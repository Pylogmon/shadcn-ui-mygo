package display

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type ChartKind int

const (
	BarChart ChartKind = iota
	LineChart
	AreaChart
	PieChart
	DonutChart
	RadarChart
	RadialChart
)

// ChartOrientation chooses the direction in which BarChart values grow.
// The zero value preserves vertical bars. Other chart kinds ignore it.
type ChartOrientation uint8

const (
	ChartVertical ChartOrientation = iota
	ChartHorizontal
)

// ChartCartesianProps configures Cartesian axes and marks. Zero values keep
// existing axes, grid, cursor and radius. It has no effect on polar charts.
type ChartCartesianProps struct {
	HideValueAxis, HideCategoryAxis, HideGrid, HideCursor bool
	// FormatCategory changes axis text only; inspection keeps original labels.
	FormatCategory func(string) string
	// FormatValue formats numeric axis ticks (fractions for normalized stacks).
	// Nil falls back to ChartProps.FormatValue; tooltips always keep raw data.
	FormatValue func(float64) string
	// CategoryAxisMaxWidth bounds category text. Horizontal bars default to 60
	// DIPs; vertical charts have no added cap. Invalid widths use that default.
	CategoryAxisMaxWidth float32
	// BarRadius nil or invalid (negative/nonfinite) uses 3 DIPs; 0 is square.
	BarRadius *float32
}

type ChartSeries struct {
	Name   string
	Values []float64
	Color  ui.Color
}

type ChartProps struct {
	Labels      []string
	Series      []ChartSeries
	Kind        ChartKind
	Orientation ChartOrientation
	Cartesian   ChartCartesianProps
	Height      float32
	Label       string
	// Stacked uses diverging stacks for bars/areas and concentric rings for
	// pies/donuts. Normalized uses shares of each category's total absolute
	// magnitude, preserving diverging positive/negative sides.
	Stacked, Normalized bool
	// Legend adds semantic series/category keys below the plot.
	Legend bool
	// Selected is a caller-owned category index; -1 means no inspected item.
	// Without it the chart keeps its own local selection, initially -1.
	Selected *int
	// HideTooltip hides the visual tooltip without disabling keyboard inspection.
	HideTooltip    bool
	FormatValue    func(float64) string
	TooltipContent func(ChartTooltipData)
	// InnerRadius is the donut hole fraction (default .6), clamped to 0.. .9.
	InnerRadius float32
	// MaxValue is the radial/radar domain maximum; zero derives it from data.
	MaxValue float64
	// CategoryColors overrides fills in pie/donut/unstacked radial plots and
	// single-series bars. That bar mode also uses category legend/tooltip colors.
	CategoryColors []ui.Color
}

type ChartLegendItem struct {
	Label string
	Color ui.Color
}
type ChartTooltipItem struct {
	Label string
	Value float64
	Color ui.Color
}
type ChartTooltipData struct {
	Label       string
	Items       []ChartTooltipItem
	FormatValue func(float64) string
}

// ChartLegend exposes the same human-readable labels used by the marks and
// tooltip. It may also be composed separately from a plot.
func ChartLegend(c *ui.Context, items []ChartLegendItem) ui.Element {
	t := theme.TokensOf(c)
	return ui.Row(c).FillWidth().Wrap().Justify(ui.Center).Gap(16).Role(ui.RoleGroup).Label("Chart legend").Children(func() {
		for _, item := range items {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(9, 9).Radius(2).Background(item.Color).AccessHidden(true)
				ui.Text(c, item.Label).FontSize(12).TextColor(t.MutedForeground)
			})
		}
	})
}

// ChartTooltip provides the shadcn surface and series indicators. It can be
// used by TooltipContent to compose custom values without replacing selection.
func ChartTooltip(c *ui.Context, data ChartTooltipData) ui.Element {
	t := theme.TokensOf(c)
	format := data.FormatValue
	if format == nil {
		format = chartFormat
	}
	return ui.Column(c).Gap(6).MinWidth(140).MaxWidth(280).Padding(10).Radius(t.RadiusMD()).
		Background(t.Popover).TextColor(t.PopoverForeground).Border(1, t.Border).
		Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Children(func() {
		ui.Text(c, data.Label).FontSize(12).FontWeight(500)
		for _, item := range data.Items {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(8, 8).Radius(2).Background(item.Color).AccessHidden(true)
				ui.Text(c, item.Label).Grow(1).FontSize(12).TextColor(t.MutedForeground)
				ui.Text(c, format(item.Value)).FontSize(12).FontWeight(500)
			})
		}
	})
}

