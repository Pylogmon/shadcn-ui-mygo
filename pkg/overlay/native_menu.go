package overlay

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// NativeDropdownMenu uses MyGo's platform menu adapter. With the default
// ui.SystemMenus policy the OS owns the popup, so it may extend beyond the
// parent window. SetMenuPresentation(ui.PopoverMenus) deliberately opts into
// MyGo's in-window fallback, useful for embedded previews and headless demos.
func NativeDropdownMenu(c *ui.Context, p MenuProps, trigger func(ui.Element)) string {
	chosen := ""
	b := ui.ButtonBase(c)
	if trigger != nil {
		trigger(b)
	} else {
		button(c, b, "Open menu", false)
	}
	b.Menu(func(m *ui.Menu) { nativeMenuItems(m, p.Items, "", &chosen) })
	return chosen
}

// NativeContextMenu preserves the system's context-key/secondary-click menu,
// including nested submenus, checked/radio states and disabled actions.
func NativeContextMenu(c *ui.Context, target ui.Element, p MenuProps) string {
	chosen := ""
	target.ContextMenu(func(m *ui.Menu) { nativeMenuItems(m, p.Items, "", &chosen) })
	return chosen
}

// NativeMenubar is a group of native popup triggers. It is distinct from an
// application's top-level operating-system menu bar, owned by mygo.Window.
func NativeMenubar(c *ui.Context, items []MenubarItem) string {
	t := theme.TokensOf(c)
	chosen := ""
	ui.Row(c).Gap(2).Padding(3).Radius(t.RadiusMD()).Border(1, t.Border).Background(t.Background).
		FocusGroup(ui.Horizontal).Role(ui.RoleMenuBar).Label("Menu bar").Children(func() {
		for i, item := range items {
			key := any(i)
			if item.Key != "" {
				key = item.Key
			}
			value := NativeDropdownMenu(c.Key(key), MenuProps{Items: item.Items}, func(b ui.Element) {
				button(c, b.Disabled(item.Disabled), item.Label, false).Height(28).Border(0, ui.Transparent)
			})
			if value != "" {
				chosen = value
			}
		}
	})
	return chosen
}

func nativeMenuItems(m *ui.Menu, items []MenuItem, path string, chosen *string) {
	group := ""
	for i, item := range items {
		if item.Separator {
			m.Separator()
			continue
		}
		if item.Group != "" && item.Group != group {
			group = item.Group
			m.Item(group).Disabled(true)
		}
		id := item.Key
		if id == "" {
			id = path + strconv.Itoa(i)
		}
		var row *ui.MenuItem
		if len(item.Children) > 0 {
			row = m.Submenu(item.Label, func(sub *ui.Menu) { nativeMenuItems(sub, item.Children, id+"/", chosen) })
		} else {
			row = m.Item(item.Label)
		}
		row.ID(id).Disabled(item.Disabled)
		if item.Checked != nil {
			row.Checked(*item.Checked)
		} else if item.Radio {
			row.Radio(item.Selected)
		}
		if item.NativeIcon != nil {
			row.Icon(item.NativeIcon)
		}
		if item.Destructive {
			row.Danger()
		}
		if item.Shortcut != "" {
			row.Trailing(item.Shortcut)
		}
		if row.Chosen() && !item.Disabled && len(item.Children) == 0 {
			if item.Checked != nil {
				*item.Checked = !*item.Checked
			}
			*chosen = item.Key
		}
	}
}
