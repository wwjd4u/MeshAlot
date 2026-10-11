# MeshAlot M13 — Gate 17: Backup, Restore, and Rollback Readiness

Date: 2026-10-11 (US Central)

**Status: PARTIAL PASS / GATE 17 OPEN.** This is a sanitized continuation
record, not permission to deploy, change production, start a service, apply
migrations, or disrupt provider nodes.

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
2. Privately inventory, protect, and hash **known-good active API binary,
   systemd unit/drop-ins, Caddy proxy configuration, and required environment
   material** without publishing their contents or restarting production.
   Compare these with the private Step 2 read-only evidence.
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
