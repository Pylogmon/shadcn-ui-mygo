package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/chartcatalog"
	"testing"
)

func TestGalleryChartSourceDirectoryKeepsDraftsAndBoundSearch(t *testing.T) {
	s := New()
	s.Page = "display"
	s.Email = "draft@example.test"
	s.Query = "global query"
	r := ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		ui.Scroll(c).Fill().Children(func() {
			if s.Page == "display" {
				s.chartSourcePreview(c)
			} else {
				ui.Text(c, "Another page")
			}
		})
	}, 900, 1050, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Find chart example"); err != nil {
		t.Fatal(err)
	}
	r.Type("chart-tooltip-advanced")
	if s.chartCatalogue.Query != "chart-tooltip-advanced" || !r.HasText("800 kcal") {
		t.Fatal("source directory search did not render the selected actual tooltip")
	}
	model := s.chartCatalogue.Model("chart-tooltip-advanced")
	model.Selected = 3
	s.Page = "other"
	r.SetDark(true)
	s.Page = "display"
	r.SetDark(false)
	if s.chartCatalogue.Model("chart-tooltip-advanced").Selected != 3 || s.Email != "draft@example.test" || s.Query != "global query" {
		t.Fatal("chart directory changed Gallery draft/global-search state")
	}
	count := 0
	for _, entry := range chartcatalog.Entries() {
		if entry.Section == "catalogue" {
			count++
		}
	}
	if count != 70 {
		t.Fatalf("Gallery directory has %d source entries", count)
	}
}
