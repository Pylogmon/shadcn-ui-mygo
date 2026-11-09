package display

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/carouselautoplay"
)

const (
	CarouselAutoplayPlay         CarouselEventType = "autoplay:play"
	CarouselAutoplayStop         CarouselEventType = "autoplay:stop"
	CarouselAutoplaySelect       CarouselEventType = "autoplay:select"
	CarouselAutoplayTimerSet     CarouselEventType = "autoplay:timerset"
	CarouselAutoplayTimerStopped CarouselEventType = "autoplay:timerstopped"
)

type CarouselAutoplayDelayInput struct {
	Snaps    []float64
	Snapshot CarouselControllerSnapshot
}

// CarouselAutoplayDelayResolver is explicitly application-owned and pure. Its
// values may be retained; it must not capture a Context, Element or frame facade.
type CarouselAutoplayDelayResolver struct {
	resolve func(CarouselAutoplayDelayInput) []time.Duration
}

func NewCarouselAutoplayDelayResolver(resolve func(CarouselAutoplayDelayInput) []time.Duration) *CarouselAutoplayDelayResolver {
	return &CarouselAutoplayDelayResolver{resolve: resolve}
}

// CarouselAutoplayRootResolver is an application-owned pure native rootNode
// resolver. A zero result uses the viewport. No UI frame objects may be captured.
type CarouselAutoplayRootResolver struct{ resolve func(ui.Handle) ui.Handle }

func NewCarouselAutoplayRootResolver(resolve func(ui.Handle) ui.Handle) *CarouselAutoplayRootResolver {
	return &CarouselAutoplayRootResolver{resolve: resolve}
}

type CarouselAutoplayOptions struct {
	Active                                                                               *bool
	Delay                                                                                *time.Duration
	DelayResolver                                                                        *CarouselAutoplayDelayResolver
	Jump, PlayOnInit, StopOnFocusIn, StopOnInteraction, StopOnMouseEnter, StopOnLastSnap *bool
	RootResolver                                                                         *CarouselAutoplayRootResolver
	Root                                                                                 *ui.Handle
	Breakpoints                                                                          []CarouselAutoplayBreakpoint
}
type CarouselAutoplayBreakpoint struct {
	Key     string
	When    CarouselMedia
	Options CarouselAutoplayOptions
}

