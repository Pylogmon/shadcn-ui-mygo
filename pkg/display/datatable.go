package display

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type DataColumn[T any] struct {
	ID, Title string
	Value     func(T) string
	Compare   func(T, T) int
	Width     float32
	// Cell composes arbitrary native content, including row actions/editors.
	// Value remains the sort/search/accessibility representation of that cell.
	Cell        func(*ui.Context, T)
	FilterValue func(T) string
	// Filter overrides exact-match OR filtering for this column.
	Filter                                      func(T, []string) bool
	MinWidth, MaxWidth                          float32
	Align                                       ui.Align
	Fixed, RowHeader, Unsortable, AlwaysVisible bool
}

// DataTableState owns sorting, column filters/visibility, selection and paging.
// Selection uses stable, comparable row keys. Supply Key when source rows may
// be inserted/replaced; otherwise selection follows their original indices.
type DataTableState struct {
	Query          string
	Sort           ui.SortOrder
	Page, PageSize int
	List           ui.ListState
	Hidden         map[string]bool
	Filters        map[string][]string
	Selection      ui.Selection[any]
	ownSelection   bool
}
type DataTableProps[T any] struct {
	Columns                                  []DataColumn[T]
	Rows                                     []T
	Key                                      func(T) any
	Search                                   func(T, string) bool
	Label                                    string
	Searchable, ShowColumnToggle, Selectable bool
	FilterPlaceholder                        string
	// SelectionActions receives selected records from the complete source, not
	// just the visible page. The caller decides what each action does.
	SelectionActions func([]T)
	Toolbar          func()
	Height           float32
}
type DataFacet struct {
	Value string
	Count int
}
type dataRecord[T any] struct {
	value T
	index int
}

func dataColumnID[T any](col DataColumn[T]) string {
	if col.ID != "" {
		return col.ID
	}
	return col.Title
}
func columnFilterValue[T any](col DataColumn[T], row T) string {
	if col.FilterValue != nil {
		return col.FilterValue(row)
	}
	if col.Value != nil {
		return col.Value(row)
	}
	return ""
}
func (s *DataTableState) SetFilter(column string, values []string) {
	if s.Filters == nil {
		s.Filters = map[string][]string{}
	}
	if len(values) == 0 {
		delete(s.Filters, column)
	} else {
		s.Filters[column] = slices.Clone(values)
	}
	s.Page = 0
}
func (s *DataTableState) SetColumnVisible(column string, visible bool) {
	if s.Hidden == nil {
		s.Hidden = map[string]bool{}
	}
	if visible {
		delete(s.Hidden, column)
	} else {
		s.Hidden[column] = true
	}
}
func dataFiltered[T any](state *DataTableState, p DataTableProps[T], except string) []dataRecord[T] {
	rows := make([]dataRecord[T], 0, len(p.Rows))
	query := strings.ToLower(strings.TrimSpace(state.Query))
	for index, row := range p.Rows {
		match := query == ""
		if !match && p.Search != nil {
			match = p.Search(row, query)
		} else if !match {
			for _, col := range p.Columns {
				if col.Value != nil && strings.Contains(strings.ToLower(col.Value(row)), query) {
					match = true
					break
				}
			}
		}
		if !match {
			continue
		}
		for _, col := range p.Columns {
			id := dataColumnID(col)
			values := state.Filters[id]
			if id == except || len(values) == 0 {
				continue
			}
			accepted := false
			if col.Filter != nil {
				accepted = col.Filter(row, values)
			} else {
				accepted = slices.Contains(values, columnFilterValue(col, row))
			}
			if !accepted {
				match = false
				break
			}
		}
		if match {
			rows = append(rows, dataRecord[T]{row, index})
		}
	}
	return rows
}
func dataRecords[T any](state *DataTableState, p DataTableProps[T]) ([]dataRecord[T], []dataRecord[T]) {
	rows := dataFiltered(state, p, "")
	for _, col := range p.Columns {
		if dataColumnID(col) == state.Sort.Column && !col.Unsortable && !col.RowHeader {
			slices.SortStableFunc(rows, func(a, b dataRecord[T]) int {
				if state.Sort.Descending {
					a, b = b, a
				}
				if col.Compare != nil {
					return col.Compare(a.value, b.value)
				}
				if col.Value != nil {
					return cmp.Compare(col.Value(a.value), col.Value(b.value))
				}
				return 0
			})
			break
		}
	}
	size := state.PageSize
	if size <= 0 {
		size = 10
	}
	pages := max(1, (len(rows)+size-1)/size)
	state.Page = max(0, min(state.Page, pages-1))
	start := state.Page * size
	return rows[start:min(start+size, len(rows))], rows
}

