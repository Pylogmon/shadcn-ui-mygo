package gallery

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
)

func (s *State) otpPreview(c *ui.Context) {
	control.Field(c, control.FieldProps{Label: tr(c, "Verification code"), Description: tr(c, "Enter the six-digit code from your authenticator.")}, func() ui.Element {
		return control.InputOTP(c, &s.OTP, control.InputOTPProps{Length: 6, Groups: []int{3, 3}, DigitsOnly: true,
			OnChange: func(string) { s.formStatus = "" }, OnComplete: func(string) { s.formStatus = "Verification code complete" }})
	})
	if s.formStatus != "" {
		ui.Text(c, s.formStatus).Role(ui.RoleStatus).FontSize(14)
	}
}

func (s *State) choiceVariantsPreview(c *ui.Context) {
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Workspace access", "Select all, or choose individual permissions.")
			state := ui.CheckMixed
			if s.readAccess && s.writeAccess {
				state = ui.CheckOn
			} else if !s.readAccess && !s.writeAccess {
				state = ui.CheckOff
			}
			control.CheckboxState(c, &state, tr(c, "All permissions"), control.CheckboxProps{Description: tr(c, "Manage reading and editing together."), OnStateChange: func(value ui.CheckState) {
				s.readAccess, s.writeAccess = value == ui.CheckOn, value == ui.CheckOn
			}})
			ui.Column(c).PaddingStart(24).Gap(12).Children(func() {
				control.Checkbox(c, &s.readAccess, tr(c, "Read projects"), control.CheckboxProps{})
				control.Checkbox(c, &s.writeAccess, tr(c, "Edit projects"), control.CheckboxProps{})
			})
			display.Separator(c, false)
			control.Switch(c, &s.quietNotifications, control.SwitchProps{Size: control.Small, Label: tr(c, "Quiet notifications"), Description: tr(c, "Send a daily summary instead of individual updates.")})
		})
	}, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Compose your message", "Choose several formatting options with the keyboard.")
			joined := float32(0)
			control.ToggleGroup(c, &s.formatting, control.ToggleGroupProps{Label: tr(c, "Text formatting"), Multiple: true, Outline: true, Spacing: &joined,
				Items: localizedOptions(c, []control.Option{{Value: "bold", Label: "Bold"}, {Value: "italic", Label: "Italic"}, {Value: "underline", Label: "Underline"}}),
				ItemContent: func(child *ui.Context, option control.Option, _ bool) {
					text := ui.Text(child, tr(child, option.Label)).FontSize(14)
					if option.Value == "bold" {
						text.FontWeight(700)
					} else if option.Value == "italic" {
						text.Italic()
					} else {
						text.Underline()
					}
				}})
			control.MultiSelect(c, &s.formatting, control.SelectProps{
				Label: tr(c, "Selected formatting"), Placeholder: tr(c, "Choose formatting"),
				Options: localizedOptions(c, []control.Option{{Value: "bold", Label: "Bold"}, {Value: "italic", Label: "Italic"}, {Value: "underline", Label: "Underline"}}),
				ValuesContent: func(child *ui.Context, view control.SelectValueView) {
					label := fmt.Sprintf("%d styles selected", len(view.Values))
					if len(view.Values) == 1 {
						label = "1 style selected"
					}
					ui.Text(child, tr(child, label)).SingleLine()
				},
			}).MaxWidth(260)
			control.Field(c, control.FieldProps{Label: tr(c, "Invite code"), Description: tr(c, "Letters, numbers and symbols are welcome. Pasted codes are capitalized.")}, func() ui.Element {
				return control.InputOTP(c, &s.inviteCode, control.InputOTPProps{Length: 6, Groups: []int{2, 4}, PasteTransform: strings.ToUpper})
			})
		})
	})
}
