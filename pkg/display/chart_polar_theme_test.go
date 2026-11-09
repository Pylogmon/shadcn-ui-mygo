package display

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Read the native legend swatch, rather than deriving the expected color using
// the chart's palette helpers. Wide-gamut tokens follow the same native paint
// conversion for both this swatch and a mark.
func polarThemeLegendColor(t *testing.T, r *ui.Tester, name string) color.RGBA {
	t.Helper()
	legend, ok := r.AccessByRole(ui.RoleGroup, "Chart legend")
	if !ok {
		t.Fatal("missing native legend")
	}
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleText && n.Name == name && n.Bounds.Y >= legend.Bounds.Y {
			return r.Image().RGBAAt(int(n.Bounds.X-10), int(n.Bounds.Y+n.Bounds.H/2))
		}
	}
	t.Fatalf("missing legend item %q", name)
	return color.RGBA{}
}

func polarThemePixelCount(r *ui.Tester, b ui.Rect, want color.RGBA) int {
	count := 0
	image := r.Image()
	for y := int(b.Y); y < int(b.Y+b.H); y++ {
		for x := int(b.X); x < int(b.X+b.W); x++ {
			if image.RGBAAt(x, y) == want {
				count++
			}
		}
	}
	return count
}

func TestPolarScopedBlueMarksUseTheActualLegendPalette(t *testing.T) {
	cases := []struct {
		name             string
		kind             ChartKind
		stacked, missing bool
	}{
		{"pie", PieChart, false, false}, {"stacked-pie", PieChart, true, false},
		{"donut", DonutChart, false, false}, {"stacked-donut", DonutChart, true, false},
		{"radial", RadialChart, false, false}, {"stacked-radial", RadialChart, true, false},
		{"radar", RadarChart, false, false}, {"missing-radar", RadarChart, false, true},
	}
	for _, dark := range []bool{false, true} {
		for _, test := range cases {
			t.Run(fmt.Sprint(test.name, "/dark=", dark), func(t *testing.T) {
				values := []float64{4, 3, 2, 4}
				if test.missing {
					values[1] = math.NaN()
				}
				p := ChartProps{Kind: test.kind, Stacked: test.stacked, Height: 240, Label: "Scoped blue", Labels: []string{"Alpha", "Beta", "Gamma", "Delta"}, Legend: true, HideTooltip: true,
					Series: []ChartSeries{{Name: "Desktop", Values: values}, {Name: "Mobile", Values: []float64{1, 2, 3, 1}}}}
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c) // The window remains Neutral while this subtree is Blue.
					tokens := theme.Preset("blue", dark)
					theme.Scope(c, tokens).Fill().Background(tokens.Background).Padding(20).Children(func() { Chart(c, p) })
				}, 400, 340)
				// Sample final scoped palette rather than the initial sector entrance.
				r.SetPreferences(ui.Preferences{ReduceMotion: true})
				plot, ok := r.AccessByRole(ui.RoleGroup, "Scoped blue")
				if !ok {
					t.Fatal("missing polar plot")
				}
				names := p.Labels
				if test.kind == RadarChart || test.kind == RadialChart && test.stacked {
					names = []string{"Desktop", "Mobile"}
				}
				for _, name := range names {
					swatch := polarThemeLegendColor(t, r, name)
					if swatch.R == swatch.G && swatch.G == swatch.B {
						t.Fatalf("probe legend %q is achromatic: %v", name, swatch)
					}
					if count := polarThemePixelCount(r, plot.Bounds, swatch); count < 2 {
						t.Fatalf("%q scoped legend%v has only %d matching native mark pixels", name, swatch, count)
					}
				}
			})
		}
	}
}

