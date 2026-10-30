# Private archive key-access verifier

This is a separate tool from the synthetic-only nine-file bundle checker.
It automates the operator-local key-access portion of G17 recovery preparation.
It does not extract an archive, restore data, inspect PostgreSQL, start services,
access an enrolled node or close a gate. CI uses synthetic fixtures only.

## Interface and acceptance

Use an operator-controlled Linux/macOS terminal with Python 3.8 or later,
GnuPG/gpgconf and private pinentry. Linux/GnuPG 2.4.7 and 2.4.8 have been tested;
actual Mac execution and the three real backup copies remain NOT TESTED. Required GnuPG
options are checked before a key prompt. Source provenance, the approved device
and the digest from the existing independently verified receipt remain operator
responsibilities. Do not upload an archive or key to GitHub or run real archives
through Actions to automate the private key-entry boundary.

    python3 -B tools/recovery/verify_archive_key.py \
      --archive /operator/private/existing-archive.gpg \
      --expected-sha256 EXISTING_REVIEWED_64_CHARACTER_SHA256 \
      --timeout 120

The example contains a deliberately non-executable digest placeholder. The key
is entered only into GnuPG's private pinentry. There is no passphrase argument,
environment-variable input or key file accepted by this tool. Standard input
must be a terminal; the public CLI refuses headless execution. Automated tests
call the verifier only with public synthetic fixtures and a synthetic pinentry.

Exit zero and PASS_KEY_ACCESS establish authenticated AES256 decryption of the
independently pinned bytes, with the original bytes and fingerprinted metadata
unchanged. They do not prove the human obtained the key from RoboForm: receipts
explicitly say OPERATOR_ATTESTATION_REQUIRED for the vault source. Restoring the
contents is NOT_TESTED and gate acceptance is NOT_ASSESSED. Failure exits one
with a redacted reason and no automatic retry. Argument misuse exits two.

## Handling and bounds

- Opens the final source component without following a symlink, requires a
  regular single-link file owned by the current user without group/other write
  access, and checks the pinned digest before launching a prompt.
- Rechecks ciphertext and metadata after decryption. Reads may update access
  time. This is for owner-controlled inputs, not a hostile concurrent-writer
  filesystem boundary or proof of backup provenance.
- Uses a new private temporary GnuPG home. It suppresses default gpg options,
  disables symmetric and external pinentry caches, disables automatic key
  retrieval/import and uses no pre-existing keyring or operator key file.
- Sends plaintext only to the operating system's null device. Captures at most
  64 KiB of GnuPG status in memory; raw status, embedded filenames and diagnostics
  are excluded from JSON. No plaintext or status-output file is created.
- Requires successful GnuPG exit plus symmetric decryption, AES256 and verified
  MDC/AEAD status. An unencrypted or unsupported-cipher input cannot pass from
  exit status alone.
- Uses a maximum 4-GiB input and 64-KiB hash chunks. The default deadline is
  120 seconds separately for each hash phase and GnuPG/pinentry phase; the CLI
  accepts 0.05–600 seconds. Hash deadlines are checked between file reads.
  Prompt time is included: retrieve the vault entry before starting the check.
- On failure/timeout/cancellation, terminates the owned crypto process and
  requests shutdown of components belonging only to the private temporary home.
  Unverified crypto-process or agent cleanup blocks acceptance. Temporary files
  are removed; original archives, original key files and vault notes are kept.

The executable is found using the operator's trusted PATH. This helper does not
authenticate the installation, prove a secure desktop or clear clipboard history.
Keep the private terminal/device within the reviewed recovery scope. Clear the
current clipboard after key entry; never paste a key into this chat or CI.

## Focused synthetic verification

    python3 -B -m unittest discover -s tools/recovery \
      -p 'test_verify_archive_key.py' -v

GnuPG and gpgconf are required; missing tools fail instead of silently skipping
the suite. Temporary fixtures use no real backup, key, account or node identity.
The public synthetic pinentry supplies only a test passphrase. Tests cover valid
MDC/AEAD access, wrong-key/cache isolation with one prompt per attempt, modified
ciphertext, wrong receipt before prompting, unsupported cipher, timeout/cleanup,
unsafe source links/modes, invalid budgets, source mutation, oversized status,
redacted receipts and both refused headless and successful terminal CLI paths.

The existing synthetic bundle checks, application, migration, ACL, Caddy and node
tests need not be repeated for this independent tool. See the dated G17 evidence
record for actual source/runner receipts and the remaining real-archive check.
The pinned MS-02 run passed all 13 tests with zero skips; no real archive or
vault key was used. [Execution evidence](../../docs/m13-g17-archive-key-verifier.md).
