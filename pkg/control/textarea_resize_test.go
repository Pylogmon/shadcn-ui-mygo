package control

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func textareaResizeAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	e, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %v %q", role, label)
	}
	return e
}

func textareaResizeTester(value *string, p *TextareaProps, resize *TextareaResizeProps, rtl bool, changed *int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		root := ui.Column(c).Width(360).Padding(20).Gap(12)
		if rtl {
			root.Direction(ui.RightToLeft)
		}
		root.Children(func() {
			parts := TextareaWith(c, value, *p, *resize)
			if changed != nil && parts.Root.Changed() {
				*changed++
			}
		})
	}, 500, 500)
}

func TestTextareaResizeAutoDragCaptureAndCancelRestore(t *testing.T) {
	value := strings.Repeat("A native line\n", 6) + "Last line"
	state := TextareaResizeState{}
	p := TextareaProps{Label: "Notes"}
	var heights []float32
	resize := TextareaResizeProps{State: &state, MaxHeight: 360, OnResize: func(height float32) { heights = append(heights, height) }}
	changed := 0
	r := textareaResizeTester(&value, &p, &resize, false, &changed)
	initial := textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds
	h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
	if state.Height != 0 || initial.H <= 64 || h.Now != float64(initial.H) {
		t.Fatalf("auto sizing/AX state=%g field=%+v splitter=%+v", state.Height, initial, h)
	}
	x, y := h.Bounds.X+h.Bounds.W/2, h.Bounds.Y+h.Bounds.H/2
	r.Press(x, y)
	r.Move(x+100, y+50)
	if math.Abs(float64(state.Height-initial.H-50)) > .25 {
		t.Fatalf("captured resize=%g start=%g", state.Height, initial.H)
	}
	r.CancelPress()
	if state.Height != 0 || textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds.H != initial.H {
		t.Fatalf("cancel did not restore auto: %g", state.Height)
	}
	if len(heights) != 2 || heights[0] != initial.H+50 || heights[1] != 0 || changed != 2 {
		t.Fatalf("resize/cancel callbacks=%v Changed=%d", heights, changed)
	}
	h = textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
	x, y = h.Bounds.X+8, h.Bounds.Y+8
	r.Press(x, y)
	r.Move(x+200, y+1000)
	r.Release(x+200, y+1000)
	if state.Height != 360 || textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds.H != 360 {
		t.Fatalf("capture/max=%g", state.Height)
	}
	n := len(heights)
	r.Move(x, y-100)
	r.Frame()
	if state.Height != 360 || len(heights) != n {
		t.Fatal("release retained active capture")
	}
}

func TestTextareaResizeKeyboardNativeAXAndJSONPersistence(t *testing.T) {
	value := "draft"
	state := TextareaResizeState{Height: 100}
	p := TextareaProps{Label: "Notes"}
	callbacks := 0
	resize := TextareaResizeProps{State: &state, MinHeight: 64, MaxHeight: 300, Step: 4, OnResize: func(float32) { callbacks++ }}
	r := textareaResizeTester(&value, &p, &resize, false, nil)
	h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
	if !h.Horizontal || h.Vertical || h.Min != 64 || h.Max != 300 || h.Now != 100 || h.Step != 4 || len(h.Controls) != 1 {
		t.Fatalf("splitter AX=%+v", h)
	}
	r.ClickAt(h.Bounds.X+8, h.Bounds.Y+8)
	r.Key(0, ui.KeyDown)
	r.Key(ui.Shift, ui.KeyDown)
	r.Key(0, ui.KeyUp)
	if state.Height != 140 || !r.Focused("Resize Notes") {
		t.Fatalf("keys height=%g", state.Height)
	}
	r.Key(0, ui.KeyHome)
	if state.Height != 64 {
		t.Fatalf("Home=%g", state.Height)
	}
	r.Key(0, ui.KeyEnd)
	if state.Height != 300 {
		t.Fatalf("End=%g", state.Height)
	}
	h = textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
	// The pinned native AccessDecrement bridge synthesizes Left for an LTR splitter.
	r.Key(0, ui.KeyLeft)
	if state.Height != 296 {
		t.Fatalf("native AX decrement=%g", state.Height)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var restored TextareaResizeState
	if err = json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	before := callbacks
	state = restored
	r.Frame()
	if !r.Focused("Resize Notes") || callbacks != before || textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds.H != 296 {
		t.Fatal("JSON load lost focus/height or reported user change")
	}
}

func TestTextareaResizeKeepsNativeEditorUndoCaretAndFieldTarget(t *testing.T) {
	value := "initial"
	state := TextareaResizeState{Height: 100}
	calls := 0
	created := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(320).Padding(20).Children(func() {
			Field(c, FieldProps{Label: "Message", Error: "Required"}, func() ui.Element {
				parts := TextareaWith(c, &value, TextareaProps{OnChange: func(string) { calls++ }}, TextareaResizeProps{State: &state, MaxHeight: 240})
				ui.Local(parts.Input, "resize-test-editor", func() bool { created++; return true })
				return parts.Root
			})
		})
	}, 500, 500)
	field := textareaResizeAccess(t, r, ui.RoleTextField, "Message")
	if !field.Multiline || !field.Invalid || len(field.ErrorMessage) != 1 {
		t.Fatalf("compound field target=%+v", field)
	}
	if created != 1 {
		t.Fatal("native editor was constructed more than once")
	}
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+20)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("typed text")
	if value != "typed text" || calls != 1 {
		t.Fatalf("native edit=%q callbacks=%d", value, calls)
	}
	h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize textarea")
	r.ClickAt(h.Bounds.X+8, h.Bounds.Y+8)
	r.Key(0, ui.KeyDown)
	field = textareaResizeAccess(t, r, ui.RoleTextField, "Message")
	if created != 1 {
		t.Fatal("resizing replaced native editor")
	}
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+20)
	r.Key(ui.Cmd, ui.KeyZ)
	if value != "initial" || calls != 2 {
		t.Fatalf("resize lost native undo=%q callbacks=%d", value, calls)
	}
	value = strings.Repeat("A native line\n", 40) + "Last"
	r.Frame()
	r.Key(ui.Cmd, ui.KeyEnd)
	caret, active := r.TextCaret()
	if !active || caret.Y < field.Bounds.Y || caret.Y+caret.H > field.Bounds.Y+field.Bounds.H {
		t.Fatalf("fixed native caret not scrolled into resized editor: caret=%+v field=%+v", caret, field.Bounds)
	}
}

