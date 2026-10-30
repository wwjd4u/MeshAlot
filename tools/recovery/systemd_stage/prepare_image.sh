#!/bin/sh
# Runs as root in the named disposable package-assembly container only.
set -eu
test -f /.dockerenv
test "$(pwd)" = /opt/stage
test "$(id -u)" = 0
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends \
    python3 qemu-system-x86 qemu-utils genisoimage seabios
rm -rf /var/lib/apt/lists/*
python3 -B build_seed.py
genisoimage -quiet -output seed.iso -volid cidata -joliet -rock \
    user-data meta-data network-config
rm user-data meta-data network-config
echo '04e8f2461a16a44e78a370284471f29616cdd9baf1893df25807b71f985b6c94  cloud.img' | sha256sum -c -
chmod -R a+rX /opt/stage
dpkg-query -W -f='M13_G17_STAGE_PACKAGE=${Package}=${Version}\n' \
    python3 qemu-system-x86 qemu-utils genisoimage seabios
