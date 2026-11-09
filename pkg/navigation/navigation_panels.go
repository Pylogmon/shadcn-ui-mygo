package navigation

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// NavigationMenuItem is a route trigger or independently hosted content panel.
// Content receives the CHILD Context and a route action. Its local controls
// survive sibling switches. Use Revision when external closure data changes.
// Zero Height measures content; zero Width inherits NavigationMenuProps.Width.
type NavigationMenuItem struct {
	Key, Label, Description string
	Icon                    *ui.SVG
	Disabled                bool
	Links                   []NavItem
	Columns                 int
	Width, Height           float32
	Content                 func(child *ui.Context, choose func(string))
}

type NavigationMenuProps struct {
	Items                 []NavigationMenuItem
	Label                 string
	Width, MaxHeight      float32
	Gap                   int
	OpenDelay, CloseDelay time.Duration
	// Duration defaults to the upstream 350ms viewport/content transition.
	// Negative durations and reduced motion apply changes immediately.
	Duration time.Duration
	NoLoop   bool
	Revision uint64
	OnSelect func(string)
	// OnTab optionally focuses a neighboring real parent control when Tab
	// crosses a panel boundary. Without it, focus returns to the trigger.
	// Custom content retains native Tab traversal among its own controls.
	OnTab func(backward bool)
}

type navigationPanelState struct {
	active                                                                  string
	open, keyboard, enter, enterLast, changed, focusWindow, keyboardPending bool
	suppressed, hoverCandidate                                              string
	enteredAt, leftAt                                                       time.Time
	childHover, childFocus                                                  bool
	handles                                                                 map[string]*ui.Handle
	heights                                                                 map[string]float32
	mounted                                                                 map[string]bool
	bounds                                                                  map[string]ui.Rect
	revision                                                                uint64
	options, selected                                                       string
	direction                                                               int
	sibling                                                                 int
	focus                                                                   string
	parent                                                                  ui.Services
	tabbed, tabBack                                                         bool
	pointerMoved, pointerLocked                                             bool
	pointerHover                                                            string
}

type navigationIndicatorKey struct{ key string }

func navigationKey(items []NavigationMenuItem, key string) int {
	for i := range items {
		if items[i].Key == key {
			return i
		}
	}
	return -1
}
func navigationHasPanel(item NavigationMenuItem) bool {
	return item.Content != nil || len(item.Links) > 0
}
func navigationNext(items []NavigationMenuItem, index, direction int, loop bool) int {
	for step := 1; step <= len(items); step++ {
		at := index + direction*step
		if !loop && (at < 0 || at >= len(items)) {
			return index
		}
		at = (at%len(items) + len(items)) % len(items)
		if !items[at].Disabled {
			return at
		}
	}
	return -1
}
func navigationRoute(items []NavItem, route string) bool {
	for _, it := range items {
		if it.Key == route || navigationRoute(it.Children, route) {
			return true
		}
	}
	return false
}
func navigationOptions(p NavigationMenuProps) string {
	var b strings.Builder
	var links func([]NavItem)
	links = func(items []NavItem) {
		for _, it := range items {
			fmt.Fprintf(&b, "%q/%q/%q/%q/%t/%p;", it.Key, it.Label, it.Description, it.Group, it.Disabled, it.Icon)
			links(it.Children)
		}
	}
	for _, it := range p.Items {
		fmt.Fprintf(&b, "%q/%q/%q/%t/%p/%v/%v/%d;", it.Key, it.Label, it.Description, it.Disabled, it.Icon, it.Width, it.Height, it.Columns)
		fmt.Fprintf(&b, "%t/%p;", it.Content != nil, it.Content)
		links(it.Links)
	}
	fmt.Fprintf(&b, "revision:%d", p.Revision)
	return b.String()
}
func (s *navigationPanelState) close(suppress, restore bool) {
	if suppress {
		s.suppressed = s.active
	}
	if restore {
		if h := s.handles[s.active]; h != nil {
			h.Focus()
		}
	}
	s.open, s.keyboard, s.enter, s.focusWindow, s.keyboardPending = false, false, false, false, false
	s.childHover, s.childFocus = false, false
	s.enteredAt, s.leftAt = time.Time{}, time.Time{}
	s.parent.Invalidate()
}
func (s *navigationPanelState) activate(items []NavigationMenuItem, index int, keyboard bool) {
	if index < 0 || index >= len(items) || items[index].Disabled || !navigationHasPanel(items[index]) {
		return
	}
	if s.active != items[index].Key {
		previous := navigationKey(items, s.active)
		s.direction = 1
		if previous > index {
			s.direction = -1
		}
		s.active = items[index].Key
		s.revision++
	}
	s.open, s.keyboard = true, keyboard
	s.mounted[items[index].Key] = true
	s.enteredAt, s.leftAt = time.Time{}, time.Time{}
	if keyboard {
		s.enter = true
		s.focusWindow = true
		s.keyboardPending = true
		s.pointerLocked = true
	}
}

