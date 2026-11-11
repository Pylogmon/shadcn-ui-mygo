package desktop

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type selectState struct {
	open, changed bool
	menu          MenuState
	bound         *string
	host          *Host
	key           string
	closed        bool
}

func (s *selectState) Close() error {
	s.closed, s.open, s.changed = true, false, false
	if s.host != nil {
		if e := s.host.entries[s.key]; e != nil && e.open == &s.open {
			// A removed/rebound owner cannot lend its open pointer to an exit
			// animation or close a newer owner that already reused this key.
			s.host.closeEntry(e)
			delete(s.host.entries, s.key)
			if s.host.backend != nil {
				s.host.backend.InvalidateParent()
			}
		}
	}
	s.bound, s.host = nil, nil
	return nil
}

// Select shows a shadcn-styled selection list in an independent native Flyout.
// Host and key identify the parent and popup; state remains caller-controlled.
// NativeSelect is the distinct operating-system menu implementation.
func Select(c *ui.Context, host *Host, key string, selected *string, p control.SelectProps) (ui.Element, error) {
	if selected == nil {
		panic("desktop.Select: selected is required")
	}
	host = resolveHost(c, host)
	t := theme.TokensOf(c)
	anchor := ui.ButtonBase(c).Role(ui.RoleComboBox).Disabled(p.Disabled).Invalid(p.Invalid)
	s := ui.Local(anchor, "desktop.select", func() selectState { return selectState{bound: selected, host: host, key: key} })
	if s.bound != selected || s.host != host || s.key != key {
		s.Close()
		*s = selectState{bound: selected, host: host, key: key}
	}
	control.StyleButton(c, anchor, "", control.ButtonProps{Variant: control.Outline, Size: p.Size, FontSize: p.FontSize, Disabled: p.Disabled}).
		Justify(ui.Start).MinWidth(144).Border(1, t.Input).Background(t.Background).TextColor(t.Foreground)
	if p.Invalid {
		anchor.Border(1, t.Destructive)
	}
	if p.Label != "" {
		anchor.Label(p.Label)
	}
	label := p.Placeholder
	if label == "" {
		label = "Select an option"
	}
	var chosen control.Option
	found := false
	items := make([]overlay.MenuItem, 0, len(p.Options))
	// Select commits a value and closes, independently of menu radio defaults.
	closeOnSelect := true
	for _, o := range p.Options {
		text := o.Label
		if text == "" {
			text = o.Value
		}
		if !o.Separator && o.Value == *selected {
			chosen = o
			label = text
			found = true
		}
		items = append(items, overlay.MenuItem{Key: o.Value, Label: text, Disabled: o.Disabled, Separator: o.Separator, Group: o.Group, Radio: !o.Separator, Selected: o.Value == *selected, CloseOnSelect: &closeOnSelect})
	}
	anchor.Value(label).Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).Children(func() {
			if found && p.TriggerContent != nil {
				p.TriggerContent(c, chosen)
			} else {
				text := ui.Text(c, label).SingleLine()
				if !found {
					text.TextColor(t.MutedForeground)
				}
			}
		})
		indicator.ChevronDown(c).TextColor(t.MutedForeground)
	})
	if anchor.IsDisabled() {
		s.open = false
	}
	reference := float32(14)
	if p.Size == control.ExtraSmall || p.Size == control.IconExtraSmall {
		reference = 12
	}
	font := theme.ControlFontSize(c, reference, p.FontSize)
	if p.Content.FontSize > 0 && !math.IsInf(float64(p.Content.FontSize), 0) && !math.IsNaN(float64(p.Content.FontSize)) {
		font = theme.ControlFontSize(c, reference, p.Content.FontSize)
	}
	_, err := Menu(c, host, key, anchor, &s.open, &s.menu, MenuProps{selectList: true, Items: items, Width: max(144, anchor.Bounds().W), FontSize: font / theme.TextScale(c), Label: p.Label,
		ItemContent: func(child *ui.Context, item overlay.MenuItem) {
			for _, option := range p.Options {
				if !option.Separator && option.Value == item.Key {
					if p.ItemContent != nil {
						p.ItemContent(child, option)
					} else {
						ui.Text(child, item.Label).SingleLine()
					}
					return
				}
			}
		},
		OnChoose: func(value string) {
			if !s.closed && s.open && s.bound == selected && s.host == host && s.key == key && value != *selected {
				*selected = value
				s.changed = true
				if p.OnChange != nil {
					p.OnChange(value)
				}
			}
		},
	})
	if s.changed {
		anchor.SetChanged()
		s.changed = false
	}
	return anchor, err
}

// NativeSelect always uses the OS system-menu route. Its native painting and
// placement are intentionally independent of shadcn's Flyout selection list.
func NativeSelect(c *ui.Context, selected *string, p control.SelectProps) ui.Element {
	return control.NativeSelect(c, selected, p)
}
