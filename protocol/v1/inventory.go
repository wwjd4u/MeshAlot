package v1

import "time"

const HardwareInventorySchemaVersion = "m7-v1"

type HardwareInventory struct {
	SchemaVersion string                   `json:"schema_version"`
	CollectedAt   time.Time                `json:"collected_at"`
	OS            OperatingSystemInventory `json:"os"`
	CPU           CPUInventory             `json:"cpu"`
	Memory        MemoryInventory          `json:"memory"`
	GPUs          []GPUInventory           `json:"gpus"`
	Acceleration  []AccelerationRuntime    `json:"acceleration_runtimes"`
	AIRuntimes    []AIRuntimeInventory     `json:"ai_runtimes"`
	Models        []ModelInventory         `json:"models"`
}

type OperatingSystemInventory struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	Kernel       string `json:"kernel,omitempty"`
}

type CPUInventory struct {
	Model         string `json:"model"`
	PhysicalCores int    `json:"physical_cores,omitempty"`
	LogicalCores  int    `json:"logical_cores"`
}

type MemoryInventory struct {
	TotalBytes uint64 `json:"total_bytes"`
}

type GPUInventory struct {
	Vendor          string  `json:"vendor"`
	Model           string  `json:"model"`
	DeviceType      string  `json:"device_type"`
	MemoryType      string  `json:"memory_type"`
	UsableVRAMBytes *uint64 `json:"usable_vram_bytes,omitempty"`
	KernelDriver    string  `json:"kernel_driver,omitempty"`
	DriverName      string  `json:"driver_name,omitempty"`
	DriverVersion   string  `json:"driver_version,omitempty"`
}

type AccelerationRuntime struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Status  string `json:"status"`
}

type AIRuntimeInventory struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Status  string `json:"status"`
}

type ModelInventory struct {
	Name         string `json:"name"`
	Runtime      string `json:"runtime,omitempty"`
	Format       string `json:"format"`
	Quantization string `json:"quantization,omitempty"`
	SizeBytes    uint64 `json:"size_bytes,omitempty"`
}
