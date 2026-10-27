#!/usr/bin/env python3
"""Offline integrity preflight for explicitly synthetic G17 recovery bundles.

Never executes artifacts, queries a database, extracts archives, or closes a gate.
"""
import argparse
from collections import Counter
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import sys


MAX_MANIFEST = 65536
MAX_ARTIFACT = 64 * 1024 * 1024
EXPECTED_KINDS = Counter(api_binary=1, systemd_unit=1, systemd_drop_in=6, caddy_config=1)
FIELDS = {"format_version", "material", "purpose", "source_commit", "postgres_major",
          "schema_version", "runtime_grants_schema_version", "artifacts"}


class Blocked(Exception):
    pass


def require(condition, code):
    if not condition:
        raise Blocked(code)


def unique_object(pairs):
    obj = {}
    for key, value in pairs:
        require(key not in obj, "DUPLICATE_JSON_KEY")
        obj[key] = value
    return obj


def safe_parts(value):
    require(isinstance(value, str) and bool(value), "INVALID_PATH")
    path = PurePosixPath(value)
    require(not path.is_absolute() and "\\" not in value and
            all(p not in ("", ".", "..") for p in value.split("/")) and
            not any(ord(c) < 32 or ord(c) == 127 for c in value), "UNSAFE_PATH")
    return path.parts


def checked_metadata(info, directory=False):
    require((stat.S_ISDIR(info.st_mode) if directory else stat.S_ISREG(info.st_mode)),
            "UNSAFE_FILE_TYPE")
    require(info.st_uid == os.geteuid() and not info.st_mode & 0o022,
            "UNSAFE_OWNERSHIP_OR_MODE")
    if not directory:
        require(info.st_nlink == 1, "HARDLINK_NOT_ALLOWED")


def read_member(root_fd, name, limit, collect=False):
    """Open each component relative to a pinned directory; never follow links."""
    parts = safe_parts(name)
    current = os.dup(root_fd)
    try:
        for part in parts[:-1]:
            child = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW,
                            dir_fd=current)
            os.close(current)
            current = child
            checked_metadata(os.fstat(current), directory=True)
        fd = os.open(parts[-1], os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK,
                     dir_fd=current)
        with os.fdopen(fd, "rb") as stream:
            before = os.fstat(stream.fileno())
            checked_metadata(before)
            require(before.st_size <= limit, "FILE_TOO_LARGE")
            digest, data, count = hashlib.sha256(), bytearray(), 0
            while chunk := stream.read(65536):
                count += len(chunk)
                require(count <= limit, "FILE_TOO_LARGE")
                digest.update(chunk)
                if collect:
                    data.extend(chunk)
            after = os.fstat(stream.fileno())
            require((before.st_size, before.st_mtime_ns, before.st_ctime_ns) ==
                    (after.st_size, after.st_mtime_ns, after.st_ctime_ns) and
                    count == before.st_size, "FILE_CHANGED_DURING_READ")
            return digest.hexdigest(), count, bytes(data)
    finally:
        os.close(current)


def check(root, manifest_name, expected_source, expected_manifest):
    require(bool(re.fullmatch(r"[0-9a-f]{40}", expected_source)), "INVALID_SOURCE_PIN")
    require(bool(re.fullmatch(r"[0-9a-f]{64}", expected_manifest)), "INVALID_MANIFEST_PIN")
    root_fd = os.open(root, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        checked_metadata(os.fstat(root_fd), directory=True)
        digest, _, raw = read_member(root_fd, manifest_name, MAX_MANIFEST, collect=True)
        require(digest == expected_manifest, "MANIFEST_DIGEST_MISMATCH")
        try:
            manifest = json.loads(raw, object_pairs_hook=unique_object)
        except (ValueError, UnicodeError, RecursionError):
            raise Blocked("INVALID_JSON") from None
        require(isinstance(manifest, dict) and set(manifest) == FIELDS, "INVALID_MANIFEST_FIELDS")
        require(type(manifest["format_version"]) is int and manifest["format_version"] == 1,
                "UNSUPPORTED_FORMAT")
        require(manifest["material"] == "synthetic_only", "MATERIAL_SCOPE_NOT_ALLOWED")
        require(manifest["source_commit"] == expected_source, "SOURCE_PIN_MISMATCH")
        require(type(manifest["postgres_major"]) is int and manifest["postgres_major"] == 16,
                "POSTGRES_MAJOR_MISMATCH")
        purpose = manifest["purpose"]
        require(isinstance(purpose, str) and purpose in ("baseline_recovery", "m13_acceptance"),
                "INVALID_PURPOSE")
        schema = 8 if purpose == "baseline_recovery" else 10
        require(type(manifest["schema_version"]) is int and manifest["schema_version"] == schema,
                "SCHEMA_PURPOSE_MISMATCH")
        require(type(manifest["runtime_grants_schema_version"]) is int and
                manifest["runtime_grants_schema_version"] == schema, "GRANTS_SCHEMA_MISMATCH")
        artifacts = manifest["artifacts"]
        require(isinstance(artifacts, list) and len(artifacts) == 9, "INVALID_ARTIFACT_COUNT")
        seen, kinds, total = {manifest_name}, Counter(), 0
        for entry in artifacts:
            require(isinstance(entry, dict) and set(entry) == {"kind", "path", "sha256"},
                    "INVALID_ARTIFACT_FIELDS")
            kind, name, digest = entry["kind"], entry["path"], entry["sha256"]
            require(isinstance(kind, str) and kind in EXPECTED_KINDS, "INVALID_ARTIFACT_KIND")
            safe_parts(name)
            require(name not in seen, "DUPLICATE_ARTIFACT_PATH")
            require(isinstance(digest, str) and bool(re.fullmatch(r"[0-9a-f]{64}", digest)),
                    "INVALID_DIGEST")
            actual, count, _ = read_member(root_fd, name, MAX_ARTIFACT)
            require(actual == digest, "ARTIFACT_DIGEST_MISMATCH")
            seen.add(name)
            kinds[kind] += 1
            total += count
        require(kinds == EXPECTED_KINDS, "INCOMPLETE_BUNDLE")
        return {"status": "READY_FOR_REVIEW", "scope": "OFFLINE_SYNTHETIC_INTEGRITY_ONLY",
                "gate_acceptance": "NOT_ASSESSED",
                "runtime_acceptance": "NOT_TESTED", "artifact_count": len(artifacts),
                "bytes_checked": total, "source_commit": expected_source}
    finally:
        os.close(root_fd)


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, required=True, help="owner-controlled synthetic bundle")
    parser.add_argument("--manifest", default="manifest.json", help="relative to root")
    parser.add_argument("--expected-source", required=True, help="independently reviewed immutable SHA")
    parser.add_argument("--expected-manifest", required=True, help="independently reviewed manifest SHA256")
    args = parser.parse_args(argv)
    try:
        report = check(args.root, args.manifest, args.expected_source, args.expected_manifest)
    except (Blocked, OSError) as error:
        # Never leak private paths, raw I/O errors, manifest values, or contents.
        report = {"status": "BLOCKED", "reason": str(error) if isinstance(error, Blocked)
                  else "LOCAL_INPUT_UNAVAILABLE_OR_UNSAFE", "runtime_acceptance": "NOT_TESTED",
                  "gate_acceptance": "NOT_ASSESSED"}
    print(json.dumps(report, sort_keys=True))
    return 0 if report["status"] == "READY_FOR_REVIEW" else 1


if __name__ == "__main__":
    sys.exit(main())
