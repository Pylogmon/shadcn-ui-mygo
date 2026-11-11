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

// This is an application model: drafts, checked values, public popup state and
// handles. Contexts, Elements and child view closures belong to the current build.
type menubarCompositionPreviewState struct {
	ready, native                           bool
	recipe, language, sideDirection         string
	lastRecipe, lastLanguage, lastDirection string
	checks                                  map[string]*bool
	selections                              map[string]string
	bars                                    map[string]*desktop.MenubarState
	nativeIcons                             map[*ui.SVG]ui.MenuIcon
	dialog                                  bool
	dialogReturn                            ui.Handle
	lastAction, menuError                   string
	commandCount                            int
}

var menubarCompositionRecipes = []control.Option{
	{Value: "demo", Label: "File, Edit, View and Profiles"},
	{Value: "basic", Label: "Basic"},
	{Value: "submenu", Label: "With Submenu"},
	{Value: "sides", Label: "Six sides"},
	{Value: "checkbox", Label: "With Checkboxes"},
	{Value: "radio", Label: "With Radio"},
	{Value: "icons", Label: "With Icons (docs)"},
	{Value: "create-icons", Label: "With Icons (Create)"},
	{Value: "shortcuts", Label: "With Shortcuts"},
	{Value: "format", Label: "Format"},
	{Value: "insert", Label: "Insert"},
	{Value: "destructive", Label: "Destructive"},
	{Value: "dialog", Label: "In Dialog"},
	{Value: "inset", Label: "With Inset"},
	{Value: "rtl", Label: "Translated File, Edit, View and Profiles"},
}

func (p *menubarCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native = true, native
	p.recipe, p.language, p.sideDirection = "demo", "ar", "ltr"
	p.lastRecipe, p.lastLanguage, p.lastDirection = p.recipe, p.language, p.sideDirection
	p.checks = map[string]*bool{}
	p.selections = map[string]string{}
	p.bars = map[string]*desktop.MenubarState{}
	p.nativeIcons = map[*ui.SVG]ui.MenuIcon{}
}

func (p *menubarCompositionPreviewState) close(h *desktop.Host) {
	for key, bar := range p.bars {
		bar.Open = false
		if h != nil {
			h.Close("source-menubar:" + key)
		}
	}
}

func (p *menubarCompositionPreviewState) sync(h *desktop.Host, native bool) {
	if p.native != native || p.lastRecipe != p.recipe || p.lastLanguage != p.language || p.lastDirection != p.sideDirection {
		p.close(h)
		p.dialog, p.menuError = false, ""
	}
	if p.native != native && h != nil {
		h.Close("source-menubar-config")
		h.Close("source-menubar-language")
		h.Close("source-menubar-direction")
	}
	p.native, p.lastRecipe, p.lastLanguage, p.lastDirection = native, p.recipe, p.language, p.sideDirection
}

func (p *menubarCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.menuError = err.Error()
	}
}

