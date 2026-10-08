# MeshAlot M13 — Gate 2 Closeout (2026-10-07)

**Gate 2: PASS (isolated implementation and remote validation).**
**M13 milestone: IN PROGRESS — not deployed.**

## Verified source
- Repo: `wwjd4u/MeshAlot`
- Work branch: `m13-persistent-heartbeat`
- Baseline `main`: `1016c9b058690f34058dfab9cde7927e421fa469`
- Tested code commit: `e730d890efcf274743797fbf52572d48809f0e33`
- M13 Gate 1 contract/test commit: `b2ffe3a8927f79226a94765ecb270b98286d88c5`
- All implementation changes isolated to M13 branch. `main` unchanged.

## Gate 2 implementation
- Added route `GET /v1/agent/connect` in `server/service.go`.
- Added `server/m13_websocket.go` with Gorilla WebSocket v1.5.3.
- Refuses operation without PostgreSQL-backed enrolled-node identity.
- Requires direct TLS, or HTTPS reverse-proxy forwarding from verified loopback address (for the localhost Caddy architecture). Rejects plaintext remote WebSocket connections.
- Rejects browser `Origin` headers; this endpoint is for machine agents, not browsers.
- Sends a cryptographically random 32-byte, per-connection challenge, valid for a single authentication attempt with a 10-second timeout.
- Requires a domain-separated Ed25519 signature by the enrolled node; retrieves its existing stored public key. A signature for another challenge or node does not authenticate.
- Sends `authenticated` only after successful verification. The agent receives no jobs and cannot submit operational telemetry in this gate; application data is explicitly rejected until a future gate enables it.
- Adds restricted input sizes, bounded authentication time, idle deadline, ping/pong keepalive, and generic failures without exposing node/key details.
- Does **not** modify enrollment, stored identities, historical reports, production API, or existing legacy heartbeat.

## Remote validation proof
- Runner: `meshalot-ms02` on `wwjd4u-MS-02-Ultra`.
- Read-only validation workflow: `.github/workflows/ms02-m13-gate2.yml` on `infra/ms02-runner-gateway`.
- Final successful run: https://github.com/wwjd4u/MeshAlot/actions/runs/37706404774
- Job ID: `113081702252`.
- Workflow fetched **exact** commit `e730d890efcf274743797fbf52572d48809f0e33` to a disposable checkout.
- Logs confirm `GATE2_GO_FMT=PASS`, `go mod verify`, `go test ./server -run '^TestM13' -v -count=1`, `go test ./... -count=1`, `go vet ./...`, and `M13_GATE2_VALIDATION=PASS`.
- Dedicated Gate 2 tests passed: database requirement; authentication and rejection of pre-telemetry data; replayed signatures; unregistered nodes; plaintext transport and browser-origin rejection.
- Validation is isolated and does not prove production availability.

## Preserved boundaries
- No production deployment or infrastructure/network changes.
- Node002 enrollment and historical M8 reports untouched.
- V100 evidence and benchmark untouched; no rehash or benchmark rerun.
- No Tailscale.

## Next work
**M13 Gate 3:** Add bounded operational heartbeat telemetry on the authenticated connection and test that only the authenticated node can update its own operational status. Select explicit initial heartbeat thresholds in tests/design; do not reinterpret previously floated values as governing requirements. Subsequent gates must implement automatic stale/offline classification, reconnect/backoff, internet interruption and reboot recovery before M13 can be closed.
