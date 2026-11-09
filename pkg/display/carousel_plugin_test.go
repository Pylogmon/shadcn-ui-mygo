package display

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func carouselPluginTester(api *CarouselController, plugins []CarouselPlugin, count *int, extra func(*ui.Context)) *ui.Tester {
	selected := 0
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(40).Children(func() {
			CarouselContainer(c, &selected, CarouselContainerProps{Count: *count, Height: 120, Align: CarouselAlignStart, Controller: api, Plugins: plugins}, func(c *ui.Context) {
				CarouselContent(c, CarouselContentProps{Spacing: CarouselValue(float32(0))}, func(c *ui.Context) {
					for i := range *count {
						CarouselItem(c, i, CarouselItemProps{Key: i}, func(c *ui.Context) { ui.Textf(c, "Card%d", i) })
					}
				})
			})
			if extra != nil {
				extra(c)
			}
		})
	}, 680, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func TestCarouselPluginsOfficialLifecycleAndMap(t *testing.T) {
	var source struct {
		Lifecycle []struct {
			Name string
			Map  []struct {
				Name        string
				SourceIndex int
			}
			Trace []string
		}
	}
	data, err := os.ReadFile("testdata/shadcn-carousel/autoplay/plugins-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	if len(source.Lifecycle) != 4 {
		t.Fatal("missing source lifecycle fixtures")
	}
	for _, tc := range source.Lifecycle {
		t.Run(tc.Name, func(t *testing.T) {
			var trace []string
			makePlugin := func(id, name string, active *bool) *CarouselCustomPlugin {
				p, err := NewCarouselPlugin(name, CarouselPluginOptions{Active: active}, CarouselPluginHooks{Init: func(CarouselPluginFrame) { trace = append(trace, "init:"+id) }, Destroy: func(CarouselPluginFrame) { trace = append(trace, "destroy:"+id) }})
				if err != nil {
					t.Fatal(err)
				}
				return p
			}
			a, b, c, d := makePlugin("A", "alpha", CarouselValue(true)), makePlugin("B", "beta", CarouselValue(false)), makePlugin("C", "gamma", nil), makePlugin("D", "alpha", CarouselValue(true))
			list := []CarouselPlugin{a, b, c}
			switch tc.Name {
			case "duplicate-name":
				list = []CarouselPlugin{a, d, c}
			case "duplicate-identity":
				list = []CarouselPlugin{a, a}
			case "all-inactive":
				list = []CarouselPlugin{b}
			}
			api, _ := NewCarouselController(CarouselOptions{})
			api.SetPlugins(list)
			count := 3
			r := carouselPluginTester(api, nil, &count, nil)
			registry := api.Plugins()
			if len(registry) != len(tc.Map) {
				t.Fatalf("map=%v source=%v", registry, tc.Map)
			}
			for _, entry := range tc.Map {
				if registry[entry.Name] != list[entry.SourceIndex] {
					t.Fatal("source last-name identity lost")
				}
			}
			api.SetPlugins([]CarouselPlugin{})
			r.Frame()
			api.SetPlugins([]CarouselPlugin{})
			r.Frame()
			if tc.Name == "active-inactive-replacement" {
				b = makePlugin("B", "beta", CarouselValue(true))
				api.SetPlugins([]CarouselPlugin{b, d})
				r.Frame()
				api.SetPlugins([]CarouselPlugin{})
				r.Frame()
			}
			if !slices.Equal(trace, tc.Trace) {
				t.Fatalf("native lifecycle %v official %v", trace, tc.Trace)
			}
		})
	}
}
func TestCarouselAutoplayOfficialNestedListenerOrder(t *testing.T) {
	type record struct {
		Kind, Event string
		Playing     bool
		Remaining   *float64
	}
	var source struct {
		Ordering []struct {
			Kind, UserRegistration string
			Trace                  []record
		}
	}
	data, err := os.ReadFile("testdata/shadcn-carousel/autoplay/plugins-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	if len(source.Ordering) != 4 {
		t.Fatal("missing source nested ordering fixtures")
	}
	for _, tc := range source.Ordering {
		t.Run(tc.Kind+"/"+tc.UserRegistration, func(t *testing.T) {
			auto, _ := NewCarouselAutoplay(CarouselAutoplayOptions{})
			api, _ := NewCarouselController(CarouselOptions{})
			var trace []record
			var now time.Time
			sample := func(kind string, event CarouselEventType) {
				var remaining *float64
				if d, ok := auto.TimeUntilNext(now); ok {
					remaining = CarouselValue(float64(d) / float64(time.Millisecond))
				}
				trace = append(trace, record{kind, string(event), auto.IsPlaying(), remaining})
			}
			user := NewCarouselListener(func(_ *CarouselController, e CarouselEventType) { sample("user", e) })
			if tc.UserRegistration == "before" {
				api.On(CarouselEventType(tc.Kind), user)
			}
			for _, event := range []CarouselEventType{CarouselAutoplayStop, CarouselAutoplayTimerStopped} {
				api.On(event, NewCarouselListener(func(_ *CarouselController, e CarouselEventType) { sample("emit", e) }))
			}
			api.SetPlugins([]CarouselPlugin{auto})
			count := 3
			r := carouselPluginTester(api, nil, &count, nil)
			now = r.Now()
			if tc.UserRegistration == "after" {
				api.On(CarouselEventType(tc.Kind), user)
			}
			trace = nil
			sample("emit", CarouselEventType(tc.Kind))
			api.Emit(CarouselEventType(tc.Kind))
			if !reflect.DeepEqual(trace, tc.Trace) {
				t.Fatalf("nested native=%+v source=%+v", trace, tc.Trace)
			}
		})
	}
}
func TestCarouselAutoplayNativeTimerCommandsAndSelectedDelay(t *testing.T) {
	resolver := NewCarouselAutoplayDelayResolver(func(input CarouselAutoplayDelayInput) []time.Duration {
		if len(input.Snaps) != 3 {
			t.Fatalf("resolver saw %v", input.Snaps)
		}
		return []time.Duration{100 * time.Millisecond, 300 * time.Millisecond, 700 * time.Millisecond}
	})
	auto, _ := NewCarouselAutoplay(CarouselAutoplayOptions{DelayResolver: resolver, Jump: CarouselValue(true)})
	api, _ := NewCarouselController(CarouselOptions{})
	api.SetPlugins([]CarouselPlugin{auto})
	count := 3
	r := carouselPluginTester(api, nil, &count, nil)
	if !auto.IsPlaying() {
		t.Fatal("playOnInit not connected")
	}
	r.Advance(50 * time.Millisecond)
	api.ScrollTo(2, true)
	r.Frame()
	if d, ok := auto.TimeUntilNext(r.Now()); !ok || d != 650*time.Millisecond {
		t.Fatalf("current selected delay query %v %v", d, ok)
	}
	var events []string
	for _, event := range []CarouselEventType{CarouselSelect, CarouselAutoplaySelect, CarouselAutoplayTimerSet} {
		api.On(event, NewCarouselListener(func(api *CarouselController, e CarouselEventType) {
			events = append(events, fmt.Sprintf("%s:%d", e, api.SelectedScrollSnap()))
		}))
	}
	r.Advance(49 * time.Millisecond)
	if api.SelectedScrollSnap() != 2 {
		t.Fatal("source deadline fired early")
	}
	r.Advance(time.Millisecond)
	if api.SelectedScrollSnap() != 0 || !slices.Equal(events, []string{"select:0", "autoplay:select:0", "autoplay:timerset:0"}) {
		t.Fatalf("timer action was deferred past source plugin event: %v selected%d", events, api.SelectedScrollSnap())
	}
	auto.Stop()
	r.Frame()
	if auto.IsPlaying() {
		t.Fatal("stop did not cancel state")
	}
	r.Advance(time.Second)
	if api.SelectedScrollSnap() != 0 {
		t.Fatal("stopped timer navigated")
	}
	auto.Play(false)
	r.Frame()
	r.Advance(50 * time.Millisecond)
	auto.Reset()
	r.Frame()
	r.Advance(99 * time.Millisecond)
	if api.SelectedScrollSnap() != 0 {
		t.Fatal("reset retained old deadline")
	}
	r.Advance(time.Millisecond)
	if api.SelectedScrollSnap() != 1 || api.EngineSnapshot().Options.Duration != 25 {
		t.Fatal("play/reset did not run controller action")
	}
	count = 1
	r.Frame()
	if auto.IsPlaying() {
		t.Fatal("reInit one snap retained old timer")
	}
	auto.Play()
	r.Frame()
	if auto.IsPlaying() {
		t.Fatal("one snap manually revived invalid timer")
	}
	count = 3
	r.Frame()
	if !auto.IsPlaying() {
		t.Fatal("new lifecycle did not initialize default playback")
	}
	api.SetPlugins([]CarouselPlugin{})
	r.Frame()
	if auto.IsPlaying() || len(api.Plugins()) != 0 {
		t.Fatal("clearing plugins retained playback")
	}
}
func TestCarouselCustomPluginFrameActionsExpireAndMedia(t *testing.T) {
	var saved CarouselPluginActions
	inits, destroys, frames := 0, 0, 0
	plugin, err := NewCarouselPlugin("application", CarouselPluginOptions{Breakpoints: []CarouselPluginBreakpoint{{Key: "small", When: CarouselMedia{MaxWidth: CarouselValue(float32(600))}, Active: CarouselValue(false)}}}, CarouselPluginHooks{Init: func(f CarouselPluginFrame) {
		inits++
		saved = f.Actions
		if !f.Actions.ScrollTo(1, true) {
			t.Fatal("missing synchronous action")
		}
		f.Actions.After(100 * time.Millisecond)
	}, Destroy: func(CarouselPluginFrame) { destroys++ }, Frame: func(CarouselPluginFrame) { frames++ }})
	if err != nil {
		t.Fatal(err)
	}
	api, _ := NewCarouselController(CarouselOptions{})
	api.SetPlugins([]CarouselPlugin{plugin})
	count := 3
	r := carouselPluginTester(api, nil, &count, nil)
	if inits != 1 || api.SelectedScrollSnap() != 1 || saved.ScrollNext(true) || saved.Emit("expired") {
		t.Fatal("frame facade escaped or initialization action was queued")
	}
	before := frames
	r.Advance(100 * time.Millisecond)
	if frames <= before {
		t.Fatal("plugin requested deadline was not scheduled")
	}
	r.SetSize(580, 300)
	if destroys != 1 || len(api.Plugins()) != 1 {
		t.Fatal("inactive media plugin map/lifecycle lost")
	}
	r.SetSize(680, 300)
	if inits != 2 {
		t.Fatal("matching media did not reinitialize plugin")
	}
	api.ReInitWithPlugins(CarouselOptions{StartIndex: CarouselValue(0)}, []CarouselPlugin{})
	r.Frame()
	if destroys != 2 || api.SelectedScrollSnap() != 0 {
		t.Fatal("combined options/plugins reInit lost explicit state")
	}
}

