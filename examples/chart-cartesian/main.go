// The 30 fixed shadcn Cartesian examples, using only public kit/native APIs.
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
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

//go:embed recipes.json
var recipeJSON []byte

type recipe struct {
	ID, Title   string
	Data        []map[string]any
	Margins     map[string]float32
	Series      []string
	Colors      map[string]int
	Labels      map[string]string
	CursorFalse bool
	Tooltip     struct {
		Indicator, NameKey                            string
		HideLabel, HideIndicator, DateLabel, DateYear bool
		Width                                         float32
	}
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
	r := a.recipe()
	colors := theme.TokensOf(c).Charts()
	rows := r.Data
	if r.ID == "chart-area-interactive" {
		days := 90
		if a.rangeID == "30d" {
			days = 30
		} else if a.rangeID == "7d" {
			days = 7
		}
		end, _ := time.Parse("2006-01-02", "2024-06-30")
		cutoff := end.AddDate(0, 0, -days)
		rows = nil
		for _, row := range r.Data {
			when, _ := time.Parse("2006-01-02", row["date"].(string))
			if !when.Before(cutoff) {
				rows = append(rows, row)
			}
		}
	}
	p := display.ChartProps{Label: r.Title, Height: 300, Selected: &a.selected, Animation: display.ChartAnimationProps{Disabled: !a.animate}}
	p.Cartesian = display.ChartCartesianProps{HideValueAxis: true, HideCursor: true, CategoryTickMargin: ptr(float32(10)), BandPadding: &display.ChartBandPadding{Inner: .2, Outer: .1}, GroupGap: ptr(float32(4))}
	p.Cartesian.Margin = &display.ChartInsets{Top: r.Margins["top"], Right: r.Margins["right"], Bottom: r.Margins["bottom"], Left: r.Margins["left"]}
	p.Cartesian.FormatCategory = func(s string) string {
		if len(s) > 3 {
			return s[:3]
		}
		return s
	}
	category := "month"
	if _, ok := rows[0]["browser"]; ok {
		category = "browser"
		p.Cartesian.FormatCategory = title
	}
	if _, ok := rows[0]["date"]; ok {
		category = "date"
		p.Cartesian.FormatCategory = func(s string) string { t, _ := time.Parse("2006-01-02", s); return t.Format("Jan 2") }
	}
	for _, row := range rows {
		p.Labels = append(p.Labels, row[category].(string))
	}
	keys := []string{"desktop"}
	if _, ok := rows[0]["visitors"]; ok {
		keys = []string{"visitors"}
	}
	if strings.HasPrefix(r.ID, "chart-area-") {
		p.Kind = display.AreaChart
		p.Cartesian.PointScale = true
		p.Cartesian.Curve = display.ChartCurveNatural
		p.Cartesian.CategoryTickMargin = ptr(float32(8))
	}
	if strings.HasPrefix(r.ID, "chart-line-") {
		p.Kind = display.LineChart
		p.Cartesian.PointScale = true
		p.Cartesian.Curve = display.ChartCurveNatural
		p.Cartesian.CategoryTickMargin = ptr(float32(8))
	}
	p.Cartesian.HideCursor = r.CursorFalse
	if p.Kind == display.BarChart && !r.CursorFalse {
		p.Cartesian.Cursor = &display.ChartCursorProps{Band: true}
	}
	switch r.ID {
	case "chart-area-stacked", "chart-area-icons", "chart-area-legend", "chart-area-axes", "chart-area-gradient", "chart-area-interactive":
		keys = []string{"mobile", "desktop"}
		p.Stacked = true
	case "chart-area-stacked-expand":
		keys = []string{"other", "mobile", "desktop"}
		p.Stacked, p.Normalized = true, true
	case "chart-bar-multiple", "chart-bar-stacked", "chart-line-multiple":
		keys = []string{"desktop", "mobile"}
		p.Stacked = r.ID == "chart-bar-stacked"
	case "chart-bar-interactive", "chart-line-interactive":
		keys = []string{a.seriesID}
	}
	if len(r.Series) > 0 {
		keys = r.Series
	}
	for _, key := range keys {
		index := r.Colors[key]
		s := display.ChartSeries{Name: title(key), Color: colors[index]}
		for _, row := range rows {
			v, _ := row[key].(float64)
			s.Values = append(s.Values, v)
		}
		if p.Kind == display.AreaChart {
			s.Style.FillOpacity = ptr(float32(.4))
		}
		p.Series = append(p.Series, s)
	}
	switch r.ID {
	case "chart-area-linear", "chart-line-linear":
		p.Cartesian.Curve = display.ChartCurveLinear
	case "chart-area-step", "chart-line-step":
		p.Cartesian.Curve = display.ChartCurveStep
	case "chart-line-multiple", "chart-line-interactive":
		p.Cartesian.Curve = display.ChartCurveMonotone
	case "chart-area-axes":
		p.Cartesian.HideValueAxis = false
		p.Cartesian.ValueTickCount = 3
	}
	if r.ID == "chart-area-gradient" || r.ID == "chart-area-interactive" {
		for j := range p.Series {
			col := p.Series[j].Color
			p.Series[j].Style.Gradient = &ui.LinearGradient{Angle: 180, Stops: []ui.GradientStop{{Color: col.Alpha(.8), At: .05}, {Color: col.Alpha(.1), At: .95}}}
			if r.ID == "chart-area-interactive" {
				p.Series[j].Style.FillOpacity = ptr(float32(1))
			}
		}
	}
	if r.ID == "chart-area-stacked-expand" {
		p.Series[0].Style.FillOpacity = ptr(float32(.1))
	}
	if strings.HasSuffix(r.ID, "-interactive") {
		p.Cartesian.CategoryMinTickGap = ptr(float32(32))
		p.Cartesian.CategoryTickMargin = ptr(float32(8))
	}
	if r.ID == "chart-area-legend" || r.ID == "chart-bar-stacked" {
		p.Legend = true
	}
	if r.ID == "chart-bar-horizontal" || r.ID == "chart-bar-mixed" || r.ID == "chart-bar-label-custom" {
		p.Orientation = display.ChartHorizontal
		p.Cartesian.HideGrid = r.ID != "chart-bar-label-custom"
	}
	if p.Kind == display.BarChart {
		p.Cartesian.BarRadius = ptr(float32(4))
	}
	if r.ID == "chart-bar-horizontal" || r.ID == "chart-bar-mixed" {
		p.Cartesian.BarRadius = ptr(float32(5))
	}
	if r.ID == "chart-bar-default" || r.ID == "chart-bar-label" {
		p.Cartesian.BarRadius = ptr(float32(8))
	}
	if r.ID == "chart-bar-interactive" {
		p.Cartesian.BarRadius = ptr(float32(0))
	}
	if r.ID == "chart-bar-stacked" {
		p.Series[0].Style.BarRadius = &display.ChartCornerRadii{BottomLeft: 4, BottomRight: 4}
		p.Series[1].Style.BarRadius = &display.ChartCornerRadii{TopLeft: 4, TopRight: 4}
	}
	if r.ID == "chart-bar-mixed" || r.ID == "chart-bar-active" || r.ID == "chart-line-dots-colors" || r.ID == "chart-line-label-custom" {
		for i := range rows {
			p.CategoryColors = append(p.CategoryColors, colors[i%len(colors)])
		}
	}
	if r.ID == "chart-bar-active" {
		p.Cartesian.BarRadius = ptr(float32(8))
		p.Series[0].Style.BarContent = func(_ *ui.Context, d display.ChartDatum, e ui.Element) {
			e.Opacity(1)
			if d.Index != 2 {
				return
			}
			col := d.Color
			e.Background(col.Alpha(.8)).DrawOver(func(painter *ui.Painter, r ui.Rect) { painter.StrokePath(roundedRect(r, 8).Dash(4, 4, 4), 2, col) })
		}
	}
	if r.ID == "chart-bar-negative" {
		p.Cartesian.HideCategoryAxis = true
		p.Cartesian.BarRadius = ptr(float32(0))
		p.Cartesian.FormatCategory = nil
		for _, v := range p.Series[0].Values {
			col := colors[0]
			if v < 0 {
				col = colors[1]
			}
			p.CategoryColors = append(p.CategoryColors, col)
		}
	}
	for j := range p.Series {
		s := &p.Series[j]
		if p.Kind == display.LineChart {
			s.Style.ActiveDot = &display.ChartDotProps{Radius: 4, Stroke: ptr(theme.TokensOf(c).Background), StrokeWidth: ptr(float32(2))}
			if r.ID != "chart-line-default" && r.ID != "chart-line-linear" && r.ID != "chart-line-step" && r.ID != "chart-line-multiple" && r.ID != "chart-line-interactive" {
				s.Style.Dots = &display.ChartDotProps{Radius: 3, Fill: &s.Color, StrokeWidth: ptr(float32(2))}
				s.Style.ActiveDot.Radius = 6
			}
		}
	}
	if r.ID == "chart-line-dots-colors" {
		p.Cartesian.HideCategoryAxis = true
		p.Series[0].Style.Dots = &display.ChartDotProps{Radius: 5, StrokeWidth: ptr(float32(2)), Content: func(_ *ui.Context, d display.ChartDatum, e ui.Element) {
			col := colors[d.Index%len(colors)]
			e.Background(col).Border(1, col)
		}}
	}
	if r.ID == "chart-line-dots-custom" {
		p.Series[0].Style.Dots = &display.ChartDotProps{Radius: 12, StrokeWidth: ptr(float32(0)), Content: func(_ *ui.Context, d display.ChartDatum, e ui.Element) {
			col := d.Color
			e.Background(ui.Transparent).Border(0, ui.Transparent).Draw(func(painter *ui.Painter, r ui.Rect) {
				x, y := r.X+r.W/2, r.Y+r.H/2
				painter.Line(x, r.Y, x, y-4, 2, col)
				painter.Line(x, y+4, x, r.Y+r.H, 2, col)
				painter.StrokePath((&ui.Path{}).Circle(x, y, 4), 2, col)
			})
		}}
	}
	if r.ID == "chart-line-label-custom" {
		p.Cartesian.HideCategoryAxis = true
	}
	if r.ID == "chart-bar-label" || r.ID == "chart-line-label" || r.ID == "chart-line-label-custom" || r.ID == "chart-bar-negative" {
		label := display.ChartLabelProps{Offset: ptr(float32(12))}
		if r.ID == "chart-line-label-custom" || r.ID == "chart-bar-negative" {
			label.Formatter = func(d display.ChartDatum) string { return title(d.Label) }
		}
		if r.ID == "chart-bar-negative" {
			label.Offset = ptr(float32(5))
		}
		p.Series[0].Style.Labels = []display.ChartLabelProps{label}
	}
	if r.ID == "chart-bar-label-custom" {
		p.Cartesian.HideCategoryAxis = true
		p.Series[0].Style.Labels = []display.ChartLabelProps{{Position: display.ChartLabelInsideLeft, Offset: ptr(float32(8)), Color: ptr(theme.TokensOf(c).Background), Formatter: func(d display.ChartDatum) string { return d.Label }}, {Position: display.ChartLabelRight, Offset: ptr(float32(8))}}
	}
	// Content knobs belong to the existing source-aware content helper, rather
	// than a second private tooltip renderer in the example.
	indicator := display.ChartIndicatorDot
	if r.Tooltip.Indicator == "line" {
		indicator = display.ChartIndicatorLine
	}
	if r.Tooltip.Indicator == "dashed" {
		indicator = display.ChartIndicatorDashed
	}
	p.TooltipContent = func(d display.ChartTooltipData) {
		payload := make([]display.ChartPayloadItem, len(d.Items))
		for i, item := range d.Items {
			col := item.Color
			payload[i] = display.ChartPayloadItem{Name: item.Label, DataKey: strings.ToLower(item.Label), Value: item.Value, Color: &col}
		}
		label := d.Label
		if r.Tooltip.DateLabel {
			when, _ := time.Parse("2006-01-02", label)
			format := "Jan 2"
			if r.Tooltip.DateYear {
				format = "Jan 2, 2006"
			}
			label = when.Format(format)
		}
		content := display.ChartTooltipContent(c, display.ChartTooltipContentProps{Active: true, Label: &display.ChartContentLabel{Text: label}, Payload: payload, Indicator: indicator, HideLabel: r.Tooltip.HideLabel, HideIndicator: r.Tooltip.HideIndicator, NameKey: r.Tooltip.NameKey})
		if r.Tooltip.Width > 0 {
			content.Width(r.Tooltip.Width)
		}
	}
	return p
}

func roundedRect(r ui.Rect, radius float32) *ui.Path {
	k := float32(.55228475)
	rad := min(radius, r.W/2, r.H/2)
	x, y, w, h := r.X, r.Y, r.W, r.H
	return (&ui.Path{}).MoveTo(x+rad, y).LineTo(x+w-rad, y).CubeTo(x+w-rad+k*rad, y, x+w, y+rad-k*rad, x+w, y+rad).LineTo(x+w, y+h-rad).CubeTo(x+w, y+h-rad+k*rad, x+w-rad+k*rad, y+h, x+w-rad, y+h).LineTo(x+rad, y+h).CubeTo(x+rad-k*rad, y+h, x, y+h-rad+k*rad, x, y+h-rad).LineTo(x, y+rad).CubeTo(x, y+rad-k*rad, x+rad-k*rad, y, x+rad, y).Close()
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
