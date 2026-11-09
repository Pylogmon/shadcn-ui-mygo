package conversation

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func questionnaireSourceTester(s *QuestionnaireState, p *QuestionnaireProps, build func(*QuestionnaireParts)) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() { QuestionnaireWith(c, s, *p, build) })
	}, 560, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func questionnaireAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	v, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing native %v %q: %+v", role, name, r.AccessAll())
	}
	return v
}
func TestQuestionnaireSourceOptionalRequiresAnswerOrExplicitSkipAndSubmitsPerEvent(t *testing.T) {
	var state QuestionnaireState
	calls := 0
	p := QuestionnaireProps{Questions: []Question{{ID: "one", Title: "First", Required: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}, {ID: "two", Title: "Optional", Freeform: true}}, OnSubmit: func(map[string]Answer) { calls++ }}
	r := questionnaireSourceTester(&state, &p, nil)
	progress := questionnaireAccess(t, r, ui.RoleProgress, "Questionnaire progress")
	if progress.Min != 1 || progress.Max != 2 || progress.Now != 1 {
		t.Fatalf("progress is not real native range %+v", progress)
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Previous"); ok {
		t.Fatal("source previous must be hidden on first item")
	}
	if err := r.Click("Next"); err != nil {
		t.Fatal(err)
	}
	if !r.HasText("Choose an answer to continue.") || !r.Focused("Alpha") {
		t.Fatal("source invalid action did not expose error and focus first native answer")
	}
	r.Click("Alpha")
	r.Key(0, ui.KeyEnter)
	if state.Index != 1 || !r.Focused("Optional") {
		t.Fatalf("selected native Enter did not continue/focus fieldset: index=%d nodes=%+v", state.Index, r.AccessAll())
	}
	r.Click("Submit")
	if calls != 0 || !r.HasText("Choose an answer or skip this question.") {
		t.Fatal("optional unanswered item silently submitted")
	}
	r.Click("Skip")
	if calls != 1 || !state.Answers["two"].Skipped {
		t.Fatal("explicit optional skip did not submit")
	}
	if _, ok := state.FormValues(p.Questions)["two"]; ok {
		t.Fatal("skipped control leaked into form values")
	}
	r.Frame()
	r.Frame()
	if calls != 1 {
		t.Fatal("submission callback replayed without an input event")
	}
	r.Click("Submit")
	if calls != 2 {
		t.Fatal("source compound locked submission after first valid event")
	}
}
func TestQuestionnaireSourceSingleFreeformExcludesRetainedDraftAndMultipleUsesDOMOrder(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "answer", Title: "Answer", Required: true, Freeform: true, InputLabel: "Other", Choices: []control.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}}}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Other")
	r.Type("Custom draft")
	r.Click("Alpha")
	if state.Answers["answer"].Text != "Custom draft" || !slices.Equal(state.FormValues(p.Questions)["answer"], []string{"a"}) {
		t.Fatalf("single choice failed to exclude retained input draft: %+v", state.Answers)
	}
	r.Click("Other")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("New answer")
	if !slices.Equal(state.FormValues(p.Questions)["answer"], []string{"New answer"}) || questionnaireAccess(t, r, ui.RoleRadio, "Alpha").Checked {
		t.Fatal("typing did not select freeform exclusively")
	}
	p.Questions[0].Multiple = true
	state.Answers["answer"] = Answer{}
	r.Frame()
	r.Click("Beta")
	r.Click("Alpha")
	r.Click("Other")
	r.Type("Also text")
	if got := state.FormValues(p.Questions)["answer"]; !slices.Equal(got, []string{"a", "b", "Also text"}) {
		t.Fatalf("multiple form values are not DOM order followed by freeform: %v", got)
	}
}
func TestQuestionnaireSourceShortcutsAndIMEKeepNativeEditing(t *testing.T) {
	var state QuestionnaireState
	calls := 0
	p := QuestionnaireProps{Shortcuts: QuestionnaireNumbers, Questions: []Question{{ID: "one", Title: "Choose", Required: true, Freeform: true, InputLabel: "Other", Choices: []control.Option{{Value: "blocked", Label: "Blocked", Disabled: true}, {Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}}, OnSubmit: func(map[string]Answer) { calls++ }}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.Key2)
	if !slices.Equal(state.Answers["one"].Values, []string{"b"}) || !r.Focused("Beta") {
		t.Fatal("numeric source shortcut failed to skip disabled option")
	}
	r.Click("Other")
	r.Compose("候", 1)
	r.Key(ui.Cmd, ui.KeyEnter)
	if calls != 0 {
		t.Fatal("IME confirmation submitted the questionnaire")
	}
	r.Type("候補")
	r.Key(0, ui.Key1)
	if len(state.Answers["one"].Values) > 0 {
		t.Fatal("answer shortcut intercepted native typing")
	}
	r.Key(ui.Cmd, ui.KeyEnter)
	if calls != 1 {
		t.Fatal("CmdEnter did not confirm the composed input")
	}
}
func TestQuestionnaireSourceSkipRetainsDraftButExcludesItUntilEdited(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "one", Title: "Optional", Freeform: true, InputLabel: "Other"}, {ID: "two", Title: "Last", Required: true, Choices: []control.Option{{Value: "ok", Label: "Okay"}}}}}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Other")
	r.Type("Keep draft")
	r.Click("Skip")
	r.Click("Previous")
	if value := questionnaireAccess(t, r, ui.RoleTextField, "Other").Value; value != "Keep draft" {
		t.Fatalf("skip destroyed native input draft %q", value)
	}
	if _, ok := state.FormValues(p.Questions)["one"]; ok {
		t.Fatal("skipped draft still submitted")
	}
	r.Click("Other")
	r.Type(" updated")
	if state.Status(p.Questions[0], 0) != QuestionnaireAnswered {
		t.Fatal("editing a skipped input did not restore answered status")
	}
}
func TestQuestionnaireSourceDefaultsResumeResetAndDisabledCollection(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{DefaultItem: "second", Questions: []Question{{ID: "first", Title: "First", Required: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}}, DefaultAnswer: Answer{Values: []string{"a"}}}, {ID: "disabled", Title: "Excluded", Disabled: true, Required: true}, {ID: "second", Title: "Second", Freeform: true, InputLabel: "Saved input", DefaultAnswer: Answer{Text: "Saved draft"}}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Progress()
		for n, q := range p.Questions {
			parts.Item(questionID(q, n), func(item *QuestionnaireItemParts) {
				item.Title(q.Title)
				item.Choices(func(c *ui.Context) {
					for _, o := range q.Choices {
						item.Choice(o.Value)
					}
					if q.Freeform {
						item.Input()
					}
				})
				item.Error()
			})
		}
		parts.Actions(func(*ui.Context) { parts.Reset(); parts.Previous(); parts.Skip(); parts.Submit() })
	})
	if state.Index != 2 || questionnaireAccess(t, r, ui.RoleProgress, "Questionnaire progress").Max != 2 {
		t.Fatal("resume or disabled collection source order was ignored")
	}
	r.Click("Saved input")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Changed")
	r.Click("Reset")
	if state.Index != 2 || questionnaireAccess(t, r, ui.RoleTextField, "Saved input").Value != "Saved draft" {
		t.Fatal("source reset failed to restore default answers and initial item")
	}
	if _, ok := state.FormValues(p.Questions)["disabled"]; ok {
		t.Fatal("disabled fieldset leaked into submitted form values")
	}
}
func TestQuestionnaireSourceNativeSelectionSurvivesHiddenInactiveMount(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "one", Title: "Editable", Required: true, Freeform: true, InputLabel: "Draft"}, {ID: "two", Title: "Second", Required: true, Choices: []control.Option{{Value: "ok", Label: "Okay"}}}}}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Draft")
	r.Type("Original selection")
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(ui.Cmd, ui.KeyEnter)
	if state.Index != 1 {
		t.Fatal("could not continue from real input")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Draft"); ok {
		t.Fatal("inactive native input remained in AX")
	}
	r.Click("Previous")
	r.Key(0, ui.KeyDown)
	r.Type("Replaced")
	if state.Answers["one"].Text != "Replaced" {
		t.Fatalf("inactive item remounted or lost real native selection: %q", state.Answers["one"].Text)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("plain source questionnaire did not settle idle")
	}
}

