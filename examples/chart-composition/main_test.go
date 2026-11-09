package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func exampleTester(s *appState, width int) *ui.Tester {
	s.appearance.ReduceMotion = true
	return ui.NewTesterWith(s.view, width, 1000, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}

func exampleAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, label)
	if !ok || a.Offscreen || a.Bounds.W <= 0 || a.Bounds.H <= 0 {
		t.Fatalf("missing visible native %v %q: %+v", role, label, a)
	}
	return a
}

func exampleClick(t *testing.T, r *ui.Tester, role ui.Role, label string) {
	t.Helper()
	a := exampleAccess(t, r, role, label)
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
}

func examplePixel(r *ui.Tester, x, y float32) color.RGBA {
	return r.Image().RGBAAt(int(x), int(y))
}

func exampleColorCount(img *image.RGBA, rect ui.Rect, want color.RGBA) int {
	count := 0
	for y := int(rect.Y); y < int(rect.Y+rect.H); y++ {
		for x := int(rect.X); x < int(rect.X+rect.W); x++ {
			if img.RGBAAt(x, y) == want {
				count++
			}
		}
	}
	return count
}

func TestCompositionRawBarsLabelsThresholdAndNativeInspection(t *testing.T) {
	s := newState(false, false)
	r := exampleTester(s, 1000)
	plot := exampleAccess(t, r, ui.RoleGroup, "Monthly revenue plot").Bounds
	if _, ok := r.AccessByRole(ui.RoleGroup, "Inspected month"); ok {
		t.Fatal("the default example invented an inspected point")
	}
	for _, pair := range []struct {
		month string
		value float32
	}{{"January", 186}, {"February", 305}, {"March", 237}, {"April", 173}, {"May", 209}, {"June", 214}} {
		bar := exampleAccess(t, r, ui.RoleImage, fmt.Sprintf("%s revenue %.0f", pair.month, pair.value)).Bounds
		if math.Abs(float64(bar.H-plot.H*pair.value/400)) > 1 || bar.W < 30 {
			t.Fatalf("native raw mark does not scale its data: %s bar=%v plot=%v", pair.month, bar, plot)
		}
		if got := examplePixel(r, bar.X+bar.W/2, bar.Y+bar.H*.7); got != (color.RGBA{37, 99, 235, 255}) {
			t.Fatalf("live config color did not reach raw %s mark: %v", pair.month, got)
		}
		// A real Anchor paints each value above its bar, outside the bar box.
		label := ui.Rect{X: bar.X + bar.W/2 - 18, Y: bar.Y - 20, W: 36, H: 16}
		if exampleColorCount(r.Image(), label, color.RGBA{10, 10, 10, 255}) < 3 {
			t.Fatalf("%s label did not paint above its raw mark", pair.month)
		}
	}
	if got := examplePixel(r, plot.X+2, plot.Y+plot.H*.375); got != (color.RGBA{234, 88, 12, 255}) {
		t.Fatalf("native RuleY threshold did not paint at $250k: %v", got)
	}
	if !r.HasText("Monthly target") || !r.HasText("Revenue") {
		t.Fatal("the config labels did not reach the composed legend")
	}
	march := exampleAccess(t, r, ui.RoleImage, "March revenue 237").Bounds
	r.Move(march.X+march.W/2, plot.Y+plot.H*.8)
	if s.selected != 2 || !r.HasText("March") || !r.HasText("237") {
		t.Fatalf("native pointer did not inspect March: index=%d", s.selected)
	}
	if got := exampleAccess(t, r, ui.RoleGroup, "Monthly revenue plot").Value; got != "March: $237k revenue" {
		t.Fatalf("native Select value = %q", got)
	}
	// The application retains its inspection when the pointer leaves so that
	// the value can be edited. Its native accessible value must agree even on
	// that exact leave frame, without relying on a later incidental rebuild.
	r.Move(4, 4)
	if s.selected != 2 || !r.HasText("March") || !r.HasText("237") {
		t.Fatal("leaving the plot discarded the app-owned inspection")
	}
	if got := exampleAccess(t, r, ui.RoleGroup, "Monthly revenue plot").Value; got != "March: $237k revenue" {
		t.Fatalf("retained tooltip and native plot value disagree after pointer leave: %q", got)
	}
	r.ClickAt(march.X+march.W/2, plot.Y+plot.H*.8)
	if !r.Focused("Monthly revenue plot") {
		t.Fatal("the native plot did not take keyboard focus")
	}
	r.Key(0, ui.KeyEnd)
	if s.selected != 5 || !r.HasText("June") {
		t.Fatal("native End did not inspect the final visible month")
	}
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyRight)
	if s.selected != 1 || !r.HasText("February") {
		t.Fatal("native Home/Right did not inspect February")
	}
	r.Key(0, ui.KeyEscape)
	if s.selected != -1 || !exampleAccess(t, r, ui.RoleTextField, "Inspected revenue").Disabled {
		t.Fatal("Escape did not clear inspection and disable its editor")
	}
	if got := exampleAccess(t, r, ui.RoleGroup, "Monthly revenue plot").Value; got != "" {
		t.Fatalf("clearing inspection left a stale native plot value: %q", got)
	}
}

