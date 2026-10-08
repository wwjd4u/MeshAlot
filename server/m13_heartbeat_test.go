package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func m13HeartbeatFixture(now time.Time) protocol.M13Heartbeat {
	ram := uint64(60 << 30)
	vram := uint64(18 << 30)
	cpu, gpu, latency := 24.5, 19.4, 32.1
	return protocol.M13Heartbeat{
		Type: "heartbeat", Online: true,
		GPUUtilizationPercent: &gpu, AvailableVRAMBytes: &vram,
		AvailableRAMBytes: &ram, CPULoadPercent: &cpu,
		ActiveJobState: "idle", RecentLatencyMS: &latency,
		AvailabilityMode: "normal", ObservedAt: now.UTC(),
	}
}

func m13TelemetryTestServer(t *testing.T, lookup m13PublicKeyLookup, record m13HeartbeatRecorder) (*httptest.Server, *websocket.Dialer) {
	t.Helper()
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		serveM13WebSocketWithTelemetry(w,r,lookup,record)
	}))
	t.Cleanup(s.Close)
	pool := x509.NewCertPool()
	pool.AddCert(s.Certificate())
	return s, &websocket.Dialer{HandshakeTimeout: 5*time.Second,TLSClientConfig: &tls.Config{RootCAs:pool}}
}

func m13AuthenticateTelemetry(t *testing.T, dialer *websocket.Dialer, server *httptest.Server, private ed25519.PrivateKey) *websocket.Conn {
	t.Helper()
	ws,_,err:=dialer.Dial("wss"+strings.TrimPrefix(server.URL,"https"),nil)
	if err!=nil {t.Fatal(err)}
	challenge:=m13ReadChallenge(t,ws)
	sig,err:=protocol.SignM13Challenge(private,m13TestNodeID,challenge)
	if err!=nil {t.Fatal(err)}
	if err=ws.WriteJSON(m13ProofFrame{Type:"authenticate",NodeID:m13TestNodeID,Signature:sig}); err!=nil {t.Fatal(err)}
	var ack m13AcceptedFrame
	if err=ws.ReadJSON(&ack); err!=nil {t.Fatal(err)}
	if ack.Type!="authenticated" || ack.NodeID!=m13TestNodeID {t.Fatalf("bad auth: %+v",ack)}
	return ws
}

func TestM13TelemetryAuthenticatedNodeOnly(t *testing.T) {
	pub,priv,err:=ed25519.GenerateKey(rand.Reader)
	if err!=nil {t.Fatal(err)}
	lookup:=func(_ context.Context,id string)(string,error) {
		if id!=m13TestNodeID {return "",sql.ErrNoRows}
		return base64.RawStdEncoding.EncodeToString(pub),nil
	}
	type item struct {node string; heartbeat protocol.M13Heartbeat}
	got:=make(chan item,2)
	record:=func(_ context.Context,id string,h protocol.M13Heartbeat)error {
		got<-item{id,h}
		return nil
	}
	s,dialer:=m13TelemetryTestServer(t,lookup,record)
	ws:=m13AuthenticateTelemetry(t,dialer,s,priv)
	defer ws.Close()
	for i:=0; i<2; i++ {
		h:=m13HeartbeatFixture(time.Now().UTC())
		if i==1 {h.ActiveJobState="running";h.ManualPause=true}
		if err:=ws.WriteJSON(h);err!=nil {t.Fatal(err)}
		var ack m13HeartbeatAckFrame
		if err:=ws.ReadJSON(&ack);err!=nil {t.Fatal(err)}
		if ack.Type!="heartbeat_ack" {t.Fatalf("no acknowledgement: %+v",ack)}
		select {
		case result:=<-got:
			if result.node!=m13TestNodeID || result.heartbeat.ActiveJobState!=h.ActiveJobState {
				t.Fatalf("identity or payload changed: %+v",result)
			}
		case <-time.After(3*time.Second): t.Fatal("no heartbeat recorded")
		}
	}
}

func TestM13TelemetryRejectsNodeSpoof(t *testing.T) {
	pub,priv,err:=ed25519.GenerateKey(rand.Reader)
	if err!=nil {t.Fatal(err)}
	writes:=make(chan string,1)
	lookup:=func(_ context.Context,_ string)(string,error) {
		return base64.RawStdEncoding.EncodeToString(pub),nil
	}
	record:=func(_ context.Context,id string,_ protocol.M13Heartbeat)error {writes<-id; return nil}
	s,dialer:=m13TelemetryTestServer(t,lookup,record)
	ws:=m13AuthenticateTelemetry(t,dialer,s,priv)
	defer ws.Close()
	frame,err:=json.Marshal(m13HeartbeatFixture(time.Now()))
	if err!=nil {t.Fatal(err)}
	m:=make(map[string]any)
	if err=json.Unmarshal(frame,&m);err!=nil {t.Fatal(err)}
	m["node_id"]="00000000-0000-4000-8000-000000000002"
	if err=ws.WriteJSON(m);err!=nil {t.Fatal(err)}
	_,_,err=ws.ReadMessage()
	if !websocket.IsCloseError(err,websocket.ClosePolicyViolation) {
		t.Fatalf("spoofed node not rejected: %v",err)
	}
	select {
	case <-writes:t.Fatal("spoofed heartbeat was stored")
	default:
	}
}

func TestM13TelemetryRejectsInvalidPayloadAndFailedWrite(t *testing.T) {
	now:=time.Now().UTC()
	for _, raw:=range []string{
		`{}`,
		`{"type":"heartbeat","online":true}`,
		`{"type":"heartbeat","online":true}{"type":"heartbeat"}`,
	} {
		if _,err:=m13DecodeHeartbeat([]byte(raw),now);err==nil {t.Fatalf("accepted %q",raw)}
	}
	pub,priv,err:=ed25519.GenerateKey(rand.Reader)
	if err!=nil {t.Fatal(err)}
	lookup:=func(context.Context,string)(string,error) {return base64.RawStdEncoding.EncodeToString(pub),nil}
	s,dialer:=m13TelemetryTestServer(t,lookup,func(context.Context,string,protocol.M13Heartbeat)error {
		return errors.New("injected database failure")
	})
	ws:=m13AuthenticateTelemetry(t,dialer,s,priv)
	defer ws.Close()
	if err=ws.WriteJSON(m13HeartbeatFixture(time.Now().UTC()));err!=nil {t.Fatal(err)}
	_,_,err=ws.ReadMessage()
	if !websocket.IsCloseError(err,websocket.CloseInternalServerErr) {
		t.Fatalf("failed database write acknowledged or wrong close: %v",err)
	}
}
