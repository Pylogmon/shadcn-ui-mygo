package control

import (
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
		if o.Value == *selected {
			label, chosen = optionLabel(o), true
			break
		}
	}
	if label == "" {
		label = "Select an option"
	}
	tr.Children(func() {
		text := ui.Text(c, label).Grow(1).SingleLine()
		if !chosen {
			text.TextColor(t.MutedForeground)
		}
		ui.Icon(c, chevron).FontSize(16).TextColor(t.MutedForeground)
	})
	parts.Popup(func(panel ui.Element) {
		panel.MinWidth(max(float32(128), tr.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).
			Radius(t.Radius*.8).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).
			Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Transition(motion.Popup(c))
		panel.Children(func() {
			for _, o := range p.Options {
				e := parts.Item(o.Value).Disabled(o.Disabled).MinHeight(32).Padding(6, 8).Gap(8).
					Radius(t.Radius * .6).AlignItems(ui.Center).AlignSelf(ui.Stretch).FontSize(14)
				if e.Highlighted() {
					e.Background(t.Accent).TextColor(t.AccentForeground)
				}
				if o.Disabled {
					e.Opacity(.5).Cursor(ui.CursorNotAllowed)
				}
				e.Children(func() {
					ui.Text(c, optionLabel(o)).Grow(1).SingleLine()
					if *selected == o.Value {
						ui.Icon(c, check).FontSize(14)
					}
				})
			}
		})
	})
	return tr
}

// NativeSelect exposes MyGo's platform-aware selection behavior with shadcn styling.
// Desktop MyGo renders its native Go popup, with the same keyboard and accessibility API as Select.
func NativeSelect(c *ui.Context, selected *string, p SelectProps) ui.Element {
	return Select(c, selected, p)
}

type ComboboxProps struct {
	Options                       []Option
	Placeholder, Label, EmptyText string
	Disabled, Invalid             bool
}

// Combobox supports free text and filtered value/label suggestions. A chosen option writes its Value.
func Combobox(c *ui.Context, value *string, p ComboboxProps) ui.Element {
	t := theme.TokensOf(c)
	parts := ui.ComboboxBase(c, value)
	input := styleInput(c, parts.Input, InputProps{Placeholder: p.Placeholder, Label: p.Label, Disabled: p.Disabled, Invalid: p.Invalid})
	query := strings.ToLower(strings.TrimSpace(*value))
	shown := make([]Option, 0, len(p.Options))
	for _, o := range p.Options {
		if query == "" || !parts.Filtering() || strings.Contains(strings.ToLower(o.Value+" "+o.Label), query) {
			shown = append(shown, o)
		}
	}
	parts.Popup(func(panel ui.Element) {
		panel.MinWidth(max(float32(160), input.Bounds().W)).MaxHeight(280).Padding(4).Margin(4, 0, 0, 0).
			Radius(t.Radius*.8).Border(1, t.Border).Background(t.Popover).TextColor(t.PopoverForeground).
			Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .12)).Transition(motion.Popup(c))
		panel.Children(func() {
			if len(shown) == 0 {
				empty := p.EmptyText
				if empty == "" {
					empty = "No results found."
				}
				ui.Text(c, empty).FontSize(14).Padding(12).TextColor(t.MutedForeground)
			}
			for _, o := range shown {
				item := parts.Item(o.Value).Disabled(o.Disabled).Padding(6, 8).MinHeight(32).
					Radius(t.Radius * .6).AlignSelf(ui.Stretch).FontSize(14)
				if item.Highlighted() {
					item.Background(t.Accent).TextColor(t.AccentForeground)
				}
				if o.Disabled {
					item.Opacity(.5)
				}
				item.Children(func() { ui.Text(c, optionLabel(o)).SingleLine() })
			}
		})
	})
	if choice, ok := parts.Chosen(); ok {
		*value = choice
	}
	return input
}
