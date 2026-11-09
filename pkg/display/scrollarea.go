package display

import (
	"fmt"
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type ScrollbarVisibility uint8

const (
	// ScrollbarsAlways shows the bars whenever their axes overflow.
	ScrollbarsAlways ScrollbarVisibility = iota
	// ScrollbarsAuto shows them on hover, focus within the area or thumb capture.
	ScrollbarsAuto
	// ScrollbarsHidden keeps native wheel, keyboard and AX scrolling available.
	ScrollbarsHidden
)

type ScrollAreaPosition struct{ X, Y, MaxX, MaxY float32 }

type ScrollAreaProps struct {
	// Zero Orientation is vertical. Horizontal|Vertical scrolls both ways.
	Orientation        ui.Orientation
	State              *ui.ScrollState
	Label, Description string
	Scrollbars         ScrollbarVisibility
	// NativeScrollbars explicitly selects the backend's overlay indicators.
	// The default uses shadcn's 10-DIP tracks and semantic Border thumbs.
	NativeScrollbars bool
	// Zero Thickness and MinThumbSize use 10 and 16 DIPs respectively.
	Thickness, MinThumbSize float32
	// Source bars are pointer controls without an extra Tab stop. An opt-in
	// focusable bar uses the supported native slider role and numeric range.
	FocusableScrollbars bool
	// OnScroll receives the accepted offset after native layout/clamping.
	// Initial placement is silent; wheel, keyboard, drag and app changes notify.
	OnScroll func(ScrollAreaPosition)
}

type ScrollAreaParts struct {
	Root, Viewport, VerticalBar, HorizontalBar, Corner ui.Element
	State                                              *ui.ScrollState
}

type scrollAreaViewportKey struct{}
type scrollAreaVerticalKey struct{}
type scrollAreaHorizontalKey struct{}
type scrollAreaCornerKey struct{}
type scrollAreaStateKey struct{}
type scrollAreaRuntimeKey struct{}
type scrollBarRuntimeKey struct{}
type scrollBarThumbKey struct{}

type scrollAreaRuntime struct {
	width, height, maxX, maxY float32
	lastX, lastY              float32
	initialized, changed      bool
	vertical, horizontal      *scrollBarRuntime
}

// ScrollArea retains its original contract: the returned element is the actual
// native vertical viewport, including fluent TrackScroll/ScrollSnap support.
func ScrollArea(c *ui.Context, content func()) ui.Element {
	return ui.Scroll(c).Fill().Children(content)
}

// ScrollAreaWith separates the layout surface from its real native viewport.
// Size and decorate Root, and use Props.State or Parts.State to move the
// content. It preserves inherited text styles and the complete current theme.
func ScrollAreaWith(c *ui.Context, p ScrollAreaProps, content func(*ui.Context)) ScrollAreaParts {
	axis := p.Orientation
	valid := axis == 0 || axis == ui.Horizontal || axis == ui.Vertical || axis == ui.Horizontal|ui.Vertical
	if axis == 0 || !valid {
		axis = ui.Vertical
	}
	if p.Scrollbars > ScrollbarsHidden {
		p.Scrollbars, valid = ScrollbarsAlways, false
	}
	root := ui.Box(c).Fill().MinWidth(0).MinHeight(0).Clip().Role(ui.RoleNone).Invalid(!valid)
	local := ui.Local(root, scrollAreaStateKey{}, func() ui.ScrollState { return ui.ScrollState{} })
	state := p.State
	if state == nil {
		state = local
	}
	if !scrollFinite(state.X) || !scrollFinite(state.Y) {
		resetScrollMotion(state, scrollNumber(state.X), scrollNumber(state.Y))
		root.Invalid(true)
	}
	s := ui.Local(root, scrollAreaRuntimeKey{}, func() scrollAreaRuntime { return scrollAreaRuntime{} })
	prepareBar := func(bar **scrollBarRuntime, active bool) {
		if old := *bar; old != nil {
			if old.freezePending && old.config.state != nil {
				resetScrollMotion(old.config.state, old.pendingX, old.pendingY)
				old.freezePending = false
			}
			if !active || old.config.state != state {
				if !active && old.config.state != nil && old.focus.Focused(c) && !root.IsDisabled() {
					old.config.state.Handle.Focus()
				}
				old.drag = nil
			}
			if !active {
				*bar = nil
			}
		}
	}
	customActive := !p.NativeScrollbars && p.Scrollbars != ScrollbarsHidden && !root.IsDisabled()
	prepareBar(&s.vertical, customActive && axis&ui.Vertical != 0)
	prepareBar(&s.horizontal, customActive && axis&ui.Horizontal != 0)
	if s.changed {
		root.SetChanged()
		s.changed = false
	}
	parts := ScrollAreaParts{Root: root, State: state}
	root.Children(func() {
		ctx := c.Key(scrollAreaViewportKey{})
		switch axis {
		case ui.Horizontal:
			parts.Viewport = ui.ScrollHorizontal(ctx)
		case ui.Horizontal | ui.Vertical:
			parts.Viewport = ui.ScrollBoth(ctx)
		default:
			parts.Viewport = ui.Scroll(ctx)
		}
		viewport := parts.Viewport.Fill().MinWidth(0).MinHeight(0).TrackScroll(state).FocusRing(false).Invalid(!valid).
			Label(p.Label).Description(p.Description)
		hasX, hasY := axis&ui.Horizontal != 0 && state.MaxX > .5, axis&ui.Vertical != 0 && state.MaxY > .5
		if hasX || hasY {
			viewport.Focusable()
		}
		if viewport.FocusVisible() {
			root.Shadow(0, 0, 0, 3, theme.TokensOf(c).Ring.Alpha(.5))
		}
		viewport.Scrollbars(p.NativeScrollbars && p.Scrollbars != ScrollbarsHidden).
			KeepScrollbars(p.Scrollbars == ScrollbarsAlways)
		viewport.Draw(func(_ *ui.Painter, bounds ui.Rect) {
			geometryChanged := s.width != bounds.W || s.height != bounds.H || s.maxX != state.MaxX || s.maxY != state.MaxY
			s.width, s.height, s.maxX, s.maxY = bounds.W, bounds.H, state.MaxX, state.MaxY
			x, y := state.X, state.Y
			if s.initialized && (s.lastX != x || s.lastY != y) {
				s.changed = true
				if p.OnScroll != nil {
					p.OnScroll(ScrollAreaPosition{state.X, state.Y, state.MaxX, state.MaxY})
				}
				c.Invalidate()
			}
			s.lastX, s.lastY, s.initialized = x, y, true
			if geometryChanged {
				// One correction frame lays out the bars from measured native bounds.
				c.Invalidate()
			}
		}).Children(func() {
			if content != nil {
				content(c)
			}
		})
		if !p.NativeScrollbars {
			visible := p.Scrollbars != ScrollbarsHidden
			if p.Scrollbars == ScrollbarsAuto {
				visible = root.Hovered() || root.FocusWithin()
			}
			common := ScrollBarProps{Thickness: p.Thickness, MinThumbSize: p.MinThumbSize, Focusable: p.FocusableScrollbars, Hidden: !visible, Disabled: p.Scrollbars == ScrollbarsHidden,
				Corner: hasX && hasY}
			if axis&ui.Vertical != 0 {
				props := common
				props.Label = "Vertical scroll bar"
				parts.VerticalBar = ScrollBar(c.Key(scrollAreaVerticalKey{}), state, props)
				s.vertical = ui.Local(parts.VerticalBar, scrollBarRuntimeKey{}, func() scrollBarRuntime { return scrollBarRuntime{} })
			}
			if axis&ui.Horizontal != 0 {
				props := common
				props.Orientation, props.Label = ui.Horizontal, "Horizontal scroll bar"
				parts.HorizontalBar = ScrollBar(c.Key(scrollAreaHorizontalKey{}), state, props)
				s.horizontal = ui.Local(parts.HorizontalBar, scrollBarRuntimeKey{}, func() scrollBarRuntime { return scrollBarRuntime{} })
			}
			if hasX && hasY {
				thickness, _ := scrollBarNumber(p.Thickness, 10, 4)
				corner := ui.Box(c.Key(scrollAreaCornerKey{})).Absolute().Bottom(0).Size(thickness, thickness).IgnoreInput(true).AccessHidden(true)
				if c.Direction() == ui.RightToLeft {
					corner.Left(0)
				} else {
					corner.Right(0)
				}
				parts.Corner = corner
			}
		}
	})
	return parts
}

type ScrollBarProps struct {
	// Zero is vertical; horizontal uses the viewport's logical X offset.
	Orientation ui.Orientation
	// ViewportSize zero uses the tracked viewport's previous layout bounds.
	ViewportSize, Thickness, MinThumbSize, Step float32
	Corner, Hidden, Disabled, Focusable         bool
	Label                                       string
	ThumbColor                                  *ui.Color
	// OnScrollIntent runs before a valid manual press, wheel or focused scroll
	// key changes the offset. Programmatic scrolling is silent. Use model-only
	// callbacks, for example to release transcript auto-follow.
	OnScrollIntent func()
}

type scrollBarConfig struct {
	state                                 *ui.ScrollState
	horizontal, rtl, disabled, focusable  bool
	view, reach, thickness, minimum, step float32
}

type scrollBarDrag struct {
	from, offset, before float64
	reach, travel        float64
}

type scrollBarRuntime struct {
	config                         scrollBarConfig
	length, origin, expected       float32
	initialized, changed, geometry bool
	drag                           *scrollBarDrag
	focus                          ui.Handle
	freezePending                  bool
	pendingX, pendingY             float32
	onScrollIntent                 func()
}

func scrollFinite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
func scrollNumber(v float32) float32 {
	if !scrollFinite(v) {
		return 0
	}
	return max(0, v)
}
func scrollBarNumber(v, fallback, minimum float32) (float32, bool) {
	if v == 0 {
		return fallback, true
	}
	if !scrollFinite(v) || v < minimum {
		return fallback, false
	}
	return v, true
}

// Resetting the caller-owned public value stops a prior ScrollTo animation.
// The stable native Handle and measured range remain attached to this viewport.
func resetScrollMotion(state *ui.ScrollState, x, y float32) {
	*state = ui.ScrollState{Handle: state.Handle, X: x, Y: y, MaxX: state.MaxX, MaxY: state.MaxY}
}

func (s *scrollBarRuntime) offset() float32 {
	if s.config.state == nil {
		return 0
	}
	if s.config.horizontal {
		return s.config.state.X
	}
	return s.config.state.Y
}

func (s *scrollBarRuntime) write(offset float64) {
	next := float32(max(0, min(float64(s.config.reach), offset)))
	state := s.config.state
	x, y := state.X, state.Y
	if s.config.horizontal {
		x = next
	} else {
		y = next
	}
	if x != state.X || y != state.Y {
		if s.freezePending {
			state.X, state.Y = x, y
			s.pendingX, s.pendingY = x, y
		} else {
			resetScrollMotion(state, x, y)
		}
		s.expected, s.changed = next, true
	}
}

func (s *scrollBarRuntime) freezeMotion() {
	state := s.config.state
	resetScrollMotion(state, state.X, state.Y)
	// A same-position ScrollTo cancels the runtime's wheel/reveal glide before
	// the next view builds. It has no distance to animate. The builder then
	// restores newer captured input before retracking the native viewport.
	s.pendingX, s.pendingY, s.freezePending = state.X, state.Y, true
	state.ScrollTo(state.X, state.Y, ui.Animation{})
}

func (s *scrollBarRuntime) thumb() (start, length, travel float32) {
	cfg := s.config
	inner := max(0, s.length-2)
	if cfg.view <= 0 || cfg.reach <= 0 {
		return 1, inner, 0
	}
	length = min(inner, max(cfg.minimum, float32(float64(inner)*float64(cfg.view)/(float64(cfg.view)+float64(cfg.reach)))))
	travel = max(0, inner-length)
	portion := float32(0)
	if cfg.reach > 0 {
		portion = max(0, min(1, s.offset()/cfg.reach))
	}
	if cfg.horizontal && cfg.rtl {
		portion = 1 - portion
	}
	return 1 + portion*travel, length, travel
}

func (s *scrollBarRuntime) input(ev ui.InputEvent) bool {
	cfg := s.config
	if cfg.disabled || cfg.state == nil || !s.geometry {
		s.drag = nil
		return false
	}
	if s.drag != nil && (s.offset() != s.expected || cfg.state.Scrolling() && !s.freezePending) {
		s.drag = nil // app-owned changes take precedence over an old capture
	}
	axis := ev.Y
	if cfg.horizontal {
		axis = ev.X
	}
	switch ev.Kind {
	case ui.InputPointerDown:
		if ev.Button != 0 || !scrollFinite(axis) {
			return false
		}
		if s.onScrollIntent != nil {
			s.onScrollIntent()
		}
		start, length, travel := s.thumb()
		s.freezeMotion()
		if travel <= 0 {
			return true
		}
		before := s.offset()
		if axis < start || axis > start+length {
			portion := float64((axis - 1 - length/2) / travel)
			if cfg.horizontal && cfg.rtl {
				portion = 1 - portion
			}
			s.write(portion * float64(cfg.reach))
		}
		s.expected = s.offset()
		s.drag = &scrollBarDrag{from: float64(s.origin) + float64(axis), offset: float64(s.offset()), before: float64(before), reach: float64(cfg.reach), travel: float64(travel)}
		return true
	case ui.InputPointerMove, ui.InputPointerUp:
		if s.drag == nil || ev.Kind == ui.InputPointerUp && ev.Button != 0 {
			return false
		}
		if ev.Canceled {
			s.write(s.drag.before)
		} else if scrollFinite(axis) {
			delta := float64(s.origin) + float64(axis) - s.drag.from
			if cfg.horizontal && cfg.rtl {
				delta = -delta
			}
			s.write(s.drag.offset + delta*s.drag.reach/s.drag.travel)
		}
		if ev.Kind == ui.InputPointerUp {
			s.drag = nil
		}
		return true
	case ui.InputScroll:
		if ev.Mods&^ui.Shift != 0 {
			return false
		}
		delta := ev.DY
		if cfg.horizontal {
			delta = ev.DX
			if ev.Mods&ui.Shift != 0 && delta == 0 {
				delta = ev.DY
			} else if cfg.rtl {
				delta = -delta
			}
		}
		before := s.offset()
		if scrollFinite(delta) {
			if delta != 0 && s.onScrollIntent != nil {
				s.onScrollIntent()
			}
			s.freezeMotion()
			s.write(float64(before) + float64(delta))
		}
		return s.offset() != before
	case ui.InputKeyDown:
		if !cfg.focusable || ev.Mods&^ui.Shift != 0 {
			return false
		}
		delta := float64(cfg.step)
		if ev.Mods&ui.Shift != 0 {
			delta *= 10
		}
		next := float64(s.offset())
		switch ev.Key {
		case ui.KeyRight:
			if cfg.horizontal && cfg.rtl {
				delta = -delta
			}
			next += delta
		case ui.KeyLeft:
			if cfg.horizontal && cfg.rtl {
				delta = -delta
			}
			next -= delta
		case ui.KeyDown:
			if cfg.horizontal && cfg.rtl {
				delta = -delta
			}
			next += delta
		case ui.KeyUp:
			if cfg.horizontal && cfg.rtl {
				delta = -delta
			}
			next -= delta
		case ui.KeyHome:
			next = 0
		case ui.KeyEnd:
			next = float64(cfg.reach)
		case ui.KeyPageDown, ui.KeyPageUp, ui.KeySpace:
			delta = float64(max(cfg.view-40, cfg.view/2))
			if ev.Key == ui.KeyPageUp || ev.Key == ui.KeySpace && ev.Mods&ui.Shift != 0 {
				delta = -delta
			}
			next += delta
		default:
			return false
		}
		if s.onScrollIntent != nil {
			s.onScrollIntent()
		}
		s.drag = nil
		s.freezeMotion()
		s.write(next)
		return true
	}
	return false
}

// ScrollBar is an independently composable track for a native TrackScroll
// viewport. It manipulates that viewport's real ScrollState, never a parallel
// visual offset. Build it after the matching viewport, inside a bounded box.
func ScrollBar(c *ui.Context, state *ui.ScrollState, p ScrollBarProps) ui.Element {
	horizontal := p.Orientation == ui.Horizontal
	valid := p.Orientation == 0 || p.Orientation == ui.Vertical || horizontal
	thickness, thicknessOK := scrollBarNumber(p.Thickness, 10, 4)
	minimum, minimumOK := scrollBarNumber(p.MinThumbSize, 16, .01)
	step, stepOK := scrollBarNumber(p.Step, 40, .01)
	valid = valid && thicknessOK && minimumOK && stepOK && state != nil
	root := ui.Box(c).Absolute().ZIndex(1).Invalid(!valid)
	s := ui.Local(root, scrollBarRuntimeKey{}, func() scrollBarRuntime { return scrollBarRuntime{} })
	root.Bind(&s.focus)
	viewport, reach, view := ui.Element{}, float32(0), p.ViewportSize
	if state != nil {
		viewport = c.Resolve(state.Handle)
		if s.freezePending && s.config.state == state {
			resetScrollMotion(state, s.pendingX, s.pendingY)
			if viewport.Valid() {
				viewport.TrackScroll(state)
			}
		}
		s.freezePending = false
		bounds := viewport.LayoutBounds()
		if horizontal {
			reach = scrollNumber(state.MaxX)
			if view == 0 {
				view = bounds.W
			}
		} else {
			reach = scrollNumber(state.MaxY)
			if view == 0 {
				view = bounds.H
			}
		}
	}
	if !scrollFinite(view) || view < 0 {
		view, valid = 0, false
	}
	rtl := c.Direction() == ui.RightToLeft
	corner := float32(0)
	if p.Corner {
		corner = thickness
	}
	if horizontal {
		root.Left(0).Right(corner).Bottom(0).Height(thickness)
		if rtl {
			root.Left(corner).Right(0)
		}
	} else {
		root.Top(0).Bottom(corner).Width(thickness)
		if rtl {
			root.Left(0)
		} else {
			root.Right(0)
		}
		// Native AX increment maps Right to positive values for this slider.
		root.Direction(ui.LeftToRight)
	}
	shown := state != nil && reach > .5 && view > 0
	disabled := p.Disabled || !shown || !valid || root.IsDisabled()
	config := scrollBarConfig{state: state, horizontal: horizontal, rtl: rtl, disabled: disabled, focusable: p.Focusable,
		view: view, reach: reach, thickness: thickness, minimum: minimum, step: step}
	if !s.initialized || config != s.config || s.drag != nil && (s.offset() != s.expected || state.Scrolling()) {
		s.drag = nil
	}
	s.config, s.initialized, s.onScrollIntent = config, true, p.OnScrollIntent
	hidden := p.Hidden && s.drag == nil
	if s.changed {
		root.SetChanged()
		s.changed = false
	}
	root.Invalid(!valid).Disabled(disabled).DimDisabled(false).IgnoreInput(hidden || !shown).AccessHidden(true)
	if hidden || !shown {
		root.Opacity(0)
	}
	if p.Focusable && shown && !hidden {
		label := p.Label
		if label == "" {
			label = "Vertical scroll bar"
			if horizontal {
				label = "Horizontal scroll bar"
			}
		}
		orientation := ui.Vertical
		if horizontal {
			orientation = ui.Horizontal
		}
		root.Focusable().Role(ui.RoleSlider).AccessHidden(false).Label(label).RoleDescription("scroll bar").Orientation(orientation).
			Range(0, float64(reach), float64(scrollNumber(s.offset()))).Step(float64(step)).Value(fmt.Sprintf("%.0f of %.0f", s.offset(), reach)).FocusRing(false)
		if viewport.Valid() {
			root.Controls(viewport)
		}
		if root.FocusVisible() {
			root.Shadow(0, 0, 0, 3, theme.TokensOf(c).Ring.Alpha(.5))
		}
	} else {
		root.KeepFocus()
	}
	root.Draw(func(_ *ui.Painter, bounds ui.Rect) {
		length, origin := bounds.H, bounds.Y
		if horizontal {
			length, origin = bounds.W, bounds.X
		}
		if s.length != length {
			s.drag = nil
			s.length = length
			c.Invalidate()
		}
		s.origin, s.geometry = origin, true
	}).HandleInput(s.input)
	return root.Children(func() {
		start, length, _ := s.thumb()
		color := theme.TokensOf(c).Border
		if p.ThumbColor != nil {
			color = *p.ThumbColor
		}
		thumb := ui.Box(c.Key(scrollBarThumbKey{})).Absolute().Radius(999).Background(color).IgnoreInput(true).AccessHidden(true)
		motion.ApplyTransition(c, thumb)
		if horizontal {
			thumb.Left(start).Top(2).Width(length).Height(max(0, thickness-3))
		} else {
			thumb.Top(start).Left(2).Height(length).Width(max(0, thickness-3))
		}
	})
}
