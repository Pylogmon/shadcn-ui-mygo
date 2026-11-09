package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

func (s *State) conversation(c *ui.Context) {
	s.conversationPreview.initialize()
	s.heading(c, "A little conversation.", "Review a proposal, try file states and keep your notes together.")
	columns(c, func() { s.discussionPreview(c) }, func() { s.workspaceQuestions(c) })
	columns(c, func() { s.conversationFilesPreview(c) }, func() { s.conversationTracePreview(c) })
	s.conversationPreviewDialogs(c)
}
func (s *State) workspaceQuestions(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "Set up your workspace", "A controlled multi-step questionnaire.")
		conversation.Questionnaire(c, &s.Questionnaire, conversation.QuestionnaireProps{Questions: []conversation.Question{
			{ID: "role", Title: "What are you building?", Required: true, Choices: []control.Option{{Value: "desktop", Label: "Desktop application"}, {Value: "tool", Label: "Developer tool"}, {Value: "creative", Label: "Creative project"}}},
			{ID: "priorities", Title: "What matters most?", Multiple: true, Choices: []control.Option{{Value: "keyboard", Label: "Keyboard navigation"}, {Value: "themes", Label: "Flexible themes"}, {Value: "motion", Label: "Smooth animation"}}},
			{ID: "notes", Title: "Anything else?", Freeform: true, Placeholder: "Tell us more…"},
		}, OnSubmit: func(map[string]conversation.Answer) {
			overlay.Toast(c, ui.Toast{Title: "Workspace preferences collected"})
		}})
	})
}
