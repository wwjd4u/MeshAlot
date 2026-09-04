BEGIN;
DO $$
BEGIN
    IF current_database() !~ '^meshalot(_[a-z0-9]+)*$' THEN
        RAISE EXCEPTION 'Refusing runtime grants outside a MeshAlot database';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_roles
        WHERE rolname='meshalot' AND rolcanlogin
          AND NOT rolsuper AND NOT rolcreatedb AND NOT rolcreaterole
          AND NOT rolreplication AND NOT rolbypassrls
    ) THEN
        RAISE EXCEPTION 'meshalot runtime role missing or unsafe';
    END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO meshalot;
GRANT SELECT (id,email,password_hash) ON users TO meshalot;

GRANT SELECT ON nodes, node_status TO meshalot;
GRANT INSERT (user_id,node_key,agent_version) ON nodes TO meshalot;
GRANT UPDATE (agent_version) ON nodes TO meshalot;
GRANT INSERT (node_id,status,observed_at) ON node_status TO meshalot;
GRANT UPDATE (status,mode,observed_at,last_heartbeat) ON node_status TO meshalot;

GRANT SELECT (node_id,score,observed_at) ON compute_benchmarks, network_benchmarks TO meshalot;
GRANT SELECT (id,user_id,job_id,transaction_type,amount_microunits,created_at) ON wallet_transactions TO meshalot;
GRANT SELECT (rate_microunits,effective_at) ON pricing_rates TO meshalot;
GRANT SELECT (id,consumer_user_id,provider_node_id,status,created_at) ON jobs TO meshalot;

GRANT SELECT (user_id,token_hash,expires_at) ON user_sessions TO meshalot;
GRANT INSERT (user_id,token_hash,expires_at) ON user_sessions TO meshalot;
GRANT DELETE ON user_sessions TO meshalot;
COMMIT;
