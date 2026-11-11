package gallery

import (
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

// Persistent models contain application values and public handles only. Menu
// callbacks mutate these values; they never retain a parent-frame Context.
type contextMenuCompositionPreviewState struct {
	ready, native, dialog                      bool
	recipe, language, lastRecipe, lastLanguage string
	action, popupError                         string
	returnFocus                                ui.Handle
	drafts                                     map[string]*contextMenuCompositionDraft
	menus                                      map[string]*contextMenuCompositionMenu
	icons                                      map[*ui.SVG]ui.MenuIcon
}

type contextMenuCompositionDraft struct {
	bookmarks, urls, developer bool
	person, theme              string
}

type contextMenuCompositionMenu struct {
	open bool
	menu desktop.MenuState
}

var contextMenuCompositionRecipes = []control.Option{
	{Value: "basic", Label: "Basic context menu"},
	{Value: "demo", Label: "Browser context menu"},
	{Value: "icons", Label: "Clipboard icons"},
	{Value: "sides", Label: "Four physical sides"},
	{Value: "shortcuts", Label: "Navigation and save shortcuts"},
	{Value: "submenu", Label: "More Tools submenu"},
	{Value: "groups", Label: "File and Edit groups"},
	{Value: "checkboxes", Label: "Three checkboxes"},
	{Value: "radio", Label: "People and Theme radio groups"},
	{Value: "destructive", Label: "Edit, Share and Delete"},
	{Value: "rtl", Label: "Translated browser menu"},
	{Value: "create-basic", Label: "Create: Basic"},
	{Value: "create-icons", Label: "Create: With Icons"},
	{Value: "create-sides", Label: "Create: Six Sides"},
	{Value: "create-shortcuts", Label: "Create: With Shortcuts"},
	{Value: "create-submenu", Label: "Create: With Submenu"},
	{Value: "create-groups", Label: "Create: With Groups, Labels & Separators"},
	{Value: "create-checkboxes", Label: "Create: With Checkboxes"},
	{Value: "create-radio", Label: "Create: With Radio Group"},
	{Value: "create-destructive", Label: "Create: With Destructive Items"},
	{Value: "create-dialog", Label: "Create: In Dialog"},
	{Value: "create-inset", Label: "Create: With Inset"},
}

func (p *contextMenuCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native = true, native
	p.recipe, p.language = "basic", "ar"
	p.lastRecipe, p.lastLanguage = p.recipe, p.language
	p.drafts = map[string]*contextMenuCompositionDraft{}
	p.menus = map[string]*contextMenuCompositionMenu{}
	p.icons = map[*ui.SVG]ui.MenuIcon{}
}

func (p *contextMenuCompositionPreviewState) draft() *contextMenuCompositionDraft {
	if d := p.drafts[p.recipe]; d != nil {
		return d
	}
	d := &contextMenuCompositionDraft{bookmarks: true, developer: true, person: "pedro", theme: "light"}
	if p.recipe == "create-inset" {
		d.theme = "system"
	}
	p.drafts[p.recipe] = d
	return d
}

func (p *contextMenuCompositionPreviewState) close(h *desktop.Host) {
	for key, m := range p.menus {
		m.open = false
		m.menu.Reset()
		if h != nil {
			h.Close(key)
		}
	}
}

func (p *contextMenuCompositionPreviewState) sync(h *desktop.Host, native bool) {
	if p.native != native || p.recipe != p.lastRecipe || p.language != p.lastLanguage {
		p.close(h)
		p.dialog, p.popupError = false, ""
	}
	if p.native != native && h != nil {
		h.Close("source-context-config")
		h.Close("source-context-language")
	}
	p.native, p.lastRecipe, p.lastLanguage = native, p.recipe, p.language
}

func (p *contextMenuCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *contextMenuCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	p.sync(h, native)
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: tr(c, "Context menu compositions")}, display.CardSlots{
		Title: tr(c, "Context menu compositions"), Description: tr(c, "Secondary-click targets, grouped actions and independent system menus."),
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-context-config", &p.recipe, control.SelectProps{Label: tr(child, "Context menu configuration"), Options: localizedOptions(child, contextMenuCompositionRecipes)})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-context-language", &p.language, control.SelectProps{Label: tr(child, "Context menu language"), Options: localizedOptions(child, statusLanguages())})
			}
			p.sync(h, native)
			ui.Column(child.Key("context-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(16).Children(func() {
				if p.recipe == "rtl" {
					ui.Column(child).FillWidth().MinWidth(0).AlignItems(ui.Center).Direction(childDirection(p.language)).Children(func() { p.draw(child, h, native) })
				} else {
					p.draw(child, h, native)
				}
			})
			if p.action != "" {
				ui.Text(child, tr(child, p.action)).FontSize(13).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
}

