"""Trusted MS-02 only: pinned synthetic M13 service integration fixture.

This is intentionally not a general-purpose VM launcher or archive runner.
No paths, endpoints, credentials or command overrides are accepted from users.
"""
import hashlib
import io
import json
import os
from pathlib import Path
import platform
import pwd
import shutil
import signal
import subprocess
import tempfile
import tarfile
import time
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parent
CLOUD_URL = "https://cloud-images.ubuntu.com/releases/noble/release-20261017/ubuntu-24.04-server-cloudimg-amd64.img"
CLOUD_SHA256 = "04e8f2461a16a44e78a370284471f29616cdd9baf1893df25807b71f985b6c94"
BASE_IMAGE = "ubuntu@sha256:fe81560990fae86dc4743202b2cb4d3df98d0bcf949173cecef15ce1127ef852"
FILES = ("prepare_image.sh", "build_seed.py", "guest.py", "safetar.py", "boot_vm.py")
APP = "4f67ee54e5ebca084e38e306e723ae0acd08e77d"
PG_IMAGE = "postgres@sha256:cbf42f0edc946c0bc1b8eeed62403734e1f25dc67a3491455d07a92fdb558b28"
CADDY_IMAGE = "caddy@sha256:040e9f7480b80b6d4a7e5013a21159b950a63dcbdb956e38abe2387fb28d9ec0"
GO_IMAGE = "golang@sha256:757779acac4af1b349a20f357c7296097b4a0b89da4ad0e370b339060077282a"
DOCKER = ["docker", "--host=unix:///var/run/docker.sock"]
LIMIT = 5 * 1024 ** 3
environment = dict(os.environ)
for selector in ("DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_TLS", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH"):
    environment.pop(selector, None)


def docker(*args, timeout=15, check=True):
    result = subprocess.run(DOCKER + list(args), env=environment, capture_output=True,
                            text=True, timeout=timeout)
    if check and result.returncode:
        # Fixed classifications only: never dump Docker config, URLs carrying
        # credentials, raw package output, private paths or the process env.
        operation = args[0] + ("_" + args[1] if args[0] in ("image", "container") else "")
        emit("FAILED_DOCKER_OPERATION", operation)
        emit("FAILED_DOCKER_EXIT", result.returncode)
        for line in result.stdout.splitlines():
            if line.startswith(("M13_G17_SERVICE_ASSEMBLY_", "M13_G17_SERVICE_BUILD_")):
                print(line, flush=True)
        lower = (result.stderr + result.stdout).lower()
        categories = {
            "manifest unknown": "MANIFEST_UNKNOWN", "no such image": "IMAGE_NOT_FOUND",
            "invalid reference format": "INVALID_REFERENCE", "unauthorized": "UNAUTHORIZED",
            "denied": "ACCESS_DENIED", "toomanyrequests": "REGISTRY_RATE_LIMIT",
            "pull rate limit": "REGISTRY_RATE_LIMIT", "too many requests": "REGISTRY_RATE_LIMIT",
            "certificate": "TLS_CERTIFICATE", "no such file": "FILE_NOT_FOUND",
            "could not find": "FILE_NOT_FOUND", "not found": "NOT_FOUND",
            "permission": "PERMISSION", "network is unreachable": "NETWORK_UNREACHABLE",
            "temporary failure": "TEMPORARY_FAILURE", "error resolving": "NAME_RESOLUTION",
            "unsupported": "UNSUPPORTED", "unknown flag": "UNKNOWN_FLAG",
            "operation not permitted": "PERMISSION", "oci runtime": "CONTAINER_RUNTIME",
            "cgroup": "CGROUP", "apparmor": "APPARMOR", "seccomp": "SECCOMP",
            "not signed": "APT_SIGNATURE", "signature": "APT_SIGNATURE",
            "no_pubkey": "APT_SIGNING_KEY", "no space left": "DISK_SPACE",
            "could not resolve": "NAME_RESOLUTION", "dpkg": "PACKAGE_INSTALL",
            "logging driver": "LOGGING_DRIVER", "compression": "LOG_COMPRESSION",
        }
        found = sorted({label for fragment, label in categories.items() if fragment in lower})
        emit("FAILED_DOCKER_REASON_CLASSES", ",".join(found) or "UNCLASSIFIED")
        result.check_returncode()
    return result


def emit(name, value):
    print("M13_G17_SERVICE_" + name + "=" + str(value), flush=True)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *args, **kwargs):
        raise ValueError("Unexpected cloud source redirect")


