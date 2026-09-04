package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func (s *Service) dashboard(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	result, err := s.postgres.Dashboard(ctx, user.ID)
	if err != nil {
		s.databaseError(w, err)
		return
	}
	writeJSON(w, 200, result)
}

func (s *Service) accountNodes(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	nodes, err := s.postgres.NodesForUser(ctx, user.ID)
	if err != nil {
		s.databaseError(w, err)
		return
	}
	writeJSON(w, 200, protocol.NodesResponse{Nodes: nodes})
}

func (s *Service) accountNode(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	node, err := s.postgres.NodeForUser(ctx, user.ID, r.PathValue("nodeID"))
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, 404, "node not found")
		return
	}
	if err != nil {
		s.databaseError(w, err)
		return
	}
	writeJSON(w, 200, node)
}

func (s *Service) wallet(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	result, err := s.postgres.Wallet(ctx, user.ID)
	if err != nil {
		s.databaseError(w, err)
		return
	}
	writeJSON(w, 200, result)
}

func (s *Service) jobs(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	result, err := s.postgres.Jobs(ctx, user.ID)
	if err != nil {
		s.databaseError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
