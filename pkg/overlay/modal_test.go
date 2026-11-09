package overlay

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestDialogClosingContentLeavesInputAndAccessibilityImmediately(t *testing.T) {
	open := true
	actions := 0
	r := ui.NewTester(view(func(c *ui.Context) {
		Dialog(c, &open, DialogProps{Title: "Publish"}, func() {
			if ui.Button(c, "Publish now").Clicked() {
				actions++
			}
		})
	}), 600, 420)
	r.Advance(300 * time.Millisecond)
	open = false
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleButton, "Publish now"); ok {
		t.Fatal("closing dialog still exposes an actionable button")
	}
	r.Key(0, ui.KeyEnter)
	if actions != 0 {
		t.Fatal("closing dialog accepted an action")
	}
}

func TestDialogNestedFocusInitialRestoreAndClosingTrap(t *testing.T) {
	outer, inner, actions := false, false, 0
	var first, innerFirst, innerRestore, restore ui.Handle
	r := ui.NewTester(view(func(c *ui.Context) {
		if ui.Button(c, "Open outer").Clicked() {
			outer = true
		}
		if ui.Button(c, "Restore here").Bind(&restore).Clicked() {
			actions++
		}
		Dialog(c, &outer, DialogProps{Title: "Outer", HideClose: true, InitialFocus: &first, ReturnFocus: &restore,
			Footer: func() { ui.Button(c, "Outer footer") }}, func() {
			ui.Button(c, "Outer first").Bind(&first)
			if ui.Button(c, "Open inner").Bind(&innerRestore).Clicked() {
				inner = true
			}
			Dialog(c, &inner, DialogProps{Title: "Inner", HideClose: true, InitialFocus: &innerFirst, ReturnFocus: &innerRestore}, func() {
				ui.Button(c, "Inner first").Bind(&innerFirst)
				if ui.Button(c, "Close inner").Clicked() {
					inner = false
				}
			})
		})
	}), 600, 480)
	if err := r.Click("Open outer"); err != nil {
		t.Fatal(err)
	}
	r.Advance(200 * time.Millisecond)
	if !r.Focused("Outer first") {
		t.Fatal("initial focus did not use bound control")
	}
	if err := r.Click("Open inner"); err != nil {
		t.Fatal(err)
	}
	r.Advance(200 * time.Millisecond)
	if !r.Focused("Inner first") {
		t.Fatal("nested initial focus did not use bound control")
	}
	for i := 0; i < 6; i++ {
		r.Key(0, ui.KeyTab)
		if !r.Focused("Inner first") && !r.Focused("Close inner") {
			t.Fatal("Tab escaped nested native modal")
		}
	}
	if err := r.Click("Close inner"); err != nil {
		t.Fatal(err)
	}
	// Pointer input over the covered window must remain blocked during exit.
	r.ClickAt(30, 82)
	if actions != 0 {
		t.Fatal("underlying window became interactive during modal exit")
	}
	r.Advance(200 * time.Millisecond)
	if !r.Focused("Open inner") || !outer {
		t.Fatalf("nested removal did not restore outer opener: outer=%v focus=%+v", outer, focusedModalControls(r))
	}
	r.Key(0, ui.KeyEscape)
	r.Advance(200 * time.Millisecond)
	if outer || !r.Focused("Restore here") {
		t.Fatal("explicit return focus was lost on native scope removal")
	}
}

func focusedModalControls(r *ui.Tester) []string {
	var names []string
	for _, a := range r.AccessAll() {
		if a.Focused {
			names = append(names, a.Name)
		}
	}
	return names
}

func TestDialogExitCanReverseWithoutReplacingNativeModal(t *testing.T) {
	open := true
	r := ui.NewTester(view(func(c *ui.Context) {
		Dialog(c, &open, DialogProps{Title: "Reversible", HideClose: true}, func() { ui.Button(c, "Reversible action") })
	}), 600, 480)
	r.Advance(200 * time.Millisecond)
	open = false
	r.Frame()
	r.Advance(30 * time.Millisecond)
	open = true
	r.Frame()
	r.Advance(200 * time.Millisecond)
	if _, ok := r.AccessByRole(ui.RoleDialog, "Reversible"); !ok {
		t.Fatal("reopening during exit removed the modal")
	}
	if err := r.Click("Reversible action"); err != nil {
		t.Fatal(err)
	}
	if !r.Focused("Reversible action") || r.FramePending() {
		t.Fatal("reopened content remained inert or animation never settled")
	}
}

func TestAlertDialogPendingDisablesRepeatedActionAndEscape(t *testing.T) {
	open, pending, allowCancel, accepted, canceled := true, false, false, 0, 0
	r := ui.NewTester(view(func(c *ui.Context) {
		if AlertDialog(c, &open, AlertDialogProps{Title: "Publish changes", ActionLabel: "Publish", Pending: pending,
			AllowCancelWhilePending: allowCancel, OnCancel: func() { canceled++ }}) {
			accepted++
			pending = true
			c.Invalidate()
		}
	}), 600, 480)
	r.Advance(200 * time.Millisecond)
	if err := r.Click("Publish"); err != nil {
		t.Fatal(err)
	}
	a, _ := r.AccessByRole(ui.RoleButton, "Publish")
	cancel, _ := r.AccessByRole(ui.RoleButton, "Cancel")
	d, _ := r.AccessByRole(ui.RoleAlertDialog, "Publish changes")
	if accepted != 1 || !a.Disabled || !a.Busy || !cancel.Disabled || !d.Busy || !open {
		t.Fatalf("pending state incorrect: accepted=%d action=%+v cancel=%+v dialog=%+v", accepted, a, cancel, d)
	}
	r.Key(0, ui.KeyEnter)
	r.Key(0, ui.KeyEscape)
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	if accepted != 1 || !open || canceled != 0 {
		t.Fatal("pending alert accepted another action or dismissal")
	}
	allowCancel = true
	r.Frame()
	if err := r.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if open || canceled != 1 {
		t.Fatal("explicit cancellation during pending did not work")
	}
}

