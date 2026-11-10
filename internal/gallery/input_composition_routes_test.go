package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorInputCompositionSearchEditAndPageReturn(t *testing.T) {
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
	oldName, oldBio := s.Name, s.Bio
	navigate("Input", "controls")
	a := revealCompositionControl(t, r, "Forms & controls", "Input compositions", ui.RoleTextField, "API Key")
	if !a.Password || *s.inputCompositions.drafts["demo"].texts["api"] != "" {
		t.Fatal("the source API Key lost its protected empty native editor")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	const draft = "sk-local-gallery-draft"
	r.Type(draft)
	r.Frame()
	if *s.inputCompositions.drafts["demo"].texts["api"] != draft || s.inputCompositions.drafts["demo"].changes == 0 || s.inputCompositions.action != "Local input edit: API Key" {
		t.Fatal("the actual source API Key failed to update its independent draft")
	}
	navigate("Badge", "content")
	navigate("Input", "controls")
	a = revealCompositionControl(t, r, "Forms & controls", "Input compositions", ui.RoleTextField, "API Key")
	if s.inputCompositions.recipe != "demo" || *s.inputCompositions.drafts["demo"].texts["api"] != draft || !a.Password || s.Name != oldName || s.Bio != oldBio {
		t.Fatal("Input page return lost the protected source draft or changed the original Profile")
	}
}
