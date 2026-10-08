package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const m13TestNodeID = "00000000-0000-4000-8000-000000000001"

func m13TestServer(t *testing.T, lookup m13PublicKeyLookup) (*httptest.Server, *websocket.Dialer) {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveM13WebSocket(w, r, lookup)
	}))
	t.Cleanup(server.Close)
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	dialer := &websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}
	return server, dialer
}

func m13TestURL(server *httptest.Server) string {
	return "wss" + strings.TrimPrefix(server.URL, "https")
}

func m13ReadChallenge(t *testing.T, ws *websocket.Conn) string {
	t.Helper()
	_ = ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	var first m13ChallengeFrame
	if err := ws.ReadJSON(&first); err != nil {
		t.Fatal(err)
	}
	if first.Type != "challenge" || first.Challenge == "" {
		t.Fatalf("invalid challenge frame: %+v", first)
	}
	return first.Challenge
}

func TestM13WebSocketRequiresDatabase(t *testing.T) {
	service := New(slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	req := httptest.NewRequest(http.MethodGet, "https://example.com/v1/agent/connect", nil)
	recorder := httptest.NewRecorder()
	service.Handler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, want 503", recorder.Code)
	}
}

func TestM13WebSocketAuthAndNoTelemetryYet(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil { t.Fatal(err) }
	lookup := func(_ context.Context, nodeID string) (string, error) {
		if nodeID != m13TestNodeID { return "", sql.ErrNoRows }
		return base64.RawStdEncoding.EncodeToString(publicKey), nil
	}
	server, dialer := m13TestServer(t, lookup)
	ws, _, err := dialer.Dial(m13TestURL(server), nil)
	if err != nil { t.Fatal(err) }
	defer ws.Close()
	challenge := m13ReadChallenge(t, ws)
	signature, err := protocol.SignM13Challenge(privateKey, m13TestNodeID, challenge)
	if err != nil { t.Fatal(err) }
	if err := ws.WriteJSON(m13ProofFrame{Type: "authenticate", NodeID: m13TestNodeID, Signature: signature}); err != nil {
		t.Fatal(err)
	}
	var accepted m13AcceptedFrame
	if err := ws.ReadJSON(&accepted); err != nil { t.Fatal(err) }
	if accepted.Type != "authenticated" || accepted.NodeID != m13TestNodeID {
		t.Fatalf("unexpected authentication response: %+v", accepted)
	}
	// A signed node is still not allowed to submit telemetry or request jobs yet.
	if err := ws.WriteJSON(map[string]string{"type":"heartbeat"}); err != nil { t.Fatal(err) }
	_, _, err = ws.ReadMessage()
	if !websocket.IsCloseError(err, websocket.CloseUnsupportedData) {
		t.Fatalf("expected unsupported-data close, got: %v", err)
	}
}

func TestM13WebSocketRejectsReplayedSignature(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil { t.Fatal(err) }
	lookup := func(_ context.Context, nodeID string) (string, error) {
		if nodeID != m13TestNodeID { return "", sql.ErrNoRows }
		return base64.RawStdEncoding.EncodeToString(publicKey), nil
	}
	server, dialer := m13TestServer(t, lookup)
	first, _, err := dialer.Dial(m13TestURL(server), nil)
	if err != nil { t.Fatal(err) }
	challenge1 := m13ReadChallenge(t, first)
	signature, err := protocol.SignM13Challenge(privateKey, m13TestNodeID, challenge1)
	if err != nil { t.Fatal(err) }
	_ = first.Close()

	second, _, err := dialer.Dial(m13TestURL(server), nil)
	if err != nil { t.Fatal(err) }
	defer second.Close()
	challenge2 := m13ReadChallenge(t, second)
	if challenge1 == challenge2 { t.Fatal("reused challenge") }
	if err := second.WriteJSON(m13ProofFrame{Type:"authenticate", NodeID:m13TestNodeID, Signature:signature}); err != nil {
		t.Fatal(err)
	}
	_, _, err = second.ReadMessage()
	if !websocket.IsCloseError(err, websocket.ClosePolicyViolation) {
		t.Fatalf("expected replay rejection, got %v", err)
	}
}

func TestM13WebSocketRejectsUnregisteredNode(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil { t.Fatal(err) }
	server, dialer := m13TestServer(t, func(context.Context,string)(string,error) {
		return "", sql.ErrNoRows
	})
	ws, _, err := dialer.Dial(m13TestURL(server), nil)
	if err != nil { t.Fatal(err) }
	defer ws.Close()
	challenge := m13ReadChallenge(t, ws)
	signature, err := protocol.SignM13Challenge(privateKey, m13TestNodeID, challenge)
	if err != nil { t.Fatal(err) }
	if err := ws.WriteJSON(m13ProofFrame{Type:"authenticate",NodeID:m13TestNodeID,Signature:signature}); err != nil { t.Fatal(err) }
	_, _, err = ws.ReadMessage()
	if !websocket.IsCloseError(err, websocket.ClosePolicyViolation) {
		t.Fatalf("expected unknown node rejection, got: %v", err)
	}
}

func TestM13WebSocketRejectsPlaintextAndBrowserOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveM13WebSocket(w, r, func(context.Context, string) (string, error) {
			return "", sql.ErrNoRows
		})
	})
	plain := httptest.NewServer(handler)
	defer plain.Close()
	_, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(plain.URL, "http"), nil)
	if err == nil { t.Fatal("plaintext WebSocket was accepted") }
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("plaintext status=%v, want 403", response)
	}
	_ = response.Body.Close()

	server, dialer := m13TestServer(t, func(context.Context, string) (string, error) {
		return "", sql.ErrNoRows
	})
	header := make(http.Header)
	header.Set("Origin", "https://unauthorized.example")
	_, response, err = dialer.Dial(m13TestURL(server), header)
	if err == nil { t.Fatal("browser-origin WebSocket was accepted") }
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("origin status=%v, want 403", response)
	}
	_ = response.Body.Close()
}
