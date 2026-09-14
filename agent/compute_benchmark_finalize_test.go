package agent

import (
	"strings"
	"testing"
	"time"
)

func TestFinalizeComputeBenchmarkRunsAllSuccessful(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	benchmark, err :=
		finalizeComputeBenchmarkRuns(
			computeBenchmarkFinalizeInput{
				CollectedAt: time.Date(
					2026,
					9,
					14,
					23,
					30,
					0,
					0,
					time.UTC,
				),

				Condition: "baseline",

				Quantization: "Q4_K_M",

				ModelPath: "/models/m9-standard.gguf",

				Runs: runs,
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if len(benchmark.Samples) !=
		m9StandardRepetitions {
		t.Fatalf(
			"samples = %d",
			len(benchmark.Samples),
		)
	}

	if benchmark.Model !=
		"m9-standard.gguf" {
		t.Fatalf(
			"model = %q",
			benchmark.Model,
		)
	}

	if benchmark.RuntimeVersion !=
		"8cf427ff-5163" {
		t.Fatalf(
			"runtime version = %q",
			benchmark.RuntimeVersion,
		)
	}

	for i, sample := range benchmark.Samples {
		if !sample.Success {
			t.Fatalf(
				"run %d unexpectedly failed",
				i+1,
			)
		}

		if sample.Error != "" {
			t.Fatalf(
				"run %d contains error %q",
				i+1,
				sample.Error,
			)
		}
	}
}

func TestFinalizeComputeBenchmarkRunsPreservesFailure(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	throttled := true

	runs[2] =
		computeBenchmarkRunMeasurement{
			Telemetry: computeBenchmarkRunTelemetry{
				TimeToFirstToken: 325 *
					time.Millisecond,

				PeakSystemRAM: 3 *
					1024 *
					1024 *
					1024,

				Throttled: &throttled,
			},

			Error: "llama-bench execution failed: signal killed",
		}

	benchmark, err :=
		finalizeComputeBenchmarkRuns(
			computeBenchmarkFinalizeInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				Quantization: "Q4_K_M",

				ModelPath: "/models/m9-standard.gguf",

				Runs: runs,
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	failed := benchmark.Samples[2]

	if failed.Success {
		t.Fatal(
			"failed run was reported successful",
		)
	}

	if failed.Error !=
		"llama-bench execution failed: signal killed" {
		t.Fatalf(
			"failure text = %q",
			failed.Error,
		)
	}

	if failed.TimeToFirstTokenMs !=
		325 {
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

	if failed.Throttled == nil ||
		!*failed.Throttled {
		t.Fatal(
			"known throttling on failed run was lost",
		)
	}
}

func TestFinalizeComputeBenchmarkRunsAllowsAllFailed(
	t *testing.T,
) {
	runs := make(
		[]computeBenchmarkRunMeasurement,
		m9StandardRepetitions,
	)

	for i := range runs {
		runs[i].Error =
			"benchmark repetition failed"
	}

	benchmark, err :=
		finalizeComputeBenchmarkRuns(
			computeBenchmarkFinalizeInput{
				CollectedAt: time.Now().UTC(),

				Condition: "loaded",

				Quantization: "Q4_K_M",

				ModelPath: "/models/m9-standard.gguf",

				Runs: runs,
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if benchmark.RuntimeVersion != "" {
		t.Fatalf(
			"runtime version = %q, want empty when no run succeeded",
			benchmark.RuntimeVersion,
		)
	}

	if benchmark.Model !=
		"m9-standard.gguf" {
		t.Fatal(
			"configured model was not preserved",
		)
	}

	for i, sample := range benchmark.Samples {
		if sample.Success {
			t.Fatalf(
				"run %d unexpectedly succeeded",
				i+1,
			)
		}

		if sample.Error == "" {
			t.Fatalf(
				"run %d lost failure text",
				i+1,
			)
		}
	}
}

func TestFinalizeComputeBenchmarkRunsRejectsMissingFailureText(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	runs[1] =
		computeBenchmarkRunMeasurement{}

	_, err :=
		finalizeComputeBenchmarkRuns(
			computeBenchmarkFinalizeInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				ModelPath: "/models/m9-standard.gguf",

				Runs: runs,
			},
		)

	if err == nil {
		t.Fatal(
			"empty failed run without error text was accepted",
		)
	}
}

func TestFinalizeComputeBenchmarkRunsRejectsLongFailure(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	runs[0] =
		computeBenchmarkRunMeasurement{
			Error: strings.Repeat(
				"x",
				m9ComputeBenchmarkRunErrorMax+1,
			),
		}

	_, err :=
		finalizeComputeBenchmarkRuns(
			computeBenchmarkFinalizeInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				ModelPath: "/models/m9-standard.gguf",

				Runs: runs,
			},
		)

	if err == nil {
		t.Fatal(
			"oversized failure text was accepted",
		)
	}
}

func TestFinalizeComputeBenchmarkRunsRejectsMetadataChange(
	t *testing.T,
) {
	runs := validIndependentRunFixture()

	runs[3].Throughput.Backend =
		"CPU"

	_, err :=
		finalizeComputeBenchmarkRuns(
			computeBenchmarkFinalizeInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				ModelPath: "/models/m9-standard.gguf",

				Runs: runs,
			},
		)

	if err == nil {
		t.Fatal(
			"successful run metadata change was accepted",
		)
	}
}
