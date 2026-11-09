package conversation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type Question struct {
	ID, Title, Description       string
	Choices                      []control.Option
	Required, Multiple, Freeform bool
	InputLabel, Placeholder      string
}
type Answer struct {
	Values  []string
	Text    string
	Skipped bool
}
type QuestionnaireState struct {
	Index     int
	Answers   map[string]Answer
	Errors    map[string]string
	Submitted bool
}
type QuestionnaireProps struct {
	Questions []Question
	Disabled  bool
	OnSubmit  func(map[string]Answer)
}

func questionID(q Question, index int) string {
	if q.ID != "" {
		return q.ID
	}
	return fmt.Sprintf("question-%d", index)
}
func (s *QuestionnaireState) init() {
	if s.Answers == nil {
		s.Answers = make(map[string]Answer)
	}
	if s.Errors == nil {
		s.Errors = make(map[string]string)
	}
}

func validValues(q Question, a Answer) []string {
	var values []string
	for _, v := range a.Values {
		for _, choice := range q.Choices {
			if !choice.Disabled && choice.Value == v && !slices.Contains(values, v) {
				values = append(values, v)
				break
			}
		}
	}
	if !q.Multiple && len(values) > 1 {
		return values[:1]
	}
	return values
}
func answerError(q Question, a Answer) string {
	if !q.Required || len(validValues(q, a)) > 0 || q.Freeform && strings.TrimSpace(a.Text) != "" {
		return ""
	}
	return "Please answer this question."
}

// Select toggles a valid choice, preserving a multi-selection's order.
// It ignores disabled/unknown values and invalidates a prior submission.
func (s *QuestionnaireState) Select(q Question, index int, value string) {
	allowed := false
	for _, o := range q.Choices {
		if o.Value == value && !o.Disabled {
			allowed = true
			break
		}
	}
	if !allowed {
		return
	}
	s.init()
	id := questionID(q, index)
	a := s.Answers[id]
	a.Skipped = false
	if q.Multiple {
		if i := slices.Index(a.Values, value); i >= 0 {
			a.Values = slices.Delete(a.Values, i, i+1)
		} else {
			a.Values = append(a.Values, value)
		}
	} else {
		a.Values = []string{value}
	}
	s.Answers[id] = a
	delete(s.Errors, id)
	s.Submitted = false
}

// Next validates the active question. On the last question it validates all
// questions, and only then marks the complete response submitted.
func (s *QuestionnaireState) Next(questions []Question) bool {
	s.init()
	if len(questions) == 0 {
		return false
	}
	s.Index = max(0, min(s.Index, len(questions)-1))
	q := questions[s.Index]
	id := questionID(q, s.Index)
	if err := answerError(q, s.Answers[id]); err != "" {
		s.Errors[id] = err
		return false
	}
	delete(s.Errors, id)
	if s.Index < len(questions)-1 {
		s.Index++
		return true
	}
	return s.Submit(questions)
}
func (s *QuestionnaireState) Previous() {
	s.Index = max(0, s.Index-1)
}
func (s *QuestionnaireState) Skip(questions []Question) bool {
	if len(questions) == 0 {
		return false
	}
	s.Index = max(0, min(s.Index, len(questions)-1))
	q := questions[s.Index]
	if q.Required {
		return false
	}
	s.init()
	id := questionID(q, s.Index)
	s.Answers[id] = Answer{Skipped: true}
	delete(s.Errors, id)
	if s.Index < len(questions)-1 {
		s.Index++
		return true
	}
	return s.Submit(questions)
}
func (s *QuestionnaireState) Submit(questions []Question) bool {
	s.init()
	if len(questions) == 0 {
		return false
	}
	firstInvalid := -1
	for i, q := range questions {
		id := questionID(q, i)
		a := s.Answers[id]
		if err := answerError(q, a); err != "" {
			s.Errors[id] = err
			if firstInvalid < 0 {
				firstInvalid = i
			}
		} else {
			delete(s.Errors, id)
		}
	}
	if firstInvalid >= 0 {
		s.Index = firstInvalid
		s.Submitted = false
		return false
	}
	s.Submitted = true
	return true
}

// Result copies the ordered question set into a stable answer map. Unknown,
// disabled and duplicate choices are excluded; freeform text is only returned
// by freeform questions. The caller can safely retain or mutate the result.
func (s *QuestionnaireState) Result(questions []Question) map[string]Answer {
	result := make(map[string]Answer, len(questions))
	for i, q := range questions {
		id := questionID(q, i)
		a := s.Answers[id]
		a.Values = validValues(q, a)
		if !q.Freeform {
			a.Text = ""
		}
		result[id] = a
	}
	return result
}

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
