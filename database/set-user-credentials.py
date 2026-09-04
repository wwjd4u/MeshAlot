#!/usr/bin/env python3
"""Set login credentials for an existing MeshAlot user without echoing the password."""
import argparse
import getpass
import re
import subprocess
import sys


def sql_text(value: str) -> str:
    return "convert_from(decode('%s','hex'),'UTF8')" % value.encode("utf-8").hex()


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--database", required=True)
    parser.add_argument("--user-id", required=True)
    parser.add_argument("--stdin", action="store_true", help="read email and password as two stdin lines")
    args = parser.parse_args(argv)
    if not re.fullmatch(r"meshalot(?:_[a-z0-9]+)*", args.database):
        parser.error("Use an explicit MeshAlot database name")
    if not re.fullmatch(r"[0-9a-fA-F-]{36}", args.user_id):
        parser.error("Invalid user ID")
    if args.stdin:
        email = sys.stdin.readline().rstrip("\n")
        password = sys.stdin.readline().rstrip("\n")
    else:
        email = input("Email: ").strip()
        password = getpass.getpass("Password: ")
    email = email.strip().lower()
    if not email or "@" not in email or len(email) > 320:
        raise SystemExit("Invalid email")
    if len(password) < 12 or len(password) > 1024:
        raise SystemExit("Password must be between 12 and 1024 characters")
    sql = f"""\\set ON_ERROR_STOP on
BEGIN;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE id='{args.user_id}'::uuid) THEN
        RAISE EXCEPTION 'MeshAlot user not found';
    END IF;
END $$;
UPDATE users
SET email={sql_text(email)},
    password_hash=crypt({sql_text(password)},gen_salt('bf',12))
WHERE id='{args.user_id}'::uuid
RETURNING email;
COMMIT;
"""
    result = subprocess.run(
        ["psql", "-X", "--no-password", "--dbname=" + args.database, "--set=ON_ERROR_STOP=1", "--file=-"],
        input=sql, text=True, check=False)
    if result.returncode:
        return result.returncode
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
