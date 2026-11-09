package desktop

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// MenubarState belongs to one menubar in one parent window. Active is a stable
// top-level item key; Open is controlled and is also cleared by Host cleanup.
type MenubarState struct {
	Open   bool
	Active string

	menu                   MenuState
	chosen                 string
	lastActive, focused    string
	wasOpen                bool
	contentStamp, revision uint64
	triggers               map[string]*ui.Handle
	pointerOver            bool
	pointerX, pointerY     float32
	pressedTrigger         string
	pressCloses            bool
}

type MenubarProps struct {
	Items            []overlay.MenubarItem
	Label            string
	Width, MaxHeight float32
	Placement        mygo.Placement
	Gap              int
	// Structural menu changes refresh automatically. Revision covers data used
	// by ItemContent or callbacks without a structural item change.
	Revision         uint64
	KeepOpenOnSelect bool
	ItemContent      func(*ui.Context, overlay.MenuItem)
	OnChoose         func(string)
	OnDismiss        func()
	// OnTab runs after dismissing and restoring the parent trigger. Applications
	// can focus an adjacent parent control with a stable Handle; MyGo has no
	// cross-window Tab-forwarding API. Without it, the next Tab leaves the bar.
	OnTab func(backward bool)
}

func menubarKey(item overlay.MenubarItem, index int) string {
	if item.Key != "" {
		return item.Key
	}
	return fmt.Sprintf("@%d", index)
}

func menubarIndex(items []overlay.MenubarItem, key string) int {
	for i, item := range items {
		if menubarKey(item, i) == key {
			return i
		}
	}
	return -1
}

func menubarNext(items []overlay.MenubarItem, from, delta int) string {
	if len(items) == 0 {
		return ""
	}
	for step := 1; step <= len(items); step++ {
		index := ((from+delta*step)%len(items) + len(items)) % len(items)
		if !items[index].Disabled {
			return menubarKey(items[index], index)
		}
	}
	return ""
}

func (s *MenubarState) restore() {
	if handle := s.triggers[s.Active]; handle != nil {
		handle.Focus()
	}
}

