package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorSliderCompositionSearchChangeAndPageReturn(t *testing.T) {
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
	thumb := func() ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(ui.RoleSlider, "Slider")
		v, _ := r.AccessByRole(ui.RoleRegion, "Forms & controls")
		if !ok || a.Bounds.Y < v.Bounds.Y || a.Bounds.Y+a.Bounds.H > v.Bounds.Y+v.Bounds.H {
			t.Fatal("source Slider search did not reveal its real native thumb")
		}
		return a
	}
	oldSlider := s.Slider
	navigate("Slider", "controls")
	a := thumb()
	if s.sliderCompositions.drafts["demo"].values[0] != 75 {
		t.Fatal("the source Slider default was lost")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	before, count := s.sliderCompositions.drafts["demo"].values[0], s.sliderCompositions.drafts["demo"].changes
	r.Key(0, ui.KeyRight)
	r.Frame()
	value := s.sliderCompositions.drafts["demo"].values[0]
	if value != before-1 || s.sliderCompositions.drafts["demo"].changes != count+1 {
		t.Fatal("the real source Slider failed one RTL keyboard step")
	}
	navigate("Badge", "content")
	navigate("Slider", "controls")
	thumb()
	if s.sliderCompositions.recipe != "demo" || s.sliderCompositions.drafts["demo"].values[0] != value || s.sliderCompositions.drafts["demo"].changes != count+1 || s.Slider != oldSlider {
		t.Fatal("Slider page return lost the source value or changed the original Volume")
	}
}
