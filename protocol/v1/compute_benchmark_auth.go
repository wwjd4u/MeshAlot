package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	ComputeBenchmarkSubmissionPath    = "/v1/agent/compute-benchmark"
	ComputeBenchmarkSignatureHeader   = "X-MeshAlot-Signature"
	ComputeBenchmarkSigningDomain     = "MESHALOT-M9-COMPUTE-BENCHMARK"
	ComputeBenchmarkSubmissionVersion = "m9-v1"
)

type ComputeBenchmarkSubmission struct {
	Version   string           `json:"version"`
	NodeID    string           `json:"node_id"`
	ReportID  string           `json:"report_id"`
	SignedAt  time.Time        `json:"signed_at"`
	Benchmark ComputeBenchmark `json:"benchmark"`
}

type ComputeBenchmarkSubmissionResponse struct {
	ReportID string `json:"report_id"`
	Accepted bool   `json:"accepted"`
}

func ComputeBenchmarkSigningMessage(body []byte) []byte {
	sum := sha256.Sum256(body)

	value := ComputeBenchmarkSigningDomain +
		"\nPOST\n" +
		ComputeBenchmarkSubmissionPath +
		"\n" +
		hex.EncodeToString(sum[:])

	return []byte(value)
}
