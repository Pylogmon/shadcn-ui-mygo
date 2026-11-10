package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorItemCompositionSearchActionAndPageReturn(t *testing.T) {
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
	navigate("Item", "content")
	if !r.HasText("Project files") || !r.HasText("Empty states") {
		t.Fatal("Item compositions displaced the original Content examples")
	}
	a := revealPreviewPart(t, r, "Content & feedback", ui.RoleButton, "Run basic item action")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	if s.itemCompositions.action != "Local item action: Basic Item" {
		t.Fatal("the real source Item action did not update its independent application model")
	}
	navigate("Popover", "overlays")
	navigate("Item", "content")
	if s.itemCompositions.recipe != "demo" || s.itemCompositions.action != "Local item action: Basic Item" || s.contentPinned {
		t.Fatal("Item page return lost its configuration or changed the original file model")
	}
	if revealPreviewPart(t, r, "Content & feedback", ui.RoleButton, "Run basic item action").Bounds.W <= 0 {
		t.Fatal("the source Item action did not remain reachable after page return")
	}
}
