//go:build darwin

package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	m9DarwinBenchmarkServerHost = "127.0.0.1"
	m9DarwinBenchmarkServerPort = 18190

	m9DarwinBenchmarkServerStartupTimeout = 3 * time.Minute
	m9DarwinBenchmarkServerHTTPTimeout    = 2 * time.Second
	m9DarwinBenchmarkServerPollInterval   = 250 * time.Millisecond
	m9DarwinBenchmarkServerStopTimeout    = 10 * time.Second
)

type darwinBenchmarkServerProcess interface {
	Wait() error
	Signal(os.Signal) error
	Kill() error
}

type darwinBenchmarkServerStarter interface {
	Start(
		context.Context,
		string,
		...string,
	) (
		darwinBenchmarkServerProcess,
		error,
	)
}

type darwinBenchmarkServerOSStarter struct{}

type darwinBenchmarkServerOSProcess struct {
	cmd *exec.Cmd
}

func (darwinBenchmarkServerOSStarter) Start(
	ctx context.Context,
	name string,
	args ...string,
) (
	darwinBenchmarkServerProcess,
	error,
) {
	cmd := exec.CommandContext(
		ctx,
		name,
		args...,
	)

	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &darwinBenchmarkServerOSProcess{
		cmd: cmd,
	}, nil
}

func (p *darwinBenchmarkServerOSProcess) Wait() error {
	return p.cmd.Wait()
}

func (p *darwinBenchmarkServerOSProcess) Signal(
	signal os.Signal,
) error {
	return p.cmd.Process.Signal(signal)
}

func (p *darwinBenchmarkServerOSProcess) Kill() error {
	return p.cmd.Process.Kill()
}

type darwinComputeBenchmarkServer struct {
	process darwinBenchmarkServerProcess

	baseURL string

	waitDone chan error
}

// darwinBenchmarkServerCommand defines the benchmark-only CPU/Accelerate
// llama-server used for M9 TTFT measurement.
//
// Darwin intentionally uses zero GPU layers here. The Node002 recovery build
// is CPU + Apple Accelerate with Metal disabled.
//
// The server is loopback-only, offline, single-slot, and uses the exact
// standardized M9 context size.
func darwinBenchmarkServerCommand(
	serverPath string,
	modelPath string,
) (
	string,
	[]string,
	string,
	error,
) {
	serverPath = strings.TrimSpace(serverPath)

	if serverPath == "" {
		return "", nil, "", errors.New(
			"llama-server path is required",
		)
	}

	modelPath = strings.TrimSpace(modelPath)

	if modelPath == "" {
		return "", nil, "", errors.New(
			"benchmark model path is required",
		)
	}

	args := []string{
		"--model",
		modelPath,

		"--offline",

		"--n-gpu-layers",
		"0",

		"--ctx-size",
		fmt.Sprintf(
			"%d",
			m9StandardContextSize,
		),

		"--parallel",
		"1",

		"--flash-attn",
		"off",

		"--host",
		m9DarwinBenchmarkServerHost,

		"--port",
		fmt.Sprintf(
			"%d",
			m9DarwinBenchmarkServerPort,
		),
	}

	baseURL := fmt.Sprintf(
		"http://%s:%d",
		m9DarwinBenchmarkServerHost,
		m9DarwinBenchmarkServerPort,
	)

	return serverPath,
		args,
		baseURL,
		nil
}

func startDarwinComputeBenchmarkServer(
	ctx context.Context,
	starter darwinBenchmarkServerStarter,
	client computeBenchmarkHTTPDoer,
	serverPath string,
	modelPath string,
) (
	*darwinComputeBenchmarkServer,
	error,
) {
	if ctx == nil {
		return nil, errors.New(
			"benchmark server context is nil",
		)
	}

	if starter == nil {
		starter = darwinBenchmarkServerOSStarter{}
	}

	if client == nil {
		client = &http.Client{
			Timeout: m9DarwinBenchmarkServerHTTPTimeout,
		}
	}

	command,
		args,
		baseURL,
		err := darwinBenchmarkServerCommand(
		serverPath,
		modelPath,
	)

	if err != nil {
		return nil, err
	}

	process,
		err := starter.Start(
		ctx,
		command,
		args...,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"start Darwin M9 llama-server: %w",
			err,
		)
	}

	server := &darwinComputeBenchmarkServer{
		process: process,
		baseURL: baseURL,
		waitDone: make(
			chan error,
			1,
		),
	}

	go func() {
		server.waitDone <- process.Wait()
	}()

	startupCtx,
		cancel := context.WithTimeout(
		ctx,
		m9DarwinBenchmarkServerStartupTimeout,
	)

	defer cancel()

	if err := waitForDarwinBenchmarkServerHealth(
		startupCtx,
		client,
		baseURL,
		server.waitDone,
	); err != nil {

		stopCtx,
			stopCancel := context.WithTimeout(
			context.Background(),
			m9DarwinBenchmarkServerStopTimeout,
		)

		_ = server.Stop(stopCtx)

		stopCancel()

		return nil, err
	}

	return server, nil
}

func waitForDarwinBenchmarkServerHealth(
	ctx context.Context,
	client computeBenchmarkHTTPDoer,
	baseURL string,
	processDone <-chan error,
) error {
	if ctx == nil {
		return errors.New(
			"benchmark health context is nil",
		)
	}

	if client == nil {
		return errors.New(
			"benchmark health HTTP client is nil",
		)
	}

	baseURL,
		err := validateLocalBenchmarkBaseURL(
		baseURL,
	)

	if err != nil {
		return err
	}

	for {
		req,
			err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			baseURL+"/health",
			nil,
		)

		if err != nil {
			return err
		}

		resp,
			requestErr := client.Do(req)

		if requestErr == nil {
			_,
				_ = io.Copy(
				io.Discard,
				io.LimitReader(
					resp.Body,
					64<<10,
				),
			)

			_ = resp.Body.Close()

			if resp.StatusCode ==
				http.StatusOK {

				return nil
			}
		}

		select {
		case processErr := <-processDone:

			if processErr == nil {
				return errors.New(
					"Darwin M9 llama-server exited before becoming healthy",
				)
			}

			return fmt.Errorf(
				"Darwin M9 llama-server exited before becoming healthy: %w",
				processErr,
			)

		case <-ctx.Done():

			return fmt.Errorf(
				"wait for Darwin M9 llama-server health: %w",
				ctx.Err(),
			)

		case <-time.After(
			m9DarwinBenchmarkServerPollInterval,
		):
		}
	}
}

// Stop signals only the process handle created by this benchmark server.
// It never scans for or signals unrelated llama.cpp processes.
func (s *darwinComputeBenchmarkServer) Stop(
	ctx context.Context,
) error {
	if s == nil ||
		s.process == nil {

		return nil
	}

	if ctx == nil {
		return errors.New(
			"benchmark server stop context is nil",
		)
	}

	_ = s.process.Signal(
		os.Interrupt,
	)

	select {
	case <-s.waitDone:
		return nil

	case <-ctx.Done():

		_ = s.process.Kill()

		return ctx.Err()
	}
}
