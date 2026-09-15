//go:build linux

package agent

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

const m9LinuxBenchmarkOutputLimit = 2 << 20

type linuxComputeBenchmarkOSRunner struct{}

func (linuxComputeBenchmarkOSRunner) Run(
	ctx context.Context,
	name string,
	args ...string,
) (linuxComputeBenchmarkCommandOutput, error) {
	cmd :=
		exec.CommandContext(
			ctx,
			name,
			args...,
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

	err := cmd.Run()

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

	_, _ = b.w.Write(p)

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
// Merely compiling this function does not execute a benchmark or contact any
// service.
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
