package main

import (
	"log/slog"
	"net/http"
	"os"
	"github.com/wwjd4u/MeshAlot/server"
)
func main(){
	logger:=slog.New(slog.NewJSONHandler(os.Stdout,nil))
	addr:=env("MESHALOT_LISTEN_ADDR",":8080")
	token:=env("MESHALOT_DEV_ENROLLMENT_TOKEN","dev-enrollment-token")
	logger.Info("control API starting","addr",addr)
	if err:=http.ListenAndServe(addr,server.New(logger,token).Handler());err!=nil{logger.Error("server stopped","error",err);os.Exit(1)}
}
func env(name,fallback string)string{if v:=os.Getenv(name);v!=""{return v};return fallback}
