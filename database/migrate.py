#!/usr/bin/env python3
"""Small psql-backed migration runner; Python standard library only.

Run as the database migration owner, never as the runtime application role.
Migration SQL is trusted repository code. This is not a SQL sandbox.
"""
import argparse
import hashlib
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parent
TESTED = {
    "000001_core.up.sql": "3b7a8cf883d1997fe4e9027c4a909bb10ee14fc72a2055f9ab42c9401f0bc8fc",
    "000002_audit_and_ledger.up.sql": "e4327b9223dfe4719b1d06bc11477a6359b0b6a360cae7bc9bdbc0f0bc1b56ed",
}


def literal(value):
    return "'" + value.replace("'", "''") + "'"


def load_migrations(directory):
    migrations = []
    for path in sorted(directory.glob("*.up.sql")):
        match = re.fullmatch(r"([0-9]{6})_[a-z0-9_]+\.up\.sql", path.name)
        if not match:
            raise ValueError("Invalid migration filename: " + path.name)
        raw = path.read_bytes()
        lines = raw.decode("utf-8").strip().splitlines()
        if len(lines) < 3 or lines[0].strip() != "BEGIN;" or lines[-1].strip() != "COMMIT;":
            raise ValueError("Migration must have a single outer BEGIN/COMMIT: " + path.name)
        body = "\n".join(lines[1:-1])
        if any(line.lstrip().startswith("\\") for line in lines):
            raise ValueError("psql commands are not permitted in migrations")
        # PL/pgSQL BEGIN/END inside functions are allowed. New migrations must
        # not contain transaction-control SQL or nontransactional DDL.
        if re.search(r"(?im)^\s*(COMMIT|ROLLBACK|START\s+TRANSACTION)\b", body):
            raise ValueError("Nested transaction control is not permitted")
        migrations.append((int(match[1]), path.name, hashlib.sha256(raw).hexdigest(), body))
    versions = [m[0] for m in migrations]
    if not migrations or versions != list(range(1, len(migrations) + 1)):
        raise ValueError("Migration versions must be unique and contiguous from 000001")
    return migrations


def adoption_test_body():
    lines = (ROOT / "tests/core.sql").read_text().splitlines()
    if lines.count("BEGIN;") != 1 or lines.count("ROLLBACK;") != 1:
        raise ValueError("Unexpected core test transaction structure")
    return "\n".join(line for line in lines
                     if line not in ("BEGIN;", "ROLLBACK;") and not line.startswith("\\echo"))


