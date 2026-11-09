package gallery

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
)

func compositionTester(s *State, part string, width int, reduce bool) *ui.Tester {
	s.Page = "conversation"
	return ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(12).Children(func() {
			if s.Page != "conversation" {
				ui.Text(c, "Another page")
				return
			}
			switch part {
			case "message":
				s.messageComposition(c)
			case "markers":
				s.markerComposition(c)
			case "chat":
				s.conversationChatComposition(c)
			default:
				s.conversationComposition(c)
			}
		})
	}, width, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: reduce}})
}

func TestCreatorConversationCompositionFollowSwitchControlsRealJumpMode(t *testing.T) {
	s := New()
	r := compositionTester(s, "chat", 620, true)
	p := &s.ConversationComposition
	advancedClick(t, r, "Follow new replies")
	if p.follow || advancedAccess(t, r, ui.RoleSwitch, "Follow new replies").Checked {
		t.Fatal("native switch did not disable following")
	}
	advancedClick(t, r, "Back to beginning")
	advancedClick(t, r, "Jump to latest")
	if math.Abs(float64(p.scroll.Scroll.Y-p.scroll.Scroll.MaxY)) > .5 || p.scroll.Mode() != conversation.MessageScrollerFree {
		t.Fatal("jump ignored the application-owned follow setting")
	}
	advancedClick(t, r, "Follow new replies")
	if !advancedAccess(t, r, ui.RoleSwitch, "Follow new replies").Checked || p.scroll.Mode() != conversation.MessageScrollerFollowing {
		t.Fatal("enabling following at the actual live edge did not rearm it")
	}
}

func TestCreatorConversationCompositionWholePageNarrowInputAndPersistence(t *testing.T) {
	for _, width := range []int{500, 1100} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			s := New()
			s.Page, s.Config.ReduceMotion = "conversation", true
			r := ui.NewTesterWith(s.View, width, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			conversationReveal(t, r, "Copy team update")
			advancedClick(t, r, "Copy team update")
			if r.Clipboard() != compositionUpdate {
				t.Fatal("complete Message was not wired into the real Gallery page")
			}
			conversationReveal(t, r, "Chat message")
			field := advancedAccess(t, r, ui.RoleTextField, "Chat message")
			panel := advancedAccess(t, r, ui.RoleRegion, "Messages & questions")
			if field.Bounds.X < panel.Bounds.X || field.Bounds.X+field.Bounds.W > panel.Bounds.X+panel.Bounds.W {
				t.Fatalf("narrow message editor escaped the native page: field%v panel%v", field.Bounds, panel.Bounds)
			}
			advancedClick(t, r, "Chat message")
			r.Type("A draft to keep")
			s.Page = "display"
			r.Frame()
			s.Page = "conversation"
			r.Frame()
			conversationReveal(t, r, "Chat message")
			if advancedAccess(t, r, ui.RoleTextField, "Chat message").Value != "A draft to keep" {
				t.Fatal("actual page navigation discarded the draft")
			}
			advancedClick(t, r, "Chat message")
			r.Key(ui.Cmd, ui.KeyA)
			r.Type("A question from the real page")
			r.Key(0, ui.KeyEnter)
			if !s.ConversationComposition.streaming {
				t.Fatal("the full-page native editor did not send")
			}
			conversationReveal(t, r, "Stop reply")
			advancedClick(t, r, "Stop reply")
			if !r.Focused("Chat message") {
				t.Fatal("whole-page Stop did not return focus to its native editor")
			}
		})
	}
}

func compositionFinishReply(t *testing.T, r *ui.Tester, p *conversationCompositionState) {
	t.Helper()
	for n := 0; p.streaming && n < 12; n++ {
		r.Advance(650 * time.Millisecond)
	}
	if p.streaming {
		t.Fatal("the actual timed reply never completed")
	}
}

func TestCreatorConversationCompositionShowsCompleteMessageAndCopiesVisibleUpdate(t *testing.T) {
	s := New()
	r := ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		ui.Column(c).Fill().Padding(20).Children(func() { s.conversationComposition(c) })
	}, 1100, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.HasText("A shared update") {
		t.Fatal("conversation page lacks its complete message composition")
	}
	advancedClick(t, r, "Copy team update")
	if r.Clipboard() != "The review is ready. Check the message guide before sharing it with the team." {
		t.Fatal("message footer did not copy its actual visible update")
	}
}

