package conversation

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestMessageScrollerLifecycleRemovedViewportRejectsEdgeAndClearsVisibility(t *testing.T) {
	var s MessageScrollerState
	show := true
	zero := float32(0)
	r := ui.NewTesterWith(func(c *ui.Context) {
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{}, func(c *ui.Context) {
			if show {
				MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
					MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
						for _, m := range scrollerRows(5) {
							MessageScrollerItem(c, MessageScrollerItemProps{ID: m.id}, func(c *ui.Context) { ui.Text(c, m.id).Height(m.height) })
						}
					})
				})
			}
		})
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	show = false
	r.Frame()
	if s.ScrollToEnd() || s.ScrollToStart() {
		t.Error("edge command accepted after only the viewport unmounted")
	}
	if s.ScrollToMessage("message-2") {
		t.Error("removed native row still registered by ID")
	}
	if v := s.Visibility(); len(v.VisibleMessageIDs) != 0 || v.CurrentAnchorID != "" {
		t.Errorf("removed viewport retained visible native rows: %+v", v)
	}
}

func TestMessageScrollerLifecycleEmptyThenReloadAutoScrollUsesSourceFirstContentBranch(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(8)
	p := MessageScrollerProviderProps{AutoScroll: true}
	r := scrollerFixture(t, &s, &messages, &p)
	r.Advance(200 * time.Millisecond)
	r.Scroll(100, 100, 0, -80)
	if s.Mode() != MessageScrollerFree {
		t.Fatal("probe failed to release native follow")
	}
	messages = nil
	r.Frame()
	messages = scrollerRows(6)
	r.Frame()
	if s.Scroll.Y != 280 || s.Mode() != MessageScrollerFollowing {
		t.Fatalf("source previousItemCount=0 + autoScroll should reopen at live edge; y=%v mode=%v", s.Scroll.Y, s.Mode())
	}
}

func TestMessageScrollerLifecycleEmptyMountedEdgeCommandKeepsEarlierPendingID(t *testing.T) {
	var s MessageScrollerState
	s.ScrollToMessage("message-2")
	messages := []scrollerTestMessage{}
	p := MessageScrollerProviderProps{}
	r := scrollerFixture(t, &s, &messages, &p)
	if !s.ScrollToStart() {
		t.Fatal("live empty viewport should accept edge command")
	}
	r.Frame()
	messages = scrollerRows(6)
	r.Frame()
	if s.Scroll.Y != 160 {
		t.Fatalf("pending ID was erased by source edge command: y=%v", s.Scroll.Y)
	}
}

func TestMessageScrollerLifecycleLateSecondInstanceOffscreenScaleDoesNotPingPong(t *testing.T) {
	var states [2]MessageScrollerState
	second := false
	height := float32(80)
	r := ui.NewTesterWith(func(c *ui.Context) {
		ui.Column(c).Fill().Clip().Children(func() {
			count := 1
			if second {
				count = 2
			}
			for k := 0; k < count; k++ {
				MessageScrollerWith(c.Key(k), &states[k], MessageScrollerProviderProps{AutoScroll: true}, func(c *ui.Context) {
					MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
						MessageScrollerContent(c, MessageScrollerContentProps{}, func(c *ui.Context) {
							for i := 0; i < 5; i++ {
								MessageScrollerItem(c, MessageScrollerItemProps{ID: fmt.Sprintf("%d-%d", k, i)}, func(c *ui.Context) { ui.Box(c).Height(height) })
							}
						})
					})
				}).Height(200).Shrink(0).Scale(.75, .75).Draw(func(_ *ui.Painter, _ ui.Rect) { second = true })
			}
		})
	}, 320, 150, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	scrollerIdle(t, r)
	height = 140
	r.Frame()
	for k := range states {
		if want := 5*height + 4*32 - 200; states[k].Scroll.Y != want {
			t.Fatalf("offscreen%d grew to y=%v want%v", k, states[k].Scroll.Y, want)
		}
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerLifecycleNestedNativeScrollStillConsumesWheelAtTranscriptEdge(t *testing.T) {
	var s MessageScrollerState
	var child ui.ScrollState
	zero := float32(0)
	r := ui.NewTesterWith(func(c *ui.Context) {
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{AutoScroll: true}, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
					MessageScrollerItem(c, MessageScrollerItemProps{ID: "top"}, func(c *ui.Context) { ui.Box(c).Height(300) })
					MessageScrollerItem(c, MessageScrollerItemProps{ID: "nested"}, func(c *ui.Context) {
						ui.Scroll(c).Height(100).FillWidth().TrackScroll(&child).Role(ui.RoleRegion).Label("Inner native scroll").Children(func() { ui.Box(c).Height(500).Shrink(0) })
					})
				})
			})
		})
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(200 * time.Millisecond)
	node, ok := r.AccessByRole(ui.RoleRegion, "Inner native scroll")
	if !ok {
		t.Fatal("missing child native viewport")
	}
	before := s.Scroll.Y
	r.Scroll(node.Bounds.X+30, node.Bounds.Y+30, 0, 40)
	if child.Y != 40 || s.Scroll.Y != before {
		t.Fatalf("ancestor containment stole child wheel: child=%v transcript=%v want transcript%v", child.Y, s.Scroll.Y, before)
	}
}

