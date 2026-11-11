package control_test

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestNumberInputStepperSharesOneBorderWithEditor(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, wrapped := range []bool{false, true} {
			for _, scale := range []float32{1, 1.5} {
				t.Run(fmt.Sprintf("dark=%v/wrapped=%v/scale=%g", dark, wrapped, scale), func(t *testing.T) {
					n := 3.
					var bounds ui.Rect
					r := ui.NewTesterWith(func(c *ui.Context) {
						tokens := theme.Preset("orange", dark)
						c.SetTheme(tokens.Theme())
						ui.Column(c).Padding(20).Background(tokens.Background).Children(func() {
							build := func() ui.Element {
								return control.NumberInput(c, &n, control.NumberInputProps{Label: "Width", Stepper: &control.NumberInputStepperProps{
									Decrement: control.ButtonProps{Variant: control.Ghost, Size: control.IconExtraSmall},
									Increment: control.ButtonProps{Variant: control.Ghost, Size: control.IconExtraSmall},
								}})
							}
							var root ui.Element
							if wrapped {
								root = control.InputGroup(c, control.InputGroupProps{Prefix: "px"}, build)
							} else {
								root = build()
							}
							root.Width(240).DrawOver(func(_ *ui.Painter, r ui.Rect) { bounds = r })
						})
					}, 400, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true, TextScale: scale}})
					editor := numberText(t, r, "Width").Bounds
					more, ok := r.AccessByRole(ui.RoleButton, "Increase value")
					if !ok || more.Bounds.X < editor.X+editor.W || more.Bounds.X+more.Bounds.W >= bounds.X+bounds.W {
						t.Fatalf("step button left the shared input surface: root=%v editor=%v button=%+v", bounds, editor, more)
					}
					img := r.Image()
					x := int(more.Bounds.X + more.Bounds.W/2)
					for _, y := range []int{int(bounds.Y), int(bounds.Y + bounds.H - 1)} {
						outside := img.RGBAAt(int(bounds.X+bounds.W+4), y)
						if pixel := img.RGBAAt(x, y); max(pixel.R, outside.R)-min(pixel.R, outside.R) <= 2 && max(pixel.G, outside.G)-min(pixel.G, outside.G) <= 2 && max(pixel.B, outside.B)-min(pixel.B, outside.B) <= 2 {
							t.Fatalf("input border does not surround its step buttons: pixel=%v outside=%v root=%v", pixel, outside, bounds)
						}
					}
					// The editor must not keep an inner right border, including when
					// NumberInput itself is composed inside an existing InputGroup.
					x, y := int(editor.X+editor.W-1), int(editor.Y+editor.H/2)
					if img.RGBAAt(x, y) != img.RGBAAt(x-3, y) {
						t.Fatal("number editor retained a separate inner border")
					}
				})
			}
		}
	}
}

func TestNumberInputStepperSharesNativeCommitAndBoundaryRules(t *testing.T) {
	n, lo, hi := 0., 0., 5.
	empty, readOnly, disabled := false, false, false
	step := 2.
	var changes []float64
	var presence []bool
	changed, submitted := 0, 0
	r := typedTester(func(c *ui.Context) {
		ui.Column(c).Disabled(disabled).Children(func() {
			e := control.NumberInput(c, &n, control.NumberInputProps{
				Label: "Quantity", Min: &lo, Max: &hi, Step: step, Empty: &empty,
				ReadOnly: readOnly, AutoFocus: true, Stepper: &control.NumberInputStepperProps{},
				OnChange:      func(v float64) { changes = append(changes, v) },
				OnEmptyChange: func(v bool) { presence = append(presence, v) },
				OnSubmit:      func(float64) { submitted++ },
			})
			if e.Changed() {
				changed++
			}
		})
	})
	button := func(name string) ui.AccessInfo {
		t.Helper()
		a, ok := r.AccessByRole(ui.RoleButton, name)
		if !ok {
			t.Fatal("missing step button", name)
		}
		return a
	}
	assertDisabled := func(decrement, increment bool) {
		t.Helper()
		if button("Decrease value").Disabled != decrement || button("Increase value").Disabled != increment {
			t.Fatalf("wrong boundary state: decrement=%+v increment=%+v", button("Decrease value"), button("Increase value"))
		}
	}
	assertDisabled(true, false)
	typedRoleClick(t, r, ui.RoleButton, "Increase value")
	if n != 2 || changed != 1 || !slices.Equal(changes, []float64{2}) || !r.Focused("Quantity") {
		t.Fatalf("button failed native commit/focus: value=%g changes=%v Changed=%d", n, changes, changed)
	}
	r.Key(0, ui.KeyUp)
	typedRoleClick(t, r, ui.RoleButton, "Increase value")
	if n != 5 || !slices.Equal(changes, []float64{2, 4, 5}) || changed != 3 {
		t.Fatal("button and keyboard did not share clamping or change reporting", n, changes, changed)
	}
	assertDisabled(false, true)
	r.Key(0, ui.KeyUp)
	r.Frame()
	if len(changes) != 3 || changed != 3 || submitted != 0 {
		t.Fatal("boundary no-op emitted a change or submit")
	}
	r.Key(0, ui.KeyDown)
	replaceNumber(r, "-")
	if n != 3 || !numberText(t, r, "Quantity").Invalid {
		t.Fatal("incomplete draft replaced committed value")
	}
	typedRoleClick(t, r, ui.RoleButton, "Decrease value")
	if n != 1 || numberText(t, r, "Quantity").Value != "1" || numberText(t, r, "Quantity").Invalid {
		t.Fatal("step button did not canonicalize from the last committed value")
	}
	empty = true
	r.Frame()
	before := len(changes)
	typedRoleClick(t, r, ui.RoleButton, "Decrease value")
	if empty || n != 0 || len(changes) != before+1 || !slices.Equal(presence, []bool{false}) {
		t.Fatal("empty step did not start from zero and publish presence once", n, empty, presence)
	}
	assertDisabled(true, false)
	readOnly = true
	r.Frame()
	assertDisabled(true, true)
	r.Key(0, ui.KeyUp)
	if n != 0 || !numberText(t, r, "Quantity").ReadOnly {
		t.Fatal("readonly step changed the model")
	}
	readOnly, disabled = false, true
	r.Frame()
	assertDisabled(true, true)
	if !numberText(t, r, "Quantity").Disabled {
		t.Fatal("disabled ancestor failed to reach the real editor")
	}
	disabled, step = false, -1
	r.Frame()
	assertDisabled(true, true)
	if a := numberText(t, r, "Quantity"); !a.Invalid || !a.ReadOnly {
		t.Fatal("invalid step did not disable both editing paths")
	}
	step, n = 2, 1
	r.Frame()
	typedRoleClick(t, r, ui.RoleTextField, "Quantity")
	r.Key(ui.Cmd, ui.KeyA)
	before = len(changes)
	r.Compose("3", 1)
	assertDisabled(true, true)
	if n != 1 || len(changes) != before {
		t.Fatal("IME preedit was committed")
	}
	r.Key(0, ui.KeyEnter)
	if n != 3 || len(changes) != before+1 || submitted != 0 {
		t.Fatal("candidate Enter stepped or submitted instead of committing once")
	}
	assertDisabled(false, false)
}

