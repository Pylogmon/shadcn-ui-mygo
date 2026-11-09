package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func bookingAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(role, name)
	if !ok || n.Offscreen {
		t.Fatalf("missing booking control %v %q", role, name)
	}
	return n
}
func bookingClick(t *testing.T, r *ui.Tester, role ui.Role, name string) {
	t.Helper()
	n := bookingAccess(t, r, role, name)
	r.ClickAt(n.Bounds.X+n.Bounds.W/2, n.Bounds.Y+n.Bounds.H/2)
	r.Frame()
}
func bookingType(t *testing.T, r *ui.Tester, name, text string) {
	t.Helper()
	bookingClick(t, r, ui.RoleTextField, name)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type(text)
	r.Frame()
}
func bookingWait(t *testing.T, r *ui.Tester, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !ready() && time.Now().Before(deadline) {
		r.Frame()
		time.Sleep(time.Millisecond)
	}
	if !ready() {
		t.Fatal("booking picker result was not applied")
	}
}
func bookingSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	if directory := os.Getenv("GALLERY_BOOKING_SNAPSHOT_DIR"); directory != "" {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(directory, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, r.Image())
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			t.Fatalf("booking snapshot: %v/%v", err, closeErr)
		}
	}
}

func TestBookingInputPreviewWorkflow(t *testing.T) {
	p := &typedInputPreviewState{}
	show, dark := true, false
	requests := make(chan mygo.OpenDialogOptions, 3)
	replies := make(chan []string)
	completed := make(chan struct{}, 3)
	defer close(replies)
	p.picker = func(options mygo.OpenDialogOptions) ([]string, error) {
		requests <- options
		paths := <-replies
		completed <- struct{}{}
		return paths, nil
	}
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(24).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			if show {
				p.view(c, nil)
			} else {
				ui.Text(c, "Browsing another Gallery page")
			}
		})
	}, 1000, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	r.SetRegion(i18n.Region{Locale: "en-US", HourCycle: 24, FirstWeekday: time.Monday})
	select {
	case <-requests:
		t.Fatal("file dialog launched without user activation")
	default:
	}
	bookingSnapshot(t, r, "booking-light")
	if err := r.Click("$"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("249.5")
	r.Key(0, ui.KeyEnter)
	if p.amount != 249.5 || p.feedback != "Budget updated to $249.50." {
		t.Fatalf("real amount editor/submit: %g %q", p.amount, p.feedback)
	}
	bookingType(t, r, "Booking amount", "-")
	r.Key(0, ui.KeyEnter)
	if p.amount != 249.5 || bookingAccess(t, r, ui.RoleTextField, "Booking amount").Value != "249.5" {
		t.Fatal("invalid numeric draft did not restore the accepted amount")
	}
	before := p.date
	bookingClick(t, r, ui.RolePopUpButton, "Booking date")
	r.Key(0, ui.KeyEscape)
	if p.date != before {
		t.Fatal("canceling the native calendar changed departure")
	}
	bookingClick(t, r, ui.RolePopUpButton, "Booking date")
	bookingClick(t, r, ui.RoleButton, "November 12, 2026")
	bookingClick(t, r, ui.RoleStepper, "Booking time hours")
	r.Key(0, ui.KeyUp)
	if p.date.Day() != 12 || p.date.Hour() != 16 || p.date.Minute() != 30 || p.date.Location() != before.Location() || p.draft.Text != "November 12, 2026" {
		t.Fatalf("shared native date/time changed civil values: %v %+v", p.date, p.draft)
	}
	bookingType(t, r, "Departure query", "2026-12-04")
	r.Key(0, ui.KeyEnter)
	if p.date.Month() != time.December || p.date.Day() != 4 || p.date.Hour() != 16 || p.feedback != "Departure date confirmed." {
		t.Fatalf("application parse/submit: %v %q", p.date, p.feedback)
	}
	before = p.date
	month := p.draft.Month
	bookingType(t, r, "Departure query", "tomorrow")
	if p.date != before || p.draft.Month != month || p.draft.Text != "tomorrow" || p.draft.Error == "" || !bookingAccess(t, r, ui.RoleTextField, "Departure query").Invalid || !bookingAccess(t, r, ui.RoleButton, "Save booking").Disabled {
		t.Fatalf("invalid exact-date draft did not preserve accepted booking: %+v", p.draft)
	}
	bookingClick(t, r, ui.RoleButton, "Restore departure")
	if p.date != before || p.draft.Error != "" || p.draft.Text != "December 04, 2026" {
		t.Fatal("restore action did not recover accepted departure")
	}
	bookingType(t, r, "Departure query", "2026-12-04")
	r.Key(0, ui.KeyDown)
	r.Frame()
	bookingClick(t, r, ui.RoleButton, "Next month")
	if p.date != before || p.draft.Text != "2026-12-04" {
		t.Fatal("calendar browsing changed accepted date or raw draft")
	}
	bookingClick(t, r, ui.RoleButton, "January 5, 2027")
	if p.date.Year() != 2027 || p.date.Month() != time.January || p.date.Day() != 5 || p.date.Hour() != 16 || p.draft.Open || p.draft.Text != "January 05, 2027" || !r.Focused("Choose departure date") {
		t.Fatalf("calendar selection did not publish text/date/focus: %v %+v", p.date, p.draft)
	}
	before = p.date
	bookingClick(t, r, ui.RoleMenuButton, "Departure calendar")
	if err := r.ChooseMenuItem("Persian / Jalali"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if p.calendar != "persian" || p.date != before {
		t.Fatal("calendar model switch changed the real departure date")
	}
	bookingClick(t, r, ui.RoleTextField, "Departure query")
	r.Key(0, ui.KeyDown)
	r.Frame()
	if !p.draft.Open || !r.HasText("دی ۱۴۰۵") {
		t.Fatal("selected model did not reach the real popup calendar")
	}
	r.Key(0, ui.KeyLeft)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !p.date.Equal(before.AddDate(0, 0, 1)) || p.draft.Open || p.draft.Text != "January 06, 2027" {
		t.Fatalf("RTL Persian day selection did not share date/text: %v %+v", p.date, p.draft)
	}
	bookingClick(t, r, ui.RoleButton, "Booking attachment")
	select {
	case options := <-requests:
		if options.Parent != nil || !options.Multiple || options.Title != "Attach booking documents" || len(options.Filters) != 1 || !slices.Equal(options.Filters[0].Extensions, []string{"txt", "pdf", "csv"}) {
			t.Fatalf("value-only native picker options: %+v", options)
		}
	case <-time.After(time.Second):
		t.Fatal("file button did not activate picker")
	}
	replies <- []string{"/tmp/itinerary.pdf", "/tmp/budget.csv"}
	bookingWait(t, r, func() bool { return len(p.paths) == 2 })
	if !r.HasText("Documents: itinerary.pdf, budget.csv") {
		t.Fatal("chosen documents have no readable booking result")
	}
	bookingClick(t, r, ui.RoleButton, "Booking attachment")
	select {
	case <-requests:
	case <-time.After(time.Second):
		t.Fatal("second file activation missing")
	}
	replies <- nil
	bookingWait(t, r, func() bool { return !bookingAccess(t, r, ui.RoleButton, "Booking attachment").Busy })
	if len(p.paths) != 2 {
		t.Fatal("picker cancel removed accepted attachments")
	}
	bookingClick(t, r, ui.RoleButton, "Booking attachment")
	select {
	case <-requests:
	case <-time.After(time.Second):
		t.Fatal("pending file activation missing")
	}
	show = false
	r.Frame()
	replies <- []string{"/tmp/stale.pdf"}
	for range 3 {
		select {
		case <-completed:
		case <-time.After(time.Second):
			t.Fatal("pending picker did not return")
		}
	}
	show = true
	r.Frame()
	if p.amount != 249.5 || p.date.Day() != 6 || p.calendar != "persian" || !slices.Equal(p.paths, []string{"/tmp/itinerary.pdf", "/tmp/budget.csv"}) || p.draft.Text != "January 06, 2027" {
		t.Fatal("page remount lost application state or applied stale dialog work")
	}
	bookingClick(t, r, ui.RoleButton, "Save booking")
	if !strings.HasPrefix(p.saved, "$249.50 · 2027-01-06 16:30") || !strings.HasSuffix(p.saved, "2 attachment(s)") || !r.HasText("Booking saved: "+p.saved) {
		t.Fatalf("saved booking is not connected to accepted values: %q", p.saved)
	}
	dark = true
	r.Frame()
	bookingClick(t, r, ui.RoleButton, "Choose departure date")
	bookingSnapshot(t, r, "booking-dark-calendar")
}
