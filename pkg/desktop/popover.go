// Package desktop bridges shadcn components to independent native windows.
// Each Host belongs to one explicit parent window; it has no global current
// window or shared UI Context. Use it on the UI thread. Background work must
// enter through the parent's Window.Update before changing application state.
package desktop

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Surface is one native flyout. Close may deliver its closed callback later.
// Methods run on the same UI thread as the owning Host.
type Surface interface {
	SetAnchor(ui.Rect)
	Invalidate()
	Close()
}

// SizeSurface optionally resizes a live Flyout without discarding its child
// tree. Only the requested width/height use this capability; other immutable
// native options still require a replacement. Dimensions are window DIPs.
type SizeSurface interface{ SetSize(width, height int) }

// FocusSurface reports actual keyboard ownership, independently of a native
// Flyout's inherited active appearance. Backends may expose it for focus-aware
// compositions; a missing capability is reported as unknown by FocusState.
type FocusSurface interface{ IsFocused() bool }

// MaterialSurface updates a live native window material without remounting its content.
type MaterialSurface interface{ SetVibrancy(mygo.Vibrancy) }

// DismissInterceptor is an optional surface capability. The callback returns
// true when Host retained the surface to animate its dismissal, so the native
// close event must be prevented. Backends without it preserve immediate close
// behavior; the native MyGo adapter implements it for plain Flyouts.
type DismissInterceptor interface{ SetDismissHandler(func() bool) }

// Backend is injectable so lifecycle and independent-context behavior can be
// tested without a desktop session. Open must install onClosed before showing
// the surface. It supplies a NEW Context to view, never the parent's context.
type Backend interface {
	Open(Spec, func(*ui.Context), func()) (Surface, error)
	InvalidateParent()
}

// Spec is the normalized native configuration. Parent ownership lives in the
// backend, so a test backend need not fabricate native Window pointers.
type Spec struct {
	Anchor                                   ui.Rect
	Placement                                mygo.Placement
	Width, Height, Gap                       int
	FitContent                               bool
	MinWidth, MinHeight, MaxWidth, MaxHeight int
	Focusable, SystemPopover                 bool
	IgnoreMouseEvents, ShowInactive          bool
	CornerRadius                             int
	Vibrancy                                 mygo.Vibrancy
}

type PopoverProps struct {
	// Private profiles keep menu source motion separate from Popover's preset.
	menuMotion                               menuMotionKind
	menuIdentity                             string
	menuContext                              bool
	Label                                    string
	Placement                                mygo.Placement
	Width, Height, Gap                       int
	FitContent                               bool
	MinWidth, MinHeight, MaxWidth, MaxHeight int
	NonFocusable                             bool
	// NoFocusRestore leaves the parent's existing keyboard focus unchanged on
	// dismissal. Hover-opened inactive windows use it to avoid turning a
	// pointer-only preview into anchor focus (and immediately reopening it).
	NoFocusRestore bool
	// IgnoreMouseEvents makes tooltips transparent to pointer input. ShowInactive
	// leaves initial keyboard focus in the parent while permitting later focus.
	IgnoreMouseEvents, ShowInactive bool
	// Plain opts out of macOS's material, arrow and system popover animation.
	// A plain flyout still has its own native window and system shadow. On
	// Windows/Linux the system-popover option has no effect.
	Plain bool
	// SystemStyle keeps the independent child's platform theme instead of
	// inheriting shadcn tokens from the parent. NativePopover sets this; use
	// ui.* controls inside its callback to retain platform control styling.
	SystemStyle bool
	// ContentPadding is 16 DIPs by default. A pointer to zero gives menus and
	// other compound content complete control of their internal spacing.
	ContentPadding *float32
	// ContentRole replaces the default dialog role; use RoleGroup for a
	// container whose content exposes its own menu/list semantics.
	ContentRole ui.Role
	// ManualTrigger leaves open entirely caller-controlled rather than toggling
	// it for a click on the anchor. Useful for hover cards and tooltips.
	ManualTrigger bool
	// Revision invalidates the child when external content changes. Its own
	// native controls redraw themselves, and theme/anchor changes are detected.
	Revision uint64
	// OnDismiss runs for native outside/Escape/parent closure, after open has
	// become false. Programmatic Close and disappearing anchors do not call it.
	OnDismiss func()
}

