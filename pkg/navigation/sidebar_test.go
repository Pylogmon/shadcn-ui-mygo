package navigation

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func sidebarTestItems() []SidebarItem {
	return []SidebarItem{
		{Key: "home", Label: "Home", Group: "Workspace", Badge: "4"},
		{Key: "projects", Label: "Projects", DefaultOpen: true, Children: []SidebarItem{
			{Key: "current", Label: "Current project"},
			{Key: "off", Label: "Unavailable project", Disabled: true},
		}},
		{Key: "settings", Label: "Settings", Group: "Account", Size: SidebarItemSmall, Outline: true},
	}
}
func sidebarTestAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	info, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q: %+v", role, label, r.AccessAll())
	}
	return info
}
func sidebarTestClick(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	if err := r.Click(label); err != nil {
		t.Fatal(err)
	}
}
func sidebarTestView(t *testing.T, s *SidebarState, p *SidebarLayoutProps, route *string, main func(*ui.Context), reduced bool) (*ui.Tester, *panelBackend, *int) {
	t.Helper()
	b := &panelBackend{}
	host := desktop.NewHostWithBackend(b)
	changes := new(int)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		host.BeginFrame()
		defer host.EndFrame()
		e, err := SidebarWith(c, host, "workspace", route, s, *p, main)
		if err != nil {
			t.Fatal(err)
		}
		if e.Changed() {
			*changes++
		}
	}, 1000, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
	return r, b, changes
}

func TestSidebarNestedDisclosureSelectionAndDisabledRoutes(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	selects, groups := 0, 0
	p := SidebarLayoutProps{Label: "Workspace navigation", Items: sidebarTestItems(), OnSelect: func(string) { selects++ }, OnGroupChange: func(string, bool) { groups++ }}
	r, _, changes := sidebarTestView(t, s, &p, &route, func(c *ui.Context) { ui.Button(c, "Editor action") }, false)
	home := sidebarTestAccess(t, r, ui.RoleLink, "Home")
	if home.Current != ui.CurrentPage {
		t.Fatal("current route has no native current-page state")
	}
	if !sidebarTestAccess(t, r, ui.RoleDisclosure, "Projects").Expanded {
		t.Fatal("default-expanded branch is closed")
	}
	sidebarTestClick(t, r, "Current project")
	if route != "current" || selects != 1 || *changes != 1 {
		t.Fatalf("route=%q selects=%d changes=%d", route, selects, *changes)
	}
	if err := r.Click("Unavailable project"); err != nil && !strings.Contains(err.Error(), "disabled") {
		t.Fatal(err)
	}
	if route != "current" || selects != 1 {
		t.Fatal("disabled child changed the route")
	}
	sidebarTestClick(t, r, "Projects")
	if s.Groups["projects"] || groups != 1 {
		t.Fatal("native disclosure did not update controlled group state")
	}
	r.Advance(300 * time.Millisecond)
	if _, ok := r.AccessByRole(ui.RoleLink, "Current project"); ok {
		t.Fatal("closed nested route is still accessible")
	}
	r.Key(0, ui.KeyRight)
	r.Advance(300 * time.Millisecond)
	if !s.Groups["projects"] || groups != 2 {
		t.Fatal("Right did not expand focused group")
	}
	disabled := p.Items[1]
	disabled.Disabled = true
	p.Items[1] = disabled
	r.Frame()
	child := sidebarTestAccess(t, r, ui.RoleLink, "Current project")
	if !child.Disabled {
		t.Fatal("disabled group did not disable its descendant links")
	}
}

