# M13 G17/1 — private original/vault comparison preparation

November 7, 2026 (America/Chicago). G17/1 BLOCKED; G17/2–7 NOT STARTED.

Fresh remote M13 ref matched handoff commit a18d87fcefcc10db31606d1f52e751798c95ecda.
The initial clean checkout was the older main baseline on local branch `work`.
Fetched the existing M13 line and selected its matching local branch, without
creating another worktree, merging main, deploying or dispatching workflows.
The infrastructure ref still matched 8a108a74498dc19e111f64cbc67eac1588c9e5bd.
These are repository observations, not production measurements.

Prepared [operator-local helper](../tools/recovery/compare_original_vault.py),
[handling and acceptance contract](../tools/recovery/ORIGINAL_VAULT_COMPARISON.md)
and focused synthetic fixtures. File selection is grounded in immutable custody
source b00ad2faed0de865c27244b3596528a5d0723bf9, not a filename search.
The confirmed host/account need not be requested again. The helper enforces
identity locally before accessing the selected original.

Required outcome: exact original first-line record versus one hidden vault
entry, with unchanged original and only MATCH/MISMATCH/STOP output. Resources:
Python standard library, temporary synthetic fixtures and pseudo-terminals.
Stop after validating the helper and delivering the private local command.
Real original bytes, vault values, archives and production resources are excluded.

Local focused result: 12 tests PASS, zero skips, 0.074 seconds. Covers exact
match/mismatch, significant spaces, unterminated record, unsafe/multiline/empty
records, unsafe modes, symbolic/hard links, nonblocking FIFO rejection, original
mutation/replacement, hidden terminal entry, cancellation, timeout, overflow,
multiline paste, terminal restoration and sanitized refused cloud CLI output.
Manual source review confirmed no network, subprocess, GPG or secret output path
in the helper. No accepted synthetic GPG or recovery tests were repeated.

Tooling implemented/tested locally; operator execution NOT TESTED. Original-file
current availability and vault equality remain UNVERIFIED. PostgreSQL archive
key access remains BLOCKED from the prior bad-passphrase receipts; ciphertext
was unchanged in those receipts. Other two archive key checks remain NOT TESTED.
No original, vault note, archive, production service or held diagnostic changed.
MATCH would establish comparison only, not decryption/recovery/gate closure.

## Operator execution receipt — November 7

Operator supplied the private MS-02 terminal result: **MATCH**, using the
published source 2fdb4b7567412921dc5417240b6e733fe4997efa and its SHA256 pin
14cf22961c05c816ff25a9326e408a79bfd9bdafd2489d1e7f26d73e4d123a8b.
The displayed launch command includes successful hash verification before the
helper. This is operator-provided evidence, not remote execution by the agent.

Under the reviewed helper contract, MATCH establishes that the selected original
file passed identity/ownership/type/mode checks, its first-line record matched
the entered value, and its bytes/fingerprinted metadata remained unchanged.
Vault origin remains the operator's attestation. One optional final LF is record
framing; no other trimming or normalization occurred. No secret was supplied to
chat. Current original availability and this comparison are now PASS.

This does not prove the same bytes reached the earlier Mac GnuPG prompt or
establish authenticated archive decryption. The prior bad-passphrase receipts
stand, PostgreSQL archive key access remains BLOCKED, other two key checks remain
NOT TESTED, and G17/1 remains BLOCKED. Do not repeat the comparison or edit either
source. Next review the existing archive/receipt association and Mac key-entry
path before proposing a targeted action; no unchanged batch retry, secret
transfer, production access or held diagnostic is authorized by MATCH.

## Operator clarification — PostgreSQL Safenote label

The operator clarified that the word “key” is not part of the saved PostgreSQL
vault value. Treat it as a label, not as credential bytes. This is operator
clarification, not an instruction to strip arbitrary prefixes or edit the vault.
The prior MATCH remains accepted; do not repeat it.

