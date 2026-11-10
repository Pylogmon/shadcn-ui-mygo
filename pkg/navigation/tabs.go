package navigation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type TabItem struct {
	Key, Label string
	Icon       *ui.SVG
	Disabled   bool
}
type TabsVariant uint8

const (
	TabsDefault TabsVariant = iota
	TabsLine
)

type TabsProps struct {
	Items    []TabItem
	Vertical bool
	Label    string
	Variant  TabsVariant
	// Manual keeps arrow navigation separate from selection. Enter or Space
	// activates the focused tab, which suits panels that load asynchronously.
	Manual bool
	// NoLoop stops arrow navigation at the ends instead of wrapping.
	NoLoop bool
}
type tabsIndicatorKey struct{ selected *int }

func nextTab(items []TabItem, from, direction int, loop bool) int {
	n := len(items)
	for step := 1; step <= n; step++ {
		at := from + direction*step
		if !loop && (at < 0 || at >= n) {
			return from
		}
		at = (at%n + n) % n
		if !items[at].Disabled {
			return at
		}
	}
	return -1
}

// Tabs uses native buttons and a roving focus group. Only the orientation's
// arrows move between tabs; disabled tabs are skipped. The selected trigger
// controls and labels a focusable native region representing its tab panel.
func Tabs(c *ui.Context, selected *int, p TabsProps, panel func(int)) ui.Element {
	if len(p.Items) == 0 {
		*selected = -1
	} else if *selected < 0 || *selected >= len(p.Items) || p.Items[*selected].Disabled {
		*selected = nextTab(p.Items, -1, 1, false)
	}
	t := theme.TokensOf(c)
	root := ui.Column(c).Gap(8)
	if p.Vertical {
		root.Row()
	}
	root.Children(func() {
		orientation := ui.Horizontal
		if p.Vertical {
			orientation = ui.Vertical
		}
		list := ui.Row(c).Shrink(0).Padding(3).Gap(2).Radius(t.Radius).Background(t.Muted).TextColor(t.MutedForeground).
			Role(ui.RoleTabList).Label(p.Label).Orientation(orientation).FocusGroup(orientation).AlignSelf(ui.Start)
		if p.Vertical {
			list.Column().AlignItems(ui.Stretch)
		}
		if p.Variant == TabsLine {
			list.Background(ui.Transparent).Padding(0).Gap(8).Radius(0)
		}
		tabs := make([]ui.Element, len(p.Items))
		focus := -1
		list.Children(func() {
			for i, item := range p.Items {
				key := any(i)
				if item.Key != "" {
					key = item.Key
				}
				tab := motion.ApplyTransition(c, ui.ButtonBase(c.Key(key)).Role(ui.RoleTab).Label(item.Label).Disabled(item.Disabled).Checked(i == *selected).
					Height(30).PaddingX(12).Gap(6).Radius(t.RadiusMD()).FontSize(14).FontWeight(500).TextColor(t.MutedForeground))
				tabs[i] = tab
				back, forth := ui.KeyLeft, ui.KeyRight
				if c.Direction() == ui.RightToLeft {
					back, forth = forth, back
				}
				if p.Vertical {
					back, forth = ui.KeyUp, ui.KeyDown
				}
				switch {
				case tab.Shortcut(0, back):
					focus = nextTab(p.Items, i, -1, !p.NoLoop)
				case tab.Shortcut(0, forth):
					focus = nextTab(p.Items, i, 1, !p.NoLoop)
				case tab.Shortcut(0, ui.KeyHome):
					focus = nextTab(p.Items, -1, 1, false)
				case tab.Shortcut(0, ui.KeyEnd):
					focus = nextTab(p.Items, len(p.Items), -1, false)
				}
				if tab.Clicked() && !item.Disabled {
					*selected = i
				}
				if focus >= 0 && !p.Manual {
					*selected = focus
				}
				if item.Disabled {
					tab.DimDisabled(false).Opacity(.5)
				}
				tab.Children(func() {
					if i == *selected {
						indicator := ui.Box(c).Absolute().ZIndex(-1).IgnoreInput(true)
						if !motion.Reduced(c) {
							tr := motion.Transition(c)
							tr.Match = tabsIndicatorKey{selected}
							indicator.Transition(tr)
						}
						if p.Variant == TabsLine {
							if p.Vertical {
								indicator.Right(0).Top(0).Bottom(0).Width(2)
							} else {
								indicator.Left(0).Right(0).Bottom(0).Height(2)
							}
							indicator.Background(t.Foreground)
						} else {
							indicator.Top(0).Bottom(0).Left(0).Right(0).Radius(t.RadiusMD()).Background(t.Background).Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, .06))
							if t.Dark {
								indicator.Background(t.Input.Alpha(.3)).Border(1, t.Input)
							}
						}
						tab.TextColor(t.Foreground)
					}
					if item.Icon != nil {
						ui.Icon(c, item.Icon).Size(16, 16)
					}
					ui.Text(c, item.Label).SingleLine()
				})
			}
		})
		if focus >= 0 && focus < len(tabs) {
			tabs[focus].Focus()
		}
		// Selection may change after an earlier tab was built, so publish the
		// current state on every trigger before accessibility commits the frame.
		for i, tab := range tabs {
			tab.Checked(i == *selected)
		}
		if panel != nil && *selected >= 0 && *selected < len(p.Items) {
			active := *selected
			region := ui.Column(c.Key("panel")).Grow(1).MinWidth(0).Role(ui.RoleRegion).RoleDescription("tab panel").LabelledBy(tabs[active]).Focusable().Children(func() { panel(active) })
			tabs[active].Controls(region)
		}
	})
	return root
}
