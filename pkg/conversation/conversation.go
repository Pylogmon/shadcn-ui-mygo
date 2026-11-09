// Package conversation provides native conversation composition primitives.
// It covers basic attachment actions, bubble/message layout, keyed virtual
// scrolling, markers, expandable traces and controlled multi-step questions.
// It does not implement upstream DOM APIs, upload transport, drag-and-drop,
// AI SDK adapters, or MessageScroller's turn anchoring/stream pause heuristics.
// Applications own transport, durable state, cancellation and branching.
package conversation

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type AttachmentProps struct {
	Name, MediaType string
	SizeBytes       int64
	// State is ready (default), uploading, or error. Progress is 0..1.
	State            string
	Progress         float64
	Disabled         bool
	Icon             *ui.SVG
	Preview          func()
	OnOpen, OnRemove func()
}

func FileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", max(0, bytes))
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}

func Attachment(c *ui.Context, p AttachmentProps) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).Gap(8).Padding(12).Radius(t.RadiusMD()).Background(t.Card).Border(1, t.Border).
		Label(p.Name).Role(ui.RoleGroup).Children(func() {
		if p.Preview != nil {
			p.Preview()
		}
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			if p.Icon != nil {
				ui.Icon(c, p.Icon).FontSize(24).TextColor(t.MutedForeground)
			}
			ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
				ui.Text(c, p.Name).FontWeight(500).SingleLine()
				meta := FileSize(p.SizeBytes)
				if p.MediaType != "" {
					meta = p.MediaType + " · " + meta
				}
				if p.State == "uploading" {
					meta = fmt.Sprintf("Uploading · %.0f%%", max(0, min(1, p.Progress))*100)
				}
				if p.State == "error" {
					meta = "Upload failed"
				}
				fg := t.MutedForeground
				if p.State == "error" {
					fg = t.Destructive
				}
				ui.Text(c, meta).FontSize(12).TextColor(fg).Role(ui.RoleStatus)
			})
			if p.OnOpen != nil && control.Button(c, "Open", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Disabled: p.Disabled || p.State == "uploading"}).Clicked() {
				p.OnOpen()
			}
			if p.OnRemove != nil && control.Button(c, "Remove", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Disabled: p.Disabled, Label: "Remove " + p.Name}).Clicked() {
				p.OnRemove()
			}
		})
	})
}

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

type TraceStatus string

const (
	Pending  TraceStatus = "pending"
	Running  TraceStatus = "running"
	Complete TraceStatus = "complete"
	Error    TraceStatus = "error"
)

type TraceStep struct {
	ID, Title, Description, Detail string
	Status                         TraceStatus
	Icon                           *ui.SVG
	Children                       []TraceStep
	OnAction                       func()
}
type TraceState struct{ Expanded map[string]bool }

func Trace(c *ui.Context, state *TraceState, steps []TraceStep) ui.Element {
	if state == nil {
		panic("conversation: Trace requires caller-owned state")
	}
	if state.Expanded == nil {
		state.Expanded = make(map[string]bool)
	}
	t := theme.TokensOf(c)
	return ui.Column(c).Gap(6).Role(ui.RoleGroup).Label("Trace").Children(func() {
		for i, step := range steps {
			id := step.ID
			if id == "" {
				id = fmt.Sprintf("step-%d", i)
			}
			ctx := c.Key(id)
			ui.Column(ctx).Gap(6).Children(func() {
				fg := t.MutedForeground
				if step.Status == Error {
					fg = t.Destructive
				} else if step.Status == Complete {
					fg = t.Foreground
				}
				expandable := step.Detail != "" || len(step.Children) > 0
				line := ui.ButtonBase(ctx).Row().Gap(8).Padding(6).Radius(t.RadiusSM()).TextColor(fg).Label(step.Title).Busy(step.Status == Running).Cursor(ui.CursorPointer)
				if expandable {
					line.Expanded(state.Expanded[id])
				}
				line.Children(func() {
					if step.Icon != nil {
						ui.Icon(ctx, step.Icon).FontSize(14)
					}
					ui.Text(ctx, step.Title).FontWeight(500)
					if step.Description != "" {
						ui.Text(ctx, step.Description).FontSize(12).TextColor(t.MutedForeground)
					}
					if expandable {
						indicator := "›"
						if state.Expanded[id] {
							indicator = "⌄"
						}
						ui.Text(ctx, indicator)
					}
				})
				if line.Clicked() {
					if expandable {
						state.Expanded[id] = !state.Expanded[id]
					}
					if step.OnAction != nil {
						step.OnAction()
					}
				}
				if state.Expanded[id] && expandable {
					ui.Column(ctx.Key("detail")).Padding(8, 12).Margin(0, 0, 0, 12).Gap(6).Border(1, t.Border).Radius(t.RadiusSM()).Transition(motion.Collapse(ctx)).Children(func() {
						if strings.TrimSpace(step.Detail) != "" {
							ui.Text(ctx, step.Detail).Selectable().Font("monospace").FontSize(12)
						}
						if len(step.Children) > 0 {
							Trace(ctx, state, step.Children)
						}
					})
				}
			})
		}
	})
}
