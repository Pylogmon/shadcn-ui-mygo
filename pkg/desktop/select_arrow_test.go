package desktop

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/internal/indicator"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var selectArrowReference = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>`))

type selectArrowInk struct {
	bounds image.Rectangle
	mask   []byte
}

func selectArrowPixels(t *testing.T, im *image.RGBA) selectArrowInk {
	t.Helper()
	box := image.Rectangle{Min: im.Bounds().Max, Max: im.Bounds().Min}
	for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
			p := im.RGBAAt(x, y)
			if int(min(p.R, p.B))-int(p.G) > 20 {
				box.Min.X = min(box.Min.X, x)
				box.Min.Y = min(box.Min.Y, y)
				box.Max.X = max(box.Max.X, x+1)
				box.Max.Y = max(box.Max.Y, y+1)
			}
		}
	}
	if box.Empty() {
		t.Fatal("no real MutedForeground arrow ink")
	}
	mask := []byte{}
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			p := im.RGBAAt(x, y)
			v := max(0, int(min(p.R, p.B))-int(p.G))
			mask = append(mask, byte(v))
		}
	}
	return selectArrowInk{box, mask}
}
func selectArrowReferencePixels(t *testing.T) selectArrowInk {
	r := ui.NewTester(func(c *ui.Context) {
		ui.Box(c).Size(16, 16).Background(ui.RGB(255, 255, 255)).Children(func() { ui.Icon(c, selectArrowReference).Size(16, 16).FontSize(16).TextColor(ui.RGB(255, 0, 255)) })
	}, 16, 16)
	return selectArrowPixels(t, r.Image())
}
func TestSelectArrowHasFixedVectorPixelsAcrossTriggers(t *testing.T) {
	reference := selectArrowReferencePixels(t)
	for _, route := range []string{"flyout", "system", "in-window"} {
		for _, rtl := range []bool{false, true} {
			for _, style := range []struct {
				name   string
				size   control.Size
				font   string
				custom bool
				long   bool
			}{
				{"gallery12", control.DefaultSize, "system-ui", false, false},
				{"xs-monospace", control.ExtraSmall, "monospace", false, false},
				{"large-serif", control.Large, "serif", false, false},
				{"long-custom", control.Small, "monospace", true, true},
			} {
				t.Run(fmt.Sprintf("%s/rtl=%v/%s", route, rtl, style.name), func(t *testing.T) {
					selected := "one"
					label := "One"
					if style.long {
						label = strings.Repeat("Long selected label ", 8)
					}
					props := control.SelectProps{Label: "Arrow choice", Size: style.size, Options: []control.Option{{Value: "one", Label: label}, {Value: "two", Label: "Two"}}}
					if style.custom {
						props.TriggerContent = func(c *ui.Context, o control.Option) { ui.Text(c, o.Label).SingleLine().FontSize(22) }
					}
					backend := &fakeBackend{}
					host := NewHostWithBackend(backend)
					r := ui.NewTesterWith(func(c *ui.Context) {
						tokens := *theme.Light()
						tokens.MutedForeground = ui.RGB(255, 0, 255)
						c.SetTheme(tokens.Theme())
						host.BeginFrame()
						defer host.EndFrame()
						row := ui.Column(c).Padding(12).Font(style.font).FontSize(12)
						if rtl {
							row.Direction(ui.RightToLeft)
						}
						row.Children(func() {
							var trigger ui.Element
							switch route {
							case "flyout":
								var err error
								trigger, err = Select(c, host, "choice", &selected, props)
								if err != nil {
									panic(err)
								}
							case "system":
								trigger = control.NativeSelect(c, &selected, props)
							default:
								trigger = control.Select(c, &selected, props)
							}
							trigger.Width(220).FontSize(12)
						})
					}, 300, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
					arrow := selectArrowPixels(t, r.Image())
					if arrow.bounds.Size() != reference.bounds.Size() || !bytes.Equal(arrow.mask, reference.mask) {
						t.Fatalf("arrow differs from fixed16DIP native vector: ink=%v reference=%v", arrow.bounds, reference.bounds)
					}
					trigger, ok := r.Find("Arrow choice")
					if !ok {
						t.Fatal("trigger missing")
					}
					centerY := float64(arrow.bounds.Min.Y+arrow.bounds.Max.Y) / 2
					if math.Abs(centerY-float64(trigger.Y+trigger.H/2)) > .51 {
						t.Fatalf("arrow not vertically centered: ink%v trigger%v", arrow.bounds, trigger)
					}
					if float32(arrow.bounds.Min.X) < trigger.X || float32(arrow.bounds.Max.X) > trigger.X+trigger.W {
						t.Fatalf("long content pushed arrow outside trigger: ink%v trigger%v", arrow.bounds, trigger)
					}
					relativeX := float32(arrow.bounds.Min.X+arrow.bounds.Max.X)/2 - trigger.X
					if !rtl && relativeX < trigger.W/2 || rtl && relativeX > trigger.W/2 {
						t.Fatalf("arrow not on logical end: ink%v trigger%v", arrow.bounds, trigger)
					}
					for _, node := range r.AccessAll() {
						if node.Role == ui.RoleImage {
							t.Fatalf("decorative arrow exposed as image: %+v", node)
						}
					}
					if node, ok := r.Access("Arrow choice"); !ok || node.Name != "Arrow choice" {
						t.Fatalf("arrow changed trigger accessible name: %+v", node)
					}
				})
			}
		}
	}
}

func TestSelectArrowPreservesSystemAndFlyoutSelectionRoutes(t *testing.T) {
	for _, route := range []string{"flyout", "system", "in-window"} {
		t.Run(route, func(t *testing.T) {
			value := "one"
			changes := 0
			backend := &fakeBackend{}
			host := NewHostWithBackend(backend)
			r := ui.NewTesterWith(func(c *ui.Context) {
				tokens := *theme.Light()
				tokens.MutedForeground = ui.RGB(255, 0, 255)
				c.SetTheme(tokens.Theme())
				host.BeginFrame()
				defer host.EndFrame()
				p := control.SelectProps{Label: "Choose plan", Options: []control.Option{{Value: "one", Label: "One"}, {Value: "blocked", Label: "Unavailable", Disabled: true}, {Value: "two", Label: "Two"}}}
				var e ui.Element
				switch route {
				case "flyout":
					var err error
					e, err = Select(c, host, "plan", &value, p)
					if err != nil {
						panic(err)
					}
				case "system":
					e = control.NativeSelect(c, &value, p)
				default:
					e = control.Select(c, &value, p)
				}
				e.Width(220).FontSize(12)
				if e.Changed() {
					changes++
				}
			}, 300, 130, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			arrow := selectArrowPixels(t, r.Image())
			r.ClickAt(float32(arrow.bounds.Min.X+arrow.bounds.Max.X)/2, float32(arrow.bounds.Min.Y+arrow.bounds.Max.Y)/2)
			switch route {
			case "flyout":
				if len(backend.surfaces) != 1 || len(r.Menu()) != 0 {
					t.Fatal("Flyout used the system-menu route")
				}
				child := ui.NewTesterWith(backend.surfaces[0].view, 260, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				child.Key(0, ui.KeyEnd)
				child.Key(0, ui.KeyEnter)
				r.Frame()
			case "system":
				if len(backend.surfaces) != 0 || len(r.Menu()) != 3 {
					t.Fatal("system menu was replaced by a Flyout")
				}
				if err := r.ChooseMenuItem("Unavailable"); err == nil {
					t.Fatal("disabled system option selected")
				}
				if err := r.ChooseMenuItem("Two"); err != nil {
					t.Fatal(err)
				}
			default:
				r.Key(0, ui.KeyHome)
				r.Key(0, ui.KeyDown)
				r.Key(0, ui.KeyEnter)
				r.Advance(time.Second)
				r.Frame()
			}
			if value != "two" || changes != 1 {
				t.Fatalf("selection/change changed: value%q changes%d", value, changes)
			}
			r.Frame()
			if changes != 1 {
				t.Fatal("arrow change emitted another selection event")
			}
		})
	}
}

func TestSelectArrowIndicatorKeepsSixteenDIPBoxUnderFontAndFlexPressure(t *testing.T) {
	for _, font := range []struct {
		family string
		size   float32
	}{{"monospace", 12}, {"serif", 28}} {
		t.Run(font.family, func(t *testing.T) {
			var actual ui.Rect
			r := ui.NewTester(func(c *ui.Context) {
				ui.Row(c).Width(36).Height(44).AlignItems(ui.Center).Font(font.family).FontSize(font.size).LineHeight(3).Children(func() {
					ui.Text(c, "Long label under flex pressure").Width(36).MinWidth(0).Shrink(1)
					indicator.ChevronDown(c).TextColor(ui.RGB(255, 0, 255)).DrawOver(func(_ *ui.Painter, r ui.Rect) { actual = r })
				})
			}, 80, 60)
			if actual.W != 16 || actual.H != 16 {
				t.Fatalf("decorative SVG box shrank or inherited line height: %v", actual)
			}
			ink := selectArrowPixels(t, r.Image())
			if math.Abs(float64(ink.bounds.Min.Y+ink.bounds.Max.Y)/2-22) > .51 {
				t.Fatalf("vector not centered inside fixed16 box: ink%v layout%v", ink.bounds, actual)
			}
		})
	}
}
