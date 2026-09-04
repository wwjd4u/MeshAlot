package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// PostgresStore owns database connectivity. Agent enrollment remains attached to
// the configured POC owner while web requests are scoped by authenticated user ID.
type PostgresStore struct {
	db    *sql.DB
	owner string
}

var ErrNodeOwnership = errors.New("node belongs to a different owner")

func OpenPostgres(ctx context.Context, dsn, owner string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.New("invalid database configuration")
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var exists bool
	err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1::uuid)", owner).Scan(&exists)
	if err != nil || !exists {
		db.Close()
		return nil, errors.New("database unavailable or POC owner missing")
	}
	rows, err := db.QueryContext(ctx, "SELECT n.agent_version,s.mode,s.last_heartbeat FROM nodes n JOIN node_status s ON s.node_id=n.id LIMIT 0")
	if err != nil {
		db.Close()
		return nil, errors.New("node persistence schema unavailable")
	}
	rows.Close()
	return &PostgresStore{db: db, owner: owner}, nil
}

func (p *PostgresStore) Close() error                   { return p.db.Close() }
func (p *PostgresStore) Ping(ctx context.Context) error { return p.db.PingContext(ctx) }

func (p *PostgresStore) Enroll(ctx context.Context, nodeID, version string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "INSERT INTO nodes(user_id,node_key,agent_version) VALUES ($1::uuid,$2,$3) ON CONFLICT(node_key) DO NOTHING", p.owner, nodeID, version)
	if err != nil {
		return err
	}
	var id, owner string
	err = tx.QueryRowContext(ctx, "SELECT id::text,user_id::text FROM nodes WHERE node_key=$1 FOR UPDATE", nodeID).Scan(&id, &owner)
	if err != nil {
		return err
	}
	var owned bool
	if err = tx.QueryRowContext(ctx, "SELECT $1::uuid=$2::uuid", owner, p.owner).Scan(&owned); err != nil {
		return err
	}
	if !owned {
		return ErrNodeOwnership
	}
	if _, err = tx.ExecContext(ctx, "UPDATE nodes SET agent_version=$2 WHERE id=$1::uuid", id, version); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO node_status(node_id,status,observed_at) VALUES ($1::uuid,'enrolled',now()) ON CONFLICT(node_id) DO NOTHING", id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (p *PostgresStore) Heartbeat(ctx context.Context, nodeID, mode string) error {
	result, err := p.db.ExecContext(ctx, `UPDATE node_status s SET status='online',mode=$3,
        observed_at=now(),last_heartbeat=now() FROM nodes n
        WHERE s.node_id=n.id AND n.node_key=$1 AND n.user_id=$2::uuid`, nodeID, p.owner, mode)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (p *PostgresStore) Nodes(ctx context.Context) ([]protocol.Node, error) {
	return p.NodesForUser(ctx, p.owner)
}

func (p *PostgresStore) NodesForUser(ctx context.Context, userID string) ([]protocol.Node, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT n.node_key,n.agent_version,s.status,s.mode,s.last_heartbeat
        FROM nodes n JOIN node_status s ON s.node_id=n.id WHERE n.user_id=$1::uuid ORDER BY n.node_key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodes := make([]protocol.Node, 0)
	for rows.Next() {
		var n protocol.Node
		var heartbeat sql.NullTime
		if err = rows.Scan(&n.NodeID, &n.AgentVersion, &n.Status, &n.Mode, &heartbeat); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		if heartbeat.Valid {
			n.LastHeartbeat = heartbeat.Time.UTC()
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}
