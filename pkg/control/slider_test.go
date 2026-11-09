package control

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func sliderAccess(t *testing.T, tt *ui.Tester, label string) ui.AccessInfo {
	t.Helper()
	info, ok := tt.AccessByRole(ui.RoleSlider, label)
	if !ok {
		t.Fatalf("missing native slider thumb %q", label)
	}
	return info
}

func TestRangeSliderPointerCaptureGapAndCancel(t *testing.T) {
	low, high, changes := 20.0, 80.0, 0
	tt := tester(func(c *ui.Context) {
		if RangeSlider(c, &low, &high, SliderProps{Label: "Price", Step: 10, MinGap: 20}).Width(240).Changed() {
			changes++
		}
	})
	lo := sliderAccess(t, tt, "Price minimum")
	hi := sliderAccess(t, tt, "Price maximum")
	if lo.Min != 0 || lo.Max != 60 || hi.Min != 40 || hi.Max != 100 || lo.Step != 10 {
		t.Fatalf("constrained accessibility ranges: low=%+v high=%+v", lo, hi)
	}
	// Start on the track near the upper endpoint; capture keeps this thumb
	// active even after the pointer leaves both the track and the window.
	group, _ := tt.AccessByRole(ui.RoleGroup, "Price")
	tt.Press(group.Bounds.X+200, group.Bounds.Y+10)
	tt.Move(group.Bounds.X+400, group.Bounds.Y+80)
	if low != 20 || high != 100 {
		t.Fatalf("nearest-thumb captured drag = [%v %v]", low, high)
	}
	tt.Move(group.Bounds.X-100, group.Bounds.Y+80)
	if low != 20 || high != 40 {
		t.Fatalf("drag crossed requested gap: [%v %v]", low, high)
	}
	tt.CancelPress()
	if low != 20 || high != 80 {
		t.Fatalf("cancel did not restore controlled endpoints: [%v %v]", low, high)
	}
	prior := changes
	tt.Frame()
	tt.Frame()
	if changes != prior || changes == 0 {
		t.Fatalf("Changed repeated or missing: %d -> %d", prior, changes)
	}
	// Clicking the lower thumb after the upper thumb had focus must move
	// keyboard focus to the lower slot, rather than keeping the old one.
	lo = sliderAccess(t, tt, "Price minimum")
	tt.ClickAt(lo.Bounds.X+8, lo.Bounds.Y+8)
	tt.Key(0, ui.KeyEnd)
	if low != 60 || high != 80 {
		t.Fatalf("pointer focus followed by End = [%v %v]", low, high)
	}
}

func TestSliderValuesKeyboardFocusFormattingAndIdentity(t *testing.T) {
	values := []float64{10, 50, 90}
	p := SliderProps{Label: "Bands", Step: 5, MinStepsBetweenValues: 2,
		ThumbIDs: []string{"bass", "mid", "treble"}, ThumbLabels: []string{"Bass", "Mid", "Treble"},
		FormatValue: func(i int, v float64) string { return fmt.Sprintf("Band %d: %.0f dB", i+1, v) }}
	tt := tester(func(c *ui.Context) { SliderValues(c, &values, p).Width(240) })
	mid := sliderAccess(t, tt, "Mid")
	if mid.Value != "Band 2: 50 dB" || mid.Now != 50 || mid.Min != 20 || mid.Max != 80 || mid.Step != 5 {
		t.Fatalf("middle thumb accessibility = %+v", mid)
	}
	tt.Key(0, ui.KeyTab)
	if !sliderAccess(t, tt, "Bass").Focused {
		t.Fatal("first ordered thumb is not a native tab stop")
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	if !slices.Equal(values, []float64{10, 80, 90}) || !sliderAccess(t, tt, "Mid").Focused {
		t.Fatalf("middle End or native focus = %v", values)
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyRight)
	if values[1] != 25 {
		t.Fatalf("middle Home then Right = %v", values)
	}
	tt.Key(0, ui.KeyPageUp)
	if values[1] != 35 {
		t.Fatalf("PageUp = %v", values)
	}
	// Replacing the controlled slice and labels keeps the same keyed thumb's
	// native focus. Keys then operate on the new caller-owned storage.
	values = []float64{5, 45, 95}
	p.ThumbLabels[1] = "Middle"
	tt.Frame()
	if !sliderAccess(t, tt, "Middle").Focused {
		t.Fatal("controlled replacement lost stable thumb focus")
	}
	tt.Key(0, ui.KeyRight)
	if values[1] != 50 {
		t.Fatalf("new slice binding was not used: %v", values)
	}
	tt.Key(ui.Shift, ui.KeyTab)
	if !sliderAccess(t, tt, "Bass").Focused {
		t.Fatal("Shift+Tab did not return to previous thumb")
	}
}

func TestSliderOrientationAndRTL(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(fmt.Sprintf("vertical=%v", vertical), func(t *testing.T) {
			value := 20.0
			tt := tester(func(c *ui.Context) {
				ui.Box(c).Direction(ui.RightToLeft).Children(func() {
					Slider(c, &value, SliderProps{Label: "Level", Step: 10, Vertical: vertical}).Width(160).Height(160)
				})
			})
			info := sliderAccess(t, tt, "Level")
			if info.Vertical != vertical || info.Horizontal == vertical {
				t.Fatalf("orientation metadata = %+v", info)
			}
			tt.ClickAt(info.Bounds.X+8, info.Bounds.Y+8)
			tt.Key(0, ui.KeyRight)
			want := 10.0
			if vertical {
				want = 30
			}
			if value != want {
				t.Fatalf("RTL Right = %v, want %v", value, want)
			}
			tt.Key(0, ui.KeyUp)
			if value != want+10 {
				t.Fatalf("Up did not increase: %v", value)
			}
			group, _ := tt.AccessByRole(ui.RoleGroup, "Level")
			if vertical {
				tt.ClickAt(group.Bounds.X+group.Bounds.W/2, group.Bounds.Y)
			} else {
				tt.ClickAt(group.Bounds.X, group.Bounds.Y+group.Bounds.H/2)
			}
			if value != 100 {
				t.Fatalf("high endpoint pointer = %v", value)
			}
		})
	}
}

