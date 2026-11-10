package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func dropdownSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_DROPDOWN_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("snapshot %v/%v", err, closeErr)
	}
}

// A public MyGo Tester/Backend flow exercises both source routes. The OS adapter
// is verified through native menu tracking; screenshots represent styled child
// content rendered by MyGo's software tester, not real OS windows.
func TestDropdownSourceCompositionWorkflow(t *testing.T) {
	p := &dropdownCompositionPreviewState{}
	b := &creatorPopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark := true, false
	parent := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("neutral", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() { h.BeginFrame(); defer h.EndFrame(); p.view(c, h, native) })
	}, 620, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	choose := func(key string) {
		t.Helper()
		r := dropdownRecipe(key)
		if err := parent.Click("Dropdown Menu configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := parent.ChooseMenuItem(r.Title); err != nil {
				t.Fatal(err)
			}
		} else {
			child := b.surfaces[len(b.surfaces)-1].tester
			child.Key(0, ui.KeyHome)
			for _, candidate := range dropdownSourceRecipes {
				if candidate.Key == key {
					break
				}
				child.Key(0, ui.KeyDown)
			}
			child.Key(0, ui.KeyEnter)
		}
		parent.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("recipe %s built %s err=%s", key, p.recipe, p.popupError)
		}
	}
	if p.recipe != "demo" || len(dropdownSourceRecipes) != 27 {
		t.Fatal("source default or complete collection missing")
	}
	for _, r := range dropdownSourceRecipes {
		choose(r.Key)
	}
	// Real platform-image handles preserve the source icons on the OS route.
	iconItems := p.items(dropdownRecipe("complex"))
	p.withNativeIcons(iconItems)
	first := iconItems[0].NativeIcon
	if first == nil || iconItems[2].Children[0].NativeIcon == nil {
		t.Fatal("native source icons or nested icons missing")
	}
	again := p.items(dropdownRecipe("complex"))
	p.withNativeIcons(again)
	if first != again[0].NativeIcon {
		t.Fatal("platform menu image handle was not reused")
	}
	choose("checkboxes")
	if err := parent.Click("Open"); err != nil {
		t.Fatal(err)
	}
	if err := parent.ChooseMenuItem("Activity Bar"); err == nil {
		t.Fatal("disabled source checkbox accepted")
	}
	if err := parent.ChooseMenuItem("Status Bar"); err != nil {
		t.Fatal(err)
	}
	if *p.model(dropdownRecipe("checkboxes")).checks["showStatusBar"] {
		t.Fatal("source checkbox did not toggle")
	}
	choose("complex")
	model := p.model(dropdownRecipe("complex"))
	if !*model.checks["notifications.email"] || model.radios["theme"] != "light" {
		t.Fatal("docs complex defaults changed")
	}
	if err := parent.Click("Complex Menu"); err != nil {
		t.Fatal(err)
	}
	if err := parent.ChooseMenuItem("Settings", "Notifications", "Email Notifications"); err != nil {
		t.Fatal(err)
	}
	if *model.checks["notifications.email"] {
		t.Fatal("third-level source action did not share its real check model")
	}
	if len(b.surfaces) != 0 {
		t.Fatal("system native menus opened themed popup windows")
	}
	native = false
	parent.Frame()
	choose("create-with-inset")
	if err := parent.Click("Open"); err != nil {
		t.Fatal(err)
	}
	inset := b.surfaces[len(b.surfaces)-1]
	inset.tester.Advance(time.Second)
	if inset.spec.SystemPopover || parent.HasText("Bookmarks") || !inset.tester.HasText("Bookmarks") {
		t.Fatal("styled source menu was not an independent Flyout")
	}
	bookmarks, ok := inset.tester.AccessByRole(ui.RoleMenuItemCheckBox, "Bookmarks")
	if !ok || !bookmarks.Checked {
		t.Fatal("inset source default lost")
	}
	inset.tester.ClickAt(bookmarks.Bounds.X+bookmarks.Bounds.W/2, bookmarks.Bounds.Y+bookmarks.Bounds.H/2)
	parent.Frame()
	if *p.model(dropdownRecipe("create-with-inset")).checks["showBookmarks"] || !h.IsOpen("dropdown-source:create-with-inset:main") {
		t.Fatal("inset checkbox failed to update while remaining open")
	}
	inset.tester.Key(0, ui.KeyEnd)
	inset.tester.Key(0, ui.KeyRight)
	if !inset.tester.HasText("Save Page...") {
		t.Fatal("inset submenu omitted source actions")
	}
	// This injected backend does not implement the native window fitter. Size
	// its software viewport explicitly to the two measured panel widths.
	inset.tester.SetSize(356, inset.spec.Height)
	inset.tester.Advance(time.Second)
	if a, ok := inset.tester.AccessByRole(ui.RoleMenuItem, "Save Page..."); !ok || a.Bounds.X < 0 || a.Bounds.X+a.Bounds.W > 356 {
		t.Fatal("inset submenu outside expanded software viewport")
	}
	dropdownSnapshot(t, inset.tester, "dropdown-inset-light")
	inset.tester.Key(0, ui.KeyEnter)
	parent.Frame()
	if p.action != "Local dropdown action: Save Page..." {
		t.Fatal("source submenu action was inert")
	}
	dark = true
	parent.Frame()
	choose("checkboxes-icons")
	if err := parent.Click("Notifications"); err != nil {
		t.Fatal(err)
	}
	icons := b.surfaces[len(b.surfaces)-1]
	icons.tester.Advance(time.Second)
	dropdownSnapshot(t, icons.tester, "dropdown-checkbox-icons-dark")
	icons.tester.Key(0, ui.KeyEscape)
	parent.Frame()
	choose("radio-group")
	if err := parent.Click("Open"); err != nil {
		t.Fatal(err)
	}
	radio := b.surfaces[len(b.surfaces)-1]
	radio.tester.Key(0, ui.KeyHome)
	radio.tester.Key(0, ui.KeyEnter)
	radio.tester.Key(0, ui.KeyDown)
	radio.tester.Key(0, ui.KeyEnter)
	// No parent Frame: mutual exclusion must be true in this live child now.
	top, topOK := radio.tester.AccessByRole(ui.RoleMenuItemRadio, "Top")
	bottom, bottomOK := radio.tester.AccessByRole(ui.RoleMenuItemRadio, "Bottom")
	if !topOK || !bottomOK || top.Checked || !bottom.Checked {
		t.Fatal("source radios were not mutually exclusive in the current child frame")
	}
	radio.tester.Key(0, ui.KeyEscape)
	parent.Frame()
	choose("complex")
	if err := parent.Click("Complex Menu"); err != nil {
		t.Fatal(err)
	}
	complex := b.surfaces[len(b.surfaces)-1]
	complex.tester.Advance(time.Second)
	complex.tester.Key(0, ui.KeyHome)
	complex.tester.Key(0, ui.KeyDown)
	complex.tester.Key(0, ui.KeyDown)
	complex.tester.Key(0, ui.KeyRight)
	complex.tester.Key(0, ui.KeyDown)
	complex.tester.Key(0, ui.KeyDown)
	complex.tester.Key(0, ui.KeyRight)
	if !complex.tester.HasText("Project Gamma") {
		t.Fatal("docs complex omitted its third panel")
	}
	complex.tester.SetSize(536, complex.spec.Height)
	complex.tester.Advance(time.Second)
	if a, ok := complex.tester.AccessByRole(ui.RoleMenuItem, "Project Gamma"); !ok || a.Bounds.X < 0 || a.Bounds.X+a.Bounds.W > 536 {
		t.Fatal("third panel outside expanded software viewport")
	}
	dropdownSnapshot(t, complex.tester, "dropdown-complex-dark")
	complex.tester.Key(0, ui.KeyEscape)
	complex.tester.Key(0, ui.KeyEscape)
	complex.tester.Key(0, ui.KeyEscape)
	parent.Frame()
	choose("rtl")
	if err := parent.Click("افتح القائمة"); err != nil {
		t.Fatal(err)
	}
	rtl := b.surfaces[len(b.surfaces)-1]
	rtl.tester.Advance(time.Second)
	rtl.tester.Key(0, ui.KeyHome)
	rtl.tester.Key(0, ui.KeyLeft)
	if !rtl.tester.HasText("الملف الشخصي") {
		t.Fatal("source RTL submenu did not use logical keyboard direction")
	}
	rtl.tester.SetSize(292, rtl.spec.Height)
	rtl.tester.Advance(time.Second)
	if a, ok := rtl.tester.AccessByRole(ui.RoleMenuItem, "الملف الشخصي"); !ok || a.Bounds.X < 0 || a.Bounds.X+a.Bounds.W > 292 {
		t.Fatal("RTL panel outside expanded software viewport")
	}
	dropdownSnapshot(t, rtl.tester, "dropdown-rtl-dark")
	rtl.tester.Key(0, ui.KeyEscape)
	rtl.tester.Key(0, ui.KeyEscape)
	parent.Frame()
	choose("create-sides")
	for _, side := range []string{"inline-start", "left", "top", "bottom", "right", "inline-end"} {
		label := side
		if side == "inline-start" {
			label = "inline start"
		}
		if side == "inline-end" {
			label = "inline end"
		}
		if err := parent.Click(label); err != nil {
			t.Fatal(err)
		}
		surface := b.surfaces[len(b.surfaces)-1]
		if surface.spec.Placement != popupPlacement(side, ui.LeftToRight) {
			t.Fatalf("source side %s is %v", side, surface.spec.Placement)
		}
		surface.tester.Key(0, ui.KeyEscape)
		parent.Frame()
	}
	choose("avatar")
	if err := parent.Click("Open shadcn avatar menu"); err != nil {
		t.Fatal(err)
	}
	docsAvatar := b.surfaces[len(b.surfaces)-1]
	if docsAvatar.spec.Placement != mygo.PlacementBottomEnd {
		t.Fatal("docs avatar source end alignment lost")
	}
	docsAvatar.tester.Key(0, ui.KeyEscape)
	parent.Frame()
	choose("create-with-avatar")
	if err := parent.Click("Open shadcn avatar menu above"); err != nil {
		t.Fatal(err)
	}
	avatar := b.surfaces[len(b.surfaces)-1]
	if avatar.spec.Placement != mygo.PlacementTopEnd {
		t.Fatal("source top/end avatar placement lost")
	}
	avatar.tester.Key(0, ui.KeyEscape)
	parent.Frame()
	choose("create-in-dialog")
	if err := parent.Click("Open dropdown dialog"); err != nil {
		t.Fatal(err)
	}
	if !parent.HasText("Dropdown Menu Example") {
		t.Fatal("source dialog missing")
	}
	if err := parent.Click("Open Menu"); err != nil {
		t.Fatal(err)
	}
	dialog := b.surfaces[len(b.surfaces)-1]
	if !dialog.tester.HasText("Developer Tools") {
		dialog.tester.Key(0, ui.KeyHome)
		for range 3 {
			dialog.tester.Key(0, ui.KeyDown)
		}
		dialog.tester.Key(0, ui.KeyRight)
	}
	if !dialog.tester.HasText("Developer Tools") {
		t.Fatal("dialog source submenu incomplete")
	}
	dialog.tester.Key(0, ui.KeyEnd)
	dialog.tester.Key(0, ui.KeyEnter)
	parent.Frame()
	if p.action != "Local dropdown action: Developer Tools" || !p.dialog {
		t.Fatal("dialog menu action disturbed surrounding modal")
	}
	if len(parent.OpenedURLs()) != 0 || p.popupError != "" {
		t.Fatal("placeholder action left local application")
	}
}
