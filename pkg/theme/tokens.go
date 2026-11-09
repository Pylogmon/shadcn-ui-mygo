// Package theme implements shadcn's semantic color pairs in native Go.
package theme

//go:generate go run ./internal/generate

import (
	"github.com/egoist/mygo/ui"
	"slices"
)

// Tokens is a complete theme value. Copy it to customize a single application
// or subtree. Lengths are DIPs; the default radius is .625rem at 16 DIPs/rem.
// DestructiveForeground, FontSize and Spacing are native extensions to the CSS
// registry. Every other value is generated from the pinned upstream sources.
type Tokens struct {
	Dark                                     bool
	Background, Foreground                   ui.Color
	Card, CardForeground                     ui.Color
	Popover, PopoverForeground               ui.Color
	Primary, PrimaryForeground               ui.Color
	Secondary, SecondaryForeground           ui.Color
	Muted, MutedForeground                   ui.Color
	Accent, AccentForeground                 ui.Color
	Destructive, DestructiveForeground       ui.Color
	Border, Input, Ring                      ui.Color
	Chart1, Chart2, Chart3, Chart4, Chart5   ui.Color
	Sidebar, SidebarForeground               ui.Color
	SidebarPrimary, SidebarPrimaryForeground ui.Color
	SidebarAccent, SidebarAccentForeground   ui.Color
	SidebarBorder, SidebarRing               ui.Color
	Radius, FontSize, Spacing                float32
}

// Light and Dark return independent copies of the pinned registry Neutral
// theme, including its monochrome chart colors. Registry values take precedence
// where the documentation's example scaffold has different chart values.
func Light() *Tokens { return Preset("neutral", false) }
func Dark() *Tokens  { return Preset("neutral", true) }

func (t *Tokens) RadiusSM() float32 { return t.Radius * .6 }
func (t *Tokens) RadiusMD() float32 { return t.Radius * .8 }
func (t *Tokens) RadiusLG() float32 { return t.Radius }
func (t *Tokens) RadiusXL() float32 { return t.Radius * 1.4 }
func (t *Tokens) Charts() [5]ui.Color {
	return [5]ui.Color{t.Chart1, t.Chart2, t.Chart3, t.Chart4, t.Chart5}
}

// Presets returns all 24 official registry themes, with Neutral first.
// Seven base palettes carry full surfaces; brand themes merge the registry's
// partial token overrides onto Neutral. Results may be edited independently.
func Presets() []string {
	names := []string{"neutral"}
	for name := range registryPresets {
		if name != "neutral" {
			names = append(names, name)
		}
	}
	slices.Sort(names[1:])
	return names
}
func Preset(name string, dark bool) *Tokens {
	p, ok := registryPresets[name]
	if !ok {
		p = registryPresets["neutral"]
	}
	t := p.light
	if dark {
		t = p.dark
	}
	return &t
}

// PaletteNames returns the 26 pinned Tailwind scale names alphabetically.
func PaletteNames() []string {
	names := make([]string, 0, len(paletteScales))
	for name := range paletteScales {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Shades returns the ordered stops present in every Tailwind palette.
func Shades() []int { return []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950} }

// Palette looks up a color scale stop. Unknown names and shades return false.
func Palette(name string, shade int) (ui.Color, bool) {
	scale, ok := paletteScales[name]
	if !ok {
		return ui.Color{}, false
	}
	i := -1
	switch {
	case shade == 50:
		i = 0
	case shade == 950:
		i = 10
	case shade >= 100 && shade <= 900 && shade%100 == 0:
		i = shade / 100
	}
	if i < 0 {
		return ui.Color{}, false
	}
	return scale[i], true
}
