# MeshAlot M13 Gate 17 Step 2 — OAuth Verified; VM Audit Requires SSH Key — 2026-10-08

**STATUS: GCP AUTHENTICATION AND VM METADATA READ PASS; HOST CONFIGURATION AUDIT NOT RUN.** No production application, database, service or Caddy change.

## Operator and scope
- User confirmed on 2026-10-08 that `gcloud auth login` was completed in the Ubuntu Terminal on the MS-02.
- Previously approved read-only inspection of `meshalot-control-01`. No permission yet to create/register SSH keys, alter GCP project/instance SSH metadata, modify IAM, migrate DB, deploy API, start services, or reboot.
- Production `main` at inspection: `1016c9b058690f34058dfab9cde7927e421fa469`, unchanged.
- M13 development branch at start of evidence write: `196af04685e044be513af5552d892aa814256010`; last tested M13 code candidate remains `4f67ee54e5ebca084e38e306e723ae0acd08e77d`.
- User-local `gcloud` CLI v588.0.0 installed in `/home/jason_guynes/google-cloud-sdk/bin/gcloud`, no shell/system package changes.

## Verified auth and project access
- Existing read-only MS-02 validation re-run: https://github.com/wwjd4u/MeshAlot/actions/runs/37792288758, latest successful attempt, job `113455652498`.
- `M13_IAP_AUTH_ACTIVE=PASS`.
- `M13_GCP_PROJECT_READ=PASS` for `meshalot-poc`.
- `M13_CONTROL_INSTANCE_READ=PASS` for `meshalot-control-01`, zone `us-south1-a`.
- `M13_EXISTING_GOOGLE_SSH_KEY=NOT_PRESENT`.
- `M13_IAP_SSH_TUNNEL=NOT_ATTEMPTED` and `M13_NO_SERVER_OR_PRODUCTION_CHANGES=PASS`.
- Google account email and access/refresh tokens were never printed in logs.

## Verified read-only Google Compute VM metadata
- Run: https://github.com/wwjd4u/MeshAlot/actions/runs/37819357162; job `113456094813`.
- Google Compute Engine API returned VM status `RUNNING`.
- OS Login in available project/VM metadata: `UNSET` (do not infer organization-policy defaults from this alone).
- Instance-level SSH key metadata: `ABSENT`.
- Project-level SSH key metadata: `PRESENT`; values and usernames were not printed.
- `block-project-ssh-keys`: `UNSET`.
- VM has a private network interface; a public address also exists, but IAP remains the prescribed remote-access path. No direct SSH was attempted.
- MS-02 `~/.ssh/google_compute_engine` private key: `NOT_PRESENT`.
- `M13_VM_SSH_POLICY_READ_ONLY=PASS`, `M13_CONTROL_VM_SSH=NOT_ATTEMPTED`, `M13_CONTROL_VM_METADATA=NOT_CHANGED`.

## Required next permission
The existing Google Cloud project already has SSH access metadata, but MS-02 has no corresponding private key. Before using IAP SSH, ask the user for **separate explicit authorization** to:
1. Generate a dedicated Ed25519 key pair **locally on MS-02** with restrictive filesystem permissions. Never send the private key to GitHub, Google Drive, logs or chat.
2. Register its **public** key in a narrowly scoped way for `meshalot-control-01` (prefer only instance-level, never overwrite or delete existing project keys). This is a GCP metadata/access-control write and was NOT covered by the read-only server-audit approval.
3. Use the newly authorized key only for the previously approved read-only metadata audit through the IAP route, without sudo modifications or service restarts. Verify the access path and the hostname before reading host configuration.

After the user approves, first inspect any existing metadata/OS Login restrictions safely and design an additive operation with explicit rollback. Stop if the effect cannot be confined to the intended VM. The staged audit script `deployments/google/m13-gate17-step2-readonly-audit.sh` has passed wrong-host and syntax validation but has **not** run on production.

**Never** claim Gate 17 Step 2 PASS until the actual host service, Caddy, database migration status and backup inventory are verified. No M13 production deploy, Node001/Node002 re-enrollment, M8/M9 history mutation, V100 benchmark change, or Tailscale.
