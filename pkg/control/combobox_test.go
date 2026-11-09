package control

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"image"
	"slices"
	"strings"
	"testing"
)

func selectionComboInput(t *testing.T, r *ui.Tester, label string) ui.AccessInfo {
	t.Helper()
	// PopupInput has both a combobox/dialog trigger and a list-autocomplete
	// combobox editor. Prefer the actual editor while that dialog is open.
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleComboBox && n.Name == label && n.HasPopup == ui.PopupListBox {
			return n
		}
	}
	n, ok := r.AccessByRole(ui.RoleComboBox, label)
	if !ok {
		t.Fatalf("missing combobox %q: %v", label, r.AccessAll())
	}
	return n
}
func openSelectionCombo(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	n := selectionComboInput(t, r, label)
	r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
	r.Frame()
}
func clickComboOption(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	n, ok := r.AccessByRole(ui.RoleOption, label)
	if !ok {
		t.Fatalf("missing option %q: %v", label, r.AccessAll())
	}
	r.ClickAt(n.Bounds.X+n.Bounds.W/2, n.Bounds.Y+n.Bounds.H/2)
	r.Frame()
}
func TestComboboxSelectionSeparatesValueLabelAndQuery(t *testing.T) {
	value := ComboboxValue{}
	q := ""
	changes, inputs := 0, 0
	options := []Option{{Value: "next", Label: "Next.js"}, {Value: "go", Label: "Go"}, {Value: "", Label: "None"}}
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &value, &q, ComboboxWithProps{Label: "Framework", Options: options, ShowClear: true, OnValueChange: func(v ComboboxValue) { changes++ }, OnInputChange: func(string) { inputs++ }}).Width(260)
	})
	openSelectionCombo(t, r, "Framework")
	r.Type("Next")
	if value.Valid || q != "Next" {
		t.Fatalf("typing committed value or changed query: %+v %q", value, q)
	}
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleOption && n.Focused {
			t.Fatal("default unexpectedly auto-highlighted first item")
		}
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value.Valid || selectionComboInput(t, r, "Framework").Expanded {
		t.Fatalf("unhighlighted Enter failed to dismiss/reset without selection: %+v %q", value, q)
	}
	r.Advance(comboMotionDuration)
	r.Frame()
	if q != "" {
		t.Fatalf("completed Enter exit failed to reset query: %q", q)
	}
	openSelectionCombo(t, r, "Framework")
	r.Type("Next")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value != (ComboboxValue{"next", true}) || q != "Next.js" || changes != 1 {
		t.Fatalf("selection=%+v query=%q calls=%d", value, q, changes)
	}
	if r.HasText("None") {
		t.Fatal("single selection left popup open")
	}
	if selectionComboInput(t, r, "Framework").AutoComplete != ui.AutoCompleteList {
		t.Fatal("missing native list autocomplete")
	}
	click(t, r, "Clear Framework")
	if value.Valid || q != "" || changes != 2 {
		t.Fatalf("clear=%+v %q calls%d", value, q, changes)
	}
	openSelectionCombo(t, r, "Framework")
	clickComboOption(t, r, "None")
	if value != (ComboboxValue{"", true}) || q != "None" {
		t.Fatalf("empty identity lost: %+v %q", value, q)
	}
	old := changes
	value = ComboboxValue{"go", true}
	r.Frame()
	if q != "Go" || changes != old {
		t.Fatalf("programmatic value=%+v q=%q callbacks=%d", value, q, changes)
	}
	r.Frame()
	if changes != old {
		t.Fatal("idle repeated selection callback")
	}
	if inputs == 0 {
		t.Fatal("missing input callback")
	}
}
func TestComboboxKeyboardSkipsDisabledLoopsAndControlsHighlight(t *testing.T) {
	for _, auto := range []bool{false, true} {
		for _, noLoop := range []bool{false, true} {
			t.Run(map[bool]string{false: "manual", true: "auto"}[auto]+map[bool]string{false: "/loop", true: "/no-loop"}[noLoop], func(t *testing.T) {
				value := ComboboxValue{}
				q := ""
				r := tester(func(c *ui.Context) {
					ComboboxWith(c, &value, &q, ComboboxWithProps{Label: "Language", AutoHighlight: auto, NoLoop: noLoop, ShowClear: true, Options: []Option{{Value: "off", Label: "Unavailable", Disabled: true}, {Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}})
				})
				openSelectionCombo(t, r, "Language")
				r.Advance(comboMotionDuration)
				r.Frame()
				r.Key(0, ui.KeyDown)
				r.Frame()
				if comboBehaviorActive(r) != "Go" {
					t.Fatal("initial Down did not skip disabled option")
				}
				r.Key(0, ui.KeyUp)
				r.Frame()
				if !auto && !noLoop {
					if !r.Focused("Language") || !selectionComboInput(t, r, "Language").Expanded {
						t.Fatal("manual loop did not pass through native input without closing")
					}
					r.Key(0, ui.KeyUp)
				}
				r.Key(0, ui.KeyEnter)
				r.Frame()
				want := "rust"
				if noLoop {
					want = "go"
				}
				if value != (ComboboxValue{want, true}) {
					t.Fatalf("selected=%+v want%s", value, want)
				}
				click(t, r, "Clear Language")
			})
		}
	}
}
func TestComboboxClearAutoHighlightAndDismissal(t *testing.T) {
	value := ComboboxValue{"go", true}
	q := ""
	opens := []bool{}
	other := ""
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &value, &q, ComboboxWithProps{Label: "Language", ShowClear: true, AutoHighlight: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}, OnOpenChange: func(v bool) { opens = append(opens, v) }})
		Input(c, &other, InputProps{Label: "Other"})
	})
	openSelectionCombo(t, r, "Language")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Ru")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value.Value != "rust" || q != "Rust" {
		t.Fatalf("auto selection %+v %q", value, q)
	}
	openSelectionCombo(t, r, "Language")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("wrong")
	r.Key(0, ui.KeyTab)
	r.Frame()
	if !r.Focused("Other") {
		t.Fatalf("Tab did not reach Other: %+v", r.AccessAll())
	}
	r.Advance(comboMotionDuration)
	r.Frame()
	if q != "Rust" || value.Value != "rust" {
		t.Fatalf("blur failed restore %+v %q", value, q)
	}
	openSelectionCombo(t, r, "Language")
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if selectionComboInput(t, r, "Language").Expanded {
		t.Fatal("Escape did not dismiss")
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if value.Valid {
		t.Fatal("closed Escape did not clear selected value")
	}
	if len(opens) < 4 {
		t.Fatalf("open events=%v", opens)
	}
}
func TestComboboxCompositionDoesNotChooseOrRemoveChip(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		v := ComboboxValue{}
		q := ""
		r := tester(func(c *ui.Context) {
			ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Language", AutoHighlight: true, Options: []Option{{Value: "zh", Label: "你好"}}})
		})
		openSelectionCombo(t, r, "Language")
		r.Compose("你好", 2)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if v.Valid {
			t.Fatalf("candidate Enter chose %+v", v)
		}
	})
	t.Run("multiple", func(t *testing.T) {
		v := []string{"go"}
		q := ""
		r := tester(func(c *ui.Context) {
			MultiComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Language", AutoHighlight: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "zh", Label: "你好"}}})
		})
		openSelectionCombo(t, r, "Language")
		r.Compose("你好", 2)
		r.Key(0, ui.KeyBackspace)
		r.Frame()
		if !slices.Equal(v, []string{"go"}) {
			t.Fatalf("IME backspace removed %+v", v)
		}
		r.Compose("你好", 2)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if !slices.Equal(v, []string{"go"}) {
			t.Fatalf("IME Enter chose %+v", v)
		}
	})
}
func TestMultiComboboxTogglesKeepsVisibleAndSharesChipField(t *testing.T) {
	values := []string{}
	q := ""
	calls := 0
	r := tester(func(c *ui.Context) {
		MultiComboboxWith(c, &values, &q, ComboboxWithProps{Label: "Language", ShowClear: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}, {Value: "off", Label: "Unavailable", Disabled: true}}, OnValuesChange: func(v []string) {
			calls++
			if len(v) > 0 {
				v[0] = "mutated callback"
			}
		}}).Width(300)
	})
	openSelectionCombo(t, r, "Language")
	if list, ok := r.AccessByRole(ui.RoleListBox, "Language options"); !ok || !list.Multiselectable {
		t.Fatal("multiple listbox semantics missing")
	}
	clickComboOption(t, r, "Go")
	if !slices.Equal(values, []string{"go"}) || calls != 1 {
		t.Fatalf("selected=%v calls=%d", values, calls)
	}
	n := selectionComboInput(t, r, "Language")
	if !n.Expanded {
		t.Fatal("unfiltered selection closed popup")
	}
	if _, ok := r.AccessByRole(ui.RoleOption, "Go"); !ok {
		t.Fatal("selected option disappeared")
	}
	clickComboOption(t, r, "Go")
	if len(values) != 0 || calls != 2 {
		t.Fatalf("toggle=%v calls=%d", values, calls)
	}
	r.Type("Rust")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(values, []string{"rust"}) || q != "" {
		t.Fatalf("filtered selection=%v %q", values, q)
	}
	if selectionComboInput(t, r, "Language").Expanded {
		t.Fatal("filtered external selection kept popup open")
	}
	chip, ok := r.AccessByRole(ui.RoleButton, "Remove Rust")
	if !ok {
		t.Fatal("missing remove button")
	}
	input := selectionComboInput(t, r, "Language")
	if chip.Bounds.Y > input.Bounds.Y+input.Bounds.H || input.Bounds.Y > chip.Bounds.Y+chip.Bounds.H {
		t.Fatal("chips and input occupy different rows with available room")
	}
	click(t, r, "Clear Language")
	if len(values) != 0 || q != "" {
		t.Fatal("clear multiple failed")
	}
}
func TestComboboxPopupSearchAndNativeFocusReturn(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Country", PopupInput: true, Placeholder: "Choose country", InputPlaceholder: "Find country", Options: []Option{{Value: "us", Label: "United States"}, {Value: "jp", Label: "Japan"}}}).Width(260)
	})
	click(t, r, "Country")
	if !r.Focused("Country") {
		t.Fatalf("popup editor not focused: %v", r.AccessAll())
	}
	r.Type("Japan")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if v != (ComboboxValue{"jp", true}) || q != "Japan" {
		t.Fatalf("popup selection %+v %q", v, q)
	}
	if !r.Focused("Country") {
		t.Fatal("popup did not restore trigger focus")
	}
	if !r.HasText("Japan") {
		t.Fatal("trigger did not show selected label")
	}
	r.Advance(comboMotionDuration)
	r.Frame()
	if q != "" {
		t.Fatalf("popup query survived completed exit: %q", q)
	}
}

