package display

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

// ChartAnimationProps configures native data geometry animation. Nil Duration
// and Begin use the fixed Recharts 3.8.0 family defaults: Bar 400ms/0, Pie and
// Donut 1500ms/400ms, and Line/Area/Radar/Radial 1500ms/0. Explicit nonpositive
// Duration completes immediately; negative Begin is zero. Nil Ease uses CSS
// ease (.25,.1,.25,1), rather than the native cubic ease-out. Ease is evaluated
// only during the current build and is never retained in state or Draw.
// Reduced motion, Instant, Disabled, a hidden window or an offscreen plot
// finish at the target. Showing it again does not replay unchanged data.
type ChartAnimationProps struct {
	Disabled        bool
	Duration, Begin *time.Duration
	Ease            ui.Easing
}

var chartSourceEase = ui.CubicBezier(.25, .1, .25, 1)

type chartAnimationKey struct{}

// A datum contains only normalized, finite geometry. Its meanings are family
// specific: rectangle edges; x/y/baseline; pie span; or radial start/end.
// Missing measurements have valid=false, never a fabricated zero observation.
type chartGeometryDatum struct {
	v     [4]float64
	valid bool
}
type chartGeometryPlan struct{ points, seeds []chartGeometryDatum }
type chartGeometryFrame struct {
	points           []chartGeometryDatum
	progress, reveal float64
	lineStart        float64
	state            *chartSeriesAnimation
}
type chartSeriesAnimation struct {
	from, current, target             []chartGeometryDatum
	start                             time.Time
	duration, begin                   time.Duration
	raw, progress, reveal, fromReveal float64
	longest, lineStart                float64
	active                            bool
}
type chartGeometryAnimation struct {
	family                       ChartKind
	initialized, visibilityKnown bool
	series                       []chartSeriesAnimation
}

func (s *chartGeometryAnimation) Close() error { clear(s.series); s.series = nil; return nil }

func chartAnimationSettings(kind ChartKind, p ChartAnimationProps) (time.Duration, time.Duration, ui.Easing) {
	duration := 1500 * time.Millisecond
	begin := time.Duration(0)
	if kind == BarChart {
		duration = 400 * time.Millisecond
	}
	if kind == PieChart || kind == DonutChart {
		begin = 400 * time.Millisecond
	}
	if p.Duration != nil {
		duration = max(0, *p.Duration)
	}
	if p.Begin != nil {
		begin = max(0, *p.Begin)
	}
	ease := p.Ease
	if ease == nil {
		ease = chartSourceEase
	}
	return duration, begin, ease
}
func chartGeometryEqual(a, b []chartGeometryDatum) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].valid != b[i].valid {
			return false
		}
		if a[i].valid && a[i].v != b[i].v {
			return false
		}
	}
	return true
}
func chartCopyGeometry(a []chartGeometryDatum) []chartGeometryDatum {
	return append([]chartGeometryDatum(nil), a...)
}
func chartGeometryHasData(a []chartGeometryDatum) bool {
	for _, p := range a {
		if p.valid {
			return true
		}
	}
	return false
}

func chartGeometryDrawable(points []chartGeometryDatum, kind ChartKind) bool {
	if kind == LineChart || kind == AreaChart {
		for i := 1; i < len(points); i++ {
			if points[i-1].valid && points[i].valid {
				return true
			}
		}
		return false
	}
	return chartGeometryHasData(points)
}

