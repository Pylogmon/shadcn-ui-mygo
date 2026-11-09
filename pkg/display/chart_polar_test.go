package display

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type polarOracleSector struct{ Cx, Cy, InnerRadius, OuterRadius, StartAngle, EndAngle, Value float64 }

func polarNear(t *testing.T, got, want float64) {
	t.Helper()
	if !finite(got) || !finite(want) || math.Abs(got-want) > 1e-4 {
		t.Fatalf("source geometry got%.9g want%.9g", got, want)
	}
}
func TestChartPolarGeometryMatchesActualOfficialFunctions(t *testing.T) {
	var oracle struct {
		PieCases []struct {
			Settings struct {
				Name                     string
				InnerRadius, OuterRadius any
			}
			Result []polarOracleSector
		}
		RadialCases []struct {
			Start, End, Inner, Outer, Maximum float64
			Result                            []polarOracleSector
		}
		Radar struct {
			Ticks  []float64
			Result []struct{ X, Y, Radius, Angle, Value float64 }
		}
	}
	b, err := os.ReadFile("testdata/shadcn-chart/polar/polar-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	capture := func(p ChartProps) chartPolarFrame {
		var f chartPolarFrame
		ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			c.DisableAnimations()
			area := ui.Box(c).Size(300, 250)
			f = chartPolarSourceFrame(c, area, p, ui.Rect{W: 300, H: 250})
		}, 300, 250)
		return f
	}
	for _, sample := range oracle.PieCases {
		p := ChartProps{Kind: PieChart, Polar: &ChartPolarProps{}, Labels: []string{"A", "B", "C", "D", "E"}, Series: []ChartSeries{{Values: []float64{275, 200, 187, 173, 90}}}}
		p.Polar.InnerRadius = cartesianPtr(ChartRadiusDIP(float32(sample.Settings.InnerRadius.(float64))))
		if v, ok := sample.Settings.OuterRadius.(float64); ok {
			p.Polar.OuterRadius = cartesianPtr(ChartRadiusDIP(float32(v)))
		}
		f := capture(p)
		if len(f.sectors) != len(sample.Result) {
			t.Fatal("source pie sector count")
		}
		for i, s := range f.sectors {
			want := sample.Result[i]
			polarNear(t, s.StartAngle, want.StartAngle)
			polarNear(t, s.EndAngle, want.EndAngle)
			polarNear(t, float64(s.InnerRadius.resolve(120)), want.InnerRadius)
			polarNear(t, float64(s.OuterRadius.resolve(120)), want.OuterRadius)
		}
	}
	for _, sample := range oracle.RadialCases {
		values := []float64{275, 200, 187, 173, 90}
		labels := []string{"A", "B", "C", "D", "E"}
		if sample.End == 180 {
			values = []float64{570}
			labels = []string{"one"}
		}
		p := ChartProps{Kind: RadialChart, Labels: labels, Series: []ChartSeries{{Values: values}}, MaxValue: sample.Maximum, Polar: &ChartPolarProps{StartAngle: &sample.Start, EndAngle: &sample.End, InnerRadius: cartesianPtr(ChartRadiusDIP(float32(sample.Inner))), OuterRadius: cartesianPtr(ChartRadiusDIP(float32(sample.Outer)))}}
		f := capture(p)
		for i, s := range f.sectors {
			want := sample.Result[i]
			polarNear(t, s.StartAngle, want.StartAngle)
			polarNear(t, s.EndAngle, want.EndAngle)
			polarNear(t, float64(s.InnerRadius.resolve(120)), want.InnerRadius)
			polarNear(t, float64(s.OuterRadius.resolve(120)), want.OuterRadius)
		}
	}
	p := ChartProps{Kind: RadarChart, Polar: &ChartPolarProps{}, Labels: []string{"M0", "M1", "M2", "M3", "M4", "M5"}, Series: []ChartSeries{{Values: []float64{186, 305, 237, 73, 209, 214}}}}
	f := capture(p)
	polarNear(t, f.maximum, oracle.Radar.Ticks[len(oracle.Radar.Ticks)-1])
	for i, v := range f.radar[0] {
		polarNear(t, 150+v.v[0], oracle.Radar.Result[i].X)
		polarNear(t, 125+v.v[1], oracle.Radar.Result[i].Y)
	}
	for _, kind := range []ChartKind{PieChart, DonutChart, RadialChart, RadarChart} {
		p := ChartProps{Kind: kind, Label: "Empty source", Polar: &ChartPolarProps{Grid: &ChartPolarGridProps{}}}
		r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); c.DisableAnimations(); Chart(c, p) }, 300, 250)
		if !r.HasText("No data") {
			t.Fatalf("empty source kind%d invented data", kind)
		}
	}
}

