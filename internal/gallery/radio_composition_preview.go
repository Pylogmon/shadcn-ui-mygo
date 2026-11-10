package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Application-owned drafts only. Every Field scope, Element and callback is
// constructed with the current view Context and is consumed during that build.
type radioCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	values                               map[string]*string
	changes                              map[string]int
}

var radioCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Spacing"},
	{Value: "description", Label: "Spacing descriptions"},
	{Value: "choice-card", Label: "Plan choice cards"},
	{Value: "fieldset", Label: "Subscription plan fieldset"},
	{Value: "disabled", Label: "One disabled option"},
	{Value: "invalid", Label: "Invalid notification preferences"},
	{Value: "rtl", Label: "Translated spacing descriptions"},
	{Value: "create-basic", Label: "Create: Basic"},
	{Value: "create-descriptions", Label: "Create: With Descriptions"},
	{Value: "create-fieldset", Label: "Create: With FieldSet"},
	{Value: "create-grid", Label: "Create: Grid Layout"},
	{Value: "create-disabled", Label: "Create: Disabled Group"},
	{Value: "create-invalid", Label: "Create: Invalid"},
}

type radioSourceOption struct {
	value, label, description string
	disabled                  bool
}
type radioSourceRecipe struct {
	initial, legend, description  string
	options                       []radioSourceOption
	card, grid, invalid, disabled bool
	legendLabel, medium           bool
	width                         float32
}

func radioSourceRecipeFor(key, language string) radioSourceRecipe {
	spacing := []radioSourceOption{{value: "default", label: "Default"}, {value: "comfortable", label: "Comfortable"}, {value: "compact", label: "Compact"}}
	r := radioSourceRecipe{initial: "comfortable", options: spacing, width: 384}
	switch key {
	case "demo":
		r.medium = true
	case "description", "rtl":
		words := radioSourceWords(language)
		if key != "rtl" {
			words = radioSourceWords("en")
		}
		for i, value := range []string{"default", "comfortable", "compact"} {
			r.options[i].label, r.options[i].description = words[value], words[value+"Description"]
		}
		r.medium = true
	case "choice-card", "create-descriptions":
		r.initial, r.card, r.medium = "plus", true, true
		r.options = []radioSourceOption{
			{value: "plus", label: "Plus", description: "For individuals and small teams"},
			{value: "pro", label: "Pro", description: "For growing businesses"},
			{value: "enterprise", label: "Enterprise", description: "For large teams and enterprises"},
		}
		if key == "choice-card" {
			for i := range r.options {
				r.options[i].description += "."
			}
		}
	case "fieldset":
		r.initial, r.legend, r.description, r.legendLabel, r.width = "monthly", "Subscription Plan", "Yearly and lifetime plans offer significant savings.", true, 320
		r.options = []radioSourceOption{{value: "monthly", label: "Monthly ($9.99/month)"}, {value: "yearly", label: "Yearly ($99.99/year)"}, {value: "lifetime", label: "Lifetime ($299.99)"}}
	case "disabled":
		r.initial = "option2"
		r.options = []radioSourceOption{{value: "option1", label: "Disabled", disabled: true}, {value: "option2", label: "Option 2"}, {value: "option3", label: "Option 3"}}
	case "invalid", "create-invalid":
		r.initial, r.legend, r.description, r.invalid = "email", "Notification Preferences", "Choose how you want to receive notifications.", true
		r.legendLabel = key == "invalid"
		if key == "invalid" {
			r.width = 320
		}
		r.options = []radioSourceOption{{value: "email", label: "Email only"}, {value: "sms", label: "SMS only"}, {value: "both", label: "Both Email & SMS"}}
	case "create-fieldset":
		r.initial, r.legend, r.description = "medium", "Battery Level", "Choose your preferred battery level."
		r.options = []radioSourceOption{{value: "high", label: "High"}, {value: "medium", label: "Medium"}, {value: "low", label: "Low"}}
	case "create-grid":
		r.initial, r.card, r.grid, r.medium = "medium", true, true, true
		r.options = []radioSourceOption{{value: "small", label: "Small"}, {value: "medium", label: "Medium"}, {value: "large", label: "Large"}, {value: "xlarge", label: "X-Large"}}
	case "create-disabled":
		r.initial, r.disabled = "option2", true
		r.options = []radioSourceOption{{value: "option1", label: "Option 1"}, {value: "option2", label: "Option 2"}, {value: "option3", label: "Option 3"}}
	}
	return r
}

func (p *radioCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.values, p.changes = map[string]*string{}, map[string]int{}
}

func (p *radioCompositionPreviewState) selected(key string) *string {
	if value := p.values[key]; value != nil {
		return value
	}
	value := radioSourceRecipeFor(key, p.language).initial
	p.values[key] = &value
	return &value
}

