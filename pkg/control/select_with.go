package control

import "github.com/egoist/mygo/ui"

// SelectWithProps is the standard shadcn template over independent Select slots.
// ItemContent and ValueContent build non-interactive native visuals. The select
// owns listbox activation, accessible labels, focus and selected indicators.
type SelectWithProps struct {
	Root         SelectRootProps
	Trigger      SelectTriggerProps
	Content      SelectContentProps
	Placeholder  string
	ItemContent  func(*ui.Context, Option)
	ValueContent func(*ui.Context, SelectValueView)
}

func SelectWith(c *ui.Context, state *SelectState, p SelectWithProps) ui.Element {
	var trigger ui.Element
	SelectRoot(c, state, p.Root, func(b *SelectScope) {
		trigger = selectTemplate(c, b, p)
	})
	return trigger
}

func selectTemplate(c *ui.Context, b *SelectScope, p SelectWithProps) ui.Element {
	trigger := b.Trigger(c, p.Trigger, func() {
		var content func(SelectValueView)
		if p.ValueContent != nil {
			content = func(v SelectValueView) { p.ValueContent(c, v) }
		}
		b.Value(c, SelectValueProps{Placeholder: p.Placeholder}, content)
	})
	b.Content(c, p.Content, func() {
		b.List(c, SelectListProps{}, func() {
			options := b.f.options
			for i := 0; i < len(options); {
				if options[i].Separator {
					b.Separator(c)
					i++
					continue
				}
				start, group := i, options[i].Group
				for i < len(options) && !options[i].Separator && options[i].Group == group {
					i++
				}
				end := i
				b.Group(c.Key(start), group, func() {
					for _, option := range options[start:end] {
						var content func()
						if p.ItemContent != nil {
							content = func() { p.ItemContent(c, option) }
						}
						b.Item(c, option.Value, SelectItemProps{}, content)
					}
				})
			}
		})
		b.ScrollUpButton(c)
		b.ScrollDownButton(c)
	})
	return trigger
}

// MultiSelect is a no-editor, multiple-selection listbox using the same native
// slots as SelectWith. Selections stay open and toggle without editing a query.
// NativeSelect remains the operating system's single-selection menu adapter.
func MultiSelect(c *ui.Context, selected *[]string, p SelectProps) ui.Element {
	if selected == nil {
		panic("control.MultiSelect: selected is required")
	}
	state := SelectState{Values: *selected}
	var trigger ui.Element
	selectRoot(c, nil, &state, SelectRootProps{Options: p.Options, Multiple: true, Disabled: p.Disabled}, true, true, func(b *SelectScope) {
		trigger = selectTemplate(c, b, SelectWithProps{
			Root:    SelectRootProps{Options: p.Options, Multiple: true, Disabled: p.Disabled},
			Trigger: SelectTriggerProps{Label: p.Label, Size: p.Size, Invalid: p.Invalid},
			Content: p.Content, Placeholder: p.Placeholder, ItemContent: p.ItemContent,
			ValueContent: p.ValuesContent,
		})
	})
	*selected = state.Values
	return trigger
}
