package navigation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

func menuItems(items []NavItem) []overlay.MenuItem {
	out := make([]overlay.MenuItem, len(items))
	for i, it := range items {
		out[i] = overlay.MenuItem{Key: it.Key, Label: it.Label, Icon: it.Icon, Disabled: it.Disabled, Group: it.Group, Children: menuItems(it.Children)}
	}
	return out
}

type navMenuState struct{ open []bool }

// InWindowNavigationMenu explicitly confines route dropdowns to the parent
// window. NativeNavigationMenu and NavigationMenuWith provide independent menus.
func InWindowNavigationMenu(c *ui.Context, items []NavItem) string {
	chosen := ""
	root := ui.Row(c).Gap(4).Role(ui.RoleNavigation).Label("Main navigation").FocusGroup(ui.Horizontal)
	s := ui.Local(root, "shadcn.navigation-menu", func() navMenuState { return navMenuState{} })
	if len(s.open) != len(items) {
		s.open = make([]bool, len(items))
	}
	root.Children(func() {
		for i, it := range items {
			if len(it.Children) > 0 {
				key := overlay.InWindowDropdownMenu(c.Key(i), &s.open[i], overlay.MenuProps{Items: menuItems(it.Children)}, func(b ui.Element) { navButton(c, b.Disabled(it.Disabled), it.Label+" ⌄") })
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

// NavigationMenu retains the original parent-bounded route menu.
// Deprecated: use NavigationMenuWith for an independent shadcn Flyout, or
// NativeNavigationMenu for operating-system route menus.
func NavigationMenu(c *ui.Context, items []NavItem) string { return InWindowNavigationMenu(c, items) }