func (o CarouselAutoplayOptions) Validate() error {
	if o.Delay != nil && *o.Delay < 0 {
		return &CarouselOptionError{Field: "Autoplay.Delay", Code: CarouselInvalidOption}
	}
	if o.RootResolver != nil && o.RootResolver.resolve == nil {
		return &CarouselOptionError{Field: "Autoplay.RootResolver", Code: CarouselInvalidOption}
	}
	if o.DelayResolver != nil && o.DelayResolver.resolve == nil {
		return &CarouselOptionError{Field: "Autoplay.DelayResolver", Code: CarouselInvalidOption}
	}
	keys := map[string]bool{}
	for _, r := range o.Breakpoints {
		if r.Key == "" || keys[r.Key] || len(r.Options.Breakpoints) > 0 {
			return &CarouselOptionError{Field: "Autoplay.Breakpoints", Code: CarouselInvalidOption}
		}
		keys[r.Key] = true
		if err := (CarouselOptions{Breakpoints: []CarouselBreakpoint{{Key: r.Key, When: r.When}}}).Validate(); err != nil {
			return err
		}
		if err := r.Options.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func carouselCloneAutoplayOptions(o CarouselAutoplayOptions) CarouselAutoplayOptions {
	o.Active = carouselCopyValue(o.Active)
	o.Delay = carouselCopyValue(o.Delay)
	o.Root = carouselCopyValue(o.Root)
	o.Jump, o.PlayOnInit = carouselCopyValue(o.Jump), carouselCopyValue(o.PlayOnInit)
	o.StopOnFocusIn, o.StopOnInteraction = carouselCopyValue(o.StopOnFocusIn), carouselCopyValue(o.StopOnInteraction)
	o.StopOnMouseEnter, o.StopOnLastSnap = carouselCopyValue(o.StopOnMouseEnter), carouselCopyValue(o.StopOnLastSnap)
	o.Breakpoints = slices.Clone(o.Breakpoints)
	for i := range o.Breakpoints {
		r := &o.Breakpoints[i]
		r.When = carouselCloneMedia(r.When)
		r.Options = carouselCloneAutoplayOptions(r.Options)
	}
	return o
}
func carouselResolveAutoplayOptions(o CarouselAutoplayOptions, env CarouselEnvironment) CarouselAutoplayOptions {
	r := CarouselAutoplayOptions{Active: CarouselValue(true), Delay: CarouselValue(4 * time.Second), Jump: CarouselValue(false), PlayOnInit: CarouselValue(true), StopOnFocusIn: CarouselValue(true), StopOnInteraction: CarouselValue(true), StopOnMouseEnter: CarouselValue(false), StopOnLastSnap: CarouselValue(false)}
	apply := func(p CarouselAutoplayOptions) {
		if p.Active != nil {
			r.Active = p.Active
		}
		if p.Delay != nil {
			r.Delay = p.Delay
			r.DelayResolver = nil
		}
		if p.DelayResolver != nil {
			r.DelayResolver = p.DelayResolver
		}
		if p.Jump != nil {
			r.Jump = p.Jump
		}
		if p.PlayOnInit != nil {
			r.PlayOnInit = p.PlayOnInit
		}
		if p.StopOnFocusIn != nil {
			r.StopOnFocusIn = p.StopOnFocusIn
		}
		if p.StopOnInteraction != nil {
			r.StopOnInteraction = p.StopOnInteraction
		}
		if p.StopOnMouseEnter != nil {
			r.StopOnMouseEnter = p.StopOnMouseEnter
		}
		if p.StopOnLastSnap != nil {
			r.StopOnLastSnap = p.StopOnLastSnap
		}
		if p.Root != nil {
			r.Root = p.Root
			r.RootResolver = nil
		}
		if p.RootResolver != nil {
			r.RootResolver = p.RootResolver
			r.Root = nil
		}
	}
	apply(o)
	for _, b := range o.Breakpoints {
		if b.When.matches(env) {
			apply(b.Options)
		}
	}
	return r
}

type CarouselAutoplay struct {
	definition carouselPluginDefinition
	options    CarouselAutoplayOptions
	session    *carouselAutoplaySession
	playing    bool
	started    float64
	delays     []float64
	controller *CarouselController
	carry      carouselautoplay.Carry
}

func NewCarouselAutoplay(options CarouselAutoplayOptions) (*CarouselAutoplay, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	a := &CarouselAutoplay{options: carouselCloneAutoplayOptions(options)}
	a.definition.name = "autoplay"
	a.definition.autoplay = a
	return a, nil
}
func (a *CarouselAutoplay) Name() string { return "autoplay" }
func (a *CarouselAutoplay) carouselPluginDefinition() *carouselPluginDefinition {
	if a == nil {
		return nil
	}
	return &a.definition
}

// Play/Stop/Reset enqueue in their owner's command FIFO. Calling them before
// binding has no effect. Plugin-frame actions use the synchronous state machine.
func (a *CarouselAutoplay) Play(jump ...bool) {
	var override *bool
	if len(jump) > 0 {
		override = CarouselValue(jump[0])
	}
	a.command("play", override)
}
func (a *CarouselAutoplay) Stop()  { a.command("stop", nil) }
func (a *CarouselAutoplay) Reset() { a.command("reset", nil) }
func (a *CarouselAutoplay) command(kind string, jump *bool) {
	if a == nil {
		return
	}
	a.definition.mu.Lock()
	owner := a.definition.owner
	a.definition.mu.Unlock()
	if owner != nil {
		owner.enqueue(carouselCommand{kind: carouselCommandAutoplay, plugin: &a.definition, pluginAction: kind, jumpOverride: jump})
	}
}
func (a *CarouselAutoplay) IsPlaying() bool {
	if a == nil {
		return false
	}
	a.definition.mu.Lock()
	defer a.definition.mu.Unlock()
	return a.playing
}

// TimeUntilNext uses the current frame/test clock supplied by the caller. Like
// the source, it uses the currently selected snap's delay, which can differ from
// the already scheduled deadline after a manual selection. False matches source null.
func (a *CarouselAutoplay) TimeUntilNext(now time.Time) (time.Duration, bool) {
	if a == nil {
		return 0, false
	}
	a.definition.mu.Lock()
	started, delays, owner := a.started, slices.Clone(a.delays), a.controller
	a.definition.mu.Unlock()
	if started == 0 || owner == nil {
		return 0, false
	}
	index := owner.SelectedScrollSnap()
	if index < 0 || index >= len(delays) {
		return 0, false
	}
	return time.Duration((delays[index] - float64(now.UnixMilli()) + started) * float64(time.Millisecond)), true
}

type carouselAutoplaySession struct {
	plugin                   *CarouselAutoplay
	core                     carouselautoplay.Core
	options                  CarouselAutoplayOptions
	binding                  *carouselPluginBinding
	delays                   []float64
	hidden, hovered, focused bool
	focusSlide               ui.Handle
	deadline                 time.Time
}

func (s *carouselAutoplaySession) publish() {
	a := s.plugin
	a.definition.mu.Lock()
	defer a.definition.mu.Unlock()
	if a.session != s {
		return
	}
	a.playing = s.core.IsPlaying()
	a.delays = slices.Clone(s.delays)
	a.started = s.core.Started()
}
func (s *carouselAutoplaySession) emit(frame *carouselControllerBuild) func(carouselautoplay.Effect) {
	return func(effect carouselautoplay.Effect) {
		s.publish()
		if effect.Event == string(CarouselAutoplayTimerSet) {
			now := s.binding.api.lastNow
			if frame != nil {
				now = frame.now
			}
			index := s.core.Selected()
			if index >= 0 && index < len(s.delays) {
				// Native After has a 1ms minimum. Keep even a zero source delay
				// asynchronous, rather than advancing during the init/reset build.
				s.deadline = now.Add(max(time.Millisecond, time.Duration(s.delays[index]*float64(time.Millisecond))))
			}
		} else if effect.Event == string(CarouselAutoplayTimerStopped) {
			s.deadline = time.Time{}
		}
		switch effect.Kind {
		case "scrollNext":
			if frame != nil {
				frame.scroll(frame.engine.selected+1, -1, effect.Jump, frame.now)
			}
		case "scrollTo":
			if frame != nil {
				frame.scroll(effect.Index, 0, effect.Jump, frame.now)
			}
		case "event":
			s.binding.api.Emit(CarouselEventType(effect.Event))
		}
		s.publish()
	}
}
func (s *carouselAutoplaySession) syncSelected() {
	s.core.SetSelected(s.binding.api.SelectedScrollSnap())
}
