package display

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestBadgeLegacyUsesVegaHeightAndSoftPrimaryFill(t *testing.T) {
	var reference ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		tokens := theme.TokensOf(c)
		ui.Column(c).Fill().Padding(20).Gap(12).Background(tokens.Background).Children(func() {
			Badge(c, "Soft", BadgeSoft).Label("Soft badge")
			ui.Box(c).Size(60, 20).Background(tokens.Primary.Alpha(.1)).Draw(func(_ *ui.Painter, bounds ui.Rect) { reference = bounds })
		})
	}, 240, 120)
	box, ok := r.Find("Soft badge")
	if !ok || box.H != 20 {
		t.Fatalf("Vega badge should be exactly 20 DIPs high: %+v", box)
	}
	image := r.Image()
	got := image.RGBAAt(int(box.X+box.W/2), int(box.Y+3))
	want := image.RGBAAt(int(reference.X+reference.W/2), int(reference.Y+3))
	if got != want {
		t.Fatalf("soft fill should use the theme primary tone: got=%v want=%v", got, want)
	}
}

func TestBadgeAllVariantsFitTextAndUseThemedTones(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprintf("dark=%v", dark), func(t *testing.T) {
			var tokens *theme.Tokens
			r := ui.NewTester(func(c *ui.Context) {
				appearance := theme.Controller{Mode: theme.LightMode}
				if dark {
					appearance.Mode = theme.DarkMode
				}
				appearance.Apply(c)
				tokens = theme.TokensOf(c)
				ui.Column(c).Fill().Padding(20).Gap(12).Background(tokens.Background).Children(func() {
					for _, variant := range []BadgeVariant{BadgeDefault, BadgeSecondary, BadgeOutline, BadgeSoft, BadgeGhost, BadgeLink} {
						BadgeWith(c.Key(variant), "Ready", BadgeProps{Variant: variant}).Label(fmt.Sprint("badge-", variant))
					}
				})
			}, 300, 250)
			want := []ui.Color{tokens.Primary, tokens.Secondary, tokens.Background, tokens.Primary.Alpha(.1), tokens.Background, tokens.Background}
			if dark {
				want[3] = tokens.Primary.Alpha(.2)
			}
			for variant, tone := range want {
				box, ok := r.Find(fmt.Sprint("badge-", variant))
				if !ok || box.H != 20 || box.W <= 20 || box.W >= 150 {
					t.Fatalf("badge should fit content rather than stretch: variant=%v box=%+v", variant, box)
				}
				actual := r.Image().RGBAAt(int(box.X+box.W/2), int(box.Y+3))
				alpha := float32(tone.A) / 255
				expected := color.RGBA{
					R: uint8(math.Round(float64(float32(tone.R)*alpha + float32(tokens.Background.R)*(1-alpha)))),
					G: uint8(math.Round(float64(float32(tone.G)*alpha + float32(tokens.Background.G)*(1-alpha)))),
					B: uint8(math.Round(float64(float32(tone.B)*alpha + float32(tokens.Background.B)*(1-alpha)))), A: 255,
				}
				if actual != expected {
					t.Fatalf("variant=%v color=%v want=%v", variant, actual, expected)
				}
			}
		})
	}
}

