package navigation

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type panelBackend struct {
	surfaces                       []*panelSurface
	invalidates                    int
	applicationInactive, holdFocus bool
}
type panelSurface struct {
	spec                         desktop.Spec
	view                         func(*ui.Context)
	closed                       func()
	tester                       *ui.Tester
	sizes                        [][2]int
	focuses, closes, invalidates int
	owner                        *panelBackend
	focused                      bool
}

func (b *panelBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	s := &panelSurface{spec: spec, view: view, closed: closed, owner: b}
	b.surfaces = append(b.surfaces, s)
	return s, nil
}
func (b *panelBackend) InvalidateParent()   { b.invalidates++ }
func (s *panelSurface) SetAnchor(r ui.Rect) { s.spec.Anchor = r }
func (s *panelSurface) SetSize(w, h int) {
	s.spec.Width, s.spec.Height = w, h
	s.sizes = append(s.sizes, [2]int{w, h})
	if s.tester != nil {
		s.tester.SetSize(w, h)
	}
}
func (s *panelSurface) Invalidate() {
	s.invalidates++
	if s.tester != nil {
		s.tester.Frame()
	}
}
func (s *panelSurface) Focus() {
	s.focuses++
	if !s.owner.holdFocus {
		s.focused = true
	}
}
func (s *panelSurface) IsFocused() bool         { return s.focused }
func (b *panelBackend) ApplicationActive() bool { return !b.applicationInactive }
func (s *panelSurface) Close()                  { s.closes++; s.closed() }
func (s *panelSurface) child() *ui.Tester {
	if s.tester == nil {
		s.tester = ui.NewTester(s.view, s.spec.Width, s.spec.Height)
		s.tester.Advance(500 * time.Millisecond)
	}
	return s.tester
}
func panelItems() []NavigationMenuItem {
	return []NavigationMenuItem{
		{Key: "start", Label: "Getting started", Width: 320, Height: 180, Links: []NavItem{{Key: "intro", Label: "Introduction", Description: "Reusable native components."}, {Key: "install", Label: "Installation", Description: "Install dependencies and structure your app."}}},
		{Key: "off", Label: "Unavailable", Disabled: true, Links: []NavItem{{Key: "hidden", Label: "Forbidden route"}}},
		{Key: "components", Label: "Components", Width: 520, Height: 260, Columns: 2, Links: []NavItem{{Key: "alert", Label: "Alert Dialog", Description: "Interrupt with an important response."}, {Key: "tabs", Label: "Tabs", Description: "Layered native content."}, {Key: "off", Label: "Disabled link", Disabled: true}}},
		{Key: "docs", Label: "Documentation"},
	}
}
func panelParent(t *testing.T, b *panelBackend, h *desktop.Host, selected *string, p *NavigationMenuProps, changes *int, visible *bool) *ui.Tester {
	t.Helper()
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Padding(16).Gap(10).Children(func() {
			if visible == nil || *visible {
				e, err := NavigationMenuWith(c, h, "nav", selected, *p)
				if err != nil {
					t.Fatal(err)
				}
				if e.Changed() {
					*changes++
				}
			}
			ui.Button(c, "After navigation")
		})
	}, 900, 480)
}
func panelMove(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	b, ok := r.Find(label)
	if !ok {
		t.Fatalf("missing %s: %v", label, r.Texts())
	}
	r.Move(b.X+b.W/2, b.Y+b.H/2)
}

type animatedPanelBackend struct{ panelBackend }
type animatedPanelSurface struct {
	*panelSurface
	dismiss func() bool
}

func (b *animatedPanelBackend) Open(spec desktop.Spec, view func(*ui.Context), closed func()) (desktop.Surface, error) {
	s, err := b.panelBackend.Open(spec, view, closed)
	if err != nil {
		return nil, err
	}
	return &animatedPanelSurface{panelSurface: s.(*panelSurface)}, nil
}
func (s *animatedPanelSurface) SetDismissHandler(handler func() bool) { s.dismiss = handler }

