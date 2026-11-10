package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func questionnaireRecipesAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing %q role%d", name, role)
	}
	return a
}
func questionnaireRecipesClick(t *testing.T, r *ui.Tester, role ui.Role, name string) {
	t.Helper()
	a := questionnaireRecipesAccess(t, r, role, name)
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
}
func questionnaireRecipesConfiguration(t *testing.T, r *ui.Tester, index int) {
	t.Helper()
	questionnaireRecipesClick(t, r, ui.RoleComboBox, "Questionnaire configuration")
	r.Key(0, ui.KeyHome)
	for i := 0; i < index; i++ {
		r.Key(0, ui.KeyDown)
	}
	r.Key(0, ui.KeyEnter)
}
func questionnaireRecipesSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_QUESTIONNAIRE_RECIPE_SNAPSHOT_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(f, r.Image()); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}

// One actual application flow covers the three fixed configurations, switching
// drafts and leaving their owner page. It is not a component-family matrix.
func TestQuestionnaireRecipesConditionalResumeValidationAndPageReturn(t *testing.T) {
	var p questionnaireRecipesState
	visible := true
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(24).Gap(16).Children(func() {
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave questionnaire page").Clicked() {
					visible = false
				}
				if ui.Button(c, "Return to questionnaire page").Clicked() {
					visible = true
				}
			})
			ui.Scroll(c).Fill().Children(func() {
				if visible {
					p.view(c)
				} else {
					ui.Text(c, "Another page")
				}
			})
		})
	}, 800, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	click := func(name string) { questionnaireRecipesClick(t, r, ui.RoleButton, name) }
	radio := func(name string) { questionnaireRecipesClick(t, r, ui.RoleRadio, name) }

	// Local skips the disabled environment; changing to cloud changes the same
	// collection's total and adds its required environment step.
	if !questionnaireRecipesAccess(t, r, ui.RoleRadio, "Local workspace").Checked || questionnaireRecipesAccess(t, r, ui.RoleProgress, "Execution plan progress").Max != 2 {
		t.Fatal("local defaults or disabled collection count absent")
	}
	click("Next")
	if p.conditional.item != "approval" {
		t.Fatal("local navigation did not skip environment")
	}
	radio("Before writing files")
	click("Save execution plan")
	if _, ok := p.conditional.result["environment"]; ok || !r.HasText("Execution plan saved · Runtime: Local workspace · Environment: Not applicable · Approval: Before writing files") {
		t.Fatal("disabled environment leaked into submitted local plan")
	}
	click("Previous")
	radio("Cloud workspace")
	if p.conditional.message != "" {
		t.Fatal("edited plan still presented the prior successful submission as current feedback")
	}
	if questionnaireRecipesAccess(t, r, ui.RoleProgress, "Execution plan progress").Max != 3 {
		t.Fatal("accepted runtime choice did not refresh dependent metadata")
	}
	click("Next")
	if p.conditional.item != "environment" {
		t.Fatal("cloud did not add environment step")
	}
	questionnaireRecipesSnapshot(t, r, "conditional-cloud")
	click("Next")
	if p.conditional.item != "environment" || !r.Focused("Preview") {
		t.Fatal("newly enabled required environment bypassed native validation")
	}
	radio("Staging")
	click("Next")
	click("Save execution plan")
	if !slices.Equal(p.conditional.result["environment"].Values, []string{"staging"}) {
		t.Fatal("latest environment was not submitted")
	}
	click("Previous")
	click("Previous")
	radio("Local workspace")
	click("Next")
	click("Save execution plan")
	if _, ok := p.conditional.result["environment"]; ok {
		t.Fatal("later-disabled prior answer remained successful form data")
	}

	// Saved defaults hydrate the second step, and Reset restores those defaults
	// rather than clearing them. Notes use the real native editor.
	questionnaireRecipesConfiguration(t, r, 1)
	if p.recipe != "resume" || p.resume.item != "verification" || !questionnaireRecipesAccess(t, r, ui.RoleCheckBox, "Run migration tests").Checked || !questionnaireRecipesAccess(t, r, ui.RoleCheckBox, "Run the typecheck").Checked {
		t.Fatal("saved migration did not resume at its saved checkpoint")
	}
	questionnaireRecipesSnapshot(t, r, "saved-migration")
	questionnaireRecipesClick(t, r, ui.RoleCheckBox, "Run the typecheck")
	questionnaireRecipesClick(t, r, ui.RoleCheckBox, "Perform a manual smoke test")
	click("Next")
	questionnaireRecipesClick(t, r, ui.RoleTextField, "Saved migration note")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Changed draft")
	click("Reset changes")
	if p.resume.item != "verification" || !questionnaireRecipesAccess(t, r, ui.RoleCheckBox, "Run the typecheck").Checked || questionnaireRecipesAccess(t, r, ui.RoleCheckBox, "Perform a manual smoke test").Checked || !r.HasText("Saved answers restored") {
		t.Fatal("Reset did not restore saved choices and checkpoint")
	}
	click("Next")
	if questionnaireRecipesAccess(t, r, ui.RoleTextField, "Saved migration note").Value != "Keep the existing public API stable." {
		t.Fatal("Reset cleared or retained dirty note instead of saved default")
	}
	questionnaireRecipesClick(t, r, ui.RoleTextField, "Saved migration note")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Keep rollout reversible.")
	click("Update draft")
	if p.resume.result["notes"].Text != "Keep rollout reversible." || !slices.Equal(p.resume.result["verification"].Values, []string{"tests", "typecheck"}) {
		t.Fatal("Update draft callback did not receive current answers")
	}

	// A valid per-item answer combination can still fail a cross-answer rule.
	// Validation returns to and focuses the earlier invalid native answer.
	questionnaireRecipesConfiguration(t, r, 2)
	radio("Concise summary")
	click("Next")
	radio("Public audience")
	click("Validate answers")
	if p.validation.item != "detail" || p.validation.result != nil || !r.HasText("Public answers need enough context. Choose a complete answer.") || !r.Focused("Concise summary") {
		t.Fatal("cross-answer failure did not return to the earlier invalid item")
	}
	questionnaireRecipesSnapshot(t, r, "response-policy-error")
	radio("Complete answer")
	click("Next")
	click("Validate answers")
	if !slices.Equal(p.validation.result["detail"].Values, []string{"complete"}) || !slices.Equal(p.validation.result["audience"].Values, []string{"public"}) || !strings.HasPrefix(p.validation.message, "Agent response configured") {
		t.Fatal("corrected policy failed current submit callback")
	}
	click("Leave questionnaire page")
	click("Return to questionnaire page")
	if p.recipe != "validation" || p.validation.item != "audience" || !r.Focused("Return to questionnaire page") {
		t.Fatal("page return lost checkpoint or replayed stale focus")
	}
	questionnaireRecipesConfiguration(t, r, 1)
	if p.resume.item != "notes" || questionnaireRecipesAccess(t, r, ui.RoleTextField, "Saved migration note").Value != "Keep rollout reversible." || p.resume.result["notes"].Text != "Keep rollout reversible." {
		t.Fatal("configuration/page return discarded the independent migration draft")
	}
}
