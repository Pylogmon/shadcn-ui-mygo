package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type chartPolarGridPaint struct {
	kind         ChartPolarGridKind
	radii        []ChartRadius
	angles       []float64
	radial       bool
	stroke, fill ui.Color
	fills        []ui.Color
	width        float32
}
type chartPolarRadarPaint struct {
	points       []chartGeometryDatum
	fill, stroke ui.Color
	width        float32
}
type chartSourcePolarPaint struct {
	kind                    ChartKind
	margin                  ChartInsets
	inner, outer            ChartRadius
	grid                    *chartPolarGridPaint
	layers, tracks          []ChartPolarSector
	radar                   []chartPolarRadarPaint
	angles                  []float64
	axisColor               ui.Color
	axisLine, axisTickLines bool
	tickSize                float32
	labelLines              []chartPolarLineLabel
}
type chartPolarLineLabel struct {
	sector ChartPolarSector
	offset float32
}

func chartPolarPaintOf(c *ui.Context, p ChartProps, frame chartPolarFrame, layers []ChartPolarSector) chartSourcePolarPaint {
	t := theme.TokensOf(c)
	start, end := chartPolarAngles(p)
	inner, outer := chartPolarRadii(p, 0)
	paint := chartSourcePolarPaint{kind: p.Kind, margin: ChartInsets{Top: 5, Right: 5, Bottom: 5, Left: 5}, inner: inner, outer: outer, layers: append([]ChartPolarSector(nil), layers...), tracks: append([]ChartPolarSector(nil), frame.tracks...), axisColor: t.Border.Alpha(.5)}
	if p.Polar.Margin != nil {
		paint.margin = *p.Polar.Margin
	}
	for i := range p.Labels {
		paint.angles = append(paint.angles, start+(end-start)*float64(i)/float64(max(1, len(p.Labels))))
	}
	if axis := p.Polar.AngleAxis; axis != nil {
		paint.axisLine = !axis.HideLine
		paint.axisTickLines = !axis.HideTickLines
		paint.tickSize = chartNonnegative(axis.TickSize, 8)
		if axis.Color != nil {
			paint.axisColor = *axis.Color
		}
	}
	if grid := p.Polar.Grid; grid != nil {
		g := &chartPolarGridPaint{kind: grid.Kind, radial: true, stroke: t.Border.Alpha(.5), width: chartNonnegative(grid.StrokeWidth, 1), radii: append([]ChartRadius(nil), grid.Radii...), angles: append([]float64(nil), grid.Angles...), fills: append([]ui.Color(nil), grid.Fills...)}
		if grid.RadialLines != nil {
			g.radial = *grid.RadialLines
		}
		if grid.Stroke != nil {
			g.stroke = *grid.Stroke
		}
		if grid.Fill != nil {
			g.fill = *grid.Fill
		}
		if grid.Radii == nil {
			if p.Kind == RadarChart {
				for i := 0; i < 5; i++ {
					g.radii = append(g.radii, chartRadiusMix(inner, outer, float32(i)/4))
				}
			} else {
				for i := range p.Labels {
					g.radii = append(g.radii, chartRadiusMix(inner, outer, (float32(i)+.5)/float32(max(1, len(p.Labels)))))
				}
			}
		}
		if grid.Angles == nil {
			if p.Kind == RadarChart {
				g.angles = append([]float64(nil), paint.angles...)
			} else {
				for _, tick := range chart.NewLinear(0, frame.maximum).Ticks(10) {
					g.angles = append(g.angles, start+(end-start)*tick.Value/frame.maximum)
				}
			}
		}
		angles := g.angles[:0]
		for _, angle := range g.angles {
			if finite(angle) {
				angles = append(angles, angle)
			}
		}
		g.angles = angles
		paint.grid = g
	}
	for j, points := range frame.radar {
		col := chartColor(c, p, j)
		s := p.Series[j].Style
		stroke, width := ui.Transparent, float32(0)
		if s.Stroke != nil {
			stroke = *s.Stroke
			width = chartNonnegative(s.StrokeWidth, 1)
		}
		paint.radar = append(paint.radar, chartPolarRadarPaint{append([]chartGeometryDatum(nil), points...), col.Alpha(chartOpacity(s.FillOpacity, 1)), stroke, width})
	}
	if frame.finished {
		for _, sector := range frame.sectors {
			if s := p.Series[sector.Datum.Series].Style.Polar; s != nil {
				for _, label := range s.Labels {
					if label.Position == ChartPolarLabelOutside && !label.HideLine {
						paint.labelLines = append(paint.labelLines, chartPolarLineLabel{sector, chartNonnegative(label.Offset, 20)})
					}
				}
			}
		}
	}
	return paint
}
func chartPolarPaintBox(m ChartInsets, r ui.Rect) chartPolarBox {
	m.Top, m.Right, m.Bottom, m.Left = chartFinite32(m.Top, 0), chartFinite32(m.Right, 0), chartFinite32(m.Bottom, 0), chartFinite32(m.Left, 0)
	w, h := r.W-m.Left-m.Right, r.H-m.Top-m.Bottom
	return chartPolarBox{r.X + m.Left + w/2, r.Y + m.Top + h/2, max(0, min(float32(math.Abs(float64(w))), float32(math.Abs(float64(h))))/2)}
}
func chartPolarPolygon(cx, cy, r float32, angles []float64) *ui.Path {
	path := &ui.Path{}
	for i, a := range angles {
		x, y := chartPolarPoint(cx, cy, r, a)
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	return path.Close()
}
func chartPolarGridPath(kind ChartPolarGridKind, box chartPolarBox, radius float32, angles []float64) *ui.Path {
	if kind == ChartPolarCircle {
		return (&ui.Path{}).Circle(box.cx, box.cy, radius)
	}
	return chartPolarPolygon(box.cx, box.cy, radius, angles)
}
func (paint chartSourcePolarPaint) draw(painter *ui.Painter, r ui.Rect) {
	box := chartPolarPaintBox(paint.margin, r)
	if g := paint.grid; g != nil {
		for i, radius := range g.radii {
			rad := radius.resolve(box.maximum)
			if rad <= 0 {
				continue
			}
			path := chartPolarGridPath(g.kind, box, rad, g.angles)
			fill := g.fill
			if i < len(g.fills) {
				fill = g.fills[i]
			}
			if fill.A > 0 {
				painter.FillPath(path, fill)
			}
			if g.width > 0 && g.stroke.A > 0 {
				painter.StrokePath(path, g.width, g.stroke)
			}
		}
		if g.radial {
			for _, angle := range g.angles {
				x0, y0 := chartPolarPoint(box.cx, box.cy, paint.inner.resolve(box.maximum), angle)
				x1, y1 := chartPolarPoint(box.cx, box.cy, paint.outer.resolve(box.maximum), angle)
				painter.Line(x0, y0, x1, y1, g.width, g.stroke)
			}
		}
	}
	if paint.axisLine {
		painter.StrokePath(chartPolarPolygon(box.cx, box.cy, paint.outer.resolve(box.maximum), paint.angles), 1, paint.axisColor)
	}
	if paint.axisTickLines {
		for _, angle := range paint.angles {
			x0, y0 := chartPolarPoint(box.cx, box.cy, paint.outer.resolve(box.maximum), angle)
			x1, y1 := chartPolarPoint(box.cx, box.cy, paint.outer.resolve(box.maximum)+paint.tickSize, angle)
			painter.Line(x0, y0, x1, y1, 1, paint.axisColor)
		}
	}
	for _, radar := range paint.radar {
		path := &ui.Path{}
		complete, inRun := len(radar.points) >= 3, false
		for _, point := range radar.points {
			if !point.valid {
				complete = false
				inRun = false
				continue
			}
			x, y := box.cx+float32(point.v[0]), box.cy+float32(point.v[1])
			if !inRun {
				path.MoveTo(x, y)
				inRun = true
			} else {
				path.LineTo(x, y)
			}
		}
		if complete {
			path.Close()
			if radar.fill.A > 0 {
				painter.FillPath(path, radar.fill)
			}
		}
		if radar.width > 0 && radar.stroke.A > 0 {
			painter.StrokePath(path, radar.width, radar.stroke)
		}
	}
	for _, list := range [][]ChartPolarSector{paint.tracks, paint.layers} {
		for _, sector := range list {
			path := chartPolarSectorPath(box.cx, box.cy, sector.InnerRadius.resolve(box.maximum), sector.OuterRadius.resolve(box.maximum), sector.StartAngle, sector.EndAngle, chartFinite32(sector.CornerRadius, 0))
			if sector.Color.A > 0 {
				painter.FillPath(path, sector.Color)
			}
			if sector.StrokeWidth > 0 && finite(float64(sector.StrokeWidth)) && sector.Stroke.A > 0 {
				painter.StrokePath(path, sector.StrokeWidth, sector.Stroke)
			}
		}
	}
	for _, label := range paint.labelLines {
		sector := label.sector
		angle := (sector.StartAngle + sector.EndAngle) / 2
		x0, y0 := chartPolarPoint(box.cx, box.cy, sector.OuterRadius.resolve(box.maximum), angle)
		x1, y1 := chartPolarPoint(box.cx, box.cy, sector.OuterRadius.resolve(box.maximum)+label.offset, angle)
		painter.Line(x0, y0, x1, y1, 1, sector.Color)
	}
}

func chartSourcePolarHit(p ChartProps, frame chartPolarFrame, layers []ChartPolarSector, x, y float32, r ui.Rect) (int, int) {
	box := chartSourcePolarBox(p, ui.Rect{W: r.W, H: r.H})
	if p.Kind == RadialChart {
		inner, outer := chartPolarRadii(p, 0)
		lo, hi := inner.resolve(box.maximum), outer.resolve(box.maximum)
		distance := float32(math.Hypot(float64(x-box.cx), float64(y-box.cy)))
		start, end := chartPolarAngles(p)
		angle := math.Atan2(float64(box.cy-y), float64(x-box.cx)) * 180 / math.Pi
		if distance <= 0 || distance < lo || distance > hi || hi <= lo || !chartPolarAngleContains(start, end, angle) || len(p.Labels) == 0 {
			return -1, -1
		}
		// Source RadialBarChart defaults to axis inspection: the entire
		// category band is active, including the blank portion of its track.
		index := min(len(p.Labels)-1, int((distance-lo)/(hi-lo)*float32(len(p.Labels))))
		return 0, index
	}
	if p.Kind == RadarChart {
		inner, outer := chartPolarRadii(p, 0)
		distance := float32(math.Hypot(float64(x-box.cx), float64(y-box.cy)))
		if distance < inner.resolve(box.maximum) || distance > outer.resolve(box.maximum) || distance <= 0 {
			return -1, -1
		}
		angle := math.Atan2(float64(box.cy-y), float64(x-box.cx)) * 180 / math.Pi
		start, end := chartPolarAngles(p)
		step := (end - start) / float64(max(1, len(p.Labels)))
		if step == 0 {
			return -1, -1
		}
		best, index := math.Inf(1), -1
		for i := range p.Labels {
			a := start + step*float64(i)
			distance := math.Abs(math.Remainder(angle-a, 360))
			if distance < best {
				best, index = distance, i
			}
		}
		return 0, index
	}
	for i := len(layers) - 1; i >= 0; i-- {
		s := layers[i]
		path := chartPolarSectorPath(box.cx, box.cy, s.InnerRadius.resolve(box.maximum), s.OuterRadius.resolve(box.maximum), s.StartAngle, s.EndAngle, s.CornerRadius)
		if chartPolarSectorContains(path, s, x, y) {
			return s.Datum.Series, s.Datum.Index
		}
	}
	return -1, -1
}

func chartPolarSectorContains(path *ui.Path, s ChartPolarSector, x, y float32) bool {
	if path.Contains(x, y) {
		return true
	}
	if s.StrokeWidth <= 0 || !finite(float64(s.StrokeWidth)) || s.Stroke.A == 0 {
		return false
	}
	// Source SVG item hit testing includes its painted stroke. Sampling a
	// half-stroke disk against the public fill path retains rounded corners
	// without reading Path internals (subpixel boundary error is <0.05 DIP
	// for the source's maximum5-DIP stroke).
	r := s.StrokeWidth / 2
	for i := 0; i < 16; i++ {
		angle := 2 * math.Pi * float64(i) / 16
		if path.Contains(x+r*float32(math.Cos(angle)), y+r*float32(math.Sin(angle))) {
			return true
		}
	}
	return false
}

func chartPolarAngleContains(start, end, angle float64) bool {
	span := math.Abs(end - start)
	if span >= 360 {
		return true
	}
	delta := math.Mod((angle-start)*math.Copysign(1, end-start), 360)
	if delta < 0 {
		delta += 360
	}
	return delta <= span+1e-7
}
