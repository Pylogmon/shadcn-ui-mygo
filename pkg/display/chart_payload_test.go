package display_test

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestChartRawPayloadUsesLatestValuesAndDetachedOriginalRows(t *testing.T) {
	selected := 0
	raw := map[string]any{"date": "2024-07-16", "running": 380.0, "swimming": 420.0, "extra": map[string]any{"units": []string{"kcal"}}}
	p := display.ChartProps{Label: "Full payload", Height: 240, Labels: []string{"2024-07-16", "2024-07-17"}, Data: []map[string]any{raw, {"running": 520.0}}, Series: []display.ChartSeries{{Name: "Running label", DataKey: "running", Values: []float64{380, 520}}, {Name: "Swimming label", DataKey: 0, TooltipName: "swimming", Values: []float64{420, math.NaN()}}}, Selected: &selected}
	var seen display.ChartTooltipData
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		p.TooltipContent = func(d display.ChartTooltipData) {
			seen = d
			display.ChartTooltipContent(c, display.ChartTooltipContentProps{Active: true, Payload: d.Payload, Formatter: func(c *ui.Context, f display.ChartTooltipFormatData) {
				if f.Index == 0 {
					ui.Text(c, "raw total")
					if f.Payload["running"] != p.Data[d.Index]["running"] {
						t.Error("formatter did not receive original row")
					}
				}
			}})
		}
		display.ChartContainer(c, display.ChartContainerProps{}, func(c *ui.Context) { display.Chart(c, p) }).AspectRatio(0)
	}, 600, 320, ui.TesterOptions{})
	if seen.Index != 0 || len(seen.Payload) != 2 || seen.Payload[0].Name != "running" || seen.Payload[1].DataKey != 0 || seen.Payload[1].Name != "swimming" || seen.Payload[0].Value != 380.0 || !r.HasText("raw total") {
		t.Fatalf("full public payload absent: %+v", seen)
	}
	p.Series[1].Data = []map[string]any{{"swimming": 420.0, "seriesSource": "pool"}}
	r.SetDark(true)
	if seen.Payload[1].Payload["seriesSource"] != "pool" {
		t.Fatal("series dataset override was not preserved")
	}
	saved := seen.Payload[0]
	saved.Payload["running"] = 1.0
	saved.Payload["extra"].(map[string]any)["units"].([]string)[0] = "changed"
	if raw["running"] != 380.0 || raw["extra"].(map[string]any)["units"].([]string)[0] != "kcal" {
		t.Fatal("formatter payload mutated caller data")
	}
	raw["running"] = 900.0
	p.Series[0].Values[0] = 900
	r.Frame()
	r.Advance(100 * time.Millisecond)
	if seen.Payload[0].Value != 900.0 || seen.Payload[0].Payload["running"] != 900.0 || seen.Items[0].Value != 900 {
		t.Fatal("payload exposed animated/stale values")
	}
	selected = 1
	r.Frame()
	if len(seen.Payload) != 1 || seen.Index != 1 {
		t.Fatalf("missing value was not filtered with its metadata: %+v", seen)
	}
}
func TestChartReversedCategoryMirrorsBandsWithoutReorderingSeries(t *testing.T) {
	selected := -1
	p := display.ChartProps{Label: "Reversed plot", Height: 240, Labels: []string{"First", "Second", "Third"}, Series: []display.ChartSeries{{Name: "Desktop", Values: []float64{10, 20, 30}}, {Name: "Mobile", Values: []float64{5, 10, 15}}}, Selected: &selected, Animation: display.ChartAnimationProps{Disabled: true}, Cartesian: display.ChartCartesianProps{ReversedCategory: true}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Direction(ui.RightToLeft).Children(func() { display.Chart(c, p) })
	}, 600, 320, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	a, ok := r.Find("Desktop First: 10")
	b, _ := r.Find("Mobile First: 5")
	last, _ := r.Find("Desktop Third: 30")
	if !ok || a.X <= last.X || a.X >= b.X {
		t.Fatalf("category or grouped series order incorrect: %v %v %v", a, b, last)
	}
	plot, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	r.Move(plot.Bounds.X+plot.Bounds.W*.95, plot.Bounds.Y+plot.Bounds.H/2)
	if selected != 0 {
		t.Fatalf("rightmost source datum pointer: %d", selected)
	}
	r.Click(p.Label)
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyLeft)
	if selected != 1 {
		t.Fatalf("physical Left should inspect second source datum: %d", selected)
	}
	r.Key(0, ui.KeyRight)
	if selected != 0 {
		t.Fatal("Right did not return to first datum")
	}
	r.Key(0, ui.KeyEscape)
	if selected != -1 {
		t.Fatal("Escape lost inspection semantics")
	}
	// Zero/default remains the original physical order, even under RTL text.
	p.Cartesian.ReversedCategory = false
	r.Frame()
	a, _ = r.Find("Desktop First: 10")
	last, _ = r.Find("Desktop Third: 30")
	if a.X >= last.X {
		t.Fatal("default physical order changed")
	}
}

