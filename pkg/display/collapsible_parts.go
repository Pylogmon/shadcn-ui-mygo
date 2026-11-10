package display

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// CollapsibleRootProps separates the native disclosure state from its layout.
// KeepMounted retains content after its first opening. As with CollapsibleWith,
// explicit Focus/AutoFocus in retained children must be conditional on Open.
type CollapsibleRootProps struct {
	Label                 string
	Disabled, KeepMounted bool
	OnOpenChange          func(bool)
}

type CollapsibleTriggerProps struct {
	Label    string
	Disabled bool
	// Style changes the existing native disclosure trigger. It must not retain
	// the Element. Nil supplies the usual text-and-chevron appearance.
	Style func(ui.Element)
}

type CollapsibleContentProps struct{ Label string }

// CollapsibleContentParts contains this build's actual clipping root and inner
// content. Put padding/gap on Content so the native height animation measures
// them; Root can take grid span, width, margin and alignment declarations.
type CollapsibleContentParts struct{ Root, Content ui.Element }

type collapsibleScopeState struct {
	mounted bool
	panel   ui.Handle
}

// CollapsibleScope is valid only during the CollapsibleRoot callback. Trigger
// and Content may be declared in either order and in separate native layouts.
// Its temporary references and builders are cleared before Root returns.
type CollapsibleScope struct {
	root       ui.Element
	open       *bool
	props      CollapsibleRootProps
	state      *collapsibleScopeState
	part       ui.CollapsibleParts
	trigger    ui.Element
	content    CollapsibleContentParts
	carrier    ui.Element
	contentCtx *ui.Context
	build      func(*ui.Context)
	alive      bool
	hasContent bool
}

// CollapsibleRoot composes a native disclosure without fixing the trigger and
// content next to each other. It requires one Trigger; Content is optional.
// Content builds once at the end of this same callback, after the trigger has
// handled its input. Nothing is queued for another frame or stored in a model.
func CollapsibleRoot(c *ui.Context, open *bool, p CollapsibleRootProps, build func(*CollapsibleScope)) ui.Element {
	if open == nil {
		panic("display.CollapsibleRoot: open state must not be nil")
	}
	root := ui.Column(c).FillWidth().MinWidth(0).Shrink(0)
	s := &CollapsibleScope{root: root, open: open, props: p, alive: true}
	s.state = ui.Local(root, "collapsible-parts", func() collapsibleScopeState { return collapsibleScopeState{} })
	defer func() { *s = CollapsibleScope{} }()
	root.Children(func() {
		if build != nil {
			build(s)
		}
		s.finish(c)
	})
	return root
}

func (s *CollapsibleScope) check() {
	if s == nil || !s.alive {
		panic("display.CollapsibleScope: parts are valid only during CollapsibleRoot's build callback")
	}
}

func (s *CollapsibleScope) Open() bool { s.check(); return *s.open }

// Progress is MyGo's current 200ms native disclosure progress. Call it after
// Trigger, or from Content's deferred builder, for an animated decorative icon.
func (s *CollapsibleScope) Progress() float32 {
	s.check()
	if !s.trigger.Valid() {
		if *s.open {
			return 1
		}
		return 0
	}
	return s.part.Progress()
}

func (s *CollapsibleScope) Trigger(c *ui.Context, p CollapsibleTriggerProps, children func(*ui.Context)) ui.Element {
	s.check()
	if s.trigger.Valid() {
		panic("display.CollapsibleRoot: declare one Trigger")
	}
	s.part = ui.CollapsibleBase(c.Key("collapsible-trigger"), s.open)
	tr := s.part.Trigger
	s.trigger = tr
	label := p.Label
	if label == "" {
		label = s.props.Label
	}
	tr.Label(label)
	if p.Style != nil {
		p.Style(tr)
	} else {
		t := theme.TokensOf(c)
		tr.FillWidth().Padding(8).Gap(8).AlignItems(ui.Center).Radius(t.RadiusMD())
	}
	tr.Disabled(s.props.Disabled || p.Disabled)
	if s.props.Disabled || p.Disabled {
		tr.Opacity(.5)
	}
	tr.Children(func() {
		if children != nil {
			children(c)
		} else {
			ui.Text(c, label).Grow(1).MinWidth(0)
			indicator.ChevronDown(c).Rotate(180 * s.part.Progress())
		}
	})
	if tr.Changed() {
		s.root.SetChanged()
		if s.props.OnOpenChange != nil {
			s.props.OnOpenChange(*s.open)
		}
	}
	return tr.Expanded(*s.open)
}

// Content declares its real native position now, and builds its children once
// Root has processed Trigger. A hidden theme carrier reenters the declaration's
// Theme only for the build; it is not an ancestor of the actual content, whose
// ordinary font/size/color/direction continue to inherit from its real layout.
func (s *CollapsibleScope) Content(c *ui.Context, p CollapsibleContentProps, children func(*ui.Context)) CollapsibleContentParts {
	s.check()
	if s.hasContent {
		panic("display.CollapsibleRoot: declare one Content")
	}
	s.hasContent = true
	s.contentCtx, s.build = c, children
	label := p.Label
	if label == "" {
		label = s.props.Label
	}
	panel := ui.Column(c.Key("collapsible-content")).FillWidth().MinWidth(0).Shrink(0).ClipY().Role(ui.RoleRegion).Label(label)
	panel.Bind(&s.state.panel)
	panel.Children(func() {
		s.content.Content = ui.Column(c.Key("collapsible-inner")).FillWidth().MinWidth(0).Shrink(0)
		s.carrier = ui.Themed(c.Key("collapsible-build-theme"), c.Theme()).Hidden(true).Background(ui.Transparent).AccessHidden(true).IgnoreInput(true)
	})
	s.content.Root = panel
	return s.content
}

func (s *CollapsibleScope) finish(c *ui.Context) {
	if !s.trigger.Valid() {
		panic("display.CollapsibleRoot: declare a Trigger in the build callback")
	}
	if !s.hasContent {
		return
	}
	if *s.open {
		s.state.mounted = true
	}
	if !*s.open && s.state.panel.FocusWithin(c) {
		s.trigger.Focus()
	}
	progress := s.part.Progress()
	mount := s.state.mounted && (*s.open || progress > 0 || s.props.KeepMounted)
	fullyClosed := !*s.open && progress <= 0
	panel, inner := s.content.Root, s.content.Content
	panel.Hidden(fullyClosed).AccessHidden(!*s.open).IgnoreInput(!*s.open).Disabled(!*s.open).DimDisabled(false)
	s.trigger.Controls(panel)
	if !mount {
		return
	}
	s.carrier.Children(func() {
		inner.Children(func() {
			if s.build != nil {
				s.build(s.contentCtx)
			}
		})
	})
	height := ui.Local(inner, "natural-height", func() float32 { return 0 })
	// A newly opened panel can clip all its pixels at progress zero, so Draw
	// may not run yet. Native layout still measures the non-shrinking inner;
	// use its published previous-frame box to seed the next animation frame.
	if previous := inner.LayoutBounds().H; previous > 0 {
		*height = previous
	}
	current := s.contentCtx
	inner.Draw(func(_ *ui.Painter, r ui.Rect) {
		if *height != r.H {
			*height = r.H
			current.Invalidate()
		}
	})
	if progress < 1 {
		panel.Height(max(0, progress) * *height)
	}
}
