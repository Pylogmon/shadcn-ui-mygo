package overlay

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func commandSelected(r *ui.Tester) (ui.AccessInfo, bool) {
	for _, info := range r.AccessAll() {
		if info.Role == ui.RoleOption && info.Selected {
			return info, true
		}
	}
	return ui.AccessInfo{}, false
}

func requireCommandHighlight(t *testing.T, r *ui.Tester, name string) ui.AccessInfo {
	t.Helper()
	selected, ok := commandSelected(r)
	if !ok || selected.Name != name || selected.Disabled {
		t.Fatalf("highlight = %+v, want enabled %q", selected, name)
	}
	count := 0
	for _, info := range r.AccessAll() {
		if info.Role == ui.RoleOption && info.Selected {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("%d selected options", count)
	}
	return selected
}

func TestCommandDynamicDisabledResultCannotStayHighlighted(t *testing.T) {
	query := ""
	items := []CommandItem{{Value: "first", Label: "First"}, {Value: "second", Label: "Second"}}
	r := ui.NewTester(view(func(c *ui.Context) {
		Command(c, &query, CommandProps{AutoFocus: true, Items: items})
	}), 480, 400)
	items[0].Disabled = true
	r.Frame()
	selected, ok := commandSelected(r)
	if !ok || selected.Name != "Second" || selected.Disabled {
		t.Fatalf("highlight did not move off the disabled result: %+v", selected)
	}
	items[1].Disabled = true
	r.Frame()
	r.Key(0, ui.KeyDown)
	if selected, ok := commandSelected(r); ok {
		t.Fatalf("all-disabled list selected an option: %+v", selected)
	}
}

func TestCommandEndScrollsHighlightedResultIntoView(t *testing.T) {
	query, selected := "", ""
	items := make([]CommandItem, 300)
	for i := range items {
		items[i] = CommandItem{Value: fmt.Sprint(i), Label: fmt.Sprintf("Command %03d", i)}
	}
	r := ui.NewTester(view(func(c *ui.Context) {
		if value := Command(c, &query, CommandProps{AutoFocus: true, Width: 360, Items: items}); value != "" {
			selected = value
		}
	}), 480, 400)
	r.Key(0, ui.KeyEnd)
	last, ok := commandSelected(r)
	if !ok || last.Name != "Command 299" || last.Bounds.Y < 40 || last.Bounds.Y+last.Bounds.H > 400 {
		t.Fatalf("End did not expose the last result: %+v", last)
	}
	r.Key(0, ui.KeyEnter)
	if selected != "299" {
		t.Fatalf("End/Enter chose %q", selected)
	}
}

func TestCommandTenThousandRowsAreVirtualAndFocusTracksVisibleOption(t *testing.T) {
	query := ""
	items := make([]CommandItem, 10000)
	for i := range items {
		items[i] = CommandItem{Value: fmt.Sprint(i), Label: fmt.Sprintf("Command %05d", i)}
	}
	items[len(items)-1].Keywords = []string{"unique last"}
	built := 0
	r := ui.NewTester(view(func(c *ui.Context) {
		Command(c, &query, CommandProps{Items: items, Width: 360, AutoFocus: true, ItemContent: func(row *ui.Context, item CommandItem, active bool) {
			built++
			ui.Text(row, item.Label)
		}})
	}), 480, 400)
	built = 0
	r.Frame()
	if built <= 0 || built > 50 {
		t.Fatalf("one frame built %d of 10000 options; list is not virtual", built)
	}
	if !r.Focused("Search commands") {
		t.Fatal("native input does not hold keyboard focus")
	}
	if _, active := r.TextCaret(); !active {
		t.Fatal("native input did not expose an IME caret")
	}
	built = 0
	r.Key(0, ui.KeyEnd)
	last := requireCommandHighlight(t, r, "Command 09999")
	if !last.Focused || last.Offscreen || last.PosInSet != 10000 || last.SetSize != 10000 || last.Bounds.H <= 0 || last.Bounds.Y+last.Bounds.H > 360 {
		t.Fatalf("last option was not revealed/active for AX: %+v", last)
	}
	if built > 100 {
		t.Fatalf("End built %d rows instead of jumping virtually", built)
	}
	if _, exists := r.AccessByRole(ui.RoleOption, "Command 00000"); exists {
		t.Fatal("initial option still mounted after virtual jump to End")
	}
	input, _ := r.AccessByRole(ui.RoleComboBox, "Search commands")
	if input.HasPopup != ui.PopupListBox || input.AutoComplete != ui.AutoCompleteList || !input.Expanded || len(input.Controls) != 1 || input.Controls[0] != "Commands" {
		t.Fatalf("combobox/list relationship = %+v", input)
	}
	r.Key(0, ui.KeyHome)
	requireCommandHighlight(t, r, "Command 00000")
	// Actual native scrolling builds a later slice while preserving keyboard focus.
	list, _ := r.AccessByRole(ui.RoleListBox, "Commands")
	r.Scroll(list.Bounds.X+20, list.Bounds.Y+20, 0, 600)
	if _, exists := r.AccessByRole(ui.RoleOption, "Command 00000"); exists {
		t.Fatal("wheel did not move the virtual viewport")
	}
	if !r.Focused("Search commands") {
		t.Fatal("scroll transferred native keyboard focus")
	}
	r.Key(0, ui.KeyHome)
	if first := requireCommandHighlight(t, r, "Command 00000"); first.Offscreen || !first.Focused {
		t.Fatalf("Home did not reveal an already-active item after manual scroll: %+v", first)
	}
	// A unique alias isolates viewport reset from the default fuzzy matcher's
	// intentional repeated-key tolerance (09999 can also match 00999).
	query = "uniquelast"
	r.Frame()
	result := requireCommandHighlight(t, r, "Command 09999")
	if result.PosInSet != 1 || result.SetSize != 1 || !result.Focused || result.Offscreen {
		t.Fatalf("filtered result did not reset scrolling and AX positions: %+v", result)
	}
}

func TestCommandStableKeysDuplicateValuesAndAsyncChanges(t *testing.T) {
	query := ""
	items := []CommandItem{{Key: "one", Value: "same", Label: "First"}, {Key: "two", Value: "same", Label: "Second"}, {Key: "three", Label: "Third"}}
	var picked []string
	r := ui.NewTester(view(func(c *ui.Context) {
		Command(c, &query, CommandProps{Items: items, AutoFocus: true, OnSelect: func(item CommandItem) { picked = append(picked, item.Key) }})
	}), 480, 400)
	r.Key(0, ui.KeyDown)
	requireCommandHighlight(t, r, "Second")
	items = []CommandItem{items[1], items[2], items[0]}
	r.Frame()
	requireCommandHighlight(t, r, "Second")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	r.KeyRepeat(0, ui.KeyEnter)
	if len(picked) != 1 || picked[0] != "two" {
		t.Fatalf("duplicate value/stable key selection = %v", picked)
	}
	items = items[1:]
	r.Frame()
	requireCommandHighlight(t, r, "Third")
	query = "First"
	r.Frame()
	requireCommandHighlight(t, r, "First")
	// Server-side filtering keeps even nonmatching labels with NoFilter.
	query = "server query"
	r = ui.NewTester(view(func(c *ui.Context) { Command(c, &query, CommandProps{Items: items, AutoFocus: true, NoFilter: true}) }), 480, 400)
	requireCommandHighlight(t, r, "Third")
}

func TestCommandKeyboardLoopBoundsPagesAndNativeEditing(t *testing.T) {
	query := ""
	items := make([]CommandItem, 20)
	for i := range items {
		items[i] = CommandItem{Label: fmt.Sprintf("Option %02d", i), Disabled: i == 1 || i == 19}
	}
	noLoop := false
	r := ui.NewTester(view(func(c *ui.Context) {
		Command(c, &query, CommandProps{Items: items, AutoFocus: true, MaxHeight: 160, NoLoop: noLoop, NoFilter: true})
	}), 480, 400)
	r.Key(0, ui.KeyUp)
	requireCommandHighlight(t, r, "Option 18")
	r.Key(0, ui.KeyDown)
	requireCommandHighlight(t, r, "Option 00")
	r.Key(0, ui.KeyDown)
	requireCommandHighlight(t, r, "Option 02")
	r.Key(0, ui.KeyPageDown)
	requireCommandHighlight(t, r, "Option 06")
	r.Key(0, ui.KeyPageUp)
	requireCommandHighlight(t, r, "Option 02")
	noLoop = true
	r.Frame()
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyDown)
	requireCommandHighlight(t, r, "Option 18")
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyUp)
	requireCommandHighlight(t, r, "Option 00")
	r.Type("alpha beta")
	r.Key(0, ui.KeyLeft)
	r.Type("X")
	if query != "alpha betXa" {
		t.Fatalf("native caret/text keys were intercepted: %q", query)
	}
	r.Key(ui.Super, ui.KeyA)
	r.Type("new query")
	if query != "new query" {
		t.Fatalf("native select-all was intercepted: %q", query)
	}
}

