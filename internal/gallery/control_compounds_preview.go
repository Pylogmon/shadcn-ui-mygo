package gallery

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Application values and public menu/scroll handles survive configuration and
// page changes. Frame Contexts, Elements and builders are never stored here.
type controlCompoundsPreviewState struct {
	ready, native, voice, following              bool
	recipe, language, label                      string
	query, message, amount, currency, name, task string
	mediaLevel, additions                        int
	action, avatarAction, error, copied          string
	copySerial                                   uint64
	more, follow, avatarMenu                     controlCompoundMenu
	taskOpen                                     bool
	nameHandle                                   ui.Handle
}

type controlCompoundMenu struct {
	open  bool
	state desktop.MenuState
}

var controlCompoundRecipes = []control.Option{
	{Value: "demo", Label: "Mail actions"},
	{Value: "orientation", Label: "Vertical controls"},
	{Value: "size", Label: "Small, default and large"},
	{Value: "nested", Label: "Nested message input"},
	{Value: "separator", Label: "Copy and paste"},
	{Value: "split", Label: "Split button"},
	{Value: "input", Label: "Search input"},
	{Value: "input-group", Label: "Voice input group"},
	{Value: "dropdown", Label: "Follow with menu"},
	{Value: "select", Label: "Currency and amount"},
	{Value: "popover", Label: "Task popover"},
	{Value: "text", Label: "Text and label"},
	{Value: "rtl", Label: "Translated RTL actions"},
}

func (p *controlCompoundsPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language, p.label = true, native, "demo", "ar", "personal"
	p.currency, p.amount, p.mediaLevel = "$", "10.00", 50
}

func (p *controlCompoundsPreviewState) syncRoute(h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native == native {
		return
	}
	p.native = native
	p.error = ""
	p.more.open, p.follow.open, p.avatarMenu.open, p.taskOpen = false, false, false, false
	p.more.state.Reset()
	p.follow.state.Reset()
	p.avatarMenu.state.Reset()
	if h != nil {
		for _, key := range []string{"compound-mail", "compound-follow", "compound-avatar", "compound-task", "compound-currency", "compound-recipe", "compound-language", "compound-voice-tip"} {
			h.Close(key)
		}
	}
}

func (p *controlCompoundsPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) ui.Element {
	if native {
		e := control.NativeSelect(c, value, props)
		if e.Clicked() {
			p.error = ""
		}
		return e
	}
	e, err := desktop.Select(c, h, key, value, props)
	if e.Clicked() {
		p.error = ""
	}
	if err != nil {
		p.error = err.Error()
	}
	return e
}

func (p *controlCompoundsPreviewState) menu(c *ui.Context, h *desktop.Host, native bool, key, label string, state *controlCompoundMenu, items []overlay.MenuItem, decorate func(ui.Element)) string {
	var chosen string
	var err error
	render := func(e ui.Element) {
		decorate(e)
		if e.Clicked() {
			p.error = ""
		}
	}
	if native {
		chosen, err = desktop.NativeMenu(c, overlay.MenuProps{Items: items}, render)
	} else {
		anchor := ui.ButtonBase(c).Label(tr(c, label))
		render(anchor)
		width := float32(160)
		if key == "compound-follow" {
			width = 176
		} else if key == "compound-avatar" {
			width = 128
		}
		chosen, err = desktop.Menu(c, h, key, anchor, &state.open, &state.state, desktop.MenuProps{Label: tr(c, label), Items: items, Width: width, Placement: mygo.PlacementBottomEnd, Gap: 6})
	}
	if err != nil {
		p.error = err.Error()
	}
	return chosen
}

func (p *controlCompoundsPreviewState) iconMenu(c *ui.Context, h *desktop.Host, native bool, key, label string, state *controlCompoundMenu, icon *ui.SVG, items []overlay.MenuItem) string {
	return p.menu(c, h, native, key, label, state, items, func(e ui.Element) {
		control.StyleButton(c, e, "", control.ButtonProps{Variant: control.Outline, Size: control.Icon, Label: tr(c, label), Icon: icon})
	})
}

