package agent

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type computeBenchmarkRunTelemetry struct {
	TimeToFirstToken time.Duration
	PeakSystemRAM    uint64

	PeakVRAMBytes      *uint64
	TemperatureCelsius *float64
	Throttled          *bool
}

type computeBenchmarkAssemblyInput struct {
	CollectedAt  time.Time
	Condition    string
	Quantization string

	Throughput llamaBenchThroughput
	Telemetry  []computeBenchmarkRunTelemetry
}

// assembleComputeBenchmark combines independently measured M9 data.
//
// Throughput, TTFT, RAM, VRAM, thermal, and throttling measurements remain
// independently collected. This function only combines matching repetitions;
// it does not invent missing measurements.
func assembleComputeBenchmark(
	input computeBenchmarkAssemblyInput,
) (protocol.ComputeBenchmark, error) {
	var benchmark protocol.ComputeBenchmark

	if input.CollectedAt.IsZero() {
		return benchmark, errors.New(
			"compute benchmark collection time is required",
		)
	}

	condition := strings.TrimSpace(input.Condition)
	if condition == "" {
		return benchmark, errors.New(
			"compute benchmark condition is required",
		)
	}

	throughput := input.Throughput

	if throughput.PromptTokens <= 0 ||
		throughput.GeneratedTokens <= 0 {
		return benchmark, errors.New(
			"compute benchmark token workload is invalid",
		)
	}

	runCount := len(
		throughput.PromptTokensPerSecond,
	)

	if runCount == 0 ||
		len(throughput.GenerationTokensPerSecond) != runCount ||
		len(input.Telemetry) != runCount {
		return benchmark, errors.New(
			"compute benchmark repetition counts do not match",
		)
	}

	if runCount != m9StandardRepetitions {
		return benchmark, fmt.Errorf(
			"compute benchmark repetition count = %d, want %d",
			runCount,
			m9StandardRepetitions,
		)
	}

	model := strings.TrimSpace(
		throughput.ModelFilename,
	)

	if model == "" {
		return benchmark, errors.New(
			"compute benchmark model is missing",
		)
	}

	runtimeVersion := strings.TrimSpace(
		throughput.RuntimeVersion,
	)

	if runtimeVersion == "" {
		return benchmark, errors.New(
			"compute benchmark runtime version is missing",
		)
	}

	samples := make(
		[]protocol.ComputeBenchmarkSample,
		0,
		runCount,
	)

	for i := 0; i < runCount; i++ {
		promptTPS :=
			throughput.PromptTokensPerSecond[i]

		generationTPS :=
			throughput.GenerationTokensPerSecond[i]

		telemetry := input.Telemetry[i]

		if !finiteNonNegative(promptTPS) ||
			promptTPS <= 0 ||
			!finiteNonNegative(generationTPS) ||
			generationTPS <= 0 {
			return benchmark, errors.New(
				"compute benchmark throughput sample is invalid",
			)
		}

		if telemetry.TimeToFirstToken <= 0 {
			return benchmark, errors.New(
				"compute benchmark TTFT sample is invalid",
			)
		}

		if telemetry.PeakSystemRAM == 0 {
			return benchmark, errors.New(
				"compute benchmark RAM sample is invalid",
			)
		}

		samples = append(
			samples,
			protocol.ComputeBenchmarkSample{
				Run: i + 1,

				PromptTokens: throughput.PromptTokens,

				GeneratedTokens: throughput.GeneratedTokens,

				PromptTokensPerSecond: promptTPS,

				GenerationTokensPerSecond: generationTPS,

				TimeToFirstTokenMs: float64(
					telemetry.TimeToFirstToken,
				) /
					float64(time.Millisecond),

				PeakSystemRAMBytes: telemetry.PeakSystemRAM,

				PeakVRAMBytes: telemetry.PeakVRAMBytes,

				TemperatureCelsius: telemetry.TemperatureCelsius,

				Throttled: telemetry.Throttled,

				Success: true,
			},
		)
	}

	return protocol.ComputeBenchmark{
		SchemaVersion: protocol.ComputeBenchmarkSchemaVersion,

		CollectedAt: input.CollectedAt.UTC(),

		WorkloadID: m9StandardWorkloadID,

		Condition: condition,

		Runtime: "llama.cpp",

		RuntimeVersion: runtimeVersion,

		Model: filepath.Base(model),

		Quantization: strings.TrimSpace(
			input.Quantization,
		),

		ContextSize: m9StandardContextSize,

		Samples: samples,
	}, nil
}
