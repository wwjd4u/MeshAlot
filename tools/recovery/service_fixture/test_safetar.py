import io
from pathlib import Path
import stat
import tarfile
import tempfile
import unittest

from .safetar import unpack_rootfs


class RootfsExtractionTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name)

    def archive(self, entries):
        archive = self.base / "public.tar"
        with tarfile.open(archive, "w:") as stream:
            for name, kind, value in entries:
                member = tarfile.TarInfo(name)
                member.mode = 0o4755
                if kind == "file":
                    member.size = len(value)
                    stream.addfile(member, io.BytesIO(value))
                else:
                    member.type = {"symlink":tarfile.SYMTYPE,"hardlink":tarfile.LNKTYPE,"device":tarfile.CHRTYPE}[kind]
                    member.linkname = value
                    stream.addfile(member)
        return archive

    def test_absolute_symlink_rebased_and_setuid_removed(self):
        archive = self.archive([("bin/program","file",b"public fixture"),("usr/bin/link","symlink","/bin/program")])
        target = self.base / "root"
        unpack_rootfs(archive,target)
        self.assertEqual((target / "usr/bin/link").read_bytes(), b"public fixture")
        self.assertEqual(stat.S_IMODE((target / "bin/program").stat().st_mode), 0o755)
        self.assertEqual((target / "usr/bin/link").resolve(), target / "bin/program")

    def test_parent_traversal_cannot_write_outside(self):
        archive = self.archive([("../outside","file",b"bad")])
        with self.assertRaises(ValueError):
            unpack_rootfs(archive,self.base / "root")
        self.assertFalse((self.base / "outside").exists())

    def test_absolute_member_refused(self):
        archive = self.archive([("/outside","file",b"bad")])
        with self.assertRaises(ValueError):
            unpack_rootfs(archive,self.base / "root")

    def test_escaping_symbolic_link_refused(self):
        archive = self.archive([("link","symlink","../../outside")])
        with self.assertRaises(ValueError):
            unpack_rootfs(archive,self.base / "root")

    def test_hardlink_must_resolve_to_regular_member(self):
        archive = self.archive([("real","file",b"public"),("alias","hardlink","real")])
        target = self.base / "root"
        unpack_rootfs(archive,target)
        self.assertEqual((target / "alias").read_bytes(), b"public")
        self.assertEqual((target / "alias").stat().st_ino,(target / "real").stat().st_ino)

    def test_escaping_hardlink_refused(self):
        archive = self.archive([("alias","hardlink","../outside")])
        with self.assertRaises(ValueError):
            unpack_rootfs(archive,self.base / "root")

    def test_special_device_refused(self):
        archive = self.archive([("device","device","")])
        with self.assertRaises(ValueError):
            unpack_rootfs(archive,self.base / "root")

    def test_link_parent_collision_cannot_redirect_file_write(self):
        archive = self.archive([("redirect","symlink","/outside"),("redirect/data","file",b"public")])
        with self.assertRaises(FileExistsError):
            unpack_rootfs(archive,self.base / "root")
        self.assertFalse((self.base / "outside").exists())
