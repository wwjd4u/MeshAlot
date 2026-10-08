# M13 Linux provider agent service (NOT installed)

These are **packaging templates only**, not proof of service installation or
reboot recovery. The MS-02 runner may build and run tests against a temporary
copy; it must not install, enable, start, or restart this unit.

## Safety contract

- Use the Linux account that **already owns** the M6 enrolled Ed25519
  identity, as the systemd instance name meshalot-agent@<username>.service.
  Do not enroll again, change file permissions or move the private identity.
- Install the compiled agent under /usr/local/bin/meshalot-agent only with
  explicit operator approval and version/commit verification.
- Place a root-owned 0600 environment file at
  /etc/meshalot/agent-<username>.env only after approval.
- Provide all required variables explicitly:
  MESHALOT_SERVER_URL (HTTPS), MESHALOT_IDENTITY_PATH (existing JSON),
  MESHALOT_PROVIDER_MODE (normal/away/maximum-earnings), and
  MESHALOT_MANUAL_PAUSE (true/false). Blank values cause the agent to
  fail closed. Unknown job state is used until actual local job integration.
- Service is outbound-only on TLS and uses the automatic reconnect supervisor
  from Gate 9. It neither exposes TCP ports nor permits remote shell/jobs.
- Provider controls entered in the static environment file are **not** yet
  dynamically synchronized with the local provider-control store. Until that
  integration is verified, this service must not be enabled unattended.
- Never install this draft against production before the M13 server, migration,
  and restart/sleep/interruption acceptance gates have passed.

## Planned acceptance checks — not yet performed

With separate permission and the correct node owner, a controlled Node001
installation must verify: existing key fingerprint unchanged, actual NVIDIA
GPU/RAM/CPU telemetry visible, safe manual-pause behaviour, loss and return of
internet connection, server offline threshold, restart-on-boot, and successful
reconnection without re-enrollment. A comparable launchd path and a real
sleep/wake test are needed for Node002.

This repo does NOT run systemctl enable, systemctl start, sudo service
installation, production DB migration, or host reboot.
