# MeshAlot M13 — Gate 17 Step 3A Readiness / Step 3B Blocker — 2026-10-08

**Step 3A PASS. Step 3B NOT EXECUTED. Step 3C NOT EXECUTED. M13 still OPEN.**

## Authorization and scope
- Operator approved a NEW private backup of the live production PostgreSQL `meshalot` database and a verified restore into an isolated test DB, preserving all existing data and configurations.
- Authorized: read-only snapshot of live DB, private local backup and disposable test restore.
- Not authorized: production data modification, schema migrations, API/Caddy/systemd restart, reboot, or changes to node identities, M8/M9 scores or V100 evidence.

## Verified Step 3A
- MS-02 self-hosted runner workflow: https://github.com/wwjd4u/MeshAlot/actions/runs/37824023544 (job `113472060388`): PASS.
- Existing IAP SSH key and fixed host target confirmed; no SSH key registration change.
- On live `meshalot-control-01`: `pg_dump` and `psql` available with an existing noninteractive PostgreSQL read-only-access path. Database size under the preflight 2 GiB ceiling.
- Production PostgreSQL **major version 16** and tracked schema migration max **8** verified again.
- Local MS-02 `gpg`, `openssl`, Docker, cached PostgreSQL 16 image and backup/restore disk headroom available.
- No production DB writes, migrations, or service changes.

## Safety stop
- An attempt to submit the narrowly scoped, user-authorized production-backup workflow through the connected execution path was refused before repository write or remote execution.
- **No production backup was created by the assistant**, no database export was streamed, no restore was performed, and no new backup encryption key was generated.
- Independently verified the `infra/ms02-runner-gateway` HEAD remained `ca872553e1341559d50349df5e01af4bbcc2b6bd`, which contains only the Step 3A preflight workflow. No Step 3B backup job appeared in GitHub Actions history.
- Production `main` remained `1016c9b058690f34058dfab9cde7927e421fa469` unchanged.
- Do not substitute an uncontrolled cloud upload or an unverified remote export to bypass this stop.

## Next supported approach
Request a direct, operator-controlled production backup via a trusted administrative session or a separately authorized supported backup service. Before execution, review the destination, encryption/recovery key handling, expected backup size, archive integrity verification, and isolated PostgreSQL16 restore plan. Never copy production data or private credentials to GitHub Actions logs, GitHub repository files, Google Drive or chat.

**Do not mark Gate 17 Step 3 complete until a real backup and isolated restore are both verified.** Maintain the one-gate-at-a-time rule.
