//go:build linux

package agent

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	inventoryCommandTimeout = 4 * time.Second
	inventoryOutputLimit    = 256 << 10
	maxDiscoveredModels     = 200
	maxModelWalkDepth       = 6
)

type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type osCommandRunner struct{}

func (osCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &limitedBuffer{w: &stdout, n: inventoryOutputLimit}
	cmd.Stderr = &limitedBuffer{w: &stderr, n: inventoryOutputLimit}
	err := cmd.Run()
	out := stdout.Bytes()
	if len(out) == 0 {
		out = stderr.Bytes()
	}
	return append([]byte(nil), out...), err
}

type limitedBuffer struct {
	w *bytes.Buffer
	n int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	original := len(p)
	if b.n <= 0 {
		return original, nil
	}
	if len(p) > b.n {
		p = p[:b.n]
	}
	_, _ = b.w.Write(p)
	b.n -= len(p)
	return original, nil
}

func CollectHardwareInventory(ctx context.Context) (protocol.HardwareInventory, error) {
	return collectHardwareInventoryLinux(ctx, osCommandRunner{}, os.UserHomeDir, os.ReadFile, os.Stat)
}

func collectHardwareInventoryLinux(
	ctx context.Context,
	runner commandRunner,
	homeDir func() (string, error),
	readFile func(string) ([]byte, error),
	stat func(string) (fs.FileInfo, error),
) (protocol.HardwareInventory, error) {
	inv := protocol.HardwareInventory{
		SchemaVersion: protocol.HardwareInventorySchemaVersion,
		CollectedAt:   time.Now().UTC(),
		GPUs:          []protocol.GPUInventory{},
		Acceleration:  []protocol.AccelerationRuntime{},
		AIRuntimes:    []protocol.AIRuntimeInventory{},
		Models:        []protocol.ModelInventory{},
	}

	inv.OS = detectLinuxOS(ctx, runner, readFile)
	inv.CPU = detectLinuxCPU(ctx, runner)
	inv.Memory = detectLinuxMemory(readFile)
	inv.GPUs = detectLinuxGPUs(ctx, runner)
	inv.Acceleration = detectLinuxAcceleration(ctx, runner, stat)

	home, _ := homeDir()
	inv.AIRuntimes = detectLinuxAIRuntimes(ctx, runner, home, stat)
	inv.Models = detectLinuxModels(ctx, runner, home)

	if inv.OS.Name == "" || inv.CPU.LogicalCores <= 0 || inv.Memory.TotalBytes == 0 {
		return inv, errors.New("hardware inventory missing required OS, CPU, or memory information")
	}
	return inv, nil
}

func detectLinuxOS(ctx context.Context, runner commandRunner, readFile func(string) ([]byte, error)) protocol.OperatingSystemInventory {
	result := protocol.OperatingSystemInventory{Architecture: runtime.GOARCH}
	if raw, err := readFile("/etc/os-release"); err == nil {
		values := parseOSRelease(string(raw))
		result.Name = firstNonEmpty(values["NAME"], values["ID"], "Linux")
		result.Version = firstNonEmpty(values["VERSION"], values["VERSION_ID"])
	}
	if out, err := runBounded(ctx, runner, "uname", "-r"); err == nil {
		result.Kernel = strings.TrimSpace(string(out))
	}
	if out, err := runBounded(ctx, runner, "uname", "-m"); err == nil && strings.TrimSpace(string(out)) != "" {
		result.Architecture = strings.TrimSpace(string(out))
	}
	if result.Name == "" {
		result.Name = "Linux"
	}
	return result
}

func parseOSRelease(raw string) map[string]string {
	values := map[string]string{}
	s := bufio.NewScanner(strings.NewReader(raw))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		values[strings.TrimSpace(key)] = value
	}
	return values
}

