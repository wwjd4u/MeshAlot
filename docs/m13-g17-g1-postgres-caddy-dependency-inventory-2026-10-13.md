# MeshAlot M13 — Gate 17 / G17-1: PostgreSQL and Caddy dependency inventory
Date: 2026-10-13 (US Central)
Status: **STATIC SOURCE INVENTORY COMPLETE; LIVE RUNTIME AND RECOVERY VALIDATION PENDING**
Governing: AI Mesh POC Build, Test, and Release Guide, Milestone 13; fixed seven-checkpoint M13 Gate 17 plan.

## Scope and evidence

This record catalogs recovery dependencies using only previously archived Gate 17
evidence and committed source files on `m13-persistent-heartbeat`.
**No production SSH session, database connection, private audit report read,
archive decryption, secret-value read, service restart, Docker container creation,
Node001/Node002 change, IAM/SSH update, or V100 benchmark was performed.**

- [Gate 17 readiness log](m13-gate17-backup-restore-rollback-readiness-2026-10-11.md):
  production PostgreSQL major 16, migrations 1–8; two later migrations remain
  unapplied in production. Prior isolated logical restoration of the encrypted
  database archive PASS; no cluster-global role/ACL restoration proven.
- [Fixed Gate 17 checklist](m13-gate17-fixed-completion-sequence-2026-10-13.md):
  G17/1 recovery dependencies before G17/2 isolated rollback rehearsal.
- [Migration runner](../database/migrate.py):
  migration execution requires administrator/owner role, not runtime role;
  ledger `meshalot_meta.schema_migrations` preserves version names and SHA-256;
  reject drift or gaps. No force-history change.
- [Runtime grants SQL](../database/runtime-grants.sql):
  restricted `meshalot` LOGIN role must **NOT** have SUPERUSER,
  CREATEDB, CREATEROLE, REPLICATION or BYPASSRLS. It must be separate
  from the schema migration owner. Its table/column/function grants
  are deliberately narrow.
- [PostgreSQL persistence guide](postgres-persistence.md):
  runtime role is non-owner and non-superuser; secret DSN stays private;
  ordinary logical DB content does not reconstruct cluster-global roles.
