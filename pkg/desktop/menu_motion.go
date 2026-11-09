package desktop

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// These are the fixed Vega menu utilities, rather than the generic Popover
// preset. Base Menubar composes DropdownMenuContent, including its closed
// utilities; the separate Radix Menubar wrapper has a different class contract.
type menuMotionKind uint8

const (
	menuMotionNone menuMotionKind = iota
	menuMotionPopup
	menuMotionBar
)
const menuMotionDuration = 100 * time.Millisecond

var menuMotionEase = ui.CubicBezier(.25, .1, .25, 1)

type menuSide uint8

const (
	menuSideUnknown menuSide = iota
	menuSideBottom
	menuSideTop
	menuSideLeft
	menuSideRight
)

type menuAttachment struct {
	side        menuSide
	x, y        float32
	originKnown bool
}

// Base UI uses a fake zero-size arrow even when no arrow is rendered. Its
// cross-axis origin points at the anchor center, clamped by arrowPadding; its
// main-axis origin is beyond the facing edge by the actual gap. Only separated
// rectangles prove a side. Overlap is explicitly unknown: no requested side
// is substituted for a compositor's unobservable collision decision.
func menuActualAttachment(anchor, content ui.Rect, padding float32) menuAttachment {
	a := menuAttachment{x: .5, y: .5}
	for _, v := range []float32{anchor.X, anchor.Y, anchor.W, anchor.H, content.X, content.Y, content.W, content.H} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return a
		}
	}
	if anchor.W < 0 || anchor.H < 0 || content.W <= 0 || content.H <= 0 {
		return a
	}
	// Diagonal disjoint rectangles can be explained by two sides. They do
	// not prove which placement the native compositor selected.
	sides := []menuSide{}
	if content.Y >= anchor.Y+anchor.H {
		sides = append(sides, menuSideBottom)
	}
	if content.Y+content.H <= anchor.Y {
		sides = append(sides, menuSideTop)
	}
	if content.X+content.W <= anchor.X {
		sides = append(sides, menuSideLeft)
	}
	if content.X >= anchor.X+anchor.W {
		sides = append(sides, menuSideRight)
	}
	if len(sides) != 1 {
		return a
	}
	return menuAttachmentForSide(anchor, content, sides[0], padding)
}

func menuAttachmentForSide(anchor, content ui.Rect, side menuSide, padding float32) menuAttachment {
	a := menuAttachment{side: side, x: .5, y: .5}
	if side == menuSideUnknown || content.W <= 0 || content.H <= 0 {
		return a
	}
	if a.side == menuSideBottom || a.side == menuSideTop {
		if a.side == menuSideBottom {
			a.y = (anchor.Y + anchor.H - content.Y) / content.H
		} else {
			a.y = (anchor.Y - content.Y) / content.H
		}
		pad := min(padding, content.W/2-1)
		a.x = max(pad, min(content.W-pad, anchor.X+anchor.W/2-content.X)) / content.W
	} else {
		if a.side == menuSideRight {
			a.x = (anchor.X + anchor.W - content.X) / content.W
		} else {
			a.x = (anchor.X - content.X) / content.W
		}
		pad := min(padding, content.H/2-1)
		a.y = max(pad, min(content.H-pad, anchor.Y+anchor.H/2-content.Y)) / content.H
	}
	a.originKnown = true
	return a
}

type menuPose struct{ opacity, scale, x, y float32 }

var menuPoseShown = menuPose{opacity: 1, scale: 1}

func menuPoseHidden(a menuAttachment, enter bool) menuPose {
	p := menuPose{scale: .95}
	if enter {
		switch a.side {
		case menuSideBottom:
			p.y = -8
		case menuSideTop:
			p.y = 8
		case menuSideLeft:
			p.x = 8
		case menuSideRight:
			p.x = -8
		}
	}
	return p
}

type menuTween struct {
	initialized bool
	entering    bool
	from, to    menuPose
	at          time.Time
}

