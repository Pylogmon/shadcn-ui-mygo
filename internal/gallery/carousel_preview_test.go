package gallery

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
	r := carouselPreviewTester(s, 1000, 1800, false)
	advancedClick(t, r, "Next project")
	r.Advance(100 * time.Millisecond)
	r.Frame()
	if !slices.Equal(s.carousels.horizontalView.Visible, []int{0, 1, 2}) {
		t.Fatalf("moving multi-item view was not actually partial: %+v", s.carousels.horizontalView)
	}
	r.Advance(200 * time.Millisecond)
	r.Frame()
	if !slices.Equal(s.carousels.horizontalView.Visible, []int{1, 2}) {
		t.Fatal("native compatibility transition failed to settle")
	}
	advancedClick(t, r, "Next vertical project")
	r.Key(0, ui.KeyDown)
	r.Advance(300 * time.Millisecond)
	r.Frame()
	if s.carousels.vertical != 2 || !slices.Equal(s.carousels.verticalView.Visible, []int{2, 3}) || s.carousels.horizontal != 1 {
		t.Fatal("vertical keys did not navigate the independent vertical track")
	}
	advancedClick(t, r, "Next RTL project")
	r.Key(0, ui.KeyLeft)
	r.Advance(300 * time.Millisecond)
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
			r := carouselPreviewTester(s, width, 1800, true)
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
