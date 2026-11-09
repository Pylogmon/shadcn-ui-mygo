package gallery

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Keep both real cards visible, so an offscreen animation policy cannot make
// a motion test accidentally pass through its immediate-finish path.
func chartMotionTester(s *State, reduce bool) *ui.Tester {
	s.Page = "display"
	return ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(12).Children(func() {
			ui.Button(c, "Motion page anchor")
			if s.Page == "display" {
				s.chartCompositionPreview(c)
			} else {
				ui.Text(c, "Another preview")
			}
		})
	}, 760, 1500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduce}})
}

func chartMotionPixels(t *testing.T, r *ui.Tester, label string) []byte {
	t.Helper()
	info := advancedAccess(t, r, ui.RoleGroup, label)
	if info.Offscreen || info.Bounds.W <= 0 || info.Bounds.H <= 0 || info.Bounds.Y < 0 || info.Bounds.Y+info.Bounds.H > 1500 {
		t.Fatalf("motion evidence plot is not actually visible: %s %+v", label, info)
	}
	b := info.Bounds
	box := image.Rect(int(math.Ceil(float64(b.X))), int(math.Ceil(float64(b.Y))), int(math.Floor(float64(b.X+b.W))), int(math.Floor(float64(b.Y+b.H))))
	copy := image.NewRGBA(image.Rect(0, 0, box.Dx(), box.Dy()))
	draw.Draw(copy, copy.Bounds(), r.Image(), box.Min, draw.Src)
	return copy.Pix
}

func chartMotionRawValue(t *testing.T, r *ui.Tester, label, expected string) {
	t.Helper()
	if value := advancedAccess(t, r, ui.RoleGroup, label).Value; !strings.Contains(value, expected) {
		t.Fatalf("latest raw data did not reach %s accessibility during motion: got %q want %q", label, value, expected)
	}
}

func TestCreatorChartMotionUpdatesPaintIntermediateGeometryAndExposeLatestRawValues(t *testing.T) {
	for _, family := range []string{"bar", "line", "area", "pie", "donut", "radar", "radial"} {
		t.Run(family, func(t *testing.T) {
			s := New()
			s.chartPreview.initialize()
			p := &s.chartPreview
			p.family, p.selected, p.familySelected = family, 2, 2
			if family == "bar" {
				p.family = "line"
			}
			r := chartMotionTester(s, false)
			if !advancedAccess(t, r, ui.RoleSwitch, "Animate chart changes").Checked {
				t.Fatal("the application did not default to animated changes")
			}
			r.Advance(2500 * time.Millisecond)
			plot, action, expected := "Visitor story", "Update story counts", "March, Desktop: 312, Mobile: 240"
			delay := 900 * time.Millisecond
			if family == "bar" {
				plot, action, expected, delay = "Visitor comparison", "Update visitor counts", "March, Desktop: 252", 160*time.Millisecond
			} else if family == "pie" || family == "donut" || family == "radial" {
				expected = "Firefox, Visitors: 125"
			}
			before := chartMotionPixels(t, r, plot)
			advancedClick(t, r, action)
			chartMotionRawValue(t, r, plot, expected)
			if family == "bar" && p.storyUpdated || family != "bar" && p.updated {
				t.Fatal("one card's action updated the other card's data")
			}
			r.Advance(delay)
			middle := chartMotionPixels(t, r, plot)
			chartMotionRawValue(t, r, plot, expected)
			r.Advance(3 * time.Second)
			final := chartMotionPixels(t, r, plot)
			if bytes.Equal(before, final) || bytes.Equal(middle, final) || bytes.Equal(before, middle) {
				t.Fatal("a real data action did not produce distinct starting, intermediate, and final plot pixels")
			}
			if r.FramePending() {
				t.Fatal("the settled chart demo kept requesting idle frames")
			}
			if family != "bar" {
				advancedClick(t, r, "Restore story counts")
				original := "March, Desktop: 237, Mobile: 120"
				if family == "pie" || family == "donut" || family == "radial" {
					original = "Firefox, Visitors: 187"
				}
				chartMotionRawValue(t, r, plot, original)
				r.Advance(3 * time.Second)
				if !bytes.Equal(before, chartMotionPixels(t, r, plot)) || p.storyUpdated {
					t.Fatal("Restore did not return the actual plotted story to its original counts")
				}
			}
		})
	}
}

