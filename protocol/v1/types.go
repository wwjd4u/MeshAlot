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
	NodeID   string `json:"node_id"`
	Accepted bool   `json:"accepted"`
}
type HeartbeatRequest struct {
	NodeID     string    `json:"node_id"`
	ObservedAt time.Time `json:"observed_at"`
	Mode       string    `json:"mode"`
}
type Node struct {
	NodeID        string    `json:"node_id"`
	AgentVersion  string    `json:"agent_version"`
	Status        string    `json:"status"`
	Mode          string    `json:"mode"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}
type NodesResponse struct {
	Nodes []Node `json:"nodes"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SessionResponse struct {
	User User `json:"user"`
}

type DashboardResponse struct {
	BalanceMicrounits       int64    `json:"balance_microunits"`
	OnlineNodes             int      `json:"online_nodes"`
	ComputeScore            *float64 `json:"compute_score"`
	NetworkScore            *float64 `json:"network_score"`
	CurrentStatus           string   `json:"current_status"`
	CurrentRateMicrounits   *int64   `json:"current_rate_microunits"`
	TodayEarningsMicrounits int64    `json:"today_earnings_microunits"`
}

type WalletTransaction struct {
	ID               string    `json:"id"`
	JobID            string    `json:"job_id,omitempty"`
	TransactionType  string    `json:"transaction_type"`
	AmountMicrounits int64     `json:"amount_microunits"`
	CreatedAt        time.Time `json:"created_at"`
}

type WalletResponse struct {
	BalanceMicrounits int64               `json:"balance_microunits"`
	Transactions      []WalletTransaction `json:"transactions"`
}

type JobSummary struct {
	ID             string    `json:"id"`
	ProviderNodeID string    `json:"provider_node_id,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type JobsResponse struct {
	Jobs []JobSummary `json:"jobs"`
}
