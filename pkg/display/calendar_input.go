package display

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// DatePickerInputState is application-owned text, displayed month and open
// state. Initialize Text explicitly, as in the fixed controlled-input recipes.
// Keep its address stable while mounted. Error records the last parse attempt;
// invalid text preserves the caller's last valid date and displayed month.
type DatePickerInputState struct {
	Text  string
	Month time.Time
	Open  bool
	Error string
}

// DateParser is supplied by the application. The kit does not infer a language,
// evaluate JavaScript Date strings or embed a natural-language parser. The
// returned time keeps the parser's instant, location and wall-clock fields.
type DateParser func(string) (time.Time, error)

type DatePickerInputProps struct {
	Input        control.InputProps
	Calendar     CalendarProps
	Parse        DateParser
	Format       func(time.Time) string
	TriggerLabel string
	// Nil uses the fixed editable recipe's bottom-end placement, 10-DIP gap
	// and -8-DIP alignment offset. The natural-language recipe uses 8 and 0.
	SideOffset, AlignOffset *float32
	// OnChange runs in the current build only after an accepted date changes.
	OnChange func(time.Time)
}

type DatePickerInputParts struct {
	Root, Editor, Trigger, Popup ui.Element
	// ValueChanged distinguishes a date publication from raw editing. Root's
	// native Changed also reports raw text edits through the InputGroup.
	ValueChanged bool
}

type dateInputMountKey struct{ state *DatePickerInputState }
type dateInputRuntime struct {
	editor, trigger                   ui.Handle
	openRequested, enabled, composing bool
	marked, commitPending             bool
	nativeMarkedSeen                  bool
	parsed, closed                    bool
	lastText                          string
}

func (s *dateInputRuntime) Close() error { *s = dateInputRuntime{closed: true}; return nil }
func (s *dateInputRuntime) input(event ui.InputEvent) bool {
	if !s.enabled || s.composing || event.Kind != ui.InputKeyDown || event.Mods != 0 {
		return false
	}
	if event.Key == ui.KeyDown {
		s.openRequested = true
		return true
	}
	return false
}
func (s *dateInputRuntime) observe(event ui.InputEvent) {
	switch event.Kind {
	case ui.InputCompose:
		s.marked = event.Text != ""
		if !s.marked {
			s.commitPending = true
		}
	case ui.InputText:
		s.marked, s.commitPending = false, true
	}
}