func TestCompositionRangeUsesRealThumbsAndRemovesExcludedRawMarks(t *testing.T) {
	s := newState(false, true)
	r := exampleTester(s, 760)
	exampleClick(t, r, ui.RoleSlider, "First month")
	r.Key(0, ui.KeyRight)
	if s.low != 1 || !r.HasText("February – June") || s.selected != 2 {
		t.Fatalf("native minimum range move lost its visible inspection: range=[%v,%v] inspected=%d", s.low, s.high, s.selected)
	}
	if _, ok := r.AccessByRole(ui.RoleImage, "January revenue 186"); ok {
		t.Fatal("excluded January raw mark remained in native tree")
	}
	plot := exampleAccess(t, r, ui.RoleGroup, "Monthly revenue plot").Bounds
	feb := exampleAccess(t, r, ui.RoleImage, "February revenue 305").Bounds
	if feb.X > plot.X+plot.W*.1 || feb.W < plot.W*.11 {
		t.Fatalf("range filtering did not relayout the native Band scale: bar=%v plot=%v", feb, plot)
	}
	exampleClick(t, r, ui.RoleSlider, "Last month")
	r.Key(0, ui.KeyHome)
	if s.high != 2 || !r.HasText("February – March") {
		t.Fatalf("constrained native high endpoint = %v", s.high)
	}
	exampleClick(t, r, ui.RoleSlider, "First month")
	r.Key(0, ui.KeyHome)
	exampleClick(t, r, ui.RoleSlider, "Last month")
	r.Key(0, ui.KeyLeft)
	if s.high != 1 || s.selected != -1 {
		t.Fatalf("excluding March did not clear its app-owned inspection: high=%v selected=%d", s.high, s.selected)
	}
	if _, ok := r.AccessByRole(ui.RoleGroup, "Inspected month"); ok {
		t.Fatal("an out-of-range payload remained in the tooltip")
	}
}

func TestCompositionNativeEditorValidatesAndRepaintsRawMark(t *testing.T) {
	s := newState(false, true)
	r := exampleTester(s, 1000)
	exampleClick(t, r, ui.RoleTextField, "Inspected revenue")
	r.Command("selectAll")
	r.Type("NaN")
	r.Key(0, ui.KeyEnter)
	if s.revenue[2] != 237 || !exampleAccess(t, r, ui.RoleTextField, "Inspected revenue").Invalid || !r.HasText("Enter a number from 0 to 400.") {
		t.Fatal("invalid native text input mutated data or failed to expose validation")
	}
	r.Command("selectAll")
	r.Type("380")
	r.Key(0, ui.KeyEnter)
	plot := exampleAccess(t, r, ui.RoleGroup, "Monthly revenue plot").Bounds
	bar := exampleAccess(t, r, ui.RoleImage, "March revenue 380").Bounds
	if s.revenue[2] != 380 || exampleAccess(t, r, ui.RoleTextField, "Inspected revenue").Invalid || math.Abs(float64(bar.H-plot.H*.95)) > 1 {
		t.Fatalf("committed editor value did not reach real mark geometry: value=%v bar=%v plot=%v", s.revenue[2], bar, plot)
	}
	if !r.HasText("March") || !r.HasText("380") || !r.Focused("Inspected revenue") {
		t.Fatal("moving into the native editor discarded the inspected month or focus")
	}
	r.Command("selectAll")
	r.Type("125")
	exampleClick(t, r, ui.RoleButton, "Apply revenue")
	bar = exampleAccess(t, r, ui.RoleImage, "March revenue 125").Bounds
	if math.Abs(float64(bar.H-plot.H*.3125)) > 1 {
		t.Fatal("native Apply button left the old raw mark painted")
	}
}

