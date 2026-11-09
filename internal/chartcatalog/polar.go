package chartcatalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"strings"
)

//go:embed polar.json
var polarJSON []byte

type PolarRecipe struct {
	ID, Title string
	Arrays    map[string][]map[string]any
	Colors    map[string]int
	Labels    map[string]string
	Margins   map[string]float32
	Height    float32
	Series    []string
}

func PolarRecipes() []PolarRecipe {
	var v []PolarRecipe
	if err := json.Unmarshal(polarJSON, &v); err != nil {
		panic(err)
	}
	return v
}
func (r PolarRecipe) rows() []map[string]any {
	if rows := r.Arrays["chartData"]; rows != nil {
		return rows
	}
	return r.Arrays["desktopData"]
}
func (r PolarRecipe) category() string {
	if _, ok := r.rows()[0]["browser"]; ok {
		return "browser"
	}
	return "month"
}
func PolarProps(c *ui.Context, r PolarRecipe, o Options) display.ChartProps {
	rows := r.rows()
	tokens := theme.TokensOf(c)
	palette := tokens.Charts()
	category := r.category()
	p := display.ChartProps{TooltipFollowInspection: true, Kind: display.PieChart, Label: r.Title, Height: r.Height, Selected: o.Selected, Animation: display.ChartAnimationProps{Disabled: !o.Animate}, Polar: &display.ChartPolarProps{SelectedSeries: o.SelectedSeries}}
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
		s := display.ChartSeries{Data: data, DataKey: key, Name: r.Labels[key], Color: palette[r.Colors[key]%len(palette)], Style: display.ChartSeriesStyle{Polar: &display.ChartPolarSeriesProps{}}}
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
			if row[category].(string) == o.Month {
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
		payload := d.Payload
		if p.Kind == display.PieChart || p.Kind == display.DonutChart || p.Kind == display.RadialChart {
			for i := range payload {
				payload[i].Name = d.Label
			}
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
	p.Data = rows
	return p
}

func center(c *ui.Context, value float64, label string, size float32) {
	ui.Column(c).AlignItems(ui.Center).Gap(2).Children(func() {
		ui.Text(c, message.NewPrinter(language.English).Sprintf("%.0f", value)).FontSize(size).FontWeight(700)
		ui.Text(c, label).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
	})
}