func TestChartPolarRoundedSectorsRenderLikeOfficialSVGPaths(t *testing.T) {
	var oracle struct {
		SectorCases []struct {
			Props struct{ Cx, Cy, InnerRadius, OuterRadius, StartAngle, EndAngle, CornerRadius float64 }
			Path  string
		}
	}
	b, err := os.ReadFile("testdata/shadcn-chart/polar/polar-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	for _, sample := range oracle.SectorCases {
		v := sample.Props
		path := chartPolarSectorPath(float32(v.Cx), float32(v.Cy), float32(v.InnerRadius), float32(v.OuterRadius), v.StartAngle, v.EndAngle, float32(v.CornerRadius))
		actual := ui.Render(func(c *ui.Context) {
			ui.Box(c).Fill().Background(ui.RGB(255, 255, 255)).Draw(func(p *ui.Painter, _ ui.Rect) { p.FillPath(path, ui.RGB(20, 80, 180)) })
		}, 300, 250, 1)
		svg := ui.MustParseSVG([]byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 250"><path d="%s" fill="#1450b4"/></svg>`, sample.Path)))
		expected := ui.Render(func(c *ui.Context) {
			ui.Box(c).Fill().Background(ui.RGB(255, 255, 255)).Children(func() { ui.Icon(c, svg).Size(300, 250).TextColor(ui.RGB(20, 80, 180)) })
		}, 300, 250, 1)
		// SVG icons and Painter paths have separate native flattening/cache
		// rasterization. Compare the filled contour within one device pixel,
		// rather than equating their edge antialias coverage/color bytes.
		inside := func(img interface{ RGBAAt(int, int) color.RGBA }, x, y int) bool { return img.RGBAAt(x, y).R < 138 }
		far, filled := 0, 0
		for y := 1; y < 249; y++ {
			for x := 1; x < 299; x++ {
				a, b := inside(actual, x, y), inside(expected, x, y)
				if a {
					filled++
				}
				if a == b {
					continue
				}
				aNear, bNear := false, false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						aNear = aNear || inside(expected, x+dx, y+dy) == a
						bNear = bNear || inside(actual, x+dx, y+dy) == b
					}
				}
				if !aNear || !bNear {
					far++
				}
			}
		}
		if far != 0 || filled < 500 {
			t.Fatalf("official sector %.0f..%.0f corner%.0f has%d contour mismatches beyond1px (%d filled)", v.StartAngle, v.EndAngle, v.CornerRadius, far, filled)
		}
	}
}

func sourcePolarTester(p *ChartProps) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, *p) }) }, 340, 310, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func TestChartPolarPublicExpandedSectorHitsRespectTheGapAndHole(t *testing.T) {
	selected := -1
	p := ChartProps{Kind: DonutChart, Label: "Expanded source", Height: 250, Selected: &selected, Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Visitors", Values: []float64{1, 1}, Style: ChartSeriesStyle{Polar: &ChartPolarSeriesProps{Shape: func(s ChartPolarSector) []ChartPolarSector {
		if s.Datum.Index != 0 {
			return []ChartPolarSector{s}
		}
		extra := s
		extra.InnerRadius, extra.OuterRadius = s.OuterRadius.Add(12), s.OuterRadius.Add(25)
		s.OuterRadius = s.OuterRadius.Add(10)
		return []ChartPolarSector{s, extra}
	}}}}}, Polar: &ChartPolarProps{InnerRadius: cartesianPtr(ChartRadiusDIP(60)), OuterRadius: cartesianPtr(ChartRadiusDIP(90))}}
	r := sourcePolarTester(&p)
	bounds := horizontalChartAccess(t, r, p.Label).Bounds
	cx, cy := bounds.X+bounds.W/2, bounds.Y+bounds.H/2
	r.Move(cx, cy-110)
	if selected != 0 {
		t.Fatal("expanded outer sector ring was not a real hit target")
	}
	r.Move(cx, cy-101)
	if selected != -1 {
		t.Fatal("empty space between the two layers hit a sector")
	}
	r.Move(cx, cy)
	if selected != -1 {
		t.Fatal("donut center hit")
	}
	r.Move(cx, cy+80)
	if selected != 1 {
		t.Fatal("source positive angle did not put second half below center")
	}
	p.Polar.StrokeWidth = cartesianPtr(float32(5))
	r.Frame()
	r.Move(cx, cy-101)
	if selected != 0 {
		t.Fatal("painted source stroke in the layered gap was not an item hit")
	}
}

