package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorStatusCompositionSearchEditingAndPageReturn(t *testing.T) {
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
		a := revealPreviewPart(t, r, "Content & feedback", role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	navigate("Empty", "content")
	if !r.HasText("Project files") || !r.HasText("Empty states") {
		t.Fatal("source compositions displaced the original Content examples")
	}
	click(ui.RoleMenuButton, "Empty configuration")
	if err := r.ChooseMenuItem("Search input group"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	click(ui.RoleTextField, "Search missing pages")
	r.Type("local page")
	r.Frame()
	if s.statusCompositions.emptyRecipe != "search" || s.statusCompositions.queries["search"] == nil || *s.statusCompositions.queries["search"] != "local page" {
		t.Fatal("the actual Empty search editor did not bind its application draft")
	}
	navigate("AlertDialog", "overlays")
	navigate("EmptySearch", "content")
	if s.statusCompositions.emptyRecipe != "search" || *s.statusCompositions.queries["search"] != "local page" || s.contentProject != "" {
		t.Fatal("page return lost the source search or changed the original Empty form")
	}
	if revealPreviewPart(t, r, "Content & feedback", ui.RoleTextField, "Search missing pages").Value != "local page" {
		t.Fatal("the retained Empty search did not return to its native editor")
	}
}
