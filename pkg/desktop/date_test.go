package desktop

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func desktopCivil(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
func pickerClickRole(t *testing.T, r *ui.Tester, role ui.Role, label string) {
	t.Helper()
	info, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing %q role %v: %v", label, role, r.Texts())
	}
	r.ClickAt(info.Bounds.X+info.Bounds.W/2, info.Bounds.Y+info.Bounds.H/2)
}
func pickerChild(t *testing.T, b *fakeBackend) *ui.Tester {
	t.Helper()
	if len(b.surfaces) == 0 {
		t.Fatal("native date surface was not created")
	}
	r := ui.NewTester(b.surfaces[len(b.surfaces)-1].view, 680, 500)
	r.Advance(250 * time.Millisecond)
	return r
}

func TestDesktopDatePickerPairsSeparateContextAndCommitOnce(t *testing.T) {
	for _, native := range []bool{false, true} {
		t.Run(map[bool]string{false: "shadcn", true: "native"}[native], func(t *testing.T) {
			b := &fakeBackend{}
			h := NewHostWithBackend(b)
			loc := time.FixedZone("Test offset", -5*3600)
			value := time.Date(2026, 11, 9, 14, 35, 10, 123, loc)
			changed, callbacks := 0, 0
			parentTokens := theme.Preset("violet", true)
			var parentContext *ui.Context
			var childPrimary, childNativeAccent ui.Color
			props := DatePickerProps{Label: "Project date", Calendar: display.CalendarProps{Location: loc, MinDate: desktopCivil(2026, 11, 9), MaxDate: desktopCivil(2026, 11, 13), Disabled: func(d time.Time) bool { return d.Day() == 10 }, Hidden: func(d time.Time) bool { return d.Day() == 11 }}, OnChange: func(time.Time) { callbacks++ }, Footer: func(child *ui.Context) {
				if child == parentContext {
					t.Fatal("picker reused parent frame context")
				}
				childPrimary, childNativeAccent = theme.TokensOf(child).Primary, child.Theme().Accent
				ui.Text(child, "Child footer")
			}}
			parent := ui.NewTester(func(c *ui.Context) {
				parentContext = c
				c.SetTheme(parentTokens.Theme())
				h.BeginFrame()
				defer h.EndFrame()
				ui.Column(c).Padding(16).Children(func() {
					var e ui.Element
					var err error
					if native {
						e, err = NativeDatePicker(c, h, "project-date", &value, props)
					} else {
						e, err = DatePicker(c, h, "project-date", &value, props)
					}
					if err != nil {
						t.Fatal(err)
					}
					if e.Changed() {
						changed++
					}
				})
			}, 420, 160)
			if err := parent.Click("Project date"); err != nil {
				t.Fatal(err)
			}
			if b.surfaces[0].spec.SystemPopover != native || !b.surfaces[0].spec.FitContent {
				t.Fatalf("paired native configuration %+v", b.surfaces[0].spec)
			}
			if parent.HasText("November 2026") {
				t.Fatal("date calendar painted into the parent tree")
			}
			child := pickerChild(t, b)
			if !child.HasText("Child footer") {
				t.Fatal("typed child footer did not render")
			}
			if native {
				if childPrimary != childNativeAccent || childPrimary == parentTokens.Primary {
					t.Fatal("native date calendar did not alias its own platform accent")
				}
			} else if childPrimary != parentTokens.Primary {
				t.Fatal("shadcn date calendar lost its semantic theme")
			}
			child.Key(0, ui.KeyRight)
			if value.Day() != 9 || changed != 0 {
				t.Fatal("keyboard browsing committed before activation")
			}
			child.Key(0, ui.KeyEnter)
			parent.Frame()
			if value.Day() != 12 || value.Hour() != 14 || value.Minute() != 35 || value.Second() != 10 || value.Nanosecond() != 123 || value.Location() != loc {
				t.Fatalf("selection lost constraints or wall clock: %v", value)
			}
			if changed != 1 || callbacks != 1 {
				t.Fatalf("change events=%d callbacks=%d", changed, callbacks)
			}
			parent.Frame()
			if changed != 1 || callbacks != 1 {
				t.Fatal("date commit was delivered twice")
			}
			if !parent.Focused("Project date") {
				t.Fatal("selection did not restore the current parent trigger")
			}
		})
	}
}

