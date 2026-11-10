package control

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func colorPickerTester(view func(*ui.Context), dark bool, width int) *ui.Tester {
	appearance := theme.Controller{Mode: theme.LightMode}
	if dark {
		appearance.Mode = theme.DarkMode
	}
	return ui.NewTesterWith(func(c *ui.Context) {
		appearance.Apply(c)
		t := theme.TokensOf(c)
		ui.Column(c).Fill().Padding(24).Gap(28).Background(t.Background).Children(func() { view(c) })
	}, width, 600, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}

func TestColorPickerPresetSelectionReportsOnceAndRemainsSelected(t *testing.T) {
	value := DefaultColorOptions()[0].Color
	var callbacks, changes []ui.Color
	r := colorPickerTester(func(c *ui.Context) {
		if ColorPicker(c, &value, ColorPickerProps{Label: "Accent", OnValueChange: func(color ui.Color) {
			callbacks = append(callbacks, color)
		}}).Changed() {
			changes = append(changes, value)
		}
	}, true, 720)
	choiceActivate(t, r, "Green")
	want := DefaultColorOptions()[5].Color
	if value != want || !slices.Equal(callbacks, []ui.Color{want}) || !slices.Equal(changes, callbacks) {
		t.Fatalf("preset edit: value=%v callbacks=%v changes=%v", value, callbacks, changes)
	}
	if !choiceAccess(t, r, ui.RoleRadio, "Green").Checked || choiceAccess(t, r, ui.RoleRadio, "Blue").Checked {
		t.Fatal("same-frame native checked states did not follow the model")
	}
	choiceActivate(t, r, "Green")
	r.Frame()
	if value != want || len(callbacks) != 1 || len(changes) != 1 {
		t.Fatal("reselecting or repainting published another edit")
	}
	value = DefaultColorOptions()[1].Color
	r.Frame()
	if !choiceAccess(t, r, ui.RoleRadio, "Purple").Checked || len(callbacks) != 1 {
		t.Fatal("external controlled value did not update silently")
	}
}

func TestColorPickerNativeKeyboardDisabledSkipAndStableKeys(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprint(rtl), func(t *testing.T) {
			colors := DefaultColorOptions()[:3]
			colors[1].Disabled = true
			value := colors[0].Color
			r := colorPickerTester(func(c *ui.Context) {
				root := ui.Column(c)
				if rtl {
					root.Direction(ui.RightToLeft)
				}
				root.Children(func() {
					ColorPicker(c, &value, ColorPickerProps{Colors: colors})
					ui.Button(c, "After colors")
				})
			}, false, 720)
			r.Key(0, ui.KeyTab)
			if !r.Focused("Blue") {
				t.Fatal("Tab did not enter at the selected preset")
			}
			next := ui.KeyRight
			if rtl {
				next = ui.KeyLeft
			}
			r.Key(0, next)
			if value != colors[2].Color || !r.Focused("Magenta") {
				t.Fatal("native arrow navigation did not skip disabled presets")
			}
			slices.Reverse(colors)
			r.Frame()
			if !r.Focused("Magenta") {
				t.Fatal("reordering colors lost keyed native focus")
			}
			r.Key(0, ui.KeyEnd)
			if value != colors[2].Color || !r.Focused("Blue") {
				t.Fatal("End used the previous palette order")
			}
			r.Key(0, ui.KeyTab)
			if !r.Focused("After colors") {
				t.Fatal("presets added multiple Tab stops")
			}
		})
	}
}

