package conversation

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/shimmer"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type MarkerVariant uint8

const (
	Inline MarkerVariant = iota
	Bordered
	Separator
)

const (
	MarkerDefault   = Inline
	MarkerBorder    = Bordered
	MarkerSeparator = Separator
)

type MarkerTriggerKind uint8

const (
	MarkerButton MarkerTriggerKind = iota
	MarkerAnchor
)

type MarkerTrigger struct {
	Kind MarkerTriggerKind
	// Label names an anchor with arbitrary native visual content. MarkerText
	// supplies a natural name; a fluent Label on the returned root overrides it.
	// Native Link accepts text children only, so compound visuals occupy a
	// separate inert row. Interactive content belongs in a static Marker.
	Label, URL string
	Disabled   bool
	Handle     *ui.Handle
	OnActivate func()
}

type MarkerProps struct {
	Variant          MarkerVariant
	Icon             *ui.SVG
	Busy             bool
	Key, Label       string
	Status, Disabled bool
	Trigger          *MarkerTrigger
	// Shimmer affects the convenience API's text only. Compound compositions
	// opt in through MarkerText; a spinner belongs in a decorative icon slot.
	Shimmer bool
}

type MarkerIconProps struct{ Content func(*ui.Context) }
type MarkerTextProps struct {
	Shimmer bool
	// Color supplies a custom currentColor to the glyph-only shimmer. Plain
	// text otherwise inherits native parent color and font normally.
	Color       *ui.Color
	MeasureFont *ui.Font
}
type markerScopeKey struct{}
type markerScope struct {
	active  bool
	variant MarkerVariant
	base    ui.Color
	text    string
}
type markerIdentity string
type markerPart uint8

const (
	markerBefore markerPart = iota
	markerIcon
	markerContent
	markerAfter
	markerTarget
	markerVisual
)

func markerInherited(c *ui.Context) markerScope {
	return *ui.Local(c.Root(), markerScopeKey{}, func() markerScope { return markerScope{} })
}
func withMarkerScope(c *ui.Context, value markerScope, content func(*ui.Context)) string {
	s := ui.Local(c.Root(), markerScopeKey{}, func() markerScope { return markerScope{} })
	previous := *s
	*s = value
	defer func() { *s = previous }()
	if content != nil {
		content(c)
	}
	return s.text
}

// Marker preserves the text/icon convenience entry point. Static markers are
// not live regions: explicitly choose Status or Busy for advisory updates.
func Marker(c *ui.Context, text string, p MarkerProps) ui.Element {
	if p.Label == "" && p.Trigger != nil && p.Trigger.Label == "" {
		p.Label = text
	}
	return MarkerWith(c, p, func(child *ui.Context) {
		if p.Icon != nil {
			MarkerIcon(child.Key(markerIcon), p.Icon)
		}
		MarkerContent(child.Key(markerContent), func(textContext *ui.Context) { MarkerText(textContext, text, MarkerTextProps{Shimmer: p.Shimmer}) })
	})
}

