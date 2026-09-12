package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestDecodeNetworkBenchmarkSubmissionBodyStrict(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.September,
		12,
		15,
		0,
		0,
		0,
		time.UTC,
	)

	valid := protocol.NetworkBenchmarkSubmission{
		Version:  protocol.NetworkBenchmarkSubmissionVersion,
		NodeID:   "11111111-1111-4111-8111-111111111111",
		ReportID: "22222222-2222-4222-8222-222222222222",
		SignedAt: now,
		Benchmark: protocol.NetworkBenchmark{
			SchemaVersion:     protocol.NetworkBenchmarkSchemaVersion,
			CollectedAt:       now,
			IPv4Available:     true,
			IPv6Available:     true,
			DownloadMbps:      100,
			UploadMbps:        50,
			LatencyMs:         40,
			JitterMs:          5,
			PacketLossPercent: 0,
			PreferredPath:     "en0",
		},
	}

	body, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := decodeNetworkBenchmarkSubmissionBody(body)
	if err != nil {
		t.Fatalf(
			"valid network benchmark body rejected: %v",
			err,
		)
	}

	if decoded.ReportID != valid.ReportID {
		t.Fatal("report ID changed during decode")
	}

	unknown := append(
		append([]byte(nil), body[:len(body)-1]...),
		[]byte(`,"unexpected":true}`)...,
	)

	if _, err := decodeNetworkBenchmarkSubmissionBody(
		unknown,
	); !errors.Is(
		err,
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal(
			"unknown top-level field was not rejected",
		)
	}

	trailing := append(
		append([]byte(nil), body...),
		[]byte(`{"second":true}`)...,
	)

	if _, err := decodeNetworkBenchmarkSubmissionBody(
		trailing,
	); !errors.Is(
		err,
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal(
			"trailing JSON value was not rejected",
		)
	}
}

func TestNetworkBenchmarkBodyLimitIsOneMiB(t *testing.T) {
	if maxNetworkBenchmarkSubmissionBodyBytes != 1024*1024 {
		t.Fatalf(
			"unexpected network benchmark body limit: %d",
			maxNetworkBenchmarkSubmissionBodyBytes,
		)
	}
}

func TestNetworkBenchmarkRouteIsRegistered(t *testing.T) {
	service := New(
		slog.Default(),
		"",
	)

	request := httptest.NewRequest(
		http.MethodPost,
		protocol.NetworkBenchmarkSubmissionPath,
		strings.NewReader("{}"),
	)

	recorder := httptest.NewRecorder()

	service.Handler().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"network benchmark route status = %d, want %d",
			recorder.Code,
			http.StatusServiceUnavailable,
		)
	}
}
