package gallery

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func chartPreviewTester(s *State, width int) *ui.Tester {
	s.Page = "display"
	return ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(12).Children(func() {
			ui.Button(c, "Chart page anchor")
			if s.Page == "display" {
				ui.Scroll(c).Fill().Children(func() { s.chartCompositionPreview(c) })
			} else {
				ui.Text(c, "Another preview")
			}
		})
	}, width, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}

func chartPreviewInspect(t *testing.T, r *ui.Tester, keys ...ui.Key) ui.AccessInfo {
	t.Helper()
	advancedClick(t, r, "Visitor comparison")
	for _, key := range keys {
		r.Key(0, key)
	}
	return advancedAccess(t, r, ui.RoleGroup, "Visitor comparison")
}

func TestCreatorChartOrientationUsesActualNativeCategoryKeys(t *testing.T) {
	s := New()
	r := chartPreviewTester(s, 800)
	info := chartPreviewInspect(t, r, ui.KeyHome, ui.KeyDown)
	if s.chartPreview.selected != 1 || !strings.Contains(info.Value, "February, Desktop: 305") || !strings.Contains(info.Description, "Up and Down") {
		t.Fatalf("horizontal preview did not inspect the actual second row: %+v", info)
	}
	advancedClick(t, r, "Vertical bars")
	info = chartPreviewInspect(t, r, ui.KeyHome, ui.KeyRight)
	if s.chartPreview.selected != 1 || !strings.Contains(info.Value, "February, Desktop: 305") || !strings.Contains(info.Description, "Left and Right") {
		t.Fatalf("vertical preview did not switch actual category navigation: %+v", info)
	}
	r.Key(0, ui.KeyEscape)
	if s.chartPreview.selected != -1 {
		t.Fatal("native chart Escape did not clear inspection")
	}
}

func TestCreatorChartGroupedStackedAndShareKeepOriginalInspectedCounts(t *testing.T) {
	s := New()
	r := chartPreviewTester(s, 800)
	for _, label := range []string{"Side by side", "Stacked totals", "Share of total"} {
		advancedClick(t, r, label)
		info := chartPreviewInspect(t, r, ui.KeyHome)
		if !strings.Contains(info.Value, "Desktop: 186") || !strings.Contains(info.Value, "Mobile: 80") || !r.HasText("186") || !r.HasText("80") {
			t.Fatalf("%s replaced original inspected values: %+v", label, info)
		}
	}
	advancedClick(t, r, "Show axis values")
	if !r.HasText("100%") {
		t.Fatal("share axis did not use the independent percentage formatter")
	}
	info := chartPreviewInspect(t, r, ui.KeyEnd)
	if !strings.Contains(info.Value, "June, Desktop: 214, Mobile: 140") || r.HasText("21400%") {
		t.Fatalf("percentage axis changed the original tooltip values: %+v", info)
	}
}

func TestCreatorChartBrowserColorsLegendAndNegativeChangeAreRealViews(t *testing.T) {
	s := New()
	r := chartPreviewTester(s, 800)
	advancedClick(t, r, "Month-to-month change")
	info := chartPreviewInspect(t, r, ui.KeyHome, ui.KeyDown, ui.KeyDown)
	if !strings.Contains(info.Value, "March, Desktop change: -207") {
		t.Fatalf("change view did not render the negative observation: %+v", info)
	}
	advancedClick(t, r, "Browsers")
	advancedClick(t, r, "Show chart legend")
	advancedAccess(t, r, ui.RoleGroup, "Chart legend")
	if !advancedAccess(t, r, ui.RoleSwitch, "Month-to-month change").Disabled || !r.HasText("Chrome") || !r.HasText("Firefox") {
		t.Fatal("browser view did not switch the native controls and category legend")
	}
	colors := map[[3]uint8]bool{}
	image := r.Image()
	for _, mark := range []string{"Visitors Chrome: 275", "Visitors Safari: 200", "Visitors Firefox: 187", "Visitors Edge: 173", "Visitors Other: 90"} {
		bounds, ok := r.Find(mark)
		if !ok || bounds.W <= 0 || bounds.H <= 0 {
			t.Fatalf("browser mark is absent: %s %v", mark, bounds)
		}
		pixel := image.RGBAAt(int(bounds.X+bounds.W/2), int(bounds.Y+bounds.H/2))
		colors[[3]uint8{pixel.R, pixel.G, pixel.B}] = true
	}
	if len(colors) < 5 {
		t.Fatalf("browser bars lost their distinct category fills: %v", colors)
	}
	info = chartPreviewInspect(t, r, ui.KeyEnd)
	if !strings.Contains(info.Value, "Other, Visitors: 90") {
		t.Fatalf("browser inspection kept month data: %+v", info)
	}
}

