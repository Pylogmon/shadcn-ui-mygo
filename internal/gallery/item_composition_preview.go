package gallery

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Persistent state is application data and the public menu model only. Every
// rich item/menu slot receives its current window Context.
type itemCompositionPreviewState struct {
	ready, native, menuOpen                              bool
	recipe, language, action, popupError, selectedPerson string
	menu                                                 desktop.MenuState
}

var itemCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Basic and verified items"}, {Value: "variant", Label: "Three item variants"},
	{Value: "size", Label: "Three item sizes"}, {Value: "icon", Label: "Security icon"},
	{Value: "avatar", Label: "User and team avatars"}, {Value: "image", Label: "Music with cover images"},
	{Value: "group", Label: "People group"}, {Value: "header", Label: "Model image headers"},
	{Value: "link", Label: "Native item links"}, {Value: "dropdown", Label: "People dropdown"}, {Value: "rtl", Label: "Translated RTL items"},
	{Value: "create-default", Label: "Create: Default"}, {Value: "create-outline", Label: "Create: Outline"}, {Value: "create-muted", Label: "Create: Muted"},
	{Value: "create-default-small", Label: "Create: Default small"}, {Value: "create-outline-small", Label: "Create: Outline small"}, {Value: "create-muted-small", Label: "Create: Muted small"},
	{Value: "create-default-xs", Label: "Create: Default extra small"}, {Value: "create-outline-xs", Label: "Create: Outline extra small"}, {Value: "create-muted-xs", Label: "Create: Muted extra small"},
	{Value: "create-default-links", Label: "Create: Default links"}, {Value: "create-outline-links", Label: "Create: Outline links"}, {Value: "create-muted-links", Label: "Create: Muted links"},
	{Value: "create-default-group", Label: "Create: Default group"}, {Value: "create-outline-group", Label: "Create: Outline group"}, {Value: "create-muted-group", Label: "Create: Muted group"},
	{Value: "create-separator", Label: "Create: Separators"}, {Value: "create-header", Label: "Create: Headers"}, {Value: "create-footer", Label: "Create: Footers"}, {Value: "create-both", Label: "Create: Headers and footers"},
	{Value: "create-default-image", Label: "Create: Default images"}, {Value: "create-outline-image", Label: "Create: Outline images"}, {Value: "create-outline-image-small", Label: "Create: Outline images small"}, {Value: "create-outline-image-xs", Label: "Create: Outline images extra small"}, {Value: "create-muted-image", Label: "Create: Muted images"},
}

func (p *itemCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
}

func (p *itemCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
		return
	}
	if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *itemCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		p.native, p.popupError, p.menuOpen = native, "", false
		p.menu.Reset()
		if h != nil {
			for _, key := range []string{"item-recipe", "item-language", "item-people"} {
				h.Close(key)
			}
		}
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Item compositions"}, display.CardSlots{
		Title: "Item compositions", Description: "Independent actions, real links, images and source collections.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "item-recipe", &p.recipe, control.SelectProps{Label: "Item configuration", Options: itemCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "item-language", &p.language, control.SelectProps{Label: "Item language", Options: statusLanguages()})
			}
			ui.Column(child.Key("item-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() {
				build := func() { p.draw(child, h, native) }
				if p.recipe == "rtl" {
					ui.Column(child).FillWidth().MinWidth(0).Direction(childDirection(p.language)).Children(build)
				} else {
					build()
				}
			})
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
}

func itemRecipeShape(recipe string) (display.ItemVariant, display.ItemSize) {
	v, s := display.ItemDefault, display.ItemDefaultSize
	if strings.Contains(recipe, "outline") {
		v = display.ItemOutline
	} else if strings.Contains(recipe, "muted") {
		v = display.ItemMuted
	}
	if strings.HasSuffix(recipe, "-small") {
		s = display.ItemSmall
	} else if strings.HasSuffix(recipe, "-xs") {
		s = display.ItemExtraSmall
	}
	return v, s
}

