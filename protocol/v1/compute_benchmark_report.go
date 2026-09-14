package v1

import "time"

type ComputeBenchmarkReport struct {
	ReportID   string           `json:"report_id"`
	NodeID     string           `json:"node_id"`
	ReceivedAt time.Time        `json:"received_at"`
	Score      float64          `json:"score"`
	Benchmark  ComputeBenchmark `json:"benchmark"`
}
