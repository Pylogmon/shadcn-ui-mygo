package conversation

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
)

type QuestionnaireShortcutMode string

const (
	QuestionnaireLetters QuestionnaireShortcutMode = "letters"
	QuestionnaireNumbers QuestionnaireShortcutMode = "numbers"
)

type QuestionnaireItemStatus string

const (
	QuestionnaireUnanswered QuestionnaireItemStatus = "unanswered"
	QuestionnaireAnswered   QuestionnaireItemStatus = "answered"
	QuestionnaireSkipped    QuestionnaireItemStatus = "skipped"
)

type QuestionnaireAnswerMode uint8

const (
	QuestionnaireAutomaticAnswer QuestionnaireAnswerMode = iota
	QuestionnaireChoiceAnswer
	QuestionnaireInputAnswer
	// Combined is a multiple item with both a fixed answer and selected input.
	QuestionnaireCombinedAnswer
	// Draft keeps native input text without making it a successful answer.
	QuestionnaireDraftAnswer
)

type QuestionnaireResetEvent struct{ prevented bool }

func (e *QuestionnaireResetEvent) PreventDefault() { e.prevented = true }

type QuestionnaireSnapshot struct {
	Current, Total int
	First, Last    bool
	Item           string
	Status         QuestionnaireItemStatus
}
type questionnaireActionKind uint8

const (
	qPrevious questionnaireActionKind = iota
	qNext
	qSkip
	qSubmit
	qReset
	qConfirm
	qShortcut
	qMoveFocus
)

type questionnaireAction struct {
	kind  questionnaireActionKind
	value string
}
type questionnaireAnswerControl struct {
	handle                                            ui.Handle
	value, text                                       string
	input, disabled, selected, composing, initialized bool
	controlled                                        bool
	controlledChecked                                 bool
	controlledText                                    string
	resetVersion                                      uint64
}
type questionnaireRadioValue struct {
	value    string
	selected bool
}
type questionnaireRuntime struct {
	item           ui.Handle
	controls       map[string]*questionnaireAnswerControl
	order          []*questionnaireAnswerControl
	radio          questionnaireRadioValue
	hadAnswer      bool
	defaults       Answer
	reported       Answer
	reportedStatus QuestionnaireItemStatus
	composing      bool
	attempted      bool
	multiple       bool
	crossError     string
	crossAnswer    Answer
}

func copyQuestionnaireAnswer(a Answer) Answer { a.Values = slices.Clone(a.Values); return a }
func (s *QuestionnaireState) effectiveQuestion(q Question, index int) Question {
	if r := s.runtime[questionID(q, index)]; r != nil {
		q.Choices = slices.Clone(q.Choices)
		for _, answer := range r.order {
			if answer.input {
				q.Freeform = true
				continue
			}
			found := false
			for _, choice := range q.Choices {
				if choice.Value == answer.value {
					found = true
					break
				}
			}
			if !found {
				q.Choices = append(q.Choices, control.Option{Value: answer.value, Disabled: answer.disabled})
			}
		}
	}
	return q
}

// Status reports successful mounted answers or an intentional skip.
func (s *QuestionnaireState) Status(q Question, index int) QuestionnaireItemStatus {
	q = s.effectiveQuestion(q, index)
	a := s.Answers[questionID(q, index)]
	if a.Skipped {
		return QuestionnaireSkipped
	}
	if len(s.successfulChoices(q, index, a)) > 0 || !s.inputDisabled(q, index) && q.Freeform && questionnaireInputSelected(q, a) {
		return QuestionnaireAnswered
	}
	return QuestionnaireUnanswered
}

// SetText selects an input answer (or deselects it when blank), preserving a
// multiple item's fixed choices. It is the model equivalent of an accepted edit.
func (s *QuestionnaireState) SetText(q Question, index int, text string) {
	s.init()
	id := questionID(q, index)
	a := s.Answers[id]
	a.Text = text
	a.Skipped = false
	if !q.Multiple {
		a.Values = nil
	}
	questionnaireInputMode(&a, strings.TrimSpace(text) != "")
	s.Answers[id] = a
	delete(s.Errors, id)
	s.Submitted = false
}

// JumpTo changes the local active item for resume or validator-directed editing.
// It does not validate the current item. UI focus follows on the next build.
// With controlled Props.Item, update that application value instead.
func (s *QuestionnaireState) JumpTo(questions []Question, id string) bool {
	for index, q := range questions {
		if questionID(q, index) == id && !q.Disabled {
			s.Index = index
			s.pendingName = id
			s.focusAttempts = 0
			s.pendingInvalid = false
			return true
		}
	}
	return false
}

