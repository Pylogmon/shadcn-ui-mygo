package display

import (
	"fmt"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestTableStaticStyleHooksConfigureLiveNativeRowsAndCells(t *testing.T) {
	builds, rows, cells := 0, 0, 0
	background := ui.RGB(20, 55, 100)
	foreground := ui.RGB(230, 40, 30)
	r := ui.NewTesterWith(func(c *ui.Context) {
		builds++
		theme.Apply(c)
		TableWith(c, TableOptions{Label: "Styled"}, TableBody(TableRow(TableRowProps{Key: "value", Style: func(row ui.Element) {
			rows++
			if !row.Valid() {
				t.Fatal("row style received an expired frame element")
			}
			row.Height(56).Background(background)
		}}, TableCell("Value", TableCellProps{Style: func(cell ui.Element) {
			cells++
			cell.Label("Styled value").Description("Native cell description").FontSize(18).TextColor(foreground)
		}}))))
	}, 300, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	cell := tableAccess(t, r, ui.RoleCell, "Styled value")
	if cell.Bounds.H != 56 || cell.Description != "Native cell description" || rows != builds || cells != builds {
		t.Fatalf("live styling did not reach native geometry/metadata or ran outside its build: cell=%+v callbacks=%d/%d builds=%d", cell, rows, cells, builds)
	}
	if pixel := r.Image().RGBAAt(2, 25); pixel != (color.RGBA{20, 55, 100, 255}) {
		t.Fatalf("row style did not override the actual source background: %v", pixel)
	}
	colored := 0
	for y := 0; y < 56; y++ {
		for x := 0; x < 100; x++ {
			pixel := r.Image().RGBAAt(x, y)
			if pixel.R > 100 && int(pixel.R) > int(pixel.B)*2 {
				colored++
			}
		}
	}
	if colored < 10 {
		t.Fatal("cell style did not change its actual native glyph colors")
	}
}

func TestTableCompatibleCaptionUsesSourceMargin(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Table(c, TableProps{Columns: []ui.TableColumn{{Title: "Invoice"}}, Rows: [][]string{{"INV001"}}, Caption: "Recent invoices"})
	}, 480, 300)
	caption, ok := r.Find("Recent invoices")
	if !ok {
		t.Fatal("caption missing")
	}
	var bounds ui.Rect
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleTable {
			bounds = node.Bounds
			break
		}
	}
	if gap := caption.Y - (bounds.Y + bounds.H); math.Abs(float64(gap-16)) > .1 {
		t.Fatalf("source caption margin is 16 DIP, got %g", gap)
	}
}

func tableAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing native role %v %q: %+v", role, name, r.AccessAll())
	}
	return a
}

func TestTableStaticCompoundSectionsShareColumnsAndFooterSpan(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{Label: "Invoices", Columns: []TableColumn{{Key: "invoice", Width: 100}, {Key: "status"}, {Key: "amount", Align: ui.End}}},
			TableCaption("Recent invoices"),
			TableHeader(TableRow(TableRowProps{Key: "head"}, TableHead("Invoice"), TableHead("Status"), TableHead("Amount"))),
			TableBody(TableRow(TableRowProps{Key: "one"}, TableRowHead("INV001"), TableCell("Paid"), TableCell("$250.00"))),
			TableFooter(TableRow(TableRowProps{Key: "total"}, TableCell("Total", TableCellProps{ColSpan: 2}), TableCell("$250.00"))),
		)
	}, 480, 300)
	table := tableAccess(t, r, ui.RoleTable, "Invoices")
	if table.RowCount != 3 || table.ColumnCount != 3 {
		t.Fatalf("native table dimensions=%+v", table)
	}
	head := tableAccess(t, r, ui.RoleColumnHeader, "Invoice")
	rowHead := tableAccess(t, r, ui.RoleRowHeader, "INV001")
	if math.Abs(float64(head.Bounds.W-100)) > .1 || head.Bounds.X != rowHead.Bounds.X || head.Bounds.W != rowHead.Bounds.W {
		t.Fatalf("column width did not carry across sections: head=%v body=%v", head.Bounds, rowHead.Bounds)
	}
	if head.Bounds.H != 40 {
		t.Fatalf("source header height=%g, want40", head.Bounds.H)
	}
	total := tableAccess(t, r, ui.RoleCell, "Total")
	if total.ColumnSpan != 2 || total.ColumnIndex != 1 || total.RowIndex != 3 || total.Bounds.W <= head.Bounds.W {
		t.Fatalf("native footer span/index was lost: span=%d col=%d row=%d width=%g first=%g", total.ColumnSpan, total.ColumnIndex, total.RowIndex, total.Bounds.W, head.Bounds.W)
	}
	caption, _ := r.Find("Recent invoices")
	if gap := caption.Y - table.Bounds.Y - table.Bounds.H; math.Abs(float64(gap-16)) > .1 {
		t.Fatalf("source caption margin=%g, want16", gap)
	}
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleRow && a.RowIndex == 1 && a.Bounds.H != 40 {
			t.Fatalf("source header row height=%g, want40", a.Bounds.H)
		}
	}
}

