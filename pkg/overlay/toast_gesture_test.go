package overlay

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func toastGestureTester(entry *toastEntry, props *ToasterProps, position *ToastPosition, actions ...func()) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		root := ui.Box(c).Absolute().Left(100).Top(100).Size(200, 80).Label("Swipe toast")
		toastSwipe(c, root, entry, *props, *position)
		root.Translate(entry.drag.X, entry.drag.Y).Children(func() {
			if ui.Button(c, "Toast action").Size(80, 30).Clicked() && len(actions) > 0 {
				actions[0]()
			}
		})
	}, 600, 400)
}

func TestToastSwipeDistanceAndVelocityBoundaries(t *testing.T) {
	for _, test := range []struct {
		name     string
		distance float32
		elapsed  time.Duration
		dismiss  bool
	}{
		{"short slow", 44, time.Second, false},
		{"distance exactly 45", 45, time.Second, true},
		{"velocity exactly point11", 11, 100 * time.Millisecond, false},
		{"velocity above point11", 12, 100 * time.Millisecond, true},
		{"quick short flick", 10, 50 * time.Millisecond, true},
		{"no elapsed time", 3, 0, true},
		{"axis dead zone", 1, 0, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
			props, position := ToasterProps{}, ToastBottomRight
			r := toastGestureTester(entry, &props, &position)
			r.Press(220, 150)
			r.Advance(test.elapsed)
			r.Move(220+test.distance, 150)
			r.Release(220+test.distance, 150)
			if got := entry.pending == toastDismiss; got != test.dismiss {
				t.Fatalf("dismiss=%v want=%v drag=%v", got, test.dismiss, entry.drag)
			}
			if !test.dismiss {
				r.Advance(400 * time.Millisecond)
				if entry.drag != (ui.Point{}) || entry.swiping || r.FramePending() {
					t.Fatalf("short swipe failed to settle: drag=%v swiping=%v pendingFrame=%v", entry.drag, entry.swiping, r.FramePending())
				}
			}
		})
	}
}

func TestToastSwipeDefaultDirectionsArePhysicalInLTRAndRTL(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		for _, test := range []struct {
			position ToastPosition
			delta    ui.Point
		}{
			{ToastTopLeft, ui.Point{X: -60}},
			{ToastTopLeft, ui.Point{Y: -60}},
			{ToastTopCenter, ui.Point{Y: -60}},
			{ToastTopRight, ui.Point{X: 60}},
			{ToastTopRight, ui.Point{Y: -60}},
			{ToastBottomLeft, ui.Point{X: -60}},
			{ToastBottomLeft, ui.Point{Y: 60}},
			{ToastBottomCenter, ui.Point{Y: 60}},
			{ToastBottomRight, ui.Point{X: 60}},
			{ToastBottomRight, ui.Point{Y: 60}},
		} {
			t.Run(string(test.position)+"/"+map[ui.Direction]string{ui.LeftToRight: "LTR", ui.RightToLeft: "RTL"}[direction]+"/"+map[bool]string{true: "horizontal", false: "vertical"}[test.delta.X != 0], func(t *testing.T) {
				entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
				props, position := ToasterProps{}, test.position
				r := toastGestureTester(entry, &props, &position)
				r.SetDirection(direction)
				r.Press(220, 150)
				r.Advance(time.Second)
				r.Move(220+test.delta.X, 150+test.delta.Y)
				if entry.drag != test.delta {
					t.Fatalf("physical direction reversed: drag=%v want=%v", entry.drag, test.delta)
				}
				r.Release(220+test.delta.X, 150+test.delta.Y)
				if entry.pending != toastDismiss {
					t.Fatal("default outward swipe did not dismiss")
				}
			})
		}
	}
}

func TestToastSwipeOppositeDirectionUsesSourceDampingAndResets(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomRight
	r := toastGestureTester(entry, &props, &position)
	r.Press(220, 150)
	r.Advance(time.Second)
	r.Move(120, 150)
	want := float32(-100) / 6.5
	if math.Abs(float64(entry.drag.X-want)) > .001 || entry.drag.Y != 0 {
		t.Fatalf("opposite swipe did not use source damping: got=%v wantX=%v", entry.drag, want)
	}
	r.Release(120, 150)
	if entry.pending != toastNone {
		t.Fatal("slow damped reverse swipe dismissed")
	}
	r.Advance(200 * time.Millisecond)
	if entry.drag.X <= want || entry.drag.X >= 0 || !r.FramePending() {
		t.Fatalf("source400ms return did not interpolate: drag=%v pendingFrame=%v", entry.drag, r.FramePending())
	}
	r.Advance(200 * time.Millisecond)
	if entry.drag != (ui.Point{}) || r.FramePending() {
		t.Fatal("opposite swipe did not finish returning")
	}
}

