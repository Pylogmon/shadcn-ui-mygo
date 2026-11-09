package display

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func carouselTester(selected *int, p *CarouselProps, onChange func(), slide func(*ui.Context, int), reduce bool) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Fill().Padding(16).Children(func() {
			if CarouselWith(c, selected, *p, func(i int) { slide(c, i) }).Changed() && onChange != nil {
				onChange()
			}
		})
	}, 400, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduce}})
}
func TestCarouselKeyboardLoopDotsAndSlideSemantics(t *testing.T) {
	selected, changes, callbacks := 0, 0, 0
	p := CarouselProps{Count: 3, Label: "Photo tour", Loop: true, Dots: true, Height: 120, OnSelect: func(int) { callbacks++ }}
	r := carouselTester(&selected, &p, func() { changes++ }, func(c *ui.Context, i int) { ui.Textf(c, "Photo %d", i) }, false)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyRight)
	if selected != 1 {
		t.Fatal("Right did not move focused carousel")
	}
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyRight)
	if selected != 0 {
		t.Fatal("loop next did not wrap to first")
	}
	r.Key(0, ui.KeyLeft)
	if selected != 2 {
		t.Fatal("loop previous did not wrap to last")
	}
	if err := r.Click("Go to slide 2"); err != nil {
		t.Fatal(err)
	}
	if selected != 1 || changes != 5 || callbacks != 5 {
		t.Fatalf("selection=%d changes=%d callbacks=%d", selected, changes, callbacks)
	}
	region, ok := r.AccessByRole(ui.RoleRegion, "Photo tour")
	if !ok || region.RoleDescription != "carousel" {
		t.Fatal("carousel region metadata missing")
	}
	slides := 0
	for _, info := range r.AccessAll() {
		if info.RoleDescription == "slide" {
			slides++
			if info.Name != "Slide 2 of 3" {
				t.Fatalf("inactive slide exposed: %+v", info)
			}
		}
	}
	if slides != 1 {
		t.Fatalf("accessible slides=%d", slides)
	}
	if _, ok := r.AccessByRole(ui.RoleStatus, "Slide 2 of 3"); !ok {
		t.Fatal("slide change has no live status")
	}
	p.Disabled = true
	r.Frame()
	r.Key(0, ui.KeyLeft)
	if selected != 1 {
		t.Fatal("disabled carousel moved")
	}
	p.Disabled = false
	p.Count = 0
	r.Frame()
	if selected != 0 {
		t.Fatal("empty carousel was not normalized")
	}
	next, _ := r.AccessByRole(ui.RoleButton, "Next slide")
	previous, _ := r.AccessByRole(ui.RoleButton, "Previous slide")
	if !next.Disabled || !previous.Disabled {
		t.Fatal("empty controls are active")
	}
}
func TestCarouselOrientationRTLAndEditableContentKeepsItsKeys(t *testing.T) {
	selected := 0
	values := []string{"alpha", "beta", "gamma"}
	p := CarouselProps{Count: 3, Loop: true, Height: 120}
	r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) { ui.TextInput(c, &values[i]).Label("Edit " + values[i]) }, false)
	r.SetDirection(ui.RightToLeft)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyRight)
	if selected != 2 {
		t.Fatalf("RTL right=%d", selected)
	}
	r.Key(0, ui.KeyLeft)
	if selected != 0 {
		t.Fatal("RTL left did not move forward")
	}
	p.Vertical = true
	r.Frame()
	r.Key(0, ui.KeyRight)
	if selected != 0 {
		t.Fatal("vertical carousel used horizontal key")
	}
	r.Key(0, ui.KeyDown)
	if selected != 1 {
		t.Fatal("vertical Down did not move")
	}
	r.Advance(400 * time.Millisecond)
	if err := r.Click("Edit beta"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyLeft)
	r.Key(0, ui.KeyRight)
	r.Type("X")
	if selected != 1 || values[1] == "beta" {
		t.Fatal("carousel stole its editable child's keys")
	}
	selected = 2
	r.Frame()
	r.Advance(400 * time.Millisecond)
	r.Key(0, ui.KeyUp)
	if selected != 1 {
		t.Fatal("focus was stranded in a hidden slide after a controlled page change")
	}
}
func TestCarouselNativeSwipeSnapBoundsAndCancellation(t *testing.T) {
	selected := 0
	p := CarouselProps{Count: 3, Height: 140, DragThreshold: 40}
	r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) {
		ui.Box(c).Fill().Background([]ui.Color{ui.RGB(230, 0, 0), ui.RGB(0, 180, 0), ui.RGB(0, 0, 230)}[i])
	}, false)
	viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
	x, y := viewport.Bounds.X+viewport.Bounds.W*.7, viewport.Bounds.Y+viewport.Bounds.H*.5
	r.Press(x, y)
	r.Move(x-170, y)
	r.Release(x-170, y)
	if selected != 1 {
		t.Fatalf("left swipe selected=%d", selected)
	}
	r.Advance(500 * time.Millisecond)
	selected = 2
	r.Frame()
	r.Advance(400 * time.Millisecond)
	r.Press(x, y)
	r.Move(x-170, y)
	r.Release(x-170, y)
	if selected != 2 {
		t.Fatal("bounded swipe passed last slide")
	}
	r.Advance(500 * time.Millisecond)
	p.Vertical = true
	selected = 0
	r.Frame()
	r.Advance(400 * time.Millisecond)
	viewport, _ = r.AccessByRole(ui.RoleGroup, "Slides")
	x, y = viewport.Bounds.X+viewport.Bounds.W*.5, viewport.Bounds.Y+viewport.Bounds.H*.8
	r.Press(x, y)
	r.Move(x, y-90)
	r.Release(x, y-90)
	if selected != 1 {
		t.Fatalf("up swipe selected=%d", selected)
	}
	r.Advance(500 * time.Millisecond)
	r.Press(x, y)
	r.Move(x, y-80)
	r.SetWindowActive(false)
	r.Release(x, y-80)
	if selected != 1 {
		t.Fatal("cancelled gesture selected a new slide")
	}
	r.SetWindowActive(true)
	r.Advance(500 * time.Millisecond)
	r.Press(x, y)
	r.Move(x, y-80)
	r.CancelPress()
	if selected != 1 {
		t.Fatal("system cancelled gesture selected a new slide")
	}
}

