package control

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func selectSourceOptions(n int) []Option {
	options := make([]Option, n)
	for i := range options {
		options[i] = Option{Value: fmt.Sprint(i), Label: fmt.Sprintf("Item %02d", i)}
	}
	return options
}
func selectSourceOpen(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	n, ok := r.AccessByRole(ui.RoleComboBox, label)
	if !ok {
		t.Fatalf("missing no-editor trigger: %v", r.AccessAll())
	}
	r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
	r.Frame()
}
func selectSourceOption(t *testing.T, r *ui.Tester, label string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(ui.RoleOption, label)
	if !ok {
		t.Fatalf("missing %q: %v", label, r.AccessAll())
	}
	return n
}

func TestSelectSourceNoEditorMultipleAndIndependentSlots(t *testing.T) {
	state := SelectState{}
	changes := 0
	closeOnChange := false
	var submitted []string
	var derivedValues []string
	var scope *SelectScope
	r := tester(func(c *ui.Context) {
		SelectRoot(c, &state, SelectRootProps{Options: selectSourceOptions(3), Multiple: true, OnValuesChange: func(_ []string, _ *SelectChange) {
			changes++
			if closeOnChange {
				state.Open = false
			}
		}}, func(b *SelectScope) {
			scope = b
			ui.Row(c).Children(func() {
				b.Trigger(c, SelectTriggerProps{Label: "Fruit"}, func() {
					b.Value(c, SelectValueProps{Placeholder: "Choose fruits"}, func(v SelectValueView) { ui.Text(c, fmt.Sprintf("%d selected", len(v.Values))) })
				}).Width(200)
				if Button(c, "Submit", ButtonProps{}).Clicked() {
					state.Values = []string{"2"}
					submitted = b.FormValues()
				}
			})
			b.Content(c, SelectContentProps{}, func() {
				b.List(c, SelectListProps{Label: "Fruits"}, func() { b.Group(c, "Fresh", func() { b.Collection(c, nil) }) })
				b.ScrollUpButton(c)
				b.ScrollDownButton(c)
			})
		})
		// Like the calendar caption, this binding is reconstructed on every
		// build. Convenience widgets retain their native identity independently
		// of that temporary address; Changed publishes the derived values.
		derived := slices.Clone(derivedValues)
		if MultiSelect(c, &derived, SelectProps{Label: "Derived selection", Options: selectSourceOptions(3)}).Changed() {
			derivedValues = slices.Clone(derived)
		}
	})
	selectSourceOpen(t, r, "Fruit")
	if n, _ := r.AccessByRole(ui.RoleComboBox, "Fruit"); n.HasPopup != ui.PopupListBox || !n.Expanded {
		t.Fatalf("trigger semantics %+v", n)
	}
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleTextField || n.Role == ui.RoleDialog || n.AutoComplete == ui.AutoCompleteList {
			t.Fatalf("no-editor Select introduced editor/dialog: %+v", n)
		}
	}
	list, ok := r.AccessByRole(ui.RoleListBox, "Fruits")
	if !ok || !list.Multiselectable {
		t.Fatalf("missing multiple listbox %+v", list)
	}
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyEnter)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeySpace)
	r.Frame()
	if !state.Open || !slices.Equal(state.Values, []string{"0", "1"}) || changes != 2 {
		t.Fatalf("multi selection %+v calls%d", state, changes)
	}
	if !selectSourceOption(t, r, "Item 00").Selected || !selectSourceOption(t, r, "Item 01").Selected {
		t.Fatal("actual selected options not exposed")
	}
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(state.Values, []string{"1"}) || changes != 3 {
		t.Fatalf("toggle %+v calls%d", state, changes)
	}
	closeOnChange = true
	item := selectSourceOption(t, r, "Item 02")
	r.ClickAt(item.Bounds.X+8, item.Bounds.Y+item.Bounds.H/2)
	r.Frame()
	if state.Open {
		t.Fatal("application-controlled close from selection callback did not apply")
	}
	if err := r.Click("Submit"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !slices.Equal(submitted, []string{"2"}) {
		t.Fatalf("form read did not observe same-build application binding: %v", submitted)
	}
	selectSourceOpen(t, r, "Derived selection")
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(derivedValues, []string{"0"}) {
		t.Fatalf("temporary multiple binding lost first native choice: %v", derivedValues)
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	trigger, _ := r.AccessByRole(ui.RoleComboBox, "Derived selection")
	if !slices.Equal(derivedValues, []string{"0", "1"}) || !trigger.Expanded {
		t.Fatalf("temporary multiple binding lost open/values: %v expanded%v", derivedValues, trigger.Expanded)
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	defer func() {
		if recover() == nil {
			t.Fatal("expired scope accepted form access")
		}
	}()
	scope.FormValues()
}

func TestSelectSourceMenuProfilePreservesMeasuredItem(t *testing.T) {
	state := SelectState{}
	var runtime *selectRuntime
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		theme.MenuScope(c, theme.Preset("blue", false), theme.MenuOptions{Color: theme.MenuInvertedTranslucent, InvertedTokens: theme.Preset("blue", true)}).Fill().Padding(20).Gap(20).Children(func() {
			ui.Box(c).Height(36)
			SelectRoot(c, &state, SelectRootProps{Options: []Option{{Value: "go", Label: "Go"}}}, func(b *SelectScope) {
				runtime = b.f.s
				b.Trigger(c, SelectTriggerProps{Label: "Profile"}, nil)
				b.Content(c, SelectContentProps{}, func() {
					b.List(c, SelectListProps{}, func() {
						b.Group(c, "", func() {
							b.Item(c, "go", SelectItemProps{}, func() { ui.Box(c).Size(16, 16).Background(ui.RGB(255, 0, 0)) })
						})
					})
				})
			})
		})
	}, 400, 320)
	selectSourceOpen(t, r, "Profile")
	if area := menuRedArea(r.Image()); area < 250 {
		t.Fatalf("measured item clipped: area%d panel%+v items%+v content%v alignment%v,%v/%v", area, runtime.panelBox, runtime.geometry, runtime.contentHeight, runtime.alignX, runtime.alignY, runtime.alignHeight)
	}
}

