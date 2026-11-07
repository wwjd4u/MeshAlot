# M13 G17 — archive key-access tooling evidence

**Current archive key access: ALL THREE PASS** — PostgreSQL, API_ROLLBACK and
ENV_SUPPLEMENT each authenticated their pinned AES256 ciphertext with
source_unchanged=true. Do not repeat these checks. Explicit human vault-source
attestation is pending; the receipts cannot prove retrieval from RoboForm.
Restore NOT_TESTED; gate NOT_ASSESSED. G17/1 remains BLOCKED by real PG globals/
ownership/ACL recovery, TLS recovery dependencies and archived service rehearsal.
The rejected diagnostic hold remains. See [the final key-access receipts and remaining scope](m13-g17-private-original-vault-comparison-2026-11-07.md).

**Latest key-access status:** PostgreSQL PASS; API_ROLLBACK PASS, each authenticated
AES256 with source_unchanged=true. Only ENV_SUPPLEMENT remains NOT TESTED; use
its existing sidecar and the same API rollback vault value. Do not repeat either
passed archive. Vault origin remains operator attestation. Restore NOT_TESTED;
gate NOT_ASSESSED; G17/1 BLOCKED. See [the cumulative receipts](m13-g17-private-original-vault-comparison-2026-11-07.md).

**Latest operator result: PostgreSQL key access PASS**, authenticated pinned
AES256 archive, source_unchanged=true, accepted fixed verifier, 600-second limit.
Successful entry was pasted through a text page, not manually typed. Earlier
real manual-entry failure attributions were corrected by the operator. Exact
paste-failure cause remains unknown. Do not repeat PostgreSQL or passed tests.
Next only API_ROLLBACK then ENV_SUPPLEMENT with the existing API rollback vault
value; both remain NOT TESTED. Restore NOT_TESTED; gate NOT_ASSESSED; G17/1 BLOCKED.
See [the cumulative receipt and corrections](m13-g17-private-original-vault-comparison-2026-11-07.md).

**Latest real PostgreSQL result:** requested manual-entry attempt again returned
GNUPG_REPORTED_BAD_PASSPHRASE, source_unchanged=true. Public typed-entry PASS,
original/vault MATCH and ciphertext SAME stand. Pause further real retries.
Next locate/review the original encryption command: repository custody evidence
checks key permissions but deliberately does not test its contents against the
archive. That relationship remains unresolved; backup loss is not established.
G17/1 BLOCKED. See [the cumulative receipt](m13-g17-private-original-vault-comparison-2026-11-07.md).

**Latest public Mac result:** manual typing PASS_KEY_ACCESS with unchanged
synthetic ciphertext; operator explicitly confirmed typing. The earlier paste
attempt failed. One targeted PostgreSQL-only check using manual entry is now
justified by this new evidence; no clipboard paste, variant guesses, batch
rerun or gate closure. Actual typed real-key check NOT TESTED. See [the cumulative receipt](m13-g17-private-original-vault-comparison-2026-11-07.md).

**Newest Mac diagnostic:** known public synthetic fixture reached the manual
paste workflow and returned GNUPG_REPORTED_BAD_PASSPHRASE with source unchanged,
real_archives=NOT_USED. Real backup/vault inputs were absent. Next is one manually
typed public-value comparison of entry methods; real retries remain paused.
No root cause or gate closure is established. See [the cumulative receipt](m13-g17-private-original-vault-comparison-2026-11-07.md).

**Latest targeted Mac retry:** after fixing terminal-local gpgconf PATH discovery,
PostgreSQL again returned GNUPG_REPORTED_BAD_PASSPHRASE, source_unchanged=true.
Original/vault MATCH and ciphertext SAME stand. Further real retries are paused;
next is one public synthetic paste-to-actual-pinentry workflow check, without
real keys/archives. G17/1 BLOCKED. See [the cumulative receipt](m13-g17-private-original-vault-comparison-2026-11-07.md).

**Latest operator evidence:** original/vault MATCH, original archive/restore-receipt
association PASS, and current Mac ciphertext SAME as that original. These are
comparison/provenance results, not authenticated decryption. Last failed Mac
entry already excluded the label “key”. G17/1 remains BLOCKED. Preserve passing
checks; next clarify the Mac private-prompt entry path, without secret output
or an unchanged batch retry. See [the cumulative receipt](m13-g17-private-original-vault-comparison-2026-11-07.md).

**November 7 update:** operator-local original PostgreSQL key versus vault entry
comparison returned **MATCH**, using pinned source and hash. Original file checks
and unchanged-source checks passed under the helper contract. This supersedes
historical statements below that original availability/equality are unverified.
Archive key access remains BLOCKED; API/environment checks NOT TESTED; G17/1
BLOCKED. See [the comparison receipt and limits](m13-g17-private-original-vault-comparison-2026-11-07.md).
Do not repeat the comparison or unchanged archive batch.

