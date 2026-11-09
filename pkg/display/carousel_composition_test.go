package display

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func carouselComposedTester(selected *int, props *CarouselContainerProps, spacing *float32, basis func(int) CarouselBasis, snapshot *CarouselSnapshot, content func(*ui.Context, int), reduced bool) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(64).Children(func() {
			CarouselContainer(c, selected, *props, func(c *ui.Context) {
				CarouselContent(c, CarouselContentProps{Spacing: spacing}, func(c *ui.Context) {
					for i := range props.Count {
						CarouselItem(c, i, CarouselItemProps{Key: fmt.Sprintf("item-%d", i), Basis: basis(i)}, func(c *ui.Context) { content(c, i) })
					}
				})
				CarouselPrevious(c, CarouselControlProps{})
				CarouselNext(c, CarouselControlProps{})
				*snapshot = CarouselSnapshotOf(c)
			})
		})
	}, 600, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
}

var carouselTestColors = []ui.Color{ui.RGB(220, 0, 0), ui.RGB(0, 170, 0), ui.RGB(0, 0, 220), ui.RGB(200, 100, 0), ui.RGB(170, 0, 170)}

func carouselColorSlide(c *ui.Context, i int) { ui.Box(c).Fill().Background(carouselTestColors[i]) }

func TestCarouselCompositionMeasuredFractionsSpacingAndSnapSelection(t *testing.T) {
	selected, calls := 4, 0
	props := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, OnSelect: func(int) { calls++ }}
	var snapshot CarouselSnapshot
	r := carouselComposedTester(&selected, &props, nil, func(int) CarouselBasis { return CarouselFraction(.5) }, &snapshot, carouselColorSlide, true)
	if selected != 3 || len(snapshot.Snaps) != 4 || snapshot.SlideCount != 5 || snapshot.CanScrollNext {
		t.Fatalf("selected slide count was used for snaps: selected=%d snapshot=%+v", selected, snapshot)
	}
	if calls != 0 {
		t.Fatal("geometry normalization emitted a user selection")
	}
	if !slices.Equal(snapshot.Visible, []int{3, 4}) {
		t.Fatalf("last two items visible=%v", snapshot.Visible)
	}
	viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
	if snapshot.Snaps[1] != -(viewport.Bounds.W+16)/2 {
		t.Fatalf("basis did not use measured native viewport: snaps=%v viewport=%+v", snapshot.Snaps, viewport.Bounds)
	}
	previous, _ := r.AccessByRole(ui.RoleButton, "Previous slide")
	next, _ := r.AccessByRole(ui.RoleButton, "Next slide")
	if previous.Bounds.X != viewport.Bounds.X-48 || next.Bounds.X != viewport.Bounds.X+viewport.Bounds.W+16 || previous.Bounds.Y+previous.Bounds.H/2 != viewport.Bounds.Y+viewport.Bounds.H/2 {
		t.Fatalf("controls are not independently outside/centered: previous=%+v next=%+v viewport=%+v", previous.Bounds, next.Bounds, viewport.Bounds)
	}
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleStatus {
			t.Fatal("composed primitive added an unrequested status")
		}
	}
	if err := r.Click("Previous slide"); err != nil {
		t.Fatal(err)
	}
	if selected != 2 || calls != 1 {
		t.Fatalf("previous did not select a snap exactly once: selected=%d calls=%d", selected, calls)
	}
	selected = 0
	r.Frame()
	im := r.Image()
	y := int(viewport.Bounds.Y + viewport.Bounds.H/2)
	red, green := color.RGBA{R: 220, A: 255}, color.RGBA{G: 170, A: 255}
	if im.RGBAAt(int(viewport.Bounds.X+30), y) != red || im.RGBAAt(int(viewport.Bounds.X+viewport.Bounds.W-30), y) != green {
		t.Fatal("two native items did not paint across the viewport")
	}
	gap := im.RGBAAt(int(viewport.Bounds.X+(viewport.Bounds.W+16)/2-8), y)
	if gap == red || gap == green {
		t.Fatal("default source spacing was not compensated between item contents")
	}
	r.SetSize(780, 500)
	viewport, _ = r.AccessByRole(ui.RoleGroup, "Slides")
	if snapshot.Snaps[1] != -(viewport.Bounds.W+16)/2 || selected != 0 {
		t.Fatalf("resize left stale geometry: %+v viewport=%+v", snapshot, viewport.Bounds)
	}
}

