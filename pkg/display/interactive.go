package display

import (
	"github.com/egoist/mygo/ui"
)

func ScrollArea(c *ui.Context, content func()) ui.Element {
	return ui.Scroll(c).Fill().Children(content)
}

type ResizableProps struct {
	Vertical bool
	Min, Max float32
}

func Resizable(c *ui.Context, size *float32, p ResizableProps, first, second func()) ui.Element {
	if p.Min > 0 {
		*size = max(p.Min, *size)
	}
	if p.Max > 0 {
		*size = min(p.Max, *size)
	}
	if p.Vertical {
		return ui.SplitVertical(c, size, first, second)
	}
	return ui.Split(c, size, first, second)
}
func Direction(c *ui.Context, dir ui.Direction, content func()) ui.Element {
	return ui.Column(c).Direction(dir).Children(content)
}