func TestComboboxDynamicOptionsDisabledAndCallbacks(t *testing.T) {
	for _, ancestor := range []bool{false, true} {
		t.Run(map[bool]string{false: "explicit", true: "ancestor"}[ancestor], func(t *testing.T) {
			value := ComboboxValue{"go", true}
			q := ""
			disabled := false
			calls, changed := 0, 0
			options := []Option{{Value: "go", Label: "Go old"}, {Value: "rust", Label: "Rust"}}
			r := tester(func(c *ui.Context) {
				ui.Column(c).Disabled(ancestor && disabled).Children(func() {
					root := ComboboxWith(c, &value, &q, ComboboxWithProps{Label: "Language", Options: options, Disabled: !ancestor && disabled, OnValueChange: func(ComboboxValue) { calls++ }})
					if root.Changed() {
						changed++
					}
				})
			})
			if q != "Go old" {
				t.Fatalf("initial label=%q", q)
			}
			options[0].Label = "Go new"
			r.Frame()
			if q != "Go new" || calls != 0 {
				t.Fatalf("async label=%q callback%d", q, calls)
			}
			openSelectionCombo(t, r, "Language")
			r.Advance(comboMotionDuration)
			r.Frame()
			if comboBehaviorActive(r) != "Go new" {
				t.Fatal("opening did not highlight selected stable identity")
			}
			options = options[1:]
			r.Frame()
			r.Key(0, ui.KeyEnter)
			r.Frame()
			if value.Value != "go" || calls != 0 {
				t.Fatalf("removed active chosen=%+v calls%d", value, calls)
			}
			r.Key(0, ui.KeyDown)
			r.Key(0, ui.KeyEnter)
			r.Frame()
			if value.Value != "rust" || calls != 1 || changed == 0 {
				t.Fatalf("real choice=%+v calls%d changed%d", value, calls, changed)
			}
			openSelectionCombo(t, r, "Language")
			disabled = true
			r.Frame()
			for _, n := range r.AccessAll() {
				if n.Role == ui.RoleOption {
					t.Fatal("disabled combobox retained active AX options")
				}
			}
			r.Key(0, ui.KeyEnter)
			if calls != 1 {
				t.Fatal("disabled choice emitted callback")
			}
		})
	}
}
func TestComboboxLabelOnlyLocaleFilterGroupsAndCustomContent(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	custom := false
	options := []Option{{Value: "secret", Label: "Café", Group: "Drinks"}, {Separator: true}, {Value: "b", Label: "Tea", Group: "Drinks"}, {Value: "c", Label: "Milk", Group: "Other"}}
	r := tester(func(c *ui.Context) {
		p := ComboboxWithProps{Label: "Drink", Options: options, Locale: "en", AutoHighlight: true, ItemContent: func(c *ui.Context, o Option) { ui.Text(c, "Drink: "+optionLabel(o)).Label("Custom " + o.Label) }}
		if custom {
			p.Filter = func(o Option, s string) bool { return strings.Contains(o.Value, s) }
		}
		ComboboxWith(c, &v, &q, p)
	})
	openSelectionCombo(t, r, "Drink")
	if _, ok := r.AccessByRole(ui.RoleGroup, "Drinks"); !ok {
		t.Fatal("named option group missing")
	}
	if _, ok := r.AccessByRole(ui.RoleSeparator, ""); !ok {
		t.Fatal("group separator semantics missing")
	}
	r.Type("cafe")
	if _, ok := r.AccessByRole(ui.RoleOption, "Café"); !ok {
		t.Fatal("diacritic-insensitive label missing")
	}
	r.Key(0, ui.KeyEnter)
	if v.Value != "secret" || q != "Café" {
		t.Fatalf("custom option selected wrong %+v %q", v, q)
	}
	openSelectionCombo(t, r, "Drink")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("secret")
	if _, ok := r.AccessByRole(ui.RoleOption, "Café"); ok {
		t.Fatal("default searched serialized identity")
	}
	custom = true
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleOption, "Café"); !ok {
		t.Fatal("explicit custom filter not applied")
	}
}
func TestMultiComboboxChipKeyboardRTLStableReorderAndLimits(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(map[bool]string{false: "ltr", true: "rtl"}[rtl], func(t *testing.T) {
			values := []string{"go", "rust"}
			q := ""
			hide := false
			disabled := false
			r := tester(func(c *ui.Context) {
				if rtl {
					c.Root().Direction(ui.RightToLeft)
				}
				MultiComboboxWith(c, &values, &q, ComboboxWithProps{Label: "Language", HideRemove: hide, Disabled: disabled, MaxSelected: 2, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}, {Value: "swift", Label: "Swift"}}}).Width(360)
			})
			openSelectionCombo(t, r, "Language")
			r.Key(0, ui.KeyHome)
			previous, next := ui.KeyLeft, ui.KeyRight
			if rtl {
				previous, next = next, previous
			}
			r.Key(0, previous)
			r.Frame()
			if !r.Focused("Rust") {
				t.Fatalf("previous at caret0 did not focus last chip: %+v", r.AccessAll())
			}
			values = []string{"rust", "go"}
			r.Frame()
			if !r.Focused("Rust") {
				t.Fatal("chip focus changed after stable-value reorder")
			}
			r.Key(0, next)
			r.Frame()
			if !r.Focused("Go") {
				t.Fatal("logical next did not follow reordered chip")
			}
			hide = true
			r.Frame()
			if r.HasText("Remove Go") {
				t.Fatal("hidden remove button still shown")
			}
			r.Key(0, ui.KeyBackspace)
			r.Frame()
			if !slices.Equal(values, []string{"rust"}) || !r.Focused("Rust") {
				t.Fatalf("keyboard removal=%v focused=%v", values, r.AccessAll())
			}
			r.Key(0, ui.KeyEnter)
			r.Frame()
			if !r.Focused("Language") {
				t.Fatal("chip Enter did not return editor")
			}
			r.Key(0, ui.KeyDown)
			r.Key(0, ui.KeyDown)
			r.Key(0, ui.KeyDown)
			r.Key(0, ui.KeyEnter)
			r.Frame()
			if len(values) > 2 {
				t.Fatal("max selected ignored")
			}
			disabled = true
			r.Frame()
			old := slices.Clone(values)
			r.Key(0, ui.KeyBackspace)
			if !slices.Equal(values, old) {
				t.Fatal("disabled removed chip")
			}
		})
	}
}
func TestComboboxPreservesNativeEditorSelectionAndNoKeyLeak(t *testing.T) {
	v := ComboboxValue{}
	q := "draft"
	other := ""
	prefix := false
	options := []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}
	r := tester(func(c *ui.Context) {
		if prefix {
			ui.Text(c, "Inserted sibling")
		}
		ComboboxWith(c.Key("stable-combo"), &v, &q, ComboboxWithProps{Label: "Language", Options: options})
		Input(c.Key("stable-other"), &other, InputProps{Label: "Other"})
	})
	openSelectionCombo(t, r, "Language")
	r.Key(ui.Cmd, ui.KeyA)
	prefix = true
	options = []Option{options[1], options[0]}
	r.Frame()
	r.Type("replacement")
	if q != "replacement" {
		t.Fatalf("native selection lost=%q", q)
	}
	r.Key(0, ui.KeyTab)
	r.Frame()
	if !r.Focused("Other") {
		t.Fatal("Tab after open lost following native editor")
	}
	r.Type("outside")
	if other != "outside" {
		t.Fatal("outside native editor lost focus on popup close")
	}
	r.Key(ui.Shift, ui.KeyTab)
	r.Frame()
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEscape)
	r.Frame()
	r.Key(0, ui.KeyTab)
	r.Frame()
	if !r.Focused("Other") {
		t.Fatal("closed popup leaked Context.Key")
	}
}