func TestSelectSourceLegacyProfileAfterCombobox(t *testing.T) {
	value, query, selected := ComboboxValue{}, "", ""
	var box ui.Rect
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		theme.MenuScope(c, theme.Preset("blue", false), theme.MenuOptions{Color: theme.MenuInvertedTranslucent, InvertedTokens: theme.Preset("blue", true)}).Fill().Padding(20).Gap(20).Children(func() {
			ComboboxWith(c, &value, &query, ComboboxWithProps{Label: "Combo", AutoHighlight: true, Options: []Option{{Value: "go", Label: "Go"}}})
			Select(c, &selected, SelectProps{Label: "Legacy", Options: []Option{{Value: "go", Label: "Go"}}, ItemContent: func(c *ui.Context, _ Option) {
				ui.Box(c).Size(16, 16).Background(ui.RGB(255, 0, 0)).Draw(func(_ *ui.Painter, b ui.Rect) { box = b })
			}})
		})
	}, 400, 320)
	if err := r.Click("Combo"); err != nil {
		t.Fatal(err)
	}
	r.Type("Go")
	r.Key(0, ui.KeyEnter)
	if err := r.Click("Legacy"); err != nil {
		t.Fatal(err)
	}
	if area := menuRedArea(r.Image()); area < 250 {
		option, _ := r.AccessByRole(ui.RoleOption, "Go")
		list, _ := r.AccessByRole(ui.RoleListBox, "")
		trigger, _ := r.AccessByRole(ui.RoleComboBox, "Legacy")
		t.Fatalf("legacy clipped area%d box%+v option%+v list%+v trigger%+v", area, box, option.Bounds, list.Bounds, trigger.Bounds)
	}
}