func TestChartInspectionTooltipUsesAxisGapPointerAndEdgeFlip(t *testing.T) {
	near := func(got, want float32) bool {
		return !math.IsNaN(float64(got)) && !math.IsInf(float64(got), 0) && math.Abs(float64(got-want)) <= 1
	}
	finiteRect := func(r ui.Rect) bool {
		for _, v := range []float32{r.X, r.Y, r.W, r.H} {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return false
			}
		}
		return true
	}
	selected := 0
	var tip ui.Rect
	p := display.ChartProps{Label: "Positioned chart", Height: 240, Labels: []string{"First", "Middle", "Last"}, Series: []display.ChartSeries{{Name: "Visitors", Values: []float64{10, 20, 30}}}, Selected: &selected, TooltipFollowInspection: true, Animation: display.ChartAnimationProps{Disabled: true}, Cartesian: display.ChartCartesianProps{HideValueAxis: true, HideGrid: true}, TooltipContent: func(d display.ChartTooltipData) {}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		p.TooltipContent = func(d display.ChartTooltipData) {
			display.ChartTooltip(c, d).Width(140).DrawOver(func(_ *ui.Painter, r ui.Rect) { tip = r })
		}
		ui.Column(c).Fill().Padding(20).Children(func() { display.Chart(c, p) })
	}, 600, 320, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	first, _ := r.Find("Visitors First: 10")
	plot, _ := r.AccessByRole(ui.RoleGroup, p.Label)
	// Pinned getTooltipTranslateXY has a 10px positive offset; measure the
	// actual native content box and actual axis band, not the private helper.
	// Official combineCoordinateForDefaultIndex uses chart height, including
	// the category axis, and its top offset rather than the plot midpoint.
	if !near(tip.Y, 20+240.0/2+10) {
		t.Fatalf("defaultIndex non-axis coordinate excluded the native axis: %v", tip)
	}
	if !near(tip.X, first.X+first.W/2+10) {
		t.Fatalf("tooltip not beside inspected first category: %v mark%v", tip, first)
	}
	r.Move(plot.Bounds.X+plot.Bounds.W*.95, plot.Bounds.Y+plot.Bounds.H*.95)
	if selected != 2 || !finiteRect(tip) || tip.X+tip.W > plot.Bounds.X+plot.Bounds.W+1 || tip.Y+tip.H >= plot.Bounds.Y+plot.Bounds.H {
		t.Fatalf("source right/bottom edge flip absent: %d %v area%v", selected, tip, plot.Bounds)
	}
	// Legacy callers retain their established fixed corner, including custom content.
	p.TooltipFollowInspection = false
	r.Key(0, ui.KeyHome)
	r.Frame()
	if !near(tip.X, plot.Bounds.X+8) || !near(tip.Y, plot.Bounds.Y+8) {
		t.Fatalf("default placement changed: %v", tip)
	}
}