type Host struct {
	backend Backend
	entries map[string]*entry
	frame   uint64
}
type entry struct {
	open                                     *bool
	props                                    PopoverProps
	spec                                     Spec
	content                                  func(*ui.Context)
	tokens                                   theme.Tokens
	menu                                     theme.MenuSnapshot
	nativeTheme                              *ui.Theme
	reduced                                  bool
	anchor                                   ui.Handle
	surface                                  Surface
	generation                               uint64
	seen                                     uint64
	awaitedAnchor                            bool
	pendingParent                            bool
	observedAnchor                           ui.Rect
	padding                                  float32
	presentation                             ui.MenuPresentation
	direction                                ui.Direction
	closing, nativeDismiss, removeAfterClose bool
	presented                                bool
}

// NewHostWithBackend creates an isolated host, primarily for backend adapters
// and deterministic tests. A nil backend returns an error from Popover.
func NewHostWithBackend(backend Backend) *Host {
	return &Host{backend: backend, entries: make(map[string]*entry)}
}

// BeginFrame/EndFrame bracket the parent view. EndFrame closes popovers whose
// anchors were not registered in this frame, such as after navigation. Calls
// can be omitted only when the caller explicitly closes disappearing content.
func (h *Host) BeginFrame() {
	if h != nil {
		h.frame++
	}
}
func (h *Host) EndFrame() {
	if h == nil {
		return
	}
	for key, e := range h.entries {
		if e.seen != h.frame {
			h.closeEntry(e)
			delete(h.entries, key)
		}
	}
}