func TestNavigationPanelSurfaceFadesWithoutBackingOrTeardownFlash(t *testing.T) {
	b := &animatedPanelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := "intro"
	changes := 0
	p := NavigationMenuProps{Items: panelItems()}
	parent := panelParent(t, &b.panelBackend, h, &route, &p, &changes, nil)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyDown)
	if len(b.surfaces) != 1 {
		t.Fatalf("opened %d navigation surfaces", len(b.surfaces))
	}
	surface := b.surfaces[0]
	child := ui.NewTester(surface.view, surface.spec.Width, surface.spec.Height)
	alpha := func() uint64 {
		var total uint64
		for i := 3; i < len(child.Image().Pix); i += 4 {
			total += uint64(child.Image().Pix[i])
		}
		return total
	}
	if got := alpha(); got != 0 {
		t.Fatalf("navigation entry painted a backing before its content: alpha=%d", got)
	}
	child.Advance(500 * time.Millisecond)
	shown := alpha()
	h.Close("nav")
	child.Frame()
	child.Advance(35 * time.Millisecond)
	if leaving := alpha(); leaving == 0 || leaving >= shown || surface.closes != 0 {
		t.Fatalf("navigation surface did not fade: leaving=%d shown=%d closes=%d", leaving, shown, surface.closes)
	}
	child.Advance(150 * time.Millisecond)
	child.Frame()
	if got := alpha(); got != 0 || child.FramePending() || surface.closes != 1 {
		t.Fatalf("navigation teardown flashed or kept rendering: alpha=%d closes=%d pending=%v", got, surface.closes, child.FramePending())
	}
}

func TestNavigationPanelsIndependentContextChooseOnceAndCurrentRoute(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := "intro"
	changes, calls := 0, 0
	p := NavigationMenuProps{Items: panelItems(), OnSelect: func(string) { calls++ }}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyDown)
	if len(b.surfaces) != 1 || !b.surfaces[0].spec.ShowInactive || b.surfaces[0].spec.SystemPopover || b.surfaces[0].focuses != 1 {
		t.Fatalf("independent styled focusable flyout not opened: %+v", b.surfaces)
	}
	child := b.surfaces[0].child()
	if parent.HasText("Reusable native components.") || !child.HasText("Reusable native components.") {
		t.Fatal("panel content context was not independent")
	}
	if !child.Focused("Introduction") {
		t.Fatalf("Down did not enter first native link: %+v", child.AccessAll())
	}
	current, ok := child.AccessByRole(ui.RoleLink, "Introduction")
	if !ok || current.Current != ui.CurrentPage || current.Description != "Reusable native components." {
		t.Fatalf("link metadata=%+v", current)
	}
	child.Key(0, ui.KeyDown)
	child.Key(0, ui.KeyEnter)
	parent.Frame()
	if route != "install" || changes != 1 || calls != 1 || h.IsOpen("nav") {
		t.Fatalf("route=%s changes=%d calls=%d open=%v", route, changes, calls, h.IsOpen("nav"))
	}
	parent.Frame()
	if changes != 1 || calls != 1 {
		t.Fatal("child action repeated")
	}
	if !parent.Focused("Getting started") {
		t.Fatal("choosing did not restore current parent trigger")
	}
}

func TestNavigationPanelHoverIntentGapCancellationAndEscapeSuppression(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Items: panelItems()}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	panelMove(t, parent, "Getting started")
	parent.Advance(199 * time.Millisecond)
	if len(b.surfaces) != 0 {
		t.Fatal("hover opened before intent delay")
	}
	parent.Move(850, 450)
	parent.Advance(time.Second)
	if len(b.surfaces) != 0 {
		t.Fatal("cancelled hover opened later")
	}
	panelMove(t, parent, "Getting started")
	parent.Advance(200 * time.Millisecond)
	parent.Frame()
	if len(b.surfaces) != 1 {
		t.Fatalf("surfaces=%d", len(b.surfaces))
	}
	child := b.surfaces[0].child()
	child.SetWindowActive(false)
	if parent.Focused("Getting started") || b.surfaces[0].focuses != 0 {
		t.Fatal("pointer opening stole keyboard focus")
	}
	parent.Move(850, 450)
	parent.Advance(150 * time.Millisecond)
	panelMove(t, child, "Introduction")
	parent.Frame()
	parent.Advance(time.Second)
	if !h.IsOpen("nav") {
		t.Fatal("cross-window pointer gap closed hovered child")
	}
	panelMove(t, parent, "Getting started")
	child.Key(0, ui.KeyEscape)
	parent.Frame()
	if h.IsOpen("nav") {
		t.Fatal("Escape did not close hover panel")
	}
	panelMove(t, parent, "Getting started")
	parent.Advance(time.Second)
	if h.IsOpen("nav") {
		t.Fatal("Escape immediately reopened same hovered trigger")
	}
	parent.Move(850, 450)
	panelMove(t, parent, "Getting started")
	parent.Advance(201 * time.Millisecond)
	if !h.IsOpen("nav") {
		t.Fatal("new pointer entry stayed suppressed")
	}
}

