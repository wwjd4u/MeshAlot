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