func TestCommandBindingsCompositionDisabledAndOnce(t *testing.T) {
	query := ""
	items := []CommandItem{{Value: "first", Label: "First"}, {Value: "run", Label: "Run", Shortcut: "⌘K", Binding: &CommandBinding{Modifiers: ui.Super, Key: ui.KeyK}}, {Label: "Disabled", Disabled: true, Binding: &CommandBinding{Modifiers: ui.Super, Key: ui.KeyD}}}
	var picks []string
	disabled := false
	r := ui.NewTester(view(func(c *ui.Context) {
		if selected := Command(c, &query, CommandProps{Items: items, AutoFocus: true, Disabled: disabled, NoFilter: true}); selected != "" {
			picks = append(picks, selected)
		}
	}), 480, 400)
	r.Key(ui.Super, ui.KeyK)
	r.Frame()
	r.KeyRepeat(ui.Super, ui.KeyK)
	r.Key(ui.Super|ui.Shift, ui.KeyK)
	r.Key(ui.Super, ui.KeyD)
	if len(picks) != 1 || picks[0] != "run" {
		t.Fatalf("shortcut did not activate exactly once with exact modifiers: %v", picks)
	}
	if option, _ := r.AccessByRole(ui.RoleOption, "Run"); option.KeyShortcuts == "" {
		t.Fatal("real shortcut was not exposed to AX")
	}
	r.Compose("中文", 2)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Key(ui.Super, ui.KeyK)
	if len(picks) != 1 {
		t.Fatalf("IME composition activated command: %v", picks)
	}
	requireCommandHighlight(t, r, "First")
	r.Compose("", 0)
	r.Key(0, ui.KeyDown)
	requireCommandHighlight(t, r, "Run")
	r.Key(0, ui.KeyEnter)
	if len(picks) != 2 || picks[1] != "run" {
		t.Fatalf("post-composition submit = %v", picks)
	}
	queryBeforeDisable := query
	disabled = true
	r.Frame()
	r.Key(ui.Super, ui.KeyK)
	r.Key(0, ui.KeyEnter)
	r.Type("ignored")
	if len(picks) != 2 || query != queryBeforeDisable {
		t.Fatalf("disabled command accepted input: picks=%v query=%q", picks, query)
	}
	if _, ok := commandSelected(r); ok {
		t.Fatal("disabled command retained active option")
	}
}

