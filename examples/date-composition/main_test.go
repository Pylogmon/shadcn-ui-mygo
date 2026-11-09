package main

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func dateTester(s *appState) *ui.Tester {
	s.appearance.ReduceMotion = true
	return ui.NewTesterWith(s.view, 1000, 1000, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
}
func dateAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(role, name)
	if !ok || n.Offscreen {
		t.Fatalf("missing native %v %q", role, name)
	}
	return n
}
func dateClick(t *testing.T, r *ui.Tester, role ui.Role, name string) {
	t.Helper()
	n := dateAccess(t, r, role, name)
	r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
	r.Frame()
}
func dateType(t *testing.T, r *ui.Tester, name, text string) {
	t.Helper()
	dateClick(t, r, ui.RoleTextField, name)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type(text)
	r.Frame()
}
func popupClick(t *testing.T, r *ui.Tester, popup, label string) {
	t.Helper()
	p := dateAccess(t, r, ui.RoleDialog, popup)
	for _, n := range r.AccessAll() {
		if n.Role == ui.RoleButton && n.Name == label && n.Bounds.X >= p.Bounds.X && n.Bounds.Y >= p.Bounds.Y && n.Bounds.X+n.Bounds.W <= p.Bounds.X+p.Bounds.W && n.Bounds.Y+n.Bounds.H <= p.Bounds.Y+p.Bounds.H {
			r.ClickAt(n.Bounds.X+8, n.Bounds.Y+n.Bounds.H/2)
			r.Frame()
			return
		}
	}
	t.Fatalf("missing %q inside %q", label, popup)
}

func TestDateCompositionPersianCalendarAndRealInstant(t *testing.T) {
	s := newState()
	r := dateTester(s)
	original := s.selected
	if !r.HasText("خرداد ۱۴۰۴") || !r.HasText("۲۲") {
		t.Fatal("Persian model recipe missing")
	}
	dateClick(t, r, ui.RoleButton, "جمعه ۲۳-ام خرداد ۱۴۰۴")
	if s.selected.Day() != 13 || s.selected.Hour() != 14 || s.selected.Minute() != 35 || s.selected.Location() != original.Location() {
		t.Fatalf("choosing a Persian day changed local clock %v", s.selected)
	}
	dateClick(t, r, ui.RoleSwitch, "Month and year menus")
	dateAccess(t, r, ui.RoleComboBox, "Month")
	dateAccess(t, r, ui.RoleComboBox, "Year")
	before := s.selected
	dateClick(t, r, ui.RoleMenuButton, "Calendar model")
	if err := r.ChooseMenuItem("Gregorian"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if s.selected != before || s.system != "gregorian" || !strings.Contains(s.describe(s.selected), "Jalali 1404/03/23") {
		t.Fatalf("calendar switch changed the selected instant %v", s.selected)
	}
}

func TestDateCompositionEditableExactAndApplicationRelativeParser(t *testing.T) {
	s := newState()
	r := dateTester(s)
	dateType(t, r, "Subscription date", "2025-07-04")
	if s.exactDate.Month() != time.July || s.exactDate.Day() != 4 || s.exact.Text != "2025-07-04" || s.exact.Error != "" {
		t.Fatalf("editable exact %+v %v", s.exact, s.exactDate)
	}
	r.Key(0, ui.KeyDown)
	r.Frame()
	before := s.exactDate
	popupClick(t, r, "Choose subscription date calendar", "Next month")
	if s.exactDate != before || s.exact.Text != "2025-07-04" {
		t.Fatal("browsing modified an accepted text date")
	}
	popupClick(t, r, "Choose subscription date calendar", "August 5, 2025")
	if s.exactDate.Month() != time.August || s.exactDate.Day() != 5 || s.exact.Text != "August 05, 2025" || s.exact.Open || !strings.HasPrefix(s.feedback, "Subscription:") {
		t.Fatalf("calendar feedback %+v %q", s.exact, s.feedback)
	}
	dateType(t, r, "Schedule date", "tomorrow")
	if !s.naturalDate.Equal(s.reference.AddDate(0, 0, 1)) || s.natural.Text != "tomorrow" {
		t.Fatalf("application relative parse %+v %v", s.natural, s.naturalDate)
	}
	valid := s.naturalDate
	dateType(t, r, "Schedule date", "the complete NLP grammar is app-owned")
	if s.naturalDate != valid || s.natural.Error == "" {
		t.Fatal("invalid application grammar overwrote accepted date")
	}
	dateType(t, r, "Schedule date", "next week")
	if !s.naturalDate.Equal(s.reference.AddDate(0, 0, 7)) || s.natural.Error != "" || !strings.HasPrefix(s.feedback, "Schedule:") {
		t.Fatalf("relative feedback %+v %q", s.natural, s.feedback)
	}
}
