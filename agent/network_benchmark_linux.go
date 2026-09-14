//go:build linux

package agent

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	networkBenchmarkOutputLimit = 512 << 10

	networkBenchmarkRouteTimeout      = 5 * time.Second
	networkBenchmarkThroughputTimeout = 45 * time.Second
	networkBenchmarkProbeTimeout      = 8 * time.Second

	networkBenchmarkProbeCount = 10

	networkBenchmarkControlURL = "https://api.meshalot.com/v1/health"

	linuxNetworkBenchmarkDownloadBytes = 25_000_000
	linuxNetworkBenchmarkUploadBytes   = 10_000_000

	linuxNetworkBenchmarkThroughputSamples = 3
	linuxNetworkBenchmarkPercentile        = 0.90

	linuxNetworkBenchmarkDownloadURL = "https://speed.cloudflare.com/__down?bytes=25000000"

	linuxNetworkBenchmarkUploadURL = "https://speed.cloudflare.com/__up"
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

// CollectNetworkBenchmark performs one full Linux M8 benchmark.
//
// Throughput is measured against Cloudflare's download/upload test
// endpoints. Control-region availability, latency, jitter, and loss are
// measured separately against the MeshAlot control endpoint.
//
// Full throughput testing is intentionally occasional.
func CollectNetworkBenchmark(
	ctx context.Context,
) (protocol.NetworkBenchmark, error) {
	return collectNetworkBenchmarkLinux(
		ctx,
		networkBenchmarkOSRunner{},
		time.Now,
	)
}

