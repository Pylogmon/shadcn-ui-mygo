package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorLabelCompositionSearchToggleAndPageReturn(t *testing.T) {
	s := New()
	s.Config.NativeMenus, s.Config.ReduceMotion = true, true
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	navigate := func(query, page string) {
		t.Helper()
		r.Key(ui.Cmd, ui.KeyK)
		r.Frame()
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(query)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if s.Page != page || s.Command {
			t.Fatalf("search %q reached %q", query, s.Page)
		}
	}
	caption := func() ui.AccessInfo {
		t.Helper()
		card, hasCard := r.AccessByRole(ui.RoleGroup, "Label compositions")
		viewport, _ := r.AccessByRole(ui.RoleRegion, "Forms & controls")
		for _, a := range r.AccessAll() {
			if a.Role != ui.RoleText || a.Name != "Accept terms and conditions" || !hasCard || a.Bounds.Y < card.Bounds.Y || a.Bounds.Y+a.Bounds.H > card.Bounds.Y+card.Bounds.H {
				continue
			}
			if a.Bounds.Y < viewport.Bounds.Y || a.Bounds.Y+a.Bounds.H > viewport.Bounds.Y+viewport.Bounds.H {
				t.Fatal("Label search did not reveal its actual source caption")
			}
			return a
		}
		t.Fatal("missing caption inside the actual source Label card")
		return ui.AccessInfo{}
	}
	oldChecked := s.Checked
	navigate("Label", "controls")
	a := caption()
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	checked := s.labelCompositions.checks["demo:terms"]
	if checked == nil || !*checked || s.labelCompositions.changes["demo"] != 1 {
		t.Fatal("clicking the standalone source label failed to toggle its real checkbox once")
	}
	navigate("Item", "content")
	navigate("Label", "controls")
	caption()
	if s.labelCompositions.recipe != "demo" || s.labelCompositions.checks["demo:terms"] != checked || !*checked || s.labelCompositions.changes["demo"] != 1 || s.Checked != oldChecked {
		t.Fatal("Label page return lost independent source state or changed the original Terms control")
	}
}
