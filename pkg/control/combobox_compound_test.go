package control

import (
	"fmt"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func comboCompoundStrings(c *ui.Context, state *ComboboxState[string], popup bool, onSubmit func()) {
	ComboboxRoot(c, state, ComboboxRootProps[string]{Items: []string{"Go", "Rust", "Swift"}, Adapter: ComboboxStringAdapter()}, func(b *ComboboxScope[string]) {
		if popup {
			b.Trigger(c, ComboboxTriggerProps{Label: "Compound"}, func() { b.Value(c, ComboboxValueProps{Placeholder: "Choose"}) })
		} else {
			parts := b.Input(c, ComboboxInputProps{Label: "Compound", Placeholder: "Search"})
			parts.Root.Width(240)
			if onSubmit != nil {
				parts.Editor.OnSubmit(onSubmit)
			}
		}
		b.Content(c, ComboboxContentProps{}, func() {
			if popup {
				b.Input(c, ComboboxInputProps{Label: "Compound", Placeholder: "Search", HideTrigger: true})
			}
			b.Empty(c, "No results")
			b.List(c, ComboboxListProps{}, func() {
				b.Collection(c, func(c *ui.Context, item string, _ int) {
					b.Item(c, item, ComboboxItemProps{}, func() { ui.Text(c, item) })
				})
			})
		})
	})
}

func TestComboboxCompoundIndependentSlotsUseNativeEditorAndSelection(t *testing.T) {
	for _, popup := range []bool{false, true} {
		t.Run(map[bool]string{false: "input", true: "trigger-dialog"}[popup], func(t *testing.T) {
			state := ComboboxState[string]{}
			submits := 0
			r := tester(func(c *ui.Context) { comboCompoundStrings(c, &state, popup, func() { submits++ }) })
			comboBehaviorOpen(t, r, "Compound")
			r.Type("Ru")
			r.Key(0, ui.KeyDown)
			r.Key(0, ui.KeyEnter)
			r.Advance(120 * time.Millisecond)
			r.Frame()
			if state.Value != (ComboboxSelection[string]{"Rust", true}) || state.Open {
				t.Fatalf("compound native choice %+v", state)
			}
			if !popup && state.Query != "Rust" {
				t.Fatalf("external editor label=%q", state.Query)
			}
			if popup && state.Query != "" {
				t.Fatalf("completed popup query=%q", state.Query)
			}
			if !popup {
				comboBehaviorOpen(t, r, "Compound")
				r.Key(ui.Cmd, ui.KeyA)
				r.Type("missing")
				r.Key(0, ui.KeyEnter)
				r.Frame()
				if submits != 1 {
					t.Fatalf("public native Editor.OnSubmit calls=%d", submits)
				}
			}
		})
	}
}

func TestComboboxCompoundMountUnloadsAndUsesFreshCallbacks(t *testing.T) {
	state := ComboboxState[string]{}
	visible := true
	version := 1
	calls := []int{}
	other := ""
	var old, current *comboRuntime
	var expired *ComboboxScope[string]
	r := tester(func(c *ui.Context) {
		if visible {
			v := version
			ComboboxRoot(c, &state, ComboboxRootProps[string]{Items: []string{"Go"}, Adapter: ComboboxStringAdapter(), OnQueryChange: func(string, *ComboboxChange) { calls = append(calls, v) }}, func(b *ComboboxScope[string]) {
				current = b.f.s
				expired = b
				b.Input(c, ComboboxInputProps{Label: "Compound"})
				b.Content(c, ComboboxContentProps{}, func() {
					b.Empty(c, "Empty")
					b.List(c, ComboboxListProps{}, func() {
						b.Collection(c, func(c *ui.Context, v string, _ int) { b.Item(c, v, ComboboxItemProps{}, nil) })
					})
				})
			})
		}
		Input(c, &other, InputProps{Label: "Other"})
	})
	comboBehaviorOpen(t, r, "Compound")
	r.Type("G")
	old = current
	visible = false
	r.Frame()
	if !old.closed {
		t.Fatal("headless mount retained its internal session after unmount")
	}
	if err := r.Click("Other"); err != nil {
		t.Fatal(err)
	}
	r.Type("next")
	if !slices.Equal(calls, []int{1}) {
		t.Fatalf("unmounted query callback survived: %v", calls)
	}
	visible = true
	version = 2
	state.Open = false
	r.Frame()
	if current == old || current.closed {
		t.Fatal("remount reused the disposed session")
	}
	comboBehaviorOpen(t, r, "Compound")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Go")
	if !slices.Equal(calls, []int{1, 2}) {
		t.Fatalf("remount invoked stale props: %v", calls)
	}
	func() {
		defer func() {
			if message := recover(); message != "control.ComboboxScope: slots must be built during ComboboxRoot" {
				t.Fatalf("expired public scope was not rejected: %v", message)
			}
		}()
		expired.FormValues()
	}()
}

// These tests exercise the independent public slots through actual native
// layout, input, AX and scroll integration, rather than duplicating helpers.
func TestComboboxCompoundContentPlacementAndAnchorSizing(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		for _, side := range []ComboboxSide{ComboboxSideBottom, ComboboxSideTop, ComboboxSideLeft, ComboboxSideRight, ComboboxSideInlineStart, ComboboxSideInlineEnd} {
			state := ComboboxState[string]{Open: true}
			var anchor, panel ui.Rect
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Absolute().Left(300).Top(250).Width(180).Children(func() {
					ComboboxRoot(c, &state, ComboboxRootProps[string]{Items: []string{"Go", "Rust"}, Adapter: ComboboxStringAdapter()}, func(b *ComboboxScope[string]) {
						parts := b.Input(c, ComboboxInputProps{Label: "Geometry"})
						anchor = parts.Editor.LayoutBounds()
						p := b.Content(c, ComboboxContentProps{Side: side}, func() {
							b.Empty(c, "Empty")
							b.List(c, ComboboxListProps{}, func() {
								b.Collection(c, func(c *ui.Context, v string, _ int) { b.Item(c, v, ComboboxItemProps{}, nil) })
							})
						})
						panel = p.LayoutBounds()
					})
				})
			}, 800, 600)
			r.SetDirection(direction)
			r.Advance(120 * time.Millisecond)
			r.Frame()
			physical := comboPhysicalSide(side, direction)
			gap := float32(0)
			switch physical {
			case ComboboxSideTop:
				gap = anchor.Y - panel.Y - panel.H
			case ComboboxSideBottom:
				gap = panel.Y - anchor.Y - anchor.H
			case ComboboxSideLeft:
				gap = anchor.X - panel.X - panel.W
			case ComboboxSideRight:
				gap = panel.X - anchor.X - anchor.W
			}
			if comboCompoundAbs(gap-6) > 1 || comboCompoundAbs(panel.W-anchor.W-28) > 1 {
				t.Fatalf("%v/%s anchor=%+v panel=%+v gap=%g", direction, side, anchor, panel, gap)
			}
			if panel.X < 5 || panel.Y < 5 || panel.X+panel.W > 795 || panel.Y+panel.H > 595 {
				t.Fatalf("panel escaped available window: %+v", panel)
			}
		}
	}
	// Explicit anchors take the source data-chips branch (no added 28DIP),
	// even if the anchor is an arbitrary caller-owned element.
	state := ComboboxState[string]{Open: true}
	var handle ui.Handle
	var anchor, panel ui.Rect
	zero := float32(0)
	r := tester(func(c *ui.Context) {
		ComboboxRoot(c, &state, ComboboxRootProps[string]{Items: []string{"Go"}, Adapter: ComboboxStringAdapter()}, func(b *ComboboxScope[string]) {
			b.Input(c, ComboboxInputProps{Label: "Explicit"})
			a := ui.Box(c).Size(120, 20).Bind(&handle)
			anchor = a.LayoutBounds()
			p := b.Content(c, ComboboxContentProps{Anchor: &handle, SideOffset: &zero}, func() { b.List(c, ComboboxListProps{}, func() { b.Item(c, "Go", ComboboxItemProps{}, nil) }) })
			panel = p.LayoutBounds()
		})
	})
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if comboCompoundAbs(panel.W-anchor.W) > 1 || comboCompoundAbs(panel.Y-anchor.Y-anchor.H) > 1 {
		t.Fatalf("explicit zero-offset anchor=%+v panel=%+v", anchor, panel)
	}
}

