package gallery

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// This adapter implements only public desktop contracts. It queues native
// resize/invalidation/close work between frames and supplies observed screen
// rectangles independently of the requested Spec.Placement.
// Tester has no native Flyout FitContent negotiation hook. A roomy child
// canvas keeps all real panels visible; these checks make no claim about the
// platform's automatic window sizing. Panel pixels/AX/input remain native.
type menuMotionBackend struct {
	surfaces    []*menuMotionSurface
	parentDirty bool
	actualSide  string
	reduced     bool
}

type menuMotionSurface struct {
	backend                       *menuMotionBackend
	spec                          desktop.Spec
	view                          func(*ui.Context)
	closed                        func()
	dismiss                       func() bool
	r                             *ui.Tester
	dirty, closePending, didClose bool
}

func (b *menuMotionBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	s := &menuMotionSurface{backend: b, spec: spec, view: view, closed: closed, dirty: true}
	b.surfaces = append(b.surfaces, s)
	return s, nil
}
func (b *menuMotionBackend) InvalidateParent()               { b.parentDirty = true }
func (s *menuMotionSurface) SetAnchor(r ui.Rect)             { s.spec.Anchor = r; s.dirty = true }
func (s *menuMotionSurface) SetSize(w, h int)                { s.spec.Width, s.spec.Height = w, h; s.dirty = true }
func (s *menuMotionSurface) Invalidate()                     { s.dirty = true }
func (s *menuMotionSurface) Close()                          { s.closePending = true }
func (s *menuMotionSurface) Focus()                          { s.dirty = true }
func (s *menuMotionSurface) SetDismissHandler(f func() bool) { s.dismiss = f }
func (s *menuMotionSurface) canvasSize() (int, int) {
	return max(600, s.spec.Width), max(400, s.spec.Height)
}
func (s *menuMotionSurface) Geometry() (anchor, content ui.Rect, known bool) {
	anchor = s.spec.Anchor
	anchor.X += 120
	anchor.Y += 80
	w, h := s.canvasSize()
	if s.r != nil {
		w, h = s.r.Image().Bounds().Dx(), s.r.Image().Bounds().Dy()
	}
	content = ui.Rect{W: float32(w), H: float32(h)}
	switch s.backend.actualSide {
	case "left":
		content.X, content.Y = anchor.X-content.W-8, anchor.Y
	case "top":
		content.X, content.Y = anchor.X, anchor.Y-content.H-8
	case "bottom":
		content.X, content.Y = anchor.X, anchor.Y+anchor.H+8
	default:
		content.X, content.Y = anchor.X+anchor.W+8, anchor.Y
	}
	return anchor, content, true
}

type menuMotionDriver struct {
	s      *State
	b      *menuMotionBackend
	parent *ui.Tester
}

func newMenuMotionDriver(width int, native, reduced, dark, rtl bool) *menuMotionDriver {
	s := New()
	s.Page, s.Config.NativeMenus = "navigation", native
	if dark {
		s.Theme.Mode = theme.DarkMode
	}
	b := &menuMotionBackend{actualSide: "right", reduced: reduced}
	s.Popovers = desktop.NewHostWithBackend(b)
	d := &menuMotionDriver{s: s, b: b}
	d.parent = ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		s.Popovers.BeginFrame()
		defer s.Popovers.EndFrame()
		ui.Column(c).Fill().Padding(16).Gap(10).Children(func() {
			ui.Button(c, "Menu page anchor")
			ui.Scroll(c).Fill().Role(ui.RoleRegion).Label("Menu preview page").Children(func() {
				ui.Column(c).FillWidth().Padding(8).Children(func() {
					if s.Page == "navigation" {
						s.navigationPreview(c)
					} else {
						ui.Text(c, "Another preview")
					}
				})
			})
		})
	}, width, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
	if rtl {
		d.parent.SetDirection(ui.RightToLeft)
	}
	return d
}

