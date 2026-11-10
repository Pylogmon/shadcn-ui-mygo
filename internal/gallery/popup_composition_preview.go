package gallery

import (
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Application drafts, open values and public handles only. Every child view
// receives its own current Context and builds its own native relationships.
type popupCompositionPreviewState struct {
	ready, native                        bool
	popoverRecipe, hoverRecipe, language string
	lastPopover, lastHover, lastLanguage string
	dimensions                           [4]string
	form                                 [2]string
	opens                                map[string]*bool
	popoverDialog, hoverDialog           bool
	popoverReturn, hoverReturn           ui.Handle
	popupError                           string
}

var popupPopoverRecipes = []control.Option{
	{Value: "dimensions", Label: "Four dimensions"},
	{Value: "basic", Label: "Basic header"},
	{Value: "form", Label: "Dimensions form"},
	{Value: "alignments", Label: "Three alignments"},
	{Value: "sides", Label: "Six sides"},
	{Value: "dialog", Label: "Popover in Dialog"},
	{Value: "rtl", Label: "Translated dimensions"},
}

var popupHoverRecipes = []control.Option{
	{Value: "profile", Label: "Next.js profile"},
	{Value: "sides", Label: "Four physical sides"},
	{Value: "logical", Label: "Six physical and logical sides"},
	{Value: "dialog", Label: "Hover card in Dialog"},
	{Value: "rtl", Label: "Translated product"},
}

var popupSides = []string{"inline-start", "left", "top", "bottom", "right", "inline-end"}

func (p *popupCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native = true, native
	p.popoverRecipe, p.hoverRecipe, p.language = "dimensions", "profile", "ar"
	p.lastPopover, p.lastHover, p.lastLanguage = p.popoverRecipe, p.hoverRecipe, p.language
	p.dimensions = [4]string{"100%", "300px", "25px", "none"}
	p.form = [2]string{"100%", "25px"}
	p.opens = map[string]*bool{}
}

func (p *popupCompositionPreviewState) open(key string) *bool {
	if v := p.opens[key]; v != nil {
		return v
	}
	v := false
	p.opens[key] = &v
	return &v
}

func (p *popupCompositionPreviewState) close(h *desktop.Host, prefix string) {
	for key, value := range p.opens {
		if strings.HasPrefix(key, prefix) {
			*value = false
		}
	}
	if h != nil {
		for _, suffix := range []string{"dimensions", "basic", "form", "dialog", "profile", "left", "top", "bottom", "right", "inline-start", "inline-end", "start", "center", "end"} {
			h.Close(prefix + suffix)
		}
	}
}

func (p *popupCompositionPreviewState) sync(h *desktop.Host, native bool) {
	if p.native != native || p.lastPopover != p.popoverRecipe || p.lastHover != p.hoverRecipe || p.lastLanguage != p.language {
		p.popupError = ""
	}
	if p.native != native || p.lastPopover != p.popoverRecipe || p.lastLanguage != p.language {
		p.close(h, "source-popover:")
		p.popoverDialog = false
	}
	if p.native != native || p.lastHover != p.hoverRecipe || p.lastLanguage != p.language {
		p.close(h, "source-hover:")
		p.hoverDialog = false
	}
	if p.native != native && h != nil {
		h.Close("source-popover-config")
		h.Close("source-hover-config")
		h.Close("source-popup-language")
	}
	p.native, p.lastPopover, p.lastHover, p.lastLanguage = native, p.popoverRecipe, p.hoverRecipe, p.language
}

func (p *popupCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *popupCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) (ui.Element, ui.Element) {
	p.initialize(native)
	p.sync(h, native)
	popover := display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Popover compositions"}, display.CardSlots{
		Title: "Popover compositions", Description: "Independent editable windows, alignment and nested dialogs.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-popover-config", &p.popoverRecipe, control.SelectProps{Label: "Popover configuration", Options: popupPopoverRecipes})
			if p.popoverRecipe == "rtl" || p.hoverRecipe == "rtl" {
				p.choose(child, h, native, "source-popup-language", &p.language, control.SelectProps{Label: "Popup language", Options: statusLanguages()})
			}
			p.sync(h, native)
			p.popoverExamples(child, h, native)
		},
	})
	hover := display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Hover card compositions"}, display.CardSlots{
		Title: "Hover card compositions", Description: "Source profiles, delayed previews and directional placement.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-hover-config", &p.hoverRecipe, control.SelectProps{Label: "Hover card configuration", Options: popupHoverRecipes})
			p.sync(h, native)
			p.hoverExamples(child, h, native)
			if p.popupError != "" {
				ui.Text(child, p.popupError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
	return popover, hover
}

func popupPlacement(side string, direction ui.Direction) mygo.Placement {
	switch side {
	case "inline-start":
		if direction == ui.RightToLeft {
			return mygo.PlacementRight
		}
		return mygo.PlacementLeft
	case "inline-end":
		if direction == ui.RightToLeft {
			return mygo.PlacementLeft
		}
		return mygo.PlacementRight
	case "left":
		return mygo.PlacementLeft
	case "right":
		return mygo.PlacementRight
	case "top":
		return mygo.PlacementTop
	case "start":
		if direction == ui.RightToLeft {
			return mygo.PlacementBottomEnd
		}
		return mygo.PlacementBottomStart
	case "end":
		if direction == ui.RightToLeft {
			return mygo.PlacementBottomStart
		}
		return mygo.PlacementBottomEnd
	default:
		return mygo.PlacementBottom
	}
}

func popupSideLabel(side, language string) string {
	values := map[string][6]string{
		"ar": {"بداية السطر", "يسار", "أعلى", "أسفل", "يمين", "نهاية السطر"},
		"he": {"תחילת השורה", "שמאל", "למעלה", "למטה", "ימין", "סוף השורה"},
		"en": {"Inline Start", "Left", "Top", "Bottom", "Right", "Inline End"},
	}
	v, ok := values[language]
	if !ok {
		v = values["en"]
	}
	for i, s := range popupSides {
		if s == side {
			return v[i]
		}
	}
	return side
}

func popupWords(language string) (string, string) {
	switch language {
	case "ar":
		return "الأبعاد", "تعيين الأبعاد للطبقة."
	case "he":
		return "מימדים", "הגדר את המימדים לשכבה."
	default:
		return "Dimensions", "Set the dimensions for the layer."
	}
}

func popupHeader(c *ui.Context, native bool, title, description string, role ui.Role, dimensions bool, content func(*ui.Context)) {
	gap, font := float32(4), float32(14)
	if dimensions {
		gap = 8
	}
	muted := theme.TokensOf(c).MutedForeground
	if native {
		gap, font, muted = c.Theme().Space(1), c.Theme().FontSize, c.Theme().TextMuted
	}
	group := ui.Column(c).FillWidth().MinWidth(0).Gap(16).Role(role)
	group.Children(func() {
		ui.Column(c).FillWidth().MinWidth(0).Gap(gap).Children(func() {
			level := 2
			if dimensions {
				level = 4
			}
			t := ui.Text(c, title).FontSize(font).FontWeight(500).Role(ui.RoleHeading).Level(level)
			if dimensions && !native {
				t.LineHeight(1)
			}
			d := ui.Text(c, description).FontSize(font).TextColor(muted)
			group.LabelledBy(t).DescribedBy(d)
		})
		if content != nil {
			content(c)
		}
	})
}

func (p *popupCompositionPreviewState) popover(c *ui.Context, h *desktop.Host, native bool, key, label, kind, side, language string) {
	visible, size := label, control.DefaultSize
	if kind == "alignment" {
		size = control.Small
	}
	if kind == "basic" || kind == "form" || kind == "dialog" {
		visible = "Open Popover"
	}
	if kind == "dimensions" {
		visible = "Open popover"
	}
	anchor := control.Button(c, visible, control.ButtonProps{Label: label, Variant: control.Outline, Size: size})
	if h == nil {
		if anchor.Clicked() || *p.open(key) {
			p.popupError = "Opening this popover requires a parent-bound Host."
		}
		*p.open(key) = false
		return
	}
	width, height := 288, 160
	if kind == "dimensions" {
		width, height = 320, 300
	}
	if kind == "form" {
		width, height = 256, 240
	}
	if kind == "side" || kind == "alignment" {
		width, height = 160, 100
	}
	props := desktop.PopoverProps{Label: label + " popup", Placement: popupPlacement(side, c.Direction()), Width: width, Height: height, FitContent: true, MaxHeight: 480, Gap: 4, ContentRole: ui.RoleNone}
	draw := desktop.Popover
	if native {
		draw = desktop.NativePopover
	}
	err := draw(c, h, key, anchor, p.open(key), props, func(child *ui.Context) {
		if kind == "side" || kind == "alignment" {
			text := "Popover on " + strings.ReplaceAll(side, "-", " ")
			if kind == "alignment" {
				text = "Aligned to " + side
			}
			ui.Text(child, text).FontSize(popupFont(child, native)).Role(ui.RoleDialog).Label(text)
			return
		}
		title, description := popupWords(language)
		if kind == "dialog" {
			title, description = "Popover in Dialog", "This popover appears inside a dialog. Click the button to open it."
		}
		popupHeader(child, native, title, description, ui.RoleDialog, kind == "dimensions", func(current *ui.Context) {
			if kind == "dimensions" || kind == "form" {
				p.fields(current, native, kind)
			}
		})
	})
	if err != nil {
		p.popupError = err.Error()
	}
}

func popupFont(c *ui.Context, native bool) float32 {
	if native {
		return c.Theme().FontSize
	}
	return 14
}

func (p *popupCompositionPreviewState) fields(c *ui.Context, native bool, kind string) {
	labels := []string{"Width", "Max. width", "Height", "Max. height"}
	values := []*string{&p.dimensions[0], &p.dimensions[1], &p.dimensions[2], &p.dimensions[3]}
	if kind == "form" {
		labels, values = []string{"Width", "Height"}, []*string{&p.form[0], &p.form[1]}
	}
	gap := float32(8)
	if kind == "form" {
		gap = 16
	}
	ui.Column(c).FillWidth().MinWidth(0).Gap(gap).Children(func() {
		for i, name := range labels {
			row := ui.Row(c.Key(name)).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Center)
			if kind == "dimensions" {
				row.Grid().Columns(3).Gap(16)
			}
			row.Children(func() {
				caption := ui.Text(c, name).FontSize(popupFont(c, native)).FontWeight(500).MinWidth(0)
				if kind == "form" {
					caption.WidthPercent(50).Shrink(0)
				}
				var input ui.Element
				if native {
					input = ui.TextInput(c, values[i]).Label(name)
				} else {
					input = control.Input(c, values[i], control.InputProps{Label: name})
				}
				input.LabelledBy(caption).MinWidth(0)
				if kind == "dimensions" {
					input.ColumnSpan(2).FillWidth()
					if !native {
						input.Height(32)
					}
				} else {
					input.Basis(0).Grow(1)
				}
			})
		}
	})
}

