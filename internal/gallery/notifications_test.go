package gallery

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func notificationPreviewTester(s *State) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		s.Theme.Apply(c)
		c.DisableAnimations()
		ui.Column(c).Fill().Padding(24).Children(func() {
			if s.Page == "overlays" {
				s.notificationsPreview(c)
			} else {
				ui.Text(c, "Another preview")
			}
		})
		// The application-owned viewport stays mounted across preview pages.
		s.notificationViewport(c)
	}, 1100, 800)
}

func TestCreatorNotificationActionsCanKeepOpenAndCancel(t *testing.T) {
	s := New()
	s.Page = "overlays"
	r := notificationPreviewTester(s)
	advancedClick(t, r, "Success")
	advancedClick(t, r, "Keep open after action")
	advancedClick(t, r, "Show notification")
	advancedClick(t, r, "Review")
	if s.toastResult != "Notification action selected" || !r.HasText("Workspace saved") {
		t.Fatal("actual notification action did not run or prevent its dismissal")
	}
	advancedClick(t, r, "Later")
	r.Advance(250 * time.Millisecond)
	if s.toastResult != "Notification postponed" || r.HasText("Workspace saved") {
		t.Fatal("cancel did not run and remove the preview notification")
	}
}

func TestCreatorUploadUpdatesSameNotificationAndSurvivesPageChanges(t *testing.T) {
	s := New()
	s.Page = "overlays"
	r := notificationPreviewTester(s)
	advancedClick(t, r, "Start upload")
	id := s.toastUploadID
	r.Advance(10 * time.Second)
	if id == "" || !r.HasText("Uploading your file") {
		t.Fatal("loading notification expired or never rendered")
	}
	s.Page = "text-layout"
	r.Frame()
	if !r.HasText("Uploading your file") {
		t.Fatal("changing preview pages removed an app-owned notification")
	}
	s.Page = "overlays"
	r.Frame()
	advancedClick(t, r, "Complete upload")
	toasts := s.toastPreview.Toasts()
	if !r.HasText("Upload complete") || r.HasText("Uploading your file") || len(toasts) != 1 || toasts[0].ID != id || toasts[0].Type != "success" {
		t.Fatal("completion added a second notification instead of updating the loading toast")
	}
	if s.toastResult != "Upload notification updated" || !advancedAccess(t, r, ui.RoleButton, "Complete upload").Disabled {
		t.Fatal("completion status/action availability did not follow the update")
	}
	r.Advance(4250 * time.Millisecond)
	if r.HasText("Upload complete") {
		t.Fatal("completed loading notification did not resume normal expiry")
	}
}

func TestCreatorNotificationPositionChangesActualViewport(t *testing.T) {
	for _, choice := range []string{"Top left", "Top center", "Top right", "Bottom left", "Bottom center", "Bottom right"} {
		t.Run(choice, func(t *testing.T) {
			s := New()
			s.Page = "overlays"
			r := notificationPreviewTester(s)
			advancedClick(t, r, choice)
			advancedClick(t, r, "Show notification")
			bounds, ok := r.Find("Event created")
			if !ok {
				t.Fatal("notification did not render")
			}
			if strings.HasPrefix(choice, "Top") && bounds.Y > 200 || strings.HasPrefix(choice, "Bottom") && bounds.Y < 500 {
				t.Fatalf("vertical position did not reach the native viewport: %+v", bounds)
			}
			if strings.HasSuffix(choice, "left") && bounds.X > 200 || strings.HasSuffix(choice, "center") && (bounds.X < 300 || bounds.X > 600) || strings.HasSuffix(choice, "right") && bounds.X < 650 {
				t.Fatalf("horizontal position did not reach the native viewport: %+v", bounds)
			}
		})
	}
}

