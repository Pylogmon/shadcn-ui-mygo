package control

import (
	"math"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var check = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="m5 12 4 4L19 6"/></svg>`))
var minus = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><path d="M5 12h14"/></svg>`))

type choiceSlot uint8

const (
	choiceIndicator choiceSlot = iota
	choiceCaption
	choiceToggleContent
	choiceToggleIcon
	choiceToggleEndIcon
)

type choiceOptionKey struct {
	value      string
	occurrence int
}

type switchDirectionKey struct{}

type CheckboxProps struct {
	Disabled, Indeterminate, Invalid bool
	Description                      string
	// Content replaces the visual label/description. It is non-interactive;
	// label and Description still name the actual native checkbox.
	Content         func(*ui.Context)
	OnCheckedChange func(bool)
	OnStateChange   func(ui.CheckState)
}

// Checkbox preserves the boolean API. Indeterminate is a caller-controlled
// display flag; OnStateChange can clear it after activation. CheckboxState
// supplies a single fully controlled native three-state value instead.
func Checkbox(c *ui.Context, checked *bool, label string, p CheckboxProps) ui.Element {
	state := ui.CheckOff
	if *checked {
		state = ui.CheckOn
	}
	if p.Indeterminate {
		state = ui.CheckMixed
	}
	q := p
	q.OnCheckedChange, q.OnStateChange = nil, nil
	root := CheckboxState(c, &state, label, q)
	if root.Changed() {
		*checked = state == ui.CheckOn
		if p.OnCheckedChange != nil {
			p.OnCheckedChange(*checked)
		}
		if p.OnStateChange != nil {
			p.OnStateChange(state)
		}
	}
	return root
}

// CheckboxState is a real native mixed checkbox. Mixed activation chooses On;
// another activation chooses Off. Props.Indeterminate only applies to Checkbox.
func CheckboxState(c *ui.Context, state *ui.CheckState, label string, p CheckboxProps) ui.Element {
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	root := ui.CheckboxStateBase(c, state)
	rich := choiceControl(root, label, p.Description, p.Disabled, p.Invalid, p.Content != nil)
	t := theme.TokensOf(c)
	root.Children(func() {
		visual := ui.Box(c.Key(choiceIndicator)).Size(16, 16).Shrink(0).Radius(4).Center().IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false)
		visual.Shadow(0, 1, 2, 0, ui.RGB(0, 0, 0).Alpha(.05))
		if rich {
			visual.Margin(2, 0, 0, 0)
		}
		choiceIndicatorStyle(c, root, visual, *state != ui.CheckOff, p.Invalid)
		visual.TextColor(t.PrimaryForeground).Children(func() {
			if *state == ui.CheckOn {
				ui.Icon(c, check).Size(14, 14)
			} else if *state == ui.CheckMixed {
				ui.Icon(c, minus).Size(14, 14)
			}
		})
		choiceCaptionContent(c, label, p.Description, p.Content)
	})
	if root.Changed() {
		if p.OnCheckedChange != nil {
			p.OnCheckedChange(*state == ui.CheckOn)
		}
		if p.OnStateChange != nil {
			p.OnStateChange(*state)
		}
	}
	return root
}

