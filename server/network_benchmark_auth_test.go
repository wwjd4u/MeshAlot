package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestNetworkBenchmarkSignatureAndValidation(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC)
	submission := validNetworkBenchmarkSubmissionFixture(now)

	body, err := json.Marshal(submission)
	if err != nil {
		t.Fatal(err)
	}

	signature := ed25519.Sign(
		privateKey,
		protocol.NetworkBenchmarkSigningMessage(body),
	)

	publicValue := base64.RawStdEncoding.EncodeToString(publicKey)
	signatureValue := base64.RawStdEncoding.EncodeToString(signature)

	if err := verifyNetworkBenchmarkSignature(
		publicValue,
		signatureValue,
		body,
	); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}

	if err := validateNetworkBenchmarkSubmission(
		submission,
		now,
	); err != nil {
		t.Fatalf("valid submission rejected: %v", err)
	}

	changed := append([]byte(nil), body...)
	changed[len(changed)-1] ^= 1

	if !errors.Is(
		verifyNetworkBenchmarkSignature(
			publicValue,
			signatureValue,
			changed,
		),
		ErrNetworkBenchmarkSignature,
	) {
		t.Fatal("modified body was not rejected")
	}

	wrongPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	if !errors.Is(
		verifyNetworkBenchmarkSignature(
			base64.RawStdEncoding.EncodeToString(wrongPublic),
			signatureValue,
			body,
		),
		ErrNetworkBenchmarkSignature,
	) {
		t.Fatal("wrong public key was not rejected")
	}
}

func TestNetworkBenchmarkTimestampWindow(t *testing.T) {
	now := time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC)

	old := validNetworkBenchmarkSubmissionFixture(
		now.Add(-networkBenchmarkSignatureMaxSkew - time.Second),
	)

	if !errors.Is(
		validateNetworkBenchmarkSubmission(old, now),
		ErrNetworkBenchmarkTimestamp,
	) {
		t.Fatal("stale network benchmark timestamp was not rejected")
	}

	future := validNetworkBenchmarkSubmissionFixture(
		now.Add(networkBenchmarkSignatureMaxSkew + time.Second),
	)

	if !errors.Is(
		validateNetworkBenchmarkSubmission(future, now),
		ErrNetworkBenchmarkTimestamp,
	) {
		t.Fatal("future network benchmark timestamp was not rejected")
	}
}

func TestNetworkBenchmarkValidationRejectsBadValues(t *testing.T) {
	now := time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC)

	badVersion := validNetworkBenchmarkSubmissionFixture(now)
	badVersion.Version = "wrong"

	if !errors.Is(
		validateNetworkBenchmarkSubmission(badVersion, now),
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal("invalid submission version was not rejected")
	}

	badNode := validNetworkBenchmarkSubmissionFixture(now)
	badNode.NodeID = "not-a-node"

	if !errors.Is(
		validateNetworkBenchmarkSubmission(badNode, now),
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal("invalid node ID was not rejected")
	}

	badReport := validNetworkBenchmarkSubmissionFixture(now)
	badReport.ReportID = "not-a-report"

	if !errors.Is(
		validateNetworkBenchmarkSubmission(badReport, now),
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal("invalid report ID was not rejected")
	}

	badSchema := validNetworkBenchmarkSubmissionFixture(now)
	badSchema.Benchmark.SchemaVersion = "wrong"

	if !errors.Is(
		validateNetworkBenchmarkSubmission(badSchema, now),
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal("invalid benchmark schema was not rejected")
	}

	badLoss := validNetworkBenchmarkSubmissionFixture(now)
	badLoss.Benchmark.PacketLossPercent = 101

	if !errors.Is(
		validateNetworkBenchmarkSubmission(badLoss, now),
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal("invalid packet loss was not rejected")
	}

	badPath := validNetworkBenchmarkSubmissionFixture(now)
	badPath.Benchmark.PreferredPath = ""

	if !errors.Is(
		validateNetworkBenchmarkSubmission(badPath, now),
		ErrInvalidNetworkBenchmarkSubmission,
	) {
		t.Fatal("empty preferred path was not rejected")
	}
}

func validNetworkBenchmarkSubmissionFixture(
	signedAt time.Time,
) protocol.NetworkBenchmarkSubmission {
	return protocol.NetworkBenchmarkSubmission{
		Version:  protocol.NetworkBenchmarkSubmissionVersion,
		NodeID:   "11111111-1111-4111-8111-111111111111",
		ReportID: "22222222-2222-4222-8222-222222222222",
		SignedAt: signedAt,
		Benchmark: protocol.NetworkBenchmark{
			SchemaVersion:     protocol.NetworkBenchmarkSchemaVersion,
			CollectedAt:       signedAt,
			IPv4Available:     true,
			IPv6Available:     true,
			DownloadMbps:      500.25,
			UploadMbps:        40.75,
			LatencyMs:         22.4,
			JitterMs:          2.1,
			PacketLossPercent: 0,
			PreferredPath:     "ipv6",
		},
	}
}
