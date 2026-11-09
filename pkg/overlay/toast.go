package overlay

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Toast uses MyGo's window-owned queue: time pauses on hover/focus and screen
// readers receive announcements. Use c.CloseToast(id) to dismiss explicitly.
func Toast(c *ui.Context, toast ui.Toast) string { return c.AddToast(toast) }

// Toaster builds the styled toast viewport once per view. Toasts have close and
// optional action buttons, preserve focus and animate when added or removed.
func Toaster(c *ui.Context) ui.Element {
	t := theme.TokensOf(c)
	return ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Padding(24).Gap(12).AlignItems(ui.End)
		for _, toast := range toasts {
			parts := ui.ToastBase(c, toast)
			root := parts.Root.Row().Width(356).MaxWidthPercent(100).Gap(12).Padding(16).Radius(t.Radius).
				Border(1, t.Border).Background(t.Background).TextColor(t.Foreground).
				Shadow(0, 4, 16, 0, ui.RGBA(0, 0, 0, 0.1)).Transition(motion.Popup(c))
			if toast.Type == "error" {
				root.Border(1, t.Destructive).TextColor(t.Destructive)
			}
			root.Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
					ui.Text(c, toast.Title).FontSize(14).FontWeight(600)
					if toast.Description != "" {
						ui.Text(c, toast.Description).FontSize(13).TextColor(t.MutedForeground)
					}
				})
				if toast.Action != "" {
					button(c, parts.ActionButton(), toast.Action, false).Height(32)
				}
				close := parts.CloseButton().Label("Dismiss notification").Size(20, 20).Radius(4).Center()
				if close.Hovered() {
					close.Background(t.Accent)
				}
				close.Children(func() { ui.Text(c, "×").FontSize(18) })
			})
		}
	})
}
