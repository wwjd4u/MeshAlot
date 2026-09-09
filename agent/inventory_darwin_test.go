//go:build darwin

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDarwinGPUs(t *testing.T) {
	raw := `
Graphics/Displays:

    Intel UHD Graphics 630:

      Chipset Model: Intel UHD Graphics 630
      Type: GPU
      Bus: Built-In
      VRAM (Dynamic, Max): 1536 MB
      Vendor: Intel
      Metal Support: Metal 3

    AMD Radeon Pro 5500M:

      Chipset Model: AMD Radeon Pro 5500M
      Type: GPU
      Bus: PCIe
      VRAM (Total): 8 GB
      Vendor: AMD (0x1002)
      Metal Support: Metal 3
`

	gpus := parseDarwinGPUs(raw)

	if len(gpus) != 2 {
		t.Fatalf("expected 2 GPUs, got %d: %#v", len(gpus), gpus)
	}

	intel := gpus[0]

	if intel.Vendor != "Intel" {
		t.Fatalf("unexpected Intel vendor: %#v", intel)
	}

	if intel.Model != "Intel UHD Graphics 630" {
		t.Fatalf("unexpected Intel model: %#v", intel)
	}

	if intel.DeviceType != "integrated" {
		t.Fatalf("Intel GPU should be integrated: %#v", intel)
	}

	if intel.MemoryType != "shared" {
		t.Fatalf("Intel GPU should use shared memory: %#v", intel)
	}

	if intel.UsableVRAMBytes != nil {
		t.Fatalf(
			"integrated GPU must not invent dedicated VRAM: %#v",
			intel,
		)
	}

	amd := gpus[1]

	if amd.Vendor != "AMD" {
		t.Fatalf("unexpected AMD vendor: %#v", amd)
	}

	if amd.Model != "AMD Radeon Pro 5500M" {
		t.Fatalf("unexpected AMD model: %#v", amd)
	}

	if amd.DeviceType != "discrete" {
		t.Fatalf("AMD GPU should be discrete: %#v", amd)
	}

	if amd.MemoryType != "dedicated" {
		t.Fatalf("AMD GPU should use dedicated memory: %#v", amd)
	}

	if amd.UsableVRAMBytes == nil {
		t.Fatalf("AMD dedicated VRAM missing: %#v", amd)
	}

	wantVRAM := uint64(8 * 1024 * 1024 * 1024)

	if *amd.UsableVRAMBytes != wantVRAM {
		t.Fatalf(
			"AMD VRAM got %d want %d",
			*amd.UsableVRAMBytes,
			wantVRAM,
		)
	}
}

func TestDarwinMetalVersion(t *testing.T) {
	raw := `
      Metal Support: Metal 3
      Metal Support: Metal 3
`

	if got := darwinMetalVersion(raw); got != "Metal 3" {
		t.Fatalf("Metal version got %q want %q", got, "Metal 3")
	}
}

func TestParseDarwinVRAM(t *testing.T) {
	got, ok := parseDarwinVRAM("8 GB")

	if !ok {
		t.Fatal("expected 8 GB to parse")
	}

	want := uint64(8 * 1024 * 1024 * 1024)

	if got != want {
		t.Fatalf("VRAM got %d want %d", got, want)
	}
}

func TestDiscoverDarwinGGUFAggregatesMultipart(t *testing.T) {
	root := t.TempDir()

	files := map[string]int{
		"Example-Q4_K_M-00001-of-00002.gguf": 100,
		"Example-Q4_K_M-00002-of-00002.gguf": 200,
	}

	for name, size := range files {
		err := os.WriteFile(
			filepath.Join(root, name),
			make([]byte, size),
			0600,
		)

		if err != nil {
			t.Fatal(err)
		}
	}

	models := discoverDarwinGGUF(root)

	if len(models) != 1 {
		t.Fatalf(
			"expected 1 model, got %d: %#v",
			len(models),
			models,
		)
	}

	model := models[0]

	if model.Name != "Example-Q4_K_M" {
		t.Fatalf("unexpected model name: %#v", model)
	}

	if model.Quantization != "Q4_K_M" {
		t.Fatalf("unexpected quantization: %#v", model)
	}

	if model.SizeBytes != 300 {
		t.Fatalf("size got %d want 300", model.SizeBytes)
	}
}

func TestDiscoverDarwinGGUFRejectsOutsideSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	target := filepath.Join(
		outside,
		"Outside-Q4_K_M.gguf",
	)

	if err := os.WriteFile(
		target,
		make([]byte, 4096),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(
		root,
		"Outside-Q4_K_M.gguf",
	)

	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	models := discoverDarwinGGUF(root)

	if len(models) != 0 {
		t.Fatalf(
			"outside-root symlink should be ignored: %#v",
			models,
		)
	}
}

func TestDarwinPathInside(t *testing.T) {
	root := "/Users/example/models"

	tests := []struct {
		target string
		want   bool
	}{
		{"/Users/example/models/a.gguf", true},
		{"/Users/example/models/sub/a.gguf", true},
		{"/Users/example/other/a.gguf", false},
	}

	for _, test := range tests {
		if got := darwinPathInside(root, test.target); got != test.want {
			t.Fatalf(
				"darwinPathInside(%q, %q)=%v want %v",
				root,
				test.target,
				got,
				test.want,
			)
		}
	}
}

func TestDarwinGlobalModelCap(t *testing.T) {
	root := t.TempDir()

	for i := 0; i < darwinModelLimit+5; i++ {
		name := fmt.Sprintf(
			"Model-%03d-Q4_K_M.gguf",
			i,
		)

		if err := os.WriteFile(
			filepath.Join(root, name),
			[]byte{0},
			0600,
		); err != nil {
			t.Fatal(err)
		}
	}

	models := discoverDarwinGGUF(root)

	if len(models) != darwinModelLimit {
		t.Fatalf(
			"model count got %d want exactly %d",
			len(models),
			darwinModelLimit,
		)
	}
}

func TestParseDarwinOllamaList(t *testing.T) {
	raw := `NAME              ID              SIZE      MODIFIED
qwen3:latest      abc123          18 GB     2 days ago
`

	models := parseDarwinOllamaList(raw)

	if len(models) != 1 {
		t.Fatalf("expected 1 Ollama model: %#v", models)
	}

	if models[0].Name != "qwen3:latest" {
		t.Fatalf("unexpected model: %#v", models[0])
	}

	if models[0].Runtime != "Ollama" {
		t.Fatalf("unexpected runtime: %#v", models[0])
	}

	if models[0].Format != "ollama" {
		t.Fatalf("unexpected format: %#v", models[0])
	}

	if models[0].SizeBytes != 18_000_000_000 {
		t.Fatalf(
			"size got %d want 18000000000",
			models[0].SizeBytes,
		)
	}
}
