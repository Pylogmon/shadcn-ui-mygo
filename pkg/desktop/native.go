package desktop

import (
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"

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

func (b *nativeBackend) ApplicationActive() bool {
	return b.parent != nil && !b.parent.IsDestroyed() && mygo.App.IsActive()
}

type nativeSurface struct {
	window      *mygo.Window
	dismiss     func() bool
	anchor      mygo.Rectangle
	geometry    nativeSurfaceGeometry
	geometryOff []func()
}

type nativeSurfaceGeometry struct {
	anchor, content ui.Rect
	known           bool
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
	anchor := nativeRect(s.Anchor)
	w := mygo.NewFlyout(nativeFlyoutOptions(b.parent, s, view))
	created = w
	native := &nativeSurface{window: w, anchor: anchor}
	w.OnClose(func(event *mygo.CloseEvent) {
		if native.dismiss != nil && native.dismiss() {
			event.PreventDefault()
		}
	})
	w.OnClosed(func() {
		native.releaseGeometryObservers()
		native.dismiss = nil
		onClosed()
	})
	native.observeGeometry(b.parent)
	w.SetIgnoreMouseEvents(s.IgnoreMouseEvents)
	if s.ShowInactive {
		w.ShowInactive()
	} else {
		w.Show()
	}
	return native, nil
}

func nativeFlyoutOptions(parent *mygo.Window, s Spec, view func(*ui.Context)) mygo.FlyoutOptions {
	return mygo.FlyoutOptions{
		Parent: parent, Anchor: nativeRect(s.Anchor), Placement: s.Placement,
		Width: s.Width, Height: s.Height, Gap: s.Gap, FitContent: s.FitContent,
		MinWidth: s.MinWidth, MinHeight: s.MinHeight, MaxWidth: s.MaxWidth, MaxHeight: s.MaxHeight,
		Focusable: s.Focusable, Popover: s.SystemPopover, Vibrancy: s.Vibrancy,
		CornerRadius: s.CornerRadius, Shadow: !s.DisableShadow, Hidden: true, Content: ui.View(view),
	}
}

func (s *nativeSurface) SetDismissHandler(handler func() bool) { s.dismiss = handler }
func (b *nativeBackend) InvalidateParent() {
	if b.parent != nil && !b.parent.IsDestroyed() {
		b.parent.Invalidate()
	}
}
func (s *nativeSurface) SetAnchor(r ui.Rect) {
	s.anchor = nativeRect(r)
	s.window.SetAnchor(s.anchor)
	s.geometryChanged()
}
func (s *nativeSurface) SetSize(width, height int) { s.window.SetSize(width, height) }
func (s *nativeSurface) Invalidate()               { s.window.Invalidate() }
func (s *nativeSurface) IsFocused() bool {
	return !s.window.IsDestroyed() && s.window.IsFocused()
}
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

// Geometry observes native bounds rather than treating the requested side as
// the result of collision handling. In particular, FitContent can resize and
// reposition a child after its initial frame.
func (s *nativeSurface) Geometry() (anchor, content ui.Rect, known bool) {
	if s == nil || s.window == nil || s.window.IsDestroyed() {
		return ui.Rect{}, ui.Rect{}, false
	}
	parent := s.window.Parent()
	if parent == nil || parent.IsDestroyed() {
		return ui.Rect{}, ui.Rect{}, false
	}
	if nativeGeometryPossibleWayland(runtime.GOOS, os.Getenv("WAYLAND_DISPLAY"), os.Getenv("WAYLAND_SOCKET"), os.Getenv("XDG_SESSION_TYPE")) {
		// There is no public resolved-compositor-position getter. Environment
		// hints identify a possible Wayland session, not the actual GTK backend;
		// keep placement unknown rather than infer it from unproven origins.
		return ui.Rect{}, ui.Rect{}, false
	}
	p, child := parent.ContentBounds(), s.window.ContentBounds()
	if runtime.GOOS == "windows" {
		// Windows bounds use the scale of the display containing most of each
		// window. Different scales cannot be added as one screen-DIP plane.
		pd, cd := mygo.Screen.DisplayMatching(p), mygo.Screen.DisplayMatching(child)
		if pd.ScaleFactor <= 0 || cd.ScaleFactor <= 0 || pd.ScaleFactor != cd.ScaleFactor {
			return ui.Rect{}, ui.Rect{}, false
		}
	}
	return nativeFlyoutGeometry(s.anchor, p, child)
}

func nativeGeometryPossibleWayland(platform, display, socket, session string) bool {
	return platform == "linux" && (display != "" || socket != "" || strings.EqualFold(strings.TrimSpace(session), "wayland"))
}

func nativeFlyoutGeometry(anchor, parent, child mygo.Rectangle) (ui.Rect, ui.Rect, bool) {
	if anchor.Width <= 0 || anchor.Height <= 0 || parent.Width <= 0 || parent.Height <= 0 || child.Width <= 0 || child.Height <= 0 {
		return ui.Rect{}, ui.Rect{}, false
	}
	return ui.Rect{X: float32(parent.X) + float32(anchor.X), Y: float32(parent.Y) + float32(anchor.Y), W: float32(anchor.Width), H: float32(anchor.Height)},
		ui.Rect{X: float32(child.X), Y: float32(child.Y), W: float32(child.Width), H: float32(child.Height)}, true
}

func (s *nativeSurface) observeGeometry(parent *mygo.Window) {
	s.geometry.anchor, s.geometry.content, s.geometry.known = s.Geometry()
	s.geometryOff = []func(){
		s.window.OnShow(s.geometryChanged),
		s.window.OnMove(s.geometryChanged),
		s.window.OnResize(s.geometryChanged),
		parent.OnMove(s.geometryChanged),
		parent.OnResize(s.geometryChanged),
		mygo.Screen.OnDisplaysChanged(s.geometryChanged),
	}
}

func (s *nativeSurface) geometryChanged() {
	if s.window == nil || s.window.IsDestroyed() {
		return
	}
	anchor, content, known := s.Geometry()
	next := nativeSurfaceGeometry{anchor: anchor, content: content, known: known}
	if next != s.geometry {
		s.geometry = next
		if s.window.IsVisible() {
			s.window.Invalidate()
		}
	}
}

func (s *nativeSurface) releaseGeometryObservers() {
	off := s.geometryOff
	s.geometryOff = nil
	for _, remove := range off {
		remove()
	}
	s.geometry = nativeSurfaceGeometry{}
}

func nativeRect(r ui.Rect) mygo.Rectangle {
	x, y := int(math.Floor(float64(r.X))), int(math.Floor(float64(r.Y)))
	return mygo.Rectangle{X: x, Y: y, Width: max(0, int(math.Ceil(float64(r.X+r.W)))-x), Height: max(0, int(math.Ceil(float64(r.Y+r.H)))-y)}
}

func (s *nativeSurface) SetVibrancy(value mygo.Vibrancy) {
	if !s.window.IsDestroyed() {
		s.window.SetVibrancy(value)
	}
}
