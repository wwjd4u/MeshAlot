# Public synthetic Mac paste check

Use only after the recorded real PostgreSQL retry again reported bad passphrase,
with original/vault MATCH and Mac/original ciphertext SAME already established.
The accepted 21-test suite used an automated synthetic pinentry. This single
manual check addresses a different unresolved interface: current Mac clipboard
paste into the real installed private GnuPG prompt. No real archive/key, vault
entry, external volume, production access or new installation is involved.

Run hash-pinned `check_mac_public_paste.py` beside the hash-pinned unchanged
`verify_archive_key.py` from fc040c80b40106396b618217d190df9b14a72a64, on the
Mac with MacGPG2/bin on PATH. Retrieve only these two public scripts. The helper
also independently hashes the verifier before loading it. It accepts no input
paths, passphrase arguments or secrets. Do not open RoboForm or copy a real key.

The helper builds a tiny public fixture encrypted with known public test value
`MeshAlot-PUBLIC-paste-check-2026`, using a separate temporary GnuPG home and
loopback entry solely for public fixture creation. It shuts down that generation
home's agents before starting the accepted verifier, which uses a different new
private home and the actual installed pinentry. It puts the public test value on
the current clipboard without a trailing newline. At the private GnuPG prompt,
press Command-V once and confirm. Do not type/copy a real vault value.

Acceptance: PASS_KEY_ACCESS and unchanged synthetic ciphertext. Failure stops;
no automatic retry or real-archive follow-on. This does not test RoboForm copy
selection or diagnose a particular credential. Even PASS leaves real key access
and G17/1 BLOCKED. Share only the sanitized JSON with PUBLIC_SYNTHETIC_PASTE_ONLY.

The fixture and temporary homes are removed. Current clipboard is cleared on
exit, and failure to clear it blocks acceptance; history/sync clearing is not
claimed. Clipboard contents are replaced, not backed up. Temporary files contain
only public synthetic data. Core-verifier failure/redaction/cleanup checks remain
unchanged. New fixture generation is capped at 30 seconds; core phases at 120;
cleanup commands at five. Private prompt interaction is the operator test still
pending, not a claim from Linux tests.

Focused new tests (not the accepted verifier suite):

    python3 -B -m unittest discover -s tools/recovery -p 'test_check_mac_public_paste.py' -v

A Mac operator run stopped during setup before the prompt. The revised wrapper
adds fixed `stage` and `failure_kind` fields on exceptions, without raw errors
or private paths. A failure before PUBLIC_INTERACTIVE_VERIFIER is not a paste
result. Reuse of the wrapper check is justified by that concrete setup failure;
no accepted core tests or actual backup attempts need repeating.

The diagnostic Mac run stopped at PUBLIC_FIXTURE_ENCRYPTION. Local public-only
reproduction with a long temporary home failed at the same stage and emitted
agent-related diagnostics. The Mac root cause is not confirmed by that receipt.
The fixture generator now uses a separate short, owner-only random /tmp home,
cleans its agent before deleting it, and never accesses the existing keyring.
The accepted verifier and its own private-home handling are unchanged. Long
fixture-directory regression and public diagnostic-redaction tests were added.
On generation failure, a temporary public-only diagnostic file is read up to
8 KiB and removed; only fixed diagnostic enums are reported, never raw stderr
or paths. A new public-only attempt is warranted by the specific setup failure.
