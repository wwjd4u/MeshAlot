BEGIN;
SET LOCAL search_path = public, pg_catalog;

-- POC credits only. Refunds, fees, and real-money settlement are later migrations.
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check
    CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled'));
ALTER TABLE wallet_transactions ALTER COLUMN job_id SET NOT NULL;
ALTER TABLE wallet_transactions ADD CONSTRAINT wallet_kind_and_sign_check CHECK (
    (transaction_type = 'earning' AND amount_microunits > 0) OR
    (transaction_type = 'usage' AND amount_microunits < 0)
);
ALTER TABLE wallet_transactions ADD CONSTRAINT wallet_nonempty_key
    CHECK (length(btrim(idempotency_key)) > 0);
-- A different idempotency key must not charge/pay the same job twice.
ALTER TABLE wallet_transactions ADD CONSTRAINT wallet_one_entry_per_kind
    UNIQUE (job_id, transaction_type);
ALTER TABLE pricing_rates ADD CONSTRAINT pricing_nonnegative CHECK (rate_microunits >= 0);

CREATE FUNCTION meshalot_reject_mutation() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    RAISE EXCEPTION '% is append-only: % forbidden', TG_TABLE_NAME, TG_OP
        USING ERRCODE = '23514';
END;
$$;

-- Statement triggers also reject TRUNCATE and no-op UPDATE/DELETE statements.
CREATE TRIGGER wallet_append_only BEFORE UPDATE OR DELETE OR TRUNCATE
    ON wallet_transactions FOR EACH STATEMENT EXECUTE FUNCTION meshalot_reject_mutation();
CREATE TRIGGER events_append_only BEFORE UPDATE OR DELETE OR TRUNCATE
    ON job_events FOR EACH STATEMENT EXECUTE FUNCTION meshalot_reject_mutation();
CREATE TRIGGER usage_append_only BEFORE UPDATE OR DELETE OR TRUNCATE
    ON usage_records FOR EACH STATEMENT EXECUTE FUNCTION meshalot_reject_mutation();
CREATE TRIGGER jobs_preserve_history BEFORE DELETE OR TRUNCATE
    ON jobs FOR EACH STATEMENT EXECUTE FUNCTION meshalot_reject_mutation();

CREATE FUNCTION meshalot_guard_node_identity() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    IF (NEW.id, NEW.user_id, NEW.node_key) IS DISTINCT FROM
       (OLD.id, OLD.user_id, OLD.node_key) THEN
        RAISE EXCEPTION 'node identity and ownership are immutable' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER node_identity BEFORE UPDATE ON nodes
    FOR EACH ROW EXECUTE FUNCTION meshalot_guard_node_identity();

CREATE FUNCTION meshalot_guard_job() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.status <> 'queued' THEN
            RAISE EXCEPTION 'new jobs must be queued' USING ERRCODE = '23514';
        END IF;
    ELSE
        IF (NEW.id, NEW.consumer_user_id, NEW.model_id, NEW.created_at) IS DISTINCT FROM
           (OLD.id, OLD.consumer_user_id, OLD.model_id, OLD.created_at)
           OR (OLD.provider_node_id IS NOT NULL AND
               NEW.provider_node_id IS DISTINCT FROM OLD.provider_node_id) THEN
            RAISE EXCEPTION 'job identity and assigned parties are immutable' USING ERRCODE = '23514';
        END IF;
        IF OLD.status IN ('completed', 'failed', 'cancelled') AND NEW IS DISTINCT FROM OLD THEN
            RAISE EXCEPTION 'terminal jobs are immutable' USING ERRCODE = '23514';
        END IF;
        IF NEW.status <> OLD.status AND NOT (
            (OLD.status = 'queued' AND NEW.status IN ('running', 'failed', 'cancelled')) OR
            (OLD.status = 'running' AND NEW.status IN ('completed', 'failed', 'cancelled'))
        ) THEN
            RAISE EXCEPTION 'invalid job status transition' USING ERRCODE = '23514';
        END IF;
    END IF;
    IF NEW.status IN ('running', 'completed') AND NEW.provider_node_id IS NULL THEN
        RAISE EXCEPTION 'running/completed job requires provider' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER job_guard BEFORE INSERT OR UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION meshalot_guard_job();

CREATE FUNCTION meshalot_record_job_status() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO public.job_events(job_id, event_type, payload)
        VALUES (NEW.id, 'created', jsonb_build_object('status', NEW.status));
    ELSIF NEW.status IS DISTINCT FROM OLD.status THEN
        INSERT INTO public.job_events(job_id, event_type, payload)
        VALUES (NEW.id, 'status_changed', jsonb_build_object('from', OLD.status, 'to', NEW.status));
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER job_status_audit AFTER INSERT OR UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION meshalot_record_job_status();

CREATE FUNCTION meshalot_check_usage() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
DECLARE j public.jobs%ROWTYPE;
BEGIN
    SELECT * INTO j FROM public.jobs WHERE id = NEW.job_id FOR SHARE;
    IF NOT FOUND OR j.provider_node_id IS NULL OR
       j.status NOT IN ('completed', 'failed', 'cancelled') THEN
        RAISE EXCEPTION 'usage needs a terminal job and assigned provider' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER usage_guard BEFORE INSERT ON usage_records
    FOR EACH ROW EXECUTE FUNCTION meshalot_check_usage();

CREATE FUNCTION meshalot_check_wallet() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
DECLARE j public.jobs%ROWTYPE; provider uuid;
BEGIN
    SELECT * INTO j FROM public.jobs WHERE id = NEW.job_id FOR SHARE;
    IF NOT FOUND OR j.status <> 'completed' OR NOT EXISTS (
        SELECT 1 FROM public.usage_records WHERE job_id = NEW.job_id
    ) THEN
        RAISE EXCEPTION 'wallet entries require completed job and usage' USING ERRCODE = '23514';
    END IF;
    SELECT user_id INTO provider FROM public.nodes WHERE id = j.provider_node_id;
    IF (NEW.transaction_type = 'earning' AND NEW.user_id IS DISTINCT FROM provider)
       OR (NEW.transaction_type = 'usage' AND NEW.user_id IS DISTINCT FROM j.consumer_user_id) THEN
        RAISE EXCEPTION 'wallet user does not match job party' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER wallet_guard BEFORE INSERT ON wallet_transactions
    FOR EACH ROW EXECUTE FUNCTION meshalot_check_wallet();

CREATE VIEW wallet_balances AS
SELECT u.id AS user_id, COALESCE(sum(w.amount_microunits), 0) AS balance_microunits
FROM users u LEFT JOIN wallet_transactions w ON w.user_id = u.id GROUP BY u.id;

CREATE VIEW job_usage_parties AS
SELECT r.id AS usage_id, r.job_id, j.consumer_user_id, j.provider_node_id,
       n.user_id AS provider_user_id, r.metrics
FROM usage_records r JOIN jobs j ON j.id = r.job_id JOIN nodes n ON n.id = j.provider_node_id;

CREATE INDEX job_events_job_id_idx ON job_events(job_id, id);
CREATE INDEX wallet_transactions_user_id_idx ON wallet_transactions(user_id, created_at);
COMMIT;
