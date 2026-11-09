package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// cleanupCache removes only the isolated directory created by verify.
// Go's downloaded modules contain read-only directories and files. Restore
// owner permissions before descending into each directory; never follow links
// into a different cache or the source checkout.
func cleanupCache(root string) error {
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mode := info.Mode().Perm() | 0200
		if entry.IsDir() {
			mode |= 0700
		}
		return os.Chmod(path, mode)
	}); err != nil {
		return err
	}
	return os.RemoveAll(root)
}
