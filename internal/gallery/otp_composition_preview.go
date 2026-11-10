package gallery

import (
	"regexp"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Only application values survive a view. The current native editor and Field
// scope are rebuilt in place; no Context, Element or callback is retained here.
type otpCompositionPreviewState struct {
	ready, native                        bool
	recipe, language, action, popupError string
	drafts                               map[string]*string
	changes                              map[string]int
	selections                           map[string][2]int
	submitted                            map[string]string
	actions                              map[string]int
}

var otpCompositionRecipes = []control.Option{
	{Value: "demo", Label: "Six populated slots"},
	{Value: "pattern", Label: "Digits Only"},
	{Value: "separator", Label: "Three pairs"},
	{Value: "disabled", Label: "Disabled populated code"},
	{Value: "controlled", Label: "Controlled value"},
	{Value: "invalid", Label: "Invalid three pairs"},
	{Value: "four", Label: "Four digits"},
	{Value: "alphanumeric", Label: "Alphanumeric code"},
	{Value: "form", Label: "Verify your login (docs)"},
	{Value: "rtl", Label: "Translated verification code"},
	{Value: "create-form", Label: "Create: Verify your login"},
	{Value: "create-simple", Label: "Create: Simple"},
	{Value: "create-pattern", Label: "Create: Digits Only"},
	{Value: "create-separator", Label: "Create: With Separator"},
	{Value: "create-alphanumeric", Label: "Create: Alphanumeric"},
	{Value: "create-disabled", Label: "Create: Disabled"},
	{Value: "create-four", Label: "Create: 4 Digits"},
	{Value: "create-invalid", Label: "Create: Invalid State"},
}

// These are the anchored ASCII patterns exported by input-otp; the native
// legacy Alphanumeric preset also accepts Unicode letters and is not used.
var otpSourceDigits = regexp.MustCompile(`^[0-9]+$`)
var otpSourceAlphanumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func (p *otpCompositionPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready, p.native, p.recipe, p.language = true, native, "demo", "ar"
	p.drafts, p.changes, p.selections = map[string]*string{}, map[string]int{}, map[string][2]int{}
	p.submitted, p.actions = map[string]string{}, map[string]int{}
}

func (p *otpCompositionPreviewState) draft() *string {
	if value := p.drafts[p.recipe]; value != nil {
		return value
	}
	value := ""
	switch p.recipe {
	case "demo", "disabled", "rtl", "create-separator", "create-disabled":
		value = "123456"
	case "invalid", "create-invalid":
		value = "000000"
	}
	p.drafts[p.recipe] = &value
	return &value
}

func (p *otpCompositionPreviewState) choose(c *ui.Context, h *desktop.Host, native bool, key string, value *string, props control.SelectProps) {
	if native {
		control.NativeSelect(c, value, props)
	} else if _, err := desktop.Select(c, h, key, value, props); err != nil {
		p.popupError = err.Error()
	}
}

func (p *otpCompositionPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		if h != nil {
			h.Close("source-otp-config")
			h.Close("source-otp-language")
		}
		p.native, p.popupError = native, ""
	}
	display.CardWith(c, display.CardProps{Size: display.CardSmall, Label: "Input OTP compositions"}, display.CardSlots{
		Title: "Input OTP compositions", Description: "Source code slots, native selection, controlled values and local verification drafts.",
		Content: func(child *ui.Context) {
			p.choose(child, h, native, "source-otp-config", &p.recipe, control.SelectProps{Label: "Input OTP configuration", Options: otpCompositionRecipes})
			if p.recipe == "rtl" {
				p.choose(child, h, native, "source-otp-language", &p.language, control.SelectProps{Label: "Input OTP language", Options: statusLanguages()})
			}
			preview := ui.Column(child.Key("otp-source:" + p.recipe)).FillWidth().MinWidth(0).AlignItems(ui.Center)
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

func otpSourceLabel(language string) string {
	switch language {
	case "ar":
		return "رمز التحقق"
	case "he":
		return "קוד אימות"
	default:
		return "Verification code"
	}
}

func (p *otpCompositionPreviewState) editor(c *ui.Context, label string, props control.InputOTPProps) ui.Element {
	props.Label = label
	props.OnChange = func(string) { p.changes[p.recipe]++ }
	e := control.InputOTP(c, p.draft(), props)
	a, b := e.TextSelection()
	p.selections[p.recipe] = [2]int{a, b}
	return e
}

func (p *otpCompositionPreviewState) draw(c *ui.Context) {
	if p.recipe == "form" || p.recipe == "create-form" {
		p.form(c)
		return
	}
	props := control.InputOTPProps{Length: 6}
	label, description, field := "One-time code", "", false
	switch p.recipe {
	case "pattern", "create-pattern":
		label, field, props.Pattern = "Digits Only", true, otpSourceDigits
	case "separator", "invalid", "create-separator", "create-invalid":
		props.Groups = []int{2, 2, 2}
		if p.recipe == "invalid" || p.recipe == "create-invalid" {
			props.Invalid = true
		}
		if p.recipe == "create-separator" {
			label, field = "With Separator", true
		} else if p.recipe == "create-invalid" {
			label, description, field = "Invalid State", "Example showing the invalid error state.", true
		}
	case "alphanumeric", "create-alphanumeric":
		props.Groups, props.Pattern = []int{3, 3}, otpSourceAlphanumeric
		if p.recipe == "create-alphanumeric" {
			label, description, field = "Alphanumeric", "Accepts both letters and numbers.", true
		}
	case "disabled", "create-disabled":
		props.Groups, props.Disabled = []int{3, 3}, true
		if p.recipe == "create-disabled" {
			label, field = "Disabled", true
		}
	case "four", "create-four":
		props.Length, props.Pattern = 4, otpSourceDigits
		if p.recipe == "create-four" {
			label, description, field = "4 Digits", "Common pattern for PIN codes.", true
		}
	case "rtl":
		label, field = otpSourceLabel(p.language), true
	case "create-simple":
		label, field, props.Groups = "Simple", true, []int{3, 3}
	}
	ui.Column(c).FillWidth().MinWidth(0).MaxWidth(320).Gap(8).AlignItems(ui.Center).Children(func() {
		if field {
			fieldProps := control.FieldProps{}
			if p.recipe == "create-invalid" {
				fieldProps.Error = "Invalid code. Please try again."
			}
			fieldRoot := control.FieldWith(c, fieldProps, func(f *control.FieldScope) {
				// Source invalid state belongs to slots/error, not the Field label.
				f.Label(c, control.FieldLabelProps{Text: label}, nil).TextColor(theme.TokensOf(c).Foreground)
				if description != "" {
					f.Description(c, description)
				}
				f.Control(c, func() ui.Element { return p.editor(c, label, props) })
			})
			if p.recipe == "pattern" {
				fieldRoot.MaxWidth(216)
			}
		} else {
			p.editor(c, label, props)
		}
		if p.recipe == "controlled" {
			text := "Enter your one-time password."
			if value := *p.draft(); value != "" {
				text = "You entered: " + value
			}
			ui.Text(c, text).FontSize(14).TextAlign(ui.Center)
		}
	})
}

func (p *otpCompositionPreviewState) intent(action string) {
	p.actions[action]++
	p.action = "Local OTP action: " + action
}

func otpSourceLink(c *ui.Context, text string) ui.Element {
	link := ui.Link(c, text, "#").Underline().UnderlineOffset(4).TextColor(theme.TokensOf(c).MutedForeground)
	if link.Hovered() {
		link.TextColor(theme.TokensOf(c).Primary)
	}
	return link
}

func (p *otpCompositionPreviewState) form(c *ui.Context) {
	owner := ui.Box(c).FillWidth().MinWidth(0).MaxWidth(448)
	router := ui.Local(owner, "otp-hash-router", func() *ui.Router { return ui.NewRouter("/otp-examples") })
	owner.Children(func() {
		(*router).View(c, func(_ *ui.Route) {
			display.CardWith(c, display.CardProps{}, display.CardSlots{
				Title: "Verify your login",
				Header: func(child *ui.Context) {
					display.CardTitle(child, "Verify your login")
					ui.RichText(child).FillWidth().MinWidth(0).FontSize(14).TextColor(theme.TokensOf(child).MutedForeground).Children(func() {
						ui.Text(child, "Enter the verification code we sent to your email address: ")
						ui.Text(child, "m@example.com").FontWeight(500)
						ui.Text(child, ".")
					})
				},
				Content: func(child *ui.Context) {
					control.FieldWith(child, control.FieldProps{Required: true}, func(f *control.FieldScope) {
						ui.Row(child).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Center).Children(func() {
							f.Label(child, control.FieldLabelProps{Label: "Verification code"}, func() { ui.Text(child, "Verification code") }).Grow(1).MinWidth(0)
							if control.Button(child, "Resend Code", control.ButtonProps{Variant: control.Outline, Size: control.ExtraSmall, Icon: outlined.Refresh()}).Shrink(0).Clicked() {
								p.intent("Resend Code")
							}
						})
						f.Control(child, func() ui.Element {
							var input ui.Element
							ui.ScrollHorizontal(child).FillWidth().MinWidth(0).Role(ui.RoleRegion).Label("Verification code slots").Children(func() {
								gap := float32(8)
								props := control.InputOTPProps{Length: 6, Groups: []int{3, 3}, SlotWidth: 48, SlotHeight: 64, Gap: &gap, SlotContent: func(current *ui.Context, slot control.InputOTPSlotProps) {
									control.InputOTPSlot(current, slot).FontSize(20)
								}}
								if p.recipe == "form" {
									props.SlotWidth, props.SlotHeight, gap = 44, 48, 16
								}
								input = p.editor(child, "Verification code", props)
							})
							return input
						})
						f.DescriptionWith(child, "I no longer have access to this email address.", func() {
							if otpSourceLink(child, "I no longer have access to this email address.").Clicked() {
								p.intent("I no longer have access to this email address.")
							}
						})
					})
				},
				Footer: func(child *ui.Context) {
					gap := float32(8)
					if p.recipe == "form" {
						gap = 12
					}
					ui.Column(child).FillWidth().MinWidth(0).Gap(gap).Children(func() {
						if control.Button(child, "Verify", control.ButtonProps{}).FillWidth().Clicked() {
							p.submitted[p.recipe] = *p.draft()
							p.intent("Verify")
						}
						ui.RichText(child).FillWidth().MinWidth(0).FontSize(14).TextColor(theme.TokensOf(child).MutedForeground).Children(func() {
							ui.Text(child, "Having trouble signing in? ")
							if otpSourceLink(child, "Contact support").Clicked() {
								p.intent("Contact support")
							}
						})
					})
				},
			})
		}).FillWidth().MinWidth(0)
	})
}
