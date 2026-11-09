package display

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/carouselphysics"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

func carouselControllerContainer(c *ui.Context, selected *int, p CarouselContainerProps, children func(*ui.Context)) ui.Element {
	if selected == nil {
		panic("display: CarouselContainer requires selected snap state")
	}
	p.Count = max(0, p.Count)
	label := p.Label
	if label == "" {
		label = "Carousel"
	}
	root := ui.Box(c).FillWidth().MinWidth(0).Role(ui.RoleRegion).RoleDescription("carousel").Label(label).Focusable().Disabled(p.Disabled)
	if p.Key != "" {
		root.Key(carouselContainerKey(p.Key))
	}
	state := ui.Local(root, "shadcn.carousel.composition", func() carouselCompositionState { return carouselCompositionState{spacing: 16} })
	if len(state.bases) != p.Count {
		next := make([]CarouselBasis, p.Count)
		copy(next, state.bases)
		state.bases = next
	}
	e := ui.Local(root, "shadcn.carousel.controller.engine", func() carouselControllerEngine { return carouselControllerEngine{} })
	root.Bind(&e.region)
	api := p.Controller
	if api == nil {
		owned := ui.Local(root, "shadcn.carousel.controller.owned", func() *CarouselController { a, _ := NewCarouselController(CarouselOptions{}); return a })
		api = *owned
	}
	lease := ui.Local(root, "shadcn.carousel.controller.lease", func() carouselControllerLease { return carouselControllerLease{} })
	if lease.controller != api {
		lease.Close()
		*lease = carouselControllerLease{controller: api}
		region := e.region
		*e = carouselControllerEngine{region: region}
	}
	previousOwner, previousServices, previouslyBound := api.ownerState()
	transfer := previouslyBound && previousOwner != e.region && previousServices == c.Services() && !c.Resolve(previousOwner).Valid()
	generation, attachError := api.attach(e.region, c.Services(), transfer)
	if attachError == nil {
		lease.generation = generation
	}
	width, height := c.Size()
	direction := CarouselLTR
	if c.Direction() == ui.RightToLeft {
		direction = CarouselRTL
	}
	axis := CarouselHorizontal
	if p.Vertical {
		axis = CarouselVertical
	}
	env := CarouselEnvironment{Width: width, Height: height, Direction: direction, Reduced: motion.Reduced(c), Dark: c.Theme() != nil && c.Theme().Dark}
	if !e.initialized {
		e.options = carouselMergeOptions(CarouselOptions{Axis: &axis, Direction: &direction, Align: &CarouselAlignment{Mode: p.Align}, Contain: &p.Contain, Loop: &p.Loop}, api.baseOptions())
		e.lastControlled = *selected
		e.plugins = api.basePlugins()
		e.initialized = true
	}
	frame := &carouselControllerBuild{api: api, engine: e, lease: lease, hooks: p.Hooks, options: p.Options, root: root, contextEnvironment: env, optionsError: attachError, onSelect: p.OnSelect, now: c.Now()}
	api.mu.Lock()
	previousFrame := api.pluginFrame
	if attachError == nil {
		api.pluginFrame = frame
		api.lastNow = frame.now
	}
	api.mu.Unlock()
	defer func() { api.mu.Lock(); api.pluginFrame = previousFrame; api.mu.Unlock() }()
	frame.previousOptions = e.options
	if p.Options != nil {
		if err := p.Options.Validate(); err != nil {
			frame.optionsError = err
			frame.invalidOptions = true
		} else if !e.propsSet || !carouselOptionsSame(e.lastProps, *p.Options) {
			e.options = carouselMergeOptions(e.options, *p.Options)
			frame.forceReInit = true
			frame.startIndex = p.Options.StartIndex
		}
	}
	if p.Plugins != nil && (!e.pluginPropsSet || !carouselSamePlugins(e.lastPluginProps, p.Plugins)) {
		if err := carouselValidatePlugins(p.Plugins); err != nil {
			frame.optionsError = err
			frame.invalidOptions = true
		} else {
			e.plugins = slices.Clone(p.Plugins)
			e.lastPluginProps = slices.Clone(p.Plugins)
			e.pluginPropsSet = true
			frame.forceReInit = true
		}
	}
	resolved := carouselResolveOptions(e.options, env)
	if e.ready && (resolved.Axis != e.resolved.Axis || resolved.Direction != e.resolved.Direction || !carouselMediaEnvironmentSame(e.lastEnvironment, env, e.options.Breakpoints)) {
		frame.forceReInit = true
	}
	if e.pluginMediaPending {
		frame.forceReInit = true
		e.pluginMediaPending = false
	}
	if e.ready && carouselPluginMediaChanged(e.plugins, e.lastEnvironment, env) {
		if env.Width != e.lastEnvironment.Width || env.Height != e.lastEnvironment.Height {
			// Resolve a size-driven plugin media change after the viewport has had one
			// real layout. This also works with watchResize=false and avoids Init at an
			// obsolete viewport size followed by a second resize-driven Init.
			e.pluginMediaPending = true
			c.Invalidate()
		} else {
			frame.forceReInit = true
		}
	}
	e.lastEnvironment = env
	p.Vertical = resolved.Axis == CarouselVertical
	state.vertical, state.rtl = p.Vertical, resolved.Direction == CarouselRTL
	if e.viewportBox.W > 0 {
		state.extent = e.viewportBox.W
		if p.Vertical {
			state.extent = e.viewportBox.H
		}
	}
	root.Direction(ui.LeftToRight)
	if state.rtl {
		root.Direction(ui.RightToLeft)
	}
	frame.resolvedDirection(root, resolved.Direction)
	if api.Snapshot().Destroyed {
		e.scheduler.Stop()
		e.pointerEvents = nil
		e.pendingTab = false
	}
	// Always subscribe to visibility, including instant/reduced-motion builds.
	hidden := c.WindowHidden()
	if hidden != e.hidden {
		e.hidden = hidden
		e.scheduler.Reset()
		e.lastTick = time.Time{}
	}
	// Sample steady animation before composing independent siblings. Pending
	// input/commands run first in finalize so a press/retarget cannot see an extra
	// seek at the old target merely because the owner happened to rebuild.
	if e.ready && frame.optionsError == nil && !hidden && len(e.pointerEvents) == 0 && !api.hasCommands() {
		if frame.tick(c.Now(), motion.Reduced(c) || p.Duration < 0) && !(motion.Reduced(c) || p.Duration < 0) {
			c.AnimationFrame()
		}
	}
	layout := carouselMeasure(state.extent, state.spacing, state.bases, CarouselAlignStart, CarouselTrimSnaps, false)
	state.geometry = layout
	b := &carouselCompositionBuild{root: root, state: state, p: p, selected: selected, seen: make([]bool, p.Count), usedKeys: make(map[any]bool), geometry: layout, controller: frame}
	if e.ready && e.resolved.Active && !api.Snapshot().Destroyed {
		b.position = float32(e.offset)
	}
	b.choose = func(target int) {
		if p.Disabled || attachError != nil {
			return
		}
		delta := target - *selected
		if delta == 1 {
			api.ScrollNext(false)
		} else if delta == -1 {
			api.ScrollPrev(false)
		} else {
			api.ScrollTo(target, false)
		}
		c.Invalidate()
	}
	b.keys = func(element ui.Element) {
		if !element.Focused() || p.Disabled || attachError != nil {
			return
		}
		previous, next := ui.KeyLeft, ui.KeyRight
		if state.rtl {
			previous, next = next, previous
		}
		if p.Vertical {
			previous, next = ui.KeyUp, ui.KeyDown
		}
		switch {
		case element.Shortcut(0, previous):
			b.choose(*selected - 1)
		case element.Shortcut(0, next):
			b.choose(*selected + 1)
		case element.Shortcut(0, ui.KeyHome):
			api.ScrollTo(0, false)
		case element.Shortcut(0, ui.KeyEnd):
			api.ScrollTo(len(e.geometry.snaps)-1, false)
		}
	}
	b.keys(root)
	root.ObserveInput(func(event ui.InputEvent) {
		if event.Kind == ui.InputKeyDown && event.Key == ui.KeyTab {
			e.lastTab = event.Time
			e.pendingTab = true
		}
	})
	root.Children(func() {
		scope := ui.Local(c.Root(), carouselCompositionKey{}, func() carouselCompositionScope { return carouselCompositionScope{} })
		previous := *scope
		scope.build = b
		defer func() { *scope = previous }()
		if children != nil {
			children(c)
		}
		frame.finalize(c, b)
	})
	return root
}

