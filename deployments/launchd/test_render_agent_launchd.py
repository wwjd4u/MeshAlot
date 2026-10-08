import os
import plistlib
import stat
import tempfile
import unittest
from pathlib import Path

from render_agent_launchd import stage_launch_agent


class LaunchAgentStagingTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="meshalot-launchd-test-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.binary = self.root / "meshalot-agent"
        self.binary.write_bytes(b"fake test executable - never run")
        self.binary.chmod(0o700)
        self.identity = self.root / "existing-identity.json"
        self.identity.write_bytes(b"fake identity fixture - never enrolled")
        self.identity.chmod(0o600)
        self.output = self.root / "staged.plist"

    def stage(self, **overrides):
        opts = {
            "binary": str(self.binary),
            "identity": str(self.identity),
            "server": "https://api.meshalot.com",
            "mode": "away",
            "manual_pause": "true",
            "output": str(self.output),
        }
        opts.update(overrides)
        return stage_launch_agent(**opts)

    def test_stages_valid_login_agent_without_enrollment(self):
        path = self.stage()
        self.assertEqual(path, self.output)
        self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)
        data = plistlib.loads(path.read_bytes())
        self.assertEqual(data["Label"], "com.meshalot.agent")
        self.assertEqual(data["ProgramArguments"][1], "connect")
        self.assertNotIn("enroll", data["ProgramArguments"])
        self.assertEqual(data["ProgramArguments"][2:4], ["--server", "https://api.meshalot.com"])
        self.assertEqual(data["ProgramArguments"][4:6], ["--identity", str(self.identity)])
        self.assertEqual(data["ProgramArguments"][6:10], ["--mode", "away", "--manual-pause", "true"])
        self.assertTrue(data["RunAtLoad"])
        self.assertEqual(data["KeepAlive"], {"SuccessfulExit": False})
        self.assertEqual(data["ThrottleInterval"], 15)
        self.assertEqual(self.identity.read_bytes(), b"fake identity fixture - never enrolled")

    def test_fails_closed_for_missing_or_insecure_identity(self):
        for invalid in [str(self.root / "missing"), "relative/identity"]:
            with self.subTest(identity=invalid):
                with self.assertRaises(ValueError):
                    self.stage(identity=invalid)
                self.assertFalse(self.output.exists())
        self.identity.chmod(0o644)
        with self.assertRaises(ValueError):
            self.stage()
        self.assertFalse(self.output.exists())
        self.identity.chmod(0o600)
        link = self.root / "identity-link"
        link.symlink_to(self.identity)
        with self.assertRaises(ValueError):
            self.stage(identity=str(link))
        self.assertFalse(self.output.exists())

    def test_rejects_insecure_or_malformed_targets(self):
        for target in ["", "http://api.meshalot.com", "ws://localhost:8080",
                       "https://user:pass@api.meshalot.com",
                       "https://api.meshalot.com/path",
                       "https://api.meshalot.com/?token=abc",
                       "https://api.meshalot.com/#a"]:
            with self.subTest(target=target):
                with self.assertRaises(ValueError):
                    self.stage(server=target)
                self.assertFalse(self.output.exists())

    def test_requires_explicit_sharing_configuration(self):
        for mode in ["", "admin", "maximum"]:
            with self.subTest(mode=mode):
                with self.assertRaises(ValueError):
                    self.stage(mode=mode)
        for pause in ["", "yes", "0"]:
            with self.subTest(pause=pause):
                with self.assertRaises(ValueError):
                    self.stage(manual_pause=pause)
        self.assertFalse(self.output.exists())

    def test_refuses_overwrite_and_incorrect_paths(self):
        self.stage()
        first = self.output.read_bytes()
        with self.assertRaises(FileExistsError):
            self.stage()
        self.assertEqual(self.output.read_bytes(), first)
        with self.assertRaises(ValueError):
            self.stage(output="relative.plist")
        self.assertEqual(self.output.read_bytes(), first)

    def test_no_agent_execution_or_launchd_mutation(self):
        # The fake binary is not a functioning program. Rendering must not
        # execute it and must not create a LaunchAgents directory.
        self.stage()
        self.assertFalse((self.root / "Library" / "LaunchAgents").exists())


if __name__ == "__main__":
    unittest.main()