// Popover synchronizes a controlled native flyout with an anchor. Its child
// callback must build using the child Context parameter; never close over and
// reuse the parent's Context. Normal Go app state may be shared on the UI thread.
func (h *Host) Popover(c *ui.Context, key string, anchor ui.Element, open *bool, p PopoverProps, content func(*ui.Context)) error {
	if h == nil || h.backend == nil {
		return errors.New("desktop: popover needs a parent-bound Host")
	}
	if key == "" {
		return errors.New("desktop: popover needs a stable nonempty key")
	}
	if c == nil || !c.Valid() || !anchor.Valid() {
		return errors.New("desktop: popover must be registered with a live parent context and anchor")
	}
	if open == nil {
		return errors.New("desktop: popover needs caller-owned open state")
	}
	if h.entries == nil {
		h.entries = make(map[string]*entry)
	}
	e := h.entries[key]
	if e == nil {
		e = &entry{}
		h.entries[key] = e
	}
	e.seen = h.frame
	if e.open != nil && e.open != open {
		*e.open = false
	}
	e.open = open
	e.content = content
	anchor.Bind(&e.anchor)
	if !p.ManualTrigger && anchor.Clicked() {
		*open = !*open
	}
	if p.NonFocusable {
		anchor.HasPopup(ui.PopupNone).Expanded(false)
	} else {
		anchor.HasPopup(ui.PopupDialog).Expanded(*open)
	}
	previousMenu := e.menu
	e.menu = theme.MenuOf(c)
	if !*open {
		e.props = p
		e.reduced = motion.Reduced(c)
		h.beginClose(e, false, false)
		return nil
	}
	if e.closing {
		// The same surface and animation identity survive a close/open reversal.
		e.closing, e.nativeDismiss, e.removeAfterClose = false, false, false
		if e.surface != nil {
			e.surface.Invalidate()
		}
	}
	// Read the anchor's committed geometry after parent layout. A single
	// correction frame updates a popup after programmatic layout/scroll moves;
	// stable anchors schedule nothing. This separate, invisible paint observer
	// preserves any custom Draw/DrawOver functions on the caller's anchor.
	c.Root().Children(func() {
		ui.Box(c.Key("shadcn.desktop.anchor:" + key)).Absolute().Fill().IgnoreInput(true).AccessHidden(true).Draw(func(_ *ui.Painter, _ ui.Rect) {
			if h.entries[key] != e || e.open == nil || !*e.open {
				return
			}
			bounds := e.anchor.Bounds(c)
			// Resolve only within this paint pass to include visual transforms
			// (scroll/translation/scale), while retaining only the stable Handle.
			if current := c.Resolve(e.anchor); current.Valid() {
				bounds = current.Bounds()
			}
			if bounds != e.observedAnchor {
				e.observedAnchor = bounds
				if bounds != e.spec.Anchor {
					c.Invalidate()
				}
			}
		})
	})
	tokens := *theme.TokensOf(c)
	spec, err := normalize(p, anchor.Bounds(), tokens.Radius)
	if err != nil {
		return err
	}
	if p.menuMotion != menuMotionNone && !p.SystemStyle && !spec.SystemPopover && e.menu.Translucent() && !c.Preferences().ReduceTransparency && !c.Preferences().HighContrast {
		spec.Vibrancy = mygo.VibrancyMenu
	}
	if spec.Anchor.W <= 0 || spec.Anchor.H <= 0 {
		if !e.awaitedAnchor {
			e.awaitedAnchor = true
			c.Invalidate()
		}
		return nil
	}
	e.awaitedAnchor = false
	reduced := c.Instant() || c.Preferences().ReduceMotion
	padding := float32(16)
	if p.ContentPadding != nil {
		padding = max(0, *p.ContentPadding)
	}
	presentation := c.MenuPresentation()
	menuSnapshot := theme.MenuOf(c)
	changed := e.nativeTheme == nil || previousMenu != menuSnapshot || e.tokens != tokens || e.reduced != reduced || e.props.Label != p.Label || e.props.Revision != p.Revision || e.props.SystemStyle != p.SystemStyle || e.padding != padding || e.props.ContentRole != p.ContentRole || e.props.menuMotion != p.menuMotion || e.props.menuIdentity != p.menuIdentity || e.props.menuContext != p.menuContext || e.direction != c.Direction() || !p.SystemStyle && e.presentation != presentation
	if e.nativeTheme == nil || e.tokens != tokens || previousMenu != menuSnapshot {
		e.tokens = tokens
		e.nativeTheme = tokens.ThemeWithMenus(menuSnapshot.ForNativeWindow())
	}
	e.reduced = reduced
	e.padding, e.presentation = padding, presentation
	e.direction = c.Direction()
	// Changing immutable window options needs a fresh native surface. Anchor,
	// complete tokens and content do not require reopening an existing one.
	previous := e.spec
	previous.Anchor = spec.Anchor
	var material MaterialSurface
	if e.surface != nil && previous.Vibrancy != spec.Vibrancy {
		if capability, ok := e.surface.(MaterialSurface); ok {
			previous.Vibrancy = spec.Vibrancy
			material = capability
			changed = true
		}
	}
	var resize SizeSurface
	if e.surface != nil && (previous.Width != spec.Width || previous.Height != spec.Height) {
		if sized, ok := e.surface.(SizeSurface); ok {
			withoutSize := previous
			withoutSize.Width, withoutSize.Height = spec.Width, spec.Height
			if withoutSize == spec {
				previous = withoutSize
				resize = sized
			}
		}
	}
	if e.surface != nil && previous != spec {
		h.releaseSurface(e)
	}
	e.props = p
	oldAnchor := e.spec.Anchor
	e.spec = spec
	if e.surface != nil {
		if material != nil {
			material.SetVibrancy(spec.Vibrancy)
		}
		if resize != nil {
			resize.SetSize(spec.Width, spec.Height)
		}
		h.configureDismiss(e, e.surface, e.generation)
		if oldAnchor != spec.Anchor {
			e.surface.SetAnchor(spec.Anchor)
		}
		if changed {
			e.surface.Invalidate()
		}
		return nil
	}
	e.generation++
	generation := e.generation
	surface, err := h.backend.Open(spec, func(child *ui.Context) { h.view(child, e, generation) }, func() {
		if e.generation != generation {
			return
		}
		e.generation++
		e.surface = nil
		remove := e.removeAfterClose
		e.presented = false
		e.closing, e.nativeDismiss, e.removeAfterClose = false, false, false
		if remove {
			h.deleteEntry(e)
		}
		if e.open != nil {
			*e.open = false
		}
		if !e.props.NonFocusable && !e.props.NoFocusRestore {
			e.anchor.Focus()
		}
		if e.props.OnDismiss != nil {
			e.props.OnDismiss()
		}
		h.backend.InvalidateParent()
	})
	if err != nil {
		*open = false
		return fmt.Errorf("desktop: opening %q: %w", key, err)
	}
	if surface == nil {
		*open = false
		return errors.New("desktop: backend returned an absent native surface")
	}
	if e.generation != generation || !*open {
		surface.Close()
		return nil
	}
	e.surface = surface
	h.configureDismiss(e, surface, generation)
	return nil
}

