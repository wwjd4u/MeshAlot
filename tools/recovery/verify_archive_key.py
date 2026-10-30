#!/usr/bin/env python3
"""Local, interactive key-access check for a separately pinned encrypted archive.

Discards decrypted output. Never restores, extracts, connects or closes a gate.
Real archive use belongs in the operator's private terminal, never CI.
"""
import argparse
import hashlib
import math
import os
from pathlib import Path
import re
import selectors
import shutil
import stat
import subprocess
import sys
import tempfile
import time


MAX_ARCHIVE = 4 * 1024 ** 3
MAX_STATUS = 65536


class Blocked(Exception):
    pass


def require(condition, reason):
    if not condition:
        raise Blocked(reason)


def receipt(status="BLOCKED", reason="UNVERIFIED", unchanged="NOT_ASSESSED"):
    return {"status": status, "reason": reason, "source_unchanged": unchanged,
            "scope": "LOCAL_ARCHIVE_KEY_ACCESS_ONLY",
            "vault_source": "OPERATOR_ATTESTATION_REQUIRED",
            "gate_acceptance": "NOT_ASSESSED", "restore_acceptance": "NOT_TESTED"}


def metadata(fd):
    info = os.fstat(fd)
    require(stat.S_ISREG(info.st_mode) and info.st_nlink == 1,
            "UNSAFE_SOURCE_TYPE_OR_LINK")
    require(info.st_uid == os.geteuid() and not info.st_mode & 0o022,
            "UNSAFE_SOURCE_OWNERSHIP_OR_MODE")
    require(0 < info.st_size <= MAX_ARCHIVE, "SOURCE_SIZE_NOT_ALLOWED")
    return (info.st_dev, info.st_ino, info.st_mode, info.st_uid, info.st_nlink,
            info.st_size, info.st_mtime_ns, info.st_ctime_ns)


def fingerprint(fd, seconds):
    before = metadata(fd)
    deadline = time.monotonic() + seconds
    digest, count = hashlib.sha256(), 0
    os.lseek(fd, 0, os.SEEK_SET)
    while chunk := os.read(fd, 65536):
        require(time.monotonic() <= deadline, "HASH_TIMEOUT")
        count += len(chunk)
        require(count <= MAX_ARCHIVE, "SOURCE_SIZE_NOT_ALLOWED")
        digest.update(chunk)
    require(count == before[5] and metadata(fd) == before, "SOURCE_CHANGED")
    return digest.hexdigest(), before


def _prepare_private_home(home):
    # The agent must not obtain the passphrase from an external/keychain cache.
    config = home / "gpg-agent.conf"
    config.write_text("no-allow-external-cache\ndefault-cache-ttl 0\nmax-cache-ttl 0\n")
    config.chmod(0o600)


def _run_crypto(command, fd, seconds, environment):
    """Bound captured metadata without imposing file limits on pinentry."""
    process = subprocess.Popen(command, stdin=subprocess.DEVNULL,
                               stdout=subprocess.PIPE, stderr=subprocess.DEVNULL,
                               pass_fds=(fd,), env=environment, start_new_session=True)
    data = bytearray()
    failure = None
    try:
        os.set_blocking(process.stdout.fileno(), False)
        deadline = time.monotonic() + seconds
        with selectors.DefaultSelector() as selector:
            selector.register(process.stdout, selectors.EVENT_READ)
            while True:
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    failure = "DECRYPTION_TIMEOUT"
                    break
                ready = selector.select(min(remaining, 0.1))
                finished = process.poll() is not None
                if ready or finished:
                    while True:
                        try:
                            chunk = os.read(process.stdout.fileno(), 8192)
                        except BlockingIOError:
                            break
                        if not chunk:
                            if selector.get_map():
                                selector.unregister(process.stdout)
                            break
                        if len(data) + len(chunk) > MAX_STATUS:
                            failure = "STATUS_LIMIT_REACHED"
                            break
                        data.extend(chunk)
                if failure or finished:
                    break
        if failure is None and process.returncode != 0:
            failure = "DECRYPTION_FAILED"
        return failure, bytes(data)
    finally:
        if process.poll() is None:
            process.kill()
        try:
            process.wait(timeout=2)
        except subprocess.TimeoutExpired:
            raise Blocked("CRYPTO_PROCESS_CLEANUP_UNVERIFIED") from None
        finally:
            process.stdout.close()


def _cleanup(gpgconf, home):
    try:
        result = subprocess.run(
            [gpgconf, "--homedir", str(home), "--kill", "all"],
            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=5)
        return result.returncode == 0
    except (OSError, subprocess.SubprocessError):
        return False


