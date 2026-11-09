package overlay

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestDrawerNativeDragCancelClickAndVelocity(t *testing.T) {
	for _, handleOnly := range []bool{false, true} {
		t.Run(fmt.Sprint("handleOnly=", handleOnly), func(t *testing.T) {
			open, clicks, index, changed := true, 0, 1, -1
			r := ui.NewTester(view(func(c *ui.Context) {
				Drawer(c, &open, DrawerProps{Title: "Drag drawer", Size: 360, HandleOnly: handleOnly, ShowSwipeHandle: true,
					SnapPoints: []DrawerSnapPoint{{Size: 180}, {Size: 360}}, ActiveSnap: &index, OnSnapChange: func(i int) { changed = i }}, func() {
					if ui.Button(c, "Drawer action").Clicked() {
						clicks++
					}
				})
			}), 600, 600)
			r.Advance(time.Second)
			if err := r.Click("Drawer action"); err != nil {
				t.Fatal(err)
			}
			if clicks != 1 || !open || index != 1 {
				t.Fatalf("click was interpreted as a drag: clicks=%d open=%v index=%d", clicks, open, index)
			}
			dragOrigin := func() (float32, float32) {
				if handleOnly {
					h, _ := r.AccessByRole(ui.RoleSlider, "Resize drawer")
					return h.Bounds.X + h.Bounds.W/2, h.Bounds.Y + h.Bounds.H/2
				}
				p, _ := r.AccessByRole(ui.RoleDialog, "Drag drawer")
				return p.Bounds.X + p.Bounds.W/2, p.Bounds.Y + 24
			}
			x, y := dragOrigin()
			r.Press(x, y)
			r.Advance(20 * time.Millisecond)
			r.Move(x, y+120)
			moved, _ := r.AccessByRole(ui.RoleDialog, "Drag drawer")
			if moved.Bounds.Y < 330 {
				t.Fatalf("native drag did not move panel: %+v", moved.Bounds)
			}
			r.CancelPress()
			r.Advance(time.Second)
			settled, _ := r.AccessByRole(ui.RoleDialog, "Drag drawer")
			if !open || index != 1 || changed != -1 || math.Abs(float64(settled.Bounds.Y-240)) > 1 {
				t.Fatalf("cancel changed snap/dismissed: open=%v index=%d callback=%d box=%+v", open, index, changed, settled.Bounds)
			}
			// Window blur cancels native capture just like pointercancel.
			x, y = dragOrigin()
			r.Press(x, y)
			r.Move(x, y+110)
			r.SetFocused(false)
			r.Advance(time.Second)
			settled, _ = r.AccessByRole(ui.RoleDialog, "Drag drawer")
			if !open || index != 1 || changed != -1 || math.Abs(float64(settled.Bounds.Y-240)) > 1 {
				t.Fatal("window blur did not cancel native drag")
			}
			r.SetFocused(true)
			// A quick flick projects past the dismissal threshold, although
			// the actual displacement is less than a snap interval.
			x, y = dragOrigin()
			r.Press(x, y)
			r.Advance(10 * time.Millisecond)
			r.Move(x, y+80)
			r.Advance(10 * time.Millisecond)
			r.Release(x, y+100)
			if open {
				t.Fatal("fast native flick was not projected to dismissal")
			}
			r.Advance(time.Second)
			if _, ok := r.AccessByRole(ui.RoleDialog, "Drag drawer"); ok || r.FramePending() {
				t.Fatal("dismissed drawer retained content or animation frames")
			}
		})
	}
}

