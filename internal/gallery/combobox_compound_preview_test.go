package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func compoundPreviewTester(s *State) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		theme.Scope(c, s.Config.Tokens(false)).Fill().Padding(20).Gap(20).Children(func() { s.comboboxCompoundPreview(c) })
	}, 1200, 1600)
}
func compoundOpen(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	comboboxClick(t, r, ui.RoleComboBox, label)
	r.Advance(120 * time.Millisecond)
	r.Frame()
}
func TestCreatorCompoundContentExplorerAndNativeChoice(t *testing.T) {
	s := New()
	r := compoundPreviewTester(s)
	for _, label := range []string{"Right to left", "12 DIP gap", "Icon addon"} {
		if err := r.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	p := &s.comboboxCompound
	if !p.rtl || !p.extraGap || !p.iconAddon {
		t.Fatal("placement configuration switches did not update")
	}
	compoundOpen(t, r, "Placement framework")
	r.Type("Remi")
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if !p.layout.Value.Valid || p.layout.Value.Value != "Remix" || p.layout.Open {
		t.Fatalf("configured native Content choice=%+v", p.layout)
	}
}
func TestCreatorCompoundObjectChipsPopupFormAndPermissions(t *testing.T) {
	s := New()
	r := compoundPreviewTester(s)
	p := &s.comboboxCompound
	compoundOpen(t, r, "Trip destinations")
	r.Type("Germany")
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if len(p.destination.Values) != 2 || p.destination.Values[1].Code != "de" {
		t.Fatalf("typed chips selected %+v", p.destination.Values)
	}
	if err := r.Click("Submit trip"); err != nil {
		t.Fatal(err)
	}
	if p.submitted != "Trip saved: Japan, Germany" {
		t.Fatalf("real form result=%q", p.submitted)
	}
	if err := r.Click("Hide chip remove buttons"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Remove Japan"); ok {
		t.Fatal("no-remove configuration left a remove action")
	}
	if err := r.Click("Show destination error"); err != nil {
		t.Fatal(err)
	}
	editor := comboboxAccess(t, r, ui.RoleComboBox, "Trip destinations")
	if !editor.Invalid {
		t.Fatal("Field error did not reach actual editor")
	}
	if err := r.Click("Read only destination"); err != nil {
		t.Fatal(err)
	}
	editor = comboboxAccess(t, r, ui.RoleComboBox, "Trip destinations")
	if !editor.ReadOnly {
		t.Fatal("readonly configuration did not reach native editor")
	}
	if err := r.Click("Read only destination"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Disable destination field"); err != nil {
		t.Fatal(err)
	}
	if !comboboxAccess(t, r, ui.RoleComboBox, "Trip destinations").Disabled {
		t.Fatal("multiple disabled configuration did not disable actual editor")
	}
	if err := r.Click("Disable destination field"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Group separators"); err != nil {
		t.Fatal(err)
	}
	if err := r.Click("Search in dialog popup"); err != nil {
		t.Fatal(err)
	}
	compoundOpen(t, r, "Trip destinations")
	r.Type("Norway")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if len(p.destination.Values) != 2 {
		t.Fatalf("disabled destination chosen: %+v", p.destination.Values)
	}
}
func TestCreatorCompoundGridVirtualCollectionAndDialog(t *testing.T) {
	s := New()
	r := compoundPreviewTester(s)
	p := &s.comboboxCompound
	compoundOpen(t, r, "Grid framework")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyRight)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if !p.grid.Value.Valid || p.grid.Value.Value != "Remix" {
		t.Fatalf("grid navigation=%+v", p.grid)
	}
	compoundOpen(t, r, "Large collection search")
	r.Type("099")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if !p.virtual.Value.Valid || p.virtual.Value.Value != "Item 099" {
		t.Fatalf("virtual native choice=%+v", p.virtual)
	}
	if err := r.Click("Choose framework in dialog"); err != nil {
		t.Fatal(err)
	}
	r.Advance(150 * time.Millisecond)
	r.Frame()
	compoundOpen(t, r, "Dialog framework")
	r.Type("Astro")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if !p.dialog || !p.modal.Value.Valid || p.modal.Value.Value != "Astro" {
		t.Fatalf("nested dialog choice=%+v dialog=%v", p.modal, p.dialog)
	}
	if err := r.Click("Confirm framework"); err != nil {
		t.Fatal(err)
	}
	r.Advance(150 * time.Millisecond)
	r.Frame()
	if p.dialog {
		t.Fatal("dialog confirm did not close")
	}
}
