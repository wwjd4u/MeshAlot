package agent

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOrchestrateComputeBenchmarkRuns(
	t *testing.T,
) {
	throughputCalls := 0
	ttftCalls := 0
	throttled := false

	benchmark, err :=
		orchestrateComputeBenchmarkRuns(
			context.Background(),
			computeBenchmarkOrchestratorInput{
				CollectedAt: time.Date(
					2026,
					9,
					14,
					23,
					45,
					0,
					0,
					time.UTC,
				),

				Condition: "baseline",

				Quantization: "Q4_K_M",

				ModelPath: "/models/m9-standard.gguf",

				CollectThroughput: func(
					_ context.Context,
					run int,
				) (
					llamaBenchThroughput,
					computeBenchmarkRunTelemetry,
					error,
				) {
					throughputCalls++

					return llamaBenchThroughput{
							RuntimeVersion: "8cf427ff-5163",

							ModelFilename: "/models/m9-standard.gguf",

							ModelType: "test 7B Q4_K - Medium",

							Backend: "Metal",

							PromptTokens: m9StandardPromptTokens,

							GeneratedTokens: m9StandardGeneratedTokens,

							PromptTokensPerSecond: []float64{
								float64(
									100 + run,
								),
							},

							GenerationTokensPerSecond: []float64{
								float64(
									20 + run,
								),
							},
						},
						computeBenchmarkRunTelemetry{
							PeakSystemRAM: uint64(
								4+run,
							) *
								1024 *
								1024 *
								1024,

							Throttled: &throttled,
						},
						nil
				},

				CollectTTFT: func(
					_ context.Context,
					run int,
				) (
					time.Duration,
					error,
				) {
					ttftCalls++

					return time.Duration(
							200+run*10,
						) *
							time.Millisecond,
						nil
				},
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if throughputCalls !=
		m9StandardRepetitions {
		t.Fatalf(
			"throughput calls = %d",
			throughputCalls,
		)
	}

	if ttftCalls !=
		m9StandardRepetitions {
		t.Fatalf(
			"TTFT calls = %d",
			ttftCalls,
		)
	}

	if len(benchmark.Samples) !=
		m9StandardRepetitions {
		t.Fatalf(
			"samples = %d",
			len(benchmark.Samples),
		)
	}

	for i, sample := range benchmark.Samples {
		if !sample.Success {
			t.Fatalf(
				"run %d unexpectedly failed",
				i+1,
			)
		}

		if sample.Run != i+1 {
			t.Fatalf(
				"run number = %d, want %d",
				sample.Run,
				i+1,
			)
		}
	}
}

func TestOrchestrateComputeBenchmarkRunsPreservesThroughputFailure(
	t *testing.T,
) {
	throughputCalls := 0
	ttftCalls := 0
	throttled := false

	benchmark, err :=
		orchestrateComputeBenchmarkRuns(
			context.Background(),
			computeBenchmarkOrchestratorInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				Quantization: "Q4_K_M",

				ModelPath: "/models/m9-standard.gguf",

				CollectThroughput: func(
					_ context.Context,
					run int,
				) (
					llamaBenchThroughput,
					computeBenchmarkRunTelemetry,
					error,
				) {
					throughputCalls++

					if run == 3 {
						return llamaBenchThroughput{},
							computeBenchmarkRunTelemetry{
								PeakSystemRAM: 3 *
									1024 *
									1024 *
									1024,

								Throttled: &throttled,
							},
							errors.New(
								"process exited 1",
							)
					}

					return validOrchestratorThroughput(
							run,
						),
						computeBenchmarkRunTelemetry{
							PeakSystemRAM: 4 *
								1024 *
								1024 *
								1024,

							Throttled: &throttled,
						},
						nil
				},

				CollectTTFT: func(
					_ context.Context,
					run int,
				) (
					time.Duration,
					error,
				) {
					ttftCalls++

					return time.Duration(
							300+run,
						) *
							time.Millisecond,
						nil
				},
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if throughputCalls != 5 ||
		ttftCalls != 5 {
		t.Fatalf(
			"collectors stopped early: throughput=%d ttft=%d",
			throughputCalls,
			ttftCalls,
		)
	}

	failed := benchmark.Samples[2]

	if failed.Success {
		t.Fatal(
			"failed throughput run became successful",
		)
	}

	if failed.Error !=
		"throughput: process exited 1" {
		t.Fatalf(
			"failure text = %q",
			failed.Error,
		)
	}

	if failed.TimeToFirstTokenMs !=
		303 {
		t.Fatalf(
			"partial TTFT = %v",
			failed.TimeToFirstTokenMs,
		)
	}

	if failed.PeakSystemRAMBytes !=
		3*1024*1024*1024 {
		t.Fatalf(
			"partial RAM = %d",
			failed.PeakSystemRAMBytes,
		)
	}
}

func TestOrchestrateComputeBenchmarkRunsPreservesTTFTFailure(
	t *testing.T,
) {
	throttled := false

	benchmark, err :=
		orchestrateComputeBenchmarkRuns(
			context.Background(),
			computeBenchmarkOrchestratorInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				ModelPath: "/models/m9-standard.gguf",

				CollectThroughput: func(
					_ context.Context,
					run int,
				) (
					llamaBenchThroughput,
					computeBenchmarkRunTelemetry,
					error,
				) {
					return validOrchestratorThroughput(
							run,
						),
						computeBenchmarkRunTelemetry{
							PeakSystemRAM: 4 *
								1024 *
								1024 *
								1024,

							Throttled: &throttled,
						},
						nil
				},

				CollectTTFT: func(
					_ context.Context,
					run int,
				) (
					time.Duration,
					error,
				) {
					if run == 4 {
						return 0,
							errors.New(
								"stream ended before first token",
							)
					}

					return 250 *
							time.Millisecond,
						nil
				},
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	failed := benchmark.Samples[3]

	if failed.Success {
		t.Fatal(
			"failed TTFT run became successful",
		)
	}

	if failed.Error !=
		"ttft: stream ended before first token" {
		t.Fatalf(
			"failure text = %q",
			failed.Error,
		)
	}

	if failed.PeakSystemRAMBytes !=
		4*1024*1024*1024 {
		t.Fatal(
			"successful throughput RAM measurement was lost",
		)
	}

	if failed.TimeToFirstTokenMs != 0 {
		t.Fatal(
			"failed TTFT was invented",
		)
	}
}

func TestOrchestrateComputeBenchmarkRunsRejectsNilCollectors(
	t *testing.T,
) {
	input :=
		computeBenchmarkOrchestratorInput{
			CollectedAt: time.Now().UTC(),

			Condition: "baseline",

			ModelPath: "/models/m9-standard.gguf",
		}

	if _, err :=
		orchestrateComputeBenchmarkRuns(
			context.Background(),
			input,
		); err == nil {
		t.Fatal(
			"nil collectors were accepted",
		)
	}
}

func TestComputeBenchmarkRunFailure(
	t *testing.T,
) {
	got := computeBenchmarkRunFailure(
		errors.New(
			"bench failed",
		),
		errors.New(
			"first token failed",
		),
	)

	want :=
		"throughput: bench failed; ttft: first token failed"

	if got != want {
		t.Fatalf(
			"failure = %q, want %q",
			got,
			want,
		)
	}
}

func validOrchestratorThroughput(
	run int,
) llamaBenchThroughput {
	return llamaBenchThroughput{
		RuntimeVersion: "8cf427ff-5163",

		ModelFilename: "/models/m9-standard.gguf",

		ModelType: "test 7B Q4_K - Medium",

		Backend: "Metal",

		PromptTokens: m9StandardPromptTokens,

		GeneratedTokens: m9StandardGeneratedTokens,

		PromptTokensPerSecond: []float64{
			float64(
				100 + run,
			),
		},

		GenerationTokensPerSecond: []float64{
			float64(
				20 + run,
			),
		},
	}
}
