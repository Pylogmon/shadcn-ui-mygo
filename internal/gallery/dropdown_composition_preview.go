package gallery

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type dropdownSourceItem struct {
	Label, Group, Shortcut, Check, Radio, Value         string
	Icon                                                *ui.SVG
	Separator, Inset, GroupInset, Disabled, Destructive bool
	Children                                            []dropdownSourceItem
}
type dropdownSourceRecipe struct {
	Key, Title, Trigger string
	Width               float32
	Items               []dropdownSourceItem
	Checks              map[string]bool
	Radios              map[string]string
}
type dropdownSourceModel struct {
	checks map[string]*bool
	radios map[string]string
}
type dropdownSourceMenu struct {
	open  bool
	state desktop.MenuState
}

// Only application data, public menu state and a stable focus handle persist.
// Child content is always built with the Context supplied by the active window.
type dropdownCompositionPreviewState struct {
	ready, native, dialog                          bool
	recipe, previous, language, action, popupError string
	models                                         map[string]*dropdownSourceModel
	menus                                          map[string]*dropdownSourceMenu
	nativeIcons                                    map[*ui.SVG]ui.MenuIcon
	dialogReturn                                   ui.Handle
}

func (p *dropdownCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.models = map[string]*dropdownSourceModel{}
	p.menus = map[string]*dropdownSourceMenu{}
	p.nativeIcons = map[*ui.SVG]ui.MenuIcon{}
}
func dropdownRecipe(key string) dropdownSourceRecipe {
	for _, r := range dropdownSourceRecipes {
		if r.Key == key {
			return r
		}
	}
	return dropdownRecipe("demo")
}
func (p *dropdownCompositionPreviewState) model(r dropdownSourceRecipe) *dropdownSourceModel {
	if m := p.models[r.Key]; m != nil {
		return m
	}
	m := &dropdownSourceModel{checks: map[string]*bool{}, radios: map[string]string{}}
	for k, v := range r.Checks {
		value := v
		m.checks[k] = &value
	}
	for k, v := range r.Radios {
		m.radios[k] = v
	}
	p.models[r.Key] = m
	return m
}
func (p *dropdownCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
		return
	}
	if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}
func (p *dropdownCompositionPreviewState) closeMenus(h *desktop.Host) {
	for key, m := range p.menus {
		if h != nil {
			h.Close(key)
		}
		m.open = false
		m.state.Reset()
	}
}
func (p *dropdownCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if native != p.native {
		p.closeMenus(h)
		p.native = native
		p.popupError = ""
		if h != nil {
			h.Close("dropdown-recipe")
			h.Close("dropdown-language")
		}
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Dropdown Menu compositions"}, display.CardSlots{
		Title: "Dropdown Menu compositions", Description: "Docs and Create menus with local actions and independent source state.",
		Content: func(child *ui.Context) {
			options := make([]control.Option, 0, len(dropdownSourceRecipes))
			for _, r := range dropdownSourceRecipes {
				options = append(options, control.Option{Value: r.Key, Label: r.Title})
			}
			p.choose(child, h, native, "dropdown-recipe", &p.recipe, control.SelectProps{Label: "Dropdown Menu configuration", Options: options})
			if p.previous != p.recipe {
				p.closeMenus(h)
				p.dialog = false
				p.previous = p.recipe
			}
			if p.recipe == "rtl" {
				p.choose(child, h, native, "dropdown-language", &p.language, control.SelectProps{Label: "Dropdown Menu language", Options: statusLanguages()})
			}
			ui.Column(child.Key("dropdown-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() {
				if p.recipe == "rtl" {
					ui.Column(child).Direction(childDirection(p.language)).Children(func() { p.draw(child, h, native) })
				} else {
					p.draw(child, h, native)
				}
			})
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		}})
}
func dropdownItemKey(path string, index int) string { return fmt.Sprintf("%s/%d", path, index) }
func (p *dropdownCompositionPreviewState) items(r dropdownSourceRecipe) []overlay.MenuItem {
	model := p.model(r)
	var build func([]dropdownSourceItem, string) []overlay.MenuItem
	build = func(source []dropdownSourceItem, path string) []overlay.MenuItem {
		result := make([]overlay.MenuItem, 0, len(source))
		radioGroups := map[string]string{}
		for i, s := range source {
			key := dropdownItemKey(path, i)
			label, group := s.Label, s.Group
			if r.Key == "rtl" {
				label = dropdownTranslation(p.language, label)
				group = dropdownTranslation(p.language, group)
			}
			// Group is also the SDK radio mutual-exclusion identity. Carry the
			// source caption onto every sibling radio; the renderer suppresses
			// repeated captions, and the child updates atomically before a
			// parent frame rebuilds from application state.
			if s.Radio != "" {
				if group != "" {
					radioGroups[s.Radio] = group
				} else {
					group = radioGroups[s.Radio]
				}
			}
			item := overlay.MenuItem{Key: key, Label: label, Group: group, Shortcut: s.Shortcut, Icon: s.Icon, Separator: s.Separator, Inset: s.Inset, GroupInset: s.GroupInset, Disabled: s.Disabled, Destructive: s.Destructive}
			if s.Check != "" {
				item.Checked = model.checks[s.Check]
			}
			if s.Radio != "" {
				item.Radio = true
				item.Selected = model.radios[s.Radio] == s.Value
			}
			item.Children = build(s.Children, key)
			result = append(result, item)
		}
		return result
	}
	return build(r.Items, r.Key)
}