Whether the last failed Mac attempt included that label or already used the
same clarified value remains pending operator confirmation. This determines
whether the failure has a concrete input correction to test. Pause additional
provenance-check development while resolving that historical input question.
No new archive attempt or production operation was performed.

## Last Mac input clarified; targeted association check prepared

The operator confirmed the last failed Mac attempt already excluded the label
and used the same clarified value that produced MATCH. This resolves the pending
historical-input question above by operator attestation; it does not independently
measure the bytes received by pinentry. Label removal is not grounds for another
unchanged decryption attempt. No variant guessing or source edits are justified.

Prepared [original ciphertext/receipt association check](../tools/recovery/PG_ARCHIVE_ASSOCIATION.md)
to obtain the original archive identity tied to the existing creation/restore
record, for comparison with the Mac selection. This targeted check addresses the
unresolved cross-device selection question after MATCH; it does not repeat the
logical restore, GPG suite or archive decryption. Resources are Python standard
library and temporary public synthetic ciphertext/receipts. Stop after one
read-only original association result; Mac comparison follows only on PASS.

Nine focused synthetic tests PASS, zero skips, 0.041 seconds. Checks cover valid
association; each wrong receipt field; wrong checksum digest/path; changed
ciphertext; unsafe modes/links; malformed/duplicate JSON; timeout; concurrent
receipt modification; and sanitized refused cloud CLI. Source review confirms
no key-file reference/access, network, GPG or production operation. Actual
original archive check NOT TESTED, Mac cross-comparison NOT TESTED. Existing
MATCH stands; G17/1 BLOCKED; other two archive-key checks NOT TESTED.

## Operator original archive association receipt and Mac next check

Operator returned ARCHIVE_ASSOCIATION=PASS and
CIPHERTEXT_SHA256=f60283f0a5563a70c363cdf863c52ecce9eec410d6112411844d948e4f3a699e.
This supersedes the original archive NOT TESTED status above. Under the published
helper contract, the original ciphertext matches its existing checksum/path and
creation/restore receipt, with source identity/metadata checks passing. This is
operator-provided evidence; no original archive or receipt entered cloud tools.
Original/vault MATCH stands. Neither receipt establishes current decryption.

Prepared a Mac-only check of the documented existing PostgreSQL external copy
against this now-recorded original ciphertext digest. Required outcome: SAME
for unchanged, safely selected ciphertext; otherwise MISMATCH or STOP. Resources:
Python standard library and public synthetic files for development. Actual check
uses only the existing encrypted APFS volume and PostgreSQL encrypted archive,
without key input, decryption, other archives or production access. Stop after
one result; do not retry the failed key-access batch.

Seven focused local synthetic tests PASS, zero skips, 0.032 seconds: same and
different ciphertext, unsafe modes/links, empty/FIFO input, timeout, concurrent
change and sanitized refused Linux CLI. Source review confirms no GPG, secret
file access, subprocess or network. Actual Mac check remains NOT TESTED. SAME
would establish current-copy identity only, not retrospectively prove the earlier
launcher selection. G17/1 BLOCKED; other archive key checks NOT TESTED.

## Operator Mac ciphertext result — SAME

Operator supplied MAC_POSTGRESQL_CIPHERTEXT=SAME from the Mac terminal using
published source 250e095e3bced10393f73edea1cb09c3e8eb906b and script hash
7cdb21aacca5a6656ce09d82bcadb14d154d5c90b47eec2aef411fe81f1a505d.
The shown command verifies the hash before invoking the helper. Under that
contract, the safely selected current Mac PostgreSQL encrypted copy matches
original receipt digest
f60283f0a5563a70c363cdf863c52ecce9eec410d6112411844d948e4f3a699e,
with unchanged source fingerprints. Current cross-device ciphertext identity:
PASS by operator-provided evidence. This supersedes Mac NOT TESTED above.
No decryption or key entry was performed by this check.

