# Synthetic systemd VM prerequisite

This prepares the isolated target needed by Gate 17. It is **not** a MeshAlot
application test, recovery-archive executor or G17/2 rollback rehearsal. Actual
PG globals, production TLS recovery, the Mac/vault drill and archive approval
remain separate. The rejected PostgreSQL diagnostic is not implemented here.

The operator authorized routine nonproduction testing/configuration and asked
for work without terminal copy/paste while traveling. `run_stage.py` accepts no
arguments and refuses execution outside the exact existing MS-02 runner,
repository, infrastructure ref, host and user. The reviewed workflow fetches
these public files at an immutable M13 SHA with independent file-digest pins.
It never checks out private host files, imports secrets or mounts the workspace.

## Required outcome and stopping point

Boot an actual systemd guest with no external interface or route; prove a
restricted synthetic service's configuration precedence, authorization denial,
start/stop, persistent fixture across restart, deliberate missing-executable
failure and configuration recovery. Then power off and remove only this
attempt's container/image/workspace. Stop on failure; no automatic retry.
Success establishes a **synthetic systemd target prerequisite only**. Keep
G17/1 blocked and G17/2 not started until their real prerequisites are met.

## Resources and boundaries

- Existing MS-02; no new cloud resource, cost, host package, service, SSH/IAM,
  KVM access or device-permission change. Software emulation uses QEMU TCG.
- Allocation preflight: at least 4 logical CPUs, 8 GiB currently available RAM
  and 12 GiB free temporary filesystem space. This is warranted by this new
  allocation, not a reason to repeat the old general readiness job.
- Named package-assembly container: 2 CPU quota, 3 GiB RAM/swap total, 128 PIDs,
  no privilege/device/host mount/port, no-new-privileges; standard Docker bridge
  only for public Ubuntu packages. It is stopped/removed on timeout or failure.
  No production connection or operator credential is supplied.
- Fixed Ubuntu 24.04 OCI digest and dated Canonical cloud image, SHA256 pinned
  from Canonical's HTTPS checksum listing. This is **not** a claim of detached
  image-signature verification. APT verifies repository signatures; actual
  installed package versions are reported. Package versions are not locked,
  so a later rebuild must record its new tooling image/versions.
- QEMU runtime: unprivileged UID/GID 65534, all capabilities dropped,
  no-new-privileges, read-only container root, no Docker socket/host mounts or
  extra devices, no port publication, Docker network `none`, and QEMU `-nic none`.
  Uses 2 CPU quota, 3 GiB total RAM/swap cap, 64 PIDs, 16 MiB shared memory.
- Guest: 2 virtual CPUs, 1024 MiB RAM, base disk at most 8 GiB virtual, no external
  backing chain. All writes go to a disposable qcow2 overlay in a 768 MiB
  size-capped tmpfs (`noexec,nosuid,nodev`). No physical storage attachment.
- Deadlines: cloud download 240 seconds checked between reads (20-second socket
  timeout); package assembly 600 seconds; VM boot/test/shutdown 360 seconds;
  outer attach 420 seconds; workflow 35 minutes. Startup and fault checks inside
  the guest have their own shorter bounds. A command timeout fails, then cleanup
  runs. SIGTERM also enters cleanup; hard host/runner loss cannot guarantee it.
- Only four packaging/probe files plus the pinned public cloud disk
  is copied to the assembly container. A unique per-attempt image is committed
  locally, used, then removed; no image registry publication or artifact upload.
- Container configuration is inspected before start and after exit. Require
  exact limits, isolation and exit status; OOM or incomplete poweroff fails.
  Guest serial output is memory-capped at 2 MiB and not dumped into shared logs.
  Report only fixed check names, status, versions, source/image hashes and time.
- Named containers and unique tooling-image tag are removed and checked absent.
  No global cache prune, existing image/container removal, marker reset or
  cleanup of any backup/identity path. Public base-image cache may remain.

No API, database, Caddy, production-derived config, actual account/session/node
identity, passphrase, recovery vault or backup disk is used. The synthetic HTTP
probe is loopback-only in the guest with a public fixture token; it does not
implement the MeshAlot customer workflow or establish its authentication.

## New acceptance checks

The actual guest must report all 16 checks true: PID 1 systemd; only loopback;
no external IPv4/IPv6 route; unit static verification; restricted UID service
start; drop-in precedence; no-new-privileges; no capabilities; denied unauthorized
synthetic request; stopped service; state preserved after restart; detected
deliberate fault; restored configuration; original fixture integrity; final stop.
Guest poweroff and outer boundary/cleanup checks must also pass. Performance
evidence is elapsed boot/test/shutdown time under these limits, not a product
latency or load benchmark. Do not repeat prior PG ACL, migration, Caddy/WSS,
bundle/key-tool, application, V100 or real-node acceptance suites for this work.

## Source pins

- OCI base: `ubuntu@sha256:fe81560990fae86dc4743202b2cb4d3df98d0bcf949173cecef15ce1127ef852`.
- Cloud disk: `https://cloud-images.ubuntu.com/releases/noble/release-20261017/ubuntu-24.04-server-cloudimg-amd64.img`.
- SHA256: `04e8f2461a16a44e78a370284471f29616cdd9baf1893df25807b71f985b6c94`.
- Checksum source: `https://cloud-images.ubuntu.com/releases/noble/release-20261017/SHA256SUMS`
  (also matched Canonical's current-release listing during review).

See the [execution evidence](../../../docs/m13-g17-systemd-stage-prerequisite.md)
for implemented, tested and gate status. No local/manual execution is requested
from the traveling operator.
