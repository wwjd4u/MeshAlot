//go:build darwin

package agent

import (
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
	darwinCommandTimeout = 8 * time.Second
	darwinOutputLimit    = 256 << 10
	darwinModelLimit     = 200
	darwinModelMaxDepth  = 6
)

type darwinRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type darwinOSRunner struct{}

func (darwinOSRunner) Run(
	ctx context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &darwinLimitedBuffer{
		w: &stdout,
		n: darwinOutputLimit,
	}
	cmd.Stderr = &darwinLimitedBuffer{
		w: &stderr,
		n: darwinOutputLimit,
	}

	err := cmd.Run()

	out := stdout.Bytes()
	if len(out) == 0 {
		out = stderr.Bytes()
	}

	return append([]byte(nil), out...), err
}

type darwinLimitedBuffer struct {
	w *bytes.Buffer
	n int
}

func (b *darwinLimitedBuffer) Write(p []byte) (int, error) {
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

func runDarwin(
	ctx context.Context,
	runner darwinRunner,
	name string,
	args ...string,
) ([]byte, error) {
	runCtx, cancel := context.WithTimeout(
		ctx,
		darwinCommandTimeout,
	)
	defer cancel()

	return runner.Run(runCtx, name, args...)
}

func CollectHardwareInventory(
	ctx context.Context,
) (protocol.HardwareInventory, error) {
	return collectHardwareInventoryDarwin(
		ctx,
		darwinOSRunner{},
		os.UserHomeDir,
		os.Lstat,
		os.Stat,
	)
}

func collectHardwareInventoryDarwin(
	ctx context.Context,
	runner darwinRunner,
	homeDir func() (string, error),
	lstat func(string) (fs.FileInfo, error),
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

	inv.OS = detectDarwinOS(ctx, runner)
	inv.CPU = detectDarwinCPU(ctx, runner)
	inv.Memory = detectDarwinMemory(ctx, runner)

	displayRaw, _ := runDarwin(
		ctx,
		runner,
		"/usr/sbin/system_profiler",
		"SPDisplaysDataType",
	)

	inv.GPUs = parseDarwinGPUs(string(displayRaw))
	inv.Acceleration = detectDarwinAcceleration(
		ctx,
		runner,
		string(displayRaw),
		lstat,
		stat,
	)

	home, _ := homeDir()

	inv.AIRuntimes = detectDarwinAIRuntimes(
		ctx,
		runner,
		home,
		lstat,
		stat,
	)

	inv.Models = detectDarwinModels(
		ctx,
		runner,
		home,
		lstat,
		stat,
	)

	if inv.OS.Name == "" ||
		inv.CPU.LogicalCores <= 0 ||
		inv.Memory.TotalBytes == 0 {
		return inv, errors.New(
			"hardware inventory missing required OS, CPU, or memory information",
		)
	}

	return inv, nil
}

func detectDarwinOS(
	ctx context.Context,
	runner darwinRunner,
) protocol.OperatingSystemInventory {
	result := protocol.OperatingSystemInventory{
		Name:         "macOS",
		Architecture: runtime.GOARCH,
	}

	if out, err := runDarwin(
		ctx,
		runner,
		"/usr/bin/sw_vers",
		"-productName",
	); err == nil {
		if value := strings.TrimSpace(string(out)); value != "" {
			result.Name = value
		}
	}

	if out, err := runDarwin(
		ctx,
		runner,
		"/usr/bin/sw_vers",
		"-productVersion",
	); err == nil {
		result.Version = strings.TrimSpace(string(out))
	}

	if out, err := runDarwin(
		ctx,
		runner,
		"/usr/bin/uname",
		"-r",
	); err == nil {
		result.Kernel = strings.TrimSpace(string(out))
	}

	if out, err := runDarwin(
		ctx,
		runner,
		"/usr/bin/uname",
		"-m",
	); err == nil {
		if value := strings.TrimSpace(string(out)); value != "" {
			result.Architecture = value
		}
	}

	return result
}

func detectDarwinCPU(
	ctx context.Context,
	runner darwinRunner,
) protocol.CPUInventory {
	result := protocol.CPUInventory{
		LogicalCores: runtime.NumCPU(),
	}

	if out, err := runDarwin(
		ctx,
		runner,
		"/usr/sbin/sysctl",
		"-n",
		"machdep.cpu.brand_string",
	); err == nil {
		result.Model = strings.TrimSpace(string(out))
	}

	if out, err := runDarwin(
		ctx,
		runner,
		"/usr/sbin/sysctl",
		"-n",
		"hw.physicalcpu",
	); err == nil {
		if value, parseErr := strconv.Atoi(
			strings.TrimSpace(string(out)),
		); parseErr == nil && value > 0 {
			result.PhysicalCores = value
		}
	}

	if out, err := runDarwin(
		ctx,
		runner,
		"/usr/sbin/sysctl",
		"-n",
		"hw.logicalcpu",
	); err == nil {
		if value, parseErr := strconv.Atoi(
			strings.TrimSpace(string(out)),
		); parseErr == nil && value > 0 {
			result.LogicalCores = value
		}
	}

	return result
}

func detectDarwinMemory(
	ctx context.Context,
	runner darwinRunner,
) protocol.MemoryInventory {
	out, err := runDarwin(
		ctx,
		runner,
		"/usr/sbin/sysctl",
		"-n",
		"hw.memsize",
	)

	if err != nil {
		return protocol.MemoryInventory{}
	}

	value, err := strconv.ParseUint(
		strings.TrimSpace(string(out)),
		10,
		64,
	)

	if err != nil {
		return protocol.MemoryInventory{}
	}

	return protocol.MemoryInventory{
		TotalBytes: value,
	}
}

type darwinGPUFields struct {
	model  string
	vendor string
	bus    string
	vram   string
}

func parseDarwinGPUs(raw string) []protocol.GPUInventory {
	var result []protocol.GPUInventory
	var current *darwinGPUFields

	flush := func() {
		if current == nil ||
			strings.TrimSpace(current.model) == "" {
			current = nil
			return
		}

		gpu := protocol.GPUInventory{
			Vendor:     darwinVendor(current.vendor, current.model),
			Model:      strings.TrimSpace(current.model),
			DeviceType: "unknown",
			MemoryType: "unknown",
		}

		switch strings.ToLower(strings.TrimSpace(current.bus)) {
		case "built-in":
			gpu.DeviceType = "integrated"
			gpu.MemoryType = "shared"

		case "pcie", "pci":
			gpu.DeviceType = "discrete"
			gpu.MemoryType = "dedicated"

			if vram, ok := parseDarwinVRAM(current.vram); ok {
				gpu.UsableVRAMBytes = &vram
			}
		}

		result = append(result, gpu)
		current = nil
	}

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "Chipset Model" {
			flush()

			current = &darwinGPUFields{
				model: value,
			}
			continue
		}

		if current == nil {
			continue
		}

		switch {
		case key == "Vendor":
			current.vendor = value

		case key == "Bus":
			current.bus = value

		case strings.HasPrefix(key, "VRAM"):
			current.vram = value
		}
	}

	flush()

	return result
}

