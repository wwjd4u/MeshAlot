package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// An opt-in integration test for a disposable PostgreSQL 16 instance only.
// The workflow supplies a non-networked, Unix-socket-only server and destroys
// its entire cluster on completion. NEVER set this DSN to production.
func TestM13StalePostgresIntegration(t *testing.T) {
	dsn := os.Getenv("MESHALOT_M13_STALE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated M13 PostgreSQL integration DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var dbName, username string
	if err := db.QueryRowContext(ctx, "SELECT current_database(),current_user").Scan(&dbName, &username); err != nil {
		t.Fatal(err)
	}
	if dbName != "meshalot_m13test" || username != "postgres" {
		t.Fatal("M13 stale test refuses non-disposable database or role")
	}
	var suffixBytes [8]byte
	if _,err := rand.Read(suffixBytes[:]); err != nil {t.Fatal(err)}
	suffix := hex.EncodeToString(suffixBytes[:])
	addUser := func(label string) string {
		t.Helper()
		var id string
		email := fmt.Sprintf("m13-stale-%s-%s@example.invalid",label,suffix)
		if err := db.QueryRowContext(ctx,
			"INSERT INTO users(email) VALUES($1) RETURNING id::text",email).Scan(&id);err!=nil {t.Fatal(err)}
		return id
	}
	owner1,owner2:=addUser("a"),addUser("b")
	created:=make(map[string]time.Time)
	makeNode:=func(owner,kind string,legacy bool,last time.Time)string{
		t.Helper()
		var nodeID string
		key:=fmt.Sprintf("m13-%s-%s",kind,suffix)
		if err:=db.QueryRowContext(ctx,
			"INSERT INTO nodes(user_id,node_key,agent_version,identity_public_key) VALUES($1::uuid,$2,'m13-test',$3) RETURNING id::text",
			owner,key,"test-public-key-"+kind+"-"+suffix).Scan(&nodeID);err!=nil {t.Fatal(err)}
		var telemetry any = `{"type":"heartbeat","online":true}`
		if legacy {telemetry=nil}
		if _,err:=db.ExecContext(ctx,
			`INSERT INTO node_status(node_id,status,observed_at,last_heartbeat,mode,m13_telemetry)
			  VALUES($1::uuid,'online',$2,$2,'normal',$3::jsonb)`,nodeID,last,telemetry);err!=nil {t.Fatal(err)}
		created[key]=last
		return key
	}
	now:=time.Now().UTC().Truncate(time.Microsecond)
	staleA:=makeNode(owner1,"stale-a",false,now.Add(-120*time.Second))
	freshA:=makeNode(owner1,"fresh-a",false,now.Add(-15*time.Second))
	legacyA:=makeNode(owner1,"legacy-a",true,now.Add(-7*24*time.Hour))
	staleB:=makeNode(owner2,"stale-b",false,now.Add(-120*time.Second))
	store:=&PostgresStore{db:db,owner:owner1}
	lookupStatus:=func(user,key string)(protocol.Node,error) {
		return store.NodeForUser(ctx,user,key)
	}
	check:=func(user,key,want string){
		t.Helper()
		n,err:=lookupStatus(user,key)
		if err!=nil {t.Fatal(err)}
		if n.Status!=want {t.Fatalf("node %s: status=%q want %q",key,n.Status,want)}
	}
	// The API read paths must show offline immediately, BEFORE the sweep.
	check(owner1,staleA,"offline")
	check(owner1,freshA,"online")
	check(owner1,legacyA,"online")
	check(owner2,staleB,"offline")
	if _,err:=lookupStatus(owner1,staleB);!errors.Is(err,sql.ErrNoRows) {
		t.Fatalf("other account node leaked: %v",err)
	}
	nodes,err:=store.NodesForUser(ctx,owner1)
	if err!=nil {t.Fatal(err)}
	if len(nodes)!=3 {t.Fatalf("owner1 node count=%d want 3",len(nodes))}
	seen:=map[string]string{}
	for _,n:=range nodes {seen[n.NodeID]=n.Status}
	if seen[staleA]!="offline" || seen[freshA]!="online" || seen[legacyA]!="online"{
		t.Fatalf("incorrect per-user online states: %+v",seen)
	}
	dash,err:=store.Dashboard(ctx,owner1)
	if err!=nil {t.Fatal(err)}
	if dash.OnlineNodes!=2 {
		t.Fatalf("dashboard counted stale node before sweep: %+v",dash)
	}
	// The sweep must persist offline status for BOTH stale M13 nodes while
	// preserving the complete last heartbeat and latest telemetry payload.
	updated,err:=store.MarkStaleM13NodesOffline(ctx)
	if err!=nil {t.Fatal(err)}
	if updated!=2 {t.Fatalf("sweep affected %d rows, want 2",updated)}
	updated,err=store.MarkStaleM13NodesOffline(ctx)
	if err!=nil || updated!=0 {t.Fatalf("second sweep changed rows %d: %v",updated,err)}
	check(owner1,staleA,"offline")
	check(owner1,freshA,"online")
	check(owner1,legacyA,"online")
	check(owner2,staleB,"offline")
	for _,key:=range []string{staleA,freshA,legacyA,staleB}{
		var status string
		var last time.Time
		var telemetry sql.NullString
		if err:=db.QueryRowContext(ctx,
			`SELECT s.status,s.last_heartbeat,s.m13_telemetry::text
			  FROM node_status s JOIN nodes n ON n.id=s.node_id WHERE n.node_key=$1`,
			key).Scan(&status,&last,&telemetry);err!=nil {t.Fatal(err)}
		if !last.Equal(created[key]) {t.Fatalf("heartbeat timestamp modified for %s",key)}
		if telemetry.Valid==(key==legacyA) {t.Fatalf("telemetry unexpectedly changed for %s",key)}
	}
	// A genuine fresh authenticated heartbeat brings back the same node
	// with its existing node ID, without reenrollment or history erasure.
	ram:=uint64(32<<30);cpu:=float64(12);lat:=float64(11)
	h:=protocol.M13Heartbeat{
		Type:"heartbeat",Online:true,AvailableRAMBytes:&ram,CPULoadPercent:&cpu,
		ActiveJobState:"idle",RecentLatencyMS:&lat,AvailabilityMode:"normal",
		ObservedAt:time.Now().UTC(),
	}
	if err:=store.RecordM13Heartbeat(ctx,staleA,h);err!=nil {t.Fatal(err)}
	check(owner1,staleA,"online")
	dash,err=store.Dashboard(ctx,owner1)
	if err!=nil {t.Fatal(err)}
	if dash.OnlineNodes!=3 {
		t.Fatalf("dashboard did not recover after heartbeat: %+v",dash)
	}
	t.Log("M13_POSTGRES_STALE_ONLINE_RECOVERY_AND_ACCOUNT_ISOLATION=PASS")
}
