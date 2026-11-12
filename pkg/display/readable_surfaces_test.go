package display_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Deliberately give surfaces very different luminance from Background. Testing
// only stock Neutral surfaces can hide an incorrect page-background fallback.
func TestColoredActionsUseTheirPaintedSurface(t *testing.T) {
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 12 12" fill="currentColor"><path d="M0 0h12v12H0z"/></svg>`))
	for _, dark := range []bool{false, true} {
		for _, provider := range []string{"card", "card-slots", "sidebar", "sidebar-layout", "popover", "explicit", "nested-alpha"} {
			for _, variant := range []struct {
				name   string
				button control.Variant
				badge  display.BadgeVariant
			}{{"soft", control.Soft, display.BadgeSoft}, {"ghost", control.Ghost, display.BadgeGhost},
				{"outline", control.Outline, display.BadgeOutline}, {"link", control.Link, display.BadgeLink}} {
				for _, badge := range []bool{false, true} {
					for _, accent := range []ui.Color{ui.RGB(170, 105, 0), ui.RGBA(120, 40, 220, .3)} {
						t.Run(fmt.Sprintf("dark=%v/%s/%s/badge=%v/alpha=%d", dark, provider, variant.name, badge, accent.A), func(t *testing.T) {
							tokens := theme.Preset("amber", dark)
							tokens.Card, tokens.Sidebar, tokens.Popover = ui.RGB(45, 57, 80), ui.RGB(235, 214, 180), ui.RGB(65, 48, 70)
							custom := ui.RGB(35, 80, 65)
							if dark {
								tokens.Card, tokens.Sidebar, tokens.Popover = ui.RGB(235, 214, 180), ui.RGB(45, 57, 80), ui.RGB(245, 235, 250)
								custom = ui.RGB(245, 235, 200)
							}
							open, selected := true, ""
							state := navigation.SidebarState{}
							mobile := false
							r := ui.NewTesterWith(func(c *ui.Context) {
								c.SetTheme(tokens.Theme())
								build := func(c *ui.Context) {
									var surface *ui.Color
									if provider == "explicit" {
										surface = &custom
									}
									if badge {
										display.BadgeWith(c, "", display.BadgeProps{Label: "Colored", Variant: variant.badge, Color: &accent, Surface: surface, Icon: icon, Button: true})
									} else {
										control.Button(c, "", control.ButtonProps{Label: "Colored", Variant: variant.button, Color: &accent, Surface: surface, Icon: icon, Size: control.Icon})
									}
								}
								ui.Column(c).Fill().Padding(24).Background(tokens.Background).Children(func() {
									switch provider {
									case "card":
										display.Card(c, func() { build(c) })
									case "card-slots":
										display.CardWith(c, display.CardProps{}, display.CardSlots{Content: build})
									case "sidebar":
										navigation.Sidebar(c, &selected, navigation.SidebarProps{}, func() { build(c) }).Height(120)
									case "sidebar-layout":
										_, err := navigation.SidebarWith(c, nil, "surface", &selected, &state,
											navigation.SidebarLayoutProps{Mobile: &mobile, HideTrigger: true, Header: build}, nil)
										if err != nil {
											t.Fatal(err)
										}
									case "popover":
										anchor := control.Button(c, "Anchor", control.ButtonProps{})
										overlay.InWindowPopover(c, anchor, &open, overlay.PopoverProps{}, func() { build(c) })
									case "explicit":
										ui.Column(c).Padding(12).Background(custom).Children(func() { build(c) })
									case "nested-alpha":
										outer, inner := ui.RGBA(0, 0, 0, .6), ui.RGBA(245, 180, 80, .2)
										ui.Column(c).Padding(12).Background(outer).Children(func() {
											theme.WithSurface(c, outer, func() {
												ui.Column(c).Padding(12).Background(inner).Children(func() { theme.WithSurface(c, inner, func() { build(c) }) })
											})
										})
									}
								})
							}, 700, 280, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
							b, ok := r.Find("Colored")
							if !ok {
								t.Fatal("missing colored action")
							}
							x, y := b.X+b.W/2, b.Y+b.H/2
							for _, state := range []string{"idle", "hover", "pressed", "focus"} {
								switch state {
								case "hover":
									r.Move(x, y)
								case "pressed":
									r.Press(x, y)
								case "focus":
									r.Release(x, y)
									r.Move(690, 270)
									for i := 0; i < 10 && !r.Focused("Colored"); i++ {
										r.Key(0, ui.KeyTab)
									}
									if !r.Focused("Colored") {
										t.Fatal("colored action lost keyboard focus")
									}
								}
								img := r.Image()
								fg, bg := img.RGBAAt(int(x), int(y)), img.RGBAAt(int(x), int(b.Y+3))
								if ratio := theme.Contrast(ui.RGB(fg.R, fg.G, fg.B), ui.RGB(bg.R, bg.G, bg.B)); ratio < 4.5 {
									t.Fatalf("%s rendered contrast %.6f fg=%v bg=%v", state, ratio, fg, bg)
								}
							}
						})
					}
				}
			}
		}
	}
}

func TestColoredFeedbackContrastDuringAnimation(t *testing.T) {
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 12 12" fill="currentColor"><path d="M0 0h12v12H0z"/></svg>`))
	for _, dark := range []bool{false, true} {
		for _, badge := range []bool{false, true} {
			t.Run(fmt.Sprintf("dark=%v/badge=%v", dark, badge), func(t *testing.T) {
				tokens := theme.Preset("amber", dark)
				r := ui.NewTester(func(c *ui.Context) {
					c.SetTheme(tokens.Theme())
					ui.Column(c).Fill().Padding(20).Background(tokens.Sidebar).Children(func() {
						theme.WithSurface(c, tokens.Sidebar, func() {
							if badge {
								display.BadgeWith(c, "", display.BadgeProps{Label: "Animated", Variant: display.BadgeSoft, Icon: icon, Button: true})
							} else {
								control.Button(c, "", control.ButtonProps{Label: "Animated", Variant: control.Soft, Icon: icon, Size: control.Icon})
							}
						})
					})
				}, 180, 100)
				b, _ := r.Find("Animated")
				x, y := b.X+b.W/2, b.Y+b.H/2
				for _, over := range []bool{true, false} {
					if over {
						r.Move(x, y)
					} else {
						r.Move(170, 90)
					}
					for i := 0; i < 20; i++ {
						r.Advance(10 * time.Millisecond)
						img := r.Image()
						fg, bg := img.RGBAAt(int(x), int(y)), img.RGBAAt(int(x), int(b.Y+3))
						if ratio := theme.Contrast(ui.RGB(fg.R, fg.G, fg.B), ui.RGB(bg.R, bg.G, bg.B)); ratio < 4.5 {
							t.Fatalf("hover=%v frame=%d contrast=%.6f", over, i, ratio)
						}
					}
				}
			})
		}
	}
}

// At opposite-polarity endpoints no single foreground can pass. Native color
// interpolation must be skipped for this rare custom-surface case.
func TestColoredFeedbackOnOppositePolarityStates(t *testing.T) {
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 12 12" fill="currentColor"><path d="M0 0h12v12H0z"/></svg>`))
	for _, kind := range []string{"button", "badge", "sidebar"} {
		t.Run(kind, func(t *testing.T) {
			tokens := theme.Dark()
			tokens.Background = ui.RGB(80, 80, 80)
			tokens.Sidebar = tokens.Background
			accent := ui.RGB(255, 255, 255)
			r := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(tokens.Theme())
				ui.Column(c).Fill().Padding(20).Background(tokens.Background).Children(func() {
					switch kind {
					case "button":
						control.Button(c, "", control.ButtonProps{Label: "Rare", Variant: control.Soft, Color: &accent, Size: control.Icon, Icon: icon})
					case "badge":
						display.BadgeWith(c, "", display.BadgeProps{Label: "Rare", Variant: display.BadgeSoft, Color: &accent, Button: true, Icon: icon})
					case "sidebar":
						navigation.SidebarMenuButton(c, navigation.SidebarMenuButtonProps{Label: "Rare", Color: &accent, Active: true}, func(c *ui.Context) { ui.Icon(c, icon).Size(12, 12) }).Justify(ui.Center)
					}
				})
			}, 180, 100)
			b, _ := r.Find("Rare")
			x, y := b.X+b.W/2, b.Y+b.H/2
			for _, over := range []bool{true, false} {
				if over {
					r.Move(x, y)
				} else {
					r.Move(170, 90)
				}
				for i := 0; i < 20; i++ {
					r.Advance(10 * time.Millisecond)
					img := r.Image()
					fg, bg := img.RGBAAt(int(x), int(y)), img.RGBAAt(int(x), int(b.Y+3))
					if ratio := theme.Contrast(ui.RGB(fg.R, fg.G, fg.B), ui.RGB(bg.R, bg.G, bg.B)); ratio < 4.5 {
						t.Fatalf("over=%v frame=%d contrast=%.6f", over, i, ratio)
					}
				}
			}
		})
	}
}

