# Gate 17 offline recovery bundle preflight

This README covers the synthetic-only check_bundle.py interface. The separate
[private archive key-access verifier](ARCHIVE_KEY_CHECK.md) automates a later
operator-local recovery-key check; its CI uses synthetic encrypted files only.
It does not relax this bundle tool's material boundary or permit real archives
or keys in GitHub/Actions.

This tool implements one **offline integrity prerequisite**, not a full recovery
acceptance harness. It prepares the existing G17 path while the rejected
service-identity diagnostic remains paused. It has no SSH, subprocess, network,
SQL, archive extraction, service management or workflow-dispatch functionality.
Do not supply production-derived files, recovery archives, environment files,
TLS keys, credentials or enrolled-node identities under the current scope.

## Verified baseline and integration

Remote refs checked during this session:

- M13 evidence head: `3b67f57afcda8bdd98989a2b3ae68d42c1d1b19c`.
- Accepted M13 application candidate: `4f67ee54e5ebca084e38e306e723ae0acd08e77d`.
- Infrastructure evidence head: `54bb072463f257c1e1b75484244dfa82ee894b50`.
- Main: `1016c9b058690f34058dfab9cde7927e421fa469`.

The GitHub comparison from the accepted M13 candidate to the evidence head
contains documentation and two audit scripts, with no application, migration or
test changes. This is a source-equality observation, not fresh runtime evidence.
The ordinary Git fetch initially could not reach the configured proxy. Using the
executor's explicit network permission restored normal GitHub HTTPS fetching;
no proxy, credential or network policy was changed. Development continues on the
existing M13 line, local branch `m13-persistent-heartbeat`, based on the exact
evidence SHA above. The checkout has a
main-only fetch mapping, so the local M13 branch has no upstream tracking
configuration. Remote refs and immutable SHAs were checked explicitly. The
existing Go CI workflow now includes the focused Python tests for future
PR/main checks. Its triggers are unchanged: a push to M13 does not run it.
Publication of this preparation change does not close G17 or deploy the app.
The temporary preparation branch/worktree was removed at the operator's request;
all new work is in the original checkout.

## Interface and boundaries

Requires Linux and Python 3.8 or later; verified here with Python 3.12.14. Supply
an owner-controlled synthetic fixture directory and **independently reviewed**
manifest/source pins:

```sh
python3 -B tools/recovery/check_bundle.py \
  --root /path/to/private-synthetic-fixture \
  --manifest manifest.json \
  --expected-source REVIEWED_40_CHARACTER_SOURCE_SHA \
  --expected-manifest REVIEWED_64_CHARACTER_MANIFEST_SHA256
```

Placeholders are deliberately not executable pins. Do not generate the expected
digest from an unreviewed manifest and call that independent verification.
There are no implicit source pins or production destinations.

The UTF-8 JSON manifest has exactly these keys:

| Key | Accepted value |
| --- | --- |
| `format_version` | Integer `1` |
| `material` | `synthetic_only` |
| `purpose` | `baseline_recovery` or `m13_acceptance` |
| `source_commit` | Exact externally supplied source SHA |
| `postgres_major` | Integer `16` |
| `schema_version` | `8` for baseline recovery; `10` for M13 acceptance |
| `runtime_grants_schema_version` | Must equal the selected schema version |
| `artifacts` | Exactly nine objects with `kind`, `path`, `sha256` |

Artifact kinds match the recorded **partial** rollback package: one `api_binary`,
one `systemd_unit`, six `systemd_drop_in`, one `caddy_config`. All paths are unique,
relative, and link-free; all digests are lowercase SHA256. Environment and TLS
recovery dependencies are separate and are not included in this nine-file proof.

Exit 0 means `READY_FOR_REVIEW`, with `runtime_acceptance=NOT_TESTED` and
`gate_acceptance=NOT_ASSESSED`. Exit 1 means `BLOCKED` with a
sanitized error class. Argument misuse exits 2 through argparse. Neither outcome
proves production recovery or grants execution permission. JSON reports contain
no bundle paths, file contents, raw I/O errors or secret values.