func TestTableStaticSelectedExpandedAndLastBorderPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			tokens := theme.Light()
			if dark {
				tokens = theme.Dark()
			}
			expanded, collapsed := true, false
			r := ui.NewTester(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
					TableWith(c, TableOptions{Label: "States"}, TableBody(
						TableRow(TableRowProps{Key: "selected", Selected: true}, TableCell("Selected")),
						TableRow(TableRowProps{Key: "expanded", Expanded: &expanded}, TableCell("Expanded")),
						TableRow(TableRowProps{Key: "collapsed", Expanded: &collapsed}, TableCell("Collapsed")),
					))
				})
			}, 400, 260)
			selected := tableAccess(t, r, ui.RoleCell, "Selected")
			expand := tableAccess(t, r, ui.RoleCell, "Expanded")
			collapse := tableAccess(t, r, ui.RoleCell, "Collapsed")
			pixel := func(a ui.AccessInfo) color.RGBA {
				return r.Image().RGBAAt(int(a.Bounds.X+a.Bounds.W-4), int(a.Bounds.Y+a.Bounds.H/2))
			}
			want := color.RGBA{tokens.Muted.R, tokens.Muted.G, tokens.Muted.B, 255}
			if got := pixel(selected); got != want {
				t.Fatalf("selected pixel=%v want=%v", got, want)
			}
			base := color.RGBA{tokens.Background.R, tokens.Background.G, tokens.Background.B, 255}
			if got := pixel(collapse); got != base {
				t.Fatalf("collapsed aria-expanded=false applied source expanded background: %v want=%v", got, base)
			}
			if got := pixel(expand); got == base || got == want {
				t.Fatalf("expanded background did not paint source muted/50: %v", got)
			}
			bottom := r.Image().RGBAAt(int(collapse.Bounds.X+collapse.Bounds.W-4), int(collapse.Bounds.Y+collapse.Bounds.H-1))
			if bottom != base {
				t.Fatalf("last body row still has a bottom border: %v want=%v", bottom, base)
			}
		})
	}
}

func TestTableStaticCustomNativeControlsKeepIdentityAcrossRowInsertion(t *testing.T) {
	rows := []string{"one", "two"}
	values := map[string]*string{"one": new(string), "two": new(string), "before": new(string)}
	clicked := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		body := make([]TableRowProps, 0, len(rows))
		for _, id := range rows {
			body = append(body, TableRow(TableRowProps{Key: id}, TableCell(id), TableCell("", TableCellProps{Key: "editor", Content: func(child *ui.Context) {
				control.Input(child, values[id], control.InputProps{Label: "Edit " + id}).Width(120)
			}}), TableCell("", TableCellProps{Content: func(child *ui.Context) {
				if control.Button(child, "Save "+id, control.ButtonProps{Size: control.Small}).Clicked() {
					clicked++
				}
			}})))
		}
		TableWith(c, TableOptions{Label: "Editable"}, TableBody(body...))
	}, 600, 320)
	if err := r.Click("Edit two"); err != nil {
		t.Fatal(err)
	}
	r.Type("draft")
	r.Key(ui.Cmd, ui.KeyA)
	rows = append([]string{"before"}, rows...)
	r.Frame()
	if !r.Focused("Edit two") || *values["two"] != "draft" {
		t.Fatal("keyed row insertion replaced the live native editor")
	}
	r.Type("replacement")
	if *values["two"] != "replacement" {
		t.Fatalf("native selection was lost on insertion: %q", *values["two"])
	}
	if err := r.Click("Save two"); err != nil {
		t.Fatal(err)
	}
	if clicked != 1 {
		t.Fatal("native cell action did not execute once")
	}
}

