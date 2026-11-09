package control

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var chevron = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>`))

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
	tr.MinHeight(h).PaddingX(min(px, 12)).Gap(8).AlignItems(ui.Center).Radius(t.Radius*.8).
		Background(t.Background).TextColor(t.Foreground).FontSize(14).Border(1, t.Input).
		Disabled(p.Disabled).Invalid(p.Invalid).Cursor(ui.CursorPointer).Transition(motion.Transition(c))
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
		ui.Icon(c, chevron).FontSize(16).TextColor(t.MutedForeground)
	})
	parts.Popup(func(panel ui.Element) {
		panel.MinWidth(max(float32(128), tr.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).
			Radius(t.Radius*.8).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).
			Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Transition(motion.Popup(c))
		panel.Children(func() {
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
						e.Background(t.Accent).TextColor(t.AccentForeground)
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
	return ButtonGroupItem(c, tr)
}

// NativeSelect uses MyGo's system-menu adapter instead of SelectBase's in-window
// listbox. The native OS owns popup placement, keyboard navigation and painting,
// and can display beyond the parent window. A caller explicitly choosing MyGo's
// drawn menu presentation is respected. Native options use text, groups and
// selected marks; arbitrary ItemContent belongs to the styled Select instead.
func NativeSelect(c *ui.Context, selected *string, p SelectProps) ui.Element {
	t := theme.TokensOf(c)
	h, px := metrics(p.Size)
	tr := ui.ButtonBase(c).MinHeight(h).PaddingX(min(px, 12)).Gap(8).AlignItems(ui.Center).Radius(t.RadiusMD()).Background(t.Background).TextColor(t.Foreground).FontSize(14).Border(1, t.Input).Disabled(p.Disabled).Invalid(p.Invalid).Cursor(ui.CursorPointer).Transition(motion.Transition(c))
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
		ui.Icon(c, chevron).FontSize(16).TextColor(t.MutedForeground)
	})
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
	t := theme.TokensOf(c)
	parts := ui.ComboboxBase(c, value)
	input := styleInput(c, parts.Input, InputProps{Placeholder: p.Placeholder, Label: p.Label, Disabled: p.Disabled, Invalid: p.Invalid})
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
		panel.MinWidth(max(float32(160), input.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).
			Radius(t.Radius*.8).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).
			Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Transition(motion.Popup(c))
		panel.Children(func() {
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
					item := parts.Item(o.Value).Disabled(o.Disabled).Padding(6, 8).MinHeight(32).
						Radius(t.Radius * .6).AlignSelf(ui.Stretch).FontSize(14).Label(optionLabel(o))
					if item.Highlighted() {
						item.Background(t.Accent).TextColor(t.AccentForeground)
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
	if choice, ok := parts.Chosen(); ok {
		*value = choice
	}
	return input
}

func optionSeparator(c *ui.Context) {
	ui.Box(c).Height(1).FillWidth().Margin(4, 0).Background(theme.TokensOf(c).Border).AccessHidden(true)
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
	t := theme.TokensOf(c)
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
					if Button(c, label+" ×", ButtonProps{Variant: Secondary, Size: ExtraSmall, Label: "Remove " + label, Disabled: p.Disabled}).Clicked() {
						*selected = slices.DeleteFunc(*selected, func(v string) bool { return v == value })
						refocus = true
					}
				}
			})
		}
		parts := ui.ComboboxBase(c.Key("search"), query)
		input = styleInput(c, parts.Input, InputProps{Placeholder: p.Placeholder, Label: p.Label, Disabled: p.Disabled, Invalid: p.Invalid}).Bind(handle).FillWidth()
		if refocus {
			input.Focus()
		}
		input.ObserveInput(func(ev ui.InputEvent) {
			if !p.Disabled && ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyBackspace && *query == "" && len(*selected) > 0 {
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
			panel.MinWidth(max(float32(160), input.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).Radius(t.RadiusMD()).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).Transition(motion.Popup(c)).Children(func() {
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
						item := parts.Item(o.Value).Label(optionLabel(o)).Disabled(disabled).Padding(6, 8).MinHeight(32).Radius(t.RadiusSM()).AlignSelf(ui.Stretch).FontSize(14)
						if item.Highlighted() {
							item.Background(t.Accent).TextColor(t.AccentForeground)
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
		if value, ok := parts.Chosen(); ok && !p.Disabled && !slices.Contains(*selected, value) && (p.MaxSelected <= 0 || len(*selected) < p.MaxSelected) {
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
