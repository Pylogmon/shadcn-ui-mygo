package overlay

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// DrawerSnapPoint is a visible extent: Size is in DIPs; Fraction is a share of
// the window along the swipe axis. Set exactly one. Mixed units are supported.
// Resolved points must be strictly increasing; closing is a separate outcome.
type DrawerSnapPoint struct{ Size, Fraction float32 }

type DrawerProps struct {
	Title, Description string
	// Height preserves the original vertical API; Size works for every side.
	Height, Width, Size, MaxSize float32
	Side                         Placement
	LogicalSide                  bool
	SnapPoints                   []DrawerSnapPoint
	// ActiveSnap indexes SnapPoints (or the single default point), starting at 0.
	// Controlled changes animate without reporting OnSnapChange.
	ActiveSnap *int
	// HandleOnly limits swiping to the real visible handle. Otherwise the
	// panel's unclaimed background also swipes. Interactive children retain
	// their own gestures; MyGo does not hand their presses to parent draggables.
	HandleOnly, ShowSwipeHandle, DisableSwipe                        bool
	HideClose, HideHeader, NonModal                                  bool
	PreventOutsideDismiss, PreventEscapeDismiss, PreventSwipeDismiss bool
	InitialFocus, ReturnFocus                                        *ui.Handle
	ContentPadding, ContentGap                                       *float32
	Footer                                                           func()
	OnSnapChange                                                     func(index int)
	OnOpenChange                                                     func(bool)
	// DismissThreshold is the fraction below the smallest snap needed to close,
	// default 1/3. Native velocity projection also recognizes a fast flick.
	DismissThreshold float32
}

type drawerState struct {
	offset, shown                           ui.Point
	index                                   int
	initialized, wasOpen, entering, present bool
	configuration                           string
	nested                                  int
	held, handleDriver                      bool
	cancelGesture, cancelHandle             bool
	cancelStarted                           bool
	suspendedFrames                         int
}
type drawerStack struct{ rendering []*drawerState }

func drawerAxis(p ui.Point, side Placement) float32 {
	switch side {
	case Top:
		return -p.Y
	case Left:
		return -p.X
	case Right:
		return p.X
	default:
		return p.Y
	}
}
func drawerPoint(v float32, side Placement) ui.Point {
	switch side {
	case Top:
		return ui.Point{Y: -v}
	case Left:
		return ui.Point{X: -v}
	case Right:
		return ui.Point{X: v}
	default:
		return ui.Point{Y: v}
	}
}
func drawerBounds(size float32, side Placement) ui.Rect {
	switch side {
	case Top:
		return ui.Rect{Y: -size, H: size}
	case Left:
		return ui.Rect{X: -size, W: size}
	case Right:
		return ui.Rect{W: size}
	default:
		return ui.Rect{H: size}
	}
}
func drawerPoints(p DrawerProps, extent float32, side Placement) (full float32, points []float32, valid bool) {
	valid = side >= Bottom && side <= Right
	maximum, good := modalDimension(p.MaxSize, max(float32(1), extent-16))
	valid = valid && good
	maximum = min(max(float32(1), extent), maximum)
	fallback := p.Height
	if side == Left || side == Right {
		fallback = p.Width
	}
	fallback, good = modalDimension(fallback, 384)
	valid = valid && good
	full, good = modalDimension(p.Size, fallback)
	valid = valid && good
	full = min(maximum, full)
	if len(p.SnapPoints) == 0 {
		return full, []float32{full}, valid
	}
	for _, point := range p.SnapPoints {
		size, fraction := point.Size, point.Fraction
		if !drawerFinite(size) || !drawerFinite(fraction) || size < 0 || fraction < 0 || fraction > 1 || (size > 0) == (fraction > 0) {
			valid = false
			continue
		}
		if fraction > 0 {
			size = extent * fraction
		}
		size = min(maximum, size)
		if size <= 0 || len(points) > 0 && size <= points[len(points)-1] {
			valid = false
		}
		points = append(points, size)
		full = max(full, size)
	}
	if !valid || len(points) != len(p.SnapPoints) {
		return full, []float32{full}, false
	}
	return full, points, true
}
func drawerFinite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }

