package display

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type AlertProps struct {
	Title, Description string
	Destructive        bool
	// Content retains the original callback API. New compositions can use
	// AlertWith's typed child-context slots instead.
	Content func()
	Label   string
	Icon    *ui.SVG
}

type AlertSlots struct {
	// Icon, Title and Description replace their matching props when non-nil.
	// Content follows the description, and Action remains at the logical end.
	Icon, Title, Description, Content, Action func(*ui.Context)
}

type alertSlotKey uint8

const (
	alertBodyKey alertSlotKey = iota
	alertTextKey
	alertIconKey
	alertTitleKey
	alertDescriptionKey
	alertContentKey
	alertActionKey
	alertActionContentKey
)

type alertLayout struct{ width, actionWidth, actionHeight float32 }

// Alert exposes MyGo's assertive, atomic live-region role without taking
// keyboard focus. Content controls, such as an undo action or a link, keep
// their own native keyboard and accessibility semantics.
func Alert(c *ui.Context, p AlertProps) ui.Element { return AlertWith(c, p, AlertSlots{}) }

// AlertWith composes a leading icon, wrapping title/description and an action.
// The action participates in native layout, reserving its actual width rather
// than painting over text when it is wider than the usual upstream undo button.
func AlertWith(c *ui.Context, p AlertProps, slots AlertSlots) ui.Element {
	t := theme.TokensOf(c)
	foreground := t.CardForeground
	if p.Destructive {
		foreground = t.Destructive
	}
	root := ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Start).Gap(16).Padding(12, 16).
		Radius(t.RadiusLG()).Border(1, t.Border).Background(t.Card).TextColor(foreground).
		FontSize(14).LineHeight(1.5).Role(ui.RoleAlert)
	if p.Label != "" {
		root.Label(p.Label)
	}
	s := ui.Local(root, "alert-layout", func() alertLayout { return alertLayout{} })
	hasBody := p.Icon != nil || slots.Icon != nil || p.Title != "" || p.Description != "" || p.Content != nil || slots.Title != nil || slots.Description != nil || slots.Content != nil
	minimumBody := float32(0)
	if hasBody {
		minimumBody = 96
	}
	stacked := slots.Action != nil && s.width > 0 && s.width-34 < s.actionWidth+16+minimumBody
	if stacked {
		root.Column().Gap(12)
	}
	if slots.Action != nil {
		root.Draw(func(_ *ui.Painter, bounds ui.Rect) {
			if s.width != bounds.W {
				s.width = bounds.W
				c.Invalidate()
			}
		})
	}
	root.Children(func() {
		body := ui.Row(c.Key(alertBodyKey)).MinWidth(0).AlignItems(ui.Start).Gap(10)
		if stacked {
			body.FillWidth().Shrink(0)
		} else {
			body.Basis(0).Grow(1)
		}
		body.Children(func() {
			if slots.Icon != nil || p.Icon != nil {
				ui.Box(c.Key(alertIconKey)).Size(16, 16).Shrink(0).Margin(2, 0, 0, 0).IgnoreInput(true).AccessHidden(true).Children(func() {
					if slots.Icon != nil {
						slots.Icon(c)
					} else {
						ui.Icon(c, p.Icon).Size(16, 16).AccessHidden(true)
					}
				})
			}
			ui.Column(c.Key(alertTextKey)).Basis(0).Grow(1).MinWidth(0).Gap(2).Children(func() {
				if slots.Title != nil || p.Title != "" {
					ui.Column(c.Key(alertTitleKey)).FillWidth().MinWidth(0).FontWeight(500).Children(func() {
						if slots.Title != nil {
							slots.Title(c)
						} else {
							AlertTitle(c, p.Title)
						}
					})
				}
				if slots.Description != nil || p.Description != "" {
					ui.Column(c.Key(alertDescriptionKey)).FillWidth().MinWidth(0).Gap(16).TextColor(alertDescriptionColor(c, p.Destructive)).Children(func() {
						if slots.Description != nil {
							slots.Description(c)
						} else {
							AlertDescription(c, p.Description, p.Destructive)
						}
					})
				}
				if slots.Content != nil || p.Content != nil {
					ui.Column(c.Key(alertContentKey)).FillWidth().MinWidth(0).Gap(12).Children(func() {
						if slots.Content != nil {
							slots.Content(c)
						}
						if p.Content != nil {
							p.Content()
						}
					})
				}
			})
		})
		if slots.Action != nil {
			// Keep one stable hierarchy across inline/stacked layouts so an
			// action editor retains its caret and state on a container resize.
			actions := ui.ScrollHorizontal(c.Key(alertActionKey)).MinWidth(0).Shrink(0).Height(max(24, s.actionHeight)).Label("Alert actions")
			if stacked {
				actions.FillWidth()
			} else {
				actions.Width(max(44, s.actionWidth)).Margin(-2, 0, -1, 0).MarginEnd(-4)
			}
			actions.Children(func() {
				AlertAction(c.Key(alertActionContentKey), slots.Action).MinWidth(44).Draw(func(_ *ui.Painter, bounds ui.Rect) {
					if s.actionWidth != bounds.W || s.actionHeight != bounds.H {
						s.actionWidth, s.actionHeight = bounds.W, bounds.H
						c.Invalidate()
					}
				})
			})
		}
	})
	return root
}

func AlertTitle(c *ui.Context, title string) ui.Element {
	return ui.Text(c, title).FontSize(14).FontWeight(500).LineHeight(1.5)
}

func alertDescriptionColor(c *ui.Context, destructive bool) ui.Color {
	t := theme.TokensOf(c)
	if destructive {
		return t.Destructive.Alpha(.9)
	}
	return t.MutedForeground
}

func AlertDescription(c *ui.Context, description string, destructive ...bool) ui.Element {
	red := len(destructive) > 0 && destructive[0]
	return ui.Text(c, description).FontSize(14).LineHeight(1.5).TextColor(alertDescriptionColor(c, red))
}

func AlertAction(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Row(c).Shrink(0).Gap(8).AlignItems(ui.Center).AlignSelf(ui.Start).Children(func() {
		if content != nil {
			content(c)
		}
	})
}