func (frame *carouselControllerBuild) resolvedDirection(root ui.Element, direction CarouselDirection) {
	if direction == CarouselRTL {
		root.Direction(ui.RightToLeft)
	} else {
		root.Direction(ui.LeftToRight)
	}
}
func carouselMediaEnvironmentSame(a, b CarouselEnvironment, rules []CarouselBreakpoint) bool {
	for _, rule := range rules {
		if rule.When.matches(a) != rule.When.matches(b) {
			return false
		}
	}
	return true
}

func carouselControllerContent(c *ui.Context, b *carouselCompositionBuild, p CarouselContentProps, children func(*ui.Context)) ui.Element {
	if b.content {
		panic("display: CarouselContainer accepts one CarouselContent")
	}
	b.content = true
	s, frame := b.state, b.controller
	e := frame.engine
	spacing := float32(16)
	if p.Spacing != nil {
		spacing = *p.Spacing
		if !carouselFinite(spacing) || spacing < 0 {
			spacing = 0
		}
	}
	s.spacing = spacing
	b.geometry = carouselMeasure(s.extent, spacing, s.bases, CarouselAlignStart, CarouselTrimSnaps, false)
	height := b.p.Height
	if !carouselFinite(height) || height < 0 {
		height = 0
	}
	if b.p.Vertical && height == 0 {
		height = 240
	}
	viewport := ui.Box(c.Key(carouselViewportPart)).FillWidth().MinWidth(0).Clip().Role(ui.RoleGroup).Label("Slides").Bind(&e.root)
	viewport.ObserveInput(func(event ui.InputEvent) { e.recordPointer(event, false) })
	// Declaring capture does not move the viewport. A non-consuming handler
	// preserves child buttons, while the viewport receives its own captured
	// moves/releases even beyond its bounds. The SDK exposes no ancestor takeover.
	canCapture := frame.optionsError == nil && e.ready && len(e.geometry.snaps) > 0 && e.resolved.Active && e.resolved.WatchDrag && !b.p.NoDrag && !b.p.Disabled && !frame.api.Snapshot().Destroyed
	if canCapture {
		_, _, frame.captureHeld = viewport.Dragged()
		viewport.HandleInput(func(event ui.InputEvent) bool { e.recordPointer(event, true); return false })
	}
	frame.viewport = viewport
	if height > 0 {
		viewport.Height(height)
	}
	viewport.Draw(func(_ *ui.Painter, r ui.Rect) {
		e.viewportBox = r
		extent := r.W
		if b.p.Vertical {
			extent = r.H
		}
		if s.extent != extent {
			s.extent = extent
			c.Invalidate()
		}
	})
	var track ui.Element
	viewport.Children(func() {
		track = ui.Row(c.Key(carouselTrackPart)).FillWidth().AlignItems(ui.Stretch).Bind(&e.track)
		frame.track = track
		if b.p.Vertical {
			track.Column().Height(s.extent+spacing).Margin(-spacing, 0, 0, 0).Translate(0, b.position)
		} else {
			if s.extent > 0 {
				track.Width(s.extent + spacing)
			}
			track.MarginStart(-spacing)
			if height > 0 {
				track.Height(height)
			}
			x := b.position
			if s.rtl {
				x = -x
			}
			track.Translate(x, 0)
		}
		track.Children(func() {
			b.inContent = true
			defer func() { b.inContent = false }()
			if children != nil {
				children(c)
			}
		})
	})
	for _, seen := range b.seen {
		if !seen {
			panic("display: CarouselContent must build every declared item exactly once")
		}
	}
	live := carouselMeasure(s.extent, spacing, s.bases, CarouselAlignStart, CarouselTrimSnaps, false)
	if !live.equal(b.geometry) {
		c.Invalidate()
	}
	s.geometry = live
	return track
}