func darwinVendor(vendor string, model string) string {
	value := strings.ToLower(vendor + " " + model)

	switch {
	case strings.Contains(value, "intel"):
		return "Intel"

	case strings.Contains(value, "amd"),
		strings.Contains(value, "advanced micro devices"),
		strings.Contains(value, "ati"):
		return "AMD"

	case strings.Contains(value, "nvidia"):
		return "NVIDIA"

	case strings.Contains(value, "apple"):
		return "Apple"
	}

	return "Unknown"
}

var darwinVRAMPattern = regexp.MustCompile(
	`(?i)^\s*([0-9]+(?:\.[0-9]+)?)\s*(KB|MB|GB|TB)\s*$`,
)

func parseDarwinVRAM(value string) (uint64, bool) {
	match := darwinVRAMPattern.FindStringSubmatch(
		strings.TrimSpace(value),
	)

	if len(match) != 3 {
		return 0, false
	}

	number, err := strconv.ParseFloat(match[1], 64)
	if err != nil || number < 0 {
		return 0, false
	}

	multipliers := map[string]float64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	multiplier := multipliers[strings.ToUpper(match[2])]
	if multiplier == 0 {
		return 0, false
	}

	return uint64(number * multiplier), true
}

func detectDarwinAcceleration(
	ctx context.Context,
	runner darwinRunner,
	profiler string,
	lstat func(string) (fs.FileInfo, error),
	stat func(string) (fs.FileInfo, error),
) []protocol.AccelerationRuntime {
	var result []protocol.AccelerationRuntime

	if metal := darwinMetalVersion(profiler); metal != "" {
		result = append(
			result,
			protocol.AccelerationRuntime{
				Name:    "Metal",
				Version: metal,
				Status:  "available",
			},
		)
	}

	for _, candidate := range []string{
		"/usr/local/bin/vulkaninfo",
		"/opt/homebrew/bin/vulkaninfo",
	} {
		path, ok := trustedDarwinExecutable(
			candidate,
			lstat,
			stat,
		)
		if !ok {
			continue
		}

		out, err := runDarwin(
			ctx,
			runner,
			path,
			"--summary",
		)
		if err != nil {
			continue
		}

		version := ""

		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(
				line,
				"Vulkan Instance Version:",
			) {
				version = strings.TrimSpace(
					strings.TrimPrefix(
						line,
						"Vulkan Instance Version:",
					),
				)
				break
			}
		}

		result = append(
			result,
			protocol.AccelerationRuntime{
				Name:    "Vulkan",
				Version: version,
				Status:  "available",
			},
		)

		break
	}

	return result
}

