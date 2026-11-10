package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorCardCompositionSearchEditingAndPageReturn(t *testing.T) {
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
	click := func(role ui.Role, label string) {
		t.Helper()
		a := revealPreviewPart(t, r, "Cards", role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	navigate("Card", "overview")
	if !r.HasText("Create an account") || !r.HasText("Card compositions") {
		t.Fatal("Cards did not retain the original examples beside source compositions")
	}
	click(ui.RoleMenuButton, "Card configuration")
	if err := r.ChooseMenuItem("Login"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	click(ui.RoleTextField, "Email *")
	r.Type("review@example.com")
	r.Frame()
	if s.cardCompositions.login[0].email != "review@example.com" || s.cardCompositions.error != "" {
		t.Fatal("the actual source Card email editor did not update its independent model")
	}
	navigate("Payments", "display")
	navigate("CardSpacing", "overview")
	if s.cardCompositions.recipe != "login" || s.cardCompositions.login[0].email != "review@example.com" || s.Email != "sofia@example.com" {
		t.Fatal("Cards page return lost the source draft or changed the original account model")
	}
	if revealPreviewPart(t, r, "Cards", ui.RoleTextField, "Email *").Value != "review@example.com" {
		t.Fatal("the retained source Card draft did not return to its real editor")
	}
}
