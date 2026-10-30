# M13 G17 — independent vault recovery-key check

Prepared: October 29, 2026 (America/Chicago).

**LOCAL TOOL IMPLEMENTED / SYNTHETIC TESTS PASS;
REAL ARCHIVE/VAULT CHECK NOT EXECUTED.**

**Implementation checkpoint:** [the local interactive verifier](../tools/recovery/ARCHIVE_KEY_CHECK.md)
now automates digest checks, authenticated discard-only decryption, status
redaction, timeout and private-agent cleanup. Its focused tests use synthetic
material only. This is implementation/verification of tooling, not execution
against the actual Toshiba copies or a completed vault recovery check. The Mac
portion is deferred while the operator is traveling; no copy/paste or key entry
is requested during travel.

## Required outcome and scope

Prove that the operator can retrieve the existing recovery passphrases from
RoboForm and authenticate/decrypt the existing, digest-pinned encrypted backup
copies without the original MS-02 passphrase files. The three approved assets
are the PostgreSQL logical archive, nine-member API/systemd/Caddy archive and
environment supplement. The latter two use the existing API-rollback passphrase.

This check discards decrypted output locally. It does not extract files, restore
a database, start services, inspect private environment values, create new
backups/keys, read TLS keys, contact production or reproduce the rejected
PostgreSQL diagnostic. It cannot establish role/ACL recovery or G17/2 acceptance.

The operator has approved continued testing/configuration. No further generic
permission to continue is needed. Exact device/tool/artifact identification is
still required; stop on unexpected material, changed digests or unsafe handling.

## Device and resource contract

- Prefer the existing Mac and encrypted APFS MeshAlot-Backup volume. This tests
  recovery independently of MS-02 and uses native encrypted-APFS support.
- The operator subsequently reported the drive connected to the Mac. Successful
  unlock, actual mount, native tool availability and the local artifact locations
  are not yet verified. This information does not require another general
  permission to continue.
- Use only the existing encrypted backup volume. Leave the TOSHIBA EXT Time
  Machine volume and all existing files intact. The APFS volumes share one
  physical disk; this check does not create another failure domain.
- Use a private local terminal and existing GnuPG installation. Check support
  for disabling the symmetric passphrase cache before attempting decryption.
  Do not assume that GnuPG or a particular pinentry program is installed.
- One archive at a time, streaming to the null device, no plaintext output file,
  network service, container, cloud resource or production session. Bound the
  selected input by its previously recorded identity/digest and observed size.
  Define an operator timeout before execution; do not loop failed attempts.

## Mac metadata batch

Run this once in the operator's private Mac Terminal. It prints the OS/GnuPG
version, cache-option support and selected volume fields. It does not decrypt an
archive or request a passphrase. Share only these metadata results; do not paste
private artifact paths or keys. Missing GnuPG, unexpected volume name/filesystem,
locked encryption or missing mount evidence stops this lane for configuration.
Run it only when parked; no terminal operation is requested during travel.

    (
      [ "$(uname -s)" = "Darwin" ] || { printf 'STOP_NOT_MAC\n'; exit 1; }
      printf 'MACOS_VERSION='
      sw_vers -productVersion
      if command -v python3 >/dev/null 2>&1; then
        python3 --version 2>/dev/null
      else
        printf 'PYTHON3_AVAILABLE=False\n'
      fi
      if command -v gpg >/dev/null 2>&1; then
        printf 'GPG_AVAILABLE=True\n'
        gpg --no-options --version 2>/dev/null | head -n 1
        gpg --no-options --dump-options 2>/dev/null |
          awk '$0 == "--no-symkey-cache" {found=1} END {print "GPG_FRESH_SYMMETRIC_KEY_SUPPORTED=" (found ? "True" : "UNVERIFIED")}'
      else
        printf 'GPG_AVAILABLE=False\n'
      fi
      if recovery_volume_info="$(diskutil info "/Volumes/MeshAlot-Backup" 2>/dev/null)"; then
        printf 'VOLUME_METADATA_READ=PASS\n'
        printf '%s\n' "$recovery_volume_info" |
          awk '/Volume Name:|Mounted:|File System Personality:|FileVault:/ {print}'
      else
        printf 'VOLUME_METADATA_READ=UNVERIFIED\n'
      fi
    )

The volume-read result only means diskutil returned metadata. The selected
fields must still establish the expected mounted/unlocked encrypted APFS
volume; a directory name or successful command alone is insufficient.

## Execution sequence after device/tool identification

1. Confirm Darwin, the actual mounted/unlocked encrypted APFS backup volume,
   Python 3.8+, GnuPG/gpgconf/cache-option support and private interactive pinentry.
   The verifier performs SHA256 checks itself. Record only classifications and
   nonsecret version/size.
   Missing information is UNVERIFIED; do not format, repair, change encryption,
   install a filesystem driver or modify an existing volume to pass a check.
2. Select each original encrypted copy locally using the existing private
   receipt. Pin its SHA256 to the previously verified value. An expected digest
   newly computed from the same unreviewed file is not independent provenance.
   Keep exact paths, receipts and raw errors out of chat and shared logs.
3. Retrieve the appropriate existing passphrase from its RoboForm Safenote.
   Enter it only in GnuPG's private interactive prompt. Never put it in a shell
   argument, environment variable, command history, script, CI secret or chat.
   Disable GnuPG's symmetric cache so success requires the retrieved key.
4. With the reviewed input selected, decrypt once to the local null device.
   Suppress raw GnuPG diagnostics from shared output. Record a zero exit only
   as authenticated decryption PASS for that exact encrypted asset. Failure,
   cancellation, unsupported integrity format or timeout is BLOCKED/INCOMPLETE;
   stop rather than change credentials or retry automatically.
5. Recheck that the encrypted bytes retain the pinned SHA256. Do not alter mode,
   owner, archive contents or sidecars. Clear the current clipboard after key
   entry; do not claim that clipboard history or external sync was audited.
6. Repeat only for the other two distinct required assets. Keep original
   encrypted archives, MS-02 key files and RoboForm notes intact. Record no
   passphrases, archive contents, plaintext, account/node identities or paths.

The verifier's concrete invocation is documented in its linked README. Select
the actual input/pin only after the Mac's tools, mount and private receipt are
verified. This document does not guess private paths, install software or submit
a real-archive command remotely.

## Acceptance and stopping point

PASS requires all three independently pinned copies to authenticate/decrypt
using freshly retrieved vault keys, with ciphertext unchanged and no plaintext
files or shared secret output. Record each asset separately, plus the device,
tool version, elapsed time and input size using sanitized evidence.

Stop after the three required assets. Do not repeat the prior logical DB
restore, ACL tests, synthetic globals/Caddy tests or benchmarks. Partial success
does not close this key check. A successful key check still leaves real PG16
globals/ownership, TLS recovery requirements and the full service rehearsal open.

G17/1 remains BLOCKED; G17/2–7 NOT STARTED; M13 OPEN / NOT RELEASE-READY.