func TestSidebarResponsiveUsesContainerAndPreservesIndependentModes(t *testing.T) {
	s := &SidebarState{Collapsed: true}
	route := "home"
	p := SidebarLayoutProps{Items: sidebarTestItems(), Collapse: SidebarIcon, AvailableWidth: 500}
	r, _, _ := sidebarTestView(t, s, &p, &route, nil, true)
	if !s.IsMobile() {
		t.Fatal("embedded container width did not select mobile mode")
	}
	if _, ok := r.AccessByRole(ui.RoleNavigation, "Sidebar"); ok {
		t.Fatal("closed mobile mode still rendered desktop sidebar")
	}
	sidebarTestClick(t, r, "Toggle sidebar")
	if !s.MobileOpen || !s.Collapsed {
		t.Fatal("mobile toggle changed desktop collapse state")
	}
	if _, ok := r.AccessByRole(ui.RoleDialog, "Sidebar"); !ok {
		t.Fatal("mobile navigation is not a modal sheet")
	}
	r.Key(0, ui.KeyEscape)
	r.Advance(300 * time.Millisecond)
	if s.MobileOpen || !r.Focused("Toggle sidebar") {
		t.Fatal("mobile Escape did not close and restore trigger")
	}
	p.AvailableWidth = 1000
	r.Frame()
	if s.IsMobile() || !s.Collapsed {
		t.Fatal("desktop transition lost independent desktop state")
	}
	nav := sidebarTestAccess(t, r, ui.RoleNavigation, "Sidebar")
	if math.Abs(float64(nav.Bounds.W-48)) > 1 {
		t.Fatalf("icon sidebar width=%v", nav.Bounds.W)
	}
}

func TestSidebarMeasuredEmbeddingAndFixedMode(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	p := SidebarLayoutProps{Items: sidebarTestItems()}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		e, err := SidebarWith(c, nil, "embed", &route, s, p, nil)
		if err != nil {
			t.Fatal(err)
		}
		e.Width(500)
	}, 1000, 500)
	if !s.IsMobile() {
		t.Fatal("sidebar used full window instead of its 500 DIP container")
	}
	p.Collapse = SidebarFixed
	r.Frame()
	if s.IsMobile() {
		t.Fatal("fixed sidebar unexpectedly became a mobile sheet")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Toggle sidebar"); ok {
		t.Fatal("fixed sidebar exposes a collapse trigger")
	}
	if _, ok := r.AccessByRole(ui.RoleNavigation, "Sidebar"); !ok {
		t.Fatal("fixed sidebar disappeared")
	}
}

func TestSidebarMobileTrapSelectionAndViewportTransition(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	mobile := true
	toggles := 0
	p := SidebarLayoutProps{Items: sidebarTestItems(), Label: "Workspace", Mobile: &mobile, OnToggle: func(bool, bool) { toggles++ }}
	r, _, _ := sidebarTestView(t, s, &p, &route, func(c *ui.Context) { ui.Button(c, "Background action") }, false)
	sidebarTestClick(t, r, "Toggle sidebar")
	r.Advance(300 * time.Millisecond)
	if !s.MobileOpen || toggles != 1 {
		t.Fatal("trigger did not open mobile navigation once")
	}
	for range 9 {
		r.Key(0, ui.KeyTab)
		if r.Focused("Background action") || r.Focused("Toggle sidebar") {
			t.Fatal("mobile modal Tab escaped to background")
		}
	}
	sidebarTestClick(t, r, "Settings")
	r.Advance(300 * time.Millisecond)
	if route != "settings" || s.MobileOpen || toggles != 2 || !r.Focused("Toggle sidebar") {
		t.Fatalf("mobile selection route=%q open=%v toggles=%d", route, s.MobileOpen, toggles)
	}
	sidebarTestClick(t, r, "Toggle sidebar")
	r.Advance(300 * time.Millisecond)
	mobile = false
	r.Frame()
	r.Advance(300 * time.Millisecond)
	if s.MobileOpen || s.IsMobile() {
		t.Fatal("desktop resize left mobile modal open")
	}
	if toggles != 4 {
		t.Fatalf("responsive close notification count=%d, want4", toggles)
	}
	if _, ok := r.AccessByRole(ui.RoleDialog, "Workspace"); ok {
		t.Fatal("desktop retained a stale modal")
	}
	if _, ok := r.AccessByRole(ui.RoleNavigation, "Workspace"); !ok {
		t.Fatal("desktop navigation not restored")
	}
}