// buttonGroups is wired to Forms & controls by the Gallery owner. The source
// configurations share application drafts; selecting a recipe never resets them.
func (p *controlCompoundsPreviewState) buttonGroups(c *ui.Context, h *desktop.Host, native bool) {
	p.syncRoute(h, native)
	display.CardWith(c.Key("control-compound-groups"), display.CardProps{Label: tr(c, "Button group examples")}, display.CardSlots{
		Title: tr(c, "Button groups"), Description: tr(c, "Joined actions, editors and native popup choices."),
		Content: func(child *ui.Context) {
			ui.Column(child).FillWidth().MinWidth(0).Gap(20).Children(func() {
				before := p.recipe
				p.choose(child, h, native, "compound-recipe", &p.recipe, control.SelectProps{Label: tr(child, "Button group configuration"), Options: localizedOptions(child, controlCompoundRecipes)}).MaxWidth(360)
				if before != p.recipe {
					p.error = ""
					p.taskOpen, p.more.open, p.follow.open = false, false, false
					if h != nil {
						h.Close("compound-task")
						h.Close("compound-mail")
						h.Close("compound-follow")
					}
				}
				ui.Column(child.Key(p.recipe)).FillWidth().MinWidth(0).Gap(14).Children(func() {
					p.renderGroup(child, h, native)
				})
				if p.action != "" {
					ui.Text(child, tr(child, p.action)).FontSize(13).Role(ui.RoleStatus)
				}
				if p.error != "" {
					ui.Text(child, tr(child, p.error)).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
				}
			})
		},
	})
}

func (p *controlCompoundsPreviewState) actionButton(c *ui.Context, caption, label string, props control.ButtonProps) ui.Element {
	if props.Label == "" {
		props.Label = label
	}
	b := control.Button(c, tr(c, caption), props)
	if b.Clicked() {
		p.action = label + " selected locally."
	}
	return b
}