func darwinMetalVersion(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "Metal Support:") {
			continue
		}

		value := strings.TrimSpace(
			strings.TrimPrefix(line, "Metal Support:"),
		)

		if value != "" &&
			!strings.EqualFold(value, "unsupported") {
			return value
		}
	}

	return ""
}

func detectDarwinAIRuntimes(
	ctx context.Context,
	runner darwinRunner,
	home string,
	lstat func(string) (fs.FileInfo, error),
	stat func(string) (fs.FileInfo, error),
) []protocol.AIRuntimeInventory {
	result := []protocol.AIRuntimeInventory{}

	if ollama, ok := firstDarwinExecutable(
		[]string{
			"/usr/local/bin/ollama",
			"/opt/homebrew/bin/ollama",
			"/Applications/Ollama.app/Contents/Resources/ollama",
		},
		lstat,
		stat,
	); ok {
		out, runErr := runDarwin(
			ctx,
			runner,
			ollama,
			"--version",
		)

		status := "available"
		if runErr != nil {
			status = "installed_unavailable"
		}

		result = append(
			result,
			protocol.AIRuntimeInventory{
				Name:    "Ollama",
				Version: darwinTrailingVersion(string(out)),
				Status:  status,
			},
		)
	}

	llamaCandidates := []string{
		filepath.Join(
			home,
			"llama.cpp",
			"build",
			"bin",
			"llama-server",
		),
		filepath.Join(
			home,
			"llama.cpp",
			"build-metal",
			"bin",
			"llama-server",
		),
		"/usr/local/bin/llama-server",
		"/opt/homebrew/bin/llama-server",
	}

	if llama, ok := firstDarwinExecutable(
		llamaCandidates,
		lstat,
		stat,
	); ok {
		out, runErr := runDarwin(
			ctx,
			runner,
			llama,
			"--version",
		)

		status := "available"
		if runErr != nil {
			status = "installed_unavailable"
		}

		result = append(
			result,
			protocol.AIRuntimeInventory{
				Name:    "llama.cpp",
				Version: darwinTrailingVersion(string(out)),
				Status:  status,
			},
		)
	}

	return result
}

