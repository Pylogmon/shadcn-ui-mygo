package display

import (
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// TableColumn sizes a shared static column. Width is its preferred width;
// unbroken text and native controls can make it wider. Zero shares surplus
// room with other automatic columns. Align uses MyGo's logical alignment.
type TableColumn struct {
	Key                       string
	Width, MinWidth, MaxWidth float32
	Align                     ui.Align
}

type TableOptions struct {
	Label, Description string
	Columns            []TableColumn
	MinWidth           float32
	// Scroll is caller-owned horizontal position. Nil uses MyGo's local scroll.
	Scroll *ui.ScrollState
}

// TableCellProps holds a cell's content. Content replaces Text and builds
// actual native controls in the keyed cell; it is never called to measure it.
type TableCellProps struct {
	Key, Text string
	Content   func(*ui.Context)
	// Style configures the actual native cell after its source defaults. It
	// runs once during the current build; do not retain the Element.
	Style   func(ui.Element)
	ColSpan int
	Align   *ui.Align
	// Checkbox is the public equivalent of CSS :has([role=checkbox]): it
	// removes physical right padding, including in RTL. MyGo exposes no
	// descendant-role query; set this when custom content contains a checkbox.
	Checkbox bool
	Medium   bool
	kind     tableCellKind
}

type TableRowProps struct {
	Key      string
	Cells    []TableCellProps
	Selected bool
	// Style configures the live native row before its cells are built.
	Style func(ui.Element)
	// Expanded follows a child control's aria-expanded state for the source's
	// :has([aria-expanded=true]) background. The child owns its AX action.
	Expanded *bool
}

type TableSection struct {
	rows    []TableRowProps
	kind    tableSectionKind
	caption string
}

type tableSectionKind uint8

const (
	tableHeader tableSectionKind = iota
	tableBody
	tableFooter
	tableCaption
)

type tableCellKind uint8

const (
	tableCell tableCellKind = iota
	tableHead
	tableRowHead
)

// The compound constructors are descriptions, not widgets or retained frame
// objects. TableWith renders them together so all sections share column sizes.
func TableHeader(rows ...TableRowProps) TableSection {
	return TableSection{rows: rows, kind: tableHeader}
}
func TableBody(rows ...TableRowProps) TableSection { return TableSection{rows: rows, kind: tableBody} }
func TableFooter(rows ...TableRowProps) TableSection {
	return TableSection{rows: rows, kind: tableFooter}
}
func TableCaption(text string) TableSection { return TableSection{kind: tableCaption, caption: text} }
func TableRow(props TableRowProps, cells ...TableCellProps) TableRowProps {
	if len(cells) > 0 {
		props.Cells = cells
	}
	return props
}
func TableCell(text string, options ...TableCellProps) TableCellProps {
	return tableDescribeCell(text, tableCell, options)
}
func TableHead(text string, options ...TableCellProps) TableCellProps {
	return tableDescribeCell(text, tableHead, options)
}
func TableRowHead(text string, options ...TableCellProps) TableCellProps {
	return tableDescribeCell(text, tableRowHead, options)
}
func tableDescribeCell(text string, kind tableCellKind, options []TableCellProps) TableCellProps {
	p := TableCellProps{}
	if len(options) > 0 {
		p = options[0]
	}
	p.Text, p.kind = text, kind
	return p
}

type tableIdentity struct {
	Kind    uint8
	Name    string
	Ordinal int
}
type tableCellIdentity struct{ Section, Row, Cell tableIdentity }
type tableLayoutKey struct{}
type tableLayout struct {
	viewportWidth float32
	widths        map[tableCellIdentity]float32
	primed        bool
}

type tableProbeCoordinatorKey struct{}
type tableProbeSentinelKey struct{ layout *tableLayout }
type tableProbeCoordinator struct {
	wave       *tableProbeWave
	correction bool
}
type tableProbeWave struct {
	seen     map[*tableLayout]bool
	drawn    bool
	followup bool
}

// All static tables in one window share a post-layout correction wave. A
// callback in the window overlay paints even when its table is clipped out;
// it carries no drawing, input or accessibility content. A repeated table
// before painting marks another native build pass, and invalidates callbacks
// from the abandoned pass. Only the latest wave may request one follow-up.
func tableProbe(c *ui.Context, layout *tableLayout) {
	coordinator := ui.Local(c.Root(), tableProbeCoordinatorKey{}, func() tableProbeCoordinator { return tableProbeCoordinator{} })
	wave := coordinator.wave
	if wave == nil || wave.drawn || wave.seen[layout] {
		wave = &tableProbeWave{seen: map[*tableLayout]bool{}, followup: !coordinator.correction}
		coordinator.wave = wave
	}
	wave.seen[layout] = true
	if !layout.primed {
		layout.primed = true
		wave.followup = true
	}
	services := c.Services()
	ui.Overlay(c, func() {
		ui.Box(c).Key(tableProbeSentinelKey{layout}).Absolute().Left(0).Top(0).Size(1, 1).
			Role(ui.RoleNone).AccessHidden(true).IgnoreInput(true).PassThrough().
			Draw(func(_ *ui.Painter, _ ui.Rect) {
				if coordinator.wave != wave || wave.drawn {
					return
				}
				wave.drawn = true
				coordinator.correction = wave.followup
				if wave.followup {
					services.Invalidate()
				}
			})
	})
}

type tablePreparedCell struct {
	props     TableCellProps
	key       tableCellIdentity
	col, span int
}
type tablePreparedRow struct {
	props TableRowProps
	key   tableIdentity
	cells []tablePreparedCell
	index int
}
type tablePreparedSection struct {
	kind tableSectionKind
	key  tableIdentity
	rows []tablePreparedRow
}

func tableFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

// Native bounds are float32 DIPs. Above this point even a one-DIP step is
// lost, and summing column widths can overflow. Such dimensions are invalid
// layout inputs rather than useful table sizes; virtual row counts stay free.
const tableMaxDimension float32 = 1 << 24

func tableNonnegative(value float32, invalid *bool) float32 {
	if !tableFinite(value) || value < 0 || value > tableMaxDimension {
		*invalid = true
		return 0
	}
	return value
}
func tableIdentityFor(kind uint8, name string, index int, seen map[string]int) tableIdentity {
	if name == "" {
		name = "index:" + strconv.Itoa(index)
	} else {
		name = "key:" + name
	}
	ordinal := seen[name]
	seen[name]++
	return tableIdentity{kind, name, ordinal}
}

// TableWith is the static source table: natural-height rows, shared measured
// columns, horizontal overflow, footer spans and real native cell contents.
// It adds no selection, sorting, filtering or expansion behavior. Applications
// own those states and actions. For large/lazy data use the virtualized Table.
func TableWith(c *ui.Context, p TableOptions, sections ...TableSection) ui.Element {
	invalid := false
	minimum := tableNonnegative(p.MinWidth, &invalid)
	n := len(p.Columns)
	if n == 0 {
		for _, section := range sections {
			for _, row := range section.rows {
				count := 0
				for _, cell := range row.Cells {
					span := max(1, cell.ColSpan)
					if span > 32767-count {
						count = 32767
						invalid = true
						break
					}
					count += span
				}
				n = max(n, count)
			}
		}
	}
	// Grid cell indexes use int16 in the public runtime. Bound invalid inputs
	// before passing them into native layout; static tables are not virtual data.
	if n > 32767 {
		n = 32767
		invalid = true
	}
	columns := make([]TableColumn, n)
	copy(columns, p.Columns)
	for i := range columns {
		col := &columns[i]
		col.Width = tableNonnegative(col.Width, &invalid)
		col.MinWidth = tableNonnegative(col.MinWidth, &invalid)
		col.MaxWidth = tableNonnegative(col.MaxWidth, &invalid)
		if col.MaxWidth > 0 && col.MaxWidth < col.MinWidth {
			col.MaxWidth = col.MinWidth
			invalid = true
		}
		if col.Align > ui.End {
			col.Align = ui.Start
			invalid = true
		}
	}
	prepared := make([]tablePreparedSection, 0, len(sections))
	caption := ""
	rowIndex := 0
	sectionSeen := map[string]int{}
	active := map[tableCellIdentity]bool{}
	for _, section := range sections {
		if section.kind == tableCaption {
			caption = section.caption
			continue
		}
		sk := tableIdentityFor(uint8(section.kind), strconv.Itoa(int(section.kind)), 0, sectionSeen)
		ps := tablePreparedSection{kind: section.kind, key: sk}
		rowSeen := map[string]int{}
		for i, row := range section.rows {
			rk := tableIdentityFor(4, row.Key, i, rowSeen)
			rowIndex++
			pr := tablePreparedRow{props: row, key: rk, index: rowIndex}
			cellSeen := map[string]int{}
			col := 0
			for j, cell := range row.Cells {
				if col >= n {
					invalid = true
					break
				}
				span := cell.ColSpan
				if span < 0 {
					span = 1
					invalid = true
				}
				if span == 0 {
					span = 1
				}
				if span > n-col {
					span = n - col
					invalid = true
				}
				if cell.Align != nil && *cell.Align > ui.End {
					value := ui.Start
					cell.Align = &value
					invalid = true
				}
				name := cell.Key
				if name == "" && columns[col].Key != "" {
					name = columns[col].Key
				}
				ck := tableCellIdentity{sk, rk, tableIdentityFor(5, name, j, cellSeen)}
				pr.cells = append(pr.cells, tablePreparedCell{cell, ck, col, span})
				active[ck] = true
				col += span
			}
			ps.rows = append(ps.rows, pr)
		}
		prepared = append(prepared, ps)
	}
	viewport := ui.ScrollHorizontal(c).FillWidth().MinWidth(0)
	if p.Scroll != nil {
		for _, offset := range []*float32{&p.Scroll.X, &p.Scroll.Y} {
			if !tableFinite(*offset) {
				*offset = 0
				invalid = true
			}
		}
		viewport.TrackScroll(p.Scroll)
	}
	layout := ui.Local(viewport, tableLayoutKey{}, func() tableLayout { return tableLayout{widths: map[tableCellIdentity]float32{}} })
	if width := viewport.LayoutBounds().W; tableFinite(width) && width > 0 {
		layout.viewportWidth = width
	}
	for key := range layout.widths {
		if !active[key] {
			delete(layout.widths, key)
		}
	}
	widths := make([]float32, n)
	for i, col := range columns {
		widths[i] = max(col.Width, col.MinWidth)
	}
	for _, section := range prepared {
		for _, row := range section.rows {
			for _, cell := range row.cells {
				padding := float32(16)
				if cell.props.Checkbox {
					padding = 8
				}
				want, ok := layout.widths[cell.key]
				if !ok && cell.props.Content == nil {
					weight := 400
					if cell.props.Medium || cell.props.kind != tableCell || section.kind == tableFooter {
						weight = 500
					}
					want, _ = c.MeasureText(0, ui.Span{Text: cell.props.Text, Size: 14, Weight: weight, NoLetterSpacing: true})
				}
				want += padding
				current := float32(0)
				for j := cell.col; j < cell.col+cell.span; j++ {
					current += widths[j]
				}
				if want > current {
					extra := (want - current) / float32(cell.span)
					for j := cell.col; j < cell.col+cell.span; j++ {
						widths[j] += extra
					}
				}
			}
		}
	}
	total := float32(0)
	for i, col := range columns {
		if col.MaxWidth > 0 {
			widths[i] = max(col.MinWidth, min(widths[i], col.MaxWidth))
		}
		widths[i] = max(widths[i], 1)
		total += widths[i]
	}
	available := max(minimum, layout.viewportWidth, total)
	if total < available && n > 0 {
		auto := false
		for _, col := range columns {
			auto = auto || col.Width == 0
		}
		remaining := available - total
		for remaining > .1 {
			flex := 0
			for i, col := range columns {
				if (!auto || col.Width == 0) && (col.MaxWidth == 0 || widths[i] < col.MaxWidth) {
					flex++
				}
			}
			if flex == 0 {
				break
			}
			share := remaining / float32(flex)
			used := float32(0)
			for i, col := range columns {
				if auto && col.Width != 0 {
					continue
				}
				add := share
				if col.MaxWidth > 0 {
					add = min(add, max(float32(0), col.MaxWidth-widths[i]))
				}
				widths[i] += add
				used += add
			}
			if used <= .01 {
				break
			}
			remaining -= used
		}
	}
	// An empty table and capped columns still keep the source w-full surface.
	total = available
	tracks := make([]ui.Track, n)
	for i, width := range widths {
		tracks[i] = ui.Fixed(width)
	}
	services := c.Services()
	viewport.Draw(func(_ *ui.Painter, r ui.Rect) {
		if tableFinite(r.W) && math.Abs(float64(layout.viewportWidth-r.W)) > .1 {
			layout.viewportWidth = r.W
			services.Invalidate()
		}
	})
	t := theme.TokensOf(c)
	result := viewport.Invalid(invalid).Children(func() {
		ui.Column(c).Width(max(total, minimum)).MinWidth(0).Gap(0).Children(func() {
			label := p.Label
			if label == "" {
				label = caption
			}
			table := ui.Column(c).FillWidth().Gap(0).FontSize(14).LineHeight(10.0 / 7).TextColor(t.Foreground).
				Role(ui.RoleTable).Label(label).Description(p.Description).RowCount(rowIndex).ColumnCount(n).Invalid(invalid)
			table.Children(func() {
				for _, section := range prepared {
					group := ui.Column(c.Key(section.key)).FillWidth().Gap(0).Role(ui.RoleNone)
					if section.kind == tableFooter {
						group.Background(t.Muted.Alpha(.5)).FontWeight(500)
						border := t.Border
						group.DrawOver(func(painter *ui.Painter, bounds ui.Rect) {
							painter.Fill(ui.Rect{X: bounds.X, Y: bounds.Y, W: bounds.W, H: min(float32(1), bounds.H)}, border, 0)
						})
					}
					group.Children(func() {
						for i, row := range section.rows {
							r := ui.Grid(c.Key(row.key)).FillWidth().ColumnTracks(tracks...).Gap(0).AlignItems(ui.Stretch).Role(ui.RoleRow).RowIndex(row.index)
							border := section.kind == tableHeader || i < len(section.rows)-1
							if border {
								r.InsetShadow(0, -1, 0, 0, t.Border)
							}
							if row.props.Selected {
								r.Background(t.Muted).AccessSelected(true)
							} else if r.Hovered() || row.props.Expanded != nil && *row.props.Expanded {
								r.Background(t.Muted.Alpha(.5))
							}
							if !motion.Reduced(c) {
								r.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Ease: ui.CubicBezier(.4, 0, .2, 1), Colors: true})
							}
							if row.props.Style != nil {
								row.props.Style(r)
							}
							r.Children(func() {
								for _, cell := range row.cells {
									pc := cell.props
									align := columns[cell.col].Align
									if pc.Align != nil {
										align = *pc.Align
									}
									if align > ui.End {
										align = ui.Start
									}
									// The source explicitly text-left on th; td inherits direction.
									if pc.kind != tableCell && pc.Align == nil && columns[cell.col].Align == ui.Start && c.Direction() == ui.RightToLeft {
										align = ui.End
									}
									box := ui.Column(c.Key(cell.key.Cell)).ColumnStart(cell.col + 1).ColumnSpan(cell.span).MinWidth(0).Padding(8).Justify(ui.Center).AlignItems(align).
										Role(ui.RoleCell).RowIndex(row.index).ColumnIndex(cell.col + 1)
									if pc.Checkbox {
										box.Padding(8, 0, 8, 8)
									}
									if pc.kind != tableCell {
										box.FontWeight(500).PaddingY(0).MinHeight(40)
										if pc.kind == tableHead {
											box.Role(ui.RoleColumnHeader)
										} else {
											box.Role(ui.RoleRowHeader)
										}
									} else if pc.Medium {
										box.FontWeight(500)
									}
									if pc.Content == nil {
										box.Label(pc.Text)
									}
									if pc.Style != nil {
										pc.Style(box)
									}
									box.Children(func() {
										inner := ui.Row(c).AlignSelf(align).Shrink(0).MinWidth(0).AlignItems(ui.Center)
										inner.Children(func() {
											if pc.Content != nil {
												pc.Content(c)
											} else if pc.Text != "" {
												ui.Text(c, pc.Text).SingleLine().Role(ui.RoleNone)
											}
										})
										key := cell.key
										// Draw is culled outside the viewport; committed native
										// layout is still available while constructing this cell.
										if width := inner.LayoutBounds().W; tableFinite(width) && math.Abs(float64(layout.widths[key]-width)) > .1 {
											layout.widths[key] = width
											services.Invalidate()
										}
										inner.Draw(func(_ *ui.Painter, bounds ui.Rect) {
											if tableFinite(bounds.W) && math.Abs(float64(layout.widths[key]-bounds.W)) > .1 {
												layout.widths[key] = bounds.W
												services.Invalidate()
											}
										})
									})
								}
							})
						}
					})
				}
			})
			if caption != "" {
				ui.Box(c.Key(tableIdentity{Kind: 8})).Height(16)
				ui.Text(c.Key(tableIdentity{Kind: 7}), caption).FillWidth().TextAlign(ui.Center).FontSize(14).LineHeight(10.0 / 7).TextColor(t.MutedForeground)
			}
		})
	})
	tableProbe(c, layout)
	return result
}
