package control

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestChoiceSwitchUsesVegaTrackSize(t *testing.T) {
	on := false
	r := tester(func(c *ui.Context) { Switch(c, &on, SwitchProps{Label: ""}) })
	switchInfo, ok := r.AccessByRole(ui.RoleSwitch, "")
	if !ok {
		for _, info := range r.AccessAll() {
			if info.Role == ui.RoleSwitch {
				switchInfo, ok = info, true
				break
			}
		}
	}
	if !ok || switchInfo.Bounds.W != 32 || math.Abs(float64(switchInfo.Bounds.H-18.4)) > .01 {
		t.Fatalf("switch track does not use source32×18.4 sizing: %+v", switchInfo.Bounds)
	}
}

func choiceAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	info, ok := r.AccessByRole(role, name)
	if !ok {
		t.Fatalf("missing native %v %q: %+v", role, name, r.AccessAll())
	}
	return info
}

func choiceActivate(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	if err := r.Click(name); err != nil {
		t.Fatal(err)
	}
}

func TestChoiceCheckboxNativeMixedBooleanAndDisabledSemantics(t *testing.T) {
	state, disabled, calls := ui.CheckMixed, false, 0
	var reported []ui.CheckState
	r := tester(func(c *ui.Context) {
		CheckboxState(c, &state, "Select all", CheckboxProps{Disabled: disabled, Description: "Select every row", Invalid: true,
			OnStateChange: func(current ui.CheckState) { calls++; reported = append(reported, current) }})
		ui.Button(c, "Continue")
	})
	info := choiceAccess(t, r, ui.RoleCheckBox, "Select all")
	if !info.Mixed || info.Checked || !info.Invalid || info.Description != "Select every row" {
		t.Fatalf("mixed semantics=%+v", info)
	}
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeySpace)
	if state != ui.CheckOn || calls != 1 || !slices.Equal(reported, []ui.CheckState{ui.CheckOn}) {
		t.Fatalf("mixed click did not choose On once: state=%v calls=%v", state, reported)
	}
	if after := choiceAccess(t, r, ui.RoleCheckBox, "Select all"); !after.Checked || after.Mixed {
		t.Fatalf("On native state=%+v", after)
	}
	r.Key(0, ui.KeySpace)
	if state != ui.CheckOff || calls != 2 {
		t.Fatal("On activation did not choose Off")
	}
	disabled = true
	r.Frame()
	choiceActivate(t, r, "Select all")
	if state != ui.CheckOff || calls != 2 {
		t.Fatal("disabled checkbox still changed")
	}
	if !choiceAccess(t, r, ui.RoleCheckBox, "Select all").Disabled {
		t.Fatal("disabled mixed checkbox lost native state")
	}
	checked, mixed, boolCalls := false, true, 0
	r = tester(func(c *ui.Context) {
		Checkbox(c, &checked, "Legacy mixed", CheckboxProps{Indeterminate: mixed, OnCheckedChange: func(on bool) { boolCalls++; mixed = false }})
	})
	choiceActivate(t, r, "Legacy mixed")
	if !checked || mixed || boolCalls != 1 || choiceAccess(t, r, ui.RoleCheckBox, "Legacy mixed").Mixed {
		t.Fatal("legacy controlled mixed flag could not transition through its callback")
	}
}

func TestChoiceSwitchSizesDirectionColorsAndNoIdle(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			for _, size := range []Size{DefaultSize, Small} {
				t.Run(fmt.Sprintf("dark=%v/rtl=%v/size=%d", dark, rtl, size), func(t *testing.T) {
					on, calls := false, 0
					var tokens *theme.Tokens
					appearance := theme.LightMode
					if dark {
						appearance = theme.DarkMode
					}
					controller := theme.Controller{Mode: appearance}
					r := ui.NewTesterWith(func(c *ui.Context) {
						controller.Apply(c)
						tokens = theme.TokensOf(c)
						root := ui.Column(c).Fill().Padding(20).Background(tokens.Background)
						if rtl {
							root.Direction(ui.RightToLeft)
						}
						root.Children(func() {
							Switch(c, &on, SwitchProps{Size: size, OnCheckedChange: func(bool) { calls++ }}).Label("Power")
						})
					}, 240, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
					info := choiceAccess(t, r, ui.RoleSwitch, "Power")
					width, height, inset := float32(32), float32(18.4), float32(9)
					if size == Small {
						width, height, inset = 24, 14, 7
					}
					if info.Bounds.W != width || math.Abs(float64(info.Bounds.H-height)) > .01 {
						t.Fatalf("wrong switch size: %+v", info.Bounds)
					}
					knobX := inset
					if rtl {
						knobX = width - inset
					}
					assertColor := func(x float32, expected ui.Color) {
						t.Helper()
						p := r.Image().RGBAAt(int(info.Bounds.X+x), int(info.Bounds.Y+height/2))
						if p.R != expected.R || p.G != expected.G || p.B != expected.B {
							t.Fatalf("knob color at%g=%v want%v", x, p, expected)
						}
					}
					knob := tokens.Background
					if dark {
						knob = tokens.Foreground
					}
					assertColor(knobX, knob)
					choiceActivate(t, r, "Power")
					if !on || calls != 1 || !choiceAccess(t, r, ui.RoleSwitch, "Power").Checked {
						t.Fatal("switch did not update native checked state once")
					}
					if dark {
						knob = tokens.PrimaryForeground
					}
					assertColor(width-knobX, knob)
					assertColor(knobX, tokens.Primary)
					r.Frame()
					if r.FramePending() {
						t.Fatal("reduced-motion switch kept invalidating")
					}
				})
			}
		}
	}
}

