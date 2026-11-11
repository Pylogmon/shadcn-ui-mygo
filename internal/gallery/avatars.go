package gallery

import (
	"errors"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

var avatarPortrait = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 80 80"><path fill="#bfdbfe" d="M0 0h80v80H0z"/><circle fill="#f5d0a9" cx="40" cy="31" r="16"/><path fill="#1e3a8a" d="M10 80c0-20 12-30 30-30s30 10 30 30"/><path fill="#713f12" d="M24 30c-5-20 28-27 32 0-10-2-17-7-20-12-2 8-7 12-12 12"/></svg>`))
var avatarCheck = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m3 8 3 3 7-7"/></svg>`))

func (s *State) avatarsPreview(c *ui.Context) {
	owner := ui.Column(c).FillWidth()
	mode := ui.Local(owner, "avatar-image-mode", func() int { return 0 })
	owner.Children(func() {
		display.CardWith(c, display.CardProps{Label: tr(c, "Profile portraits"), Size: display.CardSmall}, display.CardSlots{
			Title: tr(c, "Avatars"), Description: tr(c, "Portraits, initials, status and a few more people."),
			Content: func(child *ui.Context) {
				ui.Row(child).Gap(16).AlignItems(ui.Center).Children(func() {
					for _, scale := range []display.AvatarSize{display.AvatarSmall, display.AvatarDefault, display.AvatarLarge} {
						p := display.AvatarProps{Key: string(rune('a' + scale)), Image: avatarPortrait, Fallback: "SD", Label: "Sofia Davis", Scale: scale,
							Badge: &display.AvatarBadgeProps{Label: tr(child, "Online"), Icon: avatarCheck}}
						switch *mode {
						case 1:
							p.Image = nil
						case 2:
							p.Image, p.Loading = nil, true
						case 3:
							p.Image, p.Error = nil, errors.New("preview image unavailable")
						}
						display.Avatar(child, p)
					}
				})
				ui.Row(child).Wrap().Gap(6).Children(func() {
					for i, label := range []string{"Image", "Initials", "Loading", "Unavailable"} {
						selected := *mode == i
						if control.Toggle(child.Key(i), &selected, tr(child, label), control.ToggleProps{}).Height(28).PaddingX(8).FontSize(12).Changed() {
							*mode = i
							child.Invalidate()
						}
					}
				})
				ui.Text(child, []string{"Image ready", "Showing initials", "Loading image…", "Image unavailable — showing initials"}[*mode]).FontSize(12).Role(ui.RoleStatus)
			},
			Footer: func(child *ui.Context) {
				display.AvatarGroup(child, display.AvatarGroupProps{Label: tr(child, "Project team"), Max: 3, Total: 6, Items: []display.AvatarProps{
					{Key: "sofia", Label: "Sofia Davis", Image: avatarPortrait},
					{Key: "jackson", Label: "Jackson Lee", Fallback: "JL"},
					{Key: "olivia", Label: "Olivia Martin", Fallback: "OM"},
				}})
			},
		})
	})
}
