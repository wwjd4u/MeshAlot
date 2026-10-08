\set ON_ERROR_STOP on
BEGIN;
DO $$
BEGIN
	IF current_database() !~ '^meshalot(_[a-z0-9]+)*$' THEN
		RAISE EXCEPTION 'Refusing M13 runtime test outside a MeshAlot database';
	END IF;
	IF NOT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema='public' AND table_name='node_status'
		  AND column_name='m13_telemetry' AND data_type='jsonb'
	) THEN
		RAISE EXCEPTION 'M13 telemetry column missing';
	END IF;
	IF NOT has_column_privilege('meshalot','node_status','m13_telemetry','UPDATE') THEN
		RAISE EXCEPTION 'Runtime role cannot update M13 telemetry';
	END IF;
	IF has_table_privilege('meshalot','node_status','DELETE') THEN
		RAISE EXCEPTION 'Runtime role must not delete node status';
	END IF;
END $$;
SELECT 'M13 TELEMETRY RUNTIME PRIVILEGES PASS' AS result;
ROLLBACK;