func (d *menuMotionDriver) synchronizeClocks() {
	now := d.parent.Now()
	for _, s := range d.b.surfaces {
		if s.r != nil && !s.didClose && s.r.Now().After(now) {
			now = s.r.Now()
		}
	}
	if d.parent.Now().Before(now) {
		d.parent.Advance(now.Sub(d.parent.Now()))
	}
	for _, s := range d.b.surfaces {
		if s.r != nil && !s.didClose && s.r.Now().Before(now) {
			s.r.Advance(now.Sub(s.r.Now()))
		}
	}
}

func (d *menuMotionDriver) flush(t *testing.T) {
	t.Helper()
	for pass := 0; pass < 16; pass++ {
		work := false
		if d.b.parentDirty {
			d.b.parentDirty = false
			d.parent.Frame()
			work = true
		}
		for _, s := range d.b.surfaces {
			if s.didClose {
				continue
			}
			if s.closePending {
				s.closePending = false
				s.didClose = true
				s.closed()
				work = true
				continue
			}
			if s.r == nil {
				w, h := s.canvasSize()
				s.r = ui.NewTesterWith(s.view, w, h, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: d.b.reduced}})
				work = true
			}
			w, h := s.canvasSize()
			if s.r.Image().Bounds().Dx() != w || s.r.Image().Bounds().Dy() != h {
				s.r.SetSize(w, h)
				work = true
			}
			if s.dirty {
				s.dirty = false
				s.r.Frame()
				work = true
			}
		}
		d.synchronizeClocks()
		if !work {
			return
		}
	}
	t.Fatal("public Flyout adapter did not settle queued native work")
}

func (d *menuMotionDriver) advance(t *testing.T, elapsed time.Duration) {
	t.Helper()
	d.synchronizeClocks()
	d.parent.Advance(elapsed)
	for _, s := range d.b.surfaces {
		if s.r != nil && !s.didClose {
			s.r.Advance(elapsed)
		}
	}
	d.flush(t)
}

func menuMotionReveal(t *testing.T, r *ui.Tester, name string) ui.Rect {
	t.Helper()
	for attempt := 0; attempt < 8; attempt++ {
		a, ok := r.Access(name)
		panel, pok := r.AccessByRole(ui.RoleRegion, "Menu preview page")
		if !ok || !pok {
			t.Fatalf("missing Gallery target/viewport: %q", name)
		}
		if !a.Offscreen && a.Bounds.H > 0 && a.Bounds.Y >= panel.Bounds.Y && a.Bounds.Y+a.Bounds.H <= panel.Bounds.Y+panel.Bounds.H {
			return a.Bounds
		}
		r.Scroll(panel.Bounds.X+panel.Bounds.W-8, panel.Bounds.Y+panel.Bounds.H/2, 0, a.Bounds.Y+a.Bounds.H/2-panel.Bounds.Y-panel.Bounds.H/2)
	}
	t.Fatalf("Gallery could not reveal %q", name)
	return ui.Rect{}
}

func (d *menuMotionDriver) open(t *testing.T, kind string) *menuMotionSurface {
	t.Helper()
	switch kind {
	case "project":
		menuMotionReveal(t, d.parent, "Project actions")
		advancedClick(t, d.parent, "Project actions")
	case "context":
		b := menuMotionReveal(t, d.parent, "Project workspace")
		d.parent.ClickAt(b.X+b.W/2, b.Y+b.H/2)
		d.parent.Key(ui.Shift, ui.KeyF10)
	case "menubar":
		menuMotionReveal(t, d.parent, "File")
		advancedClick(t, d.parent, "File")
	default:
		t.Fatalf("unknown public preview %q", kind)
	}
	d.flush(t)
	if d.s.Config.NativeMenus {
		return nil
	}
	if len(d.b.surfaces) == 0 {
		t.Fatal("Gallery did not open an independent Flyout")
	}
	return d.b.surfaces[len(d.b.surfaces)-1]
}

