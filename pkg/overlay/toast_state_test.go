package overlay

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestToasterStateGeneratedIDSkipsExplicitIDs(t *testing.T) {
	var s ToasterState
	explicit := s.Show(ui.Toast{ID: "sonner-2", Title: "Explicit"})
	automatic := s.Show(ui.Toast{Title: "Automatic"})
	if automatic == explicit {
		t.Fatalf("automatic ID %q collided with an existing explicit ID", automatic)
	}
	s.Dismiss(automatic)
	models := s.Toasts()
	if len(models) != 1 || models[0].ID != explicit {
		t.Fatalf("dismissing generated notification affected explicit notification: %v", models)
	}
}

func TestToasterStateReopensSameIDBeforeAndDuringRemoval(t *testing.T) {
	for _, duringRemoval := range []bool{false, true} {
		t.Run(map[bool]string{false: "pending", true: "closing"}[duringRemoval], func(t *testing.T) {
			var s ToasterState
			closed := 0
			s.Show(ui.Toast{ID: "same", Title: "Old", Timeout: -1, OnClose: func() { closed++ }})
			r := ui.NewTester(func(c *ui.Context) { s.advance(c, ToasterProps{}, false) }, 300, 100)
			identity := s.entries[0]
			key := identity.key
			s.Dismiss("same")
			if duringRemoval {
				r.Frame()
			}
			s.Show(ui.Toast{ID: "same", Title: "New", Timeout: -1})
			r.Frame()
			r.Advance(time.Second)
			if len(s.entries) != 1 || s.entries[0] != identity || s.entries[0].key != key {
				t.Fatalf("same-ID reopen lost its entry identity or created a duplicate: %v", s.entries)
			}
			models := s.Toasts()
			if len(models) != 1 || models[0].Title != "New" {
				t.Fatalf("reopened model disappeared: %v", models)
			}
			wantClosed := 0
			if duringRemoval {
				wantClosed = 1
			}
			if closed != wantClosed {
				t.Fatalf("old close callback ran %d times, want %d", closed, wantClosed)
			}
		})
	}
}

func TestToasterStateUpdateAnnouncesChangedText(t *testing.T) {
	var s ToasterState
	s.Show(ui.Toast{ID: "upload", Title: "Uploading", Type: "loading"})
	r := ui.NewTester(func(c *ui.Context) { s.advance(c, ToasterProps{}, false) }, 300, 100)
	r.Announcements()
	if !s.Update("upload", ui.Toast{Title: "Upload complete", Description: "3 files saved", Type: "success"}) {
		t.Fatal("existing notification was not updated")
	}
	r.Frame()
	got := strings.Join(r.Announcements(), " ")
	if !strings.Contains(got, "Upload complete") || !strings.Contains(got, "3 files saved") {
		t.Fatalf("updated text was not announced: %q", got)
	}
}

func TestToasterStateRemovalReleasesEntryReferences(t *testing.T) {
	var s ToasterState
	s.Show(ui.Toast{ID: "old", Title: "Old", Timeout: -1, OnClose: func() {}})
	r := ui.NewTester(func(c *ui.Context) { s.advance(c, ToasterProps{}, false) }, 300, 100)
	backing := s.entries[:cap(s.entries)]
	s.Dismiss("old")
	r.Frame()
	r.Advance(250 * time.Millisecond)
	if len(s.entries) != 0 {
		t.Fatal("closed notification was not removed")
	}
	if backing[0] != nil {
		t.Fatal("removed notification still retained its callbacks in the slice backing array")
	}
}

func TestToasterStateCloseCallbackCanReopenSameID(t *testing.T) {
	var s ToasterState
	closed := 0
	s.Show(ui.Toast{ID: "same", Title: "Old", Timeout: time.Second, OnClose: func() {
		closed++
		s.Show(ui.Toast{ID: "same", Title: "New", Timeout: -1})
	}})
	r := ui.NewTester(func(c *ui.Context) { s.advance(c, ToasterProps{}, false) }, 300, 100)
	r.Advance(time.Second)
	r.Advance(250 * time.Millisecond)
	r.Advance(time.Second)
	models := s.Toasts()
	if closed != 1 || len(models) != 1 || models[0].Title != "New" {
		t.Fatalf("close callback reentry replayed or lost the new notification: closed=%d models=%v", closed, models)
	}
}

