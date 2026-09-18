package agent

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type fakeComputeBenchmarkTransportDoer struct {
	calls int

	statusCode int
	response   string
	err        error

	method    string
	url       string
	signature string
	body      []byte
}

func (f *fakeComputeBenchmarkTransportDoer) Do(
	req *http.Request,
) (*http.Response, error) {
	f.calls++

	f.method =
		req.Method

	f.url =
		req.URL.String()

	f.signature =
		req.Header.Get(
			protocol.
				ComputeBenchmarkSignatureHeader,
		)

	body, err :=
		io.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	f.body =
		append(
			[]byte(nil),
			body...,
		)

	if f.err != nil {
		return nil, f.err
	}

	return &http.Response{
		StatusCode: f.statusCode,

		Status: http.StatusText(
			f.statusCode,
		),

		Body: io.NopCloser(
			strings.NewReader(
				f.response,
			),
		),

		Header: make(http.Header),
	}, nil
}

func TestPostSignedComputeBenchmark(
	t *testing.T,
) {
	const reportID = "11111111-2222-4333-8444-555555555555"

	body :=
		[]byte(
			`{"version":"m9-v1","report_id":"` +
				reportID +
				`"}`,
		)

	client :=
		&fakeComputeBenchmarkTransportDoer{
			statusCode: http.StatusCreated,

			response: `{"report_id":"` +
				reportID +
				`","accepted":true}`,
		}

	result, err :=
		PostSignedComputeBenchmark(
			context.Background(),
			client,
			"https://api.example.test/",
			body,
			"test-signature",
			reportID,
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

	if client.method !=
		http.MethodPost {

		t.Fatalf(
			"method = %q",
			client.method,
		)
	}

	wantURL :=
		"https://api.example.test" +
			protocol.
				ComputeBenchmarkSubmissionPath

	if client.url != wantURL {
		t.Fatalf(
			"URL = %q, want %q",
			client.url,
			wantURL,
		)
	}

	if client.signature !=
		"test-signature" {

		t.Fatal(
			"signature header changed",
		)
	}

	if string(client.body) !=
		string(body) {

		t.Fatal(
			"signed body changed during transport",
		)
	}

	if !result.Accepted ||
		result.ReportID != reportID {

		t.Fatalf(
			"unexpected response: %#v",
			result,
		)
	}
}

func TestPostSignedComputeBenchmarkRejectsPlainRemoteHTTP(
	t *testing.T,
) {
	client :=
		&fakeComputeBenchmarkTransportDoer{}

	_, err :=
		PostSignedComputeBenchmark(
			context.Background(),
			client,
			"http://192.0.2.10:8080",
			[]byte(`{"version":"m9-v1"}`),
			"signature",
			"11111111-2222-4333-8444-555555555555",
		)

	if err == nil {
		t.Fatal(
			"plain remote HTTP was accepted",
		)
	}

	if client.calls != 0 {
		t.Fatal(
			"remote HTTP endpoint was contacted",
		)
	}
}

func TestPostSignedComputeBenchmarkAllowsLoopbackHTTP(
	t *testing.T,
) {
	const reportID = "11111111-2222-4333-8444-555555555555"

	client :=
		&fakeComputeBenchmarkTransportDoer{
			statusCode: http.StatusCreated,

			response: `{"report_id":"` +
				reportID +
				`","accepted":true}`,
		}

	_, err :=
		PostSignedComputeBenchmark(
			context.Background(),
			client,
			"http://127.0.0.1:18180",
			[]byte(`{"version":"m9-v1"}`),
			"signature",
			reportID,
		)

	if err != nil {
		t.Fatal(err)
	}

	if client.calls != 1 {
		t.Fatal(
			"loopback HTTP was not attempted",
		)
	}
}

func TestPostSignedComputeBenchmarkPreservesReportIDOnTransportFailure(
	t *testing.T,
) {
	const reportID = "11111111-2222-4333-8444-555555555555"

	client :=
		&fakeComputeBenchmarkTransportDoer{
			err: errors.New(
				"connection reset",
			),
		}

	_, err :=
		PostSignedComputeBenchmark(
			context.Background(),
			client,
			"https://api.example.test",
			[]byte(`{"version":"m9-v1"}`),
			"signature",
			reportID,
		)

	if err == nil {
		t.Fatal(
			"transport failure was accepted",
		)
	}

	if !strings.Contains(
		err.Error(),
		reportID,
	) {
		t.Fatalf(
			"transport error lost report ID: %v",
			err,
		)
	}
}

func TestPostSignedComputeBenchmarkRejectsMismatchedConfirmation(
	t *testing.T,
) {
	const reportID = "11111111-2222-4333-8444-555555555555"

	client :=
		&fakeComputeBenchmarkTransportDoer{
			statusCode: http.StatusCreated,

			response: `{"report_id":"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee","accepted":true}`,
		}

	_, err :=
		PostSignedComputeBenchmark(
			context.Background(),
			client,
			"https://api.example.test",
			[]byte(`{"version":"m9-v1"}`),
			"signature",
			reportID,
		)

	if err == nil {
		t.Fatal(
			"mismatched report confirmation was accepted",
		)
	}
}