func (p *itemCompositionPreviewState) actionButton(c *ui.Context, text, label, result string, variant control.Variant, size control.Size, icon *ui.SVG) ui.Element {
	e := control.Button(c, text, control.ButtonProps{Label: label, Variant: variant, Size: size, Icon: icon})
	if e.Clicked() {
		p.action = "Local item action: " + result
	}
	return e
}
func (p *itemCompositionPreviewState) primary(label string) *display.ItemPrimary {
	return &display.ItemPrimary{Kind: display.ItemPrimaryLink, Label: label, OnActivate: func() { p.action = "Local item link: " + label }}
}
func itemInbox() *display.ItemMediaProps {
	return &display.ItemMediaProps{Variant: display.ItemMediaIcon, Icon: outlined.Inbox(), Decorative: true}
}
func itemLinkContent(title, description string, icon *ui.SVG) func(*ui.Context) {
	return func(c *ui.Context) {
		ui.Row(c).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Basis(0).Grow(1).MinWidth(0).Gap(4).Children(func() {
				display.ItemTitle(c, title)
				if description != "" {
					display.ItemDescription(c, description)
				}
			})
			material.Icon(c, icon, material.Props{Size: 16}).Shrink(0)
		})
	}
}

func (p *itemCompositionPreviewState) group(c *ui.Context, label string, items []display.ItemProps, gap float32) ui.Element {
	return display.ItemGroup(c, display.ItemGroupProps{Label: label, Items: items, Gap: &gap}).MaxWidth(448)
}

func (p *itemCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	switch p.recipe {
	case "demo", "rtl":
		p.demo(c)
	case "variant":
		items := []display.ItemProps{}
		for i, row := range [][2]string{{"Default Variant", "Transparent background with no border."}, {"Outline Variant", "Outlined style with a visible border."}, {"Muted Variant", "Muted background for secondary content."}} {
			items = append(items, display.ItemProps{Key: row[0], Title: row[0], Description: row[1], Variant: display.ItemVariant(i), Media: itemInbox()})
		}
		p.group(c, "Item variants", items, 24)
	case "size":
		items := []display.ItemProps{}
		for i, row := range [][2]string{{"Default Size", "The standard size for most use cases."}, {"Small Size", "A compact size for dense layouts."}, {"Extra Small Size", "The most compact size available."}} {
			items = append(items, display.ItemProps{Key: row[0], Title: row[0], Description: row[1], Variant: display.ItemOutline, Size: display.ItemSize(i), Media: itemInbox()})
		}
		p.group(c, "Item sizes", items, 24)
	case "icon":
		p.group(c, "Security item", []display.ItemProps{{Key: "security", Title: "Security Alert", Description: "New login detected from unknown device.", Variant: display.ItemOutline, Media: &display.ItemMediaProps{Variant: display.ItemMediaIcon, Icon: outlined.Shield(), Decorative: true}, Actions: func(child *ui.Context) {
			p.actionButton(child, "Review", "Review security alert", "Security Alert", control.Outline, control.Small, nil)
		}}}, 24).MaxWidth(512)
	case "avatar":
		p.avatars(c)
	case "group":
		p.people(c)
	case "image":
		p.music(c)
	case "header":
		p.models(c)
	case "dropdown":
		p.peopleMenu(c, h, native)
	case "link":
		p.group(c, "Item links", []display.ItemProps{
			{Key: "documentation", Title: "Visit our documentation", Description: "Learn how to get started with our components.", Primary: p.primary("Visit our documentation"), Content: itemLinkContent("Visit our documentation", "Learn how to get started with our components.", outlined.ChevronRight())},
			{Key: "external", Title: "External resource", Description: "Opens in a new tab with security attributes.", Variant: display.ItemOutline, Primary: p.primary("External resource"), Content: itemLinkContent("External resource", "Opens in a new tab with security attributes.", outlined.OpenInNew())},
		}, 16)
	default:
		p.create(c)
	}
}

func itemDemoWords(language string) (string, string, string, string) {
	switch language {
	case "ar":
		return "عنصر أساسي", "عنصر بسيط يحتوي على عنوان ووصف.", "إجراء", "تم التحقق من ملفك الشخصي."
	case "he":
		return "פריט בסיסי", "פריט פשוט עם כותרת ותיאור.", "פעולה", "הפרופיל שלך אומת."
	}
	return "Basic Item", "A simple item with title and description.", "Action", "Your profile has been verified."
}
func (p *itemCompositionPreviewState) demo(c *ui.Context) {
	language := "en"
	if p.recipe == "rtl" {
		language = p.language
	}
	title, desc, action, verified := itemDemoWords(language)
	actionLabel := "Run basic item action"
	if p.recipe == "rtl" {
		actionLabel = action
	}
	p.group(c, "Basic item examples", []display.ItemProps{
		{Key: "basic", Title: title, Description: desc, Variant: display.ItemOutline, Actions: func(child *ui.Context) {
			p.actionButton(child, action, actionLabel, title, control.Outline, control.Small, nil)
		}},
		{Key: "verified", Title: verified, Size: display.ItemSmall, Variant: display.ItemOutline, Media: &display.ItemMediaProps{Icon: outlined.Verified(), IconSize: 20, Decorative: true}, Primary: p.primary(verified), Content: itemLinkContent(verified, "", outlined.ChevronRight())},
	}, 24)
}

