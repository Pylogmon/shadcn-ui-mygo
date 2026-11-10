package conversation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
)

// Question defines logical order and defaults. IDs and choice values must be
// unique within their scope. Compound Choice slots may supply native answers
// without duplicating optional Choices metadata.
type Question struct {
	ID, Title, Description string
	// Nil Choices uses registered native answer order for shortcuts in
	// QuestionnaireWith; a nonnil slice declares authoritative shortcut order.
	Choices                      []control.Option
	Required, Multiple, Freeform bool
	InputLabel, Placeholder      string
	Disabled, Invalid            bool
	Error                        string
	DefaultAnswer                Answer
	Validate                     func(Answer) string
	OnStatusChange               func(QuestionnaireItemStatus)
}
type Answer struct {
	Values []string
	Text   string
	// Inputs holds independent native input drafts once an item has more than
	// the legacy unnamed slot, or uses an explicit input Key. Text remains the
	// first unnamed slot's compatible value. Result includes successful inputs
	// only; State.Answers also retains inactive drafts and their selection.
	Inputs  []QuestionnaireInputValue
	Skipped bool
	// Mode records selection independently from the native text draft. Zero
	// retains legacy inference for restored Answers that predate compound slots.
	Mode QuestionnaireAnswerMode
}

// QuestionnaireInputValue identifies one input answer within a question.
// Key is an explicit Input Key, empty for the first unnamed slot, or a reserved
// NUL-prefixed identity for subsequent unnamed slots in declaration order.
// Text need not be unique: successful controls submit their values separately.
type QuestionnaireInputValue struct {
	Key, Text string
	Selected  bool
}

// QuestionnaireState is caller owned. Retain it for drafts across page changes;
// its private UI state stores models and public Handles, never frame Elements.
type QuestionnaireState struct {
	Index                   int
	Answers                 map[string]Answer
	Errors                  map[string]string
	Submitted               bool
	runtime                 map[string]*questionnaireRuntime
	initialized             bool
	activeName, pendingName string
	pendingInvalid          bool
	focusAttempts           int
	pending                 []questionnaireAction
	defaultItem             string
	lastIndex               int
	resetVersion            uint64
	owner                   *questionnaireOwner
	fallbackRequested       string
}

// QuestionnaireProps configures both convenience and compound rendering.
// DefaultItem/Item, callbacks other than OnSubmit, and Validators apply to
// QuestionnaireWith. The legacy Questionnaire keeps its original contract.
type QuestionnaireProps struct {
	Questions   []Question
	Disabled    bool
	OnSubmit    func(map[string]Answer)
	DefaultItem string
	// Item is a source-style controlled active name. Requests call OnItemChange;
	// the caller decides whether to update this value. Nil uses State.Index.
	Item           *string
	OnItemChange   func(string)
	Shortcuts      QuestionnaireShortcutMode
	OnStatusChange func(string, QuestionnaireItemStatus)
	OnAnswerChange func(string, Answer)
	OnReset        func(*QuestionnaireResetEvent)
	// ValidateAll runs after per-item validation on an explicit submission.
	// Its errors survive until an answer changes; transport remains app owned.
	ValidateAll func(map[string]Answer) map[string]string
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

// Select is the legacy choice coordinator. It toggles a valid choice, preserving a multi-selection's order.
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

// Next is the legacy required-only coordinator. It validates the active question. On the last question it validates all
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

// Previous is the legacy positional coordinator.
func (s *QuestionnaireState) Previous() {
	s.Index = max(0, s.Index-1)
}

// Skip is the legacy coordinator; it clears the skipped draft.
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

// Submit is the legacy required-only validator.
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
		if q.Disabled {
			continue
		}
		id := questionID(q, i)
		a := s.Answers[id]
		// Legacy callers never registered compound controls, and zero Mode keeps
		// their original ability to return fixed values plus freeform text.
		if s.runtime[id] == nil && a.Mode == QuestionnaireAutomaticAnswer && a.Inputs == nil {
			a.Values = validValues(q, a)
			if !q.Freeform {
				a.Text = ""
			}
			result[id] = a
			continue
		}
		q = s.effectiveQuestion(q, i)
		a.Values = s.successfulChoices(q, i, a)
		if a.Skipped {
			a.Values = nil
		}
		if a.Inputs != nil {
			a.Inputs = s.successfulInputs(q, i, a)
			a.Text = ""
			for _, input := range a.Inputs {
				if input.Key == "" {
					a.Text = input.Text
				}
			}
			questionnaireInputMode(&a, len(a.Inputs) > 0)
			result[id] = a
			continue
		}
		if a.Skipped {
			a.Values = nil
			a.Text = ""
		}
		if !q.Freeform || !questionnaireInputSelected(q, a) || s.inputDisabled(q, i) {
			a.Text = ""
		}
		result[id] = a
	}
	return result
}
