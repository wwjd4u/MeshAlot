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

func TestDecodeComputeBenchmarkSubmissionBodyStrict(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.September,
		14,
		21,
		0,
		0,
		0,
		time.UTC,
	)

	valid := protocol.ComputeBenchmarkSubmission{
		Version: protocol.ComputeBenchmarkSubmissionVersion,

		NodeID: "11111111-1111-4111-8111-111111111111",

		ReportID: "22222222-2222-4222-8222-222222222222",

		SignedAt: now,

		Benchmark: protocol.ComputeBenchmark{
			SchemaVersion: protocol.ComputeBenchmarkSchemaVersion,

			CollectedAt: now,
			WorkloadID:  "m9-standard-v1",
			Condition:   "baseline",
			Runtime:     "llama.cpp",
			Model:       "test-model",
			ContextSize: 4096,

			Samples: []protocol.ComputeBenchmarkSample{
				{
					Run:             1,
					PromptTokens:    512,
					GeneratedTokens: 128,

					PromptTokensPerSecond: 250,

					GenerationTokensPerSecond: 35,

					TimeToFirstTokenMs: 300,

					PeakSystemRAMBytes: 8 * 1024 * 1024 * 1024,

					Success: true,
				},
			},
		},
	}

	body, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err :=
		decodeComputeBenchmarkSubmissionBody(body)

	if err != nil {
		t.Fatalf(
			"valid compute benchmark body rejected: %v",
			err,
		)
	}

	if decoded.ReportID != valid.ReportID {
		t.Fatal("report ID changed during decode")
	}

	if decoded.Benchmark.WorkloadID !=
		valid.Benchmark.WorkloadID {
		t.Fatal("workload ID changed during decode")
	}

	unknown := append(
		append([]byte(nil), body[:len(body)-1]...),
		[]byte(`,"unexpected":true}`)...,
	)

	if _, err := decodeComputeBenchmarkSubmissionBody(
		unknown,
	); !errors.Is(
		err,
		ErrInvalidComputeBenchmarkSubmission,
	) {
		t.Fatal(
			"unknown top-level field was not rejected",
		)
	}

	trailing := append(
		append([]byte(nil), body...),
		[]byte(`{"second":true}`)...,
	)

	if _, err := decodeComputeBenchmarkSubmissionBody(
		trailing,
	); !errors.Is(
		err,
		ErrInvalidComputeBenchmarkSubmission,
	) {
		t.Fatal(
			"trailing JSON value was not rejected",
		)
	}
}

func TestComputeBenchmarkBodyLimitIsOneMiB(
	t *testing.T,
) {
	if maxComputeBenchmarkSubmissionBodyBytes !=
		1024*1024 {
		t.Fatalf(
			"unexpected compute benchmark body limit: %d",
			maxComputeBenchmarkSubmissionBodyBytes,
		)
	}
}

func TestComputeBenchmarkRouteIsRegistered(
	t *testing.T,
) {
	service := New(
		slog.Default(),
		"",
	)

	request := httptest.NewRequest(
		http.MethodPost,
		protocol.ComputeBenchmarkSubmissionPath,
		strings.NewReader("{}"),
	)

	recorder := httptest.NewRecorder()

	service.Handler().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code !=
		http.StatusServiceUnavailable {
		t.Fatalf(
			"compute benchmark route status = %d, want %d",
			recorder.Code,
			http.StatusServiceUnavailable,
		)
	}
}