func legacyComboClick(t *testing.T, r *ui.Tester) {
	t.Helper()
	n, ok := r.AccessByRole(ui.RoleComboBox, "Language")
	if !ok {
		t.Fatal("missing legacy combo")
	}
	r.ClickAt(n.Bounds.X+8, n.Bounds.Y+8)
	r.Frame()
}
func TestLegacyComboDisabledCannotChooseKeyboard(t *testing.T) {
	value := ""
	r := tester(func(c *ui.Context) {
		Combobox(c, &value, ComboboxProps{Label: "Language", Options: []Option{{Value: "off", Label: "Unavailable", Disabled: true}, {Value: "go", Label: "Go"}}})
	})
	legacyComboClick(t, r)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value == "off" {
		t.Fatal("disabled option selected by native Enter")
	}
	t.Logf("value=%q", value)
}
func TestLegacyMultiDisabledSkippedByKeyboard(t *testing.T) {
	selected := []string{}
	query := ""
	r := tester(func(c *ui.Context) {
		MultiCombobox(c, &selected, &query, MultiComboboxProps{Label: "Language", Options: []Option{{Value: "off", Label: "Unavailable", Disabled: true}, {Value: "go", Label: "Go"}}})
	})
	legacyComboClick(t, r)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(selected, []string{"go"}) {
		t.Fatalf("keyboard did not skip disabled option, selected=%v query=%q", selected, query)
	}
}
func TestLegacyMultiBackspaceOnlyRemovesChipOutsideIME(t *testing.T) {
	selected := []string{"go"}
	query := ""
	r := tester(func(c *ui.Context) {
		MultiCombobox(c, &selected, &query, MultiComboboxProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go"}, {Value: "zh", Label: "中文"}}})
	})
	legacyComboClick(t, r)
	r.Compose("你好", 2)
	r.Key(0, ui.KeyBackspace)
	r.Frame()
	if !slices.Equal(selected, []string{"go"}) {
		t.Fatalf("IME backspace removed selected chip: selected=%v query=%q", selected, query)
	}
}
func TestLegacyMultiFinalTextBackspaceDoesNotRemoveChip(t *testing.T) {
	selected := []string{"go"}
	query := "x"
	r := tester(func(c *ui.Context) {
		MultiCombobox(c, &selected, &query, MultiComboboxProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go"}}})
	})
	legacyComboClick(t, r)
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyBackspace)
	r.Frame()
	if !slices.Equal(selected, []string{"go"}) || query != "" {
		t.Fatalf("text backspace mutated chip or query: selected=%v query=%q", selected, query)
	}
}
func TestLegacyComboAdvertisesListAutocomplete(t *testing.T) {
	value := ""
	r := tester(func(c *ui.Context) {
		Combobox(c, &value, ComboboxProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go"}}})
	})
	a, _ := r.AccessByRole(ui.RoleComboBox, "Language")
	if a.AutoComplete != ui.AutoCompleteList {
		t.Fatalf("AX autocomplete=%v want list", a.AutoComplete)
	}
}

func TestLegacyComboDisabledWhileOpenCannotChoose(t *testing.T) {
	value := ""
	disabled := false
	r := tester(func(c *ui.Context) {
		Combobox(c, &value, ComboboxProps{Label: "Language", Disabled: disabled, Options: []Option{{Value: "go", Label: "Go"}}})
	})
	legacyComboClick(t, r)
	r.Key(0, ui.KeyDown)
	disabled = true
	r.Frame()
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value != "" {
		t.Fatalf("disabled open combo chose %q", value)
	}
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleListItem {
			t.Fatalf("disabled combo retained option AX %+v", a)
		}
	}
}
func TestLegacyComboCandidateEnterDoesNotChoose(t *testing.T) {
	value := ""
	r := tester(func(c *ui.Context) {
		Combobox(c, &value, ComboboxProps{Label: "Language", Options: []Option{{Value: "zh", Label: "你好"}}})
	})
	legacyComboClick(t, r)
	r.Compose("你好", 2)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value != "你好" {
		t.Fatalf("candidate Enter chose option too early: value=%q", value)
	}
}
func TestLegacyMultiRemoveFocusSurvivesReorder(t *testing.T) {
	selected := []string{"go", "rust"}
	query := ""
	r := tester(func(c *ui.Context) {
		MultiCombobox(c, &selected, &query, MultiComboboxProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}})
	})
	r.Key(0, ui.KeyTab)
	r.Frame()
	if !r.Focused("Remove Go") {
		t.Fatalf("initial Tab did not focus Remove Go: AX=%+v", r.AccessAll())
	}
	selected = []string{"rust", "go"}
	r.Frame()
	if !r.Focused("Remove Go") {
		t.Fatalf("keyed chip focus moved on reorder: AX=%+v", r.AccessAll())
	}
}
func TestLegacyComboChoiceReportsChanged(t *testing.T) {
	value := ""
	changes := 0
	r := tester(func(c *ui.Context) {
		if Combobox(c, &value, ComboboxProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go"}}}).Changed() {
			changes++
		}
	})
	legacyComboClick(t, r)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if value != "go" || changes != 1 {
		t.Fatalf("choice value=%q Changed count=%d", value, changes)
	}
}

