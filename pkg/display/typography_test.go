package display

import (
	"image/color"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestTypographyUsesPinnedNativeLineMetricsAndH2Divider(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(400).Padding(20).Gap(8).Children(func() {
			Typography(c, "Heading", H2).Label("Heading two")
			Typography(c, "A paragraph", Paragraph).Label("Paragraph")
			Typography(c, "Small", Small).Label("Small text")
		})
	}, 500, 300)
	for _, sample := range []struct {
		name   string
		height float32
	}{{"Heading two", 45}, {"Paragraph", 28}} {
		bounds, ok := r.Find(sample.name)
		if !ok || bounds.H != sample.height {
			t.Fatalf("pinned line metrics for %q: %+v want height=%g", sample.name, bounds, sample.height)
		}
	}
}

func typographyTester(view func(*ui.Context), width, height int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) { theme.Apply(c); ui.Column(c).Fill().Padding(20).Children(func() { view(c) }) }, width, height)
}

func TestTypographyProseSourceSpacingHeadingLevelsAndNoScopeLeak(t *testing.T) {
	r := typographyTester(func(c *ui.Context) {
		TypographyProse(c, func(child *ui.Context) {
			Typography(child, "First", H1)
			Typography(child, "Opening paragraph", Paragraph)
			Typography(child, "Second", H2)
			Typography(child, "Another paragraph", Paragraph)
			Typography(child, "Third", H3)
			Typography(child, "Fourth", H4)
		}).Label("Prose")
		Typography(c, "Outside prose", Paragraph)
	}, 600, 650)
	for i, name := range []string{"First", "Second", "Third", "Fourth"} {
		h, ok := r.AccessByRole(ui.RoleHeading, name)
		if !ok || h.Level != i+1 {
			t.Fatalf("native heading %q=%+v", name, h)
		}
	}
	for _, gap := range []struct {
		before, after string
		value         float32
	}{{"First", "Opening paragraph", 24}, {"Opening paragraph", "Second", 40}, {"Second", "Another paragraph", 24}, {"Another paragraph", "Third", 32}, {"Third", "Fourth", 32}} {
		a, _ := r.Find(gap.before)
		b, _ := r.Find(gap.after)
		if b.Y-a.Y-a.H != gap.value {
			t.Fatalf("source flow %s→%s=%g want%g", gap.before, gap.after, b.Y-a.Y-a.H, gap.value)
		}
	}
	prose, _ := r.Find("Prose")
	outside, _ := r.Find("Outside prose")
	if outside.Y != prose.Y+prose.H {
		t.Fatal("prose spacing leaked to sibling typography")
	}
}

func TestTypographyH2DividerAndQuoteUseSemanticPixelsInBothDirections(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			var tokens *theme.Tokens
			r := ui.NewTester(func(c *ui.Context) {
				controller := theme.Controller{Mode: theme.LightMode}
				if dark {
					controller.Mode = theme.DarkMode
				}
				controller.Apply(c)
				tokens = theme.TokensOf(c)
				root := ui.Column(c).Fill().Padding(20).Gap(12).Background(tokens.Background)
				if rtl {
					root.Direction(ui.RightToLeft)
				}
				root.Children(func() {
					Typography(c, "Section", H2).Label("Section frame")
					TypographyRich(c, Blockquote, func(child *ui.Context) { ui.Text(child, "Quoted words") }).Label("Quote frame")
				})
			}, 500, 250)
			h, _ := r.Find("Section frame")
			q, _ := r.Find("Quote frame")
			im := r.Image()
			want := typographyComposite(tokens.Background, tokens.Border)
			if got := im.RGBAAt(int(h.X+8), int(h.Y+h.H-1)); got != want {
				t.Fatalf("H2 actual divider=%v want=%v", got, want)
			}
			x := q.X
			if rtl {
				x = q.X + q.W - 1
			}
			if got := im.RGBAAt(int(x), int(q.Y+q.H/2)); got != want {
				t.Fatalf("quote logical border=%v want=%v", got, want)
			}
			left, right := q.X+q.W, q.X
			for py := int(q.Y); py < int(q.Y+q.H); py++ {
				for px := int(q.X + 2); px < int(q.X+q.W-2); px++ {
					p := im.RGBAAt(px, py)
					if p.R == tokens.Foreground.R && p.G == tokens.Foreground.G && p.B == tokens.Foreground.B {
						left = min(left, float32(px))
						right = max(right, float32(px+1))
					}
				}
			}
			inset := left - q.X
			if rtl {
				inset = q.X + q.W - right
			}
			if inset < 24 || inset > 32 {
				t.Fatalf("actual quote ink ignored 24+2 logical inset: rtl=%v inset=%g", rtl, inset)
			}
		}
	}
}

