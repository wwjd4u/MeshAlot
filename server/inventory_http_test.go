package server

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestDecodeInventorySubmissionBodyStrict(t *testing.T) {
	now := time.Date(
		2026,
		9,
		7,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	valid := protocol.InventorySubmission{
		Version:  protocol.InventorySubmissionVersion,
		NodeID:   "11111111-1111-4111-8111-111111111111",
		ReportID: "22222222-2222-4222-8222-222222222222",
		SignedAt: now,
		Inventory: protocol.HardwareInventory{
			SchemaVersion: protocol.HardwareInventorySchemaVersion,
			CollectedAt:   now,
			OS: protocol.OperatingSystemInventory{
				Name:         "Ubuntu",
				Version:      "26.04.1",
				Architecture: "x86_64",
			},
			CPU: protocol.CPUInventory{
				Model:        "Test CPU",
				LogicalCores: 8,
			},
			Memory: protocol.MemoryInventory{
				TotalBytes: 1024,
			},
		},
	}

	body, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := decodeInventorySubmissionBody(body)
	if err != nil {
		t.Fatalf("valid body rejected: %v", err)
	}

	if decoded.ReportID != valid.ReportID {
		t.Fatal("report ID changed during decode")
	}

	unknown := append(
		append([]byte(nil), body[:len(body)-1]...),
		[]byte(`,"unexpected":true}`)...,
	)

	if _, err := decodeInventorySubmissionBody(unknown); !errors.Is(
		err,
		ErrInvalidInventorySubmission,
	) {
		t.Fatal("unknown top-level field was not rejected")
	}

	trailing := append(
		append([]byte(nil), body...),
		[]byte(`{"second":true}`)...,
	)

	if _, err := decodeInventorySubmissionBody(trailing); !errors.Is(
		err,
		ErrInvalidInventorySubmission,
	) {
		t.Fatal("trailing JSON value was not rejected")
	}
}

func TestInventoryBodyLimitIsOneMiB(t *testing.T) {
	if maxInventorySubmissionBodyBytes != 1024*1024 {
		t.Fatalf(
			"unexpected inventory body limit: %d",
			maxInventorySubmissionBodyBytes,
		)
	}
}
