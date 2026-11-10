package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorKbdCompositionSearchActionAndPageReturn(t *testing.T) {
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
			t.Fatalf("search %q reached %q", query, s.Page)
		}
	}
	navigate("Kbd", "text-layout")
	oldCode := s.codeLanguage
	selector, ok := r.AccessByRole(ui.RoleMenuButton, "Kbd configuration")
	viewport, _ := r.AccessByRole(ui.RoleRegion, "Text & layout")
	if !ok || selector.Bounds.Y < viewport.Bounds.Y || selector.Bounds.Y+selector.Bounds.H > viewport.Bounds.Y+viewport.Bounds.H {
		t.Fatal("Kbd search did not reveal its actual source configuration")
	}
	r.ClickAt(selector.Bounds.X+selector.Bounds.W/2, selector.Bounds.Y+selector.Bounds.H/2)
	if err := r.ChooseMenuItem("Accept button"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	cap, ok := r.Find("⏎")
	button, hasButton := r.AccessByRole(ui.RoleButton, "Accept")
	if !ok || !hasButton || cap.Y < button.Bounds.Y || cap.Y+cap.H > button.Bounds.Y+button.Bounds.H || cap.Y < viewport.Bounds.Y || cap.Y+cap.H > viewport.Bounds.Y+viewport.Bounds.H {
		t.Fatal("the source Accept button did not contain its visible inert keycap")
	}
	r.ClickAt(cap.X+cap.W/2, cap.Y+cap.H/2)
	r.Frame()
	if s.kbdCompositions.accepted != 1 || s.kbdCompositions.action != "Local Kbd action: Accept" {
		t.Fatal("clicking the inert keycap failed to activate its actual native parent once")
	}
	navigate("Item", "content")
	navigate("Keyboard", "text-layout")
	if s.kbdCompositions.recipe != "button" || s.kbdCompositions.accepted != 1 || s.codeLanguage != oldCode || s.Command {
		t.Fatal("Kbd page return lost independent source state or changed the original code example")
	}
}
