# MeshAlot M13 — Gate 17 fixed completion sequence
Date: 2026-10-13 (US Central)
Status: **IN PROGRESS — NOT RELEASE-READY**

## Purpose and authority
This execution checklist condenses, but does not replace, the Google Drive
**AI Mesh POC Build, Test, and Release Guide**, Milestone 13, and
[Gate 17 controlled rollout plan](docs/m13-gate17-controlled-live-rollout-plan-2026-10-07.md).
No release approval, production write, node reboot, WAN interruption,
secret export or new infrastructure expense is inferred from this document.
Work down this checklist in order; close each item only against recorded
PASS evidence. If blocked, keep its place and resolve the blocker instead
of skipping ahead to unrelated work.

## Verified unchanged baselines (at 2026-10-13 preflight)
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469`,
  unchanged as of initial GitHub read.
- Working branch `m13-persistent-heartbeat` is **79 commits ahead, 0
  behind** `main` at this baseline; this is *not* a deployment.
- Original exact application candidate evidence: `4f67ee54e5ebca084e38e306e723ae0acd08e77d`.
  Any future cutover must pin and revalidate the intended immutable release.
- Gates 1–16 passed code/isolated tests. Production PostgreSQL is v16,
  migrations 1–8; migrations 9–10 have **not** been applied to production.
- [Step 3B/3C/3D/3E evidence](m13-gate17-backup-restore-rollback-readiness-2026-10-11.md):
  original GPG-encrypted database and nine-item partial API/systemd/Caddy
  rollback archives, isolated logical DB restore PASS, identical SHA-256
  on separate encrypted Toshiba APFS volume, two RoboForm passphrases
  operator-confirmed. Existing copies and recovery keys preserved.
- Step 3E.5: two systemd EnvironmentFiles *declarations* resolve to
  regular files; private owner/mode/size metadata captured safely with
  pre-existing noninteractive stat; **file contents were not backed up**,
  never exposed, and are not covered by the nine-item rollback archive.
- Step 3E.1: bounded primary Caddyfile inspection found no `import`
  directives; additional runtime/TLS dependencies are not fully certified.
- Read-only trusted gateway workflow on the MS-02 runner completed
  successfully October 12. Do not assume its online state forever.

## Closure gates — strict order

### G17/1 of 7 — Complete recoverability and isolated staging preflight
**Status: IN PROGRESS.**
- [x] Encrypted database archive and off-MS-02 checksum copy, prior isolated
  logical restore and original key custody evidence.
- [x] Encrypted partial API/systemd/Caddy rollback archive and off-MS-02
  checksum copy, original rollback key custody evidence.
- [x] Metadata-only classification of both systemd environment-file entries.
- [x] With express operator approval, capture the two systemd
  EnvironmentFiles declarations into **new encrypted local** GPG AES256
  supplemental archive using the existing API-rollback passphrase.
  [Successful job](https://github.com/wwjd4u/MeshAlot/actions/runs/38404900223).
  **Original source and other archives preserved.**
- [x] Transfer existing new encrypted supplement and its checksum
  from MS-02 to the existing Toshiba encrypted APFS volume via strict
  host-key SSH. Operator Mac Terminal confirmed
  `M13_G17_ENVSUPP_EXTERNAL_COPY=PASS` and
  `M13_G17_ENVSUPP_EXTERNAL_SHA256=PASS`. No recovery passphrase
  transferred and original MS-02 files preserved.
- [x] Identify code-level PG16 global role/grant, Caddy/TLS and
  systemd runtime recovery dependencies **without accessing secrets or
  production**. See [G17/1 static dependency inventory](m13-g17-g1-postgres-caddy-dependency-inventory-2026-10-13.md).
- [x] Execute **new targeted PostgreSQL 16 runtime ACL acceptance**
  in the approved isolated 2-CPU/4-GiB, network-none, synthetic-only
  staging environment, using exact pinned M13 source and tracked grants.
  Allowed operations and 8 expected-denied operations PASS.
  [Evidence](https://github.com/wwjd4u/MeshAlot/actions/runs/38426678379).
  **No real production global role/membership restoration has been tested.**
- [x] Exercise a **new synthetic-only PostgreSQL16 cluster-global role and
  database ownership reconstruction** on the approved isolated MS-02 stage.
  Separate nonadmin migration owner and restricted runtime role were exported
  through password-free globals, then restored with a synthetic logical
  database archive into a **fresh disposable PostgreSQL16 cluster**. Database,
  table, function and migration-ledger ownership, tracked grants, preserved
  fixture and two expected-denied SQLSTATE 42501 actions all PASS. [Run](https://github.com/wwjd4u/MeshAlot/actions/runs/38440768518).
  Stage cleanup PASS; the one-time test was disabled and [follow-up SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38440900971).
  This is **NOT** recovery of the actual production roles or PostgreSQL globals.
- [x] **Operator-approved, metadata-only production Caddy/TLS inspection**
  using the existing unexpired pinned IAP SSH route: Caddy active,
  CLI v2.11.4, 0 nonstandard modules, 1 primary-file reverse_proxy
  directive, 0 bounded-parser imports/explicit TLS/storage directives.
  Candidate certificate data root remains UNVERIFIED, not absent.
  [Read-only evidence](https://github.com/wwjd4u/MeshAlot/actions/runs/38442058347); workflow disabled with
  [follow-up SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38442168184). No secrets or production changes.
- [ ] Verify remaining **full recovery dependencies**, including real
  PostgreSQL global roles/ownership, Caddy/TLS runtime materials,
  and stage-only service rollback; synthetic ACL PASS is not equivalent
  to complete production recovery.
- [x] Operator approved existing MS-02 as a disposable **PG16 synthetic-only** staging target, bounded to 2 logical CPUs, 4 GiB RAM, Docker `--network=none`, no host bind mounts or ports. [Successful one-time bootstrap and isolation proof](https://github.com/wwjd4u/MeshAlot/actions/runs/38419118910); container and tmpfs cleaned. This is **not** a complete rollback rehearsal.
- [x] One-time MS-02 runner staging resource preflight PASS on 2026-10-13;
  **no staging environment has been created**. [Evidence](https://github.com/wwjd4u/MeshAlot/actions/runs/38399945655).

### G17/2 of 7 — Isolated rollback rehearsal
**Status: NOT STARTED.** Requires explicit scope approval for use of
production-derived encrypted archive bytes/passphrases, if used.
- Stage disposable PG16 with synthetic fixtures; do *not* rerun the
  already-passed production-data logical restore by default.
- Rehearse known-good API binary, systemd unit/drop-ins, primary Caddyfile
  using remapped stage-only paths, **never** production service paths.
- Satisfy required dependencies by approved isolated test substitutes;
  prove runtime ownership, permissions, loopback health, TCP/TLS/WSS
  proxy upgrades and negative authentication.
- Inject a disposable failure, revert to staged known-good state, verify
  recovery and preserve sanitized reports. Discard stage-only material
  per the operator-approved plan.

### G17/3 of 7 — Operator-controlled production cutover readiness
**Status: NOT STARTED.**
- Recheck source SHA, release artifact, DB migration ledger and roles,
  Caddy routes, per-node authorization, rollback criteria and budgets.
- Approve exact production window *separately*, migration 9/10,
  binary/Caddy/systemd updates, and controlled stage/canary scope.
  Until approval, do not deploy or alter prod.

### G17/4 of 7 — Restricted production service acceptance
**Status: NOT STARTED.**
- Under separately approved cutover, execute exact tracked migration
  procedure and restricted grants. Verify every migration checksum and
  historical records are preserved.
- Validate production HTTPS/WSS, login, dashboard, historical data,
  challenge authentication, telemetry and error/reconnect behavior,
  using pinned release. No shell exposure or inbound agent ports.

### G17/5 of 7 — Node001 supervised manual acceptance
**Status: NOT STARTED.**
- Existing Node001 identity and benchmark fingerprints unchanged.
- Manually start one M13 agent connection with provider mode **paused**;
  truthfully report GPU/VRAM/RAM/CPU/latency/availability.
- Validate M11 provider-control settings are authoritative after restart,
  not hardcoded. No unattended startup or marketplace scheduling until
  safeguards work.

### G17/6 of 7 — Real Node001 outage and reboot recovery
**Status: NOT STARTED.** Physical operation requires separately approved
time window with user present, never while driving.
- Disconnect WAN by approved physical means, wait >=90 seconds, confirm
  dashboard offline; reconnect and prove autonomous reconnect.
- Deliberately restart Node001, prove agent automatic recovery and
  resource-control enforcement. Preserve V100 evidence/Ollama/Hermes.

### G17/7 of 7 — Node002 and final monitoring/closeout
**Status: NOT STARTED.** Physical Mac sleep/wake and launch-at-login
need a supervised user window.
- Preserve original Node002 enrollment and the original M8 report.
- Confirm sleep/wake, disconnect, login/reboot and reconnect without
  re-enrollment; GPU fields may be absent when unsupported.
- Observe production health, authentication, telemetry correctness,
  budgets and operating logs; final security/rollback review.
- Archive sanitized evidence in GitHub and handoff documentation in
  Google Drive when the milestone is actually passed.
- Close Milestone 13 only if the governing functional pass criteria
  and explicit user signoff are met.

## What can be done while the operator is driving
Read-only GitHub/runner inspection, review of existing evidence, test
planning, sanitized documentation, isolated synthetic development/test
work **after staging approval**, and safe code fixes to the M13 branch.
No driving-time physical WAN interruption, reboot, production cutover,
secret handling that requires operator clipboard/vault, Node002
sleep/wake or background work claims without an actual scheduled task.

## Hard stop and preservation policy
No new accounts, new SSH key identities, Tailscale, host shell backdoors,
unapproved inbound ports, changes to control-plane production, node
re-enrollment, replacement of existing encrypted archives, or V100
benchmarks rerun. Keep Time Machine backup volume preserved and Toshiba
encrypted archive key material separately protected in RoboForm. Production
data, passwords, plaintext configs and encryption keys must not enter
GitHub Actions logs, GitHub repo, Google Drive or chat.

## 2026-10-13 environmental supplement update

Two pinned production systemd environment-file entries were captured
in a single, privately stored, separately AES256-encrypted archive on
MS-02. The existing API rollback GPG passphrase was reused without
a new key; its value remained private. The new ciphertext was verified
in-memory and the old backup files remain intact. One-time workflow
was disabled; the later job was [SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38405043699).

The operator's Mac Terminal subsequently confirmed both
`M13_G17_ENVSUPP_EXTERNAL_COPY=PASS` and
`M13_G17_ENVSUPP_EXTERNAL_SHA256=PASS`.
**Supplemental off-machine ciphertext copy and checksum verification
are now COMPLETE.** This does not establish full recovery readiness,
because global PostgreSQL roles/grants, Caddy/TLS runtime dependencies,
and isolated rollback staging are not yet proven.

## 2026-10-13 recovery dependency source inventory

The operator confirmed the Toshiba safely ejected. Read-only GitHub
source inventory identified the required PostgreSQL 16 runtime role
properties, separate migration owner, post-migration runtime grants,
Caddy/TLS runtime components, and the currently unsupported use
of generic Docker Compose `postgres:17` as a PG16 recovery fixture.
See [G17/1 dependency inventory](m13-g17-g1-postgres-caddy-dependency-inventory-2026-10-13.md).
**Source-level planning PASS**, but no live PG role/ACL recovery,
Caddy/TLS runtime validation or isolated staging execution yet.

## Next *single* action
Review the remaining Caddy/TLS and PostgreSQL **global**
ownership/recovery dependencies before G17/2. Synthetic runtime ACLs
are verified. Do not open or decrypt production-derived archives,
use vault credentials, access real node identities or deploy to live
services without specific authorization. The next executable stage
test should close an untested recovery dependency rather than repeat
the completed migration or ACL tests.

## 2026-10-13 staged-resource readiness snapshot (read-only)
The trusted MS-02 runner independently returned 14 logical CPUs,
117 GiB `MemAvailable` (integer floor), 625 GiB filesystem available
(integer floor), and an accessible Docker daemon.
`M13_G17_STAGE_CAPACITY_PRELIMINARY=SUFFICIENT_FOR_PLANNING`.
The workflow did **not** create any container, test container network
isolation, read recovery credentials, connect to production,
run any V100 benchmark, modify a provider node or allocate new cloud
resources. [Preflight run](https://github.com/wwjd4u/MeshAlot/actions/runs/38399945655).
The one-time job was disabled and its
[next execution skipped](https://github.com/wwjd4u/MeshAlot/actions/runs/38400025350).

**Planning conclusion:** The *existing* MS-02 runner has sufficient local
resources to propose disposable, synthetic-first, isolated staging without
new cloud costs. This does **not** authorize a Docker container, run a
production-derived archive, or confirm network isolation. Choose staging
target and scope with the operator before running stage tests.
## 2026-10-13 approved temporary stage execution

Operator authorized a disposable local stage on MS-02 (no new cloud
cost or production access). First bootstrap [run 38418971253](https://github.com/wwjd4u/MeshAlot/actions/runs/38418971253)
started its container but stopped at an overly strict Docker
`NetworkSettings.Networks` structure assertion; cleanup PASS, no
data or backups touched. The corrected one-time
[run 38419118910](https://github.com/wwjd4u/MeshAlot/actions/runs/38419118910)
completed successfully:

- Local cached `postgres:16-alpine`, **no image pull**.
- `--network=none`, no host-port publication/bind mounts, 2-CPU and
  4-GiB memory limits, read-only root, no capabilities, no new privileges,
  user `70:70`, temporary in-memory PostgreSQL data/socket/tmpfs.
- PostgreSQL 16 socket healthy; synthetic-only fixture inserted/read.
  Synthetic owner/runtime roles had restricted non-admin attributes.
  This is a role-policy bootstrap, **not** full tracked application
  role/ACL recovery or a real production restore.
- Private sanitized evidence and SHA-256 PASS; no SSH/IAP, production
  secrets, original archive bytes, historical V100 benchmark or
  enrolled node identity accessed.
- All disposable containers and files deleted; explicit cleanup PASS.
  The one-time workflow was disabled; its
  [next run skipped](https://github.com/wwjd4u/MeshAlot/actions/runs/38419245216).

**G17/1 remains IN PROGRESS** until staged role/grant dependencies,
Caddy/TLS dependencies and recovery strategy meet acceptance.
**G17/2 remains NOT STARTED**; no known-good API/config archive has been
used in staging. M13 remains OPEN.

## 2026-10-13 targeted PostgreSQL16 runtime ACL acceptance

[One-time PG16 synthetic role/grant test](https://github.com/wwjd4u/MeshAlot/actions/runs/38426678379)
completed successfully on the approved MS-02 Docker runner. Using pinned
application source `4f67ee54e5ebca084e38e306e723ae0acd08e77d`,
the tracked migrations were applied **only as a fixture setup** for the
new actual ACL tests; this is not a repeat claim of the already-passed
8→9→10 migration upgrade test.

- Exactly PostgreSQL16, `--network=none`, 2 CPUs/4 GiB memory,
  no bind mounts or host ports; in-memory disposable database.
- Stage-only `meshalot` role was LOGIN, non-SUPERUSER, no CREATEDB,
  CREATEROLE, REPLICATION, BYPASSRLS or inherited administrative roles.
- Actual checked-in `database/runtime-grants.sql` executed in stage
  after tracked migrations 1–10. Authorized column reads, node telemetry
  UPDATE and node-rating function execution PASSED.
- **Eight intentional forbidden actions all returned PostgreSQL
  SQLSTATE 42501**, including restricted user fields, unauthorized user
  updates, direct ratings insertion, wallet deletion, schema/table DDL,
  migration-ledger reads and role creation. No unwanted table/role was
  created.
- Worktree unchanged; no real node IDs, recovery keys, private backups,
  production connections or V100 tests accessed.
- Container and all disposable temporary files removed: explicit PASS.
  One-time workflow subsequently disabled; [follow-up run SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38426776392).

**G17/1 still IN PROGRESS** because production global role/owner
reconstruction, actual Caddy/TLS runtime dependency recovery and an
isolated API/systemd/Caddy rollback have not been proven. This ACL test
does not authorize production migration or cutover. G17/2 has not
started and M13 Gate 17 remains OPEN.


## 2026-10-13 synthetic PostgreSQL16 global ownership reconstruction

The operator-approved isolated recovery test [run 38440768518](https://github.com/wwjd4u/MeshAlot/actions/runs/38440768518)
completed successfully on the trusted MS-02 runner. The source was pinned to
`4f67ee54e5ebca084e38e306e723ae0acd08e77d`; no live services were used.
A synthetic PostgreSQL16 source cluster created separate, nonadmin
`meshalot_stage_owner` and restricted `meshalot` roles. The migrations were
used solely to establish the fixture schema and owner (not to repeat the
previously accepted migration acceptance). The tracked runtime grants were
applied. The synthetic globals were exported without role passwords, and a
custom-format logical database archive was created. The entire source
cluster was then disposed of before a **new clean PostgreSQL16 cluster**
restored those synthetic roles, owners, grants and schema/data.

Acceptance PASS: PG16 with Docker network none, no ports/bind mounts,
read-only container, 2 logical CPUs and 4 GiB memory per sequential stage;
synthetic global roles reconstructed; migration/runtime role separation and
nonadministrative flags retained; database, tracked tables, migration ledger
and security-definer function ownership restored; migrations 1–10 and the
synthetic record intact; restricted runtime SELECT remained allowed; restricted
UPDATE and migration-ledger read rejected with SQLSTATE 42501. Temporary
containers and synthetic dump files removed, explicit cleanup PASS.
The workflow was disabled and [next execution SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38440900971).

**Remaining unverified:** actual production cluster global roles, memberships,
password/authentication material, schema owner inventory, Caddy module/TLS
certificate and storage recovery, systemd service runtime dependencies, and
full API/systemd/Caddy stage-only rollback. No production archive, recovery
key, production DB, SSH session, real node ID, V100 benchmark or live
service was accessed. **G17/1 stays OPEN; G17/2 has NOT started.**

## 2026-10-13 approved Caddy TLS metadata inventory

[One-time approved production read-only run](https://github.com/wwjd4u/MeshAlot/actions/runs/38442058347) **PASS**.
Existing authorized IAP SSH was still valid; no access renewal, new key,
production configuration change or service restart. Caddy service active,
installed CLI version **v2.11.4**, **0 nonstandard modules** reported.
The service references its readable primary Caddyfile, whose bounded
line-level scan found **1 reverse_proxy** and **0** imports, explicit
TLS/storage directives and obvious manual key references. Actual active
binary digest, semantically adapted configuration, certificate automation
and real TLS storage location were **not** verified. `/var/lib/caddy`
exists; deeper candidate data root **UNVERIFIED** rather than missing.
Full private audit/sidecar kept restricted on MS-02; no certificate or
private-key content accessed. One-time workflow disabled and
[subsequent run SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38442168184).

**Remaining G17/1 gap:** real Caddy version-matched isolated HTTPS/WSS and
failure/recovery test, plus production certificate-storage/renewal strategy
and API/systemd/Caddy rollback dependencies. G17/2 NOT STARTED.
No production cutover authorization has been given.
