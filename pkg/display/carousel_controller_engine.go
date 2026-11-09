package display

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/carouselphysics"
)

type carouselSelectedItem struct{ canonical, public ui.Handle }

type carouselControllerEngine struct {
	ready, initialized                         bool
	options                                    CarouselOptions
	resolved                                   CarouselResolvedOptions
	geometry                                   carouselEngineGeometry
	selectedItems                              []carouselSelectedItem
	selected, previous                         int
	lastControlled                             int
	location, offset, previousLocation, target float64
	lastTick                                   time.Time
	root, track, region                        ui.Handle
	inView, notInView                          []int
	observedLayout                             carouselGeometry
	observedHandles                            []ui.Handle
	resizeBaseline                             carouselGeometry
	resizeBaselineHandles                      []ui.Handle
	lastProps                                  CarouselOptions
	propsSet                                   bool
	lastEnvironment                            CarouselEnvironment
	pressed, focusWithin                       bool
	lastTab                                    time.Time
	lastError                                  error
	intersections                              []carouselIntersection
	pendingTab                                 bool
	pointerEvents                              []carouselObservedPointer
	invalidProps                               bool
	focusedSlide                               ui.Handle
	viewportBox                                ui.Rect
	body                                       *carouselphysics.Body
	scheduler                                  carouselphysics.Scheduler
	epoch, scheduledAt                         time.Time
	hidden                                     bool
	captured                                   bool
	tracker                                    carouselphysics.DragTracker
	pointerStart, pointerLast, crossStart      float64
	pointerMouse, suppressClick, preventScroll bool
}

type carouselIntersection struct {
	intersecting bool
	threshold    int
}
type carouselObservedPointer struct {
	kind      ui.InputKind
	event     CarouselPointerEvent
	delivered bool
}

type carouselControllerBuild struct {
	api                   *CarouselController
	engine                *carouselControllerEngine
	lease                 *carouselControllerLease
	hooks                 CarouselWatchHooks
	options               *CarouselOptions
	optionsError          error
	contextEnvironment    CarouselEnvironment
	root, viewport, track ui.Element
	items                 []ui.Element
	captureHeld           bool
	forceReInit           bool
	startIndex            *int
	onSelect              func(int)
	invalidOptions        bool
	previousOptions       CarouselOptions
}

func (engine *carouselControllerEngine) snapshot(api *CarouselController) CarouselControllerSnapshot {
	g := engine.geometry
	s := CarouselControllerSnapshot{Ready: engine.ready, Active: engine.resolved.Active, Selected: engine.selected, Previous: engine.previous, Progress: g.progress(engine.location), Root: engine.root, Container: engine.track, Region: engine.region, InView: slices.Clone(engine.inView), NotInView: slices.Clone(engine.notInView), Error: engine.lastError}
	s.SnapOffsets = slices.Clone(g.snaps)
	for _, snap := range g.snaps {
		s.SnapList = append(s.SnapList, g.progress(snap))
	}
	for _, item := range engine.selectedItems {
		s.Slides = append(s.Slides, item.public)
	}
	s.CanScrollPrevious = g.counter(engine.selected-1) != engine.selected
	s.CanScrollNext = g.counter(engine.selected+1) != engine.selected
	s.Engine = CarouselEngineSnapshot{Location: engine.location, OffsetLocation: engine.offset, PreviousLocation: engine.previousLocation, Target: engine.target, Viewport: g.view, Content: g.total, Options: engine.resolved, SlideRegistry: g.registry}
	if engine.body != nil {
		s.Engine.Velocity = engine.body.Velocity()
		s.Engine.Duration = engine.body.Duration()
		s.Engine.Friction = engine.body.Friction()
		s.Engine.PointerDown = engine.captured
		s.Engine.ClickSuppressed = engine.suppressClick
	}
	return s
}

func (frame *carouselControllerBuild) publish(event CarouselEventType) {
	frame.engine.syncPhysics()
	frame.api.publish(frame.engine.snapshot(frame.api), event)
	if event == CarouselSelect {
		frame.root.SetChanged()
		if frame.onSelect != nil {
			frame.onSelect(frame.engine.selected)
		}
	}
}

