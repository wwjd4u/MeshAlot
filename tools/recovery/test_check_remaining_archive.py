"""New receipt/selection wrapper checks; no real keys or crypto invocation."""
from pathlib import Path
import tempfile
import unittest
from types import SimpleNamespace
from unittest.mock import Mock
import check_remaining_archive as c

class RemainingTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(dir=Path.cwd())
        self.addCleanup(self.temp.cleanup)
        self.folder = Path(self.temp.name)
        self.asset = 'API_ROLLBACK'
        self.name = c.ASSETS[self.asset]
        self.receipt = self.folder / (self.name + '.sha256')
        self.data = b'a' * 64 + b'  /original/' + self.name.encode() + b'\n'
        self.receipt.write_bytes(self.data)
        self.receipt.chmod(0o600)
    def test_single_selection_and_budget(self):
        fn = Mock(return_value={'status':'PASS_KEY_ACCESS','source_unchanged':True})
        report = c.run_asset(self.asset, SimpleNamespace(verify_archive=fn), self.folder)
        fn.assert_called_once_with(self.folder / self.name, 'a'*64, 600)
        self.assertEqual(report['asset'], self.asset)
    def test_environment_uses_its_own_receipt(self):
        name = c.ASSETS['ENV_SUPPLEMENT']
        receipt = self.folder / (name + '.sha256')
        receipt.write_bytes(b'b'*64 + b'  ' + name.encode() + b'\n')
        receipt.chmod(0o600)
        fn = Mock(return_value={'status':'PASS_KEY_ACCESS'})
        c.run_asset('ENV_SUPPLEMENT', SimpleNamespace(verify_archive=fn), self.folder)
        fn.assert_called_once_with(self.folder / name, 'b'*64, 600)
    def test_both_record_formats(self):
        for data in (self.data, self.data.rstrip(b'\n'), b'b'*64+b' *'+self.name.encode()+b'\n'):
            self.assertEqual(len(c.parse_receipt(data, self.name)), 64)
    def test_wrong_name_malformed_multiline(self):
        for data in (b'', b'x'*64+b'  '+self.name.encode(), self.data+self.data,
                     b'a'*64+b'  other.gpg\n', b'a'*64+b'  '+self.name.encode()+b'\r\n'):
            with self.assertRaises(c.Stop): c.parse_receipt(data, self.name)
    def test_unsafe_receipt(self):
        self.receipt.chmod(0o644)
        fn = Mock()
        with self.assertRaises(c.Stop):
            c.run_asset(self.asset, SimpleNamespace(verify_archive=fn), self.folder)
        fn.assert_not_called()
    def test_changed_receipt(self):
        def mutate(*args):
            self.receipt.write_bytes(b'b'*64+b'  '+self.name.encode()+b'\n')
            return {'status':'PASS_KEY_ACCESS'}
        with self.assertRaises(c.Stop):
            c.run_asset(self.asset, SimpleNamespace(verify_archive=mutate), self.folder)
    def test_failed_crypto_not_retried(self):
        fn = Mock(return_value={'status':'BLOCKED','reason':'GNUPG_REPORTED_BAD_PASSPHRASE'})
        report = c.run_asset(self.asset, SimpleNamespace(verify_archive=fn), self.folder)
        self.assertEqual(report['status'], 'BLOCKED')
        self.assertEqual(fn.call_count, 1)
    def test_postgresql_cannot_be_selected(self):
        fn = Mock()
        with self.assertRaises(c.Stop):
            c.run_asset('POSTGRESQL', SimpleNamespace(verify_archive=fn), self.folder)
        fn.assert_not_called()

if __name__ == '__main__': unittest.main()
