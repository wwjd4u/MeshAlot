#!/usr/bin/env python3
"""Isolated Milestone 5 authentication/account-separation integration test.

Run as the postgres OS user. The dedicated meshalot_m5_test database and test
fixtures are retained for audit; only the temporary test API process is stopped.
"""
from http.cookiejar import CookieJar
import json
import os
from pathlib import Path
import pwd
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[1]
DATABASE = "meshalot_m5_test"
PORT = 18180
BASE = f"http://127.0.0.1:{PORT}"


def literal(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def run(command, **kwargs):
    return subprocess.run(command, text=True, check=True, **kwargs)


def psql(sql: str, database=DATABASE):
    run(["psql", "-X", "--no-password", "--dbname=" + database,
         "--set=ON_ERROR_STOP=1", "--file=-"], input=sql,
        stdout=subprocess.DEVNULL)


def ensure_database():
    exists = subprocess.check_output(
        ["psql", "-X", "--no-password", "--dbname=postgres", "--tuples-only",
         "--no-align", "--command", f"SELECT 1 FROM pg_database WHERE datname={literal(DATABASE)}"],
        text=True).strip()
    if not exists:
        run(["createdb", "--owner=postgres", DATABASE], stdout=subprocess.DEVNULL)
    run([sys.executable, str(ROOT / "migrate.py"), "apply", "--database", DATABASE],
        stdout=subprocess.DEVNULL)


def request(opener, method, path, body=None, expected=200):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(BASE + path, data=data, method=method,
                                 headers={"Content-Type": "application/json"})
    try:
        with opener.open(req, timeout=5) as response:
            payload = None if response.status == 204 else json.load(response)
            if response.status != expected:
                raise RuntimeError(f"{path}: expected {expected}, got {response.status}")
            return payload
    except urllib.error.HTTPError as error:
        payload = json.loads(error.read().decode() or "{}")
        if error.code != expected:
            raise RuntimeError(f"{path}: expected {expected}, got {error.code}: {payload.get('error','')}")
        return payload


def main():
    if pwd.getpwuid(os.geteuid()).pw_name != "postgres":
        raise SystemExit("Run this test as the postgres OS user")
    if len(sys.argv) != 2:
        raise SystemExit("usage: milestone5_auth.py /path/to/meshalot-server")
    binary = Path(sys.argv[1]).resolve()
    if not binary.is_file():
        raise SystemExit("server binary not found")
    with socket.socket() as probe:
        try:
            probe.bind(("127.0.0.1", PORT))
        except OSError:
            raise SystemExit(f"test port {PORT} is already in use; live API is not touched")

    ensure_database()
    suffix = uuid.uuid4().hex
    user_a, user_b = str(uuid.uuid4()), str(uuid.uuid4())
    node_a_id, node_b_id = str(uuid.uuid4()), str(uuid.uuid4())
    node_a, node_b = f"m5-a-{suffix}", f"m5-b-{suffix}"
    job = str(uuid.uuid4())
    email_a, email_b = f"m5-a-{suffix}@example.invalid", f"m5-b-{suffix}@example.invalid"
    password_a, password_b = "M5-A-" + uuid.uuid4().hex, "M5-B-" + uuid.uuid4().hex
    rate_kind = "m5-" + suffix
    fixture_sql = f"""\\set ON_ERROR_STOP on
BEGIN;
INSERT INTO users(id,email,password_hash) VALUES
('{user_a}'::uuid,{literal(email_a)},crypt({literal(password_a)},gen_salt('bf',12))),
('{user_b}'::uuid,{literal(email_b)},crypt({literal(password_b)},gen_salt('bf',12)));
INSERT INTO nodes(id,user_id,node_key,agent_version) VALUES
('{node_a_id}'::uuid,'{user_a}'::uuid,{literal(node_a)},'m5-test'),
('{node_b_id}'::uuid,'{user_b}'::uuid,{literal(node_b)},'m5-test');
INSERT INTO node_status(node_id,status,observed_at,mode,last_heartbeat) VALUES
('{node_a_id}'::uuid,'online',now(),'available',now()),
('{node_b_id}'::uuid,'online',now(),'available',now());
INSERT INTO compute_benchmarks(node_id,payload,score,observed_at) VALUES ('{node_a_id}'::uuid,'{{}}',81,now());
INSERT INTO network_benchmarks(node_id,payload,score,observed_at) VALUES ('{node_a_id}'::uuid,'{{}}',91,now());
INSERT INTO pricing_rates(workload_type,rate_microunits,inputs,effective_at) VALUES ({literal(rate_kind)},125,'{{}}',now());
INSERT INTO jobs(id,consumer_user_id,provider_node_id,status) VALUES ('{job}'::uuid,'{user_b}'::uuid,'{node_a_id}'::uuid,'queued');
UPDATE jobs SET status='running' WHERE id='{job}'::uuid;
UPDATE jobs SET status='completed' WHERE id='{job}'::uuid;
INSERT INTO usage_records(job_id,metrics) VALUES ('{job}'::uuid,'{{"tokens":100}}');
INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES
('{user_a}'::uuid,'{job}'::uuid,'earning',1000,{literal(job + ':earning')}),
('{user_b}'::uuid,'{job}'::uuid,'usage',-1000,{literal(job + ':usage')});
COMMIT;
"""
    psql(fixture_sql)

    env = os.environ.copy()
    env.update({
        "MESHALOT_LISTEN_ADDR": f"127.0.0.1:{PORT}",
        "MESHALOT_DATABASE_DSN": f"host=/var/run/postgresql dbname={DATABASE} user=postgres sslmode=disable",
        "MESHALOT_POC_USER_ID": user_a,
        "MESHALOT_DEV_ENROLLMENT_TOKEN": "m5-test-enrollment-token-not-production-0001",
    })
    process = subprocess.Popen([str(binary)], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        deadline = time.time() + 20
        while time.time() < deadline:
            if process.poll() is not None:
                raise RuntimeError("temporary Milestone 5 API exited during startup")
            try:
                with urllib.request.urlopen(BASE + "/v1/health", timeout=1) as response:
                    if response.status == 200:
                        break
            except Exception:
                time.sleep(.25)
        else:
            raise RuntimeError("temporary Milestone 5 API did not become healthy")

        anon = urllib.request.build_opener()
        error = request(anon, "GET", "/v1/auth/me", expected=401)
        if "sign in" not in error.get("error", ""):
            raise RuntimeError("unauthenticated error is not understandable")
        error = request(anon, "POST", "/v1/auth/login", {"email": email_a, "password": "wrong"}, expected=401)
        if "invalid email or password" != error.get("error"):
            raise RuntimeError("invalid-login error is not understandable")

        jar_a, jar_b = CookieJar(), CookieJar()
        client_a = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar_a))
        client_b = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar_b))
        request(client_a, "POST", "/v1/auth/login", {"email": email_a, "password": password_a})
        request(client_b, "POST", "/v1/auth/login", {"email": email_b, "password": password_b})

        nodes_a = request(client_a, "GET", "/v1/account/nodes")["nodes"]
        nodes_b = request(client_b, "GET", "/v1/account/nodes")["nodes"]
        if [n["node_id"] for n in nodes_a] != [node_a] or [n["node_id"] for n in nodes_b] != [node_b]:
            raise RuntimeError("node account separation failed")
        request(client_a, "GET", "/v1/account/nodes/" + node_b, expected=404)

        wallet_a = request(client_a, "GET", "/v1/account/wallet")
        wallet_b = request(client_b, "GET", "/v1/account/wallet")
        if wallet_a["balance_microunits"] != 1000 or wallet_b["balance_microunits"] != -1000:
            raise RuntimeError("wallet account separation failed")
        if any(item["amount_microunits"] < 0 for item in wallet_a["transactions"]):
            raise RuntimeError("user A can see user B wallet entry")
        if any(item["amount_microunits"] > 0 for item in wallet_b["transactions"]):
            raise RuntimeError("user B can see user A wallet entry")

        for client in (client_a, client_b):
            request(client, "GET", "/v1/dashboard")
            request(client, "GET", "/v1/account/jobs")
        request(client_a, "POST", "/v1/auth/logout", expected=204)
        request(client_a, "GET", "/v1/auth/me", expected=401)
        request(client_b, "POST", "/v1/auth/logout", expected=204)
        print("MILESTONE 5 AUTH/API TEST PASSED")
        print("PASS: sign in, sign out, node separation, wallet separation, dashboard/jobs APIs, understandable errors")
        print(f"Audit fixtures retained in {DATABASE}; live service untouched")
    finally:
        if process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill(); process.wait(timeout=5)


if __name__ == "__main__":
    main()
