# Remaining archive key checks — one asset per invocation

PostgreSQL key access has operator-provided PASS. Do not repeat it. The remaining
approved checks are API_ROLLBACK and ENV_SUPPLEMENT, both with the existing
API-rollback RoboForm value. Never use the PostgreSQL value for these assets.

Run hash-verified `check_remaining_archive.py API_ROLLBACK` beside the accepted
hash-pinned `verify_archive_key.py` source, on the existing Mac with MacGPG2/bin
on PATH and the established encrypted APFS volume mounted/unlocked. Stop after
one receipt. ENV_SUPPLEMENT is a separate later invocation after API PASS.
No option can select PostgreSQL and no loop retries or advances to another asset.

The wrapper uses the existing independently transferred `.sha256` sidecar for
the exact chosen archive. It never computes a new expected digest from an
unreviewed archive. The original receipt's GNU text/binary format, basename,
bounded single-line content, regular single-link owner-only 0600 file identity
and unchanged bytes/metadata are verified. The mounted volume, canonical path,
owner-only 0700 folder and unchanged directory identity are also required.
This is association with an existing accepted receipt, not a new independent
proof of provenance. APFS encryption/physical UUID are established by prior
operator evidence, not independently reprobed by this wrapper.

The accepted core independently verifies pinned and unchanged ciphertext,
authenticated AES256, zero GnuPG exit, fresh private-agent/cache isolation and
cleanup, and discards decrypted output locally. Deadline is 600 seconds per
core phase, including input. Passphrase input belongs only in GnuPG's private
prompt. The wrapper never receives it. No extraction, restore, production,
account/node access, secret logging or new backup/key is involved.

Use the exact existing API rollback saved value, with no label, trimming,
normalization or variants. The successful PostgreSQL route was operator-reported
paste through a text page, not manual typing; its exact transformation was not
measured. Keep any temporary text surface private and unsaved, and do not put
values in chat, shell arguments, files, repository, CI or Drive. Current clipboard
is cleared on exit; cleanup failure blocks acceptance. No history/sync audit or
reliable memory erasure is claimed. Existing key sources remain unchanged.

PASS requires status=PASS_KEY_ACCESS and source_unchanged=true. Vault retrieval
remains operator attestation. Any other result stops; do not rerun a passed
asset. Restore NOT_TESTED and gate NOT_ASSESSED apply even after both pass.

Focused new wrapper checks:

    python3 -B -m unittest discover -s tools/recovery -p 'test_check_remaining_archive.py' -v
