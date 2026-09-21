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

type fakeLinuxBenchmarkServerProcess struct {
	wait chan error

	mu sync.Mutex

	signals int
	kills   int

	finished bool
}

func newFakeLinuxBenchmarkServerProcess() *fakeLinuxBenchmarkServerProcess {

	return &fakeLinuxBenchmarkServerProcess{
		wait: make(
			chan error,
			1,
		),
	}
}

func (p *fakeLinuxBenchmarkServerProcess) Wait() error {
	return <-p.wait
}

func (p *fakeLinuxBenchmarkServerProcess) Signal(
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

func (p *fakeLinuxBenchmarkServerProcess) Kill() error {
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

type fakeLinuxBenchmarkServerStarter struct {
	calls int

	name string
	args []string

	process *fakeLinuxBenchmarkServerProcess
}

func (s *fakeLinuxBenchmarkServerStarter) Start(
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
			newFakeLinuxBenchmarkServerProcess()
	}

	return s.process,
		nil
}

type fakeLinuxBenchmarkHealthClient struct {
	calls int

	statuses []int

	requestURL string
}

func (c *fakeLinuxBenchmarkHealthClient) Do(
	req *http.Request,
) (
	*http.Response,
	error,
) {
	c.calls++

	c.requestURL =
		req.URL.String()

	status :=
		http.StatusOK

	if len(c.statuses) > 0 {
		index :=
			c.calls - 1

		if index >=
			len(c.statuses) {

			index =
				len(c.statuses) - 1
		}

		status =
			c.statuses[index]
	}

	return &http.Response{
		StatusCode: status,

		Body: io.NopCloser(
			strings.NewReader(
				`{"status":"ok"}`,
			),
		),

		Header: make(http.Header),
	}, nil
}

func TestLinuxBenchmarkServerCommand(
	t *testing.T,
) {
	command,
		args,
		baseURL,
		err :=
		linuxBenchmarkServerCommand(
			"/opt/llama-server",
			"/models/m9.gguf",
		)

	if err != nil {
		t.Fatal(err)
	}

	if command !=
		m9LinuxBashPath {

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

	required :=
		[]string{
			m9LinuxOneAPISetvarsPath,
			"/opt/llama-server",
			"--model /models/m9.gguf",
			"--offline",
			"--device SYCL0",
			"--n-gpu-layers 999",
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
}

func TestWaitForLinuxBenchmarkServerHealth(
	t *testing.T,
) {
	client :=
		&fakeLinuxBenchmarkHealthClient{
			statuses: []int{
				http.StatusServiceUnavailable,
				http.StatusOK,
			},
		}

	ctx,
		cancel :=
		context.WithTimeout(
			context.Background(),
			2*time.Second,
		)

	defer cancel()

	err :=
		waitForLinuxBenchmarkServerHealth(
			ctx,
			client,
			"http://127.0.0.1:18190",
			nil,
		)

	if err != nil {
		t.Fatal(err)
	}

	if client.calls != 2 {
		t.Fatalf(
			"health calls = %d",
			client.calls,
		)
	}

	if client.requestURL !=
		"http://127.0.0.1:18190/health" {

		t.Fatalf(
			"health URL = %q",
			client.requestURL,
		)
	}
}

func TestStartAndStopLinuxComputeBenchmarkServer(
	t *testing.T,
) {
	process :=
		newFakeLinuxBenchmarkServerProcess()

	starter :=
		&fakeLinuxBenchmarkServerStarter{
			process: process,
		}

	client :=
		&fakeLinuxBenchmarkHealthClient{
			statuses: []int{
				http.StatusOK,
			},
		}

	ctx,
		cancel :=
		context.WithCancel(
			context.Background(),
		)

	defer cancel()

	server, err :=
		startLinuxComputeBenchmarkServer(
			ctx,
			starter,
			client,
			"/opt/llama-server",
			"/models/m9.gguf",
		)

	if err != nil {
		t.Fatal(err)
	}

	if starter.calls != 1 {
		t.Fatalf(
			"starter calls = %d",
			starter.calls,
		)
	}

	if server.baseURL !=
		"http://127.0.0.1:18190" {

		t.Fatalf(
			"base URL = %q",
			server.baseURL,
		)
	}

	stopCtx,
		stopCancel :=
		context.WithTimeout(
			context.Background(),
			time.Second,
		)

	defer stopCancel()

	if err :=
		server.Stop(
			stopCtx,
		); err != nil {

		t.Fatal(err)
	}

	process.mu.Lock()
	defer process.mu.Unlock()

	if process.signals != 1 {
		t.Fatalf(
			"signals = %d",
			process.signals,
		)
	}

	if process.kills != 0 {
		t.Fatalf(
			"kills = %d",
			process.kills,
		)
	}
}

func TestLinuxBenchmarkServerCommandRejectsEmptyModel(
	t *testing.T,
) {
	if _, _, _, err :=
		linuxBenchmarkServerCommand(
			"/opt/llama-server",
			" ",
		); err == nil {

		t.Fatal(
			"empty model path was accepted",
		)
	}
}
