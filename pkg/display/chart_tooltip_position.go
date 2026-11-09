package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
)

type chartTooltipAnchor struct{ x, y, w, h float32 }
type chartTooltipSize struct{ w, h float32 }

// This is the fixed source getTooltipTranslateXY policy: offset=10, no escape,
// no reverseDirection and no explicit position. Source modules are archived.
func chartTooltipAxisPosition(at, size, extent float32) float32 {
	positive := at + 10
	if positive+size > extent {
		return max(0, at-size-10)
	}
	return max(0, positive)
}
func chartPositionTooltip(c *ui.Context, tip ui.Element, a chartTooltipAnchor) {
	size := ui.Local(tip, "inspection-tooltip-size", func() chartTooltipSize { return chartTooltipSize{128, 48} })
	service := c.Services()
	tip.Draw(func(_ *ui.Painter, r ui.Rect) {
		if math.Abs(float64(size.w-r.W)) > .25 || math.Abs(float64(size.h-r.H)) > .25 {
			size.w, size.h = r.W, r.H
			service.Invalidate()
		}
	})
	tip.Left(chartTooltipAxisPosition(a.x, size.w, a.w)).Top(chartTooltipAxisPosition(a.y, size.h, a.h))
}
func chartTooltipObservedArea(c *ui.Context, area ui.Element) ui.Rect {
	size := ui.Local(area, "inspection-tooltip-viewport", func() chartTooltipSize { return chartTooltipSize{} })
	service := c.Services()
	area.DrawOver(func(_ *ui.Painter, r ui.Rect) {
		if math.Abs(float64(size.w-r.W)) > .25 || math.Abs(float64(size.h-r.H)) > .25 {
			size.w, size.h = r.W, r.H
			service.Invalidate()
		}
	})
	r := area.LayoutBounds()
	if size.w > 0 && size.h > 0 {
		r.W, r.H = size.w, size.h
	}
	return r
}
func chartCartesianTooltipAnchor(c *ui.Context, area, outer ui.Element, p ChartProps, band chart.Band, index int) chartTooltipAnchor {
	r := area.LayoutBounds()
	if p.TooltipFollowInspection {
		r = chartTooltipObservedArea(c, area)
	}
	a := chartTooltipAnchor{x: r.W / 2, y: r.H / 2, w: r.W, h: r.H}
	// Source defaultIndex takes the non-axis coordinate from the complete
	// chart and its axis offset; category axes are outside the plot viewport.
	box := outer.LayoutBounds()
	ox, oy := outer.LocalToWindow(0, 0)
	ax, ay := area.LocalToWindow(0, 0)
	if box.W > 0 && box.H > 0 {
		a.x = (box.W - (ax - ox)) / 2
		a.y = (box.H - (ay - oy)) / 2
	}
	if index >= 0 && index < len(p.Labels) {
		at := chartCategoryAt(p, band.At(index))
		if horizontalBars(p) {
			a.y = at * r.H
		} else {
			a.x = at * r.W
		}
	}
	if x, y, over := area.PointerPosition(); over {
		if horizontalBars(p) {
			a.x = x
		} else {
			a.y = y
		}
	}
	return a
}
func chartSourcePolarTooltipAnchor(c *ui.Context, area ui.Element, p ChartProps, frame chartPolarFrame, index, series int, r ui.Rect) chartTooltipAnchor {
	if p.TooltipFollowInspection {
		r = chartTooltipObservedArea(c, area)
	}
	a := chartTooltipAnchor{x: r.W / 2, y: r.H / 2, w: r.W, h: r.H}
	if x, y, over := area.PointerPosition(); over {
		a.x, a.y = x, y
		return a
	}
	for _, s := range frame.sectors {
		if s.Datum.Index == index && s.Datum.Series == series {
			a.x, a.y = s.Datum.X*r.W, (1-s.Datum.Y)*r.H
			break
		}
	}
	return a
}
