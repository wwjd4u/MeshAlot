package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	NetworkBenchmarkSubmissionPath    = "/v1/agent/network-benchmark"
	NetworkBenchmarkSignatureHeader   = "X-MeshAlot-Signature"
	NetworkBenchmarkSigningDomain     = "MESHALOT-M8-NETWORK-BENCHMARK"
	NetworkBenchmarkSubmissionVersion = "m8-v1"
)

type NetworkBenchmarkSubmission struct {
	Version   string           `json:"version"`
	NodeID    string           `json:"node_id"`
	ReportID  string           `json:"report_id"`
	SignedAt  time.Time        `json:"signed_at"`
	Benchmark NetworkBenchmark `json:"benchmark"`
}

type NetworkBenchmarkSubmissionResponse struct {
	ReportID string `json:"report_id"`
	Accepted bool   `json:"accepted"`
}

func NetworkBenchmarkSigningMessage(body []byte) []byte {
	sum := sha256.Sum256(body)

	value := NetworkBenchmarkSigningDomain +
		"\nPOST\n" +
		NetworkBenchmarkSubmissionPath +
		"\n" +
		hex.EncodeToString(sum[:])

	return []byte(value)
}
