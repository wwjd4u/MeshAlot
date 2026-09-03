package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type Service struct {
	logger *slog.Logger
	devToken string
	mu sync.RWMutex
	nodes map[string]protocol.Node
}

func New(logger *slog.Logger, devToken string) *Service {
	return &Service{logger: logger, devToken: devToken, nodes: make(map[string]protocol.Node)}
}
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.health)
	mux.HandleFunc("POST /v1/enroll", s.enroll)
	mux.HandleFunc("POST /v1/heartbeat", s.heartbeat)
	mux.HandleFunc("GET /v1/nodes", s.listNodes)
	return middleware(mux)
}
func (s *Service) health(w http.ResponseWriter, _ *http.Request) {
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
	s.mu.Lock()
	node, ok := s.nodes[req.NodeID]
	if ok { node.Status="online"; node.Mode=req.Mode; node.LastHeartbeat=req.ObservedAt.UTC(); s.nodes[req.NodeID]=node }
	s.mu.Unlock()
	if !ok { writeError(w, 404, "node is not enrolled"); return }
	w.WriteHeader(204)
}
func (s *Service) listNodes(w http.ResponseWriter, _ *http.Request) {
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
