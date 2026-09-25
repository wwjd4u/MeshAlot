//go:build darwin

package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	m9DarwinStandardTTFTPrompt = "MeshAlot M9 benchmark prompt"

	m9DarwinTTFTRequestTimeout = 2 * time.Minute
)

type darwinComputeBenchmarkRunCollector func(
	context.Context,
	string,
	string,
) (
	darwinLlamaBenchRunResult,
	error,
)

type darwinComputeBenchmarkOrchestratorInput struct {
	CollectedAt  time.Time
	Condition    string
	Quantization string

	LlamaBenchPath  string
	LlamaServerPath string
	ModelPath       string

	ServerStarter darwinBenchmarkServerStarter

	HealthClient computeBenchmarkHTTPDoer
	TTFTClient   computeBenchmarkHTTPDoer

	Now func() time.Time

	CollectRun darwinComputeBenchmarkRunCollector

	StopTimeout time.Duration
}

// orchestrateDarwinComputeBenchmark runs the standardized M9 repetitions
// serially.
//
// Important memory-safety difference from Linux:
//
// Darwin does NOT keep the TTFT llama-server resident while llama-bench runs.
// The 30B CPU model can consume roughly 30 GiB per process on Node002.
// Therefore each repetition is:
//
//	llama-bench -> exits -> TTFT server -> measure -> stop server
//
// This preserves the standardized throughput and TTFT workloads without
// intentionally holding two complete CPU model instances in RAM at once.
func orchestrateDarwinComputeBenchmark(
	ctx context.Context,
	input darwinComputeBenchmarkOrchestratorInput,
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

	llamaBenchPath :=
		strings.TrimSpace(
			input.LlamaBenchPath,
		)

	if llamaBenchPath == "" {
		return benchmark,
			errors.New(
				"llama-bench path is required",
			)
	}

	modelPath :=
		strings.TrimSpace(
			input.ModelPath,
		)

	if _, err :=
		m9LlamaBenchArgs(
			modelPath,
		); err != nil {

		return benchmark,
			err
	}

	if _, _, _, err :=
		darwinBenchmarkServerCommand(
			input.LlamaServerPath,
			modelPath,
		); err != nil {

		return benchmark,
			err
	}

	collectRun := input.CollectRun

	if collectRun == nil {
		collectRun =
			collectDarwinLlamaBenchRun
	}

	ttftClient := input.TTFTClient

	if ttftClient == nil {
		ttftClient = &http.Client{
			Timeout: m9DarwinTTFTRequestTimeout,
		}
	}

	now := input.Now

	if now == nil {
		now = time.Now
	}

	stopTimeout :=
		input.StopTimeout

	if stopTimeout <= 0 {
		stopTimeout =
			m9DarwinBenchmarkServerStopTimeout
	}

	return orchestrateComputeBenchmarkRuns(
		ctx,
		computeBenchmarkOrchestratorInput{
			CollectedAt: input.CollectedAt,

			Condition: input.Condition,

			Quantization: input.Quantization,

			ModelPath: modelPath,

			CollectThroughput: func(
				runCtx context.Context,
				_ int,
			) (
				llamaBenchThroughput,
				computeBenchmarkRunTelemetry,
				error,
			) {
				result,
					err := collectRun(
					runCtx,
					llamaBenchPath,
					modelPath,
				)

				telemetry :=
					computeBenchmarkRunTelemetry{
						PeakSystemRAM: result.PeakSystemRAM,

						Throttled: result.Throttled,
					}

				return result.Throughput,
					telemetry,
					err
			},

			CollectTTFT: func(
				runCtx context.Context,
				_ int,
			) (
				time.Duration,
				error,
			) {
				server,
					err := startDarwinComputeBenchmarkServer(
					runCtx,
					input.ServerStarter,
					input.HealthClient,
					input.LlamaServerPath,
					modelPath,
				)

				if err != nil {
					return 0, err
				}

				duration,
					measureErr := measureLlamaServerTTFT(
					runCtx,
					ttftClient,
					server.baseURL,
					m9DarwinStandardTTFTPrompt,
					now,
				)

				stopCtx,
					cancel := context.WithTimeout(
					context.Background(),
					stopTimeout,
				)

				stopErr := server.Stop(
					stopCtx,
				)

				cancel()

				if stopErr != nil {
					measureErr =
						errors.Join(
							measureErr,
							fmt.Errorf(
								"stop Darwin M9 benchmark server: %w",
								stopErr,
							),
						)
				}

				return duration,
					measureErr
			},
		},
	)
}
