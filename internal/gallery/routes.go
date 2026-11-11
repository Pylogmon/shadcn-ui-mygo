package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
)

func (s *State) breadcrumbPreview(c *ui.Context) {
	items := []navigation.BreadcrumbItem{
		{Key: "home", Label: "shadcn/ui", URL: "https://ui.shadcn.com/create"},
		{Key: "components", Label: tr(c, "Components"), URL: "https://ui.shadcn.com/docs/components"},
		{Key: "forms", Label: tr(c, "Forms")},
		{Key: "patterns", Label: tr(c, "Patterns")},
		{Key: "navigation", Label: tr(c, "Navigation")},
	}
	props := navigation.BreadcrumbProps{MaxItems: 3, EllipsisLabel: tr(c, "Show parent routes"), OnSelect: func(item navigation.BreadcrumbItem) {
		s.breadcrumbSelection = "Selected parent: " + item.Label
	}}
	var err error
	if s.Config.NativeMenus {
		_, err = navigation.NativeBreadcrumbWith(c, "gallery-routes", items, props)
	} else {
		_, err = navigation.BreadcrumbWith(c, s.Popovers, "gallery-routes", items, props)
	}
	if err != nil {
		s.popupError(c, err)
	}
	if s.breadcrumbSelection != "" {
		ui.Text(c, s.breadcrumbSelection).Role(ui.RoleStatus).FontSize(14)
	}
}

func (s *State) paginationPreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "Pagination", "Explore a million pages with a bounded navigation window.")
		navigation.PaginationWith(c, &s.PageNumber, navigation.PaginationProps{Pages: 1_000_000, Label: tr(c, "Preview pages")})
		ui.Text(c, fmt.Sprintf(tr(c, "Page %d of 1,000,000"), s.PageNumber)).Role(ui.RoleStatus).FontSize(14)
	})
}
