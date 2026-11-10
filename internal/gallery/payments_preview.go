package gallery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type paymentPreviewRow struct {
	ID, Status, Email string
	Amount            float64
}

// The five records in the fixed data-table-demo.tsx, in original order.
var paymentPreviewRows = []paymentPreviewRow{
	{"m5gr84i9", "success", "ken99@example.com", 316},
	{"3u1reuv4", "success", "Abe45@example.com", 242},
	{"derv1ws0", "processing", "Monserrat44@example.com", 837},
	{"5kma53ae", "success", "Silas22@example.com", 874},
	{"bhqecj4p", "failed", "carmella@example.com", 721},
}

type paymentPreviewMenu struct {
	open  bool
	state desktop.MenuState
}

// App-owned query, sort, selection, visibility and paging use public table
// values. Menu state/scroll survive frames; no Context/Element is retained.
type paymentsPreviewState struct {
	ready, native                  bool
	table                          display.DataTableState
	visible                        [3]bool
	columns, pageSize              paymentPreviewMenu
	rows                           map[string]*paymentPreviewMenu
	scroll                         ui.ScrollState
	detailID, detailKind           string
	feedback, popupError, copiedID string
	copySerial                     uint64
}

func (p *paymentsPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native = true, native
	p.table.PageSize = 10
	p.visible = [3]bool{true, true, true}
	p.rows = map[string]*paymentPreviewMenu{}
}
func paymentsDataProps() display.DataTableProps[paymentPreviewRow] {
	return display.DataTableProps[paymentPreviewRow]{Rows: paymentPreviewRows, Key: func(row paymentPreviewRow) any { return row.ID }, Search: func(row paymentPreviewRow, query string) bool {
		return strings.Contains(strings.ToLower(row.Email), query)
	}, Columns: []display.DataColumn[paymentPreviewRow]{
		{ID: "status", Title: "Status", Value: func(row paymentPreviewRow) string { return row.Status }, Unsortable: true},
		{ID: "email", Title: "Email", Value: func(row paymentPreviewRow) string { return strings.ToLower(row.Email) }, Compare: func(a, b paymentPreviewRow) int {
			return strings.Compare(strings.ToLower(a.Email), strings.ToLower(b.Email))
		}},
		{ID: "amount", Title: "Amount", Value: func(row paymentPreviewRow) string { return fmt.Sprintf("$%.2f", row.Amount) }, Unsortable: true},
	}}
}
func (p *paymentsPreviewState) filteredRows(props display.DataTableProps[paymentPreviewRow]) []paymentPreviewRow {
	// Ask the delivered row model for the complete filtered set, using only its
	// public value inputs. Sorting/filter algorithms are not duplicated here.
	all := display.DataTableState{Query: p.table.Query, Sort: p.table.Sort, Filters: p.table.Filters, PageSize: max(1, len(props.Rows))}
	rows, _ := display.DataRows(&all, props)
	return rows
}
func (p *paymentsPreviewState) copyID(c *ui.Context, id string) {
	p.copySerial++
	serial := p.copySerial
	p.copiedID = ""
	p.feedback = "Copying payment ID…"
	c.WriteClipboardThen(id, func(err error) {
		if serial != p.copySerial {
			return
		}
		if err != nil {
			p.feedback = "Could not copy payment ID " + id + ". Try again."
			return
		}
		p.copiedID = id
		p.feedback = "Copied payment ID " + id + "."
	})
}
func (p *paymentsPreviewState) menu(c *ui.Context, h *desktop.Host, native bool, key, label, caption string, items []overlay.MenuItem, state *paymentPreviewMenu) string {
	props := control.ButtonProps{Variant: control.Outline, Size: control.Small, Label: label}
	if caption == "" {
		props.Variant, props.Size, props.Icon = control.Ghost, control.IconExtraSmall, outlined.MoreHoriz()
	}
	content := func(e ui.Element) {
		if caption != "" {
			e.Children(func() { material.Icon(c, outlined.ExpandMore(), material.Props{Size: 14}) })
		}
		if e.Clicked() {
			p.popupError = ""
		}
	}
	chosen := ""
	var err error
	if native {
		chosen, err = desktop.NativeMenu(c, overlay.MenuProps{Items: items}, func(e ui.Element) { control.StyleButton(c, e, caption, props); content(e) })
	} else {
		anchor := control.Button(c, caption, props)
		content(anchor)
		chosen, err = desktop.DropdownMenu(c, h, key, anchor, &state.open, &state.state, desktop.MenuProps{Label: label, Items: items, Placement: mygo.PlacementBottomEnd, Width: 210, Gap: 4})
	}
	if err != nil {
		p.popupError = "Could not open payment menu: " + err.Error()
	}
	return chosen
}
func (p *paymentsPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		p.native = native
		for key, state := range p.rows {
			state.open = false
			state.state.Reset()
			if h != nil {
				h.Close("payment-actions-" + key)
			}
		}
		p.columns.open, p.pageSize.open = false, false
		p.columns.state.Reset()
		p.pageSize.state.Reset()
		if h != nil {
			h.Close("payment-columns")
			h.Close("payment-page-size")
		}
	}
	t := theme.TokensOf(c)
	display.Card(c.Key("payments-source-preview"), func() {
		ui.Column(c).FillWidth().Gap(16).Children(func() {
			display.Typography(c, "Payments", display.H3)
			ui.Text(c, "Filter recent payments, choose columns, and select the rows on each page.").TextColor(t.MutedForeground)
			ui.Row(c).FillWidth().Wrap().Gap(8).AlignItems(ui.Center).Children(func() {
				input := control.Input(c, &p.table.Query, control.InputProps{Label: "Filter payment emails", Placeholder: "Filter emails…", OnChange: func(string) { p.table.Page = 0 }}).Width(300).MaxWidthPercent(100)
				if input.Changed() {
					c.Invalidate()
				}
				ui.Box(c).Grow(1)
				items := []overlay.MenuItem{{Key: "status", Label: "Status", Checked: &p.visible[0]}, {Key: "email", Label: "Email", Checked: &p.visible[1]}, {Key: "amount", Label: "Amount", Checked: &p.visible[2]}}
				if chosen := p.menu(c, h, native, "payment-columns", "Payment columns", "Columns", items, &p.columns); chosen != "" {
					for i, id := range []string{"status", "email", "amount"} {
						if chosen == id {
							p.table.SetColumnVisible(id, p.visible[i])
							break
						}
					}
					c.Invalidate()
				}
			})
			props := paymentsDataProps()
			page, total := display.DataRows(&p.table, props)
			amountAlign := ui.End
			if c.Direction() == ui.RightToLeft {
				amountAlign = ui.Start
			}
			headerAlign := ui.Start
			columns := []display.TableColumn{{Key: "select", Width: 48}}
			headers := []display.TableCellProps{display.TableHead("", display.TableCellProps{Key: "select", Checkbox: true, Align: &headerAlign, Style: func(e ui.Element) { e.Label("Payment selection") }, Content: func(child *ui.Context) {
				count := 0
				for _, row := range page {
					if p.table.Selection.Has(row.ID) {
						count++
					}
				}
				state := ui.CheckOff
				if count > 0 && count < len(page) {
					state = ui.CheckMixed
				} else if count > 0 {
					state = ui.CheckOn
				}
				if control.CheckboxState(child, &state, "Select all payments on page", control.CheckboxProps{Disabled: len(page) == 0, Content: func(*ui.Context) {}}).Changed() {
					for _, row := range page {
						if state == ui.CheckOn {
							p.table.Selection.Add(row.ID)
						} else {
							p.table.Selection.Remove(row.ID)
						}
					}
					child.Invalidate()
				}
			}})}
			if !p.table.Hidden["status"] {
				columns = append(columns, display.TableColumn{Key: "status", Width: 120})
				headers = append(headers, display.TableHead("Status", display.TableCellProps{Key: "status", Align: &headerAlign}))
			}
			if !p.table.Hidden["email"] {
				columns = append(columns, display.TableColumn{Key: "email", Width: 240})
				headers = append(headers, display.TableHead("", display.TableCellProps{Key: "email", Align: &headerAlign, Style: func(e ui.Element) {
					direction := ui.SortNone
					if p.table.Sort.Column == "email" {
						direction = ui.SortAscending
						if p.table.Sort.Descending {
							direction = ui.SortDescending
						}
					}
					e.Label("Email").Sort(direction)
				}, Content: func(child *ui.Context) {
					if control.Button(child, "Email", control.ButtonProps{Label: "Sort payment emails", Variant: control.Ghost, Size: control.Small, Icon: outlined.SwapVert()}).Clicked() {
						p.table.Sort = ui.SortOrder{Column: "email", Descending: p.table.Sort.Column == "email" && !p.table.Sort.Descending}
						child.Invalidate()
					}
				}}))
			}
			if !p.table.Hidden["amount"] {
				columns = append(columns, display.TableColumn{Key: "amount", Width: 112, Align: amountAlign})
				headers = append(headers, display.TableHead("Amount", display.TableCellProps{Key: "amount", Align: &amountAlign}))
			}
			columns = append(columns, display.TableColumn{Key: "actions", Width: 56})
			headers = append(headers, display.TableHead("", display.TableCellProps{Key: "actions", Align: &headerAlign, Style: func(e ui.Element) { e.Label("Payment actions") }}))
			body := make([]display.TableRowProps, 0, len(page))
			for _, record := range page {
				row := record
				cells := []display.TableCellProps{display.TableCell("", display.TableCellProps{Key: "select", Checkbox: true, Content: func(child *ui.Context) {
					selected := p.table.Selection.Has(row.ID)
					if control.Checkbox(child, &selected, "Select payment "+row.ID, control.CheckboxProps{Content: func(*ui.Context) {}}).Changed() {
						if selected {
							p.table.Selection.Add(row.ID)
						} else {
							p.table.Selection.Remove(row.ID)
						}
						child.Invalidate()
					}
				}})}
				if !p.table.Hidden["status"] {
					cells = append(cells, display.TableCell(strings.ToUpper(row.Status[:1])+row.Status[1:], display.TableCellProps{Key: "status"}))
				}
				if !p.table.Hidden["email"] {
					cells = append(cells, display.TableCell(strings.ToLower(row.Email), display.TableCellProps{Key: "email"}))
				}
				if !p.table.Hidden["amount"] {
					cells = append(cells, display.TableCell(fmt.Sprintf("$%.2f", row.Amount), display.TableCellProps{Key: "amount", Medium: true}))
				}
				cells = append(cells, display.TableCell("", display.TableCellProps{Key: "actions", Content: func(child *ui.Context) {
					state := p.rows[row.ID]
					if state == nil {
						state = &paymentPreviewMenu{}
						p.rows[row.ID] = state
					}
					switch p.menu(child, h, native, "payment-actions-"+row.ID, "Payment actions "+row.ID, "", []overlay.MenuItem{{Key: "copy", Label: "Copy payment ID", Icon: outlined.ContentCopy(), Group: "Actions"}, {Separator: true}, {Key: "customer", Label: "View customer"}, {Key: "details", Label: "View payment details"}}, state) {
					case "copy":
						p.copyID(child, row.ID)
					case "customer":
						p.detailID, p.detailKind = row.ID, "customer"
					case "details":
						p.detailID, p.detailKind = row.ID, "details"
					}
				}}))
				body = append(body, display.TableRow(display.TableRowProps{Key: row.ID, Selected: p.table.Selection.Has(row.ID)}, cells...))
			}
			if len(body) == 0 {
				center := ui.Center
				body = append(body, display.TableRow(display.TableRowProps{Key: "empty-payments"}, display.TableCell("No results.", display.TableCellProps{Key: "empty", ColSpan: len(columns), Align: &center, Style: func(e ui.Element) { e.Height(96) }})))
			}
			display.TableWith(c.Key("payments-table"), display.TableOptions{Label: "Recent payments", Scroll: &p.scroll, MinWidth: 600, Columns: columns}, display.TableHeader(display.TableRow(display.TableRowProps{Key: "header"}, headers...)), display.TableBody(body...))
			filtered := p.filteredRows(props)
			selected := 0
			for _, row := range filtered {
				if p.table.Selection.Has(row.ID) {
					selected++
				}
			}
			ui.Row(c).FillWidth().Wrap().Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, fmt.Sprintf("%d of %d row(s) selected.", selected, total)).Grow(1).FontSize(13).TextColor(t.MutedForeground)
				ui.Text(c, "Rows per page").FontSize(13)
				if chosen := p.menu(c, h, native, "payment-page-size", "Payments page size", strconv.Itoa(p.table.PageSize), []overlay.MenuItem{{Key: "2", Label: "2 per page (demo)"}, {Key: "5", Label: "5 per page"}, {Key: "10", Label: "10 per page"}}, &p.pageSize); chosen != "" {
					size, err := strconv.Atoi(chosen)
					if err == nil {
						p.table.PageSize, p.table.Page = size, 0
						c.Invalidate()
					}
				}
				pages := max(1, (total+p.table.PageSize-1)/p.table.PageSize)
				ui.Text(c, fmt.Sprintf("Page %d of %d", p.table.Page+1, pages)).FontSize(13)
				if control.Button(c, "Previous", control.ButtonProps{Label: "Previous payments page", Variant: control.Outline, Size: control.Small, Disabled: p.table.Page == 0}).Clicked() {
					p.table.Page--
					c.Invalidate()
				}
				if control.Button(c, "Next", control.ButtonProps{Label: "Next payments page", Variant: control.Outline, Size: control.Small, Disabled: p.table.Page+1 >= pages}).Clicked() {
					p.table.Page++
					c.Invalidate()
				}
			})
			if p.detailID != "" {
				for _, row := range paymentPreviewRows {
					if row.ID != p.detailID {
						continue
					}
					ui.Column(c.Key("payment-details")).FillWidth().Padding(12).Gap(6).Border(1, t.Border).Radius(t.RadiusMD()).Children(func() {
						if p.detailKind == "customer" {
							ui.Text(c, "Customer contact").FontWeight(600)
							ui.Text(c, row.Email)
						} else {
							ui.Text(c, "Payment details · "+row.ID).FontWeight(600)
							ui.Text(c, fmt.Sprintf("%s · $%.2f · %s", row.Status, row.Amount, row.Email))
						}
						if control.Button(c, "Close details", control.ButtonProps{Variant: control.Ghost, Size: control.ExtraSmall}).Clicked() {
							p.detailID = ""
							c.Invalidate()
						}
					})
					break
				}
			}
			if p.feedback != "" {
				ui.Text(c, p.feedback).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(c, p.popupError).Role(ui.RoleAlert)
			}
		})
	}).Role(ui.RoleGroup).Label("Payments example")
}
