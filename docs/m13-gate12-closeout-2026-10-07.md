# MeshAlot M13 — Gate 12 Closeout — 2026-10-07

**Gate 12 PASS — integrated agent + real WSS server + restricted-role PostgreSQL offline/reconnect recovery.**
**M13 remains IN PROGRESS; no production deployment or live Node001/Node002 interruption/reboot test.**

## Exact source and scope
- Gate 11 closeout baseline: `c29d6ebedc741cd73fb462e733aa94758c4e6129`.
- Integrated test initial source: `1bc8368b82be7f29435181f528a5f1b6bc4f656a`.
- MS-02 runner gofmt and fast-forward push: `0cbe54b7b6925baecd8a2057a39c0b02dfffabec`.
- Test source: `server/m13_vertical_slice_postgres_test.go`.
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469` unchanged.

## End-to-end proof
- Built a disposable network-disabled PostgreSQL 16 container with a Unix socket and all 10 migrations.
- Applied existing restricted grants for the non-root `meshalot` runtime DB account; admin role was used only to prepare isolated test users/identities and force a simulated stale timestamp.
- Created an isolated new fake node identity (not Node001 or Node002), inserted its public key through test fixtures, retained its private identity in a 0700/0600 temporary directory.
- Started the production-shaped `Service.NewWithPostgres` API handler in an isolated local TLS server and ran the real persistent agent WebSocket client using a pinned test certificate.
- Confirmed server verified the existing Ed25519 key, stored first operational heartbeat and latest telemetry with server timestamp, and showed the node online on both node detail and dashboard.
- Closed the first actual hijacked TLS socket to simulate transport loss, set the **isolated DB only** to a 121-second-old heartbeat, verified immediate offline display and the sweep persisting offline.
- Confirmed the real agent established a distinct second TLS socket, signed the fresh challenge with the same key, submitted a new heartbeat, and returned the original node to online on the dashboard.
- Rechecked the identity bytes before and after; unchanged. No enrollment call on reconnect. No remote shell/job feature.

## Verification evidence
- Workflow: https://github.com/wwjd4u/MeshAlot/actions/runs/37714020634
- Job: `113106250246`.
- `TestM13VerticalSliceRealPostgres`: PASS.
- All ten schema migrations and M13 runtime privilege tests: PASS.
- Full `go test ./...` and `go vet ./...`: PASS.
- Log markers: `M13_VERTICAL_TLS_AUTH_DB_OFFLINE_RECONNECT_DASHBOARD=PASS`, `M13_GATE12_END_TO_END=PASS`, `M13_GATE12_CODE=PASS`.
- The runner pushed only its formatting-only test source correction to M13 branch.

## Remaining acceptance blockers
- Production-shaped Caddy HTTPS reverse proxy / localhost forwarded-protocol path has not yet been exercised end-to-end; the Gate 12 TLS server used direct TLS to the handler. Test that path in Gate 13.
- Neither real physical Node001 nor Node002 was disconnected, slept, woken or rebooted. Live service packaging was not installed.
- Existing provider settings in service packaging are explicit/static until a persistent owner-control integration is verified. Do not enable real node service prematurely.
- M13 server/database changes remain on the isolated development branch.
