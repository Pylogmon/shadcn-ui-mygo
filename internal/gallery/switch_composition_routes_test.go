package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// Source cards can contain the same labels as the original examples. Locate
// the real target inside its card, then scroll that target into the viewport.
func revealCompositionControl(t *testing.T, r *ui.Tester, region, card string, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	for range 6 {
		container, hasCard := r.AccessByRole(ui.RoleGroup, card)
		viewport, hasViewport := r.AccessByRole(ui.RoleRegion, region)
		if !hasCard || !hasViewport {
			t.Fatalf("missing source card %q or viewport %q", card, region)
		}
		for _, a := range r.AccessAll() {
			b, c, v := a.Bounds, container.Bounds, viewport.Bounds
			if a.Role != role || a.Name != name || b.Y < c.Y || b.Y+b.H > c.Y+c.H || b.X < c.X-1 || b.X+b.W > c.X+c.W+1 {
				continue
			}
			if b.Y >= v.Y && b.Y+b.H <= v.Y+v.H {
				if b.W <= 0 || b.H <= 0 || b.X < v.X-1 || b.X+b.W > v.X+v.W+1 {
					t.Fatalf("source target %q exceeds its scroll viewport: %v in %v", name, b, v)
				}
				return a
			}
			r.Scroll(v.X+v.W-12, v.Y+v.H/2, 0, b.Y+b.H/2-v.Y-v.H/2)
			break
		}
	}
	t.Fatalf("could not reveal source target %q in card %q", name, card)
	return ui.AccessInfo{}
}

func TestCreatorSwitchCompositionSearchToggleAndPageReturn(t *testing.T) {
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
	navigate("Switch", "controls")
	if !r.HasText("Workspace access") || !r.HasText("Profile") {
		t.Fatal("source Switch compositions displaced existing Forms examples")
	}
	a := revealCompositionControl(t, r, "Forms & controls", "Switch compositions", ui.RoleSwitch, "Airplane Mode")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	if !*s.switchCompositions.values["demo:airplane"] || s.switchCompositions.changes != 1 || s.switchCompositions.action != "Local switch action: Airplane Mode" {
		t.Fatal("the real source Switch failed to change its independent value exactly once")
	}
	navigate("Badge", "content")
	navigate("Switch", "controls")
	if s.switchCompositions.recipe != "demo" || !*s.switchCompositions.values["demo:airplane"] || s.switchCompositions.changes != 1 || !s.Notifications || s.quietNotifications {
		t.Fatal("Switch page return lost source state or changed the original notification model")
	}
	if !revealCompositionControl(t, r, "Forms & controls", "Switch compositions", ui.RoleSwitch, "Airplane Mode").Checked {
		t.Fatal("the retained source Switch value did not reach its actual native target")
	}
}
