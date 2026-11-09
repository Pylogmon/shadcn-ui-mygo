package control

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func comboBehaviorOpen(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	for _, n := range r.AccessAll() {
		if n.Name == label && (n.Role == ui.RoleComboBox || n.Role == ui.RoleButton) {
			r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
			r.Advance(120 * time.Millisecond)
			r.Frame()
			return
		}
	}
	t.Fatalf("missing native trigger/editor %q", label)
}

func comboBehaviorActive(r *ui.Tester) string {
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleOption && n.Focused {
			return n.Name
		}
	}
	return ""
}

func TestComboboxSourceExternalEmptyInputClearsOnlySingle(t *testing.T) {
	for _, mode := range []string{"single", "popup", "multiple", "readonly", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			v := ComboboxValue{"go", true}
			values := []string{"go"}
			q := ""
			changes := 0
			r := tester(func(c *ui.Context) {
				p := ComboboxWithProps{Label: "Source", PopupInput: mode == "popup", ReadOnly: mode == "readonly", Disabled: mode == "disabled", Options: []Option{{Value: "go", Label: "Go"}}, OnValueChange: func(ComboboxValue) { changes++ }, OnValuesChange: func([]string) { changes++ }}
				if mode == "multiple" {
					MultiComboboxWith(c, &values, &q, p)
				} else {
					ComboboxWith(c, &v, &q, p)
				}
			})
			comboBehaviorOpen(t, r, "Source")
			if mode == "multiple" {
				r.Type("x")
			}
			r.Key(ui.Cmd, ui.KeyA)
			r.Key(0, ui.KeyBackspace)
			r.Frame()
			if mode == "single" {
				if v.Valid || q != "" || changes != 1 {
					t.Fatalf("empty native edit value=%+v query=%q calls=%d", v, q, changes)
				}
			} else if !v.Valid || !slices.Equal(values, []string{"go"}) || changes != 0 {
				t.Fatalf("%s empty edit mutated selection %+v/%v calls%d", mode, v, values, changes)
			}
		})
	}
}

func TestComboboxSourceDefaultTrimAndCustomRawFilter(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(map[bool]string{false: "default-trim", true: "custom-raw"}[custom], func(t *testing.T) {
			v := ComboboxValue{}
			q := ""
			received := ""
			r := tester(func(c *ui.Context) {
				p := ComboboxWithProps{Label: "Source", Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}}
				if custom {
					p.Filter = func(o Option, raw string) bool { received = raw; return o.Value == "go" }
				}
				ComboboxWith(c, &v, &q, p)
			})
			comboBehaviorOpen(t, r, "Source")
			r.Type("  Go  ")
			r.Frame()
			if _, ok := r.AccessByRole(ui.RoleOption, "Go"); !ok {
				t.Fatalf("whitespace query did not retain Go: %q", q)
			}
			if q != "  Go  " || custom && received != q {
				t.Fatalf("query was normalized in editor/custom callback: q=%q callback=%q", q, received)
			}
		})
	}
}

func TestComboboxSourceCompositionKeepsCommittedResultsAndEvents(t *testing.T) {
	for _, popup := range []bool{false, true} {
		t.Run(map[bool]string{false: "external", true: "popup"}[popup], func(t *testing.T) {
			v := ComboboxValue{"go", true}
			q := ""
			inputs, changes := 0, 0
			r := tester(func(c *ui.Context) {
				ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", PopupInput: popup, AutoHighlight: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}, OnInputChange: func(string) { inputs++ }, OnValueChange: func(ComboboxValue) { changes++ }})
			})
			comboBehaviorOpen(t, r, "Source")
			r.Key(ui.Cmd, ui.KeyA)
			r.Compose("Ru", 2)
			r.Frame()
			if _, ok := r.AccessByRole(ui.RoleOption, "Go"); !ok {
				t.Fatal("preedit prematurely filtered committed list")
			}
			if _, ok := r.AccessByRole(ui.RoleStatus, "No items found."); ok {
				t.Fatal("preedit prematurely displayed empty state")
			}
			if inputs != 0 || changes != 0 || v != (ComboboxValue{"go", true}) {
				t.Fatalf("preedit published events/value: inputs%d changes%d value%+v", inputs, changes, v)
			}
			r.Type("Rust")
			r.Frame()
			if q != "Rust" || inputs != 1 || changes != 0 {
				t.Fatalf("composition commit query=%q inputCalls%d valueCalls%d", q, inputs, changes)
			}
			if _, ok := r.AccessByRole(ui.RoleOption, "Go"); ok {
				t.Fatal("commit did not update filtered results")
			}
			if _, ok := r.AccessByRole(ui.RoleOption, "Rust"); !ok {
				t.Fatal("commit lost matching option")
			}
		})
	}
}