func TestCommandGroupsCustomContentFilteringEmptyAndLoading(t *testing.T) {
	query := ""
	loading := false
	items := []CommandItem{{Label: "Print", Group: "Documents", Keywords: []string{"paper"}}, {Separator: true}, {Label: "Settings", Description: "Preferences", Group: "Tools", Checked: true}}
	picks := 0
	r := ui.NewTester(view(func(c *ui.Context) {
		Command(c, &query, CommandProps{Items: items, AutoFocus: true, Loading: loading,
			ItemContent:    func(row *ui.Context, item CommandItem, active bool) { ui.Text(row, "Custom "+item.Label) },
			EmptyContent:   func(c *ui.Context) { ui.Text(c, "Nothing matches").Role(ui.RoleStatus) },
			LoadingContent: func(c *ui.Context) { ui.Text(c, "Fetching results").Role(ui.RoleStatus) },
			OnSelect:       func(CommandItem) { picks++ },
		})
	}), 480, 400)
	if _, ok := r.AccessByRole(ui.RoleHeading, "Documents"); !ok {
		t.Fatal("missing first group heading")
	}
	if _, ok := r.AccessByRole(ui.RoleHeading, "Tools"); !ok {
		t.Fatal("missing second group heading")
	}
	if _, ok := r.AccessByRole(ui.RoleSeparator, ""); !ok {
		t.Fatal("missing separator")
	}
	if _, ok := r.Access("Custom Settings"); !ok {
		t.Fatal("custom row builder not used")
	}
	query = "paper"
	r.Frame()
	requireCommandHighlight(t, r, "Print")
	if _, ok := r.AccessByRole(ui.RoleHeading, "Tools"); ok {
		t.Fatal("empty group was still rendered")
	}
	if _, ok := r.AccessByRole(ui.RoleSeparator, ""); ok {
		t.Fatal("search rendered an orphan separator")
	}
	query = "unmatched"
	r.Frame()
	r.Key(0, ui.KeyEnter)
	if _, ok := r.AccessByRole(ui.RoleStatus, "Nothing matches"); !ok || picks != 0 {
		t.Fatal("empty custom status or inertness failed")
	}
	query, loading = "", true
	r.Frame()
	r.Key(0, ui.KeyEnter)
	if _, ok := r.AccessByRole(ui.RoleStatus, "Fetching results"); !ok || picks != 0 {
		t.Fatal("loading custom status or inertness failed")
	}
	if selected, ok := commandSelected(r); ok {
		t.Fatalf("loading selected a result: %+v", selected)
	}
	loading = false
	r.Frame()
	requireCommandHighlight(t, r, "Print")
	query = "match"
	r = ui.NewTester(view(func(c *ui.Context) {
		Command(c, &query, CommandProps{Items: items, AutoFocus: true, Filter: func(item CommandItem, query string) float64 {
			if item.Label == "Settings" {
				return 2
			}
			return 1
		}})
	}), 480, 400)
	requireCommandHighlight(t, r, "Settings")
}

