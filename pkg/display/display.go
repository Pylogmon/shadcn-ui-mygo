// Package display provides shadcn's surfaces, typography and status components.
package display

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func Card(c *ui.Context, content func()) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).FillWidth().Gap(24).Padding(24).Radius(t.Radius*1.4).
		Background(t.Card).TextColor(t.CardForeground).Border(1, t.Border).Children(content)
}
func CardHeader(c *ui.Context, title, description string) ui.Element {
	return ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, title).FontSize(18).FontWeight(600)
		if description != "" {
			ui.Text(c, description).FontSize(14).TextColor(theme.TokensOf(c).MutedForeground)
		}
	})
}
func CardContent(c *ui.Context, content func()) ui.Element {
	return ui.Column(c).FillWidth().Gap(12).Children(content)
}
func CardFooter(c *ui.Context, content func()) ui.Element {
	return ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(content)
}

type BadgeVariant int

const (
	BadgeDefault BadgeVariant = iota
	BadgeSecondary
	BadgeOutline
	BadgeDestructive
)

func Badge(c *ui.Context, text string, variant BadgeVariant) ui.Element {
	t := theme.TokensOf(c)
	bg, fg, border := t.Primary, t.PrimaryForeground, ui.Transparent
	switch variant {
	case BadgeSecondary:
		bg, fg = t.Secondary, t.SecondaryForeground
	case BadgeOutline:
		bg, fg, border = ui.Transparent, t.Foreground, t.Border
	case BadgeDestructive:
		bg, fg = t.Destructive, t.DestructiveForeground
	}
	return ui.Row(c).Padding(2, 8).Radius(999).Border(1, border).Background(bg).TextColor(fg).FontSize(12).FontWeight(500).Children(func() { ui.Text(c, text) })
}

type AlertProps struct {
	Title, Description string
	Destructive        bool
	Content            func()
}

func Alert(c *ui.Context, p AlertProps) ui.Element {
	t := theme.TokensOf(c)
	fg := t.Foreground
	if p.Destructive {
		fg = t.Destructive
	}
	return ui.Column(c).FillWidth().Gap(6).Padding(16).Radius(t.Radius).Border(1, t.Border).Background(t.Card).TextColor(fg).Role(ui.RoleAlert).Children(func() {
		if p.Title != "" {
			ui.Text(c, p.Title).FontWeight(500)
		}
		if p.Description != "" {
			ui.Text(c, p.Description).FontSize(14).LineHeight(1.5)
		}
		if p.Content != nil {
			p.Content()
		}
	})
}

type SeparatorProps struct {
	// Semantic exposes a named/oriented separator to assistive technology.
	// The default is decorative, as in shadcn's separator primitive.
	Semantic bool
	Label    string
}

func Separator(c *ui.Context, vertical bool, props ...SeparatorProps) ui.Element {
	e := ui.Box(c).Background(theme.TokensOf(c).Border).Shrink(0)
	p := SeparatorProps{}
	if len(props) > 0 {
		p = props[0]
	}
	if p.Semantic {
		e.Role(ui.RoleSeparator).Label(p.Label)
		if vertical {
			e.Orientation(ui.Vertical)
		} else {
			e.Orientation(ui.Horizontal)
		}
	} else {
		e.AccessHidden(true)
	}
	if vertical {
		return e.Width(1).AlignSelf(ui.Stretch)
	}
	return e.Height(1).FillWidth()
}
func AspectRatio(c *ui.Context, ratio float32, content func()) ui.Element {
	if ratio <= 0 {
		ratio = 1
	}
	return ui.Box(c).FillWidth().AspectRatio(ratio).Clip().Children(content)
}
func Kbd(c *ui.Context, key string) ui.Element {
	t := theme.TokensOf(c)
	return ui.Row(c).Padding(2, 5).Radius(4).Background(t.Muted).Border(1, t.Border).TextColor(t.MutedForeground).FontSize(12).Children(func() { ui.Text(c, key) })
}

type TextStyle int

const (
	Paragraph TextStyle = iota
	H1
	H2
	H3
	H4
	Lead
	Large
	Small
	Muted
	Blockquote
	InlineCode
)

