package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const networkBenchmarkSignatureMaxSkew = 5 * time.Minute

var (
	ErrInvalidNetworkBenchmarkSubmission = errors.New("invalid network benchmark submission")
	ErrNetworkBenchmarkSignature         = errors.New("invalid network benchmark signature")
	ErrNetworkBenchmarkTimestamp         = errors.New("network benchmark timestamp outside allowed window")
)

func verifyNetworkBenchmarkSignature(
	publicKeyValue string,
	signatureValue string,
	body []byte,
) error {
	publicKeyBytes, err := base64.RawStdEncoding.DecodeString(
		strings.TrimSpace(publicKeyValue),
	)
	if err != nil || len(publicKeyBytes) != ed25519.PublicKeySize {
		return ErrNetworkBenchmarkSignature
	}

	signatureBytes, err := base64.RawStdEncoding.DecodeString(
		strings.TrimSpace(signatureValue),
	)
	if err != nil || len(signatureBytes) != ed25519.SignatureSize {
		return ErrNetworkBenchmarkSignature
	}

	if !ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.NetworkBenchmarkSigningMessage(body),
		signatureBytes,
	) {
		return ErrNetworkBenchmarkSignature
	}

	return nil
}

func validateNetworkBenchmarkSubmission(
	submission protocol.NetworkBenchmarkSubmission,
	now time.Time,
) error {
	if submission.Version != protocol.NetworkBenchmarkSubmissionVersion ||
		!validUUIDv4(submission.NodeID) ||
		!validUUIDv4(submission.ReportID) {
		return ErrInvalidNetworkBenchmarkSubmission
	}

	if submission.SignedAt.IsZero() {
		return ErrNetworkBenchmarkTimestamp
	}

	skew := now.UTC().Sub(submission.SignedAt.UTC())
	if skew < 0 {
		skew = -skew
	}

	if skew > networkBenchmarkSignatureMaxSkew {
		return ErrNetworkBenchmarkTimestamp
	}

	benchmark := submission.Benchmark

	collectionSkew := submission.SignedAt.UTC().Sub(
		benchmark.CollectedAt.UTC(),
	)
	if collectionSkew < 0 {
		collectionSkew = -collectionSkew
	}

	if collectionSkew > networkBenchmarkSignatureMaxSkew {
		return ErrNetworkBenchmarkTimestamp
	}

	if benchmark.SchemaVersion != protocol.NetworkBenchmarkSchemaVersion ||
		benchmark.CollectedAt.IsZero() ||
		strings.TrimSpace(benchmark.PreferredPath) == "" {
		return ErrInvalidNetworkBenchmarkSubmission
	}

	if benchmark.DownloadMbps < 0 ||
		benchmark.UploadMbps < 0 ||
		benchmark.LatencyMs < 0 ||
		benchmark.JitterMs < 0 ||
		benchmark.PacketLossPercent < 0 ||
		benchmark.PacketLossPercent > 100 {
		return ErrInvalidNetworkBenchmarkSubmission
	}

	return nil
}
