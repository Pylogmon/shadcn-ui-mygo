package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorButtonCompositionSearchActionAndPageReturn(t *testing.T) {
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
	navigate("Button", "controls")
	if !r.HasText("Profile") || !r.HasText("Make a choice") {
		t.Fatal("source Button compositions displaced existing Forms examples")
	}
	card := revealPreviewPart(t, r, "Forms & controls", ui.RoleGroup, "Button compositions").Bounds
	found := false
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleButton && a.Name == "Button" && a.Bounds.Y >= card.Y && a.Bounds.Y+a.Bounds.H <= card.Y+card.H {
			r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
			found = true
			break
		}
	}
	if !found {
		t.Fatal("the source Button was not reachable inside its own card")
	}
	r.Frame()
	if s.buttonCompositions.action != "Local button action: Button" || s.buttonCompositions.counts["demo"] != 1 {
		t.Fatal("the real source button failed to update its independent model exactly once")
	}
	navigate("Badge", "content")
	navigate("Button", "controls")
	if s.buttonCompositions.recipe != "demo" || s.buttonCompositions.action != "Local button action: Button" || s.buttonCompositions.counts["demo"] != 1 || s.Name != "Sofia Davis" {
		t.Fatal("Button page return lost source state or changed the existing Profile draft")
	}
	revealPreviewPart(t, r, "Forms & controls", ui.RoleGroup, "Button compositions")
}