// MarkerWith builds the source's direct icon/content/control slots. A Trigger
// chooses an actual native button or URL link. Without one it adds no input.
// The returned visual root reports Changed for an activation. Trigger.Handle
// binds the actual native focus target; Label also names that target.
func MarkerWith(c *ui.Context, p MarkerProps, content func(*ui.Context)) ui.Element {
	t := theme.TokensOf(c)
	var root, target ui.Element
	anchor := p.Trigger != nil && p.Trigger.Kind == MarkerAnchor
	if p.Trigger == nil || anchor {
		root = ui.Row(c)
	} else {
		root = ui.ButtonBase(c)
		target = root
	}
	root.Row().FillWidth().MinHeight(16).Gap(8).AlignItems(ui.Center).FontSize(14).LineHeight(10.0 / 7).TextColor(t.MutedForeground).
		TextAlign(ui.Start).Disabled(p.Disabled).DimDisabled(false).Busy(p.Busy)
	if c.Direction() == ui.RightToLeft {
		root.TextAlign(ui.End)
	}
	if p.Key != "" {
		root.Key(markerIdentity(p.Key))
	}
	label := p.Label
	base := t.MutedForeground
	if p.Trigger != nil {
		root.Disabled(p.Disabled || p.Trigger.Disabled).Cursor(ui.CursorPointer)
		if label == "" {
			label = p.Trigger.Label
		}
		if root.Hovered() && !p.Disabled && !p.Trigger.Disabled {
			base = t.Foreground
			root.TextColor(base)
		}
		if anchor {
			root.Underline().UnderlineOffset(3)
		}
	}
	if p.Status || p.Busy {
		root.Role(ui.RoleStatus)
	}
	if label != "" {
		root.Label(label)
	}
	if p.Variant == Bordered {
		root.BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Padding(0, 0, 8, 0)
	}
	buildContent := func() {
		if p.Variant == Separator {
			ui.Box(c.Key(markerBefore)).Grow(1).Shrink(1).MinWidth(0).Height(1).Background(t.Border).Margin(0, 4, 0, 0).AccessHidden(true).IgnoreInput(true)
		}
		natural := withMarkerScope(c, markerScope{active: true, variant: p.Variant, base: base}, content)
		if anchor && label == "" {
			label = natural
			if label == "" {
				label = p.Trigger.URL
			}
			root.Label(label)
		}
		if p.Variant == Separator {
			ui.Box(c.Key(markerAfter)).Grow(1).Shrink(1).MinWidth(0).Height(1).Background(t.Border).Margin(0, 0, 0, 4).AccessHidden(true).IgnoreInput(true)
		}
	}
	root.Children(func() {
		if !anchor {
			buildContent()
			return
		}
		// Keep URL/router navigation, Enter-only activation, focus and disabled
		// handling on an actual native Link. Its text-only child restriction
		// requires a separate decorative visual row for icon/control slots.
		root.Role(ui.RoleNone)
		target = ui.Link(c.Key(markerTarget), "", p.Trigger.URL).Label(label).LabelledBy(root).Absolute().Fill().Disabled(p.Disabled || p.Trigger.Disabled).DimDisabled(false)
		if target.Hovered() && !target.IsDisabled() {
			base = t.Foreground
			root.TextColor(base)
		}
		ui.Row(c.Key(markerVisual)).FillWidth().MinWidth(0).Gap(8).AlignItems(ui.Center).IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(buildContent)
	})
	if p.Trigger != nil && target.Valid() {
		if p.Trigger.Handle != nil {
			target.Bind(p.Trigger.Handle)
		}
		if target.Clicked() {
			root.SetChanged()
			if p.Trigger.OnActivate != nil {
				p.Trigger.OnActivate()
			}
		}
	}
	return root
}

func MarkerIcon(c *ui.Context, icon *ui.SVG, props ...MarkerIconProps) ui.Element {
	p := MarkerIconProps{}
	if len(props) > 0 {
		p = props[0]
	}
	return ui.Box(c).Size(16, 16).Shrink(0).Center().AccessHidden(true).IgnoreInput(true).Children(func() {
		scope := ui.Local(c.Root(), markerScopeKey{}, func() markerScope { return markerScope{} })
		previousText := scope.text
		defer func() { scope.text = previousText }()
		if p.Content != nil {
			p.Content(c)
		} else if icon != nil {
			ui.Icon(c, icon).Size(16, 16).AccessHidden(true).IgnoreInput(true)
		}
	})
}

func MarkerContent(c *ui.Context, content func(*ui.Context)) ui.Element {
	root := ui.Column(c).MinWidth(0).Shrink(1)
	if markerInherited(c).variant == Separator {
		root.Shrink(0).TextAlign(ui.Center)
	}
	return root.Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func MarkerText(c *ui.Context, text string, props ...MarkerTextProps) ui.Element {
	markerCollectText(c, text)
	p := MarkerTextProps{}
	if len(props) > 0 {
		p = props[0]
	}
	if !p.Shimmer {
		root := ui.Text(c, text).MinWidth(0)
		if p.Color != nil {
			root.TextColor(*p.Color)
		}
		if text == "" {
			root.Size(0, 0)
		}
		return root
	}
	scope := markerInherited(c)
	base := theme.TokensOf(c).MutedForeground
	if scope.active {
		base = scope.base
	}
	if p.Color != nil {
		base = *p.Color
	}
	measure := p.MeasureFont
	if measure == nil {
		measure = &ui.Font{Family: c.Theme().Font, Size: 14, Weight: 400, LineHeight: 10.0 / 7}
	}
	root := shimmer.Text(c, text, shimmer.Props{Active: text != "", Base: base, MeasureFont: measure}).MinWidth(0)
	if text == "" {
		root.Size(0, 0)
	}
	return root
}

// MarkerLink keeps native URL/router/keyboard semantics. The source's links
// stay underlined, with a three-DIP offset, including before hover.
func MarkerLink(c *ui.Context, text, url string) ui.Element {
	markerCollectText(c, text)
	base := theme.TokensOf(c).MutedForeground
	if scope := markerInherited(c); scope.active {
		base = scope.base
	}
	link := ui.Link(c, text, url).TextColor(base).Underline().UnderlineOffset(3)
	if link.Hovered() {
		link.TextColor(theme.TokensOf(c).Foreground)
	}
	return link
}

func markerCollectText(c *ui.Context, text string) {
	scope := ui.Local(c.Root(), markerScopeKey{}, func() markerScope { return markerScope{} })
	if scope.active && text != "" {
		if scope.text != "" {
			scope.text += " "
		}
		scope.text += text
	}
}
