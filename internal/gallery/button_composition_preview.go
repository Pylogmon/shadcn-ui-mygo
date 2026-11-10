package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// This model holds application values only. Native targets, visual elements and
// router view callbacks are built with the current Context, never retained here.
type buttonCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	actions                              map[string]string
	counts                               map[string]int
}

var buttonCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Demo"},
	{Value: "size", Label: "Sizes"},
	{Value: "default", Label: "Default"},
	{Value: "outline", Label: "Outline"},
	{Value: "secondary", Label: "Secondary"},
	{Value: "ghost", Label: "Ghost"},
	{Value: "destructive", Label: "Destructive"},
	{Value: "link", Label: "Link variant"},
	{Value: "with-icon", Label: "With Icon"},
	{Value: "icon", Label: "Icon"},
	{Value: "rounded", Label: "Rounded"},
	{Value: "spinner", Label: "Loading"},
	{Value: "render", Label: "Render as Link"},
	{Value: "rtl", Label: "Translated buttons"},
	{Value: "create-variants", Label: "Create: Variants & Sizes"},
	{Value: "create-icon-right", Label: "Create: Icon Right"},
	{Value: "create-icon-left", Label: "Create: Icon Left"},
	{Value: "create-icon-only", Label: "Create: Icon Only"},
	{Value: "create-invalid", Label: "Create: Invalid States"},
	{Value: "create-examples", Label: "Create: Examples"},
}

type buttonCompositionSpec struct {
	text, label                                       string
	variant                                           control.Variant
	size                                              control.Size
	icon                                              *ui.SVG
	end, spin, disabled, invalid, round, mirror, link bool
	// DOM order and data-icon need not agree in the original examples.
	inlineStart, unmarked, underlined bool
}

func (p *buttonCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.actions, p.counts = map[string]string{}, map[string]int{}
}

func (p *buttonCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
		return
	}
	if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *buttonCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("button-source-configuration")
			h.Close("button-source-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Button compositions"}, display.CardSlots{
		Title: "Button compositions", Description: "Source variants, sizes, icon positions and native button or link targets.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "button-source-configuration", &p.recipe, control.SelectProps{Label: "Button configuration", Options: buttonCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "button-source-language", &p.language, control.SelectProps{Label: "Button language", Options: statusLanguages()})
			}
			p.action = p.actions[p.recipe]
			ui.Column(child.Key("button-source:" + p.recipe)).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
				if p.recipe == "rtl" {
					ui.Column(child).Direction(childDirection(p.language)).Children(func() { p.draw(child) })
				} else {
					p.draw(child)
				}
			})
			if p.action != "" {
				ui.Text(child, p.action).Role(ui.RoleStatus).FontSize(13)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert).TextColor(theme.TokensOf(child).Destructive)
			}
		},
	})
}

func buttonCompositionContent(c *ui.Context, s buttonCompositionSpec) {
	drawIcon := func() {
		if s.spin {
			indicator.Spinner(c, indicator.Props{Decorative: true})
			return
		}
		if s.icon == nil {
			return
		}
		size := float32(16)
		if s.size == control.ExtraSmall || s.size == control.IconExtraSmall {
			size = 12
		}
		e := ui.Icon(c, s.icon).Size(size, size).Shrink(0).IgnoreInput(true).AccessHidden(true)
		if s.mirror && c.Direction() == ui.RightToLeft {
			e.Rotate(180)
		}
	}
	if !s.end {
		drawIcon()
	}
	if s.text != "" {
		text := ui.Text(c, s.text).SingleLine().AccessHidden(true)
		if s.underlined {
			text.Underline().UnderlineOffset(4)
		}
	}
	if s.end {
		drawIcon()
	}
}

func (p *buttonCompositionPreviewState) activate(label string, link bool) {
	kind := "action"
	if link {
		kind = "link"
	}
	p.action = "Local button " + kind + ": " + label
	p.actions[p.recipe] = p.action
	p.counts[p.recipe]++
}