func choiceWhiteKnobCenter(r *ui.Tester, bounds ui.Rect) float32 {
	total, count := float32(0), float32(0)
	for x := 1; x < int(bounds.W)-1; x++ {
		pixel := r.Image().RGBAAt(int(bounds.X)+x, int(bounds.Y+bounds.H/2))
		if pixel.R >= 248 && pixel.G >= 248 && pixel.B >= 248 {
			total, count = total+float32(x), count+1
		}
	}
	return total / max(1, count)
}

func TestChoiceSwitchReversalAndLiveDirectionStayInsideTrack(t *testing.T) {
	on := false
	r := tester(func(c *ui.Context) { Switch(c, &on, SwitchProps{}).Label("Power") })
	bounds := choiceAccess(t, r, ui.RoleSwitch, "Power").Bounds
	start := choiceWhiteKnobCenter(r, bounds)
	choiceActivate(t, r, "Power")
	r.Advance(35 * time.Millisecond)
	mid := choiceWhiteKnobCenter(r, bounds)
	if mid <= start+1 || mid >= 23.5 {
		t.Fatalf("switch did not animate to the right: start=%g middle=%g", start, mid)
	}
	choiceActivate(t, r, "Power")
	r.Advance(35 * time.Millisecond)
	reversed := choiceWhiteKnobCenter(r, bounds)
	if reversed >= mid {
		t.Fatalf("switch did not reverse from current position: mid=%g reverse=%g", mid, reversed)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reversed switch never settled")
	}
	on = true
	r.Frame()
	r.Advance(time.Second)
	r.SetDirection(ui.RightToLeft)
	bounds = choiceAccess(t, r, ui.RoleSwitch, "Power").Bounds
	if center := choiceWhiteKnobCenter(r, bounds); center < 6 || center > 11 {
		t.Fatalf("live direction mirrored layout but not thumb offset: %g", center)
	}
}

func TestChoiceRadioGroupOrientationNoLoopRichLabelsAndChanges(t *testing.T) {
	for _, orientation := range []ui.Orientation{ui.Horizontal, ui.Vertical} {
		t.Run(fmt.Sprint(orientation), func(t *testing.T) {
			selected, changes := "one", 0
			var callback []string
			r := tester(func(c *ui.Context) {
				group := RadioGroup(c, &selected, RadioGroupProps{Label: "Storage", Orientation: orientation, NoLoop: true, Options: []Option{
					{Value: "one", Label: "One", Description: "First plan"}, {Value: "disabled", Label: "Disabled", Disabled: true}, {Value: "two", Label: "Two", Description: "Second plan"},
				}, OnValueChange: func(value string) { callback = append(callback, value) }})
				if group.Changed() {
					changes++
				}
				ui.Button(c, "After radios")
			})
			g := choiceAccess(t, r, ui.RoleRadioGroup, "Storage")
			if g.Vertical != (orientation == ui.Vertical) || g.Horizontal != (orientation == ui.Horizontal) {
				t.Fatalf("group orientation=%+v", g)
			}
			r.Key(0, ui.KeyTab)
			if !r.Focused("One") {
				t.Fatal("radio group did not enter at its selected option")
			}
			next, previous, orthogonal := ui.KeyRight, ui.KeyLeft, ui.KeyDown
			if orientation == ui.Vertical {
				next, previous, orthogonal = ui.KeyDown, ui.KeyUp, ui.KeyRight
			}
			r.Key(0, orthogonal)
			if selected != "one" {
				t.Fatal("orthogonal key changed the explicit axis")
			}
			r.Key(0, previous)
			if selected != "one" || !r.Focused("One") {
				t.Fatal("NoLoop wrapped before the first enabled option")
			}
			r.Key(0, next)
			if selected != "two" || !r.Focused("Two") || changes != 1 || !slices.Equal(callback, []string{"two"}) {
				t.Fatalf("native disabled skip/change=%q changes=%d calls=%v", selected, changes, callback)
			}
			r.Key(0, next)
			if selected != "two" || changes != 1 {
				t.Fatal("NoLoop wrapped after the last option")
			}
			r.Key(0, ui.KeyTab)
			if !r.Focused("After radios") {
				t.Fatal("radio group added extra Tab stops")
			}
			if choiceAccess(t, r, ui.RoleRadio, "Two").Description != "Second plan" {
				t.Fatal("rich option lost native description")
			}
		})
	}
}

