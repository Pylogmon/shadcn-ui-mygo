package conversation

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type Alignment uint8

const (
	Start Alignment = iota
	End
)

type BubbleVariant uint8

const (
	Primary BubbleVariant = iota
	Secondary
	Muted
	Tinted
	Outline
	Destructive
	Ghost
)

type Reaction struct {
	Label              string
	Count              int
	Selected, Disabled bool
	OnClick            func()
}
type BubbleProps struct {
	Variant   BubbleVariant
	Align     Alignment
	Reactions []Reaction
	// Expanded enables a caller-controlled disclosure around content.
	Expanded *bool
	Summary  string
}

func Bubble(c *ui.Context, p BubbleProps, content func()) ui.Element {
	t := theme.TokensOf(c)
	bg, fg, border := t.Primary, t.PrimaryForeground, ui.Transparent
	switch p.Variant {
	case Secondary:
		bg, fg = t.Secondary, t.SecondaryForeground
	case Muted:
		bg, fg = t.Muted, t.Foreground
	case Tinted:
		bg, fg = theme.Mix(t.Background, t.Primary, .1), t.Foreground
	case Outline:
		bg, fg, border = t.Background, t.Foreground, t.Border
	case Destructive:
		bg, fg = t.Destructive, t.DestructiveForeground
	case Ghost:
		bg, fg = ui.Transparent, t.Foreground
	}
	e := ui.Column(c).Gap(8).Padding(12, 16).Radius(t.Radius).Background(bg).TextColor(fg).Border(1, border).MaxWidthPercent(80)
	if p.Align == End {
		e.AlignSelf(ui.End)
	} else {
		e.AlignSelf(ui.Start)
	}
	if p.Variant == Ghost {
		e.MaxWidthPercent(100).Padding(0)
	}
	return e.Children(func() {
		if p.Expanded == nil || *p.Expanded {
			if content != nil {
				content()
			}
		} else {
			ui.Text(c, p.Summary)
		}
		if p.Expanded != nil {
			label := "Show more"
			if *p.Expanded {
				label = "Show less"
			}
			if control.Button(c, label, control.ButtonProps{Variant: control.Ghost, Size: control.Small}).Expanded(*p.Expanded).Clicked() {
				*p.Expanded = !*p.Expanded
			}
		}
		if len(p.Reactions) > 0 {
			ui.Row(c).Gap(4).Children(func() {
				for i, r := range p.Reactions {
					label := r.Label
					if r.Count > 0 {
						label += fmt.Sprintf(" %d", r.Count)
					}
					variant := control.Outline
					if r.Selected {
						variant = control.Secondary
					}
					if control.Button(c.Key(i), label, control.ButtonProps{Variant: variant, Size: control.Small, Disabled: r.Disabled}).Clicked() && r.OnClick != nil {
						r.OnClick()
					}
				}
			})
		}
	})
}
