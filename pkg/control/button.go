// Package control implements shadcn's form and action components for MyGo.
// State is owned by the caller; native MyGo controls provide focus, keyboard,
// clipboard, and accessibility behavior.
package control

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type Variant uint8

const (
	Default Variant = iota
	Secondary
	Outline
	Ghost
	Destructive
	Link
)

type Size uint8

const (
	DefaultSize Size = iota
	Small
	Large
	Icon
)

type ButtonProps struct {
	Variant  Variant
	Size     Size
	Disabled bool
	// Icon is optional; icon-only buttons should also set Label.
	Icon  *ui.SVG
	Label string
}

func metrics(size Size) (height, padding float32) {
	switch size {
	case Small:
		return 32, 12
	case Large:
		return 40, 24
	case Icon:
		return 36, 0
	default:
		return 36, 16
	}
}

func Button(c *ui.Context, label string, p ButtonProps) ui.Element {
	return StyleButton(c, ui.ButtonBase(c), label, p)
}

// StyleButton also styles a native menu or dialog trigger without replacing its behavior.
func StyleButton(c *ui.Context, b ui.Element, label string, p ButtonProps) ui.Element {
	t := theme.TokensOf(c)
	h, px := metrics(p.Size)
	b.Height(h).PaddingX(px).Gap(8).Radius(t.Radius * .8).FontSize(14).FontWeight(500).
		AlignItems(ui.Center).Justify(ui.Center).Cursor(ui.CursorPointer).
		Disabled(p.Disabled).Transition(motion.Transition(c))
	bg, fg, border := t.Primary, t.PrimaryForeground, ui.Transparent
	switch p.Variant {
	case Secondary:
		bg, fg = t.Secondary, t.SecondaryForeground
	case Outline:
		bg, fg, border = t.Background, t.Foreground, t.Input
	case Ghost:
		bg, fg = ui.Transparent, t.Foreground
	case Destructive:
		bg, fg = t.Destructive, t.DestructiveForeground
	case Link:
		bg, fg = ui.Transparent, t.Primary
	}
	if !p.Disabled && (b.Hovered() || b.Pressed()) {
		switch p.Variant {
		case Outline, Ghost:
			bg, fg = t.Accent, t.AccentForeground
		case Link:
		default:
			bg = bg.Mix(t.Background, .12)
		}
	}
	b.Background(bg).TextColor(fg).Border(1, border)
	if p.Size == Icon {
		b.Width(h)
	}
	if p.Label != "" {
		b.Label(p.Label)
	}
	if p.Disabled {
		b.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	b.Children(func() {
		if p.Icon != nil {
			ui.Icon(c, p.Icon).FontSize(16)
		}
		if label != "" {
			ui.Text(c, label).SingleLine()
		}
	})
	return b
}

// ButtonGroup joins actions into one keyboard-navigation group.
func ButtonGroup(c *ui.Context, content func()) ui.Element {
	return ui.Row(c).Gap(1).AlignItems(ui.Center).FocusGroup(ui.Horizontal).Children(content)
}

type ToggleProps struct {
	Size              Size
	Outline, Disabled bool
}

func Toggle(c *ui.Context, on *bool, label string, p ToggleProps) ui.Element {
	t := theme.TokensOf(c)
	b := ui.ToggleBase(c, on)
	v := Ghost
	if p.Outline {
		v = Outline
	}
	StyleButton(c, b, label, ButtonProps{Variant: v, Size: p.Size, Disabled: p.Disabled})
	if *on {
		b.Background(t.Accent).TextColor(t.AccentForeground)
	}
	return b
}
