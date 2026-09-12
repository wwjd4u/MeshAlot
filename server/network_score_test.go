package server

import "testing"

func TestCalculateNetworkScoreExpectedExamples(t *testing.T) {
	tests := []struct {
		name              string
		ipv4              bool
		ipv6              bool
		downloadMbps      float64
		uploadMbps        float64
		latencyMs         float64
		jitterMs          float64
		packetLossPercent float64
		want              int
	}{
		{
			name:              "perfect reference connection",
			ipv4:              true,
			ipv6:              true,
			downloadMbps:      1000,
			uploadMbps:        1000,
			latencyMs:         10,
			jitterMs:          2,
			packetLossPercent: 0,
			want:              100,
		},
		{
			name:              "M8 excellent design example",
			ipv4:              true,
			ipv6:              true,
			downloadMbps:      946,
			uploadMbps:        891,
			latencyMs:         21,
			jitterMs:          1.8,
			packetLossPercent: 0,
			want:              96,
		},
		{
			name:              "upload constrained to 38 Mbps",
			ipv4:              true,
			ipv6:              true,
			downloadMbps:      200,
			uploadMbps:        38,
			latencyMs:         30,
			jitterMs:          5,
			packetLossPercent: 0.2,
			want:              68,
		},
		{
			name:              "high packet loss strongly penalized",
			ipv4:              true,
			ipv6:              true,
			downloadMbps:      500,
			uploadMbps:        500,
			latencyMs:         25,
			jitterMs:          5,
			packetLossPercent: 3,
			want:              58,
		},
		{
			name:              "poor single stack connection",
			ipv4:              true,
			ipv6:              false,
			downloadMbps:      25,
			uploadMbps:        5,
			latencyMs:         150,
			jitterMs:          30,
			packetLossPercent: 5,
			want:              16,
		},
		{
			name:              "no IP connectivity",
			ipv4:              false,
			ipv6:              false,
			downloadMbps:      1000,
			uploadMbps:        1000,
			latencyMs:         1,
			jitterMs:          0,
			packetLossPercent: 0,
			want:              0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CalculateNetworkScore(
				test.ipv4,
				test.ipv6,
				test.downloadMbps,
				test.uploadMbps,
				test.latencyMs,
				test.jitterMs,
				test.packetLossPercent,
			)

			if got != test.want {
				t.Fatalf("CalculateNetworkScore() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestCalculateNetworkScoreRespondsToConnectionChanges(t *testing.T) {
	baseline := CalculateNetworkScore(
		true, true,
		500, 500,
		25, 5,
		0,
	)

	uploadConstrained := CalculateNetworkScore(
		true, true,
		500, 38,
		25, 5,
		0,
	)

	highLoss := CalculateNetworkScore(
		true, true,
		500, 500,
		25, 5,
		3,
	)

	if uploadConstrained >= baseline {
		t.Fatalf(
			"constraining upload did not reduce score: baseline=%d constrained=%d",
			baseline,
			uploadConstrained,
		)
	}

	if highLoss >= baseline {
		t.Fatalf(
			"packet loss did not reduce score: baseline=%d high_loss=%d",
			baseline,
			highLoss,
		)
	}
}
