package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	nav "github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Native modal/editor state belongs to the SDK. The application retains only
// drafts, values, public Handles and local snapshots, never a frame or scope.
type dialogCompositionPreviewState struct {
	ready, native                bool
	recipe, language, lastRecipe string
	action, popupError           string
	drafts                       map[string]*dialogCompositionDraft
}

type dialogCompositionDraft struct {
	open                                   bool
	name, username, link, nickname, about  string
	tab, chatTheme, accent, spoken, voice  string
	taskPush, taskEmail, custom, twoFactor bool
	trigger, nameTarget                    ui.Handle
	saved                                  map[string]string
	changes, saves, actions                int
}

var dialogCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Edit profile (docs)"},
	{Value: "usage", Label: "Basic dialog (guide)"},
	{Value: "share", Label: "Custom close / share link"},
	{Value: "no-close", Label: "No close button (docs)"},
	{Value: "sticky", Label: "Sticky footer (docs)"},
	{Value: "scrollable", Label: "Scrollable content (docs)"},
	{Value: "rtl", Label: "Translated profile"},
	{Value: "create-form", Label: "Create: With Form"},
	{Value: "create-scrollable", Label: "Create: Scrollable Content"},
	{Value: "create-sticky", Label: "Create: With Sticky Footer"},
	{Value: "create-no-close", Label: "Create: No Close Button"},
	{Value: "create-chat", Label: "Create: Chat Settings"},
}

func (p *dialogCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.lastRecipe, p.language = true, native, "demo", "demo", "ar"
	p.drafts = map[string]*dialogCompositionDraft{}
}

func (p *dialogCompositionPreviewState) draft() *dialogCompositionDraft {
	if d := p.drafts[p.recipe]; d != nil {
		return d
	}
	d := &dialogCompositionDraft{name: "Pedro Duarte", username: "@peduarte", link: "https://ui.shadcn.com/docs/installation", tab: "general", chatTheme: "system", accent: "default", spoken: "en", voice: "samantha", custom: true}
	p.drafts[p.recipe] = d
	return d
}

func (p *dialogCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) ui.Element {
	if native {
		return control.NativeSelect(c, value, props).MinWidth(0)
	}
	e, err := desktop.Select(c, h, key, value, props)
	if err != nil {
		p.popupError = err.Error()
	}
	return e.MinWidth(0)
}

func (p *dialogCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native || p.lastRecipe != p.recipe {
		if d := p.drafts[p.lastRecipe]; d != nil {
			d.open = false
		}
		if h != nil {
			for _, key := range []string{"source-dialog-config", "source-dialog-language", "source-dialog-theme", "source-dialog-accent", "source-dialog-spoken", "source-dialog-voice", "source-dialog-help"} {
				h.Close(key)
			}
		}
		p.native, p.lastRecipe, p.popupError = native, p.recipe, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Dialog compositions"}, display.CardSlots{
		Title: "Dialog compositions", Description: "Source native modals, profile drafts, fixed actions and complete chat settings.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-dialog-config", &p.recipe, control.SelectProps{Label: "Dialog configuration", Options: dialogCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-dialog-language", &p.language, control.SelectProps{Label: "Dialog language", Options: statusLanguages()})
			}
			ui.Column(child.Key("dialog-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() { p.draw(child, h, native) })
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert)
			}
		},
	})
}

type dialogSourceWords struct{ open, title, description, name, username, cancel, save string }

func dialogSourceProfileWords(language string) dialogSourceWords {
	switch language {
	case "ar":
		return dialogSourceWords{"فتح الحوار", "تعديل الملف الشخصي", "قم بإجراء تغييرات على ملفك الشخصي هنا. انقر فوق حفظ عند الانتهاء.", "الاسم", "اسم المستخدم", "إلغاء", "حفظ التغييرات"}
	case "he":
		return dialogSourceWords{"פתח דיאלוג", "ערוך פרופיל", "בצע שינויים בפרופיל שלך כאן. לחץ על שמור כשתסיים.", "שם", "שם משתמש", "בטל", "שמור שינויים"}
	default:
		return dialogSourceWords{"Open Dialog", "Edit profile", "Make changes to your profile here. Click save when you're done.", "Name", "Username", "Cancel", "Save changes"}
	}
}

const dialogSourceLorem = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

