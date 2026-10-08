package agent

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/gorilla/websocket"
)

const (
	M13InitialRetryDelay = 1 * time.Second
	M13MaximumRetryDelay = 60 * time.Second
	M13BackoffResetAfter = 2 * time.Minute
	m13MaximumJitter     = 500 * time.Millisecond
)

// M13ReconnectBackoff returns a bounded exponential delay for consecutive
// failed/outage sessions. failures==1 starts at 1s, capped at 60s.
func M13ReconnectBackoff(failures uint32) time.Duration {
	delay := M13InitialRetryDelay
	for i := uint32(1); i < failures && delay < M13MaximumRetryDelay; i++ {
		if delay >= M13MaximumRetryDelay/2 {
			return M13MaximumRetryDelay
		}
		delay *= 2
	}
	return delay
}

func m13WaitReconnect(ctx context.Context, base time.Duration) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Jitter prevents thousands of recovered sleeping hosts reconnecting
	// to the control plane at the same instant. Never reduce base delay.
	jitter, err := rand.Int(rand.Reader, big.NewInt(int64(m13MaximumJitter/time.Millisecond)+1))
	if err == nil {
		base += time.Duration(jitter.Int64()) * time.Millisecond
	}
	timer := time.NewTimer(base)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type m13ReconnectRun func(context.Context) error
type m13ReconnectWait func(context.Context, time.Duration) error

// runM13ReconnectLoop is injectable for deterministic simulated-outage tests.
// A run returning without cancellation is unexpected even when err==nil;
// retry it rather than silently leaving the provider offline.
func runM13ReconnectLoop(ctx context.Context, run m13ReconnectRun,
	wait m13ReconnectWait, now func() time.Time) error {
	if ctx == nil || run == nil || wait == nil || now == nil {
		return errors.New("M13 retry-loop dependencies are required")
	}
	var failures uint32
	for {
		if ctx.Err() != nil {
			return nil
		}
		started := now()
		_ = run(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if !now().Before(started.Add(M13BackoffResetAfter)) {
			failures = 0
		}
		if failures < 32 {
			failures++
		}
		if err := wait(ctx, M13ReconnectBackoff(failures)); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}

// RunM13PersistentSession keeps the outbound authenticated WSS connection
// available despite network drops and wake-from-sleep interruptions.
// It REUSES the passed existing identity; it never enrolls a new node,
// opens an inbound port, executes a remote command, or modifies sharing policy.
func RunM13PersistentSession(ctx context.Context, controlURL string,
	identity Identity, sample M13HeartbeatSampler,
	interval time.Duration, dialer *websocket.Dialer) error {
	if ctx == nil {
		return errors.New("context required")
	}
	if err := validateIdentity(identity); err != nil {
		return err
	}
	if _, err := M13WebSocketURL(controlURL); err != nil {
		return err
	}
	if sample == nil {
		return errors.New("heartbeat sampler is required")
	}
	if interval < 10*time.Millisecond || interval > 5*time.Minute {
		return errors.New("heartbeat interval outside supported bounds")
	}
	if dialer != nil && dialer.TLSClientConfig != nil && dialer.TLSClientConfig.InsecureSkipVerify {
		return errors.New("TLS certificate verification must not be disabled")
	}
	return runM13ReconnectLoop(ctx,
		func(ctx context.Context) error {
			return RunM13Session(ctx, controlURL, identity, sample, interval, dialer)
		}, m13WaitReconnect, time.Now)
}
