package display

import (
	"bytes"
	"math"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestSpinnerInheritsCurrentColor(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(ui.RGB(255, 255, 255)).TextColor(ui.RGB(220, 20, 40)).Children(func() { Spinner(c).Label("Working") })
	}, 100, 100)
	box, ok := r.Find("Working")
	if !ok {
		t.Fatal("missing spinner")
	}
	colored := 0
	im := r.Image()
	for y := int(box.Y); y < int(box.Y+box.H); y++ {
		for x := int(box.X); x < int(box.X+box.W); x++ {
			p := im.RGBAAt(x, y)
			if int(p.R) > int(p.G)+40 && int(p.R) > int(p.B)+40 {
				colored++
			}
		}
	}
	if colored < 10 {
		t.Fatal("spinner did not use its parent's current text color")
	}
}

func TestSpinnerReducedMotionPixelsStayStillOnUnrelatedFrames(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(theme.TokensOf(c).Background).Children(func() { Spinner(c) })
	}, 100, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(325 * time.Millisecond)
	if !bytes.Equal(before, r.Image().Pix) || r.FramePending() {
		t.Fatal("reduced-motion spinner changed pixels or requested another frame")
	}
}

func TestSkeletonUsesVegaRadiusAndPulseCurve(t *testing.T) {
	tokens := theme.Light()
	tokens.Radius = 16
	tokens.Background, tokens.Muted = ui.RGB(255, 255, 255), ui.RGB(10, 10, 220)
	view := func(reference bool, opacity float32) func(*ui.Context) {
		return func(c *ui.Context) {
			theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
				ui.Column(c).Padding(20).Children(func() {
					if reference {
						ui.Box(c).Size(80, 40).Radius(12.8).Background(tokens.Muted).Opacity(opacity)
					} else {
						Skeleton(c, SkeletonProps{Width: 80, Height: 40})
					}
				})
			})
		}
	}
	r := ui.NewTester(view(false, 0), 130, 100)
	reference := ui.NewTester(view(true, 1), 130, 100)
	if !bytes.Equal(r.Image().Pix, reference.Image().Pix) {
		t.Fatal("skeleton radius did not use pinned CSS calc(radius * .8)")
	}
	r.Advance(250 * time.Millisecond)
	// CSS pulse has opacity .5 at 50%, using cubic-bezier(.4,0,.6,1).
	// At 250ms/2s the first half is .25, whose eased progress is .135307.
	reference = ui.NewTester(view(true, .9323464), 130, 100)
	actual, want := r.Image().RGBAAt(60, 40), reference.Image().RGBAAt(60, 40)
	if actual != want {
		t.Fatalf("pulse did not follow the source easing: pixel=%v reference=%v", actual, want)
	}
}

func TestSkeletonInstantFramesRemainStatic(t *testing.T) {
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		c.DisableAnimations()
		ui.Column(c).Fill().Background(theme.TokensOf(c).Background).Children(func() { Skeleton(c, SkeletonProps{Width: 80, Height: 40}) })
	}, 130, 100)
	before := append([]byte(nil), r.Image().Pix...)
	r.Advance(750 * time.Millisecond)
	if !bytes.Equal(before, r.Image().Pix) || r.FramePending() {
		t.Fatal("instant skeleton changed pixels or kept an animation alive")
	}
}

func TestSpinnerMatchesPinnedSVGAndOneSecondLinearRotation(t *testing.T) {
	source, err := os.ReadFile("../../internal/indicator/loader-circle.svg")
	if err != nil {
		t.Fatal(err)
	}
	icon := ui.MustParseSVG(source)
	view := func(reference bool, degrees float32) func(*ui.Context) {
		return func(c *ui.Context) {
			theme.Apply(c)
			ui.Column(c).Fill().Padding(20).TextColor(ui.RGB(40, 80, 220)).Background(ui.RGB(255, 255, 255)).Children(func() {
				if reference {
					ui.Icon(c, icon).Size(24, 24).Rotate(degrees)
				} else {
					Spinner(c, SpinnerProps{Size: 24, Label: "Processing", Description: "Loading your workspace"})
				}
			})
		}
	}
	r := ui.NewTester(view(false, 0), 100, 100)
	initial := append([]byte(nil), r.Image().Pix...)
	for _, angle := range []float32{0, 90} {
		if angle != 0 {
			r.Advance(250 * time.Millisecond)
		}
		reference := ui.NewTester(view(true, angle), 100, 100)
		if !bytes.Equal(r.Image().Pix, reference.Image().Pix) {
			t.Fatalf("loading glyph/rotation differs from pinned SVG at %g degrees", angle)
		}
	}
	progress, ok := r.AccessByRole(ui.RoleProgress, "Processing")
	if !ok || progress.Now >= progress.Min || progress.Bounds.W != 24 || progress.Bounds.H != 24 || progress.Description != "Loading your workspace" {
		t.Fatalf("native unknown-progress semantics/size=%+v", progress)
	}
	r.Advance(750 * time.Millisecond)
	if !bytes.Equal(initial, r.Image().Pix) || !r.FramePending() {
		t.Fatal("spinner did not complete one linear turn in one second")
	}
}