func detectDarwinModels(
	ctx context.Context,
	runner darwinRunner,
	home string,
	lstat func(string) (fs.FileInfo, error),
	stat func(string) (fs.FileInfo, error),
) []protocol.ModelInventory {
	models := map[string]protocol.ModelInventory{}

	if ollama, ok := firstDarwinExecutable(
		[]string{
			"/usr/local/bin/ollama",
			"/opt/homebrew/bin/ollama",
			"/Applications/Ollama.app/Contents/Resources/ollama",
		},
		lstat,
		stat,
	); ok {
		if out, err := runDarwin(
			ctx,
			runner,
			ollama,
			"list",
		); err == nil {
			for _, model := range parseDarwinOllamaList(
				string(out),
			) {
				key := "ollama\x00" + model.Name
				models[key] = model

				if len(models) >= darwinModelLimit {
					break
				}
			}
		}
	}

	roots := []string{
		filepath.Join(home, ".cache", "huggingface", "hub"),
		filepath.Join(home, ".cache", "llama.cpp"),
		filepath.Join(home, "models"),
	}

rootLoop:
	for _, root := range roots {
		for _, model := range discoverDarwinGGUF(root) {
			key := "gguf\x00" +
				strings.ToLower(model.Name) +
				fmt.Sprintf("\x00%d", model.SizeBytes)

			models[key] = model

			if len(models) >= darwinModelLimit {
				break rootLoop
			}
		}
	}

	result := make(
		[]protocol.ModelInventory,
		0,
		len(models),
	)

	for _, model := range models {
		result = append(result, model)
	}

	sort.Slice(
		result,
		func(i, j int) bool {
			if result[i].Runtime != result[j].Runtime {
				return result[i].Runtime < result[j].Runtime
			}

			return result[i].Name < result[j].Name
		},
	)

	if len(result) > darwinModelLimit {
		result = result[:darwinModelLimit]
	}

	return result
}

func trustedDarwinExecutable(
	path string,
	lstat func(string) (fs.FileInfo, error),
	stat func(string) (fs.FileInfo, error),
) (string, bool) {
	info, err := lstat(path)
	if err != nil {
		return "", false
	}

	resolved := path

	if info.Mode()&os.ModeSymlink != 0 {
		value, err := filepath.EvalSymlinks(path)
		if err != nil {
			return "", false
		}

		resolved = value
	}

	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", false
	}

	targetInfo, err := stat(resolved)
	if err != nil ||
		!targetInfo.Mode().IsRegular() ||
		targetInfo.Mode().Perm()&0111 == 0 {
		return "", false
	}

	if !darwinExecutableAllowed(path, resolved) {
		return "", false
	}

	return resolved, true
}

func firstDarwinExecutable(
	candidates []string,
	lstat func(string) (fs.FileInfo, error),
	stat func(string) (fs.FileInfo, error),
) (string, bool) {
	for _, candidate := range candidates {
		if path, ok := trustedDarwinExecutable(
			candidate,
			lstat,
			stat,
		); ok {
			return path, true
		}
	}

	return "", false
}

func darwinExecutableAllowed(
	candidate string,
	resolved string,
) bool {
	candidate, _ = filepath.Abs(candidate)
	resolved, _ = filepath.Abs(resolved)

	allowedRoots := []string{
		"/usr/local",
		"/opt/homebrew",
		"/Applications/Ollama.app",
	}

	if home, err := os.UserHomeDir(); err == nil {
		allowedRoots = append(
			allowedRoots,
			filepath.Join(home, "llama.cpp"),
		)
	}

	for _, root := range allowedRoots {
		root, err := filepath.Abs(root)
		if err != nil {
			continue
		}

		if darwinPathInside(root, candidate) &&
			darwinPathInside(root, resolved) {
			return true
		}
	}

	return false
}

func darwinPathInside(root string, target string) bool {
	rel, err := filepath.Rel(root, target)

	if err != nil ||
		filepath.IsAbs(rel) ||
		rel == ".." ||
		strings.HasPrefix(
			rel,
			".."+string(filepath.Separator),
		) {
		return false
	}

	return true
}

func parseDarwinOllamaList(
	raw string,
) []protocol.ModelInventory {
	var result []protocol.ModelInventory

	for i, line := range strings.Split(raw, "\n") {
		if i == 0 &&
			strings.HasPrefix(
				strings.TrimSpace(line),
				"NAME",
			) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		model := protocol.ModelInventory{
			Name:    fields[0],
			Runtime: "Ollama",
			Format:  "ollama",
		}

		if len(fields) >= 4 {
			if size, ok := darwinHumanBytes(
				fields[2],
				fields[3],
			); ok {
				model.SizeBytes = size
			}
		}

		result = append(result, model)

		if len(result) >= darwinModelLimit {
			break
		}
	}

	return result
}

