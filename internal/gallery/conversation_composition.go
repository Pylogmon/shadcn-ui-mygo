package gallery

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

const compositionUpdate = "The review is ready. Check the message guide before sharing it with the team."
const compositionGuideURL = "https://github.com/shadcn-ui/ui/blob/8ab27715521cbd4f87ae6886cd989cbb678f6de7/apps/v4/content/docs/components/base/message.mdx"

var compositionReply = []string{
	"Start with the reading flow. ",
	"Keep each turn together, with enough space between the messages to follow the conversation. ",
	"New replies can keep moving into view while you're reading the latest message. ",
	"If you scroll back to an earlier turn, your place should stay put as the reply grows. ",
	"You can load older messages without losing that place, too. ",
	"Choose Jump to latest whenever you're ready to catch up, then keep the conversation going.",
}

type compositionMessage struct {
	id, sender, text string
	user             bool
}
type conversationCompositionState struct {
	ready, follow, streaming, sendRequested bool
	draft, notice, streamID                 string
	messages                                []compositionMessage
	sequence, history, chunk                int
	nextTick                                time.Time
	scroll                                  conversation.MessageScrollerState
	editor, send, stop                      ui.Handle
}

func (p *conversationCompositionState) initialize() {
	if p.ready {
		return
	}
	p.ready, p.follow = true, true
	p.messages = []compositionMessage{
		{id: "opening-question", sender: "You", user: true, text: "How can we make a long conversation comfortable to read?"},
		{id: "opening-answer", sender: "Assistant", text: "Give every turn a clear shape. Keep the author and their message together, and leave room between topics."},
		{id: "context-question", sender: "You", user: true, text: "What happens when I scroll back to an earlier message?"},
		{id: "context-answer", sender: "Assistant", text: "Your place stays yours. New text can arrive below without pulling you away from the message you're reading."},
		{id: "history-question", sender: "You", user: true, text: "I'd also like to revisit older decisions without losing my draft."},
		{id: "history-answer", sender: "Assistant", text: "Try loading earlier messages, then write a new reply below. Your draft stays with this conversation."},
	}
}
func (p *conversationCompositionState) sendMessage(now time.Time) {
	text := strings.TrimSpace(p.draft)
	if p.streaming || text == "" {
		return
	}
	p.sequence++
	id := fmt.Sprintf("composed-%d", p.sequence)
	p.messages = append(p.messages, compositionMessage{id: id + "-you", sender: "You", text: text, user: true}, compositionMessage{id: id + "-reply", sender: "Assistant"})
	p.draft, p.notice, p.streamID = "", "", id+"-reply"
	p.streaming, p.chunk = true, 0
	p.nextTick = now.Add(600 * time.Millisecond)
}
func (p *conversationCompositionState) advance(now time.Time) {
	if !p.streaming {
		return
	}
	if p.nextTick.IsZero() {
		p.nextTick = now.Add(600 * time.Millisecond)
	}
	if now.Before(p.nextTick) {
		return
	}
	for i := range p.messages {
		if p.messages[i].id != p.streamID {
			continue
		}
		p.messages[i].text += compositionReply[p.chunk]
		p.chunk++
		if p.chunk == len(compositionReply) {
			p.streaming = false
			p.nextTick = time.Time{}
			p.notice = "Reply complete."
		} else {
			p.nextTick = now.Add(600 * time.Millisecond)
		}
		return
	}
	p.streaming = false
	p.nextTick = time.Time{}
}
func (p *conversationCompositionState) prependHistory() {
	p.history++
	prefix := fmt.Sprintf("history-%d", p.history)
	older := []compositionMessage{
		{id: prefix + "-you", sender: "You", user: true, text: "Earlier, we agreed to leave room for both quick questions and longer replies."},
		{id: prefix + "-reply", sender: "Assistant", text: "And to keep the reading position steady when more context arrives. This is that earlier part of the conversation."},
	}
	p.messages = append(older, p.messages...)
	p.notice = "Earlier messages loaded."
}