// DataRows returns the filtered, stable-sorted visible page without changing
// caller data. Filters combine with AND across columns, OR within a column.
func DataRows[T any](state *DataTableState, p DataTableProps[T]) ([]T, int) {
	page, all := dataRecords(state, p)
	rows := make([]T, len(page))
	for i, row := range page {
		rows[i] = row.value
	}
	return rows, len(all)
}

// DataFacetValues counts values after global/other-column filters, ignoring
// the column's own filter so the user can add alternatives without losing them.
func DataFacetValues[T any](state *DataTableState, p DataTableProps[T], column string) []DataFacet {
	var target *DataColumn[T]
	for i := range p.Columns {
		if dataColumnID(p.Columns[i]) == column {
			target = &p.Columns[i]
			break
		}
	}
	if target == nil {
		return nil
	}
	counts := map[string]int{}
	for _, row := range dataFiltered(state, p, column) {
		counts[columnFilterValue(*target, row.value)]++
	}
	values := make([]DataFacet, 0, len(counts))
	for value, count := range counts {
		values = append(values, DataFacet{value, count})
	}
	slices.SortFunc(values, func(a, b DataFacet) int { return cmp.Compare(a.Value, b.Value) })
	return values
}

func tableAction(c *ui.Context, label string, disabled bool) ui.Element {
	t := theme.TokensOf(c)
	e := ui.ButtonBase(c).Height(32).PaddingX(12).Radius(t.RadiusMD()).Border(1, t.Border).Disabled(disabled).TextColor(t.Foreground).Transition(motion.Transition(c))
	if e.Hovered() {
		e.Background(t.Accent)
	}
	return e.Children(func() { ui.Text(c, label).FontSize(13) })
}
func tableCheckbox(c *ui.Context, label string, on bool) ui.Element {
	t := theme.TokensOf(c)
	e := ui.ButtonBase(c).Role(ui.RoleCheckBox).Checked(on).Label(label).Size(16, 16).Padding(0).Radius(4).Border(1, t.Input).Center().FocusRingColor(t.Ring)
	if on {
		e.Background(t.Primary).Border(1, t.Primary).TextColor(t.PrimaryForeground)
	}
	return e.Children(func() {
		if on {
			ui.Text(c, "✓").FontSize(11)
		}
	})
}
func dataRecordKey[T any](p DataTableProps[T], row dataRecord[T]) any {
	if p.Key != nil {
		return p.Key(row.value)
	}
	return row.index
}

// DataTableFacetFilter is a composable multi-select facet popup. Its counts
// typically come from DataFacetValues, which honors the other active filters.
func DataTableFacetFilter(c *ui.Context, state *DataTableState, column, label string, options []DataFacet) ui.Element {
	count := len(state.Filters[column])
	text := label
	if count > 0 {
		text = fmt.Sprintf("%s (%d)", label, count)
	}
	return tableAction(c, text, false).Menu(func(m *ui.Menu) {
		for _, option := range options {
			on := slices.Contains(state.Filters[column], option.Value)
			if m.Item(option.Value).ID(column + ":" + option.Value).Checked(on).Subtitle(fmt.Sprintf("%d matching rows", option.Count)).Chosen() {
				values := slices.Clone(state.Filters[column])
				if i := slices.Index(values, option.Value); i >= 0 {
					values = slices.Delete(values, i, i+1)
				} else {
					values = append(values, option.Value)
				}
				state.SetFilter(column, values)
			}
		}
		if count > 0 {
			m.Separator()
			if m.Item("Clear filter").ID("clear:" + column).Chosen() {
				state.SetFilter(column, nil)
			}
		}
	})
}

