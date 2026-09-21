//go:build linux

package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// LinuxComputeBenchmarkExecutionOptions contains only the real execution
// settings required to run one complete M9 Linux benchmark session.
type LinuxComputeBenchmarkExecutionOptions struct {
	CollectedAt time.Time

	Condition string

	Quantization string

	LlamaBenchPath string

	LlamaServerPath string

	ModelPath string
}

type linuxComputeBenchmarkExecutionInput struct {
	CollectedAt time.Time

	Condition string

	Quantization string

	LlamaBenchPath string

	LlamaServerPath string

	ModelPath string

	ServerStarter linuxBenchmarkServerStarter

	HealthClient computeBenchmarkHTTPDoer

	TTFTClient computeBenchmarkHTTPDoer

	Now func() time.Time

	CollectRun linuxComputeBenchmarkRunCollector

	StopTimeout time.Duration
}

// ExecuteLinuxComputeBenchmark executes one complete real M9 Linux benchmark
// session.
//
// It starts exactly one benchmark-owned loopback llama-server, executes the
// standardized five-run M9 workload, and stops only that server before
// returning.
func ExecuteLinuxComputeBenchmark(
	ctx context.Context,
	options LinuxComputeBenchmarkExecutionOptions,
) (
	protocol.ComputeBenchmark,
	error,
) {
	return executeLinuxComputeBenchmark(
		ctx,
		linuxComputeBenchmarkExecutionInput{
			CollectedAt: options.CollectedAt,

			Condition: options.Condition,

			Quantization: options.Quantization,

			LlamaBenchPath: options.LlamaBenchPath,

			LlamaServerPath: options.LlamaServerPath,

			ModelPath: options.ModelPath,
		},
	)
}

func executeLinuxComputeBenchmark(
	ctx context.Context,
	input linuxComputeBenchmarkExecutionInput,
) (
	benchmark protocol.ComputeBenchmark,
	retErr error,
) {
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

	// Validate the server command completely before any process starts.
	if _, _, _, err :=
		linuxBenchmarkServerCommand(
			input.LlamaServerPath,
			input.ModelPath,
		); err != nil {

		return benchmark,
			err
	}

	server, err :=
		startLinuxComputeBenchmarkServer(
			ctx,
			input.ServerStarter,
			input.HealthClient,
			input.LlamaServerPath,
			input.ModelPath,
		)

	if err != nil {
		return benchmark,
			err
	}

	stopTimeout :=
		input.StopTimeout

	if stopTimeout <= 0 {
		stopTimeout =
			m9LinuxBenchmarkServerStopTimeout
	}

	defer func() {
		stopCtx,
			cancel :=
			context.WithTimeout(
				context.Background(),
				stopTimeout,
			)

		defer cancel()

		if stopErr :=
			server.Stop(
				stopCtx,
			); stopErr != nil {

			retErr =
				errors.Join(
					retErr,
					fmt.Errorf(
						"stop M9 benchmark server: %w",
						stopErr,
					),
				)
		}
	}()

	benchmark,
		retErr =
		orchestrateLinuxComputeBenchmark(
			ctx,
			linuxComputeBenchmarkOrchestratorInput{
				CollectedAt: input.CollectedAt,

				Condition: input.Condition,

				Quantization: input.Quantization,

				LlamaBenchPath: input.LlamaBenchPath,

				ModelPath: input.ModelPath,

				TTFTBaseURL: server.baseURL,

				HTTPClient: input.TTFTClient,

				Now: input.Now,

				CollectRun: input.CollectRun,
			},
		)

	return benchmark,
		retErr
}
