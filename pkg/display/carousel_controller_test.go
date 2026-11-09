package display

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestCarouselControllerListenerIdentityAndSourceEmitSnapshot(t *testing.T) {
	api, err := NewCarouselController(CarouselOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var trace []string
	b := NewCarouselListener(func(*CarouselController, CarouselEventType) { trace = append(trace, "B") })
	c := NewCarouselListener(func(*CarouselController, CarouselEventType) { trace = append(trace, "C") })
	a := NewCarouselListener(func(api *CarouselController, _ CarouselEventType) {
		trace = append(trace, "A")
		api.Off(CarouselSelect, b).On(CarouselSelect, c)
	})
	api.On(CarouselSelect, a).On(CarouselSelect, b).On(CarouselSelect, b).Emit(CarouselSelect)
	trace = append(trace, "|")
	api.Emit(CarouselSelect)
	if !slices.Equal(trace, []string{"A", "B", "B", "|", "A", "C"}) {
		t.Fatalf("official EventHandler trace mismatch: %v", trace)
	}
	api.On(CarouselDestroy, NewCarouselListener(func(api *CarouselController, _ CarouselEventType) { api.On(CarouselSelect, b) }))
	api.Destroy()
	trace = nil
	api.Emit(CarouselSelect)
	if len(trace) != 0 {
		t.Fatal("destroy callback's new subscription escaped source clear")
	}
	api.On(CarouselSelect, b).Emit(CarouselSelect)
	if !slices.Equal(trace, []string{"B"}) {
		t.Fatal("explicit On/Emit after Destroy was incorrectly disabled")
	}
	if err := api.ReInit(CarouselOptions{}); err == nil {
		t.Fatal("destroyed engine silently reactivated")
	}
}

func TestCarouselControllerPartialOptionsValidateAtomically(t *testing.T) {
	api, err := NewCarouselController(CarouselOptions{Loop: CarouselValue(true), Active: CarouselValue(true)})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.Configure(CarouselOptions{Loop: CarouselValue(false), StartIndex: CarouselValue(0)}); err != nil {
		t.Fatal(err)
	}
	r := carouselResolveOptions(api.baseOptions(), CarouselEnvironment{})
	if r.Loop || !r.Active || r.StartIndex != 0 || r.SlidesToScroll.Count != 1 {
		t.Fatalf("partial false/zero/default lost: %+v", r)
	}
	err = api.Configure(CarouselOptions{Loop: CarouselValue(true), Duration: CarouselValue(25.0)})
	var problem *CarouselOptionError
	if !errors.As(err, &problem) || problem.Code != CarouselRequiresPhysics || problem.Field != "Duration" {
		t.Fatalf("pending source physics was silently accepted: %v", err)
	}
	if carouselResolveOptions(api.baseOptions(), CarouselEnvironment{}).Loop {
		t.Fatal("invalid patch partly changed configuration")
	}
	patch := CarouselOptions{Breakpoints: []CarouselBreakpoint{
		{Key: "medium", When: CarouselMedia{MinWidth: CarouselValue(float32(400))}, Options: CarouselOptions{Loop: CarouselValue(true)}},
		{Key: "large", When: CarouselMedia{MinWidth: CarouselValue(float32(600))}, Options: CarouselOptions{Loop: CarouselValue(false)}},
	}}
	if err := api.Configure(patch); err != nil {
		t.Fatal(err)
	}
	if !carouselResolveOptions(api.baseOptions(), CarouselEnvironment{Width: 500}).Loop || carouselResolveOptions(api.baseOptions(), CarouselEnvironment{Width: 700}).Loop {
		t.Fatal("ordered matching media patches lost source precedence")
	}
}

func carouselControllerTester(api *CarouselController, selected *int, props *CarouselContainerProps, reduction ...bool) *ui.Tester {
	reduced := true
	if len(reduction) > 0 {
		reduced = reduction[0]
	}
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(64).Children(func() {
			CarouselContainer(c, selected, *props, func(c *ui.Context) {
				CarouselContent(c, CarouselContentProps{}, func(c *ui.Context) {
					for i := range props.Count {
						CarouselItem(c, i, CarouselItemProps{Key: fmt.Sprintf("item%d", i), Basis: CarouselFraction(.5)}, func(c *ui.Context) { carouselColorSlide(c, i) })
					}
				})
				CarouselPrevious(c, CarouselControlProps{})
				CarouselNext(c, CarouselControlProps{})
			})
		})
	}, 728, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
}

