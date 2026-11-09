package main

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

// This is one inventory/build check for the exact 30 requested source recipes,
// not an additional combinatorial styling matrix.
func TestThirtySourceRecipesBuildThroughPublicAPIs(t *testing.T) {
	all := recipes()
	if len(all) != 30 {
		t.Fatalf("source inventory has%d recipes", len(all))
	}
	seen := map[string]bool{}
	for _, source := range all {
		if seen[source.ID] {
			t.Fatalf("duplicate source%s", source.ID)
		}
		seen[source.ID] = true
		a := newApp(source.ID)
		a.animate = false
		a.selected = 1
		r := ui.NewTesterWith(a.view, 1000, 780, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		plot, ok := r.AccessByRole(ui.RoleGroup, source.Title)
		if !ok || plot.Bounds.W < 300 || plot.Bounds.H < 200 {
			t.Fatalf("source%s missing actual plot: %+v", source.ID, plot)
		}
		painted := 0
		img := r.Image()
		for y := int(plot.Bounds.Y); y < int(plot.Bounds.Y+plot.Bounds.H) && y < img.Rect.Max.Y; y++ {
			for x := int(plot.Bounds.X); x < int(plot.Bounds.X+plot.Bounds.W) && x < img.Rect.Max.X; x++ {
				v := img.RGBAAt(x, y)
				if int(v.B) > int(v.R)+30 || int(v.G) > int(v.R)+30 || int(v.R) > int(v.B)+30 {
					painted++
				}
			}
		}
		if painted < 30 {
			t.Fatalf("source%s has no actual marks (%d colored pixels)", source.ID, painted)
		}
	}
}

func TestCartesianSourceInteractiveControlsChangeRenderedData(t *testing.T) {
	a := newApp("chart-bar-interactive")
	a.animate = false
	a.selected = 1
	r := ui.NewTesterWith(a.view, 1000, 800, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	total := float64(0)
	for _, row := range a.recipe().Data {
		total += row["mobile"].(float64)
	}
	if err := r.Click(fmt.Sprintf("Mobile · %.0f", total)); err != nil {
		t.Fatal(err)
	}
	plot, _ := r.AccessByRole(ui.RoleGroup, a.recipe().Title)
	if a.seriesID != "mobile" || plot.Value == "" {
		t.Fatalf("real series button did not update chart data: %s %+v", a.seriesID, plot)
	}
	a.id = "chart-area-interactive"
	a.selected = -1
	r.Frame()
	if err := r.Click("Date range"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
	if a.rangeID != "7d" {
		t.Fatalf("native date selector did not commit source range: %s", a.rangeID)
	}
	plot, _ = r.AccessByRole(ui.RoleGroup, a.recipe().Title)
	r.ClickAt(plot.Bounds.X+plot.Bounds.W/2, plot.Bounds.Y+plot.Bounds.H/2)
	r.Key(0, ui.KeyHome)
	if a.selected != 0 || plot.Bounds.W < 300 {
		t.Fatal("filtered plot did not retain native keyboard inspection")
	}
}
