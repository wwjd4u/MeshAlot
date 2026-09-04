package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIdentityPersistsAndPermissionsAreRestrictive(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "private")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "identity.json")
	first, created, err := LoadOrCreateIdentity(path)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected identity to be created")
	}
	second, created, err := LoadOrCreateIdentity(path)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("expected identity to be reused")
	}
	if first.NodeID != second.NodeID || first.PublicKey != second.PublicKey || first.PrivateKey != second.PrivateKey {
		t.Fatal("identity changed across reload")
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("identity permissions too broad: %04o", info.Mode().Perm())
		}
	}
}

func TestIdentityRejectsBroadPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission test")
	}
	dir := filepath.Join(t.TempDir(), "private")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "identity.json")
	if _, _, err := LoadOrCreateIdentity(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadIdentity(path); err == nil {
		t.Fatal("expected broad identity permissions to be rejected")
	}
}
