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
	display.CardWith(c, display.CardProps{Label: "Workspace sidebar preview", Size: display.CardSmall, ContentBleed: true}, display.CardSlots{
		Title: "Workspace sidebar", Description: "Collapse to icons or resize the preview to use mobile navigation.",
		Content: func(child *ui.Context) {
			ui.Box(child).FillWidth().Height(340).Clip().Children(func() {
				_, err := navigation.SidebarWith(child, s.Popovers, "preview-workspace-sidebar", &s.sidebarRoute, &s.Sidebar,
					navigation.SidebarLayoutProps{
						Label: "Project workspace", ContentLabel: "Workspace content", Variant: navigation.SidebarInset, Collapse: navigation.SidebarIcon, Width: 208,
						Header: func(ctx *ui.Context) {
							label := "Acme Studio"
							if s.Sidebar.Collapsed && !s.Sidebar.IsMobile() {
								label = "AC"
							}
							ui.Text(ctx, label).FontWeight(600).SingleLine()
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
							{Key: "home", Label: "Home", Group: "Workspace"},
							{Key: "inbox", Label: "Inbox", Badge: "4"},
							{Key: "projects", Label: "Projects", DefaultOpen: true, Children: []navigation.SidebarItem{{Key: "website", Label: "Website"}, {Key: "mobile", Label: "Mobile app"}}},
							{Key: "settings", Label: "Settings", Group: "Account"},
						},
					}, func(ctx *ui.Context) {
						ui.Scroll(ctx).Fill().Padding(16).Gap(12).Children(func() {
							display.Typography(ctx, "Your workspace", display.H3)
							ui.Text(ctx, "Selected: "+s.sidebarRoute).FontSize(14)
							control.Field(ctx, control.FieldProps{Label: "Workspace owner"}, func() ui.Element {
								return control.Input(ctx, &s.Name, control.InputProps{})
							})
							ui.Text(ctx, "Press ⌘/Ctrl+B here to toggle navigation.").FontSize(12)
						})
					})
				if err != nil {
					s.popupError(child, err)
				}
			})
		},
	})
}