func TestToastSwipeAxisLocksAtFirstMovement(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomRight
	r := toastGestureTester(entry, &props, &position)
	r.Press(220, 150)
	r.Move(223, 152)
	if entry.swipeAxis != 1 || entry.drag != (ui.Point{X: 3}) {
		t.Fatalf("first movement did not lockX: axis=%d drag=%v", entry.swipeAxis, entry.drag)
	}
	r.Advance(time.Second)
	r.Move(230, 250)
	if entry.drag != (ui.Point{X: 10}) || entry.swipeAxis != 1 {
		t.Fatalf("later dominantY changed locked axis: axis=%d drag=%v", entry.swipeAxis, entry.drag)
	}
	r.Release(230, 250)
	if entry.pending != toastNone {
		t.Fatal("unlocked vertical delta incorrectly dismissed")
	}
}

func TestToastSwipeExplicitDirectionsAndEmptyDirections(t *testing.T) {
	for _, test := range []struct {
		name       string
		directions []ToastSwipeDirection
		delta      ui.Point
		want       ui.Point
	}{
		{"override", []ToastSwipeDirection{ToastSwipeLeft}, ui.Point{X: -60}, ui.Point{X: -60}},
		{"empty disables", []ToastSwipeDirection{}, ui.Point{X: 60}, ui.Point{}},
		{"vertical excludes horizontal", []ToastSwipeDirection{ToastSwipeTop}, ui.Point{X: 60}, ui.Point{}},
		{"center horizontal unavailable", nil, ui.Point{X: 60}, ui.Point{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
			props, position := ToasterProps{SwipeDirections: test.directions}, ToastBottomCenter
			r := toastGestureTester(entry, &props, &position)
			r.Press(220, 150)
			r.Advance(time.Second)
			r.Move(220+test.delta.X, 150+test.delta.Y)
			if entry.drag != test.want {
				t.Fatalf("directions drag=%v want=%v", entry.drag, test.want)
			}
			r.Release(220+test.delta.X, 150+test.delta.Y)
			if (entry.pending == toastDismiss) != (test.want != (ui.Point{})) {
				t.Fatalf("directions dismissal=%v", entry.pending)
			}
		})
	}
}

func TestToastSwipeDoesNotStartOnNativeChildButton(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomRight
	clicked := 0
	r := toastGestureTester(entry, &props, &position, func() { clicked++ })
	button, ok := r.Find("Toast action")
	if !ok {
		t.Fatal("missing action")
	}
	x, y := button.X+button.W/2, button.Y+button.H/2
	r.Press(x, y)
	r.Move(x+300, y)
	r.Release(x+300, y)
	if entry.drag != (ui.Point{}) || entry.pending != toastNone || entry.swiping || clicked != 0 {
		t.Fatalf("child button began toast swipe: drag=%v pending=%v swiping=%v clicks=%d", entry.drag, entry.pending, entry.swiping, clicked)
	}
	if err := r.Click("Toast action"); err != nil {
		t.Fatal(err)
	}
	if clicked != 1 || entry.pending != toastNone {
		t.Fatal("gesture registration suppressed native action click")
	}
}

func TestToastSwipeSelectableContentKeepsNativeTextSelection(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	start, end := 0, 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		root := ui.Box(c).Absolute().Left(100).Top(100).Size(200, 80)
		toastSwipe(c, root, entry, ToasterProps{}, ToastBottomRight)
		root.Translate(entry.drag.X, entry.drag.Y).Children(func() {
			text := ui.Text(c, "Selectable toast details").Selectable().Width(200).Height(30)
			start, end = text.TextSelection()
		})
	}, 600, 400)
	text, ok := r.Find("Selectable toast details")
	if !ok {
		t.Fatal("missing selectable custom content")
	}
	r.Press(text.X+2, text.Y+10)
	r.Move(text.X+100, text.Y+10)
	r.Release(text.X+100, text.Y+10)
	if end <= start || entry.drag != (ui.Point{}) || entry.pending != toastNone || entry.swiping {
		t.Fatalf("toast stole native selection: selection=%d:%d drag=%v pending=%v swiping=%v", start, end, entry.drag, entry.pending, entry.swiping)
	}
}

