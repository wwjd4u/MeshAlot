# M13 G17 — archive key-access tooling evidence

**Local synthetic verification: PASS (13 focused tests).
MS-02 verification: pending execution receipt below.
Real backup/vault check: NOT RUN. G17/1: BLOCKED.**

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
