package display

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// The reference is the native equivalent of the fixed base chart.tsx default
// dot composition and Vega tooltip classes, including its two shadow layers.
func chartTooltipSourceReference(c *ui.Context, t *theme.Tokens) ui.Element {
	return ui.Column(c).MinWidth(128).Gap(6).Padding(6, 10).Radius(t.RadiusLG()).FontSize(12).LineHeight(4.0/3).
		Background(t.Background).Border(1, ui.RGBA(t.Border.R, t.Border.G, t.Border.B, float32(t.Border.A)/255*.5)).
		Shadow(0, 20, 25, -5, ui.RGBA(0, 0, 0, .1)).Shadow(0, 8, 10, -6, ui.RGBA(0, 0, 0, .1)).Children(func() {
		ui.Text(c, "January").FontWeight(500)
		ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(10, 10).Shrink(0).Radius(2).Background(ui.RGB(180, 40, 80)).AccessHidden(true)
			ui.Row(c).Grow(1).MinWidth(0).Justify(ui.SpaceBetween).AlignItems(ui.Center).LineHeight(1).Children(func() {
				ui.Text(c, "A").TextColor(t.MutedForeground)
				ui.Text(c, "1111").Font("monospace").FontWeight(500).FontFeatures("tnum").TextColor(t.Foreground)
			})
		})
	})
}

func TestChartTooltipVegaActualSurfaceGlyphsAndInheritedFont(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
			for _, family := range []string{"serif", "monospace"} {
				t.Run(fmt.Sprint(dark, direction, family), func(t *testing.T) {
					render := func(reference bool) *ui.Tester {
						return ui.NewTester(func(c *ui.Context) {
							tokens := theme.Light()
							if dark {
								tokens = theme.Dark()
							}
							tokens.Radius = 13
							tokens.Border = ui.RGBA(200, 50, 110, .24)
							tokens.Popover = ui.RGB(140, 20, 170)
							tokens.PopoverForeground = ui.RGB(250, 240, 10)
							native := tokens.Theme()
							native.Font = family
							ui.Themed(c, native).Fill().Background(tokens.Background).Padding(40).TextColor(ui.RGB(20, 100, 80)).FontWeight(700).Children(func() {
								if reference {
									chartTooltipSourceReference(c, tokens)
									return
								}
								ChartTooltip(c, ChartTooltipData{Label: "January", Items: []ChartTooltipItem{{Label: "A", Value: 1111, Color: ui.RGB(180, 40, 80)}}})
							})
						}, 360, 180)
					}
					actual, reference := render(false), render(true)
					actual.SetDirection(direction)
					reference.SetDirection(direction)
					if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
						t.Fatal("tooltip differs from source native surface/glyph reference")
					}
					for _, node := range actual.AccessAll() {
						if node.Focusable {
							t.Fatal("default passive tooltip acquired input")
						}
					}
				})
			}
		}
	}
}

func TestChartLegendVegaSwatchTopInsetAndInheritedForegroundPixels(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			render := func(reference bool) *ui.Tester {
				return ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					ui.Column(c).Fill().Padding(20).Font("serif").FontWeight(700).TextColor(ui.RGB(170, 30, 80)).Children(func() {
						if !reference {
							ChartLegend(c, []ChartLegendItem{{Label: "Desktop", Color: ui.RGB(30, 150, 90)}})
							return
						}
						ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.Center).Gap(16).Padding(12, 0, 0, 0).FontSize(12).LineHeight(4.0 / 3).Children(func() {
							ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
								ui.Box(c).Size(8, 8).Shrink(0).Radius(2).Background(ui.RGB(30, 150, 90)).AccessHidden(true)
								ui.Text(c, "Desktop")
							})
						})
					})
				}, 360, 100)
			}
			actual, reference := render(false), render(true)
			actual.SetDirection(direction)
			reference.SetDirection(direction)
			if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
				t.Fatal("legend lost source8DIP swatch/12DIP inset or foreground/font inheritance")
			}
		})
	}
}