func TestCommandPointerSelectionKeepsNativeInputAndDoesNotOverrideKeyboard(t *testing.T) {
	query, chosen := "", ""
	r := ui.NewTester(view(func(c *ui.Context) {
		if v := Command(c, &query, CommandProps{AutoFocus: true, Items: []CommandItem{{Label: "First"}, {Label: "Second"}, {Label: "Third"}}}); v != "" {
			chosen = v
		}
	}), 480, 400)
	second, _ := r.AccessByRole(ui.RoleOption, "Second")
	r.Move(second.Bounds.X+10, second.Bounds.Y+10)
	requireCommandHighlight(t, r, "Second")
	r.Key(0, ui.KeyDown)
	r.Frame()
	requireCommandHighlight(t, r, "Third")
	if err := r.Click("Second"); err != nil {
		t.Fatal(err)
	}
	if chosen != "Second" || !r.Focused("Search commands") {
		t.Fatalf("click: chosen=%q input focus=%v", chosen, r.Focused("Search commands"))
	}
}

func TestCommandDialogHiddenHeaderMetadataAndSelectionClosesOnce(t *testing.T) {
	query, open := "", false
	var picks []string
	r := ui.NewTester(view(func(c *ui.Context) {
		if ui.Button(c, "Open commands").Clicked() {
			open = true
		}
		if selected := CommandDialog(c, &open, &query, CommandProps{Items: []CommandItem{{Label: "Run task"}}, DialogTitle: "Actions", DialogDescription: "Choose a task"}); selected != "" {
			picks = append(picks, selected)
		}
	}), 640, 480)
	if err := r.Click("Open commands"); err != nil {
		t.Fatal(err)
	}
	r.Advance(time.Second)
	dialog, ok := r.AccessByRole(ui.RoleDialog, "Actions")
	if !ok || dialog.Description != "Choose a task" {
		t.Fatalf("dialog metadata=%+v", dialog)
	}
	if _, ok := r.AccessByRole(ui.RoleHeading, "Actions"); ok {
		t.Fatal("CommandDialog rendered its screen-reader-only header")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Close dialog"); ok {
		t.Fatal("CommandDialog close button should be opt-in")
	}
	if !r.Focused("Search commands") {
		t.Fatal("CommandDialog did not focus native input")
	}
	r.Key(0, ui.KeyEnter)
	r.KeyRepeat(0, ui.KeyEnter)
	r.Advance(time.Second)
	if open || len(picks) != 1 || picks[0] != "Run task" || !r.Focused("Open commands") {
		t.Fatalf("dialog open=%v picks=%v return focus=%v", open, picks, r.Focused("Open commands"))
	}
}

