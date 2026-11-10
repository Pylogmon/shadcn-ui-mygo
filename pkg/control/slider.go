package control

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type SliderProps struct {
	// Min and Max default to 0 and 100 when both are zero. Reversed bounds
	// are accepted. Step <= 0 permits continuous pointer input.
	Min, Max, Step float64
	Label          string
	Disabled       bool
	// Invalid marks application validation independently of impossible numeric
	// constraints. It preserves keyboard/pointer editing of a valid domain.
	Invalid  bool
	Vertical bool
	// Adjacent thumbs cannot cross. The larger of these two gaps applies;
	// a positive Step rounds the gap up to a whole number of steps.
	MinGap                float64
	MinStepsBetweenValues int
	// Thumbs keep their ordered slot identity, even when values meet. IDs
	// optionally give slots explicit, unique identities across list edits.
	ThumbIDs    []string
	ThumbLabels []string
	// FormatValue supplies accessible value text; the numeric range remains
	// available to assistive technology independently of this text.
	FormatValue func(index int, value float64) string
}

// Slider is a single native focusable slider thumb with shadcn styling.
func Slider(c *ui.Context, value *float64, p SliderProps) ui.Element {
	return slider(c, sliderBinding{
		read: func() []float64 {
			if value == nil {
				return nil
			}
			return []float64{*value}
		},
		write: func(v []float64) { *value = v[0] },
	}, p)
}

// RangeSlider selects two ordered endpoints. Both pointers must be distinct.
func RangeSlider(c *ui.Context, low, high *float64, p SliderProps) ui.Element {
	return slider(c, sliderBinding{
		read: func() []float64 {
			if low == nil || high == nil || low == high {
				return nil
			}
			return []float64{*low, *high}
		},
		write: func(v []float64) { *low, *high = v[0], v[1] },
	}, p)
}

// SliderValues selects an arbitrary number of ordered values. External values
// are sorted, clamped and snapped without reporting a user change. An empty
// slice renders an inert track. Impossible gap constraints are disabled and
// exposed as invalid rather than silently reducing the requested gap.
func SliderValues(c *ui.Context, values *[]float64, p SliderProps) ui.Element {
	return slider(c, sliderBinding{
		read: func() []float64 {
			if values == nil {
				return nil
			}
			return slices.Clone(*values)
		},
		write: func(v []float64) { copy(*values, v) },
	}, p)
}

type sliderBinding struct {
	read  func() []float64
	write func([]float64)
}

type sliderState struct {
	width, height float32
	dragging      int
	focused       int
	focus         bool
	changed       bool
	before        []float64
}

type sliderDomain struct {
	lo, hi, step, gap float64
	invalid           bool
}

