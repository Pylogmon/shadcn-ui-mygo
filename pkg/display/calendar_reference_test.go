package display

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCalendarFixedReferenceArchive(t *testing.T) {
	raw, err := os.ReadFile("testdata/shadcn-calendar/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Commit string
		Files  []struct{ Archive, SHA256 string }
	}
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m.Commit != "8ab27715521cbd4f87ae6886cd989cbb678f6de7" || len(m.Files) != 28 {
		t.Fatal("wrong fixed calendar references")
	}
	for _, file := range m.Files {
		raw, err := os.ReadFile(filepath.Join("testdata/shadcn-calendar", file.Archive))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(raw)
		if hex.EncodeToString(hash[:]) != file.SHA256 {
			t.Errorf("changed fixed source %s", file.Archive)
		}
	}
}
