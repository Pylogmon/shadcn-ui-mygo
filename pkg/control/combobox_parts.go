package control

import (
	"math"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// comboFrame is deliberately frame-owned. It is the only place that holds UI
// objects, current bindings, render functions and application callbacks.
type comboEmptyFrame struct {
	c       *ui.Context
	element ui.Element
	text    string
}
type comboItemFrame struct {
	c                  *ui.Context
	element, indicator ui.Element
	key                string
	active             bool
}

type comboFrame struct {
	hosted                                                                               comboHostedFrame
	c                                                                                    *ui.Context
	s                                                                                    *comboRuntime
	p                                                                                    ComboboxWithProps
	single                                                                               *ComboboxValue
	values                                                                               *[]string
	query                                                                                *string
	boundOpen                                                                            *bool
	multiple, alive, justMounted, externalOpen, changed                                  bool
	always, grid, virtualized                                                            bool
	slotDisabled                                                                         map[string]bool
	gridCells                                                                            []comboGridCell
	queryControlled                                                                      bool
	inputBefore                                                                          string
	manual                                                                               bool
	manualEmpty                                                                          []comboEmptyFrame
	manualItems                                                                          []comboItemFrame
	manualListEmpty                                                                      bool
	resolveSelection                                                                     func()
	manualEntries                                                                        []Option
	prepared                                                                             bool
	queryCanceled                                                                        bool
	contentHeight, contentGap                                                            float32
	reveal                                                                               func(int)
	highlighted                                                                          func(string, bool)
	completed                                                                            func(bool)
	label, groupLabel                                                                    func(string) string
	mount, primary, inputRoot, triggerRoot, chipsRoot, editorElement, panel, listElement ui.Element
	inContent                                                                            bool
	itemsSynced                                                                          bool
	labelText                                                                            string
	itemIndex                                                                            int
	row                                                                                  int
	inRow                                                                                bool
	parentGroup                                                                          *groupBuild
	beforeGroupItems                                                                     int
}

func (s *comboRuntime) Close() error { *s = comboRuntime{closed: true}; return nil }

func newComboFrame(c *ui.Context, identity any, single *ComboboxValue, values *[]string, query *string, open *bool, p ComboboxWithProps, multiple bool) *comboFrame {
	mount := ui.Box(c).Key(comboMountKey{identity}).Absolute().Size(0, 0).IgnoreInput(true).AccessHidden(true)
	s := ui.Local(mount, "shadcn.control.selection-combobox", func() comboRuntime { return comboRuntime{chips: map[string]*ui.Handle{}} })
	f := &comboFrame{c: c, s: s, p: p, single: single, values: values, query: query, boundOpen: open, multiple: multiple, alive: true, justMounted: !s.initialized, mount: mount}
	if open != nil && !f.justMounted {
		p.PopupInput = s.popupInput
		f.p.PopupInput = p.PopupInput
	}
	f.label = func(key string) string { return comboLabel(f.p.Options, key) }
	f.groupLabel = func(key string) string { return key }
	if scope := groupScope(c); len(scope.active) > 0 {
		f.parentGroup = scope.active[len(scope.active)-1]
		f.beforeGroupItems = len(f.parentGroup.items)
	}
	s.enabled, s.multiple, s.autoHighlight, s.noLoop, s.direction = !mount.IsDisabled() && !p.Disabled && !p.ReadOnly, multiple, p.AutoHighlight, p.NoLoop, c.Direction()
	if open != nil {
		if !s.boundOpenInitialized {
			s.open = *open
			s.wasOpen = *open
			s.lastBoundOpen = *open
			s.boundOpenInitialized = true
		} else if *open != s.lastBoundOpen {
			s.open = *open
			f.externalOpen = true
		}
	}
	if !s.initialized {
		if open == nil && !multiple && single.Valid && *query == "" && !p.PopupInput {
			*query = f.label(single.Value)
		}
		s.committedQuery, s.previousFilterQuery = *query, *query
		s.initialized = true
		s.completeInitialized, s.completeOpen = !s.open, false
	}
	if open == nil && !multiple && !s.open && !p.PopupInput && (*single != s.previousValue || single.Valid && f.label(single.Value) != s.previousLabel) {
		if single.Valid {
			*query = f.label(single.Value)
		} else {
			*query = ""
		}
	}
	if !s.composing && !s.compositionCommit {
		s.committedQuery = *query
		if open != nil && s.previousQuery != *query {
			s.filtering = true
		}
	}
	if !s.enabled {
		s.open, s.hasActive, s.composing = false, false, false
	}
	return f
}
func (f *comboFrame) prepare() {
	if f.prepared {
		return
	}
	f.prepared = true
	s, p, multiple := f.s, f.p, f.multiple
	if s.hasTypeaheadValue {
		s.hasTypeaheadValue = false
		if s.enabled && p.PopupInput && !multiple && !s.open {
			f.change([]string{s.typeaheadValue})
		}
	}
	if s.clear {
		s.clear = false
		f.change(nil)
		f.setQuery("")
		s.hasActive, s.filtering = false, false
		s.input.Focus()
	}
	if s.hasRemove {
		s.hasRemove = false
		f.remove(s.remove)
	}
	s.selected = slices.Clone(*f.values)
	f.syncItems()
	if s.hasPending && !f.manual {
		key := s.pending
		s.hasPending = false
		f.choose(key)
	}
}

func (f *comboFrame) check() { comboRequireScope(f.alive) }
func (f *comboFrame) setQuery(q string) {
	if *f.query != q {
		f.inputBefore = *f.query
		*f.query = q
		if !f.s.composing {
			f.s.committedQuery = q
		}
		if f.p.OnInputChange != nil {
			f.p.OnInputChange(q)
		}
		f.s.committedQuery = *f.query
		f.queryCanceled = false
	}
	f.itemsSynced = false
}
func (f *comboFrame) change(values []string) {
	if f.multiple {
		if !slices.Equal(*f.values, values) {
			*f.values = slices.Clone(values)
			if f.p.OnValuesChange != nil {
				f.p.OnValuesChange(slices.Clone(values))
			}
			f.changed = true
			if f.primary.Valid() {
				f.primary.SetChanged()
			}
		}
	} else {
		v := ComboboxValue{}
		if len(values) > 0 {
			v = ComboboxValue{values[0], true}
		}
		if *f.single != v {
			*f.single = v
			*f.values = slices.Clone(values)
			if f.p.OnValueChange != nil {
				f.p.OnValueChange(v)
			}
			f.changed = true
			if f.primary.Valid() {
				f.primary.SetChanged()
			}
		}
	}
	f.s.selected = slices.Clone(*f.values)
	f.itemsSynced = false
}
func (f *comboFrame) remove(key string) {
	f.change(slices.DeleteFunc(slices.Clone(*f.values), func(v string) bool { return v == key }))
}
func (f *comboFrame) choose(key string) {
	if !f.s.enabled {
		return
	}
	f.syncItems()
	var o Option
	found := false
	for _, e := range f.s.items {
		if e.option.Value == key && !e.option.Separator && !e.option.Disabled {
			o = e.option
			found = true
			break
		}
	}
	if !found {
		return
	}
	filtered := f.s.filtering && comboTrimQuery(*f.query) != ""
	if !f.multiple || filtered && !f.p.PopupInput {
		f.s.captureClose(f.s.filterQuery(*f.query))
	}
	if f.multiple {
		values := slices.Clone(*f.values)
		if slices.Contains(values, key) {
			values = slices.DeleteFunc(values, func(v string) bool { return v == key })
		} else {
			values = append(values, key)
		}
		f.change(values)
		if filtered && !f.p.PopupInput {
			f.s.open = false
		}
		if filtered {
			f.setQuery("")
		}
	} else {
		f.change([]string{key})
		if !f.p.PopupInput {
			f.setQuery(optionLabel(o))
		}
		f.s.open = false
	}
	f.s.filtering, f.s.hasActive, f.s.pointerHighlight = false, false, false
	if f.p.PopupInput {
		f.s.trigger.Focus()
	} else {
		f.s.input.Focus()
	}
}
func (f *comboFrame) syncItems() {
	if f.itemsSynced {
		return
	}
	f.itemsSynced = true
	s := f.s
	query := s.filterQuery(*f.query)
	filtering := s.filtering
	if s.closing {
		query, filtering = s.closeQuery, s.closeFiltering
	}
	s.items = comboEntries(f.p, query, filtering, *f.values)
	for i := range s.items {
		s.items[i].option.Disabled = s.items[i].option.Disabled || s.slotDisabled[s.items[i].option.Value]
	}
	for i := range s.items {
		if s.items[i].option.Separator {
			s.items[i].option.Disabled = true
		}
	}
	if s.previousFilterQuery != query {
		s.hasActive = false
		if s.filtering && f.p.AutoHighlight && comboTrimQuery(query) != "" {
			s.edge(false)
		}
	}
	if s.hasActive {
		valid := false
		for _, e := range s.items {
			if !e.option.Separator && e.option.Value == s.active && (!e.option.Disabled || e.selected) {
				valid = true
				break
			}
		}
		if !valid {
			s.hasActive = false
		}
	}
	if s.open && !s.wasOpen && !s.filtering && !s.hasActive {
		s.highlightSelected()
	}
	if (f.always && s.open || s.filtering && f.p.AutoHighlight && comboTrimQuery(query) != "") && !s.hasActive {
		s.edge(false)
	}
}
func (f *comboFrame) presence() (comboPose, bool) {
	s := f.s
	if s.wasOpen && !s.open {
		s.captureClose(s.filterQuery(*f.query))
	}
	if s.open {
		s.closing = false
	}
	pose, present := s.motion.update(f.c, s.open)
	if s.closing {
		if f.multiple && !f.p.PopupInput {
			f.restoreQuery("")
		}
		if !present {
			s.closing, s.filtering, s.hasActive = false, false, false
			if f.multiple || f.p.PopupInput || !f.single.Valid {
				f.restoreQuery("")
			} else {
				f.restoreQuery(f.label(f.single.Value))
			}
			s.committedQuery = *f.query
		}
	}
	return pose, present
}
func (f *comboFrame) restoreQuery(query string) {
	if f.boundOpen != nil {
		f.setQuery(query)
	} else {
		*f.query = query
	}
}
func (f *comboFrame) bindRoot(root ui.Element) {
	if !f.primary.Valid() {
		f.primary = root
		if f.changed {
			root.SetChanged()
		}
	}
	ref := ui.Local(root, "shadcn.control.combobox-controller", func() comboRuntimeReference { return comboRuntimeReference{f.s} })
	ref.runtime = f.s
	if f.p.Disabled {
		root.Disabled(true)
	}
	root.DimDisabled(false)
}
func (f *comboFrame) input(c *ui.Context, p ComboboxInputProps) ComboboxInputParts {
	f.check()
	f.prepare()
	f.p.Size = p.Size
	if f.hosted.native {
		return f.nativeHostedInput(c, p)
	}
	t := theme.TokensOf(c)
	f.labelText = p.Label
	if f.labelText == "" {
		f.labelText = f.p.Label
	}
	root := ui.Row(c.Key("combobox-input")).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(0).Radius(t.RadiusMD()).TextColor(t.Foreground)
	f.bindRoot(root)
	f.inputRoot = root
	invalid := fieldInputInvalid(c, p.Invalid || p.Addons.Invalid)
	root.Disabled(f.p.Disabled || p.Addons.Disabled)
	textInputAppearance(c, root, invalid)
	if t.Dark {
		root.Background(t.Input.Alpha(.3))
	}
	height, _ := metrics(p.Size)
	scale := comboTextScale(c)
	height = max(height, inputFont(c, p.Size)*inputLineHeight(p.Size)+2)
	root.MinHeight(height).Padding(0)
	if f.inContent {
		root.Margin(4, 4, 0, 4).MinHeight(max(float32(32), 20*scale+12)).Height(max(float32(32), 20*scale+12)).Padding(0).Background(t.Input.Alpha(.3)).Border(1, t.Input.Alpha(.3)).NoShadow()
	}
	var editor ui.Element
	renderInline := func() {
		if p.Addons.InlineStart != nil {
			ui.Row(c.Key("combobox-addon-start")).Shrink(0).Padding(6, 0).PaddingStart(8).Gap(8).AlignItems(ui.Center).FontSize(14).TextColor(t.MutedForeground).Children(p.Addons.InlineStart)
		}
		if p.Addons.Prefix != "" {
			ui.Text(c, p.Addons.Prefix).TextColor(t.MutedForeground)
		}
		editor = f.editor(c, p, f.inContent, false)
		if p.Addons.InlineStart != nil || p.Addons.Prefix != "" {
			editor.PaddingStart(6)
		}
		if p.ShowClear && len(*f.values) > 0 {
			f.clear(c, "Clear "+f.labelText)
		} else if !f.inContent {
			if !p.HideTrigger {
				f.triggerIcon(c, f.labelText)
			}
		}
		if p.Addons.Suffix != "" {
			ui.Text(c, p.Addons.Suffix).TextColor(t.MutedForeground)
		}
		if p.Addons.InlineEnd != nil {
			ui.Row(c.Key("combobox-addon-end")).Shrink(0).Padding(6, 0).PaddingEnd(8).Gap(8).AlignItems(ui.Center).FontSize(14).TextColor(t.MutedForeground).Children(p.Addons.InlineEnd)
		}
	}
	if p.Addons.BlockStart != nil || p.Addons.BlockEnd != nil {
		root.Column()
		root.Children(func() {
			if p.Addons.BlockStart != nil {
				ui.Row(c).FillWidth().Padding(8, 10, 2, 10).Children(p.Addons.BlockStart)
			}
			ui.Row(c).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(renderInline)
			if p.Addons.BlockEnd != nil {
				ui.Row(c).FillWidth().Padding(2, 10, 8, 10).Children(p.Addons.BlockEnd)
			}
		})
	} else {
		root.Children(renderInline)
	}

	handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	editor.Bind(handle)
	if editor.FocusVisible() && !invalid {
		root.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
	}
	if root.IsDisabled() {
		root.Opacity(.5)
	}
	if appearance := ui.Local(c.Root(), inputGroupAppearanceKey{}, func() inputGroupAppearance { return inputGroupAppearance{} }); len(appearance.active) > 0 {
		root.Border(0, ui.Transparent).Background(ui.Transparent).NoShadow()
	}
	root.HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind == ui.InputPointerDown && ev.Button == 0 && f.s.enabled {
			f.s.input.Focus()
		}
		return false
	})
	ButtonGroupItem(c, root)
	return ComboboxInputParts{root, editor}
}
func comboTextScale(c *ui.Context) float32 {
	scale := c.Preferences().TextScale
	if scale <= 0 || math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) {
		return 1
	}
	return scale
}
func (f *comboFrame) editor(c *ui.Context, p ComboboxInputProps, inPopup, chips bool) ui.Element {
	f.check()
	f.prepare()
	if !inPopup && !f.multiple && !f.queryControlled && f.single.Valid && ((f.justMounted && *f.query == "") || !f.s.open && (*f.single != f.s.previousValue || f.label(f.single.Value) != f.s.previousLabel)) {
		*f.query = f.label(f.single.Value)
		f.s.committedQuery = *f.query
	}
	f.p.PopupInput = inPopup
	f.s.popupInput = inPopup
	label := p.Label
	if label == "" {
		label = f.labelText
	}
	placeholder := p.Placeholder
	if placeholder == "" && inPopup {
		placeholder = "Search…"
	}
	var input ui.Element
	if f.hosted.native {
		input = ui.TextInput(c.Key("combobox-editor"), f.query).Label(label).Placeholder(placeholder).Disabled(f.p.Disabled).ReadOnly(f.p.ReadOnly).Invalid(p.Invalid)
	} else {
		input = styleInput(c, ui.TextInputBase(c.Key("combobox-editor"), f.query), InputProps{Label: label, Placeholder: placeholder, Disabled: f.p.Disabled, ReadOnly: f.p.ReadOnly, Invalid: p.Invalid, Size: p.Size}).Border(0, ui.Transparent).Radius(0).Background(ui.Transparent).FocusRing(false).NoShadow().Padding(0)
	}
	input.Bind(&f.s.input).Role(ui.RoleComboBox).HasPopup(ui.PopupListBox).AutoComplete(ui.AutoCompleteList).Expanded(f.s.open).Grow(1).MinWidth(0).HandleInput(f.s.inputEvent).ObserveInput(f.s.observe)
	if !f.hosted.native {
		height, _ := metrics(p.Size)
		height = max(height, inputFont(c, p.Size)*inputLineHeight(p.Size)+2)
		input.Height(height - 2)
		if !chips && !inPopup {
			input.Padding(4, 6, 4, 10)
		}
		if inPopup {
			input.Height(max(float32(30), 20*comboTextScale(c)+10)).Padding(4, 8).FontSize(14).LineHeight(20.0 / 14)
		}
		if chips {
			input.MinWidth(64).Height(max(float32(20), 20*comboTextScale(c))).FontSize(14).LineHeight(20.0 / 14)
		}
	}
	f.editorElement = input
	f.s.enabled = f.s.enabled && !input.IsDisabled() && !f.p.ReadOnly
	if !f.s.enabled {
		f.s.open, f.s.hasActive, f.s.composing = false, false, false
	}
	wasComposing := f.s.composing
	f.s.composing = input.Composing()
	notified := false
	commit := func() {
		if notified {
			return
		}
		if input.Composing() {
			return
		}
		notified = true
		f.s.compositionCommit = false
		f.inputBefore = f.s.committedQuery
		if f.p.OnInputChange != nil {
			f.p.OnInputChange(*f.query)
		}
		f.s.committedQuery = *f.query
		if f.queryCanceled {
			f.queryCanceled = false
			return
		}
		f.s.open, f.s.filtering, f.s.pointerHighlight = true, true, false
		f.s.hasActive = false
		if !f.multiple && !inPopup && *f.query == "" {
			f.change(nil)
		}
		f.itemsSynced = false
	}
	input.OnChange(commit)
	if f.s.compositionCommit || wasComposing && !f.s.composing {
		commit()
	}
	if input.Clicked() && f.s.enabled {
		f.s.open = true
		f.itemsSynced = false
	}
	f.s.queryEmpty = *f.query == ""
	f.s.selectionStart, f.s.selectionEnd = input.TextSelection()
	if f.listElement.Valid() {
		input.Controls(f.listElement)
	}
	return input
}
func (f *comboFrame) triggerIcon(c *ui.Context, label string) ui.Element {
	t := theme.TokensOf(c)
	height, _ := metrics(f.p.Size)
	size := min(float32(24), height-2)
	button := ui.ButtonBase(c.Key("combobox-trigger-icon")).Size(size, size).MarginEnd(4).KeepFocus().TabStop(false).Label("Show " + label + " options").Disabled(!f.s.enabled).FocusRing(false).Children(func() { indicator.ChevronDown(c).TextColor(t.MutedForeground) })
	if button.Clicked() {
		f.s.toggle(*f.query)
		f.itemsSynced = false
		f.s.input.Focus()
	}
	return button
}
func (f *comboFrame) trigger(c *ui.Context, p ComboboxTriggerProps, children func()) ui.Element {
	f.check()
	f.prepare()
	root := ui.Element{}
	if p.Render != nil {
		root = p.Render(c)
	} else {
		root = ui.ButtonBase(c)
		button := p.Button
		if button.Variant == Default {
			button.Variant = Outline
		}
		button.Label = p.Label
		button.Disabled = button.Disabled || f.p.Disabled || f.p.ReadOnly
		StyleButton(c, root, "", button)
	}
	f.bindRoot(root)
	f.triggerRoot = root
	f.labelText = p.Label
	inside := !f.editorElement.Valid() || f.inContent
	f.p.PopupInput = inside
	f.s.popupInput = inside
	f.s.enabled = f.s.enabled && !root.IsDisabled()
	if !f.s.enabled {
		f.s.open, f.s.hasActive = false, false
	}
	root.ReadOnly(f.p.ReadOnly).Bind(&f.s.trigger).Role(ui.RoleComboBox).HasPopup(ui.PopupDialog).Expanded(f.s.open).Gap(8).Focusable().FocusRing(false).HandleInput(f.s.triggerEvent).Label(p.Label)
	if !inside {
		root.HasPopup(ui.PopupListBox).TabStop(false).KeepFocus()
	}
	handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	root.Bind(handle)
	root.Children(func() {
		if children != nil {
			children()
		} else {
			label := f.p.Placeholder
			if !f.multiple && f.single.Valid {
				label = f.label(f.single.Value)
			}
			if f.multiple && len(*f.values) > 0 {
				labels := []string{}
				for _, key := range *f.values {
					labels = append(labels, f.label(key))
				}
				label = strings.Join(labels, ", ")
			}
			if label == "" {
				label = "Select an option"
			}
			ui.Text(c, label).Grow(1).SingleLine()
		}
		indicator.ChevronDown(c).TextColor(theme.TokensOf(c).MutedForeground)
	})
	if root.Clicked() && f.s.enabled {
		f.s.toggle(*f.query)
		f.itemsSynced = false
	}
	ButtonGroupItem(c, root)
	return root
}
func (f *comboFrame) clear(c *ui.Context, label string) ui.Element {
	f.check()
	if len(*f.values) == 0 {
		return ui.Element{}
	}
	height, _ := metrics(f.p.Size)
	size := min(float32(24), height-2)
	button := ui.ButtonBase(c.Key("combobox-clear")).Size(size, size).MarginEnd(4).KeepFocus().TabStop(false).Label(label).Disabled(!f.s.enabled).FocusRing(false).Children(func() { indicator.Close(c) })
	if button.Clicked() {
		f.change(nil)
		f.setQuery("")
		f.s.hasActive, f.s.filtering = false, false
		f.s.input.Focus()
	}
	return button
}
func (f *comboFrame) chips(c *ui.Context, p ComboboxChipsProps, children func()) ui.Element {
	f.check()
	f.prepare()
	t := theme.TokensOf(c)
	root := ui.Row(c.Key("combobox-chips")).FillWidth().MinWidth(0).Wrap().MinHeight(36).Padding(6, 10).Gap(6).AlignItems(ui.Center).Radius(t.RadiusMD()).TextColor(t.Foreground).Role(ui.RoleToolbar)
	if len(*f.values) > 0 {
		root.Padding(6, 6)
	}
	textInputAppearance(c, root, fieldInputInvalid(c, p.Invalid))
	if t.Dark {
		root.Background(t.Input.Alpha(.3))
	}
	f.bindRoot(root)
	f.chipsRoot = root
	f.inputRoot = root
	f.labelText = p.Label
	root.Children(children)
	if f.editorElement.Valid() {
		handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
		f.editorElement.Bind(handle)
	}
	if root.FocusWithin() && !p.Invalid {
		root.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
	}
	root.HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind == ui.InputPointerDown && ev.Button == 0 && f.s.enabled {
			f.s.input.Focus()
		}
		return false
	})
	ButtonGroupItem(c, root)
	return root
}
func (f *comboFrame) chip(c *ui.Context, key, label string, p ComboboxChipProps, children func()) ui.Element {
	f.check()
	t := theme.TokensOf(c)
	if f.s.chips[key] == nil {
		f.s.chips[key] = &ui.Handle{}
	}
	chip := ui.Row(c.Key(choiceOptionKey{value: key})).Height(22).Padding(2, 6).Gap(4).AlignItems(ui.Center).Radius(t.RadiusSM()).Background(t.Muted).TextColor(t.Foreground).FontSize(12).FontWeight(500).Focusable().TabStop(false).Label(label).Bind(f.s.chips[key]).HandleInput(func(ev ui.InputEvent) bool { return f.s.chipEvent(key, ev) })
	if !p.HideRemove {
		chip.PaddingEnd(0)
	}
	chip.Children(func() {
		if children != nil {
			children()
		} else {
			ui.Text(c, label).SingleLine()
		}
		if !p.HideRemove {
			remove := ui.ButtonBase(c).Size(24, 22).MarginStart(-4).KeepFocus().TabStop(false).Label("Remove " + label).Disabled(!f.s.enabled).FocusRing(false).Opacity(.5).Children(func() { indicator.Close(c).Size(12, 12) })
			if remove.Hovered() {
				remove.Opacity(1)
			}
			if remove.Clicked() {
				f.remove(key)
				f.s.input.Focus()
			}
		}
	})
	return chip
}
func (f *comboFrame) empty(c *ui.Context, text string) ui.Element {
	f.check()
	f.syncItems()
	n := 0
	for _, entry := range f.s.items {
		if !entry.option.Separator {
			n++
		}
	}
	status := ui.Row(c.Key("combobox-empty")).FillWidth().Role(ui.RoleStatus).Justify(ui.Center).TextAlign(ui.Center).FontSize(14).TextColor(theme.TokensOf(c).MutedForeground)
	if f.hosted.native {
		status.FontSize(c.Theme().FontSize).TextColor(c.Theme().TextMuted)
	}
	if f.manual {
		// Manual items register later in this build. Populate this owned status
		// once from the final collection, before native layout/AX are submitted.
		f.manualEmpty = append(f.manualEmpty, comboEmptyFrame{c: c, element: status, text: text})
		return status
	}
	if n == 0 {
		status.Padding(8, 0).Children(func() { ui.Text(c, text) })
	} else {
		status.Height(0).IgnoreInput(true).AccessHidden(true)
	}
	return status
}
func (f *comboFrame) groupTitle(c *ui.Context, label string) ui.Element {
	if f.hosted.native {
		return ui.Text(c, label).Padding(c.Theme().Space(1.5), c.Theme().Space(2.5)).FontSize(c.Theme().FontSize).TextColor(c.Theme().TextMuted)
	}
	return ui.Text(c, label).Padding(6, 8).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
}
func (f *comboFrame) separator(c *ui.Context) ui.Element {
	return ui.Box(c).Height(1).FillWidth().Margin(4, -4).Background(theme.MenuOf(c).Separator()).Role(ui.RoleSeparator).Orientation(ui.Horizontal)
}
func (f *comboFrame) list(c *ui.Context, p ComboboxListProps, children func()) ui.Element {
	f.check()
	f.syncItems()
	label := p.Label
	if label == "" {
		label = f.labelText + " options"
	}
	list := ui.Scroll(c.Key("combobox-list")).TabStop(false).FillWidth().Padding(4).Scrollbars(false).Role(ui.RoleListBox).Multiselectable(f.multiple).Label(label).HandleInput(f.s.pointerEvent)
	if f.grid {
		list.Role(ui.RoleGrid).ColumnCount(f.s.gridColumns)
	}
	n := 0
	for _, e := range f.s.items {
		if !e.option.Separator {
			n++
		}
	}
	if n == 0 {
		list.Padding(0)
	}
	f.manualListEmpty = n == 0
	maxHeight := float32(252)
	if f.panel.Valid() {
		available := f.contentAvailable(f.panel)
		maxHeight = max(float32(0), min(maxHeight, available.H-36))
		if f.contentHeight > 0 {
			maxHeight = max(float32(1), min(maxHeight, f.contentHeight-36))
		}
	}
	if f.hosted.native {
		list.Scrollbars(true).Padding(c.Theme().Space(1))
		// Native text metrics can exceed the source's 36DIP input allowance.
		// Keep the scroll viewport inside the independent window at large fonts.
		inputHeight := f.inputRoot.LayoutBounds().H
		if inputHeight <= 0 {
			inputHeight = c.Theme().FontSize*1.4 + c.Theme().Space(3) + 2
		}
		if f.contentHeight > 0 {
			maxHeight = min(maxHeight, max(float32(1), f.contentHeight-inputHeight-2*c.Theme().Space(1)))
		}
	}
	list.MaxHeight(maxHeight)
	if p.Scroll != nil {
		list.TrackScroll(p.Scroll)
	}
	f.listElement = list
	f.itemIndex = 0
	if f.editorElement.Valid() {
		f.editorElement.Controls(list).Expanded(f.s.open)
	}
	_, _, over := list.PointerPosition()
	if f.s.pointerOver && !over && f.s.pointerHighlight {
		f.s.hasActive, f.s.pointerHighlight = false, false
	}
	f.s.pointerOver = over
	list.Children(children)
	f.s.pointerMoved = false
	return list
}
func (f *comboFrame) item(c *ui.Context, key string, p ComboboxItemProps, children func()) ui.Element {
	f.check()
	f.syncItems()
	entry := comboEntry{}
	found := false
	index := -1
	n := 0
	for _, e := range f.s.items {
		if e.option.Separator {
			continue
		}
		if e.option.Value == key {
			entry, found, index = e, true, n
		}
		n++
	}
	if !found {
		return ui.Element{}
	}
	if f.slotDisabled == nil {
		f.slotDisabled = map[string]bool{}
	}
	f.slotDisabled[key] = p.Disabled
	for i := range f.s.items {
		if f.s.items[i].option.Value == key {
			baseDisabled := false
			for _, option := range f.p.Options {
				if option.Value == key {
					baseDisabled = option.Disabled
					break
				}
			}
			f.s.items[i].option.Disabled = baseDisabled || p.Disabled
			entry.option.Disabled = baseDisabled || p.Disabled
		}
	}
	t := theme.TokensOf(c)
	option := ui.ButtonBase(c.Key(choiceOptionKey{value: key})).Role(ui.RoleOption).KeepFocus().TabStop(false).Disabled(entry.option.Disabled || !f.s.enabled).DimDisabled(false).Label(optionLabel(entry.option)).PositionInSet(index+1, n).AccessSelected(entry.selected).FillWidth().MinWidth(0).MinHeight(32).Padding(6, 8).PaddingEnd(32).Gap(8).AlignItems(ui.Center).Radius(t.RadiusSM()).FontSize(14).FocusRing(false).HandleInput(f.s.pointerEvent)
	if f.grid && f.inRow {
		option.Role(ui.RoleCell).RowIndex(f.row + 1).ColumnIndex(f.itemIndex + 1)
	}
	if f.grid {
		row, col := index/comboGridColumns(f.s.gridColumns), index%comboGridColumns(f.s.gridColumns)
		if f.inRow {
			row, col = f.row, f.itemIndex
		}
		f.gridCells = append(f.gridCells, comboGridCell{key: key, row: row, column: col})
	}
	f.itemIndex++
	active := f.s.hasActive && f.s.active == key
	pointerMoved := f.s.pointerMoved && !f.s.pointerTouch
	if active && f.s.pointerHighlight && pointerMoved && !option.Hovered() {
		f.s.hasActive, f.s.pointerHighlight = false, false
		active = false
	}
	if pointerMoved && option.Hovered() && !entry.option.Disabled && f.s.enabled {
		f.s.active, f.s.hasActive, f.s.pointerHighlight = key, true, true
		active = true
	}
	if active {
		background, foreground := theme.MenuOf(c).Highlight(false)
		option.Background(background).TextColor(foreground)
		if f.editorElement.Valid() {
			f.editorElement.ActiveDescendant(option)
		}
		if f.s.reveal {
			option.ScrollIntoView()
			f.s.reveal = false
		}
	}
	if f.hosted.native {
		native := c.Theme()
		option.Padding(native.Space(1.5), native.Space(2.5)).PaddingEnd(native.Space(7)).MinHeight(native.FontSize + native.Space(3)).Radius(native.Radius).FontSize(native.FontSize).DimDisabled(true)
		if active {
			option.Background(native.Accent).TextColor(native.AccentText)
		} else if option.Hovered() {
			option.Background(native.SurfaceHover).TextColor(native.Text)
		}
	}
	if entry.option.Disabled && !f.hosted.native {
		option.Opacity(.5)
	}
	option.Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).Children(func() {
			if children != nil {
				children()
			} else {
				ui.Text(c, optionLabel(entry.option)).SingleLine()
			}
		})
		if f.manual {
			check := ui.Box(c).Absolute().End(8).TopPercent(50).Size(16, 16).Translate(0, -8).IgnoreInput(true).AccessHidden(true)
			f.manualItems = append(f.manualItems, comboItemFrame{c: c, element: option, indicator: check, key: key, active: active})
		} else if entry.selected {
			ui.Box(c).Absolute().End(8).TopPercent(50).Size(16, 16).Translate(0, -8).Children(func() { indicator.Check(c) })
		}
	})
	if option.Clicked() {
		f.choose(key)
	}
	return option
}
func (f *comboFrame) content(c *ui.Context, p ComboboxContentProps, children func()) ui.Element {
	f.check()
	f.prepare()
	f.syncItems()
	_, present := f.presence()
	if !present {
		return ui.Element{}
	}
	anchor := f.editorElement
	if f.p.PopupInput || !anchor.Valid() {
		anchor = f.triggerRoot
	}
	explicit := p.Anchor != nil
	if explicit {
		anchor = c.Resolve(*p.Anchor)
	}
	if !anchor.Valid() {
		return ui.Element{}
	}
	side := p.Side
	if side == "" {
		side = ComboboxSideBottom
	}
	align := p.Align
	if align == "" {
		align = ComboboxAlignStart
	}
	gap := float32(6)
	if p.SideOffset != nil {
		gap = *p.SideOffset
	}
	f.contentGap = gap
	at, self := comboAnchors(side, align, c.Direction())
	menu := theme.MenuOf(c).ForScene()
	t := &menu.Tokens
	var panel ui.Element
	ui.Overlay(c, func() {
		panel = ui.Themed(c, menu.Theme()).Key(f.s).Role(ui.RolePopup).AttachTo(anchor, at, self).AttachEdgeMargin(5).Clip().Radius(t.RadiusMD()).Border(1, t.Foreground.Alpha(.1)).Background(t.Popover).TextColor(t.PopoverForeground).Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).IgnoreInput(!f.s.open).AccessHidden(!f.s.open)
		menu.Surface(c, panel)
		panel.Clicked()
		f.panel = panel
		if f.p.PopupInput {
			panel.Role(ui.RoleDialog).Label(f.labelText + " popup")
			if f.triggerRoot.Valid() {
				f.triggerRoot.Controls(panel)
			}
		} else {
			panel.KeepFocus()
		}
		physical := comboPhysicalSide(side, c.Direction())
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
		bounds := anchor.LayoutBounds()
		width := bounds.W
		if !explicit {
			width += 28
		}
		if p.Width > 0 {
			width = max(width, p.Width)
		}
		avail := f.contentAvailable(panel)
		width = min(max(float32(1), width), avail.W)
		panel.Width(width).MaxWidth(avail.W)
		height := p.MaxHeight
		if height <= 0 || math.IsNaN(float64(height)) || math.IsInf(float64(height), 0) {
			height = 288
		}
		f.contentHeight = min(height, avail.H)
		panel.MaxHeight(f.contentHeight)
		if p.Style != nil {
			p.Style(panel)
		}
		previous := f.inContent
		f.inContent = true
		panel.Children(children)
		f.inContent = previous
		if f.s.open {
			outside := panel.PressedOutside()
			if panel.OverlayShortcut(0, ui.KeyEscape) || outside {
				f.s.captureClose(f.s.filterQuery(*f.query))
				f.s.open = false
				if f.p.PopupInput && !outside {
					f.s.trigger.Focus()
				}
			}
		}
	})
	f.s.motion.observe(anchor, panel)
	pose, _ := f.presence()
	panel.IgnoreInput(!f.s.open).AccessHidden(!f.s.open)
	f.s.motion.apply(c, anchor, panel, pose)
	if f.p.PopupInput && f.s.open && (f.justMounted || !f.s.wasOpen) && f.editorElement.Valid() {
		f.editorElement.Focus()
	}
	return panel
}
func comboPhysicalSide(side ComboboxSide, direction ui.Direction) ComboboxSide {
	if side == ComboboxSideInlineStart {
		if direction == ui.RightToLeft {
			return ComboboxSideRight
		}
		return ComboboxSideLeft
	}
	if side == ComboboxSideInlineEnd {
		if direction == ui.RightToLeft {
			return ComboboxSideLeft
		}
		return ComboboxSideRight
	}
	return side
}
func comboAnchors(side ComboboxSide, align ComboboxAlign, direction ui.Direction) (ui.Anchor, ui.Anchor) {
	side = comboPhysicalSide(side, direction)
	if side == ComboboxSideLeft {
		switch align {
		case ComboboxAlignCenter:
			return ui.AnchorLeft, ui.AnchorRight
		case ComboboxAlignEnd:
			return ui.AnchorBottomLeft, ui.AnchorBottomRight
		default:
			return ui.AnchorTopLeft, ui.AnchorTopRight
		}
	}
	if side == ComboboxSideRight {
		switch align {
		case ComboboxAlignCenter:
			return ui.AnchorRight, ui.AnchorLeft
		case ComboboxAlignEnd:
			return ui.AnchorBottomRight, ui.AnchorBottomLeft
		default:
			return ui.AnchorTopRight, ui.AnchorTopLeft
		}
	}
	if side == ComboboxSideTop {
		switch align {
		case ComboboxAlignCenter:
			return ui.AnchorTop, ui.AnchorBottom
		case ComboboxAlignEnd:
			return ui.AnchorTopEnd, ui.AnchorBottomEnd
		default:
			return ui.AnchorTopStart, ui.AnchorBottomStart
		}
	}
	switch align {
	case ComboboxAlignCenter:
		return ui.AnchorBottom, ui.AnchorTop
	case ComboboxAlignEnd:
		return ui.AnchorBottomEnd, ui.AnchorTopEnd
	default:
		return ui.AnchorBottomStart, ui.AnchorTopStart
	}
}
func (f *comboFrame) contentAvailable(panel ui.Element) ui.Rect {
	w, h := f.c.Size()
	available := ui.Rect{W: max(float32(1), w-10), H: max(float32(1), h-10)}
	// The first pass uses the containing window. Native AttachTo measures and
	// flips within it; later passes use its observed side and actual anchor gap.
	a, b := f.s.motion.anchorBounds, f.s.motion.panelBounds
	if b.W > 0 && b.H > 0 {
		switch f.s.motion.attachment.side {
		case comboSideTop:
			available.H = max(float32(1), a.Y-f.contentGap-5)
		case comboSideBottom:
			available.H = max(float32(1), h-a.Y-a.H-f.contentGap-5)
		case comboSideLeft:
			available.W = max(float32(1), a.X-f.contentGap-5)
		case comboSideRight:
			available.W = max(float32(1), w-a.X-a.W-f.contentGap-5)
		}
	}
	return available
}
func (f *comboFrame) finish() {
	defer func() {
		f.alive = false
		f.reveal = nil
		f.highlighted = nil
		f.completed = nil
		f.label = nil
		f.groupLabel = nil
		f.resolveSelection = nil
		f.manualEmpty, f.manualItems = nil, nil
	}()
	f.reconcileManual()
	if !f.hosted.detached && f.primary.Valid() && f.primary.FocusLeft() && !comboFocusWithin(f.inputRoot) && !comboFocusWithin(f.triggerRoot) && !comboFocusWithin(f.chipsRoot) && !comboFocusWithin(f.panel) {
		f.s.open, f.s.composing = false, false
		f.s.typeahead, f.s.typeaheadText = "", ""
	}
	if !f.hosted.detached {
		f.presence()
	}
	if f.s.wasOpen != f.s.open && !f.externalOpen && f.p.OnOpenChange != nil {
		f.p.OnOpenChange(f.s.open)
	}
	if f.boundOpen != nil {
		*f.boundOpen = f.s.open
		f.s.lastBoundOpen = *f.boundOpen
	}
	if f.s.reveal && f.virtualized && f.reveal != nil && f.s.hasActive {
		index := 0
		for _, e := range f.s.items {
			if e.option.Separator {
				continue
			}
			if e.option.Value == f.s.active {
				f.reveal(index)
				f.c.Invalidate()
				break
			}
			index++
		}
	}
	if f.highlighted != nil && (f.s.lastHasActive != f.s.hasActive || f.s.hasActive && f.s.lastActive != f.s.active) {
		f.highlighted(f.s.active, f.s.hasActive)
	}
	f.s.lastActive, f.s.lastHasActive = f.s.active, f.s.hasActive
	if f.grid && f.listElement.Valid() {
		rows, cols := 0, f.s.gridColumns
		for _, cell := range f.gridCells {
			rows, cols = max(rows, cell.row+1), max(cols, cell.column+1)
		}
		f.listElement.RowCount(rows).ColumnCount(cols)
	}
	f.s.gridCells = slices.Clone(f.gridCells)
	if f.manual && (f.manualEntries != nil || f.panel.Valid()) {
		f.s.manualOptions = slices.Clone(f.manualEntries)
	}
	if !f.virtualized || f.slotDisabled != nil {
		f.s.slotDisabled = f.slotDisabled
	}
	if f.completed != nil && f.completionReady() && (!f.s.completeInitialized || f.s.completeOpen != f.s.open) {
		f.s.completeOpen, f.s.completeInitialized = f.s.open, true
		f.completed(f.s.open)
	}
	f.s.wasOpen = f.s.open
	f.s.previousQuery = *f.query
	f.s.previousFilterQuery = f.s.filterQuery(*f.query)
	if !f.multiple {
		f.s.previousValue = *f.single
		f.s.previousLabel = f.label(f.single.Value)
	}
	if f.s.open {
		f.s.typeahead, f.s.typeaheadText = "", ""
	}
	if f.parentGroup != nil {
		f.parentGroup.items = f.parentGroup.items[:f.beforeGroupItems]
		if f.primary.Valid() {
			ButtonGroupItem(f.c, f.primary)
		}
	}
}

