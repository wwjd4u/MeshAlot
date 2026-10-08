package v1

import (
	"errors"
	"math"
	"time"
)

// M13Heartbeat is a single operational snapshot over an authenticated channel.
// Deliberately no node_id: only the proven connection identity selects the node.
// A nil GPU utilization/VRAM pair means the node has no reporting GPU.
type M13Heartbeat struct {
	Type                  string    `json:"type"`
	Online                bool      `json:"online"`
	GPUUtilizationPercent *float64  `json:"gpu_utilization_percent"`
	AvailableVRAMBytes    *uint64   `json:"available_vram_bytes"`
	AvailableRAMBytes     *uint64   `json:"available_ram_bytes"`
	CPULoadPercent        *float64  `json:"cpu_load_percent"`
	ActiveJobState        string    `json:"active_job_state"`
	RecentLatencyMS       *float64  `json:"recent_latency_ms"`
	AvailabilityMode      string    `json:"availability_mode"`
	ManualPause           bool      `json:"manual_pause"`
	ObservedAt            time.Time `json:"observed_at"`
}

const (
	M13MaxHeartbeatClockSkew = 5 * time.Minute
	M13MaxResourceBytes      = uint64(1) << 50
	M13MaxLatencyMS          = 60_000
)

// Validate rejects malformed, stale, or implausible operational data.
// Heartbeat status time is ALWAYS persisted using the server clock.
func (h M13Heartbeat) Validate(now time.Time) error {
	if h.Type != "heartbeat" || !h.Online {
		return errors.New("invalid heartbeat state")
	}
	if h.ObservedAt.IsZero() {
		return errors.New("missing heartbeat timestamp")
	}
	skew := now.UTC().Sub(h.ObservedAt.UTC())
	if skew < 0 {
		skew = -skew
	}
	if skew > M13MaxHeartbeatClockSkew {
		return errors.New("heartbeat timestamp outside allowed window")
	}
	if h.AvailableRAMBytes == nil || *h.AvailableRAMBytes > M13MaxResourceBytes {
		return errors.New("invalid available RAM")
	}
	if h.CPULoadPercent == nil || !m13Percent(*h.CPULoadPercent) {
		return errors.New("invalid CPU load percentage")
	}
	if h.RecentLatencyMS == nil || math.IsNaN(*h.RecentLatencyMS) ||
		math.IsInf(*h.RecentLatencyMS, 0) ||
		*h.RecentLatencyMS < 0 || *h.RecentLatencyMS > M13MaxLatencyMS {
		return errors.New("invalid latency")
	}
	if (h.GPUUtilizationPercent == nil) != (h.AvailableVRAMBytes == nil) {
		return errors.New("GPU and VRAM must both be present or absent")
	}
	if h.GPUUtilizationPercent != nil && (!m13Percent(*h.GPUUtilizationPercent) ||
		*h.AvailableVRAMBytes > M13MaxResourceBytes) {
		return errors.New("invalid GPU telemetry")
	}
	switch h.ActiveJobState {
	case "idle", "running", "stopping", "unknown":
	default:
		return errors.New("invalid active job state")
	}
	switch h.AvailabilityMode {
	case "normal", "away", "maximum-earnings":
	default:
		return errors.New("invalid availability mode")
	}
	return nil
}

func m13Percent(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 100
}
