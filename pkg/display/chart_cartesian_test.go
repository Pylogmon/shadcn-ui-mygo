package display

import (
	"encoding/json"
	"image"
	"math"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func cartesianPtr[T any](v T) *T { return &v }

func TestChartCartesianPathsMatchOfficialD3Oracle(t *testing.T) {
	var oracle struct {
		Cases []struct {
			Name, Curve string
			Points      [][2]float64
			Commands    [][]json.RawMessage
		}
	}
	data, err := os.ReadFile("testdata/shadcn-chart/cartesian/curve-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	for _, sample := range oracle.Cases {
		kind := map[string]ChartCurve{"natural": ChartCurveNatural, "monotone": ChartCurveMonotone, "linear": ChartCurveLinear, "step": ChartCurveStep}[sample.Curve]
		points := make([]chartGeometryDatum, len(sample.Points))
		for i, v := range sample.Points {
			points[i] = chartGeometryDatum{valid: true, v: [4]float64{v[0] / 500, 1 - v[1]/400}}
		}
		want := &ui.Path{}
		for _, cmd := range sample.Commands {
			var op string
			json.Unmarshal(cmd[0], &op)
			v := make([]float32, len(cmd)-1)
			for i := range v {
				json.Unmarshal(cmd[i+1], &v[i])
			}
			switch op {
			case "M":
				want.MoveTo(v[0], v[1])
			case "L":
				want.LineTo(v[0], v[1])
			case "C":
				want.CubeTo(v[0], v[1], v[2], v[3], v[4], v[5])
			case "Z":
				want.Close()
			}
		}
		got := chartCartesianPath(ui.Rect{W: 500, H: 400}, points, kind, false)
		render := func(path *ui.Path) *image.RGBA {
			return ui.Render(func(c *ui.Context) {
				ui.Box(c).Fill().Background(ui.RGB(255, 255, 255)).Draw(func(p *ui.Painter, _ ui.Rect) { p.StrokePath(path, 2, ui.RGB(20, 70, 180)) })
			}, 520, 420, 1)
		}
		a, b := render(got), render(want)
		different := 0
		for i := 0; i < len(a.Pix); i += 4 {
			if absInt(int(a.Pix[i])-int(b.Pix[i])) > 6 || absInt(int(a.Pix[i+1])-int(b.Pix[i+1])) > 6 || absInt(int(a.Pix[i+2])-int(b.Pix[i+2])) > 6 {
				different++
			}
		}
		if different > 12 {
			t.Fatalf("official %s/%s path differs at %d native pixels", sample.Curve, sample.Name, different)
		}
	}
}
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func TestChartCartesianSignedInnerMarginsAndScopedTickContent(t *testing.T) {
	selected := -1
	seenRaw := false
	p := ChartProps{Label: "Inset plot", Height: 260, HideTooltip: true, Selected: &selected, Labels: []string{"January", "February"}, Series: []ChartSeries{{Name: "Count", Values: []float64{20, 40}}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, Margin: &ChartInsets{Top: 12, Right: 20, Bottom: 14, Left: -10}}}
	r := horizontalChartTester(&p)
	area := horizontalChartAccess(t, r, p.Label).Bounds
	horizontalNear(t, area.X, 10)
	horizontalNear(t, area.Y, 32)
	horizontalNear(t, area.W, 550)
	horizontalNear(t, area.H, 234)
	p.Cartesian.HideValueAxis = false
	p.Cartesian.ValueTickCount = 3
	p.Cartesian.ValueTickMargin = cartesianPtr(float32(17))
	p.Cartesian.TickContent = func(c *ui.Context, tick ChartAxisTick) {
		if !tick.Category && tick.Value == 40 {
			seenRaw = true
		}
		ui.Text(c, "tick "+tick.Label).NoWrap()
	}
	r.Frame()
	if !seenRaw || !r.HasText("tick 40") {
		t.Fatal("current-build tick content did not receive original numeric tick")
	}
	r.Click(p.Label)
	r.Key(0, ui.KeyHome)
	if selected != 0 || horizontalChartAccess(t, r, p.Label).Value != "January, Count: 20" {
		t.Fatal("inner layout/formatting changed native raw inspection")
	}
}

func TestChartCartesianNativeBarCornersLabelsAndGroupGap(t *testing.T) {
	selected := -1
	p := ChartProps{Label: "Labeled groups", Height: 260, HideTooltip: true, Selected: &selected, Labels: []string{"Jan", "Feb"}, Series: []ChartSeries{{Name: "A", Values: []float64{40, 20}, Color: ui.RGB(220, 30, 30), Style: ChartSeriesStyle{BarRadius: &ChartCornerRadii{TopLeft: 14, TopRight: 14}, Labels: []ChartLabelProps{{Offset: cartesianPtr(float32(8)), Formatter: func(d ChartDatum) string { return "raw " + d.Label }}}}}, {Name: "B", Values: []float64{30, 10}, Color: ui.RGB(20, 40, 220)}}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, GroupGap: cartesianPtr(float32(6))}}
	r := horizontalChartTester(&p)
	a := horizontalChartMark(t, r, "A Jan: 40")
	b := horizontalChartMark(t, r, "B Jan: 30")
	horizontalNear(t, b.X-(a.X+a.W), 6)
	if color := r.Image().RGBAAt(int(a.X+1), int(a.Y+1)); color.R > 180 && color.G < 100 {
		t.Fatal("top-left corner did not round")
	}
	if color := r.Image().RGBAAt(int(a.X+1), int(a.Y+a.H-2)); color.R < 180 || color.G > 100 {
		t.Fatal("square bottom-left corner was rounded")
	}
	if !r.HasText("raw Jan") || !r.HasText("raw Feb") {
		t.Fatal("public label formatter did not build native text")
	}
	p.Orientation = ChartHorizontal
	p.Series[0].Style.Labels = []ChartLabelProps{{Position: ChartLabelInsideLeft, Offset: cartesianPtr(float32(8)), Formatter: func(d ChartDatum) string { return d.Label }}, {Position: ChartLabelRight, Offset: cartesianPtr(float32(8))}}
	r.Frame()
	a = horizontalChartMark(t, r, "A Jan: 40")
	b = horizontalChartMark(t, r, "B Jan: 30")
	horizontalNear(t, b.Y-(a.Y+a.H), 6)
	if !r.HasText("40") {
		t.Fatal("horizontal right value label missing")
	}
	selected = 0
	cursorColor := ui.RGB(90, 120, 90)
	p.Cartesian.Cursor = &ChartCursorProps{Band: true, Color: &cursorColor}
	r.Frame()
	area := horizontalChartAccess(t, r, p.Label).Bounds
	if got := r.Image().RGBAAt(int(area.X+area.W*.9), int(a.Y+a.H+3)); got.R != 90 || got.G != 120 || got.B != 90 {
		t.Fatalf("category cursor did not fill behind grouped marks: %v", got)
	}
}

