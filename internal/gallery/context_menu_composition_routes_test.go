package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorContextMenuCompositionSearchActionAndPageReturn(t *testing.T) {
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
	navigate("ContextMenu", "navigation")
	if !r.HasText("Project workspace") || !r.HasText("Manual activation") {
		t.Fatal("source Context menus displaced the original Navigation examples")
	}
	revealPreviewPart(t, r, "Navigation", ui.RoleGroup, "Context menu target")
	if err := r.RightClick("Context menu target"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("Back"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if s.contextMenuCompositions.action != "Local context action: Back" {
		t.Fatal("the actual source system context menu did not update its independent model")
	}
	navigate("DropdownMenu", "overlays")
	navigate("ContextMenu", "navigation")
	if s.contextMenuCompositions.recipe != "basic" || s.contextMenuCompositions.action != "Local context action: Back" || s.pendingMenuAction != "" {
		t.Fatal("Context menu page return lost source state or changed the existing project menu")
	}
	revealPreviewPart(t, r, "Navigation", ui.RoleGroup, "Context menu target")
}