- [Existing synthetic isolated v8→v10 result](https://github.com/wwjd4u/MeshAlot/actions/runs/38166687485):
  `postgres:16-alpine`, Docker `--network=none`,
  temporary PostgreSQL 16 database. Previously passed; do **not**
  repeat simply to reconfirm.
- [General docker-compose](../deployments/docker-compose.yml) currently
  specifies `postgres:17`. **Do not use this generic file for
  M13 recovery staging**, because the actual production database
  and restore artifact are PostgreSQL major **16**.
  Use the previously validated PG16 staging image path.
- [M13 reverse proxy isolated test](../server/m13_reverse_proxy_postgres_test.go):
  simulated Caddy-shaped TLS termination/loopback reverse proxy test
  is useful prior evidence, but does **not** demonstrate restoration
  of real production Caddy modules, certificate material or storage.
- [Step 3E.1 read-only inspection](https://github.com/wwjd4u/MeshAlot/actions/runs/38295531951):
  primary Caddyfile readable, bounded import scan found 0 import
  directives and 0 parse faults. That is limited to the examined
  primary file; alternate configuration sources and runtime/TLS
  dependencies are not independently certified.
- [Step 3E.5 environment metadata](https://github.com/wwjd4u/MeshAlot/actions/runs/38309656972):
  two systemd EnvironmentFiles declarations resolve to regular files
  (not necessarily two distinct underlying files). Existing supplemental
  AES256 GPG archive is now copied to Toshiba with operator SHA-256
  PASS; its values remain protected and are not reproduced here.

## Dependency classification table

| Recovery prerequisite | Evidence today | Unfinished acceptance condition |
| --- | --- | --- |
| PostgreSQL 16 engine/image | PG16 production, restore and isolated synthetic PG16 v8→v10 test PASS | Pin PG16 for new *authorized* staging; no generic compose PG17 |
| Runtime role `meshalot` | Declarative code enforces LOGIN and five no-admin flags | Verify actual restored-stage role properties without using production secrets |
| Schema/database ownership and migration administrator | Migration runner expects separate owner; no test proves full role ownership recreation | Document role provisioning and ownership/GRANT procedure for stage; verify actual global roles and grants safely |
| Runtime table, column and function permissions | `runtime-grants.sql` source lists intentional narrow grants | Test restricted and forbidden operations using only synthetic stage credentials |
| Migration order | Production 1–8; isolated 8→9→10 passed | Apply runtime grants for M13 **only after** staging migrations 9 and 10, because grant source references `node_ratings` and `node_status.m13_telemetry` |
| PostgreSQL custom logical backup | Encrypted/verified, isolated restore previously passed | Logical restore alone omits cluster-global role definitions/memberships; establish approved role recovery method |
| Primary Caddyfile | Nine-item encrypted rollback package hash verified; zero static imports observed | Test actual pinned Caddy version/adapters, config-validation and loopback/WSS behavior in isolated stage |
| TLS/certificate storage | No independent recovery proof | Identify cert issuer/storage, cert/key recovery policy and file ownership **without exporting private TLS keys into CI** |
| Caddy runtime plugins/config sources | Not fully inventoried | Identify required runtime modules and alternate config paths; retain blocker if unknown |
| Systemd user/working directory/unit and drop-ins | Archived unit + six drop-ins, hash-pinned | Validate ownership, mode, ordering, environment-file dependencies and clean restart in isolated stage |
| Two environment-file entries | Source metadata PASS; separate encrypted supplemental archive, off-machine checksum PASS | Validate recovery with the existing RoboForm API rollback key in an explicitly approved isolated test; never log secrets |
| Recovery-key custody | Two operator-verified RoboForm Safenotes | Test vault-sourced recovery deliberately, offline, with separate approval; no key travels with Toshiba |
| Existing provider identities, history and V100 evidence | Preserved; unrelated | No re-enrollment, identity regeneration, benchmark retest, or unapproved provider controls |

## Required safe test order in the future approved isolated environment

1. Approve target and safety scope, including no privileged host mounts,
   inbound provider ports, production network access, or customer credentials.
2. Pin PostgreSQL16 from existing validated local image. Use only synthetic
   database fixtures and synthetic stage-only roles/DSNs at first.
3. Install schema migrations 1–8 in stage, then verify the prior tested
   8→9→10 additive sequence **only if not duplicating a previously accepted
   identical rehearsal**; focus new testing on owners, global roles, ACLs and
   service runtime rather than repeating the original migration test.
4. Verify stage migration administrator is distinct from runtime user;
   stage `meshalot` role must have no unsafe cluster privileges.
   Apply the tracked runtime-grants script only in a disposable schema
   where its v9/v10 referenced tables and columns exist.
5. With separate operator approval, use the already encrypted known-good
   API/systemd/Caddy archive and supplemental systemd environment archive
   in a private disposable staging target, not on production or as plaintext
   files in GitHub Actions logs. Sensitive recovery keys stay operator-controlled.
6. Validate real Caddy version, routing and WSS upgrade in isolated stage,
   using stage-owned test certificates and no public DNS/network changes.
7. Rehearse a stage-only failure and rollback. Keep synthetic DB and logs
   free of real node IDs, accounts or API keys. Prove safe return to
   known-good state before any G17/2 PASS.
8. Explicitly record remaining dependencies as BLOCKED if real PG roles,
   TLS/storage or access control cannot be reproduced without production
   privileges. Do not invent coverage.

## Decisions and boundaries

**This step is PASS for code-level dependency identification only.**
It is **NOT** live database role/ACL verification, actual Caddy/TLS
recovery, a staging run, a production test, or a full rollback rehearsal.
G17/1 still requires a stage-target approval and actual dependency
recovery acceptance; G17/2 is NOT STARTED; M13 Gate 17 remains OPEN.

**Next single action:** operator approve a disposable, resource-capped,
network-isolated PG16/synthetic stage on existing MS-02, with no
production-derived secrets, no re-enrollment and no persistent agent.
Until that approval, do not create any container or staging data.

## Addendum — 2026-10-13 approved isolated PG16 bootstrap

With the operator's explicit approval, an ephemeral PG16-only staging
container was exercised on the existing MS-02. [Successful bootstrap run](https://github.com/wwjd4u/MeshAlot/actions/runs/38419118910)
established a restricted 2-CPU/4-GiB, network-none, cap-dropped,
read-only-root container with temporary in-memory PostgreSQL16
data and only synthetic test records. Restricted synthetic
owner and runtime roles were checked for nonadministrative flags.
The failed first inspector run [38418971253](https://github.com/wwjd4u/MeshAlot/actions/runs/38418971253)
also cleaned successfully; the final run's explicit cleanup passed.
No production source credentials, backup archives, V100 evidence,
node identities or plaintext secrets were used. [Rerun skipped](https://github.com/wwjd4u/MeshAlot/actions/runs/38419245216)
after disabling the one-time workflow.

**What this resolves:** operator staging-target choice,
container-level isolation/resources, local PG16 availability,
synthetic fixture and elementary restricted-role policy test.
**Still not proved:** actual tracked M13 schema-grant behavior;
PostgreSQL roles/owner/ACL reconstruction from a backup;
application/Caddy/TLS restoration; any use of the encrypted
production-derived archives; and real network/reboot acceptance.
Stage resources were deleted after the approved test and are
not still running. G17/1 remains IN PROGRESS; Gate 17 OPEN.
