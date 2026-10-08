#!/usr/bin/env python3
"""Stage a macOS LaunchAgent plist without installing or loading it.

This intentionally does not call launchctl, write to ~/Library/LaunchAgents,
generate identities, read private key bytes, or deploy software.
"""
import argparse
import os
import plistlib
import stat
from pathlib import Path
from urllib.parse import urlsplit


def _regular_owned_file(raw: str, name: str, executable: bool) -> Path:
    path = Path(raw).expanduser()
    if not path.is_absolute() or path.is_symlink() or not path.is_file():
        raise ValueError(f"{name} must be an existing non-symlink absolute file")
    info = path.stat()
    if info.st_uid != os.getuid():
        raise ValueError(f"{name} must be owned by the current login user")
    if executable and not os.access(path, os.X_OK):
        raise ValueError(f"{name} is not executable")
    if not executable and (stat.S_IMODE(info.st_mode) & 0o077):
        raise ValueError("existing node identity permissions must be 0600 or stricter")
    return path


def _validate_server(server: str) -> str:
    try:
        parsed = urlsplit(server)
        _ = parsed.port
    except ValueError as exc:
        raise ValueError("invalid HTTPS control-plane URL") from exc
    if (parsed.scheme != "https" or not parsed.hostname or parsed.username
            or parsed.password or parsed.path not in ("", "/")
            or parsed.query or parsed.fragment):
        raise ValueError("control-plane base URL must be plain HTTPS without credentials or subpaths")
    return server


def stage_launch_agent(binary: str, identity: str, server: str,
                       mode: str, manual_pause: str, output: str) -> Path:
    if mode not in ("normal", "away", "maximum-earnings"):
        raise ValueError("explicit provider mode is required")
    if manual_pause not in ("true", "false"):
        raise ValueError("explicit true/false manual-pause value is required")
    binary_path = _regular_owned_file(binary, "agent binary", executable=True)
    identity_path = _regular_owned_file(identity, "existing node identity", executable=False)
    _validate_server(server)
    output_path = Path(output).expanduser()
    if not output_path.is_absolute():
        raise ValueError("staged plist output must be an absolute path")
    parent = output_path.parent.resolve(strict=True)
    if not parent.is_dir():
        raise ValueError("staging directory must exist")
    if len(parent.parts) >= 2 and parent.parts[-2:] == ("Library", "LaunchAgents"):
        raise ValueError("refusing to install directly into LaunchAgents")
    # Never follow an existing symlink or overwrite another staging file.
    arguments = [
        str(binary_path), "connect",
        "--server", server,
        "--identity", str(identity_path),
        "--mode", mode,
        "--manual-pause", manual_pause,
        "--job-state", "unknown",
        "--interval", "30s",
    ]
    service = {
        "Label": "com.meshalot.agent",
        "ProgramArguments": arguments,
        "RunAtLoad": True,
        "KeepAlive": {"SuccessfulExit": False},
        "ThrottleInterval": 15,
        "ProcessType": "Background",
    }
    fd = os.open(output_path, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    try:
        with os.fdopen(fd, "wb") as handle:
            plistlib.dump(service, handle, fmt=plistlib.FMT_XML)
            handle.flush()
            os.fsync(handle.fileno())
    except Exception:
        output_path.unlink(missing_ok=True)
        raise
    return output_path


def main() -> None:
    parser = argparse.ArgumentParser(description="Stage (never install) MeshAlot M13 LaunchAgent")
    parser.add_argument("--binary", required=True, help="existing built agent executable")
    parser.add_argument("--identity", required=True, help="existing 0600 enrolled identity file")
    parser.add_argument("--server", required=True, help="HTTPS control-plane base URL")
    parser.add_argument("--mode", required=True, choices=("normal", "away", "maximum-earnings"))
    parser.add_argument("--manual-pause", required=True, choices=("true", "false"))
    parser.add_argument("--output", required=True, help="absolute path OUTSIDE ~/Library/LaunchAgents")
    args = parser.parse_args()
    try:
        path = stage_launch_agent(args.binary, args.identity, args.server,
                                  args.mode, args.manual_pause, args.output)
    except (ValueError, OSError) as exc:
        parser.error(str(exc))
    print(f"staged LaunchAgent plist (NOT installed): {path}")


if __name__ == "__main__":
    main()