func TestComboboxCompoundTypedControlledChangesAndCallbacks(t *testing.T) {
	type choice struct {
		ID, Version int
		Name        string
	}
	items := []choice{{1, 1, "Alpha"}, {2, 1, "Beta"}}
	state := ComboboxState[choice]{Value: ComboboxSelection[choice]{choice{1, 1, "Old Alpha"}, true}}
	adapter := ComboboxAdapter[choice]{Key: func(v choice) string { return strconv.Itoa(v.ID) }, Label: func(v choice) string { return v.Name }, FormValue: func(v choice) string { return "id=" + strconv.Itoa(v.ID) }, Equal: func(a, b choice) bool { return a.ID == b.ID && a.Version == b.Version }}
	rejectQuery, rejectValue, rejectOpen := false, false, false
	highlights, complete := []string{}, []bool{}
	var submitted []string
	r := tester(func(c *ui.Context) {
		ComboboxRoot(c, &state, ComboboxRootProps[choice]{Items: items, Adapter: adapter, QueryControlled: true, AutoHighlight: ComboboxHighlightAlways,
			OnQueryChange: func(q string, e *ComboboxChange) {
				if rejectQuery {
					e.Cancel()
				}
			}, OnValueChange: func(_ ComboboxSelection[choice], e *ComboboxChange) {
				if rejectValue {
					e.Cancel()
				}
			}, OnOpenChange: func(_ bool, e *ComboboxChange) {
				if rejectOpen {
					e.Cancel()
				}
			},
			OnHighlighted: func(v ComboboxSelection[choice], _ *ComboboxChange) {
				if v.Valid {
					highlights = append(highlights, v.Value.Name)
				}
			}, OnOpenChangeComplete: func(open bool) { complete = append(complete, open) },
		}, func(b *ComboboxScope[choice]) {
			p := b.Input(c, ComboboxInputProps{Label: "Typed"})
			formValues := b.FormValues()
			p.Editor.OnSubmit(func() { submitted = formValues })
			b.Content(c, ComboboxContentProps{}, func() {
				b.Empty(c, "Empty")
				b.List(c, ComboboxListProps{}, func() {
					b.Collection(c, func(c *ui.Context, v choice, _ int) { b.Item(c, v, ComboboxItemProps{}, nil) })
				})
			})
		})
	})
	if state.Query != "" {
		t.Fatalf("controlled empty query overwritten: %q", state.Query)
	}
	rejectOpen = true
	comboBehaviorOpen(t, r, "Typed")
	if state.Open {
		t.Fatal("canceled open applied")
	}
	rejectOpen = false
	comboBehaviorOpen(t, r, "Typed")
	if !state.Open {
		t.Fatal("open was not controlled")
	}
	if comboBehaviorActive(r) != "Alpha" {
		t.Fatalf("always highlight=%q", comboBehaviorActive(r))
	}
	rejectQuery = true
	r.Type("B")
	r.Frame()
	if state.Query != "" {
		t.Fatalf("canceled query=%q", state.Query)
	}
	rejectQuery = false
	r.Type("B")
	r.Frame()
	if state.Query != "B" || comboBehaviorActive(r) != "Beta" {
		t.Fatalf("query/always active=%q/%q", state.Query, comboBehaviorActive(r))
	}
	rejectValue = true
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if state.Value.Value.Name != "Old Alpha" {
		t.Fatalf("canceled value=%+v", state.Value)
	}
	// Programmatic changes use the newest object, while Equal independently
	// determines item selection even when stable widget keys are reused.
	state.Value = ComboboxSelection[choice]{choice{1, 2, "Replacement Alpha"}, true}
	state.Query = ""
	state.Open = true
	r.Frame()
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if option, ok := r.AccessByRole(ui.RoleOption, "Alpha"); !ok || option.Selected {
		t.Fatalf("Key incorrectly overrode explicit Equal: %+v", option)
	}
	state.Query = "missing"
	r.Frame()
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(submitted, []string{"id=1"}) {
		t.Fatalf("real Editor submission=%v", submitted)
	}
	if !slices.Contains(highlights, "Beta") || !slices.Contains(complete, true) || !slices.Contains(complete, false) {
		t.Fatalf("current-frame lifecycle callbacks highlight=%v complete=%v", highlights, complete)
	}
	// Multiple widget keys can intentionally share selection equality. The
	// first source item wins deterministically, independently of map iteration.
	adapter.Equal = func(a, b choice) bool { return true }
	stable := ComboboxState[choice]{Value: ComboboxSelection[choice]{choice{9, 1, "Equal choice"}, true}, Open: true}
	first := tester(func(c *ui.Context) {
		ComboboxRoot(c, &stable, ComboboxRootProps[choice]{Items: items, Adapter: adapter}, func(b *ComboboxScope[choice]) {
			b.Input(c, ComboboxInputProps{Label: "Stable equality"})
			b.Content(c, ComboboxContentProps{}, func() {
				b.List(c, ComboboxListProps{}, func() {
					b.Collection(c, func(c *ui.Context, v choice, _ int) { b.Item(c, v, ComboboxItemProps{}, nil) })
				})
			})
		})
	})
	first.Advance(120 * time.Millisecond)
	for range 3 {
		first.Frame()
		alpha, _ := first.AccessByRole(ui.RoleOption, "Alpha")
		beta, _ := first.AccessByRole(ui.RoleOption, "Beta")
		if !alpha.Selected || beta.Selected {
			t.Fatalf("Equal selection did not follow source order: alpha%v beta%v", alpha.Selected, beta.Selected)
		}
	}
}

