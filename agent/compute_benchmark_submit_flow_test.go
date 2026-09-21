package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type computeBenchmarkSubmitFlowDoer struct {
	calls int

	fail error

	submission protocol.ComputeBenchmarkSubmission
	signature  string
}

func (f *computeBenchmarkSubmitFlowDoer) Do(
	req *http.Request,
) (*http.Response, error) {
	f.calls++

	body, err :=
		io.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	if err :=
		json.Unmarshal(
			body,
			&f.submission,
		); err != nil {

		return nil, err
	}

	f.signature =
		req.Header.Get(
			protocol.
				ComputeBenchmarkSignatureHeader,
		)

	if f.fail != nil {
		return nil, f.fail
	}

	response, err :=
		json.Marshal(
			protocol.
				ComputeBenchmarkSubmissionResponse{
				ReportID: f.submission.ReportID,

				Accepted: true,
			},
		)

	if err != nil {
		return nil, err
	}

	return &http.Response{
		StatusCode: http.StatusCreated,

		Status: "201 Created",

		Body: io.NopCloser(
			strings.NewReader(
				string(response),
			),
		),

		Header: make(http.Header),
	}, nil
}

func m9SubmitFlowBenchmark(
	now time.Time,
) protocol.ComputeBenchmark {
	return protocol.ComputeBenchmark{
		SchemaVersion: protocol.
			ComputeBenchmarkSchemaVersion,

		CollectedAt: now,

		WorkloadID: "m9-standard-v1",

		Condition: "baseline",

		Runtime: "llama.cpp",

		RuntimeVersion: "test-version",

		Model: "test-model.gguf",

		Quantization: "Q4_K_M",

		ContextSize: 4096,

		Samples: []protocol.ComputeBenchmarkSample{
			{
				Run: 1,

				PromptTokens: 512,

				GeneratedTokens: 128,

				PromptTokensPerSecond: 120,

				GenerationTokensPerSecond: 30,

				TimeToFirstTokenMs: 200,

				PeakSystemRAMBytes: 8 *
					1024 *
					1024 *
					1024,

				Success: true,
			},
		},
	}
}

func TestSubmitComputeBenchmark(
	t *testing.T,
) {
	identity, err :=
		generateIdentity()

	if err != nil {
		t.Fatal(err)
	}

	now :=
		time.Date(
			2026,
			9,
			18,
			9,
			0,
			0,
			0,
			time.UTC,
		)

	client :=
		&computeBenchmarkSubmitFlowDoer{}

	result, err :=
		SubmitComputeBenchmark(
			context.Background(),
			identity,
			m9SubmitFlowBenchmark(now),
			"https://api.example.test",
			now,
			client,
		)

	if err != nil {
		t.Fatal(err)
	}

	if client.calls != 1 {
		t.Fatalf(
			"HTTP calls = %d",
			client.calls,
		)
	}

	if !validUUIDv4(
		result.
			Submission.
			ReportID,
	) {
		t.Fatalf(
			"invalid report ID: %q",
			result.
				Submission.
				ReportID,
		)
	}

	if result.Submission.NodeID !=
		identity.NodeID {

		t.Fatal(
			"submission used wrong node identity",
		)
	}

	if client.submission.ReportID !=
		result.Submission.ReportID {

		t.Fatal(
			"transport body used different report ID",
		)
	}

	if strings.TrimSpace(
		client.signature,
	) == "" {

		t.Fatal(
			"signature header was missing",
		)
	}

	if !result.Response.Accepted ||
		result.Response.ReportID !=
			result.Submission.ReportID {

		t.Fatalf(
			"unexpected confirmation: %#v",
			result.Response,
		)
	}
}

func TestSubmitComputeBenchmarkPreservesReportIDOnTransportFailure(
	t *testing.T,
) {
	identity, err :=
		generateIdentity()

	if err != nil {
		t.Fatal(err)
	}

	now :=
		time.Now().UTC()

	client :=
		&computeBenchmarkSubmitFlowDoer{
			fail: errors.New(
				"connection reset",
			),
		}

	result, err :=
		SubmitComputeBenchmark(
			context.Background(),
			identity,
			m9SubmitFlowBenchmark(now),
			"https://api.example.test",
			now,
			client,
		)

	if err == nil {
		t.Fatal(
			"transport failure was accepted",
		)
	}

	reportID :=
		result.
			Submission.
			ReportID

	if !validUUIDv4(reportID) {
		t.Fatalf(
			"report ID was lost after transport failure: %q",
			reportID,
		)
	}

	if !strings.Contains(
		err.Error(),
		reportID,
	) {
		t.Fatalf(
			"error lost report ID: %v",
			err,
		)
	}

	if client.calls != 1 {
		t.Fatalf(
			"HTTP calls = %d",
			client.calls,
		)
	}
}

func TestSubmitComputeBenchmarkRejectsInvalidServerBeforeSigning(
	t *testing.T,
) {
	identity, err :=
		generateIdentity()

	if err != nil {
		t.Fatal(err)
	}

	now :=
		time.Now().UTC()

	client :=
		&computeBenchmarkSubmitFlowDoer{}

	result, err :=
		SubmitComputeBenchmark(
			context.Background(),
			identity,
			m9SubmitFlowBenchmark(now),
			"http://192.0.2.10:8080",
			now,
			client,
		)

	if err == nil {
		t.Fatal(
			"invalid remote HTTP server was accepted",
		)
	}

	if result.Submission.ReportID != "" {
		t.Fatal(
			"report ID was generated before server validation",
		)
	}

	if client.calls != 0 {
		t.Fatal(
			"invalid server was contacted",
		)
	}
}

func TestSubmitComputeBenchmarkRejectsWrongSchemaBeforeTransport(
	t *testing.T,
) {
	identity, err :=
		generateIdentity()

	if err != nil {
		t.Fatal(err)
	}

	now :=
		time.Now().UTC()

	benchmark :=
		m9SubmitFlowBenchmark(now)

	benchmark.SchemaVersion =
		"wrong-version"

	client :=
		&computeBenchmarkSubmitFlowDoer{}

	result, err :=
		SubmitComputeBenchmark(
			context.Background(),
			identity,
			benchmark,
			"https://api.example.test",
			now,
			client,
		)

	if err == nil {
		t.Fatal(
			"wrong benchmark schema was accepted",
		)
	}

	if result.Submission.ReportID != "" {
		t.Fatal(
			"invalid benchmark produced a retained report ID",
		)
	}

	if client.calls != 0 {
		t.Fatal(
			"invalid benchmark reached transport",
		)
	}
}
