package overlay

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type DialogProps struct {
	Title, Description                          string
	Width, Height, MaxHeight                    float32
	HideClose, HideHeader                       bool
	PreventOutsideDismiss, PreventEscapeDismiss bool
	// InitialFocus and ReturnFocus identify real controls, never frame Elements.
	InitialFocus, ReturnFocus *ui.Handle
	// ContentPadding/Gap accept explicit zero; nil uses 24 DIPs.
	ContentPadding, ContentGap *float32
	// NoScroll lets content with its own bounded list manage scrolling.
	NoScroll bool
	Footer   func()
	// OnOpenChange reports close requests made by this component. Changing the
	// caller-owned open value before building it does not report a user event.
	OnOpenChange func(bool)
}

type presence struct {
	active, entered, closing, returnPending bool
}
type modalOptions struct {
	preventOutside, preventEscape, nonModal bool
	duration                                time.Duration
	initialFocus, returnFocus               *ui.Handle
	onOpenChange                            func(bool)
}

// present retains the native modal scope through the exit. Closing content
// keeps its pixels but accepts neither input nor accessibility actions.
func present(c *ui.Context, open *bool, p modalOptions, draw func(back, panel ui.Element, progress float32)) ui.Element {
	if open == nil {
		return ui.Element{}
	}
	marker := ui.Box(c).Absolute().Size(0, 0).AccessHidden(true).IgnoreInput(true)
	s := ui.Local(marker, "shadcn.presence", func() presence { return presence{} })
	before := *open
	if *open {
		s.active, s.closing, s.returnPending = true, false, false
	} else if s.active {
		s.closing = true
	}
	if !s.active {
		if s.returnPending {
			s.returnPending = false
			if p.returnFocus != nil {
				p.returnFocus.Focus(c)
			}
		}
		return ui.Element{}
	}
	shown := true
	finished := false
	opening := !s.entered && *open
	paint := func(back, panel ui.Element) {
		if !shown && (p.preventOutside && back.Clicked() || p.preventEscape && !back.Clicked()) {
			shown = true
		}
		if !s.entered {
			panel.Jump("presence", 0)
			s.entered = true
		}
		target := float32(1)
		if s.closing {
			target = 0
		}
		duration := p.duration
		if duration <= 0 {
			duration = 100 * time.Millisecond
		}
		progress := target
		if motion.Reduced(c) {
			progress = panel.Jump("presence", target)
		} else {
			progress = panel.AnimateWith("presence", target, duration, motion.EaseOut)
		}
		if s.closing {
			panel.Disabled(true).DimDisabled(false).IgnoreInput(true).AccessHidden(true)
		}
		finished = s.closing && progress <= 0
		draw(back, panel, progress)
	}
	var result ui.Element
	var backdrop ui.Element
	if p.nonModal {
		ui.Overlay(c, func() {
			backdrop = ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).PassThrough()
			backdrop.Children(func() {
				result = ui.Box(c).Role(ui.RoleDialog)
				result.Clicked() // panel background absorbs presses above the window
			})
		})
	} else {
		result = ui.DialogBase(c, &shown, func(back, _ ui.Element) { backdrop = back })
	}
	// Native Overlay records the opener when its builder returns. Append
	// content afterwards so content AutoFocus cannot replace that opener.
	// Children keeps the real panel's Context, ancestry and native trap.
	result.Children(func() { paint(backdrop, result) })
	if p.nonModal && ((!p.preventOutside && result.PressedOutside()) || (!p.preventEscape && result.OverlayShortcut(0, ui.KeyEscape))) {
		shown = false
	}
	if opening && p.initialFocus != nil && *open {
		p.initialFocus.Focus(c)
	}
	if !shown {
		*open = false
		s.closing = true
		c.Invalidate()
	}
	if before != *open && p.onOpenChange != nil {
		p.onOpenChange(*open)
	}
	if finished {
		*s = presence{returnPending: p.returnFocus != nil}
		c.Invalidate()
	}
	return result
}

func modalDimension(v, fallback float32) (float32, bool) {
	if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
		return fallback, false
	}
	if v <= 0 {
		return fallback, true
	}
	return v, true
}
func modalSpace(v *float32, fallback float32) float32 {
	if v == nil || math.IsNaN(float64(*v)) || math.IsInf(float64(*v), 0) {
		return fallback
	}
	return max(0, *v)
}

