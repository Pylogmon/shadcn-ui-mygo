package overlay

import (
	"math"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// CommandBinding supplies real keys; CommandItem.Shortcut is display text.
// Character keys require Ctrl or Super: ordinary typing, Shift and Alt
// characters stay with the native editor.
type CommandBinding struct {
	Modifiers ui.Modifiers
	Key       ui.Key
}

type CommandItem struct {
	// Key keeps highlight and row state stable when duplicate values are reordered.
	// Without it identity is Value and its occurrence in Items.
	Key, Value, Label, Description, Group, Shortcut string
	Keywords                                        []string
	Icon                                            *ui.SVG
	Disabled, Separator, Checked                    bool
	Binding                                         *CommandBinding
}

type CommandProps struct {
	Items                               []CommandItem
	Label, InputLabel, ListLabel        string
	Placeholder, EmptyText, LoadingText string
	Width, MaxHeight, RowHeight         float32
	AutoFocus, Disabled, Loading        bool
	NoLoop, NoFilter, ShowScrollbar     bool
	// Filter returns a positive match score. Higher scores come first within a
	// group; groups stay together and are ordered by their best matching item.
	// The default uses cmdk's fuzzy score for Label, with a distinct Value and
	// Keywords as aliases. An empty query retains registration order.
	Filter func(CommandItem, string) float64
	// ItemContent builds non-interactive option content with the current row's context.
	ItemContent                  func(*ui.Context, CommandItem, bool)
	EmptyContent, LoadingContent func(*ui.Context)
	// OnSelect is an alternative to using Command's returned event-frame value.
	OnSelect                       func(CommandItem)
	DialogTitle, DialogDescription string
	ShowClose                      bool
}

type commandKey struct {
	explicit, value string
	ordinal         int
}
type commandEntry struct {
	item      CommandItem
	key       commandKey
	score     float64
	heading   string
	separator bool
}
type commandState struct {
	list                                            ui.ListState
	items                                           []commandEntry
	active, pending                                 commandKey
	hasActive, hasPending                           bool
	position                                        int
	query                                           string
	initialized, enabled, noLoop, composing, reveal bool
	height, rowHeight                               float32
	pointerX, pointerY                              float32
	pointerOver                                     bool
}

func commandEntries(p CommandProps, query string) []commandEntry {
	entries := make([]commandEntry, 0, len(p.Items))
	occurrences := make(map[commandKey]int)
	separator := false
	for _, item := range p.Items {
		if item.Separator {
			separator = true
			continue
		}
		if item.Value == "" {
			item.Value = item.Label
		}
		if item.Label == "" {
			item.Label = item.Value
		}
		if item.Value == "" {
			continue
		}
		key := commandKey{explicit: item.Key}
		if item.Key == "" {
			key.value = item.Value
		}
		key.ordinal = occurrences[key]
		occurrences[commandKey{explicit: key.explicit, value: key.value}]++
		score := 1.0
		if !p.NoFilter && query != "" {
			if p.Filter != nil {
				score = p.Filter(item, query)
			} else {
				score = commandDefaultScore(item, query)
			}
		}
		if score > 0 && !math.IsNaN(score) {
			entries = append(entries, commandEntry{item: item, key: key, score: score, separator: separator && len(entries) > 0 && strings.TrimSpace(query) == ""})
			separator = false
		}
	}
	if query != "" && !p.NoFilter {
		groupScores := make(map[string]float64)
		groupOrder := make(map[string]int)
		groupSeparators := make(map[string]bool)
		for i, entry := range entries {
			group := entry.item.Group
			if _, exists := groupOrder[group]; !exists {
				groupOrder[group] = i
			}
			groupScores[group] = max(groupScores[group], entry.score)
			groupSeparators[group] = groupSeparators[group] || entry.separator
		}
		sort.SliceStable(entries, func(i, j int) bool {
			a, b := entries[i], entries[j]
			if a.item.Group == b.item.Group {
				return a.score > b.score
			}
			// cmdk keeps ungrouped results before the ranked group sections.
			if a.item.Group == "" || b.item.Group == "" {
				return a.item.Group == ""
			}
			if groupScores[a.item.Group] != groupScores[b.item.Group] {
				return groupScores[a.item.Group] > groupScores[b.item.Group]
			}
			return groupOrder[a.item.Group] < groupOrder[b.item.Group]
		})
		for i := range entries {
			group := entries[i].item.Group
			if group != "" {
				entries[i].separator = i > 0 && entries[i-1].item.Group != group && groupSeparators[group]
			}
		}
	}
	if len(entries) > 0 {
		entries[0].separator = false
	}
	group := ""
	for i := range entries {
		if entries[i].item.Group != group {
			group = entries[i].item.Group
			entries[i].heading = group
		}
	}
	return entries
}

func (s *commandState) eligible(i int) bool {
	return s.enabled && i >= 0 && i < len(s.items) && !s.items[i].item.Disabled
}
func (s *commandState) setActive(i int) {
	if !s.eligible(i) {
		s.hasActive = false
		return
	}
	if !s.hasActive || s.active != s.items[i].key || s.position != i {
		s.reveal = true
	}
	s.position, s.active, s.hasActive = i, s.items[i].key, true
}
func (s *commandState) ensureActive(query string) {
	reset := !s.initialized || s.query != query
	s.query, s.initialized = query, true
	if !reset && s.hasActive {
		for i := range s.items {
			if s.items[i].key == s.active && s.eligible(i) {
				s.setActive(i)
				return
			}
		}
	}
	s.hasActive = false
	start := 0
	if !reset && len(s.items) > 0 {
		start = min(max(0, s.position), len(s.items)-1)
	}
	for i := start; i < len(s.items); i++ {
		if s.eligible(i) {
			s.setActive(i)
			return
		}
	}
	for i := start - 1; i >= 0; i-- {
		if s.eligible(i) {
			s.setActive(i)
			return
		}
	}
}
func (s *commandState) move(delta int, wrap bool) {
	if len(s.items) == 0 {
		return
	}
	i := s.position
	if !s.hasActive {
		i = -1
		if delta < 0 {
			i = len(s.items)
		}
	}
	for range s.items {
		i += delta
		if i < 0 || i >= len(s.items) {
			if !wrap {
				return
			}
			i = (i + len(s.items)) % len(s.items)
		}
		if s.eligible(i) {
			s.setActive(i)
			return
		}
	}
}
func (s *commandState) edge(last bool) {
	start, delta := 0, 1
	if last {
		start, delta = len(s.items)-1, -1
	}
	for i := start; i >= 0 && i < len(s.items); i += delta {
		if s.eligible(i) {
			s.setActive(i)
			return
		}
	}
}

func commandEntryHeight(entry commandEntry, rowHeight float32) float32 {
	height := commandOptionHeight(entry.item, rowHeight)
	if entry.heading != "" {
		height += 28
	}
	if entry.separator {
		height += 9
	}
	return height
}

func (s *commandState) page(delta int) {
	if !s.hasActive {
		s.edge(delta < 0)
		return
	}
	start, target := s.position, s.position
	// Keep the current option's height as overlap between pages. Account for
	// description lines, section prefixes and disabled rows that still occupy
	// space, rather than treating every command as a 32-DIP item.
	budget := max(float32(1), s.height-commandOptionHeight(s.items[start].item, s.rowHeight))
	distance := float32(0)
	for i := start + delta; i >= 0 && i < len(s.items); i += delta {
		height := commandEntryHeight(s.items[i], s.rowHeight)
		if target != start && distance+height > budget {
			break
		}
		distance += height
		if s.eligible(i) {
			target = i
		}
		if target != start && distance >= budget {
			break
		}
	}
	s.setActive(target)
}

// input holds no previous-frame Elements or Contexts.
func (s *commandState) input(ev ui.InputEvent) bool {
	if ev.Kind != ui.InputKeyDown || !s.enabled || s.composing || commandTypesText(ev) {
		return false
	}
	for i := range s.items {
		binding := s.items[i].item.Binding
		if s.eligible(i) && binding != nil && binding.Key != ui.KeyUnknown && binding.Key == ev.Key && binding.Modifiers == ev.Mods {
			if !ev.Repeat {
				s.pending, s.hasPending = s.items[i].key, true
			}
			return true
		}
	}
	if ev.Mods != 0 {
		return false
	}
	switch ev.Key {
	case ui.KeyDown:
		s.move(1, !s.noLoop)
	case ui.KeyUp:
		s.move(-1, !s.noLoop)
	case ui.KeyHome:
		s.edge(false)
	case ui.KeyEnd:
		s.edge(true)
	case ui.KeyPageUp, ui.KeyPageDown:
		delta := 1
		if ev.Key == ui.KeyPageUp {
			delta = -1
		}
		s.page(delta)
	case ui.KeyEnter:
		if s.hasActive && !ev.Repeat {
			s.pending, s.hasPending = s.active, true
		}
	default:
		return false
	}
	if ev.Key != ui.KeyEnter && s.hasActive {
		s.reveal = true
	}
	return true
}

// Native TextInput handles text/composition through its platform text adapter,
// before HandleInput. Observe the edit stream so a candidate-choosing Enter or
// navigation key never becomes a command, even if native caret editing clears
// the marked text before the next frame.
func (s *commandState) observe(ev ui.InputEvent) {
	if ev.Kind == ui.InputCompose {
		s.composing = ev.Text != ""
	} else if ev.Kind == ui.InputText {
		s.composing = false
	}
}

// Match MyGo's native text editor: Shift and Alt can produce characters, so
// they alone never turn a character key into a command accelerator.
func commandTypesText(ev ui.InputEvent) bool {
	if ev.Mods & ^(ui.Shift|ui.Alt) != 0 {
		return false
	}
	switch ev.Key {
	case ui.KeyUnknown, ui.KeySpace, ui.KeyMinus, ui.KeyEqual, ui.KeyComma, ui.KeyPeriod, ui.KeySlash, ui.KeySemicolon, ui.KeyQuote,
		ui.KeyBracketLeft, ui.KeyBracketRight, ui.KeyBackslash, ui.KeyBackquote:
		return true
	}
	return ev.Key >= ui.Key0 && ev.Key <= ui.Key9 || ev.Key >= ui.KeyA && ev.Key <= ui.KeyZ
}

func commandBindingValid(binding *CommandBinding) bool {
	return binding != nil && binding.Key != ui.KeyUnknown && !commandTypesText(ui.InputEvent{Key: binding.Key, Mods: binding.Modifiers})
}
func commandOptionHeight(item CommandItem, rowHeight float32) float32 {
	if item.Description != "" {
		return max(48, rowHeight)
	}
	return rowHeight
}

// Command provides a native text input and virtual command list. The highlight
// follows stable keys, skips disabled items and scrolls with arrows, Home/End and
// Page Up/Down. Enter and Binding choose once. It returns the chosen Value in the
// event frame, or an empty string; OnSelect is an alternative notification.
func Command(c *ui.Context, query *string, p CommandProps) string {
	return command(c, query, p, false)
}

func command(c *ui.Context, query *string, p CommandProps, inDialog bool) string {
	if query == nil {
		panic("overlay.Command: query must not be nil")
	}
	t := theme.TokensOf(c)
	root := ui.Column(c).FillWidth().Gap(0).Padding(4).Radius(t.RadiusXL()).Clip().Background(t.Popover).TextColor(t.PopoverForeground).Role(ui.RoleGroup).Label(p.Label).Disabled(p.Disabled).Busy(p.Loading)
	if p.Width > 0 && !math.IsInf(float64(p.Width), 0) {
		root.Width(p.Width)
	}
	s := ui.Local(root, "shadcn.command", func() commandState { return commandState{position: -1} })
	s.enabled, s.noLoop = !p.Disabled && !root.IsDisabled() && !p.Loading, p.NoLoop
	s.items = commandEntries(p, *query)
	s.ensureActive(*query)
	s.rowHeight = p.RowHeight
	if s.rowHeight < 24 || math.IsNaN(float64(s.rowHeight)) || math.IsInf(float64(s.rowHeight), 0) {
		s.rowHeight = 32
	}
	maxHeight := p.MaxHeight
	if maxHeight <= 0 || math.IsNaN(float64(maxHeight)) || math.IsInf(float64(maxHeight), 0) {
		maxHeight = 288
	}
	chosen := ""
	var selected CommandItem
	selectItem := func(i int) {
		if chosen == "" && s.eligible(i) {
			selected, chosen = s.items[i].item, s.items[i].item.Value
		}
	}
	if s.hasPending {
		for i := range s.items {
			if s.items[i].key == s.pending {
				selectItem(i)
				break
			}
		}
		s.hasPending = false
	}
	root.Children(func() {
		placeholder := p.Placeholder
		if placeholder == "" {
			placeholder = "Type a command or search…"
		}
		inputLabel := p.InputLabel
		if inputLabel == "" {
			inputLabel = "Search commands"
		}
		var input ui.Element
		rightMargin := float32(4)
		if inDialog && p.ShowClose {
			rightMargin = 44
		}
		ui.Row(c).Height(32).Margin(4, rightMargin, 0, 4).Padding(0, 8).Gap(8).AlignItems(ui.Center).Radius(t.RadiusLG()).Background(t.Input.Alpha(.3)).Border(1, t.Input.Alpha(.3)).Children(func() {
			ui.Icon(c, commandSearchIcon).Size(16, 16).TextColor(t.MutedForeground).Opacity(.5).AccessHidden(true)
			input = ui.TextInputBase(c, query).Grow(1).MinWidth(0).Height(30).FontSize(14).Label(inputLabel).Placeholder(placeholder).Background(ui.Transparent).Border(0, ui.Transparent).Padding(0).FocusRing(false).Role(ui.RoleComboBox).HasPopup(ui.PopupListBox).AutoComplete(ui.AutoCompleteList).Expanded(true).HandleInput(s.input).ObserveInput(s.observe)
			if p.AutoFocus && !input.IsDisabled() {
				input.AutoFocus()
			}
			if !input.Focused() {
				s.composing = false
			}
		})
		// Prefix headings/separators keep native virtual row identities and AX set
		// positions counting commands rather than decorative section rows.
		height := float32(8)
		for _, entry := range s.items {
			height += commandEntryHeight(entry, s.rowHeight)
		}
		if p.Loading || len(s.items) == 0 {
			height = 80
		}
		s.height = min(height, maxHeight)
		listLabel := p.ListLabel
		if listLabel == "" {
			listLabel = "Commands"
		}
		count := len(s.items)
		if p.Loading {
			count = 0
		}
		list := ui.List(c, &s.list, count).Height(s.height).FillWidth().Padding(4).Gap(0).Role(ui.RoleListBox).Label(listLabel).Busy(p.Loading).Scrollbars(p.ShowScrollbar)
		input.Controls(list)
		list.Draw(func(_ *ui.Painter, r ui.Rect) { s.height = r.H })
		x, y, over := list.PointerPosition()
		pointerMoved := over && (!s.pointerOver || s.pointerX != x || s.pointerY != y)
		s.pointerX, s.pointerY, s.pointerOver = x, y, over
		s.list.Key = func(i int) any { return s.items[i].key }
		if s.reveal && s.hasActive {
			s.list.ScrollIntoView(s.position)
			s.reveal = false
		}
		firstVisible, lastVisible := s.list.Visible()
		list.Rows(func(row ui.ListRow) {
			rc, i := row.Context, row.Index
			entry := s.items[i]
			item := entry.item
			ui.Column(rc).FillWidth().Gap(0).Children(func() {
				if entry.separator {
					ui.Box(rc).Height(1).Margin(4, -4).Background(t.Border).Role(ui.RoleSeparator)
				}
				if entry.heading != "" {
					ui.Text(rc, entry.heading).Height(28).Padding(6, 8).FontSize(12).FontWeight(500).TextColor(t.MutedForeground).Role(ui.RoleHeading).Level(3)
				}
				active := s.hasActive && s.active == entry.key
				radius := t.RadiusSM()
				if inDialog {
					radius = t.RadiusLG()
				}
				option := ui.ButtonBase(rc).Role(ui.RoleOption).KeepFocus().TabStop(false).Label(item.Label).Description(item.Description).Disabled(item.Disabled || p.Loading).DimDisabled(false).Padding(6, 8).Gap(8).Height(commandOptionHeight(item, s.rowHeight)).FillWidth().Radius(radius).AccessSelected(active).PositionInSet(i+1, len(s.items))
				if pointerMoved && option.Hovered() && s.eligible(i) {
					s.setActive(i)
					active = true
					option.AccessSelected(true)
					c.Invalidate()
				}
				if active {
					option.Background(t.Muted).TextColor(t.Foreground)
					// Native List exposes the last laid-out visible range and asks
					// for another build when it changes. Never point AX focus at
					// overscan, or at an option below its visible section heading.
					if i >= firstVisible && i <= lastVisible {
						box, viewport := option.Bounds(), list.Bounds()
						if viewport.W > 8 && viewport.H > 8 && box.W == 0 && box.H == 0 {
							c.Invalidate() // one pass to lay out a newly keyed option
						} else if viewport.W > 8 && viewport.H > 8 && box.W > 0 && box.H > 0 && box.Y+box.H > viewport.Y+4 && box.Y < viewport.Y+viewport.H-4 {
							input.ActiveDescendant(option)
						}
					}
				}
				if item.Disabled || p.Loading {
					option.Opacity(.5)
				}
				if commandBindingValid(item.Binding) {
					option.KeyShortcuts(item.Binding.Modifiers, item.Binding.Key)
				}
				option.Children(func() {
					if p.ItemContent != nil {
						p.ItemContent(rc, item, active)
						return
					}
					secondary := t.MutedForeground
					if active {
						secondary = t.Foreground
					}
					if item.Icon != nil {
						ui.Icon(rc, item.Icon).Size(16, 16).TextColor(secondary).AccessHidden(true)
					}
					ui.Column(rc).Grow(1).MinWidth(0).Gap(2).Children(func() {
						ui.Text(rc, item.Label).FontSize(14).SingleLine()
						if item.Description != "" {
							ui.Text(rc, item.Description).FontSize(12).TextColor(t.MutedForeground).SingleLine()
						}
					})
					if item.Shortcut != "" {
						ui.Text(rc, item.Shortcut).FontSize(12).TextColor(secondary)
					} else if item.Checked {
						ui.Text(rc, "✓").FontSize(14).AccessHidden(true)
					}
				})
				if option.Clicked() {
					selectItem(i)
				}
			})
		})
		if p.Loading || len(s.items) == 0 {
			// A status outside the list is never a selectable synthetic command.
			list.Height(0)
			ui.Column(c).Height(min(80, maxHeight)).FillWidth().AlignItems(ui.Center).Justify(ui.Center).Role(ui.RoleStatus).Busy(p.Loading).Children(func() {
				if p.Loading && p.LoadingContent != nil {
					p.LoadingContent(c)
				} else if !p.Loading && p.EmptyContent != nil {
					p.EmptyContent(c)
				} else {
					text := p.EmptyText
					if p.Loading {
						text = p.LoadingText
						if text == "" {
							text = "Loading commands…"
						}
					} else if text == "" {
						text = "No results found."
					}
					ui.Text(c, text).FontSize(14).TextColor(t.MutedForeground).Role(ui.RoleStatus)
				}
			})
		}
	})
	if chosen != "" && p.OnSelect != nil {
		p.OnSelect(selected)
	}
	return chosen
}

var commandSearchIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>`))

// CommandDialog combines Command with a focus-trapping dialog. Choosing closes
// the dialog; keep query in caller state to preserve the search.
func CommandDialog(c *ui.Context, open *bool, query *string, p CommandProps) string {
	chosen := ""
	p.AutoFocus = true
	title, description := p.DialogTitle, p.DialogDescription
	if title == "" {
		title = "Command palette"
	}
	if description == "" {
		description = "Search for a command to run"
	}
	width := p.Width
	if width <= 0 {
		width = 448
	}
	windowWidth, windowHeight := c.Size()
	// The palette starts one third down the window. Bound its own virtual
	// viewport to the remaining space instead of nesting another Scroll.
	// Normally top is one third down. On a short window move upwards to leave
	// room for the input, optional close button and at least one standard row.
	top := min(windowHeight/3, max(float32(0), windowHeight-16-86))
	panelLimit := max(float32(1), windowHeight-top-16)
	listLimit := max(float32(1), panelLimit-46)
	if p.MaxHeight <= 0 || math.IsNaN(float64(p.MaxHeight)) || math.IsInf(float64(p.MaxHeight), 0) {
		p.MaxHeight = min(288, listLimit)
	} else {
		p.MaxHeight = min(p.MaxHeight, listLimit)
	}
	zero := float32(0)
	panel := Dialog(c, open, DialogProps{Title: title, Description: description, Width: width, MaxHeight: panelLimit, HideHeader: true, HideClose: !p.ShowClose, NoScroll: true, ContentPadding: &zero, ContentGap: &zero}, func() {
		chosen = command(c, query, p, true)
		if chosen != "" {
			*open = false
		}
	})
	panel.Absolute().Top(top).LeftPercent(50).TranslatePercent(-50, 0).MaxWidth(max(float32(1), windowWidth-32)).MaxHeight(panelLimit).Radius(theme.TokensOf(c).RadiusXL()).Clip()
	return chosen
}
