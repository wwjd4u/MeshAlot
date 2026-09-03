package server_test

import("encoding/json";"io";"log/slog";"net/http";"net/http/httptest";"testing";"github.com/wwjd4u/MeshAlot/agent";protocol "github.com/wwjd4u/MeshAlot/protocol/v1";"github.com/wwjd4u/MeshAlot/server")
func TestVerticalSlice(t *testing.T){
	api:=httptest.NewServer(server.New(slog.New(slog.NewTextHandler(io.Discard,nil)),"test-token").Handler());defer api.Close()
	client:=agent.NewClient(api.URL,"node-test-001","test-token")
	if err:=client.Enroll(t.Context());err!=nil{t.Fatal(err)}
	if err:=client.Heartbeat(t.Context());err!=nil{t.Fatal(err)}
	resp,err:=http.Get(api.URL+"/v1/nodes");if err!=nil{t.Fatal(err)};defer resp.Body.Close()
	var got protocol.NodesResponse;if err=json.NewDecoder(resp.Body).Decode(&got);err!=nil{t.Fatal(err)}
	if len(got.Nodes)!=1||got.Nodes[0].Status!="online"{t.Fatalf("unexpected nodes: %+v",got.Nodes)}
}