func TestCarouselControllerNativeJumpMatchesOfficialEngineTrace(t *testing.T) {
	api, err := NewCarouselController(CarouselOptions{})
	if err != nil {
		t.Fatal(err)
	}
	selected := 0
	p := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, Controller: api}
	r := carouselControllerTester(api, &selected, &p)
	if !api.Snapshot().Ready {
		t.Fatal("controller did not bind the resolved native layout")
	}
	if got := api.ScrollSnapList(); len(got) != 4 || got[0] != 0 || got[3] != 1 {
		t.Fatalf("normalized source snap list=%v", got)
	}
	var trace []string
	listener := NewCarouselListener(func(api *CarouselController, event CarouselEventType) {
		trace = append(trace, fmt.Sprintf("%s:%d:%d:%.3f", event, api.SelectedScrollSnap(), api.PreviousScrollSnap(), api.ScrollProgress()))
	})
	api.On(CarouselScroll, listener).On(CarouselSelect, listener).On(CarouselSettle, listener)
	api.ScrollNext(true)
	if api.SelectedScrollSnap() != 0 {
		t.Fatal("queued native command changed committed state before its owner frame")
	}
	r.Frame()
	if !slices.Equal(trace, []string{"scroll:0:0:0.333", "select:1:0:0.333"}) {
		t.Fatalf("official jump trace mismatch: %v", trace)
	}
	if selected != 1 || r.FramePending() {
		t.Fatalf("jump did not commit/stop: selected=%d pending=%v", selected, r.FramePending())
	}
	if api.RootHandle() == (ui.Handle{}) || api.ContainerHandle() == (ui.Handle{}) || api.RegionHandle() == api.RootHandle() || len(api.SlideHandles()) != 5 {
		t.Fatal("source viewport/track/item handle mapping missing")
	}
}

func TestCarouselControllerReInitFIFOAndCurrentErrorRecovery(t *testing.T) {
	api, _ := NewCarouselController(CarouselOptions{})
	selected := 0
	valid := CarouselOptions{Loop: CarouselValue(false)}
	p := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, Controller: api, Options: &valid}
	r := carouselControllerTester(api, &selected, &p)
	var states []int
	listener := NewCarouselListener(func(api *CarouselController, _ CarouselEventType) { states = append(states, api.SelectedScrollSnap()) })
	api.On(CarouselReInit, listener)
	api.ReInit(CarouselOptions{StartIndex: CarouselValue(1)})
	api.ReInit(CarouselOptions{StartIndex: CarouselValue(2)})
	r.Frame()
	if !slices.Equal(states, []int{1, 2}) || api.PreviousScrollSnap() != 2 {
		t.Fatalf("queued configuration read the final desired state early: %v state=%+v", states, api.Snapshot())
	}
	states = nil
	api.Configure(CarouselOptions{Loop: CarouselValue(true)})
	api.ScrollTo(4, true)
	api.Configure(CarouselOptions{Loop: CarouselValue(false)})
	var selections []int
	api.On(CarouselSelect, NewCarouselListener(func(api *CarouselController, _ CarouselEventType) {
		selections = append(selections, api.SelectedScrollSnap())
	}))
	r.Frame()
	if !slices.Equal(selections, []int{4}) || !slices.Equal(states, []int{2, 3}) {
		t.Fatalf("configure/scroll/configure lost intermediate source state: select=%v reinit=%v", selections, states)
	}
	states = nil
	invalid := CarouselOptions{Duration: CarouselValue(25.0)}
	p.Options = &invalid
	r.Frame()
	if api.Snapshot().Error == nil {
		t.Fatal("invalid frame options were not reported")
	}
	p.Options = &valid
	r.Frame()
	if api.Snapshot().Error != nil || len(states) != 0 || api.SelectedScrollSnap() != 3 {
		t.Fatalf("restoring old valid props left an error or reinitialized: error=%v states=%v", api.Snapshot().Error, states)
	}
}

