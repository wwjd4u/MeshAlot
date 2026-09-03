package v1

import "time"

const Version = "v1"

type HealthResponse struct {
	Service, Status, Protocol string
}

type EnrollRequest struct {
	NodeID          string `json:"node_id"`
	EnrollmentToken string `json:"enrollment_token"`
	AgentVersion    string `json:"agent_version"`
}
type EnrollResponse struct {
	NodeID string `json:"node_id"`
	Accepted bool `json:"accepted"`
}
type HeartbeatRequest struct {
	NodeID string `json:"node_id"`
	ObservedAt time.Time `json:"observed_at"`
	Mode string `json:"mode"`
}
type Node struct {
	NodeID string `json:"node_id"`
	AgentVersion string `json:"agent_version"`
	Status string `json:"status"`
	Mode string `json:"mode"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}
type NodesResponse struct { Nodes []Node `json:"nodes"` }