func chartGeometryFrom(current []chartGeometryDatum, plan chartGeometryPlan, kind ChartKind, lineHadDrawable bool) []chartGeometryDatum {
	from := chartCopyGeometry(plan.seeds)
	filtered := []chartGeometryDatum(nil)
	if kind == BarChart {
		for _, point := range current {
			if point.valid {
				filtered = append(filtered, point)
			}
		}
	}
	ordinal := 0
	for i, target := range plan.points {
		if !target.valid {
			continue
		}
		previous := i
		source := current
		switch kind {
		case BarChart:
			source = filtered
			previous = ordinal
			ordinal++
		case LineChart, AreaChart, RadarChart:
			if len(plan.points) > 0 {
				previous = i * len(current) / len(plan.points)
			}
		}
		if previous >= 0 && previous < len(source) && source[previous].valid {
			from[i] = source[previous]
		} else if kind == LineChart && lineHadDrawable && (previous < 0 || previous >= len(source)) {
			// An empty previous array is different from a cold null ref.
			// After an earlier drawable Line, new unmatched points use the
			// source animateNewValues offscreen seed.
			from[i].v[0], from[i].v[1] = 2, .5
		} else if previous >= 0 && previous < len(source) && (kind == LineChart || kind == AreaChart) {
			// A source point with a missing Y still has its numeric X.
			// Interpolating a missing endpoint returns the new Y directly;
			// it must not invent zero or move the whole point offscreen.
			from[i].v[0] = source[previous].v[0]
		}
	}
	return from
}
func chartLerpGeometry(from, to []chartGeometryDatum, t float64) []chartGeometryDatum {
	if t >= 1 {
		return chartCopyGeometry(to)
	}
	result := chartCopyGeometry(to)
	for i := range result {
		if !result[i].valid {
			continue
		}
		for j := range result[i].v {
			result[i].v[j] = from[i].v[j] + (to[i].v[j]-from[i].v[j])*t
		}
	}
	return result
}
func chartSafeEase(ease ui.Easing, t float64) float64 {
	v := float64(ease(float32(t)))
	if !finite(v) {
		return t
	}
	return min(1, max(0, v))
}

func animateChartGeometry(c *ui.Context, owner ui.Element, kind ChartKind, p ChartAnimationProps, plans []chartGeometryPlan) []chartGeometryFrame {
	state := ui.Local(owner, chartAnimationKey{}, func() chartGeometryAnimation { return chartGeometryAnimation{} })
	// Pie and donut are one source component; changing the hole applies the new
	// target radius immediately while continuing/interpolating its angle spans.
	family := kind
	if family == DonutChart {
		family = PieChart
	}
	if state.initialized && state.family != family {
		state.Close()
		*state = chartGeometryAnimation{}
	}
	first := !state.initialized
	state.initialized = true
	state.family = family
	if len(state.series) > len(plans) {
		clear(state.series[len(plans):])
		state.series = state.series[:len(plans)]
	}
	for len(state.series) < len(plans) {
		state.series = append(state.series, chartSeriesAnimation{})
	}
	duration, begin, ease := chartAnimationSettings(kind, p)
	reduced := motion.Reduced(c) || p.Disabled || duration <= 0
	hidden := c.WindowHidden()
	visible := owner.InView()
	// InView deliberately describes the preceding layout. Probe exactly once
	// before treating false as offscreen, so a visible first mount still enters.
	waiting := first && !visible && !reduced && !hidden
	offscreen := !visible && state.visibilityKnown
	state.visibilityKnown = true
	if waiting {
		c.Invalidate()
	}
	frames := make([]chartGeometryFrame, len(plans))
	now := c.Now()
	for j, plan := range plans {
		s := &state.series[j]
		changed := !chartGeometryEqual(s.target, plan.points)
		settingsChanged := s.duration != duration || s.begin != begin
		if changed || settingsChanged || s.target == nil {
			from := chartGeometryFrom(s.current, plan, kind, s.longest > 0)
			initial := s.target == nil
			if kind == LineChart && s.longest <= 0 {
				initial = true
				from = chartCopyGeometry(plan.points)
			}
			s.from = from
			s.current = chartCopyGeometry(from)
			s.target = chartCopyGeometry(plan.points)
			s.start = now
			s.duration = duration
			s.begin = begin
			s.raw = 0
			s.progress = 0
			s.lineStart = s.longest
			s.fromReveal = s.reveal
			if initial {
				s.fromReveal = 0
			}
			// A Line can have unchanged coordinates yet still owe an absolute
			// length reveal after missing-Y history. Do not short-circuit it
			// merely because interpolation snapped those missing Ys to target.
			s.active = chartGeometryDrawable(s.target, kind) && (!chartGeometryEqual(s.from, s.target) || kind == LineChart && changed || (kind == LineChart || kind == AreaChart) && s.fromReveal < 1)
			// Pie/Radial/Radar seeds already express their geometric entrance.
			if !s.active {
				s.reveal = 1
				s.progress = 1
				s.current = chartCopyGeometry(s.target)
			}
		}
		if reduced || hidden || offscreen || !chartGeometryDrawable(s.target, kind) {
			s.active = false
			s.current = chartCopyGeometry(s.target)
			s.from = nil
			s.progress = 1
			s.reveal = 1
			s.raw = 1
		} else if s.active {
			elapsed := now.Sub(s.start) - s.begin
			raw := min(1, max(0, float64(elapsed)/float64(s.duration)))
			if raw >= 1 {
				s.progress = 1
				s.active = false
				s.from = nil
				s.current = chartCopyGeometry(s.target)
				s.reveal = 1
			} else {
				// Re-evaluate only the current easing. Normalized remaining increments
				// preserve the displayed pose even when callers change Ease mid-flight;
				// no application function needs to survive this synchronous build.
				before, after := chartSafeEase(ease, s.raw), chartSafeEase(ease, raw)
				if raw > s.raw && before < 1 {
					s.progress += (1 - s.progress) * min(1, max(0, (after-before)/(1-before)))
				}
				s.current = chartLerpGeometry(s.from, s.target, s.progress)
				s.reveal = s.fromReveal + (1-s.fromReveal)*s.progress
			}
			s.raw = raw
			if s.active && !waiting {
				if elapsed < 0 {
					c.After(-elapsed)
				} else {
					c.AnimationFrame()
				}
			}
		}
		frames[j] = chartGeometryFrame{points: s.current, progress: s.progress, reveal: s.reveal, lineStart: s.lineStart, state: s}
	}
	return frames
}