func (h *Host) configureDismiss(e *entry, surface Surface, generation uint64) {
	if interceptor, ok := surface.(DismissInterceptor); ok {
		if e.props.SystemStyle || e.spec.SystemPopover {
			interceptor.SetDismissHandler(nil)
			return
		}
		interceptor.SetDismissHandler(func() bool {
			if e.generation != generation || e.surface != surface {
				return false
			}
			return h.beginClose(e, true, false)
		})
	}
}

func (h *Host) view(c *ui.Context, e *entry, generation uint64) {
	if e.generation != generation {
		return
	}
	if !e.props.SystemStyle {
		c.SetTheme(e.nativeTheme)
		c.SetMenuPresentation(e.presentation)
	}
	if e.reduced {
		c.DisableAnimations()
	}
	t := e.tokens
	root := ui.Column(c).Gap(12).Padding(e.padding).Background(t.Popover).TextColor(t.PopoverForeground).Radius(t.Radius).Direction(e.direction)
	frame := menuFrame{}
	if e.props.menuMotion != menuMotionNone && !e.props.SystemStyle && !e.spec.SystemPopover {
		root.Background(ui.Transparent).Gap(0).Padding(0)
		state := ui.Local(root, menuRootMotionKey{}, func() menuRootMotion { return menuRootMotion{identity: e.props.menuIdentity} })
		if state.identity != e.props.menuIdentity {
			*state = menuRootMotion{identity: e.props.menuIdentity}
		}
		attachment := menuAttachment{x: .5, y: .5}
		observed := menuGeometry{context: e.props.menuContext}
		if geometry, ok := e.surface.(GeometrySurface); ok {
			if anchor, content, known := geometry.Geometry(); known {
				observed.anchor, observed.content, observed.known = anchor, content, true
				padding := float32(5)
				if observed.context {
					padding = 0
				}
				attachment = menuActualAttachment(anchor, content, padding)
			}
		}
		pose := state.tween.update(c, !e.closing, attachment)
		frame = menuFrame{active: true, closing: e.closing, pose: pose, attachment: attachment, geometry: observed}
		e.presented = true
		if e.closing {
			root.IgnoreInput(true).AccessHidden(true)
		}
		if e.closing && pose.opacity <= 0 {
			h.finishExit(e)
			return
		}
	} else if !e.props.SystemStyle && !e.spec.SystemPopover {
		// MyGo intentionally skips presence animation on a view's very first
		// tree. Measure invisibly once, then add the visible keyed panel on the
		// next frame so a newly created native window really enters. Measuring
		// still includes content for FitContent; reduced motion shows it at once.
		entered := ui.Local(c.Root(), "shadcn.desktop.entered", func() bool { return false })
		if !*entered && !motion.Reduced(c) {
			root.Key("measuring").Opacity(0).IgnoreInput(true)
			*entered = true
			c.Invalidate()
		} else {
			root.Key("visible").Transition(motion.Popup(c))
			e.presented = true
		}
		// Keep one measurable native surface alive while its styled content
		// exits. Input and accessibility leave immediately, pixels leave over
		// 100 ms; switching back to open reverses from the current opacity.
		target, duration := float32(1), 150*time.Millisecond
		if e.closing {
			target, duration = 0, 100*time.Millisecond
		}
		progress := root.AnimateWith("shadcn.desktop.presence", target, duration, motion.EaseOut)
		if e.presented {
			root.Opacity(progress).Scale(.95+.05*progress, .95+.05*progress).Translate(0, -4*(1-progress))
		}
		if e.closing {
			root.IgnoreInput(true).AccessHidden(true)
		}
		if e.closing && (motion.Reduced(c) || progress <= 0) {
			h.finishExit(e)
			return
		}
	}
	if e.props.SystemStyle {
		native := c.Theme()
		root.Background(native.Surface).TextColor(native.Text).Radius(native.Radius)
	}
	if !e.spec.FitContent {
		root.Fill()
	} else {
		root.FillWidth()
	}
	if c.Vibrancy() {
		root.Background(ui.Transparent)
		if e.props.menuMotion != menuMotionNone && e.menu.Translucent() && !c.Preferences().ReduceTransparency && !c.Preferences().HighContrast {
			// The native root otherwise paints the theme background over the
			// system material. Menu panels supply their own 70% tint above it.
			c.Root().Background(ui.Transparent)
		}
	}
	label := e.props.Label
	if label == "" {
		label = "Popover"
	}
	role := e.props.ContentRole
	if role == ui.RoleAuto {
		role = ui.RoleDialog
		if e.props.NonFocusable {
			role = ui.RoleGroup
		}
	}
	root.Role(role).Label(label)
	root.ObserveInput(func(ev ui.InputEvent) {
		switch ev.Kind {
		case ui.InputKeyDown, ui.InputText, ui.InputCommand, ui.InputPointerUp:
			e.pendingParent = true
		}
	}).Children(func() {
		if e.content != nil {
			if frame.active {
				withMenuFrame(c, frame, func() { e.content(c) })
			} else {
				e.content(c)
			}
		}
	})
	// Native input arrives before controls process their Clicked/Changed
	// state. Notify the parent only after the child content has applied it.
	// Closing itself by writing *open=false is also observed here; the parent
	// then reconciles the native window after this child's frame has finished.
	if e.pendingParent || e.open != nil && !*e.open && !e.closing {
		e.pendingParent = false
		h.backend.InvalidateParent()
	}
}

