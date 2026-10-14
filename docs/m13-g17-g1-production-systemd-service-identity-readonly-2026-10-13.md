# MeshAlot M13 — Gate 17/1: approved read-only production API systemd service-account metadata

Date: 2026-10-13 (America/Chicago); GitHub Actions timestamps may read 2026-10-14 UTC
Status: **READ-ONLY SERVICE-ACCOUNT METADATA INSPECTION PASS; POSTGRESQL AUTH/ROLE INVENTORY UNVERIFIED; G17/1 OPEN.**
G17/2–G17/7 **NOT STARTED**. M13 **OPEN / NOT RELEASE-READY**.

## Operator approval and hard scope

The operator explicitly approved ONE production read-only inspection of the MeshAlot API's declared systemd OS service-account metadata on the known host `meshalot-control-01`. Authorization required the **already existing, still-valid, host-key-verified, VM-specific SSH identity** with **at least 12 minutes remaining** and specifically **excluded** SSH renewal, sudo, privilege escalation, PostgreSQL connections, credential reads, environment-file contents, `pg_hba.conf` inspection, production changes, and automatic retries.

The inspection was implemented only on `infra/ms02-runner-gateway`, using the established `meshalot-ms02` GitHub Actions runner and exact existing Ed25519 identity. No credentials or private keys were added to GitHub or Actions.

## Preflight and single remote inspection evidence

- [Source-pinned syntax preflight 38492822821](https://github.com/wwjd4u/MeshAlot/actions/runs/38492822821): **PASS** for Bash syntax, embedded Python AST, and pinned source blob SHA. Both production access and SSH renewal explicitly **NOT ATTEMPTED** in preflight. The one-time preflight was disabled after success; [follow-up SKIPPED 38492874186](https://github.com/wwjd4u/MeshAlot/actions/runs/38492874186).
- [Read-only production systemd metadata inspection 38492890931](https://github.com/wwjd4u/MeshAlot/actions/runs/38492890931): **PASS**. The previously registered IAP/SSH identity met the 12-minute minimum validity margin, the strict pinned host key was enforced, and the wrong-host guard passed. The one-time attempt marker was claimed.
- An allowlisted, bounded `systemctl list-unit-files meshalot*.service` check found exactly one appropriate service unit. A restricted `systemctl show` read only the properties `User`, `DynamicUser`, `LoadState`, and `ActiveState`, with field validation and no other systemd properties.
- Sanitized returned markers: `M13_G17_UNIT_METADATA=PASS_READONLY`, `M13_G17_UNIT_DECLARED_USER_CLASS=EXPLICIT_MESHALOT`, `M13_G17_UNIT_DYNAMIC_USER=NO`, and `M13_G17_UNIT_ACTUAL_DB_AUTHORIZATION=UNVERIFIED`. This supports that the **declared service `User=meshalot`** is an explicit, non-dynamic account, and that the selected service unit was **loaded and active at inspection time**.
- The inspection explicitly reported `M13_G17_UNIT_SUDO_OR_PRIVILEGE_ESCALATION=NONE`, `M13_G17_UNIT_ENVIRONMENT_OR_CREDENTIALS=NOT_READ`, `M13_G17_UNIT_POSTGRESQL_CONNECTION=NOT_ATTEMPTED`, `M13_G17_UNIT_PRODUCTION_CHANGES=NONE`, and `M13_G17_UNIT_PRIVATE_EVIDENCE_SHA256=PASS`. Full metadata and diagnostic details remain in owner-restricted MS-02 private evidence files; no raw output was copied into GitHub, Drive, Actions artifacts, or chat.
- The one-time inspection workflow was [disabled in commit 54bb072](https://github.com/wwjd4u/MeshAlot/commit/54bb072463f257c1e1b75484244dfa82ee894b50); the [next run 38492939948 was SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38492939948). The original one-shot marker is preserved and must not be reset or reused.

## Technical interpretation

The checked-in `database/runtime-grants.sql` requires a restricted PostgreSQL role `meshalot` with LOGIN and no SUPERUSER/CREATEDB/CREATEROLE/REPLICATION/BYPASSRLS attributes; that role is distinct from a migration/owner principal. The earlier [one-shot nonprivileged PG16 diagnostic 38491337730](https://github.com/wwjd4u/MeshAlot/actions/runs/38491337730) returned `BLOCKED_DATABASE_ROLE_ABSENT` because the **SSH Linux identity's matching PostgreSQL role** was absent. That finding did **not** prove the application `meshalot` role absent.

The newly verified systemd declaration `User=meshalot` **matches the planned application-side OS identity in source** and the isolated-test example `database/setup-runtime-test.sql`. It does not certify the actual current database role, `CONNECT` permission, PostgreSQL local peer/HBA policy, service credential/DSN method, or ownership/ACL inventory. Do not infer a successful PostgreSQL authentication from the systemd metadata PASS.

## Next single proposed authorization (NOT executed)

A subsequent narrow PostgreSQL connection feasibility test under the **already existing** OS service account `meshalot` would require **fresh explicit approval** of any OS identity transition method (including `sudo -n -u meshalot` if a pre-existing restricted policy permits it). Such an approval would not authorize a new sudoers entry, password prompt, root login, administrator database identity, role creation or grants, config alteration, environment-file reads, inbound port, permanent SSH access or migration. Fail closed if the existing host-verified SSH validity margin is insufficient, if no pre-existing permitted OS transition is available, or if the existing PostgreSQL role/auth path does not accept `SELECT 1`.

A successful later `SELECT 1` only proves limited connection feasibility. Actual production PG16 cluster-global roles/memberships and schema ownership/ACL restoration would still require a separately approved read-only metadata assessment and protected recovery-asset plan. Current Caddy/TLS private storage remains **UNVERIFIED**; G17/2 full API/systemd/Caddy isolated rollback rehearsal **NOT STARTED**.

Preserve encrypted PostgreSQL/API/systemd/Caddy/environment-file backups and independent Toshiba copies, existing Node001 and Node002 identities and original Node002 M8 report, original V100 evidence, and all accepted isolated synthetic PostgreSQL/Caddy PASS results. Do not repeat completed tests.

The full October 13 conversation verbatim archive in Google Drive remains **UNVERIFIED**. This record is a **sanitized checkpoint**, not a verbatim transcript.

**Gate 17/1 IN PROGRESS. M13 OPEN. No unapproved production DB access or deployment.**
