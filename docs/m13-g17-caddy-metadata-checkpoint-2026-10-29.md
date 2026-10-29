# M13 G17/1 — consolidated Caddy metadata checkpoint

Recorded: October 29, 2026 (America/Chicago). Observations were supplied by the
operator in this conversation on October 27–28. These are operator-reported
console results, not a new automated inspection, signed receipt or CI run.

**Metadata inspection: PARTIAL PASS. G17/1: BLOCKED. G17/2–7: NOT STARTED.
M13: OPEN / NOT RELEASE-READY.**

## Scope and source

The operator used their normal administrator console on meshalot-control-01.
The commands read allowlisted service, binary, listener, directory and local
Caddy admin-API metadata. Selected filesystem/process metadata reads used
existing noninteractive administrative privileges. This was an independent
Caddy inspection; it did not reproduce the paused PostgreSQL diagnostic.

The observations below are point-in-time evidence. They do not establish that
the service has remained unchanged since the inspection. Exact storage paths,
raw logs, environment values, ACME account data and private keys are excluded.

## Observed results and limits

| Item | Operator-reported result | Supported conclusion |
| --- | --- | --- |
| Host and service | Host meshalot-control-01; loaded, active, running; User=caddy, Group=caddy, DynamicUser=no; initial MainPID=556 | PASS for the reported service metadata |
| systemd state directory | StateDirectory empty; StateDirectoryMode=0755 | No explicit StateDirectory declaration; the mode alone says nothing about the actual TLS data directory |
| Global storage API | GET /config/storage returned JSON null | No explicit global storage configuration was observed |
| AppDataDir discovery | Current-PID startup entry found; journal-access warning False | The process-reported application-data path was identified locally; its value was not published |
| Data-directory metadata | Directory, mode 0700, owner caddy, group caddy | PASS for existence/type/owner/mode at the identified path; not proof of persistence, effective policy-specific storage, backups or successful writes |
| Running executable | SHA256 b7105518e3ed1c0761f232e44fc09345535533c9cb0abf0e12809416c7ac64d9; installed CLI digest matched; CLI v2.11.4 | PASS for the reported running-binary/CLI match, guarded by stable MainPID and InvocationID |
| Admin listener | Process match True, loopback-only True, IPv4 endpoint present True | PASS for attributing the port-2019 listener to Caddy at that inspection |
| TLS automation descendants | Three HTTP 400 responses, each at config/apps/tls/automation; service identity stable True | The requested descendants were not traversable in the exported JSON; this is not evidence that runtime automation is disabled |
| HTTP server srv0 | listen=[":443"]; automatic_https=null; first route/first matcher host=["api.meshalot.com"]; all HTTP 200; service identity stable True | PASS for these specific exported HTTP-server fields; supports default automatic-HTTPS configuration for the inspected route |
| Targeted automation journal query | Exit 124, zero parsed entries, target-domain mention False, parser exit 0, invocation stable True | BLOCKED by timeout. Empty input does not establish absence of an automation event; do not reuse False as a TLS failure finding |

Earlier generic UNVERIFIED outputs and unprivileged directory-stat failure did
not establish a missing directory or disabled TLS. The later successful
administrative stat provides the limited directory evidence above. No passing
application, migration, ACL, synthetic Caddy, V100 or node test was repeated.

## Interpretation against Caddy v2.11.4 source

The admin API reads exported configuration JSON, not a complete serialization
of every provisioned runtime default. Caddy's HTTP app provisions the TLS app;
the TLS app creates a default public automation policy when no explicit
automation object is supplied. Qualifying HTTP hostnames can enable automatic
HTTPS without an explicit TLS automation subtree.

References reviewed for these semantics:

- [Admin configuration traversal](https://github.com/caddyserver/caddy/blob/v2.11.4/admin.go#L1153).
- [HTTP app provisions the TLS app](https://github.com/caddyserver/caddy/blob/v2.11.4/modules/caddyhttp/app.go#L183).
- [Default TLS automation policy](https://github.com/caddyserver/caddy/blob/v2.11.4/modules/caddytls/tls.go#L289).
- [Automatic HTTPS defaults and exclusions](https://github.com/caddyserver/caddy/blob/v2.11.4/modules/caddyhttp/autohttps.go#L33).

Accordingly, the observed fields support default automatic-HTTPS configuration.
They do not prove a completed renewal, current certificate issuer/account,
successful future ACME validation, all-server/all-route coverage, storage
durability or restoration of TLS state. No production handshake, issuance,
rotation or renewal was performed by this consolidation.

## Remaining acceptance and next work

| Requirement | Disposition |
| --- | --- |
| Real PG16 roles, memberships, owners, grants/authentication and recoverable globals asset | BLOCKED / UNVERIFIED; the rejected diagnostic remains paused |
| Primary TLS recovery strategy | Previously documented clean reissuance at an approved recovery/cutover; these metadata reads do not execute that strategy |
| Actual issuer/account and renewal outcome; recovery-target storage, DNS and CA/challenge reachability | UNVERIFIED / NOT TESTED; retain as target-specific recovery requirements |
| Protected TLS-storage fallback backup and isolated restore | NOT VERIFIED; do not infer inclusion in the nine-file archive or environment supplement |
| Independent use of vault-sourced recovery keys | NOT TESTED; proceed with the separately described private key check when the existing encrypted backup volume is available |
| Full archived API/systemd/Caddy fault-and-restore rehearsal | NOT STARTED; G17/2 entry requirements remain in force |

The key check is defined in [the bounded vault recovery-key procedure](m13-g17-vault-recovery-key-check-2026-10-29.md).
It is a distinct missing recovery proof, not a repeat of the accepted logical
database restore. Do not continue the timed-out journal scan or probe individual
configuration leaves without a specific new acceptance need.

## Current operating authorization

On October 29 the operator authorized routine testing, configuration and
continued preparation without repeated permission requests, reserving production
pushes and material-risk decisions. Preserve the specific rejected-operation
safe stop, original evidence/backups and Node001/Node002 identities. A missing
device, artifact or tool is an information dependency, not a reason to request
the same general continuation approval again. No gate criterion is waived.

This checkpoint changes documentation only. No production connection, database
operation, archive decryption, SSH renewal, workflow dispatch, service change or
new runtime test was performed while preparing it.