def download(path):
    digest = hashlib.sha256()
    size = 0
    started = time.monotonic()
    opener = urllib.request.build_opener(NoRedirect())
    with opener.open(CLOUD_URL, timeout=20) as response, path.open("xb") as target:
        while True:
            block = response.read(1024 ** 2)
            if not block:
                break
            size += len(block)
            if size > 800 * 1024 ** 2 or time.monotonic() - started > 240:
                raise ValueError("Cloud download budget")
            digest.update(block)
            target.write(block)
    assert digest.hexdigest() == CLOUD_SHA256
    emit("CLOUD_IMAGE_SHA256", CLOUD_SHA256)
    emit("CLOUD_IMAGE_BYTES", size)


def validate_container(config, image_id):
    host = config["HostConfig"]
    assert host["NetworkMode"] == "none" and not host.get("PortBindings")
    assert not host.get("Privileged") and host["CapDrop"] == ["ALL"]
    assert "no-new-privileges" in host["SecurityOpt"]
    assert host["ReadonlyRootfs"] and config["Config"]["User"] == "65534:65534"
    assert host["NanoCpus"] == 2 * 10 ** 9
    assert host["Memory"] == LIMIT and host["MemorySwap"] == LIMIT
    assert host["PidsLimit"] == 64
    assert not host.get("Binds") and not host.get("Devices") and not config.get("Mounts")
    assert host["RestartPolicy"]["Name"] == "no"
    assert config["Image"] == image_id
    assert host["Tmpfs"] == {"/tmp": "rw,noexec,nosuid,nodev,size=1610612736,mode=1777"}
    assert config["Config"]["Labels"].get("meshalot.g17.service-fixture") == name


def sha(path):
    value = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 ** 2), b""):
            value.update(block)
    return value.hexdigest()


