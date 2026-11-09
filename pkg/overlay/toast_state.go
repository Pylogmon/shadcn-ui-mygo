package overlay

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

type ToastPosition string

const (
	ToastTopLeft      ToastPosition = "top-left"
	ToastTopCenter    ToastPosition = "top-center"
	ToastTopRight     ToastPosition = "top-right"
	ToastBottomLeft   ToastPosition = "bottom-left"
	ToastBottomCenter ToastPosition = "bottom-center"
	ToastBottomRight  ToastPosition = "bottom-right"
)

type ToastInsets struct{ Top, Right, Bottom, Left float32 }
type ToastHotkey struct {
	Modifiers ui.Modifiers
	Key       ui.Key
}
type ToastSwipeDirection string

const (
	ToastSwipeTop    ToastSwipeDirection = "top"
	ToastSwipeRight  ToastSwipeDirection = "right"
	ToastSwipeBottom ToastSwipeDirection = "bottom"
	ToastSwipeLeft   ToastSwipeDirection = "left"
)

type ToastIcons struct{ Success, Info, Warning, Error, Loading, Close *ui.SVG }

type ToasterProps struct {
	State                                   *ToasterState
	Position                                ToastPosition
	Width                                   float32
	Offset, MobileOffset                    *ToastInsets
	Gap                                     *float32
	VisibleToasts                           *int
	Duration                                time.Duration
	Expand, CloseButton, RichColors, Invert bool
	Label, CloseLabel                       string
	SwipeDirections                         []ToastSwipeDirection
	Hotkey                                  *ToastHotkey
	ReturnFocus                             *ui.Handle
	Icons                                   ToastIcons
}

// ToastOptions supplements the compatible native Toast model. A nil boolean
// inherits the viewport default. Content replaces the title/description column;
// with Unstyled it renders the complete notification body without default chrome.
type ToastOptions struct {
	Position                             ToastPosition
	CloseButton, RichColors, Dismissible *bool
	Invert                               bool
	Cancel                               string
	OnCancel                             func()
	Action                               func(*ToastActionEvent)
	OnDismiss, OnAutoClose               func(ui.Toast)
	Icon                                 *ui.SVG
	HideIcon                             bool
	Unstyled                             bool
	Content                              func(*ui.Context)
}

type ToastActionEvent struct {
	Toast     ui.Toast
	prevented bool
}

func (e *ToastActionEvent) PreventDefault() { e.prevented = true }

type toastReason uint8

const (
	toastNone toastReason = iota
	toastDismiss
	toastAutoClose
	toastAction
	toastCancel
)

type toastEntry struct {
	model            ui.Toast
	options          ToastOptions
	key              uint64
	remaining        time.Duration
	duration         time.Duration
	last             time.Time
	reset, announced bool
	closing          time.Time
	pending          toastReason
	height           float32
	drag             ui.Point
	// Gesture state contains numeric event data only; never frame elements.
	swipeStart ui.Point
	swipeAt    time.Time
	swipeAxis  int
	swiping    bool
	swipeExit  bool
	handle     ui.Handle
	version    uint64
}

// ToasterState owns an independent Sonner queue. Keep it in application state,
// and mutate it on the UI thread (or from Window.Update). A zero value is ready
// to use. Unlike MyGo's legacy queue its visible limit is configurable.
type ToasterState struct {
	entries     []*toastEntry
	next        uint64
	services    ui.Services
	paused      bool
	expanded    bool
	list        ui.Handle
	focusInside bool
	mount       *toastMount
}

// A nonzero-sized marker distinguishes viewport lifetimes without retaining
// an Element or Context. The renderer owns it through ui.Local.
type toastMount struct{ identity byte }

// attach preserves caller-owned models across a viewport's unmount, while a
// newly mounted viewport starts a fresh lifetime. Already closed entries have
// delivered their callbacks and must not reappear; pending dismissals remain
// pending and are delivered once by the next live build.
func (s *ToasterState) attach(marker *toastMount) {
	if s.mount == marker {
		return
	}
	s.mount = marker
	s.paused, s.expanded, s.focusInside = false, false, false
	s.list = ui.Handle{}
	live := s.entries[:0]
	for _, e := range s.entries {
		if !e.closing.IsZero() {
			continue
		}
		e.reset, e.announced = true, false
		e.last = time.Time{}
		e.handle = ui.Handle{}
		e.resetGesture()
		live = append(live, e)
	}
	clear(s.entries[len(live):])
	s.entries = live
}

