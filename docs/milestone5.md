# Milestone 5 — User Login and Web Dashboard

The Google Drive **AI Mesh POC Build, Test, and Release Guide** remains the source of truth for milestone gates.

## Milestone 4 reconciliation

Milestone 4 is **PASSED**. The older Milestone 4 section in `docs/milestones.md` was written before the final live verification and is stale where it still says the process-restart test or controlled live PostgreSQL deployment is pending.

Verified completion includes:

- tracked PostgreSQL migrations through `000003_node_persistence.up.sql` in the production `meshalot` database;
- restricted `meshalot` runtime-login permissions;
- append-only wallet, usage, and job-event protections plus job audit history;
- PostgreSQL-backed enrollment and heartbeat persistence;
- stable node identity and state across an actual API process restart;
- controlled live deployment on `meshalot-control-01` with the API healthy at `https://api.meshalot.com`;
- live enrollment and heartbeat persistence through a service restart;
- CPU, memory, disk, and HTTPS monitoring policies enabled, with HTTPS checks passing and notification delivery verified.

The one-off monitoring delivery-test policy is disabled rather than deleted.

## Milestone 5 status

Status: **PASSED**

Initial implementation commit: `ac4d1e8`

Production release commit: `9631776d5cbb07903c3f018b49cf9304c5e8200b`

Production release ID: `9631776d5cbb-20260904-162618`

Implemented without changing the existing agent enrollment/heartbeat path:

- PostgreSQL migration `000004_user_auth_and_sessions.up.sql` for password hashes and server-side web sessions;
- restricted runtime grants for authentication, user-scoped nodes, benchmarks, jobs, wallet data, rates, and sessions;
- email/password sign-in and sign-out endpoints;
- random HttpOnly browser sessions with only SHA-256 token hashes persisted in PostgreSQL;
- authenticated `/v1/auth/me` session check;
- user-scoped dashboard, node-list, node-detail, wallet, and job-history APIs;
- responsive React sign-in and dashboard shell;
- navigation/pages for Dashboard, My Nodes, Node Details, Compute Marketplace, Run Job, Job History, Wallet, and Node Economics;
- required dashboard cards for account balance, online nodes, compute score, network score, current status, current rate, and today's simulated earnings;
- isolated two-user API test covering sign-in, sign-out, node separation, wallet separation, dashboard/jobs access, and understandable authentication errors;
- credential-setting helper that does not echo the password or store plaintext credentials in source control.

### Milestone 5 release gate — PASSED

All required Milestone 5 release gates were demonstrated on September 4, 2026:

1. Go regression tests and the production web build passed from committed source.
2. Migration 4 applied and verified in an isolated PostgreSQL database before production deployment.
3. The real `meshalot` OS user and restricted PostgreSQL login passed the isolated runtime test.
4. Exact runtime privilege assertions passed, including required database `CONNECT`, schema/table/column rights, and rejection of excess write privileges.
5. Two separate isolated test users could not see each other's nodes or wallet entries.
6. Sign in, sign out, session invalidation, dashboard, nodes, wallet, and jobs APIs passed.
7. The exact production backend and React web release were staged and checksum-verified before cutover.
8. Production migration `000004_user_auth_and_sessions.up.sql` applied and verified successfully.
9. Narrowed production runtime grants and production privilege assertions passed.
10. Login credentials were set on the existing POC owner without exposing the password, hash, DSN, user UUID, or email in logs.
11. The live backend cut over to `/opt/meshalot/releases/9631776d5cbb-20260904-162618/meshalot-server` and remained healthy.
12. The existing live node remained present; an authenticated heartbeat returned HTTP 204 and persisted successfully.
13. Caddy validated and reloaded successfully, serving the React dashboard and SPA routes while preserving `/v1/*` reverse proxying.
14. Production `/`, `/dashboard`, and `/v1/health` returned HTTP 200; unauthenticated `/v1/auth/me` returned HTTP 401.
15. Production login returned HTTP 200 with a `Secure`, `HttpOnly`, `SameSite=Lax` session cookie; authenticated account APIs returned HTTP 200; logout returned HTTP 204; the invalidated session then returned HTTP 401.
16. HSTS, `X-Content-Type-Options: nosniff`, and `Referrer-Policy: no-referrer` were verified after deployment.
17. Five public HTTPS health checks averaged approximately 0.082 seconds during the production technical gate.
18. Final production state showed migrations 1–4 verified, one existing user, one existing node, one node-status record, no recent backend errors, and local/public health HTTP 200.
19. Rollback material was preserved at `/opt/meshalot/backups/9631776d5cbb-20260904-162618-cutover-20260904-163659`.
20. Final browser verification passed: the user signed in successfully, viewed the production dashboard, and confirmed the deployed web experience worked as expected.

No files were intentionally deleted during the Milestone 5 production staging or cutover workflow.

## Archive coverage caveat

`MeshAlot-Verbatim-Transcript-Part-2026-09-04.txt` is a **partial** transcript. It does not include the entire earlier conversation, and the later alert-setup conversation and subsequent handoff were not yet saved to Drive when this milestone began. Do not describe the archive as complete.
