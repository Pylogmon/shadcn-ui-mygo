package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCreatorOTPCompositionSearchEditAndPageReturn(t *testing.T) {
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
	code := func() ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(ui.RoleTextField, "One-time code")
		v, _ := r.AccessByRole(ui.RoleRegion, "Forms & controls")
		if !ok || a.Bounds.Y < v.Bounds.Y || a.Bounds.Y+a.Bounds.H > v.Bounds.Y+v.Bounds.H {
			t.Fatal("source OTP search did not reveal its real native editor")
		}
		return a
	}
	oldOTP := s.OTP
	navigate("InputOTP", "controls")
	a := code()
	if a.Value != "123456" {
		t.Fatal("the source default populated OTP was lost")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("987654")
	r.Frame()
	if *s.otpCompositions.drafts["demo"] != "987654" || s.otpCompositions.changes["demo"] == 0 {
		t.Fatal("the source native OTP did not update its independent draft")
	}
	navigate("Badge", "content")
	navigate("InputOTP", "controls")
	if code().Value != "987654" || *s.otpCompositions.drafts["demo"] != "987654" || s.otpCompositions.recipe != "demo" || s.OTP != oldOTP {
		t.Fatal("OTP page return lost its source draft or changed the original OTP model")
	}
}