func TestLegacyComboOpenThenDisableRejectsPointerChoice(t *testing.T) {
	value := ""
	disabled := false
	r := tester(func(c *ui.Context) {
		Combobox(c, &value, ComboboxProps{Label: "Language", Disabled: disabled, Options: []Option{{Value: "go", Label: "Go"}}})
	})
	legacyComboClick(t, r)
	disabled = true
	r.Frame()
	if a, ok := r.AccessByRole(ui.RoleListItem, "Go"); ok {
		r.ClickAt(a.Bounds.X+8, a.Bounds.Y+8)
		r.Frame()
	}
	if value != "" {
		t.Fatalf("disabled combo accepted popup pointer choice %q", value)
	}
}
func TestLegacyMultiCandidateEnterDoesNotAddChip(t *testing.T) {
	selected := []string{"go"}
	query := ""
	r := tester(func(c *ui.Context) {
		MultiCombobox(c, &selected, &query, MultiComboboxProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go"}, {Value: "zh", Label: "你好"}}})
	})
	legacyComboClick(t, r)
	r.Compose("你好", 2)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(selected, []string{"go"}) || query != "你好" {
		t.Fatalf("candidate Enter added suggestion chip too early selected=%v query=%q", selected, query)
	}
}

func TestComboboxNativeSizeBoundsAndTextScale(t *testing.T) {
	for _, tc := range []struct {
		name   string
		size   Size
		height float32
	}{{"default", DefaultSize, 36}, {"small", Small, 32}, {"extra-small", ExtraSmall, 24}, {"large", Large, 40}} {
		t.Run(tc.name, func(t *testing.T) {
			v := ComboboxValue{}
			q := ""
			var handle ui.Handle
			var bounds ui.Rect
			r := tester(func(c *ui.Context) {
				ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Language", Size: tc.size, Options: []Option{{Value: "go", Label: "Go"}}}).Bind(&handle).Width(240).Draw(func(_ *ui.Painter, r ui.Rect) { bounds = r })
			})
			if box := bounds; box.H != tc.height {
				t.Fatalf("native size %s height=%v want%v", tc.name, box.H, tc.height)
			}
			input := selectionComboInput(t, r, "Language")
			if input.Bounds.Y < bounds.Y || input.Bounds.Y+input.Bounds.H > bounds.Y+bounds.H {
				t.Fatal("native editor clipped outside trigger")
			}
		})
	}
	v := ComboboxValue{}
	q := ""
	var handle ui.Handle
	var bounds ui.Rect
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Language", Size: ExtraSmall, Options: []Option{{Value: "go", Label: "Go"}}}).Bind(&handle).Width(220).Draw(func(_ *ui.Painter, r ui.Rect) { bounds = r })
	})
	preferences := ui.Preferences{}
	preferences.TextScale = 2
	r.SetPreferences(preferences)
	r.Frame()
	if bounds.H < 34 {
		t.Fatalf("native scaled text clipped by XS height: %v", bounds)
	}
}
func TestComboboxFieldAndInputGroupKeepNativeTargetAndFocusRing(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	var handle ui.Handle
	r := tester(func(c *ui.Context) {
		Field(c, FieldProps{Label: "Language field", Error: "Choose language", Description: "Select one"}, func() ui.Element {
			return InputGroup(c, InputGroupProps{Prefix: "Code"}, func() ui.Element {
				return ComboboxWith(c, &v, &q, ComboboxWithProps{Options: []Option{{Value: "go", Label: "Go"}}}).Bind(&handle)
			})
		})
	})
	input := selectionComboInput(t, r, "Language field")
	if !input.Invalid || !slices.Contains(input.ErrorMessage, "Choose language") || !slices.Contains(input.DescribedBy, "Select one") {
		t.Fatalf("Field did not name/describe/error native editor: %+v", input)
	}
	r.Key(0, ui.KeyTab)
	r.Type("Go")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if v != (ComboboxValue{"go", true}) {
		t.Fatalf("InputGroup native keyboard chose %+v", v)
	}
	q = ""
	v = ComboboxValue{}
	var outer ui.Handle
	var outerBounds ui.Rect
	r = tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Plain", Options: []Option{{Value: "go", Label: "Go"}}}).Bind(&outer).Width(240).Draw(func(_ *ui.Painter, r ui.Rect) { outerBounds = r })
	})
	box := outerBounds
	p := r.Image().RGBAAt(int(box.X)-2, int(box.Y+box.H/2))
	r.Key(0, ui.KeyTab)
	r.Frame()
	focused := r.Image().RGBAAt(int(box.X)-2, int(box.Y+box.H/2))
	if p == focused {
		t.Fatal("keyboard focus did not draw outer input-group ring")
	}
}
func TestComboboxPopupSearchTabOrderAndEscapeReturn(t *testing.T) {
	for _, mode := range []string{"closed", "open", "typed"} {
		t.Run(mode, func(t *testing.T) {
			v := ComboboxValue{"go", true}
			q := ""
			before, after := "", ""
			r := tester(func(c *ui.Context) {
				Input(c, &before, InputProps{Label: "Before"})
				ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Language", PopupInput: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}})
				Input(c, &after, InputProps{Label: "After"})
			})
			click(t, r, "Language")
			if mode == "closed" {
				r.Key(0, ui.KeyEscape)
				r.Frame()
				if !r.Focused("Language") {
					t.Fatal("Escape did not return popup trigger focus")
				}
			} else if mode == "typed" {
				r.Type("wrong")
			}
			r.Key(0, ui.KeyTab)
			r.Frame()
			if !r.Focused("After") {
				t.Fatalf("popup-input %s Tab did not follow anchor order: %+v", mode, r.AccessAll())
			}
			r.Type("next")
			if after != "next" {
				t.Fatal("native next editor lost focus")
			}
		})
	}
}