func TestTableStaticEmptyCustomSlotDoesNotMoveSiblingEditor(t *testing.T) {
	show := false
	value := ""
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{}, TableBody(TableRow(TableRowProps{Key: "one"},
			TableCell("", TableCellProps{Key: "optional", Content: func(child *ui.Context) {
				if show {
					ui.Text(child, "Now visible")
				}
			}}),
			TableCell("", TableCellProps{Key: "editor", Content: func(child *ui.Context) {
				control.Input(child, &value, control.InputProps{Label: "Stable editor"}).Width(140)
			}}),
		)))
	}, 500, 240)
	if err := r.Click("Stable editor"); err != nil {
		t.Fatal(err)
	}
	r.Type("before")
	show = true
	r.Frame()
	r.Type(" after")
	if value != "before after" || !r.Focused("Stable editor") {
		t.Fatalf("empty custom slot moved native sibling identity: %q", value)
	}
}

func TestTableVirtualPathBuildsOnlyVisibleLazyRows(t *testing.T) {
	var state ui.ListState
	state.Key = func(row int) any { return row }
	built := map[int]bool{}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Table(c, TableProps{Label: "Large", Columns: []ui.TableColumn{{Title: "Record", ID: "record"}}, State: &state, Height: 200, RowCount: 100000, CellContent: func(child *ui.Context, row, col int) { built[row] = true; ui.Text(child, fmt.Sprintf("Row %d", row)) }})
	}, 500, 320)
	if len(built) > 40 || !r.HasText("Row 0") || built[99999] {
		t.Fatalf("lazy path eagerly built data: count=%d last=%v", len(built), built[99999])
	}
	state.ScrollToEnd()
	r.Frame()
	r.Advance(time.Second)
	if !r.HasText("Row 99999") || len(built) > 80 {
		t.Fatalf("native virtual scroll did not reach100k end cheaply: count=%d", len(built))
	}
}

func TestTableStaticCheckboxSpacingUsesPhysicalRightInLTRAndRTL(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			checked := false
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				TableWith(c, TableOptions{Label: "Selection", Columns: []TableColumn{{Width: 24}, {}}},
					TableBody(TableRow(TableRowProps{Key: "item"}, TableCell("", TableCellProps{Checkbox: true, Content: func(child *ui.Context) {
						control.Checkbox(child, &checked, "", control.CheckboxProps{}).Label("Select item")
					}}), TableCell("Item"))),
				)
			}, 300, 180)
			r.SetDirection(direction)
			check := tableAccess(t, r, ui.RoleCheckBox, "Select item")
			var cell ui.AccessInfo
			for _, a := range r.AccessAll() {
				if a.Role == ui.RoleCell && a.ColumnIndex == 1 {
					cell = a
					break
				}
			}
			if math.Abs(float64(cell.Bounds.X+cell.Bounds.W-check.Bounds.X-check.Bounds.W)) > .1 || cell.Bounds.W != 24 {
				t.Fatalf("checkbox kept physical right padding in direction%d: cell=%v checkbox=%v", direction, cell.Bounds, check.Bounds)
			}
			if err := r.Click("Select item"); err != nil {
				t.Fatal(err)
			}
			if !checked {
				t.Fatal("custom native checkbox did not update caller state")
			}
		})
	}
}

func TestTableStaticOverflowScrollsWithoutWrappingCells(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			var scroll ui.ScrollState
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				TableWith(c, TableOptions{Label: "Wide", Scroll: &scroll, MinWidth: 640, Columns: []TableColumn{{Width: 320}, {Width: 320}}}, TableBody(TableRow(TableRowProps{}, TableCell("Start"), TableCell("End"))))
			}, 260, 180)
			r.SetDirection(direction)
			if scroll.MaxX < 379 || scroll.MaxY != 0 {
				t.Fatalf("table did not create native horizontal overflow: %+v", scroll)
			}
			before := tableAccess(t, r, ui.RoleCell, "Start").Bounds
			scroll.X = scroll.MaxX
			r.Frame()
			after := tableAccess(t, r, ui.RoleCell, "Start").Bounds
			if math.Abs(float64(after.X-before.X)) < 379 || after.H != before.H {
				t.Fatalf("native horizontal scrolling did not move nowrap content: beforeX=%g afterX=%g H=%g->%g scrollX=%g maxX=%g", before.X, after.X, before.H, after.H, scroll.X, scroll.MaxX)
			}
		})
	}
}

