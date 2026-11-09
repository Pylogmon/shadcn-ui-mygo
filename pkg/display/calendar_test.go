package display

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func civil(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestCalendarRangeSelectionConstraintsAndDisabledInterior(t *testing.T) {
	var value DateRange
	p := CalendarProps{DefaultMonth: civil(2026, 11, 1), Min: 2, Max: 5, ExcludeDisabled: true, Disabled: func(d time.Time) bool { return d.Day() == 12 }}
	changed := 0
	tt := displayTester(func(c *ui.Context) {
		if CalendarRange(c, &value, p).Changed() {
			changed++
		}
	})
	clickAccessible(t, tt, ui.RoleButton, "November 9, 2026")
	if !sameDate(value.From, civil(2026, 11, 9)) || !value.To.IsZero() {
		t.Fatal(value)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 10, 2026")
	if !value.To.IsZero() || changed != 1 {
		t.Fatalf("minimum nights accepted: %v changes=%d", value, changed)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 15, 2026")
	if !value.To.IsZero() || changed != 1 {
		t.Fatalf("maximum nights accepted: %v changes=%d", value, changed)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 13, 2026")
	if !sameDate(value.From, civil(2026, 11, 13)) || !value.To.IsZero() {
		t.Fatalf("disabled interior did not restart: %v", value)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 16, 2026")
	if !sameDate(value.From, civil(2026, 11, 13)) || !sameDate(value.To, civil(2026, 11, 16)) {
		t.Fatal(value)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 8, 2026")
	clickAccessible(t, tt, ui.RoleButton, "November 5, 2026")
	if !sameDate(value.From, civil(2026, 11, 5)) || !sameDate(value.To, civil(2026, 11, 8)) {
		t.Fatalf("reverse endpoint selection: %v", value)
	}
}

func TestCalendarMultipleNativeSelectionAndLimits(t *testing.T) {
	values := []time.Time{civil(2026, 11, 9)}
	original := values
	tt := displayTester(func(c *ui.Context) { CalendarMultiple(c, &values, CalendarProps{Min: 1, Max: 2}) })
	clickAccessible(t, tt, ui.RoleButton, "November 10, 2026")
	if len(values) != 2 || len(original) != 1 {
		t.Fatalf("selection=%v original=%v", values, original)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 11, 2026")
	if len(values) != 2 {
		t.Fatalf("maximum ignored: %v", values)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 9, 2026")
	if len(values) != 1 || values[0].Day() != 10 {
		t.Fatal(values)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 10, 2026")
	if len(values) != 1 {
		t.Fatalf("minimum ignored: %v", values)
	}
	item, _ := tt.AccessByRole(ui.RoleButton, "November 10, 2026")
	if !item.Checked {
		t.Fatalf("selected day not exposed: %+v", item)
	}
	cell, ok := tt.AccessByRole(ui.RoleCell, "November 10, 2026")
	if !ok || !cell.Selected {
		t.Fatalf("grid selection not exposed on day cell: %+v", cell)
	}
}

func TestCalendarBrowseKeyboardSkipsDisabledAndCommits(t *testing.T) {
	value := civil(2026, 11, 9)
	tt := displayTester(func(c *ui.Context) {
		CalendarWith(c, &value, CalendarProps{Browse: true, AutoFocus: true, Disabled: func(d time.Time) bool { return d.Day() == 10 }})
	})
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if value.Day() != 9 {
		t.Fatalf("browse selected early: %v", value)
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if value.Day() != 11 {
		t.Fatalf("keyboard did not skip disabled date: %v", value)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 10, 2026")
	if value.Day() != 11 {
		t.Fatalf("disabled pointer selection: %v", value)
	}
}

func TestCalendarMonthBoundaryKeyboardAndControlledMonth(t *testing.T) {
	value := time.Date(2028, 1, 31, 14, 35, 10, 123, time.UTC)
	month := civil(2028, 1, 1)
	tt := displayTester(func(c *ui.Context) {
		CalendarWith(c, &value, CalendarProps{Month: &month, AutoFocus: true, MinDate: civil(2028, 1, 1), MaxDate: civil(2028, 2, 29), Required: true})
	})
	tt.Key(0, ui.KeyPageDown)
	tt.Frame()
	if value.Month() != time.February || value.Day() != 29 || value.Hour() != 14 || value.Nanosecond() != 123 || month.Month() != time.February {
		t.Fatalf("month move normalized incorrectly: date=%v month=%v", value, month)
	}
	tt.Key(0, ui.KeyPageDown)
	tt.Frame()
	if value.Month() != time.February {
		t.Fatalf("navigation exceeded bounds: %v", value)
	}
	clickAccessible(t, tt, ui.RoleButton, "Next month")
	if month.Month() != time.February {
		t.Fatalf("button exceeded bounds: %v", month)
	}
}

func TestCalendarLocationPreservesWallClockAndCivilDSTNights(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	value := time.Date(2026, 3, 6, 23, 35, 10, 123, time.UTC)
	tt := displayTester(func(c *ui.Context) { CalendarWith(c, &value, CalendarProps{Location: loc, Required: true}) })
	clickAccessible(t, tt, ui.RoleButton, "March 10, 2026")
	if value.Day() != 10 || value.Hour() != 18 || value.Minute() != 35 || value.Location() != loc || value.Nanosecond() != 123 {
		t.Fatalf("timezone/wall clock changed: %v", value)
	}
	if n := calendarDaysBetween(time.Date(2026, 3, 7, 0, 0, 0, 0, loc), time.Date(2026, 3, 9, 0, 0, 0, 0, loc)); n != 2 {
		t.Fatalf("DST nights=%d", n)
	}
}

func TestDateRangePickerWaitsForSecondEndpointAndRestoresFocus(t *testing.T) {
	value := DateRange{}
	changed := 0
	tt := displayTester(func(c *ui.Context) {
		if DateRangePicker(c, &value, DatePickerProps{Label: "Trip dates", Calendar: CalendarProps{DefaultMonth: civil(2026, 11, 1)}}).Changed() {
			changed++
		}
	})
	displayClick(t, tt, "Pick a date range")
	clickAccessible(t, tt, ui.RoleButton, "November 9, 2026")
	if !tt.HasText("November 2026") || value.From.Day() != 9 || !value.To.IsZero() {
		t.Fatalf("first endpoint prematurely closed: %v", value)
	}
	clickAccessible(t, tt, ui.RoleButton, "November 12, 2026")
	if tt.HasText("November 2026") || value.To.Day() != 12 || changed != 2 {
		t.Fatalf("range did not commit/close: %v changes=%d", value, changed)
	}
	if !tt.Focused("Trip dates") {
		t.Fatal("range picker did not return focus to trigger")
	}
}

func TestCalendarRequiredSingleAndHiddenDates(t *testing.T) {
	value := civil(2026, 11, 9)
	tt := displayTester(func(c *ui.Context) {
		CalendarWith(c, &value, CalendarProps{Required: true, Hidden: func(d time.Time) bool { return d.Day() == 12 }})
	})
	clickAccessible(t, tt, ui.RoleButton, "November 9, 2026")
	if value.IsZero() {
		t.Fatal("required selection cleared")
	}
	if _, ok := tt.AccessByRole(ui.RoleButton, "November 12, 2026"); ok {
		t.Fatal("hidden date remains accessible")
	}
}

func TestCalendarMultipleMonthsRespectNavigationBounds(t *testing.T) {
	value := DateRange{}
	month := civil(2026, 12, 1)
	tt := displayTester(func(c *ui.Context) {
		CalendarRange(c, &value, CalendarProps{Month: &month, StartMonth: civil(2026, 11, 1), EndMonth: civil(2026, 12, 1), NumberOfMonths: 2})
	})
	if !tt.HasText("November 2026") || !tt.HasText("December 2026") || tt.HasText("January 2027") {
		t.Fatalf("bounded visible months=%v", tt.Texts())
	}
	clickAccessible(t, tt, ui.RoleButton, "Next month")
	if tt.HasText("January 2027") {
		t.Fatal("last-month navigation escaped range")
	}
	clickAccessible(t, tt, ui.RoleButton, "Previous month")
	if tt.HasText("October 2026") {
		t.Fatal("first-month navigation escaped range")
	}
}

func TestCalendarDropdownCaptionDoesNotSelectDates(t *testing.T) {
	value := civil(2026, 11, 9)
	month := civil(2026, 11, 1)
	tt := displayTester(func(c *ui.Context) {
		CalendarWith(c, &value, CalendarProps{Month: &month, StartMonth: civil(2025, 1, 1), EndMonth: civil(2027, 12, 1), Caption: CalendarCaptionDropdown})
	})
	clickAccessible(t, tt, ui.RolePopUpButton, "Month")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if month.Month() != time.December || value.Month() != time.November {
		t.Fatalf("month dropdown changed selection/wrong month: date=%v month=%v", value, month)
	}
	clickAccessible(t, tt, ui.RolePopUpButton, "Year")
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if month.Year() != 2025 || value.Year() != 2026 {
		t.Fatalf("year dropdown changed selection/wrong year: date=%v month=%v", value, month)
	}
}

func TestDatePickerPresetsAndTimeUseNativeControls(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*3600)
	value := time.Date(2026, 11, 9, 14, 35, 10, 123, loc)
	changed := 0
	tt := displayTester(func(c *ui.Context) {
		if DatePicker(c, &value, DatePickerProps{Label: "Schedule", Format: "2006/01/02", ShowTime: true, Presets: []DatePreset{{Label: "Tomorrow", Date: civil(2026, 11, 10)}}}).Changed() {
			changed++
		}
	})
	displayClick(t, tt, "2026/11/09")
	displayClick(t, tt, "Tomorrow")
	if value.Day() != 10 || value.Hour() != 14 || value.Location() != loc || changed != 1 {
		t.Fatalf("preset value=%v changes=%d", value, changed)
	}
	clickAccessible(t, tt, ui.RoleStepper, "Time hours")
	tt.Key(0, ui.KeyUp)
	tt.Frame()
	if value.Day() != 10 || value.Hour() != 15 || value.Minute() != 35 || value.Nanosecond() != 123 || changed != 2 {
		t.Fatalf("time editor changed civil date/clock: %v changes=%d", value, changed)
	}
}

func TestDateRangePresetRejectsDisabledInterior(t *testing.T) {
	value := DateRange{}
	tt := displayTester(func(c *ui.Context) {
		DateRangePicker(c, &value, DatePickerProps{Calendar: CalendarProps{DefaultMonth: civil(2026, 11, 1), ExcludeDisabled: true, Disabled: func(d time.Time) bool { return d.Day() == 10 }}, RangePresets: []DateRangePreset{{Label: "Unavailable trip", Range: DateRange{From: civil(2026, 11, 9), To: civil(2026, 11, 12)}}, {Label: "Weekend", Range: DateRange{From: civil(2026, 11, 14), To: civil(2026, 11, 16)}}}})
	})
	displayClick(t, tt, "Pick a date range")
	clickAccessible(t, tt, ui.RoleButton, "Unavailable trip")
	if !value.From.IsZero() {
		t.Fatal(value)
	}
	displayClick(t, tt, "Weekend")
	if value.From.Day() != 14 || value.To.Day() != 16 {
		t.Fatalf("valid range preset=%v", value)
	}
}

func TestCalendarRTLAndYearKeyboardNavigation(t *testing.T) {
	value := civil(2028, 2, 29)
	tt := displayTester(func(c *ui.Context) {
		ui.Column(c).Direction(ui.RightToLeft).Children(func() { CalendarWith(c, &value, CalendarProps{Required: true, AutoFocus: true}) })
	})
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if value.Day() != 28 {
		t.Fatalf("RTL right did not move to previous civil day: %v", value)
	}
	tt.Key(ui.Shift, ui.KeyPageDown)
	tt.Frame()
	if value.Year() != 2029 || value.Month() != time.February || value.Day() != 28 {
		t.Fatalf("year keyboard navigation: %v", value)
	}
}

func TestCalendarPagedAndDisabledNavigation(t *testing.T) {
	value := DateRange{}
	month := civil(2026, 11, 1)
	disabled := false
	tt := displayTester(func(c *ui.Context) {
		CalendarRange(c, &value, CalendarProps{Month: &month, NumberOfMonths: 2, PagedNavigation: true, DisableNavigation: disabled, AutoFocus: true})
	})
	clickAccessible(t, tt, ui.RoleButton, "Next month")
	if month.Year() != 2027 || month.Month() != time.January {
		t.Fatalf("paged navigation advanced wrong number of months: %v", month)
	}
	disabled = true
	tt.Frame()
	clickAccessible(t, tt, ui.RoleButton, "Next month")
	if month.Month() != time.January {
		t.Fatal("disabled navigation changed month")
	}
}

func TestDateRangeTimeControlsAndDateChangesPreserveEndpointClocks(t *testing.T) {
	value := DateRange{From: time.Date(2026, 11, 9, 14, 35, 10, 123, time.UTC), To: time.Date(2026, 11, 12, 18, 50, 20, 456, time.UTC)}
	tt := displayTester(func(c *ui.Context) { DateRangePicker(c, &value, DatePickerProps{Label: "Travel", ShowTime: true}) })
	clickAccessible(t, tt, ui.RoleStepper, "End time minutes")
	tt.Key(0, ui.KeyUp)
	tt.Frame()
	if value.To.Day() != 12 || value.To.Minute() != 51 {
		t.Fatalf("end time edit=%v", value.To)
	}
	displayClick(t, tt, "Nov 9, 2026 – Nov 12, 2026")
	clickAccessible(t, tt, ui.RoleButton, "November 15, 2026")
	clickAccessible(t, tt, ui.RoleButton, "November 20, 2026")
	if value.From.Day() != 15 || value.From.Hour() != 14 || value.From.Minute() != 35 || value.From.Nanosecond() != 123 || value.To.Day() != 20 || value.To.Hour() != 18 || value.To.Minute() != 51 || value.To.Nanosecond() != 456 {
		t.Fatalf("endpoint clocks lost on date change: %v %v", value.From, value.To)
	}
}
