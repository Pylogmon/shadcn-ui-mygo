// Package tooltip shares the in-window native tooltip between compound
// controls and overlay without coupling those public component packages.
package tooltip

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// InWindow keeps native tooltip timing, focus and accessibility behavior.
// Position may customize attachment; nil uses the standard bottom-start hint.
func InWindow(c *ui.Context, anchor ui.Element, text string, hoverOnly bool, position func(ui.Element)) ui.Element {
	if text == "" {
		return ui.Element{}
	}
	anchor.Description(text)
	hovered := anchor.Hovered()
	t := theme.TokensOf(c)
	var tip ui.Element
	ui.TooltipBase(c, anchor, func(panel ui.Element) {
		if hoverOnly && !hovered {
			panel.Size(0, 0).AccessHidden(true)
			return
		}
		tip = panel
		if position != nil {
			position(panel)
		} else {
			panel.AttachTo(anchor, ui.AnchorBottomStart, ui.AnchorTopStart).Margin(6, 0, 0, 0).AttachedOrigin()
		}
		panel.Padding(6, 12).Radius(t.Radius).MaxWidth(320).Background(t.Primary).
			TextColor(t.PrimaryForeground).FontSize(12).Transition(motion.Popup(c)).Children(func() { ui.Text(c, text) })
	})
	return tip
}
