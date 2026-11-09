package control_test

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypedInputFixedOriginals(t *testing.T) {
	const root = "testdata/shadcn-input"
	raw, err := os.ReadFile(root + "/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Commit string `json:"commit"`
		Files  []struct {
			Path    string `json:"path"`
			Archive string `json:"archive"`
			SHA256  string `json:"sha256"`
			GitBlob string `json:"git_blob"`
			Bytes   int    `json:"bytes"`
		} `json:"files"`
	}
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Commit != "8ab27715521cbd4f87ae6886cd989cbb678f6de7" || len(doc.Files) != 18 {
		t.Fatal("fixed input scope changed")
	}
	for _, f := range doc.Files {
		b, err := os.ReadFile(filepath.Join(root, f.Archive))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		blob := sha1.New()
		fmt.Fprintf(blob, "blob %d%c", len(b), 0)
		blob.Write(b)
		if len(b) != f.Bytes || hex.EncodeToString(sum[:]) != f.SHA256 || hex.EncodeToString(blob.Sum(nil)) != f.GitBlob {
			t.Fatal("fixed original changed", f.Path)
		}
	}
	types, err := os.ReadFile(root + "/source/apps/v4/registry/bases/base/examples/input-example.tsx")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"password", "tel", "url", "search", "number", "date", "time", "file"} {
		if !strings.Contains(string(types), `type="`+kind+`"`) {
			t.Fatal("type absent from actual fixed collection", kind)
		}
	}
}