// Platform menu images are cached public handles; no frame Context or Element
// is retained. The OS decides how template images appear in its native menus.
func (p *dropdownCompositionPreviewState) withNativeIcons(items []overlay.MenuItem) {
	for i := range items {
		item := &items[i]
		if item.Icon != nil {
			icon := p.nativeIcons[item.Icon]
			if icon == nil {
				created, err := mygo.NewMenuIconSVG(item.Icon, true)
				if err != nil {
					p.popupError = err.Error()
				} else {
					icon = created
					p.nativeIcons[item.Icon] = icon
				}
			}
			item.NativeIcon = icon
		}
		p.withNativeIcons(item.Children)
	}
}

func (p *dropdownCompositionPreviewState) selected(r dropdownSourceRecipe, key string) {
	model := p.model(r)
	var find func([]dropdownSourceItem, string) bool
	find = func(items []dropdownSourceItem, path string) bool {
		for i, item := range items {
			itemKey := dropdownItemKey(path, i)
			if itemKey == key {
				if item.Radio != "" {
					model.radios[item.Radio] = item.Value
				}
				label := item.Label
				if r.Key == "rtl" {
					label = dropdownTranslation(p.language, label)
				}
				p.action = "Local dropdown action: " + label
				return true
			}
			if find(item.Children, itemKey) {
				return true
			}
		}
		return false
	}
	find(r.Items, r.Key)
}
func (p *dropdownCompositionPreviewState) trigger(c *ui.Context, b ui.Element, label, avatar string) ui.Element {
	if avatar == "" {
		return control.StyleButton(c, b, label, control.ButtonProps{Label: label, Variant: control.Outline})
	}
	b = control.StyleButton(c, b, "", control.ButtonProps{Label: label, Variant: control.Ghost, Size: control.Icon}).Radius(999)
	if avatar == "account" {
		b = control.StyleButton(c, b, "", control.ButtonProps{Label: label, Variant: control.Outline}).Height(48).Width(200).PaddingX(8).Justify(ui.Start)
	}
	return b.Children(func() {
		fallback := "CN"
		if avatar != "account" {
			fallback = "LR"
		}
		display.Avatar(c, display.AvatarProps{Image: cardMeetingImages[0], Fallback: fallback, Label: "shadcn", Size: 32})
		if avatar == "account" {
			ui.Column(c).Grow(1).MinWidth(0).Gap(0).AlignItems(ui.Start).Children(func() {
				ui.Text(c, "shadcn").FontSize(14).FontWeight(600)
				ui.Text(c, "shadcn@example.com").FontSize(12).TextColor(theme.TokensOf(c).MutedForeground)
			})
			ui.Icon(c, outlined.UnfoldMore()).Size(16, 16).AccessHidden(true)
		}
	})
}
func (p *dropdownCompositionPreviewState) menu(c *ui.Context, h *desktop.Host, native bool, r dropdownSourceRecipe, id, label, side, avatar string) {
	key := "dropdown-source:" + r.Key + ":" + id
	m := p.menus[key]
	if m == nil {
		m = &dropdownSourceMenu{}
		p.menus[key] = m
	}
	items := p.items(r)
	if native {
		p.withNativeIcons(items)
		value, err := desktop.NativeDropdownMenu(c.Key(id), overlay.MenuProps{Items: items, Width: r.Width}, func(b ui.Element) { p.trigger(c, b, label, avatar) })
		if err != nil {
			p.popupError = err.Error()
		}
		if value != "" {
			p.selected(r, value)
		}
		return
	}
	anchor := p.trigger(c, ui.ButtonBase(c.Key(id)), label, avatar)
	width := r.Width
	if avatar == "account" {
		width = max(224, anchor.LayoutBounds().W)
	}
	placement := popupPlacement(side, c.Direction())
	if avatar == "top" {
		placement = mygo.PlacementTopEnd
	}
	_, err := desktop.DropdownMenu(c, h, key, anchor, &m.open, &m.state, desktop.MenuProps{Label: r.Title + " menu", Items: items, Width: width, MaxHeight: 600, Placement: placement, Gap: 4, OnChoose: func(key string) { p.selected(r, key) }})
	if err != nil {
		p.popupError = err.Error()
	}
}
func (p *dropdownCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	r := dropdownRecipe(p.recipe)
	switch p.recipe {
	case "create-sides":
		ui.Row(c).Wrap().Gap(8).Justify(ui.Center).Children(func() {
			for _, side := range []string{"inline-start", "left", "top", "bottom", "right", "inline-end"} {
				p.menu(c, h, native, r, side, strings.ReplaceAll(side, "-", " "), side, "")
			}
		})
	case "create-with-avatar":
		ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
			p.menu(c, h, native, r, "account", "Open shadcn account menu", "start", "account")
			p.menu(c, h, native, r, "top", "Open shadcn avatar menu above", "top", "top")
		})
	case "avatar":
		p.menu(c, h, native, r, "avatar", "Open shadcn avatar menu", "end", "avatar")
	case "create-in-dialog":
		b := control.Button(c, "Open Dialog", control.ButtonProps{Label: "Open dropdown dialog", Variant: control.Outline}).Bind(&p.dialogReturn)
		if b.Clicked() {
			p.dialog = true
		}
		overlay.Dialog(c, &p.dialog, overlay.DialogProps{Title: "Dropdown Menu Example", Description: "Click the button below to see the dropdown menu.", ReturnFocus: &p.dialogReturn}, func() { p.menu(c, h, native, r, "dialog", "Open Menu", "start", "") })
	default:
		label := r.Trigger
		if p.recipe == "rtl" {
			label = dropdownTranslation(p.language, "{t.open}")
		}
		p.menu(c, h, native, r, "main", label, "start", "")
	}
}

