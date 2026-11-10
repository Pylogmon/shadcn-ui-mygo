package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorTextareaCompositionSearchEditAndPageReturn(t *testing.T) {
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
	oldBio := s.Bio
	navigate("Textarea", "controls")
	a := revealCompositionControl(t, r, "Forms & controls", "Textarea compositions", ui.RoleTextField, "Type your message here.")
	if !a.Multiline || a.Value != "" {
		t.Fatal("the source textarea lost its empty native multiline editor")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Type("First source line")
	r.Key(0, ui.KeyEnter)
	r.Type("Second source line")
	r.Frame()
	const draft = "First source line\nSecond source line"
	if *s.textareaCompositions.drafts["demo"] != draft || s.textareaCompositions.changes["demo"] == 0 || s.textareaCompositions.sends != 0 {
		t.Fatal("the source textarea failed native multiline editing or invented a submit")
	}
	navigate("Badge", "content")
	navigate("Textarea", "controls")
	a = revealCompositionControl(t, r, "Forms & controls", "Textarea compositions", ui.RoleTextField, "Type your message here.")
	if s.textareaCompositions.recipe != "demo" || *s.textareaCompositions.drafts["demo"] != draft || a.Value != draft || !a.Multiline || s.Bio != oldBio {
		t.Fatal("Textarea page return lost the independent native draft or changed the original Bio")
	}
}
