package desktop

import (
	"bytes"
	"fmt"
	"github.com/egoist/mygo"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestMenuMotionEntryCompletesAtSourceDuration(t *testing.T) {
	b := &presenceBackend{}
	h := NewHostWithBackend(b)
	open, state := true, MenuState{}
	_ = ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		a := control.Button(c, "Open", control.ButtonProps{})
		_, err := Menu(c, h, "menu", a, &open, &state, MenuProps{Items: []overlay.MenuItem{{Key: "one", Label: "One"}}})
		if err != nil {
			panic(err)
		}
	}, 400, 240)
	child := ui.NewTester(b.surfaces[0].view, 288, 200)
	child.Advance(110 * time.Millisecond)
	early := append([]byte(nil), child.Image().Pix...)
	child.Advance(300 * time.Millisecond)
	if !bytes.Equal(early, child.Image().Pix) {
		t.Fatal("menu still entering after the source 100ms duration")
	}
}

func TestMenuMotionSubmenuRetainsExitPixels(t *testing.T) {
	state := MenuState{}
	items := []overlay.MenuItem{{Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "child", Label: "Child"}}}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(12).Children(func() {
			MenuContent(c, &state, items, MenuContentProps{ItemContent: func(c *ui.Context, item overlay.MenuItem) {
				if item.Key == "child" {
					ui.Box(c).Size(16, 16).Background(ui.RGB(255, 0, 0))
				} else {
					ui.Text(c, item.Label)
				}
			}})
		})
	}, 600, 300)
	r.Key(0, ui.KeyRight)
	r.Advance(300 * time.Millisecond)
	child, ok := r.AccessByRole(ui.RoleMenuItem, "Child")
	if !ok {
		t.Fatal("submenu missing")
	}
	x, y := int(child.Bounds.X+12), int(child.Bounds.Y+16)
	before := r.Image().RGBAAt(x, y)
	r.Key(0, ui.KeyLeft)
	if got := r.Image().RGBAAt(x, y); got != before {
		t.Fatalf("submenu disappeared at close t0: %v -> %v", before, got)
	}
	if _, ok := r.AccessByRole(ui.RoleMenuItem, "Child"); ok {
		t.Fatal("exiting submenu remains accessible")
	}
	r.Advance(50 * time.Millisecond)
	if !r.FramePending() {
		t.Fatal("submenu did not request intermediate exit frames")
	}
	if got := r.Image().RGBAAt(x, y); got == before {
		t.Fatal("exit painted no intermediate pose")
	}
	r.Advance(100 * time.Millisecond)
	if r.HasText("Child") {
		t.Fatal("submenu retained after exit deadline")
	}
}

