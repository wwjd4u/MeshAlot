//go:build darwin

package agent

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeDarwinM9ServerProcess struct {
	wait chan error

	mu sync.Mutex

	signals int
	kills   int

	finished bool
}

func newFakeDarwinM9ServerProcess() *fakeDarwinM9ServerProcess {
	return &fakeDarwinM9ServerProcess{
		wait: make(
			chan error,
			1,
		),
	}
}

func (p *fakeDarwinM9ServerProcess) Wait() error {
	return <-p.wait
}

func (p *fakeDarwinM9ServerProcess) Signal(
	_ os.Signal,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.signals++

	if !p.finished {
		p.finished = true
		p.wait <- nil
	}

	return nil
}

func (p *fakeDarwinM9ServerProcess) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.kills++

	if !p.finished {
		p.finished = true

		p.wait <- errors.New(
			"killed",
		)
	}

	return nil
}

type fakeDarwinM9ServerStarter struct {
	mu sync.Mutex

	calls int

	names []string
	args  [][]string

	processes []*fakeDarwinM9ServerProcess
}

func (s *fakeDarwinM9ServerStarter) Start(
	_ context.Context,
	name string,
	args ...string,
) (
	darwinBenchmarkServerProcess,
	error,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++

	s.names = append(
		s.names,
		name,
	)

	s.args = append(
		s.args,
		append(
			[]string(nil),
			args...,
		),
	)

	process :=
		newFakeDarwinM9ServerProcess()

	s.processes = append(
		s.processes,
		process,
	)

	return process, nil
}

type fakeDarwinM9HTTPClient struct {
	mu sync.Mutex

	healthCalls int
	ttftCalls   int
}

func (c *fakeDarwinM9HTTPClient) Do(
	req *http.Request,
) (
	*http.Response,
	error,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch {
	case req.Method ==
		http.MethodGet &&
		req.URL.Path ==
			"/health":

		c.healthCalls++

		return &http.Response{
			StatusCode: http.StatusOK,

			Body: io.NopCloser(
				strings.NewReader(
					`{"status":"ok"}`,
				),
			),

			Header: make(http.Header),
		}, nil

	case req.Method ==
		http.MethodPost &&
		req.URL.Path ==
			"/completion":

		c.ttftCalls++

		return &http.Response{
			StatusCode: http.StatusOK,

			Body: io.NopCloser(
				strings.NewReader(
					"data: {\"content\":\"x\",\"tokens\":[1],\"stop\":false}\n\n",
				),
			),

			Header: make(http.Header),
		}, nil

	default:

		return nil,
			errors.New(
				"unexpected Darwin M9 HTTP request",
			)
	}
}

func TestDarwinBenchmarkServerCommandCPUOnly(
	t *testing.T,
) {
	command,
		args,
		baseURL,
		err := darwinBenchmarkServerCommand(
		"/opt/llama-server",
		"/models/m9.gguf",
	)

	if err != nil {
		t.Fatal(err)
	}

	if command !=
		"/opt/llama-server" {

		t.Fatalf(
			"command = %q",
			command,
		)
	}

	if baseURL !=
		"http://127.0.0.1:18190" {

		t.Fatalf(
			"base URL = %q",
			baseURL,
		)
	}

	joined :=
		strings.Join(
			args,
			" ",
		)

	required := []string{
		"--model /models/m9.gguf",
		"--offline",
		"--n-gpu-layers 0",
		"--ctx-size 4096",
		"--parallel 1",
		"--flash-attn off",
		"--host 127.0.0.1",
		"--port 18190",
	}

	for _, value := range required {
		if !strings.Contains(
			joined,
			value,
		) {
			t.Fatalf(
				"server command missing %q: %s",
				value,
				joined,
			)
		}
	}

	if strings.Contains(
		joined,
		"SYCL",
	) {
		t.Fatalf(
			"Darwin server command contains Linux GPU backend: %s",
			joined,
		)
	}
}