func TestNavigationPanelSiblingSwitchResizesLiveSurfaceAndPreservesContentState(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	created := 0
	p := NavigationMenuProps{Items: panelItems()}
	p.Items[0].Content = func(c *ui.Context, choose func(string)) {
		root := ui.Column(c)
		count := ui.Local(root, "counter", func() int { created++; return 0 })
		root.Children(func() {
			if ui.Button(c, fmt.Sprintf("Counter %d", *count)).Clicked() {
				*count++
			}
			ui.Button(c, "Navigate custom").OnClick(func() { choose("custom") })
		})
	}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	if err := parent.Click("Getting started"); err != nil {
		t.Fatal(err)
	}
	surface := b.surfaces[0]
	child := surface.child()
	baseline := created
	if err := child.Click("Counter 0"); err != nil {
		t.Fatal(err)
	}
	parent.Frame()
	panelMove(t, parent, "Components")
	parent.Frame()
	child.Frame()
	parent.Advance(100 * time.Millisecond)
	if len(b.surfaces) != 1 || surface.spec.Width <= 320 || surface.spec.Width >= 520 || surface.spec.Height <= 180 || surface.spec.Height >= 260 {
		t.Fatalf("native viewport did not animate in place: surfaces=%d spec=%+v sizes=%v", len(b.surfaces), surface.spec, surface.sizes)
	}
	parent.Advance(500 * time.Millisecond)
	child.Advance(500 * time.Millisecond)
	if surface.spec.Width != 520 || surface.spec.Height != 260 || len(surface.sizes) == 0 {
		t.Fatalf("viewport dimensions=%+v", surface.spec)
	}
	if _, ok := child.AccessByRole(ui.RoleButton, "Counter 1"); ok {
		t.Fatal("inactive custom control remained accessible")
	}
	panelMove(t, parent, "Getting started")
	parent.Advance(500 * time.Millisecond)
	child.Advance(500 * time.Millisecond)
	if created != baseline || !child.HasText("Counter 1") {
		t.Fatalf("content remounted or lost state created=%d baseline=%d texts=%v", created, baseline, child.Texts())
	}
	parent.Move(850, 450)
	child.Move(10, 10)
	parent.Frame()
	parent.Advance(time.Second)
	child.Advance(time.Second)
	if parent.FramePending() || child.FramePending() {
		t.Fatal("settled viewport kept scheduling frames")
	}
}

func TestNavigationPanelKeyboardSiblingRTLAndDynamicItems(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprintf("rtl=%v", rtl), func(t *testing.T) {
			b := &panelBackend{}
			h := desktop.NewHostWithBackend(b)
			route := ""
			changes := 0
			p := NavigationMenuProps{Items: panelItems()}
			parent := panelParent(t, b, h, &route, &p, &changes, nil)
			if rtl {
				parent.SetDirection(ui.RightToLeft)
			}
			parent.Key(0, ui.KeyTab)
			parent.Key(0, ui.KeyDown)
			child := b.surfaces[0].child()
			if rtl {
				child.SetDirection(ui.RightToLeft)
			}
			next := ui.KeyRight
			if rtl {
				next = ui.KeyLeft
			}
			child.Key(0, next)
			parent.Frame()
			child.Frame()
			if !child.Focused("Alert Dialog") {
				t.Fatalf("sibling key did not skip disabled trigger: %+v", child.AccessAll())
			}
			disabled, _ := child.AccessByRole(ui.RoleLink, "Disabled link")
			if !disabled.Disabled {
				t.Fatal("disabled route not exposed as disabled")
			}
			child.Key(0, ui.KeyEscape)
			parent.Frame()
			if !parent.Focused("Components") {
				t.Fatal("Escape restored obsolete trigger")
			}
			parent.Key(0, ui.KeyDown)
			child = b.surfaces[len(b.surfaces)-1].child()
			p.Items = p.Items[:1]
			parent.Frame()
			if h.IsOpen("nav") {
				t.Fatal("removed active item retained child")
			}
		})
	}
}

func TestNavigationPanelArbitraryChildContextAndRevisionRefresh(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	external := "alpha"
	var parentContext *ui.Context
	p := NavigationMenuProps{Items: []NavigationMenuItem{{Key: "custom", Label: "Custom", Content: func(child *ui.Context, choose func(string)) {
		if child == parentContext {
			t.Fatal("parent Context reused")
		}
		ui.Text(child, external)
		if ui.Button(child, "Choose custom").Clicked() {
			choose("made-by-app")
		}
	}}}}
	parent := ui.NewTester(func(c *ui.Context) {
		parentContext = c
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		e, err := NavigationMenuWith(c, h, "custom", &route, p)
		if err != nil {
			t.Fatal(err)
		}
		if e.Changed() {
			changes++
		}
	}, 900, 400)
	parent.Click("Custom")
	child := b.surfaces[0].child()
	if !child.HasText("alpha") {
		t.Fatal("custom content absent")
	}
	external = "beta"
	p.Revision++
	parent.Frame()
	if !child.HasText("beta") {
		t.Fatal("external content revision did not refresh child")
	}
	child.Click("Choose custom")
	parent.Frame()
	if route != "made-by-app" || changes != 1 {
		t.Fatal("custom choose did not publish once")
	}
}

