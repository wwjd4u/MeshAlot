//go:build linux

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type fakeRunner map[string]fakeCommand

type fakeCommand struct {
	out string
	err error
}

func (f fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name + "\x00" + strings.Join(args, "\x00")
	cmd, ok := f[key]
	if !ok {
		return nil, errors.New("not installed")
	}
	return []byte(cmd.out), cmd.err
}

func TestNode001Parsers(t *testing.T) {
	lspci := `00:02.0 VGA compatible controller [0300]: Intel Corporation Arrow Lake-S [Intel Graphics] [8086:7d67] (rev 06)
	DeviceName: Onboard - Video
	Subsystem: Intel Corporation Device [8086:2212]
	Kernel driver in use: i915
	Kernel modules: i915, xe
`
	vulkan := `Vulkan Instance Version: 1.4.341

GPU0:
	apiVersion         = 1.4.335
	driverVersion      = 26.0.8
	vendorID           = 0x8086
	deviceID           = 0x7d67
	deviceType         = PHYSICAL_DEVICE_TYPE_INTEGRATED_GPU
	deviceName         = Intel(R) Graphics (ARL)
	driverID           = DRIVER_ID_INTEL_OPEN_SOURCE_MESA
	driverName         = Intel open-source Mesa driver
	driverInfo         = Mesa 26.0.8-1ubuntu0.3
GPU1:
	apiVersion         = 1.4.335
	driverVersion      = 26.0.8
	deviceType         = PHYSICAL_DEVICE_TYPE_CPU
	deviceName         = llvmpipe (LLVM 21.1.8, 256 bits)
	driverName         = llvmpipe
	driverInfo         = Mesa 26.0.8-1ubuntu0.3 (LLVM 21.1.8)
`
	gpus := parseLSPCIGPUs(lspci)
	if len(gpus) != 1 || gpus[0].Vendor != "Intel" || gpus[0].KernelDriver != "i915" {
		t.Fatalf("unexpected lspci parse: %#v", gpus)
	}
	devices := parseVulkanSummary(vulkan)
	if len(devices) != 2 {
		t.Fatalf("expected Intel GPU plus llvmpipe CPU device, got %#v", devices)
	}
	if devices[0].Model != "Intel(R) Graphics (ARL)" || devices[0].DeviceType != "integrated" || devices[0].MemoryType != "shared" || devices[0].DriverVersion != "26.0.8-1ubuntu0.3" {
		t.Fatalf("unexpected Vulkan GPU: %#v", devices[0])
	}
	if devices[1].DeviceType != "cpu" {
		t.Fatalf("llvmpipe must remain a CPU device: %#v", devices[1])
	}
}

func TestNvidiaVRAMParsing(t *testing.T) {
	gpus := []protocol.GPUInventory{{Vendor: "NVIDIA", Model: "NVIDIA Corporation AD102 [GeForce RTX 4090]", KernelDriver: "nvidia"}}
	gpus = mergeNvidiaSMI(gpus, "NVIDIA GeForce RTX 4090, 24564, 580.82.09\n")
	if len(gpus) != 1 || gpus[0].UsableVRAMBytes == nil {
		t.Fatalf("NVIDIA GPU/VRAM not merged: %#v", gpus)
	}
	if *gpus[0].UsableVRAMBytes != 24564*1024*1024 || gpus[0].MemoryType != "dedicated" || gpus[0].DriverVersion != "580.82.09" {
		t.Fatalf("unexpected NVIDIA inventory: %#v", gpus[0])
	}
}

