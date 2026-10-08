package server

import (
	"context"
	"log/slog"
	"time"
)

const (
	// Three missed default 30s heartbeats trigger offline classification.
	// M13 status is reported using server time, not untrusted client time.
	M13StaleAfter         = 90 * time.Second
	M13StaleSweepInterval = 15 * time.Second
)

// M13EffectiveNodeStatus ensures the dashboard never displays an M13 node
// as online after its heartbeat expires, even before a periodic sweep runs.
// Legacy nodes with no M13 snapshot keep their previous behaviour.
func M13EffectiveNodeStatus(status string, lastHeartbeat time.Time,
	hasM13Telemetry bool, now time.Time) string {
	if status == "online" && hasM13Telemetry &&
		(lastHeartbeat.IsZero() || lastHeartbeat.Before(now.Add(-M13StaleAfter))) {
		return "offline"
	}
	return status
}

// MarkStaleM13NodesOffline modifies only status for expired M13 nodes.
// It deliberately preserves last_heartbeat, telemetry, scores, history,
// identity, mode, and enrolled/non-M13 node records.
func (p *PostgresStore) MarkStaleM13NodesOffline(ctx context.Context) (int64, error) {
	result, err := p.db.ExecContext(ctx, `UPDATE node_status
		SET status='offline'
		WHERE status='online'
		  AND m13_telemetry IS NOT NULL
		  AND (last_heartbeat IS NULL OR
		       last_heartbeat < now() - ($1::integer * interval '1 second'))`,
		int(M13StaleAfter/time.Second))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

type M13StaleSweepFunc func(context.Context) (int64, error)

// RunM13StaleMonitor runs a bounded DB housekeeping sweep until cancellation.
// It never acts without an explicit caller, so isolated tests remain safe.
func RunM13StaleMonitor(ctx context.Context, interval time.Duration,
	sweep M13StaleSweepFunc, logger *slog.Logger) {
	if ctx == nil || sweep == nil || logger == nil || interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			workCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			count, err := sweep(workCtx)
			cancel()
			if err != nil {
				logger.Error("M13 stale-node status update failed")
			} else if count > 0 {
				logger.Info("M13 stale nodes marked offline", "count", count)
			}
		}
	}
}
