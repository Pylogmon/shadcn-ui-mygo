package control

import (
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type InputProps struct {
	Placeholder, Label                    string
	Size                                  Size
	Disabled, Invalid, ReadOnly, Password bool
}

func styleInput(c *ui.Context, e ui.Element, p InputProps) ui.Element {
	t := theme.TokensOf(c)
	h, px := metrics(p.Size)
	border := t.Input
	if p.Invalid {
		border = t.Destructive
	}
	e.MinHeight(h).PaddingX(min(px, 12)).PaddingY(7).Radius(t.Radius*.8).
		FontSize(14).Background(t.Background).TextColor(t.Foreground).
		Border(1, border).PlaceholderColor(t.MutedForeground).
		Placeholder(p.Placeholder).Disabled(p.Disabled).ReadOnly(p.ReadOnly).
		Invalid(p.Invalid).Cursor(ui.CursorText).Transition(motion.Transition(c)).
		FocusRingColor(t.Ring.Alpha(.5)).FocusRingOffset(0)
	if p.Label != "" {
		e.Label(p.Label)
	}
	if p.Disabled {
		e.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	if p.Invalid {
		e.FocusRingColor(t.Destructive.Alpha(.4))
	}
	return e
}

func Input(c *ui.Context, value *string, p InputProps) ui.Element {
	e := styleInput(c, ui.TextInputBase(c, value), p)
	if p.Password {
		e.Password()
	}
	return e
}

type TextareaProps struct {
	Placeholder, Label          string
	MinRows, MaxRows            int
	Disabled, Invalid, ReadOnly bool
}

func Textarea(c *ui.Context, value *string, p TextareaProps) ui.Element {
	lo, hi := p.MinRows, p.MaxRows
	if lo <= 0 {
		lo = 3
	}
	if hi <= 0 {
		hi = 8
	}
	hi = max(lo, hi)
	return styleInput(c, ui.TextAreaBase(c, value).Lines(lo, hi), InputProps{
		Placeholder: p.Placeholder, Label: p.Label, Disabled: p.Disabled, Invalid: p.Invalid, ReadOnly: p.ReadOnly,
	})
}

func Label(c *ui.Context, text string) ui.Element {
	return ui.Text(c, text).FontSize(14).FontWeight(500).TextColor(theme.TokensOf(c).Foreground)
}

type FieldProps struct {
	Label, Description, Error string
	Required, Disabled        bool
}

// Field links its label and helper text to the returned editor for assistive technology.
func Field(c *ui.Context, p FieldProps, editor func() ui.Element) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).Gap(7).Disabled(p.Disabled).Children(func() {
		label := p.Label
		if p.Required {
			label += " *"
		}
		var title ui.Element
		if label != "" {
			title = Label(c, label)
		}
		input := editor()
		if label != "" {
			input.LabelledBy(title)
		}
		if p.Error != "" {
			desc := ui.Text(c, p.Error).FontSize(13).TextColor(t.Destructive).Role(ui.RoleAlert)
			input.Invalid(true).DescribedBy(desc)
		} else if p.Description != "" {
			desc := ui.Text(c, p.Description).FontSize(13).TextColor(t.MutedForeground)
			input.DescribedBy(desc)
		}
	})
}

type InputGroupProps struct {
	Prefix, Suffix    string
	Disabled, Invalid bool
}

// InputGroup surrounds an editor with units, shortcuts, or supporting text.
func InputGroup(c *ui.Context, p InputGroupProps, editor func() ui.Element) ui.Element {
	t := theme.TokensOf(c)
	border := t.Input
	if p.Invalid {
		border = t.Destructive
	}
	return ui.Row(c).AlignItems(ui.Center).Gap(8).PaddingX(12).Radius(t.Radius*.8).
		Border(1, border).Background(t.Background).Disabled(p.Disabled).Children(func() {
		if p.Prefix != "" {
			ui.Text(c, p.Prefix).FontSize(14).TextColor(t.MutedForeground)
		}
		e := editor().Grow(1).MinWidth(0).Border(0, ui.Transparent).Background(ui.Transparent)
		if p.Invalid {
			e.Invalid(true)
		}
		if p.Suffix != "" {
			ui.Text(c, p.Suffix).FontSize(14).TextColor(t.MutedForeground)
		}
	})
}

type InputOTPProps struct {
	Length                          int
	Label                           string
	Alphanumeric, Disabled, Invalid bool
	OnComplete                      func(string)
}

// InputOTP uses a single real editor, so paste, selection, and Backspace retain native behavior.
func InputOTP(c *ui.Context, value *string, p InputOTPProps) ui.Element {
	n := p.Length
	if n <= 0 {
		n = 6
	}
	clean := make([]rune, 0, n)
	for _, ch := range *value {
		if len(clean) < n && (ch >= '0' && ch <= '9' || p.Alphanumeric && unicode.IsLetter(ch)) {
			clean = append(clean, ch)
		}
	}
	*value = string(clean)
	t := theme.TokensOf(c)
	root := ui.Row(c).Gap(0).AlignSelf(ui.Start).Disabled(p.Disabled)
	previous := ui.Local(root, "completed", func() string { return *value })
	if *previous != *value {
		*previous = *value
		if len(clean) == n && p.OnComplete != nil {
			p.OnComplete(*value)
		}
	}
	var input ui.Element
	root.Children(func() {
		input = ui.TextInputBase(c, value).Absolute().Left(0).Right(0).Top(0).Bottom(0).
			Opacity(0).FocusRing(false).Disabled(p.Disabled).Invalid(p.Invalid)
		label := p.Label
		if label == "" {
			label = "One-time code"
		}
		input.Label(label)
		for i := range n {
			border := t.Input
			if p.Invalid {
				border = t.Destructive
			}
			cell := ui.Box(c.Key(i)).Size(40, 40).Border(1, border).Center().IgnoreInput(true)
			if i == 0 || i == n-1 {
				cell.Radius(t.Radius * .8)
			}
			if input.Focused() && (i == len(clean) || i == n-1 && len(clean) == n) {
				cell.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.3))
			}
			cell.Children(func() {
				if i < len(clean) {
					ui.Text(c, string(clean[i])).FontSize(14).TextColor(t.Foreground)
				}
			})
		}
	})
	return input
}

// Validator returns an empty string for a valid value, or a user-facing error.
type Validator func(string) string

func Required(message string) Validator {
	return func(s string) string {
		if strings.TrimSpace(s) == "" {
			return message
		}
		return ""
	}
}

func Validate(value string, validators ...Validator) string {
	for _, check := range validators {
		if err := check(value); err != "" {
			return err
		}
	}
	return ""
}
