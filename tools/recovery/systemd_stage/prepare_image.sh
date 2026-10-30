#!/bin/sh
# Runs as root in the named disposable package-assembly container only.
set -eu
stage_phase=CONTAINER_GUARD
trap 'stage_rc=$?; printf "M13_G17_STAGE_ASSEMBLY_PHASE=%s\nM13_G17_STAGE_ASSEMBLY_EXIT=%s\n" "$stage_phase" "$stage_rc"' EXIT
test -f /.dockerenv
test "$(pwd)" = /opt/stage
test "$(id -u)" = 0
export DEBIAN_FRONTEND=noninteractive
stage_phase=APT_INDEX
apt-get update -qq
stage_phase=APT_PACKAGES
apt-get install -y -qq --no-install-recommends \
    python3 qemu-system-x86 qemu-utils genisoimage seabios
rm -rf /var/lib/apt/lists/*
stage_phase=NOCLOUD_SEED
python3 -B build_seed.py
genisoimage -quiet -output seed.iso -volid cidata -joliet -rock \
    user-data meta-data network-config
rm user-data meta-data network-config
stage_phase=IMAGE_DIGEST
echo '04e8f2461a16a44e78a370284471f29616cdd9baf1893df25807b71f985b6c94  cloud.img' | sha256sum -c -
chmod -R a+rX /opt/stage
stage_phase=COMPLETE
dpkg-query -W -f='M13_G17_STAGE_PACKAGE=${Package}=${Version}\n' \
    python3 qemu-system-x86 qemu-utils genisoimage seabios
