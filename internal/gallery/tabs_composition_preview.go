package gallery

import (
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	nav "github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Only application values and the public menu state survive a view. Tabs slots
// always use their current synchronous Root; no Context, Element or callback is
// kept by this model.
type tabsCompositionPreviewState struct {
	ready, native                                    bool
	recipe, language, lastRecipe, action, popupError string
	drafts                                           map[string]*tabsCompositionDraft
	menuOpen                                         bool
	menu                                             desktop.MenuState
}
type tabsCompositionDraft struct {
	value, second, query, savedQuery, savedTab string
	changes, actions                           int
}
type tabsSourceRecipe struct {
	id, label                                   string
	values                                      []string
	line, vertical, icons, iconOnly, comparison bool
	disabled, content                           string
}

var tabsSourceRecipes = []tabsSourceRecipe{
	{id: "demo", label: "Dashboard cards (docs)", values: []string{"overview", "analytics", "reports", "settings"}, content: "cards"},
	{id: "usage", label: "Account and password (guide)", values: []string{"account", "password"}, content: "usage"},
	{id: "line", label: "Line (docs)", values: []string{"overview", "analytics", "reports"}, line: true},
	{id: "vertical", label: "Vertical triggers (docs)", values: []string{"account", "password", "notifications"}, vertical: true},
	{id: "disabled", label: "Disabled (docs)", values: []string{"home", "settings"}, disabled: "settings"},
	{id: "icons", label: "Icons (docs)", values: []string{"preview", "code"}, icons: true},
	{id: "rtl", label: "Translated dashboard", values: []string{"overview", "analytics", "reports", "settings"}, content: "cards"},
	{id: "create-basic", label: "Create: Basic", values: []string{"home", "settings"}},
	{id: "create-line", label: "Create: Line", values: []string{"overview", "analytics", "reports"}, line: true},
	{id: "create-comparison", label: "Create: Variants Alignment", values: []string{"overview", "analytics"}, comparison: true},
	{id: "create-disabled", label: "Create: Disabled", values: []string{"home", "settings"}, disabled: "settings"},
	{id: "create-icons", label: "Create: With Icons", values: []string{"preview", "code"}, icons: true},
	{id: "create-icon-only", label: "Create: Icon Only", values: []string{"home", "search", "settings"}, icons: true, iconOnly: true},
	{id: "create-multiple", label: "Create: Multiple", values: []string{"overview", "analytics", "reports", "settings"}},
	{id: "create-content", label: "Create: With Content", values: []string{"account", "password", "notifications"}, content: "account"},
	{id: "create-line-content", label: "Create: Line With Content", values: []string{"account", "password", "notifications"}, line: true, content: "account"},
	{id: "create-line-disabled", label: "Create: Line Disabled", values: []string{"overview", "analytics", "reports"}, line: true, disabled: "reports"},
	{id: "create-dropdown", label: "Create: With Dropdown", values: []string{"overview", "analytics", "reports"}, content: "dashboard"},
	{id: "create-vertical", label: "Create: Vertical", values: []string{"account", "password", "notifications"}, vertical: true, content: "vertical"},
	{id: "create-input", label: "Create: With Input and Button", values: []string{"overview", "analytics"}, content: "dashboard"},
}

func tabsCompositionOptions() []control.Option {
	options := make([]control.Option, len(tabsSourceRecipes))
	for i, r := range tabsSourceRecipes {
		options[i] = control.Option{Value: r.id, Label: r.label}
	}
	return options
}
func (p *tabsCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.lastRecipe = p.recipe
	p.drafts = map[string]*tabsCompositionDraft{}
}
func (p *tabsCompositionPreviewState) selectedRecipe() tabsSourceRecipe {
	for _, r := range tabsSourceRecipes {
		if r.id == p.recipe {
			return r
		}
	}
	return tabsSourceRecipes[0]
}
func (p *tabsCompositionPreviewState) draft(r tabsSourceRecipe) *tabsCompositionDraft {
	if d := p.drafts[r.id]; d != nil {
		return d
	}
	d := &tabsCompositionDraft{value: r.values[0], second: "overview"}
	p.drafts[r.id] = d
	return d
}
func (p *tabsCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}
func (p *tabsCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native || p.lastRecipe != p.recipe {
		if h != nil {
			h.Close("source-tabs-more")
			if p.native != native {
				h.Close("source-tabs-config")
				h.Close("source-tabs-language")
			}
		}
		p.menuOpen = false
		p.native, p.lastRecipe, p.popupError = native, p.recipe, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: tr(c, "Tabs compositions")}, display.CardSlots{Title: tr(c, "Tabs compositions"), Description: tr(c, "Source native tab lists, independent panels, dashboard cards and real header controls."), Content: func(child *ui.Context) {
		p.choose(child, h, native, "source-tabs-config", &p.recipe, control.SelectProps{Label: tr(child, "Tabs configuration"), Options: localizedOptions(child, tabsCompositionOptions())})
		if p.recipe == "rtl" {
			p.choose(child, h, native, "source-tabs-language", &p.language, control.SelectProps{Label: tr(child, "Tabs language"), Options: localizedOptions(child, statusLanguages())})
		}
		preview := ui.Column(child.Key("tabs-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
		if p.recipe == "rtl" {
			preview.Direction(childDirection(p.language))
		}
		preview.Children(func() { p.draw(child, h, native) })
		if p.action != "" {
			ui.Text(child, tr(child, p.action)).FontSize(13).Role(ui.RoleStatus)
		}
		if p.popupError != "" {
			ui.Text(child, p.popupError).Role(ui.RoleAlert)
		}
	}})
}
func tabsSourceLabel(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
func tabsSourceIcon(value string) *ui.SVG {
	switch value {
	case "home":
		return outlined.Home()
	case "search":
		return outlined.Search()
	case "settings":
		return outlined.Settings()
	case "code":
		return outlined.Code()
	default:
		return outlined.Window()
	}
}
func (p *tabsCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	r := p.selectedRecipe()
	d := p.draft(r)
	width := float32(400)
	if r.id == "rtl" {
		width = 384
	}
	if r.id == "create-input" {
		width = 512
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(width).AlignItems(ui.Center).Children(func() {
		if r.comparison {
			ui.ScrollHorizontal(c).FillWidth().MinWidth(0).Label(tr(c, "Tabs variant comparison viewport")).Children(func() {
				ui.Row(c).Gap(16).Children(func() {
					p.group(c, h, native, r, d, &d.value, "default", false)
					p.group(c, h, native, r, d, &d.second, "line", true)
				})
			})
		} else {
			p.group(c, h, native, r, d, &d.value, "main", r.line)
		}
	})
}
func (p *tabsCompositionPreviewState) group(c *ui.Context, h *desktop.Host, native bool, r tabsSourceRecipe, d *tabsCompositionDraft, value *string, key string, line bool) {
	root := nav.TabsRoot(c, value, nav.TabsRootProps{Key: r.id + ":" + key, Label: tr(c, "Source tabs ") + key, Vertical: r.vertical, OnValueChange: func(string) { d.changes++ }}, func(current *ui.Context) {
		list := func() {
			variant := nav.TabsDefault
			if line {
				variant = nav.TabsLine
			}
			nav.TabsList(current, nav.TabsListProps{Label: tr(current, "Source tabs ") + key, Variant: variant}, func(current *ui.Context) {
				for _, v := range r.values {
					label := tabsSourceLabel(v)
					if r.id == "rtl" {
						label = tabsSourceTranslations[p.language][v]
					}
					if v == "settings" && r.disabled == v {
						label = "Disabled"
					}
					props := nav.TabsTriggerProps{Label: tr(current, label), Disabled: r.disabled == v, HideLabel: r.iconOnly}
					if r.icons {
						props.Icon = tabsSourceIcon(v)
					}
					nav.TabsTrigger(current, v, props)
				}
			})
		}
		if r.id == "create-input" || r.id == "create-dropdown" {
			headerMinimum := float32(400)
			if r.id == "create-input" {
				headerMinimum = 512
			}
			ui.ScrollHorizontal(current).FillWidth().MinWidth(0).Label(tr(current, "Tabs header viewport")).Children(func() {
				ui.Row(current).FillWidth().MinWidth(headerMinimum).Gap(16).AlignItems(ui.Center).Children(func() {
					list()
					ui.Box(current).Grow(1)
					if r.id == "create-input" {
						ui.Row(current).Gap(8).AlignItems(ui.Center).Children(func() {
							control.Input(current, &d.query, control.InputProps{Label: tr(current, "Search tabs"), Placeholder: tr(current, "Search...")}).Width(176).Shrink(0)
							if control.Button(current, tr(current, "Action"), control.ButtonProps{}).Shrink(0).Clicked() {
								d.savedQuery, d.savedTab = d.query, *value
								d.actions++
								p.action = "Local tabs action: Action"
							}
						})
					} else {
						p.more(current, h, native, d)
					}
				})
			})
		} else if r.content != "" && !r.vertical {
			ui.ScrollHorizontal(current).FillWidth().MinWidth(0).Label(tr(current, "Source tab list viewport")).Children(list)
		} else {
			list()
		}
		if r.content != "" {
			panel := ui.Column(current).MinWidth(0)
			if r.vertical {
				panel.Basis(0).Grow(1)
			} else {
				panel.FillWidth()
			}
			if r.content != "cards" && r.content != "usage" {
				panel.Padding(24).Radius(theme.TokensOf(current).Radius).Border(1, theme.TokensOf(current).Border)
			}
			panel.Children(func() {
				values := r.values
				if r.id == "create-input" {
					values = []string{"overview", "analytics", "reports"}
				}
				for _, v := range values {
					nav.TabsContent(current, v, nav.TabsContentProps{}, func(current *ui.Context) { p.panel(current, r, v) })
				}
			})
		}
	})
	if r.content != "" {
		root.FillWidth()
	} else {
		root.AlignSelf(ui.Center)
	}
}
func (p *tabsCompositionPreviewState) panel(c *ui.Context, r tabsSourceRecipe, value string) {
	if r.content == "cards" {
		language := "en"
		if r.id == "rtl" {
			language = p.language
		}
		words := tabsSourceTranslations[language]
		display.CardWith(c, display.CardProps{}, display.CardSlots{Title: tr(c, words[value+"Title"]), Description: tr(c, words[value+"Desc"]), Content: func(current *ui.Context) {
			ui.Text(current, words[value+"Content"]).FontSize(14).TextColor(theme.TokensOf(current).MutedForeground)
		}})
		return
	}
	text := ""
	if r.content == "usage" {
		text = map[string]string{"account": "Make changes to your account here.", "password": "Change your password here."}[value]
	} else if r.content == "dashboard" {
		text = map[string]string{"overview": "View your dashboard metrics and key performance indicators.", "analytics": "Detailed analytics and insights about your data.", "reports": "Generate and view custom reports."}[value]
	} else {
		text = map[string]string{"account": "Manage your account preferences and profile information.", "password": "Update your password to keep your account secure.", "notifications": "Configure how you receive notifications and alerts."}[value]
		if r.content == "vertical" {
			if value == "password" {
				text += " Use a strong password with a mix of letters, numbers, and symbols."
			}
			if value == "notifications" {
				text += " Choose which types of notifications you want to receive and how you want to receive them."
			}
		}
	}
	ui.Text(c, tr(c, text)).FontSize(14)
}
func (p *tabsCompositionPreviewState) more(c *ui.Context, h *desktop.Host, native bool, d *tabsCompositionDraft) {
	items := []overlay.MenuItem{{Key: "settings", Label: tr(c, "Settings")}, {Key: "export", Label: tr(c, "Export")}, {Separator: true}, {Key: "archive", Label: tr(c, "Archive")}}
	decorate := func(e ui.Element) {
		control.StyleButton(c, e, "", control.ButtonProps{Variant: control.Ghost, Size: control.IconSmall, Label: tr(c, "More options"), Icon: outlined.MoreHoriz()}).Size(32, 32)
	}
	choice := ""
	var err error
	if native {
		choice, err = desktop.NativeMenu(c, overlay.MenuProps{Items: items}, decorate)
	} else {
		anchor := ui.ButtonBase(c)
		decorate(anchor)
		choice, err = desktop.DropdownMenu(c, h, "source-tabs-more", anchor, &p.menuOpen, &p.menu, desktop.MenuProps{Label: tr(c, "More options"), Items: items, Width: 192, Placement: mygo.PlacementBottomEnd, Gap: 4})
	}
	if err != nil {
		p.popupError = err.Error()
	}
	if choice != "" {
		d.actions++
		p.action = "Local tabs menu: " + tabsSourceLabel(choice)
	}
}

var tabsSourceTranslations = map[string]map[string]string{
	"en": {
		"overview":         "Overview",
		"analytics":        "Analytics",
		"reports":          "Reports",
		"settings":         "Settings",
		"overviewTitle":    "Overview",
		"overviewDesc":     "View your key metrics and recent project activity. Track progress across all your active projects.",
		"overviewContent":  "You have 12 active projects and 3 pending tasks.",
		"analyticsTitle":   "Analytics",
		"analyticsDesc":    "Track performance and user engagement metrics. Monitor trends and identify growth opportunities.",
		"analyticsContent": "Page views are up 25% compared to last month.",
		"reportsTitle":     "Reports",
		"reportsDesc":      "Generate and download your detailed reports. Export data in multiple formats for analysis.",
		"reportsContent":   "You have 5 reports ready and available to export.",
		"settingsTitle":    "Settings",
		"settingsDesc":     "Manage your account preferences and options. Customize your experience to fit your needs.",
		"settingsContent":  "Configure notifications, security, and themes.",
	},
	"ar": {
		"overview":         "نظرة عامة",
		"analytics":        "التحليلات",
		"reports":          "التقارير",
		"settings":         "الإعدادات",
		"overviewTitle":    "نظرة عامة",
		"overviewDesc":     "عرض مقاييسك الرئيسية وأنشطة المشروع الأخيرة. تتبع التقدم عبر جميع مشاريعك النشطة.",
		"overviewContent":  "لديك ١٢ مشروعًا نشطًا و٣ مهام معلقة.",
		"analyticsTitle":   "التحليلات",
		"analyticsDesc":    "تتبع مقاييس الأداء ومشاركة المستخدمين. راقب الاتجاهات وحدد فرص النمو.",
		"analyticsContent": "زادت مشاهدات الصفحة بنسبة ٢٥٪ مقارنة بالشهر الماضي.",
		"reportsTitle":     "التقارير",
		"reportsDesc":      "إنشاء وتنزيل تقاريرك التفصيلية. تصدير البيانات بتنسيقات متعددة للتحليل.",
		"reportsContent":   "لديك ٥ تقارير جاهزة ومتاحة للتصدير.",
		"settingsTitle":    "الإعدادات",
		"settingsDesc":     "إدارة تفضيلات حسابك وخياراته. تخصيص تجربتك لتناسب احتياجاتك.",
		"settingsContent":  "تكوين الإشعارات والأمان والسمات.",
	},
	"he": {
		"overview":         "סקירה כללית",
		"analytics":        "אנליטיקה",
		"reports":          "דוחות",
		"settings":         "הגדרות",
		"overviewTitle":    "סקירה כללית",
		"overviewDesc":     "הצג את המדדים העיקריים שלך ופעילות הפרויקט האחרונה. עקוב אחר התקדמות בכל הפרויקטים הפעילים שלך.",
		"overviewContent":  "יש לך 12 פרויקטים פעילים ו-3 משימות ממתינות.",
		"analyticsTitle":   "אנליטיקה",
		"analyticsDesc":    "עקוב אחר ביצועים ומדדי מעורבות משתמשים. עקוב אחר מגמות וזהה הזדמנויות צמיחה.",
		"analyticsContent": "צפיות בדף עלו ב-25% בהשוואה לחודש שעבר.",
		"reportsTitle":     "דוחות",
		"reportsDesc":      "צור והורד את הדוחות המפורטים שלך. ייצא נתונים בפורמטים מרובים לניתוח.",
		"reportsContent":   "יש לך 5 דוחות מוכנים וזמינים לייצוא.",
		"settingsTitle":    "הגדרות",
		"settingsDesc":     "נהל את העדפות החשבון והאפשרויות שלך. התאם אישית את החוויה שלך כך שתתאים לצרכים שלך.",
		"settingsContent":  "הגדר התראות, אבטחה וערכות נושא.",
	},
}
