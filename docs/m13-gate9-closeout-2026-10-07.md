# MeshAlot — M13 Gate 9 Closeout — 2026-10-07

**Gate 9 PASS — automatic secure WebSocket reconnection after interrupted transport.**
**M13 remains IN PROGRESS; no production deployment.**

## Verified source
- Gate 8 closeout baseline: `c9aac610091a4b12782393fc2339e95e4824a0bd`.
- Initial bounded-reconnect implementation: `130c473b73c49f251fec87f19b6c09495adea037`.
- Focused test cleanup commit: `0645e28818f0fad3c25fafd3ef3300fae9c29b7e`.
- Validated runner formatting and push commit: `ecda9f2a64924bf611ed4e314aaf70e5c8694fed`.
- Production `main` remains `1016c9b058690f34058dfab9cde7927e421fa469`.

## Behaviour and limits
- `agent/m13_reconnect.go` implements a context-cancellable, single-node outbound reconnect supervisor.
- Exponential backoff 1s, 2s, 4s, 8s, 16s, 32s, capped at 60s; adds cryptographic jitter of at most 500ms.
- Backoff resets after a connection survives at least two minutes.
- Handles unexpected session termination and transport failures without permanently abandoning the node or opening inbound ports.
- Repeats Ed25519 challenge authentication using the **existing** identity every time. Does not enroll, overwrite identities, accept arbitrary jobs, expose a remote shell, or change provider resource controls.
- Preserves TLS certificate checks and fails immediately for invalid existing identity, insecure control URL, disabled TLS validation, nil sampler, unsupported interval, or nil context.
- `agent/cmd/meshalot-agent/connect.go` now uses the automatic reconnect supervisor when invoked with explicit provider-mode and pause settings.
- The process must be kept running or restarted on machine boot in a later gate; on OS sleep, heartbeats pause and the server treats the node as offline after the M13 threshold. Reconnect on wake depends on the OS and network service resuming; hardware sleep/wake testing remains pending.

## Test evidence
- Runner: `meshalot-ms02` on MS-02 Ultra.
- [Successful GitHub Actions run](https://github.com/wwjd4u/MeshAlot/actions/runs/37713417794)
- Job: `113104337485`.
- Unit tests verify delay bounds, controlled retries, cancellation, backoff reset after a stable session, timer errors, and early rejection of unsafe configuration.
- Integration `TestM13PersistentSessionRecoversAfterTLSDrop` used an isolated local TLS WebSocket server which cut the first authenticated session and accepted a second authenticated session from the exact same Ed25519 identity. **PASS**.
- Existing Gate 7 CLI identity-preservation tests: **PASS**.
- Full `go test ./...`, focused `go test -race`, `go vet ./...`, Darwin x86_64 and arm64 cross-compilation: **PASS**.
- MS-02 runner committed and pushed formatting-only source updates to M13 branch; no production deploy.

## Next
**M13 Gate 10:** Create safe service-startup packaging for Node001 Linux (systemd), with explicitly supplied existing identity and provider settings. Validate the service template and failure-safe configuration without enabling or changing live host services. Then schedule controlled on-host installation, reboot, sleep/wake and internet-interruption proof before declaring M13 complete. Node002 live Mac validation is also pending. Preserve V100 evidence and historical M8 reports.
