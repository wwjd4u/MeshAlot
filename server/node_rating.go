package server

// NodeTier is the provider-facing tier assigned to a node.
//
// A tier is descriptive and must not replace the individual measurements used
// for workload scheduling.
type NodeTier string

const (
	NodeTierUnrated    NodeTier = "unrated"
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

	ComputeKnown      bool
	NetworkKnown      bool
	ReliabilityKnown  bool
	AvailabilityKnown bool
	TrustKnown        bool

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
		checkKnownMinimum(&result, "compute", rating.ComputeScore, rating.ComputeKnown, 40)
		checkKnownMinimum(&result, "network", rating.NetworkScore, rating.NetworkKnown, 25)
		checkKnownMinimum(&result, "reliability", rating.ReliabilityScore, rating.ReliabilityKnown, 50)
		checkKnownMinimum(&result, "availability", rating.AvailabilityScore, rating.AvailabilityKnown, 50)
		checkKnownMinimum(&result, "trust", rating.TrustScore, rating.TrustKnown, 50)

	case WorkloadClassLargeData:
		checkKnownMinimum(&result, "compute", rating.ComputeScore, rating.ComputeKnown, 50)
		checkKnownMinimum(&result, "network", rating.NetworkScore, rating.NetworkKnown, 70)
		checkKnownMinimum(&result, "reliability", rating.ReliabilityScore, rating.ReliabilityKnown, 60)
		checkKnownMinimum(&result, "availability", rating.AvailabilityScore, rating.AvailabilityKnown, 60)
		checkKnownMinimum(&result, "trust", rating.TrustScore, rating.TrustKnown, 60)

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

func checkKnownMinimum(
	result *EligibilityResult,
	dimension string,
	value int,
	known bool,
	minimum int,
) {
	if !known {
		result.Limitations = append(
			result.Limitations,
			dimension+" score evidence unavailable",
		)
		return
	}

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
	if !rating.ComputeKnown ||
		!rating.NetworkKnown ||
		!rating.ReliabilityKnown ||
		!rating.AvailabilityKnown ||
		!rating.TrustKnown {
		return NodeTierUnrated
	}

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