func (s *State) conversationComposition(c *ui.Context) {
	s.ConversationComposition.initialize()
	ui.Column(c).FillWidth().MinWidth(0).Gap(20).Children(func() {
		columns(c, func() { s.messageComposition(c) }, func() { s.markerComposition(c) })
		s.conversationChatComposition(c)
	})
}
func (s *State) messageComposition(c *ui.Context) {
	p := &s.ConversationComposition
	p.initialize()
	display.Card(c, func() {
		s.cardHeader(c, "A shared update", "A profile, a little context and an action worth keeping together.")
		conversation.MessageWith(c, conversation.MessageProps{Key: "team-update", Label: "Team update"}, conversation.MessageSlots{
			Avatar: func(c *ui.Context) { display.Avatar(c, display.AvatarProps{Fallback: "MC", Size: 32}) },
			Header: func(c *ui.Context) { ui.Text(c, "Mira Chen · Just now") },
			Content: func(c *ui.Context) {
				conversation.BubbleWith(c, conversation.BubbleProps{Variant: conversation.Outline, Trigger: &conversation.BubbleTrigger{Kind: conversation.BubbleLink, Label: "Open the message guide", URL: compositionGuideURL}}, conversation.BubbleSlots{Content: func(c *ui.Context, face conversation.BubbleContentProps) {
					conversation.BubbleContent(c, face, func() {
						ui.Row(c).MinWidth(0).Gap(10).AlignItems(ui.Start).Children(func() {
							ui.Box(c).Size(24, 24).Shrink(0).Center().Radius(6).Background(theme.TokensOf(c).Muted).AccessHidden(true).Children(func() { ui.Text(c, "↗") })
							ui.Column(c).MinWidth(0).Gap(4).Children(func() {
								ui.Text(c, "Message guide").FontWeight(500)
								ui.Text(c, compositionUpdate)
							})
						})
					})
				}})
			},
			Footer: func(c *ui.Context) {
				ui.Text(c, "Shared with the team")
				if control.Button(c, "Copy team update", control.ButtonProps{Variant: control.Ghost, Size: control.ExtraSmall}).Clicked() {
					c.WriteClipboard(compositionUpdate)
					p.notice = "Team update copied."
				}
			},
		})
	})
}
func (s *State) markerComposition(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "A little context", "Small updates can separate a day or mark a moment.")
		conversation.Marker(c, "Project room created", conversation.MarkerProps{})
		conversation.Marker(c, "Design notes added", conversation.MarkerProps{Variant: conversation.MarkerBorder})
		conversation.Marker(c, "Today", conversation.MarkerProps{Variant: conversation.MarkerSeparator})
	})
}

type compositionChatOwnerKey struct{}
type compositionChatOwner struct{ state *conversationCompositionState }

func (o *compositionChatOwner) Close() error {
	if o.state != nil {
		o.state.nextTick = time.Time{}
	}
	o.state = nil
	return nil
}