func TestSelectSourceLargeListRevealAndHoverArrows(t *testing.T) {
	state := SelectState{Value: SelectSelection{"80", true}}
	falseValue := false
	var scroll ui.ScrollState
	var runtime *selectRuntime
	r := tester(func(c *ui.Context) {
		SelectRoot(c, &state, SelectRootProps{Options: selectSourceOptions(100)}, func(b *SelectScope) {
			runtime = b.f.s
			b.Trigger(c, SelectTriggerProps{Label: "Large"}, nil).Width(180)
			b.Content(c, SelectContentProps{AlignItemWithTrigger: &falseValue, MaxHeight: 180}, func() {
				b.List(c, SelectListProps{Scroll: &scroll}, func() { b.Group(c, "", func() { b.Collection(c, nil) }) })
				b.ScrollUpButton(c)
				b.ScrollDownButton(c)
			})
		})
	})
	selectSourceOpen(t, r, "Large")
	r.Advance(120 * time.Millisecond)
	r.Frame()
	n := selectSourceOption(t, r, "Item 80")
	list, _ := r.AccessByRole(ui.RoleListBox, "")
	if scroll.Y <= 0 || n.Bounds.Y < list.Bounds.Y || n.Bounds.Y+n.Bounds.H > list.Bounds.Y+list.Bounds.H {
		t.Fatalf("selected item not revealed item%+v list%+v scroll%+v", n.Bounds, list.Bounds, scroll)
	}
	r.Key(0, ui.KeyHome)
	r.Frame()
	before := scroll.Y
	p := runtime.panelBox
	r.Move(p.X+p.W/2, p.Y+p.H-12)
	r.Advance(90 * time.Millisecond)
	r.Frame()
	if scroll.Y <= before || runtime.hasActive {
		t.Fatalf("hover arrow did not scroll one item/clear highlight: before%v after%v active%v", before, scroll.Y, runtime.hasActive)
	}
	r.Move(390, 350)
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if state.Value != (SelectSelection{"99", true}) || state.Open {
		t.Fatalf("End choose %+v", state)
	}
}

