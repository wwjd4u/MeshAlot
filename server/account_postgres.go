package server

import (
	"context"
	"database/sql"
	"fmt"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func (p *PostgresStore) Dashboard(ctx context.Context, userID string) (protocol.DashboardResponse, error) {
	var result protocol.DashboardResponse
	if err := p.db.QueryRowContext(ctx, `SELECT
        COALESCE((SELECT sum(amount_microunits) FROM wallet_transactions WHERE user_id=$1::uuid),0),
        COALESCE((SELECT count(*) FROM nodes n JOIN node_status s ON s.node_id=n.id WHERE n.user_id=$1::uuid AND s.status='online'),0),
        COALESCE((SELECT sum(amount_microunits) FROM wallet_transactions WHERE user_id=$1::uuid AND amount_microunits>0 AND created_at>=date_trunc('day',now())),0)`, userID).Scan(&result.BalanceMicrounits, &result.OnlineNodes, &result.TodayEarningsMicrounits); err != nil {
		return result, err
	}
	result.CurrentStatus = "Ready"
	if result.OnlineNodes == 0 {
		result.CurrentStatus = "No nodes online"
	}
	var score sql.NullFloat64
	if err := p.db.QueryRowContext(ctx, `SELECT cb.score FROM compute_benchmarks cb JOIN nodes n ON n.id=cb.node_id
        WHERE n.user_id=$1::uuid AND cb.score IS NOT NULL ORDER BY cb.observed_at DESC LIMIT 1`, userID).Scan(&score); err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if score.Valid {
		v := score.Float64
		result.ComputeScore = &v
	}
	score = sql.NullFloat64{}
	if err := p.db.QueryRowContext(ctx, `SELECT nb.score FROM network_benchmarks nb JOIN nodes n ON n.id=nb.node_id
        WHERE n.user_id=$1::uuid AND nb.score IS NOT NULL ORDER BY nb.observed_at DESC LIMIT 1`, userID).Scan(&score); err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if score.Valid {
		v := score.Float64
		result.NetworkScore = &v
	}
	var rate sql.NullInt64
	if err := p.db.QueryRowContext(ctx, `SELECT rate_microunits FROM pricing_rates ORDER BY effective_at DESC LIMIT 1`).Scan(&rate); err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if rate.Valid {
		v := rate.Int64
		result.CurrentRateMicrounits = &v
	}
	return result, nil
}

func (p *PostgresStore) NodeForUser(ctx context.Context, userID, nodeID string) (protocol.Node, error) {
	var node protocol.Node
	var heartbeat sql.NullTime
	err := p.db.QueryRowContext(ctx, `SELECT n.node_key,n.agent_version,s.status,s.mode,s.last_heartbeat
        FROM nodes n JOIN node_status s ON s.node_id=n.id
        WHERE n.user_id=$1::uuid AND n.node_key=$2`, userID, nodeID).Scan(&node.NodeID, &node.AgentVersion, &node.Status, &node.Mode, &heartbeat)
	if heartbeat.Valid {
		node.LastHeartbeat = heartbeat.Time.UTC()
	}
	return node, err
}

func (p *PostgresStore) Wallet(ctx context.Context, userID string) (protocol.WalletResponse, error) {
	var result protocol.WalletResponse
	if err := p.db.QueryRowContext(ctx, "SELECT COALESCE(sum(amount_microunits),0) FROM wallet_transactions WHERE user_id=$1::uuid", userID).Scan(&result.BalanceMicrounits); err != nil {
		return result, err
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id::text,job_id::text,transaction_type,amount_microunits,created_at
        FROM wallet_transactions WHERE user_id=$1::uuid ORDER BY created_at DESC,id DESC LIMIT 100`, userID)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	result.Transactions = make([]protocol.WalletTransaction, 0)
	for rows.Next() {
		var item protocol.WalletTransaction
		var jobID sql.NullString
		if err = rows.Scan(&item.ID, &jobID, &item.TransactionType, &item.AmountMicrounits, &item.CreatedAt); err != nil {
			return result, fmt.Errorf("scan wallet transaction: %w", err)
		}
		if jobID.Valid {
			item.JobID = jobID.String
		}
		item.CreatedAt = item.CreatedAt.UTC()
		result.Transactions = append(result.Transactions, item)
	}
	return result, rows.Err()
}

func (p *PostgresStore) Jobs(ctx context.Context, userID string) (protocol.JobsResponse, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT j.id::text,n.node_key,j.status,j.created_at
        FROM jobs j LEFT JOIN nodes n ON n.id=j.provider_node_id
        WHERE j.consumer_user_id=$1::uuid OR n.user_id=$1::uuid
        ORDER BY j.created_at DESC,j.id DESC LIMIT 100`, userID)
	if err != nil {
		return protocol.JobsResponse{}, err
	}
	defer rows.Close()
	result := protocol.JobsResponse{Jobs: make([]protocol.JobSummary, 0)}
	for rows.Next() {
		var item protocol.JobSummary
		var nodeID sql.NullString
		if err = rows.Scan(&item.ID, &nodeID, &item.Status, &item.CreatedAt); err != nil {
			return result, fmt.Errorf("scan job: %w", err)
		}
		if nodeID.Valid {
			item.ProviderNodeID = nodeID.String
		}
		item.CreatedAt = item.CreatedAt.UTC()
		result.Jobs = append(result.Jobs, item)
	}
	return result, rows.Err()
}
