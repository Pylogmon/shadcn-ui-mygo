package display

import (
	"fmt"
	"math"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func chartCategoryAt(p ChartProps, at float32) float32 {
	if p.Cartesian.ReversedCategory {
		return 1 - at
	}
	return at
}
func chartCategoryBand(p ChartProps, span chart.Span) chart.Span {
	if p.Cartesian.ReversedCategory {
		return chart.Span{Lo: 1 - span.Hi, Hi: 1 - span.Lo}
	}
	return span
}
func horizontalBars(p ChartProps) bool { return p.Kind == BarChart && p.Orientation == ChartHorizontal }
func chartCategoryFills(p ChartProps) bool {
	return p.Kind == BarChart && len(p.Series) == 1 && len(p.CategoryColors) > 0
}
func barColor(c *ui.Context, p ChartProps, series, category int) ui.Color {
	if chartCategoryFills(p) {
		return categoryColor(c, p, category)
	}
	return chartColor(c, p, series)
}

type chartBarKey struct{ Series, Category int }

func chartBarRadius(p ChartProps) float32 {
	if v := p.Cartesian.BarRadius; v != nil && finite(float64(*v)) && *v >= 0 {
		return *v
	}
	return 3
}

// A band scale grows from zero. Plot Y fractions grow upward, so reflect the
// category band after grouping it to keep both categories and series top down.
func horizontalBarBand(band chart.Span) chart.Span {
	return chart.Span{Lo: 1 - band.Hi, Hi: 1 - band.Lo}
}

func cartesianAxes(c *ui.Context, p ChartProps, band chart.Band, value chart.Linear, direction ui.Direction, scale float64, extent float32) chart.Axes {
	var categoryTicks []chart.Tick
	if !p.Cartesian.HideCategoryAxis {
		categoryTicks = band.Ticks()
	}
	for i := range categoryTicks {
		categoryTicks[i].At = chartCategoryAt(p, categoryTicks[i].At)
		if horizontalBars(p) {
			categoryTicks[i].At = 1 - categoryTicks[i].At
		}
		if p.Cartesian.FormatCategory != nil {
			categoryTicks[i].Label = p.Cartesian.FormatCategory(categoryTicks[i].Label)
		}
	}
	if p.Cartesian.CategoryMinTickGap != nil {
		gap := chartNonnegative(p.Cartesian.CategoryMinTickGap, 0)
		slices.SortStableFunc(categoryTicks, func(a, b chart.Tick) int {
			aa, bb := a.At, b.At
			if horizontalBars(p) {
				aa, bb = 1-aa, 1-bb
			}
			if aa < bb {
				return -1
			}
			if aa > bb {
				return 1
			}
			return 0
		})
		kept := make([]chart.Tick, 0, len(categoryTicks))
		edge := float32(math.Inf(1))
		for i := len(categoryTicks) - 1; i >= 0; i-- {
			tick := categoryTicks[i]
			w, h := ui.ShapeText(tick.Label, ui.Font{Size: 12}, 0).Size()
			at, size := tick.At, w
			if horizontalBars(p) {
				at, size = 1-tick.At, h
			}
			center := at * extent
			if len(kept) == 0 || center+size/2+gap <= edge {
				kept = append(kept, tick)
				edge = center - size/2
			}
		}
		slices.Reverse(kept)
		categoryTicks = kept
	}
	var valueTicks []chart.Tick
	if !p.Cartesian.HideValueAxis {
		valueTicks = value.Ticks(chartTickCount(p))
	}
	axes := chart.Axes{Left: valueTicks, Bottom: categoryTicks}
	if horizontalBars(p) {
		axes.Left, axes.Bottom = categoryTicks, valueTicks
	}
	cap := p.Cartesian.CategoryAxisMaxWidth
	if !finite(float64(cap)) || cap <= 0 {
		cap = 0
		if horizontalBars(p) {
			cap = 60
		}
	}
	axes.Label = func(side chart.Side, tick chart.Tick) {
		categorySide := side == chart.Bottom
		if horizontalBars(p) {
			categorySide = side == chart.Left
		}
		if p.Cartesian.TickContent != nil {
			index := -1
			if categorySide {
				for i := range p.Labels {
					at := chartCategoryAt(p, band.At(i))
					if horizontalBars(p) {
						at = 1 - at
					}
					if math.Abs(float64(at-tick.At)) < 1e-5 {
						index = i
						break
					}
				}
			}
			v := tick.Value
			if !categorySide && !(p.Stacked && p.Normalized && p.Kind != LineChart) {
				v *= scale
			}
			p.Cartesian.TickContent(c, ChartAxisTick{Side: side, Category: categorySide, Index: index, Value: v, Label: tick.Label})
			return
		}
		label := ui.Text(c, tick.Label).NoWrap().Direction(direction).TextColor(theme.TokensOf(c).MutedForeground)
		if categorySide && cap > 0 {
			label.MaxWidth(cap).SingleLine()
		}
	}
	return axes
}

func cartesianValueScale(p ChartProps, scale, lo, hi float64) chart.Linear {
	if lo == hi {
		hi = lo + 1
	}
	value := chart.NewLinear(lo, hi).Nice(chartTickCount(p))
	format := p.Cartesian.FormatValue
	if format == nil {
		format = p.FormatValue
	}
	normalized := p.Stacked && p.Normalized && p.Kind != LineChart
	if scale != 1 && !normalized {
		value.Format = func(v float64) string {
			original := v * scale
			if !finite(original) || v != 0 && original == 0 {
				return fmt.Sprintf("%g × %g", v, scale)
			}
			if format != nil {
				return format(original)
			}
			return chartFormat(original)
		}
	} else {
		value.Format = format
	}
	return value
}

type cartesianInspection struct {
	hovering         bool
	windowX, windowY float32
	last             int
}

// PointerPosition is in untransformed element coordinates. LayoutBounds is
// the matching denominator; Bounds would select the wrong row under Scale.
// Compare pointer motion in window coordinates so a layout/axis change under
// a resting pointer does not overwrite a keyboard or controlled selection.
func inspectCartesian(c *ui.Context, area ui.Element, p ChartProps, selected *int, band chart.Band, valid []int) {
	area.Focusable()
	state := ui.Local(area, "cartesian-inspection", func() cartesianInspection { return cartesianInspection{last: -1} })
	cursor := slices.Index(valid, *selected)
	if cursor < 0 {
		*selected = -1
	}
	horizontal := horizontalBars(p)
	forward, back := ui.KeyRight, ui.KeyLeft
	if horizontal {
		forward, back = ui.KeyDown, ui.KeyUp
	}
	if p.Cartesian.ReversedCategory {
		forward, back = back, forward
	}
	keyboard := false
	move := func(at int) {
		if len(valid) > 0 {
			*selected = valid[max(0, min(len(valid)-1, at))]
			keyboard = true
		}
	}
	switch {
	case len(valid) == 0:
	case area.Shortcut(0, forward):
		move(cursor + 1)
	case area.Shortcut(0, back):
		if cursor < 0 {
			move(len(valid) - 1)
		} else {
			move(cursor - 1)
		}
	case area.Shortcut(0, ui.KeyHome):
		move(0)
	case area.Shortcut(0, ui.KeyEnd):
		move(len(valid) - 1)
	case *selected >= 0 && area.Shortcut(0, ui.KeyEscape):
		*selected = -1
		keyboard = true
	}
	if x, y, over := area.PointerPosition(); over {
		layout := area.LayoutBounds()
		extent, at := layout.W, x
		windowX, windowY := area.LocalToWindow(x, y)
		changed := !state.hovering || math.Abs(float64(windowX-state.windowX)) > .01
		if horizontal {
			extent, at = layout.H, y
			changed = !state.hovering || math.Abs(float64(windowY-state.windowY)) > .01
		}
		if !keyboard && changed && extent > 0 && len(valid) > 0 {
			positions := make([]float32, len(valid))
			for i, index := range valid {
				positions[i] = band.At(index)
			}
			fraction := at / extent
			if p.Cartesian.ReversedCategory {
				fraction = 1 - fraction
			}
			cursor = chart.Nearest(positions, fraction)
			*selected = valid[cursor]
		}
		state.hovering, state.windowX, state.windowY = true, windowX, windowY
	} else if state.hovering {
		state.hovering = false
		*selected = -1
	}
	if *selected >= 0 {
		description := describeChart(c, p, *selected)
		area.Value(description)
		if keyboard && state.last != *selected {
			c.Announce(description)
		}
	}
	state.last = *selected
}