func TestSelectSourceClosedTypeaheadDisabledAndComposition(t *testing.T) {
	state := SelectState{Value: SelectSelection{"a", true}}
	options := []Option{{Value: "a", Label: "Apple"}, {Value: "ap", Label: "Apricot", Disabled: true}, {Value: "b", Label: "Blueberry"}, {Value: "bl", Label: "Blackberry"}, {Value: "n", Label: "New York"}}
	readOnly, disabled := false, false
	calls := 0
	r := tester(func(c *ui.Context) {
		SelectWith(c, &state, SelectWithProps{Root: SelectRootProps{Options: options, ReadOnly: readOnly, Disabled: disabled, OnValueChange: func(_ SelectSelection, _ *SelectChange) { calls++ }}, Trigger: SelectTriggerProps{Label: "Typeahead"}}).Width(200)
	})
	n, _ := r.AccessByRole(ui.RoleComboBox, "Typeahead")
	r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
	r.Key(0, ui.KeyEscape)
	r.Type("b")
	r.Type("l")
	r.Frame()
	if state.Value.Value != "b" || state.Open || calls != 2 {
		t.Fatalf("closed prefix selected/opened incorrectly %+v calls%d", state, calls)
	}
	r.Advance(760 * time.Millisecond)
	r.Type("b")
	r.Frame()
	if state.Value.Value != "bl" {
		t.Fatalf("750ms prefix reset/cycle selected %+v", state)
	}
	r.Advance(760 * time.Millisecond)
	r.Type("a")
	r.Frame()
	if state.Value.Value != "a" {
		t.Fatalf("closed typeahead selected disabled Apricot: %+v", state)
	}
	old := calls
	r.Advance(760 * time.Millisecond)
	r.Compose("Bl", 2)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if state.Open || calls != old || state.Value.Value != "a" {
		t.Fatal("IME preedit/Enter activated the no-editor trigger")
	}
	r.Type("Bl")
	r.Frame()
	if state.Value.Value != "b" || state.Open {
		t.Fatalf("IME commit failed source typeahead %+v", state)
	}
	readOnly = true
	r.Frame()
	old = calls
	r.Type("n")
	r.Key(0, ui.KeyDown)
	r.Frame()
	if state.Open || calls != old {
		t.Fatal("readonly trigger changed")
	}
	readOnly = false
	disabled = true
	r.Frame()
	r.Type("a")
	r.Key(0, ui.KeyDown)
	r.Frame()
	if state.Open || calls != old {
		t.Fatal("disabled trigger changed")
	}
	// Base deliberately permits focus on an aria-disabled option after open,
	// while its initial arrow opening and closed typeahead skip it.
	disabled = false
	state.Value = SelectSelection{"a", true}
	r.Frame()
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyDown)
	r.Frame()
	if n := selectSourceOption(t, r, "Apricot"); !n.Focused || !n.Disabled {
		t.Fatalf("source disabled focus contract %+v", n)
	}
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if state.Value.Value != "a" || !state.Open {
		t.Fatal("disabled focused item committed")
	}
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if state.Value.Value != "n" || state.Open {
		t.Fatalf("source non-loop end %+v", state)
	}

	// A custom native trigger keeps one half-opacity disabled layer, while
	// both explicit and ancestor disablement still reject real native input.
	var pixelState SelectState
	explicit, inherited := false, false
	pixelTester := ui.NewTester(func(c *ui.Context) {
		tokens := theme.Preset("blue", false)
		theme.Scope(c, tokens).Fill().Padding(20).Background(ui.RGB(0, 0, 0)).Children(func() {
			ui.Column(c).Disabled(inherited).Children(func() {
				SelectRoot(c, &pixelState, SelectRootProps{Options: selectSourceOptions(2), Disabled: explicit}, func(b *SelectScope) {
					b.Trigger(c, SelectTriggerProps{Label: "Disabled pixels", Render: func(current *ui.Context) ui.Element {
						return ui.ButtonBase(current).Size(180, 36).Background(ui.RGB(255, 255, 255)).Border(0, ui.Transparent).FocusRing(false)
					}}, nil)
					b.Content(c, SelectContentProps{}, func() {
						b.List(c, SelectListProps{}, func() { b.Group(c, "", func() { b.Collection(c, nil) }) })
					})
				})
			})
		})
	}, 220, 100)
	assertPixels := func(want uint8) {
		t.Helper()
		a, ok := pixelTester.AccessByRole(ui.RoleComboBox, "Disabled pixels")
		if !ok {
			t.Fatal("missing actual native select trigger for disabled pixels")
		}
		p := color.NRGBAModel.Convert(pixelTester.Image().At(int(a.Bounds.X+10), int(a.Bounds.Y+10))).(color.NRGBA)
		if p.R != want || p.G != want || p.B != want || p.A != 255 {
			t.Fatalf("explicit=%v inherited=%v pixel=%v want=%d", explicit, inherited, p, want)
		}
		if want == 128 {
			if !a.Disabled {
				t.Fatal("disabled trigger lost its native disabled state")
			}
			pixelTester.ClickAt(a.Bounds.X+8, a.Bounds.Y+a.Bounds.H/2)
			pixelTester.Key(0, ui.KeyDown)
			pixelTester.Frame()
			if pixelState.Open || pixelState.Value.Valid {
				t.Fatal("disabled native trigger accepted pointer or keyboard input")
			}
		}
	}
	assertPixels(255)
	explicit = true
	pixelTester.Frame()
	assertPixels(128)
	explicit, inherited = false, true
	pixelTester.Frame()
	assertPixels(128)
}

