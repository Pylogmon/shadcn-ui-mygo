package display

import (
	"github.com/egoist/mygo/ui"
)

func Direction(c *ui.Context, dir ui.Direction, content func()) ui.Element {
	return ui.Column(c).Direction(dir).Children(content)
}