func TestTextareaResizeReconfigurationCancelsOldGesture(t *testing.T) {
	for _, change := range []string{"external", "limits", "state", "disabled", "direction"} {
		t.Run(change, func(t *testing.T) {
			value := "draft"
			state := TextareaResizeState{Height: 100}
			replacement := TextareaResizeState{Height: 150}
			p := TextareaProps{Label: "Notes"}
			rtl := false
			callbacks := 0
			resize := TextareaResizeProps{State: &state, MaxHeight: 300, OnResize: func(float32) { callbacks++ }}
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				root := ui.Column(c).Width(320).Padding(20)
				if rtl {
					root.Direction(ui.RightToLeft)
				}
				root.Children(func() { TextareaWith(c, &value, p, resize) })
			}, 500, 500)
			h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
			x, y := h.Bounds.X+8, h.Bounds.Y+8
			r.Press(x, y)
			r.Move(x, y+20)
			if state.Height != 120 {
				t.Fatalf("drag=%g", state.Height)
			}
			switch change {
			case "external":
				state.Height = 180
			case "limits":
				resize.MaxHeight = 160
			case "state":
				resize.State = &replacement
			case "disabled":
				resize.Disabled = true
			case "direction":
				rtl = true
			}
			r.Frame()
			before := resize.State.Height
			beforeCallbacks := callbacks
			r.Move(x, y+100)
			r.CancelPress()
			if resize.State.Height != before || callbacks != beforeCallbacks {
				t.Fatalf("old gesture overwrote %s reconfiguration: height=%g want=%g callbacks=%d want=%d", change, resize.State.Height, before, callbacks, beforeCallbacks)
			}
		})
	}
}

func TestTextareaResizeDisabledReadOnlyRTLAndLocalState(t *testing.T) {
	value := "copy me"
	p := TextareaProps{Label: "Notes", ReadOnly: true}
	resize := TextareaResizeProps{MaxHeight: 200}
	r := textareaResizeTester(&value, &p, &resize, true, nil)
	field := textareaResizeAccess(t, r, ui.RoleTextField, "Notes")
	h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
	if h.Bounds.X != field.Bounds.X+1 {
		t.Fatalf("RTL grip is not logical end: field=%+v grip=%+v", field.Bounds, h.Bounds)
	}
	r.ClickAt(h.Bounds.X+8, h.Bounds.Y+8)
	r.Key(0, ui.KeyDown)
	if textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds.H != 72 {
		t.Fatal("read-only editor did not resize/local state did not persist")
	}
	resize.Disabled = true
	r.Frame()
	field = textareaResizeAccess(t, r, ui.RoleTextField, "Notes")
	h = textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
	if !h.Disabled || field.Disabled {
		t.Fatal("resize disabled affected editor")
	}
	p.ReadOnly = false
	r.Frame()
	r.ClickAt(field.Bounds.X+20, field.Bounds.Y+20)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("edited")
	if value != "edited" {
		t.Fatal("resize-disabled editor stopped native input")
	}
	resize.Disabled = false
	p.Disabled = true
	r.Frame()
	if !textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Disabled || !textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes").Disabled {
		t.Fatal("textarea disabled did not cover both targets")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("settled resize control scheduled idle work")
	}
}