func TestSelectSourceControlledCancellationTabAndMount(t *testing.T) {
	state := SelectState{}
	show, cancelOpen, cancelValue := true, true, true
	version := 1
	calls := []int{}
	var runtime, old *selectRuntime
	var expired *SelectScope
	manualDisabled := false
	r := tester(func(c *ui.Context) {
		Button(c, "Before", ButtonProps{})
		if show {
			v := version
			SelectRoot(c, &state, SelectRootProps{OnOpenChange: func(_ bool, e *SelectChange) {
				if cancelOpen {
					e.Cancel()
				}
			}, OnValueChange: func(_ SelectSelection, e *SelectChange) {
				calls = append(calls, v)
				if cancelValue {
					e.Cancel()
				}
			}}, func(b *SelectScope) {
				runtime = b.f.s
				expired = b
				b.Trigger(c, SelectTriggerProps{Label: "Manual"}, nil)
				b.Content(c, SelectContentProps{}, func() {
					b.List(c, SelectListProps{Label: "Manual items"}, func() {
						b.Group(c, "Choices", func() {
							b.Item(c, "zero", SelectItemProps{Label: "Zero", Disabled: manualDisabled}, nil)
							b.Item(c, "one", SelectItemProps{Label: "One"}, nil)
						})
					})
				})
			})
		}
		Button(c, "After", ButtonProps{})
	})
	selectSourceOpen(t, r, "Manual")
	if state.Open {
		t.Fatal("open cancellation ignored")
	}
	cancelOpen = false
	selectSourceOpen(t, r, "Manual")
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if state.Value.Valid || state.Open || !slices.Equal(calls, []int{1}) {
		t.Fatalf("value cancellation/source close %+v calls%v", state, calls)
	}
	cancelValue = false
	version = 2
	r.Frame()
	selectSourceOpen(t, r, "Manual")
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if state.Value != (SelectSelection{"one", true}) || !slices.Equal(calls, []int{1, 2}) {
		t.Fatalf("current frame callback %+v %v", state, calls)
	}
	selectSourceOpen(t, r, "Manual")
	for _, name := range []string{"Zero", "One"} {
		n := selectSourceOption(t, r, name)
		if n.SetSize != 2 {
			t.Fatalf("manual first-build set size %+v", n)
		}
	}
	focusBefore := []string{}
	for _, n := range r.AccessAll() {
		if n.Focused {
			focusBefore = append(focusBefore, n.Name)
		}
	}
	r.Key(0, ui.KeyTab)
	r.Frame()
	after, _ := r.AccessByRole(ui.RoleButton, "After")
	if state.Open || !after.Focused {
		names := []string{}
		for _, n := range r.AccessAll() {
			if n.Focused {
				names = append(names, n.Name)
			}
		}
		t.Fatalf("Tab did not leave popup to following control open%v after%v focus%v prior%v", state.Open, after.Focused, names, focusBefore)
	}
	selectSourceOpen(t, r, "Manual")
	r.Key(ui.Shift, ui.KeyTab)
	r.Frame()
	before, _ := r.AccessByRole(ui.RoleButton, "Before")
	if state.Open || !before.Focused {
		t.Fatalf("Shift+Tab did not leave before trigger open%v before%v", state.Open, before.Focused)
	}
	manualDisabled = true
	r.Frame()
	selectSourceOpen(t, r, "Manual")
	if !selectSourceOption(t, r, "Zero").Disabled {
		t.Fatal("manual disabled metadata stale")
	}
	r.Key(0, ui.KeyEscape)
	manualDisabled = false
	r.Frame()
	selectSourceOpen(t, r, "Manual")
	if selectSourceOption(t, r, "Zero").Disabled {
		t.Fatal("manual disabled state latched from old frame")
	}
	r.Key(0, ui.KeyEscape)
	old = runtime
	show = false
	r.Frame()
	r.Frame()
	if !old.closed {
		t.Fatal("removed mount retained its runtime")
	}
	show = true
	r.Frame()
	if runtime == old || runtime.closed {
		t.Fatal("remount reused closed runtime")
	}
	for _, call := range []func(){func() { expired.FormValues() }, func() { expired.Value(nil, SelectValueProps{}, nil) }, func() { expired.Collection(nil, nil) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expired scope accepted a public slot")
				}
			}()
			call()
		}()
	}
}

