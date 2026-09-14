package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var ErrComputeBenchmarkReplay = errors.New(
	"compute benchmark report already accepted",
)

// InsertComputeBenchmark writes an authenticated compute benchmark through
// the narrowly scoped SECURITY DEFINER function created by M9 migration 8.
//
// The runtime database role intentionally has no direct INSERT, UPDATE, or
// DELETE privilege on compute_benchmarks.
func (p *PostgresStore) InsertComputeBenchmark(
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
		`SELECT public.insert_compute_benchmark(
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
			return ErrComputeBenchmarkReplay
		}

		return err
	}

	if inserted != 1 {
		return sql.ErrNoRows
	}

	return nil
}
