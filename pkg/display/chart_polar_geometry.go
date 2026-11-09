package display

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// ChartRadius resolves Fraction of the available radius plus Offset DIPs.
// This keeps percentages responsive and lets shape callbacks express outer+10
// or an extra ring from outer+12 to outer+25 without retaining layout objects.
type ChartRadius struct{ Fraction, Offset float32 }

func ChartRadiusDIP(v float32) ChartRadius        { return ChartRadius{Offset: v} }
func ChartRadiusPercent(v float32) ChartRadius    { return ChartRadius{Fraction: v / 100} }
func (r ChartRadius) Add(dip float32) ChartRadius { r.Offset += dip; return r }
func (r ChartRadius) resolve(maximum float32) float32 {
	return max(0, chartFinite32(chartFinite32(r.Fraction, 0)*maximum+chartFinite32(r.Offset, 0), 0))
}
func chartRadiusMix(a, b ChartRadius, t float32) ChartRadius {
	return ChartRadius{Fraction: a.Fraction + (b.Fraction-a.Fraction)*t, Offset: a.Offset + (b.Offset-a.Offset)*t}
}

// Angles are source degrees: 0 points right, positive turns counterclockwise.
func chartPolarPoint(cx, cy, r float32, angle float64) (float32, float32) {
	a := angle * math.Pi / 180
	return cx + r*float32(math.Cos(a)), cy - r*float32(math.Sin(a))
}
func chartPolarArc(path *ui.Path, cx, cy, r float32, start, end float64) {
	if r <= 0 || !finite(start) || !finite(end) {
		return
	}
	n := max(1, int(math.Ceil(math.Abs(end-start)/90)))
	for i := range n {
		a := (start + (end-start)*float64(i)/float64(n)) * math.Pi / 180
		b := (start + (end-start)*float64(i+1)/float64(n)) * math.Pi / 180
		k := 4.0 / 3 * math.Tan((b-a)/4)
		x0, y0 := float64(cx)+float64(r)*math.Cos(a), float64(cy)-float64(r)*math.Sin(a)
		x1, y1 := float64(cx)+float64(r)*math.Cos(b), float64(cy)-float64(r)*math.Sin(b)
		path.CubeTo(float32(x0-k*float64(r)*math.Sin(a)), float32(y0-k*float64(r)*math.Cos(a)), float32(x1+k*float64(r)*math.Sin(b)), float32(y1+k*float64(r)*math.Cos(b)), float32(x1), float32(y1))
	}
}
func chartPolarArcBetween(path *ui.Path, cx, cy, r, x0, y0, x1, y1 float32, sign float64) {
	a := math.Atan2(float64(cy-y0), float64(x0-cx)) * 180 / math.Pi
	b := math.Atan2(float64(cy-y1), float64(x1-cx)) * 180 / math.Pi
	if sign > 0 {
		for b < a {
			b += 360
		}
	} else {
		for b > a {
			b -= 360
		}
	}
	chartPolarArc(path, cx, cy, r, a, b)
}

type chartPolarTangent struct {
	cx, cy, circleX, circleY, lineX, lineY float32
	theta                                  float64
}

func chartSectorTangent(cx, cy, r, corner float32, angle, sign float64, inner bool) chartPolarTangent {
	centerR := r - corner
	if inner {
		centerR = r + corner
	}
	if centerR <= 0 {
		return chartPolarTangent{}
	}
	theta := math.Asin(min(1, float64(corner/centerR))) * 180 / math.Pi
	x, y := chartPolarPoint(cx, cy, centerR, angle+sign*theta)
	sx, sy := chartPolarPoint(cx, cy, r, angle+sign*theta)
	lx, ly := chartPolarPoint(cx, cy, centerR*float32(math.Cos(theta*math.Pi/180)), angle)
	return chartPolarTangent{x, y, sx, sy, lx, ly, theta}
}

