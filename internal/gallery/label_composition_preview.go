package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

// Application-owned data only; labels link actual native targets during the
// current build and no Element or FieldScope survives that build.
type labelCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	texts                                map[string]*string
	checks                               map[string]*bool
	changes                              map[string]int
	selections                           map[string][2]int
	paymentSaved                         map[string]string
	paymentSubmits                       int
}

var labelCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Checkbox label (docs)"},
	{Value: "rtl", Label: "Translated checkbox label"},
	{Value: "usage", Label: "Guide: Standalone label"},
	{Value: "field-usage", Label: "Guide: Label in Field"},
	{Value: "field-example", Label: "Related Field payment example"},
	{Value: "create-checkbox", Label: "Create: With Checkbox"},
	{Value: "create-input", Label: "Create: With Input"},
	{Value: "create-disabled", Label: "Create: Disabled"},
	{Value: "create-textarea", Label: "Create: With Textarea"},
}

func (p *labelCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.texts, p.checks, p.changes, p.selections = map[string]*string{}, map[string]*bool{}, map[string]int{}, map[string][2]int{}
}
func (p *labelCompositionPreviewState) text(key string) *string {
	key = p.recipe + ":" + key
	if value := p.texts[key]; value != nil {
		return value
	}
	value := ""
	p.texts[key] = &value
	return &value
}
func (p *labelCompositionPreviewState) checked(key string, initial bool) *bool {
	key = p.recipe + ":" + key
	if value := p.checks[key]; value != nil {
		return value
	}
	value := initial
	p.checks[key] = &value
	return &value
}
func (p *labelCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) ui.Element {
	if native {
		return control.NativeSelect(c, value, props).MinWidth(0)
	}
	e, err := desktop.Select(c, h, key, value, props)
	if err != nil {
		p.popupError = err.Error()
	}
	return e.MinWidth(0)
}
func (p *labelCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			for _, key := range []string{"source-label-config", "source-label-language", "source-label-month", "source-label-year"} {
				h.Close(key)
			}
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: tr(c, "Label compositions")}, display.CardSlots{
		Title: tr(c, "Label compositions"), Description: tr(c, "Source standalone labels and actual native checkbox, input and multiline targets."),
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-label-config", &p.recipe, control.SelectProps{Label: tr(child, "Label configuration"), Options: localizedOptions(child, labelCompositionRecipes)})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-label-language", &p.language, control.SelectProps{Label: tr(child, "Label language"), Options: localizedOptions(child, statusLanguages())})
			}
			preview := ui.Column(child.Key("label-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
			if p.recipe == "rtl" {
				preview.Direction(childDirection(p.language))
			}
			preview.Children(func() { p.draw(child, h, native) })
			if p.action != "" {
				ui.Text(child, tr(child, p.action)).FontSize(13).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(child, p.popupError).Role(ui.RoleAlert)
			}
		},
	})
}

func labelSourceTerms(language string) string {
	switch language {
	case "ar":
		return "قبول الشروط والأحكام"
	case "he":
		return "קבל תנאים והגבלות"
	default:
		return "Accept terms and conditions"
	}
}

func (p *labelCompositionPreviewState) checkbox(c *ui.Context, label string, field bool) {
	build := func() {
		e := control.Checkbox(c, p.checked("terms", false), "", control.CheckboxProps{OnCheckedChange: func(bool) { p.changes[p.recipe]++ }})
		control.LabelFor(c, label, e)
	}
	if field {
		control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal}, func(_ *control.FieldScope) { build() })
	} else {
		ui.Row(c).Gap(8).AlignSelf(ui.Center).Children(build)
	}
}

// A label declared before its target keeps source visual order; the actual
// native editor links back to that same label once constructed.
func (p *labelCompositionPreviewState) plainField(c *ui.Context, label, placeholder string, multiline, disabled bool) {
	control.FieldWith(c, control.FieldProps{}, func(f *control.FieldScope) {
		caption := control.Label(c, tr(c, label))
		e := f.Control(c, func() ui.Element {
			changed := func(string) { p.changes[p.recipe]++ }
			if multiline {
				return control.Textarea(c, p.text("value"), control.TextareaProps{Label: tr(c, label), Placeholder: tr(c, placeholder), Disabled: disabled, OnChange: changed})
			}
			return control.Input(c, p.text("value"), control.InputProps{Label: tr(c, label), Placeholder: tr(c, placeholder), Disabled: disabled, OnChange: changed})
		})
		e.Label(tr(c, label)).LabelledBy(caption)
		if disabled {
			caption.Opacity(.5).Cursor(ui.CursorNotAllowed)
		}
		a, b := e.TextSelection()
		p.selections[p.recipe] = [2]int{a, b}
	})
}

