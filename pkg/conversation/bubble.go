package conversation

import (
	"fmt"
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type Alignment uint8

const (
	Start Alignment = iota
	End
)

type BubbleVariant uint8

const (
	Primary BubbleVariant = iota
	Secondary
	Muted
	Tinted
	Outline
	Destructive
	Ghost
)
const BubbleDefault = Primary

type Reaction struct {
	Key                string
	Label              string
	Count              int
	Selected, Disabled bool
	OnClick            func()
	OnSelectedChange   func(bool)
}
type BubbleTriggerKind uint8

const (
	BubbleButton BubbleTriggerKind = iota
	BubbleLink
)

// BubbleTrigger chooses a real native button or URL link for the content face.
// Put other interactive controls in a plain face rather than nesting buttons.
type BubbleTrigger struct {
	Kind       BubbleTriggerKind
	Label, URL string
	Disabled   bool
	Handle     *ui.Handle
	OnActivate func()
	// TextOnly selects MyGo's direct text-only Link. It derives a natural
	// name from text children, but cannot contain native boxes or icons.
	// The default compound link accepts arbitrary inert visuals and uses
	// Label (or URL when omitted) for its actual native Link target.
	TextOnly bool
}
type BubbleProps struct {
	Key      string
	Variant  BubbleVariant
	Align    Alignment
	Disabled bool
	// MaxWidth adds a DIP cap to the source's 80% (ghost 100%) maximum. Zero
	// leaves the source maximum; negative/nonfinite values expose Invalid.
	MaxWidth  float32
	Trigger   *BubbleTrigger
	Reactions []Reaction
	// Expanded and Summary retain the caller-controlled convenience disclosure.
	Expanded *bool
	Summary  string
}
type BubbleContentProps struct {
	Variant  BubbleVariant
	Align    Alignment
	Disabled bool
	Trigger  *BubbleTrigger
}
type BubbleSlots struct {
	// Content receives inherited face props. It can render one or more explicit
	// BubbleContent faces; leave nil for an empty compound shell.
	Content   func(*ui.Context, BubbleContentProps)
	Reactions func(*ui.Context)
}
type BubbleReactionSide uint8

const (
	BubbleReactionBottom BubbleReactionSide = iota
	BubbleReactionTop
)

type BubbleReactionAlign uint8

const (
	BubbleReactionEnd BubbleReactionAlign = iota
	BubbleReactionStart
)

type BubbleReactionsProps struct {
	Side  BubbleReactionSide
	Align BubbleReactionAlign
	Label string
	// Interactive removes the passive pill's padding, as :has(button) does.
	Interactive bool
}
type bubbleSlot uint8

const (
	bubbleFace bubbleSlot = iota
	bubbleReactions
	bubbleDisclosure
	bubbleLinkTarget
	bubbleLinkVisual
)

type bubbleIdentity struct{ key string }
type bubbleReactionIdentity struct {
	key        string
	occurrence int
}
type bubbleMessageScopeKey struct{}
type bubbleMessageScope struct {
	active bool
	align  Alignment
}

// withBubbleMessageAlignment applies the source's message ancestor rule only
// during the live Children call. The native root owns this numeric local scope.
func withBubbleMessageAlignment(c *ui.Context, align Alignment, content func()) {
	scope := ui.Local(c.Root(), bubbleMessageScopeKey{}, func() bubbleMessageScope { return bubbleMessageScope{} })
	previous := *scope
	*scope = bubbleMessageScope{true, align}
	defer func() { *scope = previous }()
	if content != nil {
		content()
	}
}
func bubbleInheritedAlignment(c *ui.Context, align Alignment) Alignment {
	scope := ui.Local(c.Root(), bubbleMessageScopeKey{}, func() bubbleMessageScope { return bubbleMessageScope{} })
	if scope.active && scope.align == End {
		return End
	}
	return align
}

func BubbleGroup(c *ui.Context, content func()) ui.Element {
	return ui.Column(c).MinWidth(0).Gap(8).Children(content)
}

// Bubble preserves the original convenience API, wrapping its children in one
// content face. BubbleWith exposes the source's separate content/reaction slots.
func Bubble(c *ui.Context, p BubbleProps, content func()) ui.Element {
	return BubbleWith(c, p, BubbleSlots{Content: func(c *ui.Context, face BubbleContentProps) {
		BubbleContent(c.Key(bubbleFace), face, func() {
			if p.Expanded == nil || *p.Expanded {
				if content != nil {
					content()
				}
			} else {
				ui.Text(c, p.Summary)
			}
			if p.Expanded != nil {
				label := "Show more"
				if *p.Expanded {
					label = "Show less"
				}
				if control.Button(c.Key(bubbleDisclosure), label, control.ButtonProps{Variant: control.Link, Size: control.Small, Disabled: p.Disabled}).Expanded(*p.Expanded).Clicked() {
					*p.Expanded = !*p.Expanded
				}
			}
		})
	}})
}

func BubbleWith(c *ui.Context, p BubbleProps, slots BubbleSlots) ui.Element {
	align := bubbleInheritedAlignment(c, p.Align)
	root := ui.Column(c).MinWidth(0).Gap(4).AlignItems(ui.Start).MaxWidthPercent(80).Disabled(p.Disabled).DimDisabled(false)
	if p.Key != "" {
		root.Key(bubbleIdentity{p.Key})
	}
	if p.Variant == Ghost {
		root.MaxWidthPercent(100)
	}
	if align == End {
		root.AlignSelf(ui.End)
	} else {
		root.AlignSelf(ui.Start)
	}
	if math.IsNaN(float64(p.MaxWidth)) || math.IsInf(float64(p.MaxWidth), 0) || p.MaxWidth < 0 {
		root.Invalid(true)
	} else if p.MaxWidth > 0 {
		root.MaxWidth(p.MaxWidth)
	}
	return root.Children(func() {
		if slots.Content != nil {
			slots.Content(c, BubbleContentProps{p.Variant, align, p.Disabled, p.Trigger})
		}
		if slots.Reactions != nil {
			slots.Reactions(c)
		} else if len(p.Reactions) > 0 {
			BubbleReactions(c.Key(bubbleReactions), BubbleReactionsProps{Interactive: true}, func() {
				occurrences := map[string]int{}
				for _, reaction := range p.Reactions {
					key := reaction.Key
					if key == "" {
						key = reaction.Label
					}
					occurrence := occurrences[key]
					occurrences[key]++
					reaction.Key = "" // The scoped occurrence key also supports duplicate labels.
					BubbleReaction(c.Key(bubbleReactionIdentity{key, occurrence}), reaction)
				}
			})
		}
	})
}

// BubbleContent is the clipped, styled face. Changed reports an activation;
// OnActivate and Handle refer to its real native target. Compound URL faces
// preserve native routing while their inert visuals can contain native boxes.
func BubbleContent(c *ui.Context, p BubbleContentProps, content func()) ui.Element {
	var face, target ui.Element
	compoundLink := p.Trigger != nil && p.Trigger.Kind == BubbleLink && !p.Trigger.TextOnly
	if p.Trigger == nil {
		face = ui.Column(c)
	} else if compoundLink {
		face = ui.Column(c).Role(ui.RoleNone)
	} else if p.Trigger.Kind == BubbleLink {
		face = ui.Link(c, "", p.Trigger.URL)
		target = face
	} else {
		face = ui.ButtonBase(c)
		target = face
	}
	face.Column().MinWidth(0).MaxWidthPercent(100).AlignItems(ui.Stretch).TextAlign(ui.Start).Gap(0).Padding(8, 12).Radius(theme.TokensOf(c).RadiusXL()).Border(1, ui.Transparent).FontSize(14).LineHeight(1.625).Clip().Disabled(p.Disabled).DimDisabled(false).FocusRing(false)
	if p.Align == End {
		face.AlignSelf(ui.End)
	} else {
		face.AlignSelf(ui.Start)
	}
	if p.Trigger != nil {
		if p.Trigger.Kind == BubbleButton && c.Direction() == ui.RightToLeft {
			face.TextAlign(ui.End)
		}
		label := p.Trigger.Label
		if compoundLink && label == "" {
			label = p.Trigger.URL
		}
		face.Label(label).Disabled(p.Disabled || p.Trigger.Disabled).Cursor(ui.CursorPointer)
	}
	face.Children(func() {
		if !compoundLink {
			if content != nil {
				content()
			}
			return
		}
		// Native Link is a text node, so containers cannot be its children.
		// Keep the complete clickable face on an actual Link and construct the
		// source's non-interactive visuals in a separate native sibling.
		label := p.Trigger.Label
		if label == "" {
			label = p.Trigger.URL
		}
		target = ui.Link(c.Key(bubbleLinkTarget), "", p.Trigger.URL).Label(label).LabelledBy(face).
			Absolute().Left(-1).Top(-1).Right(-1).Bottom(-1).Disabled(p.Disabled || p.Trigger.Disabled).DimDisabled(false).FocusRing(false)
		ui.Column(c.Key(bubbleLinkVisual)).MinWidth(0).MaxWidthPercent(100).Gap(0).IgnoreInput(true).
			Disabled(true).DimDisabled(false).AccessHidden(true).Children(content)
	})
	if p.Trigger != nil && p.Trigger.Handle != nil {
		target.Bind(p.Trigger.Handle)
	}
	t := theme.TokensOf(c)
	interactive := p.Trigger != nil && !p.Disabled && !p.Trigger.Disabled
	hover := interactive && target.Hovered()
	bg, fg, border := bubbleColors(t, p.Variant, hover)
	face.Background(bg).Border(1, border)
	inheritColor := p.Variant == Muted || ((p.Variant == Outline || p.Variant == Ghost) && !hover)
	if !inheritColor || p.Trigger != nil && p.Trigger.Kind == BubbleLink && p.Trigger.TextOnly {
		face.TextColor(fg)
	}
	if p.Variant == Ghost {
		face.Padding(0).Radius(0)
	}
	if p.Trigger != nil {
		motion.ApplyTransition(c, face)
		if target.FocusVisible() {
			face.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
		}
	}
	if p.Trigger != nil && target.Clicked() {
		face.SetChanged()
		if p.Trigger.OnActivate != nil {
			p.Trigger.OnActivate()
		}
	}
	return face
}

// BubbleReactions overlaps the content without consuming layout space. The
// source align classes are physical left/right even inside an RTL bubble.
func BubbleReactions(c *ui.Context, p BubbleReactionsProps, content func()) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Row(c).Absolute().ZIndex(10).AlignItems(ui.Center).Justify(ui.Center).Shrink(0).Gap(4).Padding(2, 6).Radius(999).Background(t.Muted).TextColor(t.Foreground).FontSize(14).LineHeight(1.5).Shadow(0, 0, 0, 3, t.Card)
	if p.Interactive {
		root.Padding(0)
	}
	if p.Side == BubbleReactionTop {
		root.Top(0).TranslatePercent(0, -75)
	} else {
		root.Bottom(0).TranslatePercent(0, 75)
	}
	if p.Align == BubbleReactionStart {
		root.Left(12)
	} else {
		root.Right(12)
	}
	if p.Label != "" {
		if p.Interactive {
			root.Role(ui.RoleGroup)
		} else {
			root.Role(ui.RoleImage)
		}
		root.Label(p.Label)
	}
	return root.Children(content)
}

// BubbleReaction is a controlled native pressed button. The callbacks update
// application state; Selected remains the source of truth on the next build.
func BubbleReaction(c *ui.Context, r Reaction) ui.Element {
	label := r.Label
	if r.Count > 0 {
		label += fmt.Sprintf(" %d", r.Count)
	}
	variant := control.Outline
	if r.Selected {
		variant = control.Secondary
	}
	button := control.Button(c, label, control.ButtonProps{Variant: variant, Size: control.ExtraSmall, Disabled: r.Disabled}).Role(ui.RoleToggleButton).Checked(r.Selected)
	if r.Key != "" {
		button.Key(bubbleReactionIdentity{r.Key, 0})
	}
	if button.Clicked() {
		if r.OnSelectedChange != nil {
			r.OnSelectedChange(!r.Selected)
		}
		if r.OnClick != nil {
			r.OnClick()
		}
	}
	return button
}