func TestCreatorChartUpdatesReachNativeMarksAndRestoreWithoutAnExtraInput(t *testing.T) {
	s := New()
	r := chartPreviewTester(s, 800)
	advancedClick(t, r, "Update visitor counts")
	if !advancedAccess(t, r, ui.RoleButton, "Update visitor counts").Disabled || advancedAccess(t, r, ui.RoleButton, "Restore visitor counts").Disabled {
		t.Fatal("update did not settle its real native actions")
	}
	if _, ok := r.Find("Desktop January: 220"); !ok {
		t.Fatal("updated counts did not reach the already-built native plot")
	}
	info := chartPreviewInspect(t, r, ui.KeyHome)
	if !strings.Contains(info.Value, "January, Desktop: 220") {
		t.Fatalf("updated chart inspected stale data: %+v", info)
	}
	advancedClick(t, r, "Restore visitor counts")
	if _, ok := r.Find("Desktop January: 186"); !ok {
		t.Fatal("restore did not update the native plot in its own action")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reduced chart preview kept requesting idle frames")
	}
}

func TestCreatorChartViewAndSavedValuesSurvivePreviewNavigation(t *testing.T) {
	s := New()
	r := chartPreviewTester(s, 800)
	advancedClick(t, r, "Vertical bars")
	advancedClick(t, r, "Browsers")
	advancedClick(t, r, "Show chart legend")
	advancedClick(t, r, "Update visitor counts")
	s.Page = "content"
	r.Frame()
	advancedClick(t, r, "Chart page anchor")
	s.Page = "display"
	r.Frame()
	if !r.Focused("Chart page anchor") || !advancedAccess(t, r, ui.RoleRadio, "Vertical bars").Checked || !advancedAccess(t, r, ui.RoleRadio, "Browsers").Checked || !advancedAccess(t, r, ui.RoleSwitch, "Show chart legend").Checked {
		t.Fatal("chart page reset application choices or stole focus on return")
	}
	info := chartPreviewInspect(t, r, ui.KeyHome)
	if !strings.Contains(info.Value, "Chrome, Visitors: 302") {
		t.Fatalf("application-owned updated data was discarded: %+v", info)
	}
}

func chartPreviewReveal(t *testing.T, r *ui.Tester, name string) ui.Rect {
	t.Helper()
	for attempt := 0; attempt < 6; attempt++ {
		target, exists := r.Access(name)
		panel, panelExists := r.AccessByRole(ui.RoleRegion, "Data & display")
		if !exists || !panelExists {
			t.Fatalf("missing actual chart target/panel: %q %+v", name, target)
		}
		if !target.Offscreen && target.Bounds.H > 0 && target.Bounds.Y >= panel.Bounds.Y && target.Bounds.Y+target.Bounds.H <= panel.Bounds.Y+panel.Bounds.H {
			return target.Bounds
		}
		delta := target.Bounds.Y + target.Bounds.H/2 - panel.Bounds.Y - panel.Bounds.H/2
		r.Scroll(panel.Bounds.X+panel.Bounds.W-8, panel.Bounds.Y+panel.Bounds.H/2, 0, delta)
	}
	t.Fatalf("actual Gallery could not reveal %q", name)
	return ui.Rect{}
}

func TestCreatorChartWholePageNarrowAndWideKeepPlotAndControlsReachable(t *testing.T) {
	for _, width := range []int{500, 1450} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			s := New()
			s.Page, s.Config.ReduceMotion = "display", true
			r := ui.NewTester(s.View, width, 900)
			if width == 500 {
				r.SetDirection(ui.RightToLeft)
			}
			chartPreviewReveal(t, r, "Browsers")
			advancedClick(t, r, "Browsers")
			plot := chartPreviewReveal(t, r, "Visitor comparison")
			if plot.W < 150 || plot.H < 200 {
				t.Fatalf("full-page chart plot was compressed: %v", plot)
			}
			info := chartPreviewInspect(t, r, ui.KeyHome, ui.KeyDown)
			if !strings.Contains(info.Value, "Safari, Visitors: 200") {
				t.Fatalf("whole-page chart was not actually usable: %+v", info)
			}
			chartPreviewReveal(t, r, "Update visitor counts")
			advancedClick(t, r, "Update visitor counts")
			card := advancedAccess(t, r, ui.RoleGroup, "Visitor explorer").Bounds
			for _, label := range []string{"Horizontal bars", "Vertical bars", "Browsers", "Show chart legend", "Update visitor counts", "Restore visitor counts"} {
				control, ok := r.Access(label)
				if !ok || control.Bounds.X < card.X-.5 || control.Bounds.X+control.Bounds.W > card.X+card.W+.5 || math.IsNaN(float64(control.Bounds.Y)) {
					t.Fatalf("wrapped control escaped its chart card: %s %+v card%v", label, control, card)
				}
			}
		})
	}
}

