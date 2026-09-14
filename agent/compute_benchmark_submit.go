package agent

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func BuildSignedComputeBenchmarkSubmission(
	identity Identity,
	benchmark protocol.ComputeBenchmark,
	now time.Time,
) (
	protocol.ComputeBenchmarkSubmission,
	[]byte,
	string,
	error,
) {
	if err := validateIdentity(identity); err != nil {
		return protocol.ComputeBenchmarkSubmission{}, nil, "", err
	}

	if benchmark.SchemaVersion != protocol.ComputeBenchmarkSchemaVersion {
		return protocol.ComputeBenchmarkSubmission{},
			nil,
			"",
			errors.New("unsupported compute benchmark schema")
	}

	reportID, err := newUUIDv4()
	if err != nil {
		return protocol.ComputeBenchmarkSubmission{}, nil, "", err
	}

	submission := protocol.ComputeBenchmarkSubmission{
		Version:   protocol.ComputeBenchmarkSubmissionVersion,
		NodeID:    identity.NodeID,
		ReportID:  reportID,
		SignedAt:  now.UTC(),
		Benchmark: benchmark,
	}

	body, err := json.Marshal(submission)
	if err != nil {
		return protocol.ComputeBenchmarkSubmission{}, nil, "", err
	}

	privateKeyBytes, err := base64.RawStdEncoding.DecodeString(
		identity.PrivateKey,
	)
	if err != nil || len(privateKeyBytes) != ed25519.PrivateKeySize {
		return protocol.ComputeBenchmarkSubmission{},
			nil,
			"",
			errors.New("identity private key is invalid")
	}

	signature := ed25519.Sign(
		ed25519.PrivateKey(privateKeyBytes),
		protocol.ComputeBenchmarkSigningMessage(body),
	)

	return submission,
		body,
		base64.RawStdEncoding.EncodeToString(signature),
		nil
}
