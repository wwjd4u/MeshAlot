"""Runner construction tests; PostgreSQL execution is a separate release gate."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("migrate", ROOT / "migrate.py")
migrate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(migrate)


class MigrationTests(unittest.TestCase):
    def setUp(self):
        self.files = migrate.load_migrations(ROOT / "migrations")

    def test_pinned_files_unchanged(self):
        self.assertEqual({m[1]: m[2] for m in self.files}, migrate.TESTED)

    def test_apply_is_single_transaction(self):
        sql = migrate.build_sql(self.files, "meshalot_test", "apply")
        self.assertEqual(sql.splitlines().count("BEGIN;"), 1)
        self.assertEqual(sql.splitlines().count("COMMIT;"), 1)
        self.assertLess(sql.index("pg_advisory_xact_lock"), sql.index("CREATE SCHEMA"))
        self.assertIn("checksum/name mismatch", sql)
        self.assertIn("history has a gap", sql)
        self.assertEqual(sql.count("\\if :pending"), 2)
        self.assertIn("Untracked existing schema", sql)

    def test_adoption_does_not_reexecute_schema(self):
        sql = migrate.build_sql(self.files, "meshalot_test", "adopt-tested")
        self.assertNotIn("CREATE TABLE users", sql)
        self.assertIn("SAVEPOINT core_verification", sql)
        self.assertIn("ROLLBACK TO SAVEPOINT core_verification", sql)
        self.assertIn("adopted_test", sql)

    def test_adoption_refuses_other_database(self):
        with self.assertRaises(ValueError):
            migrate.build_sql(self.files, "meshalot", "adopt-tested")

    def test_adoption_refuses_altered_hash(self):
        altered = list(self.files)
        version, name, _, body = altered[0]
        altered[0] = (version, name, "0" * 64, body)
        with self.assertRaises(ValueError):
            migrate.build_sql(altered, "meshalot_test", "adopt-tested")

    def test_status_does_not_create_persistent_objects(self):
        sql = migrate.build_sql(self.files, "meshalot_test", "status")
        self.assertNotIn("CREATE SCHEMA", sql)
        self.assertNotIn("INSERT INTO meshalot_meta", sql)
        self.assertTrue(sql.endswith("ROLLBACK;\n"))

    def test_invalid_migrations(self):
        for name, body in [
            ("000002_gap.up.sql", "BEGIN;\nSELECT 1;\nCOMMIT;"),
            ("000001_bad.up.sql", "SELECT 1;"),
            ("000001_bad.up.sql", "BEGIN;\n\\! echo unsafe\nCOMMIT;"),
            ("000001_bad.up.sql", "BEGIN;\nROLLBACK;\nCOMMIT;"),
        ]:
            with self.subTest(name=name, body=body), tempfile.TemporaryDirectory() as folder:
                Path(folder, name).write_text(body)
                with self.assertRaises(ValueError):
                    migrate.load_migrations(Path(folder))

    def test_psql_failure_is_propagated(self):
        with patch.object(migrate.subprocess, "run") as run:
            run.return_value.returncode = 3
            self.assertEqual(migrate.main(["apply", "--database", "meshalot_test"]), 3)
            self.assertIn("--set=ON_ERROR_STOP=1", run.call_args.args[0])
            self.assertNotIn("shell", run.call_args.kwargs)


if __name__ == "__main__":
    unittest.main()
