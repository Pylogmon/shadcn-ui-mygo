// Package all explicitly includes the complete archived Material Symbols
// catalog for dynamic name lookup. Applications using a fixed set of icons
// should instead import material/outlined, material/rounded or material/sharp.
package all

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/internal/materialsvg"
)

//go:embed catalog.zip
var catalog []byte

var ErrNotFound = errors.New("material: unknown symbol name")
var ErrVariant = errors.New("material: invalid symbol variant")

var resolved sync.Map

var index = sync.OnceValues(func() (map[string]*zip.File, []string) {
	r, err := zip.NewReader(bytes.NewReader(catalog), int64(len(catalog)))
	if err != nil {
		panic(fmt.Errorf("material: embedded catalog: %w", err))
	}
	files := make(map[string]*zip.File, len(r.File))
	names := []string{}
	for _, file := range r.File {
		files[file.Name] = file
		if strings.HasPrefix(file.Name, "outlined/") && strings.HasSuffix(file.Name, ".svg") && !strings.HasSuffix(file.Name, ".fill.svg") {
			names = append(names, strings.TrimSuffix(strings.TrimPrefix(file.Name, "outlined/"), ".svg"))
		}
	}
	sort.Strings(names)
	return files, names
})

// Names returns an independent alphabetically sorted list of symbol names.
func Names() []string {
	_, names := index()
	return slices.Clone(names)
}

// Lookup returns a lazily parsed SVG. The default variant is Outlined,
// Unfilled; variants never fall back silently to a different family or fill.
// Name matching is exact, using Google's lower-case snake_case symbol names.
func Lookup(name string, variants ...material.Variant) (*ui.SVG, error) {
	v := material.Variant{}
	if len(variants) > 1 {
		return nil, ErrVariant
	}
	if len(variants) == 1 {
		v = variants[0]
	}
	if v.Style > material.Sharp || v.Fill > material.Filled {
		return nil, ErrVariant
	}
	// This is an embedded catalog lookup, not a path or network lookup.
	if name == "" || strings.ContainsAny(name, "/\\.") {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	style := [...]string{"outlined", "rounded", "sharp"}[v.Style]
	suffix := ".svg"
	if v.Fill == material.Filled {
		suffix = ".fill.svg"
	}
	key := style + "/" + name + suffix
	if svg, ok := resolved.Load(key); ok {
		return svg.(*ui.SVG), nil
	}
	files, _ := index()
	file := files[key]
	if file == nil {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	r, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("material: read %q: %w", name, err)
	}
	b, readErr := io.ReadAll(r)
	closeErr := r.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, fmt.Errorf("material: read %q: %w", name, err)
	}
	svg := materialsvg.Parse(string(b))
	actual, _ := resolved.LoadOrStore(key, svg)
	return actual.(*ui.SVG), nil
}

// Must is convenient for constant, known symbol names; it panics on an unknown
// name or invalid variant. Use Lookup for application/user-supplied names.
func Must(name string, variants ...material.Variant) *ui.SVG {
	svg, err := Lookup(name, variants...)
	if err != nil {
		panic(err)
	}
	return svg
}