func TestQuestionnaireSourceControlledSkipRejectedEditAndRetainedDraft(t *testing.T) {
	var state QuestionnaireState
	checked, value := true, ""
	calls := 0
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Controlled", Choices: []control.Option{{Value: "fixed", Label: "Fixed"}}, Freeform: true}}, OnSubmit: func(map[string]Answer) { calls++ }}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Item("a", func(i *QuestionnaireItemParts) {
			i.Title("Controlled")
			i.Choice("fixed", QuestionnaireChoiceProps{Checked: &checked})
			i.Input(QuestionnaireInputProps{Label: "Controlled input", Value: &value})
		})
		parts.Actions(func(*ui.Context) { parts.Skip(); parts.Submit() })
	})
	r.Click("Skip")
	if calls != 1 || !state.Answers["a"].Skipped || questionnaireAccess(t, r, ui.RoleRadio, "Fixed").Checked {
		t.Fatalf("unchanged controlled checked resurrected skip: %+v", state.Answers)
	}
	r.Click("Controlled input")
	r.Type("Rejected")
	if !state.Answers["a"].Skipped || len(state.FormValues(p.Questions)) != 0 || questionnaireAccess(t, r, ui.RoleTextField, "Controlled input").Value != "" {
		t.Fatal("rejected controlled edit selected an answer")
	}
	value = "Draft"
	checked = false
	r.Frame()
	if !slices.Equal(state.FormValues(p.Questions)["a"], []string{"Draft"}) {
		t.Fatal("accepted controlled value did not select input")
	}
	checked = true
	r.Frame()
	if !slices.Equal(state.FormValues(p.Questions)["a"], []string{"fixed"}) {
		t.Fatalf("unchanged controlled draft reclaimed choice: %+v", state.Answers)
	}
	r.Frame()
	if state.Answers["a"].Mode != QuestionnaireChoiceAnswer {
		t.Fatal("controlled synchronization reran without prop change")
	}
}
func TestQuestionnaireSourceValidationStaysAttemptedAndSkippedExternalErrorIsValid(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Answer", Required: true, Multiple: true, Choices: []control.Option{{Value: "fixed", Label: "Fixed"}}}}}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Submit")
	r.Click("Fixed")
	if r.HasText("Choose an answer to continue.") {
		t.Fatal("valid answer retained error")
	}
	r.Click("Fixed")
	if !r.HasText("Choose an answer to continue.") {
		t.Fatal("deselecting answer forgot attempted validation")
	}
	p.Questions[0].Required = false
	p.Questions[0].Invalid = true
	r.Frame()
	r.Click("Skip")
	if questionnaireAccess(t, r, ui.RoleGroup, "Answer").Invalid || r.HasText("This answer is invalid.") {
		t.Fatal("intentional optional skip retained external invalidity")
	}
}
func TestQuestionnaireSourceInitialDefaultDoesNotStealFocusAndReorderKeepsIdentity(t *testing.T) {
	var state QuestionnaireState
	outside := ""
	p := QuestionnaireProps{DefaultItem: "b", Questions: []Question{{ID: "a", Title: "Alpha", Freeform: true, InputLabel: "A input"}, {ID: "b", Title: "Beta", Freeform: true, InputLabel: "B input"}}}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			control.Input(c, &outside, control.InputProps{Label: "Outside"}).AutoFocus()
			QuestionnaireWith(c, &state, p, nil)
		})
	}, 560, 700)
	if !r.Focused("Outside") {
		t.Fatal("initial defaultItem stole outside autofocus")
	}
	r.Click("B input")
	r.Type("Keep selection")
	r.Key(ui.Cmd, ui.KeyA)
	p.Questions[0], p.Questions[1] = p.Questions[1], p.Questions[0]
	r.Frame()
	if state.Index != 0 || !r.Focused("B input") {
		t.Fatal("reordered named item lost active identity/focus")
	}
	r.Type("Replacement")
	if state.Answers["b"].Text != "Replacement" {
		t.Fatal("reordered named item lost native editor selection")
	}
}
func TestQuestionnaireSourceScopedShortcutsDoNotCaptureOtherFormOrExternalInput(t *testing.T) {
	var left, right QuestionnaireState
	a, b := 0, 0
	outside := ""
	p := QuestionnaireProps{Shortcuts: QuestionnaireLetters, Questions: []Question{{ID: "a", Title: "First form", Required: true, Choices: []control.Option{{Value: "one", Label: "Left one"}, {Value: "two", Label: "Left two"}}}}, OnSubmit: func(map[string]Answer) { a++ }}
	q := QuestionnaireProps{Shortcuts: QuestionnaireNumbers, Questions: []Question{{ID: "b", Title: "Second form", Required: true, Choices: []control.Option{{Value: "one", Label: "Right one"}, {Value: "two", Label: "Right two"}}}}, OnSubmit: func(map[string]Answer) { b++ }}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			QuestionnaireWith(c.Key("left"), &left, p, nil)
			QuestionnaireWith(c.Key("right"), &right, q, nil)
			control.Input(c, &outside, control.InputProps{Label: "Outside"})
		})
	}, 600, 900)
	r.Click("Left one")
	r.Key(0, ui.KeyB)
	r.Key(ui.Cmd, ui.KeyEnter)
	if a != 1 || b != 0 || !slices.Equal(left.Answers["a"].Values, []string{"two"}) || len(right.Answers["b"].Values) > 0 {
		t.Fatal("left root shortcut affected another form")
	}
	r.Click("Right one")
	r.Key(0, ui.Key2)
	r.Key(0, ui.KeyEnter)
	if a != 1 || b != 1 || !slices.Equal(right.Answers["b"].Values, []string{"two"}) {
		t.Fatal("numeric shortcut escaped its focused form")
	}
	r.Click("Outside")
	r.Type("abc12")
	r.Key(0, ui.KeyEnter)
	r.Key(ui.Cmd, ui.KeyEnter)
	r.Key(0, ui.KeyA)
	r.Key(0, ui.Key1)
	if a != 1 || b != 1 || outside != "abc12" {
		t.Fatal("external editor was captured by questionnaire shortcuts")
	}
}