func collectNetworkBenchmarkLinux(
	ctx context.Context,
	runner networkBenchmarkRunner,
	now func() time.Time,
) (protocol.NetworkBenchmark, error) {
	var benchmark protocol.NetworkBenchmark

	if now == nil {
		return benchmark, errors.New(
			"network benchmark clock is nil",
		)
	}

	preferredPath, err := determineLinuxPreferredPath(
		ctx,
		runner,
	)
	if err != nil {
		return benchmark, err
	}

	if strings.HasPrefix(preferredPath, "tailscale") {
		return benchmark, errors.New(
			"Tailscale path is not permitted for MeshAlot benchmark",
		)
	}

	ipv4Probe := runNetworkControlProbe(
		ctx,
		runner,
		"-4",
		preferredPath,
	)

	ipv6Probe := runNetworkControlProbe(
		ctx,
		runner,
		"-6",
		preferredPath,
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

	downloadMbps, err := collectLinuxThroughputSamples(
		ctx,
		runner,
		probeFamily,
		preferredPath,
		false,
		"",
		linuxNetworkBenchmarkDownloadBytes,
	)
	if err != nil {
		return benchmark, err
	}

	uploadPath, cleanupUpload, err :=
		createLinuxUploadPayload(
			linuxNetworkBenchmarkUploadBytes,
		)
	if err != nil {
		return benchmark, err
	}
	defer cleanupUpload()

	uploadMbps, err := collectLinuxThroughputSamples(
		ctx,
		runner,
		probeFamily,
		preferredPath,
		true,
		uploadPath,
		linuxNetworkBenchmarkUploadBytes,
	)
	if err != nil {
		return benchmark, err
	}

	// Warm the selected MeshAlot control path.
	// This probe is deliberately excluded from latency/jitter/loss.
	_ = runNetworkControlProbe(
		ctx,
		runner,
		probeFamily,
		preferredPath,
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
				preferredPath,
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
		SchemaVersion: protocol.NetworkBenchmarkSchemaVersion,
		CollectedAt:   now().UTC(),

		IPv4Available: ipv4Available,
		IPv6Available: ipv6Available,

		DownloadMbps: downloadMbps,
		UploadMbps:   uploadMbps,

		LatencyMs: latencyMs,
		JitterMs:  jitterMs,

		PacketLossPercent: packetLossPercent,
		PreferredPath:     preferredPath,
	}

	return benchmark, nil
}

func determineLinuxPreferredPath(
	ctx context.Context,
	runner networkBenchmarkRunner,
) (string, error) {
	ipv4Raw, ipv4Err := runNetworkBenchmarkCommand(
		ctx,
		runner,
		networkBenchmarkRouteTimeout,
		"/usr/sbin/ip",
		"-4",
		"route",
		"get",
		"1.1.1.1",
	)

	if ipv4Err == nil {
		if path, err := parseLinuxRouteInterface(
			ipv4Raw,
		); err == nil {
			return path, nil
		}
	}

	ipv6Raw, ipv6Err := runNetworkBenchmarkCommand(
		ctx,
		runner,
		networkBenchmarkRouteTimeout,
		"/usr/sbin/ip",
		"-6",
		"route",
		"get",
		"2606:4700:4700::1111",
	)

	if ipv6Err == nil {
		if path, err := parseLinuxRouteInterface(
			ipv6Raw,
		); err == nil {
			return path, nil
		}
	}

	return "",
		errors.New(
			"default Linux network interface not found",
		)
}

func parseLinuxRouteInterface(
	raw []byte,
) (string, error) {
	fields := strings.Fields(string(raw))

	for i := 0; i+1 < len(fields); i++ {
		if fields[i] != "dev" {
			continue
		}

		path := strings.TrimSpace(fields[i+1])

		if path != "" {
			return path, nil
		}
	}

	return "",
		errors.New(
			"Linux route output does not contain device",
		)
}

func parseLinuxThroughputResult(
	raw []byte,
	expectedBytes int64,
) (float64, error) {
	fields := strings.Fields(string(raw))

	if len(fields) != 3 {
		return 0,
			errors.New(
				"invalid Linux throughput result",
			)
	}

	statusCode, err := strconv.Atoi(fields[0])
	if err != nil ||
		statusCode < 200 ||
		statusCode > 299 {

		return 0,
			errors.New(
				"Linux throughput endpoint returned invalid HTTP status",
			)
	}

	transferredBytes, err := strconv.ParseFloat(
		fields[1],
		64,
	)
	if err != nil ||
		!finiteNonNegative(transferredBytes) {

		return 0,
			errors.New(
				"invalid Linux throughput byte count",
			)
	}

	if transferredBytes < float64(expectedBytes) {
		return 0,
			errors.New(
				"Linux throughput transfer was incomplete",
			)
	}

	speedBytesPerSecond, err := strconv.ParseFloat(
		fields[2],
		64,
	)
	if err != nil ||
		!finiteNonNegative(speedBytesPerSecond) ||
		speedBytesPerSecond <= 0 {

		return 0,
			errors.New(
				"invalid Linux throughput speed",
			)
	}

	return speedBytesPerSecond * 8.0 / 1_000_000.0,
		nil
}

func collectLinuxThroughputSamples(
	ctx context.Context,
	runner networkBenchmarkRunner,
	family string,
	preferredPath string,
	upload bool,
	uploadPath string,
	expectedBytes int64,
) (float64, error) {
	samples := make(
		[]float64,
		0,
		linuxNetworkBenchmarkThroughputSamples,
	)

	for i := 0; i < linuxNetworkBenchmarkThroughputSamples; i++ {
		var (
			raw []byte
			err error
		)

		if upload {
			raw, err = runNetworkBenchmarkCommand(
				ctx,
				runner,
				networkBenchmarkThroughputTimeout,
				"/usr/bin/curl",
				family,
				"--interface",
				preferredPath,
				"--connect-timeout",
				"5",
				"--max-time",
				"45",
				"-sS",
				"-H",
				"Content-Type: application/octet-stream",
				"--data-binary",
				"@"+uploadPath,
				"-o",
				"/dev/null",
				"-w",
				"%{http_code} %{size_upload} %{speed_upload}\n",
				linuxNetworkBenchmarkUploadURL,
			)
		} else {
			raw, err = runNetworkBenchmarkCommand(
				ctx,
				runner,
				networkBenchmarkThroughputTimeout,
				"/usr/bin/curl",
				family,
				"--interface",
				preferredPath,
				"--connect-timeout",
				"5",
				"--max-time",
				"45",
				"-sS",
				"-o",
				"/dev/null",
				"-w",
				"%{http_code} %{size_download} %{speed_download}\n",
				linuxNetworkBenchmarkDownloadURL,
			)
		}

		if err != nil {
			if upload {
				return 0, errors.New(
					"Linux upload throughput test failed",
				)
			}

			return 0, errors.New(
				"Linux download throughput test failed",
			)
		}

		mbps, err := parseLinuxThroughputResult(
			raw,
			expectedBytes,
		)
		if err != nil {
			return 0, err
		}

		samples = append(samples, mbps)
	}

	return linuxPercentile(
		samples,
		linuxNetworkBenchmarkPercentile,
	)
}

func linuxPercentile(
	values []float64,
	percentile float64,
) (float64, error) {
	if len(values) == 0 {
		return 0, errors.New(
			"Linux throughput samples are empty",
		)
	}

	if percentile < 0 || percentile > 1 {
		return 0, errors.New(
			"Linux throughput percentile is invalid",
		)
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	index := float64(len(sorted)-1) * percentile
	lower := int(index)
	upper := lower

	if float64(lower) != index {
		upper++
	}

	if lower == upper {
		return sorted[lower], nil
	}

	fraction := index - float64(lower)

	return sorted[lower] +
			(sorted[upper]-sorted[lower])*fraction,
		nil
}

func createLinuxUploadPayload(
	size int64,
) (
	string,
	func(),
	error,
) {
	if size <= 0 {
		return "",
			func() {},
			errors.New(
				"Linux upload payload size must be positive",
			)
	}

	file, err := os.CreateTemp(
		"",
		"meshalot-m8-upload-*",
	)
	if err != nil {
		return "",
			func() {},
			errors.New(
				"create Linux upload payload",
			)
	}

	path := file.Name()

	cleanup := func() {
		_ = os.Remove(path)
	}

	if err := file.Truncate(size); err != nil {
		_ = file.Close()
		cleanup()

		return "",
			func() {},
			errors.New(
				"size Linux upload payload",
			)
	}

	if err := file.Close(); err != nil {
		cleanup()

		return "",
			func() {},
			errors.New(
				"close Linux upload payload",
			)
	}

	return path, cleanup, nil
}

func runNetworkControlProbe(
	ctx context.Context,
	runner networkBenchmarkRunner,
	family string,
	preferredPath string,
) networkProbeSample {
	raw, err := runNetworkBenchmarkCommand(
		ctx,
		runner,
		networkBenchmarkProbeTimeout,
		"/usr/bin/curl",
		family,
		"--interface",
		preferredPath,
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

	// Any real HTTP response proves the control path was reachable.
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
