package server

import (
	"math"
	"sort"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	computeScorePromptPoints     = 20.0
	computeScoreGenerationPoints = 40.0
	computeScoreTTFTPoints       = 15.0
	computeScoreStabilityPoints  = 15.0
	computeScoreSuccessPoints    = 10.0

	// These are M9 v1 saturation points, not hardware requirements.
	// Measurements beyond these values remain preserved in the benchmark.
	computeScorePromptFullTPS     = 1000.0
	computeScoreGenerationFullTPS = 100.0

	computeScoreTTFTBestMs  = 100.0
	computeScoreTTFTWorstMs = 5000.0

	computeScoreStableCV = 0.05
	computeScorePoorCV   = 0.30

	computeScoreThrottleMultiplier = 0.90
)

// ComputeScoreBreakdown preserves the reasons behind the aggregate score.
// Later dashboard/scheduler work can explain a node's ranking without relying
// on a single opaque number.
type ComputeScoreBreakdown struct {
	Score int

	SuccessfulRuns int
	TotalRuns      int

	MedianPromptTokensPerSecond     float64
	MedianGenerationTokensPerSecond float64
	MedianTimeToFirstTokenMs        float64

	PromptPoints     float64
	GenerationPoints float64
	TTFTPoints       float64
	StabilityPoints  float64
	SuccessPoints    float64

	VariationCoefficient float64
	ThrottlingObserved   bool
}

// CalculateComputeScore calculates the authoritative M9 Compute Score.
//
// Only successful benchmark samples contribute performance measurements.
// Failed samples remain important because they reduce the success component.
// The agent never supplies the authoritative score.
func CalculateComputeScore(
	benchmark protocol.ComputeBenchmark,
) ComputeScoreBreakdown {
	result := ComputeScoreBreakdown{
		TotalRuns: len(benchmark.Samples),
	}

	if len(benchmark.Samples) == 0 {
		return result
	}

	promptRates := make([]float64, 0, len(benchmark.Samples))
	generationRates := make([]float64, 0, len(benchmark.Samples))
	ttfts := make([]float64, 0, len(benchmark.Samples))

	for _, sample := range benchmark.Samples {
		if sample.Throttled != nil && *sample.Throttled {
			result.ThrottlingObserved = true
		}

		if !sample.Success {
			continue
		}

		result.SuccessfulRuns++

		promptRates = append(
			promptRates,
			sample.PromptTokensPerSecond,
		)

		generationRates = append(
			generationRates,
			sample.GenerationTokensPerSecond,
		)

		ttfts = append(
			ttfts,
			sample.TimeToFirstTokenMs,
		)
	}

	if result.SuccessfulRuns == 0 {
		return result
	}

	result.MedianPromptTokensPerSecond =
		medianFloat64(promptRates)

	result.MedianGenerationTokensPerSecond =
		medianFloat64(generationRates)

	result.MedianTimeToFirstTokenMs =
		medianFloat64(ttfts)

	result.PromptPoints =
		computeScorePromptPoints *
			clamp01(
				result.MedianPromptTokensPerSecond/
					computeScorePromptFullTPS,
			)

	result.GenerationPoints =
		computeScoreGenerationPoints *
			clamp01(
				result.MedianGenerationTokensPerSecond/
					computeScoreGenerationFullTPS,
			)

	result.TTFTPoints = decreasingLinearScore(
		result.MedianTimeToFirstTokenMs,
		computeScoreTTFTBestMs,
		computeScoreTTFTWorstMs,
		computeScoreTTFTPoints,
	)

	result.SuccessPoints =
		computeScoreSuccessPoints *
			(float64(result.SuccessfulRuns) /
				float64(result.TotalRuns))

	result.VariationCoefficient =
		computeBenchmarkVariationCoefficient(
			promptRates,
			generationRates,
			ttfts,
		)

	// M9 requires repeated measurements and understood variance.
	// One or two successful runs do not establish stability.
	if result.SuccessfulRuns >= 3 {
		result.StabilityPoints =
			decreasingLinearScore(
				result.VariationCoefficient,
				computeScoreStableCV,
				computeScorePoorCV,
				computeScoreStabilityPoints,
			)
	}

	total :=
		result.PromptPoints +
			result.GenerationPoints +
			result.TTFTPoints +
			result.StabilityPoints +
			result.SuccessPoints

	if result.ThrottlingObserved {
		total *= computeScoreThrottleMultiplier
	}

	result.Score = int(
		math.Round(
			clamp(total, 0.0, 100.0),
		),
	)

	return result
}

func computeBenchmarkVariationCoefficient(
	promptRates []float64,
	generationRates []float64,
	ttfts []float64,
) float64 {
	values := []float64{
		coefficientOfVariation(promptRates),
		coefficientOfVariation(generationRates),
		coefficientOfVariation(ttfts),
	}

	total := 0.0

	for _, value := range values {
		total += value
	}

	return total / float64(len(values))
}

func coefficientOfVariation(
	values []float64,
) float64 {
	if len(values) < 2 {
		return 1.0
	}

	mean := 0.0

	for _, value := range values {
		mean += value
	}

	mean /= float64(len(values))

	if mean <= 0 {
		return 1.0
	}

	variance := 0.0

	for _, value := range values {
		delta := value - mean
		variance += delta * delta
	}

	variance /= float64(len(values))

	return math.Sqrt(variance) / mean
}

func medianFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64(nil), values...)

	sort.Float64s(sorted)

	middle := len(sorted) / 2

	if len(sorted)%2 == 1 {
		return sorted[middle]
	}

	return (sorted[middle-1] + sorted[middle]) / 2
}
