package conversation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type MarkerVariant uint8

const (
	Inline MarkerVariant = iota
	Bordered
	Separator
)

type MarkerProps struct {
	Variant MarkerVariant
	Icon    *ui.SVG
	Busy    bool
}

func Marker(c *ui.Context, text string, p MarkerProps) ui.Element {
	t := theme.TokensOf(c)
	e := ui.Row(c).Gap(8).AlignItems(ui.Center).TextColor(t.MutedForeground).FontSize(12).Role(ui.RoleStatus).Label(text).Busy(p.Busy)
	if p.Variant == Bordered {
		e.Padding(10, 12).Border(1, t.Border).Radius(t.RadiusMD())
	}
	return e.Children(func() {
		if p.Variant == Separator {
			ui.Box(c).Grow(1).Height(1).Background(t.Border)
		}
		if p.Icon != nil {
			ui.Icon(c, p.Icon).FontSize(14)
		}
		ui.Text(c, text).SingleLine()
		if p.Variant == Separator {
			ui.Box(c).Grow(1).Height(1).Background(t.Border)
		}
	})
}