Together with original/vault MATCH and original association PASS, current
archive selection and saved-value equality are established within the recorded
contracts. The operator confirms the last failed Mac attempt already excluded
the word “key”. These observations do not measure the bytes received by the
historical GnuPG private prompt, retrospectively prove its selected input, or
establish successful key access. Do not repeat these passed checks, edit vault
or original files, guess variants, or rerun the unchanged three-archive batch.

Next narrow the Mac entry path: operator clarification of whether the last
failed prompt received a direct RoboForm paste, manual typing, or another copy
is pending. No passphrase contents, fragments or passphrase hashes are sought.
No new tool/runtime test is warranted for this documentation-only update.
G17/1 BLOCKED; PostgreSQL authenticated key access unresolved; API_ROLLBACK and
ENV_SUPPLEMENT key checks NOT TESTED; no merge/main deployment or production
access performed.

## Operator clarification and explicit single PostgreSQL retry request

The operator described the prior Mac input as copy/paste from RoboForm, resolving
the preceding entry-method question. No different source or changed credential
is established from that answer alone. The operator then explicitly requested
re-doing the Mac test now that the intended copy is understood. This supports
one deliberate PostgreSQL-only key-access attempt with fresh exact vault-value
entry, not a claim that the prior failure's cause is known or a batch rerun.

Prepared a private Mac launcher using unchanged accepted verifier source
fc040c80b40106396b618217d190df9b14a72a64, independently pinned script SHA256
4b05bd2ff0b1efa7a1a5c37b4a420affb334412824bd5aad60bdb97b871771df,
and the original-associated, Mac-matched ciphertext digest recorded above.
It checks Darwin, clipboard-tool availability, mounted/canonical volume/folder,
then invokes the core exactly once for PostgreSQL. Existing core checks require
pinned/unchanged ciphertext, authenticated AES256, successful GnuPG exit, fresh
private-agent isolation and cleanup. Decrypted output goes to the null device.
Current clipboard is cleared on launcher exit, with explicit warning if that
fails; clipboard-history/sync clearing is not claimed. No other archive is run.

Acceptance is PASS_KEY_ACCESS plus source_unchanged=true. Any other result stops
without retry, key modification, variant guessing or relaxed checks. Vault origin
remains operator attestation. Even PASS is not restore or gate acceptance.
Bash wrapper and embedded Python syntax: PASS. The accepted 21-test verifier
suite was not repeated because its code is unchanged. Actual targeted retry
NOT TESTED pending operator receipt. No archive/key/production access occurred
in preparation. G17/1 remains BLOCKED; other two key checks NOT TESTED.
The proposed additional public-fixture paste check is superseded by the user's
explicit targeted real-check request; it was not built or run.

## Targeted real Mac retry BLOCKED; tool discovery resolved

The first launch returned GNUPG_TOOLS_UNAVAILABLE, source_unchanged=NOT_ASSESSED;
no prompt opened and no key was tested. Operator metadata then established gpg
on PATH, gpgconf absent from PATH, and both executables present in the existing
MacGPG2/bin directory. A terminal-local PATH prefix yielded GNUPG_TOOLS=READY.
No install, persistent shell configuration change or new permission was needed.

The subsequent targeted PostgreSQL attempt returned operator-provided receipt:
status=BLOCKED, reason=GNUPG_REPORTED_BAD_PASSPHRASE, source_unchanged=true,
scope=LOCAL_ARCHIVE_KEY_ACCESS_ONLY, vault_source=OPERATOR_ATTESTATION_REQUIRED,
restore_acceptance=NOT_TESTED, gate_acceptance=NOT_ASSESSED. The displayed command
used accepted fixed source and independent script hash plus the associated
ciphertext digest. This supersedes targeted retry NOT TESTED above. It is not a
new missing-tool failure, proof of vault corruption or proof the backup is lost.
No further unchanged real retry, variant guesses or secret transfer is justified.

