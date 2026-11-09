package chartcatalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

//go:embed additional.json
var additionalJSON []byte

type AdditionalRecipe struct {
	ID, Title, Family, Section, Source string
	Description                        string
	Data                               []map[string]any
}

func AdditionalRecipes() []AdditionalRecipe {
	var r []AdditionalRecipe
	if err := json.Unmarshal(additionalJSON, &r); err != nil {
		panic(err)
	}
	return r
}
func short(s string) string { r := []rune(s); return string(r[:min(3, len(r))]) }
func tooltipContent(c *ui.Context, d display.ChartTooltipData, p display.ChartTooltipContentProps) ui.Element {
	p.Active = true
	p.Payload = d.Payload
	if p.Label == nil {
		p.Label = &display.ChartContentLabel{Text: d.Label}
	}
	return display.ChartTooltipContent(c, p)
}
func TooltipProps(c *ui.Context, r AdditionalRecipe, o Options) display.ChartProps {
	colors := theme.TokensOf(c).Charts()
	p := display.ChartProps{TooltipFollowInspection: true, Kind: display.BarChart, Label: r.Title, Height: 280, Data: r.Data, Selected: o.Selected, Stacked: true, Animation: display.ChartAnimationProps{Disabled: !o.Animate}, Cartesian: display.ChartCartesianProps{HideGrid: true, HideValueAxis: true, HideCursor: true, CategoryTickMargin: ptr(float32(10)), BandPadding: &display.ChartBandPadding{Inner: .2, Outer: .1}}}
	for _, row := range r.Data {
		p.Labels = append(p.Labels, row["date"].(string))
	}
	p.Cartesian.FormatCategory = func(s string) string { d, _ := time.Parse("2006-01-02", s); return d.Format("Mon") }
	for i, key := range []string{"running", "swimming"} {
		series := display.ChartSeries{Name: title(key), DataKey: key, Color: colors[i]}
		for _, row := range r.Data {
			series.Values = append(series.Values, row[key].(float64))
		}
		if i == 0 {
			series.Style.BarRadius = &display.ChartCornerRadii{BottomLeft: 4, BottomRight: 4}
		} else {
			series.Style.BarRadius = &display.ChartCornerRadii{TopLeft: 4, TopRight: 4}
		}
		p.Series = append(p.Series, series)
	}
	p.TooltipContent = func(d display.ChartTooltipData) {
		cp := display.ChartTooltipContentProps{}
		switch r.ID {
		case "chart-tooltip-indicator-line":
			cp.Indicator = display.ChartIndicatorLine
		case "chart-tooltip-indicator-none":
			cp.HideIndicator = true
		case "chart-tooltip-label-none":
			cp.HideIndicator = true
			cp.HideLabel = true
		case "chart-tooltip-label-custom":
			cp.LabelKey = "activities"
			cp.Indicator = display.ChartIndicatorLine
		case "chart-tooltip-label-formatter":
			cp.LabelFormatter = func(c *ui.Context, label display.ChartContentLabel, _ []display.ChartPayloadItem) {
				date, _ := time.Parse("2006-01-02", label.Text)
				ui.Text(c, date.Format("January 2, 2006"))
			}
		case "chart-tooltip-icons":
			cp.HideLabel = true
		case "chart-tooltip-formatter", "chart-tooltip-advanced":
			cp.HideLabel = true
			cp.Formatter = func(c *ui.Context, f display.ChartTooltipFormatData) {
				advanced := r.ID == "chart-tooltip-advanced"
				tokens := theme.TokensOf(c)
				ui.Column(c).Grow(1).MinWidth(0).Gap(6).Children(func() {
					ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
						if advanced && f.Item.Color != nil {
							ui.Box(c).Size(10, 10).Shrink(0).Radius(2).Background(*f.Item.Color)
						}
						item, ok := display.ChartPayloadConfig(c, f.Item, fmt.Sprint(f.Name))
						label := fmt.Sprint(f.Name)
						if ok {
							label = item.Label
						}
						ui.Text(c, label).Grow(1).TextColor(tokens.MutedForeground)
						ui.Row(c).AlignItems(ui.Baseline).Gap(2).Children(func() {
							ui.Text(c, fmt.Sprint(f.Value)).Font("monospace").FontFeatures("tnum").FontWeight(500)
							ui.Text(c, "kcal").TextColor(tokens.MutedForeground)
						})
					})
					if advanced && f.Index == 1 {
						a, _ := f.Payload["running"].(float64)
						b, _ := f.Payload["swimming"].(float64)
						ui.Row(c).FillWidth().Padding(6, 0, 0, 0).DrawOver(func(p *ui.Painter, r ui.Rect) { p.Line(r.X, r.Y, r.X+r.W, r.Y, 1, tokens.Border) }).Children(func() {
							ui.Text(c, "Total").Grow(1).FontWeight(500)
							ui.Text(c, fmt.Sprintf("%g kcal", a+b)).Font("monospace").FontFeatures("tnum")
						})
					}
				})
			}
		}
		e := tooltipContent(c, d, cp)
		if r.ID == "chart-tooltip-formatter" {
			e.MinWidth(130)
		}
		if r.ID == "chart-tooltip-advanced" {
			e.Width(180)
		}
	}
	return p
}
func tooltipConfig(c *ui.Context, r AdditionalRecipe) display.ChartConfig {
	palette := theme.TokensOf(c).Charts()
	config := display.ChartConfig{"running": {Label: "Running", Color: &palette[0]}, "swimming": {Label: "Swimming", Color: &palette[1]}, "activities": {Label: "Activities"}}
	if r.ID == "chart-tooltip-icons" {
		a, b := config["running"], config["swimming"]
		a.Icon = func(c *ui.Context) ui.Element { return material.Icon(c, outlined.Footprint()) }
		b.Icon = func(c *ui.Context) ui.Element { return material.Icon(c, outlined.Waves()) }
		config["running"], config["swimming"] = a, b
	}
	return config
}

