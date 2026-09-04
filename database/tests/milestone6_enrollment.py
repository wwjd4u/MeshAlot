#!/usr/bin/env python3
"""Isolated Milestone 6 identity/enrollment integration and performance test.

Run as the postgres OS user with paths to the candidate server and agent binaries.
The dedicated meshalot_m6_test database and unique test artifact directory are
retained for audit. The live API and production database are never touched.
"""
from concurrent.futures import ThreadPoolExecutor
from hashlib import sha256
from http.cookiejar import CookieJar
import base64
import json
import os
from pathlib import Path
import pwd
import socket
import statistics
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[1]
DATABASE = "meshalot_m6_test"
PORT = 18181
BASE = f"http://127.0.0.1:{PORT}"


def literal(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def run(command, **kwargs):
    return subprocess.run(command, text=True, check=True, **kwargs)


def psql(sql: str, database=DATABASE, capture=False):
    result = run(
        ["psql", "-X", "--no-password", "--dbname=" + database,
         "--set=ON_ERROR_STOP=1", "--tuples-only", "--no-align", "--file=-"],
        input=sql,
        stdout=subprocess.PIPE if capture else subprocess.DEVNULL,
    )
    return result.stdout.strip() if capture else ""


def ensure_database():
    exists = subprocess.check_output(
        ["psql", "-X", "--no-password", "--dbname=postgres", "--tuples-only",
         "--no-align", "--command", f"SELECT 1 FROM pg_database WHERE datname={literal(DATABASE)}"],
        text=True,
    ).strip()
    if not exists:
        run(["createdb", "--owner=postgres", DATABASE], stdout=subprocess.DEVNULL)
    run([sys.executable, str(ROOT / "migrate.py"), "apply", "--database", DATABASE],
        stdout=subprocess.DEVNULL)


def request(opener, method, path, body=None, expected=200):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(
        BASE + path,
        data=data,
        method=method,
        headers={"Content-Type": "application/json"},
    )
    try:
        with opener.open(req, timeout=5) as response:
            payload = None if response.status == 204 else json.load(response)
            if response.status != expected:
                raise RuntimeError(f"{path}: expected {expected}, got {response.status}")
            return payload
    except urllib.error.HTTPError as error:
        payload = json.loads(error.read().decode() or "{}")
        if error.code != expected:
            raise RuntimeError(
                f"{path}: expected {expected}, got {error.code}: {payload.get('error','')}"
            )
        return payload


def login(email, password):
    jar = CookieJar()
    client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar))
    request(client, "POST", "/v1/auth/login", {"email": email, "password": password})
    return client


def issue_code(client):
    payload = request(client, "POST", "/v1/account/enrollment-codes", {}, expected=201)
    code = payload.get("enrollment_code", "")
    if not code.startswith("mesh_") or len(code) < 40:
        raise RuntimeError("issued enrollment code does not meet format/entropy expectations")
    return code


def run_agent(agent_binary, identity_path, code, expected_success):
    identity_path.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
    os.chmod(identity_path.parent, 0o700)
    env = os.environ.copy()
    env["MESHALOT_ENROLLMENT_CODE"] = code
    result = subprocess.run(
        [str(agent_binary), "enroll", "--server", BASE, "--identity", str(identity_path)],
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        timeout=15,
    )
    if expected_success is True and result.returncode != 0:
        raise RuntimeError("agent enrollment unexpectedly failed")
    if expected_success is False and result.returncode == 0:
        raise RuntimeError("agent enrollment unexpectedly succeeded")
    return result


def read_identity(path: Path):
    mode = path.stat().st_mode & 0o777
    if mode & 0o077:
        raise RuntimeError(f"identity file permissions too broad: {mode:04o}")
    data = json.loads(path.read_text())
    if not data.get("node_id") or not data.get("public_key") or not data.get("private_key"):
        raise RuntimeError("identity file is incomplete")
    return data


def token_hash(code: str) -> str:
    return sha256(code.encode()).hexdigest()


