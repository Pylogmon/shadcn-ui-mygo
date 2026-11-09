package control

import "github.com/egoist/mygo/ui"

// Stored navigation metadata contains values only, never elements or builders.
type comboGridCell struct {
	key         string
	row, column int
}

func (s *comboRuntime) gridMove(key ui.Key) bool {
	if !s.grid || !s.open {
		return false
	}
	vertical := key == ui.KeyUp || key == ui.KeyDown
	if !vertical && key != ui.KeyLeft && key != ui.KeyRight {
		return false
	}
	items := []int{}
	for i, e := range s.items {
		if !e.option.Separator {
			items = append(items, i)
		}
	}
	if len(items) == 0 {
		return true
	}
	current := -1
	for n, i := range items {
		if s.hasActive && s.items[i].option.Value == s.active {
			current = n
			break
		}
	}
	if current < 0 {
		if vertical {
			s.edge(key == ui.KeyUp)
		}
		return true
	}
	cols := comboGridColumns(s.gridColumns)
	positions := make([]comboGridCell, len(items))
	for n, i := range items {
		positions[n] = comboGridCell{key: s.items[i].option.Value, row: n / cols, column: n % cols}
		for _, cell := range s.gridCells {
			if cell.key == positions[n].key {
				positions[n] = cell
				break
			}
		}
	}
	row, column := positions[current].row, positions[current].column
	lastRow := 0
	for _, cell := range positions {
		lastRow = max(lastRow, cell.row)
	}
	step := 1
	if key == ui.KeyUp || key == ui.KeyLeft && s.direction != ui.RightToLeft || key == ui.KeyRight && s.direction == ui.RightToLeft {
		step = -1
	}
	if vertical {
		for nextRow, tries := row+step, 0; tries <= lastRow; tries, nextRow = tries+1, nextRow+step {
			if nextRow < 0 || nextRow > lastRow {
				if s.noLoop {
					return true
				}
				if nextRow < 0 {
					nextRow = lastRow
				} else {
					nextRow = 0
				}
			}
			candidate, bestCol := -1, -1
			for n, cell := range positions {
				if cell.row == nextRow && cell.column <= column && cell.column > bestCol && s.eligible(items[n]) {
					candidate, bestCol = n, cell.column
				}
			}
			if candidate >= 0 {
				s.setActive(items[candidate])
				return true
			}
		}
		return true
	}
	rowItems := []int{}
	at := -1
	for n, cell := range positions {
		if cell.row == row {
			if n == current {
				at = len(rowItems)
			}
			rowItems = append(rowItems, n)
		}
	}
	for n, tries := at+step, 0; tries < len(rowItems); n, tries = n+step, tries+1 {
		if n < 0 || n >= len(rowItems) {
			if s.noLoop {
				return true
			}
			n = (n + len(rowItems)) % len(rowItems)
		}
		if s.eligible(items[rowItems[n]]) {
			s.setActive(items[rowItems[n]])
			return true
		}
	}
	return true
}

// Row supplies real native row/cell accessibility and row-aware vertical grid
// navigation, including uneven rows. Index is zero based in the filtered grid.
func (b *ComboboxScope[T]) Row(c *ui.Context, index int, children func()) ui.Element {
	f := b.f
	f.check()
	previousRow, previousColumn, previousInside := f.row, f.itemIndex, f.inRow
	f.row, f.itemIndex, f.inRow = index, 0, true
	defer func() { f.row, f.itemIndex, f.inRow = previousRow, previousColumn, previousInside }()
	return ui.Row(c.Key(struct{ ComboRow int }{index})).FillWidth().Role(ui.RoleRow).RowIndex(index + 1).Children(children)
}

func comboGridColumns(columns int) int {
	if columns <= 0 {
		return 2
	}
	return columns
}