func TestNativeNavigationMenuAlwaysUsesSystemMenus(t *testing.T) {
	route := "intro"
	changes := 0
	parent := ui.NewTester(func(c *ui.Context) {
		c.SetMenuPresentation(ui.PopoverMenus)
		if NativeNavigationMenu(c, &route, []NavItem{{Key: "start", Label: "Get started", Children: []NavItem{{Key: "intro", Label: "Introduction"}, {Key: "install", Label: "Install"}}}, {Key: "docs", Label: "Documentation"}}).Changed() {
			changes++
		}
		if c.MenuPresentation() != ui.PopoverMenus {
			t.Fatal("system policy leaked")
		}
	}, 600, 200)
	parent.Click("Get started")
	if strings.Join(parent.Menu(), ",") != "Introduction,Install" || parent.HasText("Install") {
		t.Fatalf("not a native system menu: menu=%v texts=%v", parent.Menu(), parent.Texts())
	}
	if err := parent.ChooseMenuItem("Install"); err != nil {
		t.Fatal(err)
	}
	if route != "install" || changes != 1 {
		t.Fatalf("route=%s changes=%d", route, changes)
	}
	parent.Click("Documentation")
	if route != "docs" || changes != 2 {
		t.Fatal("direct route did not choose")
	}
}

func TestNavigationPanelTabExitUsesRealParentFocusTarget(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	var before, after ui.Handle
	p := NavigationMenuProps{Items: panelItems(), OnTab: func(backward bool) {
		if backward {
			before.Focus()
		} else {
			after.Focus()
		}
	}}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		ui.Column(c).Children(func() {
			ui.Button(c, "Before").Bind(&before)
			if _, err := NavigationMenuWith(c, h, "nav", &route, p); err != nil {
				t.Fatal(err)
			}
			ui.Button(c, "After").Bind(&after)
		})
	}, 900, 400)
	for _, backward := range []bool{false, true} {
		parent.Click("Getting started")
		parent.Frame()
		child := b.surfaces[len(b.surfaces)-1].child()
		mods := ui.Modifiers(0)
		if backward {
			mods = ui.Shift
		}
		child.Key(mods, ui.KeyTab)
		parent.Frame()
		want := "After"
		if backward {
			want = "Before"
		}
		if h.IsOpen("nav") || !parent.Focused(want) {
			t.Fatalf("Tab exit backward=%v did not focus %s: open=%v parent=%+v child=%+v", backward, want, h.IsOpen("nav"), parent.AccessAll(), child.AccessAll())
		}
	}
}

func TestNavigationPanelMeasuredHeightReducedMotionAndNativeGeometry(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Items: panelItems()}
	p.Items[0].Height = 0
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.SetPreferences(ui.Preferences{ReduceMotion: true})
	parent.Click("Getting started")
	parent.Frame()
	surface := b.surfaces[0]
	child := surface.child()
	parent.Frame()
	child.Frame()
	parent.Frame()
	if surface.spec.Height == 180 || surface.spec.Height < 120 {
		t.Fatalf("content height not measured: %+v", surface.spec)
	}
	oldAnchor := surface.spec.Anchor
	panelMove(t, parent, "Components")
	parent.Frame()
	child.Frame()
	if len(b.surfaces) != 1 || surface.spec.Width != 520 || surface.spec.Height != 260 || surface.spec.Anchor.X <= oldAnchor.X {
		t.Fatalf("reduced motion did not apply native geometry immediately: %+v", surface.spec)
	}
	parent.Move(850, 450)
	parent.Frame()
	child.Frame()
	parent.Frame()
	if parent.FramePending() || child.FramePending() {
		t.Fatal("reduced navigation viewport requested animation frames")
	}
}

func TestNavigationPanelDrawnActiveWithoutPointerOrFocusStillCloses(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Items: panelItems()}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	panelMove(t, parent, "Getting started")
	parent.Advance(201 * time.Millisecond)
	parent.Frame()
	child := b.surfaces[0].child()
	// Native Flyouts can inherit active appearance while shown inactive. A
	// drawn-active window is not evidence of keyboard focus in its content.
	parent.Move(850, 450)
	child.Move(890, 470)
	parent.Frame()
	parent.Advance(301 * time.Millisecond)
	if h.IsOpen("nav") {
		t.Fatal("active appearance kept an abandoned pointer panel open")
	}
	parent.Advance(time.Second)
	child.Advance(time.Second)
	parent.Frame()
	if parent.FramePending() || child.FramePending() {
		t.Fatal("closed hover timer left idle frames")
	}
}

