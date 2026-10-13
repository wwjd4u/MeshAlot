# MeshAlot M13 — Gate 17/1 PostgreSQL 16 real-global-role recovery design

Date: 2026-10-13 (America/Chicago)  
Status: **PLAN RECORDED; REAL PRODUCTION GLOBALS/OWNERSHIP RECOVERY UNVERIFIED; G17/1 STILL OPEN**  
Scope: **documentation only**. This commit does not authorize production database access, elevated privileges, secret reads, archive use, restore testing, service changes, or a release.

## 1. Evidence already accepted — do not repeat

- Production uses **PostgreSQL major version 16**, with tracked schema migrations **1–8** applied. M13 migrations **9–10 are not in production**. Production `main` at the preflight was `1016c9b058690f34058dfab9cde7927e421fa469`; the intended M13 application source evidence is `4f67ee54e5ebca084e38e306e723ae0acd08e77d`. Revalidate both before any later cutover.
- The operator-controlled encrypted PostgreSQL logical archive, a checksum-matching independent copy on the encrypted Toshiba APFS volume, and a **prior isolated logical restore** are recorded PASS in [Gate 17 recovery readiness](m13-gate17-backup-restore-rollback-readiness-2026-10-11.md). This is *not* proof of actual cluster-wide roles or authentication restoration.
- A [synthetic PostgreSQL16 runtime permissions test](https://github.com/wwjd4u/MeshAlot/actions/runs/38426678379) already PASSed; eight disallowed actions returned SQLSTATE 42501.
- A [synthetic PostgreSQL16 global-role and ownership rebuild](https://github.com/wwjd4u/MeshAlot/actions/runs/38440768518) already PASSed using a **fresh disposable destination cluster**, a password-free global-role export, and a synthetic logical backup. Source and destination containers were cleaned, and the one-time workflow was disabled with a [subsequent SKIPPED run](https://github.com/wwjd4u/MeshAlot/actions/runs/38440900971). This validates the *method with synthetic roles*, **not the actual production role catalog**.
- Existing encrypted API/systemd/Caddy rollback archive and its separate encrypted environment-file supplement have off-host SHA256-verified copies. Neither is evidence of a PostgreSQL cluster-global backup.
- Do not use the generic `deployments/docker-compose.yml` PG17 image for this PG16 recovery procedure.

## 2. The actual gap

A database-level `pg_dump` archive does **not** include cluster-global objects such as PostgreSQL roles, memberships, tablespaces or global configuration-parameter privileges. `pg_restore` by default attempts to restore database-object owners and grants, but only if the destination has appropriate roles and restore privileges. A restore made with `--no-owner` / `--no-acl` can demonstrate data and schema recovery only; it is **not** proof of production identity and permission recovery.

The following remain **UNVERIFIED for production**:
- Actual cluster role definitions, authentication state, role memberships/inheritance and privileged attributes; separate migration-owner identity; ownership of database, schemas, tables, functions, security-definer functions, migration ledger and sequences.
- Whether the production cluster uses custom tablespaces, extensions, non-default global configuration grants or role-specific defaults needed by the API.
- A complete independently encrypted and restore-tested **production** globals/ownership recovery artifact and its protected off-host copy. **Do not claim it already exists.**
- Runtime role service authentication and compatibility with the recovered PostgreSQL catalog without using any actual production credential or enrolled node identity in shared staging.

## 3. Planned operator-controlled source inventory — NOT YET AUTHORIZED

If the operator separately approves a bounded, read-only **production database metadata** assessment, use an existing authorized, time-limited, host-verified route only; no credential renewal, privilege escalation, writes, exports of passwords or database rows. Confirm the approved read-only principal can query the necessary metadata **without** accessing `pg_authid` or other password/hash-bearing values.

- Collect *privately and minimally* role count, `pg_roles` nonsecret attribute classes, memberships/dependencies, ownership relationship classes, tablespace and extension requirements, database/template properties, and runtime-to-owner separation. `pg_roles` intentionally masks password values.
- Never copy real role names beyond necessary pre-approved classification, raw SQL rows, connection strings, usernames/password hashes, `pg_hba.conf`, secret paths or configuration contents into public GitHub, Actions logs/artifacts, Drive or chat. Keep the detailed read-only audit in the operator-controlled, access-restricted local evidence store; publish only an aggregate PASS/BLOCKED matrix and its hash.
- Record whether a read-only inventory has *sufficient visibility*. Missing access means **BLOCKED**, not missing roles and not permission to escalate.
- Cross-check against the pinned source: `database/migrate.py`, `database/runtime-grants.sql`, and the [dependency inventory](m13-g17-g1-postgres-caddy-dependency-inventory-2026-10-13.md). Distinguish **actual production v8** from **synthetic staged v10**; no direct 9/10 production migration is authorized.

## 4. Protected recovery-asset design — requires separate authorization

For an approved recovery window, a trained operator with appropriately authorized database privileges may capture a PG16 `pg_dumpall --globals-only --no-role-passwords`-style globals artifact, or a more narrowly reviewed equivalent. Such a dump can contain operationally sensitive role, membership and tablespace details even when password hashes are excluded. Treat it as **private recovery material**:

1. Review exact scope, source privilege requirements and expected output **before running**. Do not connect to the production cluster from a shared runner merely because this runbook exists; do not expose SQL output to terminal transcripts or CI.
2. Encrypt at source to operator-controlled restricted storage, with secure handling of intermediate buffers, integrity digest, and a separately verified independent **encrypted** copy. Reuse or create a recovery key *only under an explicit key-custody approval*; keep vault-held keys separate from archives and never place them in GitHub, Drive, Actions secrets/logs/artifacts or chat.
3. A `--no-role-passwords` restore yields **NULL role passwords**, so separately provision fresh credentials under the operator's private policy and verify the intended auth method. Do not expect restored password authentication to work until that happens. Do not store any password or hash in the role artifact.
4. Verify the artifact securely **without replacing** the existing database dump, Caddy/rollback archives, Toshiba copies, Time Machine data, original operator safenotes or either enrolled provider identity.
5. If source privilege, ownership or tablespace dependencies require an unapproved escalation or an unsafe output channel, **STOP**. A plan is not permission to use those privileges.

**Current status: NOT EXECUTED.** No real globals dump is known to have been captured or verified.

## 5. Future isolated restore sequence — design, NOT execution approval

Only after separately approved stage scope and recovery-asset/key use:

1. Establish a **fresh isolated PG16** target with the already accepted MS-02 synthetic-stage safeguards: sequential resource-capped disposable test environment; no production route, host bind mounts, inbound ports, privileged host path, real sessions or identities. Review the new scope if a real production-derived archive is involved; existing *synthetic-only* staging approval does **not** extend to it.
2. Independently verify ciphertext/hash against the private manifest **on receipt**. Examine recovered SQL for potentially unsafe or unexpected commands *within approved private handling*. An SQL dump/restore can execute code; never replay unreviewed source SQL against an operating cluster.
3. Reconstruct required **cluster-global roles, memberships, tablespaces and global configuration grants** *before* an ownership/ACL-preserving database restore. Resolve bootstrap role collisions explicitly; do not blindly use `--clean`, `DROP ROLE`, `REASSIGN OWNED`, `pg_restore --clean` or superuser login in production.
4. Keep the schema/migration owner distinct from the restricted `meshalot` runtime LOGIN role. The runtime role must not be SUPERUSER, CREATEDB, CREATEROLE, REPLICATION or BYPASSRLS and must not gain those capabilities indirectly through unsafe memberships. Validate role grant inheritance and database/schema/table/function ownership.
5. In a private **operator-approved** recovery stage only, restore the existing logical archive with its intended ownership and privileges, resolving extensions/tablespaces first. Do not relabel a `--no-owner --no-acl` test as complete recovery. Avoid exposing actual dataset content or any production credential to shared CI/logs. Prefer synthetic-only fixtures unless a specifically controlled production-derived-data stage has been separately authorized.
6. Verify migrated schema history and SHA256 ledger without changing it: **source production 1–8**, not 9–10. Test trusted API database access with fresh **stage-only** credentials, authorized column/function access and explicit expected-denied operations. Confirm database, ledger, schema, owner, extension and grant fidelity and no unexpected administrator membership; preserve test evidence with no production rows.
7. Stage-only API/systemd/Caddy startup, negative authentication, HTTPS/WSS upgrade, intentional failure and rollback belong to **G17/2** and require separate scope approval. Do not mistake the already-passed isolated synthetic Caddy test for a complete archived service rollback.
8. Prove stage resource deletion, no network/secret leakage, immutable source archives and independent operator evidence review. A test outcome cannot be **PASS** if privileged access, owner/ACL fidelity, stage isolation, or cleanup is unknown.

**Fallback if actual globals were never captured:** manual role/ownership recreation requires a separately operator-verified private inventory and an isolated acceptance test. Re-running `pg_restore --no-owner --no-acl` is not a substitute. Keep the recovery item BLOCKED until comparable evidence exists.

## 6. Evidence matrix / stop rules

| Item | Current evidence | Status |
| --- | --- | --- |
| Encrypted logical DB, independent ciphertext copy and isolated database-only restore | Previously audited | **PASS — limited to database-level restore** |
| Synthetic PG16 roles/grants, forbidden operations and new-cluster global rebuild | Isolated accepted jobs | **PASS — synthetic only** |
| Real production global-role/owner/member/auth dependency inventory | No approved production DB metadata inspection here | **OPEN / UNVERIFIED** |
| Approved, protected PG16 globals/role recovery asset and independent copy | None confirmed | **OPEN / NOT CAPTURED** |
| Real role-and-ACL-preserving restore into approved private isolated stage | Not performed | **OPEN / NOT TESTED** |
| Full API/systemd/Caddy restore and fault rollback | G17/2 not started | **OPEN** |

**Immediate stop:** missing necessary privileges, changed production version or migration ledger, unexpected privileged memberships, role/ownership mismatch, unknown tablespace/extension, archive digest difference, production network route, exposed credential/hash/private data, or a required production change without a new explicit authorization.

### The next single checkpoint

Present a **narrow read-only production PostgreSQL16 role/ownership *metadata-only* scope** for operator approval. Do not start an inspection before approval. Separately request approval later for any actual globals capture and later still for use of real recovery archives/keys in an isolated private restore. Keep production untouched during planning.

**G17/1 IN PROGRESS. G17/2–7 NOT STARTED. M13 OPEN.** This document does not close Gate 17.

## References

- [Gate 17 fixed completion sequence](m13-gate17-fixed-completion-sequence-2026-10-13.md)
- [G17/1 PostgreSQL/Caddy dependency inventory](m13-g17-g1-postgres-caddy-dependency-inventory-2026-10-13.md)
- [PostgreSQL 16 — pg_dumpall](https://www.postgresql.org/docs/16/app-pg-dumpall.html): cluster globals, password-free export semantics and restore precautions.
- [PostgreSQL 16 — pg_dump](https://www.postgresql.org/docs/16/app-pgdump.html): single-database boundaries.
- [PostgreSQL 16 — pg_restore](https://www.postgresql.org/docs/16/app-pgrestore.html): ownership/ACL restoration and trust warning.
- [PostgreSQL 16 — pg_roles view](https://www.postgresql.org/docs/16/view-pg-roles.html): non-password role metadata.