func TestMenuMotionRetainsNativeEditorBufferAcrossReversal(t *testing.T) {
	state := MenuState{}
	draft := "original draft"
	items := []overlay.MenuItem{{Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "draft", Label: "Draft item"}}}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(12).Children(func() {
			MenuContent(c, &state, items, MenuContentProps{ItemContent: func(c *ui.Context, item overlay.MenuItem) {
				if item.Key == "draft" {
					ui.TextInputBase(c, &draft).Label("Nested editor")
				} else {
					ui.Text(c, item.Label)
				}
			}})
		})
	}, 600, 300)
	r.Key(0, ui.KeyRight)
	r.Advance(150 * time.Millisecond)
	if err := r.Click("Nested editor"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("saved draft")
	if draft != "saved draft" {
		t.Fatalf("native editor input intercepted: %q", draft)
	}
	r.Key(ui.Cmd, ui.KeyA)
	// Closing is app state here, so the editor's actual selection is retained.
	state.Path = nil
	state.ActiveDepth = 0
	r.Frame()
	r.Advance(30 * time.Millisecond)
	state.Path = []string{"more"}
	state.ActiveDepth = 1
	r.Frame()
	if err := r.Click("Nested editor"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("replacement")
	if draft != "replacement" {
		t.Fatalf("native editor lost its editable buffer on reversal: %q", draft)
	}
}

type motionNativeHarness struct {
	parent, child     *ui.Tester
	host              *Host
	surface           *presenceSurface
	open              bool
	menu              MenuState
	bar               MenubarState
	chosen, dismissed int
}

func newMotionNativeHarness(t *testing.T, family string) *motionNativeHarness {
	t.Helper()
	h := &motionNativeHarness{open: true, bar: MenubarState{Active: "file", Open: true}}
	b := &presenceBackend{}
	h.host = NewHostWithBackend(b)
	items := []overlay.MenuItem{{Key: "leaf", Label: "Leaf"}}
	content := func(c *ui.Context, item overlay.MenuItem) { ui.Box(c).Size(32, 20).Background(ui.RGB(255, 0, 0)) }
	h.parent = ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.host.BeginFrame()
		defer h.host.EndFrame()
		if family == "bar" {
			_, err := Menubar(c, h.host, "audit", &h.bar, MenubarProps{Items: []overlay.MenubarItem{{Key: "file", Label: "File", Items: items}}, ItemContent: content, OnChoose: func(string) { h.chosen++ }, OnDismiss: func() { h.dismissed++ }})
			if err != nil {
				panic(err)
			}
		} else {
			anchor := control.Button(c, "Owner", control.ButtonProps{})
			_, err := Menu(c, h.host, "audit", anchor, &h.open, &h.menu, MenuProps{Items: items, ItemContent: content, OnChoose: func(string) { h.chosen++ }, OnDismiss: func() { h.dismissed++ }})
			if err != nil {
				panic(err)
			}
		}
	}, 480, 240)
	if len(b.surfaces) != 1 {
		t.Fatalf("surfaces=%d", len(b.surfaces))
	}
	h.surface = b.surfaces[0]
	h.child = ui.NewTester(h.surface.view, 420, 220)
	return h
}
func (h *motionNativeHarness) reopen(family string) {
	if family == "bar" {
		h.bar.Open = true
	} else {
		h.open = true
	}
	h.parent.Frame()
	h.child.Frame()
}
func motionNativeRed(t *testing.T, r *ui.Tester) (x, y int, p color.RGBA) {
	t.Helper()
	image := r.Image()
	var minX, minY, maxX, maxY int
	minX, minY = image.Bounds().Max.X, image.Bounds().Max.Y
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			c := color.NRGBAModel.Convert(image.RGBAAt(x, y)).(color.NRGBA)
			if c.R > 240 && c.G < 230 && c.B < 230 && int(c.R)-int(c.G) > 25 {
				minX = min(minX, x)
				minY = min(minY, y)
				maxX = max(maxX, x)
				maxY = max(maxY, y)
			}
		}
	}
	if minX > maxX || minY > maxY {
		t.Fatal("no real red marker pixels")
	}
	x, y = (minX+maxX)/2, (minY+maxY)/2
	return x, y, image.RGBAAt(x, y)
}

func TestMenuMotionSourceNativeMidpointAndDeadline(t *testing.T) {
	for _, family := range []string{"menu", "bar"} {
		t.Run(family, func(t *testing.T) {
			h := newMotionNativeHarness(t, family)
			h.child.Advance(50 * time.Millisecond)
			_, _, p := motionNativeRed(t, h.child)
			// CSS ease(.25,.1,.25,1) is 0.802403 at half of a 100ms fade.
			// The native surface stays transparent behind the fading panel.
			if p.A < 201 || p.A > 208 || p.R != p.A || p.G != 0 || p.B != 0 {
				t.Fatalf("50ms source midpoint got %v, want alpha~.8024", p)
			}
			h.child.Advance(60 * time.Millisecond)
			settled := append([]byte(nil), h.child.Image().Pix...)
			h.child.Advance(500 * time.Millisecond)
			if !bytes.Equal(settled, h.child.Image().Pix) || h.child.FramePending() {
				t.Fatal("entry did not settle by source100ms")
			}
		})
	}
}

