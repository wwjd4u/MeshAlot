# Private PostgreSQL original-key comparison

Scope: operator-local, read-only reconciliation on the already confirmed MS-02
host/account. No SSH, production service access, GPG, archive read/decryption,
vault edit or key-file modification. Never run this with real inputs in CI,
cloud tools, shared terminals or terminal recording. Retrieve the existing
PostgreSQL Safenote before starting. Only MATCH, MISMATCH or STOP may be shared.

The source file selection is fixed from the immutable custody workflow:
https://github.com/wwjd4u/MeshAlot/blob/b00ad2faed0de865c27244b3596528a5d0723bf9/.github/workflows/ms02-m13-gate17-local-evidence-readonly.yml
The helper neither searches for keys nor accepts a path or secret argument.

Run the reviewed, hash-verified helper with `python3 -I -B compare_original_vault.py`
in a private local terminal on MS-02. Paste only the saved PostgreSQL vault
value once at the hidden prompt, then press Enter. Do not paste the note label,
edit the saved entry, guess variants or send any part of the value to chat.
Ctrl-C cancels. Input is limited to 4096 bytes and 120 seconds. Multiline paste,
control characters and overflow stop the check. Backspace is deliberately not
an editing operation; it stops the check. A STOP is not a mismatch.

Before content access, the helper checks Linux/host/account/home, terminal-only
execution, no CI flags, regular owner-only 0600 single-link key file and 0700
key directory. It traverses each path component through no-follow descriptors,
rejecting group/other writable ancestors. It retains descriptors and verifies
identities and metadata through completion, then rereads bytes before reporting.
No permissions are broadened. Reads can update access time; unchanged means
bytes and fingerprinted identity/mode/owner/link/size/mtime/ctime, excluding atime.
This is an owner-controlled filesystem check, not protection from a malicious
same-user writer capable of racing checks or a compromised terminal/OS.

Comparison contract: GnuPG passphrase files represent a first-line value. One
final LF, if present, is treated explicitly as the file record terminator.
Additional LF, CR, NUL or an empty record stops the check. Every other byte,
including leading/trailing spaces and punctuation, is compared exactly using
constant-time comparison. No Unicode normalization, label removal, whitespace
trimming or variant attempts occur. The terminal supplies bytes in its current
encoding; a mismatch alone does not diagnose why the sources differ.

Core dumps are disabled before reading. Secrets exist briefly in Python memory;
Python cannot guarantee complete memory erasure. No secret output, digest,
length, private path, traceback, log, plaintext file or child process is created.
The terminal settings are restored on ordinary errors, signals and timeout;
SIGKILL/power loss cannot be caught. Clear the current clipboard privately after
entry; this does not audit clipboard history or sync.

MATCH establishes equality only under this record contract; it does not establish
archive key access, recovery, or gate acceptance. MISMATCH stops for custody
review. STOP means a precondition, input, timeout or unchanged-source check did
not pass. Do not retry automatically or change originals based on either result.

Synthetic verification only:

    python3 -B -m unittest discover -s tools/recovery -p 'test_compare_original_vault.py' -v

No accepted GPG, application, restore or service tests need repeating for this
helper. Real comparison remains NOT TESTED until the operator returns a result.