// NavigationMenuWith renders shadcn route links and content in a native Flyout
// which may extend beyond the parent. Host lifecycle brackets belong to the
// parent view. selected is a caller-owned route, updated once by a chosen link;
// Changed is reported by the returned root on the next parent frame.
func NavigationMenuWith(c *ui.Context, host *desktop.Host, key string, selected *string, p NavigationMenuProps) (ui.Element, error) {
	if c == nil || !c.Valid() || key == "" || selected == nil {
		return ui.Element{}, errors.New("navigation: NavigationMenuWith needs a live context, stable key and route pointer")
	}
	t := theme.TokensOf(c)
	label := p.Label
	if label == "" {
		label = "Main navigation"
	}
	root := ui.Row(c.Key("navigation-menu:" + key)).Gap(4).AlignItems(ui.Center).Role(ui.RoleNavigation).Label(label).FocusGroup(ui.Horizontal)
	s := ui.Local(root, "panels", func() navigationPanelState {
		return navigationPanelState{handles: map[string]*ui.Handle{}, heights: map[string]float32{}, bounds: map[string]ui.Rect{}, mounted: map[string]bool{}}
	})
	s.parent = c.Services()
	root.ObserveInput(func(ev ui.InputEvent) {
		if ev.Kind == ui.InputPointerMove {
			s.pointerMoved = true
		} else if ev.Kind == ui.InputKeyDown {
			s.pointerLocked = true
		}
	})
	if s.changed {
		root.SetChanged()
		s.changed = false
	}
	options := navigationOptions(p)
	if options != s.options || *selected != s.selected {
		s.options, s.selected = options, *selected
		s.revision++
	}
	seen := map[string]bool{}
	for _, it := range p.Items {
		if it.Key == "" || seen[it.Key] {
			return root, errors.New("navigation: panel items need unique nonempty keys")
		}
		seen[it.Key] = true
		if _, ok := s.handles[it.Key]; !ok {
			s.handles[it.Key] = &ui.Handle{}
		}
		if math.IsNaN(float64(it.Width)) || math.IsInf(float64(it.Width), 0) || math.IsNaN(float64(it.Height)) || math.IsInf(float64(it.Height), 0) {
			return root, errors.New("navigation: panel sizes must be finite")
		}
	}
	for k := range s.handles {
		if !seen[k] {
			delete(s.handles, k)
			delete(s.heights, k)
			delete(s.bounds, k)
			delete(s.mounted, k)
		}
	}
	active := navigationKey(p.Items, s.active)
	if s.open && (active < 0 || p.Items[active].Disabled || !navigationHasPanel(p.Items[active])) {
		s.close(false, false)
	}
	if s.sibling != 0 && active >= 0 {
		index := navigationNext(p.Items, active, s.sibling, !p.NoLoop)
		s.sibling = 0
		if index >= 0 && navigationHasPanel(p.Items[index]) {
			s.activate(p.Items, index, true)
		} else if index >= 0 {
			s.close(false, false)
			s.focus = p.Items[index].Key
		}
	}
	choose := func(route string) {
		if route == "" {
			return
		}
		if *selected != route {
			*selected = route
			s.changed = true
		}
		if p.OnSelect != nil {
			p.OnSelect(route)
		}
		s.close(true, s.keyboard)
	}
	var triggers []ui.Element
	hover := -1
	focus := -1
	openKey := false
	root.Children(func() {
		for i, it := range p.Items {
			trigger := motion.ApplyTransition(c, ui.ButtonBase(c.Key(it.Key)).Bind(s.handles[it.Key]).Label(it.Label).Description(it.Description).Disabled(it.Disabled).
				Height(36).PaddingX(12).Gap(6).Radius(t.RadiusMD()).FontSize(14).FontWeight(500).TextColor(t.Foreground))
			trigger.Draw(func(_ *ui.Painter, bounds ui.Rect) {
				if s.bounds[it.Key] != bounds {
					s.bounds[it.Key] = bounds
					s.parent.Invalidate()
				}
			})
			triggers = append(triggers, trigger)
			if navigationHasPanel(it) {
				trigger.HasPopup(ui.PopupDialog).Expanded(s.open && s.active == it.Key)
				if navigationRoute(it.Links, *selected) {
					trigger.Current(ui.CurrentPage)
				}
			} else {
				trigger.Role(ui.RoleLink).Current(ui.CurrentNone)
				if it.Key == *selected {
					trigger.Current(ui.CurrentPage)
				}
			}
			if it.Disabled {
				trigger.Opacity(.5)
			}
			if trigger.Hovered() && !it.Disabled {
				hover = i
				trigger.Background(t.Accent).TextColor(t.AccentForeground)
			}
			back, forth := ui.KeyLeft, ui.KeyRight
			fromUp := trigger.Shortcut(0, ui.KeyUp)
			fromDown := trigger.Shortcut(0, ui.KeyDown)
			if c.Direction() == ui.RightToLeft {
				back, forth = forth, back
			}
			switch {
			case trigger.Shortcut(0, back):
				focus = navigationNext(p.Items, i, -1, !p.NoLoop)
			case trigger.Shortcut(0, forth):
				focus = navigationNext(p.Items, i, 1, !p.NoLoop)
			case trigger.Shortcut(0, ui.KeyHome):
				focus = navigationNext(p.Items, -1, 1, false)
			case trigger.Shortcut(0, ui.KeyEnd):
				focus = navigationNext(p.Items, len(p.Items), -1, false)
			case navigationHasPanel(it) && (fromDown || fromUp):
				s.enterLast = fromUp
				s.suppressed = ""
				s.activate(p.Items, i, true)
				openKey = true
			}
			if trigger.Clicked() && !it.Disabled {
				if navigationHasPanel(it) {
					if s.open && s.active == it.Key {
						s.close(true, false)
					} else {
						s.suppressed = ""
						s.activate(p.Items, i, true)
						openKey = true
					}
				} else {
					choose(it.Key)
				}
			}
			trigger.Children(func() {
				if it.Icon != nil {
					ui.Icon(c, it.Icon).Size(16, 16)
				}
				ui.Text(c, it.Label).SingleLine()
				if navigationHasPanel(it) {
					angle := float32(0)
					if s.open && s.active == it.Key {
						angle = 180
					}
					ui.Text(c, "⌄").AccessHidden(true).Rotate(navigationAnimate(trigger, "caret", angle, navigationDuration(c, p)))
				}
				if s.open && s.active == it.Key {
					transition := ui.ElementTransition{Duration: navigationDuration(c, p), Match: navigationIndicatorKey{key}, Ease: motion.EaseOut, Colors: true}
					ui.Box(c).Absolute().Bottom(-5).LeftPercent(50).MarginX(-4).Size(8, 8).Rotate(45).Background(t.Popover).Border(1, t.Border).IgnoreInput(true).AccessHidden(true).Transition(transition)
					trigger.Background(t.Accent).TextColor(t.AccentForeground)
				}
			})
		}
	})
	if focus >= 0 && focus < len(triggers) {
		triggers[focus].Focus()
		if s.open {
			if navigationHasPanel(p.Items[focus]) {
				s.activate(p.Items, focus, true)
			} else {
				s.close(false, false)
			}
		}
	}
	if s.focus != "" {
		if h := s.handles[s.focus]; h != nil {
			h.Focus()
		}
		s.focus = ""
	}
	active = navigationKey(p.Items, s.active)
	hoverKey := ""
	if hover >= 0 {
		hoverKey = p.Items[hover].Key
	}
	if s.pointerMoved || hoverKey != s.pointerHover {
		s.pointerLocked = false
	}
	allowHover := !s.pointerLocked
	s.pointerMoved, s.pointerHover = false, hoverKey
	if s.open && root.Shortcut(0, ui.KeyEscape) {
		s.close(true, s.keyboard)
	}
	if s.suppressed != "" && (hover < 0 || p.Items[hover].Key != s.suppressed) {
		s.suppressed = ""
	}
	appActive, appKnown := false, false
	if host != nil {
		appActive, appKnown = host.ApplicationState()
	}
	transferred := s.keyboardPending && appKnown && appActive
	inactive := c.WindowHidden() || appKnown && !appActive || !c.WindowActive() && !s.childFocus && !transferred
	if inactive {
		if s.open {
			s.close(false, false)
		}
	} else if !openKey {
		childHot := s.childHover || s.childFocus || transferred
		if allowHover && hover >= 0 && navigationHasPanel(p.Items[hover]) && p.Items[hover].Key != s.suppressed {
			candidate := p.Items[hover].Key
			if s.hoverCandidate != candidate {
				s.hoverCandidate = candidate
				s.enteredAt = c.Now()
			}
			delay := p.OpenDelay
			if delay == 0 {
				delay = 200 * time.Millisecond
			}
			if s.open && s.active != candidate {
				delay = 0
			}
			if !s.open || s.active != candidate {
				if delay < 0 || c.Now().Sub(s.enteredAt) >= delay {
					s.activate(p.Items, hover, false)
				} else {
					c.After(delay - c.Now().Sub(s.enteredAt))
				}
			}
			s.leftAt = time.Time{}
		} else {
			s.hoverCandidate = ""
			s.enteredAt = time.Time{}
			focusedTrigger := active >= 0 && triggers[active].Focused() && s.keyboard
			if s.open && !childHot && !focusedTrigger {
				delay := p.CloseDelay
				if delay == 0 {
					delay = 300 * time.Millisecond
				}
				if s.leftAt.IsZero() {
					s.leftAt = c.Now()
				}
				if delay < 0 || c.Now().Sub(s.leftAt) >= delay {
					s.close(false, false)
				} else {
					c.After(delay - c.Now().Sub(s.leftAt))
				}
			} else {
				s.leftAt = time.Time{}
			}
		}
	}
	active = navigationKey(p.Items, s.active)
	for i, tr := range triggers {
		if navigationHasPanel(p.Items[i]) {
			tr.Expanded(s.open && s.active == p.Items[i].Key)
		}
	}
	if active < 0 {
		if host != nil {
			host.Close(key)
		}
		return root, nil
	}
	if host == nil {
		if s.open {
			s.close(true, false)
			return root, errors.New("navigation: opening NavigationMenuWith requires a parent-bound Host")
		}
		return root, nil
	}
	width := navigationWidth(p, p.Items[active])
	maxHeight := p.MaxHeight
	if maxHeight <= 0 {
		maxHeight = 480
	}
	if math.IsNaN(float64(width)) || math.IsInf(float64(width), 0) || math.IsNaN(float64(maxHeight)) || math.IsInf(float64(maxHeight), 0) {
		s.close(true, false)
		return root, errors.New("navigation: viewport sizes must be finite")
	}
	height := p.Items[active].Height
	if height <= 0 {
		height = s.heights[s.active]
	}
	if height <= 0 {
		height = 180
	}
	height = max(40, min(maxHeight, height))
	duration := navigationDuration(c, p)
	w := navigationAnimate(root, "viewport-width", width, duration)
	hh := navigationAnimate(root, "viewport-height", height, duration)
	zero := float32(0)
	gap := p.Gap
	if gap == 0 {
		gap = 8
	}
	// The virtual anchor follows the selected trigger's actual committed box.
	// Moving it keeps the same native viewport while its position interpolates.
	bounds := s.bounds[p.Items[active].Key]
	x := navigationAnimate(root, "viewport-anchor-x", bounds.X, duration)
	y := navigationAnimate(root, "viewport-anchor-y", bounds.Y, duration)
	aw := navigationAnimate(root, "viewport-anchor-width", bounds.W, duration)
	ah := navigationAnimate(root, "viewport-anchor-height", bounds.H, duration)
	var anchor ui.Element
	ui.Overlay(c.Key("navigation-anchor:"+key), func() {
		anchor = ui.Box(c).Absolute().Left(x).Top(y).Width(aw).Height(ah).IgnoreInput(true).AccessHidden(true)
	})
	props := desktop.PopoverProps{Label: p.Items[active].Label, Width: max(1, int(math.Ceil(float64(w)))), Height: max(1, int(math.Ceil(float64(hh)))), Gap: gap,
		ManualTrigger: true, ShowInactive: true, NoFocusRestore: true, ContentPadding: &zero, ContentRole: ui.RoleGroup, Revision: s.revision,
		OnDismiss: func() { s.close(true, s.keyboard) }}
	err := desktop.Popover(c, host, key, anchor, &s.open, props, func(child *ui.Context) {
		navPanelView(child, host, key, s, p, choose)
	})
	if err != nil {
		s.close(true, false)
		return root, err
	}
	if s.focusWindow && host.Focus(key) {
		s.focusWindow = false
		host.Refresh(key)
	}
	return root, nil
}

