package desktop

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// MenuState belongs to the application. Path identifies open submenu parents;
// Highlighted holds one item key per panel. Keys need be unique only among
// siblings, so duplicate leaf keys in different submenus remain independent.
// Empty MenuItem keys get a positional identity for navigation only.
type MenuState struct {
	Path, Highlighted  []string
	ActiveDepth        int
	Chosen             string
	ChosenPath         []string
	search             string
	searchAt           time.Time
	focus, reveal      bool
	pointerOver        bool
	pointerX, pointerY float32
	wasOpen            bool
	contextX, contextY float32
	contextFocus       ui.Handle
}

func (s *MenuState) Reset() { *s = MenuState{focus: true} }

type MenuContentProps struct {
	Label                         string
	Width, MaxHeight              float32
	NoAutoFocus, KeepOpenOnSelect bool
	OnChoose                      func(string)
	OnClose                       func()
	ItemContent                   func(*ui.Context, overlay.MenuItem)
}

func menuItemKey(item overlay.MenuItem, index int) string {
	if item.Key != "" {
		return item.Key
	}
	return fmt.Sprintf("@%d", index)
}
func menuItemIndex(items []overlay.MenuItem, key string) int {
	for i, item := range items {
		if menuItemKey(item, i) == key {
			return i
		}
	}
	return -1
}
func menuEnabled(item overlay.MenuItem) bool { return !item.Disabled && !item.Separator }
func firstMenuItem(items []overlay.MenuItem) string {
	for i, item := range items {
		if menuEnabled(item) {
			return menuItemKey(item, i)
		}
	}
	return ""
}
func menuLevels(state *MenuState, items []overlay.MenuItem) [][]overlay.MenuItem {
	levels := [][]overlay.MenuItem{items}
	for depth, key := range state.Path {
		index := menuItemIndex(levels[depth], key)
		if index < 0 || !menuEnabled(levels[depth][index]) || len(levels[depth][index].Children) == 0 {
			state.Path = state.Path[:depth]
			break
		}
		levels = append(levels, levels[depth][index].Children)
	}
	if len(state.Highlighted) > len(levels) {
		state.Highlighted = state.Highlighted[:len(levels)]
	}
	for len(state.Highlighted) < len(levels) {
		state.Highlighted = append(state.Highlighted, "")
	}
	for depth, list := range levels {
		if depth < len(state.Path) {
			state.Highlighted[depth] = state.Path[depth]
		} else {
			index := menuItemIndex(list, state.Highlighted[depth])
			if index < 0 || !menuEnabled(list[index]) {
				state.Highlighted[depth] = firstMenuItem(list)
			}
		}
	}
	state.ActiveDepth = max(0, min(state.ActiveDepth, len(levels)-1))
	return levels
}
func menuAtPath(items []overlay.MenuItem, path []string, depth int) []overlay.MenuItem {
	for i := 0; i < depth && i < len(path); i++ {
		index := menuItemIndex(items, path[i])
		if index < 0 {
			return nil
		}
		items = items[index].Children
	}
	return items
}
func menuHighlight(state *MenuState, depth int, key string) {
	for len(state.Highlighted) <= depth {
		state.Highlighted = append(state.Highlighted, "")
	}
	state.Highlighted[depth] = key
	state.ActiveDepth = depth
	if len(state.Path) > depth && state.Path[depth] != key {
		state.Path = state.Path[:depth]
		state.Highlighted = state.Highlighted[:depth+1]
	}
	state.focus, state.reveal = true, true
}
func menuOpenChild(state *MenuState, items []overlay.MenuItem, depth int, focus bool) {
	list := menuAtPath(items, state.Path, depth)
	if depth >= len(state.Highlighted) {
		return
	}
	index := menuItemIndex(list, state.Highlighted[depth])
	if index < 0 || !menuEnabled(list[index]) || len(list[index].Children) == 0 {
		return
	}
	state.Path = append(slices.Clone(state.Path[:min(depth, len(state.Path))]), menuItemKey(list[index], index))
	if len(state.Highlighted) > depth+1 {
		state.Highlighted = state.Highlighted[:depth+1]
	}
	state.Highlighted = append(state.Highlighted, firstMenuItem(list[index].Children))
	if focus {
		state.ActiveDepth = depth + 1
		state.focus, state.reveal = true, true
	}
}
func menuCloseChild(state *MenuState) {
	if state.ActiveDepth <= 0 {
		return
	}
	state.Path = state.Path[:state.ActiveDepth-1]
	state.ActiveDepth--
	state.Highlighted = state.Highlighted[:state.ActiveDepth+1]
	state.focus, state.reveal = true, true
	state.search = ""
}
func menuActivate(state *MenuState, items []overlay.MenuItem, depth int, p MenuContentProps) {
	list := menuAtPath(items, state.Path, depth)
	if depth >= len(state.Highlighted) {
		return
	}
	index := menuItemIndex(list, state.Highlighted[depth])
	if index < 0 || !menuEnabled(list[index]) {
		return
	}
	item := &list[index]
	if len(item.Children) > 0 {
		menuOpenChild(state, items, depth, true)
		return
	}
	if item.Checked != nil {
		*item.Checked = !*item.Checked
	} else if item.Radio {
		for i := range list {
			if list[i].Radio && list[i].Group == item.Group {
				list[i].Selected = i == index
			}
		}
	}
	state.Chosen = item.Key
	state.ChosenPath = append(slices.Clone(state.Path[:min(depth, len(state.Path))]), menuItemKey(*item, index))
	if p.OnChoose != nil {
		p.OnChoose(item.Key)
	}
	if !p.KeepOpenOnSelect && p.OnClose != nil {
		p.OnClose()
	}
}

