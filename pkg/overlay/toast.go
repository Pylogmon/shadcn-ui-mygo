package overlay

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Toast preserves MyGo's window-owned queue and fallback renderer. For a
// configurable Sonner queue, keep a ToasterState and pass it to ToasterWith.
func Toast(c *ui.Context, toast ui.Toast) string { return c.AddToast(toast) }
func Toaster(c *ui.Context) ui.Element           { return ToasterWith(c, ToasterProps{}) }

type toastLocalKey struct{}
type toastNativeKey struct{ id string }
type toastAppKey struct{ id uint64 }
type toastPositionKey ToastPosition
type toastViewportState struct {
	mount                *toastMount
	order                map[*toastEntry]uint64
	sequence             uint64
	native               map[string]*toastEntry
	state                ToasterState
	returnFocus          *ui.Handle
	focused              bool
	hovered, interacting bool
}

func (s *toastViewportState) Close() error {
	if s.focused && s.returnFocus != nil {
		s.returnFocus.Focus()
	}
	return nil
}

type toastRenderEntry struct {
	entry  *toastEntry
	native bool
}

// ToasterWith builds one native overlay, combining legacy native notifications
// and the optional application queue in the same position stacks. Native
// notifications retain MyGo's fixed latest-three enumeration and timer policy.
func ToasterWith(c *ui.Context, props ToasterProps) ui.Element {
	return ui.ToastViewportBase(c, func(viewport ui.Element, native []ui.Toast) {
		local := ui.Local(viewport, toastLocalKey{}, func() toastViewportState {
			return toastViewportState{native: map[string]*toastEntry{}, order: map[*toastEntry]uint64{}, mount: &toastMount{}}
		})
		state := props.State
		if state == nil {
			state = &local.state
		}
		state.attach(local.mount)
		hotkey := ToastHotkey{ui.Alt, ui.KeyT}
		if props.Hotkey != nil {
			hotkey = *props.Hotkey
		}
		label := props.Label
		if label == "" {
			label = "Notifications"
		}
		viewport.Role(ui.RoleRegion).Label(label).KeyShortcuts(hotkey.Modifiers, hotkey.Key).Padding(0).Gap(0)
		local.returnFocus = props.ReturnFocus
		viewport.ObserveInput(func(ev ui.InputEvent) {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
				local.interacting = true
			}
			if ev.Kind == ui.InputPointerUp && ev.Button == 0 {
				local.interacting = false
			}
		})
		focused := viewport.FocusWithin()
		if local.focused && !focused {
			state.expanded = false
			if props.ReturnFocus != nil {
				props.ReturnFocus.Focus(c)
			}
		}
		local.focused = focused
		hotkeyPressed := c.Shortcut(hotkey.Modifiers, hotkey.Key)
		if hotkeyPressed {
			state.expanded = true
			state.list.Focus(c)
		}
		if focused && viewport.Shortcut(0, ui.KeyEscape) {
			state.expanded = false
			if props.ReturnFocus != nil {
				props.ReturnFocus.Focus(c)
			}
		}
		hovered := viewport.Hovered()
		if hovered {
			state.expanded = true
		} else if local.hovered && !local.interacting {
			state.expanded = false
		} else if !focused && !hotkeyPressed && !local.interacting {
			state.expanded = false
		}
		local.hovered = hovered
		interacting := local.interacting
		for _, e := range state.entries {
			interacting = interacting || e.swiping
		}
		pause := c.WindowHidden() || state.expanded || interacting
		state.advance(c, props, pause)
		position := toastValidPosition(props.Position)
		groups := map[ToastPosition][]toastRenderEntry{}
		present := map[string]bool{}
		liveOrder := map[*toastEntry]bool{}
		register := func(e *toastEntry) {
			liveOrder[e] = true
			if local.order[e] == 0 {
				local.sequence++
				local.order[e] = local.sequence
			}
		}
		// Discover each lane oldest first; subsequent arrivals share one ordering.
		for _, model := range native {
			if e := local.native[model.ID]; e != nil {
				register(e)
			}
		}
		for i := 0; i < len(native); i++ {
			model := native[i]
			present[model.ID] = true
			e := local.native[model.ID]
			if e == nil {
				e = &toastEntry{height: 54}
				local.native[model.ID] = e
			}
			e.model = model
			register(e)
			groups[position] = append(groups[position], toastRenderEntry{e, true})
		}
		for id := range local.native {
			if !present[id] {
				delete(local.native, id)
			}
		}
		// Each lane retains its native order; application entries are newest first.
		for i := 0; i < len(state.entries); i++ {
			e := state.entries[i]
			register(e)
			pos := position
			if e.options.Position != "" {
				pos = toastValidPosition(e.options.Position)
			}
			groups[pos] = append(groups[pos], toastRenderEntry{e, false})
		}
		for e := range local.order {
			if !liveOrder[e] {
				delete(local.order, e)
			}
		}
		for pos, entries := range groups {
			sort.SliceStable(entries, func(i, j int) bool { return local.order[entries[i].entry] > local.order[entries[j].entry] })
			groups[pos] = entries
		}
		w, _ := c.Size()
		offset := ToastInsets{24, 24, 24, 24}
		if props.Offset != nil {
			offset = *props.Offset
		}
		if w <= 600 {
			offset = ToastInsets{16, 16, 16, 16}
			if props.MobileOffset != nil {
				offset = *props.MobileOffset
			}
		}
		offset = toastSafeInsets(offset)
		width := props.Width
		if !toastFinite(width) || width <= 0 {
			width = 356
		}
		width = min(width, max(float32(1), w-offset.Left-offset.Right))
		if w <= 600 {
			width = max(1, w-offset.Left-offset.Right)
		}
		gap := float32(14)
		if props.Gap != nil && toastFinite(*props.Gap) && *props.Gap >= 0 {
			gap = *props.Gap
		}
		limit := 3
		if props.VisibleToasts != nil {
			limit = max(0, *props.VisibleToasts)
		}
		first := true
		for _, pos := range []ToastPosition{ToastTopLeft, ToastTopCenter, ToastTopRight, ToastBottomLeft, ToastBottomCenter, ToastBottomRight} {
			entries := groups[pos]
			if len(entries) == 0 {
				continue
			}
			expanded := props.Expand || state.expanded
			visible := min(limit, len(entries))
			total := float32(0)
			for i := 0; i < visible; i++ {
				if expanded {
					total += entries[i].entry.height
					if i > 0 {
						total += gap
					}
				} else if i == 0 {
					total = entries[0].entry.height
				} else {
					total += gap
				}
			}
			list := ui.Box(c.Key(toastPositionKey(pos))).Absolute().Width(width).Height(total).PassThrough().Role(ui.RoleList).Label("Notifications")
			if first {
				list.Bind(&state.list).Focusable().TabStop(false)
				first = false
			}
			if strings.HasPrefix(string(pos), "top") {
				list.Top(offset.Top)
			} else {
				list.Bottom(offset.Bottom)
			}
			if strings.HasSuffix(string(pos), "left") {
				list.Left(offset.Left)
			} else if strings.HasSuffix(string(pos), "center") {
				list.LeftPercent(50).TranslatePercent(-50, 0)
			} else {
				list.Right(offset.Right)
			}
			list.Children(func() {
				displacement := float32(0)
				// Build older rows first so the front toast paints and receives input last.
				for i := len(entries) - 1; i >= 0; i-- {
					item := entries[i]
					e := item.entry
					if expanded {
						displacement = 0
						for j := 0; j < i; j++ {
							displacement += entries[j].entry.height + gap
						}
					} else {
						displacement = float32(i) * gap
					}
					var root ui.Element
					var parts ui.ToastParts
					if item.native {
						parts = ui.ToastBase(c, e.model)
						root = parts.Root
					} else {
						root = ui.Box(c.Key(toastAppKey{e.key})).Role(ui.RoleListItem).Label(e.model.Title).Focusable().TabStop(true).Bind(&e.handle)
					}
					root.Absolute().Left(0).Right(0).ZIndex(len(entries) - i)
					if strings.HasPrefix(string(pos), "top") {
						root.Top(displacement).TransformOrigin(.5, 0)
					} else {
						root.Bottom(displacement).TransformOrigin(.5, 1)
					}
					hidden := i >= limit
					inert := hidden || !e.closing.IsZero()
					root.AccessHidden(inert).IgnoreInput(inert).Disabled(inert).DimDisabled(false).TabStop(!inert)
					if hidden {
						root.Opacity(0)
					}
					if !expanded && i > 0 {
						root.Height(entries[0].entry.height).Scale(max(.05, 1-float32(i)*.05), max(.05, 1-float32(i)*.05))
					}
					exitY := float32(0)
					if !e.closing.IsZero() {
						age := c.Now().Sub(e.closing)
						// Sonner removes at 200ms even though normal CSS transitions
						// last 400ms. Collapsed back rows use opacity200/transform500.
						duration := 400 * time.Millisecond
						alphaDuration := duration
						if !expanded && i > 0 {
							duration = 500 * time.Millisecond
							alphaDuration = 200 * time.Millisecond
						}
						ease := ui.CubicBezier(.25, .1, .25, 1)
						if e.swipeExit {
							duration = 200 * time.Millisecond
							alphaDuration = duration
							ease = ui.CubicBezier(0, 0, .58, 1)
						}
						travel := ease(min(1, float32(age)/float32(duration)))
						opacity := 1 - ease(min(1, float32(age)/float32(alphaDuration)))
						if motion.Reduced(c) {
							travel = 1
							opacity = 0
						} else {
							c.After(time.Second / 60)
						}
						root.Opacity(opacity)
						if !e.swipeExit {
							exitY = e.height * travel
							if !expanded && i > 0 {
								exitY = e.height * .4 * travel
							} else if strings.HasPrefix(string(pos), "top") {
								exitY = -exitY
							}
						}
					}
					toastStyle(c, root, e, props, item.native, parts, state, !expanded && i > 0)
					if !item.native {
						toastSwipe(c, root, e, props, pos)
						root.Translate(e.drag.X, e.drag.Y+exitY)
					}
					services := c.Services()
					root.Draw(func(_ *ui.Painter, r ui.Rect) {
						if (expanded || i == 0) && r.H > 0 && math.Abs(float64(e.height-r.H)) > .1 {
							e.height = r.H
							services.Invalidate()
						}
					})
					if !motion.Reduced(c) {
						y := float32(100)
						if strings.HasPrefix(string(pos), "top") {
							y = -100
						}
						root.Transition(ui.ElementTransition{Duration: 400 * time.Millisecond, ExitDuration: 200 * time.Millisecond, Ease: ui.CubicBezier(.25, .1, .25, 1), Enter: &ui.Motion{YPercent: y}})
					}
				}
			})
		}
		// Flush dismissals generated by this frame's actual native button events.
		state.advance(c, props, pause)
	})
}