func TestPolarExplicitColorsKeepOverridesAndScopedFallbacks(t *testing.T) {
	tests := []struct {
		name             string
		kind             ChartKind
		stacked, missing bool
	}{
		{"pie", PieChart, false, false}, {"stacked-pie", PieChart, true, false},
		{"donut", DonutChart, false, false}, {"stacked-donut", DonutChart, true, false},
		{"radial", RadialChart, false, false}, {"stacked-radial", RadialChart, true, false},
		{"radar", RadarChart, false, false}, {"missing-radar", RadarChart, false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			explicitCategory := ui.RGB(200, 35, 65)
			explicitSeries := ui.RGB(210, 95, 20)
			values := []float64{4, 3, 2, 4}
			if test.missing {
				values[1] = math.NaN()
			}
			p := ChartProps{Kind: test.kind, Stacked: test.stacked, Height: 240, Label: "Explicit scoped chart", Labels: []string{"Alpha", "Beta", "Gamma", "Delta"}, Legend: true, HideTooltip: true,
				CategoryColors: []ui.Color{explicitCategory}, Series: []ChartSeries{{Name: "Desktop", Values: values, Color: explicitSeries}, {Name: "Mobile", Values: []float64{1, 2, 3, 1}}}}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				tk := theme.Preset("blue", false)
				theme.Scope(c, tk).Fill().Background(tk.Background).Padding(20).Children(func() { Chart(c, p) })
			}, 400, 340)
			// This is a final palette/scope pixel contract, not entry timing.
			r.SetPreferences(ui.Preferences{ReduceMotion: true})
			plot, _ := r.AccessByRole(ui.RoleGroup, p.Label)
			names := p.Labels
			expected := explicitCategory
			if test.kind == RadarChart || test.kind == RadialChart && test.stacked {
				names = []string{"Desktop", "Mobile"}
				expected = explicitSeries
			}
			for i, name := range names {
				swatch := polarThemeLegendColor(t, r, name)
				if i == 0 && swatch != (color.RGBA{R: expected.R, G: expected.G, B: expected.B, A: 255}) {
					t.Fatalf("explicit %q swatch changed: %v", name, swatch)
				}
				if count := polarThemePixelCount(r, plot.Bounds, swatch); count < 2 {
					t.Fatalf("explicit/fallback %q legend%v not painted (%d pixels)", name, swatch, count)
				}
			}
		})
	}
}

func TestPolarSiblingScopesAndLiveThemeChangesPreserveNativeInspection(t *testing.T) {
	brand := "blue"
	leftSelected, rightSelected, windowSelected := -1, -1, -1
	r := ui.NewTester(func(c *ui.Context) {
		neutral := theme.Light()
		c.SetTheme(neutral.Theme())
		ui.Column(c).Fill().Background(neutral.Background).Padding(12).Children(func() {
			ui.Row(c).FillWidth().Gap(12).Children(func() {
				for i, preset := range []string{brand, "green", "neutral"} {
					tk := theme.Preset(preset, false)
					title := []string{"Left", "Right", "Window"}[i]
					selected := []*int{&leftSelected, &rightSelected, &windowSelected}[i]
					p := ChartProps{Kind: DonutChart, Height: 220, Label: title + " scoped plot", Labels: []string{title + " alpha", title + " beta"}, Legend: true, HideTooltip: true, Selected: selected, Series: []ChartSeries{{Name: "Visitors", Values: []float64{1, 1}}}}
					if i == 2 {
						ui.Column(c).Width(270).Shrink(0).Children(func() { Chart(c, p) })
						continue
					}
					theme.Scope(c, tk).Width(270).Shrink(0).Background(tk.Background).Children(func() { Chart(c, p) })
				}
			})
		})
	}, 880, 300)
	// Isolate scoped palette and native inspection from entry geometry.
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	check := func() {
		for _, title := range []string{"Left", "Right", "Window"} {
			b, ok := r.AccessByRole(ui.RoleGroup, title+" scoped plot")
			if !ok {
				t.Fatal("missing plot", title)
			}
			swatch := polarThemeLegendColor(t, r, title+" alpha")
			if count := polarThemePixelCount(r, b.Bounds, swatch); count < 50 {
				t.Fatalf("%s palette leaked: swatch%v matches%d pixels", title, swatch, count)
			}
		}
	}
	check()
	beforeRight := polarThemeLegendColor(t, r, "Right alpha")
	beforeWindow := polarThemeLegendColor(t, r, "Window alpha")
	beforeLeft := polarThemeLegendColor(t, r, "Left alpha")
	if err := r.Click("Left scoped plot"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyHome)
	if leftSelected != 0 || rightSelected != -1 {
		t.Fatalf("native inspection selected left%d right%d", leftSelected, rightSelected)
	}
	brand = "rose"
	r.Frame()
	check()
	if got := polarThemeLegendColor(t, r, "Left alpha"); got == beforeLeft {
		t.Fatal("live scope palette did not update")
	}
	if got := polarThemeLegendColor(t, r, "Right alpha"); got != beforeRight {
		t.Fatal("left update leaked into sibling scope")
	}
	if got := polarThemeLegendColor(t, r, "Window alpha"); got != beforeWindow {
		t.Fatal("scoped update leaked into Neutral window")
	}
	plot, _ := r.AccessByRole(ui.RoleGroup, "Left scoped plot")
	if !plot.Focused || leftSelected != 0 {
		t.Fatalf("theme change lost native focus/selection: %+v selected%d", plot, leftSelected)
	}
	r.Key(0, ui.KeyEnd)
	if leftSelected != 1 || rightSelected != -1 || windowSelected != -1 {
		t.Fatalf("post-theme keyboard selected left%d right%d", leftSelected, rightSelected)
	}
}

