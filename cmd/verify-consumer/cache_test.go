package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupCacheRemovesReadonlyModuleDirectories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "isolated-cache")
	nested := filepath.Join(root, "modules", "example.org", "module@v1.0.0", "nested")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "module.go"), []byte("package example\n"), 0444); err != nil {
		t.Fatal(err)
	}
	for dir := nested; dir != root; dir = filepath.Dir(dir) {
		if err := os.Chmod(dir, 0555); err != nil {
			t.Fatal(err)
		}
	}
	if err := cleanupCache(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("isolated module cache remains: %v", err)
	}
	if err := cleanupCache(root); err != nil {
		t.Fatalf("cleanup was not idempotent: %v", err)
	}
}

func TestCleanupCacheDoesNotFollowDirectorySymlinks(t *testing.T) {
	root := filepath.Join(t.TempDir(), "isolated-cache")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	file := filepath.Join(outside, "retained.txt")
	if err := os.WriteFile(file, []byte("retained"), 0444); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}
	if err := cleanupCache(root); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil || string(data) != "retained" || before.Mode() != after.Mode() {
		t.Fatalf("cleanup changed a symlink target: %q mode%v ->%v err%v", data, before.Mode(), after.Mode(), err)
	}
	if _, err := os.Lstat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("isolated cache remains: %v", err)
	}
}