Prepared [one public synthetic clipboard/private-prompt check](../tools/recovery/MAC_PUBLIC_PASTE.md)
to test the actual human paste interface omitted by automated pinentry fixtures.
This is a newly targeted workflow check, not a rerun of the accepted 21-test
suite. It uses no real archive/key or external volume. Required outcome: known
public clipboard input authenticates the tiny known public fixture via actual
Mac pinentry, with unchanged fixture and cleanup. Stop after one result.

Five focused local tests PASS, zero skips, 1.389 seconds: pinned verifier loading,
wrong-source refusal, actual public fixture roundtrip with a separate GnuPG home,
generation-failure cleanup, and refused cloud CLI. A final review corrected the
new wrapper to block on temporary-context cleanup exceptions. Tests were rerun
for that material error-path change. No existing accepted suite was repeated.
Actual Mac public paste check NOT TESTED pending operator output. G17/1 BLOCKED;
API_ROLLBACK/ENV_SUPPLEMENT key access NOT TESTED; production untouched.

## Public Mac paste check stopped during setup

Operator returned status=BLOCKED, reason=PUBLIC_PASTE_SETUP_OR_CLEANUP_STOP,
scope=PUBLIC_SYNTHETIC_PASTE_ONLY, real_archives=NOT_USED from source
eba7bdcb2f874e29faa8bf0ebf1b7bc34f09fac8 with the pinned wrapper/verifier.
No public-prompt instruction appeared. This establishes a setup failure before
interactive verifier invocation, not a failed paste or real key attempt.
The wrapper's generic error gives no evidence for a specific root cause.

Updated only the new public wrapper with fixed stage labels (platform, CI,
terminal, tools, verifier loading, temporary directory, fixture files/encryption/
hash, clipboard, interactive core and cleanup) and a fixed failure-kind enum.
No raw exceptions, stderr, paths, values or hashes are added to receipts.
Generation failures retain the encryption stage when cleanup succeeds; failed
generation-agent cleanup blocks explicitly. No source-integrity, crypto,
original-file or accepted core-verifier checks were relaxed. No path workaround,
new installation, real retry or production operation was attempted.

Five focused wrapper checks PASS after this diagnostic change, including
assertions for encryption-failure stage retention and sanitized guard-stage
output; the accepted 21-test suite was not repeated. The next operator run is
the same single public-only workflow with new diagnostics, justified by this
specific setup failure. Stop after one receipt. Actual paste path NOT TESTED;
real PostgreSQL key access BLOCKED, ciphertext unchanged in its last receipt;
API_ROLLBACK/ENV_SUPPLEMENT NOT TESTED; G17/1 BLOCKED.

## Public fixture encryption failure isolated; short generator home fix

Operator receipt from a89f629833a056a80e55d8dad9fa6012fc307e1a:
status=BLOCKED, stage=PUBLIC_FIXTURE_ENCRYPTION,
failure_kind=CHECK_OR_LOCAL_ERROR, reason=PUBLIC_PASTE_SETUP_OR_CLEANUP_STOP,
scope=PUBLIC_SYNTHETIC_PASTE_ONLY, real_archives=NOT_USED. No paste prompt was
reached. This supersedes the earlier unknown setup stage, not real key evidence.

A targeted local public-only long-home reproduction failed in the same
PUBLIC_FIXTURE_ENCRYPTION stage. A separate public GnuPG probe showed an
agent-related diagnostic; no literal socket-length message was detected.
This supports a compatibility hypothesis, not a confirmed Mac root cause.

Changed only the public fixture generator to use a short random owner-only
/tmp GnuPG home, with agent cleanup before context removal. The fixture folder
can remain on the normal temporary filesystem. Existing user keyrings and
accepted core-verifier source/requirements are untouched. Failure diagnostics
are captured to a temporary public-only file, read with an 8-KiB bound and mapped
to fixed enums. No raw diagnostics, private paths or real secrets leave it.