// Popover opens an independent native flyout with uniform shadcn tokens,
// surfaces and geometry. It can extend outside its parent window.
func Popover(c *ui.Context, h *Host, key string, anchor ui.Element, open *bool, p PopoverProps, content func(*ui.Context)) error {
	p.Plain = true
	p.SystemStyle = false
	return h.Popover(c, key, anchor, open, p, content)
}

// NativePopover chooses system popover material/shape/animation and the child
// window's own platform theme. It deliberately does not inherit a parent
// shadcn preset. Platform ui.* controls may be composed in its child callback.
func NativePopover(c *ui.Context, h *Host, key string, anchor ui.Element, open *bool, p PopoverProps, content func(*ui.Context)) error {
	p.Plain = false
	p.SystemStyle = true
	return h.Popover(c, key, anchor, open, p, content)
}

func normalize(p PopoverProps, anchor ui.Rect, radius float32) (Spec, error) {
	for _, v := range []float32{anchor.X, anchor.Y, anchor.W, anchor.H, radius} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return Spec{}, errors.New("desktop: nonfinite anchor or radius")
		}
	}
	if p.ContentPadding != nil {
		padding := float64(*p.ContentPadding)
		if math.IsNaN(padding) || math.IsInf(padding, 0) || padding < 0 {
			return Spec{}, errors.New("desktop: content padding must be finite and nonnegative")
		}
	}
	if p.Width < 0 || p.Height < 0 || p.Gap < 0 || p.MinWidth < 0 || p.MinHeight < 0 || p.MaxWidth < 0 || p.MaxHeight < 0 {
		return Spec{}, errors.New("desktop: flyout dimensions and gap must be nonnegative")
	}
	if p.MaxWidth > 0 && p.MinWidth > p.MaxWidth || p.MaxHeight > 0 && p.MinHeight > p.MaxHeight {
		return Spec{}, errors.New("desktop: flyout minimum exceeds maximum")
	}
	placement := p.Placement
	if placement == "" {
		placement = mygo.PlacementBottomStart
	}
	switch placement {
	case mygo.PlacementBottomStart, mygo.PlacementBottom, mygo.PlacementBottomEnd, mygo.PlacementTopStart, mygo.PlacementTop, mygo.PlacementTopEnd, mygo.PlacementRightStart, mygo.PlacementRight, mygo.PlacementRightEnd, mygo.PlacementLeftStart, mygo.PlacementLeft, mygo.PlacementLeftEnd:
	default:
		return Spec{}, fmt.Errorf("desktop: unknown placement %q", placement)
	}
	width, height, gap := p.Width, p.Height, p.Gap
	if width == 0 {
		width = 288
	}
	if height == 0 {
		height = 200
	}
	if gap == 0 {
		gap = 4
	}
	return Spec{Anchor: anchor, Placement: placement, Width: width, Height: height, Gap: gap, FitContent: p.FitContent, MinWidth: p.MinWidth, MinHeight: p.MinHeight, MaxWidth: p.MaxWidth, MaxHeight: p.MaxHeight, Focusable: !p.NonFocusable, SystemPopover: !p.Plain && !p.NonFocusable, IgnoreMouseEvents: p.IgnoreMouseEvents, ShowInactive: p.ShowInactive, CornerRadius: max(0, int(math.Round(float64(radius))))}, nil
}

