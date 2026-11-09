package display

import (
	"math"
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type TextStyle int

const (
	Paragraph TextStyle = iota
	H1
	H2
	H3
	H4
	Lead
	Large
	Small
	Muted
	Blockquote
	InlineCode
)

type typographyFlowKey struct{}
type typographyFlow struct {
	count     int
	bottom    float32
	suspended bool
}
type typographyFlows struct{ active []*typographyFlow }

func typographyFlowOf(c *ui.Context) *typographyFlow {
	s := ui.Local(c.Root(), typographyFlowKey{}, func() typographyFlows { return typographyFlows{} })
	if len(s.active) == 0 {
		return nil
	}
	return s.active[len(s.active)-1]
}

// TypographyProse supplies the document's spacing without installing global
// styles. Nested text, links and lists keep their actual native semantics.
func TypographyProse(c *ui.Context, content func(*ui.Context)) ui.Element {
	root := ui.Column(c).FillWidth().MinWidth(0).Gap(0).Role(ui.RoleArticle).FontSize(16).LineHeight(1.5).TextColor(theme.TokensOf(c).Foreground)
	s := ui.Local(c.Root(), typographyFlowKey{}, func() typographyFlows { return typographyFlows{} })
	before := len(s.active)
	s.active = append(s.active, &typographyFlow{})
	defer func() { s.active = s.active[:before] }()
	return root.Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func typographyFlowBlock(c *ui.Context, e ui.Element, style TextStyle) ui.Element {
	s := typographyFlowOf(c)
	if s == nil || s.suspended {
		return e
	}
	top, bottom := float32(0), float32(0)
	switch style {
	case Paragraph, Lead:
		if s.count > 0 {
			top = 24
		}
	case H2:
		if s.count > 0 {
			top = 40
		}
	case H3, H4:
		if s.count > 0 {
			top = 32
		}
	case Blockquote:
		top = 24
	}
	e.Margin(max(0, top-s.bottom), 0, bottom)
	s.count++
	s.bottom = bottom
	return e
}

func typographyFlowContainer(c *ui.Context, e ui.Element, content func()) ui.Element {
	s := typographyFlowOf(c)
	if s == nil || s.suspended {
		return e.Children(content)
	}
	e.Margin(max(0, float32(24)-s.bottom), 0, 24)
	s.count++
	s.bottom = 24
	s.suspended = true
	defer func() { s.suspended = false }()
	return e.Children(content)
}

func typographyAppearance(c *ui.Context, e ui.Element, style TextStyle) ui.Element {
	t := theme.TokensOf(c)
	e.MinWidth(0).TextColor(t.Foreground).FontSize(16).FontWeight(400).LetterSpacing(0).TextAlign(ui.Start).LineHeight(1.5)
	switch style {
	case H1:
		e.FontSize(36).FontWeight(800).LineHeight(40.0 / 36.0).LetterSpacing(-.9).Balance().Role(ui.RoleHeading).Level(1)
	case H2:
		e.FontSize(30).FontWeight(600).LineHeight(1.2).LetterSpacing(-.75).Padding(0, 0, 8, 0).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Role(ui.RoleHeading).Level(2)
	case H3:
		e.FontSize(24).FontWeight(600).LineHeight(4.0 / 3.0).LetterSpacing(-.6).Role(ui.RoleHeading).Level(3)
	case H4:
		e.FontSize(20).FontWeight(600).LineHeight(1.4).LetterSpacing(-.5).Role(ui.RoleHeading).Level(4)
	case Paragraph:
		e.LineHeight(1.75)
	case Lead:
		e.FontSize(20).LineHeight(1.4).TextColor(t.MutedForeground)
	case Large:
		e.FontSize(18).FontWeight(600).LineHeight(28.0 / 18.0)
	case Small:
		e.FontSize(14).FontWeight(500).LineHeight(1)
	case Muted:
		e.FontSize(14).LineHeight(20.0 / 14.0).TextColor(t.MutedForeground)
	case Blockquote:
		e.Italic().PaddingStart(24).BorderColor(t.Border)
		if c.Direction() == ui.RightToLeft {
			e.BorderWidth(0, 2, 0, 0)
		} else {
			e.BorderWidth(0, 0, 0, 2)
		}
	case InlineCode:
		e.FontSize(14).FontWeight(600).Font("monospace").LineHeight(1.5)
	}
	return typographyFlowBlock(c, e, style)
}

func Typography(c *ui.Context, text string, style TextStyle) ui.Element {
	if style == InlineCode {
		return typographyAppearance(c, ui.RichText(c, TypographyCodeSpan(c, text)), style).AlignSelf(ui.Start)
	}
	return typographyAppearance(c, ui.Text(c, text), style)
}

// TypographyRich wraps native inline Text/Link/RichText children as one
// paragraph. Arbitrary layout controls belong outside this text-only callback.
func TypographyRich(c *ui.Context, style TextStyle, content func(*ui.Context)) ui.Element {
	root := typographyAppearance(c, ui.RichText(c), style)
	s := typographyFlowOf(c)
	if s != nil {
		previous := s.suspended
		s.suspended = true
		defer func() { s.suspended = previous }()
	}
	return root.Children(func() {
		if content != nil {
			content(c)
		}
	})
}

// TypographyCodeSpan has real native inline padding and rounded background;
// ordinary element padding is not preserved by MyGo inline text merging.
func TypographyCodeSpan(c *ui.Context, text string) ui.Span {
	return ui.Span{Text: text, Font: "monospace", Size: 14, Weight: 600, Color: theme.TokensOf(c).Foreground,
		Background: theme.TokensOf(c).Muted, Radius: 4, PaddingX: 4.8, PaddingY: 3.2, NoLetterSpacing: true}
}

func TypographyCode(c *ui.Context, text string) ui.Element {
	return ui.RichText(c, TypographyCodeSpan(c, text)).AlignSelf(ui.Start).MinWidth(0).FontSize(14).LineHeight(1.5)
}

func TypographyLink(c *ui.Context, text, url string) ui.Element {
	t := theme.TokensOf(c)
	return ui.Link(c, text, url).FontWeight(500).TextColor(t.Primary).Underline().UnderlineOffset(4).FocusRingColor(t.Ring)
}

type TypographyListItem struct {
	Key, Text string
	Content   func(*ui.Context)
}
type TypographyListProps struct {
	Label   string
	Ordered bool
	Start   int
}
type typographyItemKey struct {
	kind    uint8
	key     string
	ordinal int
}

func TypographyList(c *ui.Context, items []TypographyListItem, options ...TypographyListProps) ui.Element {
	p := TypographyListProps{}
	if len(options) > 0 {
		p = options[0]
	}
	root := ui.Column(c).FillWidth().MinWidth(0).Gap(8).Padding(8, 0, 0, 0).PaddingStart(24).Role(ui.RoleList).Label(p.Label).
		FontSize(16).LineHeight(1.5).TextColor(theme.TokensOf(c).Foreground)
	start := p.Start
	if start == 0 {
		start = 1
	}
	if len(items) > 0 && start > math.MaxInt-(len(items)-1) {
		start = 1
		root.Invalid(true)
	}
	markerWidth, indent := float32(16), float32(24)
	if p.Ordered {
		for i := range items {
			width, _ := c.MeasureText(0, ui.Span{Text: strconv.Itoa(start+i) + ".", Size: 16, Weight: 400, NoLetterSpacing: true})
			markerWidth = max(markerWidth, width)
		}
		indent = max(indent, markerWidth+8)
		root.PaddingStart(indent)
	}
	return typographyFlowContainer(c, root, func() {
		occurrences := map[string]int{}
		for i, item := range items {
			key := item.Key
			if key == "" {
				key = "index:" + strconv.Itoa(i)
			} else {
				key = "key:" + key
			}
			ordinal := occurrences[key]
			occurrences[key]++
			ui.Column(c.Key(typographyItemKey{0, key, ordinal})).FillWidth().MinWidth(0).Role(ui.RoleListItem).PositionInSet(i+1, len(items)).Children(func() {
				marker := "•"
				if p.Ordered {
					marker = strconv.Itoa(start+i) + "."
				}
				mark := ui.Text(c.Key(typographyItemKey{1, "marker", 0}), marker).Absolute().Top(0).Width(markerWidth).IgnoreInput(true).AccessHidden(true).TextAlign(ui.End)
				if c.Direction() == ui.RightToLeft {
					mark.Right(-indent)
				} else {
					mark.Left(-indent)
				}
				ui.Column(c.Key(typographyItemKey{2, "content", 0})).FillWidth().MinWidth(0).Children(func() {
					if item.Content != nil {
						item.Content(c)
					} else {
						ui.Text(c, item.Text)
					}
				})
			})
		}
	})
}

type TypographyTableColumn struct {
	Key, Title string
	Align      ui.Align
}
type TypographyTableRow struct {
	Key   string
	Cells []string
}
type TypographyTableProps struct {
	Label   string
	Columns []TypographyTableColumn
	Rows    []TypographyTableRow
	// MinWidth permits a readable horizontally scrolling table in a narrow
	// native viewport. Zero lets equal columns wrap inside the available width.
	MinWidth    float32
	CellContent func(*ui.Context, int, int)
}

func TypographyTable(c *ui.Context, p TypographyTableProps) ui.Element {
	t := theme.TokensOf(c)
	viewport := ui.ScrollHorizontal(c).FillWidth().MinWidth(0)
	minimum := p.MinWidth
	if minimum < 0 || math.IsNaN(float64(minimum)) || math.IsInf(float64(minimum), 0) {
		minimum = 0
		viewport.Invalid(true)
	}
	return typographyFlowContainer(c, viewport, func() {
		table := ui.Column(c).FillWidth().MinWidth(minimum).Gap(0).Border(1, t.Border).Role(ui.RoleTable).Label(p.Label).RowCount(len(p.Rows) + 1).ColumnCount(len(p.Columns)).FontSize(16).LineHeight(1.5).TextColor(t.Foreground)
		if len(p.Columns) == 0 {
			table.Border(0, ui.Transparent).RowCount(0)
			return
		}
		table.Children(func() {
			columnKeys := make([]typographyItemKey, len(p.Columns))
			columnOccurrences := map[string]int{}
			for i, column := range p.Columns {
				key := column.Key
				if key == "" {
					key = "index:" + strconv.Itoa(i)
				} else {
					key = "key:" + key
				}
				ordinal := columnOccurrences[key]
				columnOccurrences[key]++
				columnKeys[i] = typographyItemKey{4, key, ordinal}
			}
			buildRow := func(row int, key string, cells []string, header bool, ordinal int) {
				r := ui.Grid(c.Key(typographyItemKey{3, key, ordinal})).Columns(len(p.Columns)).FillWidth().Gap(0).Role(ui.RoleRow).RowIndex(row + 1)
				if row < len(p.Rows) {
					r.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
				}
				if !header && row%2 == 0 {
					r.Background(t.Muted)
				}
				r.Children(func() {
					for col, column := range p.Columns {
						cell := ui.Column(c.Key(columnKeys[col])).MinWidth(0).Padding(8, 16).Role(ui.RoleCell).ColumnIndex(col + 1).TextAlign(column.Align)
						if column.Align == 0 {
							cell.TextAlign(ui.Start)
						}
						if col < len(p.Columns)-1 {
							if c.Direction() == ui.RightToLeft {
								cell.BorderWidth(0, 0, 0, 1)
							} else {
								cell.BorderWidth(0, 1, 0, 0)
							}
							cell.BorderColor(t.Border)
						}
						if header {
							cell.Role(ui.RoleColumnHeader).FontWeight(700)
						} else if p.CellContent == nil && col < len(cells) {
							cell.Label(cells[col])
						}
						cell.Children(func() {
							if header {
								ui.Text(c, column.Title)
							} else if p.CellContent != nil {
								p.CellContent(c, row-1, col)
							} else if col < len(cells) {
								ui.Text(c, cells[col])
							}
						})
					}
				})
			}
			buildRow(0, "header", nil, true, 0)
			occurrences := map[string]int{}
			for i, row := range p.Rows {
				key := row.Key
				if key == "" {
					key = "index:" + strconv.Itoa(i)
				} else {
					key = "key:" + key
				}
				ordinal := occurrences[key]
				occurrences[key]++
				buildRow(i+1, key, row.Cells, false, ordinal)
			}
		})
	})
}