func (p *contextMenuCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	if p.recipe == "create-dialog" {
		b := control.Button(c, tr(c, "Open Dialog"), control.ButtonProps{Variant: control.Outline}).Bind(&p.returnFocus)
		if b.Clicked() {
			p.dialog = true
		}
		overlay.Dialog(c, &p.dialog, overlay.DialogProps{Title: tr(c, "Context Menu Example"), Description: tr(c, "Right click on the area below to see the context menu."), ReturnFocus: &p.returnFocus, OnOpenChange: func(open bool) {
			if !open {
				p.close(h)
			}
		}}, func() { p.target(c, h, native, "dialog", "Right click here", "") })
		if !p.dialog {
			p.close(h)
		}
		return
	}
	if strings.HasSuffix(p.recipe, "sides") {
		sides := []string{"top", "right", "bottom", "left"}
		columns, width := 2, float32(384)
		if p.recipe == "create-sides" {
			sides = []string{"inline-start", "left", "top", "bottom", "right", "inline-end"}
			width = 560
			if w, _ := c.Size(); w >= 560 {
				columns = 3
			}
		}
		ui.Grid(c).Columns(columns).FillWidth().MaxWidth(width).Gap(16).Children(func() {
			for _, side := range sides {
				label := "Right click (" + side + ")"
				if p.recipe == "create-sides" {
					label = popupSideLabel(side, "en")
				}
				p.target(c, h, native, side, label, side)
			}
		})
		return
	}
	label := "Right click here"
	if p.recipe == "rtl" {
		label = contextMenuWords(p.language)["rightClick"]
	}
	p.target(c, h, native, "target", label, "")
}

