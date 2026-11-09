package display

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/shimmer"
)

// ShimmerTextProps configures the pinned shimmer utility on actual native text.
type ShimmerTextProps struct {
	// Static disables the sweep. Desktop reduced motion also renders plain text.
	Static bool
	// Base is the source currentColor, including alpha. Nil uses the native
	// theme's Text color; set Base for a different ancestor color. While active,
	// configure the gradient here instead of applying TextColor afterward.
	Base *ui.Color
	Font ui.Font
	// MeasureFont supplies known inherited metrics for the default 3ch + 40 DIP
	// band without changing glyph styles. Spread can set the band explicitly.
	MeasureFont                  *ui.Font
	Highlight                    *ui.Color
	Duration                     time.Duration
	Angle, Spread, SpreadPercent *float32
	Once, Reverse                bool
}

// ShimmerText preserves MyGo's text shaping, truncation, layout and semantics.
// The default sweep is linear, lasts two seconds and follows the text direction.
func ShimmerText(c *ui.Context, text string, options ...ShimmerTextProps) ui.Element {
	p := ShimmerTextProps{}
	if len(options) > 0 {
		p = options[0]
	}
	base := c.Theme().Text
	if p.Base != nil {
		base = *p.Base
	}
	return shimmer.Text(c, text, shimmer.Props{
		Active: !p.Static, Base: base, Font: p.Font, MeasureFont: p.MeasureFont,
		Highlight: p.Highlight, Duration: p.Duration, Angle: p.Angle,
		Spread: p.Spread, SpreadPercent: p.SpreadPercent, Once: p.Once, Reverse: p.Reverse,
	})
}
