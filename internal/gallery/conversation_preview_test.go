package gallery

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
)

func conversationPreviewTester(s *State, part string) *ui.Tester {
	return ui.NewTesterWith(func(c *ui.Context) {
		s.Theme.Apply(c)
		if s.Config.ReduceMotion {
			c.DisableAnimations()
		}
		ui.Column(c).Fill().Padding(24).Gap(20).Children(func() {
			ui.Button(c, "Page focus anchor")
			if s.Page != "conversation" {
				ui.Text(c, "Another preview")
				return
			}
			switch part {
			case "discussion":
				s.discussionPreview(c)
			case "files":
				s.conversationFilesPreview(c)
			case "trace":
				s.conversationTracePreview(c)
			default:
				s.conversation(c)
			}
		})
		if part != "" && s.Page == "conversation" {
			s.conversationPreviewDialogs(c)
		}
	}, 1100, 850, ui.TesterOptions{})
}
func TestCreatorConversationReactionReplyAndNotesPersistAcrossPages(t *testing.T) {
	s := New()
	s.Page = "conversation"
	s.Config.ReduceMotion = true
	r := conversationPreviewTester(s, "discussion")
	advancedClick(t, r, "Helpful 3")
	if !s.conversationPreview.helpful || !advancedAccess(t, r, ui.RoleToggleButton, "Helpful 4").Checked {
		t.Fatal("actual native reaction did not update controlled pressed state")
	}
	advancedClick(t, r, "Conversation reply")
	r.Type("I'll check the keyboard path.")
	if s.conversationPreview.reply != "I'll check the keyboard path." || !r.HasText("I'll check the keyboard path.") {
		t.Fatal("native reply edit did not update the composed ghost bubble")
	}
	advancedClick(t, r, "Open review notes")
	if !s.conversationPreview.notesOpen {
		t.Fatal("native bubble content did not open its sheet")
	}
	advancedClick(t, r, "Review notes editor")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Verify labels and focus in dark mode.")
	advancedClick(t, r, "Save review notes")
	r.Advance(300 * time.Millisecond)
	if s.conversationPreview.notesOpen || s.conversationPreview.reviewNote != "Verify labels and focus in dark mode." {
		t.Fatal("sheet native edit/save callback failed")
	}
	s.Page = "display"
	r.Frame()
	s.Page = "conversation"
	r.Frame()
	if !s.conversationPreview.helpful || !r.HasText("Helpful 4") || s.conversationPreview.reply == "" || s.conversationPreview.reviewNote == "" {
		t.Fatal("conversation state was kept in a discarded frame/local")
	}
}
func TestCreatorAttachmentCancelRetryCompleteAreApplicationStateAndOpenRealDialog(t *testing.T) {
	s := New()
	s.Page = "conversation"
	s.Config.ReduceMotion = true
	r := conversationPreviewTester(s, "files")
	advancedClick(t, r, "Cancel audit-report.pdf")
	if s.conversationPreview.fileState != conversation.AttachmentError || !r.HasText("Transfer canceled. You can retry the report.") {
		t.Fatal("native cancel did not change the application model")
	}
	advancedClick(t, r, "Retry audit-report.pdf")
	if s.conversationPreview.fileState != conversation.AttachmentUploading || s.conversationPreview.fileProgress != .25 {
		t.Fatal("retry did not restart the application progress model")
	}
	advancedClick(t, r, "Complete transfer")
	if s.conversationPreview.fileState != conversation.AttachmentDone || s.conversationPreview.fileProgress != 1 || !r.HasText("Report transfer complete.") {
		t.Fatal("completion did not follow the current application state")
	}
	advancedClick(t, r, "Preview report")
	advancedAccess(t, r, ui.RoleDialog, "Report preview")
	if !r.HasText("audit-report.json") {
		t.Fatal("attachment callback did not compose the native code preview")
	}
	advancedClick(t, r, "Close report preview")
	r.Advance(300 * time.Millisecond)
	if s.conversationPreview.previewOpen {
		t.Fatal("dialog close did not update the controlled model")
	}
	s.Page = "themes"
	r.Frame()
	s.Page = "conversation"
	r.Frame()
	if s.conversationPreview.fileState != conversation.AttachmentDone {
		t.Fatal("file transfer state disappeared across preview pages")
	}
}
func TestCreatorAttachmentFiveStatesSizingOrientationAndRemoveRestore(t *testing.T) {
	s := New()
	s.Page = "conversation"
	s.Config.ReduceMotion = true
	r := conversationPreviewTester(s, "files")
	for _, choice := range []struct{ label, state string }{{"Idle", conversation.AttachmentIdle}, {"Uploading", conversation.AttachmentUploading}, {"Processing", conversation.AttachmentProcessing}, {"Failed", conversation.AttachmentError}, {"Done", conversation.AttachmentDone}} {
		advancedClick(t, r, choice.label)
		if s.conversationPreview.fileState != choice.state {
			t.Fatalf("state %q was only a display value", choice.label)
		}
	}
	advancedClick(t, r, "Vertical files")
	advancedClick(t, r, "Extra small size")
	if s.conversationPreview.orientation() != ui.Vertical || s.conversationPreview.size() != conversation.AttachmentExtraSmall {
		t.Fatal("presentation choices did not reach public attachment props")
	}
	group := advancedAccess(t, r, ui.RoleScroll, "Review attachments")
	if group.Bounds.H != 190 {
		t.Fatalf("vertical strip did not resize actual viewport: %+v", group.Bounds)
	}
	advancedClick(t, r, "Remove audit-report.pdf")
	if !s.conversationPreview.reportRemoved {
		t.Fatal("remove callback failed")
	}
	advancedClick(t, r, "Restore report")
	if s.conversationPreview.reportRemoved || s.conversationPreview.fileState != conversation.AttachmentIdle {
		t.Fatal("restore did not rebuild caller-owned attachment")
	}
	s.Page = "controls"
	r.Frame()
	s.Page = "conversation"
	r.Frame()
	if s.conversationPreview.fileSize != "xs" || s.conversationPreview.fileOrientation != "vertical" {
		t.Fatal("attachment presentation settings did not persist")
	}
}
func TestCreatorTraceNativeEditingStatusesReferenceAndCrossPageState(t *testing.T) {
	s := New()
	s.Page = "conversation"
	s.Config.ReduceMotion = true
	r := conversationPreviewTester(s, "trace")
	advancedClick(t, r, "Review accessibility")
	advancedClick(t, r, "Review note")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Add a visible focus treatment.")
	if s.conversationPreview.reviewNote != "Add a visible focus treatment." {
		t.Fatal("Trace detail used a fake editor or lost its application value")
	}
	advancedClick(t, r, "Report a problem")
	if s.conversationPreview.traceStatus != conversation.Error {
		t.Fatal("error action did not update TraceStatus")
	}
	advancedClick(t, r, "Finish review")
	if s.conversationPreview.traceStatus != conversation.Complete {
		t.Fatal("done action did not update TraceStatus")
	}
	advancedClick(t, r, "Resume review")
	if s.conversationPreview.traceStatus != conversation.Running {
		t.Fatal("resume action did not restore running status")
	}
	advancedClick(t, r, "Open shadcn reference")
	if urls := r.OpenedURLs(); len(urls) != 1 || urls[0] != "https://ui.shadcn.com/create" {
		t.Fatalf("Trace URL bypassed native Link: %v", urls)
	}
	s.Page = "text-layout"
	r.Frame()
	s.Page = "conversation"
	r.Frame()
	if !s.conversationPreview.trace.Expanded["accessibility"] || advancedAccess(t, r, ui.RoleTextField, "Review note").Value != "Add a visible focus treatment." {
		t.Fatal("Trace disclosure/editor state did not survive preview navigation")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("reduced conversation preview left shader or disclosure animation running")
	}
}