func TestChoiceToggleGroupVerticalRTLLoopAndImmediateMultiple(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			t.Run(fmt.Sprintf("vertical=%v/rtl=%v", vertical, rtl), func(t *testing.T) {
				selected := []string{}
				var calls [][]string
				options := []Option{{Value: "one", Label: "One"}, {Value: "off", Label: "Off", Disabled: true}, {Value: "two", Label: "Two"}}
				r := tester(func(c *ui.Context) {
					wrapper := ui.Column(c)
					if rtl {
						wrapper.Direction(ui.RightToLeft)
					}
					wrapper.Children(func() {
						ToggleGroup(c, &selected, ToggleGroupProps{Items: options, Label: "Formatting", Multiple: true, Vertical: vertical, NoLoop: true,
							OnValueChange: func(value []string) { calls = append(calls, value) }})
					})
					ui.Button(c, "After toggles")
				})
				r.Key(0, ui.KeyTab)
				next, previous := ui.KeyRight, ui.KeyLeft
				if rtl {
					next, previous = previous, next
				}
				if vertical {
					next, previous = ui.KeyDown, ui.KeyUp
				}
				r.Key(0, previous)
				if !r.Focused("One") {
					t.Fatal("NoLoop wrapped to the last toggle")
				}
				r.Key(0, ui.KeySpace)
				r.Key(0, next)
				if !r.Focused("Two") || !slices.Equal(selected, []string{"one"}) {
					t.Fatal("focus arrows selected a toggle or did not skip disabled")
				}
				r.Key(0, ui.KeySpace)
				if !slices.Equal(selected, []string{"one", "two"}) || len(calls) != 2 || !slices.Equal(calls[1], selected) {
					t.Fatalf("multiple updates=%v callback=%v", selected, calls)
				}
				r.Key(0, next)
				if !r.Focused("Two") {
					t.Fatal("NoLoop wrapped to the first toggle")
				}
				r.Key(0, ui.KeyTab)
				if !r.Focused("After toggles") {
					t.Fatal("toggle group did not remain one Tab stop")
				}
			})
		}
	}
}

func TestChoiceFieldInvalidVisualScopeAndNativeState(t *testing.T) {
	on, other := false, false
	r := tester(func(c *ui.Context) {
		Field(c, FieldProps{Label: "Required acceptance", Error: "Accept to continue"}, func() ui.Element { return Checkbox(c, &on, "", CheckboxProps{}) })
		Checkbox(c, &other, "Outside field", CheckboxProps{})
	})
	field := choiceAccess(t, r, ui.RoleCheckBox, "Required acceptance")
	outside := choiceAccess(t, r, ui.RoleCheckBox, "Outside field")
	if !field.Invalid || outside.Invalid {
		t.Fatal("choice invalid state did not honor the enclosing field scope")
	}
	pixel := r.Image().RGBAAt(int(field.Bounds.X+8), int(field.Bounds.Y-2))
	if pixel.R <= pixel.G+15 {
		t.Fatalf("field metadata changed after construction but its choice ring did not: %v", pixel)
	}
}

