//go:build linux

package agent

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

type m13StubRunner struct {
	data []byte
	err  error
}

func (f m13StubRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	if name != "nvidia-smi" || len(args) != 2 {
		return nil, errors.New("unexpected telemetry command")
	}
	return f.data, f.err
}

func TestM13LinuxCPUAndMemoryParsers(t *testing.T) {
	total, idle, err := m13CPUStat([]byte("cpu  100 0 50 850 0 0 0 0 0 0\n"))
	if err != nil || total != 1000 || idle != 850 {
		t.Fatalf("CPU stats=%d,%d,%v", total, idle, err)
	}
	for _, s := range []string{"", "intr 0", "cpu 1 2 3", "cpu 1 2 xx 4 5 6 7 8"} {
		if _, _, err := m13CPUStat([]byte(s)); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
	ram, err := m13MemAvailable([]byte("MemTotal:  200000 kB\nMemAvailable: 65536 kB\n"))
	if err != nil || ram != 65536*1024 {
		t.Fatalf("RAM=%d error=%v", ram, err)
	}
	if _, err := m13MemAvailable([]byte("MemTotal: 1000 kB\n")); err == nil {
		t.Fatal("accepted absent available RAM")
	}
}
func TestM13LinuxNvidiaSnapshot(t *testing.T) {
	gpu, free, ok := m13NvidiaSnapshot([]byte("18, 20000\n"))
	if !ok || gpu != 18 || free != 20000*(1<<20) {
		t.Fatalf("GPU=%g VRAM=%d ok=%v", gpu, free, ok)
	}
	for _, bad := range []string{"", "N/A, 100", "101, 500", "10, N/A", "-1, 100", "50, 9999999999999999"} {
		if _, _, ok := m13NvidiaSnapshot([]byte(bad)); ok {
			t.Fatalf("accepted %q", bad)
		}
	}
	if _, _, ok := m13NvidiaSnapshot([]byte(strings.Repeat("Z", 60))); ok {
		t.Fatal("accepted junk")
	}
}
func TestM13LinuxSamplerWithIsolatedCounters(t *testing.T) {
	counter := 0
	read := func(path string) ([]byte, error) {
		switch path {
		case "/proc/stat":
			counter++
			if counter == 1 {
				return []byte("cpu 100 0 50 850 0 0 0 0\n"), nil
			}
			return []byte("cpu 120 0 60 920 0 0 0 0\n"), nil
		case "/proc/meminfo":
			return []byte("MemAvailable: 65536 kB\n"), nil
		default:
			return nil, errors.New("unexpected file")
		}
	}
	h, err := sampleM13Linux(context.Background(),
		M13SamplerOptions{AvailabilityMode: "normal", ActiveJobState: "idle", RecentLatencyMS: 12.5},
		read, m13StubRunner{data: []byte("18, 20000\n")}, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	if h.CPULoadPercent == nil || math.Abs(*h.CPULoadPercent-30) > 0.01 {
		t.Fatalf("unexpected CPU load: %+v", h.CPULoadPercent)
	}
	if h.AvailableRAMBytes == nil || *h.AvailableRAMBytes != 65536*1024 {
		t.Fatalf("bad RAM: %+v", h.AvailableRAMBytes)
	}
	if h.GPUUtilizationPercent == nil || *h.GPUUtilizationPercent != 18 ||
		h.AvailableVRAMBytes == nil || *h.AvailableVRAMBytes != 20000*(1<<20) {
		t.Fatalf("bad GPU counters: %+v", h)
	}
}
func TestM13LinuxSamplerMissingGPUIsUnknown(t *testing.T) {
	counter := 0
	read := func(path string) ([]byte, error) {
		switch path {
		case "/proc/stat":
			counter++
			if counter == 1 {
				return []byte("cpu 10 0 0 90 0 0 0 0\n"), nil
			}
			return []byte("cpu 15 0 0 95 0 0 0 0\n"), nil
		case "/proc/meminfo":
			return []byte("MemAvailable: 4096 kB\n"), nil
		}
		return nil, errors.New("invalid path")
	}
	h, err := sampleM13Linux(context.Background(),
		M13SamplerOptions{AvailabilityMode: "away", ActiveJobState: "unknown", RecentLatencyMS: 1},
		read, m13StubRunner{err: errors.New("GPU unavailable")}, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	if h.GPUUtilizationPercent != nil || h.AvailableVRAMBytes != nil {
		t.Fatal("missing GPU must not be reported as zero")
	}
}
func TestM13LinuxHostSamplerSmoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h, err := SampleM13Heartbeat(ctx, M13SamplerOptions{
		AvailabilityMode: "normal", ActiveJobState: "unknown", RecentLatencyMS: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("host snapshot: CPU=%.1f%% availableRAM=%d GPUreported=%t (passive only)",
		*h.CPULoadPercent, *h.AvailableRAMBytes, h.GPUUtilizationPercent != nil)
}