func TestCarouselPaintsSlidingPagesAndReducedMotionSettles(t *testing.T) {
	for _, reduce := range []bool{false, true} {
		t.Run(map[bool]string{false: "animated", true: "reduced"}[reduce], func(t *testing.T) {
			selected := 0
			p := CarouselProps{Count: 3, Height: 120, Duration: 600 * time.Millisecond, Loop: true}
			colors := []ui.Color{ui.RGB(220, 0, 0), ui.RGB(0, 170, 0), ui.RGB(0, 0, 220)}
			r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) { ui.Box(c).Fill().Background(colors[i]) }, reduce)
			if err := r.Click("Next slide"); err != nil {
				t.Fatal(err)
			}
			viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
			count := func(want color.RGBA) int {
				im := r.Image()
				found := 0
				y := int(viewport.Bounds.Y + viewport.Bounds.H/2)
				for x := int(viewport.Bounds.X + 2); x < int(viewport.Bounds.X+viewport.Bounds.W-2); x++ {
					if im.RGBAAt(x, y) == want {
						found++
					}
				}
				return found
			}
			red, green, blue := color.RGBA{R: 220, A: 255}, color.RGBA{G: 170, A: 255}, color.RGBA{B: 220, A: 255}
			if reduce {
				if r.FramePending() {
					t.Fatal("reduced-motion carousel scheduled animation")
				}
				if count(green) < 100 {
					t.Fatal("reduced carousel did not paint destination immediately")
				}
			} else {
				if !r.FramePending() {
					t.Fatal("slide motion scheduled no frames")
				}
				r.Advance(100 * time.Millisecond)
				if count(red) == 0 || count(green) == 0 {
					t.Fatal("slide transition did not paint parts of both adjacent pages")
				}
			}
			r.Advance(800 * time.Millisecond)
			if count(green) < 100 || r.FramePending() {
				t.Fatal("slide motion did not settle on destination")
			}
			selected = 2
			r.Frame()
			r.Advance(800 * time.Millisecond)
			if count(blue) < 100 {
				t.Fatal("last page not painted")
			}
			if err := r.Click("Next slide"); err != nil {
				t.Fatal(err)
			}
			if selected != 0 {
				t.Fatal("loop did not wrap")
			}
			if !reduce {
				r.Advance(100 * time.Millisecond)
				if count(blue) == 0 || count(red) == 0 {
					t.Fatal("loop did not slide directly between the last and first pages")
				}
			}
			r.Advance(800 * time.Millisecond)
			if count(red) < 100 || r.FramePending() {
				t.Fatal("loop boundary did not settle directly on first page")
			}
		})
	}
}