func TestToasterStatePauseKeepsRemainingDuration(t *testing.T) {
	var s ToasterState
	pause := false
	s.Show(ui.Toast{ID: "saved", Title: "Saved", Timeout: time.Second})
	r := ui.NewTester(func(c *ui.Context) { s.advance(c, ToasterProps{}, pause) }, 300, 100)
	r.Advance(500 * time.Millisecond)
	pause = true
	r.Frame()
	r.Advance(10 * time.Second)
	if len(s.Toasts()) != 1 {
		t.Fatal("paused notification expired")
	}
	pause = false
	r.Frame()
	r.Advance(450 * time.Millisecond)
	if len(s.Toasts()) != 1 {
		t.Fatal("remaining duration was discarded")
	}
	r.Advance(60 * time.Millisecond)
	if len(s.Toasts()) != 0 {
		t.Fatal("notification did not expire after its remaining duration")
	}
}

func TestToasterStateViewportRemountStartsFreshLifetime(t *testing.T) {
	var s ToasterState
	mounted := true
	s.Show(ui.Toast{ID: "saved", Title: "Saved", Timeout: time.Second})
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		if mounted {
			ToasterWith(c, ToasterProps{State: &s})
		}
	}, 640, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Advance(750 * time.Millisecond)
	mounted = false
	r.Frame()
	r.Advance(5 * time.Second)
	mounted = true
	r.Frame()
	if len(s.Toasts()) != 1 {
		t.Fatal("unmounted wall-clock time expired the caller-owned notification on remount")
	}
	r.Advance(900 * time.Millisecond)
	if len(s.Toasts()) != 1 {
		t.Fatal("remounted notification did not receive a fresh lifetime")
	}
	r.Advance(110 * time.Millisecond)
	if len(s.Toasts()) != 0 {
		t.Fatal("remounted notification did not expire after its fresh lifetime")
	}
}

func TestToasterStatePendingDismissSurvivesViewportRemount(t *testing.T) {
	var s ToasterState
	mounted := true
	closed, dismissed := 0, 0
	s.Show(ui.Toast{ID: "old", Title: "Old", Timeout: -1, OnClose: func() { closed++ }}, ToastOptions{
		OnDismiss: func(ui.Toast) { dismissed++ },
	})
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		if mounted {
			ToasterWith(c, ToasterProps{State: &s})
		}
	}, 640, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Announcements()
	mounted = false
	r.Frame()
	s.Dismiss("old")
	r.Advance(time.Second)
	if closed != 0 || dismissed != 0 {
		t.Fatal("unmounted viewport executed callbacks without a live build")
	}
	mounted = true
	r.Frame()
	r.Advance(time.Second)
	if closed != 1 || dismissed != 1 || len(s.Toasts()) != 0 {
		t.Fatalf("pending dismissal replayed or was lost on remount: close=%d dismiss=%d models=%v", closed, dismissed, s.Toasts())
	}
	if got := r.Announcements(); len(got) != 0 {
		t.Fatalf("already dismissed notification was announced on remount: %q", got)
	}
}

func TestToasterStateNativeLoadingUpdateAnnouncesCompletionOnce(t *testing.T) {
	var s ToasterState
	s.Show(ui.Toast{ID: "upload", Title: "Uploading", Type: "loading"})
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ToasterWith(c, ToasterProps{State: &s})
	}, 640, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.Announcements()
	s.Update("upload", ui.Toast{Title: "Upload complete", Description: "3 files saved", Type: "success"})
	r.Frame()
	got := strings.Join(r.Announcements(), " ")
	if strings.Count(got, "Upload complete") != 1 || strings.Count(got, "3 files saved") != 1 {
		t.Fatalf("native update did not announce the completion exactly once: %q", got)
	}
	r.Frame()
	if got := r.Announcements(); len(got) != 0 {
		t.Fatalf("idle build replayed the completion announcement: %q", got)
	}
}

