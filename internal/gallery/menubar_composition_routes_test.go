package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorMenubarCompositionSearchCommandAndPageReturn(t *testing.T) {
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
	navigate("Menubar", "navigation")
	if !r.HasText("Project workspace") || !r.HasText("Context menu compositions") {
		t.Fatal("source Menubar displaced the existing Navigation examples")
	}
	card := revealPreviewPart(t, r, "Navigation", ui.RoleGroup, "Menubar compositions").Bounds
	found := false
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleMenuButton && a.Name == "File" && a.Bounds.Y >= card.Y && a.Bounds.Y+a.Bounds.H <= card.Y+card.H {
			r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
			found = true
			break
		}
	}
	if !found {
		t.Fatal("the source Menubar's File trigger was not reachable in its card")
	}
	if err := r.ChooseMenuItem("New Tab"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if s.menubarCompositions.lastAction != "New Tab" || s.menubarCompositions.commandCount != 1 || !r.HasText("Last command: New Tab") {
		t.Fatal("the real source system-menu command did not update its independent model exactly once")
	}
	navigate("DropdownMenu", "overlays")
	navigate("Menubar", "navigation")
	if s.menubarCompositions.recipe != "demo" || s.menubarCompositions.commandCount != 1 || s.menubarCompositions.lastAction != "New Tab" || s.pendingMenuAction != "" {
		t.Fatal("Menubar page return lost source state or changed the existing project menu")
	}
	revealPreviewPart(t, r, "Navigation", ui.RoleGroup, "Menubar compositions")
}
