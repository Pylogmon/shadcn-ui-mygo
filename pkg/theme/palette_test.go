package theme

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCompleteTailwindScalesAndKnownCoordinates(t *testing.T) {
	names := PaletteNames()
	if len(names) != 26 || !slices.IsSorted(names) {
		t.Fatalf("palette names: %v", names)
	}
	for _, name := range names {
		for _, shade := range Shades() {
			c, ok := Palette(name, shade)
			if !ok || c.A != 255 {
				t.Errorf("missing opaque %s-%d: %+v", name, shade, c)
			}
		}
	}
	// These samples include the newer base families and CSS's absent hue
	// syntax. Coordinates are read independently from the pinned CSS source.
	for _, tc := range []struct {
		name  string
		shade int
		want  ui.Color
	}{
		{"blue", 500, OKLCH(.623, .214, 259.815)},
		{"neutral", 950, ui.RGB(10, 10, 10)},
		{"mauve", 500, OKLCH(.542, .034, 322.5)},
		{"olive", 500, OKLCH(.58, .031, 107.3)},
		{"mist", 500, OKLCH(.56, .021, 213.5)},
		{"taupe", 500, OKLCH(.547, .021, 43.1)},
	} {
		got, ok := Palette(tc.name, tc.shade)
		if !ok || got != tc.want {
			t.Errorf("%s-%d=%+v want %+v", tc.name, tc.shade, got, tc.want)
		}
	}
	for _, shade := range []int{-100, 0, 51, 150, 1000} {
		if _, ok := Palette("blue", shade); ok {
			t.Errorf("invalid shade %d accepted", shade)
		}
	}
	if _, ok := Palette("typo", 500); ok {
		t.Fatal("unknown palette accepted")
	}
	names[0] = "changed"
	shades := Shades()
	shades[0] = 0
	if PaletteNames()[0] == "changed" || Shades()[0] != 50 {
		t.Fatal("returned metadata aliases library state")
	}
	if parsed, err := ParseColor("oklch(14.5% 0 none)"); err != nil || parsed != ui.RGB(10, 10, 10) {
		t.Fatalf("missing hue: %+v %v", parsed, err)
	}
}

func TestRegistryPresetFidelityAndInheritance(t *testing.T) {
	if names := Presets(); len(names) != 24 || names[0] != "neutral" {
		t.Fatalf("registry preset inventory: %v", names)
	}
	for _, name := range []string{"neutral", "stone", "zinc", "mauve", "olive", "mist", "taupe"} {
		if !slices.Contains(Presets(), name) {
			t.Errorf("missing official base %s", name)
		}
	}
	light, dark := Light(), Dark()
	if light.Chart1 != OKLCH(.87, 0, 0) || dark.Chart5 != OKLCH(.269, 0, 0) {
		t.Fatal("Neutral did not use pinned registry chart colors")
	}
	blue := Preset("blue", false)
	if blue.Primary != OKLCH(.488, .243, 264.376) || blue.PrimaryForeground != OKLCH(.97, .014, 254.604) {
		t.Fatal("blue primary differs from registry")
	}
	if blue.SidebarPrimary != OKLCH(.546, .245, 262.881) || blue.Chart2 != OKLCH(.623, .214, 259.815) {
		t.Fatal("blue chart/sidebar overrides missing")
	}
	if blue.Background != light.Background || blue.Border != light.Border || blue.Ring != light.Ring {
		t.Fatal("partial brand theme failed to inherit neutral tokens")
	}
	if got := Preset("blue", true).Primary; got != OKLCH(.424, .199, 265.638) {
		t.Fatal("dark primary was approximated rather than sourced")
	}
	if got := Preset("mauve", false); got.Foreground != OKLCH(.145, .008, 326) || got.Ring != OKLCH(.711, .019, 323.02) {
		t.Fatal("complete base theme lost its surfaces")
	}
	if *Preset("unknown", false) != *light {
		t.Fatal("unknown preset fallback changed")
	}
	blue.Primary = ui.RGB(1, 2, 3)
	if Preset("blue", false).Primary == blue.Primary {
		t.Fatal("presets share mutable returned tokens")
	}
}