func TestDesktopDateTimeDraftCancelConstraintsAndApply(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	value := time.Date(2026, 11, 9, 14, 30, 0, 0, time.UTC)
	initial := value
	changes, cancels := 0, 0
	p := DatePickerProps{Label: "Date and time", ShowTime: true, Calendar: display.CalendarProps{Location: time.UTC, DefaultMonth: desktopCivil(2026, 11, 1), MinDate: desktopCivil(2026, 11, 9), MaxDate: desktopCivil(2026, 11, 12), Disabled: func(d time.Time) bool { return d.Hour() == 15 }}, OnCancel: func() { cancels++ }}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		e, err := DatePicker(c, h, "datetime", &value, p)
		if err != nil {
			t.Fatal(err)
		}
		if e.Changed() {
			changes++
		}
	}, 400, 170)
	if err := parent.Click("Date and time"); err != nil {
		t.Fatal(err)
	}
	child := pickerChild(t, b)
	pickerClickRole(t, child, ui.RoleButton, "November 10, 2026")
	if value != initial || changes != 0 {
		t.Fatal("date-time draft published before Apply")
	}
	pickerClickRole(t, child, ui.RoleButton, "Cancel")
	parent.Frame()
	if value != initial || changes != 0 || cancels != 1 {
		t.Fatalf("cancel value=%v changes=%d cancels=%d", value, changes, cancels)
	}
	if err := parent.Click("Date and time"); err != nil {
		t.Fatal(err)
	}
	child = pickerChild(t, b)
	pickerClickRole(t, child, ui.RoleButton, "November 10, 2026")
	pickerClickRole(t, child, ui.RoleStepper, "Time hours")
	child.Key(0, ui.KeyUp)
	// 15:30 is forbidden by the complete-timestamp matcher; native segmented
	// editing is rolled back to 14:30 before the next accepted key.
	child.Key(0, ui.KeyDown)
	pickerClickRole(t, child, ui.RoleButton, "Apply")
	parent.Frame()
	if value.Day() != 10 || value.Hour() != 13 || value.Minute() != 30 || changes != 1 {
		t.Fatalf("date-time constraints or commit failed value=%v changes=%d", value, changes)
	}
}

func TestDesktopDateRangeIsolatesPartialSelectionAcrossDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	value := display.DateRange{}
	changes := 0
	p := DatePickerProps{Label: "Trip dates", Format: "2006-01-02", Calendar: display.CalendarProps{Location: loc, DefaultMonth: time.Date(2026, 3, 1, 0, 0, 0, 0, loc), NumberOfMonths: 1, Min: 2, Max: 2, ExcludeDisabled: true}}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		e, err := DateRangePicker(c, h, "trip", &value, p)
		if err != nil {
			t.Fatal(err)
		}
		if e.Changed() {
			changes++
		}
	}, 450, 160)
	if err := parent.Click("Trip dates"); err != nil {
		t.Fatal(err)
	}
	child := pickerChild(t, b)
	pickerClickRole(t, child, ui.RoleButton, "March 7, 2026")
	parent.Frame()
	if !value.From.IsZero() || changes != 0 {
		t.Fatalf("incomplete range leaked into parent: %v", value)
	}
	pickerClickRole(t, child, ui.RoleButton, "March 8, 2026")
	if !value.To.IsZero() {
		t.Fatal("minimum two nights was ignored")
	}
	pickerClickRole(t, child, ui.RoleButton, "March 9, 2026")
	parent.Frame()
	if value.From.Day() != 7 || value.To.Day() != 9 || changes != 1 || value.To.Sub(value.From) != 47*time.Hour {
		t.Fatalf("civil DST range=%v events=%d", value, changes)
	}
	if !parent.HasText("2026-03-07 – 2026-03-09") {
		t.Fatalf("range display format lost: %v", parent.Texts())
	}
}

