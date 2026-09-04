package server

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type Service struct {
	postgres *PostgresStore
	logger *slog.Logger
	devToken string
	mu sync.RWMutex
	nodes map[string]protocol.Node
}

func New(logger *slog.Logger, devToken string) *Service {
	return &Service{logger: logger, devToken: devToken, nodes: make(map[string]protocol.Node)}
}
func NewWithPostgres(logger *slog.Logger, token string, store *PostgresStore) *Service {
	s := New(logger, token)
	s.postgres = store
	return s
}
func (s *Service) authenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.postgres != nil && (s.devToken == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+s.devToken)) != 1) {
			writeError(w,401,"authentication required"); return
		}
		next(w,r)
	}
}
func (s *Service) databaseError(w http.ResponseWriter, err error) {
	if errors.Is(err,sql.ErrNoRows) { writeError(w,404,"node is not enrolled"); return }
	if errors.Is(err,ErrNodeOwnership) { writeError(w,409,"node ownership conflict"); return }
	s.logger.Error("database operation failed") // Never log DSNs or credentials.
	writeError(w,503,"database unavailable")
}
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.health)
	mux.HandleFunc("POST /v1/enroll", s.enroll)
	mux.HandleFunc("POST /v1/heartbeat", s.authenticated(s.heartbeat))
	mux.HandleFunc("GET /v1/nodes", s.authenticated(s.listNodes))
	return middleware(mux)
}
func (s *Service) health(w http.ResponseWriter, r *http.Request) {
	if s.postgres != nil {
		ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second); defer cancel()
		if err:=s.postgres.Ping(ctx);err!=nil {s.databaseError(w,err);return}
	}
	writeJSON(w, 200, map[string]string{"service":"meshalot-control","status":"ok","protocol":protocol.Version})
}
func (s *Service) enroll(w http.ResponseWriter, r *http.Request) {
	var req protocol.EnrollRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || strings.TrimSpace(req.NodeID) == "" {
		writeError(w, 400, "invalid enrollment request"); return
	}
	if s.devToken == "" || req.EnrollmentToken != s.devToken {
		writeError(w, 401, "invalid enrollment token"); return
	}
	if s.postgres != nil {
		ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
		if err:=s.postgres.Enroll(ctx,req.NodeID,req.AgentVersion);err!=nil {s.databaseError(w,err);return}
		writeJSON(w,201,protocol.EnrollResponse{NodeID:req.NodeID,Accepted:true});return
	}
	node := protocol.Node{NodeID:req.NodeID, AgentVersion:req.AgentVersion, Status:"enrolled", Mode:"available"}
	s.mu.Lock(); s.nodes[node.NodeID] = node; s.mu.Unlock()
	s.logger.Info("node enrolled", "node_id", node.NodeID)
	writeJSON(w, 201, protocol.EnrollResponse{NodeID:node.NodeID, Accepted:true})
}
func (s *Service) heartbeat(w http.ResponseWriter, r *http.Request) {
	var req protocol.HeartbeatRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil || strings.TrimSpace(req.NodeID) == "" {
		writeError(w, 400, "invalid heartbeat"); return
	}
	if s.postgres != nil {
		ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
		if err:=s.postgres.Heartbeat(ctx,req.NodeID,req.Mode);err!=nil {s.databaseError(w,err);return}
		w.WriteHeader(204);return
	}
	s.mu.Lock()
	node, ok := s.nodes[req.NodeID]
	if ok { node.Status="online"; node.Mode=req.Mode; node.LastHeartbeat=req.ObservedAt.UTC(); s.nodes[req.NodeID]=node }
	s.mu.Unlock()
	if !ok { writeError(w, 404, "node is not enrolled"); return }
	w.WriteHeader(204)
}
func (s *Service) listNodes(w http.ResponseWriter, r *http.Request) {
	if s.postgres != nil {
		ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
		nodes,err:=s.postgres.Nodes(ctx);if err!=nil {s.databaseError(w,err);return}
		writeJSON(w,200,protocol.NodesResponse{Nodes:nodes});return
	}
	s.mu.RLock(); nodes := make([]protocol.Node,0,len(s.nodes))
	for _, n := range s.nodes { nodes=append(nodes,n) }; s.mu.RUnlock()
	writeJSON(w, 200, protocol.NodesResponse{Nodes:nodes})
}
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		w.Header().Set("Content-Type","application/json")
		w.Header().Set("Access-Control-Allow-Origin","http://localhost:5173")
		next.ServeHTTP(w,r)
	})
}
func writeError(w http.ResponseWriter,status int,message string){writeJSON(w,status,map[string]string{"error":message})}
func writeJSON(w http.ResponseWriter,status int,value any){w.WriteHeader(status);_ = json.NewEncoder(w).Encode(value)}
