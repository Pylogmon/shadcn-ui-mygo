package display

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestCalendarPersianPinnedArithmeticOracle(t *testing.T) {
	raw, err := os.ReadFile("testdata/shadcn-calendar/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Vectors []struct {
			Persian, Gregorian, Roundtrip [3]int
			Days                          int
			Leap                          bool
		}
		Labels []struct {
			Persian                   [3]int
			Caption, Label, Day, Week string
			DaysInMonth               int
			NextMonth                 [3]int
		}
	}
	if err = json.Unmarshal(raw, &oracle); err != nil {
		t.Fatal(err)
	}
	if len(oracle.Vectors) != 432 || len(oracle.Labels) != 3 {
		t.Fatal("unexpected fixed-source corpus")
	}
	model := modelFor(CalendarPersian)
	for _, v := range oracle.Vectors {
		p := PersianDate{v.Persian[0], v.Persian[1], v.Persian[2]}
		date, err := p.Time(time.UTC)
		if err != nil {
			t.Fatalf("%+v: %v", p, err)
		}
		if got := [3]int{date.Year(), int(date.Month()), date.Day()}; got != v.Gregorian {
			t.Fatalf("source Gregorian %+v: %v != %v", p, got, v.Gregorian)
		}
		back, err := PersianDateOf(date)
		if err != nil || back != p {
			t.Fatalf("roundtrip %+v: %+v/%v", p, back, err)
		}
		length := calendarDaysBetween(model.start(date), model.add(model.start(date), 1))
		if length != v.Days {
			t.Fatalf("source month length %+v: %d != %d", p, length, v.Days)
		}
	}
	for _, v := range oracle.Labels {
		date, _ := (PersianDate{v.Persian[0], v.Persian[1], v.Persian[2]}).Time(time.UTC)
		if model.label("en-US", date, true) != v.Caption || model.label("en-US", date, false) != v.Label || model.dayNumber(date) != v.Day || model.week(date, time.Saturday) != v.Week {
			t.Fatalf("fixed faIR caption/day/label/week differ: %+v", v)
		}
		next := model.move(date, 1)
		y, m, d := model.parts(next)
		if [3]int{y, m, d} != v.NextMonth {
			t.Fatalf("source addMonths clamp %+v", v)
		}
	}
	if _, err := (PersianDate{1404, 12, 30}).Time(time.UTC); err == nil {
		t.Fatal("common-year Esfand 30 normalized silently")
	}
	if _, err := PersianDateOf(time.Time{}); err == nil {
		t.Fatal("zero selection accepted as a Persian day")
	}
}

func TestCalendarPersianNativeModelAndCivilClock(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Fatal(err)
	}
	value := time.Date(2025, 6, 12, 14, 35, 10, 123, loc)
	month := value
	r := displayTester(func(c *ui.Context) {
		CalendarWith(c, &value, CalendarProps{System: CalendarPersian, Location: loc, Month: &month, AutoFocus: true})
	})
	model := modelFor(CalendarPersian)
	if !r.HasText("خرداد ۱۴۰۴") || !r.HasText("۲۲") {
		t.Fatal("fixed Persian caption/numerals missing")
	}
	selected, ok := r.AccessByRole(ui.RoleButton, "پنج‌شنبه ۲۲-ام خرداد ۱۴۰۴")
	if !ok || !selected.Checked {
		t.Fatalf("fixed source selected day %+v", selected)
	}
	saturday, _ := r.AccessByRole(ui.RoleColumnHeader, "ش")
	friday, _ := r.AccessByRole(ui.RoleColumnHeader, "ج")
	if saturday.Bounds.X <= friday.Bounds.X {
		t.Fatal("Persian did not default to RTL/Saturday-first")
	}
	r.Key(0, ui.KeyLeft)
	r.Frame()
	if p, _ := PersianDateOf(value); p != (PersianDate{1404, 3, 23}) || value.Hour() != 14 || value.Minute() != 35 || value.Nanosecond() != 123 || value.Location() != loc {
		t.Fatalf("logical day move changed civil clock: %v (%+v)", value, p)
	}
	value = time.Date(2025, 3, 20, 14, 35, 10, 123, loc)
	month = value
	r.Frame()
	r.Key(ui.Shift, ui.KeyPageDown)
	r.Frame()
	if p, _ := PersianDateOf(value); p != (PersianDate{1404, 12, 29}) || value.Hour() != 14 || value.Location() != loc {
		t.Fatalf("source leap-year clamp: %v (%+v)", value, p)
	}
	r.Key(0, ui.KeyHome)
	r.Frame()
	if p, _ := PersianDateOf(value); p.Day != 1 {
		t.Fatalf("Home used Gregorian month %+v", p)
	}
	r.Key(0, ui.KeyEnd)
	r.Frame()
	if p, _ := PersianDateOf(value); p.Day != 29 {
		t.Fatalf("End used Gregorian month %+v", p)
	}
	if !r.HasText(model.label("en-US", value, true)) {
		t.Fatal("native caption did not follow Persian keyboard month")
	}
}