func TestColorPickerCustomPanelHexAlphaDismissAndReopen(t *testing.T) {
	value := DefaultColorOptions()[0].Color
	var callbacks, changes []ui.Color
	r := colorPickerTester(func(c *ui.Context) {
		if ColorPicker(c, &value, ColorPickerProps{Custom: true, OnValueChange: func(color ui.Color) {
			callbacks = append(callbacks, color)
		}}).Changed() {
			changes = append(changes, value)
		}
		ui.Button(c, "Outside colors")
	}, true, 760)
	well := choiceAccess(t, r, ui.RoleColorWell, "Custom color")
	blue := choiceAccess(t, r, ui.RoleRadio, "Blue")
	if well.Bounds.W != blue.Bounds.W || well.Bounds.H != blue.Bounds.H || well.Bounds.X >= blue.Bounds.X {
		t.Fatalf("custom well differs from preset size or order: %+v / %+v", well.Bounds, blue.Bounds)
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Hex"); ok {
		t.Fatal("custom panel opened without activation")
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyEnter)
	choiceAccess(t, r, ui.RoleSlider, "Hue")
	choiceAccess(t, r, ui.RoleSlider, "Opacity")
	choiceActivate(t, r, "Hex")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("#12345680")
	want := ui.Hex("#12345680")
	if value != want || len(callbacks) != 1 || !slices.Equal(changes, callbacks) || callbacks[0] != want {
		t.Fatalf("native RGBA edit was not published once: %v %v %v", value, callbacks, changes)
	}
	if choiceAccess(t, r, ui.RoleColorWell, "Custom color").Value != "#12345680" || choiceAccess(t, r, ui.RoleRadio, "Blue").Checked {
		t.Fatal("custom edit did not update the well and clear preset selection")
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleTextField, "Hex"); ok {
		t.Fatal("Escape did not dismiss the panel")
	}
	if value != want || len(callbacks) != 1 {
		t.Fatal("dismissal changed the accepted live color")
	}
	choiceActivate(t, r, "Custom color")
	if choiceAccess(t, r, ui.RoleTextField, "Hex").Value != "#12345680" {
		t.Fatal("reopening the panel lost its controlled color")
	}
	choiceActivate(t, r, "Outside colors")
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleTextField, "Hex"); ok {
		t.Fatal("outside press did not dismiss the panel")
	}
	choiceActivate(t, r, "Teal")
	if value != DefaultColorOptions()[6].Color || len(callbacks) != 2 {
		t.Fatal("preset selection failed after closing the custom panel")
	}
}

func TestColorPickerDisabledAncestorAndOptionalWell(t *testing.T) {
	value := DefaultColorOptions()[0].Color
	disabled, custom, calls := true, true, 0
	r := colorPickerTester(func(c *ui.Context) {
		ui.Column(c).Disabled(disabled).Children(func() {
			ColorPicker(c, &value, ColorPickerProps{Custom: custom, OnValueChange: func(ui.Color) { calls++ }})
		})
	}, false, 760)
	choiceActivate(t, r, "Custom color")
	choiceActivate(t, r, "Green")
	if calls != 0 || value != DefaultColorOptions()[0].Color {
		t.Fatal("disabled ancestor allowed a color edit")
	}
	if !choiceAccess(t, r, ui.RoleRadio, "Green").Disabled || !choiceAccess(t, r, ui.RoleColorWell, "Custom color").Disabled {
		t.Fatal("native controls lost inherited disabled state")
	}
	disabled = false
	r.Frame()
	choiceActivate(t, r, "Custom color")
	choiceAccess(t, r, ui.RoleTextField, "Hex")
	disabled = true
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleTextField, "Hex"); ok {
		t.Fatal("disabling an open picker left the panel editable")
	}
	disabled, custom = false, false
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleColorWell, "Custom color"); ok {
		t.Fatal("optional custom well remained mounted")
	}
	choiceActivate(t, r, "Green")
	if calls != 1 {
		t.Fatal("re-enabling presets did not restore input")
	}
}

