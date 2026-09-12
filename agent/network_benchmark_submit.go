package agent

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func BuildSignedNetworkBenchmarkSubmission(
	identity Identity,
	benchmark protocol.NetworkBenchmark,
	now time.Time,
) (
	protocol.NetworkBenchmarkSubmission,
	[]byte,
	string,
	error,
) {
	if err := validateIdentity(identity); err != nil {
		return protocol.NetworkBenchmarkSubmission{}, nil, "", err
	}

	if benchmark.SchemaVersion != protocol.NetworkBenchmarkSchemaVersion {
		return protocol.NetworkBenchmarkSubmission{},
			nil,
			"",
			errors.New("unsupported network benchmark schema")
	}

	reportID, err := newUUIDv4()
	if err != nil {
		return protocol.NetworkBenchmarkSubmission{}, nil, "", err
	}

	submission := protocol.NetworkBenchmarkSubmission{
		Version:   protocol.NetworkBenchmarkSubmissionVersion,
		NodeID:    identity.NodeID,
		ReportID:  reportID,
		SignedAt:  now.UTC(),
		Benchmark: benchmark,
	}

	body, err := json.Marshal(submission)
	if err != nil {
		return protocol.NetworkBenchmarkSubmission{}, nil, "", err
	}

	privateKeyBytes, err := base64.RawStdEncoding.DecodeString(identity.PrivateKey)
	if err != nil || len(privateKeyBytes) != ed25519.PrivateKeySize {
		return protocol.NetworkBenchmarkSubmission{},
			nil,
			"",
			errors.New("identity private key is invalid")
	}

	signature := ed25519.Sign(
		ed25519.PrivateKey(privateKeyBytes),
		protocol.NetworkBenchmarkSigningMessage(body),
	)

	return submission,
		body,
		base64.RawStdEncoding.EncodeToString(signature),
		nil
}
