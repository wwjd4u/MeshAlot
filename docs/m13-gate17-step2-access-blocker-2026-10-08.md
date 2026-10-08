# MeshAlot — M13 Gate 17 Step 2 — Read-only Access Blocker — 2026-10-08

**Gate 17 Step 2: ACCESS CHECK PASS / PRODUCTION AUDIT BLOCKED.** The operator explicitly approved the **read-only** inspection of production `meshalot-control-01` API/systemd/Caddy configuration, tracked PostgreSQL migrations, and backup inventory. No permission was granted to install tools, create cloud credentials, modify server state, apply migrations, restart services, or reboot machines.

## Verified Git and safety state
- Production `main` unchanged: `1016c9b058690f34058dfab9cde7927e421fa469`.
- Last fully tested M13 candidate code (Gate17 Step1): `4f67ee54e5ebca084e38e306e723ae0acd08e77d`.
- M13 branch at audit script creation: `5948f9edda3b1f673c0e7b66583118897acc4c91` (only a staged read-only audit script was added this time).
- Prior Gate17 Step1 evidence: `docs/m13-gate17-step1-readonly-preflight-2026-10-08.md`; prior read-only MS-02 runner test PASS.
- No Node001/Node002 identity changes, V100 benchmark reruns or evidence modifications.

## Access probe (actual MS-02 runner)
- [MS-02 Google Cloud access-preflight workflow](https://github.com/wwjd4u/MeshAlot/actions/runs/37773588137), job `113298850994`: workflow PASS.
- `SSH_CLIENT=AVAILABLE`.
- `GOOGLE_CLOUD_CLI=NOT_INSTALLED`.
- `EXISTING_GCLOUD_LOGIN=UNKNOWN`, `MESHALOT_PROJECT_ACCESS=UNKNOWN`.
- `KNOWN_GCP_SSH_KEY=NOT_FOUND`.
- `CONTROL_SERVER_SSH_ALIAS=NOT_FOUND`.
- `M13_GATE17_STEP2_NO_PRODUCTION_CHANGES=PASS`.
- The documented historical connection to `meshalot-control-01` is the authorized `gcloud compute ssh ... --tunnel-through-iap` route. **No IAP SSH connection was attempted**, because the runner has neither that CLI nor the required authenticated access.

## Prepared and tested audit script
- Staged, NOT executed on production: `deployments/google/m13-gate17-step2-readonly-audit.sh` on the isolated M13 Git branch.
- The script refuses a host whose `hostname -s` differs from `meshalot-control-01`; never prints raw Caddyfile, systemd `Environment`, `ExecStart` arguments, credentials, tokens or private keys.
- Intended checks once approved IAP access exists: unit names/health/drop-in paths, active release symbolic links and safe binary SHA-256, source HEAD, sanitized Caddy directive presence, SELECT-only tracked migration versions/checksums, backup directory names only and backend health.
- PostgreSQL SELECT is bracketed by `BEGIN READ ONLY`/`ROLLBACK` and uses `sudo -n` so no credential prompt or SQL writes occur.
- [Safety validation](https://github.com/wwjd4u/MeshAlot/actions/runs/37773919868), job `113299933477`: syntax PASS, wrong-host refusal PASS, static safety checks PASS, Git worktree clean.
- `M13_GATE17_STEP2_HOST_ACCESS=BLOCKED_NO_IAP_AUTH`.
- `M13_GATE17_STEP2_PRODUCTION_AUDIT=NOT_EXECUTED`.

## Required next user authorization
Ask permission to **install/configure the Google Cloud CLI only on the MS-02** (ideally isolated in the existing user's home, not system-wide), followed by the operator's one-time authenticated Google IAP sign-in. This is a separate local-machine change that was not authorized by the earlier read-only server-inspection approval. Do not silently install gcloud, mint credentials, change IAM permissions, create remote server accounts or bypass IAP. After the user explicitly authorizes the local client setup and completes any required sign-in, rerun a read-only access check and then the scoped remote audit.

The later production migration, Caddy changes, release cutover, provider service installation, physical WAN disconnect, sleep/wake and reboot still require distinct approval per the controlled M13 rollout plan.
