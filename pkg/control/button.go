// Package control implements shadcn's form and action components for MyGo.
// State is owned by the caller; native MyGo controls provide focus, keyboard,
// clipboard, and accessibility behavior.
package control

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
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
	ExtraSmall
	IconExtraSmall
	IconSmall
	IconLarge
)

type ButtonProps struct {
	Variant Variant
	Size    Size
	// FontSize overrides the theme/Size font in unscaled DIPs. Zero inherits.
	FontSize          float32
	Disabled, Invalid bool
	// Loading exposes busy state, replaces the leading icon with a spinner and
	// suppresses activation without changing the accessible label.
	Loading bool
	// Icon is optional; icon-only buttons should also set Label.
	Icon  *ui.SVG
	Label string
}

func metrics(size Size) (height, padding float32) {
	switch size {
	case Small:
		return 32, 10
	case Large:
		return 40, 10
	case ExtraSmall:
		return 24, 8
	case IconExtraSmall:
		return 24, 0
	case IconSmall:
		return 32, 0
	case IconLarge:
		return 40, 0
	case Icon:
		return 36, 0
	default:
		return 36, 10
	}
}

func Button(c *ui.Context, label string, p ButtonProps) ui.Element {
	return StyleButton(c, ui.ButtonBase(c), label, p)
}

// StyleButton also styles a native menu or dialog trigger without replacing its behavior.
func StyleButton(c *ui.Context, b ui.Element, label string, p ButtonProps) ui.Element {
	t := theme.TokensOf(c)
	h, px := metrics(p.Size)
	reference, lineHeight := float32(14), float32(20.0/14)
	if p.Size == ExtraSmall || p.Size == IconExtraSmall {
		reference, lineHeight = 12, 4.0/3
	}
	font := theme.ControlFontSize(c, reference, p.FontSize)
	h = max(h, font*lineHeight+8)
	motion.ApplyTransition(c, b.Height(h).PaddingX(px).Gap(6).Radius(t.RadiusMD()).FontSize(font).LineHeight(lineHeight).FontWeight(500).
		AlignItems(ui.Center).Justify(ui.Center).Cursor(ui.CursorPointer).
		Disabled(p.Disabled || p.Loading).Busy(p.Loading).Invalid(p.Invalid).FocusRing(false))
	if p.Size == ExtraSmall || p.Size == IconExtraSmall {
		b.Gap(4).Radius(min(t.RadiusMD(), 8))
	} else if p.Size == Small || p.Size == IconSmall {
		b.Radius(min(t.RadiusMD(), 10))
		if p.Size == Small {
			b.Gap(4)
		}
	}
	iconOnly := p.Size == Icon || p.Size == IconExtraSmall || p.Size == IconSmall || p.Size == IconLarge
	if !iconOnly && (p.Icon != nil || p.Loading) {
		inset := float32(8)
		if p.Size == ExtraSmall || p.Size == Small {
			inset = 6
		}
		b.PaddingStart(inset)
	}
	bg, fg, border := t.Primary, t.PrimaryForeground, ui.Transparent
	switch p.Variant {
	case Secondary:
		bg, fg = t.Secondary, t.SecondaryForeground
	case Outline:
		bg, fg, border = t.Background, t.Foreground, t.Border
		if t.Dark {
			bg, border = t.Input.Alpha(.3), t.Input
		}
	case Ghost:
		bg, fg = ui.Transparent, t.Foreground
	case Destructive:
		bg, fg = t.Destructive.Alpha(.1), t.Destructive
		if t.Dark {
			bg = t.Destructive.Alpha(.2)
		}
	case Link:
		bg, fg = ui.Transparent, t.Primary
	}
	if !p.Disabled && !p.Loading && (b.Hovered() || b.Pressed()) {
		switch p.Variant {
		case Outline, Ghost:
			bg, fg = t.Muted, t.Foreground
			if t.Dark {
				if p.Variant == Outline {
					bg = t.Input.Alpha(.5)
				} else {
					bg = t.Muted.Alpha(.5)
				}
			}
		case Destructive:
			bg = t.Destructive.Alpha(.2)
			if t.Dark {
				bg = t.Destructive.Alpha(.3)
			}
		case Link:
		case Secondary:
			bg = theme.Mix(t.Secondary, t.Foreground, .05)
		default:
			bg = bg.Alpha(.8)
		}
	}
	b.Background(bg).TextColor(fg).Border(1, border)
	if p.Invalid {
		b.Border(1, t.Destructive)
		b.Shadow(0, 0, 0, 3, t.Destructive.Alpha(.2))
	}
	if b.FocusVisible() {
		ring := t.Ring.Alpha(.5)
		border = t.Ring
		if p.Variant == Destructive || p.Invalid {
			ring = t.Destructive.Alpha(.2)
			border = t.Destructive.Alpha(.4)
			if t.Dark {
				ring = t.Destructive.Alpha(.4)
			}
		}
		b.Border(1, border).Shadow(0, 0, 0, 3, ring)
	}
	if iconOnly {
		b.Width(h)
	}
	if p.Label != "" {
		b.Label(p.Label)
	}
	if p.Disabled {
		b.DimDisabled(false).Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	if p.Loading {
		b.Cursor(ui.CursorWait)
	}
	b.Children(func() {
		if p.Loading {
			indicator.Spinner(c, indicator.Props{Decorative: true})
		} else if p.Icon != nil {
			size := float32(16)
			if p.Size == ExtraSmall || p.Size == IconExtraSmall {
				size = 12
			}
			ui.Icon(c, p.Icon).Size(size, size)
		}
		if label != "" {
			text := ui.Text(c, label).SingleLine()
			if p.Variant == Link && b.Hovered() && !p.Disabled && !p.Loading {
				text.Underline().UnderlineOffset(4)
			}
		}
	})
	return ButtonGroupItem(c, b)
}
