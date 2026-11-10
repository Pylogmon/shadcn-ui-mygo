package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorSelectCompositionSearchChoiceAndPageReturn(t *testing.T) {
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
	trigger := func() ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(ui.RoleComboBox, "Source fruit")
		v, _ := r.AccessByRole(ui.RoleRegion, "Forms & controls")
		if !ok || a.Bounds.Y < v.Bounds.Y || a.Bounds.Y+a.Bounds.H > v.Bounds.Y+v.Bounds.H {
			t.Fatal("source Select search did not reveal its real native trigger")
		}
		return a
	}
	oldFramework := s.Framework
	navigate("Select", "controls")
	a := trigger()
	if s.selectCompositions.drafts["demo"].choices["main"].Value.Valid {
		t.Fatal("the source Select lost its unselected default")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	apple, ok := r.AccessByRole(ui.RoleOption, "Apple")
	if !ok || apple.Offscreen {
		t.Fatal("the source Select did not expose its actual Apple option")
	}
	r.ClickAt(apple.Bounds.X+apple.Bounds.W/2, apple.Bounds.Y+apple.Bounds.H/2)
	r.Frame()
	state := s.selectCompositions.drafts["demo"].choices["main"]
	if !state.Value.Valid || state.Value.Value != "apple" || state.Open || s.selectCompositions.drafts["demo"].changes != 1 || s.selectCompositions.action != "Local select choice: apple" {
		t.Fatal("the actual source option failed one selection and close")
	}
	navigate("Badge", "content")
	navigate("Select", "controls")
	if trigger().Value != "Apple" || s.selectCompositions.recipe != "demo" || state.Value.Value != "apple" || s.selectCompositions.drafts["demo"].changes != 1 || s.Framework != oldFramework {
		t.Fatal("Select page return lost the independent source value or changed the original Framework")
	}
}
