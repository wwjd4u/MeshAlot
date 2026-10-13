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

## Addendum — synthetic PostgreSQL16 owner and globals recovery (2026-10-13)

[Isolated MS-02 recovery run](https://github.com/wwjd4u/MeshAlot/actions/runs/38440768518) PASS; staged PostgreSQL16 source and
fresh independent destination ran **sequentially**, each with 2-CPU/4-GiB cap,
Docker network none, no ports/binds, dropped capabilities and disposable
in-memory data. The pinned migration source set up fixtures owned by a
separate, nonadmin synthetic `meshalot_stage_owner` role. Runtime grants
were applied to restricted `meshalot`. Password-free `pg_dumpall` global
role definitions plus a synthetic custom-format database dump were restored
to the fresh cluster **after deleting the source cluster**.

Actual test PASS: nonadmin global role attributes; distinct migration versus
runtime role without administrative memberships; database and public table
ownership; security-definer function and migration ledger ownership; all 10
migration records; synthetic record preservation; authorized runtime read;
forbidden runtime UPDATE and ledger read both SQLSTATE 42501. Temporary
clusters and files cleanup PASS. The one-time workflow is disabled with
[follow-up SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38440900971).

**Scope distinction:** This proves a synthetic procedure and role/owner
reconstruction, not recovery of real production cluster-global metadata.
The real database archive does not include `pg_dumpall` globals. Production
roles, memberships/authentication, Caddy/TLS runtime storage and service
rollback remain **unverified** and require their own approved recovery plan.
The earlier source-inventory and bootstrap sections reflect their historical
evidence dates; this addendum supersedes their pre-test synthetic gap status.
G17/1 remains IN PROGRESS, G17/2 NOT STARTED, no production write.

## Addendum — approved production Caddy/TLS metadata-only inventory (2026-10-13)

The operator authorized a strictly **read-only** Caddy/TLS dependency inspection
using the existing time-limited, host-key-verified IAP SSH route. [Successful
one-time MS-02 runner inspection](https://github.com/wwjd4u/MeshAlot/actions/runs/38442058347) reused valid access, and its
[post-disable execution was SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38442168184). The full sanitized JSON audit
and SHA-256 sidecar remain restricted locally on MS-02. No Caddyfile contents,
certificates, TLS private keys, vault materials, environment values, user
credentials or production changes entered CI logs or GitHub.

**Observed metadata:**
- Caddy service **active**; installed **Caddy CLI v2.11.4**. The CLI's
  `list-modules --skip-standard` reported **0 nonstandard modules**.
  This does not independently prove the active service executable has
  the same digest or full module set as the inspected CLI.
- The systemd Caddy service references the primary
  `/etc/caddy/Caddyfile`, readable under existing access.
- Bounded line-level parsing found **1 reverse_proxy** directive and **0**
  `import`, explicit `tls`, explicit `storage` or obvious manually
  referenced key-file directives. This is *not* a semantic Caddy adaptation
  test and cannot establish active automated certificate issuers or storage.
- `/var/lib/caddy` is a directory. The *candidate* data root
  `/var/lib/caddy/.local/share/caddy` was **UNVERIFIED** under existing
  read-only privileges; do not equate inaccessible with missing. The
  resolved active storage location, certificate renewal strategy and
  backup-versus-regeneration decision remain **UNVERIFIED**.
- No private TLS key/certificate bytes or production configurations were
  exported; no service restart, access renewal, IAM/SSH/DNS change,
  backup modification or provider-node operation.

**Recovery decision:** the metadata inspection is PASS, but **Caddy/TLS
recovery readiness is NOT PROVEN**. Next prepare stage-only actual Caddy
v2.11.4 HTTPS/WebSocket/negative-auth and safe recovery checks with
synthetic certificates/identities, preserving the earlier successful
Go proxy simulation without repeating it as new evidence. A separate
approved/private method will be necessary for verifying real production
certificate storage/recovery, if still required; never export TLS keys to
GitHub, GitHub Actions or chat. G17/1 remains open; G17/2 has not begun.

## 2026-10-13 pinned Caddy isolated-stage image reference — no image pull

The Docker Hub **official `library/caddy`** image detail page for
`2.11.4-alpine` identifies `linux/amd64` and Caddy `v2.11.4`,
matching the previously observed production **CLI version** (the
running service binary digest is not proven identical). Source:
[Official Docker Hub image details](https://hub.docker.com/layers/library/caddy/2.11.4-alpine/images/sha256-040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0).

- Published multi-platform index digest: `sha256:6aeddd44c3078b0f9a35206472a11420648a79c184603ef95957d0a20044cb2b`.
- Published **linux/amd64 image manifest digest**: `sha256:040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0`.
- Locked future test reference:
  `docker.io/library/caddy@sha256:040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0`, with
  `--platform=linux/amd64`.
- Require registry digest match and Caddy runtime version validation **before
  stage acceptance**. Digest pin identifies immutable image content; it does
  not prove that the image matches the live Caddy binary, supply chain
  signature, or production TLS certificate material.
- [MS-02 one-time read-only preflight](https://github.com/wwjd4u/MeshAlot/actions/runs/38442965161)
  PASS: Docker, OpenSSL, Go, Python3, sha256sum available; Caddy v2.11.4
  not installed and neither matching Caddy image tag cached.
  No pull, container creation or production access in the preflight.
  The workflow was disabled and
  [follow-up execution SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38443288594).

**Not authorized by the digest lookup:** pulling the new image onto the
MS-02 host, starting test containers, accessing real TLS keys or certificates,
accessing production, or changing any provider resources. Next execution
requires explicit operator approval for a digest-only image download and
a resource-bounded, network-none synthetic Caddy staging exercise. The
production certificate data-root and complete Caddy/TLS recovery remain
unverified; G17/1 still open, G17/2 not started.
