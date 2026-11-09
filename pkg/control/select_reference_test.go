package control

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSelectFixedReferenceArchive(t *testing.T) {
	data, err := os.ReadFile("testdata/shadcn-select/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit string
		Files  []struct{ Archive, SHA256 string }
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "8ab27715521cbd4f87ae6886cd989cbb678f6de7" || len(manifest.Files) != 35 {
		t.Fatalf("unexpected fixed corpus %s/%d", manifest.Commit, len(manifest.Files))
	}
	for _, file := range manifest.Files {
		raw, err := os.ReadFile(filepath.Join("testdata/shadcn-select", file.Archive))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(raw)
		if hex.EncodeToString(hash[:]) != file.SHA256 {
			t.Errorf("changed fixed source %s", file.Archive)
		}
	}
}