func TestSidebarLayoutSidesVariantsAndReversibleGeometry(t *testing.T) {
	for _, variant := range []SidebarVariant{SidebarStandard, SidebarFloating, SidebarInset} {
		t.Run(fmt.Sprint(variant), func(t *testing.T) {
			s := &SidebarState{}
			route := "home"
			p := SidebarLayoutProps{Items: sidebarTestItems(), Variant: variant, AvailableWidth: 1000}
			r, _, _ := sidebarTestView(t, s, &p, &route, nil, false)
			main := func() ui.Rect { return sidebarTestAccess(t, r, ui.RoleRegion, "Content").Bounds }
			initial := main()
			if initial.X < 255 {
				t.Fatalf("main does not reserve expanded sidebar space: %+v", initial)
			}
			sidebarTestClick(t, r, "Toggle sidebar")
			r.Advance(80 * time.Millisecond)
			r.Frame()
			middle := main()
			if middle.X >= initial.X || middle.X <= 0 {
				t.Fatalf("no actual width transition: first=%v middle=%v", initial.X, middle.X)
			}
			sidebarTestClick(t, r, "Toggle sidebar")
			r.Frame()
			if math.Abs(float64(main().X-middle.X)) > 1 {
				t.Fatal("collapse reversal jumped geometry")
			}
			r.Advance(400 * time.Millisecond)
			if math.Abs(float64(main().X-initial.X)) > 1 || r.FramePending() {
				t.Fatal("reopened sidebar did not settle")
			}
			r.SetDirection(ui.RightToLeft)
			nav := sidebarTestAccess(t, r, ui.RoleNavigation, "Sidebar")
			if nav.Bounds.X <= main().X {
				t.Fatal("logical start did not switch to RTL right")
			}
			p.Side = SidebarLeft
			r.Frame()
			nav = sidebarTestAccess(t, r, ui.RoleNavigation, "Sidebar")
			if nav.Bounds.X >= main().X {
				t.Fatal("physical left moved with RTL")
			}
		})
	}
}

func TestSidebarIconTooltipsUseExplicitIndependentHost(t *testing.T) {
	s := &SidebarState{Collapsed: true}
	route := "home"
	p := SidebarLayoutProps{Items: sidebarTestItems(), Collapse: SidebarIcon, AvailableWidth: 1000}
	p.Items[0].Tooltip = "Workspace home overview"
	r, b, _ := sidebarTestView(t, s, &p, &route, nil, true)
	home := sidebarTestAccess(t, r, ui.RoleLink, "Home")
	r.Move(home.Bounds.X+home.Bounds.W/2, home.Bounds.Y+home.Bounds.H/2)
	if len(b.surfaces) != 1 {
		t.Fatalf("independent tooltip surfaces=%d", len(b.surfaces))
	}
	if !b.surfaces[0].spec.IgnoreMouseEvents || b.surfaces[0].spec.Focusable {
		t.Fatal("tooltip intercepts native input")
	}
	if !b.surfaces[0].child().HasText("Workspace home overview") {
		t.Fatal("tooltip did not render its own child context")
	}
	if r.HasText("Workspace home overview") {
		t.Fatal("collapsed tooltip painted its label into the parent")
	}
	s.Collapsed = false
	r.Frame()
	if b.surfaces[0].closes != 1 {
		t.Fatal("expanding the sidebar left a detached tooltip")
	}
}

func TestSidebarShortcutScopeStableIdentityAndStatePersistence(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	p := SidebarLayoutProps{Items: sidebarTestItems(), Collapse: SidebarIcon, AvailableWidth: 1000}
	r, _, _ := sidebarTestView(t, s, &p, &route, func(c *ui.Context) { ui.Button(c, "Editor action").AutoFocus() }, true)
	r.Key(ui.Cmd, ui.KeyB)
	if !s.Collapsed {
		t.Fatal("active layout shortcut did not collapse")
	}
	r.Key(ui.Ctrl, ui.KeyB)
	if s.Collapsed {
		t.Fatal("Ctrl+B did not expand")
	}
	sidebarTestClick(t, r, "Settings")
	p.Items[0], p.Items[2] = p.Items[2], p.Items[0]
	r.Frame()
	if !r.Focused("Settings") {
		t.Fatal("stable route key lost focus after reorder")
	}
	s.Groups["projects"] = false
	r.Frame()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var restored SidebarState
	if err = json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Collapsed != s.Collapsed || restored.Groups["projects"] {
		t.Fatalf("caller state did not persist: %s", data)
	}
	if r.FramePending() {
		t.Fatal("stable sidebar scheduled idle redraws")
	}
}

