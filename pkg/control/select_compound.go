package control

import (
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// SelectSelection distinguishes no selection from a selected empty identity.
type SelectSelection struct {
	Value string
	Valid bool
}

// SelectState holds application values only. Keep its address stable while
// mounted; native gestures update the same values that the application reads.
type SelectState struct {
	Value  SelectSelection
	Values []string
	Open   bool
}

// SelectChange describes a request in the current build. Cancel preserves the
// caller's value. Events and scopes must not be retained as subscriptions.
type SelectChange struct {
	Reason   string
	canceled bool
}

func (e *SelectChange) Cancel()        { e.canceled = true }
func (e *SelectChange) Canceled() bool { return e.canceled }

type SelectRootProps struct {
	// Nil Options registers manually rendered Item slots. Non-nil Options also
	// supplies closed-trigger labels and typeahead before content is opened.
	Options []Option
	// UnselectedValue identifies an optional item which clears the single
	// selection. Without it, an empty string is an ordinary valid identity.
	UnselectedValue              *string
	Multiple, Disabled, ReadOnly bool
	OnValueChange                func(SelectSelection, *SelectChange)
	OnValuesChange               func([]string, *SelectChange)
	OnOpenChange                 func(bool, *SelectChange)
}

type SelectTriggerProps struct {
	Label             string
	Size              Size
	Disabled, Invalid bool
	// Render supplies an actual native button. The slot installs select roles,
	// input and focus behavior; its visuals still belong to the application.
	Render func(*ui.Context) ui.Element
}

type SelectValueProps struct{ Placeholder string }
type SelectValueView struct {
	Value    SelectSelection
	Values   []string
	Options  []Option
	Multiple bool
}

type SelectSide string
type SelectAlign string

const (
	SelectSideBottom      SelectSide  = "bottom"
	SelectSideTop         SelectSide  = "top"
	SelectSideLeft        SelectSide  = "left"
	SelectSideRight       SelectSide  = "right"
	SelectSideInlineStart SelectSide  = "inline-start"
	SelectSideInlineEnd   SelectSide  = "inline-end"
	SelectAlignStart      SelectAlign = "start"
	SelectAlignCenter     SelectAlign = "center"
	SelectAlignEnd        SelectAlign = "end"
)

type SelectContentProps struct {
	Side  SelectSide
	Align SelectAlign
	// Nil uses the fixed wrapper's 4-DIP offset and item-aligned layout. False
	// opts into ordinary side placement and the source's 100ms content motion.
	SideOffset           *float32
	AlignItemWithTrigger *bool
	AlignOffset          float32
	Width, MaxHeight     float32
	Anchor               *ui.Handle
	Style                func(ui.Element)
}

type SelectListProps struct {
	Label  string
	Scroll *ui.ScrollState
}
type SelectItemProps struct {
	Label    string
	Disabled bool
}

// SelectScope contains independently composable, current-build native slots.
// Content does not imply List or Group: callers choose the actual layout.
type SelectScope struct {
	f *selectFrame
}

type selectMountKey struct{ identity any }
type selectItemKey struct{ value string }
type selectIntent struct {
	kind, value, reason string
	open                bool
}
type selectItemGeometry struct {
	value     string
	box, text ui.Rect
}
type selectRuntime struct {
	initialized, closed, open, enabled, multiple       bool
	legacySkipDisabled, touch, pointerActive           bool
	active                                             string
	hasActive                                          bool
	options                                            []Option
	selected                                           []string
	intents                                            []selectIntent
	trigger                                            ui.Handle
	moveFocus, reveal, opening                         bool
	openDirection                                      int
	typeahead                                          string
	typeaheadStart, typeaheadMatch                     int
	typeaheadAt                                        time.Time
	composing                                          bool
	scroll                                             ui.ScrollState
	motion                                             comboMotion
	geometry                                           []selectItemGeometry
	panelBox, valueBox, triggerBox                     ui.Rect
	contentHeight                                      float32
	alignX, alignY, alignHeight                        float32
	alignPreferredX                                    float32
	alignWindowW, alignWindowH, alignWidth, alignLimit float32
	aligned, measured, alignedScroll, alignTopPinned   bool
	arrowDirection                                     int
	arrowAt                                            time.Time
}

func (s *selectRuntime) Close() error { *s = selectRuntime{closed: true}; return nil }

type selectItemFrame struct {
	value             string
	item, text, check ui.Element
	disabled          bool
}
type selectValueFrame struct {
	text        ui.Element
	placeholder string
}
type selectFrame struct {
	c                   *ui.Context
	s                   *selectRuntime
	state               *SelectState
	p                   SelectRootProps
	alive, manual       bool
	changed             bool
	trigger, value      ui.Element
	panel, list         ui.Element
	group               ui.Element
	options             []Option
	items               []selectItemFrame
	values              []selectValueFrame
	inContent, inList   bool
	content             SelectContentProps
	scroll              *ui.ScrollState
	itemAligned, hidden bool
}

// SelectRoot is layout-transparent. A zero-size, inaccessible native mount
// releases its runtime when removed. Persistent state contains only values,
// geometry and Handles; callbacks and Elements belong to this build alone.
func SelectRoot(c *ui.Context, state *SelectState, p SelectRootProps, build func(*SelectScope)) {
	selectRoot(c, state, state, p, false, false, build)
}

func selectRoot(c *ui.Context, identity any, state *SelectState, p SelectRootProps, localOpen, legacySkipDisabled bool, build func(*SelectScope)) {
	if state == nil {
		panic("control.SelectRoot: state is required")
	}
	mount := ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true).AccessHidden(true)
	// Convenience bindings may be temporary values reconstructed each build,
	// as in a calendar caption. Their native mount follows ordinary positional/
	// caller-key identity; explicit Root state retains its documented stable key.
	if identity != nil {
		mount.Key(selectMountKey{identity})
	}
	s := ui.Local(mount, "shadcn.control.select", func() selectRuntime {
		return selectRuntime{typeaheadStart: -1, typeaheadMatch: -1}
	})
	if localOpen {
		state.Open = s.open
	}
	f := &selectFrame{c: c, s: s, state: state, p: p, alive: true, manual: p.Options == nil, scroll: &s.scroll}
	if p.Options != nil {
		f.options = slices.Clone(p.Options)
	} else {
		f.options = slices.Clone(s.options)
	}
	f.validateOptions()
	s.options = slices.Clone(f.options)
	s.enabled, s.multiple, s.legacySkipDisabled = !p.Disabled && !p.ReadOnly, p.Multiple, legacySkipDisabled
	f.selected()
	if !s.initialized {
		s.initialized, s.open = true, state.Open
		if s.open {
			s.opening, s.moveFocus, s.reveal = true, true, true
		}
	} else if state.Open != s.open {
		if !state.Open {
			s.trigger.Focus()
		}
		s.open, s.opening = state.Open, state.Open
		s.moveFocus, s.reveal = state.Open, state.Open
		s.measured, s.alignedScroll = false, false
	}
	for _, intent := range s.intents {
		switch intent.kind {
		case "open":
			f.setOpen(intent.open, intent.reason)
		case "choose":
			f.choose(intent.value, intent.reason)
		}
	}
	s.intents = nil
	if s.open != state.Open {
		s.opening, s.moveFocus, s.reveal = state.Open, state.Open, state.Open
		if state.Open {
			s.measured, s.alignedScroll = false, false
		} else {
			s.trigger.Focus()
		}
	}
	s.open = state.Open
	f.selected()
	if s.opening {
		f.initialHighlight()
	}
	if build != nil {
		build(&SelectScope{f: f})
	}
	f.finish()
}

