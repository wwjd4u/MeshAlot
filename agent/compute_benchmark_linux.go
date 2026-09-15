package agent

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	m9LinuxBenchmarkRunTimeout = 10 * time.Minute
	m9LinuxTimePath            = "/usr/bin/time"
)

var m9LinuxTemperatureGlobs = []string{
	"/sys/class/thermal/thermal_zone*/temp",
	"/sys/class/hwmon/hwmon*/temp*_input",
}

const m9LinuxThrottleGlob = "/sys/devices/system/cpu/cpu*/thermal_throttle/*_throttle_count"

type linuxComputeBenchmarkCommandOutput struct {
	Stdout []byte
	Stderr []byte
}

type linuxComputeBenchmarkRunner interface {
	Run(
		context.Context,
		string,
		...string,
	) (linuxComputeBenchmarkCommandOutput, error)
}

type linuxComputeTelemetryReader interface {
	Glob(string) ([]string, error)
	ReadFile(string) ([]byte, error)
}

type linuxComputeTelemetrySnapshot struct {
	TemperatureCelsius *float64
	ThrottleCounts     map[string]uint64
}

type linuxLlamaBenchRunResult struct {
	Throughput llamaBenchThroughput

	PeakSystemRAM      uint64
	TemperatureCelsius *float64
	Throttled          *bool
}

// parseLinuxTimePeakRSS parses GNU time -v output.
//
// Linux reports maximum resident set size in KiB. M9 stores bytes.
func parseLinuxTimePeakRSS(
	raw []byte,
) (uint64, error) {
	const prefix = "Maximum resident set size (kbytes):"

	for _, line := range strings.Split(
		string(raw),
		"\n",
	) {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(
			line,
			prefix,
		) {
			continue
		}

		value := strings.TrimSpace(
			strings.TrimPrefix(
				line,
				prefix,
			),
		)

		kib, err := strconv.ParseUint(
			value,
			10,
			64,
		)
		if err != nil || kib == 0 {
			return 0, errors.New(
				"invalid Linux peak RSS value",
			)
		}

		const bytesPerKiB uint64 = 1024

		if kib > ^uint64(0)/bytesPerKiB {
			return 0, errors.New(
				"Linux peak RSS value overflows bytes",
			)
		}

		return kib * bytesPerKiB, nil
	}

	return 0, errors.New(
		"Linux peak RSS was not found",
	)
}

func parseLinuxMilliCelsius(
	raw []byte,
) (float64, error) {
	value := strings.TrimSpace(
		string(raw),
	)

	milli, err := strconv.ParseInt(
		value,
		10,
		64,
	)
	if err != nil || milli <= 0 {
		return 0, errors.New(
			"invalid Linux temperature value",
		)
	}

	celsius := float64(milli) / 1000

	if celsius > 150 {
		return 0, errors.New(
			"Linux temperature value is out of range",
		)
	}

	return celsius, nil
}

func readLinuxComputeTelemetry(
	reader linuxComputeTelemetryReader,
) linuxComputeTelemetrySnapshot {
	if reader == nil {
		return linuxComputeTelemetrySnapshot{}
	}

	snapshot :=
		linuxComputeTelemetrySnapshot{}

	var maxTemperature *float64

	for _, pattern := range m9LinuxTemperatureGlobs {

		paths, err :=
			reader.Glob(pattern)
		if err != nil {
			continue
		}

		for _, path := range paths {
			raw, err :=
				reader.ReadFile(path)
			if err != nil {
				continue
			}

			temperature, err :=
				parseLinuxMilliCelsius(raw)
			if err != nil {
				continue
			}

			if maxTemperature == nil ||
				temperature > *maxTemperature {

				value := temperature
				maxTemperature = &value
			}
		}
	}

	snapshot.TemperatureCelsius =
		maxTemperature

	paths, err :=
		reader.Glob(
			m9LinuxThrottleGlob,
		)

	if err == nil {
		counts :=
			make(map[string]uint64)

		for _, path := range paths {
			raw, err :=
				reader.ReadFile(path)
			if err != nil {
				continue
			}

			count, err :=
				strconv.ParseUint(
					strings.TrimSpace(
						string(raw),
					),
					10,
					64,
				)
			if err != nil {
				continue
			}

			counts[path] = count
		}

		if len(counts) > 0 {
			snapshot.ThrottleCounts =
				counts
		}
	}

	return snapshot
}

