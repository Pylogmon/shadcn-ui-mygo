package control

import (
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var check = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="m5 12 4 4L19 6"/></svg>`))
var minus = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><path d="M5 12h14"/></svg>`))

type CheckboxProps struct{ Disabled, Indeterminate bool }

func Checkbox(c *ui.Context, checked *bool, label string, p CheckboxProps) ui.Element {
	t := theme.TokensOf(c)
	state := ui.CheckOff
	if *checked {
		state = ui.CheckOn
	}
	if p.Indeterminate {
		state = ui.CheckMixed
	}
	root := ui.CheckboxStateBase(c, &state).Gap(8).AlignItems(ui.Center).
		Disabled(p.Disabled).Cursor(ui.CursorPointer).FontSize(14).TextColor(t.Foreground)
	if root.Changed() {
		*checked = state == ui.CheckOn
	}
	if p.Disabled {
		root.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	root.Children(func() {
		bg, border := t.Background, t.Input
		if state != ui.CheckOff {
			bg, border = t.Primary, t.Primary
		}
		ui.Box(c).Size(16, 16).Radius(4).Border(1, border).Background(bg).Center().
			TextColor(t.PrimaryForeground).Transition(motion.Transition(c)).Children(func() {
			if state == ui.CheckOn {
				ui.Icon(c, check).FontSize(12)
			}
			if state == ui.CheckMixed {
				ui.Icon(c, minus).FontSize(12)
			}
		})
		if label != "" {
			ui.Text(c, label).SingleLine()
		}
	})
	return root
}

type SwitchProps struct {
	Label    string
	Disabled bool
}

func Switch(c *ui.Context, on *bool, p SwitchProps) ui.Element {
	t := theme.TokensOf(c)
	e := ui.SwitchBase(c, on).Gap(8).AlignItems(ui.Center).Disabled(p.Disabled).
		FontSize(14).TextColor(t.Foreground).Cursor(ui.CursorPointer)
	if p.Label != "" {
		e.Label(p.Label)
	}
	bg := t.Input
	x := float32(0)
	if *on {
		bg, x = t.Primary, 14
	}
	if p.Disabled {
		e.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	e.Children(func() {
		ui.Box(c).Size(32, 18).Radius(100).Padding(2).Background(bg).Shrink(0).
			Transition(motion.Transition(c)).Children(func() {
			thumb := ui.Box(c).Size(14, 14).Radius(100).Background(t.Background)
			thumb.Translate(thumb.AnimateTo("thumb", x, motion.Thumb(c)), 0)
		})
		if p.Label != "" {
			ui.Text(c, p.Label).SingleLine()
		}
	})
	return e
}

// Option is a stable value/label pair used by choice controls.
type Option struct {
	Value, Label string
	Disabled     bool
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
	Label                string
	Options              []Option
	Horizontal, Disabled bool
}

func RadioGroup(c *ui.Context, selected *string, p RadioGroupProps) ui.Element {
	g := ui.RadioGroup(c, func() {
		for _, o := range p.Options {
			if o.Separator {
				continue
			}
			Radio(c.Key(o.Value), selected, o.Value, optionLabel(o), o.Disabled || p.Disabled)
		}
	}).Label(p.Label).Gap(10).Disabled(p.Disabled)
	if p.Horizontal {
		g.Row().Gap(16)
	}
	return g
}

func Radio(c *ui.Context, selected *string, value, label string, disabled bool) ui.Element {
	t := theme.TokensOf(c)
	e := ui.RadioBase(c, selected, value).Gap(8).AlignItems(ui.Center).Disabled(disabled).
		Cursor(ui.CursorPointer).FontSize(14).TextColor(t.Foreground)
	if disabled {
		e.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	e.Children(func() {
		ui.Box(c).Size(16, 16).Radius(8).Border(1, t.Primary).Center().Children(func() {
			if *selected == value {
				ui.Box(c).Size(8, 8).Radius(4).Background(t.Primary)
			}
		})
		if label != "" {
			ui.Text(c, label).SingleLine()
		}
	})
	return e
}

type ToggleGroupProps struct {
	Label                       string
	Items                       []Option
	Multiple, Outline, Disabled bool
	Size                        Size
}

type toggleValue struct{ On, Last bool }

func ToggleGroup(c *ui.Context, selected *[]string, p ToggleGroupProps) ui.Element {
	root := ui.Row(c).Gap(2).AlignSelf(ui.Start).FocusGroup(ui.Horizontal).Label(p.Label).Disabled(p.Disabled)
	states := make([]*toggleValue, len(p.Items))
	for i, item := range p.Items {
		on := slices.Contains(*selected, item.Value)
		st := ui.Local(root, "toggle:"+item.Value, func() toggleValue { return toggleValue{on, on} })
		if st.On != st.Last {
			if st.On {
				if !p.Multiple {
					*selected = []string{item.Value}
				} else if !on {
					*selected = append(*selected, item.Value)
				}
			} else {
				*selected = slices.DeleteFunc(*selected, func(value string) bool { return value == item.Value })
			}
		}
		states[i] = st
	}
	root.Children(func() {
		for i, item := range p.Items {
			st := states[i]
			st.On = slices.Contains(*selected, item.Value)
			st.Last = st.On
			Toggle(c.Key(item.Value), &st.On, optionLabel(item), ToggleProps{Size: p.Size, Outline: p.Outline, Disabled: p.Disabled || item.Disabled})
		}
	})
	return root
}
