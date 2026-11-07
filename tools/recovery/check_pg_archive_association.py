#!/usr/bin/env python3
"""Read-only original ciphertext/receipt association; never opens a key file."""
import contextlib
import hashlib
import json
import os
from pathlib import Path
import pwd
import signal
import socket
import stat
import sys
import time

RUN = 'm13-g17-step3b-20261009T152401Z-4ccaf982'
FOLDER = Path('/home/jason_guynes/.local/share/meshalot-m13-private-backups') / RUN
ARCHIVE = 'meshalot-production-pg16.dump.gpg'
RECEIPT = 'm13-step3c-restore-evidence.json'
MAX_ARCHIVE = 2 * 1024**3

class Stop(Exception):
    pass

def require(ok):
    if not ok:
        raise Stop()

def pin(s):
    return (s.st_dev, s.st_ino, s.st_mode, s.st_uid, s.st_gid,
            s.st_nlink, s.st_size, s.st_mtime_ns, s.st_ctime_ns)

@contextlib.contextmanager
def opened(folder):
    fds, links, files = [], [], {}
    try:
        parent = os.open('/', os.O_RDONLY | os.O_DIRECTORY)
        fds.append(parent)
        for i, name in enumerate(folder.parts[1:]):
            fd = os.open(name, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW,
                         dir_fd=parent)
            fds.append(fd)
            s = os.fstat(fd)
            require(s.st_uid in (0, os.getuid()) and not s.st_mode & 0o022)
            if i >= len(folder.parts[1:]) - 2:
                require(s.st_uid == os.getuid() and stat.S_IMODE(s.st_mode) == 0o700)
            links.append((parent, name, fd, pin(s)))
            parent = fd
        for name, limit in ((ARCHIVE, MAX_ARCHIVE), (ARCHIVE + '.sha256', 8192),
                            (RECEIPT, 65536)):
            fd = os.open(name, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=parent)
            fds.append(fd)
            s = os.fstat(fd)
            require(stat.S_ISREG(s.st_mode) and s.st_uid == os.getuid())
            require(stat.S_IMODE(s.st_mode) == 0o600 and s.st_nlink == 1)
            require(0 < s.st_size < limit)
            links.append((parent, name, fd, pin(s)))
            files[name] = fd
        def unchanged():
            for p, n, f, expected in links:
                require(pin(os.fstat(f)) == expected)
                require(pin(os.stat(n, dir_fd=p, follow_symlinks=False)) == expected)
        unchanged()
        yield files, unchanged
    finally:
        for fd in reversed(fds):
            os.close(fd)

def small(fd, limit):
    data = os.read(fd, limit + 1)
    require(0 < len(data) <= limit)
    return data

def unique_object(pairs):
    result = {}
    for key, value in pairs:
        require(key not in result)
        result[key] = value
    return result

def verify(folder=FOLDER):
    deadline = time.monotonic() + 120
    with opened(folder) as (files, unchanged):
        fd = files[ARCHIVE]
        digest = hashlib.sha256()
        count = 0
        while True:
            require(time.monotonic() < deadline)
            block = os.read(fd, 65536)
            if not block:
                break
            count += len(block)
            require(count < MAX_ARCHIVE)
            digest.update(block)
        require(count == os.fstat(fd).st_size)
        value = digest.hexdigest()
        require(small(files[ARCHIVE + '.sha256'], 8192).decode('utf-8') ==
                value + '  ' + str(folder / ARCHIVE) + '\n')
        record = json.loads(small(files[RECEIPT], 65536), object_pairs_hook=unique_object)
        require(isinstance(record, dict))
        expected = {'run_id': RUN, 'archive_sha256': value, 'archive_size_bytes': count,
                    'evidence_source': 'Operator-provided terminal results',
                    'postgres_version': 16, 'isolated_full_restore': 'PASS',
                    'migrations_1_through_8': 'PASS', 'restored_table_reads': 'PASS',
                    'production_restore': 'NOT PERFORMED'}
        require(all(type(record.get(k)) is type(v) and record[k] == v
                    for k, v in expected.items()))
        unchanged()
        return value

def main():
    try:
        require(len(sys.argv) == 1 and sys.platform == 'linux')
        require(not os.environ.get('CI') and not os.environ.get('GITHUB_ACTIONS'))
        require(socket.gethostname().split('.')[0] == 'wwjd4u-MS-02-Ultra')
        account = pwd.getpwuid(os.getuid())
        require(os.getuid() != 0 and account.pw_name == 'jason_guynes' and
                account.pw_dir == '/home/jason_guynes')
        require(sys.stdin.isatty() and sys.stdout.isatty() and sys.stderr.isatty())
        def stop_signal(signum, frame):
            raise Stop()
        for sig in (signal.SIGALRM, signal.SIGTERM, signal.SIGHUP, signal.SIGINT):
            signal.signal(sig, stop_signal)
        signal.alarm(125)
        digest = verify()
        signal.alarm(0)
    except BaseException:
        print('ARCHIVE_ASSOCIATION=STOP')
        return 1
    print('ARCHIVE_ASSOCIATION=PASS')
    print('CIPHERTEXT_SHA256=' + digest)
    return 0

if __name__ == '__main__':
    sys.exit(main())