func TestMenuMotionBaseRetainedExitNativeInputAndReversal(t *testing.T) {
	for _, family := range []string{"menu", "bar"} {
		t.Run(family, func(t *testing.T) {
			h := newMotionNativeHarness(t, family)
			h.child.Advance(200 * time.Millisecond)
			x, y, before := motionNativeRed(t, h.child)
			h.surface.dismiss()
			h.parent.Frame()
			h.child.Frame()
			if h.surface.closes != 0 || h.dismissed != 0 {
				t.Fatalf("Base exit not retained: close%d dismiss%d", h.surface.closes, h.dismissed)
			}
			if p := h.child.Image().RGBAAt(x, y); p != before {
				t.Fatalf("close t0 did not preserve painted pose: %v -> %v", before, p)
			}
			h.child.Key(0, ui.KeyEnter)
			h.child.ClickAt(float32(x), float32(y))
			h.surface.dismiss()
			if h.chosen != 0 || h.dismissed != 0 {
				t.Fatal("exiting control activated or delivered early/repeated native dismissal")
			}
			h.child.Advance(30 * time.Millisecond)
			h.child.Frame()
			middle := h.child.Image().RGBAAt(x, y)
			if middle == before || h.surface.closes != 0 {
				t.Fatal("exit has no real intermediate pixels")
			}
			h.reopen(family)
			if after := h.child.Image().RGBAAt(x, y); after != middle {
				t.Fatalf("reverse discontinuity: %v -> %v", middle, after)
			}
			h.child.Advance(150 * time.Millisecond)
			if h.surface.closes != 0 || h.dismissed != 0 {
				t.Fatal("cancelled dismissal destroyed surface or callback")
			}
			if _, ok := h.child.AccessByRole(ui.RoleMenuItem, "Leaf"); !ok {
				t.Fatal("reverse did not restore real native item")
			}
			h.child.Key(0, ui.KeyEnter)
			h.parent.Frame()
			h.child.Frame()
			h.child.Advance(110 * time.Millisecond)
			h.parent.Frame()
			if h.chosen != 1 || h.dismissed != 0 || h.surface.closes != 1 {
				t.Fatalf("choose/close policy got chosen%d dismiss%d close%d", h.chosen, h.dismissed, h.surface.closes)
			}
			h.child.Advance(time.Second)
			if h.child.FramePending() {
				t.Fatal("closed menu retained animation work")
			}
		})
	}
}

func TestMenuMotionRunningHiddenOrReducedHasNoIdleLoop(t *testing.T) {
	for _, family := range []string{"menu", "bar"} {
		for _, mode := range []string{"hidden", "reduced"} {
			t.Run(family+"/"+mode, func(t *testing.T) {
				h := newMotionNativeHarness(t, family)
				h.child.Advance(25 * time.Millisecond)
				if mode == "hidden" {
					h.child.SetWindowHidden(true)
				} else {
					h.child.SetPreferences(ui.Preferences{ReduceMotion: true})
				}
				h.child.Frame()
				h.child.Advance(time.Second)
				if h.child.FramePending() {
					t.Fatal("running entry kept requesting frames after native preference")
				}
				_, _, p := motionNativeRed(t, h.child)
				if p.G != 0 || p.B != 0 {
					t.Fatalf("running entry did not finish: %v", p)
				}
				h.surface.dismiss()
				h.parent.Frame()
				h.child.Frame()
				h.parent.Frame()
				if h.surface.closes != 1 || h.dismissed != 1 {
					t.Fatalf("native preference exit got close%d dismiss%d", h.surface.closes, h.dismissed)
				}
				h.child.Advance(time.Second)
				if h.child.FramePending() {
					t.Fatal("instant exit retained frames")
				}
			})
		}
	}
}

type motionNativeGeometryBackend struct {
	presenceBackend
	last            *motionNativeGeometrySurface
	anchor, content ui.Rect
}
type motionNativeGeometrySurface struct {
	*presenceSurface
	unknown         bool
	anchor, content ui.Rect
}