func TestTextareaResizeInvalidBoundsStayFiniteAndDoNotDisableEditing(t *testing.T) {
	for _, resize := range []TextareaResizeProps{{MinHeight: 200, MaxHeight: 100}, {MinHeight: float32(math.NaN())}, {MaxHeight: float32(math.Inf(1))}, {Step: -1}, {State: &TextareaResizeState{Height: float32(math.NaN())}}} {
		value := ""
		p := TextareaProps{Label: "Notes"}
		r := textareaResizeTester(&value, &p, &resize, false, nil)
		field := textareaResizeAccess(t, r, ui.RoleTextField, "Notes")
		h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
		if !h.Disabled || !h.Invalid || math.IsNaN(h.Min) || math.IsInf(h.Max, 0) || math.IsNaN(h.Now) || math.IsInf(float64(field.Bounds.H), 0) {
			t.Fatalf("invalid config leaked nonfinite layout/AX: field=%+v splitter=%+v", field, h)
		}
		r.ClickAt(field.Bounds.X+20, field.Bounds.Y+20)
		r.Type("a")
		if value != "a" {
			t.Fatal("invalid resize constraints disabled native editor")
		}
	}
}

func TestTextareaResizeDefaultWindowBoundAndScaledNativeCaret(t *testing.T) {
	value := strings.Repeat("Line in the native editor\n", 50) + "Last line"
	state := TextareaResizeState{}
	p := TextareaProps{Label: "Notes", AutoFocus: true}
	callbacks := 0
	resize := TextareaResizeProps{State: &state, OnResize: func(float32) { callbacks++ }}
	r := textareaResizeTester(&value, &p, &resize, false, nil)
	field := textareaResizeAccess(t, r, ui.RoleTextField, "Notes")
	if state.Height != 0 || field.Bounds.H != 468 || textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes").Max != 468 {
		t.Fatalf("default native auto cap=%+v state=%g", field.Bounds, state.Height)
	}
	r.SetSize(500, 250)
	if state.Height != 0 || textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds.H != 218 || callbacks != 0 || !r.Focused("Notes") {
		t.Fatal("window bound replaced focus or published auto changes")
	}
	state.Height = 1000
	r.Frame()
	if state.Height != 218 || callbacks != 0 {
		t.Fatal("external fixed height was not silently normalized")
	}
	state.Height = 64
	resize.MinHeight = 32
	r.SetPreferences(ui.Preferences{TextScale: 2})
	field = textareaResizeAccess(t, r, ui.RoleTextField, "Notes")
	r.Key(ui.Cmd, ui.KeyEnd)
	caret, active := r.TextCaret()
	if field.Bounds.H < 66 || !active || caret.Y < field.Bounds.Y+1 || caret.Y+caret.H > field.Bounds.Y+field.Bounds.H-1 {
		t.Fatalf("scaled caret clipped in resized textarea: field=%+v caret=%+v", field.Bounds, caret)
	}
	resize.MaxHeight = 800
	state.Height = 600
	r.Frame()
	if textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds.H != 600 || callbacks != 0 {
		t.Fatal("explicit max was capped to the window or reported as a native resize")
	}
}

func TestTextareaResizeCallbackOverrideCancelsRemainingCapture(t *testing.T) {
	value := "draft"
	state := TextareaResizeState{}
	calls := 0
	p := TextareaProps{Label: "Notes"}
	resize := TextareaResizeProps{State: &state, MaxHeight: 300, OnResize: func(float32) { calls++; state.Height = 0; value = "caller replacement" }}
	r := textareaResizeTester(&value, &p, &resize, false, nil)
	h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
	x, y := h.Bounds.X+8, h.Bounds.Y+8
	r.Press(x, y)
	r.Move(x, y+30)
	if state.Height != 0 || value != "caller replacement" || calls != 1 {
		t.Fatal("resize callback override was overwritten or repeated")
	}
	r.Move(x, y+100)
	r.CancelPress()
	if state.Height != 0 || calls != 1 {
		t.Fatal("capture continued after callback replaced controlled height")
	}
	r.Frame()
	if calls != 1 {
		t.Fatal("callback repeated on idle frame")
	}
}

