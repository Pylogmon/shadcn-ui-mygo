package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorDisclosureCompositionSearchEditingAndPageReturn(t *testing.T) {
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
		a := revealPreviewPart(t, r, "Data & display", role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	navigate("Collapsible", "display")
	click(ui.RoleTextField, "Radius X")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("18")
	click(ui.RoleDisclosure, "Expand radius settings")
	click(ui.RoleTextField, "Expanded Radius X")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("9")
	r.Frame()
	if s.disclosureCompositions.radius[0] != "18" || s.disclosureCompositions.radius[2] != "9" || s.disclosureCompositions.popupError != "" {
		t.Fatal("the independent trigger and actual radius editors did not update their application drafts")
	}
	navigate("Empty", "content")
	navigate("RadiusSettings", "display")
	if s.disclosureCompositions.recipe != "settings" || s.disclosureCompositions.radius[0] != "18" || s.disclosureCompositions.radius[2] != "9" {
		t.Fatal("Disclosure page return lost its independent configuration or drafts")
	}
	if revealPreviewPart(t, r, "Data & display", ui.RoleTextField, "Radius X").Value != "18" || revealPreviewPart(t, r, "Data & display", ui.RoleTextField, "Expanded Radius X").Value != "9" {
		t.Fatal("retained radius drafts did not return to their actual native editors")
	}
}
