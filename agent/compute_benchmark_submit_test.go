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

func TestBuildSignedComputeBenchmarkSubmission(t *testing.T) {
	identity, err := generateIdentity()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(
		2026,
		9,
		14,
		19,
		0,
		0,
		0,
		time.UTC,
	)

	vram := uint64(8 * 1024 * 1024 * 1024)

	benchmark := protocol.ComputeBenchmark{
		SchemaVersion:  protocol.ComputeBenchmarkSchemaVersion,
		CollectedAt:    now,
		WorkloadID:     "m9-standard-v1",
		Condition:      "baseline",
		Runtime:        "llama.cpp",
		RuntimeVersion: "test-version",
		Model:          "test-model",
		Quantization:   "Q4_K_M",
		ContextSize:    4096,
		Samples: []protocol.ComputeBenchmarkSample{
			{
				Run:                       1,
				PromptTokens:              512,
				GeneratedTokens:           128,
				PromptTokensPerSecond:     120.25,
				GenerationTokensPerSecond: 31.75,
				TimeToFirstTokenMs:        185.5,
				PeakSystemRAMBytes:        12 * 1024 * 1024 * 1024,
				PeakVRAMBytes:             &vram,
				Success:                   true,
			},
		},
	}

	submission, body, signature, err :=
		BuildSignedComputeBenchmarkSubmission(
			identity,
			benchmark,
			now,
		)
	if err != nil {
		t.Fatal(err)
	}

	if submission.Version !=
		protocol.ComputeBenchmarkSubmissionVersion {
		t.Fatalf(
			"unexpected submission version: %s",
			submission.Version,
		)
	}

	if submission.NodeID != identity.NodeID {
		t.Fatal("submission used wrong node ID")
	}

	if !validUUIDv4(submission.ReportID) {
		t.Fatalf(
			"report ID is not UUIDv4: %s",
			submission.ReportID,
		)
	}

	if !submission.SignedAt.Equal(now) {
		t.Fatalf(
			"unexpected signed timestamp: %v",
			submission.SignedAt,
		)
	}

	var decoded protocol.ComputeBenchmarkSubmission

	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Version !=
		protocol.ComputeBenchmarkSubmissionVersion {
		t.Fatal("encoded submission version changed")
	}

	if decoded.ReportID != submission.ReportID {
		t.Fatal("encoded report ID changed")
	}

	if decoded.Benchmark.SchemaVersion !=
		protocol.ComputeBenchmarkSchemaVersion {
		t.Fatal("encoded benchmark schema changed")
	}

	if decoded.Benchmark.WorkloadID != "m9-standard-v1" {
		t.Fatalf(
			"encoded workload ID changed: %q",
			decoded.Benchmark.WorkloadID,
		)
	}

	if len(decoded.Benchmark.Samples) != 1 {
		t.Fatalf(
			"encoded sample count = %d, want 1",
			len(decoded.Benchmark.Samples),
		)
	}

	publicKeyBytes, err := base64.RawStdEncoding.DecodeString(
		identity.PublicKey,
	)
	if err != nil {
		t.Fatal(err)
	}

	signatureBytes, err := base64.RawStdEncoding.DecodeString(
		signature,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.ComputeBenchmarkSigningMessage(body),
		signatureBytes,
	) {
		t.Fatal("valid compute benchmark signature did not verify")
	}

	changed := append([]byte(nil), body...)
	changed[len(changed)-1] ^= 1

	if ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.ComputeBenchmarkSigningMessage(changed),
		signatureBytes,
	) {
		t.Fatal(
			"modified compute benchmark body verified unexpectedly",
		)
	}

	payload := string(body)

	if strings.Contains(payload, identity.PrivateKey) {
		t.Fatal(
			"private key leaked into compute benchmark request",
		)
	}

	if strings.Contains(payload, identity.PublicKey) {
		t.Fatal(
			"public key unnecessarily leaked into compute benchmark request",
		)
	}

	if strings.Contains(payload, "private_key") ||
		strings.Contains(payload, "public_key") {
		t.Fatal(
			"identity key fields appeared in compute benchmark request",
		)
	}
}

func TestBuildSignedComputeBenchmarkSubmissionRejectsWrongSchema(
	t *testing.T,
) {
	identity, err := generateIdentity()
	if err != nil {
		t.Fatal(err)
	}

	benchmark := protocol.ComputeBenchmark{
		SchemaVersion: "wrong-version",
	}

	_, _, _, err = BuildSignedComputeBenchmarkSubmission(
		identity,
		benchmark,
		time.Now().UTC(),
	)

	if err == nil {
		t.Fatal("unsupported compute benchmark schema was accepted")
	}
}
