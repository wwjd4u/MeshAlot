package agent

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestM13ReconnectBackoffBounds(t *testing.T) {
	tests := []struct {
		fail uint32
		want time.Duration
	}{
		{0, time.Second}, {1, time.Second}, {2, 2 * time.Second},
		{3, 4 * time.Second}, {4, 8 * time.Second}, {6, 32 * time.Second},
		{7, 60 * time.Second}, {8, 60 * time.Second}, {1000, 60 * time.Second},
	}
	for _, tt := range tests {
		if got := M13ReconnectBackoff(tt.fail); got != tt.want {
			t.Fatalf("failed sessions=%d delay=%s want=%s", tt.fail, got, tt.want)
		}
	}
}

func TestM13ReconnectLoopRetriesAndCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	var delays []time.Duration
	run := func(context.Context) error {
		calls++
		if calls == 4 {
			cancel()
			return nil
		}
		return fmt.Errorf("injected network drop %d", calls)
	}
	wait := func(_ context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	}
	if err := runM13ReconnectLoop(ctx, run, wait, time.Now); err != nil {
		t.Fatal(err)
	}
	if calls != 4 || len(delays) != 3 ||
		delays[0] != time.Second || delays[1] != 2*time.Second || delays[2] != 4*time.Second {
		t.Fatalf("incorrect retry policy calls=%d delays=%v", calls, delays)
	}
}

func TestM13ReconnectLoopResetsAfterSustainedSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	current := time.Date(2026, 10, 8, 1, 0, 0, 0, time.UTC)
	clock := func() time.Time { return current }
	calls := 0
	var delays []time.Duration
	run := func(context.Context) error {
		calls++
		if calls == 3 {
			current = current.Add(3 * time.Minute)
		}
		if calls == 4 {
			cancel()
			return nil
		}
		return errors.New("connection interrupted")
	}
	wait := func(_ context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	}
	if err := runM13ReconnectLoop(ctx, run, wait, clock); err != nil {
		t.Fatal(err)
	}
	if len(delays) != 3 || delays[0] != time.Second ||
		delays[1] != 2*time.Second || delays[2] != time.Second {
		t.Fatalf("healthy reconnect reset failed: %v", delays)
	}
}

func TestM13ReconnectLoopWaitFailureAndShutdown(t *testing.T) {
	ctx := context.Background()
	expected := errors.New("injected retry timer failure")
	err := runM13ReconnectLoop(ctx, func(context.Context) error { return errors.New("drop") },
		func(context.Context, time.Duration) error { return expected }, time.Now)
	if !errors.Is(err, expected) {
		t.Fatalf("wait error not propagated: %v", err)
	}
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	err = runM13ReconnectLoop(cancelCtx, func(context.Context) error {
		calls++
		return errors.New("drop")
	}, func(ctx context.Context, _ time.Duration) error {
		cancel()
		return ctx.Err()
	}, time.Now)
	if err != nil || calls != 1 {
		t.Fatalf("cancellation didn't stop retries: %v calls=%d", err, calls)
	}
}

func TestM13PersistentSessionRejectsInvalidConfiguration(t *testing.T) {
	id, _ := m13TestAgentIdentity(t)
	bad := id
	bad.NodeID = "not-an-existing-identity"
	if err := RunM13PersistentSession(context.Background(),
		"https://api.meshalot.com", bad, m13TestSample, 30*time.Second, nil); err == nil {
		t.Fatal("accepted invalid node identity")
	}
	if err := RunM13PersistentSession(context.Background(),
		"http://api.meshalot.com", id, m13TestSample, 30*time.Second, nil); err == nil {
		t.Fatal("accepted insecure WebSocket target")
	}
	if err := RunM13PersistentSession(context.Background(),
		"https://api.meshalot.com", id, m13TestSample, 30*time.Second,
		&websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}); err == nil {
		t.Fatal("accepted disabled TLS certificate checking")
	}
	if err := RunM13PersistentSession(context.Background(),
		"https://api.meshalot.com", id, nil, 30*time.Second, nil); err == nil {
		t.Fatal("accepted nil local sampler")
	}
	if err := RunM13PersistentSession(nil,
		"https://api.meshalot.com", id, m13TestSample, 30*time.Second, nil); err == nil {
		t.Fatal("accepted missing context")
	}
}

// Simulate a dropped TLS transport followed by a recovered connection.
// Both accepted sessions must present the same enrolled Ed25519 node ID/key.
func TestM13PersistentSessionRecoversAfterTLSDrop(t *testing.T) {
	id, pub := m13TestAgentIdentity(t)
	var connects atomic.Int32
	connectedAgain := make(chan struct{}, 1)
	serverErrors := make(chan error, 4)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agent/connect" {
			select {
			case serverErrors <- errors.New("wrong route"):
			default:
			}
			return
		}
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			select {
			case serverErrors <- err:
			default:
			}
			return
		}
		defer ws.Close()
		attempt := connects.Add(1)
		nonce, err := protocol.NewM13Challenge()
		if err != nil {
			select {
			case serverErrors <- err:
			default:
			}
			return
		}
		if err = ws.WriteJSON(map[string]string{"type": "challenge", "challenge": nonce}); err != nil {
			return
		}
		var proof struct {
			Type      string `json:"type"`
			NodeID    string `json:"node_id"`
			Signature string `json:"signature"`
		}
		if err = ws.ReadJSON(&proof); err != nil {
			return
		}
		if proof.Type != "authenticate" || proof.NodeID != id.NodeID {
			select {
			case serverErrors <- errors.New("reconnected with different ID"):
			default:
			}
			return
		}
		if err = protocol.VerifyM13Challenge(ed25519.PublicKey(pub), proof.NodeID, nonce, proof.Signature); err != nil {
			select {
			case serverErrors <- err:
			default:
			}
			return
		}
		if err = ws.WriteJSON(map[string]string{"type": "authenticated", "node_id": id.NodeID}); err != nil {
			return
		}
		var h protocol.M13Heartbeat
		if err = ws.ReadJSON(&h); err != nil {
			return
		}
		if err = h.Validate(time.Now().UTC()); err != nil {
			select {
			case serverErrors <- err:
			default:
			}
			return
		}
		if err = ws.WriteJSON(map[string]string{"type": "heartbeat_ack"}); err != nil {
			return
		}
		if attempt == 1 {
			// Abrupt transport loss: next agent heartbeat/ack must fail and
			// the reconnect loop must authenticate using the SAME key.
			_ = ws.UnderlyingConn().Close()
			return
		}
		select {
		case connectedAgain <- struct{}{}:
		default:
		}
		// Sleep briefly so cancellation, rather than another transport drop,
		// terminates the recovered client's long-running session.
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	dialer := &websocket.Dialer{HandshakeTimeout: 3 * time.Second,
		TLSClientConfig: &tls.Config{RootCAs: pool}}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- RunM13PersistentSession(ctx, server.URL, id, m13TestSample,
			20*time.Millisecond, dialer)
	}()
	select {
	case <-connectedAgain:
		if connects.Load() < 2 {
			t.Fatal("did not establish recovered connection")
		}
		cancel()
	case err := <-serverErrors:
		t.Fatalf("isolated TLS server rejected reconnect: %v", err)
	case err := <-done:
		t.Fatalf("retry supervisor exited early: %v", err)
	case <-ctx.Done():
		t.Fatal("agent did not reconnect after isolated transport loss")
	}
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("agent didn't stop on cancellation: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("retry supervisor ignored cancellation")
	}
	if count := connects.Load(); count < 2 {
		t.Fatalf("expected two signed sessions, got %d", count)
	}
}
