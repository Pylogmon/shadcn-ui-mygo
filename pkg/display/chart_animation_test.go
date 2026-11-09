package display

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func chartMotionDuration(d time.Duration) *time.Duration { return &d }
func chartMotionProps(kind ChartKind) ChartProps {
	red := ui.RGB(210, 30, 40)
	return ChartProps{Kind: kind, Height: 240, Label: "Animated geometry", Labels: []string{"A", "B", "C", "Anchor"}, Series: []ChartSeries{{Name: "Value", Color: red, Values: []float64{20, 80, 40, 100}}}, CategoryColors: []ui.Color{red, red, red, red}, HideTooltip: true, MaxValue: 100,
		Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}, Animation: ChartAnimationProps{Duration: chartMotionDuration(200 * time.Millisecond), Begin: chartMotionDuration(0), Ease: ui.Linear}}
}
func chartMotionTester(p *ChartProps) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(theme.TokensOf(c).Background).Children(func() { Chart(c, *p) })
	}, 600, 340)
}
func chartMotionInk(r *ui.Tester) (int, image.Rectangle) {
	img := r.Image()
	count := 0
	b := image.Rectangle{Min: image.Pt(img.Bounds().Max.X, img.Bounds().Max.Y)}
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			v := img.RGBAAt(x, y)
			if int(v.R)-int(v.G) > 20 && int(v.R)-int(v.B) > 20 {
				count++
				b.Min.X = min(b.Min.X, x)
				b.Min.Y = min(b.Min.Y, y)
				b.Max.X = max(b.Max.X, x+1)
				b.Max.Y = max(b.Max.Y, y+1)
			}
		}
	}
	return count, b
}
func chartMotionSnapshot(r *ui.Tester) []byte { return append([]byte(nil), r.Image().Pix...) }
func chartMotionHasInkNear(r *ui.Tester, x, y float32) bool {
	img := r.Image()
	for py := int(y) - 2; py <= int(y)+2; py++ {
		for px := int(x) - 2; px <= int(x)+2; px++ {
			v := img.RGBAAt(px, py)
			if int(v.R)-int(v.G) > 20 && int(v.R)-int(v.B) > 20 {
				return true
			}
		}
	}
	return false
}

func TestChartAnimationEveryFamilyEntersWithRealGeometry(t *testing.T) {
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			p := chartMotionProps(kind)
			r := chartMotionTester(&p)
			before, _ := chartMotionInk(r)
			if before > 6 {
				t.Fatalf("source entrance already painted %d mark pixels", before)
			}
			r.Advance(100 * time.Millisecond)
			middle, middleBox := chartMotionInk(r)
			midImage := chartMotionSnapshot(r)
			r.Advance(150 * time.Millisecond)
			final, finalBox := chartMotionInk(r)
			if middle <= 0 || middle >= final || bytes.Equal(midImage, r.Image().Pix) {
				t.Fatalf("no genuine geometry middle frame: pixels %d→%d bbox%v→%v", middle, final, middleBox, finalBox)
			}
			// Uniform opacity would keep the same spatial extent. Each entrance must
			// change an actual edge/path endpoint or sector/polygon boundary.
			if middleBox == finalBox {
				t.Fatalf("geometry extent did not move: %v", middleBox)
			}
			if r.FramePending() {
				t.Fatal("completed data animation kept requesting frames")
			}
		})
	}
}

func TestChartAnimationUpdatesAndRapidRetargetUseDisplayedPose(t *testing.T) {
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			p := chartMotionProps(kind)
			r := chartMotionTester(&p)
			r.Advance(250 * time.Millisecond)
			old := chartMotionSnapshot(r)
			p.Series[0].Values[0] = 60
			r.Frame()
			if !bytes.Equal(old, r.Image().Pix) {
				t.Fatal("data update jumped before its first animation step")
			}

			r.Advance(100 * time.Millisecond)
			middle := chartMotionSnapshot(r)
			if bytes.Equal(old, middle) {
				t.Fatal("updated mark/path has no intermediate geometry")
			}
			p.Series[0].Values[0] = 10
			r.Frame()
			if !bytes.Equal(middle, r.Image().Pix) {
				t.Fatal("rapid reversal restarted from an old target instead of displayed geometry")
			}
			r.Advance(250 * time.Millisecond)
			if bytes.Equal(middle, r.Image().Pix) || r.FramePending() {
				t.Fatal("redirected animation did not finish and idle")
			}
		})
	}
}

func TestChartAnimationRedirectsAnUnfinishedEntranceContinuously(t *testing.T) {
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			p := chartMotionProps(kind)
			r := chartMotionTester(&p)
			r.Advance(100 * time.Millisecond)
			middle := chartMotionSnapshot(r)
			p.Series[0].Values[0] = 65
			r.Frame()
			if !bytes.Equal(middle, r.Image().Pix) {
				t.Fatal("redirecting entry reset geometry or its visible reveal")
			}
			r.Advance(250 * time.Millisecond)
			if r.FramePending() {
				t.Fatal("redirected entry failed to settle")
			}
		})
	}
}

func TestChartAnimationBarSignedReversalCollapsesAtZero(t *testing.T) {
	for _, orientation := range []ChartOrientation{ChartVertical, ChartHorizontal} {
		for _, negative := range []bool{false, true} {
			t.Run(fmt.Sprint(orientation, negative), func(t *testing.T) {
				p := chartMotionProps(BarChart)
				p.Orientation = orientation
				p.Labels = []string{"Moving", "Positive", "Negative"}
				p.CategoryColors = nil
				start := 10.0
				if negative {
					start = -10
				}
				p.Series[0].Values = []float64{start, 10, -10}
				r := chartMotionTester(&p)
				r.Advance(250 * time.Millisecond)
				p.Series[0].Values[0] = -start
				r.Frame()
				r.Advance(100 * time.Millisecond)
				r.Frame()
				mark, ok := r.Find(fmt.Sprintf("Value Moving: %g", -start))
				if !ok {
					t.Fatal("missing keyed native bar")
				}
				extent := mark.H
				if orientation == ChartHorizontal {
					extent = mark.W
				}
				if extent > .01 {
					t.Fatalf("signed rectangle must cross zero at half progress, got %v", mark)
				}
				r.Advance(150 * time.Millisecond)
				final, _ := r.Find(fmt.Sprintf("Value Moving: %g", -start))
				if orientation == ChartHorizontal && final.W < 100 || orientation == ChartVertical && final.H < 60 {
					t.Fatalf("reversed bar did not grow on its other side: %v", final)
				}
			})
		}
	}
}

