package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestThirtyOnePolarSourceRecipesUsePublicNativePlots(t *testing.T) {
	all := recipes()
	if len(all) != 31 {
		t.Fatalf("source inventory%d", len(all))
	}
	seen := map[string]bool{}
	for _, r := range all {
		if seen[r.ID] {
			t.Fatal("duplicate source", r.ID)
		}
		seen[r.ID] = true
		a := newApp(r.ID)
		a.animate = false
		a.selected = -1
		native := ui.NewTesterWith(a.view, 900, 720, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		plot, ok := native.AccessByRole(ui.RoleGroup, r.Title)
		if !ok || plot.Bounds.W < 300 || plot.Bounds.H < 200 {
			t.Fatalf("missing source plot%s: %+v", r.ID, plot)
		}
		image := native.Image()
		painted := 0
		for y := int(plot.Bounds.Y); y < int(plot.Bounds.Y+plot.Bounds.H) && y < image.Rect.Max.Y; y++ {
			for x := int(plot.Bounds.X); x < int(plot.Bounds.X+plot.Bounds.W) && x < image.Rect.Max.X; x++ {
				c := image.RGBAAt(x, y)
				if int(c.B) > int(c.R)+25 || int(c.G) > int(c.R)+25 {
					painted++
				}
			}
		}
		if painted < 30 {
			t.Fatalf("source%s has no actual native marks (%d)", r.ID, painted)
		}
	}
}

func TestPolarSourceActiveMonthChangesShapeAndCenter(t *testing.T) {
	a := newApp("chart-pie-interactive")
	a.animate = false
	r := ui.NewTesterWith(a.view, 900, 740, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.HasText("186") {
		t.Fatal("initial source month center missing")
	}
	if err := r.Click("Active month"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
	if a.month != "may" || !r.HasText("209") {
		t.Fatalf("actual native source picker did not change May center: %s %v", a.month, r.Texts())
	}
	plot, _ := r.AccessByRole(ui.RoleGroup, a.recipe().Title)
	r.ClickAt(plot.Bounds.X+plot.Bounds.W/2, plot.Bounds.Y+30)
	r.Key(0, ui.KeyHome)
	if a.selected != 0 || !r.HasText("209") {
		t.Fatal("inspection changed the independent selected-month center")
	}
}