func TestSliderConstraintsNormalizationAndDisabled(t *testing.T) {
	values, changes := []float64{95, -10, math.NaN()}, 0
	disabled := false
	p := SliderProps{Label: "Values", Min: 100, Max: 0, Step: 10, MinGap: 15}
	tt := tester(func(c *ui.Context) {
		p.Disabled = disabled
		if SliderValues(c, &values, p).Width(240).Changed() {
			changes++
		}
	})
	if !slices.Equal(values, []float64{0, 20, 100}) || changes != 0 {
		t.Fatalf("external normalization = %v, changes=%d", values, changes)
	}
	first := sliderAccess(t, tt, "Values Thumb 1")
	tt.ClickAt(first.Bounds.X+8, first.Bounds.Y+8)
	tt.Key(0, ui.KeyEnd)
	if values[0] != 0 || changes != 0 {
		t.Fatalf("blocked thumb generated a change: %v, %d", values, changes)
	}
	disabled = true
	tt.Frame()
	last := sliderAccess(t, tt, "Values Thumb 3")
	if !last.Disabled {
		t.Fatal("disabled state missing from native thumb")
	}
	tt.ClickAt(last.Bounds.X+8, last.Bounds.Y+8)
	tt.Key(0, ui.KeyHome)
	if !slices.Equal(values, []float64{0, 20, 100}) {
		t.Fatalf("disabled slider changed %v", values)
	}
	disabled, p.MinGap = false, 60
	tt.Frame()
	if group, _ := tt.AccessByRole(ui.RoleGroup, "Values"); !group.Invalid || !group.Disabled {
		t.Fatalf("impossible minimum gap not reported: %+v", group)
	}
}

func TestSliderFiniteMetadataEmptyAndConstant(t *testing.T) {
	for _, p := range []SliderProps{
		{Min: math.NaN(), Max: 5}, {Min: 0, Max: math.Inf(1)},
		{Min: 3, Max: 3}, {Min: -math.MaxFloat64, Max: math.MaxFloat64},
		{Step: math.Inf(1)},
	} {
		value := math.Inf(1)
		p.Label = "Finite"
		tt := tester(func(c *ui.Context) { Slider(c, &value, p).Width(240) })
		info := sliderAccess(t, tt, "Finite")
		for _, v := range []float64{info.Min, info.Max, info.Now, info.Step} {
			if !finite(v) {
				t.Fatalf("nonfinite accessibility value %v from %+v", v, p)
			}
		}
	}
	values := []float64{}
	tt := tester(func(c *ui.Context) {
		SliderValues(c, &values, SliderProps{Label: "Empty"}).Width(240)
		Slider(c, nil, SliderProps{Label: "Nil"}).Width(240)
	})
	for _, label := range []string{"Empty", "Nil"} {
		if info, ok := tt.AccessByRole(ui.RoleGroup, label); !ok || !info.Disabled {
			t.Fatalf("empty slider %q not inert: %+v", label, info)
		}
	}
}

func TestSliderDecimalStepsAndHugeRange(t *testing.T) {
	values := []float64{0, .3, .6}
	tt := tester(func(c *ui.Context) {
		SliderValues(c, &values, SliderProps{Label: "Decimal", Max: 1, Step: .1, MinStepsBetweenValues: 3}).Width(240)
	})
	info := sliderAccess(t, tt, "Decimal Thumb 2")
	if math.Abs(values[1]-.3) > 1e-15 || math.Abs(info.Min-.3) > 1e-15 {
		t.Fatalf("three decimal steps became a larger gap: %v, %+v", values, info)
	}
	values = []float64{-math.MaxFloat64, 0, math.MaxFloat64}
	tt = tester(func(c *ui.Context) {
		SliderValues(c, &values, SliderProps{Label: "Huge", Min: -math.MaxFloat64, Max: math.MaxFloat64, MinGap: math.MaxFloat64}).Width(240)
	})
	if !slices.Equal(values, []float64{-math.MaxFloat64, 0, math.MaxFloat64}) {
		t.Fatalf("finite extreme endpoints overflowed normalization: %v", values)
	}
	for _, label := range []string{"Huge Thumb 1", "Huge Thumb 2", "Huge Thumb 3"} {
		info := sliderAccess(t, tt, label)
		if !finite(info.Min) || !finite(info.Max) || !finite(info.Now) || info.Invalid {
			t.Fatalf("extreme range accessibility = %+v", info)
		}
	}
}