func TestSelectSourceSidesLogicalAlignmentAndItemAlignment(t *testing.T) {
	falseValue := false
	for _, side := range []SelectSide{SelectSideTop, SelectSideBottom, SelectSideLeft, SelectSideRight, SelectSideInlineStart, SelectSideInlineEnd} {
		t.Run(string(side), func(t *testing.T) {
			state := SelectState{Open: true}
			var panel, trigger ui.Rect
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				ui.Box(c).Absolute().Left(300).Top(220).Children(func() {
					SelectRoot(c, &state, SelectRootProps{Options: selectSourceOptions(3)}, func(b *SelectScope) {
						tr := b.Trigger(c, SelectTriggerProps{Label: "Side"}, nil).Width(160)
						p := b.Content(c, SelectContentProps{Side: side, Align: SelectAlignStart, AlignItemWithTrigger: &falseValue, AlignOffset: 7}, func() { b.List(c, SelectListProps{}, func() { b.Group(c, "", func() { b.Collection(c, nil) }) }) })
						trigger = tr.LayoutBounds()
						if p.Valid() {
							panel = p.LayoutBounds()
						}
					})
				})
			}, 800, 600, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			r.SetDirection(ui.RightToLeft)
			r.Frame()
			r.Frame()
			physical := comboPhysicalSide(ComboboxSide(side), ui.RightToLeft)
			if panel.W != trigger.W {
				t.Fatalf("anchor width panel%+v trigger%+v", panel, trigger)
			}
			switch physical {
			case ComboboxSideTop:
				if panel.Y+panel.H > trigger.Y-3 {
					t.Fatalf("top %+v %+v", panel, trigger)
				}
			case ComboboxSideBottom:
				if panel.Y < trigger.Y+trigger.H+3 {
					t.Fatalf("bottom %+v %+v", panel, trigger)
				}
			case ComboboxSideLeft:
				if panel.X+panel.W > trigger.X-3 {
					t.Fatalf("left %+v %+v", panel, trigger)
				}
			case ComboboxSideRight:
				if panel.X < trigger.X+trigger.W+3 {
					t.Fatalf("right %+v %+v", panel, trigger)
				}
			}
		})
	}
	state := SelectState{Value: SelectSelection{"2", true}, Open: true}
	var runtime *selectRuntime
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Box(c).Absolute().Left(150).Top(200).Children(func() {
			SelectRoot(c, &state, SelectRootProps{Options: selectSourceOptions(5)}, func(b *SelectScope) {
				runtime = b.f.s
				b.Trigger(c, SelectTriggerProps{Label: "Aligned"}, nil).Width(200)
				b.Content(c, SelectContentProps{}, func() { b.List(c, SelectListProps{}, func() { b.Group(c, "", func() { b.Collection(c, nil) }) }) })
			})
		})
	}, 600, 500)
	r.Frame()
	r.Frame()
	selected := selectSourceOption(t, r, "Item 02")
	if math.Abs(float64(selected.Bounds.Y+selected.Bounds.H/2-runtime.valueBox.Y-runtime.valueBox.H/2)) > 1.1 {
		t.Fatalf("selected row not aligned to actual value: item%+v value%+v", selected.Bounds, runtime.valueBox)
	}
	if runtime.motion.from != comboPoseShown || runtime.motion.to != comboPoseShown {
		t.Fatal("item-aligned source popup animated")
	}
}

