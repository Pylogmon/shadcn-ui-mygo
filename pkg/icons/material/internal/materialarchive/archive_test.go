package materialarchive

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestRoundTripIsDeterministicAndPreservesEveryByte(t *testing.T) {
	svgs := map[string][]byte{}
	for i := range 257 {
		svgs[fmt.Sprintf("rounded/icon_%03d.svg", i)] = []byte(fmt.Sprintf("<svg>  %d\n</svg> \n", i))
	}
	encoded, err := Encode(svgs)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Encode(svgs)
	if err != nil || !bytes.Equal(encoded, again) {
		t.Fatal("archive is not reproducible", err)
	}
	catalog, err := Open(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Keys()) != len(svgs) || len(catalog.blocks) != 3 {
		t.Fatal("lost symbols or incorrect block boundaries")
	}
	for key, want := range svgs {
		got, err := catalog.Read(key)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("%s: changed original bytes: %v", key, err)
		}
	}
	if _, err := catalog.Read("missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("missing SVG did not report not-exist", err)
	}
}

func TestOpenAndLookupDoNotInflateUnrequestedBlocks(t *testing.T) {
	svgs := map[string][]byte{}
	for i := range 129 {
		svgs[fmt.Sprintf("icon_%03d", i)] = []byte("<svg>unchanged</svg>")
	}
	encoded, err := Encode(svgs)
	if err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(bytes.NewReader(encoded), int64(len(encoded)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range r.File {
		if file.Name != "blocks/0001" {
			continue
		}
		offset, err := file.DataOffset()
		if err != nil {
			t.Fatal(err)
		}
		clear(encoded[offset : uint64(offset)+file.CompressedSize64])
	}
	catalog, err := Open(encoded)
	if err != nil {
		t.Fatal("opening the index inflated a corrupted unused block", err)
	}
	if _, err := catalog.Read("icon_000"); err != nil {
		t.Fatal("reading one block inflated a different block", err)
	}
	if _, err := catalog.Read("icon_128"); err == nil {
		t.Fatal("requested corrupted block did not fail checksum/decompression")
	}
}

func TestPinnedCatalogMatchesEveryOriginalSourceHash(t *testing.T) {
	encoded, err := os.ReadFile("../../all/catalog.zip")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := Open(encoded)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open("../../sources/index.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	r, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	var index struct {
		Files []struct {
			Path, SHA256 string
			Bytes        int
		}
	}
	if err := json.NewDecoder(r).Decode(&index); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, source := range index.Files {
		parts := strings.Split(source.Path, "/")
		if len(parts) != 5 || parts[0] != "symbols" || parts[1] != "web" {
			continue
		}
		style := strings.TrimPrefix(parts[3], "materialsymbols")
		name := parts[2]
		if parts[4] == name+"_fill1_24px.svg" {
			name += ".fill"
		}
		key := style + "/" + name + ".svg"
		got, err := catalog.Read(key)
		if err != nil || len(got) != source.Bytes || fmt.Sprintf("%x", sha256.Sum256(got)) != source.SHA256 {
			t.Fatalf("%s differs from the pinned upstream SVG: %v", key, err)
		}
		count++
	}
	if count != 24942 || len(catalog.Keys()) != count {
		t.Fatalf("catalog coverage=%d entries=%d", count, len(catalog.Keys()))
	}
}
