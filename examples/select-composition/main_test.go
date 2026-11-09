package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func exampleTester(s *appState) *ui.Tester {
	s.theme.ReduceMotion = true
	return ui.NewTesterWith(s.view, 1050, 1000, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func exampleAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(role, name)
	if !ok || n.Offscreen {
		names := []string{}
		for _, n := range r.AccessAll() {
			names = append(names, n.Name)
		}
		t.Fatalf("missing native %v %q: %v", role, name, names)
	}
	return n
}
func exampleClick(t *testing.T, r *ui.Tester, role ui.Role, name string) {
	t.Helper()
	n := exampleAccess(t, r, role, name)
	r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
	r.Frame()
}
func recipe(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	exampleClick(t, r, ui.RoleMenuButton, "Recipe")
	if err := r.ChooseMenuItem(label); err != nil {
		t.Fatal(err)
	}
	r.Frame()
}
func choose(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	exampleClick(t, r, ui.RoleOption, label)
}

func TestSelectWorkbenchBasicNullGroupsMultipleAndForm(t *testing.T) {
	s := newState()
	r := exampleTester(s)
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	choose(t, r, "Apple")
	if !s.selection.Value.Valid || s.selection.Value.Value != "apple" || s.selection.Open {
		t.Fatalf("basic %+v", s.selection)
	}
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.selection.Value.Valid {
		t.Fatal("source null item did not clear")
	}
	recipe(t, r, "Groups")
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	choose(t, r, "Carrot")
	if s.selection.Value.Value != "carrot" {
		t.Fatalf("group %+v", s.selection)
	}
	recipe(t, r, "Multiple")
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	choose(t, r, "Apple")
	choose(t, r, "Banana")
	if !s.selection.Open || !slices.Equal(s.selection.Values, []string{"apple", "banana"}) || !r.HasText("2 fruits selected") {
		t.Fatalf("multiple %+v", s.selection)
	}
	r.Key(0, ui.KeyEscape)
	submit := exampleAccess(t, r, ui.RoleButton, "Submit selection")
	trigger := exampleAccess(t, r, ui.RoleComboBox, "Your selection")
	if submit.Bounds.Y >= trigger.Bounds.Y+trigger.Bounds.H {
		t.Fatal("submit button was not adjacent to the select")
	}
	exampleClick(t, r, ui.RoleButton, "Submit selection")
	if s.submitted != "Submitted: apple, banana" {
		t.Fatalf("form submission %q", s.submitted)
	}
}

func TestSelectWorkbenchRichPlanLargeListAndConfiguration(t *testing.T) {
	s := newState()
	r := exampleTester(s)
	recipe(t, r, "Subscription plan")
	if !r.HasText("Perfect for individuals getting started.") {
		t.Fatal("rich Plan value not rendered")
	}
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	choose(t, r, "Enterprise")
	if s.selection.Value.Value != "Enterprise" || !r.HasText("Advanced features for large organizations.") {
		t.Fatalf("rich Plan selected %+v", s.selection)
	}
	recipe(t, r, "Large list")
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	r.Key(0, ui.KeyEnd)
	r.Frame()
	last := exampleAccess(t, r, ui.RoleOption, "Item 99")
	if last.Bounds.H < 20 {
		t.Fatalf("long collection did not reveal %+v", last.Bounds)
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.selection.Value.Value != "99" {
		t.Fatalf("large list %+v", s.selection)
	}
	exampleClick(t, r, ui.RoleSwitch, "Align item")
	exampleClick(t, r, ui.RoleSwitch, "Right to left")
	exampleClick(t, r, ui.RoleSwitch, "Small trigger")
	exampleClick(t, r, ui.RoleMenuButton, "List side")
	if err := r.ChooseMenuItem("Inline end"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	exampleClick(t, r, ui.RoleCheckBox, "Inline field")
	if !s.inline {
		t.Fatal("inline field flow did not activate")
	}
	exampleClick(t, r, ui.RoleCheckBox, "Icons")
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	trigger := exampleAccess(t, r, ui.RoleComboBox, "Your selection")
	list := exampleAccess(t, r, ui.RoleListBox, "")
	if trigger.Bounds.H != 32 || list.Bounds.X+list.Bounds.W > trigger.Bounds.X+1 {
		t.Fatalf("small logical placement: trigger%+v list%+v", trigger.Bounds, list.Bounds)
	}
	r.Key(0, ui.KeyEscape)
	exampleClick(t, r, ui.RoleCheckBox, "Disabled")
	if !exampleAccess(t, r, ui.RoleComboBox, "Your selection").Disabled {
		t.Fatal("disabled configuration missing")
	}
	exampleClick(t, r, ui.RoleCheckBox, "Disabled")
	exampleClick(t, r, ui.RoleCheckBox, "Invalid")
	if !exampleAccess(t, r, ui.RoleComboBox, "Your selection").Invalid {
		t.Fatal("invalid Field configuration missing")
	}
	recipe(t, r, "Aligned disabled item")
	exampleClick(t, r, ui.RoleComboBox, "Your selection")
	if !exampleAccess(t, r, ui.RoleOption, "Grapes").Disabled || !exampleAccess(t, r, ui.RoleOption, "Banana").Focused {
		t.Fatal("aligned selected/disabled source recipe did not apply")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.selection.Value.Value != "banana" || !s.selection.Open {
		t.Fatal("disabled item activated in aligned recipe")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.selection.Value.Value != "pineapple" {
		t.Fatalf("aligned source recipe value %+v", s.selection.Value)
	}

}

func TestSelectCompositionManualDialogAndSystemBoundary(t *testing.T) {
	s := newState()
	r := exampleTester(s)
	exampleClick(t, r, ui.RoleComboBox, "Workflow stages")
	choose(t, r, "Draft")
	choose(t, r, "Published")
	if !slices.Equal(s.manual.Values, []string{"review", "draft", "published"}) {
		t.Fatalf("manual slots %+v", s.manual)
	}
	r.Key(0, ui.KeyEscape)
	exampleClick(t, r, ui.RoleButton, "Choose in dialog")
	exampleClick(t, r, ui.RoleComboBox, "Dialog fruit")
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.dialogSelection.Value.Value != "pineapple" || !s.dialog {
		t.Fatalf("nested native popup %+v", s.dialogSelection)
	}
	exampleClick(t, r, ui.RoleButton, "Done")
	if s.dialog {
		t.Fatal("dialog did not close")
	}
	exampleClick(t, r, ui.RoleMenuButton, "System fruit")
	if !slices.Contains(r.Menu(), "Apple") {
		t.Fatalf("missing OS menu: %v", r.Menu())
	}
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleOption && strings.Contains(n.Name, "Apple") {
			t.Fatal("OS menu option painted inside parent")
		}
	}
	if err := r.ChooseMenuItem("Apple"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if s.native != "apple" {
		t.Fatalf("OS binding %q", s.native)
	}
}
