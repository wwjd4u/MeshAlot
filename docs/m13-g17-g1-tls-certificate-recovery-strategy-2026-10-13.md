# MeshAlot M13 — Gate 17/1: Caddy TLS certificate recovery decision and safe runbook

Date: 2026-10-13 (US Central)  
Status: **RECOVERY STRATEGY DOCUMENTED; ACTUAL PRODUCTION TLS RECOVERY UNVERIFIED — G17/1 IN PROGRESS**  
Decision owner: operator. Scope authorized: sanitized planning and GitHub documentation only, **not** production access, certificate issuance, credential use, archive decryption, new infrastructure, or deployment.

**October 29 continuation:** the scope line and observations below describe the
October 13 checkpoint. Later operator-supervised, metadata-only Caddy results
are recorded in [the consolidated October 29 checkpoint](m13-g17-caddy-metadata-checkpoint-2026-10-29.md).
Selected service/binary/data-directory/admin-listener fields passed; the inspected
HTTP route supports default automatic HTTPS. Actual renewal, issuer/account,
storage durability and protected TLS-state restoration remain unverified.
Routine testing/configuration/preparation is now authorized without repeated
continuation prompts; production pushes and material-risk decisions remain
separate. This update does not authorize forced issuance or restart a service.

## Decision

**Primary recovery path: allow Caddy to obtain a fresh publicly trusted certificate for `api.meshalot.com` only during a separately approved real recovery/cutover, using its existing validated automatic-HTTPS configuration and issuer policy.** Reissue instead of copying a live TLS private key through remote-runner/public storage.

**Fallback: recover the existing persistent Caddy storage from a separately approved, protected, encrypted and restore-tested private backup when the primary path cannot be safely completed.** This is a planned fallback, **not a presently verified backup**. The current nine-member API/systemd/Caddy rollback archive contains the primary Caddyfile, **not** a verified backup of Caddy's persistent TLS data directory. A production TLS private-key backup must never be inferred from the presence of the Caddyfile, or from the existing encrypted systemd environment supplement.

This is a **recovery strategy**, not authorization to delete Caddy state or to force renewal now. In normal production operation, **keep existing, persistent Caddy storage intact and let Caddy renew normally**.

## Observed MeshAlot facts versus unknowns