func TestCarouselCompositionPartialVisibleEditorsRemainNative(t *testing.T) {
	selected := 0
	props := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, Duration: 600 * time.Millisecond}
	values := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	var snapshot CarouselSnapshot
	r := carouselComposedTester(&selected, &props, nil, func(int) CarouselBasis { return CarouselFraction(.5) }, &snapshot, func(c *ui.Context, i int) {
		ui.TextInput(c, &values[i]).Label(fmt.Sprintf("Editor %d", i))
	}, false)
	if _, ok := r.AccessByRole(ui.RoleTextField, "Editor 1"); !ok {
		t.Fatal("second visible editor was hidden")
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Editor 2"); ok {
		t.Fatal("fully offscreen editor was exposed")
	}
	if err := r.Click("Editor 1"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyLeft)
	r.Type("X")
	if selected != 0 || values[1] == "beta" || values[0] != "alpha" {
		t.Fatalf("second item was not independently editable: selected=%d values=%v", selected, values)
	}
	if err := r.Click("Next slide"); err != nil {
		t.Fatal(err)
	}
	r.Advance(100 * time.Millisecond)
	if !slices.Equal(snapshot.Visible, []int{0, 1, 2}) {
		t.Fatalf("partial items were dropped during motion: %v", snapshot.Visible)
	}
	for _, index := range snapshot.Visible {
		if _, ok := r.AccessByRole(ui.RoleTextField, fmt.Sprintf("Editor %d", index)); !ok {
			t.Fatalf("visible editor %d missing from AX", index)
		}
	}
	if _, ok := r.AccessByRole(ui.RoleTextField, "Editor 3"); ok {
		t.Fatal("completely offscreen editor became an AX/input target")
	}
	r.Advance(time.Second)
	if _, ok := r.AccessByRole(ui.RoleTextField, "Editor 0"); ok {
		t.Fatal("departed editor remained exposed after settling")
	}
}

func TestCarouselCompositionVerticalRTLAndChangingBasis(t *testing.T) {
	for _, axis := range []string{"vertical", "rtl"} {
		t.Run(axis, func(t *testing.T) {
			selected := 0
			props := CarouselContainerProps{Count: 5, Height: 270, Align: CarouselAlignStart, Vertical: axis == "vertical"}
			spacing, fraction := float32(4), float32(.5)
			var snapshot CarouselSnapshot
			r := carouselComposedTester(&selected, &props, &spacing, func(int) CarouselBasis { return CarouselFraction(fraction) }, &snapshot, carouselColorSlide, true)
			if axis == "rtl" {
				r.SetDirection(ui.RightToLeft)
			}
			viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
			if len(snapshot.Snaps) != 4 || !slices.Equal(snapshot.Visible, []int{0, 1}) {
				t.Fatalf("two-up source geometry: %+v", snapshot)
			}
			im := r.Image()
			x, y := int(viewport.Bounds.X+viewport.Bounds.W/2), int(viewport.Bounds.Y+30)
			if axis == "rtl" {
				if px := im.RGBAAt(int(viewport.Bounds.X+viewport.Bounds.W-30), int(viewport.Bounds.Y+100)); px != (color.RGBA{R: 220, A: 255}) {
					t.Fatalf("RTL first item did not paint on right: %v", px)
				}
			} else if px := im.RGBAAt(x, y); px != (color.RGBA{R: 220, A: 255}) {
				t.Fatalf("vertical first item did not paint above second: %v", px)
			}
			selected = 3
			r.Frame()
			fraction = 1.0 / 3
			r.Frame()
			if selected != 2 || len(snapshot.Snaps) != 3 || snapshot.CanScrollNext {
				t.Fatalf("three-up reinit did not clamp to snap count: selected=%d snapshot=%+v", selected, snapshot)
			}
		})
	}
}