func (p *controlCompoundsPreviewState) renderGroup(c *ui.Context, h *desktop.Host, native bool) {
	switch p.recipe {
	case "demo", "rtl":
		p.mailGroup(c, h, native)
	case "orientation":
		control.ButtonGroup(c, func() {
			if control.Button(c, "", control.ButtonProps{Variant: control.Outline, Size: control.Icon, Icon: outlined.Add(), Label: tr(c, "Increase media level")}).Clicked() {
				p.mediaLevel = min(100, p.mediaLevel+10)
			}
			if control.Button(c, "", control.ButtonProps{Variant: control.Outline, Size: control.Icon, Icon: outlined.Remove(), Label: tr(c, "Decrease media level")}).Clicked() {
				p.mediaLevel = max(0, p.mediaLevel-10)
			}
		}, control.ButtonGroupProps{Label: tr(c, "Media controls"), Vertical: true})
		ui.Text(c, fmt.Sprintf(tr(c, "Media level: %d"), p.mediaLevel)).Role(ui.RoleStatus)
	case "size":
		for _, size := range []struct {
			label      string
			text, icon control.Size
		}{{"Small", control.Small, control.IconSmall}, {"Default", control.DefaultSize, control.Icon}, {"Large", control.Large, control.IconLarge}} {
			control.ButtonGroup(c.Key(size.label), func() {
				for _, label := range []string{size.label, "Button", "Group"} {
					p.actionButton(c, label, size.label+" "+label, control.ButtonProps{Variant: control.Outline, Size: size.text})
				}
				p.addButton(c, size.icon)
			}, control.ButtonGroupProps{Label: size.label + " button group"})
		}
	case "split":
		control.ButtonGroup(c, func() {
			p.actionButton(c, "Button", "Split primary action", control.ButtonProps{Variant: control.Secondary})
			control.ButtonGroupSeparator(c)
			p.addButton(c, control.Icon)
		}, control.ButtonGroupProps{Label: tr(c, "Split button")})
		ui.Text(c, fmt.Sprintf(tr(c, "Added locally: %d"), p.additions)).Role(ui.RoleStatus)
	case "separator":
		control.ButtonGroup(c, func() {
			if control.Button(c, tr(c, "Copy"), control.ButtonProps{Variant: control.Secondary, Size: control.Small, Label: tr(c, "Copy grouped draft")}).Clicked() {
				p.copySerial++
				serial := p.copySerial
				text := p.message
				p.action = "Copying the grouped draft…"
				c.WriteClipboardThen(text, func(err error) {
					if serial != p.copySerial {
						return
					}
					if err != nil {
						p.action = "Could not copy the grouped draft."
						return
					}
					p.copied = text
					p.action = "Copied the grouped draft."
				})
			}
			control.ButtonGroupSeparator(c)
			if control.Button(c, tr(c, "Paste"), control.ButtonProps{Variant: control.Secondary, Size: control.Small, Label: tr(c, "Paste grouped draft")}).Clicked() {
				p.message = c.ReadClipboard()
				p.action = "Pasted into the grouped draft."
			}
		}, control.ButtonGroupProps{Label: tr(c, "Copy and paste group")})
		ui.Text(c, tr(c, "Draft: ")+p.message).FontSize(13)
	case "input":
		control.ButtonGroup(c, func() {
			control.Input(c, &p.query, control.InputProps{Label: tr(c, "Grouped search"), Placeholder: tr(c, "Search...")}).MinWidth(0).Grow(1)
			if control.Button(c, "", control.ButtonProps{Variant: control.Outline, Size: control.Icon, Icon: outlined.Search(), Label: tr(c, "Search grouped text")}).Clicked() {
				p.action = "Local search: " + p.query
			}
		}, control.ButtonGroupProps{Label: tr(c, "Search group")}).FillWidth().MaxWidth(440)
	case "nested", "input-group":
		p.messageGroup(c, h, native)
	case "dropdown":
		control.ButtonGroup(c, func() {
			caption := "Follow"
			if p.following {
				caption = "Following"
			}
			if control.Button(c, tr(c, caption), control.ButtonProps{Variant: control.Outline, Label: tr(c, "Follow conversation")}).Clicked() {
				p.following = !p.following
				p.action = "Local follow: " + strconv.FormatBool(p.following)
			}
			items := []overlay.MenuItem{{Key: "mute", Label: tr(c, "Mute Conversation")}, {Key: "read", Label: tr(c, "Mark as Read")}, {Key: "report", Label: tr(c, "Report Conversation")}, {Key: "block", Label: tr(c, "Block User")}, {Key: "share", Label: tr(c, "Share Conversation")}, {Key: "copy", Label: tr(c, "Copy Conversation")}, {Separator: true}, {Key: "delete", Label: tr(c, "Delete Conversation"), Destructive: true}}
			if chosen := p.iconMenu(c, h, native, "compound-follow", "Conversation options", &p.follow, outlined.ExpandMore(), items); chosen != "" {
				p.action = "Local conversation action: " + chosen
			}
		}, control.ButtonGroupProps{Label: tr(c, "Follow group")})
	case "select":
		control.ButtonGroup(c, func() {
			control.ButtonGroup(c, func() {
				props := control.SelectProps{Label: tr(c, "Grouped currency"), Options: localizedOptions(c, []control.Option{{Value: "$", Label: "$ US Dollar"}, {Value: "€", Label: "€ Euro"}, {Value: "£", Label: "£ British Pound"}}), TriggerContent: func(child *ui.Context, o control.Option) { ui.Text(child, o.Value).Font("monospace") }}
				control.ButtonGroupItem(c, p.choose(c, h, native, "compound-currency", &p.currency, props).MinWidth(72))
				control.Input(c, &p.amount, control.InputProps{Label: tr(c, "Grouped amount"), Placeholder: "10.00"}).MinWidth(0).Width(150)
			}, control.ButtonGroupProps{Label: tr(c, "Currency and amount")})
			control.ButtonGroup(c, func() {
				if control.Button(c, "", control.ButtonProps{Variant: control.Outline, Size: control.Icon, Icon: outlined.ArrowForward(), Label: tr(c, "Preview grouped amount")}).Clicked() {
					amount, err := strconv.ParseFloat(strings.TrimSpace(p.amount), 64)
					if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
						p.action = "Enter a finite amount."
					} else {
						p.action = fmt.Sprintf("Local amount preview: %s%.2f", p.currency, amount)
					}
				}
			}, control.ButtonGroupProps{Label: tr(c, "Amount action")})
		}, control.ButtonGroupProps{Label: tr(c, "Nested amount group")})
	case "popover":
		control.ButtonGroup(c, func() {
			p.actionButton(c, "Copilot", "Copilot task", control.ButtonProps{Variant: control.Outline, Icon: outlined.SmartToy()})
			anchor := control.Button(c, "", control.ButtonProps{Variant: control.Outline, Size: control.Icon, Icon: outlined.ExpandMore(), Label: tr(c, "Open grouped task")})
			p.taskPopover(c, h, native, anchor)
		}, control.ButtonGroupProps{Label: tr(c, "Task group")})
	case "text":
		control.ButtonGroup(c, func() {
			tag := control.ButtonGroupText(c, "Name")
			control.Input(c, &p.name, control.InputProps{Label: tr(c, "Grouped name"), Placeholder: tr(c, "Type something here..."), Handle: &p.nameHandle}).LabelledBy(tag).MinWidth(0).Grow(1)
		}, control.ButtonGroupProps{Label: tr(c, "Text and input group")}).FillWidth().MaxWidth(440)
	}
}

