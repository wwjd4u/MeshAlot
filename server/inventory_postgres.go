package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

var ErrInventoryReplay = errors.New("inventory report already accepted")

func (p *PostgresStore) InventoryPublicKey(
	ctx context.Context,
	nodeID string,
) (string, error) {

	var publicKey string

	// Agent authentication is node-scoped. The stored Ed25519 public key
	// authenticates the specific node regardless of which account enrolled it.
	err := p.db.QueryRowContext(
		ctx,
		`SELECT identity_public_key
		 FROM nodes
		 WHERE node_key=$1
		   AND identity_public_key IS NOT NULL
		   AND identity_public_key<>''`,
		nodeID,
	).Scan(&publicKey)

	if err != nil {
		return "", err
	}

	return publicKey, nil
}

func (p *PostgresStore) InsertInventory(
	ctx context.Context,
	reportID string,
	nodeID string,
	payload []byte,
	receivedAt time.Time,
) error {

	result, err := p.db.ExecContext(
		ctx,
		`INSERT INTO hardware_inventory(
		     id,
		     node_id,
		     payload,
		     observed_at
		 )
		 SELECT
		     $1::uuid,
		     n.id,
		     $3::jsonb,
		     $4
		 FROM nodes n
		 WHERE n.node_key=$2`,
		reportID,
		nodeID,
		payload,
		receivedAt.UTC(),
	)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) &&
			string(pqErr.Code) == "23505" {
			return ErrInventoryReplay
		}

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

func (p *PostgresStore) LatestInventoryForUser(
	ctx context.Context,
	userID string,
	nodeID string,
) (protocol.InventoryReport, error) {

	var report protocol.InventoryReport
	var payload []byte

	err := p.db.QueryRowContext(
		ctx,
		`SELECT
		     h.id::text,
		     h.observed_at,
		     h.payload
		 FROM hardware_inventory h
		 JOIN nodes n
		   ON n.id=h.node_id
		 WHERE n.node_key=$1
		   AND n.user_id=$2::uuid
		 ORDER BY h.observed_at DESC, h.id DESC
		 LIMIT 1`,
		nodeID,
		userID,
	).Scan(
		&report.ReportID,
		&report.ReceivedAt,
		&payload,
	)

	if err != nil {
		return protocol.InventoryReport{}, err
	}

	if err := json.Unmarshal(
		payload,
		&report.Inventory,
	); err != nil {
		return protocol.InventoryReport{},
			fmt.Errorf("decode hardware inventory: %w", err)
	}

	report.NodeID = nodeID
	report.ReceivedAt = report.ReceivedAt.UTC()

	return report, nil
}
