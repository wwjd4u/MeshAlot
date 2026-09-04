BEGIN;
ALTER TABLE nodes ADD COLUMN agent_version text NOT NULL DEFAULT '';
ALTER TABLE node_status ADD COLUMN mode text NOT NULL DEFAULT 'available';
ALTER TABLE node_status ADD COLUMN last_heartbeat timestamptz;
COMMIT;