func TestNumberInputStepperPresenceRedirectAndNativeTooltip(t *testing.T) {
	n, empty, reject := 7., true, true
	presence, changes := 0, 0
	icon := ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24"><path d="M12 4v16M4 12h16" stroke="currentColor"/></svg>`))
	r := typedTester(func(c *ui.Context) {
		control.NumberInput(c, &n, control.NumberInputProps{Label: "Optional", Empty: &empty,
			Stepper: &control.NumberInputStepperProps{
				Increment:        control.ButtonProps{Icon: icon, Label: "More", Variant: control.Outline},
				IncrementTooltip: "Add one item",
			},
			OnEmptyChange: func(bool) {
				presence++
				if reject {
					empty, n = true, 9
				}
			},
			OnChange: func(float64) { changes++ },
		})
	})
	more, ok := r.AccessByRole(ui.RoleButton, "More")
	if !ok || more.Description != "Add one item" || more.Disabled {
		t.Fatalf("icon step button lost its label or description: %+v", more)
	}
	r.Move(more.Bounds.X+more.Bounds.W/2, more.Bounds.Y+more.Bounds.H/2)
	r.Advance(time.Second)
	if !r.HasText("Add one item") {
		t.Fatal("stepper tooltip did not open on hover")
	}
	typedRoleClick(t, r, ui.RoleButton, "More")
	if !empty || n != 9 || presence != 1 || changes != 0 || numberText(t, r, "Optional").Value != "" {
		t.Fatal("stepper bypassed the presence callback's rejection/storage redirect")
	}
	reject = false
	typedRoleClick(t, r, ui.RoleButton, "More")
	if empty || n != 1 || presence != 2 || changes != 1 {
		t.Fatal("accepted empty step did not use the shared commit path")
	}
}

func TestNumberInputStepperFiniteExtremesAndInvalidLimits(t *testing.T) {
	n, lo, hi, step := 0., 0., 0., 1.
	r := typedTester(func(c *ui.Context) {
		control.NumberInput(c, &n, control.NumberInputProps{Label: "Bounded", Min: &lo, Max: &hi, Step: step, Stepper: &control.NumberInputStepperProps{}})
	})
	for _, state := range []struct{ n, lo, hi, step float64 }{
		{0, 0, 0, 1}, {0, 2, 1, 1}, {0, 0, 10, math.NaN()},
		{math.NaN(), 0, 10, 1}, {math.Inf(1), 0, 10, 1},
	} {
		n, lo, hi, step = state.n, state.lo, state.hi, state.step
		r.Frame()
		for _, label := range []string{"Decrease value", "Increase value"} {
			if b, ok := r.AccessByRole(ui.RoleButton, label); !ok || !b.Disabled {
				t.Fatalf("invalid/fixed state enabled %s: %+v", label, state)
			}
		}
	}
	n, lo, hi, step = math.MaxFloat64/2, -math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	r.Frame()
	typedRoleClick(t, r, ui.RoleButton, "Increase value")
	if n != math.MaxFloat64 {
		t.Fatal("finite step overflow did not clamp to the finite maximum", n)
	}
}