func TestQuestionnaireSourceArrowNavigationRespectsEditorCaretAndRadioSelection(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Choose", Required: true, Freeform: true, InputLabel: "Other", Choices: []control.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}, {ID: "b", Title: "Next item", Required: true, Freeform: true}}}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Alpha")
	r.Key(0, ui.KeyRight)
	if !r.Focused("Beta") || !questionnaireAccess(t, r, ui.RoleRadio, "Beta").Checked || state.Index != 0 {
		t.Fatal("radio arrows navigated form instead of native selection")
	}
	r.Key(0, ui.KeyDown)
	if !r.Focused("Other") {
		t.Fatal("vertical navigation did not reach empty native freeform")
	}
	r.Type("hello")
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyRight)
	r.Type("X")
	if state.Answers["a"].Text != "hXello" || state.Index != 0 {
		t.Fatalf("horizontal form policy consumed native caret: %q", state.Answers["a"].Text)
	}
	r.Key(0, ui.KeyUp)
	if !r.Focused("Other") {
		t.Fatal("nonempty native text arrow moved answer focus")
	}
	r.Key(ui.Cmd, ui.KeyEnter)
	r.Key(0, ui.KeyLeft)
	if state.Index != 0 {
		t.Fatal("fieldset left did not navigate previous item")
	}
	r.Key(0, ui.KeyRight)
	if state.Index != 1 {
		t.Fatal("answered fieldset right did not navigate next item")
	}
}
func TestQuestionnaireSourceControlledNavigationCanRejectAndAcceptAndResetCanPrevent(t *testing.T) {
	var state QuestionnaireState
	active := "a"
	requested := ""
	accept := false
	preventReset := true
	p := QuestionnaireProps{Item: &active, Questions: []Question{{ID: "a", Title: "First", Required: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}}}, {ID: "b", Title: "Second", Freeform: true}}, OnItemChange: func(name string) {
		requested = name
		if accept {
			active = name
		}
	}, OnReset: func(e *QuestionnaireResetEvent) {
		if preventReset {
			e.PreventDefault()
		}
	}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		for n, q := range p.Questions {
			parts.Item(questionID(q, n), func(i *QuestionnaireItemParts) {
				i.Title(q.Title)
				for _, o := range q.Choices {
					i.Choice(o.Value)
				}
				if q.Freeform {
					i.Input()
				}
			})
		}
		parts.Actions(func(*ui.Context) { parts.Previous(); parts.Next(); parts.Reset() })
	})
	r.Click("Alpha")
	r.Click("Next")
	if requested != "b" || state.Index != 0 || active != "a" {
		t.Fatal("rejected controlled navigation mutated current item")
	}
	accept = true
	r.Click("Next")
	if state.Index != 1 || !r.Focused("Second") {
		t.Fatal("accepted controlled navigation failed focus")
	}
	r.Click("Reset")
	if len(state.Answers["a"].Values) != 1 {
		t.Fatal("prevented reset still cleared answers")
	}
	preventReset = false
	r.Click("Reset")
	if active != "a" || state.Index != 0 || len(state.Answers["a"].Values) > 0 {
		t.Fatal("reset did not request controlled initial item/defaults")
	}
}
func TestQuestionnaireSourceCustomValidationAllTargetsEarlierNativeInput(t *testing.T) {
	var state QuestionnaireState
	submitted := 0
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Contact", Required: true, Freeform: true, InputLabel: "Email"}, {ID: "b", Title: "Confirm", Required: true, Choices: []control.Option{{Value: "ok", Label: "Okay"}}}}, ValidateAll: func(a map[string]Answer) map[string]string {
		if a["a"].Text != "valid" {
			return map[string]string{"a": "Use valid"}
		}
		return nil
	}, OnSubmit: func(map[string]Answer) { submitted++ }}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Email")
	r.Type("invalid")
	r.Key(ui.Cmd, ui.KeyEnter)
	r.Click("Okay")
	r.Click("Submit")
	if state.Index != 0 || !r.Focused("Email") || !r.HasText("Use valid") || submitted != 0 {
		t.Fatalf("cross validation did not reveal earlier input: index%d errors%v", state.Index, state.Errors)
	}
	if !questionnaireAccess(t, r, ui.RoleTextField, "Email").Invalid {
		t.Fatal("native input lacks invalid AX state")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("valid")
	r.Key(ui.Cmd, ui.KeyEnter)
	r.Click("Submit")
	if submitted != 1 {
		t.Fatal("accepted correction did not submit")
	}
}
func TestQuestionnaireSourceDisabledSlotsAndInactiveValidatorUseSuccessfulControls(t *testing.T) {
	var state QuestionnaireState
	blocked := true
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "First", Multiple: true, Required: true, Freeform: true, Choices: []control.Option{{Value: "blocked", Label: "Blocked"}, {Value: "ok", Label: "Okay"}}}, {ID: "b", Title: "Last", Required: true, Choices: []control.Option{{Value: "yes", Label: "Yes"}}}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		for n, q := range p.Questions {
			parts.Item(questionID(q, n), func(i *QuestionnaireItemParts) {
				i.Title(q.Title)
				for _, o := range q.Choices {
					i.Choice(o.Value, QuestionnaireChoiceProps{Disabled: o.Value == "blocked" && blocked})
				}
				if q.Freeform {
					i.Input(QuestionnaireInputProps{Label: "Draft", Disabled: blocked, Validate: func(string) string { return "Invalid draft" }})
				}
				i.Error()
			})
		}
		parts.Actions(func(*ui.Context) { parts.Previous(); parts.Next(); parts.Submit() })
	})
	state.Answers["a"] = Answer{Values: []string{"blocked"}, Text: "Disabled draft"}
	r.Frame()
	r.Click("Next")
	if state.Index != 0 || state.Status(p.Questions[0], 0) != QuestionnaireUnanswered || len(state.FormValues(p.Questions)["a"]) > 0 {
		t.Fatal("disabled custom slots were treated as successful controls")
	}
	r.Click("Okay")
	r.Click("Next")
	r.Click("Yes")
	r.Click("Submit")
	if !state.Submitted {
		t.Fatal("disabled input validator blocked an inactive valid item")
	}
}
func TestQuestionnaireSourceMultipleToSingleNormalizesPersistentlyAndFormValueOrder(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Answers", Multiple: true, Freeform: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Item("a", func(i *QuestionnaireItemParts) {
			i.Title("Answers")
			i.Input(QuestionnaireInputProps{Label: "First input"})
			i.Choice("b")
			i.Choice("a")
		})
		parts.Submit()
	})
	r.Click("Alpha")
	r.Click("Beta")
	r.Click("First input")
	r.Type("First")
	if !slices.Equal(state.FormValues(p.Questions)["a"], []string{"First", "b", "a"}) {
		t.Fatalf("form values do not follow rendered control order: %v", state.FormValues(p.Questions))
	}
	p.Questions[0].Multiple = false
	r.Frame()
	p.Questions[0].Multiple = true
	r.Frame()
	if len(state.Answers["a"].Values) > 0 || state.Answers["a"].Mode != QuestionnaireInputAnswer {
		t.Fatal("switching back to multiple resurrected discarded selections")
	}
}

