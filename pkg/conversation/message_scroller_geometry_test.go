package conversation

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestMessageScrollerCompoundNativeScrollbarDragAndFollowInterruption(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprint(rtl), func(t *testing.T) {
			var s MessageScrollerState
			messages := scrollerRows(8)
			p := MessageScrollerProviderProps{AutoScroll: true, DefaultScrollPosition: MessageScrollerStart}
			r := scrollerFixture(t, &s, &messages, &p)
			if rtl {
				r.SetDirection(ui.RightToLeft)
			}
			x := float32(314)
			if rtl {
				x = 6
			}
			r.Move(x, 20)
			// A native thumb press must supersede an older queued command.
			s.ScrollToEnd()
			r.Press(x, 20)
			if s.Scroll.Y != 0 || s.Mode() != MessageScrollerFree {
				t.Fatalf("queued command overrode native thumb intent: y=%v mode=%v", s.Scroll.Y, s.Mode())
			}
			r.Move(x, 60)
			bar := ui.Scrollbar{Box: ui.Rect{W: 320, H: 200}, View: 200, Content: 640}
			start, _ := bar.Thumb()
			want := bar.OffsetAt(start + 40)
			if math.Abs(float64(s.Scroll.Y-want)) > .02 || s.Mode() != MessageScrollerFree {
				t.Fatalf("native drag y=%v want=%v mode=%v", s.Scroll.Y, want, s.Mode())
			}
			r.CancelPress()
			if math.Abs(float64(s.Scroll.Y-want)) > .02 || s.Scroll.Scrolling() {
				t.Fatalf("cancel did not retain MyGo's accepted offset: y=%v moving=%v", s.Scroll.Y, s.Scroll.Scrolling())
			}
			s.ScrollToEnd()
			r.Frame()
			r.Scroll(x, 100, 0, -50)
			if s.Scroll.Y != 390 || s.Mode() != MessageScrollerFree {
				t.Fatalf("native track wheel y=%v mode=%v", s.Scroll.Y, s.Mode())
			}
			r.SetPreferences(ui.Preferences{})
			s.ScrollToEnd(MessageScrollerScrollOptions{Smooth: true})
			r.Frame()
			r.Advance(30 * time.Millisecond)
			bar.Offset = s.Scroll.Y
			start, length := bar.Thumb()
			y := start + length/2
			r.Move(x, y)
			r.Press(x, y)
			r.Move(x, y-20)
			r.Release(x, y-20)
			if s.Scroll.Scrolling() || s.Mode() != MessageScrollerFree {
				t.Fatalf("native drag did not interrupt glide: mode=%v moving=%v", s.Mode(), s.Scroll.Scrolling())
			}
			accepted := s.Scroll.Y
			messages[len(messages)-1].height += 60
			r.Frame()
			if s.Scroll.Y != accepted || s.Mode() != MessageScrollerFree {
				t.Fatal("streaming resumed following after manual native scrolling")
			}
			scrollerIdle(t, r)
		})
	}
}

