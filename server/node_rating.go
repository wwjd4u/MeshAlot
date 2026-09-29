package server

// NodeTier is the provider-facing tier assigned to a node.
//
// A tier is descriptive and must not replace the individual measurements used
// for workload scheduling.
type NodeTier string

const (
	NodeTierBasic      NodeTier = "basic"
	NodeTierStandard   NodeTier = "standard"
	NodeTierGold       NodeTier = "gold"
	NodeTierPlatinum   NodeTier = "platinum"
	NodeTierDatacenter NodeTier = "datacenter"
)

// WorkloadClass identifies a class of work with its own eligibility
// requirements.
type WorkloadClass string

const (
	WorkloadClassTextInference WorkloadClass = "text-inference"
	WorkloadClassLargeData     WorkloadClass = "large-data"
)

// NodeRating preserves the independent measurements used by the scheduler.
//
// Do not collapse these dimensions into one authoritative overall score.
// Different workload classes may care about the dimensions differently.
type NodeRating struct {
	ComputeScore      int
	NetworkScore      int
	ReliabilityScore  int
	AvailabilityScore int
	TrustScore        int

	Tier NodeTier
}

// EligibilityResult explains whether a node may receive a particular workload.
//
// Reasons explain why the node qualifies.
// Limitations explain which measurements prevent or restrict eligibility.
type EligibilityResult struct {
	Eligible bool

	Reasons     []string
	Limitations []string
}

// EvaluateNodeEligibility determines whether a node may receive a workload.
//
// Eligibility is workload-specific. Each score dimension remains an independent
// scheduling input; no aggregate overall score is used.
func EvaluateNodeEligibility(
	rating NodeRating,
	workload WorkloadClass,
) EligibilityResult {
	result := EligibilityResult{}

	switch workload {
	case WorkloadClassTextInference:
		checkMinimum(&result, "compute", rating.ComputeScore, 40)
		checkMinimum(&result, "network", rating.NetworkScore, 25)
		checkMinimum(&result, "reliability", rating.ReliabilityScore, 50)
		checkMinimum(&result, "availability", rating.AvailabilityScore, 50)
		checkMinimum(&result, "trust", rating.TrustScore, 50)

	case WorkloadClassLargeData:
		checkMinimum(&result, "compute", rating.ComputeScore, 50)
		checkMinimum(&result, "network", rating.NetworkScore, 70)
		checkMinimum(&result, "reliability", rating.ReliabilityScore, 60)
		checkMinimum(&result, "availability", rating.AvailabilityScore, 60)
		checkMinimum(&result, "trust", rating.TrustScore, 60)

	default:
		result.Limitations = append(
			result.Limitations,
			"unsupported workload class",
		)
		return result
	}

	result.Eligible = len(result.Limitations) == 0

	if result.Eligible {
		result.Reasons = append(
			result.Reasons,
			"node satisfies workload-specific requirements",
		)
	}

	return result
}

func checkMinimum(
	result *EligibilityResult,
	dimension string,
	value int,
	minimum int,
) {
	if value >= minimum {
		return
	}

	result.Limitations = append(
		result.Limitations,
		dimension+" score below workload requirement",
	)
}

// DetermineNodeTier assigns a provider-facing tier from the node's weakest
// measured dimension.
//
// Tier is descriptive only. Scheduling eligibility continues to use the
// individual Compute, Network, Reliability, Availability, and Trust scores.
func DetermineNodeTier(rating NodeRating) NodeTier {
	lowest := rating.ComputeScore

	values := []int{
		rating.NetworkScore,
		rating.ReliabilityScore,
		rating.AvailabilityScore,
		rating.TrustScore,
	}

	for _, value := range values {
		if value < lowest {
			lowest = value
		}
	}

	switch {
	case lowest >= 90:
		return NodeTierDatacenter
	case lowest >= 80:
		return NodeTierPlatinum
	case lowest >= 70:
		return NodeTierGold
	case lowest >= 50:
		return NodeTierStandard
	default:
		return NodeTierBasic
	}
}