func TestDrawerNonModalLetsBackgroundReceiveOutsidePress(t *testing.T) {
	open, actions, changes := true, 0, 0
	r := ui.NewTester(view(func(c *ui.Context) {
		if ui.Button(c, "Background action").Clicked() {
			actions++
		}
		Drawer(c, &open, DrawerProps{Title: "Nonmodal", Size: 240, NonModal: true, DisableSwipe: true,
			OnOpenChange: func(bool) { changes++ }}, func() { ui.Button(c, "Nonmodal action") })
	}), 600, 600)
	r.Advance(time.Second)
	if err := r.Click("Background action"); err != nil {
		t.Fatal(err)
	}
	if actions != 1 || open || changes != 1 {
		t.Fatalf("outside press not delivered/dismissed once: actions=%d open=%v changes=%d", actions, open, changes)
	}
	r.Advance(time.Second)
	if _, ok := r.AccessByRole(ui.RoleDialog, "Nonmodal"); ok || r.FramePending() {
		t.Fatal("nonmodal drawer retained a panel or frame")
	}
}

func TestDrawerSnapPointsControlledKeyboardAndInvalidArrays(t *testing.T) {
	open, index, callbacks := true, 0, 0
	points := []DrawerSnapPoint{{Size: 120}, {Fraction: .5}, {Fraction: 1}}
	r := ui.NewTester(view(func(c *ui.Context) {
		Drawer(c, &open, DrawerProps{Title: "Snaps", HandleOnly: true, SnapPoints: points, ActiveSnap: &index,
			OnSnapChange: func(int) { callbacks++ }}, func() { ui.Button(c, "Snap action") })
	}), 600, 600)
	r.Advance(time.Second)
	d, _ := r.AccessByRole(ui.RoleDialog, "Snaps")
	if math.Abs(float64(d.Bounds.Y-480)) > 1 {
		t.Fatalf("DIP snap visible extent is wrong: %+v", d.Bounds)
	}
	index = 1
	r.Frame()
	r.Advance(time.Second)
	d, _ = r.AccessByRole(ui.RoleDialog, "Snaps")
	if callbacks != 0 || math.Abs(float64(d.Bounds.Y-300)) > 1 {
		t.Fatalf("controlled fraction failed or reported user callback: callbacks=%d box=%+v", callbacks, d.Bounds)
	}
	if err := r.Click("Resize drawer"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyEnd)
	r.Advance(time.Second)
	d, _ = r.AccessByRole(ui.RoleDialog, "Snaps")
	if index != 2 || callbacks != 1 || math.Abs(float64(d.Bounds.Y-16)) > 1 {
		t.Fatalf("handle keyboard snap failed: index=%d callbacks=%d box=%+v", index, callbacks, d.Bounds)
	}
	// Invalid arrays remain visibly usable and marked invalid, but never
	// silently reorder points or write an unrelated controlled index.
	points = []DrawerSnapPoint{{Size: 200}, {Size: 100}}
	index = 7
	r.Frame()
	r.Advance(time.Second)
	d, _ = r.AccessByRole(ui.RoleDialog, "Snaps")
	h, _ := r.AccessByRole(ui.RoleSlider, "Resize drawer")
	if !d.Invalid || !h.Disabled || index != 7 {
		t.Fatalf("invalid snap array silently normalized: dialog=%+v handle=%+v index=%d", d, h, index)
	}
}

func TestDrawerReducedMotionAndSides(t *testing.T) {
	for _, side := range []Placement{Bottom, Top, Left, Right} {
		t.Run(fmt.Sprint(side), func(t *testing.T) {
			open := false
			r := ui.NewTester(view(func(c *ui.Context) {
				Drawer(c, &open, DrawerProps{Title: "Side drawer", Side: side, Size: 240, HandleOnly: true}, func() { ui.Button(c, "Side action") })
			}), 600, 500)
			r.SetPreferences(ui.Preferences{ReduceMotion: true})
			open = true
			r.Frame()
			d, ok := r.AccessByRole(ui.RoleDialog, "Side drawer")
			if !ok || r.FramePending() {
				t.Fatal("reduced-motion opening did not settle immediately")
			}
			handle, _ := r.AccessByRole(ui.RoleSlider, "Resize drawer")
			if (side == Top || side == Bottom) && !handle.Vertical || (side == Left || side == Right) && !handle.Horizontal {
				t.Fatalf("handle accessibility orientation differs from native drag axis: vertical=%v horizontal=%v", handle.Vertical, handle.Horizontal)
			}
			wantX, wantY := float32(0), float32(0)
			if side == Bottom {
				wantY = 260
			} else if side == Right {
				wantX = 360
			}
			if math.Abs(float64(d.Bounds.X-wantX))+math.Abs(float64(d.Bounds.Y-wantY)) > 1 {
				t.Fatalf("wrong physical side: %+v", d.Bounds)
			}
			open = false
			r.Frame()
			if _, ok := r.AccessByRole(ui.RoleDialog, "Side drawer"); ok || r.FramePending() {
				t.Fatal("reduced-motion closing retained a panel or frame")
			}
		})
	}
}