func TestDesktopDatePresetsValidateAndNativeRangeKeepsPlatformColors(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	value := display.DateRange{From: desktopCivil(2026, 11, 9), To: desktopCivil(2026, 11, 10)}
	changes := 0
	props := DatePickerProps{Label: "Native trip", Calendar: display.CalendarProps{Location: time.UTC, Hidden: func(d time.Time) bool { return d.Day() == 12 }, ExcludeDisabled: true, Disabled: func(d time.Time) bool { return d.Day() == 14 }}, RangePresets: []display.DateRangePreset{
		{Label: "Hidden endpoint", Range: display.DateRange{From: desktopCivil(2026, 11, 9), To: desktopCivil(2026, 11, 12)}},
		{Label: "Disabled interior", Range: display.DateRange{From: desktopCivil(2026, 11, 13), To: desktopCivil(2026, 11, 15)}},
		{Label: "Valid trip", Range: display.DateRange{From: desktopCivil(2026, 11, 16), To: desktopCivil(2026, 11, 18)}},
	}}
	parent := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(theme.Preset("rose", true).Theme())
		h.BeginFrame()
		defer h.EndFrame()
		e, err := NativeDateRangePicker(c, h, "native-trip", &value, props)
		if err != nil {
			t.Fatal(err)
		}
		if e.Changed() {
			changes++
		}
	}, 450, 160)
	if err := parent.Click("Native trip"); err != nil {
		t.Fatal(err)
	}
	child := pickerChild(t, b)
	for _, label := range []string{"Hidden endpoint", "Disabled interior"} {
		button, ok := child.AccessByRole(ui.RoleButton, label)
		if !ok || !button.Disabled {
			t.Fatalf("invalid preset selectable: %q %+v", label, button)
		}
	}
	pickerClickRole(t, child, ui.RoleButton, "Valid trip")
	parent.Frame()
	if value.From.Day() != 16 || value.To.Day() != 18 || changes != 1 {
		t.Fatalf("native range preset %v changes=%d", value, changes)
	}
}

func TestDesktopDatePickerDismissExternalValueAndMissingHost(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	value := desktopCivil(2026, 11, 9)
	changes, cancels := 0, 0
	props := DatePickerProps{Label: "Date", Confirm: true, Calendar: display.CalendarProps{Location: time.UTC}, OnCancel: func() { cancels++ }}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		e, err := DatePicker(c, h, "date", &value, props)
		if err != nil {
			t.Fatal(err)
		}
		if e.Changed() {
			changes++
		}
	}, 400, 160)
	if err := parent.Click("Date"); err != nil {
		t.Fatal(err)
	}
	child := pickerChild(t, b)
	pickerClickRole(t, child, ui.RoleButton, "November 10, 2026")
	b.surfaces[0].dismiss()
	parent.Frame()
	if value.Day() != 9 || changes != 0 || cancels != 1 {
		t.Fatal("outside dismissal published a draft")
	}
	if err := parent.Click("Date"); err != nil {
		t.Fatal(err)
	}
	child = pickerChild(t, b)
	value = desktopCivil(2026, 11, 12)
	parent.Frame()
	child.Frame()
	button, ok := child.AccessByRole(ui.RoleButton, "November 12, 2026")
	if !ok || !button.Checked || changes != 0 {
		t.Fatalf("external controlled value did not refresh draft %+v events=%d", button, changes)
	}
	props.Disabled = true
	parent.Frame()
	if e, _ := parent.AccessByRole(ui.RoleButton, "Date"); !e.Disabled || e.Expanded {
		t.Fatalf("disabled picker kept open state %+v", e)
	}
	missingError := false
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		_, err := DatePicker(c, nil, "no-host", &value, DatePickerProps{Label: "No host"})
		missingError = missingError || err != nil
	}, 400, 160)
	if missingError {
		t.Fatal("closed date picker unexpectedly needs a native window")
	}
	if err := r.Click("No host"); err != nil {
		t.Fatal(err)
	}
	if !missingError || r.HasText("November 2026") {
		t.Fatal("missing host failed to report opening error or silently fell back")
	}
}

func TestDesktopDatePickerRejectsParentFooterAndInvalidCompleteTime(t *testing.T) {
	value := time.Date(2026, 11, 9, 14, 0, 0, 0, time.UTC)
	var footerErr error
	ui.NewTester(func(c *ui.Context) {
		_, footerErr = DatePicker(c, nil, "footer", &value, DatePickerProps{Calendar: display.CalendarProps{Footer: func() {}}})
	}, 400, 160)
	if footerErr == nil {
		t.Fatal("unsafe zero-argument footer accepted")
	}
	loc := time.UTC
	p := display.CalendarProps{MinDate: desktopCivil(2026, 11, 9), MaxDate: desktopCivil(2026, 11, 12), Disabled: func(d time.Time) bool { return d.Hour() == 15 }, Hidden: func(d time.Time) bool { return d.Day() == 10 }}
	for _, v := range []time.Time{value.Add(time.Hour), desktopCivil(2026, 11, 10), desktopCivil(2026, 11, 13)} {
		if pickerValidDate(v, p, loc) {
			t.Fatalf("invalid timestamp was accepted: %v", v)
		}
	}
	if pickerValidRange(display.DateRange{From: value, To: value.Add(-time.Hour)}, p, loc) {
		t.Fatal("same-day inverted time range accepted")
	}
}