func TestChartCartesianGradientAndCustomMarkSnapshots(t *testing.T) {
	{
		// Compare public Chart paint to direct native paint on the same
		// background. An existing alpha is multiplied only by FillOpacity.
		color := ui.RGBA(200, 20, 40, .5)
		style := ChartSeriesStyle{}
		reference := func(c *ui.Context) { ui.Box(c).Size(400, 100).Background(color) }
		pixels := ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			c.DisableAnimations()
			c.Root().Background(ui.RGB(255, 255, 255))
			ui.Column(c).Fill().Children(func() {
				Chart(c, ChartProps{Height: 100, Labels: []string{"One"}, Series: []ChartSeries{{Name: "Value", Values: []float64{10}, Color: color, Style: style}}, Cartesian: ChartCartesianProps{HideCategoryAxis: true, HideValueAxis: true, HideGrid: true}})
				reference(c)
			})
		}, 400, 220)
		img := pixels.Image()
		if got, want := img.RGBAAt(200, 80), img.RGBAAt(200, 180); got != want {
			t.Errorf("default translucent bar differs from same native color: got%v reference%v", got, want)
		}
		style = ChartSeriesStyle{FillOpacity: cartesianPtr(float32(.4)), Gradient: &ui.LinearGradient{Angle: 180, Stops: []ui.GradientStop{{Color: ui.RGBA(200, 20, 40, .8), At: .05}, {Color: ui.RGBA(200, 20, 40, .1), At: .95}}}}
		reference = func(c *ui.Context) {
			ui.Box(c).Size(400, 100).LinearGradient(ui.LinearGradient{Angle: 180, Stops: []ui.GradientStop{{Color: ui.RGBA(200, 20, 40, .32), At: .05}, {Color: ui.RGBA(200, 20, 40, .04), At: .95}}})
		}
		pixels.Frame()
		img = pixels.Image()
		for _, y := range []int{10, 90} {
			if got, want := img.RGBAAt(200, y), img.RGBAAt(200, y+100); got != want {
				t.Errorf("gradient stop alpha times fill opacity differs at y%d: got%v reference%v", y, got, want)
			}
		}
	}
	col := ui.RGB(220, 20, 20)
	blue := ui.RGB(20, 20, 220)
	selected := 1
	bars, dots := 0, 0
	p := ChartProps{Kind: AreaChart, Label: "Paint snapshot", Height: 240, Selected: &selected, HideTooltip: true, Labels: []string{"A", "B", "C"}, Series: []ChartSeries{{Name: "Value", Values: []float64{100, 100, 100}, Color: col, Style: ChartSeriesStyle{FillOpacity: cartesianPtr(float32(1)), StrokeWidth: cartesianPtr(float32(0)), Gradient: &ui.LinearGradient{Angle: 180, From: col, To: blue}}}}, Cartesian: ChartCartesianProps{PointScale: true, HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}}
	r := horizontalChartTester(&p)
	area := horizontalChartAccess(t, r, p.Label).Bounds
	top := r.Image().RGBAAt(int(area.X+area.W/2), int(area.Y+30))
	bottom := r.Image().RGBAAt(int(area.X+area.W/2), int(area.Y+area.H-30))
	if top.R < top.B || bottom.B < bottom.R {
		t.Fatalf("actual area gradient not vertical: %v %v", top, bottom)
	}
	paint := chartCartesianStyle(p, 0, col)
	p.Series[0].Style.Gradient.From = ui.RGB(0, 200, 0)
	if paint.gradient.From != col {
		t.Fatal("retained paint borrowed mutable app gradient")
	}
	p.Kind = BarChart
	p.Series[0].Style.BarContent = func(_ *ui.Context, d ChartDatum, e ui.Element) {
		bars++
		if d.Value != 100 {
			t.Fatal("custom bar got animation-scaled value")
		}
		e.Border(3, blue)
	}
	r.Frame()
	if bars < 3 {
		t.Fatal("native bar customization not connected")
	}
	p.Kind = LineChart
	p.Cartesian.Curve = ChartCurveNatural
	p.Series[0].Style.Dots = &ChartDotProps{Radius: 10, Content: func(_ *ui.Context, d ChartDatum, e ui.Element) { dots++; e.Background(blue).Label("dot " + d.Label) }}
	p.Series[0].Style.ActiveDot = &ChartDotProps{Disabled: true}
	r.Frame()
	dot := horizontalChartMark(t, r, "dot B")
	horizontalNear(t, dot.W, 21)
	if dots < 3 {
		t.Fatal("custom idle dots did not build")
	}
}