type chartInspection struct {
	hovering bool
	x, y     float32
	last     int
}

// Chart renders native GPU paths/marks, with one keyboard stop for inspection.
// Arrows/Home/End move between valid categories and Escape clears selection.
// Pie, donut and radial plots use non-negative finite values; missing Cartesian
// observations break a line rather than inventing a measurement.
func Chart(c *ui.Context, p ChartProps) ui.Element {
	root := ui.Column(c).FillWidth().Gap(12)
	selected := p.Selected
	if selected == nil {
		selected = ui.Local(root, "chart-selection", func() int { return -1 })
	}
	if *selected < -1 || *selected >= len(p.Labels) {
		*selected = -1
	}
	height := p.Height
	if height <= 0 || !finite(float64(height)) {
		height = 220
	}
	return root.Children(func() {
		switch p.Kind {
		case PieChart, DonutChart, RadialChart, RadarChart:
			polarChart(c, p, selected, height)
		default:
			cartesianChart(c, p, selected, height)
		}
		if p.Legend {
			ChartLegend(c, chartLegendItems(c, p))
		}
	})
}

func chartFormat(v float64) string { return fmt.Sprintf("%g", v) }
func finite(v float64) bool        { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func chartColor(c *ui.Context, p ChartProps, series int) ui.Color {
	if series >= 0 && series < len(p.Series) && p.Series[series].Color != (ui.Color{}) {
		return p.Series[series].Color
	}
	colors := theme.TokensOf(c).Charts()
	return colors[max(0, series)%len(colors)]
}
func seriesName(p ChartProps, i int) string {
	if p.Series[i].Name != "" {
		return p.Series[i].Name
	}
	return fmt.Sprintf("Series %d", i+1)
}
func chartLegendItems(c *ui.Context, p ChartProps) []ChartLegendItem {
	var items []ChartLegendItem
	if p.Kind == PieChart || p.Kind == DonutChart || p.Kind == RadialChart && !p.Stacked || chartCategoryFills(p) {
		for i, label := range p.Labels {
			items = append(items, ChartLegendItem{label, categoryColor(c, p, i)})
		}
	} else {
		for i := range p.Series {
			items = append(items, ChartLegendItem{seriesName(p, i), chartColor(c, p, i)})
		}
	}
	return items
}
func categoryColor(c *ui.Context, p ChartProps, index int) ui.Color {
	if index >= 0 && index < len(p.CategoryColors) && p.CategoryColors[index] != (ui.Color{}) {
		return p.CategoryColors[index]
	}
	return chartColor(c, ChartProps{}, index)
}
func tooltipData(c *ui.Context, p ChartProps, index int) ChartTooltipData {
	d := ChartTooltipData{FormatValue: p.FormatValue}
	if index < 0 || index >= len(p.Labels) {
		return d
	}
	d.Label = p.Labels[index]
	for j, s := range p.Series {
		if index >= len(s.Values) || !finite(s.Values[index]) {
			continue
		}
		color := chartColor(c, p, j)
		if chartCategoryFills(p) {
			color = categoryColor(c, p, index)
		}
		if p.Kind == PieChart || p.Kind == DonutChart || p.Kind == RadialChart || p.Kind == RadarChart {
			if s.Values[index] < 0 {
				continue
			}
			if p.Kind != RadialChart || !p.Stacked {
				color = categoryColor(c, p, index)
			}
		}
		d.Items = append(d.Items, ChartTooltipItem{seriesName(p, j), s.Values[index], color})
	}
	return d
}
func describeChart(c *ui.Context, p ChartProps, index int) string {
	d := tooltipData(c, p, index)
	format := d.FormatValue
	if format == nil {
		format = chartFormat
	}
	parts := []string{d.Label}
	for _, v := range d.Items {
		parts = append(parts, v.Label+": "+format(v.Value))
	}
	return strings.Join(parts, ", ")
}
func chartTooltipAt(c *ui.Context, p ChartProps, index int) {
	if p.HideTooltip || index < 0 {
		return
	}
	d := tooltipData(c, p, index)
	if len(d.Items) == 0 {
		return
	}
	ui.Column(c).Absolute().Top(8).Left(8).ZIndex(10).IgnoreInput(true).AccessHidden(true).Children(func() {
		if p.TooltipContent != nil {
			p.TooltipContent(d)
		} else {
			ChartTooltip(c, d)
		}
	})
}

func validCategories(p ChartProps, positive bool) []int {
	var out []int
	for i := range p.Labels {
		for j, s := range p.Series {
			if positive && !p.Stacked && j > 0 && p.Kind != RadarChart {
				break
			}
			if i < len(s.Values) && finite(s.Values[i]) && (!positive || s.Values[i] > 0) {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

// stackSegments sanitizes non-finite values before passing them to the native
// stacking primitive; Infinity must not poison every segment in that category.
func stackSegments(p ChartProps) [][]chart.Segment {
	values := make([][]float64, len(p.Series))
	for j, s := range p.Series {
		values[j] = make([]float64, len(p.Labels))
		for i := range p.Labels {
			if i < len(s.Values) && finite(s.Values[i]) {
				values[j][i] = s.Values[i]
			}
		}
	}
	how := chart.StackStandard
	if p.Normalized {
		how = chart.StackNormalized
	}
	return chart.Stack(values, how)
}

func cartesianChart(c *ui.Context, p ChartProps, selected *int, height float32) ui.Element {
	t := theme.TokensOf(c)
	plot, scale := scaledChartValues(p)
	hi, lo := float64(0), float64(0)
	var stacked [][]chart.Segment
	if p.Stacked && p.Kind != LineChart {
		stacked = stackSegments(plot)
	}
	for j, s := range plot.Series {
		for i, v := range s.Values {
			if i >= len(p.Labels) {
				break
			}
			if finite(v) {
				if len(stacked) > 0 {
					lo = min(lo, stacked[j][i].Lo)
					hi = max(hi, stacked[j][i].Hi)
				} else {
					hi = max(hi, v)
					lo = min(lo, v)
				}
			}
		}
	}
	x := chart.NewBand(p.Labels...).Padding(.3, .15)
	y := cartesianValueScale(p, scale, lo, hi)
	direction := c.Direction()
	axes := cartesianAxes(c, p, x, y, direction)
	horizontal := horizontalBars(p)
	valid := validCategories(p, false)
	return chart.Plot(c, axes, func(area ui.Element) {
		label := p.Label
		if label == "" {
			label = "Chart"
		}
		description := "Use Left and Right arrows to inspect values; Home and End move to the first and last category; Escape clears inspection."
		if horizontal {
			description = "Use Up and Down arrows to inspect values; Home and End move to the first and last category; Escape clears inspection."
		}
		area.Label(label).Role(ui.RoleGroup).Description(description).Direction(direction)
		inspectCartesian(c, area, p, selected, x, valid)
		drawers := []func(*ui.Painter, ui.Rect){}
		if !p.Cartesian.HideGrid {
			gridScale := y
			gridScale.Format = nil
			gridTicks := gridScale.Ticks(4)
			drawers = append(drawers, func(painter *ui.Painter, r ui.Rect) {
				for _, tick := range gridTicks {
					if horizontal {
						painter.Line(chart.X(r, tick.At), r.Y, chart.X(r, tick.At), r.Y+r.H, 1, t.Border)
					} else {
						painter.Line(r.X, chart.Y(r, tick.At), r.X+r.W, chart.Y(r, tick.At), 1, t.Border)
					}
				}
			})
		}

		for j, s := range plot.Series {
			color := chartColor(c, p, j)
			n := min(len(s.Values), len(p.Labels))
			for i := range n {
				if !finite(s.Values[i]) {
					continue
				}
				band := x.Band(i)
				if p.Kind == BarChart {
					from, to := 0.0, s.Values[i]
					if len(stacked) > 0 {
						from, to = stacked[j][i].Lo, stacked[j][i].Hi
					} else {
						width := (band.Hi - band.Lo) / float32(max(len(p.Series), 1))
						band.Lo += float32(j) * width
						band.Hi = band.Lo + width
					}
					xSpan, ySpan := band, y.Span(from, to)
					if horizontal {
						xSpan, ySpan = ySpan, horizontalBarBand(band)
					}
					mark := motion.ApplyTransition(c, chart.Rect(c.Key(chartBarKey{j, i}), xSpan, ySpan).Background(barColor(c, p, j, i)).Radius(chartBarRadius(p)).Label(fmt.Sprintf("%s %s: %g", seriesName(p, j), p.Labels[i], p.Series[j].Values[i])))
					if *selected == i {
						mark.Opacity(.8)
					}
				}
			}
			if p.Kind != BarChart {
				segments := finiteSegments(s.Values[:n], x, y)
				if len(stacked) > 0 {
					segments = stackedSegments(s.Values[:n], stacked[j], x, y)
				}
				isArea, zeroBaseline := p.Kind == AreaChart, y.At(0)
				drawers = append(drawers, func(painter *ui.Painter, r ui.Rect) {
					for _, segment := range segments {
						if isArea {
							base := segment.Base
							if len(base) == 0 {
								base = make([]float32, len(segment.X))
								for i := range base {
									base[i] = zeroBaseline
								}
							}
							painter.FillPath(chart.Area(r, segment.X, base, segment.Y, chart.CurveMonotone), color.Alpha(.2))
						}
						painter.StrokePath(chart.Line(r, segment.X, segment.Y, chart.CurveMonotone), 2, color)
					}
				})
				if *selected >= 0 && *selected < n && finite(s.Values[*selected]) {
					v := s.Values[*selected]
					if len(stacked) > 0 {
						v = stacked[j][*selected].Hi
					}
					chart.Point(c, x.At(*selected), y.At(v), 7, 7).Radius(4).Background(color).Border(2, t.Background).IgnoreInput(true)
				}
			}
		}
		area.Draw(func(painter *ui.Painter, r ui.Rect) {
			for _, draw := range drawers {
				draw(painter, r)
			}
		})
		if *selected >= 0 && !p.Cartesian.HideCursor {
			if horizontal {
				chart.RuleY(c, 1-x.At(*selected), 1).Background(t.Border).IgnoreInput(true)
			} else {
				chart.RuleX(c, x.At(*selected), 1).Background(t.Border).IgnoreInput(true)
			}
		}
		chartTooltipAt(c, p, *selected)
	}).Height(height).FillWidth().MinWidth(0).Gap(8).FontSize(12).Direction(ui.LeftToRight)
}

type plotSegment struct{ X, Y, Base []float32 }

// Non-finite measurements are missing observations, so a line must break at
// them rather than connect the observations on either side with invented data.
func finiteSegments(values []float64, x chart.Band, y chart.Linear) []plotSegment {
	var out []plotSegment
	var segment plotSegment
	flush := func() {
		if len(segment.X) > 0 {
			out = append(out, segment)
			segment = plotSegment{}
		}
	}
	for i, v := range values {
		if !finite(v) {
			flush()
			continue
		}
		segment.X = append(segment.X, x.At(i))
		segment.Y = append(segment.Y, y.At(v))
	}
	flush()
	return out
}
func stackedSegments(values []float64, stack []chart.Segment, x chart.Band, y chart.Linear) []plotSegment {
	var out []plotSegment
	var segment plotSegment
	flush := func() {
		if len(segment.X) > 0 {
			out = append(out, segment)
			segment = plotSegment{}
		}
	}
	for i, v := range values {
		if !finite(v) {
			flush()
			continue
		}
		segment.X = append(segment.X, x.At(i))
		segment.Y = append(segment.Y, y.At(stack[i].Hi))
		segment.Base = append(segment.Base, y.At(stack[i].Lo))
	}
	flush()
	return out
}

func polarGeometry(r ui.Rect) (float32, float32, float32) {
	return r.X + r.W/2, r.Y + r.H/2, max(0, min(r.W, r.H)/2-22)
}
func positiveValues(p ChartProps, series int) []float64 {
	values := make([]float64, len(p.Labels))
	if series >= len(p.Series) {
		return values
	}
	for i := range values {
		if i < len(p.Series[series].Values) && finite(p.Series[series].Values[i]) && p.Series[series].Values[i] > 0 {
			values[i] = p.Series[series].Values[i]
		}
	}
	return values
}

// Scale only geometry for extreme magnitudes. Tooltips retain caller values;
// axis labels undo this factor. This avoids overflowing a finite domain/stack.
func scaledChartValues(p ChartProps) (ChartProps, float64) {
	maxValue := 0.0
	for _, series := range p.Series {
		for i, v := range series.Values {
			if i < len(p.Labels) && finite(v) {
				maxValue = max(maxValue, math.Abs(v))
			}
		}
	}
	if maxValue == 0 || maxValue >= 1e-100 && maxValue <= 1e100 {
		return p, 1
	}
	q := p
	q.Series = make([]ChartSeries, len(p.Series))
	for j, series := range p.Series {
		q.Series[j] = series
		q.Series[j].Values = make([]float64, len(series.Values))
		for i, v := range series.Values {
			q.Series[j].Values[i] = v / maxValue
		}
	}
	if finite(q.MaxValue) && q.MaxValue > 0 {
		q.MaxValue /= maxValue
	}
	return q, maxValue
}
func pieSlices(values []float64) []chart.Slice {
	largest := 0.0
	for _, value := range values {
		if finite(value) {
			largest = max(largest, value)
		}
	}
	scaled := make([]float64, len(values))
	if largest > 0 {
		for i, value := range values {
			if finite(value) && value > 0 {
				scaled[i] = value / largest
			}
		}
	}
	return chart.Pie(scaled, .025)
}
func polarMaximum(p ChartProps) float64 {
	if finite(p.MaxValue) && p.MaxValue > 0 {
		return p.MaxValue
	}
	hi := 1.0
	for _, s := range p.Series {
		for i, v := range s.Values {
			if i < len(p.Labels) && finite(v) {
				hi = max(hi, v)
			}
		}
	}
	return hi
}

func polarChart(c *ui.Context, p ChartProps, selected *int, height float32) ui.Element {
	t := theme.TokensOf(c)
	label := p.Label
	if label == "" {
		label = "Chart"
	}
	area := ui.Box(c).Width(height).MaxWidthPercent(100).AspectRatio(1).AlignSelf(ui.Center).Label(label).Role(ui.RoleGroup).Description("Use arrow keys to inspect categories; Home and End move to the first and last category; Escape clears inspection.")
	plot, _ := scaledChartValues(p)
	valid := validCategories(p, true)
	if p.Kind == RadarChart || p.Kind == RadialChart {
		valid = nil
		for i := range p.Labels {
			for j, s := range p.Series {
				if p.Kind == RadialChart && !p.Stacked && j > 0 {
					break
				}
				if i < len(s.Values) && finite(s.Values[i]) && s.Values[i] >= 0 {
					valid = append(valid, i)
					break
				}
			}
		}
	}
	if !slices.Contains(valid, *selected) {
		*selected = -1
	}
	inspectPolar(c, area, selected, valid, func(x, y float32, r ui.Rect) int { return polarHit(p, x, y, r) }, func(i int) string { return describeChart(c, p, i) })
	area.Draw(func(painter *ui.Painter, r ui.Rect) {
		cx, cy, radius := polarGeometry(r)
		if radius <= 0 {
			return
		}
		switch p.Kind {
		case PieChart, DonutChart:
			rings := min(len(p.Series), 1)
			if p.Stacked {
				rings = len(p.Series)
			}
			if rings == 0 {
				return
			}
			inner := float32(0)
			if p.Kind == DonutChart {
				inner = p.InnerRadius
				if inner <= 0 {
					inner = .6
				}
				inner = min(.9, max(0, inner))
			}
			for j := range rings {
				outerR := radius - float32(j)*(radius*(1-inner))/float32(rings)
				innerR := radius - float32(j+1)*(radius*(1-inner))/float32(rings)
				if rings > 1 {
					innerR += 2
				}
				values := positiveValues(p, j)
				for i, slice := range pieSlices(values) {
					if values[i] <= 0 {
						continue
					}
					color := categoryColor(c, p, i)
					outer := outerR
					if *selected == i {
						outer += 3
					}
					painter.FillPath(chart.Sector(cx, cy, innerR, outer, slice), color)
				}
			}
		case RadialChart:
			values := positiveValues(plot, 0)
			maxValue := polarMaximum(plot)
			var stacks [][]chart.Segment
			if p.Stacked {
				positive := plot
				positive.Normalized = false
				positive.Series = make([]ChartSeries, len(plot.Series))
				for j, series := range plot.Series {
					positive.Series[j] = series
					positive.Series[j].Values = positiveValues(plot, j)
				}
				stacks = stackSegments(positive)
				if p.MaxValue <= 0 {
					for _, series := range stacks {
						for _, seg := range series {
							maxValue = max(maxValue, seg.Hi)
						}
					}
				}
			}
			count := len(values)
			if count == 0 {
				return
			}
			band := radius / float32(count+1)
			for i, v := range values {
				outer := radius - float32(i)*band
				inner := outer - band*.65
				track := chart.Slice{Start: 0, End: 2 * math.Pi}
				painter.FillPath(chart.Sector(cx, cy, inner, outer, track), t.Muted)
				if len(stacks) > 0 {
					for j, series := range stacks {
						seg := series[i]
						if seg.Hi <= seg.Lo {
							continue
						}
						slice := chart.Slice{Start: 2 * math.Pi * min(1, seg.Lo/maxValue), End: 2 * math.Pi * min(1, seg.Hi/maxValue)}
						painter.FillPath(chart.Sector(cx, cy, inner, outer, slice), chartColor(c, p, j))
					}
					if *selected == i {
						painter.StrokePath(chart.Sector(cx, cy, inner, outer, track), 2, t.Foreground)
					}
				} else if v > 0 {
					slice := chart.Slice{Start: 0, End: 2 * math.Pi * min(1, v/maxValue)}
					painter.FillPath(chart.Sector(cx, cy, inner, outer, slice), categoryColor(c, p, i))
					if *selected == i {
						painter.StrokePath(chart.Sector(cx, cy, inner, outer, slice), 2, t.Foreground)
					}
				}
			}
		case RadarChart:
			count := len(p.Labels)
			if count < 3 {
				return
			}
			maxValue := polarMaximum(plot)
			for level := 1; level <= 4; level++ {
				path := radarPath(cx, cy, radius*float32(level)/4, count, nil, maxValue)
				painter.StrokePath(path, 1, t.Border)
			}
			for i := range count {
				px, py := chart.Polar(cx, cy, radius, 2*math.Pi*float64(i)/float64(count))
				painter.Line(cx, cy, px, py, 1, t.Border)
			}
			for j, s := range plot.Series {
				color := chartColor(c, p, j)
				complete := len(s.Values) >= count
				for i := range count {
					if i >= len(s.Values) || !finite(s.Values[i]) || s.Values[i] < 0 {
						complete = false
						break
					}
				}
				if complete {
					path := radarPath(cx, cy, radius, count, s.Values, maxValue)
					painter.FillPath(path, color.Alpha(.15))
					painter.StrokePath(path, 2, color)
				} else {
					for _, path := range radarMissingPaths(cx, cy, radius, count, s.Values, maxValue) {
						painter.StrokePath(path, 2, color)
					}
				}
				if *selected >= 0 && *selected < len(s.Values) && finite(s.Values[*selected]) && s.Values[*selected] >= 0 {
					px, py := chart.Polar(cx, cy, radius*float32(min(1, max(0, s.Values[*selected]/maxValue))), 2*math.Pi*float64(*selected)/float64(count))
					painter.FillPath((&ui.Path{}).Circle(px, py, 4), color)
				}
			}
		}
	})
	return area.Children(func() {
		if p.Kind == RadarChart && len(p.Labels) >= 3 {
			for i, label := range p.Labels {
				angle := 2 * math.Pi * float64(i) / float64(len(p.Labels))
				ui.Text(c, label).Absolute().LeftPercent(50+44*float32(math.Sin(angle))).TopPercent(50-44*float32(math.Cos(angle))).TranslatePercent(-50, -50).MaxWidth(90).SingleLine().FontSize(11).TextColor(t.MutedForeground).IgnoreInput(true)
			}
		}
		chartTooltipAt(c, p, *selected)
		if len(valid) == 0 {
			ui.Text(c, "No data").FontSize(13).TextColor(t.MutedForeground).Absolute().TopPercent(50).LeftPercent(50).TranslatePercent(-50, -50)
		}
	})
}

func radarPath(cx, cy, radius float32, count int, values []float64, maxValue float64) *ui.Path {
	path := &ui.Path{}
	for i := range count {
		at := float64(1)
		if values != nil {
			at = 0
			if i < len(values) && finite(values[i]) {
				at = min(1, max(0, values[i]/maxValue))
			}
		}
		x, y := chart.Polar(cx, cy, radius*float32(at), 2*math.Pi*float64(i)/float64(max(1, count)))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	return path.Close()
}

func radarMissingPaths(cx, cy, radius float32, count int, values []float64, maxValue float64) []*ui.Path {
	var out []*ui.Path
	var path *ui.Path
	for i := range count {
		if i >= len(values) || !finite(values[i]) || values[i] < 0 {
			path = nil
			continue
		}
		x, y := chart.Polar(cx, cy, radius*float32(min(1, max(0, values[i]/maxValue))), 2*math.Pi*float64(i)/float64(count))
		if path == nil {
			path = (&ui.Path{}).MoveTo(x, y)
			out = append(out, path)
		} else {
			path.LineTo(x, y)
		}
	}
	return out
}

func polarHit(p ChartProps, x, y float32, r ui.Rect) int {
	// PointerPosition is relative to the element; use a relative plot box.
	cx, cy, radius := polarGeometry(ui.Rect{W: r.W, H: r.H})
	dx, dy := x-cx, y-cy
	distance := float32(math.Hypot(float64(dx), float64(dy)))
	if distance > radius+3 || radius <= 0 {
		return -1
	}
	switch p.Kind {
	case PieChart, DonutChart:
		inner := float32(0)
		if p.Kind == DonutChart {
			inner = p.InnerRadius
			if inner <= 0 {
				inner = .6
			}
			inner = min(.9, max(0, inner))
		}
		if distance < radius*inner {
			return -1
		}
		rings := min(len(p.Series), 1)
		if p.Stacked {
			rings = len(p.Series)
		}
		if rings == 0 {
			return -1
		}
		ring := min(rings-1, int((radius-distance)/(radius*(1-inner)/float32(rings))))
		if ring < 0 {
			return -1
		}
		slices := pieSlices(positiveValues(p, ring))
		if i, ok := chart.SliceAt(slices, dx, dy, radius*inner, radius+3); ok {
			return i
		}
	case RadialChart:
		count := len(p.Labels)
		if count == 0 {
			return -1
		}
		band := radius / float32(count+1)
		index := int((radius - distance) / band)
		if index >= 0 && index < count {
			outer := radius - float32(index)*band
			if distance >= outer-band*.65 {
				return index
			}
		}
	case RadarChart:
		if distance < 5 {
			return -1
		}
		angle := math.Atan2(float64(dx), float64(-dy))
		if angle < 0 {
			angle += 2 * math.Pi
		}
		if len(p.Labels) > 0 {
			return int(math.Floor(angle/(2*math.Pi)*float64(len(p.Labels))+.5)) % len(p.Labels)
		}
	}
	return -1
}

func inspectPolar(c *ui.Context, area ui.Element, index *int, valid []int, hit func(float32, float32, ui.Rect) int, describe func(int) string) {
	area.Focusable()
	st := ui.Local(area, "polar-inspection", func() chartInspection { return chartInspection{last: -1} })
	cursor := slices.Index(valid, *index)
	keyboard := false
	move := func(at int) {
		if len(valid) > 0 {
			*index = valid[max(0, min(len(valid)-1, at))]
			keyboard = true
		}
	}
	switch {
	case area.Shortcut(0, ui.KeyRight) || area.Shortcut(0, ui.KeyDown):
		move(cursor + 1)
	case area.Shortcut(0, ui.KeyLeft) || area.Shortcut(0, ui.KeyUp):
		if cursor < 0 {
			move(len(valid) - 1)
		} else {
			move(cursor - 1)
		}
	case area.Shortcut(0, ui.KeyHome):
		move(0)
	case area.Shortcut(0, ui.KeyEnd):
		move(len(valid) - 1)
	case area.Shortcut(0, ui.KeyEscape):
		*index = -1
		keyboard = true
	}
	if x, y, over := area.PointerPosition(); over {
		if !keyboard && (!st.hovering || x != st.x || y != st.y) {
			candidate := hit(x, y, area.Bounds())
			*index = -1
			if slices.Contains(valid, candidate) {
				*index = candidate
			}
		}
		st.hovering, st.x, st.y = true, x, y
	} else if st.hovering {
		st.hovering = false
		*index = -1
	}
	if *index >= 0 {
		value := describe(*index)
		area.Value(value)
		if keyboard && st.last != *index {
			c.Announce(value)
		}
	}
	st.last = *index
}
