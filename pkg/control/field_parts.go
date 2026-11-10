package control

import (
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// FieldScope is valid only during FieldWith's build callback. Declare the
// visible parts in any order and register the actual editor or choice with
// Control. Relationships are applied after every part has been built.
type FieldScope struct {
	c                *ui.Context
	root             ui.Element
	p                FieldProps
	error            fieldErrorContent
	errorNode        ui.Element
	label            ui.Element
	target           ui.Element
	described        []ui.Element
	choices          []ui.Element
	alive            bool
	horizontal       bool
	hasContent       bool
	errorBuilt       bool
	descriptionBuilt bool
}

// FieldWith composes a Field's native parts without prescribing their order.
// Unlike the legacy convenience Field, a horizontal Content groups the label
// and description together, alongside its Control. Props.Label supplies a
// fallback accessible name when the composition deliberately omits a label.
// No Context, Element, scope or builder is retained between builds.
func FieldWith(c *ui.Context, p FieldProps, build func(*FieldScope)) ui.Element {
	messages := p.Errors
	if p.Error != "" {
		messages = append([]string{p.Error}, messages...)
	}
	f := &FieldScope{c: c, p: p, error: resolveFieldError(FieldErrorProps{Errors: messages, Label: p.ErrorLabel, Content: p.ErrorContent}), alive: true}
	f.horizontal = fieldIsHorizontal(c, p)
	f.root = ui.Column(c).FillWidth().MinWidth(0).Gap(12).Role(ui.RoleGroup).Disabled(p.Disabled).DimDisabled(false)
	if f.horizontal {
		f.root.Row().AlignItems(ui.Center)
	}
	if p.Invalid || f.error.present() {
		f.root.TextColor(theme.TokensOf(c).Destructive)
	}
	defer func() { f.alive = false }()
	f.root.Children(func() {
		fieldEditor(c, p.Invalid || f.error.present(), func() ui.Element {
			if build != nil {
				build(f)
			}
			if !f.descriptionBuilt && p.Description != "" {
				f.Description(c, p.Description)
			}
			if !f.errorBuilt {
				f.Error(c)
			}
			return f.root
		})
		f.finish()
	})
	return f.root
}

func fieldIsHorizontal(c *ui.Context, p FieldProps) bool {
	if p.Orientation != FieldResponsive {
		return p.Orientation == FieldHorizontal
	}
	width := p.ResponsiveWidth
	s := ui.Local(c.Root(), "shadcn.control.field-group", func() fieldGroupLayout { return fieldGroupLayout{} })
	if width <= 0 && s.active {
		width = s.width
	}
	if width <= 0 {
		width, _ = c.Size()
	}
	return width >= 448
}

func (f *FieldScope) check() {
	if f == nil || !f.alive {
		panic("control.FieldScope: parts are valid only during FieldWith's build callback")
	}
}

// Content groups a label and its helper text in one flexible column.
func (f *FieldScope) Content(c *ui.Context, build func()) ui.Element {
	f.check()
	f.hasContent = true
	return FieldContent(c, build)
}

// Control registers the composition's single native target. InputGroup and
// TextareaWith roots resolve to their actual editor using the same target
// contract as Field. Button-only layout fields may omit Control entirely.
func (f *FieldScope) Control(c *ui.Context, build func() ui.Element) ui.Element {
	f.check()
	if f.target.Valid() {
		panic("control.FieldScope: a field has one Control; compose multiple fields in FieldGroup")
	}
	if build == nil {
		return ui.Element{}
	}
	root := fieldEditor(c, f.p.Invalid || f.error.present(), build)
	f.target = fieldTarget(c, root)
	return root
}

// Label may appear before or after Control, or wrap a whole choice card.
// Rich children require Label (or Text) to name the actual native target.
func (f *FieldScope) Label(c *ui.Context, p FieldLabelProps, build func()) ui.Element {
	f.check()
	if p.Target.Valid() {
		panic("control.FieldScope: register the target with Control; Target is for standalone FieldLabel")
	}
	if f.label.Valid() {
		panic("control.FieldScope: a field has one Label")
	}
	p.Disabled = p.Disabled || f.p.Disabled
	if p.Text == "" && p.Label == "" && build == nil {
		p.Text = f.p.Label
	}
	if f.p.Required && build == nil && p.Text != "" {
		p.Text += " *"
	}
	label := FieldLabel(c, p, build)
	if f.horizontal && !p.Choice {
		label.Grow(1)
	}
	if f.p.Invalid || f.error.present() {
		label.TextColor(theme.TokensOf(c).Destructive)
	}
	f.label = label
	if p.Choice {
		f.choices = append(f.choices, label)
	}
	return label
}

func (f *FieldScope) Description(c *ui.Context, text string) ui.Element {
	f.check()
	f.descriptionBuilt = true
	e := FieldDescription(c, text)
	if e.Valid() {
		f.described = append(f.described, e)
	}
	return e
}

// DescriptionWith supports real native links or other rich helper content.
// Label is the complete accessible description; children are the visible UI.
func (f *FieldScope) DescriptionWith(c *ui.Context, label string, build func()) ui.Element {
	f.check()
	f.descriptionBuilt = true
	e := FieldDescriptionWith(c, label, build)
	if e.Valid() {
		f.described = append(f.described, e)
	}
	return e
}

// Error places the FieldProps error at the caller's chosen position. When
// omitted it is appended after the composition. Empty errors create no node.
func (f *FieldScope) Error(c *ui.Context) ui.Element {
	f.check()
	if f.errorBuilt {
		panic("control.FieldScope: Error may be declared once")
	}
	f.errorBuilt = true
	e := renderFieldError(c, f.error)
	f.errorNode = e
	if e.Valid() {
		f.described = append(f.described, e)
	}
	return e
}

func (f *FieldScope) finish() {
	if f.horizontal && f.hasContent {
		f.root.AlignItems(ui.Start)
	}
	if !f.target.Valid() {
		return
	}
	if f.label.Valid() {
		f.target.LabelledBy(f.label)
	} else if f.p.Label != "" {
		f.target.Label(f.p.Label)
	}
	if f.p.Required {
		f.target.Description("Required")
	}
	for _, d := range f.described {
		f.target.DescribedBy(d)
	}
	if f.p.Invalid {
		f.target.Invalid(true)
	}
	if f.error.present() {
		f.target.Invalid(true).ErrorMessage(f.errorNode)
	}
	for _, card := range f.choices {
		if f.target.FocusVisible() {
			t := theme.TokensOf(f.c)
			card.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
		}
	}
}

// FieldLabelProps explicitly supplies the pure selection state for a choice
// card. Native controls keep their own checked/selected semantics; this helper
// never creates a second checkbox/radio or a second Tab stop.
type FieldLabelProps struct {
	Text, Label                string
	Choice, Selected, Disabled bool
	// Target is optional for standalone labels and must belong to this build.
	// Within FieldWith use Control for order-independent association instead.
	Target ui.Element
}

func FieldContent(c *ui.Context, build func()) ui.Element {
	return ui.Column(c).Grow(1).MinWidth(0).Gap(4).LineHeight(1.375).Children(build)
}

func FieldLabel(c *ui.Context, p FieldLabelProps, build func()) ui.Element {
	name := p.Label
	if name == "" {
		name = p.Text
	}
	if name == "" && build != nil {
		panic("control.FieldLabel: rich content needs Label or Text")
	}
	t := theme.TokensOf(c)
	e := ui.Row(c).MinWidth(0).Gap(8).AlignItems(ui.Center).FontSize(14).FontWeight(500).LineHeight(1.375).Label(name).Disabled(p.Disabled).DimDisabled(false)
	if p.Choice {
		e.Column().FillWidth().Padding(12).Radius(t.RadiusMD()).Border(1, t.Border)
		if p.Selected {
			alpha, border := float32(.05), float32(.3)
			if t.Dark {
				alpha, border = .1, .2
			}
			e.Background(t.Primary.Alpha(alpha)).Border(1, t.Primary.Alpha(border))
		} else if e.Hovered() && !p.Disabled {
			e.Background(t.Muted.Alpha(.5))
		}
	}
	if p.Disabled {
		e.Opacity(.5)
	}
	e.Children(func() {
		if build != nil {
			build()
		} else if p.Text != "" {
			ui.Text(c, p.Text)
		}
	})
	if p.Target.Valid() {
		fieldTarget(c, p.Target).LabelledBy(e)
	}
	return e
}

func FieldTitle(c *ui.Context, text string) ui.Element {
	return ui.Text(c, text).MinWidth(0).FontSize(14).FontWeight(500).LineHeight(1.375)
}

func FieldDescription(c *ui.Context, text string) ui.Element {
	if text == "" {
		return ui.Element{}
	}
	return ui.Text(c, text).MinWidth(0).FontSize(14).FontWeight(400).LineHeight(1.5).TextColor(theme.TokensOf(c).MutedForeground)
}

func FieldDescriptionWith(c *ui.Context, label string, build func()) ui.Element {
	if build == nil {
		return FieldDescription(c, label)
	}
	if label == "" {
		panic("control.FieldDescriptionWith: rich content needs an accessible label")
	}
	return ui.Column(c).MinWidth(0).Gap(4).FontSize(14).FontWeight(400).LineHeight(1.5).TextColor(theme.TokensOf(c).MutedForeground).Label(label).Children(build)
}

type FieldLegendVariant uint8

const (
	FieldLegendDefault FieldLegendVariant = iota
	FieldLegendLabel
)

func FieldLegend(c *ui.Context, text string, variant ...FieldLegendVariant) ui.Element {
	font := float32(16)
	if len(variant) > 0 && variant[0] == FieldLegendLabel {
		font = 14
	}
	return ui.Text(c, text).FontSize(font).FontWeight(500).LineHeight(1.5).Margin(0, 0, 12, 0)
}

type FieldSetSlots struct {
	Legend, Description func(*ui.Context) ui.Element
	Content             func(*ui.Context)
}

// FieldSetWith allows native rich legend/description content while keeping
// the group named and described. Callbacks are consumed in this build only.
func FieldSetWith(c *ui.Context, p FieldSetProps, slots FieldSetSlots) ui.Element {
	root := ui.Column(c).FillWidth().MinWidth(0).Gap(24).Role(ui.RoleGroup).Label(p.Legend).Disabled(p.Disabled).DimDisabled(false)
	root.Children(func() {
		if slots.Legend != nil || slots.Description != nil || p.Legend != "" || p.Description != "" {
			// HTML's legend is outside the fieldset's flex formatting context.
			// A native header group preserves its 12px margin and the following
			// description's -6px adjustment without adding the content gap.
			ui.Column(c).FillWidth().MinWidth(0).Children(func() {
				var legend, description ui.Element
				if slots.Legend != nil {
					legend = slots.Legend(c)
				} else if p.Legend != "" {
					legend = FieldLegend(c, p.Legend)
				}
				if legend.Valid() {
					root.LabelledBy(legend)
				}
				if slots.Description != nil {
					description = slots.Description(c)
				} else if p.Description != "" {
					description = FieldDescription(c, p.Description)
				}
				if description.Valid() {
					if legend.Valid() {
						description.Margin(-6, 0, 0, 0)
					}
					root.DescribedBy(description)
				} else if legend.Valid() {
					// A lone HTML legend's margin is the only spacing before
					// the fieldset body; it is not also a flex-gap participant.
					root.Gap(0)
				}
			})
		}
		if slots.Content != nil {
			ui.Column(c).FillWidth().MinWidth(0).Gap(24).Children(func() { slots.Content(c) })
		}
	})
	return root
}

// FieldSeparator draws a decorative rule, optionally interrupted by real text.
func FieldSeparator(c *ui.Context, text string) ui.Element {
	t := theme.TokensOf(c)
	line := func() { ui.Box(c).Height(1).Grow(1).Background(t.Border).AccessHidden(true).IgnoreInput(true) }
	return ui.Row(c).FillWidth().MinWidth(0).Height(20).MarginY(-8).AlignItems(ui.Center).Children(func() {
		line()
		if text != "" {
			ui.Text(c, text).PaddingX(8).FontSize(14).FontWeight(400).TextColor(t.MutedForeground).Background(t.Background).Shrink(0)
			line()
		}
	})
}
