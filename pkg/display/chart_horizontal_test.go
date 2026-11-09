package display

import (
	"bytes"
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"image/color"
	"math"
	"strings"
	"testing"
	"time"
)

func horizontalChartTester(p *ChartProps) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, *p) }) }, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func horizontalChartAccess(t *testing.T, r *ui.Tester, name string) ui.AccessInfo {
	t.Helper()
	v, ok := r.AccessByRole(ui.RoleGroup, name)
	if !ok {
		t.Fatalf("missing plot %q: %+v", name, r.AccessAll())
	}
	return v
}
func horizontalChartMark(t *testing.T, r *ui.Tester, name string) ui.Rect {
	t.Helper()
	v, ok := r.Find(name)
	if !ok {
		t.Fatalf("missing native mark %q: %+v", name, r.AccessAll())
	}
	return v
}
func horizontalNear(t *testing.T, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > .15 {
		t.Fatalf("geometry got%v want%v", got, want)
	}
}
func TestChartHorizontalGroupsAreNativeRowsInCategoryAndSeriesOrder(t *testing.T) {
	p := ChartProps{Kind: BarChart, Orientation: ChartHorizontal, Height: 260, Label: "Horizontal plot", HideTooltip: true, Labels: []string{"January", "February"}, Series: []ChartSeries{{Name: "Desktop", Values: []float64{20, 40}, Color: ui.RGB(220, 30, 30)}, {Name: "Mobile", Values: []float64{10, 30}, Color: ui.RGB(30, 30, 220)}}}
	r := horizontalChartTester(&p)
	area := horizontalChartAccess(t, r, p.Label).Bounds
	a := horizontalChartMark(t, r, "Desktop January: 20")
	b := horizontalChartMark(t, r, "Mobile January: 10")
	c := horizontalChartMark(t, r, "Desktop February: 40")
	horizontalNear(t, a.X, area.X)
	horizontalNear(t, b.X, area.X)
	horizontalNear(t, a.W, 2*b.W)
	if !(a.Y < b.Y && b.Y < c.Y && a.W > a.H) {
		t.Fatalf("horizontal categories/groups not top-to-bottom rows: area%v a%v b%v c%v", area, a, b, c)
	}
	label, ok := r.AccessByRole(ui.RoleText, "January")
	if !ok || label.Bounds.X+label.Bounds.W > area.X {
		t.Fatal("category axis is not beside the horizontal rows")
	}
	pixel := r.Image().RGBAAt(int(a.X+a.W/2), int(a.Y+a.H/2))
	if pixel.R < 180 || pixel.B > 80 {
		t.Fatalf("actual horizontal series fill missing: %v", pixel)
	}
}
func TestChartHorizontalInspectionUsesYAndNativeScopedKeys(t *testing.T) {
	selected := -1
	outside := ""
	p := ChartProps{Orientation: ChartHorizontal, Height: 260, Label: "Inspect rows", Labels: []string{"First", "Missing", "Third", "Last"}, Selected: &selected, Series: []ChartSeries{{Name: "Visitors", Values: []float64{10, math.NaN(), 30, 40}}}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p); control.Input(c, &outside, control.InputProps{Label: "Outside editor"}) })
	}, 600, 440, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	area := horizontalChartAccess(t, r, p.Label).Bounds
	r.Click(p.Label)
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyDown)
	if selected != 2 || horizontalChartAccess(t, r, p.Label).Value != "Third, Visitors: 30" {
		t.Fatalf("native Down did not skip missing observation: %d", selected)
	}
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyUp)
	if selected != 2 {
		t.Fatal("native Up did not inspect previous valid row")
	}
	r.Key(0, ui.KeyEscape)
	if selected != -1 || horizontalChartAccess(t, r, p.Label).Value != "" {
		t.Fatal("Escape did not clear real native inspection value")
	}
	r.Move(area.X+area.W*.8, area.Y+area.H*.12)
	if selected != 0 {
		t.Fatalf("top row pointer chose%d", selected)
	}
	r.Move(area.X+area.W*.8, area.Y+area.H*.88)
	if selected != 3 {
		t.Fatalf("pointer Y change did not inspect bottom row: %d", selected)
	}
	r.Move(1, 1)
	if selected != -1 {
		t.Fatal("pointer leave retained inspected category")
	}
	r.Click("Outside editor")
	r.Key(0, ui.KeyDown)
	r.Type("Text")
	if selected != -1 || outside != "Text" {
		t.Fatal("chart shortcut escaped its focused native plot")
	}
}
func TestChartCartesianHiddenAxesReclaimNativePlotTracks(t *testing.T) {
	p := ChartProps{Orientation: ChartHorizontal, Height: 240, Label: "Bare plot", Labels: []string{"January", "February"}, Series: []ChartSeries{{Values: []float64{10, 20}}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}}
	r := horizontalChartTester(&p)
	area := horizontalChartAccess(t, r, p.Label).Bounds
	horizontalNear(t, area.W, 560)
	horizontalNear(t, area.H, 240)
	if _, ok := r.AccessByRole(ui.RoleText, "January"); ok {
		t.Fatal("hidden category axis stayed in native accessibility tree")
	}
}
func TestChartHorizontalFractionalAndTinyValuesUseActualDomain(t *testing.T) {
	for _, scale := range []float64{1, 1e-99} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			p := ChartProps{Orientation: ChartHorizontal, Height: 220, Label: "Fractional plot", Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Value", Values: []float64{.2 * scale, .4 * scale}}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true}}
			r := horizontalChartTester(&p)
			area := horizontalChartAccess(t, r, p.Label).Bounds
			a := horizontalChartMark(t, r, fmt.Sprintf("Value A: %g", .2*scale))
			b := horizontalChartMark(t, r, fmt.Sprintf("Value B: %g", .4*scale))
			if b.W < area.W*.8 || a.W < area.W*.4 {
				t.Fatalf("finite small measurements collapsed into fallback domain: area%v a%v b%v", area, a, b)
			}
			horizontalNear(t, a.W, b.W/2)
		})
	}
}
func TestChartHorizontalLongCategoryHasBudgetAndOriginalInspectionLabel(t *testing.T) {
	label := strings.Repeat("Long category ", 60)
	selected := -1
	p := ChartProps{Orientation: ChartHorizontal, Label: "Long labels", Labels: []string{label, "Short"}, Selected: &selected, Series: []ChartSeries{{Name: "Visitors", Values: []float64{10, 20}}}}
	r := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p) }) }, 300, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	area := horizontalChartAccess(t, r, p.Label).Bounds
	text, ok := r.AccessByRole(ui.RoleText, label)
	if !ok || text.Bounds.W > 60.1 {
		t.Fatalf("native category label has no horizontal width budget: %v", text.Bounds)
	}
	if area.W < 150 {
		t.Fatalf("long native category consumed plot budget: %+v", area)
	}
	r.Click(p.Label)
	r.Key(0, ui.KeyHome)
	if !strings.HasPrefix(horizontalChartAccess(t, r, p.Label).Value, label+",") {
		t.Fatal("axis truncation changed accessible original data")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reduced horizontal chart did not settle")
	}
}

