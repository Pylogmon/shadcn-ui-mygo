package conversation

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type messageScrollerScopeKey struct{}
type messageScrollerScope struct{ state *MessageScrollerState }

func scrollerScope(c *ui.Context) *MessageScrollerState {
	s := ui.Local(c.Root(), messageScrollerScopeKey{}, func() messageScrollerScope { return messageScrollerScope{} }).state
	if s == nil {
		panic("conversation: MessageScroller parts require a provider")
	}
	return s
}
func MessageScrollerProvider(c *ui.Context, s *MessageScrollerState, p MessageScrollerProviderProps, content func(*ui.Context)) {
	if s == nil {
		panic("conversation: MessageScrollerProvider requires caller-owned state")
	}
	scope := ui.Local(c.Root(), messageScrollerScopeKey{}, func() messageScrollerScope { return messageScrollerScope{} })
	previous := scope.state
	scope.state = s
	defer func() { scope.state = previous }()
	if s.initialized && s.props.AutoScroll != p.AutoScroll {
		s.autoChanged = true
	}
	s.props = p
	s.services = c.Services()
	s.now = c.Now()
	s.invalid = false
	for _, v := range []*float32{p.ScrollMargin, p.ScrollPreviousItemPeek, p.ScrollEdgeThreshold} {
		if v != nil && !scrollerFinite(*v) {
			s.invalid = true
		}
	}
	if !scrollerFinite(s.Scroll.X) || !scrollerFinite(s.Scroll.Y) {
		x, y := s.Scroll.X, s.Scroll.Y
		if !scrollerFinite(x) {
			x = 0
		}
		if !scrollerFinite(y) {
			y = 0
		}
		s.Scroll = ui.ScrollState{Handle: s.Scroll.Handle, X: x, Y: y, MaxX: s.Scroll.MaxX, MaxY: s.Scroll.MaxY}
		s.invalid = true
	}
	_ = c.WindowHidden()
	position := p.DefaultScrollPosition
	if position == "" {
		position = MessageScrollerEnd
	}
	if position != MessageScrollerStart && position != MessageScrollerEnd && position != MessageScrollerLastAnchor {
		position = MessageScrollerEnd
		s.invalid = true
	}
	if !s.initialized {
		s.initialized = true
		s.pending = position != MessageScrollerStart && !s.defaultApplied
		s.mode = MessageScrollerFree
		if p.AutoScroll {
			s.mode = MessageScrollerFollowing
		}
		s.preserve = true
		s.handled = map[*messageScrollerItem]bool{}
	} else if position != s.defaultPosition {
		s.defaultApplied = false
	}
	s.defaultPosition = position
	if !s.autoUntil.IsZero() {
		if !s.now.Before(s.autoUntil) {
			s.autoUntil = time.Time{}
		} else {
			c.After(s.autoUntil.Sub(s.now))
		}
	}
	if content != nil {
		content(c)
	}
}
func MessageScrollerWith(c *ui.Context, s *MessageScrollerState, p MessageScrollerProviderProps, content func(*ui.Context)) ui.Element {
	var root ui.Element
	MessageScrollerProvider(c, s, p, func(c *ui.Context) { root = MessageScrollerRoot(c, content) })
	return root
}
func MessageScrollerRoot(c *ui.Context, content func(*ui.Context)) ui.Element {
	s := scrollerScope(c)
	root := ui.Column(c).Fill().MinWidth(0).MinHeight(0).Clip().Invalid(s.invalid)
	owner := ui.Local(root, messageScrollerOwnerKey{}, func() messageScrollerOwner { return messageScrollerOwner{} })
	if owner.state != s {
		owner.Close()
		owner.state = s
	}
	s.owner = owner
	s.mounted = true
	s.root = root.LayoutBounds()
	root.Children(func() {
		if content != nil {
			content(c)
		}
	})
	messageScrollerProbe(c, s)
	return root
}
func MessageScrollerViewport(c *ui.Context, p MessageScrollerViewportProps, content func(*ui.Context)) ui.Element {
	s := scrollerScope(c)
	s.viewportMounted = true
	label := p.Label
	if label == "" {
		label = "Messages"
	}
	s.preserve = p.PreserveScrollOnPrepend == nil || *p.PreserveScrollOnPrepend
	viewport := ui.Scroll(c).Fill().MinWidth(0).MinHeight(0).TrackScroll(&s.Scroll).Role(ui.RoleRegion).Label(label).Focusable().FocusRing(false).Scrollbars(false)
	owner := ui.Local(viewport, messageScrollerViewportOwnerKey{}, func() messageScrollerViewportOwner { return messageScrollerViewportOwner{} })
	if owner.state != s {
		owner.Close()
		owner.state = s
	}
	s.viewportOwner = owner
	// CSS "thin" leaves dimensions to the UA. Map its thumb to the current
	// MyGo backend's thin width and reserve the matching track at logical end.
	// This source-mapped bar controls Scroll directly and keeps its hit target
	// when programmatic scrolling makes the thumb transparent.
	thumbWidth := c.Theme().ScrollbarWidth
	if !scrollerFinite(thumbWidth) || thumbWidth <= 0 {
		thumbWidth = 6
	}
	trackWidth := max(float32(4), thumbWidth+3)
	viewport.PaddingEnd(trackWidth)
	s.viewport = viewport.LayoutBounds()
	s.width, s.height = s.viewport.W, s.viewport.H
	if s.pending {
		viewport.Opacity(0).IgnoreInput(true).AccessHidden(true)
	}
	if s.height > 0 && s.Scroll.MaxY > 0 {
		fraction := min(float32(.12), 40/s.height) * ui.CubicBezier(.42, 0, .58, 1)(min(float32(1), max(float32(0), (s.Scroll.MaxY-s.Scroll.Y)/96)))
		viewport.MaskLinearGradient(ui.LinearGradient{Angle: 180, Stops: []ui.GradientStop{
			{Color: ui.RGB(0, 0, 0), At: 0}, {Color: ui.RGB(0, 0, 0), At: 1 - fraction}, {Color: ui.Transparent, At: 1},
		}})
	}
	services := c.Services()
	viewport.Draw(func(_ *ui.Painter, r ui.Rect) {
		if s.lastAcceptedY != s.Scroll.Y || s.lastMaxY != s.Scroll.MaxY || s.height != r.H {
			s.lastAcceptedY, s.lastMaxY = s.Scroll.Y, s.Scroll.MaxY
			services.Invalidate()
		}
	})
	viewport.HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind == ui.InputScroll || ev.Kind == ui.InputKeyDown && messageScrollerScrollKey(ev.Key) {
			s.userIntent()
		}
		// Native child scroll containers must get their default routing. The
		// flat-transcript opt-in cannot distinguish them at an outer edge.
		return p.ContainScrollAtEdges && ev.Kind == ui.InputScroll && (ev.DY < 0 && s.Scroll.Y <= 0 || ev.DY > 0 && s.Scroll.Y >= s.Scroll.MaxY)
	})
	viewport.ObserveInput(func(ev ui.InputEvent) {
		if ev.Kind == ui.InputKeyDown && messageScrollerScrollKey(ev.Key) {
			s.userIntent()
		}
	})
	// Prepare captured bar input before reconciling content: the public bar
	// may restore a queued accepted offset while canceling a native glide.
	bar := messageScrollerBar(c, s, trackWidth)
	viewport.Children(func() {
		if content != nil {
			content(c)
		}
	})
	if s.contentOwner == nil && s.request != nil && s.request.edge != "" {
		s.contentBottom = s.Scroll.MaxY + s.height
		request := s.request
		s.edge(request.edge, request.options.Smooth)
		s.request = nil
	}
	if motion.Reduced(c) && s.smoothTarget != nil && s.Scroll.Scrolling() {
		s.move = s.smoothTarget
		s.moveSmooth = false
		s.smoothTarget = nil
	}
	if s.move != nil {
		target := *s.move
		if !s.moveSmooth || motion.Reduced(c) {
			// Retargeting a native glide to its accepted coordinate cancels
			// its velocity before the immediate assignment is adopted.
			if s.Scroll.Scrolling() {
				s.Scroll.ScrollTo(s.Scroll.X, s.Scroll.Y, ui.Animation{Duration: time.Nanosecond})
				viewport.TrackScroll(&s.Scroll)
			}
			s.Scroll.Y = target
			viewport.TrackScroll(&s.Scroll)
			s.move = nil
		} else if target <= s.Scroll.MaxY+.5 {
			s.Scroll.ScrollTo(s.Scroll.X, target, ui.Smooth)
			viewport.TrackScroll(&s.Scroll)
			s.smoothTarget = &target
			s.move = nil
		}
	}
	if s.AutoScrolling() {
		bar.Opacity(0)
	}
	return viewport
}
func messageScrollerBar(c *ui.Context, s *MessageScrollerState, trackWidth float32) ui.Element {
	color := theme.TokensOf(c).Border
	var bar ui.Element
	ui.Box(c.Key(messageScrollerBarKey{})).Absolute().End(0).Top(s.viewport.Y-s.root.Y).Size(trackWidth, s.height).Role(ui.RoleNone).PassThrough().HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind != ui.InputScroll {
			return false
		}
		s.userIntent()
		return ev.DY < 0 && s.Scroll.Y <= 0 || ev.DY > 0 && s.Scroll.Y >= s.Scroll.MaxY
	}).Children(func() {
		bar = display.ScrollBar(c, &s.Scroll, display.ScrollBarProps{Thickness: trackWidth, ViewportSize: s.height, ThumbColor: &color, OnScrollIntent: s.userIntent})
		bar.ObserveInput(func(ev ui.InputEvent) {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
				s.barPressed = true
			}
			if ev.Kind == ui.InputPointerUp && ev.Button == 0 {
				s.barPressed = false
			}
		})
		s.bar = bar.LayoutBounds()
	})
	return bar
}