func detectLinuxCPU(ctx context.Context, runner commandRunner) protocol.CPUInventory {
	result := protocol.CPUInventory{LogicalCores: runtime.NumCPU()}
	out, err := runBounded(ctx, runner, "lscpu")
	if err != nil {
		return result
	}
	fields := parseColonFields(string(out))
	result.Model = strings.TrimSpace(fields["Model name"])
	if v, err := strconv.Atoi(strings.TrimSpace(fields["CPU(s)"])); err == nil && v > 0 {
		result.LogicalCores = v
	}
	cores, _ := strconv.Atoi(strings.TrimSpace(fields["Core(s) per socket"]))
	sockets, _ := strconv.Atoi(strings.TrimSpace(fields["Socket(s)"]))
	if cores > 0 && sockets > 0 {
		result.PhysicalCores = cores * sockets
	}
	return result
}

func parseColonFields(raw string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if ok {
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return values
}

func detectLinuxMemory(readFile func(string) ([]byte, error)) protocol.MemoryInventory {
	raw, err := readFile("/proc/meminfo")
	if err != nil {
		return protocol.MemoryInventory{}
	}
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "MemTotal:" {
			kb, err := strconv.ParseUint(fields[1], 10, 64)
			if err == nil {
				return protocol.MemoryInventory{TotalBytes: kb * 1024}
			}
		}
	}
	return protocol.MemoryInventory{}
}

func detectLinuxGPUs(ctx context.Context, runner commandRunner) []protocol.GPUInventory {
	var gpus []protocol.GPUInventory
	out, err := runBounded(ctx, runner, "lspci", "-nnk")
	if err == nil {
		gpus = parseLSPCIGPUs(string(out))
	}
	vulkan, err := runBounded(ctx, runner, "vulkaninfo", "--summary")
	if err == nil {
		devices := parseVulkanSummary(string(vulkan))
		physical := make([]protocol.GPUInventory, 0, len(devices))
		for _, dev := range devices {
			if dev.DeviceType != "cpu" {
				physical = append(physical, dev)
			}
		}
		for _, dev := range physical {
			matched := false
			for i := range gpus {
				if (len(gpus) == 1 && len(physical) == 1) || sameGPU(gpus[i], dev) {
					gpus[i].Model = firstNonEmpty(dev.Model, gpus[i].Model)
					gpus[i].DeviceType = firstNonEmpty(dev.DeviceType, gpus[i].DeviceType)
					gpus[i].DriverName = dev.DriverName
					gpus[i].DriverVersion = dev.DriverVersion
					if gpus[i].DeviceType == "integrated" {
						gpus[i].MemoryType = "shared"
					}
					matched = true
					break
				}
			}
			if !matched {
				gpus = append(gpus, dev)
			}
		}
	}
	if out, err := runBounded(ctx, runner, "nvidia-smi", "--query-gpu=name,memory.total,driver_version", "--format=csv,noheader,nounits"); err == nil {
		gpus = mergeNvidiaSMI(gpus, string(out))
	}
	return gpus
}

func mergeNvidiaSMI(gpus []protocol.GPUInventory, raw string) []protocol.GPUInventory {
	for _, line := range strings.Split(raw, "\n") {
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		model := strings.TrimSpace(parts[0])
		mib, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		if model == "" || err != nil {
			continue
		}
		bytes := mib * 1024 * 1024
		driver := strings.TrimSpace(parts[2])
		matched := false
		for i := range gpus {
			if strings.EqualFold(gpus[i].Vendor, "NVIDIA") && sameGPU(gpus[i], protocol.GPUInventory{Model: model}) {
				gpus[i].Vendor = "NVIDIA"
				gpus[i].Model = model
				gpus[i].DeviceType = "discrete"
				gpus[i].MemoryType = "dedicated"
				gpus[i].UsableVRAMBytes = &bytes
				gpus[i].DriverName = "NVIDIA"
				gpus[i].DriverVersion = driver
				matched = true
				break
			}
		}
		if !matched {
			gpus = append(gpus, protocol.GPUInventory{
				Vendor: "NVIDIA", Model: model, DeviceType: "discrete", MemoryType: "dedicated",
				UsableVRAMBytes: &bytes, DriverName: "NVIDIA", DriverVersion: driver,
			})
		}
	}
	return gpus
}

var pciGPUHeader = regexp.MustCompile(`(?i)(VGA compatible controller|3D controller|Display controller).*?:\s*(.+)$`)