// SVG dash revelation resets at each missing-data subpath. Each run shares
// the same absolute visible-length budget computed from the whole series.
func TestChartAnimationLineMissingRunsShareAbsoluteRevealBudget(t *testing.T) {
	p := chartMotionProps(LineChart)
	p.Labels = []string{"A", "B", "Missing", "D", "E", "F"}
	p.Series[0].Values = []float64{4, 4, math.NaN(), 4, 4, 4}
	p.Animation.Duration = chartMotionDuration(300 * time.Millisecond)
	r := chartMotionTester(&p)
	r.Advance(100 * time.Millisecond)
	r.Frame()
	plot, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	band := chart.NewBand(p.Labels...).Padding(.3, .15)
	x := func(i int) float32 { return plot.Bounds.X + band.At(i)*plot.Bounds.W }
	if !chartMotionHasInkNear(r, x(1), plot.Bounds.Y) {
		t.Fatal("first subpath did not finish at its global length budget")
	}
	if !chartMotionHasInkNear(r, x(4), plot.Bounds.Y) {
		t.Fatal("later subpath did not receive the same absolute reveal budget")
	}
	if chartMotionHasInkNear(r, x(5), plot.Bounds.Y) {
		t.Fatal("later subpath revealed beyond the shared absolute budget")
	}
	r.Advance(250 * time.Millisecond)
	if !chartMotionHasInkNear(r, x(5), plot.Bounds.Y) {
		t.Fatal("final path endpoint was not revealed")
	}
}

func TestChartAnimationAreaEntersByClippingFinalValues(t *testing.T) {
	p := chartMotionProps(AreaChart)
	p.Series[0].Values = []float64{100, 100, 100, 100}
	r := chartMotionTester(&p)
	r.Advance(100 * time.Millisecond)
	plot, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	if !chartMotionHasInkNear(r, plot.Bounds.X+plot.Bounds.W*.25, plot.Bounds.Y+3) {
		t.Fatal("area grew Y values from baseline instead of clipping final curve")
	}
	if chartMotionHasInkNear(r, plot.Bounds.X+plot.Bounds.W*.75, plot.Bounds.Y+3) {
		t.Fatal("area revealed beyond its growing X clip")
	}
}

func TestChartAnimationSourceDefaultsUseFamilyTimingAndEase(t *testing.T) {
	t.Run("bar400CSS-ease", func(t *testing.T) {
		p := chartMotionProps(BarChart)
		p.Animation = ChartAnimationProps{}
		r := chartMotionTester(&p)
		r.Advance(200 * time.Millisecond)
		mid, _ := r.Find("Value Anchor: 100")
		r.Advance(250 * time.Millisecond)
		final, _ := r.Find("Value Anchor: 100")
		ratio := mid.H / final.H
		if ratio < .79 || ratio > .82 {
			t.Fatalf("400ms CSS ease half-time geometry ratio=%v", ratio)
		}
		if r.FramePending() {
			t.Fatal("400ms source bar did not finish")
		}
	})
	for _, kind := range []ChartKind{PieChart, DonutChart} {
		t.Run(fmt.Sprint(kind, "/400+1500"), func(t *testing.T) {
			p := chartMotionProps(kind)
			p.Animation = ChartAnimationProps{}
			r := chartMotionTester(&p)
			r.Advance(350 * time.Millisecond)
			if count, _ := chartMotionInk(r); count != 0 {
				t.Fatal("source pie began before its 400ms delay")
			}
			r.Advance(800 * time.Millisecond)
			middle, _ := chartMotionInk(r)
			r.Advance(650 * time.Millisecond)
			r.Frame()
			if !r.FramePending() {
				t.Fatal("source pie finished before its 1900ms total deadline")
			}
			r.Advance(150 * time.Millisecond)
			final, _ := chartMotionInk(r)
			if middle <= 0 || middle >= final || r.FramePending() {
				t.Fatalf("pie 1500ms geometry after400ms: middle%d final%d pending%v", middle, final, r.FramePending())
			}
		})
	}
}

func TestChartAnimationSemanticsAndNativeEditorUseLatestData(t *testing.T) {
	p := chartMotionProps(LineChart)
	selected := 0
	p.Selected = &selected
	draft := "Initial"
	seen := 0.0
	p.HideTooltip = false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() {
			copy := p
			copy.TooltipContent = func(data ChartTooltipData) {
				if len(data.Items) > 0 {
					seen = data.Items[0].Value
				}
				ui.Text(c, "Latest raw values")
			}
			Chart(c, copy)
			ui.TextInput(c, &draft).Label("Outside native editor")
		})
	}, 600, 380)
	if err := r.Click("Outside native editor"); err != nil {
		t.Fatal(err)
	}
	r.Command("selectAll")
	r.Compose("中", 1)
	p.Series[0].Values[0] = 65
	r.Frame()
	r.Advance(50 * time.Millisecond)
	r.Type("中文 preserved")
	if draft != "中文 preserved" {
		t.Fatalf("geometry rebuild reset native IME/selection: %q", draft)
	}
	plot, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	if seen != 65 || plot.Value != "A, Value: 65" {
		t.Fatalf("animation exposed stale semantic values: tooltip%v AX%q", seen, plot.Value)
	}
}

func chartAnimationRedPixels(r *ui.Tester, b ui.Rect) int {
	image := r.Image()
	n := 0
	for y := int(b.Y); y < int(b.Y+b.H); y++ {
		for x := int(b.X); x < int(b.X+b.W); x++ {
			p := image.RGBAAt(x, y)
			if p.R > 180 && p.G < 80 && p.B < 80 {
				n++
			}
		}
	}
	return n
}