func MessageScrollerContent(c *ui.Context, p MessageScrollerContentProps, content func(*ui.Context)) ui.Element {
	s := scrollerScope(c)
	s.paddingStart, s.paddingEnd = p.PaddingStart, p.PaddingEnd
	valid := scrollerFinite(s.paddingStart) && scrollerFinite(s.paddingEnd) && (p.Gap == nil || scrollerFinite(*p.Gap))
	if !scrollerFinite(s.paddingStart) {
		s.paddingStart = 0
	}
	if !scrollerFinite(s.paddingEnd) {
		s.paddingEnd = 0
	}
	s.paddingStart, s.paddingEnd = max(0, s.paddingStart), max(0, s.paddingEnd)
	s.gap = 32
	if p.Gap != nil {
		s.gap = max(0, scrollerValue(p.Gap, 32))
	}
	root := ui.Column(c).FillWidth().MinWidth(0).MinHeightPercent(100).Shrink(0).Gap(s.gap).Padding(s.paddingStart, 0, s.paddingEnd, 0).Role(ui.RoleLog).Invalid(!valid)
	owner := ui.Local(root, messageScrollerContentOwnerKey{}, func() messageScrollerContentOwner { return messageScrollerContentOwner{} })
	if owner.state != s {
		owner.Close()
		owner.state = s
	}
	s.contentOwner = owner
	if p.Label != "" {
		root.Label(p.Label)
	}
	s.content = root.LayoutBounds()
	s.items = nil
	root.Children(func() {
		if content != nil {
			content(c)
		}
		s.reconcile()
		if s.spacer > 0 {
			ui.Box(c.Key(messageScrollerSpacerKey{})).Height(s.spacer).Margin(-s.gap, 0, 0, 0).Shrink(0).AccessHidden(true).IgnoreInput(true)
		}
	})
	return root
}
func MessageScrollerItem(c *ui.Context, p MessageScrollerItemProps, content func(*ui.Context)) ui.Element {
	s := scrollerScope(c)
	root := ui.Column(c).FillWidth().MinWidth(0).Shrink(0)
	if p.Key != nil {
		root.Key(p.Key)
	} else if p.ID != "" {
		root.Key(p.ID)
	}
	item := ui.Local(root, messageScrollerItemKey{}, func() messageScrollerItem { return messageScrollerItem{} })
	item.id, item.anchor = p.ID, p.ScrollAnchor
	r := root.LayoutBounds()
	item.measured = r.W > 0
	item.top, item.height = r.Y-s.content.Y, r.H
	s.items = append(s.items, item)
	return root.Children(func() {
		if content != nil {
			content(c)
		}
	})
}