func TestCompositionIndicatorsAndThemeUseLiveResolvedColors(t *testing.T) {
	s := newState(false, true)
	r := exampleTester(s, 1000)
	count := func(want color.RGBA) int {
		return exampleColorCount(r.Image(), exampleAccess(t, r, ui.RoleGroup, "Inspected month").Bounds, want)
	}
	blue := color.RGBA{37, 99, 235, 255}
	dot := count(blue)
	exampleClick(t, r, ui.RoleRadio, "Line")
	line := count(blue)
	exampleClick(t, r, ui.RoleRadio, "Dashed")
	dashed := count(blue)
	if dot < 20 || line < 10 || dashed < 3 || dot == line || line == dashed {
		t.Fatalf("indicator controls did not paint distinct real marks: dot=%d line=%d dashed=%d", dot, line, dashed)
	}
	exampleClick(t, r, ui.RoleSwitch, "Hide indicator")
	if got := count(blue); got != 0 || s.selected != 2 || !r.HasText("237") {
		t.Fatalf("hidden indicator did not preserve data without its visual mark: count=%d", got)
	}
	exampleClick(t, r, ui.RoleSwitch, "Hide indicator")
	exampleClick(t, r, ui.RoleSwitch, "Dark appearance")
	darkBlue := color.RGBA{147, 197, 253, 255}
	bar := exampleAccess(t, r, ui.RoleImage, "March revenue 237").Bounds
	if !s.dark || examplePixel(r, bar.X+bar.W/2, bar.Y+bar.H*.8) != darkBlue || count(darkBlue) < 3 {
		t.Fatal("dark theme colors did not reach both raw marks and tooltip")
	}
	legend := exampleAccess(t, r, ui.RoleGroup, "Chart legend").Bounds
	if exampleColorCount(r.Image(), legend, darkBlue) < 20 {
		t.Fatal("the legend did not share the newly resolved dark config color")
	}
}

func TestCompositionNativeNoteSurvivesRangeThemeAndIMECandidate(t *testing.T) {
	s := newState(false, true)
	s.note = ""
	r := exampleTester(s, 760)
	exampleClick(t, r, ui.RoleTextField, "Chart note")
	r.Compose("季度", 2)
	// The app can replace filter/theme state without remounting its editor.
	s.low, s.high, s.dark = 1, 4, true
	r.Frame()
	if !r.Focused("Chart note") {
		t.Fatal("native IME focus was lost during a chart rebuild")
	}
	// Enter must commit the existing candidate after both rebuilds; typing the
	// same string here would hide an accidental editor remount.
	r.Key(0, ui.KeyEnter)
	r.Type(" review")
	if s.note != "季度\n review" {
		t.Fatalf("the original native editor did not commit exactly once: %q", s.note)
	}
	r.Command("selectAll")
	r.Command("copy")
	if r.Clipboard() != "季度\n review" || !strings.Contains(exampleAccess(t, r, ui.RoleTextField, "Chart note").Value, "季度") {
		t.Fatal("the composed note did not retain native selection/clipboard behavior")
	}
}

func TestCompositionRTLTimelineKeepsPhysicalSelectionAndNarrowControlsReachable(t *testing.T) {
	s := newState(true, false)
	r := exampleTester(s, 500)
	r.SetDirection(ui.RightToLeft)
	plot := exampleAccess(t, r, ui.RoleGroup, "Monthly revenue plot").Bounds
	jan := exampleAccess(t, r, ui.RoleImage, "January revenue 186").Bounds
	jun := exampleAccess(t, r, ui.RoleImage, "June revenue 214").Bounds
	if jan.X >= jun.X || plot.W < 300 || plot.X < 0 || plot.X+plot.W > 500 {
		t.Fatalf("explicit LTR timeline was mirrored or overflowed an RTL window: plot=%v January=%v June=%v", plot, jan, jun)
	}
	r.ClickAt(jan.X+jan.W/2, plot.Y+plot.H*.8)
	r.Key(0, ui.KeyRight)
	if s.selected != 1 || !r.HasText("February") {
		t.Fatalf("physical native Right no longer matched the plot: selected=%d", s.selected)
	}
	for _, name := range []string{"Inspected revenue", "Chart note"} {
		box := exampleAccess(t, r, ui.RoleTextField, name).Bounds
		if box.X < 0 || box.X+box.W > 500 || box.W < 180 {
			t.Fatalf("narrow editor is clipped: %s %v", name, box)
		}
	}
	if r.FramePending() {
		t.Fatal("the static composed example keeps requesting frames")
	}
}