func (p *controlCompoundsPreviewState) addButton(c *ui.Context, size control.Size) {
	variant := control.Outline
	if p.recipe == "split" {
		variant = control.Secondary
	}
	if control.Button(c, "", control.ButtonProps{Variant: variant, Size: size, Icon: outlined.Add(), Label: tr(c, "Add grouped item")}).Clicked() {
		p.additions++
		p.action = fmt.Sprintf("Added locally: %d", p.additions)
	}
}

func (p *controlCompoundsPreviewState) messageGroup(c *ui.Context, h *desktop.Host, native bool) {
	build := func() {
		control.ButtonGroup(c, func() {
			control.ButtonGroup(c, func() { p.addButton(c, control.Icon) })
			control.ButtonGroup(c, func() {
				control.InputGroup(c, control.InputGroupProps{InlineEnd: func() {
					if p.recipe == "nested" {
						anchor := material.Icon(c, outlined.GraphicEq(), material.Props{Size: 16})
						p.voiceTooltip(c, h, native, anchor)
						return
					}
					b := control.Toggle(c, &p.voice, "", control.ToggleProps{Size: control.IconExtraSmall, Label: tr(c, "Voice Mode"), Icon: outlined.GraphicEq()})
					if p.voice {
						if theme.TokensOf(c).Dark {
							b.Background(ui.RGB(154, 52, 18)).TextColor(ui.RGB(255, 237, 213))
						} else {
							b.Background(ui.RGB(255, 237, 213)).TextColor(ui.RGB(194, 65, 12))
						}
					}
					p.voiceTooltip(c, h, native, b)
				}}, func() ui.Element {
					placeholder := "Send a message..."
					if p.recipe == "input-group" && p.voice {
						placeholder = "Record and send audio..."
					}
					return control.Input(c, &p.message, control.InputProps{Label: tr(c, "Grouped message"), Placeholder: tr(c, placeholder), Disabled: p.recipe == "input-group" && p.voice})
				})
			}).Grow(1).MinWidth(0)
		}, control.ButtonGroupProps{Label: tr(c, "Nested message group")}).FillWidth().MaxWidth(480)
	}
	if p.recipe == "input-group" {
		tokens := *theme.TokensOf(c)
		tokens.Radius = 999
		theme.Scope(c, &tokens).FillWidth().Children(build)
	} else {
		build()
	}
}

