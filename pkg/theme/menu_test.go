package theme

import (
	"image/color"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestMenuSnapshotsScopeInversionAccentAndNativeIsolation(t *testing.T) {
	for _, colorMode := range []MenuColor{MenuDefault, MenuInverted, MenuDefaultTranslucent, MenuInvertedTranslucent} {
		light, dark := Preset("blue", false), Preset("blue", true)
		dark.Popover = ui.RGB(31, 36, 48)
		var root Tokens
		var menu, after, native MenuSnapshot
		ui.NewTester(func(c *ui.Context) {
			Apply(c)
			MenuScope(c, light, MenuOptions{Color: colorMode, Accent: MenuBold, InvertedTokens: dark}).Children(func() {
				root = *TokensOf(c)
				menu = MenuOf(c)
				NativeScope(c).Children(func() { native = MenuOf(c) })
			})
			after = MenuOf(c)
		}, 100, 80)
		inverted := colorMode == MenuInverted || colorMode == MenuInvertedTranslucent
		trans := colorMode == MenuDefaultTranslucent || colorMode == MenuInvertedTranslucent
		if root.Dark || menu.Tokens.Dark != inverted || menu.Translucent() != trans || menu.Animated() == trans {
			t.Fatalf("%s root/menu=%+v/%+v", colorMode, root, menu)
		}
		if inverted && menu.Tokens.Popover != dark.Popover {
			t.Fatal("force-dark lost supplied palette")
		}
		if trans {
			if menu.Accent != MenuSubtle || root.Accent != light.Accent {
				t.Fatal("translucent did not normalize bold to subtle")
			}
		} else if root.Accent != root.Primary || root.AccentForeground != root.PrimaryForeground {
			t.Fatal("bold did not map theme-wide accent pair")
		}
		if root.SidebarAccent != light.SidebarAccent {
			t.Fatal("bold changed sidebar accent despite source excluding it")
		}
		if after.Color != MenuDefault || native.Color != MenuDefault || native.Translucent() {
			t.Fatal("profile escaped its scope or entered NativeScope")
		}
		dark.Popover = ui.RGB(255, 0, 255)
		if inverted && menu.Tokens.Popover == dark.Popover {
			t.Fatal("snapshot retained caller token pointer")
		}
	}
}

func TestMenuScopeChoosesItsOwnPresetPairAndKeepsCustomPaletteExplicit(t *testing.T) {
	controller := Controller{Preset: "neutral", Mode: LightMode, MenuColor: MenuInverted, MenuAccent: MenuBold}
	rose := Preset("rose", false)
	custom := Light()
	custom.Popover = ui.RGB(90, 110, 130)
	customDark := Dark()
	customDark.Popover = ui.RGB(20, 30, 40)
	var standard, unpaired, paired MenuSnapshot
	ui.NewTester(func(c *ui.Context) {
		controller.Apply(c)
		Scope(c, rose).Children(func() { standard = MenuOf(c) })
		Scope(c, custom).Children(func() { unpaired = MenuOf(c) })
		MenuScope(c, custom, MenuOptions{Color: MenuInverted, Accent: MenuBold, InvertedTokens: customDark}).Children(func() { paired = MenuOf(c) })
	}, 100, 80)
	darkRose := Preset("rose", true)
	if standard.Tokens.Primary != darkRose.Primary || standard.Tokens.Accent != darkRose.Primary || !standard.HasInvertedPalette() {
		t.Fatal("standard child scope borrowed the ancestor Neutral pair")
	}
	if unpaired.Tokens.Popover != custom.Popover || !unpaired.Tokens.Dark || unpaired.HasInvertedPalette() {
		t.Fatal("custom child silently borrowed a different dark palette")
	}
	if paired.Tokens.Popover != customDark.Popover || !paired.HasInvertedPalette() {
		t.Fatal("explicit custom pair was lost")
	}
}

func TestMenuBackdropUsesRealGaussianSaturationAndTint(t *testing.T) {
	tokens := Light()
	tokens.Popover = ui.RGB(255, 255, 255)
	background := ui.RGB(102, 153, 128)
	r := ui.NewTester(func(c *ui.Context) {
		MenuScope(c, tokens, MenuOptions{Color: MenuDefaultTranslucent}).Fill().Background(background).Children(func() {
			surface := ui.Box(c).Absolute().Left(80).Top(80).Size(160, 160)
			MenuOf(c).Surface(c, surface)
		})
	}, 320, 320)
	got := r.Image().RGBAAt(160, 160)
	rgb := []float64{102.0 / 255, 153.0 / 255, 128.0 / 255}
	l := .213*rgb[0] + .715*rgb[1] + .072*rgb[2]
	actual := []uint8{got.R, got.G, got.B}
	for i, v := range rgb {
		want := 255 * (.7 + .3*math.Max(0, math.Min(1, l+(v-l)*1.5)))
		if math.Abs(float64(actual[i])-want) > 2 {
			t.Fatalf("source 70%%/1.5 saturation pixel=%v channel%d expected%.2f", got, i, want)
		}
	}
	// A tint-only implementation leaves a step at x=160. A real sigma40
	// backdrop has a smooth gradient on both sides of that boundary.
	blur := ui.NewTester(func(c *ui.Context) {
		MenuScope(c, tokens, MenuOptions{Color: MenuDefaultTranslucent}).Fill().Background(ui.RGB(0, 0, 0)).Children(func() {
			ui.Box(c).Absolute().Left(160).Top(0).Size(160, 320).Background(ui.RGB(255, 255, 255))
			panel := ui.Box(c).Absolute().Left(32).Top(32).Size(256, 256)
			MenuOf(c).Surface(c, panel)
		})
	}, 320, 320)
	a, b, c, d := blur.Image().RGBAAt(80, 160).R, blur.Image().RGBAAt(128, 160).R, blur.Image().RGBAAt(192, 160).R, blur.Image().RGBAAt(240, 160).R
	for i, x := range []int{80, 128, 192, 240} {
		pixel := []uint8{a, b, c, d}[i]
		cdf := .5 * (1 + math.Erf((float64(x)+.5-160)/(40*math.Sqrt2)))
		want := 255 * (.7 + .3*cdf)
		if math.Abs(float64(pixel)-want) > 3 {
			t.Fatalf("sigma40 source profile at x%d =%d expected%.2f", x, pixel, want)
		}
	}
	if !(int(b) > int(a)+5 && int(c) < int(d)-5 && a < b && b < c && c < d) {
		t.Fatalf("not a Gaussian backdrop gradient: %d %d %d %d", a, b, c, d)
	}
}

func TestMenuBackdropReportsActualSceneSystemOrOpaqueCapability(t *testing.T) {
	tokens := Light()
	tokens.Popover = ui.RGB(220, 230, 240)
	nativeWindow := false
	var mode MenuBackdrop
	r := ui.NewTester(func(c *ui.Context) {
		MenuScope(c, tokens, MenuOptions{Color: MenuDefaultTranslucent}).Fill().Background(ui.RGB(10, 20, 30)).Children(func() {
			profile := MenuOf(c)
			if nativeWindow {
				profile = profile.ForNativeWindow()
			}
			mode = profile.Backdrop(c)
			profile.Surface(c, ui.Box(c).Fill())
		})
	}, 64, 64)
	if mode != MenuSceneBackdrop {
		t.Fatal("in-window scene backdrop unavailable")
	}
	r.SetPreferences(ui.Preferences{ReduceTransparency: true})
	if got := r.Image().RGBAAt(32, 32); mode != MenuOpaque || got != (color.RGBA{220, 230, 240, 255}) {
		t.Fatalf("reduce transparency fallback=%v/%v", mode, got)
	}
	r.SetPreferences(ui.Preferences{})
	nativeWindow = true
	r.Frame()
	if mode != MenuOpaque {
		t.Fatal("claimed desktop sampling in a native window without actual material")
	}
	// This public Tester flag simulates the host capability; it does not
	// claim a measured macOS/Windows material kernel or desktop image.
	r.SetVibrancy(true)
	if mode != MenuSystemBackdrop {
		t.Fatal("did not use the host's reported window material")
	}
	r.SetPreferences(ui.Preferences{HighContrast: true})
	if mode != MenuOpaque {
		t.Fatal("high contrast did not select the opaque surface")
	}
}