func dialogSourceHeader(c *ui.Context, title, description string) {
	t := theme.TokensOf(c)
	ui.Column(c).FillWidth().MinWidth(0).Shrink(0).Gap(8).Children(func() {
		theme.StyleHeading(c, ui.Text(c, title).FontSize(14).FontWeight(500).LineHeight(1).PaddingEnd(24).Role(ui.RoleHeading).Level(2))
		ui.Text(c, description).FontSize(14).TextColor(t.MutedForeground)
	})
}

func (p *dialogCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	d := p.draft()
	words := dialogSourceProfileWords("en")
	width := float32(448)
	hideClose, profile, scrollable, sticky := false, false, false, false
	if p.recipe == "rtl" {
		words = dialogSourceProfileWords(p.language)
	}
	switch p.recipe {
	case "demo", "rtl":
		width, profile = 384, true
	case "create-form":
		words.open, profile = "Edit Profile", true
		words.description += " Your profile will be updated immediately."
	case "usage":
		words.open, words.title, words.description = "Open", "Are you absolutely sure?", "This action cannot be undone. This will permanently delete your account and remove your data from our servers."
	case "share":
		words.open, words.title, words.description = "Share", "Share link", "Anyone who has this link will be able to view this."
	case "no-close", "create-no-close":
		words.open, words.title, words.description, hideClose = "No Close Button", "No Close Button", "This dialog doesn't have a close button in the top-right corner.", true
	case "scrollable", "create-scrollable":
		words.open, words.title, words.description, scrollable = "Scrollable Content", "Scrollable Content", "This is a dialog with scrollable content.", true
	case "sticky":
		words.open, words.title, words.description, scrollable, sticky = "Sticky Footer", "Sticky Footer", "This dialog has a sticky footer that stays visible while the content scrolls.", true, true
	case "create-sticky":
		words.open, words.title, words.description, scrollable, sticky = "Sticky Footer", "Scrollable Content", "This is a dialog with scrollable content.", true, true
	case "create-chat":
		words.open, words.title, words.description = "Chat Settings", "Chat Settings", "Customize your chat settings: theme, accent color, spoken language, voice, personality, and custom instructions."
	}
	trigger := control.Button(c, words.open, control.ButtonProps{Variant: control.Outline}).Bind(&d.trigger)
	if trigger.Clicked() {
		d.open = true
	}
	build := func() {
		props := overlay.DialogProps{Title: words.title, Description: words.description, Width: width, HideClose: true, HideHeader: true, NoScroll: true, ReturnFocus: &d.trigger}
		if profile {
			props.InitialFocus = &d.nameTarget
		}
		if profile || sticky || p.recipe == "share" || p.recipe == "create-no-close" {
			props.Footer = func() {
				windowWidth, _ := c.Size()
				fit := func(button ui.Element) ui.Element {
					if windowWidth < 640 {
						button.FillWidth()
					}
					return button
				}
				if profile {
					if fit(control.Button(c, words.cancel, control.ButtonProps{Variant: control.Outline})).Clicked() {
						d.open = false
					}
					if fit(control.Button(c, words.save, control.ButtonProps{})).Clicked() {
						d.saved = map[string]string{"name": d.name, "username": d.username}
						d.saves++
						p.action = "Local dialog action: Save changes"
					}
				} else {
					closeButton := func() {
						if fit(control.Button(c, "Close", control.ButtonProps{Variant: control.Outline})).Clicked() {
							d.open = false
						}
					}
					if p.recipe == "share" {
						ui.Row(c).FillWidth().Justify(ui.Start).Children(func() {
							if fit(control.Button(c, "Close", control.ButtonProps{})).Clicked() {
								d.open = false
							}
						})
					} else {
						closeButton()
					}
				}
			}
		}
		panel := overlay.Dialog(c, &d.open, props, func() {
			ui.Column(c).FillWidth().MinWidth(0).MinHeight(0).Gap(24).Children(func() {
				dialogSourceHeader(c, words.title, words.description)
				switch {
				case profile:
					control.FieldGroup(c, func() {
						for _, field := range []struct {
							label  string
							value  *string
							handle *ui.Handle
						}{{words.name, &d.name, &d.nameTarget}, {words.username, &d.username, nil}} {
							control.FieldWith(c.Key(field.label), control.FieldProps{}, func(f *control.FieldScope) {
								label := f.Label(c, control.FieldLabelProps{Text: field.label}, nil)
								if p.recipe != "create-form" {
									label.LineHeight(1)
								}
								f.Control(c, func() ui.Element {
									return control.Input(c, field.value, control.InputProps{Label: field.label, Handle: field.handle, OnChange: func(string) { d.changes++ }})
								})
							})
						}
					})
				case p.recipe == "share":
					control.Input(c, &d.link, control.InputProps{Label: "Link", ReadOnly: true})
				case scrollable:
					_, height := c.Size()
					fraction := float32(.5)
					if strings.HasPrefix(p.recipe, "create-") {
						fraction = .7
					}
					ui.Scroll(c).FillWidth().MinHeight(0).Shrink(1).MaxHeight(height * fraction).Scrollbars(false).Role(ui.RoleRegion).Label("Dialog source scroll").Children(func() {
						ui.Column(c).FillWidth().Shrink(0).Children(func() {
							for i := 0; i < 10; i++ {
								ui.Text(c.Key(i), dialogSourceLorem).FontSize(14).LineHeight(1.5).Margin(0, 0, 16, 0)
							}
						})
					})
				case p.recipe == "create-chat":
					p.chat(c, h, native, d)
				}
			})
		})
		if !hideClose && panel.Valid() {
			// A real current-modal button supplies the source Ghost/IconSmall
			// face. Appending children to the returned native panel retains its
			// modal focus/disabled/exit scope; no frame Element is stored.
			panel.Children(func() {
				if control.Button(c, "", control.ButtonProps{Variant: control.Ghost, Size: control.IconSmall, Label: "Close dialog", Icon: outlined.Close()}).Absolute().Top(16).End(16).Clicked() {
					d.open = false
					c.Invalidate()
				}
			})
		}
	}
	if p.recipe == "rtl" {
		ui.Column(c).Direction(childDirection(p.language)).Children(build)
	} else {
		build()
	}
}

