package desktop

import (
	"errors"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// TooltipProps configures a non-interactive, pointer-transparent native Flyout.
// The default open and close delays are zero, as shadcn's TooltipProvider.
type TooltipProps struct {
	Placement mygo.Placement
	// Side optionally overrides Placement's physical side. Logical sides use
	// the parent reading direction; Placement still supplies start/end alignment.
	Side TooltipSide
	// Description is required by TooltipWith and NativeTooltipWith. Rich
	// visuals do not replace the trigger's assistive-technology description.
	Description           string
	Gap, MaxWidth         int
	OpenDelay, CloseDelay time.Duration
	Disabled              bool
	Revision              uint64
	OnOpenChange          func(bool)
}

// HoverCardProps configures an interactive preview in an independent window.
// Zero delays use 700 ms open / 300 ms close; negative delays mean immediate.
// A pointer-opened card does not take keyboard focus. Down or Tab while its
// trigger is focused enters the card, and Escape closes it to the trigger.
type HoverCardProps struct {
	Label, Description                       string
	Placement                                mygo.Placement
	Width, Height, Gap                       int
	MinWidth, MinHeight, MaxWidth, MaxHeight int
	OpenDelay, CloseDelay                    time.Duration
	Disabled, NonInteractive                 bool
	Revision                                 uint64
	OnOpenChange                             func(bool)
}

type hoverKey struct{ kind, key string }

type hoverApplicationState struct{ known, active bool }

// WindowActive describes a rendered window and can lag a native focus transfer.
// Only explicit application activity can bridge that gap; unknown backends keep
// the conservative window-based dismissal behavior.
func hoverApplication(h *Host) hoverApplicationState {
	active, known := h.ApplicationState()
	return hoverApplicationState{known: known, active: active}
}

// Flyouts inherit their owner's drawn-active appearance even when ShowInactive
// leaves the keyboard in the parent. Only real window focus may keep a pointer
// preview alive after the pointer leaves both windows. Injected surfaces can
// report IsFocused. Without that capability, an exact body focus proves initial
// keyboard ownership; FocusLeft then tracks a move into its descendants or out
// of the window without mistaking a merely retained UI focus for ownership.
func hoverWindowFocused(h *Host, key string, body ui.Element, wasFocused, focusLeft bool) bool {
	if h != nil {
		if focused, known := h.FocusState(key); known {
			return focused
		}
	}
	return body.Focused() || wasFocused && !focusLeft && body.FocusWithin()
}

type hoverFlyoutState struct {
	open, suppressed                           bool
	childHover, childFocus, childWindowFocus   bool
	enterChild, keyboardChild, keyboardPending bool
	enteredAt, leftAt                          time.Time
	text                                       string
	lastRevision, renderRevision               uint64
	onOpenChange                               func(bool)
	anchor                                     ui.Handle
}

func (s *hoverFlyoutState) setOpen(open bool) {
	if s.open == open {
		return
	}
	s.open = open
	if s.onOpenChange != nil {
		s.onOpenChange(open)
	}
}
func (s *hoverFlyoutState) dismiss() {
	if s.keyboardChild && (s.keyboardPending || s.childWindowFocus) {
		s.anchor.Focus()
	}
	s.suppressed = true
	s.enteredAt, s.leftAt = time.Time{}, time.Time{}
	s.childHover, s.childFocus, s.childWindowFocus = false, false, false
	s.enterChild, s.keyboardChild, s.keyboardPending = false, false, false
	s.setOpen(false)
}
func (s *hoverFlyoutState) cancel() {
	if s.keyboardChild && (s.keyboardPending || s.childWindowFocus) {
		s.anchor.Focus()
	}
	s.enteredAt, s.leftAt = time.Time{}, time.Time{}
	s.childHover, s.childFocus, s.childWindowFocus = false, false, false
	s.enterChild, s.keyboardChild, s.keyboardPending = false, false, false
	s.setOpen(false)
}

func hoverArgs(c *ui.Context, key string, anchor ui.Element) error {
	if c == nil || !c.Valid() || !anchor.Valid() {
		return errors.New("desktop: hover content needs a live parent context and anchor")
	}
	if key == "" {
		return errors.New("desktop: hover content needs a stable nonempty key")
	}
	return nil
}

// advanceHover runs the timer only in the parent Context. Child windows report
// transitions, not timestamps or parent-frame Elements, so clocked tests and
// native frames both cancel/re-enter the same state machine without idle loops.
func advanceHover(c *ui.Context, s *hoverFlyoutState, anchorHot, immediate bool, openDelay, closeDelay time.Duration, interactive bool, application hoverApplicationState) {
	keyboardActive := interactive && s.keyboardPending && application.known && application.active
	childHot := interactive && (s.childHover || s.childFocus || s.childWindowFocus || keyboardActive)
	hot := anchorHot || childHot
	if c.WindowHidden() || application.known && !application.active || !c.WindowActive() && (!interactive || !s.childWindowFocus) && !keyboardActive {
		if s.open || !s.enteredAt.IsZero() {
			s.dismiss()
		}
		return
	}
	if s.suppressed {
		if !hot {
			s.suppressed = false
		} else {
			s.cancel()
			return
		}
	}
	now := c.Now()
	if hot {
		s.leftAt = time.Time{}
		if s.enteredAt.IsZero() {
			s.enteredAt = now
		}
		if s.open || immediate || openDelay <= 0 || now.Sub(s.enteredAt) >= openDelay {
			s.setOpen(true)
			return
		}
		c.After(openDelay - now.Sub(s.enteredAt))
		return
	}
	s.enteredAt = time.Time{}
	if !s.open {
		s.leftAt = time.Time{}
		return
	}
	if s.leftAt.IsZero() {
		s.leftAt = now
	}
	if closeDelay <= 0 || now.Sub(s.leftAt) >= closeDelay {
		s.cancel()
	} else {
		c.After(closeDelay - now.Sub(s.leftAt))
	}
}

// Tooltip uses an independent native Flyout with shadcn's inverse surface.
// It never takes focus or mouse input and can extend beyond its parent window.
func Tooltip(c *ui.Context, h *Host, key string, anchor ui.Element, text string, p TooltipProps) error {
	return tooltip(c, h, key, anchor, text, p, false, nil)
}

// TooltipWith builds inert rich content in the independent child Context.
// Give Description and advance Revision when application content changes.
// The callback must never capture a parent Context or Element.
func TooltipWith(c *ui.Context, h *Host, key string, anchor ui.Element, p TooltipProps, content func(*ui.Context)) error {
	if p.Description == "" {
		return errors.New("desktop: rich Tooltip needs an accessible Description")
	}
	return tooltip(c, h, key, anchor, p.Description, p, false, content)
}

// NativeTooltipWith retains the independent window's platform theme for rich
// inert visuals. Like NativeTooltip, it is not an OS tooltip-widget API.
func NativeTooltipWith(c *ui.Context, h *Host, key string, anchor ui.Element, p TooltipProps, content func(*ui.Context)) error {
	if p.Description == "" {
		return errors.New("desktop: rich Tooltip needs an accessible Description")
	}
	return tooltip(c, h, key, anchor, p.Description, p, true, content)
}

// NativeTooltip uses a pointer-transparent native Flyout and its child's
// platform theme. MyGo does not expose an operating-system tooltip-widget API;
// this is deliberately a platform-themed native window, not such a widget.
func NativeTooltip(c *ui.Context, h *Host, key string, anchor ui.Element, text string, p TooltipProps) error {
	return tooltip(c, h, key, anchor, text, p, true, nil)
}

func tooltip(c *ui.Context, h *Host, key string, anchor ui.Element, text string, p TooltipProps, native bool, content func(*ui.Context)) error {
	if err := hoverArgs(c, key, anchor); err != nil {
		return err
	}
	anchor.Description(text)
	placement, err := tooltipPlacement(p, c.Direction())
	if err != nil {
		return err
	}
	s := ui.Local(anchor, hoverKey{"tooltip", key}, func() hoverFlyoutState { return hoverFlyoutState{} })
	s.onOpenChange = p.OnOpenChange
	if s.text != text || s.lastRevision != p.Revision {
		s.text, s.lastRevision = text, p.Revision
		s.renderRevision++
	}
	anchorHot := anchor.Hovered() || anchor.FocusWithin()
	if p.Disabled || text == "" {
		s.cancel()
	} else {
		if (s.open || !s.enteredAt.IsZero()) && c.Shortcut(0, ui.KeyEscape) || anchor.Clicked() {
			s.dismiss()
		}
		advanceHover(c, s, anchorHot, anchor.FocusWithin(), max(0, p.OpenDelay), max(0, p.CloseDelay), false, hoverApplicationState{})
	}
	if h == nil {
		if s.open {
			s.dismiss()
			return errors.New("desktop: opening a Tooltip requires a parent-bound Host")
		}
		return nil
	}
	maxWidth := p.MaxWidth
	if maxWidth <= 0 {
		maxWidth = 320
	}
	maxWidth = max(32, maxWidth)
	windowWidth := maxWidth
	if !native {
		windowWidth += 2 * tooltipArrowInset
	}
	zero := float32(0)
	err = h.Popover(c, key, anchor, &s.open, PopoverProps{
		Label: text, Placement: placement, Width: windowWidth, Height: 32, Gap: p.Gap,
		FitContent: true, MaxWidth: windowWidth, NonFocusable: true, Plain: true,
		SystemStyle: native, IgnoreMouseEvents: true, ShowInactive: true,
		tooltipSurface: true,
		ContentPadding: &zero, ContentRole: ui.RoleTooltip, ManualTrigger: true,
		Revision: s.renderRevision, OnDismiss: s.dismiss, NoFocusRestore: true,
	}, func(child *ui.Context) {
		var background, foreground ui.Color
		if native {
			nativeTheme := child.Theme()
			background, foreground = nativeTheme.Inverse, nativeTheme.InverseText
			if background == ui.Transparent {
				background = nativeTheme.Text
			}
			if foreground == ui.Transparent {
				foreground = nativeTheme.Background
			}
		} else {
			t := theme.TokensOf(child)
			background, foreground = t.Foreground, t.Background
		}
		tooltipContent(child, h, key, maxWidth, background, foreground, !native, func(child *ui.Context) {
			if content != nil {
				content(child)
			} else {
				ui.Text(child, text).MaxWidth(float32(maxWidth - 24))
			}
		})
	})
	if err != nil {
		s.dismiss()
	}
	return err
}

// HoverCard shows interactive shadcn content in a Flyout without stealing
// keyboard focus on hover. content must build with its child Context.
func HoverCard(c *ui.Context, h *Host, key string, anchor ui.Element, p HoverCardProps, content func(*ui.Context)) error {
	return hoverCard(c, h, key, anchor, p, content, false)
}

// NativeHoverCard retains the independent child's platform theme. It uses a
// plain native Flyout shown inactive: macOS NSPopover always takes focus, so
// it is intentionally not used for a pointer-opened preview card.
func NativeHoverCard(c *ui.Context, h *Host, key string, anchor ui.Element, p HoverCardProps, content func(*ui.Context)) error {
	return hoverCard(c, h, key, anchor, p, content, true)
}

func hoverCard(c *ui.Context, h *Host, key string, anchor ui.Element, p HoverCardProps, content func(*ui.Context), native bool) error {
	if err := hoverArgs(c, key, anchor); err != nil {
		return err
	}
	s := ui.Local(anchor, hoverKey{"hover-card", key}, func() hoverFlyoutState { return hoverFlyoutState{} })
	anchor.Bind(&s.anchor)
	s.onOpenChange = p.OnOpenChange
	if p.Description != "" {
		anchor.Description(p.Description)
	}
	anchorHot := anchor.Hovered() || anchor.FocusWithin()
	interactive := !p.NonInteractive
	openDelay, closeDelay := p.OpenDelay, p.CloseDelay
	if openDelay == 0 {
		openDelay = 700 * time.Millisecond
	}
	if closeDelay == 0 {
		closeDelay = 300 * time.Millisecond
	}
	if p.Disabled {
		s.cancel()
	} else {
		if (s.open || !s.enteredAt.IsZero()) && c.Shortcut(0, ui.KeyEscape) {
			s.dismiss()
		}
		if interactive && (anchor.Shortcut(0, ui.KeyDown) || s.open && anchor.Shortcut(0, ui.KeyTab)) {
			s.suppressed = false
			s.enterChild, s.keyboardChild, s.keyboardPending = true, true, true
		}
		advanceHover(c, s, anchorHot, anchor.FocusWithin() || s.enterChild, openDelay, closeDelay, interactive, hoverApplication(h))
	}
	if h == nil {
		if s.open {
			s.dismiss()
			return errors.New("desktop: opening a HoverCard requires a parent-bound Host")
		}
		return nil
	}
	width := p.Width
	if width == 0 {
		width = 320
	}
	minWidth, maxWidth := p.MinWidth, p.MaxWidth
	if minWidth == 0 && maxWidth == 0 {
		minWidth, maxWidth = width, width
	}
	label := p.Label
	if label == "" {
		label = "Preview"
	}
	props := PopoverProps{
		Label: label, Placement: p.Placement, Width: width, Height: p.Height, Gap: p.Gap,
		FitContent: true, MinWidth: minWidth, MinHeight: p.MinHeight, MaxWidth: maxWidth, MaxHeight: p.MaxHeight,
		NonFocusable: !interactive, Plain: true, SystemStyle: native, ShowInactive: true, IgnoreMouseEvents: !interactive,
		ContentRole: ui.RoleNone, ManualTrigger: true, Revision: p.Revision, NoFocusRestore: true,
		OnDismiss: func() {
			// Native destruction can precede the child's blur frame. The closed
			// surface no longer owns the keyboard; preserve whichever parent
			// control received it. Explicit child Escape restores before closing.
			s.childWindowFocus, s.keyboardPending = false, false
			s.dismiss()
		},
	}
	err := h.Popover(c, key, anchor, &s.open, props, func(child *ui.Context) {
		body := ui.Column(child).FillWidth().Gap(8).Role(ui.RoleGroup).Label(label).Disabled(!interactive)
		if interactive {
			// The explicit keyboard-entry group is a real focus target, not a
			// hidden parent-frame sentinel. It survives the Host's entry animation.
			body.Focusable()
			if s.keyboardChild && !body.FocusWithin() {
				body.Focus()
			}
			if body.Focused() && body.Shortcut(ui.Shift, ui.KeyTab) || body.Shortcut(0, ui.KeyEscape) {
				s.dismiss()
				h.backend.InvalidateParent()
			}
		}
		body.Children(func() {
			if content != nil {
				content(child)
			}
		})
		// Observe the native keyboard boundary as well as its drawn appearance;
		// WindowActive may stay true for both owner and inactive child. This
		// dependency schedules the child frame that reports a real focus change.
		focusLeft := body.FocusLeft()
		hover, windowFocus := body.Hovered(), hoverWindowFocused(h, key, body, s.childWindowFocus, focusLeft)
		focus := windowFocus && body.FocusWithin()
		if !interactive {
			focus, windowFocus = false, false
		}
		// Application activity bridges only the unconfirmed native transfer.
		// Once the child actually owns the keyboard, ordinary hover/focus rules
		// resume so moving to another parent control can close the preview.
		confirmed := s.keyboardPending && windowFocus
		if confirmed {
			s.keyboardPending = false
		}
		if s.childHover != hover || s.childFocus != focus || s.childWindowFocus != windowFocus || confirmed {
			s.childHover, s.childFocus, s.childWindowFocus = hover, focus, windowFocus
			h.backend.InvalidateParent()
		}
	})
	// Preview cards are descriptive content, not menu/dialog trigger controls.
	anchor.HasPopup(ui.PopupNone).Expanded(false)
	if err != nil {
		s.dismiss()
		return err
	}
	if s.enterChild && h.Focus(key) {
		s.enterChild = false
		h.Refresh(key)
	}
	return nil
}
