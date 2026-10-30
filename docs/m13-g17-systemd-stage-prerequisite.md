# M13 G17 — synthetic systemd target prerequisite

**Implementation prepared; actual MS-02 execution pending. G17/1 BLOCKED.
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

Static Python/shell syntax and whitespace checks pass. An unstarted local Docker
metadata fixture confirmed how tmpfs and no-new-privileges are represented by
the daemon; it was removed without execution. The real VM acceptance checks are
**NOT RUN** at this source checkpoint. No previously passing suite was repeated.

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

The later actual runner receipt will be recorded here. Even a passing synthetic
target test cannot certify the real recovery dependencies or start G17/2 out of
order. Actual archive/key use, real globals/ownership recovery, TLS prerequisites,
full archived service replay and production cutover remain separately held.
