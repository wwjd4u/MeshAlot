package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

type m13HeartbeatRecorder func(context.Context, string, protocol.M13Heartbeat) error

type m13HeartbeatAckFrame struct {
	Type string `json:"type"`
}

// Frames never specify node identity; strict decoding also prohibits an
// unrecognized node_id field. This prevents one authenticated node from
// attempting to update another node's operational state.
func m13DecodeHeartbeat(frame []byte, now time.Time) (protocol.M13Heartbeat, error) {
	var h protocol.M13Heartbeat
	if len(frame) == 0 || len(frame) > int(m13MaxFrameBytes) {
		return h, errors.New("invalid heartbeat size")
	}
	d := json.NewDecoder(bytes.NewReader(frame))
	d.DisallowUnknownFields()
	if err := d.Decode(&h); err != nil {
		return protocol.M13Heartbeat{}, errors.New("invalid heartbeat JSON")
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		return protocol.M13Heartbeat{}, errors.New("unexpected heartbeat trailing data")
	}
	if err := h.Validate(now); err != nil {
		return protocol.M13Heartbeat{}, err
	}
	return h, nil
}