def prepare_assets(context, scratch):
    pins = json.loads((ROOT / "app-source.json").read_text())
    assert pins["application_commit"] == APP and len(pins["files"]) == 49
    source = scratch / "source"
    source.mkdir()
    url = "https://codeload.github.com/wwjd4u/MeshAlot/tar.gz/" + APP
    with urllib.request.build_opener(NoRedirect()).open(url, timeout=30) as response:
        data = response.read(64 * 1024 ** 2 + 1)
    assert len(data) <= 64 * 1024 ** 2
    found = set()
    with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as archive:
        for item in archive:
            top, separator, relative = item.name.partition("/")
            if relative not in pins["files"]:
                continue
            assert top == "MeshAlot-" + APP and item.isfile() and item.size <= 1024 ** 2
            assert relative not in found
            value = archive.extractfile(item).read(1024 ** 2 + 1)
            assert hashlib.sha256(value).hexdigest() == pins["files"][relative]
            target = source / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(value)
            found.add(relative)
    assert found == set(pins["files"])
    emit("APPLICATION_SOURCE_FILES", len(found))
    emit("APPLICATION_COMMIT", APP)
    payload = context / "payload"
    payload.mkdir()
    shutil.copytree(source / "database", payload / "database")
    shutil.copyfile(ROOT / "probe.go", source / "stage-probe.go")
    # A digest-pinned compiler receives only reviewed public source. It has a
    # hard cgroup budget and no host mounts, checkout, credentials or host cache.
    if docker("image", "inspect", GO_IMAGE, check=False).returncode:
        docker("pull", "--quiet", "--platform", "linux/amd64", GO_IMAGE, timeout=180)
    script = """set -eu
    phase=API_BUILD
    trap 'rc=$?; printf "M13_G17_SERVICE_BUILD_PHASE=%s\\nM13_G17_SERVICE_BUILD_EXIT=%s\\n" "$phase" "$rc"' EXIT
    printf 'M13_G17_SERVICE_COMPILER='; go version
    mkdir /out
    go build -p=2 -mod=readonly -buildvcs=false -trimpath -ldflags='-s -w' -o /out/api ./server/cmd/meshalot-server
    phase=PROBE_BUILD
    go build -p=2 -mod=readonly -buildvcs=false -trimpath -ldflags='-s -w' -o /out/probe stage-probe.go
    phase=COMPLETE
    """
    docker("create", "--name", compiler, "--label", "meshalot.g17.service-fixture=" + name,
           "--network", "bridge", "--security-opt", "no-new-privileges", "--cpus", "2",
           "--memory", "3g", "--memory-swap", "3g", "--pids-limit", "128",
           "--log-driver", "local", "--log-opt", "max-size=1m", "--log-opt", "max-file=1",
           "--log-opt", "compress=false", "--workdir", "/work",
           "--env", "CGO_ENABLED=0", "--env", "GOOS=linux", "--env", "GOARCH=amd64",
           "--env", "GOTOOLCHAIN=local", "--env", "GOENV=off", "--env", "GOMAXPROCS=2",
           "--env", "GOMEMLIMIT=512MiB", GO_IMAGE, "/bin/sh", "-ec", script)
    compiler_meta = json.loads(docker("inspect", compiler).stdout)[0]
    assert not compiler_meta.get("Mounts") and not compiler_meta["HostConfig"].get("Binds")
    assert not compiler_meta["HostConfig"]["Privileged"] and not compiler_meta["HostConfig"].get("Devices")
    assert compiler_meta["HostConfig"]["NanoCpus"] == 2 * 10 ** 9
    assert compiler_meta["HostConfig"]["Memory"] == 3 * 1024 ** 3
    docker("cp", str(source) + "/.", compiler + ":/work", timeout=30)
    compiled = docker("start", "--attach", compiler, timeout=600)
    assert json.loads(docker("inspect", compiler).stdout)[0]["State"]["ExitCode"] == 0
    for line in compiled.stdout.splitlines():
        if line.startswith("M13_G17_SERVICE_COMPILER="):
            print(line, flush=True)
    for target in ("api", "probe"):
        docker("cp", compiler + ":/out/" + target, str(payload / target), timeout=20)
    emit("STATIC_LINUX_BUILD", "PASS")
    for image, container in ((PG_IMAGE, pg_source), (CADDY_IMAGE, caddy_source)):
        if docker("image", "inspect", image, check=False).returncode:
            docker("pull", "--quiet", "--platform", "linux/amd64", image, timeout=180)
        docker("create", "--name", container, "--label", "meshalot.g17.service-fixture=" + name,
               "--network", "none", "--read-only", "--entrypoint", "/bin/true", image)
        metadata = json.loads(docker("inspect", container).stdout)[0]
        assert metadata["State"]["Status"] == "created" and not metadata["HostConfig"].get("Binds")
        assert all(m["Type"] == "volume" and m.get("Name") for m in metadata.get("Mounts", []))
        owned_volumes.update(m["Name"] for m in metadata.get("Mounts", []))
    # These fresh source containers NEVER start. No existing container is read.
    docker("export", "--output", str(payload / "pg-rootfs.tar"), pg_source, timeout=90)
    assert (payload / "pg-rootfs.tar").stat().st_size <= 1024 ** 3
    docker("cp", caddy_source + ":/usr/bin/caddy", str(payload / "caddy"), timeout=45)
    manifest = {"application_commit": APP, "material": "public_synthetic_only",
                "postgres_image": PG_IMAGE, "caddy_image": CADDY_IMAGE,
                "files": {str(p.relative_to(payload)): sha(p) for p in sorted(payload.rglob("*")) if p.is_file()}}
    (context / "assets.json").write_text(json.dumps(manifest, sort_keys=True) + "\n")
    emit("PUBLIC_ASSET_MANIFEST_SHA256", sha(context / "assets.json"))
    emit("POSTGRES_OCI", PG_IMAGE)
    emit("CADDY_OCI", CADDY_IMAGE)
    for target in ("api", "probe", "caddy", "pg-rootfs.tar"):
        emit("ARTIFACT_SHA256_" + target.upper().replace(".", "_"), manifest["files"][target])


