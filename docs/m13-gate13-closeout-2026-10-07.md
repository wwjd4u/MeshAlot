# MeshAlot M13 — Gate 13 Closeout — 2026-10-07

**Gate 13 PASS — production-shaped TLS proxy / loopback backend / long-lived WSS connection.**
**M13 IN PROGRESS, not deployed to production.**

## Exact source
- Prior Gate 12 closeout baseline: `cbb15ead6e2eeb2bce73d6b708986fa9a2af013a`.
- Initial Gate 13 test commit: `4a323589246e4fd943f45c6d37a59359ac198c47`.
- Final MS-02 runner format-and-push test commit: `3fb8c221b379cafe6fececcf86635168f26a453d`.
- Production `main`: `1016c9b058690f34058dfab9cde7927e421fa469` unchanged.

## Verified isolated architecture
- Real agent's signed WSS challenge response to a test HTTPS edge proxy, then an unencrypted loopback HTTP backend with the trusted HTTPS-forwarded protocol header, then real M13 server handler and disposable PostgreSQL 16.
- Direct plaintext WebSocket to backend WITHOUT verified HTTPS reverse proxy rejected with 403.
- The backend HTTP server was deliberately configured with extremely short 250ms ReadTimeout, WriteTimeout and IdleTimeout. A WSS connection successfully submitted and persisted two heartbeats spaced at least 400ms apart, showing the upgrade did not incorrectly inherit ordinary HTTP request timeouts.
- Latest `node_status`, same enrolled identity and dashboard online-node count verified through PostgreSQL after the second heartbeat.
- Test used a new isolated fake node/key pair, local root-of-trust test certificate, and non-root `meshalot` DB role; Node001 and Node002 identities untouched.
- This test models the *Caddy-style trust boundary*; it does not assert a real Caddy configuration change or production deployment.

## Runner proof
- [Successful run](https://github.com/wwjd4u/MeshAlot/actions/runs/37714237556), job `113106937822`.
- `TestM13ReverseProxyRealPostgres`: PASS.
- The focused test was also rerun with Go race detection: PASS.
- All ten DB migrations, restricted runtime grants, full Go test suite, `go vet` and formatting verification: PASS.
- Log markers: `M13_PROXY_TLS_LOOPBACK_UPGRADE_LONG_LIVED_WSS=PASS`, `M13_GATE13_END_TO_END=PASS`, `M13_GATE13_CODE=PASS`.
- First infrastructure workflow creation had invalid YAML from a substitution error and therefore did not launch any job. Its corrected workflow ran successfully. No production or node changes.

## Remaining live gates
1. Confirm the real Node001 identity path and existing service/sampler prerequisites **read-only**.
2. Review Caddy's deployed WSS forwarding policy and DB migration plan; capture rollback artifacts.
3. Controlled, explicitly authorized staging deployment and migration; check WSS from Node001.
4. Verify Node001 provider controls, disconnect/reconnect, 90s offline classification, reboot recovery and no re-enrollment.
5. Verify Node002 launchd, real Mac telemetry, sleep/wake and reconnect without identity change. Do not modify Node002 historical reports.
6. Only after live acceptance and security checks consider controlled M13 production cutover; do not merge or deploy this branch automatically.
