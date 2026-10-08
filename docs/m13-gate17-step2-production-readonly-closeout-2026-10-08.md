# MeshAlot M13 — Gate 17 Step 2 Closeout — 2026-10-08

**Status: PASS — live production control-server configuration inspected read-only via authenticated Google IAP SSH.**
**Milestone 13 remains OPEN. This closeout is NOT approval to deploy, migrate, restart, interrupt WAN, or reboot.**

## Authorized access and privacy
- User explicitly approved generating a dedicated local MS-02 SSH key and registering its public half specifically on the VM `meshalot-control-01`. The user expressly requested confidentiality of project data.
- Dedicated Ed25519 private key remains **only** in a mode-0700 user-local directory on MS-02, with mode-0600 key. No private keys, access tokens, credential details, Caddy contents, raw DB records, user SSH names, or active API binary paths/hashes were uploaded to GitHub or Google Drive.
- Google Compute Engine instance `ssh-keys` metadata received **one dedicated time-limited public key** for the existing authorized Linux SSH username, with 24-hour expiration. The existing *project-wide* SSH keys and all unrelated VM instance metadata were compared before/after and unchanged. No new IAM role or VM instance was created.
- Host authentication used the intended Google IAP tunnel, and the SSH host key was stored privately on the MS-02 for strict verification on subsequent connections.
- All detailed audit reports and SHA-256 checksum files remain exclusively in a private mode-0700, user-local MS-02 directory. Only sanitized status/counts were logged to GitHub Actions. Backup archive contents and production configuration were not copied.

## Git branch and accepted candidate
- Production `main` unchanged after all checks: `1016c9b058690f34058dfab9cde7927e421fa469`.
- M13 gate17 development branch baseline for read-only scripts: `7e43a7334e3a7c58786ff80eb2cb550a42071a03`.
- The previously validated M13 code candidate is `4f67ee54e5ebca084e38e306e723ae0acd08e77d`; subsequent M13 branch changes to date have been operational audit scripts or evidence files, not a production rollout.
- Separate runner infrastructure branch contains all temporary workflows; none were added to production `main`.

## Ordered verification
1. **SSH safety preflight:** https://github.com/wwjd4u/MeshAlot/actions/runs/37821623956. VM OS Login metadata not explicitly enabled; VM-specific `ssh-keys` initially absent; project-level keys preserved. Two plausible existing usernames were found, so registration paused for account match resolution.
2. **Existing user match:** https://github.com/wwjd4u/MeshAlot/actions/runs/37821760082. The authenticated Google account's local part matched a pre-existing GCE SSH metadata username. No username/email or key material was printed.
3. **Dedicated key registration:** https://github.com/wwjd4u/MeshAlot/actions/runs/37821919435. Key-pair consistency, restrictive local permissions, 24-hour metadata expiry, instance-only registration and preservation of other metadata: PASS.
4. **Google IAP SSH host verification:** https://github.com/wwjd4u/MeshAlot/actions/runs/37822118540. Authenticated SSH to exact expected `meshalot-control-01` host PASS. Its SSH host key was stored on the MS-02 and rechecked strictly in all further sessions.
5. **First private server audit:** https://github.com/wwjd4u/MeshAlot/actions/runs/37822331815. Production host name confirmed; one MeshAlot service identified; Caddy reverse-proxy/API route/loopback references present; PostgreSQL tracked versions 1–8 contiguous; four existing backup directories inventoried. The initial assumed local port was not the live Caddy upstream; release symlinks were not at the two initially checked paths.
6. **Service health and proxy follow-up:** https://github.com/wwjd4u/MeshAlot/actions/runs/37822655699. Actual MeshAlot service active; two Caddy loopback targets examined and one correctly returned HTTP 200 at `/v1/health`. Service binary wasn't readable through the nonprivileged `/proc` route, so further read-only verification was needed.
7. **Final binary and migration verification:** https://github.com/wwjd4u/MeshAlot/actions/runs/37822857351. Privileged *read-only* inspection verified active API binary path and SHA-256, preserving exact details in private local report only. Production PostgreSQL migration entries 1–8 exactly matched the SHA-256 checksums of the committed source migration files. **Production migrations 9 and 10 are NOT APPLIED.** The private report checksums passed.

## Deployment consequences and next gate
- **Gate 17 Step 2 is complete**, but M13 is not deployed.
- Production is at tracked migration version **8**, not version 9. The existing Gate16 isolated database test covered version9→10; any future production release must account for BOTH unapplied migrations **9 and 10**. Do not skip version 9 or rewrite applied migration history.
- Four backup directories were found, but their existence alone is NOT a tested, restorable database backup. A verified private backup and restore test remains required before any release.
- Production API and Caddy were functioning during the audit; no binary, configuration, deployment, migration or service restart occurred.
- The only intentional production-side access-control change during this step was the user-authorized instance-specific SSH public key (expires after 24 hours). The expired metadata entry may still remain until separately removed; **do not delete it without authorization**.
- **NEXT — Gate 17 Step 3:** obtain explicit user approval to capture a new private, recoverable PostgreSQL database backup and known-good release/config rollback material, test restore into a disposable isolated DB (never into production), and preserve private evidence. Obtain a separate approval for any production deployment/migrations and separate approval for disruptive Node001 WAN/reboot and Node002 Mac sleep/wake tests.
- Never re-enroll Node001/Node002, alter historical M8/M9 scores, rerun/overwrite original V100 evidence, use Tailscale, or expose remote host shell to arbitrary execution.
