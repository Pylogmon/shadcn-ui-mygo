package display

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// ChartInsets are physical top/right/bottom/left DIPs. They affect the inner
// Cartesian plot and axes, not the surrounding card or legend.
type ChartInsets struct{ Top, Right, Bottom, Left float32 }
type ChartBandPadding struct{ Inner, Outer float32 }
type ChartCursorProps struct {
	Band  bool
	Color *ui.Color
}
type chartAxisLayout struct {
	bounds ui.Rect
	primed bool
}

type ChartCurve uint8

const (
	ChartCurveMonotone ChartCurve = iota
	ChartCurveNatural
	ChartCurveLinear
	ChartCurveStep
)

// ChartAxisTick describes the displayed tick; Value is the original value
// (normalized shares for a normalized stack), independent of its formatted text.
type ChartAxisTick struct {
	Side     chart.Side
	Category bool
	Index    int
	Value    float64
	Label    string
}

// ChartDatum keeps original measurements for custom marks and label content.
// X/Y are fractions of the plot, with Y growing upward. Active describes the
// inspected category, not the animation's intermediate data. Callbacks run
// synchronously during Chart; do not retain Context or Element after the build.
type ChartDatum struct {
	Series, Index int
	Label, Name   string
	Value         float64
	Color         ui.Color
	X, Y          float32
	Active        bool
}

type ChartCornerRadii struct{ TopLeft, TopRight, BottomRight, BottomLeft float32 }

// ChartDotProps enables an idle dot when assigned to Dots. ActiveDot nil keeps
// the existing 7-DIP active dot. Disabled suppresses it; nonpositive Radius uses
// 4 DIPs. Radius is the SVG-style centerline radius: configured dots include
// half the stroke on each outside edge. Content customizes the native point element, including its size, paint
// and children, during this build. Neither callback nor Context enters Draw.
type ChartDotProps struct {
	Disabled     bool
	Radius       float32
	Fill, Stroke *ui.Color
	StrokeWidth  *float32
	Content      func(*ui.Context, ChartDatum, ui.Element)
}

type ChartLabelPosition uint8

const (
	ChartLabelTop ChartLabelPosition = iota
	ChartLabelInsideLeft
	ChartLabelRight
)

// ChartLabelProps implements the Cartesian placements used by the fixed source
// LabelList examples. Offset nil uses 5 DIPs; explicit 0 is supported. Formatter
// receives original values and category labels. Content replaces native text.
// Labels appear after data animation, as in the source, and are decorative AX
// content: the plot's inspected value remains the original measurement.
type ChartLabelProps struct {
	Position  ChartLabelPosition
	Offset    *float32
	Color     *ui.Color
	FontSize  float32
	Formatter func(ChartDatum) string
	Content   func(*ui.Context, ChartDatum)
}

// ChartSeriesStyle is optional per-series Cartesian styling. Nil opacity,
// stroke and width preserve existing defaults; explicit transparent/zero values
// remain meaningful. Gradient uses public native linear-gradient stops, copied
// before Draw. BarContent can style or draw over the real native bar element.
// Callbacks are build-only and never stored in the geometry animation state.
type ChartSeriesStyle struct {
	Curve           *ChartCurve
	FillOpacity     *float32
	Stroke          *ui.Color
	StrokeWidth     *float32
	Gradient        *ui.LinearGradient
	BarRadius       *ChartCornerRadii
	Dots, ActiveDot *ChartDotProps
	BarContent      func(*ui.Context, ChartDatum, ui.Element)
	Labels          []ChartLabelProps
}

func chartFinite32(v float32, fallback float32) float32 {
	if !finite(float64(v)) {
		return fallback
	}
	return v
}
func chartNonnegative(v *float32, fallback float32) float32 {
	if v == nil {
		return fallback
	}
	return max(0, chartFinite32(*v, fallback))
}
func chartOpacity(v *float32, fallback float32) float32 {
	return min(1, chartNonnegative(v, fallback))
}
func chartCurveOf(p ChartProps, series int) ChartCurve {
	k := p.Cartesian.Curve
	if s := p.Series[series].Style.Curve; s != nil {
		k = *s
	}
	if k > ChartCurveStep {
		return ChartCurveMonotone
	}
	return k
}
func chartNativeCurve(k ChartCurve) chart.Curve {
	switch k {
	case ChartCurveLinear:
		return chart.CurveLinear
	case ChartCurveStep:
		return chart.CurveStep
	}
	return chart.CurveMonotone
}

