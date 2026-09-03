package main

import("context";"flag";"log/slog";"os";"github.com/wwjd4u/MeshAlot/agent")
func main(){
	serverURL:=flag.String("server","http://127.0.0.1:8080","control API URL")
	nodeID:=flag.String("node","meshalot-development-node-001","persistent node ID placeholder")
	token:=flag.String("token","dev-enrollment-token","development enrollment token");flag.Parse()
	logger:=slog.New(slog.NewJSONHandler(os.Stdout,nil));client:=agent.NewClient(*serverURL,*nodeID,*token)
	if err:=client.Enroll(context.Background());err!=nil{logger.Error("enrollment failed","error",err);os.Exit(1)}
	if err:=client.Heartbeat(context.Background());err!=nil{logger.Error("heartbeat failed","error",err);os.Exit(1)}
	logger.Info("placeholder agent cycle completed","node_id",*nodeID)
}
