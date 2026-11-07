# Original PostgreSQL archive association

This operator-local MS-02 check addresses a specific unresolved question after
original/vault MATCH and a Mac failure with the same clarified value: identify
the ciphertext associated with the original creation/restore record so that its
digest can be compared with the Mac selection. It does not repeat decryption or
the accepted logical restore. No secret input, passphrase file access, production
connection, GPG invocation, container or child process is involved.

Selection and receipt fields come directly from immutable custody source:
https://github.com/wwjd4u/MeshAlot/blob/b00ad2faed0de865c27244b3596528a5d0723bf9/.github/workflows/ms02-m13-gate17-local-evidence-readonly.yml

Use only the hash-pinned published helper in the confirmed private MS-02 terminal:

    python3 -I -B check_pg_archive_association.py

The CLI accepts no arguments, verifies the existing host/account and refuses CI
and headless execution. It walks the fixed path through no-follow descriptors,
requires safe ancestor ownership/modes, owner-only 0700 backup/run directories,
and owner-only regular single-link 0600 input files. The original archive is
bounded below 2 GiB, checksum below 8 KiB, receipt below 64 KiB. Content hashing
has a 120-second deadline and a 125-second overall alarm. Original bytes and
fingerprinted identity/metadata are checked through completion; access times
can change from reading. No original file is written. These are owner-controlled
filesystem checks, not protection against malicious same-user races or hardware
failure. A receipt is existing local evidence, not a new independent signature.

Acceptance requires the actual encrypted archive digest to match both the exact
existing checksum/path association and the historical restore JSON fields,
including creation run ID, size, PG16, migrations 1–8 and isolated restore PASS.
Malformed or duplicate JSON fields, unsafe sources, mutation or timeout STOP.
Only fixed classifications and the ciphertext SHA256 are printed. This digest
identifies encrypted archive bytes; it is not a passphrase digest. Raw receipt
contents, paths and exceptions are never printed.

Share only ARCHIVE_ASSOCIATION=PASS and CIPHERTEXT_SHA256, or
ARCHIVE_ASSOCIATION=STOP. Stop after this one check. Do not change originals or
retry automatically. PASS enables a subsequent Mac selection comparison; it
does not prove current key access, recovery, historical prompt bytes or gate
acceptance. The check remains NOT TESTED on actual inputs until operator output.

Focused synthetic verification:

    python3 -B -m unittest discover -s tools/recovery -p 'test_check_pg_archive_association.py' -v

## Mac ciphertext comparison after original association PASS

`compare_mac_pg_ciphertext.py` pins the operator-reported original ciphertext
SHA256 f60283f0a5563a70c363cdf863c52ecce9eec410d6112411844d948e4f3a699e.
Run its published, independently hash-checked source only on the existing Mac,
with the previously verified encrypted APFS MeshAlot-Backup volume mounted and
unlocked. It selects only the PostgreSQL archive in the documented existing
external-copy folder. No file picker, key entry, decryption, sidecar rewrite,
archive transfer or production connection occurs. No other archive is checked.

It requires Darwin, a private terminal, an actual volume mount, canonical paths,
owner-only folder/file modes, a regular single-link file, unchanged descriptors
and named-file metadata, a size below 2 GiB and a 120-second hashing deadline
(125-second overall alarm). It performs no writes; reading can update atime.
Output is only MAC_POSTGRESQL_CIPHERTEXT=SAME, MISMATCH or STOP. SAME identifies
current bytes as matching the operator-reported original receipt digest; it
does not retrospectively prove which bytes the earlier Mac launcher selected
or authenticate decryption, vault retrieval or gate acceptance. On MISMATCH or
STOP, preserve all files and stop. No automatic retry or repair is permitted.

This check does not independently reverify APFS encryption or physical disk
identity; use the established operator volume evidence. Filesystem checks assume
an owner-controlled device rather than a hostile writer or compromised OS.

Focused synthetic verification:

    python3 -B -m unittest discover -s tools/recovery -p 'test_compare_mac_pg_ciphertext.py' -v
