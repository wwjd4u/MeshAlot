# MeshAlot M13 — Gate 17: Backup, Restore, and Rollback Readiness

Date: 2026-10-11 (US Central)

**Status: PARTIAL PASS / GATE 17 OPEN.** This is a sanitized continuation
record, not permission to deploy, change production, start a service, apply
migrations, or disrupt provider nodes.

**Operator continuation (2026-10-12):** Step 3D external encrypted archive copy and four-file integrity verification are PASS from operator-provided Mac Terminal output; two recovery passphrases were saved and checked in separate RoboForm Safenotes per operator confirmation, with clipboard clearing PASS. Gate 17 remains OPEN; see appended checkpoint. Historic sections below retain their 2026-10-11 as-of status.

## Verified baselines

- Governing document: **AI Mesh POC Build, Test, and Release Guide**,
  Milestone 13 — persistent agent connection and heartbeat.
- Controlled rollout plan:
  [M13 Gate 17 controlled live rollout](m13-gate17-controlled-live-rollout-plan-2026-10-07.md).
- Production `main` observed unchanged:
  `1016c9b058690f34058dfab9cde7927e421fa469`.
- Tested M13 application candidate (immutable source):
  `4f67ee54e5ebca084e38e306e723ae0acd08e77d`.
- M13 code/isolated Gates 1–16: previously passed, **not** evidence of a
  production heartbeat/WSS deployment or real interruption/reboot acceptance.
- Production PostgreSQL: major 16; migration history **1–8**. Migrations
  **9 and 10 remain unapplied**. The application and Caddy remain deployed
  at the existing production release; no Gate 17 cutover was performed.

## Gate 17 Step 3: backup/restore evidence

