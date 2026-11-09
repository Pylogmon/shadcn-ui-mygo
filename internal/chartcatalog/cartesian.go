package chartcatalog

import (
	_ "embed"
	"encoding/json"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"strings"
	"time"
)

//go:embed cartesian.json
var cartesianJSON []byte

type CartesianRecipe struct {
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

func CartesianRecipes() []CartesianRecipe {
	var v []CartesianRecipe
	if err := json.Unmarshal(cartesianJSON, &v); err != nil {
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

func CartesianProps(c *ui.Context, r CartesianRecipe, o Options) display.ChartProps {
	colors := theme.TokensOf(c).Charts()
	rows := r.Data
	if r.ID == "chart-area-interactive" {
		days := 90
		if o.Range == "30d" {
			days = 30
		} else if o.Range == "7d" {
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
	p := display.ChartProps{TooltipFollowInspection: true, Label: r.Title, Height: 300, Selected: o.Selected, Animation: display.ChartAnimationProps{Disabled: !o.Animate}}
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
		keys = []string{o.Series}
	}
	if len(r.Series) > 0 {
		keys = r.Series
	}
	for _, key := range keys {
		index := r.Colors[key]
		s := display.ChartSeries{DataKey: key, Name: title(key), Color: colors[index]}
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
		payload := d.Payload
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
	p.Data = rows
	return p
}

func roundedRect(r ui.Rect, radius float32) *ui.Path {
	k := float32(.55228475)
	rad := min(radius, r.W/2, r.H/2)
	x, y, w, h := r.X, r.Y, r.W, r.H
	return (&ui.Path{}).MoveTo(x+rad, y).LineTo(x+w-rad, y).CubeTo(x+w-rad+k*rad, y, x+w, y+rad-k*rad, x+w, y+rad).LineTo(x+w, y+h-rad).CubeTo(x+w, y+h-rad+k*rad, x+w-rad+k*rad, y+h, x+w-rad, y+h).LineTo(x+rad, y+h).CubeTo(x+rad-k*rad, y+h, x, y+h-rad+k*rad, x, y+h-rad).LineTo(x, y+rad).CubeTo(x, y+rad-k*rad, x+rad-k*rad, y, x+rad, y).Close()
}
