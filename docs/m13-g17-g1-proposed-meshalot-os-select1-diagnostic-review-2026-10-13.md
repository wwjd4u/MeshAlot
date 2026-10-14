# MeshAlot — M13 G17/1: proposed existing-service-identity PostgreSQL connection feasibility diagnostic

Prepared: October 13, 2026 (America/Chicago)  
Status: **DRAFT FOR OPERATOR REVIEW — NOT APPROVED TO EXECUTE / NOT EXECUTED**  
Milestone: **M13 OPEN**. Gate 17/1: **IN PROGRESS**. Gate 17/2–17/7: **NOT STARTED**.

## Approval boundary

Operator response **"yes"** approved *preparation of a narrow diagnostic for review only*. This does **not** approve use of `sudo`, changing Linux identity, a PostgreSQL connection, SSH extension, privilege or database modification, any new workflow execution, or privileged service reading. No permanent access, new accounts, ports, keys, role grants, auth-policy edits, credential reads, or deployment is proposed.

This is a **non-executable review document**, not a GitHub Actions workflow or executable production script. No production access was made to create it.

## Grounded existing evidence

1. [One-time PostgreSQL connection diagnostic 38491337730](https://github.com/wwjd4u/MeshAlot/actions/runs/38491337730): `BLOCKED_DATABASE_ROLE_ABSENT` because the SSH Linux account's corresponding PostgreSQL role was not present. No successful `SELECT 1`; does **not** establish that MeshAlot's planned application role is absent. The one-time workflow was disabled and follow-up SKIPPED.
2. [One-time systemd metadata inspection 38492890931](https://github.com/wwjd4u/MeshAlot/actions/runs/38492890931): the active production API systemd unit explicitly declares `User=meshalot`, `DynamicUser=no`. The systemd declaration does not prove a PostgreSQL role, effective process UID, authentication mapping, or DB grants. No database query or sudo was performed. Its one-time workflow was disabled and follow-up SKIPPED.
3. [Runtime grants](../database/runtime-grants.sql): `meshalot` is intended to be a restricted non-owner LOGIN DB role without administrative attributes; distinct migration owner. This is an intended source contract, **not verified production state**.
4. [Persistence guidance](postgres-persistence.md): production application DSN is private and must not be read or copied into logs, GitHub or chat. The isolated test uses a `meshalot` Linux account/role peer-auth scenario, but production peer/HBA behavior remains unverified.

## Proposed *future* single attempt — requires separate approval

**Purpose:** Test whether the existing `meshalot` Linux service identity, if accessible through an already-permitted noninteractive privilege transition, can authenticate as the existing PostgreSQL `meshalot` role over the **local Unix socket** for the sole SQL statement `SELECT 1`. The goal is limited connection feasibility, **not** a catalog, owner, role membership, credentials, customer-data or recovery inspection.

**Hard preconditions (must all pass before any remote connection):**

- Fresh explicit operator approval for **one** narrowly scoped `sudo -n -u meshalot -- /usr/bin/psql …` attempt, acknowledging this is an OS identity transition, even though it is not a root login.
- Use only the original pinned-host-key Google IAP SSH route to `meshalot-control-01` via `meshalot-ms02` with the previously registered exact Ed25519 SSH identity. Require the VM-scoped SSH registration to have **more than 12 minutes remaining**. If expired or insufficient, **stop without SSH** and request **separate** approval for any extension.
- Current runner, Linux account, host, project, zone, private key/known-hosts ownership and modes must match previously verified expectations. No new SSH identity, project-wide key, IAM expansion, SSH config bypass or incoming port.
- The existing SSH account must be non-root. The production systemd evidence remains the reference for `User=meshalot`; no new service-file or EnvironmentFile reads.
- Do not modify or install sudoers policy. If noninteractive `sudo -n -u meshalot` with this **exact invocation** is not already permitted, fail closed; no password prompt and no alternate `sudo` command or retry.

**Candidate single SQL command, for review only; DO NOT EXECUTE from this document:**

```bash
/usr/bin/env -i PATH=/usr/bin:/bin HOME=/nonexistent LC_ALL=C \
  /usr/bin/sudo -n -u meshalot -- \
  /usr/bin/psql -X -w -qAt -v ON_ERROR_STOP=1 \
    --dbname='host=/var/run/postgresql dbname=meshalot user=meshalot passfile=/dev/null connect_timeout=5 sslmode=disable' \
    -c 'SELECT 1'
```

Why this command is bounded:

- `env -i` clears inherited environment including `PGPASSWORD`, `PGSERVICE` and other potentially sensitive libpq environment inputs **before invoking sudo**. The later implementation must separately verify the actual sudo environment policy, not assume it.
- `sudo -n` does not ask for a password. If exact, existing permission is unavailable, the operation stops with **BLOCKED_SUDO_TRANSITION_NOT_PREAUTHORIZED**. No `sudoers` file is viewed or changed.
- `-X` avoids loading user `.psqlrc`. `-w` disables password prompts **but by itself does not prevent** a stored password-file lookup. Explicit libpq `passfile=/dev/null` prevents use of `~/.pgpass` as a fallback. Do **not** read credential files or allow environment-sourced DSNs. PostgreSQL 16 docs: [libpq connection parameters](https://www.postgresql.org/docs/16/libpq-connect.html), [psql](https://www.postgresql.org/docs/16/app-psql.html).
- The explicit socket-directory `host=/var/run/postgresql` prevents connecting to a TCP endpoint; `sslmode=disable` is scoped to this Unix socket, not public TLS.
- Database/user are set explicitly to `meshalot`, unlike the previous SSH-user-to-PG-user mismatch. No DB administrative identity is requested.
- The sole SQL text is `SELECT 1`. No role listing, `pg_authid`, permissions enumeration, `pg_hba.conf`, production rows, backup, configuration change or migration.

**When later implemented as a separately reviewed one-time job:**

- Keep it *disabled* at job/dispatch level until approval; ensure no schedule or self-retry.
- Validate Bash syntax, embedded Python syntax and exact source hash **offline before** any production access. Do not rerun already accepted migrations, ACL, globals, V100, Caddy or service tests.
- Claim a **fresh exclusive, persistent 0600 one-attempt marker** in a 0700 owner-controlled MS-02 directory (never reset an old marker). No second execution on fail or Actions rerun.
- Recheck the existing SSH validity margin before launching SSH. Use strict pinned host-key validation, exact target and a remote wrong-host guard; non-root process only. SSH must be outbound via the existing IAP route, no inbound port.
- Capture only `stdout == "1"` as connection success. Collect raw `stdout/stderr` **only in owner-only MS-02 local audit files**; never expose raw errors, usernames, role membership, environment data or credentials to Actions logs/artifacts, the public repository, Drive or ChatGPT. Record a local SHA-256 sidecar.
- Emit only fixed allowlisted classifications: `CONNECTED_SELECT1`, `BLOCKED_SUDO_TRANSITION_NOT_PREAUTHORIZED`, `BLOCKED_PG_ROLE_ABSENT`, `BLOCKED_PEER_AUTH`, `BLOCKED_AUTH_REQUIRED`, `BLOCKED_DB_CONNECT_DENIED`, `BLOCKED_SOCKET_UNAVAILABLE`, `BLOCKED_SSH_EXPIRED_OR_INVALID`, `BLOCKED_UNEXPECTED`. A successful GitHub job **must not be conflated with successful database access**.
- Fail closed on any mismatched identity, host, expiry, missing/surprising binary, unsupported auth method, unknown response, or insufficiency of existing sudo permissions. No credential fallback, privilege escalation, new role, grants, `pg_hba.conf` changes or automatic retries.
- After exactly one authorized attempt, disable the one-time workflow and verify a SKIPPED follow-up; preserve sanitized GitHub evidence and update Google Drive checkpoint when allowed.

## Decision required after reviewing this plan

**A.** Approve only a newly prepared and offline-validated one-time diagnostic with exact bounded pre-existing noninteractive transition `sudo -n -u meshalot -- /usr/bin/psql …` and **only** `SELECT 1`. This approval must be explicit and separate from SSH renewal.  
**B.** If the pre-existing SSH registration has expired, a *further separate* approval is required to renew the **existing** VM-only key with a time-limited expiry. Never infer a renewal from diagnostic approval.  
**C.** If `sudo` is not already authorized or passwordless peer auth is unavailable, STOP and request a separately scoped operator-assisted access decision. Do not add rights just to pass an audit.

**No action A or B has been approved by the current preparation-only instruction. Nothing in this proposal is an executable approval.**

A successful future connection would establish only that one nonsecret read-only query was accepted under the existing `meshalot` identity. Production PG16 global roles/memberships, schema owner and ACL inventory, protected globals recovery asset, Caddy/TLS storage and full stage rollback remain separate gate requirements.

Preserve Node001 and Node002 identities, original Node002 M8 report, encrypted backups and off-machine Toshiba copies, and historical V100 evidence. G17/1 **IN PROGRESS**; G17/2–7 **NOT STARTED**; M13 **OPEN**. This is a technical review document, **not a verbatim transcript**, and the October 13 complete verbatim Drive archive remains **UNVERIFIED**.
