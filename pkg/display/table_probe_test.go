package display

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
	"math"
	"testing"
	"time"
)

func TestTableStaticRTLSpansShareActualWidthsAndNativeIndices(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{Label: "Spans", Columns: []TableColumn{{Width: 100}, {Width: 110}, {Width: 120}}},
			TableHeader(TableRow(TableRowProps{Key: "h"}, TableHead("First"), TableHead("Second"), TableHead("Third"))),
			TableBody(TableRow(TableRowProps{Key: "r"}, TableCell("Wide", TableCellProps{ColSpan: 2}), TableCell("Last"))),
		)
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.SetDirection(ui.RightToLeft)
	first, _ := r.AccessByRole(ui.RoleColumnHeader, "First")
	second, _ := r.AccessByRole(ui.RoleColumnHeader, "Second")
	wide, _ := r.AccessByRole(ui.RoleCell, "Wide")
	if math.Abs(float64(wide.Bounds.W-first.Bounds.W-second.Bounds.W)) > .1 || wide.ColumnSpan != 2 || wide.ColumnIndex != 1 || wide.RowIndex != 2 || math.Abs(float64(wide.Bounds.X-second.Bounds.X)) > .1 {
		t.Fatalf("RTL span/index mismatch: first=%v second=%v wide=%+v", first.Bounds, second.Bounds, wide)
	}
}

func TestTableStaticContentShrinkAndIMEKeepRealIdentity(t *testing.T) {
	value, wide := "", true
	var scroll ui.ScrollState
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		text := "short"
		if wide {
			text = "A long unbroken table column that creates horizontal overflow"
		}
		TableWith(c, TableOptions{Scroll: &scroll}, TableBody(TableRow(TableRowProps{Key: "r"}, TableCell(text, TableCellProps{Key: "text"}), TableCell("", TableCellProps{Key: "editor", Content: func(c *ui.Context) { ui.TextInput(c, &value).Label("IME editor").Width(120) }}))))
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	scroll.X = scroll.MaxX
	r.Frame()
	if err := r.Click("IME editor"); err != nil {
		t.Fatal(err)
	}
	r.Compose("文", 1)
	wide = false
	r.Frame()
	r.Compose("文档", 2)
	r.Type("文档")
	if value != "文档" || !r.Focused("IME editor") || scroll.MaxX != 0 {
		t.Fatalf("width shrink lost actual editor/IME: value=%q focus=%v max=%v", value, r.Focused("IME editor"), scroll.MaxX)
	}
	if r.FramePending() {
		t.Fatal("reduced static table did not settle")
	}
}

func TestTableStaticNativeCancelDoesNotInvokeAction(t *testing.T) {
	clicked, builds, slots := 0, 0, 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		builds++
		theme.Apply(c)
		TableWith(c, TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("", TableCellProps{Content: func(c *ui.Context) {
			slots++
			if ui.Button(c, "Real action").Clicked() {
				clicked++
			}
		}}))))
	}, 400, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	button, _ := r.Find("Real action")
	r.Press(button.X+10, button.Y+10)
	r.CancelPress()
	r.Advance(time.Second)
	if clicked != 0 || builds != slots || r.FramePending() {
		t.Fatalf("cancel/callback/idle failed: clicks=%d builds=%d slots=%d frame=%v", clicked, builds, slots, r.FramePending())
	}
}

func TestTableStaticReducedMotionStopsActiveHoverTransition(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("Hover"))))
	}, 400, 180)
	cell, _ := r.AccessByRole(ui.RoleCell, "Hover")
	r.Move(cell.Bounds.X+cell.Bounds.W-8, cell.Bounds.Y+cell.Bounds.H/2)
	r.Advance(50 * time.Millisecond)
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	if r.FramePending() {
		t.Fatal("active row hover kept requesting frames after reduced-motion preference")
	}
}

