// Package display provides shadcn's surfaces, typography and status components.
package display

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func Card(c *ui.Context, content func()) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).FillWidth().Gap(24).Padding(24).Radius(t.Radius*1.4).
		Background(t.Card).TextColor(t.CardForeground).Border(1, t.Border).Children(func() { theme.WithSurface(c, t.Card, content) })
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