func TestToasterStateUpdateKeepsOrderOptionsAndIndependentSnapshot(t *testing.T) {
	var s ToasterState
	closeButton, dismissible := false, false
	s.Show(ui.Toast{ID: "first", Title: "First", Timeout: -1}, ToastOptions{
		Position: ToastTopLeft, CloseButton: &closeButton, Dismissible: &dismissible, Cancel: "Keep editing",
	})
	s.Show(ui.Toast{ID: "second", Title: "Second", Timeout: -1})
	first := s.entries[0]
	if !s.Update("first", ui.Toast{ID: "ignored", Title: "Updated", Timeout: -1}) {
		t.Fatal("existing ID was not updated")
	}
	models := s.Toasts()
	if len(models) != 2 || models[0].ID != "first" || models[0].Title != "Updated" || models[1].ID != "second" || s.entries[0] != first {
		t.Fatalf("update changed the ID, queue position, or identity: %v", models)
	}
	options := first.options
	if options.Position != ToastTopLeft || options.Cancel != "Keep editing" || toastBool(options.CloseButton, true) || toastBool(options.Dismissible, true) {
		t.Fatalf("omitted options did not retain explicit false overrides: %+v", options)
	}
	models[0].Title = "Changed snapshot"
	if got := s.Toasts()[0].Title; got != "Updated" {
		t.Fatalf("changing the returned snapshot changed live state: %q", got)
	}
	if s.Update("missing", ui.Toast{Title: "Unexpected"}) {
		t.Fatal("updating a missing ID created a notification")
	}
	s.Update("first", ui.Toast{Title: "Cleared options", Timeout: -1}, ToastOptions{})
	if first.options.Cancel != "" || first.options.Position != "" {
		t.Fatal("explicit empty options did not clear prior overrides")
	}
}

func TestToasterStateDismissAllReentrantShowPreservesFreshNotification(t *testing.T) {
	var s ToasterState
	closed, dismissed, automatic := 0, 0, 0
	s.Show(ui.Toast{ID: "first", Title: "First", Timeout: -1, OnClose: func() {
		closed++
		s.Show(ui.Toast{ID: "fresh", Title: "Fresh", Timeout: -1})
	}}, ToastOptions{OnDismiss: func(ui.Toast) { dismissed++ }, OnAutoClose: func(ui.Toast) { automatic++ }})
	s.Show(ui.Toast{ID: "second", Title: "Second", Timeout: -1, OnClose: func() { closed++ }}, ToastOptions{
		OnDismiss: func(ui.Toast) { dismissed++ }, OnAutoClose: func(ui.Toast) { automatic++ },
	})
	r := ui.NewTester(func(c *ui.Context) { s.advance(c, ToasterProps{}, false) }, 300, 100)
	s.Dismiss("")
	r.Frame()
	r.Advance(time.Second)
	models := s.Toasts()
	if closed != 2 || dismissed != 2 || automatic != 0 || len(models) != 1 || models[0].ID != "fresh" {
		t.Fatalf("dismiss-all captured future notifications or replayed callbacks: close=%d dismiss=%d auto=%d models=%v", closed, dismissed, automatic, models)
	}
}

func TestToasterStateSameIDUpdateKeepsNativeActionFocus(t *testing.T) {
	var s ToasterState
	s.Show(ui.Toast{ID: "saved", Title: "Saved", Timeout: -1, Action: "Undo"}, ToastOptions{
		Action: func(event *ToastActionEvent) { event.PreventDefault() },
	})
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ToasterWith(c, ToasterProps{State: &s})
	}, 640, 300, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	if !r.Focused("Undo") {
		t.Fatal("actual native action did not receive focus")
	}
	s.Show(ui.Toast{ID: "saved", Title: "Updated", Timeout: -1, Action: "Undo"})
	r.Frame()
	if !r.Focused("Undo") {
		t.Fatal("same-ID update replaced the focused native action")
	}
	if len(s.Toasts()) != 1 || s.Toasts()[0].Title != "Updated" {
		t.Fatal("same-ID update did not replace the actual model")
	}
}