func (e *toastEntry) resetGesture() {
	e.drag, e.swipeStart = ui.Point{}, ui.Point{}
	e.swipeAt = time.Time{}
	e.swipeAxis, e.swiping = 0, false
	e.swipeExit = false
}

func (s *ToasterState) Show(model ui.Toast, options ...ToastOptions) string {
	if model.ID != "" {
		if s.Update(model.ID, model, options...) {
			return model.ID
		}
	}
	s.next++
	if model.ID == "" {
		for {
			model.ID = fmt.Sprintf("sonner-%d", s.next)
			used := false
			for _, e := range s.entries {
				if e.model.ID == model.ID {
					used = true
					break
				}
			}
			if !used {
				break
			}
			s.next++
		}
	}
	e := &toastEntry{model: model, key: s.next, reset: true, height: 54, version: 1}
	if len(options) > 0 {
		e.options = options[0]
	}
	s.entries = append(s.entries, e)
	s.services.Invalidate()
	return model.ID
}

// Update replaces the native model while preserving identity, position in the
// queue and native focus. Updating an entry before its removal revives the
// same identity. Omitted options retain the existing options.
func (s *ToasterState) Update(id string, model ui.Toast, options ...ToastOptions) bool {
	for _, e := range s.entries {
		if e.model.ID == id {
			model.ID = id
			e.model = model
			e.reset = true
			e.announced = false
			e.last = time.Time{}
			e.version++
			e.pending = toastNone
			e.closing = time.Time{}
			e.resetGesture()
			if len(options) > 0 {
				e.options = options[0]
			}
			s.services.Invalidate()
			return true
		}
	}
	return false
}

// Dismiss starts the source's 200ms removal lifecycle. An empty ID dismisses
// all notifications. Dismissible only controls user dismissal, not this API.
func (s *ToasterState) Dismiss(id string) {
	for _, e := range s.entries {
		if (id == "" || e.model.ID == id) && e.closing.IsZero() && e.pending == toastNone {
			e.pending = toastDismiss
		}
	}
	s.services.Invalidate()
}

func (s *ToasterState) Toasts() []ui.Toast {
	result := make([]ui.Toast, 0, len(s.entries))
	for _, e := range s.entries {
		if e.closing.IsZero() && e.pending == toastNone {
			result = append(result, e.model)
		}
	}
	return result
}

func (s *ToasterState) advance(c *ui.Context, props ToasterProps, pause bool) {
	s.services = c.Services()
	now := c.Now()
	// Copy the traversal because application callbacks may Show/Update/Dismiss.
	for _, e := range append([]*toastEntry(nil), s.entries...) {
		if !e.closing.IsZero() {
			continue
		}
		duration := e.model.Timeout
		if duration == 0 {
			duration = props.Duration
		}
		if duration == 0 {
			duration = 4 * time.Second
		}
		if e.reset || e.duration != duration {
			e.duration, e.remaining = duration, duration
			e.last = now
			e.reset = false
		}
		if !e.announced && e.pending == toastNone {
			c.Announce(e.model.Title + " " + e.model.Description)
			e.announced = true
		}
		if e.pending == toastNone && e.model.Type != "loading" && e.remaining > 0 {
			if !s.paused {
				e.remaining -= now.Sub(e.last)
			}
			if e.remaining <= 0 {
				e.pending = toastAutoClose
			} else if !pause {
				c.After(e.remaining)
			}
		}
		e.last = now
		if reason := e.pending; reason != toastNone {
			e.pending = toastNone
			e.closing = now
			model, opts := e.model, e.options
			if model.OnClose != nil {
				model.OnClose()
			}
			if reason == toastAutoClose && opts.OnAutoClose != nil {
				opts.OnAutoClose(model)
			}
			if reason == toastDismiss && opts.OnDismiss != nil {
				opts.OnDismiss(model)
			}
		}
	}
	s.paused = pause
	live := s.entries[:0]
	for _, e := range s.entries {
		if !e.closing.IsZero() {
			left := 200*time.Millisecond - now.Sub(e.closing)
			if left <= 0 {
				continue
			}
			c.After(left)
		}
		live = append(live, e)
	}
	clear(s.entries[len(live):])
	s.entries = live
}

func toastBool(value *bool, fallback bool) bool {
	if value != nil {
		return *value
	}
	return fallback
}
func toastValidPosition(position ToastPosition) ToastPosition {
	switch position {
	case ToastTopLeft, ToastTopCenter, ToastTopRight, ToastBottomLeft, ToastBottomCenter, ToastBottomRight:
		return position
	}
	return ToastBottomRight
}
