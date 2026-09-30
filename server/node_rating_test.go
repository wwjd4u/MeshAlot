package server

import "testing"

func TestEvaluateNodeEligibilityDiffersByWorkload(t *testing.T) {
	highComputeSlowNetwork := NodeRating{
		ComputeScore:      90,
		NetworkScore:      35,
		ReliabilityScore:  85,
		AvailabilityScore: 85,
		TrustScore:        85,
		Tier:              NodeTierGold,
	}
	markAllEvidenceKnown(&highComputeSlowNetwork)

	textResult := EvaluateNodeEligibility(
		highComputeSlowNetwork,
		WorkloadClassTextInference,
	)

	largeDataResult := EvaluateNodeEligibility(
		highComputeSlowNetwork,
		WorkloadClassLargeData,
	)

	if !textResult.Eligible {
		t.Fatalf(
			"text inference should remain eligible: limitations=%v",
			textResult.Limitations,
		)
	}

	if largeDataResult.Eligible {
		t.Fatal(
			"large-data workload should not be eligible on slow network",
		)
	}
}

func TestEvaluateNodeEligibilityAllowsFastNetworkLowerComputeForLargeData(
	t *testing.T,
) {
	lowerComputeFastNetwork := NodeRating{
		ComputeScore:      55,
		NetworkScore:      90,
		ReliabilityScore:  85,
		AvailabilityScore: 85,
		TrustScore:        85,
		Tier:              NodeTierStandard,
	}
	markAllEvidenceKnown(&lowerComputeFastNetwork)

	result := EvaluateNodeEligibility(
		lowerComputeFastNetwork,
		WorkloadClassLargeData,
	)

	if !result.Eligible {
		t.Fatalf(
			"large-data workload should be eligible: limitations=%v",
			result.Limitations,
		)
	}
}

func TestEvaluateNodeEligibilityUsesIndependentDimensions(
	t *testing.T,
) {
	node := NodeRating{
		ComputeScore:      95,
		NetworkScore:      95,
		ReliabilityScore:  20,
		AvailabilityScore: 90,
		TrustScore:        90,
		Tier:              NodeTierPlatinum,
	}
	markAllEvidenceKnown(&node)

	result := EvaluateNodeEligibility(
		node,
		WorkloadClassTextInference,
	)

	if result.Eligible {
		t.Fatal(
			"high compute and network must not hide inadequate reliability",
		)
	}

	if len(result.Limitations) == 0 {
		t.Fatal("expected reliability limitation")
	}
}

func TestEvaluateNodeEligibilityRejectsUnknownWorkload(
	t *testing.T,
) {
	node := NodeRating{
		ComputeScore:      100,
		NetworkScore:      100,
		ReliabilityScore:  100,
		AvailabilityScore: 100,
		TrustScore:        100,
		Tier:              NodeTierDatacenter,
	}

	result := EvaluateNodeEligibility(
		node,
		WorkloadClass("unknown"),
	)

	if result.Eligible {
		t.Fatal("unknown workload class must not be eligible")
	}

	if len(result.Limitations) == 0 {
		t.Fatal("unknown workload should explain its limitation")
	}
}

func TestDetermineNodeTierUsesAllDimensions(t *testing.T) {
	tests := []struct {
		name   string
		rating NodeRating
		want   NodeTier
	}{
		{
			name: "datacenter",
			rating: NodeRating{
				ComputeScore:      95,
				NetworkScore:      94,
				ReliabilityScore:  96,
				AvailabilityScore: 97,
				TrustScore:        98,
			},
			want: NodeTierDatacenter,
		},
		{
			name: "platinum",
			rating: NodeRating{
				ComputeScore:      88,
				NetworkScore:      84,
				ReliabilityScore:  86,
				AvailabilityScore: 85,
				TrustScore:        89,
			},
			want: NodeTierPlatinum,
		},
		{
			name: "gold",
			rating: NodeRating{
				ComputeScore:      78,
				NetworkScore:      76,
				ReliabilityScore:  72,
				AvailabilityScore: 75,
				TrustScore:        79,
			},
			want: NodeTierGold,
		},
		{
			name: "standard",
			rating: NodeRating{
				ComputeScore:      65,
				NetworkScore:      60,
				ReliabilityScore:  55,
				AvailabilityScore: 62,
				TrustScore:        68,
			},
			want: NodeTierStandard,
		},
		{
			name: "weak dimension prevents higher tier",
			rating: NodeRating{
				ComputeScore:      95,
				NetworkScore:      95,
				ReliabilityScore:  40,
				AvailabilityScore: 95,
				TrustScore:        95,
			},
			want: NodeTierBasic,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rating := test.rating
			markAllEvidenceKnown(&rating)

			got := DetermineNodeTier(rating)

			if got != test.want {
				t.Fatalf(
					"DetermineNodeTier() = %q, want %q",
					got,
					test.want,
				)
			}
		})
	}
}

func TestEvaluateNodeEligibilityRejectsUnknownDimensionEvidence(
	t *testing.T,
) {
	node := NodeRating{
		ComputeScore:      90,
		NetworkScore:      90,
		ReliabilityScore:  90,
		AvailabilityScore: 90,
		TrustScore:        90,

		ComputeKnown:      true,
		NetworkKnown:      true,
		ReliabilityKnown:  false,
		AvailabilityKnown: true,
		TrustKnown:        true,
	}

	result := EvaluateNodeEligibility(
		node,
		WorkloadClassTextInference,
	)

	if result.Eligible {
		t.Fatal(
			"node with unknown reliability evidence must not be eligible",
		)
	}

	if len(result.Limitations) == 0 {
		t.Fatal("expected unknown-evidence limitation")
	}
}

func markAllEvidenceKnown(rating *NodeRating) {
	rating.ComputeKnown = true
	rating.NetworkKnown = true
	rating.ReliabilityKnown = true
	rating.AvailabilityKnown = true
	rating.TrustKnown = true
}

func TestDetermineNodeTierIsUnratedWithUnknownEvidence(
	t *testing.T,
) {
	node := NodeRating{
		ComputeScore:      95,
		NetworkScore:      95,
		ReliabilityScore:  95,
		AvailabilityScore: 95,
		TrustScore:        95,

		ComputeKnown:      true,
		NetworkKnown:      true,
		ReliabilityKnown:  false,
		AvailabilityKnown: true,
		TrustKnown:        true,
	}

	got := DetermineNodeTier(node)

	if got != NodeTierUnrated {
		t.Fatalf(
			"DetermineNodeTier() = %q, want %q",
			got,
			NodeTierUnrated,
		)
	}
}