func TestChartAnimationEmptyThenFirstDataUsesSourceEntrance(t *testing.T) {
	for _, kind := range []ChartKind{LineChart, AreaChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			duration := time.Second
			p := ChartProps{Kind: kind, Label: "First data arrival", Height: 200, HideTooltip: true, Animation: ChartAnimationProps{Duration: &duration, Ease: ui.Linear}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}, Series: []ChartSeries{{Name: "Signal", Color: ui.RGB(220, 30, 30)}}}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
			}, 500, 300)
			p.Labels = []string{"A", "B", "C"}
			p.Series[0].Values = []float64{20, 80, 40}
			r.Frame()
			area, _ := r.AccessByRole(ui.RoleGroup, p.Label)
			if got := chartAnimationRedPixels(r, area.Bounds); got != 0 {
				t.Fatalf("first nonempty data painted %d final-path pixels at entrance t0", got)
			}
			r.Advance(250 * time.Millisecond)
			if chartAnimationRedPixels(r, area.Bounds) == 0 {
				t.Fatal("first data never began revealing")
			}
		})
	}
}

func TestChartAnimationLineRedirectionContinuesAbsoluteLengthAndRawInspection(t *testing.T) {
	duration := time.Second
	selected := -1
	p := ChartProps{Kind: LineChart, Label: "Redirected line", Height: 200, HideTooltip: true, Selected: &selected, Labels: []string{"A", "B"}, Animation: ChartAnimationProps{Duration: &duration, Ease: ui.Linear}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}, Series: []ChartSeries{{Name: "Signal", Color: ui.RGB(220, 30, 30), Values: []float64{50, 50}}, {Name: "Anchor", Color: ui.RGB(30, 30, 220), Values: []float64{100, 100}}}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
	}, 500, 300)
	r.Advance(250 * time.Millisecond)
	before := bytes.Clone(r.Image().Pix)
	p.Series[0].Values = []float64{60, 60}
	r.Frame()
	if !bytes.Equal(before, r.Image().Pix) {
		t.Fatal("data redirection reset the already revealed path at the same timestamp")
	}
	if err := r.Click(p.Label); err != nil {
		t.Fatal(err)
	}
	r.Move(1, 1)
	r.Key(0, ui.KeyHome)
	a, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	if a.Value != "A, Signal: 60, Anchor: 100" {
		t.Fatalf("animated inspector announced interpolated/old data: %q", a.Value)
	}
	r.Advance(250 * time.Millisecond)
	image := r.Image()
	furthest := float32(-1)
	for y := int(a.Bounds.Y); y < int(a.Bounds.Y+a.Bounds.H); y++ {
		for x := int(a.Bounds.X); x < int(a.Bounds.X+a.Bounds.W); x++ {
			v := image.RGBAAt(x, y)
			if v.R > 180 && v.G < 80 && v.B < 80 {
				furthest = max(furthest, float32(x))
			}
		}
	}
	// The red horizontal path spans half the plot width. Its old revealed
	// length is .25*.5W; another .25*.5W follows the redirect, so its visible
	// endpoint is x=.25W+.125W+.125W=.5W. A relative restart differs visibly.
	expected := a.Bounds.X + a.Bounds.W*.5
	if math.Abs(float64(furthest-expected)) > 2 {
		t.Fatalf("absolute revealed-length continuation ended at%g expected%g", furthest, expected)
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	a, _ = r.AccessByRole(ui.RoleGroup, p.Label)
	if a.Value != "A, Signal: 60, Anchor: 100" {
		t.Fatal("reduced completion changed the latest semantic values")
	}
	r.Advance(2 * time.Second)
	if r.FramePending() {
		t.Fatal("reduced redirected line kept an animation active")
	}
}

func TestChartAnimationLineFirstDrawablePathAfterMissingDataReveals(t *testing.T) {
	duration := time.Second
	p := ChartProps{Kind: LineChart, Label: "First drawable line", Height: 200, HideTooltip: true, Labels: []string{"A", "B", "C"}, Animation: ChartAnimationProps{Duration: &duration, Ease: ui.Linear}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}, Series: []ChartSeries{{Name: "Signal", Color: ui.RGB(220, 30, 30), Values: []float64{math.NaN(), math.NaN(), math.NaN()}}}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
	}, 500, 300)
	r.Advance(2 * time.Second)
	p.Series[0].Values = []float64{20, 80, 40}
	r.Frame()
	area, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	r.Advance(250 * time.Millisecond)
	middle := chartAnimationRedPixels(r, area.Bounds)
	r.Advance(time.Second)
	final := chartAnimationRedPixels(r, area.Bounds)
	if middle == 0 || middle >= final {
		t.Fatalf("first drawable Line skipped source stroke-length entrance: mid250ms=%d final=%d", middle, final)
	}
}

func TestChartAnimationFirstConnectedLineAfterSingletonUsesTargetPath(t *testing.T) {
	for _, isolated := range []bool{false, true} {
		t.Run(fmt.Sprint(isolated), func(t *testing.T) {
			duration := time.Second
			p := ChartProps{Kind: LineChart, Label: "First connected path", Height: 200, HideTooltip: true, Labels: []string{"A"}, Animation: ChartAnimationProps{Duration: &duration, Ease: ui.Linear}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}, Series: []ChartSeries{{Name: "Signal", Color: ui.RGB(220, 30, 30), Values: []float64{20}}}}
			if isolated {
				p.Labels = []string{"A", "B", "C"}
				p.Series[0].Values = []float64{math.NaN(), 20, math.NaN()}
			}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
			}, 500, 300)
			r.Advance(2 * time.Second)
			p.Labels = []string{"A", "B", "C"}
			p.Series[0].Values = []float64{20, 80, 40}
			r.Frame()
			r.Advance(250 * time.Millisecond)
			area, _ := r.AccessByRole(ui.RoleGroup, p.Label)
			image := r.Image()
			var middle [][2]int
			red := func(x, y int) bool { p := image.RGBAAt(x, y); return p.R > 180 && p.G < 80 && p.B < 80 }
			for y := int(area.Bounds.Y); y < int(area.Bounds.Y+area.Bounds.H); y++ {
				for x := int(area.Bounds.X); x < int(area.Bounds.X+area.Bounds.W); x++ {
					if red(x, y) {
						middle = append(middle, [2]int{x, y})
					}
				}
			}
			r.Advance(time.Second)
			image = r.Image()
			stray := 0
			for _, p := range middle {
				near := false
				for dy := -2; dy <= 2; dy++ {
					for dx := -2; dx <= 2; dx++ {
						if red(p[0]+dx, p[1]+dy) {
							near = true
						}
					}
				}
				if !near {
					stray++
				}
			}
			if len(middle) == 0 || stray > max(3, len(middle)/100) {
				t.Fatalf("first connected Line morphed from a zero-length predecessor instead of revealing final curve: midpixels%d off-target%d", len(middle), stray)
			}
		})
	}
}