func TestSidebarInvalidKeysDimensionsAndMissingTooltipHost(t *testing.T) {
	state := &SidebarState{}
	route := "home"
	var got error
	p := SidebarLayoutProps{Items: []SidebarItem{{Key: "same", Label: "One"}, {Key: "same", Label: "Two"}}}
	r := ui.NewTester(func(c *ui.Context) { _, got = SidebarWith(c, nil, "bad", &route, state, p, nil) }, 1000, 500)
	if got == nil {
		t.Fatal("duplicate route identity was accepted")
	}
	p.Items = sidebarTestItems()
	p.Width = float32(math.NaN())
	r.Frame()
	if got == nil {
		t.Fatal("nonfinite geometry was accepted")
	}
	p.Width = 0
	p.Collapse = SidebarIcon
	p.AvailableWidth = 1000
	state.Collapsed = true
	r.Frame()
	if got != nil {
		t.Fatal(got)
	}
	home := sidebarTestAccess(t, r, ui.RoleLink, "Home")
	r.Move(home.Bounds.X+4, home.Bounds.Y+4)
	if got == nil || !strings.Contains(got.Error(), "Host") {
		t.Fatal("missing tooltip Host silently rendered a parent-bounded fallback")
	}
}

func TestSidebarShortcutsStayInTheFocusedLayout(t *testing.T) {
	a, b := &SidebarState{}, &SidebarState{}
	routeA, routeB := "home", "home"
	p := SidebarLayoutProps{Items: sidebarTestItems(), Collapse: SidebarIcon, AvailableWidth: 1000}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Row(c).Fill().Children(func() {
			if _, err := SidebarWith(c, nil, "first", &routeA, a, p, func(c *ui.Context) { ui.Button(c, "First editor").AutoFocus() }); err != nil {
				t.Fatal(err)
			}
			if _, err := SidebarWith(c, nil, "second", &routeB, b, p, func(c *ui.Context) { ui.Button(c, "Second editor") }); err != nil {
				t.Fatal(err)
			}
		})
	}, 1200, 500)
	r.Key(ui.Cmd, ui.KeyB)
	if !a.Collapsed || b.Collapsed {
		t.Fatal("shortcut toggled an unrelated sidebar")
	}
	sidebarTestClick(t, r, "Second editor")
	r.Key(ui.Cmd, ui.KeyB)
	if !a.Collapsed || !b.Collapsed {
		t.Fatal("shortcut did not follow actual layout focus")
	}
}

func TestSidebarInitiallyOffcanvasDoesNotMountHiddenAutofocus(t *testing.T) {
	s := &SidebarState{Collapsed: true}
	route := "home"
	value := ""
	built := 0
	p := SidebarLayoutProps{Items: sidebarTestItems(), AvailableWidth: 1000, Header: func(c *ui.Context) { built++; ui.TextInput(c, &value).Label("Sidebar search").AutoFocus() }}
	r, _, _ := sidebarTestView(t, s, &p, &route, func(c *ui.Context) { ui.Button(c, "Main action").AutoFocus() }, true)
	if built != 0 || !r.Focused("Main action") {
		t.Fatal("initial offcanvas content captured focus")
	}
	sidebarTestClick(t, r, "Toggle sidebar")
	if built == 0 {
		t.Fatal("opening did not build the lazy sidebar header")
	}
}

