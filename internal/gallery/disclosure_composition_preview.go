package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Only application values survive configuration/page changes. All native
// elements, disclosure scopes and builders belong to the current frame.
type disclosureCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	opens                                map[string]*bool
	radius                               [4]string
	radiusHandles                        [4]ui.Handle
	treeTab, selectedFile                string
}

var disclosureCompositionRecipes = []control.Option{
	{Value: "settings", Label: "Radius settings"},
	{Value: "tree", Label: "File tree"},
	{Value: "order", Label: "Order details"},
	{Value: "product", Label: "Product details"},
	{Value: "collapsible-rtl", Label: "Translated order"},
	{Value: "accordion-demo", Label: "Shipping questions"},
	{Value: "accordion-basic", Label: "Account questions"},
	{Value: "accordion-borders", Label: "Joined billing questions"},
	{Value: "accordion-multiple", Label: "Preference questions"},
	{Value: "accordion-disabled", Label: "Disabled account question"},
	{Value: "accordion-card", Label: "Subscription card"},
	{Value: "accordion-rtl", Label: "Translated account questions"},
	{Value: "create-basic", Label: "Create: Basic accordion"},
	{Value: "create-multiple", Label: "Create: Long multiple accordion"},
	{Value: "create-borders", Label: "Create: Separate borders"},
	{Value: "create-card", Label: "Create: Subscription card"},
	{Value: "create-disabled", Label: "Create: Disabled accordion"},
}

func (p *disclosureCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language, p.treeTab = true, native, "settings", "ar", "explorer"
	p.opens = map[string]*bool{}
	p.radius = [4]string{"0", "0", "0", "0"}
}

func (p *disclosureCompositionPreviewState) open(key string, initial bool) *bool {
	if v := p.opens[key]; v != nil {
		return v
	}
	v := initial
	p.opens[key] = &v
	return &v
}

func (p *disclosureCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) ui.Element {
	if native {
		return desktop.NativeSelect(c, value, props)
	}
	e, err := desktop.Select(c, h, key, value, props)
	if err != nil {
		p.popupError = err.Error()
	}
	return e
}

