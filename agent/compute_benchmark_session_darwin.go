//go:build darwin

package agent

import (
	"context"
	"errors"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// DarwinComputeBenchmarkExecutionOptions contains the execution settings
// required for one complete M9 Darwin benchmark session.
type DarwinComputeBenchmarkExecutionOptions struct {
	CollectedAt time.Time

	Condition string

	Quantization string

	LlamaBenchPath string

	LlamaServerPath string

	ModelPath string
}

type darwinComputeBenchmarkExecutionInput struct {
	CollectedAt time.Time

	Condition string

	Quantization string

	LlamaBenchPath string

	LlamaServerPath string

	ModelPath string

	ServerStarter darwinBenchmarkServerStarter

	HealthClient computeBenchmarkHTTPDoer

	TTFTClient computeBenchmarkHTTPDoer

	Now func() time.Time

	CollectRun darwinComputeBenchmarkRunCollector

	StopTimeout time.Duration
}

// ExecuteDarwinComputeBenchmark executes one complete real M9 Darwin session.
//
// Throughput and TTFT remain sequential on Darwin so that the large CPU model
// is not intentionally resident in two separate llama.cpp processes at once.
func ExecuteDarwinComputeBenchmark(
	ctx context.Context,
	options DarwinComputeBenchmarkExecutionOptions,
) (
	protocol.ComputeBenchmark,
	error,
) {
	return executeDarwinComputeBenchmark(
		ctx,
		darwinComputeBenchmarkExecutionInput{
			CollectedAt: options.CollectedAt,

			Condition: options.Condition,

			Quantization: options.Quantization,

			LlamaBenchPath: options.LlamaBenchPath,

			LlamaServerPath: options.LlamaServerPath,

			ModelPath: options.ModelPath,
		},
	)
}

func executeDarwinComputeBenchmark(
	ctx context.Context,
	input darwinComputeBenchmarkExecutionInput,
) (
	protocol.ComputeBenchmark,
	error,
) {
	var benchmark protocol.ComputeBenchmark

	if ctx == nil {
		return benchmark,
			errors.New(
				"compute benchmark context is nil",
			)
	}

	if input.CollectedAt.IsZero() {
		return benchmark,
			errors.New(
				"compute benchmark collection time is required",
			)
	}

	if strings.TrimSpace(
		input.Condition,
	) == "" {

		return benchmark,
			errors.New(
				"compute benchmark condition is required",
			)
	}

	if strings.TrimSpace(
		input.Quantization,
	) == "" {

		return benchmark,
			errors.New(
				"compute benchmark quantization is required",
			)
	}

	if strings.TrimSpace(
		input.LlamaBenchPath,
	) == "" {

		return benchmark,
			errors.New(
				"llama-bench path is required",
			)
	}

	if _, err :=
		m9LlamaBenchArgs(
			input.ModelPath,
		); err != nil {

		return benchmark,
			err
	}

	// Validate the TTFT server command before any throughput process starts.
	if _, _, _, err :=
		darwinBenchmarkServerCommand(
			input.LlamaServerPath,
			input.ModelPath,
		); err != nil {

		return benchmark,
			err
	}

	return orchestrateDarwinComputeBenchmark(
		ctx,
		darwinComputeBenchmarkOrchestratorInput{
			CollectedAt: input.CollectedAt,

			Condition: input.Condition,

			Quantization: input.Quantization,

			LlamaBenchPath: input.LlamaBenchPath,

			LlamaServerPath: input.LlamaServerPath,

			ModelPath: input.ModelPath,

			ServerStarter: input.ServerStarter,

			HealthClient: input.HealthClient,

			TTFTClient: input.TTFTClient,

			Now: input.Now,

			CollectRun: input.CollectRun,

			StopTimeout: input.StopTimeout,
		},
	)
}