func TestDrawerScrollBoundariesDoNotDragOrDismiss(t *testing.T) {
	open, actions := true, 0
	r := ui.NewTester(view(func(c *ui.Context) {
		Drawer(c, &open, DrawerProps{Title: "Scrollable drawer", Size: 360, ShowSwipeHandle: true}, func() {
			for i := 0; i < 40; i++ {
				ui.Text(c, fmt.Sprintf("Drawer row %02d", i)).Height(30)
			}
			if ui.Button(c, "Final drawer action").Clicked() {
				actions++
			}
		})
	}), 600, 600)
	r.Advance(time.Second)
	d, _ := r.AccessByRole(ui.RoleDialog, "Scrollable drawer")
	body, _ := r.Access("Drawer content")
	x, y := body.Bounds.X+body.Bounds.W/2, body.Bounds.Y+body.Bounds.H/2
	// Overscroll at each end keeps the panel anchored and preserves content
	// scrolling. Native child scroll never hands its pointer gesture to a
	// parent Draggable; a swipe begins on the handle or panel background.
	r.Scroll(x, y, 0, -2000)
	r.Advance(500 * time.Millisecond)
	top, _ := r.AccessByRole(ui.RoleDialog, "Scrollable drawer")
	if !open || top.Bounds != d.Bounds {
		t.Fatal("scrolling beyond the top moved/dismissed drawer")
	}
	r.Scroll(x, y, 0, 4000)
	r.Advance(500 * time.Millisecond)
	bottom, _ := r.AccessByRole(ui.RoleDialog, "Scrollable drawer")
	last, _ := r.AccessByRole(ui.RoleButton, "Final drawer action")
	if !open || bottom.Bounds != d.Bounds || last.Bounds.Y < body.Bounds.Y || last.Bounds.Y+last.Bounds.H > body.Bounds.Y+body.Bounds.H+1 {
		t.Fatalf("bottom scrolling moved panel or did not reveal final action: drawer=%+v body=%+v final=%+v", bottom.Bounds, body.Bounds, last.Bounds)
	}
	if err := r.Click("Final drawer action"); err != nil {
		t.Fatal(err)
	}
	if actions != 1 || !open {
		t.Fatal("scrolled content action lost to drawer gesture")
	}
}

