package display

import (
	"github.com/egoist/mygo/ui"
)

type TableProps struct {
	Columns []ui.TableColumn
	Rows    [][]string
	Caption string
}

func Table(c *ui.Context, p TableProps) ui.Element {
	root := ui.Column(c).FillWidth().Gap(8)
	state := ui.Local(root, "table", func() ui.ListState { return ui.ListState{} })
	return root.Children(func() {
		ui.Table(c, state, p.Columns, len(p.Rows), func(row, col int) {
			if col < len(p.Rows[row]) {
				ui.Text(c, p.Rows[row][col]).FontSize(14)
			}
		}).FillWidth().Height(float32(min(len(p.Rows), 10)+1) * 40)
		if p.Caption != "" {
			Typography(c, p.Caption, Muted).AlignSelf(ui.Center)
		}
	})
}
