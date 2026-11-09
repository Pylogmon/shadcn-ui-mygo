package navigation

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type PaginationProps struct {
	Pages, Siblings int
	// Explicit counts permit zero. Nil SiblingCount retains legacy Siblings
	// (zero means one); nil BoundaryCount shows the first and last page.
	SiblingCount, BoundaryCount                         *int
	Label, PreviousLabel, NextLabel                     string
	PreviousText, NextText                              string
	PageLabel                                           func(int) string
	Compact, AlwaysShowText, Disabled, HidePreviousNext bool
	// URL supplies actual native Link targets instead of buttons. PageContent
	// builds non-interactive visuals; the native button/link owns activation.
	URL          func(int) string
	PageContent  func(*ui.Context, int, bool)
	OnPageChange func(int)
	// MaxVisible bounds generated numeric controls, default 4096. An oversized
	// requested window is Invalid/inert rather than silently truncated.
	MaxVisible int
}

type paginationKey struct {
	kind uint8
	page int
}
type paginationRange struct{ first, last int }

// paginationNumbers touches only requested windows, not every page. Arithmetic
// uses distances from bounds so counts near the largest int cannot overflow.
func paginationNumbers(pages, current, siblings, boundaries, budget int) ([]int, bool) {
	if pages <= 0 {
		return nil, true
	}
	siblings, boundaries = max(0, siblings), max(0, boundaries)
	left := current - min(current-1, siblings)
	right := current + min(pages-current, siblings)
	ranges := []paginationRange{}
	if boundaries > 0 {
		ranges = append(ranges, paginationRange{1, min(pages, boundaries)})
	}
	ranges = append(ranges, paginationRange{left, right})
	if boundaries > 0 {
		ranges = append(ranges, paginationRange{pages - min(pages-1, boundaries-1), pages})
	}
	// Three already mostly sorted ranges: a small insertion sort avoids work
	// proportional to the total number of pages.
	for i := 1; i < len(ranges); i++ {
		for j := i; j > 0 && ranges[j].first < ranges[j-1].first; j-- {
			ranges[j], ranges[j-1] = ranges[j-1], ranges[j]
		}
	}
	merged := ranges[:0]
	for _, r := range ranges {
		if len(merged) > 0 {
			last := &merged[len(merged)-1]
			if r.first <= last.last || last.last < pages && r.first == last.last+1 {
				last.last = max(last.last, r.last)
				continue
			}
		}
		merged = append(merged, r)
	}
	count := 0
	for _, r := range merged {
		n := r.last - r.first
		if n >= budget-count {
			return nil, false
		}
		count += n + 1
	}
	numbers := make([]int, 0, count)
	for _, r := range merged {
		for n := r.first; ; n++ {
			numbers = append(numbers, n)
			if n == r.last {
				break
			}
		}
	}
	return numbers, true
}

// Pagination preserves the one-based legacy entry point and reports only a
// user's page change. PaginationWith returns the composable native container.
func Pagination(c *ui.Context, page *int, p PaginationProps) bool {
	if p.PreviousLabel == "" {
		p.PreviousLabel = "‹ Previous"
	}
	if p.NextLabel == "" {
		p.NextLabel = "Next ›"
	}
	return PaginationWith(c, page, p).Changed()
}

