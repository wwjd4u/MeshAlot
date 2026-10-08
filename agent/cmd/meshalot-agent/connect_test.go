package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wwjd4u/MeshAlot/agent"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func m13ConnectTestSnapshot(ctx context.Context, opts agent.M13SamplerOptions) (protocol.M13Heartbeat, error) {
	if err := ctx.Err(); err != nil {
		return protocol.M13Heartbeat{}, err
	}
	ram := uint64(64 << 30)
	cpu := 13.5
	rtt := opts.RecentLatencyMS
	return protocol.M13Heartbeat{
		Type:              "heartbeat",
		Online:            true,
		AvailableRAMBytes: &ram,
		CPULoadPercent:    &cpu,
		ActiveJobState:    opts.ActiveJobState,
		RecentLatencyMS:   &rtt,
		AvailabilityMode:  opts.AvailabilityMode,
		ManualPause:       opts.ManualPause,
		ObservedAt:        time.Now().UTC(),
	}, nil
}

func m13ConnectTestOptions(identityPath string) []string {
	return []string{
		"--server", "https://meshalot.test",
		"--identity", identityPath,
		"--mode", "away",
		"--manual-pause", "true",
		"--job-state", "running",
		"--interval", "15s",
	}
}

// Identity fixtures must live under a 0700 directory even when Go's
// temporary test root is created with broader permissions on a runner.
func m13PrivateIdentityFixture(t *testing.T) string {
	t.Helper()
	directory := filepath.Join(t.TempDir(), "private")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(directory, "identity.json")
}

func TestM13ConnectRequiresExplicitProviderState(t *testing.T) {
	var called bool
	loader := func(string) (agent.Identity, error) {
		called = true
		return agent.Identity{}, errors.New("must not read identity")
	}
	session := func(context.Context, string, agent.Identity, agent.M13HeartbeatSampler, time.Duration) error {
		t.Fatal("must not open network session")
		return nil
	}
	options := [][]string{
		{"--manual-pause", "false"},
		{"--mode", "normal"},
		{"--mode", "normal", "--manual-pause", "maybe"},
		{"--mode", "invalid", "--manual-pause", "false"},
		{"--mode", "normal", "--manual-pause", "false", "--job-state", "execute-shell"},
		{"--mode", "normal", "--manual-pause", "false", "--interval", "1ms"},
		{"--mode", "normal", "--manual-pause", "false", "--server", "http://localhost:8180"},
		{"--mode", "normal", "--manual-pause", "false", "enroll-code"},
	}
	for _, args := range options {
		if err := runConnectWithDeps(context.Background(), args, loader, m13ConnectTestSnapshot, session); err == nil {
			t.Fatalf("unsafe/missing connect settings accepted: %q", args)
		}
	}
	if called {
		t.Fatal("read local identity while options were invalid")
	}
}

func TestM13ConnectMissingIdentityNeverEnrolls(t *testing.T) {
	path := filepath.Join(t.TempDir(), "identity.json")
	var dialed bool
	session := func(context.Context, string, agent.Identity, agent.M13HeartbeatSampler, time.Duration) error {
		dialed = true
		return nil
	}
	err := runConnectWithDeps(context.Background(), m13ConnectTestOptions(path),
		agent.LoadIdentity, m13ConnectTestSnapshot, session)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing identity must fail closed, got %v", err)
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("connect created an identity file: %v", err)
	}
	if dialed {
		t.Fatal("started connection without existing identity")
	}
}

