// Command generate imports the fixed Material Symbols SVG subset from a local
// official Git checkout. It never downloads missing objects.
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"go/format"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	repository    = "https://github.com/google/material-design-icons"
	commit        = "10ecc14790612ca8fd164e3f4ffe9c16fc50d5b7"
	formatVersion = 1
	expectedNames = 4157
	module        = "github.com/pylogmon/shadcn-ui-mygo"
)

var styles = []string{"outlined", "rounded", "sharp"}

type symbol struct {
	name, style string
	fill        int
}

type sourceEntry struct {
	Path    string `json:"path"`
	GitBlob string `json:"gitBlob"`
	SHA256  string `json:"sha256"`
	Bytes   int    `json:"bytes"`
}

type sourceIndex struct {
	FormatVersion int           `json:"formatVersion"`
	Repository    string        `json:"repository"`
	Commit        string        `json:"commit"`
	Files         []sourceEntry `json:"files"`
}

type nameMapping struct {
	Name   string `json:"name"`
	GoName string `json:"goName"`
}

type generatedFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

type manifest struct {
	FormatVersion  int             `json:"formatVersion"`
	Repository     string          `json:"repository"`
	Commit         string          `json:"commit"`
	License        string          `json:"license"`
	Scope          string          `json:"scope"`
	Weight         int             `json:"weight"`
	Grade          int             `json:"grade"`
	OpticalSize    int             `json:"opticalSize"`
	Fills          []int           `json:"fills"`
	Styles         []string        `json:"styles"`
	Names          int             `json:"names"`
	SVGVariants    int             `json:"svgVariants"`
	UniqueGitBlobs int             `json:"uniqueGitBlobs"`
	SourceIndex    string          `json:"sourceIndex"`
	Verification   string          `json:"verification"`
	NamesMapping   []nameMapping   `json:"namesMapping"`
	GeneratedFiles []generatedFile `json:"generatedFiles"`
}

func main() {
	repo := flag.String("repo", "", "local official Git checkout containing all fixed-commit objects; required")
	out := flag.String("out", "pkg/icons/material", "output Material Symbols package directory")
	flag.Parse()
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./pkg/icons/material/internal/generate -repo /path/to/material-design-icons [-out pkg/icons/material]")
		os.Exit(2)
	}
	if err := generate(*repo, *out); err != nil {
		fmt.Fprintln(os.Stderr, "material symbols:", err)
		os.Exit(1)
	}
}

func offlineCommand(repo string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	// Partial clones must fail when an object is absent instead of contacting
	// the promisor remote. Generation therefore remains an offline operation.
	cmd.Env = append(os.Environ(), "GIT_NO_LAZY_FETCH=1", "GIT_TERMINAL_PROMPT=0")
	return cmd
}

func selected(path string) (symbol, bool) {
	parts := strings.Split(path, "/")
	if len(parts) != 5 || parts[0] != "symbols" || parts[1] != "web" {
		return symbol{}, false
	}
	style := strings.TrimPrefix(parts[3], "materialsymbols")
	if style != "outlined" && style != "rounded" && style != "sharp" {
		return symbol{}, false
	}
	s := symbol{name: parts[2], style: style}
	switch parts[4] {
	case parts[2] + "_24px.svg":
	case parts[2] + "_fill1_24px.svg":
		s.fill = 1
	default:
		return symbol{}, false
	}
	return s, true
}

func treeIndex(repo string) (map[string]string, error) {
	cmd := offlineCommand(repo, "ls-tree", "-r", "--full-tree", commit, "--", "LICENSE", "README.md", "symbols/web")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()
	index := make(map[string]string)
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid Git tree row")
		}
		_, isSymbol := selected(parts[1])
		if !isSymbol && parts[1] != "LICENSE" && parts[1] != "README.md" {
			continue
		}
		header := strings.Fields(parts[0])
		if len(header) != 3 || header[1] != "blob" {
			return nil, fmt.Errorf("not a blob: %s", parts[1])
		}
		index[parts[1]] = header[2]
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf("read fixed Git tree: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if len(index) != expectedNames*6+2 || index["LICENSE"] == "" || index["README.md"] == "" {
		return nil, fmt.Errorf("fixed source inventory: got %d paths, want %d", len(index), expectedNames*6+2)
	}
	return index, nil
}

