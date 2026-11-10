package gallery

import (
	"github.com/egoist/mygo/ui"
)

func (s *State) conversation(c *ui.Context) {
	s.conversationPreview.initialize()
	s.heading(c, "A little conversation.", "Review a proposal, try file states and keep your notes together.")
	columns(c, func() { s.discussionPreview(c) }, func() { s.workspaceQuestions(c) })
	columns(c, func() { s.conversationFilesPreview(c) }, func() { s.conversationTracePreview(c) })
	s.conversationComposition(c)
	s.conversationPreviewDialogs(c)
}
func (s *State) workspaceQuestions(c *ui.Context) {
	s.workspaceQuestionnaire(c)
	s.questionnaireRecipes.view(c)
}