// choiceControl keeps the native control as the only keyboard/AX target. The
// visual children can be rich labels without creating nested native controls.
func choiceControl(root ui.Element, label, description string, disabled, invalid, custom bool) bool {
	root.AlignSelf(ui.Start).MinWidth(0).MaxWidthPercent(100).Gap(8).AlignItems(ui.Center).
		Label(label).Description(description).Disabled(disabled).DimDisabled(false).Invalid(invalid).
		FocusRing(false).Cursor(ui.CursorPointer).FontSize(14).LineHeight(1.375)
	rich := description != "" || custom
	if rich {
		root.AlignItems(ui.Start)
	}
	if root.IsDisabled() {
		root.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	return rich
}
func choiceCaptionContent(c *ui.Context, label, description string, content func(*ui.Context)) {
	if label == "" && description == "" && content == nil {
		return
	}
	ui.Column(c.Key(choiceCaption)).MinWidth(0).Gap(4).IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false).Children(func() {
		if content != nil {
			content(c)
			return
		}
		if label != "" {
			ui.Text(c, label)
		}
		if description != "" {
			ui.Text(c, description).FontSize(12).LineHeight(1.5).TextColor(theme.TokensOf(c).MutedForeground)
		}
	})
}
func choiceIndicatorStyle(c *ui.Context, control, visual ui.Element, checked, invalid bool) {
	t := theme.TokensOf(c)
	background, border := ui.Transparent, t.Input
	if t.Dark {
		background = t.Input.Alpha(.3)
	}
	if checked {
		background, border = t.Primary, t.Primary
	}
	if control.FocusVisible() {
		border = t.Ring
		visual.Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
	}
	if invalid {
		border = t.Destructive
		alpha := float32(.2)
		if t.Dark {
			border, alpha = t.Destructive.Alpha(.5), .4
		} else if checked {
			border = t.Primary
		}
		visual.Shadow(0, 0, 0, 3, t.Destructive.Alpha(alpha))
	}
	visual.Background(background).Border(1, border)
	motion.ApplyTransition(c, visual)
}

type SwitchProps struct {
	Label, Description string
	Disabled, Invalid  bool
	// Small maps to the source's 24×14-DIP switch; other sizes use 32×18.4.
	Size            Size
	Content         func(*ui.Context)
	OnCheckedChange func(bool)
}

func Switch(c *ui.Context, on *bool, p SwitchProps) ui.Element {
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	t := theme.TokensOf(c)
	root := ui.SwitchBase(c, on)
	rich := choiceControl(root, p.Label, p.Description, p.Disabled, p.Invalid, p.Content != nil)
	width, height, diameter := float32(32), float32(18.4), float32(16)
	if p.Size == Small {
		width, height, diameter = 24, 14, 12
	}
	background, knob := t.Input, t.Background
	if t.Dark {
		background, knob = t.Input.Alpha(.8), t.Foreground
	}
	x := float32(0)
	if *on {
		background = t.Primary
		if t.Dark {
			knob = t.PrimaryForeground
		}
		x = width - diameter - 2
	}
	root.Children(func() {
		track := ui.Row(c.Key(choiceIndicator)).Size(width, height).Shrink(0).Radius(999).AlignItems(ui.Center).
			Border(1, ui.Transparent).Background(background).IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false)
		motion.ApplyTransition(c, track)
		track.Shadow(0, 1, 2, 0, ui.RGB(0, 0, 0).Alpha(.05))
		if rich {
			track.Margin(2, 0, 0, 0)
		}
		if root.FocusVisible() {
			track.Border(1, t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
		}
		if p.Invalid {
			border, alpha := t.Destructive, float32(.2)
			if t.Dark {
				border, alpha = border.Alpha(.5), .4
			}
			track.Border(1, border).Shadow(0, 0, 0, 3, t.Destructive.Alpha(alpha))
		}
		track.Children(func() {
			thumb := ui.Box(c).Size(diameter, diameter).Shrink(0).Radius(999).Background(knob)
			direction := c.Direction()
			if direction == ui.RightToLeft {
				x = -x
			}
			previousDirection := ui.Local(thumb, switchDirectionKey{}, func() ui.Direction { return direction })
			if *previousDirection != direction {
				// Layout mirrors the logical starting edge immediately. Mirror
				// the offset too rather than animating from the old physical edge.
				thumb.Jump("thumb", x)
				*previousDirection = direction
			}
			position := x
			if motion.Reduced(c) {
				position = thumb.Jump("thumb", x)
			} else {
				position = thumb.AnimateTo("thumb", x, motion.Thumb(c))
			}
			thumb.Translate(position, 0)
		})
		choiceCaptionContent(c, p.Label, p.Description, p.Content)
	})
	if root.Changed() && p.OnCheckedChange != nil {
		p.OnCheckedChange(*on)
	}
	return root
}

