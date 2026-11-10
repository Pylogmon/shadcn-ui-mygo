package gallery

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
)

func TestColorPickerSearchNarrowRTLAndPageReturn(t *testing.T) {
	s := New()
	s.Config.NativeMenus, s.Config.ReduceMotion = true, true
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	navigate := func(query string) {
		t.Helper()
		r.Key(ui.Cmd, ui.KeyK)
		r.Frame()
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(query)
		r.Key(0, ui.KeyEnter)
		r.Frame()
	}
	navigate("ColorPicker")
	if s.Page != "controls" || s.pendingPreview != "" {
		t.Fatal("ColorPicker search did not reveal its preview")
	}
	viewport := bookingAccess(t, r, ui.RoleRegion, "Forms & controls")
	well := bookingAccess(t, r, ui.RoleColorWell, "Custom color")
	if well.Bounds.Y < viewport.Bounds.Y || well.Bounds.Y+well.Bounds.H > viewport.Bounds.Y+viewport.Bounds.H {
		t.Fatal("search did not bring the real color well into view")
	}
	var first, last ui.AccessInfo
	for _, option := range control.DefaultColorOptions() {
		swatch := bookingAccess(t, r, ui.RoleRadio, option.Label)
		if swatch.Bounds.X < viewport.Bounds.X || swatch.Bounds.X+swatch.Bounds.W > viewport.Bounds.X+viewport.Bounds.W {
			t.Fatalf("narrow RTL palette overflowed its viewport: %+v in %+v", swatch.Bounds, viewport.Bounds)
		}
		if option.Label == "Blue" {
			first = swatch
		} else if option.Label == "White" {
			last = swatch
		}
	}
	if last.Bounds.Y <= first.Bounds.Y {
		t.Fatal("narrow palette did not wrap to another line")
	}
	bookingClick(t, r, ui.RoleRadio, "Circle")
	bookingClick(t, r, ui.RoleRadio, "Green")
	if s.colorPicker.shape != "circle" || s.colorPicker.value != control.DefaultColorOptions()[5].Color {
		t.Fatal("Gallery shape and color controls did not edit the preview")
	}
	navigate("Badge")
	navigate("颜色选择")
	if s.Page != "controls" || s.colorPicker.shape != "circle" || s.colorPicker.value != control.DefaultColorOptions()[5].Color {
		t.Fatal("page return lost the color picker configuration or value")
	}
	bookingClick(t, r, ui.RoleSwitch, "Custom color well")
	if _, ok := r.AccessByRole(ui.RoleColorWell, "Custom color"); ok {
		t.Fatal("Gallery custom-well switch did not remove the well")
	}
}