func (s *menuTween) sample(now time.Time) menuPose {
	if !s.initialized {
		return menuPoseShown
	}
	f := float32(now.Sub(s.at)) / float32(menuMotionDuration)
	if f >= 1 {
		return s.to
	}
	if f <= 0 {
		return s.from
	}
	f = menuMotionEase(f)
	lerp := func(a, b float32) float32 { return a + (b-a)*f }
	return menuPose{lerp(s.from.opacity, s.to.opacity), lerp(s.from.scale, s.to.scale), lerp(s.from.x, s.to.x), lerp(s.from.y, s.to.y)}
}
func (s *menuTween) update(c *ui.Context, shown bool, a menuAttachment) menuPose {
	target := menuPoseShown
	if !shown {
		target = menuPoseHidden(a, false)
	}
	if !s.initialized {
		*s = menuTween{initialized: true, entering: shown, from: menuPoseHidden(a, true), to: target, at: c.Now()}
	}
	if target != s.to {
		s.from = s.sample(c.Now())
		s.to = target
		s.at = c.Now()
		s.entering = false
	}
	if s.entering {
		// FitContent may reveal or flip the actual side after the first
		// layout. CSS's data-side entry variables update without restarting
		// the animation; an interrupted/reversed pose must not be reseeded.
		seed := menuPoseHidden(a, true)
		s.from.x, s.from.y = seed.x, seed.y
	}
	if motion.Reduced(c) || c.WindowHidden() || !theme.MenuOf(c).Animated() {
		s.from, s.to = target, target
		s.entering = false
	}
	p := s.sample(c.Now())
	if p != s.to {
		c.AnimationFrame()
	} else {
		s.entering = false
	}
	return p
}
func menuApplyPose(e ui.Element, p menuPose, a menuAttachment) ui.Element {
	if a.originKnown {
		e.TransformOrigin(a.x, a.y)
	}
	return e.Opacity(p.opacity).Scale(p.scale, p.scale).Translate(p.x, p.y)
}

type menuGeometry struct {
	anchor, content ui.Rect
	known           bool
	context         bool
}
type menuFrameKey struct{}
type menuFrame struct {
	active, closing bool
	pose            menuPose
	attachment      menuAttachment
	geometry        menuGeometry
}

func menuFrameOf(c *ui.Context) menuFrame {
	return *ui.Local(c.Root(), menuFrameKey{}, func() menuFrame { return menuFrame{} })
}
func withMenuFrame(c *ui.Context, value menuFrame, build func()) {
	s := ui.Local(c.Root(), menuFrameKey{}, func() menuFrame { return menuFrame{} })
	previous := *s
	*s = value
	defer func() { *s = previous }()
	build()
}

type menuRootMotionKey struct{}
type menuRootMotion struct {
	tween    menuTween
	identity string
}

type menuPanelKey string
type menuDepthKey int
type menuPanelsKey struct{}
type menuPanelPresence struct {
	path      []string
	tween     menuTween
	live      bool
	highlight string
	pose      menuPose
	bounds    ui.Rect
}
type menuPanels struct {
	panels  []*menuPanelPresence
	anchors map[string]ui.Rect
}

func (s *menuPanels) Close() error {
	clear(s.panels)
	s.panels = nil
	clear(s.anchors)
	s.anchors = nil
	return nil
}
func menuPathKey(path []string) string { return fmt.Sprintf("%q", path) }
func (s *menuPanels) sync(state *MenuState, levels int, closing bool) {
	for _, p := range s.panels {
		p.live = false
	}
	for depth := 1; depth < levels; depth++ {
		path := state.Path[:depth]
		var panel *menuPanelPresence
		for _, p := range s.panels {
			if slices.Equal(p.path, path) {
				panel = p
				break
			}
		}
		if panel == nil {
			panel = &menuPanelPresence{path: slices.Clone(path)}
			s.panels = append(s.panels, panel)
		}
		panel.live = !closing
		if depth < len(state.Highlighted) {
			panel.highlight = state.Highlighted[depth]
		}
	}
}
