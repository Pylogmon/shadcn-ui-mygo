package display

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func displayTester(view func(*ui.Context)) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(20).Gap(12).Children(func() { view(c) })
	}, 600, 600)
}

func displayClick(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	tt.Frame()
}

func clickAccessible(t *testing.T, tt *ui.Tester, role ui.Role, label string) {
	t.Helper()
	info, ok := tt.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q: %+v", role, label, tt.AccessAll())
	}
	tt.ClickAt(info.Bounds.X+info.Bounds.W/2, info.Bounds.Y+info.Bounds.H/2)
	tt.Advance(time.Second)
	tt.Frame()
}

func TestAccordionSingleAndKeyboard(t *testing.T) {
	first, second := false, false
	tt := displayTester(func(c *ui.Context) {
		Accordion(c, AccordionProps{Single: true, Items: []AccordionItem{
			{Key: "one", Title: "First", Open: &first, Content: func() { ui.Text(c, "First panel") }},
			{Key: "two", Title: "Second", Open: &second, Content: func() { ui.Text(c, "Second panel") }},
		}})
	})
	displayClick(t, tt, "First")
	if !first || second || !tt.HasText("First panel") {
		t.Fatalf("first open state = %v,%v texts=%v", first, second, tt.Texts())
	}
	displayClick(t, tt, "Second")
	if first || !second || tt.HasText("First panel") || !tt.HasText("Second panel") {
		t.Fatalf("second open state = %v,%v texts=%v", first, second, tt.Texts())
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeySpace)
	tt.Advance(time.Second)
	tt.Frame()
	if !first || second {
		t.Fatalf("keyboard open state = %v,%v", first, second)
	}
}

func TestAccordionSingleNormalizesInitialState(t *testing.T) {
	first, second := true, true
	displayTester(func(c *ui.Context) {
		Accordion(c, AccordionProps{Single: true, Items: []AccordionItem{
			{Key: "one", Title: "First", Open: &first}, {Key: "two", Title: "Second", Open: &second},
		}})
	})
	if !first || second {
		t.Fatalf("single accordion kept conflicting initial state %v,%v", first, second)
	}
}

func TestCalendarChangesDateAndPreservesClock(t *testing.T) {
	loc := time.FixedZone("local", 8*3600)
	date := time.Date(2026, time.November, 9, 14, 35, 10, 123, loc)
	changed := false
	tt := displayTester(func(c *ui.Context) {
		if Calendar(c, &date).Changed() {
			changed = true
		}
	})
	clickAccessible(t, tt, ui.RoleButton, "November 12, 2026")
	if date.Day() != 12 || date.Hour() != 14 || date.Minute() != 35 || date.Nanosecond() != 123 || date.Location() != loc || !changed {
		t.Fatalf("calendar selection = %v changed=%v", date, changed)
	}
}

func TestDatePickerBrowsesThenCommitsAndCloses(t *testing.T) {
	date := time.Date(2026, time.November, 9, 0, 0, 0, 0, time.UTC)
	changed := false
	tt := displayTester(func(c *ui.Context) {
		if DatePicker(c, &date, DatePickerProps{Label: "Release date"}).Changed() {
			changed = true
		}
	})
	displayClick(t, tt, "Nov 9, 2026")
	anchor, _ := tt.AccessByRole(ui.RoleButton, "Release date")
	month, ok := tt.Find("November 2026")
	if !ok {
		t.Fatalf("datepicker did not show November: %v", tt.Texts())
	}
	if month.Y < anchor.Bounds.Y+anchor.Bounds.H {
		t.Fatal("calendar overlaps its trigger")
	}
	clickAccessible(t, tt, ui.RoleButton, "Next month")
	if date.Month() != time.November || changed || !tt.HasText("December 2026") {
		t.Fatalf("browsing committed date=%v changed=%v texts=%v", date, changed, tt.Texts())
	}
	clickAccessible(t, tt, ui.RoleButton, "December 12, 2026")
	if date.Month() != time.December || date.Day() != 12 || !changed {
		t.Fatalf("picked date=%v changed=%v", date, changed)
	}
	if tt.HasText("December 2026") {
		t.Fatal("calendar popup stayed open after selection")
	}
}

func TestDatePickerEmptyStartsAtCurrentMonth(t *testing.T) {
	date := time.Time{}
	var now time.Time
	tt := displayTester(func(c *ui.Context) { now = c.Now(); DatePicker(c, &date, DatePickerProps{}) })
	displayClick(t, tt, "Pick a date")
	if !tt.HasText(now.Format("January 2006")) {
		t.Fatalf("empty picker texts %v, want month %s", tt.Texts(), now.Format("January 2006"))
	}
	if !date.IsZero() {
		t.Fatalf("opening empty picker committed %v", date)
	}
}

func TestCarouselClampsAndStopsAtBoundaries(t *testing.T) {
	selected, count := 9, 3
	tt := displayTester(func(c *ui.Context) {
		Carousel(c, &selected, count, func(i int) { ui.Text(c, fmt.Sprintf("Slide %d", i)) })
	})
	if selected != 2 || !tt.HasText("Slide 2") {
		t.Fatalf("initial carousel=%d texts=%v", selected, tt.Texts())
	}
	clickAccessible(t, tt, ui.RoleButton, "Next slide")
	if selected != 2 {
		t.Fatalf("last slide advanced to %d", selected)
	}
	clickAccessible(t, tt, ui.RoleButton, "Previous slide")
	if selected != 1 || !tt.HasText("Slide 1") {
		t.Fatalf("previous carousel=%d texts=%v", selected, tt.Texts())
	}
	count = 0
	tt.Frame()
	if selected != 0 || tt.HasText("Slide 1") || !tt.HasText("0 / 0") {
		t.Fatalf("empty carousel=%d texts=%v", selected, tt.Texts())
	}
}

