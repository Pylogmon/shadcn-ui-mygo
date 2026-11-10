package control

import (
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// ComboboxValue distinguishes a selected empty string from no selection.
// Values are option identities; their Labels are the text shown to the user.
type ComboboxValue struct {
	Value string
	Valid bool
}

// ComboboxWithProps configures the selection comboboxes. Unlike the legacy
// free-text Combobox, typing edits the query and choosing commits an option.
// Clearing a single external editor also clears its selected value, matching
// Base UI. AutoHighlight is opt-in; manual loops include the unhighlighted input.
type ComboboxWithProps struct {
	Options                       []Option
	Label, Placeholder, EmptyText string
	// PopupInput uses a native button with combobox/dialog accessibility semantics
	// and places a list-autocomplete combobox editor inside its nonmodal dialog.
	PopupInput                                                bool
	InputPlaceholder                                          string
	Disabled, Invalid, ReadOnly                               bool
	ShowClear, HideTrigger, AutoHighlight, NoLoop, HideRemove bool
	Size                                                      Size
	MaxSelected                                               int
	MaxHeight                                                 float32
	// Locale controls source search collation, including case, diacritics,
	// width and ignored punctuation within fixed UTF-16 query windows. A custom
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
	previousLabel                                             string
	previousQuery                                             string
	previousValue                                             ComboboxValue
	pointerOver                                               bool
	closing                                                   bool
	closeQuery                                                string
	closeFiltering                                            bool
	motion                                                    comboMotion
	committedQuery, previousFilterQuery                       string
	compositionCommit                                         bool
	pointerHighlight                                          bool
	pointerMoved, pointerTouch                                bool
	typeahead, typeaheadText                                  string
	typeaheadAt                                               time.Time
	typeaheadPrevious, typeaheadMatch                         int
	typeaheadIME                                              bool
	typeaheadValue                                            string
	hasTypeaheadValue                                         bool
	closed, boundOpenInitialized, lastBoundOpen               bool
	lastActive                                                string
	lastHasActive, completeInitialized, completeOpen          bool
	grid                                                      bool
	gridColumns                                               int
	manualOptions                                             []Option
	slotDisabled                                              map[string]bool
	gridCells                                                 []comboGridCell
}

func (s *comboRuntime) filterQuery(query string) string {
	if s.composing || s.compositionCommit {
		return s.committedQuery
	}
	return query
}

func (s *comboRuntime) highlightSelected() bool {
	if len(s.selected) == 0 {
		return false
	}
	value := s.selected[len(s.selected)-1]
	for _, e := range s.items {
		if !e.option.Separator && e.option.Value == value {
			s.active, s.hasActive, s.reveal, s.pointerHighlight = value, true, true, false
			return true
		}
	}
	return false
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
	s.pointerHighlight = false
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
			if s.hasActive && !s.autoHighlight {
				s.hasActive, s.pointerHighlight = false, false
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
	if ev.Kind == ui.InputKeyDown || ev.Kind == ui.InputText {
		s.pointerHighlight = false
	}
	switch ev.Kind {
	case ui.InputCompose:
		s.composing = ev.Text != ""
		if s.enabled && strings.TrimSpace(ev.Text) != "" {
			s.open = true
		}
	case ui.InputText:
		s.compositionCommit = s.composing
		s.composing = false
	}
}

func (s *comboRuntime) pointerEvent(ev ui.InputEvent) bool {
	if ev.Kind == ui.InputPointerMove {
		s.pointerMoved = true
		s.pointerTouch = ev.Pointer == ui.PointerTouch
	}
	return false
}
func (s *comboRuntime) inputEvent(ev ui.InputEvent) bool {
	if !s.enabled || s.composing || ev.Kind != ui.InputKeyDown || ev.Mods != 0 {
		return false
	}
	if s.gridMove(ev.Key) {
		return true
	}
	switch ev.Key {
	case ui.KeyDown, ui.KeyUp:
		d := 1
		if ev.Key == ui.KeyUp {
			d = -1
		}
		if !s.open {
			s.open = true
			if s.highlightSelected() {
				return true
			}
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
	if !s.enabled || ev.Kind != ui.InputKeyDown {
		return false
	}
	if ev.Mods&^ui.Shift == 0 && ev.Key != ui.KeySpace && comboPrintableKey(ev) {
		s.input.Focus()
		return false // Native text, composition, selection and undo stay in the editor.
	}
	if ev.Mods != 0 {
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
	return false
}

func comboPrintableKey(ev ui.InputEvent) bool {
	if ev.Key >= ui.KeyA && ev.Key <= ui.KeyZ || ev.Key >= ui.Key0 && ev.Key <= ui.Key9 {
		return true
	}
	switch ev.Key {
	case ui.KeyUnknown:
		return ev.Code == ui.CodeUnknown
	case ui.KeySpace, ui.KeyMinus, ui.KeyEqual, ui.KeyComma, ui.KeyPeriod, ui.KeySlash, ui.KeySemicolon, ui.KeyQuote, ui.KeyBracketLeft, ui.KeyBracketRight, ui.KeyBackslash, ui.KeyBackquote:
		return true
	}
	return false
}

func comboKeyText(ev ui.InputEvent) string {
	if ev.Key >= ui.KeyA && ev.Key <= ui.KeyZ {
		return string(rune('a' + ev.Key - ui.KeyA))
	}
	if ev.Key >= ui.Key0 && ev.Key <= ui.Key9 {
		if ev.Mods&ui.Shift != 0 {
			return string(")!@#$%^&*("[ev.Key-ui.Key0])
		}
		return string(rune('0' + ev.Key - ui.Key0))
	}
	plain, shifted := "", ""
	switch ev.Key {
	case ui.KeySpace:
		plain, shifted = " ", " "
	case ui.KeyMinus:
		plain, shifted = "-", "_"
	case ui.KeyEqual:
		plain, shifted = "=", "+"
	case ui.KeyComma:
		plain, shifted = ",", "<"
	case ui.KeyPeriod:
		plain, shifted = ".", ">"
	case ui.KeySlash:
		plain, shifted = "/", "?"
	case ui.KeySemicolon:
		plain, shifted = ";", ":"
	case ui.KeyQuote:
		plain, shifted = "'", "\""
	case ui.KeyBracketLeft:
		plain, shifted = "[", "{"
	case ui.KeyBracketRight:
		plain, shifted = "]", "}"
	case ui.KeyBackslash:
		plain, shifted = "\\", "|"
	case ui.KeyBackquote:
		plain, shifted = "`", "~"
	}
	if ev.Mods&ui.Shift != 0 {
		return shifted
	}
	return plain
}

// The closed single-value trigger has Base's 750ms startsWith session. Unlike
// open list navigation, Base does not pass disabledIndices to this typeahead.
func (s *comboRuntime) typeaheadInput(text string, now time.Time) {
	if now.Sub(s.typeaheadAt) >= 750*time.Millisecond {
		s.typeahead = ""
	}
	items := make([]Option, 0, len(s.items))
	for _, e := range s.items {
		if !e.option.Separator {
			items = append(items, e.option)
		}
	}
	for _, ch := range text {
		if !unicode.IsPrint(ch) {
			continue
		}
		key := strings.ToLower(string(ch))
		fresh := s.typeahead == ""
		if fresh {
			s.typeaheadPrevious = -1
			for i, o := range items {
				if len(s.selected) > 0 && o.Value == s.selected[0] {
					s.typeaheadPrevious = i
					break
				}
			}
		}
		rapid := true
		for _, o := range items {
			rs := []rune(strings.ToLower(optionLabel(o)))
			if len(rs) > 1 && rs[0] == rs[1] {
				rapid = false
				break
			}
		}
		if rapid && s.typeahead == key {
			s.typeahead = ""
			s.typeaheadPrevious = s.typeaheadMatch
		}
		s.typeahead += key
		s.typeaheadAt = now
		found := -1
		for n := range len(items) {
			i := (s.typeaheadPrevious + 1 + n + len(items)) % len(items)
			if strings.HasPrefix(strings.ToLower(optionLabel(items[i])), s.typeahead) {
				found = i
				break
			}
		}
		if found >= 0 {
			s.typeaheadMatch = found
			s.typeaheadValue, s.hasTypeaheadValue = items[found].Value, true
		} else if key != " " {
			s.typeahead = ""
		}
	}
}

func (s *comboRuntime) triggerEvent(ev ui.InputEvent) bool {
	if !s.enabled {
		return false
	}
	if ev.Kind == ui.InputCompose {
		s.typeaheadIME = ev.Text != ""
		return false
	}
	if ev.Kind == ui.InputKeyUp {
		s.typeaheadText = ""
		return false
	}
	if ev.Kind == ui.InputKeyDown && ev.Mods == 0 && s.gridMove(ev.Key) {
		return true
	}
	if ev.Kind == ui.InputKeyDown && ev.Mods == 0 && (ev.Key == ui.KeyDown || ev.Key == ui.KeyUp) {
		d := 1
		if ev.Key == ui.KeyUp {
			d = -1
		}
		if !s.open {
			s.open = true
			if s.highlightSelected() {
				return true
			}
		}
		s.move(d)
		s.input.Focus()
		return true
	}
	if s.open || s.multiple || ev.Mods&(ui.Ctrl|ui.Super|ui.Alt) != 0 {
		return false
	}
	if ev.Time.Sub(s.typeaheadAt) >= 750*time.Millisecond {
		s.typeahead = ""
	}
	if ev.Kind == ui.InputText {
		if s.typeaheadIME {
			s.typeaheadIME = false
			return false
		}
		if strings.EqualFold(ev.Text, s.typeaheadText) {
			s.typeaheadText = ""
			return true
		}
		s.typeaheadInput(ev.Text, ev.Time)
		return ev.Text != ""
	}
	if ev.Kind == ui.InputKeyDown && !s.typeaheadIME {
		key := comboKeyText(ev)
		if key == "" || key == " " && s.typeahead == "" {
			return false
		}
		s.typeaheadInput(key, ev.Time)
		s.typeaheadText = key
		return true
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
	var filter comboLabelFilter
	if p.Filter == nil {
		q = comboTrimQuery(q)
	}
	if filtering && q != "" && p.Filter == nil {
		filter = newComboLabelFilter(q, p.Locale)
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
			} else if !filter.contains(optionLabel(o)) {
				continue
			}
		}
		selectedItem := slices.Contains(selected, o.Value)
		if p.MaxSelected > 0 && len(selected) >= p.MaxSelected && !selectedItem {
			o.Disabled = true
		}
		out = append(out, comboEntry{o, selectedItem})
	}
	if filtering {
		// Separators are collection boundaries, not surviving results. Empty
		// groups disappear with their children; retain one rule only between
		// two remaining runs, never before/after the filtered collection.
		filtered := out[:0]
		pending := false
		for _, e := range out {
			if e.option.Separator {
				pending = len(filtered) > 0
				continue
			}
			if pending {
				filtered = append(filtered, comboEntry{option: Option{Separator: true, Disabled: true}})
				pending = false
			}
			filtered = append(filtered, e)
		}
		out = filtered
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
	f := newComboFrame(c, query, single, selected, query, nil, p, multiple)
	defer f.finish()
	var root ui.Element
	if p.PopupInput {
		root = f.trigger(c, ComboboxTriggerProps{Label: p.Label, Button: ButtonProps{Variant: Outline, Disabled: p.Disabled || p.ReadOnly, Invalid: p.Invalid, Size: p.Size}}, func() {
			label := p.Placeholder
			var option Option
			if !multiple && single.Valid {
				label = f.label(single.Value)
				for _, o := range p.Options {
					if !o.Separator && o.Value == single.Value {
						option = o
						break
					}
				}
			}
			if multiple && len(*selected) > 0 {
				labels := []string{}
				for _, key := range *selected {
					labels = append(labels, f.label(key))
				}
				label = strings.Join(labels, ", ")
			}
			if label == "" {
				label = "Select an option"
			}
			if p.TriggerContent != nil && !multiple && single.Valid {
				ui.Row(c).Grow(1).MinWidth(0).Children(func() { p.TriggerContent(c, option) })
			} else {
				ui.Text(c, label).Grow(1).SingleLine()
			}
		}).FillWidth()
	} else if multiple {
		root = f.chips(c, ComboboxChipsProps{Label: p.Label, Invalid: p.Invalid}, func() {
			seen := map[string]bool{}
			for _, key := range slices.Clone(*selected) {
				if seen[key] {
					continue
				}
				seen[key] = true
				f.chip(c, key, f.label(key), ComboboxChipProps{HideRemove: p.HideRemove}, nil)
			}
			f.editor(c, ComboboxInputProps{Label: p.Label, Placeholder: p.Placeholder, Size: p.Size, Invalid: p.Invalid}, false, true)
			if p.ShowClear && len(*selected) > 0 {
				f.clear(c, "Clear "+p.Label)
			} else if !p.HideTrigger {
				f.triggerIcon(c, p.Label)
			}
		})
	} else {
		root = f.input(c, ComboboxInputProps{Label: p.Label, Placeholder: p.Placeholder, Size: p.Size, Invalid: p.Invalid, HideTrigger: p.HideTrigger, ShowClear: p.ShowClear}).Root
	}
	content := ComboboxContentProps{MaxHeight: p.MaxHeight}
	var anchor ui.Handle
	if multiple && !p.PopupInput {
		root.Bind(&anchor)
		content.Anchor = &anchor
	}
	f.content(c, content, func() {
		if p.PopupInput {
			f.input(c, ComboboxInputProps{Label: p.Label, Placeholder: p.InputPlaceholder, Size: p.Size, Invalid: p.Invalid, HideTrigger: true, ShowClear: p.ShowClear})
		}
		empty := p.EmptyText
		if empty == "" {
			empty = "No items found."
		}
		f.empty(c, empty)
		f.list(c, ComboboxListProps{Label: p.Label + " options"}, func() {
			render := func(e comboEntry) {
				o := e.option
				f.item(c, o.Value, ComboboxItemProps{}, func() {
					if p.ItemContent != nil {
						p.ItemContent(c, o)
					} else {
						ui.Text(c, optionLabel(o)).SingleLine()
					}
				})
			}
			for i := 0; i < len(f.s.items); {
				if f.s.items[i].option.Separator {
					f.separator(c)
					i++
					continue
				}
				start, end, name := i, i+1, f.s.items[i].option.Group
				for end < len(f.s.items) && !f.s.items[end].option.Separator && f.s.items[end].option.Group == name {
					end++
				}
				if name != "" {
					ui.Column(c.Key(struct {
						Name  string
						Start int
					}{name, start})).FillWidth().Role(ui.RoleGroup).Label(name).Children(func() {
						f.groupTitle(c, name)
						for j := start; j < end; j++ {
							render(f.s.items[j])
						}
					})
				} else {
					for j := start; j < end; j++ {
						render(f.s.items[j])
					}
				}
				i = end
			}
		})
	})
	return root
}