func TestNavigationPanelCustomEditorsKeepTabAndCaretKeys(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	first, second := "alpha", "beta"
	p := NavigationMenuProps{Items: []NavigationMenuItem{{Key: "form", Label: "Form", Content: func(c *ui.Context, choose func(string)) {
		ui.TextInput(c, &first).Label("First name")
		ui.TextInput(c, &second).Label("Last name")
	}}, panelItems()[2]}}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Click("Form")
	parent.Frame()
	child := b.surfaces[0].child()
	child.Key(0, ui.KeyTab)
	if !child.Focused("First name") {
		t.Fatal("Tab did not enter custom editor")
	}
	child.Key(0, ui.KeyRight)
	child.Type("X")
	if !strings.Contains(first, "X") || !h.IsOpen("nav") {
		t.Fatal("sibling navigation stole editor caret key")
	}
	child.Key(0, ui.KeyTab)
	if !child.Focused("Last name") {
		t.Fatal("custom native Tab traversal was replaced by route-group exit")
	}
	child.Key(0, ui.KeyEscape)
	parent.Frame()
	if h.IsOpen("nav") || !parent.Focused("Form") {
		t.Fatal("custom form Escape did not return to trigger")
	}
}

func TestNavigationPanelUpEntersLastLinkAndNoHostDoesNotFakePopup(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Items: panelItems()}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyUp)
	child := b.surfaces[0].child()
	if !child.Focused("Installation") {
		t.Fatal("Up did not enter last enabled route")
	}
	var openError error
	standalone := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		_, err := NavigationMenuWith(c, nil, "standalone", &route, p)
		if err != nil {
			openError = err
		}
	}, 900, 300)
	if openError != nil {
		t.Fatal("closed trigger required Host")
	}
	standalone.Click("Getting started")
	if openError == nil || standalone.HasText("Reusable native components.") {
		t.Fatal("missing Host silently drew parent-window content")
	}
}

func TestNavigationPanelPaintsDirectionalContentAndViewportTransition(t *testing.T) {
	for _, reduce := range []bool{false, true} {
		t.Run(fmt.Sprintf("reduced=%v", reduce), func(t *testing.T) {
			b := &panelBackend{}
			h := desktop.NewHostWithBackend(b)
			route := ""
			changes := 0
			p := NavigationMenuProps{Items: []NavigationMenuItem{{Key: "red", Label: "Red panel", Width: 320, Height: 180, Content: func(c *ui.Context, _ func(string)) { ui.Box(c).FillWidth().Height(150).Background(ui.RGB(220, 0, 0)) }}, {Key: "blue", Label: "Blue panel", Width: 520, Height: 260, Content: func(c *ui.Context, _ func(string)) { ui.Box(c).FillWidth().Height(230).Background(ui.RGB(0, 0, 220)) }}}}
			parent := panelParent(t, b, h, &route, &p, &changes, nil)
			if reduce {
				parent.SetPreferences(ui.Preferences{ReduceMotion: true})
			}
			panelMove(t, parent, "Red panel")
			parent.Advance(201 * time.Millisecond)
			parent.Frame()
			surface := b.surfaces[0]
			child := surface.child()
			parent.Frame()
			oldAnchor := surface.spec.Anchor
			panelMove(t, parent, "Blue panel")
			parent.Frame()
			child.Frame()
			if !reduce {
				parent.Advance(75 * time.Millisecond)
				child.Advance(75 * time.Millisecond)
				if surface.spec.Width <= 320 || surface.spec.Width >= 520 || surface.spec.Anchor.X <= oldAnchor.X {
					t.Fatal("native size/position did not interpolate")
				}
				red, blue := 0, 0
				im := child.Image()
				for y := 15; y < 150; y++ {
					for x := 0; x < im.Bounds().Dx(); x++ {
						pixel := im.RGBAAt(x, y)
						if pixel.R > pixel.B+30 {
							red++
						}
						if pixel.B > pixel.R+30 {
							blue++
						}
					}
				}
				if red < 100 || blue < 100 {
					t.Fatalf("content transition did not paint both panels: red=%d blue=%d", red, blue)
				}
			} else if surface.spec.Width != 520 || surface.spec.Height != 260 {
				t.Fatal("reduced geometry did not apply immediately")
			}
			parent.Advance(500 * time.Millisecond)
			child.Advance(500 * time.Millisecond)
			parent.Frame()
			child.Frame()
			if surface.spec.Width != 520 || surface.spec.Height != 260 || parent.FramePending() || child.FramePending() {
				t.Fatal("viewport transition did not settle")
			}
			blue := child.Image().RGBAAt(30, 30)
			if blue.B < 180 || blue.R > 20 {
				t.Fatalf("destination panel did not settle on its pixels: %v", blue)
			}
		})
	}
}

