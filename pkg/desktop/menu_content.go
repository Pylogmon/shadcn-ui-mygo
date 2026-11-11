package desktop

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
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
	selectList       bool
	fontSize         float32
	Label            string
	Width, MaxHeight float32
	// ItemHeight is the full row height in DIPs, including custom ItemContent.
	// Positive finite values override the font-aware default (at least 32); group labels/separators
	// keep their own metrics. Match the height to the content's native layout.
	ItemHeight  float32
	FontSize    float32
	NoAutoFocus bool
	// KeepOpenOnSelect prevents closing for every row. By default actions close
	// and checkbox/radio rows stay open; MenuItem.CloseOnSelect overrides a row.
	KeepOpenOnSelect bool
	OnChoose         func(string)
	OnClose          func()
	// OnSwitch receives -1/1 for a logical previous/next top-level menu when
	// a horizontal key has no submenu action. Menubars use it across windows.
	OnSwitch func(delta int)
	// OnTab replaces the default Tab close path, allowing a compound menubar
	// to restore its parent focus and then select an explicit adjacent target.
	OnTab       func(backward bool)
	ItemContent func(*ui.Context, overlay.MenuItem)
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
	closeOnSelect := item.Checked == nil && !item.Radio
	if item.CloseOnSelect != nil {
		closeOnSelect = *item.CloseOnSelect
	}
	if closeOnSelect && !p.KeepOpenOnSelect && p.OnClose != nil {
		p.OnClose()
	}
}

