// Package navigation provides native shadcn navigation components.
package navigation

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type BreadcrumbItem struct {
	Key, Label string
	Disabled   bool
}

// Breadcrumb returns the Key of an activated ancestor; the last crumb is the
// current page, exposed to assistive technology as such rather than a button.
func Breadcrumb(c *ui.Context, items []BreadcrumbItem) string {
	t := theme.TokensOf(c)
	chosen := ""
	ui.Row(c).Gap(8).AlignItems(ui.Center).Role(ui.RoleNavigation).Label("Breadcrumb").Children(func() {
		for i, item := range items {
			if i > 0 {
				ui.Text(c, "›").FontSize(14).TextColor(t.MutedForeground).AccessHidden(true)
			}
			if i == len(items)-1 {
				ui.Text(c, item.Label).FontSize(14).TextColor(t.Foreground).Current(ui.CurrentPage)
				continue
			}
			b := navButton(c, ui.ButtonBase(c.Key(i)).Disabled(item.Disabled), item.Label).Height(28).PaddingX(0).TextColor(t.MutedForeground)
			if b.Hovered() {
				b.TextColor(t.Foreground)
			}
			if b.Clicked() {
				chosen = item.Key
			}
		}
	})
	return chosen
}

func navButton(c *ui.Context, b ui.Element, label string) ui.Element {
	t := theme.TokensOf(c)
	b.Height(36).PaddingX(12).Radius(t.Radius).Gap(8).FontSize(14).FontWeight(500).TextColor(t.Foreground).Transition(motion.Transition(c))
	if b.Hovered() {
		b.Background(t.Accent).TextColor(t.AccentForeground)
	}
	b.Children(func() { ui.Text(c, label).SingleLine() })
	return b
}

type PaginationProps struct {
	Pages    int
	Siblings int
}

// Pagination uses one-based page numbers and bounds both previous/next and
// numbered links. It reports a change in the frame a page is activated.
func Pagination(c *ui.Context, page *int, p PaginationProps) bool {
	if p.Pages <= 0 {
		return false
	}
	*page = max(1, min(*page, p.Pages))
	before := *page
	t := theme.TokensOf(c)
	ui.Row(c).Gap(4).AlignItems(ui.Center).Role(ui.RoleNavigation).Label("Pagination").Children(func() {
		if navButton(c, ui.ButtonBase(c.Key("previous")).Disabled(*page <= 1), "‹ Previous").Clicked() {
			*page--
		}
		siblings := p.Siblings
		if siblings <= 0 {
			siblings = 1
		}
		last := 0
		for n := 1; n <= p.Pages; n++ {
			if n != 1 && n != p.Pages && (n < *page-siblings || n > *page+siblings) {
				continue
			}
			if last > 0 && n-last > 1 {
				ui.Text(c, "…").Width(36).TextAlign(ui.Center).TextColor(t.MutedForeground).AccessHidden(true)
			}
			b := navButton(c, ui.ButtonBase(c.Key(n)).Label("Page "+strconv.Itoa(n)), strconv.Itoa(n)).Width(36).PaddingX(0)
			if n == *page {
				b.Border(1, t.Border).Current(ui.CurrentPage)
			}
			if b.Clicked() {
				*page = n
			}
			last = n
		}
		if navButton(c, ui.ButtonBase(c.Key("next")).Disabled(*page >= p.Pages), "Next ›").Clicked() {
			*page++
		}
	})
	return before != *page
}

// NavItem is a route entry. Key is returned on activation; applications decide
// whether that changes an in-app route or opens an external URL.
type NavItem struct {
	Key, Label, Description, Group string
	Icon                           *ui.SVG
	Disabled                       bool
	Children                       []NavItem
}

func menuItems(items []NavItem) []overlay.MenuItem {
	out := make([]overlay.MenuItem, len(items))
	for i, it := range items {
		out[i] = overlay.MenuItem{Key: it.Key, Label: it.Label, Icon: it.Icon, Disabled: it.Disabled, Group: it.Group, Children: menuItems(it.Children)}
	}
	return out
}

type navMenuState struct{ open []bool }

// NavigationMenu provides a horizontal route menu with keyboard accessible
// dropdowns. Groups and nested items are represented by native menu semantics.
func NavigationMenu(c *ui.Context, items []NavItem) string {
	chosen := ""
	root := ui.Row(c).Gap(4).Role(ui.RoleNavigation).Label("Main navigation").FocusGroup(ui.Horizontal)
	s := ui.Local(root, "shadcn.navigation-menu", func() navMenuState { return navMenuState{} })
	if len(s.open) != len(items) {
		s.open = make([]bool, len(items))
	}
	root.Children(func() {
		for i, it := range items {
			if len(it.Children) > 0 {
				key := overlay.DropdownMenu(c.Key(i), &s.open[i], overlay.MenuProps{Items: menuItems(it.Children)}, func(b ui.Element) { navButton(c, b.Disabled(it.Disabled), it.Label+" ⌄") })
				if key != "" {
					chosen = key
				}
			} else if navButton(c, ui.ButtonBase(c.Key(i)).Disabled(it.Disabled), it.Label).Clicked() {
				chosen = it.Key
			}
		}
	})
	return chosen
}

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
		b := ui.ButtonBase(c.Key(it.Key + strconv.Itoa(i))).Label(it.Label).Description(it.Description).Disabled(it.Disabled).
			Height(34).PaddingX(8).Gap(8).Justify(ui.Start).Radius(t.Radius).TextColor(t.SidebarForeground).FocusRingColor(t.SidebarRing).Transition(motion.Transition(c))
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