func DataTable[T any](c *ui.Context, state *DataTableState, p DataTableProps[T]) ui.Element {
	t := theme.TokensOf(c)
	rows, filtered := dataRecords(state, p)
	total := len(filtered)
	columns := []ui.TableColumn{}
	visible := []int{}
	if p.Selectable {
		columns = append(columns, ui.TableColumn{ID: "__selection", Title: "Select", Width: 54, Fixed: true})
		visible = append(visible, -1)
	}
	for i, col := range p.Columns {
		id := dataColumnID(col)
		if state.Hidden[id] && !col.AlwaysVisible {
			continue
		}
		columns = append(columns, ui.TableColumn{ID: id, Title: col.Title, Width: col.Width, MinWidth: col.MinWidth, MaxWidth: col.MaxWidth, Align: col.Align, Fixed: col.Fixed, RowHeader: col.RowHeader, Sortable: !col.Unsortable && !col.RowHeader})
		visible = append(visible, i)
	}
	state.List.Sort = &state.Sort
	if p.Key != nil || p.Selectable {
		state.List.Key = func(i int) any { return dataRecordKey(p, rows[i]) }
	} else {
		state.List.Key = nil
	}
	if p.Selectable {
		state.List.Selection = &state.Selection
		state.ownSelection = true
	} else if state.ownSelection {
		state.List.Selection = nil
		state.ownSelection = false
	}
	root := ui.Column(c).FillWidth().Gap(12)
	return root.Children(func() {
		if p.Searchable || p.ShowColumnToggle || p.Toolbar != nil {
			ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
				if p.Searchable {
					placeholder := p.FilterPlaceholder
					if placeholder == "" {
						placeholder = "Filter rows…"
					}
					input := ui.TextInputBase(c, &state.Query).Label("Filter rows").Placeholder(placeholder).Height(34).Width(260).MaxWidthPercent(65).PaddingX(10).Radius(t.RadiusMD()).Border(1, t.Input)
					if input.Changed() {
						state.Page = 0
					}
				}
				if p.Toolbar != nil {
					p.Toolbar()
				}
				ui.Box(c).Grow(1)
				if p.ShowColumnToggle {
					tableAction(c, "Columns", false).Menu(func(m *ui.Menu) {
						for _, col := range p.Columns {
							id := dataColumnID(col)
							on := !state.Hidden[id] || col.AlwaysVisible
							if m.Item(col.Title).ID(id).Checked(on).Disabled(col.AlwaysVisible).Chosen() {
								state.SetColumnVisible(id, !on)
							}
						}
					})
				}
			})
		}
		if len(columns) > 0 {
			height := p.Height
			if height <= 0 {
				height = float32(max(len(rows), 1)+1) * 40
			}
			ui.Table(c, &state.List, columns, len(rows), func(row, col int) {
				record := rows[row]
				index := visible[col]
				if index < 0 {
					key := dataRecordKey(p, record)
					label := fmt.Sprintf("Select row %v", key)
					b := tableCheckbox(c, label, state.Selection.Has(key))
					if b.Clicked() {
						if state.Selection.Has(key) {
							state.Selection.Remove(key)
						} else {
							state.Selection.Add(key)
						}
					}
					return
				}
				column := p.Columns[index]
				if column.Cell != nil {
					column.Cell(c, record.value)
				} else if column.Value != nil {
					ui.Text(c, column.Value(record.value)).FontSize(14)
				}
			}).FillWidth().Height(height).Label(p.Label).Border(1, t.Border).Radius(t.Radius)
		} else {
			ui.Text(c, "No visible columns.").FontSize(14).Padding(20).TextColor(t.MutedForeground)
		}
		if total == 0 {
			Typography(c, "No results.", Muted)
		}
		if p.Selectable {
			ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
				selectedFiltered := 0
				for _, row := range filtered {
					if state.Selection.Has(dataRecordKey(p, row)) {
						selectedFiltered++
					}
				}
				ui.Text(c, fmt.Sprintf("%d of %d row(s) selected.", selectedFiltered, total)).Grow(1).FontSize(13).TextColor(t.MutedForeground)
				if tableAction(c, "Select page", len(rows) == 0).Clicked() {
					for _, row := range rows {
						state.Selection.Add(dataRecordKey(p, row))
					}
				}
				if tableAction(c, "Clear selection", state.Selection.Len() == 0).Clicked() {
					state.Selection.Clear()
				}
				if p.SelectionActions != nil {
					selected := []T{}
					for i, row := range p.Rows {
						if state.Selection.Has(dataRecordKey(p, dataRecord[T]{row, i})) {
							selected = append(selected, row)
						}
					}
					p.SelectionActions(selected)
				}
			})
		}
		ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("%d results · Page %d", total, state.Page+1)).Grow(1).FontSize(13).TextColor(t.MutedForeground)
			if tableAction(c, "Previous", state.Page == 0).Clicked() {
				state.Page--
			}
			size := state.PageSize
			if size <= 0 {
				size = 10
			}
			if tableAction(c, "Next", (state.Page+1)*size >= total).Clicked() {
				state.Page++
			}
		})
	})
}
