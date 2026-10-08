# MeshAlot M13 — Gate 16 Closeout — 2026-10-07

**Gate 16 PASS — tracked migration 9→10 upgrade is compatible and preserves historical records in a disposable PostgreSQL 16 database.**
**No production DB migration, API deployment, repository main merge or agent-service enablement. M13 is IN PROGRESS.**

## Pinned source and environment
- Last Gate 15 closeout commit: `91a9ea6a8e17e187e7fbee1f237ed2b8c12e1302`.
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469` unchanged.
- Runner: `meshalot-ms02` on the MS-02 Ultra.
- [Successful workflow](https://github.com/wwjd4u/MeshAlot/actions/runs/37714771872), job `113108637142`.
- Database: ephemeral, Docker `postgres:16-alpine`, `--network none`, Unix socket only.
- Migration test ran the actual repository `database/migrate.py` as an independent tracked version/checksum runner, via a temporary host psql wrapper into the disposable container. The real repo migration files were unchanged.

## Upgrade validation
1. All 8 migration-runner construction tests PASS, including original pinned migrations 1 and 2.
2. Ran tracked `apply` for migrations 1–9, verified schema_migrations contains exactly versions 1–9.
3. Inserted a fake legacy node with a pre-M13 heartbeat timestamp, mode/status and a historic Network Score 67/report JSON payload. No real node IDs or historical reports were used.
4. Added the unchanged migration `000010_m13_heartbeat.up.sql` to a temporary copy of the migration directory and ran tracked `apply`. The runner recorded version 10 and verified its checksum.
5. Subsequent `status` PASS and repeated `apply` skipped version 10 without creating a duplicate migration record.
6. Verified historic fake node status/mode/heartbeat remained unchanged, nullable M13 telemetry was NULL for this older node, and the historical network benchmark payload and score were intact.
7. Applied restricted runtime grants and passed the M13 telemetry privilege test.
8. Deliberately changed only a temporary copy of migration 10's comment; tracked `status` correctly rejected the checksum mismatch. Restored the exact original file and verified `status` again.

Proof markers:
- `M13_TRACKED_V9=PASS`
- `M13_TRACKED_UPGRADE_HISTORY_PRESERVED=PASS`
- `M13 TELEMETRY RUNTIME PRIVILEGES PASS`
- `M13_GATE16_CHECKSUM_DRIFT_REJECTED=PASS`
- `M13_GATE16_TRACKED_MIGRATION=PASS`

## Release constraints
- Production PostgreSQL must be independently backed up and its currently recorded migration status inspected before a separately approved controlled deployment.
- If schema migration execution fails, inspect status, checksum and current state before retrying. Do not rewrite applied history.
- Migration 10 is additive and safe to leave in place if a new binary must be rolled back; never drop `m13_telemetry` blindly as a rollback operation.
- Server service, reverse-proxy config, persisted provider controls, identity, agent daemon install, physical Node001 reconnect/reboot, and Node002 sleep/wake tests are still pending.
- V100 baseline, benchmark results and evidence hashes untouched. No Tailscale.

**Gate 17:** controlled live acceptance planning and explicit operator approval before any production host/service/DB modification.
