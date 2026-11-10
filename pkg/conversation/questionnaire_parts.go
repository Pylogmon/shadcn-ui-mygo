package conversation

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type QuestionnaireProgressProps struct {
	Label   string
	Content func(*ui.Context, QuestionnaireSnapshot)
}
type QuestionnaireActionProps struct {
	Label    string
	Disabled bool
	// Nil keeps the source default (outline for Previous/Skip/Reset).
	Variant *control.Variant
	Size    control.Size
	OnClick func(*QuestionnaireResetEvent)
}
type QuestionnaireChoiceEvent struct {
	Value     string
	Checked   bool
	prevented bool
}

func (e *QuestionnaireChoiceEvent) PreventDefault() { e.prevented = true }

type QuestionnaireChoiceProps struct {
	Label, Description string
	Disabled           bool
	Checked            *bool
	DefaultChecked     bool
	Handle             *ui.Handle
	Content            func(*ui.Context, bool)
	OnChange           func(*QuestionnaireChoiceEvent)
}
type QuestionnaireInputEvent struct {
	Value     string
	prevented bool
}

func (e *QuestionnaireInputEvent) PreventDefault() { e.prevented = true }

type QuestionnaireInputProps struct {
	// Key preserves this input's identity across reordering/insertion. Empty
	// keys use declaration order; the first unnamed slot also uses Answer.Text.
	// Explicit keys must be unique within the item and cannot contain NUL.
	Key                                  string
	Label, Placeholder                   string
	Disabled, ReadOnly, Password, Search bool
	Value                                *string
	DefaultValue                         string
	Handle                               *ui.Handle
	Validate                             func(string) string
	OnChange                             func(string)
	// OnInput may prevent selection/filled-state changes. An uncontrolled native
	// editor keeps its edited text, matching a prevented DOM change; Value is
	// authoritative for controlled editors. OnChange receives accepted edits.
	OnInput func(*QuestionnaireInputEvent)
	// Editor replaces the real native text control. Return its actual native
	// focusable Element; the string is a persistent native model. Disabled,
	// ReadOnly, Password, Search and Invalid still apply to that actual control.
	// Native multiline caret/IME behavior takes priority over form shortcuts.
	Editor func(*ui.Context, *string) ui.Element
}

// Parts are valid only inside this build. Durable state contains models and
// public Handles, never these Elements or the Context.
type QuestionnaireParts struct {
	Root            ui.Element
	State           QuestionnaireSnapshot
	c               *ui.Context
	model           *QuestionnaireState
	props           QuestionnaireProps
	disabled        bool
	inactive        bool
	choiceHandlers  map[string]func(bool)
	inputValidators map[string][]questionnaireInputValidator
}

type questionnaireInputValidator struct {
	key      string
	validate func() string
}

// QuestionnaireItemParts is build scoped. Active lets custom compositions gate
// imperative focus/entry animation while inactive controls remain mounted/inert.
type QuestionnaireItemParts struct {
	Root              ui.Element
	Active            bool
	Status            QuestionnaireItemStatus
	Invalid, Disabled bool
	Question          Question
	c                 *ui.Context
	root              *QuestionnaireParts
	runtime           *questionnaireRuntime
	index             int
	id                string
	controls          []ui.Element
	errorCount        int
	inputCount        int
}
type questionnaireSlot uint8

const (
	qProgressSlot questionnaireSlot = iota
	qActionsSlot
	qInputSlot
	qChoicesSlot
	qIndicatorSlot
	qLabelSlot
	qShortcutSlot
	qErrorSlot
)

type questionnaireItemKey string
type questionnaireChoiceKey string

type questionnaireLiveValue struct {
	text        string
	initialized bool
}