func DocsProps(c *ui.Context, r AdditionalRecipe, o Options, language string) (display.ChartProps, display.ChartConfig) {
	palette := []ui.Color{ui.RGB(37, 99, 235), ui.RGB(96, 165, 250)}
	labels := map[string]string{"desktop": "Desktop", "mobile": "Mobile"}
	p := display.ChartProps{TooltipFollowInspection: true, Kind: display.BarChart, Label: r.Title, Height: 250, Data: r.Data, Selected: o.Selected, Animation: display.ChartAnimationProps{Disabled: !o.Animate}, HideTooltip: r.ID == "docs-chart-example" || r.ID == "docs-chart-example-grid" || r.ID == "docs-chart-example-axis", Cartesian: display.ChartCartesianProps{HideGrid: r.ID == "docs-chart-example", HideValueAxis: true, HideCategoryAxis: r.ID == "docs-chart-example" || r.ID == "docs-chart-example-grid", CategoryTickMargin: ptr(float32(10)), BarRadius: ptr(float32(4)), BandPadding: &display.ChartBandPadding{Inner: .2, Outer: .1}, GroupGap: ptr(float32(4)), Cursor: &display.ChartCursorProps{Band: true}}}
	keys := []string{"desktop", "mobile"}
	category := "month"
	translate := func(s string) string { return s }
	if r.ID == "docs-chart-rtl" {
		colors := theme.TokensOf(c).Charts()
		palette = []ui.Color{colors[1], colors[0]}
		p.Cartesian.ReversedCategory = language != "en"
		months := map[string][]string{"en": {"January", "February", "March", "April", "May", "June"}, "ar": {"يناير", "فبراير", "مارس", "أبريل", "مايو", "يونيو"}, "he": {"ינואר", "פברואר", "מרץ", "אפריל", "מאי", "יוני"}}
		if language == "ar" {
			labels = map[string]string{"desktop": "سطح المكتب", "mobile": "الجوال"}
		} else if language == "he" {
			labels = map[string]string{"desktop": "מחשב", "mobile": "נייד"}
		}
		translate = func(s string) string {
			for i, key := range []string{"january", "february", "march", "april", "may", "june"} {
				if strings.EqualFold(key, s) {
					if list, ok := months[language]; ok {
						return list[i]
					}
					return months["en"][i]
				}
			}
			return s
		}
	}
	if r.ID == "docs-chart-demo" {
		colors := theme.TokensOf(c).Charts()
		palette = []ui.Color{colors[1], colors[0]}
		keys = []string{o.Series}
		category = "date"
		p.Cartesian.BarRadius = ptr(float32(0))
		p.Cartesian.Margin = &display.ChartInsets{Left: 12, Right: 12}
		p.Cartesian.CategoryTickMargin = ptr(float32(8))
		p.Cartesian.CategoryMinTickGap = ptr(float32(32))
	}
	for _, row := range r.Data {
		p.Labels = append(p.Labels, row[category].(string))
	}
	config := display.ChartConfig{"views": {Label: "Page Views"}}
	for _, key := range []string{"desktop", "mobile"} {
		index := 0
		if key == "mobile" {
			index = 1
		}
		col := palette[index]
		config[key] = display.ChartConfigItem{Label: labels[key], Color: &col}
	}
	for _, key := range keys {
		entry := config[key]
		series := display.ChartSeries{Name: entry.Label, DataKey: key, Color: *entry.Color}
		for _, row := range r.Data {
			series.Values = append(series.Values, row[key].(float64))
		}
		p.Series = append(p.Series, series)
	}
	p.Cartesian.FormatCategory = func(s string) string { return short(translate(s)) }
	if category == "date" {
		p.Cartesian.FormatCategory = func(s string) string { d, _ := time.Parse("2006-01-02", s); return d.Format("Jan 2") }
	}
	p.TooltipContent = func(d display.ChartTooltipData) {
		cp := display.ChartTooltipContentProps{Label: &display.ChartContentLabel{Text: translate(d.Label)}}
		if r.ID == "docs-chart-demo" {
			cp.NameKey = "views"
			cp.LabelFormatter = func(c *ui.Context, l display.ChartContentLabel, _ []display.ChartPayloadItem) {
				date, _ := time.Parse("2006-01-02", l.Text)
				ui.Text(c, date.Format("Jan 2, 2006"))
			}
		}
		e := tooltipContent(c, d, cp)
		if r.ID == "docs-chart-demo" {
			e.Width(150)
		}
	}
	return p, config
}

