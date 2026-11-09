package display

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestChartInspectionKeyboardSkipsMissingCategories(t *testing.T) {
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			selected := -1
			tt := displayTester(func(c *ui.Context) {
				Chart(c, ChartProps{Kind: kind, Label: "Revenue plot", Labels: []string{"Jan", "Missing", "Mar", "Apr"},
					Series: []ChartSeries{{Name: "Desktop", Values: []float64{30, math.NaN(), 40, 50}}}, Selected: &selected, Legend: true})
			})
			if err := tt.Click("Revenue plot"); err != nil {
				t.Fatal(err)
			}
			tt.Key(0, ui.KeyHome)
			if selected != 0 {
				t.Fatalf("Home selected %d", selected)
			}
			tt.Key(0, ui.KeyRight)
			if selected != 2 {
				t.Fatalf("Right selected missing observation %d", selected)
			}
			info, ok := tt.AccessByRole(ui.RoleGroup, "Revenue plot")
			if !ok || info.Value != "Mar, Desktop: 40" {
				t.Fatalf("accessible inspection: %+v", info)
			}
			if !tt.HasText("Mar") || !tt.HasText("40") {
				t.Fatalf("tooltip absent: %v", tt.Texts())
			}
			tt.Key(0, ui.KeyEnd)
			if selected != 3 {
				t.Fatalf("End selected %d", selected)
			}
			tt.Key(0, ui.KeyEscape)
			if selected != -1 {
				t.Fatalf("Escape selected %d", selected)
			}
		})
	}
}

func TestDonutPointerHitTestingHonorsHoleAndSliceGeometry(t *testing.T) {
	selected := -1
	p := ChartProps{Kind: DonutChart, Label: "Visitors", Labels: []string{"Desktop", "Mobile"},
		Series: []ChartSeries{{Name: "Visitors", Values: []float64{1, 1}}}, Selected: &selected, Height: 240}
	tt := displayTester(func(c *ui.Context) { Chart(c, p) })
	// Hit testing here is for the completed, full donut.
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	info, ok := tt.AccessByRole(ui.RoleGroup, "Visitors")
	if !ok {
		t.Fatal("missing plot")
	}
	r := info.Bounds
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	if selected != -1 {
		t.Fatalf("donut hole selected category %d", selected)
	}
	radius := min(r.W, r.H)/2 - 22
	tt.Move(r.X+r.W/2+radius*.8, r.Y+r.H/2)
	if selected != 0 {
		t.Fatalf("right semicircle selected %d", selected)
	}
	tt.Move(r.X+r.W/2-radius*.8, r.Y+r.H/2)
	if selected != 1 {
		t.Fatalf("left semicircle selected %d", selected)
	}
}

func TestChartStacksUseDivergingAndNormalizedDomains(t *testing.T) {
	p := ChartProps{Labels: []string{"A", "B"}, Series: []ChartSeries{{Values: []float64{3, -2}}, {Values: []float64{4, -3}}}}
	stack := stackSegments(p)
	if stack[0][0].Lo != 0 || stack[1][0].Lo != 3 || stack[1][0].Hi != 7 || stack[1][1].Lo != -5 || stack[1][1].Hi != -2 {
		t.Fatalf("diverging stacks %+v", stack)
	}
	p.Normalized = true
	stack = stackSegments(p)
	if math.Abs(stack[1][0].Hi-1) > 1e-9 || math.Abs(stack[1][1].Lo+1) > 1e-9 {
		t.Fatalf("normalized stacks %+v", stack)
	}
	p.Series[0].Values = []float64{math.Inf(1), math.NaN()}
	stack = stackSegments(p)
	for _, series := range stack {
		for _, span := range series {
			if !finite(span.Lo) || !finite(span.Hi) {
				t.Fatalf("invalid stack %+v", stack)
			}
		}
	}
}