func (frame *carouselControllerBuild) scroll(index, direction int, jump bool, now time.Time) {
	e := frame.engine
	g := e.geometry
	if !e.ready || !e.resolved.Active || frame.api.Snapshot().Destroyed || len(g.snaps) == 0 {
		return
	}
	index = g.counter(index)
	if e.body == nil {
		return
	}
	e.body.UseBaseFriction().UseBaseDuration()
	if jump {
		e.body.UseDuration(0)
	}
	target := e.targeter().ByIndex(e.body.State.Target, index, float64(direction))
	frame.applyPhysicsTarget(target, now)
}

func (frame *carouselControllerBuild) tick(now time.Time, instant bool) bool {
	e := frame.engine
	if !e.ready || e.body == nil || !e.scheduler.Active() {
		return false
	}
	if !instant && (now == e.lastTick || !now.After(e.scheduledAt)) {
		return true
	}
	e.lastTick = now
	if instant {
		e.body.UseDuration(0)
		frame.physicsUpdate()
		frame.physicsRender(1)
		if e.scheduler.Active() {
			frame.physicsUpdate()
			frame.physicsRender(1)
		}
	} else {
		e.scheduler.Advance(float64(now.Sub(e.epoch))/float64(time.Millisecond), frame.physicsUpdate, frame.physicsRender)
	}
	frame.publish("")
	return e.scheduler.Active()
}

func (e *carouselControllerEngine) syncPhysics() {
	if e.body != nil {
		e.location, e.offset, e.previousLocation, e.target = e.body.State.Location, e.body.State.Offset, e.body.State.Previous, e.body.State.Target
	}
}
func (e *carouselControllerEngine) targeter() carouselphysics.Targeter {
	minimum, maximum := e.geometry.limit()
	return carouselphysics.Targeter{Snaps: e.geometry.snaps, Content: e.geometry.total, Limit: carouselphysics.Limit{Min: minimum, Max: maximum}, Loop: e.geometry.loop}
}
func (e *carouselControllerEngine) bounds() carouselphysics.Bounds {
	return carouselphysics.Bounds{Limit: e.targeter().Limit, View: e.geometry.view}
}
func (frame *carouselControllerBuild) startPhysics(now time.Time) {
	e := frame.engine
	if !e.scheduler.Active() {
		e.scheduledAt = now
	}
	e.scheduler.Start()
}
func (frame *carouselControllerBuild) physicsUpdate() {
	e := frame.engine
	if !e.geometry.loop {
		e.bounds().Constrain(e.body, e.captured)
	}
	e.body.Seek()
	e.syncPhysics()
}
func (frame *carouselControllerBuild) physicsRender(alpha float64) {
	e := frame.engine
	settled := e.body.Settled()
	if !e.geometry.loop {
		settled = settled && !e.bounds().ShouldConstrain(e.body)
	}
	if settled && !e.captured {
		e.scheduler.Stop()
		frame.publish(CarouselSettle)
	}
	if !settled {
		frame.publish(CarouselScroll)
	}
	e.body.Interpolate(alpha)
	if e.geometry.loop {
		e.body.Wrap(e.geometry.total, e.targeter().Limit)
	}
	e.syncPhysics()
}
func (frame *carouselControllerBuild) applyPhysicsTarget(target carouselphysics.Target, now time.Time) {
	e := frame.engine
	changed := target.Index != e.selected
	e.body.State.Target += target.Distance
	if target.Distance != 0 {
		if e.body.Duration() != 0 {
			frame.startPhysics(now)
		} else {
			frame.physicsUpdate()
			frame.physicsRender(1)
			frame.physicsUpdate()
		}
	}
	if changed {
		e.previous, e.selected = e.selected, target.Index
		frame.publish(CarouselSelect)
	} else {
		frame.publish("")
	}
}