func TestCarouselCompositionStableKeysAndScopeIsolation(t *testing.T) {
	selected := 0
	ids := []string{"a", "b", "c"}
	values := map[string]*string{}
	created := map[string]int{}
	for _, id := range ids {
		value := id
		values[id] = &value
	}
	var nested, outer CarouselSnapshot
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(64).Children(func() {
			CarouselContainer(c, &selected, CarouselContainerProps{Count: len(ids), Height: 100, Align: CarouselAlignStart}, func(c *ui.Context) {
				CarouselContent(c, CarouselContentProps{}, func(c *ui.Context) {
					for i, id := range ids {
						CarouselItem(c, i, CarouselItemProps{Key: id, Basis: CarouselFraction(.5)}, func(c *ui.Context) {
							box := ui.Column(c)
							ui.Local(box, "created", func() int { created[id]++; return 1 })
							box.Children(func() { ui.TextInput(c, values[id]).Label("Key " + id) })
						})
					}
				})
				innerSelected := 0
				CarouselContainer(c, &innerSelected, CarouselContainerProps{Key: "inner", Count: 0}, func(c *ui.Context) { nested = CarouselSnapshotOf(c) })
				outer = CarouselSnapshotOf(c)
			})
		})
	}, 600, 400)
	if nested.SlideCount != 0 || outer.SlideCount != 3 || len(outer.Snaps) != 2 {
		t.Fatalf("nested scope leaked: outer=%+v nested=%+v", outer, nested)
	}
	if err := r.Click("Key b"); err != nil {
		t.Fatal(err)
	}
	r.Type("X")
	r.Key(ui.Cmd, ui.KeyA)
	ids[0], ids[1] = ids[1], ids[0]
	r.Frame()
	if !r.Focused("Key b") {
		t.Fatal("reordering a still-visible keyed editor lost native focus")
	}
	r.Type("Y")
	if created["a"] != 1 || created["b"] != 1 || created["c"] != 1 || *values["b"] != "Y" || *values["a"] != "a" {
		t.Fatalf("keyed state/input moved with positional item: created=%v a=%q b=%q", created, *values["a"], *values["b"])
	}
	value := "d"
	values["d"] = &value
	ids = append(ids, "d")
	r.Frame()
	if !r.Focused("Key b") || created["b"] != 1 || created["d"] != 1 || len(outer.Snaps) != 3 {
		t.Fatalf("appending an item stole a still-visible editor: focused=%v created=%v snapshot=%+v", r.Focused("Key b"), created, outer)
	}
	// Invalid keys are a safe positional fallback, never a native hash panic.
	keys := []any{[]string{"bad"}, map[string]int{"bad": 1}, struct{ Value any }{[]int{1}}, math.NaN(), "duplicate", "duplicate"}
	keyCreated := make([]int, len(keys))
	r = ui.NewTester(func(c *ui.Context) {
		CarouselContainer(c, &selected, CarouselContainerProps{Count: len(keys), Height: 80}, func(c *ui.Context) {
			CarouselContent(c, CarouselContentProps{}, func(c *ui.Context) {
				for i, key := range keys {
					CarouselItem(c, i, CarouselItemProps{Key: key}, func(c *ui.Context) {
						ui.Local(ui.Box(c), "instance", func() int { keyCreated[i]++; return 1 })
					})
				}
			})
		})
	}, 300, 200)
	r.Frame()
	for i, created := range keyCreated {
		if created != 1 {
			t.Fatalf("invalid/duplicate key %d did not retain a distinct positional item: %v", i, keyCreated)
		}
	}
}