func TestCommandOptionsAreNotTabStopsAndBlurClearsComposition(t *testing.T) {
	query := ""
	picked := ""
	r := ui.NewTester(view(func(c *ui.Context) {
		if value := Command(c, &query, CommandProps{AutoFocus: true, NoFilter: true, Items: []CommandItem{{Label: "First"}, {Label: "Second"}}}); value != "" {
			picked = value
		}
		ui.Button(c, "Next field")
	}), 480, 400)
	r.Compose("marked", 6)
	r.Key(0, ui.KeyTab)
	if !r.Focused("Next field") {
		t.Fatal("Tab stopped at a virtual command option")
	}
	r.Key(ui.Shift, ui.KeyTab)
	if !r.Focused("Search commands") {
		t.Fatal("Shift+Tab did not return to the native input")
	}
	r.Key(0, ui.KeyDown)
	requireCommandHighlight(t, r, "Second")
	r.Key(0, ui.KeyEnter)
	if picked != "Second" {
		t.Fatalf("blur left composition/selection stuck: %q", picked)
	}
}

func TestCommandAncestorDisabledNeverAutoFocusesOrSubmits(t *testing.T) {
	query, picked := "", ""
	r := ui.NewTester(view(func(c *ui.Context) {
		ui.Button(c, "Enabled outside").AutoFocus()
		ui.Column(c).Disabled(true).Children(func() {
			picked = Command(c, &query, CommandProps{AutoFocus: true, Items: []CommandItem{{Label: "Unavailable"}}})
		})
	}), 480, 400)
	if !r.Focused("Enabled outside") {
		t.Fatal("command autofocus escaped its disabled ancestor")
	}
	r.Key(0, ui.KeyEnter)
	r.Type("should not type")
	if query != "" || picked != "" {
		t.Fatalf("disabled ancestor accepted query=%q pick=%q", query, picked)
	}
	if _, ok := commandSelected(r); ok {
		t.Fatal("disabled ancestor kept a selected option")
	}
}

