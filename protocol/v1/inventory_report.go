package v1

import "time"

type InventoryReport struct {
	ReportID   string            `json:"report_id"`
	NodeID     string            `json:"node_id"`
	ReceivedAt time.Time         `json:"received_at"`
	Inventory  HardwareInventory `json:"inventory"`
}