func TestNativeDateRangeTimeDraftPreservesEndpointClocksAndRejectsInversion(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*3600)
	value := display.DateRange{From: time.Date(2026, 11, 9, 9, 15, 0, 0, loc), To: time.Date(2026, 11, 9, 10, 45, 0, 0, loc)}
	initial := value
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	changes := 0
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		e, err := NativeDateRangePicker(c, h, "range-time", &value, DatePickerProps{Label: "Range time", ShowTime: true, Calendar: display.CalendarProps{Location: loc, NumberOfMonths: 1}})
		if err != nil {
			t.Fatal(err)
		}
		if e.Changed() {
			changes++
		}
	}, 400, 160)
	if err := parent.Click("Range time"); err != nil {
		t.Fatal(err)
	}
	child := pickerChild(t, b)
	pickerClickRole(t, child, ui.RoleButton, "November 12, 2026")
	pickerClickRole(t, child, ui.RoleButton, "November 12, 2026")
	if !pickerRangeEqual(value, initial) || changes != 0 {
		t.Fatal("date-time range leaked before confirmation")
	}
	pickerClickRole(t, child, ui.RoleStepper, "Start time hours")
	child.Key(0, ui.KeyUp)
	child.Key(0, ui.KeyUp) // 11:15 would be later than the 10:45 endpoint.
	pickerClickRole(t, child, ui.RoleButton, "Apply")
	parent.Frame()
	if changes != 1 || value.From.Day() != 12 || value.To.Day() != 12 || value.From.Hour() != 10 || value.From.Minute() != 15 || value.To.Hour() != 10 || value.To.Minute() != 45 || value.From.Location() != loc || value.To.Location() != loc {
		t.Fatalf("range clock preservation/inversion: %+v changes=%d", value, changes)
	}
}

func TestDesktopDatePickerRefreshesLiveConstraintsWithoutIdleRedraws(t *testing.T) {
	b := &fakeBackend{}
	h := NewHostWithBackend(b)
	value := desktopCivil(2026, 11, 9)
	p := DatePickerProps{Label: "Live constraints", Confirm: true, Calendar: display.CalendarProps{Location: time.UTC, MinDate: desktopCivil(2026, 11, 9)}}
	parent := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		if _, err := DatePicker(c, h, "live-date", &value, p); err != nil {
			t.Fatal(err)
		}
	}, 400, 160)
	if err := parent.Click("Live constraints"); err != nil {
		t.Fatal(err)
	}
	child := pickerChild(t, b)
	surface := b.surfaces[len(b.surfaces)-1]
	before := surface.invalidates
	p.Calendar.MinDate = desktopCivil(2026, 11, 12)
	parent.Frame()
	if surface.invalidates <= before {
		t.Fatal("live date constraint did not invalidate open child")
	}
	child.Frame()
	day, _ := child.AccessByRole(ui.RoleButton, "November 9, 2026")
	apply, _ := child.AccessByRole(ui.RoleButton, "Apply")
	if !day.Disabled || !apply.Disabled || value.Day() != 9 {
		t.Fatalf("updated constraint bypassed: day=%+v apply=%+v value=%v", day, apply, value)
	}
	before = surface.invalidates
	for range 3 {
		parent.Frame()
	}
	if surface.invalidates != before {
		t.Fatal("stable date options caused idle child invalidation")
	}
	before = surface.invalidates
	p.Calendar.MinDate = value
	p.Calendar.System = display.CalendarPersian
	parent.Frame()
	child.Frame()
	if surface.invalidates <= before || len(b.surfaces) != 1 || !child.HasText("آبان ۱۴۰۵") || value.Day() != 9 {
		t.Fatal("live month model did not refresh the same child without changing its real value")
	}
	direction := ui.LeftToRight
	p.Calendar.Direction = &direction
	before = surface.invalidates
	parent.Frame()
	child.Frame()
	saturday, _ := child.AccessByRole(ui.RoleColumnHeader, "ش")
	friday, _ := child.AccessByRole(ui.RoleColumnHeader, "ج")
	if surface.invalidates <= before || saturday.Bounds.X >= friday.Bounds.X {
		t.Fatal("explicit direction did not refresh Persian child")
	}
	direction = ui.RightToLeft
	before = surface.invalidates
	parent.Frame()
	child.Frame()
	saturday, _ = child.AccessByRole(ui.RoleColumnHeader, "ش")
	friday, _ = child.AccessByRole(ui.RoleColumnHeader, "ج")
	if surface.invalidates <= before || saturday.Bounds.X <= friday.Bounds.X {
		t.Fatal("changed direction value at the same address did not refresh child")
	}
}
