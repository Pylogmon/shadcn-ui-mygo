package main

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func exampleTabs(s *appState) *ui.Tester {
	return ui.NewTesterWith(s.view, 1000, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func TestFixedTabsRecipesRenderNativeSlots(t *testing.T) {
	if len(recipes) != 13 {
		t.Fatal("fixed collection changed")
	}
	for _, r := range recipes {
		t.Run(r.ID, func(t *testing.T) {
			s := newState(false, r.ID)
			view := exampleTabs(s)
			lists, tabs := 0, 0
			for _, a := range view.AccessAll() {
				if a.Role == ui.RoleTabList {
					lists++
					if a.Vertical != r.Vertical {
						t.Fatal("source orientation not native")
					}
				}
				if a.Role == ui.RoleTab {
					tabs++
					if a.Name == "" {
						t.Fatal("native icon trigger has no name")
					}
				}
			}
			wantLists, wantTabs := 1, len(r.Values)
			if r.Compare {
				wantLists, wantTabs = 2, 4
			}
			if lists != wantLists || tabs != wantTabs {
				t.Fatalf("lists%d tabs%d, source needs%d/%d", lists, tabs, wantLists, wantTabs)
			}
			if r.Content {
				tab, _ := view.AccessByRole(ui.RoleTab, label(r.Values[0]))
				if !slices.Equal(tab.Controls, []string{label(r.Values[0])}) {
					t.Fatal("source external panel not connected")
				}
			}
		})
	}
}
func TestTabsHeaderInputActionMenuAndPanelDraft(t *testing.T) {
	s := newState(false, "input")
	r := exampleTabs(s)
	if err := r.Click("Search tabs"); err != nil {
		t.Fatal(err)
	}
	r.Type("native query")
	r.Key(0, ui.KeyLeft)
	if s.query != "native query" || *s.values["input"] != "overview" {
		t.Fatal("tab list captured external editor")
	}
	if err := r.Click("Action"); err != nil {
		t.Fatal(err)
	}
	if s.status != "Action for overview: native query" {
		t.Fatal("action did not use actual editor state")
	}
	r.Click("Note for Overview")
	r.Type("A kept draft")
	r.Click("Analytics")
	if !r.HasText("Detailed analytics and insights about your data.") {
		t.Fatal("panel did not switch")
	}
	r.Click("Overview")
	note, ok := r.AccessByRole(ui.RoleTextField, "Note for Overview")
	if !ok || note.Value != "A kept draft" {
		t.Fatal("application panel draft was lost")
	}
	s.recipe = "dropdown"
	r.Frame()
	if err := r.Click("More options"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.AccessByRole(ui.RoleMenuItem, "Export"); !ok {
		t.Fatal("header menu was only a visual placeholder")
	}
	r.Click("Export")
	if s.status != "Menu export for overview" || s.menu || *s.values["dropdown"] != "overview" {
		t.Fatal("native menu action changed tab group or failed to close")
	}
}