func (p *popupCompositionPreviewState) popoverExamples(c *ui.Context, h *desktop.Host, native bool) {
	switch p.popoverRecipe {
	case "dimensions", "basic", "form":
		label := "Open dimensions popover"
		if p.popoverRecipe == "basic" {
			label = "Open basic popover"
		}
		if p.popoverRecipe == "form" {
			label = "Open dimensions form"
		}
		align := "start"
		if p.popoverRecipe == "dimensions" {
			align = "center"
		}
		p.popover(c, h, native, "source-popover:"+p.popoverRecipe, label, p.popoverRecipe, align, "en")
	case "alignments":
		ui.Row(c).FillWidth().MinWidth(0).Wrap().Gap(24).Children(func() {
			for _, align := range []string{"start", "center", "end"} {
				p.popover(c, h, native, "source-popover:"+align, strings.ToUpper(align[:1])+align[1:], "alignment", align, "en")
			}
		})
	case "sides", "rtl":
		build := func() {
			groups := [][]string{popupSides[:3], popupSides[3:]}
			gap := float32(8)
			if p.popoverRecipe == "rtl" {
				groups = [][]string{{"left", "top", "bottom", "right"}, {"inline-start", "inline-end"}}
				gap = 16
			}
			ui.Column(c).FillWidth().MinWidth(0).Gap(gap).Children(func() {
				for _, sides := range groups {
					row := ui.Row(c).FillWidth().MinWidth(0).Wrap().Gap(8)
					if p.popoverRecipe == "rtl" {
						row.Justify(ui.Center)
					}
					row.Children(func() {
						for _, side := range sides {
							kind, language := "side", "en"
							if p.popoverRecipe == "rtl" {
								kind, language = "translated", p.language
							}
							p.popover(c, h, native, "source-popover:"+side, popupSideLabel(side, language), kind, side, language)
						}
					})
				}
			})
		}
		if p.popoverRecipe == "rtl" {
			ui.Column(c).FillWidth().Direction(childDirection(p.language)).Children(build)
		} else {
			build()
		}
	case "dialog":
		anchor := control.Button(c, "Open Dialog", control.ButtonProps{Label: "Open popover dialog", Variant: control.Outline}).Bind(&p.popoverReturn)
		if anchor.Clicked() {
			p.popoverDialog = true
		}
		overlay.Dialog(c, &p.popoverDialog, overlay.DialogProps{Title: "Popover Example", Description: "Click the button below to see the popover.", ReturnFocus: &p.popoverReturn}, func() {
			p.popover(c, h, native, "source-popover:dialog", "Open popover in dialog", "dialog", "start", "en")
		})
	}
}

