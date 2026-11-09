// Natural spline algorithm derived from d3-shape, Copyright 2010-2022 Mike Bostock.
// Permission to use, copy, modify, and/or distribute this software for any purpose
// with or without fee is hereby granted, provided that the above copyright notice
// and this permission notice appear in all copies.
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
// FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
// OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
// THIS SOFTWARE.
package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
)

// Natural control points follow d3-shape 3.2.0's fixed natural.js (ISC),
// archived with its license and the official-function oracle in testdata.
// They solve the tridiagonal system separately for X and Y. No native Path
// internals are read, and missing observations are separate spline runs.
func chartNaturalControls(x []float64) (a, b []float64) {
	n := len(x) - 1
	if n <= 0 {
		return nil, nil
	}
	a, b = make([]float64, n), make([]float64, n)
	r := make([]float64, n)
	a[0], b[0], r[0] = 0, 2, x[0]+2*x[1]
	for i := 1; i < n-1; i++ {
		a[i], b[i], r[i] = 1, 4, 4*x[i]+2*x[i+1]
	}
	a[n-1], b[n-1], r[n-1] = 2, 7, 8*x[n-1]+x[n]
	for i := 1; i < n; i++ {
		m := a[i] / b[i-1]
		b[i] -= m
		r[i] -= m * r[i-1]
	}
	a[n-1] = r[n-1] / b[n-1]
	for i := n - 2; i >= 0; i-- {
		a[i] = (r[i] - a[i+1]) / b[i]
	}
	b[n-1] = (x[n] + a[n-1]) / 2
	for i := 0; i < n-1; i++ {
		b[i] = 2*x[i+1] - a[i+1]
	}
	return
}

type chartNaturalSegment struct{ from, c1, c2, to [2]float64 }

func chartNaturalSegments(coords [][2]float64) []chartNaturalSegment {
	if len(coords) < 3 {
		return nil
	}
	x, y := make([]float64, len(coords)), make([]float64, len(coords))
	for i, v := range coords {
		x[i], y[i] = v[0], v[1]
	}
	ax, bx := chartNaturalControls(x)
	ay, by := chartNaturalControls(y)
	out := make([]chartNaturalSegment, len(coords)-1)
	for i := range out {
		out[i] = chartNaturalSegment{coords[i], [2]float64{ax[i], ay[i]}, [2]float64{bx[i], by[i]}, coords[i+1]}
	}
	return out
}

func chartEachRun(points []chartGeometryDatum, fn func([]chartGeometryDatum)) {
	for start := 0; start < len(points); {
		for start < len(points) && !points[start].valid {
			start++
		}
		end := start
		for end < len(points) && points[end].valid {
			end++
		}
		if end > start {
			fn(points[start:end])
		}
		start = max(end, start+1)
	}
}
func chartRunCoordinates(r ui.Rect, points []chartGeometryDatum, baseline, reverse bool) [][2]float64 {
	out := make([][2]float64, len(points))
	for i := range out {
		at := i
		if reverse {
			at = len(points) - 1 - i
		}
		point := points[at]
		y := point.v[1]
		if baseline {
			y = point.v[2]
		}
		out[i] = [2]float64{float64(r.X) + point.v[0]*float64(r.W), float64(r.Y) + (1-y)*float64(r.H)}
	}
	return out
}
func chartAppendNatural(path *ui.Path, coords [][2]float64, connect bool) {
	if len(coords) == 0 {
		return
	}
	x, y := float32(coords[0][0]), float32(coords[0][1])
	if connect {
		path.LineTo(x, y)
	} else {
		path.MoveTo(x, y)
	}
	if len(coords) == 2 {
		path.LineTo(float32(coords[1][0]), float32(coords[1][1]))
		return
	}
	for _, s := range chartNaturalSegments(coords) {
		path.CubeTo(float32(s.c1[0]), float32(s.c1[1]), float32(s.c2[0]), float32(s.c2[1]), float32(s.to[0]), float32(s.to[1]))
	}
}
func chartCartesianPath(r ui.Rect, points []chartGeometryDatum, curve ChartCurve, area bool) *ui.Path {
	if curve != ChartCurveNatural {
		xs, ys, base := chartPathCoordinates(points)
		if area {
			return chart.Area(r, xs, base, ys, chartNativeCurve(curve))
		}
		return chart.Line(r, xs, ys, chartNativeCurve(curve))
	}
	path := &ui.Path{}
	chartEachRun(points, func(run []chartGeometryDatum) {
		chartAppendNatural(path, chartRunCoordinates(r, run, false, false), false)
		if area {
			chartAppendNatural(path, chartRunCoordinates(r, run, true, true), true)
			path.Close()
		}
	})
	return path
}
func chartCurveLengthFor(r ui.Rect, points []chartGeometryDatum, curve ChartCurve) float64 {
	if curve == ChartCurveMonotone || curve > ChartCurveStep {
		return chartCurveLength(r, points)
	}
	total := 0.0
	chartEachRun(points, func(run []chartGeometryDatum) {
		coords := chartRunCoordinates(r, run, false, false)
		if curve == ChartCurveNatural && len(coords) > 2 {
			for _, s := range chartNaturalSegments(coords) {
				total += chartCubicLength(s.from, s.c1, s.c2, s.to, 0)
			}
		} else {
			for i := 1; i < len(coords); i++ {
				dx, dy := coords[i][0]-coords[i-1][0], coords[i][1]-coords[i-1][1]
				if curve == ChartCurveStep {
					total += math.Abs(dx) + math.Abs(dy)
				} else {
					total += math.Hypot(dx, dy)
				}
			}
		}
	})
	return total
}