func TestQuestionnaireSourceControlledInvalidNameRequestsFallbackWithoutInventingActiveItem(t *testing.T) {
	var state QuestionnaireState
	active := "missing"
	requested := ""
	calls := 0
	p := QuestionnaireProps{Item: &active, OnItemChange: func(name string) { requested = name; calls++ }, Questions: []Question{{ID: "a", Title: "First", Freeform: true}}}
	r := questionnaireSourceTester(&state, &p, nil)
	if requested != "a" || calls != 1 || state.Index != -1 {
		t.Fatalf("invalid controlled name silently activated item: requested=%s calls%d index%d", requested, calls, state.Index)
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Your answer"); ok {
		t.Fatal("unaccepted controlled fallback exposed answer")
	}
	r.Frame()
	if calls != 1 {
		t.Fatal("controlled fallback request replayed while parent refused")
	}
	active = "a"
	r.Frame()
	if !r.Focused("First") {
		t.Fatalf("accepting fallback did not honor pending item focus: index%d pending%q active%q %+v", state.Index, state.pendingName, state.activeName, r.AccessAll())
	}
}
func TestQuestionnaireSourceCancellableNativeEventsAndDynamicDefaults(t *testing.T) {
	var state QuestionnaireState
	prevent := true
	choiceDefault := false
	inputDefault := "Saved"
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Editable", Freeform: true, Choices: []control.Option{{Value: "fixed", Label: "Fixed"}}}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Item("a", func(i *QuestionnaireItemParts) {
			i.Title("Editable")
			i.Choice("fixed", QuestionnaireChoiceProps{DefaultChecked: choiceDefault, OnChange: func(e *QuestionnaireChoiceEvent) {
				if prevent {
					e.PreventDefault()
				}
			}})
			i.Input(QuestionnaireInputProps{Label: "Draft", DefaultValue: inputDefault, OnInput: func(e *QuestionnaireInputEvent) {
				if prevent {
					e.PreventDefault()
				}
			}})
		})
		parts.Actions(func(*ui.Context) { parts.Reset(); parts.Submit() })
	})
	r.Click("Fixed")
	if questionnaireAccess(t, r, ui.RoleRadio, "Fixed").Checked {
		t.Fatal("prevented native choice change persisted")
	}
	r.Click("Draft")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Rejected")
	if state.Answers["a"].Text != "Rejected" || questionnaireAccess(t, r, ui.RoleTextField, "Draft").Value != "Rejected" {
		t.Fatal("prevented uncontrolled edit did not retain native value")
	}
	prevent = false
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Edited")
	choiceDefault = true
	inputDefault = ""
	r.Frame()
	if state.Answers["a"].Text != "Edited" {
		t.Fatal("changed defaults replaced live edit")
	}
	r.Click("Reset")
	if !questionnaireAccess(t, r, ui.RoleRadio, "Fixed").Checked || questionnaireAccess(t, r, ui.RoleTextField, "Draft").Value != "" {
		t.Fatalf("reset did not adopt latest defaults: %+v", state.Answers)
	}
}
func TestQuestionnaireSourceUnmountClearsIMEAndQueuedWorkWithoutRetainingFrame(t *testing.T) {
	var state QuestionnaireState
	mounted := true
	calls := 0
	outside := ""
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Answer", Required: true, Freeform: true, InputLabel: "Draft"}}, OnSubmit: func(map[string]Answer) { calls++ }}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			if mounted {
				QuestionnaireWith(c.Key("q"), &state, p, nil)
			}
			control.Input(c, &outside, control.InputProps{Label: "Outside"})
		})
	}, 560, 700)
	r.Click("Draft")
	r.Type("Keep")
	r.Compose("未", 1)
	state.pending = append(state.pending, questionnaireAction{kind: qSubmit})
	mounted = false
	r.Frame()
	if len(state.pending) > 0 {
		t.Fatal("unmounted owner retained queued frame work")
	}
	// An application can discard the entire state; a remount of this retained
	// model starts a fresh owner and must never replay old frame work.
	mounted = true
	r.Frame()
	r.Click("Draft")
	r.Key(ui.Cmd, ui.KeyEnter)
	if calls != 1 {
		t.Fatal("unmounted composition blocked remounted native confirmation")
	}
	r.Frame()
	if calls != 1 {
		t.Fatal("remount replayed callback")
	}
	r.Click("Outside")
	p.Disabled = true
	r.Frame()
	r.Key(ui.Cmd, ui.KeyEnter)
	if calls != 1 {
		t.Fatal("disabled form intercepted outside confirmation")
	}
}
func TestQuestionnaireSourceCustomNativeEditorAndProgressComposition(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Custom", Required: true}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Progress(QuestionnaireProgressProps{Label: "Custom progress", Content: func(c *ui.Context, s QuestionnaireSnapshot) { ui.Text(c, "Checkpoint "+string(rune('0'+s.Current))) }})
		parts.Item("a", func(i *QuestionnaireItemParts) {
			ui.Column(i.c).Padding(24).Gap(6).Border(1, theme.TokensOf(i.c).Border).Children(func() {
				i.Title("Custom")
				i.Description("Native multiline editor")
				i.Input(QuestionnaireInputProps{Label: "Custom editor", Editor: func(c *ui.Context, value *string) ui.Element {
					return control.Textarea(c, value, control.TextareaProps{Label: "Custom editor", Rows: 2})
				}})
				i.Error()
			})
		})
		parts.Submit()
	})
	r.Click("Custom editor")
	r.Type("First")
	r.Key(0, ui.KeyEnter)
	r.Key(ui.Cmd, ui.KeyEnter)
	// Selected freeform Enter has the source confirm behavior; no fake textarea
	// or text selection is substituted for the actual custom native control.
	if !state.Submitted || state.Answers["a"].Text != "First\n" {
		t.Fatal("custom editor did not participate in native form behavior")
	}
	progress := questionnaireAccess(t, r, ui.RoleProgress, "Custom progress")
	if progress.Now != 1 || !r.HasText("Checkpoint 1") {
		t.Fatal("custom progress lost real range")
	}
}
func TestQuestionnaireSourceVegaGeometryAndPixelsAcrossThemeAndBreakpoints(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, width := range []int{500, 700, 900} {
			t.Run(fmt.Sprintf("dark%v/width%d", dark, width), func(t *testing.T) {
				var state QuestionnaireState
				var tk *theme.Tokens
				if dark {
					tk = theme.Dark()
				} else {
					tk = theme.Light()
				}
				p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Title", Description: "Description", Freeform: true, InputLabel: "Draft", Choices: []control.Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}}}}
				r := ui.NewTesterWith(func(c *ui.Context) {
					theme.Scope(c, tk).Fill().Background(ui.RGB(50, 90, 150)).Padding(20).Children(func() { QuestionnaireWith(c, &state, p, nil) })
				}, width, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				title := questionnaireAccess(t, r, ui.RoleText, "Title").Bounds
				desc, _ := r.Find("Description")
				one := questionnaireAccess(t, r, ui.RoleRadio, "One")
				two := questionnaireAccess(t, r, ui.RoleRadio, "Two")
				input := questionnaireAccess(t, r, ui.RoleTextField, "Draft")
				near := func(got, want float32) {
					t.Helper()
					if math.Abs(float64(got-want)) > .1 {
						t.Fatalf("source geometry got%v want%v title%v desc%v one%v", got, want, title, desc, one.Bounds)
					}
				}
				near(title.H, 24)
				near(desc.Y, title.Y+title.H)
				near(desc.H, 20)
				near(one.Bounds.Y, desc.Y+desc.H+20)
				near(one.Bounds.H, 49.25)
				near(two.Bounds.Y, one.Bounds.Y+one.Bounds.H+12)
				want := float32(36)
				if width < 640 {
					want = 44
				}
				near(input.Bounds.H, want)
				if !dark {
					got := r.Image().RGBAAt(int(one.Bounds.X+8), int(one.Bounds.Y+24))
					if got != (color.RGBA{50, 90, 150, 255}) {
						t.Fatalf("ordinary choice lost source transparent surface: %v", got)
					}
				}
				r.Click("One")
				got := r.Image().RGBAAt(int(one.Bounds.X+8), int(one.Bounds.Y+24))
				expected := color.RGBA{tk.Muted.R, tk.Muted.G, tk.Muted.B, tk.Muted.A}
				if got != expected {
					t.Fatalf("checked native choice face got%v want%v", got, expected)
				}
				r.Advance(time.Second)
				if r.FramePending() {
					t.Fatal("reduced source UI requested idle frames")
				}
			})
		}
	}
}

