package control

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

// Fixed Base/Vega content utilities: duration-100, CSS ease, fade/zoom-95,
// and side-specific slide-in-*-2. The closed target has no translation.
const comboMotionDuration = 100 * time.Millisecond

var comboMotionEase = ui.CubicBezier(.25, .1, .25, 1)

type comboSide uint8

const (
	comboSideUnknown comboSide = iota
	comboSideBottom
	comboSideTop
	comboSideLeft
	comboSideRight
)

type comboAttachment struct {
	side comboSide
	x, y float32
}

// Only observed, separated boxes prove placement. In particular an attachment
// requested below the trigger is not evidence that collision layout put it there.
func comboActualAttachment(anchor, panel ui.Rect) comboAttachment {
	a := comboAttachment{x: .5, y: .5}
	for _, v := range []float32{anchor.X, anchor.Y, anchor.W, anchor.H, panel.X, panel.Y, panel.W, panel.H} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return a
		}
	}
	if anchor.W < 0 || anchor.H < 0 || panel.W <= 0 || panel.H <= 0 {
		return a
	}
	n := 0
	if panel.Y >= anchor.Y+anchor.H {
		a.side = comboSideBottom
		n++
	}
	if panel.Y+panel.H <= anchor.Y {
		a.side = comboSideTop
		n++
	}
	if panel.X+panel.W <= anchor.X {
		a.side = comboSideLeft
		n++
	}
	if panel.X >= anchor.X+anchor.W {
		a.side = comboSideRight
		n++
	}
	if n != 1 {
		return comboAttachment{x: .5, y: .5}
	}
	// Base's zero-size positioning arrow projects the anchor center, with
	// default arrowPadding=5. Its facing-axis origin includes the actual gap.
	if a.side == comboSideTop || a.side == comboSideBottom {
		pad := min(float32(5), panel.W/2-1)
		a.x = max(pad, min(panel.W-pad, anchor.X+anchor.W/2-panel.X)) / panel.W
		if a.side == comboSideBottom {
			a.y = (anchor.Y + anchor.H - panel.Y) / panel.H
		} else {
			a.y = (anchor.Y - panel.Y) / panel.H
		}
	} else {
		pad := min(float32(5), panel.H/2-1)
		a.y = max(pad, min(panel.H-pad, anchor.Y+anchor.H/2-panel.Y)) / panel.H
		if a.side == comboSideRight {
			a.x = (anchor.X + anchor.W - panel.X) / panel.W
		} else {
			a.x = (anchor.X - panel.X) / panel.W
		}
	}
	return a
}

type comboPose struct{ opacity, scale, x, y float32 }

var comboPoseShown = comboPose{opacity: 1, scale: 1}

func comboPoseHidden(a comboAttachment, entering bool) comboPose {
	p := comboPose{scale: .95}
	if entering {
		switch a.side {
		case comboSideBottom:
			p.y = -8
		case comboSideTop:
			p.y = 8
		case comboSideLeft:
			p.x = 8
		case comboSideRight:
			p.x = -8
		}
	}
	return p
}

type comboMotion struct {
	initialized, entering, shown bool
	from, to                     comboPose
	at                           time.Time
	attachment                   comboAttachment
	anchorBounds, panelBounds    ui.Rect
}

func (s *comboMotion) sample(now time.Time) comboPose {
	f := float32(now.Sub(s.at)) / float32(comboMotionDuration)
	if f >= 1 {
		return s.to
	}
	if f <= 0 {
		return s.from
	}
	f = comboMotionEase(f)
	lerp := func(a, b float32) float32 { return a + (b-a)*f }
	return comboPose{lerp(s.from.opacity, s.to.opacity), lerp(s.from.scale, s.to.scale), lerp(s.from.x, s.to.x), lerp(s.from.y, s.to.y)}
}

func (s *comboMotion) update(c *ui.Context, open bool) (comboPose, bool) {
	if !s.initialized {
		s.attachment = comboAttachment{x: .5, y: .5}
	}
	target := comboPoseHidden(s.attachment, false)
	if open {
		target = comboPoseShown
	}
	if !s.initialized {
		s.initialized = true
		s.from, s.to, s.at = target, target, c.Now()
		if open {
			s.from, s.entering = comboPoseHidden(s.attachment, true), true
		}
	} else if open != s.shown {
		s.from, s.to, s.at = s.sample(c.Now()), target, c.Now()
		s.entering = false
		// A genuinely unmounted popup starts a new entry. A reversal keeps
		// its current native subtree and visual pose instead.
		if open && s.from == comboPoseHidden(s.attachment, false) {
			s.from, s.entering = comboPoseHidden(s.attachment, true), true
		}
	}
	s.shown = open
	if s.entering {
		seed := comboPoseHidden(s.attachment, true)
		s.from.x, s.from.y = seed.x, seed.y
	}
	if motion.Reduced(c) || c.WindowHidden() {
		s.from, s.to, s.entering = target, target, false
	}
	p := s.sample(c.Now())
	if p != s.to {
		c.AnimationFrame()
	} else {
		s.entering = false
	}
	return p, open || p != target
}

func (s *comboMotion) observe(anchor, panel ui.Element) {
	// LayoutBounds also exposes the last completed layout during the view.
	// This matters at entry opacity zero, when the painter can skip Draw.
	if b := panel.LayoutBounds(); b.W > 0 && b.H > 0 {
		s.attachment = comboActualAttachment(anchor.LayoutBounds(), b)
	}
}

func (s *comboMotion) apply(c *ui.Context, anchor, panel ui.Element, p comboPose) {
	panel.TransformOrigin(s.attachment.x, s.attachment.y).Opacity(p.opacity).Scale(p.scale, p.scale).Translate(p.x, p.y)
	// Draw sees this frame's final collision placement. The callback belongs
	// only to this frame; persistent state contains boxes, never UI objects.
	panel.Draw(func(_ *ui.Painter, _ ui.Rect) {
		a, b := anchor.LayoutBounds(), panel.LayoutBounds()
		next := comboActualAttachment(a, b)
		if next != s.attachment {
			s.attachment = next
			c.Invalidate()
		}
		s.anchorBounds, s.panelBounds = a, b
	})
}
