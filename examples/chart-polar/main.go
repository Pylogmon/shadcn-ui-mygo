// The 31 fixed shadcn Pie, Radar and Radial recipes through public APIs.
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed recipes.json
var recipeJSON []byte

type recipe struct {
	ID, Title string
	Arrays    map[string][]map[string]any
	Colors    map[string]int
	Labels    map[string]string
	Margins   map[string]float32
	Height    float32
	Series    []string
}

func recipes() []recipe {
	var v []recipe
	if err := json.Unmarshal(recipeJSON, &v); err != nil {
		panic(err)
	}
	return v
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
	r := a.recipe()
	rows := r.rows()
	tokens := theme.TokensOf(c)
	palette := tokens.Charts()
	category := r.category()
	p := display.ChartProps{Kind: display.PieChart, Label: r.Title, Height: r.Height, Selected: &a.selected, Animation: display.ChartAnimationProps{Disabled: !a.animate}, Polar: &display.ChartPolarProps{SelectedSeries: &a.selectedSeries}}
	if len(r.Margins) > 0 {
		p.Polar.Margin = &display.ChartInsets{Top: r.Margins["top"], Right: r.Margins["right"], Bottom: r.Margins["bottom"], Left: r.Margins["left"]}
	}
	if strings.HasPrefix(r.ID, "chart-radar-") {
		p.Kind = display.RadarChart
		p.Polar.AngleAxis = &display.ChartPolarAngleAxisProps{}
		p.Polar.Grid = &display.ChartPolarGridProps{}
	}
	if strings.HasPrefix(r.ID, "chart-radial-") {
		p.Kind = display.RadialChart
		p.Polar.InnerRadius = ptr(display.ChartRadiusDIP(30))
		p.Polar.OuterRadius = ptr(display.ChartRadiusDIP(110))
	}
	for _, row := range rows {
		label := row[category].(string)
		p.Labels = append(p.Labels, label)
		index, ok := r.Colors[label]
		if !ok {
			index = 0
		}
		p.CategoryColors = append(p.CategoryColors, palette[index%len(palette)])
	}
	for _, key := range r.Series {
		data := rows
		if other, ok := r.Arrays[key+"Data"]; ok {
			data = other
		}
		s := display.ChartSeries{Name: r.Labels[key], Color: palette[r.Colors[key]%len(palette)], Style: display.ChartSeriesStyle{Polar: &display.ChartPolarSeriesProps{}}}
		if s.Name == "" {
			s.Name = title(key)
		}
		for _, row := range data {
			s.Values = append(s.Values, row[key].(float64))
		}
		if p.Kind == display.RadarChart && key == "desktop" {
			s.Style.FillOpacity = ptr(float32(.6))
		}
		p.Series = append(p.Series, s)
	}
	if r.ID == "chart-pie-stacked" {
		p.Stacked = true
		p.Series[0].Style.Polar.OuterRadius = ptr(display.ChartRadiusDIP(60))
		p.Series[1].Style.Polar.InnerRadius = ptr(display.ChartRadiusDIP(70))
		p.Series[1].Style.Polar.OuterRadius = ptr(display.ChartRadiusDIP(90))
	}
	if strings.Contains(r.ID, "donut") || r.ID == "chart-pie-interactive" {
		p.Kind = display.DonutChart
		p.Polar.InnerRadius = ptr(display.ChartRadiusDIP(60))
	}
	if r.ID == "chart-pie-donut-active" || r.ID == "chart-pie-donut-text" || r.ID == "chart-pie-interactive" {
		p.Polar.StrokeWidth = ptr(float32(5))
	}
	if r.ID == "chart-pie-separator-none" {
		p.Polar.StrokeWidth = ptr(float32(0))
	}
	if r.ID == "chart-pie-donut-active" {
		p.Series[0].Style.Polar.Shape = func(s display.ChartPolarSector) []display.ChartPolarSector {
			if s.Datum.Index == 0 {
				s.OuterRadius = s.OuterRadius.Add(10)
			}
			return []display.ChartPolarSector{s}
		}
	}
	if r.ID == "chart-pie-interactive" {
		active := 0
		for i, row := range rows {
			if row[category].(string) == a.month {
				active = i
			}
		}
		p.Series[0].Style.Polar.Shape = func(s display.ChartPolarSector) []display.ChartPolarSector {
			if s.Datum.Index != active {
				return []display.ChartPolarSector{s}
			}
			outer := s
			outer.InnerRadius, outer.OuterRadius = s.OuterRadius.Add(12), s.OuterRadius.Add(25)
			s.OuterRadius = s.OuterRadius.Add(10)
			return []display.ChartPolarSector{s, outer}
		}
		value := rows[active]["desktop"].(float64)
		p.Polar.Center = func(c *ui.Context, _ display.ChartPolarCenter) { center(c, value, "Visitors", 30) }
	}
	if r.ID == "chart-pie-donut-text" {
		p.Polar.Center = func(c *ui.Context, d display.ChartPolarCenter) { center(c, d.Total, "Visitors", 30) }
	}
	if r.ID == "chart-pie-label" || r.ID == "chart-pie-label-custom" {
		p.Series[0].Style.Polar.Labels = []display.ChartPolarLabelProps{{Position: display.ChartPolarLabelOutside, HideLine: r.ID == "chart-pie-label-custom"}}
	}
	if r.ID == "chart-pie-label-list" {
		p.Series[0].Style.Polar.Labels = []display.ChartPolarLabelProps{{Color: &tokens.Background, Formatter: func(d display.ChartDatum) string { return r.Labels[d.Label] }}}
	}
	if r.ID == "chart-radar-grid-none" {
		p.Polar.Grid = nil
	}
	if strings.Contains(r.ID, "grid-circle") {
		p.Polar.Grid.Kind = display.ChartPolarCircle
	}
	if strings.Contains(r.ID, "no-lines") || r.ID == "chart-radar-grid-custom" || r.ID == "chart-radar-lines-only" {
		p.Polar.Grid.RadialLines = ptr(false)
	}
	if r.ID == "chart-radar-grid-custom" {
		p.Polar.Grid.Radii = []display.ChartRadius{display.ChartRadiusDIP(90)}
	}
	if r.ID == "chart-radar-grid-fill" || r.ID == "chart-radar-grid-circle-fill" {
		col := p.Series[0].Color.Alpha(.2)
		stroke := tokens.Border.Alpha(.2)
		p.Polar.Grid.Fill = &col
		p.Polar.Grid.Stroke = &stroke
		p.Series[0].Style.FillOpacity = ptr(float32(.5))
	}
	if r.ID == "chart-radar-dots" || r.ID == "chart-radar-grid-none" || r.ID == "chart-radar-grid-circle" || r.ID == "chart-radar-grid-circle-no-lines" {
		p.Series[0].Style.Dots = &display.ChartDotProps{Radius: 4, StrokeWidth: ptr(float32(0))}
	}
	if r.ID == "chart-radar-lines-only" {
		for j := range p.Series {
			p.Series[j].Style.FillOpacity = ptr(float32(0))
			p.Series[j].Style.Stroke = &p.Series[j].Color
			p.Series[j].Style.StrokeWidth = ptr(float32(2))
		}
	}
	if r.ID == "chart-radar-radius" {
		p.Polar.AngleAxis = nil
		p.Polar.RadiusAxis = &display.ChartPolarRadiusAxisProps{Angle: 60, TextAlign: ui.Center, HideLine: true, Color: &tokens.Foreground}
	}
	if r.ID == "chart-radar-label-custom" {
		p.Polar.AngleAxis.Content = func(c *ui.Context, d display.ChartPolarTick) {
			e := ui.Column(c).AlignItems(d.TextAlign)
			if d.Index == 0 {
				e.Translate(0, -10)
			}
			e.Children(func() {
				ui.Text(c, fmt.Sprintf("%g/%g", d.Values[0], d.Values[1])).FontSize(13).FontWeight(500)
				ui.Text(c, d.Label).FontSize(12).TextColor(tokens.MutedForeground)
			})
		}
	}
	if r.ID == "chart-radial-simple" || r.ID == "chart-radial-label" || r.ID == "chart-radial-shape" || r.ID == "chart-radial-text" {
		p.Polar.Background = true
	}
	if r.ID == "chart-radial-grid" {
		p.Polar.OuterRadius = ptr(display.ChartRadiusDIP(100))
		p.Polar.Grid = &display.ChartPolarGridProps{Kind: display.ChartPolarCircle}
	}
	if r.ID == "chart-radial-label" {
		p.Polar.StartAngle, p.Polar.EndAngle = ptr(float64(-90)), ptr(float64(380))
		white := ui.RGB(255, 255, 255)
		p.Series[0].Style.Polar.Labels = []display.ChartPolarLabelProps{{Position: display.ChartPolarLabelInsideStart, FontSize: 11, Color: &white, Formatter: func(d display.ChartDatum) string { return title(d.Label) }}}
	}
	if r.ID == "chart-radial-shape" || r.ID == "chart-radial-text" {
		inner, outer, end := float32(65), float32(95), 100.0
		radii := []display.ChartRadius{display.ChartRadiusDIP(86), display.ChartRadiusDIP(74)}
		if r.ID == "chart-radial-text" {
			inner, outer, end = 80, 90, 250
			radii = []display.ChartRadius{display.ChartRadiusDIP(90), display.ChartRadiusDIP(80)}
			p.Polar.CornerRadius = 10
		}
		p.Polar.EndAngle = ptr(end)
		p.Polar.InnerRadius = ptr(display.ChartRadiusDIP(inner))
		p.Polar.OuterRadius = ptr(display.ChartRadiusDIP(outer))
		transparent := ui.Transparent
		p.Polar.Grid = &display.ChartPolarGridProps{Kind: display.ChartPolarCircle, RadialLines: ptr(false), Radii: radii, Fills: []ui.Color{tokens.Muted, tokens.Background}, Stroke: &transparent}
		p.Polar.Center = func(c *ui.Context, d display.ChartPolarCenter) { center(c, d.Total, "Visitors", 36) }
	}
	if r.ID == "chart-radial-stacked" {
		p.Stacked = true
		p.Polar.EndAngle = ptr(float64(180))
		p.Polar.InnerRadius = ptr(display.ChartRadiusDIP(80))
		p.Polar.OuterRadius = ptr(display.ChartRadiusDIP(110))
		p.Polar.CornerRadius = 5
		p.Polar.Center = func(c *ui.Context, d display.ChartPolarCenter) {
			ui.Box(c).Translate(0, -16).Children(func() { center(c, d.Total, "Visitors", 24) })
		}
	}
	if r.ID == "chart-radar-legend" {
		p.Legend = true
	}
	p.TooltipContent = func(d display.ChartTooltipData) {
		payload := make([]display.ChartPayloadItem, len(d.Items))
		for i, item := range d.Items {
			col := item.Color
			name := strings.ToLower(item.Label)
			dataKey := name
			row := map[string]any{}
			for _, candidate := range rows {
				if candidate[category] == d.Label {
					row = candidate
					break
				}
			}
			if p.Kind == display.PieChart || p.Kind == display.DonutChart || p.Kind == display.RadialChart {
				name = d.Label
			}
			payload[i] = display.ChartPayloadItem{Name: name, DataKey: dataKey, Value: item.Value, Color: &col, Fields: row, Payload: row}
		}
		indicator := display.ChartIndicatorDot
		hide := p.Kind != display.RadarChart
		if r.ID == "chart-radar-radius" {
			indicator = display.ChartIndicatorLine
		}
		label := &display.ChartContentLabel{Text: d.Label}
		if r.ID == "chart-pie-stacked" {
			hide = false
			indicator = display.ChartIndicatorLine
			if len(d.Items) > 0 {
				label.Text = d.Items[0].Label
			}
		}
		display.ChartTooltipContent(c, display.ChartTooltipContentProps{Active: true, Label: label, Payload: payload, Indicator: indicator, HideLabel: hide})
	}
	return p
}

func center(c *ui.Context, value float64, label string, size float32) {
	ui.Column(c).AlignItems(ui.Center).Gap(2).Children(func() {
		ui.Text(c, message.NewPrinter(language.English).Sprintf("%.0f", value)).FontSize(size).FontWeight(700)
		ui.Text(c, label).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
	})
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
