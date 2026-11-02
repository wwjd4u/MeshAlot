# M13 G17 — archive key-access tooling evidence

**Current local synthetic verification: PASS (18 focused tests).
Earlier pinned MS-02 revision: PASS (13 tests, zero skips).
Two real Mac PostgreSQL key attempts: BLOCKED, source unchanged.
G17/1: BLOCKED.**

## November 2, 2026 — failed Mac attempt and safe failure classification

The operator reports macOS 26.7.1, Python 3.13.15 and MacGPG2/gpgconf 2.5.21,
with the existing external encrypted APFS backup volume mounted and unlocked.
The first PostgreSQL-archive attempt used verifier source
63fc0d0366517f4c6c993cd2c1d9bf6dcf830078 with its independently pinned script hash.
The operator supplied this sanitized receipt:

    status=BLOCKED; reason=DECRYPTION_FAILED; source_unchanged=true
    scope=LOCAL_ARCHIVE_KEY_ACCESS_ONLY
    vault_source=OPERATOR_ATTESTATION_REQUIRED
    restore_acceptance=NOT_TESTED; gate_acceptance=NOT_ASSESSED

The verifier reached and passed the original ciphertext-receipt comparison,
private-agent cleanup and final unchanged-source check. GnuPG returned nonzero.
The operator confirms a private prompt appeared and the submitted value was
copied from the PostgreSQL RoboForm Safenote. That is source attestation, not
proof that the copied value was accepted. The original verifier discarded
failure details, so this receipt cannot distinguish a rejected value from a
prompt, agent or other GnuPG failure. No cause or passphrase-format correction
is inferred. Original archives, receipts and vault notes remain preserved.

The existing verifier now maps its bounded GnuPG status to fixed, sanitized
failure reasons. No new raw diagnostic output, passphrase input path, retry
loop, file extraction, service connection or permission change was added.
Actual synthetic wrong-key, cancelled-prompt and missing-pinentry cases prove
the distinction. The initial synthetic test run exposed two classification
failures: GnuPG emits BAD_PASSPHRASE after cancellation/missing pinentry too.
Specific prompt/agent statuses now take precedence over that secondary symptom.

Final focused result: **18 tests PASS, zero skips, in 7.706 seconds** on the
local Linux/GnuPG runtime. This repeats the affected verifier suite because its
failure handling changed after a specific real failure; no application,
service, archive-restore, migration, node or historical acceptance suite ran.
The earlier 13-test MS-02 receipt below applies to its original source, not this
revision. Successful real Mac key access and restore remain unverified.

Next supported action: one explicitly targeted, operator-local PostgreSQL
attempt with the reviewed updated verifier and the same existing checksum,
to obtain safe error classification. Keep the complete vault value unchanged;
do not trim unverified characters. Stop on BLOCKED and review that reason;
do not guess variants, alter permissions or change the original artifacts.
No production operation or rejected service-identity diagnostic is authorized
by this local failure-handling change.

### Second Mac attempt — generic failure persists

The operator subsequently supplied the targeted attempt using source
69a8918ac02a30b8b83a90edd72d5299cff8ea9a and its pinned verifier digest. It again
reported BLOCKED / DECRYPTION_FAILED / source_unchanged=true, with restore
NOT_TESTED and gate NOT_ASSESSED. The operator confirms the private prompt
appeared and the key was submitted. The fixed classifier found no recognized
specific failure; its generic result does not establish wrong credentials,
corrupt ciphertext or a Mac-runtime cause. Do not retry the real archive again
without evidence supporting a next change.

Next supported action: execute the existing pinned 18-test synthetic suite on
the Mac, using only disposable public fixture data and automatic synthetic
pinentry. This new runtime check is justified by two unresolved real failures;
the accepted Linux/MS-02 suites are reused, not repeated on those hosts. The
Mac launcher verifies both public source hashes, uses private temporary homes,
and may report only allowlisted status tags/numeric GnuPG error codes for
otherwise generic synthetic failures. It does not name/read a real archive,
request a vault key or contact production. Its shell/Python syntax was reviewed;
Mac execution remains NOT TESTED until the operator returns actual output.
Stop after that one synthetic run and assess its result. Preserve all originals
and all other gate, production and identity holds.

The remaining sections preserve the original tooling and runner evidence.

## Objective and permitted work

The operator is traveling and requested work requiring no copy/paste or physical
operation. Routine nonproduction testing/configuration was already approved.
The Mac backup/vault lane is deferred until the operator is parked. This change
automates its later local key-access check and validates the automation now,
using only temporary public synthetic fixtures. No real key or archive is used.
The rejected PostgreSQL diagnostic remains paused and is not reproduced.

## Implementation

- [Verifier](../tools/recovery/verify_archive_key.py) and
  [interface/limits](../tools/recovery/ARCHIVE_KEY_CHECK.md).
- [Focused tests](../tools/recovery/test_verify_archive_key.py).
- Starting M13 source: 6d9d6a9477a1d53c5b43de753d767c73eb84321d.
- Source receipt is the immutable M13 commit used by the new runner workflow;
  the actual workflow/run receipt is appended after execution.

