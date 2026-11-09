package display

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
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

type DataColumn[T any] struct {
	ID, Title string
	Value     func(T) string
	Compare   func(T, T) int
	Width     float32
}

// DataTableState owns sorting, filtering and paging independently of rendering.
// Native list selection can be configured on List and follows stable row keys.
type DataTableState struct {
	Query          string
	Sort           ui.SortOrder
	Page, PageSize int
	List           ui.ListState
}
type DataTableProps[T any] struct {
	Columns []DataColumn[T]
	Rows    []T
	Key     func(T) any
	Search  func(T, string) bool
	Label   string
}

func DataRows[T any](state *DataTableState, p DataTableProps[T]) ([]T, int) {
	rows := make([]T, 0, len(p.Rows))
	query := strings.ToLower(strings.TrimSpace(state.Query))
	for _, row := range p.Rows {
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
		if match {
			rows = append(rows, row)
		}
	}
	for _, col := range p.Columns {
		if col.ID == state.Sort.Column {
			slices.SortStableFunc(rows, func(a, b T) int {
				if state.Sort.Descending {
					a, b = b, a
				}
				if col.Compare != nil {
					return col.Compare(a, b)
				}
				if col.Value != nil {
					return cmp.Compare(col.Value(a), col.Value(b))
				}
				return 0
			})
			break
		}
	}
	total := len(rows)
	size := state.PageSize
	if size <= 0 {
		size = 10
	}
	pages := max(1, (total+size-1)/size)
	state.Page = max(0, min(state.Page, pages-1))
	start := state.Page * size
	return rows[start:min(start+size, total)], total
}
func DataTable[T any](c *ui.Context, state *DataTableState, p DataTableProps[T]) ui.Element {
	t := theme.TokensOf(c)
	rows, total := DataRows(state, p)
	columns := make([]ui.TableColumn, len(p.Columns))
	for i, col := range p.Columns {
		columns[i] = ui.TableColumn{ID: col.ID, Title: col.Title, Width: col.Width, Sortable: true}
	}
	state.List.Sort = &state.Sort
	if p.Key != nil {
		state.List.Key = func(i int) any { return p.Key(rows[i]) }
	} else {
		state.List.Key = nil
	}
	return ui.Column(c).FillWidth().Gap(12).Children(func() {
		ui.Table(c, &state.List, columns, len(rows), func(row, col int) {
			if f := p.Columns[col].Value; f != nil {
				ui.Text(c, f(rows[row])).FontSize(14)
			}
		}).FillWidth().Height(float32(max(len(rows), 1)+1)*40).Label(p.Label).Border(1, t.Border).Radius(t.Radius)
		if total == 0 {
			Typography(c, "No results.", Muted)
		}
		ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("%d results · Page %d", total, state.Page+1)).Grow(1).FontSize(13).TextColor(t.MutedForeground)
			button := func(label string, disabled bool) bool {
				return ui.ButtonBase(c).Height(32).PaddingX(12).Radius(t.Radius*.8).Border(1, t.Border).Disabled(disabled).Children(func() { ui.Text(c, label) }).Clicked()
			}
			if button("Previous", state.Page == 0) {
				state.Page--
			}
			size := state.PageSize
			if size <= 0 {
				size = 10
			}
			if button("Next", (state.Page+1)*size >= total) {
				state.Page++
			}
		})
	})
}

type ChartKind int

const (
	BarChart ChartKind = iota
	LineChart
	AreaChart
)

type ChartSeries struct {
	Name   string
	Values []float64
	Color  ui.Color
}
type ChartProps struct {
	Labels []string
	Series []ChartSeries
	Kind   ChartKind
	Height float32
	Label  string
}

func Chart(c *ui.Context, p ChartProps) ui.Element {
	t := theme.TokensOf(c)
	height := p.Height
	if height <= 0 {
		height = 220
	}
	hi := float64(1)
	lo := float64(0)
	for _, s := range p.Series {
		for _, v := range s.Values {
			if !math.IsNaN(v) && !math.IsInf(v, 0) {
				hi = max(hi, v)
				lo = min(lo, v)
			}
		}
	}
	x := chart.NewBand(p.Labels...).Padding(.3, .15)
	y := chart.NewLinear(lo, hi).Nice(4)
	return chart.Plot(c, chart.Axes{Left: y.Ticks(4), Bottom: x.Ticks()}, func(area ui.Element) {
		area.Draw(func(painter *ui.Painter, r ui.Rect) {
			for _, tick := range y.Ticks(4) {
				painter.Line(r.X, chart.Y(r, tick.At), r.X+r.W, chart.Y(r, tick.At), 1, t.Border)
			}
		})
		for seriesIndex, s := range p.Series {
			color := s.Color
			if color == (ui.Color{}) {
				palette := t.Charts()
				color = palette[seriesIndex%len(palette)]
			}
			n := min(len(s.Values), len(p.Labels))
			for i := range n {
				if math.IsNaN(s.Values[i]) || math.IsInf(s.Values[i], 0) {
					continue
				}
				band := x.Band(i)
				if p.Kind == BarChart {
					width := (band.Hi - band.Lo) / float32(max(len(p.Series), 1))
					start := band.Lo + float32(seriesIndex)*width
					chart.Rect(c, chart.Span{Lo: start, Hi: start + width}, y.Span(0, s.Values[i])).Background(color).Radius(3).Label(fmt.Sprintf("%s %s: %g", s.Name, p.Labels[i], s.Values[i]))
				}
			}
			if p.Kind != BarChart {
				segments := finiteSegments(s.Values[:n], x, y)
				area.Draw(func(painter *ui.Painter, r ui.Rect) {
					for _, segment := range segments {
						if p.Kind == AreaChart {
							base := make([]float32, len(segment.X))
							for i := range base {
								base[i] = y.At(0)
							}
							painter.FillPath(chart.Area(r, segment.X, base, segment.Y, chart.CurveMonotone), color.Alpha(.15))
						}
						painter.StrokePath(chart.Line(r, segment.X, segment.Y, chart.CurveMonotone), 2, color)
					}
				})
			}
		}
	}).Height(height).FillWidth().Gap(8).TextColor(t.MutedForeground).FontSize(12).Label(p.Label)
}

type plotSegment struct{ X, Y []float32 }

// Non-finite measurements are missing observations, so a line must break at
// them rather than connect the observations on either side with invented data.
func finiteSegments(values []float64, x chart.Band, y chart.Linear) []plotSegment {
	var out []plotSegment
	var segment plotSegment
	flush := func() {
		if len(segment.X) > 0 {
			out = append(out, segment)
			segment = plotSegment{}
		}
	}
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			flush()
			continue
		}
		segment.X = append(segment.X, x.Band(i).Mid())
		segment.Y = append(segment.Y, y.At(v))
	}
	flush()
	return out
}
