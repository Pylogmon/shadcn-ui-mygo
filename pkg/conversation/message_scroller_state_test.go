package conversation

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

type scrollerTestMessage struct {
	id     string
	height float32
	anchor bool
}

func scrollerIdle(t *testing.T, r *ui.Tester) {
	t.Helper()
	r.Advance(time.Second)
	for i := 0; i < 20 && r.FramePending(); i++ {
		r.Frame()
	}
	if r.FramePending() {
		t.Fatal("scroller continues scheduling frames at rest")
	}
}
func TestMessageScrollerCompoundNativeWheelReleasesAndRearmsFollowing(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(8)
	p := MessageScrollerProviderProps{AutoScroll: true}
	r := scrollerFixture(t, &s, &messages, &p)
	r.Advance(200 * time.Millisecond)
	r.Scroll(100, 100, 0, -80)
	if s.Scroll.Y != 360 || s.Mode() != MessageScrollerFree {
		t.Fatalf("wheel y=%v mode=%v", s.Scroll.Y, s.Mode())
	}
	messages[len(messages)-1].height = 140
	r.Frame()
	if s.Scroll.Y != 360 {
		t.Fatalf("growth took reader to%v", s.Scroll.Y)
	}
	r.Scroll(100, 100, 0, 1000)
	if s.Mode() != MessageScrollerFollowing || s.Scroll.Y != 500 {
		t.Fatalf("rearm y=%v mode=%v", s.Scroll.Y, s.Mode())
	}
	messages = append(messages, scrollerTestMessage{id: "more", height: 80})
	r.Frame()
	if s.Scroll.Y != 580 {
		t.Fatalf("following y=%v", s.Scroll.Y)
	}
	scrollerIdle(t, r)
}
func TestMessageScrollerCompoundQueuedIDWinsDefaultAndAlignments(t *testing.T) {
	var s MessageScrollerState
	if !s.ScrollToMessage("message-2") {
		t.Fatal("empty queue rejected")
	}
	messages := []scrollerTestMessage{}
	p := MessageScrollerProviderProps{AutoScroll: true}
	r := scrollerFixture(t, &s, &messages, &p)
	if s.Pending() {
		t.Fatal("empty viewport remains pending")
	}
	messages = scrollerRows(6)
	r.Frame()
	if s.Scroll.Y != 160 || s.Mode() != MessageScrollerJumping {
		t.Fatalf("queued y=%v mode=%v", s.Scroll.Y, s.Mode())
	}
	for _, tc := range []struct {
		align MessageScrollerAlign
		y     float32
	}{{MessageScrollerAlignStart, 160}, {MessageScrollerAlignCenter, 100}, {MessageScrollerAlignEnd, 40}, {MessageScrollerAlignNearest, 40}} {
		if !s.ScrollToMessage("message-2", MessageScrollerScrollOptions{Align: tc.align}) {
			t.Fatal("missing ID")
		}
		r.Frame()
		if s.Scroll.Y != tc.y {
			t.Errorf("%s y=%v want%v", tc.align, s.Scroll.Y, tc.y)
		}
	}
	if s.ScrollToMessage("missing") {
		t.Fatal("nonempty unknown ID accepted")
	}
	scrollerIdle(t, r)
}
func TestMessageScrollerCompoundLastAnchorVisibilityAndPrepend(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(6)
	messages[2].anchor = true
	p := MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerLastAnchor}
	r := scrollerFixture(t, &s, &messages, &p)
	if s.Scroll.Y != 96 || s.Visibility().CurrentAnchorID != "message-2" || !reflect.DeepEqual(s.Visibility().VisibleMessageIDs, []string{"message-2", "message-3"}) {
		t.Fatalf("last-anchor y=%v visibility=%+v", s.Scroll.Y, s.Visibility())
	}
	before, _ := r.Find("message-2")
	messages = append([]scrollerTestMessage{{id: "older", height: 60, anchor: true}}, messages...)
	r.Frame()
	after, _ := r.Find("message-2")
	if after.Y != before.Y || s.Scroll.Y != 156 {
		t.Fatalf("prepend rect %v→%v y=%v", before, after, s.Scroll.Y)
	}
	scrollerIdle(t, r)
}
func TestMessageScrollerCompoundSameCountReplacedAnchorAndBulkAnchors(t *testing.T) {
	for _, auto := range []bool{false, true} {
		t.Run(fmt.Sprint(auto), func(t *testing.T) {
			var s MessageScrollerState
			messages := scrollerRows(6)
			p := MessageScrollerProviderProps{AutoScroll: auto}
			r := scrollerFixture(t, &s, &messages, &p)
			messages[5] = scrollerTestMessage{id: "replaced", height: 40, anchor: true}
			r.Frame()
			if s.Scroll.Y != 336 || s.Mode() != MessageScrollerAnchored {
				t.Fatalf("replace y=%v mode=%v", s.Scroll.Y, s.Mode())
			}
			s.ScrollToEnd()
			r.Frame()
			messages = append(messages, scrollerTestMessage{id: "batch-a", height: 80, anchor: true}, scrollerTestMessage{id: "batch-b", height: 80, anchor: true})
			r.Frame()
			want := float32(376)
			if auto {
				want = 400
			}
			if s.Scroll.Y != want {
				t.Fatalf("batch auto%v y=%v want%v", auto, s.Scroll.Y, want)
			}
			scrollerIdle(t, r)
		})
	}
}
func TestMessageScrollerCompoundReleasedTailConvergesWithoutIdleFrames(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(4)
	p := MessageScrollerProviderProps{AutoScroll: true}
	r := scrollerFixture(t, &s, &messages, &p)
	r.Advance(200 * time.Millisecond)
	messages = append(messages, scrollerTestMessage{id: "turn", height: 40, anchor: true})
	r.Frame()
	r.Scroll(100, 100, 0, -120)
	if s.Mode() != MessageScrollerFree {
		t.Fatalf("gesture mode=%v", s.Mode())
	}
	messages = append(messages, scrollerTestMessage{id: "stream", height: 30})
	r.Frame()
	if s.TailSpace() != 96 || s.Scroll.Y != 136 {
		t.Fatalf("reader above spacer moved: y=%v tail=%v", s.Scroll.Y, s.TailSpace())
	}
	scrollerIdle(t, r)
	r.Scroll(100, 100, 0, 1000)
	if s.TailSpace() != 0 || s.Mode() != MessageScrollerFollowing {
		t.Fatalf("finish tail=%v mode=%v y=%v", s.TailSpace(), s.Mode(), s.Scroll.Y)
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundTallCenterUsesPaddedInset(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(5)
	messages[2].height = 400
	p := MessageScrollerProviderProps{DefaultScrollPosition: MessageScrollerStart}
	r := scrollerFixture(t, &s, &messages, &p)
	s.ScrollToMessage("message-2", MessageScrollerScrollOptions{Align: MessageScrollerAlignCenter})
	r.Frame()
	if s.Scroll.Y != 260 {
		t.Fatalf("tall row center y=%v want260", s.Scroll.Y)
	}
}
func TestMessageScrollerCompoundBareScrollAwayAndDefaultChange(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(8)
	p := MessageScrollerProviderProps{AutoScroll: true}
	r := scrollerFixture(t, &s, &messages, &p)
	r.Advance(200 * time.Millisecond)
	s.Scroll.Y = 80
	r.Frame()
	if s.Mode() != MessageScrollerFree {
		t.Fatalf("bare offset did not release mode%v", s.Mode())
	}
	messages[7].height = 200
	r.Frame()
	if s.Scroll.Y != 80 {
		t.Fatalf("growth moved bare reader %v", s.Scroll.Y)
	}
	p.DefaultScrollPosition = MessageScrollerStart
	r.Frame()
	if s.Scroll.Y != 0 || s.Pending() {
		t.Fatalf("live start y%v pending%v", s.Scroll.Y, s.Pending())
	}
	p.DefaultScrollPosition = MessageScrollerEnd
	r.Frame()
	if s.Scroll.Y != 560 {
		t.Fatalf("live end y%v", s.Scroll.Y)
	}
	scrollerIdle(t, r)
}
func TestMessageScrollerCompoundAnchoredCollapseAndTallTurnDoesNotFollow(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(5)
	messages[2].anchor = true
	p := MessageScrollerProviderProps{AutoScroll: true, DefaultScrollPosition: MessageScrollerLastAnchor}
	r := scrollerFixture(t, &s, &messages, &p)
	if s.Scroll.Y != 96 {
		t.Fatalf("open %v", s.Scroll.Y)
	}
	messages[3].height = 0
	messages[4].height = 0
	r.Frame()
	if s.Scroll.Y != 96 || s.TailSpace() != 56 || s.Mode() != MessageScrollerAnchored {
		t.Fatalf("collapse y%v tail%v mode%v", s.Scroll.Y, s.TailSpace(), s.Mode())
	}
	messages[2].height = 400
	r.Frame()
	// Consuming reserved room hands over to follow output.
	if s.Mode() != MessageScrollerFollowing {
		t.Fatalf("reserved handoff mode%v", s.Mode())
	}
	messages = append(messages, scrollerTestMessage{id: "tall-turn", height: 400, anchor: true})
	r.Frame()
	if s.Mode() != MessageScrollerAnchored || s.TailSpace() != 0 {
		t.Fatalf("tall place mode%v tail%v", s.Mode(), s.TailSpace())
	}
	y := s.Scroll.Y
	messages[len(messages)-1].height = 600
	r.Frame()
	if s.Scroll.Y != y || s.Mode() != MessageScrollerAnchored {
		t.Fatalf("tall turn y%v→%v mode%v", y, s.Scroll.Y, s.Mode())
	}
	scrollerIdle(t, r)
}

func TestMessageScrollerCompoundAutoScrollReenabledResumesExistingFollowMode(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(8)
	p := MessageScrollerProviderProps{AutoScroll: true}
	r := scrollerFixture(t, &s, &messages, &p)
	p.AutoScroll = false
	r.Frame()
	messages = append(messages, scrollerTestMessage{id: "while-paused", height: 80})
	r.Frame()
	if s.Scroll.Y != 440 {
		t.Fatalf("auto false moved%v", s.Scroll.Y)
	}
	p.AutoScroll = true
	r.Frame()
	if s.Scroll.Y != 520 {
		t.Fatalf("reenabling existing follow mode did not follow%v", s.Scroll.Y)
	}
	scrollerIdle(t, r)
}

func scrollerFixture(t *testing.T, s *MessageScrollerState, messages *[]scrollerTestMessage, p *MessageScrollerProviderProps) *ui.Tester {
	t.Helper()
	zero := float32(0)
	return ui.NewTesterWith(func(c *ui.Context) {
		MessageScrollerWith(c, s, *p, func(c *ui.Context) {
			MessageScrollerViewport(c, MessageScrollerViewportProps{}, func(c *ui.Context) {
				MessageScrollerContent(c, MessageScrollerContentProps{Gap: &zero}, func(c *ui.Context) {
					for _, m := range *messages {
						MessageScrollerItem(c, MessageScrollerItemProps{ID: m.id, ScrollAnchor: m.anchor}, func(c *ui.Context) { ui.Text(c, m.id).Height(m.height) })
					}
				})
			})
		})
	}, 320, 200, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func scrollerRows(n int) []scrollerTestMessage {
	rows := make([]scrollerTestMessage, n)
	for i := range rows {
		rows[i] = scrollerTestMessage{id: fmt.Sprintf("message-%d", i), height: 80}
	}
	return rows
}
func TestMessageScrollerCompoundOpeningAndFollow(t *testing.T) {
	for _, auto := range []bool{false, true} {
		t.Run(fmt.Sprint(auto), func(t *testing.T) {
			var s MessageScrollerState
			messages := scrollerRows(4)
			p := MessageScrollerProviderProps{AutoScroll: auto}
			r := scrollerFixture(t, &s, &messages, &p)

			if s.Scroll.Y != 120 {
				t.Fatalf("default end Y=%v want120", s.Scroll.Y)
			}
			messages = append(messages, scrollerTestMessage{id: "appended", height: 80})
			r.Frame()
			want := float32(120)
			if auto {
				want = 200
			}
			if s.Scroll.Y != want {
				t.Fatalf("append auto=%v Y=%v want%v", auto, s.Scroll.Y, want)
			}
		})
	}
}
func TestMessageScrollerCompoundAnchorConsumesTailAndHandsOff(t *testing.T) {
	var s MessageScrollerState
	messages := scrollerRows(4)
	p := MessageScrollerProviderProps{AutoScroll: true}
	r := scrollerFixture(t, &s, &messages, &p)
	messages = append(messages, scrollerTestMessage{id: "new-turn", height: 40, anchor: true})
	r.Frame()
	if s.Scroll.Y != 256 || s.TailSpace() != 96 || s.Mode() != MessageScrollerAnchored {
		t.Fatalf("anchor y=%v tail=%v mode=%v", s.Scroll.Y, s.TailSpace(), s.Mode())
	}
	messages = append(messages, scrollerTestMessage{id: "reply", height: 40})
	r.Frame()
	if s.Scroll.Y != 256 || s.TailSpace() != 56 {
		t.Fatalf("stream y=%v tail=%v", s.Scroll.Y, s.TailSpace())
	}
	messages[len(messages)-1].height = 140
	r.Frame()
	if s.Scroll.Y != 300 || s.TailSpace() != 0 || s.Mode() != MessageScrollerFollowing {
		t.Fatalf("handoff y=%v tail=%v mode=%v", s.Scroll.Y, s.TailSpace(), s.Mode())
	}
}
