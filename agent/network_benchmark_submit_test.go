package agent

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestBuildSignedNetworkBenchmarkSubmission(t *testing.T) {
	identity, err := generateIdentity()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 9, 12, 13, 30, 0, 0, time.UTC)

	benchmark := protocol.NetworkBenchmark{
		SchemaVersion:     protocol.NetworkBenchmarkSchemaVersion,
		CollectedAt:       now,
		IPv4Available:     true,
		IPv6Available:     true,
		DownloadMbps:      500.25,
		UploadMbps:        40.75,
		LatencyMs:         22.4,
		JitterMs:          2.1,
		PacketLossPercent: 0.0,
		PreferredPath:     "ipv6",
	}

	submission, body, signature, err :=
		BuildSignedNetworkBenchmarkSubmission(identity, benchmark, now)

	if err != nil {
		t.Fatal(err)
	}

	if submission.Version != protocol.NetworkBenchmarkSubmissionVersion {
		t.Fatalf("unexpected submission version: %s", submission.Version)
	}

	if submission.NodeID != identity.NodeID {
		t.Fatal("submission used wrong node ID")
	}

	if !validUUIDv4(submission.ReportID) {
		t.Fatalf("report ID is not UUIDv4: %s", submission.ReportID)
	}

	if !submission.SignedAt.Equal(now) {
		t.Fatalf("unexpected signed timestamp: %v", submission.SignedAt)
	}

	var decoded protocol.NetworkBenchmarkSubmission
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Version != protocol.NetworkBenchmarkSubmissionVersion {
		t.Fatal("encoded submission version changed")
	}

	if decoded.ReportID != submission.ReportID {
		t.Fatal("encoded report ID changed")
	}

	if decoded.Benchmark.SchemaVersion != protocol.NetworkBenchmarkSchemaVersion {
		t.Fatal("encoded benchmark schema changed")
	}

	publicKeyBytes, err := base64.RawStdEncoding.DecodeString(identity.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	signatureBytes, err := base64.RawStdEncoding.DecodeString(signature)
	if err != nil {
		t.Fatal(err)
	}

	if !ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.NetworkBenchmarkSigningMessage(body),
		signatureBytes,
	) {
		t.Fatal("valid network benchmark signature did not verify")
	}

	changed := append([]byte(nil), body...)
	changed[len(changed)-1] ^= 1

	if ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.NetworkBenchmarkSigningMessage(changed),
		signatureBytes,
	) {
		t.Fatal("modified network benchmark body verified unexpectedly")
	}

	payload := string(body)

	if strings.Contains(payload, identity.PrivateKey) {
		t.Fatal("private key leaked into network benchmark request")
	}

	if strings.Contains(payload, identity.PublicKey) {
		t.Fatal("public key unnecessarily leaked into network benchmark request")
	}

	if strings.Contains(payload, "private_key") ||
		strings.Contains(payload, "public_key") {
		t.Fatal("identity key fields appeared in network benchmark request")
	}
}