func TestToastSwipeSecondaryButtonDoesNotStart(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomRight
	r := toastGestureTester(entry, &props, &position)
	r.PressButton(1, 220, 150)
	r.Move(500, 150)
	r.ReleaseButton(1, 500, 150)
	if entry.drag != (ui.Point{}) || entry.pending != toastNone || entry.swiping {
		t.Fatalf("secondary press swiped: drag=%v pending=%v swiping=%v", entry.drag, entry.pending, entry.swiping)
	}
}

func TestToastSwipeLoadingAndNonDismissibleDoNotStart(t *testing.T) {
	disabled := false
	for _, test := range []struct {
		name  string
		entry toastEntry
	}{
		{"loading", toastEntry{model: ui.Toast{ID: "loading", Type: "loading"}}},
		{"non dismissible", toastEntry{model: ui.Toast{ID: "fixed"}, options: ToastOptions{Dismissible: &disabled}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			entry := test.entry
			props, position := ToasterProps{}, ToastBottomRight
			r := toastGestureTester(&entry, &props, &position)
			r.Press(220, 150)
			r.Move(500, 150)
			r.Release(500, 150)
			if entry.drag != (ui.Point{}) || entry.pending != toastNone || entry.swiping {
				t.Fatalf("ineligible toast swiped: drag=%v pending=%v swiping=%v", entry.drag, entry.pending, entry.swiping)
			}
		})
	}
}

func TestToastSwipeCancelAndWindowBlurNeverDismiss(t *testing.T) {
	for _, blur := range []bool{false, true} {
		t.Run(map[bool]string{false: "pointer cancel", true: "window blur"}[blur], func(t *testing.T) {
			entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
			props, position := ToasterProps{}, ToastBottomRight
			r := toastGestureTester(entry, &props, &position)
			r.Press(220, 150)
			r.Advance(100 * time.Millisecond)
			r.Move(320, 150)
			if blur {
				r.SetFocused(false)
			} else {
				r.CancelPress()
			}
			if entry.pending != toastNone || entry.swiping {
				t.Fatalf("canceled capture dismissed: pending=%v swiping=%v", entry.pending, entry.swiping)
			}
			r.Advance(400 * time.Millisecond)
			if entry.drag != (ui.Point{}) || r.FramePending() {
				t.Fatal("canceled capture did not return to rest")
			}
		})
	}
}

func TestToastSwipeUpdateCancelsOldCaptureUntilNextPress(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomRight
	r := toastGestureTester(entry, &props, &position)
	r.Press(220, 150)
	r.Move(250, 150)
	entry.version++
	r.Frame()
	r.Move(400, 150)
	r.Release(400, 150)
	if entry.drag != (ui.Point{}) || entry.pending != toastNone || entry.swiping {
		t.Fatal("old capture affected updated notification")
	}
	r.Press(220, 150)
	r.Move(280, 150)
	r.Release(280, 150)
	if entry.pending != toastDismiss {
		t.Fatal("fresh press did not recover after update")
	}
}

func TestToastSwipeReducedMotionReturnIsImmediate(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomRight
	r := toastGestureTester(entry, &props, &position)
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	r.Press(220, 150)
	r.Advance(time.Second)
	r.Move(250, 150)
	r.Release(250, 150)
	if entry.drag != (ui.Point{}) || entry.pending != toastNone || r.FramePending() {
		t.Fatalf("reduced return moved: drag=%v pending=%v frame=%v", entry.drag, entry.pending, r.FramePending())
	}
}

func TestToastSwipeExitMovesOneToastWidthInPhysicalDirection(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomLeft
	r := toastGestureTester(entry, &props, &position)
	r.Press(220, 150)
	r.Advance(time.Second)
	r.Move(160, 150)
	r.Release(160, 150)
	if entry.drag != (ui.Point{X: -60}) || entry.pending != toastDismiss {
		t.Fatalf("wrong swipe exit start: drag=%v pending=%v", entry.drag, entry.pending)
	}
	r.Advance(100 * time.Millisecond)
	if entry.drag.X >= -60 || entry.drag.X <= -260 || entry.drag.Y != 0 {
		t.Fatalf("exit did not interpolate left: %v", entry.drag)
	}
	r.Advance(100 * time.Millisecond)
	if entry.drag != (ui.Point{X: -260}) || r.FramePending() {
		t.Fatalf("wrong exit destination or scheduling: drag=%v frame=%v", entry.drag, r.FramePending())
	}
}

