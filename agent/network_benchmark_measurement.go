package agent

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

type networkQualityResult struct {
	DownloadBitsPerSecond float64 `json:"dl_throughput"`
	UploadBitsPerSecond   float64 `json:"ul_throughput"`
	InterfaceName         string  `json:"interface_name"`
}

func parseNetworkQualityResult(
	raw []byte,
) (
	downloadMbps float64,
	uploadMbps float64,
	interfaceName string,
	err error,
) {
	if strings.TrimSpace(string(raw)) == "" {
		return 0, 0, "", errors.New("networkQuality output is empty")
	}

	var result networkQualityResult

	if err := json.Unmarshal(raw, &result); err != nil {
		return 0, 0, "", errors.New("invalid networkQuality JSON")
	}

	if !finiteNonNegative(result.DownloadBitsPerSecond) {
		return 0, 0, "", errors.New("invalid download throughput")
	}

	if !finiteNonNegative(result.UploadBitsPerSecond) {
		return 0, 0, "", errors.New("invalid upload throughput")
	}

	interfaceName = strings.TrimSpace(result.InterfaceName)
	if interfaceName == "" {
		return 0, 0, "", errors.New("networkQuality interface is empty")
	}

	return result.DownloadBitsPerSecond / 1_000_000.0,
		result.UploadBitsPerSecond / 1_000_000.0,
		interfaceName,
		nil
}

type networkProbeSample struct {
	ConnectDuration time.Duration
	Success         bool
}

// summarizeNetworkProbes calculates control-path measurements.
//
// Latency is the median successful TCP connection time.
// Jitter is the mean absolute change between consecutive successful
// connection times.
// Packet loss is the percentage of control probes that failed.
//
// ICMP is deliberately not used because the MeshAlot control endpoint
// has already been proven reachable by HTTPS while not answering ICMP.
func summarizeNetworkProbes(
	samples []networkProbeSample,
) (
	latencyMs float64,
	jitterMs float64,
	packetLossPercent float64,
	err error,
) {
	if len(samples) == 0 {
		return 0, 0, 0, errors.New("no control probes supplied")
	}

	successful := make([]float64, 0, len(samples))

	for _, sample := range samples {
		if !sample.Success {
			continue
		}

		if sample.ConnectDuration <= 0 {
			return 0, 0, 0, errors.New(
				"successful control probe has invalid duration",
			)
		}

		successful = append(
			successful,
			float64(sample.ConnectDuration)/float64(time.Millisecond),
		)
	}

	failed := len(samples) - len(successful)

	packetLossPercent =
		(float64(failed) / float64(len(samples))) * 100.0

	if len(successful) == 0 {
		return 0,
			0,
			packetLossPercent,
			errors.New("no successful control probes")
	}

	sorted := append([]float64(nil), successful...)
	sort.Float64s(sorted)

	middle := len(sorted) / 2

	if len(sorted)%2 == 0 {
		latencyMs = (sorted[middle-1] + sorted[middle]) / 2.0
	} else {
		latencyMs = sorted[middle]
	}

	if len(successful) > 1 {
		var totalVariation float64

		for i := 1; i < len(successful); i++ {
			totalVariation += math.Abs(
				successful[i] - successful[i-1],
			)
		}

		jitterMs =
			totalVariation / float64(len(successful)-1)
	}

	return latencyMs, jitterMs, packetLossPercent, nil
}

func finiteNonNegative(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0
}