// A native Path has no public length getter. Integrate the public monotone
// curve geometry independently so interrupted Line entrances can retain the
// longest revealed absolute length, rather than restarting a relative trim.
// Adaptive subdivision bounds the control-polygon/chord difference to .001
// DIP per cubic (the browser/native flatteners also have finite precision).
func chartCurveLength(r ui.Rect, points []chartGeometryDatum) float64 {
	total := 0.0
	for start := 0; start < len(points); {
		for start < len(points) && !points[start].valid {
			start++
		}
		end := start
		for end < len(points) && points[end].valid {
			end++
		}
		if end-start > 1 {
			coords := make([][2]float64, end-start)
			for i := range coords {
				p := points[start+i]
				coords[i] = [2]float64{p.v[0] * float64(r.W), (1 - p.v[1]) * float64(r.H)}
			}
			n := len(coords)
			slopes := make([]float64, n)
			slope := func(a, b [2]float64) float64 {
				if b[0] == a[0] {
					return 0
				}
				return (b[1] - a[1]) / (b[0] - a[0])
			}
			if n >= 3 {
				for i := 1; i < n-1; i++ {
					h0, h1 := coords[i][0]-coords[i-1][0], coords[i+1][0]-coords[i][0]
					if h0 == 0 || h1 == 0 {
						continue
					}
					s0, s1 := slope(coords[i-1], coords[i]), slope(coords[i], coords[i+1])
					if math.Signbit(s0) != math.Signbit(s1) || s0 == 0 || s1 == 0 {
						continue
					}
					value := 2 * min(math.Abs(s0), math.Abs(s1), .5*math.Abs((s0*h1+s1*h0)/(h0+h1)))
					if s0 < 0 {
						value = -value
					}
					slopes[i] = value
				}
				slopes[0] = (3*slope(coords[0], coords[1]) - slopes[1]) / 2
				slopes[n-1] = (3*slope(coords[n-2], coords[n-1]) - slopes[n-2]) / 2
			}
			for i := 1; i < n; i++ {
				a, b := coords[i-1], coords[i]
				if n < 3 {
					total += math.Hypot(b[0]-a[0], b[1]-a[1])
					continue
				}
				dx := (b[0] - a[0]) / 3
				c1, c2 := [2]float64{a[0] + dx, a[1] + dx*slopes[i-1]}, [2]float64{b[0] - dx, b[1] - dx*slopes[i]}
				total += chartCubicLength(a, c1, c2, b, 0)
			}
		}
		start = max(end, start+1)
	}
	return total
}
func chartCubicLength(a, b, c, d [2]float64, depth int) float64 {
	distance := func(a, b [2]float64) float64 { return math.Hypot(a[0]-b[0], a[1]-b[1]) }
	chord := distance(a, d)
	polygon := distance(a, b) + distance(b, c) + distance(c, d)
	if depth >= 12 || polygon-chord < .001 {
		return (polygon + chord) / 2
	}
	mid := func(a, b [2]float64) [2]float64 { return [2]float64{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2} }
	ab, bc, cd := mid(a, b), mid(b, c), mid(c, d)
	abc, bcd := mid(ab, bc), mid(bc, cd)
	center := mid(abc, bcd)
	return chartCubicLength(a, ab, abc, center, depth+1) + chartCubicLength(center, bcd, cd, d, depth+1)
}