func TestBadgeSoftColorsAndInteraction(t *testing.T) {
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
			for _, mode := range []string{"static", "button", "url"} {
				t.Run(fmt.Sprintf("dark=%v/%s/%s", dark, tc.name, mode), func(t *testing.T) {
					calls := 0
					p := BadgeProps{Variant: BadgeSoft, Color: tc.color, Icon: badgeTestIcon, Label: "Soft badge", Button: mode == "button"}
					if mode != "static" {
						p.OnActivate = func() { calls++ }
					}
					if mode == "url" {
						p.URL = "https://example.com/soft"
					}
					r := ui.NewTesterWith(func(c *ui.Context) {
						theme.Scope(c, tokens).Fill().Padding(20).Background(tokens.Background).Children(func() {
							BadgeWith(c, "", p)
						})
					}, 180, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
					box, ok := r.Find(p.Label)
					if !ok {
						t.Fatal("missing Soft badge")
					}
					x, y := box.X+box.W/2, box.Y+box.H/2
					base := color.RGBA{R: tokens.Background.R, G: tokens.Background.G, B: tokens.Background.B, A: 255}
					over := func(bg color.RGBA, fg ui.Color) color.RGBA {
						a := float32(fg.A) / 255
						channel := func(b, f uint8) uint8 { return uint8(float32(b)*(1-a) + float32(f)*a + .5) }
						return color.RGBA{R: channel(bg.R, fg.R), G: channel(bg.G, fg.G), B: channel(bg.B, fg.B), A: 255}
					}
					pixel := func(name string, y float32, want color.RGBA) {
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
					quiet := over(base, chosen.Alpha(quietAlpha))
					pixel("quiet fill", box.Y+3, quiet)
					iconPixel := r.Image().RGBAAt(int(x), int(y))
					if theme.Contrast(ui.RGB(iconPixel.R, iconPixel.G, iconPixel.B), ui.RGB(quiet.R, quiet.G, quiet.B)) < 4.5 {
						t.Fatal("unreadable inherited icon")
					}
					r.Move(x, y)
					if mode == "static" {
						pixel("static hover fill", box.Y+3, quiet)
					} else {
						pixel("hover fill", box.Y+3, over(base, chosen.Alpha(activeAlpha)))
						r.Press(x, y)
						pixel("pressed fill", box.Y+3, over(base, chosen.Alpha(activeAlpha)))
						r.Release(x, y)
						r.Move(170, 90)
						r.Key(0, ui.KeyTab)
						if !r.Focused(p.Label) {
							t.Fatal("Soft badge lost keyboard focus")
						}
						pixel("custom focus ring", box.Y-2, over(base, chosen.Alpha(focusAlpha)))
						r.Key(0, ui.KeyEnter)
						if calls != 2 {
							t.Fatalf("pointer and keyboard activations=%d want 2", calls)
						}
						p.Invalid = true
						r.Frame()
						pixel("invalid focus ring", box.Y-2, over(base, tokens.Destructive.Alpha(focusAlpha)))
						p.Invalid = false
					}
					p.Color = new(ui.RGB(40, 180, 210))
					r.Frame()
					pixel("updated fill", box.Y+3, over(base, p.Color.Alpha(quietAlpha)))
					if mode != "static" {
						pixel("updated focus ring", box.Y-2, over(base, p.Color.Alpha(focusAlpha)))
					}
				})
			}
		}
	}
}

func TestBadgeLogicalIconSlotsAndSingleAccessibleName(t *testing.T) {
	r := displayTester(func(c *ui.Context) {
		BadgeWith(c, "Verified", BadgeProps{Label: "Verified account", Description: "Identity confirmed", Button: true,
			Leading:  func(child *ui.Context) { ui.Box(child).Fill().Background(ui.RGB(255, 0, 0)).Label("Start slot") },
			Trailing: func(child *ui.Context) { ui.Box(child).Fill().Background(ui.RGB(0, 0, 255)).Label("End slot") },
		})
	})
	button, ok := r.AccessByRole(ui.RoleButton, "Verified account")
	if !ok || button.Description != "Identity confirmed" {
		t.Fatalf("native badge button metadata=%+v", button)
	}
	start, _ := r.Find("Start slot")
	end, _ := r.Find("End slot")
	if start.W != 12 || start.H != 12 || start.X-button.Bounds.X != 7 || button.Bounds.X+button.Bounds.W-end.X-end.W != 7 || start.X >= end.X {
		t.Fatalf("logical icon metrics button=%+v start=%+v end=%+v", button.Bounds, start, end)
	}
	for _, node := range r.AccessAll() {
		if node.Name == "Start slot" || node.Name == "End slot" || node.Name == "Verified" {
			t.Fatalf("badge exposed decorative child accessibility: %+v", node)
		}
	}
	r.SetDirection(ui.RightToLeft)
	start, _ = r.Find("Start slot")
	end, _ = r.Find("End slot")
	button, _ = r.AccessByRole(ui.RoleButton, "Verified account")
	if start.X <= end.X || button.Bounds.X+button.Bounds.W-start.X-start.W != 7 || end.X-button.Bounds.X != 7 {
		t.Fatalf("logical slots did not flip in RTL: button=%+v start=%+v end=%+v", button.Bounds, start, end)
	}
}

