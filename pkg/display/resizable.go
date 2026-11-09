package display

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type ResizableProps struct {
	Vertical bool
	Min, Max float32
}

// Resizable preserves the two-pane native Split API: size, Min and Max are
// DIPs, rather than the percentage units of ResizablePanelGroup.
func Resizable(c *ui.Context, size *float32, p ResizableProps, first, second func()) ui.Element {
	if size == nil {
		return ui.Box(c).Invalid(true)
	}
	valid := panelFinite(float64(p.Min)) && panelFinite(float64(p.Max)) && (p.Min <= 0 || p.Max <= 0 || p.Min <= p.Max)
	if !panelFinite(float64(*size)) {
		*size = 40
	}
	if panelFinite(float64(p.Min)) && p.Min > 0 {
		*size = max(p.Min, *size)
	}
	if panelFinite(float64(p.Max)) && p.Max > 0 {
		*size = min(p.Max, *size)
	}
	if first == nil {
		first = func() {}
	}
	if second == nil {
		second = func() {}
	}
	var split ui.Element
	if p.Vertical {
		split = ui.SplitVertical(c, size, first, second)
	} else {
		split = ui.Split(c, size, first, second)
	}
	return split.Invalid(!valid).Disabled(!valid)
}

// ResizableLayout maps stable panel IDs to percentages of the available pane
// space (excluding the separators). It can be saved and restored with encoding/json.
type ResizableLayout map[string]float64

type ResizablePanelProps struct {
	ID, Label string
	// All sizes are percentages. MaxSize == 0 means 100; absent layout entries
	// use DefaultSize, or an equal share when DefaultSize is zero.
	DefaultSize, MinSize, MaxSize float64
	Collapsible                   bool
	CollapsedSize                 float64
	Content                       func(*ui.Context)
}

type ResizablePanelGroupProps struct {
	Panels                         []ResizablePanelProps
	Label                          string
	Vertical, Disabled, WithHandle bool
	// Step is the percentage moved by an arrow key, default 2. Shift moves ten
	// steps. Enter collapses/restores an adjacent collapsible panel.
	Step     float64
	OnLayout func(ResizableLayout)
}

type panelLimit struct {
	min, max, collapsed float64
	collapsible         bool
}
type panelDrag struct {
	boundary          int
	start, room, sign float64
	before            []float64
}
type panelGroupState struct {
	room     float64
	origins  map[int]float64
	drag     *panelDrag
	changed  bool
	config   string
	last     []float64
	expanded map[string]float64
}