func TestCarouselCompositionInstantAndLayoutChangesAtSameClock(t *testing.T) {
	selected := 0
	fraction := float32(.5)
	props := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart, Duration: -1}
	var snapshot CarouselSnapshot
	r := carouselComposedTester(&selected, &props, nil, func(int) CarouselBasis { return CarouselFraction(fraction) }, &snapshot, carouselColorSlide, false)
	selected = 3
	r.Frame() // No Advance: negative duration must not become a 1ms tween.
	if !slices.Equal(snapshot.Visible, []int{3, 4}) || r.FramePending() {
		t.Fatalf("instant selection retained old pose: %+v pending=%v", snapshot, r.FramePending())
	}
	fraction = 1.0 / 3
	r.Frame()
	if selected != 2 || len(snapshot.Snaps) != 3 || !slices.Equal(snapshot.Visible, []int{2, 3, 4}) || r.FramePending() {
		t.Fatalf("same-clock reconfiguration retained old pose: selected=%d %+v pending=%v", selected, snapshot, r.FramePending())
	}
	viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
	if px := r.Image().RGBAAt(int(viewport.Bounds.X+30), int(viewport.Bounds.Y+60)); px != (color.RGBA{B: 220, A: 255}) {
		t.Fatalf("instant geometry snapshot and actual pixels diverged: %v", px)
	}
}

func TestCarouselCompositionLoopOptionReinitializesVirtualSnap(t *testing.T) {
	selected, spacing := 0, float32(0)
	props := CarouselContainerProps{Count: 5, Height: 100, Align: CarouselAlignStart, Loop: true, Duration: -1}
	var snapshot CarouselSnapshot
	r := carouselComposedTester(&selected, &props, &spacing, func(int) CarouselBasis { return CarouselFraction(.5) }, &snapshot, carouselColorSlide, false)
	if err := r.Click("Previous slide"); err != nil {
		t.Fatal(err)
	}
	if selected != 4 || !snapshot.Loop {
		t.Fatalf("loop did not wrap by snap: selected=%d %+v", selected, snapshot)
	}
	if err := r.Click("Next slide"); err != nil {
		t.Fatal(err)
	}
	props.Loop = false
	r.Frame()
	if selected != 0 || snapshot.Loop || !slices.Equal(snapshot.Visible, []int{0, 1}) || snapshot.CanScrollPrevious || len(snapshot.Snaps) != 4 {
		t.Fatalf("loop reinit retained an unrelated virtual occurrence: selected=%d %+v", selected, snapshot)
	}
}

func TestCarouselCompositionMixedLoopRapidRetargetKeepsViewportPainted(t *testing.T) {
	for _, axis := range []string{"horizontal", "rtl", "vertical"} {
		t.Run(axis, func(t *testing.T) {
			selected, spacing := 0, float32(0)
			props := CarouselContainerProps{Count: 4, Height: 210, Vertical: axis == "vertical", Loop: true, Duration: 600 * time.Millisecond}
			var snapshot CarouselSnapshot
			sizes := []float32{180, 260, 210, 280}
			r := carouselComposedTester(&selected, &props, &spacing, func(i int) CarouselBasis { return CarouselFixed(sizes[i]) }, &snapshot, carouselColorSlide, false)
			if axis == "rtl" {
				r.SetDirection(ui.RightToLeft)
			}
			if !snapshot.Loop {
				t.Fatalf("valid mixed geometry did not loop: %+v", snapshot)
			}
			viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
			pixels := func() []color.RGBA {
				im := r.Image()
				var row []color.RGBA
				length := int(viewport.Bounds.W)
				if props.Vertical {
					length = int(viewport.Bounds.H)
				}
				for at := 2; at < length-2; at++ {
					x, y := int(viewport.Bounds.X)+at, int(viewport.Bounds.Y+viewport.Bounds.H/2)
					if props.Vertical {
						x, y = int(viewport.Bounds.X+viewport.Bounds.W/2), int(viewport.Bounds.Y)+at
					}
					row = append(row, im.RGBAAt(x, y))
				}
				return row
			}
			painted := func() {
				t.Helper()
				missing := 0
				for _, px := range pixels() {
					if px != (color.RGBA{R: 220, A: 255}) && px != (color.RGBA{G: 170, A: 255}) && px != (color.RGBA{B: 220, A: 255}) && px != (color.RGBA{R: 200, G: 100, A: 255}) {
						missing++
					}
				}
				// Each independently painted fractional boundary can antialias
				// one pixel. Mixed multi-item views contain more than one.
				if missing > max(2, len(snapshot.Visible)) {
					t.Fatalf("zero-gap loop lost %d pixels", missing)
				}
			}
			painted()
			for _, action := range []string{"Previous slide", "Next slide", "Next slide", "Next slide", "Previous slide", "Next slide"} {
				before := pixels()
				if err := r.Click(action); err != nil {
					t.Fatal(err)
				}
				changed := 0
				for i, pixel := range pixels() {
					if pixel != before[i] {
						changed++
					}
				}
				if changed > max(2, len(snapshot.Visible)) {
					t.Fatalf("%s jumped %d painted pixels at the same clock", action, changed)
				}
				painted()
				for range 4 {
					r.Advance(25 * time.Millisecond)
					painted()
				}
			}
			r.Advance(time.Second)
			painted()
			if r.FramePending() {
				t.Fatal("settled composed loop kept requesting frames")
			}
		})
	}
}