// reconcileManual completes only frame-owned native slots. It never calls
// the application's build/render functions a second time or waits for a frame.
func (f *comboFrame) reconcileManual() {
	if !f.manual || f.manualEntries == nil && !f.panel.Valid() {
		return
	}
	f.p.Options = slices.Clone(f.manualEntries)
	f.s.slotDisabled = f.slotDisabled
	if f.resolveSelection != nil {
		f.resolveSelection()
	}
	f.itemsSynced = false
	f.syncItems()
	f.s.hasPending = false
	count := 0
	positions := map[string]int{}
	selected := map[string]bool{}
	for _, entry := range f.s.items {
		if entry.option.Separator {
			continue
		}
		count++
		positions[entry.option.Value] = count
		selected[entry.option.Value] = entry.selected
	}
	for _, empty := range f.manualEmpty {
		if count == 0 {
			empty.element.Padding(8, 0).Children(func() { ui.Text(empty.c, empty.text) })
		} else {
			empty.element.Height(0).IgnoreInput(true).AccessHidden(true)
		}
	}
	if f.listElement.Valid() && f.manualListEmpty != (count == 0) {
		if count == 0 {
			f.listElement.Padding(0)
		} else {
			f.listElement.Padding(4)
		}
	}
	if f.editorElement.Valid() {
		f.editorElement.ActiveDescendant(ui.Element{})
	}
	for _, item := range f.manualItems {
		position, exists := positions[item.key]
		item.element.PositionInSet(position, count).AccessSelected(selected[item.key])
		if !exists {
			item.element.IgnoreInput(true).AccessHidden(true)
			continue
		}
		active := f.s.hasActive && f.s.active == item.key
		if active != item.active {
			t := theme.TokensOf(item.c)
			if active {
				background, foreground := theme.MenuOf(item.c).Highlight(false)
				item.element.Background(background).TextColor(foreground)
			} else {
				item.element.Background(ui.Transparent).TextColor(t.PopoverForeground)
			}
		}
		if active {
			if f.editorElement.Valid() {
				f.editorElement.ActiveDescendant(item.element)
			}
			if f.s.reveal {
				item.element.ScrollIntoView()
				f.s.reveal = false
			}
		}
		if selected[item.key] {
			item.indicator.Children(func() { indicator.Check(item.c) })
		}
	}
}

func comboFocusWithin(element ui.Element) bool { return element.Valid() && element.FocusWithin() }