func TestChartSurfacesEmptyPayloadAndAbsentTooltipLabelReserveNoExtraLine(t *testing.T) {
	var emptyTip, emptyLegend, tip ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().AlignItems(ui.Start).Children(func() {
			ChartTooltip(c, ChartTooltipData{}).Draw(func(_ *ui.Painter, r ui.Rect) { emptyTip = r })
			ChartLegend(c, nil).Draw(func(_ *ui.Painter, r ui.Rect) { emptyLegend = r })
			ChartTooltip(c, ChartTooltipData{Items: []ChartTooltipItem{{Label: "A", Value: 1}}}).Draw(func(_ *ui.Painter, r ui.Rect) { tip = r })
		})
	}, 360, 140)
	label, _ := r.Find("A")
	value, _ := r.Find("1")
	// Native Text keeps at least its font metrics even with LineHeight(1).
	// Do not mask that native floor with a fixed glyph box. The absent source
	// heading must still contribute no line or gap: only one native row and
	// the two six-DIP paddings plus border remain.
	wantHeight := max(float32(10), max(label.H, value.H)) + 14
	if emptyTip.H != 0 || emptyLegend.H != 0 || tip.H != wantHeight || tip.W != 128 {
		t.Fatalf("empty heights %.2f/%.2f; populated width/height %.2f/%.2f", emptyTip.H, emptyLegend.H, tip.W, tip.H)
	}
	if _, ok := r.AccessByRole(ui.RoleGroup, "Chart legend"); ok {
		t.Fatal("empty legend exposes a meaningless AX group")
	}
}

func TestChartLegendCompositionKeepsOneTwelveDIPPlotGap(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Chart(c, ChartProps{Labels: []string{"Jan", "Feb"}, Series: []ChartSeries{{Name: "Desktop", Values: []float64{1, 2}}}, Legend: true, Height: 100}).Role(ui.RoleGroup).Label("Whole chart")
	}, 400, 220)
	whole, ok := r.AccessByRole(ui.RoleGroup, "Whole chart")
	if !ok {
		t.Fatal("missing chart")
	}
	legend, ok := r.AccessByRole(ui.RoleGroup, "Chart legend")
	if !ok {
		t.Fatal("missing legend")
	}
	if whole.Bounds.H != 128 || legend.Bounds.Y != 100 || legend.Bounds.H != 28 {
		t.Fatalf("source legend pt12 doubled the plot gap: whole=%v legend=%v", whole.Bounds, legend.Bounds)
	}
}

func TestChartTooltipSourceDoesNotCapWideFormattedNativeValue(t *testing.T) {
	value := "888888888888888888888888888888888888888888888"
	var bounds ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().AlignItems(ui.Start).Children(func() {
			ChartTooltip(c, ChartTooltipData{Items: []ChartTooltipItem{{Label: "Desktop", Value: 1}}, FormatValue: func(float64) string { return value }}).Draw(func(_ *ui.Painter, r ui.Rect) { bounds = r })
		})
	}, 650, 160)
	text, ok := r.Find(value)
	if !ok || bounds.W <= 280 || text.H > 18 || text.X < bounds.X || text.X+text.W > bounds.X+bounds.W {
		t.Fatalf("source unconstrained tooltip value was clipped/wrapped: box=%v text=%v", bounds, text)
	}
}

func TestChartLegendMissingOptionalLabelDoesNotReserveNativeTextLine(t *testing.T) {
	var bounds ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ChartLegend(c, []ChartLegendItem{{Color: ui.RGB(30, 100, 180)}}).Draw(func(_ *ui.Painter, r ui.Rect) { bounds = r })
	}, 300, 100)
	if bounds.H != 20 || r.HasText("Series 1") {
		t.Fatalf("missing source config label added a blank line: height=%v", bounds.H)
	}
}