func chartAnimationTestPoint(x, y float64) chartGeometryDatum {
	return chartGeometryDatum{v: [4]float64{x, y}, valid: true}
}
func TestChartAnimationLengthKnownLengthsAndMissingRuns(t *testing.T) {
	p := chartAnimationTestPoint
	tests := []struct {
		name   string
		points []chartGeometryDatum
		r      ui.Rect
		want   float64
	}{{"empty", nil, ui.Rect{W: 300, H: 400}, 0}, {"singleton", []chartGeometryDatum{p(.4, .3)}, ui.Rect{W: 300, H: 400}, 0}, {"pythagoras", []chartGeometryDatum{p(0, 0), p(1, 1)}, ui.Rect{W: 300, H: 400}, 500}, {"collinear-monotone", []chartGeometryDatum{p(0, 0), p(.2, .2), p(1, 1)}, ui.Rect{W: 300, H: 400}, 500}, {"flat", []chartGeometryDatum{p(0, .3), p(.3, .3), p(.7, .3), p(1, .3)}, ui.Rect{W: 300, H: 400}, 300}, {"vertical", []chartGeometryDatum{p(.2, .1), p(.2, .9)}, ui.Rect{W: 300, H: 400}, 320}, {"missing-runs", []chartGeometryDatum{p(0, 0), p(.2, .2), {}, p(.3, .8), {}, p(.5, 0), p(1, 1)}, ui.Rect{W: 300, H: 400}, 100 + math.Hypot(150, 400)}, {"nonuniform-resize", []chartGeometryDatum{p(0, 0), p(1, 1)}, ui.Rect{X: 30, Y: 120, W: 900, H: 120}, math.Hypot(900, 120)}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := chartCurveLength(test.r, test.points); math.Abs(got-test.want) > 1e-6 {
				t.Fatalf("independent known length %g expected %g", got, test.want)
			}
		})
	}
}

// For this symmetric peak, the fixed monotone contract has a first cubic
// derivative (100, -300+300t²). Use an independent Simpson quadrature of that
// analytic speed, not the helper's control-polygon subdivision algorithm.
func TestChartAnimationLengthCurvedPeakIndependentIntegral(t *testing.T) {
	const steps = 65536
	speed := func(t float64) float64 { return math.Hypot(100, -300+300*t*t) }
	sum := speed(0) + speed(1)
	for i := 1; i < steps; i++ {
		weight := float64(2)
		if i%2 == 1 {
			weight = 4
		}
		sum += weight * speed(float64(i)/steps)
	}
	want := 2 * sum / (3 * steps)
	got := chartCurveLength(ui.Rect{W: 200, H: 200}, []chartGeometryDatum{chartAnimationTestPoint(0, 0), chartAnimationTestPoint(.5, 1), chartAnimationTestPoint(1, 0)})
	if math.Abs(got-want) > .001 {
		t.Fatalf("curved absolute length got%.9f analytic integral%.9f", got, want)
	}
}

func TestChartAnimationLengthNativeTrimStopsAtKnownAnchorAfterResize(t *testing.T) {
	for _, missing := range []bool{false, true} {
		for _, size := range [][2]int{{460, 320}, {1000, 160}, {200, 700}} {
			t.Run(fmt.Sprintf("missing%v/%dx%d", missing, size[0], size[1]), func(t *testing.T) {
				p := chartAnimationTestPoint
				points := []chartGeometryDatum{p(0, .1), p(.25, .1), p(.5, .9), p(1, .3)}
				if missing {
					points = []chartGeometryDatum{p(0, .1), p(.25, .1), {}, p(.55, .4), p(.8, .8), p(1, .4)}
				}
				var box ui.Rect
				r := ui.NewTester(func(c *ui.Context) {
					ui.Box(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() {
						ui.Box(c).Fill().Draw(func(painter *ui.Painter, b ui.Rect) {
							box = b
							xs, ys := make([]float32, len(points)), make([]float32, len(points))
							for i, point := range points {
								xs[i], ys[i] = float32(math.NaN()), float32(math.NaN())
								if point.valid {
									xs[i], ys[i] = float32(point.v[0]), float32(point.v[1])
								}
							}
							total := chartCurveLength(b, points)
							fraction := float32(float64(b.W) * .25 / total)
							painter.StrokePath(chart.Line(b, xs, ys, chart.CurveMonotone).Trim(0, fraction).Cap(ui.ButtCap), 2, ui.RGB(220, 30, 30))
						})
					})
				}, size[0], size[1])
				check := func() {
					image := r.Image()
					endX, endY := box.X+box.W*.25, box.Y+box.H*.9
					furthest := float32(-1)
					near := 0
					for y := 0; y < image.Bounds().Dy(); y++ {
						for x := 0; x < image.Bounds().Dx(); x++ {
							pixel := image.RGBAAt(x, y)
							if pixel.R > 180 && pixel.G < 80 && pixel.B < 80 {
								furthest = max(furthest, float32(x))
								if math.Hypot(float64(float32(x)-endX), float64(float32(y)-endY)) < 3 {
									near++
								}
							}
						}
					}
					if near == 0 || math.Abs(float64(furthest-endX)) > 2 {
						t.Fatalf("public native Trim endpoint disagrees with absolute length: rect%v expectedend(%g,%g) lastx%g near%d", box, endX, endY, furthest, near)
					}
				}
				check()
				r.SetSize(size[0]+123, size[1]+57)
				check()
			})
		}
	}
}

