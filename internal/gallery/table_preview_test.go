package gallery

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
)

func tablePreviewTester(s *State, backend *creatorPopupBackend, width int) *ui.Tester {
	s.Popovers = desktop.NewHostWithBackend(backend)
	return ui.NewTester(func(c *ui.Context) {
		s.Theme.Apply(c)
		c.DisableAnimations()
		s.Popovers.BeginFrame()
		defer s.Popovers.EndFrame()
		ui.Column(c).Fill().Padding(20).Children(func() {
			if s.Page == "display" {
				s.tableCompositionPreview(c)
			} else {
				ui.Text(c, "Another page")
			}
		})
	}, width, 750)
}

func TestCreatorTableNativeSelectionEditingAndKeyedReorder(t *testing.T) {
	s := New()
	s.Page = "display"
	r := tablePreviewTester(s, &creatorPopupBackend{}, 1300)
	advancedAccess(t, r, ui.RoleTable, "Editable invoices")
	advancedClick(t, r, "Select INV001")
	if !s.tablePreview.records["INV001"].selected {
		t.Fatal("checkbox did not update application selection")
	}
	selected := false
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleRow && node.RowIndex == 2 {
			selected = node.Selected
		}
	}
	if !selected {
		t.Fatal("selected row did not expose native selection")
	}
	advancedClick(t, r, "Note INV001")
	r.Type("Keep this record")
	r.Key(ui.Cmd, ui.KeyA)
	slices.Reverse(s.tablePreview.order)
	r.Frame()
	r.Type("Updated after reorder")
	if s.tablePreview.records["INV001"].note != "Updated after reorder" || s.tablePreview.records["INV004"].note != "" {
		t.Fatal("reordering redirected native editor focus or selection to another record")
	}
	s.Page = "overview"
	r.Frame()
	s.Page = "display"
	r.Frame()
	if advancedAccess(t, r, ui.RoleTextField, "Note INV001").Value != "Updated after reorder" || !advancedAccess(t, r, ui.RoleCheckBox, "Select INV001").Checked {
		t.Fatal("leaving the preview lost app-owned note/selection")
	}
}

func TestCreatorTableStatusAndActionsUseSelectedMenuFamily(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(fmt.Sprint(native), func(t *testing.T) {
			s := New()
			s.Page = "display"
			s.Config.NativeMenus = native
			backend := &creatorPopupBackend{}
			r := tablePreviewTester(s, backend, 1300)
			choose := func(trigger, choice string) {
				t.Helper()
				advancedClick(t, r, trigger)
				if native {
					if len(backend.surfaces) != 0 || !slices.Contains(r.Menu(), choice) {
						t.Fatal("table did not open an OS menu")
					}
					if err := r.ChooseMenuItem(choice); err != nil {
						t.Fatal(err)
					}
				} else {
					popup := backend.surfaces[len(backend.surfaces)-1]
					if popup.spec.SystemPopover {
						t.Fatal("themed table menu did not use its independent Flyout")
					}
					for _, node := range r.AccessAll() {
						if node.Role == ui.RoleMenuItem || node.Role == ui.RoleMenuItemRadio {
							t.Fatal("table popup items leaked into the parent window")
						}
					}
					popup.tester.Advance(time.Second)
					advancedClick(t, popup.tester, choice)
					r.Frame()
				}
			}
			choose("Status INV001", "Processing")
			if s.tablePreview.records["INV001"].status != "Processing" {
				t.Fatal("selected status did not reach the actual row model")
			}
			choose("Actions INV001", "View invoice")
			if s.pendingMenuAction != "View invoice INV001" {
				t.Fatalf("row action targeted wrong record: %q", s.pendingMenuAction)
			}
		})
	}
}

func TestCreatorTableEmptyFooterAndHorizontalViewport(t *testing.T) {
	s := New()
	s.Page = "display"
	r := tablePreviewTester(s, &creatorPopupBackend{}, 620)
	if s.tablePreview.scroll.MaxX < 200 {
		t.Fatal("wide native cells did not create a horizontal viewport")
	}
	s.tablePreview.scroll.ScrollTo(s.tablePreview.scroll.MaxX, 0, ui.Animation{})
	r.Frame()
	if advancedAccess(t, r, ui.RoleButton, "Actions INV001").Bounds.X >= 620 {
		t.Fatal("scroll did not reveal actual trailing row action")
	}
	advancedClick(t, r, "Show empty table")
	if !r.HasText("No invoices to display.") || !r.HasText("$0.00") || !r.HasText("A list of recent invoices.") {
		t.Fatal("empty composition omitted footer or caption")
	}
	emptyCell := false
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleCell && node.Name == "No invoices to display." {
			emptyCell = true
			if node.ColumnSpan != 6 {
				t.Fatal("empty state did not span every column")
			}
		}
	}
	if !emptyCell {
		t.Fatal("empty state lost its native table cell semantics")
	}
	advancedClick(t, r, "Show empty table")
	if !r.HasText("$1200.00") || r.HasText("No invoices to display.") {
		t.Fatal("restoring rows did not restore the real total")
	}
}

func TestCreatorTableIsReachableByComponentSearch(t *testing.T) {
	s := New()
	r := ui.NewTester(s.View, 1280, 900)
	advancedClick(t, r, "Search components")
	r.Advance(200 * time.Millisecond)
	r.Type("Table")
	r.Key(0, ui.KeyEnter)
	r.Advance(200 * time.Millisecond)
	if s.Page != "display" || s.Command || !r.HasText("Table composition") {
		t.Fatal("Table search did not reach its real Gallery preview")
	}
}
