# MeshAlot M13 — Gate 17 Controlled Live Rollout Plan (NOT EXECUTED)
Date: 2026-10-07

**Status: READY FOR OPERATOR REVIEW — BLOCKED ON EXPLICIT AUTHORIZATION.**
M13 Gates 1–16 have isolated code, database, reverse-proxy simulation,
read-only Node001 identity and local-host smoke evidence. They do NOT prove
live production availability or real reboot/sleep/WAN recovery.

## Evidence prerequisites
- Governing milestone: AI Mesh POC Build, Test, and Release Guide, Milestone 13.
- Work branch: `m13-persistent-heartbeat`; keep `main` and live release unchanged
  until a controlled acceptance decision.
- `docs/m13-gate2-closeout-2026-10-07.md` through
  `docs/m13-gate16-closeout-2026-10-07.md` preserve detailed proof.
- Node001 existing identity file was read-only validated at
  `/home/jason_guynes/.config/meshalot/agent-identity.json` (Gate 14);
  Gate 15 proved a real one-session local TLS/PG signed heartbeat and NVIDIA
  GPU/VRAM telemetry without modifying identity or enrolled state.
- Node002 Mac enrollment and original M8 report MUST NOT be replaced.
- V100 tests and hashed evidence are closed; do not rerun by default.

## Operator decision points (none preapproved by this plan)
- Explicit permission to inspect production control-server state and backups
  beyond public health endpoints.
- Explicit permission and chosen target for a production-shaped staging
  deployment; no extra Google Cloud cost or DNS/network changes without approval.
- Separate authorization for production schema migration/runtime grants,
  binary/API or Caddy updates, and Node001 agent service installation.
- Explicit timing/approval for *disruptive* physical WAN interruption, host
  sleep/wake, Mac login/reboot or Node001 reboot. Do not surprise an active user.

## Controlled rollout sequence — stop on any failed check
1. Re-verify exact branch SHAs and Go test/proxy/migration evidence, confirm
   no unexpected commit or source changes. Pin immutable candidate release SHA.
2. Read-only inspect production PostgreSQL `meshalot_meta.schema_migrations`,
   current running binary/source SHA, systemd service/drop-ins, environment
   ownership and Caddy route policy. Do not dump credentials or secret fields.
3. Capture an approved private production PostgreSQL backup and verified
   restore check (outside the public GitHub repo), plus known-good running API
   binary/service configuration and reverse-proxy rollback material. Keep
   prior release operational; record artifact IDs in a PRIVATE evidence ledger.
4. Prefer an isolated staging/canary route and database before live production
   cutover. Use outbound `wss://` and trusted loopback TLS forwarding.
   Validate host settings, Caddy upstream WebSocket upgrades, public TLS,
   IPv4 and IPv6 and lack of inbound provider ports.
5. After a separate explicit release approval, apply additive tracked
   `000010_m13_heartbeat.up.sql` with `database/migrate.py apply` as
   migration owner, then apply restricted runtime grants and check that all
   10 schema versions and recorded checksums verify. Never force an altered
   history entry or apply to an untracked production schema.
6. Stage and start the exact candidate API binary in the approved environment
   with preserved private configuration. Verify existing HTTPS health, login,
   dashboard, node/report history, and new WSS challenge authentication.
7. Run the Node001 agent **manually first**, as the existing identity owner,
   using its original read-only identity and explicit **paused** provider
   mode; check real GPU/RAM/CPU telemetry, correct RTT, no remote jobs,
   no re-enrollment, unchanged private fingerprint and Node001 original
   network/compute benchmark records.
8. With operator approval, interrupt Node001's network; after at least 90s
   of absent M13 heartbeats verify offline status. Restore network, verify
   exponential-backoff WSS reconnection with the same node ID and online
   status, including newly measured telemetry and no access to the host shell.
9. Only after the manual test, verify the local provider-resource-control
   source is authoritative and any paused/scheduled settings will be honored
   after restart. The current staged service templates require explicit
   static values; they are NOT proof of automatic provider-control sync.
   Do not enable unattended service until that integration is accepted.
10. With separate approval, install/enable only the new Node001 agent
    systemd service. Confirm reboot recovery, never touching the existing
    Ollama/Hermes configuration or V100 evidence.
11. Test Node002 on the real Mac in a separate authorized window: existing
    identity intact, launch-at-login packaging installed only on approval,
    sleep/wake and disconnect/reconnect proved, required operational metrics
    truthful (GPU unknown fields remain absent when unsupported).
12. Observe production stability, dashboard/authorization boundaries,
    DB and cloud costs, TLS and reconnect logs. Close M13 only if all
    governing acceptance criteria pass. Otherwise leave M13 in POC and
    fix only the failing gate.

## Rollback and stop criteria
- If PostgreSQL migration/checksum status deviates, STOP, inspect, do not
  rewrite history or run a blind migration/restore.
- If the new API fails, revert the API binary/systemd config and any Caddy
  changes to the known-good release; additive nullable `m13_telemetry` may
  remain for compatibility. Do not blindly drop a column or overwrite records.
- If the new agent behaves incorrectly, stop/disable only its new service;
  keep original enrolled Node001/Node002 identities and node history.
- If provider settings are unknown or cannot be enforced on restart,
  remain manually paused; do not schedule third-party work.
- If a host heats up, loses GPU access or impacts the owner's workloads,
  stop the new test, preserve logs and restore prior host state.
- Never use Tailscale, host port forwarding, arbitrary command execution,
  real-money settlement, or unapproved unrelated service changes.

## Next operator input
Confirm the target environment, whether a staging/canary deploy is approved,
and when it is safe to run physical Node001 and Node002 tests. Nothing here
authorizes execution of a production migration or service deployment.