func (p *controlCompoundsPreviewState) voiceTooltip(c *ui.Context, h *desktop.Host, native bool, anchor ui.Element) {
	if h == nil {
		return
	}
	if anchor.Clicked() {
		p.error = ""
	}
	if native {
		_ = desktop.NativeTooltip(c, h, "compound-voice-tip", anchor, "Voice Mode", desktop.TooltipProps{})
	} else {
		_ = desktop.Tooltip(c, h, "compound-voice-tip", anchor, "Voice Mode", desktop.TooltipProps{})
	}
}

func (p *controlCompoundsPreviewState) taskPopover(c *ui.Context, h *desktop.Host, native bool, anchor ui.Element) {
	if h == nil {
		if anchor.Clicked() {
			p.error = "A parent-bound Host is needed to open the task popover."
		}
		return
	}
	if anchor.Clicked() {
		p.error = ""
	}
	props := desktop.PopoverProps{Label: tr(c, "Grouped task"), Width: 340, Height: 250, FitContent: true, MaxHeight: 320, Placement: mygo.PlacementBottomEnd}
	content := func(child *ui.Context) {
		ui.Text(child, tr(child, "Start a new task with Copilot")).FontWeight(600)
		ui.Text(child, tr(child, "Describe your task in natural language."))
		if native {
			ui.TextArea(child, &p.task).Label(tr(child, "Task Description")).Placeholder(tr(child, "I need to...")).FillWidth().Height(96)
		} else {
			control.Textarea(child, &p.task, control.TextareaProps{Label: tr(child, "Task Description"), Placeholder: tr(child, "I need to..."), Rows: 3})
		}
		ui.Text(child, tr(child, "Your task draft stays in this preview.")).FontSize(12)
	}
	var err error
	if native {
		err = desktop.NativePopover(c, h, "compound-task", anchor, &p.taskOpen, props, content)
	} else {
		err = desktop.Popover(c, h, "compound-task", anchor, &p.taskOpen, props, content)
	}
	if err != nil {
		p.error = err.Error()
	}
}

