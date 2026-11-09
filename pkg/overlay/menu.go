package overlay

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// MenuItem supports nested submenus, checkbox items and radio selections.
// Key identifies the result; Shortcut is display text rather than a key binding.
type MenuItem struct {
	Key, Label, Shortcut             string
	Icon                             *ui.SVG
	Disabled, Destructive, Separator bool
	Group                            string
	Checked                          *bool
	Radio                            bool
	Selected                         bool
	Children                         []MenuItem
}

type MenuProps struct {
	Items []MenuItem
	Width float32
}

func menuPanel(c *ui.Context, panel ui.Element, width float32) {
	surface(c, panel).Padding(4).Gap(0)
	if width <= 0 {
		width = 192
	}
	panel.MinWidth(width)
}

// DropdownMenu opens on pointer press or Enter, Space, Up and Down. The menu
// supports typeahead, disabled-item skipping, nested arrows and Escape.
func DropdownMenu(c *ui.Context, open *bool, p MenuProps, trigger func(ui.Element)) string {
	m := ui.MenuBase(c, open)
	if trigger != nil {
		trigger(m.Trigger)
	} else {
		button(c, m.Trigger, "Open menu", false)
	}
	chosen := ""
	m.Popup(func(panel ui.Element) {
		menuPanel(c, panel, p.Width)
		panel.Children(func() { menuItems(c, m, p.Items, p.Width, &chosen) })
	})
	return chosen
}

// ContextMenu opens on secondary click, Control-click on macOS, or Shift+F10.
func ContextMenu(c *ui.Context, target ui.Element, p MenuProps) string {
	open := ui.Local(target, "shadcn.context-menu", func() bool { return false })
	m := ui.ContextMenuBase(c, target, open)
	chosen := ""
	m.Popup(func(panel ui.Element) {
		menuPanel(c, panel, p.Width)
		panel.Children(func() { menuItems(c, m, p.Items, p.Width, &chosen) })
	})
	return chosen
}

func menuItems(c *ui.Context, m ui.MenuParts, items []MenuItem, width float32, chosen *string) {
	t := theme.TokensOf(c)
	group := ""
	for i := range items {
		it := items[i]
		if it.Separator {
			ui.Box(c.Key(i)).Height(1).Margin(4, -4).Background(t.Border)
			continue
		}
		if it.Group != "" && it.Group != group {
			group = it.Group
			ui.Text(c, group).Padding(6, 8).FontSize(12).FontWeight(500).TextColor(t.MutedForeground)
		}
		var row ui.Element
		switch {
		case len(it.Children) > 0:
			row = m.SubmenuItem()
		case it.Checked != nil:
			row = m.CheckItem(*it.Checked)
		case it.Radio:
			row = m.RadioItem(it.Selected)
		default:
			row = m.Item()
		}
		row.Label(it.Label).Disabled(it.Disabled).Height(32).Padding(6, 8).Gap(8).Radius(max(2, t.Radius-2)).
			FontSize(14).TextColor(t.PopoverForeground).Transition(motion.Transition(c))
		if row.Highlighted() {
			row.Background(t.Accent).TextColor(t.AccentForeground)
		}
		if it.Destructive {
			row.TextColor(t.Destructive)
			if row.Highlighted() {
				row.Background(t.Destructive.Alpha(0.1))
			}
		}
		if it.Disabled {
			row.Opacity(0.5)
		}
		row.Children(func() {
			if it.Checked != nil || it.Radio {
				mark := ""
				if it.Checked != nil && *it.Checked {
					mark = "✓"
				}
				if it.Radio && it.Selected {
					mark = "●"
				}
				ui.Text(c, mark).Width(16).FontSize(12)
			} else if it.Icon != nil {
				ui.Icon(c, it.Icon).Size(16, 16)
			}
			ui.Text(c, it.Label).Grow(1).SingleLine()
			if it.Shortcut != "" {
				ui.Text(c, it.Shortcut).FontSize(12).TextColor(t.MutedForeground).MarginStart(24)
			}
			if len(it.Children) > 0 {
				ui.Text(c, "›").FontSize(16)
			}
		})
		if len(it.Children) > 0 {
			m.Submenu(row, func(panel ui.Element) {
				menuPanel(c, panel, width)
				panel.Children(func() { menuItems(c, m, it.Children, width, chosen) })
			})
		}
		if row.Clicked() && !it.Disabled && len(it.Children) == 0 {
			if it.Checked != nil {
				*it.Checked = !*it.Checked
			}
			*chosen = it.Key
		}
	}
}

type MenubarItem struct {
	Label    string
	Items    []MenuItem
	Disabled bool
}
type menubarState struct{ open []bool }

// Menubar is a horizontal group of native menu triggers. Tab enters the group,
// Left and Right move among triggers, and each menu has full menu navigation.
func Menubar(c *ui.Context, items []MenubarItem) string {
	t := theme.TokensOf(c)
	root := ui.Row(c).Gap(2).Padding(3).Radius(t.Radius).Border(1, t.Border).Background(t.Background).FocusGroup(ui.Horizontal).Role(ui.RoleMenuBar).Label("Menu bar")
	s := ui.Local(root, "shadcn.menubar", func() menubarState { return menubarState{} })
	if len(s.open) != len(items) {
		s.open = make([]bool, len(items))
	}
	chosen := ""
	root.Children(func() {
		for i, item := range items {
			value := DropdownMenu(c.Key(i), &s.open[i], MenuProps{Items: item.Items}, func(trigger ui.Element) {
				button(c, trigger.Disabled(item.Disabled), item.Label, false).Border(0, ui.Transparent).Height(28)
				if s.open[i] {
					trigger.Background(t.Accent)
				}
				if trigger.Hovered() && !item.Disabled {
					for j := range s.open {
						if j != i && s.open[j] {
							s.open[j], s.open[i] = false, true
						}
					}
				}
			})
			if value != "" {
				chosen = value
			}
		}
	})
	return chosen
}
