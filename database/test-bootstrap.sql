\set ON_ERROR_STOP on
-- Never deploy this test harness against an application database.
SELECT current_database() = 'meshalot_test' AS test_database \gset
\if :test_database
\else
    \echo 'REFUSED: database must be meshalot_test'
    -- Force nonzero psql status, including noninteractive invocation.
    SELECT 1 / 0;
\endif
DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public') THEN
        RAISE EXCEPTION 'Test database is not empty. Do not delete it; inspect existing tables first.';
    END IF;
END;
$$;
\ir migrations/000001_core.up.sql
\ir migrations/000002_audit_and_ledger.up.sql
\ir tests/core.sql