func TestComboboxSourceLoopIncludesInputOnlyInManualMode(t *testing.T) {
	for _, auto := range []bool{false, true} {
		for _, noLoop := range []bool{false, true} {
			t.Run(map[bool]string{false: "manual", true: "auto"}[auto]+map[bool]string{false: "/loop", true: "/no-loop"}[noLoop], func(t *testing.T) {
				v := ComboboxValue{}
				q := ""
				r := tester(func(c *ui.Context) {
					ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", AutoHighlight: auto, NoLoop: noLoop, Options: []Option{{Value: "off", Label: "Unavailable", Disabled: true}, {Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}})
				})
				comboBehaviorOpen(t, r, "Source")
				if got := comboBehaviorActive(r); got != "" {
					t.Fatalf("pointer open unselected active=%q", got)
				}
				want := []string{"Alpha", "Beta", "", "Alpha"}
				if auto {
					want = []string{"Alpha", "Beta", "Alpha", "Beta"}
				}
				if noLoop {
					want = []string{"Alpha", "Beta", "Beta", "Beta"}
				}
				for i, label := range want {
					r.Key(0, ui.KeyDown)
					r.Frame()
					if got := comboBehaviorActive(r); got != label {
						t.Fatalf("Down%d active=%q want%q", i+1, got, label)
					}
				}
			})
		}
	}
}

func TestComboboxSourceOpensAtSelectedValueAndReveals(t *testing.T) {
	for _, openBy := range []string{"pointer", "down", "up"} {
		t.Run(openBy, func(t *testing.T) {
			v := ComboboxValue{"item-29", true}
			q := ""
			options := make([]Option, 30)
			for i := range options {
				options[i] = Option{Value: "item-" + strings.Repeat("x", i)}
			}
			options[29] = Option{Value: "item-29", Label: "Selected last"}
			r := tester(func(c *ui.Context) {
				ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", MaxHeight: 150, Options: options})
			})
			if openBy == "pointer" {
				comboBehaviorOpen(t, r, "Source")
			} else {
				n := selectionComboInput(t, r, "Source")
				r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
				r.Key(0, ui.KeyEscape)
				r.Advance(120 * time.Millisecond)
				key := ui.KeyDown
				if openBy == "up" {
					key = ui.KeyUp
				}
				r.Key(0, key)
				r.Advance(120 * time.Millisecond)
				r.Frame()
			}
			if got := comboBehaviorActive(r); got != "Selected last" {
				t.Fatalf("%s open active=%q", openBy, got)
			}
			n, ok := r.AccessByRole(ui.RoleOption, "Selected last")
			list, _ := r.AccessByRole(ui.RoleListBox, "Source options")
			if !ok || n.Offscreen || n.Bounds.Y < list.Bounds.Y || n.Bounds.Y+n.Bounds.H > list.Bounds.Y+list.Bounds.H+.1 {
				t.Fatalf("selected option not revealed: %+v list%+v", n, list)
			}
		})
	}
}

func TestComboboxSourcePointerLeaveResetsOnlyPointerHighlight(t *testing.T) {
	for _, auto := range []bool{false, true} {
		t.Run(map[bool]string{false: "manual", true: "auto"}[auto], func(t *testing.T) {
			v := ComboboxValue{}
			q := ""
			r := tester(func(c *ui.Context) {
				ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", AutoHighlight: auto, Options: []Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}})
			})
			comboBehaviorOpen(t, r, "Source")
			b, _ := r.AccessByRole(ui.RoleOption, "Beta")
			r.Move(b.Bounds.X+12, b.Bounds.Y+b.Bounds.H/2)
			r.Frame()
			if comboBehaviorActive(r) != "Beta" {
				t.Fatal("pointer did not highlight Beta")
			}
			r.Move(390, 340)
			r.Frame()
			if got := comboBehaviorActive(r); got != "" {
				t.Fatalf("pointer leave retained %q", got)
			}
			r.Key(0, ui.KeyDown)
			r.Frame()
			r.Move(389, 340)
			r.Frame()
			if comboBehaviorActive(r) != "Alpha" {
				t.Fatal("outside pointer movement cleared keyboard highlight")
			}
		})
	}
}

