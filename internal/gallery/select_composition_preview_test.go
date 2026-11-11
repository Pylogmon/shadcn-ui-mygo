package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func selectCompositionSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_SELECT_COMPOSITION_SNAPSHOT_DIR")
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
	if closeErr := f.Close(); err != nil || closeErr != nil {
		t.Fatalf("select PNG %v/%v", err, closeErr)
	}
}

// One public native-input flow verifies complete source listboxes. The source
// controls stay in-window; only configuration menus use the desktop Host.
func TestSelectCompositionsSourceWorkflow(t *testing.T) {
	p := &selectCompositionPreviewState{}
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
	access := func(role ui.Role, label string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing native role=%v label=%q", role, label)
		}
		return a
	}
	click := func(role ui.Role, label string) {
		t.Helper()
		a := access(role, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
	}
	choose := func(recipe string) {
		t.Helper()
		r.Key(0, ui.KeyEscape)
		r.Frame()
		index := -1
		for i, option := range selectCompositionRecipes {
			if option.Value == recipe {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown source recipe %s", recipe)
		}
		if native {
			click(ui.RoleMenuButton, "Select configuration")
			if err := r.ChooseMenuItem(selectCompositionRecipes[index].Label); err != nil {
				selectCompositionSnapshot(t, r, "select-configuration-diagnostic")
				t.Logf("selector bounds=%+v", access(ui.RoleMenuButton, "Select configuration").Bounds)
				t.Fatalf("configuration %s: %v", recipe, err)
			}
		} else {
			click(ui.RoleComboBox, "Select configuration")
			child := b.surfaces[len(b.surfaces)-1]
			child.tester.Key(0, ui.KeyHome)
			for range index {
				child.tester.Key(0, ui.KeyDown)
			}
			child.tester.Key(0, ui.KeyEnter)
			if !child.didClose {
				t.Fatal("styled configuration selector did not close its child")
			}
		}
		r.Frame()
		if p.recipe != recipe || p.popupError != "" {
			t.Fatalf("source recipe %s reached %s error=%s", recipe, p.recipe, p.popupError)
		}
	}
	if len(selectCompositionRecipes) != 22 || p.recipe != "demo" || access(ui.RoleComboBox, "Source fruit").Bounds.W != 192 {
		t.Fatal("source recipe count/default or docs192-DIP trigger changed")
	}
	click(ui.RoleComboBox, "Source fruit")
	access(ui.RoleGroup, "Fruits")
	click(ui.RoleOption, "Apple")
	if v := p.draft().choice("main").Value; !v.Valid || v.Value != "apple" || p.draft().changes != 1 || p.draft().choice("main").Open {
		t.Fatal("source pointer selection did not commit exactly once and close")
	}
	for _, recipe := range selectCompositionRecipes {
		choose(recipe.Value)
	}
	choose("align")
	if v := p.draft().choice("aligned-demo").Value; !v.Valid || v.Value != "banana" || !access(ui.RoleSwitch, "Align Item").Checked {
		t.Fatal("docs Align Item lost its Banana/true defaults")
	}
	click(ui.RoleSwitch, "Align Item")
	if p.draft().align {
		t.Fatal("the source Align Item switch did not update actual placement state")
	}
	click(ui.RoleComboBox, "Source fruit")
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p.draft().choice("aligned-demo").Value.Value != "pineapple" {
		t.Fatal("the real unaligned source listbox ignored native keyboard input")
	}
	choose("groups")
	click(ui.RoleComboBox, "Source fruit")
	access(ui.RoleGroup, "Fruits")
	access(ui.RoleGroup, "Vegetables")
	if _, ok := r.AccessByRole(ui.RoleOption, "Select a fruit"); ok {
		t.Fatal("source groups gained a placeholder row absent from their rendered collection")
	}
	click(ui.RoleOption, "Spinach")
	if p.draft().choice("main").Value.Value != "spinach" {
		t.Fatal("the vegetable group lost its real option")
	}
	choose("disabled")
	if !access(ui.RoleComboBox, "Source fruit").Disabled {
		t.Fatal("disabled source did not reach the native trigger")
	}
	click(ui.RoleComboBox, "Source fruit")
	if p.draft().choice("main").Open || p.draft().changes != 0 {
		t.Fatal("disabled source opened or committed a value")
	}
	choose("invalid")
	if a := access(ui.RoleComboBox, "Fruit"); !a.Invalid || a.Disabled || !strings.Contains(a.Description, "Please select a fruit.") {
		t.Fatal("source invalid Field lost native invalid/error association or blocked input")
	}
	click(ui.RoleComboBox, "Fruit")
	click(ui.RoleOption, "Banana")
	choose("create-aligned")
	click(ui.RoleComboBox, "Source fruit")
	if !access(ui.RoleOption, "Grapes").Disabled {
		t.Fatal("source item-aligned example lost its disabled Grapes item")
	}
	click(ui.RoleOption, "Grapes")
	if p.draft().choice("main").Value.Valid || !p.draft().choice("main").Open {
		t.Fatal("disabled Grapes committed or dismissed the source list")
	}
	choose("create-plan")
	if p.draft().choice("plan").Value.Value != "Starter" || !r.HasText("Perfect for individuals getting started.") || access(ui.RoleComboBox, "Subscription Plan").Bounds.W != 288 {
		t.Fatal("the source rich Plan value/default/width was collapsed")
	}
	click(ui.RoleComboBox, "Subscription Plan")
	r.Key(0, ui.KeyEnd)
	r.Frame()
	selectCompositionSnapshot(t, r, "select-plan-light")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p.draft().choice("plan").Value.Value != "Enterprise" || !r.HasText("Advanced features for large organizations.") {
		t.Fatal("source object identity and rich selected value diverged")
	}
	choose("create-multiple")
	click(ui.RoleComboBox, "Source fruits")
	click(ui.RoleOption, "Apple")
	click(ui.RoleOption, "Watermelon")
	if !slices.Equal(p.draft().choice("main").Values, []string{"apple", "watermelon"}) || !p.draft().choice("main").Open || !r.HasText("2 fruits selected") || p.draft().changes != 2 {
		t.Fatal("source multiple selection lost open state, final rows, count rendering or single callbacks")
	}
	selectCompositionSnapshot(t, r, "select-multiple-light")
	choose("create-large")
	click(ui.RoleComboBox, "Source fruit")
	r.Key(0, ui.KeyEnd)
	r.Frame()
	r.Frame()
	last := access(ui.RoleOption, "Item 99")
	viewport := access(ui.RoleListBox, "Source fruit options")
	if last.Offscreen || last.Bounds.Y < viewport.Bounds.Y || last.Bounds.Y+last.Bounds.H > viewport.Bounds.Y+viewport.Bounds.H {
		t.Fatal("the 100-item source collection did not reveal its native last row")
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p.draft().choice("main").Value.Value != "item-99" {
		t.Fatal("the full source 100-item collection was truncated or renamed")
	}
	choose("scrollable")
	dark = true
	r.Frame()
	click(ui.RoleComboBox, "Source timezone")
	r.Key(0, ui.KeyEnd)
	r.Frame()
	r.Frame()
	access(ui.RoleGroup, "South America")
	last = access(ui.RoleOption, "Chile Standard Time")
	viewport = access(ui.RoleListBox, "Source timezone options")
	if last.Offscreen || last.Bounds.Y < viewport.Bounds.Y || last.Bounds.Y+last.Bounds.H > viewport.Bounds.Y+viewport.Bounds.H {
		t.Fatalf("native timezone End did not reveal its full last row: item=%+v viewport=%+v", last, viewport)
	}
	selectCompositionSnapshot(t, r, "select-timezones-dark")
	r.ClickAt(last.Bounds.X+last.Bounds.W/2, last.Bounds.Y+last.Bounds.H/2)
	r.Frame()
	if p.draft().choice("main").Value.Value != "clt" || len(selectSourceTimezones) != 27 {
		t.Fatalf("source timezone pointer selection failed: last=%+v viewport=%+v value=%+v open=%v", last.Bounds, viewport.Bounds, p.draft().choice("main").Value, p.draft().choice("main").Open)
	}
	choose("create-icons")
	if access(ui.RoleComboBox, "Small chart type").Bounds.H >= access(ui.RoleComboBox, "Default chart type").Bounds.H {
		t.Fatal("the two original chart trigger sizes were collapsed")
	}
	click(ui.RoleComboBox, "Small chart type")
	click(ui.RoleOption, "Bar")
	if p.draft().choice("0").Value.Value != "bar" || p.draft().choice("1").Value.Valid {
		t.Fatal("source icon triggers shared their selection state")
	}
	choose("create-button")
	click(ui.RoleComboBox, "Small fruit")
	click(ui.RoleOption, "Blueberry")
	click(ui.RoleButton, "Submit small fruit")
	if p.draft().submits != 1 || !slices.Equal(p.draft().submitted, []string{"blueberry"}) || p.action != "Local select action: Submit" {
		t.Fatal("the real source button did not capture its local selection once")
	}
	choose("create-inline")
	click(ui.RoleTextField, "Source inline search")
	r.Type("editable inline draft")
	click(ui.RoleComboBox, "Source filter")
	click(ui.RoleOption, "Active")
	click(ui.RoleMenuButton, "Source sort order")
	if err := r.ChooseMenuItem("Date"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.draft().query != "editable inline draft" || p.draft().sort != "date" || p.draft().choice("main").Value.Value != "active" {
		t.Fatal("source inline input/filter/NativeSelect values were collapsed")
	}
	choose("create-dialog")
	click(ui.RoleButton, "Open Dialog")
	access(ui.RoleDialog, "Select Example")
	click(ui.RoleComboBox, "Dialog fruit")
	click(ui.RoleOption, "Pineapple")
	if p.draft().choice("main").Value.Value != "pineapple" || !p.draft().dialog {
		t.Fatal("select choice escaped or replaced the original Dialog")
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if p.draft().dialog {
		t.Fatal("the source Dialog refused its native Escape dismissal")
	}
	choose("rtl")
	click(ui.RoleMenuButton, "Select language")
	if err := r.ChooseMenuItem("עברית"); err != nil {
		t.Fatal(err)
	}
	r.SetSize(360, 900)
	r.Frame()
	if access(ui.RoleComboBox, "Source fruit").Bounds.W != 128 {
		t.Fatal("the translated source128-DIP trigger width was lost")
	}
	click(ui.RoleComboBox, "Source fruit")
	access(ui.RoleGroup, "פירות")
	access(ui.RoleGroup, "ירקות")
	selectCompositionSnapshot(t, r, "select-hebrew-dark")
	click(ui.RoleOption, "תרד")
	if p.draft().choice("main").Value.Value != "spinach" {
		t.Fatal("translated source labels lost stable selected identities")
	}
	visible = false
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleComboBox, "Source fruit"); ok {
		t.Fatal("leaving source examples retained native listboxes")
	}
	visible = true
	r.Frame()
	if p.draft().choice("main").Value.Value != "spinach" {
		t.Fatal("page return lost the independent translated selection")
	}
	r.SetSize(620, 900)
	native = false
	r.Frame()
	choose("demo")
	if p.draft().choice("main").Value.Value != "apple" || p.draft().changes != 1 || access(ui.RoleComboBox, "Source fruit").Value != "Apple" {
		t.Fatal("selector style/recipe changes crossed independent source drafts")
	}
}
