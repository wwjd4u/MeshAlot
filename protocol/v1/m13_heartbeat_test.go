package v1

import (
	"math"
	"testing"
	"time"
)

func m13Float(v float64) *float64 { return &v }
func m13Uint(v uint64) *uint64 { return &v }

func m13TestHeartbeat(now time.Time) M13Heartbeat {
	return M13Heartbeat{
		Type: "heartbeat", Online: true,
		GPUUtilizationPercent: m13Float(32.5),
		AvailableVRAMBytes: m13Uint(16 << 30),
		AvailableRAMBytes: m13Uint(64 << 30),
		CPULoadPercent: m13Float(48),
		ActiveJobState: "idle",
		RecentLatencyMS: m13Float(24.3),
		AvailabilityMode: "normal",
		ObservedAt: now,
	}
}

func TestM13HeartbeatValidation(t *testing.T) {
	now := time.Now().UTC()
	good := m13TestHeartbeat(now)
	if err := good.Validate(now); err != nil { t.Fatal(err) }
	noGPU := good
	noGPU.GPUUtilizationPercent = nil
	noGPU.AvailableVRAMBytes = nil
	if err := noGPU.Validate(now); err != nil { t.Fatalf("CPU-only node: %v", err) }
	paused := good
	paused.ManualPause = true
	paused.AvailabilityMode = "maximum-earnings"
	if err := paused.Validate(now); err != nil { t.Fatal(err) }
	tests := []struct {
		name string
		change func(*M13Heartbeat)
	}{
		{"offline", func(h *M13Heartbeat) {h.Online = false}},
		{"wrong frame type", func(h *M13Heartbeat) {h.Type="job"}},
		{"missing timestamp", func(h *M13Heartbeat) {h.ObservedAt=time.Time{}}},
		{"old timestamp", func(h *M13Heartbeat) {h.ObservedAt=now.Add(-6*time.Minute)}},
		{"future timestamp", func(h *M13Heartbeat) {h.ObservedAt=now.Add(6*time.Minute)}},
		{"RAM omitted", func(h *M13Heartbeat) {h.AvailableRAMBytes=nil}},
		{"RAM excessive", func(h *M13Heartbeat) {h.AvailableRAMBytes=m13Uint(M13MaxResourceBytes+1)}},
		{"CPU omitted", func(h *M13Heartbeat) {h.CPULoadPercent=nil}},
		{"CPU negative", func(h *M13Heartbeat) {h.CPULoadPercent=m13Float(-1)}},
		{"CPU over 100", func(h *M13Heartbeat) {h.CPULoadPercent=m13Float(101)}},
		{"CPU not finite", func(h *M13Heartbeat) {h.CPULoadPercent=m13Float(math.NaN())}},
		{"GPU missing VRAM", func(h *M13Heartbeat) {h.AvailableVRAMBytes=nil}},
		{"GPU over 100", func(h *M13Heartbeat) {h.GPUUtilizationPercent=m13Float(101)}},
		{"VRAM excessive", func(h *M13Heartbeat) {h.AvailableVRAMBytes=m13Uint(M13MaxResourceBytes+1)}},
		{"latency missing", func(h *M13Heartbeat) {h.RecentLatencyMS=nil}},
		{"latency negative", func(h *M13Heartbeat) {h.RecentLatencyMS=m13Float(-1)}},
		{"latency excessive", func(h *M13Heartbeat) {h.RecentLatencyMS=m13Float(M13MaxLatencyMS+1)}},
		{"bad job state", func(h *M13Heartbeat) {h.ActiveJobState="execute-shell"}},
		{"bad availability mode", func(h *M13Heartbeat) {h.AvailabilityMode="admin"}},
	}
	for _, tc := range tests {
		t.Run(tc.name,func(t *testing.T) {
			bad:=good
			tc.change(&bad)
			if err:=bad.Validate(now); err==nil {t.Fatal("accepted invalid heartbeat")}
		})
	}
}
