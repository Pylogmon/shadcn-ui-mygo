package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/gallery"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestSnapshotIncludesCompletedActivityBarsWithoutChangingPreferences(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint("dark=", dark), func(t *testing.T) {
			state := gallery.New()
			state.Theme.Mode = theme.LightMode
			if dark {
				state.Theme.Mode = theme.DarkMode
			}
			if state.Config.ReduceMotion {
				t.Fatal("fixture should exercise normal motion")
			}
			actual := renderSnapshot(state)
			if state.Config.ReduceMotion {
				t.Fatal("snapshot changed the application's motion preference")
			}
			// Independent normal-motion reference: wait through the public 400ms
			// Bar entrance, then sample each real native mark's final bounds.
			reference := ui.NewTester(state.View, 1280, 900)
			reference.Advance(time.Second)
			for _, label := range []string{"Builds Mon: 32", "Builds Tue: 48", "Builds Wed: 36", "Builds Thu: 72", "Builds Fri: 56", "Builds Sat: 88"} {
				mark, ok := reference.AccessByRole(ui.RoleGroup, label)
				if !ok || mark.Bounds.W <= 4 || mark.Bounds.H <= 4 || mark.Offscreen {
					t.Fatalf("reference native mark is not visible: %s %+v", label, mark)
				}
				x, y := int(mark.Bounds.X+mark.Bounds.W/2), int(mark.Bounds.Y+mark.Bounds.H/2)
				want := reference.Image().RGBAAt(x, y)
				if got := actual.RGBAAt(x, y); got != want {
					t.Errorf("snapshot omitted the completed %s mark at (%d,%d): got %v want %v", label, x, y, got, want)
				}
			}
		})
	}
}
