package gallery

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Persistent values belong to the application. Rich slots and menu callbacks
// are created only in the current window build; none are kept in this model.
type statusCompositionPreviewState struct {
	ready, native, enabled, undone bool
	alertRecipe, emptyRecipe       string
	alertLanguage, emptyLanguage   string
	alertAction, emptyAction       string
	popupError                     string
	refreshes                      int
	queries                        map[string]*string
}

var statusAlertRecipes = []control.Option{
	{Value: "basic", Label: "Account updated"},
	{Value: "color", Label: "Amber subscription notice"},
	{Value: "enable", Label: "Enable dark mode action"},
	{Value: "docs-payment", Label: "Payment successful"},
	{Value: "docs-feature", Label: "New feature available"},
	{Value: "docs-failed", Label: "Payment failed"},
	{Value: "rtl", Label: "Translated RTL alerts"},
	{Value: "basic-title", Label: "Title only"},
	{Value: "basic-both", Label: "Title and description"},
	{Value: "basic-description", Label: "Description only"},
	{Value: "icon-title-link", Label: "Icon and title link"},
	{Value: "icon-description-links", Label: "Icon and two inline links"},
	{Value: "icon-both", Label: "Icon, title and description"},
	{Value: "long-title", Label: "Long title"},
	{Value: "long-description", Label: "Long description"},
	{Value: "long-both", Label: "Long title and description"},
	{Value: "session", Label: "Expired session"},
	{Value: "payment-list", Label: "Destructive payment checklist"},
	{Value: "undo", Label: "Undo action"},
	{Value: "badge", Label: "Badge action slot"},
}

var statusEmptyRecipes = []control.Option{
	{Value: "projects", Label: "No projects"},
	{Value: "avatar", Label: "Offline user"},
	{Value: "team", Label: "No team members"},
	{Value: "notifications", Label: "Muted notifications"},
	{Value: "cloud", Label: "Dashed cloud storage"},
	{Value: "search", Label: "Search input group"},
	{Value: "rtl", Label: "Translated RTL projects"},
	{Value: "create-projects", Label: "Create: Basic projects"},
	{Value: "no-results", Label: "Create: No search results"},
	{Value: "border-search", Label: "Create: Bordered search"},
	{Value: "posts", Label: "Create: No posts"},
	{Value: "muted-search", Label: "Create: Muted search"},
	{Value: "card-projects", Label: "Create: Empty in Card"},
}

func (p *statusCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native = true, native
	p.alertRecipe, p.emptyRecipe = "basic", "projects"
	p.alertLanguage, p.emptyLanguage = "ar", "ar"
	p.queries = map[string]*string{}
}

func (p *statusCompositionPreviewState) query(key string) *string {
	if q := p.queries[key]; q != nil {
		return q
	}
	q := ""
	p.queries[key] = &q
	return &q
}

