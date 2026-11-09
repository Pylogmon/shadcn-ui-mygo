package control

import (
	"math"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

// ComboboxValue distinguishes a selected empty string from no selection.
// Values are option identities; their Labels are the text shown to the user.
type ComboboxValue struct {
	Value string
	Valid bool
}

// ComboboxWithProps configures the selection comboboxes. Unlike the legacy
// free-text Combobox, typing only changes query; choosing an option changes the
// selected value. AutoHighlight is opt-in, and arrow navigation loops by default.
type ComboboxWithProps struct {
	Options                       []Option
	Label, Placeholder, EmptyText string
	// PopupInput places the search editor inside the popup, below a button.
	PopupInput                                                bool
	InputPlaceholder                                          string
	Disabled, Invalid, ReadOnly                               bool
	ShowClear, HideTrigger, AutoHighlight, NoLoop, HideRemove bool
	Size                                                      Size
	MaxSelected                                               int
	MaxHeight                                                 float32
	// Locale controls case/diacritic/width-insensitive label matching. A custom
	// Filter receives the unmodified query and can search additional metadata.
	Locale         string
	Filter         func(Option, string) bool
	ItemContent    func(*ui.Context, Option)
	TriggerContent func(*ui.Context, Option)
	OnValueChange  func(ComboboxValue)
	OnValuesChange func([]string)
	OnInputChange  func(string)
	OnOpenChange   func(bool)
}

type comboboxComposition struct{ composing, suppressChoice bool }

func (s *comboboxComposition) observe(ev ui.InputEvent) {
	switch ev.Kind {
	case ui.InputCompose:
		s.composing = ev.Text != ""
	case ui.InputText:
		s.composing = false
	case ui.InputKeyDown:
		s.suppressChoice = s.composing && ev.Key == ui.KeyEnter
	}
}

type comboEntry struct {
	option   Option
	selected bool
}
type comboRuntime struct {
	open, wasOpen, initialized, enabled, multiple, popupInput bool
	autoHighlight, noLoop, composing, filtering, queryEmpty   bool
	active                                                    string
	hasActive, reveal                                         bool
	items                                                     []comboEntry
	pending                                                   string
	hasPending                                                bool
	remove                                                    string
	hasRemove, clear                                          bool
	input, trigger                                            ui.Handle
	chips                                                     map[string]*ui.Handle
	selected                                                  []string
	direction                                                 ui.Direction
	selectionStart, selectionEnd                              int
	previousQuery                                             string
	previousLabel                                             string
	previousValue                                             ComboboxValue
	pointerX, pointerY                                        float32
	pointerOver                                               bool
	closing                                                   bool
	closeQuery                                                string
	closeFiltering                                            bool
	motion                                                    comboMotion
}

func (s *comboRuntime) captureClose(query string) {
	if !s.closing {
		s.closing = true
		s.closeQuery, s.closeFiltering = query, s.filtering
	}
}

func (s *comboRuntime) toggle(query string) {
	if s.open {
		s.captureClose(query)
		s.open = false
	} else {
		s.open = true
		if !s.closing {
			s.filtering = false
		}
	}
	s.hasActive = false
}

func (s *comboRuntime) eligible(i int) bool {
	return i >= 0 && i < len(s.items) && !s.items[i].option.Disabled
}
func (s *comboRuntime) setActive(i int) {
	if !s.eligible(i) {
		s.hasActive = false
		return
	}
	if !s.hasActive || s.active != s.items[i].option.Value {
		s.reveal = true
	}
	s.active, s.hasActive = s.items[i].option.Value, true
}
func (s *comboRuntime) move(delta int) {
	i := -1
	if delta < 0 {
		i = len(s.items)
	}
	if s.hasActive {
		for n, e := range s.items {
			if e.option.Value == s.active {
				i = n
				break
			}
		}
	}
	for range len(s.items) {
		i += delta
		if i < 0 || i >= len(s.items) {
			if s.noLoop {
				return
			}
			i = (i + len(s.items)) % len(s.items)
		}
		if s.eligible(i) {
			s.setActive(i)
			return
		}
	}
}
func (s *comboRuntime) edge(last bool) {
	i, d := 0, 1
	if last {
		i, d = len(s.items)-1, -1
	}
	for ; i >= 0 && i < len(s.items); i += d {
		if s.eligible(i) {
			s.setActive(i)
			return
		}
	}
}
func (s *comboRuntime) observe(ev ui.InputEvent) {
	switch ev.Kind {
	case ui.InputCompose:
		s.composing = ev.Text != ""
	case ui.InputText:
		s.composing = false
	}
}
func (s *comboRuntime) inputEvent(ev ui.InputEvent) bool {
	if !s.enabled || s.composing || ev.Kind != ui.InputKeyDown || ev.Mods != 0 {
		return false
	}
	switch ev.Key {
	case ui.KeyDown, ui.KeyUp:
		s.open = true
		d := 1
		if ev.Key == ui.KeyUp {
			d = -1
		}
		s.move(d)
		return true
	case ui.KeyEnter:
		if s.open {
			if s.hasActive {
				s.pending, s.hasPending = s.active, true
				return true
			}
			// An unhighlighted Enter dismisses suggestions but remains a
			// native editor submission gesture for the enclosing form.
			s.open = false
			if s.popupInput {
				s.trigger.Focus()
			}
		}
	case ui.KeyEscape:
		if s.open {
			s.open = false
			s.hasActive = false
			if s.popupInput {
				s.trigger.Focus()
			}
			return true
		}
		if len(s.selected) > 0 {
			s.clear = true
			return true
		}
	case ui.KeyBackspace:
		if s.multiple && s.queryEmpty && len(s.selected) > 0 {
			s.remove, s.hasRemove = s.selected[len(s.selected)-1], true
			return true
		}
	case ui.KeyLeft, ui.KeyRight:
		previous := ui.KeyLeft
		if s.direction == ui.RightToLeft {
			previous = ui.KeyRight
		}
		if s.multiple && ev.Key == previous && s.selectionStart == 0 && s.selectionEnd == 0 && len(s.selected) > 0 {
			if h := s.chips[s.selected[len(s.selected)-1]]; h != nil {
				h.Focus()
				return true
			}
		}
	}
	return false
}
func (s *comboRuntime) chipEvent(value string, ev ui.InputEvent) bool {
	if !s.enabled || ev.Kind != ui.InputKeyDown || ev.Mods != 0 {
		return false
	}
	i := slices.Index(s.selected, value)
	if i < 0 {
		return false
	}
	previous, next := ui.KeyLeft, ui.KeyRight
	if s.direction == ui.RightToLeft {
		previous, next = next, previous
	}
	switch ev.Key {
	case previous, next:
		j := i + 1
		if ev.Key == previous {
			j = i - 1
		}
		if j < 0 || j >= len(s.selected) {
			s.input.Focus()
		} else if h := s.chips[s.selected[j]]; h != nil {
			h.Focus()
		}
		return true
	case ui.KeyBackspace, ui.KeyDelete:
		s.remove, s.hasRemove = value, true
		j := i
		if j == len(s.selected)-1 {
			j--
		}
		if j < 0 {
			s.input.Focus()
		} else {
			if j >= i {
				j++
			}
			if h := s.chips[s.selected[j]]; h != nil {
				h.Focus()
			}
		}
		return true
	case ui.KeyEnter, ui.KeySpace:
		s.input.Focus()
		return true
	case ui.KeyDown, ui.KeyUp:
		s.open = true
		s.input.Focus()
		return true
	}
	if ev.Key >= ui.KeyA && ev.Key <= ui.KeyZ {
		s.input.Focus()
	}
	return false
}
func comboLabel(options []Option, value string) string {
	for _, o := range options {
		if !o.Separator && o.Value == value {
			return optionLabel(o)
		}
	}
	return value
}
func comboEntries(p ComboboxWithProps, q string, filtering bool, selected []string) []comboEntry {
	out := make([]comboEntry, 0, len(p.Options))
	seen := map[string]bool{}
	var pattern *search.Pattern
	if filtering && q != "" && p.Filter == nil {
		pattern = search.New(language.Make(p.Locale), search.Loose).CompileString(q)
	}
	for _, o := range p.Options {
		if o.Separator {
			out = append(out, comboEntry{option: o})
			continue
		}
		if seen[o.Value] {
			continue
		}
		seen[o.Value] = true
		if filtering && q != "" {
			if p.Filter != nil {
				if !p.Filter(o, q) {
					continue
				}
			} else if start, _ := pattern.IndexString(optionLabel(o)); start < 0 {
				continue
			}
		}
		selectedItem := slices.Contains(selected, o.Value)
		if p.MaxSelected > 0 && len(selected) >= p.MaxSelected && !selectedItem {
			o.Disabled = true
		}
		out = append(out, comboEntry{o, selectedItem})
	}
	return out
}

// ComboboxWith chooses an option Value independently of the native editor's
// query. A choice displays its Label; dismissal restores that label, while a
// popup-contained editor is cleared. Programmatic changes do not emit callbacks.
func ComboboxWith(c *ui.Context, selected *ComboboxValue, query *string, p ComboboxWithProps) ui.Element {
	if selected == nil || query == nil {
		panic("control.ComboboxWith: selected and query must not be nil")
	}
	values := []string(nil)
	if selected.Valid {
		values = []string{selected.Value}
	}
	return selectionCombobox(c, selected, &values, query, p, false)
}

// MultiComboboxWith toggles selected Values and keeps them visible in the
// list. Chips share the native input's wrapping field; PopupInput instead uses
// a button summarizing the chosen labels. A filtered choice closes
// an external-input popup; an unfiltered choice keeps it open.
func MultiComboboxWith(c *ui.Context, selected *[]string, query *string, p ComboboxWithProps) ui.Element {
	if selected == nil || query == nil {
		panic("control.MultiComboboxWith: selected and query must not be nil")
	}
	return selectionCombobox(c, nil, selected, query, p, true)
}

func selectionCombobox(c *ui.Context, single *ComboboxValue, selected *[]string, query *string, p ComboboxWithProps, multiple bool) ui.Element {
	t := theme.TokensOf(c)
	var root ui.Element
	if p.PopupInput {
		root = ui.ButtonBase(c)
	} else {
		root = ui.Row(c)
	}
	root.FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(6).Radius(t.RadiusMD()).Disabled(p.Disabled).DimDisabled(false).TextColor(t.Foreground)
	var parentGroup *groupBuild
	beforeGroupItems := 0
	if scope := groupScope(c); len(scope.active) > 0 {
		parentGroup = scope.active[len(scope.active)-1]
		beforeGroupItems = len(parentGroup.items)
	}
	s := ui.Local(root, "shadcn.control.selection-combobox", func() comboRuntime { return comboRuntime{chips: make(map[string]*ui.Handle)} })
	s.enabled, s.multiple, s.popupInput = !root.IsDisabled() && !p.ReadOnly, multiple, p.PopupInput
	s.autoHighlight, s.noLoop, s.direction = p.AutoHighlight, p.NoLoop, c.Direction()
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	groupAppearance := ui.Local(c.Root(), inputGroupAppearanceKey{}, func() inputGroupAppearance { return inputGroupAppearance{} })
	inInputGroup := len(groupAppearance.active) > 0
	for _, v := range groupAppearance.active {
		*v = *v || p.Invalid
	}
	root.Invalid(p.Invalid).Cursor(ui.CursorText)
	if t.Dark {
		root.Background(t.Input.Alpha(.3))
	} else {
		root.Background(ui.Transparent)
	}
	motion.ApplyTransition(c, root)
	textInputAppearance(c, root, p.Invalid)
	if root.IsDisabled() {
		root.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	if p.PopupInput {
		root.Background(ui.Transparent).Border(0, ui.Transparent).NoShadow()
	}
	handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	height, _ := metrics(p.Size)
	scale := c.Preferences().TextScale
	if scale <= 0 || math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) {
		scale = 1
	}
	height = max(height, inputFont(c, p.Size)*inputLineHeight(p.Size)*scale+2)
	inputHeight := height - 2
	if multiple {
		height = max(height, 36)
		inputHeight = max(float32(20), 20*scale)
	}
	root.MinHeight(height)
	if !s.initialized {
		if !multiple && single.Valid && *query == "" && !p.PopupInput {
			*query = comboLabel(p.Options, single.Value)
		}
		s.previousQuery = *query
		s.initialized = true
	}
	// External updates reconcile the displayed label without publishing user events.
	if !multiple && (*single != s.previousValue || (single.Valid && comboLabel(p.Options, single.Value) != s.previousLabel)) && !s.open && !p.PopupInput {
		if single.Valid {
			*query = comboLabel(p.Options, single.Value)
		} else {
			*query = ""
		}
	}
	if !s.enabled {
		s.open = false
		s.hasActive = false
		s.composing = false
	}
	// Logical dismissal removes input and accessibility immediately; native
	// editors and the filtered result stay mounted until their exit finishes.
	syncMotion := func() (comboPose, bool) {
		if s.wasOpen && !s.open {
			s.captureClose(*query)
		}
		if s.open {
			s.closing = false
		}
		pose, present := s.motion.update(c, s.open)
		if s.closing {
			if multiple && !p.PopupInput {
				*query = ""
			}
			if !present {
				s.closing, s.filtering, s.hasActive = false, false, false
				if multiple || p.PopupInput || !single.Valid {
					*query = ""
				} else {
					*query = comboLabel(p.Options, single.Value)
				}
			}
		}
		return pose, present
	}
	syncMotion()
	setQuery := func(q string) {
		if *query != q {
			*query = q
			if p.OnInputChange != nil {
				p.OnInputChange(q)
			}
		}
	}
	change := func(values []string) {
		if multiple {
			if !slices.Equal(*selected, values) {
				*selected = slices.Clone(values)
				root.SetChanged()
				if p.OnValuesChange != nil {
					p.OnValuesChange(slices.Clone(values))
				}
			}
		} else {
			v := ComboboxValue{}
			if len(values) > 0 {
				v = ComboboxValue{values[0], true}
			}
			if *single != v {
				*single = v
				*selected = slices.Clone(values)
				root.SetChanged()
				if p.OnValueChange != nil {
					p.OnValueChange(v)
				}
			}
		}
	}
	remove := func(value string) {
		values := slices.DeleteFunc(slices.Clone(*selected), func(v string) bool { return v == value })
		change(values)
	}
	if s.clear {
		s.clear = false
		change(nil)
		setQuery("")
		s.hasActive = false
		s.filtering = false
		s.input.Focus()
	}
	if s.hasRemove {
		s.hasRemove = false
		remove(s.remove)
	}
	s.selected = slices.Clone(*selected)
	s.items = comboEntries(p, *query, s.filtering, *selected)
	// Separators take space but are not eligible navigation destinations.
	for i := range s.items {
		if s.items[i].option.Separator {
			s.items[i].option.Disabled = true
		}
	}
	if s.previousQuery != *query {
		s.hasActive = false
		if s.filtering && p.AutoHighlight {
			s.edge(false)
		}
	}
	if s.hasActive {
		valid := false
		for i, e := range s.items {
			if e.option.Value == s.active && s.eligible(i) {
				valid = true
				break
			}
		}
		if !valid {
			s.hasActive = false
		}
	}
	var choice string
	hasChoice := s.hasPending
	s.hasPending = false
	choice = s.pending
	choose := func(value string) {
		if !s.enabled {
			return
		}
		var option Option
		found := false
		for _, e := range s.items {
			if !e.option.Separator && !e.option.Disabled && e.option.Value == value {
				option = e.option
				found = true
				break
			}
		}
		if !found {
			return
		}
		filtered := s.filtering && strings.TrimSpace(*query) != ""
		if !multiple || filtered && !p.PopupInput {
			s.captureClose(*query)
		}
		if multiple {
			values := slices.Clone(*selected)
			if slices.Contains(values, value) {
				values = slices.DeleteFunc(values, func(v string) bool { return v == value })
			} else {
				values = append(values, value)
			}
			change(values)
			if filtered && !p.PopupInput {
				s.open = false
			}
			if filtered {
				setQuery("")
			}
		} else {
			change([]string{value})
			if !p.PopupInput {
				setQuery(optionLabel(option))
			}
			s.open = false
		}
		s.filtering = false
		s.hasActive = false
		if p.PopupInput {
			s.trigger.Focus()
		} else {
			s.input.Focus()
		}
	}
	if hasChoice {
		choose(choice)
	}
	buildInput := func(ctx *ui.Context, inPopup bool) {
		input := styleInput(ctx, ui.TextInputBase(ctx.Key("editor"), query), InputProps{Label: p.Label, Placeholder: p.Placeholder, Disabled: p.Disabled, Invalid: p.Invalid, ReadOnly: p.ReadOnly, Size: p.Size}).Bind(&s.input).
			Role(ui.RoleComboBox).HasPopup(ui.PopupListBox).AutoComplete(ui.AutoCompleteList).Expanded(s.open).Grow(1).MinWidth(0).
			Border(0, ui.Transparent).Radius(0).Background(ui.Transparent).FocusRing(false).NoShadow().Padding(0).Height(inputHeight).
			HandleInput(s.inputEvent).ObserveInput(s.observe)
		if !inPopup {
			input.Bind(handle)
		}
		if multiple && !inPopup {
			input.MinWidth(64).FontSize(14).LineHeight(20.0 / 14)
		}
		if inPopup {
			placeholder := p.InputPlaceholder
			if placeholder == "" {
				placeholder = "Search…"
			}
			input.Placeholder(placeholder).Padding(4, 8).Height(max(float32(30), 20*scale+10)).FontSize(14).LineHeight(20.0 / 14)
		}

		input.OnChange(func() {
			s.open = true
			s.filtering = true
			if s.previousQuery != *query {
				s.hasActive = false
			}
			if p.OnInputChange != nil {
				p.OnInputChange(*query)
			}
		})
		if input.Clicked() && s.enabled {
			s.open = true
		}
		s.composing = input.Composing()
		s.queryEmpty = *query == ""
		s.selectionStart, s.selectionEnd = input.TextSelection()
	}
	if p.PopupInput {
		StyleButton(c, root, "", ButtonProps{Variant: Outline, Disabled: p.Disabled || p.ReadOnly, Invalid: p.Invalid, Label: p.Label, Size: p.Size})
	}
	root.Children(func() {
		if p.PopupInput {
			label := p.Placeholder
			var selectedOption Option
			if !multiple && single.Valid {
				label = comboLabel(p.Options, single.Value)
				for _, o := range p.Options {
					if !o.Separator && o.Value == single.Value {
						selectedOption = o
						break
					}
				}
			}
			if multiple && len(*selected) > 0 {
				labels := make([]string, 0, len(*selected))
				for _, v := range *selected {
					labels = append(labels, comboLabel(p.Options, v))
				}
				label = strings.Join(labels, ", ")
			}
			if label == "" {
				label = "Select an option"
			}
			trigger := root.Bind(&s.trigger).FillWidth().Gap(8).Expanded(s.open).HasPopup(ui.PopupListBox)
			trigger.Bind(handle)
			trigger.Children(func() {
				if p.TriggerContent != nil && !multiple && single.Valid {
					ui.Row(c).Grow(1).MinWidth(0).Children(func() { p.TriggerContent(c, selectedOption) })
				} else {
					ui.Text(c, label).Grow(1).SingleLine()
				}
				indicator.ChevronDown(c).TextColor(t.MutedForeground)
			})
			if trigger.Clicked() && s.enabled {
				s.toggle(*query)
			}
			trigger.HandleInput(func(ev ui.InputEvent) bool {
				if s.enabled && ev.Kind == ui.InputKeyDown && ev.Mods == 0 && (ev.Key == ui.KeyDown || ev.Key == ui.KeyUp) {
					s.open = true
					s.hasActive = false
					return true
				}
				return false
			})
			return
		}
		root.Padding(0, 10)
		if multiple {
			root.Wrap().Padding(6, 10)
		}
		if multiple {
			live := make(map[string]bool)
			for _, value := range slices.Clone(*selected) {
				if live[value] {
					continue
				}
				live[value] = true
				if s.chips[value] == nil {
					s.chips[value] = &ui.Handle{}
				}
				value := value
				label := comboLabel(p.Options, value)
				chip := ui.Row(c.Key(choiceOptionKey{value: value})).Height(22).Padding(2, 6).Gap(4).AlignItems(ui.Center).Radius(t.RadiusSM()).Background(t.Muted).TextColor(t.Foreground).FontSize(12).FontWeight(500).Focusable().TabStop(false).Label(label).Bind(s.chips[value]).HandleInput(func(ev ui.InputEvent) bool { return s.chipEvent(value, ev) })
				chip.Children(func() {
					ui.Text(c, label).SingleLine()
					if !p.HideRemove {
						removeButton := ui.ButtonBase(c).Size(16, 16).KeepFocus().TabStop(false).Label("Remove " + label).Disabled(!s.enabled).FocusRing(false).Children(func() { indicator.Close(c).Size(12, 12).FontSize(12) })
						if removeButton.Clicked() {
							remove(value)
							s.input.Focus()
						}
					}
				})
			}
			for value := range s.chips {
				if !live[value] {
					delete(s.chips, value)
				}
			}
		}
		buildInput(c, false)
		visibleClear := p.ShowClear && len(*selected) > 0
		if visibleClear {
			clear := ui.ButtonBase(c.Key("clear")).Size(min(float32(24), height-2), min(float32(24), height-2)).Shrink(0).KeepFocus().TabStop(false).Label("Clear " + p.Label).Disabled(!s.enabled).FocusRing(false).Children(func() { indicator.Close(c) })
			if clear.Clicked() {
				change(nil)
				setQuery("")
				s.hasActive = false
				s.input.Focus()
			}
		} else if !p.HideTrigger {
			trigger := ui.ButtonBase(c.Key("trigger")).Size(min(float32(24), height-2), min(float32(24), height-2)).Shrink(0).KeepFocus().TabStop(false).Label("Show " + p.Label + " options").Disabled(!s.enabled).FocusRing(false).Children(func() { indicator.ChevronDown(c).TextColor(t.MutedForeground) })
			if trigger.Clicked() {
				s.toggle(*query)
				s.input.Focus()
			}
		}
	})
	if !p.PopupInput {
		if input := c.Resolve(s.input); input.Valid() && input.FocusVisible() && !p.Invalid && !inInputGroup {
			root.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
		}
		root.HandleInput(func(ev ui.InputEvent) bool {
			if s.enabled && ev.Kind == ui.InputPointerDown && ev.Button == 0 {
				s.input.Focus()
			}
			return false
		})
	}
	if inInputGroup {
		root.Border(0, ui.Transparent).Background(ui.Transparent).NoShadow()
	}
	s.items = comboEntries(p, *query, s.filtering, *selected)
	for i := range s.items {
		if s.items[i].option.Separator {
			s.items[i].option.Disabled = true
		}
	}
	if s.previousQuery != *query {
		s.hasActive = false
		if s.filtering && p.AutoHighlight {
			s.edge(false)
		}
	}
	if s.filtering && p.AutoHighlight && !s.hasActive {
		s.edge(false)
	}
	// The attached popup participates in FocusLeft and native light dismissal.
	if root.FocusLeft() {
		s.open = false
		s.composing = false
	}
	pose, present := syncMotion()
	if s.closing {
		s.items = comboEntries(p, s.closeQuery, s.closeFiltering, *selected)
		for i := range s.items {
			if s.items[i].option.Separator {
				s.items[i].option.Disabled = true
			}
		}
	}
	maxHeight := p.MaxHeight
	if maxHeight <= 0 || math.IsNaN(float64(maxHeight)) || math.IsInf(float64(maxHeight), 0) {
		maxHeight = 288
	}
	anchor := root
	if p.PopupInput {
		if trigger := c.Resolve(s.trigger); trigger.Valid() {
			anchor = trigger
		}
	}
	var panel ui.Element
	dismissed, outside := false, false
	if present {
		ui.Overlay(c, func() {
			panel = ui.Box(c).Role(ui.RolePopup).AttachTo(anchor, ui.AnchorBottomStart, ui.AnchorTopStart)
			panel.Key(s)
			panel.Clicked() // The panel background participates in native hit testing.
			panel.MinWidth(max(float32(144), root.Bounds().W)).MaxHeight(maxHeight).Margin(6, 0, 0, 0).Clip().Radius(t.RadiusMD()).Border(1, t.Foreground.Alpha(.1)).Background(t.Popover).TextColor(t.PopoverForeground).Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).IgnoreInput(!s.open).AccessHidden(!s.open)
			if !p.PopupInput {
				panel.KeepFocus()
			}
			panel.Children(func() {
				if p.PopupInput {
					ui.Row(c.Key("popup-search")).FillWidth().Margin(4, 4, 0, 4).Height(max(float32(32), 20*scale+12)).AlignItems(ui.Center).Radius(t.RadiusMD()).Background(t.Input.Alpha(.3)).Border(1, t.Input.Alpha(.3)).Children(func() {
						buildInput(c, true)
						if p.ShowClear && len(*selected) > 0 {
							clear := ui.ButtonBase(c.Key("clear")).Size(24, 24).MarginEnd(4).Shrink(0).KeepFocus().TabStop(false).Label("Clear " + p.Label).Disabled(!s.enabled).FocusRing(false).Children(func() { indicator.Close(c) })
							if clear.Clicked() {
								change(nil)
								setQuery("")
								s.hasActive = false
								s.filtering = false
								s.input.Focus()
							}
						}
					})
				}
				list := ui.Scroll(c.Key("list")).TabStop(false).FillWidth().Padding(4).MaxHeight(max(float32(1), maxHeight-36)).Role(ui.RoleListBox).Multiselectable(multiple).Label(p.Label + " options")
				if input := c.Resolve(s.input); input.Valid() {
					input.Controls(list).Expanded(s.open)
				}
				x, y, over := list.PointerPosition()
				pointerMoved := over && (!s.pointerOver || s.pointerX != x || s.pointerY != y)
				s.pointerX, s.pointerY, s.pointerOver = x, y, over
				list.Children(func() {
					n := 0
					for _, e := range s.items {
						if !e.option.Separator {
							n++
						}
					}
					if n == 0 {
						empty := p.EmptyText
						if empty == "" {
							empty = "No items found."
						}
						ui.Text(c, empty).Padding(8, 10).FontSize(14).TextColor(t.MutedForeground).Role(ui.RoleStatus)
					}
					pos := 0
					renderOption := func(entry comboEntry) {
						o := entry.option
						pos++
						option := ui.ButtonBase(c.Key(choiceOptionKey{value: o.Value})).Role(ui.RoleOption).KeepFocus().TabStop(false).Disabled(o.Disabled || !s.enabled).DimDisabled(false).Label(optionLabel(o)).PositionInSet(pos, n).AccessSelected(entry.selected).FillWidth().MinWidth(0).MinHeight(32).Padding(6, 32, 6, 8).Gap(8).AlignItems(ui.Center).Radius(t.RadiusSM()).FontSize(14).FocusRing(false)
						active := s.hasActive && s.active == o.Value
						if pointerMoved && option.Hovered() && !o.Disabled && s.enabled {
							s.active, s.hasActive = o.Value, true
							active = true
						}
						if active {
							option.Background(t.Accent).TextColor(t.AccentForeground)
							if input := c.Resolve(s.input); input.Valid() {
								input.ActiveDescendant(option)
							}
							if s.reveal {
								option.ScrollIntoView()
								s.reveal = false
							}
						}
						if o.Disabled {
							option.Opacity(.5)
						}
						option.Children(func() {
							ui.Row(c).Grow(1).MinWidth(0).Children(func() {
								if p.ItemContent != nil {
									p.ItemContent(c, o)
								} else {
									ui.Text(c, optionLabel(o)).SingleLine()
								}
							})
							if entry.selected {
								ui.Box(c).Absolute().Right(8).TopPercent(50).Size(16, 16).Translate(0, -8).Children(func() { indicator.Check(c) })
							}
						})
						if option.Clicked() {
							choose(o.Value)
						}
					}
					for i := 0; i < len(s.items); {
						if s.items[i].option.Separator {
							ui.Box(c).Height(1).FillWidth().Margin(4, 0).Background(t.Border).Role(ui.RoleSeparator).Orientation(ui.Horizontal)
							i++
							continue
						}
						start, end, name := i, i+1, s.items[i].option.Group
						for end < len(s.items) && !s.items[end].option.Separator && s.items[end].option.Group == name {
							end++
						}
						if name != "" {
							ui.Column(c.Key(struct {
								Name  string
								Start int
							}{name, start})).FillWidth().Role(ui.RoleGroup).Label(name).Children(func() {
								optionGroup(c, name)
								for j := start; j < end; j++ {
									renderOption(s.items[j])
								}
							})
						} else {
							for j := start; j < end; j++ {
								renderOption(s.items[j])
							}
						}
						i = end
					}
				})
			})
			if s.open {
				outside = panel.PressedOutside()
				if panel.OverlayShortcut(0, ui.KeyEscape) || outside {
					s.captureClose(*query)
					s.open, dismissed = false, true
				}
			}
		})
	}
	if panel.Valid() {
		s.motion.observe(anchor, panel)
	}
	pose, _ = syncMotion()
	if panel.Valid() {
		panel.IgnoreInput(!s.open).AccessHidden(!s.open)
		s.motion.apply(c, anchor, panel, pose)
	}
	if p.PopupInput && s.open && !s.wasOpen {
		if input := c.Resolve(s.input); input.Valid() && !input.IsDisabled() {
			input.Focus()
		}
	}
	if p.PopupInput && s.wasOpen && !s.open && dismissed && !outside {
		s.trigger.Focus()
	}
	if s.wasOpen != s.open && p.OnOpenChange != nil {
		p.OnOpenChange(s.open)
	}
	s.wasOpen = s.open
	s.previousQuery = *query
	if !multiple {
		s.previousValue = *single
		s.previousLabel = comboLabel(p.Options, single.Value)
	}
	if parentGroup != nil {
		parentGroup.items = parentGroup.items[:beforeGroupItems]
	}
	return ButtonGroupItem(c, root)
}
