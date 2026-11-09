package display

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/carouselautoplay"
)

type carouselPluginBinding struct {
	plugin           CarouselPlugin
	definition       *carouselPluginDefinition
	api              *CarouselController
	generation       uint64
	initialized      bool
	listener         *CarouselListener
	autoplay         *carouselAutoplaySession
	root             ui.Handle
	deadline         time.Time
	hovered, focused bool
}

func (a *CarouselController) basePlugins() []CarouselPlugin {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.appliedPluginsSet {
		return slices.Clone(a.appliedPlugins)
	}
	return slices.Clone(a.desiredPlugins)
}
func carouselSamePlugins(a, b []CarouselPlugin) bool {
	return slices.EqualFunc(a, b, func(x, y CarouselPlugin) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.carouselPluginDefinition() == y.carouselPluginDefinition()
	})
}
func carouselPluginMediaChanged(plugins []CarouselPlugin, a, b CarouselEnvironment) bool {
	for _, p := range plugins {
		def := p.carouselPluginDefinition()
		if def.autoplay != nil {
			for _, r := range def.autoplay.options.Breakpoints {
				if r.When.matches(a) != r.When.matches(b) {
					return true
				}
			}
		} else {
			for _, r := range def.options.Breakpoints {
				if r.When.matches(a) != r.When.matches(b) {
					return true
				}
			}
		}
	}
	return false
}
func (frame *carouselControllerBuild) pluginFrame(binding *carouselPluginBinding, invoke func(CarouselPluginFrame)) {
	scope := &carouselPluginActionScope{frame: frame, binding: binding, alive: true}
	defer func() { scope.alive = false; scope.frame = nil; scope.binding = nil }()
	invoke(CarouselPluginFrame{Now: frame.now, Environment: frame.contextEnvironment, Snapshot: frame.api.Snapshot(), Root: binding.root, Hovered: binding.hovered, FocusWithin: binding.focused, Actions: CarouselPluginActions{scope: scope}})
}
func (binding *carouselPluginBinding) event(event CarouselEventType) {
	binding.api.mu.Lock()
	frame := binding.api.pluginFrame
	now := binding.api.lastNow
	binding.api.mu.Unlock()
	if binding.autoplay != nil {
		s := binding.autoplay
		s.syncSelected()
		emit := s.emit(frame)
		switch event {
		case CarouselPointerDown:
			s.core.PointerDown(emit)
		case CarouselPointerUp:
			s.core.PointerUp(float64(now.UnixMilli()), emit)
		case CarouselSlideFocusStart:
			s.core.FocusStart(emit)
		}
		s.publish()
		binding.api.services.Invalidate()
	} else if binding.definition.hooks.Event != nil {
		if frame != nil {
			frame.pluginFrame(binding, func(f CarouselPluginFrame) { binding.definition.hooks.Event(f, event) })
		} else {
			binding.definition.hooks.Event(CarouselPluginFrame{Now: now, Snapshot: binding.api.Snapshot(), Root: binding.root}, event)
		}
	}
}
func (binding *carouselPluginBinding) destroy(frame *carouselControllerBuild) {
	if !binding.initialized {
		return
	}
	hooks := binding.definition.hooks
	if binding.autoplay != nil {
		for _, event := range []CarouselEventType{CarouselPointerDown, CarouselPointerUp, CarouselSlideFocusStart} {
			binding.api.Off(event, binding.listener)
		}
		binding.autoplay.syncSelected()
		binding.autoplay.core.Destroy(binding.autoplay.emit(frame))
		auto := binding.autoplay.plugin
		auto.definition.mu.Lock()
		if auto.session == binding.autoplay {
			auto.carry = binding.autoplay.core.Carry()
		}
		auto.definition.mu.Unlock()
		binding.autoplay.publish()
	} else {
		for _, event := range hooks.Events {
			binding.api.Off(event, binding.listener)
		}
		if hooks.Destroy != nil {
			if frame != nil {
				frame.pluginFrame(binding, hooks.Destroy)
			} else {
				hooks.Destroy(CarouselPluginFrame{Now: binding.api.lastNow, Snapshot: binding.api.Snapshot(), Root: binding.root})
			}
		}
	}
	binding.deadline = time.Time{}
	def := binding.definition
	def.mu.Lock()
	if def.owner == binding.api && def.generation == binding.generation {
		def.owner = nil
	}
	def.mu.Unlock()
}
func (a *CarouselController) destroyPlugins(frame *carouselControllerBuild) {
	a.mu.Lock()
	bindings := a.pluginBindings
	a.pluginBindings = nil
	a.mu.Unlock()
	for _, b := range bindings {
		b.destroy(frame)
	}
}
func (frame *carouselControllerBuild) validatePlugins(c *ui.Context, plugins []CarouselPlugin, count int, engineActive bool) error {
	if err := carouselValidatePlugins(plugins); err != nil {
		return err
	}
	if !engineActive {
		return nil
	}
	for _, p := range plugins {
		d := p.carouselPluginDefinition()
		d.mu.Lock()
		owner := d.owner
		d.mu.Unlock()
		if owner != nil && owner != frame.api {
			return &CarouselOptionError{Field: "Plugins.Owner", Code: CarouselAlreadyBound}
		}
		if d.autoplay != nil {
			o := carouselResolveAutoplayOptions(d.autoplay.options, frame.contextEnvironment)
			if count > 1 && *o.Active {
				root := frame.engine.root
				if o.Root != nil && *o.Root != (ui.Handle{}) {
					root = *o.Root
				}
				if o.RootResolver != nil {
					if resolved := o.RootResolver.resolve(frame.engine.root); resolved != (ui.Handle{}) {
						root = resolved
					}
				}
				if !c.Resolve(root).Valid() {
					return &CarouselOptionError{Field: "Autoplay.Root", Code: CarouselUnresolvedNode}
				}
				if frame.pluginRoots == nil {
					frame.pluginRoots = map[*carouselPluginDefinition]ui.Handle{}
				}
				frame.pluginRoots[d] = root
			}
		}
	}
	return nil
}
func (frame *carouselControllerBuild) initPlugins(c *ui.Context) {
	e := frame.engine
	api := frame.api
	e.pluginBindings = nil
	api.mu.Lock()
	api.appliedPlugins = slices.Clone(e.plugins)
	api.appliedPluginsSet = true
	api.mu.Unlock()
	if !e.resolved.Active {
		return
	}
	shared := map[*carouselPluginDefinition]*carouselPluginBinding{}
	bindings := make([]*carouselPluginBinding, 0, len(e.plugins))
	pluginMap := map[string]CarouselPlugin{}
	generation := api.Snapshot().Generation
	for _, p := range e.plugins {
		d := p.carouselPluginDefinition()
		pluginMap[p.Name()] = p
		active := carouselPluginActive(d.options, frame.contextEnvironment)
		var options CarouselAutoplayOptions
		if d.autoplay != nil {
			options = carouselResolveAutoplayOptions(d.autoplay.options, frame.contextEnvironment)
			active = *options.Active
		}
		if !active {
			continue
		}
		binding := shared[d]
		if binding == nil {
			binding = &carouselPluginBinding{plugin: p, definition: d, api: api, generation: generation, root: e.root}
			shared[d] = binding
			binding.listener = NewCarouselListener(func(_ *CarouselController, event CarouselEventType) { binding.event(event) })
			if d.autoplay != nil {
				binding.autoplay = &carouselAutoplaySession{plugin: d.autoplay, binding: binding, options: options, hidden: e.hidden}
				if root, ok := frame.pluginRoots[d]; ok {
					binding.root = root
				}
				auto := binding.autoplay.plugin
				auto.definition.mu.Lock()
				carry := auto.carry
				auto.definition.mu.Unlock()
				binding.autoplay.core.RestoreCarry(carry)
				// New listeners establish a baseline; remounting alone does not
				// synthesize a source mouseenter/focusout delivery.
				if root := c.Resolve(binding.root); root.Valid() {
					binding.autoplay.hovered = root.Hovered()
				}
				binding.autoplay.focused = frame.track.Valid() && frame.track.FocusWithin()
				for _, item := range e.selectedItems {
					if item.canonical.FocusWithin(c) {
						binding.autoplay.focusSlide = item.canonical
						break
					}
				}
			}
		}
		bindings = append(bindings, binding)
		e.pluginBindings = bindings
		api.mu.Lock()
		api.pluginBindings = bindings
		api.mu.Unlock()
		binding.initialized = true
		d.mu.Lock()
		d.owner, d.generation = api, generation
		d.mu.Unlock()
		if binding.autoplay != nil {
			s := binding.autoplay
			var delays []time.Duration
			if len(e.geometry.snaps) <= 1 {
			} else if options.DelayResolver != nil {
				snap := api.Snapshot()
				delays = options.DelayResolver.resolve(CarouselAutoplayDelayInput{Snaps: slices.Clone(snap.SnapList), Snapshot: snap})
			} else {
				for range e.geometry.snaps {
					delays = append(delays, *options.Delay)
				}
			}
			if len(e.geometry.snaps) > 1 && len(delays) != len(e.geometry.snaps) {
				e.lastError = &CarouselOptionError{Field: "Autoplay.DelayResolver", Code: CarouselInvalidOption}
				continue
			}
			s.delays = nil
			valid := true
			for _, delay := range delays {
				if delay < 0 {
					valid = false
				}
				s.delays = append(s.delays, float64(delay)/float64(time.Millisecond))
			}
			if !valid {
				e.lastError = &CarouselOptionError{Field: "Autoplay.DelayResolver", Code: CarouselInvalidOption}
				continue
			}
			s.core.SetSelected(e.selected)
			s.plugin.definition.mu.Lock()
			s.plugin.session, s.plugin.controller = s, api
			s.plugin.definition.mu.Unlock()
			if len(e.geometry.snaps) > 1 {
				if e.resolved.WatchDrag {
					api.On(CarouselPointerDown, binding.listener)
					if !*options.StopOnInteraction {
						api.On(CarouselPointerUp, binding.listener)
					}
				}
				if *options.StopOnFocusIn {
					api.On(CarouselSlideFocusStart, binding.listener)
				}
			}
			s.core.Init(carouselautoplay.Options{Delays: s.delays, Jump: *options.Jump, PlayOnInit: *options.PlayOnInit, StopOnFocusIn: *options.StopOnFocusIn, StopOnInteraction: *options.StopOnInteraction, StopOnMouseEnter: *options.StopOnMouseEnter, StopOnLastSnap: *options.StopOnLastSnap, WatchDrag: e.resolved.WatchDrag}, len(e.geometry.snaps), e.geometry.loop, e.hidden, float64(frame.now.UnixMilli()), s.emit(frame))
			s.publish()
		} else {
			for _, event := range d.hooks.Events {
				api.On(event, binding.listener)
			}
			if d.hooks.Init != nil {
				frame.pluginFrame(binding, d.hooks.Init)
			}
		}
	}
	e.pluginBindings = bindings
	api.mu.Lock()
	api.pluginBindings = bindings
	api.pluginMap = pluginMap
	api.mu.Unlock()
}
func (frame *carouselControllerBuild) updatePlugins(c *ui.Context, b *carouselCompositionBuild, beforeFocus bool) {
	if frame.engine.pluginMediaPending {
		return
	}
	seen := map[*carouselPluginBinding]bool{}
	for _, binding := range frame.engine.pluginBindings {
		if seen[binding] {
			continue
		}
		seen[binding] = true
		root := c.Resolve(binding.root)
		binding.hovered = root.Valid() && root.Hovered()
		binding.focused = frame.track.Valid() && frame.track.FocusWithin()
		if s := binding.autoplay; s != nil {
			s.syncSelected()
			emit := s.emit(frame)
			now := float64(frame.now.UnixMilli())
			if beforeFocus {
				if s.hidden != frame.engine.hidden {
					s.core.Visibility(frame.engine.hidden, now, emit)
					s.hidden = frame.engine.hidden
				}
				if binding.hovered != s.hovered {
					if binding.hovered {
						s.core.MouseEnter(emit)
					} else {
						s.core.MouseLeave(now, emit)
					}
					s.hovered = binding.hovered
				}
				var focus ui.Handle
				for _, item := range frame.engine.selectedItems {
					if item.canonical.FocusWithin(c) {
						focus = item.canonical
						break
					}
				}
				if s.focused && (!binding.focused || focus != s.focusSlide || frame.engine.pendingTab) {
					s.core.FocusOut(now, emit)
				}
				s.focused, s.focusSlide = binding.focused, focus
			} else if !frame.engine.hidden && !s.deadline.IsZero() && !frame.now.Before(s.deadline) {
				s.core.Advance(now, emit)
			}
			s.publish()
		} else if !beforeFocus && binding.definition.hooks.Frame != nil {
			if !binding.deadline.IsZero() && !frame.now.Before(binding.deadline) {
				binding.deadline = time.Time{}
			}
			frame.pluginFrame(binding, binding.definition.hooks.Frame)
		}
		if !beforeFocus && !frame.engine.hidden {
			deadline := binding.deadline
			if s := binding.autoplay; s != nil {
				deadline = s.deadline
			}
			if !deadline.IsZero() {
				c.After(max(0, deadline.Sub(frame.now)))
			}
		}
	}
}
func (frame *carouselControllerBuild) autoplayCommand(command carouselCommand) bool {
	for _, binding := range frame.engine.pluginBindings {
		if binding.definition != command.plugin || binding.autoplay == nil {
			continue
		}
		s := binding.autoplay
		s.syncSelected()
		emit := s.emit(frame)
		now := float64(frame.now.UnixMilli())
		switch command.pluginAction {
		case "play":
			s.core.Play(now, command.jumpOverride, emit)
		case "stop":
			s.core.Stop(emit)
		case "reset":
			s.core.Reset(now, emit)
		}
		s.publish()
		return true
	}
	return false
}
