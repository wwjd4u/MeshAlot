package agent

import (
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestAssembleComputeBenchmark(
	t *testing.T,
) {
	vram := uint64(8 * 1024 * 1024 * 1024)
	temp := 72.5
	throttled := false

	throughput := llamaBenchThroughput{
		RuntimeVersion: "8cf427ff-5163",

		ModelFilename: "/models/m9-standard-Q4_K_M.gguf",

		ModelType: "test 7B Q4_K - Medium",

		Backend: "Metal",

		PromptTokens: 512,

		GeneratedTokens: 128,

		PromptTokensPerSecond: []float64{
			700,
			710,
			705,
			695,
			708,
		},

		GenerationTokensPerSecond: []float64{
			60,
			61,
			60.5,
			59.5,
			60.8,
		},
	}

	telemetry := make(
		[]computeBenchmarkRunTelemetry,
		m9StandardRepetitions,
	)

	for i := range telemetry {
		telemetry[i] =
			computeBenchmarkRunTelemetry{
				TimeToFirstToken: time.Duration(200+i*5) *
					time.Millisecond,

				PeakSystemRAM: uint64(8+i) *
					1024 *
					1024 *
					1024,

				PeakVRAMBytes: &vram,

				TemperatureCelsius: &temp,

				Throttled: &throttled,
			}
	}

	collectedAt := time.Date(
		2026,
		9,
		14,
		23,
		0,
		0,
		0,
		time.UTC,
	)

	benchmark, err :=
		assembleComputeBenchmark(
			computeBenchmarkAssemblyInput{
				CollectedAt: collectedAt,

				Condition: "baseline",

				Quantization: "Q4_K_M",

				Throughput: throughput,

				Telemetry: telemetry,
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if benchmark.SchemaVersion !=
		protocol.ComputeBenchmarkSchemaVersion {
		t.Fatalf(
			"schema = %q",
			benchmark.SchemaVersion,
		)
	}

	if benchmark.WorkloadID !=
		m9StandardWorkloadID {
		t.Fatalf(
			"workload = %q",
			benchmark.WorkloadID,
		)
	}

	if benchmark.Runtime != "llama.cpp" {
		t.Fatalf(
			"runtime = %q",
			benchmark.Runtime,
		)
	}

	if benchmark.Model !=
		"m9-standard-Q4_K_M.gguf" {
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

	if benchmark.ContextSize != 4096 {
		t.Fatalf(
			"context size = %d",
			benchmark.ContextSize,
		)
	}

	if len(benchmark.Samples) !=
		m9StandardRepetitions {
		t.Fatalf(
			"samples = %d",
			len(benchmark.Samples),
		)
	}

	first := benchmark.Samples[0]

	if first.Run != 1 ||
		first.PromptTokens != 512 ||
		first.GeneratedTokens != 128 ||
		first.PromptTokensPerSecond != 700 ||
		first.GenerationTokensPerSecond != 60 ||
		first.TimeToFirstTokenMs != 200 ||
		!first.Success {
		t.Fatalf(
			"first sample changed: %#v",
			first,
		)
	}

	if first.PeakVRAMBytes == nil ||
		*first.PeakVRAMBytes != vram {
		t.Fatal("VRAM telemetry changed")
	}

	if first.TemperatureCelsius == nil ||
		*first.TemperatureCelsius != temp {
		t.Fatal(
			"temperature telemetry changed",
		)
	}

	if first.Throttled == nil ||
		*first.Throttled {
		t.Fatal(
			"throttling telemetry changed",
		)
	}
}

func TestAssembleComputeBenchmarkAllowsOptionalTelemetry(
	t *testing.T,
) {
	throughput :=
		validAssemblyThroughputFixture()

	telemetry :=
		validAssemblyTelemetryFixture()

	for i := range telemetry {
		telemetry[i].PeakVRAMBytes = nil
		telemetry[i].TemperatureCelsius = nil
	}

	benchmark, err :=
		assembleComputeBenchmark(
			computeBenchmarkAssemblyInput{
				CollectedAt: time.Now().UTC(),

				Condition: "baseline",

				Quantization: "Q4_K_M",

				Throughput: throughput,

				Telemetry: telemetry,
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	for _, sample := range benchmark.Samples {
		if sample.PeakVRAMBytes != nil {
			t.Fatal(
				"unavailable VRAM was invented",
			)
		}

		if sample.TemperatureCelsius != nil {
			t.Fatal(
				"unavailable temperature was invented",
			)
		}
	}
}

func TestAssembleComputeBenchmarkRejectsMismatchedRuns(
	t *testing.T,
) {
	throughput :=
		validAssemblyThroughputFixture()

	telemetry :=
		validAssemblyTelemetryFixture()

	telemetry = telemetry[:4]

	_, err := assembleComputeBenchmark(
		computeBenchmarkAssemblyInput{
			CollectedAt: time.Now().UTC(),

			Condition: "baseline",

			Throughput: throughput,

			Telemetry: telemetry,
		},
	)

	if err == nil {
		t.Fatal(
			"mismatched repetitions were accepted",
		)
	}
}

func TestAssembleComputeBenchmarkRejectsMissingRAM(
	t *testing.T,
) {
	throughput :=
		validAssemblyThroughputFixture()

	telemetry :=
		validAssemblyTelemetryFixture()

	telemetry[2].PeakSystemRAM = 0

	_, err := assembleComputeBenchmark(
		computeBenchmarkAssemblyInput{
			CollectedAt: time.Now().UTC(),

			Condition: "baseline",

			Throughput: throughput,

			Telemetry: telemetry,
		},
	)

	if err == nil {
		t.Fatal(
			"missing RAM measurement was accepted",
		)
	}
}

func validAssemblyThroughputFixture() llamaBenchThroughput {
	return llamaBenchThroughput{
		RuntimeVersion: "test-1",

		ModelFilename: "/models/model.gguf",

		ModelType: "test",

		Backend: "CPU",

		PromptTokens: 512,

		GeneratedTokens: 128,

		PromptTokensPerSecond: []float64{
			100,
			101,
			102,
			103,
			104,
		},

		GenerationTokensPerSecond: []float64{
			20,
			21,
			22,
			23,
			24,
		},
	}
}

func validAssemblyTelemetryFixture() []computeBenchmarkRunTelemetry {
	result := make(
		[]computeBenchmarkRunTelemetry,
		m9StandardRepetitions,
	)

	throttled := false

	for i := range result {
		result[i] =
			computeBenchmarkRunTelemetry{
				TimeToFirstToken: 250 * time.Millisecond,

				PeakSystemRAM: 4 * 1024 * 1024 * 1024,

				Throttled: &throttled,
			}
	}

	return result
}
