"""Build public, synthetic NoCloud seed files; no operator input."""
import json
from pathlib import Path

root = Path(__file__).resolve().parent
config = {
    "hostname": "meshalot-synthetic-stage",
    "manage_etc_hosts": True,
    "users": [],
    "disable_root": True,
    "ssh_pwauth": False,
    "package_update": False,
    "package_upgrade": False,
    "growpart": {"mode": "off"},
    "resize_rootfs": False,
    "write_files": [{
        "path": "/opt/meshalot-stage/guest_probe.py",
        "permissions": "0700",
        "owner": "root:root",
        "content": (root / "guest_probe.py").read_text(),
    }],
    "runcmd": [["python3", "-B", "/opt/meshalot-stage/guest_probe.py"]],
}
(root / "user-data").write_text("#cloud-config\n" + json.dumps(config) + "\n")
(root / "meta-data").write_text(
    "instance-id: meshalot-synthetic-systemd-v1\nlocal-hostname: meshalot-synthetic-stage\n")
(root / "network-config").write_text("version: 2\nethernets: {}\n")
