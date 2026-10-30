# MeshAlot current milestone status

This file is the concise current-status pointer. The Google Drive **AI Mesh POC Build, Test, and Release Guide** remains the source of truth for milestone requirements and release gates.

- Milestone 0 — **PASSED**
- Milestone 1 — **PASSED**
- Milestone 2 — **PASSED**
- Milestone 3 — **PASSED**
- Milestone 4 — **PASSED**
- Milestone 5 — **PASSED**
- Milestone 6 — **PASSED**
- Milestone 7 — **PASSED**
- Milestone 8 — **PASSED**
- Milestone 9 — **PASSED**
- Milestone 10 — **PASSED**
- Milestone 11 — **PASSED**
- Milestone 12 — **PASSED**
- Milestone 13 — **IN PROGRESS — GATES 1–16 VERIFIED IN ISOLATED TESTS; LIVE ACCEPTANCE PENDING**

## Hands-free G17 preparation — archive key-access tooling

The operator is traveling; Mac/vault operations are deferred. The new
[local archive key-access verifier and evidence](m13-g17-archive-key-verifier.md)
automate digest checks, discard-only authenticated decryption and redacted
reporting. **13 focused synthetic local tests PASS; 13 pinned MS-02 tests PASS
with zero skips** ([run 41108714209](https://github.com/wwjd4u/MeshAlot/actions/runs/41108714209)).
No real backup/key was used. Actual Mac/vault
acceptance is **NOT RUN**, and Gate 17/1 remains **BLOCKED**.

## Gate 17 continuation — October 29, 2026

Gate 17/1 remains **BLOCKED**; Gate 17/2–7 are **NOT STARTED**. The operator's
October 27–28 Caddy console results are consolidated in
[the metadata checkpoint](m13-g17-caddy-metadata-checkpoint-2026-10-29.md).
They establish selected service/binary/listener/directory metadata and support
default automatic HTTPS for the inspected API route; renewal, TLS-state restore
and complete recoverability remain unverified. The timed-out journal query is
inconclusive and is not a reason to rerun passing tests.

Routine testing/configuration/preparation is authorized without repeated
continuation approval; production pushes and material-risk decisions remain
separate. The rejected PostgreSQL diagnostic stays paused. The next independent
recovery proof is [vault-key availability against the existing encrypted copies](m13-g17-vault-recovery-key-check-2026-10-29.md),
using the Mac's native encrypted-APFS support. The operator reports the drive
connected; actual mount/unlock and GnuPG availability still need verification.
No real archive has been decrypted by this continuation record.

[Automated MS-02 runner readiness](m13-g17-runner-readiness-2026-10-29.md)
**PASS**: the existing trusted gateway executed the reviewed read-only metadata
job at infrastructure SHA 81e78c042e3dcc4df75a1a8c2328e11a28f92653. Capacity is
sufficient for planning; local Docker is accessible. QEMU was not found on PATH,
so full VM-stage suitability/isolation remains **NOT TESTED**. Suitable automated work
belongs on the runner; the Mac-attached encrypted drive and private vault entry
remain operator-local steps.

## Gate 17 recovery preparation — October 26, 2026

Gate 17/1 remains **BLOCKED** on actual PG16 global-role/ownership recovery and
Caddy/TLS runtime dependencies; Gate 17/2–7 remain **NOT STARTED**. The new offline
synthetic bundle integrity preflight passed 12 focused local tests and is
included in future PR/main CI. It does not execute restored artifacts, inspect
production or establish recovery acceptance. See
[the implementation and evidence record](m13-g17-offline-bundle-preflight-2026-10-26.md).

## Milestone 4 reconciliation

The older Milestone 4 section in `docs/milestones.md` preserves evidence from an earlier point in the build and is stale where it says process-restart verification, restricted-login runtime verification, or controlled production deployment are pending. Do not interpret that older status line as the current gate result.

Final Milestone 4 verification includes the production `meshalot` PostgreSQL database, tracked migrations through `000003_node_persistence.up.sql`, restricted runtime permissions, ledger protections, PostgreSQL-backed enrollment/heartbeat persistence, actual API process-restart verification, controlled live deployment, live service-restart persistence, and completed CPU/memory/disk/HTTPS monitoring with notification delivery verified.

## Milestone 5 completion

Milestone 5 passed on September 4, 2026.

Production release commit: `9631776d5cbb07903c3f018b49cf9304c5e8200b`

Production release ID: `9631776d5cbb-20260904-162618`

Verified completion includes migration 4, narrowed runtime database permissions, real restricted-login and account-separation tests, production sign-in/sign-out, secure browser sessions, user-scoped dashboard/node/wallet/job APIs, existing-node heartbeat persistence, the React dashboard and SPA routes behind Caddy, security-header verification, successful local/public health checks, preserved rollback material, and final browser verification by the user.

See `docs/milestone5.md` for the detailed Milestone 5 implementation and release-gate evidence.

## Milestone 6 completion

Milestone 6 passed on September 4, 2026 and is operationally closed.

Release source commit: `bb7c75b4cd0797f68fb4b8a5f26c3f2de26b3a65`

Production release ID: `bb7c75b4cd07-20260904-183650`

Verified production backend SHA256: `620db365acde4660e3d95124b5c6ee7462d75c29ed17a7d580d635c671395b5d`

Verified completion includes locally generated persistent UUIDv4 and Ed25519 node identity, restrictive local private-identity permissions, public-key-only secure enrollment, short-lived one-time website enrollment codes stored only as hashes server-side, atomic anti-replay redemption, account isolation, concurrent double-redeem protection, invalid/expired/fabricated/consumed-code rejection, M6 migration and least-privilege runtime grants, the secure My Nodes enrollment UI, isolated enrollment performance testing, controlled production release with M5 rollback preserved, and a real Node 001 enrollment/reload/replay test using the Ubuntu Minisforum MS-02 Ultra.

Final production verification confirmed the real Node 001 as enrolled with M6 agent version `0.1.0-m6`, a public cryptographic identity present, the one-time token consumed and node-bound, migration 5 verified, M6 runtime privileges verified, the exact M6 backend live, the M6 systemd release override effective, and local/public API plus public web health all HTTP 200.

Operational source closeout then fast-forwarded `/opt/meshalot/source` to the exact M6 release commit without restarting the live M6 service, changing the database, or changing Caddy. The previous source ref remains preserved as local branch `backup/pre-m6-source-20260904-195308`; health remained HTTP 200 throughout.

Preserved M5 rollback directory: `/opt/meshalot/backups/bb7c75b4cd07-20260904-183650-cutover-20260904-191501`

See `docs/milestone6.md` for the detailed Milestone 6 implementation and release-gate evidence.

## Milestone 11 completion

Milestone 11 — Provider Resource Controls passed final validation on October 5, 2026.

Final validation ran on the trusted `meshalot-ms02` GitHub self-hosted runner using Go 1.24.13. The targeted provider-control tests passed, the full repository test suite passed, branch scope matched the intended four M11 files, and the runner worktree was clean.

M11 implementation merge SHA: `2165afc7927b79028ebf9c37219a360c2d7c604a`

See `docs/milestone11.md` for the detailed validation and closeout evidence.

## Next work

Milestone 12 is closed. Milestone 13 (Persistent Agent Connection and Heartbeat) has completed code, isolated PostgreSQL, WebSocket/TLS, reconnect, Linux/macOS packaging, real Node001 read-only identity/local telemetry, reverse-proxy and migration-upgrade validations through Gate 16. These are **not** evidence of real production WSS deployment, physical WAN/sleep interruption or reboot recovery. M13 remains open.

Gate 17 is a **controlled live rollout with explicit operator approval**. See `docs/m13-gate17-controlled-live-rollout-plan-2026-10-07.md` and individual `docs/m13-gate*-closeout-2026-10-07.md` evidence. Do not migrate or restart production, re-enroll Node002, or rerun V100 benchmark evidence until the approved plan requires it.

**October 13 fixed execution sequence:** `docs/m13-gate17-fixed-completion-sequence-2026-10-13.md` is the single operational checklist for G17/1–G17/7. Read-only remote MS-02 staging capacity preflight PASS (14 logical CPUs, 117 GiB available RAM, 625 GiB available disk, Docker daemon accessible). With separate approval, isolated PostgreSQL16 staging bootstrap PASS: network none, 2 CPUs, 4 GiB RAM, synthetic data and restricted synthetic roles, no production connection; disposable container and temporary storage cleaned. Evidence: https://github.com/wwjd4u/MeshAlot/actions/runs/38419118910 . Full service rollback rehearsal has NOT been executed. Supplemental AES256-encrypted systemd environment-file archive was created locally on MS-02 and independently SHA-256-verified on the external encrypted Toshiba volume by operator Mac Terminal; the two earlier archives remain preserved. Full *recovery rehearsal* (including PostgreSQL global roles/grants and Caddy/TLS dependencies), production deployment, real Node001 network/reboot and Node002 sleep/wake acceptance remain incomplete. Production `main` and existing source backups unchanged. Complete Gate 17 in order, without repeating verified tests or bypassing operator approvals.

**G17/1 permission checkpoint (2026-10-13):** [Isolated PostgreSQL 16 runtime ACL run](https://github.com/wwjd4u/MeshAlot/actions/runs/38426678379) PASS. Actual tracked M13 runtime grants were exercised against synthetic stage roles/data; authorized operations succeeded and eight unauthorized statements returned SQLSTATE 42501. Stage resources cleaned, one-time workflow disabled, production untouched. PostgreSQL cluster-global role restoration and Caddy/TLS recovery still need stage acceptance. Gate 17 OPEN.

**G17/1 synthetic ownership recovery checkpoint (2026-10-13):** [Fresh PostgreSQL 16 cluster reconstruction run](https://github.com/wwjd4u/MeshAlot/actions/runs/38440768518) **PASS**. Source and clean destination were independent, disposable, and sequential. Synthetic global nonadmin migration owner/runtime roles, database/table/function ownership, the 10-entry migration ledger, synthetic record and restricted grants survived export/restore; unauthorized operations remained blocked with SQLSTATE 42501. No production globals or archives were used, cleanup PASS. The one-time workflow was disabled and [next run SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38440900971). **Real production global-role recovery and Caddy/TLS/systemd full restoration remain open; G17/1 IN PROGRESS; G17/2 NOT STARTED; production main unchanged.**

**G17/1 approved Caddy/TLS metadata inspection (2026-10-13):** [Read-only MS-02/IAP run](https://github.com/wwjd4u/MeshAlot/actions/runs/38442058347) **PASS**. Existing access reused; production Caddy service active and installed CLI **v2.11.4**. Module query reported 0 nonstandard modules; primary Caddyfile had 1 bounded-parser reverse_proxy and 0 detected import, explicit TLS/storage or manual-key directives. Certificate data root **UNVERIFIED**, not confirmed missing; no private keys/certificates or secrets accessed. No production or access changes. One-time workflow disabled and [next run SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38442168184). **Caddy/TLS recovery is NOT yet proven; G17/1 OPEN, G17/2 NOT STARTED.**

**G17/1 Caddy pinned-image preflight (2026-10-13):** Official Docker Hub published linux/amd64 Caddy 2.11.4-alpine manifest `sha256:040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0` ([source](https://hub.docker.com/layers/library/caddy/2.11.4-alpine/images/sha256-040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0)). [MS-02 read-only staging preflight](https://github.com/wwjd4u/MeshAlot/actions/runs/38442965161) PASS; Docker/SSL tooling present, no matching image or native binary cached. One-time preflight disabled, [next execution SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38443288594). **Digest recorded only**; no image downloaded, stage started or production accessed. Need approval for future pinned image pull/synthetic HTTPS/WSS staging; TLS storage/recovery remains unproven. G17/1 IN PROGRESS, G17/2 NOT STARTED.

Milestone 10 remains closed with workload-aware eligibility preserving separate Compute, Network, Reliability, Availability, and Trust dimensions.

A parallel MeshAlot Brand/UI track is approved using the supplied network artwork as the visual direction. Brand work should remain separately auditable from technical milestone work.

## Archive coverage

`MeshAlot-Verbatim-Transcript-Part-2026-09-04.txt` is a partial transcript. The later alert-setup conversation and subsequent handoff were not yet saved to Drive when Milestone 5 began. Do not describe the archive as complete.