// Rounded sector tangencies follow the archived fixed Recharts Sector source.
// Only the normal corner mode used by the six source Radial examples is needed;
// forceCornerRadius and external corner APIs are deliberately not generalized.
func chartPolarSectorPath(cx, cy, inner, outer float32, start, end float64, corner float32) *ui.Path {
	path := &ui.Path{}
	if outer <= 0 || outer < inner || start == end || !finite(start) || !finite(end) {
		return path
	}
	delta := math.Copysign(min(math.Abs(end-start), 359.999), end-start)
	end = start + delta
	sign := math.Copysign(1, delta)
	corner = min(max(0, corner), (outer-inner)/2)
	if corner > 0 && math.Abs(delta) < 359.999 {
		a, b := chartSectorTangent(cx, cy, outer, corner, start, sign, false), chartSectorTangent(cx, cy, outer, corner, end, -sign, false)
		if math.Abs(delta) >= a.theta+b.theta {
			path.MoveTo(a.lineX, a.lineY)
			chartPolarArcBetween(path, a.cx, a.cy, corner, a.lineX, a.lineY, a.circleX, a.circleY, sign)
			chartPolarArc(path, cx, cy, outer, start+sign*a.theta, end-sign*b.theta)
			chartPolarArcBetween(path, b.cx, b.cy, corner, b.circleX, b.circleY, b.lineX, b.lineY, sign)
			if inner > 0 {
				u, v := chartSectorTangent(cx, cy, inner, corner, start, sign, true), chartSectorTangent(cx, cy, inner, corner, end, -sign, true)
				path.LineTo(v.lineX, v.lineY)
				chartPolarArcBetween(path, v.cx, v.cy, corner, v.lineX, v.lineY, v.circleX, v.circleY, sign)
				chartPolarArc(path, cx, cy, inner, end-sign*v.theta, start+sign*u.theta)
				chartPolarArcBetween(path, u.cx, u.cy, corner, u.circleX, u.circleY, u.lineX, u.lineY, sign)
			} else {
				path.LineTo(cx, cy)
			}
			return path.Close()
		}
	}
	x, y := chartPolarPoint(cx, cy, outer, start)
	path.MoveTo(x, y)
	chartPolarArc(path, cx, cy, outer, start, end)
	if inner > 0 {
		x, y = chartPolarPoint(cx, cy, inner, end)
		path.LineTo(x, y)
		chartPolarArc(path, cx, cy, inner, end, start)
	} else {
		path.LineTo(cx, cy)
	}
	return path.Close()
}

func chartPolarAngles(p ChartProps) (float64, float64) {
	start, end := 0.0, 360.0
	if p.Kind == RadarChart {
		start, end = 90, -270
	}
	if p.Polar.StartAngle != nil && finite(*p.Polar.StartAngle) {
		start = *p.Polar.StartAngle
	}
	if p.Polar.EndAngle != nil && finite(*p.Polar.EndAngle) {
		end = *p.Polar.EndAngle
	}
	return start, end
}
func chartPolarRadii(p ChartProps, j int) (ChartRadius, ChartRadius) {
	inner, outer := ChartRadiusDIP(0), ChartRadiusPercent(80)
	if p.Kind == DonutChart {
		fraction := p.InnerRadius
		if fraction <= 0 {
			fraction = .6
		}
		inner = ChartRadius{Fraction: min(.9, max(0, fraction))}
	}
	if p.Polar.InnerRadius != nil {
		inner = *p.Polar.InnerRadius
	}
	if p.Polar.OuterRadius != nil {
		outer = *p.Polar.OuterRadius
	}
	if j >= 0 && j < len(p.Series) {
		if s := p.Series[j].Style.Polar; s != nil {
			if s.InnerRadius != nil {
				inner = *s.InnerRadius
			}
			if s.OuterRadius != nil {
				outer = *s.OuterRadius
			}
		}
	}
	return inner, outer
}

// Radius axes extend their domain with the fixed adaptive five-tick algorithm.
// Angular numeric axes explicitly do not extend it (axisSelectors source).
// Values are non-negative in this profile, so zero is always the first tick.
func chartPolarNiceMaximum(value float64) float64 {
	if !finite(value) || value <= 0 {
		return 4
	}
	rough := value / 4
	digits := math.Floor(math.Log10(rough)) + 1
	power := math.Pow(10, digits)
	ratio := .05
	if digits == 1 {
		ratio = .1
	}
	step := math.Ceil(rough/power/ratio-1e-12) * ratio * power
	step = math.Ceil(step)
	result := step * 4
	if result < value || !finite(result) {
		return value
	}
	return result
}