// Option is a stable value/label pair used by choice controls.
type Option struct {
	Value, Label, Description string
	Disabled                  bool
	// Group labels adjacent options. Separator inserts a non-selectable rule.
	Group     string
	Separator bool
}

func optionLabel(o Option) string {
	if o.Label == "" {
		return o.Value
	}
	return o.Label
}

type RadioGroupProps struct {
	Label, Description   string
	Options              []Option
	Horizontal, Disabled bool
	Invalid, NoLoop      bool
	// Zero preserves native radio's four-arrow navigation. An explicit axis
	// constrains arrows and supplies native orientation semantics.
	Orientation   ui.Orientation
	ItemContent   func(*ui.Context, Option, bool)
	OnValueChange func(string)
}

func RadioGroup(c *ui.Context, selected *string, p RadioGroupProps) ui.Element {
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	before := *selected
	direction := c.Direction()
	axis := p.Orientation
	if axis == 0 {
		axis = ui.Horizontal | ui.Vertical
	}
	occurrences := map[string]int{}
	g := ui.RadioGroup(c, func() {
		for i, option := range p.Options {
			if option.Separator {
				continue
			}
			key := choiceOptionKey{option.Value, occurrences[option.Value]}
			occurrences[option.Value]++
			props := RadioProps{Disabled: option.Disabled || p.Disabled, Invalid: p.Invalid, Description: option.Description}
			if p.ItemContent != nil {
				props.Content = func(child *ui.Context) { p.ItemContent(child, option, *selected == option.Value) }
			}
			radio := RadioWith(c.Key(key), selected, option.Value, optionLabel(option), props)
			if p.NoLoop {
				radio.HandleInput(func(event ui.InputEvent) bool { return choiceStopsLoop(event, i, p.Options, axis, direction) })
			}
		}
	}).Label(p.Label).Description(p.Description).Gap(12).Disabled(p.Disabled).Invalid(p.Invalid).FocusGroup(axis)
	if p.Horizontal || p.Orientation == ui.Horizontal {
		g.Row().Gap(16)
	}
	if p.Orientation == ui.Horizontal || p.Orientation == ui.Vertical {
		g.Orientation(p.Orientation)
	}
	if before != *selected {
		g.SetChanged()
		if p.OnValueChange != nil {
			p.OnValueChange(*selected)
		}
	}
	return g
}

type RadioProps struct {
	Description       string
	Disabled, Invalid bool
	Content           func(*ui.Context)
}

func Radio(c *ui.Context, selected *string, value, label string, disabled bool) ui.Element {
	return RadioWith(c, selected, value, label, RadioProps{Disabled: disabled})
}
func RadioWith(c *ui.Context, selected *string, value, label string, p RadioProps) ui.Element {
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	t := theme.TokensOf(c)
	root := ui.RadioBase(c, selected, value)
	rich := choiceControl(root, label, p.Description, p.Disabled, p.Invalid, p.Content != nil)
	root.Children(func() {
		visual := ui.Box(c.Key(choiceIndicator)).Size(16, 16).Shrink(0).Radius(8).Center().IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false)
		if rich {
			visual.Margin(2, 0, 0, 0)
		}
		choiceIndicatorStyle(c, root, visual, *selected == value, p.Invalid)
		visual.Children(func() {
			if *selected == value {
				ui.Box(c).Size(8, 8).Radius(4).Background(t.PrimaryForeground)
			}
		})
		choiceCaptionContent(c, label, p.Description, p.Content)
	})
	return root
}

type ToggleProps struct {
	Size                       Size
	Outline, Disabled, Invalid bool
	Label, Description         string
	Icon, EndIcon              *ui.SVG
	Content                    func(*ui.Context)
	OnPressedChange            func(bool)
}