type messageScrollerItemKey struct{}
type messageScrollerBarKey struct{}
type messageScrollerSpacerKey struct{}
type messageScrollerOwnerKey struct{}
type messageScrollerOwner struct{ state *MessageScrollerState }
type messageScrollerViewportOwnerKey struct{}
type messageScrollerViewportOwner struct{ state *MessageScrollerState }
type messageScrollerContentOwnerKey struct{}
type messageScrollerContentOwner struct{ state *MessageScrollerState }

func (o *messageScrollerContentOwner) Close() error {
	if s := o.state; s != nil && s.contentOwner == o {
		s.clearContent()
		s.services.Invalidate()
	}
	o.state = nil
	return nil
}
func (s *MessageScrollerState) clearContent() {
	s.contentOwner = nil
	s.items, s.rows = nil, nil
	s.anchor, s.prepend = nil, nil
	s.handled = map[*messageScrollerItem]bool{}
	s.spacer = 0
	s.scrollable = MessageScrollerScrollable{}
	s.contentBottom = 0
	s.request = nil
	s.move, s.smoothTarget = nil, nil
	s.autoUntil = time.Time{}
	s.Scroll = ui.ScrollState{Handle: s.Scroll.Handle, X: s.Scroll.X, Y: s.Scroll.Y, MaxX: s.Scroll.MaxX, MaxY: s.Scroll.MaxY}
}