func (s *motionNativeGeometrySurface) Geometry() (ui.Rect, ui.Rect, bool) {
	return s.anchor, s.content, !s.unknown
}
func (b *motionNativeGeometryBackend) Open(spec Spec, view func(*ui.Context), closed func()) (Surface, error) {
	f := &presenceSurface{spec: spec, view: view, onClosed: closed}
	b.surfaces = append(b.surfaces, f)
	b.last = &motionNativeGeometrySurface{presenceSurface: f, anchor: b.anchor, content: b.content}
	return b.last, nil
}
func motionNativeRedCenter(t *testing.T, r *ui.Tester) (float64, float64) {
	t.Helper()
	im := r.Image()
	minX, minY, maxX, maxY := im.Bounds().Max.X, im.Bounds().Max.Y, -1, -1
	for y := 0; y < im.Bounds().Max.Y; y++ {
		for x := 0; x < im.Bounds().Max.X; x++ {
			c := color.NRGBAModel.Convert(im.RGBAAt(x, y)).(color.NRGBA)
			if c.R > 240 && c.G < 230 && c.B < 230 && int(c.R)-int(c.G) > 25 {
				minX = min(minX, x)
				minY = min(minY, y)
				maxX = max(maxX, x)
				maxY = max(maxY, y)
			}
		}
	}
	if maxX < minX || maxY < minY {
		t.Fatal("no real marker pixels for geometry")
	}
	return float64(minX+maxX) / 2, float64(minY+maxY) / 2
}
func TestMenuMotionBaseObservedGeometryOriginPixels(t *testing.T) {
	tests := []struct {
		name                             string
		anchor, content                  ui.Rect
		originX, originY, slideX, slideY float64
	}{
		{"same-width-anchor-center", ui.Rect{X: 100, Y: 100, W: 192, H: 30}, ui.Rect{X: 100, Y: 170, W: 192, H: 40}, 96, -40, 0, -8},
		{"actual-top-opposes-requested-bottom", ui.Rect{X: 100, Y: 100, W: 192, H: 30}, ui.Rect{X: 100, Y: 20, W: 192, H: 40}, 96, 80, 0, 8},
		{"crossaxis-slide-still-points-at-anchor", ui.Rect{X: 100, Y: 100, W: 80, H: 30}, ui.Rect{X: 120, Y: 170, W: 192, H: 40}, 20, -40, 0, -8},
		{"actual-right-opposes-requested-bottom", ui.Rect{X: 100, Y: 100, W: 80, H: 30}, ui.Rect{X: 220, Y: 95, W: 192, H: 42}, -40, 20, -8, 0},
		{"actual-left-opposes-requested-bottom", ui.Rect{X: 100, Y: 100, W: 80, H: 30}, ui.Rect{X: -132, Y: 95, W: 192, H: 42}, 232, 20, 8, 0},
		{"diagonal-is-unknown", ui.Rect{X: 100, Y: 100, W: 80, H: 30}, ui.Rect{X: 300, Y: 170, W: 192, H: 42}, 96, 21, 0, 0},
		{"overlap-is-unknown", ui.Rect{X: 100, Y: 100, W: 80, H: 30}, ui.Rect{X: 110, Y: 110, W: 192, H: 42}, 96, 21, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &motionNativeGeometryBackend{anchor: tt.anchor, content: tt.content}
			h := NewHostWithBackend(b)
			open, state := true, MenuState{}
			_ = ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				anchor := control.Button(c, "Owner", control.ButtonProps{})
				_, err := Menu(c, h, "geometry", anchor, &open, &state, MenuProps{Gap: 40, Placement: mygo.PlacementBottomStart, Items: []overlay.MenuItem{{Key: "leaf", Label: "Leaf"}}, ItemContent: func(c *ui.Context, _ overlay.MenuItem) { ui.Box(c).Size(32, 20).Background(ui.RGB(255, 0, 0)) }})
				if err != nil {
					panic(err)
				}
			}, 480, 240)
			r := ui.NewTester(b.surfaces[0].view, 192, 80)
			r.Advance(25 * time.Millisecond)
			r.Frame()
			earlyX, earlyY := motionNativeRedCenter(t, r)
			r.Advance(200 * time.Millisecond)
			endX, endY := motionNativeRedCenter(t, r)
			const f = .40851059135539586
			const scale = .95 + .05*f
			wantX := tt.originX + (endX-tt.originX)*scale + tt.slideX*(1-f)
			wantY := tt.originY + (endY-tt.originY)*scale + tt.slideY*(1-f)
			if delta := earlyX - wantX; delta < -.8 || delta > .8 {
				t.Fatalf("real X origin got %.2f, want %.2f (end %.2f)", earlyX, wantX, endX)
			}
			if delta := earlyY - wantY; delta < -.8 || delta > .8 {
				t.Fatalf("real Y side/origin got %.2f, want %.2f (end %.2f)", earlyY, wantY, endY)
			}
		})
	}
}

