package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

type conversationPreviewState struct {
	ready                                bool
	helpful                              bool
	reply, reviewNote                    string
	fileState, fileSize, fileOrientation string
	fileProgress                         float64
	reportRemoved                        bool
	files                                ui.ScrollState
	trace                                conversation.TraceState
	traceStatus                          conversation.TraceStatus
	result                               string
	previewOpen, notesOpen               bool
	notesTrigger                         ui.Handle
	notesReturn                          *ui.Handle
}

func (p *conversationPreviewState) initialize() {
	if p.ready {
		return
	}
	p.ready = true
	p.fileState = conversation.AttachmentUploading
	p.fileSize = "default"
	p.fileOrientation = "horizontal"
	p.fileProgress = .64
	p.traceStatus = conversation.Running
	p.reviewNote = "Check focus visibility in both appearances."
	p.result = "Ready for a review."
}
func (p *conversationPreviewState) size() conversation.AttachmentSize {
	switch p.fileSize {
	case "small":
		return conversation.AttachmentSmall
	case "xs":
		return conversation.AttachmentExtraSmall
	}
	return conversation.AttachmentDefault
}
func (p *conversationPreviewState) orientation() ui.Orientation {
	if p.fileOrientation == "vertical" {
		return ui.Vertical
	}
	return ui.Horizontal
}