func menuInput(state *MenuState, items []overlay.MenuItem, depth int, p MenuContentProps, rtl bool, ev ui.InputEvent) bool {
	list := menuAtPath(items, state.Path, depth)
	if len(list) == 0 {
		return false
	}
	menuLevels(state, items)
	current := menuItemIndex(list, state.Highlighted[depth])
	openKey, closeKey := ui.KeyRight, ui.KeyLeft
	if rtl {
		openKey, closeKey = closeKey, openKey
	}
	if ev.Kind == ui.InputText {
		text := strings.ToLower(strings.TrimSpace(ev.Text))
		text = strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, text)
		if text == "" {
			return false
		}
		if ev.Time.Sub(state.searchAt) > 700*time.Millisecond || ev.Time.Before(state.searchAt) {
			state.search = ""
		}
		old := state.search
		state.search += text
		state.searchAt = ev.Time
		prefix := state.search
		runes := []rune(prefix)
		repeat := len(runes) > 1
		for _, r := range runes {
			if r != runes[0] {
				repeat = false
				break
			}
		}
		if repeat {
			prefix = string(runes[0])
		}
		start := current + 1
		if old != "" && !repeat {
			start = max(0, current)
		}
		for offset := range len(list) {
			index := (max(0, start) + offset) % len(list)
			if menuEnabled(list[index]) && strings.HasPrefix(strings.ToLower(list[index].Label), prefix) {
				menuHighlight(state, depth, menuItemKey(list[index], index))
				break
			}
		}
		return true
	}
	if ev.Kind != ui.InputKeyDown {
		return false
	}
	switch ev.Key {
	case ui.KeyUp, ui.KeyDown:
		delta := 1
		if ev.Key == ui.KeyUp {
			delta = -1
		}
		index := current
		if index < 0 {
			if delta > 0 {
				index = -1
			} else {
				index = 0
			}
		}
		for range len(list) {
			index = (index + delta + len(list)) % len(list)
			if menuEnabled(list[index]) {
				menuHighlight(state, depth, menuItemKey(list[index], index))
				break
			}
		}
		state.search = ""
		return true
	case ui.KeyHome, ui.KeyEnd:
		index, delta := 0, 1
		if ev.Key == ui.KeyEnd {
			index, delta = len(list)-1, -1
		}
		for index >= 0 && index < len(list) {
			if menuEnabled(list[index]) {
				menuHighlight(state, depth, menuItemKey(list[index], index))
				break
			}
			index += delta
		}
		state.search = ""
		return true
	case openKey:
		menuOpenChild(state, items, depth, true)
		return true
	case closeKey:
		if depth > 0 {
			state.ActiveDepth = depth
			menuCloseChild(state)
			return true
		}
	case ui.KeyEnter, ui.KeySpace:
		menuActivate(state, items, depth, p)
		return true
	case ui.KeyEscape:
		if depth > 0 {
			state.ActiveDepth = depth
			menuCloseChild(state)
		} else if p.OnClose != nil {
			p.OnClose()
		}
		return true
	case ui.KeyTab:
		if p.OnClose != nil {
			p.OnClose()
		}
		return true
	}
	return false
}

