package chartcatalog

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func tester(s *State, e Entry) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() { s.Render(c, e) })
	}, 900, 650, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func entry(id string) Entry {
	for _, e := range Entries() {
		if e.ID == id {
			return e
		}
	}
	panic(id)
}
func focusPlot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	p, ok := r.AccessByRole(ui.RoleGroup, name)
	if !ok {
		t.Fatal("missing plot " + name)
	}
	r.ClickAt(p.Bounds.X+p.Bounds.W*.8, p.Bounds.Y+p.Bounds.H*.7)
}
func selectEnd(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	if err := r.Click(name); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
}

// One inventory/config build check, not 84 independent visual acceptances.
func TestSourceInventoryBuilds70Catalogue8Docs6Create(t *testing.T) {
	counts := map[string]int{}
	seen := map[string]bool{}
	for _, e := range Entries() {
		counts[e.Section]++
		if seen[e.ID] {
			t.Fatalf("duplicate %s", e.ID)
		}
		seen[e.ID] = true
		s := New()
		s.Animate = false
		r := tester(s, e)
		if e.ID == "docs-chart-tooltip" {
			if !r.HasText("Indicator") || !r.HasText("12,486") {
				t.Fatal("tooltip anatomy missing source content")
			}
			continue
		}
		plot, ok := r.AccessByRole(ui.RoleGroup, e.Title)
		if !ok || plot.Bounds.W < 100 || plot.Bounds.H < 100 {
			t.Fatalf("%s native plot absent: %+v", e.ID, plot)
		}
	}
	if counts["catalogue"] != 70 || counts["docs"] != 8 || counts["create"] != 6 {
		t.Fatalf("source inventory %v", counts)
	}
}
func TestTooltipNineUseSourceContentAndDefaultIndexOnlyOnce(t *testing.T) {
	for _, e := range Entries() {
		if e.Family != "tooltip" {
			continue
		}
		s := New()
		s.Animate = false
		r := tester(s, e)
		if s.Model(e.ID).Selected != 1 {
			t.Fatal("source defaultIndex1 absent")
		}
		if e.ID == "chart-tooltip-advanced" {
			if !r.HasText("Total") || !r.HasText("800 kcal") {
				t.Fatalf("source advanced total not original running+swimming (%s)", e.ID)
			}
		}
		if e.ID == "chart-tooltip-formatter" && !r.HasText("kcal") {
			t.Fatal("source formatter not used")
		}
		if e.ID == "chart-tooltip-label-custom" && !r.HasText("Activities") {
			t.Fatal("source labelKey config indirection absent")
		}
		if e.ID == "chart-tooltip-label-formatter" && !r.HasText("July 16, 2024") {
			t.Fatal("source date label formatter absent")
		}
		focusPlot(t, r, e.Title)
		r.Key(0, ui.KeyEscape)
		r.Frame()
		if s.Model(e.ID).Selected != -1 {
			t.Fatal("rebuild reasserted defaultIndex after Escape")
		}
	}
}
func TestSourceInteractiveControlsAndDocsDatasetRemainDistinct(t *testing.T) {
	s := New()
	s.Animate = false
	e := entry("chart-area-interactive")
	r := tester(s, e)
	selectEnd(t, r, "Date range")
	if s.Model(e.ID).Range != "7d" {
		t.Fatal("date range is not bound")
	}
	var p display.ChartProps
	check := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); p, _ = s.Props(c, e.ID); display.Chart(c, p) }, 700, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if len(p.Labels) != 8 || p.Labels[0] != "2024-06-23" {
		t.Fatalf("source date cutoff: %v", p.Labels)
	}
	_ = check
	e = entry("docs-chart-demo")
	r = tester(s, e)
	total := 0.0
	for _, source := range AdditionalRecipes() {
		if source.ID == e.ID {
			for _, row := range source.Data {
				total += row["mobile"].(float64)
			}
		}
	}
	if err := r.Click(fmt.Sprintf("Mobile · %.0f", total)); err != nil {
		t.Fatal(err)
	}
	focusPlot(t, r, e.Title)
	r.Key(0, ui.KeyHome)
	plot, _ := r.AccessByRole(ui.RoleGroup, e.Title)
	if s.Model(e.ID).Series != "mobile" || !strings.Contains(plot.Value, "Mobile: 150") || !strings.Contains(plot.Value, "2024-04-01") {
		t.Fatalf("docs30-row demo not live/source data: %+v", plot)
	}
	e = entry("chart-pie-interactive")
	r = tester(s, e)
	selectEnd(t, r, "Active month")
	if s.Model(e.ID).Month != "may" {
		t.Fatal("source month selection absent")
	}
	if !r.HasText("209") {
		t.Fatal("active month did not reach source center")
	}
}
func TestDocsRTLAndCreateCardsUseActualSharedConfigs(t *testing.T) {
	s := New()
	s.Animate = false
	e := entry("docs-chart-rtl")
	r := tester(s, e)
	first, ok := r.Find("سطح المكتب january: 186")
	last, _ := r.Find("سطح المكتب june: 214")
	if !ok || first.X <= last.X {
		t.Fatalf("Arabic source positions not reversed: %v %v", first, last)
	}
	selectEnd(t, r, "Chart language")
	if s.Model(e.ID).Language != "he" {
		t.Fatal("language selector not bound")
	}
	focusPlot(t, r, e.Title)
	r.Key(0, ui.KeyHome)
	plot, _ := r.AccessByRole(ui.RoleGroup, e.Title)
	if !strings.Contains(plot.Value, "מחשב") {
		t.Fatal("Hebrew labels not applied")
	}
	s.Collection = "create"
	s.ID = "create-area"
	r = ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); ui.Scroll(c).Fill().Children(func() { s.View(c) }) }, 900, 1500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.HasText("Area Chart") || !r.HasText("Radial Chart - Shape") {
		t.Fatal("Create six-card flow not built")
	}
	// Theme is read afresh by the shared public configuration, never cached CSS.
	var col ui.Color
	appearance := theme.Controller{Mode: theme.LightMode, Preset: "blue"}
	q := ui.NewTesterWith(func(c *ui.Context) {
		appearance.Apply(c)
		p, _ := s.Props(c, "create-area")
		col = p.Series[0].Color
		display.Chart(c, p)
	}, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	before := col
	appearance.Preset = "orange"
	q.SetDark(true)
	q.Frame()
	if before == col {
		t.Fatal("Create palette change did not reach shared source chart")
	}
	r.Advance(time.Millisecond)
}