func PaginationWith(c *ui.Context, page *int, p PaginationProps) ui.Element {
	label := p.Label
	if label == "" {
		label = "Pagination"
	}
	root := ui.Row(c).FillWidth().MinWidth(0).Justify(ui.Center).Role(ui.RoleNavigation).Label(label).Disabled(p.Disabled)
	if page == nil || p.Pages < 0 {
		return root.Invalid(true)
	}
	if p.Pages == 0 {
		return root
	}
	*page = max(1, min(*page, p.Pages))
	before := *page
	siblings := p.Siblings
	if siblings <= 0 {
		siblings = 1
	}
	if p.SiblingCount != nil {
		siblings = *p.SiblingCount
	}
	boundaries := 1
	if p.BoundaryCount != nil {
		boundaries = *p.BoundaryCount
	}
	budget := p.MaxVisible
	if budget <= 0 {
		budget = 4096
	}
	numbers, valid := paginationNumbers(p.Pages, *page, siblings, boundaries, budget)
	if siblings < 0 || boundaries < 0 || !valid {
		return root.Invalid(true).Disabled(true)
	}
	width, _ := c.Size()
	compact := p.Compact || !p.AlwaysShowText && width < 640
	t := theme.TokensOf(c)
	choose := func(next int) {
		if !root.IsDisabled() {
			*page = next
		}
	}
	root.Children(func() {
		ui.Row(c).Gap(4).AlignItems(ui.Center).Role(ui.RoleList).Children(func() {
			button := func(kind uint8, n int, text, accessible string, disabled bool) {
				active := kind == 0 && n == before
				// Previous/next identify controls, not their changing destination.
				// A stable native node preserves focus for repeated keyboard use.
				keyPage := n
				if kind != 0 {
					keyPage = 0
				}
				ui.Box(c).Key(paginationKey{kind, keyPage}).Role(ui.RoleListItem).Children(func() {
					url := ""
					if p.URL != nil {
						url = p.URL(n)
					}
					var outer, target ui.Element
					if url == "" {
						outer = ui.ButtonBase(c)
						target = outer
					} else {
						outer = ui.Row(c).Role(ui.RoleNone)
					}
					motion.ApplyTransition(c, outer.Height(36).MinWidth(36).PaddingX(10).Gap(6).AlignItems(ui.Center).Justify(ui.Center).Radius(t.RadiusMD()).Border(1, ui.Transparent).FontSize(14).FontWeight(500).TextColor(t.Foreground))
					// Keep the source's icon size for ordinary page numbers. Long
					// native labels need their content width to remain distinguishable.
					if kind == 0 && len(text) <= 3 || kind != 0 && compact {
						outer.Width(36).PaddingX(0)
					}
					outer.Children(func() {
						if url != "" {
							target = ui.Link(c, "", url).Absolute().Left(-1).Top(-1).Right(-1).Bottom(-1).Background(ui.Transparent).TextColor(ui.Transparent)
						}
						ui.Row(c).Gap(6).AlignItems(ui.Center).IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(func() {
							if kind == 0 && p.PageContent != nil {
								p.PageContent(c, n, active)
								return
							}
							arrowLeft, arrowRight := "‹", "›"
							if c.Direction() == ui.RightToLeft {
								arrowLeft, arrowRight = arrowRight, arrowLeft
							}
							if kind == 1 {
								ui.Text(c, arrowLeft).FontSize(16)
								if !compact {
									ui.Text(c, text).SingleLine()
								}
							} else if kind == 2 {
								if !compact {
									ui.Text(c, text).SingleLine()
								}
								ui.Text(c, arrowRight).FontSize(16)
							} else {
								ui.Text(c, text).SingleLine()
							}
						})
					})
					target.Label(accessible).Disabled(disabled || p.Disabled).FocusRing(false).Cursor(ui.CursorPointer)
					if active {
						target.Current(ui.CurrentPage)
						outer.Border(1, t.Border).Background(t.Background)
					}
					if !target.IsDisabled() && (target.Hovered() || target.Pressed()) {
						outer.Background(t.Muted)
					}
					if target.FocusVisible() {
						outer.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
					}
					if target.IsDisabled() {
						outer.Opacity(.5)
					}
					if target.Clicked() && !target.IsDisabled() {
						choose(n)
					}
				})
			}
			previous, previousLabel := p.PreviousText, p.PreviousLabel
			if previous == "" {
				previous = "Previous"
			}
			if previousLabel == "" {
				previousLabel = "Go to previous page"
			}
			next, nextLabel := p.NextText, p.NextLabel
			if next == "" {
				next = "Next"
			}
			if nextLabel == "" {
				nextLabel = "Go to next page"
			}
			if !p.HidePreviousNext {
				button(1, max(1, before-1), previous, previousLabel, before <= 1)
			}
			last := 0
			for _, n := range numbers {
				if last > 0 && n-last > 1 {
					ui.Box(c).Key(paginationKey{3, n}).Size(36, 36).Center().AccessHidden(true).Children(func() { ui.Text(c, "…").FontSize(16) })
				}
				pageLabel := "Page " + strconv.Itoa(n)
				if p.PageLabel != nil {
					pageLabel = p.PageLabel(n)
				}
				button(0, n, strconv.Itoa(n), pageLabel, false)
				last = n
			}
			if !p.HidePreviousNext {
				target := before
				if target < p.Pages {
					target++
				}
				button(2, target, next, nextLabel, before >= p.Pages)
			}
		})
	})
	if before != *page {
		root.SetChanged()
		if p.OnPageChange != nil {
			p.OnPageChange(*page)
		}
		c.Invalidate()
	}
	return root
}