func carouselControllerItemVisual(b *carouselCompositionBuild, index int, handle ui.Handle) (float32, bool) {
	g, frame := b.geometry, b.controller
	e := frame.engine
	shift := float64(0)
	for i, item := range e.selectedItems {
		if item.canonical == handle {
			shift = e.geometry.shift(i, float64(b.position))
			break
		}
	}
	if index >= len(g.sizes) {
		return 0, false
	}
	start := float64(g.starts[index]-g.spacing+b.position) + shift
	return float32(shift), start < float64(g.viewport)-.01 && start+float64(g.sizes[index]) > .01
}

func (frame *carouselControllerBuild) finalize(c *ui.Context, b *carouselCompositionBuild) {
	e, s := frame.engine, b.state
	if frame.api.Snapshot().Destroyed {
		return
	}
	if frame.optionsError != nil {
		e.lastError = frame.optionsError
		e.invalidProps = frame.invalidOptions
		frame.api.recordError(frame.optionsError)
		return
	}
	if e.invalidProps {
		e.invalidProps = false
		e.lastError = nil
		frame.publish("")
	}
	if frame.options != nil && !e.propsSet || frame.options != nil && !carouselOptionsSame(e.lastProps, *frame.options) {
		e.options = carouselMergeOptions(e.options, *frame.options)
		e.lastProps = carouselCloneOptions(*frame.options)
		e.propsSet = true
		frame.forceReInit = true
	}
	resolved := carouselResolveOptions(e.options, frame.contextEnvironment)
	live := s.geometry
	if s.extent <= 0 {
		return
	}
	handles := make([]ui.Handle, len(b.current))
	for i, item := range b.current {
		handles[i] = item.handle
	}
	slidesChanged := e.ready && !slices.Equal(handles, e.observedHandles)
	resize := e.ready && carouselTrackedResize(e.observedLayout, e.observedHandles, e.selectedItems, live, handles, 0)
	resizeDefault := e.ready && carouselTrackedResize(e.resizeBaseline, e.resizeBaselineHandles, e.selectedItems, live, handles, .5)
	force := frame.forceReInit || !e.ready
	reason := CarouselEventType("")
	if !force && slidesChanged && resolved.Active && resolved.WatchSlides && (frame.hooks.BeforeSlides == nil || frame.hooks.BeforeSlides(frame.api, CarouselSlidesEvent{Previous: slices.Clone(e.observedHandles), Current: slices.Clone(handles)})) {
		force = true
		reason = CarouselSlidesChanged
	}
	if !force && resize && resolved.Active && resolved.WatchResize && (frame.hooks.BeforeResize == nil || frame.hooks.BeforeResize(frame.api, CarouselResizeEvent{Previous: carouselLayoutRects(e.observedLayout), Current: carouselLayoutRects(live)})) && resizeDefault {
		force = true
		reason = CarouselResize
	}
	e.observedHandles = slices.Clone(handles)
	e.observedLayout = live
	commands := frame.api.takeCommands()
	if !e.ready {
		kept := commands[:0]
		for _, command := range commands {
			if command.kind == carouselCommandConfigure || command.kind == carouselCommandReInit || command.kind == carouselCommandPlugins {
				if command.kind == carouselCommandPlugins {
					e.plugins = slices.Clone(command.plugins)
				}
				e.options = carouselMergeOptions(e.options, command.patch)
			} else {
				kept = append(kept, command)
			}
		}
		commands = kept
		resolved = carouselResolveOptions(e.options, frame.contextEnvironment)
	}
	if force {
		if !frame.reinitialize(c, b, resolved, reason) {
			e.options = frame.previousOptions
			return
		}
	}
	if e.ready && *b.selected != e.lastControlled {
		frame.scroll(*b.selected, 0, false, c.Now())
		e.lastControlled = *b.selected
	}
	oldSelected := e.selected
	for _, command := range commands {
		switch command.kind {
		case carouselCommandConfigure, carouselCommandReInit, carouselCommandPlugins:
			previousPlugins := e.plugins
			if command.kind == carouselCommandPlugins {
				e.plugins = slices.Clone(command.plugins)
			}
			previousOptions := e.options
			e.options = carouselMergeOptions(e.options, command.patch)
			frame.startIndex = command.patch.StartIndex
			resolved = carouselResolveOptions(e.options, frame.contextEnvironment)
			if !frame.reinitialize(c, b, resolved, "") {
				e.options = previousOptions
				e.plugins = previousPlugins
				continue
			}
		case carouselCommandAutoplay:
			frame.autoplayCommand(command)
		case carouselCommandNext:
			frame.scroll(e.selected+1, -1, command.jump, c.Now())
		case carouselCommandPrevious:
			frame.scroll(e.selected-1, 1, command.jump, c.Now())
		case carouselCommandTo:
			frame.scroll(command.index, 0, command.jump, c.Now())
		}
	}
	*b.selected = e.selected
	e.lastControlled = e.selected
	if oldSelected != e.selected {
		c.Invalidate()
	}
	frame.updatePlugins(c, b, true)
	var focused ui.Handle
	for i, item := range e.selectedItems {
		if item.canonical.FocusWithin(c) {
			focused = item.canonical
			if e.resolved.Active && e.resolved.WatchFocus && (focused != e.focusedSlide || e.pendingTab) {
				allowed := frame.hooks.BeforeFocus == nil || frame.hooks.BeforeFocus(frame.api, CarouselFocusEvent{Slide: i, Handle: item.public, ViaTab: e.pendingTab, Time: c.Now()})
				if allowed && e.pendingTab && c.Now().Sub(e.lastTab) <= 10*time.Millisecond {
					group := slices.IndexFunc(e.geometry.registry, func(indices []int) bool { return slices.Contains(indices, i) })
					if group >= 0 {
						frame.publish(CarouselSlideFocusStart)
						frame.scroll(group, 0, true, c.Now())
						frame.publish(CarouselSlideFocus)
					}
				}
			}
			break
		}
	}
	e.focusedSlide = focused
	e.pendingTab = false
	frame.processPointers(c, b)
	frame.updatePlugins(c, b, false)
	*b.selected = e.selected
	e.lastControlled = e.selected
	if e.scheduler.Active() && !e.hidden {
		frame.tick(c.Now(), motion.Reduced(c) || b.p.Duration < 0)
	}
	if e.scheduler.Active() && !e.hidden && !(motion.Reduced(c) || b.p.Duration < 0) {
		c.AnimationFrame()
	}
	frame.updateInView(b)
	frame.publish("")
	if b.position != float32(e.offset) && e.resolved.Active {
		c.Invalidate()
	}
	for _, old := range s.bound {
		if old.handle.FocusWithin(c) && !slices.ContainsFunc(b.current, func(item carouselBoundItem) bool { return item.key == old.key && item.shown }) {
			b.root.Focus()
		}
	}
	s.bound = b.current
}