// DatePickerInput composes a real native editor and CalendarWith. Parsing is a
// current-build application callback; no Context, Element or callback is saved
// in its mounted runtime. Editor remains mounted while browsing the popup, so
// native selection, undo and Field/OnSubmit integration remain available.
func DatePickerInput(c *ui.Context, date *time.Time, state *DatePickerInputState, p DatePickerInputProps) DatePickerInputParts {
	if date == nil || state == nil || p.Parse == nil {
		panic("display.DatePickerInput: date, state and Parse are required")
	}
	mount := ui.Box(c).Key(dateInputMountKey{state}).Absolute().Size(0, 0).AccessHidden(true).IgnoreInput(true)
	runtime := ui.Local(mount, "date-picker-input", func() dateInputRuntime { return dateInputRuntime{} })
	runtime.enabled = !p.Input.Disabled && !p.Input.ReadOnly
	if runtime.openRequested {
		state.Open = runtime.enabled
		runtime.openRequested = false
	}
	if !runtime.enabled {
		state.Open = false
	}
	model := modelFor(p.Calendar.System)
	format := p.Format
	if format == nil {
		format = func(value time.Time) string {
			if value.IsZero() {
				return ""
			}
			if p.Calendar.System == CalendarPersian {
				return model.label(c.Locale(), value, false)
			}
			return value.Format("January 02, 2006")
		}
	}
	triggerLabel := p.TriggerLabel
	if triggerLabel == "" {
		triggerLabel = "Select date"
	}
	parts := DatePickerInputParts{}
	inputProps := p.Input
	inputProps.Invalid = inputProps.Invalid || state.Error != ""
	parts.Root = control.InputGroup(c, control.InputGroupProps{Disabled: p.Input.Disabled, Invalid: inputProps.Invalid, InlineEnd: func() {
		parts.Trigger = control.InputGroupButton(c, "", control.ButtonProps{Variant: control.Ghost, Size: control.IconExtraSmall, Label: triggerLabel, Icon: outlined.CalendarMonth(), Disabled: !runtime.enabled}).Bind(&runtime.trigger).HasPopup(ui.PopupDialog).Expanded(state.Open)
		if parts.Trigger.Clicked() {
			state.Open = !state.Open
		}
	}}, func() ui.Element {
		parts.Editor = control.Input(c, &state.Text, inputProps).Bind(&runtime.editor).HandleInput(runtime.input)
		return parts.Editor
	})
	// Observe the ancestor, preserving control.Input's own native composition/
	// submit observer. Preedit can temporarily delete a selected span before
	// the native editor exposes Composing; the observed marker covers that pass.
	parts.Root.ObserveInput(runtime.observe)
	nativeComposing := parts.Editor.Composing()
	if nativeComposing {
		runtime.nativeMarkedSeen = true
	}
	// Native Enter may finish a marked range without a separate InputText.
	// The initial deletion pass is masked until native marked state was seen.
	if runtime.marked && runtime.nativeMarkedSeen && !nativeComposing {
		runtime.marked = false
		runtime.commitPending = true
	}
	composing := runtime.marked || nativeComposing
	committed := !composing && (parts.Editor.Changed() || runtime.composing || runtime.commitPending)
	if committed && (!runtime.parsed || state.Text != runtime.lastText || runtime.composing) {
		runtime.lastText, runtime.parsed = state.Text, true
		parsed, err := p.Parse(state.Text)
		if err == nil && !parsed.IsZero() && unselectableDate(parsed, p.Calendar) {
			err = fmt.Errorf("Date is not available")
		}
		previousError := state.Error
		state.Error = ""
		if err != nil {
			state.Error = err.Error()
		} else {
			parts.ValueChanged = !pickerInputEqual(*date, parsed)
			*date = parsed
			if !parsed.IsZero() {
				state.Month = parsed
			}
			if parts.ValueChanged && p.OnChange != nil {
				p.OnChange(parsed)
			}
		}
		if previousError != state.Error {
			c.Invalidate()
		}
	}
	runtime.composing = composing
	if !composing {
		runtime.commitPending = false
		runtime.nativeMarkedSeen = false
	}
	parts.Editor.Invalid(p.Input.Invalid || state.Error != "")
	t := theme.TokensOf(c)
	parts.Popup = ui.PopoverBase(c, parts.Trigger, &state.Open, func(panel ui.Element) {
		gap, offset := float32(10), float32(-8)
		if p.SideOffset != nil {
			gap = *p.SideOffset
		}
		if p.AlignOffset != nil {
			offset = *p.AlignOffset
		}
		if c.Direction() == ui.RightToLeft {
			offset = -offset
		}
		panel.Role(ui.RoleDialog).Label(triggerLabel+" calendar").AttachTo(parts.Trigger, ui.AnchorBottomEnd, ui.AnchorTopEnd).Left(offset).Margin(gap, 0, 0, 0).Background(t.Popover).TextColor(t.PopoverForeground).Border(1, t.Border).Radius(t.RadiusMD()).Transition(motion.Popup(c))
		panel.Children(func() {
			working := *date
			props := p.Calendar
			props.Month = &state.Month
			props.Browse, props.AutoFocus = true, true
			if CalendarWith(c, &working, props).Changed() {
				parts.ValueChanged = !pickerInputEqual(*date, working)
				*date = working
				state.Text, state.Error = format(working), ""
				runtime.lastText, runtime.parsed = state.Text, true
				state.Open = false
				if parts.ValueChanged && p.OnChange != nil {
					p.OnChange(working)
				}
				parts.Root.SetChanged()
				parts.Trigger.Focus()
			}
		})
	})
	if parts.Popup.Valid() {
		parts.Trigger.Controls(parts.Popup)
		if parts.Popup.Dismissed() && !parts.Popup.PressedOutside() {
			parts.Trigger.Focus()
		}
		if state.Open && parts.Popup.FocusLeft() && !parts.Popup.FocusWithin() && !parts.Trigger.FocusWithin() {
			state.Open = false
			c.Invalidate()
		}
	}
	parts.Trigger.Expanded(state.Open)
	return parts
}

func pickerInputEqual(a, b time.Time) bool { return a.Equal(b) && a.Location() == b.Location() }
