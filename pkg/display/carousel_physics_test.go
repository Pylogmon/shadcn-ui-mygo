package display

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type carouselPhysicsState struct {
	Selected, Previous                                             int
	Location, Offset, PreviousLocation, Target, Velocity, Duration float64
	PointerDown                                                    bool
	Event                                                          string
}

func carouselPhysicsSample(api *CarouselController) carouselPhysicsState {
	s := api.Snapshot()
	e := s.Engine
	return carouselPhysicsState{Selected: s.Selected, Previous: s.Previous, Location: e.Location, Offset: e.OffsetLocation, PreviousLocation: e.PreviousLocation, Target: e.Target, Velocity: e.Velocity, Duration: e.Duration, PointerDown: e.PointerDown}
}
func carouselPhysicsTester(api *CarouselController, selected *int, reduced bool, extra func(*ui.Context), before, after *CarouselSnapshot) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(64).Children(func() {
			CarouselContainer(c, selected, CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, Controller: api}, func(c *ui.Context) {
				if before != nil {
					*before = CarouselSnapshotOf(c)
				}
				CarouselContent(c, CarouselContentProps{Spacing: CarouselValue(float32(0))}, func(c *ui.Context) {
					for i := range 5 {
						CarouselItem(c, i, CarouselItemProps{Key: i, Basis: CarouselFraction(.5)}, func(c *ui.Context) {
							carouselColorSlide(c, i)
							if i == 0 && extra != nil {
								extra(c)
							}
						})
					}
				})
				if after != nil {
					*after = CarouselSnapshotOf(c)
				}
			})
		})
	}, 728, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
}
func assertCarouselPhysicsState(t *testing.T, label string, got, want carouselPhysicsState) {
	t.Helper()
	if got.Selected != want.Selected || got.Previous != want.Previous || got.PointerDown != want.PointerDown || got.Event != want.Event {
		t.Fatalf("%s state got=%+v want=%+v", label, got, want)
	}
	a, b := []float64{got.Location, got.Offset, got.PreviousLocation, got.Target, got.Velocity, got.Duration}, []float64{want.Location, want.Offset, want.PreviousLocation, want.Target, want.Velocity, want.Duration}
	for i := range a {
		if math.IsNaN(a[i]) || math.IsInf(a[i], 0) || math.IsNaN(b[i]) || math.IsInf(b[i], 0) || math.Abs(a[i]-b[i]) > 1e-7 {
			t.Fatalf("%s vector%d got=%+v want=%+v", label, i, got, want)
		}
	}
}

