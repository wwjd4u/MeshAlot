# MeshAlot M13 Gate 17 — Step 1 Read-only MS-02 Readiness — 2026-10-08

**Step 1 PASS.** The next step remains explicitly gated on operator approval to inspect production host configuration. **No M13 production deployment or live service change was performed.**

## Checked revision and invariant
- Immutable candidate/source tested: `4f67ee54e5ebca084e38e306e723ae0acd08e77d` on `m13-persistent-heartbeat`.
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469` (unchanged).
- Separate MS-02 runner workflow staged on `infra/ms02-runner-gateway`, with `contents: read` only; no credentials persisted in the checked-out worktree.

## MS-02 live preflight evidence
- [Successful self-hosted GitHub Actions run](https://github.com/wwjd4u/MeshAlot/actions/runs/37764119508), job `113267436783`.
- Runner: `meshalot-ms02` on MS-02 Ultra; the job ran rather than remaining queued.
- Historical M13 Gate13–16 closeout documents and the Gate17 rollout plan were present.
- `go mod verify`, full `go test ./... -count=1`, focused Go race reconnect tests, `go vet ./...`, and a temporary `meshalot-agent` build succeeded.
- Existing Node001 identity at `/home/jason_guynes/.config/meshalot/agent-identity.json` was tested owner-matched, non-symlink, 0600, and strict parsed fingerprint matches the earlier verified values:
  - Node ID fingerprint: `2d717e508988`
  - Ed25519 public fingerprint: `1249f59eacd83ffc`
- The agent identity file's SHA-256 was compared locally before and after without printing either its hash or private key; bytes unchanged. No key generation or enrollment.
- No `meshalot-agent*` systemd unit present and no installed `meshalot-agent` binary in the runner user's PATH.
- Read-only `nvidia-smi` successfully queried GPU metadata. No V100 benchmark rerun or GPU workload started.
- Public `https://api.meshalot.com/v1/health` returned `status=ok`.
- Git worktree clean on completion. Only temporary test/build files created and cleaned up.

Runner result markers:
- `M13_GATE17_STEP1_PRODUCTION_MAIN_PRESERVED=PASS`
- `M13_GATE17_STEP1_EVIDENCE_PRESENT=PASS`
- `M13_GATE17_STEP1_BUILD_TEST_VET=PASS`
- `M13_GATE17_STEP1_EXISTING_IDENTITY_UNCHANGED=PASS`
- `M13_GATE17_STEP1_AGENT_UNIT=NOT_INSTALLED`
- `M13_GATE17_STEP1_AGENT_BINARY=NOT_INSTALLED_IN_PATH`
- `M13_GATE17_STEP1_GPU_READ_ONLY=PASS`
- `M13_GATE17_STEP1_PUBLIC_API_HEALTH=PASS`
- `M13_GATE17_STEP1_READ_ONLY_PREFLIGHT=PASS`

## Hard stop and operator approval required
Gate 17 Step 2 from `docs/m13-gate17-controlled-live-rollout-plan-2026-10-07.md` requires **explicit permission** before inspecting `meshalot-control-01` private production state. On approval, read-only inspect migration status (no user data/credentials), current API binary/release/source SHA, relevant Caddy route, systemd service/overrides, backup existence and rollback prerequisites.

This approval is not authorization to run migrations, alter Caddy, deploy binaries, install/enable an agent, disconnect Node001 or reboot either host. Those are distinct later gates and need separate approval.

Do not reenroll Node002, alter existing M8/M9 scores, modify V100 evidence, use Tailscale, or merge M13 into `main`.
