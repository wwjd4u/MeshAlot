"""Bounded extraction of a pinned public rootfs inside an already isolated guest."""
import os
from pathlib import PurePosixPath
import posixpath
import shutil
import stat
import tarfile

def unpack_rootfs(archive, destination):
    # Rootfs symlinks are installed last, with absolute targets remapped inside
    # this guest-only root. Regular-file writes cannot follow archive symlinks.
    def relative(name):
        path = PurePosixPath(name)
        if path.is_absolute() or ".." in path.parts:
            raise ValueError("rootfs_path")
        return str(path)
    destination.mkdir(mode=0o755)
    links = []
    total = 0
    count = 0
    with tarfile.open(archive, "r:") as stream:
        for item in stream:
            count += 1
            total += item.size
            if count > 50000 or total > 1024 ** 3 or item.size > 256 * 1024 ** 2:
                raise ValueError("rootfs_budget")
            name = relative(item.name)
            if name == ".":
                continue
            target = destination / name
            target.parent.mkdir(parents=True, exist_ok=True)
            if item.isdir():
                target.mkdir(exist_ok=True)
                target.chmod(item.mode & 0o1777)
            elif item.isfile():
                with stream.extractfile(item) as source, target.open("xb") as output:
                    shutil.copyfileobj(source, output, 1024 ** 2)
                target.chmod(item.mode & 0o1777)
            elif item.issym() or item.islnk():
                links.append((item, target, name))
            else:
                raise ValueError("rootfs_special_member")
    # Hard links resolve only to existing regular members, never symbolic links.
    for item, target, name in links:
        if item.islnk():
            original = destination / relative(item.linkname)
            if not stat.S_ISREG(original.lstat().st_mode):
                raise ValueError("rootfs_hardlink")
            os.link(original, target)
    for item, target, name in links:
        if item.issym():
            mapped = posixpath.normpath(item.linkname.lstrip("/") if item.linkname.startswith("/")
                                        else posixpath.join(posixpath.dirname(name), item.linkname))
            relative(mapped)
            target.symlink_to(os.path.relpath(destination / mapped, target.parent))