func (p *buttonCompositionPreviewState) button(c *ui.Context, key string, s buttonCompositionSpec) ui.Element {
	label := s.label
	if label == "" {
		label = s.text
	}
	props := control.ButtonProps{Variant: s.variant, Size: s.size, Label: label, Disabled: s.disabled, Invalid: s.invalid}
	if s.link {
		return p.link(c.Key(key), s, props)
	}
	b := control.StyleButton(c, ui.ButtonBase(c.Key(key)), "", props).Busy(s.spin)
	iconOnly := s.size == control.Icon || s.size == control.IconExtraSmall || s.size == control.IconSmall || s.size == control.IconLarge
	if (s.icon != nil || s.spin) && !iconOnly && !s.unmarked {
		padding := float32(8)
		if s.size == control.ExtraSmall || s.size == control.Small {
			padding = 6
		}
		if s.end && !s.inlineStart {
			b.PaddingEnd(padding)
		} else {
			b.PaddingStart(padding)
		}
	}
	s.underlined = s.variant == control.Link && b.Hovered() && !b.IsDisabled()
	if s.round {
		b.Radius(999)
	}
	if s.spin {
		b.Cursor(ui.CursorWait)
	}
	b.Children(func() { buttonCompositionContent(c, s) })
	if b.Clicked() {
		p.activate(label, false)
	}
	return b
}

// MyGo Link is text-only. An actual native Link draws the styled target under
// an inert button face. A public Router, held by SDK ui.Local, consumes the
// original href="#" as a same-page route, without opening an external browser.
func (p *buttonCompositionPreviewState) link(c *ui.Context, s buttonCompositionSpec, props control.ButtonProps) ui.Element {
	owner := ui.Box(c).Shrink(0)
	router := ui.Local(owner, "button-hash-router", func() *ui.Router { return ui.NewRouter("/button-examples") })
	owner.Children(func() {
		(*router).View(c, func(_ *ui.Route) {
			ui.Box(c).AlignSelf(ui.Start).Children(func() {
				target := control.StyleButton(c.Key("target"), ui.Link(c.Key("target"), "", "#"), "", props).Absolute().Fill().Label(s.text)
				control.StyleButton(c.Key("face"), ui.Row(c.Key("face")), "", props).Background(ui.Transparent).Border(0, ui.Transparent).
					ZIndex(1).IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false).Children(func() { buttonCompositionContent(c, s) })
				if target.Clicked() {
					p.activate(s.text, true)
				}
			})
		}).Grow(0).AlignSelf(ui.Start)
	})
	return owner
}

func (p *buttonCompositionPreviewState) row(c *ui.Context, key string, items ...buttonCompositionSpec) {
	ui.Row(c.Key(key)).Gap(8).Wrap().Justify(ui.Center).AlignItems(ui.Center).Children(func() {
		for i, s := range items {
			p.button(c, key+":"+string(rune('a'+i)), s)
		}
	})
}

var buttonSourceSizes = []struct {
	size, icon control.Size
	label      string
}{
	{control.ExtraSmall, control.IconExtraSmall, "extra small"}, {control.Small, control.IconSmall, "small"},
	{control.DefaultSize, control.Icon, "default"}, {control.Large, control.IconLarge, "large"},
}
var buttonSourceVariants = []string{"Default", "Secondary", "Outline", "Ghost", "Destructive", "Link"}

func (p *buttonCompositionPreviewState) matrix(c *ui.Context, kind string) {
	for i, size := range buttonSourceSizes {
		var items []buttonCompositionSpec
		for v, text := range buttonSourceVariants {
			s := buttonCompositionSpec{text: text, label: text + " (" + size.label + ")", variant: control.Variant(v), size: size.size}
			switch kind {
			case "create-icon-right":
				s.icon, s.end = outlined.ArrowForward(), true
				// This one source ArrowRight has no data-icon marker.
				s.unmarked = i == 1 && v == int(control.Outline)
			case "create-icon-left":
				s.icon = outlined.ArrowCircleLeft()
			case "create-icon-only":
				s.text, s.icon, s.size, s.label = "", outlined.ArrowForward(), size.icon, text+" arrow button ("+size.label+")"
			case "create-invalid":
				s.invalid = true
			}
			items = append(items, s)
		}
		p.row(c, "matrix-"+string(rune('a'+i)), items...)
	}
}

