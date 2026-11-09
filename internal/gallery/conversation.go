package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

func (s *State) conversation(c *ui.Context) {
	s.heading(c, "A little conversation.", "Messages, attachments, progress and guided questions.")
	columns(c, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Project discussion", "Composable messages with your own content.")
			conversation.Marker(c, "Today", conversation.MarkerProps{Variant: conversation.Separator})
			conversation.Message(c, conversation.MessageProps{Sender: "Sofia", Avatar: func() { display.Avatar(c, display.AvatarProps{Fallback: "SD", Size: 32}) }}, func() {
				conversation.Bubble(c, conversation.BubbleProps{Variant: conversation.Secondary}, func() { ui.Text(c, "Ready to build something beautiful in Go?") })
			})
			conversation.Message(c, conversation.MessageProps{Sender: "You", Align: conversation.End}, func() {
				conversation.Bubble(c, conversation.BubbleProps{Align: conversation.End}, func() { ui.Text(c, "Let's start with the design system.") })
			})
			conversation.Attachment(c, conversation.AttachmentProps{Name: "design-tokens.json", MediaType: "application/json", SizeBytes: 2480, OnOpen: func() {
				overlay.Toast(c, ui.Toast{Title: "Attachment selected", Description: "Applications supply their own file handler."})
			}})
			conversation.Trace(c, &s.Trace, []conversation.TraceStep{{ID: "tokens", Title: "Loaded design tokens", Detail: "Semantic colors, type scale and corner radii."}, {ID: "render", Title: "Rendered native components", Detail: "GPU rendering with MyGo."}})
		})
	}, func() {
		display.Card(c, func() {
			s.cardHeader(c, "Set up your workspace", "A controlled multi-step questionnaire.")
			conversation.Questionnaire(c, &s.Questionnaire, conversation.QuestionnaireProps{Questions: []conversation.Question{{ID: "role", Title: "What are you building?", Required: true, Choices: []control.Option{{Value: "desktop", Label: "Desktop application"}, {Value: "tool", Label: "Developer tool"}, {Value: "creative", Label: "Creative project"}}}, {ID: "priorities", Title: "What matters most?", Multiple: true, Choices: []control.Option{{Value: "keyboard", Label: "Keyboard navigation"}, {Value: "themes", Label: "Flexible themes"}, {Value: "motion", Label: "Smooth animation"}}}, {ID: "notes", Title: "Anything else?", Freeform: true, Placeholder: "Tell us more…"}}, OnSubmit: func(map[string]conversation.Answer) {
				overlay.Toast(c, ui.Toast{Title: "Workspace preferences collected"})
			}})
		})
	})
}
