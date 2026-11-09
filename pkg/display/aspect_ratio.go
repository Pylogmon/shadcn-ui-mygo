package display

import (
	"math"

	"github.com/egoist/mygo/ui"
)

type AspectRatioProps struct {
	Ratio float32
	// Zero Width fills the available width. Other zero bounds are unset.
	Width, MaxWidth, MaxHeight float32
	// Clip is opt-in, matching the source primitive's visible overflow.
	Clip               bool
	Label, Description string
}

func aspectFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

// AspectRatio retains the original clipped container contract.
func AspectRatio(c *ui.Context, ratio float32, content func()) ui.Element {
	return AspectRatioWith(c, AspectRatioProps{Ratio: ratio, Clip: true}, func(*ui.Context) {
		if content != nil {
			content()
		}
	})
}

// AspectRatioWith keeps content in a stable native container. A ratio that is
// nonpositive, nonfinite or cannot represent a height at the requested/window
// width uses a square and exposes Invalid instead of passing infinity to layout.
func AspectRatioWith(c *ui.Context, p AspectRatioProps, content func(*ui.Context)) ui.Element {
	valid := true
	for _, dimension := range []*float32{&p.Width, &p.MaxWidth, &p.MaxHeight} {
		if !aspectFinite(*dimension) || *dimension < 0 {
			*dimension, valid = 0, false
		}
	}
	width, _ := c.Size()
	if p.Width > 0 {
		width = p.Width
	}
	if p.MaxWidth > 0 {
		width = min(width, p.MaxWidth)
	}
	ratio := p.Ratio
	if !aspectFinite(ratio) || ratio <= 0 || float64(max(1, width))/float64(ratio) > math.MaxFloat32 {
		ratio, valid = 1, false
	}
	root := ui.Box(c).FillWidth().MinWidth(0).AspectRatio(ratio).MaxHeight(math.MaxFloat32).Invalid(!valid)
	if p.Width > 0 {
		root.Width(p.Width)
	}
	if p.MaxWidth > 0 {
		root.MaxWidth(p.MaxWidth)
	}
	if p.MaxHeight > 0 {
		root.MaxHeight(p.MaxHeight)
	}
	if p.Clip {
		root.Clip()
	}
	if p.Label != "" || p.Description != "" {
		root.Role(ui.RoleGroup).Label(p.Label).Description(p.Description)
	}
	return root.Children(func() {
		if content != nil {
			content(c)
		}
	})
}
