// Public synthetic identity and loopback-only endpoints; never a real node client.
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const nodeID = "11111111-1111-4111-8111-111111111111"
const edge = "wss://127.0.0.1:18443/v1/agent/connect"
const caPath = "/var/lib/m13-fixture-caddy/data/caddy/pki/authorities/local/root.crt"

func must(ok bool) {
	if !ok {
		panic("synthetic probe check failed")
	}
}
func require(err error) { must(err == nil) }

func main() {
	checks := map[string]bool{}
	defer func() {
		if recover() != nil {
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"status": "FAIL", "checks": checks})
			os.Exit(1)
		}
	}()
	must(len(os.Args) == 2)
	seed := sha256.Sum256([]byte("MeshAlot PUBLIC synthetic fixture identity: never use on a real node"))
	private := ed25519.NewKeyFromSeed(seed[:])
	public := base64.RawStdEncoding.EncodeToString(private.Public().(ed25519.PublicKey))
	if os.Args[1] == "public" {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]string{"node_id": nodeID, "public_key": public})
		return
	}
	initial := os.Args[1] == "initial"
	must(initial || os.Args[1] == "reconnect")
	ca, err := os.ReadFile(caPath)
	require(err)
	roots := x509.NewCertPool()
	must(roots.AppendCertsFromPEM(ca))
	dialer := &websocket.Dialer{HandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}
	if initial {
		ws, response, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:18080/v1/agent/connect", nil)
		if ws != nil {
			ws.Close()
		}
		must(err != nil && response != nil && response.StatusCode == http.StatusForbidden)
		response.Body.Close()
		checks["plaintext_backend_denied"] = true
		ws, response, err = dialer.Dial(edge, http.Header{"Origin": []string{"https://example.invalid"}})
		if ws != nil {
			ws.Close()
		}
		must(err != nil && response != nil && response.StatusCode == http.StatusForbidden)
		response.Body.Close()
		checks["browser_origin_denied"] = true
	}
	connect := func() *websocket.Conn {
		ws, _, err := dialer.Dial(edge, nil)
		require(err)
		require(ws.SetReadDeadline(time.Now().Add(5 * time.Second)))
		require(ws.SetWriteDeadline(time.Now().Add(5 * time.Second)))
		return ws
	}
	authenticate := func(ws *websocket.Conn, signature string) string {
		var challenge struct{ Type, Challenge string }
		require(ws.ReadJSON(&challenge))
		must(challenge.Type == "challenge")
		if signature == "" {
			signature, err = protocol.SignM13Challenge(private, nodeID, challenge.Challenge)
			require(err)
		}
		require(ws.WriteJSON(map[string]string{"type": "authenticate", "node_id": nodeID, "signature": signature}))
		return signature
	}
	ws := connect()
	defer ws.Close()
	signature := authenticate(ws, "")
	var accepted struct {
		Type   string
		NodeID string `json:"node_id"`
	}
	require(ws.ReadJSON(&accepted))
	must(accepted.Type == "authenticated" && accepted.NodeID == nodeID)
	checks["trusted_tls_and_signed_wss"] = true
	if initial {
		replay := connect()
		authenticate(replay, signature)
		var value any
		err = replay.ReadJSON(&value)
		replay.Close()
		must(websocket.IsCloseError(err, websocket.ClosePolicyViolation))
		checks["cross_connection_replay_denied"] = true
	}
	ram, cpu, latency := uint64(1<<30), float64(7), float64(0)
	heartbeat := func() {
		require(ws.SetReadDeadline(time.Now().Add(5 * time.Second)))
		require(ws.SetWriteDeadline(time.Now().Add(5 * time.Second)))
		require(ws.WriteJSON(protocol.M13Heartbeat{Type: "heartbeat", Online: true,
			AvailableRAMBytes: &ram, CPULoadPercent: &cpu, RecentLatencyMS: &latency,
			ActiveJobState: "unknown", AvailabilityMode: "away", ManualPause: true, ObservedAt: time.Now().UTC()}))
		var ack struct{ Type string }
		require(ws.ReadJSON(&ack))
		must(ack.Type == "heartbeat_ack")
	}
	heartbeat()
	if initial {
		time.Sleep(11 * time.Second)
		heartbeat()
		checks["wss_survives_http_timeout"] = true
	}
	checks["heartbeat_acknowledged"] = true
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"status": "PASS", "checks": checks})
}