func TestCreatorConversationUnmountClosesTransientModalsWithoutReplayingOldFocus(t *testing.T) {
	for _, part := range []string{"discussion", "files"} {
		t.Run(part, func(t *testing.T) {
			s := New()
			s.Page = "conversation"
			s.Config.ReduceMotion = true
			r := conversationPreviewTester(s, part)
			opener := "Open review notes"
			if part == "files" {
				opener = "Preview report"
				s.conversationPreview.fileState = conversation.AttachmentDone
				r.Frame()
			}
			advancedClick(t, r, opener)
			s.Page = "display"
			r.Frame()
			if s.conversationPreview.notesOpen || s.conversationPreview.previewOpen || s.conversationPreview.notesReturn != nil {
				t.Fatal("page teardown retained a transient modal or its unmounted return target")
			}
			advancedClick(t, r, "Page focus anchor")
			r.Advance(300 * time.Millisecond)
			s.Page = "conversation"
			r.Frame()
			r.Advance(300 * time.Millisecond)
			if !r.Focused("Page focus anchor") {
				t.Fatal("rebinding a previous page opener replayed a stale modal focus request")
			}
			if _, ok := r.AccessByRole(ui.RoleDialog, "Review notes"); ok {
				t.Fatal("leaving and returning reopened the notes sheet")
			}
			if _, ok := r.AccessByRole(ui.RoleDialog, "Report preview"); ok {
				t.Fatal("leaving and returning reopened the report dialog")
			}
		})
	}
}