// Every sample and event comes from the actual pinned JS Engine/DragHandler.
// The existing kernel oracle independently covers all primitive corner cases.
func TestCarouselPhysicsNativeOfficialEngine(t *testing.T) {
	var fixture struct {
		SourceSHA256 string
		FrameNS      int64
		Cases        []struct {
			Name    string
			Pointer string
			Options struct {
				Duration                             *float64
				Loop, WatchDrag, DragFree, SkipSnaps *bool
			}
			Actions []struct {
				Action struct {
					Kind  string
					Index int
					Jump  bool
					X, Y  float32
					MS    float64
				}
				State  carouselPhysicsState
				Events []carouselPhysicsState
			}
		}
	}
	data, err := os.ReadFile("testdata/shadcn-carousel/physics/engine-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SourceSHA256 != "a90f9c822bd7102e4b13c985c7926f910c13fc299a368f72c9c044eb730b2507" || len(fixture.Cases) != 12 || fixture.FrameNS != 16666666 {
		t.Fatal("unrecognized official engine fixture")
	}
	for _, tc := range fixture.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			o := tc.Options
			api, err := NewCarouselController(CarouselOptions{Duration: o.Duration, Loop: o.Loop, WatchDrag: o.WatchDrag, DragFree: o.DragFree, SkipSnaps: o.SkipSnaps})
			if err != nil {
				t.Fatal(err)
			}
			selected := 0
			var pre, post CarouselSnapshot
			r := carouselPhysicsTester(api, &selected, false, nil, &pre, &post)
			if tc.Pointer == "touch" {
				r.UsePointer(ui.PointerTouch, .5)
			}
			var events []carouselPhysicsState
			listener := NewCarouselListener(func(a *CarouselController, e CarouselEventType) {
				s := carouselPhysicsSample(a)
				s.Event = string(e)
				events = append(events, s)
			})
			for _, event := range []CarouselEventType{CarouselScroll, CarouselSelect, CarouselSettle, CarouselPointerDown, CarouselPointerUp} {
				api.On(event, listener)
			}
			for n, step := range tc.Actions {
				action := step.Action
				switch action.Kind {
				case "next":
					api.ScrollNext(action.Jump)
					r.Frame()
				case "prev":
					api.ScrollPrev(action.Jump)
					r.Frame()
				case "to":
					api.ScrollTo(action.Index, action.Jump)
					r.Frame()
				case "frame":
					r.Advance(time.Duration(fixture.FrameNS))
				case "hide":
					r.SetWindowHidden(true)
				case "show":
					r.SetWindowHidden(false)
				case "idle":
					r.Advance(time.Duration(action.MS * float64(time.Millisecond)))
				case "down":
					r.Press(64+action.X, 124)
				case "move":
					r.Move(64+action.X, 124)
				case "up":
					r.Release(64+action.X, 124)
				}
				label := tc.Name + "/" + action.Kind
				if len(events) != len(step.Events) {
					t.Fatalf("step%d %s event count got=%+v want=%+v", n, label, events, step.Events)
				}
				for i := range events {
					assertCarouselPhysicsState(t, label, events[i], step.Events[i])
				}
				assertCarouselPhysicsState(t, label, carouselPhysicsSample(api), step.State)
				if action.Kind == "frame" && (pre.SelectedSnap != post.SelectedSnap || !slices.Equal(pre.Visible, post.Visible)) {
					t.Fatalf("independent pre/post composition pose mismatch: %+v/%+v", pre, post)
				}
				if action.Kind == "frame" && (n == 2 || n == 5 || n == 12) {
					assertCarouselPhysicsPaint(t, r, step.State.Offset, o.Loop != nil && *o.Loop)
				}
				events = nil
			}
		})
	}
}

func TestCarouselPhysicsOptionsCopyMergeAndMedia(t *testing.T) {
	duration, threshold, free, skip := 40., 12., true, true
	options := CarouselOptions{Duration: &duration, DragThreshold: &threshold, DragFree: &free, SkipSnaps: &skip, Breakpoints: []CarouselBreakpoint{{Key: "wide", When: CarouselMedia{MinWidth: CarouselValue(float32(700))}, Options: CarouselOptions{Duration: CarouselValue(18.), DragThreshold: CarouselValue(0.), DragFree: CarouselValue(false)}}}}
	api, err := NewCarouselController(options)
	if err != nil {
		t.Fatal(err)
	}
	duration, threshold, free, skip = 0, 0, false, false
	base := carouselResolveOptions(api.baseOptions(), CarouselEnvironment{Width: 600})
	wide := carouselResolveOptions(api.baseOptions(), CarouselEnvironment{Width: 800})
	if base.Duration != 40 || base.DragThreshold != 12 || !base.DragFree || !base.SkipSnaps || wide.Duration != 18 || wide.DragThreshold != 0 || wide.DragFree || !wide.SkipSnaps {
		t.Fatalf("copy/partial/media lost physical fields: %+v %+v", base, wide)
	}
	for _, bad := range []float64{-1, math.NaN(), math.Inf(1)} {
		for _, o := range []CarouselOptions{{Duration: CarouselValue(bad)}, {DragThreshold: CarouselValue(bad)}} {
			if o.Validate() == nil {
				t.Fatalf("accepted invalid physical option %+v", o)
			}
		}
	}
	selected := 0
	r := carouselPhysicsTester(api, &selected, false, nil, nil, nil)
	if api.EngineSnapshot().Options.Duration != 18 {
		t.Fatal("window media did not resolve source duration")
	}
	api.Configure(CarouselOptions{Duration: CarouselValue(0.), DragFree: CarouselValue(false), SkipSnaps: CarouselValue(false)})
	r.Frame()
	r.SetSize(600, 400)
	opts := api.EngineSnapshot().Options
	if opts.Duration != 0 || opts.DragFree || opts.SkipSnaps {
		t.Fatalf("explicit zero/false not applied after media stopped matching: %+v", opts)
	}
}