func (p *contextMenuCompositionPreviewState) target(c *ui.Context, h *desktop.Host, native bool, suffix, text, side string) {
	key := "source-context:" + p.recipe + ":" + suffix
	cc := c.Key(key)
	t := theme.TokensOf(cc)
	label := "Context menu target"
	if side != "" {
		label = "Context menu target: " + side
	}
	target := ui.Box(cc).FillWidth().MinWidth(0).MaxWidth(320).AspectRatio(16.0/9).Radius(12).Border(1, t.Border).BorderStyle(ui.BorderDashed).FontSize(14).Center().Focusable().TabStop(true).Role(ui.RoleGroup).Label(tr(c, label)).Description(tr(c, text))
	if strings.HasPrefix(p.recipe, "create-") {
		target.MaxWidth(560).AspectRatio(4).Radius(8).BorderStyle(ui.BorderSolid)
	} else {
		longPress := strings.Replace(text, "Right click", "Long press", 1)
		if p.recipe == "rtl" {
			longPress = contextMenuWords(p.language)["longPress"]
		}
		target.Description(text + " / " + longPress)
	}
	target.Children(func() { ui.Text(cc, tr(c, text)).FontSize(14).AccessHidden(true) })
	items := localizedMenuItems(c, p.items())
	if native {
		p.nativeIcons(items)
		chosen, err := desktop.NativeContextMenu(cc, target, overlay.MenuProps{Items: items, Width: p.width()})
		if err != nil {
			p.popupError = err.Error()
		}
		if chosen != "" {
			p.apply(chosen)
		}
		return
	}
	m := p.menus[key]
	if m == nil {
		m = &contextMenuCompositionMenu{}
		p.menus[key] = m
	}
	props := desktop.MenuProps{Items: items, Width: p.width(), Label: tr(c, "Context menu"), OnChoose: func(key string) { p.apply(key) }}
	if side != "" {
		props.Placement = popupPlacement(side, cc.Direction())
	}
	if _, err := desktop.ContextMenu(cc, h, key, target, &m.open, &m.menu, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *contextMenuCompositionPreviewState) width() float32 {
	if p.recipe == "create-inset" {
		return 176
	}
	return 192
}

func (p *contextMenuCompositionPreviewState) nativeIcons(items []overlay.MenuItem) {
	for i := range items {
		item := &items[i]
		if item.Icon != nil {
			icon := p.icons[item.Icon]
			if icon == nil {
				if image, err := mygo.NewMenuIconSVG(item.Icon, true); err == nil {
					icon = image
					p.icons[item.Icon] = icon
				}
			}
			item.NativeIcon = icon
		}
		p.nativeIcons(item.Children)
	}
}

func (p *contextMenuCompositionPreviewState) apply(key string) {
	d := p.draft()
	switch key {
	case "person:pedro", "person:colm":
		d.person = strings.TrimPrefix(key, "person:")
	case "theme:light", "theme:dark", "theme:system":
		d.theme = strings.TrimPrefix(key, "theme:")
	}
	label := key
	var find func([]overlay.MenuItem)
	find = func(items []overlay.MenuItem) {
		for _, item := range items {
			if item.Key == key {
				label = item.Label
			}
			find(item.Children)
		}
	}
	find(p.items())
	p.action = "Local context action: " + label
}

func contextMenuAction(key, label, shortcut string, icon *ui.SVG) overlay.MenuItem {
	return overlay.MenuItem{Key: key, Label: label, Shortcut: shortcut, Icon: icon}
}
func contextMenuSeparator() overlay.MenuItem { return overlay.MenuItem{Separator: true} }

func contextMenuTools(words map[string]string, destructive bool) []overlay.MenuItem {
	items := []overlay.MenuItem{
		contextMenuAction("save-page", words["savePage"], "", nil), contextMenuAction("create-shortcut", words["createShortcut"], "", nil), contextMenuAction("name-window", words["nameWindow"], "", nil),
		contextMenuSeparator(), contextMenuAction("developer-tools", words["developerTools"], "", nil),
	}
	if destructive {
		items = append(items, contextMenuSeparator(), overlay.MenuItem{Key: "delete", Label: words["delete"], Destructive: true})
	}
	return items
}

func (p *contextMenuCompositionPreviewState) items() []overlay.MenuItem {
	d := p.draft()
	kind := strings.TrimPrefix(p.recipe, "create-")
	w := contextMenuWords("en")
	if kind == "rtl" {
		w = contextMenuWords(p.language)
	}
	back := contextMenuAction("back", w["back"], "", nil)
	forward := contextMenuAction("forward", w["forward"], "", nil)
	forward.Disabled = true
	reload := contextMenuAction("reload", w["reload"], "", nil)
	copy := contextMenuAction("copy", "Copy", "", outlined.ContentCopy())
	cut := contextMenuAction("cut", "Cut", "", outlined.ContentCut())
	paste := contextMenuAction("paste", "Paste", "", outlined.ContentPaste())
	remove := overlay.MenuItem{Key: "delete", Label: "Delete", Icon: outlined.Delete(), Destructive: true}
	person := func(group string) []overlay.MenuItem {
		return []overlay.MenuItem{{Key: "person:pedro", Label: "Pedro Duarte", Radio: true, Selected: d.person == "pedro", Group: group}, {Key: "person:colm", Label: "Colm Tuite", Radio: true, Selected: d.person == "colm", Group: group}}
	}
	themes := func() []overlay.MenuItem {
		return []overlay.MenuItem{{Key: "theme:light", Label: "Light", Radio: true, Selected: d.theme == "light", Group: "Theme"}, {Key: "theme:dark", Label: "Dark", Radio: true, Selected: d.theme == "dark", Group: "Theme"}, {Key: "theme:system", Label: "System", Radio: true, Selected: d.theme == "system", Group: "Theme"}}
	}
	switch kind {
	case "basic":
		return []overlay.MenuItem{back, forward, reload}
	case "sides":
		forward.Disabled = false
		return []overlay.MenuItem{back, forward, reload}
	case "icons":
		return []overlay.MenuItem{copy, cut, paste, contextMenuSeparator(), remove}
	case "shortcuts", "demo", "rtl":
		back.Shortcut, forward.Shortcut, reload.Shortcut = "⌘[", "⌘]", "⌘R"
		items := []overlay.MenuItem{back, forward, reload}
		if kind == "shortcuts" {
			return append(items, contextMenuSeparator(), contextMenuAction("save", "Save", "⌘S", nil), contextMenuAction("save-as", "Save As...", "⇧⌘S", nil))
		}
		if kind == "rtl" {
			items[0].Icon, items[1].Icon, items[2].Icon = outlined.ArrowBack(), outlined.ArrowForward(), outlined.Refresh()
			items = []overlay.MenuItem{{Key: "navigation", Label: w["navigation"], SubmenuWidth: 176, Children: items}}
		}
		items = append(items, overlay.MenuItem{Key: "more-tools", Label: w["moreTools"], SubmenuWidth: 176, Children: contextMenuTools(w, true)}, contextMenuSeparator(),
			overlay.MenuItem{Key: "bookmarks", Label: w["showBookmarks"], Checked: &d.bookmarks}, overlay.MenuItem{Key: "urls", Label: w["showFullUrls"], Checked: &d.urls}, contextMenuSeparator())
		return append(items, person(w["people"])...)
	case "submenu":
		copy.Icon, cut.Icon, copy.Shortcut, cut.Shortcut = nil, nil, "⌘C", "⌘X"
		return []overlay.MenuItem{copy, cut, {Key: "more-tools", Label: "More Tools", Children: contextMenuTools(w, true)}}
	case "groups":
		items := []overlay.MenuItem{
			{Key: "new-file", Label: "New File", Shortcut: "⌘N", Group: "File"}, {Key: "open-file", Label: "Open File", Shortcut: "⌘O", Group: "File"}, {Key: "save", Label: "Save", Shortcut: "⌘S", Group: "File"}, contextMenuSeparator(),
			{Key: "undo", Label: "Undo", Shortcut: "⌘Z", Group: "Edit"}, {Key: "redo", Label: "Redo", Shortcut: "⇧⌘Z", Group: "Edit"}, contextMenuSeparator(),
			{Key: "cut", Label: "Cut", Shortcut: "⌘X"}, {Key: "copy", Label: "Copy", Shortcut: "⌘C"}, {Key: "paste", Label: "Paste", Shortcut: "⌘V"}, contextMenuSeparator(),
			{Key: "delete", Label: "Delete", Shortcut: "⌫", Destructive: true},
		}
		return items
	case "checkboxes":
		return []overlay.MenuItem{{Key: "bookmarks", Label: "Show Bookmarks Bar", Checked: &d.bookmarks}, {Key: "urls", Label: "Show Full URLs", Checked: &d.urls}, {Key: "developer", Label: "Show Developer Tools", Checked: &d.developer}}
	case "radio":
		return append(append(person("People"), contextMenuSeparator()), themes()...)
	case "destructive":
		items := []overlay.MenuItem{contextMenuAction("edit", "Edit", "", outlined.Edit()), contextMenuAction("share", "Share", "", outlined.Share()), contextMenuSeparator()}
		if p.recipe == "create-destructive" {
			items = append(items, contextMenuAction("archive", "Archive", "", outlined.Archive()))
		}
		return append(items, remove)
	case "dialog":
		return []overlay.MenuItem{copy, cut, paste, contextMenuSeparator(), {Key: "more-options", Label: "More Options", Children: contextMenuTools(w, false)}, contextMenuSeparator(), remove}
	case "inset":
		copy.Group, cut.Group = "Actions", "Actions"
		paste.Group, paste.Icon, paste.Inset = "Actions", nil, true
		items := []overlay.MenuItem{copy, cut, paste, contextMenuSeparator(),
			{Key: "bookmarks", Label: "Bookmarks", Group: "Appearance", GroupInset: true, Inset: true, Checked: &d.bookmarks}, {Key: "urls", Label: "Full URLs", Group: "Appearance", Inset: true, Checked: &d.urls}, contextMenuSeparator()}
		radio := themes()
		for i := range radio {
			radio[i].Inset = true
		}
		radio[0].GroupInset = true
		items = append(items, radio...)
		return append(items, contextMenuSeparator(), overlay.MenuItem{Key: "more-options", Label: "More Options", Inset: true, Children: contextMenuTools(w, false)[:2]})
	}
	return nil
}

func contextMenuWords(language string) map[string]string {
	switch language {
	case "ar":
		return map[string]string{"rightClick": "انقر بزر الماوس الأيمن هنا", "longPress": "اضغط مطولاً هنا", "navigation": "التنقل", "back": "رجوع", "forward": "تقدم", "reload": "إعادة تحميل", "moreTools": "المزيد من الأدوات", "savePage": "حفظ الصفحة...", "createShortcut": "إنشاء اختصار...", "nameWindow": "تسمية النافذة...", "developerTools": "أدوات المطور", "delete": "حذف", "showBookmarks": "إظهار الإشارات المرجعية", "showFullUrls": "إظهار عناوين URL الكاملة", "people": "الأشخاص"}
	case "he":
		return map[string]string{"rightClick": "לחץ לחיצה ימנית כאן", "longPress": "לחץ לחיצה ארוכה כאן", "navigation": "ניווט", "back": "חזור", "forward": "קדימה", "reload": "רענן", "moreTools": "כלים נוספים", "savePage": "שמור עמוד...", "createShortcut": "צור קיצור דרך...", "nameWindow": "שם חלון...", "developerTools": "כלי מפתח", "delete": "מחק", "showBookmarks": "הצג סימניות", "showFullUrls": "הצג כתובות URL מלאות", "people": "אנשים"}
	default:
		return map[string]string{"rightClick": "Right click here", "longPress": "Long press here", "navigation": "Navigation", "back": "Back", "forward": "Forward", "reload": "Reload", "moreTools": "More Tools", "savePage": "Save Page...", "createShortcut": "Create Shortcut...", "nameWindow": "Name Window...", "developerTools": "Developer Tools", "delete": "Delete", "showBookmarks": "Show Bookmarks", "showFullUrls": "Show Full URLs", "people": "People"}
	}
}