assert os.environ.get("RUNNER_NAME") == "meshalot-ms02"
assert os.environ.get("GITHUB_REPOSITORY") == "wwjd4u/MeshAlot"
assert os.environ.get("GITHUB_REF") == "refs/heads/infra/ms02-runner-gateway"
assert platform.system() == "Linux" and platform.machine() == "x86_64"
assert platform.node().split(".")[0] == "wwjd4u-MS-02-Ultra"
assert pwd.getpwuid(os.geteuid()).pw_name == "jason_guynes"
assert Path.home() == Path("/home/jason_guynes")
assert not os.environ.get("DOCKER_HOST") and not os.environ.get("DOCKER_CONTEXT")
memory = next(int(line.split()[1]) for line in Path("/proc/meminfo").read_text().splitlines()
              if line.startswith("MemAvailable:"))
assert memory >= 10 * 1024 ** 2 and (os.cpu_count() or 0) >= 4
assert shutil.disk_usage("/tmp").free >= 12 * 1024 ** 3
docker("version", "--format", "{{.Server.Version}}")
emit("TRUSTED_RUNNER_AND_ALLOCATION_PREFLIGHT", "PASS")
emit("SCOPE", "PUBLIC_SYNTHETIC_M13_SERVICE_FIXTURE_ONLY")
os.umask(0o077)
name = "m13-g17-svc-" + uuid.uuid4().hex[:12]
builder = name + "-assembly"
pg_source, caddy_source = name + "-pg-source", name + "-caddy-source"
compiler = name + "-compiler"
tag = "meshalot-g17-service:" + name
passed = False
image_built = False
cleanup_ok = True
owned_volumes = set()


def interrupted(signum, frame):
    raise RuntimeError("stage_interrupted")


signal.signal(signal.SIGTERM, interrupted)
try:
    with tempfile.TemporaryDirectory(prefix="m13-svc-", dir="/tmp") as directory:
        context = Path(directory)
        for filename in FILES:
            shutil.copyfile(ROOT / filename, context / filename)
        with tempfile.TemporaryDirectory(prefix="m13-svc-build-", dir="/tmp") as scratch:
            prepare_assets(context, Path(scratch))
        download(context / "cloud.img")
        # Assemble in a named bounded container, rather than an anonymous build:
        # its process can be forcibly stopped and removed even on timeout.
        # Only public Ubuntu repositories need network during package assembly.
        emit("PUBLIC_PACKAGE_ASSEMBLY", "STARTING")
        if docker("image", "inspect", BASE_IMAGE, check=False).returncode:
            docker("pull", "--quiet", BASE_IMAGE, timeout=180)
        docker("create", "--name", builder,
               "--label", "meshalot.g17.service-fixture=" + name,
               "--network", "bridge", "--security-opt", "no-new-privileges",
               "--cpus", "2", "--memory", "5g", "--memory-swap", "5g",
               "--pids-limit", "128", "--restart", "no", "--workdir", "/opt/stage",
               "--log-driver", "local", "--log-opt", "max-size=1m", "--log-opt", "max-file=1",
               "--log-opt", "compress=false",
               BASE_IMAGE, "/bin/sh", "/opt/stage/prepare_image.sh")
        assembly = json.loads(docker("inspect", builder).stdout)[0]
        assert not assembly["HostConfig"]["Privileged"] and not assembly.get("Mounts")
        assert not assembly["HostConfig"].get("Binds") and not assembly["HostConfig"].get("Devices")
        assert not assembly["HostConfig"].get("PortBindings")
        assert assembly["HostConfig"]["NanoCpus"] == 2 * 10 ** 9
        assert assembly["HostConfig"]["Memory"] == LIMIT and assembly["HostConfig"]["MemorySwap"] == LIMIT
        assert "no-new-privileges" in assembly["HostConfig"]["SecurityOpt"]
        docker("cp", str(context) + "/.", builder + ":/opt/stage", timeout=90)
        built = docker("start", "--attach", builder, timeout=600)
        for line in built.stdout.splitlines():
            if line.startswith("M13_G17_SERVICE_PACKAGE="):
                print(line, flush=True)
        assert json.loads(docker("inspect", builder).stdout)[0]["State"]["ExitCode"] == 0
        docker("commit", "--change", "USER 65534:65534",
               "--change", 'ENTRYPOINT ["python3", "-B", "/opt/stage/boot_vm.py"]',
               "--change", "CMD []", builder, tag, timeout=90)
        image_built = True
        image_id = json.loads(docker("image", "inspect", tag).stdout)[0]["Id"]
        emit("TOOLING_IMAGE_ID", image_id)
        emit("PUBLIC_PACKAGE_ASSEMBLY", "PASS")
        docker("create", "--name", name,
               "--label", "meshalot.g17.service-fixture=" + name,
               "--network", "none", "--cap-drop", "ALL", "--security-opt", "no-new-privileges",
               "--read-only", "--user", "65534:65534", "--cpus", "2",
               "--memory", "5g", "--memory-swap", "5g", "--pids-limit", "64",
               "--tmpfs", "/tmp:rw,noexec,nosuid,nodev,size=1610612736,mode=1777",
               "--shm-size", "16m", "--restart", "no",
               "--log-driver", "local", "--log-opt", "max-size=1m", "--log-opt", "max-file=1",
               "--log-opt", "compress=false", tag)
        validate_container(json.loads(docker("inspect", name).stdout)[0], image_id)
        emit("CONTAINER_BOUNDARIES_BEFORE_START", "PASS")
        result = docker("start", "--attach", name, timeout=660, check=False)
        # Only explicit sanitized probe output crosses back into shared logs.
        for line in result.stdout.splitlines():
            if line.startswith("M13_G17_SERVICE_") or line.startswith("M13_G17_SERVICE_GUEST_JSON="):
                print(line, flush=True)
        config = json.loads(docker("inspect", name).stdout)[0]
        validate_container(config, image_id)
        assert result.returncode == 0 and config["State"]["ExitCode"] == 0
        assert not config["State"]["Running"] and not config["State"]["OOMKilled"]
        assert "M13_G17_SERVICE_VM_POWEROFF=PASS" in result.stdout.splitlines()
        emit("CONTAINER_BOUNDARIES_AFTER_EXIT", "PASS")
        passed = True