func TestTableStaticLateSecondTableDoesNotPingPongCorrectionFrames(t *testing.T) {
	second := false
	builds := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		builds++
		theme.Apply(c)
		ui.Column(c).Children(func() {
			TableWith(c.Key("first"), TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("First table"))))
			if second {
				TableWith(c.Key("second"), TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("Second table"))))
			}
			services := c.Services()
			ui.Box(c).Size(1, 1).Draw(func(_ *ui.Painter, _ ui.Rect) {
				if !second {
					second = true
					services.Invalidate()
				}
			})
		})
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !second || r.FramePending() {
		t.Fatalf("independent tables did not settle: second=%v builds=%d pending=%v", second, builds, r.FramePending())
	}
}

func TestTableStaticSharedWaveSettlesAcrossOffscreenMountReorderAndWidths(t *testing.T) {
	for _, reduced := range []bool{false, true} {
		t.Run(fmt.Sprint(reduced), func(t *testing.T) {
			ids := []string{"a"}
			widths := map[string]float32{"a": 130, "b": 160, "c": 220}
			values := map[string]*string{"a": new(string), "b": new(string), "c": new(string)}
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Width(400).Height(1000).Shrink(0).Children(func() {
					ui.Box(c).Height(260).Shrink(0)
					for _, id := range ids {
						columns, cells := make([]TableColumn, 10), make([]TableCellProps, 10)
						for i := range columns {
							columns[i].Width = 100
							cells[i] = TableCell("short")
						}
						cells[9] = TableCell("", TableCellProps{Key: "editor", Content: func(c *ui.Context) { ui.TextInput(c, values[id]).Width(widths[id]).Label("Editor " + id) }})
						TableWith(c.Key(id), TableOptions{Label: "Table " + id, Columns: columns}, TableBody(TableRow(TableRowProps{Key: id}, cells...))).Height(50).Shrink(0)
					}
				})
			}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
			check := func(stage string) {
				t.Helper()
				if r.FramePending() {
					t.Fatalf("%s left static window busy", stage)
				}
				current := ""
				found := map[string]bool{}
				for _, node := range r.AccessAll() {
					if node.Role == ui.RoleTable {
						current = node.Name
					}
					if node.Role == ui.RoleCell && node.ColumnIndex == 10 {
						for _, id := range ids {
							if current == "Table "+id {
								want := max(float32(100), widths[id]+16)
								if math.Abs(float64(node.Bounds.W-want)) > .1 {
									t.Fatalf("%s ignored offscreen%s content: got=%v want=%v", stage, id, node.Bounds, want)
								}
								found[id] = true
							}
						}
					}
				}
				if len(found) != len(ids) {
					t.Fatalf("%s lost native rows: found=%v ids=%v", stage, found, ids)
				}
			}
			check("initial")
			for i, next := range [][]string{{"a", "b"}, {"c", "a", "b"}, {"b", "c", "a"}, {"c", "a"}, {"a"}, {}, {"b", "c"}} {
				ids = next
				r.Frame()
				check(fmt.Sprintf("mount/reorder%d", i))
			}
			widths["b"], widths["c"] = 330, 500
			r.Frame()
			check("grow same keys")
			widths["b"], widths["c"] = 90, 120
			r.Frame()
			check("shrink same keys")
		})
	}
}

func TestTableStaticLateFirstTableDoesNotResetOtherWaveParticipants(t *testing.T) {
	ids := []string{"old"}
	changed := false
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Children(func() {
			for _, id := range ids {
				TableWith(c.Key(id), TableOptions{Label: id}, TableBody(TableRow(TableRowProps{}, TableCell(id))))
			}
			services := c.Services()
			ui.Box(c).Size(1, 1).Draw(func(_ *ui.Painter, _ ui.Rect) {
				if !changed {
					changed, ids = true, []string{"new", "old"}
					services.Invalidate()
				}
			})
		})
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.HasText("new") || r.FramePending() {
		t.Fatal("prepending a new table before previous wave participant prevented settling")
	}
}

func TestTableStaticTransformedTableMeasurementsUseLayoutCoordinates(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		TableWith(c, TableOptions{}, TableBody(TableRow(TableRowProps{}, TableCell("Scaled native content")))).Scale(.75, .75)
	}, 400, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if r.FramePending() {
		t.Fatal("transformed Bounds fought untransformed Draw widths and never settled")
	}
}