// MenuContent lays out real, measured native controls; it never creates an
// absolute in-window menu overlay. Adjacent submenu panels therefore contribute
// to a Flyout's FitContent size and may extend beyond its parent window.
func MenuContent(c *ui.Context, state *MenuState, items []overlay.MenuItem, p MenuContentProps) ui.Element {
	if state == nil {
		panic("desktop: MenuContent requires caller-owned MenuState")
	}
	t := theme.TokensOf(c)
	levels := menuLevels(state, items)
	width := p.Width
	if width <= 0 {
		width = 192
	}
	maxHeight := p.MaxHeight
	if maxHeight <= 0 {
		maxHeight = 480
	}
	label := p.Label
	if label == "" {
		label = "Menu"
	}
	root := ui.Row(c).Gap(4).AlignItems(ui.Start).Role(ui.RoleGroup).Label(label)
	x, y, over := root.PointerPosition()
	moved := over && (!state.pointerOver || state.pointerX != x || state.pointerY != y)
	state.pointerOver, state.pointerX, state.pointerY = over, x, y
	rtl := c.Direction() == ui.RightToLeft
	root.Children(func() {
		for depth, list := range levels {
			panelLabel := label
			if depth > 0 {
				panelLabel = listLabel(items, state.Path, depth)
			}
			panel := ui.Column(c.Key(fmt.Sprintf("%q", state.Path[:min(depth, len(state.Path))]))).Width(width).Padding(4).Radius(t.RadiusMD()).Background(t.Popover).TextColor(t.PopoverForeground).Border(1, t.Border).Role(ui.RoleMenu).Label(panelLabel).Focusable().InputMethod(true).TabStop(depth == state.ActiveDepth).HandleInput(func(ev ui.InputEvent) bool { return menuInput(state, items, depth, p, rtl, ev) })
			if !p.NoAutoFocus && depth == 0 {
				panel.AutoFocus()
			}
			if !p.NoAutoFocus && depth == state.ActiveDepth && !root.FocusWithin() {
				panel.Focus()
			}
			if state.focus && depth == state.ActiveDepth {
				panel.Focus()
			}
			panel.Children(func() {
				height := menuContentHeight(list)
				body := ui.Column(c).Width(width - 8)
				if height > maxHeight-8 {
					body = ui.Scroll(c).Width(width - 8).Height(max(32, maxHeight-8))
				}
				body.Children(func() {
					group := ""
					for i, item := range list {
						key := menuItemKey(item, i)
						if item.Separator {
							ui.Box(c.Key(key)).Height(1).Margin(4, 0).Background(t.Border).Role(ui.RoleSeparator)
							continue
						}
						if item.Group != "" && item.Group != group {
							group = item.Group
							ui.Text(c, group).Padding(6, 8).FontSize(12).FontWeight(500).TextColor(t.MutedForeground)
						}
						role := ui.RoleMenuItem
						if item.Checked != nil {
							role = ui.RoleMenuItemCheckBox
						} else if item.Radio {
							role = ui.RoleMenuItemRadio
						}
						row := ui.Row(c.Key(key)).FillWidth().Height(32).PaddingX(8).Gap(8).Radius(max(2, t.Radius-2)).AlignItems(ui.Center).Role(role).Label(item.Label).Disabled(item.Disabled).FontSize(14).TextColor(t.PopoverForeground).Transition(motion.Transition(c))
						highlighted := depth < len(state.Highlighted) && key == state.Highlighted[depth]
						if highlighted {
							row.Background(t.Accent).TextColor(t.AccentForeground)
							if depth == state.ActiveDepth {
								panel.ActiveDescendant(row)
								if state.reveal {
									row.ScrollIntoView()
								}
							}
						}
						if item.Checked != nil {
							row.Checked(*item.Checked)
						} else if item.Radio {
							row.Checked(item.Selected)
						}
						if len(item.Children) > 0 {
							row.HasPopup(ui.PopupMenu).Expanded(depth < len(state.Path) && state.Path[depth] == key)
						}
						if item.Destructive {
							row.TextColor(t.Destructive)
							if highlighted {
								row.Background(t.Destructive.Alpha(.1))
							}
						}
						if item.Disabled {
							row.Opacity(.5).Cursor(ui.CursorNotAllowed)
						}
						row.Children(func() {
							if item.Checked != nil || item.Radio {
								mark := ""
								if item.Checked != nil && *item.Checked {
									mark = "✓"
								}
								if item.Radio && item.Selected {
									mark = "●"
								}
								ui.Text(c, mark).Width(16).FontSize(12).AccessHidden(true)
							} else if item.Icon != nil {
								ui.Icon(c, item.Icon).Size(16, 16).AccessHidden(true)
							}
							if p.ItemContent != nil {
								ui.Column(c).Grow(1).MinWidth(0).Children(func() { p.ItemContent(c, item) })
							} else {
								ui.Text(c, item.Label).Grow(1).MinWidth(0).SingleLine()
							}
							if item.Shortcut != "" {
								ui.Text(c, item.Shortcut).FontSize(12).TextColor(t.MutedForeground).MarginStart(20)
							}
							if len(item.Children) > 0 {
								arrow := "›"
								if rtl {
									arrow = "‹"
								}
								ui.Text(c, arrow).FontSize(16).AccessHidden(true)
							}
						})
						hovered := row.Hovered()
						if moved && hovered && !item.Disabled {
							menuHighlight(state, depth, key)
							if len(item.Children) > 0 {
								menuOpenChild(state, items, depth, false)
							} else if len(state.Path) > depth {
								state.Path = state.Path[:depth]
							}
							c.Invalidate()
						}
						if row.Clicked() && !item.Disabled {
							menuHighlight(state, depth, key)
							panel.Focus()
							menuActivate(state, items, depth, p)
							c.Invalidate()
						}
					}
				})
			})
		}
	})
	state.focus, state.reveal = false, false
	return root
}
func listLabel(items []overlay.MenuItem, path []string, depth int) string {
	label := "Submenu"
	for i := 0; i < depth && i < len(path); i++ {
		index := menuItemIndex(items, path[i])
		if index < 0 {
			break
		}
		label = items[index].Label
		items = items[index].Children
	}
	return label
}
func menuContentHeight(items []overlay.MenuItem) float32 {
	height := float32(0)
	group := ""
	for _, item := range items {
		if item.Separator {
			height += 9
			continue
		}
		if item.Group != "" && item.Group != group {
			group = item.Group
			height += 28
		}
		height += 32
	}
	return height
}