type menuMotionLife struct{ closed *int }

func (p *menuMotionLife) Close() error { *p.closed++; return nil }

func TestMenuMotionSubmenuBranchRetargetRetainsKeyedNativeState(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprintf("rtl=%v", rtl), func(t *testing.T) {
			state := MenuState{}
			a, b := "A draft", "B draft"
			opened, closed := [2]int{}, [2]int{}
			var handles [2]ui.Handle
			var extent ui.Rect
			items := []overlay.MenuItem{{Key: "a", Label: "A", Children: []overlay.MenuItem{{Key: "edit", Label: "A editor"}}}, {Key: "b", Label: "B", Children: []overlay.MenuItem{{Key: "edit", Label: "B editor"}}}}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				row := ui.Row(c).Padding(12)
				if rtl {
					row.Direction(ui.RightToLeft)
				}
				row.Children(func() {
					MenuContent(c, &state, items, MenuContentProps{ItemContent: func(c *ui.Context, item overlay.MenuItem) {
						i := -1
						switch item.Label {
						case "A editor":
							i = 0
						case "B editor":
							i = 1
						}
						if i >= 0 {
							value := &a
							if i == 1 {
								value = &b
							}
							e := ui.TextInputBase(c, value).Label(item.Label).Bind(&handles[i])
							ui.Local(e, "life", func() *menuMotionLife { opened[i]++; return &menuMotionLife{&closed[i]} })
						} else {
							ui.Text(c, item.Label)
						}
					}}).Draw(func(_ *ui.Painter, r ui.Rect) { extent = r })
				})
			}, 700, 300)
			forward := ui.KeyRight
			if rtl {
				forward = ui.KeyLeft
			}
			r.Key(0, forward)
			r.Advance(150 * time.Millisecond)
			if err := r.Click("A editor"); err != nil {
				t.Fatal(err)
			}
			r.Key(ui.Cmd, ui.KeyA)
			r.Type("saved A")
			r.Key(ui.Cmd, ui.KeyA)
			beforeW := extent.W
			if err := r.Click("B"); err != nil {
				t.Fatal(err)
			}
			r.Advance(30 * time.Millisecond)
			r.Frame()
			if state.ActiveDepth != 1 || len(state.Path) != 1 || state.Path[0] != "b" {
				t.Fatalf("native branch switch=%+v", state)
			}
			if _, ok := r.AccessByRole(ui.RoleTextField, "A editor"); ok {
				t.Fatal("leaving A editor remains in AX")
			}
			if extent.W != beforeW {
				t.Fatalf("retarget widened the measured menu: %v -> %v", beforeW, extent.W)
			}
			if err := r.Click("B editor"); err != nil {
				t.Fatal(err)
			}
			r.Key(ui.Cmd, ui.KeyA)
			r.Type("saved B")
			if b != "saved B" || a != "saved A" {
				t.Fatalf("sibling editor crossed values: A%q B%q", a, b)
			}
			if err := r.Click("A"); err != nil {
				t.Fatal(err)
			}
			if opened[0] != 1 || closed[0] != 0 {
				t.Fatalf("same-path reversal remounted A: open%v close%v", opened, closed)
			}
			handles[0].Focus()
			r.Frame()
			r.Type("replacement")
			if a != "replacement" {
				t.Fatalf("native selection did not survive reversal: %q", a)
			}
			r.Advance(150 * time.Millisecond)
			if closed != [2]int{0, 1} {
				t.Fatalf("only retired branch should unmount: %v", closed)
			}
			state.Path = nil
			state.ActiveDepth = 0
			r.Frame()
			r.Advance(150 * time.Millisecond)
			if closed != [2]int{1, 1} || r.FramePending() {
				t.Fatalf("submenu cleanup open%v close%v pending%v", opened, closed, r.FramePending())
			}
		})
	}
}

