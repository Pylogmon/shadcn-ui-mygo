package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// ChartPolarProps opts into source-oriented polar composition. Nil on Chart
// preserves the original polar profile and its public defaults. The source
// profile uses degree angles (right=0, positive counterclockwise), 80% outer
// radius, 5-DIP margins, Pie/Radial 0..360 and Radar 90..-270 by default.
// An absent Grid/AngleAxis/RadiusAxis does not create that decorative component.
type ChartPolarProps struct {
	Margin                   *ChartInsets
	StartAngle, EndAngle     *float64
	InnerRadius, OuterRadius *ChartRadius
	Stroke                   *ui.Color
	StrokeWidth              *float32
	CornerRadius             float32
	Background               bool
	// SelectedSeries is the inspected Pie/Donut ring. Nil keeps local state;
	// pointer hits and PageUp/PageDown update it. Other families inspect all
	// series of a category together and ignore this option.
	SelectedSeries *int
	Grid           *ChartPolarGridProps
	AngleAxis      *ChartPolarAngleAxisProps
	RadiusAxis     *ChartPolarRadiusAxisProps
	// Center runs in the current build at the plot center. It receives raw
	// data, independent of an intermediate animated sector or polygon.
	Center func(*ui.Context, ChartPolarCenter)
}

type ChartPolarGridKind uint8

const (
	ChartPolarPolygon ChartPolarGridKind = iota
	ChartPolarCircle
)

type ChartPolarGridProps struct {
	Kind        ChartPolarGridKind
	RadialLines *bool
	// Nil radii use the source axis ticks. Explicit empty suppresses rings.
	Radii  []ChartRadius
	Angles []float64
	// Fill colors each concentric shape; Fills can override individual rings
	// in source order (for example outer muted and inner background).
	Fill, Stroke *ui.Color
	Fills        []ui.Color
	StrokeWidth  *float32
}
type ChartPolarAngleAxisProps struct {
	HideLabels, HideLine, HideTickLines bool
	TickSize                            *float32
	Color                               *ui.Color
	// Content replaces one native label; metadata contains original values.
	Content func(*ui.Context, ChartPolarTick)
}
type ChartPolarRadiusAxisProps struct {
	Angle                float64
	TextAlign            ui.Align
	HideLabels, HideLine bool
	Color                *ui.Color
	Formatter            func(float64) string
}
type ChartPolarTick struct {
	Index     int
	Label     string
	Values    []float64
	Angle     float64
	TextAlign ui.Align
}
type ChartPolarCenter struct {
	Selected       int
	SelectedSeries int
	Label          string
	Value, Total   float64
}

// ChartPolarSector is a detached, pure-value sector. Shape may return one or
// several copies with changed radii/colors, such as the source active outer
// ring. It must be pure: pointer resolution can evaluate it before and after a
// selection change in the same build. No resolver is retained by Draw/state.
type ChartPolarSector struct {
	Datum                     ChartDatum
	InnerRadius, OuterRadius  ChartRadius
	StartAngle, EndAngle      float64
	Color, Stroke             ui.Color
	StrokeWidth, CornerRadius float32
}
type ChartPolarSeriesProps struct {
	InnerRadius, OuterRadius *ChartRadius
	CornerRadius             *float32
	Shape                    func(ChartPolarSector) []ChartPolarSector
	Labels                   []ChartPolarLabelProps
}
type ChartPolarLabelPosition uint8

const (
	ChartPolarLabelMiddle ChartPolarLabelPosition = iota
	ChartPolarLabelOutside
	ChartPolarLabelInsideStart
)

type ChartPolarLabelProps struct {
	Position  ChartPolarLabelPosition
	Offset    *float32
	HideLine  bool
	Color     *ui.Color
	FontSize  float32
	Formatter func(ChartDatum) string
	Content   func(*ui.Context, ChartDatum)
}

type chartPolarFrame struct {
	sectors      []ChartPolarSector
	tracks       []ChartPolarSector
	radar        [][]chartGeometryDatum
	radarTargets [][]chartGeometryDatum
	maximum      float64
	finished     bool
}
type chartPolarLayout struct {
	bounds ui.Rect
	primed bool
}
type chartPolarAnimationKey struct{}
type chartPolarBox struct{ cx, cy, maximum float32 }

