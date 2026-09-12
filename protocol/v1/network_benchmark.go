package v1

import "time"

const NetworkBenchmarkSchemaVersion = "m8-v1"

type NetworkBenchmark struct {
	SchemaVersion     string    `json:"schema_version"`
	CollectedAt       time.Time `json:"collected_at"`
	IPv4Available     bool      `json:"ipv4_available"`
	IPv6Available     bool      `json:"ipv6_available"`
	DownloadMbps      float64   `json:"download_mbps"`
	UploadMbps        float64   `json:"upload_mbps"`
	LatencyMs         float64   `json:"latency_ms"`
	JitterMs          float64   `json:"jitter_ms"`
	PacketLossPercent float64   `json:"packet_loss_percent"`
	PreferredPath     string    `json:"preferred_path"`
}
