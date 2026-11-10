package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorDropdownCompositionSearchActionAndPageReturn(t *testing.T) {
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
	navigate("DropdownMenu", "overlays")
	if !r.HasText("Command palette") || !r.HasText("Popover compositions") {
		t.Fatal("source Dropdown menus displaced the existing Overlays examples")
	}
	a := revealPreviewPart(t, r, "Overlays", ui.RoleMenuButton, "Open")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	if err := r.ChooseMenuItem("Profile"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if s.dropdownCompositions.action != "Local dropdown action: Profile" {
		t.Fatal("the source system-menu action did not reach its independent model")
	}
	navigate("Item", "content")
	navigate("DropdownMenu", "overlays")
	if s.dropdownCompositions.recipe != "demo" || s.dropdownCompositions.action != "Local dropdown action: Profile" || s.contentPinned {
		t.Fatal("Dropdown page return lost its configuration or changed the original file model")
	}
	revealPreviewPart(t, r, "Overlays", ui.RoleMenuButton, "Open")
}
