package v1

import "time"

const ComputeBenchmarkSchemaVersion = "m9-v1"

// ComputeBenchmark contains measured AI performance.
//
// The benchmark contains raw measurements only. The authoritative Compute
// Score is calculated by the control server after authentication.
type ComputeBenchmark struct {
	SchemaVersion  string                   `json:"schema_version"`
	CollectedAt    time.Time                `json:"collected_at"`
	WorkloadID     string                   `json:"workload_id"`
	Condition      string                   `json:"condition"`
	Runtime        string                   `json:"runtime"`
	RuntimeVersion string                   `json:"runtime_version,omitempty"`
	Model          string                   `json:"model"`
	Quantization   string                   `json:"quantization,omitempty"`
	ContextSize    int                      `json:"context_size"`
	Samples        []ComputeBenchmarkSample `json:"samples"`
}

// ComputeBenchmarkSample preserves one repetition of the benchmark workload.
//
// Keeping the individual samples allows variance, instability, and abnormal
// runs to remain visible instead of storing only an average.
type ComputeBenchmarkSample struct {
	Run                       int      `json:"run"`
	PromptTokens              int      `json:"prompt_tokens"`
	GeneratedTokens           int      `json:"generated_tokens"`
	PromptTokensPerSecond     float64  `json:"prompt_tokens_per_second"`
	GenerationTokensPerSecond float64  `json:"generation_tokens_per_second"`
	TimeToFirstTokenMs        float64  `json:"time_to_first_token_ms"`
	PeakSystemRAMBytes        uint64   `json:"peak_system_ram_bytes"`
	PeakVRAMBytes             *uint64  `json:"peak_vram_bytes,omitempty"`
	TemperatureCelsius        *float64 `json:"temperature_celsius,omitempty"`
	Throttled                 *bool    `json:"throttled,omitempty"`
	Success                   bool     `json:"success"`
	Error                     string   `json:"error,omitempty"`
}
