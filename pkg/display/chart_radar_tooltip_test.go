package display

import (
	"fmt"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestRadarTooltipIndicatorsKeepSeriesColorsAcrossCategories(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprint("dark=", dark, "/explicit=", explicit), func(t *testing.T) {
				selected := 0
				seriesColors := []ui.Color{ui.RGB(210, 45, 95), ui.RGB(20, 130, 220)}
				var expectedColors []ui.Color
				var observed ChartTooltipData
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					tokens := theme.Preset("blue", dark)
					palette := tokens.Charts()
					expectedColors = palette[:2]
					configured := []ui.Color{{}, {}}
					if explicit {
						configured = seriesColors
						expectedColors = seriesColors
					}
					theme.Scope(c, tokens).Fill().Background(tokens.Background).Padding(20).Children(func() {
						ui.Column(c).Fill().Gap(12).Children(func() {
							Chart(c, ChartProps{
								Kind: RadarChart, Height: 240, Label: "Radar series", Legend: true, Selected: &selected,
								Labels: []string{"January", "February", "March", "April"},
								Series: []ChartSeries{
									{Name: "Desktop", Values: []float64{4, 3, 2, 4}, Color: configured[0]},
									{Name: "Mobile", Values: []float64{1, 2, 3, 1}, Color: configured[1]},
								},
								// Radar marks are series-colored; these categorical overrides
								// belong to pie/donut/unstacked radial plots and single bars.
								CategoryColors: []ui.Color{ui.RGB(10, 190, 50), ui.RGB(240, 175, 10)},
								TooltipContent: func(data ChartTooltipData) {
									observed = data
								},
							})
							ChartTooltip(c, observed).Role(ui.RoleGroup).Label("Radar value tooltip")
						})
					})
				}, 460, 440)
				check := func(category string) {
					if observed.Label != category || len(observed.Items) != 2 {
						t.Fatalf("wrong raw radar payload: %+v", observed)
					}
					surface, ok := r.AccessByRole(ui.RoleGroup, "Radar value tooltip")
					if !ok {
						t.Fatal("missing actual tooltip surface")
					}
					for i, name := range []string{"Desktop", "Mobile"} {
						if observed.Items[i].Color != expectedColors[i] {
							t.Fatalf("%s %s tooltip color%v differs from its series%v", category, name, observed.Items[i].Color, expectedColors[i])
						}
						var row ui.AccessInfo
						for _, node := range r.AccessAll() {
							if node.Role == ui.RoleText && node.Name == name && node.Bounds.Y >= surface.Bounds.Y && node.Bounds.Y < surface.Bounds.Y+surface.Bounds.H {
								row = node
								break
							}
						}
						if row.Name == "" {
							t.Fatal("missing rendered tooltip row", name)
						}
						want := polarThemeLegendColor(t, r, name)
						if got := r.Image().RGBAAt(int(surface.Bounds.X+16), int(row.Bounds.Y+row.Bounds.H/2)); got != want {
							t.Fatalf("actual %s indicator got%v want%v", name, got, want)
						}
					}
				}
				check("January")
				if err := r.Click("Radar series"); err != nil {
					t.Fatal(err)
				}
				r.Key(0, ui.KeyEnd)
				check("April")
				plot, _ := r.AccessByRole(ui.RoleGroup, "Radar series")
				if !plot.Focused || !strings.Contains(plot.Value, "April, Desktop: 4, Mobile: 1") {
					t.Fatalf("palette fix changed native inspection: %+v", plot)
				}
			})
		}
	}
}