func TestChartCartesianScaledPointerUsesUntransformedLayoutDimensions(t *testing.T) {
	for _, orientation := range []ChartOrientation{ChartVertical, ChartHorizontal} {
		t.Run(fmt.Sprint(orientation), func(t *testing.T) {
			selected := -1
			p := ChartProps{Orientation: orientation, Label: "Scaled plot", Height: 240, Labels: []string{"A", "B"}, Selected: &selected, Series: []ChartSeries{{Values: []float64{10, 20}}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true}}
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p).Scale(.5, .5) })
			}, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			area := horizontalChartAccess(t, r, p.Label).Bounds
			x, y := area.X+area.W*.49, area.Y+area.H*.4
			if orientation == ChartHorizontal {
				x, y = area.X+area.W*.25, area.Y+area.H*.49
			}
			r.Move(x, y)
			if selected != 0 {
				t.Fatalf("scaled native pointer fraction .49 selected%d", selected)
			}
			if orientation == ChartHorizontal {
				y = area.Y + area.H*.51
			} else {
				x = area.X + area.W*.51
			}
			r.Move(x, y)
			if selected != 1 {
				t.Fatalf("scaled native pointer fraction .51 selected%d", selected)
			}
		})
	}
}
func TestChartHorizontalNormalizedExtremeAxisUsesFractionsAndTooltipKeepsCounts(t *testing.T) {
	selected := 0
	p := ChartProps{Orientation: ChartHorizontal, Stacked: true, Normalized: true, Label: "Normalized plot", Labels: []string{"A", "B"}, Selected: &selected, Series: []ChartSeries{{Name: "Desktop", Values: []float64{1e200, -3e200}}, {Name: "Mobile", Values: []float64{3e200, -1e200}}}, FormatValue: func(v float64) string { return fmt.Sprintf("%g visits", v) }, Cartesian: ChartCartesianProps{FormatValue: func(v float64) string { return fmt.Sprintf("%.0f%%", v*100) }}}
	r := horizontalChartTester(&p)
	if !r.HasText("100%") || !r.HasText("-100%") || !r.HasText("1e+200 visits") || !r.HasText("3e+200 visits") {
		t.Fatalf("normalized domain/tooltip units mixed: %v", r.Texts())
	}
	area := horizontalChartAccess(t, r, p.Label).Bounds
	positive := horizontalChartMark(t, r, "Desktop A: 1e+200")
	negative := horizontalChartMark(t, r, "Desktop B: -3e+200")
	if positive.X < area.X+area.W*.49 || negative.X >= positive.X {
		t.Fatalf("diverging normalized marks did not share native zero baseline: area%v pos%v neg%v", area, positive, negative)
	}
}