func parseLSPCIGPUs(raw string) []protocol.GPUInventory {
	lines := strings.Split(raw, "\n")
	var result []protocol.GPUInventory
	for i := 0; i < len(lines); i++ {
		match := pciGPUHeader.FindStringSubmatch(lines[i])
		if len(match) != 3 {
			continue
		}
		model := stripPCIIDs(strings.TrimSpace(match[2]))
		gpu := protocol.GPUInventory{
			Vendor:     vendorFromText(model),
			Model:      model,
			DeviceType: "unknown",
			MemoryType: "unknown",
		}
		for j := i + 1; j < len(lines) && j <= i+6; j++ {
			line := strings.TrimSpace(lines[j])
			if line == "" || (!strings.HasPrefix(lines[j], "\t") && !strings.HasPrefix(lines[j], " ")) {
				break
			}
			if strings.HasPrefix(line, "Kernel driver in use:") {
				gpu.KernelDriver = strings.TrimSpace(strings.TrimPrefix(line, "Kernel driver in use:"))
			}
		}
		result = append(result, gpu)
	}
	return result
}

var bracketID = regexp.MustCompile(`\s*\[[0-9a-fA-F]{4}:[0-9a-fA-F]{4}\](?:\s*\(rev [0-9a-fA-F]+\))?`)

func stripPCIIDs(value string) string {
	return strings.TrimSpace(bracketID.ReplaceAllString(value, ""))
}

func vendorFromText(value string) string {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "intel"):
		return "Intel"
	case strings.Contains(lower, "nvidia"):
		return "NVIDIA"
	case strings.Contains(lower, "amd"), strings.Contains(lower, "advanced micro devices"), strings.Contains(lower, "ati"):
		return "AMD"
	default:
		return "Unknown"
	}
}

func parseVulkanSummary(raw string) []protocol.GPUInventory {
	var result []protocol.GPUInventory
	var fields map[string]string
	flush := func() {
		if fields == nil {
			return
		}
		model := fields["deviceName"]
		if model == "" {
			return
		}
		deviceType := strings.ToLower(strings.TrimPrefix(fields["deviceType"], "PHYSICAL_DEVICE_TYPE_"))
		deviceType = strings.ReplaceAll(deviceType, "_gpu", "")
		dev := protocol.GPUInventory{
			Vendor:        vendorFromText(model),
			Model:         model,
			DeviceType:    deviceType,
			MemoryType:    "unknown",
			DriverName:    fields["driverName"],
			DriverVersion: firstNonEmpty(extractMesaVersion(fields["driverInfo"]), fields["driverVersion"]),
		}
		if dev.DeviceType == "integrated" {
			dev.MemoryType = "shared"
		}
		result = append(result, dev)
	}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if regexp.MustCompile(`^GPU[0-9]+:$`).MatchString(trimmed) {
			flush()
			fields = map[string]string{}
			continue
		}
		if fields == nil {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if ok {
			fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	flush()
	return result
}

func extractMesaVersion(value string) string {
	fields := strings.Fields(value)
	for i := 0; i+1 < len(fields); i++ {
		if strings.EqualFold(fields[i], "Mesa") {
			return strings.TrimSuffix(fields[i+1], ",")
		}
	}
	return ""
}

func sameGPU(a, b protocol.GPUInventory) bool {
	aModel := normalizeGPUModel(a.Model)
	bModel := normalizeGPUModel(b.Model)
	return aModel != "" && bModel != "" && (strings.Contains(aModel, bModel) || strings.Contains(bModel, aModel))
}

func normalizeGPUModel(value string) string {
	replacer := strings.NewReplacer("[", " ", "]", " ", "(", " ", ")", " ", "-", " ", "_", " ")
	fields := strings.Fields(strings.ToLower(replacer.Replace(value)))
	ignored := map[string]bool{
		"nvidia": true, "corporation": true, "geforce": true, "intel": true,
		"advanced": true, "micro": true, "devices": true, "amd": true, "graphics": true,
	}
	kept := fields[:0]
	for _, field := range fields {
		if !ignored[field] {
			kept = append(kept, field)
		}
	}
	return strings.Join(kept, " ")
}

func detectLinuxAcceleration(ctx context.Context, runner commandRunner, stat func(string) (fs.FileInfo, error)) []protocol.AccelerationRuntime {
	var result []protocol.AccelerationRuntime
	if out, err := runBounded(ctx, runner, "vulkaninfo", "--summary"); err == nil {
		version := ""
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Vulkan Instance Version:") {
				version = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "Vulkan Instance Version:"))
				break
			}
		}
		result = append(result, protocol.AccelerationRuntime{Name: "Vulkan", Version: version, Status: "available"})
	}
	if out, err := runBounded(ctx, runner, "nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader"); err == nil && strings.TrimSpace(string(out)) != "" {
		result = append(result, protocol.AccelerationRuntime{Name: "CUDA", Status: "available"})
	}
	if _, err := runBounded(ctx, runner, "rocminfo"); err == nil {
		result = append(result, protocol.AccelerationRuntime{Name: "ROCm", Status: "available"})
	}
	if _, err := stat("/opt/intel/oneapi/compiler/latest/bin/sycl-ls"); err == nil {
		out, runErr := runBounded(ctx, runner, "/opt/intel/oneapi/compiler/latest/bin/sycl-ls")
		status := "available"
		if runErr != nil || strings.Contains(strings.ToLower(string(out)), "no platforms found") || strings.TrimSpace(string(out)) == "" {
			status = "installed_unavailable"
		}
		result = append(result, protocol.AccelerationRuntime{Name: "SYCL", Status: status})
	}
	return result
}