func carouselOptionsSame(a, b CarouselOptions) bool {
	if !carouselOptionalSame(a.Duration, b.Duration) || !carouselOptionalSame(a.DragThreshold, b.DragThreshold) || !carouselOptionalSame(a.DragFree, b.DragFree) || !carouselOptionalSame(a.SkipSnaps, b.SkipSnaps) {
		return false
	}
	if !carouselOptionalSame(a.Align, b.Align) || !carouselOptionalSame(a.Axis, b.Axis) || !carouselOptionalSame(a.Direction, b.Direction) || !carouselOptionalSame(a.Contain, b.Contain) || !carouselOptionalSame(a.SlidesToScroll, b.SlidesToScroll) || !carouselOptionalSame(a.StartIndex, b.StartIndex) || !carouselOptionalSame(a.Active, b.Active) || !carouselOptionalSame(a.Loop, b.Loop) || !carouselOptionalSame(a.WatchDrag, b.WatchDrag) || !carouselOptionalSame(a.WatchResize, b.WatchResize) || !carouselOptionalSame(a.WatchSlides, b.WatchSlides) || !carouselOptionalSame(a.WatchFocus, b.WatchFocus) {
		return false
	}
	if (a.InViewThreshold == nil) != (b.InViewThreshold == nil) {
		return false
	}
	if a.InViewThreshold != nil && !slices.Equal(*a.InViewThreshold, *b.InViewThreshold) {
		return false
	}
	if (a.Nodes == nil) != (b.Nodes == nil) {
		return false
	}
	if a.Nodes != nil {
		if !carouselOptionalSame(a.Nodes.Container, b.Nodes.Container) || (a.Nodes.Slides == nil) != (b.Nodes.Slides == nil) {
			return false
		}
		if a.Nodes.Slides != nil && !slices.Equal(*a.Nodes.Slides, *b.Nodes.Slides) {
			return false
		}
	}
	if len(a.Breakpoints) != len(b.Breakpoints) {
		return false
	}
	for i, x := range a.Breakpoints {
		y := b.Breakpoints[i]
		if x.Key != y.Key || !carouselMediaSame(x.When, y.When) || !carouselOptionsSame(x.Options, y.Options) {
			return false
		}
	}
	return true
}