func TestCarouselAutoplayNativeInteractionRootCarryAndVisibility(t *testing.T) {
	var root ui.Handle
	auto, _ := NewCarouselAutoplay(CarouselAutoplayOptions{Delay: CarouselValue(100 * time.Millisecond), Jump: CarouselValue(true), StopOnInteraction: CarouselValue(false), StopOnMouseEnter: CarouselValue(true), RootResolver: NewCarouselAutoplayRootResolver(func(ui.Handle) ui.Handle { return root })})
	api, _ := NewCarouselController(CarouselOptions{})
	api.SetPlugins([]CarouselPlugin{auto})
	selected := 0
	values := []string{"zero", "one", "two"}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(40).Bind(&root).Children(func() {
			CarouselContainer(c, &selected, CarouselContainerProps{Count: 3, Height: 120, Controller: api, Align: CarouselAlignStart}, func(c *ui.Context) {
				CarouselContent(c, CarouselContentProps{Spacing: CarouselValue(float32(0))}, func(c *ui.Context) {
					for i := range 3 {
						CarouselItem(c, i, CarouselItemProps{Key: i}, func(c *ui.Context) { ui.TextInput(c, &values[i]).Label(fmt.Sprintf("Autoplay field%d", i)) })
					}
				})
			})
			ui.ButtonBase(c).Label("Outside autoplay").Size(180, 40).Children(func() { ui.Text(c, "Outside") })
		})
	}, 680, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Move(500, 220)
	if auto.IsPlaying() {
		t.Fatal("explicit root hover did not stop")
	}
	auto.Play()
	r.Frame()
	api.ReInit(CarouselOptions{})
	r.Frame()
	if !auto.IsPlaying() {
		t.Fatal("reInit synthesized mouseenter instead of preserving source carry")
	}
	// Source retains mouseIsOver across destroy/init; a pointerDown while still
	// over that root must not stop the just-restarted timer.
	api.Emit(CarouselPointerDown)
	if !auto.IsPlaying() {
		t.Fatal("source hover carry was cleared")
	}
	r.Move(-10, 220)
	r.SetWindowHidden(true)
	before := selected
	r.Advance(time.Second)
	if auto.IsPlaying() || selected != before {
		t.Fatal("hidden autoplay kept running")
	}
	r.SetWindowHidden(false)
	r.Advance(99 * time.Millisecond)
	if selected != before {
		t.Fatal("visibility resume retained an expired deadline")
	}
	r.Advance(time.Millisecond)
	if selected == before {
		t.Fatal("visible full-delay timer did not resume")
	}
	// Recent native Tab focus uses the controller's source slideFocusStart event.
	auto.Play()
	r.Frame()
	r.Click("Outside autoplay")
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyTab)
	if !r.Focused(fmt.Sprintf("Autoplay field%d", selected)) || auto.IsPlaying() {
		t.Fatalf("native Tab did not stop playback, selected%d playing%v", selected, auto.IsPlaying())
	}
	r.Key(0, ui.KeyTab)
	if !auto.IsPlaying() {
		t.Fatal("native track focusout did not resume")
	}
	api.SetPlugins([]CarouselPlugin{})
	r.Frame()
	if auto.IsPlaying() {
		t.Fatal("removal retained timer")
	}
}

