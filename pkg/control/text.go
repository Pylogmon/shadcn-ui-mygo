package control

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type InputProps struct {
	Placeholder, Label, Description string
	Size                            Size
	// FontSize reaches the native editor; zero uses the theme and Size ratio.
	FontSize                              float32
	Disabled, Invalid, ReadOnly, Password bool
	Search, AutoFocus                     bool
	// MaxLength uses the native editor's UTF-16 limit for typed/pasted/IME text.
	// Application-owned values are not truncated; zero means no limit.
	MaxLength int
	Handle    *ui.Handle
	// Callbacks run with native bound-value input. OnPaste returns true to
	// consume a paste; OnTextKey sees non-text keys before the native editor.
	OnChange, OnSubmit func(string)
	OnPaste            func(string) bool
	OnTextKey          func(ui.TextKey) bool
}

type fieldAppearanceKey struct{}
type fieldAppearanceScope struct{ invalid bool }
type textCompositionKey struct{}
type textCompositionState struct{ composing, suppressSubmit bool }
type inputGroupAppearanceKey struct{}
type inputGroupAppearance struct{ active []*bool }

func inputGroupEditor(c *ui.Context, build func() ui.Element) (ui.Element, bool) {
	scope := ui.Local(c.Root(), inputGroupAppearanceKey{}, func() inputGroupAppearance { return inputGroupAppearance{} })
	before := len(scope.active)
	invalid := false
	scope.active = append(scope.active, &invalid)
	defer func() { scope.active = scope.active[:before] }()
	return build(), invalid
}

func fieldInputInvalid(c *ui.Context, explicit bool) bool {
	scope := ui.Local(c.Root(), fieldAppearanceKey{}, func() fieldAppearanceScope { return fieldAppearanceScope{} })
	return explicit || scope.invalid
}

func fieldEditor(c *ui.Context, invalid bool, build func() ui.Element) ui.Element {
	scope := ui.Local(c.Root(), fieldAppearanceKey{}, func() fieldAppearanceScope { return fieldAppearanceScope{} })
	previous := *scope
	scope.invalid = invalid
	defer func() { *scope = previous }()
	return build()
}

func inputReferenceFont(c *ui.Context, size Size) float32 {
	if size == ExtraSmall || size == IconExtraSmall {
		return 12
	}
	if size == Small || size == IconSmall {
		return 14
	}
	if size == Large || size == IconLarge {
		return 16
	}
	width, _ := c.Size()
	if width >= 768 {
		return 14
	}
	return 16
}

func inputFont(c *ui.Context, size Size) float32 {
	return theme.ControlFontSize(c, inputReferenceFont(c, size), 0)
}

func inputLineHeight(size Size) float32 {
	if size == ExtraSmall || size == IconExtraSmall {
		return 4.0 / 3.0
	}
	return 1.5
}

func inputPaddingY(size Size) float32 {
	if size == ExtraSmall || size == IconExtraSmall {
		return 2
	}
	return 4
}

func textInputAppearance(c *ui.Context, e ui.Element, invalid bool) {
	t := theme.TokensOf(c)
	border := t.Input
	if invalid {
		border = t.Destructive
		if t.Dark {
			border = t.Destructive.Alpha(.5)
		}
	}
	e.Border(1, border).FocusRing(false).Shadow(0, 1, 2, 0, ui.RGB(0, 0, 0).Alpha(.05))
	if invalid {
		alpha := float32(.2)
		if t.Dark {
			alpha = .4
		}
		e.Shadow(0, 0, 0, 3, t.Destructive.Alpha(alpha))
	} else if e.FocusVisible() {
		e.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
	}
}

