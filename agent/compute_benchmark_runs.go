package agent

import (
	"errors"
	"fmt"
)

type computeBenchmarkRunMeasurement struct {
	Throughput llamaBenchThroughput
	Telemetry  computeBenchmarkRunTelemetry

	// Error is populated when this repetition failed. Failed runs are retained
	// so the authoritative server score can account for benchmark instability.
	Error string
}

// combineComputeBenchmarkRuns combines exactly five independently measured
// M9 repetitions.
//
// Each input must come from one llama-bench process using -r 1 and must carry
// its own TTFT, RAM, and available thermal/resource telemetry.
//
// Runtime/model/backend metadata must remain identical across all five runs so
// that the resulting benchmark represents one comparable workload.
func combineComputeBenchmarkRuns(
	runs []computeBenchmarkRunMeasurement,
) (
	llamaBenchThroughput,
	[]computeBenchmarkRunTelemetry,
	error,
) {
	var combined llamaBenchThroughput

	if len(runs) != m9StandardRepetitions {
		return combined, nil, fmt.Errorf(
			"compute benchmark run count = %d, want %d",
			len(runs),
			m9StandardRepetitions,
		)
	}

	telemetry := make(
		[]computeBenchmarkRunTelemetry,
		0,
		m9StandardRepetitions,
	)

	for i, run := range runs {
		current := run.Throughput

		if len(current.PromptTokensPerSecond) !=
			m9LlamaBenchInvocationRepetitions ||
			len(current.GenerationTokensPerSecond) !=
				m9LlamaBenchInvocationRepetitions {
			return combined, nil, fmt.Errorf(
				"compute benchmark run %d does not contain exactly one throughput repetition",
				i+1,
			)
		}

		if current.PromptTokens <= 0 ||
			current.GeneratedTokens <= 0 {
			return combined, nil, fmt.Errorf(
				"compute benchmark run %d has invalid token counts",
				i+1,
			)
		}

		if current.RuntimeVersion == "" ||
			current.ModelFilename == "" ||
			current.ModelType == "" ||
			current.Backend == "" {
			return combined, nil, fmt.Errorf(
				"compute benchmark run %d has incomplete runtime metadata",
				i+1,
			)
		}

		if err := validateLlamaBenchSamples(
			current.PromptTokensPerSecond,
		); err != nil {
			return combined, nil, fmt.Errorf(
				"compute benchmark run %d prompt throughput: %w",
				i+1,
				err,
			)
		}

		if err := validateLlamaBenchSamples(
			current.GenerationTokensPerSecond,
		); err != nil {
			return combined, nil, fmt.Errorf(
				"compute benchmark run %d generation throughput: %w",
				i+1,
				err,
			)
		}

		if run.Telemetry.TimeToFirstToken <= 0 {
			return combined, nil, fmt.Errorf(
				"compute benchmark run %d has invalid TTFT",
				i+1,
			)
		}

		if run.Telemetry.PeakSystemRAM == 0 {
			return combined, nil, fmt.Errorf(
				"compute benchmark run %d has invalid peak RAM",
				i+1,
			)
		}

		if i == 0 {
			combined = llamaBenchThroughput{
				RuntimeVersion: current.RuntimeVersion,

				ModelFilename: current.ModelFilename,

				ModelType: current.ModelType,

				Backend: current.Backend,

				PromptTokens: current.PromptTokens,

				GeneratedTokens: current.GeneratedTokens,
			}
		} else if current.RuntimeVersion !=
			combined.RuntimeVersion ||
			current.ModelFilename !=
				combined.ModelFilename ||
			current.ModelType !=
				combined.ModelType ||
			current.Backend !=
				combined.Backend ||
			current.PromptTokens !=
				combined.PromptTokens ||
			current.GeneratedTokens !=
				combined.GeneratedTokens {
			return llamaBenchThroughput{},
				nil,
				errors.New(
					"compute benchmark run metadata changed between repetitions",
				)
		}

		combined.PromptTokensPerSecond =
			append(
				combined.PromptTokensPerSecond,
				current.PromptTokensPerSecond[0],
			)

		combined.GenerationTokensPerSecond =
			append(
				combined.GenerationTokensPerSecond,
				current.GenerationTokensPerSecond[0],
			)

		telemetry = append(
			telemetry,
			run.Telemetry,
		)
	}

	return combined, telemetry, nil
}
