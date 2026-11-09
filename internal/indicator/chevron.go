package indicator

import "github.com/egoist/mygo/ui"

var chevronDown = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>`))

// ChevronDown is the shared decorative select indicator. Both its layout box
// and SVG glyph use 16 DIPs, independently of an ancestor's font or line height.
// The vector inherits the caller's current color.
func ChevronDown(c *ui.Context) ui.Element {
	return ui.Icon(c, chevronDown).Size(16, 16).FontSize(16).Shrink(0).AlignSelf(ui.Center).IgnoreInput(true).AccessHidden(true)
}
