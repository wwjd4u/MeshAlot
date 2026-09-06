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

func TestBuildSignedInventorySubmission(t *testing.T) {
	identity, err := generateIdentity()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)

	inventory := protocol.HardwareInventory{
		SchemaVersion: protocol.HardwareInventorySchemaVersion,
		CollectedAt:   now,
		OS: protocol.OperatingSystemInventory{
			Name:         "Ubuntu",
			Version:      "26.04.1",
			Architecture: "x86_64",
		},
		CPU: protocol.CPUInventory{
			Model:         "Intel(R) Core(TM) Ultra 5 235HX",
			PhysicalCores: 14,
			LogicalCores:  14,
		},
		Memory: protocol.MemoryInventory{
			TotalBytes: 132049289216,
		},
	}

	submission, body, signature, err :=
		BuildSignedInventorySubmission(identity, inventory, now)

	if err != nil {
		t.Fatal(err)
	}

	if submission.Version != protocol.InventorySubmissionVersion {
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

	var decoded protocol.InventorySubmission
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Version != protocol.InventorySubmissionVersion {
		t.Fatal("encoded submission version changed")
	}

	if decoded.ReportID != submission.ReportID {
		t.Fatal("encoded report ID changed")
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
		protocol.InventorySigningMessage(body),
		signatureBytes,
	) {
		t.Fatal("valid inventory signature did not verify")
	}

	changed := append([]byte(nil), body...)
	changed[len(changed)-1] ^= 1

	if ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.InventorySigningMessage(changed),
		signatureBytes,
	) {
		t.Fatal("modified inventory body verified unexpectedly")
	}

	payload := string(body)

	if strings.Contains(payload, identity.PrivateKey) {
		t.Fatal("private key leaked into inventory request")
	}

	if strings.Contains(payload, identity.PublicKey) {
		t.Fatal("public key unnecessarily leaked into inventory request")
	}

	if strings.Contains(payload, "private_key") ||
		strings.Contains(payload, "public_key") {
		t.Fatal("identity key fields appeared in inventory request")
	}
}