Seven focused tests PASS, zero skips, 2.737 seconds. New checks cover successful
actual public fixture generation from a long working path and fixed agent-error
classification/redaction. Prior focused wrapper checks were rerun for this
material generator change; accepted 21-test verifier suite was not repeated.
Next: one corrected public-only Mac prompt/paste check, then stop. Mac result
NOT TESTED; actual clipboard-to-pinentry acceptance NOT TESTED. Real PostgreSQL
key-access remains BLOCKED with ciphertext unchanged in its last receipt;
other archive key checks NOT TESTED; G17/1 BLOCKED. No production activity.

## Actual Mac public-paste workflow failed with known public fixture

Operator output from pinned source fa5e62efe5950b152d09b031289d9da567dc7351
and wrapper SHA256 cf79544ee7f523d90e6e8360f5e995a32f362278c9a3e5b337d9a8557a947a49
shows the public-prompt instruction followed by status=BLOCKED,
reason=GNUPG_REPORTED_BAD_PASSPHRASE, source_unchanged=true,
scope=PUBLIC_SYNTHETIC_PASTE_ONLY, real_archives=NOT_USED.
Public fixture generation now completed; short-home setup issue no longer
blocks this run. It does not retrospectively prove the old Mac setup cause.

Known public fixture generation, accepted verifier, and actual Mac clipboard/
private-prompt workflow can produce the same bad-passphrase result without
any real vault value or archive. This fails the public paste workflow but does
not alone establish that clipboard handling is the cause or identify the bytes
received by pinentry. It is not evidence that the original backup/key is lost.
Real key/backup checks remain paused; all originals and prior evidence preserved.

Next targeted check: reuse only this tiny public fixture workflow once, changing
only human entry to manual typing of its documented known public ASCII value.
Do not paste or enter a real credential. This tests a distinct input method
against the specifically failed manual-paste workflow; accepted automated suites
and real archive checks need no repetition. The wrapper's on-screen paste
instruction is superseded for that single typed-entry diagnostic. Required
outcome: PASS_KEY_ACCESS with source_unchanged=true, or a redacted failure and
stop. A typed success would support narrowing to clipboard/paste selection;
a typed failure would leave wider actual-prompt/fixture compatibility unresolved.
It would not justify real retries or gate closure by itself.

No implementation change was needed for the manual-entry diagnostic, so no new
local test suite was run. Actual public typed entry NOT TESTED; original/vault
MATCH and ciphertext SAME stand. Real PostgreSQL key access BLOCKED; other two
archive key checks NOT TESTED; G17/1 BLOCKED. No production access/deployment.

## Public typed-entry PASS; one targeted real typed-entry check justified

Operator supplied the accepted fixed-verifier public receipt:
status=PASS_KEY_ACCESS, reason=PINNED_AES256_ARCHIVE_AUTHENTICATED,
source_unchanged=true, scope=PUBLIC_SYNTHETIC_PASTE_ONLY,
real_archives=NOT_USED. Source remained fa5e62efe5950b152d09b031289d9da567dc7351
with the previously pinned wrapper/verifier hashes. The operator explicitly
confirmed manual typing of the documented public value for this successful run.
Public typed-entry acceptance is PASS. Prompt window versus terminal presentation
was not supplied; it is not needed to reuse the successful manual entry method.

The earlier public paste attempt failed and a subsequent manually typed public
attempt passed. These distinct generated fixtures use the same known public
value and wrapper path; this supports a targeted entry-method investigation,
not proof of the exact clipboard failure or byte-level historical input. Do not
repeat either passed public checks or the original/vault/ciphertext comparisons.

Next justified action: one PostgreSQL-only invocation of unchanged accepted
verifier fc040c80b40106396b618217d190df9b14a72a64 with the existing source-hash
pin and original-associated/Mac-matched ciphertext digest. Manually type the
exact saved PostgreSQL value at GnuPG's private prompt, without clipboard paste,
variants, normalization or changes to originals. The operator's clarification
that “key” is a label remains applicable. No private value is sought in chat.
Reuse successful terminal-local MacGPG2 PATH setup; check the existing mount
and canonical folder as before. Clear current clipboard on launcher exit; do
not claim clipboard-history/sync audit. Decrypted output is discarded locally.