func (f *selectFrame) check() {
	if f == nil || !f.alive {
		panic("control.SelectScope: slots are valid only during SelectRoot's build callback")
	}
}
func (f *selectFrame) validateOptions() {
	seen := map[string]bool{}
	for _, o := range f.options {
		if o.Separator {
			continue
		}
		if seen[o.Value] {
			panic("control.SelectRoot: duplicate item identity " + o.Value)
		}
		seen[o.Value] = true
	}
}
func (f *selectFrame) selected() {
	f.s.selected = nil
	if f.p.Multiple {
		for _, value := range f.state.Values {
			if !slices.Contains(f.s.selected, value) {
				f.s.selected = append(f.s.selected, value)
			}
		}
	} else if f.state.Value.Valid {
		f.s.selected = []string{f.state.Value.Value}
	} else if f.p.UnselectedValue != nil {
		f.s.selected = []string{*f.p.UnselectedValue}
	}
}
func (f *selectFrame) setOpen(open bool, reason string) {
	if f.state.Open == open {
		f.s.open = open
		return
	}
	e := &SelectChange{Reason: reason}
	if f.p.OnOpenChange != nil {
		f.p.OnOpenChange(open, e)
	}
	if e.canceled {
		f.s.moveFocus = f.state.Open
		f.s.open = f.state.Open
		return
	}
	f.state.Open, f.s.open = open, open
	f.s.opening, f.s.moveFocus, f.s.reveal = open, open, open
	if open {
		f.s.measured, f.s.alignedScroll = false, false
	}
}
func (f *selectFrame) choose(value, reason string) {
	if f.p.Disabled || f.p.ReadOnly {
		return
	}
	for _, o := range f.options {
		if o.Separator || o.Value != value || o.Disabled {
			continue
		}
		e := &SelectChange{Reason: reason}
		if f.p.Multiple {
			next := slices.Clone(f.state.Values)
			if i := slices.Index(next, value); i >= 0 {
				next = slices.DeleteFunc(next, func(v string) bool { return v == value })
			} else {
				next = append(next, value)
			}
			if f.p.OnValuesChange != nil {
				f.p.OnValuesChange(slices.Clone(next), e)
			}
			if !e.canceled {
				f.state.Values, f.changed = next, true
			}
		} else {
			next := SelectSelection{value, true}
			if f.p.UnselectedValue != nil && value == *f.p.UnselectedValue {
				next = SelectSelection{}
			}
			if f.p.OnValueChange != nil {
				f.p.OnValueChange(next, e)
			}
			if !e.canceled && next != f.state.Value {
				f.state.Value, f.changed = next, true
			}
			if reason != "typeahead" {
				f.setOpen(false, reason)
				f.s.trigger.Focus()
			}
		}
		f.selected()
		return
	}
}