func TestComboboxPopupClearAndMultipleSelectedSummary(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "multiple"}[multiple], func(t *testing.T) {
			single := ComboboxValue{"go", true}
			values := []string{"go", "rust"}
			q := ""
			calls := 0
			r := tester(func(c *ui.Context) {
				p := ComboboxWithProps{Label: "Language", PopupInput: true, ShowClear: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}, OnValueChange: func(ComboboxValue) { calls++ }, OnValuesChange: func([]string) { calls++ }}
				if multiple {
					MultiComboboxWith(c, &values, &q, p)
				} else {
					ComboboxWith(c, &single, &q, p)
				}
			})
			if multiple && !r.HasText("Go, Rust") {
				t.Fatal("popup multi trigger did not summarize selected labels")
			}
			click(t, r, "Language")
			click(t, r, "Clear Language")
			r.Frame()
			if (multiple && len(values) != 0) || (!multiple && single.Valid) || q != "" || calls != 1 {
				t.Fatalf("popup clear single=%+v multi=%v q=%q calls=%d", single, values, q, calls)
			}
			if !r.Focused("Language") {
				t.Fatal("popup clear moved focus away from native search")
			}
		})
	}
}

func TestComboboxPopupButtonGroupMatchesJoinedNativeButtonGeometryAndPixels(t *testing.T) {
	for _, last := range []bool{false, true} {
		t.Run(map[bool]string{false: "first", true: "last"}[last], func(t *testing.T) {
			render := func(combo bool) (*ui.Tester, ui.Rect) {
				value := ComboboxValue{}
				query := ""
				var outer ui.Rect
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					ui.Column(c).Padding(20).Children(func() {
						ButtonGroup(c, func() {
							before := func() { Button(c, "Before", ButtonProps{Variant: Outline}).Width(80) }
							choice := func() {
								var root ui.Element
								if combo {
									root = ComboboxWith(c, &value, &query, ComboboxWithProps{Label: "Choice", Placeholder: "Pick", PopupInput: true})
								} else {
									root = Button(c, "Pick", ButtonProps{Variant: Outline, Label: "Choice"})
								}
								root.Width(180).Draw(func(_ *ui.Painter, r ui.Rect) { outer = r })
							}
							if last {
								before()
								choice()
							} else {
								choice()
								before()
							}
						}).Width(259)
					})
				}, 320, 100)
				return r, outer
			}
			candidate, outer := render(true)
			reference, wantOuter := render(false)
			got, ok := candidate.AccessByRole(ui.RoleComboBox, "Choice")
			if !ok {
				t.Fatal("missing popup button")
			}
			if got.HasPopup != ui.PopupDialog {
				t.Fatal("popup native button lacks combobox/dialog semantics")
			}
			want, ok := reference.AccessByRole(ui.RoleButton, "Choice")
			if !ok {
				t.Fatal("missing joined native reference")
			}
			if got.Bounds != want.Bounds || outer != wantOuter {
				t.Fatalf("joined trigger=%+v root=%+v; reference trigger=%+v root=%+v", got.Bounds, outer, want.Bounds, wantOuter)
			}
			if strings.Count(strings.Join(candidate.Texts(), "\n"), "Pick") != 1 {
				t.Fatalf("popup trigger rendered its label more than once: %v", candidate.Texts())
			}
			a, b := candidate.Image(), reference.Image()
			x0, y0, w, h := int(want.Bounds.X), int(want.Bounds.Y), int(want.Bounds.W), int(want.Bounds.H)
			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					for _, p := range []image.Point{{x0 + x, y0 + y}, {x0 + w - 1 - x, y0 + y}, {x0 + x, y0 + h - 1 - y}, {x0 + w - 1 - x, y0 + h - 1 - y}} {
						if a.RGBAAt(p.X, p.Y) != b.RGBAAt(p.X, p.Y) {
							t.Fatalf("joined border/corner pixel %v differs from ordinary native outlined button", p)
						}
					}
				}
			}
		})
	}
}

