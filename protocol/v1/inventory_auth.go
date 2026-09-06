package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	InventorySubmissionPath    = "/v1/agent/inventory"
	InventorySignatureHeader   = "X-MeshAlot-Signature"
	InventorySigningDomain     = "MESHALOT-M7-INVENTORY"
	InventorySubmissionVersion = "m7-v1"
)

type InventorySubmission struct {
	Version   string            `json:"version"`
	NodeID    string            `json:"node_id"`
	ReportID  string            `json:"report_id"`
	SignedAt  time.Time         `json:"signed_at"`
	Inventory HardwareInventory `json:"inventory"`
}

type InventorySubmissionResponse struct {
	ReportID string `json:"report_id"`
	Accepted bool   `json:"accepted"`
}

func InventorySigningMessage(body []byte) []byte {
	sum := sha256.Sum256(body)

	value := InventorySigningDomain +
		"\nPOST\n" +
		InventorySubmissionPath +
		"\n" +
		hex.EncodeToString(sum[:])

	return []byte(value)
}
