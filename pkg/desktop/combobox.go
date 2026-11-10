package desktop

import (
	"errors"
	"math"
	"slices"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// ComboboxView is a window-owned props provider, like a Popover child view.
// Capture application models only. Each call receives the CURRENT parent or
// child Context; do not capture a parent Context, Element or ComboboxScope.
// The returned adapters, filters and handlers are used synchronously in that
// build and are never retained by the host's application model.
type ComboboxView[T any] func(window *ui.Context) control.ComboboxRootProps[T]

type ComboboxProps struct {
	Label, Placeholder, SearchPlaceholder, EmptyText string
	Size                                             control.Size
	Invalid, ShowClear                               bool
	Button                                           control.ButtonProps
	// Popup.Revision (or Host.Refresh) refreshes changes to provider behavior
	// whose visible data are otherwise unchanged, such as a new Filter policy.
	Popup        PopoverProps
	SearchHandle *ui.Handle
	// OnTab runs in the next current parent build after accepting the close.
	// MyGo has no cross-window Tab-forward operation; focus an adjacent public
	// Handle here. Without it, Tab returns to the trigger, as ordinary dismissal.
	OnTab func(backward bool)
}

type comboDescriptor struct {
	key, label, form, group string
	disabled                bool
}
type comboHostState[T any] struct {
	owner                                      *control.ComboboxState[T]
	rows                                       []comboDescriptor
	flags                                      comboHostFlags
	query                                      string
	revision                                   uint64
	valueChanged, requestDismiss, nativeClosed bool
	session, closing                           bool
	tab                                        int
	selection                                  []string
	child                                      comboChildProps
}
type comboHostFlags struct {
	multiple, disabled, readonly, noLoop, controlled bool
	auto                                             control.ComboboxAutoHighlight
	locale                                           string
}

// Combobox opens an editable, shadcn-styled popup in an independent Flyout.
// State is application-owned; selected typed values are updated in the child
// event, while the parent trigger reports Changed once on its next build.
func Combobox[T any](c *ui.Context, h *Host, key string, state *control.ComboboxState[T], p ComboboxProps, view ComboboxView[T]) (ui.Element, error) {
	return hostedCombobox(c, h, key, state, p, view, false)
}

// NativeCombobox uses an independent plain Flyout with the child's platform
// theme and native base controls. It is not an operating-system ComboBox or
// static OS menu. Standard MyGo Flyouts can intercept an outside dismissal,
// allowing the current OnOpenChange handler to cancel without losing the editor.
func NativeCombobox[T any](c *ui.Context, h *Host, key string, state *control.ComboboxState[T], p ComboboxProps, view ComboboxView[T]) (ui.Element, error) {
	return hostedCombobox(c, h, key, state, p, view, true)
}

func hostedCombobox[T any](c *ui.Context, h *Host, key string, state *control.ComboboxState[T], p ComboboxProps, view ComboboxView[T], native bool) (ui.Element, error) {
	if c == nil || !c.Valid() || key == "" || state == nil || view == nil {
		return ui.Element{}, errors.New("desktop: Combobox needs a live Context, stable key, state and window-owned view")
	}
	root := ui.Column(c.Key("desktop-combobox:" + key)).FillWidth().MinWidth(0)
	s := ui.Local(root, "hosted-combobox", func() comboHostState[T] { return comboHostState[T]{owner: state} })
	if s.owner != state {
		if h != nil {
			h.Close(key)
		}
		*s = comboHostState[T]{owner: state}
	}
	props := view(c)
	if props.Adapter.Key == nil {
		return ui.Element{}, errors.New("desktop: Combobox view must provide Adapter.Key")
	}
	if props.Grid || props.Virtualized {
		return ui.Element{}, errors.New("desktop: Combobox window view uses the normal collection; compose Grid/Virtualized in control.ComboboxRoot")
	}
	rows := comboHostDescriptors(props)
	selection := comboHostSelection(props, state)
	flags := comboHostFlags{props.Multiple, props.Disabled, props.ReadOnly, props.NoLoop, props.QueryControlled, props.AutoHighlight, props.Locale}
	if !slices.Equal(rows, s.rows) || !slices.Equal(selection, s.selection) || flags != s.flags || s.query != state.Query {
		s.rows, s.flags, s.query, s.selection = rows, flags, state.Query, selection
		s.revision++
	}
	if s.nativeClosed && s.requestDismiss {
		state.Open = true
	}
	request := s.requestDismiss
	s.requestDismiss, s.nativeClosed = false, false
	// The parent has no editor/list: its job is closed typeahead and the trigger.
	// Its own filtering must never observe the child's IME preedit or duplicate
	// active-descendant/highlight/completion events from the editable child.
	parent := props
	parent.Filter = func(T, string) bool { return true }
	parent.OnHighlighted, parent.OnOpenChangeComplete = nil, nil
	var anchor ui.Element
	var popupErr error
	root.Children(func() {
		control.ComboboxRoot(c, state, parent, func(b *control.ComboboxScope[T]) {
			button := p.Button
			button.Size, button.Invalid = p.Size, p.Invalid
			if button.Variant == control.Default {
				button.Variant = control.Outline
			}
			trigger := control.ComboboxTriggerProps{Label: p.Label, Button: button}
			if native {
				trigger.Render = func(current *ui.Context) ui.Element {
					return ui.Button(current, "").Disabled(props.Disabled || props.ReadOnly || button.Disabled).Invalid(p.Invalid)
				}
			}
			anchor = b.HostedTrigger(c, trigger, request, func() {
				ui.Row(c).Grow(1).MinWidth(0).Children(func() {
					placeholder := p.Placeholder
					if placeholder == "" {
						placeholder = "Select an option"
					}
					b.Value(c, control.ComboboxValueProps{Placeholder: placeholder})
				})
			}).FillWidth()
			if native {
				anchor.DimDisabled(true)
			}
			if s.valueChanged {
				anchor.SetChanged()
				root.SetChanged()
				s.valueChanged = false
			}
			acceptedDismiss := request && !state.Open
			if acceptedDismiss && p.Popup.OnDismiss != nil {
				p.Popup.OnDismiss()
			}
			if h == nil {
				if state.Open {
					state.Open = false
					anchor.Expanded(false)
					popupErr = errors.New("desktop: opening a Combobox needs a parent-bound Host")
				}
				return
			}
			popup := p.Popup
			popup.Plain, popup.SystemStyle, popup.ManualTrigger = true, native, true
			popup.NonFocusable = false
			popup.ContentRole = ui.RoleDialog
			if popup.Label == "" {
				popup.Label = p.Label + " popup"
			}
			if popup.Placement == "" {
				popup.Placement = mygo.PlacementBottomStart
				if c.Direction() == ui.RightToLeft {
					popup.Placement = mygo.PlacementBottomEnd
				}
			}
			if popup.Width <= 0 {
				popup.Width = max(280, int(math.Ceil(float64(anchor.LayoutBounds().W+28))))
			}
			if popup.Height <= 0 {
				popup.Height = 200
			}
			if popup.MaxHeight <= 0 {
				popup.MaxHeight = 288
			}
			popup.FitContent = true
			if popup.Gap == 0 {
				popup.Gap = 6
			}
			zero := float32(0)
			if popup.ContentPadding == nil {
				popup.ContentPadding = &zero
			}
			if !native {
				popup.menuMotion = menuMotionPopup
				popup.menuIdentity = "combobox:" + key
			}
			popup.NoFocusRestore = popup.NoFocusRestore || s.tab != 0 && p.OnTab != nil
			popup.OnDismiss = func() { s.requestDismiss, s.nativeClosed = true, true }
			// Only the explicit SDK-owned provider and pure view values cross this
			// boundary. No returned parent props, Scope or Element is captured.
			childProps := comboChildProps{label: p.Label, search: p.SearchPlaceholder, empty: p.EmptyText, size: p.Size, invalid: p.Invalid, clear: p.ShowClear, height: float32(popup.MaxHeight), native: native, handle: p.SearchHandle}
			if childProps != s.child {
				s.child = childProps
				s.revision++
			}
			popup.Revision += s.revision
			popupErr = h.Popover(c, key, anchor, &state.Open, popup, func(child *ui.Context) {
				build := func() { buildHostedCombobox(child, state, s, childProps, view) }
				if native {
					theme.NativeScope(child).Children(build)
				} else {
					build()
				}
			})
			installComboDismiss(h, key, s)
			entry := h.entries[key]
			gone := entry == nil || entry.surface == nil
			if state.Open {
				s.session = true
				s.closing = false
			}
			if !state.Open && s.session {
				s.closing = true
				if gone {
					b.ResetHostedQuery()
					if props.OnOpenChangeComplete != nil {
						props.OnOpenChangeComplete(false)
					}
					s.session, s.closing = false, false
					if s.tab != 0 && p.OnTab != nil {
						p.OnTab(s.tab < 0)
					}
					s.tab = 0
				}
			}
			if state.Open && s.tab != 0 {
				s.tab = 0
			}
			anchor.Expanded(state.Open)
		})
	})
	return anchor, popupErr
}

type comboChildProps struct {
	label, search, empty   string
	size                   control.Size
	invalid, clear, native bool
	height                 float32
	handle                 *ui.Handle
}

func buildHostedCombobox[T any](c *ui.Context, state *control.ComboboxState[T], s *comboHostState[T], p comboChildProps, view ComboboxView[T]) {
	props := view(c)
	valueChange, valuesChange, complete := props.OnValueChange, props.OnValuesChange, props.OnOpenChangeComplete
	props.OnValueChange = func(v control.ComboboxSelection[T], event *control.ComboboxChange) {
		if valueChange != nil {
			valueChange(v, event)
		}
		if !event.Canceled() {
			s.valueChanged = true
		}
	}
	props.OnValuesChange = func(v []T, event *control.ComboboxChange) {
		if valuesChange != nil {
			valuesChange(v, event)
		}
		if !event.Canceled() {
			s.valueChanged = true
		}
	}
	props.OnOpenChangeComplete = func(open bool) {
		if open && complete != nil {
			complete(true)
		}
	}
	control.ComboboxRoot(c, state, props, func(b *control.ComboboxScope[T]) {
		frame := menuFrameOf(c)
		panel := b.HostedContent(c, control.ComboboxHostedContentProps{Label: p.label, Native: p.native, MaxHeight: p.height, OpenComplete: p.native || !frame.active || frame.pose == menuPoseShown, OnTab: func(backward bool) {
			s.tab = 1
			if backward {
				s.tab = -1
			}
		}}, func() {
			input := b.Input(c, control.ComboboxInputProps{Label: p.label + " search", Placeholder: p.search, Size: p.size, Invalid: p.invalid, HideTrigger: true, ShowClear: p.clear})
			if p.handle != nil {
				input.Editor.Bind(p.handle)
			}
			empty := p.empty
			if empty == "" {
				empty = "No items found."
			}
			b.Empty(c, empty)
			b.List(c, control.ComboboxListProps{Label: p.label + " options"}, func() {
				if len(props.Groups) > 0 {
					b.Groups(c, func(current *ui.Context, group control.ComboboxGroupData[T], _ int) {
						b.Group(current, group, func() {
							b.Label(current, group.Label)
							b.Collection(current, func(current *ui.Context, value T, _ int) { b.Item(current, value, control.ComboboxItemProps{}, nil) })
						})
					})
				} else {
					b.Collection(c, func(current *ui.Context, value T, _ int) { b.Item(current, value, control.ComboboxItemProps{}, nil) })
				}
			})
		})
		if frame.active && !p.native {
			menuApplyPose(panel, frame.pose, frame.attachment)
		}
	})
}

func installComboDismiss[T any](h *Host, key string, s *comboHostState[T]) {
	e := h.entries[key]
	if e == nil || e.surface == nil {
		return
	}
	interceptor, ok := e.surface.(DismissInterceptor)
	if !ok {
		return
	}
	surface, generation := e.surface, e.generation
	interceptor.SetDismissHandler(func() bool {
		if e.generation != generation || e.surface != surface {
			return false
		}
		if e.closing {
			// The current parent already accepted this close. Preserve the Host's
			// exit and intercept repeated outside gestures without queuing a
			// second cancelable close when that exit eventually completes.
			return true
		}
		if e.open == nil || !*e.open {
			return false
		}
		s.requestDismiss = true
		h.backend.InvalidateParent()
		return true
	})
}

func comboHostDescriptors[T any](p control.ComboboxRootProps[T]) []comboDescriptor {
	var rows []comboDescriptor
	add := func(value T, group string) {
		key := p.Adapter.Key(value)
		label, form := key, key
		if p.Adapter.Label != nil {
			label = p.Adapter.Label(value)
		}
		if p.Adapter.FormValue != nil {
			form = p.Adapter.FormValue(value)
		}
		disabled := false
		if p.Adapter.Disabled != nil {
			disabled = p.Adapter.Disabled(value)
		}
		rows = append(rows, comboDescriptor{key, label, form, group, disabled})
	}
	for _, v := range p.Items {
		add(v, "")
	}
	for _, g := range p.Groups {
		rows = append(rows, comboDescriptor{group: g.Key, label: g.Label})
		for _, v := range g.Items {
			add(v, g.Key)
		}
	}
	return rows
}

func comboHostSelection[T any](p control.ComboboxRootProps[T], state *control.ComboboxState[T]) []string {
	out := []string{strconv.FormatBool(state.Value.Valid), strconv.Itoa(len(state.Values))}
	add := func(value T) {
		key, label := p.Adapter.Key(value), p.Adapter.Key(value)
		if p.Adapter.Label != nil {
			label = p.Adapter.Label(value)
		}
		out = append(out, key, label)
		equal := func(other T) {
			same := p.Adapter.Key(other) == key
			if p.Adapter.Equal != nil {
				same = p.Adapter.Equal(other, value)
			}
			out = append(out, strconv.FormatBool(same))
		}
		for _, item := range p.Items {
			equal(item)
		}
		for _, group := range p.Groups {
			for _, item := range group.Items {
				equal(item)
			}
		}
	}
	if p.Multiple {
		for _, v := range state.Values {
			add(v)
		}
	} else if state.Value.Valid {
		add(state.Value.Value)
	}
	return out
}
