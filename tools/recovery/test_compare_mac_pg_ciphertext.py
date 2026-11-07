"""Synthetic files only; no Mac volume or actual backup access."""
import hashlib
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
import compare_mac_pg_ciphertext as c

class MacComparisonTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(dir=Path.cwd())
        self.addCleanup(self.temp.cleanup)
        self.folder = Path(self.temp.name)
        self.file = self.folder / c.NAME
        self.file.write_bytes(b'public fixture')
        self.file.chmod(0o600)
        self.expected = hashlib.sha256(b'public fixture').hexdigest()
    def check(self):
        return c.compare(self.folder, self.expected)
    def test_same_and_different(self):
        self.assertEqual(self.check(), 'SAME')
        self.file.write_bytes(b'different fixture')
        self.assertEqual(self.check(), 'MISMATCH')
    def test_modes(self):
        self.file.chmod(0o644)
        with self.assertRaises(c.Stop): self.check()
        self.file.chmod(0o600)
        self.folder.chmod(0o755)
        with self.assertRaises(c.Stop): self.check()
    def test_links(self):
        link = self.folder / 'link'
        os.link(self.file, link)
        with self.assertRaises(c.Stop): self.check()
        link.unlink()
        self.file.unlink()
        self.file.symlink_to('missing')
        with self.assertRaises(OSError): self.check()
    def test_empty_and_fifo(self):
        self.file.write_bytes(b'')
        with self.assertRaises(c.Stop): self.check()
        self.file.unlink()
        os.mkfifo(self.file, 0o600)
        with self.assertRaises(c.Stop): self.check()
    def test_timeout(self):
        with patch.object(c.time, 'monotonic', side_effect=[0, 121]):
            with self.assertRaises(c.Stop): self.check()
    def test_concurrent_change(self):
        original = c.os.read
        def mutate(fd, count):
            value = original(fd, count)
            if value: self.file.write_bytes(b'modified during read')
            return value
        with patch.object(c.os, 'read', side_effect=mutate):
            with self.assertRaises(c.Stop): self.check()
    def test_cli_refuses_linux(self):
        result = subprocess.run([sys.executable, '-B', c.__file__], capture_output=True, timeout=3)
        self.assertEqual((result.returncode, result.stdout, result.stderr),
                         (1, b'MAC_POSTGRESQL_CIPHERTEXT=STOP\n', b''))

if __name__ == '__main__': unittest.main()
