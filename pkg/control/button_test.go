package control

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestButtonSoftColorsAndInteraction(t *testing.T) {
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 16 16" fill="currentColor"><path d="M0 0h16v16H0z"/></svg>`))
	for _, dark := range []bool{false, true} {
		tokens := theme.Preset("blue", dark)
		for _, tc := range []struct {
			name  string
			color *ui.Color
		}{
			{name: "theme"},
			{name: "green", color: new(ui.RGB(20, 160, 80))},
			{name: "destructive", color: &tokens.Destructive},
			{name: "black", color: new(ui.RGB(0, 0, 0))},
			{name: "transparent", color: new(ui.Transparent)},
			{name: "translucent", color: new(ui.RGBA(120, 40, 220, .5))},
		} {
			t.Run(fmt.Sprintf("dark=%v/%s", dark, tc.name), func(t *testing.T) {
				p := ButtonProps{Variant: Soft, Color: tc.color, Size: Icon, Icon: icon, Label: "Soft action"}
				calls := 0
				r := ui.NewTesterWith(func(c *ui.Context) {
					theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).Children(func() {
						if Button(c, "", p).Clicked() {
							calls++
						}
					})
				}, 180, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				a, ok := r.AccessByRole(ui.RoleButton, p.Label)
				if !ok {
					t.Fatal("missing native Soft button")
				}
				box := a.Bounds
				x, y := box.X+box.W/2, box.Y+box.H/2
				base := color.RGBA{R: tokens.Background.R, G: tokens.Background.G, B: tokens.Background.B, A: 255}
				// Source-over on the opaque page also checks that custom alpha is retained.
				blend := func(bg color.RGBA, fg ui.Color) color.RGBA {
					a := float32(fg.A) / 255
					channel := func(b, f uint8) uint8 { return uint8(float32(b)*(1-a) + float32(f)*a + .5) }
					return color.RGBA{R: channel(bg.R, fg.R), G: channel(bg.G, fg.G), B: channel(bg.B, fg.B), A: 255}
				}
				pixel := func(name string, x, y float32, want color.RGBA) {
					t.Helper()
					got := r.Image().RGBAAt(int(x), int(y))
					near := func(a, b uint8) bool { return int(a)-int(b) >= -1 && int(a)-int(b) <= 1 }
					if !near(got.R, want.R) || !near(got.G, want.G) || !near(got.B, want.B) || got.A != want.A {
						t.Fatalf("%s pixel=%v want %v", name, got, want)
					}
				}
				chosen := tokens.Primary
				if p.Color != nil {
					chosen = *p.Color
				}
				quietAlpha, activeAlpha, focusAlpha := float32(.1), float32(.2), float32(.2)
				if dark {
					quietAlpha, activeAlpha, focusAlpha = .2, .3, .4
				}
				quiet := blend(base, chosen.Alpha(quietAlpha))
				pixel("quiet fill", x, box.Y+4, quiet)
				pixel("inherited icon", x, y, blend(quiet, chosen))
				r.Move(x, y)
				pixel("hover fill", x, box.Y+4, blend(base, chosen.Alpha(activeAlpha)))
				r.Press(x, y)
				pixel("pressed fill", x, box.Y+4, blend(base, chosen.Alpha(activeAlpha)))
				r.Release(x, y)
				r.Move(170, 90)
				r.Key(0, ui.KeyTab)
				if !r.Focused(p.Label) {
					t.Fatal("Soft action lost keyboard focus")
				}
				pixel("custom focus ring", x, box.Y-2, blend(base, chosen.Alpha(focusAlpha)))
				r.Key(0, ui.KeyEnter)
				if calls != 2 {
					t.Fatalf("pointer and keyboard activations=%d want 2", calls)
				}
				p.Invalid = true
				r.Frame()
				invalidRing := r.Image().RGBAAt(int(x), int(box.Y-2))
				if invalidRing.R <= invalidRing.G || invalidRing.R <= invalidRing.B {
					t.Fatalf("invalid state did not retain the theme danger color: %v", invalidRing)
				}
				a, _ = r.AccessByRole(ui.RoleButton, p.Label)
				if !a.Invalid {
					t.Fatal("Soft action lost invalid semantics")
				}
				p.Invalid, p.Color = false, new(ui.RGB(40, 180, 210))
				r.Frame()
				pixel("updated fill", x, box.Y+4, blend(base, p.Color.Alpha(quietAlpha)))
				pixel("updated focus ring", x, box.Y-2, blend(base, p.Color.Alpha(focusAlpha)))
			})
		}
	}
}

func TestButtonSoftDisabledAndLoading(t *testing.T) {
	for _, loading := range []bool{false, true} {
		t.Run(fmt.Sprint("loading=", loading), func(t *testing.T) {
			calls := 0
			r := tester(func(c *ui.Context) {
				if Button(c, "Unavailable", ButtonProps{Variant: Soft, Color: new(ui.RGB(20, 160, 80)), Disabled: !loading, Loading: loading}).Clicked() {
					calls++
				}
			})
			a, _ := r.AccessByRole(ui.RoleButton, "Unavailable")
			if !a.Disabled || a.Focusable || a.Busy != loading {
				t.Fatalf("disabled/loading semantics=%+v", a)
			}
			before := r.Image().RGBAAt(int(a.Bounds.X+4), int(a.Bounds.Y+a.Bounds.H/2))
			click(t, r, "Unavailable")
			r.Key(0, ui.KeyEnter)
			after := r.Image().RGBAAt(int(a.Bounds.X+4), int(a.Bounds.Y+a.Bounds.H/2))
			if calls != 0 || before != after {
				t.Fatalf("disabled/loading action changed: calls=%d fill=%v→%v", calls, before, after)
			}
		})
	}
}