func TestQuestionnaireSourceInactiveNativeAutofocusCannotCaptureKeyboard(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Active", Required: true, Freeform: true}, {ID: "b", Title: "Inactive", Required: true, Freeform: true}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		for n, q := range p.Questions {
			parts.Item(questionID(q, n), func(i *QuestionnaireItemParts) {
				i.Title(q.Title)
				i.Input(QuestionnaireInputProps{Label: q.Title + " input", Editor: func(c *ui.Context, v *string) ui.Element {
					e := control.Input(c, v, control.InputProps{Label: q.Title + " input"})
					if !i.Active {
						e.AutoFocus()
					}
					return e
				}})
			})
		}
		parts.Next()
	})
	if r.Focused("Inactive input") {
		t.Fatal("inactive registered custom editor autofocus bypassed native disabled/inert state")
	}
	r.Click("Active input")
	r.Type("Real active text")
	if state.Answers["b"].Text != "" || state.Answers["a"].Text != "Real active text" {
		t.Fatal("inactive hidden native input received text")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Inactive input"); ok {
		t.Fatal("inactive custom editor leaked into AX")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("inactive autofocus requested an idle frame loop")
	}
}
func TestQuestionnaireSourcePhysicalFormNavigationAndNativeRadioDirection(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			var state QuestionnaireState
			p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Choose", Required: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}}}, {ID: "b", Title: "Last", Required: true, Freeform: true}}}
			r := questionnaireSourceTester(&state, &p, nil)
			r.SetDirection(direction)
			r.Click("Alpha")
			r.Key(0, ui.KeyRight)
			target := "Beta"
			if direction == ui.RightToLeft {
				target = "Gamma"
			}
			if !r.Focused(target) || !questionnaireAccess(t, r, ui.RoleRadio, target).Checked {
				t.Fatal("radio physical horizontal direction disagrees with native RTL semantics")
			}
			r.KeyRepeat(0, ui.KeyRight)
			if r.Focused(target) {
				t.Fatal("native repeated radio arrow did not move")
			}
			r.Key(ui.Cmd, ui.KeyEnter)
			r.Key(0, ui.KeyLeft)
			if state.Index != 0 {
				t.Fatal("RTL changed physical form previous arrow")
			}
		})
	}
}

func TestQuestionnaireSourceProgressAnnouncesOnceAndErrorUsesNativeAlertRelations(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "First", Required: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}}}, {ID: "b", Title: "Last", Required: true, Freeform: true}}}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Next")
	questionnaireAccess(t, r, ui.RoleAlert, "Choose an answer to continue.")
	answer := questionnaireAccess(t, r, ui.RoleRadio, "Alpha")
	if !answer.Invalid || !slices.Contains(answer.ErrorMessage, "Choose an answer to continue.") {
		t.Fatal("invalid native answer lost alert relation")
	}
	r.Click("Alpha")
	r.Click("Next")
	if !slices.Contains(r.Announcements(), "Question 2 of 2") {
		t.Fatalf("changed progress not announced: %v", r.Announcements())
	}
	count := len(r.Announcements())
	r.Frame()
	r.Frame()
	if len(r.Announcements()) != count {
		t.Fatal("plain frames replayed live announcement")
	}
}

