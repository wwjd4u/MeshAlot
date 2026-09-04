package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wwjd4u/MeshAlot/agent"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestPostgresPersistence(t *testing.T) {
	dsn:=os.Getenv("MESHALOT_TEST_DSN")
	if dsn=="" {t.Skip("set MESHALOT_TEST_DSN for restricted-role PostgreSQL integration")}
	owner:=os.Getenv("MESHALOT_TEST_USER_ID")
	ctx:=context.Background()
	store,err:=OpenPostgres(ctx,dsn,owner);if err!=nil {t.Fatal(err)}
	defer store.Close()
	var database,role string
	var privileged bool
	err=store.db.QueryRowContext(ctx,`SELECT current_database(),current_user,
        rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls
        FROM pg_roles WHERE rolname=current_user`).Scan(&database,&role,&privileged)
	if err!=nil {t.Fatal(err)}
	if database!="meshalot_fresh_test" || role!="meshalot" || privileged {t.Fatal("requires meshalot restricted login on meshalot_fresh_test")}
	var owns bool
	err=store.db.QueryRowContext(ctx,"SELECT EXISTS(SELECT FROM pg_tables WHERE schemaname='public' AND tableowner=current_user)").Scan(&owns)
	if err!=nil || owns {t.Fatal("runtime role must not own core tables")}
	logger:=slog.New(slog.NewTextHandler(io.Discard,nil))
	const token="integration-test-only-token-not-a-production-secret"
	api:=httptest.NewServer(NewWithPostgres(logger,token,store).Handler())
	defer api.Close()
	var random [16]byte
	if _,err=rand.Read(random[:]);err!=nil {t.Fatal(err)}
	nodeID:="persistence-test-"+hex.EncodeToString(random[:])
	client:=agent.NewClient(api.URL,nodeID,token)
	if err=client.Enroll(ctx);err!=nil {t.Fatal(err)}
	if err=client.Heartbeat(ctx);err!=nil {t.Fatal(err)}
	readNode:=func(base string) protocol.Node {
		t.Helper()
		req,_:=http.NewRequest("GET",base+"/v1/nodes",nil)
		req.Header.Set("Authorization","Bearer "+token)
		resp,e:=http.DefaultClient.Do(req);if e!=nil {t.Fatal(e)}
		defer resp.Body.Close()
		if resp.StatusCode!=200 {t.Fatalf("node list: %s",resp.Status)}
		var list protocol.NodesResponse
		if e=json.NewDecoder(resp.Body).Decode(&list);e!=nil {t.Fatal(e)}
		for _,n:=range list.Nodes {if n.NodeID==nodeID {return n}}
		t.Fatal("persisted node missing");return protocol.Node{}
	}
	before:=readNode(api.URL)
	if before.Status!="online" || before.LastHeartbeat.IsZero() {t.Fatal("heartbeat not persisted")}
	var idBefore,idAfter string
	if err=store.db.QueryRowContext(ctx,"SELECT id::text FROM nodes WHERE node_key=$1",nodeID).Scan(&idBefore);err!=nil {t.Fatal(err)}
	api.Close()
	store.Close()
	// New HTTP server, service object, and connection pool; no memory carries over.
	store2,err:=OpenPostgres(ctx,dsn,owner);if err!=nil {t.Fatal(err)}
	defer store2.Close()
	api2:=httptest.NewServer(NewWithPostgres(logger,token,store2).Handler());defer api2.Close()
	after:=readNode(api2.URL)
	if before!=after {t.Fatalf("node changed across service recreation: %+v / %+v",before,after)}
	if err=agent.NewClient(api2.URL,nodeID,token).Enroll(ctx);err!=nil {t.Fatal(err)}
	if err=store2.db.QueryRowContext(ctx,"SELECT id::text FROM nodes WHERE node_key=$1",nodeID).Scan(&idAfter);err!=nil {t.Fatal(err)}
	if idAfter!=idBefore || readNode(api2.URL).LastHeartbeat!=before.LastHeartbeat {t.Fatal("reenrollment reset identity/heartbeat")}
	resp,err:=http.Get(api2.URL+"/v1/nodes");if err!=nil {t.Fatal(err)}
	resp.Body.Close();if resp.StatusCode!=401 {t.Fatal("unauthenticated node list not blocked")}
	if err=agent.NewClient(api2.URL,nodeID,"wrong-token").Heartbeat(ctx);err==nil {t.Fatal("unauthenticated heartbeat accepted")}
	store2.Close()
	resp,err=http.Get(api2.URL+"/v1/health");if err!=nil {t.Fatal(err)}
	resp.Body.Close();if resp.StatusCode!=503 {t.Fatal("closed database must fail health check")}
	t.Log("PASS: restricted login, persistent enrollment/heartbeat, service recreation, stable identity, authentication and database failure")
	t.Log("Test node retained in isolated database:",nodeID)
}