func cartesianAnimationPlans(p ChartProps, plot ChartProps, x chart.Band, y chart.Linear, stacked [][]chart.Segment) []chartGeometryPlan {
	plans := make([]chartGeometryPlan, len(plot.Series))
	for j, series := range plot.Series {
		points := make([]chartGeometryDatum, len(p.Labels))
		seeds := make([]chartGeometryDatum, len(p.Labels))
		for i := range points {
			points[i].v[0] = float64(x.At(i))
			seeds[i] = points[i]
			if i >= len(series.Values) || !finite(series.Values[i]) {
				continue
			}
			if p.Kind == BarChart {
				from, to := 0.0, series.Values[i]
				band := x.Band(i)
				if len(stacked) > 0 {
					from, to = stacked[j][i].Lo, stacked[j][i].Hi
					if series.Values[i] < 0 {
						from, to = to, from
					}
				} else {
					width := (band.Hi - band.Lo) / float32(max(len(p.Series), 1))
					band.Lo += float32(j) * width
					band.Hi = band.Lo + width
				}
				xSpan, ySpan := band, y.Span(from, to)
				if horizontalBars(p) {
					xSpan, ySpan = ySpan, horizontalBarBand(band)
				}
				point := chartGeometryDatum{valid: true, v: [4]float64{float64(xSpan.Lo), float64(xSpan.Hi), float64(ySpan.Lo), float64(ySpan.Hi)}}
				// Recharts filters zero-sized rectangles before geometric matching.
				point.valid = point.v[0] != point.v[1] && point.v[2] != point.v[3]
				seed := point
				if horizontalBars(p) {
					seed.v[0], seed.v[1] = float64(y.At(0)), float64(y.At(0))
				} else {
					seed.v[2], seed.v[3] = float64(y.At(0)), float64(y.At(0))
				}
				points[i], seeds[i] = point, seed
			} else {
				hi, lo := y.At(series.Values[i]), y.At(0)
				if len(stacked) > 0 {
					hi, lo = y.At(stacked[j][i].Hi), y.At(stacked[j][i].Lo)
				}
				point := chartGeometryDatum{valid: true, v: [4]float64{float64(x.At(i)), float64(hi), float64(lo)}}
				points[i], seeds[i] = point, point
			}
		}
		plans[j] = chartGeometryPlan{points: points, seeds: seeds}
	}
	return plans
}