func chartMotionCrop(r *ui.Tester, b ui.Rect) []byte {
	img := r.Image()
	rect := image.Rect(int(b.X), int(b.Y), int(b.X+b.W), int(b.Y+b.H)).Intersect(img.Bounds())
	out := []byte{}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		from := img.PixOffset(rect.Min.X, y)
		out = append(out, img.Pix[from:from+rect.Dx()*4]...)
	}
	return out
}
func chartMotionPlot(t *testing.T, r *ui.Tester, p ChartProps) ui.Rect {
	t.Helper()
	a, ok := r.AccessByRole(ui.RoleGroup, p.Label)
	if !ok {
		t.Fatal("missing animated plot")
	}
	return a.Bounds
}

func TestChartAnimationReducedInstantAndDisabledFinishWithoutReplay(t *testing.T) {
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		for _, mode := range []string{"reduced", "instant", "disabled"} {
			t.Run(fmt.Sprint(kind, "/", mode), func(t *testing.T) {
				p := chartMotionProps(kind)
				instant := false
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					if instant {
						c.DisableAnimations()
					}
					ui.Column(c).Fill().Padding(20).Background(theme.TokensOf(c).Background).Children(func() { Chart(c, p) })
				}, 600, 340)
				r.Advance(100 * time.Millisecond)
				switch mode {
				case "reduced":
					r.SetPreferences(ui.Preferences{ReduceMotion: true})
				case "instant":
					instant = true
					r.Frame()
				case "disabled":
					p.Animation.Disabled = true
					r.Frame()
				}
				referenceProps := p
				referenceProps.Animation.Disabled = true
				reference := chartMotionTester(&referenceProps)
				if !bytes.Equal(r.Image().Pix, reference.Image().Pix) {
					t.Fatal("stopping motion did not draw the latest final geometry")
				}
				before := chartMotionSnapshot(r)
				r.Advance(time.Second)
				if r.FramePending() || !bytes.Equal(before, r.Image().Pix) {
					t.Fatal("stopped chart did not idle")
				}
				switch mode {
				case "reduced":
					r.SetPreferences(ui.Preferences{})
				case "instant":
					instant = false
					r.Frame()
				case "disabled":
					p.Animation.Disabled = false
					r.Frame()
				}
				if r.FramePending() || !bytes.Equal(before, r.Image().Pix) {
					t.Fatal("enabling motion replayed unchanged data")
				}
			})
		}
	}
}

func TestChartAnimationHiddenAndOffscreenFinishAndResumeIdle(t *testing.T) {
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		for _, mode := range []string{"window", "offscreen", "ancestor"} {
			t.Run(fmt.Sprint(kind, "/", mode), func(t *testing.T) {
				p := chartMotionProps(kind)
				offset := float32(0)
				ancestorHidden := false
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					ui.Scroll(c).Fill().Background(theme.TokensOf(c).Background).Children(func() {
						ui.Column(c).FillWidth().Padding(20).Children(func() {
							ui.Box(c).Height(offset).Shrink(0)
							Chart(c, p).Shrink(0).Hidden(ancestorHidden)
							ui.Box(c).Height(400).Shrink(0)
						})
					})
				}, 600, 340)
				r.Advance(100 * time.Millisecond)
				if mode == "window" {
					r.SetWindowHidden(true)
				} else if mode == "ancestor" {
					ancestorHidden = true
					r.Frame()
				} else {
					offset = 600
					r.Frame()
				}
				r.Advance(time.Second)
				if r.FramePending() {
					t.Fatal("hidden/offscreen chart kept requesting frames")
				}
				if mode == "window" {
					r.SetWindowHidden(false)
				} else if mode == "ancestor" {
					ancestorHidden = false
					r.Frame()
				} else {
					offset = 0
					r.Frame()
				}
				referenceProps := p
				referenceProps.Animation.Disabled = true
				reference := chartMotionTester(&referenceProps)
				if !bytes.Equal(chartMotionCrop(r, chartMotionPlot(t, r, p)), chartMotionCrop(reference, chartMotionPlot(t, reference, p))) {
					t.Fatal("restored chart replayed or retained an unfinished pose")
				}
				r.Advance(time.Second)
				if r.FramePending() {
					t.Fatal("restored unchanged chart failed to idle")
				}
			})
		}
	}
}

func TestChartAnimationFirstOffscreenLayoutAndUnmountedDelayAreIdle(t *testing.T) {
	for _, offscreen := range []bool{false, true} {
		t.Run(fmt.Sprint(offscreen), func(t *testing.T) {
			p := chartMotionProps(PieChart)
			p.Animation.Begin = nil // Source 400ms delay.
			show := true
			offset := float32(0)
			if offscreen {
				offset = 600
			}
			builds := 0
			draft := "Remaining editor"
			r := ui.NewTester(func(c *ui.Context) {
				builds++
				theme.Apply(c)
				ui.Scroll(c).Fill().Children(func() {
					ui.Column(c).FillWidth().Children(func() {
						ui.Box(c).Height(offset).Shrink(0)
						if show {
							Chart(c, p).Shrink(0)
						}
						ui.TextInput(c, &draft).Label("Remaining editor")
						ui.Box(c).Height(400).Shrink(0)
					})
				})
			}, 600, 340)
			if offscreen {
				before := builds
				r.Advance(time.Second)
				if r.FramePending() || builds != before {
					t.Fatalf("initial offscreen probe hot-looped: builds%d→%d", before, builds)
				}
			}
			show = false
			offset = 0
			r.Frame()
			before := builds
			r.Advance(2 * time.Second)
			if r.FramePending() || builds != before {
				t.Fatalf("unmounted delayed chart kept scheduled work: builds%d→%d", before, builds)
			}
			if err := r.Click("Remaining editor"); err != nil {
				t.Fatal(err)
			}
			r.Command("selectAll")
			r.Type("Still native")
			if draft != "Still native" {
				t.Fatal("unmounting animation disturbed the remaining native editor")
			}
		})
	}
}