func typographyComposite(bg, fg ui.Color) color.RGBA {
	blend := func(b, f uint8) uint8 { return uint8((int(f)*int(fg.A) + int(b)*(255-int(fg.A)) + 127) / 255) }
	return color.RGBA{R: blend(bg.R, fg.R), G: blend(bg.G, fg.G), B: blend(bg.B, fg.B), A: 255}
}

func TestTypographyRichNativeLinksWrapRouteAndUseEnterOnly(t *testing.T) {
	router := ui.NewRouter("/docs/current")
	r := typographyTester(func(c *ui.Context) {
		router.View(c, func(_ *ui.Route) {
			TypographyRich(c, Paragraph, func(child *ui.Context) {
				ui.Text(child, "Read ")
				TypographyLink(child, "the native typography documentation", "../guide")
				ui.Text(child, " and install ")
				TypographyCode(child, "native/ui")
				ui.Text(child, " today.")
			})
		})
	}, 260, 240)
	if !r.HasText("Read the native typography documentation and install native/ui today.") {
		t.Fatalf("inline children did not join one paragraph: %v", r.Texts())
	}
	link, ok := r.AccessByRole(ui.RoleLink, "the native typography documentation")
	if !ok || link.Bounds.H < 28 {
		t.Fatal("native link did not wrap with the paragraph")
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused(link.Name) {
		t.Fatal("native inline link missing from Tab order")
	}
	r.Key(0, ui.KeySpace)
	if router.Location() != "/docs/current" {
		t.Fatal("Space activated a native Link")
	}
	r.Key(0, ui.KeyEnter)
	if router.Path() != "/guide" || len(r.OpenedURLs()) != 0 {
		t.Fatalf("native inline Router action=%s urls=%v", router.Location(), r.OpenedURLs())
	}
}

func TestTypographyInlineCodeKeepsNativePaddedBackgroundAndSelection(t *testing.T) {
	var tokens *theme.Tokens
	r := typographyTester(func(c *ui.Context) {
		tokens = theme.TokensOf(c)
		TypographyRich(c, Paragraph, func(child *ui.Context) {
			ui.Text(child, "Before ")
			TypographyCode(child, "sample").Label("Inline code")
			ui.Text(child, " after")
		}).Selectable().Label("Selectable paragraph")
	}, 600, 180)
	code, _ := r.Find("Inline code")
	before, _ := r.Find("Before ")
	if code.X <= before.X+before.W || code.W <= 40 {
		t.Fatalf("inline code had no actual layout padding: before=%+v code=%+v", before, code)
	}
	muted := 0
	im := r.Image()
	for y := int(code.Y); y < int(code.Y+code.H); y++ {
		for x := int(code.X); x < int(code.X+code.W); x++ {
			p := im.RGBAAt(x, y)
			if p.R == tokens.Muted.R && p.G == tokens.Muted.G && p.B == tokens.Muted.B {
				muted++
			}
		}
	}
	if muted < 100 {
		t.Fatal("native inline code background was lost by text merging")
	}
	if err := r.Click("Selectable paragraph"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(ui.Cmd, ui.KeyC)
	if r.Clipboard() != "Before sample after" {
		t.Fatalf("native rich selection/copy=%q", r.Clipboard())
	}
}

func TestTypographyListNativeSemanticsLogicalMarkersAndCollapsedSpacing(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		r := typographyTester(func(c *ui.Context) {
			root := ui.Column(c).FillWidth()
			if rtl {
				root.Direction(ui.RightToLeft)
			}
			root.Children(func() {
				TypographyProse(c, func(child *ui.Context) {
					Typography(child, "Introduction", Paragraph)
					TypographyList(child, []TypographyListItem{{Key: "a", Text: "First item"}, {Key: "b", Text: "Second item"}}, TypographyListProps{Label: "Instructions"})
					Typography(child, "Following paragraph", Paragraph)
				})
			})
		}, 500, 300)
		list, ok := r.AccessByRole(ui.RoleList, "Instructions")
		if !ok {
			t.Fatal("list has no native semantics")
		}
		intro, _ := r.Find("Introduction")
		first, _ := r.Find("First item")
		second, _ := r.Find("Second item")
		after, _ := r.Find("Following paragraph")
		if first.Y-intro.Y-intro.H != 32 || second.Y-first.Y-first.H != 8 || after.Y-second.Y-second.H != 24 {
			t.Fatalf("list spacing intro=%+v first=%+v second=%+v after=%+v", intro, first, second, after)
		}
		inset := first.X - list.Bounds.X
		if rtl {
			inset = list.Bounds.X + list.Bounds.W - first.X - first.W
		}
		if inset != 24 {
			t.Fatalf("list logical inset=%g", inset)
		}
		items := 0
		for _, node := range r.AccessAll() {
			if node.Role == ui.RoleListItem {
				items++
			}
			if node.Name == "•" {
				t.Fatal("decorative bullet duplicated AX")
			}
		}
		if items != 2 {
			t.Fatalf("native list item count=%d", items)
		}
	}
}

func TestTypographyKeyedListRetainsRealInteractiveChildren(t *testing.T) {
	items := []string{"a", "b"}
	draft := "draft"
	created := 0
	r := typographyTester(func(c *ui.Context) {
		var entries []TypographyListItem
		for _, key := range items {
			entries = append(entries, TypographyListItem{Key: key, Content: func(child *ui.Context) {
				if key == "a" {
					input := ui.TextInput(child, &draft).Label("List editor")
					ui.Local(input, "init", func() bool { created++; return true })
				} else {
					TypographyLink(child, "Guide", "https://example.com/guide")
				}
			}})
		}
		TypographyList(c, entries)
	}, 500, 240)
	if err := r.Click("List editor"); err != nil {
		t.Fatal(err)
	}
	r.Type("!")
	items = slices.Clone(items)
	slices.Reverse(items)
	r.Frame()
	if created != 1 || !r.Focused("List editor") || draft != "draft!" {
		t.Fatal("list reordering replaced the editor or moved focus")
	}
}

func TestTypographyTableNativeHeadersRowsZebraAndConstrainedScroll(t *testing.T) {
	var tokens *theme.Tokens
	r := typographyTester(func(c *ui.Context) {
		tokens = theme.TokensOf(c)
		TypographyTable(c, TypographyTableProps{Label: "Treasury", MinWidth: 560,
			Columns: []TypographyTableColumn{{Key: "treasury", Title: "Treasury"}, {Key: "people", Title: "People", Align: ui.End}},
			Rows:    []TypographyTableRow{{Key: "one", Cells: []string{"Empty", "Overflowing"}}, {Key: "two", Cells: []string{"Modest", "Satisfied"}}, {Key: "three", Cells: []string{"Full", "Ecstatic"}}}})
	}, 400, 320)
	table, ok := r.AccessByRole(ui.RoleTable, "Treasury")
	if !ok || table.RowCount != 4 || table.ColumnCount != 2 {
		t.Fatalf("table native model=%+v", table)
	}
	head, ok := r.AccessByRole(ui.RoleColumnHeader, "Treasury")
	if !ok {
		t.Fatal("native column header missing")
	}
	if head.Bounds.W != 279 {
		t.Fatalf("column collapsed in a narrow native viewport: %+v", head.Bounds)
	}
	modest, found := r.AccessByRole(ui.RoleCell, "Modest")
	if !found {
		t.Fatalf("cell not named Modest: %v", r.AccessAll())
	}
	pixel := r.Image().RGBAAt(int(modest.Bounds.X+4), int(modest.Bounds.Y+4))
	if pixel.R != tokens.Muted.R || pixel.G != tokens.Muted.G || pixel.B != tokens.Muted.B {
		t.Fatalf("even source row has no themed zebra: %v", pixel)
	}
	viewport := ui.Rect{}
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleScroll {
			viewport = node.Bounds
			break
		}
	}
	if viewport.W != 360 || table.Bounds.W < 560 {
		t.Fatalf("table did not remain in a bounded scroll viewport: viewport=%+v table=%+v", viewport, table.Bounds)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("static native table scheduled idle frames")
	}
}

