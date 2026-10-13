# MeshAlot M13 — G17/1 approved PostgreSQL16 read-only metadata inspection: access blocker

Date: 2026-10-13 (US Central)  
**Result: BLOCKED BEFORE PRODUCTION CONNECTION. No database role metadata was inspected.**  
**G17/1 remains IN PROGRESS. G17/2–G17/7 NOT STARTED. Milestone 13 OPEN.**

## Approval and exact scope

The operator approved a limited read-only inspection of production PostgreSQL16 role and ownership **metadata** using the **existing** IAP/SSH access route only. No password contents, database rows, new credentials, privilege escalation, production writes, restarts, schema migrations or release operations were permitted. The scope did **not** approve renewal of GCP VM SSH-key registration, new SSH identities, IAM changes or use of production backup archives.

## Evidence and verified safe stop

- [One-time approved MS-02 inspection run 38467844313](https://github.com/wwjd4u/MeshAlot/actions/runs/38467844313) concluded GitHub Actions **success**, **but the actual database inspection did not run**. This is a successful safe-stop, **not** an inspection PASS.
- The MS-02 runner verified its expected identity and checked the existing private registration's time-limited expiry **before opening any SSH connection**. The registered access was expired or did not retain the required 12-minute safety margin; the guard reported `NO_VALID_EXISTING_AUTH_MARGIN`.
- Sanitized execution markers were `M13_G17_PG_EXISTING_IAP_AUTH=BLOCKED_SAFE_STOP` and `M13_G17_PG_REMOTE_CONNECTION=NOT_ATTEMPTED`. The guard exited without executing SSH, PostgreSQL, `sudo` or any production command. No DB credential, role metadata, key contents or production records were read or sent to CI.
- No public-key registration update, privilege escalation, IAM change, database write, Caddy/API restart, new backup or node operation occurred.
- The workflow was [disabled in infra commit `0371365e3ba65afea5ebbfec7df7c7fa747eeaed`](https://github.com/wwjd4u/MeshAlot/commit/0371365e3ba65afea5ebbfec7df7c7fa747eeaed). Its [subsequent run 38467942039](https://github.com/wwjd4u/MeshAlot/actions/runs/38467942039) was **SKIPPED**. The one-time execution pathway must not be restarted or silently reenabled.

The preserved encrypted database, API/systemd/Caddy and supplemental environment-file backups were not accessed, modified or retested. Existing synthetic PG16 globals/ACL and Caddy recovery PASS records remain valid within their original restricted scope.

## Blocker and disposition

**BLOCKED — Existing IAP/SSH registration lacks the required validity margin.** This report does not determine whether the entry was already expired or simply close enough to expiry to trigger the fail-closed guard. No connection attempt was made; therefore current PostgreSQL role/ownership state, memberships, authentication settings and production globals recovery remain **UNVERIFIED**.

A retry requires **fresh, explicit operator approval** for a narrow VM-instance-only renewal of the **existing** time-limited public-key registration, and a **separate** security review if that action cannot preserve the same key, existing host verification, scope and shortest necessary expiry. Renewal is an access-control **write** and was not covered by approval to read metadata.

After any approval, check current authorization/registration policy and do not assume renewal is safe from prior runs. Use strictly host-key-verified IAP and only the approved nonprivileged current Linux identity; if the database catalog cannot be queried without `sudo`, secret DSNs, passwords or privilege elevation, stop and report **BLOCKED_DB_CATALOG_PERMISSION** rather than escalating. Full production globals backup/restoration needs separate approval regardless of read-only inspection outcome.

## Current gate outcome

- `G17/1`: **IN PROGRESS — real PG role/ownership metadata UNVERIFIED**.
- `G17/2`: **NOT STARTED** (full API/systemd/Caddy isolated rollback).
- `G17/3–7`: **NOT STARTED**.
- `PRODUCTION`: **UNMODIFIED by this attempt**.
- `NEXT ACTION`: Ask operator whether to approve only the narrow renewal of the pre-existing instance-specific time-limited SSH public-key registration, **without** new identity, broader privilege or other changes. No automatic retry.