func TestComboboxSourceChipPrintableKeysUseNativeEditor(t *testing.T) {
	for _, tc := range []struct {
		name string
		mods ui.Modifiers
		key  ui.Key
		text string
	}{{"digit", 0, ui.Key1, "1"}, {"symbol", ui.Shift, ui.Key1, "!"}, {"punctuation", 0, ui.KeySlash, "/"}, {"unicode", 0, ui.KeyUnknown, "中"}} {
		t.Run(tc.name, func(t *testing.T) {
			values := []string{"go"}
			q := ""
			r := tester(func(c *ui.Context) {
				MultiComboboxWith(c, &values, &q, ComboboxWithProps{Label: "Source", Options: []Option{{Value: "go", Label: "Go"}}})
			})
			comboBehaviorOpen(t, r, "Source")
			r.Key(0, ui.KeyHome)
			r.Key(0, ui.KeyLeft)
			r.Frame()
			if !r.Focused("Go") {
				t.Fatal("chip did not focus")
			}
			r.TypeKey(tc.mods, tc.key, tc.text)
			r.Frame()
			if q != tc.text || !r.Focused("Source") || !slices.Equal(values, []string{"go"}) {
				t.Fatalf("printable key lost native input q=%q selected=%v focus=%v", q, values, r.AccessAll())
			}
			r.Key(ui.Cmd, ui.KeyZ)
			r.Frame()
			if q != "" {
				t.Fatalf("native undo after chip typing=%q", q)
			}
		})
	}
}

func TestComboboxSourcePopupTriggerAXAndTypeahead(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	calls := 0
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", PopupInput: true, Options: []Option{{Value: "react", Label: "React"}, {Value: "rust", Label: "Rust"}, {Value: "go", Label: "Go"}}, OnValueChange: func(ComboboxValue) { calls++ }})
	})
	n, ok := r.AccessByRole(ui.RoleComboBox, "Source")
	if !ok || n.HasPopup != ui.PopupDialog {
		t.Fatalf("popup trigger source AX=%+v all%v", n, r.AccessAll())
	}
	comboBehaviorOpen(t, r, "Source")
	if _, ok := r.AccessByRole(ui.RoleDialog, ""); !ok {
		found := false
		for _, a := range r.AccessAll() {
			found = found || a.Role == ui.RoleDialog
		}
		if !found {
			t.Fatal("popup input lacks dialog container")
		}
	}
	r.Key(0, ui.KeyEscape)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	r.TypeKey(0, ui.KeyR, "r")
	r.Frame()
	if v.Value != "react" || q != "" || calls != 1 {
		t.Fatalf("closed typeahead v%+v q%q calls%d", v, q, calls)
	}
	r.TypeKey(0, ui.KeyR, "r")
	r.Frame()
	if v.Value != "rust" || calls != 2 {
		t.Fatalf("rapid first-letter cycle %+v calls%d", v, calls)
	}
	r.Advance(751 * time.Millisecond)
	r.TypeKey(0, ui.KeyG, "g")
	r.Frame()
	if v.Value != "go" || calls != 3 {
		t.Fatalf("expired typing session %+v calls%d", v, calls)
	}
	n, _ = r.AccessByRole(ui.RoleComboBox, "Source")
	if n.Expanded {
		t.Fatal("closed typeahead opened popup")
	}
}

