package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func revealPreviewPart(t *testing.T, r *ui.Tester, region string, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	for range 6 {
		part, found := r.AccessByRole(role, label)
		panel, hasPanel := r.AccessByRole(ui.RoleRegion, region)
		if !found || !hasPanel {
			t.Fatalf("missing %q in %q", label, region)
		}
		b, p := part.Bounds, panel.Bounds
		if b.Y >= p.Y && b.Y+b.H <= p.Y+p.H {
			if b.W <= 0 || b.X < p.X-1 || b.X+b.W > p.X+p.W+1 {
				t.Fatalf("preview control %q exceeds its scroll region: %v in %v", label, b, p)
			}
			return part
		}
		r.Scroll(p.X+p.W-12, p.Y+p.H/2, 0, b.Y+b.H/2-p.Y-p.H/2)
	}
	t.Fatalf("preview scroll could not reveal %q", label)
	return ui.AccessInfo{}
}

func TestCreatorPaymentsSearchEditingAndPageReturn(t *testing.T) {
	s := New()
	s.Config.ReduceMotion = true
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	r.Key(ui.Cmd, ui.KeyK)
	r.Frame()
	r.Type("Payments")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.Page != "display" || s.Command || !r.HasText("Recent invoices") || !r.HasText("Table composition") {
		t.Fatal("Payments search did not reach the composed Data page with its existing examples")
	}
	a := revealPreviewPart(t, r, "Data & display", ui.RoleTextField, "Filter payment emails")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Type("abe45")
	r.Frame()
	if s.payments.table.Query != "abe45" || !r.Focused("Filter payment emails") || !r.HasText("abe45@example.com") || s.payments.popupError != "" {
		t.Fatal("the narrow Data page did not bind the real email filter to Payments")
	}
	s.Page = "navigation"
	r.Frame()
	r.Key(ui.Cmd, ui.KeyK)
	r.Frame()
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Payment")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.Page != "display" || s.Command || s.payments.table.Query != "abe45" || s.Table.Query != "" {
		t.Fatal("Payments page return lost its filter or changed the independent invoice model")
	}
	if revealPreviewPart(t, r, "Data & display", ui.RoleTextField, "Filter payment emails").Value != "abe45" {
		t.Fatal("the retained payment query was not restored into its actual editor")
	}
}
