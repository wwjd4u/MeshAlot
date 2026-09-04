#!/usr/bin/env python3
"""Exercise Milestone 5 through the real restricted meshalot PostgreSQL login.

Run this orchestrator as root with MESHALOT_RUNTIME_TEST_DSN set in the inherited
environment. Database fixture/migration work runs as the postgres OS user, while
the temporary API process drops to the production meshalot OS user. The DSN must
identify user=meshalot and database=meshalot_m5_test. The DSN and generated
credentials are never printed. Test fixtures are retained.
"""
from http.cookiejar import CookieJar
import json
import os
from pathlib import Path
import pwd
import shlex
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[1]
DATABASE = "meshalot_m5_test"
PORT = 18181
BASE = f"http://127.0.0.1:{PORT}"
TOKEN = "m5-runtime-enrollment-token-not-production-0001"
RUNTIME_OS_USER = "meshalot"


def literal(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def as_postgres(command, **kwargs):
    return subprocess.run(
        ["sudo", "-n", "-u", "postgres", *command],
        text=True,
        check=True,
        **kwargs,
    )


def psql(sql: str):
    as_postgres(
        ["psql", "-X", "--no-password", "--dbname=" + DATABASE,
         "--set=ON_ERROR_STOP=1", "--file=-"],
        input=sql,
        stdout=subprocess.DEVNULL,
    )


def dsn_identity(dsn: str):
    if dsn.startswith(("postgresql://", "postgres://")):
        parsed = urllib.parse.urlparse(dsn)
        return urllib.parse.unquote(parsed.username or ""), parsed.path.lstrip("/")
    values = {}
    for token in shlex.split(dsn):
        if "=" in token:
            key, value = token.split("=", 1)
            values[key] = value
    return values.get("user", ""), values.get("dbname", "")


def drop_to_runtime_user():
    account = pwd.getpwnam(RUNTIME_OS_USER)
    os.setgid(account.pw_gid)
    os.initgroups(account.pw_name, account.pw_gid)
    os.setuid(account.pw_uid)


def request(opener, method, path, body=None, expected=200, headers=None):
    data = None if body is None else json.dumps(body).encode()
    request_headers = {"Content-Type": "application/json"}
    if headers:
        request_headers.update(headers)
    req = urllib.request.Request(BASE + path, data=data, method=method, headers=request_headers)
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
    if os.geteuid() != 0:
        raise SystemExit("Run this test orchestrator as root")
    if len(sys.argv) != 2:
        raise SystemExit("usage: milestone5_runtime_login.py /path/to/meshalot-server")

    binary = Path(sys.argv[1]).resolve()
    if not binary.is_file():
        raise SystemExit("server binary not found")
    try:
        pwd.getpwnam(RUNTIME_OS_USER)
    except KeyError:
        raise SystemExit("meshalot OS user not found")

    dsn = os.environ.get("MESHALOT_RUNTIME_TEST_DSN", "")
    if not dsn:
        raise SystemExit("MESHALOT_RUNTIME_TEST_DSN is required")
    user, database = dsn_identity(dsn)
    if user != "meshalot" or database != DATABASE:
        raise SystemExit("runtime test DSN must use the meshalot login and meshalot_m5_test database")

    with socket.socket() as probe:
        try:
            probe.bind(("127.0.0.1", PORT))
        except OSError:
            raise SystemExit(f"test port {PORT} is already in use; live API is not touched")

    as_postgres(
        [sys.executable, str(ROOT / "migrate.py"), "apply", "--database", DATABASE],
        stdout=subprocess.DEVNULL,
    )

    suffix = uuid.uuid4().hex
    user_a, user_b = str(uuid.uuid4()), str(uuid.uuid4())
    node_a_id, node_b_id = str(uuid.uuid4()), str(uuid.uuid4())
    node_a, node_b = f"m5-runtime-a-{suffix}", f"m5-runtime-b-{suffix}"
    enrolled_node = f"m5-runtime-enroll-{suffix}"
    job = str(uuid.uuid4())
    email_a = f"m5-runtime-a-{suffix}@example.invalid"
    email_b = f"m5-runtime-b-{suffix}@example.invalid"
    password_a = "M5-RUNTIME-A-" + uuid.uuid4().hex
    password_b = "M5-RUNTIME-B-" + uuid.uuid4().hex
    rate_kind = "m5-runtime-" + suffix

    psql(f"""\\set ON_ERROR_STOP on
BEGIN;
INSERT INTO users(id,email,password_hash) VALUES
('{user_a}'::uuid,{literal(email_a)},crypt({literal(password_a)},gen_salt('bf',12))),
('{user_b}'::uuid,{literal(email_b)},crypt({literal(password_b)},gen_salt('bf',12)));
INSERT INTO nodes(id,user_id,node_key,agent_version) VALUES
('{node_a_id}'::uuid,'{user_a}'::uuid,{literal(node_a)},'m5-runtime-test'),
('{node_b_id}'::uuid,'{user_b}'::uuid,{literal(node_b)},'m5-runtime-test');
INSERT INTO node_status(node_id,status,observed_at,mode,last_heartbeat) VALUES
('{node_a_id}'::uuid,'online',now(),'available',now()),
('{node_b_id}'::uuid,'online',now(),'available',now());
INSERT INTO compute_benchmarks(node_id,payload,score,observed_at) VALUES ('{node_a_id}'::uuid,'{{}}',82,now());
INSERT INTO network_benchmarks(node_id,payload,score,observed_at) VALUES ('{node_a_id}'::uuid,'{{}}',92,now());
INSERT INTO pricing_rates(workload_type,rate_microunits,inputs,effective_at) VALUES ({literal(rate_kind)},126,'{{}}',now());
INSERT INTO jobs(id,consumer_user_id,provider_node_id,status) VALUES ('{job}'::uuid,'{user_b}'::uuid,'{node_a_id}'::uuid,'queued');
UPDATE jobs SET status='running' WHERE id='{job}'::uuid;
UPDATE jobs SET status='completed' WHERE id='{job}'::uuid;
INSERT INTO usage_records(job_id,metrics) VALUES ('{job}'::uuid,'{{"tokens":101}}');
INSERT INTO wallet_transactions(user_id,job_id,transaction_type,amount_microunits,idempotency_key) VALUES
('{user_a}'::uuid,'{job}'::uuid,'earning',1100,{literal(job + ':earning')}),
('{user_b}'::uuid,'{job}'::uuid,'usage',-1100,{literal(job + ':usage')});
COMMIT;
""")

    env = os.environ.copy()
    env.update({
        "MESHALOT_LISTEN_ADDR": f"127.0.0.1:{PORT}",
        "MESHALOT_DATABASE_DSN": dsn,
        "MESHALOT_POC_USER_ID": user_a,
        "MESHALOT_DEV_ENROLLMENT_TOKEN": TOKEN,
    })
    env.pop("MESHALOT_RUNTIME_TEST_DSN", None)

    process = subprocess.Popen(
        [str(binary)],
        env=env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        preexec_fn=drop_to_runtime_user,
    )
    try:
        deadline = time.time() + 20
        while time.time() < deadline:
            if process.poll() is not None:
                raise RuntimeError("restricted-login test API exited during startup")
            try:
                with urllib.request.urlopen(BASE + "/v1/health", timeout=1) as response:
                    if response.status == 200:
                        break
            except Exception:
                time.sleep(.25)
        else:
            raise RuntimeError("restricted-login test API did not become healthy")

        jar_a, jar_b = CookieJar(), CookieJar()
        client_a = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar_a))
        client_b = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar_b))

        request(client_a, "POST", "/v1/auth/login", {"email": email_a, "password": password_a})
        request(client_b, "POST", "/v1/auth/login", {"email": email_b, "password": password_b})
        request(client_a, "GET", "/v1/auth/me")
        request(client_b, "GET", "/v1/auth/me")

        nodes_a = request(client_a, "GET", "/v1/account/nodes")["nodes"]
        nodes_b = request(client_b, "GET", "/v1/account/nodes")["nodes"]
        if node_a not in [n["node_id"] for n in nodes_a] or node_b in [n["node_id"] for n in nodes_a]:
            raise RuntimeError("restricted-login user A node separation failed")
        if node_b not in [n["node_id"] for n in nodes_b] or node_a in [n["node_id"] for n in nodes_b]:
            raise RuntimeError("restricted-login user B node separation failed")
        request(client_a, "GET", "/v1/account/nodes/" + node_b, expected=404)

        wallet_a = request(client_a, "GET", "/v1/account/wallet")
        wallet_b = request(client_b, "GET", "/v1/account/wallet")
        if wallet_a["balance_microunits"] != 1100 or wallet_b["balance_microunits"] != -1100:
            raise RuntimeError("restricted-login wallet separation failed")
        request(client_a, "GET", "/v1/dashboard")
        request(client_b, "GET", "/v1/dashboard")
        request(client_a, "GET", "/v1/account/jobs")
        request(client_b, "GET", "/v1/account/jobs")

        anon = urllib.request.build_opener()
        request(anon, "POST", "/v1/enroll", {
            "node_id": enrolled_node,
            "enrollment_token": TOKEN,
            "agent_version": "m5-runtime-restricted",
        }, expected=201)
        request(anon, "POST", "/v1/heartbeat", {
            "node_id": enrolled_node,
            "observed_at": "2026-09-04T00:00:00Z",
            "mode": "available",
        }, expected=204, headers={"Authorization": "Bearer " + TOKEN})
        provider_nodes = request(anon, "GET", "/v1/nodes", headers={"Authorization": "Bearer " + TOKEN})["nodes"]
        matched = [n for n in provider_nodes if n["node_id"] == enrolled_node]
        if len(matched) != 1 or matched[0]["status"] != "online" or matched[0]["mode"] != "available":
            raise RuntimeError("restricted-login enrollment/heartbeat persistence failed")

        request(client_a, "POST", "/v1/auth/logout", expected=204)
        request(client_a, "GET", "/v1/auth/me", expected=401)
        request(client_b, "POST", "/v1/auth/logout", expected=204)

        print("MILESTONE 5 RESTRICTED LOGIN TEST PASSED")
        print("PASS: meshalot OS user + DB login, sessions, account separation, dashboard/jobs, enrollment, heartbeat")
        print(f"Audit fixtures retained in {DATABASE}; live service untouched")
    finally:
        if process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)


if __name__ == "__main__":
    main()
