package conversation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type MessageProps struct {
	Sender                 string
	Align                  Alignment
	Avatar, Header, Footer func()
	Key, Label             string
	// Ghost declares that this message contains a ghost face: the source's
	// descendant selector removes header/footer horizontal padding. MyGo has
	// no public descendant data-slot query.
	Ghost bool
	// HasFooter declares a footer in a custom Children composition. The
	// standard Footer slot is detected directly, even when its callback is empty.
	HasFooter bool
}

type MessageSlots struct {
	Avatar, Header, Content, Footer func(*ui.Context)
	// Children replaces the standard avatar/content assembly. Compose the
	// exported native slot helpers in source order; they add no fake controls.
	Children func(*ui.Context)
}

type messagePart uint8

const (
	messageAvatar messagePart = iota
	messageContent
	messageHeader
	messageBody
	messageFooter
)

type messageIdentity string
type messageScopeKey struct{}
type messageScope struct {
	active, inContent, ghost, footer bool
	align                            Alignment
}

func messageInherited(c *ui.Context) messageScope {
	return *ui.Local(c.Root(), messageScopeKey{}, func() messageScope { return messageScope{} })
}
func withMessageScope(c *ui.Context, value messageScope, content func()) {
	s := ui.Local(c.Root(), messageScopeKey{}, func() messageScope { return messageScope{} })
	previous := *s
	*s = value
	defer func() { *s = previous }()
	if content != nil {
		content()
	}
}

func MessageGroup(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Column(c).MinWidth(0).Gap(8).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

// Message owns the row and metadata; compose Bubble inside content for a face.
func Message(c *ui.Context, p MessageProps, content func()) ui.Element {
	return MessageWith(c, p, MessageSlots{Content: func(*ui.Context) {
		if content != nil {
			content()
		}
	}})
}

// MessageWith composes all source slots. The end variant reverses only native
// layout, preserving source/Tab order. Applications own the controls, actions,
// image loading and transport inside each slot.
func MessageWith(c *ui.Context, p MessageProps, slots MessageSlots) ui.Element {
	if slots.Avatar == nil && p.Avatar != nil {
		slots.Avatar = func(*ui.Context) { p.Avatar() }
	}
	if slots.Header == nil {
		if p.Header != nil {
			slots.Header = func(*ui.Context) { p.Header() }
		} else if p.Sender != "" {
			slots.Header = func(child *ui.Context) { ui.Text(child, p.Sender) }
		}
	}
	if slots.Footer == nil && p.Footer != nil {
		slots.Footer = func(*ui.Context) { p.Footer() }
	}
	label := p.Label
	if label == "" {
		label = p.Sender
	}
	root := ui.Row(c).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Stretch).FontSize(14).LineHeight(10.0 / 7).Role(ui.RoleArticle).Label(label)
	if p.Key != "" {
		root.Key(messageIdentity(p.Key))
	}
	if p.Align == End {
		root.Reverse()
	}
	scope := messageScope{active: true, align: p.Align, ghost: p.Ghost, footer: p.HasFooter || slots.Children == nil && slots.Footer != nil}
	return root.Children(func() {
		withMessageScope(c, scope, func() {
			withBubbleMessageAlignment(c, p.Align, func() {
				if slots.Children != nil {
					slots.Children(c)
					return
				}
				if slots.Avatar != nil {
					MessageAvatar(c.Key(messageAvatar), slots.Avatar)
				}
				MessageContent(c.Key(messageContent), func(child *ui.Context) {
					if slots.Header != nil {
						MessageHeader(child.Key(messageHeader), slots.Header)
					}
					// The scoped context keeps an optional header from borrowing
					// the identity of native controls built through the body context.
					if slots.Content != nil {
						slots.Content(child.Key(messageBody))
					}
					if slots.Footer != nil {
						MessageFooter(child.Key(messageFooter), slots.Footer)
					}
				})
			})
		})
	})
}

// MessageSlot applies the source's end-aligned direct data-slot child rule to
// another native component (for example Attachment), without adding a wrapper.
// Bubble and the message header/footer already inherit this automatically.
// Raw text is deliberately left in normal content flow.
func MessageSlot(c *ui.Context, element ui.Element) ui.Element {
	scope := messageInherited(c)
	if scope.active && scope.inContent && scope.align == End {
		element.AlignSelf(ui.End)
	}
	return element
}

func MessageAvatar(c *ui.Context, content func(*ui.Context)) ui.Element {
	t := theme.TokensOf(c)
	avatar := ui.Column(c).MinWidth(32).Shrink(0).AlignSelf(ui.End).AlignItems(ui.Center).Justify(ui.Center).Clip().Radius(999).Background(t.Muted)
	if messageInherited(c).footer {
		avatar.Translate(0, -32)
	}
	return avatar.Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func MessageContent(c *ui.Context, content func(*ui.Context)) ui.Element {
	scope := messageInherited(c)
	scope.inContent = true
	return ui.Column(c).FillWidth().MinWidth(0).Shrink(1).Gap(10).Children(func() {
		withMessageScope(c, scope, func() {
			if content != nil {
				content(c)
			}
		})
	})
}

func MessageHeader(c *ui.Context, content func(*ui.Context)) ui.Element {
	return messageMetadata(c, false, content)
}
func MessageFooter(c *ui.Context, content func(*ui.Context)) ui.Element {
	return messageMetadata(c, true, content)
}
func messageMetadata(c *ui.Context, footer bool, content func(*ui.Context)) ui.Element {
	scope := messageInherited(c)
	row := ui.Row(c).MaxWidthPercent(100).MinWidth(0).AlignItems(ui.Center).FontSize(12).LineHeight(4.0 / 3).FontWeight(500).TextColor(theme.TokensOf(c).MutedForeground)
	if !scope.ghost {
		row.PaddingX(12)
	}
	if footer && scope.align == End {
		row.Justify(ui.End)
	}
	MessageSlot(c, row)
	return row.Children(func() {
		if content != nil {
			content(c)
		}
	})
}