func TestQuestionnaireSourceNavigationLabelOptionsKeepOutlineAndExplicitVariantOverrides(t *testing.T) {
	var state QuestionnaireState
	variant := control.Ghost
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Optional", Freeform: true}}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(ui.RGB(50, 90, 150)).Children(func() {
			QuestionnaireWith(c, &state, p, func(parts *QuestionnaireParts) {
				parts.Item("a", func(i *QuestionnaireItemParts) { i.Title("Optional"); i.Input() })
				parts.Actions(func(*ui.Context) {
					parts.Skip(QuestionnaireActionProps{Label: "Omit"})
					parts.Submit(QuestionnaireActionProps{Label: "Send", Variant: &variant})
				})
			})
		})
	}, 560, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	skip := questionnaireAccess(t, r, ui.RoleButton, "Omit").Bounds
	got := r.Image().RGBAAt(int(skip.X+5), int(skip.Y+skip.H/2))
	want := theme.Light().Background
	if got != (color.RGBA{want.R, want.G, want.B, want.A}) {
		t.Fatalf("label option lost source outline variant: %v", got)
	}
	send := questionnaireAccess(t, r, ui.RoleButton, "Send").Bounds
	got = r.Image().RGBAAt(int(send.X+5), int(send.Y+send.H/2))
	if got != (color.RGBA{50, 90, 150, 255}) {
		t.Fatalf("explicit native ghost variant failed: %v", got)
	}
}
func TestQuestionnaireSourceCollectionsShortcutLimitsAndDisabledDefinitionChanges(t *testing.T) {
	for _, mode := range []QuestionnaireShortcutMode{QuestionnaireLetters, QuestionnaireNumbers} {
		t.Run(string(mode), func(t *testing.T) {
			var state QuestionnaireState
			choices := []control.Option{}
			limit := 26
			if mode == QuestionnaireNumbers {
				limit = 9
			}
			for n := 0; n < limit+1; n++ {
				choices = append(choices, control.Option{Value: fmt.Sprint(n), Label: fmt.Sprint("Choice ", n)})
			}
			p := QuestionnaireProps{Shortcuts: mode, DefaultItem: "a", Questions: []Question{{ID: "a", Title: "Choose", Choices: choices}, {ID: "b", Title: "Remaining", Freeform: true}}}
			r := ui.NewTesterWith(func(c *ui.Context) { theme.Apply(c); QuestionnaireWith(c, &state, p, nil) }, 500, 2400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if questionnaireAccess(t, r, ui.RoleRadio, fmt.Sprint("Choice ", limit)).KeyShortcuts != "" {
				t.Fatal("shortcut overflow was assigned")
			}
			target := questionnaireAccess(t, r, ui.RoleRadio, fmt.Sprint("Choice ", limit-1))
			if target.KeyShortcuts == "" {
				t.Fatal("last supported shortcut missing AX")
			}
			r.Click("Choice 0")
			r.KeyRepeat(0, questionnaireKey(questionnaireShortcut(mode, 1)))
			if !slices.Equal(state.Answers["a"].Values, []string{"0"}) {
				t.Fatal("repeated shortcut selected new answer")
			}
			r.Key(0, questionnaireKey(questionnaireShortcut(mode, 1)))
			if !slices.Equal(state.Answers["a"].Values, []string{"1"}) {
				t.Fatal("last managed keyboard action did not select")
			}
			p.Questions[0].Disabled = true
			r.Frame()
			if state.Index != 1 || questionnaireAccess(t, r, ui.RoleProgress, "Questionnaire progress").Max != 1 {
				t.Fatal("disabling active definition did not reconcile enabled collection")
			}
		})
	}
}
func TestQuestionnaireSourceLatestDefaultItemUsedByActualReset(t *testing.T) {
	var state QuestionnaireState
	initial := "a"
	qs := []Question{{ID: "a", Title: "A", Choices: []control.Option{{Value: "yes", Label: "Yes A"}}}, {ID: "b", Title: "B", Choices: []control.Option{{Value: "yes", Label: "Yes B"}}}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs, DefaultItem: initial}, func(p *QuestionnaireParts) {
			p.Item("a", func(i *QuestionnaireItemParts) { i.Choice("yes") })
			p.Item("b", func(i *QuestionnaireItemParts) { i.Choice("yes") })
			p.Reset()
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	initial = "b"
	r.Frame()
	if err := r.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	if state.Index != 1 {
		t.Fatalf("Reset ignored latest defaultItem b: index=%d", state.Index)
	}
}

func TestQuestionnaireSourceChoicesWithoutOptionalDefinitionMetadataRemainSuccessful(t *testing.T) {
	for _, multi := range []bool{false, true} {
		t.Run(map[bool]string{false: "radio", true: "checkbox"}[multi], func(t *testing.T) {
			var state QuestionnaireState
			submitted := 0
			qs := []Question{{ID: "q", Title: "Runtime-only choices", Required: true, Multiple: multi}}
			r := ui.NewTesterWith(func(c *ui.Context) {
				QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs, OnSubmit: func(map[string]Answer) { submitted++ }}, func(p *QuestionnaireParts) {
					p.Item("q", func(i *QuestionnaireItemParts) {
						i.Choices(func(c *ui.Context) { i.Choice("runtime", QuestionnaireChoiceProps{Label: "Runtime choice"}) })
						i.Error()
					})
					p.Submit()
				})
			}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if err := r.Click("Runtime choice"); err != nil {
				t.Fatal(err)
			}
			if state.Status(qs[0], 0) != QuestionnaireAnswered || !slices.Equal(state.FormValues(qs)["q"], []string{"runtime"}) {
				t.Errorf("actual native choice excluded by absent optional definition metadata: status=%s form=%v", state.Status(qs[0], 0), state.FormValues(qs))
			}
			if err := r.Click("Submit"); err != nil {
				t.Fatal(err)
			}
			if submitted != 1 {
				t.Fatalf("actual chosen runtime answer did not submit, errors=%v", state.Errors)
			}
		})
	}
}

func TestQuestionnaireSourceRuntimeOnlyRadioGroupHasOneNativeTabStop(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Runtime", Required: true}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Item("a", func(i *QuestionnaireItemParts) {
			i.Choice("a", QuestionnaireChoiceProps{Label: "Alpha"})
			i.Choice("b", QuestionnaireChoiceProps{Label: "Beta"})
		})
		parts.Submit()
	})
	r.Click("Beta")
	r.Key(0, ui.KeyTab)
	if !r.Focused("Submit") {
		t.Fatal("selected runtime radio did not leave group on Tab")
	}
	r.Key(ui.Shift, ui.KeyTab)
	if !r.Focused("Beta") {
		t.Fatal("native group did not return to selected runtime radio")
	}
	r.Key(ui.Shift, ui.KeyTab)
	if r.Focused("Alpha") {
		t.Fatal("runtime-only group retained a second radio Tab stop")
	}
}
func TestQuestionnaireSourceControlledResetDoesNotRequestUnchangedItem(t *testing.T) {
	var state QuestionnaireState
	active := "a"
	calls := 0
	qs := []Question{{ID: "a", Title: "A", Choices: []control.Option{{Value: "yes", Label: "Yes"}}}, {ID: "b", Title: "B"}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs, Item: &active, DefaultItem: "a", OnItemChange: func(string) { calls++ }}, func(p *QuestionnaireParts) {
			p.Item("a", func(i *QuestionnaireItemParts) { i.Choice("yes") })
			p.Item("b", func(i *QuestionnaireItemParts) {})
			p.Reset()
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("controlled reset requested already active item %d times", calls)
	}
}

func TestQuestionnaireSourceRuntimeOnlyChoicesHaveNativeTabEntry(t *testing.T) {
	var state QuestionnaireState
	qs := []Question{{ID: "q", Title: "Runtime"}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs}, func(p *QuestionnaireParts) {
			p.Item("q", func(i *QuestionnaireItemParts) {
				i.Choice("one", QuestionnaireChoiceProps{Label: "Runtime one"})
				i.Choice("two", QuestionnaireChoiceProps{Label: "Runtime two"})
			})
			p.Submit()
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Key(0, ui.KeyTab)
	if !r.Focused("Runtime one") {
		t.Fatalf("no native keyboard entry to runtime choices; access=%+v", r.AccessAll())
	}
}

func TestQuestionnaireSourceUnmountedChoiceDoesNotReappearSelected(t *testing.T) {
	var state QuestionnaireState
	show := true
	qs := []Question{{ID: "q", Title: "Choice", Required: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs}, func(p *QuestionnaireParts) {
			p.Item("q", func(i *QuestionnaireItemParts) {
				i.Choice("a")
				if show {
					i.Choice("b")
				}
			})
			p.Submit()
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Beta"); err != nil {
		t.Fatal(err)
	}
	show = false
	r.Frame()
	if state.Status(qs[0], 0) != QuestionnaireUnanswered {
		t.Fatal("removed radio still successful")
	}
	show = true
	r.Frame()
	if state.Status(qs[0], 0) != QuestionnaireUnanswered {
		t.Fatalf("unmounted/remounted uncontrolled radio came back selected: %+v", state.Answers)
	}
}
func TestQuestionnaireSourceUncontrolledPreventedInputRetainsNativeDraftAndPriorSelection(t *testing.T) {
	for _, initial := range []string{"", "Initial"} {
		t.Run(map[bool]string{false: "previously-empty", true: "previously-selected"}[initial != ""], func(t *testing.T) {
			var state QuestionnaireState
			qs := []Question{{ID: "q", Title: "Custom text", Required: true, Freeform: true}}
			r := ui.NewTesterWith(func(c *ui.Context) {
				QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs}, func(p *QuestionnaireParts) {
					p.Item("q", func(i *QuestionnaireItemParts) {
						i.Input(QuestionnaireInputProps{Label: "Draft", DefaultValue: initial, OnInput: func(e *QuestionnaireInputEvent) { e.PreventDefault() }})
					})
					p.Submit()
				})
			}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			if err := r.Click("Draft"); err != nil {
				t.Fatal(err)
			}
			r.Key(ui.Cmd, ui.KeyA)
			r.Type("Edited")
			field, _ := r.AccessByRole(ui.RoleTextField, "Draft")
			if field.Value != "Edited" {
				t.Errorf("prevented uncontrolled change rolled back native DOM-equivalent value: %q", field.Value)
			}
			want := QuestionnaireUnanswered
			var form []string
			if initial != "" {
				want = QuestionnaireAnswered
				form = []string{"Edited"}
			}
			if state.Status(qs[0], 0) != want || !slices.Equal(state.FormValues(qs)["q"], form) {
				t.Fatalf("prevented uncontrolled edit changed prior selection or submitted stale text: status=%s form=%v wantstatus%s form%v", state.Status(qs[0], 0), state.FormValues(qs), want, form)
			}
		})
	}
}

func TestQuestionnaireSourceControlledNavigationDuringIMEReleasesCompositionOnNewItem(t *testing.T) {
	var state QuestionnaireState
	active := "a"
	submitted := 0
	qs := []Question{{ID: "a", Title: "First", Required: true, Freeform: true, InputLabel: "First draft"}, {ID: "b", Title: "Second", Required: true, Choices: []control.Option{{Value: "yes", Label: "Yes"}}, DefaultAnswer: Answer{Values: []string{"yes"}}}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs, Item: &active, OnSubmit: func(map[string]Answer) { submitted++ }}, nil)
	}, 500, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("First draft"); err != nil {
		t.Fatal(err)
	}
	r.Type("Saved draft")
	r.Compose("未", 1)
	active = "b"
	r.Frame()
	if !r.Focused("Second") {
		t.Fatal("controlled new fieldset did not get real native focus")
	}
	if state.runtime["a"].controls["\x00input"].composing {
		t.Fatal("probe still has native composition instead of stale questionnaire flag")
	}
	r.Key(ui.Cmd, ui.KeyEnter)
	if submitted != 1 {
		t.Fatalf("ended native IME left new item confirmation blocked: submits=%d composing=%v/%v", submitted, state.runtime["a"].composing, state.runtime["b"].composing)
	}
}