func TestM13ConnectReadsExistingIdentityWithoutMutation(t *testing.T) {
	path := m13PrivateIdentityFixture(t)
	original, created, err := agent.LoadOrCreateIdentity(path) // isolated test fixture only
	if err != nil || !created {
		t.Fatalf("failed to make isolated test identity: %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	session := func(ctx context.Context, server string, id agent.Identity,
		sample agent.M13HeartbeatSampler, interval time.Duration) error {
		called = true
		if server != "https://meshalot.test" || interval != 15*time.Second || id != original {
			t.Fatalf("incorrect connection configuration")
		}
		h, err := sample(ctx)
		if err != nil {
			return err
		}
		if h.AvailabilityMode != "away" || !h.ManualPause || h.ActiveJobState != "running" {
			t.Fatalf("provider state was silently changed: %+v", h)
		}
		return h.Validate(time.Now().UTC())
	}
	err = runConnectWithDeps(context.Background(), m13ConnectTestOptions(path),
		agent.LoadIdentity, m13ConnectTestSnapshot, session)
	if err != nil || !called {
		t.Fatalf("existing identity connection failed: called=%v err=%v", called, err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("connect changed existing identity file")
	}
}

func TestM13ConnectPropagatesCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runConnectWithDeps(ctx, m13ConnectTestOptions("nonexistent"),
		func(string) (agent.Identity, error) {
			t.Fatal("loaded identity after cancellation")
			return agent.Identity{}, nil
		},
		m13ConnectTestSnapshot,
		func(context.Context, string, agent.Identity, agent.M13HeartbeatSampler, time.Duration) error {
			t.Fatal("started session after cancellation")
			return nil
		})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestM13ConnectEndToEndIsolatedTLS(t *testing.T) {
	path := m13PrivateIdentityFixture(t)
	identity, _, err := agent.LoadOrCreateIdentity(path) // ONLY test fixture
	if err != nil {
		t.Fatal(err)
	}
	pub, err := base64.RawStdEncoding.DecodeString(identity.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	var serverErrors = make(chan error, 2)
	received := make(chan protocol.M13Heartbeat, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agent/connect" {
			serverErrors <- errors.New("unexpected endpoint")
			return
		}
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			serverErrors <- err
			return
		}
		defer ws.Close()
		challenge, err := protocol.NewM13Challenge()
		if err != nil {
			serverErrors <- err
			return
		}
		if err := ws.WriteJSON(map[string]string{"type": "challenge", "challenge": challenge}); err != nil {
			serverErrors <- err
			return
		}
		var proof struct {
			Type      string `json:"type"`
			NodeID    string `json:"node_id"`
			Signature string `json:"signature"`
		}
		if err := ws.ReadJSON(&proof); err != nil {
			serverErrors <- err
			return
		}
		if proof.Type != "authenticate" || proof.NodeID != identity.NodeID {
			serverErrors <- errors.New("wrong enrolled node")
			return
		}
		if err := protocol.VerifyM13Challenge(ed25519.PublicKey(pub), proof.NodeID, challenge, proof.Signature); err != nil {
			serverErrors <- err
			return
		}
		if err := ws.WriteJSON(map[string]string{"type": "authenticated", "node_id": identity.NodeID}); err != nil {
			serverErrors <- err
			return
		}
		var heartbeat protocol.M13Heartbeat
		if err := ws.ReadJSON(&heartbeat); err != nil {
			serverErrors <- err
			return
		}
		if err := heartbeat.Validate(time.Now().UTC()); err != nil {
			serverErrors <- err
			return
		}
		if err := ws.WriteJSON(map[string]string{"type": "heartbeat_ack"}); err != nil {
			serverErrors <- err
			return
		}
		received <- heartbeat
	}))
	defer server.Close()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	dialer := &websocket.Dialer{
		HandshakeTimeout: 3 * time.Second,
		TLSClientConfig:  &tls.Config{RootCAs: roots},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	args := []string{
		"--server", server.URL, "--identity", path, "--mode", "normal",
		"--manual-pause", "false", "--job-state", "unknown", "--interval", "10s",
	}
	session := func(ctx context.Context, url string, id agent.Identity,
		sample agent.M13HeartbeatSampler, interval time.Duration) error {
		return agent.RunM13Session(ctx, url, id, sample, interval, dialer)
	}
	done := make(chan error, 1)
	go func() {
		done <- runConnectWithDeps(ctx, args, agent.LoadIdentity, m13ConnectTestSnapshot, session)
	}()
	select {
	case h := <-received:
		if h.AvailabilityMode != "normal" || h.ActiveJobState != "unknown" || h.ManualPause {
			t.Fatalf("unexpected forwarded provider state: %+v", h)
		}
		if h.RecentLatencyMS == nil || *h.RecentLatencyMS < 0 {
			t.Fatal("missing measured TLS channel latency")
		}
		cancel()
	case err := <-serverErrors:
		t.Fatalf("isolated test server error: %v", err)
	case err := <-done:
		t.Fatalf("agent exited before first heartbeat: %v", err)
	case <-ctx.Done():
		t.Fatal("heartbeat did not reach isolated TLS server")
	}
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("agent did not exit on shutdown: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("agent did not stop on context cancellation")
	}
	after, err := agent.LoadIdentity(path)
	if err != nil || after != identity {
		t.Fatalf("identity changed after TLS session: %v", err)
	}
}