func TestBadgeButtonNativeKeyboardDisabledBusyAndChanged(t *testing.T) {
	count, changes := 0, 0
	var p = BadgeProps{Label: "Status action", Variant: BadgeSoft, Color: new(ui.RGB(20, 160, 80)), Button: true, OnActivate: func() { count++ }}
	r := displayTester(func(c *ui.Context) {
		if BadgeWith(c, "Review", p).Changed() {
			changes++
		}
		ui.Button(c, "After badge")
	})
	r.Key(0, ui.KeyTab)
	if !r.Focused("Status action") {
		t.Fatal("button badge was skipped by native Tab")
	}
	r.Key(0, ui.KeySpace)
	r.Key(0, ui.KeyEnter)
	if count != 2 || changes != 2 {
		t.Fatalf("native activation notifications=%d changes=%d", count, changes)
	}
	p.Disabled = true
	r.Frame()
	displayClick(t, r, "Status action")
	if count != 2 {
		t.Fatal("disabled badge activated")
	}
	p.Disabled, p.Busy = false, true
	r.Frame()
	info, _ := r.AccessByRole(ui.RoleButton, "Status action")
	if !info.Busy || !info.Disabled {
		t.Fatalf("busy badge did not suppress activation: %+v", info)
	}
	p.Busy = false
	r.Frame()
	displayClick(t, r, "Status action")
	if count != 3 || changes != 3 {
		t.Fatalf("ready badge did not recover: count=%d changes=%d", count, changes)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("settled badge kept requesting frames")
	}
}

func TestBadgeUsesActualNativeLinkWithURLAndEnterOnly(t *testing.T) {
	activated, changes := 0, 0
	var handle ui.Handle
	r := displayTester(func(c *ui.Context) {
		if BadgeWith(c, "Docs", BadgeProps{URL: "https://example.com/docs", Variant: BadgeSecondary, Label: "Open documentation", Handle: &handle,
			OnActivate: func() { activated++ }, EndIcon: badgeTestIcon,
		}).Changed() {
			changes++
		}
	})
	link, ok := r.AccessByRole(ui.RoleLink, "Open documentation")
	if !ok || link.Bounds.H != 20 {
		t.Fatalf("native link present=%v size=%vx%v", ok, link.Bounds.W, link.Bounds.H)
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("Open documentation") {
		t.Fatal("URL badge did not provide a native focus target")
	}
	r.Key(0, ui.KeySpace)
	if len(r.OpenedURLs()) != 0 || activated != 0 {
		t.Fatal("link took Space instead of preserving native page-scroll semantics")
	}
	r.Key(0, ui.KeyEnter)
	if len(r.OpenedURLs()) != 1 || r.OpenedURLs()[0] != "https://example.com/docs" || activated != 1 || changes != 1 {
		t.Fatalf("native URL dispatch changed=%d activated=%d urls=%v", changes, activated, r.OpenedURLs())
	}
	displayClick(t, r, "Open documentation")
	if len(r.OpenedURLs()) != 2 || activated != 2 || changes != 2 {
		t.Fatalf("pointer URL dispatch changed=%d activated=%d urls=%v", changes, activated, r.OpenedURLs())
	}
}

var badgeTestIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12" fill="currentColor"><path d="M0 0h12v12H0z"/></svg>`))

func TestBadgeLongTextClipsWithoutWrappingAndEmptySlotsKeepSiblingIdentity(t *testing.T) {
	show, draft := false, "Draft"
	r := displayTester(func(c *ui.Context) {
		BadgeWith(c, strings.Repeat("Long badge ", 8), BadgeProps{Label: "Long badge", Variant: BadgeOutline}).Width(100)
		BadgeWith(c, "", BadgeProps{Content: func(child *ui.Context) {
			if show {
				ui.Text(child, "Late content")
			}
		}, Leading: func(*ui.Context) {}, Trailing: func(*ui.Context) {}})
		ui.TextInput(c, &draft).Label("Following editor")
	})
	box, _ := r.Find("Long badge")
	if box.W != 100 || box.H != 20 {
		t.Fatalf("bounded badge wrapped or changed fixed height: %+v", box)
	}
	displayClick(t, r, "Following editor")
	show = true
	r.Frame()
	if !r.Focused("Following editor") {
		t.Fatal("empty callback leaked a key into the next sibling")
	}
}

func TestBadgeNativeLinkResolvesRelativeRouterTargets(t *testing.T) {
	router := ui.NewRouter("/notes/42")
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		router.View(c, func(route *ui.Route) {
			ui.Column(c).Fill().Padding(20).Children(func() {
				BadgeWith(c, "Edit note", BadgeProps{URL: "edit", Icon: badgeTestIcon})
				ui.Text(c, route.Path())
			})
		})
	}, 400, 200)
	displayClick(t, r, "Edit note")
	if router.Path() != "/notes/edit" || len(r.OpenedURLs()) != 0 {
		t.Fatalf("relative badge target bypassed native Router: path=%q URLs=%v", router.Path(), r.OpenedURLs())
	}
}

