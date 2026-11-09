// Package navigation provides native shadcn navigation components.
package navigation

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type BreadcrumbItem struct {
	Key, Label string
	Disabled   bool
}

// Breadcrumb returns the Key of an activated ancestor; the last crumb is the
// current page, exposed to assistive technology as such rather than a button.
func Breadcrumb(c *ui.Context, items []BreadcrumbItem) string {
	t := theme.TokensOf(c)
	chosen := ""
	ui.Row(c).Gap(8).AlignItems(ui.Center).Role(ui.RoleNavigation).Label("Breadcrumb").Children(func() {
		for i, item := range items {
			if i > 0 {
				ui.Text(c, "›").FontSize(14).TextColor(t.MutedForeground).AccessHidden(true)
			}
			if i == len(items)-1 {
				ui.Text(c, item.Label).FontSize(14).TextColor(t.Foreground).Current(ui.CurrentPage)
				continue
			}
			b := navButton(c, ui.ButtonBase(c.Key(i)).Disabled(item.Disabled), item.Label).Height(28).PaddingX(0).TextColor(t.MutedForeground)
			if b.Hovered() {
				b.TextColor(t.Foreground)
			}
			if b.Clicked() {
				chosen = item.Key
			}
		}
	})
	return chosen
}

func navButton(c *ui.Context, b ui.Element, label string) ui.Element {
	t := theme.TokensOf(c)
	b.Height(36).PaddingX(12).Radius(t.Radius).Gap(8).FontSize(14).FontWeight(500).TextColor(t.Foreground).Transition(motion.Transition(c))
	if b.Hovered() {
		b.Background(t.Accent).TextColor(t.AccentForeground)
	}
	b.Children(func() { ui.Text(c, label).SingleLine() })
	return b
}

type PaginationProps struct {
	Pages    int
	Siblings int
}

// Pagination uses one-based page numbers and bounds both previous/next and
// numbered links. It reports a change in the frame a page is activated.
func Pagination(c *ui.Context, page *int, p PaginationProps) bool {
	if p.Pages <= 0 {
		return false
	}
	*page = max(1, min(*page, p.Pages))
	before := *page
	t := theme.TokensOf(c)
	ui.Row(c).Gap(4).AlignItems(ui.Center).Role(ui.RoleNavigation).Label("Pagination").Children(func() {
		if navButton(c, ui.ButtonBase(c.Key("previous")).Disabled(*page <= 1), "‹ Previous").Clicked() {
			*page--
		}
		siblings := p.Siblings
		if siblings <= 0 {
			siblings = 1
		}
		last := 0
		for n := 1; n <= p.Pages; n++ {
			if n != 1 && n != p.Pages && (n < *page-siblings || n > *page+siblings) {
				continue
			}
			if last > 0 && n-last > 1 {
				ui.Text(c, "…").Width(36).TextAlign(ui.Center).TextColor(t.MutedForeground).AccessHidden(true)
			}
			b := navButton(c, ui.ButtonBase(c.Key(n)).Label("Page "+strconv.Itoa(n)), strconv.Itoa(n)).Width(36).PaddingX(0)
			if n == *page {
				b.Border(1, t.Border).Current(ui.CurrentPage)
			}
			if b.Clicked() {
				*page = n
			}
			last = n
		}
		if navButton(c, ui.ButtonBase(c.Key("next")).Disabled(*page >= p.Pages), "Next ›").Clicked() {
			*page++
		}
	})
	return before != *page
}

// NavItem is a route entry. Key is returned on activation; applications decide
// whether that changes an in-app route or opens an external URL.
type NavItem struct {
	Key, Label, Description, Group string
	Icon                           *ui.SVG
	Disabled                       bool
	Children                       []NavItem
}