func navigationDuration(c *ui.Context, p NavigationMenuProps) time.Duration {
	if motion.Reduced(c) {
		return -1
	}
	if p.Duration == 0 {
		return 350 * time.Millisecond
	}
	return p.Duration
}
func navigationWidth(p NavigationMenuProps, item NavigationMenuItem) float32 {
	width := item.Width
	if width <= 0 {
		width = p.Width
	}
	if width <= 0 {
		width = 480
	}
	return max(160, width)
}

func navPanelView(c *ui.Context, host *desktop.Host, key string, s *navigationPanelState, p NavigationMenuProps, choose func(string)) {
	t := theme.TokensOf(c)
	active := navigationKey(p.Items, s.active)
	if active < 0 {
		return
	}
	viewport := ui.Box(c).Fill().Clip().Role(ui.RoleNavigation).Label(p.Items[active].Label + " links")
	ownsFocus := ui.Local(viewport, "owns-window-focus", func() bool { return false })
	leaveTab := func(backward bool) {
		s.close(true, true)
		if p.OnTab != nil {
			if handle := s.handles[s.active]; handle != nil {
				handle.CancelFocus()
			}
			p.OnTab(backward)
		}
	}
	viewport.ObserveInput(func(ev ui.InputEvent) {
		if ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyTab && ev.Mods & ^ui.Shift == 0 {
			s.tabbed, s.tabBack = true, ev.Mods&ui.Shift != 0
		}
	})
	if p.Items[active].Content == nil {
		backward := viewport.Shortcut(ui.Shift, ui.KeyTab)
		forward := viewport.Shortcut(0, ui.KeyTab)
		if backward || forward {
			leaveTab(backward)
		}
	}
	if c.Shortcut(0, ui.KeyEscape) {
		s.close(true, s.keyboard)
	}
	// Horizontal keys switch trigger siblings while native text editors retain
	// their own caret keys. Up/Down remain native focus-group navigation.
	if !c.FocusTakesText() {
		back, forth := ui.KeyLeft, ui.KeyRight
		if c.Direction() == ui.RightToLeft {
			back, forth = forth, back
		}
		if viewport.Shortcut(0, back) {
			s.sibling = -1
			s.parent.Invalidate()
		}
		if viewport.Shortcut(0, forth) {
			s.sibling = 1
			s.parent.Invalidate()
		}
	}
	viewport.Children(func() {
		for index, item := range p.Items {
			if !navigationHasPanel(item) || !s.mounted[item.Key] {
				continue
			}
			visible := s.open && item.Key == s.active
			body := ui.Column(c.Key(item.Key)).Absolute().Top(0).Left(0).Width(navigationWidth(p, item)).Padding(12).Gap(4).
				Background(t.Popover).TextColor(t.PopoverForeground).Role(ui.RoleGroup).Label(item.Label).Disabled(!visible).DimDisabled(false).AccessHidden(!visible).IgnoreInput(!visible).Focusable().TabStop(false)
			if item.Content == nil {
				body.FocusGroup(ui.Vertical)
			} else {
				// The real panel region is the keyboard entry point and a native
				// Tab boundary, not a hidden synthetic sentinel. Tab traverses
				// arbitrary controls normally; returning to this region means it
				// passed the last/first control in the independent window.
				body.TabStop(true)
			}
			target := float32(0)
			if visible {
				target = 1
			}
			entered := ui.Local(body, "panel-entry", func() bool { return false })
			if !*entered {
				side := float32(s.direction)
				if c.Direction() == ui.RightToLeft {
					side = -side
				}
				body.Jump("panel-presence", 0)
				body.Jump("panel-offset", navigationWidth(p, item)*.5*side)
				*entered = true
			}
			amount := navigationAnimate(body, "panel-presence", target, navigationDuration(c, p))
			direction := float32(1)
			if index < active {
				direction = -1
			}
			if c.Direction() == ui.RightToLeft {
				direction = -direction
			}
			offset := float32(0)
			if !visible {
				offset = navigationWidth(p, item) * .5 * direction
			}
			actualOffset := navigationAnimate(body, "panel-offset", offset, navigationDuration(c, p))
			body.Opacity(amount).Translate(actualOffset, 0)
			body.Draw(func(_ *ui.Painter, r ui.Rect) {
				height := float32(math.Ceil(float64(r.H)))
				if height > 0 && s.heights[item.Key] != height {
					s.heights[item.Key] = height
					s.parent.Invalidate()
				}
			})
			var links []ui.Element
			panelChoose := func(route string) {
				if s.open && s.active == item.Key {
					choose(route)
				}
			}
			body.Children(func() {
				maxHeight := p.MaxHeight
				if maxHeight <= 0 {
					maxHeight = 480
				}
				scroll := ui.Scroll(c).FillWidth().MaxHeight(maxHeight - 24).Gap(4)
				if item.Height > 0 {
					scroll.Height(max(16, item.Height-24))
				}
				scroll.Children(func() {
					if item.Content != nil {
						item.Content(c, panelChoose)
					} else {
						navigationLinks(c, item.Links, max(1, item.Columns), s.selected, panelChoose, &links, false)
					}
				})
			})
			if visible {
				if body.Focused() {
					*ownsFocus = true
				}
				for _, link := range links {
					if link.Focused() {
						*ownsFocus = true
					}
				}
			}
			if visible && s.enter {
				if body.FocusWithin() {
					s.enter, s.enterLast = false, false
				} else {
					body.Focus()
					if len(links) > 0 {
						at := 0
						if s.enterLast {
							at = len(links) - 1
						}
						links[at].Focus()
					}
				}
			}
			if visible && item.Content != nil && body.Focused() {
				if body.Shortcut(ui.Shift, ui.KeyTab) {
					leaveTab(true)
				} else if s.tabbed {
					leaveTab(s.tabBack)
				}
			}
		}
	})
	s.tabbed = false
	if viewport.FocusLeft() {
		*ownsFocus = false
	}
	focus := *ownsFocus && viewport.FocusWithin()
	if actual, known := host.FocusState(key); known {
		focus = actual && viewport.FocusWithin()
		*ownsFocus = focus
	}
	if focus {
		s.keyboardPending = false
	}
	hover := viewport.Hovered()
	if s.childHover != hover || s.childFocus != focus {
		s.childHover, s.childFocus = hover, focus
		s.parent.Invalidate()
	}
}