Checks cover the externally pinned manifest, source agreement, nine-file
composition, each artifact digest, version pairing, bounded input sizes, regular
files, current-user ownership and absence of group/other write permission.
Directory-relative opens do not follow symlink components; hardlinks are refused.
Hashing is streamed in 64-KiB chunks, bounded to 64 MiB per artifact and 64 KiB
for the manifest. Files are not executed or modified. Metadata changes during a
read are rejected. File access times may change according to filesystem policy.

**Limits:** manifest labels do not establish that bytes are synthetic or came
from the declared commit. Trusted manifest review and artifact provenance remain
operator responsibilities. This tool does not parse executable compatibility,
systemd semantics, Caddy configuration, SQL grants, actual database schema or
runtime isolation. It is not a hostile-concurrent-writer security boundary and
should run only in an owner-controlled fixture directory. A valid checksum alone
cannot close any missing dependency or acceptance gate.

## Focused validation

```sh
python3 -B -m unittest discover -s tools/recovery -p 'test_*.py' -v
```

Tests use temporary synthetic text fixtures, not archived executables or configs.
They exercise the public CLI success/failure path as well as integrity, malformed
JSON, manifest pin, version/scope, bundle composition, path escape, symlink,
hardlink, mode and size rejection. They verify input bytes/modes/mtime remain
unchanged and success cannot be reported as runtime or gate acceptance. A child
process timeout bounds CLI tests. They do not rerun earlier MeshAlot tests.

## Remaining acceptance contract: strict gate order

G17/1 remains **BLOCKED** until real production PG16 global role/member/owner,
grant/authentication dependencies, protected independent recovery material and
actual Caddy/TLS service/storage dependencies have sufficient recorded evidence.
Synthetic reconstruction, metadata-only systemd user inspection and this tool
cannot establish those facts. The existing environment supplement and matching
off-machine copy are complete; do not reopen that earlier historical blocker.

G17/2 remains **NOT STARTED**. Before execution define an approved isolated
systemd-capable target, CPU/RAM/disk/time limits, no production network route,
stage-only paths/identities/credentials, private evidence handling and cleanup.
The previous 2-CPU/4-GiB PG16 container approval is not proof of a suitable full
systemd rehearsal environment. Any use of real archives/keys needs its own scope.

| Acceptance lane | Required evidence when execution is authorized |
| --- | --- |
| Known-good baseline recovery | Archived binary and config digests/provenance; PG16 schema v8 and matching original ACLs; dependency/owner/mode and override precedence; database-backed API startup; authenticated baseline customer workflow; deliberate stage-only fault and restore; data/ledger preservation. Use the archive's verified release provenance, not an assumed current-main identity. |
| M13 acceptance | Exact approved M13 artifact; separate schema v10 fixture, owner/runtime-role separation and v10 grants; signed synthetic-agent heartbeat through actual Caddy WSS to PG16 and account dashboard; isolation from a second synthetic account; negative authentication and plaintext-backend rejection; paused provider controls, stale/offline and reconnect behavior. |
| Both lanes | Actual restored service processes and configuration, not just httptest; source/artifact/version receipts; bounded startup/recovery and resource use against pre-agreed limits; no real enrolled identities; unchanged original backups; teardown and evidence review. |

The M13 startup schema probe reads `node_status.m13_telemetry`, and the current
grants file references v9/v10 objects. Do not pair them with the v8 baseline or
apply the latest grants to that baseline. Do not require historical baseline
binaries to implement the M13 WebSocket endpoint. A staged proxy substitute must
be identified as such; it is not actual archived API M13 acceptance.

`/v1/health` alone is insufficient: the server can start in development memory
mode without a DSN, and its health handler pings PG only when a store exists.
Require startup evidence of PG mode plus authenticated, persistent DB-backed
workflow results. Never disclose the DSN or environment-file values.

The migration runner's `apply` command applies **all pending** migrations, and
`status` uses a temporary table and advisory transaction lock. Neither is a
generic production read-only metadata diagnostic. No rollback of database
history or blind migration retry is part of the stage service rollback plan.

Reuse accepted migration, ACL, synthetic globals and Caddy tests while their
relevant source/configuration/conditions remain valid. A future actual archived
service restoration tests a distinct, previously unverified requirement. No new
runtime test is authorized by this document. Keep the rejected diagnostic
paused; this tooling does not reproduce it or provide an alternative executor.
