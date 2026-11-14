// Package all explicitly includes the complete archived Material Symbols
// catalog for dynamic name lookup. Applications using a fixed set of icons
// should instead import material/outlined, material/rounded or material/sharp.
package all

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/internal/materialarchive"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/internal/materialsvg"
)

//go:embed catalog.zip
var catalog []byte

var ErrNotFound = errors.New("material: unknown symbol name")
var ErrVariant = errors.New("material: invalid symbol variant")

var resolved sync.Map

var index = sync.OnceValues(func() (*materialarchive.Catalog, []string) {
	r, err := materialarchive.Open(catalog)
	if err != nil {
		panic(fmt.Errorf("material: embedded catalog: %w", err))
	}
	names := []string{}
	for _, key := range r.Keys() {
		if strings.HasPrefix(key, "outlined/") && strings.HasSuffix(key, ".svg") && !strings.HasSuffix(key, ".fill.svg") {
			names = append(names, strings.TrimSuffix(strings.TrimPrefix(key, "outlined/"), ".svg"))
		}
	}
	sort.Strings(names)
	return r, names
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
	b, err := files.Read(key)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	if err != nil {
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