func navigationLinks(c *ui.Context, items []NavItem, columns int, selected string, choose func(string), links *[]ui.Element, disabled bool) {
	t := theme.TokensOf(c)
	ui.Grid(c).Columns(columns).Gap(4).FillWidth().Children(func() {
		for _, item := range items {
			if len(item.Children) > 0 {
				ui.Column(c.Key(item.Key)).Gap(4).Disabled(disabled || item.Disabled).DimDisabled(false).Children(func() {
					ui.Text(c, item.Label).FontSize(12).FontWeight(500).TextColor(t.MutedForeground).Padding(8)
					navigationLinks(c, item.Children, 1, selected, choose, links, disabled || item.Disabled)
				})
				continue
			}
			disabled := disabled || item.Disabled
			link := motion.ApplyTransition(c, ui.ButtonBase(c.Key(item.Key)).Role(ui.RoleLink).Label(item.Label).Description(item.Description).Disabled(disabled).DimDisabled(false).
				Column().AlignItems(ui.Start).Gap(4).Padding(12).Radius(t.RadiusMD()).TextColor(t.Foreground)).MinWidth(0)
			if !disabled {
				*links = append(*links, link)
			} else {
				link.Opacity(.5)
			}
			if item.Key == selected {
				link.Current(ui.CurrentPage).Background(t.Accent)
			}
			if link.Hovered() || link.Focused() {
				link.Background(t.Accent).TextColor(t.AccentForeground)
			}
			link.Children(func() {
				if item.Icon != nil {
					ui.Icon(c, item.Icon).Size(16, 16)
				}
				ui.Text(c, item.Label).FontSize(14).FontWeight(500)
				if item.Description != "" {
					ui.Text(c, item.Description).FontSize(14).TextColor(t.MutedForeground)
				}
			})
			if link.Clicked() && !disabled {
				choose(item.Key)
			}
		}
	})
}

