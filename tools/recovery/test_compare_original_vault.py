"""Public synthetic fixtures only; never use the production default path."""
import os
from pathlib import Path
import pty
import subprocess
import sys
import tempfile
import termios
import threading
import time
import unittest

import compare_original_vault as c

class ComparisonTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(dir=Path.cwd())
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.key = self.root / 'fixture'
        self.key.write_bytes(b'public synthetic value\n')
        self.key.chmod(0o600)
        self.parts = self.key.parts[1:]

    def run_compare(self, value, callback=None):
        with c.source(self.parts) as (fd, check):
            def prompt():
                if callback:
                    callback()
                return value
            return c.compare(fd, check, prompt)

    def test_match_and_mismatch(self):
        self.assertEqual(self.run_compare(b'public synthetic value'), 'MATCH')
        self.assertEqual(self.run_compare(b'wrong'), 'MISMATCH')

    def test_no_normalization(self):
        self.key.write_bytes(b'  public synthetic value  \n')
        self.assertEqual(self.run_compare(b'public synthetic value'), 'MISMATCH')
        self.assertEqual(self.run_compare(b'  public synthetic value  '), 'MATCH')

    def test_unterminated_record(self):
        self.key.write_bytes(b'public')
        self.assertEqual(self.run_compare(b'public'), 'MATCH')

    def test_invalid_records(self):
        for value in (b'', b'\n', b'a\nb\n', b'a\r\n', b'a\0b', b'a' * (c.LIMIT + 1)):
            with self.subTest(value_length=len(value)):
                self.key.write_bytes(value)
                with self.assertRaises(c.Stop):
                    self.run_compare(b'public')

    def test_unsafe_modes(self):
        for mode in (0o644, 0o666, 0o400):
            self.key.chmod(mode)
            with self.assertRaises(c.Stop):
                self.run_compare(b'public')
        self.key.chmod(0o600)
        self.root.chmod(0o755)
        with self.assertRaises(c.Stop):
            self.run_compare(b'public')

    def test_symlink_and_hardlink(self):
        alias = self.root / 'alias'
        alias.symlink_to(self.key)
        with self.assertRaises(OSError):
            with c.source(alias.parts[1:]):
                self.fail('symlink accepted')
        alias.unlink()
        os.link(self.key, alias)
        with self.assertRaises(c.Stop):
            self.run_compare(b'public')

    def test_fifo_does_not_block(self):
        self.key.unlink()
        os.mkfifo(self.key, 0o600)
        with self.assertRaises(c.Stop):
            self.run_compare(b'public')

    def test_changed_source(self):
        with self.assertRaises(c.Stop):
            self.run_compare(b'public', lambda: self.key.write_bytes(b'changed'))

    def test_replaced_source(self):
        def replace():
            self.key.unlink()
            self.key.write_bytes(b'public synthetic value\n')
            self.key.chmod(0o600)
        with self.assertRaises(c.Stop):
            self.run_compare(b'public synthetic value', replace)

    def terminal_case(self, payload, budget=1):
        master, slave = pty.openpty()
        before = termios.tcgetattr(slave)
        output = []
        def run():
            try:
                output.append(c.hidden_input(slave, budget))
            except c.Stop:
                output.append('STOP')
        worker = threading.Thread(target=run)
        worker.start()
        try:
            deadline = time.monotonic() + 1
            while termios.tcgetattr(slave)[3] & termios.ECHO:
                self.assertLess(time.monotonic(), deadline)
                time.sleep(0.001)
            if payload:
                os.write(master, payload)
            worker.join(2)
            self.assertFalse(worker.is_alive())
            self.assertEqual(termios.tcgetattr(slave), before)
            self.assertEqual(output.__len__(), 1)
            return output[0]
        finally:
            os.close(master)
            os.close(slave)

    def test_hidden_terminal(self):
        self.assertEqual(self.terminal_case(b'public\r'), b'public')

    def test_cancel_timeout_overflow_multiline(self):
        for payload, budget in ((b'\x03', 1), (None, .03),
                                (b'a' * (c.LIMIT + 1), 1), (b'a\nb\n', 1)):
            with self.subTest(size=len(payload) if payload else 0):
                self.assertEqual(self.terminal_case(payload, budget), 'STOP')

    def test_cli_refuses_cloud_without_leaking(self):
        result = subprocess.run([sys.executable, '-B', c.__file__], capture_output=True, timeout=3)
        self.assertEqual(result.returncode, 1)
        self.assertEqual(result.stdout, b'STOP\n')
        self.assertEqual(result.stderr, b'')

if __name__ == '__main__':
    unittest.main()
