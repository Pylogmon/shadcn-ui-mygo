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
	Orientation               FieldOrientation
	// LabelWidth reserves label room in horizontal layouts; zero uses 140 DIPs.
	LabelWidth float32
	// ResponsiveWidth is the enclosing FieldGroup's width for responsive fields.
	// Without a FieldGroup or this override the window width is used.
	ResponsiveWidth float32
}

type FieldOrientation uint8

const (
	FieldVertical FieldOrientation = iota
	FieldHorizontal
	FieldResponsive
)

type fieldGroupLayout struct {
	width  float32
	active bool
}

// FieldGroup stacks fields and supplies its measured width for responsive
// fields. A changed layout requests one correction frame; stable groups do not
// schedule an idle render loop.
func FieldGroup(c *ui.Context, content func()) ui.Element {
	root := ui.Column(c).FillWidth().Gap(24)
	s := ui.Local(c.Root(), "shadcn.control.field-group", func() fieldGroupLayout { return fieldGroupLayout{} })
	previous := *s
	measured := ui.Local(root, "responsive-width", func() float32 { return 0 })
	root.Draw(func(_ *ui.Painter, r ui.Rect) {
		difference := *measured - r.W
		if difference > .25 || difference < -.25 {
			*measured = r.W
			c.Invalidate()
		}
	})
	width := *measured
	if width <= 0 {
		width, _ = c.Size()
	}
	s.width, s.active = width, true
	defer func() { *s = previous }()
	return root.Children(content)
}

type FieldSetProps struct {
	Legend, Description string
	Disabled            bool
}

func FieldSet(c *ui.Context, p FieldSetProps, content func()) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Column(c).FillWidth().Gap(16).Role(ui.RoleGroup).Label(p.Legend).Disabled(p.Disabled)
	return root.Children(func() {
		if p.Legend != "" {
			ui.Text(c, p.Legend).FontSize(16).FontWeight(500)
		}
		if p.Description != "" {
			d := ui.Text(c, p.Description).FontSize(14).TextColor(t.MutedForeground)
			root.DescribedBy(d)
		}
		if content != nil {
			content()
		}
	})
}

// LabelFor associates an independently built label with its editor. Native
// label activation focuses text fields and toggles checkboxes/radios/switches.
func LabelFor(c *ui.Context, text string, editor ui.Element) ui.Element {
	label := Label(c, text)
	fieldTarget(c, editor).LabelledBy(label)
	return label
}

func fieldTarget(c *ui.Context, root ui.Element) ui.Element {
	handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	if target := c.Resolve(*handle); target.Valid() {
		return target
	}
	return root
}

// Field links its label and helper text to the returned editor for assistive technology.
func Field(c *ui.Context, p FieldProps, editor func() ui.Element) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Column(c).Gap(7).FillWidth().Disabled(p.Disabled).Role(ui.RoleGroup)
	horizontal := p.Orientation == FieldHorizontal
	if p.Orientation == FieldResponsive {
		width := p.ResponsiveWidth
		s := ui.Local(c.Root(), "shadcn.control.field-group", func() fieldGroupLayout { return fieldGroupLayout{} })
		if width <= 0 && s.active {
			width = s.width
		}
		if width <= 0 {
			width, _ = c.Size()
		}
		horizontal = width >= 448
	}
	if horizontal {
		root.Row().AlignItems(ui.Start).Gap(16)
	}
	return root.Children(func() {
		label := p.Label
		if p.Required {
			label += " *"
		}
		var title ui.Element
		if label != "" {
			title = Label(c, label)
			if horizontal {
				width := p.LabelWidth
				if width <= 0 {
					width = 140
				}
				title.Width(width).Shrink(0).PaddingY(8)
			}
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(7).Children(func() {
			input := fieldTarget(c, editor())
			if label != "" {
				input.LabelledBy(title)
			}
			if p.Required {
				input.Description("Required")
			}
			if p.Description != "" {
				desc := ui.Text(c, p.Description).FontSize(14).TextColor(t.MutedForeground)
				input.DescribedBy(desc)
			}
			if p.Error != "" {
				desc := ui.Text(c, p.Error).FontSize(14).TextColor(t.Destructive).Role(ui.RoleAlert)
				input.Invalid(true).ErrorMessage(desc).DescribedBy(desc)
			}
		})
	})
}