func TestTypographyTableStableRowAndColumnKeysAndEmptyCallbacks(t *testing.T) {
	rows := []TypographyTableRow{{Key: "alpha"}, {Key: "beta"}}
	columns := []TypographyTableColumn{{Key: "one", Title: "One"}, {Key: "two", Title: "Two"}}
	draft := "saved"
	created := 0
	show := false
	r := typographyTester(func(c *ui.Context) {
		TypographyRich(c, Paragraph, func(child *ui.Context) {
			if show {
				ui.Text(child, "Now present")
			}
		})
		TypographyTable(c, TypographyTableProps{Columns: columns, Rows: rows, CellContent: func(child *ui.Context, row, col int) {
			if rows[row].Key == "alpha" && columns[col].Key == "one" {
				e := ui.TextInput(child, &draft).Label("Table editor")
				ui.Local(e, "init", func() bool { created++; return true })
			}
		}})
	}, 600, 400)
	if err := r.Click("Table editor"); err != nil {
		t.Fatal(err)
	}
	r.Type("!")
	slices.Reverse(rows)
	slices.Reverse(columns)
	show = true
	r.Frame()
	if created != 1 || !r.Focused("Table editor") || draft != "saved!" {
		t.Fatal("empty callbacks or table reordering lost native editor identity")
	}
}