func chartGradientCopy(g *ui.LinearGradient, opacity float32) *ui.LinearGradient {
	if g == nil {
		return nil
	}
	v := *g
	v.From, v.To = chartWithOpacity(v.From, opacity), chartWithOpacity(v.To, opacity)
	v.Stops = append([]ui.GradientStop(nil), g.Stops...)
	for i := range v.Stops {
		v.Stops[i].Color = chartWithOpacity(v.Stops[i].Color, opacity)
	}
	return &v
}

type chartCartesianPaint struct {
	curve        ChartCurve
	fill, stroke ui.Color
	width        float32
	gradient     *ui.LinearGradient
}

func chartCartesianStyle(p ChartProps, j int, color ui.Color) chartCartesianPaint {
	s := p.Series[j].Style
	alpha := float32(1)
	if p.Kind == AreaChart {
		alpha = .2
	}
	alpha = chartOpacity(s.FillOpacity, alpha)
	stroke := color
	if s.Stroke != nil {
		stroke = *s.Stroke
	}
	return chartCartesianPaint{chartCurveOf(p, j), chartWithOpacity(color, alpha), stroke, chartNonnegative(s.StrokeWidth, 2), chartGradientCopy(s.Gradient, alpha)}
}
func chartWithOpacity(c ui.Color, alpha float32) ui.Color { return c.Alpha(alpha) }

func chartDatum(p ChartProps, j, i int, point chartGeometryDatum, color ui.Color, selected int) ChartDatum {
	d := ChartDatum{Series: j, Index: i, Label: p.Labels[i], Name: seriesName(p, j), Value: p.Series[j].Values[i], Color: color, Active: selected == i}
	if p.Kind == BarChart {
		d.X, d.Y = float32((point.v[0]+point.v[1])/2), float32(point.v[3])
		if horizontalBars(p) {
			d.X, d.Y = float32(point.v[1]), float32((point.v[2]+point.v[3])/2)
		}
	} else {
		d.X, d.Y = float32(point.v[0]), float32(point.v[1])
	}
	return d
}

func chartStyleBar(c *ui.Context, p ChartProps, j int, d ChartDatum, mark ui.Element) {
	s := p.Series[j].Style
	mark.Background(chartWithOpacity(d.Color, chartOpacity(s.FillOpacity, 1)))
	if g := chartGradientCopy(s.Gradient, chartOpacity(s.FillOpacity, 1)); g != nil {
		mark.LinearGradient(*g)
	}
	if s.BarRadius != nil {
		r := *s.BarRadius
		mark.Radius(max(0, chartFinite32(r.TopLeft, 0)), max(0, chartFinite32(r.TopRight, 0)), max(0, chartFinite32(r.BottomRight, 0)), max(0, chartFinite32(r.BottomLeft, 0)))
	}
	if s.Stroke != nil {
		mark.Border(chartNonnegative(s.StrokeWidth, 1), *s.Stroke)
	}
	if s.BarContent != nil {
		s.BarContent(c, d, mark)
	}
}

type chartDotKey struct {
	Series, Index int
	Active        bool
}

func chartBuildDot(c *ui.Context, p ChartProps, j, i int, point chartGeometryDatum, color ui.Color, selected int, active bool, spec *ChartDotProps) {
	if spec != nil && spec.Disabled {
		return
	}
	radius, width := float32(3.5), float32(2)
	fill, stroke := color, theme.TokensOf(c).Background
	if spec != nil {
		radius = spec.Radius
		if !finite(float64(radius)) || radius <= 0 {
			radius = 4
		}
		width = chartNonnegative(spec.StrokeWidth, 1)
		stroke = color
		if spec.Fill != nil {
			fill = *spec.Fill
		}
		if spec.Stroke != nil {
			stroke = *spec.Stroke
		}
	}
	d := chartDatum(p, j, i, point, color, selected)
	outer := radius
	if spec != nil {
		outer += width / 2
	}
	mark := chart.Point(c.Key(chartDotKey{j, i, active}), d.X, d.Y, 2*outer, 2*outer).Radius(outer).Background(fill).Border(width, stroke).IgnoreInput(true).AccessHidden(true)
	if spec != nil && spec.Content != nil {
		spec.Content(c, d, mark)
	}
}

