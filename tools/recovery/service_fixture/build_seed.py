"""Build public, synthetic NoCloud seed files; no operator input."""
import json
from pathlib import Path

root = Path(__file__).resolve().parent
config = {
    "hostname": "meshalot-service-fixture",
    "manage_etc_hosts": True,
    "users": [],
    "disable_root": True,
    "ssh_pwauth": False,
    "package_update": False,
    "package_upgrade": False,
    "growpart": {"mode": "off"},
    "resize_rootfs": False,
    "write_files": [{
        "path": "/opt/meshalot-stage/guest.py",
        "permissions": "0700",
        "owner": "root:root",
        "content": (root / "guest.py").read_text(),
    }],
    "runcmd": [["python3", "-B", "/opt/meshalot-stage/guest.py"]],
}
config["write_files"].append({"path": "/opt/meshalot-stage/safetar.py", "permissions": "0600",
                              "owner": "root:root", "content": (root / "safetar.py").read_text()})
(root / "user-data").write_text("#cloud-config\n" + json.dumps(config) + "\n")
(root / "meta-data").write_text(
    "instance-id: meshalot-service-fixture-v1\nlocal-hostname: meshalot-service-fixture\n")
(root / "network-config").write_text("version: 2\nethernets: {}\n")