func (p *radioCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *radioCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-radio-config")
			h.Close("source-radio-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Radio Group compositions"}, display.CardSlots{
		Title: "Radio Group compositions", Description: "Source spacing, plan cards, native groups and independent selections.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-radio-config", &p.recipe, control.SelectProps{Label: "Radio Group configuration", Options: radioCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-radio-language", &p.language, control.SelectProps{Label: "Radio Group language", Options: statusLanguages()})
			}
			ui.Column(child.Key(p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center).Children(func() {
				if p.recipe == "rtl" {
					ui.Column(child).FillWidth().MinWidth(0).AlignItems(ui.Center).Direction(childDirection(p.language)).Children(func() { p.draw(child) })
				} else {
					p.draw(child)
				}
			})
			if p.popupError != "" {
				ui.Text(child, p.popupError).FontSize(13).TextColor(theme.TokensOf(child).Destructive).Role(ui.RoleAlert)
			}
		},
	})
}

func (p *radioCompositionPreviewState) draw(c *ui.Context) {
	r := radioSourceRecipeFor(p.recipe, p.language)
	value := p.selected(p.recipe)
	before := *value
	group := func(child *ui.Context) {
		g := ui.RadioGroup(child, func() {
			for _, option := range r.options {
				p.option(child.Key(option.value), value, option, r)
			}
		}).Gap(12).FillWidth().MinWidth(0).Role(ui.RoleRadioGroup).Label("Source radio group").FocusGroup(ui.Horizontal | ui.Vertical).Disabled(r.disabled).DimDisabled(false).Invalid(r.invalid)
		if r.grid {
			g.Grid().Columns(2).Gap(8)
		}
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(r.width).Gap(12).Children(func() {
		if r.legend != "" {
			control.FieldSetWith(c, control.FieldSetProps{Legend: r.legend, Description: r.description}, control.FieldSetSlots{
				Legend: func(child *ui.Context) ui.Element {
					variant := control.FieldLegendDefault
					if r.legendLabel {
						variant = control.FieldLegendLabel
					}
					return control.FieldLegend(child, r.legend, variant)
				},
				Content: group,
			})
		} else {
			group(c)
		}
		label := ""
		for _, option := range r.options {
			if option.value == *value {
				label = option.label
				break
			}
		}
		if before != *value {
			p.action = "Local radio choice: " + label
			p.changes[p.recipe]++
		}
		ui.Text(c, "Selected value: "+label).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground).Role(ui.RoleStatus).Label("Selected value: " + label)
	})
}

func (p *radioCompositionPreviewState) option(c *ui.Context, selected *string, option radioSourceOption, r radioSourceRecipe) {
	disabled := r.disabled || option.disabled
	control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal, Disabled: disabled, Invalid: r.invalid}, func(f *control.FieldScope) {
		buildControl := func() {
			f.Control(c, func() ui.Element {
				return control.RadioWith(c, selected, option.value, "", control.RadioProps{Disabled: disabled, Invalid: r.invalid})
			})
		}
		if r.card {
			f.Label(c, control.FieldLabelProps{Label: option.label, Choice: true, Selected: *selected == option.value}, func() {
				ui.Row(c).FillWidth().MinWidth(0).Gap(12).AlignItems(ui.Center).Children(func() {
					caption := func() {
						f.Content(c, func() { control.FieldTitle(c, option.label); f.Description(c, option.description) })
					}
					if r.grid {
						buildControl()
						caption()
					} else {
						caption()
						buildControl()
					}
				})
			})
			return
		}
		buildControl()
		caption := func() {
			label := f.Label(c, control.FieldLabelProps{Text: option.label, Disabled: disabled}, nil)
			if !r.medium {
				label.FontWeight(400)
			}
			f.Description(c, option.description)
		}
		if option.description != "" {
			f.Content(c, caption)
		} else {
			caption()
		}
	})
}

func radioSourceWords(language string) map[string]string {
	switch language {
	case "ar":
		return map[string]string{"default": "افتراضي", "defaultDescription": "تباعد قياسي لمعظم حالات الاستخدام.", "comfortable": "مريح", "comfortableDescription": "مساحة أكبر بين العناصر.", "compact": "مضغوط", "compactDescription": "تباعد أدنى للتخطيطات الكثيفة."}
	case "he":
		return map[string]string{"default": "ברירת מחדל", "defaultDescription": "ריווח סטנדרטי לרוב מקרי השימוש.", "comfortable": "נוח", "comfortableDescription": "יותר מקום בין האלמנטים.", "compact": "קומפקטי", "compactDescription": "ריווח מינימלי לפריסות צפופות."}
	default:
		return map[string]string{"default": "Default", "defaultDescription": "Standard spacing for most use cases.", "comfortable": "Comfortable", "comfortableDescription": "More space between elements.", "compact": "Compact", "compactDescription": "Minimal spacing for dense layouts."}
	}
}
