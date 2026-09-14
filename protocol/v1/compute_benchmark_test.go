package v1

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestComputeBenchmarkProtocolConstants(t *testing.T) {
	if ComputeBenchmarkSchemaVersion != "m9-v1" {
		t.Fatalf(
			"ComputeBenchmarkSchemaVersion = %q, want m9-v1",
			ComputeBenchmarkSchemaVersion,
		)
	}

	if ComputeBenchmarkSubmissionVersion != "m9-v1" {
		t.Fatalf(
			"ComputeBenchmarkSubmissionVersion = %q, want m9-v1",
			ComputeBenchmarkSubmissionVersion,
		)
	}

	if ComputeBenchmarkSubmissionPath != "/v1/agent/compute-benchmark" {
		t.Fatalf(
			"ComputeBenchmarkSubmissionPath = %q",
			ComputeBenchmarkSubmissionPath,
		)
	}

	if ComputeBenchmarkSigningDomain != "MESHALOT-M9-COMPUTE-BENCHMARK" {
		t.Fatalf(
			"ComputeBenchmarkSigningDomain = %q",
			ComputeBenchmarkSigningDomain,
		)
	}
}

func TestComputeBenchmarkJSONPreservesSamples(t *testing.T) {
	vram := uint64(8 * 1024 * 1024 * 1024)
	temp := 72.5
	throttled := false

	benchmark := ComputeBenchmark{
		SchemaVersion:  ComputeBenchmarkSchemaVersion,
		CollectedAt:    time.Date(2026, 9, 14, 18, 0, 0, 0, time.UTC),
		WorkloadID:     "m9-standard-v1",
		Condition:      "baseline",
		Runtime:        "llama.cpp",
		RuntimeVersion: "test-version",
		Model:          "test-model",
		Quantization:   "Q4_K_M",
		ContextSize:    4096,
		Samples: []ComputeBenchmarkSample{
			{
				Run:                       1,
				PromptTokens:              512,
				GeneratedTokens:           128,
				PromptTokensPerSecond:     120.25,
				GenerationTokensPerSecond: 31.75,
				TimeToFirstTokenMs:        185.5,
				PeakSystemRAMBytes:        12 * 1024 * 1024 * 1024,
				PeakVRAMBytes:             &vram,
				TemperatureCelsius:        &temp,
				Throttled:                 &throttled,
				Success:                   true,
			},
			{
				Run:                       2,
				PromptTokens:              512,
				GeneratedTokens:           128,
				PromptTokensPerSecond:     118.0,
				GenerationTokensPerSecond: 30.9,
				TimeToFirstTokenMs:        190.0,
				PeakSystemRAMBytes:        12 * 1024 * 1024 * 1024,
				Success:                   true,
			},
		},
	}

	body, err := json.Marshal(benchmark)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ComputeBenchmark
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.SchemaVersion != ComputeBenchmarkSchemaVersion {
		t.Fatalf(
			"schema version = %q",
			decoded.SchemaVersion,
		)
	}

	if decoded.WorkloadID != "m9-standard-v1" {
		t.Fatalf(
			"workload id = %q",
			decoded.WorkloadID,
		)
	}

	if decoded.Condition != "baseline" {
		t.Fatalf(
			"condition = %q",
			decoded.Condition,
		)
	}

	if decoded.Model != "test-model" ||
		decoded.Quantization != "Q4_K_M" ||
		decoded.ContextSize != 4096 {
		t.Fatalf("model metadata changed: %#v", decoded)
	}

	if len(decoded.Samples) != 2 {
		t.Fatalf(
			"sample count = %d, want 2",
			len(decoded.Samples),
		)
	}

	if decoded.Samples[0].PeakVRAMBytes == nil ||
		*decoded.Samples[0].PeakVRAMBytes != vram {
		t.Fatalf(
			"VRAM measurement changed: %#v",
			decoded.Samples[0].PeakVRAMBytes,
		)
	}

	if decoded.Samples[0].TemperatureCelsius == nil ||
		*decoded.Samples[0].TemperatureCelsius != temp {
		t.Fatalf(
			"temperature measurement changed: %#v",
			decoded.Samples[0].TemperatureCelsius,
		)
	}

	if decoded.Samples[0].Throttled == nil ||
		*decoded.Samples[0].Throttled {
		t.Fatalf(
			"throttling measurement changed: %#v",
			decoded.Samples[0].Throttled,
		)
	}
}

func TestComputeBenchmarkSigningDomainIsSeparated(t *testing.T) {
	message := string(
		ComputeBenchmarkSigningMessage(
			[]byte(`{"benchmark":"test"}`),
		),
	)

	if !strings.HasPrefix(
		message,
		"MESHALOT-M9-COMPUTE-BENCHMARK\nPOST\n/v1/agent/compute-benchmark\n",
	) {
		t.Fatalf(
			"unexpected compute benchmark signing message: %q",
			message,
		)
	}

	if strings.Contains(message, "MESHALOT-M8-NETWORK-BENCHMARK") {
		t.Fatal("M9 signing domain overlaps M8 signing domain")
	}
}
