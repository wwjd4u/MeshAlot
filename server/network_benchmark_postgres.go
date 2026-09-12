package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var ErrNetworkBenchmarkReplay = errors.New(
	"network benchmark report already accepted",
)

// InsertNetworkBenchmark writes an authenticated network benchmark through
// the narrowly scoped SECURITY DEFINER function created by M8 migration 7.
//
// The runtime database role intentionally has no direct INSERT, UPDATE, or
// DELETE privilege on network_benchmarks.
func (p *PostgresStore) InsertNetworkBenchmark(
	ctx context.Context,
	reportID string,
	nodeID string,
	payload []byte,
	score int,
	receivedAt time.Time,
) error {
	var inserted int

	err := p.db.QueryRowContext(
		ctx,
		`SELECT public.insert_network_benchmark(
		     $1::uuid,
		     $2,
		     $3::jsonb,
		     $4::numeric,
		     $5
		 )`,
		reportID,
		nodeID,
		payload,
		score,
		receivedAt.UTC(),
	).Scan(&inserted)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) &&
			string(pqErr.Code) == "23505" {

			return ErrNetworkBenchmarkReplay
		}

		return err
	}

	if inserted != 1 {
		return sql.ErrNoRows
	}

	return nil
}