// Reset restores the latest compound defaults and DefaultItem. Retained native
// controls update on the next build; controlled values remain authoritative.
func (s *QuestionnaireState) Reset(questions []Question) {
	s.init()
	for i, q := range questions {
		id := questionID(q, i)
		a := q.DefaultAnswer
		if r := s.runtime[id]; r != nil {
			a = r.defaults
		}
		s.Answers[id] = copyQuestionnaireAnswer(a)
	}
	clear(s.Errors)
	s.Submitted = false
	s.resetVersion++
	for _, r := range s.runtime {
		r.attempted = false
		r.crossError = ""
	}
	previous := s.activeName
	s.Index = 0
	if !s.JumpTo(questions, s.defaultItem) {
		for i, q := range questions {
			if !q.Disabled {
				s.Index = i
				s.pendingName = questionID(q, i)
				break
			}
		}
	}
	if s.pendingName == previous {
		s.pendingName = ""
		s.pendingInvalid = false
	}
}

// FormValues mirrors successful native form controls: skipped and disabled
// items are absent, and a single item's inactive freeform draft is excluded.
func (s *QuestionnaireState) FormValues(questions []Question) map[string][]string {
	values := map[string][]string{}
	result := s.Result(questions)
	for index, q := range questions {
		id := questionID(q, index)
		a, exists := result[id]
		if !exists || a.Skipped {
			continue
		}
		var ordered []string
		if r := s.runtime[id]; r != nil && len(r.order) > 0 {
			for _, control := range r.order {
				if control.input {
					if questionnaireInputSelected(s.effectiveQuestion(q, index), a) {
						ordered = append(ordered, a.Text)
					}
					continue
				}
				if slices.Contains(a.Values, control.value) && !slices.Contains(ordered, control.value) {
					ordered = append(ordered, control.value)
				}
			}
		} else {
			for _, choice := range q.Choices {
				if slices.Contains(a.Values, choice.Value) && !slices.Contains(ordered, choice.Value) {
					ordered = append(ordered, choice.Value)
				}
			}
		}
		if (s.runtime[id] == nil || len(s.runtime[id].order) == 0) && questionnaireInputSelected(q, a) {
			ordered = append(ordered, a.Text)
		}
		if len(ordered) > 0 {
			values[id] = ordered
		}
	}
	return values
}
func questionnaireEnabled(questions []Question) []int {
	result := []int{}
	for i, q := range questions {
		if !q.Disabled {
			result = append(result, i)
		}
	}
	return result
}
func (s *QuestionnaireState) prepare(p QuestionnaireProps) {
	s.init()
	if s.runtime == nil {
		s.runtime = map[string]*questionnaireRuntime{}
	}
	present := map[string]bool{}
	for i, q := range p.Questions {
		present[questionID(q, i)] = true
	}
	for id := range s.runtime {
		if !present[id] {
			delete(s.runtime, id)
			delete(s.Answers, id)
			delete(s.Errors, id)
		}
	}
	for i, q := range p.Questions {
		id := questionID(q, i)
		r := s.runtime[id]
		if r == nil {
			a, exists := s.Answers[id]
			if !exists {
				a = copyQuestionnaireAnswer(q.DefaultAnswer)
				s.Answers[id] = a
			}
			r = &questionnaireRuntime{controls: map[string]*questionnaireAnswerControl{}, hadAnswer: exists || len(q.DefaultAnswer.Values) > 0 || q.DefaultAnswer.Text != "", defaults: copyQuestionnaireAnswer(q.DefaultAnswer), reportedStatus: QuestionnaireUnanswered}
			s.runtime[id] = r
		}
		if r.multiple && !q.Multiple {
			a := s.Answers[id]
			for _, control := range r.order {
				if control.selected && !control.disabled {
					if control.input {
						a.Values = nil
						a.Mode = QuestionnaireInputAnswer
					} else {
						a.Values = []string{control.value}
						a.Mode = QuestionnaireChoiceAnswer
					}
					break
				}
			}
			s.Answers[id] = a
		}
		r.multiple = q.Multiple
		r.defaults = copyQuestionnaireAnswer(q.DefaultAnswer)
		r.radio = questionnaireRadioValue{}
		a := s.Answers[id]
		if v := validValues(s.effectiveQuestion(q, i), a); len(v) > 0 {
			r.radio = questionnaireRadioValue{value: v[0], selected: true}
		}
	}
	if !s.initialized {
		s.initialized = true
		s.defaultItem = p.DefaultItem
		if p.DefaultItem != "" {
			for i, q := range p.Questions {
				if questionID(q, i) == p.DefaultItem && !q.Disabled {
					s.Index = i
					break
				}
			}
		}
	}
	s.defaultItem = p.DefaultItem
	enabled := questionnaireEnabled(p.Questions)
	if len(enabled) == 0 {
		s.Index = 0
		return
	}
	if p.Item == nil && s.activeName != "" && s.Index == s.lastIndex && s.pendingName == "" {
		found := false
		for i, q := range p.Questions {
			if questionID(q, i) == s.activeName && !q.Disabled {
				s.Index = i
				found = true
				break
			}
		}
		if !found {
			s.Index = enabled[0]
		}
	}
	if p.Item != nil {
		s.Index = -1
		for i, q := range p.Questions {
			if questionID(q, i) == *p.Item && !q.Disabled {
				s.Index = i
				break
			}
		}
	}
	if p.Item == nil && (s.Index < 0 || s.Index >= len(p.Questions) || p.Questions[s.Index].Disabled) {
		s.Index = enabled[0]
	}
}
func (s *QuestionnaireState) snapshot(p QuestionnaireProps) QuestionnaireSnapshot {
	snap := QuestionnaireSnapshot{}
	enabled := questionnaireEnabled(p.Questions)
	snap.Total = len(enabled)
	for n, i := range enabled {
		if i == s.Index {
			snap.Current = n + 1
			snap.Item = questionID(p.Questions[i], i)
			snap.Status = s.Status(p.Questions[i], i)
		}
	}
	snap.First = snap.Current == 1 && snap.Total > 0
	snap.Last = snap.Current == snap.Total && snap.Total > 0
	return snap
}
func questionnaireError(q Question, status QuestionnaireItemStatus, a Answer) string {
	if q.Disabled || status == QuestionnaireSkipped && !q.Required {
		return ""
	}
	if q.Invalid {
		if q.Error != "" {
			return q.Error
		}
		return "This answer is invalid."
	}
	if q.Validate != nil {
		if message := q.Validate(copyQuestionnaireAnswer(a)); message != "" {
			return message
		}
	}
	if status == QuestionnaireAnswered {
		return ""
	}
	if q.Required {
		return "Choose an answer to continue."
	}
	return "Choose an answer or skip this question."
}