func TestSelectSourceAlignedLongListExpandsAndPreservesNativeReveal(t *testing.T) {
	state := SelectState{Value: SelectSelection{"0", true}, Open: true}
	var runtime *selectRuntime
	var scroll ui.ScrollState
	var contentMaxHeight float32
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Box(c).Absolute().Left(160).Top(200).Children(func() {
			SelectRoot(c, &state, SelectRootProps{Options: selectSourceOptions(100)}, func(b *SelectScope) {
				runtime = b.f.s
				b.Trigger(c, SelectTriggerProps{Label: "Long aligned"}, nil).Width(180)
				b.Content(c, SelectContentProps{MaxHeight: contentMaxHeight}, func() {
					b.List(c, SelectListProps{Scroll: &scroll}, func() { b.Group(c, "", func() { b.Collection(c, nil) }) })
					b.ScrollUpButton(c)
					b.ScrollDownButton(c)
				})
			})
		})
	}, 600, 600)
	r.Frame()
	r.Frame()
	initial := runtime.alignHeight
	first := selectSourceOption(t, r, "Item 00")
	if math.Abs(float64(first.Bounds.Y+first.Bounds.H/2-runtime.valueBox.Y-runtime.valueBox.H/2)) > 1.1 {
		t.Fatalf("first item did not align: item%+v value%+v", first.Bounds, runtime.valueBox)
	}
	r.Key(0, ui.KeyEnd)
	r.Frame()
	r.Frame()
	last := selectSourceOption(t, r, "Item 99")
	if last.Offscreen || scroll.Y <= 0 || runtime.alignHeight <= initial+20 || runtime.alignHeight > 580.5 {
		t.Fatalf("native reveal/expansion: last%+v scroll%v initial%v final%v", last.Bounds, scroll.Y, initial, runtime.alignHeight)
	}
	if p := runtime.panelBox; p.Y < 9.5 || p.Y+p.H > 590.5 || last.Bounds.Y < p.Y || last.Bounds.Y+last.Bounds.H > p.Y+p.H || scroll.Y < scroll.MaxY && last.Bounds.Y+last.Bounds.H > p.Y+p.H-24 {
		t.Fatalf("End left the aligned last row outside fitted panel/arrow gutters: item%+v panel%+v scroll%+v", last.Bounds, p, scroll)
	}
	// An active item may leave the viewport during ordinary user scrolling.
	// Extra frames must not turn the one-shot reveal into a scroll lock.
	before := scroll.Y
	p := runtime.panelBox
	r.Scroll(p.X+p.W/2, p.Y+p.H/2, 0, -120)
	r.Frame()
	scrolled := scroll.Y
	r.Frame()
	if scrolled >= before || scroll.Y != scrolled {
		t.Fatalf("active reveal absorbed manual scrolling: before%v user%v final%v", before, scrolled, scroll.Y)
	}
	r.Key(0, ui.KeyEnd)
	r.Frame()
	r.Frame()
	// A live resize must replace the old measured cap and reveal only once.
	// The fixed trigger extends beyond this narrow window, so the popup must
	// also clamp its independent horizontal bounds before a real row click.
	r.SetSize(300, 360)
	r.Frame()
	r.Frame()
	last = selectSourceOption(t, r, "Item 99")
	if p := runtime.panelBox; p.X < 4.5 || p.X+p.W > 295.5 || p.Y < 9.5 || p.Y+p.H > 350.5 || last.Offscreen || last.Bounds.Y < p.Y || last.Bounds.Y+last.Bounds.H > p.Y+p.H {
		t.Fatalf("live narrow/short resize retained an old popup pose: item%+v panel%+v scroll%+v", last.Bounds, p, scroll)
	}
	shrunkHeight := runtime.panelBox.H
	p = runtime.panelBox
	before = scroll.Y
	r.Scroll(p.X+p.W/2, p.Y+p.H/2, 0, -90)
	r.Frame()
	scrolled = scroll.Y
	r.Frame()
	if scrolled >= before || scroll.Y != scrolled {
		t.Fatalf("resize reveal absorbed later user scrolling: before%v user%v final%v", before, scrolled, scroll.Y)
	}
	r.Key(0, ui.KeyEnd)
	r.Frame()
	r.Frame()
	last = selectSourceOption(t, r, "Item 99")
	r.ClickAt(last.Bounds.X+last.Bounds.W/2, last.Bounds.Y+last.Bounds.H/2)
	r.Frame()
	if state.Value.Value != "99" || state.Open {
		t.Fatalf("aligned long-list pointer choice %+v", state)
	}
	selectSourceOpen(t, r, "Long aligned")
	r.SetSize(600, 600)
	r.Frame()
	r.Frame()
	if p := runtime.panelBox; p.H <= shrunkHeight+100 || p.X < 4.5 || p.X+p.W > 595.5 || p.Y < 9.5 || p.Y+p.H > 590.5 {
		t.Fatalf("restoring window size did not recover a bounded expanded popup: %+v", p)
	}
	contentMaxHeight = 220
	r.Frame()
	r.Frame()
	last = selectSourceOption(t, r, "Item 99")
	if p := runtime.panelBox; p.H > 220.5 || last.Offscreen || last.Bounds.Y < p.Y || last.Bounds.Y+last.Bounds.H > p.Y+p.H {
		t.Fatalf("current Content.MaxHeight did not replace the measured cap: item%+v panel%+v", last.Bounds, p)
	}
	contentMaxHeight = 0
	r.Frame()
	r.Frame()
	if p := runtime.panelBox; p.H <= 320 || p.Y < 9.5 || p.Y+p.H > 590.5 {
		t.Fatalf("removing the current content cap did not recover bounded height: %+v", p)
	}
	r.Key(0, ui.KeyHome)
	r.Frame()
	r.Frame()
	first = selectSourceOption(t, r, "Item 00")
	if p := runtime.panelBox; p.Y < 9.5 || p.Y+p.H > 590.5 || first.Bounds.Y < p.Y || scroll.Y > 0 && first.Bounds.Y < p.Y+24 {
		t.Fatalf("Home left the first row under an arrow or outside the window: item%+v panel%+v scroll%+v", first.Bounds, p, scroll)
	}
	r.ClickAt(first.Bounds.X+first.Bounds.W/2, first.Bounds.Y+first.Bounds.H/2)
	r.Frame()
	if state.Value.Value != "0" || state.Open || !r.Focused("Long aligned") {
		t.Fatalf("aligned first-row pointer/focus return %+v", state)
	}
}