func menuMotionKey(t *testing.T, d *menuMotionDriver, s *menuMotionSurface, key ui.Key) {
	t.Helper()
	d.synchronizeClocks()
	s.r.Key(0, key)
	d.flush(t)
}

func menuMotionSnapshot(s *menuMotionSurface) []byte { return append([]byte(nil), s.r.Image().Pix...) }

func TestCreatorMenuMotionUsesObservedPhysicalSideForNativeEntry(t *testing.T) {
	for _, side := range []string{"right", "left", "bottom", "top"} {
		t.Run(side, func(t *testing.T) {
			d := newMenuMotionDriver(1100, false, false, false, false)
			d.b.actualSide = side
			s := d.open(t, "project")
			d.advance(t, 40*time.Millisecond)
			middle := advancedAccess(t, s.r, ui.RoleMenu, "Project actions").Bounds
			d.advance(t, 100*time.Millisecond)
			final := advancedAccess(t, s.r, ui.RoleMenu, "Project actions").Bounds
			if middle.W >= final.W-.1 || middle.H >= final.H-.1 {
				t.Fatalf("native entry did not zoom its actual accessible box: middle=%v final=%v", middle, final)
			}
			movedFromSide := false
			switch side {
			case "right":
				movedFromSide = middle.X < final.X-.1
			case "left":
				movedFromSide = middle.X+middle.W > final.X+final.W+.1
			case "bottom":
				movedFromSide = middle.Y < final.Y-.1
			case "top":
				movedFromSide = middle.Y+middle.H > final.Y+final.H+.1
			}
			if !movedFromSide {
				t.Fatalf("observed %s geometry did not determine entry direction: middle=%v final=%v requested=%s", side, middle, final, s.spec.Placement)
			}
		})
	}
}

func TestCreatorMenuMotionRootsUseRealDefaultClockAndRetainInertExit(t *testing.T) {
	for _, kind := range []string{"project", "context", "menubar"} {
		t.Run(kind, func(t *testing.T) {
			d := newMenuMotionDriver(1100, false, false, false, false)
			s := d.open(t, kind)
			initial := menuMotionSnapshot(s)
			d.advance(t, 40*time.Millisecond)
			middle := menuMotionSnapshot(s)
			d.advance(t, 100*time.Millisecond)
			full := menuMotionSnapshot(s)
			if bytes.Equal(initial, middle) || bytes.Equal(middle, full) {
				t.Fatal("default menu entrance did not paint distinct native frames")
			}
			first := "Copy link"
			if kind == "menubar" {
				first = "New project"
			}
			item := advancedAccess(t, s.r, ui.RoleMenuItem, first).Bounds
			menuMotionKey(t, d, s, ui.KeyEscape)
			if s.didClose || !d.s.Popovers.IsPresent(map[string]string{"project": "navigation-project-actions", "context": "navigation-project-context", "menubar": "preview-menubar"}[kind]) {
				t.Fatal("logical close destroyed the Flyout before its source exit")
			}
			if _, ok := s.r.AccessByRole(ui.RoleMenuItem, first); ok {
				t.Fatal("retained closed menu still exposes actionable accessibility")
			}
			s.r.ClickAt(item.X+item.W/2, item.Y+item.H/2)
			d.flush(t)
			if d.s.pendingMenuAction != "" {
				t.Fatal("a click on exiting menu pixels dispatched an action")
			}
			d.advance(t, 40*time.Millisecond)
			if s.didClose {
				t.Fatal("source exit ended before its visible middle frame")
			}
			if bytes.Equal(full, menuMotionSnapshot(s)) {
				t.Fatal("retained Flyout did not paint its exit")
			}
			d.advance(t, 100*time.Millisecond)
			if !s.didClose {
				t.Fatal("completed menu exit did not release its native surface")
			}
			target := map[string]string{"project": "Project actions", "context": "Project workspace", "menubar": "File"}[kind]
			if !d.parent.Focused(target) {
				t.Fatalf("menu close did not restore the real parent trigger %q", target)
			}
		})
	}
}

