package conversation

import (
	"slices"
	"strings"
)

// A nil collection preserves the original Text/Mode model. Creating a second
// slot upgrades that item's value without changing the first editor's identity.
func questionnaireManageInputs(q Question, a *Answer) {
	if a.Inputs == nil {
		selected := questionnaireInputSelected(q, *a)
		a.Inputs = []QuestionnaireInputValue{{Text: a.Text, Selected: selected}}
	}
}

func questionnaireInputValue(q Question, a Answer, key string) (string, bool) {
	if a.Inputs == nil {
		if key == "" {
			return a.Text, questionnaireInputSelected(q, a)
		}
		return "", false
	}
	for _, input := range a.Inputs {
		if input.Key == key {
			text := input.Text
			if key == "" {
				text = a.Text
			}
			return text, input.Selected && !a.Skipped
		}
	}
	return "", false
}

func questionnaireSetInput(q Question, a *Answer, key, text string, selected bool) {
	if key == "" && a.Inputs == nil {
		a.Text = text
		if selected && !q.Multiple {
			a.Values = nil
		}
		questionnaireInputMode(a, selected)
		return
	}
	questionnaireManageInputs(q, a)
	if selected && !q.Multiple {
		a.Values = nil
		questionnaireDeselectInputs(a)
	}
	index := slices.IndexFunc(a.Inputs, func(input QuestionnaireInputValue) bool { return input.Key == key })
	if index < 0 {
		a.Inputs = append(a.Inputs, QuestionnaireInputValue{Key: key})
		index = len(a.Inputs) - 1
	}
	a.Inputs[index].Text, a.Inputs[index].Selected = text, selected
	if key == "" {
		a.Text = text
	}
	questionnaireInputMode(a, questionnaireAnyInputSelected(*a))
}

func questionnaireAnyInputSelected(a Answer) bool {
	for _, input := range a.Inputs {
		if input.Selected {
			return true
		}
	}
	return false
}

func questionnaireDeselectInputs(a *Answer) {
	for index := range a.Inputs {
		a.Inputs[index].Selected = false
	}
}

func questionnaireRemoveInput(a *Answer, key string) {
	if key == "" {
		a.Text = ""
	}
	if a.Inputs == nil {
		questionnaireInputMode(a, false)
		return
	}
	a.Inputs = slices.DeleteFunc(a.Inputs, func(input QuestionnaireInputValue) bool { return input.Key == key })
	questionnaireInputMode(a, questionnaireAnyInputSelected(*a))
}

func (s *QuestionnaireState) setInputText(q Question, index int, key, text string) {
	s.init()
	id := questionID(q, index)
	a := s.Answers[id]
	a.Skipped = false
	if a.Inputs == nil && key == "" && !q.Multiple {
		a.Values = nil
	}
	questionnaireSetInput(q, &a, key, text, strings.TrimSpace(text) != "")
	s.Answers[id] = a
	delete(s.Errors, id)
	s.Submitted = false
}

func (s *QuestionnaireState) successfulInputs(q Question, index int, a Answer) []QuestionnaireInputValue {
	inputs := make([]QuestionnaireInputValue, 0, len(a.Inputs))
	if !q.Freeform || q.Disabled || a.Skipped {
		return inputs
	}
	if r := s.runtime[questionID(q, index)]; r != nil {
		for _, control := range r.order {
			if control.input && !control.disabled {
				if text, selected := questionnaireInputValue(q, a, control.inputKey); selected {
					inputs = append(inputs, QuestionnaireInputValue{Key: control.inputKey, Text: text, Selected: true})
				}
			}
		}
		return inputs
	}
	for _, input := range a.Inputs {
		if input.Selected {
			if input.Key == "" {
				input.Text = a.Text
			}
			inputs = append(inputs, input)
		}
	}
	return inputs
}

func (s *QuestionnaireState) hasSuccessfulInput(q Question, index int, a Answer) bool {
	if a.Inputs != nil {
		return len(s.successfulInputs(q, index, a)) > 0
	}
	return q.Freeform && !s.inputDisabled(q, index) && questionnaireInputSelected(q, a)
}

func questionnaireAnswersEqual(a, b Answer) bool {
	return slices.Equal(a.Values, b.Values) && a.Text == b.Text && slices.Equal(a.Inputs, b.Inputs) && a.Skipped == b.Skipped && a.Mode == b.Mode
}
