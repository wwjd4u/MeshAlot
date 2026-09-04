\set ON_ERROR_STOP on
BEGIN;

DO $$
BEGIN
    IF current_database() !~ '^meshalot(_[a-z0-9]+)*$' THEN
        RAISE EXCEPTION 'Refusing runtime privilege test outside a MeshAlot database';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_roles
        WHERE rolname='meshalot' AND rolcanlogin
          AND NOT rolsuper AND NOT rolcreatedb AND NOT rolcreaterole
          AND NOT rolreplication AND NOT rolbypassrls
    ) THEN
        RAISE EXCEPTION 'meshalot runtime role attributes are unsafe';
    END IF;

    IF NOT has_schema_privilege('meshalot','public','USAGE') THEN
        RAISE EXCEPTION 'missing public schema usage';
    END IF;

    -- Required authentication rights.
    IF NOT has_column_privilege('meshalot','users','id','SELECT')
       OR NOT has_column_privilege('meshalot','users','email','SELECT')
       OR NOT has_column_privilege('meshalot','users','password_hash','SELECT') THEN
        RAISE EXCEPTION 'missing required user authentication read rights';
    END IF;

    -- Required node enrollment, heartbeat, and dashboard rights.
    IF NOT has_table_privilege('meshalot','nodes','SELECT')
       OR NOT has_column_privilege('meshalot','nodes','user_id','INSERT')
       OR NOT has_column_privilege('meshalot','nodes','node_key','INSERT')
       OR NOT has_column_privilege('meshalot','nodes','agent_version','INSERT')
       OR NOT has_column_privilege('meshalot','nodes','agent_version','UPDATE') THEN
        RAISE EXCEPTION 'missing required nodes rights';
    END IF;

    IF NOT has_table_privilege('meshalot','node_status','SELECT')
       OR NOT has_column_privilege('meshalot','node_status','node_id','INSERT')
       OR NOT has_column_privilege('meshalot','node_status','status','INSERT')
       OR NOT has_column_privilege('meshalot','node_status','observed_at','INSERT')
       OR NOT has_column_privilege('meshalot','node_status','status','UPDATE')
       OR NOT has_column_privilege('meshalot','node_status','mode','UPDATE')
       OR NOT has_column_privilege('meshalot','node_status','observed_at','UPDATE')
       OR NOT has_column_privilege('meshalot','node_status','last_heartbeat','UPDATE') THEN
        RAISE EXCEPTION 'missing required node_status rights';
    END IF;

    IF NOT has_column_privilege('meshalot','compute_benchmarks','node_id','SELECT')
       OR NOT has_column_privilege('meshalot','compute_benchmarks','score','SELECT')
       OR NOT has_column_privilege('meshalot','compute_benchmarks','observed_at','SELECT')
       OR NOT has_column_privilege('meshalot','network_benchmarks','node_id','SELECT')
       OR NOT has_column_privilege('meshalot','network_benchmarks','score','SELECT')
       OR NOT has_column_privilege('meshalot','network_benchmarks','observed_at','SELECT') THEN
        RAISE EXCEPTION 'missing required benchmark read rights';
    END IF;

    IF NOT has_column_privilege('meshalot','wallet_transactions','id','SELECT')
       OR NOT has_column_privilege('meshalot','wallet_transactions','user_id','SELECT')
       OR NOT has_column_privilege('meshalot','wallet_transactions','job_id','SELECT')
       OR NOT has_column_privilege('meshalot','wallet_transactions','transaction_type','SELECT')
       OR NOT has_column_privilege('meshalot','wallet_transactions','amount_microunits','SELECT')
       OR NOT has_column_privilege('meshalot','wallet_transactions','created_at','SELECT') THEN
        RAISE EXCEPTION 'missing required wallet read rights';
    END IF;

    IF NOT has_column_privilege('meshalot','pricing_rates','rate_microunits','SELECT')
       OR NOT has_column_privilege('meshalot','pricing_rates','effective_at','SELECT') THEN
        RAISE EXCEPTION 'missing required pricing read rights';
    END IF;

    IF NOT has_column_privilege('meshalot','jobs','id','SELECT')
       OR NOT has_column_privilege('meshalot','jobs','consumer_user_id','SELECT')
       OR NOT has_column_privilege('meshalot','jobs','provider_node_id','SELECT')
       OR NOT has_column_privilege('meshalot','jobs','status','SELECT')
       OR NOT has_column_privilege('meshalot','jobs','created_at','SELECT') THEN
        RAISE EXCEPTION 'missing required jobs read rights';
    END IF;

    IF NOT has_column_privilege('meshalot','user_sessions','user_id','SELECT')
       OR NOT has_column_privilege('meshalot','user_sessions','token_hash','SELECT')
       OR NOT has_column_privilege('meshalot','user_sessions','expires_at','SELECT')
       OR NOT has_column_privilege('meshalot','user_sessions','user_id','INSERT')
       OR NOT has_column_privilege('meshalot','user_sessions','token_hash','INSERT')
       OR NOT has_column_privilege('meshalot','user_sessions','expires_at','INSERT')
       OR NOT has_table_privilege('meshalot','user_sessions','DELETE') THEN
        RAISE EXCEPTION 'missing required session rights';
    END IF;

    -- Forbidden account/data mutation rights.
    IF has_table_privilege('meshalot','users','INSERT')
       OR has_table_privilege('meshalot','users','UPDATE')
       OR has_table_privilege('meshalot','users','DELETE')
       OR has_column_privilege('meshalot','users','created_at','SELECT') THEN
        RAISE EXCEPTION 'runtime role has excess users rights';
    END IF;

    IF has_table_privilege('meshalot','nodes','DELETE')
       OR has_column_privilege('meshalot','nodes','id','INSERT')
       OR has_column_privilege('meshalot','nodes','created_at','INSERT')
       OR has_column_privilege('meshalot','nodes','id','UPDATE')
       OR has_column_privilege('meshalot','nodes','user_id','UPDATE')
       OR has_column_privilege('meshalot','nodes','node_key','UPDATE') THEN
        RAISE EXCEPTION 'runtime role has excess nodes rights';
    END IF;

    IF has_table_privilege('meshalot','node_status','DELETE')
       OR has_column_privilege('meshalot','node_status','node_id','UPDATE')
       OR has_column_privilege('meshalot','node_status','mode','INSERT')
       OR has_column_privilege('meshalot','node_status','last_heartbeat','INSERT') THEN
        RAISE EXCEPTION 'runtime role has excess node_status rights';
    END IF;

    IF has_table_privilege('meshalot','compute_benchmarks','INSERT')
       OR has_table_privilege('meshalot','compute_benchmarks','UPDATE')
       OR has_table_privilege('meshalot','compute_benchmarks','DELETE')
       OR has_table_privilege('meshalot','network_benchmarks','INSERT')
       OR has_table_privilege('meshalot','network_benchmarks','UPDATE')
       OR has_table_privilege('meshalot','network_benchmarks','DELETE') THEN
        RAISE EXCEPTION 'runtime role has benchmark write rights';
    END IF;

    IF has_table_privilege('meshalot','wallet_transactions','INSERT')
       OR has_table_privilege('meshalot','wallet_transactions','UPDATE')
       OR has_table_privilege('meshalot','wallet_transactions','DELETE') THEN
        RAISE EXCEPTION 'runtime role has wallet write rights';
    END IF;

    IF has_table_privilege('meshalot','jobs','INSERT')
       OR has_table_privilege('meshalot','jobs','UPDATE')
       OR has_table_privilege('meshalot','jobs','DELETE') THEN
        RAISE EXCEPTION 'runtime role has jobs write rights';
    END IF;

    IF has_table_privilege('meshalot','pricing_rates','INSERT')
       OR has_table_privilege('meshalot','pricing_rates','UPDATE')
       OR has_table_privilege('meshalot','pricing_rates','DELETE') THEN
        RAISE EXCEPTION 'runtime role has pricing write rights';
    END IF;

    IF has_table_privilege('meshalot','user_sessions','UPDATE')
       OR has_column_privilege('meshalot','user_sessions','id','INSERT')
       OR has_column_privilege('meshalot','user_sessions','created_at','INSERT')
       OR has_column_privilege('meshalot','user_sessions','id','SELECT')
       OR has_column_privilege('meshalot','user_sessions','created_at','SELECT') THEN
        RAISE EXCEPTION 'runtime role has excess session rights';
    END IF;
END
$$;

SELECT 'MILESTONE 5 RUNTIME PRIVILEGE ASSERTIONS PASSED' AS result;
ROLLBACK;