func styleInput(c *ui.Context, e ui.Element, p InputProps) ui.Element {
	t := theme.TokensOf(c)
	bg := ui.Transparent
	if t.Dark {
		bg = t.Input.Alpha(.3)
	}
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	group := ui.Local(c.Root(), inputGroupAppearanceKey{}, func() inputGroupAppearance { return inputGroupAppearance{} })
	for _, invalid := range group.active {
		*invalid = *invalid || p.Invalid
	}
	motion.ApplyTransition(c, e.FillWidth().MinWidth(0).Padding(inputPaddingY(p.Size), 10).Radius(t.RadiusMD()).
		FontSize(theme.ControlFontSize(c, inputReferenceFont(c, p.Size), p.FontSize)).LineHeight(inputLineHeight(p.Size)).TextAlign(ui.Start).Background(bg).TextColor(t.Foreground).
		PlaceholderColor(t.MutedForeground).Placeholder(p.Placeholder).Disabled(p.Disabled).DimDisabled(false).ReadOnly(p.ReadOnly).
		Invalid(p.Invalid).Cursor(ui.CursorText))
	if p.Label != "" {
		e.Label(p.Label)
	}
	if p.Description != "" {
		e.Description(p.Description)
	}
	if p.Disabled {
		e.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	textInputAppearance(c, e, p.Invalid)
	return e
}

func configureTextInput(c *ui.Context, e ui.Element, value *string, p InputProps) ui.Element {
	e.MaxLength(p.MaxLength).OnPaste(p.OnPaste)
	var composition *textCompositionState
	if p.OnSubmit != nil || p.OnTextKey != nil {
		composition = ui.Local(e, textCompositionKey{}, func() textCompositionState { return textCompositionState{} })
		composition.composing = e.Composing()
		e.ObserveInput(func(event ui.InputEvent) {
			switch event.Kind {
			case ui.InputCompose:
				composition.composing = event.Text != ""
			case ui.InputText:
				composition.composing = false
			case ui.InputKeyDown:
				composition.suppressSubmit = composition.composing && event.Key == ui.KeyEnter
			}
		})
	}
	if p.OnTextKey != nil {
		e.OnTextKey(func(key ui.TextKey) bool {
			if composition.composing {
				return false
			}
			return p.OnTextKey(key)
		})
	}
	if p.Handle != nil {
		e.Bind(p.Handle)
	}
	if p.AutoFocus && !e.IsDisabled() {
		e.AutoFocus()
	}
	if p.OnChange != nil {
		e.OnChange(func() { p.OnChange(*value) })
	}
	if p.OnSubmit != nil {
		e.OnSubmit(func() {
			// A raw candidate-choosing Enter can commit native marked text and
			// also raise Submitted. Observe composition before that key so the
			// same gesture never submits the form or sends a message.
			if !composition.composing && !composition.suppressSubmit {
				p.OnSubmit(*value)
			}
		})
	}
	return e
}

func Input(c *ui.Context, value *string, p InputProps) ui.Element {
	e := styleInput(c, ui.TextInputBase(c, value), p)
	height, _ := metrics(p.Size)
	// Preserve exact default heights while leaving room for larger native text.
	height = max(height, theme.ControlFontSize(c, inputReferenceFont(c, p.Size), p.FontSize)*inputLineHeight(p.Size)+2*inputPaddingY(p.Size)+2)
	e.Height(height)
	if p.Password {
		e.Password()
	}
	if p.Search {
		e.Search()
	}
	return configureTextInput(c, e, value, p)
}

type TextareaProps struct {
	Placeholder, Label, Description string
	// Rows requests a fixed native line count. Otherwise the editor grows with
	// wrapped content, bounded by positive MinRows/MaxRows; zero max is unlimited.
	Rows, MinRows, MaxRows      int
	Disabled, Invalid, ReadOnly bool
	AutoFocus                   bool
	MaxLength                   int
	Handle                      *ui.Handle
	OnChange                    func(string)
	OnPaste                     func(string) bool
	OnTextKey                   func(ui.TextKey) bool
}

func Textarea(c *ui.Context, value *string, p TextareaProps) ui.Element {
	lo, hi := max(1, p.MinRows), p.MaxRows
	if hi <= 0 {
		hi = math.MaxInt32
	}
	hi = max(lo, hi)
	if p.Rows > 0 {
		lo, hi = p.Rows, p.Rows
	}
	input := InputProps{Placeholder: p.Placeholder, Label: p.Label, Description: p.Description, Disabled: p.Disabled,
		Invalid: p.Invalid, ReadOnly: p.ReadOnly, AutoFocus: p.AutoFocus, MaxLength: p.MaxLength, Handle: p.Handle,
		OnChange: p.OnChange, OnPaste: p.OnPaste, OnTextKey: p.OnTextKey}
	e := styleInput(c, ui.TextAreaBase(c, value).Lines(lo, hi), input).Padding(8, 10).MinHeight(64)
	return configureTextInput(c, e, value, input)
}
