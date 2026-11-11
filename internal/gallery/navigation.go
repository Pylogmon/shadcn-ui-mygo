package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func projectExportMenuItem() overlay.MenuItem {
	return overlay.MenuItem{Key: "export", Label: "Export", Children: []overlay.MenuItem{
		{Key: "go-source", Label: "Go source"},
		{Key: "preset-json", Label: "Preset JSON"},
	}}
}

func projectMenuItems() []overlay.MenuItem {
	return []overlay.MenuItem{
		{Key: "copy", Label: "Copy link"},
		{Key: "share", Label: "Share project"},
		projectExportMenuItem(),
		{Separator: true},
		{Key: "archive", Label: "Archive project", Destructive: true},
	}
}

func (s *State) projectContextMenu(c *ui.Context, target ui.Element, items []overlay.MenuItem) {
	items = localizedMenuItems(c, items)
	if s.Config.NativeMenus {
		chosen, err := desktop.NativeContextMenu(c, target, overlay.MenuProps{Items: items})
		if err != nil {
			s.popupError(c, err)
		}
		if chosen != "" {
			s.pendingMenuAction = chosen
		}
		return
	}
	const key = "navigation-project-context"
	if s.menuStates == nil {
		s.menuStates = make(map[string]*creatorMenu)
	}
	state := s.menuStates[key]
	if state == nil {
		state = &creatorMenu{}
		s.menuStates[key] = state
	}
	_, err := desktop.ContextMenu(c, s.Popovers, key, target, &state.Open, &state.Menu, desktop.MenuProps{
		Label: tr(c, "Project workspace actions"), Items: items,
		OnChoose: func(chosen string) { s.pendingMenuAction = chosen },
	})
	if err != nil {
		s.popupError(c, err)
	}
}

func menuActionLabel(key string) string {
	labels := map[string]string{
		"copy": "Copy link", "share": "Share project", "archive": "Archive project",
		"export": "Export", "go-source": "Go source", "preset-json": "Preset JSON",
		"new": "New project", "open": "Open project", "close": "Close project",
		"undo": "Undo", "redo": "Redo",
	}
	if label, ok := labels[key]; ok {
		return label
	}
	return key
}

func (s *State) projectMenuActionPreview(c *ui.Context) {
	if s.pendingMenuAction != "" {
		s.navigationMenuAction = menuActionLabel(s.pendingMenuAction)
	}
	if s.navigationMenuAction != "" {
		ui.Text(c, tr(c, "Last action: ")+tr(c, s.navigationMenuAction)).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
	}
}

func (s *State) cardHeader(c *ui.Context, title, description string) ui.Element {
	t := theme.TokensOf(c)
	return ui.Column(c).Gap(6).Children(func() {
		theme.StyleHeading(c, ui.Text(c, tr(c, title)).FontSize(18).FontWeight(600))
		if description != "" {
			ui.Text(c, tr(c, description)).FontSize(14).TextColor(t.MutedForeground)
		}
	})
}
func (s *State) navigationPreview(c *ui.Context) {
	s.heading(c, "Find your way.", "Tabs, routes and menus with native keyboard and focus behavior.")
	s.breadcrumbPreview(c)
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Tabs", "Arrow keys select a tab. The header action shares the selected view.")
			s.composedProjectTabs(c)
			display.Separator(c, false)
			s.cardHeader(c, "Manual activation", "Move focus with arrows; press Enter to select.")
			navigation.Tabs(c, &s.manualTab, navigation.TabsProps{Label: tr(c, "Manual tabs"), Manual: true, Items: []navigation.TabItem{{Label: tr(c, "Account")}, {Label: tr(c, "Security")}, {Label: tr(c, "Unavailable"), Disabled: true}}}, func(i int) {
				ui.Text(c, tr(c, []string{"Update your account information.", "Review your sign-in preferences.", ""}[i])).Padding(12)
			})
		})
	}, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Menu presentation", "Native menus follow the desktop; choose Shadcn Flyout to compare themed menus.")
			s.menubarPreview(c)
			menu := overlay.MenuProps{Items: projectMenuItems()}
			s.previewMenu(c, "navigation-project-actions", "Project actions", menu)
			tokens := theme.TokensOf(c)
			target := ui.Column(c).FillWidth().MinWidth(0).MinHeight(92).Padding(16).Gap(6).
				Border(1, tokens.Border).Radius(tokens.RadiusMD()).Background(tokens.Muted.Alpha(.3)).
				Focusable().Role(ui.RoleGroup).Label(tr(c, "Project workspace")).
				Description(tr(c, "Right-click or press Shift+F10 for project actions.")).Children(func() {
				ui.Text(c, tr(c, "Project workspace")).FontSize(14).FontWeight(500)
				ui.Text(c, tr(c, "Right-click or press Shift+F10 for project actions.")).FontSize(12).TextColor(tokens.MutedForeground)
			})
			s.projectContextMenu(c, target, menu.Items)
			s.projectMenuActionPreview(c)
		}).Label(tr(c, "Project menu preview"))
	})
	s.routePanelsPreview(c)
	s.componentPreview(c, "tabs", func() { s.tabsCompositions.view(c, s.Popovers, s.Config.NativeMenus) })
	s.componentPreview(c, "context-menu", func() { s.contextMenuCompositions.view(c, s.Popovers, s.Config.NativeMenus) })
	s.componentPreview(c, "menubar", func() { s.menubarCompositions.view(c, s.Popovers, s.Config.NativeMenus) })
	s.sidebarPreview(c)
	s.paginationPreview(c)
}
