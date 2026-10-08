BEGIN;
-- Add only the latest operational snapshot; no enrollment, scores, historic
-- reports or existing status values are modified by this schema migration.
ALTER TABLE node_status ADD COLUMN m13_telemetry jsonb;
COMMIT;
