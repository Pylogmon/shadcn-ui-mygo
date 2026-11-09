package display

import (
	"github.com/egoist/mygo/ui"
)

func ScrollArea(c *ui.Context, content func()) ui.Element {
	return ui.Scroll(c).Fill().Children(content)
}

func Direction(c *ui.Context, dir ui.Direction, content func()) ui.Element {
	return ui.Column(c).Direction(dir).Children(content)
}
