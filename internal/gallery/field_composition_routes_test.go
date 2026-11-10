package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorFieldCompositionSearchEditingAndPageReturn(t *testing.T) {
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
	click := func(role ui.Role, label string) {
		t.Helper()
		a := revealPreviewPart(t, r, "Forms & controls", role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	navigate("FieldComposition", "controls")
	click(ui.RoleTextField, "Name *")
	r.Type("Native field draft")
	click(ui.RoleButton, "Submit profile")
	if s.fieldCompositions.profileSaved != "Native field draft" || s.fieldCompositions.popupError != "" || !r.HasText("Saved local profile: Native field draft") {
		t.Fatal("the actual composed Field editor/action did not save its application draft")
	}
	navigate("CardSpacing", "overview")
	navigate("FieldSet", "controls")
	if s.fieldCompositions.recipe != "responsive" || s.fieldCompositions.profileSaved != "Native field draft" || s.Name != "Sofia Davis" {
		t.Fatal("Field page return lost its source model or changed the original profile")
	}
	if revealPreviewPart(t, r, "Forms & controls", ui.RoleTextField, "Name *").Value != "Native field draft" {
		t.Fatal("the retained Field draft did not return to its native editor")
	}
}
