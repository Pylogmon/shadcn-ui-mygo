package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

// Only application values survive a build. Field scopes, targets and callbacks
// are constructed with the current view, including inside a Flyout selector.
type switchCompositionPreviewState struct {
	ready, native                     bool
	recipe, language, action, lastKey string
	popupError                        string
	values                            map[string]*bool
	changes                           int
}

var switchCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Airplane Mode (docs)"},
	{Value: "description", Label: "Description (docs)"},
	{Value: "choice-card", Label: "Choice Cards"},
	{Value: "disabled", Label: "Disabled (docs)"},
	{Value: "invalid", Label: "Invalid Terms"},
	{Value: "sizes", Label: "Sizes (docs)"},
	{Value: "rtl", Label: "Translated Switch"},
	{Value: "create-basic", Label: "Create: Basic"},
	{Value: "create-description", Label: "Create: Description Choice Card"},
	{Value: "create-label", Label: "Create: Bluetooth Label"},
	{Value: "create-disabled", Label: "Create: Disabled States"},
	{Value: "create-sizes", Label: "Create: Sizes"},
}

func (p *switchCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.values = map[string]*bool{}
}

func (p *switchCompositionPreviewState) value(key string, initial bool) *bool {
	if value := p.values[key]; value != nil {
		return value
	}
	value := initial
	p.values[key] = &value
	return &value
}

func (p *switchCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *switchCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-switch-config")
			h.Close("source-switch-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: tr(c, "Switch compositions")}, display.CardSlots{
		Title: tr(c, "Switch compositions"), Description: tr(c, "Source labels, descriptions, choice cards and native switch state."),
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-switch-config", &p.recipe, control.SelectProps{Label: tr(child, "Switch configuration"), Options: localizedOptions(child, switchCompositionRecipes)})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-switch-language", &p.language, control.SelectProps{Label: tr(child, "Switch language"), Options: localizedOptions(child, statusLanguages())})
			}
			preview := ui.Column(child.Key("switch-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
			if p.recipe == "rtl" {
				preview.Direction(childDirection(p.language))
			}
			preview.Children(func() { p.draw(child) })
			if p.action != "" {
				ui.Text(child, tr(child, p.action)).FontSize(13).Role(ui.RoleStatus).Label(tr(child, p.action))
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert)
			}
		},
	})
}

type switchSourceRow struct {
	key, label, description           string
	initial, disabled, invalid, small bool
	field, card, controlLast          bool
}

const switchSourceDescription = "Focus is shared across devices, and turns off when you leave the app."

func (p *switchCompositionPreviewState) rows() []switchSourceRow {
	switch p.recipe {
	case "description", "rtl":
		label, description := "Share across devices", switchSourceDescription
		if p.recipe == "rtl" {
			switch p.language {
			case "ar":
				label, description = "المشاركة عبر الأجهزة", "يتم مشاركة التركيز عبر الأجهزة، ويتم إيقاف تشغيله عند مغادرة التطبيق."
			case "he":
				label, description = "שיתוף בין מכשירים", "המיקוד משותף בין מכשירים, וכבה כשאתה עוזב את האפליקציה."
			}
		}
		return []switchSourceRow{{key: "share", label: label, description: description, field: true, controlLast: true}}
	case "choice-card":
		return []switchSourceRow{
			{key: "share", label: "Share across devices", description: switchSourceDescription, field: true, card: true, controlLast: true},
			{key: "notifications", label: "Enable notifications", description: "Receive notifications when focus mode is enabled or disabled.", initial: true, field: true, card: true, controlLast: true},
		}
	case "disabled":
		return []switchSourceRow{{key: "disabled", label: "Disabled", disabled: true, field: true}}
	case "invalid":
		return []switchSourceRow{{key: "terms", label: "Accept terms and conditions", description: "You must accept the terms and conditions to continue.", invalid: true, field: true, controlLast: true}}
	case "sizes":
		return []switchSourceRow{{key: "small", label: "Small", small: true, field: true}, {key: "default", label: "Default", field: true}}
	case "create-basic":
		return []switchSourceRow{{key: "airplane", label: "Airplane Mode", field: true}}
	case "create-description":
		return []switchSourceRow{{key: "share", label: "Share across devices", description: switchSourceDescription, field: true, card: true, controlLast: true}}
	case "create-label":
		return []switchSourceRow{{key: "bluetooth", label: "Bluetooth", initial: true}}
	case "create-disabled":
		return []switchSourceRow{{key: "unchecked", label: "Disabled (Unchecked)", disabled: true}, {key: "checked", label: "Disabled (Checked)", initial: true, disabled: true}}
	case "create-sizes":
		return []switchSourceRow{{key: "small", label: "Small", small: true}, {key: "default", label: "Default"}}
	default:
		return []switchSourceRow{{key: "airplane", label: "Airplane Mode"}}
	}
}

func (p *switchCompositionPreviewState) draw(c *ui.Context) {
	rows := p.rows()
	group := ui.Column(c).MinWidth(0).AlignSelf(ui.Center).Gap(28)
	switch p.recipe {
	case "description", "choice-card", "invalid", "rtl", "create-description":
		group.FillWidth().MaxWidth(384)
	case "sizes":
		group.FillWidth().MaxWidth(160)
	case "create-disabled", "create-sizes":
		group.Gap(48)
	}
	group.Children(func() {
		for _, row := range rows {
			p.row(c.Key(row.key), row)
		}
	})
}

func (p *switchCompositionPreviewState) row(c *ui.Context, row switchSourceRow) {
	key := p.recipe + ":" + row.key
	value := p.value(key, row.initial)
	props := control.SwitchProps{Disabled: row.disabled, Invalid: row.invalid, OnCheckedChange: func(bool) {
		p.lastKey, p.action = key, "Local switch action: "+row.label
		p.changes++
	}}
	if row.small {
		props.Size = control.Small
	}
	if !row.field {
		props.Label = row.label
		control.Switch(c, value, props).LineHeight(1).FontWeight(500)
		return
	}
	control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal, Disabled: row.disabled, Invalid: row.invalid}, func(f *control.FieldScope) {
		buildControl := func() { f.Control(c, func() ui.Element { return control.Switch(c, value, props) }) }
		buildCaption := func() {
			if row.description == "" {
				f.Label(c, control.FieldLabelProps{Text: tr(c, row.label)}, nil)
				return
			}
			f.Content(c, func() {
				f.Label(c, control.FieldLabelProps{Text: tr(c, row.label)}, nil)
				f.Description(c, tr(c, row.description))
			})
		}
		if row.card {
			f.Label(c, control.FieldLabelProps{Label: tr(c, row.label), Choice: true, Selected: *value}, func() {
				ui.Row(c).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Start).Children(func() {
					f.Content(c, func() { control.FieldTitle(c, tr(c, row.label)); f.Description(c, tr(c, row.description)) })
					buildControl()
				})
			})
		} else if row.controlLast {
			buildCaption()
			buildControl()
		} else {
			buildControl()
			buildCaption()
		}
	})
}
