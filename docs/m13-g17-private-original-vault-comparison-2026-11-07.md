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