func TestChartLegendFiveBarCategoriesStayReadableAtNarrowWidths(t *testing.T) {
	for _, width := range []int{300, 400} {
		for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
			t.Run(fmt.Sprint(width, direction), func(t *testing.T) {
				names := []string{"Chrome", "Safari", "Firefox", "Edge", "Other"}
				r := ui.NewTester(func(c *ui.Context) {
					tokens := theme.Light()
					native := tokens.Theme()
					native.Font = "monospace"
					ui.Themed(c, native).Fill().Background(tokens.Background).Padding(20).TextColor(tokens.Foreground).Children(func() {
						Chart(c, ChartProps{Kind: BarChart, Orientation: ChartHorizontal, Labels: names,
							Series:         []ChartSeries{{Name: "Visitors", Values: []float64{275, 200, 187, 173, 90}}},
							CategoryColors: []ui.Color{ui.RGB(200, 30, 70), ui.RGB(30, 130, 70), ui.RGB(20, 100, 200), ui.RGB(180, 110, 20), ui.RGB(120, 30, 180)},
							Legend:         true, Height: 180, Cartesian: ChartCartesianProps{HideValueAxis: true, HideGrid: true, HideCursor: true},
						})
					})
				}, width+40, 350)
				r.SetDirection(direction)
				legend, ok := r.AccessByRole(ui.RoleGroup, "Chart legend")
				if !ok {
					t.Fatal("missing category legend")
				}
				for _, name := range names {
					found := false
					for _, node := range r.AccessAll() {
						if node.Role != ui.RoleText || node.Name != name || node.Bounds.Y < legend.Bounds.Y {
							continue
						}
						found = true
						if node.Offscreen || node.Bounds.X < legend.Bounds.X || node.Bounds.X+node.Bounds.W > legend.Bounds.X+legend.Bounds.W+.1 || node.Bounds.H > 16.1 {
							t.Fatalf("category %q clipped or wrapped within its name: %v legend=%v", name, node.Bounds, legend.Bounds)
						}
					}
					if !found {
						t.Fatalf("missing native category label %q", name)
					}
				}
				if width == 300 && legend.Bounds.H <= 28 {
					t.Fatal("narrow native legend did not wrap complete items")
				}
				if width == 400 && legend.Bounds.H != 28 {
					t.Fatal("wide legend no longer preserves source single-line spacing")
				}
			})
		}
	}
}

func TestChartCartesianGridSourceHalfAlphaPreservesTokenTransparency(t *testing.T) {
	for _, orientation := range []ChartOrientation{ChartVertical, ChartHorizontal} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%t", orientation, dark), func(t *testing.T) {
				horizontal := orientation == ChartHorizontal
				ticks := chart.NewLinear(0, 1).Nice(4).Ticks(4)
				at := ticks[len(ticks)/2].At
				if at <= 0 || at >= 1 {
					t.Fatal("probe requires a visible interior tick")
				}
				render := func(reference bool) *ui.Tester {
					return ui.NewTester(func(c *ui.Context) {
						tokens := theme.Light()
						if dark {
							tokens = theme.Dark()
						}
						tokens.Border = ui.RGBA(200, 40, 120, .24)
						theme.Scope(c, tokens).Fill().Background(tokens.Background).Padding(20).Children(func() {
							if reference {
								// Construct border/50 independently: locked MyGo Alpha already multiplies A.
								color := ui.RGBA(tokens.Border.R, tokens.Border.G, tokens.Border.B, float32(tokens.Border.A)/255*.5)
								ui.Box(c).FillWidth().Height(160).Draw(func(p *ui.Painter, r ui.Rect) {
									if horizontal {
										p.Line(r.X+at*r.W, r.Y, r.X+at*r.W, r.Y+r.H, 1, color)
									} else {
										p.Line(r.X, r.Y+(1-at)*r.H, r.X+r.W, r.Y+(1-at)*r.H, 1, color)
									}
								})
								return
							}
							Chart(c, ChartProps{Kind: BarChart, Orientation: orientation, Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Baseline", Values: []float64{0, 0}}}, Height: 160, Label: "Grid opacity", Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideCursor: true}})
						})
					}, 400, 220)
				}
				actual, reference := render(false), render(true)
				area, ok := actual.AccessByRole(ui.RoleGroup, "Grid opacity")
				if !ok {
					t.Fatal("missing actual plot")
				}
				x0, y0 := int(area.Bounds.X+10), int(area.Bounds.Y+(1-at)*area.Bounds.H)-1
				x1, y1 := x0+30, y0+3
				if horizontal {
					x0, y0 = int(area.Bounds.X+at*area.Bounds.W)-1, int(area.Bounds.Y+10)
					x1, y1 = x0+3, y0+30
				}
				painted := false
				background := actual.Image().RGBAAt(int(area.Bounds.X+area.Bounds.W-4), int(area.Bounds.Y+area.Bounds.H-4))
				for y := y0; y < y1; y++ {
					for x := x0; x < x1; x++ {
						want := reference.Image().RGBAAt(x, y)
						painted = painted || want != background
						if got := actual.Image().RGBAAt(x, y); got != want {
							t.Fatalf("source grid border/50 pixel=%v want%v", got, want)
						}
					}
				}
				if !painted {
					t.Fatal("grid reference painted no visible line")
				}
			})
		}
	}
}
