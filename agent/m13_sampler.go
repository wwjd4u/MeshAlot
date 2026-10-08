package agent

// M13SamplerOptions is provided from the node's local provider/job state.
// The sampler MUST NOT infer or overwrite sharing policy on its own.
type M13SamplerOptions struct {
	AvailabilityMode string
	ManualPause      bool
	ActiveJobState   string
	RecentLatencyMS  float64
}