func TestChoiceEmptyCallbacksAndStaticVisualsPreserveSiblingIdentity(t *testing.T) {
	for _, which := range []string{"checkbox", "radio", "switch", "toggle", "radio-group", "toggle-group"} {
		t.Run(which, func(t *testing.T) {
			show, on, selected, text := false, false, "one", "Draft"
			values := []string{}
			r := tester(func(c *ui.Context) {
				content := func(child *ui.Context) {
					if show {
						ui.Button(child, "Visual only button")
					}
				}
				switch which {
				case "checkbox":
					Checkbox(c, &on, "Choice", CheckboxProps{Content: content})
				case "radio":
					RadioWith(c, &selected, "one", "Choice", RadioProps{Content: content})
				case "switch":
					Switch(c, &on, SwitchProps{Label: "Choice", Content: content})
				case "toggle":
					Toggle(c, &on, "Choice", ToggleProps{Content: content})
				case "radio-group":
					RadioGroup(c, &selected, RadioGroupProps{Options: []Option{{Value: "one", Label: "Choice"}}, ItemContent: func(child *ui.Context, _ Option, _ bool) { content(child) }})
				case "toggle-group":
					ToggleGroup(c, &values, ToggleGroupProps{Items: []Option{{Value: "one", Label: "Choice"}}, ItemContent: func(child *ui.Context, _ Option, _ bool) { content(child) }})
				}
				ui.TextInput(c, &text).Label("Following editor")
				ui.Button(c, "Following action")
			})
			choiceActivate(t, r, "Following editor")
			show = true
			r.Frame()
			if !r.Focused("Following editor") {
				t.Fatal("empty callback leaked its key into the following editor")
			}
			if _, visible := r.AccessByRole(ui.RoleButton, "Visual only button"); visible {
				t.Fatal("visual label created a nested accessible button")
			}
			r.Key(0, ui.KeyTab)
			if !r.Focused("Following action") {
				t.Fatal("visual label created a hidden keyboard stop")
			}
		})
	}
}

func TestChoiceJoinedToggleGeometryAndSelectedPixels(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(fmt.Sprint(vertical), func(t *testing.T) {
			selected := []string{"one"}
			zero := float32(0)
			var muted ui.Color
			r := tester(func(c *ui.Context) {
				muted = theme.TokensOf(c).Muted
				ToggleGroup(c, &selected, ToggleGroupProps{Label: "Joined", Vertical: vertical, Outline: true, Spacing: &zero, Items: []Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}})
			})
			a, b := choiceAccess(t, r, ui.RoleToggleButton, "One"), choiceAccess(t, r, ui.RoleToggleButton, "Two")
			if vertical && math.Abs(float64(a.Bounds.Y+a.Bounds.H-b.Bounds.Y)) > .01 || !vertical && math.Abs(float64(a.Bounds.X+a.Bounds.W-b.Bounds.X)) > .01 {
				t.Fatalf("joined gap remained: a=%+v b=%+v", a.Bounds, b.Bounds)
			}
			p := r.Image().RGBAAt(int(a.Bounds.X+4), int(a.Bounds.Y+a.Bounds.H/2))
			if p != (color.RGBA{R: muted.R, G: muted.G, B: muted.B, A: muted.A}) {
				t.Fatalf("pressed toggle does not use muted: %v want %v", p, muted)
			}
		})
	}
}

func TestChoiceDefaultToggleIconSizeAndActualHoverFill(t *testing.T) {
	on := false
	var muted ui.Color
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		muted = theme.TokensOf(c).Muted
		ui.Column(c).Fill().Padding(20).Background(ui.RGB(180, 200, 220)).Children(func() {
			Toggle(c, &on, "", ToggleProps{Label: "Format", Icon: check})
		})
	}, 160, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	info := choiceAccess(t, r, ui.RoleToggleButton, "Format")
	if info.Bounds.W != 36 || info.Bounds.H != 36 {
		t.Fatalf("default icon toggle has an unsolicited border affecting source36DIP geometry: %+v", info.Bounds)
	}
	r.Move(info.Bounds.X+4, info.Bounds.Y+18)
	color := r.Image().RGBAAt(int(info.Bounds.X+4), int(info.Bounds.Y+18))
	if color.R != muted.R || color.G != muted.G || color.B != muted.B {
		t.Fatalf("default toggle hover did not actually paint muted: %v want%v", color, muted)
	}
}

func TestChoiceStableOptionIdentityKeepsNativeFocusAcrossReorder(t *testing.T) {
	for _, toggles := range []bool{false, true} {
		t.Run(fmt.Sprint(toggles), func(t *testing.T) {
			value, selected := "two", []string{"two"}
			items := []Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}, {Value: "three", Label: "Three"}}
			r := tester(func(c *ui.Context) {
				if toggles {
					ToggleGroup(c, &selected, ToggleGroupProps{Items: items})
				} else {
					RadioGroup(c, &value, RadioGroupProps{Options: items, Orientation: ui.Horizontal})
				}
			})
			choiceActivate(t, r, "Two")
			items = []Option{items[2], items[1], items[0]}
			r.Frame()
			if !r.Focused("Two") {
				t.Fatal("reordering option values lost native focus")
			}
			r.Key(0, ui.KeyRight)
			if !r.Focused("One") || !toggles && value != "one" {
				t.Fatalf("reordered native focus group used stale index: radio=%q", value)
			}
		})
	}
}

