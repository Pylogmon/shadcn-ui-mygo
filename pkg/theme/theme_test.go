package theme

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestColorConversion(t *testing.T) {
	for _, tc := range []struct {
		l, c, h float64
		want    ui.Color
	}{
		{0, 0, 0, ui.RGB(0, 0, 0)}, {1, 0, 0, ui.RGB(255, 255, 255)},
		{.145, 0, 0, ui.RGB(10, 10, 10)}, {.922, 0, 0, ui.RGB(229, 229, 229)},
		{.62795536, .2576833, 29.233885, ui.RGB(255, 0, 0)},
	} {
		got := OKLCH(tc.l, tc.c, tc.h)
		if got != tc.want {
			t.Errorf("OKLCH(%v,%v,%v)=%+v want %+v", tc.l, tc.c, tc.h, got, tc.want)
		}
	}
	if a := OKLCH(1, 0, 0, .1).A; a != 26 {
		t.Fatalf("10%% alpha=%v", a)
	}
	if got := Contrast(ui.RGB(0, 0, 0), ui.RGB(255, 255, 255)); math.Abs(got-21) > .00001 {
		t.Fatal(got)
	}
	// Transparent red contributes no red to a premultiplied blend.
	if got := Mix(ui.Color{R: 255}, ui.RGB(0, 0, 255), .5); got.R != 0 || got.B != 255 || got.A != 128 {
		t.Fatal(got)
	}
}

func TestParseColors(t *testing.T) {
	for value, want := range map[string]ui.Color{
		"#fff": ui.RGB(255, 255, 255), "#f008": {R: 255, A: 136}, "#11223344": {R: 17, G: 34, B: 51, A: 68},
		"oklch(1 0 0 / 10%)": {R: 255, G: 255, B: 255, A: 26}, "oklch(100% 0 0)": ui.RGB(255, 255, 255),
	} {
		got, err := ParseColor(value)
		if err != nil || got != want {
			t.Errorf("%s: got %+v %v want %+v", value, got, err, want)
		}
	}
	for _, value := range []string{"", "#12", "#ggg", "oklch(1 0)", "oklch(NaN 0 0)", "oklch(1 0 0 .4)", "oklch(1 0 5%)"} {
		if _, err := ParseColor(value); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
}

func TestIndependentTokensAndContrast(t *testing.T) {
	first := Light()
	first.Primary = ui.RGB(255, 0, 0)
	if Light().Primary == first.Primary {
		t.Fatal("Light returns shared mutable state")
	}
	for _, name := range Presets() {
		for _, dark := range []bool{false, true} {
			tk := Preset(name, dark)
			if Contrast(tk.Foreground, tk.Background) < 7 {
				t.Errorf("%s dark=%v poor text contrast", name, dark)
			}
			if Contrast(tk.Primary, tk.PrimaryForeground) < 4.5 {
				t.Errorf("%s dark=%v poor primary contrast: %v", name, dark, Contrast(tk.Primary, tk.PrimaryForeground))
			}
		}
	}
}

func TestScopeKeepsAllSemanticTokens(t *testing.T) {
	custom := Light()
	custom.Card = ui.RGB(211, 222, 233)
	custom.Chart3 = ui.RGB(3, 4, 5)
	custom.Sidebar = ui.RGB(6, 7, 8)
	var root, inside, after Tokens
	ui.NewTester(func(c *ui.Context) {
		Apply(c)
		root = *TokensOf(c)
		Scope(c, custom).Children(func() { inside = *TokensOf(c); ui.Text(c, "Scoped") })
		after = *TokensOf(c)
	}, 300, 200)
	if inside != *custom {
		t.Fatalf("scope lost semantic tokens: %+v", inside)
	}
	if root != after || root.Card == custom.Card {
		t.Fatal("scope leaked into parent")
	}
}

func TestControllerSwitchesAndFollowsSystem(t *testing.T) {
	s := Controller{}
	var got Tokens
	tt := ui.NewTester(func(c *ui.Context) { s.Apply(c); got = *TokensOf(c); ui.Text(c, "Theme") }, 120, 80)
	if got.Dark {
		t.Fatal("default should follow light system")
	}
	tt.SetDark(true)
	if !got.Dark {
		t.Fatal("failed to follow dark system")
	}
	s.Mode = LightMode
	tt.Frame()
	if got.Dark {
		t.Fatal("explicit light failed")
	}
	s.Mode = DarkMode
	tt.Frame()
	if !got.Dark {
		t.Fatal("explicit dark failed")
	}
	tt.SetDark(false)
	if !got.Dark {
		t.Fatal("explicit appearance changed with system")
	}
	s.Mode = System
	tt.Frame()
	if got.Dark {
		t.Fatal("return to system retained explicit dark")
	}
	s.Preset = "violet"
	tt.Frame()
	if got.Primary != Preset("violet", false).Primary {
		t.Fatal("preset change not applied")
	}
	s.ReduceMotion = true
	var instant bool
	ui.NewTester(func(c *ui.Context) { s.Apply(c); instant = c.Instant() }, 40, 40)
	if !instant {
		t.Fatal("app reduce motion did not disable animations")
	}
}

func TestPreferencesRoundTripAndRejectInvalid(t *testing.T) {
	zero := float32(0)
	want := Controller{Mode: DarkMode, Preset: "blue", Radius: 8, RadiusOverride: &zero, ReduceMotion: true}
	path := filepath.Join(t.TempDir(), "settings", "theme.json")
	if err := want.Save(path); err != nil {
		t.Fatal(err)
	}
	var got Controller
	if err := got.Load(path); err != nil {
		t.Fatal(err)
	}
	if got.Mode != want.Mode || got.Preset != want.Preset || got.Radius != 8 || got.RadiusOverride == nil || *got.RadiusOverride != 0 || !got.ReduceMotion {
		t.Fatalf("bad round trip: %+v", got)
	}
	if got.Resolve(false).Radius != 0 {
		t.Fatal("explicit square radius lost")
	}
	for _, bad := range []string{`{"mode":"sepia"}`, `{"preset":"typo"}`, `{"radius":-1}`, `{"unknown":true}`, `{} {}`} {
		if err := os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		if err := got.Load(path); err == nil {
			t.Errorf("accepted %s", bad)
		}
		if got.Mode != DarkMode || got.Preset != "blue" {
			t.Fatal("invalid JSON changed state")
		}
	}
}
