"""Run only inside the disposable guest. Never an API/archive acceptance test."""
import hashlib
import ipaddress
import json
import os
from pathlib import Path
import pwd
import subprocess
import time
import urllib.error
import urllib.request

PREFIX = "M13_G17_SYSTEMD_GUEST_JSON="
UNIT = "meshalot-synthetic-probe.service"
STATE = Path("/var/lib/meshalot-synthetic-probe")
SCRIPT = Path("/opt/meshalot-stage/daemon.py")
UNIT_PATH = Path("/etc/systemd/system") / UNIT
DROP = Path(str(UNIT_PATH) + ".d")
FIXTURE = b"public-synthetic-recovery-fixture-v1\n"
TOKEN = "public-synthetic-authorization-only"
checks = {}
started = time.monotonic()


def run(*args, check=True):
    return subprocess.run(args, capture_output=True, text=True, timeout=15, check=check)


def require(name, value):
    checks[name] = bool(value)
    if not value:
        raise RuntimeError(name)


def request(authorized=True):
    headers = {"Authorization": "Bearer " + TOKEN} if authorized else {}
    request = urllib.request.Request("http://127.0.0.1:38417/", headers=headers)
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    with opener.open(request, timeout=2) as response:
        return json.loads(response.read(4096))


def healthy():
    deadline = time.monotonic() + 25
    while time.monotonic() < deadline:
        try:
            return request()
        except (OSError, ValueError):
            time.sleep(0.2)
    raise RuntimeError("service_deadline")


try:
    require("guest_pid1_systemd", Path("/proc/1/comm").read_text().strip() == "systemd")
    require("guest_no_external_interface", sorted(p.name for p in Path("/sys/class/net").iterdir()) == ["lo"])
    for family in (4, 6):
        routes = json.loads(run("ip", "-j", "-" + str(family), "route", "show", "table", "all").stdout)
        require("guest_no_ipv" + str(family) + "_external_route", all(
            route.get("dev") == "lo" and ipaddress.ip_network(route["dst"], strict=False).is_loopback
            for route in routes))
    STATE.mkdir(mode=0o700)
    nobody = pwd.getpwnam("nobody")
    os.chown(STATE, nobody.pw_uid, nobody.pw_gid)
    (STATE / "fixture").write_bytes(FIXTURE)
    os.chown(STATE / "fixture", nobody.pw_uid, nobody.pw_gid)
    (STATE / "fixture").chmod(0o600)
    SCRIPT.write_text('''import hashlib, http.server, json, os
from pathlib import Path
state = Path("/var/lib/meshalot-synthetic-probe")
counter = state / "starts"
n = int(counter.read_text()) + 1 if counter.exists() else 1
counter.write_text(str(n))
class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.headers.get("Authorization") != "Bearer public-synthetic-authorization-only":
            self.send_response(403); self.end_headers(); return
        payload = json.dumps({"starts": n, "uid": os.getuid(),
            "precedence": os.environ.get("STAGE_PRECEDENCE"),
            "fixture": hashlib.sha256((state / "fixture").read_bytes()).hexdigest()}).encode()
        self.send_response(200); self.end_headers(); self.wfile.write(payload)
    def log_message(self, *args):
        pass
http.server.HTTPServer(("127.0.0.1", 38417), Handler).serve_forever()
''')
    SCRIPT.chmod(0o644)
    UNIT_PATH.write_text("""[Unit]
Description=MeshAlot public synthetic stage prerequisite only
[Service]
Type=simple
User=nobody
Group=nogroup
ExecStart=/usr/bin/python3 -B /opt/meshalot-stage/daemon.py
WorkingDirectory=/var/lib/meshalot-synthetic-probe
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/meshalot-synthetic-probe
Restart=no
""")
    DROP.mkdir(mode=0o755)
    (DROP / "10-baseline.conf").write_text("[Service]\nEnvironment=STAGE_PRECEDENCE=baseline\n")
    (DROP / "20-override.conf").write_text("[Service]\nEnvironment=STAGE_PRECEDENCE=override\n")
    run("systemd-analyze", "verify", str(UNIT_PATH))
    require("unit_static_verification", True)
    run("systemctl", "daemon-reload")
    run("systemctl", "start", UNIT)
    first = healthy()
    require("service_restricted_start", first["uid"] == nobody.pw_uid and first["starts"] == 1)
    require("drop_in_precedence", first["precedence"] == "override")
    pid = run("systemctl", "show", UNIT, "--property=MainPID", "--value").stdout.strip()
    status = Path("/proc/" + str(int(pid)) + "/status").read_text().splitlines()
    require("service_no_new_privileges", "NoNewPrivs:\t1" in status)
    require("service_no_capabilities", "CapEff:\t0000000000000000" in status)
    try:
        request(False)
        denied = False
    except urllib.error.HTTPError as error:
        denied = error.code == 403
    require("synthetic_authorization_denied", denied)
    run("systemctl", "stop", UNIT)
    require("service_stopped", run("systemctl", "is-active", UNIT, check=False).stdout.strip() == "inactive")
    run("systemctl", "start", UNIT)
    second = healthy()
    require("restart_state_preserved", second["starts"] == 2 and second["fixture"] == hashlib.sha256(FIXTURE).hexdigest())
    run("systemctl", "stop", UNIT)
    broken = DROP / "30-deliberate-fault.conf"
    broken.write_text("[Service]\nExecStart=\nExecStart=/nonexistent-synthetic-stage-binary\n")
    run("systemctl", "daemon-reload")
    failed = run("systemctl", "start", UNIT, check=False)
    # Type=simple can return before exec failure; read the resulting state boundedly.
    deadline = time.monotonic() + 10
    while time.monotonic() < deadline:
        if run("systemctl", "is-failed", UNIT, check=False).stdout.strip() == "failed":
            break
        time.sleep(0.1)
    require("deliberate_fault_detected", run("systemctl", "is-failed", UNIT, check=False).stdout.strip() == "failed")
    broken.unlink()
    run("systemctl", "daemon-reload")
    run("systemctl", "reset-failed", UNIT)
    run("systemctl", "start", UNIT)
    restored = healthy()
    require("service_configuration_restored", restored["starts"] == 3 and restored["precedence"] == "override")
    require("fixture_integrity_preserved", restored["fixture"] == hashlib.sha256(FIXTURE).hexdigest())
    run("systemctl", "stop", UNIT)
    require("final_service_stop", run("systemctl", "is-active", UNIT, check=False).stdout.strip() == "inactive")
    result = "PASS"
except (OSError, ValueError, KeyError, RuntimeError, subprocess.SubprocessError) as error:
    result = "FAIL"
    checks["error_class"] = type(error).__name__
finally:
    report = {"status": result, "checks": checks,
              "elapsed_seconds": round(time.monotonic() - started, 3),
              "archive_acceptance": "NOT_TESTED", "gate_acceptance": "NOT_ASSESSED"}
    with open("/dev/ttyS0", "w") as serial:
        serial.write("\n" + PREFIX + json.dumps(report, sort_keys=True) + "\n")
        serial.flush()
    os.sync()
    subprocess.run(["systemctl", "poweroff"], timeout=15, check=False,
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
