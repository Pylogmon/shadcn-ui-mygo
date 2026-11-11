package conversation

import (
	"math"
	"time"
)

func scrollerValue(p *float32, fallback float32) float32 {
	if p == nil || !scrollerFinite(*p) {
		return fallback
	}
	return *p
}
func (s *MessageScrollerState) userIntent() {
	s.mode = MessageScrollerFree
	s.anchor = nil
	s.move = nil
	s.smoothTarget = nil
	s.request = nil
}
func (s *MessageScrollerState) setSpacer(v float32) {
	v = float32(math.Ceil(float64(max(0, v))))
	if v != s.spacer {
		s.spacer = v
		s.services.Invalidate()
	}
}
func (s *MessageScrollerState) moveTo(target float32, smooth, auto bool) {
	target = max(0, target)
	if auto && math.Abs(float64(target-s.Scroll.Y)) > .5 {
		s.autoUntil = s.now.Add(180 * time.Millisecond)
	}
	if math.Abs(float64(target-s.Scroll.Y)) > .5 || s.Scroll.Scrolling() {
		s.move = &target
		s.moveSmooth = smooth
		s.services.Invalidate()
	}
}
func (s *MessageScrollerState) edge(edge MessageScrollerPosition, smooth bool) {
	s.setSpacer(0)
	s.anchor = nil
	target := float32(0)
	s.mode = MessageScrollerFree
	if edge == MessageScrollerEnd {
		target = max(0, s.contentBottom-s.height)
		if s.props.AutoScroll {
			s.mode = MessageScrollerFollowing
		}
	}
	s.moveTo(target, smooth, edge == MessageScrollerEnd)
}
func (s *MessageScrollerState) target(item *messageScrollerItem, o MessageScrollerScrollOptions, peek bool) float32 {
	margin := scrollerValue(s.props.ScrollMargin, 0)
	if o.ScrollMargin != nil {
		margin = scrollerValue(o.ScrollMargin, margin)
	}
	if peek {
		margin += scrollerValue(s.props.ScrollPreviousItemPeek, 64)
	}
	switch o.Align {
	case MessageScrollerAlignCenter:
		return item.top - s.paddingStart - (max(0, s.height-s.paddingStart-s.paddingEnd)-item.height)/2 - margin
	case MessageScrollerAlignEnd:
		return item.top - s.height + item.height + s.paddingEnd + margin
	case MessageScrollerAlignNearest:
		if item.top >= s.Scroll.Y+s.paddingStart && item.top+item.height <= s.Scroll.Y+s.height-s.paddingEnd {
			return s.Scroll.Y
		}
		if item.top < s.Scroll.Y+s.paddingStart {
			return item.top - s.paddingStart - margin
		}
		return item.top + item.height - s.height + s.paddingEnd + margin
	default:
		return item.top - s.paddingStart - margin
	}
}
func (s *MessageScrollerState) place(item *messageScrollerItem, o MessageScrollerScrollOptions, peek bool) {
	target := s.target(item, o, peek)
	s.setSpacer(target + s.height - s.contentBottom)
	s.mode = MessageScrollerJumping
	s.anchor = nil
	if peek {
		s.mode = MessageScrollerAnchored
		s.anchor = item
	}
	s.prepend = item
	s.prependY = item.top - target
	s.moveTo(target, o.Smooth, false)
}
func (s *MessageScrollerState) reconcile() {
	previousCount := s.committedItemCount
	s.itemCount = len(s.items)
	if len(s.items) == 0 {
		s.committedItemCount = 0
		s.contentBottom = s.paddingStart + s.paddingEnd
		if request := s.request; request != nil && request.edge != "" {
			s.edge(request.edge, request.options.Smooth)
			s.request = nil
		}
		s.pending = false
		s.scrollable = MessageScrollerScrollable{}
		s.rows = nil
		s.setSpacer(0)
		return
	}
	if s.height <= 0 {
		return
	}
	for _, item := range s.items {
		if !item.measured {
			return
		}
	}
	structure := len(s.rows) != len(s.items)
	geometry := s.lastHeight != s.height || s.autoChanged
	bottom := s.paddingStart + s.paddingEnd
	for i, item := range s.items {
		bottom = max(bottom, item.top+item.height+s.paddingEnd)
		if i >= len(s.rows) {
			continue
		}
		old := s.rows[i]
		structure = structure || old.item != item || old.anchor != item.anchor || old.id != item.id
		geometry = geometry || old.top != item.top || old.height != item.height
	}
	s.contentBottom = bottom
	geometry = geometry || bottom != s.lastBottom
	threshold := scrollerValue(s.props.ScrollEdgeThreshold, 8)
	rawEnd := bottom-s.Scroll.Y-s.height > threshold
	scrolledUp := s.Scroll.Y < s.lastY-.5
	if s.props.AutoScroll && !rawEnd && s.spacer == 0 && s.mode != MessageScrollerAnchored && s.mode != MessageScrollerJumping {
		s.mode = MessageScrollerFollowing
	} else if s.mode == MessageScrollerFollowing && rawEnd && scrolledUp && (!s.AutoScrolling() || !s.Scroll.Scrolling()) {
		// A native thumb drag cancels ScrollTo before reporting its offset,
		// even while the previous automatic-scroll status is still active.
		s.userIntent()
	}
	command := false
	if request := s.pendingMessage; request != nil {
		for i := len(s.items) - 1; i >= 0; i-- {
			item := s.items[i]
			if item.id == request.id {
				s.place(item, request.options, false)
				s.pendingMessage, s.request = nil, nil
				s.defaultApplied, s.pending = true, false
				command = true
				break
			}
		}
	}
	if request := s.request; !command && request != nil {
		if request.edge != "" {
			s.edge(request.edge, request.options.Smooth)
			s.request = nil
			command = true
		} else {
			for i := len(s.items) - 1; i >= 0; i-- {
				item := s.items[i]
				if item.id == request.id {
					s.place(item, request.options, false)
					s.request = nil
					s.defaultApplied, s.pending = true, false
					command = true
					break
				}
			}
		}
	}
	if !command && !s.defaultApplied {
		anchor := (*messageScrollerItem)(nil)
		if s.defaultPosition == MessageScrollerLastAnchor {
			for _, item := range s.items {
				if item.anchor {
					anchor = item
				}
			}
		}
		if anchor != nil && bottom-anchor.top > s.height {
			s.place(anchor, MessageScrollerScrollOptions{}, true)
		} else {
			edge := MessageScrollerEnd
			if s.defaultPosition == MessageScrollerStart {
				edge = MessageScrollerStart
			}
			s.edge(edge, false)
		}
		s.defaultApplied = true
		s.pending = false
		command = true
		for _, item := range s.items {
			if item.anchor {
				s.handled[item] = true
			}
		}
	}
	// The source treats content returning after an empty transcript as its
	// first-content branch, even after the mount default was already applied.
	if !command && previousCount == 0 && s.props.AutoScroll {
		s.edge(MessageScrollerEnd, false)
		command = true
	}
	prepended := false
	if !command && structure && len(s.rows) > 0 && s.preserve {
		for i, item := range s.items {
			if item == s.rows[0].item && i > 0 {
				prepended = true
				break
			}
		}
		if prepended && s.prepend != nil {
			for _, item := range s.items {
				if item == s.prepend {
					s.moveTo(item.top-s.prependY, false, false)
					break
				}
			}
		}
	}
	if !command && !prepended && structure && len(s.rows) > 0 {
		var anchor *messageScrollerItem
		count := 0
		if len(s.items) > len(s.rows) {
			for _, item := range s.items[len(s.rows):] {
				if item.anchor {
					if anchor == nil {
						anchor = item
					}
					count++
				}
			}
		} else if len(s.items) == len(s.rows) {
			for _, item := range s.items {
				if item.anchor && !s.handled[item] {
					anchor = item
					break
				}
			}
		}
		if anchor != nil {
			if s.props.AutoScroll && s.mode == MessageScrollerFollowing && count > 1 {
				s.edge(MessageScrollerEnd, false)
			} else {
				s.place(anchor, MessageScrollerScrollOptions{}, true)
			}
			s.handled[anchor] = true
			command = true
		}
	}
	// Content mutation and native resize are distinct source notifications.
	// Prepend preservation must not suppress the resize hold when a reply
	// collapses in the same update and the native viewport clamps its offset.
	if !command && (geometry || structure) {
		if !prepended && s.mode == MessageScrollerFollowing && s.props.AutoScroll {
			s.edge(MessageScrollerEnd, false)
		} else if s.mode == MessageScrollerAnchored && s.anchor != nil {
			present := false
			for _, item := range s.items {
				if item == s.anchor {
					present = true
				}
			}
			if present {
				old := s.spacer
				s.place(s.anchor, MessageScrollerScrollOptions{}, true)
				if s.props.AutoScroll && old > 0 && s.spacer == 0 {
					s.edge(MessageScrollerEnd, false)
				}
			} else {
				s.userIntent()
			}
		}
	}
	if s.props.AutoScroll && s.spacer > 0 && s.mode == MessageScrollerFree && !s.AutoScrolling() {
		overlap := s.Scroll.Y + s.height - bottom
		if overlap > 0 {
			next := max(0, bottom+s.spacer-s.Scroll.Y-s.height)
			if next == 0 {
				s.edge(MessageScrollerEnd, false)
			} else {
				s.setSpacer(next)
			}
		}
	}
	// Snapshot the real edge, excluding tail room. Following suppresses the
	// transient end gap caused by an arriving chunk until its offset catches up.
	y := s.Scroll.Y
	if s.move != nil && !s.moveSmooth {
		y = *s.move
	}
	s.scrollable = MessageScrollerScrollable{Start: y > threshold, End: bottom-y-s.height > threshold}
	if s.mode == MessageScrollerFollowing {
		s.scrollable.End = false
	}
	s.snapshotY = y
	// Capture a real visible row after applying this build's pending command.
	for _, item := range s.items {
		if item.id != "" && item.top+item.height > y && item.top < y+s.height {
			s.prepend = item
			s.prependY = item.top - y
			break
		}
	}
	s.rows = s.rows[:0]
	present := make(map[*messageScrollerItem]bool, len(s.items))
	for _, item := range s.items {
		s.rows = append(s.rows, messageScrollerRowSnapshot{item, item.id, item.anchor, item.top, item.height})
		present[item] = true
	}
	for item := range s.handled {
		if !present[item] {
			delete(s.handled, item)
		}
	}
	s.lastHeight, s.lastBottom, s.lastY = s.height, bottom, y
	s.committedItemCount = len(s.items)
	s.autoChanged = false
}