func TestChartAnimationChangingSettingsAndInvalidEaseKeepFinitePose(t *testing.T) {
	p := chartMotionProps(BarChart)
	r := chartMotionTester(&p)
	r.Advance(100 * time.Millisecond)
	before := chartMotionSnapshot(r)
	p.Animation.Ease = ui.EaseIn
	r.Frame()
	if !bytes.Equal(before, r.Image().Pix) {
		t.Fatal("changing easing jumped the displayed pose")
	}
	p.Animation.Duration = chartMotionDuration(400 * time.Millisecond)
	r.Frame()
	if !bytes.Equal(before, r.Image().Pix) {
		t.Fatal("changing duration jumped the displayed pose")
	}
	p.Animation.Ease = func(float32) float32 { return float32(math.NaN()) }
	r.Frame()
	r.Advance(500 * time.Millisecond)
	if r.FramePending() {
		t.Fatal("non-finite custom easing did not safely finish")
	}
	for _, duration := range []time.Duration{0, -time.Millisecond} {
		p := chartMotionProps(DonutChart)
		p.Animation.Duration = &duration
		r := chartMotionTester(&p)
		if n, _ := chartMotionInk(r); n < 100 || r.FramePending() {
			t.Fatal("explicit nonpositive duration did not finish immediately")
		}
	}
}

func TestChartAnimationMissingExtremeLengthKindAndStackChangesStayFinite(t *testing.T) {
	p := chartMotionProps(BarChart)
	r := chartMotionTester(&p)
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		p.Kind = kind
		p.Orientation = ChartHorizontal
		p.Stacked = true
		p.Normalized = true
		p.Labels = []string{"A", "Missing", "C", "Anchor", "Added"}
		p.Series = []ChartSeries{{Name: "S1", Color: ui.RGB(210, 30, 40), Values: []float64{math.MaxFloat64, math.NaN(), -math.MaxFloat64, math.MaxFloat64}}, {Name: "S2", Color: ui.RGB(20, 80, 200), Values: []float64{math.MaxFloat64, math.Inf(1), 3}}}
		r.Frame()
		r.Advance(70 * time.Millisecond)
		p.Labels = p.Labels[:3]
		p.Series = p.Series[:1]
		p.Series[0].Values = []float64{1e-200, math.Inf(-1), 2e-200}
		p.Stacked = false
		p.Normalized = false
		p.Orientation = ChartVertical
		r.Frame()
		r.SetDirection(ui.RightToLeft)
		r.Advance(250 * time.Millisecond)
		for _, node := range r.AccessAll() {
			for _, v := range []float32{node.Bounds.X, node.Bounds.Y, node.Bounds.W, node.Bounds.H} {
				if !finite(float64(v)) {
					t.Fatalf("kind%d nonfinite native geometry: %+v", kind, node)
				}
			}
		}
		if r.FramePending() {
			t.Fatalf("kind%d failed to idle after changing topology", kind)
		}
	}
	p.Series[0].Values = []float64{math.NaN(), math.Inf(1), math.Inf(-1)}
	r.Frame()
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("all-missing geometry kept an animation alive")
	}
}

func TestChartAnimationCustomEaseOnlyRunsDuringLiveBuild(t *testing.T) {
	live := false
	show := true
	calls := 0
	p := chartMotionProps(LineChart)
	p.Animation.Ease = func(t float32) float32 {
		if !live {
			panic("application easing retained outside build")
		}
		calls++
		return t
	}
	r := ui.NewTester(func(c *ui.Context) {
		live = true
		defer func() { live = false }()
		theme.Apply(c)
		if show {
			Chart(c, p)
		}
	}, 600, 340)
	r.Advance(100 * time.Millisecond)
	if calls == 0 {
		t.Fatal("custom easing had no actual geometry effect")
	}
	show = false
	r.Frame()
	before := calls
	r.Advance(time.Second)
	if calls != before || r.FramePending() {
		t.Fatal("unmounted chart retained application easing work")
	}
}

func TestChartAnimationPolarPositiveToZeroShrinksFromDisplayedArc(t *testing.T) {
	for _, kind := range []ChartKind{PieChart, DonutChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			duration, begin := time.Second, time.Duration(0)
			p := ChartProps{Kind: kind, Label: "Shrinking positive sector", Height: 240, HideTooltip: true, Labels: []string{"A", "B"}, CategoryColors: []ui.Color{ui.RGB(220, 30, 30), ui.RGB(30, 30, 220)}, Animation: ChartAnimationProps{Duration: &duration, Begin: &begin, Ease: ui.Linear}, Series: []ChartSeries{{Name: "Signal", Values: []float64{1, 3}}}}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
			}, 500, 340)
			r.Advance(1100 * time.Millisecond)
			a, _ := r.AccessByRole(ui.RoleGroup, p.Label)
			before := bytes.Clone(r.Image().Pix)
			beforeRed := chartAnimationRedPixels(r, a.Bounds)
			p.Series[0].Values[0] = 0
			r.Frame()
			afterRed := chartAnimationRedPixels(r, a.Bounds)
			if !bytes.Equal(before, r.Image().Pix) || beforeRed == 0 || afterRed != beforeRed {
				t.Fatalf("positive-to-zero arc vanished before interpolation: red pixels%d ->%d", beforeRed, afterRed)
			}
			r.Advance(500 * time.Millisecond)
			middle := chartAnimationRedPixels(r, a.Bounds)
			if middle <= 0 || middle >= beforeRed {
				t.Fatalf("zero-target sector did not geometrically shrink: before%d middle%d", beforeRed, middle)
			}
			r.Advance(600 * time.Millisecond)
			if final := chartAnimationRedPixels(r, a.Bounds); final != 0 {
				t.Fatalf("zero-target sector did not finish disappearing: %d", final)
			}
		})
	}
}

