package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	m9ComputeBenchmarkSubmissionTimeout = 20 * time.Second

	m9ComputeBenchmarkResponseLimit = 64 << 10
)

// PostSignedComputeBenchmark sends an already-created signed M9 submission.
//
// Signing is intentionally separate from transport. The caller therefore has
// the report ID before network activity begins and can use that report ID for
// read-only verification if a POST has an ambiguous outcome.
func PostSignedComputeBenchmark(
	ctx context.Context,
	client computeBenchmarkHTTPDoer,
	serverURL string,
	body []byte,
	signature string,
	reportID string,
) (
	protocol.ComputeBenchmarkSubmissionResponse,
	error,
) {
	var result protocol.ComputeBenchmarkSubmissionResponse

	if ctx == nil {
		return result, errors.New(
			"compute benchmark context is nil",
		)
	}

	baseURL, err :=
		normalizeComputeBenchmarkServerURL(
			serverURL,
		)

	if err != nil {
		return result, err
	}

	if len(body) == 0 {
		return result, errors.New(
			"compute benchmark submission body is empty",
		)
	}

	signature =
		strings.TrimSpace(signature)

	if signature == "" {
		return result, errors.New(
			"compute benchmark signature is required",
		)
	}

	reportID =
		strings.TrimSpace(reportID)

	if reportID == "" {
		return result, errors.New(
			"compute benchmark report ID is required",
		)
	}

	if client == nil {
		client =
			&http.Client{
				Timeout: m9ComputeBenchmarkSubmissionTimeout,
			}
	}

	req, err :=
		http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			baseURL+
				protocol.
					ComputeBenchmarkSubmissionPath,
			bytes.NewReader(body),
		)

	if err != nil {
		return result, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		protocol.
			ComputeBenchmarkSignatureHeader,
		signature,
	)

	resp, err :=
		client.Do(req)

	if err != nil {
		return result, fmt.Errorf(
			"submit compute benchmark report %s: %w",
			reportID,
			err,
		)
	}

	defer resp.Body.Close()

	responseBody, err :=
		io.ReadAll(
			io.LimitReader(
				resp.Body,
				m9ComputeBenchmarkResponseLimit,
			),
		)

	if err != nil {
		return result, fmt.Errorf(
			"read compute benchmark response for report %s: %w",
			reportID,
			err,
		)
	}

	if resp.StatusCode !=
		http.StatusCreated {

		var apiError map[string]string

		_ = json.Unmarshal(
			responseBody,
			&apiError,
		)

		if message :=
			strings.TrimSpace(
				apiError["error"],
			); message != "" {

			return result, fmt.Errorf(
				"compute benchmark report %s rejected: %s",
				reportID,
				message,
			)
		}

		return result, fmt.Errorf(
			"compute benchmark report %s rejected: %s",
			reportID,
			resp.Status,
		)
	}

	if err :=
		json.Unmarshal(
			responseBody,
			&result,
		); err != nil {

		return protocol.
				ComputeBenchmarkSubmissionResponse{},
			fmt.Errorf(
				"compute benchmark service returned an invalid response for report %s",
				reportID,
			)
	}

	if !result.Accepted {
		return protocol.
				ComputeBenchmarkSubmissionResponse{},
			fmt.Errorf(
				"compute benchmark service did not accept report %s",
				reportID,
			)
	}

	if result.ReportID != reportID {
		return protocol.
				ComputeBenchmarkSubmissionResponse{},
			fmt.Errorf(
				"compute benchmark service confirmed unexpected report ID",
			)
	}

	return result, nil
}

// Remote compute benchmark submission requires HTTPS.
//
// Plain HTTP is allowed only for a loopback-only isolated test server.
func normalizeComputeBenchmarkServerURL(
	value string,
) (string, error) {
	value =
		strings.TrimSpace(value)

	if value == "" {
		return "", errors.New(
			"compute benchmark server URL is required",
		)
	}

	parsed, err :=
		url.Parse(value)

	if err != nil {
		return "", errors.New(
			"invalid compute benchmark server URL",
		)
	}

	if parsed.Host == "" ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" ||
		(parsed.Path != "" &&
			parsed.Path != "/") {

		return "", errors.New(
			"invalid compute benchmark server base URL",
		)
	}

	host :=
		strings.ToLower(
			parsed.Hostname(),
		)

	switch parsed.Scheme {
	case "https":
		// Remote submission is permitted only over TLS.

	case "http":
		switch host {
		case "127.0.0.1",
			"localhost",
			"::1":
		default:
			return "", errors.New(
				"plain HTTP compute benchmark transport is limited to loopback",
			)
		}

	default:
		return "", errors.New(
			"compute benchmark server URL must use HTTPS or loopback HTTP",
		)
	}

	return strings.TrimRight(
		value,
		"/",
	), nil
}
