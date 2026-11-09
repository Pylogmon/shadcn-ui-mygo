// The 31 fixed shadcn Pie, Radar and Radial recipes through module-local shared
// public-API builders. See chart-composition for an external sample.
package main

import (
	"flag"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/chartcatalog"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type recipe chartcatalog.PolarRecipe

func recipes() []recipe {
	source := chartcatalog.PolarRecipes()
	out := make([]recipe, len(source))
	for i, r := range source {
		out[i] = recipe(r)
	}
	return out
}
func ptr[T any](v T) *T { return &v }
func title(v string) string {
	if v == "" {
		return v
	}
	return strings.ToUpper(v[:1]) + v[1:]
}

type app struct {
	appearance               theme.Controller
	all                      []recipe
	id, month                string
	selected, selectedSeries int
	dark, animate            bool
}

func newApp(id string) *app {
	return &app{appearance: theme.Controller{Preset: "blue"}, all: recipes(), id: id, month: "january", selected: -1, animate: true}
}
func (a *app) recipe() recipe {
	for _, r := range a.all {
		if r.ID == a.id {
			return r
		}
	}
	return a.all[0]
}
func (r recipe) rows() []map[string]any {
	if rows := r.Arrays["chartData"]; rows != nil {
		return rows
	}
	return r.Arrays["desktopData"]
}
func (r recipe) category() string {
	if _, ok := r.rows()[0]["browser"]; ok {
		return "browser"
	}
	return "month"
}
func (a *app) props(c *ui.Context) display.ChartProps {
	return chartcatalog.PolarProps(c, chartcatalog.PolarRecipe(a.recipe()), chartcatalog.Options{Selected: &a.selected, Animate: a.animate, SelectedSeries: &a.selectedSeries, Month: a.month})
}

func (a *app) view(c *ui.Context) {
	a.appearance.Mode = theme.LightMode
	if a.dark {
		a.appearance.Mode = theme.DarkMode
	}
	a.appearance.Apply(c)
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).FillWidth().Padding(28).Gap(16).Children(func() {
			display.Typography(c, "Polar charts", display.H2)
			options := make([]control.Option, len(a.all))
			for i, r := range a.all {
				options[i] = control.Option{Value: r.ID, Label: r.Title}
			}
			if control.Select(c, &a.id, control.SelectProps{Label: "Chart example", Options: options}).Changed() {
				a.selected, a.selectedSeries = -1, 0
				c.Invalidate()
			}
			ui.Row(c).Gap(18).Children(func() {
				control.Switch(c, &a.dark, control.SwitchProps{Label: "Dark appearance"})
				control.Switch(c, &a.animate, control.SwitchProps{Label: "Animate chart changes"})
			})
			r := a.recipe()
			display.Card(c, func() {
				ui.Text(c, r.Title).FontSize(20).FontWeight(600)
				ui.Text(c, "Source: "+r.ID).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
				if r.ID == "chart-pie-interactive" {
					options := make([]control.Option, len(r.rows()))
					for i, row := range r.rows() {
						key := row[r.category()].(string)
						options[i] = control.Option{Value: key, Label: r.Labels[key]}
					}
					control.Select(c, &a.month, control.SelectProps{Label: "Active month", Options: options})
				}
				config := display.ChartConfig{}
				palette := theme.TokensOf(c).Charts()
				for key, label := range r.Labels {
					item := display.ChartConfigItem{Label: label}
					if index, ok := r.Colors[key]; ok {
						col := palette[index%len(palette)]
						item.Color = &col
					}
					config[key] = item
					config[title(key)] = item
				}
				display.ChartContainer(c, display.ChartContainerProps{Key: "source-polar", Config: config}, func(c *ui.Context) {
					ui.Column(c).FillWidth().Children(func() {
						display.Chart(c, a.props(c))
						if r.ID == "chart-pie-legend" {
							payload := make([]display.ChartPayloadItem, len(r.rows()))
							for i, row := range r.rows() {
								key := row[r.category()].(string)
								col := palette[r.Colors[key]%len(palette)]
								payload[i] = display.ChartPayloadItem{DataKey: "visitors", Name: key, Color: &col, Fields: row, Payload: row}
							}
							display.ChartLegendContent(c, display.ChartLegendContentProps{Payload: payload, NameKey: "browser"})
						}
						if r.ID == "chart-radar-icons" {
							for _, key := range []string{"desktop", "mobile"} {
								item := config[key]
								arrow := "↗"
								if key == "desktop" {
									arrow = "↘"
								}
								item.Icon = func(c *ui.Context) ui.Element { return ui.Text(c, arrow).AccessHidden(true) }
								config[key] = item
							}
							display.ChartContainer(c, display.ChartContainerProps{Config: config}, func(c *ui.Context) {
								display.ChartLegendContent(c, display.ChartLegendContentProps{Payload: []display.ChartPayloadItem{{DataKey: "desktop", Color: &palette[0]}, {DataKey: "mobile", Color: &palette[1]}}})
							}).AspectRatio(0)
						}
					})
				}).AspectRatio(0)
				ui.Text(c, "Inspect with the pointer or arrow keys. Page Up / Page Down change stacked Pie rings.").FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			})
		})
	})
}
func main() {
	id := flag.String("example", "chart-pie-donut-text", "one of the 31 source IDs")
	snapshot := flag.String("snapshot", "", "write completed PNG")
	dark := flag.Bool("dark", false, "dark appearance")
	flag.Parse()
	a := newApp(*id)
	a.dark = *dark
	if *snapshot != "" {
		a.animate = false
		f, e := os.Create(*snapshot)
		if e != nil {
			log.Fatal(e)
		}
		if e = png.Encode(f, ui.Render(a.view, 900, 700, 1)); e != nil {
			log.Fatal(e)
		}
		if e = f.Close(); e != nil {
			log.Fatal(e)
		}
		return
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Polar charts", Width: 900, Height: 720, Content: ui.View(a.view)})
		w.Show()
	})
	if e := mygo.App.Run(); e != nil {
		log.Fatal(e)
	}
}