// Dialog is a constrained native modal with a real scroll viewport, native
// nested focus trapping, cancellable exit motion and optional explicit focus.
func Dialog(c *ui.Context, open *bool, p DialogProps, content func()) ui.Element {
	t := theme.TokensOf(c)
	_, windowHeight := c.Size()
	limit := max(float32(1), windowHeight-32)
	maxHeight, validMax := modalDimension(p.MaxHeight, limit)
	maxHeight = min(limit, maxHeight)
	width, validWidth := modalDimension(p.Width, 448)
	height, validHeight := modalDimension(p.Height, 0)
	return present(c, open, modalOptions{preventOutside: p.PreventOutsideDismiss, preventEscape: p.PreventEscapeDismiss, duration: 100 * time.Millisecond, initialFocus: p.InitialFocus, returnFocus: p.ReturnFocus, onOpenChange: p.OnOpenChange}, func(back, panel ui.Element, progress float32) {
		back.Padding(16).Background(ui.RGBA(0, 0, 0, .1*progress))
		panel.Column().Gap(modalSpace(p.ContentGap, 24)).Padding(modalSpace(p.ContentPadding, 24)).Width(width).MaxWidthPercent(100).MaxHeight(maxHeight).MinHeight(0).
			Radius(t.RadiusXL()).Border(1, t.Foreground.Alpha(.1)).Background(t.Popover).TextColor(t.PopoverForeground).FontSize(14).
			Label(p.Title).Description(p.Description).Opacity(progress).Scale(.95+.05*progress, .95+.05*progress).
			Invalid(!validMax || !validWidth || !validHeight).Shadow(0, 12, 40, 0, ui.RGBA(0, 0, 0, .18))
		if height > 0 {
			panel.Height(min(height, maxHeight))
		}
		panel.Children(func() {
			if !p.HideHeader {
				dialogHeader(c, p.Title, p.Description)
			}
			if content != nil {
				if p.NoScroll {
					ui.Column(c).FillWidth().MinHeight(0).Gap(12).Children(content)
				} else {
					body := ui.Scroll(c).FillWidth().MinHeight(0).Shrink(1).MaxHeight(maxHeight).Label("Dialog content")
					if height > 0 {
						body.Grow(1)
					}
					body.Children(func() { ui.Column(c).FillWidth().Shrink(0).Gap(12).Children(content) })
				}
			}
			if p.Footer != nil {
				modalFooter(c, p.Footer)
			}
			if !p.HideClose {
				modalClose(c, open, "Close dialog", 16)
			}
		})
	})
}

func dialogHeader(c *ui.Context, title, description string) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).FillWidth().Shrink(0).Gap(8).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(18).FontWeight(500).LineHeight(1).PaddingEnd(24).Role(ui.RoleHeading).Level(2)
		}
		if description != "" {
			ui.Text(c, description).FontSize(14).TextColor(t.MutedForeground)
		}
	})
}
func modalFooter(c *ui.Context, content func()) ui.Element {
	width, _ := c.Size()
	footer := ui.Row(c).FillWidth().Shrink(0).Gap(8).Justify(ui.End)
	if width < 640 {
		footer.Column().Reverse()
	}
	return footer.Children(content)
}
func modalClose(c *ui.Context, open *bool, label string, inset float32) ui.Element {
	t := theme.TokensOf(c)
	b := ui.ButtonBase(c).Label(label).Size(24, 24).Absolute().Top(inset).Right(inset).Radius(t.RadiusSM()).Center()
	if b.Hovered() {
		b.Background(t.Accent)
	}
	b.Children(func() { ui.Text(c, "×").FontSize(20).TextColor(t.MutedForeground).AccessHidden(true) })
	if b.Clicked() {
		*open = false
	}
	return b
}

type AlertDialogProps struct {
	Title, Description                      string
	CancelLabel, ActionLabel                string
	Destructive, Small                      bool
	Pending, ActionDisabled, CancelDisabled bool
	AllowCancelWhilePending                 bool
	Media, Content                          func()
	OnCancel                                func()
	InitialFocus, ReturnFocus               *ui.Handle
}