// ResizablePanelGroup renders a controlled group with real native focusable
// splitters and captured pointer input. Neighbouring panels absorb a resize in
// order, continuing past a pane that reaches its constraint. External layouts
// are normalized without Changed/OnLayout; impossible constraints are exposed as
// invalid and disable resizing without overwriting the caller's layout.
func ResizablePanelGroup(c *ui.Context, layout *ResizableLayout, p ResizablePanelGroupProps) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Box(c).Fill().MinWidth(0).MinHeight(0).Role(ui.RoleGroup).Label(p.Label)
	s := ui.Local(root, "shadcn.resizable", func() panelGroupState {
		return panelGroupState{origins: make(map[int]float64), expanded: make(map[string]float64)}
	})
	limits, config, valid := panelLimits(p.Panels)
	valid = valid && layout != nil
	values := make([]float64, len(p.Panels))
	for i, panel := range p.Panels {
		v, exists := float64(0), false
		if layout != nil {
			v, exists = (*layout)[panel.ID]
		}
		if !exists || !panelFinite(v) {
			v = panel.DefaultSize
			if !panelFinite(v) || v <= 0 {
				v = 100 / float64(max(1, len(values)))
			}
		}
		values[i] = max(0, min(100, v))
	}
	if valid {
		var ok bool
		values, ok = panelNormalize(values, limits)
		valid = ok
	}
	if valid {
		panelWrite(layout, p.Panels, values)
	} else {
		for i := range values {
			values[i] = 100 / float64(max(1, len(values)))
			limits[i] = panelLimit{max: 100}
		}
	}
	disabled := p.Disabled || !valid || len(values) < 2
	root.Invalid(!valid)
	config = fmt.Sprintf("%t/%v/%s", p.Vertical, c.Direction(), config)
	if s.config != config || s.drag != nil && !panelEqual(values, s.last) || disabled {
		s.drag = nil
	}
	s.config = config
	s.last = slices.Clone(values)
	changed := s.changed
	s.changed = false
	step := p.Step
	if !panelFinite(step) || step <= 0 {
		step = 2
	}
	step = min(step, 100)
	write := func(next []float64) {
		if panelEqual(values, next) {
			return
		}
		for i, v := range values {
			if limits[i].collapsible && v > limits[i].collapsed+1e-7 {
				s.expanded[p.Panels[i].ID] = v
			}
		}
		values = next
		panelWrite(layout, p.Panels, next)
		s.last = slices.Clone(next)
		s.changed = true
	}
	rtl := !p.Vertical && c.Direction() == ui.RightToLeft
	root.Children(func() {
		row := ui.Row(c).Fill().MinWidth(0).MinHeight(0).AlignItems(ui.Stretch)
		if p.Vertical {
			row.Column()
		}
		row.Draw(func(_ *ui.Painter, r ui.Rect) {
			room := float64(r.W) - float64(max(0, len(values)-1))
			if p.Vertical {
				room = float64(r.H) - float64(max(0, len(values)-1))
			}
			room = max(0, room)
			if room != s.room {
				s.room = room
				c.Invalidate()
			}
		})
		row.Children(func() {
			var previousSeparator ui.Element
			for i, panel := range p.Panels {
				key := any(panel.ID)
				if !valid {
					key = i
				}
				collapsed := valid && limits[i].collapsible && values[i] <= limits[i].collapsed+1e-7 || values[i] <= 1e-7
				pane := ui.Box(c.Key(key)).Basis(0).Grow(float32(values[i] / 100)).Shrink(1).MinWidth(0).MinHeight(0).Clip().
					Role(ui.RoleGroup).Label(panelLabel(panel)).AccessHidden(collapsed).IgnoreInput(collapsed)
				pane.Children(func() {
					if panel.Content != nil {
						panel.Content(c)
					}
				})
				if i+1 == len(values) {
					if collapsed && pane.FocusWithin() && previousSeparator.Valid() && !disabled {
						previousSeparator.Focus()
					}
					continue
				}
				boundary := i
				label := "Resize " + panelLabel(panel)
				sep := ui.Box(c.Key("separator:" + panel.ID)).Shrink(0).Focusable().Role(ui.RoleSplitter).
					Label(label).Controls(pane).Disabled(disabled).Invalid(!valid).FocusRingColor(t.Ring).FocusRingOffset(0)
				if p.Vertical {
					sep.Height(8).MarginY(-3.5).Orientation(ui.Horizontal).Cursor(ui.CursorResizeNS).Direction(ui.LeftToRight)
				} else {
					sep.Width(8).MarginX(-3.5).Orientation(ui.Vertical).Cursor(ui.CursorResizeEW)
				}
				if collapsed && pane.FocusWithin() && !disabled {
					sep.Focus()
				}
				previousSeparator = sep
				minimum, maximum := panelBoundaryBounds(values, limits, boundary)
				now := panelSum(values[:i+1])
				sep.Range(minimum, maximum, now).Step(step).Value(fmt.Sprintf("%.1f%%", now))
				sep.Draw(func(painter *ui.Painter, r ui.Rect) {
					origin := float64(r.X)
					if p.Vertical {
						origin = float64(r.Y)
					}
					s.origins[boundary] = origin
					line := t.Border
					if sep.Hovered() || s.drag != nil && s.drag.boundary == boundary {
						line = t.Primary
					}
					if p.Vertical {
						painter.Fill(ui.Rect{X: r.X, Y: r.Y + (r.H-1)/2, W: r.W, H: 1}, line, 0)
					} else {
						painter.Fill(ui.Rect{X: r.X + (r.W-1)/2, Y: r.Y, W: 1, H: r.H}, line, 0)
					}
					if p.WithHandle {
						w, h := float32(3), float32(12)
						if p.Vertical {
							w, h = h, w
						}
						painter.Fill(ui.Rect{X: r.X + (r.W-w)/2, Y: r.Y + (r.H-h)/2, W: w, H: h}, line, 1)
					}
				})
				sep.HandleInput(func(ev ui.InputEvent) bool {
					if disabled || ev.Button != 0 && ev.Kind != ui.InputPointerMove {
						return false
					}
					axis := float64(ev.X)
					if p.Vertical {
						axis = float64(ev.Y)
					}
					axis += s.origins[boundary]
					switch ev.Kind {
					case ui.InputPointerDown:
						if s.room <= 0 {
							return false
						}
						sign := float64(1)
						if rtl {
							sign = -1
						}
						s.drag = &panelDrag{boundary: boundary, start: axis, room: s.room, sign: sign, before: slices.Clone(values)}
						return true
					case ui.InputPointerMove, ui.InputPointerUp:
						if s.drag == nil || s.drag.boundary != boundary {
							return false
						}
						if ev.Canceled {
							write(slices.Clone(s.drag.before))
						} else {
							delta := (axis - s.drag.start) * 100 / s.drag.room * s.drag.sign
							write(panelResize(s.drag.before, limits, boundary, delta))
						}
						if ev.Kind == ui.InputPointerUp {
							s.drag = nil
						}
						return true
					}
					return false
				})
				if disabled {
					continue
				}
				back, forth := ui.KeyLeft, ui.KeyRight
				if rtl {
					back, forth = forth, back
				}
				if p.Vertical {
					back, forth = ui.KeyUp, ui.KeyDown
				}
				delta := float64(0)
				switch {
				// MyGo's native AccessIncrement/Decrement bridge sends Right/Left,
				// or Up/Down for RTL. Accept those aliases while retaining physical
				// vertical Up/Down resize semantics. A vertical splitter's own LTR
				// direction keeps its native bridge on the unambiguous Right/Left.
				case sep.Shortcut(0, back), !p.Vertical && sep.Shortcut(0, ui.KeyDown), p.Vertical && sep.Shortcut(0, ui.KeyLeft):
					delta = -step
				case sep.Shortcut(0, forth), !p.Vertical && sep.Shortcut(0, ui.KeyUp), p.Vertical && sep.Shortcut(0, ui.KeyRight):
					delta = step
				case sep.Shortcut(ui.Shift, back):
					delta = -10 * step
				case sep.Shortcut(ui.Shift, forth):
					delta = 10 * step
				case sep.Shortcut(0, ui.KeyHome):
					delta = minimum - now
				case sep.Shortcut(0, ui.KeyEnd):
					delta = maximum - now
				case sep.Shortcut(0, ui.KeyEnter):
					target := i
					if !limits[target].collapsible {
						target++
					}
					if limits[target].collapsible {
						to := limits[target].collapsed
						if values[target] <= to+1e-7 {
							to = s.expanded[p.Panels[target].ID]
							if to <= limits[target].collapsed {
								to = max(limits[target].min, p.Panels[target].DefaultSize)
							}
						}
						write(panelSetSize(values, limits, target, to))
					}
				}
				if delta != 0 {
					write(panelResize(values, limits, boundary, panelKeyboardDelta(values, limits, boundary, delta)))
				}
			}
		})
	})
	if changed || s.changed {
		root.SetChanged()
		s.changed = false
		c.Invalidate()
		if p.OnLayout != nil && layout != nil {
			p.OnLayout(panelClone(*layout))
		}
	}
	return root
}

