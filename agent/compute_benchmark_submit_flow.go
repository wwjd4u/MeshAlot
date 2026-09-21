package agent

import (
	"context"
	"errors"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type ComputeBenchmarkSubmitResult struct {
	Submission protocol.ComputeBenchmarkSubmission
	Response   protocol.ComputeBenchmarkSubmissionResponse
}

// SubmitComputeBenchmark signs an already-collected M9 benchmark and submits
// that exact signed body through the M9 HTTP transport.
//
// The generated submission is returned even when the HTTP operation fails.
// This intentionally preserves ReportID so an ambiguous POST can be checked
// read-only before any retry is considered.
func SubmitComputeBenchmark(
	ctx context.Context,
	identity Identity,
	benchmark protocol.ComputeBenchmark,
	serverURL string,
	signedAt time.Time,
	client computeBenchmarkHTTPDoer,
) (
	ComputeBenchmarkSubmitResult,
	error,
) {
	var result ComputeBenchmarkSubmitResult

	if ctx == nil {
		return result, errors.New(
			"compute benchmark context is nil",
		)
	}

	normalizedURL, err :=
		normalizeComputeBenchmarkServerURL(
			serverURL,
		)

	if err != nil {
		return result, err
	}

	submission,
		body,
		signature,
		err :=
		BuildSignedComputeBenchmarkSubmission(
			identity,
			benchmark,
			signedAt,
		)

	if err != nil {
		return result, err
	}

	result.Submission =
		submission

	response, err :=
		PostSignedComputeBenchmark(
			ctx,
			client,
			normalizedURL,
			body,
			signature,
			submission.ReportID,
		)

	if err != nil {
		return result, err
	}

	result.Response =
		response

	return result, nil
}
