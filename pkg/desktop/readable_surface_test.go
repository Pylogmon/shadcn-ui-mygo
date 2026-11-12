package desktop

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestHostedPopoverDeclaresItsActualTextSurface(t *testing.T) {
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 12 12" fill="currentColor"><path d="M0 0h12v12H0z"/></svg>`))
	for _, dark := range []bool{false, true} {
		tokens := theme.Preset("amber", dark)
		tokens.Popover = ui.RGB(245, 230, 200)
		if !dark {
			tokens.Popover = ui.RGB(40, 50, 65)
		}
		backend := &fakeBackend{}
		host := NewHostWithBackend(backend)
		open := true
		ui.NewTesterWith(func(c *ui.Context) {
			c.SetTheme(tokens.Theme())
			anchor := control.Button(c, "Open", control.ButtonProps{})
			if err := Popover(c, host, "readable", anchor, &open, PopoverProps{}, func(child *ui.Context) {
				if theme.SurfaceOf(child) != tokens.Popover {
					t.Errorf("hosted text surface=%v want %v", theme.SurfaceOf(child), tokens.Popover)
				}
				control.Button(child, "", control.ButtonProps{Variant: control.Soft, Size: control.Icon, Icon: icon, Label: "Colored"})
			}); err != nil {
				t.Fatal(err)
			}
		}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		if len(backend.surfaces) != 1 {
			t.Fatal("hosted popover did not open")
		}
		r := ui.NewTesterWith(backend.surfaces[0].view, 288, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		b, ok := r.Find("Colored")
		if !ok {
			t.Fatal("missing hosted colored action")
		}
		for _, hover := range []bool{false, true} {
			if hover {
				r.Move(b.X+b.W/2, b.Y+b.H/2)
			}
			img := r.Image()
			fg, bg := img.RGBAAt(int(b.X+b.W/2), int(b.Y+b.H/2)), img.RGBAAt(int(b.X+b.W/2), int(b.Y+3))
			if theme.Contrast(ui.RGB(fg.R, fg.G, fg.B), ui.RGB(bg.R, bg.G, bg.B)) < 4.5 {
				t.Fatalf("hosted foreground=%v background=%v", fg, bg)
			}
		}
		host.CloseAll()
	}
}
