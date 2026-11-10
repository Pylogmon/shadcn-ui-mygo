package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Keycaps are inert visuals. Only values and counters survive the current
// build; native editors and tooltip child Contexts are never retained here.
type kbdCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	drafts                               map[string]*string
	changes                              map[string]int
	accepted                             int
}

var kbdCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Modifier rows (docs)"},
	{Value: "group", Label: "Inline shortcut group"},
	{Value: "button", Label: "Accept button"},
	{Value: "tooltip", Label: "Save and Print tooltips"},
	{Value: "input-group", Label: "Search input group"},
	{Value: "rtl", Label: "RTL modifier rows"},
	{Value: "usage", Label: "Guide: Control key"},
	{Value: "usage-group", Label: "Guide: Control and B"},
	{Value: "create-basic", Label: "Create: Basic"},
	{Value: "create-modifiers", Label: "Create: Modifier Keys"},
	{Value: "create-group", Label: "Create: KbdGroup"},
	{Value: "create-arrows", Label: "Create: Arrow Keys"},
	{Value: "create-icons", Label: "Create: With Icons"},
	{Value: "create-icon-text", Label: "Create: With Icons and Text"},
	{Value: "create-input", Label: "Create: InputGroup"},
	{Value: "create-tooltip", Label: "Create: Tooltip"},
	{Value: "create-samp", Label: "Create: With samp"},
}

func (p *kbdCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.drafts, p.changes = map[string]*string{}, map[string]int{}
}

func (p *kbdCompositionPreviewState) draft() *string {
	if value := p.drafts[p.recipe]; value != nil {
		return value
	}
	value := ""
	p.drafts[p.recipe] = &value
	return &value
}

func (p *kbdCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *kbdCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			for _, key := range []string{"source-kbd-config", "source-kbd-language", "source-kbd-save", "source-kbd-print", "source-kbd-create-save"} {
				h.Close(key)
			}
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Kbd compositions"}, display.CardSlots{
		Title: "Kbd compositions", Description: "Source keycaps, inert shortcut notation and real button, input and tooltip compositions.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-kbd-config", &p.recipe, control.SelectProps{Label: "Kbd configuration", Options: kbdCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-kbd-language", &p.language, control.SelectProps{Label: "Kbd language", Options: statusLanguages()})
			}
			preview := ui.Column(child.Key("kbd-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
			if p.recipe == "rtl" {
				preview.Direction(childDirection(p.language))
			}
			preview.Children(func() { p.draw(child, h, native) })
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert)
			}
		},
	})
}

func kbdSourceGroup(c *ui.Context, keys ...string) ui.Element {
	return display.KbdGroup(c, func(child *ui.Context) {
		for _, key := range keys {
			display.Kbd(child, key)
		}
	}).AlignSelf(ui.Center)
}

func (p *kbdCompositionPreviewState) tooltip(c *ui.Context, h *desktop.Host, native bool, anchor ui.Element, key, text string, keys []string, create bool) {
	description := text
	for _, key := range keys {
		description += " " + key
	}
	props := desktop.TooltipProps{Description: description, Gap: 4, MaxWidth: 240}
	// Only copied text/keys and presentation values enter the SDK-owned child
	// provider. The child builds using its own current Context.
	content := func(child *ui.Context) {
		gap := float32(4)
		if create {
			gap = 8
		}
		row := ui.Row(child).Gap(gap).AlignItems(ui.Center)
		if create && !native {
			row.MarginEnd(-6)
		}
		row.Children(func() {
			ui.Text(child, text).SingleLine()
			display.KbdGroup(child, func(current *ui.Context) {
				for _, key := range keys {
					display.KbdWith(current, key, display.KbdProps{InTooltip: !native})
				}
			})
		})
	}
	var err error
	if native {
		err = desktop.NativeTooltipWith(c, h, key, anchor, props, content)
	} else {
		err = desktop.TooltipWith(c, h, key, anchor, props, content)
	}
	if err != nil {
		p.popupError = err.Error()
	}
}