var itemPeople = []struct{ name, email, fallback string }{{"shadcn", "shadcn@vercel.com", "CN"}, {"maxleiter", "maxleiter@vercel.com", "LR"}, {"evilrabbit", "evilrabbit@vercel.com", "ER"}}

func itemAvatar(index int, gray bool, size float32) *display.ItemMediaProps {
	return &display.ItemMediaProps{Content: func(c *ui.Context) {
		image := cardMeetingImages[index]
		if gray {
			image = itemGrayPeople[index]
		}
		display.Avatar(c, display.AvatarProps{Image: image, Fallback: itemPeople[index].fallback, Label: "@" + itemPeople[index].name, Size: size})
	}}
}
func (p *itemCompositionPreviewState) avatars(c *ui.Context) {
	p.group(c, "Item avatars", []display.ItemProps{
		{Key: "evilrabbit", Title: "Evil Rabbit", Description: "Last seen 5 months ago", Variant: display.ItemOutline, Media: itemAvatar(2, false, 40), Actions: func(child *ui.Context) {
			p.actionButton(child, "", "Invite Evil Rabbit", "Invite Evil Rabbit", control.Outline, control.IconSmall, outlined.Add()).Radius(999)
		}},
		{Key: "team", Title: "No Team Members", Description: "Invite your team to collaborate on this project.", Variant: display.ItemOutline, Media: &display.ItemMediaProps{Content: func(child *ui.Context) {
			w, _ := child.Size()
			start := 0
			if w < 640 {
				start = 2
			}
			items := []display.AvatarProps{}
			for i := start; i < 3; i++ {
				items = append(items, display.AvatarProps{Key: itemPeople[i].name, Label: "@" + itemPeople[i].name, Image: itemGrayPeople[i], Fallback: itemPeople[i].fallback})
			}
			display.AvatarGroup(child, display.AvatarGroupProps{Label: "Item team avatars", Items: items, Max: 3, Total: len(items)})
		}}, Actions: func(child *ui.Context) {
			p.actionButton(child, "Invite", "Invite team members", "Invite team members", control.Outline, control.Small, nil)
		}},
	}, 24).MaxWidth(512)
}
func (p *itemCompositionPreviewState) people(c *ui.Context) {
	items := []display.ItemProps{}
	for i, person := range itemPeople {
		i, person := i, person
		items = append(items, display.ItemProps{Key: person.name, Title: person.name, Description: person.email, Variant: display.ItemOutline, Media: itemAvatar(i, true, 32), Actions: func(child *ui.Context) {
			p.actionButton(child, "", "Invite "+person.name, "Invite "+person.name, control.Ghost, control.Icon, outlined.Add()).Radius(999)
		}})
	}
	p.group(c, "People items", items, 16).MaxWidth(384)
}

func (p *itemCompositionPreviewState) peopleMenu(c *ui.Context, h *desktop.Host, native bool) {
	items := []overlay.MenuItem{}
	for _, person := range itemPeople {
		items = append(items, overlay.MenuItem{Key: person.name, Label: person.name + " · " + person.email})
	}
	render := func(e ui.Element) {
		control.StyleButton(c, e, "Select", control.ButtonProps{Variant: control.Outline, Label: "Select item person"})
		e.Children(func() { indicator.ChevronDown(c) })
	}
	var chosen string
	var err error
	if native {
		chosen, err = desktop.NativeMenu(c, overlay.MenuProps{Items: items}, render)
	} else {
		anchor := ui.ButtonBase(c)
		render(anchor)
		chosen, err = desktop.Menu(c, h, "item-people", anchor, &p.menuOpen, &p.menu, desktop.MenuProps{Label: "Item people menu", Items: items, Width: 192, ItemHeight: 60, MaxHeight: 320, Placement: mygo.PlacementBottomEnd, Gap: 6, ItemContent: func(child *ui.Context, item overlay.MenuItem) {
			for i, person := range itemPeople {
				if person.name == item.Key {
					display.Item(child, display.ItemProps{Key: person.name, Size: display.ItemExtraSmall, Media: itemAvatar(i, true, 26), Content: func(current *ui.Context) {
						display.ItemTitle(current, person.name)
						display.ItemDescription(current, person.email, display.ItemExtraSmall).LineHeight(1)
					}}).Padding(8).FillWidth()
				}
			}
		}})
	}
	if err != nil {
		p.popupError = err.Error()
	}
	if chosen != "" {
		p.selectedPerson = chosen
		p.action = "Local person selection: " + chosen
	}
}