func TestCommandCharacterBindingNeverTakesTextInput(t *testing.T) {
	query, picked := "", ""
	r := ui.NewTester(view(func(c *ui.Context) {
		if value := Command(c, &query, CommandProps{AutoFocus: true, NoFilter: true, Items: []CommandItem{{Label: "Should not run", Binding: &CommandBinding{Key: ui.KeyK}}}}); value != "" {
			picked = value
		}
	}), 480, 400)
	r.TypeKey(0, ui.KeyK, "k")
	if picked != "" || query != "k" {
		t.Fatalf("plain character binding intercepted editing: pick=%q query=%q", picked, query)
	}
	if option, _ := r.AccessByRole(ui.RoleOption, "Should not run"); option.KeyShortcuts != "" {
		t.Fatal("AX advertised a character binding that cannot run")
	}
}

func TestCommandDialogUsesTopThirdAndConstrainedVirtualViewport(t *testing.T) {
	for _, size := range [][2]int{{640, 480}, {400, 240}, {240, 100}, {180, 64}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			query, open := "", true
			items := make([]CommandItem, 100)
			for i := range items {
				items[i] = CommandItem{Label: fmt.Sprintf("Action %03d", i)}
			}
			r := ui.NewTester(view(func(c *ui.Context) { CommandDialog(c, &open, &query, CommandProps{Items: items, ShowClose: true}) }), size[0], size[1])
			r.Advance(time.Second)
			panel, ok := r.AccessByRole(ui.RoleDialog, "Command palette")
			if !ok || panel.Bounds.X < 0 || panel.Bounds.Y < 0 || panel.Bounds.X+panel.Bounds.W > float32(size[0])+1 || panel.Bounds.Y+panel.Bounds.H > float32(size[1])-15 {
				t.Fatalf("palette escaped the remaining viewport: %+v", panel)
			}
			if size[1] >= 240 && math.Abs(float64(panel.Bounds.Y-float32(size[1])/3)) > 1 {
				t.Fatalf("palette top=%g, want third of %d", panel.Bounds.Y, size[1])
			}
			if size[0] == 640 && math.Abs(float64(panel.Bounds.W-448)) > 1 {
				t.Fatalf("default palette width=%g", panel.Bounds.W)
			}
			if panel.Bounds.X < 15 || panel.Bounds.X+panel.Bounds.W > float32(size[0])-15 {
				t.Fatalf("palette lost horizontal viewport margins: %+v", panel)
			}
			r.Key(0, ui.KeyEnd)
			last := requireCommandHighlight(t, r, "Action 099")
			list, _ := r.AccessByRole(ui.RoleListBox, "Commands")
			if list.Bounds.Y+list.Bounds.H > panel.Bounds.Y+panel.Bounds.H+1 {
				t.Fatalf("list escaped palette: list=%+v panel=%+v", list, panel)
			}
			if size[1] <= 64 && last.Focused {
				t.Fatalf("offscreen option held AX focus: %+v", last)
			}
			closeButton, ok := r.AccessByRole(ui.RoleButton, "Close dialog")
			if !ok {
				t.Fatal("opt-in close button missing")
			}
			input, _ := r.AccessByRole(ui.RoleComboBox, "Search commands")
			if input.Bounds.X+input.Bounds.W > closeButton.Bounds.X || closeButton.Bounds.Y+closeButton.Bounds.H > panel.Bounds.Y+panel.Bounds.H {
				t.Fatalf("close button obscures input or leaves palette: input=%v close=%v panel=%v", input.Bounds, closeButton.Bounds, panel.Bounds)
			}
			if err := r.Click("Close dialog"); err != nil {
				t.Fatal(err)
			}
			if open {
				t.Fatal("opt-in close button did not close the palette")
			}
		})
	}
}