func TestMenuMotionMenubarTopIdentityDoesNotReuseSiblingSubmenu(t *testing.T) {
	backend := &presenceBackend{}
	host := NewHostWithBackend(backend)
	state := MenubarState{Open: true, Active: "file"}
	draft := [2]string{"File draft", "Edit draft"}
	opened, closed := [2]int{}, [2]int{}
	menus := []overlay.MenubarItem{
		{Key: "file", Label: "File", Items: []overlay.MenuItem{{Key: "same", Label: "File more", Children: []overlay.MenuItem{{Key: "edit", Label: "File editor"}}}}},
		{Key: "edit", Label: "Edit", Items: []overlay.MenuItem{{Key: "same", Label: "Edit more", Children: []overlay.MenuItem{{Key: "edit", Label: "Edit editor"}}}}},
	}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		host.BeginFrame()
		defer host.EndFrame()
		_, err := Menubar(c, host, "bar", &state, MenubarProps{Items: menus, ItemContent: func(c *ui.Context, item overlay.MenuItem) {
			i := -1
			if item.Label == "File editor" {
				i = 0
			} else if item.Label == "Edit editor" {
				i = 1
			}
			if i >= 0 {
				e := ui.TextInputBase(c, &draft[i]).Label(item.Label)
				ui.Local(e, "life", func() *menuMotionLife { opened[i]++; return &menuMotionLife{&closed[i]} })
			} else {
				ui.Text(c, item.Label)
			}
		}})
		if err != nil {
			panic(err)
		}
	}, 500, 240)
	child := ui.NewTester(backend.surfaces[0].view, 500, 300)
	child.Advance(150 * time.Millisecond)
	child.Key(0, ui.KeyRight)
	child.Advance(150 * time.Millisecond)
	child.Click("File editor")
	child.Key(ui.Cmd, ui.KeyA)
	child.Type("Saved file")
	if err := parent.Click("Edit"); err != nil {
		t.Fatal(err)
	}
	child.Frame()
	if child.HasText("File editor") || closed[0] != 1 {
		t.Fatalf("retarget retained another top-level lifecycle: opened%v closed%v", opened, closed)
	}
	child.Key(0, ui.KeyRight)
	child.Advance(150 * time.Millisecond)
	if err := child.Click("Edit editor"); err != nil {
		t.Fatalf("%v active=%s path=%v depth=%d text=%v", err, state.Active, state.menu.Path, state.menu.ActiveDepth, child.Texts())
	}
	child.Key(ui.Cmd, ui.KeyA)
	child.Type("Saved edit")
	if draft != [2]string{"Saved file", "Saved edit"} {
		t.Fatalf("top-level editor values crossed: %v", draft)
	}
	if err := parent.Click("File"); err != nil {
		t.Fatal(err)
	}
	child.Frame()
	child.Key(0, ui.KeyRight)
	child.Advance(150 * time.Millisecond)
	if opened != [2]int{2, 1} || closed != [2]int{1, 1} {
		t.Fatalf("top-level remount policy crossed scope: open%v close%v", opened, closed)
	}
	if child.HasText("Edit editor") {
		t.Fatal("retired sibling leaked into File")
	}
}