func carouselOptionalSame[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
func carouselMediaSame(a, b CarouselMedia) bool {
	return carouselOptionalSame(a.MinWidth, b.MinWidth) && carouselOptionalSame(a.MaxWidth, b.MaxWidth) && carouselOptionalSame(a.MinHeight, b.MinHeight) && carouselOptionalSame(a.MaxHeight, b.MaxHeight) && carouselOptionalSame(a.Dark, b.Dark) && carouselOptionalSame(a.Reduced, b.Reduced) && carouselOptionalSame(a.Direction, b.Direction)
}

// Input handlers retain only this value record and the engine's model pointer.
// ObserveInput supplies descendant interaction notifications; a non-consuming
// HandleInput supplies the captured viewport's complete move/up stream.
func (e *carouselControllerEngine) recordPointer(event ui.InputEvent, delivered bool) {
	if event.Kind != ui.InputPointerDown && event.Kind != ui.InputPointerUp && (event.Kind != ui.InputPointerMove || event.Button != 0) {
		return
	}
	value := CarouselPointerEvent{Position: ui.Point{X: event.X, Y: event.Y}, Kind: event.Pointer, Cancelled: event.Canceled, Button: event.Button, Clicks: event.Clicks, Modifiers: event.Mods, Time: event.Time, Pressure: event.Pressure}
	if n := len(e.pointerEvents); n > 0 && event.Kind != ui.InputPointerMove {
		last := &e.pointerEvents[n-1]
		if last.kind == event.Kind && last.event.Time == value.Time && last.event.Position == value.Position {
			last.delivered = last.delivered || delivered
			// MyGo ObserveInput omits Pointer/Pressure; HandleInput carries the
			// actual device. Prefer its complete value when coalescing delivery.
			if delivered {
				last.event = value
			}
			return
		}
	}
	e.pointerEvents = append(e.pointerEvents, carouselObservedPointer{kind: event.Kind, event: value, delivered: delivered})
}
func (frame *carouselControllerBuild) processPointers(c *ui.Context, b *carouselCompositionBuild) {
	e := frame.engine
	events := e.pointerEvents
	e.pointerEvents = nil
	canDrag := e.ready && len(e.geometry.snaps) > 0 && e.resolved.Active && e.resolved.WatchDrag && !b.p.NoDrag && !b.p.Disabled && !frame.api.Snapshot().Destroyed
	for n, observed := range events {
		ev := observed.event
		point, cross := float64(ev.Position.X), float64(ev.Position.Y)
		if e.resolved.Axis == CarouselVertical {
			point, cross = cross, point
		}
		nowMS := float64(ev.Time.Sub(e.epoch)) / float64(time.Millisecond)
		switch observed.kind {
		case ui.InputPointerDown:
			if !canDrag {
				continue
			}
			allowed := frame.hooks.BeforeDrag == nil || frame.hooks.BeforeDrag(frame.api, ev)
			if !allowed || ev.Button != 0 || c.FocusTakesText() {
				continue
			}
			e.pressed = true
			e.pointerStart, e.crossStart = point, cross
			e.pointerMouse = ev.Kind == ui.PointerMouse
			e.suppressClick, e.preventScroll = false, false
			captured := frame.captureHeld
			// A completed batch may no longer report held. Only a delivered move/up
			// proves that this viewport, rather than a descendant button, won capture.
			for _, later := range events[n+1:] {
				if later.kind == ui.InputPointerDown {
					break
				}
				if later.delivered && (later.kind == ui.InputPointerMove || later.kind == ui.InputPointerUp) {
					captured = true
					break
				}
			}
			if captured {
				e.captured = true
				e.tracker.Down(point, nowMS)
				e.body.UseFriction(0).UseDuration(0)
				e.body.State.Target = e.body.State.Location
			}
			frame.publish(CarouselPointerDown)
		case ui.InputPointerMove:
			if !canDrag || !e.captured || !e.pressed || !observed.delivered {
				continue
			}
			if !e.pointerMouse && !e.preventScroll {
				e.preventScroll = math.Abs(point-e.pointerStart) > math.Abs(cross-e.crossStart)
				if !e.preventScroll {
					frame.releasePointer(c.Now(), nowMS, true)
					continue
				}
			}
			diff := e.tracker.Move(point, nowMS)
			if math.Abs(point-e.pointerStart) > e.resolved.DragThreshold {
				e.suppressClick = true
			}
			if e.resolved.Axis == CarouselHorizontal && e.resolved.Direction == CarouselRTL {
				diff = -diff
			}
			e.body.UseFriction(.3).UseDuration(.75)
			e.body.State.Target += diff
			frame.startPhysics(c.Now())
		case ui.InputPointerUp:
			if e.pressed {
				frame.releasePointer(c.Now(), nowMS, ev.Cancelled)
			}
		}
	}
	if e.captured && (!canDrag || !c.WindowActive() || frame.viewport.PressCanceled()) {
		frame.releasePointer(c.Now(), float64(c.Now().Sub(e.epoch))/float64(time.Millisecond), true)
	}
}
func (frame *carouselControllerBuild) releasePointer(now time.Time, timeMS float64, canceled bool) {
	e := frame.engine
	captured := e.captured
	e.captured, e.pressed, e.preventScroll = false, false, false
	if captured && e.body != nil {
		if canceled {
			// Native cancellation undoes the gesture instead of committing a fling.
			e.suppressClick = false
			e.body.UseBaseDuration().UseBaseFriction()
			frame.applyPhysicsTarget(e.targeter().ByIndex(e.body.State.Target, e.selected, 0), now)
		} else {
			sign := float64(1)
			if e.resolved.Axis == CarouselHorizontal && e.resolved.Direction == CarouselRTL {
				sign = -1
			}
			release := e.targeter().Release(e.body.State.Target, e.selected, e.tracker.Up(timeMS), e.geometry.view, e.pointerMouse, e.resolved.DragFree, e.resolved.SkipSnaps, sign)
			e.body.UseDuration(release.Duration).UseFriction(release.Friction)
			frame.applyPhysicsTarget(release.Target, now)
		}
	}
	e.pointerMouse = false
	frame.publish(CarouselPointerUp)
}
