package agent

import "fmt"

// ProviderMode describes the owner's desired provider posture.
type ProviderMode string

const (
	ProviderModeNormal          ProviderMode = "normal"
	ProviderModeAway            ProviderMode = "away"
	ProviderModeMaximumEarnings ProviderMode = "maximum-earnings"
)

// LocalReturnBehavior describes how new marketplace work should be handled
// when local activity returns to the provider machine.
type LocalReturnBehavior string

const (
	LocalReturnStopNewWork LocalReturnBehavior = "stop-new-work"
	LocalReturnContinue    LocalReturnBehavior = "continue"
)

// ProviderResourceControls contains the owner's hard limits and availability
// choices for marketplace work.
type ProviderResourceControls struct {
	MaxGPUPercent       int
	MaxVRAMBytes        uint64
	MaxRAMBytes         uint64
	MaxCPUPercent       int
	AllowedStartHour    int
	AllowedEndHour      int
	ManualPause         bool
	Mode                ProviderMode
	LocalReturnBehavior LocalReturnBehavior
}

// ProviderWorkloadResources describes the resources requested by prospective
// marketplace work.
type ProviderWorkloadResources struct {
	GPUPercent int
	VRAMBytes  uint64
	RAMBytes   uint64
	CPUPercent int
}

// ProviderResourceDecision reports whether prospective work is compatible
// with the provider's configured resource controls.
type ProviderResourceDecision struct {
	Allowed     bool
	Limitations []string
}

// ValidateProviderResourceControls checks that provider controls contain valid
// percentage and hour bounds and recognized behavior values.
func ValidateProviderResourceControls(controls ProviderResourceControls) error {
	if controls.MaxGPUPercent < 0 || controls.MaxGPUPercent > 100 {
		return fmt.Errorf("max GPU percent must be between 0 and 100")
	}
	if controls.MaxCPUPercent < 0 || controls.MaxCPUPercent > 100 {
		return fmt.Errorf("max CPU percent must be between 0 and 100")
	}
	if controls.AllowedStartHour < 0 || controls.AllowedStartHour > 23 {
		return fmt.Errorf("allowed start hour must be between 0 and 23")
	}
	if controls.AllowedEndHour < 0 || controls.AllowedEndHour > 23 {
		return fmt.Errorf("allowed end hour must be between 0 and 23")
	}
	switch controls.Mode {
	case ProviderModeNormal, ProviderModeAway, ProviderModeMaximumEarnings:
	default:
		return fmt.Errorf("unsupported provider mode %q", controls.Mode)
	}
	switch controls.LocalReturnBehavior {
	case LocalReturnStopNewWork, LocalReturnContinue:
	default:
		return fmt.Errorf("unsupported local-return behavior %q", controls.LocalReturnBehavior)
	}
	return nil
}

// EvaluateProviderResourceRequest applies the owner's hard resource ceilings
// to a prospective marketplace workload.
func EvaluateProviderResourceRequest(
	controls ProviderResourceControls,
	request ProviderWorkloadResources,
) ProviderResourceDecision {
	limitations := make([]string, 0, 4)

	if request.GPUPercent < 0 || request.GPUPercent > 100 {
		limitations = append(limitations, "requested GPU percent is outside 0-100")
	} else if request.GPUPercent > controls.MaxGPUPercent {
		limitations = append(limitations, "requested GPU exceeds provider limit")
	}
	if request.VRAMBytes > controls.MaxVRAMBytes {
		limitations = append(limitations, "requested VRAM exceeds provider limit")
	}
	if request.RAMBytes > controls.MaxRAMBytes {
		limitations = append(limitations, "requested RAM exceeds provider limit")
	}
	if request.CPUPercent < 0 || request.CPUPercent > 100 {
		limitations = append(limitations, "requested CPU percent is outside 0-100")
	} else if request.CPUPercent > controls.MaxCPUPercent {
		limitations = append(limitations, "requested CPU exceeds provider limit")
	}

	return ProviderResourceDecision{
		Allowed:     len(limitations) == 0,
		Limitations: limitations,
	}
}

// EvaluateProviderNewWork enforces provider-control validity and manual pause
// before evaluating a prospective workload against resource limits.
func EvaluateProviderNewWork(
	controls ProviderResourceControls,
	request ProviderWorkloadResources,
) ProviderResourceDecision {
	if err := ValidateProviderResourceControls(controls); err != nil {
		return ProviderResourceDecision{
			Allowed: false,
			Limitations: []string{
				"invalid provider controls: " + err.Error(),
			},
		}
	}
	if controls.ManualPause {
		return ProviderResourceDecision{
			Allowed: false,
			Limitations: []string{
				"provider sharing is manually paused",
			},
		}
	}
	return EvaluateProviderResourceRequest(controls, request)
}

// IsProviderHourAllowed reports whether the supplied local wall-clock hour is
// inside the owner's configured sharing window.
//
// Equal start/end hours mean sharing is allowed all day.
// A start hour later than the end hour represents an overnight window.
func IsProviderHourAllowed(
	controls ProviderResourceControls,
	localHour int,
) bool {
	if localHour < 0 || localHour > 23 {
		return false
	}
	start := controls.AllowedStartHour
	end := controls.AllowedEndHour
	if start == end {
		return true
	}
	if start < end {
		return localHour >= start && localHour < end
	}
	return localHour >= start || localHour < end
}

// EvaluateProviderNewWorkAtHour applies the owner's sharing schedule before
// evaluating the remaining new-work admission controls.
func EvaluateProviderNewWorkAtHour(
	controls ProviderResourceControls,
	request ProviderWorkloadResources,
	localHour int,
) ProviderResourceDecision {
	if !IsProviderHourAllowed(controls, localHour) {
		return ProviderResourceDecision{
			Allowed: false,
			Limitations: []string{
				"outside provider allowed hours",
			},
		}
	}
	return EvaluateProviderNewWork(controls, request)
}

// EvaluateProviderNewWorkWithActivity applies the owner's local-return policy
// before admitting prospective new marketplace work.
//
// localUserActive means local activity has returned to the provider machine.
// This affects admission of new work only; it does not terminate running work.
func EvaluateProviderNewWorkWithActivity(
	controls ProviderResourceControls,
	request ProviderWorkloadResources,
	localHour int,
	localUserActive bool,
) ProviderResourceDecision {
	if localUserActive &&
		controls.LocalReturnBehavior == LocalReturnStopNewWork {
		return ProviderResourceDecision{
			Allowed: false,
			Limitations: []string{
				"local user active; provider configured to stop new work",
			},
		}
	}
	return EvaluateProviderNewWorkAtHour(
		controls,
		request,
		localHour,
	)
}

// ActiveProviderWorkAction describes what the agent should do with marketplace
// work that is already running.
type ActiveProviderWorkAction string

const (
	ActiveProviderWorkContinue  ActiveProviderWorkAction = "continue"
	ActiveProviderWorkTerminate ActiveProviderWorkAction = "terminate"
)

// EvaluateActiveProviderWork protects already-running work from abrupt
// termination when provider sharing conditions change.
//
// Manual pause, allowed-hours changes, provider mode changes, and local-user
// return affect admission of new work. They do not by themselves terminate
// work already in progress.
//
// Running work may be terminated only when required for safety or when
// termination has been explicitly authorized.
func EvaluateActiveProviderWork(
	safetyTerminationRequired bool,
	terminationExplicitlyAllowed bool,
) ActiveProviderWorkAction {
	if safetyTerminationRequired || terminationExplicitlyAllowed {
		return ActiveProviderWorkTerminate
	}
	return ActiveProviderWorkContinue
}
