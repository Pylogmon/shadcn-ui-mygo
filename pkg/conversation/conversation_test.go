package conversation

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func click(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	tt.Frame()
}

func TestQuestionnaireValidationProgressAndSubmission(t *testing.T) {
	questions := []Question{
		{ID: "kind", Title: "Choose a project", Required: true, Choices: []control.Option{{Value: "native", Label: "Native app"}, {Value: "web", Label: "Website"}}},
		{ID: "features", Title: "Choose features", Multiple: true, Required: true, Choices: []control.Option{{Value: "themes", Label: "Themes"}, {Value: "motion", Label: "Motion"}, {Value: "disabled", Label: "Unavailable", Disabled: true}}},
		{ID: "note", Title: "Anything else?", Freeform: true},
	}
	s := QuestionnaireState{}
	submissions := 0
	var result map[string]Answer
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Questionnaire(c, &s, QuestionnaireProps{Questions: questions, OnSubmit: func(answers map[string]Answer) { submissions++; result = answers }}).Width(420)
	}, 460, 500)
	click(t, tt, "Next")
	if s.Index != 0 || s.Errors["kind"] == "" || !tt.HasText("Please answer this question.") {
		t.Fatal("required validation did not block navigation")
	}
	click(t, tt, "Native app")
	if !slices.Equal(s.Answers["kind"].Values, []string{"native"}) {
		t.Fatalf("single choice not saved: %+v", s.Answers)
	}
	click(t, tt, "Next")
	if s.Index != 1 {
		t.Fatalf("index=%d", s.Index)
	}
	click(t, tt, "Themes")
	click(t, tt, "Motion")
	if !slices.Equal(s.Answers["features"].Values, []string{"themes", "motion"}) {
		t.Fatalf("multiple choices lost: %+v", s.Answers)
	}
	click(t, tt, "Unavailable")
	if slices.Contains(s.Answers["features"].Values, "disabled") {
		t.Fatal("disabled option selected")
	}
	click(t, tt, "Next")
	click(t, tt, "Previous")
	if s.Index != 1 || len(s.Answers["features"].Values) != 2 {
		t.Fatal("back navigation lost answers")
	}
	click(t, tt, "Next")
	click(t, tt, "Skip")
	if !s.Submitted || submissions != 1 || !result["note"].Skipped {
		t.Fatalf("submit failed: %+v submissions=%d result=%+v", s, submissions, result)
	}
	tt.Frame()
	tt.Frame()
	if submissions != 1 {
		t.Fatal("onSubmit repeated without input")
	}
	result["features"].Values[0] = "changed"
	if s.Answers["features"].Values[0] != "themes" {
		t.Fatal("result aliases mutable state")
	}
}

func TestQuestionnaireFreeformAndStaleChoices(t *testing.T) {
	s := QuestionnaireState{Answers: map[string]Answer{"one": {Values: []string{"missing", "disabled"}, Text: "stale"}}}
	q := []Question{{ID: "one", Title: "Answer", Required: true, Freeform: false, Choices: []control.Option{{Value: "disabled", Disabled: true}}}}
	if s.Submit(q) {
		t.Fatal("unknown/disabled values or hidden text satisfied required validation")
	}
	q[0].Freeform = true
	if !s.Submit(q) {
		t.Fatal("freeform answer did not satisfy required validation")
	}
	if len(s.Result(q)["one"].Values) != 0 {
		t.Fatal("invalid values escaped in result")
	}
	if s.Skip(q) {
		t.Fatal("required question can be skipped")
	}
	var state QuestionnaireState
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		Questionnaire(c, &state, QuestionnaireProps{Questions: []Question{{ID: "text", Title: "Describe it", Required: true, Freeform: true, InputLabel: "Description"}}})
	}, 400, 360)
	field, ok := tt.AccessByRole(ui.RoleTextField, "Description")
	if !ok {
		t.Fatal("freeform field missing")
	}
	tt.ClickAt(field.Bounds.X+8, field.Bounds.Y+field.Bounds.H/2)
	tt.Type("A native application")
	click(t, tt, "Submit")
	if !state.Submitted || state.Answers["text"].Text != "A native application" {
		t.Fatalf("freeform typing failed: %+v", state)
	}
}

func TestTraceAndAttachmentActions(t *testing.T) {
	s := TraceState{}
	opened, removed := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(12).Gap(10).Children(func() {
			Trace(c, &s, []TraceStep{{ID: "build", Title: "Build application", Status: Complete, Detail: "Build succeeded"}})
			Attachment(c, AttachmentProps{Name: "report.pdf", SizeBytes: 2048, OnOpen: func() { opened++ }, OnRemove: func() { removed++ }})
		})
	}, 440, 380)
	if tt.HasText("Build succeeded") {
		t.Fatal("trace detail initially visible")
	}
	click(t, tt, "Build application")
	if !s.Expanded["build"] || !tt.HasText("Build succeeded") {
		t.Fatal("trace did not expand")
	}
	if node, ok := tt.AccessByRole(ui.RoleButton, "Build application"); !ok || !node.Expanded {
		t.Fatal("trace disclosure accessibility missing")
	}
	click(t, tt, "Build application")
	if s.Expanded["build"] || tt.HasText("Build succeeded") {
		t.Fatal("trace did not collapse")
	}
	click(t, tt, "Open")
	click(t, tt, "Remove report.pdf")
	if opened != 1 || removed != 1 {
		t.Fatalf("attachment actions: open=%d remove=%d", opened, removed)
	}
}

func TestMessageScrollerVirtualizesFollowsAndPreservesKeyedPlace(t *testing.T) {
	var state ui.ListState
	count, base := 10000, 0
	tt := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		MessageScroller(c, &state, count, MessageScrollerProps{FollowEnd: true, Key: func(i int) any { return i + base }}, func(c *ui.Context, i int) {
			ui.Text(c, fmt.Sprintf("Message %d", i+base)).Height(28)
		}).Fill()
	}, 420, 240, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	first, last := state.Visible()
	if last < 9998 || last-first > 30 || !state.AtEnd() {
		t.Fatalf("initial visible range %d..%d atEnd=%v", first, last, state.AtEnd())
	}
	state.ScrollTo(5000, ui.Start)
	tt.Frame()
	first, last = state.Visible()
	if first > 5000 || last < 5000 || last-first > 30 {
		t.Fatalf("jump range %d..%d", first, last)
	}
	before := first + base
	count += 100
	base -= 100
	tt.Frame()
	first, _ = state.Visible()
	if first+base != before {
		t.Fatalf("prepending history moved visible key %d to %d", before, first+base)
	}
	count++
	tt.Frame()
	first, _ = state.Visible()
	if first+base != before {
		t.Fatal("new message pulled reader away from history")
	}
	state.ScrollToEnd()
	tt.Frame()
	if !state.AtEnd() {
		t.Fatal("jump to latest did not restore live edge")
	}
}
