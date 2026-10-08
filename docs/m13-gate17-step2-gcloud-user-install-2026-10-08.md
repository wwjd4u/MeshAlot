# MeshAlot — M13 Gate 17 Step 2 — User-local Google Cloud CLI installation — 2026-10-08

**Google Cloud CLI installed and cryptographically verified. No Google authentication or production access was performed. The read-only production inspection remains PENDING, not PASS.**

## Authorization
The user expressly approved installation of Google Cloud CLI **within the existing MS-02 user account**, not system-wide, after the earlier Step 2 read-only inspection had been blocked for lack of Google Cloud/IAP tools. The user also expected a one-time browser sign-in. This authorization did not permit production migration, Caddy changes, installing services, re-enrollment, production deployment or reboot.

## Installation and verification
- Repo: `wwjd4u/MeshAlot`.
- Production `main` remains `1016c9b058690f34058dfab9cde7927e421fa469`.
- Runner: `meshalot-ms02` (host `wwjd4u-MS-02-Ultra`, user `jason_guynes`).
- [Successful GitHub Actions install run](https://github.com/wwjd4u/MeshAlot/actions/runs/37789030701), job `113351148856`.
- Official Google archive: `https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-x86_64.tar.gz`.
- Published SHA-256 verified before extraction: `73b6678401811a8a1b8d0ce29e043834308d9263ae6d9577c5aadc55621cd7d3` (official Google version 588.0.0 archive).
- Extracted into a disposable runner staging directory after sanitizing tar pathnames, self-tested the staged binary, then atomically moved to the user-local destination without overwriting pre-existing paths.
- Installed version: `588.0.0`.
- Binary: `/home/jason_guynes/google-cloud-sdk/bin/gcloud`.
- No `sudo`, `apt`, system-wide package change, Bash profile or shell PATH modification. No repo code/production branch deployment or runtime-model change.
- Workflow output: `GCLOUD_DOWNLOAD_OFFICIAL_SHA256=PASS`, `GCLOUD_STAGED_BINARY_SELFTEST=PASS`, `M13_GATE17_GOOGLE_CLOUD_USER_INSTALL=PASS`, `GCLOUD_AUTHENTICATION=NOT_ATTEMPTED`, `GCLOUD_PRODUCTION_ACCESS=NOT_ATTEMPTED`.

## Separate authentication check
- [Successful read-only auth preflight](https://github.com/wwjd4u/MeshAlot/actions/runs/37789201926), job `113351741291`.
- `MESHALOT_GCLOUD_ACTIVE_ACCOUNT=NOT_SIGNED_IN`.
- `MESHALOT_GCLOUD_PROJECT_READ_ACCESS=NOT_CHECKED`.
- `MESHALOT_GCLOUD_NO_PRODUCTION_CHANGES=PASS`.
- This status check did not print Google account names, access tokens or private credentials.

## Operator's next task (interactive)
On the MS-02 desktop open a Terminal and run:
`~/google-cloud-sdk/bin/gcloud auth login`

The user must personally complete Google OAuth in the browser using the account authorized for `meshalot-poc`. Do **not** collect or post authorization codes/tokens in chat, logs or GitHub. If the browser does not open, use the official `--no-launch-browser` flow locally with the user; do not expose its one-time authorization code to the runner logs or this conversation.

After the user confirms sign-in, run a fresh read-only MS-02 check that an active account is present (without printing private account details), and verify project `meshalot-poc` access. Only then use the already-approved and safety-tested IAP read-only audit of `meshalot-control-01`.

## Boundaries
- Gate 17 Step 2 production inspection **NOT YET EXECUTED**.
- No new SSH key yet; no IAP session yet; no project configuration or IAM changes.
- No production DB migration, API/Caddy changes, systemd service installation, internet interruption, host reboot or Mac sleep/wake action.
- Preserve Node001/Node002 enrolled identity, historic M8/M9 reports and original V100 hash evidence. No Tailscale.