type InputGroupProps struct {
	Prefix, Suffix                               string
	Disabled, Invalid                            bool
	InlineStart, InlineEnd, BlockStart, BlockEnd func()
	Label                                        string
}

// InputGroup surrounds an editor with units, shortcuts, or supporting text.
func InputGroup(c *ui.Context, p InputGroupProps, editor func() ui.Element) ui.Element {
	t := theme.TokensOf(c)
	border := t.Input
	if p.Invalid {
		border = t.Destructive
	}
	root := ui.Column(c).FillWidth().Radius(t.RadiusMD()).Border(1, border).Background(t.Background).Role(ui.RoleGroup).Disabled(p.Disabled)
	// Treat the compound control as one joined group item, rather than
	// accidentally joining the action buttons inside its addons.
	scope := groupScope(c)
	var parent *groupBuild
	before := 0
	if len(scope.active) > 0 {
		parent = scope.active[len(scope.active)-1]
		before = len(parent.items)
	}
	if t.Dark {
		root.Background(t.Input.Alpha(.3))
	}
	var input ui.Element
	handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	// Buttons inside an addon retain native activation; unhandled addon pointer
	// presses focus the editor, matching input-group icon/text click behavior.
	addon := func(key string, block bool, content func()) {
		if content == nil {
			return
		}
		a := ui.Row(c.Key(key)).Gap(8).AlignItems(ui.Center).TextColor(t.MutedForeground).FontSize(14)
		if block {
			a.FillWidth().Padding(8, 10)
		} else {
			a.Padding(4, 8).Shrink(0)
		}
		a.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 && !p.Disabled {
				handle.Focus()
			}
			return false
		}).Children(content)
	}
	root.Children(func() {
		addon("block-start", true, p.BlockStart)
		ui.Row(c.Key("control")).FillWidth().Gap(0).AlignItems(ui.Center).Children(func() {
			if p.Prefix != "" || p.InlineStart != nil {
				addon("inline-start", false, func() {
					if p.Prefix != "" {
						ui.Text(c, p.Prefix)
					}
					if p.InlineStart != nil {
						p.InlineStart()
					}
				})
			}
			input = editor().Bind(handle).Grow(1).MinWidth(0).Radius(0).Border(0, ui.Transparent).Background(ui.Transparent).FocusRing(false).Disabled(p.Disabled).Invalid(p.Invalid)
			if p.Label != "" {
				input.Label(p.Label)
			}
			if p.Suffix != "" || p.InlineEnd != nil {
				addon("inline-end", false, func() {
					if p.Suffix != "" {
						ui.Text(c, p.Suffix)
					}
					if p.InlineEnd != nil {
						p.InlineEnd()
					}
				})
			}
		})
		addon("block-end", true, p.BlockEnd)
	})
	if input.FocusVisible() {
		root.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
	}
	if p.Invalid {
		alpha := float32(.2)
		if t.Dark {
			alpha = .4
		}
		root.Border(1, t.Destructive).Shadow(0, 0, 0, 3, t.Destructive.Alpha(alpha))
	}
	if p.Disabled {
		root.Opacity(.5)
	}

	if parent != nil {
		parent.items = parent.items[:before]
	}
	return ButtonGroupItem(c, root)
}

func InputGroupButton(c *ui.Context, label string, props ...ButtonProps) ui.Element {
	p := ButtonProps{Variant: Ghost, Size: ExtraSmall}
	if len(props) > 0 {
		p = props[0]
	}
	if p.Size == DefaultSize {
		p.Size = ExtraSmall
	}
	return Button(c, label, p).Radius(max(0, theme.TokensOf(c).Radius-5)).PaddingX(6)
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
