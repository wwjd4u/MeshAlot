//go:build linux

package agent

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	m9StandardTTFTPrompt = "MeshAlot M9 benchmark prompt"

	m9LinuxTTFTRequestTimeout = 2 * time.Minute
)

type linuxComputeBenchmarkRunCollector func(
	context.Context,
	string,
	string,
) (linuxLlamaBenchRunResult, error)

type linuxComputeBenchmarkOrchestratorInput struct {
	CollectedAt  time.Time
	Condition    string
	Quantization string

	LlamaBenchPath string
	ModelPath      string
	TTFTBaseURL    string

	HTTPClient computeBenchmarkHTTPDoer
	Now        func() time.Time

	// CollectRun permits source-only/injected testing.
	// Real execution defaults to collectLinuxLlamaBenchRun.
	CollectRun linuxComputeBenchmarkRunCollector
}

// orchestrateLinuxComputeBenchmark connects the real Linux M9 throughput and
// resource collector with the existing five-run benchmark orchestrator and
// loopback-only TTFT collector.
//
// This function does not start llama-server. A later execution gate is
// responsible for managing a benchmark-only local server lifecycle.
func orchestrateLinuxComputeBenchmark(
	ctx context.Context,
	input linuxComputeBenchmarkOrchestratorInput,
) (protocol.ComputeBenchmark, error) {
	var benchmark protocol.ComputeBenchmark

	if ctx == nil {
		return benchmark, errors.New(
			"compute benchmark context is nil",
		)
	}

	llamaBenchPath :=
		strings.TrimSpace(
			input.LlamaBenchPath,
		)

	if llamaBenchPath == "" {
		return benchmark, errors.New(
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

		return benchmark, err
	}

	ttftBaseURL, err :=
		validateLocalBenchmarkBaseURL(
			input.TTFTBaseURL,
		)

	if err != nil {
		return benchmark, err
	}

	collectRun :=
		input.CollectRun

	if collectRun == nil {
		collectRun =
			collectLinuxLlamaBenchRun
	}

	client :=
		input.HTTPClient

	if client == nil {
		client =
			&http.Client{
				Timeout: m9LinuxTTFTRequestTimeout,
			}
	}

	now :=
		input.Now

	if now == nil {
		now = time.Now
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
				result, err :=
					collectRun(
						runCtx,
						llamaBenchPath,
						modelPath,
					)

				telemetry :=
					computeBenchmarkRunTelemetry{
						PeakSystemRAM: result.
							PeakSystemRAM,

						TemperatureCelsius: result.
							TemperatureCelsius,

						Throttled: result.
							Throttled,
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
				return measureLlamaServerTTFT(
					runCtx,
					client,
					ttftBaseURL,
					m9StandardTTFTPrompt,
					now,
				)
			},
		},
	)
}
