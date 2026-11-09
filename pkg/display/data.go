package display

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type TableProps struct {
	Columns            []ui.TableColumn
	Rows               [][]string
	Caption            string
	Label, Description string
	// State keeps native virtualization, focus, row keys, column sizing and
	// optional selection in application state. Nil uses viewport-local state.
	State *ui.ListState
	// Height bounds the native viewport. Zero keeps the compatible ten-row
	// preview limit. RowCount permits a lazy CellContent without allocating Rows.
	Height      float32
	RowCount    int
	CellContent func(*ui.Context, int, int)
}

// Table preserves the native virtualized table entry point. It builds only
// visible rows, using State.Key for stable item identity. TableWith provides
// the source's composable static headers, footer, spans and arbitrary cells.
func Table(c *ui.Context, p TableProps) ui.Element {
	root := ui.Column(c).FillWidth().Gap(16)
	state := p.State
	if state == nil {
		state = ui.Local(root, "table", func() ui.ListState { return ui.ListState{} })
	}
	n := len(p.Rows)
	if p.CellContent != nil && p.RowCount > 0 {
		n = p.RowCount
	}
	height := p.Height
	invalid := p.RowCount < 0 || math.IsNaN(float64(height)) || math.IsInf(float64(height), 0) || height < 0 || height > tableMaxDimension
	if height <= 0 || invalid {
		height = float32(min(n, 10)+1) * 40
	}
	columns := append([]ui.TableColumn(nil), p.Columns...)
	for i := range columns {
		for _, value := range []*float32{&columns[i].Width, &columns[i].MinWidth, &columns[i].MaxWidth} {
			if !tableFinite(*value) || *value < 0 || *value > tableMaxDimension {
				*value = 0
				invalid = true
			}
		}
		if columns[i].MaxWidth > 0 && columns[i].MaxWidth < columns[i].MinWidth {
			columns[i].MaxWidth = columns[i].MinWidth
			invalid = true
		}
	}
	t := theme.TokensOf(c)
	return root.Children(func() {
		ui.Table(c, state, columns, n, func(row, col int) {
			if p.CellContent != nil {
				p.CellContent(c, row, col)
			} else if row < len(p.Rows) && col < len(p.Rows[row]) {
				ui.Text(c, p.Rows[row][col]).FontSize(14).SingleLine()
			}
		}).FillWidth().Height(height).Label(p.Label).Description(p.Description).Invalid(invalid)
		if p.Caption != "" {
			ui.Text(c, p.Caption).FontSize(14).LineHeight(10.0 / 7).TextColor(t.MutedForeground).AlignSelf(ui.Center)
		}
	})
}
