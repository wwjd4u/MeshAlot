# MeshAlot M13 — Gate 10 Closeout — 2026-10-07

**Gate 10 PACKAGING PASS (dry run only).**
**No live systemd service installation, enabling, host reboot, or production deployment. M13 is not closed.**

## Scope and tested source
- Gate 9 closeout baseline: `4ae9c0a9e9da80e19dfca8e4b6d425d4bb03119b`.
- Service packaging implementation: `59e6f6eb4bd9515082e0e164251265214eec5b02`.
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469`, unchanged.

## Files created
- `deployments/systemd/meshalot-agent@.service`: templated per-existing-user systemd service, with outbound `connect` invocation, restart-on-failure, and conservative runtime privileges.
- `deployments/systemd/agent.env.example`: intentionally blank, explicit control-plane URL, existing identity path, provider availability mode and manual-pause state. Empty values fail closed.
- `deployments/systemd/README.md`: safety contract, approved-install requirements, and future acceptance tests.

## Security boundaries
- Run as the Linux account owning the **existing enrolled Ed25519 identity**; do not enroll again, regenerate private keys, change file permissions or move identity files.
- Service command uses `meshalot-agent connect`, whose Gate 9 reconnect supervisor reuses the identity and requires an explicit current sharing mode/pause decision.
- No inbound listening, arbitrary job operation, remote shell or sudo installation was added.
- Until explicit provider-control persistence/synchronization and the controlled live acceptance test, service package must **not** be enabled unattended.
- No production M13 server/database deployment or Node002 change.

## Verified on MS-02 runner
- Successful workflow: https://github.com/wwjd4u/MeshAlot/actions/runs/37713629486
- Job ID `113105003876`.
- Workflow checked out pinned source, built the Go agent in a temporary runner directory, and syntax-verified a copied, path-substituted systemd unit using `systemd-analyze verify`.
- A dry-run connect invocation with a deliberately missing identity failed as required and did not create an identity file.
- Gate 7 connect security regression tests passed.
- Markers: `M13_GATE10_SYSTEMD_VERIFY=PASS`, `M13_GATE10_NO_ENROLLMENT=PASS`, `M13_GATE10_PACKAGING=PASS`.
- No systemctl operations, live service creation, service restart or agent enablement.

## Next
Gate 11 can prepare macOS per-user startup packaging for Node002 without installation or re-enrollment. After package validation, a later controlled live test must verify Node001/Node002 connect over secure channel, correct provider controls, offline transition on disconnection, automatic reconnection after internet/sleep interruption, and node recovery after reboot. Those acceptance tests remain pending.