func TestLoadingChangesPreferencesStaticAndUnmountWithoutIdle(t *testing.T) {
	show, static, reduced := true, false, false
	view := func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Background(theme.TokensOf(c).Background).Gap(16).Children(func() {
			if show {
				Spinner(c, SpinnerProps{Static: static})
				Skeleton(c, SkeletonProps{Static: static, Width: 80, Height: 40})
			}
		})
	}
	r := ui.NewTester(view, 150, 150)
	r.Advance(300 * time.Millisecond)
	if !r.FramePending() {
		t.Fatal("visible loading primitives did not animate")
	}
	for _, mode := range []string{"reduced", "static"} {
		if mode == "reduced" {
			reduced = true
			r.SetPreferences(ui.Preferences{ReduceMotion: reduced})
		} else {
			static = true
			r.SetPreferences(ui.Preferences{})
		}
		reference := ui.NewTesterWith(view, 150, 150, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
		if !bytes.Equal(r.Image().Pix, reference.Image().Pix) || r.FramePending() {
			t.Fatalf("%s did not cancel the running primitives to their static rendering", mode)
		}
		frozen := append([]byte(nil), r.Image().Pix...)
		r.Advance(800 * time.Millisecond)
		if !bytes.Equal(frozen, r.Image().Pix) || r.FramePending() {
			t.Fatalf("%s changed pixels on unrelated frames", mode)
		}
		reduced = false
	}
	static = false
	r.Frame()
	if !r.FramePending() {
		t.Fatal("normal loading motion did not resume")
	}
	show = false
	r.Frame()
	r.Advance(3 * time.Second)
	if r.FramePending() {
		t.Fatal("removed loading primitives kept an animation alive")
	}
}

func TestLoadingFluentStylesDecorativeAndFiniteDimensions(t *testing.T) {
	for _, size := range []float32{-1, float32(math.NaN()), float32(math.Inf(1))} {
		r := ui.NewTesterWith(func(c *ui.Context) {
			theme.Apply(c)
			ui.Column(c).Width(200).Gap(12).Children(func() {
				Spinner(c, SpinnerProps{Size: size, Label: "Indicator"})
				Skeleton(c, SkeletonProps{Width: size, Height: size}).Label("Placeholder")
			})
		}, 250, 150, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		spinner, ok := r.AccessByRole(ui.RoleProgress, "Indicator")
		placeholder, exists := r.Find("Placeholder")
		if !ok || !exists || !spinner.Invalid || spinner.Bounds.W != 16 || spinner.Bounds.H != 16 || placeholder.W != 200 || placeholder.H != 16 {
			t.Fatalf("invalid dimensions escaped to layout: spinner=%+v skeleton=%+v", spinner, placeholder)
		}
	}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(200).Gap(12).Children(func() {
			Spinner(c, SpinnerProps{Decorative: true}).Size(32, 32).TextColor(ui.RGB(20, 80, 220)).Label("Decorative visual")
			Skeleton(c, SkeletonProps{}).WidthPercent(75).Label("Three quarters")
			Skeleton(c, SkeletonProps{Circle: true, Height: 40}).Label("Avatar placeholder")
			ui.Button(c, "Continue")
		})
	}, 250, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if _, exposed := r.AccessByRole(ui.RoleProgress, "Decorative visual"); exposed {
		t.Fatal("decorative spinner exposed a duplicate progress announcement")
	}
	visual, ok := r.Find("Decorative visual")
	threeQuarters, _ := r.Find("Three quarters")
	circle, _ := r.Find("Avatar placeholder")
	if !ok || visual.W != 32 || visual.H != 32 || threeQuarters.W != 150 || circle.W != 40 || circle.H != 40 {
		t.Fatalf("fluent/circle dimensions: spinner=%+v width=%+v circle=%+v", visual, threeQuarters, circle)
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("Continue") {
		t.Fatal("loading visuals inserted keyboard stops")
	}
}