func chartPathCoordinates(points []chartGeometryDatum) (xs, ys, base []float32) {
	xs = make([]float32, len(points))
	ys = make([]float32, len(points))
	base = make([]float32, len(points))
	for i, point := range points {
		xs[i], ys[i], base[i] = float32(point.v[0]), float32(point.v[1]), float32(point.v[2])
		if !point.valid {
			ys[i] = float32(math.NaN())
			base[i] = ys[i]
		}
	}
	return
}

func drawAnimatedCartesian(painter *ui.Painter, r ui.Rect, frame chartGeometryFrame, isArea bool, color ui.Color) {
	drawStyledCartesian(painter, r, frame, isArea, chartCartesianPaint{curve: ChartCurveMonotone, fill: color.Alpha(.2), stroke: color, width: 2})
}

func drawStyledCartesian(painter *ui.Painter, r ui.Rect, frame chartGeometryFrame, isArea bool, paint chartCartesianPaint) {
	draw := func() {
		if isArea {
			path := chartCartesianPath(r, frame.points, paint.curve, true)
			if paint.gradient != nil {
				painter.FillPathGradient(path, *paint.gradient)
			} else {
				painter.FillPath(path, paint.fill)
			}
			if paint.width > 0 {
				painter.StrokePath(chartCartesianPath(r, frame.points, paint.curve, false), paint.width, paint.stroke)
			}
			return
		}
		total := chartCurveLengthFor(r, frame.points, paint.curve)
		if total <= 0 {
			return
		}
		budget := total
		if frame.progress < 1 {
			budget = min(total, frame.lineStart+total*frame.progress)
		}
		frame.state.longest = max(frame.state.longest, budget)
		if budget <= 0 {
			return
		}
		path := chartCartesianPath(r, frame.points, paint.curve, false)
		// Public Path.Dash restarts at every subpath, like SVG. Its gap is
		// the current total length, at least as long as any individual run,
		// so a reveal cannot repeat inside that run. Recomputing this length
		// also avoids Recharts 3.8's stale-measurement repetition gaps.
		if budget < total {
			path.Dash(0, float32(budget), float32(total))
		}
		if paint.width > 0 {
			painter.StrokePath(path, paint.width, paint.stroke)
		}
	}
	if isArea && frame.reveal < 1 && len(frame.points) > 0 {
		start, end := frame.points[0].v[0], frame.points[len(frame.points)-1].v[0]
		width := math.Abs(end-start) * frame.reveal * float64(r.W)
		left := float64(r.X) + start*float64(r.W)
		if start > end {
			left -= width
		}
		painter.Clip(ui.Rect{X: float32(left), Y: r.Y - 2, W: float32(width), H: r.H + 4}, 0, draw)
	} else {
		draw()
	}
}