func TestQuestionnaireSourceCustomNativeEditorHonorsReadOnlyProp(t *testing.T) {
	var state QuestionnaireState
	qs := []Question{{ID: "q", Title: "Readonly draft", Freeform: true}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs}, func(p *QuestionnaireParts) {
			p.Item("q", func(i *QuestionnaireItemParts) {
				i.Input(QuestionnaireInputProps{Label: "Readonly custom", DefaultValue: "Saved", ReadOnly: true, Editor: func(c *ui.Context, value *string) ui.Element { return ui.TextInput(c, value) }})
			})
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	field, _ := r.AccessByRole(ui.RoleTextField, "Readonly custom")
	if !field.ReadOnly {
		t.Error("ReadOnly prop was not applied to actual native editor")
	}
	if err := r.Click("Readonly custom"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Changed")
	if state.Answers["q"].Text != "Saved" {
		t.Fatalf("custom native read-only editor accepted edit %q", state.Answers["q"].Text)
	}
}

func TestQuestionnaireSourceAllDisabledResetDoesNotRequestDisabledItem(t *testing.T) {
	var state QuestionnaireState
	active := "a"
	calls := 0
	qs := []Question{{ID: "a", Title: "Disabled", Disabled: true}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs, Item: &active, OnItemChange: func(string) { calls++ }}, func(p *QuestionnaireParts) { p.Item("a", func(i *QuestionnaireItemParts) {}); p.Reset() })
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("empty enabled collection reset requested disabled item %d times", calls)
	}
}
func TestQuestionnaireSourceEmptyStringChoiceValueIsARealAnswer(t *testing.T) {
	var state QuestionnaireState
	qs := []Question{{ID: "q", Title: "Preference", Required: true, Choices: []control.Option{{Value: "", Label: "No preference"}, {Value: "a", Label: "Alpha"}}}}
	submitted := 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs, OnSubmit: func(map[string]Answer) { submitted++ }}, nil)
	}, 500, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("No preference"); err != nil {
		t.Fatal(err)
	}
	if state.Status(qs[0], 0) != QuestionnaireAnswered || !slices.Equal(state.FormValues(qs)["q"], []string{""}) {
		t.Fatalf("empty string native radio value discarded: status=%s form=%v", state.Status(qs[0], 0), state.FormValues(qs))
	}
	if err := r.Click("Submit"); err != nil {
		t.Fatal(err)
	}
	if submitted != 1 {
		t.Fatalf("empty string selected choice did not submit: %v", state.Errors)
	}
}

func TestQuestionnaireSourceDisabledNavigationDoesNotAdvertiseEnterShortcut(t *testing.T) {
	var state QuestionnaireState
	qs := []Question{{ID: "q", Title: "Choice", Choices: []control.Option{{Value: "yes", Label: "Yes"}}}}
	r := ui.NewTesterWith(func(c *ui.Context) {
		QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs}, func(p *QuestionnaireParts) {
			p.Item("q", func(i *QuestionnaireItemParts) { i.Choice("yes") })
			p.Submit(QuestionnaireActionProps{Disabled: true})
		})
	}, 500, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	button, _ := r.AccessByRole(ui.RoleButton, "Submit")
	if !button.Disabled {
		t.Fatal("probe did not render disabled native action")
	}
	if button.KeyShortcuts != "" {
		t.Fatalf("disabled action advertises non-active source shortcut %q", button.KeyShortcuts)
	}
}

func TestQuestionnaireSourceMultipleToggleDoesNotReactivateInactiveSingleDraft(t *testing.T) {
	var state QuestionnaireState
	qs := []Question{{ID: "q", Title: "Answer mode", Required: true, Freeform: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}}}}
	r := ui.NewTesterWith(func(c *ui.Context) { QuestionnaireWith(c, &state, QuestionnaireProps{Questions: qs}, nil) }, 500, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Your answer"); err != nil {
		t.Fatal(err)
	}
	r.Type("Retained draft")
	if err := r.Click("Alpha"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(state.FormValues(qs)["q"], []string{"a"}) {
		t.Fatal("probe did not make input draft inactive")
	}
	qs[0].Multiple = true
	r.Frame()
	if !slices.Equal(state.FormValues(qs)["q"], []string{"a"}) {
		t.Fatalf("changing multiple reactivated an unselected freeform draft: %v", state.FormValues(qs))
	}
}