func slider(c *ui.Context, binding sliderBinding, p SliderProps) ui.Element {
	t := theme.TokensOf(c)
	values := binding.read()
	d := sliderLimits(p, len(values))
	if !d.invalid && len(values) > 0 {
		values = d.normalize(values)
		binding.write(values)
	}
	// Even an invalid domain has finite display/accessibility metadata.
	if d.invalid {
		for i, v := range values {
			values[i] = finiteClamp(v, d.lo, d.hi)
		}
	}
	disabled := p.Disabled || d.invalid || d.lo == d.hi || len(values) == 0
	invalid := fieldInputInvalid(c, p.Invalid) || d.invalid
	root := ui.Box(c).Role(ui.RoleGroup).Label(p.Label).Disabled(disabled).
		Invalid(invalid).Shrink(0).Cursor(ui.CursorPointer)
	if p.Vertical {
		root.Width(20).Height(160).Orientation(ui.Vertical)
	} else {
		root.Height(20).FillWidth().Orientation(ui.Horizontal)
	}
	if disabled {
		root.DimDisabled(false).Opacity(.5).Cursor(ui.CursorNotAllowed)
	}
	st := ui.Local(root, "shadcn.slider", func() sliderState { return sliderState{dragging: -1} })
	if st.changed {
		root.SetChanged()
		st.changed = false
	}
	if disabled || st.dragging >= len(values) {
		st.dragging, st.before = -1, nil
	}
	rtl := c.Direction() == ui.RightToLeft && !p.Vertical
	count := len(values)
	// The native input engine supplies pointer capture, including cancellation
	// and moves beyond the control. Only numeric state crosses frame boundaries.
	root.HandleInput(func(ev ui.InputEvent) bool {
		if disabled || ev.Button != 0 && ev.Kind != ui.InputPointerMove {
			return false
		}
		current := binding.read()
		if len(current) != count {
			st.dragging, st.before = -1, nil
			return false
		}
		at := func() float64 {
			axis, size := ev.X, st.width
			if p.Vertical {
				axis, size = ev.Y, st.height
			}
			f := float64((axis - 8) / max(float32(1), size-16))
			if p.Vertical || rtl {
				f = 1 - f
			}
			f = max(0, min(1, f))
			return d.lo*(1-f) + d.hi*f
		}
		change := func(v []float64) {
			if !slices.Equal(current, v) {
				binding.write(v)
				st.changed = true
			}
		}
		switch ev.Kind {
		case ui.InputPointerDown:
			st.before = slices.Clone(current)
			st.dragging = nearestSliderThumb(current, at(), st.focused, d)
			st.focused, st.focus = st.dragging, true
			change(d.move(current, st.dragging, at()))
			return true
		case ui.InputPointerMove:
			if st.dragging < 0 {
				return false
			}
			change(d.move(current, st.dragging, at()))
			return true
		case ui.InputPointerUp:
			if st.dragging < 0 {
				return false
			}
			if ev.Canceled {
				change(st.before)
			} else {
				change(d.move(current, st.dragging, at()))
			}
			st.dragging, st.before = -1, nil
			return true
		}
		return false
	})
	root.Children(func() {
		for i := range values {
			key := any(i)
			if i < len(p.ThumbIDs) && p.ThumbIDs[i] != "" {
				key = p.ThumbIDs[i]
			}
			thumb := motion.ApplyTransition(c, ui.Box(c).Key(key).Absolute().Size(16, 16).Focusable().
				Role(ui.RoleSlider).Label(sliderThumbLabel(p, i, len(values))).
				Disabled(disabled).DimDisabled(false).Invalid(invalid).FocusRing(false).
				Cursor(ui.CursorPointer).Radius(8).Border(1, t.Primary).
				Background(ui.RGB(255, 255, 255)).Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, .05)))
			if thumb.Focused() && !st.focus {
				st.focused = i
			}
			if st.focus && st.focused == i && !disabled {
				thumb.Focus()
				st.focus = false
			}
			if !disabled {
				back, forth := ui.KeyLeft, ui.KeyRight
				if rtl {
					back, forth = forth, back
				}
				step := d.step
				if step <= 0 {
					step = d.hi/100 - d.lo/100
				}
				page := max(step, d.hi/10-d.lo/10)
				v, next, handled := values[i], values[i], true
				from, to := d.bounds(values, i)
				switch {
				case thumb.Shortcut(0, back), thumb.Shortcut(0, ui.KeyDown):
					next = v - step
				case thumb.Shortcut(0, forth), thumb.Shortcut(0, ui.KeyUp):
					next = v + step
				case thumb.Shortcut(0, ui.KeyPageDown), thumb.Shortcut(ui.Shift, back), thumb.Shortcut(ui.Shift, ui.KeyDown):
					next = v - page
				case thumb.Shortcut(0, ui.KeyPageUp), thumb.Shortcut(ui.Shift, forth), thumb.Shortcut(ui.Shift, ui.KeyUp):
					next = v + page
				case thumb.Shortcut(0, ui.KeyHome):
					next = from
				case thumb.Shortcut(0, ui.KeyEnd):
					next = to
				default:
					handled = false
				}
				if handled {
					updated := d.move(values, i, next)
					if !slices.Equal(values, updated) {
						values = updated
						binding.write(values)
						root.SetChanged()
					}
				}
			}
			from, to := d.bounds(values, i)
			keyStep := d.step
			if keyStep <= 0 {
				keyStep = d.hi/100 - d.lo/100
			}
			thumb.Range(from, to, values[i]).Step(keyStep)
			text := strconv.FormatFloat(values[i], 'f', -1, 64)
			if p.FormatValue != nil {
				text = p.FormatValue(i, values[i])
			}
			thumb.Value(text)
			frac := float32(d.fraction(values[i]))
			if p.Vertical {
				frac = 1 - frac
				thumb.TopPercent(frac*100).Margin(frac*-16, 0, 0, 0).
					LeftPercent(50).MarginX(-8).Orientation(ui.Vertical)
			} else {
				if rtl {
					frac = 1 - frac
				}
				thumb.LeftPercent(frac*100).Margin(0, 0, 0, frac*-16).
					TopPercent(50).MarginY(-8).Orientation(ui.Horizontal)
			}
			if !disabled && (thumb.Hovered() || thumb.FocusVisible()) {
				thumb.Shadow(0, 0, 0, 4, t.Ring.Alpha(.5))
			}
		}
	})
	root.Draw(func(pt *ui.Painter, r ui.Rect) {
		st.width, st.height = r.W, r.H
		f0, f1 := float32(0), float32(0)
		if len(values) > 0 {
			f1 = float32(d.fraction(values[len(values)-1]))
			if len(values) > 1 {
				f0 = float32(d.fraction(values[0]))
			}
		}
		if p.Vertical {
			y0, y1 := r.Y+8+(1-f1)*max(float32(0), r.H-16), r.Y+8+(1-f0)*max(float32(0), r.H-16)
			if len(values) == 1 {
				y1 = r.Y + r.H
			}
			pt.Fill(ui.Rect{X: r.X + r.W/2 - 3, Y: r.Y, W: 6, H: r.H}, t.Muted, 3)
			pt.Fill(ui.Rect{X: r.X + r.W/2 - 3, Y: y0, W: 6, H: y1 - y0}, t.Primary, 3)
		} else {
			if rtl {
				f0, f1 = 1-f1, 1-f0
			}
			x0, x1 := r.X+8+f0*max(float32(0), r.W-16), r.X+8+f1*max(float32(0), r.W-16)
			// A single-value fill starts at the track edge, not the inset
			// thumb center used for the endpoints of a multi-thumb range.
			if len(values) == 1 {
				if rtl {
					x1 = r.X + r.W
				} else {
					x0 = r.X
				}
			}
			pt.Fill(ui.Rect{X: r.X, Y: r.Y + r.H/2 - 3, W: r.W, H: 6}, t.Muted, 3)
			pt.Fill(ui.Rect{X: x0, Y: r.Y + r.H/2 - 3, W: x1 - x0, H: 6}, t.Primary, 3)
		}
	})
	return root
}

