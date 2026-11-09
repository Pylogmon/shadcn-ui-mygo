// Package overlay provides shadcn styled native dialogs, menus and floating content.
// State belongs to the caller; keyboard handling, focus restoration and accessibility
// are provided by MyGo's native base controls.
package overlay

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type Placement int

const (
	Bottom Placement = iota
	Top
	Left
	Right
)

func attach(panel, anchor ui.Element, side Placement, gap float32) {
	switch side {
	case Top:
		panel.AttachTo(anchor, ui.AnchorTop, ui.AnchorBottom).Margin(0, 0, gap, 0)
	case Left:
		panel.AttachTo(anchor, ui.AnchorLeft, ui.AnchorRight).Margin(0, gap, 0, 0)
	case Right:
		panel.AttachTo(anchor, ui.AnchorRight, ui.AnchorLeft).Margin(0, 0, 0, gap)
	default:
		panel.AttachTo(anchor, ui.AnchorBottomStart, ui.AnchorTopStart).Margin(gap, 0, 0, 0)
	}
	panel.AttachedOrigin()
}

func surface(c *ui.Context, panel ui.Element) ui.Element {
	t := theme.TokensOf(c)
	return panel.Column().Padding(16).Gap(8).Radius(t.Radius).Border(1, t.Border).
		Background(t.Popover).TextColor(t.PopoverForeground).FontSize(14).
		Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, 0.12)).Transition(motion.Popup(c))
}

// button styles a base button while keeping all of its native behavior.
func button(c *ui.Context, b ui.Element, label string, primary bool) ui.Element {
	t := theme.TokensOf(c)
	b.Height(36).PaddingX(12).Radius(t.Radius).FontSize(14).FontWeight(500).
		Border(1, t.Border).Background(t.Background).TextColor(t.Foreground).Transition(motion.Transition(c))
	if primary {
		b.Background(t.Primary).TextColor(t.PrimaryForeground).Border(0, ui.Transparent)
		if b.Hovered() {
			b.Background(t.Primary.Alpha(0.9))
		}
	} else if b.Hovered() {
		b.Background(t.Accent).TextColor(t.AccentForeground)
	}
	b.Children(func() { ui.Text(c, label).SingleLine() })
	return b
}

type PopoverProps struct {
	Placement Placement
	Width     float32
}

// InWindowPopover toggles on an anchor click and closes on Escape or an outside press.
func InWindowPopover(c *ui.Context, anchor ui.Element, open *bool, p PopoverProps, content func()) ui.Element {
	if anchor.Clicked() {
		*open = !*open
	}
	anchor.Expanded(*open)
	return ui.PopoverBase(c, anchor, open, func(panel ui.Element) {
		surface(c, panel)
		attach(panel, anchor, p.Placement, 4)
		if p.Width > 0 {
			panel.Width(p.Width)
		} else {
			panel.Width(288)
		}
		panel.MaxWidthPercent(100).Children(content)
	})
}

type TooltipProps struct{ Placement Placement }

// InWindowTooltip also describes the trigger to assistive technology, and opens on focus.
func InWindowTooltip(c *ui.Context, anchor ui.Element, text string, p TooltipProps) ui.Element {
	if text == "" {
		return ui.Element{}
	}
	anchor.Description(text)
	t := theme.TokensOf(c)
	return ui.TooltipBase(c, anchor, func(panel ui.Element) {
		attach(panel, anchor, p.Placement, 6)
		panel.Padding(6, 12).Radius(t.Radius).MaxWidth(320).Background(t.Primary).
			TextColor(t.PrimaryForeground).FontSize(12).Transition(motion.Popup(c)).Children(func() { ui.Text(c, text) })
	})
}

type HoverCardProps struct {
	Placement             Placement
	Width                 float32
	OpenDelay, CloseDelay time.Duration
}

type hoverState struct {
	open, panelHover, suppressed bool
	enter, leave                 time.Time
}

// InWindowHoverCard keeps interactive content open while the pointer travels from the
// trigger to its panel; focus opens immediately, Escape dismisses it.
func InWindowHoverCard(c *ui.Context, anchor ui.Element, p HoverCardProps, content func()) ui.Element {
	s := ui.Local(anchor, "shadcn.hover-card", func() hoverState { return hoverState{} })
	now := c.Now()
	over := anchor.Hovered() || anchor.FocusWithin() || s.panelHover
	if !over {
		s.suppressed = false
	}
	if over && !s.suppressed {
		s.leave = time.Time{}
		if s.enter.IsZero() {
			s.enter = now
		}
		delay := p.OpenDelay
		if delay <= 0 {
			delay = 700 * time.Millisecond
		}
		if anchor.FocusWithin() || now.Sub(s.enter) >= delay {
			s.open = true
		} else {
			c.After(delay - now.Sub(s.enter))
		}
	} else {
		s.enter = time.Time{}
		if s.leave.IsZero() {
			s.leave = now
		}
		delay := p.CloseDelay
		if delay <= 0 {
			delay = 300 * time.Millisecond
		}
		if now.Sub(s.leave) >= delay {
			s.open = false
		} else if s.open {
			c.After(delay - now.Sub(s.leave))
		}
	}
	wasOpen := s.open
	s.panelHover = false
	panel := ui.PopoverBase(c, anchor, &s.open, func(panel ui.Element) {
		surface(c, panel)
		attach(panel, anchor, p.Placement, 4)
		width := p.Width
		if width <= 0 {
			width = 320
		}
		panel.Width(width).Children(content)
		s.panelHover = panel.Hovered() || panel.FocusWithin()
	})
	if wasOpen && !s.open {
		s.suppressed = true
	}
	return panel
}
