package conversation

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"strings"
)

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
