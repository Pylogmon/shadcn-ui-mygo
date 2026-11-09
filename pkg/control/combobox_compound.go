package control

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
)

// ComboboxSelection represents a selected value, including a valid zero value.
type ComboboxSelection[T any] struct {
	Value T
	Valid bool
}

// ComboboxState contains application values only. Keep its address stable while
// mounted. Changing these fields controls the next frame; native gestures update
// them. Internal editor, presence and navigation state belongs to the mount.
type ComboboxState[T any] struct {
	Value  ComboboxSelection[T]
	Values []T
	Query  string
	Open   bool
}

// ComboboxAdapter separates stable widget identity, display, submission and
// selection equality. Key is required. Equal defaults to comparing keys.
// Adapters are used during the current build and never retained by the widget.
type ComboboxAdapter[T any] struct {
	Key       func(T) string
	Label     func(T) string
	FormValue func(T) string
	Equal     func(T, T) bool
	Disabled  func(T) bool
}

func ComboboxStringAdapter() ComboboxAdapter[string] {
	return ComboboxAdapter[string]{Key: func(v string) string { return v }, Label: func(v string) string { return v }, FormValue: func(v string) string { return v }, Equal: func(a, b string) bool { return a == b }}
}
func ComboboxOptionAdapter() ComboboxAdapter[Option] {
	return ComboboxAdapter[Option]{Key: func(v Option) string { return v.Value }, Label: optionLabel, FormValue: func(v Option) string { return v.Value }, Equal: func(a, b Option) bool { return a.Value == b.Value }, Disabled: func(v Option) bool { return v.Disabled }}
}

type ComboboxGroupData[T any] struct {
	Key, Label string
	Items      []T
}
type ComboboxAutoHighlight uint8

const (
	ComboboxHighlightOff ComboboxAutoHighlight = iota
	ComboboxHighlightOnInput
	ComboboxHighlightAlways
)

// ComboboxChange is an event in the current build. Cancel rejects a proposed
// controlled change. It must not be retained as an application subscription.
type ComboboxChange struct {
	Reason   string
	canceled bool
}

func (e *ComboboxChange) Cancel()        { e.canceled = true }
func (e *ComboboxChange) Canceled() bool { return e.canceled }

type ComboboxRootProps[T any] struct {
	// Nil Items and Groups register manually rendered Item slots. An explicit
	// empty slice represents an empty data collection.
	Items                                []T
	Groups                               []ComboboxGroupData[T]
	Adapter                              ComboboxAdapter[T]
	Multiple, Disabled, ReadOnly, NoLoop bool
	// QueryControlled keeps an explicit empty query independent of the chosen
	// label; otherwise external inputs derive that label on mount/value changes.
	QueryControlled bool
	AutoHighlight   ComboboxAutoHighlight
	Locale          string
	Filter          func(T, string) bool
	// Grid enables spatial navigation and grid/gridcell accessibility. Layout
	// remains the caller's: Item geometry supplies its rows and columns.
	Grid bool
	// GridColumns defaults to two for uniform rows. Row slots support ragged rows.
	GridColumns int
	// Virtualized uses caller-rendered indexed items and Reveal to bring an
	// offscreen active index into the external native list's rendered range.
	Virtualized          bool
	Reveal               func(int)
	OnValueChange        func(ComboboxSelection[T], *ComboboxChange)
	OnValuesChange       func([]T, *ComboboxChange)
	OnQueryChange        func(string, *ComboboxChange)
	OnOpenChange         func(bool, *ComboboxChange)
	OnHighlighted        func(ComboboxSelection[T], *ComboboxChange)
	OnOpenChangeComplete func(bool)
}

type ComboboxSide string

const (
	ComboboxSideBottom      ComboboxSide = "bottom"
	ComboboxSideTop         ComboboxSide = "top"
	ComboboxSideLeft        ComboboxSide = "left"
	ComboboxSideRight       ComboboxSide = "right"
	ComboboxSideInlineStart ComboboxSide = "inline-start"
	ComboboxSideInlineEnd   ComboboxSide = "inline-end"
)

type ComboboxAlign string

const (
	ComboboxAlignStart  ComboboxAlign = "start"
	ComboboxAlignCenter ComboboxAlign = "center"
	ComboboxAlignEnd    ComboboxAlign = "end"
)

