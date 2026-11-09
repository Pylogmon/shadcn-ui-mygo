package conversation

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestMessageScrollerCompoundNativeFontInheritancePixelsAndVegaGap(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, family := range []string{"monospace", "serif"} {
			t.Run(fmt.Sprintf("%v/%s", dark, family), func(t *testing.T) {
				render := func(reference bool) *ui.Tester {
					var s MessageScrollerState
					return ui.NewTesterWith(func(c *ui.Context) {
						tokens := theme.Light()
						if dark {
							tokens = theme.Dark()
						}
						theme.Scope(c, tokens).Fill().Background(tokens.Background).Font(family).FontSize(22).FontWeight(700).LineHeight(1.6).TextColor(ui.RGB(30, 110, 150)).Children(func() {
							items := func(c *ui.Context) {
								for _, text := range []string{"MMMM inherited native text", "Second message"} {
									if reference {
										ui.Column(c).FillWidth().MinWidth(0).Shrink(0).Children(func() { ui.Text(c, text) })
									} else {
										MessageScrollerItem(c, MessageScrollerItemProps{ID: text}, func(c *ui.Context) { ui.Text(c, text) })
									}
								}
							}
							if reference {
								ui.Column(c).Fill().MinWidth(0).MinHeight(0).Clip().Children(func() {
									ui.Scroll(c).Fill().MinWidth(0).MinHeight(0).Scrollbars(false).PaddingEnd(9).Children(func() {
										ui.Column(c).FillWidth().MinWidth(0).MinHeightPercent(100).Shrink(0).Gap(32).Children(func() { items(c) })
									})
								})
							} else {
								MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
									MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) { MessageScrollerContent(c, MessageScrollerContentProps{}, items) })
								})
							}
						})
					}, 460, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
				}
				actual, reference := render(false), render(true)
				if !bytes.Equal(actual.Image().Pix, reference.Image().Pix) {
					t.Fatal("compound scopes changed inherited native glyph/font/line metrics")
				}
				first, _ := actual.Find("MMMM inherited native text")
				second, _ := actual.Find("Second message")
				if math.Abs(float64(second.Y-first.Y-first.H-32)) > .001 {
					t.Fatalf("Vega gap %v", second.Y-first.Y-first.H)
				}
				scrollerIdle(t, actual)
			})
		}
	}
}

func TestMessageScrollerCompoundSourceThumbTokenAndAutoTransparencyPixels(t *testing.T) {
	var s MessageScrollerState
	zero := float32(0)
	tokens := theme.Light()
	tokens.Border = ui.RGB(190, 20, 130)
	tokens.Background = ui.RGB(250, 245, 240)
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Scope(c, tokens).Fill().Background(tokens.Background).Children(func() {
			MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
				MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
					MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
						for i := 0; i < 8; i++ {
							MessageScrollerItem(c, MessageScrollerItemProps{ID: fmt.Sprint(i)}, func(c *ui.Context) { ui.Text(c, fmt.Sprint(i)).Height(80) })
						}
					})
				})
			})
		})
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if got := r.Image().RGBAAt(315, 20); got != (color.RGBA{190, 20, 130, 255}) {
		t.Fatalf("source Border thumb=%v", got)
	}
	s.ScrollToEnd()
	r.Frame()
	if !s.AutoScrolling() {
		t.Fatal("auto transparency flag absent")
	}
	if got := r.Image().RGBAAt(315, 170); got != (color.RGBA{250, 245, 240, 255}) {
		t.Fatalf("auto thumb did not become transparent: %v", got)
	}
	r.Advance(181 * time.Millisecond)
	r.Frame()
	if got := r.Image().RGBAAt(315, 170); got != (color.RGBA{190, 20, 130, 255}) {
		t.Fatalf("settled thumb token=%v", got)
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundFluentRTLRetainsLogicalGutterAndBar(t *testing.T) {
	var s MessageScrollerState
	zero := float32(0)
	r := ui.NewTesterWith(func(c *ui.Context) {
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
					for i := 0; i < 8; i++ {
						MessageScrollerItem(c, MessageScrollerItemProps{ID: fmt.Sprint(i)}, func(c *ui.Context) { ui.Text(c, fmt.Sprint(i)).Height(80) })
					}
				})
			})
		}).Direction(ui.RightToLeft)
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if s.bar.X != 0 {
		t.Fatalf("fluent RTL bar=%v", s.bar)
	}
	first, _ := r.Find("0")
	if first.X != 9 || first.W != 311 {
		t.Fatalf("fluent RTL stable gutter=%v", first)
	}
	scrollerIdle(t, r)
}
