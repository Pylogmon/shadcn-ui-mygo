package display

import (
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

// CarouselContainerProps configures the composed carousel. Selected values and
// OnSelect are SNAP indices, which may differ from the number of slides. Align
// defaults to center, as in the source primitive; its multiple-item examples
// explicitly choose CarouselAlignStart. Count is the number of CarouselItems.
type CarouselContainerProps struct {
	Key, Label string
	Count      int
	// Controller is application-owned. Options contains explicit source patches;
	// only supplied fields override compatibility props. Hooks are frame-only.
	Controller *CarouselController
	Options    *CarouselOptions
	// Plugins is an ordered application-owned list: nil preserves, a non-nil
	// empty slice clears. Repeated identities/names follow source ordering.
	Plugins                  []CarouselPlugin
	Hooks                    CarouselWatchHooks
	Vertical, Loop, Disabled bool
	NoDrag                   bool
	Height                   float32
	Align                    CarouselAlign
	Contain                  CarouselContain
	// Duration/DragThreshold configure the nil-Controller/nil-Options/nil-Plugins
	// compatibility profile. Source engine options live in Options; a negative
	// Duration additionally requests the native instant-motion policy.
	Duration      time.Duration
	DragThreshold float32
	OnSelect      func(int)
}

type CarouselContentProps struct {
	// Nil uses the source's 16-DIP compensated leading padding. Point to zero
	// for no spacing. Negative/nonfinite values are treated as zero.
	Spacing *float32
}

type CarouselItemProps struct {
	// Comparable, unique keys preserve editor and Local state across reorders.
	// Nil, noncomparable, nonreflexive or duplicate keys fall back to the index.
	Key   any
	Basis CarouselBasis
}

type CarouselControlProps struct {
	Label    string
	Disabled bool
	// Button overrides the default outline/icon-small styling when supplied.
	Button  *control.ButtonProps
	Content func(*ui.Context)
}

// CarouselSnapshot is detached geometry/selection state for the resolved layout.
// Snaps are logical track translations in DIPs; Visible contains slide indices
// and includes partially visible slides during native dragging/animation.
type CarouselSnapshot struct {
	SelectedSnap, SlideCount         int
	Snaps                            []float32
	Visible                          []int
	CanScrollPrevious, CanScrollNext bool
	Loop                             bool
}

type carouselCompositionKey struct{}
type carouselContainerKey string
type carouselIndexKey int
type carouselItemKey struct{ value any }
type carouselPart uint8

const (
	carouselViewportPart carouselPart = iota
	carouselTrackPart
	carouselPreviousPart
	carouselNextPart
)

type carouselBoundItem struct {
	index  int
	key    any
	handle ui.Handle
	shown  bool
}

// Persistent state contains only values and public Handles. The composition
// scope below holds this build's objects only while synchronous Children runs.
type carouselCompositionState struct {
	initialized, initialLayout  bool
	vertical, rtl               bool
	last, virtual, move, target int
	extent, spacing             float32
	bases                       []CarouselBasis
	geometry                    carouselGeometry
	drag                        ui.Point
	displayedDrag               float32
	bound                       []carouselBoundItem
}

type carouselCompositionScope struct{ build *carouselCompositionBuild }
type carouselCompositionBuild struct {
	root       ui.Element
	state      *carouselCompositionState
	p          CarouselContainerProps
	selected   *int
	choose     func(int)
	keys       func(ui.Element)
	content    bool
	inContent  bool
	seen       []bool
	usedKeys   map[any]bool
	current    []carouselBoundItem
	position   float32
	geometry   carouselGeometry
	controller *carouselControllerBuild
}

func carouselCurrent(c *ui.Context) *carouselCompositionBuild {
	scope := ui.Local(c.Root(), carouselCompositionKey{}, func() carouselCompositionScope { return carouselCompositionScope{} })
	if scope.build == nil {
		panic("display: Carousel composition requires CarouselContainer Children")
	}
	return scope.build
}

// CarouselContainer composes one CarouselContent and optional independent
// controls. Children and all composition helpers run synchronously; replacing
// the returned Element's Children does not reinstall this scope. Nested and
// sibling containers restore the prior scope. No status or controls are added.
// Carousel and CarouselWith retain their original no-gap pager API.
// Native root/control arrow shortcuts run only while that target is focused,
// preserving embedded editors' keys. Vertical uses Up/Down, horizontal follows
// RTL, and Home/End select boundary snaps. This intentionally differs from the
// web primitive's captured Left/Right handler on all descendants.
func CarouselContainer(c *ui.Context, selected *int, p CarouselContainerProps, children func(*ui.Context)) ui.Element {
	if p.Controller != nil || p.Options != nil || p.Plugins != nil {
		return carouselControllerContainer(c, selected, p, children)
	}
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
	s := ui.Local(root, "shadcn.carousel.composition", func() carouselCompositionState { return carouselCompositionState{initialLayout: true, spacing: 16} })
	if len(s.bases) != p.Count {
		bases := make([]CarouselBasis, p.Count)
		copy(bases, s.bases)
		s.bases = bases
		s.initialLayout = true
	}
	rtl := c.Direction() == ui.RightToLeft
	if !s.initialized || s.vertical != p.Vertical || s.rtl != rtl {
		s.initialized, s.initialLayout = true, true
		s.vertical, s.rtl = p.Vertical, rtl
		s.drag = ui.Point{}
	}
	geometry := carouselMeasure(s.extent, s.spacing, s.bases, p.Align, p.Contain, p.Loop)
	if !s.geometry.equal(geometry) {
		s.initialLayout = true
		s.move, s.drag = 0, ui.Point{}
		s.displayedDrag = 0
	}
	s.geometry = geometry
	count := len(s.geometry.snaps)
	if s.extent == 0 && p.Count > 0 {
		count = p.Count
	}
	if count == 0 {
		*selected = 0
	} else {
		*selected = max(0, min(*selected, count-1))
	}
	if s.initialLayout {
		s.last, s.virtual = *selected, *selected
	}
	b := &carouselCompositionBuild{root: root, state: s, p: p, selected: selected, seen: make([]bool, p.Count), usedKeys: make(map[any]bool), geometry: s.geometry}
	b.choose = func(target int) {
		count := len(s.geometry.snaps)
		if count < 2 || p.Disabled {
			return
		}
		movement := target - *selected
		if s.geometry.loop {
			target = carouselModulo(target, count)
		} else {
			target = max(0, min(target, count-1))
		}
		if target == *selected {
			return
		}
		s.move, s.target = movement, target
		*selected = target
		root.SetChanged()
		c.Invalidate()
		if p.OnSelect != nil {
			p.OnSelect(target)
		}
	}
	b.keys = func(e ui.Element) {
		if !e.Focused() || p.Disabled {
			return
		}
		previous, next := ui.KeyLeft, ui.KeyRight
		if rtl {
			previous, next = next, previous
		}
		if p.Vertical {
			previous, next = ui.KeyUp, ui.KeyDown
		}
		switch {
		case e.Shortcut(0, previous):
			b.choose(*selected - 1)
		case e.Shortcut(0, next):
			b.choose(*selected + 1)
		case e.Shortcut(0, ui.KeyHome):
			b.choose(0)
		case e.Shortcut(0, ui.KeyEnd):
			b.choose(len(s.geometry.snaps) - 1)
		}
	}
	b.keys(root)
	position, _ := carouselCompositionPose(c, b)
	b.position = position + s.displayedDrag
	root.Children(func() {
		scope := ui.Local(c.Root(), carouselCompositionKey{}, func() carouselCompositionScope { return carouselCompositionScope{} })
		previous := *scope
		scope.build = b
		defer func() { *scope = previous }()
		if children != nil {
			children(c)
		}
	})
	return root
}

// Sample in Container as well as Content so independently composed siblings
// observe the actual current animation, including when Content follows them.
func carouselCompositionPose(c *ui.Context, b *carouselCompositionBuild) (float32, time.Duration) {
	s := b.state
	if s.last != *b.selected {
		delta := *b.selected - s.last
		if s.geometry.loop && s.move != 0 && s.target == *b.selected {
			delta = s.move
		}
		s.virtual += delta
		s.last, s.move, s.drag, s.displayedDrag = *b.selected, 0, ui.Point{}, 0
	}
	if s.initialLayout {
		s.last, s.virtual = *b.selected, *b.selected
		s.drag, s.displayedDrag = ui.Point{}, 0
	}
	duration := b.p.Duration
	if duration == 0 {
		duration = 250 * time.Millisecond
	}
	if duration < 0 || motion.Reduced(c) || c.WindowHidden() || s.initialLayout {
		duration = -1
	}
	target := s.geometry.position(s.virtual)
	if duration < 0 {
		// AnimateWith clamps negative durations to 1ms. Jump truly resets the
		// retained target at this clock, including after a layout change.
		return b.root.Jump("carousel-composition-position", target), duration
	}
	return b.root.AnimateWith("carousel-composition-position", target, duration, motion.EaseOut), duration
}

// CarouselContent creates the clipped viewport and separately styled native
// track. Build exactly Count items, in index order, inside Children. Fluent
// track styling is available on the returned Element. Container.Height fixes
// the vertical viewport; use Basis and Spacing for scrolling geometry rather
// than overriding the track's or item's axial size.
func CarouselContent(c *ui.Context, p CarouselContentProps, children func(*ui.Context)) ui.Element {
	b := carouselCurrent(c)
	if b.controller != nil {
		return carouselControllerContent(c, b, p, children)
	}
	if b.content {
		panic("display: CarouselContainer accepts one CarouselContent")
	}
	b.content = true
	s := b.state
	spacing := float32(16)
	if p.Spacing != nil {
		spacing = *p.Spacing
		if !carouselFinite(spacing) || spacing < 0 {
			spacing = 0
		}
	}
	if s.spacing != spacing {
		s.spacing, s.initialLayout = spacing, true
		s.geometry = carouselMeasure(s.extent, spacing, s.bases, b.p.Align, b.p.Contain, b.p.Loop)
	}
	b.geometry = s.geometry
	position, duration := carouselCompositionPose(c, b)
	height := b.p.Height
	if !carouselFinite(height) || height < 0 {
		height = 0
	}
	if b.p.Vertical && height == 0 {
		height = 240
	}
	viewport := ui.Box(c.Key(carouselViewportPart)).FillWidth().MinWidth(0).Clip().Role(ui.RoleGroup).Label("Slides")
	if height > 0 {
		viewport.Height(height)
	}
	viewport.Draw(func(_ *ui.Painter, r ui.Rect) {
		extent := r.W
		if b.p.Vertical {
			extent = r.H
		}
		if s.extent != extent {
			s.extent, s.initialLayout = extent, true
			c.Invalidate()
		}
	})
	var track ui.Element
	viewport.Children(func() {
		track = ui.Row(c.Key(carouselTrackPart)).FillWidth().AlignItems(ui.Stretch)
		if b.p.Vertical {
			track.Column().Height(s.extent+spacing).Margin(-spacing, 0, 0, 0)
			track.Translate(0, position)
		} else {
			if s.extent > 0 {
				track.Width(s.extent + spacing)
			}
			track.MarginStart(-spacing)
			if height > 0 {
				track.Height(height)
			}
			x := position
			if s.rtl {
				x = -x
			}
			track.Translate(x, 0)
		}
		draggable := !b.p.NoDrag && !b.p.Disabled && len(s.geometry.snaps) > 1 && s.extent > 0 && !c.WindowHidden()
		s.displayedDrag = 0
		if draggable {
			axis, bounds := ui.Horizontal, ui.Rect{X: -s.extent, W: 2 * s.extent}
			if b.p.Vertical {
				axis, bounds = ui.Vertical, ui.Rect{Y: -s.extent, H: 2 * s.extent}
			}
			dragDuration := duration
			if dragDuration <= 0 {
				dragDuration = time.Nanosecond
			}
			track.Draggable(&s.drag, ui.DraggableOptions{Axis: axis, Bounds: &bounds, Animation: ui.Animation{Duration: dragDuration, Ease: motion.EaseOut}, Snap: func(projected ui.Point) ui.Point {
				if track.PressCanceled() || !c.WindowActive() {
					return ui.Point{}
				}
				threshold := b.p.DragThreshold
				if !carouselFinite(threshold) || threshold <= 0 {
					threshold = max(24, min(96, s.extent*.15))
				}
				movement := -projected.X
				if b.p.Vertical {
					movement = -projected.Y
				} else if s.rtl {
					movement = -movement
				}
				if movement >= threshold {
					b.choose(*b.selected + 1)
				} else if movement <= -threshold {
					b.choose(*b.selected - 1)
				}
				return ui.Point{}
			}})
			drag := track.DragOffset()
			if b.p.Vertical {
				s.displayedDrag = drag.Y
			} else if s.rtl {
				s.displayedDrag = -drag.X
			} else {
				s.displayedDrag = drag.X
			}
		}
		b.position = position + s.displayedDrag
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
	geometry := carouselMeasure(s.extent, spacing, s.bases, b.p.Align, b.p.Contain, b.p.Loop)
	changed := !s.geometry.equal(geometry)
	if changed {
		s.geometry, s.initialLayout = geometry, true
		*b.selected = max(0, min(*b.selected, len(geometry.snaps)-1))
		s.last, s.virtual, s.move, s.drag = *b.selected, *b.selected, 0, ui.Point{}
		b.geometry = geometry
		b.position, _ = carouselCompositionPose(c, b)
		c.Invalidate()
		// Focus decisions use the resolved layout, not the temporary previous
		// geometry used before this build had collected all new item bases.
		for i := range b.current {
			b.current[i].shown = geometry.visible(b.current[i].index, geometry.position(*b.selected))
		}
	} else if s.extent > 0 {
		s.initialLayout = false
	}
	for _, old := range s.bound {
		if old.handle.FocusWithin(c) && !slices.ContainsFunc(b.current, func(item carouselBoundItem) bool { return item.key == old.key && item.shown }) {
			b.root.Focus()
		}
	}
	s.bound = b.current
	return track
}

func CarouselItem(c *ui.Context, index int, p CarouselItemProps, children func(*ui.Context)) ui.Element {
	b := carouselCurrent(c)
	if !b.inContent || index < 0 || index >= b.p.Count || b.seen[index] || index != len(b.current) {
		panic("display: CarouselItem requires each declared index in order inside CarouselContent")
	}
	b.seen[index] = true
	key := any(carouselIndexKey(index))
	if p.Key != nil && reflect.ValueOf(p.Key).Comparable() && p.Key == p.Key {
		candidate := carouselItemKey{p.Key}
		if !b.usedKeys[candidate] {
			key = candidate
		}
	}
	b.usedKeys[key] = true
	b.state.bases[index] = p.Basis
	g := b.geometry
	shown := g.visible(index, b.position)
	item := ui.Box(c.Key(key)).Shrink(0).MinWidth(0).Role(ui.RoleGroup).RoleDescription("slide").Label(fmt.Sprintf("Slide %d of %d", index+1, b.p.Count)).AccessHidden(!shown).IgnoreInput(!shown)
	handle := ui.Local(item, "carousel.item.handle", func() ui.Handle { return ui.Handle{} })
	item.Bind(handle)
	shift := g.occurrence(index, b.position)
	if b.controller != nil {
		shift, shown = carouselControllerItemVisual(b, index, *handle)
		item.AccessHidden(!shown).IgnoreInput(!shown)
		b.controller.items = append(b.controller.items, item)
	}
	b.current = append(b.current, carouselBoundItem{index: index, key: key, handle: *handle, shown: shown})
	size := float32(0)
	if index < len(g.sizes) {
		size = g.sizes[index]
	}
	if b.p.Vertical {
		item.FillWidth().Height(size).Padding(g.spacing, 0, 0, 0).Translate(0, shift)
	} else {
		item.Width(size).PaddingStart(g.spacing)
		offset := shift
		if b.state.rtl {
			offset = -offset
		}
		item.Translate(offset, 0)
	}
	return item.Children(func() {
		if children != nil {
			children(c)
		}
	})
}

// CarouselSnapshotOf returns a detached snapshot in the current synchronous
// Container build, before or after CarouselContent. It samples current motion
// against the last resolved item/spacing layout. Content resolves new item
// declarations and refreshes the native drag sample; changed geometry converges
// through native layout invalidation. A persistent controller is a later stage.
func CarouselSnapshotOf(c *ui.Context) CarouselSnapshot {
	b := carouselCurrent(c)
	g := b.geometry
	snapshot := CarouselSnapshot{SelectedSnap: *b.selected, SlideCount: b.p.Count, Snaps: slices.Clone(g.snaps), Loop: g.loop}
	for index := range g.sizes {
		if g.visible(index, b.position) {
			snapshot.Visible = append(snapshot.Visible, index)
		}
	}
	snapshot.CanScrollPrevious, snapshot.CanScrollNext = carouselCanScroll(g, *b.selected)
	if b.controller != nil {
		e := b.controller.engine
		snapshot.SelectedSnap = e.selected
		snapshot.Snaps = nil
		snapshot.Loop = e.geometry.loop
		for _, snap := range e.geometry.snaps {
			snapshot.Snaps = append(snapshot.Snaps, float32(snap))
		}
		snapshot.CanScrollPrevious = e.geometry.counter(e.selected-1) != e.selected
		snapshot.CanScrollNext = e.geometry.counter(e.selected+1) != e.selected
		snapshot.Visible = nil
		items := b.current
		if !b.content {
			items = b.state.bound
		}
		for _, item := range items {
			// Before Content, sample the last declared Handle/layout values at
			// this frame's solver pose instead of returning an empty inventory.
			_, shown := carouselControllerItemVisual(b, item.index, item.handle)
			if shown {
				snapshot.Visible = append(snapshot.Visible, item.index)
			}
		}
	}
	return snapshot
}

func carouselCanScroll(g carouselGeometry, selected int) (previous, next bool) {
	return len(g.snaps) > 1 && (g.loop || selected > 0), len(g.snaps) > 1 && (g.loop || selected < len(g.snaps)-1)
}

func CarouselPrevious(c *ui.Context, p CarouselControlProps) ui.Element {
	return carouselControl(c, p, true)
}

func CarouselNext(c *ui.Context, p CarouselControlProps) ui.Element {
	return carouselControl(c, p, false)
}

func carouselControl(c *ui.Context, p CarouselControlProps, previous bool) ui.Element {
	b := carouselCurrent(c)
	canPrevious, canNext := carouselCanScroll(b.state.geometry, *b.selected)
	if b.controller != nil {
		e := b.controller.engine
		canPrevious = e.ready && e.geometry.counter(e.selected-1) != e.selected
		canNext = e.ready && e.geometry.counter(e.selected+1) != e.selected
	}
	label, enabled, part := "Next slide", canNext, carouselNextPart
	if previous {
		label, enabled, part = "Previous slide", canPrevious, carouselPreviousPart
	}
	if p.Label != "" {
		label = p.Label
	}
	props := control.ButtonProps{Variant: control.Outline, Size: control.IconSmall}
	if p.Button != nil {
		props = *p.Button
	}
	if p.Content != nil {
		props.Icon = nil
	}
	props.Label, props.Disabled = label, props.Disabled || p.Disabled || b.p.Disabled || !enabled
	button := control.Button(c.Key(part), "", props).Absolute().Radius(999)
	if b.p.Vertical {
		button.LeftPercent(50).TranslatePercent(-50, 0)
		if previous {
			button.Top(-48)
		} else {
			button.Bottom(-48)
		}
	} else {
		button.TopPercent(50).TranslatePercent(0, -50)
		if previous != b.state.rtl {
			button.Left(-48)
		} else {
			button.Right(-48)
		}
	}
	if props.Icon == nil {
		button.Children(func() {
			if p.Content != nil {
				p.Content(c)
				return
			}
			icon := outlined.ChevronRight()
			if previous != (!b.p.Vertical && b.state.rtl) {
				icon = outlined.ChevronLeft()
			}
			e := material.Icon(c, icon, material.Props{Size: 16})
			if b.p.Vertical {
				e.Rotate(90)
			}
		})
	}
	if !props.Disabled && !props.Loading {
		b.keys(button)
	}
	if button.Clicked() {
		delta := 1
		if previous {
			delta = -1
		}
		b.choose(*b.selected + delta)
	}
	return button
}
