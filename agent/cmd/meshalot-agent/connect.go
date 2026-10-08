package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/wwjd4u/MeshAlot/agent"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// m13ConnectSession keeps transport injectable for tests. The command itself
// never enrolls nodes, modifies identities, or starts background services.
type m13ConnectSession func(context.Context, string, agent.Identity, agent.M13HeartbeatSampler, time.Duration) error

func runConnect(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runConnectWithDeps(ctx, args, agent.LoadIdentity, agent.SampleM13Heartbeat,
		func(ctx context.Context, serverURL string, identity agent.Identity,
			sample agent.M13HeartbeatSampler, interval time.Duration) error {
			return agent.RunM13Session(ctx, serverURL, identity, sample, interval, nil)
		})
}

// Gate 7 uses explicitly supplied current provider state. We cannot assume
// the mode or pause state from prior benchmark data, and unknown job activity
// is reported as "unknown", not falsely "idle". A later gate must bind
// values to the live local provider-control store before automatic startup.
func runConnectWithDeps(ctx context.Context, args []string,
	load func(string) (agent.Identity, error),
	collect func(context.Context, agent.M13SamplerOptions) (protocol.M13Heartbeat, error),
	session m13ConnectSession) error {
	flags := flag.NewFlagSet("connect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	server := flags.String("server", env("MESHALOT_SERVER_URL", "https://api.meshalot.com"),
		"HTTPS control-plane URL")
	identityPath := flags.String("identity", "", "existing identity file; never creates identity")
	mode := flags.String("mode", "", "explicit provider mode: normal, away, maximum-earnings")
	pause := flags.String("manual-pause", "", "explicit true or false pause state")
	jobState := flags.String("job-state", "unknown", "idle, running, stopping, or unknown")
	interval := flags.Duration("interval", agent.M13DefaultHeartbeatInterval, "heartbeat interval")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("invalid connect options: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("connect accepts no positional arguments or enrollment codes")
	}
	switch *mode {
	case "normal", "away", "maximum-earnings":
	default:
		return errors.New("--mode must explicitly be normal, away, or maximum-earnings")
	}
	if *pause != "true" && *pause != "false" {
		return errors.New("--manual-pause must explicitly be true or false")
	}
	switch *jobState {
	case "idle", "running", "stopping", "unknown":
	default:
		return errors.New("--job-state must be idle, running, stopping, or unknown")
	}
	if *interval < 10*time.Second || *interval > 5*time.Minute {
		return errors.New("--interval must be between 10s and 5m")
	}
	if _, err := agent.M13WebSocketURL(strings.TrimSpace(*server)); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if load == nil || collect == nil || session == nil {
		return errors.New("connect dependencies are unavailable")
	}
	path, err := resolveIdentityPath(*identityPath)
	if err != nil {
		return err
	}
	// LoadIdentity is read-only. In particular DO NOT use LoadOrCreateIdentity,
	// generate a key, or send an enrollment request here.
	identity, err := load(path)
	if err != nil {
		return fmt.Errorf("read existing enrolled identity: %w", err)
	}
	options := agent.M13SamplerOptions{
		AvailabilityMode: *mode,
		ManualPause:      *pause == "true",
		ActiveJobState:   *jobState,
		RecentLatencyMS:  0, // replaced with measured control RTT in RunM13Session
	}
	sample := func(ctx context.Context) (protocol.M13Heartbeat, error) {
		return collect(ctx, options)
	}
	// Exactly one session. Automatic reconnection/backoff belongs to a later gate.
	return session(ctx, strings.TrimSpace(*server), identity, sample, *interval)
}