func TestNode001FullInventoryFixture(t *testing.T) {
	home := t.TempDir()
	for _, rel := range []string{
		"llama.cpp/build/bin/llama-server",
		"llama.cpp/build-sycl/bin/llama-server",
	} {
		path := filepath.Join(home, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("fixture"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	modelRoot := filepath.Join(home, "models")
	if err := os.MkdirAll(modelRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]int64{
		"DeepSeek-V4-Flash-0731-UD-IQ1_S-00001-of-00003.gguf": 5257664,
		"DeepSeek-V4-Flash-0731-UD-IQ1_S-00002-of-00003.gguf": 49093726624,
		"DeepSeek-V4-Flash-0731-UD-IQ1_S-00003-of-00003.gguf": 33440253504,
		"mmproj-Qwen3VL-30B-A3B-Instruct-F16.gguf":            1083499584,
		"Qwen3VL-30B-A3B-Instruct-Q4_K_M.gguf":                18556687168,
	}
	for name, size := range files {
		path := filepath.Join(modelRoot, name)
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(size); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	commands := fakeRunner{
		"uname\x00-r": {out: "7.0.0-30-generic\n"},
		"uname\x00-m": {out: "x86_64\n"},
		"lscpu\x00": {out: `Architecture: x86_64
CPU(s): 14
Model name: Intel(R) Core(TM) Ultra 5 235HX
Thread(s) per core: 1
Core(s) per socket: 14
Socket(s): 1
`},
		"lspci\x00-nnk": {out: `00:02.0 VGA compatible controller [0300]: Intel Corporation Arrow Lake-S [Intel Graphics] [8086:7d67] (rev 06)
	Kernel driver in use: i915
	Kernel modules: i915, xe
`},
		"vulkaninfo\x00--summary": {out: `Vulkan Instance Version: 1.4.341
GPU0:
	deviceType = PHYSICAL_DEVICE_TYPE_INTEGRATED_GPU
	deviceName = Intel(R) Graphics (ARL)
	driverName = Intel open-source Mesa driver
	driverInfo = Mesa 26.0.8-1ubuntu0.3
GPU1:
	deviceType = PHYSICAL_DEVICE_TYPE_CPU
	deviceName = llvmpipe (LLVM 21.1.8, 256 bits)
	driverName = llvmpipe
`},
		"ollama\x00--version": {out: "ollama version is 0.32.5\n"},
		"ollama\x00list": {out: `NAME                                  ID              SIZE      MODIFIED
qwen3-30b-hermes:latest               70d26e2ff80b    18 GB     2 weeks ago
qwen3-30b-local:latest                813ae2da0f3c    18 GB     2 weeks ago
qwen3:30b-a3b-instruct-2507-q4_K_M    19e422b02313    18 GB     2 weeks ago
qwen3-embedding:0.6b                  ac6da0dfba84    639 MB    4 weeks ago
qwen3:14b                             bdbd181c33f2    9.3 GB    5 weeks ago
`},
		filepath.Join(home, "llama.cpp", "build", "bin", "llama-server") + "\x00--version":      {out: "version: 0.1.0-dev (build 10454, commit 4df29be4f)\nbuilt with GNU 15.2.0 for Linux x86_64\n"},
		filepath.Join(home, "llama.cpp", "build-sycl", "bin", "llama-server") + "\x00--version": {out: "error while loading shared libraries: libsvml.so", err: errors.New("exit 127")},
		"/opt/intel/oneapi/compiler/latest/bin/sycl-ls\x00":                                     {out: "No platforms found - run with '--verbose' to get more details.\n"},
	}
	readFile := func(path string) ([]byte, error) {
		switch path {
		case "/etc/os-release":
			return []byte("PRETTY_NAME=\"Ubuntu 26.04.1 LTS\"\nNAME=\"Ubuntu\"\nVERSION_ID=\"26.04\"\nVERSION=\"26.04.1 LTS (Resolute Raccoon)\"\n"), nil
		case "/proc/meminfo":
			return []byte("MemTotal:       128954384 kB\n"), nil
		default:
			return nil, os.ErrNotExist
		}
	}
	stat := func(path string) (os.FileInfo, error) {
		if path == "/opt/intel/oneapi/compiler/latest/bin/sycl-ls" {
			return fakeInfo{name: "sycl-ls", mode: 0o755}, nil
		}
		return os.Stat(path)
	}
	inv, err := collectHardwareInventoryLinux(context.Background(), commands, func() (string, error) { return home, nil }, readFile, stat)
	if err != nil {
		t.Fatal(err)
	}
	if inv.OS.Name != "Ubuntu" || inv.OS.Version != "26.04.1 LTS (Resolute Raccoon)" || inv.OS.Architecture != "x86_64" || inv.OS.Kernel != "7.0.0-30-generic" {
		t.Fatalf("unexpected OS: %#v", inv.OS)
	}
	if inv.CPU.Model != "Intel(R) Core(TM) Ultra 5 235HX" || inv.CPU.LogicalCores != 14 || inv.CPU.PhysicalCores != 14 {
		t.Fatalf("unexpected CPU: %#v", inv.CPU)
	}
	if inv.Memory.TotalBytes != 132049289216 {
		t.Fatalf("unexpected RAM: %d", inv.Memory.TotalBytes)
	}
	if len(inv.GPUs) != 1 || inv.GPUs[0].Model != "Intel(R) Graphics (ARL)" || inv.GPUs[0].MemoryType != "shared" || inv.GPUs[0].UsableVRAMBytes != nil || inv.GPUs[0].KernelDriver != "i915" {
		t.Fatalf("unexpected GPU: %#v", inv.GPUs)
	}
	if !hasAcceleration(inv.Acceleration, "Vulkan", "available") || !hasAcceleration(inv.Acceleration, "SYCL", "installed_unavailable") {
		t.Fatalf("unexpected acceleration: %#v", inv.Acceleration)
	}
	if !hasRuntime(inv.AIRuntimes, "Ollama", "available") || !hasRuntime(inv.AIRuntimes, "llama.cpp", "available") {
		t.Fatalf("unexpected AI runtimes: %#v", inv.AIRuntimes)
	}
	if countRuntimeModels(inv.Models, "Ollama") != 5 {
		t.Fatalf("expected 5 Ollama model tags, got %#v", inv.Models)
	}
	if model, ok := findModel(inv.Models, "DeepSeek-V4-Flash-0731-UD-IQ1_S"); !ok || model.SizeBytes != 82539237792 || model.Quantization != "UD-IQ1_S" {
		t.Fatalf("multipart DeepSeek model was not aggregated correctly: %#v", model)
	}
	if _, ok := findModel(inv.Models, "mmproj-Qwen3VL-30B-A3B-Instruct-F16"); ok {
		t.Fatal("mmproj helper must not be reported as a runnable model")
	}
	payload, err := json.Marshal(inv)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), home) || strings.Contains(string(payload), "/home/") {
		t.Fatalf("inventory leaked a local filesystem path: %s", payload)
	}
}

func TestGGUFSymlinkPrivacyBoundary(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "models")

	if err := os.MkdirAll(filepath.Join(root, "blobs"), 0o755); err != nil {
		t.Fatal(err)
	}

	internal := map[string]struct {
		blob string
		size int64
	}{
		"Qwen2.5-7B-Instruct-1M-Q4_K_M.gguf": {
			blob: "blob-1m",
			size: 4683074144,
		},
		"Qwen2.5-7B-Instruct-Q4_K_M.gguf": {
			blob: "blob-standard",
			size: 4683074240,
		},
	}

	for modelName, fixture := range internal {
		target := filepath.Join(root, "blobs", fixture.blob)

		f, err := os.Create(target)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(fixture.size); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(
			filepath.Join("blobs", fixture.blob),
			filepath.Join(root, modelName),
		); err != nil {
			t.Fatal(err)
		}
	}

	outsideRoot := t.TempDir()
	outsideTarget := filepath.Join(outsideRoot, "private-blob")

	f, err := os.Create(outsideTarget)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(999); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(
		outsideTarget,
		filepath.Join(root, "Outside-Q4_K_M.gguf"),
	); err != nil {
		t.Fatal(err)
	}

	models := discoverGGUFModels(root)

	for name, fixture := range internal {
		expectedName := strings.TrimSuffix(name, ".gguf")

		model, ok := findModel(models, expectedName)
		if !ok {
			t.Fatalf("safe in-root model symlink missing: %s", expectedName)
		}
		if model.SizeBytes != uint64(fixture.size) {
			t.Fatalf(
				"wrong size for %s: got %d want %d",
				expectedName,
				model.SizeBytes,
				fixture.size,
			)
		}
	}

	if _, ok := findModel(models, "Outside-Q4_K_M"); ok {
		t.Fatal("symlink escaping approved model root must not be reported")
	}

	symlinkRoot := filepath.Join(parent, "models-link")
	if err := os.Symlink(root, symlinkRoot); err != nil {
		t.Fatal(err)
	}

	if models := discoverGGUFModels(symlinkRoot); len(models) != 0 {
		t.Fatal("symlinked model root must not be scanned")
	}
}

type fakeInfo struct {
	name string
	mode os.FileMode
}

func (f fakeInfo) Name() string           { return f.name }
func (f fakeInfo) Size() int64            { return 0 }
func (f fakeInfo) Mode() os.FileMode      { return f.mode }
func (f fakeInfo) ModTime() (t time.Time) { return }
func (f fakeInfo) IsDir() bool            { return false }
func (f fakeInfo) Sys() any               { return nil }

func hasAcceleration(items []protocol.AccelerationRuntime, name, status string) bool {
	for _, item := range items {
		if item.Name == name && item.Status == status {
			return true
		}
	}
	return false
}
func hasRuntime(items []protocol.AIRuntimeInventory, name, status string) bool {
	for _, item := range items {
		if item.Name == name && item.Status == status {
			return true
		}
	}
	return false
}
func countRuntimeModels(items []protocol.ModelInventory, runtime string) int {
	n := 0
	for _, item := range items {
		if item.Runtime == runtime {
			n++
		}
	}
	return n
}
func findModel(items []protocol.ModelInventory, name string) (protocol.ModelInventory, bool) {
	for _, item := range items {
		if item.Name == name {
			return item, true
		}
	}
	return protocol.ModelInventory{}, false
}
