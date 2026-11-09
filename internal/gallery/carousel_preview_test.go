package gallery

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func carouselPreviewTester(s *State, width, height int, reduced bool) *ui.Tester {
	s.Page = "display"
	return ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		tokens := s.Config.Tokens(c.Theme().Dark)
		theme.Scope(c, tokens).Fill().Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			ui.Scroll(c).Fill().Padding(20).Children(func() {
				if s.Page == "display" {
					s.carouselPreview(c)
				} else {
					control.Input(c, &s.Name, control.InputProps{Label: "Another carousel page draft"})
				}
			})
		})
	}, width, height, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
}

func TestCreatorCarouselLayoutSpacingAndLoopAreRealGeometry(t *testing.T) {
	s := New()
	r := carouselPreviewTester(s, 1000, 1200, true)
	p := &s.carousels
	if p.horizontalView.SlideCount != 5 || len(p.horizontalView.Snaps) != 4 || !slices.Equal(p.horizontalView.Visible, []int{0, 1}) {
		t.Fatalf("default two-item composition=%+v", p.horizontalView)
	}
	for _, choice := range []struct {
		label string
		snaps int
	}{{"One card", 5}, {"Three cards", 3}, {"Two cards", 4}} {
		advancedClick(t, r, choice.label)
		if len(p.horizontalView.Snaps) != choice.snaps || !r.HasText(fmt.Sprintf("Position 1 of %d · 5 projects", choice.snaps)) {
			t.Fatalf("%s did not change actual track and status: %+v", choice.label, p.horizontalView)
		}
	}
	before := p.horizontalView.Snaps[1]
	advancedClick(t, r, "Compact card spacing")
	if math.Abs(float64(p.horizontalView.Snaps[1]-before-6)) > .01 {
		t.Fatalf("16→4 spacing did not change half-item geometry by6: before=%g now=%v", before, p.horizontalView.Snaps)
	}
	advancedClick(t, r, "Loop project cards")
	if !p.horizontalView.Loop || advancedAccess(t, r, ui.RoleButton, "Previous project").Disabled {
		t.Fatal("loop switch did not enable the previous boundary")
	}
	advancedClick(t, r, "Previous project")
	if p.horizontal != 4 {
		t.Fatalf("loop previous selected=%d want4", p.horizontal)
	}
	advancedClick(t, r, "Show only two projects")
	if p.horizontal != 0 || p.horizontalView.Loop || len(p.horizontalView.Snaps) != 1 || !r.HasText("This layout needs more projects to loop.") {
		t.Fatalf("short collection did not normalize actual loop/snaps: %+v", p.horizontalView)
	}
	if !advancedAccess(t, r, ui.RoleButton, "Previous project").Disabled || !advancedAccess(t, r, ui.RoleButton, "Next project").Disabled {
		t.Fatal("one-snap collection left a live navigation button")
	}
	advancedClick(t, r, "Show only two projects")
	advancedClick(t, r, "Fixed card width")
	if p.horizontalView.Snaps[1] != -220 {
		t.Fatalf("fixed item geometry did not use220DIP: %v", p.horizontalView.Snaps)
	}
	r.SetSize(1100, 1200)
	if p.horizontalView.Loop || len(p.horizontalView.Snaps) < 2 || !r.HasText("This layout needs more projects to loop.") {
		t.Fatal("loop fallback incorrectly implied that every project fits in one position")
	}
}