def build_sql(migrations, database, action):
    if action == "adopt-tested":
        if database != "meshalot_test":
            raise ValueError("Adoption is restricted to meshalot_test")
        if {m[1]: m[2] for m in migrations} != TESTED:
            raise ValueError("Adoption requires the exact two previously tested migration files")

    values = ",\n".join(f"({v}, {literal(name)}, {literal(digest)})"
                         for v, name, digest, _ in migrations)
    sql = ["\\set ON_ERROR_STOP on", "BEGIN;",
           "SET LOCAL standard_conforming_strings = on;",
           "SET LOCAL search_path = pg_catalog, public;",
           "SET LOCAL lock_timeout = '15s';",
           "SELECT pg_advisory_xact_lock(743820041);",
           f"DO $$ BEGIN IF current_database() <> {literal(database)} THEN "
           "RAISE EXCEPTION 'Wrong database'; END IF; END $$;"]
    if action == "status":
        sql += ["SELECT to_regclass('meshalot_meta.schema_migrations') IS NOT NULL AS tracked \\gset",
                "\\if :tracked"]
    else:
        sql += ["CREATE SCHEMA IF NOT EXISTS meshalot_meta;",
                "REVOKE ALL ON SCHEMA meshalot_meta FROM PUBLIC;",
                "CREATE TABLE IF NOT EXISTS meshalot_meta.schema_migrations ("
                "version integer PRIMARY KEY, filename text UNIQUE NOT NULL, "
                "sha256 text NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'), "
                "registered_at timestamptz NOT NULL DEFAULT now(), "
                "registration text NOT NULL CHECK (registration IN ('executed','adopted_test')));",
                "REVOKE ALL ON meshalot_meta.schema_migrations FROM PUBLIC;"]
    sql += ["CREATE TEMP TABLE expected_migrations(version integer, filename text, sha256 text) ON COMMIT DROP;",
            "INSERT INTO expected_migrations VALUES " + values + ";",
            """DO $$ BEGIN
    IF EXISTS (
        SELECT FROM meshalot_meta.schema_migrations h
        LEFT JOIN expected_migrations e USING (version)
        WHERE e.version IS NULL OR h.filename <> e.filename OR h.sha256 <> e.sha256
    ) THEN
        RAISE EXCEPTION 'Migration checksum/name mismatch or applied file missing; do not edit history';
    END IF;
    IF EXISTS (
        SELECT FROM expected_migrations e
        WHERE e.version <= (SELECT max(version) FROM meshalot_meta.schema_migrations)
          AND NOT EXISTS (SELECT FROM meshalot_meta.schema_migrations h WHERE h.version=e.version)
    ) THEN
        RAISE EXCEPTION 'Migration history has a gap';
    END IF;
END $$;"""]
    if action == "apply":
        sql += ["""DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM meshalot_meta.schema_migrations)
       AND EXISTS (SELECT FROM pg_tables WHERE schemaname='public') THEN
        RAISE EXCEPTION 'Untracked existing schema: refusing to apply; inspect before explicit test adoption';
    END IF;
END $$;"""]
        for version, name, digest, body in migrations:
            sql += [f"SELECT NOT EXISTS (SELECT FROM meshalot_meta.schema_migrations WHERE version={version}) AS pending \\gset",
                    "\\if :pending", f"\\echo Applying {name}",
                    "SET LOCAL search_path = public, pg_catalog;", body,
                    "INSERT INTO meshalot_meta.schema_migrations(version, filename, sha256, registration) VALUES "
                    f"({version},{literal(name)},{literal(digest)},'executed');",
                    "\\else", f"\\echo Verified; skipping {name}", "\\endif"]
    elif action == "adopt-tested":
        sql += ["SELECT NOT EXISTS (SELECT FROM meshalot_meta.schema_migrations) AS untracked \\gset",
                "\\if :untracked",
                "-- Admission is only for the empty, previously verified test schema.",
                """DO $$ DECLARE t record; populated boolean; BEGIN
    IF (SELECT count(*) FROM pg_tables WHERE schemaname='public') <> 13 THEN
        RAISE EXCEPTION 'Expected exactly 13 core tables';
    END IF;
    FOR t IN SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename LOOP
        EXECUTE format('LOCK TABLE public.%I IN ACCESS EXCLUSIVE MODE', t.tablename);
        EXECUTE format('SELECT EXISTS(SELECT FROM public.%I)', t.tablename) INTO populated;
        IF populated THEN RAISE EXCEPTION 'Adoption requires empty test tables'; END IF;
    END LOOP;
END $$;""",
                "SAVEPOINT core_verification;", adoption_test_body(),
                "ROLLBACK TO SAVEPOINT core_verification;",
                "INSERT INTO meshalot_meta.schema_migrations(version, filename, sha256, registration) "
                "SELECT version, filename, sha256, 'adopted_test' FROM expected_migrations;",
                "\\echo Existing test schema verified and registered; migrations NOT re-executed",
                "\\else", "\\echo Already tracked; history checks passed", "\\endif"]

    sql += ["SELECT e.version, e.filename, h.sha256, h.registration, h.registered_at, "
            "CASE WHEN h.version IS NULL THEN 'pending' ELSE 'verified' END AS state "
            "FROM expected_migrations e LEFT JOIN meshalot_meta.schema_migrations h USING(version) ORDER BY e.version;"]
    if action == "status":
        sql += ["\\else", "\\echo UNTRACKED: no migration history exists", "\\endif", "ROLLBACK;"]
    else:
        sql += ["COMMIT;"]
    return "\n".join(sql) + "\n"


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("apply", "status", "adopt-tested"))
    parser.add_argument("--database", required=True)
    args = parser.parse_args(argv)
    if not re.fullmatch(r"meshalot(?:_[a-z0-9]+)*", args.database):
        parser.error("Use an explicit MeshAlot database name, not a connection string")
    try:
        migrations = load_migrations(ROOT / "migrations")
        sql = build_sql(migrations, args.database, args.action)
        result = subprocess.run(
            ["psql", "-X", "--no-password", "--dbname=" + args.database,
             "--set=ON_ERROR_STOP=1", "--file=-"], input=sql, text=True, check=False)
        if result.returncode:
            print("FAILED: inspect database status before retrying; do not alter migration history.", file=sys.stderr)
        return result.returncode
    except (OSError, ValueError) as error:
        print("Migration runner: " + str(error), file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