func (p *statusCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
		return
	}
	if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *statusCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) (ui.Element, ui.Element) {
	p.initialize(native)
	if p.native != native {
		p.native, p.popupError = native, ""
		if h != nil {
			for _, key := range []string{"status-alert-recipe", "status-empty-recipe", "status-alert-language", "status-empty-language"} {
				h.Close(key)
			}
		}
	}
	alert := display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Alert compositions"}, display.CardSlots{
		Title: "Alert compositions", Description: "Source notices, inline links and independent actions.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "status-alert-recipe", &p.alertRecipe, control.SelectProps{Label: "Alert configuration", Options: statusAlertRecipes})
			if p.alertRecipe == "rtl" {
				p.choose(child, h, native, "status-alert-language", &p.alertLanguage, control.SelectProps{Label: "Alert language", Options: statusLanguages()})
			}
			ui.Column(child.Key("source-alert:" + p.alertRecipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Gap(16).Children(func() {
				if p.alertRecipe == "rtl" {
					ui.Column(child).FillWidth().MinWidth(0).Direction(childDirection(p.alertLanguage)).Gap(16).Children(func() { p.alert(child) })
				} else {
					p.alert(child)
				}
			})
			if p.alertAction != "" {
				ui.Text(child, p.alertAction).FontSize(13).Role(ui.RoleStatus)
			}
		},
	})
	empty := display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Empty compositions"}, display.CardSlots{
		Title: "Empty compositions", Description: "A useful next step, real search and local actions.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "status-empty-recipe", &p.emptyRecipe, control.SelectProps{Label: "Empty configuration", Options: statusEmptyRecipes})
			if p.emptyRecipe == "rtl" {
				p.choose(child, h, native, "status-empty-language", &p.emptyLanguage, control.SelectProps{Label: "Empty language", Options: statusLanguages()})
			}
			ui.Column(child.Key("source-empty:" + p.emptyRecipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() {
				build := func(current *ui.Context) { p.empty(current) }
				if p.emptyRecipe == "rtl" {
					ui.Column(child).FillWidth().MinWidth(0).Direction(childDirection(p.emptyLanguage)).Children(func() { build(child) })
				} else if p.emptyRecipe == "card-projects" {
					display.CardWith(child, display.CardProps{}, display.CardSlots{Content: build})
				} else {
					build(child)
				}
			})
			if p.emptyAction != "" {
				ui.Text(child, p.emptyAction).FontSize(13).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
	return alert, empty
}

func statusLanguages() []control.Option {
	return []control.Option{{Value: "ar", Label: "العربية"}, {Value: "he", Label: "עברית"}, {Value: "en", Label: "English"}}
}

func (p *statusCompositionPreviewState) inline(c *ui.Context, color ui.Color, parts ...statusInlinePart) ui.Element {
	return ui.RichText(c).FillWidth().MinWidth(0).FontSize(14).LineHeight(1.5).TextColor(color).Children(func() {
		for _, part := range parts {
			if part.action == "" {
				ui.Text(c, part.text)
			} else {
				if ui.Link(c, part.text, "").TextColor(color).Underline().UnderlineOffset(4).Clicked() {
					p.alertAction = "Local alert link: " + part.action
				}
			}
		}
	})
}

type statusInlinePart struct{ text, action string }

func (p *statusCompositionPreviewState) alert(c *ui.Context) {
	t := theme.TokensOf(c)
	props := display.AlertProps{}
	slots := display.AlertSlots{}
	width := float32(512)
	if !strings.Contains(p.alertRecipe, "-") && p.alertRecipe != "session" && p.alertRecipe != "undo" && p.alertRecipe != "badge" {
		width = 448
	}
	switch p.alertRecipe {
	case "basic":
		props.Title, props.Description, props.Icon = "Account updated successfully", "Your profile information has been saved. Changes will be reflected immediately.", outlined.CheckCircle()
	case "color":
		foreground, background, border := statusAmber(900), statusAmber(50), statusAmber(200)
		if t.Dark {
			foreground, background, border = statusAmber(50), statusAmber(950), statusAmber(900)
		}
		props.Title, props.Icon = "Your subscription will expire in 3 days.", outlined.Warning()
		slots.Description = func(child *ui.Context) {
			ui.Text(child, "Renew now to avoid service interruption or upgrade to a paid plan to continue using the service.").FontSize(14).LineHeight(1.5).TextColor(foreground)
		}
		display.AlertWith(c, props, slots).MaxWidth(width).Border(1, border).Background(background).TextColor(foreground)
		return
	case "enable":
		props.Title, props.Description = "Dark mode is now available", "Enable it under your profile settings to get started."
		slots.Action = func(child *ui.Context) {
			if control.Button(child, "Enable", control.ButtonProps{Size: control.ExtraSmall}).Clicked() {
				p.enabled = true
				p.alertAction = "Dark mode preference enabled locally"
			}
		}
	case "docs-payment", "docs-feature":
		payment, paymentBody, feature, featureBody := statusAlertWords("en")
		props.Title, props.Description, props.Icon = payment, paymentBody, outlined.CheckCircle()
		if p.alertRecipe == "docs-feature" {
			props.Title, props.Description, props.Icon = feature, featureBody, outlined.Info()
		}
		width = 448
	case "docs-failed":
		props.Title, props.Description, props.Icon, props.Destructive = "Payment failed", "Your payment could not be processed. Please check your payment method and try again.", outlined.Error(), true
		width = 448
	case "rtl":
		payment, paymentBody, feature, featureBody := statusAlertWords(p.alertLanguage)
		display.Alert(c.Key("payment"), display.AlertProps{Title: payment, Description: paymentBody, Icon: outlined.CheckCircle()}).MaxWidth(448)
		display.Alert(c.Key("feature"), display.AlertProps{Title: feature, Description: featureBody, Icon: outlined.Info()}).MaxWidth(448)
		return
	case "basic-title", "basic-both":
		props.Title = "Success! Your changes have been saved."
		if p.alertRecipe == "basic-both" {
			props.Description = "This is an alert with title and description."
		}
	case "basic-description":
		props.Description = "This one has a description only. No title. No icon."
	case "icon-title-link":
		props.Icon = outlined.Error()
		slots.Title = func(child *ui.Context) {
			p.inline(child, t.CardForeground, statusInlinePart{text: "Let's try one with icon, title and a "}, statusInlinePart{text: "link", action: "title"}, statusInlinePart{text: "."}).FontWeight(500)
		}
	case "icon-description-links":
		props.Icon = outlined.Error()
		slots.Description = func(child *ui.Context) {
			p.inline(child, t.MutedForeground, statusInlinePart{text: "This one has an icon and a description only. No title. "}, statusInlinePart{text: "But it has a link", action: "first"}, statusInlinePart{text: " and a "}, statusInlinePart{text: "second link", action: "second"}, statusInlinePart{text: "."})
		}
	case "icon-both":
		props.Icon, props.Title, props.Description = outlined.Error(), "Success! Your changes have been saved", "This is an alert with icon, title and description."
	case "long-title":
		props.Icon, props.Title = outlined.Error(), "This is a very long alert title that demonstrates how the component handles extended text content and potentially wraps across multiple lines"
	case "long-description":
		props.Icon, props.Description = outlined.Error(), "This is a very long alert description that demonstrates how the component handles extended text content and potentially wraps across multiple lines"
	case "long-both":
		props.Icon, props.Title = outlined.Error(), "This is an extremely long alert title that spans multiple lines to demonstrate how the component handles very lengthy headings while maintaining readability and proper text wrapping behavior"
		props.Description = "This is an equally long description that contains detailed information about the alert. It shows how the component can accommodate extensive content while preserving proper spacing, alignment, and readability across different screen sizes and viewport widths. This helps ensure the user experience remains consistent regardless of the content length."
	case "session":
		props.Icon, props.Title, props.Description, props.Destructive = outlined.Error(), "Something went wrong!", "Your session has expired. Please log in again.", true
	case "payment-list":
		props.Icon, props.Title, props.Destructive = outlined.Error(), "Unable to process your payment.", true
		slots.Description = func(child *ui.Context) {
			p.inline(child, t.Destructive.Alpha(.9), statusInlinePart{text: "Please verify your "}, statusInlinePart{text: "billing information", action: "billing"}, statusInlinePart{text: " and try again."})
			ui.Column(child).FillWidth().MinWidth(0).Role(ui.RoleList).Label("Payment recovery checklist").Children(func() {
				for i, text := range []string{"Check your card details", "Ensure sufficient funds", "Verify billing address"} {
					ui.Row(child.Key(i)).FillWidth().MinWidth(0).Gap(6).Role(ui.RoleListItem).Label(text).PositionInSet(i+1, 3).Children(func() {
						ui.Text(child, "•").AccessHidden(true)
						ui.Text(child, text).MinWidth(0).Grow(1).AccessHidden(true)
					})
				}
			})
		}
	case "undo", "badge":
		props.Icon, props.Title = outlined.Error(), "The selected emails have been marked as spam."
		if p.alertRecipe == "undo" {
			slots.Action = func(child *ui.Context) {
				if control.Button(child, "Undo", control.ButtonProps{Size: control.ExtraSmall}).Clicked() {
					p.undone, p.alertAction = true, "Local spam action undone"
				}
			}
		} else {
			props.Description = "This is a very long alert title that demonstrates how the component handles extended text content."
			slots.Action = func(child *ui.Context) { display.Badge(child, "Badge", display.BadgeSecondary) }
		}
	}
	display.AlertWith(c, props, slots).MaxWidth(width)
}

func statusAmber(shade int) ui.Color { color, _ := theme.Palette("amber", shade); return color }

func statusAlertWords(language string) (string, string, string, string) {
	switch language {
	case "ar":
		return "تم الدفع بنجاح", "تمت معالجة دفعتك البالغة 29.99 دولارًا. تم إرسال إيصال إلى عنوان بريدك الإلكتروني.", "ميزة جديدة متاحة", "لقد أضفنا دعم الوضع الداكن. يمكنك تفعيله في إعدادات حسابك."
	case "he":
		return "התשלום בוצע בהצלחה", "התשלום שלך בסך 29.99 דולר עובד. קבלה נשלחה לכתובת האימייל שלך.", "תכונה חדשה זמינה", "הוספנו תמיכה במצב כהה. אתה יכול להפעיל אותו בהגדרות החשבון שלך."
	default:
		return "Payment successful", "Your payment of $29.99 has been processed. A receipt has been sent to your email address.", "New feature available", "We've added dark mode support. You can enable it in your account settings."
	}
}

func (p *statusCompositionPreviewState) button(c *ui.Context, label string, variant control.Variant, size control.Size, icon *ui.SVG, link bool) ui.Element {
	props := control.ButtonProps{Variant: variant, Size: size, Icon: icon, Label: label}
	var e ui.Element
	if link {
		e = control.StyleButton(c, ui.Link(c, "", ""), label, props)
	} else {
		e = control.Button(c, label, props)
	}
	if e.Clicked() {
		p.emptyAction = "Local action: " + label
	}
	return e
}

func (p *statusCompositionPreviewState) learnMore(c *ui.Context, label string) {
	t := theme.TokensOf(c)
	// Native Link is text-only. An actual Link overlays inert visual content,
	// matching the kit's existing compound Badge/Bubble link contract.
	face := control.StyleButton(c, ui.Row(c), "", control.ButtonProps{Variant: control.Link, Size: control.Small}).TextColor(t.MutedForeground)
	var target ui.Element
	face.Children(func() {
		target = ui.Link(c.Key("learn-target"), "", "").Label(label).Absolute().Left(0).Top(0).Right(0).Bottom(0).FocusRing(false).Cursor(ui.CursorPointer)
		ui.Row(c.Key("learn-visual")).Gap(6).AlignItems(ui.Center).IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(func() {
			text := ui.Text(c, label).SingleLine()
			if target.Hovered() {
				text.Underline().UnderlineOffset(4)
			}
			icon := material.Icon(c, outlined.ArrowOutward(), material.Props{Size: 16})
			if p.emptyRecipe == "rtl" && c.Direction() == ui.RightToLeft {
				icon.Rotate(270)
			}
		})
	})
	if target.FocusVisible() {
		face.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
	}
	if target.Clicked() {
		p.emptyAction = "Local action: " + label
	}
}

func (p *statusCompositionPreviewState) empty(c *ui.Context) {
	t := theme.TokensOf(c)
	props := display.EmptyProps{Label: "Source empty"}
	slots := display.EmptySlots{}
	outsideMore, lowerCase, learnLabel := false, false, "Learn More"
	projectDescription := "You haven't created any projects yet. Get started by creating your first project."
	switch p.emptyRecipe {
	case "projects", "rtl", "create-projects", "card-projects":
		props.Title, props.Description = "No Projects Yet", projectDescription
		create, importProject := "Create Project", "Import Project"
		if p.emptyRecipe == "create-projects" || p.emptyRecipe == "card-projects" {
			props.Title, create, importProject, learnLabel, lowerCase = "No projects yet", "Create project", "Import project", "Learn more", true
		}
		if p.emptyRecipe != "create-projects" {
			props.Icon, props.MediaVariant = outlined.FolderCode(), display.EmptyMediaIcon
			if p.emptyRecipe == "card-projects" {
				props.Icon = outlined.Folder()
			}
		}
		if p.emptyRecipe == "rtl" {
			props.Title, props.Description, create, importProject, learnLabel = statusEmptyWords(p.emptyLanguage)
		}
		outsideMore = p.emptyRecipe == "projects" || p.emptyRecipe == "rtl"
		slots.Content = func(child *ui.Context) {
			ui.Row(child).Wrap().Gap(8).Justify(ui.Center).Children(func() {
				p.button(child, create, control.Default, control.DefaultSize, nil, lowerCase)
				p.button(child, importProject, control.Outline, control.DefaultSize, nil, false)
			})
			if !outsideMore {
				p.learnMore(child, learnLabel)
			}
		}
	case "avatar":
		props.Title, props.Description = "User Offline", "This user is currently offline. You can leave a message to notify them or try again later."
		slots.Media = func(child *ui.Context) {
			display.Avatar(child, display.AvatarProps{Image: statusGrayAvatars[0], Fallback: "LR", Label: "@shadcn", Size: 48})
		}
		slots.Content = func(child *ui.Context) { p.button(child, "Leave Message", control.Default, control.Small, nil, false) }
	case "team":
		props.Title, props.Description = "No Team Members", "Invite your team to collaborate on this project."
		slots.Media = func(child *ui.Context) {
			display.AvatarGroup(child, display.AvatarGroupProps{Label: "Team invitation avatars", Items: []display.AvatarProps{
				{Key: "shadcn", Image: statusGrayAvatars[0], Fallback: "CN", Label: "@shadcn", Size: 48},
				{Key: "maxleiter", Image: statusGrayAvatars[1], Fallback: "LR", Label: "@maxleiter", Size: 48},
				{Key: "evilrabbit", Image: statusGrayAvatars[2], Fallback: "ER", Label: "@evilrabbit", Size: 48},
			}})
		}
		slots.Content = func(child *ui.Context) {
			p.button(child, "Invite Members", control.Default, control.Small, outlined.Add(), false)
		}
	case "notifications":
		props.Title, props.Description, props.Icon, props.MediaVariant = "No Notifications", "You're all caught up. New notifications will appear here.", outlined.Notifications(), display.EmptyMediaIcon
		slots.Description = func(child *ui.Context) { display.EmptyDescription(child, props.Description).MaxWidth(320) }
		slots.Content = func(child *ui.Context) {
			if p.button(child, "Refresh", control.Outline, control.DefaultSize, outlined.Refresh(), false).Clicked() {
				p.refreshes++
				p.emptyAction = fmt.Sprintf("Local notification refresh: %d", p.refreshes)
			}
		}
	case "cloud":
		props.Title, props.Description, props.Icon, props.MediaVariant, props.Bordered = "Cloud Storage Empty", "Upload files to your cloud storage to access them anywhere.", outlined.Cloud(), display.EmptyMediaIcon, true
		slots.Content = func(child *ui.Context) { p.button(child, "Upload Files", control.Outline, control.Small, nil, false) }
	case "search", "border-search", "muted-search":
		props.Title, props.Description = "404 - Not Found", "The page you're looking for doesn't exist. Try searching for what you need below."
		slots.Content = func(child *ui.Context) {
			icon := outlined.ProgressActivity()
			if p.emptyRecipe == "search" {
				icon = outlined.Search()
			}
			q := p.query(p.emptyRecipe)
			group := control.InputGroup(child, control.InputGroupProps{Label: "Search missing pages", InlineStart: func() { material.Icon(child, icon, material.Props{Size: 16}) }, InlineEnd: func() { display.Kbd(child, "/") }}, func() ui.Element {
				return control.Input(child, q, control.InputProps{Label: "Search missing pages", Placeholder: "Try searching for pages...", OnSubmit: func(value string) { p.emptyAction = "Local page search: " + value }})
			})
			viewportWidth, _ := child.Size()
			if p.emptyRecipe != "search" || viewportWidth >= 640 {
				group.WidthPercent(75)
			}
			ui.RichText(child).FillWidth().FontSize(14).LineHeight(1.625).TextAlign(ui.Center).TextColor(t.MutedForeground).Children(func() {
				ui.Text(child, "Need help? ")
				if ui.Link(child, "Contact support", "").Underline().UnderlineOffset(4).TextColor(t.MutedForeground).Clicked() {
					p.emptyAction = "Local help action: Contact support"
				}
			})
		}
	case "no-results":
		props.Title, props.Description, props.Muted = "No results found", "No results found for your search. Try adjusting your search terms.", true
		slots.Content = func(child *ui.Context) {
			p.button(child, "Try again", control.Default, control.DefaultSize, nil, false)
			p.learnMore(child, "Learn more")
		}
	case "posts":
		props.Title, props.Icon, props.MediaVariant = "Nothing to see here", outlined.Folder(), display.EmptyMediaIcon
		slots.Description = func(child *ui.Context) {
			ui.RichText(child).FillWidth().FontSize(14).LineHeight(1.625).TextAlign(ui.Center).TextColor(t.MutedForeground).Children(func() {
				ui.Text(child, "No posts have been created yet. Get started by ")
				if ui.Link(child, "creating your first post", "").TextColor(t.MutedForeground).Underline().UnderlineOffset(4).Clicked() {
					p.emptyAction = "Local action: creating your first post"
				}
				ui.Text(child, ".")
			})
		}
		slots.Content = func(child *ui.Context) {
			p.button(child, "New Post", control.Outline, control.DefaultSize, outlined.Add(), false)
		}
	}
	root := display.EmptyWith(c, props, slots)
	switch p.emptyRecipe {
	case "notifications":
		root.Background(t.Muted.Alpha(.3))
	case "border-search", "posts":
		root.Border(1, t.Border).BorderStyle(ui.BorderSolid)
	case "muted-search":
		root.Background(t.Muted.Alpha(.5))
	}
	if outsideMore {
		root.Children(func() { p.learnMore(c, learnLabel) })
	}
}

func statusEmptyWords(language string) (string, string, string, string, string) {
	switch language {
	case "ar":
		return "لا توجد مشاريع بعد", "لم تقم بإنشاء أي مشاريع بعد. ابدأ بإنشاء مشروعك الأول.", "إنشاء مشروع", "استيراد مشروع", "تعرف على المزيد"
	case "he":
		return "אין פרויקטים עדיין", "עדיין לא יצרת פרויקטים. התחל על ידי יצירת הפרויקט הראשון שלך.", "צור פרויקט", "ייבא פרויקט", "למד עוד"
	default:
		return "No Projects Yet", "You haven't created any projects yet. Get started by creating your first project.", "Create Project", "Import Project", "Learn More"
	}
}

// Reuse the exact source-profile bytes already embedded for Card examples.
// Each bitmap is transformed once by the same application-owned pixel filter.
var statusGrayAvatars = []*ui.Bitmap{
	cardFilteredImage(cardDecodeImage(cardShadcnBytes), 1, nil),
	cardFilteredImage(cardDecodeImage(cardMaxleiterBytes), 1, nil),
	cardFilteredImage(cardDecodeImage(cardEvilrabbitBytes), 1, nil),
}
