package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

var contentFile = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8zM14 2v6h6M8 13h8M8 17h5"/></svg>`))
var contentInbox = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m4 4-3 12v4a1 1 0 0 0 1 1h20a1 1 0 0 0 1-1v-4L20 4zM1 16h6l2 3h6l2-3h6"/></svg>`))
var contentWarning = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 3 2 21h20zM12 9v5M12 18h.01"/></svg>`))
var contentClose = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="m6 6 12 12M18 6 6 18"/></svg>`))
var contentExternal = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6M21 3l-9 9M10 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-5"/></svg>`))

func (s *State) contentHeader(title, description string) func(*ui.Context) {
	return func(c *ui.Context) {
		display.CardTitle(c, tr(c, title), display.CardSmall)
		display.CardDescription(c, tr(c, description))
	}
}

func (s *State) contents(c *ui.Context) {
	s.heading(c, "Content, with context.", "Useful status, clear actions and room for your next idea.")
	columns(c, func() {
		display.CardWith(c, display.CardProps{Size: display.CardSmall}, display.CardSlots{
			Header: s.contentHeader("Badges", "Small details for status and navigation."),
			Content: func(child *ui.Context) {
				ui.Row(child).Wrap().Gap(8).Children(func() {
					for i, label := range []string{"Default", "Secondary", "Outline", "Soft", "Ghost", "Link"} {
						display.Badge(child.Key(i), tr(child, label), display.BadgeVariant(i))
					}
				})
				ui.Row(child).Wrap().Gap(8).Children(func() {
					display.BadgeWith(child, tr(child, "Remove filter"), display.BadgeProps{Variant: display.BadgeOutline, Button: true, EndIcon: contentClose, OnActivate: func() { s.contentAction = "Filter removed" }})
					display.BadgeWith(child, tr(child, "Badge documentation"), display.BadgeProps{Variant: display.BadgeLink, URL: "https://ui.shadcn.com/docs/components/badge", EndIcon: contentExternal})
				})
			},
		})
		display.CardWith(c, display.CardProps{Size: display.CardSmall}, display.CardSlots{
			Header: s.contentHeader("Notices", "Context, recovery and a way to continue."),
			Content: func(child *ui.Context) {
				display.AlertWith(child, display.AlertProps{Title: tr(child, "Your changes are saved"), Description: tr(child, "Your workspace is ready for the next idea."), Icon: avatarCheck}, display.AlertSlots{
					Action: func(ctx *ui.Context) {
						if control.Button(ctx, tr(ctx, "Undo"), control.ButtonProps{Variant: control.Ghost, Size: control.ExtraSmall}).Clicked() {
							s.contentAction = "Changes restored"
						}
					},
				})
				display.AlertWith(child, display.AlertProps{Title: tr(child, "Unable to upload file"), Description: tr(child, "Check the connection and try again."), Destructive: true, Icon: contentWarning}, display.AlertSlots{
					Action: func(ctx *ui.Context) {
						if control.Button(ctx, tr(ctx, "Retry"), control.ButtonProps{Variant: control.Ghost, Size: control.ExtraSmall}).Clicked() {
							s.contentAction = "Upload queued"
						}
					},
				})
				if !s.contentDismissed {
					display.AlertWith(child, display.AlertProps{Title: tr(child, "A new update is available"), Description: tr(child, "You can install it when you are ready.")}, display.AlertSlots{
						Action: func(ctx *ui.Context) {
							if control.Button(ctx, "", control.ButtonProps{Label: tr(ctx, "Dismiss update"), Variant: control.Ghost, Size: control.IconExtraSmall, Icon: contentClose}).Clicked() {
								s.contentDismissed = true
							}
						},
					})
				} else if control.Button(child, tr(child, "Restore notice"), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
					s.contentDismissed = false
				}
			},
		})
	}, func() {
		display.CardWith(c, display.CardProps{Size: display.CardSmall}, display.CardSlots{
			Header: s.contentHeader("Project files", "Open a file or choose its own action."),
			Content: func(child *ui.Context) {
				display.ItemGroup(child, display.ItemGroupProps{Label: tr(child, "Workspace files"), Items: []display.ItemProps{
					{Key: "report", Title: tr(child, "Quarterly report"), Description: tr(child, "PDF · updated today"), Variant: display.ItemOutline, Size: display.ItemSmall,
						Media:   &display.ItemMediaProps{Variant: display.ItemMediaIcon, Icon: contentFile},
						Primary: &display.ItemPrimary{Label: tr(child, "Open quarterly report"), OnActivate: func() { s.contentAction = "Opened quarterly report" }},
						Actions: func(ctx *ui.Context) {
							label := "Pin report"
							if s.contentPinned {
								label = "Unpin report"
							}
							if control.Button(ctx, tr(ctx, label), control.ButtonProps{Variant: control.Ghost, Size: control.ExtraSmall}).Clicked() {
								s.contentPinned = !s.contentPinned
							}
						},
					},
					{Key: "resources", Title: tr(child, "Design resources"), Description: tr(child, "Shared with your team"), Variant: display.ItemMuted, Size: display.ItemExtraSmall,
						Media:   &display.ItemMediaProps{Variant: display.ItemMediaImage, Image: avatarPortrait, Label: tr(child, "Resource artwork")},
						Primary: &display.ItemPrimary{Kind: display.ItemPrimaryLink, Label: tr(child, "Open component documentation"), URL: "https://ui.shadcn.com/docs/components/item"},
						Actions: func(ctx *ui.Context) { display.Badge(ctx, tr(ctx, "Shared"), display.BadgeSecondary) },
					},
					{Key: "archive", Title: tr(child, "Archived workspace"), Description: tr(child, "This workspace is currently unavailable."), Size: display.ItemExtraSmall, Disabled: true,
						Primary: &display.ItemPrimary{Label: tr(child, "Open archived workspace")}},
				}})
			},
		})
		display.CardWith(c, display.CardProps{Size: display.CardSmall}, display.CardSlots{
			Header: s.contentHeader("Empty states", "A helpful starting point for something new."),
			Content: func(child *ui.Context) {
				padding := float32(20)
				display.EmptyWith(child, display.EmptyProps{Title: tr(child, "No projects yet"), Description: tr(child, "Create a project to get started."), Icon: contentInbox, MediaVariant: display.EmptyMediaIcon, Bordered: true, Padding: &padding}, display.EmptySlots{
					Content: func(ctx *ui.Context) {
						control.Input(ctx, &s.contentProject, control.InputProps{Label: tr(ctx, "New project name"), Placeholder: tr(ctx, "Project name")}).WidthPercent(75).TextAlign(ui.Start)
						if control.Button(ctx, tr(ctx, "Create project"), control.ButtonProps{Size: control.Small}).Clicked() {
							if s.contentProject == "" {
								s.contentAction = "Enter a project name to continue"
							} else {
								s.contentAction = "Created project: " + s.contentProject
							}
						}
					},
				})
			},
		})
	})
	alert, empty := s.statusCompositions.view(c, s.Popovers, s.Config.NativeMenus)
	s.revealComponent("alert", alert)
	s.revealComponent("empty", empty)
	s.componentPreview(c, "item", func() { s.itemCompositions.view(c, s.Popovers, s.Config.NativeMenus) })
	s.componentPreview(c, "badge", func() { s.badgeCompositions.view(c, s.Popovers, s.Config.NativeMenus) })
	if s.contentAction != "" {
		ui.Text(c, s.contentAction).FontSize(14).Role(ui.RoleStatus)
	}
}
