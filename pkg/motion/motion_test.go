package motion

import (
	"github.com/egoist/mygo/ui"
	"testing"
	"time"
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
