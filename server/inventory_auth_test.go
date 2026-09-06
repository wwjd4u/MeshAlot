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

func TestInventorySignatureAndValidation(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)

	submission := validInventorySubmissionFixture(now)

	body, err := json.Marshal(submission)
	if err != nil {
		t.Fatal(err)
	}

	signature := ed25519.Sign(
		privateKey,
		protocol.InventorySigningMessage(body),
	)

	publicValue := base64.RawStdEncoding.EncodeToString(publicKey)
	signatureValue := base64.RawStdEncoding.EncodeToString(signature)

	if err := verifyInventorySignature(
		publicValue,
		signatureValue,
		body,
	); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}

	if err := validateInventorySubmission(
		submission,
		now,
	); err != nil {
		t.Fatalf("valid submission rejected: %v", err)
	}

	changed := append([]byte(nil), body...)
	changed[len(changed)-1] ^= 1

	if !errors.Is(
		verifyInventorySignature(publicValue, signatureValue, changed),
		ErrInventorySignature,
	) {
		t.Fatal("modified body was not rejected")
	}

	wrongPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	if !errors.Is(
		verifyInventorySignature(
			base64.RawStdEncoding.EncodeToString(wrongPublic),
			signatureValue,
			body,
		),
		ErrInventorySignature,
	) {
		t.Fatal("wrong public key was not rejected")
	}
}

func TestInventoryTimestampWindow(t *testing.T) {
	now := time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)

	old := validInventorySubmissionFixture(
		now.Add(-inventorySignatureMaxSkew - time.Second),
	)

	if !errors.Is(
		validateInventorySubmission(old, now),
		ErrInventoryTimestamp,
	) {
		t.Fatal("stale inventory timestamp was not rejected")
	}

	future := validInventorySubmissionFixture(
		now.Add(inventorySignatureMaxSkew + time.Second),
	)

	if !errors.Is(
		validateInventorySubmission(future, now),
		ErrInventoryTimestamp,
	) {
		t.Fatal("future inventory timestamp was not rejected")
	}
}

func TestInventoryCollectionFreshness(t *testing.T) {
	now := time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)

	stale := validInventorySubmissionFixture(now)
	stale.Inventory.CollectedAt =
		now.Add(-inventorySignatureMaxSkew - time.Second)

	if !errors.Is(
		validateInventorySubmission(stale, now),
		ErrInventoryTimestamp,
	) {
		t.Fatal("stale collected_at was not rejected")
	}

	future := validInventorySubmissionFixture(now)
	future.Inventory.CollectedAt =
		now.Add(inventorySignatureMaxSkew + time.Second)

	if !errors.Is(
		validateInventorySubmission(future, now),
		ErrInventoryTimestamp,
	) {
		t.Fatal("future collected_at was not rejected")
	}
}

func TestInventoryValidationRejectsBadIDsAndSchema(t *testing.T) {
	now := time.Date(2026, 9, 5, 4, 0, 0, 0, time.UTC)

	badVersion := validInventorySubmissionFixture(now)
	badVersion.Version = "wrong"

	if !errors.Is(
		validateInventorySubmission(badVersion, now),
		ErrInvalidInventorySubmission,
	) {
		t.Fatal("invalid submission version was not rejected")
	}

	badNode := validInventorySubmissionFixture(now)
	badNode.NodeID = "not-a-node"

	if !errors.Is(
		validateInventorySubmission(badNode, now),
		ErrInvalidInventorySubmission,
	) {
		t.Fatal("invalid node ID was not rejected")
	}

	badReport := validInventorySubmissionFixture(now)
	badReport.ReportID = "not-a-report"

	if !errors.Is(
		validateInventorySubmission(badReport, now),
		ErrInvalidInventorySubmission,
	) {
		t.Fatal("invalid report ID was not rejected")
	}

	badSchema := validInventorySubmissionFixture(now)
	badSchema.Inventory.SchemaVersion = "wrong"

	if !errors.Is(
		validateInventorySubmission(badSchema, now),
		ErrInvalidInventorySubmission,
	) {
		t.Fatal("invalid inventory schema was not rejected")
	}
}

func validInventorySubmissionFixture(
	signedAt time.Time,
) protocol.InventorySubmission {

	return protocol.InventorySubmission{
		Version:  protocol.InventorySubmissionVersion,
		NodeID:   "11111111-1111-4111-8111-111111111111",
		ReportID: "22222222-2222-4222-8222-222222222222",
		SignedAt: signedAt,
		Inventory: protocol.HardwareInventory{
			SchemaVersion: protocol.HardwareInventorySchemaVersion,
			CollectedAt:   signedAt,
			OS: protocol.OperatingSystemInventory{
				Name:         "Ubuntu",
				Version:      "26.04.1",
				Architecture: "x86_64",
			},
			CPU: protocol.CPUInventory{
				Model:         "Intel CPU",
				PhysicalCores: 14,
				LogicalCores:  14,
			},
			Memory: protocol.MemoryInventory{
				TotalBytes: 132049289216,
			},
			GPUs:         []protocol.GPUInventory{},
			Acceleration: []protocol.AccelerationRuntime{},
			AIRuntimes:   []protocol.AIRuntimeInventory{},
			Models:       []protocol.ModelInventory{},
		},
	}
}