func TestChartAnimationDefaultLongFamiliesRunFor1500Milliseconds(t *testing.T) {
	for _, kind := range []ChartKind{LineChart, AreaChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			p := chartMotionProps(kind)
			p.Animation = ChartAnimationProps{}
			r := chartMotionTester(&p)
			r.Advance(750 * time.Millisecond)
			middle, _ := chartMotionInk(r)
			r.Advance(250 * time.Millisecond)
			r.Frame()
			if !r.FramePending() {
				t.Fatal("source 1500ms family finished at or before 1000ms")
			}
			r.Advance(550 * time.Millisecond)
			final, _ := chartMotionInk(r)
			if middle <= 0 || middle >= final || r.FramePending() {
				t.Fatalf("source long-family timing: middle%d final%d pending%v", middle, final, r.FramePending())
			}
		})
	}
}

func TestChartAnimationActiveLineAndAreaDotsUseTargetCoordinates(t *testing.T) {
	for _, kind := range []ChartKind{LineChart, AreaChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			p := chartMotionProps(kind)
			selected := 0
			p.Selected = &selected
			r := chartMotionTester(&p)
			r.Advance(250 * time.Millisecond)
			p.Series[0].Values[0] = 60
			r.Frame()
			b := chartMotionPlot(t, r, p)
			x := chart.NewBand(p.Labels...).Padding(.3, .15).At(0)
			if !chartMotionHasInkNear(r, b.X+x*b.W, b.Y+.4*b.H) {
				t.Fatal("source active dot remained on the animated old point instead of latest target")
			}
		})
	}
}

func TestChartAnimationPolarNoDataResetsAndStackedRadialZeroShrinks(t *testing.T) {
	for _, kind := range []ChartKind{PieChart, DonutChart} {
		t.Run(fmt.Sprint(kind, "/no-data"), func(t *testing.T) {
			p := chartMotionProps(kind)
			r := chartMotionTester(&p)
			r.Advance(250 * time.Millisecond)
			p.Series[0].Values = []float64{0, 0, 0, 0}
			r.Frame()
			if count, _ := chartMotionInk(r); count != 0 || !r.HasText("No data") || r.FramePending() {
				t.Fatal("source all-zero pie failed to reset its previous sectors")
			}
			p.Series[0].Values = []float64{20, 80, 40, 100}
			r.Frame()
			if count, _ := chartMotionInk(r); count != 0 {
				t.Fatal("new data after reset reused old sector geometry")
			}
			r.Advance(100 * time.Millisecond)
			if count, _ := chartMotionInk(r); count == 0 {
				t.Fatal("new pie did not re-enter")
			}
		})
	}
	p := chartMotionProps(RadialChart)
	p.Stacked = true
	p.Series = append(p.Series, ChartSeries{Name: "Other", Color: ui.RGB(30, 30, 220), Values: []float64{20, 20, 20, 20}})
	r := chartMotionTester(&p)
	r.Advance(250 * time.Millisecond)
	area := chartMotionPlot(t, r, p)
	old := chartAnimationRedPixels(r, area)
	p.Series[0].Values = []float64{0, 0, 0, 0}
	r.Frame()
	if got := chartAnimationRedPixels(r, area); got != old {
		t.Fatal("zero-target stacked radial reset its current sector")
	}
	r.Advance(100 * time.Millisecond)
	mid := chartAnimationRedPixels(r, area)
	r.Advance(150 * time.Millisecond)
	final := chartAnimationRedPixels(r, area)
	if mid <= 0 || mid >= old || final != 0 || r.FramePending() {
		t.Fatalf("stacked radial zero interpolation old%d mid%d final%d", old, mid, final)
	}
}

func TestChartAnimationPiePointerFollowsCurrentColorAndLatestRaw(t *testing.T) {
	for _, kind := range []ChartKind{PieChart, DonutChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			duration, begin := time.Second, time.Duration(0)
			selected := -1
			p := ChartProps{Kind: kind, Label: "Animated item hit", Height: 240, HideTooltip: true, Selected: &selected, Labels: []string{"A", "B"}, CategoryColors: []ui.Color{ui.RGB(220, 30, 30), ui.RGB(30, 30, 220)}, Animation: ChartAnimationProps{Duration: &duration, Begin: &begin, Ease: ui.Linear}, Series: []ChartSeries{{Name: "Signal", Values: []float64{1, 3}}}}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
			}, 500, 340)
			r.Advance(1100 * time.Millisecond)
			p.Series[0].Values = []float64{3, 1}
			r.Frame()
			r.Advance(250 * time.Millisecond)
			a, _ := r.AccessByRole(ui.RoleGroup, p.Label)
			x, y := a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2+(a.Bounds.W/2-22)*.8
			pixel := r.Image().RGBAAt(int(x), int(y))
			if pixel.B < 180 || pixel.R > 80 {
				t.Fatalf("probe point is not visibly current blue B: %v", pixel)
			}
			r.Press(x, y)
			r.Release(x, y)
			a, _ = r.AccessByRole(ui.RoleGroup, p.Label)
			if selected != 1 || a.Value != "B, Signal: 1" {
				t.Fatalf("visible blue B was hit as future target A: selected%d value%q", selected, a.Value)
			}
		})
	}
}

