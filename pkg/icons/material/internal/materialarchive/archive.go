// Package materialarchive stores the complete, byte-exact SVG catalog in
// independently compressed blocks. Only the index is inflated at open time.
package materialarchive

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"time"
)

const (
	entriesPerBlock = 128
	maxBlockBytes   = 1 << 20
	maxIndexBytes   = 8 << 20
)

// A location contains a block number, byte offset and SVG byte length.
type location [3]uint32

type manifest struct {
	Version int                 `json:"version"`
	Entries map[string]location `json:"entries"`
}

type Catalog struct {
	entries map[string]location
	blocks  map[string]*zip.File
}

func blockName(number uint32) string { return fmt.Sprintf("blocks/%04d", number) }

func read(file *zip.File) ([]byte, error) {
	r, err := file.Open()
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(r)
	return data, errors.Join(err, r.Close())
}

// Open reads the small index, without decompressing any SVG block. The archive
// is immutable; independent reads are safe from concurrent callers.
func Open(data []byte) (*Catalog, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	files := make(map[string]*zip.File, len(r.File))
	for _, file := range r.File {
		if files[file.Name] != nil {
			return nil, fmt.Errorf("duplicate catalog member %q", file.Name)
		}
		files[file.Name] = file
	}
	index := files["index.json"]
	if index == nil || index.UncompressedSize64 > maxIndexBytes {
		return nil, errors.New("invalid catalog index")
	}
	encoded, err := read(index)
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(encoded, &m); err != nil {
		return nil, err
	}
	if m.Version != 1 || len(m.Entries) == 0 {
		return nil, errors.New("unsupported catalog index")
	}
	for name, pos := range m.Entries {
		block := files[blockName(pos[0])]
		if name == "" || block == nil || block.UncompressedSize64 > maxBlockBytes || uint64(pos[1])+uint64(pos[2]) > block.UncompressedSize64 || pos[2] == 0 {
			return nil, fmt.Errorf("invalid catalog location for %q", name)
		}
	}
	delete(files, "index.json")
	return &Catalog{entries: m.Entries, blocks: files}, nil
}

func (c *Catalog) Keys() []string {
	keys := make([]string, 0, len(c.entries))
	for key := range c.entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Read inflates one bounded block and returns the requested original SVG.
// Unrequested SVGs are neither parsed nor retained in an uncompressed cache.
func (c *Catalog) Read(key string) ([]byte, error) {
	pos, ok := c.entries[key]
	if !ok {
		return nil, fs.ErrNotExist
	}
	data, err := read(c.blocks[blockName(pos[0])])
	if err != nil {
		return nil, err
	}
	return data[int(pos[1]):int(uint64(pos[1])+uint64(pos[2]))], nil
}

// Encode produces a deterministic ZIP. Adjacent SVGs share a compression
// dictionary within each block, while the index permits lazy name lookup.
func Encode(svgs map[string][]byte) ([]byte, error) {
	if len(svgs) == 0 {
		return nil, errors.New("empty SVG catalog")
	}
	keys := make([]string, 0, len(svgs))
	for key := range svgs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	w.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(w, flate.BestCompression)
	})
	write := func(name string, data []byte) error {
		h := zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		h.SetMode(0644)
		f, err := w.CreateHeader(&h)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	}
	m := manifest{Version: 1, Entries: make(map[string]location, len(svgs))}
	var block uint32
	for start := 0; start < len(keys); block++ {
		var raw bytes.Buffer
		end := start
		for end < len(keys) && end-start < entriesPerBlock {
			key, svg := keys[end], svgs[keys[end]]
			if key == "" || len(svg) == 0 || len(svg) > maxBlockBytes {
				return nil, fmt.Errorf("invalid SVG %q", key)
			}
			if raw.Len()+len(svg) > maxBlockBytes {
				break
			}
			m.Entries[key] = location{block, uint32(raw.Len()), uint32(len(svg))}
			raw.Write(svg)
			end++
		}
		if err := write(blockName(block), raw.Bytes()); err != nil {
			return nil, err
		}
		start = end
	}
	index, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if len(index) > maxIndexBytes {
		return nil, errors.New("catalog index exceeds limit")
	}
	if err := write("index.json", index); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
