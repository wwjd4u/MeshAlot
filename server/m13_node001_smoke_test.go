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
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/wwjd4u/MeshAlot/agent"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
	"github.com/wwjd4u/MeshAlot/server"
)

// Opt-in local-only smoke: one authenticated TLS session using the ALREADY
// enrolled Node001 key and real passive host telemetry. No enroll call, no
// prod traffic, no shell/job request, no long-lived service installation.
// This test MUST run only on the designated MS-02 runner and an ephemeral DB.
func TestM13Node001LocalIdentitySmoke(t *testing.T) {
	if os.Getenv("MESHALOT_M13_NODE001_SMOKE") != "1" {
		t.Skip("opt-in test for existing local Node001 identity")
	}
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	if hostname != "wwjd4u-MS-02-Ultra" {
		t.Fatal("refusing to read Node001 identity on unapproved host")
	}
	adminDSN := os.Getenv("MESHALOT_M13_VERTICAL_ADMIN_DSN")
	runtimeDSN := os.Getenv("MESHALOT_M13_VERTICAL_RUNTIME_DSN")
	if adminDSN == "" || runtimeDSN == "" {
		t.Fatal("disposable PostgreSQL DSNs must be provided")
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
		t.Fatal("refusing test on any non-disposable DB")
	}
	identityPath, err := agent.DefaultIdentityPath()
	if err != nil {
		t.Fatal(err)
	}
	identity, err := agent.LoadIdentity(identityPath)
	if err != nil {
		t.Fatalf("existing Node001 identity unavailable: %v", err)
	}
	original, err := os.ReadFile(identityPath)
	if err != nil {
		t.Fatal(err)
	}
	var suffixBytes [8]byte
	if _, err = rand.Read(suffixBytes[:]); err != nil {
		t.Fatal(err)
	}
	var ownerID string
	if err = admin.QueryRowContext(ctx,
		"INSERT INTO users(email) VALUES($1) RETURNING id::text",
		"m13-real-host-test-"+hex.EncodeToString(suffixBytes[:])+"@example.invalid",
	).Scan(&ownerID); err != nil {
		t.Fatal(err)
	}
	var dbNodeID string
	if err = admin.QueryRowContext(ctx,
		"INSERT INTO nodes(user_id,node_key,agent_version,identity_public_key) VALUES($1::uuid,$2,'M13-node001-smoke',$3) RETURNING id::text",
		ownerID, identity.NodeID, identity.PublicKey).Scan(&dbNodeID); err != nil {
		t.Fatal(err)
	}
	if _, err = admin.ExecContext(ctx,
		"INSERT INTO node_status(node_id,status,observed_at) VALUES($1::uuid,'enrolled',now())", dbNodeID); err != nil {
		t.Fatal(err)
	}
	store, err := server.OpenPostgres(ctx, runtimeDSN, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	const testToken = "m13-node001-local-smoke-token-not-production"
	handler := server.NewWithPostgres(log, testToken, store).Handler()
	api := httptest.NewTLSServer(handler)
	defer api.Close()
	pool := x509.NewCertPool()
	pool.AddCert(api.Certificate())
	dialer := &websocket.Dialer{HandshakeTimeout: 3 * time.Second,
		TLSClientConfig: &tls.Config{RootCAs: pool}}
	// Test-only safe provider state is explicitly paused. It does not alter
	// the user's actual local resource sharing policy or runtime configuration.
	sample := func(ctx context.Context) (protocol.M13Heartbeat, error) {
		return agent.SampleM13Heartbeat(ctx, agent.M13SamplerOptions{
			AvailabilityMode: "away", ManualPause: true,
			ActiveJobState: "unknown", RecentLatencyMS: 0,
		})
	}
	done := make(chan error, 1)
	go func() {
		done <- agent.RunM13Session(ctx, api.URL, identity, sample, 30*time.Second, dialer)
	}()
	deadline := time.Now().Add(5 * time.Second)
	var state, mode string
	var payloadType sql.NullString
	var cpuText, ramText, gpuText, vramText, paused sql.NullString
	var last sql.NullTime
	for {
		err = admin.QueryRowContext(ctx,
			`SELECT status,mode,last_heartbeat,
			  m13_telemetry->>'type',
			  m13_telemetry->>'cpu_load_percent',
			  m13_telemetry->>'available_ram_bytes',
			  m13_telemetry->>'gpu_utilization_percent',
			  m13_telemetry->>'available_vram_bytes',
			  m13_telemetry->>'manual_pause'
			FROM node_status WHERE node_id=$1::uuid`, dbNodeID,
		).Scan(&state, &mode, &last, &payloadType,
			&cpuText, &ramText, &gpuText, &vramText, &paused)
		if err != nil {
			t.Fatal(err)
		}
		if state == "online" && payloadType.Valid && payloadType.String == "heartbeat" {
			break
		}
		select {
		case err = <-done:
			t.Fatalf("existing Node001 agent failed to authenticate: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("Node001 local heartbeat did not persist")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if mode != "away" || paused.String != "true" || !last.Valid {
		t.Fatalf("unsafe or incomplete test-only status: mode=%q paused=%v last=%v", mode, paused, last)
	}
	ram, err := strconv.ParseUint(ramText.String, 10, 64)
	if err != nil || ram == 0 {
		t.Fatalf("invalid host RAM telemetry: %v", err)
	}
	cpu, err := strconv.ParseFloat(cpuText.String, 64)
	if err != nil || cpu < 0 || cpu > 100 {
		t.Fatalf("invalid host CPU telemetry: %v", err)
	}
	if !gpuText.Valid || !vramText.Valid {
		t.Fatal("installed Node001 V100 GPU/VRAM telemetry was not available")
	}
	gpu, err := strconv.ParseFloat(gpuText.String, 64)
	if err != nil || gpu < 0 || gpu > 100 {
		t.Fatalf("invalid host GPU telemetry: %v", err)
	}
	vram, err := strconv.ParseUint(vramText.String, 10, 64)
	if err != nil || vram == 0 {
		t.Fatalf("invalid Node001 VRAM telemetry: %v", err)
	}
	dash, err := store.Dashboard(ctx, ownerID)
	if err != nil || dash.OnlineNodes != 1 {
		t.Fatalf("local Node001 smoke dashboard %+v err %v", dash, err)
	}
	cancel()
	select {
	case err = <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("node agent did not stop cleanly: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("node session ignored cancellation")
	}
	after, err := os.ReadFile(identityPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Fatal("real node identity was modified")
	}
	if _, err = os.Stat(identityPath); err != nil {
		t.Fatal(err)
	}
	t.Log("M13_NODE001_REAL_IDENTITY_LOCAL_TLS_GPU_AND_DB=PASS")
}
