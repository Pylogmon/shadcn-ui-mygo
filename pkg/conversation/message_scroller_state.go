package conversation

import (
	"time"

	"github.com/egoist/mygo/ui"
)

type MessageScrollerPosition string

const (
	MessageScrollerStart      MessageScrollerPosition = "start"
	MessageScrollerEnd        MessageScrollerPosition = "end"
	MessageScrollerLastAnchor MessageScrollerPosition = "last-anchor"
)

type MessageScrollerMode string

const (
	MessageScrollerFollowing MessageScrollerMode = "following-bottom"
	MessageScrollerFree      MessageScrollerMode = "free-scrolling"
	MessageScrollerAnchored  MessageScrollerMode = "anchored-to-message"
	MessageScrollerJumping   MessageScrollerMode = "settling-jump"
)

type MessageScrollerAlign string

const (
	MessageScrollerAlignStart   MessageScrollerAlign = "start"
	MessageScrollerAlignCenter  MessageScrollerAlign = "center"
	MessageScrollerAlignEnd     MessageScrollerAlign = "end"
	MessageScrollerAlignNearest MessageScrollerAlign = "nearest"
)

type MessageScrollerScrollOptions struct {
	Align  MessageScrollerAlign
	Smooth bool
	// Nil uses the provider's margin; a pointer also permits an explicit zero.
	ScrollMargin *float32
}

type MessageScrollerProviderProps struct {
	// Defaults: no automatic following, opening at the end, an 8-DIP edge
	// tolerance, a 64-DIP previous-turn peek, and zero command margin.
	AutoScroll                                                bool
	DefaultScrollPosition                                     MessageScrollerPosition
	ScrollEdgeThreshold, ScrollPreviousItemPeek, ScrollMargin *float32
}

type MessageScrollerScrollable struct{ Start, End bool }
type MessageScrollerVisibility struct {
	CurrentAnchorID   string
	VisibleMessageIDs []string
}

type MessageScrollerViewportProps struct {
	Label                   string
	PreserveScrollOnPrepend *bool
	// MyGo dispatches ancestor input handlers before native child scrolling.
	// Opt in only for flat transcripts without nested native scroll containers.
	// The default preserves native nested scrolling and its normal chaining.
	ContainScrollAtEdges bool
}
type MessageScrollerContentProps struct {
	// These values are applied to the live native content. Commands account for
	// them without depending on an unavailable computed-style getter.
	PaddingStart, PaddingEnd float32
	// Nil uses Vega's 32-DIP row gap. A pointer permits an explicit zero gap.
	Gap   *float32
	Label string
}
type MessageScrollerItemProps struct {
	ID           string
	Key          any
	ScrollAnchor bool
}
type MessageScrollerButtonProps struct {
	Direction MessageScrollerPosition
	// Instant disables the source button's default smooth jump.
	Instant bool
	Label   string
	// OnClick may return false to prevent the default command.
	OnClick func() bool
}

type messageScrollerItem struct {
	id          string
	anchor      bool
	top, height float32
	measured    bool
}
type messageScrollerRowSnapshot struct {
	item        *messageScrollerItem
	id          string
	anchor      bool
	top, height float32
}
type messageScrollerRequest struct {
	id      string
	edge    MessageScrollerPosition
	options MessageScrollerScrollOptions
}

