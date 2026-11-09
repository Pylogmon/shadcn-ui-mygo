package gallery

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func comboboxPreviewTester(s *State, width, height int) *ui.Tester {
	var otherDraft string
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		theme.Scope(c, s.Config.Tokens(false)).Fill().Padding(20).Gap(20).Children(func() {
			if s.Page == "controls" {
				s.comboboxPreview(c)
			} else {
				control.Input(c, &otherDraft, control.InputProps{Label: "Another page draft"})
			}
		})
	}, width, height)
}

func comboboxAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	// An open popup has a combobox/dialog trigger and a list-autocomplete
	// combobox editor. Keep input/focus assertions on the actual native editor.
	if role == ui.RoleComboBox {
		for _, info := range r.AccessAll() {
			if info.Role == role && info.Name == label && info.HasPopup == ui.PopupListBox {
				return info
			}
		}
	}
	a, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing native %v %q", role, label)
	}
	return a
}

func comboboxClick(t *testing.T, r *ui.Tester, role ui.Role, label string) {
	t.Helper()
	b := comboboxAccess(t, r, role, label).Bounds
	r.ClickAt(b.X+b.W/2, b.Y+b.H/2)
}

func comboboxFocused(r *ui.Tester) []string {
	var focused []string
	for _, info := range r.AccessAll() {
		if info.Focused {
			focused = append(focused, fmt.Sprintf("%d:%s", info.Role, info.Name))
		}
	}
	return focused
}

func TestCreatorComboboxQuerySelectionClearAndExplicitKeyboardChoice(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := comboboxPreviewTester(s, 1200, 1000)
	if !s.comboboxes.framework.Valid || s.comboboxes.framework.Value != "mygo" || comboboxAccess(t, r, ui.RoleComboBox, "Framework search").Value != "MyGo" {
		t.Fatal("the initial choice was not displayed as its label")
	}
	comboboxClick(t, r, ui.RoleComboBox, "Framework search")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Fy")
	if s.comboboxes.framework.Value != "mygo" || s.comboboxes.frameworkQuery != "Fy" {
		t.Fatal("typing changed the selected option instead of its search query")
	}
	r.Key(0, ui.KeyEnter)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.framework.Value != "mygo" || s.comboboxes.frameworkQuery != "MyGo" || comboboxAccess(t, r, ui.RoleComboBox, "Framework search").Expanded {
		t.Fatal("unhighlighted Enter did not close and restore the selected framework")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Fy")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.framework.Value != "fyne" || s.comboboxes.frameworkQuery != "Fyne" || !r.HasText("Selected: Fyne") {
		t.Fatalf("keyboard choice did not update the actual model and readable result: %+v query=%q", s.comboboxes.framework, s.comboboxes.frameworkQuery)
	}
	advancedClick(t, r, "Clear Framework search")
	if s.comboboxes.framework.Valid || s.comboboxes.frameworkQuery != "" || !r.Focused("Framework search") {
		t.Fatal("the real clear action did not empty the selection and preserve editor focus")
	}
}

