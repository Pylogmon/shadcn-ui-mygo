package gallery

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// The host retains this application state across pages. It owns three separate
// drafts; selecting a configuration does not reset the others or the existing
// workspace questionnaire. No build Context, Parts or Element is retained.
type questionnaireRecipesState struct {
	ready                           bool
	recipe                          string
	conditional, resume, validation questionnaireRecipeDraft
}
type questionnaireRecipeDraft struct {
	state   conversation.QuestionnaireState
	item    string
	result  map[string]conversation.Answer
	message string
}

func (p *questionnaireRecipesState) initialize() {
	if p.ready {
		return
	}
	p.ready = true
	p.recipe = "conditional"
	p.conditional.item = "runtime"
	p.resume.item = "verification"
	p.validation.item = "detail"
}
func (p *questionnaireRecipesState) view(c *ui.Context) {
	p.initialize()
	display.Card(c, func() {
		ui.Column(c).FillWidth().MinWidth(0).Gap(20).Children(func() {
			display.Typography(c, tr(c, "Questionnaire workflows"), display.H3)
			ui.Text(c, tr(c, "Try a conditional plan, restore a saved migration, or validate a response policy.")).TextColor(theme.TokensOf(c).MutedForeground)
			control.Select(c, &p.recipe, control.SelectProps{Label: tr(c, "Questionnaire configuration"), Options: localizedOptions(c, []control.Option{{Value: "conditional", Label: "Conditional execution plan"}, {Value: "resume", Label: "Resume saved migration"}, {Value: "validation", Label: "Validate response policy"}})}).MaxWidth(360)
			ui.Column(c.Key("questionnaire-recipe-" + p.recipe)).FillWidth().MaxWidth(448).MinWidth(0).Children(func() {
				switch p.recipe {
				case "resume":
					p.resume.render(c, "resume", resumeRecipeQuestions())
				case "validation":
					p.validation.render(c, "validation", validationRecipeQuestions())
				default:
					p.conditional.render(c, "conditional", conditionalRecipeQuestions(p.conditional.state.Answers))
				}
			})
		})
	}).Role(ui.RoleGroup).Label(tr(c, "Questionnaire workflow examples"))
}
func conditionalRecipeQuestions(answers map[string]conversation.Answer) []conversation.Question {
	cloud := slices.Contains(answers["runtime"].Values, "cloud")
	return []conversation.Question{
		{ID: "runtime", Title: "Where should the agent run?", Description: "Cloud runs add an environment question to this flow.", Required: true, DefaultAnswer: conversation.Answer{Values: []string{"local"}}, Choices: []control.Option{{Value: "local", Label: "Local workspace"}, {Value: "cloud", Label: "Cloud workspace"}}},
		{ID: "environment", Title: "Which cloud environment should it use?", Disabled: !cloud, Required: true, Choices: []control.Option{{Value: "preview", Label: "Preview"}, {Value: "staging", Label: "Staging"}, {Value: "isolated", Label: "Isolated sandbox"}}},
		{ID: "approval", Title: "When should the agent request approval?", Required: true, Choices: []control.Option{{Value: "writes", Label: "Before writing files"}, {Value: "commands", Label: "Before running commands"}, {Value: "sensitive", Label: "Only for sensitive actions"}}},
	}
}
func resumeRecipeQuestions() []conversation.Question {
	return []conversation.Question{
		{ID: "change", Title: "What kind of migration is this?", Description: "This answer was saved during the previous session.", Required: true, DefaultAnswer: conversation.Answer{Values: []string{"incremental"}}, Choices: []control.Option{{Value: "incremental", Label: "Incremental migration"}, {Value: "cutover", Label: "Single cutover"}}},
		{ID: "verification", Title: "How should the migration be verified?", Description: "These checks were selected during the previous session.", Required: true, Multiple: true, DefaultAnswer: conversation.Answer{Values: []string{"tests", "typecheck"}}, Choices: []control.Option{{Value: "tests", Label: "Run migration tests"}, {Value: "typecheck", Label: "Run the typecheck"}, {Value: "manual", Label: "Perform a manual smoke test"}}},
		{ID: "notes", Title: "Anything else the agent should remember?", Description: "This note was saved with the draft.", Freeform: true, InputLabel: "Saved migration note", DefaultAnswer: conversation.Answer{Text: "Keep the existing public API stable."}},
	}
}
func validationRecipeQuestions() []conversation.Question {
	return []conversation.Question{
		{ID: "detail", Title: "How much detail should the answer include?", Description: "Choose the response depth.", Required: true, Choices: []control.Option{{Value: "summary", Label: "Concise summary"}, {Value: "complete", Label: "Complete answer"}}},
		{ID: "audience", Title: "Who will read the answer?", Description: "Public answers require complete context.", Required: true, Choices: []control.Option{{Value: "team", Label: "My team"}, {Value: "public", Label: "Public audience"}}},
	}
}
func (m *questionnaireRecipeDraft) render(c *ui.Context, recipe string, questions []conversation.Question) {
	questions = localizedQuestions(c, questions)
	initial, progress, submit := "runtime", "Execution plan progress", "Save execution plan"
	switch recipe {
	case "resume":
		initial, progress, submit = "verification", "Saved migration progress", "Update draft"
	case "validation":
		initial, progress, submit = "detail", "Response policy progress", "Validate answers"
	}
	props := conversation.QuestionnaireProps{
		Questions: questions, DefaultItem: initial, Item: &m.item,
		OnItemChange: func(item string) { m.item = item },
		// A runtime choice can change a later item's Disabled flag, so schedule its
		// current metadata build after the accepted answer is published by the SDK.
		OnAnswerChange: func(id string, answer conversation.Answer) {
			// Keep the accepted snapshot, but stop presenting its success as
			// feedback for an edited draft. A same-frame submit already holds
			// the new answer, so its current success message remains visible.
			if m.result != nil {
				saved, present := m.result[id]
				if !present || !slices.Equal(saved.Values, answer.Values) || saved.Text != answer.Text || saved.Skipped != answer.Skipped {
					m.message = ""
				}
			}
			c.Invalidate()
		},
		OnReset: func(*conversation.QuestionnaireResetEvent) { m.result = nil; m.message = "Saved answers restored" },
		OnSubmit: func(result map[string]conversation.Answer) {
			m.result = result
			switch recipe {
			case "resume":
				m.message = "Draft updated"
			case "validation":
				m.message = "Agent response configured"
			default:
				m.message = "Execution plan saved"
			}
		},
	}
	if recipe == "validation" {
		props.ValidateAll = func(answers map[string]conversation.Answer) map[string]string {
			if slices.Contains(answers["audience"].Values, "public") && slices.Contains(answers["detail"].Values, "summary") {
				return map[string]string{"detail": "Public answers need enough context. Choose a complete answer."}
			}
			return nil
		}
	}
	conversation.QuestionnaireWith(c, &m.state, props, func(parts *conversation.QuestionnaireParts) {
		parts.Root.Label(tr(c, "Questionnaire ") + recipe)
		localizedQuestionnaireProgress(c, parts, progress)
		for _, question := range questions {
			parts.Item(question.ID, func(item *conversation.QuestionnaireItemParts) {
				item.Header(func(*ui.Context) {
					item.Title(tr(c, question.Title))
					if question.Description != "" {
						item.Description(tr(c, question.Description))
					}
				})
				item.Choices(func(*ui.Context) {
					for _, choice := range question.Choices {
						item.Choice(choice.Value)
					}
					if question.Freeform {
						item.Input(conversation.QuestionnaireInputProps{Label: tr(c, question.InputLabel)})
					}
				})
				item.Error()
			})
		}
		actions := parts.Actions(func(*ui.Context) {
			if recipe == "resume" {
				parts.Reset(conversation.QuestionnaireActionProps{Label: tr(c, "Reset changes")})
			}
			parts.Previous(conversation.QuestionnaireActionProps{Label: tr(c, "Previous")})
			parts.Next(conversation.QuestionnaireActionProps{Label: tr(c, "Next")})
			parts.Submit(conversation.QuestionnaireActionProps{Label: tr(c, submit)})
		})
		if recipe == "resume" {
			actions.ColumnTracks(ui.Fr(1), ui.FitContent(), ui.FitContent(), ui.FitContent())
		}
	})
	if m.message != "" {
		message := m.message
		if m.result != nil {
			message += " · " + questionnaireRecipeResult(recipe, m.result, questions)
		}
		ui.Text(c, message).Role(ui.RoleStatus).FontSize(12).LineHeight(1.5)
	}
}
func questionnaireRecipeResult(recipe string, result map[string]conversation.Answer, questions []conversation.Question) string {
	labels := func(id string) string {
		answer := result[id]
		if answer.Skipped {
			return "Skipped"
		}
		var out []string
		for _, q := range questions {
			if q.ID == id {
				for _, option := range q.Choices {
					if slices.Contains(answer.Values, option.Value) {
						out = append(out, option.Label)
					}
				}
			}
		}
		if answer.Text != "" {
			out = append(out, answer.Text)
		}
		if len(out) == 0 {
			return "Not applicable"
		}
		return strings.Join(out, ", ")
	}
	switch recipe {
	case "resume":
		return fmt.Sprintf("Migration: %s · Verification: %s · Notes: %s", labels("change"), labels("verification"), labels("notes"))
	case "validation":
		return fmt.Sprintf("Detail: %s · Audience: %s", labels("detail"), labels("audience"))
	default:
		return fmt.Sprintf("Runtime: %s · Environment: %s · Approval: %s", labels("runtime"), labels("environment"), labels("approval"))
	}
}
