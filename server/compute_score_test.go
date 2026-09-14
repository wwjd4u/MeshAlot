package server

import (
	"testing"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestCalculateComputeScoreRanksFasterNodeHigher(
	t *testing.T,
) {
	slower := computeScoreFixture(
		250,
		20,
		650,
	)

	faster := computeScoreFixture(
		700,
		65,
		220,
	)

	slowResult := CalculateComputeScore(slower)
	fastResult := CalculateComputeScore(faster)

	if fastResult.Score <= slowResult.Score {
		t.Fatalf(
			"faster node score %d must exceed slower node score %d",
			fastResult.Score,
			slowResult.Score,
		)
	}

	if fastResult.MedianGenerationTokensPerSecond <=
		slowResult.MedianGenerationTokensPerSecond {
		t.Fatal(
			"generation measurement did not preserve ranking",
		)
	}
}

func TestCalculateComputeScoreUsesMedianAgainstOutlier(
	t *testing.T,
) {
	benchmark := computeScoreFixture(
		500,
		40,
		300,
	)

	benchmark.Samples = append(
		benchmark.Samples,
		protocol.ComputeBenchmarkSample{
			Run:                       4,
			PromptTokens:              512,
			GeneratedTokens:           128,
			PromptTokensPerSecond:     5000,
			GenerationTokensPerSecond: 1000,
			TimeToFirstTokenMs:        1,
			PeakSystemRAMBytes:        1,
			Success:                   true,
		},
	)

	result := CalculateComputeScore(benchmark)

	if result.MedianGenerationTokensPerSecond != 40 {
		t.Fatalf(
			"generation median = %.2f, want 40",
			result.MedianGenerationTokensPerSecond,
		)
	}

	if result.MedianPromptTokensPerSecond != 500 {
		t.Fatalf(
			"prompt median = %.2f, want 500",
			result.MedianPromptTokensPerSecond,
		)
	}
}

func TestCalculateComputeScorePenalizesFailures(
	t *testing.T,
) {
	clean := computeScoreFixture(
		500,
		40,
		300,
	)

	withFailure := computeScoreFixture(
		500,
		40,
		300,
	)

	withFailure.Samples = append(
		withFailure.Samples,
		protocol.ComputeBenchmarkSample{
			Run:     4,
			Success: false,
			Error:   "runtime failure",
		},
	)

	cleanResult := CalculateComputeScore(clean)
	failedResult := CalculateComputeScore(withFailure)

	if failedResult.SuccessPoints >= cleanResult.SuccessPoints {
		t.Fatalf(
			"failure did not reduce success points: clean %.2f failed %.2f",
			cleanResult.SuccessPoints,
			failedResult.SuccessPoints,
		)
	}

	if failedResult.Score >= cleanResult.Score {
		t.Fatalf(
			"failure did not reduce score: clean %d failed %d",
			cleanResult.Score,
			failedResult.Score,
		)
	}
}

func TestCalculateComputeScorePenalizesThrottling(
	t *testing.T,
) {
	normal := computeScoreFixture(
		500,
		40,
		300,
	)

	throttled := computeScoreFixture(
		500,
		40,
		300,
	)

	value := true
	throttled.Samples[1].Throttled = &value

	normalResult := CalculateComputeScore(normal)
	throttledResult := CalculateComputeScore(throttled)

	if !throttledResult.ThrottlingObserved {
		t.Fatal("throttling was not detected")
	}

	if throttledResult.Score >= normalResult.Score {
		t.Fatalf(
			"throttling did not reduce score: normal %d throttled %d",
			normalResult.Score,
			throttledResult.Score,
		)
	}
}

func TestCalculateComputeScoreRequiresRepeatedRunsForStability(
	t *testing.T,
) {
	benchmark := computeScoreFixture(
		500,
		40,
		300,
	)

	benchmark.Samples = benchmark.Samples[:2]

	result := CalculateComputeScore(benchmark)

	if result.StabilityPoints != 0 {
		t.Fatalf(
			"two runs received stability points: %.2f",
			result.StabilityPoints,
		)
	}
}

func TestCalculateComputeScoreRewardsConsistentRuns(
	t *testing.T,
) {
	consistent := computeScoreFixture(
		500,
		40,
		300,
	)

	variable := computeScoreFixture(
		500,
		40,
		300,
	)

	variable.Samples[0].PromptTokensPerSecond = 100
	variable.Samples[0].GenerationTokensPerSecond = 5
	variable.Samples[0].TimeToFirstTokenMs = 1500

	variable.Samples[2].PromptTokensPerSecond = 900
	variable.Samples[2].GenerationTokensPerSecond = 75
	variable.Samples[2].TimeToFirstTokenMs = 100

	consistentResult := CalculateComputeScore(
		consistent,
	)

	variableResult := CalculateComputeScore(
		variable,
	)

	if consistentResult.StabilityPoints <=
		variableResult.StabilityPoints {
		t.Fatalf(
			"consistent stability %.2f must exceed variable %.2f",
			consistentResult.StabilityPoints,
			variableResult.StabilityPoints,
		)
	}
}

func TestCalculateComputeScoreNoSuccessfulRunsIsZero(
	t *testing.T,
) {
	benchmark := protocol.ComputeBenchmark{
		Samples: []protocol.ComputeBenchmarkSample{
			{
				Run:     1,
				Success: false,
				Error:   "failed",
			},
			{
				Run:     2,
				Success: false,
				Error:   "failed",
			},
			{
				Run:     3,
				Success: false,
				Error:   "failed",
			},
		},
	}

	result := CalculateComputeScore(benchmark)

	if result.Score != 0 {
		t.Fatalf(
			"failed benchmark score = %d, want 0",
			result.Score,
		)
	}
}

func computeScoreFixture(
	promptTPS float64,
	generationTPS float64,
	ttftMs float64,
) protocol.ComputeBenchmark {
	samples := make(
		[]protocol.ComputeBenchmarkSample,
		0,
		3,
	)

	for run := 1; run <= 3; run++ {
		samples = append(
			samples,
			protocol.ComputeBenchmarkSample{
				Run: run,

				PromptTokens: 512,

				GeneratedTokens: 128,

				PromptTokensPerSecond: promptTPS,

				GenerationTokensPerSecond: generationTPS,

				TimeToFirstTokenMs: ttftMs,

				PeakSystemRAMBytes: 8 * 1024 * 1024 * 1024,

				Success: true,
			},
		)
	}

	return protocol.ComputeBenchmark{
		Samples: samples,
	}
}