func chartSourcePolarBox(p ChartProps, r ui.Rect) chartPolarBox {
	m := ChartInsets{Top: 5, Right: 5, Bottom: 5, Left: 5}
	if p.Polar.Margin != nil {
		m = *p.Polar.Margin
	}
	m.Top, m.Right, m.Bottom, m.Left = chartFinite32(m.Top, 0), chartFinite32(m.Right, 0), chartFinite32(m.Bottom, 0), chartFinite32(m.Left, 0)
	w, h := r.W-m.Left-m.Right, r.H-m.Top-m.Bottom
	return chartPolarBox{r.X + m.Left + w/2, r.Y + m.Top + h/2, max(0, min(float32(math.Abs(float64(w))), float32(math.Abs(float64(h))))/2)}
}

func chartPolarRawMaximum(p ChartProps) float64 {
	if finite(p.MaxValue) && p.MaxValue > 0 {
		return p.MaxValue
	}
	maximum := 0.0
	if p.Kind == RadialChart && p.Stacked {
		q, scale := scaledChartValues(p)
		q.Normalized = false
		for _, stack := range stackSegments(q) {
			for _, span := range stack {
				maximum = max(maximum, span.Hi*scale)
			}
		}
	} else {
		for _, s := range p.Series {
			for i, v := range s.Values {
				if i < len(p.Labels) && finite(v) {
					maximum = max(maximum, v)
				}
			}
		}
	}
	if p.Kind == RadarChart {
		return chartPolarNiceMaximum(maximum)
	}
	if maximum <= 0 {
		return 1
	}
	return maximum
}
func chartPolarSeriesCorner(p ChartProps, j int) float32 {
	v := max(0, chartFinite32(p.Polar.CornerRadius, 0))
	if s := p.Series[j].Style.Polar; s != nil {
		v = chartNonnegative(s.CornerRadius, v)
	}
	return v
}

