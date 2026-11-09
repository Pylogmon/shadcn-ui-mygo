package display

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestChartLegacyEmptySurfacesDoNotReserveParentGaps(t *testing.T) {
	for _, name := range []string{"tooltip", "legend"} {
		t.Run(name, func(t *testing.T) {
			populated := false
			text := "Original"
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Gap(10).Children(func() {
					ui.Box(c).Size(20, 20).Label("Before")
					if name == "tooltip" {
						data := ChartTooltipData{}
						if populated {
							data.Items = []ChartTooltipItem{{Label: "Revenue", Value: 20, Color: ui.RGB(20, 80, 220)}}
						}
						ChartTooltip(c, data)
					} else {
						var items []ChartLegendItem
						if populated {
							items = []ChartLegendItem{{Label: "Revenue", Color: ui.RGB(20, 80, 220)}}
						}
						ChartLegend(c, items)
					}
					ui.TextInput(c, &text).Label("After editor")
				})
			}, 400, 260)
			checkGap := func() {
				before, _ := r.Find("Before")
				after, _ := r.Find("After editor")
				if gap := after.Y - before.Y - before.H; gap != 10 {
					t.Fatalf("empty %s reserved a flex slot: gap%g want10", name, gap)
				}
			}
			checkGap()
			if err := r.Click("After editor"); err != nil {
				t.Fatal(err)
			}
			r.Command("selectAll")
			populated = true
			r.Frame()
			if !r.HasText("Revenue") {
				t.Fatal("nonempty surface failed to enter the native layout")
			}
			populated = false
			r.Frame()
			checkGap()
			r.Type("Retained")
			if text != "Retained" || !r.Focused("After editor") {
				t.Fatal("empty/content toggles remounted the adjacent editor")
			}
		})
	}
}
