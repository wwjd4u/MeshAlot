//go:build darwin

package agent

import (
	"errors"
	"strconv"
	"strings"
)

// parseDarwinProcessRSS parses the resident-memory value produced by:
//
//	ps -o rss= -p <pid>
//
// macOS reports RSS in KiB. M9 records the value in bytes.
func parseDarwinProcessRSS(
	raw []byte,
) (uint64, error) {
	fields := strings.Fields(string(raw))

	if len(fields) != 1 {
		return 0, errors.New(
			"invalid macOS process RSS output",
		)
	}

	kib, err := strconv.ParseUint(
		fields[0],
		10,
		64,
	)
	if err != nil || kib == 0 {
		return 0, errors.New(
			"invalid macOS process RSS value",
		)
	}

	const bytesPerKiB uint64 = 1024

	if kib > ^uint64(0)/bytesPerKiB {
		return 0, errors.New(
			"macOS process RSS value overflows bytes",
		)
	}

	return kib * bytesPerKiB, nil
}

// parseDarwinThermalState interprets:
//
//	pmset -g therm
//
// A benchmark is considered throttled when macOS reduces scheduler capacity,
// CPU speed, or the number of CPUs available to the workload.
//
// Temperature is deliberately not inferred here. On this Mac the reliable
// powermetrics temperature path requires superuser access, so M9 leaves
// temperature optional rather than inventing a value.
func parseDarwinThermalState(
	raw []byte,
	logicalCPUs int,
) (bool, error) {
	if logicalCPUs <= 0 {
		return false, errors.New(
			"logical CPU count is invalid",
		)
	}

	var (
		schedulerLimit int
		availableCPUs  int
		speedLimit     int

		haveScheduler bool
		haveAvailable bool
		haveSpeed     bool
	)

	for _, line := range strings.Split(
		string(raw),
		"\n",
	) {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		parts := strings.SplitN(
			line,
			"=",
			2,
		)

		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		parsed, err := strconv.Atoi(value)
		if err != nil {
			continue
		}

		switch key {
		case "CPU_Scheduler_Limit":
			schedulerLimit = parsed
			haveScheduler = true

		case "CPU_Available_CPUs":
			availableCPUs = parsed
			haveAvailable = true

		case "CPU_Speed_Limit":
			speedLimit = parsed
			haveSpeed = true
		}
	}

	if !haveScheduler ||
		!haveAvailable ||
		!haveSpeed {
		return false, errors.New(
			"macOS thermal state is incomplete",
		)
	}

	if schedulerLimit < 0 ||
		schedulerLimit > 100 ||
		speedLimit < 0 ||
		speedLimit > 100 ||
		availableCPUs <= 0 ||
		availableCPUs > logicalCPUs {
		return false, errors.New(
			"macOS thermal state contains invalid limits",
		)
	}

	throttled :=
		schedulerLimit < 100 ||
			speedLimit < 100 ||
			availableCPUs < logicalCPUs

	return throttled, nil
}

// parseDarwinTimePeakRSS extracts the maximum resident set size from:
//
//	/usr/bin/time -l <command>
//
// On macOS this value is reported in bytes.
func parseDarwinTimePeakRSS(
	raw []byte,
) (uint64, error) {
	for _, line := range strings.Split(
		string(raw),
		"\n",
	) {
		line = strings.TrimSpace(line)

		if !strings.HasSuffix(
			line,
			"maximum resident set size",
		) {
			continue
		}

		value := strings.TrimSpace(
			strings.TrimSuffix(
				line,
				"maximum resident set size",
			),
		)

		bytes, err := strconv.ParseUint(
			value,
			10,
			64,
		)
		if err != nil || bytes == 0 {
			return 0, errors.New(
				"invalid macOS maximum resident set size",
			)
		}

		return bytes, nil
	}

	return 0, errors.New(
		"macOS maximum resident set size not found",
	)
}