func dropdownTranslation(language, key string) string {
	if !strings.HasPrefix(key, "{t.") {
		return key
	}
	key = strings.TrimSuffix(strings.TrimPrefix(key, "{t."), "}")
	words := map[string][3]string{
		"open": {"Open", "افتح القائمة", "פתח תפריט"}, "account": {"Account", "الحساب", "חשבון"}, "profile": {"Profile", "الملف الشخصي", "פרופיל"}, "billing": {"Billing", "الفوترة", "חיוב"}, "settings": {"Settings", "الإعدادات", "הגדרות"}, "logout": {"Log out", "تسجيل الخروج", "התנתק"}, "team": {"Team", "الفريق", "הצוות"}, "inviteUsers": {"Invite users", "دعوة المستخدمين", "הזמן משתמשים"}, "email": {"Email", "البريد الإلكتروني", "אימייל"}, "message": {"Message", "رسالة", "הודעה"}, "more": {"More", "المزيد", "עוד"}, "calendar": {"Calendar", "تقويم", "יומן"}, "chat": {"Chat", "دردشة", "צ'אט"}, "webhook": {"Webhook", "خطاف ويب", "Webhook"}, "advanced": {"Advanced...", "متقدم...", "מתקדם..."}, "newTeam": {"New Team", "فريق جديد", "צוות חדש"}, "view": {"View", "عرض", "תצוגה"}, "statusBar": {"Status Bar", "شريط الحالة", "שורת סטטוס"}, "activityBar": {"Activity Bar", "شريط النشاط", "שורת פעילות"}, "panel": {"Panel", "اللوحة", "לוח"}, "position": {"Position", "الموضع", "מיקום"}, "top": {"Top", "أعلى", "למעלה"}, "bottom": {"Bottom", "أسفل", "למטה"}, "right": {"Right", "يمين", "ימין"}, "left": {"Left", "يسار", "שמאל"},
	}
	index := 0
	if language == "ar" {
		index = 1
	}
	if language == "he" {
		index = 2
	}
	return words[key][index]
}

