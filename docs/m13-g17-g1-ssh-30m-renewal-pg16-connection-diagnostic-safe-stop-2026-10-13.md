# MeshAlot M13 — G17/1 existing-SSH 30m renewal and PG16 connection diagnostic safe-stop

Date: 2026-10-13 (America/Chicago)
Status: **SSH RENEWAL PASS; SINGLE POSTGRESQL CONNECTION DIAGNOSTIC BLOCKED. NOT A DATABASE INSPECTION PASS.**
Gate 17/1: **IN PROGRESS / BLOCKED**. G17/2–G17/7: **NOT STARTED**. Milestone 13: **OPEN**.

## Separate operator approvals and authorized scope

The operator explicitly approved **one** 30-minute renewal of only the **existing Ed25519 instance-specific public SSH authorization** for `meshalot-control-01`, preserving unrelated VM/project metadata. No new key, permanent SSH access, broadened account rights, production service modification, or role/DB write was approved.

The operator additionally approved **one** read-only PostgreSQL connection diagnostic if the existing verified access was valid: exact existing nonprivileged identity, local Unix socket, no passwords/DSNs, no sudo, and only `SELECT 1`. Raw failure details were to remain owner-restricted on MS-02; GitHub reports sanitized classes only. No automatic repeat.

## Renewal: PASS, narrowly scoped

- Initial new-workflow push [run 38486753426](https://github.com/wwjd4u/MeshAlot/actions/runs/38486753426) failed at YAML workflow validation; **zero runner jobs**. This was not a renewal attempt.
- After repairing YAML quoting, [run 38486815874](https://github.com/wwjd4u/MeshAlot/actions/runs/38486815874) **PASS**. The local one-attempt guard was claimed and the runtime log verified existing Ed25519 identity, exactly one existing matching VM SSH record, 30-minute expiry, unchanged other project/instance metadata, preserved private historical registration, and no new SSH identity or production service/DB modifications. The corresponding source commit is [f80e280](https://github.com/wwjd4u/MeshAlot/commit/f80e28006371328731ff3d22d80948b3f15ed0a1).
- Renewal workflow then disabled in [commit 16650b9](https://github.com/wwjd4u/MeshAlot/commit/16650b9f2a6d23c5bf3a629567b550244ed5c4f9); subsequent [run 38486931077](https://github.com/wwjd4u/MeshAlot/actions/runs/38486931077) **SKIPPED**.
- Access is strictly temporary: do not assume the 30-minute registration remains valid or renew it automatically.

## Read-only connection diagnostic: BLOCKED, not an inspection PASS

- [One-time run 38487092673](https://github.com/wwjd4u/MeshAlot/actions/runs/38487092673) received GitHub Actions `success` because the existing safe-stop path returned normally. This **does not mean** the PostgreSQL diagnostic passed.
- Runtime evidence: `M13_G17_PG_CONNECTION_ATTEMPT=CLAIMED`, `M13_G17_PG_EXISTING_IAP_AUTH=VALID_WITH_MARGIN`, `M13_G17_PG_WRONG_HOST_GUARD=PASS`, and a shell warning that the generated remote Python heredoc was not properly terminated. The result was `M13_G17_PG_DIAGNOSTIC_CLASS=BLOCKED_SSH_OR_HOST_SAFE_STOP`.
- **No successful PostgreSQL `SELECT 1` result or diagnostic connection error class was obtained.** The historical blocker `BLOCKED_EXISTING_DB_ACCESS_UNAVAILABLE` remains unexplained. Do not assert server/socket access, role existence, peer/HBA status, or CONNECT permission from this attempt.
- The failed one-time diagnostic workflow was disabled; its malformed heredoc was corrected **in dormant source only** in [commit b9c766a](https://github.com/wwjd4u/MeshAlot/commit/b9c766aab449f4d2abc95086e28c108295f3294a). No corrected diagnostic execution occurred. [Follow-up run 38487197780](https://github.com/wwjd4u/MeshAlot/actions/runs/38487197780) was **SKIPPED**.
- Detailed SSH errors were retained only as local access-restricted evidence by the run's safe-stop handler. They were not published to Actions, GitHub, Google Drive, or chat. The attempt guard was consumed. **Do not silently retry or reset it.**
- No production database or service change was authorized or issued by the diagnostic workflow. The log does not support claiming a successful database inspection.

## Preserved outstanding recovery gates

- Earlier synthetic PostgreSQL16 ACL and global ownership reconstruction tests: **PASS — synthetic only**.
- Real production PostgreSQL16 role, memberships, owner/ACL and cluster-global recovery dependencies: **UNVERIFIED**.
- Real production Caddy/TLS storage/recovery dependencies: **UNVERIFIED**; isolated synthetic Caddy tests remain PASS.
- Full API/systemd/Caddy rollback stage: **NOT STARTED** (Gate 17/2).
- Production migrations 9–10: **NOT DEPLOYED**.
- Existing encrypted backups, independent Toshiba copies, Node001/Node002 identities, Node002 original M8 report, V100 benchmark evidence: **PRESERVE / DO NOT REPLACE OR RERUN**.
- October 13 full verbatim conversation transcript upload to Google Drive remains **UNVERIFIED**; do not call this summary a verbatim archive.

## Next security decision (NOT AUTHORIZED by this evidence)

A new actual diagnostic attempt requires **new explicit operator approval**, even though the previous run failed before classifying PostgreSQL. Revalidate the existing VM-specific SSH registration first, with a minimum validity margin. If the 30-minute authorization is expired or inadequate, **stop** and request separate SSH renewal authorization; do not infer it from approval for diagnosis.

Any later approved one-time diagnostic must use the corrected, reviewed and disabled source as reference, create its own separate non-replayable marker and log only an allowlisted error category. No sudo, key creation, password, `pg_hba.conf` modifications, database role/grant changes, data reads, production deployment or auto-retry. Continue strictly at **M13 G17/1**.

## Evidence references

- [G17/1 preceding PostgreSQL blocker](m13-g17-g1-pg16-metadata-retry-db-access-blocker-2026-10-13.md)
- [Gate 17 fixed completion sequence](m13-gate17-fixed-completion-sequence-2026-10-13.md)
