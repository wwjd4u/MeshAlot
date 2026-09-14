//go:build darwin

package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	m9DarwinBenchmarkOutputLimit = 2 << 20
	m9DarwinBenchmarkRunTimeout  = 10 * time.Minute

	m9DarwinTimePath  = "/usr/bin/time"
	m9DarwinPMSetPath = "/usr/bin/pmset"
)

type darwinComputeBenchmarkCommandOutput struct {
	Stdout []byte
	Stderr []byte
}

type darwinComputeBenchmarkRunner interface {
	Run(
		context.Context,
		string,
		...string,
	) (darwinComputeBenchmarkCommandOutput, error)
}

type darwinComputeBenchmarkOSRunner struct{}

func (darwinComputeBenchmarkOSRunner) Run(
	ctx context.Context,
	name string,
	args ...string,
) (darwinComputeBenchmarkCommandOutput, error) {
	cmd := exec.CommandContext(
		ctx,
		name,
		args...,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &m9DarwinLimitedBuffer{
		w: &stdout,
		n: m9DarwinBenchmarkOutputLimit,
	}

	cmd.Stderr = &m9DarwinLimitedBuffer{
		w: &stderr,
		n: m9DarwinBenchmarkOutputLimit,
	}

	err := cmd.Run()

	return darwinComputeBenchmarkCommandOutput{
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

type m9DarwinLimitedBuffer struct {
	w *bytes.Buffer
	n int
}

func (b *m9DarwinLimitedBuffer) Write(
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

type darwinLlamaBenchRunResult struct {
	Throughput llamaBenchThroughput

	PeakSystemRAM uint64
	Throttled     *bool
}

// measureDarwinLlamaBenchRun performs one independent M9 throughput run.
//
// It deliberately does not perform TTFT measurement. TTFT uses the separate
// loopback-only llama.cpp server helper and is combined with this result by
// the higher-level benchmark orchestrator.
//
// Peak RAM is measured by /usr/bin/time -l around the llama-bench process.
// Thermal state is sampled immediately before and after the process. If either
// thermal snapshot is unavailable, throttling remains unknown rather than
// inventing a value.
func measureDarwinLlamaBenchRun(
	ctx context.Context,
	runner darwinComputeBenchmarkRunner,
	llamaBenchPath string,
	modelPath string,
	logicalCPUs int,
) (darwinLlamaBenchRunResult, error) {
	var result darwinLlamaBenchRunResult

	if runner == nil {
		return result, errors.New(
			"compute benchmark runner is nil",
		)
	}

	llamaBenchPath =
		strings.TrimSpace(llamaBenchPath)

	if llamaBenchPath == "" {
		return result, errors.New(
			"llama-bench path is required",
		)
	}

	if logicalCPUs <= 0 {
		return result, errors.New(
			"logical CPU count is invalid",
		)
	}

	benchArgs, err :=
		m9LlamaBenchArgs(modelPath)
	if err != nil {
		return result, err
	}

	preThrottle,
		preThrottleOK :=
		readDarwinBenchmarkThrottle(
			ctx,
			runner,
			logicalCPUs,
		)

	runArgs := make(
		[]string,
		0,
		len(benchArgs)+2,
	)

	runArgs = append(
		runArgs,
		"-l",
		llamaBenchPath,
	)

	runArgs = append(
		runArgs,
		benchArgs...,
	)

	runCtx, cancel :=
		context.WithTimeout(
			ctx,
			m9DarwinBenchmarkRunTimeout,
		)
	defer cancel()

	output, err := runner.Run(
		runCtx,
		m9DarwinTimePath,
		runArgs...,
	)
	if err != nil {
		return result, fmt.Errorf(
			"llama-bench execution failed: %w",
			err,
		)
	}

	throughput, err :=
		parseLlamaBenchJSON(
			output.Stdout,
		)
	if err != nil {
		return result, fmt.Errorf(
			"parse llama-bench output: %w",
			err,
		)
	}

	peakRAM, err :=
		parseDarwinTimePeakRSS(
			output.Stderr,
		)
	if err != nil {
		return result, fmt.Errorf(
			"parse llama-bench peak RAM: %w",
			err,
		)
	}

	postThrottle,
		postThrottleOK :=
		readDarwinBenchmarkThrottle(
			ctx,
			runner,
			logicalCPUs,
		)

	var throttled *bool

	if preThrottleOK &&
		postThrottleOK {
		value :=
			preThrottle ||
				postThrottle

		throttled = &value
	}

	result = darwinLlamaBenchRunResult{
		Throughput: throughput,

		PeakSystemRAM: peakRAM,

		Throttled: throttled,
	}

	return result, nil
}

func readDarwinBenchmarkThrottle(
	ctx context.Context,
	runner darwinComputeBenchmarkRunner,
	logicalCPUs int,
) (bool, bool) {
	runCtx, cancel :=
		context.WithTimeout(
			ctx,
			5*time.Second,
		)
	defer cancel()

	output, err := runner.Run(
		runCtx,
		m9DarwinPMSetPath,
		"-g",
		"therm",
	)
	if err != nil {
		return false, false
	}

	raw := output.Stdout

	if len(raw) == 0 {
		raw = output.Stderr
	}

	throttled, err :=
		parseDarwinThermalState(
			raw,
			logicalCPUs,
		)
	if err != nil {
		return false, false
	}

	return throttled, true
}

// collectDarwinLlamaBenchRun is the real OS-backed entry point used later by
// the benchmark orchestrator.
func collectDarwinLlamaBenchRun(
	ctx context.Context,
	llamaBenchPath string,
	modelPath string,
) (darwinLlamaBenchRunResult, error) {
	return measureDarwinLlamaBenchRun(
		ctx,
		darwinComputeBenchmarkOSRunner{},
		llamaBenchPath,
		modelPath,
		runtime.NumCPU(),
	)
}
