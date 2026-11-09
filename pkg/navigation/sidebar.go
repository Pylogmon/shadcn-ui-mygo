package navigation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"strconv"
)

type SidebarProps struct {
	Items     []NavItem
	Label     string
	Width     float32
	Collapsed *bool
}

// Sidebar renders grouped route buttons and an optional collapse toggle.
// selected is the current route key and changes when an item is activated.
// It returns the container so callers can choose height and placement.
func Sidebar(c *ui.Context, selected *string, p SidebarProps, footer func()) ui.Element {
	t := theme.TokensOf(c)
	width := p.Width
	if width <= 0 {
		width = 256
	}
	collapsed := p.Collapsed != nil && *p.Collapsed
	if collapsed {
		width = 56
	}
	root := ui.Column(c).Width(width).Gap(8).Padding(8).Background(t.Sidebar).TextColor(t.SidebarForeground).Border(1, t.SidebarBorder).Role(ui.RoleNavigation).Label(p.Label)
	root.Children(func() {
		if p.Collapsed != nil {
			label := "Collapse sidebar"
			if collapsed {
				label = "Expand sidebar"
			}
			b := ui.ButtonBase(c.Key("collapse")).Label(label).Height(32).Width(32).Radius(t.Radius).Center()
			if b.Hovered() {
				b.Background(t.SidebarAccent).TextColor(t.SidebarAccentForeground)
			}
			b.Children(func() { ui.Text(c, "☰").FontSize(16) })
			if b.Clicked() {
				*p.Collapsed = !*p.Collapsed
			}
		}
		ui.Scroll(c).Grow(1).Gap(2).Children(func() { sidebarItems(c, selected, p.Items, collapsed, 0) })
		if footer != nil {
			ui.Column(c).Padding(8, 0, 0, 0).Children(footer)
		}
	})
	return root
}

func sidebarItems(c *ui.Context, selected *string, items []NavItem, collapsed bool, depth int) {
	t := theme.TokensOf(c)
	group := ""
	for i, it := range items {
		if it.Group != "" && it.Group != group && !collapsed {
			group = it.Group
			ui.Text(c, group).Padding(8).FontSize(12).FontWeight(500).TextColor(t.MutedForeground)
		}
		b := motion.ApplyTransition(c, ui.ButtonBase(c.Key(it.Key+strconv.Itoa(i))).Label(it.Label).Description(it.Description).Disabled(it.Disabled).
			Height(34).PaddingX(8).Gap(8).Justify(ui.Start).Radius(t.Radius).TextColor(t.SidebarForeground).FocusRingColor(t.SidebarRing))
		if depth > 0 && !collapsed {
			b.MarginStart(float32(depth) * 12)
		}
		if it.Key == *selected {
			b.Background(t.SidebarAccent).TextColor(t.SidebarAccentForeground).Current(ui.CurrentPage)
		} else if b.Hovered() {
			b.Background(t.SidebarAccent).TextColor(t.SidebarAccentForeground)
		}
		if it.Disabled {
			b.Opacity(0.5)
		}
		b.Children(func() {
			if it.Icon != nil {
				ui.Icon(c, it.Icon).Size(16, 16)
			} else if collapsed {
				ui.Text(c, initial(it.Label)).Width(16)
			}
			if !collapsed {
				ui.Text(c, it.Label).FontSize(14).SingleLine()
			}
		})
		if b.Clicked() {
			*selected = it.Key
		}
		if !collapsed && len(it.Children) > 0 {
			sidebarItems(c, selected, it.Children, false, depth+1)
		}
	}
}

func initial(label string) string {
	for _, r := range label {
		return string(r)
	}
	return ""
}
