package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func checkboxCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_CHECKBOX_COMPOSITION_SNAPSHOT_DIR")
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
		t.Fatalf("checkbox PNG %v/%v", err, closeErr)
	}
}

// One public-input application flow covers the fixed source fields, choice
// cards and controlled selection table. Its software image/backend evidence
// does not stand in for real OS, assistive-technology or platform-font tests.
func TestCheckboxCompositionsSourceWorkflow(t *testing.T) {
	p := &checkboxCompositionPreviewState{}
	b := &sourcePopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, visible := true, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			h.BeginFrame()
			defer h.EndFrame()
			if visible {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another application page")
			}
		})
	}, 620, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	choose := func(key string) {
		t.Helper()
		if err := r.Click("Checkbox configuration"); err != nil {
			t.Fatal(err)
		}
		if native {
			if err := r.ChooseMenuItem(checkboxRecipeLabel(key)); err != nil {
				t.Fatal(err)
			}
		} else {
			child := b.surfaces[len(b.surfaces)-1].tester
			child.Key(0, ui.KeyHome)
			for _, recipe := range checkboxCompositionRecipes {
				if recipe.Value == key {
					break
				}
				child.Key(0, ui.KeyDown)
			}
			child.Key(0, ui.KeyEnter)
		}
		r.Frame()
		if p.recipe != key || p.popupError != "" {
			t.Fatalf("source configuration %s built %s: %s", key, p.recipe, p.popupError)
		}
	}
	access := func(name string, index int) ui.AccessInfo {
		t.Helper()
		for _, a := range r.AccessAll() {
			if a.Role == ui.RoleCheckBox && a.Name == name {
				if index == 0 {
					return a
				}
				index--
			}
		}
		t.Fatalf("missing native checkbox %q occurrence %d", name, index)
		return ui.AccessInfo{}
	}
	click := func(a ui.AccessInfo) {
		t.Helper()
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	if p.recipe != "demo" || len(checkboxCompositionRecipes) != 15 {
		t.Fatal("source default or complete collection missing")
	}
	terms := "Accept terms and conditions"
	notifications := "Enable notifications"
	if access(terms, 0).Checked || !access(terms, 1).Checked || !access(notifications, 0).Disabled || access(notifications, 1).Checked {
		t.Fatal("docs Demo lost its independent source defaults")
	}
	// Click the actual caption, then Space on its real native target.
	if err := r.Click(terms); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !p.drafts["demo"].values[0] || !p.drafts["demo"].values[1] || p.changes["demo"] != 1 {
		t.Fatal("native label did not toggle just the first checkbox once")
	}
	r.Key(0, ui.KeySpace)
	r.Frame()
	if p.drafts["demo"].values[0] || p.changes["demo"] != 2 {
		t.Fatal("Space did not operate the actual focused native target")
	}
	before := p.changes["demo"]
	click(access(notifications, 0))
	if p.drafts["demo"].values[2] || p.changes["demo"] != before {
		t.Fatal("disabled source control accepted a pointer action")
	}
	if err := r.Click("You can enable or disable notifications at any time."); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !p.drafts["demo"].values[3] || p.changes["demo"] != before+1 {
		t.Fatal("the real whole-card label did not activate exactly once")
	}
	checkboxCompositionSnapshot(t, r, "checkbox-demo-light")
	for _, recipe := range checkboxCompositionRecipes {
		choose(recipe.Value)
	}
	choose("description")
	if !access(terms, 0).Checked || !strings.Contains(access(terms, 0).Description, "terms and conditions.") {
		t.Fatal("source description/default was not associated with the checkbox")
	}
	choose("invalid")
	if !access(terms, 0).Invalid {
		t.Fatal("source invalid state did not reach the actual native target")
	}
	click(access(terms, 0))
	if !p.drafts["invalid"].values[0] {
		t.Fatal("invalid source checkbox was incorrectly made disabled")
	}
	choose("group")
	if !access("Hard disks", 0).Checked || !access("External disks", 0).Checked || access("Connected servers", 0).Checked || !r.HasText("Select the items you want to show on the desktop.") {
		t.Fatal("docs desktop group lost its defaults or description")
	}
	click(access("Connected servers", 0))
	choose("create-group")
	if access("Hard disks", 0).Checked || access("External disks", 0).Checked || r.HasText("Select the items you want to show on the desktop.") {
		t.Fatal("Create group was replaced by the materially different docs group")
	}
	choose("group")
	if !access("Connected servers", 0).Checked {
		t.Fatal("independent desktop preferences did not survive configuration changes")
	}
	dark = true
	r.Frame()
	choose("create-title")
	if !access(notifications, 0).Checked || !access(notifications, 1).Disabled {
		t.Fatal("Create title cards lost enabled/default and disabled differences")
	}
	if err := r.Click("You can enable or disable notifications at any time."); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.drafts["create-title"].values[0] || p.changes["create-title"] != 1 {
		t.Fatal("Create whole-card activation did not change the native checkbox once")
	}
	click(access(notifications, 1))
	if p.drafts["create-title"].values[1] || p.changes["create-title"] != 1 {
		t.Fatal("disabled title card accepted input")
	}
	checkboxCompositionSnapshot(t, r, "checkbox-title-dark")
	choose("table")
	r.Frame()
	if !access("Select Sarah Chen", 0).Checked || access("Select all rows", 0).Checked || access("Select all rows", 0).Mixed {
		t.Fatal("source table initial selection or boolean partial-header contract changed")
	}
	click(access("Select all rows", 0))
	for _, row := range checkboxSourcePeople {
		if !access("Select "+row.name, 0).Checked {
			t.Fatal("all-selected command failed to update a real row checkbox")
		}
	}
	click(access("Select Marcus Rodriguez", 0))
	if access("Select all rows", 0).Checked || access("Select all rows", 0).Mixed || access("Select Marcus Rodriguez", 0).Checked || !access("Select David Kim", 0).Checked {
		t.Fatal("partial selection did not derive the pinned boolean header without crossing rows")
	}
	r.SetSize(820, 900)
	r.Frame()
	checkboxCompositionSnapshot(t, r, "checkbox-table-dark")
	r.SetSize(360, 900)
	r.Frame()
	if p.drafts["table"].scroll.MaxX <= 0 {
		t.Fatal("the original complete table became inaccessible in a narrow parent")
	}
	table, ok := r.AccessByRole(ui.RoleTable, "Source checkbox table")
	if !ok {
		t.Fatal("source table lost its native table semantics")
	}
	r.Scroll(200, table.Bounds.Y+20, p.drafts["table"].scroll.MaxX, 0)
	r.Frame()
	if p.drafts["table"].scroll.X <= 0 {
		t.Fatal("actual horizontal input did not reach the remaining source columns")
	}
	r.Scroll(200, table.Bounds.Y+20, -p.drafts["table"].scroll.MaxX, 0)
	r.SetSize(620, 900)
	r.Frame()
	choose("create-table")
	if !access("Select Sarah Chen", 0).Checked || access("Select David Kim", 0).Checked {
		t.Fatal("Create table did not keep independent original selection")
	}
	choose("table")
	if access("Select Marcus Rodriguez", 0).Checked || !access("Select David Kim", 0).Checked {
		t.Fatal("table selection did not survive switching to the independent Create table")
	}
	choose("rtl")
	if err := r.Click("Checkbox language"); err != nil {
		t.Fatal(err)
	}
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	hebrew := "קבל תנאים והגבלות"
	click(access(hebrew, 0))
	if !p.drafts["rtl"].values[0] || p.changes["rtl"] != 1 || !r.HasText("אתה יכול להפעיל או להשבית התראות בכל עת.") {
		t.Fatal("Hebrew source content or native action did not use its independent model")
	}
	checkboxCompositionSnapshot(t, r, "checkbox-hebrew-dark")
	visible = false
	r.Frame()
	visible = true
	r.Frame()
	if p.recipe != "rtl" || !access(hebrew, 0).Checked || p.changes["rtl"] != 1 {
		t.Fatal("source state was lost after the composition left the application page")
	}
	native = false
	r.Frame()
	choose("basic")
	if h.IsOpen("source-checkbox-config") || access(terms, 0).Checked {
		t.Fatal("Flyout selector did not close after choosing or corrupted basic defaults")
	}
	visible = false
	r.Frame()
	for _, surface := range b.surfaces {
		if !surface.didClose {
			t.Fatal("an absent composition retained an independent configuration popup")
		}
	}
}