func TestMenuMotionRemovedAndReorderedPathsUseCurrentItems(t *testing.T) {
	state := MenuState{}
	draft := "Saved"
	opened, closed, chosen := 0, 0, 0
	items := []overlay.MenuItem{{Key: "a", Label: "A", Children: []overlay.MenuItem{{Key: "editor", Label: "Current editor"}}}, {Key: "b", Label: "B"}}
	p := MenuContentProps{OnChoose: func(string) { chosen++ }, ItemContent: func(c *ui.Context, item overlay.MenuItem) {
		if item.Key == "editor" {
			e := ui.TextInputBase(c, &draft).Label(item.Label)
			ui.Local(e, "life", func() *menuMotionLife { opened++; return &menuMotionLife{&closed} })
		} else {
			ui.Text(c, item.Label)
		}
	}}
	r := ui.NewTester(func(c *ui.Context) { theme.Apply(c); MenuContent(c, &state, items, p) }, 600, 300)
	r.Key(0, ui.KeyRight)
	r.Advance(150 * time.Millisecond)
	r.Click("Current editor")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("native draft")
	items[0], items[1] = items[1], items[0]
	r.Frame()
	if opened != 1 || closed != 0 || draft != "native draft" {
		t.Fatal("reordering remounted the keyed branch")
	}
	state.Path = nil
	state.ActiveDepth = 0
	r.Frame()
	r.Advance(30 * time.Millisecond)
	r.Frame()
	items[1].Children[0].Label = "Updated editor"
	r.Frame()
	if opened != 1 || closed != 0 {
		t.Fatal("leaving path did not use current keyed content")
	}
	items = items[:1]
	r.Frame()
	if closed != 1 || chosen != 0 || r.HasText("Updated editor") {
		t.Fatalf("removed ancestor retained old callback/tree: close%d chosen%d", closed, chosen)
	}
	r.Advance(200 * time.Millisecond)
	if r.FramePending() {
		t.Fatal("deleted branch kept animation work")
	}
}

func TestMenuMotionNativeIMEAndExplicitFocusStayInertDuringExit(t *testing.T) {
	state := MenuState{}
	draft := "Saved"
	force := false
	var handle ui.Handle
	items := []overlay.MenuItem{{Key: "more", Label: "More", Children: []overlay.MenuItem{{Key: "edit", Label: "Edit"}}}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Row(c).Children(func() {
			MenuContent(c, &state, items, MenuContentProps{ItemContent: func(c *ui.Context, item overlay.MenuItem) {
				if item.Key == "edit" {
					e := ui.TextInputBase(c, &draft).Label("IME draft").Bind(&handle)
					if force {
						e.Focus()
						ui.TextInputBase(c, &draft).Label("Inserted draft").AutoFocus()
					}
				} else {
					ui.Text(c, item.Label)
				}
			}})
		})
	}, 500, 200)
	r.Key(0, ui.KeyRight)
	r.Advance(150 * time.Millisecond)
	r.Click("IME draft")
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("未", 1)
	before := draft
	state.Path = nil
	state.ActiveDepth = 0
	force = true
	r.Frame()
	handle.Focus()
	r.Frame()
	if r.Focused("IME draft") || r.Focused("Inserted draft") {
		t.Fatal("retained disabled descendant stole focus")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "IME draft"); ok {
		t.Fatal("retained IME editor is exposed to AX")
	}
	r.Type("late input")
	if draft != before {
		t.Fatalf("leaving editor accepted input or committed old IME: %q -> %q", before, draft)
	}
	r.Advance(150 * time.Millisecond)
	if r.FramePending() {
		t.Fatal("retired native editor kept frames")
	}
}