func menuMotionVariance(img *image.RGBA, b ui.Rect) int64 {
	box := image.Rect(int(math.Ceil(float64(b.X))), int(math.Ceil(float64(b.Y))), int(math.Floor(float64(b.X+b.W))), int(math.Floor(float64(b.Y+b.H))))
	if !box.In(img.Bounds()) {
		return -1
	}
	var hist [3][256]int64
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			p := img.RGBAAt(x, y)
			hist[0][p.R]++
			hist[1][p.G]++
			hist[2][p.B]++
		}
	}
	var result int64
	for ch := range hist {
		var total int64
		for _, n := range hist[ch] {
			total += n
		}
		var sum int64
		median := 0
		for value, n := range hist[ch] {
			sum += n
			if sum*2 >= total {
				median = value
				break
			}
		}
		for value, n := range hist[ch] {
			delta := value - median
			if delta < 0 {
				delta = -delta
			}
			result += int64(delta) * n
		}
	}
	return result
}

func TestCreatorMenuMotionSubmenuExitReversesAndDoesNotDispatchFromRetainedPixels(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprint("rtl=", rtl), func(t *testing.T) {
			d := newMenuMotionDriver(1100, false, false, false, rtl)
			s := d.open(t, "project")
			d.advance(t, 140*time.Millisecond)
			openKey, closeKey := ui.KeyRight, ui.KeyLeft
			if rtl {
				openKey, closeKey = closeKey, openKey
			}
			menuMotionKey(t, d, s, ui.KeyHome)
			menuMotionKey(t, d, s, ui.KeyDown)
			menuMotionKey(t, d, s, ui.KeyDown)
			menuMotionKey(t, d, s, openKey)
			d.advance(t, 140*time.Millisecond)
			panel := advancedAccess(t, s.r, ui.RoleMenu, "Export").Bounds
			leaf := advancedAccess(t, s.r, ui.RoleMenuItem, "Go source").Bounds
			width := s.r.Image().Bounds().Dx()
			menuMotionKey(t, d, s, closeKey)
			if _, ok := s.r.AccessByRole(ui.RoleMenuItem, "Go source"); ok {
				t.Fatal("closing submenu remains in the actionable tree")
			}
			d.advance(t, 40*time.Millisecond)
			closing := menuMotionVariance(s.r.Image(), panel)
			if closing <= 0 || s.r.Image().Bounds().Dx() != width {
				t.Fatalf("submenu pixels or their native window disappeared during retained exit: variance=%d width=%d→%d panel=%v", closing, width, s.r.Image().Bounds().Dx(), panel)
			}
			menuMotionKey(t, d, s, openKey)
			advancedAccess(t, s.r, ui.RoleMenuItem, "Go source")
			reopened := menuMotionVariance(s.r.Image(), panel)
			if reopened < closing/2 || len(d.b.surfaces) != 1 {
				t.Fatal("reopening reset the visible submenu to a blank entrance or replaced the native window")
			}
			d.advance(t, 140*time.Millisecond)
			menuMotionKey(t, d, s, closeKey)
			s.r.ClickAt(leaf.X+leaf.W/2, leaf.Y+leaf.H/2)
			d.flush(t)
			if d.s.pendingMenuAction != "" {
				t.Fatal("closing submenu pixels executed an action")
			}
			d.advance(t, 140*time.Millisecond)
			menuMotionKey(t, d, s, openKey)
			d.advance(t, 140*time.Millisecond)
			menuMotionKey(t, d, s, ui.KeyEnter)
			if d.s.pendingMenuAction != "go-source" || d.s.navigationMenuAction != "Go source" {
				t.Fatal("real submenu choice did not reach Gallery")
			}
			d.s.pendingMenuAction = ""
			d.advance(t, 200*time.Millisecond)
			d.parent.Frame()
			d.flush(t)
			if d.s.pendingMenuAction != "" {
				t.Fatal("submenu action replayed after its closing animation")
			}
		})
	}
}

