# MeshAlot Milestone 13 — G17/1 PostgreSQL metadata retry: existing DB access blocked

Date: 2026-10-13 (America/Chicago)
Status: **BLOCKED — NOT AN INSPECTION PASS. Gate 17/1 IN PROGRESS; G17/2–G17/7 NOT STARTED.**

## Approved actions and verified results

The operator explicitly approved the **existing VM-specific SSH public-key registration** to be renewed for **no more than 60 minutes** to retry the previously authorized **read-only production PostgreSQL16 metadata-only inspection**. No new key identity, broader project IAM, operating system privilege elevation, password use, database updates or service changes were approved.

- [60-minute one-time SSH renewal](https://github.com/wwjd4u/MeshAlot/actions/runs/38468871500): **PASS**. Logs confirm the same existing local Ed25519 identity, one matching instance-level public key entry replaced with a 60-minute expiry, unchanged other instance/project metadata, private prior-registration history preserved, and no application or database changes. Renewal workflow was [disabled](https://github.com/wwjd4u/MeshAlot/commit/af8c40089961919bde76137610ccc6a6d9471836), and the [subsequent run was SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38469100778).
- [Read-only PG16 metadata retry](https://github.com/wwjd4u/MeshAlot/actions/runs/38469143375): GitHub job concluded **success** solely because the **fail-closed path executed as designed**. `M13_G17_PG_EXISTING_IAP_AUTH=VALID_WITH_MARGIN` and `M13_G17_PG_WRONG_HOST_GUARD=PASS`. The SSH route reached the expected host, but `M13_G17_PG_REMOTE_METADATA=BLOCKED_EXISTING_DB_ACCESS_UNAVAILABLE`. **No PostgreSQL catalog summary was obtained.**
- Audit safeguards reported `M13_G17_PG_SUDO_OR_PRIVILEGE_ESCALATION=NONE`, `M13_G17_PG_PASSWORDS_OR_DATABASE_ROWS=NOT_READ`, `M13_G17_PG_PRODUCTION_CHANGES=NONE` and private local audit SHA256 **PASS**. The one-time audit workflow was [disabled](https://github.com/wwjd4u/MeshAlot/commit/37da207c6cc6ee4ac45501d5553c4586fa5bc602), with [subsequent run SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38469419094).
- No access-extension job or metadata-inspection retry is authorized to keep rerunning. The existing one-hour SSH authorization must be allowed to expire normally. No further extension is inferred.

## What the blocker actually proves

The approved command attempted local socket access as the **existing nonprivileged Linux account**, with password/DSN/service-file use deliberately disabled and connection errors **suppressed**. Its single blocked classification combines possible causes such as missing/wrong PostgreSQL socket, unavailable endpoint, nonexistent matching database role, peer-auth mismatch, missing CONNECT privilege, failed SQL permissions or timeout. The current evidence **does not identify which**. It would be inaccurate to claim a confirmed authentication policy defect or a missing role from this job alone.

In PostgreSQL16, local socket access is governed by the selected `pg_hba.conf` authentication record; `peer` authentication compares the operating-system identity with the requested DB role or an approved mapping. Merely having SSH access never grants DB access. See https://www.postgresql.org/docs/16/auth-pg-hba-conf.html and https://www.postgresql.org/docs/16/client-authentication.html.

## Least-privilege unblock sequence (proposed, NOT executed or authorized)

1. **New approval for a single read-only diagnostic, if needed.** Under a currently valid verified SSH registration (do not extend it automatically), check only the PostgreSQL local socket reachability and attempt `SELECT 1` with the same existing unprivileged OS/DB identity, with no `sudo`, password, DSN, secrets, config-file reads, or database writes. Capture raw errors **only in an owner-only local file**, never Actions logs, GitHub, Drive, or chat. Print a whitelisted error **class**, not strings containing usernames, filesystem paths, passwords, or settings. No automatic retries.
2. **Classify the cause**: server/socket unavailable; PostgreSQL role absent; `peer` authentication mismatch; password or HBA authentication requirement; database `CONNECT` denied; or SQL/catalog permission denied. Fail closed on any unknown classification. Do not change `pg_hba.conf`, grant privileges, enable `trust` authentication, add accounts/roles, read credentials, or use `sudo` as part of diagnosis.
3. **Operator-approved correction only after diagnosis**: a path/port mistake can be fixed in the *audit command* without touching production. If authentication or DB privileges are required, propose a separate, tightly controlled and auditable inspection using an **existing** permitted DB/operator identity, with specified approved local privilege transition if required. Any `sudo -u postgres`, superuser use, authentication edit, new role/grant, or password handling would require a **distinct explicit authorization**, scope review, local-only evidence safeguards, and time limits. Prefer not to change production auth/roles to accommodate an audit.
4. **Acceptance**: query only approved aggregate `pg_roles`/catalog metadata inside a read-only transaction, no password-bearing `pg_authid` or customer rows; preserve private evidence and sanitized result. This would close only the narrow metadata-inventory gap—not full actual production globals backup or role-and-ACL-preserving restore. Those require separate approvals and tests.

## Hard gates

Production migrations 9–10 remain **not deployed**. No production DB, Caddy, systemd, node identities, backup archives, key vault, WAN or consumer workload changes are authorized by this document. Preserve existing encrypted backups and all prior synthetic PASS results.

**Next decision:** Operator approval of a **single zero-privilege diagnostic** that categorizes why the existing PostgreSQL socket/role cannot be used. Only then choose a least-privilege remediation. **G17/1 OPEN; M13 OPEN.**