def _authenticated_aes256(raw):
    events = [line.split()[1:] for line in raw.splitlines()
              if line.startswith(b"[GNUPG:] ")]
    tags = {event[0] for event in events if event}
    if not {b"NEED_PASSPHRASE_SYM", b"DECRYPTION_OKAY", b"END_DECRYPTION"} <= tags:
        return False
    infos = [event[1:] for event in events if event and event[0] == b"DECRYPTION_INFO"]
    if len(infos) != 1 or len(infos[0]) < 2:
        return False
    info = infos[0]
    if info[1] != b"9":  # GnuPG's AES256 cipher identifier.
        return False
    mdc = info[0].isdigit() and int(info[0]) > 0 and b"GOODMDC" in tags
    aead = len(info) >= 3 and info[2].isdigit() and int(info[2]) > 0
    return bool(mdc or aead)


def verify_archive(archive, expected_sha256, seconds=120):
    """Uses GnuPG's private pinentry; callers never pass a key to this function."""
    unchanged = "NOT_ASSESSED"
    fd = None
    try:
        require(isinstance(expected_sha256, str) and
                bool(re.fullmatch(r"[0-9a-f]{64}", expected_sha256)), "INVALID_DIGEST_PIN")
        require(type(seconds) in (int, float) and math.isfinite(seconds) and
                0.05 <= seconds <= 600, "INVALID_TIME_LIMIT")
        require(os.name == "posix" and hasattr(os, "O_NOFOLLOW"), "UNSUPPORTED_PLATFORM")
        gpg, gpgconf = shutil.which("gpg"), shutil.which("gpgconf")
        require(bool(gpg and gpgconf), "GNUPG_TOOLS_UNAVAILABLE")
        options = subprocess.run([gpg, "--no-options", "--dump-options"],
                                 capture_output=True, timeout=5)
        required = {b"--no-symkey-cache", b"--pinentry-mode", b"--status-fd",
                    b"--no-auto-key-retrieve", b"--no-auto-key-import"}
        require(options.returncode == 0 and required <= set(options.stdout.splitlines()),
                "GNUPG_OPTIONS_UNSUPPORTED")
        fd = os.open(archive, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
        digest, before = fingerprint(fd, seconds)
        require(digest == expected_sha256, "SOURCE_DIGEST_MISMATCH")
        with tempfile.TemporaryDirectory(prefix="meshalot-g17-key-") as directory:
            home = Path(directory)
            _prepare_private_home(home)
            failure, raw = None, b""
            try:
                require(metadata(fd) == before, "SOURCE_CHANGED")
                os.lseek(fd, 0, os.SEEK_SET)
                command = [gpg, "--no-options", "--homedir", str(home),
                           "--no-symkey-cache", "--pinentry-mode", "ask",
                           "--no-auto-key-retrieve", "--no-auto-key-import",
                           "--status-fd", "1", "--yes", "--output", os.devnull,
                           "--decrypt", "/dev/fd/" + str(fd)]
                environment = dict(os.environ)
                if sys.stdin.isatty():
                    environment["GPG_TTY"] = os.ttyname(sys.stdin.fileno())
                failure, raw = _run_crypto(command, fd, seconds, environment)
            finally:
                require(_cleanup(gpgconf, home), "PRIVATE_AGENT_CLEANUP_UNVERIFIED")
        after_digest, after = fingerprint(fd, seconds)
        unchanged = after_digest == expected_sha256 and after == before
        require(unchanged, "SOURCE_CHANGED")
        require(failure is None, failure)
        require(_authenticated_aes256(raw), "AUTHENTICATED_AES256_NOT_ESTABLISHED")
        return receipt("PASS_KEY_ACCESS", "PINNED_AES256_ARCHIVE_AUTHENTICATED", True)
    except Blocked as error:
        return receipt(reason=str(error), unchanged=unchanged)
    except KeyboardInterrupt:
        return receipt(reason="OPERATOR_CANCELLED", unchanged=unchanged)
    except (OSError, ValueError, subprocess.SubprocessError):
        return receipt(reason="LOCAL_INPUT_OR_CRYPTO_TOOL_UNAVAILABLE", unchanged=unchanged)
    finally:
        if fd is not None:
            os.close(fd)


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--archive", type=Path, required=True)
    parser.add_argument("--expected-sha256", required=True,
                        help="digest from the independently reviewed existing receipt")
    parser.add_argument("--timeout", type=float, default=120,
                        help="seconds per hash/decryption phase, 0.05 to 600; default 120")
    args = parser.parse_args(argv)
    if not sys.stdin.isatty():
        report = receipt(reason="PRIVATE_INTERACTIVE_TERMINAL_REQUIRED")
    else:
        report = verify_archive(args.archive, args.expected_sha256, args.timeout)
    import json
    print(json.dumps(report, sort_keys=True))
    return 0 if report["status"] == "PASS_KEY_ACCESS" else 1


if __name__ == "__main__":
    sys.exit(main())