func TestChoiceDisabledAncestorSuppressesNativeEventsAndCallbacks(t *testing.T) {
	disabled, on, checked, radio := true, false, false, "one"
	selected, calls := []string{}, 0
	r := tester(func(c *ui.Context) {
		ui.Column(c).Disabled(disabled).Children(func() {
			Checkbox(c, &checked, "Acceptance", CheckboxProps{OnCheckedChange: func(bool) { calls++ }})
			Switch(c, &on, SwitchProps{Label: "Power", OnCheckedChange: func(bool) { calls++ }})
			RadioGroup(c, &radio, RadioGroupProps{Options: []Option{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}}, OnValueChange: func(string) { calls++ }})
			ToggleGroup(c, &selected, ToggleGroupProps{Items: []Option{{Value: "bold", Label: "Bold"}}, OnValueChange: func([]string) { calls++ }})
		})
	})
	for _, name := range []string{"Acceptance", "Power", "Two", "Bold"} {
		choiceActivate(t, r, name)
	}
	if checked || on || radio != "one" || len(selected) != 0 || calls != 0 {
		t.Fatal("choice child ignored its disabled ancestor")
	}
	disabled = false
	r.Frame()
	for _, name := range []string{"Acceptance", "Power", "Two", "Bold"} {
		choiceActivate(t, r, name)
	}
	if !checked || !on || radio != "two" || !slices.Equal(selected, []string{"bold"}) || calls != 4 {
		t.Fatalf("re-enabled controls did not deliver one event each: checked=%v on=%v radio=%q values=%v calls=%d", checked, on, radio, selected, calls)
	}
}

func TestChoiceSwitchCancelAndDynamicReducedMotionSettle(t *testing.T) {
	on, calls := false, 0
	r := tester(func(c *ui.Context) {
		Switch(c, &on, SwitchProps{OnCheckedChange: func(bool) { calls++ }}).Label("Power")
	})
	bounds := choiceAccess(t, r, ui.RoleSwitch, "Power").Bounds
	r.Press(bounds.X+16, bounds.Y+9)
	r.CancelPress()
	if on || calls != 0 {
		t.Fatal("cancelled native pointer press toggled the switch")
	}
	choiceActivate(t, r, "Power")
	r.Advance(35 * time.Millisecond)
	if !r.FramePending() {
		t.Fatal("normal switch feedback did not animate")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	r.Frame()
	if !on || calls != 1 || r.FramePending() {
		t.Fatalf("enabling reduced motion did not settle an active switch: on=%v calls=%d pending=%v center=%g", on, calls, r.FramePending(), choiceWhiteKnobCenter(r, bounds))
	}
	if center := choiceWhiteKnobCenter(r, bounds); center < 21 || center > 25 {
		t.Fatalf("reduced-motion switch stopped before its target: %g", center)
	}
}

func TestChoiceRadioUsesFilledSelectedIndicator(t *testing.T) {
	selected := "go"
	var foreground ui.Color
	r := tester(func(c *ui.Context) {
		foreground = theme.TokensOf(c).PrimaryForeground
		Radio(c, &selected, "go", "Go", false)
	})
	radio, ok := r.AccessByRole(ui.RoleRadio, "Go")
	if !ok {
		t.Fatal("missing native radio")
	}
	pixel := r.Image().RGBAAt(int(radio.Bounds.X+8), int(radio.Bounds.Y+radio.Bounds.H/2))
	if pixel.R != foreground.R || pixel.G != foreground.G || pixel.B != foreground.B {
		t.Fatalf("selected radio center does not use primaryforeground: %v want %v", pixel, foreground)
	}
}

func TestChoiceToggleGroupReportsSameEventChange(t *testing.T) {
	selected := []string{}
	var events [][]string
	r := tester(func(c *ui.Context) {
		group := ToggleGroup(c, &selected, ToggleGroupProps{Items: []Option{{Value: "bold", Label: "Bold"}, {Value: "italic", Label: "Italic"}}})
		if group.Changed() {
			events = append(events, slices.Clone(selected))
		}
	})
	if err := r.Click("Bold"); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || !slices.Equal(events[0], []string{"bold"}) {
		t.Fatalf("group did not publish current caller state in event frame: events=%v selected=%v", events, selected)
	}
	r.Frame()
	if len(events) != 1 {
		t.Fatal("repaint repeated a toggle change")
	}
}