func animatePolarPaint(c *ui.Context, area ui.Element, p ChartProps, paint *polarPaintSnapshot) {
	var plans []chartGeometryPlan
	switch p.Kind {
	case PieChart, DonutChart:
		plans = make([]chartGeometryPlan, len(paint.pies))
		for j, ring := range paint.pies {
			hasPositive := false
			for _, value := range ring.values {
				hasPositive = hasPositive || value > 0
			}
			points := make([]chartGeometryDatum, len(ring.slices))
			seeds := make([]chartGeometryDatum, len(points))
			for i, slice := range ring.slices {
				point := chartGeometryDatum{valid: hasPositive, v: [4]float64{max(0, slice.End-slice.Start-slice.Pad)}}
				points[i] = point
				seeds[i] = point
				seeds[i].v[0] = 0
			}
			plans[j] = chartGeometryPlan{points: points, seeds: seeds}
		}
	case RadialChart:
		count := len(paint.radialValues)
		series := 1
		if len(paint.radialStacks) > 0 {
			series = len(paint.radialStacks)
		}
		plans = make([]chartGeometryPlan, series)
		for j := range plans {
			points := make([]chartGeometryDatum, count)
			seeds := make([]chartGeometryDatum, count)
			for i, value := range paint.radialValues {
				start, end := 0.0, 2*math.Pi*min(1, value/paint.maximum)
				if len(paint.radialStacks) > 0 {
					span := paint.radialStacks[j][i]
					start, end = 2*math.Pi*min(1, span.Lo/paint.maximum), 2*math.Pi*min(1, span.Hi/paint.maximum)
				}
				points[i] = chartGeometryDatum{valid: true, v: [4]float64{start, end}}
				seeds[i] = points[i]
				seeds[i].v[1] = start
			}
			plans[j] = chartGeometryPlan{points: points, seeds: seeds}
		}
	case RadarChart:
		plans = make([]chartGeometryPlan, len(paint.radarValues))
		for j, values := range paint.radarValues {
			points := make([]chartGeometryDatum, paint.count)
			seeds := make([]chartGeometryDatum, paint.count)
			for i := range points {
				if paint.count < 3 || i >= len(values) || !finite(values[i]) || values[i] < 0 {
					continue
				}
				angle := 2 * math.Pi * float64(i) / float64(paint.count)
				radius := min(1, max(0, values[i]/paint.maximum))
				points[i] = chartGeometryDatum{valid: true, v: [4]float64{math.Sin(angle) * radius, -math.Cos(angle) * radius}}
				seeds[i] = chartGeometryDatum{valid: true}
			}
			plans[j] = chartGeometryPlan{points: points, seeds: seeds}
		}
	}
	frames := animateChartGeometry(c, area, p.Kind, p.Animation, plans)
	switch p.Kind {
	case PieChart, DonutChart:
		for j := range paint.pies {
			cursor := 0.0
			for i, slice := range paint.pies[j].slices {
				span := frames[j].points[i].v[0]
				if !frames[j].points[i].valid {
					span = 0
				}
				paint.pies[j].slices[i] = chart.Slice{Start: cursor, End: cursor + span + slice.Pad, Pad: slice.Pad}
				cursor += span + slice.Pad
			}
		}
	case RadialChart:
		paint.radialAngles = make([][]chart.Slice, len(frames))
		for j, frame := range frames {
			paint.radialAngles[j] = make([]chart.Slice, len(frame.points))
			for i, point := range frame.points {
				paint.radialAngles[j][i] = chart.Slice{Start: point.v[0], End: point.v[1]}
			}
		}
	case RadarChart:
		paint.radarPoints = make([][]chartGeometryDatum, len(frames))
		for j, frame := range frames {
			paint.radarPoints[j] = frame.points
		}
	}
}

func chartRadarGeometryPath(cx, cy, radius float32, points []chartGeometryDatum, closed bool) *ui.Path {
	path := &ui.Path{}
	inRun := false
	for _, point := range points {
		if !point.valid {
			inRun = false
			continue
		}
		x, y := cx+float32(point.v[0])*radius, cy+float32(point.v[1])*radius
		if !inRun {
			path.MoveTo(x, y)
			inRun = true
		} else {
			path.LineTo(x, y)
		}
	}
	if closed {
		path.Close()
	}
	return path
}

func polarAnimatedHit(p ChartProps, paint polarPaintSnapshot, x, y float32, r ui.Rect) int {
	if p.Kind != PieChart && p.Kind != DonutChart {
		return polarHit(p, x, y, r)
	}
	cx, cy, radius := polarGeometry(ui.Rect{W: r.W, H: r.H})
	dx, dy := x-cx, y-cy
	distance := float32(math.Hypot(float64(dx), float64(dy)))
	if radius <= 0 || distance > radius+3 || distance < radius*paint.inner || len(paint.pies) == 0 {
		return -1
	}
	ring := min(len(paint.pies)-1, int((radius-distance)/(radius*(1-paint.inner)/float32(len(paint.pies)))))
	if ring < 0 {
		return -1
	}
	if index, ok := chart.SliceAt(paint.pies[ring].slices, dx, dy, radius*paint.inner, radius+3); ok {
		slice := paint.pies[ring].slices[index]
		if slice.End-slice.Start <= slice.Pad+1e-12 {
			return -1
		}
		return index
	}
	return -1
}
