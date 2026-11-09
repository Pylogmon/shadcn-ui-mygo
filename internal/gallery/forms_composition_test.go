package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func revealFormPart(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	for range 6 {
		part, ok := r.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing Forms control %q", label)
		}
		panel, ok := r.AccessByRole(ui.RoleRegion, "Forms & controls")
		if !ok {
			t.Fatal("missing Forms scroll region")
		}
		b, p := part.Bounds, panel.Bounds
		if b.Y >= p.Y && b.Y+b.H <= p.Y+p.H {
			if b.W <= 0 || b.X < p.X-1 || b.X+b.W > p.X+p.W+1 {
				t.Fatalf("narrow Forms control %q exceeds its region: %v in %v", label, b, p)
			}
			return part
		}
		r.Scroll(p.X+p.W-12, p.Y+p.H/2, 0, b.Y+b.H/2-p.Y-p.H/2)
	}
	t.Fatalf("Forms scroll could not reveal %q", label)
	return ui.AccessInfo{}
}

func clickFormPart(t *testing.T, r *ui.Tester, role ui.Role, label string) {
	t.Helper()
	part := revealFormPart(t, r, role, label)
	r.ClickAt(part.Bounds.X+part.Bounds.W/2, part.Bounds.Y+part.Bounds.H/2)
	r.Frame()
}

func TestCreatorComposedFormsRemainReachableAndKeepDrafts(t *testing.T) {
	s := New()
	s.Page, s.Config.ReduceMotion = "controls", true
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	clickFormPart(t, r, ui.RoleTextField, "Booking amount")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("211.5")
	r.Key(0, ui.KeyEnter)
	if s.typedInputs.amount != 211.5 || !r.Focused("Booking amount") {
		t.Fatal("the actual narrow Forms amount editor did not publish its value")
	}
	clickFormPart(t, r, ui.RoleTextField, "Favorite site")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("native.example")
	clickFormPart(t, r, ui.RoleButton, "Favorite site toggle")
	if s.inputGroups.site != "native.example" || !s.inputGroups.favorite {
		t.Fatal("the actual Forms InputGroup editor and action did not share their model")
	}
	s.Page = "navigation"
	r.Frame()
	// Use the real global component search to return to the new Forms examples.
	r.Key(ui.Cmd, ui.KeyK)
	r.Frame()
	r.Type("DatePickerInput")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.Page != "controls" || s.Command || !r.HasText("Plan a booking") {
		t.Fatal("component search did not return to the actual date-input workflow")
	}
	if s.typedInputs.amount != 211.5 || s.inputGroups.site != "native.example" || !s.inputGroups.favorite || s.Name != "Sofia Davis" {
		t.Fatal("page changes lost composed form drafts or changed the existing profile")
	}
	if revealFormPart(t, r, ui.RoleTextField, "Booking amount").Value != "211.5" {
		t.Fatal("the restored booking model is not bound to the real editor")
	}
}
