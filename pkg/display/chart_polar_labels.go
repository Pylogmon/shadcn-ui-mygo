package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type chartPolarLabelKey struct{ Series, Index, Label int }

func chartPolarAnchor(c *ui.Context, bounds ui.Rect, x, y float32, h, v ui.Align, build func()) ui.Element {
	return chart.Anchor(c, x/max(1, bounds.W), 1-y/max(1, bounds.H), h, v, build)
}
func chartPolarNativeContent(c *ui.Context, p ChartProps, frame chartPolarFrame, selected, selectedSeries int, bounds ui.Rect) {
	t := theme.TokensOf(c)
	box := chartSourcePolarBox(p, ui.Rect{W: bounds.W, H: bounds.H})
	start, end := chartPolarAngles(p)
	inner, outer := chartPolarRadii(p, 0)
	if axis := p.Polar.AngleAxis; axis != nil && !axis.HideLabels {
		color := t.MutedForeground
		if axis.Color != nil {
			color = *axis.Color
		}
		radius := outer.resolve(box.maximum) + chartNonnegative(axis.TickSize, 8)
		for i, label := range p.Labels {
			angle := start + (end-start)*float64(i)/float64(max(1, len(p.Labels)))
			x, y := chartPolarPoint(box.cx, box.cy, radius, angle)
			h := ui.Center
			cos := math.Cos(angle * math.Pi / 180)
			if cos > 1e-5 {
				h = ui.Start
			} else if cos < -1e-5 {
				h = ui.End
			}
			values := make([]float64, len(p.Series))
			for j, s := range p.Series {
				if i < len(s.Values) {
					values[j] = s.Values[i]
				}
			}
			chartPolarAnchor(c.Key(chartPolarLabelKey{-1, i, 0}), bounds, x, y, h, ui.Center, func() {
				ui.Box(c).FontSize(12).TextColor(color).Children(func() {
					if axis.Content != nil {
						axis.Content(c, ChartPolarTick{Index: i, Label: label, Values: values, Angle: angle, TextAlign: h})
					} else {
						ui.Text(c, label).NoWrap()
					}
				})
			}).IgnoreInput(true)
		}
	}
	if axis := p.Polar.RadiusAxis; axis != nil && p.Kind == RadarChart {
		axisAngle := axis.Angle
		if !finite(axisAngle) {
			axisAngle = 0
		}
		color := t.MutedForeground
		if axis.Color != nil {
			color = *axis.Color
		}
		if !axis.HideLine {
			x0, y0 := chartPolarPoint(box.cx, box.cy, inner.resolve(box.maximum), axisAngle)
			x1, y1 := chartPolarPoint(box.cx, box.cy, outer.resolve(box.maximum), axisAngle)
			ui.Box(c).Absolute().Fill().IgnoreInput(true).Draw(func(p *ui.Painter, r ui.Rect) { p.Line(r.X+x0, r.Y+y0, r.X+x1, r.Y+y1, 1, color) })
		}
		if !axis.HideLabels {
			for i := 0; i < 5; i++ {
				value := frame.maximum * float64(i) / 4
				radius := chartRadiusMix(inner, outer, float32(i)/4).resolve(box.maximum)
				x, y := chartPolarPoint(box.cx, box.cy, radius, axisAngle)
				text := chartFormat(value)
				if axis.Formatter != nil {
					text = axis.Formatter(value)
				}
				chartPolarAnchor(c.Key(chartPolarLabelKey{-2, i, 0}), bounds, x, y, axis.TextAlign, ui.Center, func() { ui.Text(c, text).NoWrap().FontSize(12).TextColor(color) }).IgnoreInput(true)
			}
		}
	}
	if p.Kind == RadarChart {
		for j, points := range frame.radar {
			color := chartColor(c, p, j)
			style := p.Series[j].Style
			for i, point := range points {
				if !point.valid {
					continue
				}
				normalized := chartGeometryDatum{valid: true, v: [4]float64{float64((box.cx + float32(point.v[0])) / max(1, bounds.W)), float64(1 - (box.cy+float32(point.v[1]))/max(1, bounds.H))}}
				if style.Dots != nil {
					chartBuildDot(c, p, j, i, normalized, color, selected, false, style.Dots)
				}
				if i == selected {
					target := frame.radarTargets[j][i]
					normalized.v[0], normalized.v[1] = float64((box.cx+float32(target.v[0]))/max(1, bounds.W)), float64(1-(box.cy+float32(target.v[1]))/max(1, bounds.H))
					spec := style.ActiveDot
					if spec == nil {
						background := t.Background
						w := float32(2)
						spec = &ChartDotProps{Radius: 4, Stroke: &background, StrokeWidth: &w}
					}
					chartBuildDot(c, p, j, i, normalized, color, selected, true, spec)
				}
			}
		}
	}
	if frame.finished {
		for _, sector := range frame.sectors {
			style := p.Series[sector.Datum.Series].Style.Polar
			if style == nil {
				continue
			}
			for k, label := range style.Labels {
				d := sector.Datum
				d.Active = d.Index == selected
				angle := (sector.StartAngle + sector.EndAngle) / 2
				radius := (sector.InnerRadius.resolve(box.maximum) + sector.OuterRadius.resolve(box.maximum)) / 2
				offset := float32(0)
				h := ui.Center
				if label.Position == ChartPolarLabelOutside {
					offset = chartNonnegative(label.Offset, 20)
					radius = sector.OuterRadius.resolve(box.maximum) + offset
					if math.Cos(angle*math.Pi/180) >= 0 {
						h = ui.Start
					} else {
						h = ui.End
					}
				}
				color := t.Foreground
				if label.Color != nil {
					color = *label.Color
				}
				font := label.FontSize
				if font <= 0 || !finite(float64(font)) {
					font = 12
				}
				text := chartFormat(d.Value)
				if label.Formatter != nil {
					text = label.Formatter(d)
				}
				if label.Position == ChartPolarLabelInsideStart {
					margin := ChartInsets{Top: 5, Right: 5, Bottom: 5, Left: 5}
					if p.Polar.Margin != nil {
						margin = *p.Polar.Margin
					}
					chartPolarCurvedLabel(c.Key(chartPolarLabelKey{d.Series, d.Index, k}), margin, sector, text, font, color, chartNonnegative(label.Offset, 5))
					continue
				}
				x, y := chartPolarPoint(box.cx, box.cy, radius, angle)
				d.X, d.Y = x/max(1, bounds.W), 1-y/max(1, bounds.H)
				chartPolarAnchor(c.Key(chartPolarLabelKey{d.Series, d.Index, k}), bounds, x, y, h, ui.Center, func() {
					ui.Box(c).FontSize(font).TextColor(color).Children(func() {
						if label.Content != nil {
							label.Content(c, d)
						} else {
							ui.Text(c, text).NoWrap()
						}
					})
				}).IgnoreInput(true).AccessHidden(true)
			}
		}
	}
	if p.Polar.Center != nil {
		data := ChartPolarCenter{Selected: selected, SelectedSeries: selectedSeries}
		centerProps := chartPolarInspectionProps(p, selectedSeries)
		if selected >= 0 && selected < len(p.Labels) {
			data.Label = p.Labels[selected]
			if len(centerProps.Series) > 0 && selected < len(centerProps.Series[0].Values) {
				data.Value = centerProps.Series[0].Values[selected]
			}
		}
		if len(centerProps.Series) > 0 {
			series := centerProps.Series[:1]
			if p.Kind == RadialChart && p.Stacked {
				series = centerProps.Series
			}
			for _, s := range series {
				for i, v := range s.Values {
					if i < len(p.Labels) && finite(v) && v > 0 {
						data.Total += v
					}
				}
			}
		}
		chartPolarAnchor(c.Key("polar-center"), bounds, box.cx, box.cy, ui.Center, ui.Center, func() { p.Polar.Center(c, data) })
	}
}