func TestCarouselPhysicsHiddenAndReducedPolicy(t *testing.T) {
	api, _ := NewCarouselController(CarouselOptions{})
	selected := 0
	r := carouselPhysicsTester(api, &selected, false, nil, nil, nil)
	api.ScrollTo(3, false)
	r.Frame()
	r.Advance(50 * time.Millisecond)
	before := api.EngineSnapshot()
	r.SetWindowHidden(true)
	r.Advance(5 * time.Second)
	if got := api.EngineSnapshot(); got.Location != before.Location || got.OffsetLocation != before.OffsetLocation {
		t.Fatal("hidden scheduler advanced")
	}
	r.SetWindowHidden(false)
	resumed := api.EngineSnapshot()
	if math.Abs(resumed.Location-before.Location) > math.Abs(before.Target-before.Location)/2 {
		t.Fatal("visible resume caught up hidden wall time")
	}
	r.Advance(3 * time.Second)
	if r.FramePending() || math.Abs(api.EngineSnapshot().Target-api.EngineSnapshot().OffsetLocation) > .001 {
		t.Fatal("resumed physics did not settle")
	}
	reducedAPI, _ := NewCarouselController(CarouselOptions{})
	selected = 0
	reduced := carouselPhysicsTester(reducedAPI, &selected, true, nil, nil, nil)
	reducedAPI.ScrollNext(false)
	reduced.Frame()
	if selected != 1 || reduced.FramePending() || reducedAPI.EngineSnapshot().Location != -300 || reducedAPI.EngineSnapshot().OffsetLocation != -300 {
		t.Fatalf("native reduced policy left motion %+v", reducedAPI.Snapshot())
	}
}

func TestCarouselPhysicsCaptureCancelThresholdAndChildButton(t *testing.T) {
	api, _ := NewCarouselController(CarouselOptions{DragThreshold: CarouselValue(500.)})
	selected, clicks := 0, 0
	r := carouselPhysicsTester(api, &selected, false, func(c *ui.Context) {
		button := ui.ButtonBase(c).Size(180, 60).Label("Child action").Children(func() { ui.Text(c, "Action") })
		if button.Clicked() {
			clicks++
		}
	}, nil, nil)
	r.Click("Child action")
	if clicks != 1 {
		t.Fatal("carousel stole child tap")
	}
	r.Press(514, 154)
	r.Advance(20 * time.Millisecond)
	r.Move(364, 154)
	r.Advance(20 * time.Millisecond)
	r.Release(364, 154)
	// This navigation occurs BELOW the configured click threshold: the source
	// navigation threshold and click threshold are distinct contracts.
	if selected == 0 || api.EngineSnapshot().ClickSuppressed {
		t.Fatalf("click threshold used as navigation threshold: %+v", api.Snapshot())
	}
	r.Advance(3 * time.Second)
	api.Configure(CarouselOptions{DragThreshold: CarouselValue(10.)})
	api.ScrollTo(0, true)
	r.Frame()
	r.Press(514, 154)
	r.Advance(20 * time.Millisecond)
	r.Move(400, 154)
	if !api.EngineSnapshot().PointerDown || !api.EngineSnapshot().ClickSuppressed {
		t.Fatal("captured threshold/press state missing")
	}
	r.CancelPress()
	r.Advance(3 * time.Second)
	if selected != 0 || api.EngineSnapshot().PointerDown || math.Abs(api.EngineSnapshot().OffsetLocation) > .001 {
		t.Fatalf("cancel committed a fling: %+v", api.Snapshot())
	}
	// An interactive descendant owns its pointer in the public SDK; its ordinary
	// click remains usable. No parent takeover or click suppression is claimed.
	if !slices.Contains(api.SlidesInView(), 0) {
		t.Fatal("cancel lost slide visibility")
	}
}

