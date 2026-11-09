package conversation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type MessageProps struct {
	Sender                 string
	Align                  Alignment
	Avatar, Header, Footer func()
}

// Message owns the row and metadata; compose Bubble inside content for a face.
func Message(c *ui.Context, p MessageProps, content func()) ui.Element {
	e := ui.Row(c).Gap(10).AlignItems(ui.End).Role(ui.RoleArticle).Label(p.Sender)
	if p.Align == End {
		e.Justify(ui.End)
	}
	return e.Children(func() {
		avatar := func() {
			if p.Avatar != nil {
				ui.Column(c.Key("avatar")).Shrink(0).Children(p.Avatar)
			}
		}
		if p.Align == Start {
			avatar()
		}
		body := ui.Column(c.Key("body")).Gap(6).MaxWidthPercent(90).MinWidth(0)
		if p.Align == End {
			body.AlignItems(ui.End)
		}
		body.Children(func() {
			if p.Header != nil {
				p.Header()
			} else if p.Sender != "" {
				ui.Text(c, p.Sender).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			}
			if content != nil {
				withBubbleMessageAlignment(c, p.Align, content)
			}
			if p.Footer != nil {
				p.Footer()
			}
		})
		if p.Align == End {
			avatar()
		}
	})
}