func TestCreatorConversationCompositionComplexBubbleOpensNativeURLAndKeepsSlots(t *testing.T) {
	s := New()
	r := compositionTester(s, "message", 620, true)
	advancedAccess(t, r, ui.RoleArticle, "Team update")
	if !r.HasText("Mira Chen · Just now") || !r.HasText("Shared with the team") {
		t.Fatal("message header/footer slots are missing")
	}
	link := advancedAccess(t, r, ui.RoleLink, "Open the message guide")
	r.ClickAt(link.Bounds.X+3, link.Bounds.Y+link.Bounds.H/2)
	r.Key(0, ui.KeySpace)
	r.Key(0, ui.KeyEnter)
	if urls := r.OpenedURLs(); len(urls) != 2 || urls[0] != compositionGuideURL || urls[1] != compositionGuideURL {
		t.Fatalf("complex native Box/Row link lost URL or Enter-only behavior: %v", urls)
	}
}

func TestCreatorConversationCompositionMarkersAreThreeStaticEvents(t *testing.T) {
	s := New()
	r := compositionTester(s, "markers", 500, true)
	for _, text := range []string{"Project room created", "Design notes added", "Today"} {
		if !r.HasText(text) {
			t.Fatalf("missing visible marker %q", text)
		}
	}
	for _, node := range r.AccessAll() {
		if node.Role == ui.RoleStatus || node.Role == ui.RoleButton || node.Role == ui.RoleLink {
			t.Fatal("a passive marker preview acquired a live or interactive role")
		}
	}
}

func TestCreatorConversationCompositionNativeSendStreamsThenBecomesIdle(t *testing.T) {
	s := New()
	r := compositionTester(s, "chat", 620, false)
	p := &s.ConversationComposition
	before := len(p.messages)
	advancedClick(t, r, "Chat message")
	r.Type("Please explain the reading flow.")
	r.Key(0, ui.KeyEnter)
	if len(p.messages) != before+2 || p.draft != "" || !p.streaming || !r.HasText("Please explain the reading flow.") {
		t.Fatal("native Enter did not add the visible user turn and start its reply")
	}
	if !advancedAccess(t, r, ui.RoleStatus, "Writing a reply…").Busy {
		t.Fatal("writing marker did not expose actual native busy status")
	}
	r.Advance(650 * time.Millisecond)
	partial := p.messages[len(p.messages)-1].text
	if partial == "" || partial == strings.Join(compositionReply, "") || !r.HasText(partial) {
		t.Fatal("reply was replaced with an instant completion instead of rendering timed text")
	}
	compositionFinishReply(t, r, p)
	if got := p.messages[len(p.messages)-1].text; got != strings.Join(compositionReply, "") || !r.HasText(got) {
		t.Fatal("completed reply is missing from the native transcript")
	}
	if !advancedAccess(t, r, ui.RoleButton, "Stop reply").Disabled || !r.Focused("Chat message") {
		t.Fatal("completion left a disabled stop action focused or lost the editor")
	}
	r.Advance(2 * time.Second)
	if r.FramePending() {
		t.Fatal("completed conversation retained animation work")
	}
}

func TestCreatorConversationCompositionNativeIMEDoesNotSendOnCandidateEnter(t *testing.T) {
	s := New()
	r := compositionTester(s, "chat", 620, true)
	p := &s.ConversationComposition
	before := len(p.messages)
	advancedClick(t, r, "Chat message")
	r.Compose("检查", 2)
	r.Key(0, ui.KeyEnter)
	if len(p.messages) != before || p.streaming {
		t.Fatal("choosing an IME candidate sent a message")
	}
	r.Type("阅读位置")
	r.Key(0, ui.KeyEnter)
	if len(p.messages) != before+2 || p.messages[before].text != "检查阅读位置" {
		t.Fatalf("committed native Chinese input did not send exactly once: rows=%d want%d draft=%q sent=%+v", len(p.messages), before+2, p.draft, p.messages[before:])
	}
}