// NativeNavigationMenu is a route menu group backed by operating-system menus.
// Arbitrary rich content belongs to NavigationMenuWith, not to the OS adapter.
func NativeNavigationMenu(c *ui.Context, selected *string, items []NavItem) ui.Element {
	if selected == nil {
		panic("navigation: NativeNavigationMenu requires a route pointer")
	}
	policy := c.MenuPresentation()
	c.SetMenuPresentation(ui.SystemMenus)
	defer c.SetMenuPresentation(policy)
	root := ui.Row(c).Gap(4).Role(ui.RoleNavigation).Label("Main navigation").FocusGroup(ui.Horizontal)
	root.Children(func() {
		for _, item := range items {
			if len(item.Children) > 0 {
				value := ""
				ui.Button(c.Key(item.Key), item.Label).Disabled(item.Disabled).Menu(func(menu *ui.Menu) { nativeNavigationItems(menu, item.Children, *selected, &value) })
				if value != "" && value != *selected {
					*selected = value
					root.SetChanged()
				}
			} else {
				link := ui.Button(c.Key(item.Key), item.Label).Role(ui.RoleLink).Disabled(item.Disabled)
				if item.Key == *selected {
					link.Current(ui.CurrentPage)
				}
				if link.Clicked() && item.Key != *selected {
					*selected = item.Key
					root.SetChanged()
				}
			}
		}
	})
	return root
}

func nativeNavigationItems(menu *ui.Menu, items []NavItem, selected string, chosen *string) {
	group := ""
	for _, item := range items {
		if item.Group != "" && item.Group != group {
			group = item.Group
			menu.Item(group).Disabled(true)
		}
		if len(item.Children) > 0 {
			menu.Submenu(item.Label, func(sub *ui.Menu) { nativeNavigationItems(sub, item.Children, selected, chosen) }).ID(item.Key).Disabled(item.Disabled)
			continue
		}
		if menu.Item(item.Label).ID(item.Key).Disabled(item.Disabled).Radio(item.Key == selected).Chosen() && !item.Disabled {
			*chosen = item.Key
		}
	}
}

func navigationAnimate(element ui.Element, key any, target float32, duration time.Duration) float32 {
	if duration < 0 {
		return element.Jump(key, target)
	}
	return element.AnimateWith(key, target, duration, motion.EaseOut)
}
