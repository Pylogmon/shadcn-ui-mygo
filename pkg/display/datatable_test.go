package display

import (
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestDataTableSelectionPersistsAcrossPagesAndSort(t *testing.T) {
	p := records()
	p.Selectable, p.Searchable = true, true
	var actionRecords []testRecord
	p.SelectionActions = func(selected []testRecord) { actionRecords = selected }
	state := DataTableState{PageSize: 2}
	tt := displayTester(func(c *ui.Context) { DataTable(c, &state, p) })
	displayClick(t, tt, "Select row 1")
	if !state.Selection.Has(1) || state.Selection.Len() != 1 {
		t.Fatalf("checkbox selection=%v", slices.Collect(state.Selection.All()))
	}
	checkbox, ok := tt.AccessByRole(ui.RoleCheckBox, "Select row 1")
	if !ok || !checkbox.Checked {
		t.Fatalf("row checkbox did not expose selection: %+v", checkbox)
	}
	tt.Key(0, ui.KeySpace)
	if state.Selection.Has(1) {
		t.Fatal("keyboard Space did not clear the focused row checkbox")
	}
	tt.Key(0, ui.KeySpace)
	if !state.Selection.Has(1) {
		t.Fatal("keyboard Space did not select the focused row checkbox")
	}
	displayClick(t, tt, "Next")
	displayClick(t, tt, "Select row 4")
	if !state.Selection.Has(1) || !state.Selection.Has(4) || len(actionRecords) != 2 {
		t.Fatalf("page selection=%v action=%v", slices.Collect(state.Selection.All()), actionRecords)
	}
	displayClick(t, tt, "Size")
	if !state.Selection.Has(1) || !state.Selection.Has(4) {
		t.Fatal("sorting lost keyed selections")
	}
	state.Query = "Cat"
	tt.Frame()
	if !tt.HasText("1 of 1 row(s) selected.") {
		t.Fatalf("filtered selection count=%v", tt.Texts())
	}
	displayClick(t, tt, "Clear selection")
	if state.Selection.Len() != 0 || len(actionRecords) != 0 {
		t.Fatalf("clear selection=%d action=%v", state.Selection.Len(), actionRecords)
	}
}

func TestDataTableColumnVisibilityAndCustomRowAction(t *testing.T) {
	p := records()
	p.ShowColumnToggle = true
	clicked := 0
	p.Columns[0].Cell = func(c *ui.Context, row testRecord) {
		if ui.ButtonBase(c).Label(fmt.Sprintf("Inspect %d", row.ID)).Children(func() { ui.Text(c, row.Name) }).Clicked() {
			clicked = row.ID
		}
	}
	state := DataTableState{}
	tt := displayTester(func(c *ui.Context) { DataTable(c, &state, p) })
	displayClick(t, tt, "Inspect 2")
	if clicked != 2 {
		t.Fatal("custom cell action did not run")
	}
	displayClick(t, tt, "Columns")
	if err := tt.ChooseMenuItem("Size"); err != nil {
		t.Fatal(err)
	}
	if !state.Hidden["size"] {
		t.Fatal("column popup did not hide Size")
	}
	tt.Key(0, ui.KeyEscape)
	if tt.HasText("Size") {
		t.Fatalf("hidden column still rendered: %v", tt.Texts())
	}
	state.SetColumnVisible("name", false)
	tt.Frame()
	if !tt.HasText("No visible columns.") {
		t.Fatalf("all-hidden state %v", tt.Texts())
	}
	state.SetColumnVisible("size", true)
	tt.Frame()
	if !tt.HasText("Size") {
		t.Fatal("column did not return")
	}
}

func TestDataTableFacetFiltersComposeAndCountAlternatives(t *testing.T) {
	p := records()
	state := DataTableState{Page: 1, PageSize: 2}
	state.SetFilter("name", []string{"Bee"})
	rows, total := DataRows(&state, p)
	if total != 2 || state.Page != 0 || len(rows) != 2 {
		t.Fatalf("facet result=%v total=%d page=%d", rows, total, state.Page)
	}
	facets := DataFacetValues(&state, p, "name")
	if !slices.Equal(facets, []DataFacet{{"Ant", 1}, {"Bee", 2}, {"Cat", 1}}) {
		t.Fatalf("facet counts=%+v", facets)
	}
	state.SetFilter("size", []string{"3"})
	rows, total = DataRows(&state, p)
	if total != 1 || rows[0].ID != 3 {
		t.Fatalf("AND filters=%v total=%d", rows, total)
	}
	state.Query = "Bee"
	facets = DataFacetValues(&state, p, "size")
	if !slices.Equal(facets, []DataFacet{{"2", 1}, {"3", 1}}) {
		t.Fatalf("global-filtered alternatives=%+v", facets)
	}
	state.SetFilter("size", nil)
	tt := displayTester(func(c *ui.Context) {
		DataTableFacetFilter(c, &state, "size", "Sizes", DataFacetValues(&state, p, "size"))
		DataTable(c, &state, p)
	})
	displayClick(t, tt, "Sizes")
	if err := tt.ChooseMenuItem("3"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(state.Filters["size"], []string{"3"}) {
		t.Fatalf("popup facet filter=%v", state.Filters)
	}
	rows, total = DataRows(&state, p)
	if total != 1 || rows[0].ID != 3 {
		t.Fatalf("popup did not filter=%v", rows)
	}
}

func TestDataTableSearchInputAndFallbackSelectionIdentity(t *testing.T) {
	p := records()
	p.Key = nil
	p.Selectable = true
	p.Searchable = true
	state := DataTableState{PageSize: 2}
	tt := displayTester(func(c *ui.Context) { DataTable(c, &state, p) })
	displayClick(t, tt, "Select row 0")
	displayClick(t, tt, "Next")
	displayClick(t, tt, "Select page")
	if !state.Selection.Has(0) || !state.Selection.Has(2) || !state.Selection.Has(3) || state.Selection.Has(1) {
		t.Fatalf("fallback keys=%v", slices.Collect(state.Selection.All()))
	}
	displayClick(t, tt, "Filter rows")
	tt.Type("Ant")
	if state.Page != 0 || !tt.HasText("Ant") || tt.HasText("Cat") {
		t.Fatalf("search page=%d texts=%v", state.Page, tt.Texts())
	}
	if !state.Selection.Has(0) || !state.Selection.Has(3) {
		t.Fatal("filtering lost nonvisible source selections")
	}
}