var dialogChatSections = []control.Option{{Value: "general", Label: "General"}, {Value: "notifications", Label: "Notifications"}, {Value: "personalization", Label: "Personalization"}, {Value: "security", Label: "Security"}}
var dialogChatThemes = []control.Option{{Value: "light", Label: "Light"}, {Value: "dark", Label: "Dark"}, {Value: "system", Label: "System"}}
var dialogChatAccents = []control.Option{{Value: "default", Label: "Default"}, {Value: "red", Label: "Red"}, {Value: "blue", Label: "Blue"}, {Value: "green", Label: "Green"}, {Value: "purple", Label: "Purple"}, {Value: "pink", Label: "Pink"}}
var dialogChatLanguages = []control.Option{{Value: "auto", Label: "Auto"}, {Value: "en", Label: "English"}, {Value: "es", Label: "Spanish"}, {Value: "fr", Label: "French"}, {Value: "de", Label: "German"}, {Value: "it", Label: "Italian"}, {Value: "pt", Label: "Portuguese"}, {Value: "ru", Label: "Russian"}, {Value: "zh", Label: "Chinese"}, {Value: "ja", Label: "Japanese"}, {Value: "ko", Label: "Korean"}, {Value: "ar", Label: "Arabic"}, {Value: "hi", Label: "Hindi"}, {Value: "bn", Label: "Bengali"}, {Value: "te", Label: "Telugu"}, {Value: "mr", Label: "Marathi"}, {Value: "kn", Label: "Kannada"}, {Value: "ml", Label: "Malayalam"}}
var dialogChatVoices = []control.Option{{Value: "samantha", Label: "Samantha"}, {Value: "alex", Label: "Alex"}, {Value: "fred", Label: "Fred"}, {Value: "victoria", Label: "Victoria"}, {Value: "tom", Label: "Tom"}, {Value: "karen", Label: "Karen"}, {Value: "sam", Label: "Sam"}, {Value: "daniel", Label: "Daniel"}}