Required outcome: PASS_KEY_ACCESS and source_unchanged=true. On anything else,
stop without another attempt. Do not advance to the API/environment archives
within this invocation, restore anything, contact production or close G17/1.
Vault source still requires operator attestation; even PASS is key access only.
Accepted verifier code/tests are unchanged, so no local suite is rerun for this
entry-method-only command. Actual PostgreSQL typed-entry check NOT TESTED;
G17/1 BLOCKED; API/environment archive key checks NOT TESTED.

## Real PostgreSQL manually typed attempt BLOCKED

Operator supplied the requested typed-entry command's receipt:
status=BLOCKED, reason=GNUPG_REPORTED_BAD_PASSPHRASE, source_unchanged=true,
scope=LOCAL_ARCHIVE_KEY_ACCESS_ONLY, vault_source=OPERATOR_ATTESTATION_REQUIRED,
restore_acceptance=NOT_TESTED, gate_acceptance=NOT_ASSESSED. The displayed command
pins accepted verifier source/hash, the associated original ciphertext digest,
MacGPG2 PATH, existing mount/canonical folder and a 120-second limit.
This records the requested manual-entry attempt by operator output; no separate
byte-level measurement or new vault-source attestation is inferred.

Real archive typed-entry result is BLOCKED, superseding NOT TESTED above.
The successful known-public typed-entry test stands. Paste handling alone cannot
account for the real failure. Original/vault MATCH and current ciphertext SAME
also stand within their contracts. No more real attempts or passphrase variants
are justified by these results; preserve originals, vault entries and receipts.

Read-only repository review of the backup-readiness record (including original
commit 7de98be) and immutable custody workflow confirms: initial decryption/TOC
and isolated logical restore are historical operator-reported PASS; later
workflow validates checksum/receipt and key permissions only, explicitly never
reads the key. The reviewed repository records do not contain the actual original
PostgreSQL encryption/key-generation command. Current file naming/custody and
vault equality are not a fresh cryptographic test that this preserved original
key file decrypts these archive bytes. That relationship remains unresolved.
This is a documented evidence limit, not proof of a wrong key or lost backup.

Next locate the original creation command/script in prior chat/archive or private
MS-02 records and review its input method, without sharing secrets/raw history,
reading keys into cloud tools, decrypting again or bypassing existing holds.
Asked only where that source is available. No new test/tool was built or run;
no accepted suites repeated. G17/1 BLOCKED; API_ROLLBACK/ENV_SUPPLEMENT NOT TESTED;
production, key files, archives and vault entries not changed by this review.

## Corrections, timeout adjustment, and actual PostgreSQL key-access PASS

The operator clarified the earlier purported real typed attempt was actually
pasted. Supersede any attribution of that real failure to manual typing; real
manual-entry acceptance was NOT TESTED. A later manual-entry attempt reportedly
timed out before completion; no receipt was supplied for that timeout and no
key-validity conclusion follows. The operator then attempted the standalone
option `--timeout 600`, received command-not-found, and was given the full command.
The existing supported core deadline was increased to 600 seconds, without
source changes, relaxed authentication or a repeated synthetic suite.

Operator supplied the full 600-second launch and receipt:
status=PASS_KEY_ACCESS, reason=PINNED_AES256_ARCHIVE_AUTHENTICATED,
source_unchanged=true, scope=LOCAL_ARCHIVE_KEY_ACCESS_ONLY,
vault_source=OPERATOR_ATTESTATION_REQUIRED, restore_acceptance=NOT_TESTED,
gate_acceptance=NOT_ASSESSED. Source fc040c80b40106396b618217d190df9b14a72a64,
script SHA256 4b05bd2ff0b1efa7a1a5c37b4a420affb334412824bd5aad60bdb97b871771df,
and ciphertext digest f60283f0a5563a70c363cdf863c52ecce9eec410d6112411844d948e4f3a699e
were pinned in the displayed launch command. PostgreSQL key access is now PASS
by operator-provided evidence. This supersedes the earlier BLOCKED status for
current key access; historical failures remain preserved.