// beginClose returns whether native destruction is being retained for an exit.
// A removed anchor or destroyed parent uses closeEntry instead: no orphan
// window or animation timer may outlive its owning UI.
func (h *Host) beginClose(e *entry, native, remove bool) bool {
	if e.open != nil {
		*e.open = false
	}
	if e.closing {
		e.nativeDismiss = e.nativeDismiss || native
		e.removeAfterClose = e.removeAfterClose || remove
		return true
	}
	_, capable := e.surface.(DismissInterceptor)
	if e.surface != nil && capable && e.presented && !e.reduced && (e.props.menuMotion == menuMotionNone || e.menu.Animated()) && !e.props.SystemStyle && !e.spec.SystemPopover {
		e.closing, e.nativeDismiss, e.removeAfterClose = true, native, remove
		e.surface.Invalidate()
		h.backend.InvalidateParent()
		return true
	}
	if native {
		return false
	}
	h.closeEntry(e)
	if remove {
		h.deleteEntry(e)
	}
	return false
}
func (h *Host) deleteEntry(e *entry) {
	for key, current := range h.entries {
		if current == e {
			delete(h.entries, key)
			return
		}
	}
}
func (h *Host) finishExit(e *entry) {
	native, remove := e.nativeDismiss, e.removeAfterClose
	e.closing, e.nativeDismiss, e.removeAfterClose = false, false, false
	h.closeEntry(e)
	if remove {
		h.deleteEntry(e)
	}
	if native && e.props.OnDismiss != nil {
		e.props.OnDismiss()
	}
	h.backend.InvalidateParent()
}

