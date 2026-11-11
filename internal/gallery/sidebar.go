package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
)

func (s *State) sidebarPreview(c *ui.Context) {
	if s.sidebarRoute == "" {
		s.sidebarRoute = "home"
	}
	display.CardWith(c, display.CardProps{Label: tr(c, "Workspace sidebar preview"), Size: display.CardSmall, ContentBleed: true}, display.CardSlots{
		Title: tr(c, "Workspace sidebar"), Description: tr(c, "Collapse to icons or resize the preview to use mobile navigation."),
		Content: func(child *ui.Context) {
			ui.Box(child).FillWidth().Height(340).Clip().Children(func() {
				_, err := navigation.SidebarWith(child, s.Popovers, "preview-workspace-sidebar", &s.sidebarRoute, &s.Sidebar,
					navigation.SidebarLayoutProps{
						Label: tr(child, "Project workspace"), ContentLabel: "Workspace content", Variant: navigation.SidebarInset, Collapse: navigation.SidebarIcon, Width: 208,
						Header: func(ctx *ui.Context) {
							label := "Acme Studio"
							if s.Sidebar.Collapsed && !s.Sidebar.IsMobile() {
								label = "AC"
							}
							ui.Text(ctx, tr(ctx, label)).FontWeight(600).SingleLine()
						},
						Footer: func(ctx *ui.Context) {
							ui.Row(ctx).Gap(8).AlignItems(ui.Center).Children(func() {
								display.Avatar(ctx, display.AvatarProps{Label: "Sofia Davis", Fallback: "SD", Scale: display.AvatarSmall})
								if !s.Sidebar.Collapsed || s.Sidebar.IsMobile() {
									ui.Text(ctx, "Sofia Davis").FontSize(12).SingleLine()
								}
							})
						},
						Items: []navigation.SidebarItem{
							{Key: "home", Label: tr(child, "Home"), Group: tr(child, "Workspace")},
							{Key: "inbox", Label: tr(child, "Inbox"), Badge: "4"},
							{Key: "projects", Label: tr(child, "Projects"), DefaultOpen: true, Children: []navigation.SidebarItem{{Key: "website", Label: tr(child, "Website")}, {Key: "mobile", Label: tr(child, "Mobile app")}}},
							{Key: "settings", Label: tr(child, "Settings"), Group: tr(child, "Account")},
						},
					}, func(ctx *ui.Context) {
						ui.Scroll(ctx).Fill().Padding(16).Gap(12).Children(func() {
							display.Typography(ctx, tr(ctx, "Your workspace"), display.H3)
							ui.Text(ctx, tr(ctx, "Selected: ")+s.sidebarRoute).FontSize(14)
							control.Field(ctx, control.FieldProps{Label: tr(ctx, "Workspace owner")}, func() ui.Element {
								return control.Input(ctx, &s.Name, control.InputProps{})
							})
							ui.Text(ctx, tr(ctx, "Press ⌘/Ctrl+B here to toggle navigation.")).FontSize(12)
						})
					})
				if err != nil {
					s.popupError(child, err)
				}
			})
		},
	})
}
