package gallery

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

type tablePreviewRecord struct {
	selected     bool
	status, note string
}
type tablePreviewState struct {
	records map[string]*tablePreviewRecord
	order   []string
	empty   bool
	scroll  ui.ScrollState
}

func (s *State) tableCompositionPreview(c *ui.Context) {
	p := &s.tablePreview
	if p.records == nil {
		p.records = map[string]*tablePreviewRecord{}
		for _, invoice := range invoices {
			p.order = append(p.order, invoice.ID)
			p.records[invoice.ID] = &tablePreviewRecord{status: invoice.Status}
		}
	}
	display.Card(c, func() {
		s.cardHeader(c, "Table composition", "Select a row, change its status, edit a note or open its actions.")
		ui.Row(c).Gap(12).Wrap().AlignItems(ui.Center).Children(func() {
			control.Switch(c, &p.empty, control.SwitchProps{Label: "Show empty table"})
			if control.Button(c, "Reverse invoice rows", control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
				slices.Reverse(p.order)
			}
		})
		rows := make([]display.TableRowProps, 0, len(p.order))
		total := float64(0)
		for _, id := range p.order {
			if p.empty {
				break
			}
			record := p.records[id]
			amount := float64(0)
			for _, invoice := range invoices {
				if invoice.ID == id {
					amount = invoice.Amount
					break
				}
			}
			total += amount
			rows = append(rows, display.TableRow(display.TableRowProps{Key: id, Selected: record.selected},
				display.TableCell("", display.TableCellProps{Key: "selected", Checkbox: true, Content: func(c *ui.Context) {
					control.Checkbox(c, &record.selected, "Select "+id, control.CheckboxProps{Content: func(*ui.Context) {}})
				}}),
				display.TableRowHead(id, display.TableCellProps{Key: "invoice"}),
				display.TableCell("", display.TableCellProps{Key: "status", Content: func(c *ui.Context) {
					props := control.SelectProps{Label: "Status " + id, Size: control.Small,
						Options:        []control.Option{{Value: "Paid", Label: "Paid"}, {Value: "Pending", Label: "Pending"}, {Value: "Processing", Label: "Processing"}},
						TriggerContent: func(c *ui.Context, option control.Option) { display.Badge(c, option.Label, display.BadgeSecondary) },
					}
					if s.Config.NativeMenus {
						desktop.NativeSelect(c, &record.status, props)
					} else {
						if _, err := desktop.Select(c, s.Popovers, "table-status-"+id, &record.status, props); err != nil {
							s.popupError(c, err)
						}
					}
				}}),
				display.TableCell(fmt.Sprintf("$%.2f", amount), display.TableCellProps{Key: "amount"}),
				display.TableCell("", display.TableCellProps{Key: "note", Content: func(c *ui.Context) {
					control.Input(c, &record.note, control.InputProps{Label: "Note " + id, Placeholder: "Add a note", Size: control.Small}).Width(180)
				}}),
				display.TableCell("", display.TableCellProps{Key: "actions", Content: func(c *ui.Context) {
					s.previewMenu(c, "table-actions-"+id, "Actions "+id, overlay.MenuProps{Items: []overlay.MenuItem{
						{Key: "View invoice " + id, Label: "View invoice"}, {Key: "Download invoice " + id, Label: "Download receipt"},
					}})
				}}),
			))
		}
		if p.empty {
			rows = append(rows, display.TableRow(display.TableRowProps{Key: "empty"}, display.TableCell("No invoices to display.", display.TableCellProps{ColSpan: 6})))
		}
		display.TableWith(c, display.TableOptions{Label: "Editable invoices", Scroll: &p.scroll, Columns: []display.TableColumn{
			{Key: "selected", Width: 64}, {Key: "invoice", Width: 100}, {Key: "status", Width: 168},
			{Key: "amount", Width: 110, Align: ui.End}, {Key: "note", Width: 196}, {Key: "actions", Width: 160},
		}}, display.TableHeader(display.TableRow(display.TableRowProps{Key: "header"},
			display.TableHead("Select"), display.TableHead("Invoice"), display.TableHead("Status"),
			display.TableHead("Amount"), display.TableHead("Note"), display.TableHead("Actions"),
		)), display.TableBody(rows...), display.TableFooter(display.TableRow(display.TableRowProps{Key: "total"},
			display.TableCell("Total", display.TableCellProps{ColSpan: 3}), display.TableCell(fmt.Sprintf("$%.2f", total)),
			display.TableCell("", display.TableCellProps{ColSpan: 2}),
		)), display.TableCaption("A list of recent invoices."))
	})
}
