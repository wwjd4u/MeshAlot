"""Public fixture only, inside a no-NIC disposable VM. Never run on a real host."""
import hashlib
from http.cookiejar import CookieJar
import ipaddress
import json
import os
from pathlib import Path, PurePosixPath
import posixpath
import shutil
import socket
import ssl
import stat
import subprocess
from safetar import unpack_rootfs
import time
import urllib.error
import urllib.request

APP = "4f67ee54e5ebca084e38e306e723ae0acd08e77d"
ROOT = Path("/opt/meshalot-fixture")
PG = ROOT / "pg-root"
SEED = Path("/run/m13-fixture-seed")
UNITDIR = Path("/etc/systemd/system")
DB = "meshalot_m13stage"
BASE = "https://127.0.0.1:18443"
CA = Path("/var/lib/m13-fixture-caddy/data/caddy/pki/authorities/local/root.crt")
USER_A = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
USER_B = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
NODE = "11111111-1111-4111-8111-111111111111"
checks = {}
phase = "guest_boundary"
started = time.monotonic()
result = "FAIL"
performance = {}
last_failure = {}
service_failure = {}


def run(args, *, input=None, timeout=30, check=True, env=None):
    global last_failure
    value = subprocess.run(args, input=input, capture_output=True, text=True,
                           timeout=timeout, env=env)
    if check and value.returncode:
        text = (value.stdout + value.stderr).lower()
        labels = {"permission denied":"PERMISSION", "no such file":"FILE_NOT_FOUND",
                  "not found":"NOT_FOUND", "peer authentication":"PEER_AUTH", "role":"ROLE",
                  "schema":"SCHEMA", "extension":"EXTENSION", "shared memory":"SHARED_MEMORY",
                  "read-only":"READ_ONLY", "operation not permitted":"OPERATION_NOT_PERMITTED",
                  "mount":"MOUNT", "locale":"LOCALE", "invalid authentication":"AUTH_METHOD",
                  "connection":"CONNECTION", "not exist":"MISSING_OBJECT", "syntax":"SYNTAX"}
        last_failure = {"command":Path(args[0]).name,"exit":value.returncode,
                        "classes":sorted({label for fragment,label in labels.items() if fragment in text})}
        value.check_returncode()
    return value


def require(name, ok):
    checks[name] = bool(ok)
    if not ok:
        raise RuntimeError(name)


def digest(path):
    value = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 ** 2), b""):
            value.update(block)
    return value.hexdigest()




def sql(statement, database=DB, role="stage_bootstrap"):
    return run(["chroot", "--userspec=20001:20001", str(PG), "/usr/local/bin/psql",
                "-X", "--no-password", "--host=/var/lib/postgresql/fixture-socket", "--username=" + role,
                "--dbname=" + database, "--set=ON_ERROR_STOP=1", "-At", "--file=-"],
               input=statement).stdout.strip()


def service(action, name, check=True):
    return run(["systemctl", action, "m13-fixture-" + name + ".service"], check=check)


def service_state(name):
    global service_failure
    unit = "m13-fixture-" + name + ".service"
    properties = run(["systemctl", "show", unit, "-p", "ActiveState", "-p", "SubState",
                      "-p", "Result", "-p", "ExecMainCode", "-p", "ExecMainStatus"]).stdout
    service_failure = dict(line.split("=", 1) for line in properties.splitlines() if "=" in line)
    logs = run(["journalctl", "-b", "-u", unit, "-n", "30", "--output=json", "--no-pager"], check=False).stdout
    messages = []
    for line in logs.splitlines():
        try:
            message = json.loads(line).get("MESSAGE", "")
            if isinstance(message, str):
                messages.append(message)
        except (ValueError, AttributeError):
            continue
    text = " ".join(messages).lower()
    classes = {"permission denied":"PERMISSION", "no such file":"FILE_NOT_FOUND",
               "group":"GROUP", "user":"USER", "shared memory":"SHARED_MEMORY",
               "socket":"SOCKET", "mount":"MOUNT", "namespace":"NAMESPACE",
               "fatal":"FATAL", "could not":"DEPENDENCY", "exec":"EXEC"}
    service_failure["journal_classes"] = sorted({v for k,v in classes.items() if k in text})
    return service_failure