except (OSError, ValueError, AssertionError, RuntimeError, subprocess.SubprocessError) as error:
    emit("RESULT", "FAIL_" + type(error).__name__)
finally:
    try:
        # Query successful list results rather than treating any inspect error
        # (which could be a dead daemon) as proof of absence.
        owned = docker("container", "ls", "--all", "--filter",
                       "label=meshalot.g17.service-fixture=" + name, "--format", "{{.Names}}")
        names = owned.stdout.splitlines()
        assert set(names) <= {name, builder, pg_source, caddy_source, compiler}
        for container in names:
            docker("rm", "--force", "--volumes", container)
        assert not docker("container", "ls", "--all", "--filter",
                          "label=meshalot.g17.service-fixture=" + name, "--format", "{{.Names}}").stdout.strip()
        # Only image-declared anonymous volumes of our never-started source
        # containers are removed by rm --volumes. Verify their absence too.
        remaining = set(docker("volume", "ls", "--quiet").stdout.splitlines())
        assert not owned_volumes.intersection(remaining)
        # The tag is unique to this attempt. Never prune or force-remove images.
        image = docker("image", "ls", "--filter", "reference=" + tag, "--format", "{{.ID}}")
        if image.stdout.strip():
            meta = json.loads(docker("image", "inspect", tag).stdout)[0]
            assert meta["Config"]["Labels"].get("meshalot.g17.service-fixture") == name
            docker("image", "rm", tag)
        assert not docker("image", "ls", "--filter", "reference=" + tag, "--format", "{{.ID}}").stdout.strip()
    except (OSError, ValueError, AssertionError, RuntimeError, subprocess.SubprocessError):
        cleanup_ok = False
    emit("TASK_CONTAINER_AND_TOOLING_IMAGE_CLEANUP", "PASS" if cleanup_ok else "FAIL")
    emit("PRODUCTION_CONNECTION", "NONE")
    emit("REAL_ARCHIVES_OR_KEYS", "NOT_USED")
    emit("HOST_PACKAGES_SERVICES_OR_IDENTITIES_CHANGED", "NONE")
    emit("FULL_ARCHIVED_ROLLBACK", "NOT_TESTED")
    emit("GATE_ACCEPTANCE", "NOT_ASSESSED")
    emit("RESULT", "PASS" if passed and cleanup_ok else "FAIL")
raise SystemExit(0 if passed and cleanup_ok else 1)
