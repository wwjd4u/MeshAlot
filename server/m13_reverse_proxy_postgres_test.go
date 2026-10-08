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
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/wwjd4u/MeshAlot/agent"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
	"github.com/wwjd4u/MeshAlot/server"
)

// Caddy-shaped isolated TLS termination -> localhost HTTP reverse proxy ->
// enrolled-node WSS auth and PostgreSQL status. This is NOT a production Caddy
// configuration test or a production node connection.
func TestM13ReverseProxyRealPostgres(t *testing.T) {
	adminDSN:=os.Getenv("MESHALOT_M13_VERTICAL_ADMIN_DSN")
	runtimeDSN:=os.Getenv("MESHALOT_M13_VERTICAL_RUNTIME_DSN")
	if adminDSN=="" || runtimeDSN=="" {t.Skip("requires isolated M13 PG DSNs")}
	ctx,cancel:=context.WithTimeout(context.Background(),8*time.Second)
	defer cancel()
	admin,err:=sql.Open("postgres",adminDSN)
	if err!=nil{t.Fatal(err)}
	defer admin.Close()
	var database,role string
	if err=admin.QueryRowContext(ctx,"SELECT current_database(),current_user").Scan(&database,&role);err!=nil{t.Fatal(err)}
	if database!="meshalot_m13test" || role!="postgres" {t.Fatal("refusing non-disposable integration database")}
	var raw [8]byte
	if _,err=rand.Read(raw[:]);err!=nil{t.Fatal(err)}
	var owner string
	if err=admin.QueryRowContext(ctx,"INSERT INTO users(email) VALUES($1) RETURNING id::text",
		"m13-proxy-"+hex.EncodeToString(raw[:])+"@example.invalid").Scan(&owner);err!=nil{t.Fatal(err)}
	identityDirectory:=filepath.Join(t.TempDir(),"private")
	if err=os.Mkdir(identityDirectory,0700);err!=nil{t.Fatal(err)}
	path:=filepath.Join(identityDirectory,"agent-identity.json")
	id,created,err:=agent.LoadOrCreateIdentity(path)
	if err!=nil || !created{t.Fatalf("isolated identity fixture failed: %v",err)}
	before,err:=os.ReadFile(path)
	if err!=nil{t.Fatal(err)}
	var nodeRow string
	if err=admin.QueryRowContext(ctx,
		"INSERT INTO nodes(user_id,node_key,agent_version,identity_public_key) VALUES($1::uuid,$2,'M13-proxy-test',$3) RETURNING id::text",
		owner,id.NodeID,id.PublicKey).Scan(&nodeRow);err!=nil{t.Fatal(err)}
	if _,err=admin.ExecContext(ctx,
		"INSERT INTO node_status(node_id,status,observed_at) VALUES($1::uuid,'enrolled',now())",nodeRow);err!=nil{t.Fatal(err)}
	store,err:=server.OpenPostgres(ctx,runtimeDSN,owner)
	if err!=nil{t.Fatal(err)}
	defer store.Close()
	log:=slog.New(slog.NewTextHandler(io.Discard,nil))
	svc:=server.NewWithPostgres(log,"isolated-proxy-auth-test-only-dev-token",store)
	// Actual HTTP server has artificially short timeouts: the upgraded
	// WebSocket must survive longer than regular HTTP requests.
	backend:=httptest.NewUnstartedServer(svc.Handler())
	backend.Config.ReadHeaderTimeout=250*time.Millisecond
	backend.Config.ReadTimeout=250*time.Millisecond
	backend.Config.WriteTimeout=250*time.Millisecond
	backend.Config.IdleTimeout=250*time.Millisecond
	backend.Start()
	defer backend.Close()
	// Direct plaintext to the backend must FAIL despite localhost location.
	directWSURL:="ws"+strings.TrimPrefix(backend.URL,"http")+"/v1/agent/connect"
	unauth,resp,err:=websocket.DefaultDialer.Dial(directWSURL,nil)
	if err==nil {
		unauth.Close()
		t.Fatal("accepted direct unencrypted backend WebSocket without trusted proxy")
	}
	if resp==nil || resp.StatusCode!=http.StatusForbidden {
		t.Fatalf("plaintext backend status=%v (wanted 403)",resp)
	}
	_ = resp.Body.Close()
	target,err:=url.Parse(backend.URL)
	if err!=nil{t.Fatal(err)}
	proxy:=httputil.NewSingleHostReverseProxy(target)
	originalDirector:=proxy.Director
	proxy.Director=func(req *http.Request){
		originalDirector(req)
		// Model a trusted Caddy HTTPS-termination header. Never derive
		// this value from an external client-controlled request header.
		if req.TLS!=nil {req.Header.Set("X-Forwarded-Proto","https")}
	}
	edge:=httptest.NewTLSServer(proxy)
	defer edge.Close()
	certPool:=x509.NewCertPool()
	certPool.AddCert(edge.Certificate())
	dialer:=&websocket.Dialer{HandshakeTimeout:3*time.Second,
		TLSClientConfig:&tls.Config{RootCAs:certPool}}
	sample:=func(ctx context.Context)(protocol.M13Heartbeat,error){
		if err:=ctx.Err();err!=nil{return protocol.M13Heartbeat{},err}
		ram:=uint64(8<<30);cpu:=float64(5);lat:=float64(0)
		return protocol.M13Heartbeat{
			Type:"heartbeat",Online:true,AvailableRAMBytes:&ram,
			CPULoadPercent:&cpu,ActiveJobState:"unknown",
			RecentLatencyMS:&lat,AvailabilityMode:"away",
			ManualPause:true,ObservedAt:time.Now().UTC(),
		},nil
	}
	done:=make(chan error,1)
	go func(){done<-agent.RunM13Session(ctx,edge.URL,id,sample,400*time.Millisecond,dialer)}()
	waitForHeartbeat:=func(since time.Time,window time.Duration)time.Time{
		t.Helper()
		deadline:=time.Now().Add(window)
		for{
			var at sql.NullTime
			var state string
			err:=admin.QueryRowContext(ctx,
				"SELECT status,last_heartbeat FROM node_status WHERE node_id=$1::uuid",
				nodeRow).Scan(&state,&at)
			if err!=nil{t.Fatal(err)}
			if state=="online" && at.Valid && at.Time.After(since){return at.Time}
			select{
			case e:=<-done:t.Fatalf("WSS failed before next heartbeat: %v",e)
			default:
			}
			if time.Now().After(deadline){t.Fatal("proxy did not persist next heartbeat")}
			time.Sleep(15*time.Millisecond)
		}
	}
	first:=waitForHeartbeat(time.Time{},3*time.Second)
	// The next heartbeat occurs at >=400ms, past the 250ms standard HTTP
	// server timeouts. It must still use the upgraded WSS connection.
	second:=waitForHeartbeat(first,3*time.Second)
	if !second.After(first){t.Fatal("WebSocket stopped after HTTP request timeout")}
	node,err:=store.NodeForUser(ctx,owner,id.NodeID)
	if err!=nil || node.Status!="online"{
		t.Fatalf("proxy-backed node status=%+v error %v",node,err)
	}
	dash,err:=store.Dashboard(ctx,owner)
	if err!=nil || dash.OnlineNodes!=1{t.Fatalf("proxy-backed dashboard %+v err %v",dash,err)}
	cancel()
	select{
	case err=<-done:
		if err!=nil && !errors.Is(err,context.Canceled){
			t.Fatalf("WebSocket did not stop cleanly: %v",err)
		}
	case <-time.After(2*time.Second):
		t.Fatal("agent did not stop on cancellation")
	}
	after,err:=os.ReadFile(path)
	if err!=nil{t.Fatal(err)}
	if !bytes.Equal(before,after){t.Fatal("proxy traffic changed identity")}
	t.Log("M13_PROXY_TLS_LOOPBACK_UPGRADE_LONG_LIVED_WSS=PASS")
}
