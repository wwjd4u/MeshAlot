package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestM13EffectiveNodeStatus(t *testing.T) {
	now:=time.Date(2026,10,8,1,0,0,0,time.UTC)
	cases:=[]struct{
		name string
		status string
		last time.Time
		m13 bool
		want string
	}{
		{"fresh M13","online",now.Add(-30*time.Second),true,"online"},
		{"exact threshold M13","online",now.Add(-90*time.Second),true,"online"},
		{"expired M13","online",now.Add(-91*time.Second),true,"offline"},
		{"never reported M13","online",time.Time{},true,"offline"},
		{"old legacy heartbeat","online",now.Add(-48*time.Hour),false,"online"},
		{"enrolled untouched","enrolled",time.Time{},true,"enrolled"},
		{"offline stays offline","offline",now,true,"offline"},
		{"pending stays pending","pending",now,true,"pending"},
	}
	for _,tt:=range cases{
		t.Run(tt.name,func(t *testing.T){
			got:=M13EffectiveNodeStatus(tt.status,tt.last,tt.m13,now)
			if got!=tt.want {t.Fatalf("status=%q want %q",got,tt.want)}
		})
	}
}

func TestM13StaleMonitorSweepsThenStops(t *testing.T){
	ctx,cancel:=context.WithCancel(context.Background())
	calls:=make(chan struct{},2)
	finished:=make(chan struct{})
	logger:=slog.New(slog.NewTextHandler(io.Discard,nil))
	go func(){
		RunM13StaleMonitor(ctx,2*time.Millisecond,
			func(_ context.Context)(int64,error){
				select {case calls<-struct{}{}:default:}
				return 1,nil
			},logger)
		close(finished)
	}()
	select {
	case <-calls:
	case <-time.After(3*time.Second): t.Fatal("stale sweep never started")
	}
	cancel()
	select {
	case <-finished:
	case <-time.After(3*time.Second): t.Fatal("stale monitor ignored cancellation")
	}
}

func TestM13StaleMonitorHandlesErrorsWithoutStopping(t *testing.T){
	ctx,cancel:=context.WithCancel(context.Background())
	defer cancel()
	done:=make(chan struct{})
	calls:=make(chan struct{},2)
	logger:=slog.New(slog.NewTextHandler(io.Discard,nil))
	go func(){
		counter:=0
		RunM13StaleMonitor(ctx,2*time.Millisecond,
			func(_ context.Context)(int64,error){
				counter++
				select {case calls<-struct{}{}:default:}
				if counter==1 {return 0,errors.New("simulated DB error")}
				return 0,nil
			},logger)
		close(done)
	}()
	for i:=0;i<2;i++{
		select {
		case <-calls:
		case <-time.After(3*time.Second):t.Fatal("monitor did not retry after error")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3*time.Second):t.Fatal("monitor did not exit")
	}
}