func (p *disclosureCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		p.native, p.popupError = native, ""
		if h != nil {
			h.Close("disclosure-composition-recipe")
			h.Close("disclosure-composition-language")
		}
	}
	display.CardWith(c, display.CardProps{Label: tr(c, "Disclosure compositions"), Size: display.CardSmall}, display.CardSlots{
		Title: tr(c, "Disclosure compositions"), Description: tr(c, "Independent triggers, persistent drafts and source question groups."),
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "disclosure-composition-recipe", &p.recipe, control.SelectProps{Label: tr(child, "Disclosure configuration"), Options: localizedOptions(child, disclosureCompositionRecipes)})
			if strings.HasSuffix(p.recipe, "-rtl") {
				p.choose(child, h, native, "disclosure-composition-language", &p.language, control.SelectProps{Label: tr(child, "Disclosure language"), Options: localizedOptions(child, []control.Option{{Value: "ar", Label: "العربية"}, {Value: "he", Label: "עברית"}, {Value: "en", Label: "English"}})})
			}
			ui.Column(child).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() {
				p.draw(child.Key("disclosure-source:" + p.recipe))
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

func (p *disclosureCompositionPreviewState) draw(c *ui.Context) {
	switch p.recipe {
	case "settings":
		p.settings(c)
	case "tree":
		p.fileTree(c)
	case "order", "collapsible-rtl":
		p.order(c)
	case "product":
		p.product(c)
	default:
		p.accordion(c)
	}
}

func disclosureIcon(c *ui.Context, icon *ui.SVG, size float32) ui.Element {
	return material.Icon(c, icon, material.Props{Size: size}).Shrink(0).AccessHidden(true).IgnoreInput(true)
}

func (p *disclosureCompositionPreviewState) settings(c *ui.Context) {
	open := p.open("settings", false)
	display.CardWith(c, display.CardProps{Label: tr(c, "Radius settings card"), Size: display.CardSmall}, display.CardSlots{
		Title: tr(c, "Radius"), Description: tr(c, "Set the corner radius of the element."),
		Content: func(child *ui.Context) {
			display.CollapsibleRoot(child, open, display.CollapsibleRootProps{Label: tr(child, "Radius settings"), KeepMounted: true}, func(s *display.CollapsibleScope) {
				ui.Column(child).Grid().Columns(2).Grow(1).MinWidth(0).Gap(8).Children(func() {
					for i, label := range []string{"Radius X", "Radius Y"} {
						control.Input(child.Key(label), &p.radius[i], control.InputProps{Label: tr(child, label), Placeholder: "0", Handle: &p.radiusHandles[i]}).FillWidth().MinWidth(0)
					}
					// This panel is declared before the independent trigger. It shares
					// the two-column input grid without turning an editor into a button.
					parts := s.Content(child, display.CollapsibleContentProps{Label: tr(child, "Additional radius values")}, func(panel *ui.Context) {
						for i, label := range []string{"Expanded Radius X", "Expanded Radius Y"} {
							control.Input(panel.Key(label), &p.radius[i+2], control.InputProps{Label: tr(panel, label), Placeholder: "0", Handle: &p.radiusHandles[i+2]}).FillWidth().MinWidth(0)
						}
					})
					parts.Root.ColumnSpan(2)
					parts.Content.Grid().Columns(2).Gap(8)
				})
				s.Trigger(child, display.CollapsibleTriggerProps{Label: tr(child, "Expand radius settings"), Style: func(tr ui.Element) {
					control.StyleButton(child, tr, "", control.ButtonProps{Variant: control.Outline, Size: control.Icon})
				}}, func(current *ui.Context) {
					icon := outlined.OpenInFull()
					if s.Open() {
						icon = outlined.CloseFullscreen()
					}
					disclosureIcon(current, icon, 16)
				})
			}).Row().AlignItems(ui.Start).Gap(8)
		},
	}).MaxWidth(320)
}

func (p *disclosureCompositionPreviewState) product(c *ui.Context) {
	open := p.open("product", false)
	display.CardWith(c, display.CardProps{Label: tr(c, "Product details card")}, display.CardSlots{Content: func(child *ui.Context) {
		t := theme.TokensOf(child)
		root := display.CollapsibleRoot(child, open, display.CollapsibleRootProps{Label: tr(child, "Product details")}, func(s *display.CollapsibleScope) {
			s.Trigger(child, display.CollapsibleTriggerProps{Label: tr(child, "Product details"), Style: func(tr ui.Element) {
				control.StyleButton(child, tr, "", control.ButtonProps{Variant: control.Ghost}).FillWidth().Justify(ui.Start)
			}}, func(current *ui.Context) {
				ui.Text(current, tr(current, "Product details")).Grow(1)
				disclosureIcon(current, outlined.KeyboardArrowDown(), 16).Rotate(180 * s.Progress())
			})
			parts := s.Content(child, display.CollapsibleContentProps{}, func(panel *ui.Context) {
				ui.Text(panel, tr(panel, "This panel can be expanded or collapsed to reveal additional content.")).FontSize(14).LineHeight(1.5)
				if control.Button(panel, tr(panel, "Learn More"), control.ButtonProps{Size: control.ExtraSmall}).Clicked() {
					p.action = "Local product details preview"
				}
			})
			parts.Content.Padding(0, 10, 10, 10).Gap(8).AlignItems(ui.Start)
		})
		root.Radius(t.RadiusMD())
		if *open {
			root.Background(t.Muted)
		}
	}}).MaxWidth(384)
}

type disclosureOrderWords struct{ order, status, shipped, shipping, address, items, description string }

func (p *disclosureCompositionPreviewState) order(c *ui.Context) {
	w := disclosureOrderWords{"Order #4189", "Status", "Shipped", "Shipping address", "100 Market St, San Francisco", "Items", "2x Studio Headphones"}
	direction := c.Direction()
	if p.recipe == "collapsible-rtl" {
		direction = childDirection(p.language)
		if p.language == "ar" {
			w = disclosureOrderWords{"الطلب #4189", "الحالة", "تم الشحن", "عنوان الشحن", "100 Market St, San Francisco", "العناصر", "2x سماعات الاستوديو"}
		} else if p.language == "he" {
			w = disclosureOrderWords{"הזמנה #4189", "סטטוס", "נשלח", "כתובת משלוח", "100 Market St, San Francisco", "פריטים", "2x אוזניות סטודיו"}
		}
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(350).Direction(direction).Children(func() {
		t := theme.TokensOf(c)
		display.CollapsibleRoot(c, p.open(p.recipe, false), display.CollapsibleRootProps{Label: tr(c, w.order)}, func(s *display.CollapsibleScope) {
			ui.Row(c).FillWidth().Justify(ui.SpaceBetween).Gap(16).PaddingX(16).AlignItems(ui.Center).Children(func() {
				ui.Text(c, w.order).FontSize(14).FontWeight(600)
				s.Trigger(c, display.CollapsibleTriggerProps{Label: tr(c, "Toggle order details"), Style: func(tr ui.Element) {
					control.StyleButton(c, tr, "", control.ButtonProps{Variant: control.Ghost, Size: control.IconSmall})
				}}, func(current *ui.Context) { disclosureIcon(current, outlined.UnfoldMore(), 16) })
			})
			ui.Row(c).FillWidth().Justify(ui.SpaceBetween).Gap(8).Radius(t.RadiusMD()).Border(1, t.Border).Padding(8, 16).Children(func() {
				ui.Text(c, tr(c, w.status)).FontSize(14).TextColor(t.MutedForeground)
				ui.Text(c, w.shipped).FontSize(14).FontWeight(500)
			})
			parts := s.Content(c, display.CollapsibleContentProps{Label: tr(c, "Order detail content")}, func(panel *ui.Context) {
				for _, row := range [][2]string{{w.shipping, w.address}, {w.items, w.description}} {
					ui.Column(panel).FillWidth().Radius(t.RadiusMD()).Border(1, t.Border).Padding(8, 16).Children(func() {
						ui.Text(panel, row[0]).FontSize(14).FontWeight(500)
						ui.Text(panel, row[1]).FontSize(14).TextColor(t.MutedForeground)
					})
				}
			})
			parts.Content.Gap(8)
		}).Gap(8)
	})
}

type disclosureTreeNode struct {
	name     string
	children []disclosureTreeNode
}

var disclosureFiles = []disclosureTreeNode{
	{"components", []disclosureTreeNode{{"ui", []disclosureTreeNode{{name: "button.tsx"}, {name: "card.tsx"}, {name: "dialog.tsx"}, {name: "input.tsx"}, {name: "select.tsx"}, {name: "table.tsx"}}}, {name: "login-form.tsx"}, {name: "register-form.tsx"}}},
	{"lib", []disclosureTreeNode{{name: "utils.ts"}, {name: "cn.ts"}, {name: "api.ts"}}},
	{"hooks", []disclosureTreeNode{{name: "use-media-query.ts"}, {name: "use-debounce.ts"}, {name: "use-local-storage.ts"}}},
	{"types", []disclosureTreeNode{{name: "index.d.ts"}, {name: "api.d.ts"}}},
	{"public", []disclosureTreeNode{{name: "favicon.ico"}, {name: "logo.svg"}, {name: "images"}}},
	{name: "app.tsx"}, {name: "layout.tsx"}, {name: "globals.css"}, {name: "package.json"}, {name: "tsconfig.json"}, {name: "README.md"}, {name: ".gitignore"},
}

func (p *disclosureCompositionPreviewState) fileTree(c *ui.Context) {
	display.CardWith(c, display.CardProps{Label: tr(c, "Source file explorer"), Size: display.CardSmall, Spacing: 16}, display.CardSlots{
		Header: func(current *ui.Context) {
			navigation.TabsRoot(current, &p.treeTab, navigation.TabsRootProps{Label: tr(current, "Explorer view")}, func(child *ui.Context) {
				navigation.TabsList(child, navigation.TabsListProps{}, func(list *ui.Context) {
					navigation.TabsTrigger(list, "explorer", navigation.TabsTriggerProps{Label: tr(list, "Explorer")}).Grow(1)
					navigation.TabsTrigger(list, "settings", navigation.TabsTriggerProps{Label: tr(list, "Outline")}).Grow(1)
				}).FillWidth()
			}).FillWidth()
		},
		Content: func(current *ui.Context) { p.tree(current, "", disclosureFiles) },
	}).MaxWidth(256).Gap(8)
}

func (p *disclosureCompositionPreviewState) tree(c *ui.Context, parent string, nodes []disclosureTreeNode) {
	ui.Column(c).FillWidth().MinWidth(0).Gap(4).Children(func() {
		for _, node := range nodes {
			path := node.name
			if parent != "" {
				path = parent + "/" + node.name
			}
			if node.children == nil {
				control.Button(c.Key(path), node.name, control.ButtonProps{Label: tr(c, "Open ") + path, Variant: control.Link, Size: control.Small, Icon: outlined.Description()}).FillWidth().Justify(ui.Start).Gap(8).TextColor(theme.TokensOf(c).Foreground).OnClick(func() {
					p.selectedFile, p.action = path, "Local file preview: "+path
				})
				continue
			}
			display.CollapsibleRoot(c.Key(path), p.open("tree:"+path, false), display.CollapsibleRootProps{Label: tr(c, "Folder ") + path}, func(s *display.CollapsibleScope) {
				s.Trigger(c, display.CollapsibleTriggerProps{Label: tr(c, "Folder ") + path, Style: func(tr ui.Element) {
					control.StyleButton(c, tr, "", control.ButtonProps{Variant: control.Ghost, Size: control.Small}).FillWidth().Justify(ui.Start).Gap(8)
				}}, func(current *ui.Context) {
					icon := disclosureIcon(current, outlined.ChevronRight(), 16)
					if current.Direction() == ui.RightToLeft {
						icon.Rotate(180 - 90*s.Progress())
					} else {
						icon.Rotate(90 * s.Progress())
					}
					disclosureIcon(current, outlined.Folder(), 16)
					ui.Text(current, node.name).MinWidth(0)
				})
				parts := s.Content(c, display.CollapsibleContentProps{}, func(current *ui.Context) { p.tree(current, path, node.children) })
				parts.Root.Margin(4, 0, 0, 0).PaddingStart(20)
			})
		}
	})
}

type disclosureQuestion struct {
	key, title                        string
	paragraphs                        []string
	disabled, annualLink, plansButton bool
}

func (p *disclosureCompositionPreviewState) accordion(c *ui.Context) {
	recipe := p.recipe
	questions := disclosureQuestions[recipe]
	direction := c.Direction()
	if recipe == "accordion-rtl" {
		questions = disclosureTranslations[p.language]
		direction = childDirection(p.language)
	}
	initial := map[string]string{"accordion-demo": "shipping", "accordion-basic": "item-1", "accordion-borders": "billing", "accordion-multiple": "notifications", "accordion-card": "plans", "accordion-rtl": "item-1", "create-card": "plans"}[recipe]
	width := float32(512)
	if recipe == "accordion-card" {
		width = 384
	} else if recipe == "accordion-rtl" {
		width = 448
	}
	build := func(current *ui.Context) {
		props := display.AccordionProps{Single: recipe != "accordion-multiple" && recipe != "create-multiple" && recipe != "create-card", Bordered: recipe == "create-borders", Loop: true}
		for _, question := range questions {
			props.Items = append(props.Items, display.AccordionItem{Key: question.key, Title: tr(current, question.title), Open: p.open(recipe+":"+question.key, question.key == initial), Disabled: question.disabled, Content: func() {
				ui.Column(current).FillWidth().Gap(16).Children(func() {
					for index, paragraph := range question.paragraphs {
						if question.annualLink && index == 1 {
							ui.RichText(current).FillWidth().FontSize(14).LineHeight(1.5).Children(func() {
								ui.Link(current, tr(current, "Annual billing is available"), "").OnClick(func() { p.action = "Local annual-billing information" })
								ui.Text(current, strings.TrimPrefix(paragraph, "Annual billing is available"))
							})
						} else {
							ui.Text(current, paragraph).FontSize(14).LineHeight(1.5)
						}
					}
					if question.plansButton {
						if control.Button(current, tr(current, "View plans"), control.ButtonProps{Size: control.Small, Icon: outlined.NorthEast()}).AlignSelf(ui.Start).Clicked() {
							p.action = "Local subscription-plan preview"
						}
					}
				})
			}})
		}
		t := theme.TokensOf(current)
		if recipe == "accordion-borders" || recipe == "create-disabled" {
			props.StyleItem = func(_ *ui.Context, parts display.AccordionItemParts) {
				parts.Trigger.PaddingX(16)
				if parts.Content.Valid() {
					parts.Content.PaddingX(16)
				}
				if recipe == "create-disabled" {
					parts.Root.Padding(4)
					if parts.Open {
						parts.Root.Background(t.Muted.Alpha(.5))
					}
				}
			}
		}
		root := display.Accordion(current, props).FillWidth().MinWidth(0)
		if recipe == "accordion-borders" || recipe == "create-disabled" {
			root.Border(1, t.Border).Radius(t.RadiusLG()).Clip()
		}
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(width).Direction(direction).Children(func() {
		if recipe == "accordion-card" || recipe == "create-card" {
			description := "Common questions about your account, plans, payments and cancellations."
			if recipe == "create-card" {
				description = "Common questions about your account, plans, and payments"
			}
			card := display.CardWith(c, display.CardProps{Label: tr(c, "Subscription & Billing")}, display.CardSlots{Title: tr(c, "Subscription & Billing"), Description: tr(c, description), Content: build})
			if recipe == "create-card" {
				card.Gap(16)
			}
		} else {
			build(c)
		}
	})
}

// Literal question content copied from the fixed originals in source-map.json.
var disclosureQuestions = map[string][]disclosureQuestion{
	"accordion-basic": {
		{key: "item-1", title: "How do I reset my password?", paragraphs: []string{"Click on 'Forgot Password' on the login page, enter your email address, and we'll send you a link to reset your password. The link will expire in 24 hours."}},
		{key: "item-2", title: "Can I change my subscription plan?", paragraphs: []string{"Yes, you can upgrade or downgrade your plan at any time from your account settings. Changes will be reflected in your next billing cycle."}},
		{key: "item-3", title: "What payment methods do you accept?", paragraphs: []string{"We accept all major credit cards, PayPal, and bank transfers. All payments are processed securely through our payment partners."}},
	},
	"accordion-borders": {
		{key: "billing", title: "How does billing work?", paragraphs: []string{"We offer monthly and annual subscription plans. Billing is charged at the beginning of each cycle, and you can cancel anytime. All plans include automatic backups, 24/7 support, and unlimited team members."}},
		{key: "security", title: "Is my data secure?", paragraphs: []string{"Yes. We use end-to-end encryption, SOC 2 Type II compliance, and regular third-party security audits. All data is encrypted at rest and in transit using industry-standard protocols."}},
		{key: "integration", title: "What integrations do you support?", paragraphs: []string{"We integrate with 500+ popular tools including Slack, Zapier, Salesforce, HubSpot, and more. You can also build custom integrations using our REST API and webhooks."}},
	},
	"accordion-multiple": {
		{key: "notifications", title: "Notification Settings", paragraphs: []string{"Manage how you receive notifications. You can enable email alerts for updates or push notifications for mobile devices."}},
		{key: "privacy", title: "Privacy & Security", paragraphs: []string{"Control your privacy settings and security preferences. Enable two-factor authentication, manage connected devices, review active sessions, and configure data sharing preferences. You can also download your data or delete your account."}},
		{key: "billing", title: "Billing & Subscription", paragraphs: []string{"View your current plan, payment history, and upcoming invoices. Update your payment method, change your subscription tier, or cancel your subscription."}},
	},
	"accordion-card": {
		{key: "plans", title: "What subscription plans do you offer?", paragraphs: []string{"We offer three subscription tiers: Starter ($9/month), Professional ($29/month), and Enterprise ($99/month). Each plan includes increasing storage limits, API access, priority support, and team collaboration features."}},
		{key: "billing", title: "How does billing work?", paragraphs: []string{"Billing occurs automatically at the start of each billing cycle. We accept all major credit cards, PayPal, and ACH transfers for enterprise customers. You'll receive an invoice via email after each payment."}},
		{key: "cancel", title: "How do I cancel my subscription?", paragraphs: []string{"You can cancel your subscription anytime from your account settings. There are no cancellation fees or penalties. Your access will continue until the end of your current billing period."}},
	},
	"create-basic": {
		{key: "item-1", title: "Is it accessible?", paragraphs: []string{"Yes. It adheres to the WAI-ARIA design pattern."}},
		{key: "item-2", title: "Is it styled?", paragraphs: []string{"Yes. It comes with default styles that matches the other components' aesthetic."}},
		{key: "item-3", title: "Is it animated?", paragraphs: []string{"Yes. It's animated by default, but you can disable it if you prefer."}},
	},
	"create-multiple": {
		{key: "item-1", title: "What are the key considerations when implementing a comprehensive enterprise-level authentication system?", paragraphs: []string{"Implementing a robust enterprise authentication system requires careful consideration of multiple factors. This includes secure password hashing and storage, multi-factor authentication (MFA) implementation, session management, OAuth2 and SSO integration, regular security audits, rate limiting to prevent brute force attacks, and maintaining detailed audit logs. Additionally, you'll need to consider scalability, performance impact, and compliance with relevant data protection regulations such as GDPR or HIPAA."}},
		{key: "item-2", title: "How does modern distributed system architecture handle eventual consistency and data synchronization across multiple regions?", paragraphs: []string{"Modern distributed systems employ various strategies to maintain data consistency across regions. This often involves using techniques like CRDT (Conflict-Free Replicated Data Types), vector clocks, and gossip protocols. Systems might implement event sourcing patterns, utilize message queues for asynchronous updates, and employ sophisticated conflict resolution strategies. Popular solutions like Amazon's DynamoDB and Google's Spanner demonstrate different approaches to solving these challenges, balancing between consistency, availability, and partition tolerance as described in the CAP theorem."}},
	},
	"create-borders": {
		{key: "billing", title: "How does billing work?", paragraphs: []string{"We offer monthly and annual subscription plans. Billing is charged at the beginning of each cycle, and you can cancel anytime. All plans include automatic backups, 24/7 support, and unlimited team members. There are no hidden fees or setup costs."}},
		{key: "security", title: "Is my data secure?", paragraphs: []string{"Yes. We use end-to-end encryption, SOC 2 Type II compliance, and regular third-party security audits. All data is encrypted at rest and in transit using industry-standard protocols. We also offer optional two-factor authentication and single sign-on for enterprise customers."}},
		{key: "integration", title: "What integrations do you support?", paragraphs: []string{"We integrate with 500+ popular tools including Slack, Zapier, Salesforce, HubSpot, and more. You can also build custom integrations using our REST API and webhooks.", "Our API documentation includes code examples in 10+ programming languages."}},
	},
	"create-card": {
		{key: "plans", title: "What subscription plans do you offer?", paragraphs: []string{"We offer three subscription tiers: Starter ($9/month), Professional ($29/month), and Enterprise ($99/month). Each plan includes increasing storage limits, API access, priority support, and team collaboration features.", "Annual billing is available with a 20% discount. All plans include a 14-day free trial with no credit card required."}, annualLink: true, plansButton: true},
		{key: "billing", title: "How does billing work?", paragraphs: []string{"Billing occurs automatically at the start of each billing cycle. We accept all major credit cards, PayPal, and ACH transfers for enterprise customers.", "You'll receive an invoice via email after each payment. You can update your payment method or billing information anytime in your account settings. Failed payments will trigger automated retry attempts and email notifications."}},
		{key: "upgrade", title: "Can I upgrade or downgrade my plan?", paragraphs: []string{"Yes, you can change your plan at any time. When upgrading, you'll be charged a prorated amount for the remainder of your billing cycle and immediately gain access to new features.", "When downgrading, the change takes effect at the end of your current billing period, and you'll retain access to premium features until then. No refunds are provided for downgrades."}},
		{key: "cancel", title: "How do I cancel my subscription?", paragraphs: []string{"You can cancel your subscription anytime from your account settings. There are no cancellation fees or penalties. Your access will continue until the end of your current billing period.", "After cancellation, your data is retained for 30 days in case you want to reactivate. You can export all your data before or after canceling. We'd love to hear your feedback about why you're leaving."}},
		{key: "refund", title: "What is your refund policy?", paragraphs: []string{"We offer a 30-day money-back guarantee for new subscriptions. If you're not satisfied within the first 30 days, contact our support team for a full refund.", "After 30 days, we don't provide refunds for partial billing periods, but you can cancel anytime to avoid future charges. Enterprise customers have custom refund terms outlined in their contracts."}},
	},
	"create-disabled": {
		{key: "item-1", title: "Can I access my account history?", paragraphs: []string{"Yes, you can view your complete account history including all transactions, plan changes, and support tickets in the Account History section of your dashboard."}},
		{key: "item-2", title: "Premium feature information", paragraphs: []string{"This section contains information about premium features. Upgrade your plan to access this content."}, disabled: true},
		{key: "item-3", title: "How do I update my email address?", paragraphs: []string{"You can update your email address in your account settings. You'll receive a verification email at your new address to confirm the change."}},
	},
	"accordion-demo": {
		{key: "shipping", title: "What are your shipping options?", paragraphs: []string{"We offer standard (5-7 days), express (2-3 days), and overnight shipping. Free shipping on international orders."}},
		{key: "returns", title: "What is your return policy?", paragraphs: []string{"Returns accepted within 30 days. Items must be unused and in original packaging. Refunds processed within 5-7 business days."}},
		{key: "support", title: "How can I contact customer support?", paragraphs: []string{"Reach us via email, live chat, or phone. We respond within 24 hours during business days."}},
	},
	"accordion-disabled": {
		{key: "item-1", title: "Can I access my account history?", paragraphs: []string{"Yes, you can view your complete account history including all transactions, plan changes, and support tickets in the Account History section of your dashboard."}},
		{key: "item-2", title: "Premium feature information", paragraphs: []string{"This section contains information about premium features. Upgrade your plan to access this content."}, disabled: true},
		{key: "item-3", title: "How do I update my email address?", paragraphs: []string{"You can update your email address in your account settings. You'll receive a verification email at your new address to confirm the change."}},
	},
}

var disclosureTranslations = map[string][]disclosureQuestion{
	"en": {
		{key: "item-1", title: "How do I reset my password?", paragraphs: []string{"Click on 'Forgot Password' on the login page, enter your email address, and we'll send you a link to reset your password."}},
		{key: "item-2", title: "Can I change my subscription plan?", paragraphs: []string{"Yes, you can upgrade or downgrade your plan at any time from your account settings. Changes will be reflected in your next billing cycle."}},
		{key: "item-3", title: "What payment methods do you accept?", paragraphs: []string{"We accept all major credit cards, PayPal, and bank transfers. All payments are processed securely through our payment partners."}},
	},
	"ar": {
		{key: "item-1", title: "كيف يمكنني إعادة تعيين كلمة المرور؟", paragraphs: []string{"انقر على 'نسيت كلمة المرور' في صفحة تسجيل الدخول، أدخل عنوان بريدك الإلكتروني، وسنرسل لك رابطًا لإعادة تعيين كلمة المرور. سينتهي صلاحية الرابط خلال 24 ساعة."}},
		{key: "item-2", title: "هل يمكنني تغيير خطة الاشتراك الخاصة بي؟", paragraphs: []string{"نعم، يمكنك ترقية أو تخفيض خطتك في أي وقت من إعدادات حسابك. ستظهر التغييرات في دورة الفوترة التالية."}},
		{key: "item-3", title: "ما هي طرق الدفع التي تقبلونها؟", paragraphs: []string{"نقبل جميع بطاقات الائتمان الرئيسية و PayPal والتحويلات المصرفية. تتم معالجة جميع المدفوعات بأمان من خلال شركاء الدفع لدينا."}},
	},
	"he": {
		{key: "item-1", title: "איך אני מאפס את הסיסמה שלי?", paragraphs: []string{"לחץ על 'שכחתי סיסמה' בעמוד ההתחברות, הזן את כתובת האימייל שלך, ונשלח לך קישור לאיפוס הסיסמה. הקישור יפוג תוך 24 שעות."}},
		{key: "item-2", title: "האם אני יכול לשנות את תוכנית המנוי שלי?", paragraphs: []string{"כן, אתה יכול לשדרג או להוריד את התוכנית שלך בכל עת מההגדרות של החשבון שלך. השינויים יבואו לידי ביטוי במחזור החיוב הבא."}},
		{key: "item-3", title: "אילו אמצעי תשלום אתם מקבלים?", paragraphs: []string{"אנו מקבלים כרטיסי אשראי, PayPal והעברות בנקאיות."}},
	},
}