type ComboboxContentProps struct {
	Side  ComboboxSide
	Align ComboboxAlign
	// Nil uses the source's 6DIP default; a pointer to zero requests no gap.
	SideOffset  *float32
	AlignOffset float32
	// An explicit anchor follows the source data-chips sizing branch, whether
	// it refers to chips or to a caller-supplied native element.
	Anchor           *ui.Handle
	Width, MaxHeight float32
	Style            func(ui.Element)
}
type ComboboxInputProps struct {
	Label, Placeholder              string
	Size                            Size
	Invalid, HideTrigger, ShowClear bool
	Addons                          InputGroupProps
}
type ComboboxInputParts struct{ Root, Editor ui.Element }
type ComboboxTriggerProps struct {
	Label  string
	Button ButtonProps
	Render func(*ui.Context) ui.Element
}
type ComboboxListProps struct {
	Label  string
	Scroll *ui.ScrollState
}
type ComboboxItemProps struct {
	Disabled bool
}
type ComboboxChipsProps struct {
	Label   string
	Invalid bool
}
type ComboboxChipProps struct{ HideRemove bool }
type ComboboxValueProps struct{ Placeholder string }
type ComboboxValueView[T any] struct {
	Value    ComboboxSelection[T]
	Values   []T
	Multiple bool
}

// ComboboxScope exposes independent native slots. The scope, elements and
// callbacks are valid only during ComboboxRoot's build callback.
type ComboboxScope[T any] struct {
	f        *comboFrame
	state    *ComboboxState[T]
	p        ComboboxRootProps[T]
	byKey    map[string]T
	groups   map[string]ComboboxGroupData[T]
	group    string
	groupSet bool
	key      func(T) string
}

type comboMountKey struct{ identity any }
type comboRuntimeReference struct{ runtime *comboRuntime }

