// Package materialsvg shares lazy immutable SVG parsing between generated
// on-demand symbols and the explicitly imported full catalog.
package materialsvg

import (
	"fmt"
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
)

var parsed sync.Map

// Parse caches only requested SVGs. Generated icon packages have no global
// initializers or catalog tables that retain unrequested icon functions.
func Parse(source string) *ui.SVG {
	if svg, ok := parsed.Load(source); ok {
		return svg.(*ui.SVG)
	}
	svg := ui.MustParseSVG([]byte(source))
	actual, _ := parsed.LoadOrStore(source, svg)
	return actual.(*ui.SVG)
}

// Select is the implementation called by generated individual icon functions.
func Select(fill []material.Fill, regular, filled string) *ui.SVG {
	if len(fill) > 1 {
		panic("material: expected at most one fill state")
	}
	if len(fill) == 0 || fill[0] == material.Unfilled {
		return Parse(regular)
	}
	if fill[0] == material.Filled {
		return Parse(filled)
	}
	panic(fmt.Sprintf("material: invalid fill state %d", fill[0]))
}