func TestChartHorizontalDivergingStacksNormalizeAbsoluteMagnitudeAndKeepZeroInspection(t *testing.T) {
	for _, normalized := range []bool{false, true} {
		t.Run(fmt.Sprint(normalized), func(t *testing.T) {
			selected := -1
			p := ChartProps{Orientation: ChartHorizontal, Stacked: true, Normalized: normalized, Height: 300, Label: "Diverging rows", HideTooltip: true, Selected: &selected, Labels: []string{"A", "B", "Zero", "Missing"}, Series: []ChartSeries{{Name: "Desktop", Values: []float64{4, -3, 0, math.NaN()}}, {Name: "Mobile", Values: []float64{6, -1, 0, math.Inf(1)}}, {Name: "Other", Values: []float64{-2, 2, 0, math.Inf(-1)}}}}
			r := horizontalChartTester(&p)
			area := horizontalChartAccess(t, r, p.Label).Bounds
			first := horizontalChartMark(t, r, "Desktop A: 4")
			second := horizontalChartMark(t, r, "Mobile A: 6")
			negative := horizontalChartMark(t, r, "Other A: -2")
			left := horizontalChartMark(t, r, "Desktop B: -3")
			farleft := horizontalChartMark(t, r, "Mobile B: -1")
			horizontalNear(t, first.Y, second.Y)
			horizontalNear(t, first.H, second.H)
			horizontalNear(t, first.X+first.W, second.X)
			horizontalNear(t, negative.X+negative.W, first.X)
			horizontalNear(t, left.X+left.W, first.X)
			horizontalNear(t, farleft.X+farleft.W, left.X)
			if normalized {
				horizontalNear(t, first.X, area.X+area.W/2)
				horizontalNear(t, first.W, area.W/6)
				horizontalNear(t, second.W, area.W/4)
				horizontalNear(t, negative.W, area.W/12)
			}
			if _, ok := r.Find("Desktop Missing: NaN"); ok {
				t.Fatal("missing observation was drawn as a measurement")
			}
			r.Click(p.Label)
			r.Key(0, ui.KeyEnd)
			if selected != 2 || !strings.HasPrefix(horizontalChartAccess(t, r, p.Label).Value, "Zero, Desktop: 0") {
				t.Fatal("zero measurement was discarded or nonfinite category was selected")
			}
			for _, n := range r.AccessAll() {
				for _, v := range []float32{n.Bounds.X, n.Bounds.Y, n.Bounds.W, n.Bounds.H} {
					if !finite(float64(v)) {
						t.Fatalf("nonfinite native bound: %+v", n)
					}
				}
			}
		})
	}
}
func TestChartBarCategoryFillsPaintLegendAndActualTooltip(t *testing.T) {
	for _, orientation := range []ChartOrientation{ChartVertical, ChartHorizontal} {
		t.Run(fmt.Sprint(orientation), func(t *testing.T) {
			selected := -1
			red, blue, green := ui.RGB(220, 30, 30), ui.RGB(30, 30, 220), ui.RGB(30, 180, 30)
			p := ChartProps{Orientation: orientation, Height: 240, Label: "Category colors", Labels: []string{"Chrome", "Safari"}, Selected: &selected, Legend: true, CategoryColors: []ui.Color{red, blue}, Series: []ChartSeries{{Name: "Visitors", Values: []float64{10, 20}, Color: green}}}
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Children(func() {
					p.TooltipContent = func(data ChartTooltipData) {
						ui.Box(c).Size(12, 12).Background(data.Items[0].Color).Label("Actual payload color")
						ui.Text(c, fmt.Sprint(data.Items[0].Value))
					}
					Chart(c, p)
				})
			}, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			a := horizontalChartMark(t, r, "Visitors Chrome: 10")
			b := horizontalChartMark(t, r, "Visitors Safari: 20")
			pixel := r.Image().RGBAAt(int(a.X+a.W/2), int(a.Y+a.H/2))
			if pixel.R != red.R || pixel.B != red.B {
				t.Fatalf("category fill did not paint native mark: %v", pixel)
			}
			pixel = r.Image().RGBAAt(int(b.X+b.W/2), int(b.Y+b.H/2))
			if pixel.B != blue.B || pixel.R != blue.R {
				t.Fatalf("second category fill was lost: %v", pixel)
			}
			legend := horizontalChartAccess(t, r, "Chart legend").Bounds
			redCount, blueCount := 0, 0
			painted := r.Image()
			for y := int(legend.Y); y < int(legend.Y+legend.H); y++ {
				for x := int(legend.X); x < int(legend.X+legend.W); x++ {
					v := painted.RGBAAt(x, y)
					if v.R == red.R && v.B == red.B {
						redCount++
					}
					if v.B == blue.B && v.R == blue.R {
						blueCount++
					}
				}
			}
			if redCount < 10 || blueCount < 10 {
				t.Fatal("category legend did not paint both actual fills")
			}
			r.Click(p.Label)
			r.Key(0, ui.KeyHome)
			swatch := horizontalChartMark(t, r, "Actual payload color")
			pixel = r.Image().RGBAAt(int(swatch.X+6), int(swatch.Y+6))
			if pixel.R != red.R || pixel.B != red.B || !r.HasText("10") {
				t.Fatal("tooltip did not receive original value and painted category color")
			}
		})
	}
}
func TestChartGroupedBarsKeepSeriesPaletteWhenCategoryColorsExist(t *testing.T) {
	p := ChartProps{Orientation: ChartHorizontal, Label: "Grouped colors", Labels: []string{"A"}, CategoryColors: []ui.Color{ui.RGB(0, 255, 0)}, Legend: true, Series: []ChartSeries{{Name: "Desktop", Values: []float64{10}, Color: ui.RGB(220, 30, 30)}, {Name: "Mobile", Values: []float64{20}, Color: ui.RGB(30, 30, 220)}}}
	r := horizontalChartTester(&p)
	mark := horizontalChartMark(t, r, "Desktop A: 10")
	pixel := r.Image().RGBAAt(int(mark.X+mark.W/2), int(mark.Y+mark.H/2))
	if pixel.R != 220 || pixel.G != 30 {
		t.Fatal("category palette ambiguously replaced multi-series data color")
	}
	if !r.HasText("Desktop") || !r.HasText("Mobile") {
		t.Fatal("grouped legend lost series names")
	}
}
func TestChartCartesianAxesStayPhysicalInRTLAndInheritNativeFont(t *testing.T) {
	for _, orientation := range []ChartOrientation{ChartVertical, ChartHorizontal} {
		for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
			t.Run(fmt.Sprintf("%d/%d", orientation, direction), func(t *testing.T) {
				p := ChartProps{Orientation: orientation, Label: "Directional plot", HideTooltip: true, Labels: []string{"September", "October"}, Series: []ChartSeries{{Name: "Value", Values: []float64{10, 20}}}, Cartesian: ChartCartesianProps{CategoryAxisMaxWidth: 200, FormatValue: func(v float64) string { return fmt.Sprintf("N%.0f", v) }}}
				r := ui.NewTesterWith(func(c *ui.Context) {
					theme.Apply(c)
					ui.Column(c).Fill().Padding(20).Font("monospace").FontWeight(700).Direction(direction).Children(func() { ui.Text(c, "September").FontSize(12).Label("Font reference").AlignSelf(ui.Start); Chart(c, p) })
				}, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				area := horizontalChartAccess(t, r, p.Label).Bounds
				reference, _ := r.AccessByRole(ui.RoleText, "Font reference")
				category, _ := r.AccessByRole(ui.RoleText, "September")
				zero, _ := r.AccessByRole(ui.RoleText, "N0")
				max, _ := r.AccessByRole(ui.RoleText, "N20")
				horizontalNear(t, category.Bounds.W, reference.Bounds.W)
				if orientation == ChartHorizontal {
					if category.Bounds.X+category.Bounds.W > area.X+.1 {
						t.Fatalf("RTL moved category axis into wrong physical side: label%v area%v", category.Bounds, area)
					}
					horizontalNear(t, zero.Bounds.X+zero.Bounds.W/2, area.X)
					horizontalNear(t, max.Bounds.X+max.Bounds.W/2, area.X+area.W)
				} else {
					if zero.Bounds.X+zero.Bounds.W > area.X+.1 {
						t.Fatal("RTL moved numeric axis into wrong physical side")
					}
					horizontalNear(t, category.Bounds.X+category.Bounds.W/2, area.X+area.W*.25)
				}
			})
		}
	}
}
func TestChartCustomCartesianTooltipInheritsAncestorForeground(t *testing.T) {
	glyph := ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path fill="currentColor" d="M0 0h10v10H0z"/></svg>`))
	selected := 0
	foreground := ui.RGB(40, 130, 80)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).TextColor(foreground).Children(func() {
			Chart(c, ChartProps{Orientation: ChartHorizontal, Label: "Custom tooltip", Labels: []string{"A"}, Selected: &selected, Series: []ChartSeries{{Name: "Value", Values: []float64{10}}}, TooltipContent: func(ChartTooltipData) { ui.Icon(c, glyph).Size(10, 10).Label("Inherited foreground") }})
		})
	}, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	icon := horizontalChartMark(t, r, "Inherited foreground")
	pixel := r.Image().RGBAAt(int(icon.X+5), int(icon.Y+5))
	if pixel.R != foreground.R || pixel.G != foreground.G || pixel.B != foreground.B {
		t.Fatalf("plot axis color polluted actual custom tooltip: %v", pixel)
	}
}

func TestChartCartesianFormatCategoryAndVisualOptionsPreserveOriginalInspection(t *testing.T) {
	selected := -1
	p := ChartProps{Orientation: ChartHorizontal, Height: 240, Label: "Configured plot", Labels: []string{"Alpha category", "Beta category"}, Selected: &selected, Series: []ChartSeries{{Name: "Value", Values: []float64{20, 100}, Color: ui.RGB(220, 30, 30)}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true}}
	tk := theme.Light()
	tk.Border = ui.RGB(10, 30, 50)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Scope(c, tk).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
	}, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	area := horizontalChartAccess(t, r, p.Label).Bounds
	painted := func(x, y float32) bool {
		for py := int(y) - 1; py <= int(y)+1; py++ {
			for px := int(x) - 1; px <= int(x)+1; px++ {
				v := r.Image().RGBAAt(px, py)
				if v.R != 255 || v.G != 255 || v.B != 255 {
					return true
				}
			}
		}
		return false
	}
	x, y := area.X+area.W*.4, area.Y+area.H*.25
	if !painted(x, y) {
		t.Fatal("default native numeric grid was absent")
	}
	p.Cartesian.HideGrid = true
	r.Frame()
	if painted(x, y) {
		t.Fatal("HideGrid left painted grid under empty part of row")
	}
	selected = 0
	r.Frame()
	if !painted(x, y) {
		t.Fatal("selected row cursor did not paint across real plot")
	}
	p.Cartesian.HideCursor = true
	r.Frame()
	if painted(x, y) {
		t.Fatal("HideCursor left selected row rule painted")
	}
	p.Cartesian.HideCategoryAxis = false
	p.Cartesian.FormatCategory = func(s string) string { return s[:3] }
	r.Frame()
	if !r.HasText("Alp") || !strings.HasPrefix(horizontalChartAccess(t, r, p.Label).Value, "Alpha category,") {
		t.Fatal("axis formatter changed original inspection label or failed actual tick")
	}
}
func TestChartBarRadiusAllowsSquareAndRejectsInvalidValues(t *testing.T) {
	p := ChartProps{Orientation: ChartHorizontal, Height: 240, Label: "Radius plot", HideTooltip: true, Labels: []string{"A"}, Series: []ChartSeries{{Name: "Value", Values: []float64{10}, Color: ui.RGB(220, 30, 30)}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true}}
	r := horizontalChartTester(&p)
	baseline := bytes.Clone(r.Image().Pix)
	for _, value := range []float32{-1, float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		p.Cartesian.BarRadius = &value
		r.Frame()
		if !bytes.Equal(baseline, r.Image().Pix) {
			t.Fatalf("invalid radius changed default native shape: %v", value)
		}
	}
	square := float32(0)
	p.Cartesian.BarRadius = &square
	r.Frame()
	mark := horizontalChartMark(t, r, "Value A: 10")
	sample := func() color.RGBA { return r.Image().RGBAAt(int(mark.X+1), int(mark.Y+1)) }
	if v := sample(); v.R != 220 || v.G != 30 {
		t.Fatalf("radius0 did not paint square corner: %v", v)
	}
	rounded := float32(12)
	p.Cartesian.BarRadius = &rounded
	r.Frame()
	if v := sample(); v.R == 220 && v.G == 30 {
		t.Fatal("finite radius did not round actual native corner")
	}
}
func TestChartDynamicOrientationKeepsFocusedPlotAndLocalSelection(t *testing.T) {
	p := ChartProps{Height: 240, Label: "Dynamic plot", Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Value", Values: []float64{20, 40}}}}
	r := horizontalChartTester(&p)
	r.Click(p.Label)
	r.Key(0, ui.KeyEnd)
	before := horizontalChartAccess(t, r, p.Label).Value
	p.Orientation = ChartHorizontal
	r.Frame()
	if !r.Focused(p.Label) || horizontalChartAccess(t, r, p.Label).Value != before {
		t.Fatal("direction change recreated focused native plot or reset local selection")
	}
	r.Key(0, ui.KeyUp)
	if !strings.HasPrefix(horizontalChartAccess(t, r, p.Label).Value, "A,") {
		t.Fatal("new direction did not use native Up after preserving focus")
	}
	p.Series[0].Values[0] = math.NaN()
	p.Labels = append(p.Labels, "C")
	p.Series[0].Values = append(p.Series[0].Values, 60)
	r.Frame()
	r.Key(0, ui.KeyEnd)
	if !r.Focused(p.Label) || !strings.HasPrefix(horizontalChartAccess(t, r, p.Label).Value, "C,") {
		t.Fatal("stream append/missing data lost native plot identity or valid navigation")
	}
	p.Orientation = ChartVertical
	r.Frame()
	r.Frame()
	if !strings.HasPrefix(horizontalChartAccess(t, r, p.Label).Value, "C,") {
		t.Fatal("stationary pointer overrode selected category during direction layout change")
	}
	r.Key(0, ui.KeyLeft)
	if !strings.HasPrefix(horizontalChartAccess(t, r, p.Label).Value, "B,") {
		t.Fatal("vertical keyboard compatibility did not return")
	}
}
func TestChartOrientationDoesNotTransposeNonBarKinds(t *testing.T) {
	for _, kind := range []ChartKind{LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			p := ChartProps{Kind: kind, Height: 220, Label: "Unchanged plot", Labels: []string{"A", "B", "C"}, Series: []ChartSeries{{Values: []float64{10, 20, 30}}}}
			r := horizontalChartTester(&p)
			before := bytes.Clone(r.Image().Pix)
			p.Orientation = ChartHorizontal
			r.Frame()
			if !bytes.Equal(before, r.Image().Pix) {
				t.Fatal("bar-only orientation altered another chart kind")
			}
		})
	}
}
func TestChartHorizontalReducedMotionFinishesActiveNativeColorTransition(t *testing.T) {
	p := ChartProps{Orientation: ChartHorizontal, Height: 240, Label: "Animated bars", HideTooltip: true, Labels: []string{"A", "Anchor"}, Series: []ChartSeries{{Name: "Value", Values: []float64{20, 100}, Color: ui.RGB(220, 30, 30)}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true}}
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p) }) }, 600, 400)
	// First settle the independent 400ms data entrance, then isolate color feedback.
	r.Advance(time.Second)
	mark := horizontalChartMark(t, r, "Value A: 20")
	sample := func() color.RGBA { return r.Image().RGBAAt(int(mark.X+mark.W/2), int(mark.Y+mark.H/2)) }
	before := sample()
	p.Series[0].Color = ui.RGB(30, 30, 220)
	r.Frame()
	r.Advance(50 * time.Millisecond)
	middle := sample()
	if middle == before || middle == (color.RGBA{30, 30, 220, 255}) {
		t.Fatalf("native feedback did not show intermediate color: before%v middle%v", before, middle)
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	if got := sample(); got != (color.RGBA{30, 30, 220, 255}) {
		t.Fatalf("reduced preference left old native color in flight: %v", got)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("settled reduced native marks kept requesting frames")
	}
}
func TestChartHorizontalMissingObservationKeepsLaterNativeMarkIdentity(t *testing.T) {
	p := ChartProps{Orientation: ChartHorizontal, Height: 280, Label: "Keyed bars", HideTooltip: true, Labels: []string{"A", "B", "Anchor"}, CategoryColors: []ui.Color{ui.RGB(220, 30, 30), ui.RGB(30, 30, 220), ui.RGB(30, 180, 30)}, Series: []ChartSeries{{Name: "Value", Values: []float64{10, 20, 100}}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true}}
	// Isolate the existing native color/key contract from source filtered-rectangle
	// data matching, which can intentionally move the surviving geometry.
	p.Animation.Disabled = true
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p) }) }, 600, 400)
	mark := horizontalChartMark(t, r, "Value B: 20")
	p.Series[0].Values[0] = math.NaN()
	r.Frame()
	got := r.Image().RGBAAt(int(mark.X+mark.W/2), int(mark.Y+mark.H/2))
	if got != (color.RGBA{30, 30, 220, 255}) {
		t.Fatalf("removed earlier observation reused another bar's animation identity: %v", got)
	}
	if r.FramePending() {
		t.Fatal("unchanged later data animated because another observation disappeared")
	}
}

func TestChartHorizontalMissingOnlyRaggedAndMalformedHeightRemainFinite(t *testing.T) {
	for _, height := range []float32{0, -1, float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		t.Run(fmt.Sprint(height), func(t *testing.T) {
			selected := 2
			p := ChartProps{Orientation: ChartHorizontal, Height: height, Label: "Empty rows", Labels: []string{"A", "B", "C"}, Selected: &selected, Series: []ChartSeries{{Values: []float64{math.NaN(), math.Inf(1)}}, {Values: []float64{math.Inf(-1)}}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true}}
			r := horizontalChartTester(&p)
			plot := horizontalChartAccess(t, r, p.Label)
			if selected != -1 || plot.Value != "" {
				t.Fatal("missing-only native chart exposed an invented measurement")
			}
			horizontalNear(t, plot.Bounds.H, 220)
			r.Click(p.Label)
			r.Key(0, ui.KeyDown)
			r.Key(0, ui.KeyEnd)
			if selected != -1 {
				t.Fatal("empty native chart selected a missing row")
			}
			for _, n := range r.AccessAll() {
				for _, v := range []float32{n.Bounds.X, n.Bounds.Y, n.Bounds.W, n.Bounds.H} {
					if !finite(float64(v)) {
						t.Fatalf("nonfinite accessible plot: %+v", n)
					}
				}
			}
		})
	}
}
