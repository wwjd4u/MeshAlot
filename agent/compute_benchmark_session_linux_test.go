//go:build linux

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

type fakeLinuxBenchmarkSessionProcess struct {
	wait chan error

	mu sync.Mutex

	signals int
	kills   int

	finished bool
}

func newFakeLinuxBenchmarkSessionProcess() *fakeLinuxBenchmarkSessionProcess {

	return &fakeLinuxBenchmarkSessionProcess{
		wait: make(
			chan error,
			1,
		),
	}
}

func (p *fakeLinuxBenchmarkSessionProcess) Wait() error {
	return <-p.wait
}

func (p *fakeLinuxBenchmarkSessionProcess) Signal(
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

func (p *fakeLinuxBenchmarkSessionProcess) Kill() error {
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

type fakeLinuxBenchmarkSessionStarter struct {
	calls int

	name string
	args []string

	process *fakeLinuxBenchmarkSessionProcess
}

func (s *fakeLinuxBenchmarkSessionStarter) Start(
	_ context.Context,
	name string,
	args ...string,
) (
	linuxBenchmarkServerProcess,
	error,
) {
	s.calls++

	s.name =
		name

	s.args =
		append(
			[]string(nil),
			args...,
		)

	if s.process == nil {
		s.process =
			newFakeLinuxBenchmarkSessionProcess()
	}

	return s.process,
		nil
}

type fakeLinuxBenchmarkSessionHTTPClient struct {
	healthCalls int
	ttftCalls   int
}

func (c *fakeLinuxBenchmarkSessionHTTPClient) Do(
	req *http.Request,
) (
	*http.Response,
	error,
) {
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
				"unexpected benchmark HTTP request",
			)
	}
}

func TestExecuteLinuxComputeBenchmark(
	t *testing.T,
) {
	collectedAt :=
		time.Date(
			2026,
			9,
			21,
			20,
			0,
			0,
			0,
			time.UTC,
		)

	process :=
		newFakeLinuxBenchmarkSessionProcess()

	starter :=
		&fakeLinuxBenchmarkSessionStarter{
			process: process,
		}

	client :=
		&fakeLinuxBenchmarkSessionHTTPClient{}

	runCalls :=
		0

	throttled :=
		false

	collectRun :=
		func(
			_ context.Context,
			_ string,
			modelPath string,
		) (
			linuxLlamaBenchRunResult,
			error,
		) {
			runCalls++

			return linuxLlamaBenchRunResult{
				Throughput: llamaBenchThroughput{
					RuntimeVersion: "4df29be4f-10454",

					ModelFilename: modelPath,

					ModelType: "qwen3moe 30B.A3B Q4_K - Medium",

					Backend: "SYCL",

					PromptTokens: m9StandardPromptTokens,

					GeneratedTokens: m9StandardGeneratedTokens,

					PromptTokensPerSecond: []float64{
						40 +
							float64(
								runCalls,
							),
					},

					GenerationTokensPerSecond: []float64{
						9 +
							float64(
								runCalls,
							)/10,
					},
				},

				PeakSystemRAM: uint64(
					18+
						runCalls,
				) *
					1024 *
					1024 *
					1024,

				Throttled: &throttled,
			}, nil
		}

	nowCalls :=
		0

	now :=
		func() time.Time {
			value :=
				collectedAt.Add(
					time.Duration(
						nowCalls,
					) *
						250 *
						time.Millisecond,
				)

			nowCalls++

			return value
		}

	benchmark,
		err :=
		executeLinuxComputeBenchmark(
			context.Background(),
			linuxComputeBenchmarkExecutionInput{
				CollectedAt: collectedAt,

				Condition: "loaded-interference-validation",

				Quantization: "Q4_K_M",

				LlamaBenchPath: "/opt/llama-bench",

				LlamaServerPath: "/opt/llama-server",

				ModelPath: "/models/m9.gguf",

				ServerStarter: starter,

				HealthClient: client,

				TTFTClient: client,

				Now: now,

				CollectRun: collectRun,
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if starter.calls != 1 {
		t.Fatalf(
			"server starts = %d",
			starter.calls,
		)
	}

	if runCalls !=
		m9StandardRepetitions {

		t.Fatalf(
			"throughput runs = %d",
			runCalls,
		)
	}

	if client.healthCalls != 1 {
		t.Fatalf(
			"health calls = %d",
			client.healthCalls,
		)
	}

	if client.ttftCalls !=
		m9StandardRepetitions {

		t.Fatalf(
			"TTFT calls = %d",
			client.ttftCalls,
		)
	}

	if len(benchmark.Samples) !=
		m9StandardRepetitions {

		t.Fatalf(
			"sample count = %d",
			len(
				benchmark.Samples,
			),
		)
	}

	if benchmark.Condition !=
		"loaded-interference-validation" {

		t.Fatalf(
			"condition = %q",
			benchmark.Condition,
		)
	}

	if benchmark.RuntimeVersion !=
		"4df29be4f-10454" {

		t.Fatalf(
			"runtime version = %q",
			benchmark.RuntimeVersion,
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
				sample.
					TimeToFirstTokenMs,
			)
		}
	}

	process.mu.Lock()
	defer process.mu.Unlock()

	if process.signals != 1 {
		t.Fatalf(
			"server signals = %d",
			process.signals,
		)
	}

	if process.kills != 0 {
		t.Fatalf(
			"server kills = %d",
			process.kills,
		)
	}
}

func TestExecuteLinuxComputeBenchmarkStopsServerOnCancellation(
	t *testing.T,
) {
	ctx,
		cancel :=
		context.WithCancel(
			context.Background(),
		)

	process :=
		newFakeLinuxBenchmarkSessionProcess()

	starter :=
		&fakeLinuxBenchmarkSessionStarter{
			process: process,
		}

	client :=
		&fakeLinuxBenchmarkSessionHTTPClient{}

	collectRun :=
		func(
			_ context.Context,
			_ string,
			_ string,
		) (
			linuxLlamaBenchRunResult,
			error,
		) {
			cancel()

			return linuxLlamaBenchRunResult{},
				errors.New(
					"canceled test run",
				)
		}

	_,
		err :=
		executeLinuxComputeBenchmark(
			ctx,
			linuxComputeBenchmarkExecutionInput{
				CollectedAt: time.Now().UTC(),

				Condition: "loaded-interference-validation",

				Quantization: "Q4_K_M",

				LlamaBenchPath: "/opt/llama-bench",

				LlamaServerPath: "/opt/llama-server",

				ModelPath: "/models/m9.gguf",

				ServerStarter: starter,

				HealthClient: client,

				TTFTClient: client,

				CollectRun: collectRun,
			},
		)

	if err == nil {
		t.Fatal(
			"canceled benchmark was accepted",
		)
	}

	process.mu.Lock()
	defer process.mu.Unlock()

	if process.signals != 1 {
		t.Fatalf(
			"server was not stopped after cancellation; signals = %d",
			process.signals,
		)
	}

	if process.kills != 0 {
		t.Fatalf(
			"unexpected server kills = %d",
			process.kills,
		)
	}
}

func TestExecuteLinuxComputeBenchmarkRejectsInvalidInputBeforeStart(
	t *testing.T,
) {
	starter :=
		&fakeLinuxBenchmarkSessionStarter{}

	_,
		err :=
		executeLinuxComputeBenchmark(
			context.Background(),
			linuxComputeBenchmarkExecutionInput{
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
			"invalid execution input was accepted",
		)
	}

	if starter.calls != 0 {
		t.Fatalf(
			"server started before validation; calls = %d",
			starter.calls,
		)
	}
}