**Current local synthetic verification: PASS (21 focused tests).
Earlier pinned MS-02 revision: PASS (13 tests, zero skips).
Original Mac synthetic verification: FAIL (13 pass, 5 fail).
Updated Mac verification: PASS (21 tests, zero skips, 13.196 seconds).
Latest fixed-source PostgreSQL attempt: BLOCKED / GNUPG_REPORTED_BAD_PASSPHRASE,
source unchanged. API and environment archive key checks: NOT TESTED.
G17/1: BLOCKED.**

## November 5, 2026 — fixed-source PostgreSQL key attempt blocked

The operator supplied the three-asset launcher's actual output:

    VERIFIER_SOURCE=fc040c80b40106396b618217d190df9b14a72a64
    ASSET=POSTGRESQL ROBOFORM_NOTE=POSTGRESQL
    status=BLOCKED; reason=GNUPG_REPORTED_BAD_PASSPHRASE
    source_unchanged=true; vault_source=OPERATOR_ATTESTATION_REQUIRED
    scope=LOCAL_ARCHIVE_KEY_ACCESS_ONLY
    restore_acceptance=NOT_TESTED; gate_acceptance=NOT_ASSESSED

The corrected, Mac-verified tool reached the pinned archive and GnuPG reported
bad passphrase. The encrypted source retained its checksum and fingerprinted
metadata. The batch stopped before API_ROLLBACK and ENV_SUPPLEMENT, so those
two key checks are NOT TESTED. This is distinct from the older output-file
compatibility failure. It establishes rejection of the submitted value for
this input; it does not establish the cause of a vault/source mismatch or that
the original encrypted backup is unrecoverable. Latest vault-source attestation
remains explicit in the receipt. No archive or vault entry should be changed,
no character prefix should be guessed/trimmed, and no failed batch rerun.

