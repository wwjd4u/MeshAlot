package agent

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const m9ComputeBenchmarkRunErrorMax = 1024

type computeBenchmarkFinalizeInput struct {
	CollectedAt  time.Time
	Condition    string
	Quantization string
	ModelPath    string

	Runs []computeBenchmarkRunMeasurement
}

// finalizeComputeBenchmarkRuns converts five independently attempted M9 runs
// into the protocol payload.
//
// Successful runs require complete throughput, TTFT, and RAM measurements.
// Failed runs remain in the payload with Success=false and an explanatory
// error. This allows the authoritative Compute Score to penalize instability
// instead of silently discarding failed repetitions.
func finalizeComputeBenchmarkRuns(
	input computeBenchmarkFinalizeInput,
) (protocol.ComputeBenchmark, error) {
	var benchmark protocol.ComputeBenchmark

	if input.CollectedAt.IsZero() {
		return benchmark, errors.New(
			"compute benchmark collection time is required",
		)
	}

	condition := strings.TrimSpace(
		input.Condition,
	)

	if condition == "" {
		return benchmark, errors.New(
			"compute benchmark condition is required",
		)
	}

	modelPath := strings.TrimSpace(
		input.ModelPath,
	)

	if modelPath == "" {
		return benchmark, errors.New(
			"compute benchmark model path is required",
		)
	}

	modelName := filepath.Base(modelPath)

	if modelName == "" ||
		modelName == "." ||
		modelName == string(filepath.Separator) {
		return benchmark, errors.New(
			"compute benchmark model name is invalid",
		)
	}

	if len(input.Runs) !=
		m9StandardRepetitions {
		return benchmark, fmt.Errorf(
			"compute benchmark run count = %d, want %d",
			len(input.Runs),
			m9StandardRepetitions,
		)
	}

	benchmark = protocol.ComputeBenchmark{
		SchemaVersion: protocol.ComputeBenchmarkSchemaVersion,

		CollectedAt: input.CollectedAt.UTC(),

		WorkloadID: m9StandardWorkloadID,

		Condition: condition,

		Runtime: "llama.cpp",

		Model: modelName,

		Quantization: strings.TrimSpace(
			input.Quantization,
		),

		ContextSize: m9StandardContextSize,

		Samples: make(
			[]protocol.ComputeBenchmarkSample,
			0,
			m9StandardRepetitions,
		),
	}

	var (
		referenceThroughput llamaBenchThroughput
		haveReference       bool
	)

	for i, run := range input.Runs {
		runNumber := i + 1

		errorText := strings.TrimSpace(
			run.Error,
		)

		if errorText != "" {
			if len(errorText) >
				m9ComputeBenchmarkRunErrorMax {
				return protocol.ComputeBenchmark{},
					fmt.Errorf(
						"compute benchmark run %d error text is too long",
						runNumber,
					)
			}

			if run.Telemetry.TimeToFirstToken < 0 {
				return protocol.ComputeBenchmark{},
					fmt.Errorf(
						"compute benchmark run %d has invalid partial TTFT",
						runNumber,
					)
			}

			sample :=
				protocol.ComputeBenchmarkSample{
					Run: runNumber,

					PromptTokens: m9StandardPromptTokens,

					GeneratedTokens: m9StandardGeneratedTokens,

					PeakSystemRAMBytes: run.Telemetry.PeakSystemRAM,

					PeakVRAMBytes: run.Telemetry.PeakVRAMBytes,

					TemperatureCelsius: run.Telemetry.
						TemperatureCelsius,

					Throttled: run.Telemetry.Throttled,

					Success: false,

					Error: errorText,
				}

			if run.Telemetry.TimeToFirstToken > 0 {
				sample.TimeToFirstTokenMs =
					float64(
						run.Telemetry.
							TimeToFirstToken,
					) /
						float64(
							time.Millisecond,
						)
			}

			benchmark.Samples = append(
				benchmark.Samples,
				sample,
			)

			continue
		}

		current := run.Throughput

		if len(
			current.
				PromptTokensPerSecond,
		) !=
			m9LlamaBenchInvocationRepetitions ||
			len(
				current.
					GenerationTokensPerSecond,
			) !=
				m9LlamaBenchInvocationRepetitions {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d does not contain exactly one throughput repetition",
					runNumber,
				)
		}

		if current.PromptTokens <= 0 ||
			current.GeneratedTokens <= 0 {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d has invalid token counts",
					runNumber,
				)
		}

		if strings.TrimSpace(
			current.RuntimeVersion,
		) == "" ||
			strings.TrimSpace(
				current.ModelFilename,
			) == "" ||
			strings.TrimSpace(
				current.ModelType,
			) == "" ||
			strings.TrimSpace(
				current.Backend,
			) == "" {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d has incomplete runtime metadata",
					runNumber,
				)
		}

		if filepath.Base(
			current.ModelFilename,
		) != modelName {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d used an unexpected model",
					runNumber,
				)
		}

		if err := validateLlamaBenchSamples(
			current.PromptTokensPerSecond,
		); err != nil {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d prompt throughput: %w",
					runNumber,
					err,
				)
		}

		if err := validateLlamaBenchSamples(
			current.GenerationTokensPerSecond,
		); err != nil {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d generation throughput: %w",
					runNumber,
					err,
				)
		}

		if run.Telemetry.TimeToFirstToken <= 0 {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d has invalid TTFT",
					runNumber,
				)
		}

		if run.Telemetry.PeakSystemRAM == 0 {
			return protocol.ComputeBenchmark{},
				fmt.Errorf(
					"compute benchmark run %d has invalid peak RAM",
					runNumber,
				)
		}

		if !haveReference {
			referenceThroughput = current
			haveReference = true

			benchmark.RuntimeVersion =
				current.RuntimeVersion
		} else if current.RuntimeVersion !=
			referenceThroughput.RuntimeVersion ||
			current.ModelFilename !=
				referenceThroughput.
					ModelFilename ||
			current.ModelType !=
				referenceThroughput.ModelType ||
			current.Backend !=
				referenceThroughput.Backend ||
			current.PromptTokens !=
				referenceThroughput.
					PromptTokens ||
			current.GeneratedTokens !=
				referenceThroughput.
					GeneratedTokens {
			return protocol.ComputeBenchmark{},
				errors.New(
					"compute benchmark run metadata changed between successful repetitions",
				)
		}

		benchmark.Samples = append(
			benchmark.Samples,
			protocol.ComputeBenchmarkSample{
				Run: runNumber,

				PromptTokens: current.PromptTokens,

				GeneratedTokens: current.GeneratedTokens,

				PromptTokensPerSecond: current.
					PromptTokensPerSecond[0],

				GenerationTokensPerSecond: current.
					GenerationTokensPerSecond[0],

				TimeToFirstTokenMs: float64(
					run.Telemetry.
						TimeToFirstToken,
				) /
					float64(
						time.Millisecond,
					),

				PeakSystemRAMBytes: run.Telemetry.
					PeakSystemRAM,

				PeakVRAMBytes: run.Telemetry.
					PeakVRAMBytes,

				TemperatureCelsius: run.Telemetry.
					TemperatureCelsius,

				Throttled: run.Telemetry.Throttled,

				Success: true,
			},
		)
	}

	return benchmark, nil
}
