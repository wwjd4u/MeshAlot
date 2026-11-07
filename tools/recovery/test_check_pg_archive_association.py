"""Synthetic ciphertext and receipt only; no actual archive or key access."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
import check_pg_archive_association as c

class AssociationTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(dir=Path.cwd())
        self.addCleanup(self.temp.cleanup)
        self.folder = Path(self.temp.name) / 'fixture'
        self.folder.mkdir(mode=0o700)
        self.archive = self.folder / c.ARCHIVE
        self.archive.write_bytes(b'public synthetic ciphertext placeholder')
        self.digest = hashlib.sha256(self.archive.read_bytes()).hexdigest()
        self.sidecar = self.folder / (c.ARCHIVE + '.sha256')
        self.sidecar.write_text(self.digest + '  ' + str(self.archive) + '\n')
        self.record = {'run_id': c.RUN, 'archive_sha256': self.digest,
                       'archive_size_bytes': self.archive.stat().st_size,
                       'evidence_source': 'Operator-provided terminal results',
                       'postgres_version': 16, 'isolated_full_restore': 'PASS',
                       'migrations_1_through_8': 'PASS', 'restored_table_reads': 'PASS',
                       'production_restore': 'NOT PERFORMED'}
        self.receipt = self.folder / c.RECEIPT
        self.save()
        for p in self.folder.iterdir():
            p.chmod(0o600)
    def save(self):
        self.receipt.write_text(json.dumps(self.record))
    def test_associated(self):
        self.assertEqual(c.verify(self.folder), self.digest)
    def test_wrong_record_fields(self):
        for key in self.record:
            original = self.record[key]
            self.record[key] = 'incorrect'
            self.save()
            with self.subTest(key=key), self.assertRaises(c.Stop):
                c.verify(self.folder)
            self.record[key] = original
    def test_wrong_checksum_path_and_digest(self):
        for value in (self.digest + '  another-file\n', '0' * 64 + '  ' + str(self.archive) + '\n'):
            self.sidecar.write_text(value)
            with self.assertRaises(c.Stop):
                c.verify(self.folder)
    def test_changed_ciphertext(self):
        self.archive.write_bytes(b'changed')
        with self.assertRaises(c.Stop):
            c.verify(self.folder)
    def test_links_modes_and_sizes(self):
        self.archive.chmod(0o644)
        with self.assertRaises(c.Stop): c.verify(self.folder)
        self.archive.chmod(0o600)
        alias = self.folder / 'alias'
        os.link(self.archive, alias)
        with self.assertRaises(c.Stop): c.verify(self.folder)
        alias.unlink()
        self.archive.unlink()
        self.archive.symlink_to(self.sidecar)
        with self.assertRaises(OSError): c.verify(self.folder)
    def test_malformed_duplicate_receipt(self):
        for data in ('{', '[]', '{"run_id":"a","run_id":"b"}'):
            self.receipt.write_text(data)
            with self.assertRaises((c.Stop, ValueError)):
                c.verify(self.folder)
    def test_timeout(self):
        with patch.object(c.time, 'monotonic', side_effect=[0, 121]):
            with self.assertRaises(c.Stop): c.verify(self.folder)
    def test_concurrent_receipt_change(self):
        original = c.small
        def mutate(fd, limit):
            result = original(fd, limit)
            if limit == 65536:
                self.receipt.write_text('{}')
            return result
        with patch.object(c, 'small', side_effect=mutate):
            with self.assertRaises(c.Stop): c.verify(self.folder)
    def test_refused_cloud_cli_sanitized(self):
        result = subprocess.run([sys.executable, '-B', c.__file__], capture_output=True, timeout=3)
        self.assertEqual((result.returncode, result.stdout, result.stderr),
                         (1, b'ARCHIVE_ASSOCIATION=STOP\n', b''))

if __name__ == '__main__': unittest.main()
