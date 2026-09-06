package agent

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func BuildSignedInventorySubmission(
	identity Identity,
	inventory protocol.HardwareInventory,
	now time.Time,
) (protocol.InventorySubmission, []byte, string, error) {

	if err := validateIdentity(identity); err != nil {
		return protocol.InventorySubmission{}, nil, "", err
	}

	if inventory.SchemaVersion != protocol.HardwareInventorySchemaVersion {
		return protocol.InventorySubmission{}, nil, "", errors.New("unsupported hardware inventory schema")
	}

	reportID, err := newUUIDv4()
	if err != nil {
		return protocol.InventorySubmission{}, nil, "", err
	}

	submission := protocol.InventorySubmission{
		Version:   protocol.InventorySubmissionVersion,
		NodeID:    identity.NodeID,
		ReportID:  reportID,
		SignedAt:  now.UTC(),
		Inventory: inventory,
	}

	body, err := json.Marshal(submission)
	if err != nil {
		return protocol.InventorySubmission{}, nil, "", err
	}

	privateKeyBytes, err := base64.RawStdEncoding.DecodeString(identity.PrivateKey)
	if err != nil || len(privateKeyBytes) != ed25519.PrivateKeySize {
		return protocol.InventorySubmission{}, nil, "", errors.New("identity private key is invalid")
	}

	signature := ed25519.Sign(
		ed25519.PrivateKey(privateKeyBytes),
		protocol.InventorySigningMessage(body),
	)

	return submission,
		body,
		base64.RawStdEncoding.EncodeToString(signature),
		nil
}
