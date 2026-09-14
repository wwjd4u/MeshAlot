//go:build linux

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

type fakeLinuxNetworkCommandResult struct {
	output string
	err    error
}

type fakeLinuxNetworkBenchmarkRunner struct {
	ipv4 []fakeLinuxNetworkCommandResult
	ipv6 []fakeLinuxNetworkCommandResult

	ipv4Index int
	ipv6Index int
}

func (f *fakeLinuxNetworkBenchmarkRunner) Run(
	_ context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	switch name {
	case "/usr/sbin/ip":
		return []byte(
			"1.1.1.1 via 192.168.4.1 dev eno1 " +
				"src 192.168.5.131 uid 1000\n",
		), nil

	case "/usr/bin/curl":
		joined := strings.Join(args, " ")

		switch {
		case strings.Contains(
			joined,
			linuxNetworkBenchmarkDownloadURL,
		):
			return []byte(
				"200 25000000 5000000\n",
			), nil

		case strings.Contains(
			joined,
			linuxNetworkBenchmarkUploadURL,
		):
			return []byte(
				"200 10000000 2000000\n",
			), nil

		case linuxContainsArgument(args, "-4"):
			return f.nextIPv4()

		case linuxContainsArgument(args, "-6"):
			return f.nextIPv6()

		default:
			return nil,
				errors.New(
					"curl family was not forced",
				)
		}

	default:
		return nil,
			fmt.Errorf(
				"unexpected command: %s",
				name,
			)
	}
}

func (f *fakeLinuxNetworkBenchmarkRunner) nextIPv4() (
	[]byte,
	error,
) {
	if f.ipv4Index >= len(f.ipv4) {
		return nil,
			errors.New(
				"unexpected extra IPv4 probe",
			)
	}

	result := f.ipv4[f.ipv4Index]
	f.ipv4Index++

	return []byte(result.output), result.err
}

func (f *fakeLinuxNetworkBenchmarkRunner) nextIPv6() (
	[]byte,
	error,
) {
	if f.ipv6Index >= len(f.ipv6) {
		return nil,
			errors.New(
				"unexpected extra IPv6 probe",
			)
	}

	result := f.ipv6[f.ipv6Index]
	f.ipv6Index++

	return []byte(result.output), result.err
}

func linuxContainsArgument(
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

func TestCollectNetworkBenchmarkLinux(t *testing.T) {
	runner := &fakeLinuxNetworkBenchmarkRunner{
		ipv4: []fakeLinuxNetworkCommandResult{
			// IPv4 availability.
			{output: "200 0.050000\n"},

			// Warm-up.
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

		ipv6: []fakeLinuxNetworkCommandResult{
			{
				output: "000 0.000000\n",
				err:    errors.New("IPv6 unavailable"),
			},
		},
	}

	collectedAt := time.Date(
		2026,
		time.September,
		13,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	benchmark, err := collectNetworkBenchmarkLinux(
		context.Background(),
		runner,
		func() time.Time {
			return collectedAt
		},
	)
	if err != nil {
		t.Fatalf(
			"collectNetworkBenchmarkLinux() error = %v",
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

	if benchmark.IPv6Available {
		t.Fatal("IPv6 should be unavailable")
	}

	if !almostEqual(
		benchmark.DownloadMbps,
		40,
		0.000001,
	) {
		t.Fatalf(
			"download = %f, want 40",
			benchmark.DownloadMbps,
		)
	}

	if !almostEqual(
		benchmark.UploadMbps,
		16,
		0.000001,
	) {
		t.Fatalf(
			"upload = %f, want 16",
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

	if benchmark.PreferredPath != "eno1" {
		t.Fatalf(
			"preferred path = %q, want eno1",
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

func TestParseLinuxRouteInterface(t *testing.T) {
	raw := []byte(
		"1.1.1.1 via 192.168.4.1 dev eno1 " +
			"src 192.168.5.131 uid 1000\n",
	)

	got, err := parseLinuxRouteInterface(raw)
	if err != nil {
		t.Fatalf(
			"parseLinuxRouteInterface() error = %v",
			err,
		)
	}

	if got != "eno1" {
		t.Fatalf(
			"interface = %q, want eno1",
			got,
		)
	}
}

func TestParseLinuxThroughputResult(t *testing.T) {
	got, err := parseLinuxThroughputResult(
		[]byte("200 25000000 5000000\n"),
		25_000_000,
	)
	if err != nil {
		t.Fatalf(
			"parseLinuxThroughputResult() error = %v",
			err,
		)
	}

	if !almostEqual(got, 40, 0.000001) {
		t.Fatalf(
			"throughput = %f, want 40 Mbps",
			got,
		)
	}
}

func TestParseLinuxThroughputRejectsPartialTransfer(
	t *testing.T,
) {
	_, err := parseLinuxThroughputResult(
		[]byte("200 1000 5000000\n"),
		25_000_000,
	)

	if err == nil {
		t.Fatal(
			"partial Linux throughput transfer was accepted",
		)
	}
}

func TestLinuxPercentile(t *testing.T) {
	got, err := linuxPercentile(
		[]float64{
			100,
			200,
			300,
		},
		0.90,
	)
	if err != nil {
		t.Fatalf(
			"linuxPercentile() error = %v",
			err,
		)
	}

	if !almostEqual(got, 280, 0.000001) {
		t.Fatalf(
			"90th percentile = %f, want 280",
			got,
		)
	}
}

func TestLinuxPercentileRejectsEmptySamples(
	t *testing.T,
) {
	_, err := linuxPercentile(
		nil,
		0.90,
	)

	if err == nil {
		t.Fatal(
			"empty Linux throughput sample set was accepted",
		)
	}
}

func TestLinuxControlProbeAcceptsHTTPErrorResponse(
	t *testing.T,
) {
	runner := &fakeLinuxSingleNetworkRunner{
		output: "503 0.055000\n",
	}

	sample := runNetworkControlProbe(
		context.Background(),
		runner,
		"-4",
		"eno1",
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

type fakeLinuxSingleNetworkRunner struct {
	output string
	err    error
}

func (f *fakeLinuxSingleNetworkRunner) Run(
	_ context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	if name != "/usr/bin/curl" {
		return nil,
			fmt.Errorf(
				"unexpected command %q",
				name,
			)
	}

	if !strings.Contains(
		strings.Join(args, " "),
		networkBenchmarkControlURL,
	) {
		return nil,
			errors.New(
				"unexpected control URL",
			)
	}

	if !linuxContainsArgument(args, "--interface") ||
		!linuxContainsArgument(args, "eno1") {

		return nil,
			errors.New(
				"control probe did not bind eno1",
			)
	}

	return []byte(f.output), f.err
}