The operator explicitly reported copying to a text page and then pasting into
the passphrase prompt. This is NOT a real manual-typing PASS. No exact reason
for differences between earlier and successful pastes is established. The
original/vault MATCH and associated ciphertext SAME evidence stands. The receipt
still requires human vault-origin attestation; intermediary text-page handling
was not audited, and no claim about storage/sync/history cleanup is made.
User was advised to keep it private and close without saving if temporary and
unsaved. No real passphrase or archive contents entered this repository/chat.

Next: only API_ROLLBACK then ENV_SUPPLEMENT, using their shared existing API
rollback vault entry, never the PostgreSQL value. [One-asset wrapper and contract](../tools/recovery/REMAINING_ARCHIVE_KEYS.md)
reuse the accepted verifier, accepted original sidecars, existing mounted volume
and 600-second limit. No PostgreSQL selector, automatic retry or cross-asset
loop exists. Each result is returned separately, stop on failure. New focused
wrapper tests verify both asset/receipt selections, receipt formats and wrong
names, malformed/multiline receipts, unsafe modes, mutation, no retry and
PostgreSQL exclusion. No passed original/public/GPG suite was repeated.
Actual API_ROLLBACK and ENV_SUPPLEMENT key checks remain NOT TESTED. PostgreSQL
PASS is key access only, not restore/gate acceptance. G17/1 remains BLOCKED.

Remaining-wrapper local result: eight focused tests PASS, zero skips, 0.006
seconds. Manual source review confirmed one selected core invocation, fixed
whitelisted asset names, immutable verifier hash check, no key input handling,
no network in the helper, and unchanged existing receipt validation before/after
crypto. Only public helper/source retrieval uses network in the delivered
launcher. No real API/environment archive accessed during preparation.

## API rollback key-access PASS; only environment supplement remains

Operator supplied API_ROLLBACK receipt from wrapper source
9c00e10f21e222597618d36314b4b4292b4eae9d with its independent hash and accepted
fixed-verifier source/hash pinned in the displayed command:
asset=API_ROLLBACK, status=PASS_KEY_ACCESS,
reason=PINNED_AES256_ARCHIVE_AUTHENTICATED, source_unchanged=true,
scope=LOCAL_ARCHIVE_KEY_ACCESS_ONLY, vault_source=OPERATOR_ATTESTATION_REQUIRED,
restore_acceptance=NOT_TESTED, gate_acceptance=NOT_ASSESSED.
API_ROLLBACK key access is PASS by operator-provided evidence, superseding
NOT TESTED above. The wrapper checked the existing associated sidecar, not a
newly computed expected digest. No archive was extracted or restored.
The operator was instructed to use the existing API rollback RoboForm entry;
vault origin remains operator attestation rather than machine verification.

PostgreSQL PASS stands and neither passed archive is repeated. Next and only
remaining key-access check is ENV_SUPPLEMENT with the same existing API rollback
vault value, same pinned wrapper/verifier, its own existing sidecar, and the
600-second deadline. Exactly one invocation, stop after its result. No code
change or local test repetition is needed for the already-tested asset selector.
Actual ENV_SUPPLEMENT key check NOT TESTED. G17/1 remains BLOCKED by remaining
key/recovery proof; restore acceptance NOT TESTED, gate acceptance NOT_ASSESSED.
No production, key-source or archive modification occurred in recording this.

## Environment supplement PASS — all three key-access receipts accepted