func TestBadgeDisabledAncestorSuppressesNativeLinkAndCallbacks(t *testing.T) {
	disabled, calls := true, 0
	r := displayTester(func(c *ui.Context) {
		ui.Row(c).Disabled(disabled).Children(func() {
			BadgeWith(c, "Open account", BadgeProps{URL: "https://example.com/account", OnActivate: func() { calls++ }})
		})
	})
	displayClick(t, r, "Open account")
	if calls != 0 || len(r.OpenedURLs()) != 0 {
		t.Fatal("native link or callback ignored a disabled ancestor")
	}
	disabled = false
	r.Frame()
	displayClick(t, r, "Open account")
	if calls != 1 || len(r.OpenedURLs()) != 1 {
		t.Fatalf("enabled ancestor did not restore URL action: calls=%d URLs=%v", calls, r.OpenedURLs())
	}
}

func TestBadgeHoverFocusInvalidAndReducedMotionPixels(t *testing.T) {
	for _, url := range []bool{false, true} {
		t.Run(fmt.Sprint("link=", url), func(t *testing.T) {
			p := BadgeProps{Label: "Review badge", Variant: BadgeGhost, Button: true}
			if url {
				p.URL = "https://example.com/review"
			}
			tokens := theme.Preset("neutral", false)
			tokens.Ring, tokens.Destructive = ui.RGB(210, 0, 210), ui.RGB(220, 0, 0)
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				theme.Scope(c, tokens).Fill().Children(func() {
					ui.Column(c).Fill().Padding(20).Gap(16).Background(tokens.Background).Children(func() {
						BadgeWith(c, "Review", p)
						ui.Button(c, "After review")
					})
				})
			}, 240, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			box, _ := r.Find("Review badge")
			before := r.Image().RGBAAt(int(box.X+box.W/2), int(box.Y+3))
			r.Move(box.X+box.W/2, box.Y+box.H/2)
			hovered := r.Image().RGBAAt(int(box.X+box.W/2), int(box.Y+3))
			if before == hovered || hovered.R != tokens.Muted.R || hovered.G != tokens.Muted.G || hovered.B != tokens.Muted.B {
				t.Fatalf("ghost hover did not paint themed muted fill: before=%v hover=%v", before, hovered)
			}
			r.Move(230, 90)
			r.Key(0, ui.KeyTab)
			if !r.Focused("Review badge") {
				t.Fatal("keyboard focus never reached badge target")
			}
			ring := r.Image().RGBAAt(int(box.X+box.W/2), int(box.Y+box.H+1))
			if ring.R <= ring.G || ring.B <= ring.G {
				t.Fatalf("keyboard focus ring was not painted: %v", ring)
			}
			p.Invalid = true
			r.Frame()
			invalidRing := r.Image().RGBAAt(int(box.X+box.W/2), int(box.Y+box.H+1))
			if invalidRing.R <= invalidRing.B || invalidRing.R <= invalidRing.G {
				t.Fatalf("invalid state did not use destructive ring: %v", invalidRing)
			}
			role := ui.RoleButton
			if url {
				role = ui.RoleLink
			}
			info, _ := r.AccessByRole(role, "Review badge")
			if !info.Invalid {
				t.Fatal("invalid state missing from actual native target")
			}
			r.Key(0, ui.KeyTab)
			unfocused := r.Image().RGBAAt(int(box.X+box.W/2), int(box.Y+box.H+1))
			if unfocused != (color.RGBA{R: tokens.Background.R, G: tokens.Background.G, B: tokens.Background.B, A: 255}) {
				t.Fatalf("unfocused invalid badge painted a focus-only ring: %v", unfocused)
			}
			if r.FramePending() {
				t.Fatal("reduced-motion badge scheduled idle transition frames")
			}
		})
	}
}