// Menubar is a horizontal group of shadcn triggers with one independently
// hosted Flyout. Keyboard and hover sibling switching reanchor that same native
// surface. Submenu arrows take precedence over top-level switching.
// OnChoose runs in the child event; the return reports the result once on the
// next parent build. Applications should dispatch using one of these paths.
func Menubar(c *ui.Context, h *Host, key string, s *MenubarState, p MenubarProps) (string, error) {
	if c == nil || !c.Valid() || key == "" || s == nil {
		return "", errors.New("desktop: Menubar needs a live parent context, stable key and caller-owned state")
	}
	result := s.chosen
	s.chosen = ""
	keys := make(map[string]bool, len(p.Items))
	for i, item := range p.Items {
		id := menubarKey(item, i)
		if keys[id] {
			s.Open, s.wasOpen = false, false
			if h != nil {
				h.Close(key)
			}
			return result, fmt.Errorf("desktop: Menubar top-level key %q is duplicated", id)
		}
		keys[id] = true
	}
	if s.triggers == nil {
		s.triggers = make(map[string]*ui.Handle)
	}
	for id := range s.triggers {
		if !keys[id] {
			delete(s.triggers, id)
		}
	}
	for id := range keys {
		if s.triggers[id] == nil {
			s.triggers[id] = &ui.Handle{}
		}
	}
	index := menubarIndex(p.Items, s.Active)
	if index < 0 || p.Items[index].Disabled {
		if s.Active != "" {
			s.Open = false
		}
		s.Active = menubarNext(p.Items, -1, 1)
		if s.Active == "" {
			s.Open = false
		}
		if s.wasOpen && s.Active != "" {
			s.restore()
		}
	}
	label := p.Label
	if label == "" {
		label = "Menu bar"
	}
	t := theme.TokensOf(c)
	root := ui.Row(c.Key("menubar:"+key)).Height(36).Gap(4).Padding(4).Radius(t.RadiusMD()).
		Border(1, t.Border).Background(t.Background).Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, .05)).
		Role(ui.RoleMenuBar).Label(label).FocusGroup(ui.Horizontal)
	x, y, over := root.PointerPosition()
	moved := over && (!s.pointerOver || s.pointerX != x || s.pointerY != y)
	s.pointerOver, s.pointerX, s.pointerY = over, x, y
	anchors := make(map[string]ui.Element, len(p.Items))
	fromUp := false
	// A pointer press can focus a sibling before Clicked is observed. Compare
	// against the entry state, not the focus handler's newly active sibling.
	clickedActive, clickedOpen := s.Active, s.Open
	anyPressed := false
	root.Children(func() {
		for i, item := range p.Items {
			id := menubarKey(item, i)
			trigger := motion.ApplyTransition(c, ui.ButtonBase(c.Key(id)).Role(ui.RoleMenuItem).Label(item.Label).Disabled(item.Disabled).
				Height(28).PaddingX(8).Radius(t.RadiusSM()).FontSize(14).FontWeight(500).
				TextColor(t.Foreground).HasPopup(ui.PopupMenu).Expanded(s.Open && s.Active == id).
				Cursor(ui.CursorPointer)).Bind(s.triggers[id])
			anchors[id] = trigger
			if trigger.Pressed() {
				anyPressed = true
				if s.pressedTrigger == "" {
					s.pressedTrigger = id
					s.pressCloses = clickedOpen && clickedActive == id
				}
			}
			if trigger.Focused() {
				if s.Open && s.focused != id {
					s.Active = id
				}
				s.focused = id
			}
			if !item.Disabled {
				if trigger.Clicked() {
					close := clickedOpen && clickedActive == id
					if s.pressedTrigger == id {
						close = s.pressCloses
					}
					s.pressedTrigger = ""
					if close {
						s.Open = false
					} else {
						s.Active, s.Open = id, true
					}
					s.focused = id
				}
				down, up := trigger.Shortcut(0, ui.KeyDown), trigger.Shortcut(0, ui.KeyUp)
				if down || up {
					s.Active, s.Open, s.focused = id, true, id
					fromUp = up
				}
				if s.Open && moved && trigger.Hovered() {
					s.Active = id
				}
			}
			if s.Open && s.Active == id || trigger.Hovered() && !item.Disabled {
				trigger.Background(t.Muted).TextColor(t.Foreground)
			}
			if item.Disabled {
				trigger.Opacity(.5).Cursor(ui.CursorNotAllowed)
			}
			trigger.Children(func() { ui.Text(c, item.Label).SingleLine() })
		}
	})
	if !anyPressed {
		s.pressedTrigger = ""
	}
	if s.Open && (!s.wasOpen || s.lastActive != s.Active) {
		s.menu.Reset()
		s.revision++
		if fromUp {
			active := menubarIndex(p.Items, s.Active)
			if active >= 0 {
				for i := len(p.Items[active].Items) - 1; i >= 0; i-- {
					if menuEnabled(p.Items[active].Items[i]) {
						s.menu.Highlighted = []string{menuItemKey(p.Items[active].Items[i], i)}
						break
					}
				}
			}
		}
	}
	if stamp := menubarStamp(p.Items, p.Revision); stamp != s.contentStamp {
		s.contentStamp = stamp
		s.revision++
	}
	s.lastActive, s.wasOpen = s.Active, s.Open
	if h == nil {
		if s.Open {
			s.Open, s.wasOpen = false, false
			return result, errors.New("desktop: opening a Menubar requires a parent-bound Host")
		}
		return result, nil
	}
	anchor := anchors[s.Active]
	if !anchor.Valid() {
		h.Close(key)
		s.Open, s.wasOpen = false, false
		return result, nil
	}
	width, maxHeight := p.Width, p.MaxHeight
	if math.IsNaN(float64(width)) || math.IsInf(float64(width), 0) || math.IsNaN(float64(maxHeight)) || math.IsInf(float64(maxHeight), 0) {
		s.Open, s.wasOpen = false, false
		h.Close(key)
		return result, errors.New("desktop: Menubar dimensions must be finite")
	}
	if width <= 0 {
		width = 144
	}
	if maxHeight <= 0 {
		maxHeight = 480
	}
	active := menubarIndex(p.Items, s.Active)
	if active < 0 {
		return result, nil
	}
	item := p.Items[active]
	closeMenu := func(dismiss bool) {
		if !s.Open {
			return
		}
		s.Open, s.wasOpen = false, false
		h.Close(key)
		s.restore()
		if dismiss && p.OnDismiss != nil {
			p.OnDismiss()
		}
		h.backend.InvalidateParent()
	}
	zero := float32(0)
	gap := p.Gap
	if gap == 0 {
		gap = 8
	}
	popup := PopoverProps{menuMotion: menuMotionBar, menuIdentity: s.Active, Label: item.Label, Placement: p.Placement, Gap: gap,
		Width: int(math.Ceil(float64(width))), Height: 200, FitContent: true,
		MaxHeight: int(math.Ceil(float64(maxHeight))), ManualTrigger: true,
		NoFocusRestore: true, ContentPadding: &zero, ContentRole: ui.RoleGroup,
		Revision: s.revision, OnDismiss: func() {
			s.Open, s.wasOpen = false, false
			s.restore()
			if p.OnDismiss != nil {
				p.OnDismiss()
			}
		}}
	err := Popover(c, h, key, anchor, &s.Open, popup, func(child *ui.Context) {
		// Child key events can precede the parent's reanchoring frame. Read the
		// current sibling here so the next key never acts on the old menu.
		current := menubarIndex(p.Items, s.Active)
		if current < 0 || p.Items[current].Disabled {
			return
		}
		currentItem := p.Items[current]
		selectedInEvent := false
		MenuContent(child.Key("menubar-menu:"+s.Active), &s.menu, currentItem.Items, MenuContentProps{
			Label: currentItem.Label, Width: width, MaxHeight: maxHeight,
			KeepOpenOnSelect: p.KeepOpenOnSelect, ItemContent: p.ItemContent,
			OnChoose: func(value string) {
				selectedInEvent = true
				s.chosen = value
				if p.OnChoose != nil {
					p.OnChoose(value)
				}
				h.backend.InvalidateParent()
			},
			OnClose: func() { closeMenu(!selectedInEvent) },
			OnTab: func(backward bool) {
				closeMenu(true)
				if p.OnTab != nil {
					// Handle requests resolve in parent tree order. Cancel our own
					// restoration before an app explicitly chooses an earlier node.
					if trigger := s.triggers[s.Active]; trigger != nil {
						trigger.CancelFocus()
					}
					p.OnTab(backward)
				}
			},
			OnSwitch: func(delta int) {
				if !s.Open {
					return
				}
				if next := menubarNext(p.Items, menubarIndex(p.Items, s.Active), delta); next != "" && next != s.Active {
					s.Active, s.focused = next, next
					// The child can process several keys before the parent builds.
					// Enter the new root now, alongside its latest native content;
					// the parent will separately commit the new window anchor.
					if entry := h.entries[key]; entry != nil {
						entry.props.menuIdentity = next
					}
					s.menu.Reset()
					s.lastActive = next
					s.revision++
					s.restore()
					h.backend.InvalidateParent()
				}
			},
		})
	})
	// Host marks generic dialog triggers; a menubar has menu semantics.
	for id, trigger := range anchors {
		trigger.HasPopup(ui.PopupMenu).Expanded(s.Open && s.Active == id)
	}
	return result, err
}

