package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorControlCompoundsSearchEditingAndPageReturn(t *testing.T) {
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
	click := func(region string, role ui.Role, label string) {
		t.Helper()
		a := revealPreviewPart(t, r, region, role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	navigate("ButtonGroup", "controls")
	click("Forms & controls", ui.RoleMenuButton, "Button group configuration")
	if err := r.ChooseMenuItem("Search input"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	click("Forms & controls", ui.RoleTextField, "Grouped search")
	r.Type("local group")
	click("Forms & controls", ui.RoleButton, "Search grouped text")
	if s.controlCompounds.query != "local group" || s.controlCompounds.action != "Local search: local group" || s.controlCompounds.error != "" {
		t.Fatal("actual joined input/action did not update the application model")
	}
	navigate("AvatarDropdown", "display")
	click("Data & display", ui.RoleMenuButton, "Open avatar menu")
	if err := r.ChooseMenuItem("Settings"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if s.controlCompounds.avatarAction != "Local account preview: settings" {
		t.Fatal("the actual Avatar system menu did not publish its selected item")
	}
	navigate("ButtonGroup", "controls")
	if s.controlCompounds.recipe != "input" || s.controlCompounds.query != "local group" || s.Name != "Sofia Davis" {
		t.Fatal("page navigation lost the grouped draft or changed the independent profile")
	}
	if revealPreviewPart(t, r, "Forms & controls", ui.RoleTextField, "Grouped search").Value != "local group" {
		t.Fatal("the restored grouped draft was not bound to its native editor")
	}
}