func TestPolarRadarLabelsKeepScopedNativeFontAndForeground(t *testing.T) {
	for _, family := range []string{"serif", "monospace"} {
		t.Run(family, func(t *testing.T) {
			tk := theme.Preset("blue", false)
			r := ui.NewTester(func(c *ui.Context) {
				window := theme.Light().Theme()
				window.Font = "system-ui"
				c.SetTheme(window)
				native := tk.Theme()
				native.Font = family
				ui.Themed(c, native).Fill().Background(tk.Background).FontWeight(700).Padding(20).Children(func() {
					Chart(c, ChartProps{Kind: RadarChart, Height: 240, Labels: []string{"January", "February", "March", "April"}, Series: []ChartSeries{{Name: "Desktop", Values: []float64{4, 3, 2, 4}}}, HideTooltip: true})
				})
			}, 400, 300)
			var label ui.AccessInfo
			for _, node := range r.AccessAll() {
				if node.Role == ui.RoleText && node.Name == "January" {
					label = node
					break
				}
			}
			if label.Name == "" {
				t.Fatal("missing native radar text")
			}
			plot, _ := r.AccessByRole(ui.RoleGroup, "Chart")
			// Independently shape only this label at its resolved native coordinates.
			// The scope is gone by drawing time; neither font nor semantic foreground
			// may be recovered from the Neutral window's Context then.
			reference := ui.NewTester(func(c *ui.Context) {
				ui.Box(c).Fill().Background(tk.Background).Children(func() {
					ui.Box(c).Absolute().Left(plot.Bounds.X).Top(plot.Bounds.Y).Size(plot.Bounds.W, plot.Bounds.H).Children(func() {
						ui.Text(c, label.Name).Absolute().LeftPercent(50).TopPercent(6).TranslatePercent(-50, -50).MaxWidth(90).SingleLine().Font(family).FontSize(11).FontWeight(700).TextColor(tk.MutedForeground)
					})
				})
			}, 400, 300)
			// The top label leaves the radar paths below its glyphs. Preserve
			// its fractional TranslatePercent origin in the reference, then
			// compare actual glyph pixels exactly rather than guessing which
			// antialiased pixel has full foreground coverage.
			actualImage, referenceImage := r.Image(), reference.Image()
			background := referenceImage.RGBAAt(0, 0)
			checked := 0
			for y := int(label.Bounds.Y); y < int(label.Bounds.Y+label.Bounds.H); y++ {
				for x := int(label.Bounds.X); x < int(label.Bounds.X+label.Bounds.W); x++ {
					want := referenceImage.RGBAAt(x, y)
					if want == background {
						continue
					}
					checked++
					if got := actualImage.RGBAAt(x, y); got != want {
						t.Fatalf("scoped native %s glyph %d,%d got%v want%v", family, x, y, got, want)
					}
				}
			}
			if checked < 20 {
				t.Fatalf("insufficient native %s glyph coverage: %d", family, checked)
			}
		})
	}
}