func TestChartPolarPublicStackedPieInspectionChoosesOneRing(t *testing.T) {
	selected, series := -1, 0
	outside := ""
	p := ChartProps{Kind: PieChart, Stacked: true, Label: "Two source rings", Height: 250, Selected: &selected, Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Desktop", Values: []float64{10, 20}, Style: ChartSeriesStyle{Polar: &ChartPolarSeriesProps{OuterRadius: cartesianPtr(ChartRadiusDIP(60))}}}, {Name: "Mobile", Values: []float64{100, 200}, Style: ChartSeriesStyle{Polar: &ChartPolarSeriesProps{InnerRadius: cartesianPtr(ChartRadiusDIP(70)), OuterRadius: cartesianPtr(ChartRadiusDIP(90))}}}}, Polar: &ChartPolarProps{SelectedSeries: &series}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p); control.Input(c, &outside, control.InputProps{Label: "Outside polar"}) })
	}, 340, 360, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	area := horizontalChartAccess(t, r, p.Label).Bounds
	cx, cy := area.X+area.W/2, area.Y+area.H/2
	r.Move(cx, cy-80)
	if series != 1 || selected != 0 || horizontalChartAccess(t, r, p.Label).Value != "A, Mobile: 100" {
		t.Fatalf("outer ring merged item data: ring%d selected%d %+v", series, selected, horizontalChartAccess(t, r, p.Label))
	}
	r.ClickAt(cx, cy-80)
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyPageUp)
	if series != 0 || horizontalChartAccess(t, r, p.Label).Value != "A, Desktop: 10" {
		t.Fatal("Page Up did not inspect inner source ring")
	}
	r.Key(0, ui.KeyPageDown)
	if series != 1 {
		t.Fatal("Page Down did not restore outer source ring")
	}
	r.Click("Outside polar")
	r.Key(0, ui.KeyPageUp)
	r.Type("text")
	if series != 1 || outside != "text" {
		t.Fatal("polar ring key escaped focused plot")
	}
}

func TestChartPolarSourceAxisInspectionIncludesBlankTracksAndZeroRadar(t *testing.T) {
	var oracle struct {
		AxisHitCases []struct {
			Point    struct{ RelativeX, RelativeY float32 }
			Expected int
		}
	}
	data, err := os.ReadFile("testdata/shadcn-chart/polar/polar-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	selected := -1
	p := ChartProps{Kind: RadialChart, Label: "Axis source", Height: 250, Selected: &selected, Labels: []string{"Chrome", "Safari", "Firefox", "Edge", "Other"}, Series: []ChartSeries{{Name: "Visitors", Values: []float64{275, 200, 187, 173, 90}}}, Polar: &ChartPolarProps{InnerRadius: cartesianPtr(ChartRadiusDIP(30)), OuterRadius: cartesianPtr(ChartRadiusDIP(110))}}
	r := sourcePolarTester(&p)
	area := horizontalChartAccess(t, r, p.Label).Bounds
	for _, sample := range oracle.AxisHitCases {
		r.Move(area.X+sample.Point.RelativeX, area.Y+sample.Point.RelativeY)
		if selected != sample.Expected {
			t.Fatalf("source actual axis function expected%d got%d", sample.Expected, selected)
		}
	}
	p.Kind = RadarChart
	p.Labels = []string{"Zero", "B", "C"}
	p.Series[0].Values = []float64{0, 100, 50}
	p.Polar = &ChartPolarProps{}
	selected = -1
	r.Frame()
	area = horizontalChartAccess(t, r, p.Label).Bounds
	r.Move(area.X+area.W/2, area.Y+area.H/2-50)
	if selected != 0 || horizontalChartAccess(t, r, p.Label).Value != "Zero, Visitors: 0" {
		t.Fatal("zero Radar category lost source axis inspection")
	}
}

func TestChartPolarRadarGridNativeDotsAndLatestAxisData(t *testing.T) {
	green := ui.RGB(0, 180, 0)
	p := ChartProps{Kind: RadarChart, Label: "Radar source", Height: 250, Labels: []string{"A", "B", "C", "D", "E", "F"}, Series: []ChartSeries{{Name: "Value", Values: []float64{100, 100, 100, 100, 100, 100}, Color: ui.RGB(20, 40, 220), Style: ChartSeriesStyle{FillOpacity: cartesianPtr(float32(0)), Dots: &ChartDotProps{Radius: 4, StrokeWidth: cartesianPtr(float32(0)), Content: func(_ *ui.Context, d ChartDatum, e ui.Element) { e.Label("radar dot " + d.Label) }}, ActiveDot: &ChartDotProps{Disabled: true}}}}, Polar: &ChartPolarProps{Grid: &ChartPolarGridProps{Kind: ChartPolarCircle, Radii: []ChartRadius{ChartRadiusDIP(90)}, RadialLines: cartesianPtr(false), Stroke: &green}, AngleAxis: &ChartPolarAngleAxisProps{HideLine: true, HideTickLines: true, Content: func(c *ui.Context, d ChartPolarTick) { ui.Text(c, fmt.Sprintf("%s:%g", d.Label, d.Values[0])).NoWrap() }}}}
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p) }) }, 340, 310)
	dot := horizontalChartMark(t, r, "radar dot A")
	area := horizontalChartAccess(t, r, p.Label).Bounds
	cx, cy := area.X+area.W/2, area.Y+area.H/2
	if math.Abs(float64(dot.Y+dot.H/2-cy)) > 1 {
		t.Fatal("idle Radar dot did not enter from the center with the polygon")
	}
	r.Advance(750 * time.Millisecond)
	dot = horizontalChartMark(t, r, "radar dot A")
	if dot.Y+dot.H/2 >= cy-20 {
		t.Fatal("Radar idle dot did not follow animated points")
	}
	if got := r.Image().RGBAAt(int(cx+90), int(cy)); int(got.G)-int(got.R) < 60 {
		t.Fatalf("circle grid missing at its actual radius: %v", got)
	}
	p.Polar.Grid.Kind = ChartPolarPolygon
	r.Frame()
	if got := r.Image().RGBAAt(int(cx+90), int(cy)); int(got.G)-int(got.R) > 60 {
		t.Fatal("polygon grid still painted circle perimeter")
	}
	p.Series[0].Values[0] = 160
	r.Frame()
	if !r.HasText("A:160") {
		t.Fatal("custom angle label exposed animated rather than original value")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	r.Advance(2 * time.Second)
	if r.FramePending() {
		t.Fatal("new Radar profile did not settle under reduced motion")
	}
}

