# M13 G17 — synthetic service-composition prerequisite

**IMPLEMENTED; MS-02 integration NOT RUN at this source checkpoint.
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

## Pending actual acceptance

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

The new workflow must verify the pinned tool/application bytes, compile the API
and synthetic probe, boot the capped no-NIC guest, and run the complete contract:
restricted PG/API/Caddy services; trusted HTTPS sign-in/enrollment; account
separation; signed WSS to real persistent telemetry/dashboard; API fault/restore;
PG restart and fail-closed API start; ledger/identity/session/data preservation;
logout, clean shutdown, containment and owned-resource deletion.

Actual integration status remains **NOT RUN** until a provider run and its
emitted results are recorded below. This is not real global-role recovery,
private-vault decryption, archived known-good service replay, production WSS,
public TLS recovery or physical provider-node acceptance. G17/1 and later holds
remain in force even if the synthetic composition passes.