func (p *popupCompositionPreviewState) hover(c *ui.Context, h *desktop.Host, native bool, key, label, kind, side, language string) {
	variant := control.Outline
	if kind == "profile" {
		variant = control.Link
	}
	anchor := control.Button(c, label, control.ButtonProps{Variant: variant})
	delay := 100 * time.Millisecond
	if kind == "profile" || kind == "product" {
		delay = 10 * time.Millisecond
	}
	description := "This hover card appears on the " + strings.ReplaceAll(side, "-", " ") + " side of the trigger."
	if kind == "profile" {
		description = "The React Framework – created and maintained by @vercel. Joined December 2021"
	}
	if kind == "dialog" {
		description = "This hover card appears inside a dialog. Hover over the button to see it."
	}
	if kind == "product" {
		description = popupProductName(language) + " " + popupProductPrice(language)
	}
	props := desktop.HoverCardProps{Label: "Source hover card", Description: description, Placement: popupPlacement(side, c.Direction()), Width: 256, Height: 180, MaxHeight: 480, Gap: 4, OpenDelay: delay, CloseDelay: 100 * time.Millisecond}
	draw := desktop.HoverCard
	if native {
		draw = desktop.NativeHoverCard
	}
	err := draw(c, h, key, anchor, props, func(child *ui.Context) {
		font := popupFont(child, native)
		muted := theme.TokensOf(child).MutedForeground
		if native {
			muted = child.Theme().TextMuted
		}
		if kind == "profile" {
			ui.Column(child).FillWidth().MinWidth(0).Gap(2).Role(ui.RoleGroup).Label("Next.js profile").Children(func() {
				ui.Text(child, "@nextjs").FontSize(font).FontWeight(600)
				ui.Text(child, "The React Framework – created and maintained by @vercel.").FontSize(font)
				ui.Text(child, "Joined December 2021").FontSize(12).TextColor(muted).Margin(4, 0, 0, 0)
			})
		} else if kind == "product" {
			ui.Column(child).FillWidth().MinWidth(0).Gap(4).Role(ui.RoleGroup).Label(description).Children(func() {
				ui.Text(child, popupProductName(language)).FontSize(font).FontWeight(600)
				ui.Text(child, popupProductPrice(language)).FontSize(font).TextColor(muted)
			})
		} else {
			gap := float32(8)
			if kind == "docs-side" {
				gap = 4
			}
			ui.Column(child).FillWidth().MinWidth(0).Gap(gap).Role(ui.RoleGroup).Label("Hover Card").Children(func() {
				ui.Text(child, "Hover Card").FontSize(font).FontWeight(500)
				ui.Text(child, description).FontSize(font)
			})
		}
	})
	if err != nil {
		p.popupError = err.Error()
	}
}