func readBlobs(repo string, index map[string]string) (map[string][]byte, error) {
	unique := make(map[string]bool)
	for _, id := range index {
		unique[id] = true
	}
	ids := make([]string, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	cmd := offlineCommand(repo, "cat-file", "--batch")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()
	reader := bufio.NewReader(pipe)
	blobs := make(map[string][]byte, len(ids))
	for _, id := range ids {
		if _, err = fmt.Fprintln(in, id); err != nil {
			return nil, err
		}
		header, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read blob %s: %w", id, err)
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[0] != id || fields[1] != "blob" {
			return nil, fmt.Errorf("blob unavailable offline: %s", strings.TrimSpace(header))
		}
		n, err := strconv.Atoi(fields[2])
		if err != nil || n < 0 {
			return nil, fmt.Errorf("invalid blob length for %s", id)
		}
		b := make([]byte, n)
		if _, err = io.ReadFull(reader, b); err != nil {
			return nil, err
		}
		if next, err := reader.ReadByte(); err != nil || next != '\n' {
			return nil, fmt.Errorf("invalid blob terminator for %s", id)
		}
		h := sha1.New()
		fmt.Fprintf(h, "blob %d\x00", len(b))
		h.Write(b)
		if hex.EncodeToString(h.Sum(nil)) != id {
			return nil, fmt.Errorf("Git blob hash mismatch for %s", id)
		}
		blobs[id] = b
	}
	if err = in.Close(); err != nil {
		return nil, err
	}
	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf("read Git blobs: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return blobs, nil
}

func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func validateSVG(path string, b []byte) error {
	d := xml.NewDecoder(bytes.NewReader(b))
	root := false
	for {
		t, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if start, ok := t.(xml.StartElement); ok && !root {
			root = true
			if start.Name.Local != "svg" || start.Name.Space != "http://www.w3.org/2000/svg" {
				return fmt.Errorf("%s: expected SVG root", path)
			}
			attributes := make(map[string]string)
			for _, a := range start.Attr {
				attributes[a.Name.Local] = a.Value
			}
			width, widthErr := strconv.ParseFloat(strings.TrimSuffix(attributes["width"], "px"), 64)
			height, heightErr := strconv.ParseFloat(strings.TrimSuffix(attributes["height"], "px"), 64)
			viewBoxFields := strings.Fields(attributes["viewBox"])
			if widthErr != nil || heightErr != nil || width != 24 || height != 24 || len(viewBoxFields) != 0 && len(viewBoxFields) != 4 {
				return fmt.Errorf("%s: expected fixed 24px SVG dimensions and valid optional viewBox", path)
			}
		}
	}
	if !root {
		return fmt.Errorf("%s: missing SVG root", path)
	}
	return nil
}

func exportedName(name string) string {
	var b strings.Builder
	for _, part := range strings.Split(name, "_") {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}
	n := b.String()
	if len(n) > 0 && n[0] >= '0' && n[0] <= '9' {
		n = "Icon" + n
	}
	return n
}

func goNames(names []string) ([]nameMapping, error) {
	reserved, used := make(map[string]bool), make(map[string]bool)
	for _, name := range names {
		reserved[exportedName(name)] = true
	}
	mapping := make([]nameMapping, 0, len(names))
	for _, name := range names {
		base := exportedName(name)
		if !token.IsIdentifier(base) {
			return nil, fmt.Errorf("invalid Go icon name %q from %q", base, name)
		}
		chosen := base
		if used[chosen] {
			for suffix := 2; ; suffix++ {
				chosen = base + strconv.Itoa(suffix)
				if !used[chosen] && !reserved[chosen] {
					break
				}
			}
		}
		used[chosen] = true
		mapping = append(mapping, nameMapping{name, chosen})
	}
	return mapping, nil
}

func generatedGo(style string, names []nameMapping, svgs map[string][]byte) ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by material/internal/generate; DO NOT EDIT.\n// Material Symbols by Google, Apache-2.0; see ../sources/LICENSE.\n// Fixed source: %s/tree/%s.\n\n", repository, commit)
	fmt.Fprintf(&b, "// Package %s provides individual Material Symbols (%s, 24px, weight 400, grade 0).\n// Referenced functions keep only their own FILL 0/1 SVG pair reachable to the Go linker.\npackage %s\n\n", style, style, style)
	fmt.Fprintf(&b, "import (\n\"github.com/egoist/mygo/ui\"\n\"%s/pkg/icons/material\"\n\"%s/pkg/icons/material/internal/materialsvg\"\n)\n\n", module, module)
	for _, name := range names {
		regular, ok := svgs[style+"/"+name.Name+".svg"]
		if !ok {
			return nil, fmt.Errorf("missing regular SVG: %s/%s", style, name.Name)
		}
		filled, ok := svgs[style+"/"+name.Name+".fill.svg"]
		if !ok {
			return nil, fmt.Errorf("missing filled SVG: %s/%s", style, name.Name)
		}
		fmt.Fprintf(&b, "// %s returns %q; FILL defaults to material.Unfilled.\nfunc %s(fill ...material.Fill) *ui.SVG {\nreturn materialsvg.Select(fill, %s, %s)\n}\n\n", name.GoName, name.Name, name.GoName, strconv.Quote(string(regular)), strconv.Quote(string(filled)))
	}
	return format.Source(b.Bytes())
}

func catalogZIP(svgs map[string][]byte) ([]byte, error) {
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	w.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) { return flate.NewWriter(w, flate.BestCompression) })
	keys := make([]string, 0, len(svgs))
	for key := range svgs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		h := zip.FileHeader{Name: key, Method: zip.Deflate}
		h.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		h.SetMode(0644)
		f, err := w.CreateHeader(&h)
		if err != nil {
			return nil, err
		}
		if _, err = f.Write(svgs[key]); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func gzipJSON(value any) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	w.Header.ModTime = time.Time{}
	w.Header.OS = 255
	if err = json.NewEncoder(w).Encode(value); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func generate(repo, out string) error {
	index, err := treeIndex(repo)
	if err != nil {
		return err
	}
	blobs, err := readBlobs(repo, index)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(index))
	for path := range index {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	source := sourceIndex{FormatVersion: formatVersion, Repository: repository, Commit: commit}
	svgs, namesSet := make(map[string][]byte), make(map[string]bool)
	for _, path := range paths {
		b := blobs[index[path]]
		source.Files = append(source.Files, sourceEntry{path, index[path], digest(b), len(b)})
		if s, ok := selected(path); ok {
			if err = validateSVG(path, b); err != nil {
				return err
			}
			key := s.style + "/" + s.name
			if s.fill == 1 {
				key += ".fill"
			}
			key += ".svg"
			svgs[key], namesSet[s.name] = b, true
		}
	}
	if len(namesSet) != expectedNames || len(svgs) != expectedNames*6 {
		return fmt.Errorf("invalid icon inventory")
	}
	names := make([]string, 0, len(namesSet))
	for name := range namesSet {
		names = append(names, name)
	}
	sort.Strings(names)
	mapping, err := goNames(names)
	if err != nil {
		return err
	}
	outputs := make(map[string][]byte)
	for _, style := range styles {
		outputs[style+"/icons_gen.go"], err = generatedGo(style, mapping, svgs)
		if err != nil {
			return err
		}
	}
	outputs["all/catalog.zip"], err = catalogZIP(svgs)
	if err != nil {
		return err
	}
	outputs["sources/index.json.gz"], err = gzipJSON(source)
	if err != nil {
		return err
	}
	outputs["sources/LICENSE"] = blobs[index["LICENSE"]]
	outputs["sources/README.upstream.md"] = blobs[index["README.md"]]
	m := manifest{FormatVersion: formatVersion, Repository: repository, Commit: commit, License: "Apache-2.0", Scope: "Static Material Symbols SVG: weight 400, grade 0, optical size 24, FILL 0/1; variable-font axes and other static weights/sizes are not included.", Weight: 400, Grade: 0, OpticalSize: 24, Fills: []int{0, 1}, Styles: styles, Names: len(names), SVGVariants: len(svgs), UniqueGitBlobs: len(blobs), SourceIndex: "index.json.gz", Verification: "Offline fixed Git tree, every original Git blob ID and SHA-256, complete six-variant name coverage, SVG XML and fixed dimensions.", NamesMapping: mapping}
	outputPaths := make([]string, 0, len(outputs))
	for path := range outputs {
		outputPaths = append(outputPaths, path)
	}
	sort.Strings(outputPaths)
	for _, path := range outputPaths {
		m.GeneratedFiles = append(m.GeneratedFiles, generatedFile{path, digest(outputs[path]), len(outputs[path])})
	}
	outputs["sources/manifest.json"], err = json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	outputs["sources/manifest.json"] = append(outputs["sources/manifest.json"], '\n')
	outputPaths = append(outputPaths, "sources/manifest.json")
	for _, path := range outputPaths {
		target := filepath.Join(out, filepath.FromSlash(path))
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err = os.WriteFile(target, outputs[path], 0644); err != nil {
			return err
		}
	}
	fmt.Printf("Generated %d names, %d SVG variants, %d individual functions per style; catalog.zip %d bytes.\n", len(names), len(svgs), len(mapping), len(outputs["all/catalog.zip"]))
	return nil
}