func (p *menubarCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	p.sync(h, native)
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: tr(c, "Menubar compositions")}, display.CardSlots{
		Title: tr(c, "Menubar compositions"), Description: tr(c, "Source commands, checked options and nested menus."),
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-menubar-config", &p.recipe, control.SelectProps{Label: tr(child, "Menubar configuration"), Options: localizedOptions(child, menubarCompositionRecipes)})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-menubar-language", &p.language, control.SelectProps{Label: tr(child, "Menubar language"), Options: localizedOptions(child, statusLanguages())})
			}
			if p.recipe == "sides" {
				p.choose(child, h, native, "source-menubar-direction", &p.sideDirection, control.SelectProps{Label: tr(child, "Menubar direction"), Options: localizedOptions(child, []control.Option{{Value: "ltr", Label: "Left to right"}, {Value: "rtl", Label: "Right to left"}})})
			}
			p.sync(h, native)
			ui.Column(child.Key("source:" + p.recipe)).FillWidth().MinWidth(0).Gap(12).Children(func() {
				p.draw(child, h, native)
			})
			last := "No command selected"
			if p.lastAction != "" {
				last = "Last command: " + p.lastAction
			}
			ui.Text(child, last).FontSize(13).TextColor(theme.TokensOf(child).MutedForeground).Role(ui.RoleStatus).Label(last)
			if p.menuError != "" {
				ui.Text(child, p.menuError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
}

func (p *menubarCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	if p.recipe == "dialog" {
		if control.Button(c, tr(c, "Open Dialog"), control.ButtonProps{Variant: control.Outline}).Bind(&p.dialogReturn).Clicked() {
			p.dialog = true
		}
		overlay.Dialog(c, &p.dialog, overlay.DialogProps{
			Title: tr(c, "Menubar Example"), Description: tr(c, "Use the menubar below to see the menu options."), Width: 512,
			ReturnFocus: &p.dialogReturn, OnOpenChange: func(open bool) {
				if !open {
					p.close(h)
				}
			},
		}, func() { p.bar(c, h, native, "dialog", p.items("dialog"), mygo.PlacementBottomStart, 224) })
		return
	}
	if p.recipe == "sides" {
		direction := ui.LeftToRight
		if p.sideDirection == "rtl" {
			direction = ui.RightToLeft
		}
		ui.Row(c).FillWidth().Direction(direction).Wrap().Justify(ui.Center).Gap(8).Children(func() {
			for _, side := range []string{"inline-start", "left", "top", "bottom", "right", "inline-end"} {
				label := map[string]string{"inline-start": "Inline Start", "left": "Left", "top": "Top", "bottom": "Bottom", "right": "Right", "inline-end": "Inline End"}[side]
				p.bar(c, h, native, "side:"+side, []overlay.MenubarItem{{Key: side, Label: tr(c, label), Items: []overlay.MenuItem{menubarAction("New Tab", ""), menubarAction("New Window", ""), menubarAction("New Incognito Window", "")}}}, popupPlacement(side, direction), 192)
			}
		})
		return
	}
	build := func() {
		width := float32(224)
		placement := mygo.PlacementBottomStart
		if c.Direction() == ui.RightToLeft {
			placement = mygo.PlacementBottomEnd
		}
		p.bar(c, h, native, p.recipe, p.items(p.recipe), placement, width)
	}
	if p.recipe == "rtl" {
		ui.Column(c).FillWidth().MinWidth(0).Direction(childDirection(p.language)).Children(build)
	} else {
		build()
	}
}

func (p *menubarCompositionPreviewState) withNativeIcons(items []overlay.MenubarItem) []overlay.MenubarItem {
	var visit func([]overlay.MenuItem)
	visit = func(items []overlay.MenuItem) {
		for i := range items {
			item := &items[i]
			if item.Icon != nil {
				icon := p.nativeIcons[item.Icon]
				if icon == nil {
					created, err := mygo.NewMenuIconSVG(item.Icon, true)
					if err != nil {
						p.menuError = err.Error()
					} else {
						icon = created
						p.nativeIcons[item.Icon] = icon
					}
				}
				item.NativeIcon = icon
			}
			visit(item.Children)
		}
	}
	for i := range items {
		visit(items[i].Items)
	}
	return items
}

func (p *menubarCompositionPreviewState) bar(c *ui.Context, h *desktop.Host, native bool, key string, items []overlay.MenubarItem, placement mygo.Placement, width float32) {
	items = localizedMenubarItems(c, items)
	if native {
		chosen, err := desktop.NativeMenubar(c.Key(key), p.withNativeIcons(items))
		if err != nil {
			p.menuError = err.Error()
		}
		if chosen != "" {
			p.dispatch(items, chosen)
		}
		return
	}
	bar := p.bars[key]
	if bar == nil {
		bar = &desktop.MenubarState{}
		p.bars[key] = bar
	}
	_, err := desktop.Menubar(c, h, "source-menubar:"+key, bar, desktop.MenubarProps{
		Items: items, Label: tr(c, "Source menubar"), Placement: placement, Width: width, MaxHeight: 480, Gap: 4,
		Revision: uint64(p.commandCount), OnChoose: func(value string) { p.dispatch(items, value) },
	})
	if err != nil {
		p.menuError = err.Error()
	}
}

func (p *menubarCompositionPreviewState) dispatch(items []overlay.MenubarItem, chosen string) {
	if strings.HasPrefix(chosen, "radio:") {
		parts := strings.Split(chosen, ":")
		if len(parts) == 4 {
			p.selections[parts[1]+":"+parts[2]] = parts[3]
		}
	}
	var find func([]overlay.MenuItem) string
	find = func(items []overlay.MenuItem) string {
		for _, item := range items {
			if item.Key == chosen && !item.Disabled && !item.Separator && len(item.Children) == 0 {
				return item.Label
			}
			if label := find(item.Children); label != "" {
				return label
			}
		}
		return ""
	}
	for _, item := range items {
		if label := find(item.Items); label != "" {
			p.lastAction = label
			p.commandCount++
			return
		}
	}
}

func menubarAction(label, shortcut string) overlay.MenuItem {
	return overlay.MenuItem{Key: label, Label: label, Shortcut: shortcut}
}

func menubarIcon(label, shortcut string, icon *ui.SVG) overlay.MenuItem {
	item := menubarAction(label, shortcut)
	item.Icon = icon
	return item
}

func menubarSub(label string, items ...overlay.MenuItem) overlay.MenuItem {
	item := menubarAction(label, "")
	item.Children = items
	return item
}

func menubarInset(item overlay.MenuItem) overlay.MenuItem    { item.Inset = true; return item }
func menubarDisabled(item overlay.MenuItem) overlay.MenuItem { item.Disabled = true; return item }
func menubarDanger(item overlay.MenuItem) overlay.MenuItem   { item.Destructive = true; return item }
func menubarSeparator() overlay.MenuItem                     { return overlay.MenuItem{Separator: true} }
func menubarTop(label string, items ...overlay.MenuItem) overlay.MenubarItem {
	return overlay.MenubarItem{Key: label, Label: label, Items: items}
}

func menubarTopWidth(width float32, label string, items ...overlay.MenuItem) overlay.MenubarItem {
	item := menubarTop(label, items...)
	item.Width = width
	return item
}

func (p *menubarCompositionPreviewState) check(recipe, name, label string, initial bool) overlay.MenuItem {
	key := recipe + ":" + name
	value := p.checks[key]
	if value == nil {
		v := initial
		value, p.checks[key] = &v, &v
	}
	return overlay.MenuItem{Key: "check:" + key, Label: label, Checked: value}
}

func (p *menubarCompositionPreviewState) radios(recipe, group, initial string, values, labels []string, inset bool) []overlay.MenuItem {
	key := recipe + ":" + group
	selected, ok := p.selections[key]
	if !ok {
		selected, p.selections[key] = initial, initial
	}
	items := make([]overlay.MenuItem, len(values))
	for i, value := range values {
		items[i] = overlay.MenuItem{Key: "radio:" + key + ":" + value, Label: labels[i], Radio: true, Selected: selected == value, Inset: inset}
	}
	return items
}

func (p *menubarCompositionPreviewState) items(recipe string) []overlay.MenubarItem {
	sep := menubarSeparator
	a := menubarAction
	sub := menubarSub
	file := []overlay.MenuItem{a("New Tab", "⌘T"), a("New Window", "⌘N"), menubarDisabled(a("New Incognito Window", "")), sep(), a("Print...", "⌘P")}
	edit := []overlay.MenuItem{a("Undo", "⌘Z"), a("Redo", "⇧⌘Z"), sep(), a("Cut", ""), a("Copy", ""), a("Paste", "")}
	switch recipe {
	case "basic":
		return []overlay.MenubarItem{menubarTop("File", file...), menubarTop("Edit", edit...)}
	case "submenu":
		return []overlay.MenubarItem{
			menubarTop("File", sub("Share", a("Email link", ""), a("Messages", ""), a("Notes", "")), sep(), a("Print...", "⌘P")),
			menubarTop("Edit", a("Undo", "⌘Z"), a("Redo", "⇧⌘Z"), sep(), sub("Find", a("Find...", ""), a("Find Next", ""), a("Find Previous", "")), sep(), a("Cut", ""), a("Copy", ""), a("Paste", "")),
		}
	case "checkbox":
		return []overlay.MenubarItem{
			menubarTopWidth(256, "View", p.check(recipe, "bookmarks", "Always Show Bookmarks Bar", false), p.check(recipe, "urls", "Always Show Full URLs", true), sep(), menubarInset(a("Reload", "⌘R")), menubarDisabled(menubarInset(a("Force Reload", "⇧⌘R")))),
			menubarTop("Format", p.check(recipe, "strikethrough", "Strikethrough", true), p.check(recipe, "code", "Code", false), p.check(recipe, "superscript", "Superscript", false)),
		}
	case "radio":
		profiles := p.radios(recipe, "profile", "benoit", []string{"andy", "benoit", "luis"}, []string{"Andy", "Benoit", "Luis"}, false)
		profiles = append(profiles, sep(), menubarInset(a("Edit...", "")), menubarInset(a("Add Profile...", "")))
		return []overlay.MenubarItem{menubarTop("Profiles", profiles...), menubarTop("Theme", p.radios(recipe, "theme", "system", []string{"light", "dark", "system"}, []string{"Light", "Dark", "System"}, false)...)}
	case "icons", "create-icons":
		settings, help, trash := outlined.Settings(), outlined.Help(), outlined.Delete()
		if recipe == "create-icons" {
			settings, help, trash = outlined.ProgressActivity(), outlined.ProgressActivity(), outlined.ProgressActivity()
		}
		return []overlay.MenubarItem{
			menubarTop("File", menubarIcon("New File", "⌘N", outlined.Draft()), menubarIcon("Open Folder", "", outlined.Folder()), sep(), menubarIcon("Save", "⌘S", outlined.Save())),
			menubarTop("More", menubarIcon("Settings", "", settings), menubarIcon("Help", "", help), sep(), menubarDanger(menubarIcon("Delete", "", trash))),
		}
	case "shortcuts":
		return []overlay.MenubarItem{menubarTop("File", a("New Tab", "⌘T"), a("New Window", "⌘N"), a("Print...", "⌘P")), menubarTop("Edit", a("Undo", "⌘Z"), a("Redo", "⇧⌘Z"), sep(), a("Cut", "⌘X"), a("Copy", "⌘C"), a("Paste", "⌘V"))}
	case "format":
		return []overlay.MenubarItem{
			menubarTop("Format", menubarIcon("Bold", "⌘B", outlined.FormatBold()), menubarIcon("Italic", "⌘I", outlined.FormatItalic()), menubarIcon("Underline", "⌘U", outlined.FormatUnderlined()), sep(), p.check(recipe, "strikethrough", "Strikethrough", true), p.check(recipe, "code", "Code", false)),
			menubarTop("View", p.check(recipe, "ruler", "Show Ruler", false), p.check(recipe, "grid", "Show Grid", true), sep(), menubarInset(a("Zoom In", "")), menubarInset(a("Zoom Out", ""))),
		}
	case "insert":
		media := sub("Media", a("Image", ""), a("Video", ""), a("Audio", ""))
		media.Icon = outlined.Image()
		return []overlay.MenubarItem{menubarTop("Insert", media, sep(), menubarIcon("Link", "⌘K", outlined.Link()), menubarIcon("Table", "", outlined.Table())), menubarTopWidth(176, "Tools", menubarIcon("Find & Replace", "⌘F", outlined.Search()), menubarIcon("Spell Check", "", outlined.Check()))}
	case "destructive":
		return []overlay.MenubarItem{
			menubarTopWidth(160, "File", menubarIcon("New File", "⌘N", outlined.Draft()), menubarIcon("Open Folder", "", outlined.Folder()), sep(), menubarDanger(menubarIcon("Delete File", "⌘⌫", outlined.Delete()))),
			menubarTop("Account", menubarIcon("Profile", "", outlined.Person()), menubarIcon("Settings", "", outlined.Settings()), sep(), menubarDanger(menubarIcon("Sign out", "", outlined.Logout())), sep(), menubarDanger(menubarIcon("Delete", "", outlined.Delete()))),
		}
	case "dialog":
		return []overlay.MenubarItem{
			menubarTop("File", menubarIcon("Copy", "", outlined.ContentCopy()), menubarIcon("Cut", "", outlined.ContentCut()), menubarIcon("Paste", "", outlined.ContentPaste()), sep(), sub("More Options", a("Save Page...", ""), a("Create Shortcut...", ""), a("Name Window...", ""), sep(), a("Developer Tools", "")), sep(), menubarDanger(menubarIcon("Delete", "", outlined.Delete()))),
			menubarTop("Edit", a("Undo", "⌘Z"), a("Redo", "⇧⌘Z")),
		}
	case "inset":
		copy := menubarIcon("Copy", "", outlined.ContentCopy())
		copy.Group = "Actions"
		bookmarks := menubarInset(p.check(recipe, "bookmarks", "Bookmarks", true))
		bookmarks.Group, bookmarks.GroupInset = "Appearance", true
		themes := p.radios(recipe, "theme", "system", []string{"light", "dark", "system"}, []string{"Light", "Dark", "System"}, true)
		// Group also scopes same-frame radio updates. Supply it on every row;
		// MenuContent emits the shared caption only once.
		for i := range themes {
			themes[i].Group = "Theme"
		}
		themes[0].GroupInset = true
		items := []overlay.MenuItem{copy, menubarIcon("Cut", "", outlined.ContentCut()), menubarInset(a("Paste", "")), sep(), bookmarks, menubarInset(p.check(recipe, "urls", "Full URLs", false)), sep()}
		items = append(items, themes...)
		items = append(items, sep(), menubarInset(sub("More Options", a("Save Page...", ""), a("Create Shortcut...", ""))))
		return []overlay.MenubarItem{menubarTopWidth(176, "View", items...)}
	default:
		return p.demoItems(recipe)
	}
}

func (p *menubarCompositionPreviewState) demoItems(recipe string) []overlay.MenubarItem {
	a, sep := menubarAction, menubarSeparator
	words := menubarWords["en"]
	if recipe == "rtl" {
		if translated, ok := menubarWords[p.language]; ok {
			words = translated
		}
	}
	w := func(key string) string { return words[key] }
	return []overlay.MenubarItem{
		menubarTop(w("file"), a(w("newTab"), "⌘T"), a(w("newWindow"), "⌘N"), menubarDisabled(a(w("incognito"), "")), sep(), menubarSub(w("share"), a(w("email"), ""), a(w("messages"), ""), a(w("notes"), "")), sep(), a(w("print"), "⌘P")),
		menubarTop(w("edit"), a(w("undo"), "⌘Z"), a(w("redo"), "⇧⌘Z"), sep(), menubarSub(w("find"), a(w("web"), ""), sep(), a(w("findItem"), ""), a(w("next"), ""), a(w("previous"), "")), sep(), a(w("cut"), ""), a(w("copy"), ""), a(w("paste"), "")),
		menubarTopWidth(176, w("view"), p.check(recipe, "bookmarks", w("bookmarks"), false), p.check(recipe, "urls", w("urls"), true), sep(), menubarInset(a(w("reload"), "⌘R")), menubarDisabled(menubarInset(a(w("force"), "⇧⌘R"))), sep(), menubarInset(a(w("fullscreen"), "")), sep(), menubarInset(a(w("sidebar"), ""))),
		menubarTop(w("profiles"), append(p.radios(recipe, "profile", "benoit", []string{"andy", "benoit", "Luis"}, []string{"Andy", "Benoit", "Luis"}, false), sep(), menubarInset(a(w("editProfile"), "")), sep(), menubarInset(a(w("addProfile"), "")))...),
	}
}

// Verbatim labels from menubar-rtl.tsx; English also drives the docs demo.
var menubarWords = map[string]map[string]string{
	"en": {"file": "File", "newTab": "New Tab", "newWindow": "New Window", "incognito": "New Incognito Window", "share": "Share", "email": "Email link", "messages": "Messages", "notes": "Notes", "print": "Print...", "edit": "Edit", "undo": "Undo", "redo": "Redo", "find": "Find", "web": "Search the web", "findItem": "Find...", "next": "Find Next", "previous": "Find Previous", "cut": "Cut", "copy": "Copy", "paste": "Paste", "view": "View", "bookmarks": "Bookmarks Bar", "urls": "Full URLs", "reload": "Reload", "force": "Force Reload", "fullscreen": "Toggle Fullscreen", "sidebar": "Hide Sidebar", "profiles": "Profiles", "editProfile": "Edit...", "addProfile": "Add Profile..."},
	"ar": {"file": "ملف", "newTab": "علامة تبويب جديدة", "newWindow": "نافذة جديدة", "incognito": "نافذة التصفح المتخفي الجديدة", "share": "مشاركة", "email": "رابط البريد الإلكتروني", "messages": "الرسائل", "notes": "الملاحظات", "print": "طباعة...", "edit": "تعديل", "undo": "تراجع", "redo": "إعادة", "find": "بحث", "web": "البحث على الويب", "findItem": "بحث...", "next": "البحث التالي", "previous": "البحث السابق", "cut": "قص", "copy": "نسخ", "paste": "لصق", "view": "عرض", "bookmarks": "شريط الإشارات المرجعية", "urls": "عناوين URL الكاملة", "reload": "إعادة تحميل", "force": "إعادة تحميل قسري", "fullscreen": "تبديل وضع ملء الشاشة", "sidebar": "إخفاء الشريط الجانبي", "profiles": "الملفات الشخصية", "editProfile": "تعديل...", "addProfile": "إضافة ملف شخصي..."},
	"he": {"file": "קובץ", "newTab": "כרטיסייה חדשה", "newWindow": "חלון חדש", "incognito": "חלון גלישה בסתר חדש", "share": "שתף", "email": "קישור אימייל", "messages": "הודעות", "notes": "הערות", "print": "הדפס...", "edit": "ערוך", "undo": "בטל", "redo": "בצע שוב", "find": "מצא", "web": "חפש באינטרנט", "findItem": "מצא...", "next": "מצא הבא", "previous": "מצא הקודם", "cut": "גזור", "copy": "העתק", "paste": "הדבק", "view": "תצוגה", "bookmarks": "סרגל סימניות", "urls": "כתובות URL מלאות", "reload": "רענן", "force": "רענן בכוח", "fullscreen": "החלף מסך מלא", "sidebar": "הסתר סרגל צד", "profiles": "פרופילים", "editProfile": "ערוך...", "addProfile": "הוסף פרופיל..."},
}
