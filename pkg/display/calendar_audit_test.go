package display

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestCalendarAutomaticNavigationDoesNotClearCurrentDate(t *testing.T) {
	value := civil(2026, time.November, 1)
	changes := 0
	tt := displayTester(func(c *ui.Context) {
		if CalendarWith(c, &value, CalendarProps{AutoFocus: true}).Changed() {
			changes++
		}
	})
	tt.Key(0, ui.KeyHome)
	if !sameDate(value, civil(2026, time.November, 1)) || changes != 0 {
		t.Fatalf("no-op Home toggled selection: %v changes=%d", value, changes)
	}
	tt.Key(0, ui.KeyEnter)
	if !value.IsZero() || changes != 1 {
		t.Fatalf("explicit activation must still permit clearing: %v changes=%d", value, changes)
	}
}

func TestCalendarDisabledNavigationCannotSkipIntoNextMonth(t *testing.T) {
	value := civil(2026, time.November, 29)
	month := civil(2026, time.November, 1)
	tt := displayTester(func(c *ui.Context) {
		CalendarWith(c, &value, CalendarProps{Month: &month, AutoFocus: true, Required: true, DisableNavigation: true, Disabled: func(d time.Time) bool { return d.Day() == 30 }})
	})
	tt.Key(0, ui.KeyRight)
	if value.Month() != time.November || value.Day() != 29 || month.Month() != time.November {
		t.Fatalf("disabled-day skipping escaped view: date=%v month=%v", value, month)
	}
}

func TestCalendarCaptionNavigationMovesActiveCursorIntoVisibleMonth(t *testing.T) {
	value := civil(2026, time.November, 9)
	month := civil(2026, time.November, 1)
	tt := displayTester(func(c *ui.Context) {
		CalendarWith(c, &value, CalendarProps{Month: &month, Browse: true, Required: true})
	})
	clickAccessible(t, tt, ui.RoleButton, "Next month")
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("December 2026") {
		t.Fatalf("Tab did not reach visible month grid: %+v", tt.AccessAll())
	}
	tt.Key(0, ui.KeyEnter)
	if value.Month() != time.December {
		t.Fatalf("December grid activated stale month cursor: %v", value)
	}
}

func TestDatePickerEmptyTimeCannotChooseForbiddenToday(t *testing.T) {
	var value time.Time
	changes := 0
	tt := displayTester(func(c *ui.Context) {
		if DatePicker(c, &value, DatePickerProps{ShowTime: true, Calendar: CalendarProps{MinDate: civil(2100, time.January, 1)}}).Changed() {
			changes++
		}
	})
	clickAccessible(t, tt, ui.RoleStepper, "Time hours")
	tt.Key(0, ui.KeyUp)
	if !value.IsZero() || changes != 0 {
		t.Fatalf("time editor chose forbidden implicit date: %v changes=%d", value, changes)
	}
}

func TestDatePickerPresetsRespectHiddenDates(t *testing.T) {
	value := civil(2026, time.November, 9)
	tt := displayTester(func(c *ui.Context) {
		DatePicker(c, &value, DatePickerProps{Calendar: CalendarProps{Hidden: func(d time.Time) bool { return d.Day() == 12 }}, Presets: []DatePreset{{Label: "Hidden date", Date: civil(2026, time.November, 12)}}})
	})
	displayClick(t, tt, "Nov 9, 2026")
	clickAccessible(t, tt, ui.RoleButton, "Hidden date")
	if !sameDate(value, civil(2026, time.November, 9)) {
		t.Fatalf("preset bypassed hidden-day selection rule: %v", value)
	}
}

func TestDateRangePresetsRespectHiddenEndpoints(t *testing.T) {
	var value DateRange
	tt := displayTester(func(c *ui.Context) {
		DateRangePicker(c, &value, DatePickerProps{Calendar: CalendarProps{DefaultMonth: civil(2026, time.November, 1), Hidden: func(d time.Time) bool { return d.Day() == 12 }}, RangePresets: []DateRangePreset{{Label: "Hidden endpoint", Range: DateRange{From: civil(2026, time.November, 9), To: civil(2026, time.November, 12)}}}})
	})
	displayClick(t, tt, "Pick a date range")
	clickAccessible(t, tt, ui.RoleButton, "Hidden endpoint")
	if !value.From.IsZero() {
		t.Fatalf("range preset bypassed hidden endpoint: %v", value)
	}
}

func TestDatePickerTimeChangesRespectMatchers(t *testing.T) {
	value := time.Date(2026, time.November, 9, 14, 0, 0, 0, time.UTC)
	changes := 0
	tt := displayTester(func(c *ui.Context) {
		if DatePicker(c, &value, DatePickerProps{ShowTime: true, Calendar: CalendarProps{Disabled: func(d time.Time) bool { return d.Hour() == 15 }}}).Changed() {
			changes++
		}
	})
	clickAccessible(t, tt, ui.RoleStepper, "Time hours")
	tt.Key(0, ui.KeyUp)
	if value.Hour() != 14 || changes != 0 {
		t.Fatalf("time change bypassed matcher: %v changes=%d", value, changes)
	}
}

func TestDateTimeControlsCannotCommitInvalidEndpoint(t *testing.T) {
	for _, kind := range []string{"min", "max", "hidden"} {
		t.Run(kind, func(t *testing.T) {
			value := time.Date(2026, time.November, 9, 14, 0, 0, 0, time.UTC)
			props := CalendarProps{}
			switch kind {
			case "min":
				props.MinDate = civil(2026, time.November, 10)
			case "max":
				props.MaxDate = civil(2026, time.November, 8)
			case "hidden":
				props.Hidden = func(d time.Time) bool { return d.Day() == 9 }
			}
			changes := 0
			tt := displayTester(func(c *ui.Context) {
				if DatePicker(c, &value, DatePickerProps{ShowTime: true, Calendar: props}).Changed() {
					changes++
				}
			})
			clickAccessible(t, tt, ui.RoleStepper, "Time hours")
			tt.Key(0, ui.KeyUp)
			if value.Hour() != 14 || changes != 0 {
				t.Fatalf("invalid %s endpoint committed: %v", kind, value)
			}
		})
	}
	value := DateRange{From: time.Date(2026, time.November, 9, 14, 0, 0, 0, time.UTC), To: time.Date(2026, time.November, 9, 14, 30, 0, 0, time.UTC)}
	tt := displayTester(func(c *ui.Context) { DateRangePicker(c, &value, DatePickerProps{ShowTime: true}) })
	clickAccessible(t, tt, ui.RoleStepper, "End time hours")
	tt.Key(0, ui.KeyDown)
	if value.To.Before(value.From) || value.To.Hour() != 14 {
		t.Fatalf("time editor reversed range: %v to %v", value.From, value.To)
	}
}