func TestCarouselCompositionCancellationHiddenAndReducedMotion(t *testing.T) {
	selected, spacing := 0, float32(0)
	props := CarouselContainerProps{Count: 5, Height: 120, Align: CarouselAlignStart}
	var snapshot CarouselSnapshot
	r := carouselComposedTester(&selected, &props, &spacing, func(int) CarouselBasis { return CarouselFraction(.5) }, &snapshot, carouselColorSlide, false)
	viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
	x, y := viewport.Bounds.X+viewport.Bounds.W*.7, viewport.Bounds.Y+60
	r.Press(x, y)
	r.Move(x-100, y)
	r.CancelPress()
	if selected != 0 {
		t.Fatal("cancelled multi-item drag selected a snap")
	}
	r.Advance(time.Second)
	if err := r.Click("Next slide"); err != nil {
		t.Fatal(err)
	}
	r.SetWindowHidden(true)
	r.Advance(time.Hour)
	if selected != 1 || r.FramePending() {
		t.Fatalf("hidden native carousel did not suspend: selected=%d pending=%v", selected, r.FramePending())
	}
	r.SetWindowHidden(false)
	r.Advance(time.Second)
	if !slices.Equal(snapshot.Visible, []int{1, 2}) || r.FramePending() {
		t.Fatalf("restored native motion did not settle: %+v pending=%v", snapshot, r.FramePending())
	}
	r.SetPreferences(ui.Preferences{ReduceMotion: true})
	if err := r.Click("Next slide"); err != nil {
		t.Fatal(err)
	}
	if selected != 2 || !slices.Equal(snapshot.Visible, []int{2, 3}) || r.FramePending() {
		t.Fatalf("reduced motion did not settle: selected=%d %+v pending=%v", selected, snapshot, r.FramePending())
	}
}