The original [MS-02 metadata-only custody validation](https://github.com/wwjd4u/MeshAlot/blob/b00ad2faed0de865c27244b3596528a5d0723bf9/.github/workflows/ms02-m13-gate17-local-evidence-readonly.yml)
ties the original PostgreSQL archive, checksum, prior isolated-restore receipt
and separate protected passphrase file to the same backup creation record.
It verified owner/type/0600 mode without reading the passphrase. The October 12
custody record says the original passphrase files were preserved when copies
were put in separate RoboForm Safenotes. Current original-file availability
and equality to the submitted vault value remain unverified.

Next supported action: identify the operator's existing original MS-02 Linux
account/session, then check the original file's identity and compare the two
values privately on that device, read-only, without printing/exporting either
key, changing originals or decrypting the real archive again. Do not create a
GitHub workflow/secret or send key material to chat/Drive/CI for that comparison.
Existing original-account access is a resource prerequisite, not a request for
production authorization or new SSH privileges. The accepted 21-test Mac and
Linux receipts are reused. G17/1 stays BLOCKED; other recovery holds remain.

## November 5, 2026 — fixed verifier accepted on the Mac

Operator-provided output identifies source
**fc040c80b40106396b618217d190df9b14a72a64**, passes both pinned public source
hashes, and reports **21 tests PASS, zero skips, in 13.196 seconds** with
MAC_FIXED_VERIFIER_RESULT=PASS. It includes the complete interactive CLI,
authenticated MDC/AEAD access, wrong-key/cache isolation, cancellation,
tampering, timeout, status limit, source preservation, plaintext/status
separation, nonzero-exit rejection and failed-launch descriptor cleanup.

- Verifier SHA256: 4b05bd2ff0b1efa7a1a5c37b4a420affb334412824bd5aad60bdb97b871771df.
- Test SHA256: 189d0e079d27ba92bac0453554befb11af260b178d2bc528c9d712ad824368c0.

The earlier additional single-case result still referenced old source
69a8918ac02a30b8b83a90edd72d5299cff8ea9a and repeated its /dev/null.part
failure in 2.528 seconds. Preserve that historical receipt; it does not test
or invalidate the corrected source's later successful Mac result.

Mac compatibility is now verified for this tooling. No synthetic suite needs
to run again without a material change or failed result. Real vault-key access
and actual restores remain unverified. The two earlier real PostgreSQL attempts
remain BLOCKED historical evidence with source unchanged; they are not passes.

Next supported action: the already approved local, discard-only key-access
checks against the three existing checksum-pinned archive copies. A reviewed
Mac launcher selects each distinct asset once, in order: PostgreSQL, partial
API/systemd/Caddy, then environment supplement. It uses the PostgreSQL Safenote
first and the existing API-rollback Safenote for the latter two. Each invocation
uses the proven fixed source and its own fresh private GnuPG home; shell failure
stops the batch before another asset. There is no failed-attempt retry loop.
The current clipboard is cleared on exit, with a warning if that cannot be
confirmed. Original archives, checksum receipts and vault entries are retained.

The launcher uses the existing independently verified sidecars as expected
digest receipts, verifies their association with the selected archive, and
checks the mounted volume/canonical containment before invoking the core.
Shell/Python syntax review: PASS; actual three-asset batch: NOT EXECUTED.
No plaintext output, production operation or rejected diagnostic is introduced.
After first failure, preserve any earlier PASS and diagnose only what failed;
do not restart the batch or repeat passed assets. After all three pass, require
operator attestation of the specified RoboForm sources. G17/1 remains BLOCKED
on the remaining actual recovery evidence; this is not a restore or deployment.

## November 4, 2026 — output-file failure identified and routing corrected

The operator returned the pending single-case synthetic check: **1 FAIL in
2.237 seconds**, MAC_FOCUSED_RESULT=FAIL, with untruncated public-fixture stderr:

    gpg: error creating '/dev/null.part': Operation not permitted
    gpg: handle plaintext failed: Operation not permitted

This identifies a compatibility failure in the verifier's output configuration:
it supplied /dev/null as a GnuPG output filename, and MacGPG attempted a sibling
temporary output file. Successful integrity status alone could not establish
success because handling the decrypted output failed. This finding uses only
the correct public fixture key; acceptance of the actual vault key remains
unverified. No real backup was retried or original archive/key changed.

The fix requests --output - and redirects the child process's stdout to the
OS null device. A separate inherited pipe now carries GnuPG status; plaintext
does not enter Python's captured status, a disk file or shared output. Existing
64-KiB status/deadline bounds, pinned source checks, authenticated AES256,
nonzero-exit rejection and private-agent cleanup remain mandatory. Pipe
descriptors close on normal completion, timeout and child-launch failure.
No permissions were broadened and no GnuPG integrity/exit error is ignored.

The affected verifier suite passed **21 tests, zero skips, in 7.994 seconds**
on Linux / Python 3.12.14 / GnuPG 2.4.7. New process-boundary tests exercise
discarding plaintext larger than the status cap, a separate status descriptor,
good integrity statuses with nonzero exit, and descriptor cleanup on failed
launch. Existing crypto/CLI, cache isolation, cancellation, tamper, timeout,
status-limit and unsafe-source checks pass with the changed routing. Only
this materially changed verifier suite ran; other accepted suites were reused.

Next supported action: one pinned public-fixture run of the updated 21-test
suite on the Mac. Rechecking the affected cases on that runtime is required
because output/status routing changed after a reproduced Mac failure. Do not
retry real backup keys until this Mac verification passes. G17/1 remains
BLOCKED; restored data and production acceptance are NOT TESTED. No runner
dispatch, production connection, release or node operation was performed.

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

### Actual Mac synthetic result — failure reproduced with a correct dummy key

Operator-provided Mac output for the pinned source above reports **18 tests in
11.289 seconds: 13 PASS, 5 FAIL**, MAC_SYNTHETIC_RESULT=FAIL. Failures are valid
MDC-key access, AEAD-key access, the complete terminal CLI, the AES128 rejection
case (it stopped too early), and the cache-isolation case (its initial correct
key stopped too early). The supplied automatic pinentry uses only public
synthetic passphrases. No real archive or vault key was involved in this run.

For known-correct synthetic decryptions, the captured allowlisted status includes
NEED_PASSPHRASE_SYM, PINENTRY_LAUNCHED, BEGIN_DECRYPTION, DECRYPTION_INFO,
DECRYPTION_OKAY, GOODMDC, END_DECRYPTION and FAILURE; the parsed GnuPG error code
is 1 (general error). The same failure occurs independently of the private
backup and vault. This establishes a Mac runtime/verifier-path problem, not its
underlying cause, and does not establish acceptance of the actual vault key.

Keep GnuPG zero exit and authenticated AES256 mandatory. Do not accept successful
status tags while ignoring a nonzero exit, change credentials or retry a real
backup. The next supported step is one repeat of the specifically failed valid
synthetic-key test with temporary stderr capture, to identify GnuPG's error.
The reviewed launcher pins the same two public files, runs only that failed
case, and prints at most 16 KiB of public-fixture diagnostics after replacing
the temporary GnuPG-home and user-home paths. Its temporary diagnostic file
contains synthetic-run metadata only and is removed automatically. Other 13
passing Mac checks and accepted Linux/MS-02 checks are not repeated.
Launcher shell/Python syntax: PASS; actual focused Mac result: NOT TESTED.
All real recovery, production, rejected diagnostic and identity holds remain.

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
