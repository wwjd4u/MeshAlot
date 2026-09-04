# Milestone status

The Google Drive **AI Mesh POC Build, Test, and Release Guide** remains the source of truth.

## Milestone 0 — Project Foundation

Status: **PASSED — 2026-09-03**

- [x] Planned top-level structure is present.
- [x] Environments, naming, and operating rules are documented.
- [x] Minimal API/agent vertical slice and automated integration test exist.
- [x] Clone, build, and run on the primary development machine.
- [x] Run the control API and agent together on the primary development machine.
- [x] Verify commit/push from the primary development machine.
- [x] Clone and build on a second test machine.

### Primary development-machine evidence

- Date: 2026-09-03
- Machine: `wwjd4u-MS-02-Ultra`
- Operating system: Ubuntu 26.04
- Go: `go1.26.0 linux/amd64`
- Branch: `main`
- Tested commit: `591a5f8`
- Command: `go test ./...`
- Result: **PASS**
- Server integration package completed in `0.003s`; all other packages compiled successfully.

### Live API/agent evidence

- Control API: `127.0.0.1:8180` (port 8180 avoids the existing local llama.cpp service on port 8080)
- Agent: `meshalot-development-node-001`
- Agent version: `0.0.1-dev`
- Enrollment: **PASS**
- Heartbeat: **PASS**
- `GET /v1/health`: **PASS**, protocol `v1`, status `ok`
- `GET /v1/nodes`: **PASS**, node status `online`, mode `available`

### Primary source-control evidence

- Authentication: GitHub CLI authenticated as `wwjd4u`
- Git protocol: HTTPS
- Push commit: `c40cfe9`
- Remote update: `0806520..c40cfe9 main -> main`
- Local `HEAD`, `origin/main`, and GitHub `main` matched after push
- Result: **PASS**

### Second development-machine evidence

- Date: 2026-09-03
- Machine: `Jasons-MacBook-Pro`
- Operating system: macOS 26.5.1, Intel `x86_64`
- Go: `go1.27.1 darwin/amd64`
- Branch: `main`
- Tested commit: `773dfb7`
- Command: `go test ./...`
- Result: **PASS**
- Server integration package completed in `0.480s`; all other packages compiled successfully.

## Milestone 1 — Google Cloud Account and Billing Guardrails

Status: **PASSED — 2026-09-03**

- [x] Dedicated MeshAlot Google administrator account and Google Cloud project created.
- [x] Paid billing attached to the single POC project.
- [x] Monthly alert budget set to $25 with low-threshold and forecasted alerts.
- [x] Only the milestone-required Compute Engine and Cloud Logging APIs enabled.
- [x] Primary Google Cloud region selected: `us-south1` (Dallas).
- [x] No VM instances existed before network provisioning.
- [x] Unused Google-created default firewall rules and default auto-mode VPC were deleted with explicit authorization.
- [x] Post-deletion checks showed no remaining VPC networks or firewall rules.
- [x] Administrator and billing identifiers recorded in a private Drive register and omitted from this public repository.

### Milestone 1 release gate

- Build: **PASS**
- Test: **PASS**
- Pass criteria: **PASS**
- Do-not-proceed check: **PASS**
- Result: **PROCEED TO MILESTONE 2**


## Milestone 2 — Dual-Stack Google Network

Status: **PASSED — 2026-09-03**

- [x] Custom-mode VPC created with regional routing.
- [x] Dallas subnet created with private IPv4 and external IPv6.
- [x] Private Google Access enabled.
- [x] Management firewall permits TCP 22 only from Google IAP and only to tagged instances.
- [x] No public database, local-AI runtime, or unrestricted SSH ports exposed.
- [x] Temporary dual-stack Debian VM launched on the subnet.
- [x] IAP-only SSH access verified.
- [x] Outbound IPv4 returned HTTP 204.
- [x] Outbound IPv6 returned HTTP 204.
- [x] Temporary test VM and attached boot disk permanently deleted after explicit authorization.
- [x] Purpose-built VPC, subnet, and restricted IAP firewall rule retained.

