package agent

import (
	"strings"
	"testing"
)

const validM9LlamaBenchJSON = `[
  {
    "build_commit": "8cf427ff",
    "build_number": 5163,
    "backends": "Metal",
    "model_filename": "/models/m9-standard.gguf",
    "model_type": "test 7B Q4_K - Medium",
    "n_prompt": 512,
    "n_gen": 0,
    "avg_ts": 7263.609157,
    "samples_ts": [7155.74]
  },
  {
    "build_commit": "8cf427ff",
    "build_number": 5163,
    "backends": "Metal",
    "model_filename": "/models/m9-standard.gguf",
    "model_type": "test 7B Q4_K - Medium",
    "n_prompt": 0,
    "n_gen": 128,
    "avg_ts": 118.881588,
    "samples_ts": [119.03]
  }
]`

func TestM9LlamaBenchArgs(t *testing.T) {
	args, err := m9LlamaBenchArgs(
		"/models/m9-standard.gguf",
	)
	if err != nil {
		t.Fatal(err)
	}

	got := strings.Join(args, " ")

	want :=
		"-m /models/m9-standard.gguf " +
			"-p 512 -n 128 -r 1 -o json"

	if got != want {
		t.Fatalf(
			"args = %q, want %q",
			got,
			want,
		)
	}
}

func TestM9LlamaBenchArgsRequiresModel(t *testing.T) {
	if _, err := m9LlamaBenchArgs(" "); err == nil {
		t.Fatal("empty model path was accepted")
	}
}

func TestParseLlamaBenchJSON(t *testing.T) {
	result, err := parseLlamaBenchJSON(
		[]byte(validM9LlamaBenchJSON),
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.RuntimeVersion != "8cf427ff-5163" {
		t.Fatalf(
			"runtime version = %q",
			result.RuntimeVersion,
		)
	}

	if result.Backend != "Metal" {
		t.Fatalf(
			"backend = %q",
			result.Backend,
		)
	}

	if len(result.PromptTokensPerSecond) != 1 ||
		len(result.GenerationTokensPerSecond) != 1 {
		t.Fatal("unexpected repetition count")
	}

	if result.PromptTokensPerSecond[0] != 7155.74 {
		t.Fatal("prompt sample changed")
	}

	if result.GenerationTokensPerSecond[0] != 119.03 {
		t.Fatal("generation sample changed")
	}
}

func TestParseLlamaBenchRejectsMissingGeneration(t *testing.T) {
	raw := `[
	  {
	    "build_commit":"abc",
	    "build_number":1,
	    "backends":"CPU",
	    "model_filename":"model.gguf",
	    "model_type":"test",
	    "n_prompt":512,
	    "n_gen":0,
	    "samples_ts":[1]
	  }
	]`

	if _, err := parseLlamaBenchJSON(
		[]byte(raw),
	); err == nil {
		t.Fatal("missing generation result accepted")
	}
}

func TestParseLlamaBenchRejectsBadSample(t *testing.T) {
	raw := strings.Replace(
		validM9LlamaBenchJSON,
		"119.03",
		"0",
		1,
	)

	if _, err := parseLlamaBenchJSON(
		[]byte(raw),
	); err == nil {
		t.Fatal("invalid throughput sample accepted")
	}
}

func TestParseLlamaBenchRejectsMultipleInvocationRepetitions(t *testing.T) {
	raw := strings.Replace(
		validM9LlamaBenchJSON,
		`"samples_ts": [7155.74]`,
		`"samples_ts": [7155.74,7188.71]`,
		1,
	)

	raw = strings.Replace(
		raw,
		`"samples_ts": [119.03]`,
		`"samples_ts": [119.03,120.178]`,
		1,
	)

	if _, err := parseLlamaBenchJSON(
		[]byte(raw),
	); err == nil {
		t.Fatal(
			"multiple repetitions in one llama-bench process were accepted",
		)
	}
}

func TestM9StandardWorkloadConstants(t *testing.T) {
	if m9StandardWorkloadID != "m9-standard-v1" ||
		m9StandardPromptTokens != 512 ||
		m9StandardGeneratedTokens != 128 ||
		m9StandardRepetitions != 5 ||
		m9StandardContextSize != 4096 ||
		m9LlamaBenchInvocationRepetitions != 1 {
		t.Fatal("M9 standard workload changed")
	}
}
