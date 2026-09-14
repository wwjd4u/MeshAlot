package agent

import (
	"testing"
	"time"
)

func TestCombineComputeBenchmarkRuns(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	throughput, telemetry, err :=
		combineComputeBenchmarkRuns(runs)

	if err != nil {
		t.Fatal(err)
	}

	if len(
		throughput.PromptTokensPerSecond,
	) != m9StandardRepetitions {
		t.Fatalf(
			"prompt samples = %d",
			len(
				throughput.
					PromptTokensPerSecond,
			),
		)
	}

	if len(
		throughput.GenerationTokensPerSecond,
	) != m9StandardRepetitions {
		t.Fatalf(
			"generation samples = %d",
			len(
				throughput.
					GenerationTokensPerSecond,
			),
		)
	}

	if len(telemetry) !=
		m9StandardRepetitions {
		t.Fatalf(
			"telemetry samples = %d",
			len(telemetry),
		)
	}

	if throughput.PromptTokensPerSecond[0] !=
		100 ||
		throughput.PromptTokensPerSecond[4] !=
			104 {
		t.Fatal(
			"prompt repetitions were not preserved",
		)
	}

	if throughput.
		GenerationTokensPerSecond[0] !=
		20 ||
		throughput.
			GenerationTokensPerSecond[4] !=
			24 {
		t.Fatal(
			"generation repetitions were not preserved",
		)
	}

	if telemetry[0].TimeToFirstToken !=
		200*time.Millisecond ||
		telemetry[4].TimeToFirstToken !=
			240*time.Millisecond {
		t.Fatal(
			"TTFT repetitions were not preserved",
		)
	}

	if telemetry[0].PeakSystemRAM !=
		4*1024*1024*1024 ||
		telemetry[4].PeakSystemRAM !=
			8*1024*1024*1024 {
		t.Fatal(
			"RAM repetitions were not preserved",
		)
	}
}

func TestCombineComputeBenchmarkRunsRejectsWrongCount(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	_, _, err :=
		combineComputeBenchmarkRuns(
			runs[:4],
		)

	if err == nil {
		t.Fatal(
			"four-run benchmark was accepted",
		)
	}
}

func TestCombineComputeBenchmarkRunsRejectsMultiSampleInvocation(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	runs[2].
		Throughput.
		PromptTokensPerSecond =
		[]float64{
			102,
			103,
		}

	runs[2].
		Throughput.
		GenerationTokensPerSecond =
		[]float64{
			22,
			23,
		}

	_, _, err :=
		combineComputeBenchmarkRuns(runs)

	if err == nil {
		t.Fatal(
			"multi-repetition llama-bench invocation was accepted",
		)
	}
}

func TestCombineComputeBenchmarkRunsRejectsMetadataChange(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	runs[3].Throughput.Backend =
		"CPU"

	_, _, err :=
		combineComputeBenchmarkRuns(runs)

	if err == nil {
		t.Fatal(
			"backend change between runs was accepted",
		)
	}
}

func TestCombineComputeBenchmarkRunsRejectsMissingTTFT(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	runs[1].
		Telemetry.
		TimeToFirstToken = 0

	_, _, err :=
		combineComputeBenchmarkRuns(runs)

	if err == nil {
		t.Fatal(
			"run without TTFT was accepted",
		)
	}
}

func TestCombineComputeBenchmarkRunsRejectsMissingRAM(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	runs[1].
		Telemetry.
		PeakSystemRAM = 0

	_, _, err :=
		combineComputeBenchmarkRuns(runs)

	if err == nil {
		t.Fatal(
			"run without peak RAM was accepted",
		)
	}
}

func validIndependentRunFixture() []computeBenchmarkRunMeasurement {
	runs := make(
		[]computeBenchmarkRunMeasurement,
		m9StandardRepetitions,
	)

	throttled := false

	for i := range runs {
		runs[i] =
			computeBenchmarkRunMeasurement{
				Throughput: llamaBenchThroughput{
					RuntimeVersion: "8cf427ff-5163",

					ModelFilename: "/models/m9-standard.gguf",

					ModelType: "test 7B Q4_K - Medium",

					Backend: "Metal",

					PromptTokens: m9StandardPromptTokens,

					GeneratedTokens: m9StandardGeneratedTokens,

					PromptTokensPerSecond: []float64{
						float64(
							100 + i,
						),
					},

					GenerationTokensPerSecond: []float64{
						float64(
							20 + i,
						),
					},
				},

				Telemetry: computeBenchmarkRunTelemetry{
					TimeToFirstToken: time.Duration(
						200+i*10,
					) *
						time.Millisecond,

					PeakSystemRAM: uint64(
						4+i,
					) *
						1024 *
						1024 *
						1024,

					Throttled: &throttled,
				},
			}
	}

	return runs
}
