package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

// Application drafts and selection values only. Field scopes and native editor
// Elements are constructed by the current view and are never retained here.
type textareaCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	drafts                               map[string]*string
	changes                              map[string]int
	selections                           map[string][2]int
	submitted                            map[string]string
	sends                                int
}

var textareaCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Basic (docs)"},
	{Value: "field", Label: "Field with preceding description"},
	{Value: "disabled", Label: "Disabled field (docs)"},
	{Value: "invalid", Label: "Invalid field (docs)"},
	{Value: "button", Label: "Send message"},
	{Value: "rtl", Label: "Translated feedback"},
	{Value: "create-basic", Label: "Create: Basic"},
	{Value: "create-invalid", Label: "Create: Invalid editor"},
	{Value: "create-label", Label: "Create: Label with six rows"},
	{Value: "create-description", Label: "Create: Six rows and following description"},
	{Value: "create-disabled", Label: "Create: Disabled editor"},
}

func (p *textareaCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.drafts, p.changes = map[string]*string{}, map[string]int{}
	p.selections, p.submitted = map[string][2]int{}, map[string]string{}
}

func (p *textareaCompositionPreviewState) draft(key string) *string {
	if value := p.drafts[key]; value != nil {
		return value
	}
	value := ""
	p.drafts[key] = &value
	return &value
}

func (p *textareaCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *textareaCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-textarea-config")
			h.Close("source-textarea-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Textarea compositions"}, display.CardSlots{
		Title: "Textarea compositions", Description: "Source multiline editors, field order, native selection and local message drafts.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-textarea-config", &p.recipe, control.SelectProps{Label: "Textarea configuration", Options: textareaCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-textarea-language", &p.language, control.SelectProps{Label: "Textarea language", Options: statusLanguages()})
			}
			preview := ui.Column(child.Key("textarea-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
			if p.recipe == "rtl" {
				preview.Direction(childDirection(p.language))
			}
			preview.Children(func() { p.draw(child) })
			if p.action != "" {
				ui.Text(child, p.action).FontSize(13).Role(ui.RoleStatus).Label(p.action)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert)
			}
		},
	})
}

func textareaSourceWords(language string) (label, placeholder, description string) {
	switch language {
	case "ar":
		return "التعليقات", "تعليقاتك تساعدنا على التحسين...", "شاركنا أفكارك حول خدمتنا."
	case "he":
		return "משוב", "המשוב שלך עוזר לנו להשתפר...", "שתף את מחשבותיך על השירות שלנו."
	default:
		return "Feedback", "Your feedback helps us improve...", "Share your thoughts about our service."
	}
}

func (p *textareaCompositionPreviewState) draw(c *ui.Context) {
	label, placeholder, description := "", "Type your message here.", ""
	rows, field, before := 0, false, false
	disabled, fieldDisabled, invalid := false, false, false
	width := float32(0)
	switch p.recipe {
	case "demo", "button":
		width = 320
	case "field":
		label, description, field, before, width = "Message", "Enter your message below.", true, true, 320
	case "disabled":
		label, field, disabled, fieldDisabled, width = "Message", true, true, true, 320
	case "invalid":
		label, description, field, invalid, width = "Message", "Please enter a valid message.", true, true, 320
	case "rtl":
		label, placeholder, description = textareaSourceWords(p.language)
		rows, field, width = 4, true, 320
	case "create-invalid":
		invalid = true
	case "create-label":
		label, rows, field = "Message", 6, true
	case "create-description":
		label, description, rows, field = "Message", "Type your message and press enter to send.", 6, true
	case "create-disabled":
		label, field, disabled = "Message", true, true
	}
	container := ui.Column(c).FillWidth().MinWidth(0).Gap(8)
	if width > 0 {
		container.MaxWidth(width)
	}
	container.Children(func() {
		name := label
		if name == "" {
			name = placeholder
		}
		editor := func() ui.Element {
			e := control.Textarea(c, p.draft(p.recipe), control.TextareaProps{Label: name, Placeholder: placeholder, Rows: rows, Disabled: disabled, Invalid: invalid, OnChange: func(string) {
				p.changes[p.recipe]++
			}})
			a, b := e.TextSelection()
			p.selections[p.recipe] = [2]int{a, b}
			return e
		}
		if field {
			control.FieldWith(c, control.FieldProps{Disabled: fieldDisabled, Invalid: invalid}, func(f *control.FieldScope) {
				f.Label(c, control.FieldLabelProps{Text: label}, nil)
				if before {
					f.Description(c, description)
				}
				f.Control(c, editor)
				if !before && description != "" {
					f.Description(c, description)
				}
			})
		} else {
			editor()
		}
		if p.recipe == "button" && control.Button(c, "Send message", control.ButtonProps{}).FillWidth().Clicked() {
			p.submitted[p.recipe] = *p.draft(p.recipe)
			p.sends++
			p.action = "Local textarea action: Send message"
		}
	})
}
