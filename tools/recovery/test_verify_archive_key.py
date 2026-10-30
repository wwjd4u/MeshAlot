"""New crypto-wrapper checks using only temporary, synthetic encrypted material."""
import hashlib
import json
import os
from pathlib import Path
import pty
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

import verify_archive_key as verifier


KEY = b"synthetic-key-not-an-operator-credential"
PAYLOAD = b"SYNTHETIC_PLAINTEXT_MUST_NEVER_APPEAR_IN_A_REPORT\n" * 200


class ArchiveKeyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not (shutil.which("gpg") and shutil.which("gpgconf")):
            raise RuntimeError("GnuPG tools are required; do not silently skip crypto acceptance")
        cls.directory = tempfile.TemporaryDirectory(prefix="meshalot-key-tests-")
        cls.root = Path(cls.directory.name)
        cls.encrypt_home = cls.root / "encryption-home"
        cls.encrypt_home.mkdir(mode=0o700)
        cls.gpg = shutil.which("gpg")
        cls.gpgconf = shutil.which("gpgconf")
        try:
            cls.base = cls.encrypt("base.gpg")
            cls.aes128 = cls.encrypt("aes128.gpg", cipher="AES128")
            cls.aead = cls.encrypt("aead.gpg", aead=True)
        except BaseException:
            verifier._cleanup(cls.gpgconf, cls.encrypt_home)
            cls.directory.cleanup()
            raise

    @classmethod
    def encrypt(cls, name, cipher="AES256", aead=False):
        destination = cls.root / name
        read_fd, write_fd = os.pipe()
        try:
            os.write(write_fd, KEY + b"\n")
            os.close(write_fd)
            write_fd = None
            command = [cls.gpg, "--no-options", "--homedir", str(cls.encrypt_home),
                       "--batch", "--yes", "--no-symkey-cache", "--pinentry-mode", "loopback",
                       "--passphrase-fd", str(read_fd), "--cipher-algo", cipher,
                       "--set-filename", "PRIVATE_SYNTHETIC_LITERAL_FILENAME",
                       "--output", str(destination), "--symmetric"]
            if aead:
                command.insert(-1, "--force-aead")
            subprocess.run(command, input=PAYLOAD, stdout=subprocess.DEVNULL,
                           stderr=subprocess.DEVNULL, pass_fds=(read_fd,),
                           timeout=10, check=True)
        finally:
            os.close(read_fd)
            if write_fd is not None:
                os.close(write_fd)
        destination.chmod(0o600)
        return destination

    @classmethod
    def tearDownClass(cls):
        try:
            if not verifier._cleanup(cls.gpgconf, cls.encrypt_home):
                raise RuntimeError("Synthetic fixture-agent cleanup was not verified")
        finally:
            cls.directory.cleanup()

    def setUp(self):
        self.case = tempfile.TemporaryDirectory(prefix="case-", dir=self.root)
        self.folder = Path(self.case.name)
        self.archive = self.folder / "PRIVATE_SYNTHETIC_ARCHIVE_PATH.gpg"
        shutil.copyfile(self.base, self.archive)
        self.archive.chmod(0o600)
        self.homes = []
        self.pinentry_count = self.folder / "pinentry-count"

    def tearDown(self):
        for home in self.homes:
            self.assertFalse(home.exists(), "Private GnuPG home was not removed")
        self.case.cleanup()

    def digest(self):
        return hashlib.sha256(self.archive.read_bytes()).hexdigest()

    def snapshot(self):
        info = self.archive.stat()
        return self.archive.read_bytes(), info.st_mode, info.st_mtime_ns, info.st_ctime_ns

    def configure_pinentry(self, key=KEY.decode(), delay=0, mutate_source=False):
        executable = self.folder / "synthetic-pinentry"
        executable.write_text(
            "#!" + sys.executable + "\n"
            "import pathlib, sys, time\n"
            "print('OK synthetic-only pinentry', flush=True)\n"
            "for line in sys.stdin:\n"
            "    command = line.strip().split(' ', 1)[0]\n"
            "    if command == 'GETPIN':\n"
            "        with pathlib.Path(" + repr(str(self.pinentry_count)) + ").open('a') as log:\n"
            "            log.write('prompt\\n')\n"
            "        time.sleep(" + repr(delay) + ")\n"
            "        if " + repr(mutate_source) + ":\n"
            "            target = pathlib.Path(" + repr(str(self.archive)) + ")\n"
            "            target.write_bytes(target.read_bytes() + b'changed-during-prompt')\n"
            "        print('D ' + " + repr(key) + ", flush=True)\n"
            "    print('OK', flush=True)\n"
            "    if command == 'BYE':\n"
            "        break\n")
        executable.chmod(0o700)
        original = verifier._prepare_private_home

        def configure(home):
            self.homes.append(home)
            original(home)
            with (home / "gpg-agent.conf").open("a") as config:
                config.write("pinentry-program " + str(executable) + "\n")
        return patch.object(verifier, "_prepare_private_home", configure)

    def assert_report_private(self, report):
        encoded = json.dumps(report)
        for marker in (KEY.decode(), PAYLOAD[:40].decode(), str(self.archive),
                       "PRIVATE_SYNTHETIC_LITERAL_FILENAME"):
            self.assertNotIn(marker, encoded)
        self.assertEqual(report["gate_acceptance"], "NOT_ASSESSED")
        self.assertEqual(report["restore_acceptance"], "NOT_TESTED")
        self.assertEqual(report["vault_source"], "OPERATOR_ATTESTATION_REQUIRED")

    def test_valid_key_authenticates_without_plaintext_or_source_changes(self):
        before = self.snapshot()
        with self.configure_pinentry():
            report = verifier.verify_archive(self.archive, self.digest(), 5)
        self.assertEqual(report["status"], "PASS_KEY_ACCESS", report)
        self.assertIs(report["source_unchanged"], True)
        self.assertEqual(self.snapshot(), before)
        self.assertEqual(self.pinentry_count.read_text().splitlines(), ["prompt"])
        self.assert_report_private(report)

    def test_wrong_key_does_not_reuse_a_previous_success_or_retry(self):
        with self.configure_pinentry():
            good = verifier.verify_archive(self.archive, self.digest(), 5)
        self.assertEqual(good["status"], "PASS_KEY_ACCESS", good)
        with self.configure_pinentry(key="SYNTHETIC_WRONG_KEY"):
            bad = verifier.verify_archive(self.archive, self.digest(), 5)
        self.assertEqual(bad["status"], "BLOCKED", bad)
        self.assertEqual(bad["reason"], "DECRYPTION_FAILED")
        self.assertIs(bad["source_unchanged"], True)
        self.assertEqual(self.pinentry_count.read_text().splitlines(), ["prompt", "prompt"])
        self.assert_report_private(bad)

    def test_aead_authenticated_archive_is_supported(self):
        shutil.copyfile(self.aead, self.archive)
        with self.configure_pinentry():
            report = verifier.verify_archive(self.archive, self.digest(), 5)
        self.assertEqual(report["status"], "PASS_KEY_ACCESS", report)
        self.assert_report_private(report)

    def test_ciphertext_authentication_failure_is_not_a_pass(self):
        data = bytearray(self.archive.read_bytes())
        data[-10] ^= 1
        self.archive.write_bytes(data)
        before = self.snapshot()
        with self.configure_pinentry():
            report = verifier.verify_archive(self.archive, self.digest(), 5)
        self.assertEqual(report["status"], "BLOCKED", report)
        self.assertEqual(self.snapshot(), before)
        self.assert_report_private(report)

    def test_wrong_receipt_stops_before_any_key_prompt(self):
        with patch.object(verifier, "_prepare_private_home") as configure:
            report = verifier.verify_archive(self.archive, "0" * 64, 5)
        self.assertEqual(report["reason"], "SOURCE_DIGEST_MISMATCH")
        configure.assert_not_called()
        self.assert_report_private(report)

    def test_unapproved_cipher_does_not_pass_after_decryption(self):
        shutil.copyfile(self.aes128, self.archive)
        with self.configure_pinentry():
            report = verifier.verify_archive(self.archive, self.digest(), 5)
        self.assertEqual(report["reason"], "AUTHENTICATED_AES256_NOT_ESTABLISHED")
        self.assertIs(report["source_unchanged"], True)

    def test_timeout_cleans_up_without_changing_the_archive(self):
        before = self.snapshot()
        with self.configure_pinentry(delay=2):
            report = verifier.verify_archive(self.archive, self.digest(), 0.15)
        self.assertEqual(report["reason"], "DECRYPTION_TIMEOUT", report)
        self.assertEqual(self.snapshot(), before)
        self.assert_report_private(report)

    def test_source_types_links_and_write_modes_are_rejected(self):
        pin = self.digest()
        link = self.folder / "symlink"
        link.symlink_to(self.archive)
        hardlink = self.folder / "hardlink"
        os.link(self.archive, hardlink)
        with patch.object(verifier, "_prepare_private_home") as configure:
            for target in (link, hardlink):
                with self.subTest(target=target.name):
                    self.assertEqual(verifier.verify_archive(target, pin, 5)["status"], "BLOCKED")
            hardlink.unlink()
            self.archive.chmod(0o622)
            self.assertEqual(verifier.verify_archive(self.archive, pin, 5)["status"], "BLOCKED")
            configure.assert_not_called()

    def test_invalid_pins_and_unbounded_time_limits_are_rejected(self):
        with patch.object(verifier, "_prepare_private_home") as configure:
            for pin, seconds in (("bad", 5), (self.digest(), float("nan")),
                                 (self.digest(), float("inf")), (self.digest(), 0)):
                with self.subTest(pin_valid=len(pin) == 64, seconds=seconds):
                    self.assertEqual(verifier.verify_archive(self.archive, pin, seconds)["status"], "BLOCKED")
            configure.assert_not_called()

    def test_cli_refuses_headless_real_archive_use(self):
        result = subprocess.run(
            [sys.executable, "-B", str(Path(verifier.__file__)), "--archive", str(self.archive),
             "--expected-sha256", self.digest()],
            stdin=subprocess.DEVNULL, capture_output=True, text=True, timeout=5)
        self.assertEqual(result.returncode, 1)
        report = json.loads(result.stdout)
        self.assertEqual(report["reason"], "PRIVATE_INTERACTIVE_TERMINAL_REQUIRED")
        self.assertEqual(result.stderr, "")
        self.assert_report_private(report)

    def test_complete_cli_with_private_terminal_and_synthetic_pinentry(self):
        with self.configure_pinentry():
            pinentry = self.folder / "synthetic-pinentry"
            code = (
                "import sys\n"
                "sys.path.insert(0, " + repr(str(Path(verifier.__file__).parent)) + ")\n"
                "import verify_archive_key as app\n"
                "original = app._prepare_private_home\n"
                "def prepare(home):\n"
                "    original(home)\n"
                "    with (home / 'gpg-agent.conf').open('a') as config:\n"
                "        config.write(" + repr("pinentry-program " + str(pinentry) + "\n") + ")\n"
                "app._prepare_private_home = prepare\n"
                "raise SystemExit(app.main(" + repr([
                    "--archive", str(self.archive), "--expected-sha256", self.digest(),
                    "--timeout", "5"]) + "))\n")
            master, slave = pty.openpty()
            try:
                result = subprocess.run([sys.executable, "-B", "-c", code],
                                        stdin=slave, capture_output=True, text=True, timeout=10)
            finally:
                os.close(master)
                os.close(slave)
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertEqual(result.stderr, "")
        report = json.loads(result.stdout)
        self.assertEqual(report["status"], "PASS_KEY_ACCESS")
        self.assert_report_private(report)

    def test_metadata_output_limit_fails_closed(self):
        fake_gpg = self.folder / "gpg"
        options = ["--no-symkey-cache", "--pinentry-mode", "--status-fd",
                   "--no-auto-key-retrieve", "--no-auto-key-import"]
        fake_gpg.write_text(
            "#!" + sys.executable + "\nimport os, sys\n"
            "if '--dump-options' in sys.argv:\n"
            "    print(" + repr("\n".join(options)) + ")\n"
            "else:\n"
            "    os.write(1, b'x' * 131072)\n")
        fake_gpg.chmod(0o700)
        before = self.snapshot()
        with patch.dict(os.environ, {"PATH": str(self.folder) + os.pathsep + os.environ["PATH"]}):
            report = verifier.verify_archive(self.archive, self.digest(), 5)
        self.assertEqual(report["reason"], "STATUS_LIMIT_REACHED", report)
        self.assertEqual(self.snapshot(), before)
        self.assert_report_private(report)

    def test_source_change_during_key_entry_is_rejected(self):
        pin = self.digest()
        with self.configure_pinentry(mutate_source=True):
            report = verifier.verify_archive(self.archive, pin, 5)
        self.assertEqual(report["reason"], "SOURCE_CHANGED", report)
        self.assertIs(report["source_unchanged"], False)
        self.assert_report_private(report)


if __name__ == "__main__":
    unittest.main()