func TestColorPickerShapesAndSelectionPaintInBothThemes(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, shape := range []ColorSwatchShape{ColorSwatchRounded, ColorSwatchSquare, ColorSwatchCircle} {
			t.Run(fmt.Sprintf("dark=%v/shape=%v", dark, shape), func(t *testing.T) {
				value := DefaultColorOptions()[0].Color
				r := colorPickerTester(func(c *ui.Context) {
					ColorPicker(c, &value, ColorPickerProps{Custom: true, Shape: shape})
				}, dark, 760)
				a, b := choiceAccess(t, r, ui.RoleRadio, "Blue"), choiceAccess(t, r, ui.RoleRadio, "Purple")
				if a.Bounds.W != 48 || a.Bounds.H != 48 || b.Bounds.X-a.Bounds.X != 72 {
					t.Fatalf("reference size or spacing changed: %+v / %+v", a.Bounds, b.Bounds)
				}
				im := r.Image()
				pixel := func(x, y float32) ui.Color {
					v := im.RGBAAt(int(x), int(y))
					return ui.RGB(v.R, v.G, v.B)
				}
				if pixel(a.Bounds.X-5, a.Bounds.Y+24) != value {
					t.Fatal("selected swatch did not paint its separated colored ring")
				}
				checkPixels := 0
				for y := float32(12); y < 36; y++ {
					for x := float32(12); x < 36; x++ {
						if pixel(a.Bounds.X+x, a.Bounds.Y+y) == ui.RGB(255, 255, 255) {
							checkPixels++
						}
					}
				}
				if checkPixels < 12 {
					t.Fatal("selected swatch did not paint the contrast check")
				}
				corner := pixel(b.Bounds.X+1, b.Bounds.Y+1)
				if (corner == DefaultColorOptions()[1].Color) != (shape == ColorSwatchSquare) {
					t.Fatal("square and rounded swatch corners did not follow Shape")
				}
				inset := pixel(b.Bounds.X+4, b.Bounds.Y+10)
				if (inset == DefaultColorOptions()[1].Color) == (shape == ColorSwatchCircle) {
					t.Fatal("circle and rounded-square geometry did not differ")
				}
			})
		}
	}
}

func TestColorPickerCustomPaletteEmptyAndDimensions(t *testing.T) {
	value := ui.Hex("#ff003380")
	colors := []ColorOption{{Color: value}, {Color: ui.Hex("#336699"), Label: "Ocean"}}
	size, gap := float32(32), float32(0)
	r := colorPickerTester(func(c *ui.Context) {
		ColorPicker(c, &value, ColorPickerProps{Colors: colors, SwatchSize: size, Spacing: &gap})
	}, false, 320)
	a, b := choiceAccess(t, r, ui.RoleRadio, "#ff003380"), choiceAccess(t, r, ui.RoleRadio, "Ocean")
	if a.Bounds.W != size || b.Bounds.X != a.Bounds.X+size {
		t.Fatal("explicit size or zero spacing was ignored")
	}
	colors = []ColorOption{}
	r.Frame()
	for _, info := range r.AccessAll() {
		if info.Role == ui.RoleRadio {
			t.Fatal("empty palette unexpectedly rendered default colors")
		}
	}
	colors, size, gap = nil, float32(math.NaN()), -1
	r.Frame()
	a, b = choiceAccess(t, r, ui.RoleRadio, "Blue"), choiceAccess(t, r, ui.RoleRadio, "Purple")
	if a.Bounds.W != 48 || b.Bounds.X-a.Bounds.X != 72 {
		t.Fatal("invalid dimensions did not fall back to the defaults")
	}
}

func TestColorPickerCheckRemainsReadableOnWhiteAndTransparentPresets(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, color := range []ui.Color{ui.RGB(255, 255, 255), ui.Transparent} {
			value := color
			r := colorPickerTester(func(c *ui.Context) {
				ColorPicker(c, &value, ColorPickerProps{Colors: []ColorOption{{Color: color, Label: "Pale color"}}})
			}, dark, 320)
			a := choiceAccess(t, r, ui.RoleRadio, "Pale color")
			im, darkPixels := r.Image(), 0
			for y := 12; y < 36; y++ {
				for x := 12; x < 36; x++ {
					p := im.RGBAAt(int(a.Bounds.X)+x, int(a.Bounds.Y)+y)
					if p.R < 50 && p.G < 50 && p.B < 50 {
						darkPixels++
					}
				}
			}
			if darkPixels < 12 {
				t.Fatalf("check lacks contrast: dark=%v color=%v", dark, color)
			}
		}
	}
}
