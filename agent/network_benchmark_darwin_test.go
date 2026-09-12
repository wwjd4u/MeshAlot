//go:build darwin

package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type fakeNetworkCommandResult struct {
	output string
	err    error
}

type fakeNetworkBenchmarkRunner struct {
	ipv4 []fakeNetworkCommandResult
	ipv6 []fakeNetworkCommandResult

	ipv4Index int
	ipv6Index int
}

func (f *fakeNetworkBenchmarkRunner) Run(
	_ context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	switch name {
	case "/sbin/route":
		return []byte(
			"   route to: default\n" +
				"destination: default\n" +
				"    gateway: 172.20.10.1\n" +
				"  interface: en0\n",
		), nil

	case "/usr/bin/networkQuality":
		if !containsArgument(args, "-I") ||
			!containsArgument(args, "en0") ||
			!containsArgument(args, "-c") {

			return nil, errors.New(
				"unexpected networkQuality arguments",
			)
		}

		return []byte(`{
			"base_rtt": 51.50921630859375,
			"dl_throughput": 42611076,
			"interface_name": "en0",
			"ul_throughput": 13718485
		}`), nil

	case "/usr/bin/curl":
		switch {
		case containsArgument(args, "-4"):
			return f.nextIPv4()
		case containsArgument(args, "-6"):
			return f.nextIPv6()
		default:
			return nil, errors.New(
				"curl family was not forced",
			)
		}

	default:
		return nil, fmt.Errorf(
			"unexpected command: %s",
			name,
		)
	}
}

func (f *fakeNetworkBenchmarkRunner) nextIPv4() (
	[]byte,
	error,
) {
	if f.ipv4Index >= len(f.ipv4) {
		return nil, errors.New(
			"unexpected extra IPv4 probe",
		)
	}

	result := f.ipv4[f.ipv4Index]
	f.ipv4Index++

	return []byte(result.output), result.err
}

func (f *fakeNetworkBenchmarkRunner) nextIPv6() (
	[]byte,
	error,
) {
	if f.ipv6Index >= len(f.ipv6) {
		return nil, errors.New(
			"unexpected extra IPv6 probe",
		)
	}

	result := f.ipv6[f.ipv6Index]
	f.ipv6Index++

	return []byte(result.output), result.err
}

func containsArgument(
	args []string,
	want string,
) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}

	return false
}

func TestCollectNetworkBenchmarkDarwin(t *testing.T) {
	runner := &fakeNetworkBenchmarkRunner{
		ipv4: []fakeNetworkCommandResult{
			// IPv4 availability.
			{output: "200 0.050000\n"},

			// Warm-up. This large value must not affect results.
			{output: "200 0.900000\n"},

			// Ten measured probes.
			{output: "200 0.050000\n"},
			{output: "200 0.055000\n"},
			{output: "200 0.060000\n"},
			{
				output: "000 0.000000\n",
				err:    errors.New("timeout"),
			},
			{output: "200 0.065000\n"},
			{output: "200 0.070000\n"},
			{output: "200 0.075000\n"},
			{output: "200 0.080000\n"},
			{output: "200 0.085000\n"},
			{output: "200 0.090000\n"},
		},

		ipv6: []fakeNetworkCommandResult{
			// IPv6 availability.
			{output: "200 0.060000\n"},
		},
	}

	collectedAt := time.Date(
		2026,
		time.September,
		12,
		14,
		0,
		0,
		0,
		time.UTC,
	)

	benchmark, err := collectNetworkBenchmarkDarwin(
		context.Background(),
		runner,
		func() time.Time {
			return collectedAt
		},
	)
	if err != nil {
		t.Fatalf(
			"collectNetworkBenchmarkDarwin() error = %v",
			err,
		)
	}

	if benchmark.SchemaVersion !=
		protocol.NetworkBenchmarkSchemaVersion {

		t.Fatalf(
			"schema = %q, want %q",
			benchmark.SchemaVersion,
			protocol.NetworkBenchmarkSchemaVersion,
		)
	}

	if !benchmark.CollectedAt.Equal(collectedAt) {
		t.Fatalf(
			"collected_at = %v, want %v",
			benchmark.CollectedAt,
			collectedAt,
		)
	}

	if !benchmark.IPv4Available {
		t.Fatal("IPv4 should be available")
	}

	if !benchmark.IPv6Available {
		t.Fatal("IPv6 should be available")
	}

	if !almostEqual(
		benchmark.DownloadMbps,
		42.611076,
		0.000001,
	) {
		t.Fatalf(
			"download = %f, want 42.611076",
			benchmark.DownloadMbps,
		)
	}

	if !almostEqual(
		benchmark.UploadMbps,
		13.718485,
		0.000001,
	) {
		t.Fatalf(
			"upload = %f, want 13.718485",
			benchmark.UploadMbps,
		)
	}

	if !almostEqual(
		benchmark.LatencyMs,
		70,
		0.000001,
	) {
		t.Fatalf(
			"latency = %f, want 70",
			benchmark.LatencyMs,
		)
	}

	if !almostEqual(
		benchmark.JitterMs,
		5,
		0.000001,
	) {
		t.Fatalf(
			"jitter = %f, want 5",
			benchmark.JitterMs,
		)
	}

	if !almostEqual(
		benchmark.PacketLossPercent,
		10,
		0.000001,
	) {
		t.Fatalf(
			"packet loss = %f, want 10",
			benchmark.PacketLossPercent,
		)
	}

	if benchmark.PreferredPath != "en0" {
		t.Fatalf(
			"preferred path = %q, want en0",
			benchmark.PreferredPath,
		)
	}

	if runner.ipv4Index != 12 {
		t.Fatalf(
			"IPv4 probes = %d, want 12",
			runner.ipv4Index,
		)
	}

	if runner.ipv6Index != 1 {
		t.Fatalf(
			"IPv6 probes = %d, want 1",
			runner.ipv6Index,
		)
	}
}

func TestParseDarwinDefaultInterface(t *testing.T) {
	raw := []byte(`
   route to: default
destination: default
    gateway: 172.20.10.1
  interface: en0
`)

	got, err := parseDarwinDefaultInterface(raw)
	if err != nil {
		t.Fatalf(
			"parseDarwinDefaultInterface() error = %v",
			err,
		)
	}

	if got != "en0" {
		t.Fatalf(
			"interface = %q, want en0",
			got,
		)
	}
}

func TestRunNetworkControlProbeAcceptsHTTPErrorResponse(
	t *testing.T,
) {
	runner := &fakeSingleNetworkRunner{
		output: "503 0.055000\n",
	}

	sample := runNetworkControlProbe(
		context.Background(),
		runner,
		"-4",
	)

	if !sample.Success {
		t.Fatal(
			"HTTP 503 with successful connection was treated as packet loss",
		)
	}

	if sample.ConnectDuration != 55*time.Millisecond {
		t.Fatalf(
			"duration = %v, want 55ms",
			sample.ConnectDuration,
		)
	}
}

type fakeSingleNetworkRunner struct {
	output string
	err    error
}

func (f *fakeSingleNetworkRunner) Run(
	_ context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	if name != "/usr/bin/curl" {
		return nil, fmt.Errorf(
			"unexpected command %q",
			name,
		)
	}

	if !strings.Contains(
		strings.Join(args, " "),
		networkBenchmarkControlURL,
	) {
		return nil, errors.New(
			"unexpected control URL",
		)
	}

	return []byte(f.output), f.err
}
