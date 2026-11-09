package overlay

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type DialogProps struct {
	Title, Description    string
	Width                 float32
	HideClose             bool
	PreventOutsideDismiss bool
}

type presence struct{ active, entered, closing bool }

// present keeps a modal in the tree until its exit animation finishes. The
// native dialog owns the focus trap throughout, and restores focus on removal.
func present(c *ui.Context, open *bool, preventOutside bool, draw func(back, panel ui.Element, progress float32)) ui.Element {
	marker := ui.Box(c).Absolute().Size(0, 0).AccessHidden(true)
	s := ui.Local(marker, "shadcn.presence", func() presence { return presence{} })
	if *open {
		s.active, s.closing = true, false
	} else if s.active {
		s.closing = true
	}
	if !s.active {
		return ui.Element{}
	}
	shown := true
	finished := false
	result := ui.DialogBase(c, &shown, func(back, panel ui.Element) {
		if preventOutside && back.Clicked() {
			shown = true
		}
		if !s.entered {
			panel.Jump("presence", 0)
			s.entered = true
		}
		target := float32(1)
		if s.closing {
			target = 0
			back.IgnoreInput(true)
		}
		duration := motion.Transition(c).Duration
		if duration > 0 {
			duration = 200 * time.Millisecond
		}
		progress := panel.AnimateTo("presence", target, ui.Animation{Duration: duration, Ease: ui.CubicBezier(0.16, 1, 0.3, 1)})
		if duration == 0 {
			progress = panel.Jump("presence", target)
		}
		finished = s.closing && progress <= 0
		draw(back, panel, progress)
	})
	if !shown {
		*open = false
		s.closing = true
		c.Invalidate()
	}
	if finished {
		*s = presence{}
		c.Invalidate()
	}
	return result
}

// Dialog shows an accessible modal with a focus trap and a closing animation.
// Put action buttons in content and set *open=false when their work succeeds.
func Dialog(c *ui.Context, open *bool, p DialogProps, content func()) ui.Element {
	t := theme.TokensOf(c)
	return present(c, open, p.PreventOutsideDismiss, func(back, panel ui.Element, progress float32) {
		back.Padding(24).Background(ui.RGBA(0, 0, 0, 0.5*progress))
		width := p.Width
		if width <= 0 {
			width = 512
		}
		panel.Column().Gap(16).Padding(24).Width(width).MaxWidthPercent(100).MaxHeightPercent(100).
			Radius(t.Radius+2).Border(1, t.Border).Background(t.Background).TextColor(t.Foreground).
			Label(p.Title).Description(p.Description).Opacity(progress).Scale(0.95+0.05*progress, 0.95+0.05*progress).
			Shadow(0, 12, 40, 0, ui.RGBA(0, 0, 0, 0.18)).Children(func() {
			dialogHeader(c, p.Title, p.Description)
			if content != nil {
				ui.Column(c).Gap(12).Shrink(1).Children(content)
			}
			if !p.HideClose {
				b := ui.ButtonBase(c).Label("Close dialog").Size(24, 24).Absolute().Top(12).Right(12).Radius(4).Center()
				if b.Hovered() {
					b.Background(t.Accent)
				}
				b.Children(func() { ui.Text(c, "×").FontSize(20).TextColor(t.MutedForeground) })
				if b.Clicked() {
					*open = false
				}
			}
		})
	})
}

func dialogHeader(c *ui.Context, title, description string) {
	t := theme.TokensOf(c)
	if title != "" {
		ui.Text(c, title).FontSize(18).FontWeight(600).PaddingEnd(16)
	}
	if description != "" {
		ui.Text(c, description).FontSize(14).TextColor(t.MutedForeground)
	}
}

type AlertDialogProps struct {
	Title, Description       string
	CancelLabel, ActionLabel string
	Destructive              bool
}

// AlertDialog requires an explicit response. A backdrop press is ignored;
// Escape and Cancel dismiss. Action returns true without closing, allowing the
// caller to keep the dialog open if an asynchronous operation fails.
func AlertDialog(c *ui.Context, open *bool, p AlertDialogProps) bool {
	accepted := false
	Dialog(c, open, DialogProps{Title: p.Title, Description: p.Description, HideClose: true, PreventOutsideDismiss: true}, func() {
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			cancel, action := p.CancelLabel, p.ActionLabel
			if cancel == "" {
				cancel = "Cancel"
			}
			if action == "" {
				action = "Continue"
			}
			if button(c, ui.ButtonBase(c).AutoFocus(), cancel, false).Clicked() {
				*open = false
			}
			b := button(c, ui.ButtonBase(c), action, true)
			if p.Destructive {
				t := theme.TokensOf(c)
				b.Background(t.Destructive).TextColor(t.DestructiveForeground)
			}
			accepted = b.Clicked()
		})
	}).Role(ui.RoleAlertDialog)
	return accepted
}

type SheetProps struct {
	Title, Description string
	Side               Placement
	Size               float32
	HideClose          bool
}

// Sheet enters from one edge of the window, keeping modal focus semantics.
func Sheet(c *ui.Context, open *bool, p SheetProps, content func()) ui.Element {
	t := theme.TokensOf(c)
	return present(c, open, false, func(back, panel ui.Element, progress float32) {
		back.Background(ui.RGBA(0, 0, 0, 0.5*progress))
		size := p.Size
		if size <= 0 {
			size = 384
		}
		panel.Column().Gap(16).Padding(24).Background(t.Background).TextColor(t.Foreground).
			Label(p.Title).Description(p.Description).Absolute().Opacity(progress)
		offset := size * (1 - progress)
		switch p.Side {
		case Top:
			panel.Top(-offset).Left(0).Right(0).Height(size).MaxHeightPercent(90)
		case Bottom:
			panel.Bottom(-offset).Left(0).Right(0).Height(size).MaxHeightPercent(90)
		case Left:
			panel.Left(-offset).Top(0).Bottom(0).Width(size).MaxWidthPercent(90)
		default:
			panel.Right(-offset).Top(0).Bottom(0).Width(size).MaxWidthPercent(90)
		}
		panel.Border(1, t.Border).Children(func() {
			dialogHeader(c, p.Title, p.Description)
			if content != nil {
				ui.Scroll(c).Grow(1).Gap(12).Children(content)
			}
			if !p.HideClose {
				b := ui.ButtonBase(c).Label("Close sheet").Absolute().Top(12).Right(12).Size(24, 24).Center().Radius(4)
				if b.Hovered() {
					b.Background(t.Accent)
				}
				b.Children(func() { ui.Text(c, "×").FontSize(20) })
				if b.Clicked() {
					*open = false
				}
			}
		})
	})
}

type DrawerProps struct {
	Title, Description string
	Height             float32
}

// Drawer is a bottom sheet with an accessible close control and drag handle.
func Drawer(c *ui.Context, open *bool, p DrawerProps, content func()) ui.Element {
	panel := Sheet(c, open, SheetProps{Title: p.Title, Description: p.Description, Side: Bottom, Size: p.Height}, func() {
		t := theme.TokensOf(c)
		ui.Box(c).Width(80).Height(4).Radius(2).AlignSelf(ui.Center).Background(t.Muted)
		if content != nil {
			content()
		}
	})
	height := p.Height
	if height <= 0 {
		height = 384
	}
	return panel.Draggable(nil, ui.DraggableOptions{Axis: ui.Vertical, Bounds: &ui.Rect{H: height}, Snap: func(point ui.Point) ui.Point {
		if point.Y > height/3 {
			*open = false
		}
		return ui.Point{}
	}})
}