func TestMenuMotionEntryUsesLatestObservedSideAfterInitialLayout(t *testing.T) {
	for _, mode := range []string{"initially-unknown", "initial-side-flipped"} {
		t.Run(mode, func(t *testing.T) {
			b := &motionNativeGeometryBackend{anchor: ui.Rect{X: 100, Y: 100, W: 192, H: 30}, content: ui.Rect{X: 100, Y: 170, W: 192, H: 42}}
			h := NewHostWithBackend(b)
			open, state := true, MenuState{}
			_ = ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				h.BeginFrame()
				defer h.EndFrame()
				anchor := control.Button(c, "Owner", control.ButtonProps{})
				_, err := Menu(c, h, "late-geometry", anchor, &open, &state, MenuProps{Gap: 40, Placement: mygo.PlacementBottomStart, Items: []overlay.MenuItem{{Key: "leaf", Label: "Leaf"}}, ItemContent: func(c *ui.Context, _ overlay.MenuItem) { ui.Box(c).Size(32, 20).Background(ui.RGB(255, 0, 0)) }})
				if err != nil {
					panic(err)
				}
			}, 480, 240)
			surface := b.last
			surface.unknown = mode == "initially-unknown"
			r := ui.NewTester(b.surfaces[0].view, 192, 80)
			// Placement arrives after the initial zero-opacity layout, just as a
			// measured native Flyout can resize/reposition after its first frame.
			surface.unknown = false
			wantOrigin, wantSlide := float64(-40), float64(-8)
			if mode == "initial-side-flipped" {
				surface.content.Y = 18
				wantOrigin, wantSlide = 82, 8
			}
			r.Frame()
			r.Advance(25 * time.Millisecond)
			r.Frame()
			_, earlyY := motionNativeRedCenter(t, r)
			r.Advance(200 * time.Millisecond)
			_, endY := motionNativeRedCenter(t, r)
			const f = .40851059135539586
			const scale = .95 + .05*f
			wantY := wantOrigin + (endY-wantOrigin)*scale + wantSlide*(1-f)
			if delta := earlyY - wantY; delta < -.8 || delta > .8 {
				t.Fatalf("latest actual side pixel Y%.2f, source%.2f", earlyY, wantY)
			}
		})
	}
}

func TestMenuMotionMenubarChildSiblingSwitchEntersBeforeParentRebuild(t *testing.T) {
	backend := &presenceBackend{}
	host := NewHostWithBackend(backend)
	state := MenubarState{Open: true, Active: "file"}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		host.BeginFrame()
		defer host.EndFrame()
		_, err := Menubar(c, host, "bar", &state, MenubarProps{Items: []overlay.MenubarItem{
			{Key: "file", Label: "File", Items: []overlay.MenuItem{{Key: "first", Label: "First"}}},
			{Key: "edit", Label: "Edit", Items: []overlay.MenuItem{{Key: "second", Label: "Second"}}},
		}, ItemContent: func(c *ui.Context, _ overlay.MenuItem) { ui.Box(c).Size(32, 20).Background(ui.RGB(255, 0, 0)) }})
		if err != nil {
			panic(err)
		}
	}, 500, 240)
	child := ui.NewTester(backend.surfaces[0].view, 400, 200)
	child.Advance(150 * time.Millisecond)
	before := append([]byte(nil), child.Image().Pix...)
	child.Key(0, ui.KeyRight)
	if state.Active != "edit" || !child.HasText("Second") {
		t.Fatal("child did not switch the current sibling immediately")
	}
	if bytes.Equal(before, child.Image().Pix) {
		t.Fatal("new sibling skipped its entry until a parent build")
	}
	child.Advance(50 * time.Millisecond)
	child.Frame()
	_, _, middle := motionNativeRed(t, child)
	if middle.A < 201 || middle.A > 208 || middle.R != middle.A || middle.G != 0 || middle.B != 0 {
		t.Fatalf("sibling did not use source entry clock: %v", middle)
	}
	parent.Frame()
	child.Frame()
	child.Advance(100 * time.Millisecond)
	if backend.surfaces[0].closes != 0 {
		t.Fatal("sibling motion replaced the native window")
	}
}