func (p *itemCompositionPreviewState) create(c *ui.Context) {
	v, size := itemRecipeShape(p.recipe)
	switch {
	case strings.Contains(p.recipe, "-links"):
		p.createLinks(c, v)
	case strings.Contains(p.recipe, "-group"):
		p.createGroup(c, v)
	case p.recipe == "create-separator":
		p.separator(c)
	case p.recipe == "create-header" || p.recipe == "create-footer" || p.recipe == "create-both":
		p.headersFooters(c)
	case strings.Contains(p.recipe, "-image"):
		p.createImages(c, v, size)
	default:
		p.createRows(c, v, size)
	}
}

func (p *itemCompositionPreviewState) createRows(c *ui.Context, v display.ItemVariant, size display.ItemSize) {
	titles := []string{"Title Only", "Title + Button", "Title + Description", "Title + Description + Button", "Media + Title", "Media + Title + Button", "Media + Title + Description", "Media + Title + Description + Button", "Multiple Actions"}
	desc := map[int]string{2: "This is a description that provides additional context.", 3: "This item includes a title, description, and action button.", 6: "This item includes media, title, and description.", 7: "Complete item with all components: media, title, description, and button.", 8: "Item with multiple action buttons in the actions area."}
	items := []display.ItemProps{}
	for i, title := range titles {
		i, title := i, title
		item := display.ItemProps{Key: fmt.Sprint(i), Title: title, Description: desc[i], Variant: v, Size: size}
		if i >= 4 && i <= 7 {
			item.Media = itemInbox()
		}
		if i == 1 || i == 3 || i == 5 || i == 7 {
			buttonSize := control.Small
			if size == display.ItemDefaultSize && (i == 1 || i == 3) {
				buttonSize = control.DefaultSize
			}
			variant := control.Outline
			if i == 5 || i == 7 {
				variant = control.Default
			}
			item.Actions = func(child *ui.Context) {
				p.actionButton(child, "Action", "Action for "+title, title, variant, buttonSize, nil)
			}
		}
		if i == 8 {
			item.Actions = func(child *ui.Context) {
				p.actionButton(child, "Cancel", "Cancel multiple actions", "Cancel multiple actions", control.Outline, control.Small, nil)
				p.actionButton(child, "Confirm", "Confirm multiple actions", "Confirm multiple actions", control.Default, control.Small, nil)
			}
		}
		items = append(items, item)
	}
	gap := float32(24)
	if size == display.ItemSmall {
		gap = 16
	} else if size == display.ItemExtraSmall {
		gap = 12
	}
	p.group(c, "Create item collection", items, gap).MaxWidth(512)
}
func (p *itemCompositionPreviewState) createLinks(c *ui.Context, v display.ItemVariant) {
	titles := []string{"Title Only (Link)", "Title + Description (Link)", "Media + Title (Link)", "Media + Title + Description (Link)", "With Actions (Link)"}
	descriptions := []string{"", "Clickable item with title and description.", "", "Complete link item with media, title, and description.", "Link item that also has action buttons."}
	items := []display.ItemProps{}
	for i, title := range titles {
		item := display.ItemProps{Key: fmt.Sprint(i), Title: title, Description: descriptions[i], Variant: v, Primary: p.primary(title)}
		if i == 2 || i == 3 {
			item.Media = itemInbox()
		}
		if i == 4 {
			item.Actions = func(child *ui.Context) {
				p.actionButton(child, "Share", "Share linked item", "Share linked item", control.Outline, control.Small, nil)
			}
		}
		items = append(items, item)
	}
	p.group(c, "Create item links", items, 16)
}
func (p *itemCompositionPreviewState) createGroup(c *ui.Context, v display.ItemVariant) {
	items := []display.ItemProps{}
	for i, ordinal := range []string{"First", "Second", "Third"} {
		title := fmt.Sprintf("Item %d", i+1)
		desc := ordinal + " item in the group."
		if v == display.ItemOutline {
			desc = ordinal + " item with icon."
		} else if v == display.ItemMuted {
			desc = ordinal + " item in muted group."
		}
		item := display.ItemProps{Key: fmt.Sprint(i), Title: title, Description: desc, Variant: v}
		if v == display.ItemOutline {
			item.Media = itemInbox()
		}
		if v == display.ItemMuted {
			item.Actions = func(child *ui.Context) {
				p.actionButton(child, "Action", "Run "+title, title, control.Outline, control.Small, nil)
			}
		}
		items = append(items, item)
	}
	p.group(c, "Create grouped items", items, 16)
}
func (p *itemCompositionPreviewState) separator(c *ui.Context) {
	items := []display.ItemProps{}
	for _, row := range [][2]string{{"Inbox", "View all incoming messages."}, {"Sent", "View all sent messages."}, {"Drafts", "View all draft messages."}, {"Archive", "View archived messages."}} {
		items = append(items, display.ItemProps{Key: row[0], Title: row[0], Description: row[1], Variant: display.ItemOutline, Media: itemInbox()})
	}
	display.ItemGroup(c, display.ItemGroupProps{Label: "Separated mailbox items", Items: items, Separators: true}).MaxWidth(448)
}
func (p *itemCompositionPreviewState) headersFooters(c *ui.Context) {
	var data [][4]string
	switch p.recipe {
	case "create-header":
		data = [][4]string{{"Design System", "Component Library", "A comprehensive collection of reusable UI components for building consistent interfaces.", ""}, {"Marketing", "Campaign Analytics", "Track performance metrics and engagement rates across all marketing channels.", ""}, {"Engineering", "API Documentation", "Complete reference guide for all available endpoints and authentication methods.", ""}}
	case "create-footer":
		data = [][4]string{{"", "Quarterly Report Q4 2024", "Financial overview including revenue, expenses, and growth metrics for the fourth quarter.", "Last updated 2 hours ago"}, {"", "User Research Findings", "Insights from interviews and surveys conducted with 50+ users across different demographics.", "Created by Sarah Chen"}, {"", "Product Roadmap", "Planned features and improvements scheduled for the next three months.", "12 comments"}}
	default:
		data = [][4]string{{"Team Project", "Website Redesign", "Complete overhaul of the company website with modern design principles and improved user experience.", "Updated 5 minutes ago"}, {"Client Work", "Mobile App Development", "Building a cross-platform mobile application for iOS and Android with React Native.", "Status: In Progress"}, {"Documentation", "API Integration Guide", "Step-by-step instructions for integrating third-party APIs with authentication and error handling.", "Category: Technical • 3 attachments"}}
	}
	items := []display.ItemProps{}
	for i, row := range data {
		row := row
		item := display.ItemProps{Key: fmt.Sprint(i), Title: row[1], Description: row[2], Variant: display.ItemVariant(i)}
		if row[0] != "" {
			item.Header = func(child *ui.Context) { ui.Text(child, row[0]).FontSize(14).FontWeight(500) }
		}
		if row[3] != "" {
			item.Footer = func(child *ui.Context) {
				ui.Text(child, row[3]).FontSize(14).TextColor(theme.TokensOf(child).MutedForeground)
			}
		}
		items = append(items, item)
	}
	p.group(c, "Item headers and footers", items, 24).MaxWidth(512)
}

