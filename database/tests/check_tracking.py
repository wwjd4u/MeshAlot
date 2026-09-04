"""Integration checks against the adopted meshalot_test database only.

No database drops, real-file edits, or history rewrites. Failure probes use
transactions which must roll back. Run with the migration administrator role.
"""
import importlib.util
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("migrate", ROOT / "migrate.py")
migrate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(migrate)


def run(sql, expected_error=None):
    result = subprocess.run(
        ["psql", "-X", "--no-password", "--dbname=meshalot_test", "-At",
         "--set=ON_ERROR_STOP=1", "--file=-"], input=sql, text=True,
        capture_output=True, check=False)
    if expected_error:
        if result.returncode != 3 or expected_error not in result.stderr:
            raise RuntimeError("Expected rejection missing: " + expected_error + "\n" + result.stderr)
    elif result.returncode:
        raise RuntimeError(result.stderr)
    return result.stdout


def main():
    migrations = migrate.load_migrations(ROOT / "migrations")
    snapshot = "SELECT version,filename,sha256,registration,registered_at FROM meshalot_meta.schema_migrations ORDER BY version;"
    before = run(snapshot)
    if len(before.strip().splitlines()) != len(migrations):
        raise RuntimeError("Run explicit test adoption first")
    for _ in range(2):
        run(migrate.build_sql(migrations, "meshalot_test", "apply"))
    if run(snapshot) != before:
        raise RuntimeError("Repeat apply changed migration history")
    print("PASS: repeat application skips recorded migrations")

    altered = list(migrations)
    version, name, _, body = altered[0]
    altered[0] = (version, name, "0" * 64, body)
    run(migrate.build_sql(altered, "meshalot_test", "apply"), "checksum/name mismatch")
    print("PASS: altered checksum rejected without changing migration files")

    pending = migrations + [(len(migrations)+1, "999999_failure_probe.up.sql", "f"*64,
                             "CREATE TABLE meshalot_meta.rollback_probe(id integer);\nSELECT 1/0;")]
    run(migrate.build_sql(pending, "meshalot_test", "apply"), "division by zero")
    if run("SELECT to_regclass('meshalot_meta.rollback_probe') IS NULL;").strip() != "t":
        raise RuntimeError("Failed migration left a schema object behind")
    if run(snapshot) != before:
        raise RuntimeError("Failure probe changed history")
    print("PASS: failed migration rolled back schema changes and history")
    print("MESHALOT MIGRATION TRACKING TESTS PASSED")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, OSError, ValueError) as error:
        print("FAILED: " + str(error), file=sys.stderr)
        sys.exit(1)