func (p *labelCompositionPreviewState) draw(c *ui.Context, h *desktop.Host, native bool) {
	width := float32(384)
	if p.recipe == "field-example" {
		width = 448
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(width).Gap(28).Children(func() {
		switch p.recipe {
		case "demo":
			p.checkbox(c, "Accept terms and conditions", false)
		case "rtl":
			p.checkbox(c, labelSourceTerms(p.language), false)
		case "usage":
			control.Label(c, tr(c, "Your email address"))
		case "field-usage":
			control.FieldWith(c, control.FieldProps{}, func(f *control.FieldScope) {
				f.Label(c, control.FieldLabelProps{Text: tr(c, "Your email address")}, nil)
				f.Control(c, func() ui.Element {
					return control.Input(c, p.text("value"), control.InputProps{Label: tr(c, "Your email address"), OnChange: func(string) { p.changes[p.recipe]++ }})
				})
			})
		case "create-checkbox":
			p.checkbox(c, "Accept terms and conditions", true)
		case "create-input":
			p.plainField(c, "Username", "Username", false, false)
		case "create-disabled":
			p.plainField(c, "Disabled", "Disabled", false, true)
		case "create-textarea":
			p.plainField(c, "Message", "Message", true, false)
		case "field-example":
			p.payment(c, h, native)
		}
	})
}

func (p *labelCompositionPreviewState) paymentField(c *ui.Context, key, label, placeholder, description string, required, multiline bool) {
	control.FieldWith(c.Key(key), control.FieldProps{Required: required}, func(f *control.FieldScope) {
		// Required belongs to the real editor; the source caption has no star.
		f.Label(c, control.FieldLabelProps{Label: tr(c, label)}, func() { ui.Text(c, tr(c, label)) })
		f.Control(c, func() ui.Element {
			if multiline {
				return control.Textarea(c, p.text(key), control.TextareaProps{Label: tr(c, label), Placeholder: tr(c, placeholder), OnChange: func(string) { p.changes[p.recipe]++ }})
			}
			return control.Input(c, p.text(key), control.InputProps{Label: tr(c, label), Placeholder: tr(c, placeholder), OnChange: func(string) { p.changes[p.recipe]++ }})
		})
		if description != "" {
			f.Description(c, tr(c, description))
		}
	})
}

func (p *labelCompositionPreviewState) payment(c *ui.Context, h *desktop.Host, native bool) {
	control.FieldGroup(c, func() {
		control.FieldSetWith(c, control.FieldSetProps{Legend: "Payment Method", Description: tr(c, "All transactions are secure and encrypted")}, control.FieldSetSlots{Content: func(child *ui.Context) {
			control.FieldGroup(child, func() {
				p.paymentField(child, "name", "Name on Card", "Evil Rabbit", "", true, false)
				p.paymentField(child, "number", "Card Number", "1234 5678 9012 3456", "Enter your 16-digit card number", true, false)
				ui.Column(child).Grid().Columns(3).Gap(16).FillWidth().MinWidth(0).Children(func() {
					for _, kind := range []string{"month", "year"} {
						label, placeholder := "Month", "MM"
						options := []control.Option{{Value: "", Label: "MM"}}
						if kind == "year" {
							label, placeholder = "Year", "YYYY"
							options = []control.Option{{Value: "", Label: "YYYY"}}
							for year := 2024; year <= 2029; year++ {
								value := fmt.Sprint(year)
								options = append(options, control.Option{Value: value, Label: tr(child, value)})
							}
						} else {
							for month := 1; month <= 12; month++ {
								value := fmt.Sprintf("%02d", month)
								options = append(options, control.Option{Value: value, Label: tr(child, value)})
							}
						}
						control.FieldWith(child.Key(kind), control.FieldProps{}, func(f *control.FieldScope) {
							f.Label(child, control.FieldLabelProps{Text: tr(child, label)}, nil)
							f.Control(child, func() ui.Element {
								return p.choose(child, h, native, "source-label-"+kind, p.text(kind), control.SelectProps{Label: tr(child, label), Placeholder: tr(child, placeholder), Options: localizedOptions(child, options)}).FillWidth()
							})
						})
					}
					p.paymentField(child, "cvv", "CVV", "123", "", true, false)
				})
			})
		}})
		control.FieldSeparator(c, "")
		control.FieldSetWith(c, control.FieldSetProps{Legend: "Billing Address", Description: tr(c, "The billing address associated with your payment method")}, control.FieldSetSlots{Content: func(child *ui.Context) {
			control.FieldGroup(child, func() {
				control.FieldWith(child, control.FieldProps{Orientation: control.FieldHorizontal}, func(f *control.FieldScope) {
					f.Control(child, func() ui.Element {
						return control.Checkbox(child, p.checked("shipping", true), "", control.CheckboxProps{OnCheckedChange: func(bool) { p.changes[p.recipe]++ }})
					})
					f.Label(child, control.FieldLabelProps{Text: tr(child, "Same as shipping address")}, nil).FontWeight(400)
				})
			})
		}})
		control.FieldSetWith(c, control.FieldSetProps{}, control.FieldSetSlots{Content: func(child *ui.Context) {
			control.FieldGroup(child, func() { p.paymentField(child, "comments", "Comments", "Add any additional comments", "", false, true) })
		}})
		control.FieldWith(c, control.FieldProps{Orientation: control.FieldHorizontal}, func(_ *control.FieldScope) {
			if control.Button(c, tr(c, "Submit"), control.ButtonProps{}).Clicked() {
				if *p.text("name") == "" || *p.text("number") == "" || *p.text("cvv") == "" {
					p.action = "Local Label validation: Complete the required payment fields."
				} else {
					p.paymentSaved = map[string]string{}
					for _, key := range []string{"name", "number", "month", "year", "cvv", "comments"} {
						p.paymentSaved[key] = *p.text(key)
					}
					p.paymentSaved["shipping"] = fmt.Sprint(*p.checked("shipping", true))
					p.paymentSubmits++
					p.action = "Local Label action: Submit"
				}
			}
			if control.Button(c, tr(c, "Cancel"), control.ButtonProps{Variant: control.Outline}).Clicked() {
				p.action = "Local Label action: Cancel"
			}
		})
	})
}