func TestCarouselControllerDisabledWatchersKeepEngineAndLiveLayoutSeparate(t *testing.T) {
	api, _ := NewCarouselController(CarouselOptions{WatchResize: CarouselValue(false), WatchSlides: CarouselValue(false)})
	selected := 0
	p := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, Controller: api}
	r := carouselControllerTester(api, &selected, &p)
	before := api.Snapshot()
	events := 0
	api.On(CarouselReInit, NewCarouselListener(func(*CarouselController, CarouselEventType) { events++ }))
	r.SetSize(928, 400)
	after := api.Snapshot()
	viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
	first, _ := r.AccessByRole(ui.RoleGroup, "Slide 1 of 5")
	if !slices.Equal(before.SnapOffsets, after.SnapOffsets) || events != 0 || viewport.Bounds.W != 800 || first.Bounds.W != 408 {
		t.Fatalf("watchResize=false froze native layout or rebuilt engine: before=%v after=%v events=%d viewport=%+v item=%+v", before.SnapOffsets, after.SnapOffsets, events, viewport.Bounds, first.Bounds)
	}
	p.Count = 4
	r.Frame()
	if len(api.SlideHandles()) != 5 || len(api.ScrollSnapList()) != 4 || events != 0 {
		t.Fatal("watchSlides=false silently refreshed the engine's node inventory")
	}
	api.ReInit(CarouselOptions{})
	r.Frame()
	if len(api.SlideHandles()) != 4 || len(api.ScrollSnapList()) != 3 || events != 1 {
		t.Fatalf("manual reInit did not resolve the current native layout: %+v events=%d", api.Snapshot(), events)
	}
}

func TestCarouselControllerOwnerLeaseUnmountAndSameWindowMove(t *testing.T) {
	api, _ := NewCarouselController(CarouselOptions{})
	showA, showB := true, false
	selected := 0
	destroyed := 0
	api.On(CarouselDestroy, NewCarouselListener(func(*CarouselController, CarouselEventType) { destroyed++ }))
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(64).Children(func() {
			build := func(key string) {
				CarouselContainer(c, &selected, CarouselContainerProps{Key: key, Label: key, Count: 3, Height: 80, Controller: api}, func(c *ui.Context) {
					CarouselContent(c, CarouselContentProps{}, func(c *ui.Context) {
						for i := range 3 {
							CarouselItem(c, i, CarouselItemProps{}, func(c *ui.Context) { ui.Textf(c, "item%d", i) })
						}
					})
				})
			}
			if showA {
				build("a")
			}
			if showB {
				build("b")
			}
		})
	}, 600, 400, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	generation := api.Snapshot().Generation
	showB = true
	r.Frame()
	var problem *CarouselOptionError
	if !errors.As(api.Snapshot().Error, &problem) || problem.Code != CarouselAlreadyBound || !api.Snapshot().Ready {
		t.Fatalf("double binding stole the owner or omitted typed error: %+v", api.Snapshot())
	}
	if err := r.Click("b"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyRight)
	if api.SelectedScrollSnap() != 0 {
		t.Fatal("a rejected owner still controlled the accepted engine")
	}
	showB = false
	r.Frame()
	if destroyed != 0 || !api.Snapshot().Ready {
		t.Fatal("failed second owner's cleanup destroyed the live owner")
	}
	showA, showB = false, true
	r.Frame()
	if destroyed != 0 || !api.Snapshot().Ready || api.Snapshot().Generation <= generation {
		t.Fatalf("old lease destroyed a same-window replacement: destroyed=%d state=%+v", destroyed, api.Snapshot())
	}
	showB = false
	r.Frame()
	if destroyed != 1 || !api.Snapshot().Destroyed {
		t.Fatalf("final owner did not destroy exactly once: %d %+v", destroyed, api.Snapshot())
	}
}