func TestPolarChartsEmptyAndInvalidDataRemainFinite(t *testing.T) {
	for _, kind := range []ChartKind{PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			selected := 99
			tt := displayTester(func(c *ui.Context) {
				Chart(c, ChartProps{Kind: kind, Label: "Empty plot", Labels: []string{"A", "B", "C"}, Selected: &selected,
					Series: []ChartSeries{{Values: []float64{math.NaN(), math.Inf(1), math.Inf(-1)}}}, Legend: true})
			})
			if selected != -1 || !tt.HasText("No data") {
				t.Fatalf("empty chart selected=%d texts=%v", selected, tt.Texts())
			}
			for _, n := range tt.AccessAll() {
				for _, v := range []float32{n.Bounds.X, n.Bounds.Y, n.Bounds.W, n.Bounds.H} {
					if !finite(float64(v)) {
						t.Fatalf("non-finite node %+v", n)
					}
				}
			}
			_ = tt.Image()
		})
	}
}

func TestStackedChartAndCustomTooltipUseOriginalValues(t *testing.T) {
	selected := 1
	seen := ChartTooltipData{}
	tt := displayTester(func(c *ui.Context) {
		Chart(c, ChartProps{Kind: AreaChart, Stacked: true, Normalized: true, Labels: []string{"A", "B"},
			Series:   []ChartSeries{{Name: "Desktop", Values: []float64{10, 20}}, {Name: "Mobile", Values: []float64{5, 15}}},
			Selected: &selected, TooltipContent: func(data ChartTooltipData) { seen = data; ui.Text(c, "Custom values") }})
	})
	if seen.Label != "B" || len(seen.Items) != 2 || seen.Items[0].Value != 20 || seen.Items[1].Value != 15 || !tt.HasText("Custom values") {
		t.Fatalf("tooltip did not receive original data: %+v", seen)
	}
}

func TestChartsPreserveProportionsAtExtremeFiniteMagnitudes(t *testing.T) {
	slices := pieSlices([]float64{math.MaxFloat64, math.MaxFloat64})
	if len(slices) != 2 || math.Abs(slices[0].End-math.Pi) > 1e-9 || math.Abs(slices[1].End-2*math.Pi) > 1e-9 {
		t.Fatalf("overflowed pie fractions=%+v", slices)
	}
	for _, kind := range []ChartKind{BarChart, AreaChart, PieChart, DonutChart, RadarChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			selected := 0
			tt := displayTester(func(c *ui.Context) {
				Chart(c, ChartProps{Kind: kind, Stacked: true, Labels: []string{"A", "B", "C"}, Series: []ChartSeries{{Values: []float64{math.MaxFloat64, 1, math.MaxFloat64}}, {Values: []float64{math.MaxFloat64, 2, math.MaxFloat64}}}, Selected: &selected})
			})
			for _, node := range tt.AccessAll() {
				for _, v := range []float32{node.Bounds.X, node.Bounds.Y, node.Bounds.W, node.Bounds.H} {
					if !finite(float64(v)) {
						t.Fatalf("non-finite chart bound %+v", node)
					}
				}
			}
			_ = tt.Image()
		})
	}
}

func TestLineChartPaintsEverySeries(t *testing.T) {
	tt := displayTester(func(c *ui.Context) {
		Chart(c, ChartProps{Kind: LineChart, Labels: []string{"A", "B", "C"}, Series: []ChartSeries{
			{Name: "First", Values: []float64{5, 5, 5}, Color: ui.RGB(240, 30, 30)},
			{Name: "Second", Values: []float64{2, 2, 2}, Color: ui.RGB(30, 30, 240)},
		}})
	})
	// This assertion counts final series stroke colors, not entry length.
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	image := tt.Image()
	red, blue := 0, 0
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			pixel := image.RGBAAt(x, y)
			if pixel.R > 180 && pixel.G < 80 && pixel.B < 80 {
				red++
			}
			if pixel.B > 180 && pixel.R < 80 && pixel.G < 80 {
				blue++
			}
		}
	}
	if red < 50 || blue < 50 {
		t.Fatalf("paint callback lost a series: red pixels=%d blue pixels=%d", red, blue)
	}
}