func chartPolarSourceFrame(c *ui.Context, area ui.Element, p ChartProps, bounds ui.Rect) chartPolarFrame {
	start, end := chartPolarAngles(p)
	box := chartSourcePolarBox(p, bounds)
	maximum := chartPolarRawMaximum(p)
	plans := make([]chartGeometryPlan, len(p.Series))
	scaled, _ := scaledChartValues(p)
	if p.Kind == RadialChart {
		scaled.Normalized = false
	}
	var stacks [][]float64
	if p.Kind == RadialChart {
		maximum = chartPolarRawMaximum(scaled)
	}
	if p.Kind == RadialChart && p.Stacked {
		segments := stackSegments(scaled)
		stacks = make([][]float64, len(segments))
		for j, series := range segments {
			stacks[j] = make([]float64, len(p.Labels)*2)
			for i, span := range series {
				stacks[j][2*i], stacks[j][2*i+1] = span.Lo, span.Hi
			}
		}
	}
	for j, series := range p.Series {
		points, seeds := make([]chartGeometryDatum, len(p.Labels)), make([]chartGeometryDatum, len(p.Labels))
		largest, total := 0.0, 0.0
		if p.Kind == PieChart || p.Kind == DonutChart {
			for i, v := range series.Values {
				if i < len(p.Labels) && finite(v) {
					largest = max(largest, v)
				}
			}
			if largest > 0 {
				for i, v := range series.Values {
					if i < len(p.Labels) && finite(v) && v > 0 {
						total += v / largest
					}
				}
			}
		}
		for i := range points {
			if i >= len(series.Values) || !finite(series.Values[i]) || series.Values[i] < 0 {
				continue
			}
			value := series.Values[i]
			point := chartGeometryDatum{valid: true}
			switch p.Kind {
			case PieChart, DonutChart:
				if total <= 0 {
					continue
				}
				point.v[0] = math.Copysign(min(360, math.Abs(end-start)), end-start) * (value / largest) / total
			case RadialChart:
				lo, hi := 0.0, scaled.Series[j].Values[i]
				if len(stacks) > j {
					lo, hi = stacks[j][2*i], stacks[j][2*i+1]
				}
				point.v[0], point.v[1] = start+(end-start)*(lo/maximum), start+(end-start)*(hi/maximum)
			case RadarChart:
				inner, outer := chartPolarRadii(p, j)
				radius := inner.resolve(box.maximum) + (outer.resolve(box.maximum)-inner.resolve(box.maximum))*float32(value/maximum)
				angle := start + (end-start)*float64(i)/float64(max(1, len(p.Labels)))
				x, y := chartPolarPoint(0, 0, radius, angle)
				point.v[0], point.v[1] = float64(x), float64(y)
			}
			points[i] = point
			seeds[i] = point
			if p.Kind == RadialChart {
				seeds[i].v[1] = seeds[i].v[0]
			} else {
				seeds[i].v[0] = 0
				if p.Kind == RadarChart {
					seeds[i].v[1] = 0
				}
			}
		}
		plans[j] = chartGeometryPlan{points: points, seeds: seeds}
	}
	// The legacy profile stores radians and normalized Radar points; this
	// profile stores degrees and DIP points. Never reinterpret its old state.
	frames := animateChartGeometryWithKey(c, area, p.Kind, p.Animation, plans, chartPolarAnimationKey{})
	out := chartPolarFrame{maximum: maximum, finished: true}
	for _, frame := range frames {
		out.finished = out.finished && !frame.state.active
	}
	if p.Kind == RadarChart {
		out.radar = make([][]chartGeometryDatum, len(frames))
		out.radarTargets = make([][]chartGeometryDatum, len(plans))
		for j, frame := range frames {
			out.radar[j] = frame.points
			out.radarTargets[j] = plans[j].points
		}
		return out
	}
	for j, frame := range frames {
		inner, outer := chartPolarRadii(p, j)
		cursor := start
		for i, point := range frame.points {
			if !point.valid {
				continue
			}
			a, b := cursor, cursor+point.v[0]
			ri, ro := inner, outer
			if p.Kind == RadialChart {
				a, b = point.v[0], point.v[1]
				step := 1 / float32(max(1, len(p.Labels)))
				ri = chartRadiusMix(inner, outer, (float32(i)+.1)*step)
				ro = chartRadiusMix(inner, outer, (float32(i)+.9)*step)
			} else {
				cursor = b
			}
			color := categoryColor(c, p, i)
			if p.Stacked && p.Kind == RadialChart {
				color = chartColor(c, p, j)
			}
			color = color.Alpha(chartOpacity(p.Series[j].Style.FillOpacity, 1))
			stroke, width := ui.Transparent, float32(0)
			if p.Kind == PieChart || p.Kind == DonutChart {
				stroke, width = theme.TokensOf(c).Background, 1
			}
			if p.Polar.Stroke != nil {
				stroke = *p.Polar.Stroke
			}
			width = chartNonnegative(p.Polar.StrokeWidth, width)
			if p.Series[j].Style.Stroke != nil {
				stroke = *p.Series[j].Style.Stroke
			}
			width = chartNonnegative(p.Series[j].Style.StrokeWidth, width)
			sector := ChartPolarSector{Datum: ChartDatum{Series: j, Index: i, Label: p.Labels[i], Name: seriesName(p, j), Value: p.Series[j].Values[i], Color: color}, InnerRadius: ri, OuterRadius: ro, StartAngle: a, EndAngle: b, Color: color, Stroke: stroke, StrokeWidth: width, CornerRadius: chartPolarSeriesCorner(p, j)}
			mx, my := chartPolarPoint(box.cx, box.cy, (ri.resolve(box.maximum)+ro.resolve(box.maximum))/2, (a+b)/2)
			sector.Datum.X, sector.Datum.Y = mx/max(1, bounds.W), 1-my/max(1, bounds.H)
			out.sectors = append(out.sectors, sector)
			if p.Kind == RadialChart && p.Polar.Background && j == 0 {
				track := sector
				track.StartAngle, track.EndAngle = start, end
				track.Color = theme.TokensOf(c).Muted
				track.StrokeWidth = 0
				out.tracks = append(out.tracks, track)
			}
		}
	}
	return out
}