func TestComboboxCompoundGridNavigationAndNativeAccessibility(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		state := ComboboxState[string]{}
		items := []string{"A", "B", "C", "D", "E"}
		r := tester(func(c *ui.Context) {
			ComboboxRoot(c, &state, ComboboxRootProps[string]{Items: items, Adapter: ComboboxStringAdapter(), Grid: true}, func(b *ComboboxScope[string]) {
				b.Input(c, ComboboxInputProps{Label: "Grid"})
				b.Content(c, ComboboxContentProps{}, func() {
					b.List(c, ComboboxListProps{}, func() {
						b.Row(c, 0, func() { b.Item(c, "A", ComboboxItemProps{}, nil); b.Item(c, "B", ComboboxItemProps{}, nil) })
						b.Row(c, 1, func() {
							b.Item(c, "C", ComboboxItemProps{}, nil)
							b.Item(c, "D", ComboboxItemProps{Disabled: true}, nil)
							b.Item(c, "E", ComboboxItemProps{}, nil)
						})
					})
				})
			})
		})
		r.SetDirection(direction)
		comboBehaviorOpen(t, r, "Grid")
		r.Key(0, ui.KeyDown)
		forward := ui.KeyRight
		if direction == ui.RightToLeft {
			forward = ui.KeyLeft
		}
		r.Key(0, forward)
		if n, ok := r.AccessByRole(ui.RoleCell, "B"); !ok || !n.Focused || n.RowIndex != 1 || n.ColumnIndex != 2 {
			t.Fatalf("grid horizontal cell=%+v", n)
		}
		r.Key(0, ui.KeyDown)
		if n, ok := r.AccessByRole(ui.RoleCell, "C"); !ok || !n.Focused {
			t.Fatalf("ragged vertical failed to skip disabled: %+v", n)
		}
		r.Key(0, ui.KeyEnter)
		r.Advance(120 * time.Millisecond)
		r.Frame()
		if state.Value != (ComboboxSelection[string]{"C", true}) {
			t.Fatalf("grid chose %+v", state.Value)
		}
	}
}