func TestSheetPhysicalAndLogicalSidesRTL(t *testing.T) {
	for _, logical := range []bool{false, true} {
		for _, side := range []Placement{Top, Bottom, Left, Right} {
			t.Run(fmt.Sprintf("%d/logical=%t", side, logical), func(t *testing.T) {
				open := false
				r := ui.NewTester(view(func(c *ui.Context) {
					Sheet(c, &open, SheetProps{Title: "Edge", Side: side, Size: 200, LogicalSide: logical}, func() { ui.Button(c, "Sheet action") })
				}), 600, 500)
				r.SetDirection(ui.RightToLeft)
				open = true
				r.Frame()
				r.Advance(300 * time.Millisecond)
				d, ok := r.AccessByRole(ui.RoleDialog, "Edge")
				physical := logicalSide(side, logical, ui.RightToLeft)
				x, y, w, h := float32(0), float32(0), float32(600), float32(200)
				if physical == Bottom {
					y = 300
				} else if physical == Left || physical == Right {
					w, h = 200, 500
					if physical == Right {
						x = 400
					}
				}
				if !ok || math.Abs(float64(d.Bounds.X-x))+math.Abs(float64(d.Bounds.Y-y))+math.Abs(float64(d.Bounds.W-w))+math.Abs(float64(d.Bounds.H-h)) > 1 {
					t.Fatalf("wrong side or size: %+v", d.Bounds)
				}
			})
		}
	}
}

func TestAlertDialogResponsiveMediaSmallFooterAndPinnedActions(t *testing.T) {
	for _, variant := range []struct {
		width     int
		small     bool
		wantWidth float32
	}{{500, false, 320}, {800, false, 512}, {500, true, 320}} {
		t.Run(fmt.Sprintf("%d/small=%v", variant.width, variant.small), func(t *testing.T) {
			open := true
			r := ui.NewTester(view(func(c *ui.Context) {
				AlertDialog(c, &open, AlertDialogProps{Title: "Delete workspace?", Description: "All projects are included.", Small: variant.small,
					Media: func() { ui.Box(c).Fill().Label("Warning media").Role(ui.RoleImage) },
					Content: func() {
						for i := 0; i < 25; i++ {
							ui.Text(c, fmt.Sprintf("Affected project %d", i)).Height(24)
						}
					},
				})
			}), variant.width, 500)
			r.Advance(200 * time.Millisecond)
			d, _ := r.AccessByRole(ui.RoleAlertDialog, "Delete workspace?")
			media, _ := r.AccessByRole(ui.RoleImage, "Warning media")
			cancel, _ := r.AccessByRole(ui.RoleButton, "Cancel")
			action, _ := r.AccessByRole(ui.RoleButton, "Continue")
			if math.Abs(float64(d.Bounds.W-variant.wantWidth)) > 1 || d.Bounds.H > 468 {
				t.Fatalf("responsive alert bounds are wrong: %+v", d.Bounds)
			}
			if math.Abs(float64(media.Bounds.W-64))+math.Abs(float64(media.Bounds.H-64)) > 1 {
				t.Fatalf("media did not receive its source 64-DIP surface: %+v", media.Bounds)
			}
			if action.Bounds.Y+action.Bounds.H > d.Bounds.Y+d.Bounds.H-20 || cancel.Bounds.Y < d.Bounds.Y {
				t.Fatal("actions scrolled outside long alert")
			}
			if variant.small && (math.Abs(float64(action.Bounds.Y-cancel.Bounds.Y))+math.Abs(float64(action.Bounds.W-cancel.Bounds.W)) > 1) {
				t.Fatalf("small alert footer did not use two equal columns: cancel=%+v action=%+v", cancel.Bounds, action.Bounds)
			}
			if err := r.Click("Cancel"); err != nil {
				t.Fatal(err)
			}
			if open {
				t.Fatal("pinned cancel action was not reachable")
			}
		})
	}
}

func TestDialogLongContentCanScrollToLastActionWithinWindow(t *testing.T) {
	open, selected := true, false
	r := ui.NewTester(view(func(c *ui.Context) {
		Dialog(c, &open, DialogProps{Title: "Long settings"}, func() {
			for i := 0; i < 30; i++ {
				ui.Text(c, fmt.Sprintf("Setting row %d", i)).Height(28)
			}
			if ui.Button(c, "Last setting").Clicked() {
				selected = true
			}
		})
	}), 600, 420)
	r.Advance(300 * time.Millisecond)
	panel, ok := r.AccessByRole(ui.RoleDialog, "Long settings")
	if !ok {
		t.Fatal("missing modal")
	}
	r.Scroll(panel.Bounds.X+panel.Bounds.W/2, panel.Bounds.Y+panel.Bounds.H/2, 0, 2000)
	r.Advance(300 * time.Millisecond)
	button, ok := r.AccessByRole(ui.RoleButton, "Last setting")
	if !ok || button.Bounds.Y < panel.Bounds.Y || button.Bounds.Y+button.Bounds.H > panel.Bounds.Y+panel.Bounds.H+1 {
		t.Fatalf("last setting remains outside constrained modal: panel=%+v button=%+v", panel.Bounds, button.Bounds)
	}
	if err := r.Click("Last setting"); err != nil {
		t.Fatal(err)
	}
	if !selected {
		t.Fatal("last setting action inaccessible")
	}
}
