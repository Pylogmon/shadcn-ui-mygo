package gallery

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type questionnairePreviewState struct {
	ready     bool
	item      string
	reviewing bool
	result    map[string]conversation.Answer
	focus     string
}

func workspaceQuestionnaireQuestions() []conversation.Question {
	return []conversation.Question{
		{ID: "role", Title: "What are you building?", Description: "Choose the kind of project you want to make.", Required: true, Choices: []control.Option{{Value: "desktop", Label: "Desktop application"}, {Value: "tool", Label: "Developer tool"}, {Value: "creative", Label: "Creative project"}}},
		{ID: "priorities", Title: "What matters most?", Description: "Choose all that apply, or skip this question.", Multiple: true, Choices: []control.Option{{Value: "keyboard", Label: "Keyboard navigation"}, {Value: "themes", Label: "Flexible themes"}, {Value: "motion", Label: "Smooth animation"}}},
		{ID: "notes", Title: "Anything else?", Description: "Add a short note, or skip this question. Up to 280 characters.", Freeform: true, InputLabel: "Workspace notes", Placeholder: "Tell us more…"},
	}
}

func (s *State) workspaceQuestionnaire(c *ui.Context) {
	p := &s.QuestionnairePreview
	questions := workspaceQuestionnaireQuestions()
	if !p.ready {
		p.ready = true
		index := max(0, min(s.Questionnaire.Index, len(questions)-1))
		p.item = questions[index].ID
	}
	card := display.Card(c, func() {
		if p.reviewing {
			s.questionnaireResult(c, questions)
			return
		}
		description := "Choose a few preferences to make this workspace your own."
		if p.result != nil {
			description = "Update a choice, then save your preferences."
		}
		s.cardHeader(c, "Set up your workspace", description)
		conversation.QuestionnaireWith(c, &s.Questionnaire, conversation.QuestionnaireProps{
			Questions: questions, DefaultItem: "role", Item: &p.item, Shortcuts: conversation.QuestionnaireLetters,
			OnItemChange: func(id string) { p.item = id },
			OnReset:      func(*conversation.QuestionnaireResetEvent) { p.result = nil; p.focus = "role" },
			OnSubmit:     func(result map[string]conversation.Answer) { p.result = result; p.reviewing = true; p.focus = "result" },
		}, func(parts *conversation.QuestionnaireParts) {
			parts.Progress()
			for _, question := range questions {
				parts.Item(question.ID, func(item *conversation.QuestionnaireItemParts) {
					item.Header(func(*ui.Context) { item.Title(question.Title); item.Description(question.Description) })
					item.Choices(func(c *ui.Context) {
						for _, choice := range question.Choices {
							item.Choice(choice.Value)
						}
						if question.Freeform {
							item.Input(conversation.QuestionnaireInputProps{
								Label: question.InputLabel, Placeholder: question.Placeholder,
								Validate: func(value string) string {
									if utf8.RuneCountInString(value) > 280 {
										return "Keep your note to 280 characters."
									}
									return ""
								},
								Editor: func(c *ui.Context, value *string) ui.Element {
									return control.Textarea(c, value, control.TextareaProps{Label: question.InputLabel, Placeholder: question.Placeholder, Rows: 3, Invalid: item.Invalid})
								},
							}).MinHeight(104)
						}
					})
					item.Error()
					if item.Active && p.focus == question.ID {
						item.Root.Focus()
						p.focus = ""
					}
				})
			}
			parts.Actions(func(*ui.Context) {
				parts.Previous(conversation.QuestionnaireActionProps{Label: "Back"})
				parts.Skip()
				parts.Next()
				label := "Save preferences"
				if p.result != nil {
					label = "Save changes"
				}
				parts.Submit(conversation.QuestionnaireActionProps{Label: label})
			})
			ghost := control.Ghost
			parts.Reset(conversation.QuestionnaireActionProps{Label: "Reset preferences", Size: control.Small, Variant: &ghost}).AlignSelf(ui.Start)
		})
	})
	ui.Local(card, questionnairePreviewOwnerKey{}, func() questionnairePreviewOwner { return questionnairePreviewOwner{p} })
}

func (s *State) questionnaireResult(c *ui.Context, questions []conversation.Question) {
	p := &s.QuestionnairePreview
	title := display.CardTitle(c, "Your workspace preferences").Role(ui.RoleHeading).Level(3).Focusable().TabStop(false)
	if p.focus == "result" {
		title.Focus()
		p.focus = ""
	}
	display.CardDescription(c, "Your choices are ready. Change an answer or start again.")
	for _, question := range questions {
		ui.Column(c.Key(question.ID)).FillWidth().MinWidth(0).Gap(6).Children(func() {
			ui.Text(c, question.Title).FontSize(14).FontWeight(600)
			answer := p.result[question.ID]
			text := "Skipped"
			if !answer.Skipped {
				labels := []string{}
				for _, choice := range question.Choices {
					if slices.Contains(answer.Values, choice.Value) {
						labels = append(labels, choice.Label)
					}
				}
				if answer.Text != "" {
					labels = append(labels, answer.Text)
				}
				text = strings.Join(labels, ", ")
			}
			ui.Text(c, text).FontSize(14).LineHeight(1.5).TextColor(theme.TokensOf(c).MutedForeground)
			if control.Button(c, workspaceQuestionnaireEditLabel(question.ID), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
				p.reviewing = false
				p.item = question.ID
				p.focus = question.ID
				c.Invalidate()
			}
		})
	}
	if control.Button(c, "Reset preferences", control.ButtonProps{Variant: control.Ghost, Size: control.Small}).Clicked() {
		s.Questionnaire.Reset(questions)
		p.item = "role"
		p.result = nil
		p.reviewing = false
		p.focus = "role"
		c.Invalidate()
	}
}

type questionnairePreviewOwnerKey struct{}
type questionnairePreviewOwner struct{ state *questionnairePreviewState }

func (o *questionnairePreviewOwner) Close() error { o.state.focus = ""; return nil }

func workspaceQuestionnaireEditLabel(id string) string {
	switch id {
	case "role":
		return "Edit project"
	case "priorities":
		return "Edit priorities"
	default:
		return "Edit notes"
	}
}
