package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// RecordM13Heartbeat updates only the node whose Ed25519 identity was proven
// on this WebSocket connection. Never accept a node ID from the payload.
// The server clock, not the client timestamp, controls online freshness.
func (p *PostgresStore) RecordM13Heartbeat(ctx context.Context, authenticatedNodeID string, h protocol.M13Heartbeat) error {
	if err := h.Validate(time.Now().UTC()); err != nil {
		return err
	}
	payload, err := json.Marshal(h)
	if err != nil {
		return err
	}
	result, err := p.db.ExecContext(ctx, `UPDATE node_status AS s
		SET status='online',
			mode=$2,
			observed_at=now(),
			last_heartbeat=now(),
			m13_telemetry=$3::jsonb
		FROM nodes AS n
		WHERE s.node_id=n.id
		  AND n.node_key=$1
		  AND n.identity_public_key IS NOT NULL
		  AND n.identity_public_key<>''`,
		authenticatedNodeID, h.AvailabilityMode, payload)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	return nil
}