func TestCarouselCompositionSnapshotsBeforeContentSampleResolvedMotion(t *testing.T) {
	for _, mode := range []string{"horizontal", "rtl", "vertical"} {
		t.Run(mode, func(t *testing.T) {
			selected, count := 2, 5
			fraction := float32(.5)
			var before, after CarouselSnapshot
			r := ui.NewTester(func(c *ui.Context) {
				theme.Apply(c)
				ui.Column(c).Fill().Padding(64).Children(func() {
					CarouselContainer(c, &selected, CarouselContainerProps{Count: count, Height: 150, Vertical: mode == "vertical", Align: CarouselAlignStart, Duration: 600 * time.Millisecond}, func(c *ui.Context) {
						before = CarouselSnapshotOf(c)
						CarouselPrevious(c, CarouselControlProps{})
						CarouselNext(c, CarouselControlProps{})
						CarouselContent(c, CarouselContentProps{}, func(c *ui.Context) {
							for i := range count {
								CarouselItem(c, i, CarouselItemProps{Basis: CarouselFraction(fraction)}, func(c *ui.Context) { carouselColorSlide(c, i) })
							}
						})
						after = CarouselSnapshotOf(c)
					})
				})
			}, 600, 500)
			if mode == "rtl" {
				r.SetDirection(ui.RightToLeft)
			}
			consistent := func(stage string) {
				t.Helper()
				if before.SelectedSnap != after.SelectedSnap || !slices.Equal(before.Snaps, after.Snaps) || !slices.Equal(before.Visible, after.Visible) {
					t.Fatalf("%s: before=%+v after=%+v", stage, before, after)
				}
				for i := range count {
					_, exists := r.AccessByRole(ui.RoleGroup, fmt.Sprintf("Slide %d of %d", i+1, count))
					if exists != slices.Contains(before.Visible, i) {
						t.Fatalf("%s: before-content visible=%v differs from actual AX item %d=%v", stage, before.Visible, i, exists)
					}
				}
			}
			consistent("initial selected snap two")
			if !slices.Equal(before.Visible, []int{2, 3}) {
				t.Fatalf("pre-content snapshot incorrectly used pose zero: %+v", before)
			}
			if err := r.Click("Next slide"); err != nil {
				t.Fatal(err)
			}
			for _, elapsed := range []time.Duration{20 * time.Millisecond, 30 * time.Millisecond} {
				r.Advance(elapsed)
				consistent(fmt.Sprintf("motion %v", elapsed))
			}
			fraction = 1.0 / 3
			r.Frame()
			consistent("new basis")
			count = 2
			r.Frame()
			consistent("new count")
			if selected != 0 || len(before.Snaps) != 1 || !slices.Equal(before.Visible, []int{0, 1}) {
				t.Fatalf("reconfigured pre-content snapshot did not converge: selected=%d %+v", selected, before)
			}
			next, _ := r.AccessByRole(ui.RoleButton, "Next slide")
			if !next.Disabled {
				t.Fatal("a control preceding Content did not converge at a one-snap boundary")
			}
		})
	}
}

func TestCarouselCompositionTinyBasisUsesActualPaddingFloor(t *testing.T) {
	for _, mode := range []string{"horizontal", "rtl", "vertical"} {
		for _, basis := range []CarouselBasis{CarouselFixed(4), CarouselFraction(.01)} {
			t.Run(fmt.Sprintf("%s/%v", mode, basis), func(t *testing.T) {
				selected := 0
				var snapshot CarouselSnapshot
				var outer [3]ui.Rect
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					ui.Column(c).Fill().Padding(32).Children(func() {
						CarouselContainer(c, &selected, CarouselContainerProps{Count: 3, Height: 120, Vertical: mode == "vertical", Align: CarouselAlignStart}, func(c *ui.Context) {
							CarouselContent(c, CarouselContentProps{}, func(c *ui.Context) {
								for i := range 3 {
									CarouselItem(c, i, CarouselItemProps{Basis: basis}, nil).DrawOver(func(_ *ui.Painter, bounds ui.Rect) { outer[i] = bounds })
								}
							})
							snapshot = CarouselSnapshotOf(c)
						})
					})
				}, 664, 300)
				if mode == "rtl" {
					r.SetDirection(ui.RightToLeft)
				}
				if !slices.Equal(snapshot.Visible, []int{1, 2}) || len(snapshot.Snaps) != 1 {
					t.Fatalf("geometry did not use the native/CSS padding minimum: %+v", snapshot)
				}
				for i := range 3 {
					_, exists := r.AccessByRole(ui.RoleGroup, fmt.Sprintf("Slide %d of 3", i+1))
					if exists != (i > 0) {
						t.Fatalf("tiny item's actual visibility and AX differ: item%d=%v bounds=%+v", i, exists, outer)
					}
					if i > 0 {
						size := outer[i].W
						if mode == "vertical" {
							size = outer[i].H
						}
						if size != 16 {
							t.Fatalf("native outer used size=%v, want padding floor16", size)
						}
					}
				}
			})
		}
	}
}