func TestCreatorComboboxGroupsFilterCustomMetadataAndSkipDisabled(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := comboboxPreviewTester(s, 1200, 1000)
	comboboxClick(t, r, ui.RoleComboBox, "Destination country")
	r.Advance(200 * time.Millisecond)
	if !comboboxAccess(t, r, ui.RoleOption, "Norway").Disabled || !r.HasText("Asia") || !r.HasText("Europe") {
		t.Fatal("the grouped list omitted its headings or unavailable destination")
	}
	r.Type("Berlin")
	if !r.HasText("Berlin · Europe") || r.HasText("Tokyo · Asia") {
		t.Fatal("the custom option's city metadata was not actually searchable")
	}
	r.Key(0, ui.KeyEnter)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.country.Value != "de" || s.comboboxes.countryQuery != "Germany" {
		t.Fatal("custom option selection wrote a label in place of its stable value")
	}
	comboboxClick(t, r, ui.RoleComboBox, "Destination country")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("unavailable")
	if !comboboxAccess(t, r, ui.RoleOption, "Norway").Disabled || s.comboboxes.country.Value != "de" {
		t.Fatal("the unavailable search result lost its disabled state or changed the chosen country")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.country.Value != "de" || s.comboboxes.countryQuery != "Germany" || comboboxAccess(t, r, ui.RoleComboBox, "Destination country").Expanded {
		t.Fatal("unhighlighted Enter selected an unavailable country or failed to close and restore its label")
	}
}

func TestCreatorComboboxPopupSearchReturnsFocusAndResetsOnlyQuery(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := comboboxPreviewTester(s, 1200, 1000)
	comboboxClick(t, r, ui.RoleComboBox, "Office city")
	if trigger := comboboxAccess(t, r, ui.RoleComboBox, "Office city"); trigger.HasPopup != ui.PopupListBox {
		t.Fatal("the popup did not expose its list-autocomplete search editor")
	}
	r.Advance(200 * time.Millisecond)
	if !comboboxAccess(t, r, ui.RoleComboBox, "Office city").Focused {
		t.Fatal("the popup's actual native search editor did not receive focus")
	}
	r.Type("new")
	r.Key(0, ui.KeyEnter)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.city.Value != "new-york" || s.comboboxes.cityQuery != "" || !r.Focused("Office city") {
		t.Fatal("the popup choice did not reach the trigger and return its focus")
	}
	comboboxClick(t, r, ui.RoleComboBox, "Office city")
	r.Advance(200 * time.Millisecond)
	r.Type("tokyo")
	r.Key(0, ui.KeyEscape)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.city.Value != "new-york" || s.comboboxes.cityQuery != "" || !r.Focused("Office city") {
		t.Fatalf("Escape changed office/query/focus: selected=%+v query=%q focused=%v", s.comboboxes.city, s.comboboxes.cityQuery, comboboxFocused(r))
	}
}

func TestCreatorComboboxMultipleTogglesChoicesAndUsesLogicalChipKeys(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprintf("rtl=%v", rtl), func(t *testing.T) {
			s := New()
			s.Page = "controls"
			r := comboboxPreviewTester(s, 1200, 1000)
			if rtl {
				r.SetDirection(ui.RightToLeft)
			}
			comboboxClick(t, r, ui.RoleComboBox, "Project technologies")
			r.Advance(200 * time.Millisecond)
			if !comboboxAccess(t, r, ui.RoleOption, "Go").Selected {
				t.Fatal("selected technologies disappeared from the suggestion list")
			}
			comboboxClick(t, r, ui.RoleOption, "Rust")
			if !slices.Contains(s.comboboxes.technologies, "rust") || !comboboxAccess(t, r, ui.RoleComboBox, "Project technologies").Expanded {
				t.Fatal("an unfiltered choice did not keep the multi-select list open")
			}
			comboboxClick(t, r, ui.RoleOption, "Rust")
			if slices.Contains(s.comboboxes.technologies, "rust") {
				t.Fatalf("toggle retained selected=%v query=%q expanded=%v focused=%v", s.comboboxes.technologies, s.comboboxes.technologyQuery, comboboxAccess(t, r, ui.RoleComboBox, "Project technologies").Expanded, comboboxFocused(r))
			}
			r.Type("swi")
			r.Key(0, ui.KeyEnter)
			r.Advance(250 * time.Millisecond)
			if !slices.Equal(s.comboboxes.technologies, []string{"go", "typescript", "swift"}) || s.comboboxes.technologyQuery != "" {
				t.Fatalf("filtered selection did not update chips and reset the query: %v %q", s.comboboxes.technologies, s.comboboxes.technologyQuery)
			}
			r.Key(0, ui.KeyBackspace)
			if !slices.Equal(s.comboboxes.technologies, []string{"go", "typescript"}) {
				t.Fatal("empty-input Backspace did not remove the last real chip")
			}
			previous := ui.KeyLeft
			if rtl {
				previous = ui.KeyRight
			}
			r.Key(0, previous)
			r.Key(0, ui.KeyBackspace)
			if !slices.Equal(s.comboboxes.technologies, []string{"go"}) {
				t.Fatal("logical previous/deletion did not operate on the selected chip")
			}
			advancedClick(t, r, "Remove Go")
			if len(s.comboboxes.technologies) != 0 || !r.Focused("Project technologies") {
				t.Fatal("chip remove did not update the model and return editor focus")
			}
		})
	}
}