// Source menu data is derived from the pinned docs and Create examples.
var dropdownSourceRecipes = []dropdownSourceRecipe{
	{Key: "avatar", Title: "Avatar", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Account", Icon: outlined.Verified()},
		{Label: "Billing", Icon: outlined.CreditCard()},
		{Label: "Notifications", Icon: outlined.Notifications()},
		{Separator: true},
		{Label: "Sign Out", Icon: outlined.Logout()}}},
	{Key: "basic", Title: "Basic", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Profile", Group: "My Account"},
		{Label: "Billing"},
		{Label: "Settings"},
		{Separator: true},
		{Label: "GitHub"},
		{Label: "Support"},
		{Label: "API", Disabled: true}}},
	{Key: "checkboxes-icons", Title: "Checkboxes Icons", Trigger: "Notifications", Width: 192, Items: []dropdownSourceItem{{Label: "Email notifications", Group: "Notification Preferences", Check: "notifications.email", Icon: outlined.Mail()},
		{Label: "SMS notifications", Check: "notifications.sms", Icon: outlined.Chat()},
		{Label: "Push notifications", Check: "notifications.push", Icon: outlined.Notifications()}}, Checks: map[string]bool{"notifications.email": true, "notifications.sms": false, "notifications.push": true}},
	{Key: "checkboxes", Title: "Checkboxes", Trigger: "Open", Width: 160, Items: []dropdownSourceItem{{Label: "Status Bar", Group: "Appearance", Check: "showStatusBar"},
		{Label: "Activity Bar", Check: "showActivityBar", Disabled: true},
		{Label: "Panel", Check: "showPanel"}}, Checks: map[string]bool{"showStatusBar": true, "showActivityBar": false, "showPanel": false}},
	{Key: "complex", Title: "Complex", Trigger: "Complex Menu", Width: 176, Items: []dropdownSourceItem{{Label: "New File", Group: "File", Shortcut: "⌘N", Icon: outlined.Draft()},
		{Label: "New Folder", Shortcut: "⇧⌘N", Icon: outlined.Folder()},
		{Label: "Open Recent", Icon: outlined.FolderOpen(), Children: []dropdownSourceItem{{Label: "Project Alpha", Group: "Recent Projects", Icon: outlined.Code()},
			{Label: "Project Beta", Icon: outlined.Code()},
			{Label: "More Projects", Icon: outlined.MoreHoriz(), Children: []dropdownSourceItem{{Label: "Project Gamma", Icon: outlined.Code()},
				{Label: "Project Delta", Icon: outlined.Code()}}},
			{Separator: true},
			{Label: "Browse...", Icon: outlined.FolderOpen()}}},
		{Separator: true},
		{Label: "Save", Shortcut: "⌘S", Icon: outlined.Save()},
		{Label: "Export", Shortcut: "⇧⌘E", Icon: outlined.Download()},
		{Separator: true},
		{Label: "Show Sidebar", Group: "View", Check: "notifications.email", Icon: outlined.Visibility()},
		{Label: "Show Status Bar", Check: "notifications.sms", Icon: outlined.Dashboard()},
		{Label: "Theme", Icon: outlined.Palette(), Children: []dropdownSourceItem{{Label: "Light", Group: "Appearance", Radio: "theme", Value: "light", Icon: outlined.LightMode()},
			{Label: "Dark", Radio: "theme", Value: "dark", Icon: outlined.DarkMode()},
			{Label: "System", Radio: "theme", Value: "system", Icon: outlined.Computer()}}},
		{Separator: true},
		{Label: "Profile", Group: "Account", Shortcut: "⇧⌘P", Icon: outlined.Person()},
		{Label: "Billing", Icon: outlined.CreditCard()},
		{Label: "Settings", Icon: outlined.Settings(), Children: []dropdownSourceItem{{Label: "Keyboard Shortcuts", Group: "Preferences", Icon: outlined.Keyboard()},
			{Label: "Language", Icon: outlined.Translate()},
			{Label: "Notifications", Icon: outlined.Notifications(), Children: []dropdownSourceItem{{Label: "Push Notifications", Group: "Notification Types", Check: "notifications.push", Icon: outlined.Notifications()},
				{Label: "Email Notifications", Check: "notifications.email", Icon: outlined.Mail()}}},
			{Separator: true},
			{Label: "Privacy & Security", Icon: outlined.Shield()}}},
		{Separator: true},
		{Label: "Help & Support", Icon: outlined.Help()},
		{Label: "Documentation", Icon: outlined.Description()},
		{Separator: true},
		{Label: "Sign Out", Shortcut: "⇧⌘Q", Destructive: true, Icon: outlined.Logout()}}, Checks: map[string]bool{"notifications.email": true, "notifications.sms": false, "notifications.push": true}, Radios: map[string]string{"theme": "light"}},
	{Key: "demo", Title: "Demo", Trigger: "Open", Width: 160, Items: []dropdownSourceItem{{Label: "Profile", Group: "My Account", Shortcut: "⇧⌘P"},
		{Label: "Billing", Shortcut: "⌘B"},
		{Label: "Settings", Shortcut: "⌘S"},
		{Separator: true},
		{Label: "Team"},
		{Label: "Invite users", Children: []dropdownSourceItem{{Label: "Email"},
			{Label: "Message"},
			{Separator: true},
			{Label: "More..."}}},
		{Label: "New Team", Shortcut: "⌘+T"},
		{Separator: true},
		{Label: "GitHub"},
		{Label: "Support"},
		{Label: "API", Disabled: true},
		{Separator: true},
		{Label: "Log out", Shortcut: "⇧⌘Q"}}},
	{Key: "destructive", Title: "Destructive", Trigger: "Actions", Width: 128, Items: []dropdownSourceItem{{Label: "Edit", Icon: outlined.Edit()},
		{Label: "Share", Icon: outlined.Share()},
		{Separator: true},
		{Label: "Delete", Destructive: true, Icon: outlined.Delete()}}},
	{Key: "icons", Title: "Icons", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Profile", Icon: outlined.Person()},
		{Label: "Billing", Icon: outlined.CreditCard()},
		{Label: "Settings", Icon: outlined.Settings()},
		{Separator: true},
		{Label: "Log out", Destructive: true, Icon: outlined.Logout()}}},
	{Key: "radio-group", Title: "Radio Group", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Top", Group: "Panel Position", Radio: "position", Value: "top"},
		{Label: "Bottom", Radio: "position", Value: "bottom"},
		{Label: "Right", Radio: "position", Value: "right"}}, Radios: map[string]string{"position": "bottom"}},
	{Key: "radio-icons", Title: "Radio Icons", Trigger: "Payment Method", Width: 224, Items: []dropdownSourceItem{{Label: "Credit Card", Group: "Select Payment Method", Radio: "paymentMethod", Value: "card", Icon: outlined.CreditCard()},
		{Label: "PayPal", Radio: "paymentMethod", Value: "paypal", Icon: outlined.AccountBalanceWallet()},
		{Label: "Bank Transfer", Radio: "paymentMethod", Value: "bank", Icon: outlined.AccountBalance()}}, Radios: map[string]string{"paymentMethod": "card"}},
	{Key: "rtl", Title: "Rtl", Trigger: "Open", Width: 144, Items: []dropdownSourceItem{{Label: "{t.account}", Children: []dropdownSourceItem{{Label: "{t.profile}", Icon: outlined.Person()},
		{Label: "{t.billing}", Icon: outlined.CreditCard()},
		{Label: "{t.settings}", Icon: outlined.Settings()}}},
		{Separator: true},
		{Label: "{t.team}", Group: "{t.team}"},
		{Label: "{t.inviteUsers}", Children: []dropdownSourceItem{{Label: "{t.email}"},
			{Label: "{t.message}"},
			{Label: "{t.more}", Children: []dropdownSourceItem{{Label: "{t.calendar}"},
				{Label: "{t.chat}"},
				{Separator: true},
				{Label: "{t.webhook}"}}},
			{Separator: true},
			{Label: "{t.advanced}"}}},
		{Label: "{t.newTeam}", Shortcut: "⌘+T"},
		{Separator: true},
		{Label: "{t.statusBar}", Group: "{t.view}", Check: "showStatusBar"},
		{Label: "{t.activityBar}", Check: "showActivityBar"},
		{Label: "{t.panel}", Check: "showPanel"},
		{Separator: true},
		{Label: "{t.top}", Group: "{t.position}", Radio: "position", Value: "top"},
		{Label: "{t.bottom}", Radio: "position", Value: "bottom"},
		{Label: "{t.right}", Radio: "position", Value: "right"},
		{Label: "{t.left}", Radio: "position", Value: "left"},
		{Separator: true},
		{Label: "{t.logout}", Destructive: true}}, Checks: map[string]bool{"showStatusBar": true, "showActivityBar": false, "showPanel": false}, Radios: map[string]string{"position": "bottom"}},
	{Key: "shortcuts", Title: "Shortcuts", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Profile", Group: "My Account", Shortcut: "⇧⌘P"},
		{Label: "Billing", Shortcut: "⌘B"},
		{Label: "Settings", Shortcut: "⌘S"},
		{Separator: true},
		{Label: "Log out", Shortcut: "⇧⌘Q"}}},
	{Key: "submenu", Title: "Submenu", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Team"},
		{Label: "Invite users", Children: []dropdownSourceItem{{Label: "Email"},
			{Label: "Message"},
			{Label: "More options", Children: []dropdownSourceItem{{Label: "Calendly"},
				{Label: "Slack"},
				{Separator: true},
				{Label: "Webhook"}}},
			{Separator: true},
			{Label: "Advanced..."}}},
		{Label: "New Team", Shortcut: "⌘+T"}}},
	{Key: "create-basic", Title: "Create: Basic", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Profile", Group: "My Account"},
		{Label: "Billing"},
		{Label: "Settings"},
		{Separator: true},
		{Label: "GitHub"},
		{Label: "Support"},
		{Label: "API", Disabled: true}}},
	{Key: "create-sides", Title: "Create: Sides", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Profile"},
		{Label: "Billing"},
		{Label: "Settings"}}},
	{Key: "create-with-icons", Title: "Create: With Icons", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Profile", Icon: outlined.Person()},
		{Label: "Billing", Icon: outlined.CreditCard()},
		{Label: "Settings", Icon: outlined.Settings()},
		{Separator: true},
		{Label: "Log out", Destructive: true, Icon: outlined.Logout()}}},
	{Key: "create-with-shortcuts", Title: "Create: With Shortcuts", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Profile", Group: "My Account", Shortcut: "⇧⌘P"},
		{Label: "Billing", Shortcut: "⌘B"},
		{Label: "Settings", Shortcut: "⌘S"},
		{Label: "Keyboard shortcuts", Shortcut: "⌘K"},
		{Separator: true},
		{Label: "Log out", Shortcut: "⇧⌘Q"}}},
	{Key: "create-with-submenu", Title: "Create: With Submenu", Trigger: "Open", Width: 128, Items: []dropdownSourceItem{{Label: "Team"},
		{Label: "Invite users", Children: []dropdownSourceItem{{Label: "Email"},
			{Label: "Message"},
			{Separator: true},
			{Label: "More..."}}},
		{Label: "New Team", Shortcut: "⌘+T"}}},
	{Key: "create-with-checkboxes", Title: "Create: With Checkboxes", Trigger: "Checkboxes", Width: 160, Items: []dropdownSourceItem{{Label: "Status Bar", Group: "Appearance", Check: "showStatusBar", Icon: outlined.Dashboard()},
		{Label: "Activity Bar", Check: "showActivityBar", Disabled: true, Icon: outlined.VitalSigns()},
		{Label: "Panel", Check: "showPanel", Icon: outlined.ViewSidebar()}}, Checks: map[string]bool{"showStatusBar": true, "showActivityBar": false, "showPanel": false}},
	{Key: "create-with-radio", Title: "Create: With Radio Group", Trigger: "Radio Group", Width: 128, Items: []dropdownSourceItem{{Label: "Top", Group: "Panel Position", Radio: "position", Value: "top", Icon: outlined.ArrowUpward()},
		{Label: "Bottom", Radio: "position", Value: "bottom", Icon: outlined.ArrowDownward()},
		{Label: "Right", Radio: "position", Value: "right", Disabled: true, Icon: outlined.ArrowForward()}}, Radios: map[string]string{"position": "bottom"}},
	{Key: "create-with-checkboxes-icons", Title: "Create: Checkboxes with Icons", Trigger: "Notifications", Width: 224, Items: []dropdownSourceItem{{Label: "Email notifications", Group: "Notification Preferences", Check: "notifications.email", Icon: outlined.Mail()},
		{Label: "SMS notifications", Check: "notifications.sms", Icon: outlined.Chat()},
		{Label: "Push notifications", Check: "notifications.push", Icon: outlined.Notifications()}}, Checks: map[string]bool{"notifications.email": true, "notifications.sms": false, "notifications.push": true}},
	{Key: "create-with-radio-icons", Title: "Create: Radio with Icons", Trigger: "Payment Method", Width: 224, Items: []dropdownSourceItem{{Label: "Credit Card", Group: "Select Payment Method", Radio: "paymentMethod", Value: "card", Icon: outlined.CreditCard()},
		{Label: "PayPal", Radio: "paymentMethod", Value: "paypal", Icon: outlined.AccountBalanceWallet()},
		{Label: "Bank Transfer", Radio: "paymentMethod", Value: "bank", Icon: outlined.AccountBalance()}}, Radios: map[string]string{"paymentMethod": "card"}},
	{Key: "create-with-destructive", Title: "Create: With Destructive Items", Trigger: "Actions", Width: 128, Items: []dropdownSourceItem{{Label: "Edit", Icon: outlined.Edit()},
		{Label: "Share", Icon: outlined.Share()},
		{Separator: true},
		{Label: "Archive", Icon: outlined.Archive()},
		{Label: "Delete", Destructive: true, Icon: outlined.Delete()}}},
	{Key: "create-with-avatar", Title: "Create: With Avatar", Trigger: "Open", Width: 224, Items: []dropdownSourceItem{{Label: "Account", Icon: outlined.Verified()},
		{Label: "Billing", Icon: outlined.CreditCard()},
		{Label: "Notifications", Icon: outlined.Notifications()},
		{Separator: true},
		{Label: "Sign Out", Icon: outlined.Logout()}}},
	{Key: "create-in-dialog", Title: "Create: In Dialog", Trigger: "Open Menu", Width: 128, Items: []dropdownSourceItem{{Label: "Copy", Icon: outlined.ContentCopy()},
		{Label: "Cut", Icon: outlined.ContentCut()},
		{Label: "Paste", Icon: outlined.ContentPaste()},
		{Separator: true},
		{Label: "More Options", Children: []dropdownSourceItem{{Label: "Save Page..."},
			{Label: "Create Shortcut..."},
			{Label: "Name Window..."},
			{Separator: true},
			{Label: "Developer Tools"}}},
		{Separator: true},
		{Label: "Delete", Destructive: true, Icon: outlined.Delete()}}},
	{Key: "create-with-inset", Title: "Create: With Inset", Trigger: "Open", Width: 176, Items: []dropdownSourceItem{{Label: "Copy", Group: "Actions", Icon: outlined.ContentCopy()},
		{Label: "Cut", Icon: outlined.ContentCut()},
		{Label: "Paste", Inset: true},
		{Separator: true},
		{Label: "Bookmarks", Group: "Appearance", Check: "showBookmarks", Inset: true, GroupInset: true},
		{Label: "Full URLs", Check: "showUrls", Inset: true},
		{Separator: true},
		{Label: "Light", Group: "Theme", Radio: "theme", Value: "light", Inset: true, GroupInset: true},
		{Label: "Dark", Radio: "theme", Value: "dark", Inset: true},
		{Label: "System", Radio: "theme", Value: "system", Inset: true},
		{Separator: true},
		{Label: "More Options", Inset: true, Children: []dropdownSourceItem{{Label: "Save Page..."},
			{Label: "Create Shortcut..."}}}}, Checks: map[string]bool{"showBookmarks": true, "showUrls": false}, Radios: map[string]string{"theme": "system"}},
	{Key: "create-complex", Title: "Create: Complex", Trigger: "Complex Menu", Width: 224, Items: []dropdownSourceItem{{Label: "Profile", Group: "My Account", Shortcut: "⇧⌘P", Icon: outlined.Person()},
		{Label: "Billing", Shortcut: "⌘B", Icon: outlined.CreditCard()},
		{Label: "Settings", Shortcut: "⌘S", Icon: outlined.Settings()},
		{Separator: true},
		{Label: "Sidebar", Group: "View", Check: "showSidebar", Icon: outlined.ViewSidebar()},
		{Label: "Status Bar", Check: "showStatusBar", Icon: outlined.Dashboard()},
		{Separator: true},
		{Label: "Invite Users", Icon: outlined.Group(), Children: []dropdownSourceItem{{Label: "Email", Icon: outlined.Mail()},
			{Label: "Message", Icon: outlined.Chat()},
			{Separator: true},
			{Label: "More...", Icon: outlined.AddCircle()}}},
		{Separator: true},
		{Label: "Support", Icon: outlined.Help()},
		{Label: "Sign Out", Shortcut: "⇧⌘Q", Destructive: true, Icon: outlined.Logout()}}, Checks: map[string]bool{"showSidebar": true, "showStatusBar": false}},
}