// Shape once using the public native text engine, then place complete glyphs
// along the source insideStart arc. No rune splitting, private font fields or
// application callback is used by Draw; kerning/ligature clusters are retained.
func chartPolarCurvedLabel(c *ui.Context, margin ChartInsets, sector ChartPolarSector, text string, fontSize float32, color ui.Color, offset float32) {
	font := ui.Font{Family: c.Theme().Font, Size: fontSize}
	glyphs := ui.Shape(text, font)
	metrics := font.Metrics()
	baseline := (metrics.Ascent - metrics.Descent) / 2
	// Center and radii are expressed as ratios plus DIP offsets; resize therefore
	// updates the curve even before the following build observes the new size.
	ui.Box(c).Absolute().Fill().IgnoreInput(true).AccessHidden(true).Draw(func(p *ui.Painter, r ui.Rect) {
		box := chartPolarPaintBox(margin, r)
		cx, cy := box.cx, box.cy
		radius := (sector.InnerRadius.resolve(box.maximum) + sector.OuterRadius.resolve(box.maximum)) / 2
		if radius <= 0 {
			return
		}
		sign := math.Copysign(1, sector.EndAngle-sector.StartAngle)
		for _, original := range glyphs {
			glyph := original
			angle := sector.StartAngle + sign*(float64(offset)+float64(glyph.X+glyph.Advance/2)/float64(radius)*180/math.Pi)
			x, y := chartPolarPoint(cx, cy, radius, angle)
			rotation := -angle - sign*90
			glyph.X = -glyph.Advance / 2
			p.Transform(ui.Rotation(float32(rotation)), x, y, func() { p.Glyphs([]ui.Glyph{glyph}, x, y+baseline, color) })
		}
	})
}