func TestCarouselControllerSmoothClockAndLoopInlineRender(t *testing.T) {
	api, _ := NewCarouselController(CarouselOptions{})
	selected := 0
	p := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, Controller: api}
	r := carouselControllerTester(api, &selected, &p, false)
	var trace []CarouselEventType
	listener := NewCarouselListener(func(_ *CarouselController, event CarouselEventType) { trace = append(trace, event) })
	api.On(CarouselSelect, listener).On(CarouselScroll, listener).On(CarouselSettle, listener)
	api.ScrollNext(false)
	r.Frame()
	r.Frame()
	if !slices.Equal(trace, []CarouselEventType{CarouselSelect}) {
		t.Fatalf("a same-clock owner rebuild acted as a source RAF: %v", trace)
	}
	api.ScrollTo(2, true)
	r.Frame()
	if !slices.Equal(trace, []CarouselEventType{CarouselSelect, CarouselScroll, CarouselSelect}) {
		t.Fatalf("inline jump or preexisting RAF emitted at the wrong point: %v", trace)
	}
	r.Advance(17 * time.Millisecond)
	if !slices.Equal(trace, []CarouselEventType{CarouselSelect, CarouselScroll, CarouselSelect, CarouselSettle}) {
		t.Fatalf("the preexisting RAF did not settle after jump: %v", trace)
	}
	loop, _ := NewCarouselController(CarouselOptions{Loop: CarouselValue(true)})
	selected = 0
	p.Controller = loop
	r = carouselControllerTester(loop, &selected, &p, false)
	var scroll, selection CarouselControllerSnapshot
	loop.On(CarouselScroll, NewCarouselListener(func(api *CarouselController, _ CarouselEventType) { scroll = api.Snapshot() })).On(CarouselSelect, NewCarouselListener(func(api *CarouselController, _ CarouselEventType) { selection = api.Snapshot() }))
	loop.ScrollTo(4, true)
	r.Frame()
	if scroll.Selected != 0 || math.Abs(scroll.Progress+.2) > .0001 || scroll.Engine.OffsetLocation != 0 || selection.Selected != 4 || math.Abs(selection.Progress-.8) > .0001 || math.Abs(selection.Engine.Location+selection.Engine.Content*.8) > .001 {
		t.Fatalf("source loop inline-render coordinate wrap mismatch: scroll=%+v select=%+v", scroll, selection)
	}
}

func TestCarouselControllerResizeUsesReInitBaselineAndEveryDeliveryHook(t *testing.T) {
	api, _ := NewCarouselController(CarouselOptions{})
	selected := 0
	width := float32(600)
	spacing := float32(0)
	hooks, resizes := 0, 0
	allow := true
	api.On(CarouselResize, NewCarouselListener(func(*CarouselController, CarouselEventType) { resizes++ }))
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Padding(32).Width(width + 64).Children(func() {
			CarouselContainer(c, &selected, CarouselContainerProps{Count: 5, Height: 80, Align: CarouselAlignStart, Controller: api, Hooks: CarouselWatchHooks{BeforeResize: func(*CarouselController, CarouselResizeEvent) bool { hooks++; return allow }}}, func(c *ui.Context) {
				CarouselContent(c, CarouselContentProps{Spacing: &spacing}, func(c *ui.Context) {
					for i := range 5 {
						CarouselItem(c, i, CarouselItemProps{Basis: CarouselFraction(.5)}, func(c *ui.Context) { carouselColorSlide(c, i) })
					}
				})
			})
		})
	}, 800, 250)
	hooks, resizes = 0, 0
	width = 600.25
	r.Frame()
	if hooks != 1 || resizes != 0 {
		t.Fatalf("subthreshold delivery hook=%d resize=%d", hooks, resizes)
	}
	width = 600.5
	r.Frame()
	if hooks != 2 || resizes != 1 || api.EngineSnapshot().Viewport != 600.5 {
		t.Fatalf("cumulative .5 never reached the observer's init baseline: hook=%d resize=%d viewport=%v", hooks, resizes, api.EngineSnapshot().Viewport)
	}
	allow = false
	width = 600.75
	r.Frame()
	width = 601
	r.Frame()
	allow = true
	width = 601.25
	r.Frame()
	if hooks != 5 || resizes != 2 || api.EngineSnapshot().Viewport != 601.25 {
		t.Fatalf("veto incorrectly reset the observer baseline: hook=%d resize=%d viewport=%v", hooks, resizes, api.EngineSnapshot().Viewport)
	}
}
