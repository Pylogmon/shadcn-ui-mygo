package control

import (
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/tooltip"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// StyleInput applies the kit's input appearance to a real native editor. It
// does not manufacture access to a compound SDK control's private parts.
func StyleInput(c *ui.Context, e ui.Element, p InputProps) ui.Element {
	e = styleInput(c, e, p)
	height, _ := metrics(p.Size)
	return e.Height(max(height, theme.ControlFontSize(c, inputReferenceFont(c, p.Size), p.FontSize)*inputLineHeight(p.Size)+2*inputPaddingY(p.Size)+2))
}

// NumberInputStepperProps opts into two native buttons beside the editor.
// Button props configure variants, icons and accessible Label values; omitted
// icons use minus/plus text and omitted labels use Decrease/Increase value.
// Tooltips are explicit in-window hints and also describe each button.
type NumberInputStepperProps struct {
	Decrement, Increment               ButtonProps
	DecrementTooltip, IncrementTooltip string
}

type NumberInputProps struct {
	Label, Placeholder, Description string
	Size                            Size
	// FontSize reaches the actual editor, independently of the layout root.
	FontSize float32
	// Stepper nil preserves the original editor-only layout and keyboard API.
	Stepper                                *NumberInputStepperProps
	Disabled, Invalid, ReadOnly, AutoFocus bool
	Handle                                 *ui.Handle
	Min, Max                               *float64
	// Empty optionally controls whether the field has a numeric value. While
	// true, the native editor starts empty and the numeric storage is retained.
	// Nil preserves the required-value behavior of the original API.
	Empty *bool
	// OnEmptyChange can synchronously redirect Empty and numeric storage.
	// Rejecting presence retains the old storage unless the callback replaces it.
	OnEmptyChange func(bool)
	// Zero uses one. Step governs Up/Down, not HTML step-mismatch validation.
	Step               float64
	OnChange, OnSubmit func(float64)
}
type numberInputState struct {
	draft         string
	previous      float64
	previousEmpty bool
	focused       bool
}
type numberInputEditorKey struct{}
type numberInputStepKey int

func finiteNumber(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func numberLimits(p NumberInputProps) (lo, hi, step float64, valid bool) {
	lo, hi, step = -math.MaxFloat64, math.MaxFloat64, p.Step
	if step == 0 {
		step = 1
	}
	if p.Min != nil {
		lo = *p.Min
	}
	if p.Max != nil {
		hi = *p.Max
	}
	return lo, hi, step, finiteNumber(lo) && finiteNumber(hi) && lo <= hi && finiteNumber(step) && step > 0
}
func parseInputNumber(text string, lo, hi float64) (float64, bool) {
	n, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	return n, err == nil && finiteNumber(n) && n >= lo && n <= hi
}
func formatInputNumber(n float64) string { return strconv.FormatFloat(n, 'f', -1, 64) }

func nextInputNumber(value float64, empty bool, lo, hi, delta float64) (float64, bool) {
	base := value
	if empty {
		base = 0
	}
	if !finiteNumber(base) {
		return value, false
	}
	next := math.Max(lo, math.Min(hi, base+delta))
	return next, finiteNumber(next) && (empty || next != value)
}

// NumberInput is a finite typed adapter around the same native text/IME editor
// as Input. Editable draft and committed numeric value are separate. Incomplete
// or out-of-range text keeps the last value; Submit/blur canonicalizes it. Handle
// targets the actual editor, and Field/InputGroup metadata reaches that editor.
// The returned element is the layout root. Range metadata is published when both
// bounds are supplied; this is not a browser spinbutton/constraint-bubble engine.
func NumberInput(c *ui.Context, value *float64, p NumberInputProps) ui.Element {
	if value == nil {
		panic("control: NumberInput requires numeric state")
	}
	root := ui.Box(c).FillWidth().MinWidth(0)
	if p.Stepper != nil {
		root.Row().Gap(4).AlignItems(ui.Center)
	}
	isEmpty := func() bool { return p.Empty != nil && *p.Empty }
	modelText := func() string {
		if isEmpty() {
			return ""
		}
		return formatInputNumber(*value)
	}
	s := ui.Local(root, "number-draft", func() numberInputState {
		return numberInputState{draft: modelText(), previous: *value, previousEmpty: isEmpty()}
	})
	if isEmpty() != s.previousEmpty || !isEmpty() && *value != s.previous {
		s.draft = modelText()
		s.previous = *value
		s.previousEmpty = isEmpty()
	}
	lo, hi, step, configurationValid := numberLimits(p)
	before, beforeEmpty := *value, isEmpty()
	services := c.Services()
	root.Children(func() {
		editor := ui.TextInputBase(c.Key(numberInputEditorKey{}), &s.draft)
		// The editor belongs inside the component root, keeping independent keyed
		// draft/selection state when several controls share the same numeric model.
		changed := editor.Changed()
		composing := editor.Composing()
		editable := configurationValid && !p.Disabled && !p.ReadOnly && !editor.IsDisabled() && !composing
		remember := func() { s.previous, s.previousEmpty = *value, isEmpty() }
		setEmpty := func(empty bool) {
			if p.Empty == nil || *p.Empty == empty {
				return
			}
			*p.Empty = empty
			if p.OnEmptyChange != nil {
				p.OnEmptyChange(empty)
			}
		}
		commit := func(n float64) {
			wasEmpty := isEmpty()
			if n == *value && !wasEmpty {
				return
			}
			previousValue := *value
			*value = n
			setEmpty(false)
			// A presence callback can reject this commit or replace its numeric
			// storage. Roll back only our candidate, preserving a callback's value.
			if isEmpty() && *value == n {
				*value = previousValue
			}
			if !isEmpty() && p.OnChange != nil {
				p.OnChange(n)
			}
			if isEmpty() || *value != n {
				s.draft = modelText()
			}
			remember()
		}
		stepBy := func(delta float64) {
			if !editable {
				return
			}
			if n, ok := nextInputNumber(*value, isEmpty(), lo, hi, delta); finiteNumber(n) {
				if ok {
					commit(n)
				}
				s.draft = modelText()
				c.Invalidate()
			}
		}
		if changed && editable {
			if p.Empty != nil && strings.TrimSpace(s.draft) == "" {
				setEmpty(true)
				if !isEmpty() {
					s.draft = modelText()
				}
				remember()
			} else if n, ok := parseInputNumber(s.draft, lo, hi); ok {
				commit(n)
			}
		}
		if editable {
			if editor.Shortcut(0, ui.KeyUp) {
				stepBy(step)
			} else if editor.Shortcut(0, ui.KeyDown) {
				stepBy(-step)
			}
		}
		focused := editor.Focused()
		if s.focused && !focused && !composing {
			s.draft = modelText()
			c.Invalidate()
		}
		s.focused = focused
		_, draftValid := parseInputNumber(s.draft, lo, hi)
		invalid := p.Invalid || !configurationValid || (!isEmpty() && !finiteNumber(*value)) || (!composing && strings.TrimSpace(s.draft) != "" && !draftValid)
		common := InputProps{Label: p.Label, Placeholder: p.Placeholder, Description: p.Description, Size: p.Size, FontSize: p.FontSize, Disabled: p.Disabled, Invalid: invalid, ReadOnly: p.ReadOnly || !configurationValid, AutoFocus: p.AutoFocus, Handle: p.Handle}
		common.OnSubmit = func(string) {
			n, ok := parseInputNumber(s.draft, lo, hi)
			if ok && editable {
				commit(n)
				s.draft = modelText()
				if !isEmpty() && p.OnSubmit != nil {
					p.OnSubmit(*value)
				}
			} else {
				s.draft = modelText()
			}
			services.Invalidate()
		}
		StyleInput(c, editor, common)
		configureTextInput(c, editor, &s.draft, common)
		if p.Min != nil && p.Max != nil && configurationValid && !isEmpty() {
			editor.Range(lo, hi, *value).Step(step)
		}
		handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
		editor.Bind(handle)
		if p.Stepper != nil {
			editor.Grow(1)
			ui.Row(c).Shrink(0).Gap(2).Children(func() {
				stepButton := func(direction numberInputStepKey, props ButtonProps, hint string) {
					delta, text, label := step, "+", "Increase value"
					if direction < 0 {
						delta, text, label = -step, "−", "Decrease value"
					}
					_, canStep := nextInputNumber(*value, isEmpty(), lo, hi, delta)
					props.Disabled = props.Disabled || !editable || !canStep
					if props.Label == "" {
						props.Label = label
					}
					if props.Size == DefaultSize {
						props.Size = numberStepperSize(p.Size)
					}
					if props.FontSize == 0 {
						props.FontSize = p.FontSize
					}
					if props.Icon != nil {
						text = ""
					}
					button := Button(c.Key(direction), text, props)
					if hint != "" {
						tooltip.InWindow(c, button, hint, false, nil)
					}
					if button.Clicked() && !button.IsDisabled() {
						stepBy(delta)
						editor.Focus()
					}
				}
				stepButton(-1, p.Stepper.Decrement, p.Stepper.DecrementTooltip)
				stepButton(1, p.Stepper.Increment, p.Stepper.IncrementTooltip)
			})
		}
	})
	if isEmpty() != beforeEmpty || *value != before && !(math.IsNaN(*value) && math.IsNaN(before)) {
		root.SetChanged()
	}
	return root
}

func numberStepperSize(size Size) Size {
	switch size {
	case ExtraSmall, IconExtraSmall:
		return IconExtraSmall
	case Small, IconSmall:
		return IconSmall
	case Large, IconLarge:
		return IconLarge
	default:
		return Icon
	}
}

// DateInputProps configures a native localized date field.
type DateInputProps struct {
	Label, Description           string
	Size                         Size
	Disabled, Invalid, AutoFocus bool
	Handle                       *ui.Handle
	OnChange                     func(time.Time)
}

// TimeInputProps configures a native segmented time field. Handle targets its
// group. The SDK does not expose individual segment handles or autofocus.
type TimeInputProps struct {
	Label, Description string
	Size               Size
	Disabled, Invalid  bool
	Handle             *ui.Handle
	OnChange           func(time.Time)
}

func dateTimeAppearance(c *ui.Context, e ui.Element, p InputProps, handle *ui.Handle) ui.Element {
	StyleInput(c, e, p)
	if handle != nil {
		e.Bind(handle)
	}
	if p.AutoFocus && !e.IsDisabled() {
		e.AutoFocus()
	}
	return e
}

// DateInput retains the public SDK's localized calendar, keyboard behavior,
// time-of-day and Location. It binds a date value rather than browser text.
func DateInput(c *ui.Context, value *time.Time, p DateInputProps) ui.Element {
	if value == nil {
		panic("control: DateInput requires date state")
	}
	e := dateTimeAppearance(c, ui.DateInput(c, value), InputProps{Label: p.Label, Description: p.Description, Size: p.Size, Disabled: p.Disabled, Invalid: p.Invalid, AutoFocus: p.AutoFocus}, p.Handle)
	if p.OnChange != nil {
		e.OnChange(func() { p.OnChange(*value) })
	}
	return e
}

// TimeInput retains the native localized segments (including the user's hour
// cycle). Segment edits preserve the date and Location. Group label/invalid/
// description metadata belongs to the group; segment names use its Label. The
// SDK exposes no segment parts, and no ineffective ReadOnly option is offered.
func TimeInput(c *ui.Context, value *time.Time, p TimeInputProps) ui.Element {
	if value == nil {
		panic("control: TimeInput requires time state")
	}
	e := dateTimeAppearance(c, ui.TimeInput(c, value), InputProps{Label: p.Label, Description: p.Description, Size: p.Size, Disabled: p.Disabled, Invalid: p.Invalid}, p.Handle)
	if p.OnChange != nil {
		e.OnChange(func() { p.OnChange(*value) })
	}
	return e
}

// FilePicker is application-owned, runs off the UI thread and may block. It
// receives only dialog values/a native Window, never a Context or Element.
type FilePicker func(mygo.OpenDialogOptions) ([]string, error)
type FileInputProps struct {
	Label, Description, Placeholder, ChooseLabel string
	Size                                         Size
	Disabled, Invalid                            bool
	Handle                                       *ui.Handle
	Parent                                       *mygo.Window
	Title                                        string
	Filters                                      []mygo.FileFilter
	Multiple                                     bool
	Picker                                       FilePicker
	OnChange                                     func([]string)
	OnError                                      func(error)
}
type fileInputResult struct {
	paths []string
	err   error
}
type fileInputState struct {
	mu           sync.Mutex
	closed, busy bool
	generation   uint64
	bound        *[]string
	result       *fileInputResult
	err          error
}

func (s *fileInputState) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.generation++
	s.result = nil
	return nil
}
func cloneFileFilters(filters []mygo.FileFilter) []mygo.FileFilter {
	out := slices.Clone(filters)
	for i := range out {
		out[i].Extensions = slices.Clone(out[i].Extensions)
		out[i].MIMETypes = slices.Clone(out[i].MIMETypes)
	}
	return out
}

// FileInput opens a real native picker only after pointer/keyboard activation.
// Cancel preserves selection. An asynchronous result is applied in a current
// owner build; removing/rebinding the control discards stale results/callbacks.
func FileInput(c *ui.Context, value *[]string, p FileInputProps) ui.Element {
	if value == nil {
		panic("control: FileInput requires file path state")
	}
	root := ui.ButtonBase(c).FillWidth().MinWidth(0)
	s := ui.Local(root, "file-picker", func() fileInputState { return fileInputState{bound: value} })
	s.mu.Lock()
	if s.bound != value {
		s.bound = value
		s.generation++
		s.busy = false
		s.result = nil
		s.err = nil
	}
	result := s.result
	s.result = nil
	if result != nil {
		s.busy = false
		s.err = result.err
	}
	busy, operationError := s.busy, s.err
	s.mu.Unlock()
	if result != nil {
		if result.err != nil {
			if p.OnError != nil {
				p.OnError(result.err)
			}
		} else if len(result.paths) > 0 && !slices.Equal(*value, result.paths) {
			*value = slices.Clone(result.paths)
			root.SetChanged()
			if p.OnChange != nil {
				p.OnChange(slices.Clone(*value))
			}
		}
	}
	StyleInput(c, root, InputProps{Label: p.Label, Description: p.Description, Size: p.Size, Disabled: p.Disabled || busy, Invalid: p.Invalid || operationError != nil})
	root.Cursor(ui.CursorPointer).Busy(busy).AlignItems(ui.Center).Justify(ui.Start).Gap(10)
	if p.Handle != nil {
		root.Bind(p.Handle)
	}
	choose := p.ChooseLabel
	if choose == "" {
		choose = "Choose file"
		if p.Multiple {
			choose = "Choose files"
		}
	}
	names := make([]string, 0, len(*value))
	for _, path := range *value {
		names = append(names, filepath.Base(path))
	}
	text := strings.Join(names, ", ")
	if text == "" {
		text = p.Placeholder
		if text == "" {
			text = "No file selected"
		}
	}
	description := p.Description
	if len(names) > 0 {
		description = strings.TrimSpace(description + " Selected: " + strings.Join(names, ", "))
	}
	if operationError != nil {
		description = strings.TrimSpace(description + " " + operationError.Error())
	}
	root.Description(description)
	root.Children(func() {
		ui.Box(c).Shrink(0).Padding(2, 6).Children(func() { ui.Text(c, choose).FontWeight(500) })
		ui.Text(c, text).MinWidth(0).Shrink(1).SingleLine()
	})
	if root.Clicked() && !p.Disabled && !busy {
		s.mu.Lock()
		s.busy = true
		s.generation++
		generation := s.generation
		s.mu.Unlock()
		root.Busy(true)
		options := mygo.OpenDialogOptions{Parent: p.Parent, Title: p.Title, Filters: cloneFileFilters(p.Filters), Multiple: p.Multiple}
		picker := p.Picker
		if picker == nil {
			picker = mygo.Dialog.Open
		}
		services := c.Services()
		go func() {
			paths, err := picker(options)
			paths = slices.Clone(paths)
			if !options.Multiple && len(paths) > 1 {
				paths = paths[:1]
			}
			s.mu.Lock()
			if !s.closed && s.generation == generation {
				s.result = &fileInputResult{paths: paths, err: err}
			}
			s.mu.Unlock()
			services.Invalidate()
		}()
		c.Invalidate()
	}
	return root
}