The supported CLI requires a private interactive terminal. It opens only an
explicitly selected file, checks the independently reviewed SHA256 before
prompting, decrypts to the null device using a fresh private GnuPG home, disables
symmetric/external prompt caches and automatic key retrieval/import, bounds
status collection in memory, then rechecks source bytes/metadata and cleans up
the private agent/home. No passphrase argument, environment input or key file is
accepted. GnuPG owns the prompt; the wrapper never receives the passphrase.

Success reports only local key access. Vault source needs human attestation;
restoration and gate acceptance remain NOT TESTED / NOT ASSESSED. Exact private
paths, embedded filenames, raw diagnostics and plaintext are excluded from JSON.
The tool does not extract content, connect to PostgreSQL/SSH/production or run a
service. These boundaries are distinct from the paused service-user diagnostic.

## Verification contract and actual local result

Resources: Python standard library and GnuPG, isolated temporary synthetic files
and GnuPG homes, no production-derived material or existing user keyring.
Verifier caps: 4-GiB input, 64-KiB status in memory, 120 seconds per phase by
default; tests use tiny fixtures and shorter deadlines. Only the new suite runs.

Required outcomes: authenticated AES256 succeeds; wrong keys, tampering, wrong
pins, unsafe inputs, unsupported ciphers, invalid budgets, excess status and
concurrent source changes fail closed. The complete terminal CLI must succeed
with a synthetic prompt; headless CLI must refuse. Inputs remain unchanged
unless the test intentionally simulates another writer. Reports and cleanup
must preserve the private-material boundary.

Actual local environment: Linux, Python 3.12.14, GnuPG/gpgconf 2.4.7.
Final focused result: **13 PASS in 7.572 seconds**. Both authenticated MDC and AEAD
fixtures passed. The correct-key-then-wrong-key case confirms cache isolation
and one prompt per attempt. Timeout, terminal CLI, status bound and intentional
source mutation cases all passed.

The initial fixture setup could not bind a private Unix socket in the executor's
default sandbox. A new synthetic-only agent probe classified that local IPC
limit. Using the executor's supported network permission for local IPC allowed
the tests to run; no proxy, host permission, key, production route or denied
GitHub operation was changed. After initial success, review replaced an inherited
child file-size limit with a bounded status pipe to avoid affecting pinentry's
own file operations; the changed suite and complete CLI were verified again.

No accepted bundle, application, migration, runtime ACL, synthetic globals,
Caddy, benchmark or node suite was repeated. The stop point is verified new
tooling and its focused runner receipt, not another recovery rehearsal.

## Remaining acceptance

Actual Mac tools/mount, private human vault retrieval and the three independently
pinned Toshiba copies remain unverified by this implementation. A successful
synthetic run cannot close that proof. Real PG16 globals/ownership recovery,
remaining TLS recovery requirements and full archived service rehearsal remain
open. Preserve production, backup copies/key custody and Node001/Node002.

## Immutable MS-02 execution receipt

- Tool/test source: 63fc0d0366517f4c6c993cd2c1d9bf6dcf830078 on the existing M13 branch.
- Infrastructure configuration: 9c82f49d4406e8202f9583ec90791690e3ee40ca on the existing infrastructure branch.
- [Successful workflow 41108714209](https://github.com/wwjd4u/MeshAlot/actions/runs/41108714209), job 124245156663; provider timestamps 2026-10-30 18:39:05Z–18:39:47Z.
- Actual runtime: Python 3.14.4, GnuPG 2.4.8 on the trusted MS-02 runner.
- **13 tests PASS in 6.499 seconds; zero skips; temporary-workspace cleanup PASS.**
- Downloaded verifier SHA256: 51afc244edccf3ee081de1e6970d4ce6113c08e80e129c2a72c5686f609806f3.
- Downloaded test SHA256: c89060546be1383e8fccdb0039c7e96710881d21f9ea48f71db3525f2234397a.

The workflow verified trusted host/user/ref identity, fetched only these two
immutable public source files with independent byte-hash pins and no redirects,
then ran only the new focused suite. The child test process was restricted to
two allowed CPUs and a 1-GiB per-process address-space limit, with a 120-second
suite timeout and five-minute job cap. This is a synthetic local-process test,
not proof of a network-isolated full service stage. Source retrieval contacted
GitHub; the verifier itself used only local GnuPG IPC and synthetic files.

The workflow reported REAL_ARCHIVES=NOT_USED, PRODUCTION_CONNECTION=NONE and
GATE_ACCEPTANCE=NOT_ASSESSED. Static review confirmed no existing workflow matched
the new workflow's changed path. No old one-attempt marker, disabled workflow,
passing application/recovery suite or rejected diagnostic was reset or reused.

The runner execution was a required check of the new automation path on a
different supported runtime, not a repetition of historical recovery evidence.
No further runtime tests are needed for the documentation-only receipt update.
The tooling task is complete; real Mac/archive/vault acceptance remains NOT RUN.