func detectLinuxAIRuntimes(ctx context.Context, runner commandRunner, home string, stat func(string) (fs.FileInfo, error)) []protocol.AIRuntimeInventory {
	var result []protocol.AIRuntimeInventory
	if out, err := runBounded(ctx, runner, "ollama", "--version"); err == nil {
		result = append(result, protocol.AIRuntimeInventory{Name: "Ollama", Version: parseTrailingVersion(string(out)), Status: "available"})
	}
	seenLlama := false
	for _, candidate := range []string{
		filepath.Join(home, "llama.cpp", "build", "bin", "llama-server"),
		filepath.Join(home, "llama.cpp", "build-sycl", "bin", "llama-server"),
		filepath.Join(home, "llama.cpp", "build-vulkan", "bin", "llama-server"),
	} {
		if _, err := stat(candidate); err != nil {
			continue
		}
		out, runErr := runBounded(ctx, runner, candidate, "--version")
		status := "available"
		if runErr != nil {
			status = "installed_unavailable"
		}
		if !seenLlama || status == "available" {
			entry := protocol.AIRuntimeInventory{Name: "llama.cpp", Version: parseLlamaVersion(string(out)), Status: status}
			if !seenLlama {
				result = append(result, entry)
			} else if status == "available" {
				for i := range result {
					if result[i].Name == "llama.cpp" {
						result[i] = entry
					}
				}
			}
			seenLlama = true
		}
	}
	return result
}

