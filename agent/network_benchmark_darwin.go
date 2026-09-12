//go:build darwin

package agent

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	networkBenchmarkOutputLimit = 512 << 10

	networkBenchmarkRouteTimeout   = 5 * time.Second
	networkBenchmarkQualityTimeout = 35 * time.Second
	networkBenchmarkProbeTimeout   = 8 * time.Second

	networkBenchmarkProbeCount = 10

	networkBenchmarkControlURL = "https://api.meshalot.com/v1/health"
)

type networkBenchmarkRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type networkBenchmarkOSRunner struct{}

func (networkBenchmarkOSRunner) Run(
	ctx context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &networkBenchmarkLimitedBuffer{
		w: &stdout,
		n: networkBenchmarkOutputLimit,
	}

	cmd.Stderr = &networkBenchmarkLimitedBuffer{
		w: &stderr,
		n: networkBenchmarkOutputLimit,
	}

	err := cmd.Run()

	if stdout.Len() > 0 {
		return append([]byte(nil), stdout.Bytes()...), err
	}

	return append([]byte(nil), stderr.Bytes()...), err
}

type networkBenchmarkLimitedBuffer struct {
	w *bytes.Buffer
	n int
}

func (b *networkBenchmarkLimitedBuffer) Write(
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

func runNetworkBenchmarkCommand(
	ctx context.Context,
	runner networkBenchmarkRunner,
	timeout time.Duration,
	name string,
	args ...string,
) ([]byte, error) {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return runner.Run(runCtx, name, args...)
}

// CollectNetworkBenchmark performs one full macOS M8 benchmark.
//
// This is intentionally a full benchmark because networkQuality performs
// throughput testing. Scheduling logic must ensure this is run only
// occasionally. Lightweight control-path probing can be separated later
// from full throughput collection.
func CollectNetworkBenchmark(
	ctx context.Context,
) (protocol.NetworkBenchmark, error) {
	return collectNetworkBenchmarkDarwin(
		ctx,
		networkBenchmarkOSRunner{},
		time.Now,
	)
}

func collectNetworkBenchmarkDarwin(
	ctx context.Context,
	runner networkBenchmarkRunner,
	now func() time.Time,
) (protocol.NetworkBenchmark, error) {
	var benchmark protocol.NetworkBenchmark

	if now == nil {
		return benchmark, errors.New("network benchmark clock is nil")
	}

	routeRaw, err := runNetworkBenchmarkCommand(
		ctx,
		runner,
		networkBenchmarkRouteTimeout,
		"/sbin/route",
		"-n",
		"get",
		"default",
	)
	if err != nil {
		return benchmark, errors.New(
			"determine preferred network interface",
		)
	}

	preferredPath, err := parseDarwinDefaultInterface(routeRaw)
	if err != nil {
		return benchmark, err
	}

	qualityRaw, err := runNetworkBenchmarkCommand(
		ctx,
		runner,
		networkBenchmarkQualityTimeout,
		"/usr/bin/networkQuality",
		"-I",
		preferredPath,
		"-M",
		"30",
		"-c",
	)
	if err != nil {
		return benchmark, errors.New(
			"networkQuality throughput test failed",
		)
	}

	downloadMbps,
		uploadMbps,
		qualityInterface,
		err := parseNetworkQualityResult(qualityRaw)

	if err != nil {
		return benchmark, err
	}

	if qualityInterface != preferredPath {
		return benchmark, errors.New(
			"networkQuality used unexpected interface",
		)
	}

	ipv4Probe := runNetworkControlProbe(
		ctx,
		runner,
		"-4",
	)

	ipv6Probe := runNetworkControlProbe(
		ctx,
		runner,
		"-6",
	)

	ipv4Available := ipv4Probe.Success
	ipv6Available := ipv6Probe.Success

	if !ipv4Available && !ipv6Available {
		return benchmark, errors.New(
			"MeshAlot control region unavailable over IPv4 and IPv6",
		)
	}

	probeFamily := "-4"

	if !ipv4Available {
		probeFamily = "-6"
	}

	// Warm the chosen control path. This probe is deliberately excluded
	// from latency/jitter/loss because DNS, neighbor discovery, connection
	// setup, or radio wake-up can heavily distort the first measurement.
	_ = runNetworkControlProbe(
		ctx,
		runner,
		probeFamily,
	)

	samples := make(
		[]networkProbeSample,
		0,
		networkBenchmarkProbeCount,
	)

	for i := 0; i < networkBenchmarkProbeCount; i++ {
		samples = append(
			samples,
			runNetworkControlProbe(
				ctx,
				runner,
				probeFamily,
			),
		)
	}

	latencyMs,
		jitterMs,
		packetLossPercent,
		err := summarizeNetworkProbes(samples)

	if err != nil {
		return benchmark, err
	}

	benchmark = protocol.NetworkBenchmark{
		SchemaVersion:     protocol.NetworkBenchmarkSchemaVersion,
		CollectedAt:       now().UTC(),
		IPv4Available:     ipv4Available,
		IPv6Available:     ipv6Available,
		DownloadMbps:      downloadMbps,
		UploadMbps:        uploadMbps,
		LatencyMs:         latencyMs,
		JitterMs:          jitterMs,
		PacketLossPercent: packetLossPercent,
		PreferredPath:     preferredPath,
	}

	return benchmark, nil
}

func parseDarwinDefaultInterface(
	raw []byte,
) (string, error) {
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "interface:") {
			continue
		}

		value := strings.TrimSpace(
			strings.TrimPrefix(line, "interface:"),
		)

		if value != "" {
			return value, nil
		}
	}

	return "", errors.New(
		"default network interface not found",
	)
}

func runNetworkControlProbe(
	ctx context.Context,
	runner networkBenchmarkRunner,
	family string,
) networkProbeSample {
	raw, err := runNetworkBenchmarkCommand(
		ctx,
		runner,
		networkBenchmarkProbeTimeout,
		"/usr/bin/curl",
		family,
		"--connect-timeout",
		"3",
		"--max-time",
		"6",
		"-sS",
		"-o",
		"/dev/null",
		"-w",
		"%{http_code} %{time_connect}\n",
		networkBenchmarkControlURL,
	)

	if err != nil {
		return networkProbeSample{
			Success: false,
		}
	}

	fields := strings.Fields(string(raw))

	if len(fields) != 2 {
		return networkProbeSample{
			Success: false,
		}
	}

	statusCode, err := strconv.Atoi(fields[0])
	if err != nil {
		return networkProbeSample{
			Success: false,
		}
	}

	connectSeconds, err := strconv.ParseFloat(
		fields[1],
		64,
	)
	if err != nil ||
		!finiteNonNegative(connectSeconds) ||
		connectSeconds <= 0 {

		return networkProbeSample{
			Success: false,
		}
	}

	// Any real HTTP response proves that the TCP/TLS control path was
	// reachable. Do not confuse a temporary server-side HTTP error with
	// packet loss on the customer's network.
	if statusCode < 100 || statusCode > 599 {
		return networkProbeSample{
			Success: false,
		}
	}

	return networkProbeSample{
		ConnectDuration: time.Duration(
			connectSeconds * float64(time.Second),
		),
		Success: true,
	}
}
