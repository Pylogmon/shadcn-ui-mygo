package desktop

import "github.com/egoist/mygo/ui"

// GeometrySurface optionally reports a live flyout's observed geometry. Both
// rectangles use screen DIPs: anchor is the rectangle actually supplied to the
// native window, translated from its parent's content origin, and content is
// the child's current content bounds. They do not report a resolved placement
// or a platform popover's arrow position.
//
// Like Surface, Geometry must be called on the owning UI thread. known is false
// when either window is unavailable, the rectangles have no area, or their
// coordinate scales cannot safely be compared. Backends without this capability
// can still open and animate flyouts without claiming observed placement.
type GeometrySurface interface {
	Geometry() (anchor, content ui.Rect, known bool)
}
