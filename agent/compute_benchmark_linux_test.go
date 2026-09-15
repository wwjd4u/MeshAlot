package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const m9LinuxSingleRunJSON = `[
  {
    "build_commit": "8cf427ff",
    "build_number": 5163,
    "backends": "Vulkan",
    "model_filename": "/models/m9-standard.gguf",
    "model_type": "test 7B Q4_K - Medium",
    "n_prompt": 512,
    "n_gen": 0,
    "avg_ts": 7155.74,
    "samples_ts": [7155.74]
  },
  {
    "build_commit": "8cf427ff",
    "build_number": 5163,
    "backends": "Vulkan",
    "model_filename": "/models/m9-standard.gguf",
    "model_type": "test 7B Q4_K - Medium",
    "n_prompt": 0,
    "n_gen": 128,
    "avg_ts": 119.03,
    "samples_ts": [119.03]
  }
]`

type fakeLinuxComputeCommand struct {
	name   string
	args   []string
	stdout string
	stderr string
	err    error
}

type fakeLinuxComputeRunner struct {
	commands []fakeLinuxComputeCommand
	index    int
}

func (f *fakeLinuxComputeRunner) Run(
	_ context.Context,
	name string,
	args ...string,
) (linuxComputeBenchmarkCommandOutput, error) {
	if f.index >= len(f.commands) {
		return linuxComputeBenchmarkCommandOutput{},
			errors.New(
				"unexpected command",
			)
	}

	expected :=
		f.commands[f.index]

	f.index++

	if name != expected.name {
		return linuxComputeBenchmarkCommandOutput{},
			errors.New(
				"unexpected command name",
			)
	}

	gotArgs :=
		strings.Join(
			args,
			" ",
		)

	wantArgs :=
		strings.Join(
			expected.args,
			" ",
		)

	if gotArgs != wantArgs {
		return linuxComputeBenchmarkCommandOutput{},
			errors.New(
				"unexpected command arguments",
			)
	}

	return linuxComputeBenchmarkCommandOutput{
		Stdout: []byte(
			expected.stdout,
		),

		Stderr: []byte(
			expected.stderr,
		),
	}, expected.err
}

type fakeLinuxTelemetryReader struct {
	globs map[string][]string
	files map[string][][]byte
	reads map[string]int
}

func (f *fakeLinuxTelemetryReader) Glob(
	pattern string,
) ([]string, error) {
	return append(
		[]string(nil),
		f.globs[pattern]...,
	), nil
}

func (f *fakeLinuxTelemetryReader) ReadFile(
	path string,
) ([]byte, error) {
	values := f.files[path]

	if len(values) == 0 {
		return nil, errors.New(
			"missing fake file",
		)
	}

	if f.reads == nil {
		f.reads =
			make(map[string]int)
	}

	index := f.reads[path]

	if index >= len(values) {
		index = len(values) - 1
	}

	f.reads[path]++

	return append(
		[]byte(nil),
		values[index]...,
	), nil
}

func TestParseLinuxTimePeakRSS(
	t *testing.T,
) {
	raw := []byte(
		"\tMaximum resident set size (kbytes): 72128\n",
	)

	got, err :=
		parseLinuxTimePeakRSS(raw)

	if err != nil {
		t.Fatal(err)
	}

	want :=
		uint64(72128 * 1024)

	if got != want {
		t.Fatalf(
			"peak RSS = %d, want %d",
			got,
			want,
		)
	}
}

func TestParseLinuxTimePeakRSSRejectsBadInput(
	t *testing.T,
) {
	for _, raw := range []string{
		"",
		"Maximum resident set size (kbytes): 0",
		"Maximum resident set size (kbytes): nope",
	} {
		if _, err :=
			parseLinuxTimePeakRSS(
				[]byte(raw),
			); err == nil {

			t.Fatalf(
				"invalid peak RSS %q accepted",
				raw,
			)
		}
	}
}

func TestParseLinuxMilliCelsius(
	t *testing.T,
) {
	got, err :=
		parseLinuxMilliCelsius(
			[]byte("63500\n"),
		)

	if err != nil {
		t.Fatal(err)
	}

	if got != 63.5 {
		t.Fatalf(
			"temperature = %v",
			got,
		)
	}
}