func TestCreatorChartMotionSwitchAndReducedPreferenceFinishActualPlotsImmediately(t *testing.T) {
	for _, reduced := range []bool{false, true} {
		t.Run(fmt.Sprint("reduced=", reduced), func(t *testing.T) {
			s := New()
			s.chartPreview.initialize()
			s.chartPreview.family, s.chartPreview.selected, s.chartPreview.familySelected = "donut", 1, 1
			r := chartMotionTester(s, reduced)
			r.Advance(2500 * time.Millisecond)
			if !reduced {
				advancedClick(t, r, "Animate chart changes")
				if s.chartPreview.animate {
					t.Fatal("the real native switch did not disable animations")
				}
			}
			for _, item := range []struct{ plot, action, expected string }{
				{"Visitor comparison", "Update visitor counts", "February, Desktop: 328"},
				{"Visitor story", "Update story counts", "Safari, Visitors: 310"},
			} {
				before := chartMotionPixels(t, r, item.plot)
				advancedClick(t, r, item.action)
				immediate := chartMotionPixels(t, r, item.plot)
				chartMotionRawValue(t, r, item.plot, item.expected)
				r.Advance(900 * time.Millisecond)
				if bytes.Equal(before, immediate) || !bytes.Equal(immediate, chartMotionPixels(t, r, item.plot)) {
					t.Fatal("disabled/reduced chart data did not snap directly to its final visual geometry")
				}
			}
			r.Advance(3 * time.Second)
			if r.FramePending() {
				t.Fatal("disabled/reduced demo retained an animation loop")
			}
		})
	}
}

func TestCreatorChartMotionChoicesCountsAndInspectionSurvivePreviewNavigation(t *testing.T) {
	s := New()
	s.chartPreview.initialize()
	p := &s.chartPreview
	p.family, p.selected, p.familySelected = "radar", 1, 2
	r := chartMotionTester(s, true)
	advancedClick(t, r, "Animate chart changes")
	advancedClick(t, r, "Update story counts")
	s.Page = "content"
	r.Frame()
	advancedClick(t, r, "Motion page anchor")
	s.Page = "display"
	r.Frame()
	if !r.Focused("Motion page anchor") || p.animate || p.family != "radar" || p.familySelected != 2 || p.selected != 1 || !p.storyUpdated || p.updated {
		t.Fatalf("returning to the preview reset independent choices/data/selection or stole focus: %+v", p)
	}
	chartMotionRawValue(t, r, "Visitor comparison", "February, Desktop: 305")
	chartMotionRawValue(t, r, "Visitor story", "March, Desktop: 312, Mobile: 240")
	if advancedAccess(t, r, ui.RoleSwitch, "Animate chart changes").Checked || !advancedAccess(t, r, ui.RoleButton, "Update story counts").Disabled {
		t.Fatal("returning controls disagree with preserved application state")
	}
}

func TestCreatorChartMotionWholePageNarrowRTLReachesBothCardActions(t *testing.T) {
	s := New()
	s.Page, s.Config.ReduceMotion = "display", true
	r := ui.NewTester(s.View, 500, 1000)
	r.SetDirection(ui.RightToLeft)
	chartPreviewReveal(t, r, "Animate chart changes")
	advancedClick(t, r, "Animate chart changes")
	chartPreviewReveal(t, r, "Donut")
	advancedClick(t, r, "Donut")
	chartPreviewReveal(t, r, "Update story counts")
	advancedClick(t, r, "Update story counts")
	plot := chartPreviewReveal(t, r, "Visitor story")
	advancedClick(t, r, "Visitor story")
	r.Key(0, ui.KeyHome)
	chartMotionRawValue(t, r, "Visitor story", "Chrome, Visitors: 160")
	card := advancedAccess(t, r, ui.RoleGroup, "Chart style explorer").Bounds
	if plot.W < 200 || plot.H < 200 || plot.X < card.X || plot.X+plot.W > card.X+card.W {
		t.Fatalf("narrow motion controls compressed the actual plot: plot=%v card=%v", plot, card)
	}
	for _, label := range []string{"Update story counts", "Restore story counts"} {
		chartPreviewReveal(t, r, label)
		a := advancedAccess(t, r, ui.RoleButton, label)
		if a.Offscreen || a.Bounds.X < card.X || a.Bounds.X+a.Bounds.W > card.X+card.W {
			t.Fatalf("narrow RTL card action is clipped: %s %+v", label, a)
		}
	}
	if !s.chartPreview.storyUpdated || s.chartPreview.updated || s.chartPreview.animate {
		t.Fatal("narrow native actions did not keep card data and animation choices independent")
	}
}