func panelLabel(p ResizablePanelProps) string {
	if p.Label != "" {
		return p.Label
	}
	return p.ID
}
func panelFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func panelSum(v []float64) float64 {
	s := float64(0)
	for _, n := range v {
		s += n
	}
	return s
}
func panelEqual(a, b []float64) bool {
	return slices.EqualFunc(a, b, func(a, b float64) bool { return math.Abs(a-b) < 1e-7 })
}
func panelClone(p ResizableLayout) ResizableLayout {
	r := make(ResizableLayout, len(p))
	for k, v := range p {
		r[k] = v
	}
	return r
}
func panelWrite(layout *ResizableLayout, panels []ResizablePanelProps, values []float64) {
	if *layout == nil {
		*layout = make(ResizableLayout, len(panels))
	}
	for i, p := range panels {
		(*layout)[p.ID] = values[i]
	}
}
func panelLimits(panels []ResizablePanelProps) ([]panelLimit, string, bool) {
	limits := make([]panelLimit, len(panels))
	ids := map[string]bool{}
	var config strings.Builder
	valid := len(panels) > 0
	minTotal, maxTotal := float64(0), float64(0)
	for i, p := range panels {
		lo, hi, collapsed := p.MinSize, p.MaxSize, p.CollapsedSize
		if hi == 0 {
			hi = 100
		}
		if p.ID == "" || ids[p.ID] || !panelFinite(lo) || !panelFinite(hi) || !panelFinite(collapsed) || lo < 0 || hi < lo || hi > 100 || collapsed < 0 || collapsed > lo {
			valid = false
		}
		ids[p.ID] = true
		limits[i] = panelLimit{min: lo, max: hi, collapsed: collapsed, collapsible: p.Collapsible}
		if p.Collapsible {
			minTotal += collapsed
		} else {
			minTotal += lo
		}
		maxTotal += hi
		fmt.Fprintf(&config, "%q/%g/%g/%g/%t;", p.ID, lo, hi, collapsed, p.Collapsible)
	}
	return limits, config.String(), valid && minTotal <= 100+1e-7 && maxTotal >= 100-1e-7
}
func panelNormalize(values []float64, limits []panelLimit) ([]float64, bool) {
	v := slices.Clone(values)
	lo, hi := make([]float64, len(v)), make([]float64, len(v))
	for i, l := range limits {
		lo[i], hi[i] = l.min, l.max
		if l.collapsible && v[i] < (l.collapsed+l.min)/2 {
			lo[i], hi[i] = l.collapsed, l.collapsed
		}
	}
	if panelSum(lo) > 100+1e-7 || panelSum(hi) < 100-1e-7 {
		// Collapsible panes have two valid ranges, not a continuous minimum.
		// Preserve the requested collapsed modes when possible; otherwise find
		// a feasible combination rather than marking a satisfiable group invalid.
		remainingMin, remainingMax := make([]float64, len(v)+1), make([]float64, len(v)+1)
		for i := len(v) - 1; i >= 0; i-- {
			l := limits[i]
			minimum := l.min
			if l.collapsible {
				minimum = l.collapsed
			}
			remainingMin[i] = remainingMin[i+1] + minimum
			remainingMax[i] = remainingMax[i+1] + l.max
		}
		var fit func(int, float64, float64) bool
		fit = func(index int, minimum, maximum float64) bool {
			if minimum+remainingMin[index] > 100+1e-7 || maximum+remainingMax[index] < 100-1e-7 {
				return false
			}
			if index == len(v) {
				return true
			}
			l := limits[index]
			preferred := l.collapsible && v[index] < (l.collapsed+l.min)/2
			for attempt := 0; attempt < 2; attempt++ {
				closed := preferred
				if attempt == 1 {
					if !l.collapsible {
						return false
					}
					closed = !closed
				}
				lo[index], hi[index] = l.min, l.max
				if closed {
					lo[index], hi[index] = l.collapsed, l.collapsed
				}
				if fit(index+1, minimum+lo[index], maximum+hi[index]) {
					return true
				}
			}
			return false
		}
		if !fit(0, 0, 0) {
			return v, false
		}
	}
	// Euclidean projection on the capped simplex, preserving requested shares
	// rather than letting panel order determine normalization.
	low, high := float64(-1000), float64(1000)
	for range 80 {
		mid := (low + high) / 2
		sum := float64(0)
		for i, x := range v {
			sum += max(lo[i], min(hi[i], x+mid))
		}
		if sum > 100 {
			high = mid
		} else {
			low = mid
		}
	}
	for i, x := range v {
		v[i] = max(lo[i], min(hi[i], x+(low+high)/2))
	}
	return v, true
}
func panelBoundaryBounds(v []float64, limits []panelLimit, boundary int) (float64, float64) {
	lo, hi, rlo, rhi := float64(0), float64(0), float64(0), float64(0)
	for i, l := range limits {
		minSize := l.min
		if l.collapsible {
			minSize = l.collapsed
		}
		if i <= boundary {
			lo += minSize
			hi += l.max
		} else {
			rlo += minSize
			rhi += l.max
		}
	}
	return max(lo, 100-rhi), min(hi, 100-rlo)
}
func panelResize(v []float64, limits []panelLimit, boundary int, delta float64) []float64 {
	if len(v) < 2 || !panelFinite(delta) || math.Abs(delta) < 1e-7 {
		return slices.Clone(v)
	}
	left, right := make([]int, 0, boundary+1), make([]int, 0, len(v)-boundary-1)
	for i := boundary; i >= 0; i-- {
		left = append(left, i)
	}
	for i := boundary + 1; i < len(v); i++ {
		right = append(right, i)
	}
	positive := delta > 0
	want := math.Abs(delta)
	for range len(v) + 2 {
		a, usedA := panelAdjust(v, limits, left, positive, want, true)
		b, usedB := panelAdjust(v, limits, right, !positive, want, true)
		if math.Abs(usedA-usedB) < 1e-7 {
			for _, i := range right {
				a[i] = b[i]
			}
			return a
		}
		next := max(usedA, usedB)
		if math.Abs(next-want) < 1e-7 {
			break
		}
		want = next
	}
	// A collapse jump needs room on both sides. If the opposite side cannot
	// absorb that jump, stop at the expanded constraint instead of crossing it.
	want = math.Abs(delta)
	for range len(v) + 2 {
		out, a := panelAdjust(v, limits, left, positive, want, false)
		r, b := panelAdjust(v, limits, right, !positive, want, false)
		if math.Abs(a-b) < 1e-7 {
			for _, i := range right {
				out[i] = r[i]
			}
			return out
		}
		next := min(a, b)
		if next >= want-1e-7 {
			break
		}
		want = next
	}
	return slices.Clone(v)
}
func panelAdjust(v []float64, limits []panelLimit, indices []int, increase bool, amount float64, snap bool) ([]float64, float64) {
	out := slices.Clone(v)
	left := amount
	for _, i := range indices {
		l := limits[i]
		lo := l.min
		if l.collapsible && (snap || v[i] <= l.collapsed+1e-7) {
			lo = l.collapsed
		}
		to := max(lo, v[i]-left)
		if increase {
			to = min(l.max, v[i]+left)
		}
		if l.collapsible && to > l.collapsed && to < l.min {
			if snap && to >= (l.min+l.collapsed)/2 {
				to = l.min
			} else if snap {
				to = l.collapsed
			} else {
				to = v[i]
			}
		}
		used := math.Abs(to - v[i])
		out[i] = to
		left -= used
		if left <= 1e-7 {
			break
		}
	}
	return out, amount - left
}
func panelSetSize(v []float64, limits []panelLimit, index int, to float64) []float64 {
	l := limits[index]
	to = max(l.collapsed, min(l.max, to))
	if to > l.collapsed {
		to = max(l.min, to)
	}
	delta := to - v[index]
	indices := make([]int, 0, len(v)-1)
	for distance := 1; distance < len(v); distance++ {
		if i := index + distance; i < len(v) {
			indices = append(indices, i)
		}
		if i := index - distance; i >= 0 {
			indices = append(indices, i)
		}
	}
	out, used := panelAdjust(v, limits, indices, delta < 0, math.Abs(delta), false)
	if math.Abs(used-math.Abs(delta)) > 1e-7 {
		return slices.Clone(v)
	}
	out[index] = to
	return out
}

// Discrete keys step across a collapsed-size gap in one action; unlike a drag,
// repeated keys do not accumulate an uncommitted pointer distance.
func panelKeyboardDelta(v []float64, limits []panelLimit, boundary int, delta float64) float64 {
	amount := math.Abs(delta)
	for _, pair := range []struct {
		index    int
		increase bool
	}{{boundary, delta > 0}, {boundary + 1, delta < 0}} {
		i, l := pair.index, limits[pair.index]
		if !l.collapsible {
			continue
		}
		if pair.increase && v[i] <= l.collapsed+1e-7 || !pair.increase && v[i] <= l.min+1e-7 && v[i] > l.collapsed+1e-7 {
			amount = max(amount, l.min-l.collapsed)
		}
	}
	return math.Copysign(amount, delta)
}
