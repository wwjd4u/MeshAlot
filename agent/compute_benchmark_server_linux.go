//go:build linux

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
	m9LinuxBenchmarkServerHost = "127.0.0.1"

	m9LinuxBenchmarkServerPort = 18190

	m9LinuxBenchmarkServerStartupTimeout = 3 * time.Minute

	m9LinuxBenchmarkServerHTTPTimeout = 2 * time.Second

	m9LinuxBenchmarkServerPollInterval = 250 * time.Millisecond

	m9LinuxBenchmarkServerStopTimeout = 10 * time.Second
)

type linuxBenchmarkServerProcess interface {
	Wait() error
	Signal(os.Signal) error
	Kill() error
}

type linuxBenchmarkServerStarter interface {
	Start(
		context.Context,
		string,
		...string,
	) (
		linuxBenchmarkServerProcess,
		error,
	)
}

type linuxBenchmarkServerOSStarter struct{}

type linuxBenchmarkServerOSProcess struct {
	cmd *exec.Cmd
}

func (linuxBenchmarkServerOSStarter) Start(
	ctx context.Context,
	name string,
	args ...string,
) (
	linuxBenchmarkServerProcess,
	error,
) {
	cmd :=
		exec.CommandContext(
			ctx,
			name,
			args...,
		)

	cmd.Stdout =
		io.Discard

	cmd.Stderr =
		io.Discard

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &linuxBenchmarkServerOSProcess{
		cmd: cmd,
	}, nil
}

func (p *linuxBenchmarkServerOSProcess) Wait() error {
	return p.cmd.Wait()
}

func (p *linuxBenchmarkServerOSProcess) Signal(
	signal os.Signal,
) error {
	return p.cmd.Process.Signal(
		signal,
	)
}

func (p *linuxBenchmarkServerOSProcess) Kill() error {
	return p.cmd.Process.Kill()
}

type linuxComputeBenchmarkServer struct {
	process linuxBenchmarkServerProcess

	baseURL string

	waitDone chan error
}

// linuxBenchmarkServerCommand defines the benchmark-only M9 llama-server.
//
// Important:
//   - loopback only
//   - dedicated M9 port
//   - exact M9 context size
//   - SYCL0 only
//   - local model path only
//   - offline mode prevents model downloads
func linuxBenchmarkServerCommand(
	serverPath string,
	modelPath string,
) (
	string,
	[]string,
	string,
	error,
) {
	serverPath =
		strings.TrimSpace(
			serverPath,
		)

	if serverPath == "" {
		return "", nil, "", errors.New(
			"llama-server path is required",
		)
	}

	modelPath =
		strings.TrimSpace(
			modelPath,
		)

	if modelPath == "" {
		return "", nil, "", errors.New(
			"benchmark model path is required",
		)
	}

	serverArgs :=
		[]string{
			"--model",
			modelPath,

			"--offline",

			"--device",
			"SYCL0",

			"--n-gpu-layers",
			"999",

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
			m9LinuxBenchmarkServerHost,

			"--port",
			fmt.Sprintf(
				"%d",
				m9LinuxBenchmarkServerPort,
			),
		}

	commandName,
		commandArgs,
		err :=
		linuxOneAPIWrappedCommand(
			serverPath,
			serverArgs,
		)

	if err != nil {
		return "", nil, "", err
	}

	baseURL :=
		fmt.Sprintf(
			"http://%s:%d",
			m9LinuxBenchmarkServerHost,
			m9LinuxBenchmarkServerPort,
		)

	return commandName,
		commandArgs,
		baseURL,
		nil
}

func startLinuxComputeBenchmarkServer(
	ctx context.Context,
	starter linuxBenchmarkServerStarter,
	client computeBenchmarkHTTPDoer,
	serverPath string,
	modelPath string,
) (
	*linuxComputeBenchmarkServer,
	error,
) {
	if ctx == nil {
		return nil, errors.New(
			"benchmark server context is nil",
		)
	}

	if starter == nil {
		starter =
			linuxBenchmarkServerOSStarter{}
	}

	if client == nil {
		client =
			&http.Client{
				Timeout: m9LinuxBenchmarkServerHTTPTimeout,
			}
	}

	commandName,
		commandArgs,
		baseURL,
		err :=
		linuxBenchmarkServerCommand(
			serverPath,
			modelPath,
		)

	if err != nil {
		return nil, err
	}

	process, err :=
		starter.Start(
			ctx,
			commandName,
			commandArgs...,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"start M9 llama-server: %w",
			err,
		)
	}

	server :=
		&linuxComputeBenchmarkServer{
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
		cancel :=
		context.WithTimeout(
			ctx,
			m9LinuxBenchmarkServerStartupTimeout,
		)

	defer cancel()

	if err :=
		waitForLinuxBenchmarkServerHealth(
			startupCtx,
			client,
			baseURL,
			server.waitDone,
		); err != nil {

		stopCtx,
			stopCancel :=
			context.WithTimeout(
				context.Background(),
				m9LinuxBenchmarkServerStopTimeout,
			)

		_ = server.Stop(stopCtx)

		stopCancel()

		return nil, err
	}

	return server, nil
}

func waitForLinuxBenchmarkServerHealth(
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

	baseURL, err :=
		validateLocalBenchmarkBaseURL(
			baseURL,
		)

	if err != nil {
		return err
	}

	for {
		req, err :=
			http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				baseURL+
					"/health",
				nil,
			)

		if err != nil {
			return err
		}

		resp, requestErr :=
			client.Do(req)

		if requestErr == nil {
			_,
				_ =
				io.Copy(
					io.Discard,
					io.LimitReader(
						resp.Body,
						64<<10,
					),
				)

			_ =
				resp.Body.Close()

			if resp.StatusCode ==
				http.StatusOK {

				return nil
			}
		}

		select {
		case processErr :=
			<-processDone:

			if processErr == nil {
				return errors.New(
					"M9 llama-server exited before becoming healthy",
				)
			}

			return fmt.Errorf(
				"M9 llama-server exited before becoming healthy: %w",
				processErr,
			)

		case <-ctx.Done():

			return fmt.Errorf(
				"wait for M9 llama-server health: %w",
				ctx.Err(),
			)

		case <-time.After(
			m9LinuxBenchmarkServerPollInterval,
		):
		}
	}
}

// Stop only signals the process handle returned by this M9 server instance.
// It never searches for or signals other llama-server processes.
func (s *linuxComputeBenchmarkServer) Stop(
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

	_ =
		s.process.Signal(
			os.Interrupt,
		)

	select {
	case <-s.waitDone:
		return nil

	case <-ctx.Done():

		_ =
			s.process.Kill()

		return ctx.Err()
	}
}