| Check | Recorded outcome | Evidence and limits |
| --- | --- | --- |
| Step 3A remote production backup-readiness preflight | PASS | [Runner evidence](https://github.com/wwjd4u/MeshAlot/actions/runs/37824023544) |
| Step 3B operator-created encrypted archive | PASS (operator-reported) | AES-256 GPG archive, SHA-256 match, full decryption check, PostgreSQL 16 custom-format TOC readable; 156 TOC entries including 15 table-data entries |
| Step 3C isolated logical restore | PASS (operator-reported) | User-provided terminal output: isolated network-disabled PostgreSQL 16 restore completed; migrations 1–8 and restored table reads passed |
| Remote review of private archive + recovery-evidence metadata | PASS | [MS-02 local-only read-only validation](https://github.com/wwjd4u/MeshAlot/actions/runs/38159562092), verifying on-host archive hash, key file permissions, evidence claims, and PostgreSQL image. It did **not** rerun the restore |
| Step 3D independent backup/recovery copy | **PENDING** | 2 TB USB/APFS drive identified by [runner inspection](https://github.com/wwjd4u/MeshAlot/actions/runs/38161210924). User said it was mounted on MacBook, but Mac mount/free space and a copy onto it have **not** been independently verified. Existing files on drive untouched |
| Separate protected recovery-key copy | **PENDING** | Original recovery passphrase remains only in private MS-02 storage as far as verified; no key content was exported or posted |

**Security boundary:** Full production data, GPG passphrase, private SSH keys,
production environment files, and raw production audit reports are not
included in this document, GitHub Actions output, or Google Drive. Privately
held MS-02 evidence is not a substitute for off-machine recovery protection.

## Pinned offline regression after recovery checks

Trusted MS-02 runner:
[2026-10-11 Gate 17 pinned offline regression](https://github.com/wwjd4u/MeshAlot/actions/runs/38165960011).

- Pinned code SHA and production `main` baseline: PASS.
- `go mod verify`: PASS.
- `go test ./... -count=1`: PASS.
- `go test -race ./agent -run '^TestM13(Reconnect|Persistent)' -count=1`: PASS.
- `go vet ./...`: PASS.
- Temporary agent build: PASS.
- Source worktree unchanged: PASS.
- Production DB, Node001/Node002 enrollment, V100 tests, service states,
  and external APFS backups were not touched by the validation workflow.

## Additional private rollback-report integrity validation

On 2026-10-11, the trusted MS-02 runner executed a separate **local-only,
read-only** checksum verification of the three Step 2 private audit report
families. [Verified workflow](https://github.com/wwjd4u/MeshAlot/actions/runs/38166281459).

- Production control/configuration audit: 1 private report; SHA-256 PASS.
- Service/Caddy follow-up audit: 1 private report; SHA-256 PASS.
- Running API binary audit: 1 private report; SHA-256 PASS.
- No production SSH connection, file export, decryption, rollback, or
  production mutation occurred during this check.
- The reports and hashes remained under restrictive user-local MS-02
  permissions; no report contents or sensitive configuration were uploaded.
- **Known-good API binary/configuration rollback copies still NOT VERIFIED.**
  Valid audit reports alone are not deployable rollback artifacts.
- The Mac-attached APFS off-machine backup copy and separately protected
  recovery passphrase both remain **PENDING**.

## Additional synthetic production-matched migration validation

On 2026-10-11, a trusted MS-02 workflow ran a disposable, network-isolated
PostgreSQL 16 test from the exact candidate source SHA
`4f67ee54e5ebca084e38e306e723ae0acd08e77d`.
[Verified Gate 17 V8→V10 synthetic migration run](https://github.com/wwjd4u/MeshAlot/actions/runs/38166687485).

- Applied original tracked migrations **1–8** in a newly created test DB,
  matching production's known schema version without reading production data.
- Inserted only synthetic fixture identities, status/heartbeat and network
  benchmark data; no real user/node IDs, report content, private keys, or
  passwords were involved.
- Applied untouched tracked migrations **9 and 10** with the original
  `database/migrate.py`, then verified migration status and a repeated
  idempotent apply.
- Confirmed all **10** tracked versions were recorded exactly once, the
  version 9 `node_ratings` table exists, and historical synthetic node
  status, heartbeat, and Network Score 67 remained intact with NULL
  M13 telemetry for the legacy fixture.
- Pinned source and `main` baseline matched expectations; worktree remained
  unchanged. Network mode for the disposable DB was `none`; no production
  database, recovery archive/passphrase, deployed services, node identities,
  or V100 artifacts were accessed or changed.

**Meaning:** The production-matched schema-upgrade path is now supported by
a synthetic database test in addition to the earlier Gate 16 9→10 test.
It is **not** a rehearsal using the restored production backup, nor approval
for migration, release, or application cutover.

## Gate 17 local rollback-readiness classification — 2026-10-11

The trusted MS-02 runner re-evaluated the **existing, checksum-verified**
private Step 2 production audit reports **without initiating an SSH/IAP
connection**. [Successful sanitized classification](https://github.com/wwjd4u/MeshAlot/actions/runs/38167302922).

- Original production-control audit and migration history 1–8: PASS.
- Original production service and Caddy loopback health evidence: PASS.
- Original running API binary SHA-256 evidence: PASS.
- Four pre-existing production backup directories were inventoried in the
  older audit, but **their recoverability is NOT VERIFIED**.
- Actual rollback copies of the active API binary, its systemd unit/drop-ins,
  Caddy configuration, and required private environment material:
  **NOT VERIFIED**.
- Independent backup copy on the Mac-attached APFS disk: **NOT VERIFIED**.
- The original **24-hour, instance-specific IAP SSH authorization is EXPIRED**
  as of this read-only local check. Its existing metadata must not be
  changed or extended without fresh user approval; other keys and IAM
  settings must remain preserved.
- No live production connection, database export, service restart, migration,
  or host changes were made.

**Safe stop:** A fresh, limited-scope operator authorization is required
before reinstating any control-server access and completing the production
rollback-artifact capture. Preparation and checks based on existing local
evidence can continue without new cloud access. Do not infer that the
successful database-only logical restore covers binary/configuration rollback.

## Temporary VM-specific SSH renewal — 2026-10-11

The operator explicitly approved a **24-hour renewal of the same existing
Ed25519 public key**, limited to the instance `meshalot-control-01`, for
read-only rollback auditing. No new account or SSH identity was generated,
no project-wide SSH key or IAM policy changed, and no production service or
database state was modified.

- [Guarded instance metadata renewal attempt](https://github.com/wwjd4u/MeshAlot/actions/runs/38167622592)
  verified the single existing key-entry replacement, preservation of other
  VM and project metadata, and the private local registration update. The
  GitHub job concluded **failure** due to a closing-quotation typo in its
  final status line **after** those verification markers had passed; it
  must not be represented as a successful job.
- [Independent read-only verification](https://github.com/wwjd4u/MeshAlot/actions/runs/38167753474):
  **SUCCESS**, including matching local/remote key and 24-hour expiration,
  preserved previous private registration history, unchanged project login
  policy, strict pinned-host-key IAP SSH authentication, and no remote
  mutations or file exports.
- The one-time mutating renewal workflow was then **disabled at the job guard**
  on the infrastructure branch to prevent an unattended future extension
  without a new operator approval. The read-only verification workflow remains
  available.
- The temporary authorization automatically expires about 24 hours after
  the registered October 11 renewal; do not assume it remains valid across
  future handoffs. Future extensions require renewed authorization and an
  independent scope check.

**Next checkpoint:** Read-only inventory of the existing production API,
systemd/Caddy configuration and rollback directories, then separately request
permission before copying sensitive production rollback material. The off-Mac
database archive transfer and independent recovery-key protection are pending.

## Gate 17 production rollback inventory — 2026-10-11

A newly approved, **read-only** inventory ran over the existing time-limited
Google IAP SSH channel to the exact `meshalot-control-01` host.
[Successful GitHub Actions read-only inventory](https://github.com/wwjd4u/MeshAlot/actions/runs/38168379468).

- SSH authorization, pinned host identity, and wrong-host guard: **PASS**.
- Running production API binary fingerprint matched the private Step 2
  baseline (October 8): **PASS / UNCHANGED**.
- One MeshAlot application service remained active.
- One systemd unit fragment and six systemd drop-ins were fingerprinted
  privately. One active unit declared an environment file; the environment
  file's values or contents were **not** read or uploaded.
- The Caddyfile's SHA-256 was captured in the private report and was **not**
  emitted to GitHub logs. This creates an audit fingerprint, not a rollback
  copy or a comparison against a previously archived Caddyfile.
- Four existing rollback/backup directory names matched the older Step 2
  inventory. A depth-limited, metadata-only file count was saved privately;
  the files' contents, completeness and restorability remain **UNVERIFIED**.
- Detailed paths and hashes remain exclusively in the checksum-verified
  `0700` MS-02 private-audit directory, with `0600` report/checksum files.
- **No files were exported**, and no PostgreSQL changes, service restarts,
  deploys, Caddy edits, SSH key changes or Node001/Node002 identity changes
  occurred during this inventory.

**Still pending:** Independently preserved known-good API binary,
systemd/Caddy and private environment rollback artifacts; their hashes,
recovery procedures and access controls; a separate encrypted database
backup copy on the Mac-connected APFS volume; and independently protected
recovery-key storage. The inventory did **not** authorize any of those
sensitive file transfers.

## Encrypted known-good production rollback capture — 2026-10-11

The operator separately approved copying the running production API binary,
systemd unit/drop-in files, and Caddyfile **only** into restricted MS-02
storage. No environment-file contents or database records were authorized for
this export. The operation used the existing instance-only, time-limited IAP
SSH identity; no permanent remote access was added.

- [Encrypted rollback capture — PASS](https://github.com/wwjd4u/MeshAlot/actions/runs/38168945744).
  The previous attempt failed safely during local Python setup **before**
  contacting production or creating an archive.
- Captured **9** items: the hash-pinned active API executable, one systemd
  unit fragment, six systemd drop-ins, and the Caddyfile. Original source
  paths, hashes, ownership, and modes were recorded **inside** the encrypted
  archive's private recovery manifest, not in public logs.
- The production source was read only. Artifacts were streamed over pinned
  IAP SSH directly into local GPG AES-256 encryption; no plaintext production
  rollback archive was written on the MS-02 or uploaded to GitHub or Drive.
- A new unique 0600 recovery passphrase was generated and retained only in
  user-private MS-02 storage, outside the 0700 rollback archive folder.
  Its contents were never printed or published.
- The encrypted tar stream decrypted and all 9 members' SHA-256 hashes
  matched the previous private production inventory. The running executable
  also matched the earlier Step 2 fingerprint. The encrypted output has a
  private SHA-256 checksum and local 0600 evidence record.
- [Independent, local read-only verification — PASS](https://github.com/wwjd4u/MeshAlot/actions/runs/38169049080)
  confirmed archive SHA-256, private key file permissions, preserved
  evidence, and the absence of any off-machine copy. It did **not** read the
  passphrase, decrypt again or connect to production.
- [One-time capture workflow disabled](https://github.com/wwjd4u/MeshAlot/actions/runs/38168987214)
  after successful export; the next job was **skipped**. Re-exporting these
  files requires fresh authorization and an explicit workflow change.

**Limits:** The saved files are a verified **partial rollback package**, not
full disaster-recovery certification. Systemd environment-file contents,
potential additional imported Caddy configuration and other dependent
secrets were deliberately excluded. No off-machine copy of the encrypted
archive or separate protected passphrase exists as verified. No live rollback
or restoration of production binaries/configurations has been attempted.
Existing backup directories remain inventoried but not restore-tested.

## Mac-connected external backup destination and trusted SSH — 2026-10-11

Operator-provided Mac Terminal results verified the following; this section
records **user-supplied terminal evidence** rather than a remote Mac runner
audit.

- Physical external Toshiba: 2 TB APFS container. Existing encrypted
  Time Machine volume `TOSHIBA EXT` was preserved and remains a separate
  APFS Backup-role volume with an approximately 1.36 TB quota.
- The new **external** APFS volume is `MeshAlot-Backup` (singular),
  mounted at `/Volumes/MeshAlot-Backup`, reported by macOS as encrypted
  and unlocked with a **400 GB quota**. It is distinct from the protected
  Time Machine volume. Previous accidental `MeshAlot-Backups` internal
  volume was marked for removal; do not assume its deletion was
  independently verified.
- Current MS-02 LAN SSH endpoint from trusted local-only runner inventory:
  `192.168.5.98:22`, replacing the obsolete `192.168.5.131`.
  [Runner verification](https://github.com/wwjd4u/MeshAlot/actions/runs/38188173920).
- Trusted ED25519 **server host key** fingerprint read from MS-02:
  `SHA256:q8bmG7nOFS17NCNIOcAR4Z5Cc+L7O1KJd0Q66K9yEq0`.
  [Runner verification](https://github.com/wwjd4u/MeshAlot/actions/runs/38188347513).
  The Mac used `ssh-keyscan`, compared it to that exact fingerprint, and
  appended the matching public host key to `~/.ssh/known_hosts`.
- The Mac's existing `google_compute_engine` user private key was not
  accepted by the MS-02 for `jason_guynes`. **Do not assume key-based login
  works, and do not create a new SSH identity as a shortcut.**
- Mac-to-MS-02 SSH **password authentication PASSED twice** with strict
  host-key checking and explicitly disabled public-key authentication. The
  remote `hostname -s` returned `wwjd4u-MS-02-Ultra`.
- No MeshAlot encrypted archive or checksum has yet been copied onto the
  external volume as of this handoff. Both original archives and their
  original recovery keys remain in private MS-02 storage.

**NEXT CONTROLLED OPERATOR STEP:** Create a distinct private directory on
`/Volumes/MeshAlot-Backup`, then transfer **encrypted archives plus checksum
files only** from MS-02 using the now-verified password-based SSH route.
Use exclusive/no-overwrite writes and check hashes on the external destination.
Original recovery passphrases MUST NOT travel with the archives or be printed
to GitHub, Drive or chat. Do not start a new DB dump, regenerate either key,
modify production, or claim off-machine recovery until both copies are
independently SHA-verified. Use one numbered step at a time.

## Rollback readiness — explicit limitations

The [Step 2 read-only production audit](m13-gate17-step2-production-readonly-closeout-2026-10-08.md)
verified the currently active API binary hash and inspected systemd/Caddy
configuration, storing detailed reports privately on MS-02. Four older backup
directories were inventoried; **their presence does not establish that any
contains a complete, recoverable release backup**.

**Still required before a live cutover:**

1. Verify independent recovery copy of encrypted database archive and its
   SHA-256 hash on a separate device, with a separately protected recovery
   passphrase and tested recovery instructions.
2. The API binary, systemd unit/drop-ins, and primary Caddyfile now have
   a verified private encrypted copy. Review remaining **environment-file
   dependencies and any imported Caddy configuration** without publishing
   their contents; these files were deliberately excluded from the approved
   capture. Rehearse controlled recovery before claiming a complete package.
3. Prepare a controlled restore/rollback playbook and staging rehearsal,
   considering dependencies, ownership, privileges, and cluster-global roles.
   A `pg_restore --no-owner --no-privileges` logical test alone does not
   certify production permissions, global role recreation, or full disaster
   recovery.
4. Obtain a **separate explicit approval** before any new environment,
   production migrations 9 and 10, live binary/Caddy updates, agent service
   enablement, physical WAN/reboot, or Node002 Mac sleep/wake tests.

**Rollback principles:** Preserve the known-good production API, existing
database history, original node identities, M8/M9 scores, M11 provider
controls, and historical V100 hashes. If a cutover fails, revert the binary,
service and proxy to the known-good versions; do **not** blindly restore the
database, drop additive telemetry, skip a tracked migration, or overwrite
applied migration checksums. Never use Tailscale or add remote shell
backdoors.

**Current decision:** Do not close Gate 17, do not merge the M13 work branch
into production `main`, and do not advance to a live rollout until recovery
protection and all authorization gates are independently satisfied.

## Operator continuation — 2026-10-12: external archives and RoboForm

**Scope:** This section records the user's Mac Terminal output and verbal
confirmations. It does not represent a GitHub runner audit of the Mac or
independent retrieval/decryption from RoboForm. Do not copy archive bytes or
passphrases to this repository, to GitHub Actions logs, to Google Drive,
or into chat.

- **Step 3D.1 private destination: PASS.** Operator-created folder
  `/Volumes/MeshAlot-Backup/M13-G17-20261011-Encrypted-Archives` on the
  400 GB quota encrypted Toshiba APFS volume, after checking external
  device and recorded volume UUID. Directory mode `0700`.
- **Step 3D.2 existing encrypted PostgreSQL16 archive: COPY/SHA-256 PASS.**
  Copied archive plus associated checksum over strict-host-key/password
  LAN SSH from MS-02 without replacing the original archive.
- **Step 3D.3 existing encrypted API/systemd/Caddy rollback archive:
  COPY/SHA-256 PASS.** Copied only encrypted archive and corresponding
  checksum, preserving the original MS-02 file.
- **Step 3D.4 external inventory: PASS.** Operator Terminal output showed
  exactly four files; both original recorded SHA-256 digests matched the
  external copies; all four files mode `0600`, folder mode `0700`.
  The Toshiba volume is an off-MS-02 copy but shares the same physical disk
  as the existing distinct, protected Time Machine APFS volume; it is not an
  independent physical failure domain from Time Machine.
- **Step 3D.5 recovery-key vault copies: OPERATOR CONFIRMED.** Existing
  PostgreSQL16 and API rollback decryption passphrases were stored in two
  separate private RoboForm Safenotes. The operator stated both were saved
  and verified. SSH source-key checks required mode `0600`. The original
  MS-02 passphrase files were preserved. The final current clipboard clear
  was evidenced by `M13_STEP3D_RECOVERY_CLIPBOARD_CLEARED=PASS`.
  No passphrase values are included in this record. Any clipboard-history or
  Universal Clipboard retention was not independently verified.

**Not yet proven:** a separate fresh recovery/decryption test using the
RoboForm copies (not attempted); password-vault recovery if the Mac/MS-02 is
lost; complete API/Caddy/systemd rollback dependency coverage; global
PostgreSQL roles and file ownership; and a production-shaped isolated
rollback rehearsal. The previously completed database logical restore and
rollback-archive member checks must not be repeated without a concrete need.

**Next governing Gate 17 scope:** inventory absent environment-file
and imported Caddy dependencies by approved metadata-only inspection, then
plan an isolated staging rollback rehearsal. No production migrations,
services, files, identities, SSH policies, provider controls, or historic
V100 evidence were modified by these copy/vault steps. Production cutover
remains blocked pending separate explicit operator approval.


## Gate 17 Step 3E.1 — read-only rollback dependency inventory (2026-10-12)

**Authorization and limits:** Operator explicitly approved a read-only
inventory of the production systemd environment-file dependencies and
primary Caddyfile import declarations. The inspected production host was
the established `meshalot-control-01` over the previously authorized,
host-key-pinned, time-limited Google IAP SSH path. No extension of SSH
authorization, additional IAM identity, new key, migration, production
deployment, service restart, or new backup was performed.

- [Trusted runner inspection](https://github.com/wwjd4u/MeshAlot/actions/runs/38295531951):
  **SUCCESS**. One production MeshAlot service unit identified.
- Two **systemd environment-file declarations** were discovered and
  parsed. Both underlying file-metadata lookups remained **UNVERIFIED**
  under the available read-only inspection; **neither was confirmed
  missing**. No environment-file contents or secret values were accessed.
- The **primary Caddyfile was read for import directives** on the
  production host, without publishing its contents. The bounded parser
  identified **zero import directives**, zero unresolved references,
  and zero parsing faults within the inspected primary file. This does
  not certify unrelated Caddy runtime dependencies.
- Private JSON evidence and SHA-256 checksum were preserved in the
  pre-existing, owner-restricted MS-02 audit directory. GitHub logs
  contain only sanitized statuses/counts, never environment values,
  recovery passphrases, or raw Caddy configuration.
- The initial workflow-definition attempt
  [failed before starting a job](https://github.com/wwjd4u/MeshAlot/actions/runs/38295457188);
  corrected code ran successfully at the link above. The one-time
  workflow was then disabled and its [next execution was SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38295774465),
  preserving the least-privilege boundary.
- An optional attempt to arrange an additional privileged file-metadata
  follow-up was **not executed**. No privileged inspection outcome is
  claimed; environment-file presence, ownership, modes and backup
  completeness remain **pending**.

**Decision:** Step 3E.1 is **PARTIAL PASS / DEPENDENCIES UNVERIFIED**.
The existing encrypted binary/systemd/Caddy recovery archive remains a
verified *partial* package only; there is still no complete,
tested rollback of environment files, service ownership, global
PostgreSQL roles, or staged system restart. Gate 17 remains **OPEN**.
Proceed only with separately reviewed, non-disruptive recovery planning
until the unresolved dependencies are classified. No main-branch or
production configuration changes were made.

## Gate 17 Step 3E.2 — isolated rollback staging plan (2026-10-12)

**PLAN COMPLETE / EXECUTION NOT STARTED.** The [Step 3E.2
dependency checklist and isolated staging rehearsal plan](docs/m13-gate17-step3e2-isolated-rollback-staging-plan-2026-10-12.md)
defines explicit scope approvals, remaining environment-file and runtime
dependency classifications, an isolated synthetic-first staging design,
future authorized archive-use procedures, negative/failure tests,
evidence requirements, and safe-stop conditions. No environment was
created, no recovery passphrase accessed, no existing backup decrypted,
no production metadata re-inspected, and no live system or CI test
was executed as part of this planning step.

**Existing blockers:** metadata for both systemd environment-file
declarations remains unverified; complete Caddy/runtime dependencies
and PostgreSQL global roles/grants/ownership are not demonstrated;
no full rollback rehearsal is complete. The earlier Gate 17 Step 3E.1
inspection remains PARTIAL PASS. No scope increase is inferred from
the operator's approval to continue documentation work.

**Result:** 3E.2 DOCUMENTED; 3E staging execution PENDING;
Gate 17 OPEN. The next single controlled step should
classify missing environment-file metadata under separately reviewed
read-only authorization or an acceptable stage-only substitute.


## Step 3E.3 — existing production SSH access preflight (2026-10-12)

**Status: BLOCKED / SAFE STOP (no live dependency inspection).**

Operator approved a further narrowly scoped *read-only* environment-file
metadata check. Before connecting to the production VM, a one-time trusted
MS-02 runner job inspected only the existing local time-limited IAP SSH
registration metadata and permitted host identity:

- [Preflight run 38299425609](https://github.com/wwjd4u/MeshAlot/actions/runs/38299425609):
  **SUCCESS** for the local preflight, but registered authorization state
  **EXPIRED_OR_INSUFFICIENT** against the required 15-minute access margin.
- Exact results: `M13_STEP3E3_EXISTING_IAP_AUTH=EXPIRED_OR_INSUFFICIENT`;
  `M13_STEP3E3_REMOTE_ENV_METADATA=NOT_ATTEMPTED`;
  `M13_STEP3E3_PRODUCTION_CONNECTION=NONE`;
  `M13_STEP3E3_NO_ACCESS_RENEWAL=PASS`;
  `M13_STEP3E3_PRODUCTION_CHANGES=NONE`.
- No read of private SSH-key contents, no recovery-key access, no new IAM
  user, no SSH credential renewal, no production connection, no privileged
  `stat`, no service restart and no configuration change.
- The one-shot workflow was disabled after the run, and its
  [next run was SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38299488841).
  The one-time infrastructure workflow change did not modify production
  `main`.
- The **two declared systemd environment-file dependencies remain
  unverified**. The earlier 3E.1 metadata check did not prove whether
  either one exists or is missing; no change to that classification is
  warranted.

**Next decision boundary:** if live production metadata inspection is
needed, obtain fresh explicit operator authorization before renewing the
*existing*, instance-scoped SSH registration. Such renewal alone does not
confer elevated file metadata permissions; if a narrow privileged read
is required, request separate explicit approval, without changing
system permissions. No secret file contents may be exported or logged.
Staging and production rollout remain unauthorized; Gate 17 stays OPEN.

## Step 3E.3 continuation — approved 24-hour VM-only SSH renewal (2026-10-12)

**OPERATOR-APPROVED SCOPE:** The operator expressly authorized renewal
of the *existing* Ed25519 SSH registration for 24 hours, limited to
`meshalot-control-01`, to support the pending read-only environment-file
dependency inspection. The renewal did **not** authorize elevated access,
credential export, IAM policy changes, production service/database changes,
or a Gate 17 deployment.

- [One-time renewal execution](https://github.com/wwjd4u/MeshAlot/actions/runs/38301400277):
  **SUCCESS** on the trusted `meshalot-ms02` runner. The existing
  private/public keypair and username matched, the previously registered
  expiry was expired, and exactly **one matching VM-instance SSH public
  key metadata entry** was replaced with the same key and a 24-hour
  expiry. A guarded before/after inspection confirmed other instance
  metadata, project-wide metadata, login policy and VM tags unchanged.
  Original local registration JSON was preserved in its owner-restricted
  versioned history; the updated local registration used mode `0600`.
  Successful log markers include:
  `M13_G17_24H_VM_ONLY_SSH_RENEWAL=PASS`,
  `M13_G17_NEW_VM_KEY_ENTRY_VERIFIED=PASS`,
  `M13_G17_OTHER_VM_METADATA_UNCHANGED=PASS`,
  `M13_G17_PROJECT_METADATA_UNCHANGED=PASS`, and
  `M13_STEP3E3_EXISTING_VM_ONLY_SSH_RENEWAL=PASS`.
- [Independent read-only verification](https://github.com/wwjd4u/MeshAlot/actions/runs/38301521800):
  **SUCCESS**, including the exact existing Ed25519 public key,
  24-hour expiry, prior local registration history, unchanged project
  SSH policy, and a real pinned-host-key Google IAP SSH hostname-only
  connection to `meshalot-control-01`.
  `M13_STEP3E3_RENEWED_EXISTING_SSH_INDEPENDENT_VERIFIED=PASS`.
  Production service and database changes: **NONE**.
- One-time renewal workflow **disabled after success**; next
  [run 38301470342](https://github.com/wwjd4u/MeshAlot/actions/runs/38301470342)
  **SKIPPED**.
  One-time independent verification workflow **disabled after success**;
  next [run 38301589809](https://github.com/wwjd4u/MeshAlot/actions/runs/38301589809)
  **SKIPPED**. This prevents unattended re-authorization or repeated
  production access checks.
- The new 24-hour expiration was calculated during the successful
  2026-10-12 ~20:00 UTC job, so the access is scheduled to expire
  approximately 2026-10-13 ~20:00 UTC (~3 PM US Central).
  Never treat this estimate as a durable authorization; each subsequent
  connection must first check the exact private registration expiry
  and the scope/host-key safeguards.

**Current Gate 17 state:** SSH renewal and independent authentication
verification **PASS**. Environment-file metadata for the two systemd
declarations **NOT YET RECHECKED**; secret contents and any privileged
`stat` were not accessed. The known-good encrypted rollback archive
remains a verified *partial* package. An additional read-only dependency
check must not infer elevated privileges from this authorization.
Gate 17 remains **OPEN**; no production migration/deployment/restart,
Node001/Node002 identity change, backup regeneration, or V100 rerun occurred.

## Step 3E.4 — read-only environment-file metadata inspection (2026-10-12)

**Operator instruction:** Continue the agreed Gate 17 read-only metadata
inspection, using the existing time-limited VM-specific IAP identity.
If non-elevated inspection cannot access required files, STOP; do not
escalate privilege or read secret contents.

- [One-time scoped run 38305625613](https://github.com/wwjd4u/MeshAlot/actions/runs/38305625613)
  **SUCCESS** for correct host, SSH time-limit/host-key checks, command
  execution, safe metadata-only classification and private evidence SHA-256.
  The inspected target was `meshalot-control-01` (one MeshAlot service).
- `EnvironmentFiles` declaration count: **2**, matching Step 3E.1.
  Unprivileged `lstat` metadata lookups: **0 accessible**;
  **2 ACCESS_DENIED**; **0 conclusively missing**; **0 parse faults
  or unresolved references**. Two declarations are not evidence of
  two distinct files.
- No `sudo`, credential renewal, file-content reads, Caddy reread,
  shell access beyond the approved inspection, permissions/ownership
  edits, service restart, migration, data extraction, new backups,
  recovery-key use, or production configuration changes.
- Detailed path names and metadata categories are confined to the
  existing owner-restricted `0700` MS-02 private audit directory
  in `0600` JSON and corresponding checksum files. Only sanitized
  PASS/count/classification markers were written to GitHub logs.
  `M13_STEP3E4_ENV_METADATA_CLASSIFICATION=BLOCKED_REVIEW_REQUIRED`;
  `M13_STEP3E4_SECRET_CONTENTS_READ=NO`;
  `M13_STEP3E4_PRIVILEGE_ESCALATION=NONE`;
  `M13_STEP3E4_PRODUCTION_CHANGES=NONE`;
  `M13_STEP3E4_PRIVATE_EVIDENCE_SHA256=PASS`.
- The one-time inspection job was **disabled immediately after completion**;
  its next [run 38305686863](https://github.com/wwjd4u/MeshAlot/actions/runs/38305686863)
  was independently confirmed **SKIPPED**.

**Step 3E.4 result: READ-ONLY CHECK COMPLETE / SAFE-STOP ACCESS DENIED.**
Neither the existence nor the ownership/mode of either environment-file
dependency is proven, so complete rollback readiness is still **BLOCKED**.
This is an authorization/visibility gap, **not proof of lost files or
broken production service**. The previously verified encrypted archive,
RoboForm operator confirmation, Node001/Node002 identities, historical
benchmarks, V100 evidence and Time Machine backups remain preserved.

**Decision required before Step 3E.5:** Obtain explicit, separate
operator approval for a least-privilege *metadata-only* administrative
check of the two declared paths, or choose a stage-only substitute
that does not require inspecting production file metadata.
The current user's approval does **not** authorize sudo, privilege
escalation, secret-value export, a production mutation or a staging
rollout. Gate 17 remains **OPEN**.

## Step 3E.5 — approved environment-file administrative metadata-only inventory (2026-10-12)

**Outcome: PASS for metadata-only classification; Gate 17 remains OPEN.**
Operator's "let's continue" following the explicit Step 3E.5 scoped administrative
inspection proposal was treated as permission for that *specific read-only
metadata check only*. It did not authorize new SSH identities, IAM policy
changes, production file mutations, reading environment values, or deployment.

- [Approved one-time run 38309656972](https://github.com/wwjd4u/MeshAlot/actions/runs/38309656972):
  **SUCCESS**, with pinned known-host verification via the already renewed,
  time-limited, instance-specific Google IAP SSH identity to the exact
  `meshalot-control-01` host. Private evidence was saved on MS-02 with
  mode `0600` and its SHA-256 checksum verified.
- Exactly **two systemd EnvironmentFiles declarations** were evaluated.
  **Two of two** permitted non-interactive administrative `/usr/bin/stat`
  metadata lookups succeeded using existing sudo authorization; each
  referenced path classified as a **regular file**. There were no
  unauthorized-stat, unresolved-reference, or metadata format failures.
  Two declarations do not necessarily represent two distinct physical files.
- Full underlying path names, file owners/groups, permission modes and
  sizes are contained **only** in the existing restrictive private MS-02
  evidence record. No production secret contents, environment-variable
  values, recovery keys, or decrypted archive members were read or
  printed in GitHub logs, Drive, or chat.
- Job results:
  `M13_STEP3E5_ENV_DECLARATIONS=2`,
  `M13_STEP3E5_METADATA_AVAILABLE_COUNT=2`,
  `M13_STEP3E5_ENV_METADATA_CLASSIFICATION=PASS`,
  `M13_STEP3E5_SECRET_CONTENTS_READ=NO`,
  `M13_STEP3E5_PRIVILEGED_READ=EXISTING_NONINTERACTIVE_STAT_ONLY`,
  `M13_STEP3E5_PRODUCTION_CHANGES=NONE`,
  `M13_STEP3E5_PRIVATE_EVIDENCE_SHA256=PASS`.
- The one-time administrative workflow was **disabled** on infrastructure
  branch immediately after completion. Its
  [subsequent run 38309714447](https://github.com/wwjd4u/MeshAlot/actions/runs/38309714447)
  was confirmed **SKIPPED**. No unattended repeat elevated reads are enabled.

**Resolved:** previous Step 3E.1/3E.4 inability to verify the two systemd
environment-file path entries and basic filesystem metadata. This is
*metadata verification only*, not recovery authorization or evidence
that their contents are included in the existing encrypted rollback
archive (they remain explicitly excluded).

**Remaining release blockers:** establish an operator-controlled,
encrypted, separately approved method to recover required environment
values without publishing them or storing them beside archived backups;
verify systemd/Caddy/TLS runtime dependencies and PostgreSQL global
roles/permissions; run an approved isolated stage-only rollback rehearsal;
then obtain explicit approval before any production cutover.
Preserve original DB/API archives, RoboForm Safenotes, Time Machine,
Node001/Node002 identities, M8/M9 scores, M11 controls and V100 evidence.
No such additional work was executed by Step 3E.5. **Gate 17 OPEN.**

## Gate 17 / G17-1 supplemental environment recovery archive (2026-10-13)

**OPERATOR APPROVED** limited export of only the two existing systemd
`EnvironmentFiles` declarations previously pinned by Step 3E.5,
using the **existing** API-rollback GPG passphrase. Production file
values may be accessed solely to create a separate encrypted recovery
supplement. No additional identities, new keys, VM/IAM permissions,
service changes, public storage or plaintext export authorized.

- [One-time trusted runner job 38404900223](https://github.com/wwjd4u/MeshAlot/actions/runs/38404900223):
  **SUCCESS**. MS-02 independently checked the SHA-256-protected
  private Step 3E.5 metadata report and existing 24-hour host-specific
  IAP registration. The production remote script pinned the **two**
  declared paths, required/optional flags, owner, group, mode, size,
  regular-file types and unmodified file identity, and checked
  `meshalot-control-01` hostname before content access.
- The two declared environment file *entries* were read under the
  existing non-interactive admin capability, streamed in protected
  IAP SSH, and immediately encrypted to local MS-02 restricted
  storage via GPG symmetric AES256. **No plaintext files were saved
  off the production server**. Two declarations are not by themselves
  proof of two distinct underlying files.
- The **existing**, private, mode `0600`
  `m13-g17-rollback-38168945744.passphrase` was used from MS-02.
  This is the **same API Rollback Recovery Key** that the operator
  had stored in RoboForm; it was not regenerated, copied to the
  archive directory, printed or uploaded. The original nine-member
  API/systemd/Caddy archive and the separate database archive
  remain unchanged.
- New local restricted directory:
  `$HOME/.local/share/meshalot-m13-private-rollbacks/m13-g17-env-supplement-38404900223`.
  Only the separately encrypted
  `meshalot-systemd-env-supplement.tar.gpg` and associated
  `.sha256` sidecar belong in an off-machine archive transfer.
  A **private** `env-supplement-evidence.json` is held separately
  in this MS-02 folder and must **not** be uploaded to public CI,
  Google Drive or chat.
- The new ciphertext was decrypted only in-memory in a private runner
  verification pipeline (not extracted to disk), and both member
  SHA-256 values, byte lengths, owner/group/mode, source path matches
  and source-host manifest were verified against the previously pinned
  metadata. `M13_G17_ENVSUPP_TWO_PINNED_CONTENTS=PASS`;
  `M13_G17_ENVSUPP_NEW_ARCHIVE_VALIDATION=PASS`;
  `M13_G17_ENVSUPP_PRIVATE_EVIDENCE=PASS`;
  `M13_G17_ENVSUPP_LOCAL_ENCRYPTED_ARCHIVE=PASS`.
- No plaintext values, secret names/paths from the private manifest,
  passphrases or archives were published to CI logs or Google Drive.
  No production file/service/database writes, new backup of DB/API,
  re-enrollment, benchmark rerun, or Time Machine operation occurred.
- Immediately after the single capture, the workflow was disabled;
  [next run 38405043699](https://github.com/wwjd4u/MeshAlot/actions/runs/38405043699)
  confirmed **SKIPPED**. Future export requires fresh authorization.

**Classification:** Supplemental encrypted config **capture and local
integrity PASS**, not yet independent off-machine redundancy.
`OFFMACHINE_COPY=PENDING`. The Toshiba encrypted APFS archive volume
was disconnected for safekeeping and not touched by this job.
Do **not** close G17/1 until the new supplement and checksum have been
transferred to the external encrypted APFS volume and checked.
Only then review global PG roles/grants, Caddy/TLS module/storage
dependencies and isolated staging rehearsal prerequisites.

Full production restoration, separate vault-sourced recovery practice,
and the end-to-end Milestone 13 real outage/reboot acceptance remain
unproven. Gate 17 **OPEN**.
