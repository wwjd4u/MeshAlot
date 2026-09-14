package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type computeBenchmarkThroughputCollector func(
	context.Context,
	int,
) (
	llamaBenchThroughput,
	computeBenchmarkRunTelemetry,
	error,
)

type computeBenchmarkTTFTCollector func(
	context.Context,
	int,
) (
	time.Duration,
	error,
)

type computeBenchmarkOrchestratorInput struct {
	CollectedAt  time.Time
	Condition    string
	Quantization string
	ModelPath    string

	CollectThroughput computeBenchmarkThroughputCollector
	CollectTTFT       computeBenchmarkTTFTCollector
}

// orchestrateComputeBenchmarkRuns performs the five independent M9
// repetitions.
//
// Throughput/resource measurement and TTFT measurement are attempted
// independently for every repetition. A failure in one repetition does not
// discard the remaining repetitions. Instead the failed attempt is retained
// for the authoritative server-side stability and success scoring.
func orchestrateComputeBenchmarkRuns(
	ctx context.Context,
	input computeBenchmarkOrchestratorInput,
) (protocol.ComputeBenchmark, error) {
	var benchmark protocol.ComputeBenchmark

	if ctx == nil {
		return benchmark, errors.New(
			"compute benchmark context is nil",
		)
	}

	if input.CollectThroughput == nil {
		return benchmark, errors.New(
			"compute benchmark throughput collector is nil",
		)
	}

	if input.CollectTTFT == nil {
		return benchmark, errors.New(
			"compute benchmark TTFT collector is nil",
		)
	}

	runs := make(
		[]computeBenchmarkRunMeasurement,
		m9StandardRepetitions,
	)

	for i := 0; i < m9StandardRepetitions; i++ {
		if err := ctx.Err(); err != nil {
			return benchmark, fmt.Errorf(
				"compute benchmark canceled before run %d: %w",
				i+1,
				err,
			)
		}

		runNumber := i + 1

		throughput,
			telemetry,
			throughputErr :=
			input.CollectThroughput(
				ctx,
				runNumber,
			)

		if err := ctx.Err(); err != nil {
			return benchmark, fmt.Errorf(
				"compute benchmark canceled during run %d: %w",
				runNumber,
				err,
			)
		}

		ttft,
			ttftErr :=
			input.CollectTTFT(
				ctx,
				runNumber,
			)

		if err := ctx.Err(); err != nil {
			return benchmark, fmt.Errorf(
				"compute benchmark canceled during TTFT run %d: %w",
				runNumber,
				err,
			)
		}

		if ttft > 0 {
			telemetry.TimeToFirstToken =
				ttft
		}

		runs[i] =
			computeBenchmarkRunMeasurement{
				Throughput: throughput,

				Telemetry: telemetry,

				Error: computeBenchmarkRunFailure(
					throughputErr,
					ttftErr,
				),
			}
	}

	return finalizeComputeBenchmarkRuns(
		computeBenchmarkFinalizeInput{
			CollectedAt: input.CollectedAt,

			Condition: input.Condition,

			Quantization: input.Quantization,

			ModelPath: input.ModelPath,

			Runs: runs,
		},
	)
}

func computeBenchmarkRunFailure(
	throughputErr error,
	ttftErr error,
) string {
	var failures []string

	if throughputErr != nil {
		text := strings.TrimSpace(
			throughputErr.Error(),
		)

		if text == "" {
			text = "unknown throughput error"
		}

		failures = append(
			failures,
			"throughput: "+text,
		)
	}

	if ttftErr != nil {
		text := strings.TrimSpace(
			ttftErr.Error(),
		)

		if text == "" {
			text = "unknown TTFT error"
		}

		failures = append(
			failures,
			"ttft: "+text,
		)
	}

	return strings.Join(
		failures,
		"; ",
	)
}
