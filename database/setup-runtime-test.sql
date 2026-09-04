\set ON_ERROR_STOP on
SELECT current_database()='meshalot_fresh_test' AS correct_database \gset
\if :correct_database
\else
    SELECT 1/0;
\endif
BEGIN;
-- Deliberately fail on an existing role; never silently change its privileges.
CREATE ROLE meshalot LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
GRANT CONNECT ON DATABASE meshalot_fresh_test TO meshalot;
GRANT USAGE ON SCHEMA public TO meshalot;
GRANT SELECT(id) ON users TO meshalot;
GRANT SELECT, INSERT ON nodes, node_status TO meshalot;
GRANT UPDATE(agent_version) ON nodes TO meshalot;
GRANT UPDATE ON node_status TO meshalot;
INSERT INTO users(id,email) VALUES
    ('00000000-0000-4000-8000-000000000004','poc-persistence-test@example.invalid');
COMMIT;
-- Unix socket peer authentication maps the existing OS account meshalot.
-- No password, global pg_hba change, database ownership or migration grants.
