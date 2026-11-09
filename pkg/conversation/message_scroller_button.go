package conversation

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type messageScrollerButtonAnimationKey struct{}

// MessageScrollerButton is a real native button. Its inactive state is absent
// from keyboard navigation and accessibility; the visual exit can still finish.
func MessageScrollerButton(c *ui.Context, p MessageScrollerButtonProps, content ...func(*ui.Context)) ui.Element {
	s := scrollerScope(c)
	direction := p.Direction
	if direction != MessageScrollerStart {
		direction = MessageScrollerEnd
	}
	active := s.scrollable.End
	if direction == MessageScrollerStart {
		active = s.scrollable.Start
	}
	label := p.Label
	if label == "" {
		label = "Scroll to " + string(direction)
	}
	button := control.Button(c, "", control.ButtonProps{Variant: control.Secondary, Size: control.IconSmall, Label: label, Disabled: !active}).
		Absolute().LeftPercent(50).TranslatePercent(-50, 0).DimDisabled(false).AccessHidden(!active).IgnoreInput(!active)
	if direction == MessageScrollerStart {
		button.Top(16)
	} else {
		button.Bottom(16)
	}
	t := theme.TokensOf(c)
	background := t.Background
	if button.Hovered() || button.Pressed() {
		background = t.Muted
	}
	button.Background(background).TextColor(t.Foreground).Border(1, t.Border)
	if button.FocusVisible() {
		button.Border(1, t.Ring)
	}
	phase := float32(0)
	if active {
		phase = 1
	}
	if !motion.Reduced(c) {
		duration := 400 * time.Millisecond
		ease := ui.CubicBezier(.7, 0, .84, 0)
		if active {
			duration = 200 * time.Millisecond
			ease = ui.CubicBezier(.23, 1, .32, 1)
		}
		phase = button.AnimateTo(messageScrollerButtonAnimationKey{}, phase, ui.Animation{Duration: duration, Ease: ease})
	}
	y := 32 * (1 - phase)
	if direction == MessageScrollerStart {
		y = -y
	}
	button.Translate(0, y).Scale(.95+.05*phase, .95+.05*phase).Opacity(phase)
	if len(content) > 0 && content[0] != nil {
		button.Children(func() { content[0](c) })
	} else {
		button.Children(func() {
			icon := ui.Icon(c, messageScrollerArrowDownIcon()).Size(16, 16).IgnoreInput(true).AccessHidden(true)
			if direction == MessageScrollerStart {
				icon.Rotate(180)
			}
		})
	}
	if active && button.Clicked() && (p.OnClick == nil || p.OnClick()) {
		if direction == MessageScrollerStart {
			s.ScrollToStart(MessageScrollerScrollOptions{Smooth: !p.Instant})
		} else {
			s.ScrollToEnd(MessageScrollerScrollOptions{Smooth: !p.Instant})
		}
		// The web button blurs itself. A native desktop has no document body
		// focus target, so return navigation to its real scrolling region.
		s.Scroll.Handle.Focus(c)
	}
	return button
}
