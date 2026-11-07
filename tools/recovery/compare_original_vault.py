#!/usr/bin/env python3
"""Operator-local comparison only. No secret arguments, logs, or network."""
import contextlib
import hmac
import os
import pwd
import resource
import select
import signal
import socket
import stat
import sys
import termios
import time

LIMIT = 4096
BUDGET = 120
RUN = 'm13-g17-step3b-20261009T152401Z-4ccaf982'
PARTS = ('home', 'jason_guynes', '.local', 'share',
         'meshalot-m13-backup-keys', RUN + '.passphrase')

class Stop(Exception):
    pass

def require(ok):
    if not ok:
        raise Stop()

def fingerprint(s):
    return (s.st_dev, s.st_ino, s.st_mode, s.st_uid, s.st_gid,
            s.st_nlink, s.st_size, s.st_mtime_ns, s.st_ctime_ns)

@contextlib.contextmanager
def source(parts=PARTS):
    """Walk without symlinks; keep parent descriptors through final recheck."""
    fds = []
    try:
        parent = os.open('/', os.O_RDONLY | os.O_DIRECTORY)
        fds.append(parent)
        links = []
        for i, name in enumerate(parts):
            require(name not in ('', '.', '..') and '/' not in name)
            final = i == len(parts) - 1
            fd = os.open(name, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK |
                         (0 if final else os.O_DIRECTORY), dir_fd=parent)
            fds.append(fd)
            s = os.fstat(fd)
            require(s.st_uid in (0, os.getuid()) and not s.st_mode & 0o022)
            if final:
                require(stat.S_ISREG(s.st_mode) and s.st_uid == os.getuid())
                require(stat.S_IMODE(s.st_mode) == 0o600 and s.st_nlink == 1)
                require(0 < s.st_size <= LIMIT)
            elif i == len(parts) - 2:
                require(s.st_uid == os.getuid() and stat.S_IMODE(s.st_mode) == 0o700)
            links.append((parent, name, fd, fingerprint(s)))
            parent = fd
        def check():
            for p, n, f, pin in links:
                require(fingerprint(os.fstat(f)) == pin)
                require(fingerprint(os.stat(n, dir_fd=p, follow_symlinks=False)) == pin)
        check()
        yield fds[-1], check
    finally:
        for fd in reversed(fds):
            os.close(fd)

def read_key(fd):
    os.lseek(fd, 0, os.SEEK_SET)
    value = os.read(fd, LIMIT + 1)
    require(0 < len(value) <= LIMIT)
    # GnuPG --passphrase-file uses its first line. Permit one LF record
    # terminator explicitly; never strip spaces, labels, prefixes or CR.
    if value.endswith(b'\n'):
        value = value[:-1]
    require(bool(value) and b'\n' not in value and b'\r' not in value and b'\0' not in value)
    return value

def hidden_input(fd, budget=BUDGET):
    old = termios.tcgetattr(fd)
    new = termios.tcgetattr(fd)
    new[3] &= ~(termios.ECHO | termios.ECHONL | termios.ICANON | termios.ISIG)
    new[0] &= ~(termios.ICRNL | termios.INLCR | termios.IGNCR | termios.ISTRIP | termios.IXON)
    new[6][termios.VMIN] = 1
    new[6][termios.VTIME] = 0
    data = bytearray()
    deadline = time.monotonic() + budget
    try:
        termios.tcsetattr(fd, termios.TCSAFLUSH, new)
        while True:
            remaining = deadline - time.monotonic()
            require(remaining > 0 and bool(select.select([fd], [], [], remaining)[0]))
            b = os.read(fd, 1)
            require(bool(b))
            if b in (b'\r', b'\n'):
                # Reject queued multiline paste instead of accepting a prefix.
                require(not select.select([fd], [], [], 0)[0] and bool(data))
                return bytes(data)
            require(b[0] >= 32 and b[0] != 127 and len(data) < LIMIT)
            data.extend(b)
    finally:
        termios.tcsetattr(fd, termios.TCSAFLUSH, old)
        for i in range(len(data)):
            data[i] = 0

def compare(fd, check, prompt):
    original = read_key(fd)
    check()
    entered = prompt()
    check()
    require(hmac.compare_digest(original, read_key(fd)))
    check()
    return 'MATCH' if hmac.compare_digest(original, entered) else 'MISMATCH'

def main():
    result = 'STOP'
    try:
        require(len(sys.argv) == 1)
        require(sys.platform == 'linux' and not os.environ.get('CI') and not os.environ.get('GITHUB_ACTIONS'))
        require(socket.gethostname().split('.')[0] == 'wwjd4u-MS-02-Ultra')
        account = pwd.getpwuid(os.getuid())
        require(os.getuid() != 0 and account.pw_name == 'jason_guynes' and account.pw_dir == '/home/jason_guynes')
        require(sys.stdin.isatty() and sys.stdout.isatty() and sys.stderr.isatty())
        resource.setrlimit(resource.RLIMIT_CORE, (0, 0))
        def stop_signal(signum, frame):
            raise Stop()
        for sig in (signal.SIGTERM, signal.SIGHUP, signal.SIGINT, signal.SIGALRM):
            signal.signal(sig, stop_signal)
        signal.alarm(BUDGET + 5)
        with source() as (fd, check):
            with open('/dev/tty', 'r+b', buffering=0) as terminal:
                # Fixed nonsecret prompt is private terminal UI, not a receipt.
                terminal.write(b'Paste the saved PostgreSQL vault value once, then Enter (hidden; 120s): ')
                try:
                    result = compare(fd, check, lambda: hidden_input(terminal.fileno()))
                finally:
                    terminal.write(b'\n')
        signal.alarm(0)
    except BaseException:
        result = 'STOP'
    print(result)
    return 0 if result == 'MATCH' else 1

if __name__ == '__main__':
    sys.exit(main())
