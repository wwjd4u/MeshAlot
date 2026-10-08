# MeshAlot M13 — Gate 15 Closeout — 2026-10-07

**Gate 15 PASS — real existing Node001 identity with passive MS-02/V100 telemetry against an isolated local TLS/PostgreSQL server.**
**M13 not deployed to production. No enrollment or host service changes.**

## Source and scope
- Gate 14 closeout baseline: `1e18ddb781ad9b302e50b498ecc2dadfa0293443`.
- Opt-in local Node001 smoke-test implementation: `a126feecfad2c238d48ac90535346050e45dc7b3`.
- Test-fixture nullability/Go import fix: `0fa79c378ff5a5e3419d39797cb04a9c7f2dbcfc`.
- Runner formatting-only push: `819504dfde54ebc8ef888189320b9aa0e175056a`.
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469`, unchanged.
- Test file: `server/m13_node001_smoke_test.go` (opt-in only).

## What was actually verified
- The test could run only on the designated MS-02 host and only with the opt-in environment flag and disposable database credentials.
- Used `agent.DefaultIdentityPath` and **read-only** `agent.LoadIdentity` on the pre-existing real Node001 identity file. The private key stayed on the host, and no private material was printed or uploaded.
- Created an isolated disposable PostgreSQL 16 database and inserted an ephemeral test-owned node row with the same existing public key. This was NOT a MeshAlot enrollment request, production database operation, or change to the real node's identity.
- Ran the production-shaped M13 server handler over `httptest.NewTLSServer` on local loopback; the actual Node001 private identity signed the nonce-based WebSocket challenge, and its first real passive hardware heartbeat was accepted.
- The payload included real MS-02 CPU load, available RAM, and confirmed NVIDIA GPU utilization and available VRAM. The test explicitly reported provider status as **test-only paused** (away/manual-pause true) without modifying the actual provider settings or executing any AI workload.
- Verified the signed node's status and operational telemetry were stored correctly using restricted `meshalot` database privileges and displayed online on the isolated dashboard.
- Compared exact identity-file bytes before and after. Identity remained unchanged. No reboot, live WAN disconnection or server installation.
- V100 benchmark, evidence hashes and historical M8 reports were not rerun or changed.

## Runner proof
- [Successful workflow](https://github.com/wwjd4u/MeshAlot/actions/runs/37714575353), job `113108010903`.
- `TestM13Node001LocalIdentitySmoke`: PASS.
- Existing PostgreSQL migration and role permission tests: PASS.
- Full `go test ./...` and `go vet ./...`: PASS.
- Markers: `M13_NODE001_REAL_IDENTITY_LOCAL_TLS_GPU_AND_DB=PASS`,
  `M13_GATE15_EXISTING_NODE_IDENTITY_AND_GPU=PASS`,
  `M13_GATE15=PASS`.

## What still blocks live M13 closeout
- No production control-plane WebSocket endpoint or M13 database migration has been deployed.
- No systemd agent service has been installed or enabled.
- The local TLS test is not evidence of public WAN/WSS access through live Caddy.
- No deliberate physical internet loss, Node001 reboot, or Mac Node002 sleep/wake and launchd acceptance tests have been executed.
- Provider static service settings still require explicit owner verification before unattended enablement.

**Next Gate 16:** prepare a controlled staging and rollback plan, inspect deployment prerequisites read-only and request explicit operator authorization before any production DB migration, cloud deployment, service installation or disruptive node test.
