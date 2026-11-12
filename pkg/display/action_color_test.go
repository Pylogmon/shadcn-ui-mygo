package display_test

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Exercise actual native painting and input for the shared Color contract,
// including the separate focus target used by compound URL badges.
func TestActionVariantCustomColors(t *testing.T) {
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 12 12" fill="currentColor"><path d="M0 0h12v12H0z"/></svg>`))
	for _, dark := range []bool{false, true} {
		tokens := theme.Preset("blue", dark)
		quietTint, activeTint, focusTint := float32(.1), float32(.2), float32(.2)
		if dark {
			quietTint, activeTint, focusTint = .2, .3, .4
		}
		for _, chosen := range []ui.Color{
			ui.RGB(22, 163, 74), ui.RGB(0, 0, 0), ui.RGB(255, 255, 255),
			ui.Transparent, ui.RGBA(120, 40, 220, .5),
		} {
			for _, variant := range []struct {
				name   string
				button control.Variant
				badge  display.BadgeVariant
				quiet  ui.Color
				active ui.Color
			}{
				{"primary", control.Default, display.BadgeDefault, chosen, chosen.Alpha(.8)},
				{"outline", control.Outline, display.BadgeOutline, ui.Transparent, chosen.Alpha(quietTint)},
				{"soft", control.Soft, display.BadgeSoft, chosen.Alpha(quietTint), chosen.Alpha(activeTint)},
				{"ghost", control.Ghost, display.BadgeGhost, ui.Transparent, chosen.Alpha(quietTint)},
				{"link", control.Link, display.BadgeLink, ui.Transparent, ui.Transparent},
			} {
				for _, mode := range []string{"button", "static-badge", "button-badge", "url-badge"} {
					t.Run(fmt.Sprintf("dark=%v/%v/%s/%s", dark, chosen, variant.name, mode), func(t *testing.T) {
						selected := chosen
						invalid, disabled, busy, calls := false, false, false, 0
						r := ui.NewTesterWith(func(c *ui.Context) {
							theme.Scope(c, tokens).Fill().Padding(20).Background(tokens.Background).Children(func() {
								if mode == "button" {
									if control.Button(c, "", control.ButtonProps{
										Variant: variant.button, Color: &selected, Size: control.Icon, Icon: icon,
										Label: "Colored action", Invalid: invalid, Disabled: disabled, Loading: busy,
									}).Clicked() {
										calls++
									}
									return
								}
								p := display.BadgeProps{
									Variant: variant.badge, Color: &selected, Icon: icon, Label: "Colored action",
									Button: mode == "button-badge", Invalid: invalid, Disabled: disabled, Busy: busy,
								}
								if mode != "static-badge" {
									p.OnActivate = func() { calls++ }
								}
								if mode == "url-badge" {
									p.URL = "https://example.com/color"
								}
								display.BadgeWith(c, "", p)
							})
						}, 180, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
						box, ok := r.Find("Colored action")
						if !ok {
							t.Fatal("missing action")
						}
						x, y := box.X+box.W/2, box.Y+box.H/2
						fillY := box.Y + 3
						pixel := func(name string, y float32, want color.RGBA, tolerance ...int) {
							t.Helper()
							got := r.Image().RGBAAt(int(x), int(y))
							limit := 1
							if len(tolerance) > 0 {
								limit = tolerance[0]
							}
							near := func(a, b uint8) bool { return int(a)-int(b) >= -limit && int(a)-int(b) <= limit }
							if !near(got.R, want.R) || !near(got.G, want.G) || !near(got.B, want.B) || got.A != want.A {
								t.Fatalf("%s pixel=%v want %v", name, got, want)
							}
						}
						checkForeground := func(fill color.RGBA) {
							t.Helper()
							if variant.button != control.Default {
								actual := r.Image().RGBAAt(int(x), int(y))
								if ratio := theme.Contrast(ui.RGB(actual.R, actual.G, actual.B), ui.RGB(fill.R, fill.G, fill.B)); ratio < 4.5 {
									t.Fatalf("unreadable colored icon: %v on %v = %.5f", actual, fill, ratio)
								}
								return
							}
							// Filled actions must keep readable text for light, dark,
							// transparent and translucent custom fills in both states.
							got := r.Image().RGBAAt(int(x), int(y))
							foreground := ui.RGB(got.R, got.G, got.B)
							if got.R != got.G || got.G != got.B || got.R != 0 && got.R != 255 ||
								theme.Contrast(foreground, ui.RGB(fill.R, fill.G, fill.B)) < 4.5 {
								t.Fatalf("unreadable filled icon=%v against %v", got, fill)
							}
						}
						base := color.RGBA{R: tokens.Background.R, G: tokens.Background.G, B: tokens.Background.B, A: 255}
						quiet := variant.quiet

						quietFill := colorOver(base, quiet)
						pixel("quiet fill", fillY, quietFill)
						checkForeground(quietFill)
						if variant.button == control.Outline {
							// Capsule edges retain a small amount of antialiasing
							// even at the center of this icon-only badge.
							pixel("colored outline", box.Y, colorOver(quietFill, chosen), 7)
						}
						activeFill := colorOver(base, variant.active)
						if mode == "static-badge" && variant.button != control.Ghost {
							activeFill = quietFill
						}
						r.Move(x, y)
						pixel("hover fill", fillY, activeFill)
						checkForeground(activeFill)
						if mode == "static-badge" {
							return
						}
						r.Press(x, y)
						pixel("pressed fill", fillY, activeFill)
						r.Release(x, y)
						r.Move(170, 90)
						r.Key(0, ui.KeyTab)
						if !r.Focused("Colored action") {
							t.Fatal("colored action lost keyboard focus")
						}
						pixel("custom focus ring", box.Y-2, colorOver(base, chosen.Alpha(focusTint)))
						r.Key(0, ui.KeyEnter)
						if calls != 2 {
							t.Fatalf("pointer and keyboard activations=%d want 2", calls)
						}
						invalid = true
						r.Frame()
						invalidBase := base
						if mode == "button" {
							// Buttons also retain their existing invalid-state shadow.
							invalidBase = colorOver(base, tokens.Destructive.Alpha(.2))
						}
						pixel("invalid focus ring", box.Y-2, colorOver(invalidBase, tokens.Destructive.Alpha(focusTint)))
						invalid, selected = false, ui.RGB(40, 180, 210)
						r.Frame()
						pixel("updated focus ring", box.Y-2, colorOver(base, selected.Alpha(focusTint)))
						updatedFill := quietFill
						if variant.button == control.Default {
							updatedFill = colorOver(base, selected)
						} else if variant.button == control.Soft {
							updatedFill = colorOver(base, selected.Alpha(quietTint))
						}
						pixel("updated fill", fillY, updatedFill)
						checkForeground(updatedFill)
						for _, loading := range []bool{false, true} {
							disabled, busy = !loading, loading
							r.Frame()
							before := r.Image().RGBAAt(int(x), int(fillY))
							r.Press(x, y)
							r.Release(x, y)
							r.Key(0, ui.KeyEnter)
							pixel("disabled/busy fill", fillY, before)
							if calls != 2 {
								t.Fatal("disabled/busy colored action activated")
							}
						}
					})
				}
			}
		}
	}
}

func colorOver(bg color.RGBA, fg ui.Color) color.RGBA {
	a := float32(fg.A) / 255
	channel := func(b, f uint8) uint8 { return uint8(float32(b)*(1-a) + float32(f)*a + .5) }
	return color.RGBA{R: channel(bg.R, fg.R), G: channel(bg.G, fg.G), B: channel(bg.B, fg.B), A: 255}
}