func TestSidebarUsesIndependentSemanticColorsAndReducedMotionGeometry(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	p := SidebarLayoutProps{Items: sidebarTestItems(), AvailableWidth: 1000, Collapse: SidebarIcon}
	tokens := *theme.Preset("blue", true)
	tokens.Background = ui.RGB(11, 17, 25)
	tokens.Sidebar = ui.RGB(38, 22, 54)
	tokens.SidebarAccent = ui.RGB(60, 91, 42)
	r := ui.NewTesterWith(func(c *ui.Context) {
		c.SetTheme(tokens.Theme())
		if _, err := SidebarWith(c, nil, "colors", &route, s, p, nil); err != nil {
			t.Fatal(err)
		}
	}, 1000, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	nav := sidebarTestAccess(t, r, ui.RoleNavigation, "Sidebar").Bounds
	main := sidebarTestAccess(t, r, ui.RoleRegion, "Content").Bounds
	home := sidebarTestAccess(t, r, ui.RoleLink, "Home").Bounds
	check := func(x, y float32, want ui.Color) {
		t.Helper()
		if got := r.Image().RGBAAt(int(x), int(y)); got != (color.RGBA{R: want.R, G: want.G, B: want.B, A: 255}) {
			t.Fatalf("pixel at %v,%v=%v want%+v", x, y, got, want)
		}
	}
	check(nav.X+20, nav.Y+nav.H-20, tokens.Sidebar)
	check(main.X+20, main.Y+main.H-20, tokens.Background)
	check(home.X+home.W/2, home.Y+home.H/2, tokens.SidebarAccent)
	sidebarTestClick(t, r, "Toggle sidebar")
	if sidebarTestAccess(t, r, ui.RoleNavigation, "Sidebar").Bounds.W != 48 || r.FramePending() {
		t.Fatal("reduced-motion icon collapse was not immediate and idle")
	}
}

func TestSidebarRailToggleAndOffcanvasFocusRecovery(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	p := SidebarLayoutProps{Items: sidebarTestItems(), AvailableWidth: 1000}
	r, _, _ := sidebarTestView(t, s, &p, &route, nil, true)
	sidebarTestClick(t, r, "Home")
	if !r.Focused("Home") {
		t.Fatal("route did not have native keyboard focus")
	}
	s.Collapsed = true
	r.Frame()
	if !r.Focused("Toggle sidebar") {
		t.Fatal("controlled offcanvas collapse stranded route focus")
	}
	if _, ok := r.AccessByRole(ui.RoleLink, "Home"); ok {
		t.Fatal("offcanvas route is still accessible")
	}
	rail := sidebarTestAccess(t, r, ui.RoleButton, "Toggle sidebar rail")
	r.ClickAt(max(float32(1), rail.Bounds.X+rail.Bounds.W/2), rail.Bounds.Y+rail.Bounds.H/2)
	if s.Collapsed {
		t.Fatal("visible rail did not reopen offcanvas sidebar")
	}
}

func TestSidebarClipsCustomHeaderAndFooterToCollapsedPane(t *testing.T) {
	for _, footer := range []bool{false, true} {
		t.Run(fmt.Sprint(footer), func(t *testing.T) {
			s := &SidebarState{Collapsed: true}
			route := "home"
			clicks := 0
			build := func(c *ui.Context) {
				button := ui.Button(c, "Wide custom action").Width(300).Height(32).Translate(-120, 0).Background(ui.RGB(220, 10, 10))
				if button.Clicked() {
					clicks++
				}
			}
			p := SidebarLayoutProps{Items: sidebarTestItems(), Collapse: SidebarIcon, Side: SidebarRight, AvailableWidth: 1000}
			if footer {
				p.Footer = build
			} else {
				p.Header = build
			}
			r, _, _ := sidebarTestView(t, s, &p, &route, nil, true)
			action := sidebarTestAccess(t, r, ui.RoleButton, "Wide custom action").Bounds
			x, y := float32(900), action.Y+action.H/2
			if got := r.Image().RGBAAt(int(x), int(y)); got.R == 220 && got.G == 10 {
				t.Fatal("wide custom content painted over main")
			}
			r.ClickAt(x, y)
			if clicks != 0 {
				t.Fatal("clipped custom content intercepted main input")
			}
		})
	}
}

func TestSidebarRemovedTriggerFallsBackToCurrentMainFocus(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	custom := true
	p := SidebarLayoutProps{Items: sidebarTestItems(), AvailableWidth: 1000, HideTrigger: true}
	r, _, _ := sidebarTestView(t, s, &p, &route, func(c *ui.Context) {
		if custom {
			SidebarTrigger(c, s).Label("Custom toggle")
		}
	}, true)
	sidebarTestClick(t, r, "Custom toggle")
	sidebarTestClick(t, r, "Custom toggle")
	sidebarTestClick(t, r, "Home")
	custom = false
	s.Collapsed = true
	r.Frame()
	if !r.Focused("Content") {
		t.Fatal("removed trigger stranded focus in closed pane")
	}
	custom = true
	r.Frame()
	s.Collapsed = false
	r.Frame()
	sidebarTestClick(t, r, "Home")
	s.Collapsed = true
	r.Frame()
	if !r.Focused("Custom toggle") {
		t.Fatal("newly restored trigger did not replace main fallback")
	}
}

func TestSidebarMobileRemovedTriggerReturnsToMain(t *testing.T) {
	s := &SidebarState{}
	route := "home"
	custom := true
	p := SidebarLayoutProps{Items: sidebarTestItems(), AvailableWidth: 500, HideTrigger: true}
	r, _, _ := sidebarTestView(t, s, &p, &route, func(c *ui.Context) {
		if custom {
			SidebarTrigger(c, s).Label("Custom mobile toggle")
		}
	}, true)
	sidebarTestClick(t, r, "Custom mobile toggle")
	custom = false
	r.Frame()
	r.Key(0, ui.KeyEscape)
	if s.MobileOpen || !r.Focused("Content") {
		t.Fatal("removed mobile opener left modal focus stranded")
	}
}

func TestSidebarExternalTriggersBeforeAndAfterTrackCurrentMode(t *testing.T) {
	s, route := &SidebarState{}, "home"
	mobile, after := false, true
	toggles := 0
	p := SidebarLayoutProps{Items: sidebarTestItems(), Mobile: &mobile, Collapse: SidebarIcon, HideTrigger: true, OnToggle: func(bool, bool) { toggles++ }}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			SidebarTrigger(c.Key("before"), s).Label("Before layout")
			if _, err := SidebarWith(c, nil, "external", &route, s, p, nil); err != nil {
				t.Fatal(err)
			}
			if after {
				SidebarTrigger(c.Key("after"), s).Label("After layout")
			}
		})
	}, 1000, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	assertControls := func(expanded bool) {
		t.Helper()
		panel := sidebarTestAccess(t, r, ui.RoleNavigation, "Sidebar")
		for _, label := range []string{"Before layout", "After layout"} {
			button := sidebarTestAccess(t, r, ui.RoleButton, label)
			if button.Expanded != expanded || len(button.Controls) != 1 || button.Controls[0] != panel.Name {
				t.Fatalf("%s relationship=%+v panel=%+v", label, button, panel)
			}
		}
	}
	assertControls(true)
	sidebarTestClick(t, r, "After layout")
	if !s.Collapsed || toggles != 1 {
		t.Fatal("following external trigger did not toggle once")
	}
	assertControls(false)
	sidebarTestClick(t, r, "Before layout")
	if s.Collapsed || toggles != 2 {
		t.Fatal("preceding external trigger did not toggle once")
	}
	mobile = true
	r.Frame()
	for _, label := range []string{"Before layout", "After layout"} {
		button := sidebarTestAccess(t, r, ui.RoleButton, label)
		if button.HasPopup != ui.PopupDialog || button.Expanded {
			t.Fatalf("%s did not update its mobile popup state: %+v", label, button)
		}
	}
	sidebarTestClick(t, r, "Before layout")
	if !s.MobileOpen || s.Collapsed || toggles != 3 {
		t.Fatal("preceding trigger did not use current mobile mode")
	}
	if _, ok := r.AccessByRole(ui.RoleDialog, "Sidebar"); !ok {
		t.Fatal("preceding trigger did not open its mobile dialog")
	}
	r.Key(0, ui.KeyEscape)
	if !r.Focused("Before layout") {
		t.Fatal("mobile did not restore its preceding opener")
	}
	sidebarTestClick(t, r, "After layout")
	if !s.MobileOpen || s.Collapsed || toggles != 5 {
		t.Fatal("following trigger did not use current mobile mode")
	}
	if _, ok := r.AccessByRole(ui.RoleDialog, "Sidebar"); !ok {
		t.Fatal("following trigger did not open its mobile dialog")
	}
	r.Key(0, ui.KeyEscape)
	if !r.Focused("After layout") {
		t.Fatal("mobile did not restore its following opener")
	}
	after = false
	r.Frame()
	for range 4 {
		r.Frame()
	}
	if len(s.triggers) != 1 || r.FramePending() {
		t.Fatal("removed trigger stayed registered or scheduled idle work")
	}
}
