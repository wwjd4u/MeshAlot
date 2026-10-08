package agent

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	M13DefaultHeartbeatInterval = 30 * time.Second
	m13ClientAuthTimeout        = 10 * time.Second
	m13ClientAckTimeout         = 10 * time.Second
	m13ClientWriteTimeout       = 5 * time.Second
)

type M13HeartbeatSampler func(context.Context) (protocol.M13Heartbeat, error)

type m13AgentChallenge struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
}
type m13AgentProof struct {
	Type      string `json:"type"`
	NodeID    string `json:"node_id"`
	Signature string `json:"signature"`
}
type m13AgentAuthResponse struct {
	Type   string `json:"type"`
	NodeID string `json:"node_id"`
}
type m13AgentHeartbeatAck struct {
	Type string `json:"type"`
}

// M13WebSocketURL builds a strictly authenticated outbound TLS endpoint.
// No userinfo, query, fragments, arbitrary path, or insecure ws:// allowed.
func M13WebSocketURL(base string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u == nil || !strings.EqualFold(u.Scheme, "https") ||
		u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		u.Opaque != "" || (u.Path != "" && u.Path != "/") {
		return "", errors.New("M13 requires an HTTPS control-plane base URL")
	}
	u.Scheme = "wss"
	u.Path = "/v1/agent/connect"
	u.RawPath = ""
	return u.String(), nil
}

// RunM13Session establishes ONE outbound WSS session using an already enrolled
// Ed25519 identity. It sends only sampler-produced validated heartbeats; the
// server never receives the private key or enrollment code. Caller controls
// reconnection in a subsequent M13 gate.
func RunM13Session(ctx context.Context, controlURL string, identity Identity,
	sample M13HeartbeatSampler, interval time.Duration, dialer *websocket.Dialer) error {
	if err := validateIdentity(identity); err != nil {
		return err
	}
	if sample == nil {
		return errors.New("heartbeat sampler is required")
	}
	if interval < 10*time.Millisecond || interval > 5*time.Minute {
		return errors.New("heartbeat interval outside supported bounds")
	}
	endpoint, err := M13WebSocketURL(controlURL)
	if err != nil {
		return err
	}
	privateBytes, err := base64.RawStdEncoding.DecodeString(identity.PrivateKey)
	if err != nil || len(privateBytes) != ed25519.PrivateKeySize {
		return errors.New("private key is invalid")
	}
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	if dialer.TLSClientConfig != nil && dialer.TLSClientConfig.InsecureSkipVerify {
		return errors.New("TLS certificate verification must not be disabled")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	ws, response, err := dialer.DialContext(ctx, endpoint, nil)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return fmt.Errorf("secure agent connection failed: %w", err)
	}
	defer ws.Close()
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			_ = ws.Close()
		case <-stop:
		}
	}()
	ws.SetReadLimit(1024)
	_ = ws.SetReadDeadline(time.Now().Add(m13ClientAuthTimeout))
	var challenge m13AgentChallenge
	if err = ws.ReadJSON(&challenge); err != nil {
		return fmt.Errorf("read agent challenge: %w", err)
	}
	if challenge.Type != "challenge" {
		return errors.New("invalid control-plane challenge")
	}
	signature, err := protocol.SignM13Challenge(ed25519.PrivateKey(privateBytes), identity.NodeID, challenge.Challenge)
	if err != nil {
		return err
	}
	_ = ws.SetWriteDeadline(time.Now().Add(m13ClientWriteTimeout))
	if err = ws.WriteJSON(m13AgentProof{
		Type: "authenticate", NodeID: identity.NodeID, Signature: signature,
	}); err != nil {
		return fmt.Errorf("send agent authentication: %w", err)
	}
	var accepted m13AgentAuthResponse
	if err = ws.ReadJSON(&accepted); err != nil {
		return fmt.Errorf("agent authentication failed: %w", err)
	}
	if accepted.Type != "authenticated" || accepted.NodeID != identity.NodeID {
		return errors.New("control plane did not authenticate the expected node")
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		heartbeat, sampleErr := sample(ctx)
		if sampleErr != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("collect heartbeat: %w", sampleErr)
		}
		if err = heartbeat.Validate(time.Now().UTC()); err != nil {
			return fmt.Errorf("invalid sampled heartbeat: %w", err)
		}
		_ = ws.SetWriteDeadline(time.Now().Add(m13ClientWriteTimeout))
		if err = ws.WriteJSON(heartbeat); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("send heartbeat: %w", err)
		}
		_ = ws.SetReadDeadline(time.Now().Add(m13ClientAckTimeout))
		var ack m13AgentHeartbeatAck
		if err = ws.ReadJSON(&ack); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("heartbeat acknowledgement failed: %w", err)
		}
		if ack.Type != "heartbeat_ack" {
			return errors.New("invalid heartbeat acknowledgement")
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