func TestComboboxSourceFilteredCollectionsHaveNoOrphanSeparators(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", Options: []Option{{Value: "a", Label: "Alpha", Group: "First"}, {Separator: true}, {Value: "b", Label: "Beta", Group: "Second"}, {Separator: true}, {Value: "c", Label: "Charlie", Group: "Third"}}})
	})
	comboBehaviorOpen(t, r, "Source")
	r.Type("zz")
	r.Frame()
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleSeparator || n.Role == ui.RoleGroup && (n.Name == "First" || n.Name == "Second" || n.Name == "Third") {
			t.Fatalf("empty filtered collection retained %+v", n)
		}
	}
	if _, ok := r.AccessByRole(ui.RoleStatus, "No items found."); !ok {
		t.Fatal("empty status missing")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Beta")
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleGroup, "Second"); !ok {
		t.Fatal("matching group missing")
	}
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleSeparator {
			t.Fatal("single remaining group retained orphan separator")
		}
	}
}

func TestComboboxSourceCompositionRetainsPreviouslyFilteredCollection(t *testing.T) {
	for _, popup := range []bool{false, true} {
		t.Run(map[bool]string{false: "external", true: "popup"}[popup], func(t *testing.T) {
			v := ComboboxValue{}
			q := ""
			inputs := []string{}
			r := tester(func(c *ui.Context) {
				ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", PopupInput: popup, AutoHighlight: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}, OnInputChange: func(value string) { inputs = append(inputs, value) }})
			})
			comboBehaviorOpen(t, r, "Source")
			r.Type("Ru")
			r.Key(ui.Cmd, ui.KeyA)
			r.Compose("Go", 2)
			r.Frame()
			if !slices.Equal(inputs, []string{"Ru"}) {
				t.Fatalf("preedit emitted changes %v", inputs)
			}
			if _, ok := r.AccessByRole(ui.RoleOption, "Rust"); !ok {
				t.Fatal("preedit discarded last committed result")
			}
			if _, ok := r.AccessByRole(ui.RoleOption, "Go"); ok {
				t.Fatal("preedit introduced uncommitted result")
			}
			r.Type("Go")
			r.Frame()
			if !slices.Equal(inputs, []string{"Ru", "Go"}) {
				t.Fatalf("commit events=%v", inputs)
			}
			if _, ok := r.AccessByRole(ui.RoleOption, "Go"); !ok {
				t.Fatal("commit did not replace filtered collection")
			}
		})
	}
}

func TestComboboxSourceCutClearsSingleValueOnce(t *testing.T) {
	v := ComboboxValue{"go", true}
	q := ""
	changes := 0
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", Options: []Option{{Value: "go", Label: "Go"}}, OnValueChange: func(ComboboxValue) { changes++ }})
	})
	comboBehaviorOpen(t, r, "Source")
	r.Key(ui.Cmd, ui.KeyA)
	r.Command("cut")
	r.Frame()
	if v.Valid || q != "" || changes != 1 {
		t.Fatalf("native cut value%+v query%q calls%d", v, q, changes)
	}
	r.Frame()
	if changes != 1 {
		t.Fatal("idle repeated empty-input selection event")
	}
}

func TestComboboxSourcePointerPaddingAndTouch(t *testing.T) {
	for _, touch := range []bool{false, true} {
		t.Run(map[bool]string{false: "mouse-padding", true: "touch-no-hover"}[touch], func(t *testing.T) {
			v := ComboboxValue{}
			q := ""
			r := tester(func(c *ui.Context) {
				ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", Options: []Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}})
			})
			comboBehaviorOpen(t, r, "Source")
			if touch {
				r.UsePointer(ui.PointerTouch, 0)
			}
			n, _ := r.AccessByRole(ui.RoleOption, "Beta")
			r.Move(n.Bounds.X+12, n.Bounds.Y+n.Bounds.H/2)
			r.Frame()
			if touch {
				if comboBehaviorActive(r) != "" {
					t.Fatal("touch pointer acquired hover highlight")
				}
				return
			}
			if comboBehaviorActive(r) != "Beta" {
				t.Fatal("mouse did not highlight item")
			}
			list, _ := r.AccessByRole(ui.RoleListBox, "Source options")
			r.Move(list.Bounds.X+1, n.Bounds.Y+n.Bounds.H/2)
			r.Frame()
			if got := comboBehaviorActive(r); got != "" {
				t.Fatalf("item leave into list padding retained %q", got)
			}
		})
	}
}

