# MeshAlot — M13 Gate 3 Closeout — 2026-10-07

**Gate 3 (authenticated operational heartbeat telemetry): PASS — isolated code and database validation.**
**M13 remains IN PROGRESS. Do not deploy this branch to production yet.**

## Immutable verification refs

- Governing guide: AI Mesh POC Build, Test, and Release Guide (Google Drive), Milestone 13.
- Before Gate 3: `m13-persistent-heartbeat` = `610aa62e7e9a2de896c56ee42c5914826be1eda4`.
- Implementation tested: `a4a4facef735eef95b854f4162ab6ad4e8580b8e`.
- `main` unchanged: `1016c9b058690f34058dfab9cde7927e421fa469`.
- All changes to isolated M13 branch; no production deploy.

## Implementation

- `protocol/v1/m13_heartbeat.go`: bounded type/validation for online state; GPU utilization and free VRAM (both optional together for no GPU); RAM availability; CPU load percent; active job state; latency; provider availability mode; manual-pause reporting; and observation timestamp.
- Heartbeat payload has **no node_id**; strict parsing rejects unknown fields and any attempted node spoof.
- `server/m13_websocket.go`: only authenticated Ed25519 nodes can send frames. Server calls recorder exclusively with the verified identity; acknowledges only after successful recording; rejects malformed/unexpected data. No arbitrary command or job execution.
- `server/m13_heartbeat_postgres.go`: updates the authenticated node's `node_status` only. Uses server timestamps; retains the latest operational snapshot in JSONB.
- `database/migrations/000010_m13_heartbeat.up.sql`: additive nullable JSONB column, leaving existing status and historical records untouched.
- `database/runtime-grants.sql`: least-privilege UPDATE for the new telemetry column. `database/tests/milestone13_runtime_privileges.sql` checks column and runtime privileges.

## Actual remote test evidence

### Go runner (PASS)

- Runner: `meshalot-ms02`.
- Run: https://github.com/wwjd4u/MeshAlot/actions/runs/37710797181
- Job ID: `113095991491`.
- Exact source commit: `a4a4facef735eef95b854f4162ab6ad4e8580b8e`.
- `go mod verify`, focused M13 auth + heartbeat validation + socket tests, full `go test ./...` and `go vet ./...` all passed.
- Tests specifically reject malformed/stale/implausible telemetry, a forged node_id, unauthenticated identity, replayed signatures, and failed storage acknowledgement.
- `gofmt -l` produced no files requiring formatting.

### Disposable PostgreSQL 16 (PASS)

- Runner: `meshalot-ms02`.
- Run: https://github.com/wwjd4u/MeshAlot/actions/runs/37711015130
- Job ID: `113096696724`.
- Exact source commit: `a4a4facef735eef95b854f4162ab6ad4e8580b8e`.
- Local PostgreSQL client/server binaries were not installed on the runner; the first two DB preflight runs failed before migrations. The corrected test used a temporary `postgres:16-alpine` container with `--network none`, a read-only source mount, and automatic container cleanup.
- Applied all 10 migrations from scratch to an ephemeral database; re-applied restricted runtime grants; executed the M13 privilege test; verified one authorized scoped status update and zero updates for a different node; tested committed status/JSONB value inside a transaction and rolled back.
- Log markers: `M13 TELEMETRY RUNTIME PRIVILEGES PASS` and `M13_GATE3_ISOLATED_POSTGRES=PASS`.
- This was a disposable test database, **not** the production or existing test database.

## Preservation and remaining work

- No change to Node002 enrollment, M8 historical reports, V100 benchmark or evidence.
- No production migration, deployment, host port forwarding or Tailscale.
- Gate 3 implemented **server-side acceptance and persistence only**. The agent has not yet been wired to establish its own outbound persistent connection and send live heartbeat frames. No claims of live node status, automatic offline classification, reconnection, or reboot recovery yet.

**Next — M13 Gate 4:** wire the agent's existing persistent Ed25519 identity into an outbound `wss://` connection, answer the challenge, report bounded operational telemetry and handle heartbeat acknowledgements; test with isolated fake server and preserve Node002 identity. Subsequent gates: stale/offline determination, bounded reconnect/backoff, network-interruption and reboot recovery. No production deploy until milestone gates pass.
