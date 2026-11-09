package all_test

import (
	"bytes"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/all"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/rounded"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/sharp"
)

func TestEveryArchivedSymbolParsesThroughPublicNativeAPI(t *testing.T) {
	names := all.Names()
	if len(names) != 4157 || !slices.IsSorted(names) {
		t.Fatalf("pinned catalog names=%d or unsorted", len(names))
	}
	for _, name := range names {
		for _, style := range []material.Style{material.Outlined, material.Rounded, material.Sharp} {
			for _, fill := range []material.Fill{material.Unfilled, material.Filled} {
				svg, err := all.Lookup(name, material.Variant{Style: style, Fill: fill})
				if err != nil {
					t.Fatalf("%s/%d/%d: %v", name, style, fill, err)
				}
				if w, h := svg.Size(); w != 24 || h != 24 {
					t.Fatalf("%s/%d/%d dimensions=%v,%v", name, style, fill, w, h)
				}
			}
		}
	}
	copy := all.Names()
	copy[0] = "mutated"
	if all.Names()[0] == "mutated" {
		t.Fatal("caller mutated shared catalog names")
	}
}

func TestDynamicAndOnDemandSymbolsHaveIdenticalNativePixels(t *testing.T) {
	for _, tc := range []struct {
		style material.Style
		icon  func(...material.Fill) *ui.SVG
	}{{material.Outlined, outlined.Home}, {material.Rounded, rounded.Home}, {material.Sharp, sharp.Home}} {
		for _, fill := range []material.Fill{material.Unfilled, material.Filled} {
			dynamic := all.Must("home", material.Variant{Style: tc.style, Fill: fill})
			selected := tc.icon(fill)
			if dynamic != selected || !bytes.Equal(dynamic.Image(32, 32, nil).Pix, selected.Image(32, 32, nil).Pix) {
				t.Fatalf("full and on-demand variants differ: %d/%d", tc.style, fill)
			}
		}
	}
}

func TestDynamicLookupReportsUnknownNamesAndInvalidVariants(t *testing.T) {
	for _, name := range []string{"", "Home", "missing_symbol", "../home", "outlined/home", "home.svg"} {
		if svg, err := all.Lookup(name); svg != nil || !errors.Is(err, all.ErrNotFound) {
			t.Fatalf("%q = %v,%v", name, svg, err)
		}
	}
	for _, variant := range []material.Variant{{Style: 42}, {Fill: 42}} {
		if svg, err := all.Lookup("home", variant); svg != nil || !errors.Is(err, all.ErrVariant) {
			t.Fatalf("invalid variant=%v,%v", svg, err)
		}
	}
	if _, err := all.Lookup("home", material.Variant{}, material.Variant{}); !errors.Is(err, all.ErrVariant) {
		t.Fatal("multiple variant arguments were silently ignored")
	}
}

func TestConcurrentLookupSharesOnlyRequestedImmutableSVG(t *testing.T) {
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if all.Must("settings") != outlined.Settings() {
				t.Error("concurrent lookups returned separate parsed symbols")
			}
		})
	}
	wg.Wait()
}