def opener(jar=None):
    options = [urllib.request.ProxyHandler({}), urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile=CA))]
    if jar is not None:
        options.append(urllib.request.HTTPCookieProcessor(jar))
    return urllib.request.build_opener(*options)


def request(client, method, path, data=None, expected=200):
    body = None if data is None else json.dumps(data).encode()
    query = urllib.request.Request(BASE + path, data=body, method=method,
                                   headers={"Content-Type": "application/json"})
    try:
        response = client.open(query, timeout=7)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        if response.code != expected:
            raise RuntimeError("http_expectation")
        raw = response.read(65537)
        if len(raw) > 65536:
            raise ValueError("response_budget")
        return json.loads(raw) if raw and raw.startswith(b"{") else None


def healthy():
    deadline = time.monotonic() + 35
    while time.monotonic() < deadline:
        try:
            if CA.is_file():
                request(opener(), "GET", "/v1/health")
                return
        except (OSError, RuntimeError, ValueError):
            pass
        time.sleep(.25)
    raise RuntimeError("health_deadline")


def heartbeat(mode):
    value = json.loads(run([str(ROOT / "payload/probe"), mode], timeout=35).stdout)
    require("signed_wss_" + mode, value["status"] == "PASS" and all(value["checks"].values()))
    expected = 6 if mode == "initial" else 2
    require("wss_check_count_" + mode, len(value["checks"]) == expected)