// MessageScrollerState is caller-owned. Scroll tracks the real native viewport;
// its coordinates remain usable for application restore and native inspection.
// A state belongs to one mounted provider at a time.
type MessageScrollerState struct {
	Scroll                                        ui.ScrollState
	mode                                          MessageScrollerMode
	props                                         MessageScrollerProviderProps
	items                                         []*messageScrollerItem
	request                                       *messageScrollerRequest
	pendingMessage                                *messageScrollerRequest
	itemCount                                     int
	committedItemCount                            int
	anchor, prepend                               *messageScrollerItem
	prependY                                      float32
	height, width, contentBottom, spacer          float32
	paddingStart, paddingEnd, gap                 float32
	scrollable                                    MessageScrollerScrollable
	defaultPosition                               MessageScrollerPosition
	defaultApplied, initialized, pending, mounted bool
	preserve                                      bool
	lastY                                         float32
	autoUntil                                     time.Time
	now                                           time.Time
	services                                      ui.Services
	rows                                          []messageScrollerRowSnapshot
	handled                                       map[*messageScrollerItem]bool
	lastHeight, lastBottom                        float32
	lastAcceptedY, lastMaxY                       float32
	viewport                                      ui.Rect
	root                                          ui.Rect
	content                                       ui.Rect
	owner                                         *messageScrollerOwner
	move                                          *float32
	moveSmooth                                    bool
	smoothTarget                                  *float32
	snapshotY                                     float32
	viewportMounted                               bool
	viewportOwner                                 *messageScrollerViewportOwner
	contentOwner                                  *messageScrollerContentOwner
	bar                                           ui.Rect
	barPressed                                    bool
	invalid                                       bool
	autoChanged                                   bool
}

func (s *MessageScrollerState) Scrollable() MessageScrollerScrollable { return s.scrollable }
func (s *MessageScrollerState) Visibility() MessageScrollerVisibility {
	// Compute only when requested; geometry collection does not allocate a
	// visibility snapshot for transcripts with no visibility consumer.
	v := MessageScrollerVisibility{}
	y := s.snapshotY
	line := y + scrollerValue(s.props.ScrollMargin, 0) + scrollerValue(s.props.ScrollPreviousItemPeek, 64)
	for _, item := range s.items {
		if item.id == "" || !item.measured {
			continue
		}
		if item.top+item.height > line && item.top < y+s.height {
			v.VisibleMessageIDs = append(v.VisibleMessageIDs, item.id)
		}
		if item.anchor && item.top <= line+.5 {
			v.CurrentAnchorID = item.id
		}
	}
	return v
}
func (s *MessageScrollerState) Mode() MessageScrollerMode { return s.mode }
func (s *MessageScrollerState) Pending() bool             { return s.pending }
func (s *MessageScrollerState) AutoScrolling() bool       { return s.now.Before(s.autoUntil) }
func (s *MessageScrollerState) TailSpace() float32        { return s.spacer }

func (s *MessageScrollerState) ScrollToMessage(id string, options ...MessageScrollerScrollOptions) bool {
	if id == "" || len(options) > 1 {
		return false
	}
	found := false
	for _, item := range s.items {
		if item.id == id {
			found = true
		}
	}
	if s.itemCount > 0 && !found {
		return false
	}
	o := MessageScrollerScrollOptions{}
	if len(options) > 0 {
		o = options[0]
	}
	request := &messageScrollerRequest{id: id, options: o}
	if !found || !s.viewportMounted {
		s.pendingMessage = request
	} else {
		s.request = request
		s.pendingMessage = nil
	}
	s.defaultApplied = true
	s.pending = false
	s.services.Invalidate()
	return true
}
func (s *MessageScrollerState) ScrollToStart(options ...MessageScrollerScrollOptions) bool {
	return s.requestEdge(MessageScrollerStart, options)
}
func (s *MessageScrollerState) ScrollToEnd(options ...MessageScrollerScrollOptions) bool {
	return s.requestEdge(MessageScrollerEnd, options)
}
func (s *MessageScrollerState) requestEdge(edge MessageScrollerPosition, options []MessageScrollerScrollOptions) bool {
	if !s.mounted || !s.viewportMounted || len(options) > 1 {
		return false
	}
	o := MessageScrollerScrollOptions{}
	if len(options) > 0 {
		o = options[0]
	}
	s.request = &messageScrollerRequest{edge: edge, options: o}
	s.services.Invalidate()
	return true
}