func (s *State) discussionPreview(c *ui.Context) {
	p := &s.conversationPreview
	p.initialize()
	display.Card(c, func() {
		s.cardHeader(c, "Project discussion", "A message can hold a native action, editable text and pressed reactions.")
		conversation.Marker(c, "Today", conversation.MarkerProps{Variant: conversation.Separator})
		conversation.Message(c, conversation.MessageProps{Sender: "Sofia", Avatar: func() { display.Avatar(c, display.AvatarProps{Fallback: "SD", Size: 32}) }}, func() {
			conversation.BubbleWith(c, conversation.BubbleProps{Key: "proposal", Variant: conversation.Secondary}, conversation.BubbleSlots{
				Content: func(c *ui.Context, face conversation.BubbleContentProps) {
					face.Trigger = &conversation.BubbleTrigger{Label: tr(c, "Open review notes"), Handle: &p.notesTrigger, OnActivate: func() { p.notesOpen = true; p.notesReturn = &p.notesTrigger }}
					conversation.BubbleContent(c, face, func() { ui.Text(c, tr(c, "I checked the focus states. Would you review the notes before we publish?")) })
				},
				Reactions: func(c *ui.Context) {
					conversation.BubbleReactions(c, conversation.BubbleReactionsProps{Interactive: true, Label: tr(c, "Message reactions")}, func() {
						count := 3
						if p.helpful {
							count++
						}
						conversation.BubbleReaction(c, conversation.Reaction{Key: "helpful", Label: tr(c, "Helpful"), Count: count, Selected: p.helpful, OnSelectedChange: func(on bool) { p.helpful = on; p.result = "Message reaction updated." }})
					})
				},
			}).Margin(0, 0, 18, 0)
		})
		conversation.Message(c, conversation.MessageProps{Sender: "You", Align: conversation.End}, func() {
			conversation.Bubble(c, conversation.BubbleProps{Key: "reply", Variant: conversation.Tinted}, func() { ui.Text(c, tr(c, "Yes. I'll review keyboard navigation and check the report.")) })
		})
		control.Input(c, &p.reply, control.InputProps{Label: tr(c, "Conversation reply"), Placeholder: tr(c, "Write a reply…")})
		if p.reply != "" {
			conversation.Bubble(c, conversation.BubbleProps{Key: "draft", Variant: conversation.Ghost}, func() { ui.Text(c, p.reply) })
		}
	})
}
func (s *State) conversationFilesPreview(c *ui.Context) {
	p := &s.conversationPreview
	p.initialize()
	display.Card(c, func() {
		s.cardHeader(c, "Review files", "Choose a presentation and try the transfer controls.")
		control.RadioGroup(c, &p.fileState, control.RadioGroupProps{Label: tr(c, "Attachment state"), Horizontal: true, Options: localizedOptions(c, []control.Option{
			{Value: conversation.AttachmentIdle, Label: "Idle"}, {Value: conversation.AttachmentUploading, Label: "Uploading"}, {Value: conversation.AttachmentProcessing, Label: "Processing"}, {Value: conversation.AttachmentError, Label: "Failed"}, {Value: conversation.AttachmentDone, Label: "Done"},
		})}).Wrap()
		control.RadioGroup(c, &p.fileSize, control.RadioGroupProps{Label: tr(c, "Attachment size"), Horizontal: true, Options: localizedOptions(c, []control.Option{{Value: "default", Label: "Default size"}, {Value: "small", Label: "Small size"}, {Value: "xs", Label: "Extra small size"}})}).Wrap()
		control.RadioGroup(c, &p.fileOrientation, control.RadioGroupProps{Label: tr(c, "Attachment orientation"), Horizontal: true, Options: localizedOptions(c, []control.Option{{Value: "horizontal", Label: "Horizontal files"}, {Value: "vertical", Label: "Vertical files"}})}).Wrap()
		group := conversation.AttachmentGroupWith(c, conversation.AttachmentGroupProps{State: &p.files, Label: tr(c, "Review attachments")}, func(c *ui.Context) {
			if !p.reportRemoved {
				props := conversation.AttachmentProps{ID: "report", Name: "audit-report.pdf", MediaType: "application/pdf", SizeBytes: 284000, State: p.fileState, Progress: p.fileProgress, Size: p.size(), Orientation: p.orientation(), OpenLabel: "Preview report", OnOpen: func() { p.previewOpen = true }, OnCancel: func() {
					p.fileState = conversation.AttachmentError
					p.result = "Transfer canceled. You can retry the report."
				}, OnRetry: func() {
					p.fileState = conversation.AttachmentUploading
					p.fileProgress = .25
					p.result = "Report transfer restarted."
				}, OnRemove: func() { p.reportRemoved = true; p.result = "Report removed from this review." }}
				if p.orientation() == ui.Vertical {
					props.OnOpen = nil
				}
				card := conversation.Attachment(c, props)
				if p.orientation() == ui.Horizontal {
					card.Width(300)
				}
			}
			for _, file := range []struct {
				id, name, kind, state string
				bytes                 int64
			}{{"tokens", "design-tokens.json", "application/json", conversation.AttachmentIdle, 2480}, {"archive", "screenshots.zip", "application/zip", conversation.AttachmentError, 942000}, {"approved", "review-notes.md", "text/markdown", conversation.AttachmentDone, 1840}} {
				card := conversation.Attachment(c, conversation.AttachmentProps{ID: file.id, Name: file.name, MediaType: file.kind, SizeBytes: file.bytes, State: file.state, Size: p.size(), Orientation: p.orientation()})
				if p.orientation() == ui.Horizontal {
					card.Width(230)
				}
			}
		}).FillWidth()
		if p.orientation() == ui.Vertical {
			group.Height(190)
		} else {
			group.Height(84)
		}
		if p.fileState == conversation.AttachmentUploading {
			control.Slider(c, &p.fileProgress, control.SliderProps{Min: 0, Max: 1, Step: .05, Label: tr(c, "Report progress"), FormatValue: func(_ int, value float64) string { return fmt.Sprintf("%.0f%%", value*100) }})
		}
		ui.Row(c).Gap(8).Wrap().Children(func() {
			if p.orientation() == ui.Vertical && !p.reportRemoved && control.Button(c, tr(c, "Preview report"), control.ButtonProps{Size: control.Small, Variant: control.Outline}).Clicked() {
				p.previewOpen = true
			}
			if control.Button(c, tr(c, "Complete transfer"), control.ButtonProps{Size: control.Small, Disabled: p.reportRemoved || p.fileState != conversation.AttachmentUploading && p.fileState != conversation.AttachmentProcessing}).Clicked() {
				p.fileState = conversation.AttachmentDone
				p.fileProgress = 1
				p.result = "Report transfer complete."
			}
			if p.reportRemoved && control.Button(c, tr(c, "Restore report"), control.ButtonProps{Size: control.Small, Variant: control.Outline}).Clicked() {
				p.reportRemoved = false
				p.fileState = conversation.AttachmentIdle
				p.result = "Report restored to this review."
			}
		})
		ui.Text(c, p.result).FontSize(13).Role(ui.RoleStatus)
	})
}
func (s *State) conversationTracePreview(c *ui.Context) {
	p := &s.conversationPreview
	p.initialize()
	display.Card(c, func() {
		s.cardHeader(c, "Review progress", "Expand a step to edit the notes used by this review.")
		ui.Row(c).Gap(8).Wrap().Children(func() {
			if control.Button(c, tr(c, "Finish review"), control.ButtonProps{Size: control.Small, Disabled: p.traceStatus == conversation.Complete}).Clicked() {
				p.traceStatus = conversation.Complete
			}
			if control.Button(c, tr(c, "Report a problem"), control.ButtonProps{Size: control.Small, Variant: control.Outline}).Clicked() {
				p.traceStatus = conversation.Error
			}
			if control.Button(c, tr(c, "Resume review"), control.ButtonProps{Size: control.Small, Variant: control.Ghost}).Clicked() {
				p.traceStatus = conversation.Running
			}
		})
		conversation.Trace(c, &p.trace, []conversation.TraceStep{
			{ID: "tokens", Title: tr(c, "Loaded design tokens"), Status: conversation.Complete, Detail: "Semantic colors, type scale and corner radii are ready."},
			{ID: "accessibility", Title: tr(c, "Review accessibility"), Status: p.traceStatus, Detail: "Try the actual editor below, then open the notes sheet.", KeepMounted: true, Content: func(c *ui.Context) {
				control.Input(c, &p.reviewNote, control.InputProps{Label: tr(c, "Review note"), Placeholder: tr(c, "Add a review note…")})
				if control.Button(c, tr(c, "Open notes sheet"), control.ButtonProps{Size: control.Small, Variant: control.Outline}).Clicked() {
					p.notesOpen = true
					p.notesReturn = nil
				}
			}},
			{ID: "reference", Title: tr(c, "Open shadcn reference"), Status: conversation.Complete, URL: "https://ui.shadcn.com/create"},
		})
	})
}

