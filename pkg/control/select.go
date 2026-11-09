package control

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type SelectProps struct {
	Options            []Option
	Placeholder, Label string
	Size               Size
	Disabled, Invalid  bool
	// ItemContent customizes option content without replacing native option
	// behavior or its accessible text label.
	ItemContent    func(*ui.Context, Option)
	TriggerContent func(*ui.Context, Option)
}

func Select(c *ui.Context, selected *string, p SelectProps) ui.Element {
	t := theme.TokensOf(c)
	parts := ui.SelectBase(c, selected)
	tr := parts.Trigger
	h, px := metrics(p.Size)
	motion.ApplyTransition(c, tr.MinHeight(h).PaddingX(min(px, 12)).Gap(8).AlignItems(ui.Center).Radius(t.Radius*.8).
		Background(t.Background).TextColor(t.Foreground).FontSize(14).Border(1, t.Input).
		Disabled(p.Disabled).Invalid(p.Invalid).Cursor(ui.CursorPointer))
	if p.Label != "" {
		tr.Label(p.Label)
	}
	if p.Disabled {
		tr.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	if p.Invalid {
		tr.Border(1, t.Destructive)
	}
	label := p.Placeholder
	chosen := false
	for _, o := range p.Options {
		if !o.Separator && o.Value == *selected {
			label, chosen = optionLabel(o), true
			break
		}
	}
	if label == "" {
		label = "Select an option"
	}
	tr.Children(func() {
		if chosen && p.TriggerContent != nil {
			ui.Row(c).Grow(1).MinWidth(0).Children(func() {
				for _, o := range p.Options {
					if !o.Separator && o.Value == *selected {
						p.TriggerContent(c, o)
						break
					}
				}
			})
		} else {
			text := ui.Text(c, label).Grow(1).SingleLine()
			if !chosen {
				text.TextColor(t.MutedForeground)
			}
		}
		indicator.ChevronDown(c).TextColor(t.MutedForeground)
	})
	parts.Popup(func(panel ui.Element) {
		menu := theme.MenuOf(c).ForScene()
		t := &menu.Tokens
		panel.MinWidth(max(float32(128), tr.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).
			Radius(t.Radius*.8).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).
			Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Transition(motion.Menu(c))
		menu.Surface(c, panel)
		panel.Children(func() {
			ui.Themed(c, menu.Theme()).FillWidth().Children(func() {
				ui.Scroll(c).MaxHeight(272).FillWidth().Children(func() {
					group := ""
					for _, o := range p.Options {
						if o.Separator {
							optionSeparator(c)
							continue
						}
						if o.Group != "" && o.Group != group {
							group = o.Group
							optionGroup(c, group)
						}
						e := parts.Item(o.Value).Disabled(o.Disabled).MinHeight(32).Padding(6, 8).Gap(8).
							Radius(t.Radius * .6).AlignItems(ui.Center).AlignSelf(ui.Stretch).FontSize(14).Label(optionLabel(o))
						if e.Highlighted() {
							bg, fg := menu.Highlight(false)
							e.Background(bg).TextColor(fg)
						}
						if o.Disabled {
							e.Opacity(.5).Cursor(ui.CursorNotAllowed)
						}
						e.Children(func() {
							if p.ItemContent != nil {
								ui.Row(c).Grow(1).MinWidth(0).Children(func() { p.ItemContent(c, o) })
							} else {
								ui.Text(c, optionLabel(o)).Grow(1).SingleLine()
							}
							if *selected == o.Value {
								ui.Icon(c, check).FontSize(14)
							}
						})
					}
				})
			})
		})
	})
	return ButtonGroupItem(c, tr)
}

// NativeSelect uses MyGo's system-menu adapter instead of SelectBase's in-window
// listbox. The native OS owns popup placement, keyboard navigation and painting,
// and can display beyond the parent window. It explicitly selects system menus
// for this control without changing its caller's menu policy. Native options use text, groups and
// selected marks; arbitrary ItemContent belongs to the styled Select instead.
func NativeSelect(c *ui.Context, selected *string, p SelectProps) ui.Element {
	t := theme.TokensOf(c)
	h, px := metrics(p.Size)
	tr := motion.ApplyTransition(c, ui.ButtonBase(c).MinHeight(h).PaddingX(min(px, 12)).Gap(8).AlignItems(ui.Center).Radius(t.RadiusMD()).Background(t.Background).TextColor(t.Foreground).FontSize(14).Border(1, t.Input).Disabled(p.Disabled).Invalid(p.Invalid).Cursor(ui.CursorPointer))
	if p.Disabled {
		tr.Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	if p.Invalid {
		tr.Border(1, t.Destructive)
	}
	label := p.Placeholder
	var chosen Option
	found := false
	for _, o := range p.Options {
		if !o.Separator && o.Value == *selected {
			label = optionLabel(o)
			chosen = o
			found = true
			break
		}
	}
	if label == "" {
		label = "Select an option"
	}
	if p.Label != "" {
		tr.Label(p.Label)
	}
	tr.Value(label).Children(func() {
		if found && p.TriggerContent != nil {
			ui.Row(c).Grow(1).MinWidth(0).Children(func() { p.TriggerContent(c, chosen) })
		} else {
			text := ui.Text(c, label).Grow(1).SingleLine()
			if !found {
				text.TextColor(t.MutedForeground)
			}
		}
		indicator.ChevronDown(c).TextColor(t.MutedForeground)
	})
	previousPresentation := c.MenuPresentation()
	c.SetMenuPresentation(ui.SystemMenus)
	tr.Menu(func(m *ui.Menu) {
		group := ""
		for _, o := range p.Options {
			if o.Separator {
				m.Separator()
				continue
			}
			if o.Group != "" && o.Group != group {
				if group != "" {
					m.Separator()
				}
				group = o.Group
				m.Item(group).Disabled(true)
			}
			if m.Item(optionLabel(o)).ID("option:" + o.Value).Disabled(o.Disabled || p.Disabled).Radio(*selected == o.Value).Chosen() {
				if *selected != o.Value {
					*selected = o.Value
					tr.SetChanged()
				}
			}
		}
	})
	c.SetMenuPresentation(previousPresentation)
	return ButtonGroupItem(c, tr)
}

type ComboboxProps struct {
	Options                       []Option
	Placeholder, Label, EmptyText string
	Disabled, Invalid             bool
	ItemContent                   func(*ui.Context, Option)
}

// Combobox supports free text and filtered value/label suggestions. A chosen option writes its Value.
func Combobox(c *ui.Context, value *string, p ComboboxProps) ui.Element {
	parts := ui.ComboboxBase(c, value)
	input := styleInput(c, parts.Input, InputProps{Placeholder: p.Placeholder, Label: p.Label, Disabled: p.Disabled, Invalid: p.Invalid}).AutoComplete(ui.AutoCompleteList)
	composition := ui.Local(input, "combobox-composition", func() comboboxComposition { return comboboxComposition{} })
	input.ObserveInput(composition.observe)
	composition.composing = input.Composing()
	if input.IsDisabled() {
		parts.SetOpen(false)
	}
	query := strings.ToLower(strings.TrimSpace(*value))
	shown := make([]Option, 0, len(p.Options))
	for _, o := range p.Options {
		if o.Separator {
			continue
		}
		if query == "" || !parts.Filtering() || strings.Contains(strings.ToLower(o.Value+" "+o.Label), query) {
			shown = append(shown, o)
		}
	}
	parts.Popup(func(panel ui.Element) {
		menu := theme.MenuOf(c).ForScene()
		t := &menu.Tokens
		panel.MinWidth(max(float32(160), input.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).
			Radius(t.Radius*.8).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).
			Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Transition(motion.Menu(c))
		menu.Surface(c, panel)
		panel.Children(func() {
			ui.Themed(c, menu.Theme()).FillWidth().Children(func() {
				ui.Scroll(c).MaxHeight(272).FillWidth().Children(func() {
					if len(shown) == 0 {
						empty := p.EmptyText
						if empty == "" {
							empty = "No results found."
						}
						ui.Text(c, empty).FontSize(14).Padding(12).TextColor(t.MutedForeground)
					}
					group := ""
					for _, o := range shown {
						if o.Group != "" && o.Group != group {
							group = o.Group
							optionGroup(c, group)
						}
						var item ui.Element
						if o.Disabled {
							item = ui.Row(c.Key(choiceOptionKey{value: o.Value})).Role(ui.RoleListItem).Disabled(true)
						} else {
							item = parts.Item(o.Value)
						}
						item.Padding(6, 8).MinHeight(32).
							Radius(t.Radius * .6).AlignSelf(ui.Stretch).FontSize(14).Label(optionLabel(o))
						if item.Highlighted() {
							bg, fg := menu.Highlight(false)
							item.Background(bg).TextColor(fg)
						}
						if o.Disabled {
							item.Opacity(.5)
						}
						item.Children(func() {
							if p.ItemContent != nil {
								p.ItemContent(c, o)
							} else {
								ui.Text(c, optionLabel(o)).SingleLine()
							}
						})
					}
				})
			})
		})
	})
	if choice, ok := parts.Chosen(); ok && !input.IsDisabled() && !composition.suppressChoice {
		for _, o := range p.Options {
			if !o.Separator && !o.Disabled && o.Value == choice && *value != choice {
				*value = choice
				input.SetChanged()
				break
			}
		}
	}
	return input
}

func optionSeparator(c *ui.Context) {
	ui.Box(c).Height(1).FillWidth().Margin(4, 0).Background(theme.MenuOf(c).Separator()).AccessHidden(true)
}
func optionGroup(c *ui.Context, label string) {
	ui.Text(c, label).Padding(6, 8).FontSize(12).FontWeight(500).TextColor(theme.TokensOf(c).MutedForeground)
}

type MultiComboboxProps struct {
	Options                       []Option
	Label, Placeholder, EmptyText string
	Disabled, Invalid             bool
	// MaxSelected limits new choices; zero has no limit. Existing values stay
	// caller-owned even if an option is removed from the suggestion collection.
	MaxSelected int
	ItemContent func(*ui.Context, Option)
}

// MultiCombobox combines caller-owned stable option values with an independent
// search query. Selected values have keyboard-focusable remove buttons; native
// combo-box semantics retain arrow navigation, IME and clipboard editing.
func MultiCombobox(c *ui.Context, selected *[]string, query *string, p MultiComboboxProps) ui.Element {
	root := ui.Column(c).FillWidth().Gap(6)
	var input ui.Element
	handle := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	refocus := false
	root.Children(func() {
		if len(*selected) > 0 {
			ui.Row(c).Wrap().Gap(4).Children(func() {
				for _, value := range slices.Clone(*selected) {
					label := value
					for _, o := range p.Options {
						if !o.Separator && o.Value == value {
							label = optionLabel(o)
							break
						}
					}
					if Button(c.Key(choiceOptionKey{value: value}), label+" ×", ButtonProps{Variant: Secondary, Size: ExtraSmall, Label: "Remove " + label, Disabled: p.Disabled}).Clicked() {
						*selected = slices.DeleteFunc(*selected, func(v string) bool { return v == value })
						refocus = true
					}
				}
			})
		}
		parts := ui.ComboboxBase(c.Key("search"), query)
		input = styleInput(c, parts.Input, InputProps{Placeholder: p.Placeholder, Label: p.Label, Disabled: p.Disabled, Invalid: p.Invalid}).Bind(handle).FillWidth().AutoComplete(ui.AutoCompleteList)
		if refocus {
			input.Focus()
		}
		composition := ui.Local(input, "combobox-composition", func() comboboxComposition { return comboboxComposition{} })
		if input.IsDisabled() {
			parts.SetOpen(false)
		}
		composition.composing = input.Composing()
		input.ObserveInput(func(ev ui.InputEvent) {
			composition.observe(ev)
			if !p.Disabled && !composition.composing && ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyBackspace && *query == "" && len(*selected) > 0 {
				*selected = (*selected)[:len(*selected)-1]
			}
		})
		q := strings.ToLower(strings.TrimSpace(*query))
		shown := make([]Option, 0, len(p.Options))
		for _, o := range p.Options {
			if o.Separator || slices.Contains(*selected, o.Value) {
				continue
			}
			if q == "" || !parts.Filtering() || strings.Contains(strings.ToLower(o.Value+" "+o.Label), q) {
				shown = append(shown, o)
			}
		}
		parts.Popup(func(panel ui.Element) {
			menu := theme.MenuOf(c).ForScene()
			t := &menu.Tokens
			panel.MinWidth(max(float32(160), input.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).Radius(t.RadiusMD()).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).Transition(motion.Menu(c))
			menu.Surface(c, panel)
			panel.Children(func() {
				ui.Themed(c, menu.Theme()).FillWidth().Children(func() {
					ui.Scroll(c).FillWidth().MaxHeight(272).Children(func() {
						if len(shown) == 0 {
							empty := p.EmptyText
							if empty == "" {
								empty = "No results found."
							}
							ui.Text(c, empty).Padding(12).FontSize(14).TextColor(t.MutedForeground)
						}
						group := ""
						for _, o := range shown {
							if o.Group != "" && o.Group != group {
								group = o.Group
								optionGroup(c, group)
							}
							disabled := o.Disabled || p.MaxSelected > 0 && len(*selected) >= p.MaxSelected
							var item ui.Element
							if disabled {
								item = ui.Row(c.Key(choiceOptionKey{value: o.Value})).Role(ui.RoleListItem).Disabled(true)
							} else {
								item = parts.Item(o.Value)
							}
							item.Label(optionLabel(o)).Padding(6, 8).MinHeight(32).Radius(t.RadiusSM()).AlignSelf(ui.Stretch).FontSize(14)
							if item.Highlighted() {
								bg, fg := menu.Highlight(false)
								item.Background(bg).TextColor(fg)
							}
							if disabled {
								item.Opacity(.5)
							}
							item.Children(func() {
								if p.ItemContent != nil {
									p.ItemContent(c, o)
								} else {
									ui.Text(c, optionLabel(o))
								}
							})
						}
					})
				})
			})
		})
		if value, ok := parts.Chosen(); ok && !input.IsDisabled() && !composition.suppressChoice && !slices.Contains(*selected, value) && (p.MaxSelected <= 0 || len(*selected) < p.MaxSelected) {
			for _, o := range p.Options {
				if !o.Separator && !o.Disabled && o.Value == value {
					*selected = append(*selected, value)
					*query = ""
					parts.SetOpen(true)
					break
				}
			}
		}
	})

	return root
}