func (o *messageScrollerViewportOwner) Close() error {
	if s := o.state; s != nil && s.viewportOwner == o {
		s.unmountViewport()
		s.services.Invalidate()
	}
	o.state = nil
	return nil
}
func (s *MessageScrollerState) unmountViewport() {
	s.viewportMounted = false
	s.viewportOwner = nil
	s.autoUntil = time.Time{}
	s.move, s.smoothTarget = nil, nil
	s.clearContent()
	s.mode = MessageScrollerFree
	s.scrollable = MessageScrollerScrollable{}
	s.barPressed = false
	s.request = nil
	s.Scroll = ui.ScrollState{Handle: s.Scroll.Handle, X: s.Scroll.X, Y: s.Scroll.Y, MaxX: s.Scroll.MaxX, MaxY: s.Scroll.MaxY}
}

func (o *messageScrollerOwner) Close() error {
	if s := o.state; s != nil && s.owner == o {
		s.mounted = false
		s.unmountViewport()
		s.owner = nil
	}
	o.state = nil
	return nil
}

type messageScrollerProbeKey struct{}
type messageScrollerSentinelKey struct{ state *MessageScrollerState }
type messageScrollerProbeWave struct {
	seen            map[*MessageScrollerState]bool
	drawn, followup bool
}
type messageScrollerProbeCoordinator struct {
	wave       *messageScrollerProbeWave
	correction bool
}

func messageScrollerProbe(c *ui.Context, s *MessageScrollerState) {
	coordinator := ui.Local(c.Root(), messageScrollerProbeKey{}, func() messageScrollerProbeCoordinator { return messageScrollerProbeCoordinator{} })
	wave := coordinator.wave
	if wave == nil || wave.drawn || wave.seen[s] {
		wave = &messageScrollerProbeWave{seen: map[*MessageScrollerState]bool{}, followup: !coordinator.correction}
		coordinator.wave = wave
	}
	wave.seen[s] = true
	services := c.Services()
	ui.Overlay(c, func() {
		ui.Box(c).Key(messageScrollerSentinelKey{s}).Absolute().Left(0).Top(0).Size(1, 1).Role(ui.RoleNone).AccessHidden(true).IgnoreInput(true).PassThrough().Draw(func(_ *ui.Painter, _ ui.Rect) {
			if coordinator.wave != wave || wave.drawn {
				return
			}
			wave.drawn = true
			wave.seen = nil
			coordinator.correction = wave.followup
			if wave.followup {
				services.Invalidate()
			}
		})
	})
}
func messageScrollerScrollKey(key ui.Key) bool {
	switch key {
	case ui.KeyUp, ui.KeyDown, ui.KeyHome, ui.KeyEnd, ui.KeyPageUp, ui.KeyPageDown, ui.KeySpace:
		return true
	}
	return false
}

func scrollerFinite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