type chartDataLabelKey struct{ Series, Label, Index int }

func chartBuildLabels(c *ui.Context, p ChartProps, j int, plan chartGeometryPlan, selected int, color ui.Color) {
	for k, spec := range p.Series[j].Style.Labels {
		for i, point := range plan.points {
			if i >= len(p.Series[j].Values) || !finite(p.Series[j].Values[i]) {
				continue
			}
			// Zero rectangles still have a valid original datum and label anchor.
			if !point.valid && p.Kind != BarChart {
				continue
			}
			d := chartDatum(p, j, i, point, color, selected)
			x, y, h, v, dx, dy := d.X, d.Y, ui.Center, ui.End, float32(0), float32(0)
			offset := chartNonnegative(spec.Offset, 5)
			negativeHeight := p.Kind == BarChart && !horizontalBars(p) && point.v[3] < point.v[2]
			if negativeHeight {
				v, dy = ui.Start, offset
			} else {
				dy = -offset
			}
			if p.Kind == BarChart {
				left, right := float32(point.v[0]), float32(point.v[1])
				top, bottom := float32(1-point.v[3]), float32(1-point.v[2])
				if horizontalBars(p) {
					top, bottom = 1-float32(max(point.v[2], point.v[3])), 1-float32(min(point.v[2], point.v[3]))
				}
				switch spec.Position {
				case ChartLabelInsideLeft:
					x, y, h, v, dx, dy = left, 1-(top+bottom)/2, ui.Start, ui.Center, offset, 0
					if right < left {
						h, dx = ui.End, -offset
					}
				case ChartLabelRight:
					x, y, h, v, dx, dy = right, 1-(top+bottom)/2, ui.Start, ui.Center, offset, 0
					if right < left {
						h, dx = ui.End, -offset
					}
				default:
					x, y = (left+right)/2, 1-top
				}
			}
			font := spec.FontSize
			if !finite(float64(font)) || font <= 0 {
				font = 12
			}
			textColor := theme.TokensOf(c).Foreground
			if spec.Color != nil {
				textColor = *spec.Color
			}
			chart.Anchor(c.Key(chartDataLabelKey{j, k, i}), x, y, h, v, func() {
				ui.Box(c).Translate(dx, dy).FontSize(font).TextColor(textColor).Children(func() {
					if spec.Content != nil {
						spec.Content(c, d)
					} else {
						text := chartFormat(d.Value)
						if spec.Formatter != nil {
							text = spec.Formatter(d)
						}
						ui.Text(c, text).NoWrap()
					}
				})
			}).IgnoreInput(true).AccessHidden(true)
		}
	}
}

func chartTickCount(p ChartProps) int {
	if p.Cartesian.ValueTickCount > 0 {
		return min(p.Cartesian.ValueTickCount, 100)
	}
	return 4
}
func chartMargins(p ChartProps) ChartInsets {
	if p.Cartesian.Margin == nil {
		return ChartInsets{}
	}
	m := *p.Cartesian.Margin
	m.Top, m.Right, m.Bottom, m.Left = chartFinite32(m.Top, 0), chartFinite32(m.Right, 0), chartFinite32(m.Bottom, 0), chartFinite32(m.Left, 0)
	return m
}

func chartGroupGap(plans []chartGeometryPlan, horizontal bool, gap float32) {
	// Grouped plans already split a category band equally. Inset and shift
	// each member so the total category extent stays fixed with DIP gaps.
	axis := 0
	if horizontal {
		axis = 2
	}
	n := len(plans)
	for j := range plans {
		for i := range plans[j].points {
			point := &plans[j].points[i]
			seed := &plans[j].seeds[i]
			lo, hi := min(point.v[axis], point.v[axis+1]), max(point.v[axis], point.v[axis+1])
			width := hi - lo
			g := min(float64(gap), width*float64(n)/float64(n-1))
			start := lo - float64(j)*width
			index := j
			if horizontal {
				start = hi + float64(j)*width
				index = n - 1 - j
				start -= float64(n) * width
			}
			newWidth := max(0, width-g*float64(n-1)/float64(n))
			newLo := start + float64(index)*(newWidth+g)
			point.v[axis], point.v[axis+1] = newLo, newLo+newWidth
			seed.v[axis], seed.v[axis+1] = point.v[axis], point.v[axis+1]
		}
	}
}