func TestNavigationPanelDisabledNestedGroupCannotChooseOrReceiveFocus(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Items: []NavigationMenuItem{{Key: "grouped", Label: "Grouped", Links: []NavItem{{Key: "unavailable", Label: "Unavailable group", Disabled: true, Children: []NavItem{{Key: "forbidden", Label: "Forbidden"}}}, {Key: "okay", Label: "Available"}}}}}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Click("Grouped")
	parent.Frame()
	child := b.surfaces[0].child()
	forbidden, ok := child.AccessByRole(ui.RoleLink, "Forbidden")
	if !ok || !forbidden.Disabled || forbidden.Focusable {
		t.Fatalf("disabled ancestor not propagated: %+v", forbidden)
	}
	if !child.Focused("Available") {
		t.Fatal("keyboard entry tried to focus disabled descendant")
	}
	child.Click("Forbidden")
	parent.Frame()
	if route != "" || !h.IsOpen("nav") {
		t.Fatal("disabled descendant navigated")
	}
	child.Click("Available")
	parent.Frame()
	if route != "okay" || changes != 1 {
		t.Fatal("available descendant did not navigate")
	}
}

func TestNavigationPanelCustomTabBoundariesTraverseRealControls(t *testing.T) {
	for _, backward := range []bool{false, true} {
		t.Run(fmt.Sprintf("backward=%v", backward), func(t *testing.T) {
			b := &panelBackend{}
			h := desktop.NewHostWithBackend(b)
			route := ""
			changes, calls := 0, 0
			gotBackward := false
			p := NavigationMenuProps{OnTab: func(back bool) { calls++; gotBackward = back }, Items: []NavigationMenuItem{{Key: "custom", Label: "Custom", Content: func(c *ui.Context, _ func(string)) { ui.Button(c, "First"); ui.Button(c, "Last") }}}}
			parent := panelParent(t, b, h, &route, &p, &changes, nil)
			parent.Click("Custom")
			parent.Frame()
			child := b.surfaces[0].child()
			child.Key(0, ui.KeyTab)
			if !child.Focused("First") || calls != 0 {
				t.Fatal("first internal Tab left custom controls")
			}
			if !backward {
				child.Key(0, ui.KeyTab)
				if !child.Focused("Last") || calls != 0 {
					t.Fatal("last internal Tab left custom controls")
				}
			}
			mods := ui.Modifiers(0)
			if backward {
				mods = ui.Shift
			}
			child.Key(mods, ui.KeyTab)
			parent.Frame()
			if h.IsOpen("nav") || calls != 1 || gotBackward != backward || !parent.Focused("Custom") {
				t.Fatalf("custom boundary did not leave once: backward=%v calls=%d reported=%v", backward, calls, gotBackward)
			}
		})
	}
}

func TestNavigationPanelEmptyCustomRegionAndInitialBackwardTabLeave(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(fmt.Sprintf("empty=%v", empty), func(t *testing.T) {
			b := &panelBackend{}
			h := desktop.NewHostWithBackend(b)
			route := ""
			changes, calls := 0, 0
			p := NavigationMenuProps{OnTab: func(bool) { calls++ }, Items: []NavigationMenuItem{{Key: "custom", Label: "Custom", Content: func(c *ui.Context, _ func(string)) {
				if !empty {
					ui.Button(c, "Real content action")
				}
			}}}}
			parent := panelParent(t, b, h, &route, &p, &changes, nil)
			parent.Click("Custom")
			parent.Frame()
			child := b.surfaces[0].child()
			child.Key(ui.Shift, ui.KeyTab)
			parent.Frame()
			if h.IsOpen("nav") || calls != 1 {
				t.Fatal("initial backward Tab did not leave real panel region")
			}
		})
	}
}