### Milestone 2 release gate

- Build: **PASS**
- Test: **PASS**
- Pass criteria: **PASS**
- Do-not-proceed check: **PASS**
- Result: **PROCEED TO MILESTONE 3**

## Milestone 3 — First Control Server

Status: **PASSED — 2026-09-03**

- [x] Permanent dual-stack `e2-small` control-plane VM deployed in `us-south1-a`.
- [x] Static IPv4 and IPv6 addresses assigned and DNS published for `api.meshalot.com`.
- [x] Git, Go, PostgreSQL 16, and Caddy installed.
- [x] MeshAlot server built from the repository and run as the non-login `meshalot` service account.
- [x] Enrollment secret stored in a root-controlled environment file and omitted from logs and source control.
- [x] MeshAlot, PostgreSQL, and Caddy enabled as automatic system services.
- [x] MeshAlot API bound privately to `127.0.0.1:8180`; PostgreSQL bound privately to `127.0.0.1:5432`.
- [x] Public HTTPS limited to versioned `/v1/*` API routes; unrelated paths return HTTP 404.
- [x] Valid Let's Encrypt TLS certificate issued for `api.meshalot.com`.
- [x] Public IPv4 health request returned HTTP 200.
- [x] External IPv6 ingress and TLS request observed successfully; Cloud Shell itself has no IPv6 default route.
- [x] Structured JSON application and reverse-proxy logging enabled.
- [x] Two independent reboot cycles produced new boot IDs and automatic service/API recovery.
- [x] Compute Engine reset returned success; post-reset public health returned HTTP 200.
- [x] Current-boot error review found only a transient rsyslog `/dev/console` write warning.

### Milestone 3 release gate

- Build: **PASS**
- Test: **PASS**
- Pass criteria: **PASS**
- Do-not-proceed check: **PASS**
- Result: **PROCEED TO MILESTONE 4**

## Milestone 4 — Minimum Database Schema

Status: **IN PROGRESS — ISOLATED DATABASE TESTS PASSED**

- Follow-up audit/ledger migration and isolated PostgreSQL test harness added.
- Original core migration preserved; no live database or API deployment performed.
- Test schema target: `meshalot_test`; fixtures roll back after verification.
- PostgreSQL 16 isolated test execution: **PASS**, user-reported output from commit `4a7b1cd`.
- Both migrations committed successfully; all 13 core tables were created.
- Core records, automatic job audit, role reversal, reconstructed balances,
  duplicate/sign/party rejection, and append-only guards passed.
- Test fixtures rolled back; migrated test schema retained. TRUNCATE notices
  were generated by an expected-rejection test, not a successful truncation.
- Restricted test-role verification: **PASS**, user-reported output. Required
  job/usage/ledger writes succeeded; all ten forbidden operations were denied.
  Fixtures rolled back. This used SET ROLE, not a runtime login connection.
- Version/checksum runner: eight local construction tests passed.
- PostgreSQL tracking integration tests: **PASS**, user-reported output from
  commit `d83f15c`, verified 2026-09-04 UTC.
- Existing test schema registered explicitly as `adopted_test`; versions 1 and 2
  matched pinned SHA-256 checksums without re-executing their migrations.
- Repeat application skipped recorded migrations; altered checksum was rejected;
  intentionally failed migration rolled back schema changes and history.
- Final status showed both versions verified with unchanged registration times.
- Pending: runtime-role deployment and application persistence verification.
  Fresh-database application of the full tracked migration batch also remains
  to be tested before application deployment. Do not advance yet.

Implement migrations for accounts, nodes, status, hardware and network benchmarks, models, jobs and job events, usage, the append-only simulated-money wallet ledger, pricing rates, and enrollment tokens. Prove that a test user, node, benchmark, job, usage record, earning, and charge can be created and read without manual database edits; verify that job history is auditable and wallet balance can be reconstructed from ledger entries.