func TestTableStaticColumnMaximumSurvivesSurplusSpace(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{Columns: []TableColumn{{Width: 60, MaxWidth: 80}, {MaxWidth: 100}}}, TableBody(TableRow(TableRowProps{}, TableCell("First"), TableCell("Second"))))
	}, 480, 200)
	first := tableAccess(t, r, ui.RoleCell, "First")
	second := tableAccess(t, r, ui.RoleCell, "Second")
	if first.Bounds.W > 80 || second.Bounds.W > 100 {
		t.Fatalf("surplus violated column maximums: first=%g second=%g", first.Bounds.W, second.Bounds.W)
	}
}

func TestTableStaticInvalidDimensionsPreserveNativeEditing(t *testing.T) {
	for _, value := range []float32{float32(math.NaN()), float32(math.Inf(1)), -1, math.MaxFloat32} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			draft := ""
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				TableWith(c, TableOptions{Label: "Invalid", MinWidth: value, Columns: []TableColumn{{Width: value, MinWidth: value, MaxWidth: value}}}, TableBody(TableRow(TableRowProps{}, TableCell("", TableCellProps{Content: func(child *ui.Context) { control.Input(child, &draft, control.InputProps{Label: "Draft"}).Width(120) }}))))
			}, 400, 220)
			table := tableAccess(t, r, ui.RoleTable, "Invalid")
			if !table.Invalid {
				t.Fatal("invalid geometry was not exposed on native table")
			}
			for _, a := range r.AccessAll() {
				for _, v := range []float32{a.Bounds.X, a.Bounds.Y, a.Bounds.W, a.Bounds.H} {
					if !tableFinite(v) {
						t.Fatalf("nonfinite native bounds: %v", a.Bounds)
					}
				}
			}
			if err := r.Click("Draft"); err != nil {
				t.Fatal(err)
			}
			r.Type("safe")
			if draft != "safe" {
				t.Fatal("invalid dimensions disabled actual editor")
			}
		})
	}
}

func TestTableStaticSelectPopupWorksOutsideItsRowAndKeepsFocus(t *testing.T) {
	assignee := "sarah"
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{Label: "Tasks"}, TableBody(TableRow(TableRowProps{Key: "task"}, TableCell("Design homepage"), TableCell("", TableCellProps{Content: func(child *ui.Context) {
			control.Select(child, &assignee, control.SelectProps{Label: "Assignee", Size: control.Small, Options: []control.Option{{Value: "sarah", Label: "Sarah Chen"}, {Value: "marcus", Label: "Marcus Rodriguez"}}}).Width(160)
		}}))))
	}, 480, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	trigger := tableAccess(t, r, ui.RolePopUpButton, "Assignee")
	if trigger.Bounds.W != 160 || trigger.Bounds.H != 32 {
		t.Fatalf("source select dimensions=%v", trigger.Bounds)
	}
	if err := r.Click("Assignee"); err != nil {
		t.Fatal(err)
	}
	option := tableAccess(t, r, ui.RoleOption, "Marcus Rodriguez")
	if option.Bounds.Y <= trigger.Bounds.Y+trigger.Bounds.H {
		t.Fatal("native listbox did not escape the table row")
	}
	if err := r.Click("Marcus Rodriguez"); err != nil {
		t.Fatal(err)
	}
	if assignee != "marcus" || !r.Focused("Assignee") {
		t.Fatal("native select did not update caller state and restore focus")
	}
}

func TestTableStaticFooterAndHoverUseSourceLayeredPixels(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			tokens := theme.Light()
			if dark {
				tokens = theme.Dark()
			}
			tokens.Background = ui.RGB(220, 240, 250)
			tokens.Muted = ui.RGB(30, 60, 90)
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
					TableWith(c, TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("Body"))), TableFooter(TableRow(TableRowProps{}, TableCell("Footer"))))
				})
			}, 400, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			footer := tableAccess(t, r, ui.RoleCell, "Footer")
			x, y := int(footer.Bounds.X+footer.Bounds.W-5), int(footer.Bounds.Y+footer.Bounds.H/2)
			colorOf := func(c ui.Color) color.RGBA { return color.RGBA{c.R, c.G, c.B, 255} }
			background := tokens.Muted.Alpha(.5).Over(tokens.Background)
			if got := r.Image().RGBAAt(x, y); got != colorOf(background) {
				t.Fatalf("footer background=%v want=%v", got, colorOf(background))
			}
			r.Move(float32(x), float32(y))
			want := tokens.Muted.Alpha(.5).Over(background)
			if got := r.Image().RGBAAt(x, y); got != colorOf(want) {
				t.Fatalf("footer hover did not layer row over section: got=%v want=%v", got, colorOf(want))
			}
		})
	}
}

