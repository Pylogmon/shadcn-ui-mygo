package overlay

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

type toastSwipeKey struct{}

// Only model and numeric event data survive a build. Native press ownership
// keeps buttons and selectable custom content out of the toast's gesture.
type toastSwipeRuntime struct {
	entry              *toastEntry
	version            uint64
	down               ui.Point
	downAt             time.Time
	total              ui.Point
	blocked, observed  bool
	returnAt           time.Time
	returnFrom         ui.Point
	outAt              time.Time
	outFrom, outTravel ui.Point
}

func (s *toastSwipeRuntime) Close() error {
	if s.entry != nil {
		s.entry.drag, s.entry.swiping, s.entry.swipeAxis = ui.Point{}, false, 0
		s.entry.swipeExit = false
	}
	return nil
}

var toastSwipeReturnEase = ui.CubicBezier(.25, .1, .25, 1)
var toastSwipeExitEase = ui.CubicBezier(0, 0, .58, 1)

func toastSwipe(c *ui.Context, root ui.Element, entry *toastEntry, props ToasterProps, position ToastPosition) {
	if entry == nil {
		return
	}
	s := ui.Local(root, toastSwipeKey{}, func() toastSwipeRuntime {
		return toastSwipeRuntime{entry: entry, version: entry.version}
	})
	if s.entry != entry || s.version != entry.version {
		s.Close()
		*s = toastSwipeRuntime{entry: entry, version: entry.version, blocked: true}
	}
	root.ObserveInput(func(ev ui.InputEvent) {
		if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
			s.down, s.downAt = ui.Point{X: ev.X, Y: ev.Y}, ev.Time
			s.observed, s.blocked = true, false
		}
	})
	// Dragged uses native capture, continues outside the element/window, and
	// reports only the actual press target. A child button retains its press.
	dx, dy, held := root.Dragged()
	canceled := root.PressCanceled()
	if !s.outAt.IsZero() {
		progress := toastSwipeProgress(c, s.outAt, 200*time.Millisecond)
		eased := toastSwipeExitEase(progress)
		entry.drag = ui.Point{X: s.outFrom.X + s.outTravel.X*eased, Y: s.outFrom.Y + s.outTravel.Y*eased}
		return
	}
	eligible := entry.model.Type != "loading" && toastBool(entry.options.Dismissible, true) &&
		entry.pending == toastNone && entry.closing.IsZero() && !root.IsDisabled()
	if !eligible {
		entry.drag, entry.swiping, entry.swipeAxis = ui.Point{}, false, 0
		s.total, s.returnAt, s.blocked = ui.Point{}, time.Time{}, held
		return
	}
	if canceled {
		entry.swiping, entry.swipeAxis = false, 0
		s.total, s.observed = ui.Point{}, false
		s.returnFrom, s.returnAt = entry.drag, c.Now()
	} else if held {
		if s.blocked {
			return
		}
		if !entry.swiping {
			entry.swiping, entry.swipeAxis = true, 0
			entry.swipeStart, entry.swipeAt = s.down, s.downAt
			if !s.observed {
				entry.swipeAt = c.Now()
			}
			s.total, s.returnAt = ui.Point{}, time.Time{}
		}
		s.total.X += dx
		s.total.Y += dy
		if entry.swipeAxis == 0 && (math.Abs(float64(s.total.X)) > 1 || math.Abs(float64(s.total.Y)) > 1) {
			entry.swipeAxis = 2
			if math.Abs(float64(s.total.X)) > math.Abs(float64(s.total.Y)) {
				entry.swipeAxis = 1
			}
		}
		entry.drag = toastSwipeAmount(s.total, entry.swipeAxis, toastSwipeDirections(props, position))
		return
	} else if entry.swiping {
		entry.swiping = false
		amount := entry.drag.Y
		if entry.swipeAxis == 1 {
			amount = entry.drag.X
		}
		milliseconds := float64(c.Now().Sub(entry.swipeAt)) / float64(time.Millisecond)
		velocity := float64(0)
		if milliseconds > 0 {
			velocity = math.Abs(float64(amount)) / milliseconds
		} else if amount != 0 {
			velocity = math.Inf(1)
		}
		if math.Abs(float64(amount)) >= 45 || velocity > .11 {
			entry.swipeExit = true
			s.outFrom, s.outAt = entry.drag, c.Now()
			bounds := root.Bounds()
			if entry.swipeAxis == 1 {
				s.outTravel.X = float32(math.Copysign(float64(bounds.W), float64(amount)))
			} else {
				s.outTravel.Y = float32(math.Copysign(float64(bounds.H), float64(amount)))
			}
			entry.pending = toastDismiss
			c.Invalidate()
			toastSwipeProgress(c, s.outAt, 200*time.Millisecond)
			return
		}
		entry.swipeAxis = 0
		s.returnFrom, s.returnAt = entry.drag, c.Now()
	}
	if !s.returnAt.IsZero() {
		progress := toastSwipeProgress(c, s.returnAt, 400*time.Millisecond)
		amount := 1 - toastSwipeReturnEase(progress)
		entry.drag = ui.Point{X: s.returnFrom.X * amount, Y: s.returnFrom.Y * amount}
		if progress == 1 {
			s.returnAt = time.Time{}
		}
	}
}

func toastSwipeProgress(c *ui.Context, start time.Time, duration time.Duration) float32 {
	if motion.Reduced(c) {
		return 1
	}
	progress := min(float32(1), max(float32(0), float32(c.Now().Sub(start))/float32(duration)))
	if progress < 1 {
		c.AnimationFrame()
	}
	return progress
}

func toastSwipeDirections(props ToasterProps, position ToastPosition) []ToastSwipeDirection {
	if props.SwipeDirections != nil {
		return props.SwipeDirections
	}
	switch toastValidPosition(position) {
	case ToastTopLeft:
		return []ToastSwipeDirection{ToastSwipeTop, ToastSwipeLeft}
	case ToastTopCenter:
		return []ToastSwipeDirection{ToastSwipeTop}
	case ToastTopRight:
		return []ToastSwipeDirection{ToastSwipeTop, ToastSwipeRight}
	case ToastBottomLeft:
		return []ToastSwipeDirection{ToastSwipeBottom, ToastSwipeLeft}
	case ToastBottomCenter:
		return []ToastSwipeDirection{ToastSwipeBottom}
	default:
		return []ToastSwipeDirection{ToastSwipeBottom, ToastSwipeRight}
	}
}

func toastSwipeAmount(total ui.Point, axis int, directions []ToastSwipeDirection) ui.Point {
	var top, right, bottom, left bool
	for _, direction := range directions {
		switch direction {
		case ToastSwipeTop:
			top = true
		case ToastSwipeRight:
			right = true
		case ToastSwipeBottom:
			bottom = true
		case ToastSwipeLeft:
			left = true
		}
	}
	dampen := func(amount float32, allowed bool) float32 {
		if allowed {
			return amount
		}
		return amount / (1.5 + float32(math.Abs(float64(amount)))/20)
	}
	if axis == 1 && (left || right) {
		return ui.Point{X: dampen(total.X, total.X < 0 && left || total.X > 0 && right)}
	}
	if axis == 2 && (top || bottom) {
		return ui.Point{Y: dampen(total.Y, total.Y < 0 && top || total.Y > 0 && bottom)}
	}
	return ui.Point{}
}