func TestBadgeIconsInheritVariantAndReturnedElementColors(t *testing.T) {
	var fg ui.Color
	r := displayTester(func(c *ui.Context) {
		fg = theme.TokensOf(c).PrimaryForeground
		BadgeWith(c, "", BadgeProps{Label: "Default mark", Icon: badgeTestIcon})
		BadgeWith(c, "", BadgeProps{Label: "Custom mark", Icon: badgeTestIcon}).Background(ui.RGB(10, 50, 90)).TextColor(ui.RGB(70, 220, 100))
	})
	standard, _ := r.Find("Default mark")
	custom, _ := r.Find("Custom mark")
	paint := r.Image()
	got := paint.RGBAAt(int(standard.X+13), int(standard.Y+10))
	if got != (color.RGBA{R: fg.R, G: fg.G, B: fg.B, A: fg.A}) {
		t.Fatalf("default icon did not inherit variant foreground: %v", got)
	}
	got = paint.RGBAAt(int(custom.X+13), int(custom.Y+10))
	if got != (color.RGBA{R: 70, G: 220, B: 100, A: 255}) {
		t.Fatalf("returned Element color did not propagate into icon: %v", got)
	}
}

func TestBadgeLinkHoverActuallyPaintsUnderlineForDefaultAndCustomText(t *testing.T) {
	for _, mode := range []string{"static", "button", "url"} {
		for _, custom := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/custom=%v", mode, custom), func(t *testing.T) {
				p := BadgeProps{Label: "Docs badge", Variant: BadgeLink, Button: mode == "button"}
				if mode == "url" {
					p.URL = "https://example.com/docs"
				}
				text := "Docs"
				if custom {
					text = ""
					p.Content = func(child *ui.Context) { ui.Text(child, "Docs") }
				}
				r := ui.NewTesterWith(func(c *ui.Context) {
					theme.Apply(c)
					ui.Column(c).Fill().Padding(20).Background(theme.TokensOf(c).Background).Children(func() { BadgeWith(c, text, p) })
				}, 240, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				box, _ := r.Find("Docs badge")
				before := image.NewRGBA(r.Image().Bounds())
				copy(before.Pix, r.Image().Pix)
				r.Move(box.X+box.W/2, box.Y+box.H/2)
				changed := []image.Point{}
				for y := int(box.Y); y < int(box.Y+box.H); y++ {
					for x := int(box.X); x < int(box.X+box.W); x++ {
						if before.RGBAAt(x, y) != r.Image().RGBAAt(x, y) {
							changed = append(changed, image.Pt(x, y))
						}
					}
				}
				if len(changed) < 4 {
					t.Fatalf("hovered Link badge did not paint a visible underline: %d pixels changed", len(changed))
				}
				for _, point := range changed {
					if point.Y < int(box.Y+box.H/2) {
						t.Fatalf("hover changed glyph layout instead of adding an underline: %+v", point)
					}
				}
				r.Move(230, 90)
				for _, point := range changed {
					if before.RGBAAt(point.X, point.Y) != r.Image().RGBAAt(point.X, point.Y) {
						t.Fatal("hover underline remained after pointer leave")
					}
				}
				if r.FramePending() {
					t.Fatal("reduced-motion hover decoration scheduled idle frames")
				}
			})
		}
	}
}
