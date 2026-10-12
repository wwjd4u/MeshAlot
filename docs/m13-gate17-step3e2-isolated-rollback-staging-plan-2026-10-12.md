# MeshAlot — Milestone 13, Gate 17, Step 3E.2
## Isolated rollback dependency checklist and staging rehearsal plan

Prepared: 2026-10-12 (US Central)
Status: **PLAN DOCUMENTED — NOT EXECUTED; Gate 17 OPEN**
Scope: documentation and operator review only. No system inspection, file extraction,
decryption, new keys, credential renewal, staging build, deployment, or test is
authorized by this document.

### Governing evidence and immutable boundaries

- Governing guide: **AI Mesh POC Build, Test, and Release Guide**, Milestone 13.
- [Gate 17 controlled rollout plan](m13-gate17-controlled-live-rollout-plan-2026-10-07.md);
  [Gate 17 backup/restore/rollback readiness record](m13-gate17-backup-restore-rollback-readiness-2026-10-11.md).
- Production `main` was independently read as
  `1016c9b058690f34058dfab9cde7927e421fa469` when preparing this plan.
  Tested M13 application candidate in the Gate 17 evidence is
  `4f67ee54e5ebca084e38e306e723ae0acd08e77d`; neither is permission
  to deploy. Production PostgreSQL is major 16, migrations **1–8 only**;
  migrations 9–10 are not applied to production.
- Gate 17 Step 3B original encrypted logical archive had GPG/SHA-256/TOC
  inspection PASS. Step 3C network-isolated PostgreSQL16 logical restore,
  original migrations 1–8 and table reads PASS, already executed; **do not
  repeat these historical tests by default**.
- The known-good encrypted API/systemd/Caddy rollback archive has **nine**
  previously verified members: one API executable, one systemd unit
  fragment, six drop-ins and the primary Caddyfile. It is a **partial**
  rollback package; environment-file contents and any external dependencies
  were intentionally excluded. No complete service rollback is proven.
- Step 3D operator Mac Terminal verified two existing encrypted archives,
  two matching SHA-256 sidecars, exactly four restricted files on the
  separate **encrypted Toshiba APFS** `MeshAlot-Backup` volume. The existing
  Time Machine `TOSHIBA EXT` volume shares that **same physical disk** and
  remains protected. Neither archive was regenerated.
- Step 3D.5: operator reports both original passphrases stored and reopened
  successfully in separate private RoboForm Safenotes. Clipboard-clear PASS.
  An independent decryption using the RoboForm copies has **not** been
  attempted. The originals remain on MS-02; keys do **not** belong with
  the archives, in this document, GitHub, Drive, logs, or chat.