func popupProductName(language string) string {
	switch language {
	case "ar":
		return "سماعات لاسلكية"
	case "he":
		return "אוזניות אלחוטיות"
	default:
		return "Wireless Headphones"
	}
}
func popupProductPrice(language string) string {
	switch language {
	case "ar":
		return "٩٩.٩٩ $"
	case "he":
		return "99.99 $"
	default:
		return "$99.99"
	}
}

func (p *popupCompositionPreviewState) hoverExamples(c *ui.Context, h *desktop.Host, native bool) {
	switch p.hoverRecipe {
	case "profile":
		p.hover(c, h, native, "source-hover:profile", "Hover Here", "profile", "bottom", "en")
	case "sides", "logical", "rtl":
		build := func() {
			groups := [][]string{popupSides}
			if p.hoverRecipe == "sides" {
				groups = [][]string{{"left", "top", "bottom", "right"}}
			}
			if p.hoverRecipe == "rtl" {
				groups = [][]string{{"left", "top", "bottom", "right"}, {"inline-start", "inline-end"}}
			}
			ui.Column(c).FillWidth().MinWidth(0).Gap(16).Children(func() {
				for _, sides := range groups {
					ui.Row(c).FillWidth().MinWidth(0).Wrap().Gap(8).Justify(ui.Center).Children(func() {
						for _, side := range sides {
							kind, language := "side", "en"
							if p.hoverRecipe == "sides" {
								kind = "docs-side"
							}
							if p.hoverRecipe == "rtl" {
								kind, language = "product", p.language
							}
							p.hover(c, h, native, "source-hover:"+side, popupSideLabel(side, language), kind, side, language)
						}
					})
				}
			})
		}
		if p.hoverRecipe == "rtl" {
			ui.Column(c).FillWidth().Direction(childDirection(p.language)).Children(build)
		} else {
			build()
		}
	case "dialog":
		anchor := control.Button(c, "Open Dialog", control.ButtonProps{Label: "Open hover card dialog", Variant: control.Outline}).Bind(&p.hoverReturn)
		if anchor.Clicked() {
			p.hoverDialog = true
		}
		overlay.Dialog(c, &p.hoverDialog, overlay.DialogProps{Title: "Hover Card Example", Description: "Hover over the button below to see the hover card.", ReturnFocus: &p.hoverReturn}, func() {
			p.hover(c, h, native, "source-hover:dialog", "Hover me", "dialog", "bottom", "en")
		})
	}
}
