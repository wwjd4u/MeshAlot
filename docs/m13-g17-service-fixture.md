# M13 G17 — synthetic service-composition prerequisite

**SYNTHETIC SERVICE COMPOSITION PASS; all 56 guest checks and cleanup PASS.
G17/1 BLOCKED. Actual archived G17/2 rehearsal NOT STARTED.**

The [acceptance/resource/stopping contract](../tools/recovery/service_fixture/README.md)
was recorded before execution. The operator authorized nonproduction testing
without repeated permission and asked for work without terminal copy/paste
while traveling. This work addresses the unproven capacity/integration of the
new systemd target with actual PG16/API/Caddy, not the already-passing dummy
service, standalone ACL, Caddy/WSS, archive-key or application suites.

The accepted application candidate is unchanged at
`4f67ee54e5ebca084e38e306e723ae0acd08e77d`; 49 required source/database files
have independent file-digest pins. No real archive or production credential is
an input. All OS accounts, sessions, CA state, peer mapping and node identity
belong to the disposable guest and synthetic fixture.

## Before-run evidence

- Python/shell syntax checks PASS.
- Eight new rootfs extraction boundary tests PASS (0.004 seconds locally):
  rebased absolute link and setuid removal, member traversal/absolute path,
  escaping symbolic/hard link, regular hardlink, special-device rejection and
  parent-link collision. These defend a new extraction boundary.
- The actual pinned public PG16 rootfs layout passed local extraction review;
  the fresh source container never started and was removed with its anonymous
  volume. The pinned Go formatter parsed/formatted the new probe; that fresh
  source container also never started and was removed.
- No old passing suite was rerun. No main/production, existing backup, recovery
  key, Node001/Node002, host package/service, permission or attempt marker changed.

## Execution history retained

