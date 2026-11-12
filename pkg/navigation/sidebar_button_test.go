package navigation_test

import (
	"bytes"
	"fmt"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var menuGlyph = ui.MustParseSVG([]byte(`<svg viewBox="0 0 16 16" fill="currentColor"><path d="M0 0h16v16H0z"/></svg>`))

func menuPixel(r *ui.Tester, b ui.Rect) ui.Color {
	p := r.Image().RGBAAt(int(b.X+b.W/2), int(b.Y+b.H/2))
	return ui.RGB(p.R, p.G, p.B)
}

func menuBounds(t *testing.T, r *ui.Tester, name string) ui.Rect {
	t.Helper()
	b, ok := r.Find(name)
	if !ok {
		t.Fatalf("missing %q", name)
	}
	return b
}

func hasTextColor(r *ui.Tester, b ui.Rect, want ui.Color) bool {
	img := r.Image()
	for y := int(b.Y); y < int(b.Y+b.H); y++ {
		for x := int(b.X); x < int(b.X+b.W); x++ {
			p := img.RGBAAt(x, y)
			if p == (color.RGBA{R: want.R, G: want.G, B: want.B, A: 255}) {
				return true
			}
		}
	}
	return false
}

// All official presets use the real renderer and native input; label, icon and
// inert Badge must share one state foreground rather than resolving separately.
func TestSidebarMenuButtonRenderedForegroundAndTokens(t *testing.T) {
	for _, preset := range theme.Presets() {
		for _, dark := range []bool{false, true} {
			for _, mode := range []string{"official", "official-idle", "active-color", "pro", "free"} {
				t.Run(fmt.Sprintf("%s/dark=%v/%s", preset, dark, mode), func(t *testing.T) {
					tokens := theme.Preset(preset, dark)
					p := navigation.SidebarMenuButtonProps{Label: "Menu action", Active: mode == "active-color" || mode == "official"}
					if mode == "active-color" || mode == "pro" {
						p.Color = &tokens.Primary
					}
					p.Muted = mode == "free"
					r := ui.NewTesterWith(func(c *ui.Context) {
						c.SetTheme(tokens.Theme())
						ui.Column(c).Fill().Padding(20).Background(tokens.Sidebar).Children(func() {
							navigation.SidebarMenuButton(c, p, func(c *ui.Context) {
								ui.Icon(c, menuGlyph).Size(12, 12).Role(ui.RoleImage).Label("Menu icon")
								ui.Text(c, "Menu label").SingleLine()
								display.BadgeWith(c, "Pro", display.BadgeProps{Variant: display.BadgeOutline, InheritForeground: true, Label: "Inline badge"})
							})
						})
					}, 320, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
					button := menuBounds(t, r, p.Label)
					x, y := button.X+button.W/2, button.Y+button.H/2
					for _, state := range []string{"idle", "hover", "pressed", "focus"} {
						switch state {
						case "hover":
							r.Move(x, y)
						case "pressed":
							r.Press(x, y)
						case "focus":
							r.Release(x, y)
							r.Move(310, 90)
							r.Key(0, ui.KeyTab)
						}
						icon := menuPixel(r, menuBounds(t, r, "Menu icon"))
						fillPixel := r.Image().RGBAAt(int(x), int(button.Y+3))
						fill := ui.RGB(fillPixel.R, fillPixel.G, fillPixel.B)
						if p.Color != nil {
							if ratio := theme.Contrast(icon, fill); ratio < 4.5 {
								t.Fatalf("%s colored text contrast=%.5f fg=%v bg=%v", state, ratio, icon, fill)
							}
						} else {
							want := tokens.SidebarAccentForeground
							if mode == "official-idle" && (state == "idle" || state == "focus") {
								want = tokens.SidebarForeground
							}
							if mode == "free" {
								want = tokens.MutedForeground
								if state != "idle" {
									want = tokens.Foreground
								}
							}
							if icon != want {
								t.Fatalf("%s token foreground=%v want=%v", state, icon, want)
							}
						}
						for _, label := range []string{"Menu label", "Pro"} {
							if !hasTextColor(r, menuBounds(t, r, label), icon) {
								t.Fatalf("%s %s does not inherit icon foreground %v", state, label, icon)
							}
						}
					}
					access, ok := r.AccessByRole(ui.RoleButton, p.Label)
					if !ok || !access.Focusable || access.Current != ui.CurrentPage && p.Active {
						t.Fatalf("lost native/current semantics: %+v", access)
					}
				})
			}
		}
	}
}

func TestSidebarMenuButtonCompositionDisabledAndSizing(t *testing.T) {
	p := navigation.SidebarMenuButtonProps{Label: "Custom layout", FontSize: 20, Active: true}
	calls := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() {
			navigation.SidebarMenuButton(c, p, func(c *ui.Context) {
				ui.Box(c).Width(28).Shrink(0).Center().Children(func() { ui.Icon(c, menuGlyph).Size(16, 16).Label("Fixed icon") })
				ui.Text(c, "Centered label").Basis(0).Grow(1).TextAlign(ui.Center).SingleLine()
			}).FillWidth().Gap(0).OnClick(func() { calls++ })
		})
	}, 360, 120, ui.TesterOptions{Preferences: ui.Preferences{TextScale: 1.5, ReduceMotion: true}})
	b := menuBounds(t, r, p.Label)
	if b.H < 30*20/14+8 {
		t.Fatalf("scaled custom text clipped: %v", b)
	}
	icon, label := menuBounds(t, r, "Fixed icon"), menuBounds(t, r, "Centered label")
	if icon.X+icon.W > label.X || label.W <= icon.W {
		t.Fatalf("caller layout did not keep separate regions: icon=%v label=%v", icon, label)
	}
	if err := r.Click(p.Label); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyEnter)
	r.Key(0, ui.KeySpace)
	if calls != 3 {
		t.Fatalf("native pointer/Enter/Space activations=%d", calls)
	}
	p.Disabled = true
	r.Frame()
	a, _ := r.AccessByRole(ui.RoleButton, p.Label)
	if !a.Disabled || a.Focusable || a.Current != ui.CurrentPage {
		t.Fatalf("disabled semantics=%+v", a)
	}
	_ = r.Click(p.Label)
	r.Key(0, ui.KeyEnter)
	if calls != 3 {
		t.Fatal("disabled menu activated")
	}
}

func TestSidebarMenuButtonRespectsMotionPreferenceChanges(t *testing.T) {
	tokens := theme.Preset("amber", true)
	active := false
	view := func(c *ui.Context) {
		c.SetTheme(tokens.Theme())
		ui.Column(c).Fill().Padding(20).Background(tokens.Sidebar).Children(func() {
			navigation.SidebarMenuButton(c, navigation.SidebarMenuButtonProps{Label: "Animated", Active: active, Color: &tokens.Primary, Icon: menuGlyph}, nil)
		})
	}
	r := ui.NewTester(view, 220, 100)
	active = true
	r.Frame()
	r.Advance(30 * time.Millisecond)
	if !r.FramePending() {
		t.Fatal("normal feedback did not animate")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	ref := ui.NewTesterWith(view, 220, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !bytes.Equal(r.Image().Pix, ref.Image().Pix) || r.FramePending() {
		t.Fatal("reduced motion did not stop running feedback")
	}
	r.SetPreferences(ui.Preferences{})
	active = false
	r.Frame()
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("feedback animation did not settle")
	}
}