- [Step 3E.1 read-only production dependency inventory](https://github.com/wwjd4u/MeshAlot/actions/runs/38295531951):
  one MeshAlot service unit, **two systemd environment-file declarations**,
  metadata for **both unverified** under available access, **zero detected
  import directives in the primary Caddyfile** under the bounded parser,
  zero parse faults. Two declarations are not necessarily two distinct
  files. Do **not** infer that any environment file is missing. No secret
  contents were published; inspection workflow disabled, rerun SKIPPED.

### Recovery dependency checklist (do not claim PASS from presence alone)

| Component | Existing evidence | Outstanding proof / disposition |
| --- | --- | --- |
| Encrypted PostgreSQL16 logical archive | Original and off-MS-02 SHA-256 matches; one prior isolated logical restore PASS | Document exact *later* recovery recipe, global database roles, owners, ACLs, extensions, grants and dependent runtime identities. A `--no-owner --no-privileges` test does not validate these. |
| Encrypted known-good API executable | In nine-member encrypted rollback archive, original hash-pinned | In authorized **isolated** environment, prove artifact extraction and executable/architecture compatibility without touching live executable |
| Systemd unit fragment and six drop-ins | Encrypted, archived, checked against private fingerprints | Identify `User`/`Group`, `ExecStart`, `WorkingDirectory`, ordering, required paths, permissions and override precedence; staging-only `systemd-analyze verify` and start/stop behavior |
| Systemd environment-file declarations | Two declarations identified | **BLOCKED:** classify each underlying distinct path, whether optional/required, owner/mode/existence and safe recovery mechanism. Contents intentionally not archived; no path/content in public evidence |
| Primary Caddyfile | Encrypted and hash-verified; inspected bounded import directives = 0 | In isolated context validate syntax, route/upstream and WebSocket behavior. TLS certificate storage, plugins/modules and additional runtime/config sources not established |
| Caddy imported file material | No imports detected in inspected primary file | Confirm via approved runtime/config-dependency review when justified; avoid treating primary-file static parsing as complete dependency proof |
| Runtime credentials / environment secrets | None exported or examined under Step 3E.1 | **BLOCKED:** operator-approved separate encrypted escrow/recovery strategy; exclude secrets from GitHub, Google Drive, CI jobs, logs and shared staging |
| Database-to-API compatibility | Prior production baseline migrations 1–8; synthetic migration rehearsal 8→9→10 PASS | Staging acceptance must not use real node enrollments, real account sessions or existing Node001/Node002 identities; separately prove role/grant compatibility before any full rollback claim |
| Endpoint/TLS, OS ownership, file modes | Prior loopback health and selected production hashes | Stage equivalent loopback routing, restrictive permissions, service health, and failure rollback; no production IP/DNS, GCP/IAM, or certificate changes |
| Recovery-key availability | Two RoboForm Safenotes operator-verified; original passphrases private on MS-02 | Separately authorized later recovery drill using vault-sourced keys without exposing them; do not repeat prior MS-02 archive validation |
| Time Machine, V100 and node histories | Preserved as governing constraint | No changes to Time Machine data, Node001/Node002 identities or M8 report, M8/M9 benchmarks, V100 evidence, provider controls, Ollama/Hermes, backups or live release |

### Proposed isolated staging recovery rehearsal (NOT authorized to execute)

**Phase A — scope and safety preflight.**
1. Operator explicitly approves the **target staging environment**, physical/network
   boundaries, duration, resources, cost and whether production-derived archive
   material is permitted there. Prefer an **ephemeral dedicated local VM** or
   equivalently isolated host; do not use the live control server, current MS-02
   identity/config paths, or Mac's Node002 enrollment as disposable staging.
   A new cloud VM, domain, IAM or externally reachable endpoint requires separate
   approval. Do not expose a provider inbound port or host shell.
2. Prepare a disposable VM/network without route to the production database,
   control plane, agent identities, internet-facing listener, privileged host
   mounts or production credential stores. Use host-only/loopback interfaces.
   Record stage-only identifiers and allocate enough disk/RAM before proceeding.
3. Establish a **private evidence ledger** inside the approved stage: timestamps,
   explicitly authorized source artifact IDs and their original digest references,
   stage image/version, operator acceptance, test results, teardown. No customer
   data, decrypted contents, key values, real node identities or secrets in logs.

**Phase B — dependency gate (blocking).**
4. Review the two environment-file declarations with a separately approved
   metadata-only privilege scope if genuinely needed. Record only sanitized
   dependency counts and classification in shared records; store exact paths,
   permissions and ownership exclusively in operator-controlled private evidence.
   Determine how required private values can be supplied to a fresh isolated
   instance without copying production secrets into GitHub or cloud runner jobs.
5. Reconcile primary Caddyfile parse with permitted Caddy runtime/adapter
   configuration, upstream, TLS and storage expectations. Check for additional
   imported configs or runtime modules by **metadata only** before staging
   archive replay. Identify any environment-dependent route or cert behavior.
6. Classify PostgreSQL16 global roles, extensions, grants, ownership and
   production-to-stage dependencies; stop if actual restore would require
   production credentials, privilege escalation or unapproved data copies.

**Phase C — rehearse without production data first.**
7. Use **synthetic** identities and a fresh PostgreSQL16 fixture matching
   tracked schema version 8. Do not repeat the already-passed production-data
   logical restore. No live database connection.
8. Under separate explicit approval for actual archive use, retrieve the
   *existing* encrypted rollback copy through a private operator-controlled
   path (never CI logs, Drive, or public artifacts), verify **on receipt** against
   its original digest and decrypt into a disposable restricted stage only.
   Recovery passphrase comes only from the operator's separate vault; never
   pass it on a command line, in environment variables, a shell history,
   a GitHub secret or shared staging log. No unattended runner secret access.
9. From the existing private recovery manifest, stage the known-good API
   executable, unit and drop-ins, and primary Caddyfile at **remapped stage
   paths**, never `/etc/systemd`, `/etc/caddy` or live release paths of any
   operational host. Preserve and verify mapped ownership/mode constraints.
   Configure dummy stage-only environment values and local test credentials,
   not exported production credentials.
10. Validate stage service definition and Caddy configuration statically, then
    test disposable service startup, loopback health, API-to-PG16 interaction,
    authenticated negative cases, safe shutdown and restart in the isolated VM.
    Include WebSocket proxy upgrade only in the isolated/staging context;
    disconnect must not affect live nodes.

**Phase D — fail safely and certify evidence.**
11. Rehearse an intentionally broken stage-only binary/configuration switch
    and rollback to the staged known-good artifact, recording that isolated
    health is restored and original synthetic DB history remains intact.
    Do not blindly restore a live database or roll back applied migrations.
12. Confirm all acceptance checks, archive digests, known-good executable
    fingerprint, restricted permissions, staging-only service endpoints,
    no customer/node credential use and no escape route to production.
    Record dependencies still unresolved as **BLOCKED**, not PASS.
13. Shut down and clean only disposable stage resources after operator review;
    retain a redacted report and approved private evidence. Do not touch
    original encrypted archives, Time Machine, original recovery keys, Node001,
    Node002, historical benchmarks or V100 hashes.

### Entry gates, stop conditions, and deliverables

**Entry gates:** specific operator approval for a stage target and any cost;
metadata classification of environment files and required Caddy/runtime
dependencies; clear staging privilege/recovery-key boundaries; a stage-only
rollback procedure; unmodified archived artifact checksums.

**Immediate STOP if:** privilege/ownership requirements cannot be satisfied,
any required environment file is unknown, unexpected systemd/Caddy dependency
appears, stage can route to production or enrolled nodes, key or decrypted
config risks leaking to logs, any archive hash differs, a credential would be
exported to shared services, schema history deviates, or new IAM/host access
would be needed without approval.

**Evidence outputs (future, not claimed now):**
- Private approved-stage inventory/manifest and receipts for only authorized
  artifacts; SHA-256 evidence; access mode and isolation verification.
- Sanitized checklist showing static unit/Caddy validation, synthetic
  application/DB checks, controlled fault/rollback and final cleanup.
- Unresolved dependency register, explicit operator sign-off and go/no-go
  recommendation. Only after a successful approved rehearsal can the
  *staging* rollback rehearsal be marked PASS; it still does not authorize
  a production cutover.

### Current decision

**STEP 3E.2 = PLANNING COMPLETE ONLY.** No staging environment was created,
no new production inspection executed, no archive decrypted or re-tested,
no key accessed, no production change or test rerun, and no M13 release
authorization granted. **Gate 17 remains OPEN.**

**Next single approval checkpoint:** resolve the two unverified
systemd environment-file declarations via a separately reviewed minimal
read-only metadata method, or design a fully equivalent stage-only
dependency substitute; do not escalate privilege or copy secret contents
without express approval.

### Addendum — Step 3E.5 approved administrative file metadata (2026-10-12)

[Step 3E.5 metadata-only evidence](https://github.com/wwjd4u/MeshAlot/actions/runs/38309656972)
verified that both declared environment-file references resolve to regular
files and that only non-secret filesystem metadata was accessible through
pre-existing, non-interactive `stat` privileges. Paths and their
owner/mode/size metadata remain confined to the private MS-02 report.
The one-time workflow was disabled; the later execution was SKIPPED.

**Checklist delta:** existence/type/metadata visibility for both declared
environment-file references has moved from UNVERIFIED to
**PASS (metadata only)**. Their *contents, actual recovery material,
separate-key custody, systemd service restoration and application startup
in an isolated staging environment remain NOT VERIFIED*.
Historical entries above are preserved as they stood when the plan was
authored; this addendum supersedes only the metadata-visibility blocker.
This is not a staging rehearsal, an escrow approval, or a live-cutover
authorization. Gate 17 is OPEN.
