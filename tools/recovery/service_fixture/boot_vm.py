"""Unprivileged QEMU TCG boot inside the network-none containment container."""
import json
import os
import selectors
import subprocess
import time

PREFIX = b"M13_G17_SERVICE_GUEST_JSON="
BASE = "/opt/stage/cloud.img"
info = json.loads(subprocess.check_output(
    ["qemu-img", "info", "--output=json", BASE], timeout=10))
assert info["format"] == "qcow2" and "backing-filename" not in info
assert 0 < info["virtual-size"] <= 8 * 1024 ** 3
assert os.getuid() == 65534
subprocess.run(["qemu-img", "create", "-q", "-f", "qcow2", "-F", "qcow2",
                "-b", BASE, "/tmp/stage.qcow2"], check=True, timeout=10)
print("M13_G17_SERVICE_QEMU_VERSION=" + subprocess.check_output(
    ["qemu-system-x86_64", "--version"], text=True, timeout=5).splitlines()[0], flush=True)
command = ["qemu-system-x86_64", "-machine", "pc", "-accel", "tcg,thread=multi",
           "-cpu", "max", "-smp", "2", "-m", "2048", "-nodefaults",
           "-display", "none", "-monitor", "none", "-serial", "stdio",
           "-nic", "none", "-no-reboot",
           "-drive", "file=/tmp/stage.qcow2,if=virtio,format=qcow2",
           "-drive", "file=/opt/stage/seed.iso,if=virtio,format=raw,readonly=on"]
started = time.monotonic()
data = bytearray()
process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
try:
    with selectors.DefaultSelector() as selector:
        selector.register(process.stdout, selectors.EVENT_READ)
        while selector.get_map():
            if time.monotonic() - started > 600:
                raise TimeoutError("guest_boot_budget")
            for key, _ in selector.select(timeout=1):
                chunk = os.read(key.fileobj.fileno(), 65536)
                if not chunk:
                    selector.unregister(key.fileobj)
                else:
                    data.extend(chunk)
                    if len(data) > 2 * 1024 ** 2:
                        raise ValueError("guest_serial_budget")
    assert process.wait(timeout=5) == 0
    reports = [line[len(PREFIX):] for line in bytes(data).splitlines() if line.startswith(PREFIX)]
    assert len(reports) == 1
    report = json.loads(reports[0])
    assert isinstance(report, dict) and report.get("status") == "PASS"
    required = {"guest_pid1_systemd", "guest_no_external_interface", "guest_no_ipv4_external_route",
                "guest_no_ipv6_external_route", "postgres_major_16", "schema_10_complete",
                "owner_and_runtime_separate", "rooted_pg_unit_static_validation",
                "real_services_ready", "api_restricted_database_role",
                "api_restricted_uid", "caddy_restricted_uid", "pg_restricted_uid",
                "api_no_new_privileges", "caddy_no_new_privileges", "pg_no_new_privileges",
                "https_session_cookie_secure",
                "enrollment_and_replay_denial", "signed_wss_initial", "wss_check_count_initial",
                "account_a_node", "account_b_isolated", "dashboard_online_from_postgres",
                "paused_telemetry_persisted", "api_fault_detected", "session_survives_api_recovery",
                "api_refuses_missing_database", "session_survives_postgres_restart", "signed_wss_reconnect",
                "migration_ledger_preserved", "synthetic_node_identity_preserved", "fixture_rows_preserved",
                "final_dashboard_online", "logout_revokes_session", "api_stopped", "caddy_stopped", "pg_stopped"}
    assert required <= set(report["checks"]) and all(value is True for value in report["checks"].values())
    print(PREFIX.decode() + json.dumps(report, sort_keys=True), flush=True)
    print("M13_G17_SERVICE_GUEST_BOOT_AND_SHUTDOWN_SECONDS=" + str(round(time.monotonic() - started, 3)), flush=True)
    print("M13_G17_SERVICE_VM_POWEROFF=PASS", flush=True)
except (OSError, ValueError, AssertionError, TimeoutError, subprocess.SubprocessError) as error:
    # Only classifications: guest logs are deliberately not printed.
    print("M13_G17_SERVICE_BOOT=FAIL_" + type(error).__name__, flush=True)
    for marker, phrase in {"KERNEL_SEEN": b"Linux version", "SYSTEMD_SEEN": b"systemd[",
                           "CLOUD_INIT_SEEN": b"Cloud-init", "NO_BOOT_DEVICE": b"No bootable device"}.items():
        print("M13_G17_SERVICE_" + marker + "=" + str(phrase in data), flush=True)
    # Safe probe names help distinguish a real failed check from missing guest execution.
    for line in bytes(data).splitlines():
        if line.startswith(PREFIX):
            parsed = json.loads(line[len(PREFIX):])
            print(PREFIX.decode() + json.dumps(parsed, sort_keys=True), flush=True)
    raise SystemExit(1)
finally:
    if process.poll() is None:
        process.kill()
        process.wait(timeout=5)