func TestTableStaticHoverAnimatesAndReducedMotionSettles(t *testing.T) {
	render := func(reduced bool) *ui.Tester {
		return ui.NewTesterWith(func(c *ui.Context) {
			tokens := theme.Light()
			tokens.Muted = ui.RGB(20, 40, 60)
			theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() { TableWith(c, TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("Hover row")))) })
		}, 400, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
	}
	for _, reduced := range []bool{false, true} {
		t.Run(fmt.Sprint(reduced), func(t *testing.T) {
			r := render(reduced)
			cell := tableAccess(t, r, ui.RoleCell, "Hover row")
			x, y := int(cell.Bounds.X+cell.Bounds.W-5), int(cell.Bounds.Y+cell.Bounds.H/2)
			before := r.Image().RGBAAt(x, y)
			r.Move(float32(x), float32(y))
			r.Advance(75 * time.Millisecond)
			middle := r.Image().RGBAAt(x, y)
			r.Advance(100 * time.Millisecond)
			after := r.Image().RGBAAt(x, y)
			if before == after {
				t.Fatal("hover did not change actual native pixels")
			}
			if reduced {
				if middle != after || r.FramePending() {
					t.Fatal("reduced motion kept animating")
				}
			} else if middle == before || middle == after {
				t.Fatal("source color transition did not interpolate")
			}
		})
	}
}

func TestTableVirtualKeyPreservesNativeEditorAcrossRowInsertion(t *testing.T) {
	ids := []string{"one", "two"}
	drafts := map[string]*string{"one": new(string), "two": new(string), "before": new(string)}
	var state ui.ListState
	state.Key = func(row int) any { return ids[row] }
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Table(c, TableProps{Label: "Virtual editors", State: &state, Columns: []ui.TableColumn{{ID: "draft", Title: "Draft"}}, RowCount: len(ids), Height: 220, CellContent: func(child *ui.Context, row, col int) {
			id := ids[row]
			control.Input(child, drafts[id], control.InputProps{Label: "Draft " + id}).Width(160)
		}})
	}, 440, 280)
	if err := r.Click("Draft two"); err != nil {
		t.Fatal(err)
	}
	r.Type("selected")
	r.Key(ui.Cmd, ui.KeyA)
	ids = append([]string{"before"}, ids...)
	r.Frame()
	r.Type("kept")
	if *drafts["two"] != "kept" || !r.Focused("Draft two") {
		t.Fatal("native virtual row key did not retain editor selection/focus")
	}
}

func TestTableStaticEmptyCellsAndBodyKeepSourceGeometry(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprint(custom), func(t *testing.T) {
			cell := TableCell("")
			if custom {
				cell.Content = func(*ui.Context) {}
			}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				TableWith(c, TableOptions{Label: "Empty cell"}, TableBody(TableRow(TableRowProps{}, cell)))
			}, 400, 180)
			empty := tableAccess(t, r, ui.RoleCell, "")
			if empty.Bounds.H != 16 {
				t.Fatalf("empty source td should contain padding only: height=%g", empty.Bounds.H)
			}
		})
	}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{Label: "Empty body"}, TableHeader(TableRow(TableRowProps{}, TableHead("Name"))), TableBody(), TableCaption("No records yet"))
	}, 400, 180)
	table := tableAccess(t, r, ui.RoleTable, "Empty body")
	if table.RowCount != 1 || table.Bounds.H != 40 {
		t.Fatalf("empty body injected extra content: rows=%d height=%g", table.RowCount, table.Bounds.H)
	}
}

func TestTableStaticCustomContentRunsOncePerNativeBuild(t *testing.T) {
	builds, callbacks := 0, 0
	r := ui.NewTester(func(c *ui.Context) {
		builds++
		theme.Apply(c)
		TableWith(c, TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("", TableCellProps{Content: func(child *ui.Context) { callbacks++; control.Button(child, "Only once", control.ButtonProps{}) }}))))
	}, 400, 180)
	r.Frame()
	if builds != callbacks {
		t.Fatalf("custom native content was invoked for measurement: builds=%d callbacks=%d", builds, callbacks)
	}
}

func TestTableStaticRemovedContentReleasesIntrinsicColumnWidth(t *testing.T) {
	wide := true
	var scroll ui.ScrollState
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		rows := []TableRowProps{TableRow(TableRowProps{Key: "short"}, TableCell("Short"))}
		if wide {
			rows = append(rows, TableRow(TableRowProps{Key: "wide"}, TableCell("A very long unbroken native table label that must overflow the window")))
		}
		TableWith(c, TableOptions{Scroll: &scroll}, TableBody(rows...))
	}, 200, 180)
	if scroll.MaxX <= 100 {
		t.Fatal("unbroken source content was wrapped or squeezed")
	}
	wide = false
	r.Frame()
	if scroll.MaxX != 0 {
		t.Fatalf("unmounted content retained a stale intrinsic width: %g", scroll.MaxX)
	}
}