// AlertDialog returns true once per enabled action, without automatically
// closing. Set Pending during asynchronous work to prevent duplicate actions.
func AlertDialog(c *ui.Context, open *bool, p AlertDialogProps) bool {
	accepted := false
	windowWidth, _ := c.Size()
	width := float32(512)
	if p.Small || windowWidth < 640 {
		width = 320
	}
	locked := p.Pending && !p.AllowCancelWhilePending
	footer := func() {
		var row ui.Element
		if p.Small {
			row = ui.Grid(c).Columns(2).FillWidth().Gap(8)
		} else {
			row = ui.Row(c).FillWidth().Gap(8).Justify(ui.End)
			if windowWidth < 640 {
				row.Column().Reverse()
			}
		}
		row.Children(func() {
			cancel, action := p.CancelLabel, p.ActionLabel
			if cancel == "" {
				cancel = "Cancel"
			}
			if action == "" {
				action = "Continue"
			}
			cb := button(c, ui.ButtonBase(c).Disabled(p.CancelDisabled || locked), cancel, false)
			if p.Small {
				cb.FillWidth().MinWidth(0).Center()
			}
			if p.InitialFocus == nil && !cb.IsDisabled() {
				cb.AutoFocus()
			}
			if cb.Clicked() {
				*open = false
				if p.OnCancel != nil {
					p.OnCancel()
				}
			}
			b := button(c, ui.ButtonBase(c).Disabled(p.Pending || p.ActionDisabled).Busy(p.Pending), action, true)
			if p.Small {
				b.FillWidth().MinWidth(0).Center()
			}
			if p.Destructive {
				t := theme.TokensOf(c)
				b.Background(t.Destructive).TextColor(t.DestructiveForeground)
			}
			accepted = b.Clicked()
		})
	}
	dialog := Dialog(c, open, DialogProps{Title: p.Title, Description: p.Description, HideHeader: true, Width: width, HideClose: true, PreventOutsideDismiss: true, PreventEscapeDismiss: locked, InitialFocus: p.InitialFocus, ReturnFocus: p.ReturnFocus, Footer: footer}, func() {
		t := theme.TokensOf(c)
		center := p.Small || windowWidth < 640
		header := ui.Column(c).FillWidth().Shrink(0).Gap(6).AlignItems(ui.Center).TextAlign(ui.Center)
		if !center {
			header.AlignItems(ui.Start).TextAlign(ui.Start)
			if p.Media != nil {
				header.Row().Gap(24)
			}
		}
		header.Children(func() {
			if p.Media != nil {
				ui.Box(c).Size(64, 64).Shrink(0).Margin(0, 0, 8, 0).Radius(t.RadiusMD()).Background(t.Muted).Center().Children(p.Media)
			}
			text := ui.Column(c).MinWidth(0).Gap(6)
			if !center {
				text.Grow(1)
			} else {
				text.FillWidth()
			}
			text.Children(func() {
				if p.Title != "" {
					ui.Text(c, p.Title).FontSize(18).FontWeight(500).Role(ui.RoleHeading).Level(2)
				}
				if p.Description != "" {
					ui.Text(c, p.Description).FontSize(14).TextColor(t.MutedForeground)
				}
			})
		})
		if p.Content != nil {
			p.Content()
		}
	})
	dialog.Role(ui.RoleAlertDialog).Busy(p.Pending)
	return accepted
}

type SheetProps struct {
	Title, Description string
	Side               Placement
	// LogicalSide maps Left/Right to the logical start/end of the parent scope.
	LogicalSide                                 bool
	Size                                        float32
	HideClose, HideHeader                       bool
	PreventOutsideDismiss, PreventEscapeDismiss bool
	InitialFocus, ReturnFocus                   *ui.Handle
	ContentPadding, ContentGap                  *float32
	Footer                                      func()
	OnOpenChange                                func(bool)
}

func logicalSide(side Placement, logical bool, direction ui.Direction) Placement {
	if logical && direction == ui.RightToLeft {
		if side == Left {
			return Right
		}
		if side == Right {
			return Left
		}
	}
	return side
}

// Sheet is an edge modal. Left/Right remain physical unless LogicalSide is set.
func Sheet(c *ui.Context, open *bool, p SheetProps, content func()) ui.Element {
	t := theme.TokensOf(c)
	side := logicalSide(p.Side, p.LogicalSide, c.Direction())
	width, height := c.Size()
	available := width
	if side == Top || side == Bottom {
		available = height
	}
	size, valid := modalDimension(p.Size, 384)
	size = min(size, max(float32(1), available*.9))
	padding := modalSpace(p.ContentPadding, 16)
	return present(c, open, modalOptions{preventOutside: p.PreventOutsideDismiss, preventEscape: p.PreventEscapeDismiss, duration: 200 * time.Millisecond, initialFocus: p.InitialFocus, returnFocus: p.ReturnFocus, onOpenChange: p.OnOpenChange}, func(back, panel ui.Element, progress float32) {
		back.Background(ui.RGBA(0, 0, 0, .1*progress))
		panel.Column().Gap(modalSpace(p.ContentGap, 16)).Padding(padding).Background(t.Popover).TextColor(t.PopoverForeground).FontSize(14).
			Label(p.Title).Description(p.Description).Absolute().Opacity(progress).Invalid(!valid).Border(1, t.Border)
		offset := float32(40) * (1 - progress)
		switch side {
		case Top:
			panel.Top(-offset).Left(0).Right(0).Height(size)
		case Bottom:
			panel.Bottom(-offset).Left(0).Right(0).Height(size)
		case Left:
			panel.Left(-offset).Top(0).Bottom(0).Width(size)
		default:
			panel.Right(-offset).Top(0).Bottom(0).Width(size)
		}
		panel.Children(func() {
			if !p.HideHeader {
				dialogHeader(c, p.Title, p.Description)
			}
			if content != nil {
				ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Gap(12).Children(func() { ui.Column(c).Shrink(0).FillWidth().Gap(12).Children(content) })
			}
			if p.Footer != nil {
				modalFooter(c, p.Footer)
			}
			if !p.HideClose {
				modalClose(c, open, "Close sheet", 16)
			}
		})
	})
}