func TestComboboxSourcePopupTypeaheadPrefixSpaceCaseAndExpiry(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", PopupInput: true, Options: []Option{{Value: "nh", Label: "New Hampshire"}, {Value: "ny", Label: "New York"}, {Value: "go", Label: "Go"}}})
	})
	r.Key(0, ui.KeyTab)
	r.TypeKey(ui.Shift, ui.KeyN, "N")
	r.TypeKey(0, ui.KeyE, "e")
	r.TypeKey(0, ui.KeyW, "w")
	r.TypeKey(0, ui.KeySpace, " ")
	r.TypeKey(0, ui.KeyY, "y")
	r.Frame()
	if v.Value != "ny" || selectionComboInput(t, r, "Source").Expanded || q != "" {
		t.Fatalf("multi-key closed prefix failed v%+v q%q", v, q)
	}
	r.Advance(750 * time.Millisecond)
	r.Key(0, ui.KeySpace)
	r.Frame()
	if !selectionComboInput(t, r, "Source").Expanded {
		t.Fatal("expired typeahead session swallowed native Space activation")
	}
}

func TestComboboxSourcePopupTypeaheadKeepsBaseDisabledLabelRule(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	calls := 0
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", PopupInput: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust", Disabled: true}, {Value: "swift", Label: "Swift"}}, OnValueChange: func(ComboboxValue) { calls++ }})
	})
	r.Key(0, ui.KeyTab)
	r.TypeKey(0, ui.KeyR, "r")
	r.Frame()
	if v != (ComboboxValue{"rust", true}) || calls != 1 {
		t.Fatalf("closed Base typeahead should not inherit open-list disabledIndices: %+v calls%d", v, calls)
	}
	comboBehaviorOpen(t, r, "Source")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if calls != 1 {
		t.Fatal("open disabled option accepted Enter")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if v.Value != "swift" || calls != 2 {
		t.Fatalf("open navigation did not skip disabled value %+v calls%d", v, calls)
	}
}

func TestComboboxSourcePopupTypeaheadReadonlyDisabledMultipleAndIME(t *testing.T) {
	for _, mode := range []string{"readonly", "disabled", "multiple", "ime"} {
		t.Run(mode, func(t *testing.T) {
			v := ComboboxValue{}
			values := []string{}
			q := ""
			calls := 0
			r := tester(func(c *ui.Context) {
				p := ComboboxWithProps{Label: "Source", PopupInput: true, ReadOnly: mode == "readonly", Disabled: mode == "disabled", Options: []Option{{Value: "rust", Label: "Rust"}}, OnValueChange: func(ComboboxValue) { calls++ }, OnValuesChange: func([]string) { calls++ }}
				if mode == "multiple" {
					MultiComboboxWith(c, &values, &q, p)
				} else {
					ComboboxWith(c, &v, &q, p)
				}
			})
			r.Key(0, ui.KeyTab)
			if mode == "ime" {
				r.Compose("r", 1)
				r.Type("r")
			} else {
				r.TypeKey(0, ui.KeyR, "r")
			}
			r.Frame()
			if v.Valid || len(values) != 0 || q != "" || calls != 0 {
				t.Fatalf("%s triggered typeahead v%+v values%v query%q calls%d", mode, v, values, q, calls)
			}
		})
	}
}

func TestComboboxSourcePopupTypeaheadDoubleLetterLabelsDisableCycling(t *testing.T) {
	v := ComboboxValue{}
	q := ""
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &v, &q, ComboboxWithProps{Label: "Source", PopupInput: true, Options: []Option{{Value: "react", Label: "React"}, {Value: "rust", Label: "Rust"}, {Value: "llama", Label: "Llama"}}})
	})
	r.Key(0, ui.KeyTab)
	r.TypeKey(0, ui.KeyR, "r")
	r.TypeKey(0, ui.KeyR, "r")
	r.Frame()
	if v.Value != "react" {
		t.Fatalf("double-letter label did not disable rapid cycling: %+v", v)
	}
	r.TypeKey(0, ui.KeyL, "l")
	r.TypeKey(0, ui.KeyL, "l")
	r.Frame()
	if v.Value != "llama" {
		t.Fatalf("double-letter prefix did not match: %+v", v)
	}
}
