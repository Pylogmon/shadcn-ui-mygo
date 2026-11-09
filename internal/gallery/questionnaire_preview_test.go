package gallery

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func questionnairePreviewTester(s *State, width int) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		if s.Config.ReduceMotion {
			c.DisableAnimations()
		}
		theme.Scope(c, s.Config.Tokens(false)).Fill().Padding(24).Children(func() {
			ui.Button(c, "Page focus anchor")
			if s.Page == "conversation" {
				s.workspaceQuestionnaire(c)
			} else {
				ui.Text(c, "Another preview")
			}
		})
	}, width, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: s.Config.ReduceMotion}})
}
func questionnairePreviewStateForTest() *State {
	s := New()
	s.Page = "conversation"
	s.Config.ReduceMotion = true
	return s
}
func questionnairePreviewChooseFirst(t *testing.T, r *ui.Tester, choice string) {
	t.Helper()
	advancedClick(t, r, choice)
	advancedClick(t, r, "Next")
}

func TestCreatorQuestionnaireRequiredOptionalSkipAndVisibleResults(t *testing.T) {
	s := questionnairePreviewStateForTest()
	r := questionnairePreviewTester(s, 580)
	advancedClick(t, r, "Next")
	if !r.HasText("Choose an answer to continue.") || !r.Focused("Desktop application") {
		t.Fatal("required attempt did not validate and focus actual native radio")
	}
	advancedClick(t, r, "Developer tool")
	r.Key(0, ui.KeyEnter)
	if s.QuestionnairePreview.item != "priorities" {
		t.Fatal("selected native Enter did not accept controlled step request")
	}
	advancedClick(t, r, "Next")
	if !r.HasText("Choose an answer or skip this question.") || !r.Focused("Keyboard navigation") {
		t.Fatal("optional question silently advanced without an answer/Skip")
	}
	advancedClick(t, r, "Skip")
	notes := advancedAccess(t, r, ui.RoleTextField, "Workspace notes")
	if !notes.Multiline {
		t.Fatal("notes use no real native multiline editor")
	}
	advancedClick(t, r, "Save preferences")
	if s.QuestionnairePreview.reviewing || !r.HasText("Choose an answer or skip this question.") {
		t.Fatal("empty optional note silently submitted")
	}
	advancedClick(t, r, "Skip")
	if !s.QuestionnairePreview.reviewing || !s.QuestionnairePreview.result["priorities"].Skipped || !s.QuestionnairePreview.result["notes"].Skipped {
		t.Fatal("explicit last Skip did not create application result")
	}
	if !r.HasText("Your workspace preferences") || !r.HasText("Developer tool") || !r.Focused("Your workspace preferences") {
		t.Fatal("submitted result was not visible/focused")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Workspace notes"); ok {
		t.Fatal("result screen retained an active hidden editor")
	}
}
func TestCreatorQuestionnaireMultipleNotesValidationAndIME(t *testing.T) {
	s := questionnairePreviewStateForTest()
	r := questionnairePreviewTester(s, 580)
	questionnairePreviewChooseFirst(t, r, "Creative project")
	advancedClick(t, r, "Smooth animation")
	advancedClick(t, r, "Keyboard navigation")
	advancedClick(t, r, "Next")
	advancedClick(t, r, "Workspace notes")
	r.Type(strings.Repeat("界", 281))
	advancedClick(t, r, "Save preferences")
	if s.QuestionnairePreview.reviewing || !r.HasText("Keep your note to 280 characters.") || !advancedAccess(t, r, ui.RoleTextField, "Workspace notes").Invalid {
		t.Fatal("Go note validator did not block long Unicode text and expose actual invalid state")
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("优", 1)
	r.Key(ui.Cmd, ui.KeyEnter)
	if s.QuestionnairePreview.reviewing {
		t.Fatal("IME confirmation submitted the preferences")
	}
	r.Type("优先键盘")
	r.Key(ui.Cmd, ui.KeyEnter)
	if !s.QuestionnairePreview.reviewing || s.QuestionnairePreview.result["notes"].Text != "优先键盘" {
		t.Fatalf("native composition did not save actual note: %+v", s.QuestionnairePreview.result)
	}
	values := s.QuestionnairePreview.result["priorities"].Values
	if len(values) != 2 || !slices.Contains(values, "keyboard") || !slices.Contains(values, "motion") || !r.HasText("Keyboard navigation, Smooth animation") {
		t.Fatal("multiple choices did not reach the human-readable result")
	}
}
func TestCreatorQuestionnaireReviewEditControlledStepAndReset(t *testing.T) {
	s := questionnairePreviewStateForTest()
	r := questionnairePreviewTester(s, 580)
	questionnairePreviewChooseFirst(t, r, "Desktop application")
	advancedClick(t, r, "Flexible themes")
	advancedClick(t, r, "Next")
	advancedClick(t, r, "Workspace notes")
	r.Type("Initial note")
	advancedClick(t, r, "Save preferences")
	advancedClick(t, r, "Edit notes")
	if s.QuestionnairePreview.reviewing || s.QuestionnairePreview.item != "notes" || !r.Focused("Anything else?") {
		t.Fatal("Review Edit did not return/focus requested controlled item")
	}
	advancedClick(t, r, "Workspace notes")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Updated note")
	if s.QuestionnairePreview.result["notes"].Text != "Initial note" {
		t.Fatal("unsaved editor mutation changed the retained saved result")
	}
	advancedClick(t, r, "Save changes")
	if !s.QuestionnairePreview.reviewing || !r.HasText("Updated note") {
		t.Fatal("edited note did not replace the application result on save")
	}
	advancedClick(t, r, "Edit project")
	if s.QuestionnairePreview.item != "role" || !r.Focused("What are you building?") || !advancedAccess(t, r, ui.RoleRadio, "Desktop application").Checked {
		t.Fatal("Review Edit did not restore fixed choice and fieldset focus")
	}
	advancedClick(t, r, "Reset preferences")
	if s.QuestionnairePreview.result != nil || s.QuestionnairePreview.reviewing || s.QuestionnairePreview.item != "role" || len(s.Questionnaire.Answers["role"].Values) > 0 || s.Questionnaire.Answers["notes"].Text != "" {
		t.Fatal("form Reset failed to clear saved result and native-answer defaults")
	}
	advancedClick(t, r, "Next")
	if !r.HasText("Choose an answer to continue.") {
		t.Fatal("reset left answered state behind")
	}
}
func TestCreatorQuestionnaireDraftAndResultPersistAcrossPagesWithoutFocusGrab(t *testing.T) {
	s := questionnairePreviewStateForTest()
	r := questionnairePreviewTester(s, 580)
	questionnairePreviewChooseFirst(t, r, "Developer tool")
	advancedClick(t, r, "Flexible themes")
	advancedClick(t, r, "Next")
	advancedClick(t, r, "Workspace notes")
	r.Type("Keep this draft")
	s.Page = "display"
	r.Frame()
	advancedClick(t, r, "Page focus anchor")
	s.Page = "conversation"
	r.Frame()
	if s.QuestionnairePreview.item != "notes" || advancedAccess(t, r, ui.RoleTextField, "Workspace notes").Value != "Keep this draft" || !r.Focused("Page focus anchor") {
		t.Fatal("page navigation lost native draft/step or replayed stale focus")
	}
	advancedClick(t, r, "Save preferences")
	s.Page = "themes"
	r.Frame()
	advancedClick(t, r, "Page focus anchor")
	s.Page = "conversation"
	r.Frame()
	if !s.QuestionnairePreview.reviewing || !r.HasText("Keep this draft") || !r.Focused("Page focus anchor") {
		t.Fatal("saved review was tied to a discarded frame or stole focus on return")
	}
	advancedClick(t, r, "Reset preferences")
	if s.QuestionnairePreview.reviewing || s.QuestionnairePreview.result != nil || s.QuestionnairePreview.item != "role" || !r.Focused("What are you building?") {
		t.Fatal("review Reset did not reopen empty first item")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reduced-motion questionnaire preview did not settle idle")
	}
}
func TestCreatorQuestionnaireNarrowFormKeepsControlsWithinCard(t *testing.T) {
	s := questionnairePreviewStateForTest()
	r := questionnairePreviewTester(s, 390)
	questionnairePreviewChooseFirst(t, r, "Desktop application")
	advancedClick(t, r, "Keyboard navigation")
	advancedClick(t, r, "Next")
	input := advancedAccess(t, r, ui.RoleTextField, "Workspace notes").Bounds
	button := advancedAccess(t, r, ui.RoleButton, "Save preferences").Bounds
	if input.W < 200 || input.H < 104 || button.X < input.X || button.X+button.W > input.X+input.W+.1 {
		t.Fatalf("narrow actual editor/actions escaped content: input%v button%v", input, button)
	}
	advancedClick(t, r, "Workspace notes")
	r.Type("First line")
	r.Key(0, ui.KeyEnter)
	r.Type("Second line")
	if s.QuestionnairePreview.reviewing || s.Questionnaire.Answers["notes"].Text != "First line\nSecond line" {
		t.Fatal("native multiline Enter submitted instead of editing")
	}
}

func TestCreatorQuestionnaireWholeConversationPageUsesNativeFlow(t *testing.T) {
	for _, width := range []int{500, 1450} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			s := questionnairePreviewStateForTest()
			r := ui.NewTester(s.View, width, 1000)
			click := func(label string) { t.Helper(); conversationReveal(t, r, label); advancedClick(t, r, label) }
			click("Desktop application")
			click("Next")
			click("Skip")
			click("Workspace notes")
			r.Type("Whole-page draft")
			s.Page = "display"
			r.Frame()
			s.Page = "conversation"
			r.Frame()
			conversationReveal(t, r, "Workspace notes")
			if advancedAccess(t, r, ui.RoleTextField, "Workspace notes").Value != "Whole-page draft" {
				t.Fatal("actual Gallery page lost retained native questionnaire draft")
			}
			click("Save preferences")
			conversationReveal(t, r, "Your workspace preferences")
			if !s.QuestionnairePreview.reviewing {
				t.Fatal("legacy questionnaire remained wired to conversation page")
			}
			click("Edit priorities")
			click("Keyboard navigation")
			click("Next")
			click("Save changes")
			if !slices.Equal(s.QuestionnairePreview.result["priorities"].Values, []string{"keyboard"}) {
				t.Fatal("whole-page Review Edit did not clear skip and save native checkbox")
			}
			click("Reset preferences")
			conversationReveal(t, r, "Desktop application")
			if advancedAccess(t, r, ui.RoleRadio, "Desktop application").Checked || s.QuestionnairePreview.result != nil {
				t.Fatal("whole-page Reset did not rebuild empty first step")
			}
		})
	}
}

func TestCreatorQuestionnaireNoteValidationColorsActualNativeEditor(t *testing.T) {
	s := questionnairePreviewStateForTest()
	r := questionnairePreviewTester(s, 580)
	questionnairePreviewChooseFirst(t, r, "Desktop application")
	advancedClick(t, r, "Skip")
	advancedClick(t, r, "Workspace notes")
	r.Type(strings.Repeat("字", 281))
	advancedClick(t, r, "Save preferences")
	field := advancedAccess(t, r, ui.RoleTextField, "Workspace notes")
	pixel := r.Image().RGBAAt(int(field.Bounds.X), int(field.Bounds.Y+field.Bounds.H/2))
	expected := s.Config.Tokens(false).Destructive
	if pixel.R != expected.R || pixel.G != expected.G || pixel.B != expected.B {
		t.Fatalf("custom native validation has no source destructive border: got%v expected%v", pixel, expected)
	}
}