func (p *itemCompositionPreviewState) createImages(c *ui.Context, v display.ItemVariant, size display.ItemSize) {
	titles := []string{"Project Dashboard", "Document", "File Attachment"}
	descriptions := []string{"Overview of project settings and configuration.", "A document with metadata displayed.", "Complete file with image, title, description, and action button."}
	items := []display.ItemProps{}
	for i, title := range titles {
		i, title := i, title
		desc := descriptions[i]
		if size == display.ItemExtraSmall {
			desc = ""
		}
		item := display.ItemProps{Key: fmt.Sprint(i), Title: title, Description: desc, Variant: v, Size: size, Media: &display.ItemMediaProps{Variant: display.ItemMediaImage, Image: itemProjectImages[i], Label: []string{"Project", "Document", "File"}[i]}}
		if i > 0 {
			text, variant := "View", control.Outline
			if i == 2 {
				text, variant = "Download", control.Default
			}
			item.Actions = func(child *ui.Context) {
				p.actionButton(child, text, text+" "+title, text+" "+title, variant, control.Small, nil)
			}
		}
		items = append(items, item)
	}
	p.group(c, "Create image items", items, 24)
}

var itemSongs = []struct{ title, artist, album, duration string }{{"Midnight City Lights", "Neon Dreams", "Electric Nights", "3:45"}, {"Coffee Shop Conversations", "The Morning Brew", "Urban Stories", "4:05"}, {"Digital Rain", "Cyber Symphony", "Binary Beats", "3:30"}}