func TestTableStaticInvalidCallerScrollKeepsBoundsAndEditorFinite(t *testing.T) {
	scroll := ui.ScrollState{X: float32(math.NaN()), Y: float32(math.Inf(1))}
	draft := ""
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{Label: "Invalid scroll", Scroll: &scroll}, TableBody(TableRow(TableRowProps{}, TableCell("", TableCellProps{Content: func(child *ui.Context) {
			control.Input(child, &draft, control.InputProps{Label: "Scroll draft"}).Width(120)
		}}))))
	}, 400, 180)
	if !tableFinite(scroll.X) || !tableFinite(scroll.Y) {
		t.Fatal("caller scroll poisoned native offsets")
	}
	for _, a := range r.AccessAll() {
		if !tableFinite(a.Bounds.X) || !tableFinite(a.Bounds.Y) || !tableFinite(a.Bounds.W) || !tableFinite(a.Bounds.H) {
			t.Fatal("invalid scroll poisoned native bounds")
		}
	}
	if err := r.Click("Scroll draft"); err != nil {
		t.Fatal(err)
	}
	r.Type("kept")
	if draft != "kept" {
		t.Fatal("invalid scroll disabled native editing")
	}
}

func TestTableStaticFooterTopBorderRemainsAboveSelectedRow(t *testing.T) {
	tokens := theme.Light()
	tokens.Border = ui.RGB(210, 30, 90)
	tokens.Muted = ui.RGB(30, 60, 90)
	r := ui.NewTester(func(c *ui.Context) {
		theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
			TableWith(c, TableOptions{}, TableFooter(TableRow(TableRowProps{Selected: true}, TableCell("Selected footer"))))
		})
	}, 400, 180)
	cell := tableAccess(t, r, ui.RoleCell, "Selected footer")
	got := r.Image().RGBAAt(int(cell.Bounds.X+cell.Bounds.W-4), int(cell.Bounds.Y))
	want := color.RGBA{tokens.Border.R, tokens.Border.G, tokens.Border.B, 255}
	if got != want {
		t.Fatalf("selected footer covered source top border: got=%v want=%v", got, want)
	}
}

func TestTableStaticOffscreenCustomContentContributesToSharedWidth(t *testing.T) {
	columns := make([]TableColumn, 10)
	cells := make([]TableCellProps, 10)
	for i := range columns {
		columns[i].Width = 100
		cells[i] = TableCell("short")
	}
	value := "wide"
	cells[9] = TableCell("", TableCellProps{Key: "offscreen", Content: func(c *ui.Context) { ui.TextInput(c, &value).Width(300).Label("Offscreen editor") }})
	var scroll ui.ScrollState
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{Columns: columns, Scroll: &scroll}, TableBody(TableRow(TableRowProps{}, cells...)))
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	var last ui.AccessInfo
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleCell && node.ColumnIndex == 10 {
			last = node
		}
	}
	if last.Bounds.W < 316 || scroll.MaxX < 815 {
		t.Fatalf("offscreen native width was ignored: lastCell=%v scrollMax=%v", last.Bounds, scroll.MaxX)
	}
}

func TestTableStaticOffscreenSameKeyCustomWidthChangeReconciles(t *testing.T) {
	columns := make([]TableColumn, 10)
	for i := range columns {
		columns[i].Width = 100
	}
	width := float32(300)
	value := "wide"
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		cells := make([]TableCellProps, 10)
		for i := range cells {
			cells[i] = TableCell("short")
		}
		cells[9] = TableCell("", TableCellProps{Key: "offscreen", Content: func(c *ui.Context) { ui.TextInput(c, &value).Width(width).Label("Resizable offscreen editor") }})
		TableWith(c, TableOptions{Columns: columns}, TableBody(TableRow(TableRowProps{Key: "stable row"}, cells...)))
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	width = 500
	r.Frame()
	var last ui.AccessInfo
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleCell && node.ColumnIndex == 10 {
			last = node
		}
	}
	if last.Bounds.W < 516 {
		t.Fatalf("offscreen same-key native width change stayed stale: cell=%v", last.Bounds)
	}
}
