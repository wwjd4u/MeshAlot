#!/usr/bin/env python3
"""Test an isolated API process restart; never controls the live systemd service."""
import argparse
import json
import os
from pathlib import Path
import pwd
import secrets
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=Path)
    args = parser.parse_args()
    if pwd.getpwuid(os.geteuid()).pw_name != "meshalot":
        raise RuntimeError("Run as OS user meshalot, not root/postgres")
    binary = args.binary.resolve(strict=True)
    token = secrets.token_hex(32)
    node = "process-restart-test-" + str(uuid.uuid4())
    with socket.socket() as probe:
        probe.bind(("127.0.0.1", 0))
        port = probe.getsockname()[1]
    env = dict(os.environ)
    env.update({
        "MESHALOT_LISTEN_ADDR": "127.0.0.1:" + str(port),
        "MESHALOT_DATABASE_DSN": "host=/var/run/postgresql dbname=meshalot_fresh_test user=meshalot sslmode=disable connect_timeout=5",
        "MESHALOT_POC_USER_ID": "00000000-0000-4000-8000-000000000004",
        "MESHALOT_DEV_ENROLLMENT_TOKEN": token,
    })
    # No proxies for loopback traffic; never transmit the temporary token elsewhere.
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))

    def request(path, data=None, expected=200):
        body = None if data is None else json.dumps(data).encode()
        req = urllib.request.Request("http://127.0.0.1:" + str(port) + path, data=body,
                                     headers={"Authorization": "Bearer " + token,
                                              "Content-Type": "application/json"})
        with opener.open(req, timeout=5) as response:
            if response.status != expected:
                raise RuntimeError("Unexpected HTTP status")
            raw = response.read()
            return json.loads(raw) if raw else None

    def stop(process):
        # Only ever terminate the subprocess created by this test.
        if process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)

    def start():
        process = subprocess.Popen([str(binary)], env=env, stdout=subprocess.DEVNULL)
        try:
            for _ in range(30):
                if process.poll() is not None:
                    raise RuntimeError("Test API exited during startup")
                try:
                    request("/v1/health")
                    request("/v1/nodes")  # Unique token verifies the expected instance.
                    return process
                except (urllib.error.URLError, TimeoutError):
                    time.sleep(0.5)
            raise RuntimeError("Test API did not become ready")
        except BaseException:
            stop(process)
            raise

    def read_node():
        matches = [item for item in request("/v1/nodes")["nodes"] if item["node_id"] == node]
        if len(matches) != 1:
            raise RuntimeError("Expected exactly one persisted test node")
        return matches[0]

    first = start()
    first_pid = first.pid
    try:
        request("/v1/enroll", {"node_id": node, "agent_version": "process-test",
                                "enrollment_token": token}, expected=201)
        request("/v1/heartbeat", {"node_id": node, "mode": "available"}, expected=204)
        before = read_node()
        if before["status"] != "online":
            raise RuntimeError("Heartbeat did not persist")
    finally:
        stop(first)
    print("First test API stopped, PID:", first_pid, flush=True)
    second = start()
    try:
        if second.pid == first_pid:
            raise RuntimeError("PID reused: rerun test for unambiguous process evidence")
        after = read_node()
        if after != before:
            raise RuntimeError("Node changed or disappeared across process restart")
        print("Second test API PID:", second.pid)
        print("PASS: node identity, agent version, mode, status and heartbeat survived process restart")
        print("Test node retained in meshalot_fresh_test:", node)
    finally:
        stop(second)
    print("PROCESS RESTART TEST PASSED - isolated API processes stopped; live service untouched")


if __name__ == "__main__":
    try:
        main()
    except (OSError, RuntimeError, ValueError) as error:
        print("PROCESS TEST FAILED:", error, file=sys.stderr)
        sys.exit(1)