func (p *itemCompositionPreviewState) music(c *ui.Context) {
	items := []display.ItemProps{}
	for i, song := range itemSongs {
		i, song := i, song
		items = append(items, display.ItemProps{Key: song.title, Title: song.title, Description: song.artist + " · " + song.album + " · " + song.duration, Variant: display.ItemOutline, Primary: p.primary(song.title), Media: &display.ItemMediaProps{Variant: display.ItemMediaImage, Image: itemSongImages[i], Label: song.title}, Content: func(child *ui.Context) {
			ui.Row(child).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Column(child).Basis(0).Grow(1).MinWidth(0).Gap(4).Children(func() {
					ui.RichText(child).MinWidth(0).FontSize(14).FontWeight(500).MaxLines(1).Children(func() {
						ui.Text(child, song.title+" - ")
						ui.Text(child, song.album).TextColor(theme.TokensOf(child).MutedForeground)
					})
					display.ItemDescription(child, song.artist)
				})
				ui.Text(child, song.duration).Shrink(0).FontSize(14).TextColor(theme.TokensOf(child).MutedForeground).TextAlign(ui.Center)
			})
		}})
	}
	p.group(c, "Music tracks", items, 16)
}
func (p *itemCompositionPreviewState) models(c *ui.Context) {
	ui.Row(c).FillWidth().MaxWidth(576).MinWidth(0).Gap(16).AlignItems(ui.Start).Role(ui.RoleList).Label("Model image items").Children(func() {
		for i, row := range [][2]string{{"v0-1.5-sm", "Everyday tasks and UI generation."}, {"v0-1.5-lg", "Advanced thinking or reasoning."}, {"v0-2.0-mini", "Open Source model for everyone."}} {
			i, row := i, row
			ui.Column(c.Key(row[0])).Basis(0).Grow(1).MinWidth(0).Children(func() {
				display.Item(c, display.ItemProps{Key: row[0], Title: row[0], Description: row[1], Variant: display.ItemOutline, Header: func(child *ui.Context) {
					ui.Image(child, itemModelImages[i]).FillWidth().AspectRatio(1).Fit(ui.Cover).Radius(theme.TokensOf(child).RadiusSM()).Role(ui.RoleImage).Label(row[0])
				}}).Role(ui.RoleListItem).Label(row[0]).PositionInSet(i+1, 3)
			})
		}
	})
}

// Original images are embedded unchanged; source-map.json records URL, response
// type and SHA. Grayscale is an application image transform, never a fake image.
//
//go:embed item-assets/model-small.jpg
var itemModelSmall []byte

//go:embed item-assets/model-large.jpg
var itemModelLarge []byte

//go:embed item-assets/model-mini.jpg
var itemModelMini []byte

//go:embed item-assets/song-midnight.png
var itemSongMidnight []byte

//go:embed item-assets/song-coffee.png
var itemSongCoffee []byte

//go:embed item-assets/song-rain.png
var itemSongRain []byte

//go:embed item-assets/project.png
var itemProject []byte

//go:embed item-assets/document.png
var itemDocument []byte

//go:embed item-assets/file.png
var itemFile []byte

var itemModelImages = []*ui.Bitmap{cardDecodeBitmap(itemModelSmall), cardDecodeBitmap(itemModelLarge), cardDecodeBitmap(itemModelMini)}
var itemSongImages = []*ui.Bitmap{cardFilteredImage(cardDecodeImage(itemSongMidnight), 1, nil), cardFilteredImage(cardDecodeImage(itemSongCoffee), 1, nil), cardFilteredImage(cardDecodeImage(itemSongRain), 1, nil)}
var itemProjectImages = []*ui.Bitmap{cardFilteredImage(cardDecodeImage(itemProject), 1, nil), cardFilteredImage(cardDecodeImage(itemDocument), 1, nil), cardFilteredImage(cardDecodeImage(itemFile), 1, nil)}
var itemGrayPeople = []*ui.Bitmap{cardFilteredImage(cardDecodeImage(cardShadcnBytes), 1, nil), cardFilteredImage(cardDecodeImage(cardMaxleiterBytes), 1, nil), cardFilteredImage(cardDecodeImage(cardEvilrabbitBytes), 1, nil)}