func TestMessageScrollerCompoundExplicitIDWinsSimultaneousDefaultChange(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(8)
	p := MessageScrollerProviderProps{}
	r := scrollerFixture(t, &s, &messages, &p)
	p.DefaultScrollPosition = MessageScrollerStart
	s.ScrollToMessage("message-3")
	r.Frame()
	r.Frame()
	if s.Scroll.Y != 240 || s.Mode() != MessageScrollerJumping {
		t.Fatalf("new default overrode explicit ID y%v mode%v", s.Scroll.Y, s.Mode())
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundContainmentAndPrependOptOut(t *testing.T) {
	var s MessageScrollerState
	var outer ui.ScrollState
	zero := float32(0)
	preserve := false
	messages := scrollerRows(8)
	r := ui.NewTesterWith(func(c *ui.Context) {
		ui.Scroll(c).Fill().TrackScroll(&outer).Children(func() {
			MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
				MessageScrollerViewport(c, MessageScrollerViewportProps{PreserveScrollOnPrepend: &preserve, ContainScrollAtEdges: true}, func(c *ui.Context) {
					MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
						for _, m := range messages {
							MessageScrollerItem(c, MessageScrollerItemProps{ID: m.id}, func(c *ui.Context) { ui.Text(c, m.id).Height(m.height) })
						}
					})
				})
			}).Height(200).Shrink(0)
			ui.Box(c).Height(800).Shrink(0)
		})
	}, 320, 250, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Scroll(100, 100, 0, -100)
	if outer.Y != 0 || s.Scroll.Y != 0 {
		t.Fatalf("start chain outer%v inner%v", outer.Y, s.Scroll.Y)
	}
	s.ScrollToEnd()
	r.Frame()
	r.Scroll(100, 100, 0, 100)
	if outer.Y != 0 || s.Scroll.Y != 440 {
		t.Fatalf("end chain outer%v inner%v", outer.Y, s.Scroll.Y)
	}
	s.Scroll.Y = 160
	r.Frame()
	messages = append([]scrollerTestMessage{{id: "older", height: 40}}, messages...)
	r.Frame()
	if s.Scroll.Y != 160 {
		t.Fatalf("disabled prepend preservation y%v", s.Scroll.Y)
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundPaddedAlignmentNegativeMarginAndGap(t *testing.T) {
	var s MessageScrollerState
	zero := float32(0)
	margin := float32(12)
	r := ui.NewTesterWith(func(c *ui.Context) {
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart, ScrollMargin: &margin}, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero, PaddingStart: 24, PaddingEnd: 24}, func(c *ui.Context) {
					for i := 0; i < 6; i++ {
						MessageScrollerItem(c, MessageScrollerItemProps{ID: fmt.Sprint(i)}, func(c *ui.Context) { ui.Text(c, fmt.Sprint(i)).Height(80) })
					}
				})
			})
		})
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	for _, tc := range []struct {
		align MessageScrollerAlign
		y     float32
	}{{MessageScrollerAlignStart, 148}, {MessageScrollerAlignCenter, 112}, {MessageScrollerAlignEnd, 100}} {
		s.ScrollToMessage("2", MessageScrollerScrollOptions{Align: tc.align})
		r.Frame()
		if s.Scroll.Y != tc.y {
			t.Errorf("padded %s y%v want%v", tc.align, s.Scroll.Y, tc.y)
		}
	}
	negative := float32(-20)
	s.ScrollToMessage("2", MessageScrollerScrollOptions{ScrollMargin: &negative})
	r.Frame()
	if s.Scroll.Y != 180 {
		t.Fatalf("negative source margin y%v want180", s.Scroll.Y)
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundTransformsAndMultipleOffscreenDynamicMeasurement(t *testing.T) {
	for _, reduced := range []bool{false, true} {
		t.Run(fmt.Sprint(reduced), func(t *testing.T) {
			var states [3]MessageScrollerState
			heights := [3]float32{80, 90, 100}
			r := ui.NewTesterWith(func(c *ui.Context) {
				ui.Column(c).Fill().Children(func() {
					for k := range states {
						MessageScrollerWith(c.Key(k), &states[k], MessageScrollerProviderProps{AutoScroll: true}, func(c *ui.Context) {
							MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
								MessageScrollerContent(c, MessageScrollerContentProps{}, func(c *ui.Context) {
									for i := 0; i < 5; i++ {
										MessageScrollerItem(c, MessageScrollerItemProps{ID: fmt.Sprintf("%d-%d", k, i)}, func(c *ui.Context) { ui.Text(c, fmt.Sprintf("%d-%d", k, i)).Height(heights[k]) })
									}
								})
							})
						}).Height(200).Shrink(0).Scale(.75, .75)
					}
				})
			}, 320, 220, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduced}})
			for pass := 0; pass < 3; pass++ {
				for k := range states {
					want := 5*heights[k] + 4*32 - 200
					if states[k].Scroll.Y != want || states[k].Scroll.MaxY != want {
						t.Fatalf("pass%d offscreen%d y%v max%v want%v", pass, k, states[k].Scroll.Y, states[k].Scroll.MaxY, want)
					}
				}
				scrollerIdle(t, r)
				for k := range heights {
					heights[k] += 30
				}
				r.Frame()
			}
			scrollerIdle(t, r)
		})
	}
}

func TestMessageScrollerCompoundUnmountCancelsGlideAndRemountKeepsPlace(t *testing.T) {
	var s MessageScrollerState
	show := true
	zero := float32(0)
	r := ui.NewTester(func(c *ui.Context) {
		if show {
			MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
				MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
					MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
						for i := 0; i < 30; i++ {
							MessageScrollerItem(c, MessageScrollerItemProps{ID: fmt.Sprint(i)}, func(c *ui.Context) { ui.Text(c, fmt.Sprint(i)).Height(80) })
						}
					})
				})
			})
		} else {
			ui.Text(c, "Other page")
		}
	}, 320, 200)
	s.ScrollToEnd(MessageScrollerScrollOptions{Smooth: true})
	r.Frame()
	r.Advance(50 * time.Millisecond)
	y := s.Scroll.Y
	show = false
	r.Frame()
	if s.mounted || s.Scroll.Scrolling() || s.ScrollToEnd() {
		t.Fatalf("unmounted mounted%v moving%v", s.mounted, s.Scroll.Scrolling())
	}
	r.Advance(time.Second)
	scrollerIdle(t, r)
	show = true
	r.Frame()
	if math.Abs(float64(s.Scroll.Y-y)) > .5 || s.Scroll.Scrolling() {
		t.Fatalf("remount moved saved place %v→%v", y, s.Scroll.Y)
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundEmptyEdgeCommandDoesNotQueueFutureRestore(t *testing.T) {
	var s MessageScrollerState
	messages := []scrollerTestMessage{}
	p := MessageScrollerProviderProps{}
	r := scrollerFixture(t, &s, &messages, &p)
	if !s.ScrollToStart() {
		t.Fatal("mounted empty viewport rejected start")
	}
	r.Frame()
	messages = scrollerRows(6)
	r.Frame()
	if s.Scroll.Y != 280 {
		t.Fatalf("empty start overrode future opening default y%v", s.Scroll.Y)
	}
	scrollerIdle(t, r)
}