type testRecord struct {
	ID   int
	Name string
	Size int
}

func records() DataTableProps[testRecord] {
	return DataTableProps[testRecord]{Rows: []testRecord{
		{1, "Bee", 2}, {2, "Ant", 1}, {3, "Bee", 3}, {4, "Cat", 4},
	}, Columns: []DataColumn[testRecord]{
		{ID: "name", Title: "Name", Value: func(r testRecord) string { return r.Name }},
		{ID: "size", Title: "Size", Value: func(r testRecord) string { return fmt.Sprint(r.Size) }, Compare: func(a, b testRecord) int {
			if a.Size < b.Size {
				return math.MinInt
			}
			if a.Size > b.Size {
				return math.MaxInt
			}
			return 0
		}},
	}, Key: func(r testRecord) any { return r.ID }}
}

func TestDataRowsFilterSortPagingAndInputImmutability(t *testing.T) {
	p := records()
	state := DataTableState{Sort: ui.SortOrder{Column: "name"}, PageSize: 2}
	rows, total := DataRows(&state, p)
	if total != 4 || rows[0].ID != 2 || rows[1].ID != 1 {
		t.Fatalf("ascending rows=%v total=%d", rows, total)
	}
	state.Page = 1
	rows, total = DataRows(&state, p)
	if rows[0].ID != 3 || rows[1].ID != 4 {
		t.Fatalf("stable second page=%v", rows)
	}
	state.Query = " BEE "
	rows, total = DataRows(&state, p)
	if total != 2 || state.Page != 0 || rows[0].ID != 1 || rows[1].ID != 3 {
		t.Fatalf("filtered page=%v total=%d page=%d", rows, total, state.Page)
	}
	state.Query = ""
	state.Sort = ui.SortOrder{Column: "size", Descending: true}
	state.PageSize = 10
	rows, _ = DataRows(&state, p)
	if !slices.EqualFunc(rows, []testRecord{{4, "Cat", 4}, {3, "Bee", 3}, {1, "Bee", 2}, {2, "Ant", 1}}, func(a, b testRecord) bool { return a == b }) {
		t.Fatalf("descending arbitrary comparator = %v", rows)
	}
	if !slices.EqualFunc(p.Rows, records().Rows, func(a, b testRecord) bool { return a == b }) {
		t.Fatal("DataRows mutated caller's records")
	}
	state.Query = "no match"
	state.Page = math.MaxInt
	rows, total = DataRows(&state, p)
	if total != 0 || len(rows) != 0 || state.Page != 0 {
		t.Fatalf("empty page=%v total=%d page=%d", rows, total, state.Page)
	}
}

func TestChartMissingValuesBreakLines(t *testing.T) {
	segments := finiteSegments([]float64{1, 2, math.NaN(), 3, math.Inf(-1), 4, 5},
		chart.NewBand("A", "B", "C", "D", "E", "F", "G"), chart.NewLinear(0, 5))
	if len(segments) != 3 || len(segments[0].X) != 2 || len(segments[1].X) != 1 || len(segments[2].X) != 2 {
		t.Fatalf("non-finite observations did not break lines: %+v", segments)
	}
}

func TestDataTableNativeSortingPagingAndKeyRemoval(t *testing.T) {
	p := records()
	state := DataTableState{PageSize: 2}
	tt := displayTester(func(c *ui.Context) { DataTable(c, &state, p) })
	if !tt.HasText("Bee") || !tt.HasText("Ant") || tt.HasText("Cat") {
		t.Fatalf("first page texts=%v", tt.Texts())
	}
	displayClick(t, tt, "Name")
	if state.Sort.Column != "name" {
		t.Fatalf("native header did not set sort: %+v", state.Sort)
	}
	displayClick(t, tt, "Next")
	if state.Page != 1 || !tt.HasText("Cat") || tt.HasText("Ant") {
		t.Fatalf("second page=%d texts=%v", state.Page, tt.Texts())
	}
	p.Key = nil
	tt.Frame()
	if state.List.Key != nil {
		t.Fatal("DataTable kept a stale row key function after it was removed")
	}
	state.Query = "no match"
	tt.Frame()
	if state.Page != 0 || !tt.HasText("No results.") {
		t.Fatalf("filtered table page=%d texts=%v", state.Page, tt.Texts())
	}
}

func TestChartIgnoresNonFiniteData(t *testing.T) {
	for _, kind := range []ChartKind{BarChart, LineChart, AreaChart} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			tt := displayTester(func(c *ui.Context) {
				Chart(c, ChartProps{Kind: kind, Label: "Measurements", Labels: []string{"A", "B", "C", "D"}, Series: []ChartSeries{
					{Name: "Series", Values: []float64{-3, math.NaN(), math.Inf(1), 7}},
				}})
			})
			for _, info := range tt.AccessAll() {
				for _, v := range []float32{info.Bounds.X, info.Bounds.Y, info.Bounds.W, info.Bounds.H} {
					if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
						t.Fatalf("chart produced non-finite accessibility bounds: %+v", info)
					}
				}
			}
		})
	}
}
