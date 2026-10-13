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
- [ ] Transfer the **existing new encrypted supplement and checksum**
  from MS-02 to the separate Toshiba encrypted APFS volume and verify the
  checksum. Do not store the existing recovery passphrase on Toshiba.
  This off-machine step is still PENDING.
- [ ] Inventory dependency requirements for global PG roles/grants,
  TLS/Caddy modules/storage, deployment ownership and service user,
  without exporting secret values.
- [ ] Confirm a safe disposable staging **target** with approved scope,
  budget, isolation and no production identity or network reachability.
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

The **next single operator action** is reconnecting the Toshiba encrypted
APFS volume to Mac and transferring exactly the new archive and checksum,
then independently SHA256-verifying them. Do not claim G17/1 complete
until this off-machine copy passes. The rest of the seven-checkpoint
sequence is unchanged.

## Next *single* action
Complete the supplemental encrypted environment archive off-machine
copy to the **already configured** Toshiba APFS volume, verify its
checksum, then return to the remaining G17/1 dependency checklist.

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