func TestNavigationPanelCustomControlMayConsumeTab(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes, tabs := 0, 0
	p := NavigationMenuProps{Items: []NavigationMenuItem{{Key: "custom", Label: "Custom", Content: func(c *ui.Context, _ func(string)) {
		ui.Box(c).Size(100, 40).Role(ui.RoleTextField).Label("Code editor").Focusable().HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyTab {
				tabs++
				return true
			}
			return false
		})
	}}}}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Click("Custom")
	parent.Frame()
	child := b.surfaces[0].child()
	child.Key(0, ui.KeyTab)
	child.Key(0, ui.KeyTab)
	parent.Frame()
	if tabs != 1 || !h.IsOpen("nav") || !child.Focused("Code editor") {
		t.Fatal("navigation overrode a custom editor's native Tab handler")
	}
}

func TestNavigationPanelMissingHostDoesNotSpinWhileTriggerRemainsHovered(t *testing.T) {
	route := ""
	errors := 0
	p := NavigationMenuProps{Items: panelItems()}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		if _, err := NavigationMenuWith(c, nil, "absent", &route, p); err != nil {
			errors++
		}
	}, 900, 300)
	panelMove(t, parent, "Getting started")
	parent.Advance(201 * time.Millisecond)
	parent.Frame()
	parent.Advance(2 * time.Second)
	if errors != 1 || parent.FramePending() {
		t.Fatalf("missing Host retried while idle: errors=%d pending=%v", errors, parent.FramePending())
	}
}

func TestNavigationPanelCustomChooseIgnoresDuplicateAndInactiveCallbacks(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes, calls := 0, 0
	var firstChoose func(string)
	p := NavigationMenuProps{OnSelect: func(string) { calls++ }, Items: []NavigationMenuItem{{Key: "first", Label: "First panel", Content: func(c *ui.Context, choose func(string)) { firstChoose = choose; ui.Text(c, "First custom") }}, panelItems()[2]}}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Click("First panel")
	parent.Frame()
	child := b.surfaces[0].child()
	panelMove(t, parent, "Components")
	parent.Frame()
	firstChoose("stale")
	parent.Frame()
	if route != "" || calls != 0 {
		t.Fatal("inactive custom callback navigated")
	}
	panelMove(t, parent, "First panel")
	parent.Frame()
	child.Frame()
	firstChoose("one")
	firstChoose("two")
	parent.Frame()
	if route != "one" || calls != 1 || changes != 1 {
		t.Fatalf("duplicate custom callback navigated again: route=%s calls=%d changes=%d", route, calls, changes)
	}
}

func TestNavigationPanelRapidRetargetKeepsCurrentPaintPosition(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Duration: 600 * time.Millisecond, Items: []NavigationMenuItem{{Key: "first", Label: "First panel", Width: 400, Height: 200, Content: func(c *ui.Context, _ func(string)) { ui.Box(c).FillWidth().Height(176).Background(ui.RGB(220, 0, 0)) }}, {Key: "second", Label: "Second panel", Width: 400, Height: 200, Content: func(c *ui.Context, _ func(string)) { ui.Box(c).FillWidth().Height(176).Background(ui.RGB(0, 220, 0)) }}, {Key: "third", Label: "Third panel", Width: 400, Height: 200, Content: func(c *ui.Context, _ func(string)) { ui.Box(c).FillWidth().Height(176).Background(ui.RGB(0, 0, 220)) }}}}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	panelMove(t, parent, "First panel")
	parent.Advance(201 * time.Millisecond)
	parent.Frame()
	child := b.surfaces[0].child()
	parent.Frame()
	panelMove(t, parent, "Second panel")
	parent.Frame()
	child.Frame()
	parent.Advance(100 * time.Millisecond)
	child.Advance(100 * time.Millisecond)
	child.Frame()
	before := append([]byte(nil), child.Image().Pix...)
	panelMove(t, parent, "Third panel")
	parent.Frame()
	child.Frame()
	after := child.Image()
	changed := 0
	for y := 20; y < 160; y++ {
		for x := 20; x < 380; x++ {
			at := after.PixOffset(x, y)
			if string(before[at:at+4]) != string(after.Pix[at:at+4]) {
				changed++
			}
		}
	}
	if changed > 300 {
		t.Fatalf("retarget instant jumped the visible content: changed pixels=%d before=%v/%v/%v after=%v/%v/%v", changed, before[after.PixOffset(50, 50):after.PixOffset(50, 50)+4], before[after.PixOffset(150, 50):after.PixOffset(150, 50)+4], before[after.PixOffset(300, 50):after.PixOffset(300, 50)+4], after.RGBAAt(50, 50), after.RGBAAt(150, 50), after.RGBAAt(300, 50))
	}
	parent.Advance(time.Second)
	child.Advance(time.Second)
	parent.Frame()
	child.Frame()
	if len(b.surfaces) != 1 || parent.FramePending() || child.FramePending() {
		t.Fatal("rapid content retarget recreated or failed to settle")
	}
}

