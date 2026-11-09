package display

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
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
	motion                                     carouselControllerMotion
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
	dragDenied                                 bool
	invalidProps                               bool
	displayedDrag                              float64
	focusedSlide                               ui.Handle
	viewportBox                                ui.Rect
}

type carouselIntersection struct {
	intersecting bool
	threshold    int
}
type carouselObservedPointer struct {
	down  bool
	event CarouselPointerEvent
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
	drag                  float64
	forceReInit           bool
	startIndex            *int
	onSelect              func(int)
	invalidOptions        bool
	previousOptions       CarouselOptions
}

func (engine *carouselControllerEngine) snapshot(api *CarouselController) CarouselControllerSnapshot {
	g := engine.geometry
	s := CarouselControllerSnapshot{Ready: engine.ready, Active: engine.resolved.Active, Selected: engine.selected, Previous: engine.previous, Progress: g.progress(engine.location + engine.displayedDrag), Root: engine.root, Container: engine.track, Region: engine.region, InView: slices.Clone(engine.inView), NotInView: slices.Clone(engine.notInView), Error: engine.lastError}
	s.SnapOffsets = slices.Clone(g.snaps)
	for _, snap := range g.snaps {
		s.SnapList = append(s.SnapList, g.progress(snap))
	}
	for _, item := range engine.selectedItems {
		s.Slides = append(s.Slides, item.public)
	}
	s.CanScrollPrevious = g.counter(engine.selected-1) != engine.selected
	s.CanScrollNext = g.counter(engine.selected+1) != engine.selected
	s.Engine = CarouselEngineSnapshot{Location: engine.location + engine.displayedDrag, OffsetLocation: engine.offset + engine.displayedDrag, PreviousLocation: engine.previousLocation, Target: engine.target + engine.displayedDrag, Viewport: g.view, Content: g.total, Options: engine.resolved, SlideRegistry: g.registry}
	return s
}

func (frame *carouselControllerBuild) publish(event CarouselEventType) {
	frame.api.publish(frame.engine.snapshot(frame.api), event)
	if event == CarouselSelect {
		frame.root.SetChanged()
		if frame.onSelect != nil {
			frame.onSelect(frame.engine.selected)
		}
	}
}

func (frame *carouselControllerBuild) scroll(index, direction int, jump bool, now time.Time, duration time.Duration) {
	e := frame.engine
	g := e.geometry
	if !e.ready || !e.resolved.Active || frame.api.Snapshot().Destroyed || len(g.snaps) == 0 {
		return
	}
	index = g.counter(index)
	distance := g.shortcut(g.snaps[index]-e.target, direction)
	changed := index != e.selected
	e.target += distance
	if distance != 0 {
		if jump {
			// The source's inline zero-duration render runs before its index
			// update. A previously scheduled animation still gets its settle.
			active := e.motion.active
			direction := carouselSign(e.target - e.location)
			e.previousLocation = e.target
			e.location = e.target
			frame.publish(CarouselScroll)
			e.offset = e.location
			e.wrapCoordinates(direction)
			e.motion = carouselControllerMotion{from: e.location, to: e.target, at: now, duration: duration, active: active}
		} else {
			e.motion = carouselControllerMotion{from: e.location, to: e.target, at: now, duration: duration, active: true}
		}
	}
	if changed {
		e.previous, e.selected = e.selected, index
		frame.publish(CarouselSelect)
	} else {
		frame.publish("")
	}
}

func (frame *carouselControllerBuild) tick(now time.Time, instant bool) bool {
	e := frame.engine
	if !e.ready || !e.motion.active || now == e.lastTick && !instant {
		return e.motion.active
	}
	if !instant && !now.After(e.motion.at) {
		return true
	}
	e.lastTick = now
	if math.Abs(e.target-e.offset) < .001 {
		e.motion.active = false
		frame.publish(CarouselSettle)
		return false
	}
	fraction := float64(1)
	if !instant && e.motion.duration > 0 {
		fraction = float64(now.Sub(e.motion.at)) / float64(e.motion.duration)
		fraction = max(0, min(1, fraction))
		fraction = float64(motion.EaseOut(float32(fraction)))
	}
	e.previousLocation = e.location
	e.location = e.motion.from + (e.motion.to-e.motion.from)*fraction
	frame.publish(CarouselScroll)
	e.offset = e.location
	e.wrapCoordinates(carouselSign(e.location - e.previousLocation))
	frame.publish("")
	return true
}

func (e *carouselControllerEngine) wrapCoordinates(direction int) {
	if e.geometry.loop {
		minimum, maximum := e.geometry.limit()
		shift := float64(0)
		if direction > 0 && e.location > maximum+.1 {
			shift = -e.geometry.total
		} else if direction < 0 && e.location < minimum+.1 {
			shift = e.geometry.total
		}
		if shift != 0 {
			e.location += shift
			e.offset += shift
			e.previousLocation += shift
			e.target += shift
			e.motion.from += shift
			e.motion.to += shift
		}
	}
}

func carouselOptionsSame(a, b CarouselOptions) bool {
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
