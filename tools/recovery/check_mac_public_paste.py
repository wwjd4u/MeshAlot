#!/usr/bin/env python3
"""One public synthetic clipboard -> actual private pinentry check. No real inputs."""
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

PUBLIC_VALUE = b'MeshAlot-PUBLIC-paste-check-2026'
PUBLIC_DATA = b'Public synthetic clipboard and pinentry fixture. No production data.\n'
VERIFIER_SHA = '4b05bd2ff0b1efa7a1a5c37b4a420affb334412824bd5aad60bdb97b871771df'

class Stop(Exception):
    pass

def require(ok):
    if not ok:
        raise Stop()

def load_verifier(path):
    data = path.read_bytes()
    require(hashlib.sha256(data).hexdigest() == VERIFIER_SHA)
    spec = importlib.util.spec_from_file_location('public_paste_verifier', path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module

def generate(folder, gpg, gpgconf):
    home = folder / 'generation-home'
    home.mkdir(mode=0o700)
    plain = folder / 'public.txt'
    plain.write_bytes(PUBLIC_DATA)
    plain.chmod(0o600)
    archive = folder / 'public.gpg'
    try:
        result = subprocess.run(
            [gpg, '--no-options', '--homedir', str(home), '--batch', '--yes',
             '--pinentry-mode', 'loopback', '--passphrase-fd', '0',
             '--cipher-algo', 'AES256', '--symmetric', '--output', str(archive), str(plain)],
            input=PUBLIC_VALUE + b'\n', stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL, timeout=30)
        require(result.returncode == 0)
        archive.chmod(0o600)
        digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    finally:
        cleanup = subprocess.run([gpgconf, '--homedir', str(home), '--kill', 'all'],
                                 stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=5)
        require(cleanup.returncode == 0)
    return archive, digest

def main():
    report = {'status': 'BLOCKED', 'reason': 'PUBLIC_PASTE_SETUP_STOP',
              'scope': 'PUBLIC_SYNTHETIC_PASTE_ONLY', 'real_archives': 'NOT_USED'}
    clipboard_used = False
    try:
        require(len(sys.argv) == 1 and sys.platform == 'darwin' and os.getuid() != 0)
        require(not os.environ.get('CI') and not os.environ.get('GITHUB_ACTIONS'))
        require(sys.stdin.isatty() and sys.stdout.isatty() and sys.stderr.isatty())
        gpg, gpgconf, pbcopy = (shutil.which(t) for t in ('gpg', 'gpgconf', 'pbcopy'))
        require(all((gpg, gpgconf, pbcopy)))
        verifier = load_verifier(Path(__file__).with_name('verify_archive_key.py'))
        with tempfile.TemporaryDirectory(prefix='meshalot-public-paste-') as temp:
            archive, digest = generate(Path(temp), gpg, gpgconf)
            clipboard_used = True
            copied = subprocess.run([pbcopy], input=PUBLIC_VALUE, stdout=subprocess.DEVNULL,
                                    stderr=subprocess.DEVNULL, timeout=5)
            require(copied.returncode == 0)
            print('PUBLIC FIXTURE ONLY: At the GnuPG prompt, press Command-V once, then confirm.', flush=True)
            result = verifier.verify_archive(archive, digest, 120)
            report = {'status': result['status'], 'reason': result['reason'],
                      'source_unchanged': result['source_unchanged'],
                      'scope': 'PUBLIC_SYNTHETIC_PASTE_ONLY', 'real_archives': 'NOT_USED'}
    except BaseException:
        report = {'status': 'BLOCKED', 'reason': 'PUBLIC_PASTE_SETUP_OR_CLEANUP_STOP',
                  'scope': 'PUBLIC_SYNTHETIC_PASTE_ONLY', 'real_archives': 'NOT_USED'}
    finally:
        if clipboard_used:
            try:
                cleared = subprocess.run([pbcopy], input=b'', stdout=subprocess.DEVNULL,
                                         stderr=subprocess.DEVNULL, timeout=5)
                require(cleared.returncode == 0)
            except BaseException:
                report['status'] = 'BLOCKED'
                report['reason'] = 'PUBLIC_CLIPBOARD_CLEANUP_UNVERIFIED'
    print(json.dumps(report, sort_keys=True))
    return 0 if report['status'] == 'PASS_KEY_ACCESS' else 1

if __name__ == '__main__':
    sys.exit(main())