func TestCommandPageKeysAccountForDescriptionsAndSectionHeaders(t *testing.T) {
	query := ""
	items := make([]CommandItem, 1000)
	for i := range items {
		items[i] = CommandItem{Label: fmt.Sprintf("Command %03d", i), Description: "A second line", Group: fmt.Sprintf("Section %03d", i)}
	}
	r := ui.NewTester(view(func(c *ui.Context) { Command(c, &query, CommandProps{Items: items, AutoFocus: true, MaxHeight: 160}) }), 480, 400)
	r.Key(0, ui.KeyEnd)
	requireCommandHighlight(t, r, "Command 999")
	r.Key(0, ui.KeyPageUp)
	previous, ok := commandSelected(r)
	if !ok || previous.PosInSet < 998 || previous.PosInSet >= 1000 || previous.Offscreen || !previous.Focused {
		t.Fatalf("PageUp jumped more than the 160-DIP viewport of 76-DIP rows: %+v", previous)
	}
	r.Key(0, ui.KeyPageDown)
	requireCommandHighlight(t, r, "Command 999")
}

func TestCommandScoresKeepSectionsContiguousAndDoNotMoveSeparatorToTop(t *testing.T) {
	query := ""
	items := []CommandItem{{Label: "A low", Group: "A"}, {Label: "A high", Group: "A"}, {Separator: true}, {Label: "B middle", Group: "B"}}
	r := ui.NewTester(view(func(c *ui.Context) {
		Command(c, &query, CommandProps{Items: items, AutoFocus: true, Filter: func(item CommandItem, query string) float64 {
			switch item.Label {
			case "A high":
				return 3
			case "B middle":
				return 2
			default:
				return 1
			}
		}})
	}), 480, 400)
	// An empty upstream search keeps registration order and its separators.
	var emptyNames []string
	for _, info := range r.AccessAll() {
		if info.Role == ui.RoleOption {
			emptyNames = append(emptyNames, info.Name)
		}
	}
	if fmt.Sprint(emptyNames) != "[A low A high B middle]" {
		t.Fatalf("empty search lost registration order: %v", emptyNames)
	}
	query = "rank"
	r.Frame()
	var names []string
	groupA := 0
	firstOptionY, separatorY := float32(0), float32(0)
	for _, info := range r.AccessAll() {
		if info.Role == ui.RoleOption {
			names = append(names, info.Name)
			if firstOptionY == 0 {
				firstOptionY = info.Bounds.Y
			}
		}
		if info.Role == ui.RoleHeading && info.Name == "A" {
			groupA++
		}
		if info.Role == ui.RoleSeparator {
			separatorY = info.Bounds.Y
		}
	}
	if fmt.Sprint(names) != "[A high A low B middle]" || groupA != 1 || separatorY != 0 {
		t.Fatalf("scores fragmented a section: options=%v A-headings=%d separatorY=%g firstY=%g", names, groupA, separatorY, firstOptionY)
	}
}

func TestCommandOffscreenOverscanNeverBecomesActiveDescendant(t *testing.T) {
	query := ""
	items := make([]CommandItem, 100)
	for i := range items {
		items[i] = CommandItem{Label: fmt.Sprintf("Result %03d", i)}
	}
	r := ui.NewTester(view(func(c *ui.Context) { Command(c, &query, CommandProps{Items: items, AutoFocus: true, MaxHeight: 160}) }), 480, 400)
	list, _ := r.AccessByRole(ui.RoleListBox, "Commands")
	// Stay in the list's right padding, so scrolling does not hover an option.
	r.Scroll(list.Bounds.X+list.Bounds.W-2, list.Bounds.Y+20, 0, 40)
	first, exists := r.AccessByRole(ui.RoleOption, "Result 000")
	if !exists || !first.Selected || !first.Offscreen || first.Focused {
		t.Fatalf("overscan option captured AX focus after scroll: %+v", first)
	}
	input, _ := r.AccessByRole(ui.RoleComboBox, "Search commands")
	if !input.Focused || !r.Focused("Search commands") {
		t.Fatal("offscreen active option did not leave AX and keyboard focus on input")
	}
	r.Key(0, ui.KeyHome)
	if first := requireCommandHighlight(t, r, "Result 000"); !first.Focused || first.Offscreen {
		t.Fatalf("Home did not restore visible active descendant: %+v", first)
	}
}