Operator supplied ENV_SUPPLEMENT receipt from pinned wrapper source
9c00e10f21e222597618d36314b4b4292b4eae9d and unchanged accepted verifier:
asset=ENV_SUPPLEMENT, status=PASS_KEY_ACCESS,
reason=PINNED_AES256_ARCHIVE_AUTHENTICATED, source_unchanged=true,
scope=LOCAL_ARCHIVE_KEY_ACCESS_ONLY, vault_source=OPERATOR_ATTESTATION_REQUIRED,
restore_acceptance=NOT_TESTED, gate_acceptance=NOT_ASSESSED.
This supersedes ENV_SUPPLEMENT NOT TESTED above. Its own existing associated
sidecar was used as expected digest provenance. Source stayed unchanged and
no extraction or restore occurred under this discard-only verifier.

Current archive key-access matrix (operator-provided receipts):
- PostgreSQL: PASS; source unchanged.
- API_ROLLBACK: PASS; source unchanged.
- ENV_SUPPLEMENT: PASS; source unchanged.

All three key-access checks are complete. Do not repeat them or any accepted
synthetic/recovery test. The tool cannot verify human vault retrieval; requested
explicit operator attestation that both successful values came from the two
existing respective RoboForm Safenotes without saved-value edits. Vault-source
attestation remains pending until that answer; no secret contents are requested.

Read-only review of the real PG16 globals runbook, TLS recovery strategy and
October 29 Caddy metadata checkpoint confirms remaining G17/1 blockers:
actual PG16 global roles/memberships/authentication/ownership/ACL recovery and
protected globals material; target-specific TLS reissuance prerequisites and
protected fallback state; full archived API/systemd/Caddy rehearsal with correct
historical v8 service/database grants and separate approved real-artifact scope.
Synthetic composition/ACL/globals PASS does not close these. Prior logical DB
restore stands; do not repeat it. G17/2–7 NOT STARTED; M13 NOT RELEASE-READY.

The rejected PostgreSQL service diagnostic/upload remains paused. No production
route, SSH renewal, workflow dispatch, archive extraction, certificate export,
permission change, or retry of that operation is inferred from key-access PASS.
This is documentation-only receipt reconciliation; no tests were rerun. Next
technical scope is real recovery dependencies, starting with the existing PG
metadata/access hold, not another key test or a production deployment.

## Vault-origin attestation not obtained

In response to the explicit question whether both successful values came from
their respective existing RoboForm Safenotes without saved-value edits, the
operator selected: “No—one or both came from another source.” Do not claim
independent retrieval from the intended vault entries or close the vault-recovery
procedure. The three authenticated key-access receipts remain PASS; ciphertext
was unchanged in each. A different reported origin does not invalidate the
cryptographic receipts, and does not identify an incorrect saved vault value.

Requested a nonsecret clarification: which value(s) and source type only,
including whether a temporary text page was copied from the intended RoboForm
entry. No values, character fragments, key hashes or private paths are requested.
This distinguishes an intermediary from a genuinely different original source;
no key edits, archive reruns or new handling are authorized by the answer alone.
Do not infer an original-key-file read, RoboForm failure or note corruption.

Vault-source recovery status: UNVERIFIED / ATTESTATION NOT OBTAINED, superseding
any implication that only a routine affirmative confirmation remains. Archive
key-access matrix: all three PASS. G17/1 remains BLOCKED; other recovery holds
stand. Documentation-only update; no tests or actual archives accessed.

## Continue-build instruction

The operator clarified both successful values were copied into a text document
and then pasted into GnuPG. This identifies intermediaries, but does not confirm
the original intended RoboForm note source or absence of edits. The operator
then requested continuing the build. Preserve UNVERIFIED vault origin without
repeating the attestation question or treating it as a reason to rerun crypto.
All three authenticated archive key-access receipts remain PASS.

Read-only review resumed at real recovery dependencies and preserved the exact
held diagnostic. [Next-build checkpoint](m13-g17-next-build-checkpoint-2026-11-07.md)
records fresh refs, remaining requirements and new-to-handoff support escalation
evidence. No executable workaround, production operation or repeated test was
performed. G17/1 BLOCKED; G17/2–7 NOT STARTED; M13 NOT RELEASE-READY.