| Requirement | Evidence classification |
| --- | --- |
| One running production Caddy service; inspected Caddy CLI v2.11.4; primary service Caddyfile; 1 reverse proxy; bounded scan found 0 imports and 0 explicit TLS/storage overrides | Read-only inspection [38442058347](https://github.com/wwjd4u/MeshAlot/actions/runs/38442058347) PASS, limited to inspected metadata. This does not prove the running binary digest or fully adapted configuration. |
| Real Caddy v2.11.4 synthetic HTTPS + authenticated WebSocket and break/recover | Isolated [38449591265](https://github.com/wwjd4u/MeshAlot/actions/runs/38449591265) PASS; [one-time workflow disabled; follow-up SKIPPED](https://github.com/wwjd4u/MeshAlot/actions/runs/38449988211). Synthetic certificates only. |
| External encrypted backup copies of PG16 data, known-good API/systemd/Caddy components, and systemd environment supplement | Three GPG archives + .sha256 operator-verified off MS-02, with two recovery passphrases in separate RoboForm Safenotes. **Caddy TLS storage NOT known to be included.** |
| Actual production certificate issuer/account, active `CADDYPATH`/XDG data root, ACME challenge choice, service UID/GID and storage permissions | **UNVERIFIED**; do not assume the candidate directory is the actual data root. |
| DNS A/AAAA and CA reachability from the *recovery target*, challenge reachability on TCP 80 (HTTP-01) or TCP 443 (TLS-ALPN-01), and storage persistence | **NOT VALIDATED for future recovery target**. No firewall/DNS adjustment is authorized by this document. |
| Restore of live TLS storage from separately encrypted backup, or actual fresh ACME issuance after disaster | **NOT PERFORMED**. |

## Prerequisites for a future approved real recovery

1. Obtain **explicit separate operator authorization** for exact target, maintenance window, service change, DNS/firewall changes if needed, expected costs and rollback/stop conditions. No production connection solely on the strength of this document. Verify the current release source SHA and existing known-good recovery artifacts first; do not change original Node001/Node002 identities.
2. With separately approved **metadata-only** access, confirm the *actual* service executable, Caddy service user, effective/adapter config source, hostname, active storage module/root and whether certificates are auto-managed versus manually loaded. Inspect directory existence/ownership/modes without reading, printing, copying or enumerating private certificate/key **contents** into CI, Drive or chat. If additional storage/module/issuer dependencies exist, STOP and update the plan.
3. Confirm public DNS `A` and `AAAA` resolve only to the **intended recovery endpoint**; confirm both advertised address families route correctly. An unreachable AAAA record can break validation. Ensure the relevant standard ACME challenge can reach Caddy: **HTTP-01 requires public TCP 80; TLS-ALPN-01 requires public TCP 443**. Check existing firewall/security-group rules without changing them absent approval. For DNS-01, confirm it is deliberately configured and its DNS-provider dependency is available **without exporting its credentials**.
4. Confirm the recovery target has the correct matching Caddy configuration, minimal service permissions required to bind/listen, outbound reachability for the configured CA, correct time, and **persistent writable Caddy data storage** owned appropriately. Do not use a throwaway directory for live TLS state and do not delete the original Caddy data directory or ACME account state.
5. Confirm application loopback upstream is healthy and that production API/session authorization will not become inadvertently exposed while TLS is pending. Keep the original release, DB schema and services in their previously verified states until separately approved cutover.

## Primary procedure — reissuance during an authorized recovery ONLY

- Prepare fresh service/systemd/Caddy dependencies on **an independently approved target** with safe rollback of the existing API. Do not replay private archive contents on MS-02/CI by default and do not expose any production credential to public runners.
- Avoid a test run against the real `api.meshalot.com` issuer. For **synthetic** rehearsals, use Caddy local certificates or an ACME staging endpoint with a controlled stage-only domain and no real production DNS. Never repeatedly obtain production certificates just to test a script.
- If the original certificate state is genuinely lost and **the operator has approved this recovery**, start the validated Caddy service with the approved domain/config on the correct endpoint; let configured Caddy automatic HTTPS request a new public certificate. Caddy may serve/start while issuance is still pending. **Do not treat `systemctl active` alone as TLS recovered.**
- Validate externally a trusted TLS chain and SAN for `api.meshalot.com`, the correct SNI and expiry, IPv4 and IPv6 where DNS advertises both, HTTPS `/v1/health` (or approved production health route), login/session authorization boundaries, and WebSocket upgrades/negative authorization. Verify persisted Caddy storage and renewal policy metadata with **non-secret evidence**, without reading or exporting private keys.
- If issuance fails, do **not** loop retries, relax TLS checking, turn off authentication, reissue across multiple hosts, disable firewall controls indiscriminately, or publish certificate material. Stop, preserve diagnostics without secrets, keep/revert to the prior working deployment if available, and request a new operator decision.

## Contingency: restore protected original TLS state (NOT YET READY)

This requires a *future approved* backup/restore scope. Identify actual Caddy storage first. Make a private encrypted, SHA256-verified, separately controlled storage backup, maintain least-privilege owner/mode and passphrase separation, copy to independently controlled media, and test its restoration **only inside an isolated authorized target**. Preserving ACME account/storage metadata avoids unnecessary new issuance; however the storage holds private keys and other sensitive assets, so never transfer it to public GitHub, GitHub Actions logs/artifacts, Google Drive, chat, or unprotected backup media. **Do not create this archive or handle the keys based on the current authorization.**

If no securely restorable TLS state exists, prefer approved clean reissuance over unverified copies; record any service interruption. A temporary locally trusted/self-signed certificate is acceptable only for internal synthetic tests, not as an untrusted public replacement.

## Security and evidence acceptance checklist

- **PASS (already):** official digest-pinned Caddy 2.11.4 synthetic HTTPS, WSS, authorization negatives, induced stage-only 502 and recovery, complete cleanup; one-shot runner workflow disabled/confirmed skipped.
- **PLAN COMPLETE (this document):** default reissue + contingency restore and required DNS/ACME/storage verification/stop controls.
- **OPEN:** verify production TLS storage/issuer/renewal and service dependencies by newly approved metadata-only method; confirm actual PG16 cluster global roles/ownership recovery policy; conduct separately authorized *full* isolated API/systemd/Caddy rollback rehearsal using production-derived archives only with fresh explicit archive/key scope approval; prove vault recovery practice separately.
- **NOT AUTHORIZED:** production certificate issuance/rotation, live Caddy restart, privileged reads, Caddy storage/key export, new IAM/SSH identities, public-network changes, database migration 9/10, node reboot, application cutover, cloud expense.
- **Evidence rules:** write only sanitized classifications (PASS/BLOCKED/NOT RUN) and public source references to GitHub/Drive. Keep exact private paths and raw metadata in previously owner-restricted MS-02 audit files only; never claim a backup or recovery test occurred when only planned.

**Gate status after decision: G17/1 = IN PROGRESS; G17/2–G17/7 = NOT STARTED; Milestone 13 = OPEN. Production unchanged.**

## Vendor references reviewed 2026-10-13

- [Caddy Automatic HTTPS](https://caddyserver.com/docs/automatic-https) — activation, challenges, staging tests, storage and background issuance.
- [Caddy Conventions](https://caddyserver.com/docs/conventions) — TLS data directory is persistent and sensitive; directory location depends on environment and storage.
- [Caddy Global Options](https://caddyserver.com/docs/caddyfile/options) — storage overrides and CA staging settings.
- [Caddy HTTPS Quick-start](https://caddyserver.com/docs/quick-starts/https) — DNS A/AAAA and public port requirements.
- [Caddy Troubleshooting](https://caddyserver.com/docs/troubleshooting) — bad/stale AAAA records and challenge accessibility.
- [Let's Encrypt Rate Limits](https://letsencrypt.org/docs/rate-limits/) — avoid unnecessary repeated production new-order/identical-identifier requests.
