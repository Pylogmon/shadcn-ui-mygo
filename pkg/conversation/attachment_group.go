package conversation

import "github.com/egoist/mygo/ui"

type AttachmentGroupProps struct {
	State    *ui.ScrollState
	Label    string
	Disabled bool
}

type attachmentGroupStateKey struct{}
type attachmentGroupRuntimeKey struct{}
type attachmentGroupRuntime struct{ width, offset, reach float32 }

var attachmentFadeEase = ui.CubicBezier(.42, 0, .58, 1)

// AttachmentGroup is a horizontally scrolling native strip. Attachment
// roots mark their own snap starts; supply IDs when reordering attachments.
func AttachmentGroup(c *ui.Context, content func()) ui.Element {
	return AttachmentGroupWith(c, AttachmentGroupProps{}, func(*ui.Context) {
		if content != nil {
			content()
		}
	})
}

// AttachmentGroupWith returns the actual native viewport: wheel, touchpad,
// focus reveal, keyboard scrolling, snap settling and TrackScroll stay native.
func AttachmentGroupWith(c *ui.Context, p AttachmentGroupProps, content func(*ui.Context)) ui.Element {
	root := ui.ScrollHorizontal(c).MinWidth(0).Gap(12).PaddingY(4).Scrollbars(false).
		ScrollSnap(ui.SnapMandatory).Disabled(p.Disabled).DimDisabled(false)
	state := p.State
	if state == nil {
		state = ui.Local(root, attachmentGroupStateKey{}, func() ui.ScrollState { return ui.ScrollState{} })
	}
	root.TrackScroll(state)
	geometry := ui.Local(root, attachmentGroupRuntimeKey{}, func() attachmentGroupRuntime { return attachmentGroupRuntime{} })
	if state.MaxX > 0 {
		root.Focusable().FocusRing(false)
	}
	if p.Label != "" {
		root.Label(p.Label)
	}
	rtl := c.Direction() == ui.RightToLeft
	if geometry.width > 0 && state.MaxX > 0 {
		// The pinned scroll-fade-x utility reveals min(12%, 40px) over
		// the first/last96px, eased by the actual native scroll offset.
		fraction := min(float32(.12), 40/geometry.width)
		start := fraction * attachmentFadeEase(max(0, min(1, state.X/96)))
		end := fraction * attachmentFadeEase(max(0, min(1, (state.MaxX-state.X)/96)))
		angle := float32(90)
		if rtl {
			angle = 270
		}
		root.MaskLinearGradient(ui.LinearGradient{Angle: angle, Stops: []ui.GradientStop{
			{Color: ui.Transparent, At: 0},
			{Color: ui.RGB(0, 0, 0), At: start},
			{Color: ui.RGB(0, 0, 0), At: 1 - end},
			{Color: ui.Transparent, At: 1},
		}})
	}
	services := c.Services()
	root.Draw(func(_ *ui.Painter, bounds ui.Rect) {
		if geometry.width != bounds.W || geometry.offset != state.X || geometry.reach != state.MaxX {
			geometry.width, geometry.offset, geometry.reach = bounds.W, state.X, state.MaxX
			services.Invalidate()
		}
	})
	focused := state.Handle.Focused(c)
	root.HandleInput(func(ev ui.InputEvent) bool {
		// MyGo's built-in Home/End targets the vertical axis. Supply the
		// horizontal equivalent only while the viewport itself is focused;
		// descendant editors retain their own caret-navigation keys.
		if ev.Kind == ui.InputKeyDown && ev.Mods == 0 && focused && (ev.Key == ui.KeyHome || ev.Key == ui.KeyEnd) {
			target := float32(0)
			if ev.Key == ui.KeyEnd {
				target = state.MaxX
			}
			state.ScrollTo(target, state.Y, ui.Animation{})
			return true
		}
		if ev.Kind != ui.InputScroll || ev.Mods&^ui.Shift != 0 {
			return false
		}
		delta := ev.DX
		if ev.Mods&ui.Shift != 0 && delta == 0 {
			// The native shift-wheel always advances from logical start.
			delta = ev.DY
		} else if rtl {
			delta = -delta
		}
		// Native scrolling already consumes a packet when it moves. Only
		// the boundary needs containment; ordinary packets retain snapping.
		// A mixed-axis boundary packet is taken whole because HandleInput
		// cannot forward only its vertical component to an ancestor.
		return delta < 0 && state.X <= 0 || delta > 0 && state.X >= state.MaxX
	})
	return root.Children(func() {
		if content != nil {
			content(c)
		}
	})
}
