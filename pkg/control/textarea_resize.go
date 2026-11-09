package control

import (
	"fmt"
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// TextareaResizeState is caller-owned and can be saved with encoding/json.
// Height is in DIPs; zero retains native content sizing within the limits.
type TextareaResizeState struct{ Height float32 }

type TextareaResizeProps struct {
	// A nil State uses state local to this TextareaWith instance.
	State *TextareaResizeState
	// Zero MinHeight is 64. The minimum also reserves one native text line.
	// Zero MaxHeight uses the window height minus 32, at least MinHeight.
	// An explicit MaxHeight may exceed the window. Zero Step is 8 DIPs;
	// Shift with an arrow key moves ten steps.
	MinHeight, MaxHeight, Step float32
	// Disabled disables resizing; the native editor remains usable.
	Disabled bool
	Label    string
	// OnResize reports user height changes, including a canceled drag's
	// restored height (zero when the editor previously used auto sizing).
	// External state/configuration changes do not invoke it. Invalid resize
	// configuration disables the handle without changing the caller's height.
	OnResize func(float32)
}

// TextareaParts keeps the native editor and its interactive sibling separate.
// Input retains all Textarea editing/selection/undo semantics. Handle is a
// horizontal native splitter. Root and Handle report Changed for user resizing.
type TextareaParts struct{ Root, Input, Handle ui.Element }

type textareaResizeStateKey struct{}
type textareaResizeRuntimeKey struct{}
type textareaResizeInputKey struct{}
type textareaResizeHandleKey struct{}
type textareaResizeConfig struct {
	state                                *TextareaResizeState
	min, max, step                       float32
	windowWidth, windowHeight, textScale float32
	rows, minRows, maxRows               int
	disabled, rtl, readOnly              bool
	label                                string
}
type textareaResizeDrag struct {
	y, height float64
	before    float32
	moved     bool
}
type textareaResizeRuntime struct {
	config          textareaResizeConfig
	initialized     bool
	height, originY float32
	last            float32
	pending         bool
	drag            *textareaResizeDrag
}

func textareaResizeFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

func textareaResizeLimits(c *ui.Context, p TextareaResizeProps) (minimum, maximum, step float32, valid bool) {
	valid = true
	minimum = p.MinHeight
	if minimum == 0 {
		minimum = 64
	} else if !textareaResizeFinite(minimum) || minimum < 0 {
		minimum = 64
		valid = false
	}
	scale := c.Preferences().TextScale
	if !textareaResizeFinite(scale) || scale <= 0 {
		scale = 1
	}
	// Padding and border match the actual Textarea native editor. A larger
	// system text size must still fit a caret in an explicitly sized editor.
	line := float64(inputFont(c, DefaultSize))*1.5*float64(scale) + 18
	if line > math.MaxFloat32 {
		line = 64
		valid = false
	}
	minimum = max(minimum, float32(line))
	_, windowHeight := c.Size()
	maximum = max(minimum, windowHeight-32)
	if !textareaResizeFinite(maximum) {
		maximum = minimum
		valid = false
	}
	if p.MaxHeight != 0 {
		if !textareaResizeFinite(p.MaxHeight) || p.MaxHeight < minimum {
			valid = false
		} else {
			maximum = p.MaxHeight
		}
	}
	step = p.Step
	if step == 0 {
		step = 8
	} else if !textareaResizeFinite(step) || step < 0 {
		step = 8
		valid = false
	}
	return
}

func (s *textareaResizeRuntime) write(height float32) {
	if s.config.state.Height == height {
		return
	}
	s.config.state.Height = height
	s.last = height
	s.pending = true
}

// input is invoked between frames. It only reads current numeric geometry and
// caller-owned state: no Context or Element is retained by the native handler.
func (s *textareaResizeRuntime) input(event ui.InputEvent) bool {
	config := s.config
	if config.disabled {
		s.drag = nil
		return false
	}
	if config.state.Height != s.last {
		// A caller can replace the height between native events, before a rebuild.
		s.drag = nil
		s.pending = false
		return false
	}
	clamp := func(height float64) float32 {
		return float32(max(float64(config.min), min(float64(config.max), height)))
	}
	switch event.Kind {
	case ui.InputPointerDown:
		if event.Button != 0 || !textareaResizeFinite(event.Y) || s.height <= 0 {
			return false
		}
		s.drag = &textareaResizeDrag{y: float64(s.originY) + float64(event.Y), height: float64(s.height), before: config.state.Height}
		return true
	case ui.InputPointerMove, ui.InputPointerUp:
		if s.drag == nil || event.Kind == ui.InputPointerUp && event.Button != 0 {
			return false
		}
		if event.Canceled {
			s.write(s.drag.before)
		} else if textareaResizeFinite(event.Y) {
			delta := float64(s.originY) + float64(event.Y) - s.drag.y
			if delta != 0 {
				s.drag.moved = true
			}
			// A click on the grip does not turn auto sizing into a fixed height.
			if s.drag.moved {
				s.write(clamp(s.drag.height + delta))
			}
		}
		if event.Kind == ui.InputPointerUp {
			s.drag = nil
		}
		return true
	case ui.InputKeyDown:
		if event.Mods&^ui.Shift != 0 {
			return false
		}
		delta := float64(config.step)
		if event.Mods&ui.Shift != 0 {
			delta *= 10
		}
		current := s.height
		if config.state.Height > 0 {
			// Native events can arrive in a batch before the next Draw. Each
			// key continues from the accepted controlled height.
			current = config.state.Height
		}
		next := float64(current)
		switch event.Key {
		case ui.KeyUp, ui.KeyLeft:
			next -= delta
		case ui.KeyDown, ui.KeyRight:
			next += delta
		case ui.KeyHome:
			next = float64(config.min)
		case ui.KeyEnd:
			next = float64(config.max)
		default:
			return false
		}
		s.drag = nil
		nextHeight := clamp(next)
		if nextHeight != current {
			s.write(nextHeight)
		}
		return true
	}
	return false
}

// TextareaWith adds a captured native resize grip without changing Textarea's
// existing return contract. It uses one real native editor for content sizing,
// scrolling, IME, selection and undo. The independent sibling splitter can be
// disabled while the editor remains editable. No content is built inside the
// native text input, which is a leaf in both layout and accessibility.
func TextareaWith(c *ui.Context, value *string, p TextareaProps, options ...TextareaResizeProps) TextareaParts {
	resize := TextareaResizeProps{}
	if len(options) > 0 {
		resize = options[0]
	}
	root := ui.Box(c).FillWidth().MinWidth(0).Role(ui.RoleGroup).Disabled(p.Disabled).DimDisabled(false)
	local := ui.Local(root, textareaResizeStateKey{}, func() TextareaResizeState { return TextareaResizeState{} })
	state := resize.State
	if state == nil {
		state = local
	}
	s := ui.Local(root, textareaResizeRuntimeKey{}, func() textareaResizeRuntime { return textareaResizeRuntime{} })
	minimum, maximum, step, valid := textareaResizeLimits(c, resize)
	fixed := textareaResizeFinite(state.Height) && state.Height > 0
	if !textareaResizeFinite(state.Height) || state.Height < 0 {
		valid = false
	}
	fixedHeight := minimum
	if fixed {
		fixedHeight = max(minimum, min(maximum, state.Height))
		if valid {
			state.Height = fixedHeight
		}
	}
	label := resize.Label
	if label == "" {
		label = "Resize textarea"
		if p.Label != "" {
			label = "Resize " + p.Label
		}
	}
	windowWidth, windowHeight := c.Size()
	config := textareaResizeConfig{state: state, min: minimum, max: maximum, step: step, rows: p.Rows, minRows: p.MinRows, maxRows: p.MaxRows,
		disabled: resize.Disabled || !valid || root.IsDisabled(), rtl: c.Direction() == ui.RightToLeft, readOnly: p.ReadOnly, label: label, windowWidth: windowWidth, windowHeight: windowHeight, textScale: c.Preferences().TextScale}
	same := s.initialized && s.config == config && s.last == state.Height
	changed := same && s.pending
	if !same || config.disabled {
		s.drag = nil
	}
	s.config = config
	s.initialized = true
	s.last = state.Height
	s.pending = false
	root.MinHeight(minimum).MaxHeight(maximum).Invalid(!valid)
	if fixed {
		root.Height(fixedHeight)
	}
	actualHeight := max(minimum, min(maximum, s.height))
	if fixed {
		actualHeight = fixedHeight
	}
	root.Draw(func(_ *ui.Painter, bounds ui.Rect) {
		if textareaResizeFinite(bounds.H) && bounds.H != s.height {
			s.height = bounds.H
			// Draw runs in this build. One correction frame publishes actual native
			// auto height to AX; stable geometry requests no frames.
			c.Invalidate()
		}
	})
	target := ui.Local(root, "shadcn.control.editor-target", func() ui.Handle { return ui.Handle{} })
	var input, handle ui.Element
	root.Children(func() {
		input = Textarea(c.Key(textareaResizeInputKey{}), value, p).MinHeight(minimum).MaxHeight(maximum).Shrink(0).Bind(target)
		if fixed {
			input.Height(fixedHeight)
		}
		handle = ui.Box(c.Key(textareaResizeHandleKey{})).Absolute().Bottom(1).Size(16, 16).ZIndex(1).
			Focusable().Role(ui.RoleSplitter).Orientation(ui.Horizontal).Direction(ui.LeftToRight).
			Label(label).Controls(input).Range(float64(minimum), float64(maximum), float64(actualHeight)).Step(float64(step)).
			Value(fmt.Sprintf("%.0f", actualHeight)).Disabled(config.disabled).Invalid(!valid).
			Cursor(ui.CursorResizeNS).FocusRingColor(theme.TokensOf(c).Ring).FocusRingOffset(0)
		if config.rtl {
			handle.Left(1)
		} else {
			handle.Right(1)
		}
		color := theme.TokensOf(c).MutedForeground
		if config.disabled {
			handle.Opacity(.5).Cursor(ui.CursorNotAllowed)
		}
		handle.Draw(func(painter *ui.Painter, bounds ui.Rect) {
			s.originY = bounds.Y
			// Six small marks leave a recognizable corner grip in either direction.
			for row := range 3 {
				for column := 0; column <= row; column++ {
					x := bounds.X + bounds.W - 4 - float32(column*4)
					if config.rtl {
						x = bounds.X + 2 + float32(column*4)
					}
					painter.Fill(ui.Rect{X: x, Y: bounds.Y + bounds.H - 4 - float32((row-column)*4), W: 2, H: 2}, color, 0)
				}
			}
		}).HandleInput(s.input)
	})
	if changed {
		root.SetChanged()
		handle.SetChanged()
		if resize.OnResize != nil {
			resize.OnResize(state.Height)
		}
		c.Invalidate()
	}
	return TextareaParts{Root: root, Input: input, Handle: handle}
}
