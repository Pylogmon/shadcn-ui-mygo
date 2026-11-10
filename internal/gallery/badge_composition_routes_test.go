package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorBadgeCompositionSearchLinkAndPageReturn(t *testing.T) {
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
	navigate("Badge", "content")
	if !r.HasText("Project files") || !r.HasText("Empty states") {
		t.Fatal("source Badge compositions displaced existing Content examples")
	}
	a := revealPreviewPart(t, r, "Content & feedback", ui.RoleMenuButton, "Badge configuration")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	if err := r.ChooseMenuItem("Open Link"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	a = revealPreviewPart(t, r, "Content & feedback", ui.RoleLink, "Open Link")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	if s.badgeCompositions.action != "Local badge action: Open Link" || s.badgeCompositions.activations != 1 || len(r.OpenedURLs()) != 0 || s.badgeCompositions.router.Path() != "/badge-examples" {
		t.Fatal("the real source Badge link failed to activate once through its local Router")
	}
	navigate("Button", "controls")
	navigate("Badge", "content")
	if s.badgeCompositions.recipe != "link" || s.badgeCompositions.action != "Local badge action: Open Link" || s.badgeCompositions.activations != 1 || s.contentAction != "" {
		t.Fatal("Badge page return lost local source state or changed the existing Content model")
	}
	revealPreviewPart(t, r, "Content & feedback", ui.RoleLink, "Open Link")
}
