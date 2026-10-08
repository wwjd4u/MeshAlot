# MeshAlot M13 — Gate 4 Closeout — 2026-10-07

**Gate 4: PASS — isolated outbound WSS agent session implementation and validation.**
**M13 milestone is still IN PROGRESS — do not deploy this branch to production.**

## Verified source and scope

- Branch before Gate 4: `e0bddcae6fd2507a56d489b9fe8b29187963babe`.
- Initial agent session: `2389aa5eaa29e9da6566d677a9078ea777e38f41`.
- TLS bypass refusal/cancellation fix: `fc835e9c693f0ad6bb8e0df4fd81a3c7e3b0dde3`.
- Negative test: `7aabc472e49ceaab5bc8544e2841fb87c1aedb6c`.
- Runner gofmt commit (validated final source): `9baede1b502231ee256ea35d63a4c38cd862545c`.
- Files changed: `agent/m13_connection.go` and `agent/m13_connection_test.go` only.
- Production `main` remained `1016c9b058690f34058dfab9cde7927e421fa469` during validation.

## Agent session capabilities

- Outbound `wss://` derived only from a valid HTTPS control-plane base URL.
- Rejects plaintext, embedded credentials, unwanted paths/query strings/fragments, and explicit TLS-verification bypass.
- Requires a valid pre-existing local identity. Does not generate, enroll or overwrite an identity.
- Receives the server nonce, signs its domain-separated challenge using the existing Ed25519 private key, and verifies the acknowledged node ID.
- Sends only validated heartbeat frames supplied by an injectable sampler; verifies acknowledgement before sending the next frame.
- Uses deadlines, frame-size limits, and context cancellation; caller handles reconnection in a later gate.
- No job execution, shell access, or direct inbound host listening introduced.

## Trusted MS-02 runner evidence

- Run: https://github.com/wwjd4u/MeshAlot/actions/runs/37711280463
- Job: `113097531319`.
- Exact pre-test input: `7aabc472e49ceaab5bc8544e2841fb87c1aedb6c`.
- `go mod verify`, focused agent M13 tests, full `go test ./...`, `go test -race ./agent -run '^TestM13'`, `go vet ./...`, and final gofmt checks all PASSED.
- Integration tests used a temporary TLS server and explicit test CA, checked challenge/signature, heartbeat delivery, and clean cancellation.
- Negative tests reject insecure or malformed control-plane URLs, mismatched identity, plaintext fallback, missing sampler, disabled TLS verification, and incorrect authenticated node.
- After tests, the MS-02 runner itself committed the formatting-only change and fast-forward pushed M13 branch: `9baede1b502231ee256ea35d63a4c38cd862545c`.
- No production deployment.

## Boundaries and next step

- Node002 original enrolled identity, historic M8 reports, V100 evidence, runtime configuration, and provider controls untouched.
- Gate 4 provides a **session client with injected sampler**, not a running background agent service.
- **Next M13 Gate 5**: implement lightweight host operational telemetry sampler (Linux and Darwin as feasible), maintain missing-GPU conventions, and verify real values safely without benchmarks. Subsequent gates wire the CLI/service supervisor, automatic stale/offline handling, backoff, interruption recovery and reboot persistence.
