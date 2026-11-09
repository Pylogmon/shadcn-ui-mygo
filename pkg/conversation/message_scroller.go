package conversation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

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
