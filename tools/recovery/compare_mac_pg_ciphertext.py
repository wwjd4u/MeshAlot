#!/usr/bin/env python3
"""Compare only the existing Mac PostgreSQL ciphertext with the MS-02 receipt."""
import hashlib
import os
from pathlib import Path
import signal
import stat
import sys
import time

EXPECTED = 'f60283f0a5563a70c363cdf863c52ecce9eec410d6112411844d948e4f3a699e'
VOLUME = Path('/Volumes/MeshAlot-Backup')
FOLDER = VOLUME / 'M13-G17-20261011-Encrypted-Archives'
NAME = 'meshalot-production-pg16.dump.gpg'
LIMIT = 2 * 1024**3

class Stop(Exception):
    pass

def require(ok):
    if not ok:
        raise Stop()

def pin(s):
    return (s.st_dev, s.st_ino, s.st_mode, s.st_uid, s.st_gid,
            s.st_nlink, s.st_size, s.st_mtime_ns, s.st_ctime_ns)

def compare(folder=FOLDER, expected=EXPECTED):
    deadline = time.monotonic() + 120
    require(folder.resolve(strict=True) == folder)
    directory = os.open(folder, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    fd = None
    try:
        ds = os.fstat(directory)
        require(ds.st_uid == os.getuid() and stat.S_IMODE(ds.st_mode) == 0o700)
        dp = pin(ds)
        fd = os.open(NAME, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=directory)
        s = os.fstat(fd)
        require(stat.S_ISREG(s.st_mode) and s.st_uid == os.getuid())
        require(stat.S_IMODE(s.st_mode) == 0o600 and s.st_nlink == 1)
        require(0 < s.st_size < LIMIT)
        fp = pin(s)
        digest, count = hashlib.sha256(), 0
        while True:
            require(time.monotonic() < deadline)
            block = os.read(fd, 65536)
            if not block:
                break
            count += len(block)
            require(count < LIMIT)
            digest.update(block)
        require(count == s.st_size and pin(os.fstat(fd)) == fp)
        require(pin(os.stat(NAME, dir_fd=directory, follow_symlinks=False)) == fp)
        require(pin(os.fstat(directory)) == dp and pin(folder.lstat()) == dp)
        require(folder.resolve(strict=True) == folder)
        return 'SAME' if digest.hexdigest() == expected else 'MISMATCH'
    finally:
        if fd is not None:
            os.close(fd)
        os.close(directory)

def main():
    result = 'STOP'
    try:
        require(len(sys.argv) == 1 and sys.platform == 'darwin' and os.getuid() != 0)
        require(not os.environ.get('CI') and not os.environ.get('GITHUB_ACTIONS'))
        require(sys.stdin.isatty() and sys.stdout.isatty() and sys.stderr.isatty())
        require(os.path.ismount(VOLUME) and VOLUME.resolve(strict=True) == VOLUME)
        def stop_signal(signum, frame):
            raise Stop()
        for sig in (signal.SIGALRM, signal.SIGINT, signal.SIGTERM, signal.SIGHUP):
            signal.signal(sig, stop_signal)
        signal.alarm(125)
        result = compare()
        signal.alarm(0)
    except BaseException:
        result = 'STOP'
    print('MAC_POSTGRESQL_CIPHERTEXT=' + result)
    return 0 if result == 'SAME' else 1

if __name__ == '__main__':
    sys.exit(main())