func Typography(c *ui.Context, text string, style TextStyle) ui.Element {
	t := theme.TokensOf(c)
	e := ui.Text(c, text).TextColor(t.Foreground).FontSize(16).LineHeight(1.6)
	switch style {
	case H1:
		e.FontSize(36).FontWeight(800).LineHeight(1.15).Role(ui.RoleHeading).Level(1)
	case H2:
		e.FontSize(30).FontWeight(600).LineHeight(1.2).Role(ui.RoleHeading).Level(2)
	case H3:
		e.FontSize(24).FontWeight(600).LineHeight(1.25).Role(ui.RoleHeading).Level(3)
	case H4:
		e.FontSize(20).FontWeight(600).Role(ui.RoleHeading).Level(4)
	case Lead:
		e.FontSize(20).TextColor(t.MutedForeground)
	case Large:
		e.FontSize(18).FontWeight(600)
	case Small:
		e.FontSize(14).FontWeight(500)
	case Muted:
		e.FontSize(14).TextColor(t.MutedForeground)
	case Blockquote:
		e.Padding(0, 0, 0, 24).InsetShadow(2, 0, 0, 0, t.Border)
	case InlineCode:
		e.FontSize(14).Background(t.Muted).Radius(4).Padding(2, 4)
	}
	return e
}

type EmptyProps struct {
	Title, Description string
	Action             func()
}

func Empty(c *ui.Context, p EmptyProps) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(8).Padding(32).Children(func() {
		ui.Text(c, p.Title).FontSize(18).FontWeight(600)
		ui.Text(c, p.Description).FontSize(14).TextColor(t.MutedForeground).TextAlign(ui.Center)
		if p.Action != nil {
			ui.Row(c).Margin(8, 0, 0, 0).Children(p.Action)
		}
	})
}

type ItemProps struct {
	Title, Description string
	Leading, Trailing  func()
}

func Item(c *ui.Context, p ItemProps) ui.Element {
	return ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Center).Padding(12).Children(func() {
		if p.Leading != nil {
			p.Leading()
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			Typography(c, p.Title, Small)
			if p.Description != "" {
				Typography(c, p.Description, Muted)
			}
		})
		if p.Trailing != nil {
			p.Trailing()
		}
	})
}

type ProgressProps struct {
	Label, ValueText       string
	Height                 float32
	Indeterminate, Reverse bool
}

// Progress accepts fractions in 0..1; a negative value or Indeterminate shows
// activity of unknown length without pretending zero percent is known.
func Progress(c *ui.Context, value float64, props ...ProgressProps) ui.Element {
	t := theme.TokensOf(c)
	p := ProgressProps{}
	if len(props) > 0 {
		p = props[0]
	}
	if p.Indeterminate || math.IsNaN(value) || value < 0 {
		value = -1
	} else {
		value = min(1, value)
	}
	height := p.Height
	if height <= 0 || math.IsNaN(float64(height)) || math.IsInf(float64(height), 0) {
		height = 6
	}
	label := p.Label
	if label == "" {
		label = "Progress"
	}
	text := p.ValueText
	if text == "" {
		if value < 0 {
			text = "Loading"
		} else {
			text = fmt.Sprintf("%.0f%%", value*100)
		}
	}
	e := ui.Progress(c, value).Height(height).Radius(999).Background(t.Muted).TextColor(t.Primary).Label(label).Value(text)
	if p.Reverse {
		e.Reverse()
	}
	if value < 0 && motion.Reduced(c) {
		// A stable segment retains the unknown-progress cue without time-based
		// movement on unrelated UI refreshes or scheduling any animation frames.
		e.Draw(func(painter *ui.Painter, r ui.Rect) {
			painter.Fill(ui.Rect{X: r.X + r.W*.35, Y: r.Y, W: r.W * .3, H: r.H}, t.Primary, r.H/2)
		})
	}
	return e
}
func Spinner(c *ui.Context) ui.Element {
	return ui.Spinner(c).Size(16, 16).TextColor(theme.TokensOf(c).Primary).Label("Loading")
}

type SkeletonProps struct {
	Width, Height  float32
	Circle, Static bool
}

func Skeleton(c *ui.Context, p SkeletonProps) ui.Element {
	t := theme.TokensOf(c)
	h := p.Height
	if h <= 0 {
		h = 16
	}
	e := ui.Box(c).Height(h).Radius(t.Radius * .8).Background(t.Muted)
	if p.Width > 0 {
		e.Width(p.Width)
	} else {
		e.FillWidth()
	}
	if p.Circle {
		e.Width(h).Radius(h / 2)
	}
	if !p.Static && !c.Preferences().ReduceMotion {
		e.Opacity(1 - .5*e.Loop("pulse", 2*time.Second, ui.Bounce(ui.EaseInOut)))
	}
	return e
}

// Initials uses Unicode words and runes so names need not be Latin text.
func Initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return ""
	}
	first, _ := utf8.DecodeRuneInString(parts[0])
	result := string(first)
	if len(parts) > 1 {
		last, _ := utf8.DecodeRuneInString(parts[len(parts)-1])
		result += string(last)
	}
	return strings.ToUpper(result)
}
