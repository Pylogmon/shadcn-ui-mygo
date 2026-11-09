package display

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestChartUnstackedPolarOmitsUndrawnSeriesFromTooltipAndGeometry(t *testing.T) {
	for _, kind := range []ChartKind{PieChart, DonutChart, RadialChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			selected := -1
			extra := false
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				theme.Scope(c, theme.Preset("blue", false)).Fill().Padding(20).Children(func() {
					p := ChartProps{Kind: kind, Label: "Visible polar observations", Height: 240, Legend: true, Selected: &selected, Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Visible", Values: []float64{10, 20}}}}
					if extra {
						p.Series = append(p.Series, ChartSeries{Name: "Undrawn", Values: []float64{math.MaxFloat64, math.MaxFloat64}})
					}
					Chart(c, p)
				})
			}, 460, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if err := r.Click("Visible polar observations"); err != nil {
				t.Fatal(err)
			}
			r.Move(1, 1)
			r.Key(0, ui.KeyHome)
			before := bytes.Clone(r.Image().Pix)
			extra = true
			r.Frame()
			a, _ := r.AccessByRole(ui.RoleGroup, "Visible polar observations")
			if a.Value != "A, Visible: 10" {
				t.Fatalf("inspection exposed an undrawn series: %q", a.Value)
			}
			if row, ok := r.Find("Undrawn"); ok && row.W > 0 && row.H > 0 {
				t.Fatalf("native tooltip painted an undrawn row: %v", row)
			}
			if !bytes.Equal(before, r.Image().Pix) {
				t.Fatal("unpainted series changed native geometry, legend or tooltip")
			}
			r.Key(0, ui.KeyEnd)
			a, _ = r.AccessByRole(ui.RoleGroup, "Visible polar observations")
			if a.Value != "B, Visible: 20" {
				t.Fatalf("End inspected an undrawn series: %q", a.Value)
			}
			before = bytes.Clone(r.Image().Pix)
			extra = false
			r.Frame()
			if !bytes.Equal(before, r.Image().Pix) {
				t.Fatal("removing an unpainted series changed the native chart")
			}
		})
	}
}

func TestChartUnstackedRadialDomainIgnoresUndrawnHugeSeries(t *testing.T) {
	extra := false
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		theme.Scope(c, theme.Preset("blue", false)).Fill().Padding(20).Children(func() {
			p := ChartProps{Kind: RadialChart, Label: "Radial visible domain", Height: 240, HideTooltip: true, Labels: []string{"A", "B"}, Series: []ChartSeries{{Name: "Visible", Values: []float64{10, 20}}}}
			if extra {
				p.Series = append(p.Series, ChartSeries{Name: "Undrawn", Values: []float64{math.MaxFloat64, math.MaxFloat64}})
			}
			Chart(c, p)
		})
	}, 460, 360, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	a, _ := r.AccessByRole(ui.RoleGroup, "Radial visible domain")
	x, y := a.Bounds.X+a.Bounds.W/2+(a.Bounds.W/2-22)*.84, a.Bounds.Y+a.Bounds.H/2
	before := r.Image().RGBAAt(int(x), int(y))
	extra = true
	r.Frame()
	after := r.Image().RGBAAt(int(x), int(y))
	if before != after {
		t.Fatalf("undrawn huge values changed visible first-series native arc: %v -> %v", before, after)
	}
	r.Move(x, y)
	a, _ = r.AccessByRole(ui.RoleGroup, "Radial visible domain")
	if a.Value != "A, Visible: 10" {
		t.Fatalf("native hit/inspection disagreed with painted arc: %q", a.Value)
	}
}

func TestChartStackedPolarAndRadarKeepEveryPaintedSeries(t *testing.T) {
	for _, kind := range []ChartKind{PieChart, DonutChart, RadialChart, RadarChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			selected := -1
			extra := true
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				theme.Scope(c, theme.Preset("blue", false)).Fill().Padding(20).Children(func() {
					p := ChartProps{Kind: kind, Stacked: kind != RadarChart, Label: "All painted observations", Height: 240, Selected: &selected, Labels: []string{"A", "B", "C", "D"}, Series: []ChartSeries{{Name: "Visible", Values: []float64{10, 20, 10, 20}}}}
					if extra {
						p.Series = append(p.Series, ChartSeries{Name: "Second painted", Values: []float64{5, 4, 5, 4}})
					}
					Chart(c, p)
				})
			}, 460, 380, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if err := r.Click("All painted observations"); err != nil {
				t.Fatal(err)
			}
			r.Move(1, 1)
			r.Key(0, ui.KeyHome)
			a, _ := r.AccessByRole(ui.RoleGroup, "All painted observations")
			if !strings.Contains(a.Value, "Second painted: 5") {
				t.Fatalf("painted second series missing from native inspection: %q", a.Value)
			}
			if row, ok := r.Find("Second painted"); !ok || row.W <= 0 || row.H <= 0 {
				t.Fatal("painted second series missing from actual tooltip")
			}
			beforeImage := r.Image()
			var lower []byte
			for y := int(a.Bounds.Y + a.Bounds.H/2); y < int(a.Bounds.Y+a.Bounds.H); y++ {
				for x := int(a.Bounds.X); x < int(a.Bounds.X+a.Bounds.W); x++ {
					v := beforeImage.RGBAAt(x, y)
					lower = append(lower, v.R, v.G, v.B, v.A)
				}
			}
			extra = false
			r.Frame()
			afterImage := r.Image()
			var after []byte
			for y := int(a.Bounds.Y + a.Bounds.H/2); y < int(a.Bounds.Y+a.Bounds.H); y++ {
				for x := int(a.Bounds.X); x < int(a.Bounds.X+a.Bounds.W); x++ {
					v := afterImage.RGBAAt(x, y)
					after = append(after, v.R, v.G, v.B, v.A)
				}
			}
			if bytes.Equal(lower, after) {
				t.Fatal("removing a painted series did not change actual polar geometry")
			}
		})
	}
}