func (h *Host) releaseSurface(e *entry) {
	e.generation++
	surface := e.surface
	e.surface = nil
	e.presented = false
	e.closing, e.nativeDismiss, e.removeAfterClose = false, false, false
	if surface != nil {
		if interceptor, ok := surface.(DismissInterceptor); ok {
			interceptor.SetDismissHandler(nil)
		}
		surface.Close()
	}
}
func (h *Host) closeEntry(e *entry) {
	if e.open != nil {
		*e.open = false
	}
	if e.surface != nil && !e.props.NonFocusable && !e.props.NoFocusRestore {
		e.anchor.Focus()
	}
	e.awaitedAnchor = false
	h.releaseSurface(e)
}

// Close is a programmatic dismissal. A late native OnClosed from this surface
// cannot close a later flyout with the same key.
func (h *Host) Close(key string) {
	if h == nil {
		return
	}
	if e := h.entries[key]; e != nil {
		h.beginClose(e, false, true)
		h.backend.InvalidateParent()
	}
}
func (h *Host) CloseAll() {
	if h == nil {
		return
	}
	hadEntries := len(h.entries) > 0
	for key, e := range h.entries {
		h.closeEntry(e)
		delete(h.entries, key)
	}
	if hadEntries && h.backend != nil {
		h.backend.InvalidateParent()
	}
}

// Refresh redraws external content without moving or recreating the window.
func (h *Host) Refresh(key string) {
	if h == nil {
		return
	}
	if e := h.entries[key]; e != nil && e.surface != nil {
		e.surface.Invalidate()
	}
}
func (h *Host) IsOpen(key string) bool {
	if h == nil {
		return false
	}
	e := h.entries[key]
	return e != nil && e.surface != nil && !e.closing
}

// IsPresent includes a surface retained briefly for its exit animation.
func (h *Host) IsPresent(key string) bool {
	if h == nil {
		return false
	}
	e := h.entries[key]
	return e != nil && e.surface != nil
}

// Focus enters an already-present, focusable flyout on explicit keyboard intent.
// Hover opening can remain inactive until the user asks to enter its content.
func (h *Host) Focus(key string) bool {
	if h == nil {
		return false
	}
	e := h.entries[key]
	if e == nil || e.surface == nil || e.closing || e.props.NonFocusable {
		return false
	}
	if focusable, ok := e.surface.(interface{ Focus() }); ok {
		focusable.Focus()
		return true
	}
	return false
}

// FocusState reports a live surface's actual keyboard focus when the backend
// supports FocusSurface. Active appearance alone does not prove keyboard focus.
func (h *Host) FocusState(key string) (focused, known bool) {
	if h == nil {
		return false, false
	}
	entry := h.entries[key]
	if entry == nil || entry.surface == nil || entry.closing {
		return false, false
	}
	if surface, ok := entry.surface.(FocusSurface); ok {
		return surface.IsFocused(), true
	}
	return false, false
}

// ApplicationState reports whether the backend's owning application is active.
// ApplicationActive is an optional Backend capability; unknown backends do not
// infer application activity from a Flyout's active appearance.
func (h *Host) ApplicationState() (active, known bool) {
	if h != nil {
		if backend, ok := h.backend.(interface{ ApplicationActive() bool }); ok {
			return backend.ApplicationActive(), true
		}
	}
	return false, false
}

// NativeWindow exposes the MyGo surface for desktop-only geometry/automation
// integration. An injected backend that has no native window returns nil.
func (h *Host) NativeWindow(key string) *mygo.Window {
	if h == nil {
		return nil
	}
	if e := h.entries[key]; e != nil {
		if native, ok := e.surface.(interface{ NativeWindow() *mygo.Window }); ok {
			return native.NativeWindow()
		}
	}
	return nil
}