func TestCalendarPersianRangeBoundsAndCaptionModel(t *testing.T) {
	date := func(y, m, d int) time.Time {
		v, err := (PersianDate{y, m, d}).Time(time.UTC)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	value := DateRange{}
	month := date(1403, 12, 1)
	system := CalendarPersian
	p := CalendarProps{System: system, Month: &month, Location: time.UTC, NumberOfMonths: 2, Min: 2, Max: 10, MinDate: date(1403, 12, 28), MaxDate: date(1404, 1, 5)}
	r := ui.NewTester(func(c *ui.Context) { p.System = system; CalendarRange(c, &value, p) }, 600, 360)
	model := modelFor(CalendarPersian)
	if !r.HasText("اسفند ۱۴۰۳") || !r.HasText("فروردین ۱۴۰۴") {
		t.Fatal("multi-month range used Gregorian arithmetic")
	}
	if n, ok := r.AccessByRole(ui.RoleButton, model.label("en-US", date(1403, 12, 27), false)); !ok || !n.Disabled {
		t.Fatal("real date bounds did not disable Persian cell")
	}
	clickAccessible(t, r, ui.RoleButton, model.label("en-US", date(1403, 12, 28), false))
	clickAccessible(t, r, ui.RoleButton, model.label("en-US", date(1404, 1, 2), false))
	if !sameDate(value.From, date(1403, 12, 28)) || !sameDate(value.To, date(1404, 1, 2)) {
		t.Fatalf("cross-new-year range %+v", value)
	}
	before := value
	system = CalendarGregorian
	r.Frame()
	if value != before || !r.HasText("March 2025") {
		t.Fatalf("switching the month model changed real selected dates %+v", value)
	}
	selected := date(1404, 3, 22)
	month = selected
	r = displayTester(func(c *ui.Context) {
		CalendarWith(c, &selected, CalendarProps{System: CalendarPersian, Month: &month, Caption: CalendarCaptionDropdown, StartMonth: date(1403, 1, 1), EndMonth: date(1405, 12, 1)})
	})
	clickAccessible(t, r, ui.RoleComboBox, "Month")
	r.Key(0, ui.KeyEnd)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p, _ := PersianDateOf(month); p.Year != 1404 || p.Month != 12 || !sameDate(selected, date(1404, 3, 22)) {
		t.Fatalf("caption month was not Jalali: %v selected%v", month, selected)
	}
	clickAccessible(t, r, ui.RoleComboBox, "Year")
	r.Key(0, ui.KeyHome)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if p, _ := PersianDateOf(month); p.Year != 1403 || p.Month != 12 || !sameDate(selected, date(1404, 3, 22)) {
		t.Fatalf("caption year was not Jalali: %v selected%v", month, selected)
	}
}
