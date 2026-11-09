// Package material provides native rendering for Google's Material Symbols.
//
// Import outlined, rounded or sharp and call individual icon functions for
// linker-selected resources, or explicitly import material/all for name lookup.
// This package itself contains no icon catalog, font or network loader.
package material

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// Fill selects one of the two archived static fill states.
type Fill uint8

const (
	Unfilled Fill = iota
	Filled
)

// Style selects the Material Symbols family.
type Style uint8

const (
	Outlined Style = iota
	Rounded
	Sharp
)

// Variant identifies a static 24px, weight400, grade0 symbol. Size changes the
// rendered dimensions; it does not synthesize variable-font optical-size axes.
type Variant struct {
	Style Style
	Fill  Fill
}

// Props configures a rendered symbol. Zero Size uses24DIP. Without a Label,
// icons are decorative. Colors inherit the current text color and can also be
// set on the returned element with TextColor.
type Props struct {
	Size  float32
	Label string
}

// Icon draws an SVG returned by an on-demand icon function or all.Lookup.
// The fixed layout and glyph dimensions do not inherit a surrounding font's
// size or line height. Use the SVG directly in ButtonProps.Icon when the
// receiving component should control its own icon dimensions.
func Icon(c *ui.Context, svg *ui.SVG, props ...Props) ui.Element {
	p := Props{}
	if len(props) > 0 {
		p = props[0]
	}
	size := p.Size
	if size <= 0 || math.IsNaN(float64(size)) || math.IsInf(float64(size), 0) {
		size = 24
	}
	e := ui.Icon(c, svg).Size(size, size).FontSize(size).Shrink(0).AlignSelf(ui.Center).IgnoreInput(true)
	if p.Label != "" {
		e.Label(p.Label).Role(ui.RoleImage).AccessHidden(false)
	} else {
		e.AccessHidden(true)
	}
	return e
}