func TestColoredTextHonorsHighContrastPreference(t *testing.T) {
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 12 12" fill="currentColor"><path d="M0 0h12v12H0z"/></svg>`))
	for _, kind := range []string{"button", "badge", "sidebar"} {
		t.Run(kind, func(t *testing.T) {
			controller := theme.Controller{Mode: theme.DarkMode, Preset: "amber"}
			r := ui.NewTesterWith(func(c *ui.Context) {
				controller.Apply(c)
				tokens := theme.TokensOf(c)
				ui.Column(c).Fill().Padding(20).Background(tokens.Sidebar).Children(func() {
					theme.WithSurface(c, tokens.Sidebar, func() {
						switch kind {
						case "button":
							control.Button(c, "", control.ButtonProps{Label: "High contrast", Variant: control.Soft, Size: control.Icon, Icon: icon})
						case "badge":
							display.BadgeWith(c, "", display.BadgeProps{Label: "High contrast", Variant: display.BadgeSoft, Button: true, Icon: icon})
						case "sidebar":
							navigation.SidebarMenuButton(c, navigation.SidebarMenuButtonProps{Label: "High contrast", Color: &tokens.Primary, Active: true}, func(c *ui.Context) { ui.Icon(c, icon).Size(12, 12) }).Justify(ui.Center)
						}
					})
				})
			}, 180, 100, ui.TesterOptions{Preferences: ui.Preferences{HighContrast: true, ReduceMotion: true}})
			b, _ := r.Find("High contrast")
			x, y := b.X+b.W/2, b.Y+b.H/2
			for _, state := range []string{"idle", "hover", "focus"} {
				if state == "hover" {
					r.Move(x, y)
				}
				if state == "focus" {
					r.Move(170, 90)
					r.Key(0, ui.KeyTab)
				}
				img := r.Image()
				fg, bg := img.RGBAAt(int(x), int(y)), img.RGBAAt(int(x), int(b.Y+3))
				if ratio := theme.Contrast(ui.RGB(fg.R, fg.G, fg.B), ui.RGB(bg.R, bg.G, bg.B)); ratio < 7 {
					t.Fatalf("%s contrast=%.5f", state, ratio)
				}
			}
		})
	}
}
