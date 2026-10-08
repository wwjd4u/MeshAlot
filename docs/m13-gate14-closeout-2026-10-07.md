# MeshAlot M13 — Gate 14 Read-only Preflight — 2026-10-07

**Gate 14 PASS — read-only host identity prerequisite and production API health preflight.**
**M13 not deployed, merged, or enabled.**

## Verified state
- Governing M13 branch before and after: `6af3eed39ae77a4f639a11636159435b7a94f8a1`.
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469` unchanged.
- MS-02 runner name: `meshalot-ms02`; host corresponds to Node001 test machine.
- Existing read-only Node001 identity: `/home/jason_guynes/.config/meshalot/agent-identity.json`. Agent `identity` check passed strict identity validation and did not create/re-enroll it.
- Node-ID fingerprint (short SHA-256): `2d717e508988`.
- Public-key fingerprint: `1249f59eacd83ffc`. Private key bytes were never logged.
- No `meshalot-agent` executable in the runner login account PATH.
- API GET `https://api.meshalot.com/v1/health`: response reported status `ok`; no M13 deployment made.

## Remote evidence
- [Successful Actions run](https://github.com/wwjd4u/MeshAlot/actions/runs/37714365158).
- Job `113107349647`.
- Runner built a temporary M13 binary to check identity syntax and deleted it on exit; no binary/service installed.
- Queried systemd service list read-only, no install/start/stop commands.
- Markers: `M13_GATE14_GO_BUILD=PASS`, `M13_GATE14_IDENTITY=EXISTING_VALID_READ_ONLY`,
  `M13_GATE14_INSTALLED_AGENT=NOT_IN_PATH`, `M13_GATE14_PRODUCTION_HEALTH=PASS`,
  `M13_GATE14_READ_ONLY_PREFLIGHT=PASS`.

## Safety and next
- Verified the original M6 private identity is available to the runner at its default path; it must not be regenerated, copied off-host, or overwritten.
- No Node002 enrollment/history touched; no V100 benchmark rerun or evidence rehash.
- Gate 15 may safely run the real Node001 agent with its **existing** identity against a temporary localhost TLS+disposable PostgreSQL test endpoint, without deploying server changes or enrolling the real node in a new account. Use no network-facing listener and do not print private identity material.
- The controlled production/server rollout, persistent systemd installation, actual WAN interruption, host sleep/wake, reboot recovery, and real Mac Node002 acceptance remain pending. These require separate explicit approval and a rollback plan.