func TestChartPolarSourceAnimationKeepsRawInspectionAndScopedPaint(t *testing.T) {
	selected := 0
	calls := 0
	red := ui.RGBA(200, 20, 40, .5)
	p := ChartProps{Kind: PieChart, Label: "Animated source pie", Height: 250, Selected: &selected, Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Value", Values: []float64{1, 1}, Style: ChartSeriesStyle{Polar: &ChartPolarSeriesProps{Shape: func(s ChartPolarSector) []ChartPolarSector { calls++; return []ChartPolarSector{s} }}}}}, CategoryColors: []ui.Color{red, ui.RGB(20, 50, 200)}, Polar: &ChartPolarProps{StrokeWidth: cartesianPtr(float32(0))}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		c.Root().Background(ui.RGB(255, 255, 255))
		ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p) })
	}, 340, 310)
	r.Advance(300 * time.Millisecond)
	if horizontalChartAccess(t, r, p.Label).Value != "A, Value: 1" {
		t.Fatal("source delayed entrance changed raw inspection")
	}
	r.Advance(900 * time.Millisecond)
	p.Series[0].Values[0] = 4
	r.Frame()
	if horizontalChartAccess(t, r, p.Label).Value != "A, Value: 4" || !r.HasText("4") {
		t.Fatal("retargeted source tooltip/AX used intermediate values")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	r.Advance(2 * time.Second)
	before := calls
	img := r.Image()
	if calls != before {
		t.Fatal("native Draw invoked the application's Shape resolver")
	}
	area := horizontalChartAccess(t, r, p.Label).Bounds
	cx, cy := area.X+area.W/2, area.Y+area.H/2
	got := img.RGBAAt(int(cx), int(cy-50))
	ref := ui.Render(func(c *ui.Context) {
		ui.Box(c).Fill().Background(ui.RGB(255, 255, 255)).Draw(func(p *ui.Painter, _ ui.Rect) { p.FillPath((&ui.Path{}).Circle(5, 5, 5), red) })
	}, 10, 10, 1).RGBAAt(5, 5)
	if absInt(int(got.R)-int(ref.R)) > 1 || absInt(int(got.G)-int(ref.G)) > 1 || absInt(int(got.B)-int(ref.B)) > 1 || got.A != ref.A {
		t.Fatalf("polar fill alpha changed or scope paint borrowed Context: got%v native%v", got, ref)
	}
}
