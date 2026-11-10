package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorRadioCompositionSearchChoiceAndPageReturn(t *testing.T) {
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
	navigate("RadioGroup", "controls")
	a := revealCompositionControl(t, r, "Forms & controls", "Radio Group compositions", ui.RoleRadio, "Compact")
	if *s.radioCompositions.values["demo"] != "comfortable" || a.Checked {
		t.Fatal("source spacing radios lost their Comfortable default")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	if *s.radioCompositions.values["demo"] != "compact" || s.radioCompositions.changes["demo"] != 1 || s.radioCompositions.action != "Local radio choice: Compact" {
		t.Fatal("the actual source radio failed to update its independent value once")
	}
	r.Key(0, ui.KeyUp)
	r.Frame()
	if *s.radioCompositions.values["demo"] != "comfortable" || s.radioCompositions.changes["demo"] != 2 {
		t.Fatal("the source radio group's keyboard input did not select the preceding native radio once")
	}
	navigate("Badge", "content")
	navigate("RadioGroup", "controls")
	if s.radioCompositions.recipe != "demo" || *s.radioCompositions.values["demo"] != "comfortable" || s.radioCompositions.changes["demo"] != 2 || s.Plan != "pro" {
		t.Fatal("RadioGroup page return lost source state or changed the original Plan model")
	}
	if !revealCompositionControl(t, r, "Forms & controls", "Radio Group compositions", ui.RoleRadio, "Comfortable").Checked {
		t.Fatal("the retained selected value did not reach its actual native radio")
	}
}