func menuInput(state *MenuState, items []overlay.MenuItem, depth int, p MenuContentProps, rtl bool, ev ui.InputEvent) bool {
	levels := menuLevels(state, items)
	depth = max(0, min(depth, len(levels)-1))
	list := levels[depth]
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
	if ev.Mods != 0 && !(ev.Key == ui.KeyTab && ev.Mods == ui.Shift) {
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
		if current >= 0 && menuEnabled(list[current]) && len(list[current].Children) > 0 {
			menuOpenChild(state, items, depth, true)
		} else if p.OnSwitch != nil {
			p.OnSwitch(1)
		}
		return true
	case closeKey:
		if depth > 0 {
			state.ActiveDepth = depth
			menuCloseChild(state)
			return true
		}
		if p.OnSwitch != nil {
			p.OnSwitch(-1)
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
		if p.OnTab != nil {
			p.OnTab(ev.Mods&ui.Shift != 0)
		} else if p.OnClose != nil {
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
	if !menuDimensionsFinite(p.Width, p.MaxHeight, items) {
		panic("desktop: menu dimensions must be finite")
	}
	menu := theme.MenuOf(c)
	t := &menu.Tokens
	levels := menuLevels(state, items)
	width := p.Width
	if width <= 0 {
		width = 192
	}
	maxHeight := p.MaxHeight
	if maxHeight <= 0 {
		maxHeight = 480
	}
	fontSize := p.fontSize
	if fontSize <= 0 {
		fontSize = theme.ControlFontSize(c, 14, p.FontSize)
	}
	itemHeight := themedMenuItemHeight(p.ItemHeight, fontSize)
	groupHeight := menuGroupHeight(fontSize)
	viewportHeight := menuViewportHeight(maxHeight)
	label := p.Label
	if label == "" {
		label = "Menu"
	}
	// Only the individual panels paint a surface. This theme/measurement
	// wrapper stays transparent while those panels animate independently.
	root := ui.Themed(c, menu.Theme()).Background(ui.Transparent).Row().Gap(4).AlignItems(ui.Start).Role(ui.RoleGroup).Label(label)
	x, y, over := root.PointerPosition()
	moved := over && (!state.pointerOver || state.pointerX != x || state.pointerY != y)
	state.pointerOver, state.pointerX, state.pointerY = over, x, y
	rtl := c.Direction() == ui.RightToLeft
	frame := menuFrameOf(c)
	panels := ui.Local(root, menuPanelsKey{}, func() menuPanels { return menuPanels{} })
	if panels.anchors == nil {
		panels.anchors = make(map[string]ui.Rect)
	}
	seenAnchors := make(map[string]bool)
	panels.sync(state, len(levels), frame.closing)
	subAttachment := menuAttachment{side: menuSideRight, x: .5, y: .5}
	if rtl {
		subAttachment.side, subAttachment.x = menuSideLeft, 1
	}
	// Resolve retained paths against today's items. A removed/disabled ancestor
	// cancels that panel immediately; no old item or callback is kept alive.
	kept := panels.panels[:0]
	maxDepth := 0
	for _, presence := range panels.panels {
		list := menuAtPath(items, presence.path, len(presence.path))
		valid := len(list) > 0
		for depth, key := range presence.path {
			ancestors := menuAtPath(items, presence.path, depth)
			index := menuItemIndex(ancestors, key)
			if index < 0 || !menuEnabled(ancestors[index]) {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		presence.pose = presence.tween.update(c, presence.live, subAttachment)
		if !presence.live && presence.pose.opacity <= 0 {
			continue
		}
		kept = append(kept, presence)
		maxDepth = max(maxDepth, len(presence.path))
	}
	clear(panels.panels[len(kept):])
	panels.panels = kept
	// Retiring siblings overlap in one column. Reserve their maximum envelope,
	// while every live panel retains its current owner's actual width.
	depthWidths := make([]float32, maxDepth+1)
	depthWidths[0] = width
	for _, presence := range panels.panels {
		panelWidth := menuPathWidth(items, presence.path, width)
		if !presence.live {
			panelWidth = max(panelWidth, presence.bounds.W)
		}
		depth := len(presence.path)
		depthWidths[depth] = max(depthWidths[depth], panelWidth)
	}
	root.Children(func() {
		rootPanelBounds := ui.Rect{}
		renderPanel := func(depth int, list []overlay.MenuItem, path []string, presence *menuPanelPresence, live bool, highlight string) {
			panelLabel := label
			if depth > 0 {
				panelLabel = listLabel(items, path, depth)
			}
			panelWidth := menuPathWidth(items, path, width)
			role := ui.RoleMenu
			if p.selectList {
				role = ui.RoleListBox
			}
			panel := ui.Column(c.Key(menuPanelKey(menuPathKey(path)))).Width(panelWidth).Padding(4).Radius(t.RadiusMD()).Background(t.Popover).TextColor(t.PopoverForeground).Border(1, t.Border).Role(role).Label(panelLabel).InputMethod(live).TabStop(live && depth == state.ActiveDepth)
			menu.Surface(c, panel)
			if live {
				panel.Focusable()
				panel.HandleInput(func(ev ui.InputEvent) bool { return menuInput(state, items, depth, p, rtl, ev) })
			} else {
				panel.Disabled(true).DimDisabled(false).IgnoreInput(true).AccessHidden(true)
			}
			if depth == 0 && frame.active {
				attachment := frame.attachment
				if frame.geometry.known {
					geometry := frame.geometry
					bounds := panel.LayoutBounds()
					if bounds.W > 0 && bounds.H > 0 {
						geometry.content.X += bounds.X * c.Zoom()
						geometry.content.Y += bounds.Y * c.Zoom()
						geometry.content.W, geometry.content.H = bounds.W*c.Zoom(), bounds.H*c.Zoom()
					} else {
						geometry.content.W = width * c.Zoom()
						geometry.content.H = menuPanelHeight(list, itemHeight, maxHeight, groupHeight) * c.Zoom()
						if rtl {
							geometry.content.X += frame.geometry.content.W - geometry.content.W
						}
					}
					padding := float32(5)
					if geometry.context {
						padding = 0
					}
					attachment = menuAttachmentForSide(geometry.anchor, geometry.content, frame.attachment.side, padding)
				}
				menuApplyPose(panel, frame.pose, attachment)
			}
			if depth == 0 {
				rootPanelBounds = panel.LayoutBounds()
			}
			if presence != nil {
				// A retiring wider ancestor may reserve more space than this path
				// needs. Correct that space along logical start, including RTL.
				offset := float32(0)
				for level := 0; level < depth; level++ {
					offset += menuPathWidth(items, path[:level], width) - depthWidths[level]
				}
				panel.Absolute().Start(offset).Top(0)
				attachment := subAttachment
				content := presence.bounds
				if content.W <= 0 {
					content = ui.Rect{X: rootPanelBounds.X, Y: rootPanelBounds.Y, W: panelWidth, H: menuPanelHeight(list, itemHeight, maxHeight, groupHeight)}
					if rtl {
						for level := 1; level <= depth; level++ {
							content.X -= menuPathWidth(items, path[:level], width) + 4
						}
					} else {
						for level := 0; level < depth; level++ {
							content.X += menuPathWidth(items, path[:level], width) + 4
						}
					}
				}
				content.W = panelWidth
				if anchor, ok := panels.anchors[menuPathKey(path)]; ok {
					attachment = menuAttachmentForSide(anchor, content, subAttachment.side, 5)
				}
				menuApplyPose(panel, presence.pose, attachment)
				if bounds := panel.LayoutBounds(); bounds.W > 0 && bounds.H > 0 {
					if !presence.live {
						// Keep the full retiring envelope even if its owner is
						// resized again before the exit animation completes.
						bounds.W = max(bounds.W, presence.bounds.W)
					}
					presence.bounds = bounds
				}
			}
			if live && !p.NoAutoFocus && depth == 0 {
				panel.AutoFocus()
			}
			if live && !p.NoAutoFocus && depth == state.ActiveDepth && !root.FocusWithin() {
				panel.Focus()
			}
			if live && state.focus && depth == state.ActiveDepth {
				panel.Focus()
			}
			panel.Children(func() {
				height := menuContentHeight(list, itemHeight, groupHeight)
				body := ui.Column(c).Width(max(0, panelWidth-8))
				if height > viewportHeight {
					body = ui.Scroll(c).Width(max(0, panelWidth-8)).Height(viewportHeight)
				}
				body.Children(func() {
					group := ""
					for i, item := range list {
						key := menuItemKey(item, i)
						if item.Separator {
							ui.Box(c.Key(key)).Height(1).Margin(4, 0).Background(menu.Separator()).Role(ui.RoleSeparator)
							continue
						}
						if item.Group != "" && item.Group != group {
							group = item.Group
							caption := ui.Text(c, group).Padding(6, 8).FontSize(fontSize * 12 / 14).LineHeight(4.0 / 3).FontWeight(500).TextColor(t.MutedForeground)
							if item.GroupInset {
								caption.PaddingStart(32)
							}
						}
						role := ui.RoleMenuItem
						if item.Checked != nil {
							role = ui.RoleMenuItemCheckBox
						} else if item.Radio {
							role = ui.RoleMenuItemRadio
						}
						if p.selectList {
							role = ui.RoleOption
						}
						row := ui.Row(c.Key(key)).FillWidth().Height(itemHeight).PaddingX(8).Gap(8).Radius(max(2, t.Radius-2)).AlignItems(ui.Center).Role(role).Label(item.Label).Disabled(item.Disabled).FontSize(fontSize).LineHeight(20.0 / 14).TextColor(t.PopoverForeground)
						if item.Inset {
							row.PaddingStart(32)
						}
						if item.Checked != nil || item.Radio {
							row.PaddingEnd(32)
						}
						if menu.Animated() {
							motion.ApplyTransition(c, row)
						}
						highlighted := key == highlight
						if highlighted {
							bg, fg := menu.Highlight(item.Destructive)
							row.Background(bg).TextColor(fg)
							if live && depth == state.ActiveDepth {
								panel.ActiveDescendant(row)
								if state.reveal {
									row.ScrollIntoView()
								}
							}
						}
						if p.selectList {
							row.Selected(item.Selected)
						} else if item.Checked != nil {
							row.Checked(*item.Checked)
						} else if item.Radio {
							row.Checked(item.Selected)
						}
						if len(item.Children) > 0 {
							anchorKey := menuPathKey(append(slices.Clone(path), key))
							seenAnchors[anchorKey] = true
							if bounds := row.LayoutBounds(); bounds.W > 0 && bounds.H > 0 {
								panels.anchors[anchorKey] = bounds
							}
							row.HasPopup(ui.PopupMenu).Expanded(depth < len(state.Path) && state.Path[depth] == key)
						}
						if item.Destructive {
							row.TextColor(t.Destructive)
							if menu.Translucent() {
								row.TextColor(t.AccentForeground)
							}
							if highlighted {
								bg, fg := menu.Highlight(true)
								row.Background(bg).TextColor(fg)
							}
						}
						if item.Disabled {
							row.DimDisabled(false).Opacity(.5).Cursor(ui.CursorNotAllowed)
						}
						row.Children(func() {
							if item.Icon != nil {
								ui.Icon(c, item.Icon).Size(16, 16).AccessHidden(true)
							}
							if (item.Checked != nil && *item.Checked) || (item.Radio && item.Selected) {
								ui.Box(c).Absolute().End(8).TopPercent(50).Size(16, 16).Translate(0, -8).Children(func() { indicator.Check(c) })
							}
							if p.ItemContent != nil {
								ui.Column(c).Grow(1).MinWidth(0).Children(func() { p.ItemContent(c, item) })
							} else {
								ui.Text(c, item.Label).Grow(1).MinWidth(0).SingleLine()
							}
							if item.Shortcut != "" {
								shortcutColor := t.MutedForeground
								if item.Destructive && menu.Translucent() {
									shortcutColor = t.AccentForeground
								}
								ui.Text(c, item.Shortcut).FontSize(fontSize * 12 / 14).TextColor(shortcutColor)
							}
							if len(item.Children) > 0 {
								angle := float32(-90)
								if rtl {
									angle = 90
								}
								indicator.ChevronDown(c).Rotate(angle)
							}
						})
						hovered := row.Hovered()
						if live && moved && hovered && !item.Disabled {
							menuHighlight(state, depth, key)
							if len(item.Children) > 0 {
								menuOpenChild(state, items, depth, false)
							} else if len(state.Path) > depth {
								state.Path = state.Path[:depth]
							}
							c.Invalidate()
						}
						if live && row.Clicked() && !item.Disabled {
							menuHighlight(state, depth, key)
							panel.Focus()
							menuActivate(state, items, depth, p)
							c.Invalidate()
						}
					}
				})
			})
		}
		highlight := ""
		if len(state.Highlighted) > 0 {
			highlight = state.Highlighted[0]
		}
		renderPanel(0, levels[0], nil, nil, !frame.closing, highlight)
		for depth := 1; depth <= maxDepth; depth++ {
			height := float32(0)
			for _, presence := range panels.panels {
				if len(presence.path) == depth {
					list := menuAtPath(items, presence.path, depth)
					height = max(height, menuPanelHeight(list, itemHeight, maxHeight, groupHeight))
				}
			}
			ui.Box(c.Key(menuDepthKey(depth))).Width(depthWidths[depth]).Height(height).Children(func() {
				for _, presence := range panels.panels {
					if len(presence.path) == depth {
						// Same-depth replacements overlap while the old panel leaves;
						// they never widen the Flyout by another column.
						renderPanel(depth, menuAtPath(items, presence.path, depth), presence.path, presence, presence.live, presence.highlight)
					}
				}
			})
		}
	})
	state.focus, state.reveal = false, false
	for key := range panels.anchors {
		if !seenAnchors[key] {
			delete(panels.anchors, key)
		}
	}
	return root
}

// Each owner inherits the original shared width, rather than the nearest
// ancestor override. Resolve current data on every independent child build.
func menuPathWidth(items []overlay.MenuItem, path []string, fallback float32) float32 {
	width := fallback
	for _, key := range path {
		index := menuItemIndex(items, key)
		if index < 0 {
			break
		}
		item := items[index]
		width = fallback
		if item.SubmenuWidth > 0 {
			width = item.SubmenuWidth
		}
		items = item.Children
	}
	return width
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
func themedMenuItemHeight(value, fontSize float32) float32 {
	if value > 0 && !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0) {
		return value
	}
	return max(float32(32), fontSize*(20.0/14)+12)
}

func menuGroupHeight(fontSize float32) float32 {
	return fontSize*(12.0/14)*(4.0/3) + 12
}

// A panel adds four DIP padding and a one DIP border on each vertical edge.
func menuViewportHeight(maxHeight float32) float32 { return max(1, maxHeight-10) }

func menuPanelHeight(items []overlay.MenuItem, itemHeight, maxHeight float32, groupHeight ...float32) float32 {
	return min(menuContentHeight(items, itemHeight, groupHeight...), menuViewportHeight(maxHeight)) + 10
}

func menuContentHeight(items []overlay.MenuItem, itemHeight float32, groupHeight ...float32) float32 {
	height := float32(0)
	captionHeight := float32(28)
	if len(groupHeight) > 0 {
		captionHeight = groupHeight[0]
	}
	group := ""
	for _, item := range items {
		if item.Separator {
			height += 9
			continue
		}
		if item.Group != "" && item.Group != group {
			group = item.Group
			height += captionHeight
		}
		height += itemHeight
	}
	return height
}