func (p *kbdCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(560).AlignItems(ui.Center).Gap(16).Children(func() {
		switch p.recipe {
		case "demo", "rtl":
			kbdSourceGroup(c, "⌘", "⇧", "⌥", "⌃")
			display.KbdGroup(c, func(child *ui.Context) { display.Kbd(child, "Ctrl"); ui.Text(child, "+"); display.Kbd(child, "B") }).AlignSelf(ui.Center)
		case "group":
			ui.Row(c).Wrap().Gap(4).AlignItems(ui.Center).FontSize(14).TextColor(theme.TokensOf(c).MutedForeground).Children(func() {
				ui.Text(c, "Use")
				kbdSourceGroup(c, "Ctrl + B", "Ctrl + K")
				ui.Text(c, "to open the command palette")
			})
		case "usage":
			display.Kbd(c, "Ctrl").AlignSelf(ui.Center)
		case "usage-group":
			kbdSourceGroup(c, "Ctrl", "B")
		case "button":
			button := control.Button(c, "", control.ButtonProps{Label: "Accept", Variant: control.Outline}).PaddingEnd(8)
			button.Children(func() { ui.Text(c, "Accept"); display.Kbd(c, "⏎").Translate(2, 0) })
			if button.Clicked() {
				p.accepted++
				p.action = "Local Kbd action: Accept"
			}
		case "tooltip":
			control.ButtonGroup(c, func() {
				save := control.Button(c, "Save", control.ButtonProps{Variant: control.Outline})
				p.tooltip(c, h, native, save, "source-kbd-save", "Save Changes", []string{"S"}, false)
				print := control.Button(c, "Print", control.ButtonProps{Variant: control.Outline})
				p.tooltip(c, h, native, print, "source-kbd-print", "Print Document", []string{"Ctrl", "P"}, false)
			})
		case "input-group", "create-input":
			ui.Column(c).FillWidth().MinWidth(0).MaxWidth(320).Gap(24).Children(func() {
				props := control.InputGroupProps{Label: "Kbd source input"}
				placeholder := ""
				if p.recipe == "input-group" {
					placeholder = "Search..."
					props.InlineStart = func() { ui.Icon(c, outlined.Search()).Size(16, 16).AccessHidden(true) }
					props.InlineEnd = func() { display.Kbd(c, "⌘"); display.Kbd(c, "K") }
				} else {
					props.InlineStart = func() { display.Kbd(c, "Space") }
				}
				control.InputGroup(c, props, func() ui.Element {
					return control.Input(c, p.draft(), control.InputProps{Label: "Kbd source input", Placeholder: placeholder, OnChange: func(string) { p.changes[p.recipe]++ }})
				})
			})
		case "create-basic":
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				for _, key := range []string{"Ctrl", "⌘K", "Ctrl + B"} {
					display.Kbd(c, key)
				}
			})
		case "create-modifiers":
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() { display.Kbd(c, "⌘"); display.Kbd(c, "C") })
		case "create-group":
			kbdSourceGroup(c, "Ctrl", "Shift", "P")
		case "create-arrows":
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				for _, key := range []string{"↑", "↓", "←", "→"} {
					display.Kbd(c, key)
				}
			})
		case "create-icons":
			display.KbdGroup(c, func(child *ui.Context) {
				display.KbdWith(child, "", display.KbdProps{Label: "Dashed circle", Icon: outlined.MotionPhotosOn()})
				display.KbdWith(child, "", display.KbdProps{Label: "Arrow left", Icon: outlined.ArrowBack()})
				display.KbdWith(child, "", display.KbdProps{Label: "Arrow right", Icon: outlined.ArrowForward()})
			}).AlignSelf(ui.Center)
		case "create-icon-text":
			display.KbdGroup(c, func(child *ui.Context) {
				display.KbdWith(child, "Left", display.KbdProps{Icon: outlined.ArrowBack()})
				display.KbdWith(child, "Voice Enabled", display.KbdProps{Icon: outlined.MotionPhotosOn()})
			}).AlignSelf(ui.Center)
		case "create-tooltip":
			button := control.Button(c, "", control.ButtonProps{Label: "Save changes", Variant: control.Outline, Size: control.IconSmall, Icon: outlined.Save()})
			p.tooltip(c, h, native, button, "source-kbd-create-save", "Save Changes", []string{"S"}, true)
		case "create-samp":
			display.KbdWith(c, "", display.KbdProps{Label: "File", Content: func(child *ui.Context) { ui.Text(child, "File").Font("monospace").SingleLine() }}).AlignSelf(ui.Center)
		}
	})
}
