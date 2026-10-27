import copy
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

import check_bundle as checker


SOURCE = "4f67ee54e5ebca084e38e306e723ae0acd08e77d"


class BundleTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="meshalot-synthetic-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        artifacts = []
        for kind, count in checker.EXPECTED_KINDS.items():
            for index in range(count):
                name = f"{kind}-{index}.fixture"
                data = f"synthetic fixture: {name}\n".encode()
                (self.root / name).write_bytes(data)
                artifacts.append({"kind": kind, "path": name,
                                  "sha256": hashlib.sha256(data).hexdigest()})
        self.manifest = {"format_version": 1, "material": "synthetic_only",
                         "purpose": "m13_acceptance", "source_commit": SOURCE,
                         "postgres_major": 16, "schema_version": 10,
                         "runtime_grants_schema_version": 10, "artifacts": artifacts}
        self.save()

    def save(self):
        (self.root / "manifest.json").write_text(json.dumps(self.manifest))

    def run_check(self):
        self.save()
        return self.check_current()

    def manifest_digest(self):
        return hashlib.sha256((self.root / "manifest.json").read_bytes()).hexdigest()

    def check_current(self):
        return checker.check(self.root, "manifest.json", SOURCE, self.manifest_digest())

    def test_valid_bundle_is_only_ready_for_review_and_inputs_unchanged(self):
        before = {p.name: (p.read_bytes(), p.stat().st_mode, p.stat().st_mtime_ns)
                  for p in self.root.iterdir()}
        result = self.check_current()
        self.assertEqual(result["status"], "READY_FOR_REVIEW")
        self.assertEqual(result["runtime_acceptance"], "NOT_TESTED")
        self.assertEqual(result["gate_acceptance"], "NOT_ASSESSED")
        self.assertEqual(before, {p.name: (p.read_bytes(), p.stat().st_mode, p.stat().st_mtime_ns)
                                  for p in self.root.iterdir()})

    def test_schema_source_scope_and_engine_rejections(self):
        original = copy.deepcopy(self.manifest)
        cases = [("source_commit", "0" * 40, "SOURCE_PIN_MISMATCH"),
                 ("postgres_major", 17, "POSTGRES_MAJOR_MISMATCH"),
                 ("schema_version", 8, "SCHEMA_PURPOSE_MISMATCH"),
                 ("runtime_grants_schema_version", 8, "GRANTS_SCHEMA_MISMATCH"),
                 ("material", "production", "MATERIAL_SCOPE_NOT_ALLOWED"),
                 ("purpose", [], "INVALID_PURPOSE"),
                 ("format_version", True, "UNSUPPORTED_FORMAT")]
        for field, value, reason in cases:
            with self.subTest(field=field):
                self.manifest = copy.deepcopy(original)
                self.manifest[field] = value
                with self.assertRaisesRegex(checker.Blocked, reason):
                    self.run_check()

    def test_v8_recovery_requires_v8_grants(self):
        self.manifest.update(purpose="baseline_recovery", schema_version=8)
        with self.assertRaisesRegex(checker.Blocked, "GRANTS_SCHEMA_MISMATCH"):
            self.run_check()
        self.manifest["runtime_grants_schema_version"] = 8
        self.assertEqual(self.run_check()["status"], "READY_FOR_REVIEW")

    def test_modified_artifact(self):
        (self.root / self.manifest["artifacts"][0]["path"]).write_text("changed")
        with self.assertRaisesRegex(checker.Blocked, "ARTIFACT_DIGEST_MISMATCH"):
            self.run_check()

    def test_manifest_changes_require_independent_review(self):
        pin = self.manifest_digest()
        self.manifest["artifacts"][0]["sha256"] = "0" * 64
        self.save()
        with self.assertRaisesRegex(checker.Blocked, "MANIFEST_DIGEST_MISMATCH"):
            checker.check(self.root, "manifest.json", SOURCE, pin)

    def test_missing_member_and_incorrect_kind_count(self):
        self.manifest["artifacts"][0]["kind"] = "systemd_drop_in"
        with self.assertRaisesRegex(checker.Blocked, "INCOMPLETE_BUNDLE"):
            self.run_check()
        self.manifest["artifacts"].pop()
        with self.assertRaisesRegex(checker.Blocked, "INVALID_ARTIFACT_COUNT"):
            self.run_check()

    def test_path_escape_and_duplicate_paths(self):
        for name in ("../outside", "/etc/passwd", "a/../outside", "a//b", "./a", "a\\b", "a\nsecret"):
            with self.subTest(name=name), self.assertRaisesRegex(checker.Blocked, "UNSAFE_PATH"):
                checker.safe_parts(name)
        self.manifest["artifacts"][1]["path"] = self.manifest["artifacts"][0]["path"]
        with self.assertRaisesRegex(checker.Blocked, "DUPLICATE_ARTIFACT_PATH"):
            self.run_check()

    def test_symlinked_file_directory_and_root(self):
        target = self.root / self.manifest["artifacts"][0]["path"]
        target.unlink()
        target.symlink_to("manifest.json")
        with self.assertRaises(OSError):
            self.run_check()
        target.unlink()
        (self.root / "linked").symlink_to(self.root, target_is_directory=True)
        fd = os.open(self.root, os.O_RDONLY | os.O_DIRECTORY)
        try:
            with self.assertRaises(OSError):
                checker.read_member(fd, "linked/manifest.json", 65536)
        finally:
            os.close(fd)
        with self.assertRaises(OSError):
            checker.check(self.root / "linked", "manifest.json", SOURCE, self.manifest_digest())

    def test_hardlink_and_world_writable_file(self):
        target = self.root / self.manifest["artifacts"][0]["path"]
        os.link(target, self.root / "alias")
        with self.assertRaisesRegex(checker.Blocked, "HARDLINK_NOT_ALLOWED"):
            self.run_check()
        (self.root / "alias").unlink()
        target.chmod(0o666)
        with self.assertRaisesRegex(checker.Blocked, "UNSAFE_OWNERSHIP_OR_MODE"):
            self.run_check()

    def test_bad_json_and_unknown_fields(self):
        for raw in ('{"purpose":1,"purpose":2}', "{invalid", '"text"', '[' * 1500):
            (self.root / "manifest.json").write_text(raw)
            with self.subTest(raw=raw[:20]), self.assertRaises(checker.Blocked):
                self.check_current()
        self.manifest["secret"] = "must never be reported"
        with self.assertRaisesRegex(checker.Blocked, "INVALID_MANIFEST_FIELDS"):
            self.run_check()

    def test_cli_exit_codes_and_sanitized_errors(self):
        command = [sys.executable, str(Path(checker.__file__)), "--root", str(self.root),
                   "--expected-source", SOURCE, "--expected-manifest", self.manifest_digest()]
        passed = subprocess.run(command, capture_output=True, text=True, timeout=5)
        self.assertEqual(passed.returncode, 0, passed.stderr)
        self.assertEqual(json.loads(passed.stdout)["runtime_acceptance"], "NOT_TESTED")
        (self.root / self.manifest["artifacts"][0]["path"]).unlink()
        blocked = subprocess.run(command, capture_output=True, text=True, timeout=5)
        self.assertEqual(blocked.returncode, 1)
        self.assertEqual(json.loads(blocked.stdout)["status"], "BLOCKED")
        self.assertNotIn(str(self.root), blocked.stdout + blocked.stderr)

    def test_oversized_manifest_and_artifact_limits(self):
        (self.root / "manifest.json").write_bytes(b" " * (checker.MAX_MANIFEST + 1))
        with self.assertRaisesRegex(checker.Blocked, "FILE_TOO_LARGE"):
            self.check_current()
        self.save()
        fd = os.open(self.root, os.O_RDONLY | os.O_DIRECTORY)
        try:
            with self.assertRaisesRegex(checker.Blocked, "FILE_TOO_LARGE"):
                checker.read_member(fd, self.manifest["artifacts"][0]["path"], 1)
        finally:
            os.close(fd)


if __name__ == "__main__":
    unittest.main()
