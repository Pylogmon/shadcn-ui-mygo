package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorDialogCompositionSearchEditAndPageReturn(t *testing.T) {
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
	click := func(role ui.Role, name string) {
		t.Helper()
		a, ok := r.AccessByRole(role, name)
		if !ok || a.Offscreen {
			t.Fatalf("missing visible actual target %q", name)
		}
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	oldName := s.Name
	navigate("Dialog", "overlays")
	open, ok := r.AccessByRole(ui.RoleButton, "Open Dialog")
	viewport, _ := r.AccessByRole(ui.RoleRegion, "Overlays")
	if !ok || open.Bounds.Y < viewport.Bounds.Y || open.Bounds.Y+open.Bounds.H > viewport.Bounds.Y+viewport.Bounds.H {
		t.Fatal("Dialog search did not reveal the actual source opener")
	}
	click(ui.RoleButton, "Open Dialog")
	if _, ok := r.AccessByRole(ui.RoleDialog, "Edit profile"); !ok {
		t.Fatal("the source opener did not create its native modal")
	}
	click(ui.RoleTextField, "Name")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Milestone profile")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	d := s.dialogCompositions.drafts["demo"]
	if d.name != "Milestone profile" || d.changes == 0 || d.saves != 0 || !d.open {
		t.Fatal("the source native profile editor lost its draft or invented an Enter submit")
	}
	click(ui.RoleButton, "Save changes")
	if d.saves != 1 || d.saved["name"] != "Milestone profile" || !d.open || s.dialogCompositions.action != "Local dialog action: Save changes" {
		t.Fatal("the source Save did not capture once and preserve the open dialog")
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if d.open || !r.Focused("Open Dialog") {
		t.Fatal("source native Escape failed to return focus to its actual opener")
	}
	navigate("Item", "content")
	navigate("Dialog", "overlays")
	click(ui.RoleButton, "Open Dialog")
	name, ok := r.AccessByRole(ui.RoleTextField, "Name")
	if !ok || name.Value != "Milestone profile" || s.dialogCompositions.recipe != "demo" || d.saves != 1 || s.Name != oldName || s.Dialog {
		t.Fatal("Dialog page return lost the source draft or changed the original Profile/Dialog model")
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
}
