package agent

import (
	"math"
	"testing"
)

func TestM13DarwinCPUParser(t *testing.T) {
	raw := []byte("CPU usage: 0.5% user, 1.0% sys, 98.5% idle\nCPU usage: 5.5% user, 10.5% sys, 84.0% idle\n")
	pct, err := m13DarwinCPUUsage(raw)
	if err != nil || math.Abs(pct-16) > 0.0001 {
		t.Fatalf("CPU=%v error=%v", pct, err)
	}
	for _, s := range []string{"", "CPU usage: broken", "CPU usage: 120% idle"} {
		if _, err := m13DarwinCPUUsage([]byte(s)); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
func TestM13DarwinMemoryParser(t *testing.T) {
	raw := []byte("Mach Virtual Memory Statistics: (page size of 4096 bytes)\nPages free: 100.\nPages active: 200.\nPages inactive: 300.\nPages speculative: 40.\nPages purgeable: 20.\n")
	available, err := m13DarwinAvailableRAM(raw)
	if err != nil || available != 460*4096 {
		t.Fatalf("memory=%d error=%v", available, err)
	}
	for _, s := range []string{"", "Pages free: 200.", "page size of 0 bytes\nPages free: 4.\nPages inactive: 3.", "page size of 4096 bytes\nPages free: -3.\nPages inactive: 10."} {
		if _, err := m13DarwinAvailableRAM([]byte(s)); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
