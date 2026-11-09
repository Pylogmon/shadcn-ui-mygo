// Package motion contains shadcn's short feedback and presence transitions.
// Every preset honors the desktop's reduced-motion setting and instant frames.
package motion

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

const (
	Fast   = 100 * time.Millisecond
	Normal = 150 * time.Millisecond
	Slow   = 300 * time.Millisecond
)

var EaseOut = ui.CubicBezier(.16, 1, .3, 1)

func Reduced(c *ui.Context) bool { return c.Instant() || c.Preferences().ReduceMotion }

func respect(c *ui.Context, tr ui.ElementTransition) ui.ElementTransition {
	if Reduced(c) {
		return ui.ElementTransition{Duration: -1, ExitDuration: -1, Colors: tr.Colors}
	}
	return tr
}

func Transition(c *ui.Context) ui.ElementTransition {
	return respect(c, ui.ElementTransition{Duration: Normal, Ease: EaseOut, Colors: true})
}

// ApplyTransition applies the standard feedback transition. Reduced motion
// and instant frames skip registration so an in-flight native color transition
// cannot continue after the preference changes. Transition remains available
// for callers that need to customize a preset.
func ApplyTransition(c *ui.Context, element ui.Element) ui.Element {
	if !Reduced(c) {
		element.Transition(Transition(c))
	}
	return element
}
func Popup(c *ui.Context) ui.ElementTransition {
	return respect(c, ui.ElementTransition{Duration: 150 * time.Millisecond, ExitDuration: 100 * time.Millisecond, Ease: EaseOut,
		Enter: &ui.Motion{Scale: .95, Y: -4, Flip: true}, Exit: &ui.Motion{Scale: .95, Y: -4, Flip: true}})
}
func Modal(c *ui.Context) ui.ElementTransition {
	return respect(c, ui.ElementTransition{Duration: 200 * time.Millisecond, ExitDuration: 150 * time.Millisecond, Ease: EaseOut,
		Enter: &ui.Motion{Scale: .95}, Exit: &ui.Motion{Scale: .95}})
}
func Backdrop(c *ui.Context) ui.ElementTransition {
	return respect(c, ui.ElementTransition{Duration: 200 * time.Millisecond, Enter: &ui.Motion{}, Exit: &ui.Motion{}})
}
func Collapse(c *ui.Context) ui.ElementTransition {
	return respect(c, ui.ElementTransition{Duration: 200 * time.Millisecond, Ease: EaseOut, Enter: &ui.Motion{Collapse: true}, Exit: &ui.Motion{Collapse: true}})
}

type Side uint8

const (
	Bottom Side = iota
	Top
	Left
	Right
)

func Drawer(c *ui.Context, side Side) ui.ElementTransition {
	m := ui.Motion{Opacity: 1}
	switch side {
	case Top:
		m.YPercent = -100
	case Left:
		m.XPercent = -100
	case Right:
		m.XPercent = 100
	default:
		m.YPercent = 100
	}
	exit := m
	return respect(c, ui.ElementTransition{Duration: 300 * time.Millisecond, ExitDuration: 200 * time.Millisecond, Ease: EaseOut, Enter: &m, Exit: &exit})
}
func Toast(c *ui.Context) ui.ElementTransition {
	return respect(c, ui.ElementTransition{Duration: 200 * time.Millisecond, Ease: EaseOut, Enter: &ui.Motion{Y: 20}, Exit: &ui.Motion{Y: 20}})
}
func Thumb(c *ui.Context) ui.Animation {
	if Reduced(c) {
		return ui.Animation{Duration: -1}
	}
	return ui.Animation{Duration: 150 * time.Millisecond, Ease: EaseOut}
}

// Menu is the fixed Base/Vega menu CSS animation profile. Translucent Create
// presets override animation entirely; ordinary Popover remains independent.
func Menu(c *ui.Context) ui.ElementTransition {
	if !theme.MenuOf(c).Animated() {
		return ui.ElementTransition{Duration: -1, ExitDuration: -1}
	}
	return respect(c, ui.ElementTransition{Duration: 100 * time.Millisecond, ExitDuration: 100 * time.Millisecond, Ease: ui.CubicBezier(.25, .1, .25, 1), Enter: &ui.Motion{Scale: .95, Y: -8, Flip: true}, Exit: &ui.Motion{Scale: .95}})
}