func TestTextareaResizeNativeCancelBlurAndRemovalRestoreExactState(t *testing.T) {
	for _, cause := range []string{"blur", "remove"} {
		t.Run(cause, func(t *testing.T) {
			value := "draft"
			state := TextareaResizeState{Height: 100}
			visible := true
			calls := 0
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Width(300).Padding(20).Children(func() {
					if visible {
						TextareaWith(c, &value, TextareaProps{Label: "Notes"}, TextareaResizeProps{State: &state, MaxHeight: 300, OnResize: func(float32) { calls++ }})
					}
				})
			}, 500, 500)
			h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes")
			x, y := h.Bounds.X+8, h.Bounds.Y+8
			r.Press(x, y)
			r.Move(x, y+30)
			if state.Height != 130 || calls != 1 {
				t.Fatal("drag did not begin")
			}
			if cause == "blur" {
				r.SetFocused(false)
			} else {
				visible = false
				r.Frame()
			}
			if state.Height != 100 {
				t.Fatalf("native %s cancellation did not restore exact state: %g", cause, state.Height)
			}
			if cause == "blur" {
				if calls != 2 {
					t.Fatalf("blur restoration callbacks=%d", calls)
				}
				r.SetFocused(true)
			} else {
				if calls != 1 {
					t.Fatal("removed editor invoked a retained view callback")
				}
				visible = true
				r.Frame()
			}
			r.Move(x, y+100)
			r.Release(x, y+100)
			if state.Height != 100 || value != "draft" {
				t.Fatal("canceled old capture reached the editor after remount/refocus")
			}
			r.Advance(time.Second)
			if r.FramePending() {
				t.Fatal("canceled editor scheduled idle frames")
			}
		})
	}
}

func TestTextareaResizeInstancesLocalStateAndAncestorDisableRemainIndependent(t *testing.T) {
	first, second := "first", "second"
	disabled := false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(300).Padding(20).Gap(20).Children(func() {
			TextareaWith(c, &first, TextareaProps{Label: "First"})
			ui.Column(c).Disabled(disabled).Children(func() { TextareaWith(c, &second, TextareaProps{Label: "Second"}) })
		})
	}, 500, 500)
	h := textareaResizeAccess(t, r, ui.RoleSplitter, "Resize First")
	r.ClickAt(h.Bounds.X+8, h.Bounds.Y+8)
	r.Key(0, ui.KeyDown)
	if textareaResizeAccess(t, r, ui.RoleTextField, "First").Bounds.H != 72 || textareaResizeAccess(t, r, ui.RoleTextField, "Second").Bounds.H != 64 {
		t.Fatal("nil-state instances shared height")
	}
	disabled = true
	r.Frame()
	if !textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Second").Disabled || !textareaResizeAccess(t, r, ui.RoleTextField, "Second").Disabled || textareaResizeAccess(t, r, ui.RoleSplitter, "Resize First").Disabled {
		t.Fatal("ancestor disabled did not scope to its own compound editor")
	}
	r.Key(0, ui.KeyDown)
	if textareaResizeAccess(t, r, ui.RoleTextField, "First").Bounds.H != 80 {
		t.Fatal("sibling reconfiguration replaced local state or lost grip focus")
	}
}

func TestTextareaResizeNativeQueuedKeysAccumulateBeforeDraw(t *testing.T) {
	// A desktop can coalesce several native key-down events into one render.
	// Exercise the same handler without an intervening Draw, including its
	// default auto height and native accessibility Left/Right key aliases.
	for _, initial := range []float32{0, 100} {
		state := TextareaResizeState{Height: initial}
		runtime := textareaResizeRuntime{config: textareaResizeConfig{state: &state, min: 64, max: 300, step: 8}, height: 100, last: initial}
		if !runtime.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyDown}) || !runtime.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyDown}) {
			t.Fatal("native height keys were not consumed")
		}
		if state.Height != 116 {
			t.Fatalf("two queued Down events from %g = %g", initial, state.Height)
		}
		runtime.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyLeft})
		if state.Height != 108 {
			t.Fatalf("queued accessibility decrement = %g", state.Height)
		}
		runtime.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyLeft})
		if state.Height != 100 {
			t.Fatalf("queued keys did not return to last painted height: %g", state.Height)
		}
	}
}

func TestTextareaResizeInvalidConfigurationPreservesCallerHeight(t *testing.T) {
	value := "draft"
	state := TextareaResizeState{Height: 100}
	p := TextareaProps{Label: "Notes"}
	calls := 0
	resize := TextareaResizeProps{State: &state, MinHeight: 200, MaxHeight: 100, OnResize: func(float32) { calls++ }}
	r := textareaResizeTester(&value, &p, &resize, false, nil)
	if state.Height != 100 || calls != 0 || !textareaResizeAccess(t, r, ui.RoleSplitter, "Resize Notes").Invalid {
		t.Fatal("invalid constraints overwrote caller-owned height")
	}
	resize.MinHeight = 64
	resize.MaxHeight = 300
	r.Frame()
	if state.Height != 100 || textareaResizeAccess(t, r, ui.RoleTextField, "Notes").Bounds.H != 100 || calls != 0 {
		t.Fatal("correcting constraints did not restore the caller height silently")
	}
}
