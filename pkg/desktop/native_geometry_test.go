package desktop

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestNativeFlyoutGeometryUsesObservedScreenBounds(t *testing.T) {
	tests := []struct {
		name                    string
		local                   ui.Rect
		parent, child           mygo.Rectangle
		wantAnchor, wantContent ui.Rect
	}{
		{"negative-screen-and-fractional-anchor", ui.Rect{X: 10.5, Y: -20.25, W: 40.25, H: 10.5}, mygo.Rectangle{X: -1200, Y: 100, Width: 900, Height: 600}, mygo.Rectangle{X: -1190, Y: 45, Width: 240, Height: 120}, ui.Rect{X: -1190, Y: 79, W: 41, H: 12}, ui.Rect{X: -1190, Y: 45, W: 240, H: 120}},
		{"collision-flipped-content-is-not-reconstructed", ui.Rect{X: 60, Y: 80, W: 40, H: 20}, mygo.Rectangle{X: 500, Y: 300, Width: 600, Height: 400}, mygo.Rectangle{X: 410, Y: 100, Width: 250, Height: 250}, ui.Rect{X: 560, Y: 380, W: 40, H: 20}, ui.Rect{X: 410, Y: 100, W: 250, H: 250}},
		{"parent-move-uses-current-origin", ui.Rect{X: 60, Y: 80, W: 40, H: 20}, mygo.Rectangle{X: -500, Y: -300, Width: 600, Height: 400}, mygo.Rectangle{X: -440, Y: -196, Width: 250, Height: 250}, ui.Rect{X: -440, Y: -220, W: 40, H: 20}, ui.Rect{X: -440, Y: -196, W: 250, H: 250}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, c, known := nativeFlyoutGeometry(nativeRect(tt.local), tt.parent, tt.child)
			if !known || a != tt.wantAnchor || c != tt.wantContent {
				t.Fatalf("got %+v %+v %v, want %+v %+v", a, c, known, tt.wantAnchor, tt.wantContent)
			}
		})
	}
}

func TestNativeFlyoutGeometryUnknownHasNoInventedRect(t *testing.T) {
	good := mygo.Rectangle{X: 1, Y: 2, Width: 100, Height: 80}
	tests := []struct {
		name                  string
		anchor, parent, child mygo.Rectangle
	}{
		{"zero-anchor", mygo.Rectangle{}, good, good},
		{"zero-parent", good, mygo.Rectangle{}, good},
		{"zero-child", good, good, mygo.Rectangle{}},
		{"negative-child", good, good, mygo.Rectangle{Width: -1, Height: 80}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, c, known := nativeFlyoutGeometry(tt.anchor, tt.parent, tt.child)
			if known || a != (ui.Rect{}) || c != (ui.Rect{}) {
				t.Fatalf("unknown geometry leaked %+v %+v %v", a, c, known)
			}
		})
	}
	var surface GeometrySurface = (*nativeSurface)(nil)
	a, c, known := surface.Geometry()
	if known || a != (ui.Rect{}) || c != (ui.Rect{}) {
		t.Fatalf("nil native surface geometry %+v %+v %v", a, c, known)
	}
	surface = &nativeSurface{}
	a, c, known = surface.Geometry()
	if known || a != (ui.Rect{}) || c != (ui.Rect{}) {
		t.Fatalf("empty native surface geometry %+v %+v %v", a, c, known)
	}
}

func TestNativeSurfaceGeometryObserversReleaseOnceIncludingReentrantRemoval(t *testing.T) {
	surface := &nativeSurface{geometry: nativeSurfaceGeometry{known: true}}
	removed := [3]int{}
	surface.geometryOff = []func(){
		func() { removed[0]++; surface.releaseGeometryObservers() },
		func() { removed[1]++ },
		func() { removed[2]++ },
	}
	surface.releaseGeometryObservers()
	surface.releaseGeometryObservers()
	if removed != [3]int{1, 1, 1} || surface.geometryOff != nil || surface.geometry != (nativeSurfaceGeometry{}) {
		t.Fatalf("geometry listeners survived release or ran twice: removed=%v state=%+v", removed, surface)
	}
}

func TestNativeGeometryRecognizesPossibleWaylandWithoutClaimingBackend(t *testing.T) {
	for _, test := range []struct {
		name, platform, display, socket, session string
		possible                                 bool
	}{
		{"display", "linux", "wayland-0", "", "", true},
		{"inherited-socket", "linux", "", "7", "", true},
		{"session", "linux", "", "", " WayLand ", true},
		{"x11-session-with-wayland-signal", "linux", "wayland-0", "", "x11", true},
		{"x11-only", "linux", "", "", "x11", false},
		{"no-observed-signal", "linux", "", "", "", false},
		{"mac-with-stale-environment", "darwin", "wayland-0", "7", "wayland", false},
		{"windows-with-stale-environment", "windows", "wayland-0", "7", "wayland", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := nativeGeometryPossibleWayland(test.platform, test.display, test.socket, test.session); got != test.possible {
				t.Fatalf("possible Wayland=%v, want %v", got, test.possible)
			}
		})
	}
}
