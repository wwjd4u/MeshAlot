//go:build linux

package agent

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type linuxOrchestratorTTFTDoer struct {
	calls int
}

func (d *linuxOrchestratorTTFTDoer) Do(
	_ *http.Request,
) (*http.Response, error) {
	d.calls++

	stream :=
		"data: {\"prompt_progress\":{\"total\":512,\"processed\":512}}\n\n" +
			"data: {\"content\":\"A\",\"tokens\":[123],\"stop\":false}\n\n"

	return &http.Response{
		StatusCode: http.StatusOK,

		Body: io.NopCloser(
			strings.NewReader(
				stream,
			),
		),

		Header: make(
			http.Header,
		),
	}, nil
}

func TestOrchestrateLinuxComputeBenchmark(
	t *testing.T,
) {
	collectedAt :=
		time.Date(
			2026,
			9,
			17,
			8,
			0,
			0,
			0,
			time.UTC,
		)

	client :=
		&linuxOrchestratorTTFTDoer{}

	currentTime := collectedAt

	clock :=
		func() time.Time {
			value := currentTime

			currentTime =
				currentTime.Add(
					125 *
						time.Millisecond,
				)

			return value
		}

	runCalls := 0
	temperature := 63.5
	throttled := false

	benchmark, err :=
		orchestrateLinuxComputeBenchmark(
			context.Background(),
			linuxComputeBenchmarkOrchestratorInput{
				CollectedAt: collectedAt,

				Condition: "baseline",

				Quantization: "Q4_K_M",

				LlamaBenchPath: "/opt/llama-bench",

				ModelPath: "/models/m9-standard.gguf",

				TTFTBaseURL: "http://127.0.0.1:8080",

				HTTPClient: client,

				Now: clock,

				CollectRun: func(
					_ context.Context,
					_ string,
					modelPath string,
				) (
					linuxLlamaBenchRunResult,
					error,
				) {
					runCalls++

					return linuxLlamaBenchRunResult{
						Throughput: llamaBenchThroughput{
							RuntimeVersion: "8cf427ff-5163",

							ModelFilename: modelPath,

							ModelType: "test 7B Q4_K - Medium",

							Backend: "Vulkan",

							PromptTokens: m9StandardPromptTokens,

							GeneratedTokens: m9StandardGeneratedTokens,

							PromptTokensPerSecond: []float64{
								700 +
									float64(
										runCalls,
									),
							},

							GenerationTokensPerSecond: []float64{
								60 +
									float64(
										runCalls,
									),
							},
						},

						PeakSystemRAM: uint64(
							8+runCalls,
						) *
							1024 *
							1024 *
							1024,

						TemperatureCelsius: &temperature,

						Throttled: &throttled,
					}, nil
				},
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if runCalls !=
		m9StandardRepetitions {

		t.Fatalf(
			"Linux run calls = %d",
			runCalls,
		)
	}

	if client.calls !=
		m9StandardRepetitions {

		t.Fatalf(
			"TTFT calls = %d",
			client.calls,
		)
	}

	if len(benchmark.Samples) !=
		m9StandardRepetitions {

		t.Fatalf(
			"samples = %d",
			len(
				benchmark.Samples,
			),
		)
	}

	if benchmark.Runtime !=
		"llama.cpp" {

		t.Fatalf(
			"runtime = %q",
			benchmark.Runtime,
		)
	}

	if benchmark.RuntimeVersion !=
		"8cf427ff-5163" {

		t.Fatalf(
			"runtime version = %q",
			benchmark.RuntimeVersion,
		)
	}

	if benchmark.Model !=
		"m9-standard.gguf" {

		t.Fatalf(
			"model = %q",
			benchmark.Model,
		)
	}

	if benchmark.Quantization !=
		"Q4_K_M" {

		t.Fatalf(
			"quantization = %q",
			benchmark.Quantization,
		)
	}

	for i, sample := range benchmark.Samples {

		if !sample.Success {
			t.Fatalf(
				"run %d failed: %s",
				i+1,
				sample.Error,
			)
		}

		if sample.TimeToFirstTokenMs !=
			125 {

			t.Fatalf(
				"run %d TTFT = %v",
				i+1,
				sample.
					TimeToFirstTokenMs,
			)
		}

		if sample.PeakSystemRAMBytes ==
			0 {

			t.Fatalf(
				"run %d RAM missing",
				i+1,
			)
		}

		if sample.PeakVRAMBytes != nil {
			t.Fatalf(
				"run %d invented VRAM",
				i+1,
			)
		}

		if sample.TemperatureCelsius ==
			nil ||
			*sample.
				TemperatureCelsius !=
				temperature {

			t.Fatalf(
				"run %d temperature changed",
				i+1,
			)
		}

		if sample.Throttled == nil ||
			*sample.Throttled {

			t.Fatalf(
				"run %d throttling changed",
				i+1,
			)
		}
	}
}

func TestOrchestrateLinuxComputeBenchmarkRejectsRemoteTTFT(
	t *testing.T,
) {
	runCalls := 0

	_, err :=
		orchestrateLinuxComputeBenchmark(
			context.Background(),
			linuxComputeBenchmarkOrchestratorInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				Quantization: "Q4_K_M",

				LlamaBenchPath: "/opt/llama-bench",

				ModelPath: "/models/m9-standard.gguf",

				TTFTBaseURL: "http://10.0.0.10:8080",

				CollectRun: func(
					context.Context,
					string,
					string,
				) (
					linuxLlamaBenchRunResult,
					error,
				) {
					runCalls++

					return linuxLlamaBenchRunResult{},
						nil
				},
			},
		)

	if err == nil {
		t.Fatal(
			"remote TTFT URL was accepted",
		)
	}

	if runCalls != 0 {
		t.Fatal(
			"benchmark ran before TTFT URL validation",
		)
	}
}