// ComboboxRoot is layout-transparent: a zero-size absolute native mount owns
// ui.Local state but does not wrap its slots or enter the accessibility tree.
// Leaving the view releases that mount and its state through public io.Closer
// semantics. State and every callback are supplied anew on each build.
func ComboboxRoot[T any](c *ui.Context, state *ComboboxState[T], p ComboboxRootProps[T], build func(*ComboboxScope[T])) {
	if state == nil || p.Adapter.Key == nil {
		panic("control.ComboboxRoot: state and Adapter.Key are required")
	}
	if len(p.Items) > 0 && len(p.Groups) > 0 {
		panic("control.ComboboxRoot: provide Items or Groups, not both")
	}
	if p.Adapter.Label == nil {
		p.Adapter.Label = func(v T) string { return p.Adapter.Key(v) }
	}
	if p.Adapter.FormValue == nil {
		p.Adapter.FormValue = p.Adapter.Key
	}
	if p.Adapter.Equal == nil {
		p.Adapter.Equal = func(a, b T) bool { return p.Adapter.Key(a) == p.Adapter.Key(b) }
	}
	b := &ComboboxScope[T]{state: state, p: p, byKey: map[string]T{}, groups: map[string]ComboboxGroupData[T]{}}
	options := []Option{}
	add := func(item T, group string) {
		key := p.Adapter.Key(item)
		if _, exists := b.byKey[key]; exists {
			panic("control.ComboboxRoot: duplicate item key " + key)
		}
		b.byKey[key] = item
		o := Option{Value: key, Label: p.Adapter.Label(item), Group: group}
		if p.Adapter.Disabled != nil {
			o.Disabled = p.Adapter.Disabled(item)
		}
		options = append(options, o)
	}
	for _, item := range p.Items {
		add(item, "")
	}
	for _, g := range p.Groups {
		if _, exists := b.groups[g.Key]; exists {
			panic("control.ComboboxRoot: duplicate group key " + g.Key)
		}
		b.groups[g.Key] = g
		for _, item := range g.Items {
			add(item, g.Key)
		}
	}
	canonical := func(v T) string {
		ordered := options
		if b.f != nil {
			ordered = b.f.p.Options
		}
		for _, option := range ordered {
			if item, exists := b.byKey[option.Value]; exists && p.Adapter.Equal(item, v) {
				return option.Value
			}
		}
		key := p.Adapter.Key(v)
		// Identity and selection equality are separate: a new item reusing a
		// key must not silently become selected when Equal explicitly rejects it.
		if _, collision := b.byKey[key]; collision {
			for {
				key = "\x00unmatched:" + key
				if _, exists := b.byKey[key]; !exists {
					break
				}
			}
		}
		return key
	}
	b.key = canonical
	single := ComboboxValue{}
	keys := []string{}
	if p.Multiple {
		for _, v := range state.Values {
			keys = append(keys, canonical(v))
		}
	} else if state.Value.Valid {
		single = ComboboxValue{canonical(state.Value.Value), true}
		keys = []string{single.Value}
	}
	props := ComboboxWithProps{Options: options, Disabled: p.Disabled, ReadOnly: p.ReadOnly, NoLoop: p.NoLoop, AutoHighlight: p.AutoHighlight != ComboboxHighlightOff, Locale: p.Locale}
	if p.Filter != nil {
		props.Filter = func(o Option, query string) bool { return p.Filter(b.byKey[o.Value], query) }
	}
	props.OnValueChange = func(next ComboboxValue) {
		v := ComboboxSelection[T]{}
		if next.Valid {
			item, ok := b.byKey[next.Value]
			if !ok {
				return
			}
			v = ComboboxSelection[T]{item, true}
		}
		event := &ComboboxChange{Reason: "value-change"}
		if p.OnValueChange != nil {
			p.OnValueChange(v, event)
		}
		if !event.canceled {
			state.Value = v
		} else {
			single = ComboboxValue{}
			keys = nil
			if state.Value.Valid {
				single = ComboboxValue{canonical(state.Value.Value), true}
				keys = []string{single.Value}
			}
		}
	}
	props.OnValuesChange = func(next []string) {
		values := make([]T, 0, len(next))
		for _, key := range next {
			if item, ok := b.byKey[key]; ok {
				values = append(values, item)
			} else {
				for _, existing := range state.Values {
					if canonical(existing) == key {
						values = append(values, existing)
						break
					}
				}
			}
		}
		event := &ComboboxChange{Reason: "value-change"}
		if p.OnValuesChange != nil {
			p.OnValuesChange(slices.Clone(values), event)
		}
		if !event.canceled {
			state.Values = values
		} else {
			keys = nil
			for _, v := range state.Values {
				keys = append(keys, canonical(v))
			}
		}
	}
	props.OnInputChange = func(next string) {
		event := &ComboboxChange{Reason: "input-change"}
		if p.OnQueryChange != nil {
			p.OnQueryChange(next, event)
		}
		if event.canceled && b.f != nil {
			state.Query = b.f.inputBefore
			b.f.queryCanceled = true
		}
	}
	props.OnOpenChange = func(next bool) {
		event := &ComboboxChange{Reason: "open-change"}
		if p.OnOpenChange != nil {
			p.OnOpenChange(next, event)
		}
		if !event.canceled {
			state.Open = next
		} else if b.f != nil {
			b.f.s.open = state.Open
			c.Invalidate()
		}
	}
	b.f = newComboFrame(c, state, &single, &keys, &state.Query, &state.Open, props, p.Multiple)
	b.f.resolveSelection = func() {
		keys = keys[:0]
		if p.Multiple {
			for _, value := range state.Values {
				keys = append(keys, canonical(value))
			}
		} else if state.Value.Valid {
			single = ComboboxValue{Value: canonical(state.Value.Value), Valid: true}
			keys = append(keys, single.Value)
		} else {
			single = ComboboxValue{}
		}
		b.f.s.selected = slices.Clone(keys)
	}
	b.f.manual = p.Items == nil && p.Groups == nil
	if b.f.manual {
		b.f.p.Options = slices.Clone(b.f.s.manualOptions)
	}
	b.f.always = p.AutoHighlight == ComboboxHighlightAlways
	b.f.queryControlled = p.QueryControlled
	b.f.grid, b.f.virtualized, b.f.reveal = p.Grid, p.Virtualized, p.Reveal
	b.f.s.grid, b.f.s.gridColumns = p.Grid, comboGridColumns(p.GridColumns)
	b.f.itemsSynced = false
	b.f.highlighted = func(key string, valid bool) {
		if p.OnHighlighted == nil {
			return
		}
		selection := ComboboxSelection[T]{}
		if value, ok := b.byKey[key]; valid && ok {
			selection = ComboboxSelection[T]{Value: value, Valid: true}
		}
		event := &ComboboxChange{Reason: "highlight-change"}
		p.OnHighlighted(selection, event)
		if event.canceled {
			b.f.s.active, b.f.s.hasActive = b.f.s.lastActive, b.f.s.lastHasActive
		}
	}
	b.f.completed = p.OnOpenChangeComplete
	b.f.groupLabel = func(key string) string {
		if g, ok := b.groups[key]; ok {
			return g.Label
		}
		return key
	}
	b.f.label = func(key string) string {
		if state.Value.Valid && canonical(state.Value.Value) == key {
			return p.Adapter.Label(state.Value.Value)
		}
		for _, v := range state.Values {
			if canonical(v) == key {
				return p.Adapter.Label(v)
			}
		}
		if v, ok := b.byKey[key]; ok {
			return p.Adapter.Label(v)
		}
		return key
	}
	defer b.f.finish()
	if build != nil {
		build(b)
	}
}