func TestCreatorMenuMotionNativeFamiliesKeepOSSubmenusAndReturnOneAction(t *testing.T) {
	for _, kind := range []string{"project", "context", "menubar"} {
		t.Run(kind, func(t *testing.T) {
			d := newMenuMotionDriver(1100, true, false, false, false)
			d.open(t, kind)
			if len(d.b.surfaces) != 0 || !slices.Contains(d.parent.Menu(), "Export") {
				t.Fatalf("native route did not use its OS submenu: surfaces=%d menu=%v", len(d.b.surfaces), d.parent.Menu())
			}
			if err := d.parent.ChooseMenuItem("Export", "Go source"); err != nil {
				t.Fatal(err)
			}
			d.parent.Frame()
			if d.s.pendingMenuAction != "go-source" || d.s.navigationMenuAction != "Go source" || len(d.parent.Menu()) != 0 {
				t.Fatal("OS submenu choice did not reach the app once")
			}
			d.s.pendingMenuAction = ""
			d.advance(t, 200*time.Millisecond)
			if d.s.pendingMenuAction != "" || len(d.b.surfaces) != 0 {
				t.Fatal("native route gained themed motion or replayed its action")
			}
		})
	}
}

func TestCreatorMenuMotionReducedMenusCloseImmediately(t *testing.T) {
	for _, kind := range []string{"project", "menubar"} {
		t.Run(kind, func(t *testing.T) {
			d := newMenuMotionDriver(1100, false, true, false, false)
			s := d.open(t, kind)
			immediate := menuMotionSnapshot(s)
			d.advance(t, 40*time.Millisecond)
			if !bytes.Equal(immediate, menuMotionSnapshot(s)) {
				t.Fatal("reduced menu retained animated entry geometry")
			}
			menuMotionKey(t, d, s, ui.KeyEscape)
			if !s.didClose {
				t.Fatal("reduced menu retained an exit window")
			}
		})
	}
}

func TestCreatorMenuMotionNarrowRTLContextIsKeyboardReachableAndRestoresFocus(t *testing.T) {
	d := newMenuMotionDriver(500, false, true, true, true)
	s := d.open(t, "context")
	menuMotionKey(t, d, s, ui.KeyHome)
	menuMotionKey(t, d, s, ui.KeyDown)
	menuMotionKey(t, d, s, ui.KeyDown)
	menuMotionKey(t, d, s, ui.KeyLeft)
	advancedAccess(t, s.r, ui.RoleMenuItem, "Go source")
	menuMotionKey(t, d, s, ui.KeyEnter)
	if d.s.navigationMenuAction != "Go source" || !d.parent.Focused("Project workspace") {
		t.Fatal("narrow context action or focus restoration failed")
	}
	card := advancedAccess(t, d.parent, ui.RoleGroup, "Project menu preview").Bounds
	for _, label := range []string{"Project actions", "Project workspace"} {
		bounds := menuMotionReveal(t, d.parent, label)
		if bounds.X < card.X || bounds.X+bounds.W > card.X+card.W {
			t.Fatalf("narrow menu target escaped its card: %s %v", label, bounds)
		}
	}
}

func TestCreatorMenuMotionPageRemovalCancelsRetainedExitAndPendingFocus(t *testing.T) {
	d := newMenuMotionDriver(1100, false, false, false, false)
	s := d.open(t, "project")
	d.advance(t, 140*time.Millisecond)
	menuMotionKey(t, d, s, ui.KeyEscape)
	d.s.Page = "overview"
	d.parent.Frame()
	d.flush(t)
	if !s.didClose || d.s.Popovers.IsPresent("navigation-project-actions") {
		t.Fatal("leaving the page retained a detached exit window")
	}
	advancedClick(t, d.parent, "Menu page anchor")
	d.s.Page = "navigation"
	d.parent.Frame()
	d.advance(t, 200*time.Millisecond)
	if !d.parent.Focused("Menu page anchor") || d.s.pendingMenuAction != "" {
		t.Fatal("an old retained menu stole focus or replayed an action after return")
	}
}
