//go:build linux

package agent

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	m9LinuxBenchmarkOutputLimit = 2 << 20

	m9LinuxBashPath = "/bin/bash"

	m9LinuxOneAPISetvarsPath = "/opt/intel/oneapi/setvars.sh"

	m9LinuxOneAPICommandScript = `env_script="$1"; shift; command_args=("$@"); set --; source "$env_script" >/dev/null 2>&1 || exit 125; exec "${command_args[@]}"`
)

type linuxComputeBenchmarkOSRunner struct{}

// linuxOneAPIWrappedCommand creates a shell invocation that loads the
// pre-installed Intel oneAPI environment before executing the requested
// command.
//
// The executable and all command arguments are passed as positional parameters,
// not interpolated into shell source. This prevents command arguments such as a
// model filename from becoming shell syntax.
func linuxOneAPIWrappedCommand(
	name string,
	args []string,
) (
	string,
	[]string,
	error,
) {
	name =
		strings.TrimSpace(name)

	if name == "" {
		return "", nil, errors.New(
			"Linux benchmark command is empty",
		)
	}

	wrappedArgs :=
		make(
			[]string,
			0,
			len(args)+5,
		)

	wrappedArgs =
		append(
			wrappedArgs,
			"-c",
			m9LinuxOneAPICommandScript,
			"meshalot-oneapi",
			m9LinuxOneAPISetvarsPath,
			name,
		)

	wrappedArgs =
		append(
			wrappedArgs,
			args...,
		)

	return m9LinuxBashPath,
		wrappedArgs,
		nil
}

func (linuxComputeBenchmarkOSRunner) Run(
	ctx context.Context,
	name string,
	args ...string,
) (linuxComputeBenchmarkCommandOutput, error) {
	commandName,
		commandArgs,
		err :=
		linuxOneAPIWrappedCommand(
			name,
			args,
		)

	if err != nil {
		return linuxComputeBenchmarkCommandOutput{},
			err
	}

	cmd :=
		exec.CommandContext(
			ctx,
			commandName,
			commandArgs...,
		)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout =
		&m9LinuxLimitedBuffer{
			w: &stdout,
			n: m9LinuxBenchmarkOutputLimit,
		}

	cmd.Stderr =
		&m9LinuxLimitedBuffer{
			w: &stderr,
			n: m9LinuxBenchmarkOutputLimit,
		}

	err = cmd.Run()

	return linuxComputeBenchmarkCommandOutput{
		Stdout: append(
			[]byte(nil),
			stdout.Bytes()...,
		),

		Stderr: append(
			[]byte(nil),
			stderr.Bytes()...,
		),
	}, err
}

type m9LinuxLimitedBuffer struct {
	w *bytes.Buffer
	n int
}

func (b *m9LinuxLimitedBuffer) Write(
	p []byte,
) (int, error) {
	original := len(p)

	if b.n <= 0 {
		return original, nil
	}

	if len(p) > b.n {
		p = p[:b.n]
	}

	_, _ =
		b.w.Write(p)

	b.n -= len(p)

	return original, nil
}

type linuxComputeTelemetryOSReader struct{}

func (linuxComputeTelemetryOSReader) Glob(
	pattern string,
) ([]string, error) {
	return filepath.Glob(pattern)
}

func (linuxComputeTelemetryOSReader) ReadFile(
	path string,
) ([]byte, error) {
	return os.ReadFile(path)
}

// collectLinuxLlamaBenchRun is the real Linux OS-backed entry point used later
// by the M9 benchmark orchestrator.
//
// The OS runner explicitly loads the installed Intel oneAPI environment before
// starting GNU time and llama-bench. Merely compiling this function does not
// execute a benchmark or contact any service.
func collectLinuxLlamaBenchRun(
	ctx context.Context,
	llamaBenchPath string,
	modelPath string,
) (linuxLlamaBenchRunResult, error) {
	return measureLinuxLlamaBenchRun(
		ctx,
		linuxComputeBenchmarkOSRunner{},
		linuxComputeTelemetryOSReader{},
		llamaBenchPath,
		modelPath,
	)
}