func (b *ComboboxScope[T]) Input(c *ui.Context, p ComboboxInputProps) ComboboxInputParts {
	b.f.check()
	return b.f.input(c, p)
}
func (b *ComboboxScope[T]) Trigger(c *ui.Context, p ComboboxTriggerProps, children func()) ui.Element {
	b.f.check()
	return b.f.trigger(c, p, children)
}
func (b *ComboboxScope[T]) Content(c *ui.Context, p ComboboxContentProps, children func()) ui.Element {
	b.f.check()
	return b.f.content(c, p, children)
}
func (b *ComboboxScope[T]) List(c *ui.Context, p ComboboxListProps, children func()) ui.Element {
	b.f.check()
	return b.f.list(c, p, children)
}
func (b *ComboboxScope[T]) Empty(c *ui.Context, text string) ui.Element {
	b.f.check()
	return b.f.empty(c, text)
}
func (b *ComboboxScope[T]) Separator(c *ui.Context) ui.Element { b.f.check(); return b.f.separator(c) }
func (b *ComboboxScope[T]) Clear(c *ui.Context, label string) ui.Element {
	b.f.check()
	return b.f.clear(c, label)
}
func (b *ComboboxScope[T]) Label(c *ui.Context, label string) ui.Element {
	b.f.check()
	return b.f.groupTitle(c, label)
}
func (b *ComboboxScope[T]) Item(c *ui.Context, value T, p ComboboxItemProps, children func()) ui.Element {
	b.f.check()
	key := b.p.Adapter.Key(value)
	b.byKey[key] = value
	option := Option{Value: key, Label: b.p.Adapter.Label(value), Disabled: p.Disabled, Group: b.group}
	if b.p.Adapter.Disabled != nil {
		option.Disabled = option.Disabled || b.p.Adapter.Disabled(value)
	}
	if b.f.manual {
		b.registerManual(option)
		updated := false
		for i := range b.f.p.Options {
			if b.f.p.Options[i].Value == key {
				b.f.p.Options[i] = option
				updated = true
				break
			}
		}
		if !updated {
			b.f.p.Options = append(b.f.p.Options, option)
		}
		b.f.itemsSynced = false
		if b.f.s.hasPending && b.f.s.pending == key {
			b.f.s.hasPending = false
			b.f.choose(key)
		}
	} else {
		found := false
		for _, item := range b.f.p.Options {
			if item.Value == key {
				found = true
				break
			}
		}
		if !found {
			b.f.p.Options = append(b.f.p.Options, option)
			b.f.itemsSynced = false
		}
	}
	return b.f.item(c, key, p, children)
}

// Collection iterates the current group's filtered children, or the root's
// filtered collection. It renders no wrapper of its own.
func (b *ComboboxScope[T]) Collection(c *ui.Context, render func(*ui.Context, T, int)) {
	b.f.check()
	b.f.syncItems()
	index := 0
	for _, entry := range b.f.s.items {
		if entry.option.Separator || b.groupSet && entry.option.Group != b.group {
			continue
		}
		if value, ok := b.byKey[entry.option.Value]; ok && render != nil {
			render(c, value, index)
			index++
		}
	}
}

// FilteredItems returns a current-frame copy for an external native scroll
// window or custom collection renderer. Item keeps full collection positions,
// even when only a window is rendered; Reveal receives that full filtered index.
func (b *ComboboxScope[T]) FilteredItems() []T {
	b.f.check()
	b.f.syncItems()
	out := []T{}
	for _, entry := range b.f.s.items {
		if !entry.option.Separator {
			if value, ok := b.byKey[entry.option.Value]; ok {
				out = append(out, value)
			}
		}
	}
	return out
}
func (b *ComboboxScope[T]) HighlightedIndex() int {
	b.f.check()
	b.f.syncItems()
	i := 0
	for _, entry := range b.f.s.items {
		if entry.option.Separator {
			continue
		}
		if b.f.s.hasActive && entry.option.Value == b.f.s.active {
			return i
		}
		i++
	}
	return -1
}