// Successful controls are registered native answers, including custom disabled
// slots. Inactive fieldsets remain registered and keep their answers.
func (s *QuestionnaireState) successfulChoices(q Question, index int, a Answer) []string {
	values := validValues(q, a)
	if r := s.runtime[questionID(q, index)]; r != nil {
		values = slices.DeleteFunc(values, func(v string) bool {
			for _, c := range r.order {
				if !c.input && c.value == v {
					return c.disabled
				}
			}
			return true
		})
	}
	return values
}
func (s *QuestionnaireState) inputDisabled(q Question, index int) bool {
	if r := s.runtime[questionID(q, index)]; r != nil {
		for _, c := range r.order {
			if c.input {
				return c.disabled
			}
		}
		return true
	}
	return false
}

type questionnaireOwnerKey struct{}
type questionnaireOwner struct{ state *QuestionnaireState }

func (o *questionnaireOwner) Close() error {
	if o.state != nil && o.state.owner == o {
		s := o.state
		s.pending = nil
		s.pendingName = ""
		s.pendingInvalid = false
		s.owner = nil
		for _, r := range s.runtime {
			r.composing = false
			r.item = ui.Handle{}
			for _, a := range r.controls {
				a.composing = false
				a.handle = ui.Handle{}
			}
		}
	}
	return nil
}

func questionnaireInputSelected(q Question, a Answer) bool {
	if a.Skipped {
		return false
	}
	switch a.Mode {
	case QuestionnaireInputAnswer, QuestionnaireCombinedAnswer:
		return true
	case QuestionnaireAutomaticAnswer:
		return strings.TrimSpace(a.Text) != "" && (q.Multiple || len(a.Values) == 0)
	default:
		return false
	}
}
func questionnaireInputMode(a *Answer, selected bool) {
	if selected {
		a.Mode = QuestionnaireInputAnswer
		if len(a.Values) > 0 {
			a.Mode = QuestionnaireCombinedAnswer
		}
	} else {
		a.Mode = QuestionnaireDraftAnswer
		if len(a.Values) > 0 {
			a.Mode = QuestionnaireChoiceAnswer
		}
	}
}
func (s *QuestionnaireState) sourceSelect(q Question, index int, value string) {
	a := s.Answers[questionID(q, index)]
	input := q.Multiple && questionnaireInputSelected(q, a)
	s.Select(q, index, value)
	a = s.Answers[questionID(q, index)]
	questionnaireInputMode(&a, input)
	s.Answers[questionID(q, index)] = a
}