func TestMessageScrollerLifecycleNativeIntentCancelsNotYetReconciledEdgeCommand(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(8)
	p := MessageScrollerProviderProps{AutoScroll: true, DefaultScrollPosition: MessageScrollerStart}
	r := scrollerFixture(t, &s, &messages, &p)
	if !s.ScrollToEnd() {
		t.Fatal("live command rejected")
	}
	r.Scroll(100, 100, 0, 40)
	if s.Scroll.Y != 40 || s.Mode() != MessageScrollerFree {
		t.Fatalf("old queued edge command overrode later native wheel intent: y%v mode%v", s.Scroll.Y, s.Mode())
	}
}

func TestMessageScrollerLifecycleRootUnmountCancelsNotYetReconciledEdgeCommand(t *testing.T) {
	var s MessageScrollerState
	show := true
	zero := float32(0)
	r := ui.NewTesterWith(func(c *ui.Context) {
		if !show {
			return
		}
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
					for _, m := range scrollerRows(8) {
						MessageScrollerItem(c, MessageScrollerItemProps{ID: m.id}, func(c *ui.Context) { ui.Text(c, m.id).Height(m.height) })
					}
				})
			})
		})
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	s.Scroll.Y = 100
	r.Frame()
	s.ScrollToEnd()
	show = false
	r.Frame()
	show = true
	r.Frame()
	if s.Scroll.Y != 100 {
		t.Fatalf("unmounted live edge request replayed on native remount: y%v", s.Scroll.Y)
	}
}

func TestMessageScrollerLifecycleKnownIDCommandSupersedesEarlierUnknownQueue(t *testing.T) {
	var s MessageScrollerState
	s.ScrollToMessage("future")
	messages := []scrollerTestMessage{}
	p := MessageScrollerProviderProps{}
	r := scrollerFixture(t, &s, &messages, &p)
	messages = scrollerRows(2)
	r.Frame()
	if !s.ScrollToMessage("message-1") {
		t.Fatal("known message command rejected")
	}
	r.Frame()
	messages = append(messages, scrollerTestMessage{id: "future", height: 80})
	r.Frame()
	if s.Scroll.Y != 80 {
		t.Fatalf("superseded unknown ID jumped later: y%v", s.Scroll.Y)
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerLifecyclePendingUnknownIDSurvivesEdgesAndNativeIntent(t *testing.T) {
	for _, auto := range []bool{false, true} {
		for _, edge := range []MessageScrollerPosition{MessageScrollerStart, MessageScrollerEnd} {
			t.Run(fmt.Sprintf("%v/%s", auto, edge), func(t *testing.T) {
				var s MessageScrollerState
				s.ScrollToMessage("message-2")
				messages := []scrollerTestMessage{}
				p := MessageScrollerProviderProps{AutoScroll: auto}
				r := scrollerFixture(t, &s, &messages, &p)
				if edge == MessageScrollerStart {
					s.ScrollToStart()
				} else {
					s.ScrollToEnd()
				}
				r.Scroll(100, 100, 0, 40)
				messages = scrollerRows(6)
				r.Frame()
				if s.Scroll.Y != 160 || s.Mode() != MessageScrollerJumping {
					t.Fatalf("unknown queue lost after edge/gesture y%v mode%v", s.Scroll.Y, s.Mode())
				}
				scrollerIdle(t, r)
			})
		}
	}
}

func TestMessageScrollerLifecycleContentAndViewportMountSequences(t *testing.T) {
	for _, part := range []string{"viewport", "content"} {
		t.Run(part, func(t *testing.T) {
			var s MessageScrollerState
			show := true
			zero := float32(0)
			r := ui.NewTesterWith(func(c *ui.Context) {
				MessageScrollerWith(c, &s, MessageScrollerProviderProps{}, func(c *ui.Context) {
					if part == "viewport" && !show {
						return
					}
					MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
						if show {
							MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
								for _, m := range scrollerRows(8) {
									MessageScrollerItem(c, MessageScrollerItemProps{ID: m.id}, func(c *ui.Context) { ui.Text(c, m.id).Height(m.height) })
								}
							})
						}
					})
				})
			}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			s.Scroll.Y = 100
			r.Frame()
			show = false
			r.Frame()
			if s.ScrollToMessage("message-2") || len(s.Visibility().VisibleMessageIDs) > 0 || s.Scrollable() != (MessageScrollerScrollable{}) {
				t.Fatal("unmounted part retained old registrations/visibility")
			}
			accepted := s.ScrollToEnd()
			if accepted != (part == "content") {
				t.Fatalf("edge availability after %s removal=%v", part, accepted)
			}
			r.Frame()
			scrollerIdle(t, r)
			show = true
			r.Frame()
			if !s.ScrollToMessage("message-2") {
				t.Fatal("remounted row was not registered")
			}
			r.Frame()
			if s.Scroll.Y != 160 {
				t.Fatalf("remounted ID target y%v", s.Scroll.Y)
			}
			scrollerIdle(t, r)
		})
	}
}