func TestNavigationPanelConfirmedKeyboardFocusDoesNotKeepAbandonedWindowOpen(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Items: panelItems()}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Key(0, ui.KeyTab)
	parent.Key(0, ui.KeyDown)
	surface := b.surfaces[0]
	child := surface.child()
	parent.Frame()
	if !child.Focused("Introduction") {
		t.Fatal("child keyboard entry never completed")
	}
	surface.focused = false
	child.SetFocused(false)
	child.Move(890, 470)
	parent.Click("After navigation")
	parent.Frame()
	parent.Advance(301 * time.Millisecond)
	if h.IsOpen("nav") || !parent.Focused("After navigation") {
		t.Fatal("completed app-active handoff kept abandoned panel or stole new parent focus")
	}
}

func TestNavigationPanelKeyboardPendingBridgesOnlyUntilActualFocusOrAppDeactivation(t *testing.T) {
	for _, completed := range []bool{false, true} {
		t.Run(fmt.Sprintf("completed=%v", completed), func(t *testing.T) {
			b := &panelBackend{holdFocus: true}
			h := desktop.NewHostWithBackend(b)
			route := ""
			changes := 0
			p := NavigationMenuProps{Items: panelItems()}
			parent := panelParent(t, b, h, &route, &p, &changes, nil)
			parent.Key(0, ui.KeyTab)
			parent.Key(0, ui.KeyDown)
			surface := b.surfaces[0]
			child := surface.child()
			child.SetFocused(false)
			parent.SetFocused(false)
			parent.SetWindowActive(false)
			parent.Move(890, 470)
			child.Move(890, 470)
			parent.Frame()
			parent.Advance(time.Second)
			if !h.IsOpen("nav") || parent.FramePending() {
				t.Fatal("known app-active pending handoff closed or spun")
			}
			if completed {
				surface.focused = true
				child.SetFocused(true)
				child.Frame()
				parent.Frame()
				surface.focused = false
				child.SetFocused(false)
				parent.Frame()
				if h.IsOpen("nav") {
					t.Fatal("confirmed handoff continued bridging real focus loss")
				}
			}
			if !completed {
				b.applicationInactive = true
				parent.Frame()
				if h.IsOpen("nav") {
					t.Fatal("application deactivation kept pending keyboard handoff")
				}
			}
		})
	}
}

func TestNavigationPanelDoesNotMountUnvisitedAutofocusContent(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes, mounted := 0, 0
	text := "one"
	p := NavigationMenuProps{Items: []NavigationMenuItem{{Key: "form", Label: "Form", Content: func(c *ui.Context, _ func(string)) { ui.TextInput(c, &text).Label("Existing editor") }}, panelItems()[2]}}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Click("Form")
	parent.Frame()
	child := b.surfaces[0].child()
	child.Key(0, ui.KeyTab)
	p.Items = append(p.Items, NavigationMenuItem{Key: "later", Label: "Later", Content: func(c *ui.Context, _ func(string)) { mounted++; ui.Button(c, "New autofocus").AutoFocus() }})
	parent.Frame()
	child.Frame()
	if mounted != 0 || !child.Focused("Existing editor") {
		t.Fatal("unvisited inactive content ran autofocus")
	}
	child.Type("X")
	if !strings.Contains(text, "X") {
		t.Fatal("inserting inactive content stole editor input")
	}
}

func TestNavigationPanelKeyboardSiblingWinsOverStationaryPointer(t *testing.T) {
	b := &panelBackend{}
	h := desktop.NewHostWithBackend(b)
	route := ""
	changes := 0
	p := NavigationMenuProps{Items: panelItems()}
	parent := panelParent(t, b, h, &route, &p, &changes, nil)
	parent.Click("Getting started")
	parent.Frame()
	child := b.surfaces[0].child()
	child.Key(0, ui.KeyRight)
	parent.Frame()
	child.Frame()
	if _, ok := child.AccessByRole(ui.RoleLink, "Alert Dialog"); !ok {
		t.Fatal("stationary pointer over old trigger reversed keyboard sibling switch")
	}
	child.Key(0, ui.KeyRight)
	parent.Frame()
	if h.IsOpen("nav") || !parent.Focused("Documentation") {
		t.Fatal("stationary pointer reopened old panel while moving to direct route")
	}
	parent.Move(850, 450)
	panelMove(t, parent, "Getting started")
	parent.Advance(201 * time.Millisecond)
	parent.Frame()
	if !h.IsOpen("nav") {
		t.Fatal("real pointer motion did not resume hover intent")
	}
}
