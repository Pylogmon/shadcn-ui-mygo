package gallery

import (
	"maps"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
)

func TestCreatorQuestionnaireSavedCheckpointSearchAndPageReturn(t *testing.T) {
	s := New()
	s.Config.ReduceMotion = true
	r := ui.NewTester(s.View, 500, 900)
	r.SetDirection(ui.RightToLeft)
	r.Key(ui.Cmd, ui.KeyK)
	r.Frame()
	r.Type("Questionnaire")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if s.Page != "conversation" || s.Command || !r.HasText("Set up your workspace") || !r.HasText("Questionnaire workflows") {
		t.Fatal("Questionnaire search did not reach both the original and workflow examples")
	}
	workspaceAnswers := maps.Clone(s.Questionnaire.Answers)
	a := revealPreviewPart(t, r, "Messages & questions", ui.RoleComboBox, "Questionnaire configuration")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	p := &s.questionnaireRecipes
	if p.recipe != "resume" || p.resume.item != "verification" {
		t.Fatal("the real configuration control did not restore the saved checkpoint")
	}
	a = revealPreviewPart(t, r, "Messages & questions", ui.RoleCheckBox, "Run the typecheck")
	if !a.Checked {
		t.Fatal("the saved choice was not shown in the native checkbox")
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
	if slices.Contains(p.resume.state.Answers["verification"].Values, "typecheck") {
		t.Fatal("the actual workflow checkbox did not update its independent saved draft")
	}
	s.Page = "navigation"
	r.Frame()
	r.Key(ui.Cmd, ui.KeyK)
	r.Frame()
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Migration")
	r.Key(0, ui.KeyEnter)
	r.Frame()
	workspaceUnchanged := maps.EqualFunc(workspaceAnswers, s.Questionnaire.Answers, func(a, b conversation.Answer) bool {
		return slices.Equal(a.Values, b.Values) && a.Text == b.Text && a.Skipped == b.Skipped && a.Mode == b.Mode
	})
	if s.Page != "conversation" || s.Command || p.recipe != "resume" || p.resume.item != "verification" || !workspaceUnchanged {
		t.Fatalf("workflow page return: page=%q command=%v recipe=%q item=%q workspace=%+v", s.Page, s.Command, p.recipe, p.resume.item, s.Questionnaire.Answers)
	}
	if revealPreviewPart(t, r, "Messages & questions", ui.RoleCheckBox, "Run the typecheck").Checked || !questionnaireRecipesAccess(t, r, ui.RoleCheckBox, "Run migration tests").Checked {
		t.Fatal("the current native choices did not restore the retained workflow draft")
	}
}
