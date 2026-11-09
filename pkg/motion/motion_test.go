package motion

import (
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestReducedMotionPresets(t *testing.T) {
	var presets []ui.ElementTransition
	var thumb ui.Animation
	tt := ui.NewTesterWith(func(c *ui.Context) {
		presets = []ui.ElementTransition{Transition(c), Popup(c), Modal(c), Backdrop(c), Collapse(c), Drawer(c, Right), Toast(c)}
		thumb = Thumb(c)
	}, 100, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	for i, tr := range presets {
		if tr.Duration >= 0 || tr.Enter != nil || tr.Exit != nil {
			t.Errorf("preset %d animates under reduce motion: %+v", i, tr)
		}
	}
	if thumb.Duration >= 0 {
		t.Fatal("thumb animates under reduce motion")
	}
	tt.SetPreferences(ui.Preferences{})
	for i, tr := range presets {
		if tr.Duration <= 0 {
			t.Errorf("preset %d didn't restore normal animation", i)
		}
	}
}

func TestApplyTransitionCancelsRunningColorsWhenReducedMotionChanges(t *testing.T) {
	background := ui.RGB(220, 20, 40)
	border := ui.RGB(40, 200, 40)
	r := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() {
			ApplyTransition(c, ui.Box(c).Size(60, 60).Border(4, border).Background(background))
		})
	}, 80, 80)
	background, border = ui.RGB(20, 40, 220), ui.RGB(230, 180, 10)
	r.Frame()
	r.Advance(30 * time.Millisecond)
	want := color.RGBA{R: background.R, G: background.G, B: background.B, A: 255}
	if pixel := r.Image().RGBAAt(30, 30); pixel == want || !r.FramePending() {
		t.Fatalf("normal color transition did not animate: pixel=%v target=%v pending=%v", pixel, want, r.FramePending())
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	if pixel := r.Image().RGBAAt(30, 30); pixel != want {
		t.Fatalf("switching reduced motion retained an old background: pixel=%v target=%v", pixel, want)
	}
	if pixel := r.Image().RGBAAt(1, 30); pixel != (color.RGBA{R: border.R, G: border.G, B: border.B, A: 255}) {
		t.Fatalf("switching reduced motion retained an old border: %v", pixel)
	}
	if r.FramePending() {
		t.Fatal("reduced motion retained an animation frame")
	}
	background = ui.RGB(10, 120, 50)
	r.Frame()
	if pixel := r.Image().RGBAAt(30, 30); pixel != (color.RGBA{R: background.R, G: background.G, B: background.B, A: 255}) || r.FramePending() {
		t.Fatalf("reduced motion started another transition: pixel=%v pending=%v", pixel, r.FramePending())
	}
	r.SetPreferences(ui.Preferences{})
	background = ui.RGB(180, 10, 190)
	r.Frame()
	if pixel := r.Image().RGBAAt(30, 30); pixel == (color.RGBA{R: background.R, G: background.G, B: background.B, A: 255}) || !r.FramePending() {
		t.Fatalf("normal animation did not resume: pixel=%v pending=%v", pixel, r.FramePending())
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("normal transition did not settle")
	}
}

func TestPresenceActuallyEntersAndExits(t *testing.T) {
	shown := false
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() {
			if shown {
				ui.Box(c).Key("panel").Size(80, 50).Background(ui.RGB(10, 10, 10)).Transition(Popup(c))
			}
		})
	}, 100, 100)
	shown = true
	tt.Frame()
	if !tt.FramePending() {
		t.Fatal("enter did not schedule animation")
	}
	tt.Advance(200 * time.Millisecond)
	shown = false
	tt.Frame()
	if !tt.FramePending() {
		t.Fatal("exit did not schedule animation")
	}
	tt.Advance(200 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("exit never settled")
	}
}
