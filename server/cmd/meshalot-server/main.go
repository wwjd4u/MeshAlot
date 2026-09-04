package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
	"github.com/wwjd4u/MeshAlot/server"
)
func main(){
	logger:=slog.New(slog.NewJSONHandler(os.Stdout,nil))
	addr:=env("MESHALOT_LISTEN_ADDR",":8080")
	token:=env("MESHALOT_DEV_ENROLLMENT_TOKEN","dev-enrollment-token")
	svc:=server.New(logger,token)
	if dsn:=os.Getenv("MESHALOT_DATABASE_DSN");dsn!="" {
		if token=="dev-enrollment-token" || len(token)<32 {
			logger.Error("PostgreSQL mode requires an explicit POC token of at least 32 characters");os.Exit(1)
		}
		store,err:=server.OpenPostgres(context.Background(),dsn,os.Getenv("MESHALOT_POC_USER_ID"))
		if err!=nil {logger.Error("database startup failed; check configuration and migrations");os.Exit(1)}
		defer store.Close()
		svc=server.NewWithPostgres(logger,token,store)
		logger.Info("PostgreSQL storage enabled")
	} else {logger.Warn("development in-memory storage enabled")}
	logger.Info("control API starting","addr",addr)
	httpServer:=&http.Server{Addr:addr,Handler:svc.Handler(),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second,IdleTimeout:60*time.Second}
	if err:=httpServer.ListenAndServe();err!=nil{logger.Error("server stopped","error",err);os.Exit(1)}
}
func env(name,fallback string)string{if v:=os.Getenv(name);v!=""{return v};return fallback}