func (p *dialogCompositionPreviewState) chat(c *ui.Context, h *desktop.Host, native bool, d *dialogCompositionDraft) {
	w, height := c.Size()
	ui.Column(c).FillWidth().MinWidth(0).MinHeight(0).Gap(16).Children(func() {
		if w < 768 {
			control.NativeSelect(c, &d.tab, control.SelectProps{Label: "Chat settings section", Options: dialogChatSections}).FillWidth()
		}
		nav.TabsRoot(c, &d.tab, nav.TabsRootProps{Key: "dialog-chat-tabs", Label: "Chat settings tabs"}, func(current *ui.Context) {
			if w >= 768 {
				nav.TabsList(current, nav.TabsListProps{Label: "Chat settings tabs"}, func(current *ui.Context) {
					for _, o := range dialogChatSections {
						nav.TabsTrigger(current, o.Value, nav.TabsTriggerProps{Label: o.Label})
					}
				}).FillWidth().AlignSelf(ui.Stretch)
			}
			ui.Scroll(current).FillWidth().MinWidth(0).MinHeight(0).Shrink(1).MaxHeight(height*.7).Border(1, theme.TokensOf(current).Border).Radius(theme.TokensOf(current).RadiusLG()).Role(ui.RoleRegion).Label("Chat settings content").Children(func() {
				ui.Column(current).FillWidth().MinWidth(0).MinHeight(550).Padding(24).Children(func() {
					for _, section := range dialogChatSections {
						nav.TabsContent(current, section.Value, nav.TabsContentProps{}, func(current *ui.Context) { p.chatPanel(current, h, native, d, section.Value) })
					}
				})
			})
		}).FillWidth().MinHeight(0)
	})
}

func (p *dialogCompositionPreviewState) chatSelect(c *ui.Context, h *desktop.Host, native bool, d *dialogCompositionDraft, key, label, description string, value *string, options []control.Option, responsive bool) {
	orientation := control.FieldHorizontal
	if responsive {
		orientation = control.FieldResponsive
	}
	control.FieldWith(c.Key(key), control.FieldProps{Orientation: orientation, ResponsiveWidth: 400}, func(f *control.FieldScope) {
		if description != "" {
			f.Content(c, func() { f.Label(c, control.FieldLabelProps{Text: label}, nil); f.Description(c, description) })
		} else {
			f.Label(c, control.FieldLabelProps{Text: label}, nil)
		}
		f.Control(c, func() ui.Element {
			e := p.choose(c, h, native, "source-dialog-"+key, value, control.SelectProps{Label: label, Options: options}).MinWidth(125).Width(125).Shrink(0)
			if e.Changed() {
				d.changes++
			}
			return e
		})
	})
}