func TestCreatorConversationIsReachableByComponentSearch(t *testing.T) {
	for _, query := range []string{"Bubble", "Attachment", "Trace", "Message", "Questionnaire"} {
		t.Run(query, func(t *testing.T) {
			s := New()
			s.Config.ReduceMotion = true
			r := ui.NewTester(s.View, 1280, 900)
			advancedClick(t, r, "Search components")
			r.Advance(200 * time.Millisecond)
			r.Type(query)
			r.Key(0, ui.KeyEnter)
			r.Advance(200 * time.Millisecond)
			if s.Page != "conversation" || s.Command || !r.HasText("Project discussion") {
				t.Fatalf("%s search did not reach the real conversation Gallery", query)
			}
		})
	}
}

func TestCreatorConversationNormalModalCloseRestoresMountedOpener(t *testing.T) {
	for _, path := range []string{"bubble sheet", "trace sheet", "horizontal report", "vertical report"} {
		t.Run(path, func(t *testing.T) {
			s := New()
			s.Page, s.Config.ReduceMotion = "conversation", true
			part, opener, closer := "discussion", "Open review notes", "Save review notes"
			if path == "trace sheet" {
				part, opener = "trace", "Open notes sheet"
			} else if path == "horizontal report" || path == "vertical report" {
				part, opener, closer = "files", "Preview report", "Close report preview"
			}
			r := conversationPreviewTester(s, part)
			if path == "trace sheet" {
				advancedClick(t, r, "Review accessibility")
			} else if part == "files" {
				advancedClick(t, r, "Complete transfer")
				if path == "vertical report" {
					advancedClick(t, r, "Vertical files")
				}
			}
			advancedClick(t, r, opener)
			advancedClick(t, r, closer)
			r.Advance(300 * time.Millisecond)
			if !r.Focused(opener) {
				t.Fatalf("normal modal close lost still-mounted native opener %q", opener)
			}
		})
	}
}

func TestCreatorConversationTransientOwnerDuringExitDoesNotReopenOrStealFocus(t *testing.T) {
	for _, part := range []string{"discussion", "trace", "files"} {
		t.Run(part, func(t *testing.T) {
			s := New()
			s.Page = "conversation"
			r := conversationPreviewTester(s, part)
			opener, closer := "Open review notes", "Save review notes"
			if part == "trace" {
				advancedClick(t, r, "Review accessibility")
				r.Advance(250 * time.Millisecond)
				opener = "Open notes sheet"
			} else if part == "files" {
				advancedClick(t, r, "Complete transfer")
				opener, closer = "Preview report", "Close report preview"
			}
			advancedClick(t, r, opener)
			r.Advance(250 * time.Millisecond)
			advancedClick(t, r, closer)
			r.Advance(25 * time.Millisecond)
			s.Page = "controls"
			r.Frame()
			advancedClick(t, r, "Page focus anchor")
			r.Advance(400 * time.Millisecond)
			s.Page = "conversation"
			r.Frame()
			r.Advance(400 * time.Millisecond)
			if !r.Focused("Page focus anchor") || s.conversationPreview.notesOpen || s.conversationPreview.previewOpen {
				t.Fatal("leaving during real exit replayed a modal or pending opener focus")
			}
		})
	}
}

