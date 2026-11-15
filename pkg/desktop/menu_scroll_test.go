package desktop

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestMenuContentPointerHighlightPreservesUserScroll(t *testing.T) {
	for _, selectList := range []bool{false, true} {
		for _, reduced := range []bool{false, true} {
			t.Run(fmt.Sprintf("select=%v/reduced=%v", selectList, reduced), func(t *testing.T) {
				state := MenuState{}
				items := make([]overlay.MenuItem, 100)
				for i := range items {
					items[i] = overlay.MenuItem{Key: fmt.Sprint(i), Label: fmt.Sprintf("Font %02d", i)}
				}
				tt := ui.NewTesterWith(func(c *ui.Context) {
					theme.Apply(c)
					MenuContent(c, &state, items, MenuContentProps{selectList: selectList, Label: "Fonts", MaxHeight: 170})
				}, 260, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
				role, panelRole := ui.RoleMenuItem, ui.RoleMenu
				if selectList {
					role, panelRole = ui.RoleOption, ui.RoleListBox
				}
				access := func(role ui.Role, label string) ui.AccessInfo {
					t.Helper()
					a, ok := tt.AccessByRole(role, label)
					if !ok {
						t.Fatalf("missing %q with role %v", label, role)
					}
					return a
				}
				tt.Key(0, ui.KeyEnd)
				last := access(role, "Font 99")
				if last.Offscreen {
					t.Fatal("keyboard End did not reveal the last item")
				}
				panel := access(panelRole, "Fonts").Bounds
				x, y := panel.X+panel.W/2, panel.Y+panel.H/2
				tt.Move(x, y)
				previous := state.Highlighted[0]
				tt.Scroll(x, y, 0, -384)
				before := access(role, "Font 00").Bounds.Y
				if old := access(role, items[menuItemIndex(items, previous)].Label); !old.Offscreen {
					t.Fatal("manual scroll did not move the previous highlight offscreen")
				}
				// Hovering an earlier row must not reveal the previous, later
				// highlight while that stale row is still being built this frame.
				tt.Move(x+1, y)
				tt.Advance(200 * time.Millisecond)
				if after := access(role, "Font 00").Bounds.Y; math.Abs(float64(after-before)) > .1 {
					t.Fatalf("pointer highlight pulled the list back: first row y=%v, want %v", after, before)
				}
				if state.Highlighted[0] == previous {
					t.Fatal("pointer did not highlight the row under its new position")
				}
				tt.Key(0, ui.KeyHome)
				if first := access(role, "Font 00"); first.Offscreen || state.Highlighted[0] != "0" {
					t.Fatal("keyboard Home lost its reveal after pointer scrolling")
				}
			})
		}
	}
}