func TestMessageScrollerLifecycleNativeIntentCancelsKnownIDAndEdges(t *testing.T) {
	for _, command := range []string{"start", "end", "message"} {
		for _, input := range []string{"wheel", "key"} {
			t.Run(command+"/"+input, func(t *testing.T) {
				var s MessageScrollerState
				messages := scrollerRows(8)
				p := MessageScrollerProviderProps{AutoScroll: true, DefaultScrollPosition: MessageScrollerStart}
				r := scrollerFixture(t, &s, &messages, &p)
				if input == "key" {
					s.Scroll.Handle.Focus()
					r.Frame()
				}
				switch command {
				case "start":
					s.ScrollToStart()
				case "end":
					s.ScrollToEnd()
				case "message":
					s.ScrollToMessage("message-5")
				}
				if input == "key" {
					r.Key(0, ui.KeyPageDown)
				} else {
					r.Scroll(100, 100, 0, 40)
				}
				want := float32(40)
				if input == "key" {
					want = 160
				}
				if s.Scroll.Y != want || s.Mode() != MessageScrollerFree {
					t.Fatalf("stale %s overrode %s y%v mode%v", command, input, s.Scroll.Y, s.Mode())
				}
				scrollerIdle(t, r)
			})
		}
	}
}

func TestMessageScrollerLifecycleContentRemovalCancelsNativeGlide(t *testing.T) {
	var s MessageScrollerState
	show := true
	zero := float32(0)
	r := ui.NewTester(func(c *ui.Context) {
		MessageScrollerWith(c, &s, MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				if show {
					MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
						for _, m := range scrollerRows(30) {
							MessageScrollerItem(c, MessageScrollerItemProps{ID: m.id}, func(c *ui.Context) { ui.Text(c, m.id).Height(m.height) })
						}
					})
				}
			})
		})
	}, 320, 200)
	s.ScrollToEnd(MessageScrollerScrollOptions{Smooth: true})
	r.Frame()
	r.Advance(50 * time.Millisecond)
	if !s.Scroll.Scrolling() {
		t.Fatal("probe did not start native glide")
	}
	show = false
	r.Frame()
	if s.Scroll.Scrolling() || s.move != nil || s.smoothTarget != nil {
		t.Fatal("removed content retained native glide or target")
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerLifecyclePrependAndReplyCollapseKeepAnchoredReadingLine(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(5)
	messages[2].anchor = true
	p := MessageScrollerProviderProps{AutoScroll: true, DefaultScrollPosition: MessageScrollerLastAnchor}
	r := scrollerFixture(t, &s, &messages, &p)
	messages = append([]scrollerTestMessage{{id: "older", height: 40}}, messages...)
	messages[4].height = 0
	messages[5].height = 0
	r.Frame()
	if s.Scroll.Y != 136 || s.TailSpace() != 56 || s.Mode() != MessageScrollerAnchored {
		t.Fatalf("prepend/collapse lost reading line y%v tail%v mode%v", s.Scroll.Y, s.TailSpace(), s.Mode())
	}
	anchor, _ := r.Find("message-2")
	if anchor.Y != 64 {
		t.Fatalf("real anchored row top%v", anchor.Y)
	}
	scrollerIdle(t, r)
}
