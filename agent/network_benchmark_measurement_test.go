package agent

import (
	"math"
	"testing"
	"time"
)

func TestParseNetworkQualityResultRecordedM8Fixture(t *testing.T) {
	raw := []byte(`{
		"base_rtt": 51.50921630859375,
		"dl_throughput": 42611076,
		"interface_name": "en0",
		"ul_throughput": 13718485
	}`)

	download,
		upload,
		path,
		err := parseNetworkQualityResult(raw)

	if err != nil {
		t.Fatalf("parseNetworkQualityResult() error = %v", err)
	}

	if !almostEqual(download, 42.611076, 0.000001) {
		t.Fatalf("download Mbps = %f, want 42.611076", download)
	}

	if !almostEqual(upload, 13.718485, 0.000001) {
		t.Fatalf("upload Mbps = %f, want 13.718485", upload)
	}

	if path != "en0" {
		t.Fatalf("interface = %q, want en0", path)
	}
}

func TestSummarizeNetworkProbes(t *testing.T) {
	samples := []networkProbeSample{
		{
			ConnectDuration: 50 * time.Millisecond,
			Success:         true,
		},
		{
			ConnectDuration: 70 * time.Millisecond,
			Success:         true,
		},
		{
			ConnectDuration: 60 * time.Millisecond,
			Success:         true,
		},
		{
			Success: false,
		},
		{
			ConnectDuration: 80 * time.Millisecond,
			Success:         true,
		},
	}

	latency,
		jitter,
		loss,
		err := summarizeNetworkProbes(samples)

	if err != nil {
		t.Fatalf("summarizeNetworkProbes() error = %v", err)
	}

	if !almostEqual(latency, 65.0, 0.000001) {
		t.Fatalf("latency = %f, want 65", latency)
	}

	if !almostEqual(jitter, 16.6666667, 0.000001) {
		t.Fatalf(
			"jitter = %f, want approximately 16.666667",
			jitter,
		)
	}

	if !almostEqual(loss, 20.0, 0.000001) {
		t.Fatalf("loss = %f, want 20", loss)
	}
}

func TestSummarizeNetworkProbesAllFailed(t *testing.T) {
	samples := []networkProbeSample{
		{Success: false},
		{Success: false},
		{Success: false},
	}

	_,
		_,
		loss,
		err := summarizeNetworkProbes(samples)

	if err == nil {
		t.Fatal("all-failed control probes were accepted")
	}

	if !almostEqual(loss, 100.0, 0.000001) {
		t.Fatalf("loss = %f, want 100", loss)
	}
}

func TestParseNetworkQualityRejectsMissingInterface(t *testing.T) {
	raw := []byte(`{
		"dl_throughput": 100000000,
		"ul_throughput": 50000000
	}`)

	_, _, _, err := parseNetworkQualityResult(raw)

	if err == nil {
		t.Fatal("networkQuality output without interface was accepted")
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