// Drawer is an independently controlled edge panel inside the parent window.
// A native draggable sheet supplies capture, velocity projection and cancelled
// gesture rollback. Its backdrop stays stationary. HandleOnly uses that same
// native gesture on the actual handle while moving the whole panel with it.
func Drawer(c *ui.Context, open *bool, p DrawerProps, content func()) ui.Element {
	if open == nil {
		return ui.Element{}
	}
	t := theme.TokensOf(c)
	side := logicalSide(p.Side, p.LogicalSide, c.Direction())
	width, height := c.Size()
	extent := height
	if side == Left || side == Right {
		extent = width
	}
	full, points, valid := drawerPoints(p, extent, side)
	marker := ui.Box(c).Absolute().Size(0, 0).AccessHidden(true).IgnoreInput(true)
	s := ui.Local(marker, "shadcn.drawer", func() drawerState { return drawerState{} })
	stack := ui.Local(c.Root(), "shadcn.drawer-stack", func() drawerStack { return drawerStack{} })
	s.nested = 0
	if *open || s.present {
		for _, parent := range stack.rendering {
			parent.nested++
		}
	}
	index := s.index
	if p.ActiveSnap != nil {
		index = *p.ActiveSnap
	}
	if valid {
		index = max(0, min(index, len(points)-1))
		if p.ActiveSnap != nil {
			*p.ActiveSnap = index
		}
	} else {
		index = 0
	}
	var config strings.Builder
	fmt.Fprintf(&config, "%v/%v/%t/%t/%t/", side, full, p.HandleOnly, p.DisableSwipe, valid)
	for _, point := range points {
		fmt.Fprintf(&config, "%g,", point)
	}
	rest := full - points[index]
	opening := *open && !s.wasOpen
	configurationChanged := s.initialized && s.configuration != config.String()
	if (configurationChanged || p.ActiveSnap != nil && index != s.index) && s.held {
		// Stop registering the native gesture until the old capture is over.
		// One disabled build would let the same press resume its old drag.
		s.cancelGesture, s.cancelHandle = true, s.handleDriver
		s.cancelStarted, s.suspendedFrames = false, 0
	}
	if !s.initialized || s.configuration != config.String() {
		s.initialized = true
		s.configuration = config.String()
		s.offset = drawerPoint(rest, side)
	}
	if opening {
		if !s.present && !motion.Reduced(c) {
			s.offset = drawerPoint(full, side)
			s.entering = true
		} else {
			// Reverse the current native settle instead of retaining the
			// closed target when the caller reopens during the exit.
			s.offset = drawerPoint(rest, side)
			s.entering = false
		}
	} else if s.entering {
		s.offset = drawerPoint(rest, side)
		s.entering = false
	} else if index != s.index && *open {
		s.offset = drawerPoint(rest, side)
	}
	s.index = index
	s.wasOpen = *open
	closeThreshold := p.DismissThreshold
	if !drawerFinite(closeThreshold) || closeThreshold <= 0 || closeThreshold >= 1 {
		closeThreshold = 1.0 / 3
	}
	padding := modalSpace(p.ContentPadding, 16)
	gap := modalSpace(p.ContentGap, 16)
	var shownOffset ui.Point
	result := present(c, open, modalOptions{preventOutside: p.PreventOutsideDismiss, preventEscape: p.PreventEscapeDismiss, nonModal: p.NonModal, duration: 450 * time.Millisecond, initialFocus: p.InitialFocus, returnFocus: p.ReturnFocus, onOpenChange: p.OnOpenChange}, func(back, panel ui.Element, progress float32) {
		if p.NonModal {
			back.Background(ui.Transparent)
		} else {
			back.Background(ui.RGBA(0, 0, 0, .1*progress))
		}
		panel.Column().Gap(gap).Padding(padding).Background(t.Popover).TextColor(t.PopoverForeground).FontSize(14).
			Label(p.Title).Description(p.Description).Absolute().Opacity(progress).Invalid(!valid).Border(1, t.Border)
		radius := t.RadiusXL()
		switch side {
		case Top:
			panel.Top(0).Left(0).Right(0).Height(full).Radius(0, 0, radius, radius)
		case Left:
			panel.Left(0).Top(0).Bottom(0).Width(full).Radius(0, radius, radius, 0)
		case Right:
			panel.Right(0).Top(0).Bottom(0).Width(full).Radius(radius, 0, 0, radius)
		default:
			panel.Bottom(0).Left(0).Right(0).Height(full).Radius(radius, radius, 0, 0)
		}
		if !*open {
			s.offset = drawerPoint(full, side)
		}
		bounds := drawerBounds(full, side)
		axis := ui.Vertical
		if side == Left || side == Right {
			axis = ui.Horizontal
		}
		animation := ui.Animation{Duration: 450 * time.Millisecond, Ease: motion.EaseOut}
		_, panelHeld := panel.PressedButton()
		handleHeld := false
		if s.cancelGesture {
			if !s.cancelStarted {
				panel.Jump("drawer-offset", drawerAxis(s.shown, side))
				s.cancelStarted = true
			}
			s.suspendedFrames++
		}
		var draggable ui.Element
		snap := func(projected ui.Point) ui.Point {
			if !*open {
				return drawerPoint(full, side)
			}
			if draggable.PressCanceled() || !c.WindowActive() {
				return drawerPoint(full-points[s.index], side)
			}
			position := drawerAxis(projected, side)
			smallestOffset := full - points[0]
			if !p.PreventSwipeDismiss && position >= smallestOffset+points[0]*closeThreshold {
				*open = false
				c.Invalidate()
				return drawerPoint(full, side)
			}
			best := 0
			distance := float32(math.Inf(1))
			for i, point := range points {
				if d := float32(math.Abs(float64(position - (full - point)))); d < distance {
					best, distance = i, d
				}
			}
			if s.index != best {
				s.index = best
				if p.ActiveSnap != nil {
					*p.ActiveSnap = best
				}
				if p.OnSnapChange != nil {
					p.OnSnapChange(best)
				}
				c.Invalidate()
			}
			return drawerPoint(full-points[best], side)
		}
		canDrag := valid && !p.DisableSwipe && !s.cancelGesture
		dragEnabled := canDrag && *open
		if !p.HandleOnly && canDrag {
			draggable = panel
			panel.Draggable(&s.offset, ui.DraggableOptions{Axis: axis, Bounds: &bounds, Animation: animation, Snap: snap})
			shownOffset = panel.DragOffset()
		} else if !p.HandleOnly {
			shownOffset = drawerPoint(panel.AnimateWith("drawer-offset", drawerAxis(s.offset, side), 450*time.Millisecond, motion.EaseOut), side)
			panel.Translate(shownOffset.X, shownOffset.Y)
		}
		panel.Children(func() {
			if p.ShowSwipeHandle || p.HandleOnly {
				handle := ui.Box(c).Shrink(0).Center()
				if side == Left || side == Right {
					handle.Absolute().Top(0).Bottom(0).Width(20)
					if side == Left {
						handle.Right(0)
					} else {
						handle.Left(0)
					}
				} else if side == Top {
					handle.Absolute().Bottom(0).Left(0).Right(0).Height(20)
				} else {
					handle.FillWidth().Height(20)
				}
				if p.HandleOnly {
					handle.Label("Resize drawer").Focusable().Role(ui.RoleSlider).Orientation(axis).
						Range(0, float64(len(points)-1), float64(s.index)).Step(1).Value(fmt.Sprintf("Snap %d of %d", s.index+1, len(points)))
				} else {
					handle.PassThrough().AccessHidden(true)
				}
				if p.HandleOnly || s.cancelGesture && s.cancelHandle {
					_, handleHeld = handle.PressedButton()
				}
				if !dragEnabled {
					handle.Disabled(true)
				}
				handle.Children(func() {
					pill := ui.Box(c).Radius(999).Background(t.Muted).AccessHidden(true)
					if side == Left || side == Right {
						pill.Size(6, 100)
					} else {
						pill.Size(100, 6)
					}
				})
				if p.HandleOnly {
					if canDrag {
						draggable = handle
						handle.Draggable(&s.offset, ui.DraggableOptions{Axis: axis, Bounds: &bounds, Animation: animation, Snap: snap})
						shownOffset = handle.DragOffset()
						// Cancel only the handle's own drag transform. The panel
						// receives it below, so the handle follows the pointer once.
						handle.Translate(-shownOffset.X, -shownOffset.Y)
					} else {
						value := drawerAxis(s.offset, side)
						if motion.Reduced(c) {
							value = panel.Jump("drawer-offset", value)
						} else {
							value = panel.AnimateWith("drawer-offset", value, 450*time.Millisecond, motion.EaseOut)
						}
						shownOffset = drawerPoint(value, side)
					}
					panel.Translate(shownOffset.X, shownOffset.Y)
				}
				if dragEnabled && p.HandleOnly {
					next := s.index
					switch {
					case handle.Shortcut(0, ui.KeyUp), handle.Shortcut(0, ui.KeyRight):
						next++
					case handle.Shortcut(0, ui.KeyDown), handle.Shortcut(0, ui.KeyLeft):
						next--
					case handle.Shortcut(0, ui.KeyHome):
						next = 0
					case handle.Shortcut(0, ui.KeyEnd):
						next = len(points) - 1
					}
					next = max(0, min(next, len(points)-1))
					if next != s.index {
						s.index = next
						if p.ActiveSnap != nil {
							*p.ActiveSnap = next
						}
						s.offset = drawerPoint(full-points[next], side)
						if p.OnSnapChange != nil {
							p.OnSnapChange(next)
						}
						c.Invalidate()
					}
				}
			}
			if !p.HideHeader {
				dialogHeader(c, p.Title, p.Description)
			}
			if content != nil {
				ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Label("Drawer content").Children(func() {
					ui.Column(c).FillWidth().Shrink(0).Gap(12).Children(func() {
						stack.rendering = append(stack.rendering, s)
						defer func() { stack.rendering = stack.rendering[:len(stack.rendering)-1] }()
						content()
					})
				})
			}
			if p.Footer != nil {
				modalFooter(c, p.Footer)
			}
			if !p.HideClose {
				modalClose(c, open, "Close drawer", 16)
			}
		})
		if s.cancelGesture {
			s.held = panelHeld
			if s.cancelHandle {
				s.held = handleHeld
			}
			if !s.held && s.suspendedFrames >= 2 && math.Abs(float64(shownOffset.X-s.offset.X))+math.Abs(float64(shownOffset.Y-s.offset.Y)) < .1 {
				s.cancelGesture, s.cancelStarted = false, false
				c.Invalidate()
			} else if s.suspendedFrames < 2 {
				c.Invalidate()
			}
		} else {
			s.held, s.handleDriver = panelHeld, p.HandleOnly
			if p.HandleOnly {
				s.held = handleHeld
			}
		}
		s.shown = shownOffset
		if s.nested > 0 {
			scale := panel.AnimateWith("drawer-stack", max(float32(.8), 1-float32(s.nested)*.05), 450*time.Millisecond, motion.EaseOut)
			panel.Scale(scale, scale)
		} else {
			scale := panel.AnimateWith("drawer-stack", 1, 450*time.Millisecond, motion.EaseOut)
			panel.Scale(scale, scale)
		}
		if s.entering {
			c.Invalidate()
		}
	})
	nowPresent := result.Valid()
	if s.present != nowPresent {
		s.present = nowPresent
		c.Invalidate()
	}
	return result
}
