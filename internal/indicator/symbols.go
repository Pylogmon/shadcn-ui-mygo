package indicator

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

//go:embed check.svg
var checkSource []byte

//go:embed x.svg
var closeSource []byte

var checkIcon = ui.MustParseSVG(checkSource)
var closeIcon = ui.MustParseSVG(closeSource)

// Check is the fixed-size decorative selection indicator.
func Check(c *ui.Context) ui.Element { return symbol(c, checkIcon) }

// Close is the fixed-size decorative clear/remove indicator.
func Close(c *ui.Context) ui.Element { return symbol(c, closeIcon) }

func symbol(c *ui.Context, icon *ui.SVG) ui.Element {
	return ui.Icon(c, icon).Size(16, 16).FontSize(16).Shrink(0).AlignSelf(ui.Center).
		IgnoreInput(true).AccessHidden(true)
}