The first [run 41119563622](https://github.com/wwjd4u/MeshAlot/actions/runs/41119563622)
at infrastructure `88cf1c1a7bd467f7395668871f452b9171ad6fd6`, tooling source
`50241e1f0069b9334146b759bfca913b5f2ee0d1`, **FAILED** during rooted PG startup.
All 49 application source pins, Go 1.24.13 static API/probe builds, public
artifact hashes, cloud disk, package assembly, outer containment, real systemd
boot and payload integrity passed. PG did not become reachable at the proposed
socket; the guest reported CONNECTION/FILE_NOT_FOUND from psql. This does not
establish the original launch cause. Customer acceptance was **NOT RUN**.
Task containers/image/anonymous-volume and source-workspace cleanup passed.

The targeted correction supplies matching group records, uses explicit numeric
UID/GID for the rooted PG service, locates its socket outside transient /run,
and makes only the public guest executable directory traversable. Peer auth,
role separation, no-new-privileges, network isolation and data/key modes are
preserved. A bounded service-state/journal-classification report will distinguish
any remaining launch failure. The new resource/setup change warrants a follow-up
integration run; it does not warrant rerunning unchanged standalone suites.

The second [run 41120996165](https://github.com/wwjd4u/MeshAlot/actions/runs/41120996165)
at infrastructure `c33dcd8b4f39299fbb3a9f77af97b73e58945ff1`, tooling source
`46ef8ebb43e19580c09f88582fe443e353b121a6`, **FAILED** during unit static
verification. The rooted PG16 service now started, all ten pinned migrations
and runtime grants applied, and separate database ownership was verified.
Cleanup PASS. Actual customer workflow acceptance remained NOT RUN.

A local public synthetic reproduction confirmed that systemd-analyze resolves
ExecStart against its own filesystem for this check, ignoring RootDirectory.
The explicit rooted lint view passed locally. The fixture now checks an
otherwise identical lint-only PG unit copy with that single executable path
resolved for the checker; the actual service unit is unchanged. Actual rooted
startup and restricted service execution remain mandatory. No check was marked
PASS from an ignored failure and no authentication/isolation requirement changed.

The third [run 41122908774](https://github.com/wwjd4u/MeshAlot/actions/runs/41122908774)
at infrastructure `04d0afc5bf264e274ad47eabf101b4fcac411911`, tooling source
`ffea405985ae4d573649b4c1c6bbea2146e2bc38`, **FAILED** during customer checks.
All three real services, rooted unit verification, actual restricted API database
role, each service's UID/no-new-privileges, trusted HTTPS and secure HttpOnly
session cookie passed. Cleanup PASS. Heartbeat/fault/persistence acceptance had
not completed.

Source review identified a fixture expectation error: the secure-enrollment
handler returns HTTP 201 on success; this new fixture expected 200. That
expectation is corrected against the pinned source, with exact status-only HTTP
failure reporting added. No application response, authentication rule or other
acceptance criterion is changed. A follow-up is required for this failed new
customer path; old standalone suites remain reusable.

The predefined acceptance required pinned tool/application bytes, compiled API
and synthetic probe, a capped no-NIC guest, and the complete contract:
restricted PG/API/Caddy services; trusted HTTPS sign-in/enrollment; account
separation; signed WSS to real persistent telemetry/dashboard; API fault/restore;
PG restart and fail-closed API start; ledger/identity/session/data preservation;
logout, clean shutdown, containment and owned-resource deletion.

The successful receipt below supersedes the pending synthetic integration
status. The earlier failed runs remain failure evidence, not acceptance PASS.

## Actual successful MS-02 receipt

- [Run 41124047221](https://github.com/wwjd4u/MeshAlot/actions/runs/41124047221),
  job **124297833349**, completed **SUCCESS**. Provider workflow timestamps:
  **2026-10-30 20:32:36Z–20:40:15Z** (15:32:36–15:40:15 America/Chicago).
- Infrastructure SHA: `8a108a74498dc19e111f64cbc67eac1588c9e5bd`.
- Tooling source: `894c948d31c78a0338706c17807ad278a31ae7a9`;
  all eight fetched file-digest pins PASS.
- Application: `4f67ee54e5ebca084e38e306e723ae0acd08e77d`;
  all **49** source/database file pins PASS; compiler **Go 1.24.13 linux/amd64**.
  Application/protocol/database/module source remains unchanged on M13.
- QEMU **8.2.2**, package `1:8.2.2+ds-0ubuntu1.18`; Python
  `3.12.3-0ubuntu2.1`, genisoimage `9:1.1.11-3.5`, seabios `1.16.3-2`.
- Public asset manifest SHA256:
  `c0a9fc06e235410e62507fdaac3e363ca8e8c105c7bfaf254367e324361bbbbf`.
- API SHA256:
  `91867dbeac357b40a31777e8692d140c7a482f178dd3dc97ae75b7ffa05cb45a`.
- Probe SHA256:
  `89ceda41e4c64eb69bb8de89f3cc5b947c25be39ced20a2876280dae3fce4fd4`.
- Caddy SHA256:
  `b7105518e3ed1c0761f232e44fc09345535533c9cb0abf0e12809416c7ac64d9`.
- Public PG-rootfs tar SHA256:
  `e9d63662a354bd4e1366f31f61864f966c7e4ae32587ad7b0cdc1bd01f06cda2`.
- Disposable tooling-image receipt:
  `sha256:c011a1a3af28c7aa9aff6378a25e176be7394591b59f6febc52efde1f4d8ab67`.
  This task image was removed after use; it is not a retained running stage.
- Guest boot/workflow/shutdown: **268.105 seconds**, below the 600-second cap.
  Guest acceptance itself: **46.884 seconds**. No production performance claim.

All **56 emitted guest checks were true**, with no skipped path or ignored
failure. Required checks include:

| Area | Actual acceptance |
| --- | --- |
| Source/isolation | Payload digests, systemd PID 1, only loopback, no external IPv4/IPv6 route PASS |
| Database | Actual PG16, pinned 10-entry schema, separate owner/runtime role, actual restricted API database connection PASS |
| Services | Direct API/Caddy unit checks and explicit rooted PG lint view PASS; actual PG/API/Caddy restricted UIDs and no-new-privileges PASS |
| Customer authentication | Trusted HTTPS sign-in, bad/anonymous denial, secure HttpOnly cookie, successful synthetic enrollment and consumed-code denial PASS |
| Transport/telemetry | Signed WSS and acknowledged paused heartbeat through real Caddy to PG PASS; plaintext backend, browser Origin and cross-connection proof replay denied; WSS survived ordinary HTTP timeout |
| Account boundary | A sees its node and online dashboard; B sees no node and cannot read A's node PASS |
| Fault/restore | Missing stage-only API executable detected, proxy 502, original staged unit restored, same session and signed identity reconnect PASS |
| Durable workflow | API refused startup without PG; PG/API restart retained session, ledger version/digest pairs, node/key and fixture row counts; signed reconnect and final online dashboard PASS |
| Closeout | Logout revoked session; all three services stopped, seed unmounted and VM powered off PASS |

The initial WSS probe required all six of its transport/auth/heartbeat checks;
only valid reconnect checks were repeated after each deliberate service fault.
No standalone old suite was rerun. The API/probe/Caddy artifact digests stayed
identical across the fixture corrections; its application source never changed.

Outer containment inspections before and after execution, no-OOM/clean exit,
owned containers/anonymous volumes/unique image removal and temporary source
workspace cleanup all **PASS**. Raw guest logs, SQL rows, cookies and CA private
keys were not published. Public base-image caches may remain; no global prune.
The receipt used actual emitted lines, excluding echoed workflow source.

The cloud-workspace capacity metadata was also inspected while the follow-up
was queued; no alternative cloud VM/profile was created or executed. Actual
acceptance above belongs to the existing MS-02 runner.

## Material-risk review and gate decision

This closes the **synthetic service-composition prerequisite**. The tested
design supports the later controlled rehearsal; all task VMs were destroyed.
Its internal CA, public fixture credentials, synthetic node and generated units
are test substitutes. The result is not a browser/UI test, production load
benchmark, real global-role recovery, private-vault decryption, known-good v8
archive replay, public TLS recovery or physical provider-node acceptance.

**G17/1 remains BLOCKED; G17/2–7 remain NOT STARTED.** Remaining entry evidence:
actual vault-sourced decryption, protected real PG globals/ownership recovery,
remaining real TLS recovery prerequisites and approved archived service replay.
The specifically rejected PostgreSQL diagnostic remains paused. Main, real
backups, original identity material, old disabled workflows/markers and host
services were not modified. No deployment, SSH renewal or production connection.

Next supported action: use this verified design when the real recovery entry
evidence is available. Defer the independent Mac/vault check until the operator
is parked; no manual command is requested while traveling. Publish this receipt
without repeating the passing fixture: all eight tested executable/input files
remain byte-identical, and the final update changes documentation only.