func TestChartAnimationShrinkingZeroItemPointerAndKeyboardHaveSeparateEligibility(t *testing.T) {
	for _, kind := range []ChartKind{PieChart, DonutChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			duration, begin := time.Second, time.Duration(0)
			selected := -1
			p := ChartProps{Kind: kind, Label: "Shrinking zero item", Height: 240, HideTooltip: true, Selected: &selected, Labels: []string{"A", "B"}, CategoryColors: []ui.Color{ui.RGB(220, 30, 30), ui.RGB(30, 30, 220)}, Animation: ChartAnimationProps{Duration: &duration, Begin: &begin, Ease: ui.Linear}, Series: []ChartSeries{{Name: "Signal", Values: []float64{1, 3}}}}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
			}, 500, 340)
			r.Advance(1100 * time.Millisecond)
			p.Series[0].Values[0] = 0
			r.Frame()
			r.Advance(250 * time.Millisecond)
			a, _ := r.AccessByRole(ui.RoleGroup, p.Label)
			radius := a.Bounds.W/2 - 22
			x, y := a.Bounds.X+a.Bounds.W/2+radius*.4, a.Bounds.Y+a.Bounds.H/2-radius*.8*float32(math.Cos(math.Pi/6))
			pixel := r.Image().RGBAAt(int(x), int(y))
			if pixel.R < 180 || pixel.B > 80 {
				t.Fatalf("probe point is not visibly shrinking red A: %v", pixel)
			}
			r.Press(x, y)
			r.Release(x, y)
			a, _ = r.AccessByRole(ui.RoleGroup, p.Label)
			if selected != 0 || a.Value != "A, Signal: 0" {
				t.Fatalf("visible zero-target A lost current SVG item semantics: selected%d value%q", selected, a.Value)
			}
			r.Key(0, ui.KeyHome)
			r.Advance(50 * time.Millisecond)
			if selected != 1 {
				t.Fatal("resting pointer overrode keyboard's target-positive category contract")
			}
			r.Move(x+2, y)
			a, _ = r.AccessByRole(ui.RoleGroup, p.Label)
			if selected != 0 || a.Value != "A, Signal: 0" {
				t.Fatal("real pointer motion did not reacquire visible shrinking zero item")
			}
			r.Advance(800 * time.Millisecond)
			a, _ = r.AccessByRole(ui.RoleGroup, p.Label)
			if selected != 1 || a.Value != "B, Signal: 3" {
				t.Fatalf("stationary pointer kept a vanished zero sector instead of the now-visible B: %d %q", selected, a.Value)
			}
		})
	}
}

func TestChartAnimationStackedBarSignedReversalKeepsSourceOrigin(t *testing.T) {
	for _, orientation := range []ChartOrientation{ChartVertical, ChartHorizontal} {
		for _, negative := range []bool{false, true} {
			t.Run(fmt.Sprint(orientation, negative), func(t *testing.T) {
				p := chartMotionProps(BarChart)
				p.Orientation = orientation
				p.Stacked = true
				p.Labels = []string{"Moving", "Positive", "Negative"}
				start := 10.0
				if negative {
					start = -10
				}
				p.Series[0].Values = []float64{start, 20, -20}
				p.Series = append(p.Series, ChartSeries{Name: "Stack", Values: []float64{5, 0, 0}, Color: ui.RGB(30, 30, 220)})
				r := chartMotionTester(&p)
				r.Advance(250 * time.Millisecond)
				p.Series[0].Values[0] = -start
				r.Frame()
				r.Advance(100 * time.Millisecond)
				r.Frame()
				mark, _ := r.Find(fmt.Sprintf("Value Moving: %g", -start))
				extent := mark.H
				if orientation == ChartHorizontal {
					extent = mark.W
				}
				if extent > .01 {
					t.Fatalf("stacked signed rectangle crossed without collapsing: %v", mark)
				}
				r.Advance(150 * time.Millisecond)
				if r.FramePending() {
					t.Fatal("stacked reversal did not idle")
				}
			})
		}
	}
}

func TestChartAnimationScopedPaletteChangesDuringMotionUseLiveBuildColors(t *testing.T) {
	for _, kind := range []ChartKind{LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			p := chartMotionProps(kind)
			p.Series[0].Color = ui.Color{}
			p.CategoryColors = nil
			p.Legend = true
			brand := "blue"
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				tokens := theme.Preset(brand, false)
				theme.Scope(c, tokens).Fill().Padding(20).Background(tokens.Background).Children(func() { Chart(c, p) })
			}, 600, 380)
			r.Advance(100 * time.Millisecond)
			name := "A"
			if kind == LineChart || kind == AreaChart || kind == RadarChart {
				name = "Value"
			}
			before := polarThemeLegendColor(t, r, name)
			brand = "rose"
			r.Frame()
			after := polarThemeLegendColor(t, r, name)
			if before == after {
				t.Fatal("probe did not change the actual scoped palette")
			}
			plot := chartMotionPlot(t, r, p)
			if n := polarThemePixelCount(r, plot, after); n < 2 {
				t.Fatalf("in-flight mark/path reread window theme after scope exit: %v, pixels%d", after, n)
			}
			r.Advance(150 * time.Millisecond)
			if r.FramePending() {
				t.Fatal("color-only scope change restarted the geometry timeline")
			}
		})
	}
}

func TestChartAnimationMissingRunsShareAbsoluteRevealInParallel(t *testing.T) {
	duration := time.Second
	p := ChartProps{Kind: LineChart, Label: "Parallel missing runs", Height: 200, HideTooltip: true, Labels: []string{"A", "B", "C", "gap", "E", "F"}, Animation: ChartAnimationProps{Duration: &duration, Ease: ui.Linear}, Cartesian: ChartCartesianProps{HideValueAxis: true, HideCategoryAxis: true, HideGrid: true, HideCursor: true}, Series: []ChartSeries{{Name: "Signal", Color: ui.RGB(220, 30, 30), Values: []float64{50, 50, 50, math.NaN(), 50, 50}}, {Name: "Anchor", Color: ui.RGB(30, 30, 220), Values: []float64{100, 100, 100, 100, 100, 100}}}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).Children(func() { Chart(c, p) })
	}, 620, 300)
	r.Advance(250 * time.Millisecond)
	a, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	image := r.Image()
	last := [2]float32{-1, -1}
	y := int(a.Bounds.Y + a.Bounds.H*.5)
	for x := int(a.Bounds.X); x < int(a.Bounds.X+a.Bounds.W); x++ {
		v := image.RGBAAt(x, y)
		if v.R > 180 && v.G < 80 && v.B < 80 {
			run := 0
			if float32(x) > a.Bounds.X+a.Bounds.W*.6 {
				run = 1
			}
			last[run] = float32(x)
		}
	}
	// The first flat run spans W/3, the second W/6; summed length W/2.
	// At progress .25 both receive the same absolute budget W/8, rather than
	// sequential allocation or each receiving .25 of its own length.
	for run, start := range []float32{1.0 / 12, .75} {
		want := a.Bounds.X + a.Bounds.W*(start+.125)
		if math.Abs(float64(last[run]-want)) > 2 {
			t.Fatalf("SVG reset-per-subpath shared absolute reveal failed for run%d: endpoint%g expected%g rect%v", run, last[run], want, a.Bounds)
		}
	}
}