// avatar is wired to Data & display by the Gallery owner. The menu follows the
// source's four items; actions deliberately report a local preview selection.
func (p *controlCompoundsPreviewState) avatar(c *ui.Context, h *desktop.Host, native bool) {
	p.syncRoute(h, native)
	display.CardWith(c.Key("control-compound-avatar"), display.CardProps{Label: tr(c, "Avatar dropdown example"), Size: display.CardSmall}, display.CardSlots{
		Title: tr(c, "Avatar menu"), Description: tr(c, "A circular portrait trigger with a native popup menu."),
		Content: func(child *ui.Context) {
			items := []overlay.MenuItem{{Key: "profile", Label: tr(child, "Profile")}, {Key: "billing", Label: tr(child, "Billing")}, {Key: "settings", Label: tr(child, "Settings")}, {Separator: true}, {Key: "logout", Label: tr(child, "Log out"), Destructive: true}}
			if chosen := p.menu(child, h, native, "compound-avatar", "Open avatar menu", &p.avatarMenu, items, func(e ui.Element) {
				control.StyleButton(child, e, "", control.ButtonProps{Variant: control.Ghost, Size: control.Icon, Label: tr(child, "Open avatar menu")}).Radius(999)
				e.Children(func() {
					display.Avatar(child, display.AvatarProps{Image: avatarPortrait, Fallback: "CN", Decorative: true})
				})
			}); chosen != "" {
				p.avatarAction = "Local account preview: " + chosen
			}
			if p.avatarAction != "" {
				ui.Text(child, p.avatarAction).Role(ui.RoleStatus)
			}
			if p.error != "" {
				ui.Text(child, tr(child, p.error)).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
}

func (p *controlCompoundsPreviewState) mailGroup(c *ui.Context, h *desktop.Host, native bool) {
	words := compoundMailWords["en"]
	direction := c.Direction()
	if p.recipe == "rtl" {
		p.choose(c, h, native, "compound-language", &p.language, control.SelectProps{Label: tr(c, "Button group language"), Options: localizedOptions(c, []control.Option{{Value: "en", Label: "English"}, {Value: "ar", Label: "العربية"}, {Value: "he", Label: "עברית"}})}).MaxWidth(200)
		words = compoundMailWords[p.language]
		direction = ui.RightToLeft
		if p.language == "en" {
			direction = ui.LeftToRight
		}
	}
	ui.Box(c).FillWidth().Direction(direction).Children(func() {
		control.ButtonGroup(c, func() {
			width, _ := c.Size()
			if width >= 640 {
				control.ButtonGroup(c, func() {
					backIcon := outlined.ArrowBack()
					if direction == ui.RightToLeft {
						backIcon = outlined.ArrowForward()
					}
					p.actionButton(c, "", "Go Back", control.ButtonProps{Variant: control.Outline, Size: control.Icon, Icon: backIcon})
				})
			}
			control.ButtonGroup(c, func() {
				p.actionButton(c, words.archive, "Archive mail", control.ButtonProps{Variant: control.Outline})
				p.actionButton(c, words.report, "Report mail", control.ButtonProps{Variant: control.Outline})
			})
			control.ButtonGroup(c, func() {
				p.actionButton(c, words.snooze, "Snooze mail", control.ButtonProps{Variant: control.Outline})
				labels := []overlay.MenuItem{{Key: "personal", Label: tr(c, words.personal), Radio: true, Selected: p.label == "personal"}, {Key: "work", Label: tr(c, words.work), Radio: true, Selected: p.label == "work"}, {Key: "other", Label: tr(c, words.other), Radio: true, Selected: p.label == "other"}}
				items := []overlay.MenuItem{{Key: "read", Label: tr(c, words.read), Icon: outlined.MarkEmailRead()}, {Key: "archive", Label: tr(c, words.archive), Icon: outlined.Archive()}, {Separator: true}, {Key: "snooze", Label: tr(c, words.snooze), Icon: outlined.Schedule()}, {Key: "calendar", Label: tr(c, words.calendar)}, {Key: "list", Label: tr(c, words.list)}, {Key: "labels", Label: tr(c, words.labels), Icon: outlined.Label(), Children: labels}, {Separator: true}, {Key: "trash", Label: tr(c, words.trash), Destructive: true, Icon: outlined.Delete()}}
				if chosen := p.iconMenu(c, h, native, "compound-mail", "Mail options", &p.more, outlined.MoreHoriz(), items); chosen != "" {
					if chosen == "personal" || chosen == "work" || chosen == "other" {
						p.label = chosen
					}
					p.action = "Local mail action: " + chosen
				}
			})
		}, control.ButtonGroupProps{Label: tr(c, "Mail action groups")})
	})
	ui.Text(c, tr(c, "Current label: ")+p.label).FontSize(13).Role(ui.RoleStatus)
}

type compoundMailText struct{ archive, report, snooze, read, calendar, list, labels, personal, work, other, trash string }

var compoundMailWords = map[string]compoundMailText{
	"en": {"Archive", "Report", "Snooze", "Mark as Read", "Add to Calendar", "Add to List", "Label As...", "Personal", "Work", "Other", "Trash"},
	"ar": {"أرشفة", "تقرير", "تأجيل", "وضع علامة كمقروء", "إضافة إلى التقويم", "إضافة إلى القائمة", "تصنيف كـ...", "شخصي", "عمل", "آخر", "سلة المهملات"},
	"he": {"ארכיון", "דוח", "דחה", "סמן כנקרא", "הוסף ליומן", "הוסף לרשימה", "תייג כ...", "אישי", "עבודה", "אחר", "פח"},
}