func TestTypographyOrderedListLargeStartStaysFinite(t *testing.T) {
	r := typographyTester(func(c *ui.Context) {
		TypographyList(c, []TypographyListItem{{Text: "One"}, {Text: "Two"}}, TypographyListProps{Ordered: true, Start: int(^uint(0)>>1) - 1})
	}, 500, 200)
	if !r.HasText(strconv.Itoa(int(^uint(0)>>1)) + ".") {
		t.Fatal("ordered marker overflowed")
	}
	marker, _ := r.Find(strconv.Itoa(int(^uint(0)>>1)) + ".")
	text, _ := r.Find("Two")
	if marker.W < 100 || marker.X+marker.W > text.X-7 {
		t.Fatal("long ordered marker was truncated or overlapped list text")
	}
	r = typographyTester(func(c *ui.Context) {
		TypographyList(c, []TypographyListItem{{Text: "One"}, {Text: "Two"}}, TypographyListProps{Label: "Invalid range", Ordered: true, Start: int(^uint(0) >> 1)})
	}, 500, 200)
	list, _ := r.AccessByRole(ui.RoleList, "Invalid range")
	if !list.Invalid || !r.HasText("1.") || !r.HasText("2.") {
		t.Fatal("overflowed numbering range did not use an explicit invalid finite fallback")
	}
}