func TestCreatorChartOtherFamiliesRenderAndUseNativeInspection(t *testing.T) {
	for _, family := range []string{"Line", "Area", "Pie", "Donut", "Radar", "Radial"} {
		t.Run(family, func(t *testing.T) {
			s := New()
			s.chartPreview.initialize()
			var palette [5]ui.Color
			r := ui.NewTesterWith(func(c *ui.Context) {
				s.Theme.Apply(c)
				palette = theme.TokensOf(c).Charts()
				ui.Column(c).Fill().Padding(20).Children(func() { s.chartFamilyPreview(c) })
			}, 620, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			advancedClick(t, r, family)
			advancedClick(t, r, "Visitor story")
			r.Key(0, ui.KeyHome)
			value := advancedAccess(t, r, ui.RoleGroup, "Visitor story").Value
			if family == "Pie" || family == "Donut" || family == "Radial" {
				if !strings.Contains(value, "Chrome, Visitors: 275") {
					t.Fatalf("%s did not inspect its category data: %s", family, value)
				}
			} else if !strings.Contains(value, "January, Desktop: 186, Mobile: 80") {
				t.Fatalf("%s did not inspect both real series: %s", family, value)
			}
			r.Key(0, ui.KeyRight)
			if s.chartPreview.familySelected != 1 {
				t.Fatal("family view did not use the native category inspection")
			}
			plot := advancedAccess(t, r, ui.RoleGroup, "Visitor story").Bounds
			polar := family != "Line" && family != "Area"
			if polar && (plot.W != 240 || plot.H != 240) || !polar && plot.W <= 300 {
				t.Fatalf("%s used the wrong native plot geometry: %v", family, plot)
			}
			if polar {
				image := r.Image()
				cx, cy := int(plot.X+plot.W/2), int(plot.Y+plot.H/2)
				background := image.RGBAAt(int(plot.X+1), int(plot.Y+plot.H-1))
				centerInk, solidSeriesInk := 0, 0
				for y := cy - 10; y < cy+10; y++ {
					for x := cx - 10; x < cx+10; x++ {
						if image.RGBAAt(x, y) != background {
							centerInk++
						}
					}
				}
				if family == "Pie" && centerInk < 200 || (family == "Donut" || family == "Radial") && centerInk != 0 {
					t.Fatalf("%s painted the wrong real center/hole: %d", family, centerInk)
				}
				if family == "Donut" || family == "Radial" {
					inside := image.RGBAAt(cx+40, cy+15)
					if family == "Donut" && inside != background || family == "Radial" && inside == background {
						t.Fatalf("%s did not paint its actual annular geometry", family)
					}
				}
				if family == "Radar" {
					for y := int(plot.Y); y < int(plot.Y+plot.H); y++ {
						for x := int(plot.X); x < int(plot.X+plot.W); x++ {
							pixel := image.RGBAAt(x, y)
							for _, color := range palette[:2] {
								if pixel.R == color.R && pixel.G == color.G && pixel.B == color.B {
									solidSeriesInk++
									break
								}
							}
						}
					}
					if solidSeriesInk == 0 || solidSeriesInk > 5000 {
						t.Fatalf("radar chart lost its thin native polygon outlines: %d", solidSeriesInk)
					}
				}
			}
			r.Advance(time.Second)
			if r.FramePending() {
				t.Fatal("static family chart kept requesting frames")
			}
		})
	}
}

func TestCreatorChartPeerPolarChartDoesNotCaptureBarKeys(t *testing.T) {
	s := New()
	r := chartPreviewTester(s, 800)
	s.chartPreview.family = "pie"
	r.Frame()
	chartPreviewInspect(t, r, ui.KeyHome, ui.KeyDown)
	if s.chartPreview.selected != 1 || s.chartPreview.familySelected != -1 {
		t.Fatal("a peer polar chart stole the focused bar's native key")
	}
}