func TestComboboxCompoundVirtualWindowRevealsNativeScrollAndFullPositions(t *testing.T) {
	items := make([]string, 100)
	for i := range items {
		items[i] = fmt.Sprintf("Item %03d", i)
	}
	state := ComboboxState[string]{Value: ComboboxSelection[string]{items[80], true}}
	scroll := ui.ScrollState{}
	rendered := 0
	revealed := -1
	r := tester(func(c *ui.Context) {
		ComboboxRoot(c, &state, ComboboxRootProps[string]{Items: items, Adapter: ComboboxStringAdapter(), Virtualized: true, Reveal: func(index int) { revealed = index; scroll.Y = float32(index * 32) }}, func(b *ComboboxScope[string]) {
			b.Input(c, ComboboxInputProps{Label: "Window"})
			b.Content(c, ComboboxContentProps{}, func() {
				b.Empty(c, "Empty")
				b.List(c, ComboboxListProps{Scroll: &scroll}, func() {
					filtered := b.FilteredItems()
					first := min(max(0, int(scroll.Y/32)-1), max(0, len(filtered)-1))
					last := min(len(filtered), first+12)
					rendered = last - first
					ui.Box(c).Height(float32(first * 32)).FillWidth()
					for _, v := range filtered[first:last] {
						b.Item(c, v, ComboboxItemProps{}, nil)
					}
					ui.Box(c).Height(float32((len(filtered) - last) * 32)).FillWidth()
				})
			})
		})
	})
	comboBehaviorOpen(t, r, "Window")
	r.Frame()
	n, ok := r.AccessByRole(ui.RoleOption, items[80])
	if !ok || !n.Focused || n.Offscreen || n.PosInSet != 81 || n.SetSize != 100 || rendered >= 100 || revealed != 80 || scroll.Y <= 0 {
		t.Fatalf("native window reveal=%d built=%d scroll=%g item=%+v", revealed, rendered, scroll.Y, n)
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if state.Value.Value != items[81] {
		t.Fatalf("offscreen indexed selection=%+v", state.Value)
	}
}

func comboCompoundAbs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func TestComboboxCompoundManualItemsGroupsAndEmptyListLayers(t *testing.T) {
	state := ComboboxState[string]{Open: true}
	showBeta := true
	r := tester(func(c *ui.Context) {
		ComboboxRoot(c, &state, ComboboxRootProps[string]{Adapter: ComboboxStringAdapter()}, func(b *ComboboxScope[string]) {
			b.Input(c, ComboboxInputProps{Label: "Manual"})
			b.Content(c, ComboboxContentProps{}, func() {
				b.Empty(c, "Nothing here")
				b.List(c, ComboboxListProps{}, func() {
					b.Item(c, "Alpha", ComboboxItemProps{}, nil)
					if showBeta {
						b.Item(c, "Beta", ComboboxItemProps{}, nil)
					}
				})
			})
		})
	})
	// Inspect the first submitted native AX tree. No Advance, warm-up frame or
	// second application build may be needed to finalize manual registration.
	if _, exists := r.AccessByRole(ui.RoleStatus, "Nothing here"); exists {
		t.Fatal("first AX submission exposed Empty beside real manual items")
	}
	for index, label := range []string{"Alpha", "Beta"} {
		item, exists := r.AccessByRole(ui.RoleOption, label)
		if !exists || item.PosInSet != index+1 || item.SetSize != 2 {
			t.Fatalf("first AX submission %s position=%d size=%d", label, item.PosInSet, item.SetSize)
		}
	}
	r.Key(0, ui.KeyEscape)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	state.Open = false
	r.Frame()
	comboBehaviorOpen(t, r, "Manual")
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyDown)
	if comboBehaviorActive(r) != "Beta" {
		t.Fatalf("manual registration navigation=%q", comboBehaviorActive(r))
	}
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	r.Frame()
	if state.Value != (ComboboxSelection[string]{"Beta", true}) {
		t.Fatalf("manual Item selection=%+v", state.Value)
	}
	showBeta = false
	comboBehaviorOpen(t, r, "Manual")
	r.Frame()
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("missing")
	r.Frame()
	status, ok := r.AccessByRole(ui.RoleStatus, "Nothing here")
	if !ok {
		t.Fatal("Empty sibling status missing")
	}
	list, ok := r.AccessByRole(ui.RoleListBox, "Manual options")
	if !ok || list.Bounds.H > 1 || status.Bounds.H <= 0 {
		t.Fatalf("Empty/List did not have separate visible/empty layers: empty=%+v list=%+v", status.Bounds, list.Bounds)
	}
	if _, ok := r.AccessByRole(ui.RoleOption, "Beta"); ok {
		t.Fatal("removed manual Item remained rendered")
	}

	groupState := ComboboxState[string]{}
	groups := []ComboboxGroupData[string]{{Key: "", Label: "Letters", Items: []string{"Alpha", "Beta"}}, {Key: "numbers", Label: "Numbers", Items: []string{"One", "Two"}}}
	g := tester(func(c *ui.Context) {
		ComboboxRoot(c, &groupState, ComboboxRootProps[string]{Groups: groups, Adapter: ComboboxStringAdapter()}, func(b *ComboboxScope[string]) {
			b.Input(c, ComboboxInputProps{Label: "Groups"})
			b.Content(c, ComboboxContentProps{}, func() {
				b.Empty(c, "No groups")
				b.List(c, ComboboxListProps{}, func() {
					b.Groups(c, func(c *ui.Context, data ComboboxGroupData[string], index int) {
						if index > 0 {
							b.Separator(c)
						}
						b.Group(c, data, func() {
							b.Label(c, data.Label)
							b.Collection(c, func(c *ui.Context, value string, _ int) { b.Item(c, value, ComboboxItemProps{}, nil) })
						})
					})
				})
			})
		})
	})
	comboBehaviorOpen(t, g, "Groups")
	g.Type("One")
	g.Frame()
	if _, ok := g.AccessByRole(ui.RoleGroup, "Letters"); ok {
		t.Fatal("empty group retained")
	}
	if _, ok := g.AccessByRole(ui.RoleGroup, "Numbers"); !ok {
		t.Fatal("matching group missing")
	}
	if g.HasText("Alpha") || g.HasText("Beta") {
		t.Fatal("empty-key Group leaked another collection")
	}
	for _, n := range g.AccessAll() {
		if n.Role == ui.RoleSeparator {
			t.Fatal("orphan separator after group filtering")
		}
	}
}