func TestQuestionnaireSourcePreventedClearingRetainsSelectionButEnterRequiresFilledNativeValue(t *testing.T) {
	var state QuestionnaireState
	calls := 0
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Answer", Required: true, Freeform: true}}, OnSubmit: func(map[string]Answer) { calls++ }}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Item("a", func(i *QuestionnaireItemParts) {
			i.Input(QuestionnaireInputProps{Label: "Draft", DefaultValue: "Selected", OnInput: func(e *QuestionnaireInputEvent) { e.PreventDefault() }})
		})
		parts.Submit()
	})
	r.Click("Draft")
	r.Key(ui.Cmd, ui.KeyA)
	r.Key(0, ui.KeyBackspace)
	if state.Status(p.Questions[0], 0) != QuestionnaireAnswered || !slices.Equal(state.FormValues(p.Questions)["a"], []string{""}) {
		t.Fatalf("prevented native clearing lost prior selection or successful empty value: %v", state.FormValues(p.Questions))
	}
	r.Key(0, ui.KeyEnter)
	if calls != 0 {
		t.Fatal("plain Enter confirmed native empty value")
	}
	r.Key(ui.Cmd, ui.KeyEnter)
	if calls != 1 {
		t.Fatal("CmdEnter failed strict selected-state validation")
	}
}
func TestQuestionnaireSourceIMEBlurAndReturnDoesNotLeaveStaleComposition(t *testing.T) {
	var state QuestionnaireState
	outside := ""
	calls := 0
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Answer", Required: true, Freeform: true, InputLabel: "Draft"}}, OnSubmit: func(map[string]Answer) { calls++ }}
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Children(func() {
			QuestionnaireWith(c, &state, p, nil)
			control.Input(c, &outside, control.InputProps{Label: "Outside"})
		})
	}, 560, 700)
	r.Click("Draft")
	r.Type("Saved")
	r.Compose("未", 1)
	r.Click("Outside")
	r.Click("Draft")
	r.Key(ui.Cmd, ui.KeyEnter)
	if calls != 1 {
		t.Fatal("blurred composition blocked returning native editor confirmation")
	}
}

func TestQuestionnaireSourceUncontrolledCollectionFallbackAndResetReportActualNames(t *testing.T) {
	var state QuestionnaireState
	calls := []string{}
	p := QuestionnaireProps{DefaultItem: "a", Questions: []Question{{ID: "a", Title: "First", Choices: []control.Option{{Value: "yes", Label: "Yes"}}}, {ID: "b", Title: "Second", Freeform: true}}, OnItemChange: func(name string) { calls = append(calls, name) }}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		for n, q := range p.Questions {
			parts.Item(questionID(q, n), func(i *QuestionnaireItemParts) {
				i.Title(q.Title)
				for _, o := range q.Choices {
					i.Choice(o.Value)
				}
				if q.Freeform {
					i.Input()
				}
			})
		}
		parts.Actions(func(*ui.Context) { parts.Reset(); parts.Next() })
	})
	p.Questions[0].Disabled = true
	r.Frame()
	if !slices.Equal(calls, []string{"b"}) {
		t.Fatalf("enabled collection fallback was not reported once: %v", calls)
	}
	p.Questions[0].Disabled = false
	r.Frame()
	r.Click("Reset")
	if !slices.Equal(calls, []string{"b", "a"}) {
		t.Fatalf("uncontrolled reset did not report changed item: %v", calls)
	}
}

func TestQuestionnaireSourceCustomMultilineEditorKeepsEmptyVerticalCaret(t *testing.T) {
	var state QuestionnaireState
	p := QuestionnaireProps{Questions: []Question{{ID: "a", Title: "Custom editor", Required: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}}}}}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Item("a", func(i *QuestionnaireItemParts) {
			i.Choice("a")
			i.Input(QuestionnaireInputProps{Label: "Multiline", Editor: func(c *ui.Context, v *string) ui.Element {
				return control.Textarea(c, v, control.TextareaProps{Label: "Multiline", Rows: 3})
			}})
		})
		parts.Submit()
	})
	r.Click("Multiline")
	r.Type("\n\n")
	r.Key(0, ui.KeyUp)
	if !r.Focused("Multiline") {
		t.Fatal("form navigation hijacked empty multiline editor caret")
	}
	r.Type("X")
	if state.Answers["a"].Text != "\nX\n" {
		t.Fatalf("custom native caret did not move on its own line: %q", state.Answers["a"].Text)
	}
}

func TestQuestionnaireLegacyResultKeepsFreeformAlongsideFixedChoice(t *testing.T) {
	q := []Question{{ID: "legacy", Freeform: true, Choices: []control.Option{{Value: "fixed", Label: "Fixed"}}}}
	state := QuestionnaireState{Answers: map[string]Answer{"legacy": {Text: "Legacy freeform"}}}
	state.Select(q[0], 0, "fixed")
	result := state.Result(q)
	if result["legacy"].Text != "Legacy freeform" || !slices.Equal(result["legacy"].Values, []string{"fixed"}) {
		t.Fatalf("legacy automatic answer inference changed: %+v", result)
	}
	result["legacy"].Values[0] = "mutated"
	if state.Answers["legacy"].Values[0] != "fixed" {
		t.Fatal("legacy result aliased model")
	}
}

func TestQuestionnaireSourceRemovingQuestionDisposesItsUncontrolledSelection(t *testing.T) {
	var state QuestionnaireState
	question := Question{ID: "a", Title: "Removed", Required: true, Freeform: true, Choices: []control.Option{{Value: "a", Label: "Alpha"}}}
	p := QuestionnaireProps{Questions: []Question{question, {ID: "b", Title: "Remaining", Freeform: true}}}
	r := questionnaireSourceTester(&state, &p, nil)
	r.Click("Alpha")
	p.Questions = p.Questions[1:]
	r.Frame()
	if _, exists := state.Answers["a"]; exists {
		t.Fatal("removed source item retained its uncontrolled selection")
	}
	p.Questions = append([]Question{question}, p.Questions...)
	state.JumpTo(p.Questions, "a")
	r.Frame()
	if questionnaireAccess(t, r, ui.RoleRadio, "Alpha").Checked || state.Status(question, 0) != QuestionnaireUnanswered {
		t.Fatal("reinserted item revived removed selection instead of defaults")
	}
}
func TestQuestionnaireSourceNilChoiceMetadataUsesRegisteredShortcutOrder(t *testing.T) {
	for _, mode := range []QuestionnaireShortcutMode{QuestionnaireLetters, QuestionnaireNumbers} {
		t.Run(string(mode), func(t *testing.T) {
			var state QuestionnaireState
			p := QuestionnaireProps{Shortcuts: mode, Questions: []Question{{ID: "a", Title: "Runtime", Required: true}}}
			r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
				parts.Item("a", func(i *QuestionnaireItemParts) {
					i.Choice("blocked", QuestionnaireChoiceProps{Label: "Blocked", Disabled: true})
					i.Choice("b", QuestionnaireChoiceProps{Label: "Beta"})
					i.Input(QuestionnaireInputProps{Label: "Other"})
					i.Choice("a", QuestionnaireChoiceProps{Label: "Alpha"})
				})
				parts.Submit()
			})
			first := questionnaireAccess(t, r, ui.RoleRadio, "Beta")
			if first.KeyShortcuts == "" {
				t.Fatal("runtime native first answer has no assigned shortcut")
			}
			r.Click("Beta")
			r.Key(0, questionnaireKey(questionnaireShortcut(mode, 1)))
			if !r.Focused("Alpha") || !slices.Equal(state.FormValues(p.Questions)["a"], []string{"a"}) {
				t.Fatal("runtime shortcut order failed to skip disabled/native input slots")
			}
			p.Questions[0].Choices = []control.Option{}
			r.Frame()
			if questionnaireAccess(t, r, ui.RoleRadio, "Beta").KeyShortcuts != "" {
				t.Fatal("explicit empty metadata did not suppress runtime shortcut assignment")
			}
		})
	}
}