func carouselTrackedResize(baseline carouselGeometry, baselineHandles []ui.Handle, selectedItems []carouselSelectedItem, live carouselGeometry, handles []ui.Handle, threshold float64) bool {
	changed := func(a, b float32) bool {
		difference := math.Abs(float64(a - b))
		return difference > 0 && difference >= threshold
	}
	if changed(baseline.view, live.view) {
		return true
	}
	for _, item := range selectedItems {
		old := slices.Index(baselineHandles, item.canonical)
		current := slices.Index(handles, item.canonical)
		before, after := float32(0), float32(0)
		if old >= 0 && old < len(baseline.sizes) {
			before = baseline.sizes[old]
		}
		if current >= 0 && current < len(live.sizes) {
			after = live.sizes[current]
		}
		if changed(before, after) {
			return true
		}
	}
	return false
}
func carouselLayoutRects(g carouselGeometry) []ui.Rect {
	out := make([]ui.Rect, len(g.sizes))
	for i, size := range g.sizes {
		out[i] = ui.Rect{X: g.starts[i], W: size}
	}
	return out
}

func (frame *carouselControllerBuild) reinitialize(c *ui.Context, b *carouselCompositionBuild, options CarouselResolvedOptions, reason CarouselEventType) bool {
	e := frame.engine
	live := b.state.geometry
	if e.ready {
		start := e.selected
		if frame.startIndex != nil {
			start = *frame.startIndex
		}
		e.options = carouselMergeOptions(e.options, CarouselOptions{StartIndex: &start})
		options = carouselResolveOptions(e.options, frame.contextEnvironment)
	}
	if (options.Axis == CarouselVertical) != b.p.Vertical && e.viewportBox.W > 0 {
		extent := e.viewportBox.W
		if options.Axis == CarouselVertical {
			extent = e.viewportBox.H
		}
		live = carouselMeasure(extent, b.state.spacing, b.state.bases, CarouselAlignStart, CarouselTrimSnaps, false)
	}
	selected := make([]carouselSelectedItem, 0, len(b.current))
	indices := make([]int, 0, len(b.current))
	if options.HasContainer && c.Resolve(options.Container) != frame.track {
		e.lastError = &CarouselOptionError{Field: "Nodes.Container", Code: CarouselForeignNode}
		frame.publish("")
		return false
	}
	if options.HasSlides {
		for _, handle := range options.Slides {
			if handle == (ui.Handle{}) || !c.Resolve(handle).Valid() {
				e.lastError = &CarouselOptionError{Field: "Nodes.Slides", Code: CarouselUnresolvedNode}
				frame.publish("")
				return false
			}
			element := c.Resolve(handle)
			index := slices.Index(frame.items, element)
			if index < 0 {
				e.lastError = &CarouselOptionError{Field: "Nodes.Slides", Code: CarouselForeignNode}
				frame.publish("")
				return false
			}
			indices = append(indices, index)
			selected = append(selected, carouselSelectedItem{canonical: b.current[index].handle, public: handle})
		}
	} else {
		for i, item := range b.current {
			indices = append(indices, i)
			selected = append(selected, carouselSelectedItem{canonical: item.handle, public: item.handle})
		}
	}
	var starts, sizes []float64
	for _, index := range indices {
		starts = append(starts, float64(live.starts[index]))
		sizes = append(sizes, float64(live.sizes[index]))
	}
	g, err := carouselMeasureEngine(float64(live.viewport), float64(live.spacing), starts, sizes, indices, options)
	if err != nil {
		e.lastError = err
		frame.publish("")
		return false
	}
	if err := frame.validatePlugins(c, e.plugins, len(g.snaps), options.Active); err != nil {
		e.lastError = err
		frame.publish("")
		return false
	}
	frame.api.destroyPlugins(frame)
	wasReady := e.ready
	index := e.selected
	if !wasReady {
		index = options.StartIndex
		if *b.selected != 0 && e.options.StartIndex == nil {
			index = *b.selected
		}
	} else {
		index = options.StartIndex
	}
	e.geometry, e.resolved, e.selectedItems = g, options, selected
	e.resolved.Loop = g.loop
	e.observedLayout = live
	e.resizeBaseline = live
	e.resizeBaselineHandles = slices.Clone(e.observedHandles)
	frame.api.acceptOptions(e.options)
	e.selected = g.counter(index)
	e.previous = e.selected
	e.ready, e.initialized, e.lastError = true, true, nil
	e.location = 0
	if len(g.snaps) > 0 {
		e.location = g.snaps[e.selected]
	}
	e.offset, e.previousLocation, e.target = e.location, e.location, e.location
	e.body = carouselphysics.NewBody(e.location, e.resolved.Duration, .68)
	e.scheduler.Stop()
	e.epoch = c.Now()
	e.lastTick = time.Time{}
	e.captured, e.pressed = false, false
	e.pointerMouse, e.suppressClick, e.preventScroll = false, false, false
	e.inView, e.notInView = nil, nil
	e.intersections = nil
	*b.selected = e.selected
	e.lastControlled = e.selected
	frame.publish("")
	frame.initPlugins(c)
	if !wasReady {
		frame.publish(CarouselInit)
	} else {
		frame.publish(CarouselReInit)
	}
	if reason != "" {
		frame.publish(reason)
	}
	c.Invalidate()
	return true
}