func assertCarouselPhysicsPaint(t *testing.T, r *ui.Tester, offset float64, loop bool) {
	t.Helper()
	im := r.Image()
	colors := []color.RGBA{{R: 220, A: 255}, {G: 170, A: 255}, {B: 220, A: 255}, {R: 200, G: 100, A: 255}, {R: 170, B: 170, A: 255}}
	for _, x := range []float64{80, 180, 400, 550} {
		point := x - offset
		if loop {
			point = math.Mod(math.Mod(point, 1500)+1500, 1500)
		}
		index := int(math.Floor(point / 300))
		if index < 0 || index >= 5 || math.Mod(point, 300) < 2 || math.Mod(point, 300) > 298 {
			continue
		}
		if got := im.RGBAAt(64+int(x), 124); got != colors[index] {
			t.Fatalf("native paint offset%.9f x%.0f got%+v expected slide%d=%+v", offset, x, got, index, colors[index])
		}
	}
	for i := range 5 {
		start := float64(i*300) + offset
		if loop {
			for start >= 600 {
				start -= 1500
			}
			for start+300 <= 0 {
				start += 1500
			}
		}
		intersects := start < 600-.01 && start+300 > .01
		_, visible := r.AccessByRole(ui.RoleGroup, fmt.Sprintf("Slide %d of 5", i+1))
		if intersects != visible {
			t.Fatalf("native AX offset%.9f slide%d start%.9f visible%v expected%v", offset, i, start, visible, intersects)
		}
	}
}

func TestCarouselPhysicsNativeAxesAndSingleSnap(t *testing.T) {
	for _, mode := range []struct {
		name      string
		axis      CarouselAxis
		direction CarouselDirection
	}{{"rtl", CarouselHorizontal, CarouselRTL}, {"vertical", CarouselVertical, CarouselLTR}, {"verticalRTL", CarouselVertical, CarouselRTL}} {
		t.Run(mode.name, func(t *testing.T) {
			api, _ := NewCarouselController(CarouselOptions{Axis: &mode.axis, Direction: &mode.direction})
			selected := 0
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(64).Children(func() {
					CarouselContainer(c, &selected, CarouselContainerProps{Count: 5, Height: 600, Controller: api, Align: CarouselAlignStart}, func(c *ui.Context) {
						CarouselContent(c, CarouselContentProps{Spacing: CarouselValue(float32(0))}, func(c *ui.Context) {
							for i := range 5 {
								CarouselItem(c, i, CarouselItemProps{Key: i, Basis: CarouselFraction(.5)}, func(c *ui.Context) { carouselColorSlide(c, i) })
							}
						})
					})
				})
			}, 728, 728)
			x, y, mx, my := float32(514), float32(364), float32(364), float32(364)
			if mode.axis == CarouselVertical {
				x, y, mx, my = 364, 514, 364, 364
			} else {
				x, mx = 214, 364
			}
			r.Press(x, y)
			r.Advance(20 * time.Millisecond)
			r.Move(mx, my)
			r.Advance(20 * time.Millisecond)
			r.Release(mx, my)
			if selected != 1 || api.EngineSnapshot().PointerDown {
				t.Fatalf("axis drag did not choose logical next %+v", api.Snapshot())
			}
			r.Advance(5 * time.Second)
			// Upstream settles against the previous rendered offset, then interpolates;
			// the final pose can retain a tiny overshoot instead of exact snapping.
			if r.FramePending() || math.Abs(api.EngineSnapshot().OffsetLocation+300) > .01 {
				t.Fatalf("axis drag did not settle: %+v", api.Snapshot())
			}
		})
	}
	api, _ := NewCarouselController(CarouselOptions{})
	selected, count := 0, 1
	r := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() {
			CarouselContainer(c, &selected, CarouselContainerProps{Count: count, Height: 120, Controller: api}, func(c *ui.Context) {
				CarouselContent(c, CarouselContentProps{Spacing: CarouselValue(float32(0))}, func(c *ui.Context) {
					for i := range count {
						CarouselItem(c, i, CarouselItemProps{Key: i}, func(c *ui.Context) { ui.Box(c).Fill() })
					}
				})
			})
		})
	}, 600, 200)
	r.Press(400, 60)
	r.Advance(20 * time.Millisecond)
	r.Move(300, 60)
	if !api.EngineSnapshot().PointerDown || api.EngineSnapshot().Target >= 0 {
		t.Fatal("source one-snap carousel could not drag")
	}
	r.Release(700, 60)
	r.Advance(5 * time.Second)
	if selected != 0 || math.Abs(api.EngineSnapshot().OffsetLocation) > .001 {
		t.Fatal("single-snap bounds did not restore")
	}
	count = 0
	r.Frame()
	r.Press(100, 60)
	r.Move(200, 60)
	r.Release(200, 60)
	if api.EngineSnapshot().PointerDown || len(api.ScrollSnapList()) != 0 {
		t.Fatal("empty inventory captured a drag")
	}
}
