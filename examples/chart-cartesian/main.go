// The 30 fixed shadcn Cartesian examples, through module-local shared builders
// implemented with public kit/native APIs. See chart-composition for an external sample.
package main

import (
	"flag"
	"fmt"
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

type recipe chartcatalog.CartesianRecipe

func recipes() []recipe {
	source := chartcatalog.CartesianRecipes()
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
	controller            theme.Controller
	id, rangeID, seriesID string
	selected              int
	dark, animate         bool
	all                   []recipe
}

func newApp(id string) *app {
	return &app{controller: theme.Controller{Preset: "blue"}, id: id, rangeID: "90d", seriesID: "desktop", selected: -1, animate: true, all: recipes()}
}
func (a *app) recipe() recipe {
	for _, r := range a.all {
		if r.ID == a.id {
			return r
		}
	}
	return a.all[0]
}

func (a *app) props(c *ui.Context) display.ChartProps {
	return chartcatalog.CartesianProps(c, chartcatalog.CartesianRecipe(a.recipe()), chartcatalog.Options{Selected: &a.selected, Animate: a.animate, Range: a.rangeID, Series: a.seriesID})
}

func (a *app) view(c *ui.Context) {
	a.controller.Mode = theme.LightMode
	if a.dark {
		a.controller.Mode = theme.DarkMode
	}
	a.controller.Apply(c)
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).FillWidth().Padding(28).Gap(16).Children(func() {
			display.Typography(c, "Cartesian charts", display.H2)
			options := make([]control.Option, len(a.all))
			for i, r := range a.all {
				options[i] = control.Option{Value: r.ID, Label: r.Title}
			}
			if control.Select(c, &a.id, control.SelectProps{Label: "Chart example", Options: options}).Changed() {
				a.selected = -1
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
				if r.ID == "chart-area-interactive" {
					control.Select(c, &a.rangeID, control.SelectProps{Label: "Date range", Options: []control.Option{{Value: "90d", Label: "Last 3 months"}, {Value: "30d", Label: "Last 30 days"}, {Value: "7d", Label: "Last 7 days"}}})
				}
				if r.ID == "chart-bar-interactive" || r.ID == "chart-line-interactive" {
					ui.Row(c).Gap(12).Children(func() {
						for _, key := range []string{"desktop", "mobile"} {
							total := float64(0)
							for _, row := range r.Data {
								total += row[key].(float64)
							}
							if control.Button(c, fmt.Sprintf("%s · %.0f", title(key), total), control.ButtonProps{Variant: control.Outline}).Clicked() {
								a.seriesID = key
								c.Invalidate()
							}
						}
					})
				}
				config := display.ChartConfig{}
				for key, label := range r.Labels {
					colors := theme.TokensOf(c).Charts()
					col := colors[r.Colors[key]%len(colors)]
					entry := display.ChartConfigItem{Label: label}
					if _, ok := r.Colors[key]; ok {
						entry.Color = &col
					}
					config[key] = entry
					config[title(key)] = config[key]
				}
				display.ChartContainer(c, display.ChartContainerProps{Key: "source-cartesian", Config: config}, func(c *ui.Context) {
					ui.Column(c).FillWidth().Children(func() {
						props := a.props(c)
						if r.ID == "chart-area-icons" {
							props.Legend = false
						}
						display.Chart(c, props)
						if r.ID == "chart-area-icons" {
							colors := theme.TokensOf(c).Charts()
							for _, key := range []string{"mobile", "desktop"} {
								entry := config[key]
								down := key == "desktop"
								entry.Icon = func(c *ui.Context) ui.Element {
									col := theme.TokensOf(c).MutedForeground
									return ui.Box(c).Size(12, 12).AccessHidden(true).Draw(func(p *ui.Painter, r ui.Rect) {
										y1, y2 := r.Y+r.H*.7, r.Y+r.H*.2
										if down {
											y1, y2 = y2, y1
										}
										p.StrokePath((&ui.Path{}).MoveTo(r.X, y1).LineTo(r.X+r.W*.35, r.Y+r.H*.45).LineTo(r.X+r.W*.6, r.Y+r.H*.65).LineTo(r.X+r.W, y2).MoveTo(r.X+r.W*.65, y2).LineTo(r.X+r.W, y2).LineTo(r.X+r.W, y2+r.H*.3), 1.5, col)
									})
								}
								config[key] = entry
							}
							display.ChartContainer(c, display.ChartContainerProps{Config: config}, func(c *ui.Context) {
								display.ChartLegendContent(c, display.ChartLegendContentProps{Payload: []display.ChartPayloadItem{{DataKey: "mobile", Color: &colors[1]}, {DataKey: "desktop", Color: &colors[0]}}})
							}).AspectRatio(0)
						}
					})
				}).AspectRatio(0)
				ui.Text(c, "Trending up by 5.2% this month ↗").FontSize(14)
				ui.Text(c, "Inspect values with the pointer or focus the plot and use the arrow keys.").FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			})
		})
	})
}
func main() {
	id := flag.String("example", "chart-area-default", "one of the 30 source example IDs")
	snapshot := flag.String("snapshot", "", "write a completed PNG instead of opening a window")
	dark := flag.Bool("dark", false, "dark appearance")
	inspect := flag.Bool("inspect", false, "initially inspect the third point")
	flag.Parse()
	a := newApp(*id)
	a.dark = *dark
	if *inspect {
		a.selected = 2
	}
	if *snapshot != "" {
		a.animate = false
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err = png.Encode(f, ui.Render(a.view, 1000, 650, 1)); err != nil {
			log.Fatal(err)
		}
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "Cartesian charts", Width: 1000, Height: 700, Content: ui.View(a.view)})
		w.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
