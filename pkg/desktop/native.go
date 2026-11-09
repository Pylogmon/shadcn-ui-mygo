package desktop

import (
	"errors"
	"fmt"
	"math"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// NewHost binds native flyouts to parent. Create it after App.WhenReady and
// after the parent Window. The parent owns and closes its native child windows.
func NewHost(parent *mygo.Window) *Host {
	h := NewHostWithBackend(&nativeBackend{parent: parent})
	if parent != nil {
		parent.OnClosed(h.CloseAll)
	}
	return h
}

type nativeBackend struct{ parent *mygo.Window }
type nativeSurface struct {
	window  *mygo.Window
	dismiss func() bool
}

func (b *nativeBackend) Open(s Spec, view func(*ui.Context), onClosed func()) (surface Surface, err error) {
	if b.parent == nil {
		return nil, errors.New("native flyout needs an explicit parent window")
	}
	if b.parent.IsDestroyed() {
		return nil, errors.New("native flyout parent has closed")
	}
	var created *mygo.Window
	defer func() {
		if value := recover(); value != nil {
			if created != nil && !created.IsDestroyed() {
				created.Destroy()
			}
			surface = nil
			err = fmt.Errorf("native flyout creation failed: %v", value)
		}
	}()
	w := mygo.NewFlyout(mygo.FlyoutOptions{Parent: b.parent, Anchor: nativeRect(s.Anchor), Placement: s.Placement, Width: s.Width, Height: s.Height, Gap: s.Gap, FitContent: s.FitContent, MinWidth: s.MinWidth, MinHeight: s.MinHeight, MaxWidth: s.MaxWidth, MaxHeight: s.MaxHeight, Focusable: s.Focusable, Popover: s.SystemPopover, CornerRadius: s.CornerRadius, Shadow: true, Hidden: true, Content: ui.View(view)})
	created = w
	native := &nativeSurface{window: w}
	w.OnClose(func(event *mygo.CloseEvent) {
		if native.dismiss != nil && native.dismiss() {
			event.PreventDefault()
		}
	})
	w.OnClosed(onClosed)
	w.SetIgnoreMouseEvents(s.IgnoreMouseEvents)
	if s.ShowInactive {
		w.ShowInactive()
	} else {
		w.Show()
	}
	return native, nil
}
func (s *nativeSurface) SetDismissHandler(handler func() bool) { s.dismiss = handler }
func (b *nativeBackend) InvalidateParent() {
	if b.parent != nil && !b.parent.IsDestroyed() {
		b.parent.Invalidate()
	}
}
func (s *nativeSurface) SetAnchor(r ui.Rect)       { s.window.SetAnchor(nativeRect(r)) }
func (s *nativeSurface) SetSize(width, height int) { s.window.SetSize(width, height) }
func (s *nativeSurface) Invalidate()               { s.window.Invalidate() }
func (s *nativeSurface) Focus() {
	if !s.window.IsDestroyed() {
		s.window.Focus()
	}
}
func (s *nativeSurface) Close() {
	if !s.window.IsDestroyed() {
		// Closing from the flyout's own UI callback must not destroy the surface
		// while it is building/painting. Update queues deterministic owned-window
		// cleanup after the current frame and does not redraw a destroyed window.
		s.window.Update(func() {
			if !s.window.IsDestroyed() {
				s.window.Destroy()
			}
		})
	}
}
func (s *nativeSurface) NativeWindow() *mygo.Window { return s.window }
func nativeRect(r ui.Rect) mygo.Rectangle {
	x, y := int(math.Floor(float64(r.X))), int(math.Floor(float64(r.Y)))
	return mygo.Rectangle{X: x, Y: y, Width: max(0, int(math.Ceil(float64(r.X+r.W)))-x), Height: max(0, int(math.Ceil(float64(r.Y+r.H)))-y)}
}
