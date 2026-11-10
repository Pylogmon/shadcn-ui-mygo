package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorCheckboxCompositionSearchToggleAndPageReturn(t *testing.T) {
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
			t.Fatalf("component search %q reached %q instead of %q", query, s.Page, page)
		}
	}
	navigate("Checkbox", "controls")
	a := revealCompositionControl(t, r, "Forms & controls", "Checkbox compositions", ui.RoleCheckBox, "Accept terms and conditions")
	if a.Checked || !s.checkboxCompositions.drafts["demo"].values[1] {
		t.Fatal("the two source terms checkboxes lost their distinct initial defaults")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	if !s.checkboxCompositions.drafts["demo"].values[0] || s.checkboxCompositions.changes["demo"] != 1 || s.checkboxCompositions.action != "Local checkbox change: Accept terms and conditions" {
		t.Fatal("the real source checkbox failed to update its independent draft once")
	}
	navigate("Badge", "content")
	navigate("Checkbox", "controls")
	if s.checkboxCompositions.recipe != "demo" || !s.checkboxCompositions.drafts["demo"].values[0] || s.checkboxCompositions.changes["demo"] != 1 || s.Checked || !s.readAccess || s.writeAccess {
		t.Fatal("Checkbox page return lost source state or changed the original permission/terms model")
	}
	if !revealCompositionControl(t, r, "Forms & controls", "Checkbox compositions", ui.RoleCheckBox, "Accept terms and conditions").Checked {
		t.Fatal("the retained source draft did not reach the actual native checkbox")
	}
	if a, ok := r.AccessByRole(ui.RoleCheckBox, "All permissions"); !ok || !a.Mixed {
		t.Fatal("the original permission example lost its native mixed-state semantics")
	}
}