func TestLinuxBenchmarkThrottleDelta(
	t *testing.T,
) {
	before :=
		map[string]uint64{
			"core":    10,
			"package": 4,
		}

	after :=
		map[string]uint64{
			"core":    11,
			"package": 4,
		}

	got :=
		linuxBenchmarkThrottleDelta(
			before,
			after,
		)

	if got == nil || !*got {
		t.Fatal(
			"throttle increase was not detected",
		)
	}

	got =
		linuxBenchmarkThrottleDelta(
			before,
			before,
		)

	if got == nil || *got {
		t.Fatal(
			"unchanged throttle counters were not preserved as false",
		)
	}
}

func TestMeasureLinuxLlamaBenchRun(
	t *testing.T,
) {
	runner :=
		&fakeLinuxComputeRunner{
			commands: []fakeLinuxComputeCommand{
				{
					name: m9LinuxTimePath,

					args: []string{
						"-v",
						"/opt/llama-bench",
						"-m",
						"/models/m9-standard.gguf",
						"-p",
						"512",
						"-n",
						"128",
						"-r",
						"1",
						"-o",
						"json",
					},

					stdout: m9LinuxSingleRunJSON,

					stderr: "Maximum resident set size (kbytes): 72128\n",
				},
			},
		}

	temperaturePath :=
		"/sys/class/thermal/thermal_zone0/temp"

	throttlePath :=
		"/sys/devices/system/cpu/cpu0/thermal_throttle/core_throttle_count"

	reader :=
		&fakeLinuxTelemetryReader{
			globs: map[string][]string{
				m9LinuxTemperatureGlobs[0]: {
					temperaturePath,
				},

				m9LinuxThrottleGlob: {
					throttlePath,
				},
			},

			files: map[string][][]byte{
				temperaturePath: {
					[]byte("55000\n"),
					[]byte("63000\n"),
				},

				throttlePath: {
					[]byte("10\n"),
					[]byte("11\n"),
				},
			},
		}

	result, err :=
		measureLinuxLlamaBenchRun(
			context.Background(),
			runner,
			reader,
			"/opt/llama-bench",
			"/models/m9-standard.gguf",
		)

	if err != nil {
		t.Fatal(err)
	}

	if runner.index != 1 {
		t.Fatalf(
			"commands executed = %d, want 1",
			runner.index,
		)
	}

	if result.PeakSystemRAM !=
		72128*1024 {

		t.Fatalf(
			"peak RAM = %d",
			result.PeakSystemRAM,
		)
	}

	if result.TemperatureCelsius == nil ||
		*result.TemperatureCelsius != 63 {

		t.Fatalf(
			"temperature = %v",
			result.TemperatureCelsius,
		)
	}

	if result.Throttled == nil ||
		!*result.Throttled {

		t.Fatal(
			"throttling was not preserved",
		)
	}

	if len(
		result.
			Throughput.
			PromptTokensPerSecond,
	) != 1 ||
		len(
			result.
				Throughput.
				GenerationTokensPerSecond,
		) != 1 {

		t.Fatal(
			"single-run throughput changed",
		)
	}
}

func TestMeasureLinuxLlamaBenchRunAllowsUnavailableOptionalTelemetry(
	t *testing.T,
) {
	runner :=
		&fakeLinuxComputeRunner{
			commands: []fakeLinuxComputeCommand{
				{
					name: m9LinuxTimePath,

					args: []string{
						"-v",
						"/opt/llama-bench",
						"-m",
						"/models/m9-standard.gguf",
						"-p",
						"512",
						"-n",
						"128",
						"-r",
						"1",
						"-o",
						"json",
					},

					stdout: m9LinuxSingleRunJSON,

					stderr: "Maximum resident set size (kbytes): 72128\n",
				},
			},
		}

	result, err :=
		measureLinuxLlamaBenchRun(
			context.Background(),
			runner,
			nil,
			"/opt/llama-bench",
			"/models/m9-standard.gguf",
		)

	if err != nil {
		t.Fatal(err)
	}

	if result.TemperatureCelsius != nil {
		t.Fatal(
			"unavailable temperature was invented",
		)
	}

	if result.Throttled != nil {
		t.Fatal(
			"unavailable throttling was invented",
		)
	}
}