func toastStyle(c *ui.Context, root ui.Element, e *toastEntry, props ToasterProps, native bool, parts ui.ToastParts, state *ToasterState, back bool) {
	t := theme.TokensOf(c)
	bg, fg, border, desc := toastColors(t, e, props)
	family := c.Theme().Font
	if family == "" {
		family = "system-ui"
	}
	if e.options.Unstyled {
		root.Font(family).Children(func() {
			if e.options.Content != nil {
				e.options.Content(c)
			} else {
				ui.Text(c, e.model.Title)
				if e.model.Description != "" {
					ui.Text(c, e.model.Description)
				}
			}
		})
		return
	}
	root.FocusRingColor(ui.RGBA(0, 0, 0, .2)).FocusRingOffset(0).Row().AlignItems(ui.Center).Gap(6).Padding(16).Radius(t.Radius*1.8).Border(1, border).Background(bg).TextColor(fg).Font(family).FontSize(13).LineHeight(1.5).Shadow(0, 4, 12, 0, ui.RGBA(0, 0, 0, .1))
	root.Children(func() {
		content := ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Center).Gap(6)
		if back {
			content.Opacity(0).IgnoreInput(true).AccessHidden(true)
		}
		content.Children(func() {
			icon := toastIcon(e, props)
			if !e.options.HideIcon && (icon != nil || e.model.Type == "loading") {
				ui.Box(c).Size(16, 16).MarginStart(-3).MarginEnd(4).Shrink(0).Center().Children(func() {
					if icon != nil {
						ui.Icon(c, icon).Size(16, 16).MarginStart(-1).AccessHidden(true)
					} else {
						indicator.Spinner(c, indicator.Props{Size: 16, Decorative: true})
					}
				})
			}
			ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
				if e.options.Content != nil {
					e.options.Content(c)
				} else {
					if e.model.Title != "" {
						ui.Text(c, e.model.Title).FontSize(13).FontWeight(500).LineHeight(1.5)
					}
					if e.model.Description != "" {
						ui.Text(c, e.model.Description).FontSize(13).FontWeight(400).LineHeight(1.4).TextColor(desc)
					}
				}
			})
			dismissible := toastBool(e.options.Dismissible, true)
			if e.options.Cancel != "" {
				b := toastActionButton(c, ui.ButtonBase(c), e.options.Cancel, fg, ui.RGBA(0, 0, 0, .08))
				if t.Dark {
					b.Background(ui.RGBA(255, 255, 255, .3))
				}
				b.Disabled(!dismissible).DimDisabled(false)
				if b.Clicked() && dismissible {
					if e.options.OnCancel != nil {
						e.options.OnCancel()
					}
					if native {
						c.CloseToast(e.model.ID)
					} else {
						e.pending = toastCancel
						c.Invalidate()
					}
				}
			}
			if e.model.Action != "" {
				var b ui.Element
				if native {
					b = parts.ActionButton()
				} else {
					b = ui.ButtonBase(c)
				}
				b = toastActionButton(c, b, e.model.Action, bg, fg)
				if !native && b.Clicked() {
					model, options := e.model, e.options
					event := ToastActionEvent{Toast: model}
					if model.OnAction != nil {
						model.OnAction()
					}
					if options.Action != nil {
						options.Action(&event)
					}
					if !event.prevented {
						e.pending = toastAction
						c.Invalidate()
					}
				}
			}
		})
		if toastBool(e.options.CloseButton, props.CloseButton) && e.model.Type != "loading" && !back {
			var close ui.Element
			if native {
				close = parts.CloseButton()
			} else {
				close = ui.ButtonBase(c)
			}
			label := props.CloseLabel
			if label == "" {
				label = "Close toast"
			}
			closeFg, closeBorder := fg, border
			rich := toastBool(e.options.RichColors, props.RichColors) && (e.model.Type == "success" || e.model.Type == "info" || e.model.Type == "warning" || e.model.Type == "error")
			if !t.Dark && !rich {
				closeFg = ui.RGB(23, 23, 23)
				closeBorder = ui.RGB(237, 237, 237)
			}
			close.Label(label).FocusRingColor(ui.RGBA(0, 0, 0, .2)).FocusRingOffset(0).Absolute().Top(-7).Start(-7).Size(20, 20).Radius(10).Border(1, closeBorder).Background(bg).TextColor(closeFg).Center().Disabled(!toastBool(e.options.Dismissible, true)).DimDisabled(false)
			if close.Hovered() {
				close.Background(ui.RGB(248, 248, 248)).Border(1, ui.RGB(232, 232, 232))
				if t.Dark {
					close.Background(ui.RGB(31, 31, 31)).Border(1, ui.RGB(64, 64, 64))
				}
			}
			icon := props.Icons.Close
			if icon == nil {
				icon = toastCloseIcon
			}
			close.Children(func() { ui.Icon(c, icon).Size(12, 12).AccessHidden(true) })
			if !native && close.Clicked() {
				e.pending = toastDismiss
				c.Invalidate()
			}
		}
	})
}
func toastActionButton(c *ui.Context, b ui.Element, label string, fg, bg ui.Color) ui.Element {
	b.Label(label).Height(24).PaddingX(8).Radius(4).Shrink(0).Center().Background(bg).TextColor(fg).FontSize(12).FontWeight(500)
	if b.FocusVisible() {
		b.FocusRing(true).FocusRingColor(ui.RGBA(0, 0, 0, .4)).FocusRingOffset(0)
	}
	return b.Children(func() { ui.Text(c, label) })
}
func toastIcon(e *toastEntry, p ToasterProps) *ui.SVG {
	if e.options.Icon != nil {
		return e.options.Icon
	}
	switch e.model.Type {
	case "success":
		if p.Icons.Success != nil {
			return p.Icons.Success
		}
		return toastSuccessIcon
	case "info":
		if p.Icons.Info != nil {
			return p.Icons.Info
		}
		return toastInfoIcon
	case "warning":
		if p.Icons.Warning != nil {
			return p.Icons.Warning
		}
		return toastWarningIcon
	case "error":
		if p.Icons.Error != nil {
			return p.Icons.Error
		}
		return toastErrorIcon
	case "loading":
		return p.Icons.Loading
	}
	return nil
}
func toastColors(t *theme.Tokens, e *toastEntry, p ToasterProps) (bg, fg, border, desc ui.Color) {
	bg, fg, border = t.Popover, t.PopoverForeground, t.Border
	desc = ui.RGB(63, 63, 63)
	if t.Dark {
		desc = ui.RGB(232, 232, 232)
	}
	if p.Invert || e.options.Invert {
		if t.Dark {
			bg = ui.RGB(255, 255, 255)
			fg = ui.RGB(23, 23, 23)
			border = ui.RGB(242, 242, 242)
		} else {
			bg = ui.RGB(0, 0, 0)
			fg = ui.RGB(252, 252, 252)
			border = ui.RGB(51, 51, 51)
		}
	}
	if toastBool(e.options.RichColors, p.RichColors) {
		var palette [3][3]float64
		ok := true
		if t.Dark {
			switch e.model.Type {
			case "success":
				palette = [3][3]float64{{150, 100, 6}, {147, 100, 12}, {150, 86, 65}}
			case "info":
				palette = [3][3]float64{{215, 100, 6}, {223, 43, 17}, {216, 87, 65}}
			case "warning":
				palette = [3][3]float64{{64, 100, 6}, {60, 100, 9}, {46, 87, 65}}
			case "error":
				palette = [3][3]float64{{358, 76, 10}, {357, 89, 16}, {358, 100, 81}}
			default:
				ok = false
			}
		} else {
			switch e.model.Type {
			case "success":
				palette = [3][3]float64{{143, 85, 96}, {145, 92, 87}, {140, 100, 27}}
			case "info":
				palette = [3][3]float64{{208, 100, 97}, {221, 91, 93}, {210, 92, 45}}
			case "warning":
				palette = [3][3]float64{{49, 100, 97}, {49, 91, 84}, {31, 92, 45}}
			case "error":
				palette = [3][3]float64{{359, 100, 97}, {359, 100, 94}, {360, 100, 45}}
			default:
				ok = false
			}
		}
		if ok {
			bg = toastHSL(palette[0])
			border = toastHSL(palette[1])
			fg = toastHSL(palette[2])
			desc = fg
		}
	}
	return
}
func toastHSL(v [3]float64) ui.Color {
	h, s, l := v[0]/60, v[1]/100, v[2]/100
	a := s * math.Min(l, 1-l)
	channel := func(n float64) uint8 {
		k := math.Mod(n+h*2, 12)
		return uint8(math.Round(255 * (l - a*math.Max(-1, math.Min(math.Min(k-3, 9-k), 1)))))
	}
	return ui.RGB(channel(0), channel(8), channel(4))
}
func toastFinite(n float32) bool { return !math.IsNaN(float64(n)) && !math.IsInf(float64(n), 0) }
func toastSafeInsets(v ToastInsets) ToastInsets {
	fields := []*float32{&v.Top, &v.Right, &v.Bottom, &v.Left}
	for _, p := range fields {
		if !toastFinite(*p) || *p < 0 {
			*p = 0
		}
	}
	return v
}