func (s *State) conversationChatComposition(c *ui.Context) {
	p := &s.ConversationComposition
	p.initialize()
	if p.sendRequested {
		p.sendRequested = false
		p.sendMessage(c.Now())
	}
	p.advance(c.Now())
	root := display.Card(c, func() {
		s.cardHeader(c, "Keep your place", "Write a message, browse earlier turns and catch up whenever you're ready.")
		ui.Row(c).Gap(12).Wrap().AlignItems(ui.Center).Children(func() {
			if control.Button(c, "Load earlier messages", control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
				p.prependHistory()
			}
			control.Switch(c, &p.follow, control.SwitchProps{Label: "Follow new replies", Size: control.Small})
		})
		conversation.MessageScrollerWith(c, &p.scroll, conversation.MessageScrollerProviderProps{AutoScroll: p.follow}, func(c *ui.Context) {
			conversation.MessageScrollerViewport(c, conversation.MessageScrollerViewportProps{Label: "Team conversation"}, func(c *ui.Context) {
				conversation.MessageScrollerContent(c, conversation.MessageScrollerContentProps{PaddingStart: 16, PaddingEnd: 16, Label: "Conversation messages"}, func(c *ui.Context) {
					for _, message := range p.messages {
						conversation.MessageScrollerItem(c, conversation.MessageScrollerItemProps{ID: message.id, ScrollAnchor: message.user}, func(c *ui.Context) {
							s.compositionChatMessage(c, message)
						})
					}
					if p.streaming {
						conversation.MessageScrollerItem(c, conversation.MessageScrollerItemProps{ID: "writing-reply"}, func(c *ui.Context) {
							conversation.MarkerWith(c, conversation.MarkerProps{Status: true, Busy: true}, func(c *ui.Context) {
								conversation.MarkerIcon(c, nil, conversation.MarkerIconProps{Content: func(c *ui.Context) { display.Spinner(c, display.SpinnerProps{Decorative: true}) }})
								conversation.MarkerContent(c, func(c *ui.Context) {
									conversation.MarkerText(c, "Writing a reply…", conversation.MarkerTextProps{Shimmer: true})
								})
							})
						})
					}
				})
			})
			conversation.MessageScrollerButton(c.Key("beginning"), conversation.MessageScrollerButtonProps{Direction: conversation.MessageScrollerStart, Label: "Back to beginning"})
			conversation.MessageScrollerButton(c.Key("latest"), conversation.MessageScrollerButtonProps{Label: "Jump to latest"})
		}).Height(340).Shrink(0).Border(1, theme.TokensOf(c).Border).Radius(8)
		ui.Row(c).Gap(8).Wrap().AlignItems(ui.Center).Children(func() {
			control.Input(c, &p.draft, control.InputProps{Label: "Chat message", Placeholder: "Write a message…", Handle: &p.editor, OnSubmit: func(string) { p.sendRequested = true }}).Basis(0).Grow(1).MinWidth(180)
			if control.Button(c, "Send message", control.ButtonProps{Size: control.Small, Disabled: p.streaming || strings.TrimSpace(p.draft) == ""}).Bind(&p.send).Clicked() {
				p.sendRequested = true
				p.editor.Focus(c)
				c.Invalidate()
			}
			if control.Button(c, "Stop reply", control.ButtonProps{Variant: control.Outline, Size: control.Small, Disabled: !p.streaming}).Bind(&p.stop).Clicked() {
				p.streaming = false
				p.nextTick = time.Time{}
				p.notice = "Reply stopped."
				p.editor.Focus(c)
				c.Invalidate()
			}
		})
		if (!p.streaming && p.stop.Focused(c)) || (p.streaming && p.send.Focused(c)) {
			p.editor.Focus(c)
		}
		if p.notice != "" {
			ui.Text(c, p.notice).FontSize(12).TextColor(theme.TokensOf(c).MutedForeground).Role(ui.RoleStatus)
		}
	})
	owner := ui.Local(root, compositionChatOwnerKey{}, func() compositionChatOwner { return compositionChatOwner{} })
	if owner.state != p {
		owner.Close()
		owner.state = p
	}
	if p.streaming {
		c.After(max(time.Millisecond, p.nextTick.Sub(c.Now())))
	}
}

func (s *State) compositionChatMessage(c *ui.Context, message compositionMessage) {
	align, variant, avatar := conversation.Start, conversation.Ghost, "AI"
	if message.user {
		align, variant, avatar = conversation.End, conversation.Muted, "Y"
	}
	conversation.MessageWith(c, conversation.MessageProps{Align: align, Ghost: !message.user, Label: message.sender + " message"}, conversation.MessageSlots{
		Avatar: func(c *ui.Context) { display.Avatar(c, display.AvatarProps{Fallback: avatar, Size: 32}) },
		Header: func(c *ui.Context) { ui.Text(c, message.sender) },
		Content: func(c *ui.Context) {
			conversation.Bubble(c, conversation.BubbleProps{Variant: variant}, func() {
				if message.text != "" {
					ui.Text(c, message.text).Selectable()
				}
			})
		},
	})
}
