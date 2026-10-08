package server_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/wwjd4u/MeshAlot/agent"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
	"github.com/wwjd4u/MeshAlot/server"
)

// Opt-in, disposable PostgreSQL 16 + real authenticated server + real agent.
// Never execute this test against live MeshAlot databases or existing node IDs.
func TestM13VerticalSliceRealPostgres(t *testing.T) {
	adminDSN := os.Getenv("MESHALOT_M13_VERTICAL_ADMIN_DSN")
	runtimeDSN := os.Getenv("MESHALOT_M13_VERTICAL_RUNTIME_DSN")
	if adminDSN == "" || runtimeDSN == "" {
		t.Skip("requires isolated M13 vertical-slice PostgreSQL DSNs")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	admin, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	var dbName, role string
	if err = admin.QueryRowContext(ctx, "SELECT current_database(),current_user").Scan(&dbName, &role); err != nil {
		t.Fatal(err)
	}
	if dbName != "meshalot_m13test" || role != "postgres" {
		t.Fatal("refusing non-disposable integration database")
	}
	ownerBytes := make([]byte, 8)
	if _, err = rand.Read(ownerBytes); err != nil {
		t.Fatal(err)
	}
	ownerLabel := hex.EncodeToString(ownerBytes)
	var owner string
	if err = admin.QueryRowContext(ctx,
		"INSERT INTO users(email) VALUES($1) RETURNING id::text",
		"m13-vertical-"+ownerLabel+"@example.invalid").Scan(&owner); err != nil {
		t.Fatal(err)
	}
	privateDir := filepath.Join(t.TempDir(), "private")
	if err = os.Mkdir(privateDir, 0700); err != nil {
		t.Fatal(err)
	}
	identityFile := filepath.Join(privateDir, "agent-identity.json")
	id, created, err := agent.LoadOrCreateIdentity(identityFile)
	if err != nil || !created {
		t.Fatalf("isolated test identity unavailable: %v", err)
	}
	before, err := os.ReadFile(identityFile)
	if err != nil {
		t.Fatal(err)
	}
	var rowID string
	if err = admin.QueryRowContext(ctx,
		"INSERT INTO nodes(user_id,node_key,agent_version,identity_public_key) VALUES($1::uuid,$2,$3,$4) RETURNING id::text",
		owner, id.NodeID, "test-M13", id.PublicKey).Scan(&rowID); err != nil {
		t.Fatal(err)
	}
	if _, err = admin.ExecContext(ctx,
		"INSERT INTO node_status(node_id,status,observed_at) VALUES($1::uuid,'enrolled',now())", rowID); err != nil {
		t.Fatal(err)
	}
	store, err := server.OpenPostgres(ctx, runtimeDSN, owner)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	const devToken = "isolated-M13-only-not-for-production-test-token"
	svc := server.NewWithPostgres(log, devToken, store)
	// Capture actual hijacked TLS sockets to simulate an abrupt internet drop
	// while keeping the same listening endpoint available for reconnection.
	sockets := make(chan net.Conn, 4)
	api := httptest.NewUnstartedServer(svc.Handler())
	api.Config.ConnState = func(c net.Conn, state http.ConnState) {
		if state == http.StateHijacked {
			select {
			case sockets <- c:
			default:
			}
		}
	}
	api.StartTLS()
	defer api.Close()
	roots := x509.NewCertPool()
	roots.AddCert(api.Certificate())
	dialer := &websocket.Dialer{HandshakeTimeout: 3 * time.Second,
		TLSClientConfig: &tls.Config{RootCAs: roots}}
	sampler := func(ctx context.Context) (protocol.M13Heartbeat, error) {
		if err := ctx.Err(); err != nil {
			return protocol.M13Heartbeat{}, err
		}
		ram := uint64(32 << 30)
		vram := uint64(12 << 30)
		gpu, cpu, lat := float64(24), float64(16), float64(0)
		return protocol.M13Heartbeat{
			Type: "heartbeat", Online: true,
			GPUUtilizationPercent: &gpu, AvailableVRAMBytes: &vram,
			AvailableRAMBytes: &ram, CPULoadPercent: &cpu,
			ActiveJobState: "unknown", RecentLatencyMS: &lat,
			AvailabilityMode: "normal", ManualPause: true, ObservedAt: time.Now().UTC(),
		}, nil
	}
	done := make(chan error, 1)
	go func() {
		done <- agent.RunM13PersistentSession(ctx, api.URL, id, sampler,
			25*time.Millisecond, dialer)
	}()
	waitStatus := func(want string, seconds time.Duration) {
		t.Helper()
		deadline := time.Now().Add(seconds)
		for {
			n, e := store.NodeForUser(ctx, owner, id.NodeID)
			if e == nil && n.Status == want {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("node status never became %s: got %+v error %v", want, n, e)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	var first net.Conn
	select {
	case first = <-sockets:
	case err := <-done:
		t.Fatalf("agent exited before initial TLS connect: %v", err)
	case <-ctx.Done():
		t.Fatal("agent did not establish first signed TLS connection")
	}
	waitStatus("online", 2*time.Second)
	var firstTimestamp time.Time
	var jsonType, mode string
	if err = admin.QueryRowContext(ctx,
		"SELECT last_heartbeat,m13_telemetry->>'type',mode FROM node_status WHERE node_id=$1::uuid",
		rowID).Scan(&firstTimestamp, &jsonType, &mode); err != nil {
		t.Fatal(err)
	}
	if firstTimestamp.IsZero() || jsonType != "heartbeat" || mode != "normal" {
		t.Fatalf("real Postgres snapshot mismatch: %v %q %q", firstTimestamp, jsonType, mode)
	}
	dash, err := store.Dashboard(ctx, owner)
	if err != nil || dash.OnlineNodes != 1 {
		t.Fatalf("dashboard did not display initial live node: %+v %v", dash, err)
	}
	// Kill the first actual TLS socket, not the HTTP server, simulating loss
	// of internet while the endpoint remains available for later reconnect.
	_ = first.Close()
	// Simulate >90 seconds asleep without waiting real time. The test DB,
	// not the host clock or physical network, is advanced to the stale state.
	if _, err = admin.ExecContext(ctx,
		"UPDATE node_status SET last_heartbeat=now()-interval '121 seconds' WHERE node_id=$1::uuid",
		rowID); err != nil {
		t.Fatal(err)
	}
	waitStatus("offline", time.Second)
	count, err := store.MarkStaleM13NodesOffline(ctx)
	if err != nil || count != 1 {
		t.Fatalf("offline sweep count %d error %v", count, err)
	}
	dash, err = store.Dashboard(ctx, owner)
	if err != nil || dash.OnlineNodes != 0 {
		t.Fatalf("stale node still counted online: %+v %v", dash, err)
	}
	// A second TLS connection only exists after a fresh signed challenge.
	var second net.Conn
	select {
	case second = <-sockets:
	case err = <-done:
		t.Fatalf("agent exited rather than reconnecting: %v", err)
	case <-ctx.Done():
		t.Fatal("agent did not reconnect after simulated outage")
	}
	if second == first {
		t.Fatal("reused closed connection")
	}
	waitStatus("online", 3*time.Second)
	dash, err = store.Dashboard(ctx, owner)
	if err != nil || dash.OnlineNodes != 1 {
		t.Fatalf("dashboard not online after reconnect: %+v %v", dash, err)
	}
	after, err := os.ReadFile(identityFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("reconnect modified existing private identity")
	}
	cancel()
	select {
	case err = <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("agent ignored clean stop: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("agent did not stop after cancellation")
	}
	t.Log("M13_VERTICAL_TLS_AUTH_DB_OFFLINE_RECONNECT_DASHBOARD=PASS")
}
