# M13 G17 — synthetic systemd target prerequisite

**Synthetic systemd target prerequisite PASS. G17/1 BLOCKED.
G17/2 NOT STARTED. No production deployment or archived rollback acceptance.**

The operator authorized nonproduction configuration/testing without repeated
approval and requested autonomous work while traveling. The October 29 runner
receipt found no QEMU executable on PATH; it did not establish a systemd guest
or safe full rehearsal target. This new bounded test addresses that specific
missing prerequisite, using public synthetic material only.

The [implementation and acceptance contract](../tools/recovery/systemd_stage/README.md)
define the required outcome, resources, security/failure/integrity checks and
stopping point before execution. It builds QEMU in a disposable container rather
than installing on MS-02 or using its KVM device. The runtime is nonprivileged,
has no host mounts or external network, and boots a real Ubuntu systemd guest.
A synthetic loopback service tests unit precedence, restricted execution,
authorization denial, restart persistence, deliberate fault/recovery and shutdown.

Static Python/shell syntax and whitespace checks passed. An unstarted local Docker
metadata fixture confirmed how tmpfs and no-new-privileges are represented by
the daemon; it was removed without execution. The real VM acceptance checks were
not run at the initial source checkpoint; the actual receipt below supersedes
that pending status. No previously passing suite was repeated.

The first actual [run 41113449849](https://github.com/wwjd4u/MeshAlot/actions/runs/41113449849)
at infrastructure SHA `2e85bd19e27743584ab56869a9a51a7a6614793b`
failed before VM startup during package assembly. Trusted allocation preflight
and the 621284352-byte cloud-image digest passed; task-container/image and
source-workspace cleanup passed. Source was
`317489fd47e522f92c0c26b2951319e92934875e`. Only a generic subprocess failure was
reported, so the Docker wrapper now emits fixed operation/reason classifications
instead of raw diagnostics. One targeted follow-up is justified by this failed
assembly result, not by uncertainty about previously accepted suites. Canonical
cloud bytes and the OCI index/amd64 digests remain unchanged and were confirmed
at the public registry; no pin was loosened or credential changed.

## Failure evidence retained and bounded correction

| Execution | Actual result |
| --- | --- |
| [41113449849](https://github.com/wwjd4u/MeshAlot/actions/runs/41113449849), infrastructure `2e85bd19e27743584ab56869a9a51a7a6614793b` | FAIL during assembly; generic subprocess failure; no guest startup; cleanup PASS. |
| [41113832867](https://github.com/wwjd4u/MeshAlot/actions/runs/41113832867), infrastructure `911483631350b383337ad69c08d8ff8064a3ab7d` | FAIL at Docker start; operation classification added; no guest startup; cleanup PASS. |
| [41114138050](https://github.com/wwjd4u/MeshAlot/actions/runs/41114138050), infrastructure `f10236598a0269036aec701eb3247f700ac7a2ff` | FAIL at Docker start, exit 1, no assembly-script phase; cleanup PASS. |
| [41114631815](https://github.com/wwjd4u/MeshAlot/actions/runs/41114631815), infrastructure `fbaf937d20b92b8b19aba1c18d405ad828d8117d` | Empty-container startup diagnostic only, no cloud download or VM test. Default runtime runc; logging-driver initialization failed; container never started, exit state 128; cleanup PASS. |

The narrowly scoped startup diagnostic emitted only a fixed public vocabulary,
not raw Docker errors, paths, URLs or identifiers. The correction explicitly
sets `compress=false` on each task container's local logger with a single
1-MiB log file. Corrected local empty-container startup and cleanup passed;
the runner's complete corrected execution follows. No host logging default,
daemon, permission, security flag, image pin or credential was changed.
The earlier failed attempts are not PASS evidence. Each continuation addressed
the failed new assembly path, not previously accepted application/recovery tests.

## Actual MS-02 receipt

- [Successful run 41114765692](https://github.com/wwjd4u/MeshAlot/actions/runs/41114765692),
  job 124265983641, push event; provider timestamps
  **2026-10-30 19:22:45Z–19:27:05Z** (14:22:45–14:27:05 America/Chicago).
- Exact infrastructure SHA: `7364cd3c856e41eab1e2c74d8e01a223b301897f`.
- Exact immutable tooling source: `905abd6b893d6d194a44dc9138b6f4215dc3dcab`;
  five independently SHA256-pinned public source files verified before execution.
- Canonical cloud disk: **621284352 bytes**, SHA256
  `04e8f2461a16a44e78a370284471f29616cdd9baf1893df25807b71f985b6c94`, PASS.
- Disposable tooling-image receipt:
  `sha256:e54acf86d6d8a92ec867cb59931f39d020c30b8bb719d59aa8db2a08d0034c48`.
  This image was removed after use; the digest is evidence, not a retained stage.
- Installed package receipts: QEMU system/utils `1:8.2.2+ds-0ubuntu1.18`,
  Python `3.12.3-0ubuntu2.1`, genisoimage `9:1.1.11-3.5`, seabios `1.16.3-2`.
- Trusted identity/allocation preflight, public package assembly, exact container
  boundaries before start and after exit: **PASS**. No OOM or incomplete shutdown.
- Guest boot, all checks and shutdown: **229.208 seconds** within the 360-second
  budget. The service checks took **8.196 seconds**. These are prerequisite
  execution times, not product latency/load results.

All **16 guest checks PASS**: PID 1 systemd; only loopback interface; no external
IPv4 route; no external IPv6 route; static unit verification; restricted-UID
service start; drop-in precedence; no-new-privileges; no effective capabilities;
negative synthetic authorization; service stop; state preserved over restart;
deliberate missing-executable fault detected; configuration restored; original
synthetic fixture intact; final service stop. Actual VM poweroff PASS.

Task containers, unique tooling-image tag and source workspace cleanup PASS.
The successful listing checks verified absence; an arbitrary inspect error was
not treated as proof of cleanup. A public base-image cache may remain; there was
no global prune or deletion of existing resources.

Only actual emitted result lines were used, excluding echoed workflow source.
The report explicitly states `PRODUCTION_CONNECTION=NONE`,
`REAL_ARCHIVES_OR_KEYS=NOT_USED`, `FULL_ARCHIVED_ROLLBACK=NOT_TESTED` and
`GATE_ACCEPTANCE=NOT_ASSESSED`. All five tested source files remain byte-identical
when this documentation receipt is published. No additional runtime repeat is
needed for this evidence update.

## Gate decision and next supported action

The synthetic systemd-capable target prerequisite is now **PASS**, with the
resource/network/security constraints above. The VM was destroyed. Future
full-service work must recreate and inspect its own target under a reviewed
resource budget; 1 GiB guest RAM is not proof of full PG/API/Caddy capacity.

**G17/1 remains BLOCKED; G17/2–7 remain NOT STARTED.** Actual vault-sourced
recovery-key access, protected real PostgreSQL globals/ownership recovery,
remaining production TLS recovery prerequisites and full archived service
replay remain open. This synthetic service is not the MeshAlot API and supplies
no new customer-workflow, database, Caddy or archive-restore acceptance.

Reuse this target design for the later full-service rehearsal after its entry
requirements are met. Continue to defer private Mac/vault work until the operator
is parked. The specifically rejected PostgreSQL diagnostic remains paused;
production push/merge/cutover, SSH renewal and Node001/Node002 identity/physical
operations remain outside this execution. Main and existing backups are intact.
