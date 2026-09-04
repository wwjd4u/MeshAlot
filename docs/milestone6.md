# MeshAlot Milestone 6 — Agent Identity and Secure Enrollment

Status: **PASSED**

Passed: September 4, 2026

Milestone 6 goal from the project guide: install one agent binary on a computer and securely attach it to an account with persistent node identity and replay-resistant enrollment.

## Production release

- Release source commit: `bb7c75b4cd0797f68fb4b8a5f26c3f2de26b3a65`
- Release ID: `bb7c75b4cd07-20260904-183650`
- Verified production backend SHA256: `620db365acde4660e3d95124b5c6ee7462d75c29ed17a7d580d635c671395b5d`
- M6 agent version: `0.1.0-m6`
- Production API: `https://api.meshalot.com`
- Production migration: `000005_agent_identity_and_enrollment.up.sql`
- Preserved M5 rollback directory: `/opt/meshalot/backups/bb7c75b4cd07-20260904-183650-cutover-20260904-191501`
- Preserved pre-M6 production source branch: `backup/pre-m6-source-20260904-195308`

## Delivered functionality

Milestone 6 added:

- persistent UUIDv4 node identity generated locally by the agent;
- persistent Ed25519 cryptographic identity generated locally;
- restrictive local identity-file permissions;
- public-key-only enrollment data sent to the control plane;
- browser-authenticated, short-lived, one-time enrollment-code generation;
- SHA-256 hashed enrollment codes at rest;
- atomic one-time enrollment redemption;
- account ownership enforcement and node-identity conflict protection;
- invalid, expired, fabricated, and already-consumed enrollment-code rejection;
- bounded enrollment attempts;
- secure enrollment UI under **My Nodes**;
- copy-only browser presentation of the raw enrollment code without localStorage or sessionStorage persistence;
- migration 5 identity/enrollment schema additions;
- narrowed runtime PostgreSQL permissions for the new enrollment path.

## Pre-production security and performance gate

The isolated Milestone 6 integration gate passed against a dedicated `meshalot_m6_test` PostgreSQL database while the live production service remained on M5.

Verified behaviors included:

- persistent identity across a new agent process;
- invalid-code rejection;
- expired-code rejection;
- already-consumed-code replay rejection;
- cross-account node-hijack rejection;
- simultaneous double-redemption with exactly one success;
- raw enrollment codes not stored in PostgreSQL;
- private identity material absent from PostgreSQL and server logs;
- restrictive local private-key permissions;
- least-privilege runtime database rights.

Enrollment-path performance sample:

- sample size: 20
- p50: 39.4 ms
- p95: 44.6 ms

## Web build gate

The production React/TypeScript/Vite enrollment UI build passed with:

- `go test ./...` successful;
- `npm ci --ignore-scripts` successful with zero reported vulnerabilities;
- TypeScript and Vite production build successful;
- enrollment-code endpoint present in the compiled bundle;
- browser copy action present;
- no `localStorage` or `sessionStorage` use for the enrollment code;
- one-time/expiration handling visible in the UI;
- existing Job History mapping regression-checked;
- production still unchanged on M5 during the gate.

## Production staging and cutover

A versioned M6 backend, agent, source archive, and web release were staged and hash-verified before cutover.

The live cutover then passed with:

- migration 5 applied and verified;
- M6 runtime database privileges applied and tested;
- M6 backend selected through a higher-priority `60-m6-release.conf` systemd drop-in;
- existing `50-m5-release.conf` preserved untouched for rollback;
- Caddy web root switched to the versioned M6 web release;
- exact M6 backend SHA verified after restart;
- unauthenticated enrollment-code creation rejected with HTTP 401;
- malformed secure enrollment rejected with HTTP 400;
- local API, public API, public web, and versioned M6 JavaScript asset all returned HTTP 200;
- public index hash matched the staged M6 web build.

## Real Node 001 release gate

Node 001 is the Ubuntu Minisforum MS-02 Ultra used for the POC.

The real production Node 001 gate passed:

- exact M6 source checked out and agent built;
- production enrollment accepted through a website-generated one-time code;
- a new persistent local identity was created;
- node identity fingerprint remained unchanged across a new agent process;
- public-key fingerprint remained unchanged across a new agent process;
- identity file owner matched the local user;
- identity file mode was `600`;
- reuse of the same consumed enrollment code was rejected;
- replay attempt returned a non-zero exit code;
- Node 001 identity remained unchanged after the failed replay.

The final server-side verification passed:

- the real Node 001 record was securely enrolled;
- agent version matched `0.1.0-m6`;
- public cryptographic identity was present;
- node status was `enrolled`;
- one production enrollment code was consumed and atomically bound to a node;
- no production enrollment code remained currently usable at the time of verification;
- migration 5 was registered and verified;
- Milestone 6 runtime privileges passed again on production;
- exact M6 backend binary and SHA were live;
- M6 systemd release override was effective;
- local API, public API, and public web health were all HTTP 200.

## Operational source closeout

After the M6 pass was recorded, `/opt/meshalot/source` was cleanly fast-forwarded from the M5 release commit to the exact M6 release source commit `bb7c75b4cd0797f68fb4b8a5f26c3f2de26b3a65`.

Closeout verification confirmed:

- a local backup branch `backup/pre-m6-source-20260904-195308` preserved the previous production source ref;
- the exact M6 release remained in current GitHub `main` history;
- GitHub commits after the runtime release were documentation-only at closeout;
- migration status remained verified through version 5 from the synchronized production source;
- the live M6 service PID did not change during source synchronization;
- the live backend SHA remained the verified M6 SHA;
- no database changes were made during source synchronization;
- no Caddy changes were made during source synchronization;
- no rollback material was deleted;
- local API, public API, and public web all remained HTTP 200.

## Milestone 6 pass criteria

**PASSED:** A node securely joins an account, preserves its identity through restart/new-process reload, and invalid enrollment attempts are rejected.

**PASSED:** Node identity is persistent and enrollment cannot be trivially replayed.

## Scope boundary preserved

Milestone 6 did not pull forward later work. Hardware inventory, network benchmarking, compute benchmarking, node scoring, provider resource controls, runtime adapters, and persistent agent heartbeat/connection behavior remain future milestones.

Milestone 6 is technically and operationally closed. Production source, versioned runtime release, migration state, web release, and preserved rollback chain are reconciled to the completed M6 state.