func TestCreatorCarouselDraftFollowsKeyAcrossNavigationReorderAndPages(t *testing.T) {
	s := New()
	r := carouselPreviewTester(s, 1000, 1200, true)
	advancedClick(t, r, "Notes for Discover")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Ship the native app")
	advancedClick(t, r, "Next project")
	if _, ok := r.AccessByRole(ui.RoleTextField, "Notes for Discover"); ok {
		t.Fatal("fully departed slide editor remained in AX")
	}
	advancedClick(t, r, "Previous project")
	if advancedAccess(t, r, ui.RoleTextField, "Notes for Discover").Value != "Ship the native app" {
		t.Fatal("navigation discarded the real editor's draft")
	}
	advancedClick(t, r, "Reverse projects")
	if s.carousels.order[0] != 4 || s.carousels.notes[0] != "Ship the native app" {
		t.Fatal("reorder moved text to positional identity")
	}
	s.carousels.horizontal = 3
	r.Frame()
	if advancedAccess(t, r, ui.RoleTextField, "Notes for Discover").Value != "Ship the native app" {
		t.Fatal("reordered keyed slide did not show its own draft")
	}
	s.Page = "overview"
	r.Frame()
	advancedClick(t, r, "Another carousel page draft")
	r.Type("Keep this focus")
	r.Advance(time.Second)
	if !r.Focused("Another carousel page draft") {
		t.Fatal("removed carousel stole the new page's focus")
	}
	s.Page = "display"
	r.Frame()
	if s.carousels.order[0] != 4 || advancedAccess(t, r, ui.RoleTextField, "Notes for Discover").Value != "Ship the native app" {
		t.Fatal("page navigation lost the caller-owned carousel state")
	}
}

func TestCreatorCarouselNativeMotionVerticalAndRTLKeys(t *testing.T) {
	s := New()
	r := carouselPreviewTester(s, 1000, 2500, false)
	advancedClick(t, r, "Next project")
	r.Advance(100 * time.Millisecond)
	r.Frame()
	if !slices.Equal(s.carousels.horizontalView.Visible, []int{0, 1, 2}) {
		t.Fatalf("moving multi-item view was not actually partial: %+v", s.carousels.horizontalView)
	}
	r.Advance(2 * time.Second)
	r.Frame()
	if !slices.Equal(s.carousels.horizontalView.Visible, []int{1, 2}) {
		t.Fatal("source physics transition failed to settle")
	}
	advancedClick(t, r, "Next vertical project")
	r.Key(0, ui.KeyDown)
	r.Advance(2 * time.Second)
	r.Frame()
	if s.carousels.vertical != 2 || !slices.Equal(s.carousels.verticalView.Visible, []int{2, 3}) || s.carousels.horizontal != 1 {
		t.Fatal("vertical keys did not navigate the independent vertical track")
	}
	advancedClick(t, r, "Next RTL project")
	r.Key(0, ui.KeyLeft)
	r.Advance(2 * time.Second)
	r.Frame()
	if s.carousels.rtl != 2 || !slices.Equal(s.carousels.rtlView.Visible, []int{2}) {
		t.Fatal("RTL Left did not move forward along the real RTL track")
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	advancedClick(t, r, "Previous project")
	if s.carousels.horizontal != 0 || !slices.Equal(s.carousels.horizontalView.Visible, []int{0, 1}) {
		t.Fatal("Reduce motion did not immediately update the composed preview")
	}
}

func TestCreatorCarouselSmallWindowCanReachBothEndsByNativeScroll(t *testing.T) {
	s := New()
	r := carouselPreviewTester(s, 500, 600, true)
	r.Scroll(440, 500, 0, 1800)
	advancedClick(t, r, "Next RTL project")
	if s.carousels.rtl != 1 {
		t.Fatal("small-window scrolling did not reach the final example's real control")
	}
	r.Scroll(440, 500, 0, -1800)
	advancedClick(t, r, "Notes for Discover")
	r.Type("A reachable draft")
	if s.carousels.notes[0] != "A reachable draft" {
		t.Fatal("scrolling back did not restore the first native editor's input target")
	}
}

func TestCreatorCarouselResponsiveControlsRemainInsideCards(t *testing.T) {
	for _, width := range []int{500, 800, 1280} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			s := New()
			s.Theme.Mode = theme.DarkMode
			r := carouselPreviewTester(s, width, 2500, true)
			advancedClick(t, r, "Responsive cards")
			want := 5
			if width >= 1024 {
				want = 3
			} else if width >= 640 {
				want = 4
			}
			if len(s.carousels.horizontalView.Snaps) != want {
				t.Fatalf("responsive snapCount=%d want%d", len(s.carousels.horizontalView.Snaps), want)
			}
			for _, label := range []string{"Previous project", "Next project", "Previous vertical project", "Next vertical project", "Previous RTL project", "Next RTL project"} {
				n := advancedAccess(t, r, ui.RoleButton, label)
				if n.Bounds.W <= 0 || n.Bounds.X < 20 || n.Bounds.X+n.Bounds.W > float32(width-20)+.01 {
					t.Fatalf("%s overflowed %d-wide preview: %+v", label, width, n.Bounds)
				}
			}
		})
	}
}