func detectLinuxModels(ctx context.Context, runner commandRunner, home string) []protocol.ModelInventory {
	models := map[string]protocol.ModelInventory{}
	if out, err := runBounded(ctx, runner, "ollama", "list"); err == nil {
		for _, model := range parseOllamaList(string(out)) {
			models["ollama\x00"+model.Name] = model
		}
	}
	roots := []string{
		filepath.Join(home, ".cache", "huggingface", "hub"),
		filepath.Join(home, ".cache", "llama.cpp"),
		filepath.Join(home, "models"),
	}
	for _, root := range roots {
		for _, model := range discoverGGUFModels(root) {
			key := "gguf\x00" + strings.ToLower(model.Name) + fmt.Sprintf("\x00%d", model.SizeBytes)
			models[key] = model
			if len(models) >= maxDiscoveredModels {
				break
			}
		}
	}
	result := make([]protocol.ModelInventory, 0, len(models))
	for _, model := range models {
		result = append(result, model)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Runtime != result[j].Runtime {
			return result[i].Runtime < result[j].Runtime
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func parseOllamaList(raw string) []protocol.ModelInventory {
	var result []protocol.ModelInventory
	for i, line := range strings.Split(raw, "\n") {
		if i == 0 && strings.HasPrefix(strings.TrimSpace(line), "NAME") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		model := protocol.ModelInventory{Name: fields[0], Runtime: "Ollama", Format: "ollama"}
		if len(fields) >= 4 {
			if size, ok := parseHumanBytes(fields[2], fields[3]); ok {
				model.SizeBytes = size
			}
		}
		result = append(result, model)
	}
	return result
}

func parseHumanBytes(value, unit string) (uint64, bool) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil || v < 0 {
		return 0, false
	}
	multipliers := map[string]float64{"KB": 1e3, "MB": 1e6, "GB": 1e9, "TB": 1e12}
	m, ok := multipliers[strings.ToUpper(unit)]
	if !ok {
		return 0, false
	}
	return uint64(v * m), true
}

var multipartGGUF = regexp.MustCompile(`(?i)^(.*)-([0-9]{5})-of-([0-9]{5})\.gguf$`)
var quantizationPattern = regexp.MustCompile(`(?i)(Q[0-9]+(?:_[A-Z0-9]+)+|IQ[0-9]+_[A-Z0-9]+|UD-IQ[0-9]+_[A-Z0-9]+)`)

func discoverGGUFModels(root string) []protocol.ModelInventory {
	rootInfo, err := os.Lstat(root)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil
	}

	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil
	}
	resolvedRoot, err = filepath.Abs(resolvedRoot)
	if err != nil {
		return nil
	}

	type aggregate struct {
		name string
		size uint64
	}
	groups := map[string]*aggregate{}

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		depth := 0
		if rel != "." {
			depth = len(strings.Split(rel, string(filepath.Separator)))
		}

		if d.IsDir() && depth > maxModelWalkDepth {
			return filepath.SkipDir
		}

		if d.IsDir() || depth > maxModelWalkDepth {
			return nil
		}

		if !strings.EqualFold(filepath.Ext(d.Name()), ".gguf") {
			return nil
		}

		if strings.HasPrefix(strings.ToLower(d.Name()), "mmproj-") {
			return nil
		}

		var size int64

		if d.Type()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil
			}

			resolved, err = filepath.Abs(resolved)
			if err != nil || !pathInsideRoot(resolvedRoot, resolved) {
				return nil
			}

			targetInfo, err := os.Stat(resolved)
			if err != nil || !targetInfo.Mode().IsRegular() || targetInfo.Size() < 0 {
				return nil
			}

			size = targetInfo.Size()
		} else {
			fileInfo, err := d.Info()
			if err != nil || !fileInfo.Mode().IsRegular() || fileInfo.Size() < 0 {
				return nil
			}

			size = fileInfo.Size()
		}

		name := d.Name()
		key := strings.ToLower(name)

		if match := multipartGGUF.FindStringSubmatch(name); len(match) == 4 {
			name = match[1] + ".gguf"
			key = strings.ToLower(name)
		}

		entry := groups[key]
		if entry == nil {
			entry = &aggregate{name: name}
			groups[key] = entry
		}

		entry.size += uint64(size)
		return nil
	})

	result := make([]protocol.ModelInventory, 0, len(groups))

	for _, group := range groups {
		base := strings.TrimSuffix(group.name, filepath.Ext(group.name))

		result = append(result, protocol.ModelInventory{
			Name:         base,
			Runtime:      "llama.cpp",
			Format:       "GGUF",
			Quantization: parseQuantization(base),
			SizeBytes:    group.size,
		})
	}

	return result
}

func pathInsideRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(rel) {
		return false
	}

	return rel != ".." &&
		!strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func parseQuantization(name string) string {
	if match := quantizationPattern.FindString(name); match != "" {
		return match
	}
	return ""
}

func parseTrailingVersion(raw string) string {
	fields := strings.Fields(raw)
	for i := len(fields) - 1; i >= 0; i-- {
		if regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)+`).MatchString(fields[i]) {
			return strings.TrimSpace(fields[i])
		}
	}
	return ""
}

func parseLlamaVersion(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		}
	}
	return ""
}

func runBounded(parent context.Context, runner commandRunner, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, inventoryCommandTimeout)
	defer cancel()
	return runner.Run(ctx, name, args...)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
