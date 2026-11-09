package display

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestScrollBarReportsRealManualIntentBeforeOffsetAndUsesLatestCallback(t *testing.T) {
	var state ui.ScrollState
	version := 1
	var calls []int
	var before []float32
	disabled := false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Box(c).Size(220, 160).Children(func() {
			ui.Scroll(c).Fill().TrackScroll(&state).Scrollbars(false).Children(func() { ui.Box(c).Size(220, 600).Shrink(0) })
			current := version
			ScrollBar(c, &state, ScrollBarProps{Focusable: true, Label: "Intent bar", Disabled: disabled, OnScrollIntent: func() {
				calls = append(calls, current)
				before = append(before, state.Y)
			}})
		})
	}, 240, 180)
	r.Frame() // The standalone bar reads the viewport's committed native range.
	bar := scrollAccess(t, r, ui.RoleSlider, "Intent bar").Bounds
	r.Wheel(bar.X+bar.W/2, bar.Y+bar.H/2, 0, 80)
	if !slices.Equal(calls, []int{1}) || before[0] != 0 || state.Y != 80 {
		t.Fatalf("wheel intent did not precede the real native offset: calls=%v before=%v state=%g", calls, before, state.Y)
	}
	version = 7
	r.Frame()
	state.ScrollTo(0, 120, ui.Animation{})
	r.Frame()
	if len(calls) != 1 {
		t.Fatal("programmatic ScrollTo reported manual intent")
	}
	r.Key(0, ui.KeyTab)
	beforeKey := state.Y
	r.Key(0, ui.KeyDown)
	if !slices.Equal(calls, []int{1, 7}) || before[1] != beforeKey || state.Y <= beforeKey {
		t.Fatalf("key did not use current callback before scrolling: calls=%v before=%v y=%g", calls, before, state.Y)
	}
	r.Key(0, ui.KeyA)
	r.Wheel(bar.X+bar.W/2, bar.Y+bar.H/2, 0, 0)
	if len(calls) != 2 {
		t.Fatal("unhandled key or zero wheel reported scrolling intent")
	}
	disabled = true
	r.Frame()
	r.Wheel(bar.X+bar.W/2, bar.Y+bar.H/2, 0, 20)
	r.Press(bar.X+bar.W/2, bar.Y+bar.H/2)
	r.Release(bar.X+bar.W/2, bar.Y+bar.H/2)
	if len(calls) != 2 {
		t.Fatal("disabled bar reported manual intent")
	}
}

func TestScrollBarIntentCallbackChangesDoNotCancelNativeCapture(t *testing.T) {
	var state ui.ScrollState
	version, calls := 1, 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Box(c).Size(220, 160).Children(func() {
			ui.Scroll(c).Fill().TrackScroll(&state).Scrollbars(false).Children(func() { ui.Box(c).Size(220, 600).Shrink(0) })
			current := version
			ScrollBar(c, &state, ScrollBarProps{Focusable: true, Label: "Captured bar", OnScrollIntent: func() { calls += current }})
		})
	}, 240, 180)
	r.Frame()
	bar := scrollAccess(t, r, ui.RoleSlider, "Captured bar").Bounds
	x, y := bar.X+bar.W/2, bar.Y+10
	r.Press(x, y)
	version = 7
	r.Frame()
	r.Move(x, y+45)
	r.Release(x, y+45)
	if calls != 1 || state.Y <= 0 {
		t.Fatalf("new callback replayed intent or canceled native drag: calls=%d y=%g", calls, state.Y)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("idle intent callback left pending animation")
	}
}
