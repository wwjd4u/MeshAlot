# MeshAlot M13 — Gate 17/1 Caddy synthetic HTTPS/WSS recovery evidence

Date: 2026-10-13 (US Central)

**Result: PASS for the operator-approved, disposable Caddy 2.11.4 synthetic TLS/WSS fault-and-recovery test.** This is one verified subcheck of **G17/1 of 7**, not closure of G17/1, Gate 17, or Milestone 13. This record contains only sanitized metadata and links.

## Evidence and source control

- [Successful self-hosted MS-02 GitHub Actions run 38449591265](https://github.com/wwjd4u/MeshAlot/actions/runs/38449591265), job 115476752911, concluded **success**. The Caddy workflow source for that run was [infra commit `a270a943e0d3f09c4bdb5e7a9c25c3f2326e6cd9`](https://github.com/wwjd4u/MeshAlot/commit/a270a943e0d3f09c4bdb5e7a9c25c3f2326e6cd9).
- [One-time workflow disabled](https://github.com/wwjd4u/MeshAlot/commit/7d55bb68cb29e5ff94bd76ad3038f467d88b13b0) immediately after PASS. Independent [subsequent workflow run 38449988211](https://github.com/wwjd4u/MeshAlot/actions/runs/38449988211) was **SKIPPED**; the trusted runner gateway check on that push passed.
- Image reference was digest-pinned as `docker.io/library/caddy@sha256:040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0`, platform `linux/amd64`. Pull and digest/platform verification **PASS**; runtime version **v2.11.4** verified; zero nonstandard Caddy modules reported.
- The official Caddy binary carried a file capability for binding privileged ports, which was incompatible with the deliberately capability-free runtime and caused an initial `exec /usr/bin/caddy: operation not permitted` safe-stop. The disposable derived stage image removed **only that unnecessary file capability**; no container security restriction or production file was relaxed. Subsequent version/module checks and all remaining stage tests **PASS**. Earlier failed attempts are not counted as PASS.
- The disposable synthetic helper compiled locally/offline. Staging used synthetic certificates, a synthetic backend and synthetic authorization tokens only.

## Isolated execution safeguards (verified PASS)

- Trusted runner: `meshalot-ms02`, host `wwjd4u-MS-02-Ultra`.
- Docker network mode `none`; **no host ports** and **no host bind mounts**.
- Container non-root user `10001:10001`, `--cap-drop=ALL`, `no-new-privileges`, and read-only root filesystem.
- Stage resource bounds: **2 CPUs**, **4 GiB RAM/swap ceiling**, **96-process limit**, and three disposable, non-executable tmpfs locations.
- No GCP IAP/SSH connection, live Caddy config, production API process, database, existing GPG archive, recovery passphrase, Node001/Node002 identity, historical benchmark or V100 evidence used by this test.
- Stage container, derived image, newly downloaded base image and temporary work directory disposal checks: **PASS**.

## Verified functional acceptance

- Known-good Caddyfile adapted/validated; deliberately invalid Caddyfile rejected.
- Trusted synthetic HTTPS `/healthz`: **PASS**; HTTPS proxy-forwarded-protocol normalization: **PASS**.
- Private endpoint without token: **401 BLOCKED**; with synthetic authorization: **PASS**.
- WebSocket without token: **BLOCKED**; disallowed browser-origin WebSocket request: **403 BLOCKED**.
- Untrusted synthetic TLS certificate: **BLOCKED**; trusted synthetic TLS: **PASS**.
- Real authenticated WebSocket upgrade and two masked text echo frames, including after an idle interval: **PASS**.
- Deliberate upstream failure introduced by stage-only Caddy config reload: expected **502**, **PASS**.
- Reload of known-good Caddy config: **PASS**, followed by a **second full HTTPS/WSS, authorization and TLS-negative acceptance PASS**.
- Runtime logs explicitly marked `M13_G17_CADDY_ONE_TIME_SYNTHETIC_ACCEPTANCE=PASS` and `M13_G17_CADDY_STAGE_DISPOSABLE_CLEANUP=PASS`.

## Limits — G17/1 still open

This does **not** establish the SHA-256 identity of the live production Caddy service binary, production certificate/ACME storage location, renewal and restoration method, existence of a complete independent TLS archive, or full production API/systemd/Caddy recovery. Nor does it rehearse restoring the encrypted production-derived rollback archives, production PostgreSQL global roles/credentials, or a live cutover.

**Next evidence gap:** review a secure production TLS certificate-storage/renewal *recovery strategy* without exporting live private keys; confirm required service runtime dependencies; then request separate operator approval for the **G17/2 full isolated rollback rehearsal** and any production-derived archive/passphrase use. Production deployments, migrations, service changes, WAN/reboot tests and new privileges remain **UNAUTHORIZED**.

**Gate statuses:** `G17/1 IN PROGRESS`; `G17/2 NOT STARTED`; `G17 OPEN`; `M13 OPEN`. No production change.