type conversationPreviewOwnerKey struct{}
type conversationPreviewOwner struct{ state *conversationPreviewState }

func (o *conversationPreviewOwner) Close() error {
	p := o.state
	p.previewOpen = false
	p.notesOpen = false
	p.notesReturn = nil
	p.notesTrigger.CancelFocus()
	return nil
}

func (s *State) conversationPreviewDialogs(c *ui.Context) {
	p := &s.conversationPreview
	owner := ui.Box(c.Key(conversationPreviewOwnerKey{})).Absolute().Size(0, 0).AccessHidden(true).IgnoreInput(true)
	ui.Local(owner, conversationPreviewOwnerKey{}, func() conversationPreviewOwner { return conversationPreviewOwner{p} })
	if !p.notesOpen && p.notesTrigger.Focused(c) {
		p.notesReturn = nil
	}
	overlay.Dialog(c, &p.previewOpen, overlay.DialogProps{Title: tr(c, "Report preview"), Description: tr(c, "A local preview of the review report."), Width: 540}, func() {
		display.CodeBlock(c, display.CodeBlockProps{Title: "audit-report.json", Language: "json", Code: "{\n  \"review\": \"keyboard and focus\",\n  \"status\": \"ready\"\n}", Height: 150, HideCopy: false})
		if control.Button(c, tr(c, "Close report preview"), control.ButtonProps{Size: control.Small}).Clicked() {
			p.previewOpen = false
		}
	})
	overlay.Sheet(c, &p.notesOpen, overlay.SheetProps{Title: tr(c, "Review notes"), Description: tr(c, "Your notes stay with this conversation."), Side: overlay.Right, Size: 400, ReturnFocus: p.notesReturn}, func() {
		control.Textarea(c, &p.reviewNote, control.TextareaProps{Label: tr(c, "Review notes editor"), Placeholder: tr(c, "Write a review note…")}).MinHeight(160)
		if control.Button(c, tr(c, "Save review notes"), control.ButtonProps{}).Clicked() {
			p.notesOpen = false
			p.result = "Review notes saved."
		}
	})
}