func darwinHumanBytes(
	value string,
	unit string,
) (uint64, bool) {
	number, err := strconv.ParseFloat(value, 64)

	if err != nil || number < 0 {
		return 0, false
	}

	multipliers := map[string]float64{
		"KB": 1e3,
		"MB": 1e6,
		"GB": 1e9,
		"TB": 1e12,
	}

	multiplier, ok := multipliers[strings.ToUpper(unit)]
	if !ok {
		return 0, false
	}

	return uint64(number * multiplier), true
}

var darwinMultipartGGUF = regexp.MustCompile(
	`(?i)^(.*)-([0-9]{5})-of-([0-9]{5})\.gguf$`,
)

var darwinQuantization = regexp.MustCompile(
	`(?i)(Q[0-9]+(?:_[A-Z0-9]+)+|IQ[0-9]+_[A-Z0-9]+|UD-IQ[0-9]+_[A-Z0-9]+)`,
)

func discoverDarwinGGUF(
	root string,
) []protocol.ModelInventory {
	rootInfo, err := os.Lstat(root)

	if err != nil ||
		!rootInfo.IsDir() ||
		rootInfo.Mode()&os.ModeSymlink != 0 {
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

	_ = filepath.WalkDir(
		root,
		func(
			path string,
			d fs.DirEntry,
			err error,
		) error {
			if err != nil {
				return nil
			}

			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}

			depth := 0

			if rel != "." {
				depth = len(
					strings.Split(
						rel,
						string(filepath.Separator),
					),
				)
			}

			if d.IsDir() &&
				depth > darwinModelMaxDepth {
				return filepath.SkipDir
			}

			if d.IsDir() ||
				depth > darwinModelMaxDepth {
				return nil
			}

			if !strings.EqualFold(
				filepath.Ext(d.Name()),
				".gguf",
			) {
				return nil
			}

			if strings.HasPrefix(
				strings.ToLower(d.Name()),
				"mmproj-",
			) {
				return nil
			}

			var size int64

			if d.Type()&os.ModeSymlink != 0 {
				resolved, err := filepath.EvalSymlinks(path)
				if err != nil {
					return nil
				}

				resolved, err = filepath.Abs(resolved)

				if err != nil ||
					!darwinPathInside(
						resolvedRoot,
						resolved,
					) {
					return nil
				}

				info, err := os.Stat(resolved)

				if err != nil ||
					!info.Mode().IsRegular() ||
					info.Size() < 0 {
					return nil
				}

				size = info.Size()
			} else {
				info, err := d.Info()

				if err != nil ||
					!info.Mode().IsRegular() ||
					info.Size() < 0 {
					return nil
				}

				size = info.Size()
			}

			name := d.Name()
			key := strings.ToLower(name)

			if match := darwinMultipartGGUF.
				FindStringSubmatch(name); len(match) == 4 {
				name = match[1] + ".gguf"
				key = strings.ToLower(name)
			}

			entry := groups[key]

			if entry == nil {
				if len(groups) >= darwinModelLimit {
					return nil
				}

				entry = &aggregate{
					name: name,
				}
				groups[key] = entry
			}

			entry.size += uint64(size)

			return nil
		},
	)

	result := make(
		[]protocol.ModelInventory,
		0,
		len(groups),
	)

	for _, group := range groups {
		base := strings.TrimSuffix(
			group.name,
			filepath.Ext(group.name),
		)

		quant := ""

		if match := darwinQuantization.
			FindStringSubmatch(base); len(match) >= 2 {
			quant = strings.ToUpper(match[1])
		}

		result = append(
			result,
			protocol.ModelInventory{
				Name:         base,
				Runtime:      "llama.cpp",
				Format:       "GGUF",
				Quantization: quant,
				SizeBytes:    group.size,
			},
		)
	}

	sort.Slice(
		result,
		func(i, j int) bool {
			return result[i].Name < result[j].Name
		},
	)

	if len(result) > darwinModelLimit {
		result = result[:darwinModelLimit]
	}

	return result
}

var darwinVersion = regexp.MustCompile(
	`(?i)\b(v?[0-9]+(?:\.[0-9]+){1,3}(?:[-+][A-Za-z0-9._-]+)?)\b`,
)

func darwinTrailingVersion(raw string) string {
	matches := darwinVersion.FindAllString(raw, -1)

	if len(matches) == 0 {
		return ""
	}

	return strings.TrimPrefix(
		matches[len(matches)-1],
		"v",
	)
}