func (f *selectFrame) initialHighlight() {
	s := f.s
	s.active, s.hasActive = "", false
	if len(s.selected) > 0 {
		v := s.selected[len(s.selected)-1]
		for _, o := range f.options {
			if !o.Separator && o.Value == v && (!o.Disabled || !s.legacySkipDisabled) {
				s.active, s.hasActive = v, true
				return
			}
		}
	}
	if s.openDirection != 0 {
		s.move(s.openDirection, true)
	}
}
func (s *selectRuntime) move(delta int, opening bool) {
	items := []Option{}
	for _, o := range s.options {
		if !o.Separator {
			items = append(items, o)
		}
	}
	i := -1
	if delta < 0 {
		i = len(items)
	}
	if !opening && s.hasActive {
		for j, o := range items {
			if o.Value == s.active {
				i = j
				break
			}
		}
	}
	for at := i + delta; at >= 0 && at < len(items); at += delta {
		if (opening || s.legacySkipDisabled) && items[at].Disabled {
			continue
		}
		s.active, s.hasActive, s.moveFocus, s.reveal, s.pointerActive = items[at].Value, true, true, true, false
		return
	}
}

func (s *selectRuntime) input(ev ui.InputEvent) bool {
	if ev.Kind == ui.InputPointerDown {
		s.touch = ev.Pointer == ui.PointerTouch
	}
	if !s.enabled {
		return false
	}
	if ev.Kind == ui.InputCompose {
		s.composing = ev.Text != ""
		return false
	}
	if ev.Kind == ui.InputText {
		s.composing = false
		if ev.Text != "" && ev.Mods&^(ui.Shift) == 0 && (s.open || !s.multiple) {
			s.match(ev.Text, ev.Time)
			return true
		}
		return false
	}
	if ev.Kind == ui.InputKeyDown && s.composing {
		return ev.Key == ui.KeyEnter || ev.Key == ui.KeySpace || ev.Key == ui.KeyDown || ev.Key == ui.KeyUp
	}
	if ev.Kind != ui.InputKeyDown || ev.Mods&^ui.Shift != 0 {
		return false
	}
	if ev.Key == ui.KeyTab {
		if s.open {
			s.intents = append(s.intents, selectIntent{kind: "open", open: false, reason: "focus-out"})
			s.open = false
		}
		return false
	}
	if ev.Mods != 0 {
		return false
	}
	if !s.open {
		switch ev.Key {
		case ui.KeyDown, ui.KeyUp, ui.KeyEnter, ui.KeySpace:
			if ev.Key == ui.KeySpace && s.typeahead != "" && ev.Time.Sub(s.typeaheadAt) < 750*time.Millisecond {
				return false
			}
			s.openDirection = 1
			if ev.Key == ui.KeyUp {
				s.openDirection = -1
			}
			s.intents = append(s.intents, selectIntent{kind: "open", open: true, reason: "trigger-key"})
			return true
		}
		return false
	}
	switch ev.Key {
	case ui.KeyEscape:
		s.intents = append(s.intents, selectIntent{kind: "open", reason: "escape-key"})
		s.trigger.Focus()
		return true
	case ui.KeyDown:
		s.move(1, false)
		return true
	case ui.KeyUp:
		s.move(-1, false)
		return true
	case ui.KeyHome, ui.KeyEnd:
		delta := 1
		if ev.Key == ui.KeyEnd {
			delta = -1
		}
		s.move(delta, true)
		return true
	case ui.KeyEnter, ui.KeySpace:
		if ev.Key == ui.KeySpace && s.typeahead != "" && ev.Time.Sub(s.typeaheadAt) < 750*time.Millisecond {
			return false
		}
		if s.hasActive {
			s.intents = append(s.intents, selectIntent{kind: "choose", value: s.active, reason: "item-press"})
		}
		return true
	}
	return false
}

func (s *selectRuntime) match(text string, now time.Time) {
	if now.Sub(s.typeaheadAt) >= 750*time.Millisecond {
		s.typeahead = ""
	}
	s.typeaheadAt = now
	for _, typed := range text {
		letter := strings.ToLower(string(typed))
		if s.typeahead == "" {
			s.typeaheadStart = -1
			value := ""
			valid := false
			if len(s.selected) > 0 {
				value, valid = s.selected[len(s.selected)-1], true
			} else if s.open && s.hasActive {
				value, valid = s.active, true
			}
			if valid {
				for i, o := range s.options {
					if !o.Separator && o.Value == value {
						s.typeaheadStart = i
						break
					}
				}
			}
		}
		rapid := true
		for _, o := range s.options {
			if o.Separator || o.Disabled {
				continue
			}
			r := []rune(strings.ToLower(optionLabel(o)))
			if len(r) > 1 && r[0] == r[1] {
				rapid = false
				break
			}
		}
		if rapid && s.typeahead == letter {
			s.typeahead = ""
			s.typeaheadStart = s.typeaheadMatch
		}
		s.typeahead += letter
		found := -1
		for j := 0; j < len(s.options); j++ {
			i := (s.typeaheadStart + 1 + j + len(s.options)) % len(s.options)
			o := s.options[i]
			if !o.Separator && !o.Disabled && strings.HasPrefix(strings.ToLower(optionLabel(o)), s.typeahead) {
				found = i
				break
			}
		}
		if found < 0 {
			if letter != " " {
				s.typeahead = ""
			}
			continue
		}
		s.typeaheadMatch = found
		value := s.options[found].Value
		if s.open {
			s.active, s.hasActive, s.moveFocus, s.reveal = value, true, true, true
		} else {
			s.intents = append(s.intents, selectIntent{kind: "choose", value: value, reason: "typeahead"})
		}
	}
}