func TestCarouselLoopPreservesSlideLocalState(t *testing.T) {
	selected := 0
	created := [2]int{}
	p := CarouselProps{Count: 2, Height: 120, Loop: true}
	r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) {
		item := ui.Box(c)
		counter := ui.Local(item, "counter", func() int { created[i]++; return 0 })
		item.Children(func() {
			if ui.ButtonBase(c).Label("Slide counter").Children(func() { ui.Textf(c, "Count %d", *counter) }).Clicked() {
				*counter++
			}
		})
	}, true)
	if err := r.Click("Slide counter"); err != nil {
		t.Fatal(err)
	}
	for range 6 {
		if err := r.Click("Next slide"); err != nil {
			t.Fatal(err)
		}
	}
	if selected != 0 || created != [2]int{1, 1} || !r.HasText("Count 1") {
		t.Fatalf("loop remounted slide state: selected=%d created=%v texts=%v", selected, created, r.Texts())
	}
}

func TestCarouselTwoSlideLoopsRespectNavigationDirection(t *testing.T) {
	for _, next := range []bool{false, true} {
		t.Run(map[bool]string{false: "previous", true: "next"}[next], func(t *testing.T) {
			selected := 0
			p := CarouselProps{Count: 2, Height: 120, Loop: true, Duration: 600 * time.Millisecond}
			r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) {
				ui.Box(c).Fill().Background([]ui.Color{ui.RGB(220, 0, 0), ui.RGB(0, 170, 0)}[i])
			}, false)
			viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
			for turn := range 2 {
				action := "Previous slide"
				if next {
					action = "Next slide"
				}
				if err := r.Click(action); err != nil {
					t.Fatal(err)
				}
				r.Advance(100 * time.Millisecond)
				im := r.Image()
				y := int(viewport.Bounds.Y + viewport.Bounds.H/2)
				left := im.RGBAAt(int(viewport.Bounds.X)+2, y)
				right := im.RGBAAt(int(viewport.Bounds.X+viewport.Bounds.W)-2, y)
				red, green := color.RGBA{R: 220, A: 255}, color.RGBA{G: 170, A: 255}
				wantLeft, wantRight := red, green
				if next == (turn == 1) {
					wantLeft, wantRight = green, red
				}
				if left != wantLeft || right != wantRight {
					t.Fatalf("turn %d has wrong motion direction: left=%v right=%v", turn, left, right)
				}
				r.Advance(700 * time.Millisecond)
			}
		})
	}
}

func TestCarouselVerticalAndRTLTracksPaintTheirDestinations(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(map[bool]string{false: "rtl", true: "vertical"}[vertical], func(t *testing.T) {
			selected := 2
			p := CarouselProps{Count: 3, Height: 120, Vertical: vertical}
			r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) {
				ui.Box(c).Fill().Background([]ui.Color{ui.RGB(220, 0, 0), ui.RGB(0, 170, 0), ui.RGB(0, 0, 220)}[i])
			}, true)
			if !vertical {
				r.SetDirection(ui.RightToLeft)
			}
			viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
			pixel := func() color.RGBA {
				return r.Image().RGBAAt(int(viewport.Bounds.X+viewport.Bounds.W/2), int(viewport.Bounds.Y+viewport.Bounds.H/2))
			}
			if pixel() != (color.RGBA{B: 220, A: 255}) {
				t.Fatal("initial selected page did not paint")
			}
			selected = 0
			r.Frame()
			if pixel() != (color.RGBA{R: 220, A: 255}) {
				t.Fatal("controlled page change did not paint")
			}
			if err := r.Click("Next slide"); err != nil {
				t.Fatal(err)
			}
			if pixel() != (color.RGBA{G: 170, A: 255}) {
				t.Fatal("next page did not paint")
			}
		})
	}
}

