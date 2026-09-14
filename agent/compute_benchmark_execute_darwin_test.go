//go:build darwin

package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeDarwinComputeCommand struct {
	name   string
	args   []string
	stdout string
	stderr string
	err    error
}

type fakeDarwinComputeRunner struct {
	commands []fakeDarwinComputeCommand
	index    int
}

func (f *fakeDarwinComputeRunner) Run(
	_ context.Context,
	name string,
	args ...string,
) (darwinComputeBenchmarkCommandOutput, error) {
	if f.index >= len(f.commands) {
		return darwinComputeBenchmarkCommandOutput{},
			errors.New(
				"unexpected command",
			)
	}

	expected := f.commands[f.index]
	f.index++

	if name != expected.name {
		return darwinComputeBenchmarkCommandOutput{},
			errors.New(
				"unexpected command name",
			)
	}

	gotArgs := strings.Join(
		args,
		" ",
	)

	wantArgs := strings.Join(
		expected.args,
		" ",
	)

	if gotArgs != wantArgs {
		return darwinComputeBenchmarkCommandOutput{},
			errors.New(
				"unexpected command arguments",
			)
	}

	return darwinComputeBenchmarkCommandOutput{
		Stdout: []byte(expected.stdout),

		Stderr: []byte(expected.stderr),
	}, expected.err
}

const m9SingleRunJSON = `[
  {
    "build_commit": "8cf427ff",
    "build_number": 5163,
    "backends": "Metal",
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
    "backends": "Metal",
    "model_filename": "/models/m9-standard.gguf",
    "model_type": "test 7B Q4_K - Medium",
    "n_prompt": 0,
    "n_gen": 128,
    "avg_ts": 119.03,
    "samples_ts": [119.03]
  }
]`

const m9DarwinThermalNormal = `
CPU_Scheduler_Limit = 100
CPU_Available_CPUs = 16
CPU_Speed_Limit = 100
`

const m9DarwinThermalThrottled = `
CPU_Scheduler_Limit = 100
CPU_Available_CPUs = 16
CPU_Speed_Limit = 85
`

func TestMeasureDarwinLlamaBenchRun(
	t *testing.T,
) {
	runner := &fakeDarwinComputeRunner{
		commands: []fakeDarwinComputeCommand{
			{
				name: m9DarwinPMSetPath,

				args: []string{
					"-g",
					"therm",
				},

				stdout: m9DarwinThermalNormal,
			},
			{
				name: m9DarwinTimePath,

				args: []string{
					"-l",
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

				stdout: m9SingleRunJSON,

				stderr: "73805824 maximum resident set size\n",
			},
			{
				name: m9DarwinPMSetPath,

				args: []string{
					"-g",
					"therm",
				},

				stdout: m9DarwinThermalThrottled,
			},
		},
	}

	result, err :=
		measureDarwinLlamaBenchRun(
			context.Background(),
			runner,
			"/opt/llama-bench",
			"/models/m9-standard.gguf",
			16,
		)
	if err != nil {
		t.Fatal(err)
	}

	if runner.index != 3 {
		t.Fatalf(
			"commands executed = %d, want 3",
			runner.index,
		)
	}

	if result.PeakSystemRAM !=
		73805824 {
		t.Fatalf(
			"peak RAM = %d",
			result.PeakSystemRAM,
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
		result.
			Throughput.
			PromptTokensPerSecond[0] !=
			7155.74 {
		t.Fatal(
			"prompt throughput changed",
		)
	}

	if len(
		result.
			Throughput.
			GenerationTokensPerSecond,
	) != 1 ||
		result.
			Throughput.
			GenerationTokensPerSecond[0] !=
			119.03 {
		t.Fatal(
			"generation throughput changed",
		)
	}
}

func TestMeasureDarwinLlamaBenchRunLeavesThermalUnknown(
	t *testing.T,
) {
	runner := &fakeDarwinComputeRunner{
		commands: []fakeDarwinComputeCommand{
			{
				name: m9DarwinPMSetPath,

				args: []string{
					"-g",
					"therm",
				},

				err: errors.New(
					"pmset unavailable",
				),
			},
			{
				name: m9DarwinTimePath,

				args: []string{
					"-l",
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

				stdout: m9SingleRunJSON,

				stderr: "73805824 maximum resident set size\n",
			},
			{
				name: m9DarwinPMSetPath,

				args: []string{
					"-g",
					"therm",
				},

				stdout: m9DarwinThermalNormal,
			},
		},
	}

	result, err :=
		measureDarwinLlamaBenchRun(
			context.Background(),
			runner,
			"/opt/llama-bench",
			"/models/m9-standard.gguf",
			16,
		)
	if err != nil {
		t.Fatal(err)
	}

	if result.Throttled != nil {
		t.Fatal(
			"incomplete thermal observation was reported as known",
		)
	}
}

func TestMeasureDarwinLlamaBenchRunRejectsMissingPeakRAM(
	t *testing.T,
) {
	runner := &fakeDarwinComputeRunner{
		commands: []fakeDarwinComputeCommand{
			{
				name: m9DarwinPMSetPath,

				args: []string{
					"-g",
					"therm",
				},

				stdout: m9DarwinThermalNormal,
			},
			{
				name: m9DarwinTimePath,

				args: []string{
					"-l",
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

				stdout: m9SingleRunJSON,

				stderr: "123 peak memory footprint\n",
			},
		},
	}

	_, err :=
		measureDarwinLlamaBenchRun(
			context.Background(),
			runner,
			"/opt/llama-bench",
			"/models/m9-standard.gguf",
			16,
		)

	if err == nil {
		t.Fatal(
			"missing maximum RSS was accepted",
		)
	}
}

func TestMeasureDarwinLlamaBenchRunRequiresPath(
	t *testing.T,
) {
	runner := &fakeDarwinComputeRunner{}

	_, err :=
		measureDarwinLlamaBenchRun(
			context.Background(),
			runner,
			" ",
			"/models/m9-standard.gguf",
			16,
		)

	if err == nil {
		t.Fatal(
			"empty llama-bench path was accepted",
		)
	}

	if runner.index != 0 {
		t.Fatal(
			"runner was invoked for invalid path",
		)
	}
}
