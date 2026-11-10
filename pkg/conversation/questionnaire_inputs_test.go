package conversation

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
)

// One native flow follows the same item from single alternatives to multiple
// successful same-name controls. The first unnamed Input keeps legacy Text.
func TestQuestionnaireIndependentNativeInputAnswers(t *testing.T) {
	var state QuestionnaireState
	var submitted map[string]Answer
	keyed, reverse, secondDisabled, showSecond, duplicate := false, false, false, true, false
	choiceLast, validateSecond, slotDefaults := false, false, true
	generation, changedGeneration := 1, 0
	p := QuestionnaireProps{Questions: []Question{{ID: "context", Title: "Execution context", Required: true, Choices: []control.Option{{Value: "fixed", Label: "Workspace context"}}}}, OnSubmit: func(a map[string]Answer) { submitted = a }}
	r := questionnaireSourceTester(&state, &p, func(parts *QuestionnaireParts) {
		parts.Item("context", func(item *QuestionnaireItemParts) {
			item.Title("Execution context")
			item.Choices(func(*ui.Context) {
				if !choiceLast {
					item.Choice("fixed")
				}
				current := generation
				first := func() {
					props := QuestionnaireInputProps{Label: "First detail", OnChange: func(string) { changedGeneration = current }}
					if validateSecond {
						props.Validate = func(string) string { return "" }
					}
					if keyed {
						props.Key = "location"
						if slotDefaults {
							props.DefaultValue = "initial staging"
						}
					}
					item.Input(props)
				}
				second := func() {
					if showSecond {
						props := QuestionnaireInputProps{Label: "Second detail", Disabled: secondDisabled}
						if validateSecond {
							props.Validate = func(string) string { return "Review policy needs approval" }
						}
						if keyed {
							props.Key = "policy"
							if slotDefaults {
								props.DefaultValue = "initial review"
							}
							if duplicate {
								props.Key = "location"
							}
						}
						item.Input(props)
					}
				}
				if reverse {
					second()
					first()
				} else {
					first()
					second()
				}
				if choiceLast {
					item.Choice("fixed")
				}
			})
			item.Error()
		})
		parts.Actions(func(*ui.Context) {
			parts.Reset(QuestionnaireActionProps{Label: "Reset answers"})
			parts.Skip(QuestionnaireActionProps{Label: "Skip context"})
			parts.Submit()
		})
	})
	edit := func(label, value string) {
		t.Helper()
		if err := r.Click(label); err != nil {
			t.Fatal(err)
		}
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(value)
	}
	values := func(want ...string) {
		t.Helper()
		if got := state.FormValues(p.Questions)["context"]; !slices.Equal(got, want) {
			t.Fatalf("successful native controls = %q, want %q", got, want)
		}
	}
	edit("First detail", "staging")
	if got := questionnaireAccess(t, r, ui.RoleTextField, "Second detail").Value; got != "" {
		t.Fatalf("editing the first input changed the independent second draft: %q", got)
	}
	values("staging")
	edit("Second detail", "review required")
	if got := questionnaireAccess(t, r, ui.RoleTextField, "First detail").Value; got != "staging" || state.Answers["context"].Text != "staging" {
		t.Fatalf("second input replaced the legacy first draft: %q / %+v", got, state.Answers["context"])
	}
	values("review required")
	if err := r.Click("Workspace context"); err != nil {
		t.Fatal(err)
	}
	values("fixed")
	edit("First detail", "updated staging")
	values("updated staging")
	if questionnaireAccess(t, r, ui.RoleRadio, "Workspace context").Checked {
		t.Fatal("single input did not deselect the fixed answer")
	}

	p.Questions[0].Multiple = true
	r.Frame()
	if err := r.Click("Workspace context"); err != nil {
		t.Fatal(err)
	}
	edit("Second detail", "updated staging")
	values("fixed", "updated staging", "updated staging")
	if err := r.Click("Submit"); err != nil {
		t.Fatal(err)
	}
	if submitted == nil || !slices.Equal(submitted["context"].Values, []string{"fixed"}) || submitted["context"].Text != "updated staging" {
		t.Fatal("current submission lost fixed values or the compatible default Text")
	}
	if len(submitted["context"].Inputs) != 2 || submitted["context"].Inputs[0].Text != submitted["context"].Inputs[1].Text {
		t.Fatal("same-text native inputs were collapsed in the submitted answer")
	}
	choiceLast = true
	r.Frame()
	edit("First detail", "fixed")
	edit("Second detail", "different detail")
	if got := state.FormValues(p.Questions)["context"]; !slices.Equal(got, []string{"fixed", "different detail", "fixed"}) {
		t.Errorf("input text suppressed a later successful choice with the same value: %q", got)
	}
	choiceLast = false
	r.Frame()

	// Named slots keep their native model when they move. Disabling excludes a
	// successful value while retaining the draft; removal unregisters its draft.
	keyed = true
	state.Answers["context"] = Answer{}
	r.Frame()
	values("initial staging", "initial review")
	validateSecond = true
	submitted = nil
	r.Frame()
	if err := r.Click("First detail"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Ctrl, ui.KeyEnter)
	if submitted != nil || state.Errors["context"] != "Review policy needs approval" || !r.Focused("Second detail") {
		var focused []string
		for _, access := range r.AccessAll() {
			if access.Focused {
				focused = append(focused, access.Name)
			}
		}
		t.Errorf("validation did not focus the actual invalid second editor: submitted=%t, error=%q, focused=%q", submitted != nil, state.Errors["context"], focused)
	}
	validateSecond = false
	r.Frame()
	edit("First detail", "rehearsal")
	edit("Second detail", "approval required")
	reverse = true
	r.Frame()
	values("approval required", "rehearsal")
	if !r.Focused("Second detail") || questionnaireAccess(t, r, ui.RoleTextField, "First detail").Value != "rehearsal" {
		t.Fatal("keyed native input reordering replaced a draft or focused editor")
	}
	secondDisabled = true
	r.Frame()
	values("rehearsal")
	retained := slices.ContainsFunc(state.Answers["context"].Inputs, func(input QuestionnaireInputValue) bool {
		return input.Key == "policy" && input.Text == "approval required"
	})
	if !questionnaireAccess(t, r, ui.RoleTextField, "Second detail").Disabled || !retained {
		t.Fatal("disabling a slot discarded its draft")
	}
	secondDisabled = false
	r.Frame()
	values("approval required", "rehearsal")
	if questionnaireAccess(t, r, ui.RoleTextField, "Second detail").Value != "approval required" {
		t.Fatal("reenabling the native editor lost its retained draft")
	}
	showSecond = false
	r.Frame()
	values("rehearsal")
	for _, input := range state.Answers["context"].Inputs {
		if input.Key == "policy" {
			t.Fatal("removed input retained its answer registration")
		}
	}
	showSecond = true
	r.Frame()
	values("initial review", "rehearsal")
	generation = 2
	r.Frame()
	edit("First detail", "current staging")
	if changedGeneration != 2 {
		t.Fatal("input used a callback from an earlier build")
	}
	if err := r.Click("Submit"); err != nil {
		t.Fatal(err)
	}
	a := submitted["context"]
	if len(a.Inputs) != 2 || a.Inputs[0].Key != "policy" || a.Inputs[1].Key != "location" || a.Text != "" {
		t.Fatalf("keyed submission identities or legacy Text are wrong: %+v", a)
	}
	a.Inputs[0].Text = "mutated copy"
	values("initial review", "current staging")

	// A single item selects its first compatible answer on mode change. Skip
	// keeps text but excludes every answer until a new accepted interaction.
	p.Questions[0].Multiple = false
	p.Questions[0].Required = false
	r.Frame()
	values("initial review")
	if err := r.Click("Skip context"); err != nil {
		t.Fatal(err)
	}
	values()
	if questionnaireAccess(t, r, ui.RoleTextField, "First detail").Value != "current staging" || questionnaireAccess(t, r, ui.RoleTextField, "Second detail").Value != "initial review" {
		t.Fatal("skip lost independent drafts")
	}
	edit("First detail", "accepted staging")
	values("accepted staging")
	if err := r.Click("Reset answers"); err != nil {
		t.Fatal(err)
	}
	values("initial review")
	if questionnaireAccess(t, r, ui.RoleTextField, "First detail").Value != "initial staging" {
		t.Fatal("reset did not restore each input's own default")
	}
	slotDefaults = false
	p.Questions[0].Multiple = true
	p.Questions[0].DefaultAnswer = Answer{Inputs: []QuestionnaireInputValue{{Key: "location", Text: "unselected saved draft"}, {Key: "policy", Text: "accepted saved review", Selected: true}}, Mode: QuestionnaireInputAnswer}
	state.Answers["context"] = Answer{Inputs: slices.Clone(p.Questions[0].DefaultAnswer.Inputs), Mode: QuestionnaireInputAnswer}
	r.Frame()
	values("accepted saved review")
	if err := r.Click("Reset answers"); err != nil {
		t.Fatal(err)
	}
	if got := state.FormValues(p.Questions)["context"]; !slices.Equal(got, []string{"accepted saved review"}) {
		t.Errorf("reset promoted an explicitly unselected saved input draft: %q", got)
	}
	if questionnaireAccess(t, r, ui.RoleTextField, "First detail").Value != "unselected saved draft" || p.Questions[0].DefaultAnswer.Inputs[0].Selected {
		t.Fatal("restoring defaults lost the saved draft or mutated caller defaults")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("duplicate explicit Input keys were accepted")
			}
		}()
		duplicate = true
		r.Frame()
	}()
}