// QuestionnaireWith builds the strict source contract. Unlike legacy Next,
// optional unanswered items must be answered or explicitly skipped. The default
// build supplies every styled source slot; a custom build composes Cards/Dialogs
// freely. Questions supplies native logical order; no browser form/transport or
// HTML constraint-validation UI is synthesized. Use Validate/ValidateAll and
// Input.Validate for Go constraints, and Editor for other native editor types.
func QuestionnaireWith(c *ui.Context, state *QuestionnaireState, p QuestionnaireProps, build func(*QuestionnaireParts)) ui.Element {
	if state == nil {
		panic("conversation: QuestionnaireWith requires caller-owned state")
	}
	state.prepare(p)
	root := ui.Column(c).FillWidth().MinWidth(0).Gap(24).Role(ui.RoleGroup).Label("Questionnaire").Disabled(p.Disabled).DimDisabled(false)
	owner := ui.Local(root, questionnaireOwnerKey{}, func() questionnaireOwner { return questionnaireOwner{} })
	if owner.state != state {
		owner.Close()
		owner.state = state
	}
	state.owner = owner
	if p.Disabled {
		state.pending = nil
		state.pendingName = ""
		state.pendingInvalid = false
	}
	parts := QuestionnaireParts{Root: root, State: state.snapshot(p), c: c, model: state, props: p, disabled: p.Disabled, choiceHandlers: map[string]func(bool){}, inputValidators: map[string][]questionnaireInputValidator{}}
	activeRuntime := state.runtime[parts.State.Item]
	if !root.FocusWithin() && activeRuntime != nil {
		activeRuntime.composing = false
	}
	root.ObserveInput(func(ev ui.InputEvent) {
		if activeRuntime != nil {
			switch ev.Kind {
			case ui.InputCompose:
				activeRuntime.composing = ev.Text != ""
			case ui.InputText:
				activeRuntime.composing = false
			}
		}
	})

	root.Children(func() {
		if build != nil {
			build(&parts)
		} else {
			parts.Progress()
			for index, q := range p.Questions {
				id := questionID(q, index)
				parts.Item(id, func(item *QuestionnaireItemParts) {
					item.Header(func(*ui.Context) {
						item.Title(q.Title)
						if q.Description != "" {
							item.Description(q.Description)
						}
					})
					item.Choices(func(c *ui.Context) {
						for _, option := range q.Choices {
							item.Choice(option.Value)
						}
						if q.Freeform {
							item.Input()
						}
					})
					item.Error()
				})
			}
			parts.Actions(func(*ui.Context) { parts.Previous(); parts.Skip(); parts.Next(); parts.Submit() })
		}
	})
	parts.finish()
	return root
}
func (p *QuestionnaireParts) Progress(options ...QuestionnaireProgressProps) ui.Element {
	o := QuestionnaireProgressProps{}
	if len(options) > 0 {
		o = options[0]
	}
	label := o.Label
	if label == "" {
		label = "Questionnaire progress"
	}
	text := ""
	if p.State.Total > 0 {
		text = fmt.Sprintf("Question %d of %d", p.State.Current, p.State.Total)
	}
	font := ui.Font{Family: p.c.Theme().Font, Size: 12, Weight: 500}
	layout := ui.ShapeText("00000000000000", font, 0)
	width, _ := layout.Size()
	e := ui.Row(p.c.Key(qProgressSlot)).AlignSelf(ui.Start).MinWidth(width).MinHeight(16).FontSize(12).LineHeight(4.0 / 3).FontWeight(500).FontFeatures("tnum").TextColor(theme.TokensOf(p.c).MutedForeground).Role(ui.RoleProgress).Label(label)
	if p.State.Total > 0 {
		e.Range(1, float64(p.State.Total), float64(p.State.Current)).Value(text)
	}
	previous := ui.Local(e, "questionnaire-progress-live", func() questionnaireLiveValue { return questionnaireLiveValue{} })
	if !p.inactive && previous.initialized && previous.text != text && text != "" {
		p.c.Announce(text)
	}
	previous.text = text
	previous.initialized = true
	return e.Children(func() {
		if o.Content != nil {
			o.Content(p.c, p.State)
		} else {
			ui.Text(p.c, text)
		}
	})
}
func (p *QuestionnaireParts) Item(id string, build func(*QuestionnaireItemParts)) ui.Element {
	index := -1
	for n, q := range p.props.Questions {
		if questionID(q, n) == id {
			index = n
			break
		}
	}
	if index < 0 {
		return ui.Element{}
	}
	q := p.props.Questions[index]
	r := p.model.runtime[id]
	status := p.model.Status(q, index)
	previous := slices.Clone(r.order)
	r.order = r.order[:0]
	active := !q.Disabled && index == p.model.Index
	if !active {
		r.composing = false
	}
	root := ui.Column(p.c.Key(questionnaireItemKey(id))).FillWidth().MinWidth(0).Gap(20).Role(ui.RoleGroup).Label(q.Title).Focusable().TabStop(false).Bind(&r.item).Disabled(p.disabled || q.Disabled || !active).DimDisabled(false).Hidden(!active).AccessHidden(!active).IgnoreInput(!active).Invalid(p.invalid(q, index))
	item := QuestionnaireItemParts{Root: root, Active: active, Status: status, Invalid: p.invalid(q, index), Disabled: p.disabled || q.Disabled, Question: q, c: p.c, root: p, runtime: r, index: index, id: id}
	root.KeyShortcuts(ui.Super, ui.KeyEnter).KeyShortcuts(ui.Ctrl, ui.KeyEnter).KeyShortcuts(0, ui.KeyUp).KeyShortcuts(0, ui.KeyDown)
	root.Children(func() {
		before := p.inactive
		p.inactive = !active
		defer func() { p.inactive = before }()
		if build != nil {
			build(&item)
		}
	})

	// Only actual mounted native controls are successful. Removing an answer
	// disposes its local selection/default registration; hidden items keep them.
	a := p.model.Answers[id]
	for _, old := range previous {
		if !slices.Contains(r.order, old) {
			if old.input {
				questionnaireRemoveInput(&a, old.inputKey)
				questionnaireRemoveInput(&r.defaults, old.inputKey)
			} else {
				a.Values = slices.DeleteFunc(a.Values, func(v string) bool { return v == old.value })
			}
			delete(r.controls, old.value)
			p.c.Invalidate()
		}
	}
	p.model.Answers[id] = a
	if !q.Multiple {
		selected := false
		first := -1
		for j, answer := range r.order {
			if !answer.input && !answer.disabled {
				if first < 0 {
					first = j
				}
				selected = selected || answer.selected
			}
		}
		for j, answer := range r.order {
			if !answer.input && j < len(item.controls) {
				item.controls[j].TabStop(active && !answer.disabled && (answer.selected || !selected && j == first))
			}
		}
	}
	return root
}
func (i *QuestionnaireItemParts) Title(text string) ui.Element {
	e := theme.StyleHeading(i.c, ui.Text(i.c, text).FontSize(16).LineHeight(1.5).FontWeight(600).Pretty())
	i.Root.LabelledBy(e)
	return e
}
func (i *QuestionnaireItemParts) Description(text string) ui.Element {
	e := ui.Text(i.c, text).FontSize(14).LineHeight(10.0 / 7).TextColor(theme.TokensOf(i.c).MutedForeground).Pretty()
	i.Root.DescribedBy(e)
	return e
}
func (i *QuestionnaireItemParts) Choices(content func(*ui.Context)) ui.Element {
	return ui.Column(i.c.Key(qChoicesSlot)).FillWidth().MinWidth(0).Gap(12).Children(func() {
		if content != nil {
			content(i.c)
		}
	})
}
func (i *QuestionnaireItemParts) control(key string, input bool) *questionnaireAnswerControl {
	answer := i.runtime.controls[key]
	if answer == nil {
		answer = &questionnaireAnswerControl{value: key, input: input}
		i.runtime.controls[key] = answer
	}
	answer.disabled = i.root.disabled || i.Question.Disabled
	i.runtime.order = append(i.runtime.order, answer)
	return answer
}
func (i *QuestionnaireItemParts) Choice(value string, options ...QuestionnaireChoiceProps) ui.Element {
	props := QuestionnaireChoiceProps{}
	if len(options) > 0 {
		props = options[0]
	}
	for _, option := range i.Question.Choices {
		if option.Value == value {
			if props.Label == "" {
				props.Label = option.Label
			}
			if props.Description == "" {
				props.Description = option.Description
			}
			props.Disabled = props.Disabled || option.Disabled
			break
		}
	}
	if props.Label == "" {
		props.Label = value
	}
	controlState := i.control(value, false)
	controlState.disabled = controlState.disabled || props.Disabled
	model := i.root.model
	a := model.Answers[i.id]
	if props.DefaultChecked {
		if !slices.Contains(i.runtime.defaults.Values, value) {
			i.runtime.defaults.Values = append(i.runtime.defaults.Values, value)
		}
		if !controlState.initialized && !i.runtime.hadAnswer && (i.Question.Multiple || len(a.Values) == 0 && !questionnaireInputSelected(i.Question, a)) {
			model.sourceSelect(model.effectiveQuestion(i.Question, i.index), i.index, value)
			a = model.Answers[i.id]
		}
	}
	controlState.initialized = true
	if props.Checked != nil && (!controlState.controlled || controlState.controlledChecked != *props.Checked || controlState.resetVersion != model.resetVersion) {
		if *props.Checked {
			if i.Question.Multiple {
				if !slices.Contains(a.Values, value) {
					a.Values = append(a.Values, value)
				}
			} else {
				a.Values = []string{value}
				questionnaireDeselectInputs(&a)
				a.Mode = QuestionnaireChoiceAnswer
			}
			a.Skipped = false
		} else {
			a.Values = slices.DeleteFunc(a.Values, func(v string) bool { return v == value })
		}
		if i.Question.Multiple {
			questionnaireInputMode(&a, questionnaireInputSelected(i.Question, a))
		}
		model.Answers[i.id] = a
	}
	controlState.controlled = props.Checked != nil
	if props.Checked != nil {
		controlState.controlledChecked = *props.Checked
	}
	controlState.resetVersion = model.resetVersion

	checked := !a.Skipped && slices.Contains(a.Values, value)
	controlState.selected = checked
	before := checked
	var button ui.Element
	if i.Question.Multiple {
		button = ui.CheckboxBase(i.c.Key(questionnaireChoiceKey(value)), &controlState.selected)
	} else {
		if checked {
			i.runtime.radio = questionnaireRadioValue{value: value, selected: true}
		}
		button = ui.RadioBase(i.c.Key(questionnaireChoiceKey(value)), &i.runtime.radio, questionnaireRadioValue{value: value, selected: true})
	}
	button.Bind(&controlState.handle).Label(props.Label).Description(props.Description).Disabled(controlState.disabled || !i.Active).DimDisabled(false).FocusRing(false).FillWidth().MinWidth(0).MinHeight(44).Padding(14, 16).Gap(12).AlignItems(ui.Start).FontSize(14).LineHeight(1.375).Radius(theme.TokensOf(i.c).RadiusMD())
	if props.Handle != nil {
		button.Bind(props.Handle)
	}
	changed := button.Changed()
	selected := controlState.selected
	if !i.Question.Multiple {
		selected = i.runtime.radio == (questionnaireRadioValue{value: value, selected: true})
	}
	change := func(selected bool) {
		event := QuestionnaireChoiceEvent{Value: value, Checked: selected}
		if props.OnChange != nil {
			props.OnChange(&event)
		}
		a := model.Answers[i.id]
		if event.prevented {
			return
		}
		if props.Checked != nil {
			selected = *props.Checked
		}
		old := slices.Contains(a.Values, value)
		if old != selected {
			if i.Question.Multiple || selected {
				model.sourceSelect(model.effectiveQuestion(i.Question, i.index), i.index, value)
				a = model.Answers[i.id]
			}
		}
		if !i.Question.Multiple && selected {
			a.Values = []string{value}
			questionnaireDeselectInputs(&a)
			a.Mode = QuestionnaireChoiceAnswer
			a.Skipped = false
			model.Answers[i.id] = a
		}
		i.c.Invalidate()
	}
	i.root.choiceHandlers[i.id+"\x00"+value] = change
	if changed && selected != before {
		change(selected)
	}
	a = model.Answers[i.id]
	checked = !a.Skipped && slices.Contains(a.Values, value)
	controlState.selected = checked
	button.Checked(checked)

	tk := theme.TokensOf(i.c)
	invalid := i.root.invalid(i.Question, i.index)
	bg, border := ui.Transparent, tk.Input
	if tk.Dark {
		bg = tk.Input.Alpha(.2)
	}
	if button.Hovered() && !controlState.disabled {
		bg = tk.Muted.Alpha(.5)
	}
	if checked {
		bg, border = tk.Muted, tk.Primary.Alpha(.4)
	}
	if invalid {
		border = tk.Destructive
	}
	button.Background(bg).Border(1, border).Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, .05)).Invalid(invalid)
	if button.FocusVisible() {
		button.Border(1, tk.Ring).Shadow(0, 0, 0, 3, tk.Ring.Alpha(.5))
	}
	if controlState.disabled && i.Active {
		button.Opacity(.5)
	}
	questionnaireTransition(i.c, button)
	shortcut := i.shortcut(value)
	if shortcut != "" {
		button.KeyShortcuts(0, questionnaireKey(shortcut))
	}
	if checked {
		button.KeyShortcuts(0, ui.KeyEnter)
	}
	button.Children(func() {
		size := float32(4)
		if !i.Question.Multiple {
			size = 8
		}
		indicator := ui.Box(i.c.Key(qIndicatorSlot)).Size(16, 16).Shrink(0).Radius(size).Border(1, tk.Input).Background(ui.Transparent).Center().Translate(0, 1.8).IgnoreInput(true).AccessHidden(true)
		if props.Description != "" {
			indicator.Translate(0, 2)
		}
		if tk.Dark {
			indicator.Background(tk.Input.Alpha(.3))
		}
		if checked {
			indicator.Background(tk.Primary).Border(1, tk.Primary)
		}
		indicator.Children(func() {
			if checked {
				if i.Question.Multiple {
					ui.Icon(i.c, attachmentDefaultIcon("check")).Size(14, 14).TextColor(tk.PrimaryForeground)
				} else {
					ui.Box(i.c).Size(8, 8).Radius(4).Background(tk.PrimaryForeground)
				}
			}
		})
		ui.Column(i.c.Key(qLabelSlot)).Grow(1).MinWidth(0).Gap(4).Children(func() {
			if props.Content != nil {
				props.Content(i.c, checked)
			} else {
				ui.Text(i.c, props.Label)
				if props.Description != "" {
					QuestionnaireChoiceDescription(i.c, props.Description)
				}
			}
		})
		if shortcut != "" {
			ui.Box(i.c.Key(qShortcutSlot)).Size(20, 20).Shrink(0).Center().Translate(0, 1.8).Radius(tk.RadiusMD()).Border(1, tk.Input).Background(tk.Background).TextColor(tk.MutedForeground).Font("monospace").FontSize(10).FontWeight(500).LineHeight(1).Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, .05)).IgnoreInput(true).AccessHidden(true).Children(func() { ui.Text(i.c, shortcut) })
		}
	})
	i.controls = append(i.controls, button)
	return button
}
func QuestionnaireChoiceDescription(c *ui.Context, text string) ui.Element {
	return ui.Text(c, text).TextColor(theme.TokensOf(c).MutedForeground)
}
func (i *QuestionnaireItemParts) Input(options ...QuestionnaireInputProps) ui.Element {
	p := QuestionnaireInputProps{}
	if len(options) > 0 {
		p = options[0]
	}
	if p.Label == "" {
		p.Label = i.Question.InputLabel
	}
	if p.Label == "" {
		p.Label = "Your answer"
	}
	if p.Placeholder == "" {
		p.Placeholder = i.Question.Placeholder
	}
	key := p.Key
	if strings.ContainsRune(key, '\x00') {
		panic("conversation: Questionnaire Input Key cannot contain NUL")
	}
	if key == "" {
		i.inputCount++
		if i.inputCount > 1 {
			key = fmt.Sprintf("\x00input:%d", i.inputCount)
		}
	}
	for _, previous := range i.runtime.order {
		if previous.input && previous.inputKey == key {
			panic("conversation: duplicate Questionnaire Input Key: " + p.Key)
		}
	}
	controlState := i.control("\x00input"+key, true)
	controlState.inputKey = key
	controlState.disabled = controlState.disabled || p.Disabled
	model := i.root.model
	a := model.Answers[i.id]
	if key != "" {
		questionnaireManageInputs(i.Question, &a)
		model.Answers[i.id] = a
	}
	defaultText := p.DefaultValue
	if defaultText == "" {
		defaultText, _ = questionnaireInputValue(i.Question, i.Question.DefaultAnswer, key)
	}
	if key == "" && i.runtime.defaults.Inputs == nil {
		i.runtime.defaults.Text = defaultText
	} else {
		_, wasSelected := questionnaireInputValue(i.Question, i.runtime.defaults, key)
		selected := strings.TrimSpace(defaultText) != "" && (i.Question.Multiple || wasSelected || len(i.runtime.defaults.Values) == 0 && !questionnaireInputSelected(i.Question, i.runtime.defaults))
		if p.DefaultValue == "" {
			for _, input := range i.Question.DefaultAnswer.Inputs {
				if input.Key == key {
					selected = input.Selected
					break
				}
			}
		}
		questionnaireSetInput(i.Question, &i.runtime.defaults, key, defaultText, selected)
	}
	if !controlState.initialized && !i.runtime.hadAnswer && p.DefaultValue != "" {
		selected := i.Question.Multiple || len(a.Values) == 0 && !questionnaireInputSelected(i.Question, a)
		questionnaireSetInput(i.Question, &a, key, p.DefaultValue, selected)
		model.Answers[i.id] = a
	}
	controlState.initialized = true
	if p.Value != nil {
		changed := !controlState.controlled || controlState.controlledText != *p.Value || controlState.resetVersion != model.resetVersion
		_, selected := questionnaireInputValue(i.Question, a, key)
		if changed {
			selected = strings.TrimSpace(*p.Value) != ""
			if selected {
				a.Skipped = false
			}
		}
		questionnaireSetInput(i.Question, &a, key, *p.Value, selected)
		model.Answers[i.id] = a
	}

	controlState.controlled = p.Value != nil
	if p.Value != nil {
		controlState.controlledText = *p.Value
	}
	controlState.resetVersion = model.resetVersion
	controlState.text, controlState.selected = questionnaireInputValue(i.Question, a, key)
	previousText := controlState.text
	inputContext := i.c.Key(qInputSlot)
	if key != "" {
		inputContext = i.c.Key(struct {
			Slot questionnaireSlot
			Key  string
		}{qInputSlot, key})
	}
	var e ui.Element
	if p.Editor != nil {
		e = p.Editor(inputContext, &controlState.text)
	} else {
		e = control.Input(inputContext, &controlState.text, control.InputProps{Label: p.Label, Placeholder: p.Placeholder, Disabled: controlState.disabled || !i.Active, ReadOnly: p.ReadOnly, Password: p.Password, Search: p.Search, Invalid: i.root.invalid(i.Question, i.index)})
	}
	questionnaireTransition(i.c, e)
	if p.ReadOnly {
		e.ReadOnly(true)
	}
	if p.Password {
		e.Password()
	}
	if p.Search {
		e.Search()
	}
	e.Bind(&controlState.handle).Label(p.Label).Disabled(controlState.disabled || !i.Active).DimDisabled(false).Invalid(i.root.invalid(i.Question, i.index)).MinWidth(0).FillWidth().FontSize(14).MinHeight(36).Padding(4, 10)
	width, _ := i.c.Size()
	if width < 640 {
		e.MinHeight(44).FontSize(16)
	} else if width < 768 {
		e.FontSize(16)
	}
	if p.Handle != nil {
		e.Bind(p.Handle)
	}
	if e.Changed() {
		event := QuestionnaireInputEvent{Value: controlState.text}
		if p.OnInput != nil {
			p.OnInput(&event)
		}
		if !event.prevented {
			if p.OnChange != nil {
				p.OnChange(controlState.text)
			}
			if p.Value == nil {
				model.setInputText(i.Question, i.index, key, controlState.text)
			} else {
				controlState.text = *p.Value
				if controlState.text != previousText {
					model.setInputText(i.Question, i.index, key, controlState.text)
				}
			}

		} else {
			if p.Value != nil {
				controlState.text = *p.Value
			} else {
				questionnaireSetInput(i.Question, &a, key, controlState.text, controlState.selected)
				model.Answers[i.id] = a
			}
		}
		i.c.Invalidate()
	}
	if controlState.selected {
		e.KeyShortcuts(0, ui.KeyEnter)
	}
	a = model.Answers[i.id]
	_, controlState.selected = questionnaireInputValue(i.Question, a, key)
	if p.Validate != nil {
		i.root.inputValidators[i.id] = append(i.root.inputValidators[i.id], questionnaireInputValidator{key: controlState.value, validate: func() string {
			text, selected := questionnaireInputValue(i.Question, model.Answers[i.id], key)
			if selected && !controlState.disabled && strings.TrimSpace(text) != "" {
				return p.Validate(text)
			}
			return ""
		}})
	}
	controlState.composing = e.Composing()
	i.controls = append(i.controls, e)

	return e
}
func (i *QuestionnaireItemParts) Error(text ...string) ui.Element {
	message := i.root.model.Errors[i.id]
	if i.Question.Invalid && message == "" {
		message = i.Question.Error
		if message == "" {
			message = "This answer is invalid."
		}
	}
	if len(text) > 0 && text[0] != "" {
		message = text[0]
	}
	invalid := i.root.invalid(i.Question, i.index)
	i.errorCount++
	if message == "" {
		if i.Question.Required {
			message = "Choose an answer to continue."
		} else {
			message = "Choose an answer or skip this question."
		}
	}
	e := ui.Text(i.c.Key(struct {
		Slot  questionnaireSlot
		Index int
	}{qErrorSlot, i.errorCount}), message).FontSize(14).TextColor(theme.TokensOf(i.c).Destructive).Hidden(!invalid).AccessHidden(!invalid)
	if invalid {
		e.Role(ui.RoleAlert)
		i.Root.ErrorMessage(e)
		for _, control := range i.controls {
			control.ErrorMessage(e)
		}
	}

	return e
}
func (i *QuestionnaireItemParts) shortcut(value string) string {
	n := 0
	if i.Question.Choices == nil {
		for _, answer := range i.runtime.order {
			if answer.input || answer.disabled {
				continue
			}
			if answer.value == value {
				return questionnaireShortcut(i.root.props.Shortcuts, n)
			}
			n++
		}
		return ""
	}
	for _, o := range i.Question.Choices {
		if o.Disabled {
			continue
		}
		key := questionnaireShortcut(i.root.props.Shortcuts, n)
		if o.Value == value {
			return key
		}
		n++
	}
	return ""
}
func questionnaireShortcut(mode QuestionnaireShortcutMode, index int) string {
	if mode == QuestionnaireLetters && index < 26 {
		return string(rune('A' + index))
	}
	if mode == QuestionnaireNumbers && index < 9 {
		return string(rune('1' + index))
	}
	return ""
}
func questionnaireKey(value string) ui.Key {
	if len(value) != 1 {
		return ui.KeyUnknown
	}
	if value[0] >= 'A' && value[0] <= 'Z' {
		return ui.Key(int(ui.KeyA) + int(value[0]-'A'))
	}
	if value[0] >= '1' && value[0] <= '9' {
		return ui.Key(int(ui.Key1) + int(value[0]-'1'))
	}
	return ui.KeyUnknown
}
func (p *QuestionnaireParts) invalid(q Question, index int) bool {
	return !q.Disabled && !(p.model.Status(q, index) == QuestionnaireSkipped && !q.Required) && (q.Invalid || p.model.Errors[questionID(q, index)] != "")
}
func (p *QuestionnaireParts) Actions(content func(*ui.Context)) ui.Element {
	return ui.Grid(p.c.Key(qActionsSlot)).ColumnTracks(ui.Fr(1), ui.FitContent(), ui.FitContent()).FillWidth().MinWidth(0).MinHeight(36).Gap(8).AlignItems(ui.Center).Children(func() {
		if content != nil {
			content(p.c)
		}
	})
}
func (p *QuestionnaireParts) Previous(options ...QuestionnaireActionProps) ui.Element {
	return p.action(qPrevious, p.State.Total > 1 && !p.State.First, "Previous", control.Outline, options)
}
func (p *QuestionnaireParts) Skip(options ...QuestionnaireActionProps) ui.Element {
	visible := p.State.Current > 0 && !p.props.Questions[p.model.Index].Required
	return p.action(qSkip, visible, "Skip", control.Outline, options)
}
func (p *QuestionnaireParts) Next(options ...QuestionnaireActionProps) ui.Element {
	return p.action(qNext, p.State.Total > 1 && !p.State.Last, "Next", control.Default, options)
}
func (p *QuestionnaireParts) Submit(options ...QuestionnaireActionProps) ui.Element {
	return p.action(qSubmit, p.State.Total > 0 && p.State.Last, "Submit", control.Default, options)
}
func (p *QuestionnaireParts) Reset(options ...QuestionnaireActionProps) ui.Element {
	return p.action(qReset, true, "Reset", control.Outline, options)
}
func (p *QuestionnaireParts) action(kind questionnaireActionKind, visible bool, label string, variant control.Variant, options []QuestionnaireActionProps) ui.Element {
	o := QuestionnaireActionProps{}
	if len(options) > 0 {
		o = options[0]
	}
	if o.Label != "" {
		label = o.Label
	}
	if o.Variant != nil {
		variant = *o.Variant
	}
	button := control.Button(p.c.Key(kind), label, control.ButtonProps{Variant: variant, Size: o.Size, Disabled: p.disabled || o.Disabled}).Hidden(!visible).AccessHidden(!visible).IgnoreInput(!visible).Disabled(!visible || p.disabled || o.Disabled).TabStop(visible)
	column := 1
	if kind == qSkip {
		column = 2
	}
	if kind == qNext || kind == qSubmit {
		column = 3
	}
	if kind != qReset {
		button.ColumnStart(column).RowStart(1)
	}
	button.JustifySelf(ui.Start)
	if column > 1 {
		button.JustifySelf(ui.End)
	}
	width, _ := p.c.Size()
	if width < 640 {
		button.MinHeight(44)
	}
	if visible && !p.disabled && !o.Disabled && (kind == qNext || kind == qSubmit) {
		button.KeyShortcuts(0, ui.KeyEnter)
	}
	if button.Clicked() {
		event := QuestionnaireResetEvent{}
		if o.OnClick != nil {
			o.OnClick(&event)
		}
		if !event.prevented {
			p.model.pending = append(p.model.pending, questionnaireAction{kind: kind})
			p.c.Invalidate()
		}
	}
	return button
}

// Header reproduces the source fieldset legend/description block: its legend
// sits outside the fieldset's gapped content. Custom Card/Dialog headers can
// instead supply their own layout and call Title/Description within it.
func (i *QuestionnaireItemParts) Header(content func(*ui.Context)) ui.Element {
	return ui.Column(i.c).FillWidth().MinWidth(0).Gap(0).Children(func() {
		if content != nil {
			content(i.c)
		}
	})
}

func questionnaireTransition(c *ui.Context, e ui.Element) {
	if !motion.Reduced(c) {
		e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Ease: ui.CubicBezier(.4, 0, .2, 1), Colors: true})
	}
}
