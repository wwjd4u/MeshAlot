\set ON_ERROR_STOP on
SELECT current_database() = 'meshalot_test' AS test_database \gset
\if :test_database
\else
    \echo 'REFUSED: database must be meshalot_test'
    SELECT 1 / 0;
\endif
BEGIN;
-- Fixtures and helpers roll back on success or any assertion failure.
CREATE FUNCTION pg_temp.expect_error(statement text, expected_state text) RETURNS void
LANGUAGE plpgsql AS $$
DECLARE actual_state text;
BEGIN
    BEGIN
        EXECUTE statement;
    EXCEPTION WHEN OTHERS THEN
        GET STACKED DIAGNOSTICS actual_state = RETURNED_SQLSTATE;
        IF actual_state <> expected_state THEN
            RAISE EXCEPTION 'expected %, got % for %', expected_state, actual_state, statement;
        END IF;
        RETURN;
    END;
    RAISE EXCEPTION 'expected rejection but statement succeeded: %', statement;
END;
$$;

DO $$
DECLARE
    provider uuid; consumer uuid; stranger uuid; node uuid; model uuid;
    job uuid; failed_job uuid; reverse_job uuid; reverse_node uuid;
    balance numeric; hits integer; tab text; operation text;
BEGIN
    INSERT INTO users(email) VALUES ('provider-' || gen_random_uuid() || '@example.invalid') RETURNING id INTO provider;
    INSERT INTO users(email) VALUES ('consumer-' || gen_random_uuid() || '@example.invalid') RETURNING id INTO consumer;
    INSERT INTO users(email) VALUES ('stranger-' || gen_random_uuid() || '@example.invalid') RETURNING id INTO stranger;
    INSERT INTO nodes(user_id, node_key) VALUES (provider, gen_random_uuid()::text) RETURNING id INTO node;
    INSERT INTO node_status VALUES (node, 'online', now());
    INSERT INTO hardware_inventory(node_id,payload,observed_at) VALUES (node,'{"test":true}',now());
    INSERT INTO network_benchmarks(node_id,payload,score,observed_at) VALUES (node,'{}',90,now());
    INSERT INTO compute_benchmarks(node_id,payload,score,observed_at) VALUES (node,'{}',80,now());
    INSERT INTO models(name,quantization) VALUES (gen_random_uuid()::text,'test') RETURNING id INTO model;
    INSERT INTO pricing_rates(workload_type,rate_microunits,inputs,effective_at) VALUES ('test',100,'{}',now());
    INSERT INTO enrollment_tokens(user_id,token_hash,expires_at) VALUES (provider,gen_random_uuid()::text,now()+interval '1 hour');
    INSERT INTO jobs(consumer_user_id,provider_node_id,model_id,status)
        VALUES (consumer,node,model,'queued') RETURNING id INTO job;
    PERFORM pg_temp.expect_error(format('UPDATE jobs SET status = ''completed'' WHERE id = %L',job),'23514');
    PERFORM pg_temp.expect_error(format('UPDATE jobs SET status = ''bogus'' WHERE id = %L',job),'23514');
    UPDATE jobs SET status='running' WHERE id=job;
    UPDATE jobs SET status='completed' WHERE id=job;
    PERFORM pg_temp.expect_error(format('INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES (%L,%L,''earning'',100,''no-usage'')',provider,job),'23514');
    INSERT INTO usage_records(job_id,metrics) VALUES (job,'{"tokens":100}');
    INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key)
        VALUES (provider,job,'earning',1000,job||':earning'),(consumer,job,'usage',-1000,job||':usage');
    SELECT balance_microunits INTO balance FROM wallet_balances WHERE user_id=provider;
    IF balance IS DISTINCT FROM 1000 THEN RAISE EXCEPTION 'provider balance incorrect'; END IF;
    SELECT balance_microunits INTO balance FROM wallet_balances WHERE user_id=consumer;
    IF balance IS DISTINCT FROM -1000 THEN RAISE EXCEPTION 'consumer balance incorrect'; END IF;
    SELECT count(*) INTO hits FROM job_usage_parties WHERE job_id=job
        AND provider_user_id=provider AND consumer_user_id=consumer;
    IF hits <> 1 THEN RAISE EXCEPTION 'usage party trace failed'; END IF;
    PERFORM pg_temp.expect_error(format('INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES (%L,%L,''earning'',1000,%L)',provider,job,job||':earning'),'23505');
    PERFORM pg_temp.expect_error(format('INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES (%L,%L,''earning'',1000,''different-key'')',provider,job),'23505');
    PERFORM pg_temp.expect_error(format('INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES (%L,%L,''earning'',1000,''wrong-party'')',stranger,job),'23514');
    PERFORM pg_temp.expect_error(format('INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES (%L,%L,''earning'',-1,''wrong-sign'')',provider,job),'23514');
    PERFORM pg_temp.expect_error(format('INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES (%L,%L,''unknown'',1,''wrong-type'')',provider,job),'23514');

    INSERT INTO jobs(consumer_user_id,provider_node_id,status) VALUES (consumer,node,'queued') RETURNING id INTO failed_job;
    UPDATE jobs SET status='failed' WHERE id=failed_job;
    INSERT INTO job_events(job_id,event_type,payload) VALUES (failed_job,'failure_detail','{"reason":"simulated failure"}');
    INSERT INTO usage_records(job_id,metrics) VALUES (failed_job,'{"tokens":0}');
    PERFORM pg_temp.expect_error(format('INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES (%L,%L,''usage'',-1,''failed-charge'')',consumer,failed_job),'23514');
    SELECT count(*) INTO hits FROM job_events WHERE job_id=failed_job;
    IF hits <> 3 THEN RAISE EXCEPTION 'failed-job audit missing'; END IF;
    PERFORM pg_temp.expect_error(format('UPDATE jobs SET status=''running'' WHERE id=%L',failed_job),'23514');
    PERFORM pg_temp.expect_error(format('UPDATE nodes SET user_id=%L WHERE id=%L',stranger,node),'23514');
    PERFORM pg_temp.expect_error(format('UPDATE jobs SET consumer_user_id=%L WHERE id=%L',stranger,job),'23514');

    -- Reverse provider/consumer roles: the first provider spends part of its earnings.
    INSERT INTO nodes(user_id,node_key) VALUES (consumer,gen_random_uuid()::text) RETURNING id INTO reverse_node;
    INSERT INTO jobs(consumer_user_id,provider_node_id,status) VALUES (provider,reverse_node,'queued') RETURNING id INTO reverse_job;
    UPDATE jobs SET status='running' WHERE id=reverse_job;
    UPDATE jobs SET status='completed' WHERE id=reverse_job;
    INSERT INTO usage_records(job_id,metrics) VALUES (reverse_job,'{"tokens":40}');
    INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key)
        VALUES (consumer,reverse_job,'earning',400,reverse_job||':earning'),(provider,reverse_job,'usage',-400,reverse_job||':usage');
    SELECT balance_microunits INTO balance FROM wallet_balances WHERE user_id=provider;
    IF balance IS DISTINCT FROM 600 THEN RAISE EXCEPTION 'reverse-role balance incorrect'; END IF;

    FOREACH tab IN ARRAY ARRAY['wallet_transactions','job_events','usage_records'] LOOP
        PERFORM pg_temp.expect_error(format('UPDATE %I SET id=id',tab),'23514');
        PERFORM pg_temp.expect_error(format('DELETE FROM %I',tab),'23514');
        PERFORM pg_temp.expect_error(format('TRUNCATE %I',tab),'23514');
    END LOOP;
    PERFORM pg_temp.expect_error('DELETE FROM jobs','23514');
    PERFORM pg_temp.expect_error('TRUNCATE jobs CASCADE','23514');
    RAISE NOTICE 'PASS: core records, job audit, role reversal, balances, duplicate/sign/party guards, and append-only protections';
END;
$$;
ROLLBACK;
\echo 'MESHALOT DATABASE TESTS PASSED - fixture changes rolled back'
