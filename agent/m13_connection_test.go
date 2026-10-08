package agent

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func m13TestAgentIdentity(t *testing.T) (Identity, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return Identity{Version: 1, NodeID: "00000000-0000-4000-8000-000000000001",
		PublicKey:  base64.RawStdEncoding.EncodeToString(pub),
		PrivateKey: base64.RawStdEncoding.EncodeToString(priv)}, pub
}
func m13TestDialer(t *testing.T, s *httptest.Server) *websocket.Dialer {
	t.Helper()
	roots := x509.NewCertPool()
	roots.AddCert(s.Certificate())
	return &websocket.Dialer{HandshakeTimeout: 3 * time.Second, TLSClientConfig: &tls.Config{RootCAs: roots}}
}
func m13TestSample(_ context.Context) (protocol.M13Heartbeat, error) {
	ram := uint64(32 << 30)
	vram := uint64(16 << 30)
	cpu, gpu, lat := 10.0, 20.0, 60000.0
	return protocol.M13Heartbeat{
		Type: "heartbeat", Online: true, GPUUtilizationPercent: &gpu,
		AvailableVRAMBytes: &vram, AvailableRAMBytes: &ram, CPULoadPercent: &cpu,
		ActiveJobState: "idle", RecentLatencyMS: &lat,
		AvailabilityMode: "normal", ObservedAt: time.Now().UTC(),
	}, nil
}

func TestM13AgentSecureWebSocketURL(t *testing.T) {
	got, err := M13WebSocketURL("https://api.meshalot.com/")
	if err != nil || got != "wss://api.meshalot.com/v1/agent/connect" {
		t.Fatalf("unexpected WSS URL %q %v", got, err)
	}
	for _, bad := range []string{"", "http://api.meshalot.com", "ws://127.0.0.1:8180",
		"https://user:pass@api.meshalot.com",
		"https://api.meshalot.com/custom",
		"https://api.meshalot.com/?key=secret",
		"https://api.meshalot.com/#fragment",
		"api.meshalot.com",
	} {
		if _, err := M13WebSocketURL(bad); err == nil {
			t.Fatalf("accepted insecure/invalid base URL %q", bad)
		}
	}
}

func TestM13AgentSessionAuthenticatedAndHeartbeat(t *testing.T) {
	identity, pub := m13TestAgentIdentity(t)
	received := make(chan protocol.M13Heartbeat, 3)
	protocolErrors := make(chan error, 1)
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agent/connect" {
			protocolErrors <- errors.New("wrong URL path")
			return
		}
		ws, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		if err != nil {
			protocolErrors <- err
			return
		}
		defer ws.Close()
		challenge, err := protocol.NewM13Challenge()
		if err != nil {
			protocolErrors <- err
			return
		}
		if err = ws.WriteJSON(m13AgentChallenge{Type: "challenge", Challenge: challenge}); err != nil {
			protocolErrors <- err
			return
		}
		var proof m13AgentProof
		if err = ws.ReadJSON(&proof); err != nil {
			protocolErrors <- err
			return
		}
		if proof.Type != "authenticate" || proof.NodeID != identity.NodeID {
			protocolErrors <- errors.New("invalid node proof")
			return
		}
		if err = protocol.VerifyM13Challenge(pub, proof.NodeID, challenge, proof.Signature); err != nil {
			protocolErrors <- err
			return
		}
		if err = ws.WriteJSON(m13AgentAuthResponse{Type: "authenticated", NodeID: identity.NodeID}); err != nil {
			protocolErrors <- err
			return
		}
		for {
			var h protocol.M13Heartbeat
			if err = ws.ReadJSON(&h); err != nil {
				return
			}
			if err = h.Validate(time.Now().UTC()); err != nil {
				protocolErrors <- err
				return
			}
			received <- h
			if err = ws.WriteJSON(m13AgentHeartbeatAck{Type: "heartbeat_ack"}); err != nil {
				return
			}
		}
	}))
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- RunM13Session(ctx, s.URL, identity, m13TestSample, 20*time.Millisecond, m13TestDialer(t, s))
	}()
	for i := 0; i < 2; i++ {
		select {
		case h := <-received:
			if h.RecentLatencyMS == nil || *h.RecentLatencyMS < 0 || *h.RecentLatencyMS >= 60000 {
				t.Fatalf("session must report measured control RTT, not sampler placeholder: %+v", h.RecentLatencyMS)
			}
			if !h.Online || h.AvailabilityMode != "normal" {
				t.Fatalf("invalid heartbeat %+v", h)
			}
		case err := <-protocolErrors:
			t.Fatalf("test server rejected session: %v", err)
		case <-time.After(3 * time.Second):
			t.Fatal("agent did not send heartbeat")
		}
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("session did not stop cleanly: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("session failed to stop on cancellation")
	}
}

func TestM13AgentSessionRejectsIdentityAndNoHTTPFallback(t *testing.T) {
	identity, _ := m13TestAgentIdentity(t)
	bad := identity
	bad.PrivateKey = base64.RawStdEncoding.EncodeToString(make([]byte, ed25519.PrivateKeySize))
	if err := RunM13Session(context.Background(), "https://api.meshalot.com", bad,
		m13TestSample, M13DefaultHeartbeatInterval, nil); err == nil {
		t.Fatal("accepted mismatched identity")
	}
	if err := RunM13Session(context.Background(), "http://127.0.0.1:8080", identity,
		m13TestSample, M13DefaultHeartbeatInterval, nil); err == nil {
		t.Fatal("permitted plaintext fallback")
	}
	if err := RunM13Session(context.Background(), "https://api.meshalot.com", identity,
		nil, M13DefaultHeartbeatInterval, nil); err == nil {
		t.Fatal("accepted missing sampler")
	}
	if err := RunM13Session(context.Background(), "https://api.meshalot.com", identity,
		m13TestSample, M13DefaultHeartbeatInterval,
		&websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}); err == nil {
		t.Fatal("accepted disabled certificate verification")
	}
}

func TestM13AgentSessionRejectsWrongAuthenticatedNode(t *testing.T) {
	identity, _ := m13TestAgentIdentity(t)
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		challenge, _ := protocol.NewM13Challenge()
		_ = ws.WriteJSON(m13AgentChallenge{Type: "challenge", Challenge: challenge})
		var proof m13AgentProof
		if ws.ReadJSON(&proof) != nil {
			return
		}
		_ = ws.WriteJSON(m13AgentAuthResponse{Type: "authenticated", NodeID: "00000000-0000-4000-8000-000000000002"})
	}))
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := RunM13Session(ctx, s.URL, identity, m13TestSample, 20*time.Millisecond, m13TestDialer(t, s))
	if err == nil || !strings.Contains(err.Error(), "expected node") {
		t.Fatalf("accepted different node identity: %v", err)
	}
}
