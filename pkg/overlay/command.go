package overlay

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type CommandItem struct {
	Value, Label, Group, Shortcut string
	Keywords                      []string
	Icon                          *ui.SVG
	Disabled                      bool
}

type CommandProps struct {
	Items                  []CommandItem
	Placeholder, EmptyText string
	Width                  float32
	AutoFocus              bool
}

type commandState struct {
	highlighted int
	query       string
}

func matchesCommand(item CommandItem, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	for _, word := range strings.Fields(query) {
		if !strings.Contains(strings.ToLower(item.Label+" "+item.Value+" "+strings.Join(item.Keywords, " ")), word) {
			return false
		}
	}
	return true
}

// Command provides a searchable native command list. Arrow keys move the
// highlight and Enter chooses a command; results are grouped and skip disabled
// items. It returns the chosen Value in the event frame, or an empty string.
func Command(c *ui.Context, query *string, p CommandProps) string {
	t := theme.TokensOf(c)
	root := ui.Column(c).Gap(0).Radius(t.Radius).Background(t.Popover).TextColor(t.PopoverForeground)
	if p.Width > 0 {
		root.Width(p.Width)
	}
	chosen := ""
	state := ui.Local(root, "shadcn.command", func() commandState { return commandState{highlighted: -1} })
	root.Children(func() {
		placeholder := p.Placeholder
		if placeholder == "" {
			placeholder = "Type a command or search…"
		}
		input := ui.TextInputBase(c, query).Height(44).Padding(12).FontSize(14).Label("Search commands").Placeholder(placeholder).
			Background(ui.Transparent).Border(0, ui.Transparent).Radius(t.Radius).FocusRing(false).Role(ui.RoleComboBox).Expanded(true)
		if p.AutoFocus {
			input.AutoFocus()
		}
		changed := input.Changed()
		filtered := make([]CommandItem, 0, len(p.Items))
		for _, item := range p.Items {
			if matchesCommand(item, *query) {
				filtered = append(filtered, item)
			}
		}
		first := func() int {
			for i, it := range filtered {
				if !it.Disabled {
					return i
				}
			}
			return -1
		}
		if changed || state.query != *query || state.highlighted < 0 || state.highlighted >= len(filtered) {
			state.query = *query
			state.highlighted = first()
		}
		move := func(delta int) {
			if len(filtered) == 0 {
				return
			}
			for n := 0; n < len(filtered); n++ {
				state.highlighted = (state.highlighted + delta + len(filtered)) % len(filtered)
				if !filtered[state.highlighted].Disabled {
					return
				}
			}
		}
		if input.Shortcut(0, ui.KeyDown) {
			move(1)
		}
		if input.Shortcut(0, ui.KeyUp) {
			move(-1)
		}
		submit := input.Submitted()
		list := ui.Scroll(c).MaxHeight(300).Padding(4).Gap(0).Role(ui.RoleListBox).Label("Commands")
		input.Controls(list)
		list.Children(func() {
			group := ""
			for i, item := range filtered {
				if item.Group != "" && item.Group != group {
					group = item.Group
					ui.Text(c, group).Padding(6, 8).FontSize(12).FontWeight(500).TextColor(t.MutedForeground)
				}
				row := ui.ButtonBase(c.Key(item.Value)).Role(ui.RoleOption).KeepFocus().Label(item.Label).Disabled(item.Disabled).Padding(6, 8).Gap(8).Height(34).Radius(max(2, t.Radius-2)).Selected(i == state.highlighted)
				if row.Hovered() && !item.Disabled {
					state.highlighted = i
				}
				if i == state.highlighted {
					row.Background(t.Accent).TextColor(t.AccentForeground)
					input.ActiveDescendant(row)
				}
				if item.Disabled {
					row.Opacity(0.5)
				}
				row.Children(func() {
					if item.Icon != nil {
						ui.Icon(c, item.Icon).Size(16, 16)
					}
					ui.Text(c, item.Label).Grow(1).SingleLine()
					if item.Shortcut != "" {
						ui.Text(c, item.Shortcut).FontSize(12).TextColor(t.MutedForeground)
					}
				})
				if !item.Disabled && (row.Clicked() || submit && i == state.highlighted) {
					chosen = item.Value
				}
			}
			if len(filtered) == 0 {
				empty := p.EmptyText
				if empty == "" {
					empty = "No results found."
				}
				ui.Text(c, empty).Padding(24).FontSize(14).TextColor(t.MutedForeground).Role(ui.RoleStatus)
			}
		})
	})
	return chosen
}

// CommandDialog combines Command with a focus-trapping dialog. Choosing an
// item closes the dialog; keep query in caller state to preserve the search.
func CommandDialog(c *ui.Context, open *bool, query *string, p CommandProps) string {
	chosen := ""
	p.AutoFocus = true
	Dialog(c, open, DialogProps{Title: "Command palette", Description: "Search for a command to run", Width: 512}, func() {
		chosen = Command(c, query, p)
		if chosen != "" {
			*open = false
		}
	})
	return chosen
}