// linuxBenchmarkThrottleDelta marks a run throttled only when a cumulative
// Linux thermal-throttle counter increases during that benchmark.
//
// Existing historical counts do not cause a new benchmark to be marked
// throttled.
func linuxBenchmarkThrottleDelta(
	before map[string]uint64,
	after map[string]uint64,
) *bool {
	if len(before) == 0 ||
		len(after) == 0 {

		return nil
	}

	comparable := false
	throttled := false

	for path, beforeCount := range before {

		afterCount, ok :=
			after[path]

		if !ok ||
			afterCount < beforeCount {

			continue
		}

		comparable = true

		if afterCount > beforeCount {
			throttled = true
		}
	}

	if !comparable {
		return nil
	}

	return &throttled
}

func maxOptionalTemperature(
	a *float64,
	b *float64,
) *float64 {
	switch {
	case a == nil && b == nil:
		return nil

	case a == nil:
		value := *b
		return &value

	case b == nil:
		value := *a
		return &value

	default:
		value := *a

		if *b > value {
			value = *b
		}

		return &value
	}
}

// measureLinuxLlamaBenchRun performs one independent M9 throughput run using
// injected command and telemetry sources.
//
// It performs no network access and no package/model installation.
//
// Peak RAM comes from GNU time -v. Temperature and CPU thermal-throttle
// counters are best-effort local sysfs observations.
//
// Dedicated VRAM use is deliberately not inferred here. On shared-memory GPU
// systems, capacity is not the same thing as actual benchmark VRAM use.
func measureLinuxLlamaBenchRun(
	ctx context.Context,
	runner linuxComputeBenchmarkRunner,
	telemetryReader linuxComputeTelemetryReader,
	llamaBenchPath string,
	modelPath string,
) (linuxLlamaBenchRunResult, error) {
	var result linuxLlamaBenchRunResult

	if ctx == nil {
		return result, errors.New(
			"compute benchmark context is nil",
		)
	}

	if runner == nil {
		return result, errors.New(
			"compute benchmark runner is nil",
		)
	}

	llamaBenchPath =
		strings.TrimSpace(
			llamaBenchPath,
		)

	if llamaBenchPath == "" {
		return result, errors.New(
			"llama-bench path is required",
		)
	}

	benchArgs, err :=
		m9LlamaBenchArgs(modelPath)

	if err != nil {
		return result, err
	}

	before :=
		readLinuxComputeTelemetry(
			telemetryReader,
		)

	runArgs := make(
		[]string,
		0,
		len(benchArgs)+2,
	)

	runArgs = append(
		runArgs,
		"-v",
		llamaBenchPath,
	)

	runArgs = append(
		runArgs,
		benchArgs...,
	)

	runCtx, cancel :=
		context.WithTimeout(
			ctx,
			m9LinuxBenchmarkRunTimeout,
		)
	defer cancel()

	output, err :=
		runner.Run(
			runCtx,
			m9LinuxTimePath,
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
		parseLinuxTimePeakRSS(
			output.Stderr,
		)

	if err != nil {
		return result, fmt.Errorf(
			"parse llama-bench peak RAM: %w",
			err,
		)
	}

	after :=
		readLinuxComputeTelemetry(
			telemetryReader,
		)

	result =
		linuxLlamaBenchRunResult{
			Throughput: throughput,

			PeakSystemRAM: peakRAM,

			TemperatureCelsius: maxOptionalTemperature(
				before.
					TemperatureCelsius,
				after.
					TemperatureCelsius,
			),

			Throttled: linuxBenchmarkThrottleDelta(
				before.ThrottleCounts,
				after.ThrottleCounts,
			),
		}

	return result, nil
}