func TestCarouselLoopRetargetKeepsVisiblePages(t *testing.T) {
	for _, count := range []int{2, 3} {
		for _, orientation := range []string{"horizontal", "vertical", "rtl"} {
			t.Run(fmt.Sprintf("%s/%d", orientation, count), func(t *testing.T) {
				selected := 0
				p := CarouselProps{Count: count, Height: 120, Vertical: orientation == "vertical", Loop: true, Duration: 600 * time.Millisecond}
				r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) {
					ui.Box(c).Fill().Background([]ui.Color{ui.RGB(220, 0, 0), ui.RGB(0, 170, 0), ui.RGB(0, 0, 220)}[i])
				}, false)
				if orientation == "rtl" {
					r.SetDirection(ui.RightToLeft)
				}
				viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
				pixels := func() []color.RGBA {
					im := r.Image()
					var row []color.RGBA
					for axis := 2; axis < int(viewport.Bounds.W)-2; axis++ {
						x, y := int(viewport.Bounds.X)+axis, int(viewport.Bounds.Y+viewport.Bounds.H/2)
						if p.Vertical {
							if axis >= int(viewport.Bounds.H)-2 {
								break
							}
							x, y = int(viewport.Bounds.X+viewport.Bounds.W/2), int(viewport.Bounds.Y)+axis
						}
						row = append(row, im.RGBAAt(x, y))
					}
					return row
				}
				painted := func() {
					t.Helper()
					missing := 0
					for _, px := range pixels() {
						if px != (color.RGBA{R: 220, A: 255}) && px != (color.RGBA{G: 170, A: 255}) && px != (color.RGBA{B: 220, A: 255}) {
							missing++
						}
					}
					// A fractional page boundary can antialias one pixel.
					if missing > 2 {
						t.Fatalf("blank viewport across %d pixels", missing)
					}
				}
				for _, action := range []string{"Next slide", "Next slide", "Next slide", "Previous slide", "Previous slide", "controlled"} {
					before := pixels()
					if action == "controlled" {
						selected = count - 1
						r.Frame()
					} else if err := r.Click(action); err != nil {
						t.Fatal(err)
					}
					after, changed := pixels(), 0
					for i := range before {
						if before[i] != after[i] {
							changed++
						}
					}
					if changed > 2 {
						t.Fatalf("%s jumped %d pixels at the same instant", action, changed)
					}
					painted()
					for range 4 {
						r.Advance(25 * time.Millisecond)
						painted()
					}
				}
				r.Advance(800 * time.Millisecond)
				painted()
				if r.FramePending() {
					t.Fatal("retargeted carousel did not stop requesting frames")
				}
			})
		}
	}
}

func TestCarouselLoopDragKeepsAdjacentPagesPainted(t *testing.T) {
	selected := 0
	p := CarouselProps{Count: 3, Height: 120, Loop: true, DragThreshold: 40}
	r := carouselTester(&selected, &p, nil, func(c *ui.Context, i int) {
		ui.Box(c).Fill().Background([]ui.Color{ui.RGB(220, 0, 0), ui.RGB(0, 170, 0), ui.RGB(0, 0, 220)}[i])
	}, false)
	viewport, _ := r.AccessByRole(ui.RoleGroup, "Slides")
	x, y := viewport.Bounds.X+viewport.Bounds.W/2, viewport.Bounds.Y+viewport.Bounds.H/2
	r.Press(x, y)
	r.Move(x+100, y)
	im := r.Image()
	if im.RGBAAt(int(viewport.Bounds.X)+10, int(y)) != (color.RGBA{B: 220, A: 255}) || im.RGBAAt(int(viewport.Bounds.X+viewport.Bounds.W)-10, int(y)) != (color.RGBA{R: 220, A: 255}) {
		t.Fatal("loop drag did not keep first and previous pages adjacent")
	}
	r.Release(x+100, y)
	r.Advance(500 * time.Millisecond)
	if selected != 2 || r.FramePending() {
		t.Fatal("loop drag did not select and settle on the previous page")
	}
	im = r.Image()
	if im.RGBAAt(int(x), int(y)) != (color.RGBA{B: 220, A: 255}) {
		t.Fatal("loop drag did not paint its settled destination")
	}
}
