package server

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	m13AuthTimeout         = 10 * time.Second
	m13WriteTimeout        = 5 * time.Second
	m13IdleTimeout         = 90 * time.Second
	m13PingInterval        = 30 * time.Second
	m13MaxFrameBytes int64 = 1024
)

type m13PublicKeyLookup func(context.Context, string) (string, error)

type m13ChallengeFrame struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
}

type m13ProofFrame struct {
	Type      string `json:"type"`
	NodeID    string `json:"node_id"`
	Signature string `json:"signature"`
}

type m13AcceptedFrame struct {
	Type   string `json:"type"`
	NodeID string `json:"node_id"`
}

// m13WebSocket authenticates against an ALREADY enrolled Ed25519 node.
// It never enrolls a node, accepts a dev bearer token, or enables jobs.
func (s *Service) m13WebSocket(w http.ResponseWriter, r *http.Request) {
	if s.postgres == nil {
		writeError(w, http.StatusServiceUnavailable, "agent connection unavailable")
		return
	}
	serveM13WebSocketWithTelemetry(w, r, s.postgres.InventoryPublicKey, s.postgres.RecordM13Heartbeat)
}

// The control API runs behind a loopback Caddy reverse proxy in production.
// Only direct TLS or explicitly HTTPS-forwarded traffic from loopback is trusted.
// The agent must establish outbound wss://; plain ws:// is rejected here.
func m13SecureTransport(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback() &&
		strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// serveM13WebSocket accepts one proof per connection. The fresh nonce is
// connection-scoped and dies on failure, success, timeout, or disconnect.
// Heartbeat telemetry is enabled only with an authenticated recorder; jobs remain disabled.
func serveM13WebSocket(w http.ResponseWriter, r *http.Request, lookup m13PublicKeyLookup) {
	// Gate 2 compatibility: without a recorder, application frames remain disabled.
	serveM13WebSocketWithTelemetry(w, r, lookup, nil)
}

// This authenticated connection only accepts bounded heartbeats. The recorder
// receives the node ID from verified Ed25519 proof, never from the telemetry.
func serveM13WebSocketWithTelemetry(w http.ResponseWriter, r *http.Request,
	lookup m13PublicKeyLookup, record m13HeartbeatRecorder) {
	if lookup == nil {
		writeError(w, http.StatusServiceUnavailable, "agent connection unavailable")
		return
	}
	if !m13SecureTransport(r) {
		writeError(w, http.StatusForbidden, "secure connection required")
		return
	}
	if r.Header.Get("Origin") != "" {
		writeError(w, http.StatusForbidden, "browser origins not permitted")
		return
	}
	if !websocket.IsWebSocketUpgrade(r) {
		w.Header().Set("Upgrade", "websocket")
		writeError(w, http.StatusUpgradeRequired, "websocket upgrade required")
		return
	}
	challenge, err := protocol.NewM13Challenge()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "connection unavailable")
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return r.Header.Get("Origin") == "" },
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.SetReadLimit(m13MaxFrameBytes)
	_ = ws.SetReadDeadline(time.Now().Add(m13AuthTimeout))
	_ = ws.SetWriteDeadline(time.Now().Add(m13WriteTimeout))
	if err := ws.WriteJSON(m13ChallengeFrame{Type: "challenge", Challenge: challenge}); err != nil {
		return
	}
	frameType, frame, err := ws.ReadMessage()
	if err != nil || frameType != websocket.TextMessage {
		m13Reject(ws)
		return
	}
	var proof m13ProofFrame
	decoder := json.NewDecoder(bytes.NewReader(frame))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&proof) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		proof.Type != "authenticate" || !validUUIDv4(proof.NodeID) {
		m13Reject(ws)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	keyValue, err := lookup(ctx, proof.NodeID)
	cancel()
	if err != nil {
		m13Reject(ws)
		return
	}
	key, err := base64.RawStdEncoding.DecodeString(keyValue)
	if err != nil || len(key) != ed25519.PublicKeySize ||
		protocol.VerifyM13Challenge(ed25519.PublicKey(key), proof.NodeID, challenge, proof.Signature) != nil {
		m13Reject(ws)
		return
	}
	_ = ws.SetWriteDeadline(time.Now().Add(m13WriteTimeout))
	if err := ws.WriteJSON(m13AcceptedFrame{Type: "authenticated", NodeID: proof.NodeID}); err != nil {
		return
	}
	_ = ws.SetWriteDeadline(time.Time{})
	_ = ws.SetReadDeadline(time.Now().Add(m13IdleTimeout))
	ws.SetPongHandler(func(string) error {
		return ws.SetReadDeadline(time.Now().Add(m13IdleTimeout))
	})

	// Keep the authenticated connection alive with ping/pong, but do not
	// accept arbitrary application data or remote work.
	stopPings := make(chan struct{})
	defer close(stopPings)
	ticker := time.NewTicker(m13PingInterval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-stopPings:
				return
			case <-ticker.C:
				if err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(m13WriteTimeout)); err != nil {
					_ = ws.Close()
					return
				}
			}
		}
	}()
	for {
		messageType, frame, err := ws.ReadMessage()
		if err != nil {
			return
		}
		if messageType == websocket.TextMessage && record != nil {
			heartbeat, err := m13DecodeHeartbeat(frame, time.Now().UTC())
			if err != nil {
				_ = ws.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid heartbeat"),
					time.Now().Add(m13WriteTimeout))
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			err = record(ctx, proof.NodeID, heartbeat)
			cancel()
			if err != nil {
				_ = ws.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "heartbeat unavailable"),
					time.Now().Add(m13WriteTimeout))
				return
			}
			_ = ws.SetWriteDeadline(time.Now().Add(m13WriteTimeout))
			if err := ws.WriteJSON(m13HeartbeatAckFrame{Type: "heartbeat_ack"}); err != nil {
				return
			}
			_ = ws.SetWriteDeadline(time.Time{})
			_ = ws.SetReadDeadline(time.Now().Add(m13IdleTimeout))
			continue
		}
		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			_ = ws.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "telemetry not enabled"),
				time.Now().Add(m13WriteTimeout))
			return
		}
	}
}

func m13Reject(ws *websocket.Conn) {
	_ = ws.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "authentication failed"),
		time.Now().Add(m13WriteTimeout))
}