func conversationReveal(t *testing.T, r *ui.Tester, label string) ui.Rect {
	t.Helper()
	for attempt := 0; attempt < 4; attempt++ {
		target, ok := r.Find(label)
		if !ok {
			t.Fatalf("missing real Gallery label %q", label)
		}
		panel, ok := r.AccessByRole(ui.RoleRegion, "Messages & questions")
		if !ok {
			t.Fatal("missing actual Gallery scroll panel")
		}
		if target.Y >= panel.Bounds.Y && target.Y+target.H <= panel.Bounds.Y+panel.Bounds.H {
			return target
		}
		delta := target.Y + target.H/2 - panel.Bounds.Y - panel.Bounds.H/2
		r.Scroll(panel.Bounds.X+panel.Bounds.W-12, panel.Bounds.Y+panel.Bounds.H/2, 0, delta)
	}
	target, _ := r.Find(label)
	t.Fatalf("native Gallery viewport could not reveal %q at%v", label, target)
	return ui.Rect{}
}

func TestCreatorConversationWholePageNarrowCardsKeepActionsInside(t *testing.T) {
	for _, width := range []int{500, 1100, 1450} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			s := New()
			s.Page, s.Config.ReduceMotion = "conversation", true
			r := ui.NewTester(s.View, width, 900)
			conversationReveal(t, r, "Vertical files")
			advancedClick(t, r, "Vertical files")
			conversationReveal(t, r, "Remove audit-report.pdf")
			card, ok := r.AccessByRole(ui.RoleGroup, "audit-report.pdf")
			if !ok || card.Bounds.W != 120 {
				t.Fatalf("source vertical card was not120: %+v", card)
			}
			for _, label := range []string{"Cancel audit-report.pdf", "Remove audit-report.pdf"} {
				button, ok := r.AccessByRole(ui.RoleButton, label)
				if !ok || button.Bounds.X < card.Bounds.X || button.Bounds.X+button.Bounds.W > card.Bounds.X+card.Bounds.W || button.Bounds.Y < card.Bounds.Y || button.Bounds.Y+button.Bounds.H > card.Bounds.Y+card.Bounds.H {
					t.Fatalf("real vertical action escaped120 card: %q card=%v button=%v", label, card.Bounds, button.Bounds)
				}
			}
			advancedClick(t, r, "Cancel audit-report.pdf")
			conversationReveal(t, r, "Retry audit-report.pdf")
			advancedClick(t, r, "Retry audit-report.pdf")
			conversationReveal(t, r, "Preview report")
			advancedClick(t, r, "Preview report")
			advancedAccess(t, r, ui.RoleDialog, "Report preview")
			advancedClick(t, r, "Close report preview")
			if !r.Focused("Preview report") {
				t.Fatal("whole-page vertical preview did not restore external native opener")
			}
		})
	}
}

func TestCreatorConversationActualPageTraceIMEAndStateSurviveNavigation(t *testing.T) {
	s := New()
	s.Page, s.Config.ReduceMotion = "conversation", true
	r := ui.NewTester(s.View, 1100, 900)
	conversationReveal(t, r, "Review accessibility")
	advancedClick(t, r, "Review accessibility")
	conversationReveal(t, r, "Review note")
	advancedClick(t, r, "Review note")
	r.Key(ui.Cmd, ui.KeyA)
	r.Compose("文", 1)
	r.Compose("文档", 2)
	r.Type("文档")
	if s.conversationPreview.reviewNote != "文档" {
		t.Fatalf("actual full-page native composition failed: %q", s.conversationPreview.reviewNote)
	}
	s.Page = "themes"
	r.Frame()
	s.Page = "conversation"
	r.Frame()
	conversationReveal(t, r, "Review note")
	if advancedAccess(t, r, ui.RoleTextField, "Review note").Value != "文档" || !s.conversationPreview.trace.Expanded["accessibility"] || s.conversationPreview.fileState != conversation.AttachmentUploading {
		t.Fatal("actual page navigation discarded source widget model state")
	}
	if r.FramePending() {
		t.Fatal("reduced whole-page widgets did not become idle")
	}
}
