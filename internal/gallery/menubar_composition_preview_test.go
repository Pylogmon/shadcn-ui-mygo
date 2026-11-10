package gallery

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func menubarCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_MENUBAR_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("menubar image %v/%v", err, closeErr)
	}
}

// A single application workflow crosses the system-popup and independent Flyout
// routes. The public Tester/Backend evidence does not stand for a real OS matrix.
func TestMenubarCompositionsSourceWorkflow(t *testing.T) {
	p := &menubarCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(16).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 640, 850, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	click := func(label string) { t.Helper(); compoundClick(t, r, label) }
	latest := func() *sourcePopupSurface {
		t.Helper()
		if len(b.surfaces) == 0 {
			t.Fatal("no independent menu window")
		}
		s := b.surfaces[len(b.surfaces)-1]
		// The native adapter implements FitContent. This injected Backend has
		// no window fitter, so size its public test surface to the measured menu.
		if s.spec.FitContent {
			for _, a := range s.tester.AccessAll() {
				if a.Role == ui.RoleMenu {
					height := int(math.Ceil(float64(a.Bounds.Y+a.Bounds.H))) + 2
					if s.spec.MaxHeight > 0 {
						height = min(height, s.spec.MaxHeight)
					}
					if height != s.spec.Height {
						s.SetSize(s.spec.Width, height)
					}
					break
				}
			}
		}
		return s
	}
	choose := func(selector, label string) {
		t.Helper()
		click(selector)
		if native {
			if err := r.ChooseMenuItem(label); err != nil {
				t.Fatal(err)
			}
		} else {
			compoundClick(t, latest().tester, label)
		}
		r.Frame()
	}
	command := func(menu string, path ...string) {
		t.Helper()
		click(menu)
		if err := r.ChooseMenuItem(path...); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	closeChild := func() {
		t.Helper()
		latest().tester.Key(0, ui.KeyEscape)
		r.Frame()
		r.Advance(time.Second)
	}
	if p.recipe != "demo" || p.menuError != "" {
		t.Fatalf("initial state: %+v", p)
	}
	click("File")
	if len(b.surfaces) != 0 || r.HasText("New Tab") || !slices.Contains(r.Menu(), "New Tab") {
		t.Fatal("NativeMenubar did not use system popup trigger group")
	}
	if err := r.ChooseMenuItem("New Incognito Window"); err == nil {
		t.Fatal("disabled source action accepted")
	}
	if err := r.ChooseMenuItem("New Tab"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.lastAction != "New Tab" || p.commandCount != 1 {
		t.Fatal("source action did not dispatch exactly once")
	}
	command("File", "Share", "Messages")
	command("View", "Full URLs")
	command("Profiles", "Andy")
	if *p.checks["demo:urls"] || p.selections["demo:profile"] != "andy" {
		t.Fatal("native source checkbox/radio state lost")
	}
	menubarCompositionSnapshot(t, r, "menubar-native-parent-light")
	choose("Menubar configuration", "Basic")
	command("Edit", "Copy")
	choose("Menubar configuration", "With Submenu")
	command("Edit", "Find", "Find Previous")
	choose("Menubar configuration", "With Checkboxes")
	command("Format", "Code")
	if !*p.checks["checkbox:code"] || !*p.checks["checkbox:strikethrough"] || !*p.checks["checkbox:urls"] {
		t.Fatal("source checkbox defaults/changes missing")
	}
	choose("Menubar configuration", "With Radio")
	command("Theme", "Dark")
	if p.selections["radio:theme"] != "dark" || p.selections["radio:profile"] != "benoit" {
		t.Fatal("source radio defaults/changes missing")
	}
	choose("Menubar configuration", "With Icons (docs)")
	command("More", "Help")
	if len(p.nativeIcons) < 6 {
		t.Fatal("Material Symbols were not adapted to native menu images")
	}
	choose("Menubar configuration", "With Icons (Create)")
	command("More", "Delete")
	choose("Menubar configuration", "With Shortcuts")
	command("Edit", "Paste")
	choose("Menubar configuration", "Format")
	command("Format", "Bold")
	command("View", "Show Grid")
	if *p.checks["format:grid"] {
		t.Fatal("format source grid did not toggle")
	}
	choose("Menubar configuration", "Insert")
	command("Insert", "Media", "Audio")
	command("Tools", "Spell Check")
	choose("Menubar configuration", "Destructive")
	command("File", "Delete File")
	if p.lastAction != "Delete File" {
		t.Fatal("destructive sample did not report its local action")
	}
	choose("Menubar configuration", "In Dialog")
	click("Open Dialog")
	command("File", "More Options", "Developer Tools")
	if !p.dialog || p.lastAction != "Developer Tools" {
		t.Fatal("native nested menu lost its actual Dialog")
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if p.dialog {
		t.Fatal("source dialog did not dismiss")
	}
	choose("Menubar configuration", "With Inset")
	command("View", "Full URLs")
	command("View", "Dark")
	if !*p.checks["inset:bookmarks"] || !*p.checks["inset:urls"] || p.selections["inset:theme"] != "dark" {
		t.Fatal("inset source state missing")
	}
	choose("Menubar configuration", "Translated File, Edit, View and Profiles")
	choose("Menubar language", "עברית")
	command("קובץ", "שתף", "הודעות")
	if p.lastAction != "הודעות" {
		t.Fatal("source Hebrew submenu did not dispatch")
	}
	choose("Menubar configuration", "File, Edit, View and Profiles")

	// The same application model keeps its checked/radio drafts across routes.
	native, dark = false, true
	r.Frame()
	click("File")
	s := latest()
	if s.spec.SystemPopover || r.HasText("New Tab") || !s.tester.HasText("New Tab") {
		t.Fatal("styled menubar did not open independent Flyout")
	}
	before := len(b.surfaces)
	s.tester.Key(0, ui.KeyRight)
	r.Frame()
	s.tester.Frame()
	if len(b.surfaces) != before || p.bars["demo"].Active != "Edit" || !s.tester.HasText("Undo") {
		t.Fatal("Flyout sibling navigation replaced its window or retained old content")
	}
	count := p.commandCount
	compoundClick(t, s.tester, "Undo")
	r.Frame()
	r.Advance(time.Second)
	if p.lastAction != "Undo" || p.commandCount != count+1 || h.IsOpen("source-menubar:demo") || !r.Focused("Edit") {
		t.Fatal("Flyout action/focus/close missing")
	}
	r.Frame()
	if p.commandCount != count+1 {
		t.Fatal("Flyout command repeated on parent rebuild")
	}
	click("View")
	s = latest()
	if s.spec.Width != 176 {
		t.Fatal("docs Demo View did not preserve its explicit source w-44 width")
	}
	menubarCompositionSnapshot(t, s.tester, "menubar-view-width176-dark")
	if a := compoundAccess(t, s.tester, ui.RoleMenuItemCheckBox, "Full URLs"); a.Checked {
		t.Fatal("route change reset checkbox draft")
	}
	if a := compoundAccess(t, s.tester, ui.RoleMenuItem, "Force Reload"); !a.Disabled {
		t.Fatal("source disabled inset command became enabled")
	}
	compoundClick(t, s.tester, "Bookmarks Bar")
	r.Frame()
	if !h.IsOpen("source-menubar:demo") || !compoundAccess(t, s.tester, ui.RoleMenuItemCheckBox, "Bookmarks Bar").Checked {
		t.Fatal("source checkbox failed to keep its menu open and refresh checked state")
	}
	closeChild()
	click("Profiles")
	s = latest()
	if a := compoundAccess(t, s.tester, ui.RoleMenuItemRadio, "Andy"); !a.Checked {
		t.Fatal("route change reset profile draft")
	}
	compoundClick(t, s.tester, "Luis")
	r.Frame()
	if !h.IsOpen("source-menubar:demo") || !compoundAccess(t, s.tester, ui.RoleMenuItemRadio, "Luis").Checked {
		t.Fatal("source radio failed to keep its menu open and refresh selection")
	}
	closeChild()
	// These are the fixed source's explicit per-menu w-64/w-44/w-40 values;
	// Checkbox's Format has no explicit width and must retain the224 fallback.
	for _, example := range []struct {
		configuration, trigger string
		width                  int
	}{{"With Checkboxes", "View", 256}, {"With Checkboxes", "Format", 224}, {"Insert", "Tools", 176}, {"Destructive", "File", 160}} {
		choose("Menubar configuration", example.configuration)
		click(example.trigger)
		s = latest()
		if s.spec.Width != example.width {
			t.Fatalf("source %s/%s width=%d want%d", example.configuration, example.trigger, s.spec.Width, example.width)
		}
		closeChild()
	}
	choose("Menubar configuration", "With Inset")
	click("View")
	s = latest()
	if !s.tester.HasText("Actions") || !s.tester.HasText("Appearance") || !s.tester.HasText("Theme") {
		t.Fatal("actual source group captions missing")
	}
	if a, ok := s.tester.AccessByRole(ui.RoleMenuItemRadio, "Dark"); !ok || a.Offscreen || !a.Checked {
		t.Fatalf("inset radio state/viewport missing: access=%+v spec=%+v texts=%v", a, s.spec, s.tester.Texts())
	}
	menubarCompositionSnapshot(t, s.tester, "menubar-flyout-inset-dark")
	// The child can receive several selections before its parent builds. Radio
	// rows must be mutually exclusive in that same child frame, not repaired
	// only by reconstructing source data from the parent's selected value.
	uniqueTheme := func(expected string) {
		t.Helper()
		checked := 0
		for _, label := range []string{"Light", "Dark", "System"} {
			a := compoundAccess(t, s.tester, ui.RoleMenuItemRadio, label)
			if a.Checked {
				checked++
				if label != expected {
					t.Errorf("before parent.Frame: %s remains checked after choosing %s", label, expected)
				}
			}
		}
		if checked != 1 {
			t.Fatalf("before parent.Frame: %d checked theme rows after choosing %s", checked, expected)
		}
	}
	compoundClick(t, s.tester, "Light")
	uniqueTheme("Light")
	if !h.IsOpen("source-menubar:inset") || p.selections["inset:theme"] != "light" {
		t.Fatal("same-frame radio edit did not update application value and keep menu open")
	}
	compoundClick(t, s.tester, "Dark")
	uniqueTheme("Dark")
	r.Frame()
	compoundClick(t, s.tester, "Full URLs")
	r.Frame()
	if *p.checks["inset:urls"] || !h.IsOpen("source-menubar:inset") {
		t.Fatal("Flyout inset checkbox did not change application state")
	}
	closeChild()
	choose("Menubar configuration", "Six sides")
	choose("Menubar direction", "Right to left")
	for _, side := range []string{"inline-start", "left", "top", "bottom", "right", "inline-end"} {
		label := map[string]string{"inline-start": "Inline Start", "left": "Left", "top": "Top", "bottom": "Bottom", "right": "Right", "inline-end": "Inline End"}[side]
		click(label)
		s = latest()
		if s.spec.Placement != popupPlacement(side, ui.RightToLeft) || !s.tester.HasText("New Incognito Window") {
			t.Fatalf("source side %q did not map logical direction: %+v", side, s.spec)
		}
		if a := compoundAccess(t, s.tester, ui.RoleMenuItem, "New Incognito Window"); a.Disabled {
			t.Fatal("Sides source incorrectly inherited Basic disabled flag")
		}
		closeChild()
	}
	choose("Menubar configuration", "Translated File, Edit, View and Profiles")
	click("קובץ")
	s = latest()
	if s.spec.Placement != mygo.PlacementBottomEnd || !s.tester.HasText("חלון חדש") {
		t.Fatal("source RTL alignment/translation missing")
	}
	menubarCompositionSnapshot(t, s.tester, "menubar-flyout-hebrew-dark")
	closeChild()
	click("תצוגה")
	s = latest()
	if s.spec.Width != 176 {
		t.Fatal("translated View did not preserve its source w-44 width")
	}
	closeChild()
	choose("Menubar configuration", "In Dialog")
	click("Open Dialog")
	click("File")
	s = latest()
	compoundClick(t, s.tester, "Copy")
	r.Frame()
	r.Advance(time.Second)
	if !p.dialog || !r.Focused("File") || p.lastAction != "Copy" {
		t.Fatal("nested Flyout action failed to preserve dialog and restore trigger")
	}
	menubarCompositionSnapshot(t, r, "menubar-dialog-parent-dark")
	r.Key(0, ui.KeyEscape)
	r.Frame()
	r.Advance(time.Second)
	if p.dialog || !r.Focused("Open Dialog") {
		t.Fatal("parent Dialog focus did not return after child dismissed")
	}
	choose("Menubar configuration", "With Inset")
	click("View")
	visible = false
	r.Frame()
	if h.IsPresent("source-menubar:inset") || p.bars["inset"].Open {
		t.Fatal("leaving source page retained independent menu")
	}
	visible = true
	r.Frame()
	if p.menuError != "" || p.selections["demo:profile"] != "Luis" || !*p.checks["demo:bookmarks"] {
		t.Fatalf("source workflow did not retain caller-owned state: %+v", p)
	}
}

func TestMenubarCompositionsMaterialSourceDifferences(t *testing.T) {
	p := &menubarCompositionPreviewState{}
	p.initialize(false)
	find := func(items []overlay.MenuItem, label string) overlay.MenuItem {
		t.Helper()
		for _, item := range items {
			if item.Label == label {
				return item
			}
		}
		t.Fatalf("source item %q missing", label)
		return overlay.MenuItem{}
	}
	demo := p.items("demo")
	if len(demo) != 4 || len(demo[1].Items[3].Children) != 5 {
		t.Fatal("docs complete bar/find groups missing")
	}
	checkbox := p.items("checkbox")
	if !find(checkbox[0].Items, "Force Reload").Disabled || !find(checkbox[0].Items, "Reload").Inset || !*find(checkbox[1].Items, "Strikethrough").Checked {
		t.Fatal("checkbox source disabled/inset/default differences missing")
	}
	docs, create := p.items("icons"), p.items("create-icons")
	if docs[1].Items[0].Icon == create[1].Items[0].Icon || docs[0].Items[0].Icon != create[0].Items[0].Icon {
		t.Fatal("docs semantic icons/Create placeholder variation collapsed")
	}
	inset := p.items("inset")[0].Items
	if item := find(inset, "Copy"); item.Group != "Actions" || item.GroupInset || item.Inset || item.Icon == nil {
		t.Fatal("normal caption/icon row lost")
	}
	if item := find(inset, "Bookmarks"); item.Group != "Appearance" || !item.GroupInset || !item.Inset || !*item.Checked {
		t.Fatal("independent inset caption/check row lost")
	}
	if item := find(inset, "Light"); item.Group != "Theme" || !item.GroupInset || !item.Inset || !item.Radio {
		t.Fatal("inset radio caption/row lost")
	}
	if item := find(inset, "More Options"); !item.Inset || len(item.Children) != 2 {
		t.Fatal("inset submenu lost")
	}
	if item := find(p.items("destructive")[0].Items, "Delete File"); !item.Destructive || item.Shortcut != "⌘⌫" || item.Icon == nil {
		t.Fatal("source destructive icon/shortcut missing")
	}
}
