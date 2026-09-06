package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const inventorySignatureMaxSkew = 5 * time.Minute

var (
	ErrInvalidInventorySubmission = errors.New("invalid inventory submission")
	ErrInventorySignature         = errors.New("invalid inventory signature")
	ErrInventoryTimestamp         = errors.New("inventory signature timestamp outside allowed window")
)

func verifyInventorySignature(
	publicKeyValue string,
	signatureValue string,
	body []byte,
) error {

	publicKeyBytes, err := base64.RawStdEncoding.DecodeString(
		strings.TrimSpace(publicKeyValue),
	)
	if err != nil || len(publicKeyBytes) != ed25519.PublicKeySize {
		return ErrInventorySignature
	}

	signatureBytes, err := base64.RawStdEncoding.DecodeString(
		strings.TrimSpace(signatureValue),
	)
	if err != nil || len(signatureBytes) != ed25519.SignatureSize {
		return ErrInventorySignature
	}

	if !ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.InventorySigningMessage(body),
		signatureBytes,
	) {
		return ErrInventorySignature
	}

	return nil
}

func validateInventorySubmission(
	submission protocol.InventorySubmission,
	now time.Time,
) error {

	if submission.Version != protocol.InventorySubmissionVersion ||
		!validUUIDv4(submission.NodeID) ||
		!validUUIDv4(submission.ReportID) {
		return ErrInvalidInventorySubmission
	}

	if submission.SignedAt.IsZero() {
		return ErrInventoryTimestamp
	}

	skew := now.UTC().Sub(submission.SignedAt.UTC())
	if skew < 0 {
		skew = -skew
	}

	if skew > inventorySignatureMaxSkew {
		return ErrInventoryTimestamp
	}

	inventory := submission.Inventory

	collectionSkew := submission.SignedAt.UTC().Sub(
		inventory.CollectedAt.UTC(),
	)
	if collectionSkew < 0 {
		collectionSkew = -collectionSkew
	}
	if collectionSkew > inventorySignatureMaxSkew {
		return ErrInventoryTimestamp
	}

	if inventory.SchemaVersion != protocol.HardwareInventorySchemaVersion ||
		inventory.CollectedAt.IsZero() ||
		strings.TrimSpace(inventory.OS.Name) == "" ||
		strings.TrimSpace(inventory.CPU.Model) == "" ||
		inventory.CPU.LogicalCores <= 0 ||
		inventory.Memory.TotalBytes == 0 {
		return ErrInvalidInventorySubmission
	}

	if len(inventory.GPUs) > 32 ||
		len(inventory.Acceleration) > 64 ||
		len(inventory.AIRuntimes) > 64 ||
		len(inventory.Models) > 200 {
		return ErrInvalidInventorySubmission
	}

	return nil
}
