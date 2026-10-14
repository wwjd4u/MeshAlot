# MeshAlot M13 — Gate 17/1: approved SSH 60m renewal and PostgreSQL connection diagnosis

Date: 2026-10-13 (America/Chicago; Actions logs may use 2026-10-14 UTC)
Status: **SSH renewal PASS; PG16 connection failure CLASSIFIED as BLOCKED_DATABASE_ROLE_ABSENT; production catalog inventory NOT PASSED.**
Milestone 13 **OPEN / NOT RELEASE-READY**. G17/1 **IN PROGRESS**; G17/2–G17/7 **NOT STARTED**.

## Operator authorization and exact limits

The operator authorized ONE 60-minute renewal of the **existing VM-instance-specific Ed25519 public SSH authorization** for `meshalot-control-01`, conditional on offline Bash/Python syntax validation of the corrected connection diagnostic passing. The operator separately approved ONE new PG16 read-only local-socket `SELECT 1` diagnostic under currently verified access with no password, new SSH keys, privilege escalation, sudo, DB permissions update, configuration change, production deployment or repeat attempt. Preserve all other VM/project metadata, security policies, node identities and existing encrypted backups. No permanent access was approved.

## Completed source-pinned offline validation — PASS

1. Original corrected diagnostic and prior verified renewal-template syntax: [run 38490926580](https://github.com/wwjd4u/MeshAlot/actions/runs/38490926580) **PASS**. Bash `-n`, embedded Python AST checks, and exact Git source hash checks; no production/SSH access. Its workflow was disabled and the next run SKIPPED.
2. **Final** 60-minute renewal draft and new SELECT1 diagnostic draft: [run 38491177738](https://github.com/wwjd4u/MeshAlot/actions/runs/38491177738) **PASS**. The trusted MS-02 runner reported `PG16_DIAGNOSTIC_BASH_SYNTAX=PASS`, `PG16_DIAGNOSTIC_EMBEDDED_PYTHON_SYNTAX=PASS`, `PG16_DIAGNOSTIC_SOURCE_PIN=PASS`, `SSH_RENEWAL_TEMPLATE_BASH_SYNTAX=PASS`, `SSH_RENEWAL_TEMPLATE_EMBEDDED_PYTHON_SYNTAX=PASS`, `SSH_RENEWAL_TEMPLATE_SOURCE_PIN=PASS`, and `M13_G17_OFFLINE_PREFLIGHT=PASS`. No script was executed except safe syntax/AST checks, with no production access or SSH renewal. The one-time preflight was disabled in [commit b79025b](https://github.com/wwjd4u/MeshAlot/commit/b79025bdd42eeafc0280a1ba77a7dbd1555fa5f1), follow-up SKIPPED.

## One approved 60-minute SSH renewal — PASS

[Run 38491246917](https://github.com/wwjd4u/MeshAlot/actions/runs/38491246917) **PASS**. The guarded MS-02 renewal re-used the same existing private/public Ed25519 identity, replaced only the matching VM-specific public SSH metadata entry with a bounded 60-minute expiry, preserved all other instance/project metadata, and retained the previous local registration privately. Runtime audit markers included `M13_G17_60M_ATTEMPT=CLAIMED`, `M13_G17_EXISTING_PUBLIC_IDENTITY_MATCH=PASS`, `M13_G17_INSTANCE_ONLY_PUBLIC_KEY_RENEWAL=APPLIED`, `M13_G17_NEW_VM_KEY_ENTRY_VERIFIED=PASS`, `M13_G17_OTHER_VM_METADATA_UNCHANGED=PASS`, `M13_G17_PROJECT_METADATA_UNCHANGED=PASS`, `M13_G17_PREVIOUS_REGISTRATION_PRESERVED_PRIVATE=PASS`, and `M13_G17_NO_NEW_SSH_IDENTITY=PASS`.

The renewal workflow was disabled in [commit b7001e2](https://github.com/wwjd4u/MeshAlot/commit/b7001e241afe294859fac8a0e154a4ee42c683ca); its [follow-up run 38491314309](https://github.com/wwjd4u/MeshAlot/actions/runs/38491314309) was **SKIPPED**. This access remains time-limited; **do not assume it is still valid or renew it automatically**.

## One approved PG16 `SELECT 1` connection diagnostic — SAFE, failure cause identified

[Run 38491337730](https://github.com/wwjd4u/MeshAlot/actions/runs/38491337730) completed its **one authorized attempt**. The sanitized runtime markers verified `M13_G17_PG_CONNECTION_ATTEMPT=CLAIMED`, `M13_G17_PG_EXISTING_IAP_AUTH=VALID_WITH_MARGIN`, and `M13_G17_PG_WRONG_HOST_GUARD=PASS`.

**Result:** `M13_G17_PG_DIAGNOSTIC_CLASS=BLOCKED_DATABASE_ROLE_ABSENT`.

The original metadata audit had attempted local Unix-socket PostgreSQL access as the currently authorized **nonprivileged Linux account**, selecting its matching PostgreSQL role name with no credentials or sudo. The connection failed because **that requested matching PostgreSQL role does not exist**. This does **not** establish that the actual MeshAlot runtime database role is missing. It does **not** prove `CONNECT` access, pg_hba/peer policy for an existing permitted role, or production global role/ownership/ACL recovery. `SELECT 1` was attempted but **did not execute successfully**. No PostgreSQL catalog metadata was obtained.

GitHub Actions `success` describes successful execution of the fail-closed diagnostic/classification, **not a successful database connection**. Runtime controls also reported `M13_G17_PG_DETAILED_ERRORS=PRIVATE_LOCAL_ONLY`, `M13_G17_PG_NO_SUDO_OR_PRIVILEGE_ESCALATION=PASS`, `M13_G17_PG_PASSWORDS_OR_CUSTOMER_ROWS=NOT_READ`, `M13_G17_PG_PRODUCTION_CHANGES=NONE`, `M13_G17_PG_CATALOG_METADATA=NOT_QUERIED`, and `M13_G17_PG_PRIVATE_EVIDENCE_SHA256=PASS`. Raw errors remain privately stored on MS-02; no usernames, credential material, protected connection text, or customer rows were placed in this public evidence or GitHub Actions logs.

The diagnostic workflow was disabled in [commit 7816afb](https://github.com/wwjd4u/MeshAlot/commit/7816afb02914aded68160aabfc558ea0055d6650). [Follow-up run 38491394672](https://github.com/wwjd4u/MeshAlot/actions/runs/38491394672) was **SKIPPED**. Its separate non-replayable attempt marker was consumed. No automatic retry, no resetting markers, no manual invocation of completed one-shot paths.

## Least-privilege correction proposal — NOT EXECUTED

The diagnostic used an unprivileged Linux login identity that is **not** a PostgreSQL role. Do not manufacture a new role/grant or weaken authentication merely to make an audit succeed. Determine the *pre-existing* database operator/application identities and allowed local authentication route from already-approved nonsecret documentation/code or an explicitly approved operator-controlled read-only review. Choose an **existing, authorized database principal** with sufficient metadata visibility, and request **separate explicit scope approval** for any future limited catalog assessment. If using a different Linux identity, `sudo -u postgres`, elevated DB identity, secret-bearing connection settings, or a production config read becomes necessary, STOP until separately approved; prefer no production auth/permission changes. Protect nonsecret role names and topology details in restricted local evidence; publish only sanitized classes and counts.

Existing synthetic PG16 ACL and role/ownership-restoration PASS results remain accepted but do not substitute for actual production global role inventory and protected recovery asset. Caddy/TLS real storage and full rollback staging remain unverified. G17/2–G17/7 remain NOT STARTED; migrations 9–10 remain undeployed. Preserve encrypted backups and Toshiba independent copies, Node001/Node002 identities and original Node002 M8 report, and previous V100 evidence. The full October 13 verbatim archive to Google Drive remains UNVERIFIED; this sanitized report is **not** a verbatim transcript.

**Next operator decision:** approve only a separately reviewed, least-privileged inventory route using existing authorized DB access. No next production connection is authorized by this report.