func (p *buttonCompositionPreviewState) draw(c *ui.Context) {
	if strings.HasPrefix(p.recipe, "create-") && p.recipe != "create-examples" {
		p.matrix(c, p.recipe)
		return
	}
	switch p.recipe {
	case "demo":
		p.row(c, "demo", buttonCompositionSpec{text: "Button", variant: control.Outline}, buttonCompositionSpec{label: "Submit", variant: control.Outline, size: control.Icon, icon: outlined.ArrowUpward()})
	case "size":
		group := ui.Row(c).Gap(32).AlignItems(ui.Start)
		if w, _ := c.Size(); w < 640 {
			group.Column()
		}
		group.Children(func() {
			for i, size := range buttonSourceSizes {
				label := []string{"Extra Small", "Small", "Default", "Large"}[i]
				p.row(c, "size-"+size.label, buttonCompositionSpec{text: label, variant: control.Outline, size: size.size}, buttonCompositionSpec{label: "Submit (" + size.label + ")", variant: control.Outline, size: size.icon, icon: outlined.ArrowOutward()})
			}
		})
	case "default":
		p.row(c, "default", buttonCompositionSpec{text: "Button"})
	case "outline":
		p.row(c, "outline", buttonCompositionSpec{text: "Outline", variant: control.Outline})
	case "secondary":
		p.row(c, "secondary", buttonCompositionSpec{text: "Secondary", variant: control.Secondary})
	case "ghost":
		p.row(c, "ghost", buttonCompositionSpec{text: "Ghost", variant: control.Ghost})
	case "destructive":
		p.row(c, "destructive", buttonCompositionSpec{text: "Destructive", variant: control.Destructive})
	case "link":
		p.row(c, "link", buttonCompositionSpec{text: "Link", variant: control.Link})
	case "with-icon":
		p.row(c, "with-icon", buttonCompositionSpec{text: "New Branch", variant: control.Outline, icon: outlined.AccountTree()}, buttonCompositionSpec{text: "Fork", variant: control.Outline, icon: outlined.ForkRight(), end: true})
	case "icon":
		p.row(c, "icon", buttonCompositionSpec{label: "Upload", variant: control.Outline, size: control.Icon, icon: outlined.ArrowCircleUp()})
	case "rounded":
		p.row(c, "rounded", buttonCompositionSpec{text: "Get Started", round: true}, buttonCompositionSpec{label: "Submit rounded", variant: control.Outline, size: control.Icon, icon: outlined.ArrowUpward(), round: true})
	case "spinner":
		p.row(c, "spinner", buttonCompositionSpec{text: "Generating", variant: control.Outline, spin: true, disabled: true}, buttonCompositionSpec{text: "Downloading", variant: control.Secondary, spin: true, end: true, inlineStart: true, disabled: true})
	case "render":
		p.row(c, "render", buttonCompositionSpec{text: "Login", variant: control.Secondary, size: control.Small, link: true})
	case "rtl":
		button, submit, delete, loading := buttonCompositionWords(p.language)
		p.row(c, "rtl", buttonCompositionSpec{text: button, variant: control.Outline}, buttonCompositionSpec{text: delete, variant: control.Destructive}, buttonCompositionSpec{text: submit, variant: control.Outline, icon: outlined.ArrowForward(), end: true, mirror: true}, buttonCompositionSpec{label: "Add", variant: control.Outline, size: control.Icon, icon: outlined.Add()}, buttonCompositionSpec{text: loading, variant: control.Secondary, spin: true, disabled: true})
	case "create-examples":
		ui.Row(c).Gap(16).Wrap().Justify(ui.Center).AlignItems(ui.Center).Children(func() {
			p.row(c, "examples-save", buttonCompositionSpec{text: "Cancel", variant: control.Outline}, buttonCompositionSpec{text: "Submit", icon: outlined.ArrowForward(), end: true})
			p.row(c, "examples-delete", buttonCompositionSpec{text: "Delete", variant: control.Destructive}, buttonCompositionSpec{label: "Continue", size: control.Icon, icon: outlined.ArrowForward()})
			p.button(c, "examples-link", buttonCompositionSpec{text: "Link", link: true})
		})
	}
}

func buttonCompositionWords(language string) (button, submit, delete, loading string) {
	switch language {
	case "ar":
		return "زر", "إرسال", "حذف", "جاري التحميل"
	case "he":
		return "כפתור", "שלח", "מחק", "טוען"
	default:
		return "Button", "Submit", "Delete", "Loading"
	}
}
