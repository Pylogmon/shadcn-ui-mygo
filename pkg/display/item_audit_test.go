package display

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestItemPrimaryStaysInsideVeryNarrowSurface(t *testing.T) {
	for _, width := range []float32{60, 80, 100} {
		r := displayTester(func(c *ui.Context) {
			ui.Column(c).Width(width).Children(func() {
				Item(c, ItemProps{Label: "Narrow surface", Title: "Long title", Primary: &ItemPrimary{Label: "Open narrow item"}})
			})
		})
		outer, _ := r.AccessByRole(ui.RoleGroup, "Narrow surface")
		primary, _ := r.AccessByRole(ui.RoleButton, "Open narrow item")
		if primary.Bounds.X < outer.Bounds.X || primary.Bounds.X+primary.Bounds.W > outer.Bounds.X+outer.Bounds.W || primary.Bounds.W <= 0 {
			t.Fatalf("width %g primary target escaped its surface: outer=%+v primary=%+v", width, outer.Bounds, primary.Bounds)
		}
	}
}

func TestItemCustomMediaKeepsNativeControlAccessibility(t *testing.T) {
	r := displayTester(func(c *ui.Context) {
		Item(c, ItemProps{Title: "Project", Media: &ItemMediaProps{Content: func(child *ui.Context) {
			ui.Button(child, "Select project")
		}}})
	})
	r.Key(0, ui.KeyTab)
	if !r.Focused("Select project") {
		t.Fatal("custom media did not retain native interaction")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Select project"); !ok {
		t.Fatal("focusable custom media was hidden from accessibility")
	}
}
