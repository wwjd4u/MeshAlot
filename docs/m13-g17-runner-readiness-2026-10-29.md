# M13 G17 — automated MS-02 runner readiness checkpoint

Date: October 29, 2026 (America/Chicago).

**RUNNER EXECUTION / READ-ONLY METADATA: PASS.
FULL STAGE SUITABILITY: NOT TESTED. G17/1: BLOCKED.**

## Objective, authorization and scope

The operator authorized continued routine testing/configuration without repeated
permission prompts and asked that the existing GitHub runner handle automated
work. Its last recorded successful execution was October 13. Current runner and
staging capabilities were a specific unresolved dependency, so this single
read-only job was warranted; it did not repeat accepted application, ACL,
migration, Caddy, recovery or benchmark suites.

The integration's read of the runner-administration endpoint returned HTTP 403,
Resource not accessible by integration. This is an API permission limit, not
another reported safety rejection and not proof the runner is offline. No
permission setting or credential was changed. Execution was verified through
the existing trusted gateway's authorized workflow instead.

## Reviewed configuration and execution

- Existing infrastructure branch: infra/ms02-runner-gateway.
- Previous infrastructure SHA: 54bb072463f257c1e1b75484244dfa82ee894b50.
- [Reviewed configuration commit](https://github.com/wwjd4u/MeshAlot/commit/81e78c042e3dcc4df75a1a8c2328e11a28f92653): only the existing gateway workflow changed. No new branch or checkout was created; the M13 checkout/index were preserved.
- [Actual execution 40948897708](https://github.com/wwjd4u/MeshAlot/actions/runs/40948897708): completed SUCCESS at that exact source SHA. Job 123711417383 completed its metadata step successfully.
- Workflow still has read-only contents permission and a five-minute job cap.
  Exact repository, infrastructure ref, runner, OS/architecture, host, UID name
  and home-directory expectations guard execution. Private values are not logged.
- Push triggering is now restricted to changes in this gateway workflow;
  unrelated infrastructure pushes no longer repeat it. Manual triggering remains
  available. No old disabled workflow, attempt marker or rejected diagnostic was
  re-enabled, reset or reused.

Local checks passed before publication: YAML trigger/permission/runner/timeout
contract, embedded Python compilation, a single allowed subprocess targeting
only the local Docker Unix socket, and the exact one-file infrastructure diff.
Existing workflow filters were reviewed: only the trusted gateway matched the
changed path. No checkout action or remotely fetched executable was used.

## Actual sanitized result

| Observation | Result |
| --- | --- |
| Trusted runner identity | PASS |
| Logical CPUs | 14 |
| Available RAM, GiB floor | 113 |
| Available home-filesystem disk, GiB floor | 625 |
| Host PID 1 is systemd | True |
| QEMU executable present | False |
| KVM character-device read/write access reported | True |
| Local Docker Unix daemon | ACCESSIBLE |
| git / go / python3 executables present | True |
| Capacity for planning, minimum 4 CPUs / 8 GiB RAM / 15 GiB disk | SUFFICIENT |
| Full isolated service-stage suitability and isolation | NOT TESTED |
| Gate acceptance | NOT ASSESSED |

These are the emitted result lines, excluding the source-code echo in Actions
logs. For example, a BLOCKED literal appearing in echoed Python is not a failed
runtime result. Identity PASS and sufficient planning capacity match the actual
job exit status.

The only subprocess queried Docker's version through the explicit local Unix
socket, with an eight-second timeout and captured diagnostics. Context/host
overrides would have prevented that query. No container/image/VM creation, pull,
production connection, archive/key access, host configuration change, SSH
renewal, application test or node operation occurred.

## Supported next work and limits

Use the runner for suitable approved builds, synthetic tests and configuration
checks. Reuse earlier PASS evidence while its inputs and conditions remain
valid. A current execution result does not authorize broad secret access or
prove a future isolated VM's network boundaries.

QEMU was not found on the runner's PATH, so the proposed VM-based full systemd
rehearsal target is not ready by presence alone. KVM access and host systemd do not prove a safe
isolated guest. Select and verify the complete target/resource/isolation plan
before any full archived service rehearsal. Plan any tool installation under the
approved configuration scope, checking its effect on shared host services and
privileges first. Do not claim that the old PG16-only container satisfied it.

The operator reports the encrypted Toshiba drive connected to the Mac. The
[vault recovery-key check](m13-g17-vault-recovery-key-check-2026-10-29.md) still
requires Mac mount/GnuPG metadata and private human key entry. Do not move keys
into GitHub, the runner or shared logs to automate that human boundary.

Real PG16 globals/ownership recovery, remaining TLS recovery requirements and
the full archived service rehearsal remain open. Production main, existing
backups and Node001/Node002 identities were not changed by this job.