func TestDrawerNestedAndExitReversal(t *testing.T) {
	outer, inner := true, false
	r := ui.NewTester(view(func(c *ui.Context) {
		Drawer(c, &outer, DrawerProps{Title: "Outer drawer", Size: 420, HandleOnly: true}, func() {
			ui.Button(c, "Outer drawer action")
			Drawer(c, &inner, DrawerProps{Title: "Inner drawer", Size: 240, HandleOnly: true}, func() { ui.Button(c, "Inner drawer action") })
		})
	}), 600, 600)
	r.Advance(time.Second)
	inner = true
	r.Frame()
	r.Advance(time.Second)
	for i := 0; i < 6; i++ {
		r.Key(0, ui.KeyTab)
		if r.Focused("Outer drawer action") {
			t.Fatal("Tab escaped nested drawer")
		}
	}
	inner = false
	r.Frame()
	r.Advance(100 * time.Millisecond)
	inner = true
	r.Frame()
	r.Advance(time.Second)
	if err := r.Click("Inner drawer action"); err != nil {
		t.Fatal(err)
	}
	if !r.Focused("Inner drawer action") || r.FramePending() {
		t.Fatal("reopened nested drawer remained inert or animated indefinitely")
	}
	inner = false
	r.Frame()
	r.Advance(time.Second)
	d, ok := r.AccessByRole(ui.RoleDialog, "Outer drawer")
	if !ok || math.Abs(float64(d.Bounds.Y-180)) > 1 || r.FramePending() {
		t.Fatalf("parent did not unstack after nested removal: %+v", d.Bounds)
	}
}

func TestDrawerReconfigurationCancelsCapturedNativeDrag(t *testing.T) {
	for _, change := range []string{"disable", "handle", "panel", "snap", "controlled"} {
		t.Run(change, func(t *testing.T) {
			open, disabled, handleOnly, index, callbacks := true, false, false, 1, 0
			if change == "panel" {
				handleOnly = true
			}
			draft, draftInitializations := "Keep this draft", 0
			points := []DrawerSnapPoint{{Size: 180}, {Size: 360}}
			r := ui.NewTester(view(func(c *ui.Context) {
				Drawer(c, &open, DrawerProps{Title: "Dynamic drawer", Size: 360, ShowSwipeHandle: true, DisableSwipe: disabled, HandleOnly: handleOnly,
					SnapPoints: points, ActiveSnap: &index, OnSnapChange: func(int) { callbacks++ }}, func() {
					editor := ui.TextInputBase(c, &draft).Label("Dynamic draft").Height(32)
					ui.Local(editor, "draft-state", func() int { draftInitializations++; return 0 })
					ui.Button(c, "Dynamic action")
				})
			}), 600, 600)
			r.Advance(time.Second)
			d, _ := r.AccessByRole(ui.RoleDialog, "Dynamic drawer")
			x, y := d.Bounds.X+d.Bounds.W/2, d.Bounds.Y+8
			if handleOnly {
				h, _ := r.AccessByRole(ui.RoleSlider, "Resize drawer")
				x, y = h.Bounds.X+h.Bounds.W/2, h.Bounds.Y+h.Bounds.H/2
			}
			r.Press(x, y)
			r.Advance(20 * time.Millisecond)
			r.Move(x, y+120)
			switch change {
			case "disable":
				disabled = true
			case "handle":
				handleOnly = true
			case "panel":
				handleOnly = false
			case "snap":
				points = []DrawerSnapPoint{{Size: 150}, {Size: 300}}
			case "controlled":
				index = 0
			}
			r.Frame()
			// Re-enable before release as well: old native capture must not
			// resume when the old driver becomes eligible again.
			if change == "disable" {
				disabled = false
				r.Frame()
			}
			r.Advance(20 * time.Millisecond)
			r.Move(x, y+230)
			r.Release(x, y+240)
			r.Advance(time.Second)
			d, _ = r.AccessByRole(ui.RoleDialog, "Dynamic drawer")
			wantY := float32(240)
			if change == "snap" {
				wantY = 300
			}
			wantIndex := 1
			if change == "controlled" {
				wantY, wantIndex = 420, 0
			}
			if !open || index != wantIndex || callbacks != 0 || math.Abs(float64(d.Bounds.Y-wantY)) > 1 || r.FramePending() {
				t.Fatalf("stale captured drag survived configuration change: open=%v index=%d callbacks=%d box=%+v", open, index, callbacks, d.Bounds)
			}
			if draftInitializations != 1 || draft != "Keep this draft" {
				t.Fatalf("gesture reconfiguration discarded native editor state: initializations=%d value=%q", draftInitializations, draft)
			}
		})
	}
}
