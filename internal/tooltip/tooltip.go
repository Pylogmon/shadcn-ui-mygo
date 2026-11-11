// Package tooltip shares tooltip routing between compound controls, desktop
// hosts and overlay without coupling those public component packages.
package tooltip

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type rendererKey struct{}
type renderer struct {
	show func(*ui.Context, ui.Element, string) error
}

// Bind associates compound-control tooltips with this window's native host.
// The callback may retain the host, but never a Context or Element. Nil clears
// the binding; independent child windows start with their own empty binding.
func Bind(c *ui.Context, show func(*ui.Context, ui.Element, string) error) {
	if c == nil || !c.Valid() {
		return
	}
	ui.Local(c.Root(), rendererKey{}, func() renderer { return renderer{} }).show = show
}

// Show uses the bound native tooltip, or the standard in-window tooltip when
// there is no desktop host. A native failure keeps the accessible description
// and does not create a second tooltip with a different lifecycle.
func Show(c *ui.Context, anchor ui.Element, text string) {
	if text == "" {
		return
	}
	if show := ui.Local(c.Root(), rendererKey{}, func() renderer { return renderer{} }).show; show != nil {
		if err := show(c, anchor, text); err != nil {
			anchor.Description(err.Error())
		}
		return
	}
	InWindow(c, anchor, text, false, nil)
}

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
