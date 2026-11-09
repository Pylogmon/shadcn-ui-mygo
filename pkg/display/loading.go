package display

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type SpinnerProps struct {
	// Zero Size uses the source's 16 DIPs. Fluent Size/Width/Height still apply.
	Size               float32
	Label, Description string
	// Static keeps the loading cue without rotation. Reduced motion and instant
	// frames always render this same stable cue, including unrelated rebuilds.
	Static, Decorative bool
}

func Spinner(c *ui.Context, props ...SpinnerProps) ui.Element {
	p := SpinnerProps{}
	if len(props) > 0 {
		p = props[0]
	}
	return indicator.Spinner(c, indicator.Props{Size: p.Size, Label: p.Label, Description: p.Description, Static: p.Static, Decorative: p.Decorative})
}

type SkeletonProps struct {
	Width, Height  float32
	Circle, Static bool
}

type skeletonLoopKey struct{}

var skeletonPulseCurve = ui.Bounce(ui.CubicBezier(.4, 0, .6, 1))

func Skeleton(c *ui.Context, p SkeletonProps) ui.Element {
	t := theme.TokensOf(c)
	h, w := p.Height, p.Width
	invalid := false
	if h == 0 {
		h = 16
	} else if h < 0 || math.IsNaN(float64(h)) || math.IsInf(float64(h), 0) {
		h, invalid = 16, true
	}
	if w < 0 || math.IsNaN(float64(w)) || math.IsInf(float64(w), 0) {
		w, invalid = 0, true
	}
	e := ui.Box(c).Height(h).Radius(t.RadiusMD()).Background(t.Muted).Invalid(invalid)
	if w > 0 {
		e.Width(w)
	} else {
		e.FillWidth()
	}
	if p.Circle {
		e.Width(h).Radius(h / 2).Shrink(0)
	}
	if !p.Static && !motion.Reduced(c) {
		e.Opacity(1 - .5*e.Loop(skeletonLoopKey{}, 2*time.Second, skeletonPulseCurve))
	}
	return e
}
