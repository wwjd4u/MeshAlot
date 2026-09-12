package server

import "math"

const (
	networkScoreIPv4Points     = 6.0
	networkScoreIPv6Points     = 4.0
	networkScoreDownloadPoints = 15.0
	networkScoreUploadPoints   = 15.0
	networkScoreLatencyPoints  = 20.0
	networkScoreJitterPoints   = 10.0
	networkScoreLossPoints     = 30.0

	networkScoreBandwidthFullMbps = 1000.0
	networkScoreLatencyBestMs     = 10.0
	networkScoreLatencyWorstMs    = 200.0
	networkScoreJitterBestMs      = 2.0
	networkScoreJitterWorstMs     = 50.0
	networkScoreLossWorstPercent  = 5.0
)

// CalculateNetworkScore calculates the authoritative server-side network score.
//
// The agent submits raw measurements only. Individual measurements remain
// available independently so future workload scheduling does not depend on
// this aggregate score alone.
func CalculateNetworkScore(
	ipv4Available bool,
	ipv6Available bool,
	downloadMbps float64,
	uploadMbps float64,
	latencyMs float64,
	jitterMs float64,
	packetLossPercent float64,
) int {
	// A node with neither IPv4 nor IPv6 connectivity has no usable network path.
	if !ipv4Available && !ipv6Available {
		return 0
	}

	score := 0.0

	if ipv4Available {
		score += networkScoreIPv4Points
	}
	if ipv6Available {
		score += networkScoreIPv6Points
	}

	score += networkScoreDownloadPoints *
		clamp01(downloadMbps/networkScoreBandwidthFullMbps)

	score += networkScoreUploadPoints *
		clamp01(uploadMbps/networkScoreBandwidthFullMbps)

	score += decreasingLinearScore(
		latencyMs,
		networkScoreLatencyBestMs,
		networkScoreLatencyWorstMs,
		networkScoreLatencyPoints,
	)

	score += decreasingLinearScore(
		jitterMs,
		networkScoreJitterBestMs,
		networkScoreJitterWorstMs,
		networkScoreJitterPoints,
	)

	// Packet loss receives a deliberately strong nonlinear penalty.
	//
	// 0% loss = full 30 points.
	// 2% loss = 10.8 points.
	// 3% loss = 4.8 points.
	// 5%+ loss = 0 points.
	lossRatio := 1.0 - clamp(packetLossPercent, 0.0, networkScoreLossWorstPercent)/
		networkScoreLossWorstPercent
	score += networkScoreLossPoints * lossRatio * lossRatio

	score = clamp(score, 0.0, 100.0)

	return int(math.Round(score))
}

func decreasingLinearScore(value, best, worst, points float64) float64 {
	if value <= best {
		return points
	}
	if value >= worst {
		return 0.0
	}

	return points * ((worst - value) / (worst - best))
}

func clamp01(value float64) float64 {
	return clamp(value, 0.0, 1.0)
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
