// Package conversation provides native conversation composition primitives.
// It covers basic attachment actions, bubble/message layout, keyed virtual
// scrolling, markers, expandable traces and controlled multi-step questions.
// It does not implement upstream DOM APIs, upload transport, drag-and-drop,
// AI SDK adapters, or MessageScroller's turn anchoring/stream pause heuristics.
// Applications own transport, durable state, cancellation and branching.
package conversation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
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
				content()
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

type MessageScrollerProps struct {
	// Key must return a stable unique comparable ID. It preserves the visible
	// message when history is prepended and keeps widget state with messages.
	Key       func(int) any
	Label     string
	FollowEnd bool
}

// MessageScroller uses MyGo's virtual variable-height list. FollowEnd follows
// new content only at the live edge; user scrolling away suspends following.
// Call state.ScrollToEnd() for "Jump to latest", state.ScrollTo(i,ui.Start)
// for search, and state.HoldPlace() when expanding a streaming message.
func MessageScroller(c *ui.Context, state *ui.ListState, count int, p MessageScrollerProps, row func(*ui.Context, int)) ui.Element {
	if state == nil {
		panic("conversation: MessageScroller requires caller-owned ListState")
	}
	if p.Key == nil {
		panic("conversation: MessageScroller requires stable message keys")
	}
	state.Key = p.Key
	state.FollowEnd = p.FollowEnd
	if motion.Reduced(c) {
		state.FollowAnimation = nil
	} else {
		a := ui.Animation{Duration: 150000000, Ease: motion.EaseOut}
		state.FollowAnimation = &a
	}
	label := p.Label
	if label == "" {
		label = "Conversation"
	}
	return ui.List(c, state, max(0, count)).Label(label).Role(ui.RoleLog).Rows(func(r ui.ListRow) {
		ui.Column(r.Context).PaddingY(6).Children(func() {
			if row != nil {
				row(r.Context, r.Index)
			}
		})
	})
}

type MarkerVariant uint8

const (
	Inline MarkerVariant = iota
	Bordered
	Separator
)

type MarkerProps struct {
	Variant MarkerVariant
	Icon    *ui.SVG
	Busy    bool
}

func Marker(c *ui.Context, text string, p MarkerProps) ui.Element {
	t := theme.TokensOf(c)
	e := ui.Row(c).Gap(8).AlignItems(ui.Center).TextColor(t.MutedForeground).FontSize(12).Role(ui.RoleStatus).Label(text).Busy(p.Busy)
	if p.Variant == Bordered {
		e.Padding(10, 12).Border(1, t.Border).Radius(t.RadiusMD())
	}
	return e.Children(func() {
		if p.Variant == Separator {
			ui.Box(c).Grow(1).Height(1).Background(t.Border)
		}
		if p.Icon != nil {
			ui.Icon(c, p.Icon).FontSize(14)
		}
		ui.Text(c, text).SingleLine()
		if p.Variant == Separator {
			ui.Box(c).Grow(1).Height(1).Background(t.Border)
		}
	})
}
