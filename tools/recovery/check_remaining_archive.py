#!/usr/bin/env python3
"""One remaining Mac archive per invocation. No key input or PostgreSQL repeat."""
import hashlib
import importlib.util
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import subprocess
import sys

VOLUME = Path('/Volumes/MeshAlot-Backup')
FOLDER = VOLUME / 'M13-G17-20261011-Encrypted-Archives'
ASSETS = {'API_ROLLBACK': 'known-good-api-systemd-caddy.tar.gpg',
          'ENV_SUPPLEMENT': 'meshalot-systemd-env-supplement.tar.gpg'}
VERIFIER_SHA = '4b05bd2ff0b1efa7a1a5c37b4a420affb334412824bd5aad60bdb97b871771df'

class Stop(Exception):
    pass

def require(ok):
    if not ok:
        raise Stop()

def pin(s):
    return (s.st_dev, s.st_ino, s.st_mode, s.st_uid, s.st_gid,
            s.st_nlink, s.st_size, s.st_mtime_ns, s.st_ctime_ns)

def parse_receipt(data, filename):
    require(0 < len(data) <= 8192)
    match = re.fullmatch(rb'([0-9a-f]{64}) [ *]([^\r\n\x00]+)\n?', data)
    require(match is not None)
    name = match[2].decode('utf-8')
    require(PurePosixPath(name).name == filename)
    return match[1].decode('ascii')

def run_asset(asset, verifier, folder=FOLDER):
    require(asset in ASSETS)
    require(folder.resolve(strict=True) == folder)
    directory = os.open(folder, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    receipt_fd = None
    try:
        ds = os.fstat(directory)
        require(ds.st_uid == os.getuid() and stat.S_IMODE(ds.st_mode) == 0o700)
        dp = pin(ds)
        filename = ASSETS[asset]
        receipt_name = filename + '.sha256'
        receipt_fd = os.open(receipt_name, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK,
                             dir_fd=directory)
        s = os.fstat(receipt_fd)
        require(stat.S_ISREG(s.st_mode) and s.st_nlink == 1 and s.st_uid == os.getuid())
        require(stat.S_IMODE(s.st_mode) == 0o600 and 0 < s.st_size <= 8192)
        rp = pin(s)
        data = os.read(receipt_fd, 8193)
        require(len(data) == s.st_size)
        digest = parse_receipt(data, filename)
        def unchanged():
            require(pin(os.fstat(directory)) == dp and pin(folder.lstat()) == dp)
            require(folder.resolve(strict=True) == folder)
            require(pin(os.fstat(receipt_fd)) == rp)
            require(pin(os.stat(receipt_name, dir_fd=directory, follow_symlinks=False)) == rp)
            os.lseek(receipt_fd, 0, os.SEEK_SET)
            require(os.read(receipt_fd, 8193) == data)
        unchanged()
        report = verifier.verify_archive(folder / filename, digest, 600)
        unchanged()
        report['asset'] = asset
        return report
    finally:
        if receipt_fd is not None:
            os.close(receipt_fd)
        os.close(directory)

def main():
    report = {'status': 'BLOCKED', 'reason': 'REMAINING_ARCHIVE_PRECONDITION_STOP',
              'scope': 'LOCAL_ARCHIVE_KEY_ACCESS_ONLY', 'gate_acceptance': 'NOT_ASSESSED',
              'restore_acceptance': 'NOT_TESTED', 'vault_source': 'OPERATOR_ATTESTATION_REQUIRED'}
    clipboard_tool = None
    try:
        require(len(sys.argv) == 2 and sys.argv[1] in ASSETS)
        require(sys.platform == 'darwin' and os.getuid() != 0)
        require(not os.environ.get('CI') and not os.environ.get('GITHUB_ACTIONS'))
        require(sys.stdin.isatty() and sys.stdout.isatty() and sys.stderr.isatty())
        require(os.path.ismount(VOLUME) and VOLUME.resolve(strict=True) == VOLUME)
        import shutil
        clipboard_tool = shutil.which('pbcopy')
        require(bool(clipboard_tool))
        path = Path(__file__).with_name('verify_archive_key.py')
        require(hashlib.sha256(path.read_bytes()).hexdigest() == VERIFIER_SHA)
        spec = importlib.util.spec_from_file_location('remaining_verifier', path)
        verifier = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(verifier)
        print('ASSET=' + sys.argv[1] + ' ROBOFORM_NOTE=API_ROLLBACK', flush=True)
        report = run_asset(sys.argv[1], verifier)
    except BaseException:
        report['status'] = 'BLOCKED'
        report['reason'] = 'REMAINING_ARCHIVE_PRECONDITION_OR_RECEIPT_STOP'
    finally:
        if clipboard_tool:
            try:
                cleared = subprocess.run([clipboard_tool], input=b'', stdout=subprocess.DEVNULL,
                                         stderr=subprocess.DEVNULL, timeout=5)
                require(cleared.returncode == 0)
            except BaseException:
                report['status'] = 'BLOCKED'
                report['reason'] = 'CLIPBOARD_CLEANUP_UNVERIFIED'
    print(json.dumps(report, sort_keys=True))
    return 0 if report['status'] == 'PASS_KEY_ACCESS' else 1

if __name__ == '__main__':
    sys.exit(main())