func TestToastSwipeUnmountClearsModelGesture(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	mounted := true
	r := ui.NewTester(func(c *ui.Context) {
		if mounted {
			root := ui.Box(c).Absolute().Left(100).Top(100).Size(200, 80)
			toastSwipe(c, root, entry, ToasterProps{}, ToastBottomRight)
			root.Translate(entry.drag.X, entry.drag.Y)
		}
	}, 600, 400)
	r.Press(220, 150)
	r.Move(250, 150)
	mounted = false
	r.Frame()
	if entry.drag != (ui.Point{}) || entry.swiping || entry.pending != toastNone {
		t.Fatalf("unmount left captured model state: drag=%v swiping=%v pending=%v", entry.drag, entry.swiping, entry.pending)
	}
}

func TestToastSwipeRendererDismissalCallbacksAndExitGeometry(t *testing.T) {
	state := &ToasterState{}
	closed, dismissed := 0, 0
	state.Show(ui.Toast{ID: "swipe", Title: "Swiped notification", Timeout: -1, OnClose: func() { closed++ }}, ToastOptions{OnDismiss: func(ui.Toast) { dismissed++ }})
	entry := state.entries[0]
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ToasterWith(c, ToasterProps{State: state, Position: ToastBottomLeft})
	}, 800, 500)
	r.Advance(500 * time.Millisecond)
	before, ok := r.AccessByRole(ui.RoleListItem, "Swiped notification")
	if !ok {
		t.Fatal("renderer did not expose application notification")
	}
	x, y := before.Bounds.X+before.Bounds.W/2, before.Bounds.Y+before.Bounds.H/2
	r.Press(x, y)
	r.Advance(time.Second)
	r.Move(x-60, y)
	r.Release(x-60, y)
	if closed != 1 || dismissed != 1 || entry.closing.IsZero() || !entry.swipeExit || len(state.Toasts()) != 0 {
		t.Fatalf("renderer failed to accept swipe once: close=%d dismiss=%d closing=%v swipeExit=%v live=%v", closed, dismissed, entry.closing, entry.swipeExit, state.Toasts())
	}
	r.Advance(100 * time.Millisecond)
	if len(state.entries) != 1 || entry.drag.X >= -60 || entry.drag.X <= -60-before.Bounds.W || entry.drag.Y != 0 {
		t.Fatalf("renderer lost interpolated horizontal exit: entries=%d drag=%v", len(state.entries), entry.drag)
	}
	if during, ok := r.AccessByRole(ui.RoleListItem, "Swiped notification"); ok && math.Abs(float64(during.Bounds.Y-before.Bounds.Y)) > .1 {
		t.Fatalf("renderer added ordinary vertical exit to swipe: before=%v during=%v", before.Bounds, during.Bounds)
	}
	// Native timers have a minimum 1ms wake. The last 60Hz frame may land
	// a few nanoseconds before 200ms, so permit that final timer quantum.
	r.Advance(101 * time.Millisecond)
	if len(state.entries) != 0 || closed != 1 || dismissed != 1 {
		t.Fatalf("renderer did not finish source200ms lifecycle once: entries=%d close=%d dismiss=%d age=%v drag=%v frame=%v", len(state.entries), closed, dismissed, r.Now().Sub(entry.closing), entry.drag, r.FramePending())
	}
}

func TestToastSwipeDistanceThresholdCapturesOutsideAndDismisses(t *testing.T) {
	entry := &toastEntry{model: ui.Toast{ID: "swipe"}, version: 1}
	props, position := ToasterProps{}, ToastBottomRight
	r := toastGestureTester(entry, &props, &position)
	r.Press(220, 150)
	r.Advance(time.Second)
	r.Move(420, 150)
	if entry.drag != (ui.Point{X: 200}) || !entry.swiping {
		t.Fatalf("native capture did not move outside root: drag=%v swiping=%v", entry.drag, entry.swiping)
	}
	r.Release(420, 150)
	if entry.pending != toastDismiss || entry.swiping {
		t.Fatalf("completed swipe not dismissed: pending=%v swiping=%v", entry.pending, entry.swiping)
	}
}