func TestComboboxUnhighlightedEnterDismissesAndPreservesNativeSubmission(t *testing.T) {
	for _, tc := range []struct {
		name     string
		selected ComboboxValue
		typed    string
		want     string
	}{{"matching-unselected", ComboboxValue{}, "Go", ""}, {"unmatched-unselected", ComboboxValue{}, "not found", ""}, {"restore-selected-label", ComboboxValue{"go", true}, "Rust", "Go language"}} {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.selected
			query := ""
			calls, submits := 0, 0
			r := tester(func(c *ui.Context) {
				root := ComboboxWith(c, &value, &query, ComboboxWithProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go language"}, {Value: "rust", Label: "Rust language"}}, OnValueChange: func(ComboboxValue) { calls++ }})
				// This package-level composition check resolves the existing editor target
				// and reads its real public native Submitted flag; external callers are not
				// claimed to have a child-element lookup API.
				if fieldTarget(c, root).Submitted() {
					submits++
				}
			})
			openSelectionCombo(t, r, "Language")
			r.Key(ui.Cmd, ui.KeyA)
			r.Type(tc.typed)
			if value != tc.selected || query != tc.typed || calls != 0 {
				t.Fatalf("typing changed selected value: %+v q=%q calls%d", value, query, calls)
			}
			for _, n := range r.AccessAll() {
				if n.Role == ui.RoleOption && n.Focused {
					t.Fatal("default unexpectedly highlighted an option")
				}
			}
			r.Key(0, ui.KeyEnter)
			r.Frame()
			if selectionComboInput(t, r, "Language").Expanded || value != tc.selected || calls != 0 {
				t.Fatalf("unhighlighted Enter selected=%+v q=%q want=%q calls%d", value, query, tc.want, calls)
			}
			if submits != 1 {
				t.Fatalf("native Enter submission count=%d want1", submits)
			}
			r.Advance(comboMotionDuration)
			r.Frame()
			if query != tc.want || value != tc.selected || calls != 0 {
				t.Fatalf("completed Enter exit selected=%+v query=%q want=%q calls%d", value, query, tc.want, calls)
			}
		})
	}
	t.Run("multiple", func(t *testing.T) {
		values := []string{"go"}
		query := ""
		submits := 0
		r := tester(func(c *ui.Context) {
			root := MultiComboboxWith(c, &values, &query, ComboboxWithProps{Label: "Language", Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}})
			if fieldTarget(c, root).Submitted() {
				submits++
			}
		})
		openSelectionCombo(t, r, "Language")
		r.Type("Rust")
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if !slices.Equal(values, []string{"go"}) || query != "" || selectionComboInput(t, r, "Language").Expanded || submits != 1 {
			t.Fatalf("unhighlighted multiple Enter values=%v query=%q submits=%d", values, query, submits)
		}
	})
	t.Run("composition", func(t *testing.T) {
		value := ComboboxValue{}
		query := ""
		r := tester(func(c *ui.Context) {
			ComboboxWith(c, &value, &query, ComboboxWithProps{Label: "Language", Options: []Option{{Value: "zh", Label: "你好"}}})
		})
		openSelectionCombo(t, r, "Language")
		r.Compose("你好", 2)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if value.Valid || query != "你好" || !selectionComboInput(t, r, "Language").Expanded {
			t.Fatalf("IME candidate Enter changed choice/dismissed popup: %+v %q", value, query)
		}
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if value.Valid || selectionComboInput(t, r, "Language").Expanded {
			t.Fatalf("normal second Enter failed unhighlighted dismissal: %+v %q", value, query)
		}
		r.Advance(comboMotionDuration)
		r.Frame()
		if query != "" {
			t.Fatalf("completed IME dismissal did not reset query: %q", query)
		}
	})
}
