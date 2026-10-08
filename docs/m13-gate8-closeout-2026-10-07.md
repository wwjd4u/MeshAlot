# MeshAlot M13 — Gate 8 Closeout — 2026-10-07

**Gate 8: PASS — automatic stale/offline classification and offline-to-online recovery.**
**M13 remains IN PROGRESS. No production deployment.**

## Source and scope
- Initial Gate 8 implementation: `3945273c2eb0b63eff38fe44055e08245c455066`.
- Unit-tested and runner-formatted code: `fad2efdc35ec10d1f4c6e4b267fa0676b169a179`.
- Isolated PostgreSQL integration test: `f201ae0b527fef1d8e56699997ee52761649d7c9`.
- Final runner-formatted, remote-pushed integration test: `cfe2656368f8273b9eb17ad4d3a33f06203d0948`.
- Production `main` remained at `1016c9b058690f34058dfab9cde7927e421fa469`.

## Implemented behaviour
- A node previously updated via an authenticated M13 heartbeat is treated as **offline after 90 seconds without a new successful heartbeat**.
- `NodesForUser` and `NodeForUser` derive effective offline status at read time to avoid stale dashboard information before the next periodic database sweep.
- `Dashboard` online-node totals exclude expired M13 heartbeat nodes.
- A 15-second server-side sweep persists `status='offline'` for expired M13 nodes only, using server timestamps. It does not overwrite the last heartbeat timestamp, latest telemetry JSONB, provider mode, historical scoring, enrollment or node identity.
- Legacy nodes with no M13 telemetry remain on their existing semantics; this gate does not impose the new stale rule on them.
- `RecordM13Heartbeat` from Gate 3 automatically returns a previously stale node to `online`, using the already authenticated existing node identity.
- The DB availability check at startup includes the M13 column so a server requiring migration 10 fails closed if the schema is missing.
- These code changes are confined to M13 branch.

## MS-02 remote validation
### Go test and race suite
- Workflow run: https://github.com/wwjd4u/MeshAlot/actions/runs/37712727158
- Job: `113102125330`.
- Focused stale classification, sweep cancellation and retry-on-database-error tests passed.
- Full Go suite, focused race tests and go vet passed; runner committed the formatting-only adjustment to M13 branch.

### Disposable PostgreSQL 16 integration
- Workflow run: https://github.com/wwjd4u/MeshAlot/actions/runs/37713180971
- Job: `113103586154`.
- Docker PostgreSQL 16 container used `--network none` and a Unix socket only, with no exposed listener. All 10 migrations and restricted M13 runtime grants applied successfully.
- A real Go/Postgres integration test verified **offline classification on reads before the sweep**, dashboard counts, per-account node isolation, sweep changing exactly the two expired M13 nodes, no redundant second sweep, timestamp/telemetry preservation, legacy node preservation, and online recovery with the same ID following a newly accepted M13 heartbeat.
- Output: `M13_POSTGRES_STALE_ONLINE_RECOVERY_AND_ACCOUNT_ISOLATION=PASS`, `M13_GATE8_DB_RUNTIME=PASS` and `M13_GATE8_ISOLATED_DATABASE=PASS`.
- Initial PostgreSQL workflow failed because of a disposable socket-directory setup; the next run reached real SQL and uncovered duplicate **fake test fixture** public keys. Final run corrected both, passed and removed its own previous orphaned temp socket directory. Neither failed run affected production or persistent node identities.

## Sleep semantics and next
- A sleeping host generally stops heartbeats. After 90s, the control plane reports the M13 node offline and stops presenting it as schedulable online.
- **Gate 9** must implement automatic reconnect with bounded backoff after interruption, including after waking from sleep.
- Later gates must verify actual disconnect/wake/reboot and automatic service restart. This gate by itself is not evidence of successful wake/reconnect.
- No Node002 enrollment, old M8 data, V100 evidence, production API or production database changes.
