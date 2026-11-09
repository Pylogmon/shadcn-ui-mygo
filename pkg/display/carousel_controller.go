package display

import (
	"slices"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

type CarouselEventType string

const (
	CarouselInit            CarouselEventType = "init"
	CarouselPointerDown     CarouselEventType = "pointerDown"
	CarouselPointerUp       CarouselEventType = "pointerUp"
	CarouselSlidesChanged   CarouselEventType = "slidesChanged"
	CarouselSlidesInView    CarouselEventType = "slidesInView"
	CarouselScroll          CarouselEventType = "scroll"
	CarouselSelect          CarouselEventType = "select"
	CarouselSettle          CarouselEventType = "settle"
	CarouselDestroy         CarouselEventType = "destroy"
	CarouselReInit          CarouselEventType = "reInit"
	CarouselResize          CarouselEventType = "resize"
	CarouselSlideFocusStart CarouselEventType = "slideFocusStart"
	CarouselSlideFocus      CarouselEventType = "slideFocus"
)

// CarouselListener has explicit application-owned identity, like a JavaScript
// function object. On may register it repeatedly; Off removes all registrations
// of that same object. Its callback must not capture a Context or Element.
type CarouselListener struct {
	callback func(*CarouselController, CarouselEventType)
}

func NewCarouselListener(callback func(*CarouselController, CarouselEventType)) *CarouselListener {
	return &CarouselListener{callback: callback}
}

type CarouselEngineSnapshot struct {
	Location, OffsetLocation, PreviousLocation, Target float64
	Viewport, Content                                  float64
	SlideRegistry                                      [][]int
	Options                                            CarouselResolvedOptions
}

// CarouselControllerSnapshot is detached committed state. SnapList contains
// source-normalized progress values; SnapOffsets contains native DIPs. Progress
// may be outside 0..1 while overscrolling. Visible lists use engine slide indices.
type CarouselControllerSnapshot struct {
	Ready, Destroyed, Active         bool
	Generation                       uint64
	Selected, Previous               int
	SnapList, SnapOffsets            []float64
	Progress                         float64
	InView, NotInView                []int
	CanScrollPrevious, CanScrollNext bool
	Root, Container, Region          ui.Handle
	Slides                           []ui.Handle
	Engine                           CarouselEngineSnapshot
	Error                            error
}

type carouselCommandKind uint8

const (
	carouselCommandTo carouselCommandKind = iota
	carouselCommandNext
	carouselCommandPrevious
	carouselCommandConfigure
	carouselCommandReInit
)

type carouselCommand struct {
	kind  carouselCommandKind
	index int
	jump  bool
	patch CarouselOptions
}

// CarouselController belongs to application state, never a frame. Scroll and
// configuration commands are queued atomically and wake their owner through
// public weak Services; the next owner build processes every command in order.
// Snapshot reads and command enqueueing are safe from background goroutines.
// Emit/Destroy and listener callbacks run on the UI thread. Use Window.Update
// when invoking those methods from asynchronous application work.
type CarouselController struct {
	mu             sync.Mutex
	options        CarouselOptions
	appliedOptions CarouselOptions
	appliedSet     bool
	commands       []carouselCommand
	listeners      map[CarouselEventType][]*CarouselListener
	snapshot       CarouselControllerSnapshot
	services       ui.Services
	owner          ui.Handle
	bound          bool
	generation     uint64
}

func NewCarouselController(options CarouselOptions) (*CarouselController, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	return &CarouselController{options: carouselCloneOptions(options)}, nil
}

func (a *CarouselController) Configure(patch CarouselOptions) error {
	return a.configure(patch, carouselCommandConfigure)
}

func (a *CarouselController) ReInit(patch CarouselOptions) error {
	return a.configure(patch, carouselCommandReInit)
}

func (a *CarouselController) configure(patch CarouselOptions, kind carouselCommandKind) error {
	if err := patch.Validate(); err != nil {
		return err
	}
	patch = carouselCloneOptions(patch)
	a.mu.Lock()
	if a.snapshot.Destroyed {
		a.mu.Unlock()
		return &CarouselOptionError{Field: "Controller", Code: CarouselDestroyed}
	}
	a.options = carouselMergeOptions(a.options, patch)
	if a.bound {
		a.commands = append(a.commands, carouselCommand{kind: kind, patch: patch})
	}
	services := a.services
	a.mu.Unlock()
	services.Invalidate()
	return nil
}

func (a *CarouselController) enqueue(command carouselCommand) {
	if a == nil {
		return
	}
	a.mu.Lock()
	if a.snapshot.Destroyed {
		a.mu.Unlock()
		return
	}
	a.commands = append(a.commands, command)
	services := a.services
	a.mu.Unlock()
	services.Invalidate()
}

func (a *CarouselController) ScrollPrev(jump bool) {
	a.enqueue(carouselCommand{kind: carouselCommandPrevious, jump: jump})
}
func (a *CarouselController) ScrollNext(jump bool) {
	a.enqueue(carouselCommand{kind: carouselCommandNext, jump: jump})
}
func (a *CarouselController) ScrollTo(index int, jump bool) {
	a.enqueue(carouselCommand{kind: carouselCommandTo, index: index, jump: jump})
}

func (a *CarouselController) On(event CarouselEventType, listener *CarouselListener) *CarouselController {
	if a == nil || listener == nil || listener.callback == nil {
		return a
	}
	a.mu.Lock()
	if a.listeners == nil {
		a.listeners = make(map[CarouselEventType][]*CarouselListener)
	}
	a.listeners[event] = append(slices.Clone(a.listeners[event]), listener)
	a.mu.Unlock()
	return a
}

func (a *CarouselController) Off(event CarouselEventType, listener *CarouselListener) *CarouselController {
	if a == nil {
		return a
	}
	a.mu.Lock()
	kept := slices.DeleteFunc(slices.Clone(a.listeners[event]), func(entry *CarouselListener) bool { return entry == listener })
	a.listeners[event] = kept
	a.mu.Unlock()
	return a
}

// Emit only notifies listeners; it does not synthesize an engine action. A
// listener-list snapshot matches source On/Off mutation during the current emit.
// Explicit registration and Emit remain usable after engine Destroy.
func (a *CarouselController) Emit(event CarouselEventType) *CarouselController {
	if a == nil {
		return a
	}
	a.mu.Lock()
	listeners := slices.Clone(a.listeners[event])
	a.mu.Unlock()
	for _, listener := range listeners {
		listener.callback(a, event)
	}
	return a
}

// Destroy ends the engine binding and clears listeners after the destroy event,
// including listeners added by a destroy callback. It never revives that engine.
func (a *CarouselController) Destroy() {
	if a == nil {
		return
	}
	a.mu.Lock()
	if a.snapshot.Destroyed {
		a.mu.Unlock()
		return
	}
	a.snapshot.Destroyed, a.snapshot.Ready = true, false
	a.commands = nil
	services := a.services
	a.services, a.bound = ui.Services{}, false
	a.mu.Unlock()
	a.Emit(CarouselDestroy)
	a.mu.Lock()
	a.listeners = nil
	a.mu.Unlock()
	services.Invalidate()
}

func (a *CarouselController) Snapshot() CarouselControllerSnapshot {
	if a == nil {
		return CarouselControllerSnapshot{}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return carouselCloneControllerSnapshot(a.snapshot)
}

func carouselCloneControllerSnapshot(s CarouselControllerSnapshot) CarouselControllerSnapshot {
	s.SnapList, s.SnapOffsets = slices.Clone(s.SnapList), slices.Clone(s.SnapOffsets)
	s.InView, s.NotInView = slices.Clone(s.InView), slices.Clone(s.NotInView)
	s.Slides = slices.Clone(s.Slides)
	s.Engine.SlideRegistry = slices.Clone(s.Engine.SlideRegistry)
	for i := range s.Engine.SlideRegistry {
		s.Engine.SlideRegistry[i] = slices.Clone(s.Engine.SlideRegistry[i])
	}
	s.Engine.Options.InViewThreshold = slices.Clone(s.Engine.Options.InViewThreshold)
	s.Engine.Options.Slides = slices.Clone(s.Engine.Options.Slides)
	return s
}

func (a *CarouselController) SelectedScrollSnap() int   { return a.Snapshot().Selected }
func (a *CarouselController) PreviousScrollSnap() int   { return a.Snapshot().Previous }
func (a *CarouselController) ScrollSnapList() []float64 { return a.Snapshot().SnapList }
func (a *CarouselController) ScrollProgress() float64   { return a.Snapshot().Progress }
func (a *CarouselController) SlidesInView() []int       { return a.Snapshot().InView }
func (a *CarouselController) SlidesNotInView() []int    { return a.Snapshot().NotInView }
func (a *CarouselController) CanScrollPrev() bool       { return a.Snapshot().CanScrollPrevious }
func (a *CarouselController) CanScrollNext() bool       { return a.Snapshot().CanScrollNext }

// RootHandle is Embla's viewport root, not the surrounding labelled region.
func (a *CarouselController) RootHandle() ui.Handle                  { return a.Snapshot().Root }
func (a *CarouselController) ContainerHandle() ui.Handle             { return a.Snapshot().Container }
func (a *CarouselController) RegionHandle() ui.Handle                { return a.Snapshot().Region }
func (a *CarouselController) SlideHandles() []ui.Handle              { return a.Snapshot().Slides }
func (a *CarouselController) EngineSnapshot() CarouselEngineSnapshot { return a.Snapshot().Engine }

func (a *CarouselController) publish(snapshot CarouselControllerSnapshot, event CarouselEventType) {
	a.mu.Lock()
	if a.snapshot.Destroyed {
		a.mu.Unlock()
		return
	}
	snapshot.Generation = a.generation
	a.snapshot = carouselCloneControllerSnapshot(snapshot)
	a.mu.Unlock()
	if event != "" {
		a.Emit(event)
	}
}

func (a *CarouselController) recordError(err error) {
	a.mu.Lock()
	a.snapshot.Error = err
	a.mu.Unlock()
}

func (a *CarouselController) attach(owner ui.Handle, services ui.Services, transfer bool) (uint64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.snapshot.Destroyed {
		return 0, &CarouselOptionError{Field: "Controller", Code: CarouselDestroyed}
	}
	if a.bound && a.owner != owner {
		if !transfer {
			return 0, &CarouselOptionError{Field: "Controller", Code: CarouselAlreadyBound}
		}
		a.generation++
		a.owner = owner
		a.snapshot.Ready = false
	}
	if !a.bound {
		a.generation++
		a.owner, a.bound = owner, true
	}
	a.services = services
	return a.generation, nil
}

func (a *CarouselController) ownerState() (ui.Handle, ui.Services, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.owner, a.services, a.bound
}

func (a *CarouselController) takeCommands() []carouselCommand {
	a.mu.Lock()
	defer a.mu.Unlock()
	commands := a.commands
	a.commands = nil
	return commands
}

func (a *CarouselController) baseOptions() CarouselOptions {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.appliedSet {
		return carouselCloneOptions(a.appliedOptions)
	}
	return carouselCloneOptions(a.options)
}

func (a *CarouselController) acceptOptions(options CarouselOptions) {
	a.mu.Lock()
	a.appliedOptions = carouselCloneOptions(options)
	a.appliedSet = true
	a.mu.Unlock()
}

type carouselControllerLease struct {
	controller *CarouselController
	generation uint64
	closed     bool
}

func (lease *carouselControllerLease) Close() error {
	if lease == nil || lease.closed || lease.controller == nil {
		return nil
	}
	lease.closed = true
	a := lease.controller
	a.mu.Lock()
	current := a.bound && a.generation == lease.generation
	a.mu.Unlock()
	if current {
		a.Destroy()
	}
	return nil
}

type carouselControllerMotion struct {
	from, to float64
	at       time.Time
	duration time.Duration
	active   bool
}