func TestCreatorComboboxNativeCompositionAndCollectionReorderKeepEditorIdentity(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := comboboxPreviewTester(s, 1200, 1000)
	comboboxClick(t, r, ui.RoleComboBox, "Framework search")
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("原", 1)
	r.Key(0, ui.KeyEnter)
	if s.comboboxes.framework.Value != "mygo" {
		t.Fatal("the native IME confirmation selected a suggestion")
	}
	r.Type("原生")
	r.Key(ui.Cmd, ui.KeyA)
	slices.Reverse(s.comboboxes.frameworks)
	r.Frame()
	r.Type("Gio")
	if s.comboboxes.frameworkQuery != "Gio" || !r.Focused("Framework search") || s.comboboxes.framework.Value != "mygo" {
		t.Fatal("reordering options lost the real native editor, selection or chosen value")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.framework.Value != "gio" {
		t.Fatal("the latest reordered collection did not supply the chosen value")
	}
}

func TestCreatorComboboxStatesEnforceEditingAndClearRequiredErrorAfterChoice(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := comboboxPreviewTester(s, 1200, 1000)
	if !comboboxAccess(t, r, ui.RoleComboBox, "Billing plan").ReadOnly || !comboboxAccess(t, r, ui.RoleComboBox, "Managed region").Disabled || !comboboxAccess(t, r, ui.RoleComboBox, "Required destination *").Invalid {
		t.Fatal("the variants did not reach the actual native editors")
	}
	comboboxClick(t, r, ui.RoleComboBox, "Billing plan")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Free")
	if s.comboboxes.plan.Value != "pro" || s.comboboxes.planQuery != "Pro" {
		t.Fatal("the read-only value was edited")
	}
	comboboxClick(t, r, ui.RoleComboBox, "Required destination *")
	r.Type("Japan")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(250 * time.Millisecond)
	if s.comboboxes.requiredCountry.Value != "jp" || comboboxAccess(t, r, ui.RoleComboBox, "Required destination *").Invalid || r.HasText("Choose a destination.") {
		t.Fatal("a valid required choice did not clear the real field error")
	}
}

func TestCreatorComboboxLeavingPageClosesSearchWithoutStealingNewFocus(t *testing.T) {
	s := New()
	s.Page = "controls"
	r := comboboxPreviewTester(s, 1200, 1000)
	comboboxClick(t, r, ui.RoleComboBox, "Office city")
	r.Advance(200 * time.Millisecond)
	r.Type("San")
	if s.comboboxes.cityQuery != "San" || !r.Focused("Office city") {
		t.Fatalf("popup before navigation: query=%q focused=%v", s.comboboxes.cityQuery, comboboxFocused(r))
	}
	s.Page = "overview"
	r.Frame()
	advancedClick(t, r, "Another page draft")
	r.Type("Keep this focus")
	r.Advance(time.Second)
	if !r.Focused("Another page draft") || r.HasText("San Francisco") {
		t.Fatal("the removed popup retained content or stole the new page's focus")
	}
	s.Page = "controls"
	r.Frame()
	trigger := comboboxAccess(t, r, ui.RoleComboBox, "Office city")
	if trigger.HasPopup != ui.PopupDialog || trigger.Expanded || s.comboboxes.framework.Value != "mygo" || !slices.Equal(s.comboboxes.technologies, []string{"go", "typescript"}) {
		t.Fatal("returning to Forms reopened a stale popup or lost selected application state")
	}
	if _, ok := r.AccessByRole(ui.RoleDialog, "Office city popup"); ok {
		t.Fatal("returning to Forms retained a stale popup dialog")
	}
}

func comboboxReveal(t *testing.T, r *ui.Tester, label string) ui.Rect {
	t.Helper()
	for attempt := 0; attempt < 6; attempt++ {
		role := ui.RoleComboBox
		info, ok := r.AccessByRole(role, label)
		if !ok {
			t.Fatalf("missing actual Forms control %q", label)
		}
		target := info.Bounds
		panel := comboboxAccess(t, r, ui.RoleRegion, "Forms & controls")
		if target.Y >= panel.Bounds.Y && target.Y+target.H <= panel.Bounds.Y+panel.Bounds.H {
			return target
		}
		r.Scroll(panel.Bounds.X+panel.Bounds.W-12, panel.Bounds.Y+panel.Bounds.H/2, 0, target.Y+target.H/2-panel.Bounds.Y-panel.Bounds.H/2)
	}
	t.Fatalf("native Forms scroll could not reveal %q", label)
	return ui.Rect{}
}

func TestCreatorComboboxWholePageNarrowRTLUsesNaturalCardsAndReachableControls(t *testing.T) {
	for _, width := range []int{500, 800, 1450} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			s := New()
			s.Page, s.Config.ReduceMotion = "controls", true
			r := ui.NewTester(s.View, width, 900)
			r.SetDirection(ui.RightToLeft)
			for _, label := range []string{"Framework search", "Destination country", "Office city", "Project technologies", "Required destination *"} {
				bounds := comboboxReveal(t, r, label)
				panel := comboboxAccess(t, r, ui.RoleRegion, "Forms & controls")
				if bounds.X < panel.Bounds.X || bounds.X+bounds.W > panel.Bounds.X+panel.Bounds.W+1 || bounds.H < 20 {
					t.Fatalf("a real combobox editor overflowed or compressed in the narrow RTL page: %q %v inside %v", label, bounds, panel.Bounds)
				}
			}
			comboboxReveal(t, r, "Project technologies")
			comboboxClick(t, r, ui.RoleComboBox, "Project technologies")
			r.Type("rust")
			r.Key(0, ui.KeyEnter)
			if !slices.Contains(s.comboboxes.technologies, "rust") {
				t.Fatal("the whole-page narrow editor could not choose a technology")
			}
			framework := comboboxAccess(t, r, ui.RoleGroup, "Framework selection").Bounds
			destination := comboboxAccess(t, r, ui.RoleGroup, "Destination selection").Bounds
			if destination.Y < framework.Y+framework.H+10 {
				t.Fatal("stacked cards overlap instead of preserving natural content height")
			}
		})
	}
}

func TestCreatorComboboxIsReachableByComponentSearch(t *testing.T) {
	for _, query := range []string{"Combobox", "MultiCombobox", "Autocomplete"} {
		t.Run(query, func(t *testing.T) {
			s := New()
			r := ui.NewTester(s.View, 1280, 900)
			advancedClick(t, r, "Search components")
			r.Advance(250 * time.Millisecond)
			r.Type(query)
			r.Key(0, ui.KeyEnter)
			r.Advance(250 * time.Millisecond)
			if s.Page != "controls" || s.Command || !r.HasText("Grouped suggestions") || !r.HasText("Multiple selection") {
				t.Fatal("the component search did not reach its actual Forms examples")
			}
		})
	}
}