func Toggle(c *ui.Context, on *bool, label string, p ToggleProps) ui.Element {
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	t := theme.TokensOf(c)
	root := ui.ToggleBase(c, on)
	height, padding := metrics(p.Size)
	name := label
	if p.Label != "" {
		name = p.Label
	}
	root.AlignSelf(ui.Start).MinWidth(height).Height(height).PaddingX(padding).Gap(4).Radius(t.RadiusMD()).AlignItems(ui.Center).Justify(ui.Center).
		FontSize(14).FontWeight(500).Label(name).Description(p.Description).TextColor(t.Foreground).Background(ui.Transparent).Border(0, ui.Transparent).
		Disabled(p.Disabled).DimDisabled(false).Invalid(p.Invalid).FocusRing(false).Cursor(ui.CursorPointer)
	motion.ApplyTransition(c, root)
	if p.Outline {
		root.Border(1, t.Input).Shadow(0, 1, 2, 0, ui.RGB(0, 0, 0).Alpha(.05))
	}
	if p.Icon != nil && label != "" {
		edge := float32(8)
		if p.Size == Small {
			edge = 6
		}
		root.PaddingStart(edge)
	}
	if p.EndIcon != nil && label != "" {
		edge := float32(8)
		if p.Size == Small {
			edge = 6
		}
		root.PaddingEnd(edge)
	}
	if *on || root.Hovered() && !root.IsDisabled() {
		root.Background(t.Muted)
	}
	if root.FocusVisible() {
		root.BorderColor(t.Ring).Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
	}
	if p.Invalid {
		alpha := float32(.2)
		if t.Dark {
			alpha = .4
		}
		root.BorderColor(t.Destructive).Shadow(0, 0, 0, 3, t.Destructive.Alpha(alpha))
	}
	if root.IsDisabled() {
		root.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	root.Children(func() {
		ui.Row(c.Key(choiceToggleContent)).Gap(4).AlignItems(ui.Center).IgnoreInput(true).AccessHidden(true).Disabled(true).DimDisabled(false).Children(func() {
			if p.Content != nil {
				p.Content(c)
				return
			}
			if p.Icon != nil {
				ui.Icon(c.Key(choiceToggleIcon), p.Icon).Size(16, 16)
			}
			if label != "" {
				ui.Text(c, label).SingleLine()
			}
			if p.EndIcon != nil {
				ui.Icon(c.Key(choiceToggleEndIcon), p.EndIcon).Size(16, 16)
			}
		})
	})
	if root.Changed() && p.OnPressedChange != nil {
		p.OnPressedChange(*on)
	}
	return ButtonGroupItem(c, root)
}

type ToggleGroupProps struct {
	Label                       string
	Items                       []Option
	Multiple, Outline, Disabled bool
	Size                        Size
	Vertical, NoLoop, Invalid   bool
	// Spacing is in DIPs. Nil uses the source's 2×4-DIP gap; zero joins items.
	Spacing       *float32
	ItemContent   func(*ui.Context, Option, bool)
	OnValueChange func([]string)
}

func ToggleGroup(c *ui.Context, selected *[]string, p ToggleGroupProps) ui.Element {
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	axis := ui.Horizontal
	if p.Vertical {
		axis = ui.Vertical
	}
	gap := float32(8)
	if p.Spacing != nil && *p.Spacing >= 0 && !math.IsNaN(float64(*p.Spacing)) && !math.IsInf(float64(*p.Spacing), 0) {
		gap = *p.Spacing
	}
	root := ui.Row(c).Gap(gap).AlignSelf(ui.Start).FocusGroup(axis).Orientation(axis).Role(ui.RoleGroup).Label(p.Label).Disabled(p.Disabled).Invalid(p.Invalid)
	if p.Vertical {
		root.Column().AlignItems(ui.Stretch)
	}
	direction := c.Direction()
	var controls []ui.Element
	var values []string
	changed := false
	occurrences := map[string]int{}
	root.Children(func() {
		for i, item := range p.Items {
			if item.Separator {
				continue
			}
			on := slices.Contains(*selected, item.Value)
			key := choiceOptionKey{item.Value, occurrences[item.Value]}
			occurrences[item.Value]++
			props := ToggleProps{Size: p.Size, Outline: p.Outline, Disabled: p.Disabled || item.Disabled, Invalid: p.Invalid, Description: item.Description}
			if p.ItemContent != nil {
				props.Content = func(child *ui.Context) { p.ItemContent(child, item, on) }
			}
			button := Toggle(c.Key(key), &on, optionLabel(item), props)
			if p.NoLoop {
				button.HandleInput(func(event ui.InputEvent) bool { return choiceStopsLoop(event, i, p.Items, axis, direction) })
			}
			if p.Vertical {
				button.AlignSelf(ui.Stretch)
			}
			if button.Changed() {
				changed = true
				if !p.Multiple {
					*selected = nil
					if on {
						*selected = []string{item.Value}
					}
				} else if on {
					if !slices.Contains(*selected, item.Value) {
						*selected = append(*selected, item.Value)
					}
				} else {
					*selected = slices.DeleteFunc(*selected, func(value string) bool { return value == item.Value })
				}
			}
			controls, values = append(controls, button), append(values, item.Value)
		}
	})
	for i, button := range controls {
		button.Checked(slices.Contains(*selected, values[i]))
	}
	if gap == 0 {
		joinChoiceButtons(c, controls, p.Vertical, p.Outline)
	}
	if changed {
		root.SetChanged()
		c.Invalidate()
		if p.OnValueChange != nil {
			p.OnValueChange(slices.Clone(*selected))
		}
	}
	return root
}

// The real native FocusGroup handles ordinary navigation, disabled skipping,
// Home/End and Tab. Intercept only an attempted endpoint wrap when requested.
func choiceStopsLoop(event ui.InputEvent, index int, items []Option, axis ui.Orientation, direction ui.Direction) bool {
	if event.Kind != ui.InputKeyDown || event.Mods != 0 {
		return false
	}
	first, last := -1, -1
	for i, item := range items {
		if !item.Disabled && !item.Separator {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	back, forth := ui.KeyLeft, ui.KeyRight
	if direction == ui.RightToLeft {
		back, forth = forth, back
	}
	previous := event.Key == back && axis&ui.Horizontal != 0 || event.Key == ui.KeyUp && axis&ui.Vertical != 0
	next := event.Key == forth && axis&ui.Horizontal != 0 || event.Key == ui.KeyDown && axis&ui.Vertical != 0
	return previous && index == first || next && index == last
}
func joinChoiceButtons(c *ui.Context, buttons []ui.Element, vertical, outline bool) {
	radius := theme.TokensOf(c).RadiusMD()
	for i, button := range buttons {
		first, last := i == 0, i == len(buttons)-1
		corners := [4]float32{}
		if vertical {
			if first {
				corners[0], corners[1] = radius, radius
			}
			if last {
				corners[2], corners[3] = radius, radius
			}
			if outline && !first {
				button.BorderWidth(0, 1, 1, 1)
			}
		} else {
			if c.Direction() == ui.RightToLeft {
				first, last = last, first
			}
			if first {
				corners[0], corners[3] = radius, radius
			}
			if last {
				corners[1], corners[2] = radius, radius
			}
			if outline && i > 0 {
				if c.Direction() == ui.RightToLeft {
					button.BorderWidth(1, 0, 1, 1)
				} else {
					button.BorderWidth(1, 1, 1, 0)
				}
			}
		}
		button.Radius(corners[0], corners[1], corners[2], corners[3]).PaddingX(8)
		if button.FocusVisible() {
			button.ZIndex(1)
		}
	}
}