// NativeMenubar groups system-menu triggers. The OS owns their popup contents;
// this is not the application's top-level OS menu bar. MyGo's popup adapter
// does not expose sibling switching while the OS tracks a popup menu.
func NativeMenubar(c *ui.Context, items []overlay.MenubarItem) (string, error) {
	if c == nil || !c.Valid() {
		return "", errors.New("desktop: NativeMenubar needs a live parent context")
	}
	keys := make(map[string]bool, len(items))
	for i, item := range items {
		key := menubarKey(item, i)
		if keys[key] {
			return "", fmt.Errorf("desktop: NativeMenubar top-level key %q is duplicated", key)
		}
		keys[key] = true
	}
	policy := c.MenuPresentation()
	c.SetMenuPresentation(ui.SystemMenus)
	defer c.SetMenuPresentation(policy)
	chosen := ""
	ui.Row(c).Gap(2).Role(ui.RoleMenuBar).Label("Menu bar").FocusGroup(ui.Horizontal).Children(func() {
		for i, item := range items {
			value := overlay.NativeDropdownMenu(c.Key(menubarKey(item, i)), overlay.MenuProps{Items: item.Items}, func(trigger ui.Element) {
				trigger.Disabled(item.Disabled).Height(28).PaddingX(8).Radius(c.Theme().Radius).Label(item.Label).
					Children(func() { ui.Text(c, item.Label).SingleLine() })
			})
			if value != "" {
				chosen = value
			}
		}
	})
	return chosen, nil
}

func menubarStamp(items []overlay.MenubarItem, revision uint64) uint64 {
	hash := fnv.New64a()
	fmt.Fprint(hash, revision, ";")
	var visit func([]overlay.MenuItem)
	visit = func(items []overlay.MenuItem) {
		fmt.Fprint(hash, "[")
		for _, item := range items {
			checked := "none"
			if item.Checked != nil {
				checked = fmt.Sprint(*item.Checked)
			}
			fmt.Fprintf(hash, "%q/%q/%q/%q/%t/%t/%t/%t/%t/%s/%p;", item.Key, item.Label, item.Shortcut, item.Group,
				item.Disabled, item.Destructive, item.Separator, item.Radio, item.Selected, checked, item.Icon)
			visit(item.Children)
		}
		fmt.Fprint(hash, "]")
	}
	for _, item := range items {
		fmt.Fprintf(hash, "%q/%q/%t;", item.Key, item.Label, item.Disabled)
		visit(item.Items)
	}
	return hash.Sum64()
}