func TestCreatorNotificationSettingsAndClearPersistAcrossPages(t *testing.T) {
	s := New()
	s.Page = "overlays"
	r := notificationPreviewTester(s)
	advancedClick(t, r, "Rich notification colors")
	advancedClick(t, r, "Show notification close button")
	advancedClick(t, r, "Expand notifications")
	advancedClick(t, r, "Stack three updates")
	if len(s.toastPreview.Toasts()) != 3 || !r.HasText("Workspace update 1") || !r.HasText("Workspace update 3") {
		t.Fatal("stack action did not create actual notifications")
	}
	s.Page = "text-layout"
	r.Frame()
	s.Page = "overlays"
	r.Frame()
	if !s.toastRich || !s.toastClose || !s.toastExpand {
		t.Fatal("leaving preview pages lost notification settings")
	}
	advancedClick(t, r, "Clear preview notifications")
	r.Advance(250 * time.Millisecond)
	if len(s.toastPreview.Toasts()) != 0 || r.HasText("Workspace update 1") || r.HasText("Workspace update 3") {
		t.Fatal("clear preview left notifications visible or queued")
	}
}

func TestCreatorNotificationHotkeyReturnsToEnabledTrigger(t *testing.T) {
	s := New()
	s.Page = "overlays"
	r := notificationPreviewTester(s)
	advancedClick(t, r, "Stack three updates")
	r.Key(ui.Alt, ui.KeyT)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyEscape)
	if !r.Focused("Stack three updates") || len(s.toastPreview.Toasts()) != 3 {
		t.Fatal("notification shortcut did not return to its actual trigger or Escape dismissed the stack")
	}
	advancedClick(t, r, "Clear preview notifications")
	advancedClick(t, r, "Start upload")
	advancedClick(t, r, "Complete upload")
	r.Key(ui.Alt, ui.KeyT)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyEscape)
	if !r.Focused("Start upload") || !r.HasText("Upload complete") {
		t.Fatal("completed upload returned focus to its disabled action instead of an enabled trigger")
	}
}

func TestCreatorNotificationViewportUsesWindowBoundsInsidePreviewScope(t *testing.T) {
	s := New()
	s.Page = "overlays"
	s.toastPreview.Show(ui.Toast{Title: "Window bounds", Timeout: -1})
	r := ui.NewTesterWith(s.View, 1280, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	notice := advancedAccess(t, r, ui.RoleListItem, "Window bounds")
	if notice.Bounds.X+notice.Bounds.W != 1256 || notice.Bounds.Y+notice.Bounds.H != 876 {
		t.Fatalf("nested preview scope changed window-edge offsets: %+v", notice.Bounds)
	}
}

func TestCreatorNotificationReturnFocusFollowsCurrentPreviewPage(t *testing.T) {
	s := New()
	s.Page = "overlays"
	r := ui.NewTesterWith(s.View, 1280, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	advancedClick(t, r, "Start upload")
	advancedClick(t, r, "Text & layout")
	r.Key(ui.Alt, ui.KeyT)
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyEscape)
	if !r.Focused("Text & layout") || !r.HasText("Uploading your file") {
		t.Fatal("notification returned to an unmounted trigger instead of the current preview tab")
	}
	advancedClick(t, r, "Overlays")
	if r.Focused("Start upload") {
		t.Fatal("returning to Overlays replayed a stale trigger focus request")
	}
	r.Key(0, ui.KeyTab)
	if !r.Focused("Edit profile") {
		t.Fatal("returning to Overlays replayed a stale trigger focus request")
	}
}

func TestCreatorNotificationClearAlsoRemovesLegacyPreview(t *testing.T) {
	s := New()
	s.Page = "overlays"
	r := ui.NewTesterWith(s.View, 1280, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	advancedClick(t, r, "Show toast")
	advancedClick(t, r, "Start upload")
	if !r.HasText("Event created") || !r.HasText("Uploading your file") {
		t.Fatal("mixed preview queues were not displayed")
	}
	advancedClick(t, r, "Clear preview notifications")
	r.Advance(250 * time.Millisecond)
	if r.HasText("Event created") || r.HasText("Uploading your file") {
		t.Fatal("clear preview left a legacy or application notification visible")
	}
}