func chartPolarLayers(p ChartProps, frame chartPolarFrame, selected, selectedSeries int) []ChartPolarSector {
	var result []ChartPolarSector
	for _, sector := range frame.sectors {
		sector.Datum.Active = sector.Datum.Index == selected && (p.Kind == RadialChart || sector.Datum.Series == selectedSeries)
		if style := p.Series[sector.Datum.Series].Style.Polar; style != nil && style.Shape != nil {
			result = append(result, style.Shape(sector)...)
		} else {
			result = append(result, sector)
		}
	}
	return result
}

func chartSourcePolar(c *ui.Context, p ChartProps, selected *int, height float32) ui.Element {
	area := ui.Box(c).FillWidth().Height(height).Role(ui.RoleGroup).Label(p.Label).Description("Use arrow keys to inspect categories; Home and End move to the first and last category; Escape clears inspection.")
	if p.Label == "" {
		area.Label("Chart")
	}
	series := p.Polar.SelectedSeries
	if series == nil {
		series = ui.Local(area, "source-polar-series", func() int { return 0 })
	}
	if *series < 0 || *series >= len(p.Series) {
		*series = 0
	}
	if (p.Kind == PieChart || p.Kind == DonutChart) && len(p.Series) > 1 {
		area.Description("Use Left and Right arrows to inspect categories; Page Up and Page Down change rings; Home, End and Escape select or clear a value.")
		if area.Shortcut(0, ui.KeyPageDown) {
			*series = min(len(p.Series)-1, *series+1)
		} else if area.Shortcut(0, ui.KeyPageUp) {
			*series = max(0, *series-1)
		}
	}
	layout := ui.Local(area, "source-polar-layout", func() chartPolarLayout { return chartPolarLayout{} })
	if !layout.primed {
		layout.primed = true
		c.Invalidate()
	}
	bounds := area.LayoutBounds()
	if bounds.W > 0 && bounds.H > 0 {
		layout.bounds = bounds
	}
	bounds = layout.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		bounds = ui.Rect{W: height, H: height}
	}
	frame := chartPolarSourceFrame(c, area, p, bounds)
	before := *selected
	beforeSeries := *series
	layers := chartPolarLayers(p, frame, before, beforeSeries)
	valid := chartSourcePolarValid(chartPolarInspectionProps(p, *series))
	pointerValid := chartSourcePolarValid(p)
	inspectPolar(c, area, selected, valid, pointerValid, func(x, y float32, r ui.Rect) int {
		hitSeries, index := chartSourcePolarHit(p, frame, layers, x, y, r)
		if index >= 0 {
			*series = hitSeries
		}
		return index
	}, func(i int) string { return describeChart(c, chartPolarInspectionProps(p, *series), i) })
	if *selected != before || *series != beforeSeries {
		layers = chartPolarLayers(p, frame, *selected, *series)
	}
	paint := chartPolarPaintOf(c, p, frame, layers)
	services := c.Services()
	area.Draw(func(painter *ui.Painter, r ui.Rect) {
		if math.Abs(float64(r.W-layout.bounds.W)) > .25 || math.Abs(float64(r.H-layout.bounds.H)) > .25 {
			layout.bounds = ui.Rect{W: r.W, H: r.H}
			services.Invalidate()
		}
		paint.draw(painter, r)
	})
	return area.Children(func() {
		chartPolarNativeContent(c, p, frame, *selected, *series, bounds)
		chartTooltipAt(c, chartPolarInspectionProps(p, *series), *selected)
		if len(valid) == 0 {
			ui.Text(c, "No data").Absolute().LeftPercent(50).TopPercent(50).TranslatePercent(-50, -50).TextColor(theme.TokensOf(c).MutedForeground)
		}
	})
}

func chartPolarInspectionProps(p ChartProps, series int) ChartProps {
	if (p.Kind == PieChart || p.Kind == DonutChart) && series >= 0 && series < len(p.Series) {
		s := p.Series[series]
		s.Name = seriesName(p, series)
		p.Series = []ChartSeries{s}
	}
	return p
}

func chartSourcePolarValid(p ChartProps) []int {
	var result []int
	for i := range p.Labels {
		for _, series := range p.Series {
			if i >= len(series.Values) {
				continue
			}
			v := series.Values[i]
			if finite(v) && ((p.Kind == RadarChart || p.Kind == RadialChart) && v >= 0 || v > 0) {
				result = append(result, i)
				break
			}
		}
	}
	return result
}