func TestSelectSourceOrdinaryMotionRetainsInertExitAndReverses(t *testing.T) {
	falseValue := false
	state := SelectState{Value: SelectSelection{"go", true}, Open: true}
	var runtime *selectRuntime
	r := tester(func(c *ui.Context) {
		SelectRoot(c, &state, SelectRootProps{Options: []Option{{Value: "go", Label: "Go"}}}, func(b *SelectScope) {
			runtime = b.f.s
			b.Trigger(c, SelectTriggerProps{Label: "Motion"}, nil)
			b.Content(c, SelectContentProps{AlignItemWithTrigger: &falseValue}, func() {
				b.List(c, SelectListProps{}, func() {
					b.Group(c, "", func() {
						b.Item(c, "go", SelectItemProps{}, func() { ui.Box(c).Size(16, 16).Background(ui.RGB(255, 0, 0)) })
					})
				})
			})
		})
	})
	r.Advance(120 * time.Millisecond)
	r.Frame()
	state.Open = false
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleOption, "Go"); ok {
		t.Fatal("closing option remained accessible")
	}
	if n, _ := r.AccessByRole(ui.RoleComboBox, "Motion"); !n.Focused {
		t.Fatal("programmatic close retained focus in inert options")
	}
	r.Advance(50 * time.Millisecond)
	r.Frame()
	pose := runtime.motion.sample(r.Now())
	if pose.opacity <= 0 || pose.opacity >= 1 {
		t.Fatalf("exit was not retained/interpolated %+v", pose)
	}
	state.Open = true
	r.Frame()
	if !runtime.open || runtime.motion.from.opacity <= 0 || runtime.motion.from.opacity >= 1 {
		t.Fatalf("reopen did not reverse current pose %+v", runtime.motion.from)
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	r.Frame()
	if runtime.motion.sample(r.Now()) != comboPoseShown {
		t.Fatal("reduced motion did not finish active entry")
	}
	state.Open = false
	r.Frame()
	if runtime.motion.sample(r.Now()) != runtime.motion.to {
		t.Fatal("reduced close retained an active animation")
	}
}