func TestCreatorCarouselSearchEntersActualDataPage(t *testing.T) {
	for _, query := range []string{"Carousel", "Slides", "Slideshow"} {
		t.Run(query, func(t *testing.T) {
			s := New()
			r := ui.NewTester(func(c *ui.Context) { c.DisableAnimations(); s.View(c) }, 1280, 900)
			advancedClick(t, r, "Search components")
			r.Type(query)
			r.Key(0, ui.KeyEnter)
			if s.Command || s.Page != "display" || !s.carousels.initialized || !r.HasText("Carousel") {
				t.Fatalf("%q did not navigate to the actual source composition", query)
			}
		})
	}
}

func TestCreatorCarouselControllerAutoplayAndSourceConfigurations(t *testing.T) {
	s := New()
	r := carouselPreviewTester(s, 1000, 2800, true)
	p := &s.carousels
	if !p.controllers[0].Snapshot().Ready || !p.autoplay.IsPlaying() {
		t.Fatal("gallery did not bind real controller/plugin")
	}
	advancedClick(t, r, "Jump to last")
	if p.horizontal != len(p.horizontalView.Snaps)-1 || p.selectionEvents == 0 {
		t.Fatal("API example did not execute controller/listener")
	}
	advancedClick(t, r, "Instant motion")
	advancedClick(t, r, "Jump to first")
	if p.controllers[0].EngineSnapshot().Options.Duration != 0 || p.horizontal != 0 {
		t.Fatal("motion control did not configure source option")
	}
	advancedClick(t, r, "Responsive item size")
	advancedClick(t, r, "Compact card spacing")
	advancedClick(t, r, "Centered alignment")
	engine := p.controllers[0].EngineSnapshot()
	if engine.Options.Align.Mode != display.CarouselAlignCenter || math.Abs(p.controllers[0].Snapshot().SnapOffsets[1]+engine.Viewport/4) > .01 {
		t.Fatal("source spacing example was only a cosmetic preset")
	}
	r.Advance(2 * time.Second)
	if p.autoplaySelected != 1 {
		t.Fatal("source plugin delay did not advance actual cards")
	}
	advancedClick(t, r, "Pause autoplay")
	paused := p.autoplaySelected
	r.Advance(3 * time.Second)
	if p.autoplay.IsPlaying() || p.autoplaySelected != paused {
		t.Fatal("autoplay pause retained a live action")
	}
	advancedClick(t, r, "Play with instant transitions")
	r.Advance(2 * time.Second)
	if p.autoplaySelected == paused || p.controllers[3].EngineSnapshot().Duration != 0 {
		t.Fatal("jump override did not reach physics/controller")
	}
	oldController, oldAutoplay := p.controllers[0], p.controllers[3]
	saved, autoSelected := p.horizontal, p.autoplaySelected
	s.Page = "overview"
	r.Frame()
	r.Advance(3 * time.Second)
	if !oldController.Snapshot().Destroyed || !oldAutoplay.Snapshot().Destroyed || p.autoplay.IsPlaying() {
		t.Fatal("page removal retained controller/plugin")
	}
	s.Page = "display"
	r.Frame()
	if p.controllers[0] == oldController || p.controllers[3] == oldAutoplay || p.horizontal != saved || p.autoplaySelected != autoSelected || !p.autoplay.IsPlaying() {
		t.Fatal("page remount revived destroyed controller or lost application selection")
	}
}