func (b *SelectScope) Trigger(c *ui.Context, p SelectTriggerProps, children func()) ui.Element {
	f := b.f
	f.check()
	if f.trigger.Valid() {
		panic("control.SelectScope: only one Trigger is supported")
	}
	tr := ui.Element{}
	if p.Render != nil {
		tr = p.Render(c)
	} else {
		tr = ui.ButtonBase(c)
	}
	h, _ := metrics(p.Size)
	if p.Size == Small {
		h = 32
	}
	t := theme.TokensOf(c)
	invalid := fieldInputInvalid(c, p.Invalid)
	if p.Render == nil {
		motion.ApplyTransition(c, tr.MinHeight(h).Padding(4, 8).PaddingStart(10).Gap(6).AlignItems(ui.Center).Radius(t.RadiusMD()).FontSize(14).LineHeight(20.0/14).Background(ui.Transparent).TextColor(t.Foreground))
		if t.Dark {
			tr.Background(t.Input.Alpha(.3))
		}
		textInputAppearance(c, tr, invalid)
	}
	disabled := p.Disabled || f.p.Disabled
	tr.Role(ui.RoleComboBox).HasPopup(ui.PopupListBox).Expanded(f.s.open).ReadOnly(f.p.ReadOnly).Disabled(disabled).Invalid(invalid).HandleInput(f.s.input).Bind(&f.s.trigger).FocusRing(false).Cursor(ui.CursorPointer)
	if disabled {
		tr.DimDisabled(false).Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	if p.Label != "" {
		tr.Label(p.Label)
	}
	f.trigger = tr
	tr.Children(func() {
		if children != nil {
			children()
		} else {
			b.Value(c, SelectValueProps{}, nil)
		}
		indicator.ChevronDown(c).TextColor(t.MutedForeground)
	})
	if tr.Clicked() && !disabled && !f.p.ReadOnly {
		f.s.openDirection = 0
		f.setOpen(!f.s.open, "trigger-press")
		if f.s.open {
			f.initialHighlight()
		}
	}
	return ButtonGroupItem(c, tr)
}

func (b *SelectScope) Value(c *ui.Context, p SelectValueProps, children func(SelectValueView)) ui.Element {
	f := b.f
	f.check()
	row := ui.Row(c).Grow(1).MinWidth(0).Gap(6).AlignItems(ui.Center)
	f.value = row
	view := SelectValueView{Value: f.state.Value, Values: slices.Clone(f.state.Values), Options: slices.Clone(f.options), Multiple: f.p.Multiple}
	row.Children(func() {
		if children != nil {
			children(view)
			return
		}
		f.values = append(f.values, selectValueFrame{row, p.Placeholder})
	})
	if (!f.p.Multiple && !f.state.Value.Valid) || (f.p.Multiple && len(f.state.Values) == 0) {
		row.TextColor(theme.TokensOf(c).MutedForeground)
	}
	return row
}
func (f *selectFrame) valueText(placeholder string) string {
	labels := []string{}
	for _, v := range f.s.selected {
		label := v
		for _, o := range f.options {
			if !o.Separator && o.Value == v {
				label = optionLabel(o)
				break
			}
		}
		labels = append(labels, label)
	}
	if len(labels) > 0 {
		return strings.Join(labels, ", ")
	}
	if placeholder == "" {
		return "Select an option"
	}
	return placeholder
}

func (b *SelectScope) Content(c *ui.Context, p SelectContentProps, children func()) ui.Element {
	f := b.f
	f.check()
	if !f.trigger.Valid() {
		panic("control.SelectScope: Content requires an earlier Trigger")
	}
	f.content = p
	anchor := f.trigger
	if p.Anchor != nil {
		anchor = c.Resolve(*p.Anchor)
	}
	if !anchor.Valid() {
		return ui.Element{}
	}
	f.itemAligned = p.AlignItemWithTrigger == nil || *p.AlignItemWithTrigger
	f.itemAligned = f.itemAligned && !f.s.touch
	var pose comboPose
	var present bool
	if f.itemAligned {
		pose, present = comboPoseShown, f.s.open
		if f.s.open {
			f.s.motion = comboMotion{initialized: true, shown: true, from: comboPoseShown, to: comboPoseShown}
		} else {
			f.s.motion = comboMotion{}
		}
	} else {
		pose, present = f.s.motion.update(c, f.s.open)
	}
	f.hidden = !present
	if !present && !f.manual {
		return ui.Element{}
	}
	menu := theme.MenuOf(c).ForScene()
	t := &menu.Tokens
	build := func() {
		panel := ui.Themed(c, menu.Theme()).Key(f.s).Role(ui.RolePopup).Radius(t.RadiusMD()).Border(1, t.Foreground.Alpha(.1)).Background(t.Popover).TextColor(t.PopoverForeground).Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Clip().IgnoreInput(!f.s.open).AccessHidden(!f.s.open)
		menu.Surface(c, panel)
		f.panel = panel
		panel.Clicked()
		w, h := c.Size()
		width := max(float32(144), anchor.LayoutBounds().W)
		if p.Width > 0 {
			width = p.Width
		}
		width = min(width, max(float32(1), w-10))
		maxHeight := max(float32(1), h-20)
		if p.MaxHeight > 0 {
			maxHeight = min(maxHeight, p.MaxHeight)
		}
		if f.itemAligned && f.s.open {
			s := f.s
			resized := s.alignWindowW != w || s.alignWindowH != h || s.alignWidth != width || s.alignLimit != maxHeight
			if s.measured && s.aligned {
				if resized {
					// A measured pose is bounded by the current window, not the
					// window in which it opened. Re-reveal once after reflow so
					// active rows remain reachable without taking over user scroll.
					if s.alignWindowH != h || s.alignLimit != maxHeight {
						s.alignHeight = max(float32(1), min(s.contentHeight, maxHeight))
					}
					s.reveal = true
					c.Invalidate()
				}
				s.alignHeight = max(float32(1), min(s.alignHeight, maxHeight))
				s.alignX = max(float32(5), min(w-width-5, s.alignPreferredX))
				s.alignY = max(float32(10), min(h-s.alignHeight-10, s.alignY))
			}
			s.alignWindowW, s.alignWindowH, s.alignWidth, s.alignLimit = w, h, width, maxHeight
		}
		panel.Width(width).MaxWidth(max(float32(1), w-10)).MaxHeight(maxHeight)
		if f.hidden {
			panel.Absolute().Left(-10000).Top(-10000).Opacity(0).IgnoreInput(true).AccessHidden(true)
		} else {
			side, align := p.Side, p.Align
			if side == "" {
				side = SelectSideBottom
			}
			if align == "" {
				align = SelectAlignCenter
			}
			at, self := comboAnchors(ComboboxSide(side), ComboboxAlign(align), c.Direction())
			panel.AttachTo(anchor, at, self).AttachEdgeMargin(5)
			gap := float32(4)
			if p.SideOffset != nil {
				gap = *p.SideOffset
			}
			physical := comboPhysicalSide(ComboboxSide(side), c.Direction())
			switch physical {
			case ComboboxSideTop:
				panel.Margin(0, 0, gap, 0)
			case ComboboxSideLeft:
				panel.Margin(0, gap, 0, 0)
			case ComboboxSideRight:
				panel.Margin(0, 0, 0, gap)
			default:
				panel.Margin(gap, 0, 0, 0)
			}
			if physical == ComboboxSideTop || physical == ComboboxSideBottom {
				offset := p.AlignOffset
				if c.Direction() == ui.RightToLeft {
					offset = -offset
				}
				panel.Left(offset)
			} else {
				panel.Top(p.AlignOffset)
			}
			if f.itemAligned && f.s.measured && f.s.aligned {
				// AttachTo retains native popover focus/Tab ownership. Its Top/Left
				// offsets apply after edge fitting, so compensate for the fitted
				// top-left origin rather than subtracting the raw anchor position.
				a := anchor.LayoutBounds()
				x := selectTopLeftAttachmentBase(a.X, a.W, width, w, 5)
				y := selectTopLeftAttachmentBase(a.Y, a.H, f.s.alignHeight, h, 5)
				panel.AttachTo(anchor, ui.AnchorTopLeft, ui.AnchorTopLeft).Margin(0).Left(f.s.alignX - x).Top(f.s.alignY - y).Height(f.s.alignHeight).MaxHeight(f.s.alignHeight)
			}
		}
		if p.Style != nil {
			p.Style(panel)
		}
		previous := f.inContent
		f.inContent = true
		panel.Children(children)
		f.inContent = previous
		if f.s.open {
			outside := panel.PressedOutside()
			if outside || panel.OverlayShortcut(0, ui.KeyEscape) {
				reason := "escape-key"
				if outside {
					reason = "outside-press"
				}
				f.setOpen(false, reason)
				if !outside {
					f.s.trigger.Focus()
				}
			}
		}
	}
	ui.Overlay(c, build)
	if !f.panel.Valid() {
		return ui.Element{}
	}
	f.panel.IgnoreInput(!f.s.open).AccessHidden(!f.s.open)
	if !f.hidden {
		f.s.motion.observe(anchor, f.panel)
		if !f.itemAligned {
			f.s.motion.apply(c, anchor, f.panel, pose)
		}
		f.observeGeometry(c, anchor)
	}
	return f.panel
}

// The public top-left attachment can fit on either end of its anchor, then
// move into the window. Account for that base before applying an exact pose.
func selectTopLeftAttachmentBase(start, anchorSize, panelSize, limit, edge float32) float32 {
	overflow := func(at float32) float32 {
		return max(float32(0), edge-at) + max(float32(0), at+panelSize-limit+edge)
	}
	at := start
	if other := start + anchorSize - panelSize; overflow(other) < overflow(at) {
		at = other
	}
	return max(edge, min(at, limit-edge-panelSize))
}

func (b *SelectScope) List(c *ui.Context, p SelectListProps, children func()) ui.Element {
	f := b.f
	f.check()
	if !f.inContent {
		panic("control.SelectScope: List belongs inside Content")
	}
	if p.Scroll != nil {
		f.scroll = p.Scroll
	}
	list := ui.Scroll(c).TrackScroll(f.scroll).MinHeight(0).FillWidth().Scrollbars(false).Role(ui.RoleListBox).Multiselectable(f.p.Multiple).Label(p.Label)
	_, height := c.Size()
	maxHeight := max(float32(1), height-20)
	if f.content.MaxHeight > 0 {
		maxHeight = min(maxHeight, f.content.MaxHeight)
	}
	list.MaxHeight(maxHeight)
	if f.itemAligned && f.s.measured && f.s.aligned {
		list.Height(f.s.alignHeight - 2)
	}
	f.list = list
	if f.trigger.Valid() {
		f.trigger.Controls(list)
	}
	previous := f.inList
	f.inList = true
	list.Children(children)
	f.inList = previous
	return list
}
func (b *SelectScope) Group(c *ui.Context, label string, children func()) ui.Element {
	f := b.f
	f.check()
	group := ui.Column(c).FillWidth().Padding(4).Role(ui.RoleGroup).Label(label)
	group.Children(func() {
		previous := f.group
		f.group = group
		defer func() { f.group = previous }()
		if label != "" {
			b.Label(c, label)
		}
		if children != nil {
			children()
		}
	})
	return group
}
func (b *SelectScope) Label(c *ui.Context, text string) ui.Element {
	b.f.check()
	if !b.f.group.Valid() {
		panic("control.SelectScope: Label belongs inside Group")
	}
	label := ui.Text(c, text).Padding(6, 8).FontSize(12).LineHeight(4.0 / 3).TextColor(theme.TokensOf(c).MutedForeground)
	b.f.group.LabelledBy(label)
	return label
}
func (b *SelectScope) Separator(c *ui.Context) ui.Element {
	b.f.check()
	return ui.Box(c).FillWidth().Height(1).Margin(4, -4).Background(theme.TokensOf(c).Border).IgnoreInput(true).AccessHidden(true)
}
func (b *SelectScope) Collection(c *ui.Context, children func(*ui.Context, Option, int)) {
	f := b.f
	f.check()
	for i, o := range f.options {
		if children != nil {
			children(c.Key(selectItemKey{o.Value}), o, i)
		} else if o.Separator {
			b.Separator(c)
		} else {
			b.Item(c, o.Value, SelectItemProps{}, nil)
		}
	}
}
func (b *SelectScope) Item(c *ui.Context, value string, p SelectItemProps, children func()) ui.Element {
	f := b.f
	f.check()
	if !f.inList {
		panic("control.SelectScope: Item belongs inside List")
	}
	for _, old := range f.items {
		if old.value == value {
			panic("control.SelectScope: duplicate Item " + value)
		}
	}
	o := Option{Value: value, Label: p.Label, Disabled: p.Disabled}
	known := false
	for i, item := range f.options {
		if !item.Separator && item.Value == value {
			o = item
			o.Disabled = o.Disabled || p.Disabled
			if p.Label != "" {
				o.Label = p.Label
			}
			if f.manual {
				o.Disabled = p.Disabled
			}
			f.options[i] = o
			known = true
			break
		}
	}
	if !known {
		if !f.manual {
			panic("control.SelectScope: Item missing from Options: " + value)
		}
		f.options = append(f.options, o)
	}
	menu := theme.MenuOf(c).ForScene()
	t := &menu.Tokens
	item := ui.ButtonBase(c.Key(selectItemKey{value})).Role(ui.RoleOption).Label(optionLabel(o)).TabStop(false).MinHeight(32).Padding(6, 8).PaddingEnd(32).Gap(8).AlignSelf(ui.Stretch).AlignItems(ui.Center).Radius(t.RadiusSM()).FontSize(14).LineHeight(20.0 / 14).AccessDisabled(o.Disabled).DimDisabled(false).FocusRing(false).Cursor(ui.CursorDefault)
	item.HandleInput(f.s.input)
	if o.Disabled {
		item.Opacity(.5)
	}
	active := f.s.hasActive && f.s.active == value
	if !o.Disabled {
		runtime := f.s
		item.ObserveInput(func(ev ui.InputEvent) {
			if ev.Kind == ui.InputPointerMove && ev.Pointer != ui.PointerTouch && runtime.open && runtime.enabled {
				runtime.active, runtime.hasActive, runtime.moveFocus, runtime.pointerActive = value, true, true, true
			}
		})
	}

	var text, check ui.Element
	item.Children(func() {
		text = ui.Row(c).Grow(1).MinWidth(0).Gap(8).AlignItems(ui.Center).IgnoreInput(true)
		text.Children(func() {
			if children != nil {
				children()
			} else {
				ui.Text(c, optionLabel(o)).SingleLine()
			}
		})
		check = ui.Box(c).Absolute().End(8).TopPercent(50).Size(16, 16).Translate(0, -8).IgnoreInput(true).AccessHidden(true)
	})
	f.items = append(f.items, selectItemFrame{value, item, text, check, o.Disabled})
	if item.Clicked() && !f.hidden {
		f.choose(value, "item-press")
	}
	if active {
		bg, fg := menu.Highlight(false)
		item.Background(bg).TextColor(fg)
	}
	return item
}

func (b *SelectScope) ScrollUpButton(c *ui.Context) ui.Element   { return b.scrollArrow(c, -1) }
func (b *SelectScope) ScrollDownButton(c *ui.Context) ui.Element { return b.scrollArrow(c, 1) }
func (b *SelectScope) scrollArrow(c *ui.Context, direction int) ui.Element {
	f := b.f
	f.check()
	if !f.inContent {
		panic("control.SelectScope: scroll buttons belong inside Content")
	}
	if f.hidden || f.s.touch {
		return ui.Element{}
	}
	visible := f.scroll.Y > 0
	if direction > 0 {
		visible = f.scroll.Y < f.scroll.MaxY
	}
	if !visible {
		return ui.Element{}
	}
	t := theme.TokensOf(c)
	e := ui.Box(c).Absolute().Start(0).End(0).Height(24).ZIndex(10).Background(t.Popover).AccessHidden(true).AlignItems(ui.Center).Justify(ui.Center)
	if direction < 0 {
		e.Top(0)
	} else {
		e.Bottom(0)
	}
	e.Children(func() {
		if direction < 0 {
			indicator.ChevronDown(c).Rotate(180)
		} else {
			indicator.ChevronDown(c)
		}
	})
	if e.Hovered() {
		if f.s.arrowDirection != direction {
			f.s.arrowDirection, f.s.arrowAt = direction, c.Now()
		}
		remaining := 40*time.Millisecond - c.Now().Sub(f.s.arrowAt)
		if remaining <= 0 {
			f.s.hasActive = false
			f.s.arrowAt = c.Now()
			f.scroll.Y = selectArrowTarget(f.s.geometry, f.s.panelBox, direction, f.scroll.Y, f.scroll.MaxY, 24)
			c.Invalidate()
			remaining = 40 * time.Millisecond
		}
		c.After(remaining)
	} else if f.s.arrowDirection == direction {
		f.s.arrowDirection = 0
	}
	return e
}
func selectArrowTarget(items []selectItemGeometry, panel ui.Rect, direction int, y, maxY, arrowHeight float32) float32 {
	if direction < 0 {
		for i, g := range items {
			if g.box.Y >= panel.Y+arrowHeight-1 {
				if i > 0 {
					return max(float32(0), min(maxY, y+items[i-1].box.Y-panel.Y-arrowHeight))
				}
				return 0
			}
		}
		return 0
	}
	for _, g := range items {
		if g.box.Y+g.box.H > panel.Y+panel.H-arrowHeight+1 {
			return min(maxY, max(float32(0), y+g.box.Y+g.box.H-panel.Y-panel.H+arrowHeight))
		}
	}
	return maxY
}

func (f *selectFrame) observeGeometry(c *ui.Context, anchor ui.Element) {
	panel, value, list := f.panel, f.value, f.list
	items := slices.Clone(f.items)
	s, scroll := f.s, f.scroll
	itemAligned, maxHeight := f.itemAligned, f.content.MaxHeight
	reveal := s.reveal
	panel.Draw(func(_ *ui.Painter, _ ui.Rect) {
		s.panelBox, s.triggerBox = panel.LayoutBounds(), anchor.LayoutBounds()
		if value.Valid() {
			s.valueBox = value.LayoutBounds()
		} else {
			s.valueBox = s.triggerBox
		}
		s.geometry = s.geometry[:0]
		for _, item := range items {
			s.geometry = append(s.geometry, selectItemGeometry{item.value, item.item.LayoutBounds(), item.text.LayoutBounds()})
		}
		if len(items) > 0 {
			s.contentHeight = items[len(items)-1].item.LayoutBounds().Y + items[len(items)-1].item.LayoutBounds().H - items[0].item.LayoutBounds().Y + 10
		}
		if list.Valid() {
			s.contentHeight = scroll.MaxY + list.LayoutBounds().H + 2
		}
		// Native ScrollIntoView sees the scroll viewport, but not the two
		// overlaid arrow gutters. Correct only a requested keyboard/open reveal;
		// ordinary wheel/arrow scrolling must remain free to leave active behind.
		if reveal && s.open && s.hasActive && list.Valid() {
			viewport := list.LayoutBounds()
			top, bottom := viewport.Y, viewport.Y+viewport.H
			if scroll.Y > 0 {
				top += 24
			}
			if scroll.Y < scroll.MaxY {
				bottom -= 24
			}
			for _, item := range s.geometry {
				if item.value != s.active {
					continue
				}
				y := scroll.Y
				if item.box.Y < top {
					y -= top - item.box.Y
				} else if item.box.Y+item.box.H > bottom {
					y += item.box.Y + item.box.H - bottom
				}
				y = max(float32(0), min(scroll.MaxY, y))
				if y-scroll.Y > .5 || scroll.Y-y > .5 {
					scroll.Y, s.reveal = y, true
					c.Invalidate()
				}
				break
			}
		}
		if !itemAligned || !s.open {
			return
		}
		if s.measured {
			if !s.aligned {
				return
			}
			_, h := c.Size()
			limit := h - 20
			if maxHeight > 0 {
				limit = min(limit, maxHeight)
			}
			remaining := max(float32(0), limit-s.alignHeight)
			if remaining <= .5 {
				return
			}
			delta := scroll.Y
			if s.alignTopPinned {
				delta = max(float32(0), s.contentHeight-s.alignHeight-scroll.Y)
			}
			growth := min(remaining, delta)
			if s.alignTopPinned {
				growth = min(growth, max(float32(0), h-10-s.alignY-s.alignHeight))
			} else {
				growth = min(growth, max(float32(0), s.alignY-10))
			}
			if growth > .5 {
				s.alignHeight += growth
				if !s.alignTopPinned {
					s.alignY -= growth
					scroll.Y = max(float32(0), scroll.Y-growth)
				}
				c.Invalidate()
			}
			return
		}
		s.measured = true
		w, h := c.Size()
		a, p := s.triggerBox, s.panelBox
		if a.Y < 20 || a.Y+a.H > h-40 || p.H < min(float32(100), s.contentHeight) {
			s.aligned = false
			return
		}
		s.aligned = true
		text := ui.Rect{}
		if len(s.selected) > 0 {
			key := s.selected[len(s.selected)-1]
			for _, g := range s.geometry {
				if g.value == key {
					text = g.text
					break
				}
			}
		}
		if text.H <= 0 && len(s.geometry) > 0 {
			text = s.geometry[0].text
		}
		x := a.X
		if text.W > 0 {
			if c.Direction() == ui.RightToLeft {
				x = p.X + s.valueBox.X + s.valueBox.W - text.X - text.W
			} else {
				x = p.X + s.valueBox.X - text.X
			}
		}
		s.alignPreferredX = x
		s.alignX = max(float32(5), min(w-p.W-5, x))
		// The selected item's measured text center, rather than a fixed row
		// estimate, stays over the value. Bounded overflow uses the native scroll.
		itemOffset := text.Y - p.Y + scroll.Y + text.H/2
		valueCenter := s.valueBox.Y + s.valueBox.H/2
		idealTop := valueCenter - itemOffset
		top := max(float32(10), idealTop)
		height := min(s.contentHeight, h-10-top)
		topLimit := s.contentHeight - max(float32(0), 10+itemOffset-valueCenter)
		s.alignTopPinned = topLimit < height
		height = min(height, topLimit)
		if maxHeight > 0 {
			height = min(height, maxHeight)
		}
		if height+1 < min(float32(100), s.contentHeight) {
			s.aligned = false
			return
		}
		s.alignHeight = max(float32(1), height)
		s.alignY = max(float32(10), min(h-height-10, idealTop))
		if !s.alignedScroll {
			scroll.Y = max(float32(0), min(max(float32(0), s.contentHeight-height), s.alignY+itemOffset-valueCenter))
			s.alignedScroll = true
		}

		c.Invalidate()
	})
}

func (f *selectFrame) finish() {
	defer func() {
		f.alive = false
		f.c, f.state = nil, nil
		f.p = SelectRootProps{}
		f.options, f.items, f.values = nil, nil, nil
		f.trigger, f.value, f.panel, f.list = ui.Element{}, ui.Element{}, ui.Element{}, ui.Element{}
		f.group = ui.Element{}
	}()
	s := f.s
	// Application callbacks may control open directly in the same build, for
	// example closing a multiple list after a domain-specific maximum is met.
	if f.state.Open != s.open {
		if !f.state.Open {
			s.open = false
			s.trigger.Focus()
		}
		f.c.Invalidate()
	}
	if f.manual && f.panel.Valid() {
		ordered := make([]Option, 0, len(f.items))
		for _, item := range f.items {
			for _, o := range f.options {
				if !o.Separator && o.Value == item.value {
					ordered = append(ordered, o)
					break
				}
			}
		}
		f.options = ordered
	}
	f.validateOptions()
	s.options = slices.Clone(f.options)
	f.selected()
	for _, v := range f.values {
		v.text.Children(func() { ui.Text(f.c, f.valueText(v.placeholder)).SingleLine() })
	}
	if s.enabled && s.opening && f.itemAligned && !s.hasActive {
		for _, item := range f.items {
			if !item.disabled || !s.legacySkipDisabled {
				s.active, s.hasActive, s.moveFocus = item.value, true, true
				item.item.Focus()
				break
			}
		}
	}
	for i, item := range f.items {
		active := s.open && s.hasActive && s.active == item.value
		selected := slices.Contains(s.selected, item.value)
		item.item.PositionInSet(i+1, len(f.items)).AccessSelected(selected)
		if active {
			bg, fg := theme.MenuOf(f.c).ForScene().Highlight(false)
			item.item.Background(bg).TextColor(fg)
		}
		if selected {
			item.check.Children(func() { indicator.Check(f.c) })
		}
		if active && s.moveFocus && s.enabled {
			item.item.Focus()
		}
		if active && s.reveal && (!f.itemAligned || s.measured) {
			item.item.ScrollIntoView()
		}
	}

	if f.trigger.Valid() {
		f.trigger.Expanded(s.open).TabStop(!s.open).Value(f.valueText(""))
		if f.changed {
			f.trigger.SetChanged()
		}
	}
	if s.open && f.panel.Valid() && f.panel.FocusLeft() && !f.trigger.FocusWithin() && !f.panel.FocusWithin() {
		f.setOpen(false, "focus-out")
	}
	if !s.open && f.panel.Valid() {
		f.panel.AccessHidden(true).IgnoreInput(true)
		if f.itemAligned || !theme.MenuOf(f.c).Animated() {
			f.panel.Opacity(0)
		}
	}
	s.moveFocus, s.reveal, s.opening = false, false, false
	if !s.open {
		s.arrowDirection = 0
	}
}

// FormValues returns a copy suitable for application-owned form submission.
func (b *SelectScope) FormValues() []string {
	b.f.check()
	if b.f.p.Multiple {
		return slices.Clone(b.f.state.Values)
	}
	if b.f.state.Value.Valid {
		return []string{b.f.state.Value.Value}
	}
	return nil
}