try:
    require("guest_pid1_systemd", Path("/proc/1/comm").read_text().strip() == "systemd")
    require("guest_no_external_interface", sorted(p.name for p in Path("/sys/class/net").iterdir()) == ["lo"])
    for family in (4, 6):
        routes = json.loads(run(["ip", "-j", "-" + str(family), "route", "show", "table", "all"]).stdout)
        require("guest_no_ipv" + str(family) + "_external_route", all(
            route.get("dev") == "lo" and ipaddress.ip_network(route["dst"], strict=False).is_loopback for route in routes))
    phase = "verify_offline_payload"
    SEED.mkdir(mode=0o700)
    device = run(["blkid", "-L", "cidata"]).stdout.strip()
    require("seed_device", device in ("/dev/vdb", "/dev/sr0"))
    run(["mount", "-t", "iso9660", "-o", "ro,nodev,nosuid,noexec", device, str(SEED)])
    manifest = json.loads((SEED / "assets.json").read_text())
    require("artifact_source_scope", manifest["application_commit"] == APP and manifest["material"] == "public_synthetic_only")
    ROOT.mkdir(mode=0o755)
    shutil.copytree(SEED / "payload", ROOT / "payload")
    # The private assembler's umask is 077; permit the three guest services to
    # traverse this public executable directory without widening data/key modes.
    (ROOT / "payload").chmod(0o755)
    for name, expected in manifest["files"].items():
        path = PurePosixPath(name)
        if path.is_absolute() or ".." in path.parts:
            raise ValueError("asset_path")
        require("payload_digest_" + name.replace("/", "_"), digest(ROOT / "payload" / name) == expected)
    for name in ("api", "probe", "caddy"):
        (ROOT / "payload" / name).chmod(0o755)
    phase = "synthetic_pg_root"
    unpack_rootfs(ROOT / "payload/pg-rootfs.tar", PG)
    (PG / "dev").mkdir(exist_ok=True)
    if not (PG / "dev/null").exists():
        os.mknod(PG / "dev/null", stat.S_IFCHR | 0o666, os.makedev(1, 3))
    (PG / "dev/null").chmod(0o666)
    (PG / "dev/shm").mkdir(mode=0o1777, exist_ok=True)
    (PG / "dev/shm").chmod(0o1777)
    for name, uid in (("pgstage", 20001), ("meshalot", 20002), ("stagecaddy", 20003)):
        run(["groupadd", "--gid", str(uid), name])
        run(["useradd", "--system", "--uid", str(uid), "--gid", str(uid), "--no-create-home", "--shell", "/usr/sbin/nologin", name])
        with (PG / "etc/passwd").open("a") as output:
            output.write(name + ":x:" + str(uid) + ":" + str(uid) + ":fixture:/tmp:/bin/sh\n")
        with (PG / "etc/group").open("a") as output:
            output.write(name + ":x:" + str(uid) + ":\n")
    for path in (PG / "var/lib/postgresql/stage", PG / "var/lib/postgresql/fixture-socket"):
        path.mkdir(parents=True, mode=0o755, exist_ok=True)
        os.chown(path, 20001, 20001)
    run(["chroot", "--userspec=20001:20001", str(PG), "/usr/local/bin/initdb", "-D", "/var/lib/postgresql/stage",
         "--username=stage_bootstrap", "--auth-local=peer", "--auth-host=reject", "--encoding=UTF8", "--no-locale"], timeout=90)
    pgdata = PG / "var/lib/postgresql/stage"
    (pgdata / "pg_hba.conf").write_text("local all stage_bootstrap peer map=stage_local\nlocal all stage_owner peer map=stage_local\nlocal meshalot_m13stage meshalot peer\nlocal all all reject\nhost all all 0.0.0.0/0 reject\nhost all all ::/0 reject\n")
    (pgdata / "pg_ident.conf").write_text("stage_local pgstage stage_bootstrap\nstage_local pgstage stage_owner\n")
    (pgdata / "postgresql.conf").write_text("listen_addresses = ''\nunix_socket_directories = '/var/lib/postgresql/fixture-socket'\nshared_buffers = '32MB'\nmax_connections = 20\nmax_wal_size = '128MB'\nmin_wal_size = '32MB'\n")
    UNITDIR.joinpath("m13-fixture-pg.service").write_text("[Unit]\nDescription=Disposable synthetic PG16 only\n[Service]\nType=simple\nUser=20001\nGroup=20001\nRootDirectory=" + str(PG) + "\nExecStart=/usr/local/bin/postgres -D /var/lib/postgresql/stage\nNoNewPrivileges=yes\nRestart=no\nTimeoutStopSec=30\nKillSignal=SIGINT\n")
    run(["systemctl", "daemon-reload"])
    service("start", "pg")
    deadline = time.monotonic() + 25
    while True:
        try:
            version = sql("SHOW server_version_num;", "postgres")
            break
        except subprocess.SubprocessError:
            if time.monotonic() >= deadline:
                service_state("pg")
                raise
            time.sleep(.2)
    require("postgres_major_16", 160000 <= int(version) < 170000)
    sql("CREATE ROLE stage_owner LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;\nCREATE ROLE meshalot LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;\nCREATE DATABASE " + DB + " OWNER stage_owner;", "postgres")
    sql("CREATE EXTENSION pgcrypto;")
    phase = "m13_schema_and_runtime_role"
    shim = ROOT / "bin"
    shim.mkdir(mode=0o700)
    shim.joinpath("psql").write_text("#!/bin/sh\nexec /usr/sbin/chroot --userspec=20001:20001 " + str(PG) + " /usr/local/bin/psql \"$@\"\n")
    shim.joinpath("psql").chmod(0o700)
    env = dict(os.environ, PATH=str(shim) + ":/usr/bin:/bin", PGHOST="/var/lib/postgresql/fixture-socket", PGUSER="stage_owner")
    run(["python3", "-B", str(ROOT / "payload/database/migrate.py"), "apply", "--database", DB], env=env, timeout=90)
    sql((ROOT / "payload/database/runtime-grants.sql").read_text(), role="stage_owner")
    ledger_before = sql("SELECT version::text||':'||sha256 FROM meshalot_meta.schema_migrations ORDER BY version;")
    require("schema_10_complete", len(ledger_before.splitlines()) == 10 and ledger_before.splitlines()[-1].startswith("10:"))
    sql("INSERT INTO users(id,email,password_hash) VALUES ('" + USER_A + "','fixture-a@example.invalid',crypt('public-fixture-password-a',gen_salt('bf',4))),('" + USER_B + "','fixture-b@example.invalid',crypt('public-fixture-password-b',gen_salt('bf',4))); ")
    require("owner_and_runtime_separate", sql("SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname=current_database();") == "stage_owner")
    phase = "real_service_start"
    common = "\nNoNewPrivileges=yes\nPrivateTmp=yes\nProtectSystem=strict\nProtectHome=yes\nRestart=no\n"
    UNITDIR.joinpath("m13-fixture-api.service").write_text("[Unit]\nDescription=Public M13 API fixture only\n[Service]\nType=simple\nUser=meshalot\nGroup=meshalot\nExecStart=" + str(ROOT / "payload/api") + "\nEnvironment=MESHALOT_LISTEN_ADDR=127.0.0.1:18080\nEnvironment=\"MESHALOT_DATABASE_DSN=host=" + str(PG / "var/lib/postgresql/fixture-socket") + " dbname=" + DB + " user=meshalot sslmode=disable\"\nEnvironment=MESHALOT_POC_USER_ID=" + USER_A + "\nEnvironment=MESHALOT_DEV_ENROLLMENT_TOKEN=public-synthetic-fixture-only-token-2026\n" + common)
    caddy_state = Path("/var/lib/m13-fixture-caddy")
    caddy_state.mkdir(mode=0o700)
    os.chown(caddy_state, 20003, 20003)
    config = ROOT / "Caddyfile"
    config.write_text("{\n admin off\n auto_https disable_redirects\n skip_install_trust\n}\nhttps://127.0.0.1:18443 {\n bind 127.0.0.1\n tls internal\n reverse_proxy 127.0.0.1:18080\n}\n")
    UNITDIR.joinpath("m13-fixture-caddy.service").write_text("[Unit]\nDescription=Public internal-TLS fixture only\n[Service]\nType=simple\nUser=stagecaddy\nGroup=stagecaddy\nExecStart=" + str(ROOT / "payload/caddy") + " run --config " + str(config) + " --adapter caddyfile\nEnvironment=XDG_DATA_HOME=/var/lib/m13-fixture-caddy/data\nEnvironment=XDG_CONFIG_HOME=/var/lib/m13-fixture-caddy/config\nReadWritePaths=/var/lib/m13-fixture-caddy\n" + common)
    for name in ("api", "caddy", "pg"):
        run(["systemd-analyze", "verify", str(UNITDIR / ("m13-fixture-" + name + ".service"))])
    run(["systemctl", "daemon-reload"])
    service("start", "api")
    service("start", "caddy")
    healthy()
    require("real_services_ready", True)
    require("api_restricted_database_role", sql("SELECT count(*)>0 FROM pg_stat_activity WHERE datname=current_database() AND usename='meshalot';") == "t")
    for name, uid in (("api", 20002), ("caddy", 20003), ("pg", 20001)):
        pid = int(run(["systemctl", "show", "m13-fixture-" + name + ".service", "--property=MainPID", "--value"]).stdout)
        status = Path("/proc/" + str(pid) + "/status").read_text().splitlines()
        require(name + "_restricted_uid", next(line.split()[1] for line in status if line.startswith("Uid:")) == str(uid))
        require(name + "_no_new_privileges", "NoNewPrivs:\t1" in status)
    phase = "customer_workflow"
    anonymous = opener()
    request(anonymous, "GET", "/v1/auth/me", expected=401)
    request(anonymous, "POST", "/v1/auth/login", {"email":"fixture-a@example.invalid","password":"wrong"}, expected=401)
    jar_a, jar_b = CookieJar(), CookieJar()
    client_a, client_b = opener(jar_a), opener(jar_b)
    for client, email, password in ((client_a,"fixture-a@example.invalid","public-fixture-password-a"),(client_b,"fixture-b@example.invalid","public-fixture-password-b")):
        request(client, "POST", "/v1/auth/login", {"email":email,"password":password})
    require("https_session_cookie_secure", len(list(jar_a)) == 1 and all(c.secure and c.has_nonstandard_attr("HttpOnly") for c in jar_a))
    public = json.loads(run([str(ROOT / "payload/probe"), "public"]).stdout)
    code = request(client_a, "POST", "/v1/account/enrollment-codes", {}, expected=201)["enrollment_code"]
    enrollment = {"node_id":public["node_id"],"public_key":public["public_key"],"agent_version":"public-systemd-fixture","enrollment_code":code}
    request(anonymous, "POST", "/v1/agent/enroll", enrollment)
    request(anonymous, "POST", "/v1/agent/enroll", enrollment, expected=401)
    require("enrollment_and_replay_denial", True)
    heartbeat("initial")
    require("account_a_node", [n["node_id"] for n in request(client_a,"GET","/v1/account/nodes")["nodes"]] == [NODE])
    require("account_b_isolated", not request(client_b,"GET","/v1/account/nodes")["nodes"])
    request(client_b,"GET","/v1/account/nodes/" + NODE, expected=404)
    require("dashboard_online_from_postgres", request(client_a,"GET","/v1/dashboard")["online_nodes"] == 1)
    require("paused_telemetry_persisted", sql("SELECT status||':'||mode||':'||(m13_telemetry->>'manual_pause') FROM node_status;") == "online:away:true")
    identity_before = sql("SELECT node_key||':'||identity_public_key FROM nodes ORDER BY node_key;")
    phase = "api_fault_and_configuration_recovery"
    service("stop", "api")
    drop = UNITDIR / "m13-fixture-api.service.d"
    drop.mkdir()
    broken = drop / "99-deliberate-fault.conf"
    broken.write_text("[Service]\nExecStart=\nExecStart=/nonexistent-m13-synthetic-api\n")
    run(["systemctl", "daemon-reload"])
    service("start", "api", check=False)
    time.sleep(.5)
    require("api_fault_detected", service("is-failed", "api", check=False).stdout.strip() == "failed")
    request(anonymous,"GET","/v1/health", expected=502)
    broken.unlink()
    run(["systemctl", "daemon-reload"])
    service("reset-failed", "api")
    service("start", "api")
    healthy()
    require("session_survives_api_recovery", request(client_a,"GET","/v1/auth/me")["user"]["id"] == USER_A)
    heartbeat("reconnect")
    phase = "postgres_restart_and_fail_closed_startup"
    service("stop", "api")
    service("stop", "pg")
    service("start", "api", check=False)
    time.sleep(.5)
    deadline = time.monotonic() + 10
    while service("is-active", "api", check=False).stdout.strip() == "active" and time.monotonic() < deadline:
        time.sleep(.2)
    require("api_refuses_missing_database", service("is-active", "api", check=False).stdout.strip() != "active")
    service("start", "pg")
    service("reset-failed", "api")
    deadline = time.monotonic() + 20
    while True:
        try:
            sql("SELECT count(*) FROM users;")
            break
        except subprocess.SubprocessError:
            if time.monotonic() >= deadline:
                raise
            time.sleep(.2)
    service("start", "api")
    healthy()
    require("session_survives_postgres_restart", request(client_a,"GET","/v1/auth/me")["user"]["id"] == USER_A)
    # Only the valid reconnect is repeated after this material service restart.
    heartbeat("reconnect")
    require("migration_ledger_preserved", sql("SELECT version::text||':'||sha256 FROM meshalot_meta.schema_migrations ORDER BY version;") == ledger_before)
    require("synthetic_node_identity_preserved", sql("SELECT node_key||':'||identity_public_key FROM nodes ORDER BY node_key;") == identity_before)
    require("fixture_rows_preserved", sql("SELECT count(*)::text FROM users;") == "2" and sql("SELECT count(*)::text FROM nodes;") == "1")
    require("final_dashboard_online", request(client_a,"GET","/v1/dashboard")["online_nodes"] == 1)
    request(client_a,"POST","/v1/auth/logout", expected=204)
    request(client_a,"GET","/v1/auth/me", expected=401)
    request(client_b,"POST","/v1/auth/logout", expected=204)
    require("logout_revokes_session", True)
    phase = "service_shutdown"
    for name in ("caddy", "api", "pg"):
        service("stop", name)
        require(name + "_stopped", service("is-active", name, check=False).stdout.strip() == "inactive")
    run(["umount", str(SEED)])
    result = "PASS"
except (OSError, ValueError, KeyError, RuntimeError, subprocess.SubprocessError) as error:
    checks["error_class"] = type(error).__name__
finally:
    report = {"status":result,"phase":phase,"checks":checks,"application_commit":APP,
              "elapsed_seconds":round(time.monotonic() - started,3),
              "failure":last_failure if result != "PASS" else None,
              "service_failure":service_failure if result != "PASS" else None,
              "real_archive_recovery":"NOT_TESTED","gate_acceptance":"NOT_ASSESSED"}
    with open("/dev/ttyS0", "w") as output:
        output.write("\nM13_G17_SERVICE_GUEST_JSON=" + json.dumps(report,sort_keys=True) + "\n")
        output.flush()
    os.sync()
    run(["systemctl","poweroff"],check=False)
