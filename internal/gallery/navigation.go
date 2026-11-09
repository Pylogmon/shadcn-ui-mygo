package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func (s *State) cardHeader(c *ui.Context, title, description string) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, title).Font(s.Config.HeadingFamily()).FontSize(18).FontWeight(600)
		if description != "" {
			ui.Text(c, description).Font(s.Config.BodyFamily()).FontSize(14).TextColor(t.MutedForeground)
		}
	})
}
func (s *State) navigationPreview(c *ui.Context) {
	s.heading(c, "Find your way.", "Tabs, routes and menus with native keyboard and focus behavior.")
	navigation.Breadcrumb(c, []navigation.BreadcrumbItem{{Key: "home", Label: "Home"}, {Key: "components", Label: "Components"}, {Key: "navigation", Label: "Navigation"}})
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Tabs", "Arrow keys select a tab. Each trigger controls its panel.")
			navigation.Tabs(c, &s.lineTab, navigation.TabsProps{Label: "Project view", Variant: navigation.TabsLine, Items: []navigation.TabItem{{Label: "Overview"}, {Label: "Activity"}, {Label: "Settings"}}}, func(i int) {
				ui.Text(c, []string{"Your project at a glance.", "Your latest project updates.", "Manage your project preferences."}[i]).Padding(12)
			})
			display.Separator(c, false)
			s.cardHeader(c, "Manual activation", "Move focus with arrows; press Enter to select.")
			navigation.Tabs(c, &s.manualTab, navigation.TabsProps{Label: "Manual tabs", Manual: true, Items: []navigation.TabItem{{Label: "Account"}, {Label: "Security"}, {Label: "Unavailable", Disabled: true}}}, func(i int) {
				ui.Text(c, []string{"Update your account information.", "Review your sign-in preferences.", ""}[i]).Padding(12)
			})
		})
	}, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Menu presentation", "Native menus follow the desktop; choose Shadcn Flyout to compare themed menus.")
			menus := []overlay.MenubarItem{{Label: "File", Items: []overlay.MenuItem{{Key: "new", Label: "New project"}, {Key: "open", Label: "Open project"}, {Separator: true}, {Key: "close", Label: "Close project"}}}, {Label: "Edit", Items: []overlay.MenuItem{{Key: "undo", Label: "Undo"}, {Key: "redo", Label: "Redo"}}}}
			if s.Config.NativeMenus {
				if chosen := overlay.NativeMenubar(c, menus); chosen != "" {
					s.pendingMenuAction = chosen
				}
			} else {
				ui.Row(c).Gap(4).FocusGroup(ui.Horizontal).Children(func() {
					for _, menu := range menus {
						s.previewMenu(c, "navigation-menubar:"+menu.Label, menu.Label, overlay.MenuProps{Items: menu.Items})
					}
				})
			}
			menu := overlay.MenuProps{Items: []overlay.MenuItem{{Key: "copy", Label: "Copy link"}, {Key: "share", Label: "Share project"}, {Separator: true}, {Key: "archive", Label: "Archive project", Destructive: true}}}
			s.previewMenu(c, "navigation-project-actions", "Project actions", menu)
		})
	})
	display.Card(c, func() {
		s.cardHeader(c, "Pagination", "A single controlled page value, with bounds and current-page semantics.")
		navigation.Pagination(c, &s.PageNumber, navigation.PaginationProps{Pages: 20})
	})
}