def main():
    if pwd.getpwuid(os.geteuid()).pw_name != "postgres":
        raise SystemExit("Run this test as the postgres OS user")
    if len(sys.argv) != 3:
        raise SystemExit("usage: milestone6_enrollment.py /path/to/meshalot-server /path/to/meshalot-agent")
    server_binary = Path(sys.argv[1]).resolve()
    agent_binary = Path(sys.argv[2]).resolve()
    if not server_binary.is_file() or not agent_binary.is_file():
        raise SystemExit("server or agent binary not found")
    with socket.socket() as probe:
        try:
            probe.bind(("127.0.0.1", PORT))
        except OSError:
            raise SystemExit(f"test port {PORT} is already in use; live API is not touched")

    ensure_database()
    suffix = uuid.uuid4().hex
    artifact_dir = Path("/tmp") / f"meshalot-m6-test-{suffix}"
    artifact_dir.mkdir(mode=0o700)
    user_a, user_b = str(uuid.uuid4()), str(uuid.uuid4())
    email_a, email_b = f"m6-a-{suffix}@example.invalid", f"m6-b-{suffix}@example.invalid"
    password_a, password_b = "M6-A-" + uuid.uuid4().hex, "M6-B-" + uuid.uuid4().hex
    psql(f"""BEGIN;
INSERT INTO users(id,email,password_hash) VALUES
('{user_a}'::uuid,{literal(email_a)},crypt({literal(password_a)},gen_salt('bf',12))),
('{user_b}'::uuid,{literal(email_b)},crypt({literal(password_b)},gen_salt('bf',12)));
COMMIT;
""")

    env = os.environ.copy()
    env.update({
        "MESHALOT_LISTEN_ADDR": f"127.0.0.1:{PORT}",
        "MESHALOT_DATABASE_DSN": f"host=/var/run/postgresql dbname={DATABASE} user=postgres sslmode=disable",
        "MESHALOT_POC_USER_ID": user_a,
        "MESHALOT_DEV_ENROLLMENT_TOKEN": "m6-test-legacy-token-not-production-0000001",
    })
    log_path = artifact_dir / "server.log"
    log_file = open(log_path, "w", encoding="utf-8")
    os.chmod(log_path, 0o600)
    process = subprocess.Popen([str(server_binary)], env=env, stdout=log_file, stderr=subprocess.STDOUT)
    sensitive_values = []
    identity_paths = []
    try:
        deadline = time.time() + 20
        while time.time() < deadline:
            if process.poll() is not None:
                raise RuntimeError("temporary Milestone 6 API exited during startup")
            try:
                with urllib.request.urlopen(BASE + "/v1/health", timeout=1) as response:
                    if response.status == 200:
                        break
            except Exception:
                time.sleep(.25)
        else:
            raise RuntimeError("temporary Milestone 6 API did not become healthy")

        client_a = login(email_a, password_a)
        client_b = login(email_b, password_b)

        code_a1 = issue_code(client_a)
        sensitive_values.append(code_a1)
        identity_a1 = artifact_dir / "agent-a1" / "identity.json"
        identity_paths.append(identity_a1)
        run_agent(agent_binary, identity_a1, code_a1, True)
        first_identity = read_identity(identity_a1)
        status = subprocess.run(
            [str(agent_binary), "identity", "--identity", str(identity_a1)],
            text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=10,
        )
        if status.returncode != 0:
            raise RuntimeError("agent failed to reload persistent identity in a new process")
        second_identity = read_identity(identity_a1)
        if first_identity != second_identity:
            raise RuntimeError("agent identity changed across restart/reload")

        run_agent(agent_binary, identity_a1, code_a1, False)

        invalid_code = "mesh_" + base64.urlsafe_b64encode(os.urandom(32)).decode().rstrip("=")
        sensitive_values.append(invalid_code)
        invalid_identity = artifact_dir / "agent-invalid" / "identity.json"
        identity_paths.append(invalid_identity)
        run_agent(agent_binary, invalid_identity, invalid_code, False)
        invalid_before = read_identity(invalid_identity)
        run_agent(agent_binary, invalid_identity, invalid_code, False)
        if invalid_before != read_identity(invalid_identity):
            raise RuntimeError("invalid enrollment changed local identity")

        expired_code = issue_code(client_a)
        sensitive_values.append(expired_code)
        psql(f"UPDATE enrollment_tokens SET expires_at=now()-interval '1 minute' WHERE token_hash={literal(token_hash(expired_code))};")
        expired_identity = artifact_dir / "agent-expired" / "identity.json"
        identity_paths.append(expired_identity)
        run_agent(agent_binary, expired_identity, expired_code, False)

        code_b = issue_code(client_b)
        sensitive_values.append(code_b)
        identity_b = artifact_dir / "agent-b" / "identity.json"
        identity_paths.append(identity_b)
        run_agent(agent_binary, identity_b, code_b, True)
        hijack_code = issue_code(client_a)
        sensitive_values.append(hijack_code)
        run_agent(agent_binary, identity_b, hijack_code, False)
        identity_a2 = artifact_dir / "agent-a2" / "identity.json"
        identity_paths.append(identity_a2)
        run_agent(agent_binary, identity_a2, hijack_code, True)

        race_code = issue_code(client_a)
        sensitive_values.append(race_code)
        race_paths = [artifact_dir / "race-1" / "identity.json", artifact_dir / "race-2" / "identity.json"]
        identity_paths.extend(race_paths)
        with ThreadPoolExecutor(max_workers=2) as pool:
            futures = [pool.submit(run_agent, agent_binary, path, race_code, None) for path in race_paths]
            results = []
            for future in futures:
                results.append(future.result())
        successes = sum(result.returncode == 0 for result in results)
        if successes != 1:
            raise RuntimeError(f"double-redeem race allowed {successes} successes; expected exactly one")

        if int(psql(f"SELECT count(*) FROM enrollment_tokens WHERE token_hash={literal(code_a1)};", capture=True)) != 0:
            raise RuntimeError("raw enrollment code was stored in PostgreSQL")
        consumed = psql(
            f"SELECT (consumed_at IS NOT NULL AND consumed_node_id IS NOT NULL)::int FROM enrollment_tokens WHERE token_hash={literal(token_hash(code_a1))};",
            capture=True,
        )
        if consumed != "1":
            raise RuntimeError("consumed enrollment token lacks atomic node binding")

        private_values = []
        for path in identity_paths:
            if path.exists():
                private_values.append(read_identity(path)["private_key"])
        for private_value in private_values:
            count = int(psql(f"""SELECT
(SELECT count(*) FROM nodes WHERE node_key={literal(private_value)} OR agent_version={literal(private_value)} OR identity_public_key={literal(private_value)}) +
(SELECT count(*) FROM enrollment_tokens WHERE token_hash={literal(private_value)}) +
(SELECT count(*) FROM user_sessions WHERE token_hash={literal(private_value)});
""", capture=True))
            if count != 0:
                raise RuntimeError("private identity material reached PostgreSQL")

        latencies_ms = []
        for _ in range(20):
            start = time.perf_counter()
            code = issue_code(client_a)
            sensitive_values.append(code)
            public_key = base64.b64encode(os.urandom(32)).decode().rstrip("=")
            request(
                urllib.request.build_opener(),
                "POST",
                "/v1/agent/enroll",
                {
                    "node_id": str(uuid.uuid4()),
                    "enrollment_code": code,
                    "agent_version": "m6-perf-test",
                    "public_key": public_key,
                },
                expected=201,
            )
            latencies_ms.append((time.perf_counter() - start) * 1000)

        request(client_a, "POST", "/v1/auth/logout", expected=204)
        request(client_b, "POST", "/v1/auth/logout", expected=204)
    finally:
        if process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)
        log_file.close()

    log_text = log_path.read_text(errors="replace")
    for value in sensitive_values:
        if value and value in log_text:
            raise RuntimeError("raw enrollment code appeared in server logs")
    for path in identity_paths:
        if path.exists():
            private_value = read_identity(path)["private_key"]
            if private_value in log_text:
                raise RuntimeError("private identity material appeared in server logs")

    ordered = sorted(latencies_ms)
    p50 = statistics.median(ordered)
    p95 = ordered[max(0, int(len(ordered) * 0.95) - 1)]
    print("MILESTONE 6 ENROLLMENT TEST PASSED")
    print("PASS: persistent identity, one-time/expired/invalid rejection, account isolation, atomic double-redeem protection")
    print("PASS: raw codes are hashed at rest; private identity material absent from database and server logs")
    print(f"Enrollment-path sample: n={len(ordered)} p50={p50:.1f}ms p95={p95:.1f}ms")
    print(f"Audit database retained: {DATABASE}")
    print(f"Audit artifacts retained: {artifact_dir}")


if __name__ == "__main__":
    main()
