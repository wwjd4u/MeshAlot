"""New fixture construction and pin validation only; not the accepted GPG suite."""
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
import check_mac_public_paste as c

class PublicPasteTests(unittest.TestCase):
    def test_pinned_loader(self):
        v = c.load_verifier(Path(c.__file__).with_name('verify_archive_key.py'))
        self.assertTrue(callable(v.verify_archive))
    def test_wrong_source_refused(self):
        with tempfile.TemporaryDirectory() as temp:
            p = Path(temp) / 'wrong.py'
            p.write_text('raise RuntimeError("must not execute")')
            with self.assertRaises(c.Stop): c.load_verifier(p)
    def test_public_fixture_actual_roundtrip(self):
        gpg, conf = shutil.which('gpg'), shutil.which('gpgconf')
        self.assertTrue(gpg and conf)
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            archive, digest = c.generate(root, gpg, conf)
            self.assertEqual(len(digest), 64)
            home = root / 'independent-home'
            home.mkdir(mode=0o700)
            try:
                result = subprocess.run([gpg, '--no-options', '--homedir', str(home),
                    '--batch', '--pinentry-mode', 'loopback', '--passphrase-fd', '0',
                    '--decrypt', str(archive)], input=c.PUBLIC_VALUE+b'\n',
                    capture_output=True, timeout=10)
                self.assertEqual(result.returncode, 0)
                self.assertEqual(result.stdout, c.PUBLIC_DATA)
            finally:
                done = subprocess.run([conf, '--homedir', str(home), '--kill', 'all'],
                                      capture_output=True, timeout=5)
                self.assertEqual(done.returncode, 0)
    def test_generation_failure_cleans_agent(self):
        with tempfile.TemporaryDirectory() as temp:
            failed = subprocess.CompletedProcess([], 1)
            passed = subprocess.CompletedProcess([], 0)
            with patch.object(c.subprocess, 'run', side_effect=[failed, passed]) as run:
                with self.assertRaises(c.Stop): c.generate(Path(temp), 'gpg', 'gpgconf')
                self.assertIn('--kill', run.call_args_list[-1].args[0])
    def test_refused_cloud_cli(self):
        result = subprocess.run([sys.executable, '-B', c.__file__], capture_output=True, timeout=3)
        self.assertEqual(result.returncode, 1)
        self.assertIn(b'PUBLIC_PASTE_SETUP_OR_CLEANUP_STOP', result.stdout)
        self.assertEqual(result.stderr, b'')

if __name__ == '__main__': unittest.main()
