package conversation

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Questionnaire preserves the legacy convenience wizard, including required-only
// validation and its once-submitted state. Use QuestionnaireWith for source slots
// and the strict answered-or-explicitly-skipped navigation contract.
func Questionnaire(c *ui.Context, state *QuestionnaireState, p QuestionnaireProps) ui.Element {
	if state == nil {
		panic("conversation: Questionnaire requires caller-owned state")
	}
	state.init()
	tk := theme.TokensOf(c)
	return ui.Column(c).Padding(20).Gap(14).Radius(tk.Radius).Border(1, tk.Border).Background(tk.Card).Role(ui.RoleGroup).Label("Questionnaire").Children(func() {
		if len(p.Questions) == 0 {
			ui.Text(c, "No questions").TextColor(tk.MutedForeground)
			return
		}
		state.Index = max(0, min(state.Index, len(p.Questions)-1))
		q := p.Questions[state.Index]
		id := questionID(q, state.Index)
		ui.Text(c, fmt.Sprintf("Question %d of %d", state.Index+1, len(p.Questions))).FontSize(12).TextColor(tk.MutedForeground).Role(ui.RoleStatus)
		ui.Text(c, q.Title).FontSize(18).FontWeight(600)
		if q.Description != "" {
			ui.Text(c, q.Description).TextColor(tk.MutedForeground)
		}
		ui.Column(c.Key(id)).Gap(10).Children(func() {
			a := state.Answers[id]
			if q.Multiple {
				for _, option := range q.Choices {
					selected := slices.Contains(a.Values, option.Value)
					if control.Checkbox(c.Key(option.Value), &selected, option.Label, control.CheckboxProps{Disabled: p.Disabled || option.Disabled}).Changed() {
						state.Select(q, state.Index, option.Value)
						a = state.Answers[id]
					}
				}
			} else if len(q.Choices) > 0 {
				selected := ""
				if len(a.Values) > 0 {
					selected = a.Values[0]
				}
				before := selected
				control.RadioGroup(c, &selected, control.RadioGroupProps{Label: q.Title, Options: q.Choices, Disabled: p.Disabled})
				if selected != before {
					state.Select(q, state.Index, selected)
					a = state.Answers[id]
				}
			}
			if q.Freeform {
				label := q.InputLabel
				if label == "" {
					label = "Your answer"
				}
				text := a.Text
				if control.Input(c, &text, control.InputProps{Label: label, Placeholder: q.Placeholder, Disabled: p.Disabled, Invalid: state.Errors[id] != ""}).Changed() {
					a.Text = text
					a.Skipped = false
					state.Answers[id] = a
					delete(state.Errors, id)
					state.Submitted = false
				}
			}
			if err := state.Errors[id]; err != "" {
				ui.Text(c, err).TextColor(tk.Destructive).FontSize(12).Role(ui.RoleAlert)
			}
		})
		ui.Row(c).Gap(8).Children(func() {
			if control.Button(c, "Previous", control.ButtonProps{Variant: control.Outline, Disabled: p.Disabled || state.Index == 0}).Clicked() {
				state.Previous()
			}
			if !q.Required && control.Button(c, "Skip", control.ButtonProps{Variant: control.Ghost, Disabled: p.Disabled}).Clicked() {
				was := state.Submitted
				if state.Skip(p.Questions) && state.Submitted && !was && p.OnSubmit != nil {
					p.OnSubmit(state.Result(p.Questions))
				}
			}
			label := "Next"
			if state.Index == len(p.Questions)-1 {
				label = "Submit"
			}
			if control.Button(c, label, control.ButtonProps{Disabled: p.Disabled || state.Submitted}).Clicked() {
				was := state.Submitted
				if state.Next(p.Questions) && state.Submitted && !was && p.OnSubmit != nil {
					p.OnSubmit(state.Result(p.Questions))
				}
			}
		})
		if state.Submitted {
			ui.Text(c, "Response submitted").Role(ui.RoleStatus).TextColor(tk.MutedForeground)
		}
	})
}