func TestCarouselAutoplayOwnerGenerationAndRepeatedIdentity(t *testing.T) {
	auto, _ := NewCarouselAutoplay(CarouselAutoplayOptions{Delay: CarouselValue(time.Second)})
	api, _ := NewCarouselController(CarouselOptions{})
	api.SetPlugins([]CarouselPlugin{auto, auto})
	count := 3
	var events []CarouselEventType
	listener := NewCarouselListener(func(_ *CarouselController, e CarouselEventType) { events = append(events, e) })
	api.On(CarouselAutoplayPlay, listener).On(CarouselAutoplayTimerSet, listener)
	r := carouselPluginTester(api, nil, &count, nil)
	if !slices.Equal(events, []CarouselEventType{CarouselAutoplayPlay, CarouselAutoplayTimerSet, CarouselAutoplayTimerSet}) {
		t.Fatalf("same identity init did not share source state: %v", events)
	}
	other, _ := NewCarouselController(CarouselOptions{})
	other.SetPlugins([]CarouselPlugin{auto})
	second := carouselPluginTester(other, nil, &count, nil)
	if other.Snapshot().Error == nil || !auto.IsPlaying() {
		t.Fatal("cross-owner plugin replaced its accepted owner")
	}
	other.Destroy()
	second.Frame()
	auto.Stop()
	r.Frame()
	if auto.IsPlaying() {
		t.Fatal("failed owner teardown cleared the accepted binding")
	}
	// Repeated pointer listeners also share one source state: stop emits once,
	// timerstopped remains safe to repeat; explicit removal destroys no new owner.
	auto.Play()
	r.Frame()
	api.SetPlugins([]CarouselPlugin{})
	r.Frame()
	if auto.IsPlaying() {
		t.Fatal("duplicate destroy leaked playback")
	}
	api.SetPlugins([]CarouselPlugin{auto})
	r.Frame()
	if !auto.IsPlaying() {
		t.Fatal("plugin object did not reinitialize after removal")
	}
	api.Destroy()
	if auto.IsPlaying() {
		t.Fatal("controller Destroy did not destroy plugins")
	}
	moving, _ := NewCarouselController(CarouselOptions{})
	moving.SetPlugins([]CarouselPlugin{auto})
	showA, showB := true, false
	selected := 0
	destroyed := 0
	moving.On(CarouselDestroy, NewCarouselListener(func(*CarouselController, CarouselEventType) { destroyed++ }))
	move := ui.NewTesterWith(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() {
			build := func(key string) {
				CarouselContainer(c, &selected, CarouselContainerProps{Key: key, Count: 3, Height: 80, Controller: moving}, func(c *ui.Context) {
					CarouselContent(c, CarouselContentProps{}, func(c *ui.Context) {
						for i := range 3 {
							CarouselItem(c, i, CarouselItemProps{Key: i}, func(c *ui.Context) { ui.Textf(c, "item%d", i) })
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
	}, 600, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	generation := moving.Snapshot().Generation
	showA, showB = false, true
	move.Frame()
	if destroyed != 0 || moving.Snapshot().Generation <= generation || !auto.IsPlaying() {
		t.Fatal("old generation destroyed moved plugin binding")
	}
	showB = false
	move.Frame()
	if destroyed != 1 || auto.IsPlaying() {
		t.Fatal("final owner removal did not destroy plugin once")
	}
}