func CreateProps(c *ui.Context, r AdditionalRecipe, o Options) (display.ChartProps, display.ChartConfig) {
	palette := theme.TokensOf(c).Charts()
	p := display.ChartProps{TooltipFollowInspection: true, Kind: display.BarChart, Label: r.Title, Height: 250, Data: r.Data, Selected: o.Selected, Animation: display.ChartAnimationProps{Disabled: !o.Animate}, Cartesian: display.ChartCartesianProps{HideValueAxis: true, HideCursor: true, CategoryTickMargin: ptr(float32(8)), FormatCategory: short, BarRadius: ptr(float32(4)), BandPadding: &display.ChartBandPadding{Inner: .2, Outer: .1}, GroupGap: ptr(float32(4))}}
	keys := []string{"desktop", "mobile"}
	category := "month"
	if r.Family == "area" {
		keys = []string{"desktop"}
		p.Kind = display.AreaChart
		p.Cartesian.PointScale = true
		p.Cartesian.Curve = display.ChartCurveNatural
	}
	if r.Family == "line" {
		p.Kind = display.LineChart
		p.Cartesian.PointScale = true
		p.Cartesian.Curve = display.ChartCurveMonotone
	}
	if r.Family == "area" || r.Family == "line" {
		p.Cartesian.Margin = &display.ChartInsets{Left: 12, Right: 12}
	}
	if r.Family == "bar" {
		p.Cartesian.HideCursor = false
		p.Cartesian.Cursor = &display.ChartCursorProps{Band: true}
		p.Cartesian.CategoryTickMargin = ptr(float32(10))
	}
	if r.Family == "pie" || r.Family == "radial" {
		keys = []string{"visitors"}
		category = "browser"
		p.Polar = &display.ChartPolarProps{SelectedSeries: o.SelectedSeries}
	}
	if r.Family == "radar" {
		p.Kind = display.RadarChart
		p.Polar = &display.ChartPolarProps{AngleAxis: &display.ChartPolarAngleAxisProps{}, Grid: &display.ChartPolarGridProps{}}
	}
	config := display.ChartConfig{}
	for _, row := range r.Data {
		label := row[category].(string)
		p.Labels = append(p.Labels, label)
	}
	for j, key := range keys {
		s := display.ChartSeries{Name: title(key), DataKey: key, Color: palette[j]}
		for _, row := range r.Data {
			s.Values = append(s.Values, row[key].(float64))
		}
		if p.Kind == display.AreaChart {
			s.Style.FillOpacity = ptr(float32(.4))
		}
		if p.Kind == display.RadarChart && j == 0 {
			s.Style.FillOpacity = ptr(float32(.6))
		}
		config[key] = display.ChartConfigItem{Label: s.Name, Color: &s.Color}
		p.Series = append(p.Series, s)
	}
	for i, label := range p.Labels {
		if category == "browser" {
			col := palette[i%len(palette)]
			if r.Family == "radial" {
				col = palette[1]
			}
			p.CategoryColors = append(p.CategoryColors, col)
			config[label] = display.ChartConfigItem{Label: title(label), Color: &col}
		}
	}
	if r.Family == "pie" {
		p.Kind = display.DonutChart
		p.Polar.InnerRadius = ptr(display.ChartRadiusDIP(60))
		p.Polar.StrokeWidth = ptr(float32(5))
		p.Polar.Center = func(c *ui.Context, d display.ChartPolarCenter) { center(c, d.Total, "Visitors", 30) }
	}
	if r.Family == "radial" {
		p.Kind = display.RadialChart
		p.Height = 210
		p.HideTooltip = true
		p.Polar.EndAngle = ptr(float64(100))
		p.Polar.InnerRadius = ptr(display.ChartRadiusDIP(64))
		p.Polar.OuterRadius = ptr(display.ChartRadiusDIP(94))
		p.Polar.Background = true
		transparent := ui.Transparent
		p.Polar.Grid = &display.ChartPolarGridProps{Kind: display.ChartPolarCircle, RadialLines: ptr(false), Stroke: &transparent, Radii: []display.ChartRadius{display.ChartRadiusDIP(86), display.ChartRadiusDIP(74)}, Fills: []ui.Color{theme.TokensOf(c).Muted, theme.TokensOf(c).Background}}
		p.Polar.Center = func(c *ui.Context, d display.ChartPolarCenter) { center(c, d.Total, "Visitors", 36) }
	}
	p.TooltipContent = func(d display.ChartTooltipData) {
		cp := display.ChartTooltipContentProps{HideLabel: r.Family == "pie"}
		if r.Family == "area" || r.Family == "radar" {
			cp.Indicator = display.ChartIndicatorLine
		}
		if r.Family == "bar" {
			cp.Indicator = display.ChartIndicatorDashed
		}
		if r.Family == "pie" {
			for i := range d.Payload {
				d.Payload[i].Name = d.Label
			}
		}
		tooltipContent(c, d, cp)
	}
	return p, config
}
