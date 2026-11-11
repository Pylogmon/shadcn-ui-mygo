package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type ScrollbarVisibility uint8

const (
	// ScrollbarsAuto uses MyGo's default hover and drag visibility.
	ScrollbarsAuto ScrollbarVisibility = iota
	// ScrollbarsAlways keeps MyGo's bars visible whenever their axes overflow.
	ScrollbarsAlways
	// ScrollbarsHidden keeps native wheel, keyboard and AX scrolling available.
	ScrollbarsHidden
)

type ScrollAreaPosition struct{ X, Y, MaxX, MaxY float32 }

type ScrollAreaProps struct {
	// Zero Orientation is vertical. Horizontal|Vertical scrolls both ways.
	Orientation        ui.Orientation
	State              *ui.ScrollState
	Label, Description string
	// Scrollbars only controls visibility; MyGo owns the style and interaction.
	Scrollbars ScrollbarVisibility
	// OnScroll receives the accepted offset after native layout/clamping.
	// Initial placement is silent; wheel, keyboard, drag and app changes notify.
	OnScroll func(ScrollAreaPosition)
}

type ScrollAreaParts struct {
	Root, Viewport ui.Element
	State          *ui.ScrollState
}

type scrollAreaViewportKey struct{}
type scrollAreaStateKey struct{}
type scrollAreaRuntimeKey struct{}

type scrollAreaRuntime struct {
	width, height, maxX, maxY float32
	lastX, lastY              float32
	initialized, changed      bool
}

// ScrollArea returns the actual MyGo vertical viewport, including fluent
// TrackScroll/ScrollSnap support and MyGo's standard overlay scrollbars.
func ScrollArea(c *ui.Context, content func()) ui.Element {
	return ui.Scroll(c).Fill().Children(content)
}

// ScrollAreaWith separates the layout surface from its MyGo viewport.
// Size and decorate Root, and use Props.State or Parts.State to move the
// content. Scrollbar painting, geometry and input all belong to MyGo.
func ScrollAreaWith(c *ui.Context, p ScrollAreaProps, content func(*ui.Context)) ScrollAreaParts {
	axis := p.Orientation
	valid := axis == 0 || axis == ui.Horizontal || axis == ui.Vertical || axis == ui.Horizontal|ui.Vertical
	if axis == 0 || !valid {
		axis = ui.Vertical
	}
	if p.Scrollbars > ScrollbarsHidden {
		p.Scrollbars, valid = ScrollbarsAuto, false
	}
	root := ui.Box(c).Fill().MinWidth(0).MinHeight(0).Clip().Role(ui.RoleNone).Invalid(!valid)
	state := p.State
	if state == nil {
		state = ui.Local(root, scrollAreaStateKey{}, func() ui.ScrollState { return ui.ScrollState{} })
	}
	if !scrollFinite(state.X) || !scrollFinite(state.Y) {
		*state = ui.ScrollState{Handle: state.Handle, X: scrollNumber(state.X), Y: scrollNumber(state.Y), MaxX: state.MaxX, MaxY: state.MaxY}
		root.Invalid(true)
	}
	s := ui.Local(root, scrollAreaRuntimeKey{}, func() scrollAreaRuntime { return scrollAreaRuntime{} })
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
		if axis&ui.Horizontal != 0 && state.MaxX > .5 || axis&ui.Vertical != 0 && state.MaxY > .5 {
			viewport.Focusable()
		}
		if viewport.FocusVisible() {
			root.Shadow(0, 0, 0, 3, theme.TokensOf(c).Ring.Alpha(.5))
		}
		viewport.Scrollbars(p.Scrollbars != ScrollbarsHidden).KeepScrollbars(p.Scrollbars == ScrollbarsAlways)
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
				// Reflect native overflow in the viewport's keyboard focusability.
				c.Invalidate()
			}
		}).Children(func() {
			if content != nil {
				content(c)
			}
		})
	})
	return parts
}

func scrollFinite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
func scrollNumber(v float32) float32 {
	if !scrollFinite(v) {
		return 0
	}
	return max(0, v)
}