func TestChartCartesianNewCurvesAnimateAndExposeLatestRawData(t *testing.T) {
	selected := 1
	labels := 0
	p := ChartProps{Kind: LineChart, Label: "Animated natural", Height: 240, Selected: &selected, Labels: []string{"A", "B", "C"}, Series: []ChartSeries{{Name: "Value", Values: []float64{10, 30, 20}, Style: ChartSeriesStyle{Dots: &ChartDotProps{Radius: 5}, Labels: []ChartLabelProps{{Content: func(c *ui.Context, d ChartDatum) { labels++; ui.Text(c, "done "+d.Label) }}}}}}, Cartesian: ChartCartesianProps{Curve: ChartCurveNatural, PointScale: true, HideValueAxis: true, HideCategoryAxis: true, HideGrid: true}}
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { Chart(c, p) }) }, 600, 380)
	if labels != 0 || r.HasText("done A") {
		t.Fatal("data labels appeared during path entrance")
	}
	r.Advance(750 * time.Millisecond)
	before := append([]uint8(nil), r.Image().Pix...)
	p.Series[0].Values = []float64{40, 60, 15}
	r.Frame()
	if horizontalChartAccess(t, r, p.Label).Value != "B, Value: 60" || !r.HasText("60") {
		t.Fatal("animation exposed interpolated data to tooltip/AX")
	}
	r.Advance(2 * time.Second)
	if labels < 3 || !r.HasText("done A") {
		t.Fatal("labels did not appear after data animation")
	}
	same := true
	for i, v := range before {
		if v != r.Image().Pix[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("actual curve pixels did not update")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	p.Cartesian.Curve = ChartCurveStep
	p.Series[0].Values = []float64{20, 40, 10}
	r.Frame()
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reduced-motion curve did not settle")
	}
}

func TestChartCartesianMissingNaturalRunsAndNativeDateTickBudget(t *testing.T) {
	points := []chartGeometryDatum{{valid: true, v: [4]float64{0, 0}}, {valid: true, v: [4]float64{.2, .5}}, {}, {valid: true, v: [4]float64{.8, .5}}, {valid: true, v: [4]float64{1, 0}}}
	want := 2 * math.Hypot(100, 150)
	if math.Abs(chartCurveLengthFor(ui.Rect{W: 500, H: 300}, points, ChartCurveNatural)-want) > .001 {
		t.Fatal("missing natural observations were connected")
	}
	p := ChartProps{Kind: LineChart, Label: "Date budget", Labels: make([]string, 90), Series: []ChartSeries{{Values: make([]float64, 90)}}, Cartesian: ChartCartesianProps{PointScale: true, HideValueAxis: true, CategoryMinTickGap: cartesianPtr(float32(32))}}
	for i := range p.Labels {
		p.Labels[i] = time.Date(2024, 4, 1+i, 0, 0, 0, 0, time.UTC).Format("Jan 2")
		p.Series[0].Values[i] = float64(i)
	}
	r := horizontalChartTester(&p)
	count := 0
	last := ""
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleText && a.Name != "" {
			count++
			last = a.Name
		}
	}
	if count < 3 || count > 15 || !r.HasText(p.Labels[89]) {
		t.Fatalf("date ticks not thinned with preserved end: n%d last%s", count, last)
	}
}
