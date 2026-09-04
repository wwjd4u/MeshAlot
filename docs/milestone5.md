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

Status: **IN PROGRESS — CODE READY FOR ISOLATED AND LIVE GATE TESTS**

Initial implementation commit: `ac4d1e8`

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

### Milestone 5 release gate still required

Do **not** mark Milestone 5 passed until all of the following are demonstrated:

1. Go tests and the web production build pass from the committed source.
2. Migration 4 applies successfully to an isolated PostgreSQL database.
3. The restricted application role has only the required Milestone 5 access.
4. Two separate test users cannot see each other's nodes or wallet entries.
5. Sign in and sign out work.
6. All Milestone 5 pages load and API failures show understandable errors.
7. The production migration and binary/web deployment succeed without exposing secrets.
8. The existing live node remains present after deployment and its heartbeat continues to persist.
9. Public HTTPS health remains good after the web dashboard is enabled.

## Archive coverage caveat

`MeshAlot-Verbatim-Transcript-Part-2026-09-04.txt` is a **partial** transcript. It does not include the entire earlier conversation, and the later alert-setup conversation and subsequent handoff were not yet saved to Drive when this milestone began. Do not describe the archive as complete.
