package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func contextCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_CONTEXT_MENU_COMPOSITION_SNAPSHOT_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, r.Image())
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("context menu PNG %v/%v", err, closeErr)
	}
}

// A single public Tester/Backend workflow exercises actual secondary-click,
// context-key, system menu selections and independent Flyout UI. It does not
// claim a real operating-system appearance or platform test matrix.
func TestContextMenuCompositionsSourceWorkflow(t *testing.T) {
	p := &contextMenuCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(16).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave context examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show context examples").Clicked() {
					visible = true
				}
			})
			ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Role(ui.RoleRegion).Label("Context workspace").Children(func() {
				ui.Column(c).FillWidth().MinWidth(0).Gap(20).Children(func() {
					if visible {
						p.view(c, h, native)
					}
				})
			})
		})
	}, 620, 980, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, label)
		if !ok || a.Bounds.W <= 0 || a.Bounds.H <= 0 || a.Bounds.X < 0 || a.Bounds.X+a.Bounds.W > 621 || a.Bounds.Y < 0 || a.Bounds.Y+a.Bounds.H > 981 {
			t.Fatalf("missing/unbounded context control %v %q: %+v", role, label, a)
		}
		return a
	}
	click := func(role ui.Role, label string) {
		t.Helper()
		a := access(role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	latest := func() *sourcePopupSurface {
		t.Helper()
		if len(b.surfaces) == 0 {
			t.Fatal("no independent context-menu surface")
		}
		return b.surfaces[len(b.surfaces)-1]
	}
	choose := func(selector, label string) {
		t.Helper()
		role := ui.RoleButton
		if native {
			role = ui.RoleMenuButton
		}
		click(role, selector)
		if native {
			if err := r.ChooseMenuItem(label); err != nil {
				t.Fatal(err)
			}
		} else {
			index := slices.IndexFunc(contextMenuCompositionRecipes, func(option control.Option) bool { return option.Label == label })
			if index < 0 {
				t.Fatalf("unknown source configuration %q", label)
			}
			latest().tester.Key(0, ui.KeyHome)
			for range index {
				latest().tester.Key(0, ui.KeyDown)
			}
			latest().tester.Key(0, ui.KeyEnter)
		}
		r.Frame()
		if selector == "Context menu configuration" && !slices.ContainsFunc(contextMenuCompositionRecipes, func(option control.Option) bool { return option.Label == label && option.Value == p.recipe }) {
			t.Fatalf("selector did not choose %s; current=%s", label, p.recipe)
		}
	}
	open := func(label string) {
		t.Helper()
		access(ui.RoleGroup, label)
		if err := r.RightClick(label); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	selectNative := func(labels ...string) {
		t.Helper()
		if err := r.ChooseMenuItem(labels...); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	if p.recipe != "basic" || p.popupError != "" {
		t.Fatal("initial source configuration missing")
	}
	open("Context menu target")
	if len(b.surfaces) != 0 || !slices.Contains(r.Menu(), "Back") || r.HasText("Forward") {
		t.Fatal("native context menu did not stay in the system adapter")
	}
	if err := r.ChooseMenuItem("Forward"); err == nil {
		t.Fatal("source disabled Forward activated")
	}
	selectNative("Back")
	if p.action != "Local context action: Back" {
		t.Fatal("native local action did not dispatch")
	}
	// Every pinned docs/Create composition is selected through the real selector.
	for _, recipe := range contextMenuCompositionRecipes {
		choose("Context menu configuration", recipe.Label)
		if p.recipe != recipe.Value || p.popupError != "" {
			t.Fatalf("source configuration %s did not build: %s", recipe.Value, p.popupError)
		}
		if strings.HasSuffix(recipe.Value, "sides") {
			label := "Context menu target: top"
			if recipe.Value == "create-sides" {
				label = "Context menu target: inline-start"
			}
			open(label)
			selectNative("Forward") // Unlike Basic, source side demos enable it.
		} else if recipe.Value == "create-dialog" {
			click(ui.RoleButton, "Open Dialog")
			open("Context menu target")
			selectNative("More Options", "Developer Tools")
			if !p.dialog {
				t.Fatal("system child replaced the actual parent Dialog")
			}
			contextCompositionSnapshot(t, r, "context-native-dialog-parent-light")
			click(ui.RoleButton, "Close dialog")
		} else {
			open("Context menu target")
			switch strings.TrimPrefix(recipe.Value, "create-") {
			case "basic":
				selectNative("Reload")
			case "demo":
				selectNative("More Tools", "Delete")
			case "icons":
				selectNative("Copy")
			case "shortcuts":
				selectNative("Save As...")
			case "submenu":
				selectNative("More Tools", "Name Window...")
			case "groups":
				selectNative("New File")
			case "checkboxes":
				selectNative("Show Full URLs")
				if !p.draft().bookmarks || !p.draft().developer || !p.draft().urls {
					t.Fatal("source checkbox defaults/change missing")
				}
			case "radio":
				selectNative("Colm Tuite")
				if p.draft().person != "colm" || p.draft().theme != "light" {
					t.Fatal("independent People and Theme groups missing")
				}
				open("Context menu target")
				selectNative("Dark")
				if p.draft().person != "colm" || p.draft().theme != "dark" {
					t.Fatal("Theme choice changed the People group")
				}
			case "destructive":
				archive := slices.Contains(r.Menu(), "Archive")
				if archive != (recipe.Value == "create-destructive") {
					t.Fatal("docs/Create destructive material difference lost")
				}
				selectNative("Delete")
			case "rtl":
				selectNative("التنقل", "رجوع")
			case "inset":
				if p.draft().theme != "system" || !p.draft().bookmarks || p.draft().urls {
					t.Fatal("source inset defaults changed")
				}
				selectNative("Dark")
			}
		}
	}
	choose("Context menu configuration", "Translated browser menu")
	choose("Context menu language", "עברית")
	open("Context menu target")
	selectNative("כלים נוספים", "שמור עמוד...")
	if p.action != "Local context action: שמור עמוד..." || !r.HasText("לחץ לחיצה ימנית כאן") {
		t.Fatal("source Hebrew target/action missing")
	}
	// Context key opens the same focused native target.
	click(ui.RoleGroup, "Context menu target")
	r.Key(0, ui.KeyContextMenu)
	if !slices.Contains(r.Menu(), "ניווט") {
		t.Fatal("native focused Context key did not open translated menu")
	}
	selectNative("Colm Tuite")
	choose("Context menu configuration", "Create: With Inset")
	native, dark = false, true
	r.Frame()
	open("Context menu target")
	s := latest()
	if s.spec.SystemPopover || len(r.Menu()) != 0 || r.HasText("Appearance") || s.font == "monospace" {
		t.Fatal("Flyout route did not use independently styled native controls")
	}
	bookmarks, ok := s.tester.AccessByRole(ui.RoleMenuItemCheckBox, "Bookmarks")
	darkOption, darkOK := s.tester.AccessByRole(ui.RoleMenuItemRadio, "Dark")
	if !ok || !bookmarks.Checked || !darkOK || !darkOption.Checked {
		t.Fatal("system choice did not survive into independent Flyout state")
	}
	contextCompositionSnapshot(t, s.tester, "context-inset-flyout-dark")
	if err := s.tester.Click("Full URLs"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !p.draft().urls || !h.IsOpen("source-context:create-inset:target") {
		t.Fatal("independent Flyout checkbox did not update and remain open")
	}
	if err := s.tester.Click("Light"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.draft().theme != "light" || !h.IsOpen("source-context:create-inset:target") {
		t.Fatal("styled radio did not update and remain open")
	}
	if err := s.tester.Click("Dark"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if err := s.tester.Click("More Options"); err != nil {
		t.Fatal(err)
	}
	s.tester.SetSize(356, s.spec.Height)
	s.tester.Frame()
	for _, label := range []string{"Save Page...", "Create Shortcut..."} {
		if _, found := s.tester.AccessByRole(ui.RoleMenuItem, label); !found {
			t.Fatalf("inset source submenu lost %q", label)
		}
	}
	if _, found := s.tester.AccessByRole(ui.RoleMenuItem, "Name Window..."); found {
		t.Fatal("inset source submenu added a third action")
	}
	// Source checkbox/radio rows retain their menu; the OS adapter uses the
	// platform tracking behavior. Explicit dismissal ends this styled edit.
	s.tester.Key(0, ui.KeyEscape)
	s.tester.Key(0, ui.KeyEscape)
	r.Frame()
	// The styled selector itself is an independent menu, then six requested
	// placements resolve in the current parent direction (including logical sides).
	choose("Context menu configuration", "Create: Six Sides")
	r.SetDirection(ui.RightToLeft)
	for _, side := range []string{"inline-start", "left", "top", "bottom", "right", "inline-end"} {
		open("Context menu target: " + side)
		s = latest()
		if s.spec.Placement != popupPlacement(side, ui.RightToLeft) {
			t.Fatalf("current RTL side %s: %v", side, s.spec.Placement)
		}
		s.tester.Key(0, ui.KeyEscape)
		r.Frame()
	}
	r.SetDirection(ui.LeftToRight)
	choose("Context menu configuration", "Browser context menu")
	open("Context menu target")
	s = latest()
	if err := s.tester.Click("More Tools"); err != nil {
		t.Fatal(err)
	}
	s.tester.SetSize(376, max(s.spec.Height, 240))
	s.tester.Frame()
	demoRoot, _ := s.tester.AccessByRole(ui.RoleMenu, "Context menu")
	demoChild, _ := s.tester.AccessByRole(ui.RoleMenu, "More Tools")
	if demoRoot.Bounds.W != 192 || demoChild.Bounds.W != 176 || demoChild.Bounds.X-demoRoot.Bounds.X-demoRoot.Bounds.W != 4 {
		t.Fatalf("source Demo root192/sub176/gap4: root=%v child=%v", demoRoot.Bounds, demoChild.Bounds)
	}
	s.tester.Key(0, ui.KeyEscape)
	s.tester.Key(0, ui.KeyEscape)
	r.Frame()
	r.SetDirection(ui.RightToLeft)
	choose("Context menu configuration", "Translated browser menu")
	open("Context menu target")
	s = latest()
	if err := s.tester.Click("ניווט"); err != nil {
		t.Fatal(err)
	}
	// The injected Backend does not run the operating system's FitContent
	// sizing loop. Give both native panels a canvas, as menu SDK tests do.
	s.tester.SetSize(376, s.spec.Height)
	s.tester.Frame()
	r.Frame()
	forward, found := s.tester.AccessByRole(ui.RoleMenuItem, "קדימה")
	if !found || !forward.Disabled {
		t.Fatalf("translated submenu disabled Forward: found=%v disabled=%v", found, forward.Disabled)
	}
	rootPanel, rootOK := s.tester.AccessByRole(ui.RoleMenu, "Context menu")
	subPanel, subOK := s.tester.AccessByRole(ui.RoleMenu, "ניווט")
	if !s.spec.FitContent || !rootOK || !subOK || rootPanel.Bounds.W != 192 || subPanel.Bounds.W != 176 || rootPanel.Bounds.X-(subPanel.Bounds.X+subPanel.Bounds.W) != 4 {
		t.Fatalf("source RTL panels root192/sub176/gap4: root=%v child=%v", rootPanel.Bounds, subPanel.Bounds)
	}
	contextCompositionSnapshot(t, s.tester, "context-hebrew-submenu-dark")
	s.tester.Key(0, ui.KeyEscape)
	if err := s.tester.Click("כלים נוספים"); err != nil {
		t.Fatal(err)
	}
	rtlTools, _ := s.tester.AccessByRole(ui.RoleMenu, "כלים נוספים")
	if rtlTools.Bounds.W != 176 {
		t.Fatalf("source RTL More Tools width=%v", rtlTools.Bounds)
	}
	s.tester.Key(0, ui.KeyEscape)
	s.tester.Key(0, ui.KeyEscape)
	r.Frame()
	choose("Context menu configuration", "Create: In Dialog")
	click(ui.RoleButton, "Open Dialog")
	open("Context menu target")
	s = latest()
	if !p.dialog {
		t.Fatal("Flyout did not retain parent Dialog")
	}
	if err := s.tester.Click("More Options"); err != nil {
		t.Fatal(err)
	}
	s.tester.SetSize(392, max(s.spec.Height, 240))
	s.tester.Frame()
	r.Frame()
	dialogSubmenu, _ := s.tester.AccessByRole(ui.RoleMenu, "More Options")
	if dialogSubmenu.Bounds.W != 192 {
		t.Fatalf("unsized source submenu lost shared192 fallback: %v", dialogSubmenu.Bounds)
	}
	contextCompositionSnapshot(t, s.tester, "context-dialog-submenu-flyout-dark")
	if err := s.tester.Click("Save Page..."); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !p.dialog || p.action != "Local context action: Save Page..." {
		t.Fatal("child local action dismissed its actual parent")
	}
	click(ui.RoleButton, "Close dialog")
	choose("Context menu configuration", "Basic context menu")
	click(ui.RoleGroup, "Context menu target")
	target := access(ui.RoleGroup, "Context menu target")
	r.Key(0, ui.KeyContextMenu)
	r.Frame()
	s = latest()
	if s.spec.Anchor.W != 1 || s.spec.Anchor.H != 1 || s.spec.Anchor.X != target.Bounds.X || s.spec.Anchor.Y != target.Bounds.Y+target.Bounds.H {
		t.Fatal("styled Context key lost point anchoring")
	}
	click(ui.RoleButton, "Leave context examples")
	if h.IsPresent("source-context:basic:target") {
		t.Fatal("leaving application page retained a child window")
	}
	click(ui.RoleButton, "Show context examples")
	if p.popupError != "" || p.drafts["create-inset"].theme != "dark" || !p.drafts["create-inset"].urls || !p.drafts["checkboxes"].urls || !p.drafts["create-checkboxes"].urls {
		t.Fatal("independent configuration values did not survive page return")
	}
}
