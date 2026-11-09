package gallery

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestCreatorPrimitiveCompositionsShareApplicationModels(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(20).Children(func() {
			if s.Page == "controls" {
				s.choiceVariantsPreview(c)
			} else {
				s.composedProjectTabs(c)
			}
		})
	}, 1000, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	advancedClick(t, r, "Italic")
	advancedClick(t, r, "Selected formatting")
	list := advancedAccess(t, r, ui.RoleListBox, "")
	if !list.Multiselectable {
		t.Fatal("formatting menu was not an actual no-editor multiple listbox")
	}
	option := advancedAccess(t, r, ui.RoleOption, "Underline")
	r.ClickAt(option.Bounds.X+option.Bounds.W/2, option.Bounds.Y+option.Bounds.H/2)
	r.Frame()
	if !slices.Equal(s.formatting, []string{"bold", "italic", "underline"}) || !r.HasText("3 styles selected") {
		t.Fatalf("multiple selection did not share the formatting model: %v", s.formatting)
	}
	r.Key(0, ui.KeyEscape)
	if !advancedAccess(t, r, ui.RoleToggleButton, "Underline").Checked {
		t.Fatal("Select did not update the corresponding toggle")
	}
	s.Page = "navigation"
	r.Frame()
	advancedClick(t, r, "Activity")
	if s.lineTab != 1 || !r.HasText("Your latest project updates.") {
		t.Fatal("composed tabs did not update the existing application selection")
	}
	tab := advancedAccess(t, r, ui.RoleTab, "Activity")
	panel := advancedAccess(t, r, ui.RoleRegion, "Activity")
	if !slices.Equal(tab.Controls, []string{panel.Name}) {
		t.Fatal("external panel lost its native tab relation")
	}
	advancedClick(t, r, "Show current view")
	if s.status != "Current view: Activity" || s.lineTab != 1 {
		t.Fatal("header action did not use the actual selected view")
	}
	s.Page = "controls"
	r.Frame()
	if !r.HasText("3 styles selected") || !advancedAccess(t, r, ui.RoleToggleButton, "Underline").Checked {
		t.Fatal("leaving the preview discarded caller-owned formatting values")
	}
	s.Page = "navigation"
	r.Frame()
	if !advancedAccess(t, r, ui.RoleTab, "Activity").Selected || !r.HasText("Your latest project updates.") {
		t.Fatal("leaving the preview discarded the current tab")
	}
}

func TestCreatorComposedTabsFitNarrowAndPreserveNativeKeys(t *testing.T) {
	s := New()
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() { s.composedProjectTabs(c) })
	}, 360, 280, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	advancedClick(t, r, "Overview")
	r.Key(0, ui.KeyRight)
	if s.lineTab != 1 || !advancedAccess(t, r, ui.RoleTab, "Activity").Selected {
		t.Fatal("independent list did not use native arrow navigation")
	}
	for _, roleAndName := range []struct {
		role ui.Role
		name string
	}{{ui.RoleTabList, "Project view"}, {ui.RoleButton, "Show current view"}, {ui.RoleRegion, "Activity"}} {
		bounds := advancedAccess(t, r, roleAndName.role, roleAndName.name).Bounds
		if bounds.X < 19.5 || bounds.X+bounds.W > 340.5 {
			t.Fatalf("narrow composition escaped its parent: %s %+v", roleAndName.name, bounds)
		}
	}
}
