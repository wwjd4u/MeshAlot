"""Unprivileged QEMU TCG boot inside the network-none containment container."""
import json
import os
import selectors
import subprocess
import time

PREFIX = b"M13_G17_SYSTEMD_GUEST_JSON="
BASE = "/opt/stage/cloud.img"
info = json.loads(subprocess.check_output(
    ["qemu-img", "info", "--output=json", BASE], timeout=10))
assert info["format"] == "qcow2" and "backing-filename" not in info
assert 0 < info["virtual-size"] <= 8 * 1024 ** 3
assert os.getuid() == 65534
subprocess.run(["qemu-img", "create", "-q", "-f", "qcow2", "-F", "qcow2",
                "-b", BASE, "/tmp/stage.qcow2"], check=True, timeout=10)
print("M13_G17_STAGE_QEMU_VERSION=" + subprocess.check_output(
    ["qemu-system-x86_64", "--version"], text=True, timeout=5).splitlines()[0], flush=True)
command = ["qemu-system-x86_64", "-machine", "pc", "-accel", "tcg,thread=multi",
           "-cpu", "max", "-smp", "2", "-m", "1024", "-nodefaults",
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
            if time.monotonic() - started > 360:
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
    assert len(report["checks"]) == 16 and all(value is True for value in report["checks"].values())
    print(PREFIX.decode() + json.dumps(report, sort_keys=True), flush=True)
    print("M13_G17_STAGE_GUEST_BOOT_AND_SHUTDOWN_SECONDS=" + str(round(time.monotonic() - started, 3)), flush=True)
    print("M13_G17_STAGE_VM_POWEROFF=PASS", flush=True)
except (OSError, ValueError, AssertionError, TimeoutError, subprocess.SubprocessError) as error:
    # Only classifications: guest logs are deliberately not printed.
    print("M13_G17_STAGE_BOOT=FAIL_" + type(error).__name__, flush=True)
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