func (frame *carouselControllerBuild) updateInView(b *carouselCompositionBuild) {
	e := frame.engine
	if !e.ready || !e.resolved.Active {
		return
	}
	var inside, outside []int
	states := make([]carouselIntersection, len(e.selectedItems))
	changed := len(e.intersections) != len(states)
	for i, item := range e.selectedItems {
		index := slices.IndexFunc(b.current, func(candidate carouselBoundItem) bool { return candidate.handle == item.canonical })
		intersects := false
		ratio := float64(0)
		if index >= 0 && index < len(b.state.geometry.sizes) {
			g := b.state.geometry
			start := float64(g.starts[index]-g.spacing) + e.offset + e.geometry.shift(i, e.offset)
			intersects = start <= float64(g.viewport) && start+float64(g.sizes[index]) >= 0
			if g.sizes[index] > 0 {
				ratio = max(0, min(float64(g.viewport), start+float64(g.sizes[index]))-max(0, start)) / float64(g.sizes[index])
			}
		}
		if intersects {
			inside = append(inside, i)
		} else {
			outside = append(outside, i)
		}
		threshold := 0
		for _, value := range e.resolved.InViewThreshold {
			if ratio >= value {
				threshold++
			}
		}
		states[i] = carouselIntersection{intersecting: intersects, threshold: threshold}
		if i < len(e.intersections) && states[i] != e.intersections[i] {
			changed = true
		}
	}
	e.intersections = states
	e.inView, e.notInView = inside, outside
	if changed {
		frame.publish(CarouselSlidesInView)
	}
}
