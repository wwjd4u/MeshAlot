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