func TestCreatorConversationCompositionManualReadingWinsGrowthAndJumpControlsWork(t *testing.T) {
	s := New()
	r := compositionTester(s, "chat", 620, true)
	p := &s.ConversationComposition
	advancedClick(t, r, "Chat message")
	r.Type("Help me keep my place while reading a longer answer.")
	advancedClick(t, r, "Send message")
	r.Advance(650 * time.Millisecond)
	viewport := advancedAccess(t, r, ui.RoleRegion, "Team conversation").Bounds
	if viewport.H < 330 || viewport.H > 340 {
		t.Fatalf("chat viewport is not actually bounded: %v", viewport)
	}
	r.Scroll(viewport.X+30, viewport.Y+60, 0, -180)
	y := p.scroll.Scroll.Y
	if p.scroll.Mode() != conversation.MessageScrollerFree || y <= 0 {
		t.Fatal("native wheel did not release following inside the real transcript")
	}
	compositionFinishReply(t, r, p)
	if after := advancedAccess(t, r, ui.RoleRegion, "Team conversation").Bounds; after.H != viewport.H {
		t.Fatal("streamed content changed the bounded viewport height")
	}
	if math.Abs(float64(p.scroll.Scroll.Y-y)) > .5 {
		t.Fatalf("streaming pulled the reader away: %v→%v", y, p.scroll.Scroll.Y)
	}
	advancedClick(t, r, "Jump to latest")
	if math.Abs(float64(p.scroll.Scroll.Y-p.scroll.Scroll.MaxY)) > .5 || p.scroll.Mode() != conversation.MessageScrollerFollowing {
		t.Fatal("real jump control did not catch up and resume following")
	}
	advancedClick(t, r, "Back to beginning")
	if p.scroll.Scroll.Y != 0 {
		t.Fatal("source start button did not reach the actual native beginning")
	}
}

func TestCreatorConversationCompositionPrependPreservesVisibleTextAndDraft(t *testing.T) {
	s := New()
	r := compositionTester(s, "chat", 620, true)
	p := &s.ConversationComposition
	advancedClick(t, r, "Chat message")
	r.Type("Keep this draft")
	viewport := advancedAccess(t, r, ui.RoleRegion, "Team conversation").Bounds
	r.Scroll(viewport.X+30, viewport.Y+60, 0, -120)
	var text string
	var before ui.Rect
	for _, message := range p.messages {
		if bounds, ok := r.Find(message.text); ok && bounds.Y > viewport.Y+5 && bounds.Y+bounds.H < viewport.Y+viewport.H-5 {
			text, before = message.text, bounds
			break
		}
	}
	if text == "" {
		t.Fatal("probe found no actual visible message")
	}
	advancedClick(t, r, "Load earlier messages")
	after, ok := r.Find(text)
	if !ok || math.Abs(float64(after.Y-before.Y)) > .5 || advancedAccess(t, r, ui.RoleTextField, "Chat message").Value != "Keep this draft" {
		t.Fatalf("loading earlier context moved visible text or lost the native draft: %v→%v", before, after)
	}
}

func TestCreatorConversationCompositionStopAndPageChangesPreserveApplicationState(t *testing.T) {
	s := New()
	r := compositionTester(s, "chat", 620, true)
	p := &s.ConversationComposition
	advancedClick(t, r, "Chat message")
	r.Type("Start a reply")
	advancedClick(t, r, "Send message")
	r.Advance(650 * time.Millisecond)
	partial := p.messages[len(p.messages)-1].text
	s.Page = "display"
	r.Frame()
	r.Advance(2 * time.Second)
	if p.messages[len(p.messages)-1].text != partial || r.FramePending() {
		t.Fatal("unmounted preview continued generating unseen reply text")
	}
	s.Page = "conversation"
	r.Frame()
	if !r.HasText(partial) || !p.streaming {
		t.Fatal("returning to the page discarded the application-owned partial reply")
	}
	r.Advance(650 * time.Millisecond)
	if p.messages[len(p.messages)-1].text == partial {
		t.Fatal("visible reply did not resume after returning")
	}
	advancedClick(t, r, "Stop reply")
	stopped := p.messages[len(p.messages)-1].text
	r.Advance(2 * time.Second)
	if p.streaming || p.messages[len(p.messages)-1].text != stopped || !r.Focused("Chat message") || r.FramePending() {
		t.Fatal("stop did not end timed work and return focus to the native editor")
	}
}
