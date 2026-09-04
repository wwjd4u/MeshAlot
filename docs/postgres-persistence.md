# Milestone 4 API persistence (verification pending)

The two earlier migrations remain unchanged. Migration 3 adds node agent version,
mode and last-heartbeat fields. No existing table is removed.

The API uses PostgreSQL only when `MESHALOT_DATABASE_DSN` is set. A failed database
startup never falls back to memory. Otherwise the existing development memory
mode remains available. Configure `MESHALOT_POC_USER_ID` to an existing user UUID.
This fixed single-owner POC identity is not end-user authentication.

Use a restricted runtime role, not postgres or the table owner. The test setup
grants only user-ID lookup and node/status access. Migration credentials are
separate. The pool is capped at four connections, with two idle connections.

Database mode requires a non-default enrollment token of at least 32 characters.
`GET /v1/nodes` and `POST /v1/heartbeat` require `Authorization: Bearer TOKEN`.
Enrollment retains its token field. The agent now sends the bearer header.
These shared POC credentials are NOT suitable for a public multi-user marketplace.
Keep tokens and DSNs out of Git and logs; use the private service environment.
Keep the listener on `127.0.0.1`, behind the existing HTTPS proxy.

Heartbeats use server time. Re-enrollment preserves node UUID and heartbeat;
last-known online status is not proof that a node is still connected. Expiry and
offline detection remain later work. Database failure produces HTTP 503.

## Isolated verification

1. Apply tracked migrations to `meshalot_fresh_test` using `database/migrate.py`.
2. Run `database/setup-runtime-test.sql` once as postgres in that database. It
   creates database role `meshalot` for the existing service OS account. If the
   role already exists, inspect rather than dropping it or retrying blindly.
3. From the repository run as OS user meshalot:

```sh
MESHALOT_TEST_DSN='host=/var/run/postgresql dbname=meshalot_fresh_test user=meshalot sslmode=disable connect_timeout=5' \
MESHALOT_TEST_USER_ID=00000000-0000-4000-8000-000000000004 \
go test -count=1 -v ./...
```

The integration test refuses other databases and privileged/owner roles. It
creates a randomly named test node and leaves it in the isolated database; no
cleanup deletes data. It closes/recreates the HTTP service and connection pool,
then verifies persistence, stable node identity, authentication and database
failure responses. This is not a full process or VM restart test.

Before the Milestone 4 gate: compile/test on the VM, then verify a separate test
API process restart, configure the application database/runtime role, and confirm
persistence through a controlled live-service restart. Do not replace the running
binary, service environment, or apply to an application database until those
steps are explicitly reached. Alerts and the chat archive follow Milestone 4.
