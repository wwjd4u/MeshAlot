# MeshAlot M13 — Gate 17/1: Existing PostgreSQL principal and authentication-route source review

Date: 2026-10-13 (America/Chicago)
Scope: **READ-ONLY GitHub source / governing-guide review; no remote access or production execution.**
Status: **SOURCE-LEVEL CLASSIFICATION PASS; production auth/roles and global recovery UNVERIFIED.**
Gate 17/1: **IN PROGRESS / BLOCKED**. Gate 17/2–7: **NOT STARTED**. Milestone 13: **OPEN / NOT RELEASE-READY**.

## Current stop point (already verified; do not rerun)

The single approved [read-only connection diagnostic](https://github.com/wwjd4u/MeshAlot/actions/runs/38491337730) under the existing unprivileged SSH Linux identity returned `BLOCKED_DATABASE_ROLE_ABSENT`. That classification establishes that the specific PostgreSQL role requested using the SSH account's Linux username was unavailable. It **does not establish** that the planned `meshalot` application role is absent or that the database, owner, peer map or CONNECT grant is valid/invalid. `SELECT 1` did **not** succeed. Detailed errors remain locally protected.

The [approved 60-minute instance-only SSH public-key renewal](https://github.com/wwjd4u/MeshAlot/actions/runs/38491246917) passed and was disabled after its one-time use. The diagnostic one-shot workflow was also disabled and its post-disable run SKIPPED. Do not assume time-limited authorization is presently usable; no further production access or renewal is authorized by this source review.

## Source evidence — what it establishes and what it does not

| Checked-in evidence | Established source fact | Not proven in production |
| --- | --- | --- |
| [`database/runtime-grants.sql`](../database/runtime-grants.sql) | Intended application role `meshalot` is LOGIN, non-SUPERUSER, without CREATEDB, CREATEROLE, REPLICATION or BYPASSRLS; app table/column/function grants are narrow | Whether that role currently exists on the actual PG16 cluster, has valid CONNECT, or has any approved auth path |
| [`docs/postgres-persistence.md`](postgres-persistence.md) | Application uses a private `MESHALOT_DATABASE_DSN` when in PostgreSQL mode; application role is distinct from migration owner; DSN/token values must never enter GitHub/CI/chat | Actual production DSN contents, OS service identity, connection method, or ability of an audit process to authenticate |
| [`database/setup-runtime-test.sql`](../database/setup-runtime-test.sql) | **Isolated test only** creates `meshalot` login with Unix-socket peer authentication for **OS user `meshalot`**; deliberately does not alter global pg_hba | Does not certify a corresponding production OS account, role, pg_hba rule, or peer mapping |
| [`database/migrate.py`](../database/migrate.py) and [`database/README.md`](../database/README.md) | Trusted migrations are run using a separate migration/owner identity; checksums and migration ledger are protected from runtime access | Actual production owner role name, role memberships/ownership, current owner authentication |
| [historical read-only audit script](../deployments/google/m13-gate17-step2-readonly-audit.sh) | Previous, separately scoped code proposed `sudo -n -u postgres psql` for a migration-ledger query | Does **not** authorize sudo for G17/1, prove the privilege is presently configured, or justify using the postgres administrator identity |
| [generic docker-compose](../deployments/docker-compose.yml) | Development compose config sets up PostgreSQL 17 | **Not** evidence of actual production PG16 authorization/role provisioning; do not run this PG17 fixture against production recovery |

Read-only branch reference check: production `main` at `1016c9b058690f34058dfab9cde7927e421fa469`; M13 branch `m13-persistent-heartbeat` current. Existing production migration history recorded at 1–8, with 9–10 NOT DEPLOYED. Already-PASSed synthetic role, ACL and recovery tests remain accepted; do not repeat them.

The Google Drive governing **AI Mesh POC Build, Test, and Release Guide**, Milestone 13, requires outbound TLS, persistent WSS/heartbeat, stale-node offline detection and reconnect across interruption/reboot. This source inventory is a **recoverability prerequisite**, not Milestone 13 functional acceptance.

## Reasoned conclusion

1. The prior blocker was caused by trying to use the *SSH Linux identity* as a PostgreSQL role. It is a client principal selection mistake for that specific diagnostic, **not evidence of a missing production application role**.
2. The intended restricted role from the public source is `meshalot`. In test, peer authentication works only when the OS and database identities are aligned; this must **not** be extrapolated to production.
3. Authentication/ownership details in protected production environment files or PostgreSQL host-based authentication records must stay private. No credential, password, secret DSN, role password/hash, private SSH key, or protected configuration contents may be sent through CI, GitHub, Google Drive or chat.
4. The former script's use of `sudo -n -u postgres` for a different approved audit is **not current authorization**. No sudo or administrator database identity may be used absent a distinct, time-bounded operator approval.

## Proposed *next single decision* — NOT EXECUTED OR AUTHORIZED

Before another database connection, request approval for a **one-time production read-only identity/authentication-path metadata review** under a *currently valid* instance-specific, pinned-host-key IAP/SSH route. Use the same nonprivileged SSH account; first confirm existing access expiration has the required margin. Do **not** renew or extend SSH automatically.

Review only bounded, nonsecret public/systemd metadata identifying which **OS user** runs the existing production MeshAlot API, without printing any EnvironmentFile contents, ExecStart arguments, DSNs or secrets. Record only allowlisted classes in owner-only local evidence; if the permitted nonprivileged service metadata cannot establish the identity, **STOP** rather than using sudo, changing ACLs or reading protected files. Do not read `pg_hba.conf` in this phase, open a PostgreSQL connection or use the app's credential.

**After** that metadata-only finding, prepare a separate explicit approval for one suitable nonprivileged `SELECT 1` connection feasibility test as an existing DB role if its authorized local authentication path has been established. An unknown pg_hba/auth method or missing legitimate mapping is BLOCKED, not permission to use trust auth, add database roles, read private DSNs or escalate. A successful connection alone is NOT full production global-role/ownership verification; approval for catalog metadata inventory/cluster-global capture is distinct.

Alternative if no nonprivileged route exists: arrange separately approved, operator-supervised and time-limited inspection with a pre-existing permitted operator identity, with reviewed sudo/auth scope if needed. Prefer no production permission or authentication edits.

## Preservation and acceptance boundary

- No new SSH key/OS login/database role, no `pg_hba.conf` edit, no trust authentication, no privilege expansion, no incoming ports, no Tailscale, no backdoors.
- No production migration, binary/Caddy/systemd edit, live restart, backup decryption or node enrollment changes.
- Keep all encrypted database, Caddy/API and environment-file rollback archives, off-machine Toshiba copies, Node001/Node002 identities, original Node002 M8 report and V100 benchmark evidence intact.
- Production real global roles/memberships/ownership and ACL-preserving restore remain **UNVERIFIED**; actual Caddy/TLS storage is **UNVERIFIED**; G17/2 full isolated API/systemd/Caddy rollback remains **NOT STARTED**.
- Documentation created here is a sanitized technical source review, **not a verbatim transcript**. The October 13 verbatim Drive archive remains **UNVERIFIED**.

**Next gate:** request narrowly scoped operator approval before any new remote production read. G17/1 remains OPEN.