func (p *dialogCompositionPreviewState) chatPanel(c *ui.Context, h *desktop.Host, native bool, d *dialogCompositionDraft, section string) {
	control.FieldGroup(c, func() {
		sep := func() { control.FieldSeparator(c, "") }
		switch section {
		case "general":
			p.chatSelect(c, h, native, d, "theme", "Theme", "", &d.chatTheme, dialogChatThemes, false)
			sep()
			p.chatSelect(c, h, native, d, "accent", "Accent Color", "", &d.accent, dialogChatAccents, false)
			sep()
			p.chatSelect(c, h, native, d, "spoken", "Spoken Language", "For best results, select the language you mainly speak. If it's not listed, it may still be supported via auto-detection.", &d.spoken, dialogChatLanguages, true)
			sep()
			p.chatSelect(c, h, native, d, "voice", "Voice", "", &d.voice, dialogChatVoices, false)
		case "notifications":
			dialogSourceFieldset(c, func() {
				control.FieldLabel(c, control.FieldLabelProps{Text: "Responses"}, nil)
				control.FieldDescription(c, "Get notified when ChatGPT responds to requests that take time, like research or image generation.")
				checked := true
				control.Checkbox(c, &checked, "Push notifications", control.CheckboxProps{Disabled: true}).FontWeight(400)
			})
			sep()
			dialogSourceFieldset(c, func() {
				control.FieldLabel(c, control.FieldLabelProps{Text: "Tasks"}, nil)
				owner := ui.Box(c).FillWidth()
				router := ui.Local(owner, "dialog-tasks-router", func() *ui.Router { return ui.NewRouter("/dialog-examples") })
				owner.Children(func() {
					(*router).View(c, func(*ui.Route) {
						control.FieldDescriptionWith(c, "Get notified when tasks you've created have updates. Manage tasks", func() {
							ui.RichText(c).FillWidth().MinWidth(0).FontSize(14).TextColor(theme.TokensOf(c).MutedForeground).Children(func() {
								ui.Text(c, "Get notified when tasks you've created have updates. ")
								if ui.Link(c, "Manage tasks", "#").Underline().Clicked() {
									d.actions++
									p.action = "Local dialog action: Manage tasks"
								}
							})
						})
					})
				})
				ui.Column(c).Gap(12).Children(func() {
					control.Checkbox(c, &d.taskPush, "Push notifications", control.CheckboxProps{}).FontWeight(400)
					control.Checkbox(c, &d.taskEmail, "Email notifications", control.CheckboxProps{}).FontWeight(400)
				})
			})
		case "personalization":
			control.FieldWith(c, control.FieldProps{Orientation: control.FieldResponsive, ResponsiveWidth: 400}, func(f *control.FieldScope) {
				f.Label(c, control.FieldLabelProps{Text: "Nickname"}, nil)
				f.Control(c, func() ui.Element {
					return control.InputGroup(c, control.InputGroupProps{InlineEnd: func() {
						anchor := control.Button(c, "", control.ButtonProps{Variant: control.Ghost, Size: control.IconExtraSmall, Label: "Nickname help", Icon: outlined.Info()})
						props := desktop.TooltipProps{Description: "Used to identify you in the chat.", MaxWidth: 280}
						content := func(child *ui.Context) {
							ui.Row(child).Gap(8).AlignItems(ui.Center).Children(func() {
								ui.Text(child, "Used to identify you in the chat.")
								display.KbdWith(child, "N", display.KbdProps{InTooltip: !native})
							})
						}
						var err error
						if native {
							err = desktop.NativeTooltipWith(c, h, "source-dialog-help", anchor, props, content)
						} else {
							err = desktop.TooltipWith(c, h, "source-dialog-help", anchor, props, content)
						}
						if err != nil {
							p.popupError = err.Error()
						}
					}}, func() ui.Element {
						return control.Input(c, &d.nickname, control.InputProps{Label: "Nickname", Placeholder: "Broski", OnChange: func(string) { d.changes++ }})
					})
				})
			})
			sep()
			control.FieldWith(c, control.FieldProps{}, func(f *control.FieldScope) {
				f.Content(c, func() {
					f.Label(c, control.FieldLabelProps{Text: "More about you"}, nil)
					f.Description(c, "Tell us more about yourself. This will be used to help us personalize your experience.")
				})
				f.Control(c, func() ui.Element {
					return control.Textarea(c, &d.about, control.TextareaProps{Label: "More about you", Placeholder: "I'm a software engineer...", OnChange: func(string) { d.changes++ }}).MinHeight(120)
				})
			})
			sep()
			control.FieldWith(c, control.FieldProps{}, func(f *control.FieldScope) {
				f.Label(c, control.FieldLabelProps{Label: "Enable customizations", Choice: true}, func() {
					ui.Row(c).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Center).Children(func() {
						f.Content(c, func() {
							ui.Text(c, "Enable customizations").FontSize(14).FontWeight(500)
							f.Description(c, "Enable customizations to make ChatGPT more personalized.")
						})
						f.Control(c, func() ui.Element {
							return control.Switch(c, &d.custom, control.SwitchProps{})
						})
					})
				}).Border(0, ui.Transparent).Padding(0)
			})
		case "security":
			control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal}, func(f *control.FieldScope) {
				f.Content(c, func() {
					f.Label(c, control.FieldLabelProps{Text: "Multi-factor authentication"}, nil)
					f.Description(c, "Enable multi-factor authentication to secure your account. If you do not have a two-factor authentication device, you can use a one-time code sent to your email.")
				})
				f.Control(c, func() ui.Element {
					return control.Switch(c, &d.twoFactor, control.SwitchProps{})
				})
			})
			sep()
			for _, item := range []struct{ title, description, button string }{{"Log out", "Log out of your account on this device.", "Log Out"}, {"Log out of all devices", "This will log you out of all devices, including the current session. It may take up to 30 minutes for the changes to take effect.", "Log Out All"}} {
				ui.Row(c.Key(item.title)).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Center).Children(func() {
					control.FieldContent(c, func() {
						ui.Text(c, item.title).FontSize(14).FontWeight(500)
						control.FieldDescription(c, item.description)
					})
					if control.Button(c, item.button, control.ButtonProps{Variant: control.Outline, Size: control.Small}).Shrink(0).Clicked() {
						d.actions++
						p.action = "Local dialog action: " + item.button
					}
				})
				if item.button == "Log Out" {
					sep()
				}
			}
		}
	})
}

func dialogSourceFieldset(c *ui.Context, content func()) {
	control.FieldSetWith(c, control.FieldSetProps{}, control.FieldSetSlots{Content: func(*ui.Context) { content() }})
}
