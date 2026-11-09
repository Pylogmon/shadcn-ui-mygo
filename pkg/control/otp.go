package control

import (
	"math"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type InputOTPProps struct {
	// Length is the native rune-slot count, default six. Counts above 4096
	// expose Invalid and use a finite single-group fallback.
	Length             int
	Label, Placeholder string
	// The upstream 1.4 default is unrestricted. DigitsOnly/Alphanumeric are
	// optional scrubbing presets; AllowRune replaces either preset.
	DigitsOnly, Alphanumeric bool
	AllowRune                func(rune) bool
	// Pattern gates the complete candidate after paste transformation,
	// optional scrubbing and truncation. Anchor it to validate the whole code.
	Pattern *regexp.Regexp
	// PasteTransform follows input-otp 1.4.2: after replacing the native
	// selected range, selection moves to the final slot/end of the code.
	// Transforming an empty clipboard cannot create a MyGo undo transaction.
	PasteTransform                         func(string) string
	Disabled, Invalid, ReadOnly, AutoFocus bool
	// Groups splits the slots; when Length is zero it supplies the total.
	// Invalid group sizes fall back to one group and expose Invalid.
	Groups []int
	// Zero dimensions use Vega's 36 DIPs. A nil Gap uses 8 DIPs.
	SlotWidth, SlotHeight float32
	Gap                   *float32
	// Callbacks paint inert content; native input remains owned by this editor.
	HideSeparators       bool
	Separator            func(*ui.Context, int)
	SlotContent          func(*ui.Context, InputOTPSlotProps)
	Handle               *ui.Handle
	OnChange, OnComplete func(string)
}

type InputOTPSlotProps struct {
	Index                     int
	Char, Placeholder         string
	Active, HasCaret, Invalid bool
	First, Last               bool
	Width, Height             float32
}

type otpSlotKey struct{ index int }
type otpGroupKey struct{ index int }
type otpSeparatorKey struct{ index int }
type otpInputKey struct{}
type otpVisualKey struct{}

// Bound native element construction independently of the caller's integer
// range. Invalid counts render a usable single group instead of allocating
// or looping in proportion to an accidental unbounded Length.
const otpSlotLimit = 4096

type otpPaste struct{ raw, transformed string }
type otpSelection struct{ anchor, caret int }
type otpState struct {
	initialized, focused bool
	composing            bool
	previous             string
	editValue            string
	syncExternal         bool
	focus                ui.Handle
	selection            otpSelection
	compositionSelection otpSelection
	pending              *otpSelection
	paste                *otpPaste
	positions            []float32
	length               int
	slotWidth            float32
	disabled, dragging   bool
	rtl                  bool
	dragAnchor           int
}

func (s *otpState) textKey(key ui.TextKey) bool {
	if !s.rtl || s.disabled || key.Mods&^ui.Shift != 0 || key.Key != ui.KeyLeft && key.Key != ui.KeyRight {
		return false
	}
	selection := s.selection
	if s.pending != nil {
		selection = *s.pending
	}
	delta := 1
	if key.Key == ui.KeyRight {
		delta = -1
	}
	if key.Mods&ui.Shift != 0 {
		selection.caret = min(max(0, selection.caret+delta), s.length)
	} else {
		position := min(selection.anchor, selection.caret)
		position = min(max(0, position+delta), s.length)
		selection = otpSelection{position, min(position+1, s.length)}
	}
	s.pending = &selection
	return true
}

func otpGroups(p InputOTPProps) (int, []int, bool) {
	n := p.Length
	if n < 0 || n > otpSlotLimit {
		return 6, []int{6}, false
	}
	total, valid := 0, true
	for _, count := range p.Groups {
		if count < 0 || count > otpSlotLimit-total {
			if n == 0 {
				n = 6
			}
			return n, []int{n}, false
		}
		total += count
		if count == 0 {
			valid = false
		}
	}
	if n == 0 {
		n = 6
		if total > 0 {
			n = total
		}
	}
	groups := p.Groups
	if len(groups) == 0 {
		return n, []int{n}, true
	}
	if !valid || total != n {
		return n, []int{n}, false
	}
	return n, groups, true
}

func otpClean(value string, p InputOTPProps, length int) (string, []int) {
	input := []rune(value)
	output := make([]rune, 0, min(length, len(input)))
	positions := make([]int, len(input)+1)
	for i, ch := range input {
		accept := true
		switch {
		case p.AllowRune != nil:
			accept = p.AllowRune(ch)
		case p.DigitsOnly:
			accept = ch >= '0' && ch <= '9'
		case p.Alphanumeric:
			accept = ch >= '0' && ch <= '9' || unicode.IsLetter(ch)
		}
		if accept && len(output) < length {
			output = append(output, ch)
		}
		positions[i+1] = len(output)
	}
	return string(output), positions
}

func (s *otpState) indexAt(x float32) int {
	index := 0
	distance := float32(math.Inf(1))
	for i, position := range s.positions {
		if d := float32(math.Abs(float64(x - (position + s.slotWidth/2)))); d < distance {
			index, distance = i, d
		}
	}
	return min(index, s.length)
}

func (s *otpState) pointer(ev ui.InputEvent) bool {
	if s.disabled {
		return false
	}
	switch ev.Kind {
	case ui.InputPointerDown:
		if ev.Button != 0 {
			return false
		}
		index := s.indexAt(ev.X)
		if ev.Clicks >= 2 {
			s.pending = &otpSelection{0, s.length}
			s.dragging = false
			return true
		}
		s.dragAnchor = index
		selection := otpSelection{index, min(index+1, s.length)}
		if ev.Mods&ui.Shift != 0 {
			selection.anchor = s.selection.anchor
		}
		s.pending, s.dragging = &selection, true
		return true
	case ui.InputPointerMove:
		if !s.dragging {
			return false
		}
		index := s.indexAt(ev.X)
		if index < s.dragAnchor {
			s.pending = &otpSelection{min(s.dragAnchor+1, s.length), index}
		} else {
			s.pending = &otpSelection{s.dragAnchor, min(index+1, s.length)}
		}
		return true
	case ui.InputPointerUp:
		if !s.dragging {
			return false
		}
		s.dragging = false
		return true
	}
	return false
}

// InputOTP keeps one actual MyGo editor for native text services, selection,
// clipboard and undo, and paints inert slots over it. It returns that editor;
// Handle can also bind it for explicit focus and label association.
func InputOTP(c *ui.Context, value *string, p InputOTPProps) ui.Element {
	if value == nil {
		panic("control.InputOTP: value must not be nil")
	}
	p.Invalid = fieldInputInvalid(c, p.Invalid)
	n, groups, valid := otpGroups(p)
	t := theme.TokensOf(c)
	slotWidth, slotHeight := p.SlotWidth, p.SlotHeight
	for _, dimension := range []*float32{&slotWidth, &slotHeight} {
		if *dimension == 0 {
			*dimension = 36
		} else if *dimension < 0 || math.IsNaN(float64(*dimension)) || math.IsInf(float64(*dimension), 0) {
			*dimension, valid = 36, false
		}
	}
	gap := float32(8)
	if p.Gap != nil {
		if *p.Gap < 0 || math.IsNaN(float64(*p.Gap)) || math.IsInf(float64(*p.Gap), 0) {
			valid = false
		} else {
			gap = *p.Gap
		}
	}
	between64 := float64(gap)
	if !p.HideSeparators {
		between64 = between64*2 + 16
	}
	width64 := float64(n)*float64(slotWidth) + float64(len(groups)-1)*between64
	if between64 > math.MaxFloat32 || width64 > math.MaxFloat32 {
		// Individually finite float32 inputs can still overflow once groups
		// and separators combine. Never send infinite bounds to native AX,
		// layout or the painter.
		slotWidth, slotHeight, gap, valid = 36, 36, 8, false
		between64 = float64(gap)
		if !p.HideSeparators {
			between64 = between64*2 + 16
		}
		width64 = float64(n)*float64(slotWidth) + float64(len(groups)-1)*between64
	}
	width, between := float32(width64), float32(between64)
	root := ui.Box(c).Size(width, slotHeight).Shrink(0).AlignSelf(ui.Start).Disabled(p.Disabled).DimDisabled(false)
	if p.Disabled {
		root.Opacity(.5)
	}
	s := ui.Local(root, "input-otp", func() otpState { return otpState{} })
	external := !s.initialized || s.syncExternal || *value != s.previous
	if external {
		s.editValue = *value
		s.syncExternal = false
	}
	s.positions = slices.Grow(s.positions[:0], n)
	x := float32(0)
	for gi, count := range groups {
		if gi > 0 {
			x += between
		}
		for range count {
			s.positions = append(s.positions, x)
			x += slotWidth
		}
	}
	if c.Direction() == ui.RightToLeft {
		for i := range s.positions {
			s.positions[i] = width - s.positions[i] - slotWidth
		}
	}
	s.slotWidth = slotWidth
	s.rtl = c.Direction() == ui.RightToLeft
	var input ui.Element
	root.Children(func() {
		input = ui.TextInputBase(c.Key(otpInputKey{}), &s.editValue).Absolute().Fill().Opacity(0).Padding(0).Font("monospace").FontSize(14).
			FocusRing(false).Disabled(p.Disabled).ReadOnly(p.ReadOnly).Invalid(p.Invalid || !valid).CaretWidth(0)
		label := p.Label
		if label == "" {
			label = "One-time code"
		}
		input.Label(label)
		input.Bind(&s.focus)
		if p.Handle != nil {
			input.Bind(p.Handle)
		}
		if p.AutoFocus && !input.IsDisabled() {
			input.AutoFocus()
		}
		s.disabled = input.IsDisabled()
		input.HandleInput(s.pointer)
		input.OnTextKey(s.textKey)
		input.OnPaste(func(text string) bool {
			if p.PasteTransform != nil {
				s.paste = &otpPaste{raw: text, transformed: p.PasteTransform(text)}
			}
			return false
		})
		candidate := s.editValue
		anchor, caret := input.TextAnchorCaret()
		pasteApplied := false
		if paste := s.paste; paste != nil {
			s.paste = nil
			raw := []rune(candidate)
			inserted := []rune(strings.Map(func(ch rune) rune {
				if ch == '\r' || ch == '\n' {
					return ' '
				}
				return ch
			}, strings.ReplaceAll(paste.raw, "\r\n", "\n")))
			start, end := caret-len(inserted), caret
			if len(inserted) == 0 {
				// MyGo leaves an empty clipboard's selected range in place;
				// the transform may still replace it with non-empty content.
				start, end = min(anchor, caret), max(anchor, caret)
			}
			if start >= 0 && end <= len(raw) && (len(inserted) == 0 || slices.Equal(raw[start:end], inserted)) {
				candidate = string(raw[:start]) + paste.transformed + string(raw[end:])
				// input-otp 1.4.2 deliberately selects the end after transformed
				// paste, even when it replaced a slot in the middle.
				anchor, caret = len([]rune(candidate)), len([]rune(candidate))
				pasteApplied = true
			}
		}
		composing := input.Composing()
		if composing && !s.composing {
			s.compositionSelection = s.selection
		}
		restoreComposition := !composing && s.composing && !input.Changed() && !external
		if restoreComposition {
			// Native preedit can remove its selected committed range before
			// clearing the mark. An empty preedit may be cancellation, or the
			// gap before GTK sends its commit. Neither publishes a deletion.
			candidate = s.previous
			anchor, caret = s.compositionSelection.anchor, s.compositionSelection.caret
		}
		s.composing = composing
		if !composing {
			clean, mapping := otpClean(candidate, p, n)
			anchor = mapping[min(max(0, anchor), len(mapping)-1)]
			caret = mapping[min(max(0, caret), len(mapping)-1)]
			if p.Pattern != nil && clean != "" && !p.Pattern.MatchString(clean) && s.initialized && !external {
				clean = s.previous
				anchor, caret = s.selection.anchor, s.selection.caret
			}
			if clean != s.editValue {
				s.editValue = clean
				s.pending = &otpSelection{anchor, caret}
				c.Invalidate()
			}
			*value = clean
			if pasteApplied && !external && clean != s.previous {
				// Replacing native text with identical raw clipboard text need
				// not mark the editor changed, although its transform did.
				input.SetChanged()
			}
		}
		chars := []rune(*value)
		s.length = len(chars)
		focused := s.focus.Focused(c) && !input.IsDisabled()
		if s.pending != nil && !composing && !restoreComposition {
			anchor, caret = min(s.pending.anchor, len(chars)), min(s.pending.caret, len(chars))
			s.pending = nil
		} else if focused && !s.focused && !composing {
			anchor, caret = min(len(chars), n-1), len(chars)
		}
		if focused && !composing && len(chars) > 0 && anchor == caret && !(caret == len(chars) && len(chars) < n) {
			pos, offset := caret, 0
			if pos == n {
				anchor, caret = n, n-1
			} else if pos == 0 {
				anchor, caret = 0, 1
			} else {
				backward := pos < max(s.selection.anchor, s.selection.caret)
				wasInserting := s.selection.anchor == s.selection.caret && s.selection.anchor < n
				if backward && !wasInserting {
					offset = -1
				}
				if backward {
					anchor, caret = pos+offset+1, pos+offset
				} else {
					anchor, caret = pos+offset, pos+offset+1
				}
			}
		}
		a, z := input.TextAnchorCaret()
		if !composing && !restoreComposition && (a != anchor || z != caret) {
			input.SetTextSelection(anchor, caret)
		}
		s.selection = otpSelection{anchor, caret}
		s.focused = focused
		if !composing {
			accepted := *value
			overridden := false
			if s.initialized && s.previous != accepted {
				if input.Changed() && !external && p.OnChange != nil {
					p.OnChange(accepted)
					c.Invalidate()
					overridden = *value != accepted
				}
				if !overridden && len([]rune(s.previous)) < n && len(chars) == n && p.OnComplete != nil {
					p.OnComplete(accepted)
					c.Invalidate()
				}
			}
			if !overridden {
				s.previous = accepted
			}
			if *value != accepted {
				s.syncExternal = true
				c.Invalidate()
			}
			s.initialized = true
		}
		start, end := min(anchor, caret), max(anchor, caret)
		ui.Row(c.Key(otpVisualKey{})).Size(width, slotHeight).Gap(gap).IgnoreInput(true).Disabled(true).DimDisabled(false).AccessHidden(true).Children(func() {
			index := 0
			for gi, count := range groups {
				if gi > 0 && !p.HideSeparators {
					ui.Box(c.Key(otpSeparatorKey{gi})).Width(16).Height(slotHeight).Center().Children(func() {
						if p.Separator != nil {
							p.Separator(c, index)
						} else {
							InputOTPSeparator(c)
						}
					})
				}
				group := ui.Row(c.Key(otpGroupKey{gi})).Gap(0).Radius(t.RadiusMD())
				if p.Invalid || !valid {
					alpha := float32(.2)
					if t.Dark {
						alpha = .4
					}
					group.Shadow(0, 0, 0, 3, t.Destructive.Alpha(alpha))
				}
				group.Children(func() {
					for local := range count {
						props := InputOTPSlotProps{Index: index, Invalid: p.Invalid || !valid, First: local == 0, Last: local == count-1, Width: slotWidth, Height: slotHeight}
						if index < len(chars) {
							props.Char = string(chars[index])
						} else if len(chars) == 0 && index < len([]rune(p.Placeholder)) {
							props.Placeholder = string([]rune(p.Placeholder)[index])
						}
						props.Active = focused && (index >= start && index < end || start == end && index == start)
						props.HasCaret = props.Active && props.Char == "" && !p.ReadOnly && !composing
						wrapper := ui.Box(c.Key(otpSlotKey{index})).Size(slotWidth, slotHeight)
						if props.Active {
							wrapper.ZIndex(1)
						}
						wrapper.Children(func() {
							if p.SlotContent != nil {
								p.SlotContent(c, props)
							} else {
								InputOTPSlot(c, props)
							}
						})
						index++
					}
				})
			}
		})
	})
	return input
}

func InputOTPSlot(c *ui.Context, p InputOTPSlotProps) ui.Element {
	t := theme.TokensOf(c)
	width, height := p.Width, p.Height
	for _, dimension := range []*float32{&width, &height} {
		if *dimension <= 0 || math.IsNaN(float64(*dimension)) || math.IsInf(float64(*dimension), 0) {
			*dimension = 36
		}
	}
	bg := ui.Transparent
	if t.Dark {
		bg = t.Input.Alpha(.3)
	}
	border := t.Input
	if p.Invalid {
		border = t.Destructive
	}
	root := motion.ApplyTransition(c, ui.Box(c).Size(width, height).Border(1, border).BorderWidth(1, 1, 1, 0).Background(bg).TextColor(t.Foreground).Center().FontSize(14)).AccessHidden(true).IgnoreInput(true)
	if c.Direction() == ui.RightToLeft {
		root.BorderWidth(1, 0, 1, 1)
	}
	if p.First {
		root.BorderWidth(1)
	}
	radius := t.RadiusMD()
	left, right := p.First, p.Last
	if c.Direction() == ui.RightToLeft {
		left, right = right, left
	}
	if left {
		root.Radius(radius, 0, 0, radius)
	}
	if right {
		root.Radius(0, radius, radius, 0)
	}
	if left && right {
		root.Radius(radius)
	}
	if p.Active {
		ring := t.Ring.Alpha(.5)
		border = t.Ring
		if p.Invalid {
			border = t.Destructive
			ring = t.Destructive.Alpha(.2)
			if t.Dark {
				ring = t.Destructive.Alpha(.4)
			}
		}
		root.Border(1, border).ZIndex(1).Shadow(0, 0, 0, 3, ring)
	}
	caretState := ui.Local(root, "caret", func() struct {
		active bool
		start  time.Time
	} {
		return struct {
			active bool
			start  time.Time
		}{}
	})
	if p.HasCaret && !caretState.active {
		caretState.start = c.Now()
	}
	caretState.active = p.HasCaret
	root.Children(func() {
		text := p.Char
		if text == "" {
			text = p.Placeholder
		}
		if text != "" {
			e := ui.Text(c, text)
			if p.Char == "" {
				e.TextColor(t.MutedForeground)
			}
		}
		if p.HasCaret {
			visible := true
			if !motion.Reduced(c) {
				elapsed := c.Now().Sub(caretState.start)
				visible = elapsed%(time.Second) < 500*time.Millisecond
				c.After(500*time.Millisecond - elapsed%(500*time.Millisecond))
			}
			if visible {
				ui.Box(c).Absolute().Fill().Center().AccessHidden(true).Children(func() {
					ui.Box(c).Width(1).Height(16).Background(t.Foreground)
				})
			}
		}
	})
	return root
}

func InputOTPSeparator(c *ui.Context) ui.Element {
	return ui.Box(c).Size(16, 16).Center().AccessHidden(true).Children(func() { ui.Box(c).Width(10).Height(1).Background(theme.TokensOf(c).Foreground) })
}
