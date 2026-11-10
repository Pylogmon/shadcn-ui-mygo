package overlay

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// MenuItem supports nested submenus, checkbox items and radio selections.
// Key identifies the result; Shortcut is display text rather than a key binding.
type MenuItem struct {
	Key, Label, Shortcut string
	Icon                 *ui.SVG
	// NativeIcon is a reusable platform image made by mygo.NewMenuIconSVG.
	NativeIcon                       ui.MenuIcon
	Disabled, Destructive, Separator bool
	Group                            string
	// Inset increases logical-start padding from 8 to 32 pixels. GroupInset
	// applies independently to the group caption introduced by this item.
	// Native system menus ignore both visual-only flags.
	Inset, GroupInset bool
	// CloseOnSelect overrides the styled desktop menu default: actions close,
	// checkboxes and radio items stay open. KeepOpenOnSelect still takes priority.
	// Native menus and legacy InWindow menus retain MyGo selection tracking.
	CloseOnSelect *bool
	Checked       *bool
	Radio         bool
	Selected      bool
	Children      []MenuItem
}

type MenuProps struct {
	Items []MenuItem
	Width float32
}

func menuPanel(c *ui.Context, panel ui.Element, width float32) {
	profile := theme.MenuOf(c).ForScene()
	t := &profile.Tokens
	panel.Column().Padding(4).Gap(0).Radius(t.RadiusMD()).FontSize(14).Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Transition(motion.Menu(c))
	profile.Surface(c, panel)
	if width <= 0 {
		width = 192
	}
	panel.MinWidth(width)
}

// InWindowDropdownMenu opens on pointer press or Enter, Space, Up and Down. The menu
// supports typeahead, disabled-item skipping, nested arrows and Escape.
func InWindowDropdownMenu(c *ui.Context, open *bool, p MenuProps, trigger func(ui.Element)) string {
	m := ui.MenuBase(c, open)
	if trigger != nil {
		trigger(m.Trigger)
	} else {
		button(c, m.Trigger, "Open menu", false)
	}
	chosen := ""
	m.Popup(func(panel ui.Element) {
		menuPanel(c, panel, p.Width)
		panel.Children(func() {
			profile := theme.MenuOf(c).ForScene()
			ui.Themed(c, profile.Theme()).FillWidth().Children(func() { menuItems(c, m, p.Items, p.Width, &chosen) })
		})
	})
	return chosen
}

// InWindowContextMenu opens on secondary click, Control-click on macOS, or Shift+F10.
func InWindowContextMenu(c *ui.Context, target ui.Element, p MenuProps) string {
	open := ui.Local(target, "shadcn.context-menu", func() bool { return false })
	m := ui.ContextMenuBase(c, target, open)
	chosen := ""
	m.Popup(func(panel ui.Element) {
		menuPanel(c, panel, p.Width)
		panel.Children(func() {
			profile := theme.MenuOf(c).ForScene()
			ui.Themed(c, profile.Theme()).FillWidth().Children(func() { menuItems(c, m, p.Items, p.Width, &chosen) })
		})
	})
	return chosen
}

func menuItems(c *ui.Context, m ui.MenuParts, items []MenuItem, width float32, chosen *string) {
	profile := theme.MenuOf(c)
	t := &profile.Tokens
	group := ""
	for i := range items {
		it := items[i]
		if it.Separator {
			ui.Box(c.Key(i)).Height(1).Margin(4, -4).Background(profile.Separator())
			continue
		}
		if it.Group != "" && it.Group != group {
			group = it.Group
			caption := ui.Text(c, group).Padding(6, 8).FontSize(12).FontWeight(500).TextColor(t.MutedForeground)
			if it.GroupInset {
				caption.PaddingStart(32)
			}
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
		row.Label(it.Label).Disabled(it.Disabled).Height(32).Padding(6, 8).Gap(8).Radius(max(2, t.Radius-2)).FontSize(14).TextColor(t.PopoverForeground)
		if it.Inset {
			row.PaddingStart(32)
		}
		if it.Checked != nil || it.Radio {
			row.PaddingEnd(32)
		}
		if profile.Animated() {
			motion.ApplyTransition(c, row)
		}
		if row.Highlighted() {
			bg, fg := profile.Highlight(it.Destructive)
			row.Background(bg).TextColor(fg)
		}
		if it.Destructive {
			row.TextColor(t.Destructive)
			if profile.Translucent() {
				row.TextColor(t.AccentForeground)
			}
			if row.Highlighted() {
				bg, fg := profile.Highlight(true)
				row.Background(bg).TextColor(fg)
			}
		}
		if it.Disabled {
			row.Opacity(0.5)
		}
		row.Children(func() {
			if it.Icon != nil {
				ui.Icon(c, it.Icon).Size(16, 16).AccessHidden(true)
			}
			if (it.Checked != nil && *it.Checked) || (it.Radio && it.Selected) {
				ui.Box(c).Absolute().End(8).TopPercent(50).Size(16, 16).Translate(0, -8).Children(func() { indicator.Check(c) })
			}
			ui.Text(c, it.Label).Grow(1).SingleLine()
			if it.Shortcut != "" {
				shortcutColor := t.MutedForeground
				if it.Destructive && profile.Translucent() {
					shortcutColor = t.AccentForeground
				}
				ui.Text(c, it.Shortcut).FontSize(12).TextColor(shortcutColor)
			}
			if len(it.Children) > 0 {
				angle := float32(-90)
				if c.Direction() == ui.RightToLeft {
					angle = 90
				}
				indicator.ChevronDown(c).Rotate(angle)
			}
		})
		if len(it.Children) > 0 {
			m.Submenu(row, func(panel ui.Element) {
				menuPanel(c, panel, width)
				panel.Children(func() {
					profile := theme.MenuOf(c).ForScene()
					ui.Themed(c, profile.Theme()).FillWidth().Children(func() { menuItems(c, m, it.Children, width, chosen) })
				})
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
	Key      string
	Label    string
	Items    []MenuItem
	Disabled bool
}
type menubarState struct{ open []bool }

// InWindowMenubar is a horizontal group of native menu triggers. Tab enters the group,
// Left and Right move among triggers, and each menu has full menu navigation.
func InWindowMenubar(c *ui.Context, items []MenubarItem) string {
	t := theme.TokensOf(c)
	root := ui.Row(c).Gap(2).Padding(3).Radius(t.Radius).Border(1, t.Border).Background(t.Background).FocusGroup(ui.Horizontal).Role(ui.RoleMenuBar).Label("Menu bar")
	s := ui.Local(root, "shadcn.menubar", func() menubarState { return menubarState{} })
	if len(s.open) != len(items) {
		s.open = make([]bool, len(items))
	}
	chosen := ""
	root.Children(func() {
		for i, item := range items {
			value := InWindowDropdownMenu(c.Key(i), &s.open[i], MenuProps{Items: item.Items}, func(trigger ui.Element) {
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
