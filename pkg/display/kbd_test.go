package display

import (
	"fmt"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestKbdUsesVegaHeightMinimumWidthAndMutedSurface(t *testing.T) {
	var tokens *theme.Tokens
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		tokens = theme.TokensOf(c)
		ui.Column(c).Width(300).Padding(20).Background(tokens.Background).Children(func() { Kbd(c, "K").Label("Key cap") })
	}, 400, 150)
	box, ok := r.Find("Key cap")
	if !ok || box.W != 20 || box.H != 20 {
		t.Fatalf("source key cap should be content-sized with 20DIP minimum: %+v", box)
	}
	pixel := r.Image().RGBAAt(int(box.X+1), int(box.Y+box.H/2))
	if pixel.R != tokens.Muted.R || pixel.G != tokens.Muted.G || pixel.B != tokens.Muted.B {
		t.Fatalf("key cap should have muted fill through its side rather than a border: %v", pixel)
	}
}

var kbdSolidGlyph = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path fill="currentColor" d="M0 0h16v16H0z"/></svg>`))

func TestKbdGroupSpacingIconsAndNativeTextSemantics(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() {
			KbdGroup(c, func(child *ui.Context) {
				Kbd(child, "Ctrl").Label("Control key")
				KbdWith(child, "K", KbdProps{Label: "Key with icon", Icon: kbdSolidGlyph}).Label("Key with icon")
				KbdWith(child, "", KbdProps{Label: "Arrow key", Content: func(inner *ui.Context) { ui.Icon(inner, kbdSolidGlyph).Size(12, 12).Label("Visual glyph") }})
			})
		})
	}, 500, 200)
	control, _ := r.Find("Control key")
	key, _ := r.Find("Key with icon")
	glyph, _ := r.Find("Visual glyph")
	if key.X-control.X-control.W != 4 || glyph.W != 12 || glyph.H != 12 {
		t.Fatalf("source group/icon layout control=%+v key=%+v glyph=%+v", control, key, glyph)
	}
	for _, name := range []string{"Control key", "Key with icon", "Arrow key"} {
		item, ok := r.AccessByRole(ui.RoleText, name)
		if !ok || item.Focusable || item.Disabled {
			t.Fatalf("keycap should be plain native text, not an extra disabled/focusable control: %+v", item)
		}
	}
	for _, node := range r.AccessAll() {
		if node.Name == "Visual glyph" {
			t.Fatal("inert keycap visuals duplicated accessibility")
		}
	}
}

func TestKbdInButtonPreservesRealPointerAndKeyboardActivation(t *testing.T) {
	activated, visualClicks := 0, 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(12).Children(func() {
			button := ui.ButtonBase(c).Height(36).Label("Run action").Children(func() {
				KbdWith(c, "K", KbdProps{Content: func(child *ui.Context) {
					visual := ui.ButtonBase(child).Size(18, 16).Label("Inert visual").Children(func() { ui.Text(child, "K") })
					if visual.Clicked() {
						visualClicks++
					}
				}}).AlignSelf(ui.Center).Label("Key inside button")
			})
			if button.Clicked() {
				activated++
			}
			ui.Button(c, "Continue")
		})
	}, 400, 200)
	cap, _ := r.Find("Key inside button")
	r.ClickAt(cap.X+cap.W/2, cap.Y+cap.H/2)
	r.Key(0, ui.KeySpace)
	if activated != 2 || visualClicks != 0 {
		t.Fatalf("inert cap blocked parent or activated its visual: parent=%d visual=%d", activated, visualClicks)
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("Continue") {
		t.Fatal("custom keycap created an extra Tab target")
	}
	for _, node := range r.AccessAll() {
		if node.Name == "Inert visual" {
			t.Fatal("hidden keycap child appeared to accessibility")
		}
	}
}

func TestKbdTooltipColorsAreCompositedFromCurrentScopedTokens(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprintf("dark=%v", dark), func(t *testing.T) {
			var tokens *theme.Tokens
			r := ui.NewTester(func(c *ui.Context) {
				controller := theme.Controller{Mode: theme.LightMode}
				if dark {
					controller.Mode = theme.DarkMode
				}
				controller.Apply(c)
				tokens = theme.TokensOf(c)
				ui.Column(c).Fill().Padding(20).Background(tokens.Foreground).Children(func() { KbdWith(c, "S", KbdProps{InTooltip: true}).Label("Tooltip key") })
			}, 200, 100)
			bounds, _ := r.Find("Tooltip key")
			alpha := float32(.2)
			if dark {
				alpha = .1
			}
			foreground := tokens.Background.Alpha(alpha)
			blend := func(b, f uint8) uint8 {
				return uint8((int(f)*int(foreground.A) + int(b)*(255-int(foreground.A)) + 127) / 255)
			}
			pixel := r.Image().RGBAAt(int(bounds.X+1), int(bounds.Y+bounds.H/2))
			if pixel.R != blend(tokens.Foreground.R, foreground.R) || pixel.G != blend(tokens.Foreground.G, foreground.G) || pixel.B != blend(tokens.Foreground.B, foreground.B) {
				t.Fatalf("source tooltip key tint=%v", pixel)
			}
			r.Advance(time.Second)
			if r.FramePending() {
				t.Fatal("static keycap requested idle frames")
			}
		})
	}
}

func TestKbdNativeModifierNamesAndShortcutGroups(t *testing.T) {
	all := ui.Ctrl | ui.Alt | ui.Shift | ui.Super
	for _, test := range []struct {
		platform string
		want     []string
	}{{"darwin", []string{"⌃", "⌥", "⇧", "⌘"}}, {"windows", []string{"Ctrl", "Alt", "Shift", "Win"}}, {"linux", []string{"Ctrl", "Alt", "Shift", "Super"}}} {
		if got := kbdModifiers(all, test.platform); !slices.Equal(got, test.want) {
			t.Fatalf("%s physical modifier names=%v", test.platform, got)
		}
	}
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); KbdShortcut(c, ui.Cmd|ui.Shift, "P") }, 350, 100)
	for _, key := range append(kbdModifiers(ui.Cmd|ui.Shift, runtime.GOOS), "P") {
		if !r.HasText(key) {
			t.Fatalf("platform shortcut omitted %q", key)
		}
	}
}

func TestKbdEmptyCustomContentKeepsFollowingEditorIdentity(t *testing.T) {
	draft := "draft"
	show := false
	created := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(8).Children(func() {
			KbdGroup(c, func(child *ui.Context) {
				KbdWith(child, "", KbdProps{Content: func(inner *ui.Context) {
					if show {
						ui.Text(inner, "Shift")
					}
				}})
			})
			input := ui.TextInput(c, &draft).Label("After keycaps")
			ui.Local(input, "init", func() bool { created++; return true })
		})
	}, 300, 160)
	if err := r.Click("After keycaps"); err != nil {
		t.Fatal(err)
	}
	show = true
	r.Frame()
	if !r.Focused("After keycaps") || created != 1 {
		t.Fatal("empty keycap callback leaked identity")
	}
}

func TestKbdDirectionUsesLogicalGroupOrder(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Direction(ui.RightToLeft).Padding(20).Children(func() { KbdGroup(c, func(child *ui.Context) { Kbd(child, "Ctrl"); Kbd(child, "K") }) })
	}, 300, 120)
	first, _ := r.Find("Ctrl")
	last, _ := r.Find("K")
	if first.X <= last.X || first.X-last.X-last.W != 4 {
		t.Fatalf("RTL key group order/gap: first=%+v last=%+v", first, last)
	}
}
