package conversation

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
)

func (p *QuestionnaireParts) requestItem(index int, invalid bool) {
	if index < 0 || index >= len(p.props.Questions) || p.props.Questions[index].Disabled {
		return
	}
	name := questionID(p.props.Questions[index], index)
	p.model.pendingName, p.model.pendingInvalid = name, invalid
	p.model.focusAttempts = 0
	if p.props.Item == nil {
		p.model.Index = index
	}
	if p.props.OnItemChange != nil && name != p.State.Item {
		p.props.OnItemChange(name)
	}
	p.c.Invalidate()
}
func (p *QuestionnaireParts) validate(index int) bool {
	if index < 0 || index >= len(p.props.Questions) {
		return false
	}
	q := p.model.effectiveQuestion(p.props.Questions[index], index)
	if r := p.model.runtime[questionID(q, index)]; r != nil {
		r.attempted = true
	}
	id := questionID(q, index)
	message := questionnaireError(q, p.model.Status(q, index), p.model.Answers[id])
	if message == "" {
		if validate := p.inputValidators[id]; validate != nil {
			message = validate()
		}
	}
	if r := p.model.runtime[id]; r != nil && r.crossError != "" {
		a := p.model.Answers[id]
		old := r.crossAnswer
		if !slices.Equal(a.Values, old.Values) || a.Text != old.Text || a.Skipped != old.Skipped || a.Mode != old.Mode {
			r.crossError = ""
		} else if message == "" {
			message = r.crossError
		}
	}
	if message != "" {
		p.model.Errors[id] = message
		return false
	}
	delete(p.model.Errors, id)
	return true
}
func (p *QuestionnaireParts) submit() {
	first := -1
	for i, q := range p.props.Questions {
		if !q.Disabled && !p.validate(i) && first < 0 {
			first = i
		}
	}
	if first < 0 && p.props.ValidateAll != nil {
		for id, message := range p.props.ValidateAll(p.model.Result(p.props.Questions)) {
			if message == "" {
				continue
			}
			p.model.Errors[id] = message
			if r := p.model.runtime[id]; r != nil {
				r.crossError = message
				r.crossAnswer = copyQuestionnaireAnswer(p.model.Answers[id])
			}
		}
		for i, q := range p.props.Questions {
			if !q.Disabled && p.model.Errors[questionID(q, i)] != "" {
				first = i
				break
			}
		}
	}
	if first >= 0 {
		p.model.Submitted = false
		p.requestItem(first, true)
		return
	}
	p.model.Submitted = true
	if p.props.OnSubmit != nil {
		p.props.OnSubmit(p.model.Result(p.props.Questions))
	}
	p.c.Invalidate()
}
func (p *QuestionnaireParts) finish() {
	state := p.model
	// A controlled invalid name stays inactive until the caller accepts the
	// fallback request; never silently turn a controlled value into local state.
	if p.props.Item != nil && p.State.Current == 0 && p.State.Total > 0 {
		enabled := questionnaireEnabled(p.props.Questions)
		name := questionID(p.props.Questions[enabled[0]], enabled[0])
		request := *p.props.Item + "\x00" + name
		if state.fallbackRequested != request {
			state.fallbackRequested = request
			p.requestItem(enabled[0], false)
		}
	} else {
		state.fallbackRequested = ""
	}
	for index, q := range p.props.Questions {
		id := questionID(q, index)
		a := state.Answers[id]
		r := state.runtime[id]
		if r != nil && !q.Multiple && a.Mode == QuestionnaireAutomaticAnswer {
			for _, control := range r.order {
				if control.disabled {
					continue
				}
				if control.input && strings.TrimSpace(a.Text) != "" {
					a.Values = nil
					a.Mode = QuestionnaireInputAnswer
					break
				}
				if !control.input && slices.Contains(a.Values, control.value) {
					a.Values = []string{control.value}
					a.Mode = QuestionnaireChoiceAnswer
					break
				}
			}
			state.Answers[id] = a
		}
	}
	for index, q := range p.props.Questions {
		id := questionID(q, index)
		r := state.runtime[id]
		if r == nil {
			continue
		}
		a := state.Answers[id]
		if !slices.Equal(r.reported.Values, a.Values) || r.reported.Text != a.Text || r.reported.Skipped != a.Skipped || r.reported.Mode != a.Mode {
			for otherID, other := range state.runtime {
				if other.crossError != "" {
					other.crossError = ""
					delete(state.Errors, otherID)
					p.c.Invalidate()
				}
			}
			break
		}
	}

	for _, action := range append([]questionnaireAction(nil), state.pending...) {
		switch action.kind {
		case qReset:
			event := QuestionnaireResetEvent{}
			if p.props.OnReset != nil {
				p.props.OnReset(&event)
			}
			if !event.prevented {
				state.Reset(p.props.Questions)
				if p.props.OnItemChange != nil && state.Index >= 0 && state.Index < len(p.props.Questions) && !p.props.Questions[state.Index].Disabled {
					name := questionID(p.props.Questions[state.Index], state.Index)
					if name != p.State.Item {
						p.props.OnItemChange(name)
					}
				}
				p.c.Invalidate()
			}
		case qPrevious, qNext, qConfirm, qSkip, qSubmit:
			enabled := questionnaireEnabled(p.props.Questions)
			position := slices.Index(enabled, state.Index)
			if position < 0 {
				continue
			}
			switch action.kind {
			case qPrevious:
				if position > 0 {
					p.requestItem(enabled[position-1], false)
				}
			case qSkip:
				q := p.props.Questions[state.Index]
				if q.Required {
					continue
				}
				a := state.Answers[questionID(q, state.Index)]
				a.Values = nil
				a.Skipped = true
				state.Answers[questionID(q, state.Index)] = a
				delete(state.Errors, questionID(q, state.Index))
				state.Submitted = false
				if position < len(enabled)-1 {
					p.requestItem(enabled[position+1], false)
				} else {
					p.submit()
				}
			case qNext, qConfirm:
				if !p.validate(state.Index) {
					p.requestItem(state.Index, true)
					continue
				}
				if position < len(enabled)-1 {
					p.requestItem(enabled[position+1], false)
				} else if action.kind == qConfirm {
					p.submit()
				}
			case qSubmit:
				p.submit()
			}
		case qShortcut:
			if state.Index < 0 || state.Index >= len(p.props.Questions) {
				continue
			}
			q := p.props.Questions[state.Index]
			r := state.runtime[questionID(q, state.Index)]
			if r == nil {
				continue
			}
			for _, control := range r.order {
				if control.disabled || control.input || control.value != action.value {
					continue
				}
				p.choose(q, state.Index, control, !control.selected || !q.Multiple)
				control.handle.Focus(p.c)
				break
			}
		case qMoveFocus:
			if state.Index < 0 || state.Index >= len(p.props.Questions) {
				continue
			}
			q := p.props.Questions[state.Index]
			r := state.runtime[questionID(q, state.Index)]
			if r == nil {
				continue
			}
			for _, control := range r.order {
				if control.disabled || control.value != action.value {
					continue
				}
				if !control.input && !q.Multiple && !control.selected {
					p.choose(q, state.Index, control, true)
				}
				control.handle.Focus(p.c)
				break
			}
		}
	}
	state.pending = state.pending[:0]
	// Keep validation attempted until reset; editing into an unanswered state
	// exposes the error again instead of forgetting the attempted action.
	for index, q := range p.props.Questions {
		if r := state.runtime[questionID(q, index)]; r != nil && r.attempted {
			p.validate(index)
		}
	}
	// Notify only actual data/status changes, with current live callbacks.
	for index, q := range p.props.Questions {
		id := questionID(q, index)
		r := state.runtime[id]
		if r == nil {
			continue
		}
		a := state.Answers[id]
		status := state.Status(q, index)
		if r.reportedStatus != status {
			r.reportedStatus = status
			if q.OnStatusChange != nil {
				q.OnStatusChange(status)
			}
			if p.props.OnStatusChange != nil {
				p.props.OnStatusChange(id, status)
			}
			p.c.Invalidate()
		}
		if !slices.Equal(r.reported.Values, a.Values) || r.reported.Text != a.Text || r.reported.Skipped != a.Skipped || r.reported.Mode != a.Mode {
			r.reported = copyQuestionnaireAnswer(a)
			if p.props.OnAnswerChange != nil {
				p.props.OnAnswerChange(id, copyQuestionnaireAnswer(a))
			}
		}
	}
	snapshot := state.snapshot(p.props)
	if state.activeName != "" && state.activeName != snapshot.Item && state.pendingName == "" {
		state.pendingName = snapshot.Item
		if p.props.Item == nil && snapshot.Item != "" && p.props.OnItemChange != nil {
			p.props.OnItemChange(snapshot.Item)
		}
	}
	state.activeName = snapshot.Item
	state.lastIndex = state.Index
	if state.pendingName != "" && state.pendingName == snapshot.Item {
		if r := state.runtime[snapshot.Item]; r != nil {
			target := &r.item
			if state.pendingInvalid {
				var answer *questionnaireAnswerControl
				for _, control := range r.order {
					if control.disabled {
						continue
					}
					if answer == nil {
						answer = control
					}
					if control.input && control.selected {
						answer = control
						break
					}
				}
				if answer != nil {
					target = &answer.handle
				}
			}
			// An inert-to-active subtree may require one settled layout before native
			// focus accepts the target. Retry once with the current live handle only.
			if target.Focused(p.c) || state.focusAttempts >= 2 {
				state.pendingName = ""
				state.pendingInvalid = false
				state.focusAttempts = 0
			} else {
				target.Focus(p.c)
				p.c.Invalidate()
				state.focusAttempts++
			}
		}
	}

	p.installKeyboard(snapshot)
}
func (p *QuestionnaireParts) choose(q Question, index int, control *questionnaireAnswerControl, selected bool) {
	id := questionID(q, index)
	key := id + "\x00" + control.value
	if handler := p.choiceHandlers[key]; handler != nil {
		handler(selected)
		return
	}
	a := p.model.Answers[id]
	already := slices.Contains(a.Values, control.value)
	if already != selected {
		p.model.sourceSelect(p.model.effectiveQuestion(q, index), index, control.value)
		a = p.model.Answers[id]
		if !q.Multiple {
			a.Mode = QuestionnaireChoiceAnswer
		}
		p.model.Answers[id] = a
	}
	p.c.Invalidate()
}
func (p *QuestionnaireParts) installKeyboard(snapshot QuestionnaireSnapshot) {
	if snapshot.Item == "" || p.disabled {
		return
	}
	state := p.model
	r := state.runtime[snapshot.Item]
	if r == nil {
		return
	}
	var focused *questionnaireAnswerControl
	for _, control := range r.order {
		if control.handle.Focused(p.c) {
			focused = control
			break
		}
	}
	itemFocused := r.item.Focused(p.c)
	textEntry := p.c.FocusTakesText()
	multiple := p.props.Questions[state.Index].Multiple
	rtl := p.c.Direction() == ui.RightToLeft
	mode := p.props.Shortcuts
	shortcuts := map[ui.Key]string{}
	n := 0
	if p.props.Questions[state.Index].Choices == nil {
		for _, answer := range r.order {
			if answer.input || answer.disabled {
				continue
			}
			if key := questionnaireShortcut(mode, n); key != "" {
				shortcuts[questionnaireKey(key)] = answer.value
			}
			n++
		}
	}
	for _, option := range p.props.Questions[state.Index].Choices {
		if option.Disabled {
			continue
		}
		if key := questionnaireShortcut(mode, n); key != "" {
			shortcuts[questionnaireKey(key)] = option.Value
		}
		n++
	}
	handler := func(event ui.InputEvent) bool {

		if event.Kind != ui.InputKeyDown || r.composing {
			return false
		}
		queue := func(action questionnaireAction) {
			if !event.Repeat || action.kind == qMoveFocus {
				state.pending = append(state.pending, action)
			}
		}
		if event.Key == ui.KeyEnter && event.Mods&(ui.Ctrl|ui.Super) != 0 && event.Mods&^(ui.Ctrl|ui.Super) == 0 {
			queue(questionnaireAction{kind: qConfirm})
			return true
		}
		if event.Mods&^(ui.Shift) != 0 {
			return false
		}
		if event.Key == ui.KeyUp || event.Key == ui.KeyDown {
			if focused == nil && !itemFocused {
				return false
			}
			if focused != nil && focused.input && strings.TrimSpace(focused.text) != "" {
				return false
			}
			controls := []*questionnaireAnswerControl{}
			position := -1
			for _, control := range r.order {
				if !control.disabled {
					if control == focused {
						position = len(controls)
					}
					controls = append(controls, control)
				}
			}
			if len(controls) == 0 {
				return false
			}
			direction := 1
			if event.Key == ui.KeyUp {
				direction = -1
			}
			target := 0
			if position < 0 {
				if direction < 0 {
					target = len(controls) - 1
				}
				for j, control := range controls {
					if control.selected {
						target = j
						break
					}
				}
			} else {
				target = (position + direction + len(controls)) % len(controls)
			}
			if controls[target] == focused {
				return false
			}
			queue(questionnaireAction{kind: qMoveFocus, value: controls[target].value})
			return true
		}
		if event.Key == ui.KeyLeft || event.Key == ui.KeyRight {
			if textEntry || focused != nil && focused.input {
				return false
			}
			if focused != nil && !multiple {
				radios := []*questionnaireAnswerControl{}
				position := -1
				for _, control := range r.order {
					if !control.input && !control.disabled {
						if control == focused {
							position = len(radios)
						}
						radios = append(radios, control)
					}
				}
				if position < 0 || len(radios) < 2 {
					return false
				}
				direction := 1
				if event.Key == ui.KeyLeft {
					direction = -1
				}
				if rtl {
					direction = -direction
				}
				target := radios[(position+direction+len(radios))%len(radios)]
				queue(questionnaireAction{kind: qMoveFocus, value: target.value})
				return true
			}
			if event.Key == ui.KeyLeft {
				queue(questionnaireAction{kind: qPrevious})
			} else if snapshot.Status != QuestionnaireUnanswered {
				queue(questionnaireAction{kind: qNext})
			}
			return true
		}
		if event.Key == ui.KeyEnter && focused != nil {
			if focused.selected && (!focused.input || strings.TrimSpace(focused.text) != "") {
				queue(questionnaireAction{kind: qConfirm})
			}
			return true
		}
		if !textEntry && (focused == nil || !focused.input) {
			if value, ok := shortcuts[event.Key]; ok && (mode != QuestionnaireNumbers || event.Mods&ui.Shift == 0) {
				queue(questionnaireAction{kind: qShortcut, value: value})
				return true
			}
		}
		return false
	}
	// MyGo HandleInput keys belong to the focused node only. Scoped native
	// shortcuts on this root provide the source form's bubbling key policy
	// without replacing editor handlers or retaining an old Context.
	register := func(mods ui.Modifiers, key ui.Key) {
		if p.Root.Shortcut(mods, key) {
			handler(ui.InputEvent{Kind: ui.InputKeyDown, Mods: mods, Key: key, Repeat: p.c.ShortcutRepeated()})
			p.c.Invalidate()
		}
	}
	if !r.composing && (focused == nil || !focused.composing) {
		register(ui.Ctrl, ui.KeyEnter)
		register(ui.Super, ui.KeyEnter)
		register(ui.Ctrl|ui.Super, ui.KeyEnter)
		if (focused != nil && (!focused.input || strings.TrimSpace(focused.text) == "")) || itemFocused {
			register(0, ui.KeyUp)
			register(0, ui.KeyDown)
			register(ui.Shift, ui.KeyUp)
			register(ui.Shift, ui.KeyDown)
		}
		if !textEntry && (focused == nil || !focused.input) {
			register(0, ui.KeyLeft)
			register(0, ui.KeyRight)
			register(ui.Shift, ui.KeyLeft)
			register(ui.Shift, ui.KeyRight)
		}
		if focused != nil {
			register(0, ui.KeyEnter)
			register(ui.Shift, ui.KeyEnter)
		}
		if !textEntry && (focused == nil || !focused.input) {
			for key := range shortcuts {
				register(0, key)
				if mode == QuestionnaireLetters {
					register(ui.Shift, key)
				}
			}
		}
	}
}