func (b *ComboboxScope[T]) registerManual(option Option) {
	for i := range b.f.manualEntries {
		if b.f.manualEntries[i].Value == option.Value {
			b.f.manualEntries[i] = option
			return
		}
	}
	b.f.manualEntries = append(b.f.manualEntries, option)
}
func (b *ComboboxScope[T]) Group(c *ui.Context, g ComboboxGroupData[T], children func()) ui.Element {
	b.f.check()
	if b.f.manual {
		for _, value := range g.Items {
			key := b.p.Adapter.Key(value)
			b.byKey[key] = value
			option := Option{Value: key, Label: b.p.Adapter.Label(value), Group: g.Key}
			if b.p.Adapter.Disabled != nil {
				option.Disabled = b.p.Adapter.Disabled(value)
			}
			b.registerManual(option)
			updated := false
			for i := range b.f.p.Options {
				if b.f.p.Options[i].Value == key {
					b.f.p.Options[i] = option
					updated = true
					break
				}
			}
			if !updated {
				b.f.p.Options = append(b.f.p.Options, option)
			}
		}
		b.f.itemsSynced = false
	}
	previous, previousSet := b.group, b.groupSet
	b.group, b.groupSet = g.Key, true
	defer func() { b.group, b.groupSet = previous, previousSet }()
	return ui.Column(c.Key(struct{ Group string }{g.Key})).FillWidth().Role(ui.RoleGroup).Label(g.Label).Children(children)
}
func (b *ComboboxScope[T]) Groups(c *ui.Context, render func(*ui.Context, ComboboxGroupData[T], int)) {
	b.f.check()
	b.f.syncItems()
	i := 0
	for _, g := range b.p.Groups {
		visible := false
		for _, e := range b.f.s.items {
			if !e.option.Separator && e.option.Group == g.Key {
				visible = true
				break
			}
		}
		if visible && render != nil {
			render(c, g, i)
			i++
		}
	}
}
func (b *ComboboxScope[T]) Value(c *ui.Context, p ComboboxValueProps, render ...func(ComboboxValueView[T])) {
	b.f.check()
	view := ComboboxValueView[T]{Value: b.state.Value, Values: slices.Clone(b.state.Values), Multiple: b.p.Multiple}
	if len(render) > 0 && render[0] != nil {
		render[0](view)
		return
	}
	label := p.Placeholder
	if b.p.Multiple {
		labels := []string{}
		for _, v := range b.state.Values {
			labels = append(labels, b.p.Adapter.Label(v))
		}
		if len(labels) > 0 {
			label = strings.Join(labels, ", ")
		}
	} else if b.state.Value.Valid {
		label = b.p.Adapter.Label(b.state.Value.Value)
	}
	ui.Text(c, label).SingleLine()
}
func (b *ComboboxScope[T]) Chips(c *ui.Context, p ComboboxChipsProps, children func()) ui.Element {
	b.f.check()
	return b.f.chips(c, p, children)
}
func (b *ComboboxScope[T]) Chip(c *ui.Context, value T, p ComboboxChipProps, children func()) ui.Element {
	b.f.check()
	return b.f.chip(c, b.key(value), b.p.Adapter.Label(value), p, children)
}
func (b *ComboboxScope[T]) ChipsInput(c *ui.Context, p ComboboxInputProps) ui.Element {
	b.f.check()
	return b.f.editor(c, p, false, true)
}
func (b *ComboboxScope[T]) FormValues() []string {
	b.f.check()
	if b.p.Multiple {
		out := make([]string, 0, len(b.state.Values))
		for _, v := range b.state.Values {
			out = append(out, b.p.Adapter.FormValue(v))
		}
		return out
	}
	if b.state.Value.Valid {
		return []string{b.p.Adapter.FormValue(b.state.Value.Value)}
	}
	return nil
}

func comboRequireScope(alive bool) {
	if !alive {
		panic("control.ComboboxScope: slots must be built during ComboboxRoot")
	}
}