func TestExecuteDarwinComputeBenchmarkSequentialTTFTServers(
	t *testing.T,
) {
	collectedAt :=
		time.Date(
			2026,
			9,
			24,
			21,
			0,
			0,
			0,
			time.UTC,
		)

	starter :=
		&fakeDarwinM9ServerStarter{}

	client :=
		&fakeDarwinM9HTTPClient{}

	runCalls := 0

	throttled := false

	clockCalls := 0

	clock := func() time.Time {
		value :=
			collectedAt.Add(
				time.Duration(
					clockCalls,
				) *
					250 *
					time.Millisecond,
			)

		clockCalls++

		return value
	}

	benchmark,
		err := executeDarwinComputeBenchmark(
		context.Background(),
		darwinComputeBenchmarkExecutionInput{
			CollectedAt: collectedAt,

			Condition: "baseline",

			Quantization: "Q4_K_M",

			LlamaBenchPath: "/opt/llama-bench",

			LlamaServerPath: "/opt/llama-server",

			ModelPath: "/models/m9.gguf",

			ServerStarter: starter,

			HealthClient: client,

			TTFTClient: client,

			Now: clock,

			CollectRun: func(
				_ context.Context,
				_ string,
				modelPath string,
			) (
				darwinLlamaBenchRunResult,
				error,
			) {
				runCalls++

				return darwinLlamaBenchRunResult{
					Throughput: llamaBenchThroughput{
						RuntimeVersion: "4df29be4f-10454",

						ModelFilename: modelPath,

						ModelType: "qwen3moe 30B.A3B Q4_K - Medium",

						Backend: "BLAS",

						PromptTokens: m9StandardPromptTokens,

						GeneratedTokens: m9StandardGeneratedTokens,

						PromptTokensPerSecond: []float64{
							40 +
								float64(
									runCalls,
								),
						},

						GenerationTokensPerSecond: []float64{
							5 +
								float64(
									runCalls,
								)/10,
						},
					},

					PeakSystemRAM: uint64(
						30+
							runCalls,
					) *
						1024 *
						1024 *
						1024,

					Throttled: &throttled,
				}, nil
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if runCalls !=
		m9StandardRepetitions {

		t.Fatalf(
			"Darwin throughput runs = %d",
			runCalls,
		)
	}

	if starter.calls !=
		m9StandardRepetitions {

		t.Fatalf(
			"Darwin TTFT server starts = %d",
			starter.calls,
		)
	}

	if client.healthCalls !=
		m9StandardRepetitions {

		t.Fatalf(
			"Darwin health calls = %d",
			client.healthCalls,
		)
	}

	if client.ttftCalls !=
		m9StandardRepetitions {

		t.Fatalf(
			"Darwin TTFT calls = %d",
			client.ttftCalls,
		)
	}

	if len(benchmark.Samples) !=
		m9StandardRepetitions {

		t.Fatalf(
			"Darwin samples = %d",
			len(
				benchmark.Samples,
			),
		)
	}

	for i, sample := range benchmark.Samples {

		if !sample.Success {
			t.Fatalf(
				"run %d failed: %s",
				i+1,
				sample.Error,
			)
		}

		if sample.TimeToFirstTokenMs !=
			250 {

			t.Fatalf(
				"run %d TTFT = %v",
				i+1,
				sample.TimeToFirstTokenMs,
			)
		}

		if sample.PeakVRAMBytes != nil {
			t.Fatalf(
				"run %d invented VRAM",
				i+1,
			)
		}
	}

	for i, process := range starter.processes {

		process.mu.Lock()

		if process.signals != 1 {
			process.mu.Unlock()

			t.Fatalf(
				"server %d stop signals = %d",
				i+1,
				process.signals,
			)
		}

		if process.kills != 0 {
			process.mu.Unlock()

			t.Fatalf(
				"server %d kills = %d",
				i+1,
				process.kills,
			)
		}

		process.mu.Unlock()
	}
}

func TestExecuteDarwinComputeBenchmarkRejectsInvalidInputBeforeStart(
	t *testing.T,
) {
	starter :=
		&fakeDarwinM9ServerStarter{}

	_,
		err := executeDarwinComputeBenchmark(
		context.Background(),
		darwinComputeBenchmarkExecutionInput{
			CollectedAt: time.Now().UTC(),

			Condition: " ",

			Quantization: "Q4_K_M",

			LlamaBenchPath: "/opt/llama-bench",

			LlamaServerPath: "/opt/llama-server",

			ModelPath: "/models/m9.gguf",

			ServerStarter: starter,
		},
	)

	if err == nil {
		t.Fatal(
			"invalid Darwin execution input was accepted",
		)
	}

	if starter.calls != 0 {
		t.Fatalf(
			"Darwin server started before validation; calls = %d",
			starter.calls,
		)
	}
}
