# Synthetic M13 service integration fixture

This is the next G17/1 **target/integration prerequisite**, following the passing
[systemd target](../systemd_stage/README.md). It does not replay a real archive,
reconstruct real PostgreSQL globals, certify production TLS recovery, or close
G17/2. The application remains the accepted immutable candidate
`4f67ee54e5ebca084e38e306e723ae0acd08e77d`; its 49 required Go/database files are
individually pinned in `app-source.json`. No application change is included.

The traveling operator authorized routine nonproduction configuration/testing
and requested no terminal copy/paste. This fixture accepts no user arguments,
paths, endpoints, credentials or command overrides. It runs only under the exact
trusted MS-02 runner, host, user, repository and existing infrastructure ref.
The rejected production PostgreSQL diagnostic is not implemented or retried.

## Outcome and acceptance defined before execution

Prove the isolated target can run **real** PostgreSQL 16, M13 API and Caddy
services under systemd, with this synthetic customer workflow:

1. Verify source/artifact digests, no-NIC/no-route guest and outer containment.
2. Initialize a fresh private PG16 root/data directory. Create a separate
   nonadmin migration owner and restricted runtime role; apply exactly the
   pinned migrations 1–10 and tracked grants. This is not production migration.
   Require the expected major version and 10-entry ledger, not health alone.
3. Start API/Caddy/PG as three distinct restricted guest UIDs, no-new-privileges;
   confirm the API's actual restricted database connection. Caddy uses its
   internal issuer for guest-loopback IP only, with admin disabled and no public
   ACME request or trust-store installation. Clients verify its generated CA.
4. Sign in as two synthetic accounts; deny anonymous/bad login; require secure,
   HttpOnly session cookie; issue/redeem a synthetic-node enrollment code, deny
   reuse; enforce account/node separation and the online account dashboard.
5. Use the real M13 protocol with a **public deterministic test key**, no real
   identity file. Deny direct plaintext backend WSS, browser Origin and proof
   replay across connections. Require signed WSS and acknowledged paused
   heartbeats through Caddy into PG16, including after the API's 10-second
   ordinary HTTP timeout. The sampler is synthetic, not Node001 telemetry.
6. Break only the staged API ExecStart; require failed unit and Caddy 502.
   Restore the original staged definition and verify the same session and
   identity can reconnect. Do not roll back database schema/history.
7. Stop PG; require a fresh API start to fail rather than use memory mode.
   Restart PG/API, reconnect and verify the original ledger, users, node/key,
   session and dashboard. Reconnect checks repeat only after these actual faults.
8. Log out, verify session revocation, stop services, unmount seed, power off,
   verify outer exit/boundaries, remove only owned task resources and report.

**Stop on any failed acceptance, digest, isolation, resource, version, service,
workflow, persistence or cleanup check. No automatic retry or gate waiver.**
An emitted PASS must contain every required named guest check, with no skips,
all checks true, a clean emulator exit and successful cleanup. A new failure is
a reason for a targeted correction; unchanged passing suites are not repeated.

## Resources and artifact custody

- Existing MS-02 only, preflight minimum 4 logical CPUs, 10 GiB currently free
  RAM and 12 GiB free temporary storage. No cloud allocation or host install.
- Digest-pinned public Ubuntu/QEMU assembler, Go 1.24.13 compiler, PostgreSQL
  16 Alpine rootfs and Caddy 2.11.4 binary. APT verifies Ubuntu repository
  signatures; actual installed tool versions/image IDs are recorded.
- Compiler: 2 CPU quota, 3 GiB hard memory/no-extra-swap cap, 128 PIDs,
  no-new-privileges, no host mounts, no ports; only public source/dependencies.
  Compiler caches live in its disposable container. Assembly: 2 CPUs, 5 GiB,
  128 PIDs, no host mounts/devices/ports. Standard bridge is used only to fetch
  public compiler dependencies and Ubuntu packages, never production.
- PostgreSQL/Caddy source containers are fresh and **never started**; only their
  pinned public filesystems/binaries are read. Image-declared anonymous volumes
  are tracked and removed with only those owned containers; absence is checked.
- Runtime: 2 CPU quota, 5 GiB memory/no-extra-swap, 64 PIDs, read-only container
  root, UID/GID 65534, all capabilities dropped, no-new-privileges, Docker
  network none, no host bind/socket/device/port. Actual inspect must match.
- QEMU TCG: 2 virtual CPUs, 2048 MiB guest RAM, `-nic none`, no KVM/device
  permission change, no monitor or forwarding. At most 8 GiB virtual disk;
  writes confined to a 1536 MiB tmpfs qcow2 overlay with noexec/nosuid/nodev.
- Guest PG is in a private RootDirectory populated from the pinned public OCI
  rootfs. Extraction bounds member count/size, rejects traversal/devices,
  defers/rebases links and strips setuid/setgid. Eight focused security tests
  passed locally; the actual public rootfs layout also passed inspection.
  Eight unchanged tests are not rerun solely to rebuild this fixture.
- Guest database accepts only mapped local peer identities via a Unix socket;
  TCP is disabled. Both login passwords and the node key are public fixture
  constants. Real vault entries, encrypted archives, existing identity paths,
  GitHub secrets, production DSNs and recovery credentials are never supplied.
- Payload is read-only NoCloud ISO data. Its source and artifact receipts are
  checked before execution. No input artifact is extracted on the actual MS-02
  operating-system paths; rootfs execution/OS accounts exist only in the VM.
- Deadlines: 240-second cloud-download budget checked between reads (20-second
  socket timeout); individual pulls 180 seconds; compiler 600; assembly 600;
  emulator boot/workflow/shutdown 600; outer attach 660; job 50 minutes.
  Guest commands/health/readiness/proofs have shorter bounded waits. SIGTERM
  enters cleanup; hard runner loss cannot guarantee cleanup. No retry loop.
- Shared evidence contains fixed phase/check names, public hashes, versions and
  elapsed time. Raw guest logs, SQL rows, cookies, generated CA private keys and
  credentials are not printed or uploaded. Serial capture is capped at 2 MiB.
- Named task containers, their anonymous volumes, unique tooling-image tag and
  private temporary workspaces are removed; no global prune. Public base-image
  cache may remain. Existing backups, marker holds and host services are intact.

## Immutable public inputs

- Ubuntu base/cloud disk: the unchanged pins from the accepted systemd target.
- Go linux/amd64 OCI: `golang@sha256:757779acac4af1b349a20f357c7296097b4a0b89da4ad0e370b339060077282a`.
- PG16 linux/amd64 OCI: `postgres@sha256:cbf42f0edc946c0bc1b8eeed62403734e1f25dc67a3491455d07a92fdb558b28`.
- Caddy linux/amd64 OCI: `caddy@sha256:040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0`.

OCI platform manifests were verified at the official Docker registry before
publication. The new test does not rerun old standalone ACL/global restore,
Caddy recovery, bundle/key-check, V100, node or general application suites.
The reason for the new integration pass is the previously untested **actual
service composition and full customer API path in this systemd guest**.

See [the execution record](../../../docs/m13-g17-service-fixture.md). No result
here changes real archive, production push/merge/cutover, SSH or node holds.