func sliderThumbLabel(p SliderProps, i, count int) string {
	if i < len(p.ThumbLabels) && p.ThumbLabels[i] != "" {
		return p.ThumbLabels[i]
	}
	if count == 1 {
		return p.Label
	}
	suffix := fmt.Sprintf("Thumb %d", i+1)
	if count == 2 {
		suffix = []string{"minimum", "maximum"}[i]
	}
	if p.Label != "" {
		return p.Label + " " + suffix
	}
	return suffix
}

func sliderLimits(p SliderProps, count int) sliderDomain {
	d := sliderDomain{lo: p.Min, hi: p.Max, step: p.Step}
	if !finite(d.lo) || !finite(d.hi) {
		d.lo, d.hi, d.invalid = 0, 100, true
	}
	if d.lo == 0 && d.hi == 0 {
		d.hi = 100
	}
	if d.hi < d.lo {
		d.lo, d.hi = d.hi, d.lo
	}
	if !finite(d.step) || !finite(p.MinGap) {
		d.step, d.invalid = 0, true
	}
	d.step = max(0, d.step)
	d.gap = max(0, p.MinGap, float64(max(0, p.MinStepsBetweenValues))*d.step)
	if d.step > 0 && d.gap > 0 {
		// Multiplying a decimal step first can put an exact step count one
		// floating-point unit above its integer (3 * .1 / .1, for example).
		d.gap = math.Ceil(math.Nextafter(d.gap/d.step, math.Inf(-1))) * d.step
	}
	if !finite(d.gap) || count > 1 && d.gap > d.hi/float64(count-1)-d.lo/float64(count-1) {
		d.gap, d.invalid = 0, true
	}
	return d
}

func (d sliderDomain) normalize(v []float64) []float64 {
	for i := range v {
		v[i] = finiteClamp(v[i], d.lo, d.hi)
	}
	slices.Sort(v)
	// Walk backwards instead of multiplying gap by the number of slots:
	// the latter can overflow even when the endpoints themselves are finite.
	upper := make([]float64, len(v))
	limit := d.hi
	for i := len(v) - 1; i >= 0; i-- {
		upper[i] = limit
		limit -= d.gap
	}
	for i := range v {
		from, to := d.lo, upper[i]
		if i > 0 {
			from = max(from, v[i-1]+d.gap)
		}
		v[i] = d.snap(v[i], from, to)
	}
	return v
}

func (d sliderDomain) bounds(v []float64, i int) (float64, float64) {
	from, to := d.lo, d.hi
	if i > 0 {
		from = max(from, v[i-1]+d.gap)
	}
	if i+1 < len(v) {
		to = min(to, v[i+1]-d.gap)
	}
	return min(from, to), max(from, to)
}

func (d sliderDomain) move(v []float64, i int, to float64) []float64 {
	result := slices.Clone(v)
	from, until := d.bounds(v, i)
	result[i] = d.snap(to, from, until)
	return result
}

func (d sliderDomain) snap(v, from, to float64) float64 {
	v = finiteClamp(v, from, to)
	if d.step <= 0 || v == d.hi {
		return v
	}
	n := (v - d.lo) / d.step
	if !finite(n) {
		return v // A sub-ulp step or overflowing span cannot be represented.
	}
	snapped := d.lo + math.Round(n)*d.step
	if snapped < from {
		snapped = d.lo + math.Ceil((from-d.lo)/d.step)*d.step
	}
	if snapped > to {
		snapped = d.lo + math.Floor((to-d.lo)/d.step)*d.step
	}
	return finiteClamp(snapped, from, to)
}

func (d sliderDomain) fraction(v float64) float64 {
	scale := max(math.Abs(d.lo), math.Abs(d.hi), 1)
	span := d.hi/scale - d.lo/scale
	if span == 0 {
		return 0
	}
	return max(0, min(1, (v/scale-d.lo/scale)/span))
}

func nearestSliderThumb(v []float64, at float64, focused int, d sliderDomain) int {
	nearest, distance := 0, math.Inf(1)
	for i, value := range v {
		dist := math.Abs(d.fraction(at) - d.fraction(value))
		if dist < distance || dist == distance && i == focused {
			nearest, distance = i, dist
		}
	}
	return nearest
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func finiteClamp(v, lo, hi float64) float64 {
	if math.IsNaN(v) {
		return lo
	}
	return max(lo, min(hi, v))
}
